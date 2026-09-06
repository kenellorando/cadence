// source.go
// Accepts the Icecast source protocol, so the audio source connects to Cadence
// directly instead of to a separate streaming server.

package main

import (
	"bufio"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"strings"
	"time"
)

const (
	// A source that has connected but not finished its request line and headers
	// is holding a connection for nothing.
	sourceHandshakeTimeout = 10 * time.Second
	// Audio arrives continuously, so a gap this long means the source is gone
	// even though the socket has not closed.
	sourceReadTimeout = 30 * time.Second
	sourceReadBuffer  = 16 * 1024
	// The username Icecast source clients authenticate with.
	sourceUser = "source"
	// Used when CSERVER_SOURCEPORT is unset.
	defaultSourcePort = ":8081"
)

// The one connected audio source, and the listeners it feeds.
var hub = newAudioHub()

// Listens for the audio source. This is deliberately a plain TCP listener
// rather than an http.Server: the source sends "SOURCE /mount HTTP/1.0" with no
// Content-Length and no chunked encoding, and net/http gives a handler an empty
// body for such a request, so the audio would never arrive.
func serveSource(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			slog.Debug("Source listener stopped accepting.", "func", "serveSource", "error", err)
			return
		}
		go handleSourceConn(conn)
	}
}

func handleSourceConn(conn net.Conn) {
	defer conn.Close()

	if err := conn.SetReadDeadline(time.Now().Add(sourceHandshakeTimeout)); err != nil {
		return
	}
	reader := bufio.NewReaderSize(conn, sourceReadBuffer)

	requestLine, err := reader.ReadString('\n')
	if err != nil {
		slog.Debug("Source connection closed before sending a request.", "func", "handleSourceConn", "error", err)
		return
	}
	method, target, ok := parseRequestLine(requestLine)
	if !ok {
		_ = writeSourceStatus(conn, "400 Bad Request")
		return
	}
	headers, err := readHeaders(reader)
	if err != nil {
		_ = writeSourceStatus(conn, "400 Bad Request")
		return
	}
	if !sourceAuthorized(headers["authorization"]) {
		slog.Warn("Rejecting a source connection with bad credentials.", "func", "handleSourceConn", "remote", conn.RemoteAddr().String())
		_ = writeSourceStatus(conn, "401 Unauthorized")
		return
	}

	switch {
	case method == "SOURCE" || method == "PUT":
		acceptSource(conn, reader, target, headers)
	case method == "GET" && strings.HasPrefix(target, "/admin/metadata"):
		// Metadata arrives on its own connection rather than inside the audio, so
		// title, artist and album stay separate fields. Reading them here is what
		// removes the need to recover the playing song by matching a string a
		// streaming server echoed back.
		applySourceMetadata(target)
		_ = writeSourceStatus(conn, "200 OK")
	default:
		_ = writeSourceStatus(conn, "405 Method Not Allowed")
	}
}

// Reads audio until the source disconnects.
func acceptSource(conn net.Conn, reader *bufio.Reader, mount string, headers map[string]string) {
	contentType := headers["content-type"]
	if contentType == "" {
		contentType = "audio/mpeg"
	}

	// The source waits for this before sending a byte of audio.
	if err := writeSourceStatus(conn, "200 OK"); err != nil {
		return
	}

	mountpoint := strings.TrimPrefix(mount, "/")
	slog.Info(fmt.Sprintf("Audio source connected on mount <%s>.", mountpoint), "func", "acceptSource", "type", contentType)

	suspendTrackClock()
	setStreamMount(mountpoint, headers["ice-audio-info"])
	hub.sourceConnected(contentType)
	defer func() {
		hub.sourceDisconnected()
		clearStreamMount()
		slog.Info("Audio source disconnected.", "func", "acceptSource")
	}()

	buffer := make([]byte, sourceReadBuffer)
	for {
		if err := conn.SetReadDeadline(time.Now().Add(sourceReadTimeout)); err != nil {
			return
		}
		n, err := reader.Read(buffer)
		if n > 0 {
			hub.publish(buffer[:n])
		}
		if err != nil {
			slog.Debug("Audio source read ended.", "func", "acceptSource", "error", err)
			return
		}
	}
}

