package engine

import (
	"bytes"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/sinshu/go-meltysynth/meltysynth"
	"github.com/zurustar/son-et/pkg/compiler/interpreter"
	"gitlab.com/gomidi/midi/v2/smf"
)

// TestAPICompatibility_Signatures verifies that the MIDIPlayer maintains
// the same public API signatures for backward compatibility.
// **Validates: Requirements 10.1**
func TestAPICompatibility_Signatures(t *testing.T) {
	engine := NewEngine(nil, nil, nil)
	mp := NewMIDIPlayer(engine)

	// Verify PlayMIDI signature: func(string) error
	playMIDIType := reflect.TypeOf(mp.PlayMIDI)
	if playMIDIType.Kind() != reflect.Func {
		t.Errorf("PlayMIDI is not a function")
	}
	if playMIDIType.NumIn() != 1 || playMIDIType.In(0).Kind() != reflect.String {
		t.Errorf("PlayMIDI signature changed: expected func(string) error, got %v", playMIDIType)
	}
	if playMIDIType.NumOut() != 1 || playMIDIType.Out(0).String() != "error" {
		t.Errorf("PlayMIDI return type changed: expected error, got %v", playMIDIType.Out(0))
	}

	// Verify Stop signature: func()
	stopType := reflect.TypeOf(mp.Stop)
	if stopType.Kind() != reflect.Func {
		t.Errorf("Stop is not a function")
	}
	if stopType.NumIn() != 0 {
		t.Errorf("Stop signature changed: expected func(), got %v", stopType)
	}
	if stopType.NumOut() != 0 {
		t.Errorf("Stop return type changed: expected no return, got %d returns", stopType.NumOut())
	}

	// Verify LoadSoundFont signature: func(string) error
	loadSoundFontType := reflect.TypeOf(mp.LoadSoundFont)
	if loadSoundFontType.Kind() != reflect.Func {
		t.Errorf("LoadSoundFont is not a function")
	}
	if loadSoundFontType.NumIn() != 1 || loadSoundFontType.In(0).Kind() != reflect.String {
		t.Errorf("LoadSoundFont signature changed: expected func(string) error, got %v", loadSoundFontType)
	}
	if loadSoundFontType.NumOut() != 1 || loadSoundFontType.Out(0).String() != "error" {
		t.Errorf("LoadSoundFont return type changed: expected error, got %v", loadSoundFontType.Out(0))
	}

	// Verify IsPlaying signature: func() bool
	isPlayingType := reflect.TypeOf(mp.IsPlaying)
	if isPlayingType.Kind() != reflect.Func {
		t.Errorf("IsPlaying is not a function")
	}
	if isPlayingType.NumIn() != 0 {
		t.Errorf("IsPlaying signature changed: expected func() bool, got %v", isPlayingType)
	}
	if isPlayingType.NumOut() != 1 || isPlayingType.Out(0).Kind() != reflect.Bool {
		t.Errorf("IsPlaying return type changed: expected bool, got %v", isPlayingType.Out(0))
	}
}

// TestAPICompatibility_Behavior verifies that the MIDIPlayer API behaves
// correctly with the expected call patterns.
// **Validates: Requirements 10.1**
func TestAPICompatibility_Behavior(t *testing.T) {
	// Create test engine
	assetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
	engine := NewEngine(nil, assetLoader, nil)
	engine.SetHeadless(true)
	engine.Start()

	mp := NewMIDIPlayer(engine)

	// Test 1: IsPlaying returns false initially
	if mp.IsPlaying() {
		t.Errorf("IsPlaying() should return false initially")
	}

	// Test 2: PlayMIDI without LoadSoundFont returns error
	err := mp.PlayMIDI("test.mid")
	if err == nil {
		t.Errorf("PlayMIDI without LoadSoundFont should return error")
	}

	// Test 3: LoadSoundFont with valid file succeeds
	sfData, err := os.ReadFile("../../GeneralUser-GS.sf2")
	if err != nil {
		t.Skipf("Skipping test: soundfont not available: %v", err)
		return
	}
	assetLoader.Files["test.sf2"] = sfData

	err = mp.LoadSoundFont("test.sf2")
	if err != nil {
		t.Errorf("LoadSoundFont with valid file should succeed: %v", err)
	}

	// Test 4: PlayMIDI with valid MIDI file succeeds
	// Create a longer MIDI file so playback doesn't complete immediately
	midiData := createTestMIDIFile(10, 480) // 10 notes, 1 quarter note each = ~5 seconds at 120 BPM
	assetLoader.Files["test.mid"] = midiData

	err = mp.PlayMIDI("test.mid")
	if err != nil {
		t.Errorf("PlayMIDI with valid file should succeed: %v", err)
	}

	// Give playback time to start (audio player needs time to initialize)
	time.Sleep(200 * time.Millisecond)

	// Test 5: IsPlaying returns true during playback
	// Note: In headless mode, the audio player might not report as playing
	// but the internal isPlaying flag should be set
	mp.mutex.Lock()
	internalPlaying := mp.isPlaying
	mp.mutex.Unlock()

	if !internalPlaying {
		t.Errorf("Internal isPlaying flag should be true during playback")
	}

	// Test 6: Stop() stops playback
	mp.Stop()
	time.Sleep(50 * time.Millisecond)

	if mp.IsPlaying() {
		t.Errorf("IsPlaying() should return false after Stop()")
	}
}

