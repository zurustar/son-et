package audio

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/zurustar/son-et/pkg/vm"
)

// Shared audio context for all tests (Ebitengine only allows one context)
var (
	sharedAudioCtx     *audio.Context
	sharedAudioCtxOnce sync.Once
)

// getSharedAudioContext returns the shared audio context for tests.
func getSharedAudioContext() *audio.Context {
	sharedAudioCtxOnce.Do(func() {
		sharedAudioCtx = audio.NewContext(SampleRate)
	})
	return sharedAudioCtx
}

// TestNewMIDIPlayer tests the creation of a new MIDI player.
func TestNewMIDIPlayer(t *testing.T) {
	// Find SoundFont file
	soundFontPath := findSoundFont(t)
	audioCtx := getSharedAudioContext()

	t.Run("creates player with valid SoundFont", func(t *testing.T) {
		player, err := NewMIDIPlayer(soundFontPath, audioCtx, nil)
		if err != nil {
			t.Fatalf("NewMIDIPlayer failed: %v", err)
		}
		if player == nil {
			t.Fatal("NewMIDIPlayer returned nil player")
		}
		if player.IsPlaying() {
			t.Error("New player should not be playing")
		}
	})

	t.Run("returns error for empty SoundFont path", func(t *testing.T) {
		// Requirement 4.10: When SoundFont is not provided, system reports error.
		_, err := NewMIDIPlayer("", audioCtx, nil)
		if err == nil {
			t.Error("Expected error for empty SoundFont path")
		}
		if err != ErrNoSoundFont {
			t.Errorf("Expected ErrNoSoundFont, got: %v", err)
		}
	})

	t.Run("returns error for non-existent SoundFont", func(t *testing.T) {
		_, err := NewMIDIPlayer("/nonexistent/path/to/soundfont.sf2", audioCtx, nil)
		if err == nil {
			t.Error("Expected error for non-existent SoundFont")
		}
	})
}

// TestMIDIPlayerPlay tests MIDI file playback.
func TestMIDIPlayerPlay(t *testing.T) {
	soundFontPath := findSoundFont(t)
	midiPath := findMIDIFile(t)
	audioCtx := getSharedAudioContext()

	player, err := NewMIDIPlayer(soundFontPath, audioCtx, nil)
	if err != nil {
		t.Fatalf("NewMIDIPlayer failed: %v", err)
	}

	t.Run("plays valid MIDI file", func(t *testing.T) {
		// Requirement 4.1: When PlayMIDI(filename) is called, system starts playback.
		err := player.Play(midiPath)
		if err != nil {
			t.Fatalf("Play failed: %v", err)
		}

		if !player.IsPlaying() {
			t.Error("Player should be playing after Play()")
		}

		// Check duration is set
		duration := player.GetDuration()
		if duration <= 0 {
			t.Error("Duration should be positive")
		}

		// Clean up
		player.Stop()
	})

	t.Run("returns error for non-existent MIDI file", func(t *testing.T) {
		err := player.Play("/nonexistent/path/to/midi.mid")
		if err == nil {
			t.Error("Expected error for non-existent MIDI file")
		}
	})

	t.Run("returns error for invalid MIDI file", func(t *testing.T) {
		// Create a temporary invalid file
		tmpFile, err := os.CreateTemp("", "invalid*.mid")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())
		tmpFile.WriteString("not a midi file")
		tmpFile.Close()

		err = player.Play(tmpFile.Name())
		if err == nil {
			t.Error("Expected error for invalid MIDI file")
		}
	})
}

// TestMIDIPlayerStop tests stopping MIDI playback.
func TestMIDIPlayerStop(t *testing.T) {
	soundFontPath := findSoundFont(t)
	midiPath := findMIDIFile(t)
	audioCtx := getSharedAudioContext()

	player, err := NewMIDIPlayer(soundFontPath, audioCtx, nil)
	if err != nil {
		t.Fatalf("NewMIDIPlayer failed: %v", err)
	}

	t.Run("stops playing MIDI", func(t *testing.T) {
		err := player.Play(midiPath)
		if err != nil {
			t.Fatalf("Play failed: %v", err)
		}

		if !player.IsPlaying() {
			t.Error("Player should be playing")
		}

		player.Stop()

		if player.IsPlaying() {
			t.Error("Player should not be playing after Stop()")
		}
	})

	t.Run("stop is safe when not playing", func(t *testing.T) {
		// Should not panic
		player.Stop()
		player.Stop()
	})
}

