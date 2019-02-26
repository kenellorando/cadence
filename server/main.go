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
var database *sql.DB // Database abstraction interface

// Config - Primary configuration object holder
type Config struct {
	server ServerConfig
	db     DBConfig
	schema SchemaConfig
}

// ServerConfig - Webserver configuration
type ServerConfig struct {
	Domain   string
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
	Name    string
	SSLMode string
	Driver  string
	DSN     string
}

// SchemaConfig - Database schema configuration
type SchemaConfig struct {
	Table string
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

	// Webserver configurationi, err := strconv.Atoi(s)
	server.Domain = env.GetString("CSERVER_DOMAIN")
	server.LogLevel = env.GetInt("CSERVER_LOGLEVEL")
	server.Port = env.GetString("CSERVER_PORT")
	server.MusicDir = env.GetString("CSERVER_MUSIC_DIR")
	// Database server configuration
	db.Host = env.GetString("CSERVER_DB_HOST")
	db.Port = env.GetString("CSERVER_DB_PORT")
	db.User = env.GetString("CSERVER_DB_USER")
	db.Pass = env.GetString("CSERVER_DB_PASS")
	db.Name = env.GetString("CSERVER_DB_NAME")
	db.SSLMode = env.GetString("CSERVER_DB_SSLMODE")
	db.Driver = env.GetString("CSERVER_DB_DRIVER")
	db.DSN = fmt.Sprintf("host='%s' port='%s' user='%s' password='%s' sslmode='%s'", db.Host, db.Port, db.User, db.Pass, db.SSLMode)
	// Database schema configuration
	schema.Table = env.GetString("CSERVER_DB_TABLE")

	// Set the substructs of the global config
	c.server = server
	c.db = db
	c.schema = schema

	// Initialize logging
	clog.Level(c.server.LogLevel)
	clog.Info("init", fmt.Sprintf("Logging service initialized to level <%v>", c.server.LogLevel))

	// Establish a connection to the database
	clog.Debug("init", fmt.Sprintf("Establishing a connection to database server <%s:%s>", c.db.Host, c.db.Port))
	newDatabase, err := databaseConnect()
	if err != nil {
		clog.Warn("init", fmt.Sprintf("Database server connection test failed! Future database requests will also fail."))
		clog.Debug("init", "Skipping data check.")
	} else {
		// Set the global database object to the newly made pointer
		// Start the database data check
		clog.Info("init", "Database server connection test successful. Starting database auto configuration..")
		database = newDatabase
		err := databaseAutoConfig()
		if err != nil {
			clog.Warn("init", fmt.Sprintf("Auto config failed."))
		} else {
			clog.Info("init", "Database auto configurator completed building database.")
			clog.Info("init", "Starting initial database population...")
			// databasePopulate()
		}
	}
}

func main() {
	// Handle routes
	r := mux.NewRouter()

	// Subdomain 1
	s := r.Host("docs." + c.server.Domain + c.server.Port).Subrouter()
	s.PathPrefix("/").Handler(http.FileServer(http.Dir("./public/docs"))).Methods("GET")

	// List API routes firstther specific routes next
	r.HandleFunc("/api/aria1/search", ARIA1Search).Methods("POST")
	r.HandleFunc("/api/aria1/request", ARIA1Request).Methods("POST")
	// Serve other specific routes next
	r.HandleFunc("/", ServeRoot).Methods("GET")
	r.PathPrefix("/css/").Handler(http.StripPrefix("/css/", http.FileServer(http.Dir("./public/css/"))))
	r.PathPrefix("/js/").Handler(http.StripPrefix("/js/", http.FileServer(http.Dir("./public/js/"))))

	// For everything else, serve 404
	r.NotFoundHandler = http.HandlerFunc(Serve404)

	// Start server
	clog.Info("main", fmt.Sprintf("Starting webserver on port <%s>.", c.server.Port))
	clog.Fatal("main", "Server failed to start!", http.ListenAndServe(c.server.Port, r))
}
