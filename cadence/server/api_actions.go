// api_actions.go
// API function repeatable actions.

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/Jeffail/gabs"
	"github.com/RediSearch/redisearch-go/redisearch"
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
	results, _, _ := r.Metadata.Search(redisearch.NewQuery(" %" + query + "% "))
	for _, song := range results {
		var songData SongData
		// todo: error handle marshal/unmarshal
		songBytes, _ := json.Marshal(song.Properties)
		_ = json.Unmarshal(songBytes, &songData)
		queryResults = append(queryResults, songData)
	}
	return queryResults, nil
}

// Takes a title and artist string to find a song which exactly matches.
// Returns a slice of SongData of songs by relevance.
// This search should only have one result unless multiple audio files share the exact same title and artist.
func searchByTitleArtist(title string, artist string) (queryResults []SongData, err error) {
	results, _, _ := r.Metadata.Search(redisearch.NewQuery(title+" "+artist).Limit(0, 1))
	for _, song := range results {
		var songData SongData
		// todo: error handle marshal/unmarshal
		songBytes, _ := json.Marshal(song.Properties)
		_ = json.Unmarshal(songBytes, &songData)
		queryResults = append(queryResults, songData)
	}
	return queryResults, nil
}

// Takes a song ID integer.
// Returns the absolute path of the audio file.
func getPathById(id int) (path string, err error) {
	clog.Info("getPathById", fmt.Sprintf("Searching database for the path of song ID <%v>", id))
	result, err := r.Metadata.Get(fmt.Sprint(id))
	if err != nil {
		clog.Error("getPathById", "Database search failed.", err)
		return "", err
	}
	return fmt.Sprint(result.Properties["Path"]), nil
}

// Takes an absolute song path, submits the path to be queued in Liquidsoap.
// Returns the response message from Liquidsoap.
func liquidsoapRequest(path string) (message string, err error) {
	// Telnet to liquidsoap
	clog.Debug("liquidsoapRequest", "Connecting to liquidsoap service...")
	conn, err := net.Dial("tcp", c.SourceAddress+c.SourcePort)
	if err != nil {
		clog.Error("liquidsoapRequest", "Failed to connect to audio source server.", err)
		return "", err
	}
	defer conn.Close()

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
	if err != nil {
		clog.Error("liquidsoapRequest", "Failed to connect to audio source server.", err)
		return "", err
	}
	defer conn.Close()
	fmt.Fprintf(conn, "cadence1.skip\n")
	// Listen for response
	message, _ = bufio.NewReader(conn).ReadString('\n')
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
				dbPopulate()
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
	go func() {
		for {
			time.Sleep(1 * time.Second)
			resp, err := http.Get("http://" + c.StreamAddress + c.StreamPort + "/status-json.xsl")
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
			resp.Body.Close()
		}
	}()
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
