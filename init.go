package main

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/caarlos0/env"
	"github.com/kenellorando/clog"
)

// CConfig - CServer configuration
type CConfig struct {
	LogLevel int    `env:"CSERVER_LOGLEVEL"`
	MusicDir string `env:"CSERVER_MUSIC_DIR"`
}

// DBConfig - Database configuration
type DBConfig struct {
	Host string `env:"CSERVER_DB_HOST"`
	Port string `env:"CSERVER_DB_PORT"`
	User string `env:"CSERVER_DB_USER"`
	Pass string `env:"CSERVER_DB_PASS"`
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
	err := env.Parse(&db)
	if err != nil {
		clog.Error("getDBConfig", "Failed to parse database config data.", err)
	}

	return db
}

// Initializes the logger with a log level
func initLogger(l int) {
	// Initialize logging level
	logLevel := clog.Init(l)
	clog.Debug("initLogger", "Logging service set to level <"+strconv.Itoa(logLevel)+">")
}

// Establishes database connection using configuration
func connectDatabase(dbConf DBConfig) (*sql.DB, error) {
	clog.Debug("connectDatabase", "Attempting connection to database...")

	// Form a connection with the database using config
	connectInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s", dbConf.Host, dbConf.Port, dbConf.User, dbConf.Pass, dbConf.Name)
	database, err := sql.Open("postgres", connectInfo)
	if err != nil {
		clog.Error("connectDatabase", "Connection to the database failed!", err)
	}

	return database, err
}
