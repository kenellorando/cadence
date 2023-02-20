// database.go
// Metadata and rate-limit database configuration and population.

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/RediSearch/redisearch-go/redisearch"
	"github.com/dhowden/tag"
	"github.com/kenellorando/clog"
	"github.com/redis/go-redis/v9"
)

var db = RedisClient{}

type RedisClient struct {
	Metadata       *redisearch.Client
	MetadataSchema *redisearch.Schema
	RateLimit      *redis.Client
}

func newRedisClients() {
	db.Metadata = redisearch.NewClient(c.DatabaseAddress+c.DatabasePort, "metadata")
	db.RateLimit = redis.NewClient(&redis.Options{
		Addr:     c.DatabaseAddress + c.DatabasePort,
		Password: "", // no password set
		DB:       1,  // use default DB
	})
}

func metadataPopulate() error {
	db.Metadata.Drop()
	db.MetadataSchema = redisearch.NewSchema(redisearch.DefaultOptions).
		AddField(redisearch.NewNumericField("ID")).
		AddField(redisearch.NewTextField("Artist")).
		AddField(redisearch.NewTextField("Title")).
		AddField(redisearch.NewTextField("Album")).
		AddField(redisearch.NewTextField("Genre")).
		AddField(redisearch.NewNumericField("Year")).
		AddField(redisearch.NewTextField("Path"))
	if err := db.Metadata.CreateIndex(db.MetadataSchema); err != nil {
		log.Fatal(err)
	}
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
				if err := db.Metadata.Index([]redisearch.Document{doc}...); err != nil {
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

func rateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		_, err := db.RateLimit.Get(ctx, ip).Result()
		if err != nil {
			if err == redis.Nil {
				db.RateLimit.Set(ctx, ip, nil, time.Duration(c.RequestRateLimit)*time.Second)
				next.ServeHTTP(w, r)
			} else {
				w.WriteHeader(http.StatusInternalServerError) // 500
				return
			}
		} else {
			w.WriteHeader(http.StatusTooManyRequests) // 429 Too Many Requests
			return
		}
	})
}
