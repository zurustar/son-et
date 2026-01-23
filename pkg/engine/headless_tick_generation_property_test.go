package engine

import (
	"testing"
	"time"

	"github.com/zurustar/son-et/pkg/compiler/interpreter"
)

// Feature: gomidi-midi-playback, Property 9: Headless Mode Tick Generation
// **Validates: Requirements 6.2**
// For any MIDI playback in headless mode, the MIDI_Player SHALL continue generating
// FILLY ticks at the correct rate.
//
// This test verifies that UpdateHeadless() correctly calculates and delivers ticks
// based on wall-clock time in headless mode.
func TestHeadlessModeTickGeneration(t *testing.T) {
	// Create test engine in headless mode
	assetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
	engine := NewEngine(nil, assetLoader, nil)
	engine.SetHeadless(true)
	engine.Start() // Initialize context

	// Register a MIDI_TIME sequence to track tick updates
	// Add multiple instructions with waits to ensure the sequence continues
	opcodes := []interpreter.OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("tick1"), int64(1)}},
		{Cmd: interpreter.OpWait, Args: []any{int64(2)}}, // Wait 2 steps
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("tick2"), int64(2)}},
		{Cmd: interpreter.OpWait, Args: []any{int64(2)}}, // Wait 2 steps
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("tick3"), int64(3)}},
	}
	engine.RegisterMesBlock(EventMIDI_TIME, opcodes, nil, 0)

	// Create MIDI player
	mp := NewMIDIPlayer(engine)

	// Manually set up the stream and tick generator to simulate playback
	// This allows us to test UpdateHeadless() without actual MIDI playback
	tempoMap := []TempoEvent{{Tick: 0, MicrosPerBeat: 500000}} // 120 BPM
	ppq := 480
	tickGen := NewWallClockTickGenerator(MIDISampleRate, ppq, tempoMap)

	mp.stream = &MIDIStream{
		synthesizer:   nil, // Not needed for headless tick generation
		tickGenerator: tickGen,
		startTime:     time.Now(),
		engine:        engine,
		lastTick:      -1,
	}
	mp.isPlaying = true

	// Get initial sequencer state
	sequencers := engine.GetState().GetSequencers()
	if len(sequencers) != 1 {
		t.Fatalf("Expected 1 sequencer, got %d", len(sequencers))
	}
	midiSeq := sequencers[0]
	initialPC := midiSeq.GetPC()

	// Wait for enough time to pass for ticks to advance
	// At 120 BPM: 1 beat = 0.5s, 1 beat = 8 FILLY ticks
	// So 0.2s = 0.4 beats = 3.2 FILLY ticks
	time.Sleep(200 * time.Millisecond)

	// Call UpdateHeadless to generate ticks
	mp.UpdateHeadless()

	// Verify that ticks were generated and the sequence advanced
	finalPC := midiSeq.GetPC()
	if finalPC <= initialPC {
		t.Errorf("Expected sequence to advance after UpdateHeadless, but PC did not change (initial=%d, final=%d)",
			initialPC, finalPC)
	}

	// Verify that at least 1 tick was generated
	// (PC should have advanced by at least 1)
	if finalPC < 1 {
		t.Errorf("Expected at least 1 tick to be generated, but PC is %d", finalPC)
	}

	// Call UpdateHeadless again after more time
	time.Sleep(200 * time.Millisecond)
	mp.UpdateHeadless()

	// Verify that more ticks were generated
	newPC := midiSeq.GetPC()
	if newPC <= finalPC {
		t.Errorf("Expected sequence to continue advancing, but PC did not change (previous=%d, new=%d)",
			finalPC, newPC)
	}

	// Stop playback
	mp.Stop()
	if mp.IsPlaying() {
		t.Errorf("Expected player to stop after Stop()")
	}
}
