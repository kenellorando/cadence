package main

import (
	"database/sql"
	"fmt"

	env "github.com/deanishe/go-env"
	"github.com/kenellorando/clog"
)

// Config - Primary configuration object holder
type Config struct {
	server CConfig
	db     DBConfig
}

// CConfig - CServer configuration
type CConfig struct {
	MusicDir string
}

// DBConfig - Database configuration
type DBConfig struct {
	Host string
	Port string
	User string
	Pass string
	Name string
}

func getConfigs() Config {
	clog.Debug("initConfigs", "Reading environment variables for configuration...")

	// Full config object
	config := Config{}

	// Get server-related configs
	server := CConfig{}
	server.MusicDir = env.GetString("CSERVER_MUSIC_DIR")
	config.server = server

	// Get database related configs
	db := DBConfig{}
	db.Host = env.GetString("CSERVER_DB_HOST")
	db.Port = env.GetString("CSERVER_DB_PORT")
	db.User = env.GetString("CSERVER_DB_USER")
	db.Pass = env.GetString("CSERVER_DB_PASS")
	db.Name = env.GetString("CSERVER_DB_NAME")
	config.db = db

	return config
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
