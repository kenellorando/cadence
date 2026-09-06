// routes.go
// Routes incoming requests to functions written in api.go

package main

import (
	"net/http"
	"strings"
)

var radiodata_sse = newEventStream()

// Where the build puts files whose names contain a hash of their contents.
const immutablePrefix = "/_app/immutable/"

// Serves the web UI with cache headers the build's naming actually supports.
//
// Nothing set these before, so browsers fell back to heuristic caching and
// could hold on to the page without revalidating. That page names the hashed
// bundles, so a stale copy of it points at the previous build: the site keeps
// working and quietly stays one deploy behind.
//
// The hashed files can be kept forever, because a change to their contents
// changes their name. The page that names them must be revalidated every time.
func staticFiles() http.Handler {
	files := http.FileServer(http.Dir(c.RootPath + "./public/"))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, immutablePrefix) {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			// Not "no-store": revalidation is cheap and usually answers 304.
			w.Header().Set("Cache-Control", "no-cache")
		}
		files.ServeHTTP(w, r)
	})
}

func routes() *http.ServeMux {
	r := http.NewServeMux()
	r.Handle("/api/radiodata/sse", radiodata_sse)
	r.Handle("/api/search", Search())
	r.Handle("/api/request/id", rateLimitRequest(RequestID()))
	r.Handle("/api/request/bestmatch", rateLimitRequest(RequestBestMatch()))
	r.Handle("/api/nowplaying/metadata", NowPlayingMetadata())
	r.Handle("/api/nowplaying/progress", NowPlayingProgress())
	r.Handle("/api/nowplaying/albumart", rateLimitArt(NowPlayingAlbumArt()))
	r.Handle("GET /api/song/{id}/art", rateLimitSongArt(SongArt()))
	r.Handle("/api/history", History())
	r.Handle("/api/library", Library())
	r.Handle("/api/listenurl", ListenURL())
	r.Handle("/api/listeners", Listeners())
	r.Handle("/api/bitrate", Bitrate())
	r.Handle("/api/version", Version())
	r.Handle("/ready", Ready())
	if c.DevMode {
		r.Handle("/api/dev/skip", DevSkip())
	}
	r.Handle(streamPrefix, Stream())
	r.Handle("/", staticFiles())
	return r
}
