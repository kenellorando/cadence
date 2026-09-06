// ratelimit.go
// Per-client rate limiting, backed by the metadata database.

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
)

const (
	// Buckets keep the two limiters independent within one table.
	bucketRequest = "request"
	bucketArt     = "art"
	// How long an artwork allowance lasts, and how many fetches it permits. One
	// song's worth of headroom: enough for a client that legitimately refetches,
	// not enough to make artwork a bandwidth amplifier.
	artWindow = 200 * time.Second
	artLimit  = 16
	// Bounds a limiter query so a slow database cannot hold a request open.
	rateLimitTimeout = 3 * time.Second
)

// Creates the table the limiters share. Counters are per client and expire on
// their own schedule, which Postgres expresses as well as a keystore does --
// and it is a table in a database Cadence already requires, rather than a
// second stateful service operated for one map of integers.
func rateLimitInit() error {
	_, err := dbp.Exec(`CREATE TABLE IF NOT EXISTS ratelimit
	(
	   bucket  character varying(16) NOT NULL,
	   client  character varying(64) NOT NULL,
	   count   integer NOT NULL,
	   expires timestamptz NOT NULL,
	   PRIMARY KEY (bucket, client)
	)`)
	if err != nil {
		slog.Error("Couldn't create the rate limit table.", "func", "rateLimitInit", "error", err)
		return err
	}
	// Expired rows are dead weight; sweeping them keeps the table proportional
	// to live clients rather than to every client ever seen.
	if _, err = dbp.Exec(`CREATE INDEX IF NOT EXISTS ratelimit_expires_idx ON ratelimit (expires)`); err != nil {
		slog.Error("Couldn't index the rate limit table.", "func", "rateLimitInit", "error", err)
		return err
	}
	return nil
}

// Records a hit against a client's allowance and reports how many it has now
// made inside the current window. An expired window restarts at one.
//
// This is deliberately a single statement: read-then-write would let two
// concurrent requests from the same client both observe the old count and both
// be allowed through.
func rateLimitHit(bucket, client string, window time.Duration) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), rateLimitTimeout)
	defer cancel()

	const statement = `INSERT INTO ratelimit AS r (bucket, client, count, expires)
		VALUES ($1, $2, 1, now() + make_interval(secs => $3))
		ON CONFLICT (bucket, client) DO UPDATE
		   SET count   = CASE WHEN r.expires <= now() THEN 1 ELSE r.count + 1 END,
		       expires = CASE WHEN r.expires <= now()
		                      THEN now() + make_interval(secs => $3)
		                      ELSE r.expires END
		RETURNING count`

	var count int
	err := dbp.QueryRowContext(ctx, statement, bucket, client, window.Seconds()).Scan(&count)
	return count, err
}

// Clears artwork allowances, and sweeps anything else that has expired. Called
// when the song changes: the artwork has changed with it, so every client is
// entitled to fetch the new one.
func rateLimitResetArt() {
	if dbp == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), rateLimitTimeout)
	defer cancel()
	if _, err := dbp.ExecContext(ctx, `DELETE FROM ratelimit WHERE bucket = $1 OR expires <= now()`, bucketArt); err != nil {
		slog.Error("Couldn't reset artwork rate limits.", "func", "rateLimitResetArt", "error", err)
	}
}

func rateLimitRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// A rate limit of zero means the limiter is disabled. This must be handled
		// before anything is recorded, since a zero-length window would expire
		// immediately and record a row per request for no purpose.
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
		count, err := rateLimitHit(bucketRequest, ip, time.Duration(c.RequestRateLimit)*time.Second)
		if err != nil {
			slog.Error("Couldn't check the client's request allowance.", "func", "rateLimitRequest", "error", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		// One request per window: anything past the first is over the limit.
		if count > 1 {
			slog.Info(fmt.Sprintf("IP <%s> is rate limited.", ip), "func", "rateLimitRequest")
			w.WriteHeader(http.StatusTooManyRequests) // 429 Too Many Requests
			return
		}
		next.ServeHTTP(w, r)
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
		count, err := rateLimitHit(bucketArt, ip, artWindow)
		if err != nil {
			slog.Error("Couldn't check the client's artwork allowance.", "func", "rateLimitArt", "error", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		// A 304 here means "you have asked for artwork often enough that we are
		// not sending it again for now. It has not changed since you last asked,
		// so whatever you cached is still correct."
		if count > artLimit {
			slog.Info(fmt.Sprintf("IP <%s> is rate limited.", ip), "func", "rateLimitArt")
			w.WriteHeader(http.StatusNotModified) // 304 Not Modified
			return
		}
		slog.Debug(fmt.Sprintf("IP <%s> has requested artwork %d time(s) for this song.", ip, count), "func", "rateLimitArt")
		next.ServeHTTP(w, r)
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
		if forwarded := strings.TrimSpace(r.Header.Get("X-Real-IP")); forwarded != "" {
			if parsed := net.ParseIP(forwarded); parsed != nil {
				return parsed.String(), nil
			}
			slog.Debug("Ignoring an unparseable X-Real-IP header.", "func", "checkIP")
		}
	}
	return ip, nil
}
