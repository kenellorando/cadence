// database.go
// Metadata, history, and rate limiter database configuration and population.

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dhowden/tag"
	"github.com/go-redis/redis"
	"github.com/kenellorando/clog"
	//_ "github.com/nitishm/go-rejson/v4"
)

var r = RedisClient{}

type RedisClient struct {
	Metadata *redis.Client
	History  *redis.Client
	Limiter  *redis.Client
}

func dbInit() {
	r.Metadata = redis.NewClient(&redis.Options{
		Addr:     c.DatabaseAddress + c.DatabasePort,
		Password: "", // todo: c.DatabasePassword
		DB:       0,
	})
	r.History = redis.NewClient(&redis.Options{
		Addr:     c.DatabaseAddress + c.DatabasePort,
		Password: "",
		DB:       1,
	})
	r.Limiter = redis.NewClient(&redis.Options{
		Addr:     c.DatabaseAddress + c.DatabasePort,
		Password: "",
		DB:       2,
	})

	// todo: if no errs, populate metadata
	dbPopulate()
}

// func dbConfig() (newdb *sql.DB, err error) {
// 	clog.Info("dbConfig", "Setting up the database.")
// 	newdb, err = sql.Open("sqlite3", "/cadence/music-metadata.db")
// 	if err != nil {
// 		clog.Error("dbConfig", "Failed to open database file!", err)
// 		return nil, err
// 	}
// 	_, err = newdb.Exec(`DROP TABLE IF EXISTS aria`)
// 	if err != nil {
// 		clog.Error("dbConfig", "Unable to drop existing metadata table.", err)
// 		return nil, err
// 	}
// 	clog.Info("dbConfig", fmt.Sprintf("Building schema for table <%s>...", c.MetadataTable))
// 	_, err = newdb.Exec(`CREATE VIRTUAL TABLE IF NOT EXISTS aria USING FTS5(title,album,artist,genre,year,path)`) // Todo: insert 'aria' through c
// 	if err != nil {
// 		clog.Error("dbConfig", "Failed to build database table!", err)
// 		return nil, err
// 	}
// 	return newdb, nil
// }

func dbPopulate() error {
	clog.Info("dbPopulate", "Running music metadata database population.")
	_, err := os.Stat(c.MusicDir)
	if err != nil {
		if os.IsNotExist(err) {
			clog.Error("dbPopulate", "The configured target music directory was not found.", err)
			return err
		}
	}

	//insertInto := fmt.Sprintf("INSERT INTO %s (%s, %s, %s, %s, %s, %s) SELECT $1, $2, $3, $4, $5, $6", "aria", "title", "album", "artist", "genre", "year", "path")
	clog.Info("dbPopulate", fmt.Sprintf("Extracting metadata from audio files in: <%s>", c.MusicDir))

	id := 0
	err = filepath.Walk(c.MusicDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		extensions := []string{".mp3", ".flac"}
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

				fmt.Printf("count %v", id)

				song := SongData{
					ID:     id,
					Artist: tags.Artist(),
					Title:  tags.Title(),
					Album:  tags.Album(),
					Genre:  tags.Genre(),
					Year:   tags.Year(),
					Path:   path,
				}
				songInsert, _ := json.Marshal(song)

				err = r.Metadata.Set(fmt.Sprint(id), songInsert, 0).Err()

				// _, err = db.Exec(insertInto, tags.Title(), tags.Album(), tags.Artist(),
				// 	tags.Genre(), tags.Year(), path)
				if err != nil {
					clog.Error("dbPopulate", fmt.Sprintf("A problem occured populating metadata for <%s>.", path), err)
					return err
				}
				id++
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
