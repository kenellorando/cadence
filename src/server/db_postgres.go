// db_postgres.go
// Metadata database configuration and population.

package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dhowden/tag"
	"github.com/lib/pq"
)

var dbp *sql.DB

const (
	// Postgres always ships this database, and it is where a connection is made
	// to create the one Cadence actually uses.
	maintenanceDatabase = "postgres"
	// Used only if CSERVER_POSTGRESDBNAME is unset.
	defaultDatabase = "cadence"
)

// Builds a connection string for one database on the configured server.
func postgresDSN(database string) string {
	return fmt.Sprintf("host='%s' port='%s' user='%s' password='%s' dbname='%s' sslmode='%s'",
		c.PostgresAddress, c.PostgresPort, c.PostgresUser, c.PostgresPassword, database, c.PostgresSSL)
}

// Creates the configured metadata database if it does not already exist.
// A database cannot be created from a connection already inside it, so this
// connects to the default maintenance database to do the work.
func postgresCreateDatabase() error {
	maintenance, err := sql.Open("postgres", postgresDSN(maintenanceDatabase))
	if err != nil {
		slog.Error("Couldn't open a connection to the maintenance database.", "func", "postgresCreateDatabase", "error", err)
		return err
	}
	defer maintenance.Close()
	if err = maintenance.Ping(); err != nil {
		slog.Error("Couldn't ping the maintenance database.", "func", "postgresCreateDatabase", "error", err)
		return err
	}
	_, err = maintenance.Exec(fmt.Sprintf("CREATE DATABASE %s", pq.QuoteIdentifier(c.PostgresDBName)))
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "42P04" {
			// 42P04 is duplicate_database, which is the normal case on any restart.
			slog.Debug(fmt.Sprintf("Database <%s> already exists.", c.PostgresDBName), "func", "postgresCreateDatabase")
			return nil
		}
		slog.Error(fmt.Sprintf("Couldn't create database <%s>.", c.PostgresDBName), "func", "postgresCreateDatabase", "error", err)
		return err
	}
	slog.Info(fmt.Sprintf("Created database <%s>.", c.PostgresDBName), "func", "postgresCreateDatabase")
	return nil
}

func postgresInit() (err error) {
	// We wait a bit to give some leeway for Postgres to finish startup.
	// Obligatory: There's probably a better way to do this.
	time.Sleep(5 * time.Second)
	if c.PostgresDBName == "" {
		slog.Warn(fmt.Sprintf("No database name configured, defaulting to <%s>.", defaultDatabase), "func", "postgresInit")
		c.PostgresDBName = defaultDatabase
	}
	if err = postgresCreateDatabase(); err != nil {
		return err
	}
	dbp, err = sql.Open("postgres", postgresDSN(c.PostgresDBName))
	if err != nil {
		slog.Error("Couldn't open a connection to database.", "func", "postgresInit", "error", err)
		return err
	}
	err = dbp.Ping()
	if err != nil {
		slog.Error("Couldn't ping the metadata database.", "func", "postgresInit", "error", err)
		return err
	}
	// Enable fuzzystrmatch for levenshtein sorting.
	// This enables the database to return results based on search similarity.
	slog.Debug("Enabling fuzzystrmatch extension...", "func", "postgresInit")
	enableExtension := "CREATE EXTENSION fuzzystrmatch"
	_, err = dbp.Exec(enableExtension)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "42710" {
			// 42710 also indicates an existing Postgres instance configured by another Cadence instance is still running.
			slog.Debug("fuzzystrmatch already enabled on metadata database.", "func", "postgresInit")
		} else {
			slog.Error("Failed to enable fuzzystrmatch. Search will function in a degraded state.", "func", "postgresInit", "error", err)
			return err
		}
	}
	return nil
}

