// api_actions.go
// Repeatable actions used by API functions. Mostly database and audio client functions.
// No responses to clients here.

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
// Returns a slice of SongData of songs by relevance.
func searchByQuery(query string) (queryResults []SongData, err error) {
	clog.Info("dbQuery", fmt.Sprintf("Searching database for query: '%v'", query))

	selectWhereStatement := fmt.Sprintf("SELECT \"rowid\", \"artist\", \"title\",\"album\", \"genre\", \"year\" FROM %s ", c.MetadataTable) + "WHERE artist LIKE $1 OR title LIKE $2 ORDER BY rank"
	rows, err := db.Query(selectWhereStatement, "%"+query+"%", "%"+query+"%")
	if err != nil {
		clog.Error("Search", "Database search failed.", err)
		return nil, err
	}

	for rows.Next() {
		song := &SongData{}
		err = rows.Scan(&song.ID, &song.Artist, &song.Title, &song.Album, &song.Genre, &song.Year)
		if err != nil {
			clog.Error("Search", "Data scan failed.", err)
			continue
		}
		queryResults = append(queryResults, SongData{ID: song.ID, Artist: song.Artist, Title: song.Title, Album: song.Album, Genre: song.Genre, Year: song.Year})
	}

	return queryResults, nil
}

// Takes a title and artist string to find exact matches for both.
// Returns a slice of SongData of songs by relevance.
// This search should only have one result unless multiple audio files share the exact same title and artist.
func searchByTitleArtist(title string, artist string) (queryResults []SongData, err error) {
	selectStatement := fmt.Sprintf("SELECT rowid,artist,title,album,genre,year FROM %s WHERE title=\"%v\" AND artist=\"%v\";", c.MetadataTable, title, artist)
	rows, err := db.Query(selectStatement)
	if err != nil {
		clog.Error("NowPlayingMetadata", "Could not query DB.", err)
		return nil, err
	}

	for rows.Next() {
		song := &SongData{}
		err = rows.Scan(&song.ID, &song.Artist, &song.Title, &song.Album, &song.Genre, &song.Year)
		if err != nil {
			clog.Error("Search", "Data scan failed.", err)
			continue
		}
		queryResults = append(queryResults, SongData{ID: song.ID, Artist: song.Artist, Title: song.Title, Album: song.Album, Genre: song.Genre, Year: song.Year})
	}
	return queryResults, nil
}

// Takes a song ID integer.
// Returns a string absolute path of the song.
func getPathById(id int) (path string, err error) {
	clog.Info("dbQuery", fmt.Sprintf("Searching database for the path of song: '%v'", id))

	selectWhereStatement := fmt.Sprintf("SELECT \"path\" FROM %s WHERE rowid=%v", c.MetadataTable, id)

	rows, err := db.Query(selectWhereStatement)
	if err != nil {
		clog.Error("Search", "Database search failed.", err)
		return "", err
	}
	for rows.Next() {
		err = rows.Scan(&path)
		if err != nil {
			clog.Error("Search", "Data scan failed.", err)
			return "", err
		}
	}

	return path, nil
}

// Takes an absolute song path, submits the path to be queued in Liquidsoap.
// Returns the response message from Liquidsoap.
func pushRequest(path string) (message string, err error) {
	// Telnet to liquidsoap
	clog.Debug("Request", "Connecting to liquidsoap service...")
	conn, err := net.Dial("tcp", c.SourceAddress+c.SourcePort)
	defer conn.Close()
	if err != nil {
		clog.Error("Request", "Failed to connect to audio source server.", err)
		return "", err
	}

	// Push song request to source service
	fmt.Fprintf(conn, "request.push "+path+"\n")
	// Listen for response
	message, _ = bufio.NewReader(conn).ReadString('\n')
	clog.Info("Request", fmt.Sprintf("Message from audio source server: %s", message))
	// Goodbye
	fmt.Fprintf(conn, "quit"+"\n")

	return message, nil
}

