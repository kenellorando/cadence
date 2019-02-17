package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kenellorando/clog"
)

func main() {
	c, db := getConfig()
	initLogger(c.LogLevel)
	initDatabase(db)

	// Prepare frontend for serving

	r := mux.NewRouter()

	// Serve other specific routes next
	r.HandleFunc("/", ServeRoot).Methods("GET")

	// For everything else, serve 404
	r.NotFoundHandler = http.HandlerFunc(Serve404)

	// Start server
	clog.Debug("main", "ListenAndServe starting...")
	clog.Fatal("main", "ListenAndServe failed to start.", http.ListenAndServe(":8000", r))
}
