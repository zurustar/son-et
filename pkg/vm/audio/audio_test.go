// Package audio provides audio-related components for the FILLY virtual machine.
// This file contains tests for the AudioSystem integration.
package audio

import (
	"testing"
	"time"

	"github.com/zurustar/son-et/pkg/vm"
)

// Note: findSoundFont, findMIDIFile, and getSharedAudioContext are defined in midi_test.go
// Note: findSampleWAVFile is defined in wav_test.go

// TestNewAudioSystem tests the creation of a new AudioSystem.
func TestNewAudioSystem(t *testing.T) {
	soundFontPath := findSoundFont(t)
	eventQueue := vm.NewEventQueue()
	audioCtx := getSharedAudioContext()

	as, err := NewAudioSystemWithContext(soundFontPath, eventQueue, audioCtx)
	if err != nil {
		t.Fatalf("NewAudioSystemWithContext failed: %v", err)
	}
	defer as.Shutdown()

	// Verify components are initialized
	if as.GetMIDIPlayer() == nil {
		t.Error("MIDI player should be initialized")
	}
	if as.GetWAVPlayer() == nil {
		t.Error("WAV player should be initialized")
	}
	if as.GetTimer() == nil {
		t.Error("Timer should be initialized")
	}
	if as.GetAudioContext() == nil {
		t.Error("Audio context should be initialized")
	}
	if as.GetEventQueue() != eventQueue {
		t.Error("Event queue should match the provided queue")
	}
}

// TestNewAudioSystemNoSoundFont tests that NewAudioSystemWithContext returns an error when no SoundFont is provided.
func TestNewAudioSystemNoSoundFont(t *testing.T) {
	eventQueue := vm.NewEventQueue()
	audioCtx := getSharedAudioContext()

	_, err := NewAudioSystemWithContext("", eventQueue, audioCtx)
	if err == nil {
		t.Error("NewAudioSystemWithContext should return error when no SoundFont is provided")
	}
	if err != ErrNoSoundFont {
		t.Errorf("Expected ErrNoSoundFont, got: %v", err)
	}
}

// TestNewAudioSystemInvalidSoundFont tests that NewAudioSystemWithContext returns an error for invalid SoundFont.
func TestNewAudioSystemInvalidSoundFont(t *testing.T) {
	eventQueue := vm.NewEventQueue()
	audioCtx := getSharedAudioContext()

	_, err := NewAudioSystemWithContext("/nonexistent/path/soundfont.sf2", eventQueue, audioCtx)
	if err == nil {
		t.Error("NewAudioSystemWithContext should return error for invalid SoundFont path")
	}
}

// TestAudioSystemSetMuted tests the SetMuted functionality.
func TestAudioSystemSetMuted(t *testing.T) {
	soundFontPath := findSoundFont(t)
	eventQueue := vm.NewEventQueue()
	audioCtx := getSharedAudioContext()

	as, err := NewAudioSystemWithContext(soundFontPath, eventQueue, audioCtx)
	if err != nil {
		t.Fatalf("NewAudioSystemWithContext failed: %v", err)
	}
	defer as.Shutdown()

	// Initially not muted
	if as.IsMuted() {
		t.Error("AudioSystem should not be muted initially")
	}

	// Set muted
	as.SetMuted(true)
	if !as.IsMuted() {
		t.Error("AudioSystem should be muted after SetMuted(true)")
	}

	// Verify MIDI player is muted
	if !as.GetMIDIPlayer().IsMuted() {
		t.Error("MIDI player should be muted")
	}

	// Verify WAV player is muted
	if !as.GetWAVPlayer().IsMuted() {
		t.Error("WAV player should be muted")
	}

	// Unmute
	as.SetMuted(false)
	if as.IsMuted() {
		t.Error("AudioSystem should not be muted after SetMuted(false)")
	}
}

// TestAudioSystemTimer tests the timer start/stop functionality.
func TestAudioSystemTimer(t *testing.T) {
	soundFontPath := findSoundFont(t)
	eventQueue := vm.NewEventQueue()
	audioCtx := getSharedAudioContext()

	as, err := NewAudioSystemWithContext(soundFontPath, eventQueue, audioCtx)
	if err != nil {
		t.Fatalf("NewAudioSystemWithContext failed: %v", err)
	}
	defer as.Shutdown()

	// Timer should not be running initially
	if as.IsTimerRunning() {
		t.Error("Timer should not be running initially")
	}

	// Start timer
	as.StartTimer()
	if !as.IsTimerRunning() {
		t.Error("Timer should be running after StartTimer()")
	}

	// Wait for some TIME events
	time.Sleep(150 * time.Millisecond)

	// Check that TIME events were generated
	eventCount := eventQueue.Len()
	if eventCount == 0 {
		t.Error("Timer should have generated TIME events")
	}

	// Stop timer
	as.StopTimer()
	if as.IsTimerRunning() {
		t.Error("Timer should not be running after StopTimer()")
	}
}

