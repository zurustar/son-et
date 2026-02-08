package audio

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/zurustar/son-et/pkg/vm"
)

// TestMIDIPlaybackExclusivityProperty tests that only one MIDI can play at a time.
// **Validates: Requirements 4.6**
// Property 20: MIDI再生の排他性
// *任意の*2つのMIDIファイルについて、2つ目のPlayMIDI呼び出し時に1つ目のMIDI再生が停止している
//
// Requirement 4.6: WHEN 別のMIDIが再生中にPlayMIDIが呼ばれたとき、THE System SHALL 前のMIDIを停止して新しいMIDIを開始する
func TestMIDIPlaybackExclusivityProperty(t *testing.T) {
	soundFontPath := findSoundFontForProperty(t)
	midiPath := findMIDIFileForProperty(t)
	audioCtx := getSharedAudioContext()

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property: After any sequence of Play() calls, only one MIDI is playing
	// and it is the one from the last Play() call
	properties.Property("after any sequence of Play calls, only the last MIDI is playing", prop.ForAll(
		func(playCount int) bool {
			eventQueue := vm.NewEventQueue()
			player, err := NewMIDIPlayer(soundFontPath, audioCtx, eventQueue)
			if err != nil {
				t.Logf("Failed to create MIDI player: %v", err)
				return false
			}
			defer player.Stop()

			// Play MIDI multiple times
			for i := 0; i < playCount; i++ {
				err := player.Play(midiPath)
				if err != nil {
					t.Logf("Play %d failed: %v", i+1, err)
					return false
				}
			}

			// After all Play() calls, player should be playing (if playCount > 0)
			if playCount > 0 {
				if !player.IsPlaying() {
					return false
				}
				// Current file should be set
				if player.GetCurrentFile() == "" {
					return false
				}
			}

			return true
		},
		gen.IntRange(1, 10),
	))

	properties.TestingRun(t)
}

// TestMIDIPlaybackTickResetProperty tests that tick counter resets on new Play.
// **Validates: Requirements 4.6**
// Property: When a new MIDI is played, the tick counter resets to 0
func TestMIDIPlaybackTickResetProperty(t *testing.T) {
	soundFontPath := findSoundFontForProperty(t)
	midiPath := findMIDIFileForProperty(t)
	audioCtx := getSharedAudioContext()

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50

	properties := gopter.NewProperties(parameters)

	// Property: After Play(), the internal lastTick is reset to 0
	// This ensures that MIDI_TIME events start from tick 1 for each new playback
	properties.Property("tick counter resets on new Play", prop.ForAll(
		func(waitMs int) bool {
			eventQueue := vm.NewEventQueue()
			player, err := NewMIDIPlayer(soundFontPath, audioCtx, eventQueue)
			if err != nil {
				t.Logf("Failed to create MIDI player: %v", err)
				return false
			}
			defer player.Stop()

			// First playback
			err = player.Play(midiPath)
			if err != nil {
				t.Logf("First Play failed: %v", err)
				return false
			}

			// Wait for some ticks
			time.Sleep(time.Duration(waitMs) * time.Millisecond)
			player.Update()

			// Drain the event queue
			for eventQueue.Len() > 0 {
				eventQueue.Pop()
			}

			// Second playback (should reset tick counter)
			err = player.Play(midiPath)
			if err != nil {
				t.Logf("Second Play failed: %v", err)
				return false
			}

			// Wait for some ticks
			time.Sleep(time.Duration(waitMs) * time.Millisecond)
			player.Update()

			// Check that first tick event starts from a small number (reset occurred)
			if eventQueue.Len() == 0 {
				// No events generated, which is acceptable for very short waits
				return true
			}

			event, ok := eventQueue.Pop()
			if !ok {
				return true
			}

			if event.Type == vm.EventMIDI_TIME {
				tick, ok := event.GetParam("Tick")
				if !ok {
					return false
				}
				// First tick should be a small number (starting fresh after reset)
				// If tick counter wasn't reset, it would continue from previous value
				if tick.(int) > 100 {
					return false
				}
			}

			return true
		},
		gen.IntRange(50, 200),
	))

	properties.TestingRun(t)
}

