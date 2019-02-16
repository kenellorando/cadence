package main

import (
	"strconv"

	"github.com/caarlos0/env"
	"github.com/kenellorando/clog"
)

// CConfig - CServer configuration
type CConfig struct {
	LogLevel int `env:"CSERVER_LOGLEVEL"`
}

// DBConfig - Database configuration
type DBConfig struct {
	User string `env:"CSERVER_DB_USER"`
	Pass string `env:"CSERVER_DB_PASS"`
	Host string `env:"CSERVER_DB_HOST"`
	Port string `env:"CSERVER_DB_PORT"`
	Name string `env:"CSERVER_DB_NAME"`
}

// Parses environment variables for configuration
func getConfig() (CConfig, DBConfig) {
	// Read general configuration data
	ws := CConfig{}
	env.Parse(&ws)

	// Read database configuration data
	db := DBConfig{}
	env.Parse(&db)

	return ws, db
}

// Initializes the logger with a log level
func initLogger(l int) {
	// Initialize logging level
	logLevel := clog.Init(l)
	clog.Debug("initLogger", "Logging service initialized to level "+strconv.Itoa(logLevel)+".")
}

func initDatabase(db DBConfig) {
	// ...
}
