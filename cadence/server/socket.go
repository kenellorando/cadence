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

var upgrader = websocket.Upgrader{
	ReadBufferSize:  0,
	WriteBufferSize: 0,
}

// RadioData() upgrades connections for websocket
// Transfers near real-time radio updates (now playing, stream URL, listener count)
func RadioData() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("Error upgrading socket connection.", err)
			return
		}
		clog.Debug("RadioData", fmt.Sprintf("New socket connected: %s", conn.RemoteAddr()))

		type Message struct {
			Type       string
			Artist     string
			Title      string
			ListenURL  string
			Mountpoint string
			Listeners  float64
		}

		var lastArtist string
		var lastTitle string
		var lastHost string
		var lastMountpoint string
		var lastListeners float64

		for {
			clog.Debug("RadioData", "Loop start")

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

			var currentArtist = fmt.Sprintf(jsonParsed.Path("icestats.source.artist").Data().(string))
			var currentTitle = fmt.Sprintf(jsonParsed.Path("icestats.source.title").Data().(string))
			var currentHost = fmt.Sprintf(jsonParsed.Path("icestats.host").Data().(string))
			var currentMountpoint = fmt.Sprintf(jsonParsed.Path("icestats.source.server_name").Data().(string))
			var currentListeners = jsonParsed.Path("icestats.source.listeners").Data().(float64)

			var currentListenURL = currentHost + "/" + currentMountpoint

			if (lastArtist != currentArtist) || (lastTitle != currentTitle) {

				clog.Debug("RadioData", "artist or title change detected")
				//currentPicture := getAlbumArt(currentArtist, currentTitle)

				//clog.Debug("RadioData", string(currentPicture))
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
				currentListenURL = currentHost + "/" + currentMountpoint

				conn.WriteJSON(Message{Type: "StreamConnection", ListenURL: currentListenURL, Mountpoint: currentMountpoint})
				lastHost = currentHost
				lastMountpoint = currentMountpoint
			}
			if lastListeners != currentListeners {
				conn.WriteJSON(Message{Type: "Listeners", Listeners: currentListeners})
				lastListeners = currentListeners
			}
			time.Sleep(2 * time.Second)
		}
	}
}
