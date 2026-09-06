// stream.go
// Same-origin delivery of the Icecast audio stream.

package main

import (
	"io"
	"net/http"
	"strconv"
	"strings"
)

// The path the audio stream is served from, relative to the site itself.
const streamPrefix = "/stream/"

// Proxies the Icecast mount through Cadence so the audio comes from the same
// origin as the page.
//
// The client used to be handed a bare "host/mount" and told to prepend
// location.protocol. That dropped the port, so any Icecast not on the page's
// default port was unreachable; it assumed the scheme, so an HTTPS page
// pointing at a plain-HTTP stream host was blocked as mixed content; and it
// put the audio on a second hostname needing its own certificate. A relative
// path has none of those failure modes, and it works whether or not the
// bundled nginx is deployed.
func Stream() http.Handler {
	return http.StripPrefix(strings.TrimSuffix(streamPrefix, "/"), http.HandlerFunc(serveListener))
}

// The path a browser should play, or "-/-" when nothing is being broadcast.
// The sentinel is what clients already test against to decide whether the
// player is connected.
func listenPath() string {
	mountpoint := nowPlaying().Mountpoint
	if mountpoint == "" || mountpoint == "-" {
		return "-/-"
	}
	return streamPrefix + mountpoint
}

// Streams audio to one listener from the hub the source feeds.
func serveListener(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	listener, opening, ok := hub.subscribe()
	if !ok {
		// Nothing is broadcasting, so there is no stream to hold open.
		http.Error(w, "no source connected", http.StatusServiceUnavailable) // 503
		return
	}
	defer hub.unsubscribe(listener)

	_, contentType := hub.status()
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "no-cache, no-store")
	w.Header().Set("Connection", "close")
	// Live audio must not be buffered anywhere along the way, or the listener
	// hears nothing until a buffer happens to fill.
	w.Header().Set("X-Accel-Buffering", "no")

	// Players that want track names ask for them; browsers never do. Only a
	// client that asked gets metadata woven into its audio, so the stream a
	// browser receives is unaffected by any of it.
	var sink io.Writer = w
	if r.Header.Get("Icy-MetaData") == "1" {
		w.Header().Set("icy-metaint", strconv.Itoa(icyMetaInt))
		w.Header().Set("icy-name", nowPlaying().Mountpoint)
		if bitrate := nowPlaying().Bitrate; bitrate > 0 {
			w.Header().Set("icy-br", strconv.Itoa(int(bitrate)))
		}
		sink = newICYWriter(w, icyTitle)
	}
	w.WriteHeader(http.StatusOK)

	if len(opening) > 0 {
		if _, err := sink.Write(opening); err != nil {
			return
		}
		flusher.Flush()
	}

	for {
		select {
		case chunk, open := <-listener:
			if !open {
				return
			}
			if _, err := sink.Write(chunk); err != nil {
				return
			}
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}