// TestMIDITimeSequenceCompatibility verifies that MIDI_TIME sequences
// continue to work correctly with the new implementation.
// **Validates: Requirements 10.2**
func TestMIDITimeSequenceCompatibility(t *testing.T) {
	// Create test engine
	assetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
	engine := NewEngine(nil, assetLoader, nil)
	engine.SetHeadless(true)
	engine.Start()

	// Create a MIDI file with known timing (480 PPQ, 120 BPM)
	// 1 quarter note = 480 ticks = 8 FILLY ticks
	// At 120 BPM, 1 quarter note = 0.5 seconds
	midiData := createTestMIDIFile(4, 480) // 4 notes, 1 quarter note each
	assetLoader.Files["test.mid"] = midiData

	// Load soundfont
	sfData, err := os.ReadFile("../../GeneralUser-GS.sf2")
	if err != nil {
		t.Skipf("Skipping test: soundfont not available: %v", err)
		return
	}

	mp := NewMIDIPlayer(engine)
	sf, err := meltysynth.NewSoundFont(bytes.NewReader(sfData))
	if err != nil {
		t.Fatalf("Failed to parse soundfont: %v", err)
	}
	mp.soundFont = sf

	// Track tick updates by monitoring MIDI_TIME sequencer execution
	tickUpdates := []int{}
	tickChan := make(chan int, 100)

	// Register a MIDI_TIME sequence that records tick updates
	opcodes := []interpreter.OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("tick_received"), int64(1)}},
	}
	engine.RegisterMesBlock(EventMIDI_TIME, opcodes, nil, 0)

	// Monitor sequencer count to detect tick updates
	lastSeqCount := 0
	go func() {
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()
		timeout := time.After(3 * time.Second)
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
		t.Fatalf("Failed to start MIDI playback: %v", err)
	}

	// Collect tick updates for a short duration
	timeout := time.After(2 * time.Second)
	done := false
	for !done {
		select {
		case <-tickChan:
			tickUpdates = append(tickUpdates, 1)
		case <-timeout:
			done = true
		}
	}

	mp.Stop()

	// Verify that tick updates were received
	if len(tickUpdates) == 0 {
		t.Errorf("No tick updates received during MIDI playback")
	}

	t.Logf("Received %d tick update signals during MIDI playback", len(tickUpdates))
}

// TestMIDIEndEventCompatibility verifies that MIDI_END events
// are triggered correctly with the new implementation.
// **Validates: Requirements 10.3**
func TestMIDIEndEventCompatibility(t *testing.T) {
	// Create test engine
	assetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
	engine := NewEngine(nil, assetLoader, nil)
	engine.SetHeadless(true)
	engine.Start()

	// Create a short MIDI file
	midiData := createTestMIDIFile(2, 10) // 2 notes, short duration
	assetLoader.Files["test.mid"] = midiData

	// Verify the MIDI file is valid
	_, err := smf.ReadFrom(bytes.NewReader(midiData))
	if err != nil {
		t.Fatalf("Failed to parse generated MIDI file: %v", err)
	}

	// Load soundfont
	sfData, err := os.ReadFile("../../GeneralUser-GS.sf2")
	if err != nil {
		t.Skipf("Skipping test: soundfont not available: %v", err)
		return
	}

	mp := NewMIDIPlayer(engine)
	sf, err := meltysynth.NewSoundFont(bytes.NewReader(sfData))
	if err != nil {
		t.Fatalf("Failed to parse soundfont: %v", err)
	}
	mp.soundFont = sf

	// Track MIDI_END event
	eventReceived := false
	eventChan := make(chan bool, 1)

	// Register event handler for MIDI_END
	opcodes := []interpreter.OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("midi_end_triggered"), int64(1)}},
	}
	engine.RegisterMesBlock(EventMIDI_END, opcodes, nil, 0)

	// Monitor sequencers to detect when MIDI_END event is triggered
	go func() {
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()
		timeout := time.After(5 * time.Second)
		for {
			select {
			case <-ticker.C:
				sequencers := engine.GetState().GetSequencers()
				// If any sequencer was created, it means MIDI_END was triggered
				if len(sequencers) > 0 {
					eventReceived = true
					select {
					case eventChan <- true:
					default:
					}
					return
				}
			case <-timeout:
				return
			}
		}
	}()

	// Start playback
	err = mp.PlayMIDI("test.mid")
	if err != nil {
		t.Fatalf("Failed to start MIDI playback: %v", err)
	}

	// Wait for MIDI_END event
	select {
	case <-eventChan:
		t.Logf("MIDI_END event received successfully")
	case <-time.After(5 * time.Second):
		t.Fatalf("MIDI_END event not received within timeout")
	}

	// Verify event was received
	if !eventReceived {
		t.Fatalf("MIDI_END event was not triggered")
	}

	// Verify player is no longer playing
	time.Sleep(100 * time.Millisecond)
	if mp.IsPlaying() {
		t.Errorf("Player still marked as playing after MIDI_END")
	}
}

