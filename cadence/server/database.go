// database.go
// Metadata database configuration and population.

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/RediSearch/redisearch-go/redisearch"
	"github.com/dhowden/tag"
	"github.com/kenellorando/clog"
)

var r = RedisClient{}

type RedisClient struct {
	Metadata       *redisearch.Client
	MetadataSchema *redisearch.Schema
}

func dbNewClients() {
	r.Metadata = redisearch.NewClient(c.DatabaseAddress+c.DatabasePort, "metadata")
}

func dbPopulate() error {
	dbBuildSchema()
	clog.Debug("dbPopulate", "Opening given music directory.")
	_, err := os.Stat(c.MusicDir)
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
				doc := redisearch.NewDocument(fmt.Sprint(id), 1.0)
				doc.Set("ID", id).
					Set("Artist", tags.Artist()).
					Set("Title", tags.Title()).
					Set("Album", tags.Album()).
					Set("Genre", tags.Genre()).
					Set("Year", tags.Year()).
					Set("Path", path)
				if err := r.Metadata.Index([]redisearch.Document{doc}...); err != nil {
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

func dbBuildSchema() {
	r.Metadata.Drop()
	r.MetadataSchema = redisearch.NewSchema(redisearch.DefaultOptions).
		AddField(redisearch.NewNumericField("ID")).
		AddField(redisearch.NewTextField("Artist")).
		AddField(redisearch.NewTextField("Title")).
		AddField(redisearch.NewTextField("Album")).
		AddField(redisearch.NewTextField("Genre")).
		AddField(redisearch.NewNumericField("Year")).
		AddField(redisearch.NewTextField("Path"))
	if err := r.Metadata.CreateIndex(r.MetadataSchema); err != nil {
		log.Fatal(err)
	}
}
