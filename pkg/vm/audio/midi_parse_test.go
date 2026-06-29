package audio

import (
	"encoding/binary"
	"testing"
)

func buildMIDIHeader(division uint16) []byte {
	buf := make([]byte, 0, 14)
	buf = append(buf, []byte("MThd")...)
	buf = append(buf, 0, 0, 0, 6) // header length 6
	buf = append(buf, 0, 0)       // format 0
	buf = append(buf, 0, 1)       // ntrks 1
	div := make([]byte, 2)
	binary.BigEndian.PutUint16(div, division)
	buf = append(buf, div...)
	return buf
}

// TestParseMIDITempoMapTruncatedTrackNoPanic is a regression test for the bug
// where a track-length field that overran the actual data caused a
// slice-out-of-range panic. See docs/bug-hunt-findings.md finding D.
func TestParseMIDITempoMapTruncatedTrackNoPanic(t *testing.T) {
	data := buildMIDIHeader(480)
	// MTrk header declaring length 127, but only a couple of bytes follow.
	data = append(data, []byte("MTrk")...)
	data = append(data, 0, 0, 0, 127) // declared track length far beyond actual data
	data = append(data, 0x00, 0x90)   // delta=0, then a truncated event

	// Must not panic.
	events, ppq := ParseMIDITempoMap(data)

	if ppq != 480 {
		t.Errorf("ppq = %d, want 480", ppq)
	}
	// Always returns at least a default tempo entry at tick 0.
	if len(events) == 0 {
		t.Errorf("expected at least one (default) tempo event, got 0")
	}
}

// TestParseMIDITempoMapValidTrack ensures a well-formed track with a tempo
// event is still parsed correctly (no regression).
func TestParseMIDITempoMapValidTrack(t *testing.T) {
	data := buildMIDIHeader(480)

	// Build a track: delta=0, meta tempo (FF 51 03 + 3 bytes), then end of track.
	var track []byte
	track = append(track, 0x00)             // delta time
	track = append(track, 0xFF, 0x51, 0x03) // meta tempo, length 3
	track = append(track, 0x07, 0xA1, 0x20) // 500000 us/beat (120 BPM)
	track = append(track, 0x00, 0xFF, 0x2F, 0x00) // delta=0, end of track

	data = append(data, []byte("MTrk")...)
	tl := make([]byte, 4)
	binary.BigEndian.PutUint32(tl, uint32(len(track)))
	data = append(data, tl...)
	data = append(data, track...)

	events, ppq := ParseMIDITempoMap(data)
	if ppq != 480 {
		t.Errorf("ppq = %d, want 480", ppq)
	}
	found := false
	for _, e := range events {
		if e.MicrosPerBeat == 500000 {
			found = true
		}
	}
	if !found {
		t.Errorf("expected tempo event with 500000 us/beat, got %+v", events)
	}
}