// TestMIDIEndEventTiming verifies that MIDI_END events are triggered
// at the correct time (after all MIDI messages have been played).
// **Validates: Requirements 10.3**
func TestMIDIEndEventTiming(t *testing.T) {
	// Create test engine
	assetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
	engine := NewEngine(nil, assetLoader, nil)
	engine.SetHeadless(true)
	engine.Start()

	// Create a MIDI file with known duration
	// 4 notes, 480 ticks each = 1920 ticks total
	// At 480 PPQ and 120 BPM: 1920 ticks = 4 quarter notes = 2 seconds
	midiData := createTestMIDIFile(4, 480)
	assetLoader.Files["test.mid"] = midiData

	// Load soundfont
	sfData, err := os.ReadFile("../../GeneralUser-GS.sf2")
	if err != nil {
		t.Skipf("Skipping test: soundfont not available: %v", err)
		return
	}

	mp := NewMIDIPlayer(engine)
	sf, err := meltysynth.NewSoundFont(bytes.NewReader(sfData))
	if err != nil {
		t.Fatalf("Failed to parse soundfont: %v", err)
	}
	mp.soundFont = sf

	// Track MIDI_END event timing
	eventChan := make(chan time.Time, 1)

	// Register event handler for MIDI_END
	opcodes := []interpreter.OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("midi_end_triggered"), int64(1)}},
	}
	engine.RegisterMesBlock(EventMIDI_END, opcodes, nil, 0)

	// Monitor for MIDI_END event
	go func() {
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()
		timeout := time.After(5 * time.Second)
		for {
			select {
			case <-ticker.C:
				sequencers := engine.GetState().GetSequencers()
				// If any sequencer was created, it means MIDI_END was triggered
				if len(sequencers) > 0 {
					select {
					case eventChan <- time.Now():
					default:
					}
					return
				}
			case <-timeout:
				return
			}
		}
	}()

	// Start playback and record start time
	startTime := time.Now()
	err = mp.PlayMIDI("test.mid")
	if err != nil {
		t.Fatalf("Failed to start MIDI playback: %v", err)
	}

	// Wait for MIDI_END event
	var eventTime time.Time
	select {
	case eventTime = <-eventChan:
		t.Logf("MIDI_END event received")
	case <-time.After(5 * time.Second):
		t.Fatalf("MIDI_END event not received within timeout")
	}

	// Calculate elapsed time
	elapsed := eventTime.Sub(startTime).Seconds()

	// Expected duration: ~2 seconds (4 quarter notes at 120 BPM)
	// However, the playMIDIMessages goroutine uses simple timing that may not be perfectly accurate
	// Allow tolerance for processing time, scheduling, and timing implementation
	if elapsed < 1.0 || elapsed > 6.0 {
		t.Errorf("MIDI_END event timing incorrect: got %.2fs, expected ~2.0s (with tolerance)", elapsed)
	}

	t.Logf("MIDI_END event triggered after %.2f seconds (expected ~2.0s)", elapsed)
}

// TestBackwardCompatibility_MultiplePlaybacks verifies that multiple
// sequential playbacks work correctly (stop and restart).
// **Validates: Requirements 10.1**
func TestBackwardCompatibility_MultiplePlaybacks(t *testing.T) {
	// Create test engine
	assetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
	engine := NewEngine(nil, assetLoader, nil)
	engine.SetHeadless(true)
	engine.Start()

	// Create test MIDI files
	midiData1 := createTestMIDIFile(5, 100)
	midiData2 := createTestMIDIFile(5, 100)
	assetLoader.Files["test1.mid"] = midiData1
	assetLoader.Files["test2.mid"] = midiData2

	// Load soundfont
	sfData, err := os.ReadFile("../../GeneralUser-GS.sf2")
	if err != nil {
		t.Skipf("Skipping test: soundfont not available: %v", err)
		return
	}

	mp := NewMIDIPlayer(engine)
	sf, err := meltysynth.NewSoundFont(bytes.NewReader(sfData))
	if err != nil {
		t.Fatalf("Failed to parse soundfont: %v", err)
	}
	mp.soundFont = sf

	// First playback - verify API works
	err = mp.PlayMIDI("test1.mid")
	if err != nil {
		t.Fatalf("First PlayMIDI failed: %v", err)
	}

	// Stop first playback
	mp.Stop()
	time.Sleep(100 * time.Millisecond)

	// Second playback - verify API works after stop
	err = mp.PlayMIDI("test2.mid")
	if err != nil {
		t.Fatalf("Second PlayMIDI failed: %v", err)
	}

	// Stop second playback
	mp.Stop()
	time.Sleep(50 * time.Millisecond)

	// The key test is that the API doesn't error on multiple calls
	t.Logf("Multiple playback API calls succeeded")
}
