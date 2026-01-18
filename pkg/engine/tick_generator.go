package engine

import (
	"fmt"
)

// TickGenerator converts audio sample count to MIDI ticks with fractional precision
type TickGenerator struct {
	// Configuration (immutable)
	sampleRate int
	ppq        int
	tempoMap   []TempoEvent

	// State (mutable during playback)
	currentSamples    int64
	fractionalTick    float64
	lastDeliveredTick int
	tempoMapIndex     int
	currentTempo      float64 // Current tempo in BPM (cached)
}

// TempoEvent represents a tempo change at a specific tick
type TempoEvent struct {
	Tick          int
	MicrosPerBeat int
}

// NewTickGenerator creates a new tick generator with validation
func NewTickGenerator(sampleRate, ppq int, tempoMap []TempoEvent) (*TickGenerator, error) {
	// Validate sample rate
	if sampleRate <= 0 {
		return nil, fmt.Errorf("invalid sample rate: %d (must be positive)", sampleRate)
	}
	if sampleRate < 8000 || sampleRate > 192000 {
		return nil, fmt.Errorf("invalid sample rate: %d (must be between 8000 and 192000)", sampleRate)
	}

	// Validate PPQ
	if ppq <= 0 {
		return nil, fmt.Errorf("invalid PPQ: %d (must be positive)", ppq)
	}

	// Ensure we have at least one tempo event (default 120 BPM)
	if len(tempoMap) == 0 {
		tempoMap = []TempoEvent{{Tick: 0, MicrosPerBeat: 500000}} // 120 BPM
	}

	// Calculate initial tempo in BPM
	initialTempoBPM := 60000000.0 / float64(tempoMap[0].MicrosPerBeat)

	tg := &TickGenerator{
		sampleRate:        sampleRate,
		ppq:               ppq,
		tempoMap:          tempoMap,
		currentSamples:    0,
		fractionalTick:    0.0,
		lastDeliveredTick: 0,
		tempoMapIndex:     0,
		currentTempo:      initialTempoBPM,
	}

	return tg, nil
}

// ProcessSamples updates the tick position based on samples rendered
// Returns the new tick value if it has advanced, or -1 if no change
func (tg *TickGenerator) ProcessSamples(numSamples int) int {
	if numSamples <= 0 {
		return -1
	}

	tg.currentSamples += int64(numSamples)

	// Check if we need to update tempo based on current tick position
	tg.updateTempoIfNeeded()

	// Calculate tick advancement using formula:
	// ticks = (samples * tempo_bpm * ppq) / (sample_rate * 60)
	timeSec := float64(numSamples) / float64(tg.sampleRate)
	tickDelta := timeSec * (tg.currentTempo / 60.0) * float64(tg.ppq)

	// Update fractional tick
	tg.fractionalTick += tickDelta

	// Check if we crossed a tempo boundary during this buffer
	tg.updateTempoIfNeeded()

	// Check if integer part changed
	newTick := int(tg.fractionalTick)

	if newTick > tg.lastDeliveredTick {
		tg.lastDeliveredTick = newTick
		return newTick
	}

	return -1
}

// updateTempoIfNeeded checks if we've crossed into a new tempo segment
func (tg *TickGenerator) updateTempoIfNeeded() {
	currentTick := int(tg.fractionalTick)

	// Check if we need to advance to next tempo event
	for tg.tempoMapIndex+1 < len(tg.tempoMap) {
		nextEvent := tg.tempoMap[tg.tempoMapIndex+1]
		if currentTick >= nextEvent.Tick {
			// We've crossed into the next tempo segment
			tg.tempoMapIndex++
			tg.currentTempo = 60000000.0 / float64(nextEvent.MicrosPerBeat)
		} else {
			break
		}
	}
}

// GetCurrentTick returns the current integer tick position
func (tg *TickGenerator) GetCurrentTick() int {
	return tg.lastDeliveredTick
}

// GetFractionalTick returns the precise fractional tick position
func (tg *TickGenerator) GetFractionalTick() float64 {
	return tg.fractionalTick
}

// Reset resets the generator to initial state
func (tg *TickGenerator) Reset() {
	tg.currentSamples = 0
	tg.fractionalTick = 0.0
	tg.lastDeliveredTick = -1
	tg.tempoMapIndex = 0
	if len(tg.tempoMap) > 0 {
		tg.currentTempo = 60000000.0 / float64(tg.tempoMap[0].MicrosPerBeat)
	}
}
