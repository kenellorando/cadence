// handlers.go
// HTTP request function returns

package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"path"
	"strconv"
	"time"

	"github.com/Jeffail/gabs"
	"github.com/kenellorando/clog"
)

// Map of requester IPs to be locked out of making requests
var requestTimeoutIPs = make(map[string]int)

// Fileservers ////////////////////////////////////////////////////////////////////
func handleServeRoot() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clog.Info("ServeRoot", fmt.Sprintf("Client %s requesting %s%s", r.Header.Get("X-Forwarded-For"), r.Host, r.URL.Path))
		w.Header().Set("Content-type", "text/html")
		http.ServeFile(w, r, path.Dir("./public/index.html"))
	}
}

func handleServe404() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clog.Info("Serve404", fmt.Sprintf("Client %s requesting unknown resource %s%s. Returning 404.", r.Header.Get("X-Forwarded-For"), r.Host, r.URL.Path))
		w.Header().Set("Content-type", "text/html")
		http.ServeFile(w, r, path.Dir("./public/404/index.html"))
	}
}

// ARIA1 API ////////////////////////////////////////////////////////////////////
func handleARIA1Search() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clog.Debug("ARIA1Search", fmt.Sprintf("Decoding http-request data from client %s.", r.Header.Get("X-Forwarded-For")))
		// Declare object to hold r body data
		type Search struct {
			Query string `json:"search"`
		}
		var search Search

		// Decode json object
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&search)
		if err != nil {
			clog.Error("ARIA1Search", fmt.Sprintf("Failed to read http-request body from %s.", r.Header.Get("X-Forwarded-For")), err)
			return
		}

		query := search.Query
		clog.Debug("ARIA1Search", fmt.Sprintf("Search query decoded: '%v'", query))
		clog.Info("ARIA1Search", fmt.Sprintf("Querying database for: '%v'", query))

		// Query database
		selectStatement := fmt.Sprintf("SELECT \"id\", \"artist\", \"title\" FROM %s ", c.schema.Table)
		var rows *sql.Rows

		// Decide based on the format of the query if this is a special form.
		// The available fields are : "title", "album", "artist", "genre", "year"
		if startsWith(query, "songs named ") {
			// Title search
			q := query[len("songs named "):]
			selectWhereStatement := selectStatement + "WHERE title ILIKE $1 ORDER BY levenshtein($2, title) ASC"
			rows, err = database.Query(selectWhereStatement, "%"+q+"%", q)
			if err != nil {
				clog.Error("ARIA1Search", "Database search failed.", err)
				return
			}
		} else if startsWith(query, "songs by ") {
			// Artist search
			q := query[len("songs by "):]
			selectWhereStatement := selectStatement + "WHERE artist ILIKE $1 ORDER BY levenshtein($2, artist) ASC"
			rows, err = database.Query(selectWhereStatement, "%"+q+"%", q)
			if err != nil {
				clog.Error("ARIA1Search", "Database search failed.", err)
				return
			}
		} else if endsWith(query, " songs") {
			// Genre search
			q := query[:len(query)-len(" songs")]
			selectWhereStatement := selectStatement + "WHERE genre ILIKE $1 ORDER BY levenshtein($2, genre) ASC"
			rows, err = database.Query(selectWhereStatement, "%"+q+"%", q)
			if err != nil {
				clog.Error("ARIA1Search", "Database search failed.", err)
				return
			}
		} else if startsWith(query, "songs from ") {
			// Joint year/album search
			// Note that the year query doesn't use includes: "Songs from 20" shouldn't return a song made in 2009.
			q := query[len("songs from "):]
			selectWhereStatement := selectStatement + "WHERE year LIKE $1 OR ALBUM ILIKE $2 ORDER BY LEAST(levenshtein($3, year), levenshtein($4, album)) ASC"
			rows, err = database.Query(selectWhereStatement, q, "%"+q+"%", q, q)
			if err != nil {
				clog.Error("ARIA1Search", "Database search failed.", err)
				return
			}
		} else if startsWith(query, "songs released in ") {
			// Year search
			// This search also doesn't use an include-style parameter
			q := query[len("songs released in "):]
			selectWhereStatement := selectStatement + "WHERE year LIKE $1 ORDER BY levenshtein($2, year) ASC"
			rows, err = database.Query(selectWhereStatement, q, q)
			if err != nil {
				clog.Error("ARIA1Search", "Database search failed.", err)
				return
			}
		} else if startsWith(query, "songs in ") {
			// Album search
			q := query[len("songs in "):]
			selectWhereStatement := selectStatement + "WHERE album ILIKE $1 ORDER BY levenshtein($2, album) ASC"
			rows, err = database.Query(selectWhereStatement, "%"+q+"%", q)
			if err != nil {
				clog.Error("ARIA1Search", "Database search failed.", err)
				return
			}
		} else {
			// After all that work, we've concluded we don't have a special form.
			// It's been an open question since before v3.0 what exactly we should do for a general search...
			// But, it's always been the case that either title or artist search works here.
			selectWhereStatement := selectStatement + "WHERE artist ILIKE $1 OR title ILIKE $2 ORDER BY LEAST(levenshtein($3, artist), levenshtein($4, title)) ASC"
			rows, err = database.Query(selectWhereStatement, "%"+query+"%", "%"+query+"%", query, query)
			if err != nil {
				clog.Error("ARIA1Search", "Database search failed.", err)
				return
			}
		}

		// Declare object for a song
		type SongData struct {
			ID     int
			Artist string
			Title  string
		}

		// Scan the returned data and save the relevant info
		clog.Debug("ARIA1Search", "Scanning returned data...")
		var searchResults []SongData
		for rows.Next() {
			song := new(SongData)
			err := rows.Scan(&song.ID, &song.Artist, &song.Title)
			if err != nil {
				clog.Error("ARIA1Search", "Data scan failed.", err)
				return
			}
			// Add song (as SongData) to full searchResults
			searchResults = append(searchResults, SongData{ID: song.ID, Artist: song.Artist, Title: song.Title})

		}
		// Return data to client
		jsonMarshal, _ := json.Marshal(searchResults)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonMarshal)
	}
}

