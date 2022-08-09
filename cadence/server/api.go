// handlers.go
// REST API

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/Jeffail/gabs"
	"github.com/dhowden/tag"
	"github.com/gorilla/websocket"
	"github.com/kenellorando/clog"
)

// Map of requester IPs to be locked out of making requests
var requestTimeoutIPs = make(map[string]int)

// Fileservers ////////////////////////////////////////////////////////////////////
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

// Default API ////////////////////////////////////////////////////////////////////
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
			return
		}

		queryResults, err := dbQuery(search.Query)
		if err != nil {

		}

		jsonMarshal, _ := json.Marshal(queryResults)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK) // 200 OK
		w.Write(jsonMarshal)
	}
}

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

		path, err := dbQueryPath(request.ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		_, err = pushRequest(path)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}

		w.WriteHeader(http.StatusAccepted) // 202 Accepted
	}
}

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

		queryResults, err := dbQuery(rbm.Query)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		path, err := dbQueryPath(queryResults[0].ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		_, err = pushRequest(path)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusAccepted) // 202 Accepted
	}
}

func NowPlayingMetadata() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Since Cadence does maintain state on what the stream server is playing
		// it needs to check first for the simple data the stream server provides.
		clog.Debug("NowPlaying", fmt.Sprintf("Client %s requesting %s%s", r.Header.Get("X-Forwarded-For"), r.Host, r.URL.Path))

		resp, err := http.Get("http://icecast2:8000/status-json.xsl")
		if err != nil {
			clog.Error("NowPlaying", "Failed to connect to audio stream server.", err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			clog.Error("NowPlaying", "Audio stream server returned bad status", err)
			return
		}

		body, _ := io.ReadAll(resp.Body)
		jsonParsed, _ := gabs.ParseJSON([]byte(body))

		var artist, _ = jsonParsed.Path("icestats.source.artist").Data().(string)
		var title, _ = jsonParsed.Path("icestats.source.title").Data().(string)

		clog.Debug("NowPlayingMetadata", fmt.Sprintf("The stream server reports it is now playing: '%s' by '%s'.", title, artist))

		// When the title and artist are known, we query the metadata DB for all of the data it has on the playing track
		selectStatement := fmt.Sprintf("SELECT rowid,artist,title,album,genre,year FROM %s WHERE artist=\"%v\" AND title=\"%v\";", c.MetadataTable, artist, title)
		rows, err := db.Query(selectStatement)
		if err != nil {
			clog.Error("NowPlayingMetadata", "Could not query DB.", err)
			return
		}
		if err != nil {
			clog.Error("NowPlayingMetadata", "Could not query the DB.", err)
			return
		}

		// Declare object for a song
		type SongData struct {
			ID     int
			Artist string
			Title  string
			Album  string
			Genre  string
			Year   int
		}

		clog.Debug("NowPlayingMetadata", "Scanning returned data...")
		song := new(SongData)
		for rows.Next() {
			err := rows.Scan(&song.ID, &song.Artist, &song.Title, &song.Album, &song.Genre, &song.Year)
			if err != nil {
				clog.Error("NowPlayingMetadata", "Data scan failed.", err)
				return
			}
		}
		result := SongData{ID: song.ID, Artist: song.Artist, Title: song.Title, Album: song.Album, Genre: song.Genre, Year: song.Year}

		// Return data to client
		jsonMarshal, _ := json.Marshal(result)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonMarshal)
	}
}

// /api/nowplaying/albumart
func NowPlayingAlbumArt() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Since Cadence does maintain state on what the stream server is playing
		// it needs to check first for the simple data the stream server provides.
		clog.Debug("NowPlayingAlbumArt", fmt.Sprintf("Client %s requesting %s%s", r.Header.Get("X-Forwarded-For"), r.Host, r.URL.Path))

		resp, err := http.Get("http://" + c.StreamAddress + c.StreamPort + "/status-json.xsl")
		if err != nil {
			clog.Error("NowPlayingAlbumArt", "Failed to connect to audio stream server.", err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			clog.Error("NowPlayingAlbumArt", "Audio stream server returned bad status", err)
			return
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			clog.Error("NowPlayingAlbumArt", "error", err)
		}
		jsonParsed, err := gabs.ParseJSON([]byte(body))
		if err != nil {
			clog.Error("NowPlayingAlbumArt", "error", err)
		}

		var artist, _ = jsonParsed.Path("icestats.source.artist").Data().(string)
		var title, _ = jsonParsed.Path("icestats.source.title").Data().(string)

		clog.Debug("NowPlayingAlbumArt", fmt.Sprintf("The stream server reports it is now playing: '%s' by '%s'.", title, artist))

		log.Printf("%s %s", artist, title)

		selectStatement := fmt.Sprintf("SELECT path FROM %s WHERE artist=\"%v\" AND title=\"%v\";", c.MetadataTable, artist, title)
		rows, err := db.Query(selectStatement)
		if err != nil {
			clog.Error("NowPlayingAlbumArt", "Could not query the DB for a path.", err)
			return
		}
		if err != nil {
			clog.Error("NowPlayingAlbumArt", "Could not query the DB for a path.", err)
			return
		}

		var pic []byte

		for rows.Next() {
			var path string
			err := rows.Scan(&path)
			if err != nil {
				clog.Debug("NowPlayingAlbumArt", "ae")
				return
			}
			// Open a file for reading
			file, err := os.Open(path)
			if err != nil {
				clog.Error("NowPlayingAlbumArt", "Could not open music file for album art.", err)
				return
			}
			// Read metadata from the file
			tags, err := tag.ReadFrom(file)
			if err != nil {
				clog.Error("NowPlayingAlbumArt", "Could not read tags from file for album art.", err)
				return
			}
			pic = tags.Picture().Data
		}
		// Declare object for a song
		type SongData struct {
			Picture []byte
		}
		result := SongData{Picture: pic}

		// Return data to client
		jsonMarshal, _ := json.Marshal(result)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonMarshal)
	}
}

func Version() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clog.Debug("Version", fmt.Sprintf("Client %s requesting %s%s", r.Header.Get("X-Forwarded-For"), r.Host, r.URL.Path))

		// Return data to client
		type CadenceVersion struct {
			Version string
		}
		version := CadenceVersion{Version: c.Version}
		jsonMarshal, _ := json.Marshal(version)

		w.WriteHeader(http.StatusOK) // 200 Accepted
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonMarshal)
	}
}

// Status Checks ////////////////////////////////////////////////////////////////////
func Ready() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK) // 200 Accepted
	}
}

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

func tokenCheck(token string) bool {
	clog.Info("tokenCheck", fmt.Sprintf("Checking token %s...", token))

	if len(token) != 26 {
		clog.Debug("tokenCheck", fmt.Sprintf("Token %s does not satisfy length requirements.", token))
		return false
	}

	// Check the whitelist. If this fails, the whitelist is not configured. No panic is thrown, but the bypass is denied.
	b, err := ioutil.ReadFile(c.WhitelistPath)
	if err != nil {
		return false
	}
	s := string(b)

	if strings.Contains(s, token) {
		clog.Info("tokenCheck", fmt.Sprintf("Token %s is valid.", token))
		return true
	}
	clog.Info("tokenCheck", fmt.Sprintf("Token %s is invalid.", token))
	return false
}
