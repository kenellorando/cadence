package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"

	"github.com/kenellorando/clog"
)

// ServeRoot - serves the frontend root index page
func ServeRoot(w http.ResponseWriter, r *http.Request) {
	clog.Info("ServeRoot", fmt.Sprintf("Client %s requesting %s%s", r.RemoteAddr, r.Host, r.URL.Path))
	w.Header().Set("Content-type", "text/html")
	http.ServeFile(w, r, path.Dir("./public/index.html"))
}

// Serve404 - served for any requests to unknown resources
func Serve404(w http.ResponseWriter, r *http.Request) {
	clog.Info("Serve404", fmt.Sprintf("Client %s requesting unknown resource %s%s. Returning 404.", r.RemoteAddr, r.Host, r.URL.Path))
	w.Header().Set("Content-type", "text/html")
	http.ServeFile(w, r, path.Dir("./public/404/index.html"))
}

// ARIA1Search - database song searcher
// Request is received as a value in a raw JSON with key 'search'
func ARIA1Search(w http.ResponseWriter, r *http.Request) {
	clog.Debug("ARIA1Search", fmt.Sprintf("Decoding http-request data from client %s.", r.RemoteAddr))
	// Declare object to hold r body data
	type Search struct {
		Query string `json:"search"`
	}
	var search Search

	// Decode json object
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&search)
	if err != nil {
		clog.Error("ARIA1Search", fmt.Sprintf("Failed to read http-request body from %s.", r.RemoteAddr), err)
		return
	}

	query := search.Query
	clog.Debug("ARIA1Search", fmt.Sprintf("Search query decoded: '%v'", query))
	clog.Info("ARIA1Search", fmt.Sprintf("Querying database for: '%v'", query))

	// Query database
	selectStatement := fmt.Sprintf("SELECT \"id\", \"artist\", \"title\" FROM %s ", c.schema.Table)
	selectWhereStatement := fmt.Sprintf(selectStatement+"WHERE title LIKE '%%%s%%' OR artist LIKE '%%%s%%'", query, query)
	// Declare object for a song
	type SongData struct {
		id     int
		artist string
		title  string
	}
	// Run the search query on the database
	rows, err := database.Query(selectWhereStatement)
	if err != nil {
		clog.Error("ARIA1Search", "Database search failed.", err)
	}
	// Scan the returned data and save the relevant info
	clog.Debug("ARIA1Search", "Scanning returned data...")
	var searchResults []SongData
	for rows.Next() {
		song := new(SongData)
		err := rows.Scan(&song.id, &song.artist, &song.title)
		if err != nil {
			clog.Error("ARIA1Search", "Data scan failed.", err)
			return
		}
		// Add song (as SongData) to full searchResults
		searchResults = append(searchResults, SongData{id: song.id, artist: song.artist, title: song.title})

	}
	fmt.Printf("%v", searchResults)

	// Return data to client
	jsonMarshal, _ := json.Marshal(searchResults)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonMarshal)
}

// ARIA1Request - song requester
func ARIA1Request(w http.ResponseWriter, r *http.Request) {
	clog.Debug("ARIA1Request", fmt.Sprintf("Decoding http-request data from client %s.", r.RemoteAddr))

	// Declare object to hold r body data
	type Request struct {
		Path string `json:"path"`
	}
	var request Request

	// Decode json object
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		clog.Error("ARIA1Request", fmt.Sprintf("Failed to read http-request body from %s.", r.RemoteAddr), err)
		return
	}
	err = json.Unmarshal(body, &request)
	if err != nil {
		clog.Error("ARIA1Request", fmt.Sprintf("Failed to unmarshal http-request body from %s.", r.RemoteAddr), err)
		return
	}

	path := request.Path
	clog.Debug("ARIA1Request", fmt.Sprintf("Song request decoded: '%v'", path))
	clog.Info("ARIA1Request", "Connecting to liquidsoap service...")
	// Telnet to liquidsoap
	// Forward path in a request command
	// Disconnect from liquidsoap
}

// ARIA1Library - serves the library json file
func ARIA1Library(w http.ResponseWriter, r *http.Request) {
	clog.Info("ServeLibrary", fmt.Sprintf("Client %s requesting %s%s", r.RemoteAddr, r.Host, r.URL.Path))
	// Open the file, marshall the data and write it
	fileReader, _ := ioutil.ReadFile(c.server.RootPath + "./public/library.json")
	rawJSON := json.RawMessage(string(fileReader))
	jsonMarshal, _ := json.Marshal(rawJSON)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonMarshal)
}
