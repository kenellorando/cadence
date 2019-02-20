package main

import (
	"database/sql"
	"fmt"

	"github.com/kenellorando/clog"
)

// Establishes database connection using configuration
func connectDatabase(db DBConfig) (*sql.DB, error) {
	clog.Debug("connectDatabase", "Trying connection to database...")

	// Form a connection with the database using config
	connectInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s", db.Host, db.Port, db.User, db.Pass, db.Name)
	database, err := sql.Open("postgres", connectInfo)
	if err != nil {
		clog.Error("connectDatabase", "Connection to the database failed!", err)
	} else {
		clog.Info("connectDatabase", "Connected to the database.")
	}

	return database, err
}
