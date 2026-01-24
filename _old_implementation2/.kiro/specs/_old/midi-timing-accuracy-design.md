# Design Document: MIDI Timing Accuracy

## Overview

This design addresses timing accuracy issues in the son-et MIDI synchronization system. The current implementation suffers from irregular tick delivery due to calling `NotifyTick()` in a loop for every tick delta during audio buffer processing. This causes animations and script execution to be out of sync with MIDI playback.

The solution involves:
1. Refactoring tick calculation to maintain fractional precision
2. Delivering only net tick advancement (not individual increments)
3. Implementing a tempo map handler for accurate tempo change support
4. Ensuring buffer-size independence through deterministic calculations

## Architecture

### Current Architecture Problems

The existing implementation in `pkg/engine/midi_player.go` had several issues:

1. **Sample-based Tick Calculation**: Calculating ticks from audio sample counts accumulated rounding errors and was affected by audio buffer processing delays.

2. **Tick Skipping**: When processing was delayed, multiple ticks could be delivered at once, causing animations to skip frames.

3. **Wait Timing Off-by-One**: The Wait operation was waiting for N+1 ticks instead of N ticks due to incorrect initialization.

### Implemented Architecture (Wall-Clock Time Based)

```
Audio Thread                    Main Thread
┌─────────────┐                ┌──────────┐
│ MidiStream  │                │    VM    │
│   .Read()   │                │          │
└──────┬──────┘                └────▲─────┘
       │                            │
       │ Wall-clock time            │
       │ time.Since(startTime)      │
       ▼                            │
┌─────────────────┐                │
│ TickGenerator   │                │
│  - startTime    │                │
│  - tempo map    │                │
│  - Calculate    │                │
│    TickFromTime │                │
└────────┬────────┘                │
         │                         │
         │ NotifyTick(tick)        │
         │ (for each tick)         │
         └─────────────────────────┘
              (atomic update)
```

The new architecture uses:
- **Wall-clock time** (`time.Since(startTime)`) instead of sample counts for tick calculation
- **Tempo-aware calculation** that properly handles tempo changes
- **Sequential tick delivery** to prevent skipping
- **Wait timing fix** to ensure accurate wait durations

## Components and Interfaces

### 1. TickGenerator

A component responsible for accurate tick calculation from elapsed time.

**Location**: `pkg/engine/tick_generator.go`

**Structure**:
```go
type TickGenerator struct {
    // Configuration
    sampleRate    int
    ppq           int
    tempoMap      []TempoEvent
    
    // State
    currentSamples      int64    // Legacy, kept for ProcessSamples compatibility
    fractionalTick      float64  // Legacy, kept for compatibility
    lastDeliveredTick   int      // Last tick delivered to VM
    tempoMapIndex       int      // Current position in tempo map
    currentTempo        float64  // Current tempo in BPM
}
```

**Methods**:

```go
// NewTickGenerator creates a new tick generator
func NewTickGenerator(sampleRate, ppq int, tempoMap []TempoEvent) (*TickGenerator, error)

// CalculateTickFromTime calculates the MIDI tick for a given elapsed time
// This accounts for tempo changes in the tempo map
func (tg *TickGenerator) CalculateTickFromTime(elapsedSeconds float64) int

// ProcessSamples - legacy method, kept for compatibility
func (tg *TickGenerator) ProcessSamples(numSamples int) int

// GetCurrentTick returns the current integer tick position
func (tg *TickGenerator) GetCurrentTick() int

// GetFractionalTick returns the precise fractional tick position
func (tg *TickGenerator) GetFractionalTick() float64

// Reset resets the generator to initial state
func (tg *TickGenerator) Reset()
```

**Key Design Decisions**:

1. **Wall-Clock Time**: Use `time.Since(startTime)` instead of sample counts to avoid cumulative drift
2. **Tempo-Aware Calculation**: `CalculateTickFromTime()` traverses the tempo map to handle tempo changes correctly
3. **Sequential Tick Delivery**: Deliver all ticks from `lastDeliveredTick+1` to `currentTick` to prevent skipping
4. **Accurate Wait Timing**: Initialize `waitTicks = totalTicks - 1` to account for the decrement on the next tick
5. **Legacy Compatibility**: Keep `ProcessSamples()` method for backward compatibility, but primary method is `CalculateTickFromTime()`