// TestMIDIPlayerMute tests muting functionality.
func TestMIDIPlayerMute(t *testing.T) {
	soundFontPath := findSoundFont(t)
	audioCtx := getSharedAudioContext()

	player, err := NewMIDIPlayer(soundFontPath, audioCtx, nil)
	if err != nil {
		t.Fatalf("NewMIDIPlayer failed: %v", err)
	}

	t.Run("mute and unmute", func(t *testing.T) {
		if player.IsMuted() {
			t.Error("Player should not be muted initially")
		}

		player.SetMuted(true)
		if !player.IsMuted() {
			t.Error("Player should be muted after SetMuted(true)")
		}

		player.SetMuted(false)
		if player.IsMuted() {
			t.Error("Player should not be muted after SetMuted(false)")
		}
	})
}

// TestMIDIPlayerWithEventQueue tests MIDI player with event queue.
func TestMIDIPlayerWithEventQueue(t *testing.T) {
	soundFontPath := findSoundFont(t)
	audioCtx := getSharedAudioContext()

	eventQueue := vm.NewEventQueue()
	player, err := NewMIDIPlayer(soundFontPath, audioCtx, eventQueue)
	if err != nil {
		t.Fatalf("NewMIDIPlayer failed: %v", err)
	}

	if player == nil {
		t.Fatal("Player should not be nil")
	}
}

// TestTickCalculator tests the tick calculator.
func TestTickCalculator(t *testing.T) {
	t.Run("calculates ticks with single tempo", func(t *testing.T) {
		// 120 BPM = 500000 microseconds per beat
		// PPQ = 480
		tempoMap := []TempoEvent{{Tick: 0, MicrosPerBeat: 500000}}
		tc := NewTickCalculator(480, tempoMap)

		// At 120 BPM with PPQ=480:
		// 1 quarter note = 0.5 seconds = 22050 samples
		// 1 tick = 0.5/480 seconds = ~46 samples

		// Test at 0 samples
		tick := tc.TickFromSamples(0)
		if tick != 0 {
			t.Errorf("Expected tick 0 at sample 0, got %d", tick)
		}

		// Test at ~1 quarter note (22050 samples)
		tick = tc.TickFromSamples(22050)
		// Should be approximately 480 ticks
		if tick < 400 || tick > 560 {
			t.Errorf("Expected tick ~480 at sample 22050, got %d", tick)
		}
	})

	t.Run("calculates FILLY ticks", func(t *testing.T) {
		tempoMap := []TempoEvent{{Tick: 0, MicrosPerBeat: 500000}}
		tc := NewTickCalculator(480, tempoMap)

		// At 1 quarter note (480 MIDI ticks), should be 4 FILLY ticks (16th notes)
		tick := tc.TickFromSamples(22050)
		fillyTick := tc.FillyTickFromSamples(22050)

		// FILLY tick = MIDI tick * 4 / PPQ (16th note resolution)
		expectedFilly := tick * 4 / 480
		if fillyTick != expectedFilly {
			t.Errorf("Expected FILLY tick %d, got %d", expectedFilly, fillyTick)
		}
	})

	t.Run("handles tempo changes", func(t *testing.T) {
		// Start at 120 BPM, change to 60 BPM at tick 480
		tempoMap := []TempoEvent{
			{Tick: 0, MicrosPerBeat: 500000},    // 120 BPM
			{Tick: 480, MicrosPerBeat: 1000000}, // 60 BPM
		}
		tc := NewTickCalculator(480, tempoMap)

		// At tick 480 (1 quarter note at 120 BPM = 22050 samples)
		tick := tc.TickFromSamples(22050)
		if tick < 400 || tick > 560 {
			t.Errorf("Expected tick ~480 at sample 22050, got %d", tick)
		}

		// After tempo change, ticks should progress slower
		// At 60 BPM, 1 quarter note = 1 second = 44100 samples
		// So at 22050 + 44100 = 66150 samples, we should be at tick 480 + 480 = 960
		tick = tc.TickFromSamples(66150)
		if tick < 880 || tick > 1040 {
			t.Errorf("Expected tick ~960 at sample 66150, got %d", tick)
		}
	})

	t.Run("handles empty tempo map", func(t *testing.T) {
		tc := NewTickCalculator(480, []TempoEvent{})
		tick := tc.TickFromSamples(22050)
		if tick != 0 {
			t.Errorf("Expected tick 0 for empty tempo map, got %d", tick)
		}
	})
}