func postgresPopulate() error {
	dropTable := fmt.Sprintf("DROP TABLE IF EXISTS %s", c.PostgresTableName)
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
	)`, c.PostgresTableName)

	// Drop the metadata table and rebuild it to start fresh.
	slog.Debug(fmt.Sprintf("Dropping table <%s>...", c.PostgresTableName), "func", "postgresPopulate")
	_, err := dbp.Exec(dropTable)
	if err != nil {
		slog.Error("Failed to drop table. Skipping remaining autoconfig steps.", "func", "postgresPopulate", "error", err)
		return err
	}
	slog.Debug(fmt.Sprintf("Creating table <%s>...", c.PostgresTableName), "func", "postgresPopulate")
	_, err = dbp.Exec(createTable)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "42P07" {
			// 42P10 indicates an existing metadata table configured by another Cadence instance is still running.
			slog.Info("Metadata database already exists", "func", "postgresPopulate")
		} else {
			slog.Error("Failed to build database table!", "func", "postgresPopulate", "error", err)
			return err
		}
	}
	slog.Debug("Verifying music metadata directory is accessible.")
	_, err = os.Stat(c.MusicDir)
	if err != nil {
		slog.Error(fmt.Sprintf("Could not open music directory <%s> for verification.", c.MusicDir), "func", "postgresPopulate", "error", err)
		if os.IsNotExist(err) {
			slog.Error("The configured target music directory was not found.", "func", "postgresPopulate", "error", err)
			return err
		}
	}

	insertInto := fmt.Sprintf("INSERT INTO %s (%s, %s, %s, %s, %s, %s) SELECT $1, $2, $3, $4, $5, $6", c.PostgresTableName, "title", "album", "artist", "genre", "year", "path")
	slog.Debug(fmt.Sprintf("Extracting metadata from audio files in: <%s>", c.MusicDir), "func", "postgresPopulate")
	// A file we can't read is a reason to skip that file, not to abandon the
	// rest of the library. Anything skipped here is counted and reported once
	// the walk finishes.
	skipped := 0
	err = filepath.Walk(c.MusicDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			slog.Error(fmt.Sprintf("Could not access <%s> during walk, skipping.", path), "func", "postgresPopulate", "error", err)
			skipped++
			return nil
		}
		slog.Debug(fmt.Sprintf("Populate analyzing file: <%s>", path), "func", "postgresPopulate")
		if info.IsDir() {
			slog.Debug(fmt.Sprintf("<%s> is a directory, skipping.", path), "func", "postgresPopulate")
			return nil
		}
		extensions := []string{".mp3", ".flac", ".ogg"}
		for _, ext := range extensions {
			if strings.HasSuffix(path, ext) {
				file, err := os.Open(path)
				if err != nil {
					slog.Error(fmt.Sprintf("Problem opening <%s> for music population, skipping.", path), "func", "postgresPopulate", "error", err)
					skipped++
					return nil
				}
				defer file.Close()
				tags, err := tag.ReadFrom(file)
				if err != nil {
					slog.Error(fmt.Sprintf("Problem fetching tags from <%s>, skipping.", path), "func", "postgresPopulate", "error", err)
					skipped++
					return nil
				}
				_, err = dbp.Exec(insertInto, tags.Title(), tags.Album(), tags.Artist(), tags.Genre(), tags.Year(), path)
				if err != nil {
					slog.Error(fmt.Sprintf("Problem populating metadata for <%s>.", path), "func", "postgresPopulate", "error", err)
					return err
				}
				slog.Debug(fmt.Sprintf("Finished populating track: %s by %s", tags.Title(), tags.Artist()), "func", "postgresPopulate")
				break
			}
		}
		return nil
	})
	if err != nil {
		slog.Error("Music metadata database population failed, or may be incomplete.", "func", "postgresPopulate", "error", err)
		return err
	}
	if skipped > 0 {
		slog.Warn(fmt.Sprintf("Database population completed, but %d file(s) were skipped.", skipped), "func", "postgresPopulate")
		return nil
	}
	slog.Info("Database population completed.", "func", "postgresPopulate")
	return nil
}
