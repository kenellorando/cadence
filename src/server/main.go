package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	// A client that has connected but not sent a complete request header is
	// holding a connection open for nothing.
	readHeaderTimeout = 10 * time.Second
	readTimeout       = 30 * time.Second
	idleTimeout       = 120 * time.Second
	// How long in-flight requests are given to finish once a stop is requested.
	shutdownTimeout = 15 * time.Second
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
	SourcePort        string
	SourcePassword    string
	InternalStream    bool
	DevMode           bool
	LogLevel          string
}

func parseLogLevel(level string) slog.Level {
	if level == "" {
		return slog.LevelInfo
	}

	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		slog.Warn(fmt.Sprintf("Unrecognized log level %s!", level), "func", "parseLogLevel")
		return slog.LevelInfo
	}
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
	c.SourcePort = os.Getenv("CSERVER_SOURCEPORT")
	c.SourcePassword = os.Getenv("CSERVER_SOURCEPASSWORD")
	c.InternalStream, _ = strconv.ParseBool(os.Getenv("CSERVER_INTERNALSTREAM"))
	c.DevMode, _ = strconv.ParseBool(os.Getenv("CSERVER_DEVMODE"))
	c.LogLevel = os.Getenv("CSERVER_LOGLEVEL")

	slog.SetLogLoggerLevel(parseLogLevel(c.LogLevel))

	if postgresInit() == nil {
		if rateLimitInit() != nil {
			slog.Warn("Rate limiting is unavailable.", "func", "main")
		}
		if postgresPopulate() != nil {
			slog.Warn("Initial database population failed.", "func", "main")
		}
	}
	go filesystemMonitor()
	// Only one of these may drive the radio state. With the built-in source the
	// audio arrives here directly and metadata comes with it, so polling a
	// separate streaming server would only overwrite what we already know.
	if c.InternalStream {
		startAudioSource()
	} else {
		go icecastMonitor()
	}

	server := &http.Server{
		Addr:    c.Port,
		Handler: routes(),
		// WriteTimeout is deliberately left unset. /api/radiodata/sse holds its
		// response open for the lifetime of the client, and any write deadline
		// would sever every event stream on expiry. Header reads, body reads and
		// idle keep-alive connections are still bounded.
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		IdleTimeout:       idleTimeout,
	}

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.ListenAndServe()
	}()
	slog.Info(fmt.Sprintf("Starting Cadence on port <%s>.", c.Port), "func", "main")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		slog.Error("Cadence failed to start!", "func", "main", "error", err)
	case sig := <-stop:
		slog.Info(fmt.Sprintf("Received signal <%s>, shutting down.", sig), "func", "main")
		// Event stream consumers never close their side, so they are dropped
		// first. Otherwise Shutdown has nothing to wait for but its own timeout.
		radiodata_sse.Close()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			slog.Error("Shutdown did not complete cleanly.", "func", "main", "error", err)
		}
		slog.Info("Cadence stopped.", "func", "main")
	}
}