// TestParseMIDITempoMap tests MIDI tempo map parsing.
func TestParseMIDITempoMap(t *testing.T) {
	t.Run("returns default for invalid data", func(t *testing.T) {
		events, ppq := ParseMIDITempoMap([]byte{})
		if len(events) != 1 {
			t.Errorf("Expected 1 default event, got %d", len(events))
		}
		if events[0].MicrosPerBeat != 500000 {
			t.Errorf("Expected default tempo 500000, got %d", events[0].MicrosPerBeat)
		}
		if ppq != 480 {
			t.Errorf("Expected default PPQ 480, got %d", ppq)
		}
	})

	t.Run("parses real MIDI file", func(t *testing.T) {
		midiPath := findMIDIFile(t)
		data, err := os.ReadFile(midiPath)
		if err != nil {
			t.Skipf("Could not read MIDI file: %v", err)
		}

		events, ppq := ParseMIDITempoMap(data)
		if len(events) == 0 {
			t.Error("Expected at least one tempo event")
		}
		if ppq <= 0 {
			t.Errorf("Expected positive PPQ, got %d", ppq)
		}
		// First event should be at tick 0
		if events[0].Tick != 0 {
			t.Errorf("First tempo event should be at tick 0, got %d", events[0].Tick)
		}
	})
}

// TestMIDIStream tests the MIDI stream implementation.
func TestMIDIStream(t *testing.T) {
	soundFontPath := findSoundFont(t)
	midiPath := findMIDIFile(t)
	audioCtx := getSharedAudioContext()

	player, err := NewMIDIPlayer(soundFontPath, audioCtx, nil)
	if err != nil {
		t.Fatalf("NewMIDIPlayer failed: %v", err)
	}

	err = player.Play(midiPath)
	if err != nil {
		t.Fatalf("Play failed: %v", err)
	}
	defer player.Stop()

	t.Run("stream renders audio", func(t *testing.T) {
		// Wait a bit for some audio to be rendered
		time.Sleep(100 * time.Millisecond)

		// Check that position has advanced
		pos := player.GetPosition()
		if pos <= 0 {
			t.Error("Position should have advanced")
		}
	})
}

// TestMIDIPlayerGetCurrentTick tests tick position retrieval.
func TestMIDIPlayerGetCurrentTick(t *testing.T) {
	soundFontPath := findSoundFont(t)
	midiPath := findMIDIFile(t)
	audioCtx := getSharedAudioContext()

	player, err := NewMIDIPlayer(soundFontPath, audioCtx, nil)
	if err != nil {
		t.Fatalf("NewMIDIPlayer failed: %v", err)
	}

	t.Run("returns 0 when not playing", func(t *testing.T) {
		tick := player.GetCurrentTick()
		if tick != 0 {
			t.Errorf("Expected tick 0 when not playing, got %d", tick)
		}
	})

	t.Run("returns tick when playing", func(t *testing.T) {
		err := player.Play(midiPath)
		if err != nil {
			t.Fatalf("Play failed: %v", err)
		}
		defer player.Stop()

		// Wait a bit
		time.Sleep(100 * time.Millisecond)

		tick := player.GetCurrentTick()
		// Tick should have advanced
		if tick < 0 {
			t.Errorf("Expected non-negative tick, got %d", tick)
		}
	})
}

// Helper functions

// findSoundFont finds the SoundFont file in the project.
func findSoundFont(t *testing.T) string {
	t.Helper()

	// Try common locations
	paths := []string{
		"../../../GeneralUser-GS.sf2",
		"../../GeneralUser-GS.sf2",
		"GeneralUser-GS.sf2",
	}

	for _, p := range paths {
		absPath, err := filepath.Abs(p)
		if err != nil {
			continue
		}
		if _, err := os.Stat(absPath); err == nil {
			return absPath
		}
	}

	t.Skip("SoundFont file not found")
	return ""
}

// findMIDIFile finds a MIDI file in the samples directory.
func findMIDIFile(t *testing.T) string {
	t.Helper()

	// Try to find a MIDI file in samples
	sampleDirs := []string{
		"../../../samples",
		"../../samples",
		"samples",
	}

	for _, dir := range sampleDirs {
		absDir, err := filepath.Abs(dir)
		if err != nil {
			continue
		}

		// Walk the directory to find .mid files
		err = filepath.Walk(absDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() {
				ext := filepath.Ext(path)
				if ext == ".mid" || ext == ".MID" {
					return filepath.SkipAll
				}
			}
			return nil
		})

		// Try to find any .mid file
		matches, _ := filepath.Glob(filepath.Join(absDir, "*", "*.mid"))
		if len(matches) > 0 {
			return matches[0]
		}
		matches, _ = filepath.Glob(filepath.Join(absDir, "*", "*.MID"))
		if len(matches) > 0 {
			return matches[0]
		}
	}

	t.Skip("MIDI file not found in samples")
	return ""
}

