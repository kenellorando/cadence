package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

func routes() *mux.Router {
	// Handle routes
	r := mux.NewRouter()

	// Subdomains, if needed
	/*
		s := r.Host("docs." + c.server.Domain + c.server.Port).Subrouter()
		s.PathPrefix("/").Handler(http.FileServer(http.Dir("./public/docs"))).Methods("GET")
	*/

	// List API routes first
	r.HandleFunc("/api/aria1/search", ARIA1Search).Methods("POST")
	r.HandleFunc("/api/aria1/request", ARIA1Request).Methods("POST")
	r.HandleFunc("/api/aria1/library", ARIA1Library).Methods("GET")

	// Aria2
	r.HandleFunc("/api/aria2/request", ARIA2Request).Methods("POST")

	// Serve other specific routes next
	r.HandleFunc("/", handleServeRoot()).Methods("GET")
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./public/static/")))).Methods("GET")
	r.PathPrefix("/css/").Handler(http.StripPrefix("/css/", http.FileServer(http.Dir("./public/css/")))).Methods("GET")
	r.PathPrefix("/js/").Handler(http.StripPrefix("/js/", http.FileServer(http.Dir("./public/js/")))).Methods("GET")

	// For everything else, serve 404
	r.NotFoundHandler = http.HandlerFunc(Serve404)

	return r
}
