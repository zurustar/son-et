# MIDI Timing Architecture

## Overview

This document describes the MIDI timing synchronization system in son-et, which ensures accurate synchronization between MIDI playback and script execution (animations, timing-based logic).

## Architecture Summary

The MIDI timing system uses **wall-clock time** as the primary timing source instead of audio sample counts. This eliminates cumulative drift and provides deterministic, accurate tick calculation even with variable audio buffer sizes and processing delays.

### Key Components

1. **TickGenerator** (`tick_generator.go`): Calculates MIDI ticks from elapsed time
2. **MidiStream** (`midi_player.go`): Integrates tick generation with audio playback
3. **VM Tick Handling** (`engine.go`): Processes tick notifications and executes script logic

## Wall-Clock Time Based Timing

### Why Wall-Clock Time?

The previous implementation used audio sample counts to calculate ticks:
```
ticks = (samples_processed * tempo_bpm * ppq) / (sample_rate * 60)
```

**Problems with sample-based calculation:**
- Audio buffer processing delays cause cumulative drift
- Variable buffer sizes affect timing accuracy
- Rounding errors accumulate over time
- Not deterministic (depends on buffer processing patterns)

**Solution: Wall-clock time**
```
elapsed_time = time.Since(startTime).Seconds()
ticks = CalculateTickFromTime(elapsed_time)
```

**Advantages:**
- No cumulative drift from audio processing delays
- Deterministic: same elapsed time always produces same tick
- Independent of audio buffer size
- Accurate across tempo changes

### Implementation

The `MidiStream.Read()` method (called by the audio system) performs:

```go
// Get elapsed time since playback started
elapsed := time.Since(s.startTime).Seconds()

// Calculate current tick from elapsed time
currentTick := s.tickGenerator.CalculateTickFromTime(elapsed)

// Deliver all ticks from lastDeliveredTick+1 to currentTick
for tick := lastDeliveredTick + 1; tick <= currentTick; tick++ {
    NotifyTick(tick)
}
```

## Tempo-Aware Tick Calculation

### The Challenge

MIDI files can contain multiple tempo changes throughout the song. The tick calculation must account for these changes to maintain accurate synchronization.

### Algorithm

The `CalculateTickFromTime()` method traverses the tempo map:

```
1. Start from the beginning of the tempo map
2. For each tempo segment:
   a. Calculate the duration of this segment in seconds
   b. Check if elapsed time falls within this segment
   c. If yes: calculate ticks within this segment and return total
   d. If no: accumulate time and move to next segment
3. If past all tempo changes, calculate remaining ticks using last tempo
```

### Formula

For a single tempo segment:
```
ticks = elapsed_time * (tempo_bpm / 60) * ppq
```

Where:
- `elapsed_time`: Time in seconds since playback started
- `tempo_bpm`: Current tempo in beats per minute
- `ppq`: Pulses (ticks) per quarter note (typically 480)
- `60`: Conversion factor (seconds per minute)

### Example

Given a MIDI file with tempo changes:
- Tick 0-1000: 120 BPM
- Tick 1000-2000: 140 BPM
- Tick 2000+: 100 BPM

To calculate the tick at elapsed time = 15 seconds:

1. **Segment 1 (120 BPM):**
   - Duration: 1000 ticks / (120/60 * 480) = 1.042 seconds
   - Elapsed time (15s) > segment duration (1.042s)
   - Move to next segment, accumulate time

2. **Segment 2 (140 BPM):**
   - Duration: 1000 ticks / (140/60 * 480) = 0.893 seconds
   - Elapsed time (15s) > accumulated time (1.935s)
   - Move to next segment, accumulate time

3. **Segment 3 (100 BPM):**
   - Remaining time: 15 - 1.935 = 13.065 seconds
   - Ticks in segment: 13.065 * (100/60 * 480) = 10,452 ticks
   - Total: 2000 + 10,452 = 12,452 ticks

## Sequential Tick Delivery

### The Problem

If audio processing is delayed (e.g., CPU spike), multiple ticks may advance in a single audio buffer. If we only notify the latest tick, intermediate ticks are skipped, causing animations to skip frames.

### Solution

The `MidiStream.Read()` method delivers **all ticks sequentially**:

```go
// Deliver all ticks from lastDeliveredTick+1 to currentTick
for tick := lastDeliveredTick + 1; tick <= currentTick; tick++ {
    NotifyTick(tick)
}
```

**Example:**
- Last delivered tick: 100
- Current tick: 105
- Delivers: 101, 102, 103, 104, 105

This ensures that:
- No ticks are skipped
- Animations execute all frames
- Script logic sees every tick
- Synchronization remains accurate even with processing delays

## MIDI End Detection

### The Challenge

The VM needs to know when MIDI playback has finished so it can:
- Resume execution if waiting for MIDI to complete
- Trigger MIDI_END event handlers
- Terminate the program if all sequences are finished

### Implementation

The `MidiStream.Read()` method detects MIDI end by comparing ticks:

```go
if int64(currentTick) >= s.totalTicks && !s.endReported {
    s.endReported = true
    midiFinished = true  // Set global flag
    TriggerMidiEnd()     // Trigger MIDI_END event
    return len(p), nil   // Stop sending ticks
}
```

