package main

import (
	"sync"
	"testing"
	"time"
)

// Mimics icecastMonitor's write pattern against concurrent handler reads.
func TestRadioStateConcurrentAccess(t *testing.T) {
	var wg sync.WaitGroup
	stop := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; ; i++ {
			select {
			case <-stop:
				return
			default:
			}
			radioMutex.Lock()
			now.Song.Title, now.Song.Artist = "t", "a"
			now.Host, now.Mountpoint = "h", "m"
			now.Listeners, now.Bitrate = float64(i), 128
			radioMutex.Unlock()

			radioMutex.Lock()
			history = append(history, playRecord{Title: "t", Artist: "a", Ended: time.Now()})
			if len(history) > 10 {
				history = history[1:]
			}
			radioMutex.Unlock()
		}
	}()

	for r := 0; r < 8; r++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
				}
				_ = nowPlaying().Host
				_ = len(playHistory())
			}
		}()
	}

	time.Sleep(150 * time.Millisecond)
	close(stop)
	wg.Wait()
}
