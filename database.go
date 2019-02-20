package main

import (
	"database/sql"
	"fmt"

	"github.com/kenellorando/clog"
	_ "github.com/lib/pq"
)

// Check if the tables specified in the config exist
// This is only run once by init to confirm the data table
func databaseTableCheck() {
	clog.Debug("databaseTableCheck", "Running table check in database...")
	//database, _ := databaseConnect()
	//database, err := databaseConnect()
}

func databaseCreate(database *sql.DB) error {
	clog.Info("databaseCreate", fmt.Sprintf("Create database <%s>.", c.db.Name))

	createDatabase := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", c.db.Name)

	fmt.Printf("%v", createDatabase)
	_, err := database.Exec(createDatabase)
	if err != nil {
		clog.Error("databaseCreate", "Failed to create database. Skipping further creation steps.", err)
		return err
	}

	/*
		_, err = db.Exec("USE " + name)
		if err != nil {
			panic(err)
		}

		_, err = db.Exec("CREATE TABLE example ( id integer, data varchar(32) )")
		if err != nil {
			panic(err)
		}
		return err
	*/
	return err
}

// Establishes database connection using configuration,
// Confirms connection with a ping, returns a database session
// Appropriately handles connection-errors here
func databaseConnect() (*sql.DB, error) {
	clog.Info("databaseConnect", fmt.Sprintf("Trying connection to database server as user <%s>.", c.db.User))

	// Form a connection with the database using config
	database, err := sql.Open(c.db.Driver, fmt.Sprintf("host=%s port=%s user=%s password=%s sslmode=%s", c.db.Host, c.db.Port, c.db.User, c.db.Pass, c.db.SSLMode))
	defer database.Close()

	if err != nil {
		clog.Error("databaseConnect", "Connection to the database server failed!", err)
		return nil, err
	}

	// According to the go wiki, connections are deferred until queries are made
	// We ping the database here to establish the connection
	clog.Info("databaseConnect", fmt.Sprintf("Connected to database server. Pinging to open connection to <%s>...", c.db.Name))
	err = database.Ping()
	if err != nil {
		clog.Error("databaseConnect", "Ping test failed to confirm open connection.", err)
		clog.Info("databaseConnect", "Will attempt to create the database and tables defined in configuration.")
		databaseCreate(database)
	} else {
		clog.Info("databaseConnect", fmt.Sprintf("Connected successfully to database <%s>", c.db.Name))
	}

	return database, err
}