func handleARIA1Request() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clog.Debug("ARIA1Request", fmt.Sprintf("Decoding http-request data from client %s.", r.Header.Get("X-Forwarded-For")))

		requesterIP := r.Header.Get("X-Forwarded-For")

		// Declare object for a song
		type RequestResponse struct {
			Message       string
			TimeRemaining int
		}

		// If the IP is in the timeout log

		if _, ok := requestTimeoutIPs[requesterIP]; ok {
			// If the existing IP was recently logged, deny the request.
			if requestTimeoutIPs[requesterIP] > int(time.Now().Unix())-180 {
				clog.Info("ARIA1Request", fmt.Sprintf("Request denied by rate limit for client %s.", r.Header.Get("X-Forwarded-For")))

				timeRemaining := requestTimeoutIPs[requesterIP] + 180 - int(time.Now().Unix())
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

		// Declare object to hold r body data
		type Request struct {
			ID string `json:"ID"`
		}
		var request Request

		// Decode json object
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			clog.Error("ARIA1Request", fmt.Sprintf("Failed to read http-request body from %s.", r.Header.Get("X-Forwarded-For")), err)

			timeRemaining := 0
			message := fmt.Sprintf("Request not completed. Request-body is possibly malformed.")

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
			clog.Error("ARIA1Request", fmt.Sprintf("Failed to unmarshal http-request body from %s.", r.Header.Get("X-Forwarded-For")), err)

			timeRemaining := 0
			message := fmt.Sprintf("Request not completed. Request-body is possibly malformed.")

			// Return data to client
			requestResponse := RequestResponse{message, timeRemaining}
			jsonMarshal, _ := json.Marshal(requestResponse)

			w.WriteHeader(http.StatusBadRequest) // 400 Bad Request
			w.Header().Set("Content-Type", "application/json")
			w.Write(jsonMarshal)
			return
		}

		clog.Debug("ARIA1Request", fmt.Sprintf("Received a song request for song ID #%v.", request.ID))
		clog.Debug("ARIA1Request", "Searching database for corresponding path...")

		selectStatement := fmt.Sprintf("SELECT \"path\" FROM %s WHERE id=%v;", c.schema.Table, request.ID)
		rows, err := database.Query(selectStatement)
		if err != nil {
			clog.Error("ARIA1Request", "Database select failed.", err)
			timeRemaining := 0
			message := fmt.Sprintf("Request not completed. Encountered a database error.")

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
				clog.Error("ARIA1Request", "Data scan failed.", err)
				timeRemaining := 0
				message := fmt.Sprintf("Request not completed. Encountered a database error.")

				// Return data to client
				requestResponse := RequestResponse{message, timeRemaining}
				jsonMarshal, _ := json.Marshal(requestResponse)

				w.WriteHeader(http.StatusInternalServerError) // 500 Server Error
				w.Header().Set("Content-Type", "application/json")
				w.Write(jsonMarshal)
				return
			}
		}
		clog.Debug("ARIA1Request", fmt.Sprintf("Translated ID %v to path: %s", request.ID, path))

		// Telnet to liquidsoap
		clog.Debug("ARIA1Request", "Connecting to liquidsoap service...")
		conn, err := net.Dial("tcp", c.server.SourceAddress)
		if err != nil {
			clog.Error("ARIA1Request", "Failed to connect to audio source server.", err)

			timeRemaining := 0
			message := fmt.Sprintf("Request not completed. Could not submit request to stream source service.")

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
		clog.Debug("ARIA1Request", fmt.Sprintf("Message from audio source server: %s", sourceServiceResponse))

		// Disconnect from liquidsoap
		conn.Close()

		// Create or overwrite existing log times if time and request body look OK
		requestTimeoutIPs[requesterIP] = int(time.Now().Unix())

		// Return 202 OK to client
		timeRemaining := requestTimeoutIPs[requesterIP] + 180 - int(time.Now().Unix())
		message := fmt.Sprintf("Request accepted!")

		// Return data to client
		requestResponse := RequestResponse{message, timeRemaining}
		jsonMarshal, _ := json.Marshal(requestResponse)

		w.WriteHeader(http.StatusAccepted) // 202 Accepted
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonMarshal)
		return
	}
}

