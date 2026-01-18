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

	// Find the last tempo event at tick 0 (in case there are multiple)
	initialTempoIndex := 0
	for i := 1; i < len(tempoMap); i++ {
		if tempoMap[i].Tick == 0 {
			initialTempoIndex = i
		} else {
			break
		}
	}

	// Calculate initial tempo in BPM
	initialTempoBPM := 60000000.0 / float64(tempoMap[initialTempoIndex].MicrosPerBeat)

	// Debug: log tempo map
	fmt.Printf("TickGenerator: Creating with sampleRate=%d, PPQ=%d\n", sampleRate, ppq)
	fmt.Printf("TickGenerator: tempoMap has %d events\n", len(tempoMap))
	for i, ev := range tempoMap {
		bpm := 60000000.0 / float64(ev.MicrosPerBeat)
		fmt.Printf("  Event %d: Tick=%d, MicrosPerBeat=%d, BPM=%.2f\n",
			i, ev.Tick, ev.MicrosPerBeat, bpm)
	}
	fmt.Printf("TickGenerator: Using initial tempo from event %d: %.2f BPM\n", initialTempoIndex, initialTempoBPM)

	tg := &TickGenerator{
		sampleRate:        sampleRate,
		ppq:               ppq,
		tempoMap:          tempoMap,
		currentSamples:    0,
		fractionalTick:    0.0,
		lastDeliveredTick: 0,
		tempoMapIndex:     initialTempoIndex,
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

// CalculateTickFromTime calculates the MIDI tick for a given elapsed time
// This accounts for tempo changes in the tempo map
func (tg *TickGenerator) CalculateTickFromTime(elapsedSeconds float64) int {
	if len(tg.tempoMap) == 0 {
		return 0
	}

	currentTime := 0.0
	currentTick := 0

	for i := 0; i < len(tg.tempoMap); i++ {
		// Get tempo for this segment
		tempo := 60000000.0 / float64(tg.tempoMap[i].MicrosPerBeat)

		// Determine the end tick for this tempo segment
		endTick := 0
		if i+1 < len(tg.tempoMap) {
			endTick = tg.tempoMap[i+1].Tick
		} else {
			// Last segment - calculate how many ticks we can fit in remaining time
			remainingTime := elapsedSeconds - currentTime
			ticksInSegment := int(remainingTime * (tempo / 60.0) * float64(tg.ppq))
			return tg.tempoMap[i].Tick + ticksInSegment
		}

		// Calculate time duration of this tempo segment
		ticksInSegment := endTick - tg.tempoMap[i].Tick
		segmentDuration := float64(ticksInSegment) / (tempo / 60.0 * float64(tg.ppq))

		// Check if elapsed time falls within this segment
		if currentTime+segmentDuration > elapsedSeconds {
			// We're in this segment
			timeInSegment := elapsedSeconds - currentTime
			ticksInSegment := int(timeInSegment * (tempo / 60.0) * float64(tg.ppq))
			return tg.tempoMap[i].Tick + ticksInSegment
		}

		// Move to next segment
		currentTime += segmentDuration
		currentTick = endTick
	}

	return currentTick
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
