# Design Document: User Input Handling

## Overview

This design addresses the blocking execution issue in the son-et engine where `mes(TIME)` blocks prevent user interaction with the application window. The current implementation uses `sync.WaitGroup` to block the calling goroutine until sequence execution completes, which prevents the Ebiten game loop from processing input events.

The solution involves removing the blocking behavior from `RegisterSequence` for TIME mode, allowing the script goroutine to complete immediately while the VM continues executing sequences through the game loop's `Update()` calls. This maintains timing accuracy while enabling full window responsiveness.

## Architecture

### Current Architecture (Blocking)

```
Script Goroutine                Game Loop (Main Thread)
     |                                  |
     | mes(TIME) {                      |
     |   RegisterSequence()             |
     |   [BLOCKS HERE]                  |
     |                                  | Update() -> UpdateVM()
     |                                  | Draw()
     |                                  | Process Input Events
     |   <-- WaitGroup.Wait()           |
     | }                                |
     | Continue...                      |
```

**Problem**: The script goroutine blocks during `RegisterSequence`, preventing it from completing. Since the script runs in a separate goroutine started from `Run()`, this doesn't directly block the game loop. However, the busy cursor appears because:
1. The script goroutine is still running (not completed)
2. The OS detects the application has an active long-running operation
3. Window event processing may be delayed due to synchronization overhead

### Proposed Architecture (Non-blocking)

```
Script Goroutine                Game Loop (Main Thread)
     |                                  |
     | mes(TIME) {                      |
     |   RegisterSequence()             |
     |   [RETURNS IMMEDIATELY]          |
     | }                                |
     | Continue...                      |
     | Script Complete                  | Update() -> UpdateVM()
     |                                  | Draw()
     |                                  | Process Input Events (ESC, Close, etc.)
     |                                  | Check Termination Flag
```

**Solution**: Remove the `WaitGroup` blocking for TIME mode, allowing the script to complete immediately. The VM continues executing sequences through the game loop, which processes input events every frame.

## Components and Interfaces

### 1. RegisterSequence Function

**Current Signature**:
```go
func RegisterSequence(mode int, ops []OpCode, initialVars ...map[string]any)
```

**Behavior Changes**:
- Remove `WaitGroup` creation and blocking for TIME mode
- TIME mode becomes non-blocking (like MIDI_TIME mode)
- Sequences execute asynchronously through `UpdateVM()` calls
- Script goroutine completes immediately after registering all sequences

**Implementation**:
```go
func RegisterSequence(mode int, ops []OpCode, initialVars ...map[string]any) {
    // Validate OpCodes
    if err := ValidateOpCodes(ops); err != nil {
        // ... error handling
    }

    // Log registration
    modeStr := "TIME"
    if mode == MidiTime {
        modeStr = "MIDI_TIME"
    }
    fmt.Printf("[%s] RegisterSequence: mode=%s (%d ops)\n",
        time.Now().Format("15:04:05.000"), modeStr, len(ops))

    // NO BLOCKING - both modes are now non-blocking
    fmt.Printf("[%s] RegisterSequence: Non-blocking mode\n",
        time.Now().Format("15:04:05.000"))

    vmLock.Lock()

    // Set sync mode
    if mode == MidiTime {
        midiSyncMode = true
    } else {
        midiSyncMode = false
        // Ensure targetTick allows immediate execution
        if atomic.LoadInt64(&targetTick) < tickCount {
            atomic.StoreInt64(&targetTick, tickCount)
        }
    }

    // Create sequencer (no onComplete callback needed)
    mainSequencer = &Sequencer{
        commands:     ops,
        pc:           0,
        waitTicks:    0,
        active:       true,
        ticksPerStep: 12,
        vars:         vars,
        parent:       parentSeq,
        mode:         mode,
        onComplete:   nil, // No callback needed
    }

    // Add to sequencers list
    sequencers = append(sequencers, mainSequencer)
    
    vmLock.Unlock()

    // Return immediately - no blocking
}
```

### 2. Game.Update Method

