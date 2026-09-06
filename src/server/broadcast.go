// broadcast.go
// Fan-out of the audio stream from one source to many listeners.

package main

import (
	"log/slog"
	"sync"
)

const (
	// Recent audio held for a joining listener. At 192 kbps this is roughly two
	// and a half seconds, which is what lets playback start immediately instead
	// of waiting for the next few frames to arrive in real time.
	burstBytes = 64 * 1024
	// How far a listener may fall behind before it is dropped. Audio arrives at
	// a fixed rate, so a listener that cannot keep up is not going to recover;
	// holding the write would stall the source for everyone else.
	listenerBacklog = 32
)

// Carries audio from the connected source to every listener.
type audioHub struct {
	mu        sync.Mutex
	listeners map[chan []byte]struct{}
	// Trailing window of the stream, handed to a listener on connect.
	burst []byte
	// Set while a source is connected. Listeners arriving before that get a
	// 503 rather than an open connection that never produces audio.
	live        bool
	contentType string
}

func newAudioHub() *audioHub {
	return &audioHub{listeners: make(map[chan []byte]struct{})}
}

// Marks a source as connected and clears any audio held from a previous one.
func (h *audioHub) sourceConnected(contentType string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.live = true
	h.contentType = contentType
	h.burst = nil
}

// Marks the source gone and disconnects every listener, so clients reconnect
// rather than holding a stream that will never produce another byte.
func (h *audioHub) sourceDisconnected() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.live = false
	h.burst = nil
	for listener := range h.listeners {
		delete(h.listeners, listener)
		close(listener)
	}
}

func (h *audioHub) status() (bool, string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.live, h.contentType
}

func (h *audioHub) listenerCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.listeners)
}

// Sends audio to every listener and keeps a trailing copy for the next one to
// join. The caller's slice is copied: it is a read buffer that will be reused.
func (h *audioHub) publish(chunk []byte) {
	if len(chunk) == 0 {
		return
	}
	audio := make([]byte, len(chunk))
	copy(audio, chunk)

	h.mu.Lock()
	defer h.mu.Unlock()

	h.burst = append(h.burst, audio...)
	if len(h.burst) > burstBytes {
		h.burst = h.burst[len(h.burst)-burstBytes:]
	}

	for listener := range h.listeners {
		select {
		case listener <- audio:
		default:
			slog.Debug("Dropping a listener that fell behind.", "func", "audioHub.publish")
			delete(h.listeners, listener)
			close(listener)
		}
	}
}

// Registers a listener, returning its channel and the audio to send first.
// Reports false when no source is connected.
func (h *audioHub) subscribe() (chan []byte, []byte, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if !h.live {
		return nil, nil, false
	}
	listener := make(chan []byte, listenerBacklog)
	h.listeners[listener] = struct{}{}

	// Starting mid-frame gives the decoder a partial frame to choke on, which is
	// audible as a click or a dropped moment at the start of playback. Begin at
	// a frame header instead, discarding whatever precedes it.
	burst := h.burst
	if offset := frameStart(burst); offset > 0 {
		burst = burst[offset:]
	}
	opening := make([]byte, len(burst))
	copy(opening, burst)
	return listener, opening, true
}

// Removes a listener if it is still registered. publish and sourceDisconnected
// both close the channel when they remove it, so ownership is checked under the
// lock rather than closing here unconditionally.
func (h *audioHub) unsubscribe(listener chan []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.listeners[listener]; ok {
		delete(h.listeners, listener)
		close(listener)
	}
}

// Offset of the first MPEG audio frame header, or -1 if there is none. A frame
// begins with eleven set sync bits.
func frameStart(audio []byte) int {
	for i := 0; i+1 < len(audio); i++ {
		if audio[i] == 0xFF && audio[i+1]&0xE0 == 0xE0 {
			return i
		}
	}
	return -1
}
