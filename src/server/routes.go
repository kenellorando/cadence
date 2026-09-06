// routes.go
// Routes incoming requests to functions written in api.go

package main

import (
	"net/http"
)

var radiodata_sse = newEventStream()

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
	r.Handle("/", http.FileServer(http.Dir(c.RootPath+"./public/")))
	return r
}
