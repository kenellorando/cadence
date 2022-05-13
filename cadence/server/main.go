package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/kenellorando/clog"
)

// Declare globally accessible data
var c = Config{}

// Config - Primary configuration object holder
type Config struct {
	server ServerConfig
	db     DBConfig
	schema SchemaConfig
}

// ServerConfig - Webserver configuration
type ServerConfig struct {
	Version          string
	RootPath         string
	Domain           string
	RequestRateLimit int
	LogLevel         int
	Port             string
	MusicDir         string
	SourceAddress    string
	WhitelistPath    string
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
	// TODO: Make this better (do service connection retries at intervals?)
	// this pauses initialization so the postgres service may start first-- version 4C
	time.Sleep(2 * time.Second)

	// Declare substructs of the global config
	server := ServerConfig{}
	db := DBConfig{}
	schema := SchemaConfig{}

	// Webserver configuration
	server.Version = os.Getenv("CSERVER_VERSION")
	server.RootPath = os.Getenv("CSERVER_ROOTPATH")
	server.LogLevel, _ = strconv.Atoi(os.Getenv("CSERVER_LOGLEVEL"))
	server.RequestRateLimit, _ = strconv.Atoi(os.Getenv("CSERVER_REQRATELIMIT"))
	server.Port = os.Getenv("CSERVER_PORT")
	server.MusicDir = os.Getenv("CSERVER_MUSIC_DIR")
	server.SourceAddress = os.Getenv("CSERVER_SOURCEADDRESS")
	server.WhitelistPath = os.Getenv("CSERVER_WHITELIST_PATH")
	// Database server configuration
	db.Host = os.Getenv("CSERVER_DB_HOST")
	db.Port = os.Getenv("CSERVER_DB_PORT")
	db.User = os.Getenv("CSERVER_DB_USER")
	db.Pass = os.Getenv("CSERVER_DB_PASS")
	db.Name = os.Getenv("CSERVER_DB_NAME")
	db.SSLMode = os.Getenv("CSERVER_DB_SSLMODE")
	db.Driver = os.Getenv("CSERVER_DB_DRIVER")
	db.DSN = fmt.Sprintf("host='%s' port='%s' user='%s' password='%s' sslmode='%s'", db.Host, db.Port, db.User, db.Pass, db.SSLMode)
	// Database schema configuration
	schema.Table = os.Getenv("CSERVER_DB_TABLE")

	// Set the substructs of the global config
	c.server = server
	c.db = db
	c.schema = schema

	// Initialize logging
	clog.Level(c.server.LogLevel)
	clog.Info("init", fmt.Sprintf("Logging service initialized to level <%v>", c.server.LogLevel))

	newDatabase, err := dbAutoConfig()
	if err != nil {
		clog.Warn("init", "Database setup failed! Future database requests will also fail.")
		clog.Debug("init", "Skipping data check.")
	} else {
		database = newDatabase
		err = dbPopulate()
		if err != nil {
			clog.Warn("init", "Initial database population failed.")
		} else {
			clog.Debug("init", "Database population OK.")
		}
	}

	// clog.Debug("init", fmt.Sprintf("Establishing a connection to database server <%s:%s>", c.db.Host, c.db.Port))
	// newDatabase, err := databaseConnect()
	// if err != nil {
	// 	clog.Warn("init", "Database server connection test failed! Future database requests will also fail.")
	// 	clog.Debug("init", "Skipping data check.")
	// } else {
	// 	// Set the global database object to the newly made pointer
	// 	// Start the database data check
	// 	clog.Info("init", "Database server connection test successful. Starting database auto configuration..")
	// 	database = newDatabase
	// 	err := databaseAutoConfig()
	// 	if err != nil {
	// 		clog.Warn("init", "Database auto config failed. Skipping initial database population.")
	// 	} else {
	// 		clog.Info("init", "Database auto config completed building database. Starting initial database population...")
	// 		err = databasePopulate()
	// 		if err != nil {
	// 			clog.Warn("init", "Initial database population failed.")
	// 		} else {
	// 			clog.Info("init", "All initialization tasks completed successfully.")
	// 		}
	// 	}
	// }
}

func main() {
	clog.Info("main", fmt.Sprintf("Starting webserver on port <%s>.", c.server.Port))
	clog.Fatal("main", "Server failed to start!", http.ListenAndServe(c.server.Port, routes()))
}
