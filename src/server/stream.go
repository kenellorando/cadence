// stream.go
// Same-origin delivery of the Icecast audio stream.

package main

import (
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
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
	if c.InternalStream {
		return http.HandlerFunc(serveListener)
	}
	target := &url.URL{Scheme: "http", Host: c.IcecastAddress + c.IcecastPort}
	proxy := httputil.NewSingleHostReverseProxy(target)

	// Audio is an endless response. Without an explicit flush interval the
	// proxy buffers, and the listener hears nothing until the buffer fills.
	proxy.FlushInterval = -1

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		slog.Error("Couldn't proxy the audio stream.", "func", "Stream", "error", err)
		w.WriteHeader(http.StatusBadGateway) // 502 Bad Gateway
	}

	return http.StripPrefix(strings.TrimSuffix(streamPrefix, "/"), proxy)
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
	w.WriteHeader(http.StatusOK)

	if len(opening) > 0 {
		if _, err := w.Write(opening); err != nil {
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
			if _, err := w.Write(chunk); err != nil {
				return
			}
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}
