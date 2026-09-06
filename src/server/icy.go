// icy.go
// ICY metadata for players that ask for it.

package main

import (
	"fmt"
	"io"
	"strings"
)

const (
	// How much audio goes between metadata blocks. 16000 is what Icecast and
	// Shoutcast have always used, and some players assume it.
	icyMetaInt = 16000
	// A block is announced by one length byte counting 16-byte units, so this is
	// as much metadata as the protocol can carry.
	icyMaxMetadata = 16 * 255
)

// Wraps the audio stream, inserting an ICY metadata block after every
// icyMetaInt bytes. Only used for players that asked for metadata with an
// Icy-MetaData header; a browser never does, so the audio it receives is
// untouched by any of this.
type icyWriter struct {
	dst      io.Writer
	interval int
	// Bytes of audio written since the last metadata block.
	since int
	// The last title sent, so an unchanged one costs a single zero byte.
	last string
	// Reports the title to advertise. Injected so tests do not need radio state.
	title func() string
}

func newICYWriter(dst io.Writer, title func() string) *icyWriter {
	return &icyWriter{dst: dst, interval: icyMetaInt, title: title}
}

// Writes audio, breaking it at metadata boundaries. The block has to land at
// exactly the advertised interval: a player counts bytes to find it, so being
// a byte out turns metadata into noise and audio into a decoding error.
func (w *icyWriter) Write(audio []byte) (int, error) {
	written := 0
	for len(audio) > 0 {
		room := w.interval - w.since
		chunk := audio
		if len(chunk) > room {
			chunk = chunk[:room]
		}
		n, err := w.dst.Write(chunk)
		written += n
		w.since += n
		if err != nil {
			return written, err
		}
		audio = audio[len(chunk):]
		if w.since == w.interval {
			if err := w.writeMetadata(); err != nil {
				return written, err
			}
			w.since = 0
		}
	}
	return written, nil
}

func (w *icyWriter) writeMetadata() error {
	title := w.title()
	if title == w.last {
		// Nothing has changed, so the block is empty: a single zero.
		_, err := w.dst.Write([]byte{0})
		return err
	}
	w.last = title
	_, err := w.dst.Write(icyBlock(title))
	return err
}

// Builds a metadata block: a length byte counting 16-byte units, then the
// payload padded out with zeros.
func icyBlock(title string) []byte {
	payload := fmt.Sprintf("StreamTitle='%s';", icyEscape(title))
	if len(payload) > icyMaxMetadata {
		payload = payload[:icyMaxMetadata]
	}
	units := (len(payload) + 15) / 16
	block := make([]byte, 1+units*16)
	block[0] = byte(units)
	copy(block[1:], payload)
	return block
}

// A title is delimited by a single quote and terminated by a semicolon, and the
// format offers no way to escape either, so both are removed rather than left
// to truncate the field or run it into the next one.
func icyEscape(title string) string {
	return strings.NewReplacer("'", "", ";", "", "\x00", "").Replace(title)
}

// The currently playing title, in the "Artist - Title" form ICY uses.
//
// This is sent as UTF-8. The format predates any encoding agreement and older
// players assume latin-1, so a non-Latin title may render as mojibake in one of
// those -- but the alternative is transliterating or dropping characters, which
// is how a Japanese title became "*****" before Cadence served the stream
// itself. Modern players read UTF-8 correctly.
func icyTitle() string {
	playing := nowPlaying()
	artist, title := playing.Song.Artist, playing.Song.Title
	switch {
	case artist == "" || artist == "-":
		return title
	case title == "" || title == "-":
		return artist
	default:
		return artist + " - " + title
	}
}