**Current Behavior**:
- Increments tick count
- Calls `UpdateVM(currentTick)`
- Checks for mouse events (RBDOWN)

**Enhanced Behavior**:
- Add ESC key detection for termination
- Add window close detection
- Set termination flag when user requests exit
- Return `ebiten.Termination` when termination flag is set

**Implementation**:
```go
func (g *Game) Update() error {
    // Check for termination request FIRST
    if programTerminated {
        fmt.Println("Game.Update: Program terminated, returning Termination")
        return ebiten.Termination
    }

    // Check for ESC key press
    if ebiten.IsKeyPressed(ebiten.KeyEscape) {
        fmt.Println("Game.Update: ESC pressed, terminating")
        programTerminated = true
        return ebiten.Termination
    }

    // Check if window close was requested
    // Note: Ebiten handles window close automatically, but we can detect it
    // by checking if the game loop should terminate

    g.tickCount++

    // Check for mouse events
    if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
        if rbDownHandler != nil {
            TriggerRBDown()
        }
    }

    // Update VM (non-blocking)
    if !midiSyncMode {
        // TIME MODE
        tickLock.Lock()
        tickCount++
        currentTick := int(tickCount)
        tickLock.Unlock()

        UpdateVM(currentTick)
    } else {
        // MIDI SYNC MODE
        currentTarget := atomic.LoadInt64(&targetTick)

        // ... existing MIDI sync logic
    }

    return nil
}
```

### 3. Termination Handling

**New Global Variable**:
```go
var programTerminated bool
```

**Termination Sources**:
1. ESC key press in `Game.Update()`
2. Window close button (handled by Ebiten, triggers game loop exit)
3. `ExitTitle()` function call from script
4. OpCode execution error (critical failures)

**Termination Flow**:
```
User Action (ESC/Close) 
    -> Set programTerminated = true
    -> Return ebiten.Termination from Update()
    -> Ebiten closes window and exits RunGame()
    -> Application terminates
```

### 4. UpdateVM Function

**No Changes Required**:
- Already processes all active sequencers
- Already handles Wait() operations correctly
- Already marks sequences as inactive when complete
- Continues to work with non-blocking RegisterSequence

### 5. Script Execution Flow

**Current Flow**:
```go
func Run() {
    // Start script in goroutine
    go func() {
        time.Sleep(100 * time.Millisecond)
        script() // BLOCKS here during mes(TIME)
    }()

    // Start game loop (blocks until window closes)
    ebiten.RunGame(gameState)
}
```

**New Flow** (no code changes needed):
```go
func Run() {
    // Start script in goroutine
    go func() {
        time.Sleep(100 * time.Millisecond)
        script() // Returns immediately after registering sequences
        fmt.Println("Script goroutine completed")
    }()

    // Start game loop (blocks until window closes)
    // Game loop processes input events and executes VM
    ebiten.RunGame(gameState)
}
```

## Data Models

### Sequencer Structure

**No Changes Required**:
```go
type Sequencer struct {
    commands     []OpCode
    pc           int
    waitTicks    int
    active       bool
    ticksPerStep int
    vars         map[string]any
    parent       *Sequencer
    mode         int
    onComplete   func() // Will be nil for non-blocking mode
    
    // Step execution state
    inStep        bool
    stepBody      []OpCode
    stepCount     int
    stepIteration int
    stepOpIndex   int
}
```

**Note**: The `onComplete` callback is no longer needed for TIME mode, but we keep the field for backward compatibility and potential future use.

### Global State Variables

**Existing**:
```go
var (
    mainSequencer *Sequencer
    sequencers    []*Sequencer
    globalVars    map[string]any
    vmLock        sync.Mutex
    tickCount     int64
    tickLock      sync.Mutex
    midiSyncMode  bool
    targetTick    int64 // Atomic
)
```

**New**:
```go
var (
    programTerminated bool // Termination flag
)
```

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system—essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*


### Property 1: RegisterSequence Non-blocking

*For any* mes(TIME) block, calling RegisterSequence should return immediately (within 10ms) without waiting for sequence completion.

