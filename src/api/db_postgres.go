// db_postgres.go
// Metadata database configuration and population.

package main

import (
	"database/sql"
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

func postgresInit() (err error) {
	// We wait a bit to give some leeway for Postgres to finish startup.
	// Obligatory: There's probably a better way to do this.
	time.Sleep(5 * time.Second)
	dsn := fmt.Sprintf("host='%s' port='%s' user='%s' password='%s' sslmode='%s'",
		c.PostgresAddress, c.PostgresPort, c.PostgresUser, c.PostgresPassword, c.PostgresSSL)
	dbp, err = sql.Open("postgres", dsn)
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
		if err.(*pq.Error).Code == "42710" {
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
	dropDatabase := fmt.Sprintf("DROP DATABASE IF EXISTS %s", c.PostgresDBName)
	createDatabase := fmt.Sprintf("CREATE DATABASE %s", c.PostgresDBName)
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

	// Drop the database and rebuild it to start fresh.
	slog.Debug(fmt.Sprintf("Deleting existing databases named <%s>...", c.PostgresDBName), "func", "postgresPopulate")
	_, err := dbp.Exec(dropDatabase)
	if err != nil {
		slog.Error("Failed to remove existing dbp. Skipping remaining autoconfig steps.", "func", "postgresPopulate", "error", err)
		return err
	}
	slog.Debug(fmt.Sprintf("Creating database <%s>...", c.PostgresDBName), "func", "postgresPopulate")
	_, err = dbp.Exec(createDatabase)
	if err != nil {
		slog.Error("Failed to create database. Skipping remaining autoconfig steps.", "func", "postgresPopulate", "error", err)
		return err
	}
	slog.Debug(fmt.Sprintf("Dropping table <%s>...", c.PostgresTableName), "func", "postgresPopulate")
	_, err = dbp.Exec(dropTable)
	if err != nil {
		slog.Error("Failed to drop table. Skipping remaining autoconfig steps.", "func", "postgresPopulate", "error", err)
		return err
	}
	slog.Debug(fmt.Sprintf("Creating table <%s>...", c.PostgresTableName), "func", "postgresPopulate")
	_, err = dbp.Exec(createTable)
	if err != nil {
		if err.(*pq.Error).Code == "42P07" {
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
		slog.Error(fmt.Sprintf("Could not open music directory <%s> for verification.", c.MusicDir), "func", postgresPopulate, "error", err)
		if os.IsNotExist(err) {
			slog.Error("The configured target music directory was not found.", "func", "postgresPopulate", "error", err)
			return err
		}
	}

	insertInto := fmt.Sprintf("INSERT INTO %s (%s, %s, %s, %s, %s, %s) SELECT $1, $2, $3, $4, $5, $6", c.PostgresTableName, "title", "album", "artist", "genre", "year", "path")
	slog.Debug(fmt.Sprintf("Extracting metadata from audio files in: <%s>", c.MusicDir), "func", "postgresPopulate")
	err = filepath.Walk(c.MusicDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			slog.Error("Error during filepath walk", "func", "postgresPopulate", "error", err)
			return err
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
					slog.Error(fmt.Sprintf("Problem opening directory <%s> for music population.", path), "func", "postgresPopulate", "error", err)
					return err
				}
				defer file.Close()
				tags, err := tag.ReadFrom(file)
				if err != nil {
					slog.Error(fmt.Sprintf("Problem fetching tags from <%s>.", path), "func", "postgresPopulate", "error", err)
					return err
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
	slog.Info("Database population completed.", "func", "postgresPopulate")
	return nil
}