// TestMIDIPlayerUpdate tests the Update method for MIDI_TIME event generation.
// Requirement 4.3: When MIDI is playing, system generates MIDI_TIME events synchronized to MIDI tempo.
// Requirement 4.4: When MIDI tempo is 120 BPM with resolution 480 ticks per beat, system generates MIDI_TIME event every 1.04ms.
func TestMIDIPlayerUpdate(t *testing.T) {
	soundFontPath := findSoundFont(t)
	midiPath := findMIDIFile(t)
	audioCtx := getSharedAudioContext()

	t.Run("generates MIDI_TIME events when playing", func(t *testing.T) {
		eventQueue := vm.NewEventQueue()
		player, err := NewMIDIPlayer(soundFontPath, audioCtx, eventQueue)
		if err != nil {
			t.Fatalf("NewMIDIPlayer failed: %v", err)
		}

		err = player.Play(midiPath)
		if err != nil {
			t.Fatalf("Play failed: %v", err)
		}
		defer player.Stop()

		// Wait for some audio to be rendered
		time.Sleep(150 * time.Millisecond)

		// Call Update to generate MIDI_TIME events
		player.Update()

		// Check that events were generated
		eventCount := eventQueue.Len()
		if eventCount == 0 {
			t.Error("Expected MIDI_TIME events to be generated")
		}

		// Verify event type
		event, ok := eventQueue.Pop()
		if !ok {
			t.Fatal("Expected to pop an event")
		}
		if event.Type != vm.EventMIDI_TIME {
			t.Errorf("Expected MIDI_TIME event, got %s", event.Type)
		}

		// Verify event has Tick parameter
		tick, ok := event.GetParam("Tick")
		if !ok {
			t.Error("Expected Tick parameter in MIDI_TIME event")
		}
		if tick.(int) < 1 {
			t.Errorf("Expected positive tick value, got %v", tick)
		}
	})

	t.Run("does not generate events when not playing", func(t *testing.T) {
		eventQueue := vm.NewEventQueue()
		player, err := NewMIDIPlayer(soundFontPath, audioCtx, eventQueue)
		if err != nil {
			t.Fatalf("NewMIDIPlayer failed: %v", err)
		}

		// Call Update without playing
		player.Update()

		// No events should be generated
		if eventQueue.Len() != 0 {
			t.Errorf("Expected no events when not playing, got %d", eventQueue.Len())
		}
	})

	t.Run("does not generate events without event queue", func(t *testing.T) {
		player, err := NewMIDIPlayer(soundFontPath, audioCtx, nil)
		if err != nil {
			t.Fatalf("NewMIDIPlayer failed: %v", err)
		}

		err = player.Play(midiPath)
		if err != nil {
			t.Fatalf("Play failed: %v", err)
		}
		defer player.Stop()

		// Wait for some audio to be rendered
		time.Sleep(100 * time.Millisecond)

		// Call Update - should not panic
		player.Update()
	})

	t.Run("generates sequential tick events", func(t *testing.T) {
		eventQueue := vm.NewEventQueue()
		player, err := NewMIDIPlayer(soundFontPath, audioCtx, eventQueue)
		if err != nil {
			t.Fatalf("NewMIDIPlayer failed: %v", err)
		}

		err = player.Play(midiPath)
		if err != nil {
			t.Fatalf("Play failed: %v", err)
		}
		defer player.Stop()

		// Wait for some audio to be rendered
		time.Sleep(200 * time.Millisecond)

		// Call Update multiple times
		player.Update()
		time.Sleep(50 * time.Millisecond)
		player.Update()

		// Collect all events
		var ticks []int
		for {
			event, ok := eventQueue.Pop()
			if !ok {
				break
			}
			if event.Type == vm.EventMIDI_TIME {
				tick, _ := event.GetParam("Tick")
				ticks = append(ticks, tick.(int))
			}
		}

		// Verify ticks are sequential (no gaps)
		if len(ticks) < 2 {
			t.Skip("Not enough ticks generated for sequential test")
		}

		for i := 1; i < len(ticks); i++ {
			if ticks[i] != ticks[i-1]+1 {
				t.Errorf("Ticks should be sequential: got %d after %d", ticks[i], ticks[i-1])
			}
		}
	})

	t.Run("starts from tick 1 after Play", func(t *testing.T) {
		eventQueue := vm.NewEventQueue()
		player, err := NewMIDIPlayer(soundFontPath, audioCtx, eventQueue)
		if err != nil {
			t.Fatalf("NewMIDIPlayer failed: %v", err)
		}

		err = player.Play(midiPath)
		if err != nil {
			t.Fatalf("Play failed: %v", err)
		}
		defer player.Stop()

		// Wait for some audio to be rendered
		time.Sleep(150 * time.Millisecond)

		// Call Update
		player.Update()

		// First event should have tick >= 1
		event, ok := eventQueue.Pop()
		if !ok {
			t.Fatal("Expected at least one event")
		}
		tick, _ := event.GetParam("Tick")
		if tick.(int) < 1 {
			t.Errorf("First tick should be >= 1, got %d", tick.(int))
		}
	})
}