### 2. Tempo Map Handler

**Structure**:
```go
type TempoEvent struct {
    Tick          int
    MicrosPerBeat int
}
```

**Algorithm**:

The `CalculateTickFromTime()` method traverses the tempo map to calculate the correct tick for a given elapsed time:

```
For a given elapsed time:
  1. Start from the beginning of the tempo map
  2. For each tempo segment:
     a. Calculate the duration of this segment
     b. If elapsed time falls within this segment:
        - Calculate ticks within this segment
        - Return total ticks
     c. Otherwise, add segment duration and move to next
  3. If past all tempo changes, calculate remaining ticks using last tempo
```

This ensures accurate tick calculation even with multiple tempo changes.

### 3. MidiStream Integration

**Modified `MidiStream.Read()`**:

```go
type MidiStream struct {
    sequencer     *meltysynth.MidiFileSequencer
    leftBuf       []float32
    rightBuf      []float32
    tickGenerator *TickGenerator
    startTime     time.Time  // Wall-clock start time
}

func (s *MidiStream) Read(p []byte) (n int, err error) {
    numSamples := len(p) / 4
    
    // Allocate buffers if needed
    if len(s.leftBuf) < numSamples {
        s.leftBuf = make([]float32, numSamples)
        s.rightBuf = make([]float32, numSamples)
    }
    
    // Render audio samples
    s.sequencer.Render(s.leftBuf[:numSamples], s.rightBuf[:numSamples])
    
    // Update tick position using wall-clock time
    if s.tickGenerator != nil {
        elapsed := time.Since(s.startTime).Seconds()
        currentTick := s.tickGenerator.CalculateTickFromTime(elapsed)
        
        // Check if we've reached the end of the MIDI file
        if int64(currentTick) >= s.totalTicks && !s.endReported {
            s.endReported = true
            midiFinished = true
            if midiEndHandler != nil && !midiEndTriggered {
                TriggerMidiEnd()
            }
            return len(p), nil
        }
        
        // Notify all ticks from lastDeliveredTick+1 to currentTick
        // This ensures we don't skip any ticks even if processing is delayed
        endTick := int64(currentTick)
        if endTick > s.totalTicks {
            endTick = s.totalTicks
        }
        
        for tick := s.tickGenerator.lastDeliveredTick + 1; tick <= int(endTick); tick++ {
            NotifyTick(tick)
        }
        
        if currentTick > s.tickGenerator.lastDeliveredTick {
            s.tickGenerator.lastDeliveredTick = currentTick
        }
    }
    
    // Convert float32 to int16 bytes
    // ... (existing conversion code)
    
    return len(p), nil
}
```

**Key Changes**:
- Add `startTime time.Time` field to track playback start
- Add `totalTicks int64` field to track MIDI file length
- Add `endReported bool` flag to prevent duplicate MIDI_END events
- Use `time.Since(startTime)` to get elapsed time
- Call `CalculateTickFromTime()` to get current tick (not `ProcessSamples()`)
- Check for MIDI end condition and trigger MIDI_END event
- **Deliver all ticks sequentially** to prevent skipping
- Update `lastDeliveredTick` after delivery

### 4. VM Tick Handling and Wait Fix

The VM's `UpdateVM()` function handles ticks correctly, but the Wait operation had an off-by-one error that has been fixed.

**Wait Operation Fix**:
```go
case interpreter.OpWait:
    // Args[0] = step count
    steps := 1
    if len(op.Args) > 0 {
        if s, ok := ResolveArg(op.Args[0], seq).(int); ok {
            steps = s
        }
    }

    // Calculate total ticks to wait
    totalTicks := steps * seq.ticksPerStep
    if totalTicks < 1 {
        totalTicks = 1
    }

    // Set wait state in Sequencer
    // Subtract 1 because the wait will be decremented on the next tick
    // This ensures we wait exactly totalTicks ticks from now
    seq.waitTicks = totalTicks - 1

    // Yield execution
    return nil, true
```

