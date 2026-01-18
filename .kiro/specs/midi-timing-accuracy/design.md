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

The existing implementation in `pkg/engine/midi_player.go` has several issues:

1. **Inefficient Tick Delivery**: `ProcessSamples()` calls `NotifyTick()` in a loop for each tick delta:
   ```go
   delta := tick - lastTick
   for i := 0; i < delta; i++ {
       NotifyTick(lastTick + i + 1)
   }
   ```
   This causes irregular delivery patterns and performance issues.

2. **Recalculation from Scratch**: `CalculateTickFromTime()` recalculates the entire tempo map traversal on every call, which is inefficient.

3. **Loss of Fractional Precision**: Integer tick calculations accumulate rounding errors over time.

4. **Unclear VM Semantics**: The VM's `UpdateVM()` function is called once per tick, but `waitTicks` is decremented by 1 per call, creating confusion about tick advancement.

### Proposed Architecture

```
Audio Thread                    Main Thread
┌─────────────┐                ┌──────────┐
│ MidiStream  │                │    VM    │
│   .Read()   │                │          │
└──────┬──────┘                └────▲─────┘
       │                            │
       │ ProcessSamples(n)          │
       ▼                            │
┌─────────────────┐                │
│ TickGenerator   │                │
│  - samples      │                │
│  - fractional   │                │
│  - tempo map    │                │
└────────┬────────┘                │
         │                         │
         │ NotifyTick(tick)        │
         └─────────────────────────┘
              (atomic update)
```

The new architecture introduces a `TickGenerator` component that:
- Maintains fractional tick precision internally
- Tracks current position in the tempo map
- Delivers only net tick advancement
- Provides deterministic, buffer-size-independent calculations

## Components and Interfaces

### 1. TickGenerator

A new component responsible for accurate tick calculation and delivery.

**Location**: `pkg/engine/tick_generator.go`

**Structure**:
```go
type TickGenerator struct {
    // Configuration
    sampleRate    int
    ppq           int
    tempoMap      []TempoEvent
    
    // State
    currentSamples      int64
    fractionalTick      float64
    lastDeliveredTick   int
    tempoMapIndex       int
    currentTempo        float64  // Current tempo in BPM
}
```

**Methods**:

```go
// NewTickGenerator creates a new tick generator
func NewTickGenerator(sampleRate, ppq int, tempoMap []TempoEvent) (*TickGenerator, error)

// ProcessSamples updates the tick position based on samples rendered
// Returns the new tick value if it has advanced, or -1 if no change
func (tg *TickGenerator) ProcessSamples(numSamples int) int

// GetCurrentTick returns the current integer tick position
func (tg *TickGenerator) GetCurrentTick() int

// GetFractionalTick returns the precise fractional tick position
func (tg *TickGenerator) GetFractionalTick() float64

// Reset resets the generator to initial state
func (tg *TickGenerator) Reset()
```

**Key Design Decisions**:

1. **Fractional Precision**: Maintain `fractionalTick` as float64 to prevent cumulative rounding errors
2. **Single Notification**: `ProcessSamples()` returns only the new tick value, not a delta
3. **Tempo Map Caching**: Cache current tempo map index to avoid re-traversing from the beginning
4. **Deterministic Calculation**: Use pure functions based on sample count for reproducibility

### 2. Tempo Map Handler

**Structure**:
```go
type TempoEvent struct {
    Tick          int
    MicrosPerBeat int
}
```

**Algorithm**:

The tempo map handler maintains an index into the tempo map and updates it incrementally:

```
For each audio buffer:
  1. Start from current tempo map index
  2. Check if we've crossed into next tempo segment
  3. If yes, update current tempo and index
  4. Calculate tick advancement using current tempo
  5. Add to fractional tick accumulator
  6. Return integer part if it changed
```

This avoids re-traversing the entire tempo map on every call.

### 3. MidiStream Integration