func handleARIA1Library() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clog.Info("ServeLibrary", fmt.Sprintf("Client %s requesting %s%s", r.Header.Get("X-Forwarded-For"), r.Host, r.URL.Path))
		// Open the file, marshall the data and write it
		fileReader, _ := ioutil.ReadFile(c.server.RootPath + "./public/library.json")
		rawJSON := json.RawMessage(string(fileReader))
		jsonMarshal, _ := json.Marshal(rawJSON)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonMarshal)
	}
}

func handleARIA1NowPlaying() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clog.Debug("ARIA1NowPlaying", fmt.Sprintf("Client %s requesting %s%s", r.Header.Get("X-Forwarded-For"), r.Host, r.URL.Path))

		resp, err := http.Get("http://icecast2:8000/status-json.xsl")
		if err != nil {
			clog.Error("ARIA1NowPlaying", "Failed to connect to audio stream server.", err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			clog.Error("ARIA1NowPlaying", "Audio stream server returned bad status", err)
			return
		}

		body, _ := io.ReadAll(resp.Body)
		jsonParsed, _ := gabs.ParseJSON([]byte(body))

		var artist, _ = jsonParsed.Path("icestats.source.artist").Data().(string)
		var title, _ = jsonParsed.Path("icestats.source.title").Data().(string)

		clog.Info("ARIA1NowPlaying", fmt.Sprintf("Now playing: '%s' by '%s'.", title, artist))

		// Return data to client
		type NowPlayingResponse struct {
			Artist string
			Title  string
		}
		nowPlayingResponse := NowPlayingResponse{artist, title}
		jsonMarshal, _ := json.Marshal(nowPlayingResponse)
		w.WriteHeader(http.StatusAccepted) // 202 Accepted
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonMarshal)
		return
	}
}

