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
	// How long to keep waiting for Postgres to accept connections at startup,
	// and how long to pause between attempts.
	postgresStartupTimeout = 60 * time.Second
	postgresRetryInterval  = 1 * time.Second
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

// Waits for Postgres to start accepting connections. A fixed sleep was the
// previous approach, which is simultaneously too long on a warm machine and too
// short on a cold one -- and when it was too short, population failed and the
// library stayed empty with nothing reporting it.
func postgresAwait() error {
	deadline := time.Now().Add(postgresStartupTimeout)
	var lastErr error
	for attempt := 1; ; attempt++ {
		maintenance, err := sql.Open("postgres", postgresDSN(maintenanceDatabase))
		if err == nil {
			lastErr = maintenance.Ping()
			maintenance.Close()
			if lastErr == nil {
				return nil
			}
		} else {
			lastErr = err
		}
		if time.Now().After(deadline) {
			slog.Error("Postgres did not become reachable in time.", "func", "postgresAwait", "error", lastErr)
			return lastErr
		}
		slog.Debug(fmt.Sprintf("Postgres not ready yet, retrying (attempt %d).", attempt), "func", "postgresAwait")
		time.Sleep(postgresRetryInterval)
	}
}

func postgresInit() (err error) {
	if err = postgresAwait(); err != nil {
		return err
	}
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
	// Population is additive. Dropping and rebuilding the table renumbered every
	// song on every restart and on every settled change to the library, because
	// the id column is a serial. A search result held in an open tab, or a
	// request submitted moments later, then resolved to a different song
	// entirely -- silently, with a 202 Accepted. It also left a window on each
	// rebuild where the table did not exist and search returned nothing.
	createTable := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s
	(
	   id serial PRIMARY KEY,
	   title character varying(255),
	   album character varying(255),
	   artist character varying(255),
	   genre character varying(255),
	   year character varying(4),
	   path character varying(510)
	)`, c.PostgresTableName)
	// A table built by an older version has no uniqueness on path and may hold
	// duplicates, which would make the index below fail. Collapse them first,
	// keeping the lowest id so existing references stay valid.
	deduplicate := fmt.Sprintf(`DELETE FROM %s a USING %s b
		WHERE a.id > b.id AND a.path = b.path`, c.PostgresTableName, c.PostgresTableName)
	// The path is the identity of a track: it is what upserts key on, and what
	// makes a rebuild leave ids alone.
	createIndex := fmt.Sprintf(`CREATE UNIQUE INDEX IF NOT EXISTS %s_path_idx ON %s (path)`,
		c.PostgresTableName, c.PostgresTableName)
	upsert := fmt.Sprintf(`INSERT INTO %s (title, album, artist, genre, year, path)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (path) DO UPDATE SET
		   title = EXCLUDED.title, album = EXCLUDED.album, artist = EXCLUDED.artist,
		   genre = EXCLUDED.genre, year = EXCLUDED.year`, c.PostgresTableName)

	slog.Debug(fmt.Sprintf("Ensuring table <%s> exists...", c.PostgresTableName), "func", "postgresPopulate")
	if _, err := dbp.Exec(createTable); err != nil {
		slog.Error("Failed to build database table!", "func", "postgresPopulate", "error", err)
		return err
	}
	if _, err := dbp.Exec(deduplicate); err != nil {
		slog.Error("Failed to remove duplicate paths from the metadata table.", "func", "postgresPopulate", "error", err)
		return err
	}
	if _, err := dbp.Exec(createIndex); err != nil {
		slog.Error("Failed to index the metadata table by path.", "func", "postgresPopulate", "error", err)
		return err
	}

	// Search filters with a leading-wildcard ILIKE and ranks with levenshtein,
	// neither of which a B-tree can serve. Trigram indexes cover both, and cost
	// nothing to have on a small library.
	if _, err := dbp.Exec("CREATE EXTENSION IF NOT EXISTS pg_trgm"); err != nil {
		slog.Warn("Couldn't enable pg_trgm; search will scan the whole table.", "func", "postgresPopulate", "error", err)
	} else {
		for _, column := range []string{"title", "artist"} {
			index := fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s_%s_trgm_idx ON %s USING gin (%s gin_trgm_ops)",
				c.PostgresTableName, column, c.PostgresTableName, column)
			if _, err := dbp.Exec(index); err != nil {
				slog.Warn(fmt.Sprintf("Couldn't index %s for search.", column), "func", "postgresPopulate", "error", err)
			}
		}
	}

	slog.Debug("Verifying music metadata directory is accessible.")
	if _, err := os.Stat(c.MusicDir); err != nil {
		slog.Error(fmt.Sprintf("Could not open music directory <%s> for verification.", c.MusicDir), "func", "postgresPopulate", "error", err)
		if os.IsNotExist(err) {
			slog.Error("The configured target music directory was not found.", "func", "postgresPopulate", "error", err)
			return err
		}
	}

	slog.Debug(fmt.Sprintf("Extracting metadata from audio files in: <%s>", c.MusicDir), "func", "postgresPopulate")
	// A file we can't read is a reason to skip that file, not to abandon the
	// rest of the library. Anything skipped here is counted and reported once
	// the walk finishes.
	skipped := 0
	seen := make([]string, 0)
	err := filepath.Walk(c.MusicDir, func(path string, info os.FileInfo, err error) error {
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
				_, err = dbp.Exec(upsert, tags.Title(), tags.Album(), tags.Artist(), tags.Genre(), tags.Year(), path)
				if err != nil {
					slog.Error(fmt.Sprintf("Problem populating metadata for <%s>.", path), "func", "postgresPopulate", "error", err)
					return err
				}
				seen = append(seen, path)
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

	// Only prune once the walk has finished cleanly. Removing everything the
	// walk did not reach would empty the library if it had been abandoned
	// partway, which is exactly when the data is most worth keeping.
	removed, err := dbp.Exec(fmt.Sprintf("DELETE FROM %s WHERE NOT (path = ANY($1))", c.PostgresTableName), pq.Array(seen))
	if err != nil {
		slog.Error("Failed to remove songs that are no longer in the library.", "func", "postgresPopulate", "error", err)
		return err
	}
	if count, err := removed.RowsAffected(); err == nil && count > 0 {
		slog.Info(fmt.Sprintf("Removed %d song(s) no longer present in the library.", count), "func", "postgresPopulate")
	}

	if skipped > 0 {
		slog.Warn(fmt.Sprintf("Database population completed, but %d file(s) were skipped.", skipped), "func", "postgresPopulate")
		return nil
	}
	slog.Info("Database population completed.", "func", "postgresPopulate")
	return nil
}
