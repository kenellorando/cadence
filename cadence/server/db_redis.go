// db_redis.go
// Rate limit database functions.

package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/kenellorando/clog"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()
var dbr = RedisClient{}

type RedisClient struct {
	RateLimitRequest *redis.Client
	RateLimitArt     *redis.Client
}

func redisInit() {
	dbr.RateLimitRequest = redis.NewClient(&redis.Options{
		Addr:     c.RedisAddress + c.RedisPort,
		Password: "",
		DB:       0,
	})
	dbr.RateLimitArt = redis.NewClient(&redis.Options{
		Addr:     c.RedisAddress + c.RedisPort,
		Password: "",
		DB:       1,
	})
}

func rateLimitRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, err := checkIP(r)
		if err != nil {
			clog.Error("rateLimit", "Error encountered while checking IP address.", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		_, err = dbr.RateLimitRequest.Get(ctx, ip).Result()
		if err != nil {
			if err == redis.Nil {
				// redis.Nil means the IP is not in the database.
				// We create a new entry for the IP which will automatically
				// expire after the configured rate limit time expires.
				dbr.RateLimitRequest.Set(ctx, ip, nil, time.Duration(c.RequestRateLimit)*time.Second)
				next.ServeHTTP(w, r)
			} else {
				clog.Error("rateLimit", "Error while attempting to check for IP in rate limiter.", err)
				w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
				return
			}
		} else {
			clog.Debug("rateLimit", fmt.Sprintf("Client <%s> is rate limited.", ip))
			w.WriteHeader(http.StatusTooManyRequests) // 429 Too Many Requests
			return
		}
	})
}

func rateLimitArt(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, err := checkIP(r)
		if err != nil {
			clog.Error("rateLimit", "Error encountered while checking IP address.", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		_, err = dbr.RateLimitArt.Get(ctx, ip).Result()
		if err != nil {
			if err == redis.Nil {
				// redis.Nil means the IP is not in the database.
				// We create a new entry for the IP with start value 1,
				// representing the first request for art.
				dbr.RateLimitArt.Set(ctx, ip, 1, time.Duration(200)*time.Second)
				next.ServeHTTP(w, r)
			} else {
				clog.Error("rateLimit", "Error while attempting to check for IP in rate limiter.", err)
				w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
				return
			}
		} else {
			// If there is no error, the IP is at least in the database.
			// Check the value of the IP address.
			count, err := dbr.RateLimitArt.Get(ctx, ip).Int()
			if err != nil {
				clog.Error("rateLimit", "Error while converting art served value to integer.", err)
				w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
				return
			}
			// If the IP has requested artwork at least 16 times, deny the request.
			if count >= 16 {
				clog.Debug("rateLimit", fmt.Sprintf("Client <%s> is rate limited.", ip))
				w.WriteHeader(http.StatusTooManyRequests) // 429 Too Many Requests
				return
			} else {
				clog.Debug("rateLimit", fmt.Sprintf("Client <%s> is rate limited.", ip))
				dbr.RateLimitArt.Set(ctx, ip, count+1, time.Duration(200)*time.Second)
				next.ServeHTTP(w, r)
			}
		}
	})
}

func checkIP(r *http.Request) (ip string, err error) {
	// We look at the remote address and check the IP.
	// If for some reason no remote IP is there, we error to reject.
	if r.RemoteAddr != "" {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			clog.Error("checkIP", "Error while splitting client address IP and port. The request will be rejected.", err)
			return "", err
		}
		if ip == "" {
			clog.Warn("checkIP", "IP address of a client was blank, and could not be checked. The request will be rejected.")
			return "", err
		}
		return ip, nil
	}
	return "", err
}