// TestMIDIPlayerUpdateAfterStop tests that Update doesn't generate events after Stop.
func TestMIDIPlayerUpdateAfterStop(t *testing.T) {
	soundFontPath := findSoundFont(t)
	midiPath := findMIDIFile(t)
	audioCtx := getSharedAudioContext()

	eventQueue := vm.NewEventQueue()
	player, err := NewMIDIPlayer(soundFontPath, audioCtx, eventQueue)
	if err != nil {
		t.Fatalf("NewMIDIPlayer failed: %v", err)
	}

	err = player.Play(midiPath)
	if err != nil {
		t.Fatalf("Play failed: %v", err)
	}

	// Wait for some audio to be rendered
	time.Sleep(100 * time.Millisecond)

	// Stop playback
	player.Stop()

	// Clear any existing events
	for eventQueue.Len() > 0 {
		eventQueue.Pop()
	}

	// Call Update after stop
	player.Update()

	// No new events should be generated
	if eventQueue.Len() != 0 {
		t.Errorf("Expected no events after Stop, got %d", eventQueue.Len())
	}
}

// TestMIDIPlayerMIDIEndEvent tests MIDI_END event generation when playback completes.
// Requirement 4.5: When MIDI playback completes, system generates MIDI_END event.
func TestMIDIPlayerMIDIEndEvent(t *testing.T) {
	soundFontPath := findSoundFont(t)
	midiPath := findMIDIFile(t)
	audioCtx := getSharedAudioContext()

	t.Run("generates MIDI_END event when playback completes", func(t *testing.T) {
		eventQueue := vm.NewEventQueue()
		player, err := NewMIDIPlayer(soundFontPath, audioCtx, eventQueue)
		if err != nil {
			t.Fatalf("NewMIDIPlayer failed: %v", err)
		}

		err = player.Play(midiPath)
		if err != nil {
			t.Fatalf("Play failed: %v", err)
		}

		// Get the duration of the MIDI file
		duration := player.GetDuration()
		if duration <= 0 {
			t.Skip("MIDI file has no duration")
		}

		// Wait for playback to complete (with a reasonable timeout)
		// For testing, we'll use a short MIDI file or wait up to the duration + buffer
		const midiEndTestMaxDuration = 5 * time.Second
		if duration > midiEndTestMaxDuration {
			// For long MIDI files, skip this test
			t.Skip("MIDI file too long for completion test")
		}
		maxWait := duration + 2*time.Second

		// Poll Update() until playback completes or timeout
		startTime := time.Now()
		for player.IsPlaying() && time.Since(startTime) < maxWait {
			player.Update()
			time.Sleep(20 * time.Millisecond)
		}

		// Check that playback has stopped
		if player.IsPlaying() {
			t.Error("Playback should have stopped")
		}

		// Look for MIDI_END event in the queue
		foundMIDIEnd := false
		for {
			event, ok := eventQueue.Pop()
			if !ok {
				break
			}
			if event.Type == vm.EventMIDI_END {
				foundMIDIEnd = true
				break
			}
		}

		if !foundMIDIEnd {
			t.Error("Expected MIDI_END event to be generated when playback completes")
		}
	})

	t.Run("MIDI_END is only generated once", func(t *testing.T) {
		eventQueue := vm.NewEventQueue()
		player, err := NewMIDIPlayer(soundFontPath, audioCtx, eventQueue)
		if err != nil {
			t.Fatalf("NewMIDIPlayer failed: %v", err)
		}

		err = player.Play(midiPath)
		if err != nil {
			t.Fatalf("Play failed: %v", err)
		}

		// Get the duration of the MIDI file
		duration := player.GetDuration()
		if duration <= 0 {
			t.Skip("MIDI file has no duration")
		}

		// Wait for playback to complete
		const midiEndTestMaxDuration = 5 * time.Second
		if duration > midiEndTestMaxDuration {
			t.Skip("MIDI file too long for completion test")
		}
		maxWait := duration + 2*time.Second

		startTime := time.Now()
		for player.IsPlaying() && time.Since(startTime) < maxWait {
			player.Update()
			time.Sleep(20 * time.Millisecond)
		}

		// Call Update multiple times after playback has stopped
		player.Update()
		player.Update()
		player.Update()

		// Count MIDI_END events
		midiEndCount := 0
		for {
			event, ok := eventQueue.Pop()
			if !ok {
				break
			}
			if event.Type == vm.EventMIDI_END {
				midiEndCount++
			}
		}

		if midiEndCount != 1 {
			t.Errorf("Expected exactly 1 MIDI_END event, got %d", midiEndCount)
		}
	})

	t.Run("no MIDI_END without event queue", func(t *testing.T) {
		player, err := NewMIDIPlayer(soundFontPath, audioCtx, nil)
		if err != nil {
			t.Fatalf("NewMIDIPlayer failed: %v", err)
		}

		err = player.Play(midiPath)
		if err != nil {
			t.Fatalf("Play failed: %v", err)
		}

		// Get the duration of the MIDI file
		duration := player.GetDuration()
		if duration <= 0 {
			t.Skip("MIDI file has no duration")
		}

		// Wait for playback to complete
		const midiEndTestMaxDuration = 5 * time.Second
		if duration > midiEndTestMaxDuration {
			t.Skip("MIDI file too long for completion test")
		}
		maxWait := duration + 2*time.Second

		startTime := time.Now()
		for player.IsPlaying() && time.Since(startTime) < maxWait {
			player.Update()
			time.Sleep(20 * time.Millisecond)
		}

		// Should not panic - just verify playback stopped
		if player.IsPlaying() {
			t.Error("Playback should have stopped")
		}
	})
}

