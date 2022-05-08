package main

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/Jeffail/gabs"
	"github.com/gorilla/websocket"
	"github.com/kenellorando/clog"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// socketNowPlaying() upgrades connections for websocket
// and maintains a pipeline for server->client radio updates
func socketNowPlaying() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("Error upgrading socket connection.", err)
			return
		}
		log.Print("New socket connected.", conn.RemoteAddr())

		type NowPlaying struct {
			Artist string
			Title  string
		}

		var lastArtist string
		var lastTitle string

		for {
			resp, err := http.Get("http://icecast2:8000/status-json.xsl")
			if err != nil {
				clog.Error("socketNowPlaying", "Failed to connect to audio stream server.", err)
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

			if (lastArtist != currentArtist) || (lastTitle != currentTitle) {

				clog.Debug("socketNowPlaying", fmt.Sprintf("Broadcast update: %s by %s", currentTitle, currentArtist))
				// Return data to client
				type NowPlayingUpdate struct {
					Artist string
					Title  string
				}
				conn.WriteJSON(NowPlayingUpdate{Artist: currentArtist, Title: currentTitle}) // Write message to client

				lastArtist = currentArtist
				lastTitle = currentTitle
			}
		}
	}
}
