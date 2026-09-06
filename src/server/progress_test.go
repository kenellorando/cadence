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

// The same song twice in a row announces identical metadata, so the change
// cannot be seen there. Time remaining rising is the reliable signal, and
// missing it made the reported duration come out at roughly double the track.
func TestRemainingRisingRestartsTheTrackClock(t *testing.T) {
	suspendTrackClock()
	markTrackStart() // the track already playing when the source connected
	markTrackStart() // a real track start, so the clock is running

	// Most of the way through a four minute track.
	applyRemainingReading(30)
	time.Sleep(20 * time.Millisecond)
	elapsedBefore, _, _ := trackProgress()
	if elapsedBefore <= 0 {
		t.Fatalf("clock should be running, elapsed = %v", elapsedBefore)
	}

	// The same song begins again: remaining jumps back up to a full track.
	applyRemainingReading(240)

	elapsedAfter, duration, known := trackProgress()
	if !known {
		t.Fatal("progress should still be known after a track boundary")
	}
	if elapsedAfter >= elapsedBefore {
		t.Errorf("elapsed = %v, want it reset below %v", elapsedAfter, elapsedBefore)
	}
	// Roughly one track, not two.
	if duration < 200 || duration > 260 {
		t.Errorf("duration = %v, want about one track length", duration)
	}
}

// A falling reading is the normal case and must not be mistaken for a boundary.
func TestRemainingFallingLeavesTheClockAlone(t *testing.T) {
	suspendTrackClock()
	markTrackStart()
	markTrackStart()

	applyRemainingReading(240)
	time.Sleep(20 * time.Millisecond)
	before, _, _ := trackProgress()
	applyRemainingReading(235)
	after, _, _ := trackProgress()

	if after < before {
		t.Errorf("elapsed went backwards on a falling reading: %v then %v", before, after)
	}
}

// Duration used to be recomputed from elapsed on every read, so once remaining
// reached zero the two climbed together and a track ran past its own length.
func TestDurationHoldsWhenRemainingRunsOut(t *testing.T) {
	suspendTrackClock()
	markTrackStart()
	markTrackStart()

	applyRemainingReading(100)
	_, duration, known := trackProgress()
	if !known || duration < 95 || duration > 105 {
		t.Fatalf("duration = %v (known %v), want about 100", duration, known)
	}

	// The end of the track: nothing left to report.
	applyRemainingReading(0)
	time.Sleep(20 * time.Millisecond)
	elapsed, after, _ := trackProgress()

	if after != duration {
		t.Errorf("duration moved from %v to %v once remaining hit zero", duration, after)
	}
	if elapsed > after {
		t.Errorf("elapsed %v ran past the duration %v", elapsed, after)
	}
}