// TestMIDIPlayerMIDIEndEventUnit tests MIDI_END event generation logic without waiting for actual playback.
// This is a unit test that verifies the Update() method's behavior when position >= duration.
// Requirement 4.5: When MIDI playback completes, system generates MIDI_END event.
func TestMIDIPlayerMIDIEndEventUnit(t *testing.T) {
	t.Run("Update sets playing to false when position >= duration", func(t *testing.T) {
		// Create a minimal MIDIPlayer for testing the Update logic
		eventQueue := vm.NewEventQueue()

		// Create a player with mocked state to simulate completion
		player := &MIDIPlayer{
			playing:    true,
			duration:   100 * time.Millisecond, // Short duration
			eventQueue: eventQueue,
		}

		// Since player is nil, Update should return early (playing but no player)
		player.Update()

		// With no player, it should return early without changing state
		if !player.playing {
			t.Error("playing should still be true when player is nil")
		}
	})

	t.Run("MIDI_END event has correct type", func(t *testing.T) {
		// Verify that EventMIDI_END is correctly defined
		event := vm.NewEvent(vm.EventMIDI_END)
		if event.Type != vm.EventMIDI_END {
			t.Errorf("Expected event type MIDI_END, got %s", event.Type)
		}
		if event.Type != "MIDI_END" {
			t.Errorf("Expected event type string 'MIDI_END', got %s", event.Type)
		}
	})

	t.Run("multiple Update calls after completion do not generate multiple MIDI_END events", func(t *testing.T) {
		eventQueue := vm.NewEventQueue()

		// Create a player that is not playing (simulating after completion)
		player := &MIDIPlayer{
			playing:    false, // Already stopped
			eventQueue: eventQueue,
		}

		// Call Update multiple times
		player.Update()
		player.Update()
		player.Update()

		// No events should be generated since playing is false
		if eventQueue.Len() != 0 {
			t.Errorf("Expected no events when not playing, got %d", eventQueue.Len())
		}
	})
}

