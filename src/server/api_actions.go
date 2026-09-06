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
	"sync"
	"time"

	"github.com/Jeffail/gabs"
	"github.com/fsnotify/fsnotify"
)

// Every outbound call in this file talks to a service on the same compose
// network. None of them are allowed to hang: the monitor polls on a one second
// cadence, and a stalled request would silently stop all stream updates.
const serviceTimeout = 5 * time.Second

var icecastClient = &http.Client{Timeout: serviceTimeout}

var now = RadioInfo{}

// icecastMonitor writes now and history from its own goroutine once a second
// while request handlers read them, so every access is guarded. The monitor is
// the only writer, and therefore only needs the lock when it mutates; readers
// outside it must go through nowPlaying and playHistory.
var radioMutex sync.RWMutex

// Returns a copy of the current radio state, safe to use off the monitor goroutine.
func nowPlaying() RadioInfo {
	radioMutex.RLock()
	defer radioMutex.RUnlock()
	return now
}

// Returns a copy of the recently played songs, oldest first.
func playHistory() []playRecord {
	radioMutex.RLock()
	defer radioMutex.RUnlock()
	return append([]playRecord(nil), history...)
}

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
	slog.Debug(fmt.Sprintf("Searching database for query: '%v'", query), "func", "searchByQuery")
	selectWhereStatement := fmt.Sprintf("SELECT \"id\", \"artist\", \"title\",\"album\", \"genre\", \"year\" FROM %s ",
		c.PostgresTableName) + "WHERE artist ILIKE $1 OR title ILIKE $2 ORDER BY LEAST(levenshtein($3, artist), levenshtein($4, title))"
	rows, err := dbp.Query(selectWhereStatement, "%"+query+"%", "%"+query+"%", query, query)
	if err != nil {
		slog.Error("Database search failed.", "func", "searchByQuery", "error", err)
		return nil, err
	}
	defer rows.Close()
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
	if err = rows.Err(); err != nil {
		slog.Error("Failed while reading search results.", "func", "searchByQuery", "error", err)
		return nil, err
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
	defer rows.Close()
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
	if err = rows.Err(); err != nil {
		slog.Error("Failed while reading search results.", "func", "searchByTitleArtist", "error", err)
		return nil, err
	}
	return queryResults, nil
}

// Takes a song ID integer.
// Returns the absolute path of the audio file.
func getPathById(id int) (path string, err error) {
	slog.Debug(fmt.Sprintf("Searching database for the path of song: '%v'", id), "func", "getPathById")
	selectWhereStatement := fmt.Sprintf("SELECT \"path\" FROM %s WHERE id=$1", c.PostgresTableName)
	rows, err := dbp.Query(selectWhereStatement, id)
	if err != nil {
		slog.Error("Database search failed.", "func", "getPathById", "error", err)
		return "", err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&path)
		if err != nil {
			slog.Error("Data scan failed.", "func", "getPathById", "error", err)
			return "", err
		}
	}
	if err = rows.Err(); err != nil {
		slog.Error("Failed while reading path result.", "func", "getPathById", "error", err)
		return "", err
	}
	return path, nil
}

// Takes an absolute song path, submits the path to be queued in Liquidsoap.
// Returns the response message from Liquidsoap.
func liquidsoapRequest(path string) (message string, err error) {
	// Telnet to liquidsoap
	slog.Debug("Connecting to liquidsoap service...", "func", "liquidsoapRequest")
	conn, err := net.DialTimeout("tcp", c.LiquidsoapAddress+c.LiquidsoapPort, serviceTimeout)
	if err != nil {
		slog.Error("Failed to connect to audio source server.", "func", "liquidsoapRequest", "error", err)
		return "", err
	}
	defer conn.Close()
	if err = conn.SetDeadline(time.Now().Add(serviceTimeout)); err != nil {
		slog.Error("Failed to set a deadline on the audio source connection.", "func", "liquidsoapRequest", "error", err)
		return "", err
	}
	// Push song request to source service, listen for a response, and quit the telnet session.
	fmt.Fprintf(conn, "request.push %s\n", path)
	message, err = bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		slog.Error("Failed to read stream response message from audio source server.", "func", "liquidsoapRequest", "error", err)
	}
	slog.Info(fmt.Sprintf("Message from audio source server: %s", message), "func", "liquidsoapRequest")
	fmt.Fprint(conn, "quit\n")
	return message, nil
}

func liquidsoapSkip() (message string, err error) {
	slog.Debug("Connecting to liquidsoap service...", "func", "liquidsoapSkip")
	conn, err := net.DialTimeout("tcp", c.LiquidsoapAddress+c.LiquidsoapPort, serviceTimeout)
	if err != nil {
		slog.Error("Failed to connect to audio source server.", "func", "liquidsoapSkip", "error", err)
		return "", err
	}
	defer conn.Close()
	if err = conn.SetDeadline(time.Now().Add(serviceTimeout)); err != nil {
		slog.Error("Failed to set a deadline on the audio source connection.", "func", "liquidsoapSkip", "error", err)
		return "", err
	}
	fmt.Fprint(conn, "cadence1.skip\n")
	// Listen for response
	message, err = bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		slog.Error("Failed to read stream response message from audio source server.", "func", "liquidsoapSkip", "error", err)
	}
	slog.Debug(fmt.Sprintf("Message from audio source server: %s", message), "func", "liquidsoapSkip")
	fmt.Fprint(conn, "quit\n")
	return message, nil
}

