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
	dropDatabase := fmt.Sprintf("DROP DATABASE IF EXISTS %s", c.db.Name)
	createDatabase := fmt.Sprintf("CREATE DATABASE %s", c.db.Name)
	createTable := fmt.Sprintf(`CREATE TABLE %s
	(
	   id serial PRIMARY KEY,
	   title character varying(255),
	   album character varying(255),
	   artist character varying(255),
	   genre character varying(255),
	   path character varying(255)
	)
	WITH (
	   OIDS = FALSE
	)`, c.schema.Table)

	// Drop the database if it exists
	clog.Debug("databaseAutoConfig", fmt.Sprintf("Deleting existing databases named <%s>.", c.db.Name))
	_, err := database.Exec(dropDatabase)
	if err != nil {
		clog.Error("databaseAutoConfig", "Failed to remove existing database. Skipping remaining autoconfig steps.", err)
		return err
	}

	// Create the database
	clog.Debug("databaseAutoConfig", fmt.Sprintf("Creating database <%s>.", c.db.Name))
	_, err = database.Exec(createDatabase)
	if err != nil {
		clog.Error("databaseAutoConfig", "Failed to create database. Skipping remaining autoconfig steps.", err)
		return err
	}

	// Postgres has no 'USE' statements
	// In order to connect to the newly created database
	// we redefine the DSN to hold the database name
	// and reconnect using it.
	clog.Debug("databaseAutoConfig", fmt.Sprintf("Database <%s> recreated. Reconnecting to newly created database...", c.db.Name))
	c.db.DSN = fmt.Sprintf(c.db.DSN+" dbname='%s'", c.db.Name)
	database, _ = databaseConnect()
	if err != nil {
		clog.Error("databaseAutoConfig", "Failed to connect to newly created database. Skipping remaining autoconfig steps.", err)
		return err
	}

	// Build the database tables
	clog.Debug("databaseAutoConfig", fmt.Sprintf("Building database schema for table <%s>...", c.schema.Table))
	_, err = database.Exec(createTable)
	if err != nil {
		clog.Error("databaseAutoConfig", "Failed to build database table. Skipping remaining autoconfig steps.", err)
		return err
	}

	clog.Debug("databaseAutoConfig", fmt.Sprintf("Table <%s> built successfully.", c.schema.Table))

	// TODO: Populate table
	return err
}

/*
func databasePopulate() error {
	// Check if MUSIC_DIR exists. Return if err
	if _, err := os.Stat(c.server.MusicDir); err != nil {
		if os.IsNotExist(err) {
			clog.Error("databasePopulate", "The defined music directory")
			return
		}
	}

	return err
}
*/

// Establishes connection to database using configuration,
// Confirms connection with a ping, returns a database session
// Appropriately handles connection-errors here
func databaseConnect() (*sql.DB, error) {
	clog.Debug("databaseConnect", fmt.Sprintf("Connecting to database cluster <%s:%s> with set credentials...", c.db.Host, c.db.Port))

	// Initialize connection pool
	// Note that sql.Open does not actually "connect";
	// According to the go wiki, connections are deferred until queries are made
	// We ping the database here to confirm the connection
	database, err := sql.Open(c.db.Driver, c.db.DSN)
	err = database.Ping()
	if err != nil {
		clog.Error("databaseConnect", fmt.Sprintf("Failed to confirm open connection to cluster <%s:%s>", c.db.Host, c.db.Port), err)
		return nil, err
	}

	clog.Debug("databaseConnect", fmt.Sprintf("Connected successfully to database cluster <%s:%s>", c.db.Host, c.db.Port))
	return database, nil
}
