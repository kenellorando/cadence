// routes.go
// Just the router

package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

func routes() *mux.Router {
	r := mux.NewRouter()

	// REST
	r.HandleFunc("/api/search", Search()).Methods("POST")
	r.HandleFunc("/api/request/id", RequestID()).Methods("POST")
	r.HandleFunc("/api/request/bestmatch", RequestBestMatch()).Methods("POST")
	r.HandleFunc("/api/nowplaying/metadata", NowPlayingMetadata()).Methods("GET")
	r.HandleFunc("/api/nowplaying/albumart", NowPlayingAlbumArt()).Methods("GET")
	r.HandleFunc("/api/version", Version()).Methods("GET")
	// WebSocket
	r.HandleFunc("/socket/radiodata", RadioData()).Methods("GET")
	// Health
	r.HandleFunc("/ready", Ready()).Methods("GET")

	// Site
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(c.RootPath+"./public/static/")))).Methods("GET")
	r.PathPrefix("/css/").Handler(http.StripPrefix("/css/", http.FileServer(http.Dir(c.RootPath+"./public/css/")))).Methods("GET")
	r.PathPrefix("/js/").Handler(http.StripPrefix("/js/", http.FileServer(http.Dir(c.RootPath+"./public/js/")))).Methods("GET")
	r.HandleFunc("/", SiteRoot()).Methods("GET")
	r.NotFoundHandler = http.HandlerFunc(Site404())
	return r
}