**Modified `MidiStream.Read()`**:

```go
func (s *MidiStream) Read(p []byte) (n int, err error) {
    numSamples := len(p) / 4
    
    // Allocate buffers if needed
    if len(s.leftBuf) < numSamples {
        s.leftBuf = make([]float32, numSamples)
        s.rightBuf = make([]float32, numSamples)
    }
    
    // Render audio samples
    s.sequencer.Render(s.leftBuf[:numSamples], s.rightBuf[:numSamples])
    
    // Update tick position
    if s.tickGenerator != nil {
        newTick := s.tickGenerator.ProcessSamples(numSamples)
        if newTick >= 0 {
            NotifyTick(newTick)
        }
    }
    
    // Convert float32 to int16 bytes
    // ... (existing conversion code)
    
    return len(p), nil
}
```

**Key Changes**:
- Add `tickGenerator *TickGenerator` field to `MidiStream`
- Call `ProcessSamples()` once per buffer
- Only call `NotifyTick()` if tick actually advanced
- Pass the new tick value, not a delta

### 4. VM Tick Handling

The VM's `UpdateVM()` function already handles ticks correctly - it's called once per tick and decrements `waitTicks` by 1. No changes needed to VM semantics.

**Current VM Behavior** (no changes):
```go
func UpdateVM(currentTick int) {
    // ... lock and setup ...
    
    for each active sequencer {
        // Handle Wait
        if seq.waitTicks > 0 {
            seq.waitTicks--
            continue
        }
        
        // Execute instructions until Wait or End
        // ...
    }
}
```

The key insight is that `NotifyTick()` should be called once with the current tick value, and `UpdateVM()` will be called once per tick by the main loop.

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

The fundamental formula for converting samples to ticks:

```
ticks = (samples * tempo_bpm * ppq) / (sample_rate * 60)
```

Where:
- `samples`: Number of audio samples processed
- `tempo_bpm`: Current tempo in beats per minute
- `ppq`: Pulses (ticks) per quarter note
- `sample_rate`: Audio sample rate (44100 Hz)
- `60`: Conversion factor (seconds per minute)

**Derivation**:
```
time_seconds = samples / sample_rate
beats = time_seconds * (tempo_bpm / 60)
ticks = beats * ppq
      = (samples / sample_rate) * (tempo_bpm / 60) * ppq
      = (samples * tempo_bpm * ppq) / (sample_rate * 60)
```

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system—essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*


### Property 1: Tick Calculation Formula Accuracy

*For any* sample count, tempo (in BPM), PPQ value, and sample rate, when calculating ticks from samples, the result should equal `(samples * tempo * PPQ) / (sample_rate * 60)` within floating-point precision tolerance.

**Validates: Requirements 1.1, 1.4**

### Property 2: Fractional Precision Preservation

*For any* sequence of audio buffer processing calls, the internal fractional tick value should be preserved without truncation, and the cumulative error between the fractional tick and the sum of individual buffer tick calculations should remain below 0.01 ticks.

**Validates: Requirements 1.3, 5.2**

### Property 3: Tempo Change Correctness

*For any* tempo map and sample position, when samples cross a tempo boundary, subsequent tick calculations should use the new tempo value, and the tick value at the boundary should be continuous (no jumps or gaps).

**Validates: Requirements 1.2, 4.1, 4.2, 4.3, 4.4**

### Property 4: Single Tick Delivery

*For any* audio buffer size, when ProcessSamples is called and ticks advance by N positions (where N >= 1), NotifyTick should be called exactly once with the new tick value, not N times.

**Validates: Requirements 2.1, 2.2**

### Property 5: Monotonic Tick Progression

*For any* sequence of ProcessSamples calls, each delivered tick value should be strictly greater than the previous delivered tick value (monotonically increasing), with no repeated values or backwards movement.

**Validates: Requirements 2.4, 4.3**

### Property 6: Regular Tick Delivery Intervals

