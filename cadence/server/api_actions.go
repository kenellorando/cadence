// api_actions.go
// API function repeatable actions.

package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/Jeffail/gabs"
	"github.com/kenellorando/clog"
)

var now = RadioInfo{}

type RadioInfo struct {
	Song       SongData
	Host       string
	Mountpoint string
	Listeners  float64
	Bitrate    float64
}

type SongData struct {
	ID     int
	Artist string
	Title  string
	Album  string
	Genre  string
	Year   int
}

// Takes a query string to search the database.
// Returns a slice of SongData of songs ordered by relevance.
func searchByQuery(query string) (queryResults []SongData, err error) {
	clog.Info("searchByQuery", fmt.Sprintf("Searching database for query: '%v'", query))

	selectWhereStatement := fmt.Sprintf("SELECT \"rowid\", \"artist\", \"title\",\"album\", \"genre\", \"year\" FROM %s ", c.MetadataTable) + "WHERE artist LIKE $1 OR title LIKE $2 ORDER BY rank"
	rows, err := db.Query(selectWhereStatement, "%"+query+"%", "%"+query+"%")
	if err != nil {
		clog.Error("searchByQuery", "Database search failed.", err)
		return nil, err
	}

	for rows.Next() {
		song := &SongData{}
		err = rows.Scan(&song.ID, &song.Artist, &song.Title, &song.Album, &song.Genre, &song.Year)
		if err != nil {
			clog.Error("searchByQuery", "Data scan failed.", err)
			continue
		}
		queryResults = append(queryResults, SongData{ID: song.ID, Artist: song.Artist, Title: song.Title, Album: song.Album, Genre: song.Genre, Year: song.Year})
	}

	return queryResults, nil
}

// Takes a title and artist string to find a song which exactly matches.
// Returns a slice of SongData of songs by relevance.
// This search should only have one result unless multiple audio files share the exact same title and artist.
func searchByTitleArtist(title string, artist string) (queryResults []SongData, err error) {
	selectStatement := fmt.Sprintf("SELECT rowid,artist,title,album,genre,year FROM %s WHERE title=\"%v\" AND artist=\"%v\";", c.MetadataTable, title, artist)
	rows, err := db.Query(selectStatement)
	if err != nil {
		clog.Error("searchByTitleArtist", "Could not query DB.", err)
		return nil, err
	}

	for rows.Next() {
		song := &SongData{}
		err = rows.Scan(&song.ID, &song.Artist, &song.Title, &song.Album, &song.Genre, &song.Year)
		if err != nil {
			clog.Error("searchByTitleArtist", "Data scan failed.", err)
			continue
		}
		queryResults = append(queryResults, SongData{ID: song.ID, Artist: song.Artist, Title: song.Title, Album: song.Album, Genre: song.Genre, Year: song.Year})
	}

	return queryResults, nil
}

// Takes a song ID integer.
// Returns the absolute path of the audio file.
func getPathById(id int) (path string, err error) {
	clog.Info("getPathById", fmt.Sprintf("Searching database for the path of song: '%v'", id))

	selectWhereStatement := fmt.Sprintf("SELECT \"path\" FROM %s WHERE rowid=%v", c.MetadataTable, id)

	rows, err := db.Query(selectWhereStatement)
	if err != nil {
		clog.Error("getPathById", "Database search failed.", err)
		return "", err
	}
	for rows.Next() {
		err = rows.Scan(&path)
		if err != nil {
			clog.Error("getPathById", "Data scan failed.", err)
			return "", err
		}
	}

	return path, nil
}

// Takes an absolute song path, submits the path to be queued in Liquidsoap.
// Returns the response message from Liquidsoap.
func liquidsoapRequest(path string) (message string, err error) {
	// Telnet to liquidsoap
	clog.Debug("liquidsoapRequest", "Connecting to liquidsoap service...")
	conn, err := net.Dial("tcp", c.SourceAddress+c.SourcePort)
	defer conn.Close()
	if err != nil {
		clog.Error("liquidsoapRequest", "Failed to connect to audio source server.", err)
		return "", err
	}

	// Push song request to source service
	fmt.Fprintf(conn, "request.push "+path+"\n")
	// Listen for response
	message, _ = bufio.NewReader(conn).ReadString('\n')
	clog.Info("liquidsoapRequest", fmt.Sprintf("Message from audio source server: %s", message))
	// Goodbye
	fmt.Fprintf(conn, "quit"+"\n")

	return message, nil
}

