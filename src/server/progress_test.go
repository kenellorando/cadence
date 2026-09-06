package main

import (
	"testing"
	"time"
)

// Liquidsoap keeps playing across a Cadence restart, so the first track a
// reconnected source announces is already part-way through. Reporting it as
// just started shows a nearly finished song at zero, with its duration shrunk
// to whatever was left.
func TestTrackClockStaysUnsetForATrackAlreadyInProgress(t *testing.T) {
	suspendTrackClock()

	// The track that was already playing when the source connected.
	markTrackStart()
	if _, _, known := trackProgress(); known {
		t.Error("progress should be unknown for a track that was already playing")
	}
	elapsed, _, _ := trackProgress()
	if elapsed != 0 {
		t.Errorf("elapsed = %v, want 0 while the position is unknown", elapsed)
	}

	// The next track starts under observation, so the clock runs again.
	markTrackStart()
	time.Sleep(20 * time.Millisecond)
	elapsed, _, _ = trackProgress()
	if elapsed <= 0 {
		t.Errorf("elapsed = %v, want the clock running for the next track", elapsed)
	}
}
