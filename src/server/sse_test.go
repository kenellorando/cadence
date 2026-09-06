package main

import (
	"bufio"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Reads frames off a live event stream until it sees want, or the deadline
// passes. Runs the read on its own goroutine so a stream that never produces
// the frame fails the test instead of hanging it.
func awaitFrame(t *testing.T, reader *bufio.Reader, want string) {
	t.Helper()
	found := make(chan string, 1)
	go func() {
		var seen strings.Builder
		for {
			line, err := reader.ReadString('\n')
			seen.WriteString(line)
			if strings.Contains(seen.String(), want) {
				found <- ""
				return
			}
			if err != nil {
				found <- seen.String()
				return
			}
		}
	}()

	select {
	case got := <-found:
		if got != "" {
			t.Fatalf("stream ended before %q arrived, saw:\n%s", want, got)
		}
	case <-time.After(3 * time.Second):
		t.Fatalf("timed out waiting for %q", want)
	}
}

func TestEventStreamSendsHeadersAndEvents(t *testing.T) {
	stream := newEventStream()
	defer stream.Close()

	server := httptest.NewServer(stream)
	defer server.Close()

	response, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("connecting to the stream: %v", err)
	}
	defer response.Body.Close()

	if got := response.Header.Get("Content-Type"); got != "text/event-stream" {
		t.Errorf("Content-Type = %q, want text/event-stream", got)
	}
	// A cached event stream would go stale without the page ever knowing.
	if got := response.Header.Get("Cache-Control"); got != "no-cache" {
		t.Errorf("Cache-Control = %q, want no-cache", got)
	}
	if got := response.Header.Get("X-Accel-Buffering"); got != "no" {
		t.Errorf("X-Accel-Buffering = %q, want no", got)
	}

	reader := bufio.NewReader(response.Body)
	// The retry hint is written on connect, so receiving it also confirms the
	// client is registered and safe to broadcast to.
	awaitFrame(t, reader, "retry:")

	// Broadcast until the frame lands: registration completes concurrently with
	// the handler, so a single early Send could be published to nobody.
	done := make(chan struct{})
	defer close(done)
	go func() {
		for {
			select {
			case <-done:
				return
			case <-time.After(20 * time.Millisecond):
				stream.Send("title", "Slow Orbit")
			}
		}
	}()

	awaitFrame(t, reader, "event: title\ndata: Slow Orbit\n\n")
}

// A client that stops reading must be disconnected, not silently skipped: the
// browser reconnects and resyncs, whereas dropped messages leave a connection
// that looks healthy while the page shows a song from long ago.
func TestEventStreamDisconnectsClientsThatFallBehind(t *testing.T) {
	stream := newEventStream()
	defer stream.Close()

	client, ok := stream.add()
	if !ok {
		t.Fatal("could not register a client")
	}

	for i := 0; i < clientBacklog+1; i++ {
		stream.Send("title", "a song")
	}

	stream.mu.Lock()
	remaining := len(stream.clients)
	stream.mu.Unlock()
	if remaining != 0 {
		t.Fatalf("%d clients still registered, want the stalled one dropped", remaining)
	}

	// Draining what it did receive must reach a closed channel rather than block.
	for range client {
	}
}

func TestEventStreamCloseDisconnectsEveryone(t *testing.T) {
	stream := newEventStream()
	client, ok := stream.add()
	if !ok {
		t.Fatal("could not register a client")
	}

	stream.Close()

	if _, open := <-client; open {
		t.Error("client channel should be closed after Close")
	}
	// Sending and closing again must be safe: shutdown races with the monitor
	// goroutine, which keeps publishing until the process exits.
	stream.Send("title", "after close")
	stream.Close()

	if _, ok := stream.add(); ok {
		t.Error("a closed stream should not accept new clients")
	}
}
