package main

import (
	"sync"
	"testing"
	"time"

	"github.com/Jeffail/gabs"
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

// Icecast 2.5 renamed the source bitrate field, which used to panic the monitor
// goroutine and crash-loop the server. Both shapes must parse.
func TestParseIcecastStatus(t *testing.T) {
	const icecast24 = `{"icestats":{"host":"radio.example.com","source":{"artist":"The Antennas","title":"Midnight Signal","server_name":"cadence1","listeners":3,"bitrate":192}}}`
	const icecast25 = `{"icestats":{"host":"radio.example.com","source":{"artist":"The Antennas","title":"Midnight Signal","server_name":"cadence1","listeners":3,"ice-bitrate":192,"audio_bitrate":192000}}}`
	// 2.5 without ice-bitrate, leaving only the bits-per-second field.
	const bpsOnly = `{"icestats":{"host":"radio.example.com","source":{"artist":"The Antennas","title":"Midnight Signal","server_name":"cadence1","listeners":3,"audio_bitrate":192000}}}`

	for _, tc := range []struct{ name, body string }{
		{"icecast 2.4", icecast24},
		{"icecast 2.5", icecast25},
		{"bitrate in bps only", bpsOnly},
	} {
		t.Run(tc.name, func(t *testing.T) {
			parsed, err := gabs.ParseJSON([]byte(tc.body))
			if err != nil {
				t.Fatalf("parsing fixture: %v", err)
			}
			info, playing := parseIcecastStatus(parsed, RadioInfo{})
			if !playing {
				t.Fatal("expected a playing stream")
			}
			if info.Song.Artist != "The Antennas" || info.Song.Title != "Midnight Signal" {
				t.Errorf("got artist %q title %q", info.Song.Artist, info.Song.Title)
			}
			if info.Host != "radio.example.com" || info.Mountpoint != "cadence1" {
				t.Errorf("got host %q mountpoint %q", info.Host, info.Mountpoint)
			}
			if info.Listeners != 3 {
				t.Errorf("got listeners %v, want 3", info.Listeners)
			}
			if info.Bitrate != 192 {
				t.Errorf("got bitrate %v kbps, want 192", info.Bitrate)
			}
		})
	}
}

// A document missing the fields entirely must report "nothing playing" rather
// than panicking or blanking out the state we already had.
func TestParseIcecastStatusNoSource(t *testing.T) {
	parsed, err := gabs.ParseJSON([]byte(`{"icestats":{"host":"radio.example.com","server_id":"Icecast 2.5.0"}}`))
	if err != nil {
		t.Fatalf("parsing fixture: %v", err)
	}
	previous := RadioInfo{Host: "kept", Mountpoint: "kept", Listeners: 7, Bitrate: 192}
	info, playing := parseIcecastStatus(parsed, previous)
	if playing {
		t.Error("expected nothing playing when the source object is absent")
	}
	if info != previous {
		t.Errorf("state should be returned untouched, got %+v", info)
	}
}

// MP3 and AAC mounts carry ICY metadata: one combined "Artist - Title" string
// and no artist field. Requiring a separate artist reported those streams as
// silent, which is what kept Cadence from serving a format iOS can play.
func TestParseIcecastStatusICYMetadata(t *testing.T) {
	const icy = `{"icestats":{"host":"radio.example.com","source":{"title":"Vega Drift - Slow Orbit","server_name":"cadence1","listeners":2,"ice-bitrate":192}}}`
	parsed, err := gabs.ParseJSON([]byte(icy))
	if err != nil {
		t.Fatalf("parsing fixture: %v", err)
	}
	info, playing := parseIcecastStatus(parsed, RadioInfo{})
	if !playing {
		t.Fatal("an MP3 mount with ICY metadata should count as playing")
	}
	if info.Song.Artist != "Vega Drift" {
		t.Errorf("got artist %q, want %q", info.Song.Artist, "Vega Drift")
	}
	if info.Song.Title != "Slow Orbit" {
		t.Errorf("got title %q, want %q", info.Song.Title, "Slow Orbit")
	}
}

// A title with no separator carries no artist to recover, and a hyphen in the
// song name must not be mistaken for one.
func TestParseIcecastStatusICYEdgeCases(t *testing.T) {
	for _, tc := range []struct {
		name, title  string
		playing      bool
		artist, want string
	}{
		{"no separator", "Untitled", false, "", ""},
		{"hyphen inside title", "Vega Drift - Slow - Orbit", true, "Vega Drift", "Slow - Orbit"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := `{"icestats":{"source":{"title":"` + tc.title + `"}}}`
			parsed, err := gabs.ParseJSON([]byte(body))
			if err != nil {
				t.Fatalf("parsing fixture: %v", err)
			}
			info, playing := parseIcecastStatus(parsed, RadioInfo{})
			if playing != tc.playing {
				t.Fatalf("playing = %v, want %v", playing, tc.playing)
			}
			if playing && (info.Song.Artist != tc.artist || info.Song.Title != tc.want) {
				t.Errorf("got %q / %q, want %q / %q", info.Song.Artist, info.Song.Title, tc.artist, tc.want)
			}
		})
	}
}

// MP3 and AAC mounts publish no bitrate field, only an audio_info string.
func TestParseIcecastStatusAudioInfoBitrate(t *testing.T) {
	const body = `{"icestats":{"source":{"artist":"Vega Drift","title":"Slow Orbit","audio_info":"channels=2;samplerate=44100;bitrate=192"}}}`
	parsed, err := gabs.ParseJSON([]byte(body))
	if err != nil {
		t.Fatalf("parsing fixture: %v", err)
	}
	info, playing := parseIcecastStatus(parsed, RadioInfo{})
	if !playing {
		t.Fatal("expected a playing stream")
	}
	if info.Bitrate != 192 {
		t.Errorf("got bitrate %v, want 192", info.Bitrate)
	}
}
