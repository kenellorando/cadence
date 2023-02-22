// handlers.go
// API functions and fileservers.
// The functions are named exactly as their API paths are.
// See complete API documentation: https://github.com/kenellorando/cadence/wiki/API-Reference

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

// POST /api/search
// Receives a search query, which it looks in the database for.
// Returns a JSON list of text metadata (excluding art and path) of any matching songs.
func Search() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clog.Debug("Search", fmt.Sprintf("Search request from client %s.", r.RemoteAddr))
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
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		jsonMarshal, _ := json.Marshal(queryResults)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonMarshal)
	}
}

// POST /api/request/id
// Receives an integer ID of a song to request.
// This ID is translated to a filesystem path, which is passed to Liquidsoap for processing.
func RequestID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clog.Info("Request", fmt.Sprintf("Request-by-ID by client %s.", r.RemoteAddr))
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
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		path, err := getPathById(reqID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		_, err = liquidsoapRequest(path)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusAccepted) // 202 Accepted
	}
}

// POST /api/request/bestmatch
// Receives a search query, which it looks in the database for.
// The number one result of the search has its path taken and submitted to Liquidsoap for processing.
func RequestBestMatch() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clog.Debug("Search", fmt.Sprintf("Decoding http-request data from client %s.", r.RemoteAddr))
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
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		path, err := getPathById(queryResults[0].ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		_, err = liquidsoapRequest(path)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusAccepted) // 202 Accepted
	}
}

// /api/nowplaying/metadata
// Gets text metadata (excludes album art and path) of the currently playing song.
func NowPlayingMetadata() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		queryResult, err := searchByTitleArtist(now.Song.Title, now.Song.Artist)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		jsonMarshal, _ := json.Marshal(queryResult[0])
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonMarshal)
	}
}

// GET /api/nowplaying/albumart
// Gets base64 encoded album art of the currently playing song.
func NowPlayingAlbumArt() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		queryResults, err := searchByTitleArtist(now.Song.Title, now.Song.Artist)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		if len(queryResults) < 1 {
			w.WriteHeader(http.StatusNotFound) // 404 Not Found
			return
		}
		path, err := getPathById(queryResults[0].ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		file, err := os.Open(path)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		tags, err := tag.ReadFrom(file)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		if tags.Picture() == nil {
			w.WriteHeader(http.StatusNoContent) // 204 No Content
			return
		}
		type SongData struct {
			Picture []byte
		}
		result := SongData{Picture: tags.Picture().Data}
		jsonMarshal, _ := json.Marshal(result)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonMarshal)
	}
}

// GET /api/history
// Gets a list of the ten last-played songs, noting the time each ended.
func History() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jsonMarshal, _ := json.Marshal(history)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonMarshal)
	}
}

// GET /api/listenurl
// Gets the direct stream listen URL, which is a combination of host and mountpoint, set by Icecast's cadence.xml.
func ListenURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type ListenURL struct {
			ListenURL string
		}
		listenurl := ListenURL{ListenURL: string(now.Host + "/" + now.Mountpoint)}
		jsonMarshal, _ := json.Marshal(listenurl)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonMarshal)
	}
}

// GET /api/listeners
// Gets the number of active connections to Icecast's stream.
func Listeners() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type Listeners struct {
			Listeners int
		}
		listeners := Listeners{Listeners: int(now.Listeners)}
		jsonMarshal, _ := json.Marshal(listeners)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonMarshal)
	}
}

// GET /api/bitrate
// Gets the audio stream bitrate in kilobits.
func Bitrate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type Bitrate struct {
			Bitrate int
		}
		bitrate := Bitrate{Bitrate: int(now.Bitrate)}
		jsonMarshal, _ := json.Marshal(bitrate)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonMarshal)
	}
}

// GET /api/version
// Gets the current server version (set in cadence.env).
func Version() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type Version struct {
			Version string
		}
		version := Version{Version: c.Version}
		jsonMarshal, _ := json.Marshal(version)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonMarshal)
	}
}

// GET /ready
// Gets 200 OK status. Primarily used for verifying health/readiness of the API.
func Ready() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK) // 200 OK
	}
}

// GET /api/dev/skip
// Requires development mode enabled.
// Forwards a request to Liquidsoap to skip the currently playing track.
func DevSkip() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := liquidsoapSkip()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusOK) // 200 OK
	}
}
