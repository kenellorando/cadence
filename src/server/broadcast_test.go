package main

import (
	"bytes"
	"testing"
)

// A valid MPEG frame header: eleven sync bits, then layer and bitrate fields.
var frameHeader = []byte{0xFF, 0xFB, 0x90, 0x00}

func TestAudioHubRefusesListenersWithoutASource(t *testing.T) {
	hub := newAudioHub()
	if _, _, ok := hub.subscribe(); ok {
		t.Fatal("a listener should not be accepted while no source is connected")
	}
}

func TestAudioHubDeliversAudioToEveryListener(t *testing.T) {
	hub := newAudioHub()
	hub.sourceConnected("audio/mpeg")

	first, _, ok := hub.subscribe()
	if !ok {
		t.Fatal("first listener rejected")
	}
	second, _, ok := hub.subscribe()
	if !ok {
		t.Fatal("second listener rejected")
	}
	if got := hub.listenerCount(); got != 2 {
		t.Fatalf("listenerCount = %d, want 2", got)
	}

	hub.publish(append(append([]byte{}, frameHeader...), 'a', 'b'))

	for name, listener := range map[string]chan []byte{"first": first, "second": second} {
		select {
		case chunk := <-listener:
			if !bytes.HasSuffix(chunk, []byte("ab")) {
				t.Errorf("%s listener got %q", name, chunk)
			}
		default:
			t.Errorf("%s listener received nothing", name)
		}
	}
}

// The published slice is a reused read buffer, so the hub must not retain it.
func TestAudioHubCopiesPublishedAudio(t *testing.T) {
	hub := newAudioHub()
	hub.sourceConnected("audio/mpeg")
	listener, _, _ := hub.subscribe()

	buffer := append(append([]byte{}, frameHeader...), 'x')
	hub.publish(buffer)
	// Simulate the source reusing its buffer for the next read.
	for i := range buffer {
		buffer[i] = 'z'
	}

	chunk := <-listener
	if bytes.Contains(chunk, []byte("zzz")) {
		t.Errorf("hub retained the caller's buffer, got %q", chunk)
	}
}

// A joining listener gets recent audio so playback starts at once, and it has
// to begin on a frame header or the decoder opens on a partial frame.
func TestAudioHubBurstStartsOnAFrameBoundary(t *testing.T) {
	hub := newAudioHub()
	hub.sourceConnected("audio/mpeg")

	// Mid-frame bytes, then a real frame header.
	hub.publish([]byte{0x11, 0x22, 0x33})
	hub.publish(append(append([]byte{}, frameHeader...), 'p', 'a', 'y'))

	_, burst, ok := hub.subscribe()
	if !ok {
		t.Fatal("listener rejected")
	}
	if len(burst) < 2 || burst[0] != 0xFF || burst[1]&0xE0 != 0xE0 {
		t.Fatalf("burst does not start on a frame header: %x", burst)
	}
	if !bytes.HasSuffix(burst, []byte("pay")) {
		t.Errorf("burst lost the audio after the header: %x", burst)
	}
}

func TestAudioHubBurstIsBounded(t *testing.T) {
	hub := newAudioHub()
	hub.sourceConnected("audio/mpeg")
	for i := 0; i < 40; i++ {
		hub.publish(append(append([]byte{}, frameHeader...), bytes.Repeat([]byte{'d'}, 4096)...))
	}
	_, burst, _ := hub.subscribe()
	if len(burst) > burstBytes {
		t.Errorf("burst is %d bytes, want at most %d", len(burst), burstBytes)
	}
}

// A listener that stops reading is dropped rather than stalling the source,
// which would stop audio for everyone.
func TestAudioHubDropsListenersThatFallBehind(t *testing.T) {
	hub := newAudioHub()
	hub.sourceConnected("audio/mpeg")
	listener, _, _ := hub.subscribe()

	for i := 0; i < listenerBacklog+5; i++ {
		hub.publish(append(append([]byte{}, frameHeader...), byte(i)))
	}

	if got := hub.listenerCount(); got != 0 {
		t.Fatalf("listenerCount = %d, want the stalled listener dropped", got)
	}
	// Draining what it did receive must reach a closed channel rather than block.
	for range listener {
	}
}

// When the source goes away, listeners are disconnected so clients reconnect
// instead of holding a stream that will never produce another byte.
func TestAudioHubDisconnectsListenersWhenTheSourceLeaves(t *testing.T) {
	hub := newAudioHub()
	hub.sourceConnected("audio/mpeg")
	listener, _, _ := hub.subscribe()

	hub.sourceDisconnected()

	for range listener {
	}
	if live, _ := hub.status(); live {
		t.Error("hub still reports a live source")
	}
	if _, _, ok := hub.subscribe(); ok {
		t.Error("listeners should be refused once the source has gone")
	}
}

func TestFrameStart(t *testing.T) {
	for _, tc := range []struct {
		name  string
		audio []byte
		want  int
	}{
		{"header at the start", []byte{0xFF, 0xFB, 0x00}, 0},
		{"header after junk", []byte{0x01, 0x02, 0xFF, 0xE0}, 2},
		{"no header", []byte{0x01, 0x02, 0x03}, -1},
		{"sync bits incomplete", []byte{0xFF, 0x1F}, -1},
		{"empty", nil, -1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := frameStart(tc.audio); got != tc.want {
				t.Errorf("frameStart(%x) = %d, want %d", tc.audio, got, tc.want)
			}
		})
	}
}
