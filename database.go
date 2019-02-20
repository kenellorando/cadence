package main

import (
	"database/sql"
	"log"

	"github.com/kenellorando/clog"
	_ "github.com/lib/pq"
)

// Check if the tables specified in the
func databaseTableCheck() {
	clog.Debug("databaseTableCheck", "Running table check in database...")
	//database, _ := databaseConnect()
}

// Establishes database connection using configuration,
// Confirms connection with a ping, returns a database session
// Appropriately handles connection-errors here
func databaseConnect() (*sql.DB, error) {
	clog.Debug("databaseConnect", "Trying connection to database...")

	// Form a connection with the database using config
	database, err := sql.Open(c.db.Driver, c.db.DSN)
	if err != nil {
		clog.Error("databaseConnect", "Connection to the database failed!", err)
		return nil, err
	}

	// According to the go wiki, connections are deferred until queries are made
	// We ping the database here to establish the connection
	clog.Info("databaseConnect", "Connected to the database. Pinging to confirm open connection.")
	err = database.Ping()
	if err != nil {
		log.Fatal(err)
	}

	return database, err
}
