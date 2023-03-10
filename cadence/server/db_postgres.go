// db_postgres.go
// Metadata database configuration and population.

package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dhowden/tag"
	"github.com/kenellorando/clog"
	"github.com/lib/pq"
)

var dbp *sql.DB

func postgresInit() (err error) {
	// We wait a bit to give some leeway for Postgres to finish startup.
	time.Sleep(2 * time.Second)
	dsn := fmt.Sprintf("host='%s' port='%s' user='%s' password='%s' sslmode='%s'",
		c.PostgresAddress, c.PostgresPort, c.PostgresUser, c.PostgresPassword, c.PostgresSSL)
	dbp, err = sql.Open("postgres", dsn)
	if err != nil {
		clog.Error("postgresInit", "Could not open connection to database.", err)
		return err
	}
	err = dbp.Ping()
	if err != nil {
		clog.Error("postgresInit", "Could not successfully ping the metadata database.", err)
		return err
	}
	// Enable fuzzystrmatch for levenshtein sorting.
	// This enables the database to return results based on search similarity.
	clog.Debug("postgresInit", "Enabling fuzzystrmatch extension...")
	enableExtension := "CREATE EXTENSION fuzzystrmatch"
	_, err = dbp.Exec(enableExtension)
	if err != nil {
		if err.(*pq.Error).Code == "42710" {
			// 42710 also indicates an existing Postgres instance configured by another Cadence instance is still running.
			clog.Info("postgresInit", "fuzzystrmatch already enabled on metadata database.")
		} else {
			clog.Error("postgresInit", "Failed to enable fuzzystrmatch. Search may be degraded.", err)
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
	clog.Debug("postgresPopulate", fmt.Sprintf("Deleting existing databases named <%s>...", c.PostgresDBName))
	_, err := dbp.Exec(dropDatabase)
	if err != nil {
		clog.Error("postgresPopulate", "Failed to remove existing dbp. Skipping remaining autoconfig steps.", err)
		return err
	}
	clog.Debug("postgresPopulate", fmt.Sprintf("Creating database <%s>...", c.PostgresDBName))
	_, err = dbp.Exec(createDatabase)
	if err != nil {
		clog.Error("postgresPopulate", "Failed to create database. Skipping remaining autoconfig steps.", err)
		return err
	}
	clog.Debug("postgresPopulate", fmt.Sprintf("Dropping table <%s>...", c.PostgresTableName))
	_, err = dbp.Exec(dropTable)
	if err != nil {
		clog.Error("postgresPopulate", "Failed to drop table. Skipping remaining autoconfig steps.", err)
		return err
	}
	clog.Debug("postgresPopulate", fmt.Sprintf("Creating table <%s>...", c.PostgresTableName))
	_, err = dbp.Exec(createTable)
	if err != nil {
		if err.(*pq.Error).Code == "42P07" {
			// 42P10 indicates an existing metadata table configured by another Cadence instance is still running.
			clog.Info("postgresInit", "Metadata database already exists")
		} else {
			clog.Error("postgresPopulate", "Failed to build database table!", err)
			return err
		}
	}
	clog.Debug("postgresPopulate", "Verifying music metadata directory is accessible...")
	_, err = os.Stat(c.MusicDir)
	if err != nil {
		if os.IsNotExist(err) {
			clog.Error("postgresPopulate", "The configured target music directory was not found.", err)
			return err
		}
	}

	insertInto := fmt.Sprintf("INSERT INTO %s (%s, %s, %s, %s, %s, %s) SELECT $1, $2, $3, $4, $5, $6", c.PostgresTableName, "title", "album", "artist", "genre", "year", "path")
	clog.Debug("postgresPopulate", fmt.Sprintf("Extracting metadata from audio files in: <%s>", c.MusicDir))
	err = filepath.Walk(c.MusicDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			clog.Error("postgresPopulate", "Error during filepath walk", err)
			return err
		}
		clog.Debug("postgresPopulate", fmt.Sprintf("Populate analyzing file: <%s>", path))
		if info.IsDir() {
			clog.Debug("postgresPopulate", fmt.Sprintf("<%s> is a directory, skipping.", path))
			return nil
		}
		extensions := []string{".mp3", ".flac", ".ogg"}
		for _, ext := range extensions {
			if strings.HasSuffix(path, ext) {
				file, err := os.Open(path)
				if err != nil {
					clog.Error("postgresPopulate", fmt.Sprintf("A problem occured opening <%s>.", path), err)
					return err
				}
				defer file.Close()
				tags, err := tag.ReadFrom(file)
				if err != nil {
					clog.Error("postgresPopulate", fmt.Sprintf("A problem occured fetching tags from <%s>.", path), err)
					return err
				}
				_, err = dbp.Exec(insertInto, tags.Title(), tags.Album(), tags.Artist(), tags.Genre(), tags.Year(), path)
				if err != nil {
					clog.Error("postgresPopulate", fmt.Sprintf("A problem occured populating metadata for <%s>.", path), err)
					return err
				}
				clog.Debug("postgresPopulate", fmt.Sprintf("Populated: %s by %s", tags.Title(), tags.Artist()))
				break
			}
		}
		return nil
	})
	if err != nil {
		clog.Error("postgresPopulate", "Music metadata database population failed, or may be incomplete.", err)
		return err
	}
	clog.Info("postgresPopulate", "Database population completed.")
	return nil
}
