// handlers.go
// API functions and fileservers

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/Jeffail/gabs"
	"github.com/dhowden/tag"
	"github.com/gorilla/websocket"
	"github.com/kenellorando/clog"
)

func SiteRoot() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clog.Info("ServeRoot", fmt.Sprintf("Client %s requesting %s%s", r.Header.Get("X-Forwarded-For"), r.Host, r.URL.Path))
		w.Header().Set("Content-type", "text/html")
		w.WriteHeader(http.StatusOK) // 200 Accepted
		http.ServeFile(w, r, path.Dir(c.RootPath+"./public/index.html"))
	}
}

func Site404() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clog.Info("Serve404", fmt.Sprintf("Client %s requesting unknown resource %s%s. Returning 404.", r.Header.Get("X-Forwarded-For"), r.Host, r.URL.Path))
		w.Header().Set("Content-type", "text/html")
		w.WriteHeader(http.StatusOK) // 200 Accepted
		http.ServeFile(w, r, path.Dir(c.RootPath+"./public/404/index.html"))
	}
}

// /api/search
func Search() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clog.Info("Search", fmt.Sprintf("Search by client %s.", r.Header.Get("X-Forwarded-For")))

		type Search struct {
			Query string `json:"search"`
		}
		var search Search

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&search)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest) // 400 Bad Request
			return
		}

		queryResults, err := searchByQuery(search.Query)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}

		jsonMarshal, _ := json.Marshal(queryResults)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK) // 200 OK
		w.Write(jsonMarshal)
	}
}

// /api/request/id
func RequestID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clog.Info("Request", fmt.Sprintf("Request-by-ID by client %s.", r.Header.Get("X-Forwarded-For")))

		type Request struct {
			ID int `json:"ID"`
		}
		var request Request

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&request)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest) // 400 Bad Request
			return
		}

		path, err := getPathById(request.ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}
		_, err = pushRequest(path)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}

		w.WriteHeader(http.StatusAccepted) // 202 Accepted
	}
}

// /api/request/bestmatch
func RequestBestMatch() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clog.Debug("Search", fmt.Sprintf("Decoding http-request data from client %s.", r.Header.Get("X-Forwarded-For")))

		type RequestBestMatch struct {
			Query string `json:"Search"`
		}
		var rbm RequestBestMatch

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&rbm)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest) // 400 Bad Request
			return
		}

		queryResults, err := searchByQuery(rbm.Query)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}
		path, err := getPathById(queryResults[0].ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}
		_, err = pushRequest(path)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}
		w.WriteHeader(http.StatusAccepted) // 202 Accepted
	}
}

// /api/nowplaying/metadata
func NowPlayingMetadata() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		title, artist, err := getNowPlaying()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}
		queryResults, err := searchByTitleArtist(title, artist)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}
		song := queryResults[0]

		// Return data to client
		type SongData struct {
			ID     int
			Artist string
			Title  string
			Album  string
			Genre  string
			Year   int
		}
		result := SongData{ID: song.ID, Artist: song.Artist, Title: song.Title, Album: song.Album, Genre: song.Genre, Year: song.Year}
		jsonMarshal, _ := json.Marshal(result)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK) // 200 OK
		w.Write(jsonMarshal)
	}
}

// /api/nowplaying/albumart
func NowPlayingAlbumArt() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		title, artist, err := getNowPlaying()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}
		queryResults, err := searchByTitleArtist(title, artist)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}
		path, err := getPathById(queryResults[0].ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}

		file, err := os.Open(path)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}
		// Read metadata from the file
		tags, err := tag.ReadFrom(file)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}

		// Return data to client
		type SongData struct {
			Picture []byte
		}
		result := SongData{Picture: tags.Picture().Data}
		jsonMarshal, _ := json.Marshal(result)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK) // 200 OK
		w.Write(jsonMarshal)
	}
}

// /api/version
func Version() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Return data to client
		type CadenceVersion struct {
			Version string
		}
		version := CadenceVersion{Version: c.Version}
		jsonMarshal, _ := json.Marshal(version)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK) // 200 OK
		w.Write(jsonMarshal)
	}
}

// /ready
func Ready() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK) // 200 OK
	}
}

/////////////////

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
