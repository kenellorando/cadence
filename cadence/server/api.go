// handlers.go
// API functions.
// The functions are named exactly the same as their API paths.
// See complete API documentation: https://github.com/kenellorando/cadence/wiki/API-Reference

package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/dhowden/tag"
)

// POST /api/search
// Receives a search query, which it looks in the database for.
// Returns a JSON list of text metadata (excluding art and path) of any matching songs.
func Search() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Debug(fmt.Sprintf("Search request from client %s.", r.RemoteAddr), "func", "Search")
		type Search struct {
			Query string `json:"search"`
		}
		var search Search
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&search)
		if err != nil {
			slog.Error("Unable to decode search body.", "func", "Search", "error", err)
			w.WriteHeader(http.StatusBadRequest) // 400 Bad Request
			return
		}
		queryResults, err := searchByQuery(search.Query)
		if err != nil {
			slog.Error("Unable to execute search by query.", "func", "Search", "error", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		jsonMarshal, err := json.Marshal(queryResults)
		if err != nil {
			slog.Error("Failed to marshal results from the search.", "func", "Search", "error", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(jsonMarshal)
		if err != nil {
			slog.Error("Failed to write response.", "func", "Search", "error", err)
			return
		}
	}
}

// POST /api/request/id
// Receives an integer ID of a song to request.
// This ID is translated to a filesystem path, which is passed to Liquidsoap for processing.
func RequestID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info(fmt.Sprintf("Request-by-ID by client %s.", r.RemoteAddr), "func", "Request")
		type Request struct {
			ID string `json:"ID"`
		}
		var request Request
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&request)
		if err != nil {
			slog.Error("Unable to decode request.", "func", "RequestID", "error", err)
			w.WriteHeader(http.StatusBadRequest) // 400 Bad Request
			return
		}
		reqID, err := strconv.Atoi(request.ID)
		if err != nil {
			slog.Error("Unable to convert request ID to an integer.", "func", "RequestID", "error", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		path, err := getPathById(reqID)
		if err != nil {
			slog.Error("Unable to find file path by song ID.", "func", "RequestID", "error", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		_, err = liquidsoapRequest(path)
		if err != nil {
			slog.Error("Unable to submit song request.", "func", "RequestID", "error", err)
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
		slog.Debug(fmt.Sprintf("Decoding http-request data from client %s.", r.RemoteAddr), "func", "Search")
		type RequestBestMatch struct {
			Query string `json:"Search"`
		}
		var rbm RequestBestMatch
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&rbm)
		if err != nil {
			slog.Error("Unable to decode request body.", "func", "RequestBestMatch", "error", err)
			w.WriteHeader(http.StatusBadRequest) // 400 Bad Request
			return
		}
		queryResults, err := searchByQuery(rbm.Query)
		if err != nil {
			slog.Error("Unable to search by query.", "func", "RequestBestMatch", "error", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		path, err := getPathById(queryResults[0].ID)
		if err != nil {
			slog.Error("Unable to find file path by song ID", "func", "RequestBestMatch", "error", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		_, err = liquidsoapRequest(path)
		if err != nil {
			slog.Error("Unable to submit song request.", "func", "RequestBestMatch", "error", err)
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
		queryResults, err := searchByTitleArtist(now.Song.Title, now.Song.Artist)
		if err != nil {
			slog.Error("Unable to search by title and artist.", "func", "NowPlayingMetadata", "error", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		if len(queryResults) < 1 {
			slog.Warn("The currently playing song could not be found in the database. The database may not be populated.", "func", "NowPlayingMetadata")
			w.WriteHeader(http.StatusNotFound) // 404 Not Found
			return
		}
		jsonMarshal, err := json.Marshal(queryResults[0])
		if err != nil {
			slog.Error("Failed to marshal results from the search.", "func", "NowPlayingMetadata", "error", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(jsonMarshal)
		if err != nil {
			slog.Error("Failed to write response.", "func", "NowPlayingMetadata", "error", err)
			return
		}
	}
}

// GET /api/nowplaying/albumart
// Gets base64 encoded album art of the currently playing song.
func NowPlayingAlbumArt() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		queryResults, err := searchByTitleArtist(now.Song.Title, now.Song.Artist)
		if err != nil {
			slog.Error("Unable to search by title and artist.", "func", "NowPlayingAlbumArt", "error", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		if len(queryResults) < 1 {
			slog.Warn("The currently playing song could not be found in the database. The database may not be populated.", "func", "NowPlayingAlbumArt")
			w.WriteHeader(http.StatusNotFound) // 404 Not Found
			return
		}
		path, err := getPathById(queryResults[0].ID)
		if err != nil {
			slog.Error("Unable to find file path by song ID.", "func", "NowPlayingAlbumArt", "error", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		file, err := os.Open(path)
		if err != nil {
			slog.Error("Unable to open a file for album art extraction.", "func", "NowPlayingAlbumArt", "error", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		tags, err := tag.ReadFrom(file)
		if err != nil {
			slog.Error("Unable to read tags on file for art extraction.", "func", "NowPlayingAlbumArt", "error", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		if tags.Picture() == nil {
			slog.Debug("The currently playing song has no album art metadata.", "func", "NowPlayingAlbumArt")
			w.WriteHeader(http.StatusNoContent) // 204 No Content
			return
		}
		type SongData struct {
			Picture []byte
		}
		result := SongData{Picture: tags.Picture().Data}
		jsonMarshal, err := json.Marshal(result)
		if err != nil {
			slog.Error("Failed to marshal art data.", "func", "NowPlayingAlbumArt", "error", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(jsonMarshal)
		if err != nil {
			slog.Error("Failed to write response.", "func", "NowPlayingAlbumArt", "error", err)
			return
		}
	}
}

// GET /api/history
// Gets a list of the ten last-played songs, noting the time each ended.
func History() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jsonMarshal, err := json.Marshal(history)
		if err != nil {
			slog.Error("Failed to marshal play history.", "func", "History", "error", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(jsonMarshal)
		if err != nil {
			slog.Error("Failed to write response.", "func", "History", "error", err)
			return
		}
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
		jsonMarshal, err := json.Marshal(listenurl)
		if err != nil {
			slog.Error("Failed to marshal listen URL.", "func", "ListenURL", "error", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(jsonMarshal)
		if err != nil {
			slog.Error("Failed to write response.", "func", "ListenURL", "error", err)
			return
		}
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
		jsonMarshal, err := json.Marshal(listeners)
		if err != nil {
			slog.Error("Failed to marshal listeners.", "func", "Listeners", "error", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(jsonMarshal)
		if err != nil {
			slog.Error("Failed to write response.", "func", "Listeners", "error", err)
			return
		}
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
		jsonMarshal, err := json.Marshal(bitrate)
		if err != nil {
			slog.Error("Failed to marshal bitrate.", "func", "Bitrate", "error", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(jsonMarshal)
		if err != nil {
			slog.Error("Failed to write response.", "func", "Bitrate", "error", err)
			return
		}
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
		jsonMarshal, err := json.Marshal(version)
		if err != nil {
			slog.Error("Failed to marshal version.", "func", "Version", "error", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(jsonMarshal)
		if err != nil {
			slog.Error("Failed to write response.", "func", "Version", "error", err)
			return
		}
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
			slog.Error("Unable to skip the playing song.", "func", "DevSkip", "error", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusOK) // 200 OK
	}
}
