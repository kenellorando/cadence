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
				type NowPlayingUpdate struct {
					Artist string
					Title  string
				}
				conn.WriteJSON(NowPlayingUpdate{Artist: "-", Title: "-"}) // Write message to client

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
			//time.Sleep(2 * time.Second)
		}
	}
}

// socketStreamURL() maintains a server->client socket for server URL
func socketStreamURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("Error upgrading socket connection.", err)
			return
		}
		log.Print("New socket connected.", conn.RemoteAddr())

		var lastListenURL string
		var lastMountpoint string

		for {
			resp, err := http.Get("http://icecast2:8000/status-json.xsl")
			if err != nil {
				clog.Error("socketNowPlaying", "Failed to connect to audio stream server.", err)
				type StatusUpdate struct {
					ListenURL  string
					Mountpoint string
				}
				conn.WriteJSON(StatusUpdate{ListenURL: "unknown", Mountpoint: "unknown"}) // Write message to client
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				clog.Error("socketNowPlaying", "Audio stream server returned bad status", err)
				return
			}

			body, _ := io.ReadAll(resp.Body)
			jsonParsed, _ := gabs.ParseJSON([]byte(body))

			var currentListenURL = fmt.Sprintf(jsonParsed.Path("icestats.source.listenurl").Data().(string))
			var currentMountpoint = fmt.Sprintf(jsonParsed.Path("icestats.source.server_name").Data().(string))

			if lastListenURL != currentListenURL || lastMountpoint != currentMountpoint {
				clog.Info("socketStatus", fmt.Sprintf("The stream server reports a new stream URL: %s", currentListenURL))
				// Return data to client
				type StatusUpdate struct {
					ListenURL  string
					Mountpoint string
				}
				conn.WriteJSON(StatusUpdate{ListenURL: currentListenURL, Mountpoint: currentMountpoint}) // Write message to client
				lastListenURL = currentListenURL
				lastMountpoint = currentMountpoint
			}
			//time.Sleep(2 * time.Second)
		}
	}
}

// socketStreamListeners() maintains a server->client socket for listeners
func socketStreamListeners() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("Error upgrading socket connection.", err)
			return
		}
		log.Print("New socket connected.", conn.RemoteAddr())

		var lastListeners float64

		for {
			resp, err := http.Get("http://icecast2:8000/status-json.xsl")
			if err != nil {
				clog.Error("socketNowPlaying", "Failed to connect to audio stream server.", err)
				type StatusUpdate struct {
					Listeners float64
				}
				conn.WriteJSON(StatusUpdate{Listeners: -1}) // Write message to client

				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				clog.Error("socketNowPlaying", "Audio stream server returned bad status", err)
				return
			}

			body, _ := io.ReadAll(resp.Body)
			jsonParsed, _ := gabs.ParseJSON([]byte(body))

			var currentListeners = jsonParsed.Path("icestats.source.listeners").Data().(float64)

			if lastListeners != currentListeners {
				clog.Debug("socketStatus", fmt.Sprintf("Current radio listeners: %v", currentListeners))
				// Return data to client
				type StatusUpdate struct {
					Listeners float64
				}
				conn.WriteJSON(StatusUpdate{Listeners: currentListeners}) // Write message to client
				lastListeners = currentListeners
			}
			//time.Sleep(2 * time.Second)
		}
	}
}
