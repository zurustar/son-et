package engine

import (
	"sync/atomic"
)

// WallClockTickGenerator generates MIDI ticks based on wall-clock time.
// This ensures accurate timing regardless of audio buffer size.
type WallClockTickGenerator struct {
	sampleRate        int
	ppq               int
	tempoMap          []TempoEvent
	lastDeliveredTick int32 // atomic
}

// NewWallClockTickGenerator creates a new tick generator.
func NewWallClockTickGenerator(sampleRate, ppq int, tempoMap []TempoEvent) *WallClockTickGenerator {
	return &WallClockTickGenerator{
		sampleRate:        sampleRate,
		ppq:               ppq,
		tempoMap:          tempoMap,
		lastDeliveredTick: -1,
	}
}

// CalculateTickFromTime calculates the current tick from elapsed time.
// elapsed is in seconds since playback started.
func (tg *WallClockTickGenerator) CalculateTickFromTime(elapsed float64) int {
	// Convert elapsed time to ticks based on tempo map
	currentTick := 0
	remainingTime := elapsed

	for i := 0; i < len(tg.tempoMap); i++ {
		// Get current tempo event
		tempoEvent := tg.tempoMap[i]
		microsPerBeat := float64(tempoEvent.MicrosPerBeat)

		// Calculate time per MIDI tick
		// 1 quarter note = PPQ ticks
		// Time per tick = (microsPerBeat / 1000000) / PPQ
		timePerTick := (microsPerBeat / 1000000.0) / float64(tg.ppq)

		// Determine how long this tempo lasts
		var tempoDuration float64
		if i+1 < len(tg.tempoMap) {
			// Calculate ticks until next tempo change
			nextTempoTick := tg.tempoMap[i+1].Tick
			ticksInThisTempo := nextTempoTick - tempoEvent.Tick
			tempoDuration = float64(ticksInThisTempo) * timePerTick
		} else {
			// Last tempo - use all remaining time
			tempoDuration = remainingTime + 1.0 // Ensure we use all remaining time
		}

		if remainingTime <= tempoDuration {
			// Current time falls within this tempo
			ticksElapsed := int(remainingTime / timePerTick)
			currentTick = tempoEvent.Tick + ticksElapsed
			break
		} else {
			// Move to next tempo
			remainingTime -= tempoDuration
			if i+1 < len(tg.tempoMap) {
				currentTick = tg.tempoMap[i+1].Tick
			}
		}
	}

	return currentTick
}

// GetLastDeliveredTick returns the last tick that was delivered.
func (tg *WallClockTickGenerator) GetLastDeliveredTick() int {
	return int(atomic.LoadInt32(&tg.lastDeliveredTick))
}

// SetLastDeliveredTick sets the last delivered tick (for tracking).
func (tg *WallClockTickGenerator) SetLastDeliveredTick(tick int) {
	atomic.StoreInt32(&tg.lastDeliveredTick, int32(tick))
}