func liquidsoapSkip() (message string, err error) {
	clog.Debug("liquidsoapRequest", "Connecting to liquidsoap service...")
	conn, err := net.Dial("tcp", c.SourceAddress+c.SourcePort)
	defer conn.Close()
	if err != nil {
		clog.Error("liquidsoapRequest", "Failed to connect to audio source server.", err)
		return "", err
	}
	fmt.Fprintf(conn, "cadence1.skip\n")
	// Listen for response
	message, _ = bufio.NewReader(conn).ReadString('\n')
	fmt.Fprintf(conn, "quit"+"\n")
	return message, nil
}

// Returns nothing, but sends updated stream info to SSE and sets the global variables.
func icecastMonitor() {
	var prev = RadioInfo{}
	for {
		time.Sleep(1 * time.Second)
		resp, err := http.Get("http://" + c.StreamAddress + c.StreamPort + "/status-json.xsl")
		if resp != nil {
			defer resp.Body.Close()
		}
		if err != nil {
			clog.Error("icecastMonitor", "Unable to stream data from the Icecast service.", err)
			icecastDataReset()
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			clog.Debug("icecastMonitor", "Unable to connect to Icecast.")
			icecastDataReset()
			continue
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			clog.Debug("icecastMonitor", "Connected to Icecast but unable to read response.")
			icecastDataReset()
			continue
		}
		jsonParsed, err := gabs.ParseJSON([]byte(body))
		if err != nil {
			clog.Debug("icecastMonitor", "Connected to Icecast but unable to parse response.")
			icecastDataReset()
			continue
		}
		if jsonParsed.Path("icestats.source.title").Data() == nil || jsonParsed.Path("icestats.source.artist").Data() == nil {
			clog.Debug("icecastMonitor", "Connected to Icecast, but saw nothing playing.")
			icecastDataReset()
			continue
		}

		now.Song.Artist = jsonParsed.Path("icestats.source.artist").Data().(string)
		now.Song.Title = jsonParsed.Path("icestats.source.title").Data().(string)
		now.Host = jsonParsed.Path("icestats.host").Data().(string)
		now.Mountpoint = jsonParsed.Path("icestats.source.server_name").Data().(string)
		now.Listeners = jsonParsed.Path("icestats.source.listeners").Data().(float64)
		now.Bitrate = jsonParsed.Path("icestats.source.bitrate").Data().(float64)

		if (prev.Song.Title != now.Song.Title) || (prev.Song.Artist != now.Song.Artist) {
			clog.Info("icecastMonitor", fmt.Sprintf("Now Playing: %s by %s", now.Song.Title, now.Song.Artist))
			radiodata_sse.SendEventMessage(now.Song.Title, "title", "")
			radiodata_sse.SendEventMessage(now.Song.Artist, "artist", "")
			if (prev.Song.Title != "") && (prev.Song.Artist != "") {
				history = append(history, playRecord{Title: prev.Song.Title, Artist: prev.Song.Artist, Ended: time.Now()})
				radiodata_sse.SendEventMessage("update", "history", "")
			}
		}
		if (prev.Host != now.Host) || (prev.Mountpoint != now.Mountpoint) {
			clog.Info("icecastMonitor", fmt.Sprintf("Audio stream on: <%s/%s>", now.Host, now.Mountpoint))
			radiodata_sse.SendEventMessage(fmt.Sprintf(now.Host, "/", now.Mountpoint), "listenurl", "")
		}
		if prev.Listeners != now.Listeners {
			clog.Info("icecastMonitor", fmt.Sprintf("Listener count: <%v>", now.Listeners))
			radiodata_sse.SendEventMessage(fmt.Sprint(now.Listeners), "listeners", "")
		}

		prev = now
		resp.Body.Close()
	}
}

// Resets now playing, stream URL, and listener global variables to defaults. Used when Icecast is unreachable.
func icecastDataReset() {
	now.Song.Title, now.Song.Artist, now.Host, now.Mountpoint = "-", "-", "-", "-"
	now.Listeners = -1
}

var history = make([]playRecord, 0, 10)

type playRecord struct {
	Title  string
	Artist string
	Ended  time.Time
}
