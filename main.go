package main

import (
	"fmt"
	"net/http"

	env "github.com/deanishe/go-env"
	"github.com/gorilla/mux"
	"github.com/kenellorando/clog"
)

// Declare full config object globally accessible
var c = Config{}

// Config - Primary configuration object holder
type Config struct {
	server CConfig
	db     DBConfig
	schema SchemaConfig
}

// CConfig - Webserver configuration
type CConfig struct {
	LogLevel int
	Port     string
	MusicDir string
}

// DBConfig - Database server configuration
type DBConfig struct {
	Host    string
	Port    string
	User    string
	Pass    string
	SSLMode string
	Driver  string
	DSN     string
}

// SchemaConfig - Database schema configuration
type SchemaConfig struct {
	Name string
}

// Init function grabs configuration values for the server
// All configs are set in environment variables
// Default values for missing environment variables are set here.
// Init also initalizes other services with relevant values
func init() {
	// Webserver configuration
	server := CConfig{}
	server.LogLevel = env.GetInt("CSERVER_LOGLEVEL", 5)
	server.Port = env.GetString("CSERVER_WEB_PORT", ":8080")
	server.MusicDir = env.GetString("CSERVER_MUSIC_DIR", "/Default/Fake/Music/Dir")
	c.server = server

	// Database server configuration
	db := DBConfig{}
	db.Host = env.GetString("CSERVER_DB_HOST", "localhost")
	db.Port = env.GetString("CSERVER_DB_PORT", "5432")
	db.User = env.GetString("CSERVER_DB_USER", "Default_DBUser_SetEnvVar!")
	db.Pass = env.GetString("CSERVER_DB_PASS", "Default_DBPass_SetEnvVar!")
	db.SSLMode = env.GetString("CSERVER_DB_SSLMODE", "disable")
	db.Driver = env.GetString("CSERVER_DB_DRIVER", "postgres")
	db.DSN = fmt.Sprintf("host=%s port=%s user=%s password=%s sslmode=%s", db.Host, db.Port, db.User, db.Pass, db.SSLMode)
	c.db = db

	// Database schema configuration
	schema := SchemaConfig{}
	schema.Name = env.GetString("CSERVER_DB_NAME", "Default_DBName_SetEnvVar!")
	/*
		Todo: Add column names here
	*/
	c.schema = schema

	// Initialize logging
	clog.Init(c.server.LogLevel)
	clog.Info("init", fmt.Sprintf("Logging service initialized to level <%v>", c.server.LogLevel))

	// Test a connection to the database
	clog.Info("init", fmt.Sprintf("Testing a connection to database <%s:%s>", c.db.Host, c.db.Port))

	_, err := databaseConnect()
	if err != nil {
		clog.Warn("init", fmt.Sprintf("Initial test connection to the database server failed! Future server requests may also fail."))
	} else {
		clog.Info("init", "Initial test connection to database server and data succeeded.")
	}

}

func main() {
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
	clog.Info("main", fmt.Sprintf("Starting server on port `%s`.", c.server.Port))
	clog.Fatal("main", "Server failed to start!", http.ListenAndServe(c.server.Port, r))
}