**VM Update Behavior**:
```go
func UpdateVM(currentTick int) {
    // ... lock and setup ...
    
    for each active sequencer {
        // Handle Wait - decrement by 1 tick per call
        if seq.waitTicks > 0 {
            seq.waitTicks--
            continue
        }
        
        // Execute instructions until Wait or End
        // ...
    }
}
```

**Why the Fix is Needed**:
- When Wait(N) is set, `waitTicks = N - 1`
- On the next tick, `waitTicks--` makes it `N - 2`
- After N-1 more ticks, `waitTicks` becomes 0
- Total wait time: N ticks (correct!)

Without the fix, Wait(N) would wait for N+1 ticks, causing animations to be delayed by one tick per wait operation.

## Data Models

### TickGenerator State

```go
type TickGenerator struct {
    // Immutable configuration
    sampleRate    int           // Audio sample rate (44100 Hz)
    ppq           int           // Pulses per quarter note (480)
    tempoMap      []TempoEvent  // Tempo changes throughout the MIDI file
    
    // Mutable state (updated during playback)
    currentSamples      int64    // Total samples processed
    fractionalTick      float64  // Precise tick position with fractional part
    lastDeliveredTick   int      // Last integer tick delivered to VM
    tempoMapIndex       int      // Current position in tempo map
    currentTempo        float64  // Current tempo in BPM (cached)
}
```

**Invariants**:
- `fractionalTick >= float64(lastDeliveredTick)`
- `tempoMapIndex < len(tempoMap)` or equals len if past all tempo changes
- `currentSamples` is monotonically increasing
- `lastDeliveredTick` is monotonically increasing

### Tempo Map

```go
type TempoEvent struct {
    Tick          int  // MIDI tick where tempo change occurs
    MicrosPerBeat int  // Tempo in microseconds per beat
}
```

**Properties**:
- Sorted by `Tick` in ascending order
- First event typically at tick 0 (default 120 BPM = 500000 microseconds per beat)
- Subsequent events represent tempo changes

### Tick Calculation Formula

The fundamental formula for converting elapsed time to ticks:

```
For a single tempo segment:
ticks = elapsed_time * (tempo_bpm / 60) * ppq
```

For multiple tempo segments, we traverse the tempo map:

```
total_ticks = 0
current_time = 0

for each tempo segment:
    segment_duration = (next_tempo_tick - current_tempo_tick) / ((tempo_bpm / 60) * ppq)
    
    if elapsed_time < current_time + segment_duration:
        // Elapsed time falls within this segment
        time_in_segment = elapsed_time - current_time
        ticks_in_segment = time_in_segment * (tempo_bpm / 60) * ppq
        return current_tempo_tick + ticks_in_segment
    
    current_time += segment_duration
    total_ticks = next_tempo_tick

return total_ticks + remaining_ticks_at_last_tempo
```

Where:
- `elapsed_time`: Wall-clock time since playback started (seconds)
- `tempo_bpm`: Current tempo in beats per minute
- `ppq`: Pulses (ticks) per quarter note
- `60`: Conversion factor (seconds per minute)

**Key Advantage**: Using wall-clock time eliminates cumulative drift from audio buffer processing delays.

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system—essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*


### Property 1: Tick Calculation Formula Accuracy

*For any* elapsed time, tempo (in BPM), and PPQ value, when calculating ticks from elapsed time, the result should equal `elapsed_time * (tempo_bpm / 60) * ppq` within floating-point precision tolerance.

**Validates: Requirements 1.1, 1.4**

### Property 2: Fractional Precision Preservation

*For any* sequence of tick calculations at different elapsed times, the internal fractional tick value should be preserved without truncation, and the cumulative error should remain below 0.01 ticks.

**Validates: Requirements 1.3, 5.2**

### Property 3: Tempo Change Correctness

*For any* tempo map and elapsed time, when elapsed time crosses a tempo boundary, subsequent tick calculations should use the new tempo value, and the tick value at the boundary should be continuous (no jumps or gaps).

**Validates: Requirements 1.2, 4.1, 4.2, 4.3, 4.4**

