// handlers.go
// REST API

package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"strconv"
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
		http.ServeFile(w, r, path.Dir(c.RootPath+"./public/index.html"))
	}
}

func Site404() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clog.Info("Serve404", fmt.Sprintf("Client %s requesting unknown resource %s%s. Returning 404.", r.Header.Get("X-Forwarded-For"), r.Host, r.URL.Path))
		w.Header().Set("Content-type", "text/html")
		http.ServeFile(w, r, path.Dir(c.RootPath+"./public/404/index.html"))
	}
}

// Default API ////////////////////////////////////////////////////////////////////
func Search() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clog.Debug("Search", fmt.Sprintf("Decoding http-request data from client %s.", r.Header.Get("X-Forwarded-For")))
		// Declare object to hold r body data
		type Search struct {
			Query string `json:"search"`
		}
		var search Search

		// Decode json object
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&search)
		if err != nil {
			clog.Error("Search", fmt.Sprintf("Failed to read http-request body from %s.", r.Header.Get("X-Forwarded-For")), err)
			return
		}

		query := search.Query
		clog.Debug("Search", fmt.Sprintf("Search query decoded: '%v'", query))
		clog.Info("Search", fmt.Sprintf("Querying database for: '%v'", query))

		// Query database
		selectStatement := fmt.Sprintf("SELECT \"rowid\", \"artist\", \"title\",\"album\", \"genre\", \"year\" FROM %s ", c.MetadataTable)
		var rows *sql.Rows

		// Decide based on the format of the query if this is a special form.
		// The available fields are : "title", "album", "artist", "genre", "year"
		if startsWith(query, "songs named ") {
			// Title search
			q := query[len("songs named "):]
			selectWhereStatement := selectStatement + "WHERE title LIKE $1 ORDER BY rank"
			rows, err = db.Query(selectWhereStatement, "%"+q+"%")
			if err != nil {
				clog.Error("Search", "Database search failed.", err)
				return
			}
		} else if startsWith(query, "songs by ") {
			// Artist search
			q := query[len("songs by "):]
			selectWhereStatement := selectStatement + "WHERE artist LIKE $1 ORDER BY rank"
			rows, err = db.Query(selectWhereStatement, "%"+q+"%")
			if err != nil {
				clog.Error("Search", "Database search failed.", err)
				return
			}
		} else if endsWith(query, " songs") {
			// Genre search
			q := query[:len(query)-len(" songs")]
			selectWhereStatement := selectStatement + "WHERE genre LIKE $1 ORDER BY rank"
			rows, err = db.Query(selectWhereStatement, "%"+q+"%")
			if err != nil {
				clog.Error("Search", "Database search failed.", err)
				return
			}
		} else if startsWith(query, "songs from ") {
			// Joint year/album search
			// Note that the year query doesn't use includes: "Songs from 20" shouldn't return a song made in 2009.
			q := query[len("songs from "):]
			selectWhereStatement := selectStatement + "WHERE year LIKE $1 OR ALBUM LIKE $2 ORDER BY rank"
			rows, err = db.Query(selectWhereStatement, q, "%"+q+"%")
			if err != nil {
				clog.Error("Search", "Database search failed.", err)
				return
			}
		} else if startsWith(query, "songs released in ") {
			// Year search
			// This search also doesn't use an include-style parameter
			q := query[len("songs released in "):]
			selectWhereStatement := selectStatement + "WHERE year LIKE $1 ORDER BY rank"
			rows, err = db.Query(selectWhereStatement, q)
			if err != nil {
				clog.Error("Search", "Database search failed.", err)
				return
			}
		} else if startsWith(query, "songs in ") {
			// Album search
			q := query[len("songs in "):]
			selectWhereStatement := selectStatement + "WHERE album LIKE $1 ORDER BY rank"
			rows, err = db.Query(selectWhereStatement, "%"+q+"%")
			if err != nil {
				clog.Error("Search", "Database search failed.", err)
				return
			}
		} else {
			// After all that work, we've concluded we don't have a special form.
			// It's been an open question since before v3.0 what exactly we should do for a general search...
			// But, it's always been the case that either title or artist search works here.
			selectWhereStatement := selectStatement + "WHERE artist LIKE $1 OR title LIKE $2 ORDER BY rank"
			rows, err = db.Query(selectWhereStatement, "%"+query+"%", "%"+query+"%")
			if err != nil {
				clog.Error("Search", "Database search failed.", err)
				return
			}
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

		// Scan the returned data and save the relevant info
		clog.Debug("Search", "Scanning returned data...")
		var searchResults []SongData
		for rows.Next() {
			song := new(SongData)
			err := rows.Scan(&song.ID, &song.Artist, &song.Title, &song.Album, &song.Genre, &song.Year)
			if err != nil {
				clog.Error("Search", "Data scan failed.", err)
				return
			}
			// Add song (as SongData) to full searchResults
			searchResults = append(searchResults, SongData{ID: song.ID, Artist: song.Artist, Title: song.Title, Album: song.Album, Genre: song.Genre, Year: song.Year})

		}
		// Return data to client
		jsonMarshal, _ := json.Marshal(searchResults)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonMarshal)
	}
}

