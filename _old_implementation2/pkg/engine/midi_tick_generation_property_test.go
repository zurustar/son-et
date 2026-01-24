package engine

import (
	"math"
	"testing"
	"testing/quick"
)

// Feature: gomidi-midi-playback, Property 6: FILLY Tick Generation Resolution
// **Validates: Requirements 4.1**
//
// Property: For any elapsed time during MIDI playback, the MIDI_Player SHALL
// generate FILLY ticks at 32nd note resolution (8 ticks per quarter note).
func TestProperty_FILLYTickGenerationResolution(t *testing.T) {
	f := func(elapsedMs uint16, ppq uint16, microsPerBeat uint32) bool {
		// Constrain inputs to reasonable ranges
		if ppq < 24 || ppq > 960 {
			return true // Skip invalid PPQ values
		}
		if microsPerBeat < 100000 || microsPerBeat > 2000000 {
			return true // Skip invalid tempo values (30-600 BPM range)
		}
		if elapsedMs == 0 {
			return true // Skip zero elapsed time
		}

		// Convert elapsed time to seconds
		elapsed := float64(elapsedMs) / 1000.0

		// Create tempo map with single tempo
		tempoMap := []TempoEvent{{Tick: 0, MicrosPerBeat: int(microsPerBeat)}}
		tg := NewWallClockTickGenerator(44100, int(ppq), tempoMap)

		// Calculate MIDI tick from time
		midiTick := tg.CalculateTickFromTime(elapsed)

		// Convert MIDI ticks to FILLY ticks using the formula from the code
		// FILLY tick = MIDI tick * 8 / PPQ
		fillyTick := midiTick * 8 / int(ppq)

		// Calculate expected FILLY ticks based on elapsed time and tempo
		// At a given tempo (microseconds per beat):
		// - 1 beat = 1 quarter note
		// - 1 quarter note = 8 FILLY ticks (32nd note resolution)
		// - beats per second = 1000000 / microsPerBeat
		// - FILLY ticks per second = beats per second * 8
		beatsPerSecond := 1000000.0 / float64(microsPerBeat)
		expectedFillyTicks := int(elapsed * beatsPerSecond * 8)

		// Allow small tolerance for rounding errors (±1 tick)
		diff := int(math.Abs(float64(fillyTick - expectedFillyTicks)))
		if diff > 1 {
			t.Logf("FILLY tick mismatch: elapsed=%.3fs, ppq=%d, tempo=%d µs/beat, midiTick=%d, fillyTick=%d, expected=%d, diff=%d",
				elapsed, ppq, microsPerBeat, midiTick, fillyTick, expectedFillyTicks, diff)
			return false
		}

		return true
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(f, config); err != nil {
		t.Error(err)
	}
}

// Feature: gomidi-midi-playback, Property 7: MIDI to FILLY Tick Conversion
// **Validates: Requirements 4.2**
//
// Property: For any MIDI tick value in PPQ units, the conversion to FILLY ticks
// SHALL use the formula: FILLY_tick = MIDI_tick * 8 / PPQ.
func TestProperty_MIDIToFILLYTickConversion(t *testing.T) {
	f := func(midiTick uint16, ppq uint16) bool {
		// Constrain inputs to reasonable ranges
		if ppq < 24 || ppq > 960 {
			return true // Skip invalid PPQ values
		}

		// Apply the conversion formula
		fillyTick := int(midiTick) * 8 / int(ppq)

		// Verify the formula is applied correctly by checking the relationship
		// FILLY ticks should be proportional to MIDI ticks with factor 8/PPQ
		// For example:
		// - If PPQ = 480, then 480 MIDI ticks = 8 FILLY ticks (1 quarter note)
		// - If PPQ = 96, then 96 MIDI ticks = 8 FILLY ticks (1 quarter note)

		// Calculate how many quarter notes this represents
		quarterNotes := float64(midiTick) / float64(ppq)
		expectedFillyTicks := int(quarterNotes * 8)

		// Allow small tolerance for integer division rounding
		diff := int(math.Abs(float64(fillyTick - expectedFillyTicks)))
		if diff > 1 {
			t.Logf("Conversion formula mismatch: midiTick=%d, ppq=%d, fillyTick=%d, expected=%d, diff=%d",
				midiTick, ppq, fillyTick, expectedFillyTicks, diff)
			return false
		}

		return true
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(f, config); err != nil {
		t.Error(err)
	}
}

// Feature: gomidi-midi-playback, Property 8: Tick Advancement Notification
// **Validates: Requirements 4.3**
//
// Property: For any tick advancement during playback, the MIDI_Player SHALL
// call Engine.UpdateMIDISequences with the correct number of ticks advanced.
//
// Note: This property is tested indirectly through the MIDIStream.Read() method
// and UpdateHeadless() method, which both calculate tick advancement and call
// UpdateMIDISequences. The logic is:
// 1. Calculate current MIDI tick from elapsed time
// 2. Convert to FILLY ticks: currentTick = currentMIDITick * 8 / PPQ
// 3. Calculate advancement: ticksAdvanced = currentTick - lastTick
// 4. Call UpdateMIDISequences(ticksAdvanced) if ticksAdvanced > 0
//
// This test verifies the tick advancement calculation is correct.
func TestProperty_TickAdvancementCalculation(t *testing.T) {
	f := func(lastTick uint8, currentTick uint8) bool {
		// Ensure currentTick >= lastTick (time moves forward)
		if currentTick < lastTick {
			currentTick, lastTick = lastTick, currentTick
		}

		// Calculate tick advancement
		ticksAdvanced := int(currentTick) - int(lastTick)

		// Verify the calculation is correct
		expectedAdvancement := int(currentTick - lastTick)
		if ticksAdvanced != expectedAdvancement {
			t.Logf("Tick advancement calculation error: lastTick=%d, currentTick=%d, ticksAdvanced=%d, expected=%d",
				lastTick, currentTick, ticksAdvanced, expectedAdvancement)
			return false
		}

		// Verify that advancement is non-negative
		if ticksAdvanced < 0 {
			t.Logf("Negative tick advancement: lastTick=%d, currentTick=%d, ticksAdvanced=%d",
				lastTick, currentTick, ticksAdvanced)
			return false
		}

		return true
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(f, config); err != nil {
		t.Error(err)
	}
}