// How long the music library must be quiet before a change triggers a rebuild.
const librarySettleDelay = 5 * time.Second

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
		// Rebuilding drops and repopulates the entire metadata table, and a bulk
		// change to the library (copying an album in, a sync tool running) arrives
		// as a burst of events. Rebuilding per event means dozens of destructive
		// rebuilds, with search returning nothing for the duration of each. Wait
		// for the burst to go quiet and rebuild once.
		var settle *time.Timer
		var settled <-chan time.Time
		for {
			select {
			case _, ok := <-watcher.Events:
				if !ok {
					continue
				}
				slog.Debug("Change detected in music library.", "func", "fileSystemMonitor")
				if settle == nil {
					settle = time.NewTimer(librarySettleDelay)
				} else {
					// Stop reports false if the timer already fired, in which case
					// its value is still waiting in the channel and must be drained
					// before the timer can be reused.
					if !settle.Stop() {
						select {
						case <-settle.C:
						default:
						}
					}
					settle.Reset(librarySettleDelay)
				}
				settled = settle.C
			case <-settled:
				settle, settled = nil, nil
				slog.Info("Music library changes have settled, rebuilding database.", "func", "fileSystemMonitor")
				// A failed rebuild is not a reason to stop watching. Giving up here
				// left the library permanently stale until the service restarted.
				if err := postgresPopulate(); err != nil {
					slog.Error("Failed to populate.", "func", "fileSystemMonitor", "error", err)
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
	var prev = RadioInfo{}
	// Resets now playing, stream URL, and listener global variables to defaults. Used when Icecast is unreachable.
	icecastDataReset := func() {
		radioMutex.Lock()
		defer radioMutex.Unlock()
		now.Song.Title, now.Song.Artist, now.Host, now.Mountpoint = "-", "-", "-", "-"
		now.Listeners = -1
	}
	checkIcecastStatus := func() {
		resp, err := icecastClient.Get("http://" + c.IcecastAddress + c.IcecastPort + "/status-json.xsl")
		if err != nil {
			slog.Error("Unable to stream data from the Icecast service.", "func", "icecastMonitor", "error", err)
			icecastDataReset()
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			slog.Debug("Unable to connect to Icecast.", "func", "icecastMonitor")
			icecastDataReset()
			return
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			slog.Debug("Connected to Icecast but unable to read response.", "func", "icecastMonitor")
			icecastDataReset()
			return
		}
		jsonParsed, err := gabs.ParseJSON([]byte(body))
		if err != nil {
			slog.Debug("Connected to Icecast but unable to parse response.", "func", "icecastMonitor")
			icecastDataReset()
			return
		}
		if jsonParsed.Path("icestats.source.title").Data() == nil || jsonParsed.Path("icestats.source.artist").Data() == nil {
			slog.Debug("Connected to Icecast, but saw nothing playing.", "func", "icecastMonitor")
			icecastDataReset()
			return
		}

		radioMutex.Lock()
		now.Song.Artist = jsonParsed.Path("icestats.source.artist").Data().(string)
		now.Song.Title = jsonParsed.Path("icestats.source.title").Data().(string)
		now.Host = jsonParsed.Path("icestats.host").Data().(string)
		now.Mountpoint = jsonParsed.Path("icestats.source.server_name").Data().(string)
		now.Listeners = jsonParsed.Path("icestats.source.listeners").Data().(float64)
		now.Bitrate = jsonParsed.Path("icestats.source.bitrate").Data().(float64)
		radioMutex.Unlock()

		if (prev.Song.Title != now.Song.Title) || (prev.Song.Artist != now.Song.Artist) {
			slog.Info(fmt.Sprintf("Now Playing: %s by %s", now.Song.Title, now.Song.Artist), "func", "icecastMonitor")
			// Dump the artwork rate limiter database first thing before updates
			// are sent out to reset artwork request count.
			dbr.RateLimitArt.FlushDB(ctx)

			radiodata_sse.SendEventMessage(now.Song.Title, "title", "")
			radiodata_sse.SendEventMessage(now.Song.Artist, "artist", "")
			if (prev.Song.Title != "") && (prev.Song.Artist != "") {
				radioMutex.Lock()
				history = append(history, playRecord{Title: prev.Song.Title, Artist: prev.Song.Artist, Ended: time.Now()})
				if len(history) > 10 {
					history = history[1:]
				}
				radioMutex.Unlock()
				radiodata_sse.SendEventMessage("update", "history", "")
			}
		}
		if (prev.Host != now.Host) || (prev.Mountpoint != now.Mountpoint) {
			slog.Info(fmt.Sprintf("Audio stream on: <%s/%s>", now.Host, now.Mountpoint), "func", "icecastMonitor")
			radiodata_sse.SendEventMessage(now.Host+"/"+now.Mountpoint, "listenurl", "")
		}
		if prev.Listeners != now.Listeners {
			slog.Info(fmt.Sprintf("Listener count: <%v>", now.Listeners), "func", "icecastMonitor")
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
