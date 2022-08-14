// database.go
// SQLite configuration and population

package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dhowden/tag"
	"github.com/kenellorando/clog"
	_ "github.com/mattn/go-sqlite3"
)

func dbConfig() (newdb *sql.DB, err error) {
	clog.Info("dbConfig", "Setting up the database.")
	newdb, err = sql.Open("sqlite3", "/cadence/music-metadata.db")
	if err != nil {
		clog.Error("dbAutoConfig", "Failed to build database table!", err)
		return nil, err
	}
	clog.Info("dbAutoConfig", fmt.Sprintf("Building schema for table <%s>...", c.MetadataTable))
	_, err = newdb.Exec(`CREATE VIRTUAL TABLE IF NOT EXISTS aria USING FTS5(title,album,artist,genre,year,path)`) // Todo: insert 'aria' through c
	if err != nil {
		clog.Error("dbAutoConfig", "Failed to build database table!", err)
		return nil, err
	}
	return newdb, nil
}

func dbPopulate() error {
	clog.Info("dbPopulate", "Running music metadata database population.")
	_, err := os.Stat(c.MusicDir)
	if err != nil {
		if os.IsNotExist(err) {
			clog.Error("dbPopulate", "The configured target music directory was not found.", err)
			return err
		}
	}

	insertInto := fmt.Sprintf("INSERT INTO %s (%s, %s, %s, %s, %s, %s) SELECT $1, $2, $3, $4, $5, $6", "aria", "title", "album", "artist", "genre", "year", "path")
	clog.Info("dbPopulate", fmt.Sprintf("Extracting metadata from audio files in: <%s>", c.MusicDir))
	err = filepath.Walk(c.MusicDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		extensions := []string{".mp3", ".flac", ".ogg"}
		for _, ext := range extensions {
			if strings.HasSuffix(path, ext) {
				file, err := os.Open(path)
				defer file.Close()
				if err != nil {
					clog.Error("dbPopulate", fmt.Sprintf("A problem occured opening <%s>.", path), err)
					return err
				}

				tags, err := tag.ReadFrom(file)
				if err != nil {
					clog.Error("dbPopulate", fmt.Sprintf("A problem occured fetching tags from <%s>.", path), err)
					return err
				}

				_, err = db.Exec(insertInto, tags.Title(), tags.Album(), tags.Artist(),
					tags.Genre(), tags.Year(), path)
				if err != nil {
					clog.Error("dbPopulate", fmt.Sprintf("A problem occured populating metadata for <%s>.", path), err)
					return err
				}
				break
			}
		}
		return nil
	})
	if err != nil {
		clog.Error("dbPopulate", "Music metadata database population failed, or may be incomplete.", err)
		return err
	}

	clog.Info("dbPopulate", "Database population completed.")
	return nil
}
