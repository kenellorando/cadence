// database.go
// Database initialization and configuration

package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dhowden/tag"
	"github.com/kenellorando/clog"
	_ "github.com/lib/pq"
)

var database *sql.DB // Database abstraction interface

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
	   year character varying(4),
	   path character varying(510)
	)
	WITH (
	   OIDS = FALSE
	)`, c.schema.Table)
	enableExtension := "CREATE EXTENSION fuzzystrmatch"

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
	database, err = databaseConnect()
	if err != nil {
		clog.Error("databaseAutoConfig", "Failed to connect to newly created database. Skipping remaining autoconfig steps.", err)
		return err
	}

	// Enable fuzzystrmatch for levenshtein sorting
	// (sorting search results by how close they are to the query)
	clog.Debug("databaseAutoConfig", "Enabling fuzzystrmatch extension...")
	_, err = database.Exec(enableExtension)
	if err != nil {
		clog.Error("databaseAutoConfig", "Failed to enable fuzzystrmatch!", err)
		return err
	}

	// Build the database tables
	clog.Debug("databaseAutoConfig", fmt.Sprintf("Reconnected. Building database schema for table <%s>...", c.schema.Table))
	_, err = database.Exec(createTable)
	if err != nil {
		clog.Error("databaseAutoConfig", "Failed to build database table!", err)
		return err
	}

	return err
}

// Scans the env-var set music directory for audio files,
// parses their metadata and inserts them into the table.
func databasePopulate() error {
	// SQL exec statements here
	insertInto := fmt.Sprintf("INSERT INTO %s (%s, %s, %s, %s, %s, %s) SELECT $1, $2, $3, $4, $5, $6::VARCHAR WHERE NOT EXISTS (SELECT %s FROM %s WHERE %s=$6)", c.schema.Table, "title", "album", "artist", "genre", "year", "path", "path", c.schema.Table, "path")

	// Check if music directory exists. Return if err
	_, err := os.Stat(c.server.MusicDir)
	if err != nil {
		if os.IsNotExist(err) {
			clog.Error("databasePopulate", "The defined music directory was not found.", err)
			return err
		}
	}

	clog.Debug("databasePopulate", "Extracting metadata from given music directory...")

	// Recursive walk on directory
	err = filepath.Walk(c.server.MusicDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Skip non-music files
		var extensions = [...]string{".flac", ".ogg", ".mp3"}
		for _, ext := range extensions {
			if strings.HasSuffix(path, ext) {
				// Open a file for reading
				file, e := os.Open(path)
				if e != nil {
					return e
				}

				// Read metadata from the file
				tags, er := tag.ReadFrom(file)
				if er != nil {
					return er
				}

				// Insert into database
				_, err = database.Exec(insertInto, tags.Title(), tags.Album(), tags.Artist(),
					tags.Genre(), tags.Year(), path)
				if err != nil {
					panic(err)
				}

				// Close the file
				file.Close()
			} else {
				continue
			}
		}
		return nil
	})

	if err != nil {
		clog.Error("databasePopulate", "Examination of music file metadata failed!", err)
		return err
	}

	clog.Debug("databasePopulate", "Database population complete.")
	return nil
}

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
	if err != nil {
		clog.Error("databaseConnect", fmt.Sprintf("Failed to start connection to cluster <%s:%s>", c.db.Host, c.db.Port), err)
		return nil, err
	}
	err = database.Ping()
	if err != nil {
		clog.Error("databaseConnect", fmt.Sprintf("Failed to confirm ping connection to cluster <%s:%s>", c.db.Host, c.db.Port), err)
		return nil, err
	}

	clog.Debug("databaseConnect", fmt.Sprintf("Connected successfully to database cluster <%s:%s>", c.db.Host, c.db.Port))
	return database, nil
}
