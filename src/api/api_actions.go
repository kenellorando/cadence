// api_actions.go
// API interactions for Postgres, Icecast, Liquidsoap.

package main

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"

	"strings"
	"time"

	"github.com/Jeffail/gabs"
	"github.com/fsnotify/fsnotify"
)

type Playing struct {
	Song       SongData
	Host       string
	Mountpoint string
	Listeners  float64
	Bitrate    float64
	Ended  		time.Time
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

var now = Playing{}
var prev = Playing{}
var history = make([]Playing, 0, 10)

func (now *Playing) SetPlaying(title string, artist string, host string, mountpoint string, listeners float64, bitrate float64) {
	now.Song.Title = title
	now.Song.Artist = artist
	now.Host = host
	now.Mountpoint = mountpoint
	now.Listeners = listeners
	now.Bitrate = bitrate
}

func (now *Playing) ResetPlaying() {
	now.Song.Title = "-"
	now.Song.Artist = "-"
	now.Host = "-"
	now.Mountpoint = "-"
	now.Listeners = -1
}

// Takes a query string to search the database.
// Returns a slice of SongData of songs ordered by relevance.
func searchByQuery(query string) (queryResults []SongData, err error) {
	query = strings.TrimSpace(query)
	slog.Debug(fmt.Sprintf("Searching database for query: '%v'", query), "func", "searchByQuery")
	selectWhereStatement := fmt.Sprintf("SELECT \"id\", \"artist\", \"title\",\"album\", \"genre\", \"year\" FROM %s ",
		c.PostgresTableName) + "WHERE artist ILIKE $1 OR title ILIKE $2 ORDER BY LEAST(levenshtein($3, artist), levenshtein($4, title))"
	rows, err := dbp.Query(selectWhereStatement, "%"+query+"%", "%"+query+"%", query, query)
	if err != nil {
		slog.Error("Database search failed.", "func", "searchByQuery", "error", err)
		return nil, err
	}
	for rows.Next() {
		song := &SongData{}
		err = rows.Scan(&song.ID, &song.Artist, &song.Title, &song.Album, &song.Genre, &song.Year)
		if err != nil {
			slog.Error("Data scan failed.", "func", "searchByQuery", "error", err)
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
	slog.Debug(fmt.Sprintf("Searching database for: %s by %s", title, artist), "func", "searchByTitleArtist")
	selectStatement := fmt.Sprintf("SELECT id,artist,title,album,genre,year FROM %s WHERE title LIKE $1 AND artist LIKE $2;",
		c.PostgresTableName)
	rows, err := dbp.Query(selectStatement, title, artist)
	if err != nil {
		slog.Error("Could not query DB.", "func", "searchByTitleArtist", "error", err)
		return nil, err
	}
	for rows.Next() {
		song := &SongData{}
		err = rows.Scan(&song.ID, &song.Artist, &song.Title, &song.Album, &song.Genre, &song.Year)
		if err != nil {
			slog.Error("Data scan failed.", "func", "searchByTitleArtist", "error", err)
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
	slog.Debug(fmt.Sprintf("Searching database for the path of song: '%v'", id), "func", "getPathById")
	selectWhereStatement := fmt.Sprintf("SELECT \"path\" FROM %s WHERE id=%v", c.PostgresTableName, id)
	rows, err := dbp.Query(selectWhereStatement)
	if err != nil {
		slog.Error("Database search failed.", "func", "getPathById", "error", err)
		return "", err
	}
	for rows.Next() {
		err = rows.Scan(&path)
		if err != nil {
			slog.Error("Data scan failed.", "func", "getPathById", "error", err)
			return "", err
		}
	}
	return path, nil
}

// Takes an absolute song path, submits the path to be queued in Liquidsoap.
// Returns the response message from Liquidsoap.
func liquidsoapRequest(path string) (message string, err error) {
	// Telnet to liquidsoap
	slog.Debug("Connecting to liquidsoap service...", "func", "liquidsoapRequest")
	conn, err := net.Dial("tcp", c.LiquidsoapAddress+c.LiquidsoapPort)
	if err != nil {
		slog.Error("Failed to connect to audio source server.", "func", "liquidsoapRequest", "error", err)
		return "", err
	}
	defer conn.Close()
	// Push song request to source service, listen for a response, and quit the telnet session.
	fmt.Fprintf(conn, "request.push "+path+"\n")
	message, err = bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		slog.Error("Failed to read stream response message from audio source server.", "func", "liquidsoapRequest", "error", err)
	}
	slog.Info(fmt.Sprintf("Message from audio source server: %s", message), "func", "liquidsoapRequest")
	fmt.Fprintf(conn, "quit"+"\n")
	return message, nil
}

func liquidsoapSkip() (message string, err error) {
	slog.Debug("Connecting to liquidsoap service...", "func", "liquidsoapSkip")
	conn, err := net.Dial("tcp", c.LiquidsoapAddress+c.LiquidsoapPort)
	if err != nil {
		slog.Error("Failed to connect to audio source server.", "func", "liquidsoapSkip", "error", err)
		return "", err
	}
	defer conn.Close()
	fmt.Fprintf(conn, "cadence1.skip\n")
	// Listen for response
	message, err = bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		slog.Error("Failed to read stream response message from audio source server.", "func", "liquidsoapSkip", "error", err)
	}
	slog.Debug(fmt.Sprintf("Message from audio source server: %s", message), "func", "liquidsoapSkip")
	fmt.Fprintf(conn, "quit"+"\n")
	return message, nil
}

// Watches the music directory (CSERVER_MUSICDIR) for any changes, and reconfigures the database.
func filesystemMonitor() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		slog.Error("Error creating watcher.", "func", "fileSystemMonitor", "error", err)
		return
	}
	defer watcher.Close()
	err = watcher.Add(c.MusicDir)
	if err != nil {
		slog.Error("Error adding music directory to watcher.", "func", "fileSystemMonitor", "error", err)
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
				slog.Info("Change detected in music library.", "func", "fileSystemMonitor")
				err = postgresPopulate()
				if err != nil {
					slog.Error("Failed to populate.", "func", "fileSystemMonitor", "error", err)
					return
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					continue
				}
				slog.Error("Error watching music library.", "func", "fileSystemMonitor", "error", err)
			}
		}
	}()
	<-done
}

