package main

import (
	"fmt"
	"net/http"

	env "github.com/deanishe/go-env"
	"github.com/gorilla/mux"
	"github.com/kenellorando/clog"
)

// Init function performs prep work with configurations that
// need to be known *before* starting main
func init() {
	// Init defaults are declared here
	// Although all configurations are set into environment variables
	// We camn declare some defaults here for server initialization purposes
	const defaultLogLevel = 4

	// Todo:
	// make a reusable function to check if a value in the environment variable is blank

	// Set logging to log level
	logLevel := env.GetInt("CSERVER_LOGLEVEL")
	logPointer := &logLevel
	// If no log level is set, default to a value
	// Else, send the value
	if logPointer == nil {
		clog.Init(4)
		clog.Info("init", fmt.Sprintf("No default logging level was found. Using default level 4"))
	} else {
		clog.Init(logLevel)
		clog.Info("init", fmt.Sprintf("Setting logging service verbosity to <%v>", logLevel))
	}
}

func main() {
	// Get all other configurations
	c := getConfigs()

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
	clog.Info("main", fmt.Sprintf("Starting server on port %s ...", c.server.Port))
	clog.Fatal("main", "Server failed to start!", http.ListenAndServe(c.server.Port, r))
}
