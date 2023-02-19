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
)

var r = RedisClient{}

type RedisClient struct {
	Metadata *redis.Client
	History  *redis.Client
	Limiter  *redis.Client
}

func dbNewClients() {
	r.Metadata = redis.NewClient(&redis.Options{
		Addr:     c.DatabaseAddress + c.DatabasePort,
		Password: "", // todo: c.DatabasePassword
		DB:       0,
	})
}

func dbPopulate() error {
	err := r.Metadata.FlushDB().Err()
	if err != nil {
		clog.Error("dbPopulate", fmt.Sprintf("Could not flush the metadata database."), err)
		return err
	}
	clog.Debug("dbPopulate", "Opening given music directory.")
	_, err = os.Stat(c.MusicDir)
	if err != nil {
		if os.IsNotExist(err) {
			clog.Error("dbPopulate", "The configured target music directory was not found.", err)
			return err
		}
	}
	id := 0
	clog.Debug("dbPopulate", fmt.Sprintf("Extracting metadata from audio files in: <%s>", c.MusicDir))
	err = filepath.Walk(c.MusicDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		extensions := []string{".mp3", ".flac", ".ogg", ".m4a"}
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