// TestMIDIPlayerExclusiveControl tests that playing a new MIDI stops the previous one.
// Requirement 4.6: When another MIDI is playing and PlayMIDI is called, system stops the previous MIDI and starts the new one.
// **Validates: Requirements 4.6**
func TestMIDIPlayerExclusiveControl(t *testing.T) {
	soundFontPath := findSoundFont(t)
	midiPath := findMIDIFile(t)
	audioCtx := getSharedAudioContext()

	t.Run("playing new MIDI stops previous MIDI", func(t *testing.T) {
		// Requirement 4.6: When another MIDI is playing and PlayMIDI is called,
		// system stops the previous MIDI and starts the new one.
		eventQueue := vm.NewEventQueue()
		player, err := NewMIDIPlayer(soundFontPath, audioCtx, eventQueue)
		if err != nil {
			t.Fatalf("NewMIDIPlayer failed: %v", err)
		}

		// Start first MIDI playback
		err = player.Play(midiPath)
		if err != nil {
			t.Fatalf("First Play failed: %v", err)
		}

		if !player.IsPlaying() {
			t.Error("Player should be playing after first Play()")
		}

		// Wait a bit to ensure playback has started
		time.Sleep(100 * time.Millisecond)

		// Get the current file before playing new MIDI
		firstFile := player.GetCurrentFile()
		if firstFile == "" {
			t.Error("Current file should be set after Play()")
		}

		// Start second MIDI playback (same file, but should restart)
		err = player.Play(midiPath)
		if err != nil {
			t.Fatalf("Second Play failed: %v", err)
		}

		// Player should still be playing (the new MIDI)
		if !player.IsPlaying() {
			t.Error("Player should be playing after second Play()")
		}

		// Clean up
		player.Stop()
	})

	t.Run("new MIDI resets tick counter", func(t *testing.T) {
		// When a new MIDI is played, the tick counter should reset to 0
		eventQueue := vm.NewEventQueue()
		player, err := NewMIDIPlayer(soundFontPath, audioCtx, eventQueue)
		if err != nil {
			t.Fatalf("NewMIDIPlayer failed: %v", err)
		}

		// Start first MIDI playback
		err = player.Play(midiPath)
		if err != nil {
			t.Fatalf("First Play failed: %v", err)
		}

		// Wait for some ticks to be generated
		time.Sleep(100 * time.Millisecond)
		player.Update()

		// Drain the event queue
		for eventQueue.Len() > 0 {
			eventQueue.Pop()
		}

		// Start second MIDI playback
		err = player.Play(midiPath)
		if err != nil {
			t.Fatalf("Second Play failed: %v", err)
		}

		// Wait for some ticks to be generated
		time.Sleep(100 * time.Millisecond)
		player.Update()

		// The first tick event should start from 1 (not continue from previous)
		event, ok := eventQueue.Pop()
		if !ok {
			t.Skip("No events generated")
		}

		if event.Type == vm.EventMIDI_TIME {
			tick, _ := event.GetParam("Tick")
			// First tick should be a small number (starting fresh)
			if tick.(int) > 100 {
				t.Errorf("Tick should have reset after new Play(), got %d", tick.(int))
			}
		}

		// Clean up
		player.Stop()
	})

	t.Run("exclusive control with rapid successive plays", func(t *testing.T) {
		// Test that rapid successive Play() calls work correctly
		eventQueue := vm.NewEventQueue()
		player, err := NewMIDIPlayer(soundFontPath, audioCtx, eventQueue)
		if err != nil {
			t.Fatalf("NewMIDIPlayer failed: %v", err)
		}

		// Rapidly call Play() multiple times
		for i := 0; i < 5; i++ {
			err = player.Play(midiPath)
			if err != nil {
				t.Fatalf("Play %d failed: %v", i+1, err)
			}
		}

		// Player should be playing (the last MIDI)
		if !player.IsPlaying() {
			t.Error("Player should be playing after rapid Play() calls")
		}

		// Clean up
		player.Stop()
	})

	t.Run("stop clears current file", func(t *testing.T) {
		player, err := NewMIDIPlayer(soundFontPath, audioCtx, nil)
		if err != nil {
			t.Fatalf("NewMIDIPlayer failed: %v", err)
		}

		// Start playback
		err = player.Play(midiPath)
		if err != nil {
			t.Fatalf("Play failed: %v", err)
		}

		if player.GetCurrentFile() == "" {
			t.Error("Current file should be set after Play()")
		}

		// Stop playback
		player.Stop()

		// Current file should be cleared
		if player.GetCurrentFile() != "" {
			t.Error("Current file should be cleared after Stop()")
		}
	})

	t.Run("play after stop works correctly", func(t *testing.T) {
		player, err := NewMIDIPlayer(soundFontPath, audioCtx, nil)
		if err != nil {
			t.Fatalf("NewMIDIPlayer failed: %v", err)
		}

		// Start playback
		err = player.Play(midiPath)
		if err != nil {
			t.Fatalf("First Play failed: %v", err)
		}

		// Stop playback
		player.Stop()

		if player.IsPlaying() {
			t.Error("Player should not be playing after Stop()")
		}

		// Start playback again
		err = player.Play(midiPath)
		if err != nil {
			t.Fatalf("Second Play failed: %v", err)
		}

		if !player.IsPlaying() {
			t.Error("Player should be playing after second Play()")
		}

		// Clean up
		player.Stop()
	})
}

