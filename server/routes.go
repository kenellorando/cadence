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
		ID     int
		Artist string
		Title  string
	}
	// Run the search query on the database
	rows, err := database.Query(selectWhereStatement)
	if err != nil {
		clog.Error("ARIA1Search", "Database search failed.", err)
		return
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

// ARIA1Request - song requester
func ARIA1Request(w http.ResponseWriter, r *http.Request) {
	clog.Debug("ARIA1Request", fmt.Sprintf("Decoding http-request data from client %s.", r.RemoteAddr))

	// Declare object to hold r body data
	type Request struct {
		ID string `json:"ID"`
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

	clog.Debug("ARIA1Request", fmt.Sprintf("Received a song request for song ID #%v.", request.ID))

	fmt.Printf("%v", request.ID)
	selectStatement := fmt.Sprintf("SELECT \"path\" FROM %s WHERE id=%v;", c.schema.Table, request.ID)
	rows, err := database.Query(selectStatement)
	if err != nil {
		clog.Error("ARIA1Search", "Database select failed.", err)
		return
	}

	// "Every call to Scan, even the first one, must be preceded by a call to Next."
	var path string
	for rows.Next() {
		err := rows.Scan(&path)
		if err != nil {
			clog.Error("ARIA1Search", "Data scan failed.", err)
			return
		}
	}
	fmt.Printf("Received this path: %s", path)

	// Telnet to liquidsoap
	clog.Info("ARIA1Request", "Connecting to liquidsoap service...")

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
