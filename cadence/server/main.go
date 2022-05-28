package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/kenellorando/clog"
)

// Declare globally accessible data
var c = ServerConfig{}

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
	SourcePort       string
	StreamAddress    string
	StreamPort       string
	WhitelistPath    string
	MetadataTable    string
}

// Init function grabs configuration values for the server
// All configs are set in environment variables
// Default values for missing environment variables are set here.
// Init also initalizes other services with relevant values
func init() {
	// Webserver configuration
	c.Version = os.Getenv("CSERVER_VERSION")
	c.RootPath = os.Getenv("CSERVER_ROOTPATH")
	c.LogLevel, _ = strconv.Atoi(os.Getenv("CSERVER_LOGLEVEL"))
	c.RequestRateLimit, _ = strconv.Atoi(os.Getenv("CSERVER_REQRATELIMIT"))
	c.Port = os.Getenv("CSERVER_PORT")
	c.MusicDir = os.Getenv("CSERVER_MUSIC_DIR")
	c.SourceAddress = os.Getenv("CSERVER_SOURCEADDRESS")
	c.SourcePort = os.Getenv("CSERVER_SOURCEPORT")
	c.StreamAddress = os.Getenv("CSERVER_STREAMADDRESS")
	c.StreamPort = os.Getenv("CSERVER_STREAMPORT")
	c.WhitelistPath = os.Getenv("CSERVER_WHITELIST_PATH")
	c.MetadataTable = os.Getenv("CSERVER_DB_METADATA_TABLE")

	// Initialize logging
	clog.Level(c.LogLevel)
	clog.Info("init", fmt.Sprintf("Logging service initialized to level <%v>", c.LogLevel))

	newDatabase, err := dbAutoConfig()
	if err != nil {
		clog.Warn("init", "Database setup failed! Future database requests will also fail. Data check will be skipped.")
	} else {
		database = newDatabase
		err = dbPopulate()
		if err != nil {
			clog.Warn("init", "Initial database population failed.")
		} else {
			clog.Debug("init", "Database population OK.")
		}
	}
}

func main() {
	clog.Info("main", fmt.Sprintf("Starting webserver on port <%s>.", c.Port))
	clog.Fatal("main", "Server failed to start!", http.ListenAndServe(c.Port, routes()))
}
