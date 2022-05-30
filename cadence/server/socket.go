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
		log.Print("New socket connected.", conn.RemoteAddr())

		type Message struct {
			Type       string
			Artist     string
			Title      string
			Picture    []byte
			ListenURL  string
			Mountpoint string
			Listeners  float64
		}

		var lastArtist string
		var lastTitle string
		var lastListenURL string
		var lastMountpoint string
		var lastListeners float64

		for {
			resp, err := http.Get("http://" + c.StreamAddress + c.StreamPort + "/status-json.xsl")
			if err != nil {
				clog.Error("RadioData", "Failed to connect to audio stream server.", err)
				conn.WriteJSON(Message{Type: "NowPlaying", Title: "-"})
				conn.WriteJSON(Message{Type: "Listeners", Listeners: -1})
				conn.WriteJSON(Message{Type: "StreamConnection", Mountpoint: "N/A"})
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				clog.Error("RadioData", "Audio stream server returned bad status", err)
				return
			}

			body, _ := io.ReadAll(resp.Body)
			jsonParsed, _ := gabs.ParseJSON([]byte(body))

			var currentArtist = fmt.Sprintf(jsonParsed.Path("icestats.source.artist").Data().(string))
			var currentTitle = fmt.Sprintf(jsonParsed.Path("icestats.source.title").Data().(string))
			var currentHost = fmt.Sprintf(jsonParsed.Path("icestats.host").Data().(string))
			var currentMountpoint = fmt.Sprintf(jsonParsed.Path("icestats.source.server_name").Data().(string))
			var currentListeners = jsonParsed.Path("icestats.source.listeners").Data().(float64)

			var currentListenURL = currentHost + "/" + currentMountpoint

			if (lastArtist != currentArtist) || (lastTitle != currentTitle) {
				currentPicture := getAlbumArt(currentArtist, currentTitle)

				conn.WriteJSON(Message{Type: "NowPlaying", Artist: currentArtist, Title: currentTitle, Picture: currentPicture})
				lastArtist = currentArtist
				lastTitle = currentTitle
			}
			if (lastListenURL != currentListenURL) || (lastMountpoint != currentMountpoint) {
				conn.WriteJSON(Message{Type: "StreamConnection", ListenURL: currentListenURL, Mountpoint: currentMountpoint})
				lastListenURL = currentListenURL
				lastMountpoint = currentMountpoint
			}
			if lastListeners != currentListeners {
				conn.WriteJSON(Message{Type: "Listeners", Listeners: currentListeners})
				lastListeners = currentListeners
			}
			time.Sleep(1 * time.Second)
		}
	}
}
