// progress.go
// How far through the current track the broadcast is.

package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

// How often the audio source is asked how much of the track is left. The answer
// decays predictably between polls, so the page can count down locally and this
// only has to correct the drift.
const progressPollInterval = 5 * time.Second

var (
	progressMutex sync.RWMutex
	// When the current track started, taken from the moment its metadata
	// arrived rather than from any clock the source keeps.
	trackStartedAt time.Time
	// Seconds left when the source was last asked, and when that was.
	trackRemaining float64
	trackPolledAt  time.Time
	// Set when a source connects to a broadcast already in progress. The first
	// song it announces has been playing for an unknown length of time, and
	// starting the clock then would report a nearly finished track as just
	// beginning -- with a duration shrunk to whatever was left of it.
	trackPositionUnknown bool
)

// Called when a source connects. Liquidsoap keeps playing across a Cadence
// restart, so the track it announces first is already part-way through and
// there is no way to ask how far: the source reports time remaining, not
// elapsed, and its request metadata carries no duration. Rather than showing a
// wrong position, the clock stays unset until the next real track change.
func suspendTrackClock() {
	progressMutex.Lock()
	defer progressMutex.Unlock()
	trackStartedAt = time.Time{}
	trackRemaining = 0
	trackPolledAt = time.Time{}
	trackPositionUnknown = true
}

// Starts the clock for a new track. Called when the song actually changes, not
// on every metadata update, so a repeated announcement does not rewind it.
func markTrackStart() {
	progressMutex.Lock()
	defer progressMutex.Unlock()
	if trackPositionUnknown {
		// This is the track that was already playing when the source connected.
		// The one after it starts under observation, so the clock resumes then.
		trackPositionUnknown = false
		return
	}
	trackStartedAt = time.Now()
	trackRemaining = 0
	trackPolledAt = time.Time{}
}

// Elapsed and total seconds for the current track. Duration is only known once
// the source has answered, so it reports false until then and the page shows
// elapsed time without a bar it cannot yet size.
func trackProgress() (elapsed float64, duration float64, known bool) {
	progressMutex.RLock()
	defer progressMutex.RUnlock()

	if trackStartedAt.IsZero() {
		return 0, 0, false
	}
	elapsed = time.Since(trackStartedAt).Seconds()
	if trackPolledAt.IsZero() {
		return elapsed, 0, false
	}
	// The reading ages at one second per second.
	remaining := trackRemaining - time.Since(trackPolledAt).Seconds()
	if remaining < 0 {
		remaining = 0
	}
	return elapsed, elapsed + remaining, true
}

// Asks the audio source how much of the current track is left. The command is
// named for the output id, which the config sets to the mount name.
func liquidsoapRemaining() (float64, error) {
	mountpoint := nowPlaying().Mountpoint
	if mountpoint == "" || mountpoint == "-" {
		return 0, fmt.Errorf("no mount is connected")
	}
	conn, err := net.DialTimeout("tcp", c.LiquidsoapAddress+c.LiquidsoapPort, serviceTimeout)
	if err != nil {
		return 0, err
	}
	defer conn.Close()
	if err = conn.SetDeadline(time.Now().Add(serviceTimeout)); err != nil {
		return 0, err
	}
	fmt.Fprintf(conn, "%s.remaining\n", mountpoint)
	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		return 0, err
	}
	fmt.Fprint(conn, "quit\n")

	remaining, err := strconv.ParseFloat(strings.TrimSpace(line), 64)
	if err != nil {
		return 0, fmt.Errorf("unexpected reply %q: %w", strings.TrimSpace(line), err)
	}
	return remaining, nil
}

// Keeps the remaining-time reading fresh.
func trackProgressMonitor() {
	for {
		time.Sleep(progressPollInterval)
		if live, _ := hub.status(); !live {
			continue
		}
		remaining, err := liquidsoapRemaining()
		if err != nil {
			slog.Debug("Couldn't read remaining track time.", "func", "trackProgressMonitor", "error", err)
			continue
		}
		progressMutex.Lock()
		trackRemaining = remaining
		trackPolledAt = time.Now()
		progressMutex.Unlock()
	}
}
