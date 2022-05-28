package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Jeffail/gabs"
	"github.com/dhowden/tag"
	"github.com/gorilla/websocket"
	"github.com/kenellorando/clog"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  0,
	WriteBufferSize: 0,
}

// socketRadioData() upgrades connections for websocket
// Transfers near real-time radio updates (now playing, stream URL, listener count)
func socketRadioData() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("Error upgrading socket connection.", err)
			return
		}
		log.Print("New socket connected.", conn.RemoteAddr())

		type Message struct {
			Type       string
			Artist     string
			Title      string
			Picture    []byte
			ListenURL  string
			Mountpoint string
			Listeners  float64
		}

		var lastArtist string
		var lastTitle string
		var lastListenURL string
		var lastMountpoint string
		var lastListeners float64

		for {
			resp, err := http.Get("http://" + c.StreamAddress + c.StreamPort + "/status-json.xsl")
			if err != nil {
				clog.Error("socketNowPlaying", "Failed to connect to audio stream server.", err)
				conn.WriteJSON(Message{Type: "NowPlaying", Title: "-"})              // Write message to client
				conn.WriteJSON(Message{Type: "Listeners", Listeners: -1})            // Write message to client
				conn.WriteJSON(Message{Type: "StreamConnection", Mountpoint: "N/A"}) // Write message to client
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
			var currentHost = fmt.Sprintf(jsonParsed.Path("icestats.host").Data().(string))
			var currentMountpoint = fmt.Sprintf(jsonParsed.Path("icestats.source.server_name").Data().(string))
			var currentListeners = jsonParsed.Path("icestats.source.listeners").Data().(float64)

			var currentListenURL = currentHost + "/" + currentMountpoint

			if (lastArtist != currentArtist) || (lastTitle != currentTitle) {
				currentPicture := getAlbumArt(currentArtist, currentTitle)

				conn.WriteJSON(Message{Type: "NowPlaying", Artist: currentArtist, Title: currentTitle, Picture: currentPicture})
				lastArtist = currentArtist
				lastTitle = currentTitle
			}
			if (lastListenURL != currentListenURL) || (lastMountpoint != currentMountpoint) {
				conn.WriteJSON(Message{Type: "StreamConnection", ListenURL: currentListenURL, Mountpoint: currentMountpoint})
				lastListenURL = currentListenURL
				lastMountpoint = currentMountpoint
			}
			if lastListeners != currentListeners {
				conn.WriteJSON(Message{Type: "Listeners", Listeners: currentListeners})
				lastListeners = currentListeners
			}
			time.Sleep(1 * time.Second)
		}
	}
}

// This queries for the currently playing song path, then reads the data of the file directly.
func getAlbumArt(currentArtist string, currentTitle string) []byte {
	clog.Debug("getAlbumArt", "1")
	log.Printf("%s %s", currentArtist, currentTitle)
	//selectStatement := fmt.Sprintf("SELECT path FROM %s WHERE title=\"%s\" AND artist=\"%v\"", c.MetadataTable, currentTitle, currentArtist)

	selectStatement := fmt.Sprintf("SELECT path FROM %s WHERE artist=\"%v\" AND title=\"%v\";", c.MetadataTable, currentArtist, currentTitle)
	//selectStatement := fmt.Sprintf("SELECT path FROM %s WHERE rowid=1", c.MetadataTable)

	rows, err := database.Query(selectStatement)

	log.Printf("%s", selectStatement)
	log.Printf("%v", rows)
	if err != nil {
		clog.Error("getAlbumArt", "Could not query the DB for a path.", err)
		return nil
	}

	clog.Debug("getAlbumArt", "2")
	if err != nil {
		clog.Error("getAlbumArt", "Could not query the DB for a path.", err)
		return nil
	}
	var pic []byte
	var path string

	for rows.Next() {

		clog.Debug("getAlbumArt", path)
		clog.Debug("getAlbumArt", "rowsnextrun")
		err := rows.Scan(&path)
		clog.Debug("getAlbumArt", path)
		if err != nil {
			clog.Debug("getAlbumArt", "ae")
			return nil
		}
		clog.Debug("getAlbumArt", "3")
		// Open a file for reading
		file, e := os.Open(path)
		clog.Debug("getAlbumArt", "4")
		if e != nil {
			clog.Debug("getAlbumArt", "4e")
			clog.Error("getAlbumArt", "4e", e)
			return nil
		}

		// Read metadata from the file
		tags, err := tag.ReadFrom(file)

		clog.Debug("getAlbumArt", "5")
		if err != nil {
			clog.Debug("getAlbumArt", "5e")
			return nil
		}
		pic = tags.Picture().Data
	}
	return pic
}
