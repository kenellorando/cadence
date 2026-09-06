// api_actions.go
// API interactions for Postgres, Liquidsoap, and the audio source.

package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"net"

	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Every outbound call in this file talks to a service on the same compose
// network. None of them are allowed to hang: the monitor polls on a one second
// cadence, and a stalled request would silently stop all stream updates.
const serviceTimeout = 5 * time.Second

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

// Takes a song ID integer.
// Returns the absolute path of the audio file.
// Finds the library row for a song the source has announced. Album breaks ties
// where a title and artist appear more than once, and the id ordering makes the
// result deterministic when even that is ambiguous.
func resolvePlaying(title, artist, album string) (SongData, bool) {
	title, artist, album = strings.TrimSpace(title), strings.TrimSpace(artist), strings.TrimSpace(album)
	if title == "" || dbp == nil {
		return SongData{}, false
	}
	statement := fmt.Sprintf(`SELECT id, artist, title, album, genre, year FROM %s
		WHERE title = $1 AND artist = $2
		ORDER BY (album = $3) DESC, id ASC
		LIMIT 1`, c.PostgresTableName)

	song := SongData{}
	err := dbp.QueryRow(statement, title, artist, album).Scan(
		&song.ID, &song.Artist, &song.Title, &song.Album, &song.Genre, &song.Year)
	if err != nil {
		slog.Debug("Could not resolve the playing song in the library.", "func", "resolvePlaying", "error", err)
		return SongData{}, false
	}
	return song, true
}

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
		// A bulk change to the library (copying an album in, a sync tool running)
		// arrives as a burst of events, and rescanning per event means walking the
		// whole library dozens of times over. Wait for the burst to go quiet and
		// rescan once.
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

// Reads the kbps out of an audio_info string. The built-in source receives the
// same format in an ice-audio-info header.
func audioInfoBitrate(info string) (float64, bool) {
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

var previous = RadioInfo{}

// Publishes a new view of the radio and announces whatever changed. Both the
// Icecast poller and the built-in audio source feed this, so a song change
// raises the same events and the same history entry either way.
func applyRadioInfo(info RadioInfo) {
	radioMutex.Lock()
	prev := previous
	now = info
	previous = info
	radioMutex.Unlock()

	if (prev.Song.Title != info.Song.Title) || (prev.Song.Artist != info.Song.Artist) {
		slog.Info(fmt.Sprintf("Now Playing: %s by %s", info.Song.Title, info.Song.Artist), "func", "applyRadioInfo")
		// Clear artwork allowances before the update goes out: the artwork has
		// changed with the song, so every client may fetch the new one.
		rateLimitResetArt()

		radiodata_sse.Send("title", info.Song.Title)
		radiodata_sse.Send("artist", info.Song.Artist)
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
	if (prev.Host != info.Host) || (prev.Mountpoint != info.Mountpoint) {
		slog.Info(fmt.Sprintf("Audio stream on: <%s/%s>", info.Host, info.Mountpoint), "func", "applyRadioInfo")
		radiodata_sse.Send("listenurl", listenPath())
	}
	if prev.Listeners != info.Listeners {
		slog.Info(fmt.Sprintf("Listener count: <%v>", info.Listeners), "func", "applyRadioInfo")
		radiodata_sse.Send("listeners", fmt.Sprint(info.Listeners))
	}
}

// Records the song the built-in source just announced. Title, artist and album
// arrive as separate fields, so nothing has to be recovered from a combined
// string the way an Icecast status document forces.
func setNowPlaying(title, artist, album string) {
	info := nowPlaying()
	info.Song.Title = title
	info.Song.Artist = artist
	info.Song.Album = album
	// Resolve to a library row once, here, rather than matching these strings
	// again on every request for metadata or artwork. Holding the id is what
	// lets artwork be fetched by identity instead of by search.
	info.Song.ID = 0
	if song, ok := resolvePlaying(title, artist, album); ok {
		info.Song = song
	} else {
		slog.Warn(fmt.Sprintf("Playing a song that is not in the library: %s by %s", title, artist), "func", "setNowPlaying")
	}
	applyRadioInfo(info)
}

// Records the mount the built-in source connected on.
func setStreamMount(mountpoint, audioInfo string) {
	info := nowPlaying()
	info.Mountpoint = mountpoint
	if bitrate, ok := audioInfoBitrate(audioInfo); ok {
		info.Bitrate = bitrate
	}
	applyRadioInfo(info)
}

// Forgets the mount when the source disconnects, so the player reports itself
// offline rather than pointing at a stream nothing is feeding.
func clearStreamMount() {
	info := nowPlaying()
	info.Mountpoint = "-"
	info.Song = SongData{Title: "-", Artist: "-"}
	info.Listeners = -1
	applyRadioInfo(info)
}

// Reports the listener count the built-in fan-out is currently serving.
func setListeners(count int) {
	info := nowPlaying()
	info.Listeners = float64(count)
	applyRadioInfo(info)
}

var history = make([]playRecord, 0, 10)

type playRecord struct {
	Title  string
	Artist string
	Ended  time.Time
}