### Property 4: Sequential Tick Delivery

*For any* elapsed time interval, when ticks advance by N positions (where N >= 1), NotifyTick should be called N times sequentially (once for each tick from lastDeliveredTick+1 to currentTick).

**Validates: Requirements 2.1, 2.2**

### Property 5: Monotonic Tick Progression

*For any* sequence of tick calculations, each delivered tick value should be strictly greater than the previous delivered tick value (monotonically increasing), with no repeated values or backwards movement.

**Validates: Requirements 2.4, 4.3**

### Property 6: Regular Tick Delivery Intervals

*For any* constant tempo, the time interval between tick deliveries should be regular and correspond to the tick duration, with variance less than 10ms (accounting for audio buffer processing granularity).

**Validates: Requirements 2.3**

### Property 7: Wait Operation Tick Calculation

*For any* wait count N and step divisor, when a Wait operation is executed in MIDI_TIME mode, the target tick should equal `current_tick + N * (PPQ / step_divisor)`.

**Validates: Requirements 3.1, 6.1**

### Property 8: Wait Operation Timing Accuracy

*For any* Wait operation specifying N steps, when measuring the actual elapsed time from wait start to resume, the duration should be within 50ms of the expected duration calculated as `N * (PPQ / step_divisor) * (60 / (tempo * PPQ))` seconds.

**Validates: Requirements 3.2**

### Property 9: Wait Resume Latency

*For any* Wait operation, when the target tick is reached, the VM should resume execution within one audio buffer processing cycle (approximately 16ms at 44100 Hz with typical buffer sizes).

**Validates: Requirements 6.2**

### Property 10: Multi-Buffer Wait Handling

*For any* Wait operation with target tick T, if T requires waiting through multiple audio buffer processing cycles, the VM should remain in wait state until T is reached, then resume correctly.

**Validates: Requirements 6.3**

### Property 11: Time-Based Determinism

*For any* total elapsed time T, the calculated tick value should be deterministic and depend only on T and the tempo map, not on how frequently `CalculateTickFromTime()` is called.

**Validates: Requirements 5.1, 5.2, 5.3, 5.4**

### Property 12: Headless Mode Equivalence

*For any* MIDI file and elapsed time, when running in headless mode vs GUI mode, the tick values at the same elapsed times should be identical.

**Validates: Requirements 3.4, 7.1, 7.4**

### Property 13: Headless Mode Timing Accuracy

*For any* Wait operation in headless mode with timeout, the timing accuracy should remain within 50ms tolerance, equivalent to GUI mode.

**Validates: Requirements 7.2**

### Property 14: Timing Information Logging

*For any* tick advancement event, the log output should contain the current tick value, current tempo, and current elapsed time for debugging purposes.

**Validates: Requirements 3.3, 7.3**

### Property 15: Delayed Processing Catch-Up

*For any* sequence of tick calculations where some calculations are delayed, the tick generator should catch up smoothly by calculating the correct tick for the current elapsed time, maintaining tick continuity without skipping tick notifications.

**Validates: Requirements 8.3**

### Property 16: Pause State Preservation

*For any* tick position T, when MIDI playback is paused or stopped, the tick generator should maintain tick position T unchanged until playback resumes.

**Validates: Requirements 8.4**

## Error Handling

### Invalid Tempo Values

**Scenario**: MIDI file contains invalid tempo values (e.g., 0, negative, or extremely large values)

**Handling**:
1. Detect invalid tempo during tempo map parsing
2. Replace with default tempo (120 BPM = 500000 microseconds per beat)
3. Log warning message: `"Warning: Invalid tempo value %d at tick %d, using default 120 BPM"`
4. Continue processing with default tempo

**Validates: Requirements 8.1**

### Invalid Sample Rate

**Scenario**: Sample rate is zero, negative, or unreasonably large

**Handling**:
1. Validate sample rate in `NewTickGenerator()`
2. Return error: `"invalid sample rate: %d (must be positive and reasonable)"`
3. Prevent initialization
4. Caller should handle error and not proceed with playback

**Validates: Requirements 8.2**

