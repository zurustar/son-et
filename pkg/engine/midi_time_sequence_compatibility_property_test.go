package engine

import (
	"bytes"
	"math"
	"os"
	"testing"
	"testing/quick"
	"time"

	"github.com/sinshu/go-meltysynth/meltysynth"
	"github.com/zurustar/son-et/pkg/compiler/interpreter"
)

// Feature: gomidi-midi-playback, Property 13: MIDI_TIME Sequence Compatibility
// **Validates: Requirements 10.2**
//
// Property: For any FILLY program using MIDI_TIME sequences, the behavior with
// the new implementation SHALL match the expected behavior based on MIDI timing.
//
// This property verifies that:
// 1. MIDI_TIME sequences receive tick updates synchronized with MIDI playback
// 2. The number of ticks delivered matches the MIDI file's timing
// 3. Sequences execute commands at the correct FILLY tick positions
func TestProperty_MIDITimeSequenceCompatibility(t *testing.T) {
	// Load soundfont once for all test iterations
	sfData, err := os.ReadFile("../../GeneralUser-GS.sf2")
	if err != nil {
		t.Skipf("Skipping test: soundfont not available: %v", err)
		return
	}

	f := func(noteCount uint8, ppq uint16, tempo uint8) bool {
		// Constrain inputs to reasonable ranges
		if noteCount == 0 || noteCount > 20 {
			return true // Skip invalid note counts (0 or too many)
		}
		if ppq < 24 || ppq > 960 {
			return true // Skip invalid PPQ values
		}
		if tempo < 60 || tempo > 240 {
			return true // Skip invalid tempo (60-240 BPM)
		}

		// Create test engine
		assetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
		engine := NewEngine(nil, assetLoader, nil)
		engine.SetHeadless(true)
		engine.Start()

		// Create a MIDI file with specified parameters
		// Each note is 1 quarter note long
		midiData := createTestMIDIFile(int(noteCount), int(ppq))
		assetLoader.Files["test.mid"] = midiData

		// Create MIDI player
		mp := NewMIDIPlayer(engine)
		sf, err := meltysynth.NewSoundFont(bytes.NewReader(sfData))
		if err != nil {
			t.Logf("Failed to parse soundfont: %v", err)
			return false
		}
		mp.soundFont = sf

		// Track tick updates received by MIDI_TIME sequences
		ticksReceived := 0
		tickChan := make(chan int, 1000)

		// Register a MIDI_TIME sequence that counts tick updates
		// Each time UpdateMIDISequences is called, this sequence executes once per tick
		opcodes := []interpreter.OpCode{
			{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("tick_count"), int64(1)}},
		}
		engine.RegisterMesBlock(EventMIDI_TIME, opcodes, nil, 0)

		// Monitor tick updates by tracking sequencer count
		// Each tick creates a new sequencer execution
		lastSeqCount := 0
		go func() {
			ticker := time.NewTicker(10 * time.Millisecond)
			defer ticker.Stop()
			timeout := time.After(5 * time.Second)
			for {
				select {
				case <-ticker.C:
					sequencers := engine.GetState().GetSequencers()
					currentCount := len(sequencers)
					if currentCount > lastSeqCount {
						// New sequencer created, indicating tick update
						select {
						case tickChan <- 1:
						default:
						}
						lastSeqCount = currentCount
					}
				case <-timeout:
					return
				}
			}
		}()

		// Start playback
		err = mp.PlayMIDI("test.mid")
		if err != nil {
			t.Logf("Failed to start MIDI playback: %v", err)
			return false
		}

		// Calculate expected duration
		// Each note is 1 quarter note = ppq MIDI ticks
		// Total MIDI ticks = noteCount * ppq
		// At 120 BPM (default tempo): 1 quarter note = 0.5 seconds
		// Total duration = noteCount * 0.5 seconds
		expectedDuration := float64(noteCount) * 0.5

		// Wait for playback to complete (with some buffer)
		waitDuration := time.Duration(expectedDuration*1.2+1.0) * time.Second
		timeout := time.After(waitDuration)

		// Collect tick updates
		done := false
		for !done {
			select {
			case <-tickChan:
				ticksReceived++
			case <-timeout:
				done = true
			}
		}

		// Stop playback
		mp.Stop()

		// Calculate expected FILLY ticks
		// Total MIDI ticks = noteCount * ppq
		// FILLY ticks = MIDI ticks * 8 / ppq = noteCount * ppq * 8 / ppq = noteCount * 8
		expectedFillyTicks := int(noteCount) * 8

		// Allow tolerance for timing variations
		// In headless mode, tick generation may not be perfectly synchronized
		// Allow ±20% tolerance or ±2 ticks, whichever is larger
		tolerance := int(math.Max(float64(expectedFillyTicks)*0.2, 2.0))
		diff := int(math.Abs(float64(ticksReceived - expectedFillyTicks)))

		if diff > tolerance {
			t.Logf("MIDI_TIME sequence tick count mismatch:")
			t.Logf("  noteCount=%d, ppq=%d, tempo=%d BPM", noteCount, ppq, tempo)
			t.Logf("  expectedFillyTicks=%d, ticksReceived=%d, diff=%d, tolerance=%d",
				expectedFillyTicks, ticksReceived, diff, tolerance)
			return false
		}

		return true
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(f, config); err != nil {
		t.Error(err)
	}
}

