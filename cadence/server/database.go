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
	_ "github.com/mattn/go-sqlite3"
)

func dbAutoConfig() (*sql.DB, error) {
	clog.Debug("dbAutoConfig", "Starting automatic database configuration and population.")

	newdatabase, err := sql.Open("sqlite3", "/cadence/music-metadata.db")
	if err != nil {
		panic(err)
	}

	// 	// Drop the database if it exists
	// 	clog.Debug("databaseAutoConfig", fmt.Sprintf("Deleting existing databases named <%s>.", c.db.Name))
	// 	_, err := database.Exec(dropDatabase)
	// 	if err != nil {
	// 		clog.Error("databaseAutoConfig", "Failed to remove existing database. Skipping remaining autoconfig steps.", err)
	// 		return err
	// 	}

	// Build the database tables
	clog.Debug("dbAutoConfig", fmt.Sprintf("Reconnected. Building database schema for table <%s>...", c.schema.Table))
	_, err = newdatabase.Exec(`CREATE VIRTUAL TABLE aria USING FTS5
		(
		   title,
		   album,
		   artist,
		   genre,
		   year,
		   path
		)`)
	if err != nil {
		clog.Error("dbAutoConfig", "Failed to build database table!", err)
		return nil, err
	}

	return newdatabase, nil
}

func dbPopulate() error {
	clog.Debug("dbPopulate", "Starting metadata database population.")

	// SQL exec statements here
	insertInto := fmt.Sprintf("INSERT INTO %s (%s, %s, %s, %s, %s, %s) SELECT $1, $2, $3, $4, $5, $6", "aria", "title", "album", "artist", "genre", "year", "path")
	// Check if music directory exists. Return if err
	_, err := os.Stat(c.server.MusicDir)
	if err != nil {
		if os.IsNotExist(err) {
			clog.Error("dbPopulate", "The defined music directory was not found.", err)
			return err
		}
	}
	clog.Debug("dbPopulate", fmt.Sprintf("Extracting metadata from directory: %s", c.server.MusicDir))
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
				fmt.Println(insertInto)
				fmt.Println(tags.Title())
				fmt.Println(tags.Album())
				fmt.Println(tags.Artist())
				fmt.Println(tags.Genre())
				fmt.Println(tags.Year())
				fmt.Println(path)
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
		clog.Error("dbPopulate", "Examination of music file metadata failed!", err)
		return err
	}

	clog.Debug("dbPopulate", "Database population complete.")
	return nil
}

var database *sql.DB // Database abstraction interface