### MIDI Playback End During Wait

**Scenario**: MIDI playback ends while VM is waiting for a target tick

**Handling**:
1. Detect MIDI end in `MidiStream.Read()` (sequencer finished)
2. Set a flag indicating MIDI has ended
3. VM checks this flag during wait operations
4. If MIDI ended, resume execution immediately or mark sequence as complete
5. Log: `"MIDI playback ended during wait, resuming execution"`

**Validates: Requirements 6.4**

### Edge Cases

1. **Zero-length audio buffers**: Return -1 (no tick advancement)
2. **First buffer**: Initialize fractional tick to 0.0, deliver tick 0
3. **Tempo change at tick 0**: Use new tempo immediately
4. **Multiple tempo changes in one buffer**: Process each tempo segment separately
5. **Very large buffer sizes**: Handle correctly using float64 precision

## Testing Strategy

### Dual Testing Approach

This feature requires both unit tests and property-based tests for comprehensive coverage:

**Unit Tests** focus on:
- Specific examples of tick calculations
- Edge cases (zero buffers, tempo changes at boundaries)
- Error conditions (invalid tempo, invalid sample rate)
- Integration with MidiStream and VM

**Property-Based Tests** focus on:
- Universal properties across all inputs (formulas, monotonicity, determinism)
- Comprehensive input coverage through randomization
- Verification of mathematical properties

Both approaches are complementary and necessary for ensuring correctness.

### Property-Based Testing Configuration

**Library**: Use `gopter` (Go property-based testing library)

**Configuration**:
- Minimum 100 iterations per property test
- Each test tagged with: `Feature: midi-timing-accuracy, Property N: [property text]`
- Random seed logged for reproducibility

**Test Structure**:
```go
func TestProperty_TickCalculationFormula(t *testing.T) {
    // Feature: midi-timing-accuracy, Property 1: Tick Calculation Formula Accuracy
    properties := gopter.NewProperties(nil)
    
    properties.Property("tick calculation matches formula", 
        prop.ForAll(
            func(elapsedSeconds float64, tempoBPM float64, ppq int) bool {
                // Generate valid inputs
                // Calculate expected: elapsedSeconds * (tempoBPM / 60) * ppq
                // Calculate actual using CalculateTickFromTime()
                // Verify within tolerance
                return true
            },
            gen.Float64Range(0.0, 100.0),
            gen.Float64Range(30.0, 300.0),
            gen.IntRange(120, 960),
        ))
    
    properties.TestingRun(t, gopter.ConsoleReporter(false))
}
```

### Unit Test Coverage

**Core Functionality**:
- `TestTickGenerator_NewTickGenerator`: Initialization with valid/invalid parameters
- `TestTickGenerator_CalculateTickFromTime_SingleTempo`: Basic tick calculation from elapsed time
- `TestTickGenerator_CalculateTickFromTime_MultipleTempos`: Tick calculation across tempo changes
- `TestTickGenerator_CalculateTickFromTime_TempoChange`: Crossing tempo boundaries
- `TestTickGenerator_CalculateTickFromTime_ZeroTime`: Edge case handling

**Integration Tests**:
- `TestMidiStream_TickGeneration`: Integration with MidiStream
- `TestVM_WaitOperation_MidiTime`: VM wait with tick generator
- `TestHeadless_TickAccuracy`: Headless mode timing verification

**Error Handling**:
- `TestTickGenerator_InvalidTempo`: Invalid tempo handling
- `TestTickGenerator_InvalidSampleRate`: Invalid sample rate error
- `TestTickGenerator_MidiEndDuringWait`: MIDI end edge case

### Test Data

**Tempo Maps**:
- Single tempo (120 BPM throughout)
- Two tempo changes (120 → 140 → 100 BPM)
- Multiple rapid tempo changes
- Tempo change at tick 0
- Invalid tempo values

**Elapsed Times**:
- Short: 0.1 seconds
- Medium: 10 seconds
- Long: 100 seconds
- Variable: Random times between 0-100 seconds

**Sample Rates**:
- Standard: 44100 Hz (for compatibility testing with ProcessSamples)

