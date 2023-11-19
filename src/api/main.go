package main

import (
	"fmt"
	"time"
	"log/slog"
	"net/http"
	"os"
	"strconv"
)

var c = ServerConfig{}

type ServerConfig struct {
	Version           string
	RootPath          string
	RequestRateLimit  int
	Port              string
	MusicDir          string
	LiquidsoapAddress string
	LiquidsoapPort    string
	IcecastAddress    string
	IcecastPort       string
	PostgresAddress   string
	PostgresPort      string
	PostgresUser      string
	PostgresPassword  string
	PostgresDBName    string
	PostgresTableName string
	PostgresSSL       string
	RedisAddress      string
	RedisPort         string
	WhitelistPath     string
	DevMode           bool
}

func main() {
	c.Version = os.Getenv("CSERVER_VERSION")
	c.RootPath = os.Getenv("CSERVER_ROOTPATH")
	c.RequestRateLimit, _ = strconv.Atoi(os.Getenv("CSERVER_REQRATELIMIT"))
	c.Port = os.Getenv("CSERVER_PORT")
	c.MusicDir = os.Getenv("CSERVER_MUSIC_DIR")
	c.LiquidsoapAddress = os.Getenv("CSERVER_LIQUIDSOAPADDRESS")
	c.LiquidsoapPort = os.Getenv("CSERVER_LIQUIDSOAPPORT")
	c.IcecastAddress = os.Getenv("CSERVER_ICECASTADDRESS")
	c.IcecastPort = os.Getenv("CSERVER_ICECASTPORT")
	c.PostgresAddress = os.Getenv("CSERVER_POSTGRESADDRESS")
	c.PostgresPort = os.Getenv("CSERVER_POSTGRESPORT")
	c.PostgresUser = os.Getenv("CSERVER_POSTGRESUSER")
	c.PostgresPassword = os.Getenv("POSTGRES_PASSWORD")
	c.PostgresDBName = os.Getenv("CSERVER_POSTGRESDBNAME")
	c.PostgresTableName = os.Getenv("CSERVER_POSTGRESTABLENAME")
	c.PostgresSSL = os.Getenv("CSERVER_POSTGRESSSL")
	c.RedisAddress = os.Getenv("CSERVER_REDISADDRESS")
	c.RedisPort = os.Getenv("CSERVER_REDISPORT")
	c.WhitelistPath = os.Getenv("CSERVER_WHITELIST_PATH")
	c.DevMode, _ = strconv.ParseBool(os.Getenv("CSERVER_DEVMODE"))

	if postgresInit() == nil {
		if postgresPopulate() != nil {
			slog.Warn("Initial database population failed.", "func", "main")
		}
	}
	go redisInit()
	go filesystemMonitor()
	go func() {
		for {
			time.Sleep(1 * time.Second)
			icecastMonitor()
		}
	}()

	slog.Info(fmt.Sprintf("Starting Cadence on port <%s>.", c.Port), "func", "main")
	if http.ListenAndServe(c.Port, routes()) != nil {
		slog.Error("Cadence failed to start!", "func", "main")
	}
}
