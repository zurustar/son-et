package engine

import (
	"testing"
)

func TestNewWallClockTickGenerator(t *testing.T) {
	tempoMap := []TempoEvent{{Tick: 0, MicrosPerBeat: 500000}} // 120 BPM
	tg := NewWallClockTickGenerator(44100, 480, tempoMap)

	if tg == nil {
		t.Fatal("Tick generator not created")
	}

	if tg.sampleRate != 44100 {
		t.Errorf("Expected sample rate 44100, got %d", tg.sampleRate)
	}

	if tg.ppq != 480 {
		t.Errorf("Expected PPQ 480, got %d", tg.ppq)
	}

	if tg.GetLastDeliveredTick() != -1 {
		t.Errorf("Expected initial last delivered tick -1, got %d", tg.GetLastDeliveredTick())
	}
}

func TestCalculateTickFromTime_ConstantTempo(t *testing.T) {
	// 120 BPM = 500000 microseconds per beat
	// PPQ = 480 (MIDI ticks per quarter note)
	tempoMap := []TempoEvent{{Tick: 0, MicrosPerBeat: 500000}}
	tg := NewWallClockTickGenerator(44100, 480, tempoMap)

	// At 120 BPM with PPQ=480:
	// 1 beat = 0.5 seconds
	// 1 quarter note = 0.5 seconds = 480 MIDI ticks
	// Time per MIDI tick = 0.5 / 480 = 0.00104166... seconds
	// So 1 MIDI tick per ~0.001042 seconds

	tests := []struct {
		elapsed  float64
		expected int
	}{
		{0.0, 0},
		{0.0625, 60}, // 0.0625 / 0.00104166 ≈ 60 MIDI ticks
		{0.125, 120}, // 0.125 / 0.00104166 ≈ 120 MIDI ticks
		{0.5, 480},   // 1 quarter note = 480 MIDI ticks
		{1.0, 960},   // 2 quarter notes = 960 MIDI ticks
		{2.0, 1920},  // 4 quarter notes = 1920 MIDI ticks
	}

	for _, tt := range tests {
		tick := tg.CalculateTickFromTime(tt.elapsed)
		if tick != tt.expected {
			t.Errorf("At %.3f seconds, expected tick %d, got %d", tt.elapsed, tt.expected, tick)
		}
	}
}

func TestCalculateTickFromTime_TempoChange(t *testing.T) {
	// Start at 120 BPM, change to 60 BPM at tick 960 (MIDI ticks)
	// Note: The tempo map uses MIDI ticks, not FILLY ticks
	tempoMap := []TempoEvent{
		{Tick: 0, MicrosPerBeat: 500000},    // 120 BPM
		{Tick: 960, MicrosPerBeat: 1000000}, // 60 BPM at MIDI tick 960
	}
	tg := NewWallClockTickGenerator(44100, 480, tempoMap)

	// At 120 BPM with PPQ=480: time per MIDI tick = 0.5 / 480 ≈ 0.001042 seconds
	// At 60 BPM with PPQ=480: time per MIDI tick = 1.0 / 480 ≈ 0.002083 seconds

	// First 960 MIDI ticks at 120 BPM = 960 * 0.001042 ≈ 1.0 second
	tick := tg.CalculateTickFromTime(1.0)
	if tick != 960 {
		t.Errorf("At 1.0 seconds, expected tick 960, got %d", tick)
	}

	// Next 480 MIDI ticks at 60 BPM = 480 * 0.002083 ≈ 1.0 second
	// Total: 1.0 + 1.0 = 2.0 seconds, tick = 960 + 480 = 1440
	tick = tg.CalculateTickFromTime(2.0)
	if tick != 1440 {
		t.Errorf("At 2.0 seconds, expected tick 1440, got %d", tick)
	}
}

func TestSetLastDeliveredTick(t *testing.T) {
	tempoMap := []TempoEvent{{Tick: 0, MicrosPerBeat: 500000}}
	tg := NewWallClockTickGenerator(44100, 480, tempoMap)

	tg.SetLastDeliveredTick(10)
	if tg.GetLastDeliveredTick() != 10 {
		t.Errorf("Expected last delivered tick 10, got %d", tg.GetLastDeliveredTick())
	}

	tg.SetLastDeliveredTick(100)
	if tg.GetLastDeliveredTick() != 100 {
		t.Errorf("Expected last delivered tick 100, got %d", tg.GetLastDeliveredTick())
	}
}

func TestCalculateTickFromTime_FastTempo(t *testing.T) {
	// 240 BPM = 250000 microseconds per beat (twice as fast as 120 BPM)
	// PPQ = 480
	tempoMap := []TempoEvent{{Tick: 0, MicrosPerBeat: 250000}}
	tg := NewWallClockTickGenerator(44100, 480, tempoMap)

	// At 240 BPM with PPQ=480:
	// 1 beat = 0.25 seconds = 480 MIDI ticks
	// Time per MIDI tick = 0.25 / 480 ≈ 0.000521 seconds
	// So at 0.5 seconds: 0.5 / 0.000521 ≈ 960 MIDI ticks

	tick := tg.CalculateTickFromTime(0.5)
	expected := 960 // 0.5 seconds = 2 beats = 2 * 480 = 960 MIDI ticks
	if tick != expected {
		t.Errorf("At 0.5 seconds (240 BPM), expected tick %d, got %d", expected, tick)
	}
}

func TestCalculateTickFromTime_SlowTempo(t *testing.T) {
	// 60 BPM = 1000000 microseconds per beat (half as fast as 120 BPM)
	// PPQ = 480
	tempoMap := []TempoEvent{{Tick: 0, MicrosPerBeat: 1000000}}
	tg := NewWallClockTickGenerator(44100, 480, tempoMap)

	// At 60 BPM with PPQ=480:
	// 1 beat = 1.0 seconds = 480 MIDI ticks
	// Time per MIDI tick = 1.0 / 480 ≈ 0.002083 seconds
	// So at 1.0 seconds: 1.0 / 0.002083 ≈ 480 MIDI ticks

	tick := tg.CalculateTickFromTime(1.0)
	expected := 480 // 1.0 seconds = 1 beat = 480 MIDI ticks
	if tick != expected {
		t.Errorf("At 1.0 seconds (60 BPM), expected tick %d, got %d", expected, tick)
	}
}