func RequestID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clog.Debug("Request", fmt.Sprintf("Decoding http-request data from client %s.", r.Header.Get("X-Forwarded-For")))
		requesterIP := r.Header.Get("X-Forwarded-For")

		// Declare object for a song
		type RequestResponse struct {
			Message       string
			TimeRemaining int
		}

		// Declare object to hold r body data
		type Request struct {
			ID    string `json:"ID"`
			Token string `json:"Token"`
		}
		var request Request

		// Decode json object
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			clog.Error("Request", fmt.Sprintf("Failed to read http-request body from %s.", r.Header.Get("X-Forwarded-For")), err)

			timeRemaining := 0
			message := "Request not completed. Request-body is possibly malformed."

			// Return data to client
			requestResponse := RequestResponse{message, timeRemaining}
			jsonMarshal, _ := json.Marshal(requestResponse)

			w.WriteHeader(http.StatusBadRequest) // 400 Bad Request
			w.Header().Set("Content-Type", "application/json")
			w.Write(jsonMarshal)
			return
		}
		err = json.Unmarshal(body, &request)
		if err != nil {
			clog.Error("Request", fmt.Sprintf("Failed to unmarshal http-request body from %s.", r.Header.Get("X-Forwarded-For")), err)

			timeRemaining := 0
			message := "Request not completed. Request-body is possibly malformed."

			// Return data to client
			requestResponse := RequestResponse{message, timeRemaining}
			jsonMarshal, _ := json.Marshal(requestResponse)

			w.WriteHeader(http.StatusBadRequest) // 400 Bad Request
			w.Header().Set("Content-Type", "application/json")
			w.Write(jsonMarshal)
			return
		}

		// Perform a check on the token
		var tokenValid bool
		if request.Token != "" {
			tokenValid = tokenCheck(request.Token)
		}

		if tokenValid {
			clog.Info("Request", fmt.Sprintf("Client %s bypassing rate limiter using token %s.", r.Header.Get("X-Forwarded-For"), request.Token))
		} else { // Perform check on timeout log in memory
			if _, ok := requestTimeoutIPs[requesterIP]; ok {
				// If the existing IP was recently logged, deny the request.
				if requestTimeoutIPs[requesterIP] > int(time.Now().Unix())-c.RequestRateLimit {
					clog.Info("Request", fmt.Sprintf("Request denied by rate limit for client %s.", r.Header.Get("X-Forwarded-For")))

					timeRemaining := requestTimeoutIPs[requesterIP] + c.RequestRateLimit - int(time.Now().Unix())
					message := fmt.Sprintf("Request denied. Client is rate-limited for %v seconds.", timeRemaining)

					// Return data to client
					requestResponse := RequestResponse{message, timeRemaining}
					jsonMarshal, _ := json.Marshal(requestResponse)

					w.WriteHeader(http.StatusTooManyRequests) // 429 Too Many Requests
					w.Header().Set("Content-Type", "application/json")
					w.Header().Set("Retry-After", strconv.Itoa(timeRemaining))
					w.Write(jsonMarshal)
					return
				}
			}
		}

		clog.Debug("Request", fmt.Sprintf("Received a song request for song ID #%v.", request.ID))
		clog.Debug("Request", "Searching database for corresponding path...")

		selectStatement := fmt.Sprintf("SELECT \"path\" FROM %s WHERE rowid=%v;", c.MetadataTable, request.ID)
		rows, err := db.Query(selectStatement)
		if err != nil {
			clog.Error("Request", "Database select failed.", err)
			timeRemaining := 0
			message := "Request not completed. Encountered a database error."

			// Return data to client
			requestResponse := RequestResponse{message, timeRemaining}
			jsonMarshal, _ := json.Marshal(requestResponse)

			w.WriteHeader(http.StatusInternalServerError) // 500 Server Error
			w.Header().Set("Content-Type", "application/json")
			w.Write(jsonMarshal)
			return
		}

		// "Every call to Scan, even the first one, must be preceded by a call to Next."
		var path string
		for rows.Next() {
			err := rows.Scan(&path)
			if err != nil {
				clog.Error("Request", "Data scan failed.", err)
				timeRemaining := 0
				message := "Request not completed. Encountered a database error."

				// Return data to client
				requestResponse := RequestResponse{message, timeRemaining}
				jsonMarshal, _ := json.Marshal(requestResponse)

				w.WriteHeader(http.StatusInternalServerError) // 500 Server Error
				w.Header().Set("Content-Type", "application/json")
				w.Write(jsonMarshal)
				return
			}
		}
		clog.Debug("Request", fmt.Sprintf("Translated ID %v to path: %s", request.ID, path))

		// Telnet to liquidsoap
		clog.Debug("Request", "Connecting to liquidsoap service...")
		conn, err := net.Dial("tcp", c.SourceAddress+c.SourcePort)
		if err != nil {
			clog.Error("Request", "Failed to connect to audio source server.", err)

			timeRemaining := 0
			message := "Request not completed. Could not submit request to stream source service."

			// Return data to client
			requestResponse := RequestResponse{message, timeRemaining}
			jsonMarshal, _ := json.Marshal(requestResponse)

			w.WriteHeader(http.StatusServiceUnavailable) // 503 Server Error
			w.Header().Set("Content-Type", "application/json")
			w.Write(jsonMarshal)
			return
		}

		// Push request over connection
		fmt.Fprintf(conn, "request.push "+path+"\n")
		// Listen for reply
		sourceServiceResponse, _ := bufio.NewReader(conn).ReadString('\n')
		clog.Debug("Request", fmt.Sprintf("Message from audio source server: %s", sourceServiceResponse))

		fmt.Fprintf(conn, "quit"+"\n")
		// Disconnect from liquidsoap
		conn.Close()

		// Create or overwrite existing log times if time and request body look OK
		requestTimeoutIPs[requesterIP] = int(time.Now().Unix())

		// Return 202 OK to client
		timeRemaining := requestTimeoutIPs[requesterIP] + c.RequestRateLimit - int(time.Now().Unix())
		message := "Request accepted!"

		// Return data to client
		requestResponse := RequestResponse{message, timeRemaining}
		jsonMarshal, _ := json.Marshal(requestResponse)

		w.WriteHeader(http.StatusAccepted) // 202 Accepted
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonMarshal)
	}
}

