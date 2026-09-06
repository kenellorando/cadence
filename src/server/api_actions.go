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

	"strconv"
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
	// An exact match, so = rather than LIKE: song titles legitimately contain %
	// and _, which LIKE would treat as wildcards.
	selectStatement := fmt.Sprintf("SELECT id,artist,title,album,genre,year FROM %s WHERE title = $1 AND artist = $2;",
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
	// Liquidsoap 1.4 named this after the queue id ("request.push"). 2.x fixes the
	// namespace at "request_queue" and ignores the id, so the old command comes
	// back as "unknown command" and every song request is silently dropped.
	fmt.Fprintf(conn, "request_queue.push %s\n", path)
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

// Icecast's status document is remote input, so every field below is read with
// a checked assertion. An unchecked one took the whole server down when Icecast
// 2.5 renamed a field: the monitor goroutine panicked on every poll, and a
// panic there is fatal to the process, so the server crash-looped.
func icecastString(parsed *gabs.Container, path string) (string, bool) {
	value, ok := parsed.Path(path).Data().(string)
	return value, ok
}

func icecastNumber(parsed *gabs.Container, path string) (float64, bool) {
	value, ok := parsed.Path(path).Data().(float64)
	return value, ok
}

// MP3 and AAC mounts publish no bitrate field of their own, only an audio_info
// string of the form "channels=2;samplerate=44100;bitrate=192". Reading the
// kbps back out of it is the only way those mounts report a bitrate at all.
func icecastAudioInfoBitrate(parsed *gabs.Container) (float64, bool) {
	info, ok := icecastString(parsed, "icestats.source.audio_info")
	if !ok {
		return 0, false
	}
	for _, field := range strings.Split(info, ";") {
		name, value, found := strings.Cut(field, "=")
		if !found || name != "bitrate" {
			continue
		}
		bitrate, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return 0, false
		}
		return bitrate, true
	}
	return 0, false
}

// Reads stream state out of an Icecast status-json.xsl document, returning the
// updated info and whether anything is playing. Fields Icecast omits keep the
// value they already had rather than resetting, so a partial document does not
// blank out good data.
func parseIcecastStatus(parsed *gabs.Container, current RadioInfo) (RadioInfo, bool) {
	title, titleOK := icecastString(parsed, "icestats.source.title")
	if !titleOK {
		return current, false
	}
	artist, artistOK := icecastString(parsed, "icestats.source.artist")
	if !artistOK {
		// Only Ogg carries structured tags that Icecast can split into separate
		// artist and title fields. MP3 and AAC mounts carry ICY metadata, which
		// is a single "Artist - Title" string and no artist key at all, so
		// requiring both fields reports a perfectly healthy stream as silent.
		var found bool
		artist, title, found = strings.Cut(title, " - ")
		if !found {
			return current, false
		}
	}
	current.Song.Artist = artist
	current.Song.Title = title

	if host, ok := icecastString(parsed, "icestats.host"); ok {
		current.Host = host
	}
	if mountpoint, ok := icecastString(parsed, "icestats.source.server_name"); ok {
		current.Mountpoint = mountpoint
	}
	if listeners, ok := icecastNumber(parsed, "icestats.source.listeners"); ok {
		current.Listeners = listeners
	}

	// Icecast 2.4 published the source bitrate in kbps as "bitrate". 2.5 dropped
	// that key for "ice-bitrate", also kbps, alongside "audio_bitrate" in bps.
	// Read whichever the server offers so both releases are supported.
	if bitrate, ok := icecastNumber(parsed, "icestats.source.bitrate"); ok {
		current.Bitrate = bitrate
	} else if bitrate, ok := icecastNumber(parsed, "icestats.source.ice-bitrate"); ok {
		current.Bitrate = bitrate
	} else if bitrate, ok := icecastNumber(parsed, "icestats.source.audio_bitrate"); ok {
		current.Bitrate = bitrate / 1000
	} else if bitrate, ok := icecastAudioInfoBitrate(parsed); ok {
		current.Bitrate = bitrate
	}

	return current, true
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
		info, playing := parseIcecastStatus(jsonParsed, nowPlaying())
		if !playing {
			slog.Debug("Connected to Icecast, but saw nothing playing.", "func", "icecastMonitor")
			icecastDataReset()
			return
		}

		radioMutex.Lock()
		now = info
		radioMutex.Unlock()

		if (prev.Song.Title != now.Song.Title) || (prev.Song.Artist != now.Song.Artist) {
			slog.Info(fmt.Sprintf("Now Playing: %s by %s", now.Song.Title, now.Song.Artist), "func", "icecastMonitor")
			// Dump the artwork rate limiter database first thing before updates
			// are sent out to reset artwork request count.
			dbr.RateLimitArt.FlushDB(ctx)

			radiodata_sse.Send("title", now.Song.Title)
			radiodata_sse.Send("artist", now.Song.Artist)
			if (prev.Song.Title != "") && (prev.Song.Artist != "") {
				radioMutex.Lock()
				history = append(history, playRecord{Title: prev.Song.Title, Artist: prev.Song.Artist, Ended: time.Now()})
				if len(history) > 10 {
					history = history[1:]
				}
				radioMutex.Unlock()
				radiodata_sse.Send("history", "update")
			}
		}
		if (prev.Host != now.Host) || (prev.Mountpoint != now.Mountpoint) {
			slog.Info(fmt.Sprintf("Audio stream on: <%s/%s>", now.Host, now.Mountpoint), "func", "icecastMonitor")
			radiodata_sse.Send("listenurl", now.Host+"/"+now.Mountpoint)
		}
		if prev.Listeners != now.Listeners {
			slog.Info(fmt.Sprintf("Listener count: <%v>", now.Listeners), "func", "icecastMonitor")
			radiodata_sse.Send("listeners", fmt.Sprint(now.Listeners))
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
