package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/kenellorando/clog"
)

var ctx = context.Background()

var c = ServerConfig{}

// todo: rename source, stream, database to component names
type ServerConfig struct {
	Version          string
	RootPath         string
	RequestRateLimit int
	LogLevel         int
	Port             string
	MusicDir         string
	SourceAddress    string
	SourcePort       string
	StreamAddress    string
	StreamPort       string
	PostgresAddress  string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
	PostgresDBName   string
	PostgresSSL      string
	RedisAddress     string
	RedisPort        string
	WhitelistPath    string
	MetadataTable    string
	DevMode          bool
}

func main() {
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
	c.PostgresAddress = os.Getenv("CSERVER_POSTGRESADDRESS")
	c.PostgresPort = os.Getenv("CSERVER_POSTGRESPORT")
	c.PostgresUser = os.Getenv("CSERVER_POSTGRESUSER")
	c.PostgresPassword = os.Getenv("CSERVER_POSTGRESPASSWORD")
	c.PostgresDBName = os.Getenv("CSERVER_POSTGRESDBNAME")
	c.PostgresSSL = os.Getenv("CSERVER_POSTGRESSSL")
	c.RedisAddress = os.Getenv("CSERVER_REDISADDRESS")
	c.RedisPort = os.Getenv("CSERVER_REDISPORT")
	c.WhitelistPath = os.Getenv("CSERVER_WHITELIST_PATH")
	c.MetadataTable = os.Getenv("CSERVER_DB_METADATA_TABLE")
	c.DevMode, _ = strconv.ParseBool(os.Getenv("CSERVER_DEVMODE"))

	clog.Level(c.LogLevel)
	clog.Debug("init", fmt.Sprintf("Cadence Logger initialized to level <%v>.", c.LogLevel))

	redisInit()
	postgresInit()
	go filesystemMonitor()
	go icecastMonitor()

	clog.Info("main", fmt.Sprintf("Starting Cadence on port <%s>.", c.Port))
	clog.Fatal("main", "Cadence failed to start!", http.ListenAndServe(c.Port, routes()))
}
