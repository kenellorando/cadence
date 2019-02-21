package main

import (
	"database/sql"
	"fmt"

	"github.com/kenellorando/clog"
	_ "github.com/lib/pq"
)

// Creates a database and tables using configs in c.schema
// This is only called after a successful connection to the database server
// in the init function.
func databaseAutoConfig() error {
	clog.Debug("databaseAutoConfig", "Starting automatic database configuration...")

	// SQL exec statements here
	createDatabase := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", c.schema.Name)
	useDatabase := fmt.Sprintf("USE %s", c.schema.Name)

	// Create the database if it does not exist
	clog.Info("databaseAutoConfig", fmt.Sprintf("Creating database <%s> if it does not exist.", c.schema.Name))
	_, err := database.Exec(createDatabase)
	if err != nil {
		clog.Error("databaseAutoConfig", "Failed to create database. Skipping further creation steps.", err)
		return err
	}

	// Set the database in use to the one specified
	_, err = database.Exec(useDatabase)
	if err != nil {
		clog.Error("databaseAutoConfig", fmt.Sprintf("Could not switch to database <%s>", c.schema.Name), err)
	}

	//Todo: Create table using schema, then call populator?

	/*
		_, err = db.Exec("USE " + name)ß
		if err != nil {
			panic(err)
		}

		_, err = db.Exec("CREATE TABLE example ( id integer, data varchar(32) )")
		if err != nil {
			panic(err)
		}
	*/
	return err
}

// Establishes connection to database using configuration,
// Confirms connection with a ping, returns a database session
// Appropriately handles connection-errors here
func databaseConnect() (*sql.DB, error) {
	clog.Info("databaseConnect", fmt.Sprintf("Connecting to database server <%s:%s> with set credentials...", c.db.Host, c.db.Port))

	// Initialize connection pool
	// Note that sql.Open does not actually "connect";
	// According to the go wiki, connections are deferred until queries are made
	// We ping the database here to confirm the connection
	database, err := sql.Open(c.db.Driver, c.db.DSN)
	err = database.Ping()
	if err != nil {
		clog.Error("databaseConnect", fmt.Sprintf("Failed to confirm open connection to <%s:%s>", c.db.Host, c.db.Port), err)
		return nil, err
	}

	clog.Info("databaseConnect", fmt.Sprintf("Connected successfully to database <%s:%s>", c.db.Host, c.db.Port))
	return database, nil
}
