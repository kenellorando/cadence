package main

import (
	"fmt"
	"net/http"
	"os"

	env "github.com/deanishe/go-env"
	"github.com/gorilla/mux"
	"github.com/kenellorando/clog"
)

// Init function performs prep work with configurations that
// need to be known *before* starting main
func init() {
	// Init defaults are declared here
	// Although all configurations are set into environment variables
	// We can declare some defaults here for server initialization purposes
	const defaultLogLevel = 5

	// Set logging level
	// If there is no data, default to defaultLogLevel
	// Else, use the value.
	logLevel := os.Getenv("CSERVER_LOGLEVEL")
	if len(logLevel) == 0 {
		clog.Info("init", fmt.Sprintf("No default logging level was found."))
		clog.Init(defaultLogLevel)
		clog.Info("init", fmt.Sprintf("Logging level set to hardcoded default level <%v>", defaultLogLevel))
	} else {
		logLevel := env.GetInt("CSERVER_WEB_PORT")
		clog.Init(logLevel)
		clog.Info("init", fmt.Sprintf("Set logging verbosity to <%v>", logLevel))
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