// ARIA2 API ////////////////////////////////////////////////////////////////////
func handleARIA2Request() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clog.Debug("ARIA2Request", fmt.Sprintf("Decoding http-request data from client %s.", r.Header.Get("X-Forwarded-For")))
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
			clog.Error("ARIA1Request", fmt.Sprintf("Failed to read http-request body from %s.", r.Header.Get("X-Forwarded-For")), err)

			timeRemaining := 0
			message := fmt.Sprintf("Request not completed. Request-body is possibly malformed.")

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
			clog.Error("ARIA2Request", fmt.Sprintf("Failed to unmarshal http-request body from %s.", r.Header.Get("X-Forwarded-For")), err)

			timeRemaining := 0
			message := fmt.Sprintf("Request not completed. Request-body is possibly malformed.")

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

		if tokenValid == true {
			clog.Info("ARIA2Request", fmt.Sprintf("Client %s bypassing rate limiter using token %s.", r.Header.Get("X-Forwarded-For"), request.Token))
		} else { // Perform check on timeout log in memory
			if _, ok := requestTimeoutIPs[requesterIP]; ok {
				// If the existing IP was recently logged, deny the request.
				if requestTimeoutIPs[requesterIP] > int(time.Now().Unix())-180 {
					clog.Info("ARIA2Request", fmt.Sprintf("Request denied by rate limit for client %s.", r.Header.Get("X-Forwarded-For")))

					timeRemaining := requestTimeoutIPs[requesterIP] + 180 - int(time.Now().Unix())
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

		clog.Debug("ARIA2Request", fmt.Sprintf("Received a song request for song ID #%v.", request.ID))
		clog.Debug("ARIA2Request", "Searching database for corresponding path...")

		selectStatement := fmt.Sprintf("SELECT \"path\" FROM %s WHERE id=%v;", c.schema.Table, request.ID)
		rows, err := database.Query(selectStatement)
		if err != nil {
			clog.Error("ARIA2Request", "Database select failed.", err)
			timeRemaining := 0
			message := fmt.Sprintf("Request not completed. Encountered a database error.")

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
				clog.Error("ARIA2Request", "Data scan failed.", err)
				timeRemaining := 0
				message := fmt.Sprintf("Request not completed. Encountered a database error.")

				// Return data to client
				requestResponse := RequestResponse{message, timeRemaining}
				jsonMarshal, _ := json.Marshal(requestResponse)

				w.WriteHeader(http.StatusInternalServerError) // 500 Server Error
				w.Header().Set("Content-Type", "application/json")
				w.Write(jsonMarshal)
				return
			}
		}
		clog.Debug("ARIA2Request", fmt.Sprintf("Translated ID %v to path: %s", request.ID, path))

		// Telnet to liquidsoap
		clog.Debug("ARIA2Request", "Connecting to liquidsoap service...")
		conn, err := net.Dial("tcp", c.server.SourceAddress)
		if err != nil {
			clog.Error("ARIA2Request", "Failed to connect to audio source server.", err)

			timeRemaining := 0
			message := fmt.Sprintf("Request not completed. Could not submit request to stream source service.")

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
		clog.Debug("ARIA2Request", fmt.Sprintf("Message from audio source server: %s", sourceServiceResponse))

		// Disconnect from liquidsoap
		conn.Close()

		// Create or overwrite existing log times if time and request body look OK
		requestTimeoutIPs[requesterIP] = int(time.Now().Unix())

		// Return 202 OK to client
		timeRemaining := requestTimeoutIPs[requesterIP] + 180 - int(time.Now().Unix())
		message := fmt.Sprintf("Request accepted!")

		// Return data to client
		requestResponse := RequestResponse{message, timeRemaining}
		jsonMarshal, _ := json.Marshal(requestResponse)

		w.WriteHeader(http.StatusAccepted) // 202 Accepted
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonMarshal)
		return
	}
}
