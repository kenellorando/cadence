// db_redis.go
// Rate limit database functions.

package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

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
		// A rate limit of zero means the limiter is disabled. This must be handled
		// before anything is written to Redis, which reads a zero expiry as
		// "never expire" and would block the client permanently.
		if c.RequestRateLimit <= 0 {
			next.ServeHTTP(w, r)
			return
		}
		ip, err := checkIP(r)
		if err != nil {
			slog.Error("Couldn't start IP address check for request API.", "func", "rateLimitRequest", "error", err)
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
				slog.Error("Redis reported error while checking for IP.", "func", "rateLimitRequest", "error", err)
				w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
				return
			}
		} else {
			slog.Info(fmt.Sprintf("IP <%s> is rate limited.", ip), "func", "rateLimitRequest")
			w.WriteHeader(http.StatusTooManyRequests) // 429 Too Many Requests
			return
		}
	})
}

func rateLimitArt(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, err := checkIP(r)
		if err != nil {
			slog.Error("Couldn't start IP address check for artwork API.", "func", "rateLimitArt", "error", err)
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
				slog.Error("Redis reported error while checking for IP.", "func", "rateLimitArt", "error", err)
				w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
				return
			}
		} else {
			// If there is no error, the IP is at least in the database.
			// Check the value of the IP address.
			count, err := dbr.RateLimitArt.Get(ctx, ip).Int()
			if err != nil {
				slog.Error("Couldn't get the client's artwork request count.", "func", "rateLimitArt", "error", err)
				w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
				return
			}
			// We're using 16 as an arbitrary maximum number of times we expect any client to need
			// to legitimately need to get album art over the course of a duration of one song.
			// This strikes a balance between allowing a user to get album art when they need it
			// and preventing malicious users from unnecessarily consuming bandwidth.
			//
			// Basically, a 304 response means "You've requested artwork a bit too much,
			// so we're not going to send you new artwork for now. The artwork hasn't changed
			// since you last asked, so you're safe to use whatever you last cached."
			if count >= 16 {
				slog.Info(fmt.Sprintf("IP <%s> is rate limited.", ip), "func", "rateLimitArt")
				w.WriteHeader(http.StatusNotModified) // 304 Not Modified
				return
			} else {
				slog.Debug(fmt.Sprintf("IP <%s> has requested artwork %d time(s) for this song.", ip, count), "func", "rateLimitArt")
				dbr.RateLimitArt.Set(ctx, ip, count+1, time.Duration(200)*time.Second)
				next.ServeHTTP(w, r)
			}
		}
	})
}

func checkIP(r *http.Request) (string, error) {
	// We look at the remote address and check the IP.
	// If for some reason no remote IP is there, we error to reject.
	if r.RemoteAddr == "" {
		slog.Warn("A client address was blank and could not be checked. The request will be rejected.", "func", "checkIP")
		return "", errors.New("request has no remote address")
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		slog.Error("Couldn't split client address IP from port. The request will be rejected.", "func", "checkIP", "error", err)
		return "", err
	}
	if ip == "" {
		slog.Warn("A client IP was blank and could not be checked. The request will be rejected.", "func", "checkIP")
		return "", errors.New("request has a blank remote IP")
	}
	// Requests arriving through the bundled nginx proxy all carry the proxy's
	// own address, which would rate limit every listener as if they were one
	// client. Where the immediate peer is on a private network we take nginx's
	// X-Real-IP instead. The header is ignored for peers reaching the API
	// directly over a public address, where a client could set it themselves
	// to evade the limit.
	if peer := net.ParseIP(ip); peer != nil && (peer.IsLoopback() || peer.IsPrivate()) {
		if forwarded := net.ParseIP(strings.TrimSpace(r.Header.Get("X-Real-IP"))); forwarded != nil {
			return forwarded.String(), nil
		}
	}
	return ip, nil
}
