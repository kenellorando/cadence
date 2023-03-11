// api_actions.go
// API interactions for Postgres, Icecast, Liquidsoap.

package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/Jeffail/gabs"
	"github.com/fsnotify/fsnotify"
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
	Path   string
}

// Takes a query string to search the database.
// Returns a slice of SongData of songs ordered by relevance.
func searchByQuery(query string) (queryResults []SongData, err error) {
	query = strings.TrimSpace(query)
	clog.Debug("searchByQuery", fmt.Sprintf("Searching database for query: '%v'", query))
	selectWhereStatement := fmt.Sprintf("SELECT \"id\", \"artist\", \"title\",\"album\", \"genre\", \"year\" FROM %s ",
		c.PostgresTableName) + "WHERE artist ILIKE $1 OR title ILIKE $2 ORDER BY LEAST(levenshtein($3, artist), levenshtein($4, title))"
	rows, err := dbp.Query(selectWhereStatement, "%"+query+"%", "%"+query+"%", query, query)
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
		queryResults = append(queryResults,
			SongData{ID: song.ID, Artist: song.Artist, Title: song.Title, Album: song.Album, Genre: song.Genre, Year: song.Year})
	}
	return queryResults, nil
}

// Takes a title and artist string to find a song which exactly matches.
// Returns a list of SongData whose first result [0] is the first (best) match.
// This will not work if multiple songs share the exact same title and artist.
func searchByTitleArtist(title string, artist string) (queryResults []SongData, err error) {
	title, artist = strings.TrimSpace(title), strings.TrimSpace(artist)
	clog.Debug("searchByTitleArtist", fmt.Sprintf("Searching database for: %s by %s", title, artist))
	selectStatement := fmt.Sprintf("SELECT id,artist,title,album,genre,year FROM %s WHERE title LIKE $1 AND artist LIKE $2;",
		c.PostgresTableName)
	rows, err := dbp.Query(selectStatement, title, artist)
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
		queryResults = append(queryResults,
			SongData{ID: song.ID, Artist: song.Artist, Title: song.Title, Album: song.Album, Genre: song.Genre, Year: song.Year})
	}
	return queryResults, nil
}

// Takes a song ID integer.
// Returns the absolute path of the audio file.
func getPathById(id int) (path string, err error) {
	clog.Debug("getPathById", fmt.Sprintf("Searching database for the path of song: '%v'", id))
	selectWhereStatement := fmt.Sprintf("SELECT \"path\" FROM %s WHERE id=%v", c.PostgresTableName, id)
	rows, err := dbp.Query(selectWhereStatement)
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
	conn, err := net.Dial("tcp", c.LiquidsoapAddress+c.LiquidsoapPort)
	if err != nil {
		clog.Error("liquidsoapRequest", "Failed to connect to audio source server.", err)
		return "", err
	}
	defer conn.Close()
	// Push song request to source service, listen for a response, and quit the telnet session.
	fmt.Fprintf(conn, "request.push "+path+"\n")
	message, err = bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		clog.Error("liquidsoapRequest", "Failed to read stream response message from audio source server.", err)
	}
	clog.Info("liquidsoapRequest", fmt.Sprintf("Message from audio source server: %s", message))
	fmt.Fprintf(conn, "quit"+"\n")
	return message, nil
}

func liquidsoapSkip() (message string, err error) {
	clog.Debug("liquidsoapRequest", "Connecting to liquidsoap service...")
	conn, err := net.Dial("tcp", c.LiquidsoapAddress+c.LiquidsoapPort)
	if err != nil {
		clog.Error("liquidsoapRequest", "Failed to connect to audio source server.", err)
		return "", err
	}
	defer conn.Close()
	fmt.Fprintf(conn, "cadence1.skip\n")
	// Listen for response
	message, err = bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		clog.Error("liquidsoapSkip", "Failed to read stream response message from audio source server.", err)
	}
	clog.Debug("liquidsoapSkip", fmt.Sprintf("Message from audio source server: %s", message))
	fmt.Fprintf(conn, "quit"+"\n")
	return message, nil
}

// Watches the music directory (CSERVER_MUSICDIR) for any changes, and reconfigures the database.
func filesystemMonitor() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		clog.Error("fileSystemMonitor", "Error creating watcher.", err)
		return
	}
	defer watcher.Close()
	err = watcher.Add(c.MusicDir)
	if err != nil {
		clog.Error("fileSystemMonitor", "Error adding music directory to watcher.", err)
		return
	}
	done := make(chan bool)
	go func() {
		for {
			select {
			case _, ok := <-watcher.Events:
				if !ok {
					continue
				}
				clog.Info("fileSystemMonitor", "Change detected in music library.")
				postgresPopulate()
			case err, ok := <-watcher.Errors:
				if !ok {
					continue
				}
				clog.Error("fileSystemMonitor", "Error watching music library.", err)
			}
		}
	}()
	<-done
}

// Watches the Icecast status page and updates stream info for SSE.
func icecastMonitor() {
	var prev = RadioInfo{}
	// Resets now playing, stream URL, and listener global variables to defaults. Used when Icecast is unreachable.
	icecastDataReset := func() {
		now.Song.Title, now.Song.Artist, now.Host, now.Mountpoint = "-", "-", "-", "-"
		now.Listeners = -1
	}
	checkIcecastStatus := func() {
		resp, err := http.Get("http://" + c.IcecastAddress + c.IcecastPort + "/status-json.xsl")
		if err != nil {
			clog.Error("icecastMonitor", "Unable to stream data from the Icecast service.", err)
			icecastDataReset()
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			clog.Debug("icecastMonitor", "Unable to connect to Icecast.")
			icecastDataReset()
			return
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			clog.Debug("icecastMonitor", "Connected to Icecast but unable to read response.")
			icecastDataReset()
			return
		}
		jsonParsed, err := gabs.ParseJSON([]byte(body))
		if err != nil {
			clog.Debug("icecastMonitor", "Connected to Icecast but unable to parse response.")
			icecastDataReset()
			return
		}
		if jsonParsed.Path("icestats.source.title").Data() == nil || jsonParsed.Path("icestats.source.artist").Data() == nil {
			clog.Debug("icecastMonitor", "Connected to Icecast, but saw nothing playing.")
			icecastDataReset()
			return
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
				if len(history) > 10 {
					history = history[1:]
				}
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
	}
	go func() {
		for {
			time.Sleep(1 * time.Second)
			checkIcecastStatus()
		}
	}()
}

var history = make([]playRecord, 0, 10)

type playRecord struct {
	Title  string
	Artist string
	Ended  time.Time
}
