package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kenellorando/clog"
)

func main() {
	// Grab configuration
	// Start the logger and connect to the database
	c := getCConfig()
	initLogger(c.LogLevel)

	// Perform database initialization
	db := getDBConfig()
	initDatabase(db)

	// Handle routes
	r := mux.NewRouter()
	// List API routes first
	r.HandleFunc("/api/aria1/search", ARIA1Search).Methods("POST")
	r.HandleFunc("/api/aria1/request", ARIA1Request).Methods("POST")

	// Serve other specific routes next
	r.HandleFunc("/", ServeRoot).Methods("GET")

	// For everything else, serve 404
	r.NotFoundHandler = http.HandlerFunc(Serve404)

	// Start server
	clog.Debug("main", "ListenAndServe starting...")
	clog.Fatal("main", "ListenAndServe failed to start.", http.ListenAndServe(":8000", r))
}
