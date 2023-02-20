// db_redis.go
// Redis clients, with rate limiting and history functions.

package main

import (
	"net/http"
	"time"

	"github.com/RediSearch/redisearch-go/redisearch"
	"github.com/redis/go-redis/v9"
)

var dbr = RedisClient{}

type RedisClient struct {
	Metadata       *redisearch.Client
	MetadataSchema *redisearch.Schema
	RateLimit      *redis.Client
}

func newRedisClients() {
	dbr.Metadata = redisearch.NewClient(c.DatabaseAddress+c.DatabasePort, "metadata")
	dbr.RateLimit = redis.NewClient(&redis.Options{
		Addr:     c.DatabaseAddress + c.DatabasePort,
		Password: "", // no password set
		DB:       1,  // use default DB
	})
}

func rateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		_, err := dbr.RateLimit.Get(ctx, ip).Result()
		if err != nil {
			if err == redis.Nil {
				dbr.RateLimit.Set(ctx, ip, nil, time.Duration(c.RequestRateLimit)*time.Second)
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