**Validates: Requirements 1.1**

### Property 2: Game Loop Continuity

*For any* active sequence execution, the game loop Update() method should continue to be called at approximately 60 FPS (16.67ms ± 2ms per frame).

**Validates: Requirements 1.2, 3.3, 4.4, 7.4**

### Property 3: Rendering Frame Rate

*For any* script execution state, the Draw() method should be called at approximately 60 FPS, maintaining consistent frame timing.

**Validates: Requirements 1.3**

### Property 4: Concurrent Sequence Execution

*For any* set of multiple mes(TIME) blocks registered concurrently, all sequences should execute in parallel without blocking each other.

**Validates: Requirements 1.4**

### Property 5: Input Responsiveness During Wait

*For any* Wait() operation in progress, the game loop should continue processing input events (keyboard, mouse) without delay.

**Validates: Requirements 3.4**

### Property 6: Frame-Accurate Timing

*For any* mes(TIME) block execution, the tick count should increment at exactly 60 ticks per second (±5% tolerance).

**Validates: Requirements 5.1**

### Property 7: Wait Operation Accuracy

*For any* Wait(N) operation, exactly N ticks should pass before the next OpCode executes, regardless of system load.

**Validates: Requirements 5.2, 8.3**

### Property 8: Concurrent Timing Independence

*For any* set of concurrent sequences with Wait() operations, each sequence should maintain its own timing accuracy independent of other sequences.

**Validates: Requirements 5.3**

### Property 9: Termination Stops Execution

*For any* active sequence, when programTerminated flag is set to true, no new OpCodes should execute in subsequent UpdateVM() calls.

**Validates: Requirements 6.1, 6.2**

### Property 10: Update Advances VM

*For any* call to Game.Update(), the VM tick count should increment by exactly 1 in TIME mode.

**Validates: Requirements 7.1**

### Property 11: UpdateVM Non-blocking

*For any* UpdateVM() call, the execution time should be less than 5ms to avoid blocking the game loop.

**Validates: Requirements 7.3**

### Property 12: Error Recovery Responsiveness

*For any* OpCode execution error, the game loop should continue calling Update() and processing input events without interruption.

**Validates: Requirements 10.4**

## Error Handling

### Error Categories

1. **OpCode Validation Errors**
   - Detected at registration time (before execution)
   - Cause immediate panic with descriptive message
   - Prevent invalid sequences from executing

2. **OpCode Execution Errors**
   - Detected during UpdateVM() execution
   - Log error with context (sequence ID, PC, OpCode)
   - Mark sequence as inactive
   - Continue processing other sequences

3. **Termination Requests**
   - User-initiated (ESC key, window close)
   - Script-initiated (ExitTitle())
   - Set programTerminated flag
   - Return ebiten.Termination from Update()

### Error Handling Flow

```
OpCode Execution Error
    -> Log error with context
    -> Set seq.active = false
    -> Continue UpdateVM() for other sequences
    -> Game loop continues normally

User Termination Request
    -> Set programTerminated = true
    -> Return ebiten.Termination from Update()
    -> Ebiten closes window
    -> Application exits

Critical Error (Validation)
    -> Panic with error message
    -> Application terminates immediately
```

### Error Logging Format

```
VM Error: [Seq %d] [PC %d] %s: %v
```

Example:
```
VM Error: [Seq 0] [PC 15] OpLoadPic: file not found: missing.bmp
```

## Testing Strategy

### Dual Testing Approach

This feature requires both unit tests and property-based tests to ensure correctness:

**Unit Tests** focus on:
- Specific examples of non-blocking behavior
- ESC key detection and termination
- Window close event handling
- Error logging format verification
- Backward compatibility with existing scripts

**Property-Based Tests** focus on:
- RegisterSequence timing across many sequences
- Frame rate consistency under various loads
- Wait() operation accuracy with random tick counts
- Concurrent sequence execution with random operations
- Timing independence across multiple sequences

### Property-Based Testing Configuration