The VM checks the `midiFinished` flag during wait operations:

```go
if midiFinished {
    // Resume execution or terminate program
}
```

## Wait Operation Timing

### The Challenge

FILLY scripts use `Wait(N)` to pause execution for N steps. The VM must calculate the target tick and resume execution when that tick is reached.

### Implementation

When `Wait(N)` is executed:

```go
// Calculate total ticks to wait
totalTicks := N * ticksPerStep

// Set wait state (subtract 1 for decrement on next tick)
seq.waitTicks = totalTicks - 1
```

On each tick, the VM decrements the wait counter:

```go
if seq.waitTicks > 0 {
    seq.waitTicks--
    continue  // Don't execute instructions yet
}
```

**Why subtract 1?**

The wait counter is decremented on the **next** tick after Wait is set. Without the -1 adjustment, Wait(N) would wait for N+1 ticks instead of N ticks.

**Example:**
- Wait(2) is executed at tick 100
- Set waitTicks = 2 - 1 = 1
- Tick 101: waitTicks-- → 0
- Tick 102: waitTicks = 0, resume execution
- Total wait: 2 ticks ✓

## Thread Safety

### Audio Thread vs Main Thread

- **Audio Thread**: Runs `MidiStream.Read()`, calculates ticks, sends notifications
- **Main Thread**: Runs `UpdateVM()`, receives tick notifications, executes script logic

### Synchronization

Tick notifications are sent from audio thread to main thread via `NotifyTick()`:

```go
func NotifyTick(tick int) {
    // Thread-safe notification mechanism
    // (implementation uses channels or atomic operations)
}
```

**Important:** No locks are used in the audio thread to avoid audio glitches.

## Timing Accuracy Results

Testing with real MIDI files shows excellent timing accuracy:

### y_saru Sample (60 seconds)
- **Expected time**: 57.62 seconds (59,040 ticks at 128.07 BPM)
- **Actual time**: 58.02 seconds
- **Drift**: +0.40 seconds (0.69% too slow)
- **Animation skipping**: None (all MoveCast operations at correct intervals)

This represents production-quality timing accuracy for real-time MIDI synchronization.

## Key Methods Reference

### TickGenerator

#### `CalculateTickFromTime(elapsedSeconds float64) int`
**PRIMARY METHOD - Used in production**

Calculates the current MIDI tick from elapsed wall-clock time, properly handling tempo changes by traversing the tempo map.

**Used by:** `MidiStream.Read()`

#### `ProcessSamples(numSamples int) int`
**LEGACY METHOD - For testing only**

Calculates tick advancement from audio sample count. Kept for backward compatibility with property-based tests but NOT used in production code.

**Why not used:** Sample-based calculation can accumulate drift from audio buffer processing delays. Wall-clock time is more accurate.

### MidiStream

#### `Read(p []byte) (n int, err error)`
**PRIMARY AUDIO CALLBACK**

Called by the Ebiten audio system to fill audio buffers. Performs:
1. Renders audio samples from MIDI synthesizer
2. Calculates current tick from wall-clock time
3. Delivers tick notifications to VM
4. Detects MIDI playback end

## Legacy Code

The following code is kept for backward compatibility but is not actively used:

### Global Variables (midi_player.go)
```go
var (
    globalTempoMap []TempoEvent  // UNUSED: Legacy tempo map
    globalPPQ      int = 480     // UNUSED: Legacy PPQ
    currentSamples int64         // UNUSED: Legacy sample counter
)
```

### Functions (midi_player.go)
```go
func StartConductor(tempoMap []TempoEvent, ppq int)
// LEGACY: Old tick generation system, no longer used

func StartConductorStub()
// LEGACY: Placeholder function, no longer used
```

### TickGenerator Fields (tick_generator.go)
```go
type TickGenerator struct {
    currentSamples    int64   // LEGACY: Only used by ProcessSamples
    fractionalTick    float64 // LEGACY: Only used by ProcessSamples
    tempoMapIndex     int     // LEGACY: Only used by ProcessSamples
    currentTempo      float64 // LEGACY: Only used by ProcessSamples
    // ...
}
```

These fields are only used by the `ProcessSamples()` method, which is kept for testing but not used in production.

## Future Enhancements

Potential improvements to the MIDI timing system:

1. **Variable Sample Rate**: Support for different sample rates (currently hardcoded to 44100 Hz)
2. **Tempo Interpolation**: Smooth tempo transitions for gradual tempo changes
3. **Performance Metrics**: Collect and report timing accuracy statistics
4. **Adaptive Buffer Sizing**: Adjust buffer size based on system performance
5. **Multi-track Synchronization**: Synchronize multiple MIDI tracks with independent tempo maps

## References

- MIDI Specification: Standard MIDI File Format (SMF)
- Audio Programming: Real-time audio processing best practices
- Property-Based Testing: Comprehensive testing of timing properties
- Design Document: `.kiro/specs/midi-timing-accuracy/design.md`
- Requirements Document: `.kiro/specs/midi-timing-accuracy/requirements.md`
