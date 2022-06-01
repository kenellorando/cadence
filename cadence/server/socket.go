// socket.go
// WebSocket API

package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/Jeffail/gabs"
	"github.com/gorilla/websocket"
	"github.com/kenellorando/clog"
)

var upgrader = websocket.Upgrader{}

// RadioData() upgrades a connection for websocket
func RadioData() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clog.Debug("RadioDataLoop", "Received new socket connection")
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("Error upgrading socket connection.", err)
			return
		}
		defer conn.Close()

		type Message struct {
			Type       string
			Artist     string
			Title      string
			Host       string
			Mountpoint string
			ListenURL  string
			Listeners  float64
		}

		var lastArtist, lastTitle, lastHost, lastMountpoint string
		var lastListeners float64

		var currentArtist, currentTitle, currentHost, currentMountpoint string
		var currentListeners float64

		for {
			resp, err := http.Get("http://" + c.StreamAddress + c.StreamPort + "/status-json.xsl")
			if err != nil {
				clog.Error("RadioData", "Failed to connect to audio stream server.", err)
				conn.WriteJSON(Message{Type: "NowPlaying", Title: "-", Artist: "-"})
				conn.WriteJSON(Message{Type: "Listeners", Listeners: -1})
				conn.WriteJSON(Message{Type: "StreamConnection", Mountpoint: "N/A", ListenURL: "N/A"})
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				clog.Error("RadioData", "Audio stream server returned bad status", err)
				return
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				clog.Error("RadioData", "Failed to read stream server response.", err)
				conn.WriteJSON(Message{Type: "NowPlaying", Title: "-"})
				conn.WriteJSON(Message{Type: "Listeners", Listeners: -1})
				conn.WriteJSON(Message{Type: "StreamConnection", Mountpoint: "N/A"})
				return
			}
			jsonParsed, err := gabs.ParseJSON([]byte(body))
			if err != nil {
				clog.Error("RadioData", "Failed to parse stream server response.", err)
				conn.WriteJSON(Message{Type: "NowPlaying", Title: "-"})
				conn.WriteJSON(Message{Type: "Listeners", Listeners: -1})
				conn.WriteJSON(Message{Type: "StreamConnection", Mountpoint: "N/A"})
				return
			}

			currentArtist = fmt.Sprintf(jsonParsed.Path("icestats.source.artist").Data().(string))
			currentTitle = fmt.Sprintf(jsonParsed.Path("icestats.source.title").Data().(string))
			currentHost = fmt.Sprintf(jsonParsed.Path("icestats.host").Data().(string))
			currentMountpoint = fmt.Sprintf(jsonParsed.Path("icestats.source.server_name").Data().(string))
			currentListeners = jsonParsed.Path("icestats.source.listeners").Data().(float64)

			if (lastArtist != currentArtist) || (lastTitle != currentTitle) {

				clog.Debug("RadioData", "artist or title change detected")
				clog.Debug("RadioData", currentArtist)
				clog.Debug("RadioData", currentTitle)
				clog.Debug("RadioData", "Writing connection")

				err = conn.WriteJSON(Message{Type: "NowPlaying", Artist: currentArtist, Title: currentTitle})
				if err != nil {
					clog.Error("RadioData", "There was a problem writing the connection", err)
				}
				lastArtist = currentArtist
				lastTitle = currentTitle
			}
			if (lastHost != currentHost) || (lastMountpoint != currentMountpoint) {
				conn.WriteJSON(Message{Type: "StreamConnection", Host: currentHost, Mountpoint: currentMountpoint})
				lastHost = currentHost
				lastMountpoint = currentMountpoint
			}
			if lastListeners != currentListeners {
				conn.WriteJSON(Message{Type: "Listeners", Listeners: currentListeners})
				lastListeners = currentListeners
			}

			// Ping the client to maintain the connection
			// Close it in the event of an error
			err = conn.WriteMessage(websocket.PingMessage, []byte{})
			if err != nil {
				clog.Error("RadioData", "Error writing Ping.", err)
				conn.Close()
				return
			}
			time.Sleep(1 * time.Second)
		}
	}
}