- **Library**: Use `testing/quick` for Go property-based testing
- **Iterations**: Minimum 100 iterations per property test
- **Tag Format**: `// Feature: user-input-handling, Property N: [property text]`

### Test Organization

```
pkg/engine/
├── user_input_test.go           # Unit tests for input handling
├── non_blocking_test.go         # Unit tests for non-blocking execution
├── termination_test.go          # Unit tests for termination handling
├── timing_property_test.go      # Property tests for timing accuracy
├── concurrency_property_test.go # Property tests for concurrent execution
└── integration_test.go          # Integration tests with real scripts
```

### Key Test Scenarios

1. **Non-blocking Registration**
   - Register mes(TIME) block
   - Measure time to return
   - Verify < 10ms

2. **ESC Key Termination**
   - Simulate ESC key press
   - Verify programTerminated = true
   - Verify Update() returns ebiten.Termination

3. **Concurrent Sequence Execution**
   - Register multiple mes(TIME) blocks
   - Verify all execute in parallel
   - Verify timing accuracy for each

4. **Wait() Accuracy**
   - Execute Wait(N) for various N
   - Count ticks until next operation
   - Verify exactly N ticks passed

5. **Error Recovery**
   - Trigger OpCode execution error
   - Verify sequence marked inactive
   - Verify game loop continues

6. **Backward Compatibility**
   - Run existing test scripts (kuma2, robot, etc.)
   - Verify output matches expected behavior
   - Verify timing is preserved

### Integration Testing

Integration tests should verify the complete flow:

1. Start application
2. Load and execute FILLY script
3. Verify window opens and renders
4. Simulate user input (ESC, mouse clicks)
5. Verify termination works correctly
6. Verify no orphaned processes

### Headless Mode Testing

Headless mode tests should verify:

1. Sequences execute without GUI
2. Timing accuracy maintained
3. Timeout flag works correctly
4. Logs contain timestamps

## Implementation Notes

### Backward Compatibility Considerations

1. **Existing Scripts**: All existing FILLY scripts should continue to work without modification
2. **Observable Behavior**: The timing and execution order should remain identical
3. **Event Handlers**: MIDI_END, RBDOWN, RBDBLCLK handlers should trigger at the same times
4. **MIDI_TIME Mode**: Already non-blocking, should remain unchanged

### Performance Considerations

1. **UpdateVM() Execution Time**: Should complete in < 5ms to maintain 60 FPS
2. **Sequence Count**: Should handle up to 100 concurrent sequences without performance degradation
3. **Memory Usage**: Sequencer structures should be cleaned up when inactive

### Platform Considerations

1. **macOS**: Primary development platform, full testing required
2. **Headless Mode**: Must work on CI/CD systems without display
3. **Ebiten Compatibility**: Must work with Ebiten v2 game loop

### Migration Path

1. **Phase 1**: Remove WaitGroup blocking from RegisterSequence
2. **Phase 2**: Add ESC key detection to Game.Update()
3. **Phase 3**: Add termination flag checking
4. **Phase 4**: Test with existing scripts (kuma2, robot, etc.)
5. **Phase 5**: Add property-based tests for timing accuracy

### Known Limitations

1. **Window Move/Resize**: Handled by Ebiten, not directly testable
2. **Cursor State**: Controlled by OS, not directly controllable
3. **Exit Code**: Controlled by OS, not directly verifiable in tests
4. **Resource Cleanup**: Requires integration testing, not unit testable

### Future Enhancements

1. **Pause/Resume**: Allow pausing sequence execution
2. **Step Debugging**: Step through OpCodes one at a time
3. **Sequence Priority**: Allow prioritizing certain sequences
4. **Performance Monitoring**: Track UpdateVM() execution time
5. **Sequence Limits**: Limit maximum concurrent sequences

## References

- Ebiten Game Loop Documentation: https://ebiten.org/
- Go sync Package: https://pkg.go.dev/sync
- Go atomic Package: https://pkg.go.dev/sync/atomic
- FILLY Language Specification: (internal documentation)
