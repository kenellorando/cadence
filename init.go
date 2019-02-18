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

// Parses environment variables for cserver configuration
func getCConfig() CConfig {
	// Read general configuration data
	c := CConfig{}
	env.Parse(&c)

	return c
}

// Parses environment variables for database configuration
func getDBConfig() DBConfig {
	// Read database configuration data
	db := DBConfig{}
	env.Parse(&db)

	return db
}

// Initializes the logger with a log level
func initLogger(l int) {
	// Initialize logging level
	logLevel := clog.Init(l)
	clog.Debug("initLogger", "Logging service initialized to level <"+strconv.Itoa(logLevel)+">")
}

// Establishes database connection using configuration
func initDatabase(db DBConfig) {
	clog.Debug("initDatabase", "Attempting connection to database...")
	// ...
	return
}

func connectDatabase(db DBConfig) {

}
