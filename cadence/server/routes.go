// routes.go
// Rate limiter and router.

package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/kenellorando/clog"
	"gopkg.in/antage/eventsource.v1"
)

var limit_map = make(map[string]time.Time)

func rateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		expiration, ok := limit_map[ip]

		if !ok {
			clog.Debug("rateLimit", fmt.Sprintf("<%s> is a new IP.", ip))
			limit_map[ip] = time.Now().Add(time.Second * time.Duration(c.RequestRateLimit))
			next.ServeHTTP(w, r)
		} else {
			if expiration.Before(time.Now()) {
				clog.Debug("rateLimit", fmt.Sprintf("<%s> is an existing IP that is no longer rate limited.", ip))
				limit_map[ip] = time.Now().Add(time.Second * time.Duration(c.RequestRateLimit))
				next.ServeHTTP(w, r)
			} else {
				clog.Debug("rateLimit", fmt.Sprintf("<%s> is rate limited.", ip))
				w.WriteHeader(http.StatusTooManyRequests) // 429 Too Many Requests
				return
			}
		}
	})
}

var radiodata_sse = eventsource.New(nil, nil)

func routes() *http.ServeMux {
	r := http.NewServeMux()
	r.Handle("/api/radiodata/sse", radiodata_sse)
	r.Handle("/api/search", Search())
	r.Handle("/api/request/id", rateLimit(RequestID()))
	r.Handle("/api/request/bestmatch", rateLimit(RequestBestMatch()))
	r.Handle("/api/nowplaying/metadata", NowPlayingMetadata())
	r.Handle("/api/nowplaying/albumart", NowPlayingAlbumArt())
	r.Handle("/api/history", History())
	r.Handle("/api/listenurl", ListenURL())
	r.Handle("/api/listeners", Listeners())
	r.Handle("/api/bitrate", Bitrate())
	r.Handle("/api/version", Version())
	r.Handle("/ready", Ready())
	if c.DevMode {
		r.Handle("/api/dev/skip", DevSkip())
	}
	r.Handle("/", http.FileServer(http.Dir(c.RootPath+"./public/")))
	return r
}
