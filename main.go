package main

import (
	"database/sql"
	"fmt"
	"net/http"

	env "github.com/deanishe/go-env"
	"github.com/gorilla/mux"
	"github.com/kenellorando/clog"
)

// Declare globally accessible data
var c = Config{}     // Full configuration object
var database *sql.DB // Database object

// Config - Primary configuration object holder
type Config struct {
	server ServerConfig
	db     DBConfig
	schema SchemaConfig
}

// ServerConfig - Webserver configuration
type ServerConfig struct {
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
	// Declare substructs of the global config
	server := ServerConfig{}
	db := DBConfig{}
	schema := SchemaConfig{}

	// Webserver configuration
	server.LogLevel = env.GetInt("CSERVER_LOGLEVEL", 5)
	server.Port = env.GetString("CSERVER_WEB_PORT", ":8080")
	server.MusicDir = env.GetString("CSERVER_MUSIC_DIR", "/Default/Fake/Music/Dir")
	// Database server configuration
	db.Host = env.GetString("CSERVER_DB_HOST", "localhost")
	db.Port = env.GetString("CSERVER_DB_PORT", "5432")
	db.User = env.GetString("CSERVER_DB_USER", "Default_DBUser_SetEnvVar!")
	db.Pass = env.GetString("CSERVER_DB_PASS", "Default_DBPass_SetEnvVar!")
	db.SSLMode = env.GetString("CSERVER_DB_SSLMODE", "disable")
	db.Driver = env.GetString("CSERVER_DB_DRIVER", "postgres")
	db.DSN = fmt.Sprintf("host=%s port=%s user=%s password=%s sslmode=%s", db.Host, db.Port, db.User, db.Pass, db.SSLMode)
	// Database schema configuration
	schema.Name = env.GetString("CSERVER_DB_NAME", "Default_DBName_SetEnvVar!")

	// Set the substructs of the global config
	c.server = server
	c.db = db
	c.schema = schema

	// Initialize logging
	clog.Init(c.server.LogLevel)
	clog.Info("init", fmt.Sprintf("Logging service initialized to level <%v>", c.server.LogLevel))

	// Establish a connection to the database
	clog.Info("init", fmt.Sprintf("Establishing a connection to database server <%s:%s>", c.db.Host, c.db.Port))
	newDatabase, err := databaseConnect()
	if err != nil {
		clog.Warn("init", fmt.Sprintf("Initial test connection to the database server failed! Future database requests will also fail."))
		clog.Debug("init", "Skipping data check.")
	} else {
		// Set the global database object to the newly made pointer
		// Start the database data check
		clog.Debug("init", "Initial test connection to database server succeeded. Starting data check...")
		database = newDatabase
		databaseCheck()
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
	clog.Info("main", fmt.Sprintf("Starting webserver on port <%s>.", c.server.Port))
	clog.Fatal("main", "Server failed to start!", http.ListenAndServe(c.server.Port, r))
}