// Watches the Icecast status page and updates stream info for SSE.
func icecastMonitor() {
	resp, err := http.Get("http://" + c.IcecastAddress + c.IcecastPort + "/status-json.xsl")
	if err != nil {
		slog.Error("Unable to stream data from the Icecast service.", "func", "icecastMonitor", "error", err)
		now.ResetPlaying()
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		slog.Debug("Unable to connect to Icecast.", "func", "icecastMonitor")
		now.ResetPlaying()
		return
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Debug("Connected to Icecast but unable to read response.", "func", "icecastMonitor")
		now.ResetPlaying()
		return
	}
	jsonParsed, err := gabs.ParseJSON([]byte(body))
	if err != nil {
		slog.Debug("Connected to Icecast but unable to parse response.", "func", "icecastMonitor")
		now.ResetPlaying()
		return
	}
	if jsonParsed.Path("icestats.source.title").Data() == nil || jsonParsed.Path("icestats.source.artist").Data() == nil {
		slog.Debug("Connected to Icecast, but saw nothing playing.", "func", "icecastMonitor")
		now.ResetPlaying()
		return
	}

	now.SetPlaying(jsonParsed.Path("icestats.source.artist").Data().(string),
		jsonParsed.Path("icestats.source.title").Data().(string),
		jsonParsed.Path("icestats.host").Data().(string),
		jsonParsed.Path("icestats.source.server_name").Data().(string),
		jsonParsed.Path("icestats.source.listeners").Data().(float64),
		jsonParsed.Path("icestats.source.bitrate").Data().(float64),
	)

	if prev.Song != now.Song {
		slog.Info(fmt.Sprintf("Now Playing: %s by %s", now.Song.Title, now.Song.Artist), "func", "icecastMonitor")
		// Dump the artwork rate limiter database first thing before updates
		// are sent out to reset artwork request count.
		dbr.RateLimitArt.FlushDB(ctx)

		radiodata_sse.SendEventMessage(now.Song.Title, "title", "")
		radiodata_sse.SendEventMessage(now.Song.Artist, "artist", "")
		if (prev.Song.Title != "") && (prev.Song.Artist != "") {
			addToHistory := Playing{Song: prev.Song, Ended: time.Now()}
			history = append([]Playing{addToHistory}, history...)
			if len(history) > 10 {
				history = history[1:]
			}
			radiodata_sse.SendEventMessage("update", "history", "")
		}
	}
	if (prev.Host != now.Host) || (prev.Mountpoint != now.Mountpoint) {
		slog.Info(fmt.Sprintf("Audio stream on: <%s/%s>", now.Host, now.Mountpoint), "func", "icecastMonitor")
		radiodata_sse.SendEventMessage(fmt.Sprintf(now.Host, "/", now.Mountpoint), "listenurl", "")
	}
	if prev.Listeners != now.Listeners {
		slog.Info(fmt.Sprintf("Listener count: <%v>", now.Listeners), "func", "icecastMonitor")
		radiodata_sse.SendEventMessage(fmt.Sprint(now.Listeners), "listeners", "")
	}
	prev = now
}
