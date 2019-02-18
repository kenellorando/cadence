package main

import (
	"net/http"

	env "github.com/deanishe/go-env"
	"github.com/gorilla/mux"
	"github.com/kenellorando/clog"
)

const (
	defaultLogLevel = 4
)

// Init function performs prep work for some configurations
func init() {
	// Check the environment variable for a log level
	logLevel := env.GetInt("CSERVER_LOGLEVEL")
	logPointer := &logLevel
	// If no log level is set, default to a value
	// Else, send the value
	if logPointer == nil {
		clog.Init(defaultLogLevel)
	} else {
		clog.Init(logLevel)
	}
}

func main() {
	logLevel := env.GetInt("CSERVER_LOGLEVEL")
	logPoint := &logLevel
	if logPoint == nil {

	}

	// Get configuration variables (all set in environment)
	//c := getConfigs()

	// Initialize logger with

	// Test database connection before launching server

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