// Takes no arguments.
// Returns the title and artist strings actively playing on Icecast.
func getNowPlaying() (title string, artist string, err error) {
	if nowTitle == "-" || nowArtist == "-" {
		clog.Info("getNowPlaying", "Did not see anything playing in Icecast.")
	} else {
		clog.Info("getNowPlaying", fmt.Sprintf("The stream server reports it is playing: '%s' by '%s'.", title, artist))
	}
	return nowTitle, nowArtist, nil
}

var preTitle, preArtist, nowTitle, nowArtist string = "-", "-", "-", "-"
var preHost, preMountpoint, nowHost, nowMountpoint string = "-", "-", "-", "-"
var preListeners, nowListeners float64 = -1, -1

// Takes no arguments.
// Returns nothing, but sends updated stream info to the event source stream
// It watches for changes to now playing info.
func icecastMonitor() {
	for {
		radiodata_sse.SendEventMessage(nowTitle, "title", "")
		radiodata_sse.SendEventMessage(nowArtist, "artist", "")

		resp, err := http.Get("http://" + c.StreamAddress + c.StreamPort + "/status-json.xsl")
		if err != nil {
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			clog.Debug("icecastMonitor", "Unable to connect to Icecast.")
			resetIcecastData()
			continue
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			clog.Debug("icecastMonitor", "Connected to Icecast but unable to read response.")
			resetIcecastData()
			continue
		}
		jsonParsed, err := gabs.ParseJSON([]byte(body))
		if err != nil {
			clog.Debug("icecastMonitor", "Connected to Icecast but unable to parse response.")
			resetIcecastData()
			continue
		}
		if jsonParsed.Path("icestats.source.title").Data() == nil || jsonParsed.Path("icestats.source.artist").Data() == nil {
			clog.Debug("icecastMonitor", "Connected to Icecast, but did not see anything playing.")
			resetIcecastData()
			continue
		}

		nowArtist = jsonParsed.Path("icestats.source.artist").Data().(string)
		nowTitle = jsonParsed.Path("icestats.source.title").Data().(string)
		if (preTitle != nowTitle) || (preArtist != nowArtist) {
			clog.Debug("icecastMonitor", fmt.Sprintf("Now Playing: %s by %s", nowTitle, nowArtist))
			radiodata_sse.SendEventMessage(nowTitle, "title", "")
			radiodata_sse.SendEventMessage(nowArtist, "artist", "")
			preTitle = nowTitle
			preArtist = nowArtist
		}

		nowHost = jsonParsed.Path("icestats.host").Data().(string)
		nowMountpoint = jsonParsed.Path("icestats.source.server_name").Data().(string)
		if (preHost != nowHost) || (preMountpoint != nowMountpoint) {
			clog.Debug("icecastMonitor", fmt.Sprintf("Stream host: <%s>", nowHost))
			clog.Debug("icecastMonitor", fmt.Sprintf("Stream mountpoint: <%s>", nowMountpoint))
			radiodata_sse.SendEventMessage(fmt.Sprintf(nowHost, "/", nowMountpoint), "listenurl", "")
			preHost = nowHost
			preMountpoint = nowMountpoint
		}

		nowListeners = jsonParsed.Path("icestats.source.listeners").Data().(float64)
		if preListeners != nowListeners {
			clog.Debug("icecastMonitor", fmt.Sprintf("Listener count: <%v>", nowListeners))
			radiodata_sse.SendEventMessage(fmt.Sprint(nowListeners), "listeners", "")
			preListeners = nowListeners
		}

		time.Sleep(1 * time.Second)
	}
}

// Resets Icecast data to default values. Use when Icecast is unreachable.
func resetIcecastData() {
	preTitle, preArtist, nowTitle, nowArtist = "-", "-", "-", "-"
	preHost, preMountpoint, nowHost, nowMountpoint = "-", "-", "-", "-"
	preListeners, nowListeners = -1, -1
}