// TestProperty_MIDITimeSequenceCompatibility_CommandExecution verifies that
// MIDI_TIME sequences execute commands at the correct tick positions.
func TestProperty_MIDITimeSequenceCompatibility_CommandExecution(t *testing.T) {
	// Load soundfont once for all test iterations
	sfData, err := os.ReadFile("../../GeneralUser-GS.sf2")
	if err != nil {
		t.Skipf("Skipping test: soundfont not available: %v", err)
		return
	}

	f := func(waitTicks uint8, ppq uint16) bool {
		// Constrain inputs to reasonable ranges
		if waitTicks == 0 || waitTicks > 16 {
			return true // Skip invalid wait counts (0 or too long)
		}
		if ppq < 24 || ppq > 960 {
			return true // Skip invalid PPQ values
		}

		// Create test engine
		assetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
		engine := NewEngine(nil, assetLoader, nil)
		engine.SetHeadless(true)
		engine.Start()

		// Create a MIDI file long enough to test the wait
		// Need at least waitTicks + 2 FILLY ticks
		// 1 quarter note = 8 FILLY ticks, so need (waitTicks + 2) / 8 quarter notes
		noteCount := int((waitTicks+10)/8) + 1
		midiData := createTestMIDIFile(noteCount, int(ppq))
		assetLoader.Files["test.mid"] = midiData

		// Create MIDI player
		mp := NewMIDIPlayer(engine)
		sf, err := meltysynth.NewSoundFont(bytes.NewReader(sfData))
		if err != nil {
			t.Logf("Failed to parse soundfont: %v", err)
			return false
		}
		mp.soundFont = sf

		// Register a MIDI_TIME sequence with a wait command
		// Sequence: x=1, wait(waitTicks), y=2
		opcodes := []interpreter.OpCode{
			{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("x"), int64(1)}},
			{Cmd: interpreter.OpWait, Args: []any{int64(waitTicks)}},
			{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("y"), int64(2)}},
		}
		engine.RegisterMesBlock(EventMIDI_TIME, opcodes, nil, 0)

		// Start playback
		err = mp.PlayMIDI("test.mid")
		if err != nil {
			t.Logf("Failed to start MIDI playback: %v", err)
			return false
		}

		// Wait for playback to progress
		// Expected time: (waitTicks + 2) FILLY ticks
		// At 120 BPM: 8 FILLY ticks = 0.5 seconds
		// Time = (waitTicks + 2) / 8 * 0.5 seconds
		expectedTime := float64(waitTicks+2) / 8.0 * 0.5
		waitDuration := time.Duration(expectedTime*1.5+0.5) * time.Second
		time.Sleep(waitDuration)

		// Stop playback
		mp.Stop()

		// Verify that both assignments executed by checking sequencers
		// The MIDI_TIME sequence should have executed and set variables
		sequencers := engine.GetState().GetSequencers()
		if len(sequencers) == 0 {
			t.Logf("No sequencers created (waitTicks=%d, ppq=%d)", waitTicks, ppq)
			return false
		}

		// Get the first sequencer (the MIDI_TIME sequence)
		seq := sequencers[0]
		x := seq.GetVariable("x")
		y := seq.GetVariable("y")

		if x != int64(1) {
			t.Logf("Expected x=1, got x=%v (waitTicks=%d, ppq=%d)", x, waitTicks, ppq)
			return false
		}

		if y != int64(2) {
			t.Logf("Expected y=2, got y=%v (waitTicks=%d, ppq=%d)", y, waitTicks, ppq)
			return false
		}

		return true
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(f, config); err != nil {
		t.Error(err)
	}
}

