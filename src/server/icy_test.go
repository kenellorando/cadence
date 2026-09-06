package main

import (
	"bytes"
	"strings"
	"testing"
)

// A metadata block must land at exactly the advertised interval: players count
// bytes to find it, so being one byte out corrupts both the metadata and the
// audio that follows.
func TestICYWriterPlacesBlocksAtTheInterval(t *testing.T) {
	var out bytes.Buffer
	writer := newICYWriter(&out, func() string { return "The Antennas - Midnight Signal" })
	writer.interval = 16

	// Written in awkward sizes, so the writer has to split them itself.
	for _, chunk := range [][]byte{
		bytes.Repeat([]byte{'a'}, 10),
		bytes.Repeat([]byte{'b'}, 10),
		bytes.Repeat([]byte{'c'}, 12),
	} {
		if _, err := writer.Write(chunk); err != nil {
			t.Fatalf("write: %v", err)
		}
	}

	got := out.Bytes()
	// First 16 bytes are audio, then a block, then the rest of the audio.
	if !bytes.HasPrefix(got, bytes.Repeat([]byte{'a'}, 10)) {
		t.Fatalf("audio did not lead the stream: %q", got[:20])
	}
	length := int(got[16])
	if length == 0 {
		t.Fatal("expected a metadata block after the first interval")
	}
	block := string(got[17 : 17+length*16])
	if !strings.Contains(block, "StreamTitle='The Antennas - Midnight Signal';") {
		t.Errorf("block did not carry the title: %q", block)
	}
	// Audio resumes immediately after the block. 32 bytes of audio at an
	// interval of 16 crosses the boundary twice, so the tail is the remaining
	// 16 bytes plus a second, empty block: the title has not changed.
	rest := got[17+length*16:]
	if len(rest) != 17 {
		t.Fatalf("got %d bytes after the block, want 16 of audio and an empty block", len(rest))
	}
	wantTail := append(bytes.Repeat([]byte{'b'}, 4), bytes.Repeat([]byte{'c'}, 12)...)
	if !bytes.Equal(rest[:16], wantTail) {
		t.Errorf("audio did not resume cleanly after the block: got %q want %q", rest[:16], wantTail)
	}
	if rest[16] != 0 {
		t.Errorf("trailing block length = %d, want 0 for an unchanged title", rest[16])
	}
}

// Bytes reported to the caller must count audio only. Reporting the metadata
// too would tell the hub more was written than it handed over.
func TestICYWriterReportsAudioBytesOnly(t *testing.T) {
	var out bytes.Buffer
	writer := newICYWriter(&out, func() string { return "song" })
	writer.interval = 16

	audio := bytes.Repeat([]byte{'x'}, 48)
	n, err := writer.Write(audio)
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	if n != len(audio) {
		t.Errorf("reported %d bytes written, want %d", n, len(audio))
	}
	if out.Len() <= len(audio) {
		t.Error("expected metadata blocks to have been emitted alongside the audio")
	}
}

// An unchanged title costs one zero byte rather than a repeated block.
func TestICYWriterRepeatsNothingWhileTheTitleHolds(t *testing.T) {
	var out bytes.Buffer
	writer := newICYWriter(&out, func() string { return "steady" })
	writer.interval = 4

	if _, err := writer.Write(bytes.Repeat([]byte{'x'}, 12)); err != nil {
		t.Fatalf("write: %v", err)
	}
	got := out.Bytes()

	first := int(got[4])
	if first == 0 {
		t.Fatal("the first block should carry the title")
	}
	// After the first block: 4 audio bytes, then an empty block.
	next := 5 + first*16 + 4
	if got[next] != 0 {
		t.Errorf("second block length = %d, want 0 for an unchanged title", got[next])
	}
}

func TestICYWriterSendsANewBlockWhenTheTitleChanges(t *testing.T) {
	var out bytes.Buffer
	title := "first"
	writer := newICYWriter(&out, func() string { return title })
	writer.interval = 4

	if _, err := writer.Write(bytes.Repeat([]byte{'x'}, 4)); err != nil {
		t.Fatalf("write: %v", err)
	}
	title = "second"
	if _, err := writer.Write(bytes.Repeat([]byte{'x'}, 4)); err != nil {
		t.Fatalf("write: %v", err)
	}

	if !bytes.Contains(out.Bytes(), []byte("StreamTitle='second';")) {
		t.Errorf("the new title was not announced: %q", out.Bytes())
	}
}

func TestICYBlockIsWellFormed(t *testing.T) {
	block := icyBlock("Vega Drift - Slow Orbit")
	units := int(block[0])
	if len(block) != 1+units*16 {
		t.Fatalf("block is %d bytes, want %d for %d units", len(block), 1+units*16, units)
	}
	payload := string(bytes.TrimRight(block[1:], "\x00"))
	if payload != "StreamTitle='Vega Drift - Slow Orbit';" {
		t.Errorf("payload = %q", payload)
	}
}

// The format has no escape for its own delimiters, so they are removed rather
// than left to truncate the field.
func TestICYEscapeRemovesDelimiters(t *testing.T) {
	got := icyEscape("Don't Stop; Believin'")
	if strings.ContainsAny(got, "';") {
		t.Errorf("delimiters survived escaping: %q", got)
	}
}

// A title longer than the format can carry is truncated rather than overflowing
// the single length byte.
func TestICYBlockTruncatesOverlongTitles(t *testing.T) {
	block := icyBlock(strings.Repeat("x", icyMaxMetadata*2))
	if len(block) > 1+icyMaxMetadata {
		t.Errorf("block is %d bytes, want at most %d", len(block), 1+icyMaxMetadata)
	}
	if int(block[0]) > 255 {
		t.Error("length byte overflowed")
	}
}