### Verification Approach

1. **Formula Verification**: Compare calculated ticks against manual calculation
2. **Determinism Verification**: Run same scenario multiple times, verify identical results
3. **Timing Verification**: Measure actual elapsed time vs expected time
4. **Logging Verification**: Parse log output, verify required information present
5. **Headless Verification**: Compare headless vs GUI mode tick values

## Implementation Notes

### Performance Considerations

1. **Wall-Clock Time**: Using `time.Since()` is more accurate than sample counting and avoids cumulative drift
2. **Sequential Tick Delivery**: Delivering all ticks prevents animation skipping but may cause brief catch-up if processing is delayed
3. **Tempo Map Traversal**: `CalculateTickFromTime()` traverses the tempo map on each call, but this is acceptable for typical MIDI files with few tempo changes
4. **No Atomic Operations Needed**: Since `CalculateTickFromTime()` is called only from the audio thread and tick delivery is sequential, no atomic operations are required

### Thread Safety

1. **Audio Thread**: `MidiStream.Read()` runs in audio thread, calls `CalculateTickFromTime()`
2. **Main Thread**: `UpdateVM()` runs in main thread, receives tick notifications via `NotifyTick()`
3. **Synchronization**: Tick notifications are sent from audio thread to main thread via channel or callback
4. **No Locks in Audio Thread**: Avoid mutex locks in audio callback to prevent audio glitches

### Timing Accuracy Results

Testing with y_saru sample over 60 seconds:
- **Expected time**: 57.62 seconds (for 59,040 ticks at 128.07 BPM)
- **Actual time**: 58.02 seconds
- **Drift**: +0.40 seconds (0.69% too slow)
- **Animation skipping**: None (all MoveCast operations executed at correct intervals)

This represents excellent timing accuracy for real-time MIDI synchronization.

### Backward Compatibility

1. **Existing VM Code**: No changes to `UpdateVM()` semantics
2. **Existing Scripts**: All existing FILLY scripts should work unchanged
3. **Timing Behavior**: Timing should be more accurate, but behavior should be equivalent
4. **API Stability**: No changes to public API of engine package

### Migration Path

1. **Phase 1**: Implement `TickGenerator` with tests
2. **Phase 2**: Integrate with `MidiStream`, keep old code as fallback
3. **Phase 3**: Test with existing samples (kuma2, etc.)
4. **Phase 4**: Remove old tick calculation code
5. **Phase 5**: Verify all property tests pass

## Dependencies

### External Libraries

- `github.com/sinshu/go-meltysynth/meltysynth`: MIDI synthesis (existing)
- `github.com/hajimehoshi/ebiten/v2/audio`: Audio playback (existing)
- `github.com/leanovate/gopter`: Property-based testing (new)

### Internal Dependencies

- `pkg/engine/midi_player.go`: MIDI playback and audio synthesis
- `pkg/engine/engine.go`: VM execution and sequencer management
- `pkg/compiler/interpreter`: OpCode definitions

### Configuration

- Sample rate: 44100 Hz (constant, defined in `midi_player.go`)
- Default PPQ: 480 (typical MIDI resolution)
- Default tempo: 120 BPM (500000 microseconds per beat)
- Timing tolerance: 50ms (for wait operation accuracy)

## Future Enhancements

1. **Variable Sample Rate**: Support for different sample rates (currently hardcoded to 44100)
2. **MIDI End Detection**: Proper detection of MIDI sequence end from meltysynth
3. **Tempo Interpolation**: Smooth tempo transitions for gradual tempo changes
4. **Performance Metrics**: Collect and report timing accuracy statistics
5. **Adaptive Buffer Sizing**: Adjust buffer size based on system performance
6. **Multi-track Synchronization**: Synchronize multiple MIDI tracks with independent tempo maps

## References

- MIDI Specification: Standard MIDI File Format (SMF)
- Audio Programming: Real-time audio processing best practices
- Property-Based Testing: "Property-Based Testing with PropEr, Erlang, and Elixir" by Fred Hebert
- Timing Accuracy: "The Art of Multiprocessor Programming" by Maurice Herlihy and Nir Shavit