// TestAudioSystemShutdown tests the Shutdown functionality.
func TestAudioSystemShutdown(t *testing.T) {
	soundFontPath := findSoundFont(t)
	eventQueue := vm.NewEventQueue()
	audioCtx := getSharedAudioContext()

	as, err := NewAudioSystemWithContext(soundFontPath, eventQueue, audioCtx)
	if err != nil {
		t.Fatalf("NewAudioSystemWithContext failed: %v", err)
	}

	// Start timer
	as.StartTimer()

	// Shutdown
	as.Shutdown()

	// Timer should be stopped
	if as.IsTimerRunning() {
		t.Error("Timer should be stopped after Shutdown()")
	}

	// MIDI should not be playing
	if as.IsMIDIPlaying() {
		t.Error("MIDI should not be playing after Shutdown()")
	}
}

// TestAudioSystemUpdate tests the Update functionality.
func TestAudioSystemUpdate(t *testing.T) {
	soundFontPath := findSoundFont(t)
	eventQueue := vm.NewEventQueue()
	audioCtx := getSharedAudioContext()

	as, err := NewAudioSystemWithContext(soundFontPath, eventQueue, audioCtx)
	if err != nil {
		t.Fatalf("NewAudioSystemWithContext failed: %v", err)
	}
	defer as.Shutdown()

	// Update should not panic when nothing is playing
	as.Update()
}

// TestAudioSystemPlayMIDIWithoutSoundFont tests PlayMIDI error handling.
func TestAudioSystemPlayMIDIWithoutSoundFont(t *testing.T) {
	// Create a minimal AudioSystem without proper initialization
	// This tests the error path when midiPlayer is nil
	as := &AudioSystem{
		midiPlayer: nil,
	}

	err := as.PlayMIDI("test.mid")
	if err == nil {
		t.Error("PlayMIDI should return error when MIDI player is not initialized")
	}
}

// TestAudioSystemPlayWAVEWithoutPlayer tests PlayWAVE when player is nil.
func TestAudioSystemPlayWAVEWithoutPlayer(t *testing.T) {
	// Create a minimal AudioSystem without WAV player
	as := &AudioSystem{
		wavPlayer: nil,
	}

	// Should not return error, just skip
	err := as.PlayWAVE("test.wav")
	if err != nil {
		t.Errorf("PlayWAVE should not return error when WAV player is nil: %v", err)
	}
}

// TestAudioSystemStopMIDI tests the StopMIDI functionality.
func TestAudioSystemStopMIDI(t *testing.T) {
	soundFontPath := findSoundFont(t)
	eventQueue := vm.NewEventQueue()
	audioCtx := getSharedAudioContext()

	as, err := NewAudioSystemWithContext(soundFontPath, eventQueue, audioCtx)
	if err != nil {
		t.Fatalf("NewAudioSystemWithContext failed: %v", err)
	}
	defer as.Shutdown()

	// StopMIDI should not panic when nothing is playing
	as.StopMIDI()

	// MIDI should not be playing
	if as.IsMIDIPlaying() {
		t.Error("MIDI should not be playing after StopMIDI()")
	}
}

// TestAudioSystemStopAllWAV tests the StopAllWAV functionality.
func TestAudioSystemStopAllWAV(t *testing.T) {
	soundFontPath := findSoundFont(t)
	eventQueue := vm.NewEventQueue()
	audioCtx := getSharedAudioContext()

	as, err := NewAudioSystemWithContext(soundFontPath, eventQueue, audioCtx)
	if err != nil {
		t.Fatalf("NewAudioSystemWithContext failed: %v", err)
	}
	defer as.Shutdown()

	// StopAllWAV should not panic when nothing is playing
	as.StopAllWAV()
}