// Splits "SOURCE /cadence1 HTTP/1.0" into its method and target.
func parseRequestLine(line string) (method, target string, ok bool) {
	fields := strings.Fields(strings.TrimSpace(line))
	if len(fields) < 2 {
		return "", "", false
	}
	return strings.ToUpper(fields[0]), fields[1], true
}

// Reads headers up to the blank line, keyed by lowercased name.
func readHeaders(reader *bufio.Reader) (map[string]string, error) {
	headers := make(map[string]string)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			return headers, nil
		}
		name, value, found := strings.Cut(line, ":")
		if !found {
			continue
		}
		headers[strings.ToLower(strings.TrimSpace(name))] = strings.TrimSpace(value)
	}
}

// Checks HTTP Basic credentials against the configured source password.
func sourceAuthorized(header string) bool {
	if c.SourcePassword == "" {
		slog.Error("No source password is configured; refusing every source connection.", "func", "sourceAuthorized")
		return false
	}
	encoded, found := strings.CutPrefix(header, "Basic ")
	if !found {
		return false
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(encoded))
	if err != nil {
		return false
	}
	user, password, found := strings.Cut(string(decoded), ":")
	if !found {
		return false
	}
	// Constant time, so a wrong password cannot be found a character at a time.
	userOK := subtle.ConstantTimeCompare([]byte(user), []byte(sourceUser)) == 1
	passwordOK := subtle.ConstantTimeCompare([]byte(password), []byte(c.SourcePassword)) == 1
	return userOK && passwordOK
}

// Reads title, artist and album out of an /admin/metadata request and updates
// the radio state.
func applySourceMetadata(target string) {
	parsed, err := url.Parse(target)
	if err != nil {
		slog.Debug("Could not parse a metadata request.", "func", "applySourceMetadata", "error", err)
		return
	}
	query := parsed.Query()
	title := strings.TrimSpace(query.Get("title"))
	artist := strings.TrimSpace(query.Get("artist"))
	album := strings.TrimSpace(query.Get("album"))

	// Older source clients send only the combined "song" field.
	if title == "" {
		if song := strings.TrimSpace(query.Get("song")); song != "" {
			if splitArtist, splitTitle, found := strings.Cut(song, " - "); found {
				artist, title = splitArtist, splitTitle
			} else {
				title = song
			}
		}
	}
	if title == "" {
		return
	}
	setNowPlaying(title, artist, album)
}

func writeSourceStatus(conn net.Conn, status string) error {
	if err := conn.SetWriteDeadline(time.Now().Add(sourceHandshakeTimeout)); err != nil {
		return err
	}
	_, err := fmt.Fprintf(conn, "HTTP/1.0 %s\r\n\r\n", status)
	return err
}

// Opens the source listener and starts reporting listener numbers. The port is
// deliberately separate from the site's own, so it can be left unpublished:
// the source password should not be reachable from the internet.
func startAudioSource() {
	port := c.SourcePort
	if port == "" {
		port = defaultSourcePort
	}
	listener, err := net.Listen("tcp", port)
	if err != nil {
		slog.Error("Couldn't open the audio source port.", "func", "startAudioSource", "port", port, "error", err)
		return
	}
	slog.Info(fmt.Sprintf("Listening for an audio source on <%s>.", port), "func", "startAudioSource")
	go serveSource(listener)
	go trackListenerCount()
	go trackProgressMonitor()
}

// Publishes the listener count as it changes. Nothing pushes this, so it is
// sampled -- but from a number held in memory rather than an HTTP poll of
// another service.
func trackListenerCount() {
	last := -1
	for {
		time.Sleep(time.Second)
		if live, _ := hub.status(); !live {
			continue
		}
		if count := hub.listenerCount(); count != last {
			last = count
			setListeners(count)
		}
	}
}