func RequestBestMatch() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clog.Debug("Search", fmt.Sprintf("Decoding http-request data from client %s.", r.Header.Get("X-Forwarded-For")))
		requesterIP := r.Header.Get("X-Forwarded-For")

		// Declare object to hold r body data and the corresponding ID
		type RequestBestMatch struct {
			Search string `json:"Search"`
			Token  string `json:"Token"`
			Path   string // ascertained later
		}
		var rbm RequestBestMatch

		// Decode json object
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&rbm)
		if err != nil {
			clog.Error("Search", fmt.Sprintf("Failed to read http-request body from %s.", r.Header.Get("X-Forwarded-For")), err)
			return
		}

		query := rbm.Search
		clog.Debug("Search", fmt.Sprintf("Search query decoded: '%v'", query))
		clog.Info("Search", fmt.Sprintf("Querying database for: '%v'", query))

		// Query database
		selectStatement := fmt.Sprintf("SELECT \"path\" FROM %s ", c.MetadataTable)
		var rows *sql.Rows

		// Decide based on the format of the query if this is a special form.
		// The available fields are : "title", "album", "artist", "genre", "year"
		if startsWith(query, "songs named ") {
			// Title search
			q := query[len("songs named "):]
			selectWhereStatement := selectStatement + "WHERE title LIKE $1 ORDER BY rank"
			rows, err = db.Query(selectWhereStatement, "%"+q+"%")
			if err != nil {
				clog.Error("Search", "Database search failed.", err)
				return
			}
		} else if startsWith(query, "songs by ") {
			// Artist search
			q := query[len("songs by "):]
			selectWhereStatement := selectStatement + "WHERE artist LIKE $1 ORDER BY rank"
			rows, err = db.Query(selectWhereStatement, "%"+q+"%")
			if err != nil {
				clog.Error("Search", "Database search failed.", err)
				return
			}
		} else if endsWith(query, " songs") {
			// Genre search
			q := query[:len(query)-len(" songs")]
			selectWhereStatement := selectStatement + "WHERE genre LIKE $1 ORDER BY rank"
			rows, err = db.Query(selectWhereStatement, "%"+q+"%")
			if err != nil {
				clog.Error("Search", "Database search failed.", err)
				return
			}
		} else if startsWith(query, "songs from ") {
			// Joint year/album search
			// Note that the year query doesn't use includes: "Songs from 20" shouldn't return a song made in 2009.
			q := query[len("songs from "):]
			selectWhereStatement := selectStatement + "WHERE year LIKE $1 OR ALBUM LIKE $2 ORDER BY rank"
			rows, err = db.Query(selectWhereStatement, q, "%"+q+"%")
			if err != nil {
				clog.Error("Search", "Database search failed.", err)
				return
			}
		} else if startsWith(query, "songs released in ") {
			// Year search
			// This search also doesn't use an include-style parameter
			q := query[len("songs released in "):]
			selectWhereStatement := selectStatement + "WHERE year LIKE $1 ORDER BY rank"
			rows, err = db.Query(selectWhereStatement, q)
			if err != nil {
				clog.Error("Search", "Database search failed.", err)
				return
			}
		} else if startsWith(query, "songs in ") {
			// Album search
			q := query[len("songs in "):]
			selectWhereStatement := selectStatement + "WHERE album LIKE $1 ORDER BY rank"
			rows, err = db.Query(selectWhereStatement, "%"+q+"%")
			if err != nil {
				clog.Error("Search", "Database search failed.", err)
				return
			}
		} else {
			// After all that work, we've concluded we don't have a special form.
			// It's been an open question since before v3.0 what exactly we should do for a general search...
			// But, it's always been the case that either title or artist search works here.
			selectWhereStatement := selectStatement + "WHERE artist LIKE $1 OR title LIKE $2 ORDER BY rank"
			rows, err = db.Query(selectWhereStatement, "%"+query+"%", "%"+query+"%")
			if err != nil {
				clog.Error("Search", "Database search failed.", err)
				return
			}
		}

		// Declare object for a song
		type RequestResponse struct {
			Message       string
			TimeRemaining int
		}

		// Scan the returned data and save the relevant info
		clog.Debug("Search", "Scanning returned data...")
		for rows.Next() {
			err := rows.Scan(&rbm.Path)
			if err != nil {
				clog.Error("Search", "Data scan failed.", err)
				return
			}
			break
		}

		// Perform a check on the token
		var tokenValid bool
		if rbm.Token != "" {
			tokenValid = tokenCheck(rbm.Token)
		}

		if tokenValid {
			clog.Info("Request", fmt.Sprintf("Client %s bypassing rate limiter using token %s.", r.Header.Get("X-Forwarded-For"), rbm.Token))
		} else { // Perform check on timeout log in memory
			if _, ok := requestTimeoutIPs[requesterIP]; ok {
				// If the existing IP was recently logged, deny the request.
				if requestTimeoutIPs[requesterIP] > int(time.Now().Unix())-c.RequestRateLimit {
					clog.Info("Request", fmt.Sprintf("Request denied by rate limit for client %s.", r.Header.Get("X-Forwarded-For")))

					timeRemaining := requestTimeoutIPs[requesterIP] + c.RequestRateLimit - int(time.Now().Unix())
					message := fmt.Sprintf("Request denied. Client is rate-limited for %v seconds.", timeRemaining)

					// Return data to client
					requestResponse := RequestResponse{message, timeRemaining}
					jsonMarshal, _ := json.Marshal(requestResponse)

					w.WriteHeader(http.StatusTooManyRequests) // 429 Too Many Requests
					w.Header().Set("Content-Type", "application/json")
					w.Header().Set("Retry-After", strconv.Itoa(timeRemaining))
					w.Write(jsonMarshal)
					return
				}
			}
		}

		// Telnet to liquidsoap
		clog.Debug("Request", "Connecting to liquidsoap service...")
		conn, err := net.Dial("tcp", c.SourceAddress+c.SourcePort)
		if err != nil {
			clog.Error("Request", "Failed to connect to audio source server.", err)

			timeRemaining := 0
			message := "Request not completed. Could not submit request to stream source service."

			// Return data to client
			requestResponse := RequestResponse{message, timeRemaining}
			jsonMarshal, _ := json.Marshal(requestResponse)

			w.WriteHeader(http.StatusServiceUnavailable) // 503 Server Error
			w.Header().Set("Content-Type", "application/json")
			w.Write(jsonMarshal)
			return
		}

		// Push request over connection
		fmt.Fprintf(conn, "request.push "+rbm.Path+"\n")
		// Listen for reply
		sourceServiceResponse, _ := bufio.NewReader(conn).ReadString('\n')
		clog.Debug("Request", fmt.Sprintf("Message from audio source server: %s", sourceServiceResponse))

		fmt.Fprintf(conn, "quit"+"\n")
		// Disconnect from liquidsoap
		conn.Close()

		// Create or overwrite existing log times if time and request body look OK
		requestTimeoutIPs[requesterIP] = int(time.Now().Unix())

		// Return 202 OK to client
		timeRemaining := requestTimeoutIPs[requesterIP] + c.RequestRateLimit - int(time.Now().Unix())
		message := "Request accepted!"

		// Return data to client
		requestResponse := RequestResponse{message, timeRemaining}
		jsonMarshal, _ := json.Marshal(requestResponse)

		w.WriteHeader(http.StatusAccepted) // 202 Accepted
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonMarshal)
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
		w.WriteHeader(http.StatusAccepted) // 202 Accepted
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonMarshal)
	}
}

// Status Checks ////////////////////////////////////////////////////////////////////
func Ready() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted) // 202 Accepted
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