// TestAudioSystemIntegration tests the integration of all audio components.
func TestAudioSystemIntegration(t *testing.T) {
	soundFontPath := findSoundFont(t)
	eventQueue := vm.NewEventQueue()
	audioCtx := getSharedAudioContext()

	as, err := NewAudioSystemWithContext(soundFontPath, eventQueue, audioCtx)
	if err != nil {
		t.Fatalf("NewAudioSystemWithContext failed: %v", err)
	}
	defer as.Shutdown()

	// Verify all components share the same audio context
	ctx := as.GetAudioContext()
	if ctx == nil {
		t.Fatal("Audio context should not be nil")
	}

	// The MIDI player and WAV player should use the same audio context
	// This is verified by the fact that they were created with the same context
	// in NewAudioSystem

	// Test muting affects all components
	as.SetMuted(true)
	if !as.GetMIDIPlayer().IsMuted() {
		t.Error("MIDI player should be muted when AudioSystem is muted")
	}
	if !as.GetWAVPlayer().IsMuted() {
		t.Error("WAV player should be muted when AudioSystem is muted")
	}

	// Test unmuting
	as.SetMuted(false)
	if as.GetMIDIPlayer().IsMuted() {
		t.Error("MIDI player should not be muted when AudioSystem is unmuted")
	}
	if as.GetWAVPlayer().IsMuted() {
		t.Error("WAV player should not be muted when AudioSystem is unmuted")
	}
}

// TestAudioSystemWithMIDIFile tests playing a MIDI file through AudioSystem.
func TestAudioSystemWithMIDIFile(t *testing.T) {
	soundFontPath := findSoundFont(t)
	midiPath := findMIDIFile(t)
	eventQueue := vm.NewEventQueue()
	audioCtx := getSharedAudioContext()

	as, err := NewAudioSystemWithContext(soundFontPath, eventQueue, audioCtx)
	if err != nil {
		t.Fatalf("NewAudioSystemWithContext failed: %v", err)
	}
	defer as.Shutdown()

	// Mute to avoid audio output during test
	as.SetMuted(true)

	// Play MIDI file
	err = as.PlayMIDI(midiPath)
	if err != nil {
		t.Fatalf("PlayMIDI failed: %v", err)
	}

	// Verify MIDI is playing
	if !as.IsMIDIPlaying() {
		t.Error("MIDI should be playing after PlayMIDI()")
	}

	// Update a few times to generate events
	for i := 0; i < 10; i++ {
		as.Update()
		time.Sleep(10 * time.Millisecond)
	}

	// Stop MIDI
	as.StopMIDI()
	if as.IsMIDIPlaying() {
		t.Error("MIDI should not be playing after StopMIDI()")
	}
}

// TestAudioSystemWithWAVFile tests playing a WAV file through AudioSystem.
func TestAudioSystemWithWAVFile(t *testing.T) {
	soundFontPath := findSoundFont(t)
	wavPath := findSampleWAVFile()
	if wavPath == "" {
		t.Skip("WAV file not found, skipping test")
	}

	eventQueue := vm.NewEventQueue()
	audioCtx := getSharedAudioContext()

	as, err := NewAudioSystemWithContext(soundFontPath, eventQueue, audioCtx)
	if err != nil {
		t.Fatalf("NewAudioSystemWithContext failed: %v", err)
	}
	defer as.Shutdown()

	// Mute to avoid audio output during test
	as.SetMuted(true)

	// Play WAV file
	err = as.PlayWAVE(wavPath)
	if err != nil {
		t.Fatalf("PlayWAVE failed: %v", err)
	}

	// Update to clean up finished players
	as.Update()
}

// TestAudioSystemConcurrentAccess tests thread-safety of AudioSystem.
func TestAudioSystemConcurrentAccess(t *testing.T) {
	soundFontPath := findSoundFont(t)
	eventQueue := vm.NewEventQueue()
	audioCtx := getSharedAudioContext()

	as, err := NewAudioSystemWithContext(soundFontPath, eventQueue, audioCtx)
	if err != nil {
		t.Fatalf("NewAudioSystemWithContext failed: %v", err)
	}
	defer as.Shutdown()

	// Run concurrent operations
	done := make(chan bool)

	// Goroutine 1: Toggle mute
	go func() {
		for i := 0; i < 100; i++ {
			as.SetMuted(i%2 == 0)
			_ = as.IsMuted()
		}
		done <- true
	}()

	// Goroutine 2: Check status
	go func() {
		for i := 0; i < 100; i++ {
			_ = as.IsMIDIPlaying()
			_ = as.IsTimerRunning()
		}
		done <- true
	}()

	// Goroutine 3: Update
	go func() {
		for i := 0; i < 100; i++ {
			as.Update()
		}
		done <- true
	}()

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}
}
