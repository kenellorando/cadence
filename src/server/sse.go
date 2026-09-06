// sse.go
// Server-sent event delivery for live radio updates.

package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

const (
	// Written to every idle connection. Without traffic a dropped connection is
	// indistinguishable from a quiet one: the browser sees no error, never fires
	// onerror, and never reconnects, so metadata sits frozen while the audio
	// stream plays on. The comment frame also stops proxies and NATs from
	// idling the connection out in the first place.
	heartbeatInterval = 20 * time.Second
	// How long a client may fall behind before it is disconnected. Dropping a
	// slow consumer is recoverable, because the browser reconnects and refetches
	// current state; silently discarding its messages is not, because the
	// connection stays up looking healthy while the page goes stale.
	clientBacklog = 16
	// Advertised to the browser as the delay before it retries a dropped stream.
	clientRetryHint = 5 * time.Second
)

// Broadcasts events to every connected browser.
type eventStream struct {
	mu      sync.Mutex
	clients map[chan []byte]struct{}
	closed  bool
}

func newEventStream() *eventStream {
	return &eventStream{clients: make(map[chan []byte]struct{})}
}

// Queues a message for every connected client. Clients too far behind to accept
// it are disconnected rather than skipped.
func (s *eventStream) Send(event, data string) {
	message := []byte(fmt.Sprintf("event: %s\ndata: %s\n\n", event, data))

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	for client := range s.clients {
		select {
		case client <- message:
		default:
			slog.Debug("Dropping an event stream client that fell behind.", "func", "eventStream.Send")
			delete(s.clients, client)
			close(client)
		}
	}
}

// Disconnects every client. Used on shutdown.
func (s *eventStream) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	s.closed = true
	for client := range s.clients {
		delete(s.clients, client)
		close(client)
	}
}

func (s *eventStream) add() (chan []byte, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil, false
	}
	client := make(chan []byte, clientBacklog)
	s.clients[client] = struct{}{}
	return client, true
}

// Removes a client if it is still registered. Send and Close both close the
// channel when they remove it, so ownership has to be checked under the lock
// rather than closing here unconditionally.
func (s *eventStream) remove(client chan []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.clients[client]; ok {
		delete(s.clients, client)
		close(client)
	}
}

func (s *eventStream) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		slog.Error("Event stream requires a flushable response writer.", "func", "eventStream.ServeHTTP")
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	// Without no-cache an intermediary is free to buffer or cache the stream.
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	// Tells nginx not to buffer this response even where the site config has not
	// been updated to disable buffering for the route.
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	client, ok := s.add()
	if !ok {
		return
	}
	defer s.remove(client)

	if _, err := fmt.Fprintf(w, "retry: %d\n\n", clientRetryHint.Milliseconds()); err != nil {
		return
	}
	flusher.Flush()

	heartbeat := time.NewTicker(heartbeatInterval)
	defer heartbeat.Stop()

	for {
		select {
		case message, open := <-client:
			if !open {
				return
			}
			if _, err := w.Write(message); err != nil {
				return
			}
			flusher.Flush()
		case <-heartbeat.C:
			// A comment frame. Browsers ignore it, but writing it is what surfaces
			// a connection the peer has already gone away from.
			if _, err := fmt.Fprint(w, ": keepalive\n\n"); err != nil {
				return
			}
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}
