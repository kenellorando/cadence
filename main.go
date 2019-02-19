package main

import (
	"database/sql"
	"fmt"
	"net/http"

	env "github.com/deanishe/go-env"
	"github.com/gorilla/mux"
	"github.com/kenellorando/clog"
)

// Declare full config object
var c = Config{}

// Config - Primary configuration object holder
type Config struct {
	server CConfig
	db     DBConfig
}

// CConfig - CServer configuration
type CConfig struct {
	LogLevel int
	Port     string
	MusicDir string
}

// DBConfig - Database configuration
type DBConfig struct {
	Host string
	Port string
	User string
	Pass string
	Name string
}

// Init function performs prep work with configurations that
// need to be known *before* starting main
func init() {
	// Get server-related configs
	server := CConfig{}
	server.LogLevel = env.GetInt("CSERVER_LOGLEVEL", 5)
	server.Port = env.GetString("CSERVER_WEB_PORT", ":8000")
	server.MusicDir = env.GetString("CSERVER_MUSIC_DIR", "/Default/Fake/Music/Dir")
	c.server = server

	// Get database related configs
	db := DBConfig{}
	db.Host = env.GetString("CSERVER_DB_HOST", "Default_FakeDBHost")
	db.Port = env.GetString("CSERVER_DB_PORT", "Default_FakeDBPort")
	db.User = env.GetString("CSERVER_DB_USER", "Default_FakeDBUser")
	db.Pass = env.GetString("CSERVER_DB_PASS", "Default_FakeDBPass")
	db.Name = env.GetString("CSERVER_DB_NAME", "Default_FakeDBName")
	c.db = db

	clog.Info("init", "Server initialized.")
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
	clog.Info("main", fmt.Sprintf("Starting server on port %s ...", c.server.Port))
	clog.Fatal("main", "Server failed to start!", http.ListenAndServe(c.server.Port, r))
}

// Establishes database connection using configuration
func connectDatabase(db DBConfig) (*sql.DB, error) {
	clog.Debug("connectDatabase", "Attempting connection to database...")

	// Form a connection with the database using config
	connectInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s", db.Host, db.Port, db.User, db.Pass, db.Name)
	database, err := sql.Open("postgres", connectInfo)
	if err != nil {
		clog.Error("connectDatabase", "Connection to the database failed!", err)
	} else {
		clog.Info("connectDatabase", "Connected to the database.")
	}

	return database, err
}
