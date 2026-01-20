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
	tempoMap := []TempoEvent{{Tick: 0, MicrosPerBeat: 500000}}
	tg := NewWallClockTickGenerator(44100, 480, tempoMap)

	// At 120 BPM:
	// 1 beat = 0.5 seconds
	// 1 quarter note = 0.5 seconds
	// 1 32nd note = 0.5 / 8 = 0.0625 seconds
	// So 1 tick per 0.0625 seconds

	tests := []struct {
		elapsed  float64
		expected int
	}{
		{0.0, 0},
		{0.0625, 1},
		{0.125, 2},
		{0.5, 8},  // 1 quarter note = 8 ticks
		{1.0, 16}, // 2 quarter notes = 16 ticks
		{2.0, 32}, // 4 quarter notes = 32 ticks
	}

	for _, tt := range tests {
		tick := tg.CalculateTickFromTime(tt.elapsed)
		if tick != tt.expected {
			t.Errorf("At %.3f seconds, expected tick %d, got %d", tt.elapsed, tt.expected, tick)
		}
	}
}

func TestCalculateTickFromTime_TempoChange(t *testing.T) {
	// Start at 120 BPM, change to 60 BPM at tick 16
	tempoMap := []TempoEvent{
		{Tick: 0, MicrosPerBeat: 500000},   // 120 BPM
		{Tick: 16, MicrosPerBeat: 1000000}, // 60 BPM
	}
	tg := NewWallClockTickGenerator(44100, 480, tempoMap)

	// At 120 BPM: 1 tick = 0.0625 seconds
	// At 60 BPM: 1 tick = 0.125 seconds

	// First 16 ticks at 120 BPM = 16 * 0.0625 = 1.0 second
	tick := tg.CalculateTickFromTime(1.0)
	if tick != 16 {
		t.Errorf("At 1.0 seconds, expected tick 16, got %d", tick)
	}

	// Next 8 ticks at 60 BPM = 8 * 0.125 = 1.0 second
	// Total: 1.0 + 1.0 = 2.0 seconds, tick = 16 + 8 = 24
	tick = tg.CalculateTickFromTime(2.0)
	if tick != 24 {
		t.Errorf("At 2.0 seconds, expected tick 24, got %d", tick)
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
	tempoMap := []TempoEvent{{Tick: 0, MicrosPerBeat: 250000}}
	tg := NewWallClockTickGenerator(44100, 480, tempoMap)

	// At 240 BPM:
	// 1 beat = 0.25 seconds
	// 1 32nd note = 0.25 / 8 = 0.03125 seconds
	// So 1 tick per 0.03125 seconds

	tick := tg.CalculateTickFromTime(0.5)
	expected := 16 // 0.5 / 0.03125 = 16
	if tick != expected {
		t.Errorf("At 0.5 seconds (240 BPM), expected tick %d, got %d", expected, tick)
	}
}

func TestCalculateTickFromTime_SlowTempo(t *testing.T) {
	// 60 BPM = 1000000 microseconds per beat (half as fast as 120 BPM)
	tempoMap := []TempoEvent{{Tick: 0, MicrosPerBeat: 1000000}}
	tg := NewWallClockTickGenerator(44100, 480, tempoMap)

	// At 60 BPM:
	// 1 beat = 1.0 seconds
	// 1 32nd note = 1.0 / 8 = 0.125 seconds
	// So 1 tick per 0.125 seconds

	tick := tg.CalculateTickFromTime(1.0)
	expected := 8 // 1.0 / 0.125 = 8
	if tick != expected {
		t.Errorf("At 1.0 seconds (60 BPM), expected tick %d, got %d", expected, tick)
	}
}