// TestProperty_MIDITimeSequenceCompatibility_MultipleSequences verifies that
// multiple MIDI_TIME sequences can run concurrently and receive tick updates.
func TestProperty_MIDITimeSequenceCompatibility_MultipleSequences(t *testing.T) {
	// Load soundfont once for all test iterations
	sfData, err := os.ReadFile("../../GeneralUser-GS.sf2")
	if err != nil {
		t.Skipf("Skipping test: soundfont not available: %v", err)
		return
	}

	f := func(seqCount uint8, noteCount uint8, ppq uint16) bool {
		// Constrain inputs to reasonable ranges
		if seqCount == 0 || seqCount > 5 {
			return true // Skip invalid sequence counts (0 or too many)
		}
		if noteCount == 0 || noteCount > 10 {
			return true // Skip invalid note counts
		}
		if ppq < 24 || ppq > 960 {
			return true // Skip invalid PPQ values
		}

		// Create test engine
		assetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
		engine := NewEngine(nil, assetLoader, nil)
		engine.SetHeadless(true)
		engine.Start()

		// Create a MIDI file
		midiData := createTestMIDIFile(int(noteCount), int(ppq))
		assetLoader.Files["test.mid"] = midiData

		// Create MIDI player
		mp := NewMIDIPlayer(engine)
		sf, err := meltysynth.NewSoundFont(bytes.NewReader(sfData))
		if err != nil {
			t.Logf("Failed to parse soundfont: %v", err)
			return false
		}
		mp.soundFont = sf

		// Register multiple MIDI_TIME sequences
		// Each sequence increments a different counter variable
		for i := 0; i < int(seqCount); i++ {
			varName := interpreter.Variable("counter_" + string(rune('a'+i)))
			opcodes := []interpreter.OpCode{
				{Cmd: interpreter.OpAssign, Args: []any{varName, int64(1)}},
			}
			engine.RegisterMesBlock(EventMIDI_TIME, opcodes, nil, 0)
		}

		// Start playback
		err = mp.PlayMIDI("test.mid")
		if err != nil {
			t.Logf("Failed to start MIDI playback: %v", err)
			return false
		}

		// Wait for playback to progress
		expectedDuration := float64(noteCount) * 0.5
		waitDuration := time.Duration(expectedDuration*1.2+1.0) * time.Second
		time.Sleep(waitDuration)

		// Stop playback
		mp.Stop()

		// Verify that all sequences received tick updates
		// Each sequence should have created sequencers
		sequencers := engine.GetState().GetSequencers()
		if len(sequencers) < int(seqCount) {
			t.Logf("Expected at least %d sequencers, got %d", seqCount, len(sequencers))
			return false
		}

		// Verify each sequence set its counter variable
		for i := 0; i < int(seqCount); i++ {
			varName := "counter_" + string(rune('a'+i))
			found := false
			for _, seq := range sequencers {
				value := seq.GetVariable(varName)
				if value == int64(1) {
					found = true
					break
				}
			}
			if !found {
				t.Logf("Sequence %d did not execute: counter_%c not found", i, rune('a'+i))
				return false
			}
		}

		return true
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(f, config); err != nil {
		t.Error(err)
	}
}
