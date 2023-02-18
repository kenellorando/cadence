package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/go-redis/redis"
	"github.com/kenellorando/clog"
)

var c = ServerConfig{}

var db *sql.DB

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
	DatabaseAddress  string
	DatabasePort     string
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
	c.DatabaseAddress = os.Getenv("CSERVER_DBADDRESS")
	c.DatabasePort = os.Getenv("CSERVER_DBPORT")
	c.WhitelistPath = os.Getenv("CSERVER_WHITELIST_PATH")
	c.MetadataTable = os.Getenv("CSERVER_DB_METADATA_TABLE")
	c.DevMode, _ = strconv.ParseBool(os.Getenv("CSERVER_DEVMODE"))

	clog.Level(c.LogLevel)
	clog.Debug("init", fmt.Sprintf("Cadence Logger initialized to level <%v>.", c.LogLevel))

	client := redis.NewClient(&redis.Options{
		Addr:     c.DatabaseAddress,
		Password: "",
		DB:       0,
	})
	pong, err := client.Ping().Result()
	fmt.Println(pong, err)

	// dbRefresh()
	go filesystemMonitor()
	go icecastMonitor()

	clog.Info("main", fmt.Sprintf("Starting Cadence on port <%s>.", c.Port))
	clog.Fatal("main", "Cadence failed to start!", http.ListenAndServe(c.Port, routes()))
}
