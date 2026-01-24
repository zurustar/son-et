package engine

import (
	"fmt"
	"time"
)

// TickGenerator calculates MIDI ticks from elapsed wall-clock time with tempo-aware precision.
//
// # Architecture Overview
//
// The TickGenerator uses wall-clock time (time.Since(startTime)) as the primary timing source
// instead of audio sample counts. This eliminates cumulative drift caused by audio buffer
// processing delays and provides deterministic, accurate tick calculation.
//
// # Key Methods
//
//   - CalculateTickFromTime(elapsedSeconds): PRIMARY method for production use
//     Calculates the current MIDI tick from elapsed wall-clock time, properly handling
//     tempo changes by traversing the tempo map. This is the method used by MidiStream.Read().
//
//   - ProcessSamples(numSamples): LEGACY method for backward compatibility
//     Calculates tick advancement from audio sample count. Kept for testing purposes only.
//     NOT used in production code. May accumulate drift over time.
//
// # Timing Algorithm
//
// The CalculateTickFromTime method implements a tempo-aware tick calculation:
//
//  1. Traverse the tempo map to find which tempo segment contains the elapsed time
//  2. For each tempo segment before the current one:
//     - Calculate the duration of that segment in seconds
//     - Accumulate the time and tick count
//  3. For the current tempo segment:
//     - Calculate how much time has elapsed within this segment
//     - Convert that time to ticks using: ticks = time * (tempo_bpm / 60) * ppq
//  4. Return the total tick count
//
// This ensures accurate synchronization even with multiple tempo changes.
//
// # Thread Safety
//
// TickGenerator is designed to be called from the audio thread (MidiStream.Read).
// No locks are used to avoid audio glitches. Tick notifications are sent to the
// main thread via NotifyTick() which handles cross-thread communication.
type TickGenerator struct {
	// Configuration (immutable)
	sampleRate int          // Audio sample rate (typically 44100 Hz)
	ppq        int          // Pulses Per Quarter note - MIDI timing resolution (typically 480)
	tempoMap   []TempoEvent // Tempo changes throughout the MIDI file

	// State (mutable during playback)
	currentSamples    int64   // LEGACY: Total samples processed (only used by ProcessSamples)
	fractionalTick    float64 // LEGACY: Precise tick position with fractional part (only used by ProcessSamples)
	lastDeliveredTick int     // Last integer tick delivered to VM
	tempoMapIndex     int     // LEGACY: Current position in tempo map (only used by ProcessSamples)
	currentTempo      float64 // LEGACY: Current tempo in BPM cached (only used by ProcessSamples)
	lastLoggedTick    int     // Last tick that was logged (for periodic logging every 100 ticks)
}

// TempoEvent represents a tempo change at a specific MIDI tick.
//
// MIDI files can contain multiple tempo changes throughout the song.
// Each TempoEvent specifies when (at which tick) a tempo change occurs
// and what the new tempo is (in microseconds per beat).
//
// The tempo map is a sorted list of TempoEvents used by TickGenerator
// to calculate accurate tick positions across tempo changes.
type TempoEvent struct {
	Tick          int // MIDI tick where this tempo change occurs
	MicrosPerBeat int // Tempo in microseconds per beat (e.g., 500000 = 120 BPM)
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
		lastLoggedTick:    -100, // Initialize to -100 so first log happens at tick 0
	}

	return tg, nil
}

// ProcessSamples updates the tick position based on samples rendered.
//
// **LEGACY METHOD - FOR TESTING ONLY**
//
// This method calculates tick advancement from audio sample count.
// It is kept for backward compatibility with existing property-based tests
// but is NOT used in production code.
//
// **Why not used in production:**
// - Sample-based calculation can accumulate drift from audio buffer processing delays
// - Wall-clock time (CalculateTickFromTime) is more accurate and deterministic
//
// **Use CalculateTickFromTime instead for production code.**
//
// Returns the new tick value if it has advanced, or -1 if no change.
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