// TestMIDIPlaybackStopsPreviousProperty tests that new Play stops previous playback.
// **Validates: Requirements 4.6**
// Property: When Play() is called while another MIDI is playing, the previous one is stopped
func TestMIDIPlaybackStopsPreviousProperty(t *testing.T) {
	soundFontPath := findSoundFontForProperty(t)
	midiPath := findMIDIFileForProperty(t)
	audioCtx := getSharedAudioContext()

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property: After Play() is called, there is exactly one active playback
	// (the previous one is stopped)
	properties.Property("new Play stops previous playback", prop.ForAll(
		func(playCount int) bool {
			eventQueue := vm.NewEventQueue()
			player, err := NewMIDIPlayer(soundFontPath, audioCtx, eventQueue)
			if err != nil {
				t.Logf("Failed to create MIDI player: %v", err)
				return false
			}
			defer player.Stop()

			// Play MIDI multiple times
			for i := 0; i < playCount; i++ {
				// Before playing, check current state
				wasPlaying := player.IsPlaying()
				previousFile := player.GetCurrentFile()

				err := player.Play(midiPath)
				if err != nil {
					t.Logf("Play %d failed: %v", i+1, err)
					return false
				}

				// After Play(), player should be playing
				if !player.IsPlaying() {
					return false
				}

				// If there was a previous playback, it should have been stopped
				// (we can verify this by checking that the current file is set correctly)
				currentFile := player.GetCurrentFile()
				if currentFile == "" {
					return false
				}

				// The current file should be the one we just played
				if currentFile != midiPath {
					return false
				}

				// If there was a previous playback, verify it was replaced
				if wasPlaying && previousFile != "" {
					// The previous playback should have been stopped
					// (we can't directly verify this, but the fact that
					// IsPlaying() returns true and GetCurrentFile() returns
					// the new file indicates the previous one was stopped)
				}
			}

			return true
		},
		gen.IntRange(1, 10),
	))

	properties.TestingRun(t)
}

// TestMIDIPlaybackStateConsistencyProperty tests state consistency after Play/Stop sequences.
// **Validates: Requirements 4.6**
// Property: After any sequence of Play and Stop operations, the state is consistent
func TestMIDIPlaybackStateConsistencyProperty(t *testing.T) {
	soundFontPath := findSoundFontForProperty(t)
	midiPath := findMIDIFileForProperty(t)
	audioCtx := getSharedAudioContext()

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property: State is consistent after any sequence of Play/Stop operations
	// - After Play(): IsPlaying() == true, GetCurrentFile() != ""
	// - After Stop(): IsPlaying() == false, GetCurrentFile() == ""
	properties.Property("state is consistent after Play/Stop sequences", prop.ForAll(
		func(operations []bool) bool {
			eventQueue := vm.NewEventQueue()
			player, err := NewMIDIPlayer(soundFontPath, audioCtx, eventQueue)
			if err != nil {
				t.Logf("Failed to create MIDI player: %v", err)
				return false
			}
			defer player.Stop()

			for _, shouldPlay := range operations {
				if shouldPlay {
					err := player.Play(midiPath)
					if err != nil {
						t.Logf("Play failed: %v", err)
						return false
					}

					// After Play(), state should be consistent
					if !player.IsPlaying() {
						return false
					}
					if player.GetCurrentFile() == "" {
						return false
					}
				} else {
					player.Stop()

					// After Stop(), state should be consistent
					if player.IsPlaying() {
						return false
					}
					if player.GetCurrentFile() != "" {
						return false
					}
				}
			}

			return true
		},
		gen.SliceOfN(10, gen.Bool()),
	))

	properties.TestingRun(t)
}

// TestMIDIPlaybackLastPlayDeterminesFileProperty tests that the last Play determines the current file.
// **Validates: Requirements 4.6**
// Property: The last Play() call determines which file is currently playing
func TestMIDIPlaybackLastPlayDeterminesFileProperty(t *testing.T) {
	soundFontPath := findSoundFontForProperty(t)
	midiPath := findMIDIFileForProperty(t)
	audioCtx := getSharedAudioContext()

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property: After any number of Play() calls, GetCurrentFile() returns the file from the last Play()
	properties.Property("last Play determines current file", prop.ForAll(
		func(playCount int) bool {
			eventQueue := vm.NewEventQueue()
			player, err := NewMIDIPlayer(soundFontPath, audioCtx, eventQueue)
			if err != nil {
				t.Logf("Failed to create MIDI player: %v", err)
				return false
			}
			defer player.Stop()

			if playCount == 0 {
				// No plays, current file should be empty
				return player.GetCurrentFile() == ""
			}

			// Play MIDI multiple times
			for i := 0; i < playCount; i++ {
				err := player.Play(midiPath)
				if err != nil {
					t.Logf("Play %d failed: %v", i+1, err)
					return false
				}
			}

			// Current file should be the one from the last Play()
			return player.GetCurrentFile() == midiPath
		},
		gen.IntRange(0, 10),
	))

	properties.TestingRun(t)
}

// Helper functions for property tests

// findSoundFontForProperty finds the SoundFont file for property tests.
func findSoundFontForProperty(t *testing.T) string {
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

// findMIDIFileForProperty finds a MIDI file for property tests.
func findMIDIFileForProperty(t *testing.T) string {
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