*For any* constant tempo and buffer size, the time interval between tick deliveries should be regular and correspond to the buffer processing rate, with variance less than one buffer period.

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

### Property 11: Buffer Size Determinism

*For any* total sample count S, when processing S samples using different buffer size patterns (e.g., all at once vs. many small buffers), the final tick value and fractional tick should be identical within floating-point precision tolerance.

**Validates: Requirements 5.1, 5.2, 5.3, 5.4**

### Property 12: Headless Mode Equivalence

*For any* MIDI file and sample count, when running in headless mode vs GUI mode, the tick values at the same sample positions should be identical.

**Validates: Requirements 3.4, 7.1, 7.4**

### Property 13: Headless Mode Timing Accuracy

*For any* Wait operation in headless mode with timeout, the timing accuracy should remain within 50ms tolerance, equivalent to GUI mode.

**Validates: Requirements 7.2**

### Property 14: Timing Information Logging

*For any* tick advancement event, the log output should contain the current tick value, current tempo, and current sample position for debugging purposes.

**Validates: Requirements 3.3, 7.3**

### Property 15: Delayed Processing Catch-Up

*For any* sequence of audio buffer processing calls where some calls are delayed, the tick generator should catch up smoothly by processing the accumulated samples, maintaining tick continuity without skipping tick values.

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
            func(samples int64, tempoBPM float64, ppq int, sampleRate int) bool {
                // Generate valid inputs
                // Calculate expected vs actual
                // Verify within tolerance
                return true
            },
            gen.Int64Range(0, 1000000),
            gen.Float64Range(30.0, 300.0),
            gen.IntRange(120, 960),
            gen.IntRange(22050, 96000),
        ))
    
    properties.TestingRun(t, gopter.ConsoleReporter(false))
}
```

### Unit Test Coverage

**Core Functionality**:
- `TestTickGenerator_NewTickGenerator`: Initialization with valid/invalid parameters
- `TestTickGenerator_ProcessSamples_SingleBuffer`: Basic tick advancement
- `TestTickGenerator_ProcessSamples_MultipleBuffers`: Sequential processing
- `TestTickGenerator_ProcessSamples_TempoChange`: Crossing tempo boundaries
- `TestTickGenerator_ProcessSamples_ZeroSamples`: Edge case handling

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

**Buffer Sizes**:
- Small: 256 samples (~5.8ms at 44100 Hz)
- Medium: 2048 samples (~46ms at 44100 Hz)
- Large: 8192 samples (~186ms at 44100 Hz)
- Variable: Random sizes between 64-8192

**Sample Rates**:
- Standard: 44100 Hz
- High: 48000 Hz, 96000 Hz
- Low: 22050 Hz

### Verification Approach

1. **Formula Verification**: Compare calculated ticks against manual calculation
2. **Determinism Verification**: Run same scenario multiple times, verify identical results
3. **Timing Verification**: Measure actual elapsed time vs expected time
4. **Logging Verification**: Parse log output, verify required information present
5. **Headless Verification**: Compare headless vs GUI mode tick values

## Implementation Notes

### Performance Considerations

1. **Avoid Recalculation**: Cache tempo map index to avoid re-traversing from beginning
2. **Float64 Precision**: Use float64 for fractional ticks to maintain precision over long playback
3. **Atomic Operations**: Use atomic operations for `targetTick` to avoid locks in audio thread
4. **Minimal Logging**: Log only significant events (every 100 ticks) to avoid performance impact

### Thread Safety

1. **Audio Thread**: `MidiStream.Read()` runs in audio thread, calls `ProcessSamples()`
2. **Main Thread**: `UpdateVM()` runs in main thread, reads `targetTick`
3. **Synchronization**: Use `atomic.StoreInt64()` and `atomic.LoadInt64()` for `targetTick`
4. **No Locks in Audio Thread**: Avoid mutex locks in audio callback to prevent audio glitches

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
