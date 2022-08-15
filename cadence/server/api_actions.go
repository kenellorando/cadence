// api_actions.go
// Repeatable actions used by API functions. Mostly database and audio software calls.
// No direct responses to clients here.

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

// Song file metadata object
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

var preTitle, preArtist, nowTitle, nowArtist string = "-", "-", "-", "-"
var preHost, preMountpoint, nowHost, nowMountpoint string = "-", "-", "-", "-"
var preListeners, nowListeners float64 = -1, -1

// Takes no arguments.
// Returns nothing, but sends updated stream info to SSE and sets the global variables.
// It is launched as a goroutine by init.
func icecastMonitor() {
	for {
		resp, err := http.Get("http://" + c.StreamAddress + c.StreamPort + "/status-json.xsl")
		if err != nil {
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

		nowArtist = jsonParsed.Path("icestats.source.artist").Data().(string)
		nowTitle = jsonParsed.Path("icestats.source.title").Data().(string)
		if (preTitle != nowTitle) || (preArtist != nowArtist) {
			clog.Info("icecastMonitor", fmt.Sprintf("Now Playing: %s by %s", nowTitle, nowArtist))
			radiodata_sse.SendEventMessage(nowTitle, "title", "")
			radiodata_sse.SendEventMessage(nowArtist, "artist", "")
			preTitle = nowTitle
			preArtist = nowArtist
		}

		nowHost = jsonParsed.Path("icestats.host").Data().(string)
		nowMountpoint = jsonParsed.Path("icestats.source.server_name").Data().(string)
		if (preHost != nowHost) || (preMountpoint != nowMountpoint) {
			clog.Info("icecastMonitor", fmt.Sprintf("Stream host: <%s>", nowHost))
			clog.Info("icecastMonitor", fmt.Sprintf("Stream mountpoint: <%s>", nowMountpoint))
			radiodata_sse.SendEventMessage(fmt.Sprintf(nowHost, "/", nowMountpoint), "listenurl", "")
			preHost = nowHost
			preMountpoint = nowMountpoint
		}

		nowListeners = jsonParsed.Path("icestats.source.listeners").Data().(float64)
		if preListeners != nowListeners {
			clog.Info("icecastMonitor", fmt.Sprintf("Listener count: <%v>", nowListeners))
			radiodata_sse.SendEventMessage(fmt.Sprint(nowListeners), "listeners", "")
			preListeners = nowListeners
		}

		time.Sleep(1 * time.Second)
	}
}

// Resets now playing, stream URL, and listener global variables to defaults. Used when Icecast is unreachable.
func icecastDataReset() {
	preTitle, preArtist, nowTitle, nowArtist = "-", "-", "-", "-"
	preHost, preMountpoint, nowHost, nowMountpoint = "-", "-", "-", "-"
	preListeners, nowListeners = -1, -1
	time.Sleep(3 * time.Second)
}
