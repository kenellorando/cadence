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

// socketRadioData() upgrades connections for websocket
// Transfers near real-time radio updates (now playing, stream URL, listener count)
func socketRadioData() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("Error upgrading socket connection.", err)
			return
		}
		log.Print("New socket connected.", conn.RemoteAddr())

		// type NowPlaying struct {
		// 	Artist string
		// 	Title  string
		// }

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
		var lastListenURL string
		var lastMountpoint string
		var lastListeners float64

		for {
			resp, err := http.Get("http://icecast2:8000/status-json.xsl")
			if err != nil {
				clog.Error("socketNowPlaying", "Failed to connect to audio stream server.", err)
				conn.WriteJSON(Message{Type: "NowPlaying", Title: "-"})                  // Write message to client
				conn.WriteJSON(Message{Type: "Listeners", Listeners: -1})                // Write message to client
				conn.WriteJSON(Message{Type: "StreamConnection", Mountpoint: "unknown"}) // Write message to client
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				clog.Error("socketNowPlaying", "Audio stream server returned bad status", err)
				return
			}

			body, _ := io.ReadAll(resp.Body)
			jsonParsed, _ := gabs.ParseJSON([]byte(body))

			var currentArtist = fmt.Sprintf(jsonParsed.Path("icestats.source.artist").Data().(string))
			var currentTitle = fmt.Sprintf(jsonParsed.Path("icestats.source.title").Data().(string))
			var currentListenURL = fmt.Sprintf(jsonParsed.Path("icestats.source.listenurl").Data().(string))
			var currentMountpoint = fmt.Sprintf(jsonParsed.Path("icestats.source.server_name").Data().(string))
			var currentListeners = jsonParsed.Path("icestats.source.listeners").Data().(float64)

			if (lastArtist != currentArtist) || (lastTitle != currentTitle) {
				conn.WriteJSON(Message{Type: "NowPlaying", Artist: currentArtist, Title: currentTitle})
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
