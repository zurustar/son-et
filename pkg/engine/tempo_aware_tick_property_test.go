package engine

import (
	"math"
	"testing"
	"testing/quick"
)

// Feature: gomidi-midi-playback, Property 12: Tempo-Aware Tick Calculation
// **Validates: Requirements 8.1, 8.2, 8.3**
//
// Property: For any tempo map and elapsed time, the MIDI_Player SHALL calculate
// FILLY ticks correctly accounting for all tempo changes in the tempo map.
func TestProperty_TempoAwareTickCalculation(t *testing.T) {
	f := func(elapsedMs uint16, tempo1 uint16, tempo2 uint16, changePointMs uint8) bool {
		// Constrain inputs to reasonable ranges
		if elapsedMs == 0 {
			return true // Skip zero elapsed time
		}
		if tempo1 < 60 || tempo1 > 240 {
			return true // Skip invalid tempo (60-240 BPM)
		}
		if tempo2 < 60 || tempo2 > 240 {
			return true // Skip invalid tempo (60-240 BPM)
		}

		// Convert to seconds
		elapsed := float64(elapsedMs) / 1000.0
		changePoint := float64(changePointMs) / 1000.0

		// Ensure change point is before elapsed time
		if changePoint >= elapsed {
			changePoint = elapsed / 2.0 // Put change point at midpoint
		}

		ppq := 480

		// Create tempo map with two tempos
		// Convert BPM to microseconds per beat
		microsPerBeat1 := int(60000000 / float64(tempo1))
		microsPerBeat2 := int(60000000 / float64(tempo2))

		// Calculate tick at which tempo changes
		// At tempo1, how many ticks occur in changePoint seconds?
		timePerTick1 := (float64(microsPerBeat1) / 1000000.0) / float64(ppq)
		changePointTick := int(changePoint / timePerTick1)

		tempoMap := []TempoEvent{
			{Tick: 0, MicrosPerBeat: microsPerBeat1},
			{Tick: changePointTick, MicrosPerBeat: microsPerBeat2},
		}

		// Create tick generator
		tg := NewWallClockTickGenerator(44100, ppq, tempoMap)

		// Calculate MIDI tick from elapsed time
		midiTick := tg.CalculateTickFromTime(elapsed)

		// Manually calculate expected MIDI tick
		var expectedMidiTick int
		if elapsed <= changePoint {
			// All time is in first tempo
			expectedMidiTick = int(elapsed / timePerTick1)
		} else {
			// Time spans both tempos
			// Ticks in first tempo segment
			ticksInFirstSegment := changePointTick

			// Remaining time after tempo change
			remainingTime := elapsed - changePoint
			timePerTick2 := (float64(microsPerBeat2) / 1000000.0) / float64(ppq)
			ticksInSecondSegment := int(remainingTime / timePerTick2)

			expectedMidiTick = ticksInFirstSegment + ticksInSecondSegment
		}

		// Allow tolerance for rounding errors (±2 ticks)
		diff := int(math.Abs(float64(midiTick - expectedMidiTick)))
		if diff > 2 {
			t.Logf("Tempo-aware tick calculation mismatch:")
			t.Logf("  elapsed=%.3fs, tempo1=%d BPM, tempo2=%d BPM, changePoint=%.3fs",
				elapsed, tempo1, tempo2, changePoint)
			t.Logf("  changePointTick=%d, midiTick=%d, expected=%d, diff=%d",
				changePointTick, midiTick, expectedMidiTick, diff)
			return false
		}

		// Now verify FILLY tick conversion
		// FILLY tick = MIDI tick * 8 / PPQ
		fillyTick := midiTick * 8 / ppq
		expectedFillyTick := expectedMidiTick * 8 / ppq

		// Allow tolerance for integer division rounding
		fillyDiff := int(math.Abs(float64(fillyTick - expectedFillyTick)))
		if fillyDiff > 1 {
			t.Logf("FILLY tick conversion mismatch:")
			t.Logf("  midiTick=%d, fillyTick=%d, expectedFillyTick=%d, diff=%d",
				midiTick, fillyTick, expectedFillyTick, fillyDiff)
			return false
		}

		return true
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(f, config); err != nil {
		t.Error(err)
	}
}

// TestProperty_TempoAwareTickCalculation_SingleTempo tests the special case
// of a single tempo (no tempo changes).
func TestProperty_TempoAwareTickCalculation_SingleTempo(t *testing.T) {
	f := func(elapsedMs uint16, tempo uint16) bool {
		// Constrain inputs to reasonable ranges
		if elapsedMs == 0 {
			return true // Skip zero elapsed time
		}
		if tempo < 60 || tempo > 240 {
			return true // Skip invalid tempo (60-240 BPM)
		}

		// Convert to seconds
		elapsed := float64(elapsedMs) / 1000.0

		ppq := 480

		// Create tempo map with single tempo
		microsPerBeat := int(60000000 / float64(tempo))
		tempoMap := []TempoEvent{
			{Tick: 0, MicrosPerBeat: microsPerBeat},
		}

		// Create tick generator
		tg := NewWallClockTickGenerator(44100, ppq, tempoMap)

		// Calculate MIDI tick from elapsed time
		midiTick := tg.CalculateTickFromTime(elapsed)

		// Manually calculate expected MIDI tick
		timePerTick := (float64(microsPerBeat) / 1000000.0) / float64(ppq)
		expectedMidiTick := int(elapsed / timePerTick)

		// Allow tolerance for rounding errors (±2 ticks)
		diff := int(math.Abs(float64(midiTick - expectedMidiTick)))
		if diff > 2 {
			t.Logf("Single tempo tick calculation mismatch:")
			t.Logf("  elapsed=%.3fs, tempo=%d BPM, midiTick=%d, expected=%d, diff=%d",
				elapsed, tempo, midiTick, expectedMidiTick, diff)
			return false
		}

		return true
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(f, config); err != nil {
		t.Error(err)
	}
}

// TestProperty_TempoAwareTickCalculation_MultipleTempos tests tempo maps
// with multiple tempo changes.
func TestProperty_TempoAwareTickCalculation_MultipleTempos(t *testing.T) {
	f := func(elapsedMs uint16, tempo1 uint8, tempo2 uint8, tempo3 uint8) bool {
		// Constrain inputs to reasonable ranges
		if elapsedMs == 0 {
			return true // Skip zero elapsed time
		}

		// Map uint8 to BPM range (60-240)
		bpm1 := 60 + int(tempo1)%181
		bpm2 := 60 + int(tempo2)%181
		bpm3 := 60 + int(tempo3)%181

		// Convert to seconds
		elapsed := float64(elapsedMs) / 1000.0

		ppq := 480

		// Create tempo map with three tempos at evenly spaced intervals
		microsPerBeat1 := int(60000000 / float64(bpm1))
		microsPerBeat2 := int(60000000 / float64(bpm2))
		microsPerBeat3 := int(60000000 / float64(bpm3))

		// Calculate tempo change points
		timePerTick1 := (float64(microsPerBeat1) / 1000000.0) / float64(ppq)
		changePoint1 := elapsed / 3.0
		changePointTick1 := int(changePoint1 / timePerTick1)

		timePerTick2 := (float64(microsPerBeat2) / 1000000.0) / float64(ppq)
		changePoint2 := 2.0 * elapsed / 3.0
		remainingTime1 := changePoint2 - changePoint1
		ticksInSegment2 := int(remainingTime1 / timePerTick2)
		changePointTick2 := changePointTick1 + ticksInSegment2

		tempoMap := []TempoEvent{
			{Tick: 0, MicrosPerBeat: microsPerBeat1},
			{Tick: changePointTick1, MicrosPerBeat: microsPerBeat2},
			{Tick: changePointTick2, MicrosPerBeat: microsPerBeat3},
		}

		// Create tick generator
		tg := NewWallClockTickGenerator(44100, ppq, tempoMap)

		// Calculate MIDI tick from elapsed time
		midiTick := tg.CalculateTickFromTime(elapsed)

		// Manually calculate expected MIDI tick
		// Segment 1: 0 to changePoint1
		ticksInSegment1 := changePointTick1

		// Segment 2: changePoint1 to changePoint2
		// Already calculated as ticksInSegment2

		// Segment 3: changePoint2 to elapsed
		remainingTime2 := elapsed - changePoint2
		timePerTick3 := (float64(microsPerBeat3) / 1000000.0) / float64(ppq)
		ticksInSegment3 := int(remainingTime2 / timePerTick3)

		expectedMidiTick := ticksInSegment1 + ticksInSegment2 + ticksInSegment3

		// Allow larger tolerance for multiple tempo changes (±5 ticks)
		diff := int(math.Abs(float64(midiTick - expectedMidiTick)))
		if diff > 5 {
			t.Logf("Multiple tempo tick calculation mismatch:")
			t.Logf("  elapsed=%.3fs, bpm1=%d, bpm2=%d, bpm3=%d",
				elapsed, bpm1, bpm2, bpm3)
			t.Logf("  changePoints=[%.3fs, %.3fs], changeTicks=[%d, %d]",
				changePoint1, changePoint2, changePointTick1, changePointTick2)
			t.Logf("  midiTick=%d, expected=%d, diff=%d",
				midiTick, expectedMidiTick, diff)
			return false
		}

		return true
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(f, config); err != nil {
		t.Error(err)
	}
}
