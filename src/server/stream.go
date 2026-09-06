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