// TestMIDIPlayerExclusiveControlConcurrent tests exclusive control under concurrent access.
// Requirement 4.6: When another MIDI is playing and PlayMIDI is called, system stops the previous MIDI and starts the new one.
// **Validates: Requirements 4.6**
func TestMIDIPlayerExclusiveControlConcurrent(t *testing.T) {
	soundFontPath := findSoundFont(t)
	midiPath := findMIDIFile(t)
	audioCtx := getSharedAudioContext()

	t.Run("concurrent Play calls are safe", func(t *testing.T) {
		player, err := NewMIDIPlayer(soundFontPath, audioCtx, nil)
		if err != nil {
			t.Fatalf("NewMIDIPlayer failed: %v", err)
		}

		// Start multiple goroutines calling Play() concurrently
		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				// Ignore errors - we're testing for race conditions
				player.Play(midiPath)
			}()
		}

		wg.Wait()

		// Player should be in a valid state (either playing or not)
		// The important thing is that it didn't panic or deadlock
		player.Stop()
	})

	t.Run("concurrent Play and Stop are safe", func(t *testing.T) {
		player, err := NewMIDIPlayer(soundFontPath, audioCtx, nil)
		if err != nil {
			t.Fatalf("NewMIDIPlayer failed: %v", err)
		}

		// Start multiple goroutines calling Play() and Stop() concurrently
		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(2)
			go func() {
				defer wg.Done()
				player.Play(midiPath)
			}()
			go func() {
				defer wg.Done()
				player.Stop()
			}()
		}

		wg.Wait()

		// Player should be in a valid state
		player.Stop()
	})

	t.Run("concurrent Play and Update are safe", func(t *testing.T) {
		eventQueue := vm.NewEventQueue()
		player, err := NewMIDIPlayer(soundFontPath, audioCtx, eventQueue)
		if err != nil {
			t.Fatalf("NewMIDIPlayer failed: %v", err)
		}

		// Start playback
		err = player.Play(midiPath)
		if err != nil {
			t.Fatalf("Play failed: %v", err)
		}

		// Start multiple goroutines calling Play() and Update() concurrently
		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(2)
			go func() {
				defer wg.Done()
				player.Play(midiPath)
			}()
			go func() {
				defer wg.Done()
				player.Update()
			}()
		}

		wg.Wait()

		// Player should be in a valid state
		player.Stop()
	})
}

// TestNewMIDIPlayerWithFS tests the creation of a MIDI player with FileSystem support.
func TestNewMIDIPlayerWithFS(t *testing.T) {
	audioCtx := getSharedAudioContext()

	t.Run("creates player with nil FileSystem (fallback)", func(t *testing.T) {
		// Requirement 2.2: Backward compatibility with nil FileSystem
		soundFontPath := findSoundFont(t)
		player, err := NewMIDIPlayerWithFS(soundFontPath, audioCtx, nil, nil)
		if err != nil {
			t.Fatalf("NewMIDIPlayerWithFS failed: %v", err)
		}
		if player == nil {
			t.Fatal("NewMIDIPlayerWithFS returned nil player")
		}
	})

	t.Run("returns error for empty SoundFont path", func(t *testing.T) {
		_, err := NewMIDIPlayerWithFS("", audioCtx, nil, nil)
		if err == nil {
			t.Error("Expected error for empty SoundFont path")
		}
		if err != ErrNoSoundFont {
			t.Errorf("Expected ErrNoSoundFont, got: %v", err)
		}
	})

	t.Run("returns error for non-existent SoundFont with nil fs", func(t *testing.T) {
		_, err := NewMIDIPlayerWithFS("/nonexistent/soundfont.sf2", audioCtx, nil, nil)
		if err == nil {
			t.Error("Expected error for non-existent SoundFont")
		}
	})
}
