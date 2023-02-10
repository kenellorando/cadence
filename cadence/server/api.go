// handlers.go
// API functions and fileservers

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/dhowden/tag"
	"github.com/kenellorando/clog"
)

// /api/search
// Gets text metadata (excludes album art and path) of any songs matching a search query.
func Search() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clog.Info("Search", fmt.Sprintf("Search request from client %s.", r.Header.Get("X-Forwarded-For")))

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
		w.Write(jsonMarshal)
	}
}

// /api/request/id
// Posts a request submission for a specific song ID.
func RequestID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clog.Info("Request", fmt.Sprintf("Request-by-ID by client %s.", r.Header.Get("X-Forwarded-For")))

		type Request struct {
			ID string `json:"ID"`
		}
		var request Request

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&request)
		if err != nil {
			clog.Error("RequestID", "Unable to decode request.", err)
			w.WriteHeader(http.StatusBadRequest) // 400 Bad Request
			return
		}
		reqID, err := strconv.Atoi(request.ID)
		if err != nil {
			clog.Error("RequestID", "Unable to convert request ID to an integer.", err)
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}
		path, err := getPathById(reqID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}
		_, err = liquidsoapRequest(path)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}

		w.WriteHeader(http.StatusAccepted) // 202 Accepted
	}
}

// /api/request/bestmatch
// Posts a request submission for the top result of a search.
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
		_, err = liquidsoapRequest(path)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}

		w.WriteHeader(http.StatusAccepted) // 202 Accepted
	}
}

// /api/nowplaying/metadata
// Gets text metadata (excludes album art and path) of the currently playing song.
func NowPlayingMetadata() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		queryResults, err := searchByTitleArtist(now.title, now.artist)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}
		if len(queryResults) < 1 {
			w.WriteHeader(http.StatusNotFound) // 404
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
		w.Write(jsonMarshal)
	}
}

// /api/nowplaying/albumart
// Gets encoded album art of the currently playing song.
func NowPlayingAlbumArt() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		queryResults, err := searchByTitleArtist(now.title, now.artist)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}
		if len(queryResults) < 1 {
			w.WriteHeader(http.StatusNotFound) // 404
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
		if tags.Picture().Data == nil {
			w.WriteHeader(http.StatusNotFound) // 404
			return
		}

		// Return data to client
		type SongData struct {
			Picture []byte
		}
		result := SongData{Picture: tags.Picture().Data}
		jsonMarshal, _ := json.Marshal(result)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonMarshal)
	}
}

// /api/listenurl
// Gets the direct stream listen URL, which is a combination of host and mountpoint, set in Icecast's cadence.xml.
func ListenURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Return data to client
		type ListenURL struct {
			ListenURL string
		}
		listenurl := ListenURL{ListenURL: string(now.host + "/" + now.mountpoint)}
		jsonMarshal, _ := json.Marshal(listenurl)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonMarshal)
	}
}

// /api/listeners
// Gets the number of active connections with Icecast.
func Listeners() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Return data to client
		type Listeners struct {
			Listeners int
		}
		listeners := Listeners{Listeners: int(now.listeners)}
		jsonMarshal, _ := json.Marshal(listeners)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonMarshal)
	}
}

// /api/version
// Gets the current server version (set in cadence.env).
func Version() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Return data to client
		type CadenceVersion struct {
			Version string
		}
		version := CadenceVersion{Version: c.Version}
		jsonMarshal, _ := json.Marshal(version)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonMarshal)
	}
}

// /ready
// Gets 200 OK status.
func Ready() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK) // 200 OK
	}
}