// CalculateTickFromTime calculates the MIDI tick for a given elapsed time.
//
// **PRIMARY METHOD - USED IN PRODUCTION**
//
// This is the main method used by MidiStream.Read() to calculate the current
// MIDI tick position from wall-clock time. It properly accounts for tempo
// changes by traversing the tempo map.
//
// # Algorithm
//
// The method implements a tempo-aware tick calculation:
//
//  1. Start from the beginning of the tempo map
//  2. For each tempo segment:
//     a. Calculate the duration of this segment in seconds
//     b. Check if elapsed time falls within this segment
//     c. If yes: calculate ticks within this segment and return total
//     d. If no: accumulate time and move to next segment
//  3. If past all tempo changes, calculate remaining ticks using last tempo
//
// # Formula
//
// For a single tempo segment:
//
//	ticks = elapsed_time * (tempo_bpm / 60) * ppq
//
// Where:
//   - elapsed_time: Time in seconds since playback started
//   - tempo_bpm: Current tempo in beats per minute
//   - ppq: Pulses (ticks) per quarter note
//   - 60: Conversion factor (seconds per minute)
//
// # Advantages over Sample-Based Calculation
//
// - No cumulative drift from audio buffer processing delays
// - Deterministic: same elapsed time always produces same tick
// - Independent of audio buffer size
// - Accurate across tempo changes
//
// # Thread Safety
//
// This method is called from the audio thread (MidiStream.Read).
// It does not use locks to avoid audio glitches.
func (tg *TickGenerator) CalculateTickFromTime(elapsedSeconds float64) int {
	if len(tg.tempoMap) == 0 {
		return 0
	}

	currentTime := 0.0
	currentTick := 0
	currentTempo := 0.0

	for i := 0; i < len(tg.tempoMap); i++ {
		// Get tempo for this segment
		tempo := 60000000.0 / float64(tg.tempoMap[i].MicrosPerBeat)
		currentTempo = tempo

		// Determine the end tick for this tempo segment
		endTick := 0
		if i+1 < len(tg.tempoMap) {
			endTick = tg.tempoMap[i+1].Tick
		} else {
			// Last segment - calculate how many ticks we can fit in remaining time
			remainingTime := elapsedSeconds - currentTime
			ticksInSegment := int(remainingTime * (tempo / 60.0) * float64(tg.ppq))
			finalTick := tg.tempoMap[i].Tick + ticksInSegment

			// Log every 100 ticks
			if finalTick-tg.lastLoggedTick >= 100 {
				fmt.Printf("[%s] TickGenerator: tick=%d, tempo=%.2f BPM, elapsed=%.3fs\n",
					time.Now().Format("15:04:05.000"), finalTick, currentTempo, elapsedSeconds)
				tg.lastLoggedTick = (finalTick / 100) * 100 // Round down to nearest 100
			}

			return finalTick
		}

		// Calculate time duration of this tempo segment
		ticksInSegment := endTick - tg.tempoMap[i].Tick
		segmentDuration := float64(ticksInSegment) / (tempo / 60.0 * float64(tg.ppq))

		// Check if elapsed time falls within this segment
		if currentTime+segmentDuration > elapsedSeconds {
			// We're in this segment
			timeInSegment := elapsedSeconds - currentTime
			ticksInSegment := int(timeInSegment * (tempo / 60.0) * float64(tg.ppq))
			finalTick := tg.tempoMap[i].Tick + ticksInSegment

			// Log every 100 ticks
			if finalTick-tg.lastLoggedTick >= 100 {
				fmt.Printf("[%s] TickGenerator: tick=%d, tempo=%.2f BPM, elapsed=%.3fs\n",
					time.Now().Format("15:04:05.000"), finalTick, currentTempo, elapsedSeconds)
				tg.lastLoggedTick = (finalTick / 100) * 100 // Round down to nearest 100
			}

			return finalTick
		}

		// Move to next segment
		currentTime += segmentDuration
		currentTick = endTick
	}

	// Log every 100 ticks
	if currentTick-tg.lastLoggedTick >= 100 {
		fmt.Printf("[%s] TickGenerator: tick=%d, tempo=%.2f BPM, elapsed=%.3fs\n",
			time.Now().Format("15:04:05.000"), currentTick, currentTempo, elapsedSeconds)
		tg.lastLoggedTick = (currentTick / 100) * 100 // Round down to nearest 100
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
	tg.lastLoggedTick = -100 // Reset to -100 so first log happens at tick 0
	if len(tg.tempoMap) > 0 {
		tg.currentTempo = 60000000.0 / float64(tg.tempoMap[0].MicrosPerBeat)
	}
}
