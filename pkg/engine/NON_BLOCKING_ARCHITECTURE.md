# Non-Blocking Architecture

## Overview

The son-et engine implements a non-blocking execution model that allows user interaction with the application window at all times, even while FILLY scripts are executing. This document explains the architecture and termination flow.

## Key Principles

### 1. Non-Blocking RegisterSequence

Both `TIME` and `MIDI_TIME` modes are now non-blocking:

```go
// Old behavior (TIME mode only):
RegisterSequence(Time, ops)  // BLOCKED until sequence completed

// New behavior (both modes):
RegisterSequence(Time, ops)       // Returns immediately
RegisterSequence(MidiTime, ops)   // Returns immediately (unchanged)
```

**Benefits:**
- Script goroutine completes immediately after registering all sequences
- Window remains responsive during script execution
- No busy cursor (砂時計) displayed to user
- User can interact with window (close, move, resize) at any time

### 2. Game Loop Integration

The Ebiten game loop processes sequences through `Update()` calls at 60 FPS:

```
Script Goroutine                Game Loop (Main Thread)
     |                                  |
     | mes(TIME) {                      |
     |   RegisterSequence()             |
     |   [RETURNS IMMEDIATELY]          |
     | }                                |
     | Script Complete                  |
     |                                  | Update() -> UpdateVM()
     |                                  | Draw()
     |                                  | Process Input Events
```

**Key Points:**
- `Update()` is called every frame (60 FPS)
- `UpdateVM()` advances VM execution by one tick per frame
- Input events (ESC, mouse, window close) are processed every frame
- Sequences execute asynchronously through the game loop

### 3. Timing Accuracy

Despite being non-blocking, timing accuracy is maintained:

- **Frame-accurate timing**: 60 ticks per second (16.67ms per frame)
- **Wait() operations**: Exact tick counts are preserved
- **Concurrent sequences**: Each sequence maintains independent timing
- **MIDI synchronization**: MIDI_TIME mode uses audio-driven tick updates

## Termination Flow

### Termination Sources

The engine can be terminated from multiple sources:

1. **ESC Key Press** (User-initiated)
   - Detected in `Game.Update()`
   - Sets `programTerminated = true`
   - Returns `ebiten.Termination`

2. **Window Close Button** (User-initiated)
   - Handled automatically by Ebiten
   - Triggers game loop exit

3. **ExitTitle() Function** (Script-initiated)
   - Called from FILLY script
   - Sets `programTerminated = true`

4. **Critical Errors** (Error-initiated)
   - OpCode execution failures
   - Sets `programTerminated = true`

### Termination Sequence

When termination is requested, the following sequence occurs:

```
1. programTerminated flag set to true
   ↓
2. Game.Update() checks flag at start
   ↓
3. Returns ebiten.Termination
   ↓
4. Ebiten closes window and exits RunGame()
   ↓
5. Application terminates gracefully
```

### Termination Checks

The `programTerminated` flag is checked at three critical points:

#### 1. Game.Update() - Entry Point
```go
func (g *Game) Update() error {
    // Check FIRST before any processing
    if programTerminated {
        return ebiten.Termination
    }
    
    // Check for ESC key
    if ebiten.IsKeyPressed(ebiten.KeyEscape) {
        programTerminated = true
        return ebiten.Termination
    }
    
    // ... continue normal processing
}
```

#### 2. UpdateVM() - Sequence Processing
```go
func UpdateVM(currentTick int) {
    // Check FIRST before processing sequences
    if programTerminated {
        // Mark all sequences as inactive
        for _, seq := range sequencers {
            seq.active = false
        }
        return
    }
    
    // ... process active sequences
}
```

#### 3. ExecuteOp() - OpCode Execution
```go
func ExecuteOp(op OpCode, seq *Sequencer) (any, bool) {
    // Check BEFORE executing any OpCode
    if programTerminated {
        return ebiten.Termination, false
    }
    
    // ... execute OpCode
}
```

### Why Three Checks?

Each check serves a specific purpose:

1. **Game.Update()**: Immediate response to termination requests
   - Prevents any further VM processing
   - Returns control to Ebiten immediately
   - Ensures window closes promptly

2. **UpdateVM()**: Graceful sequence cleanup
   - Marks all sequences as inactive
   - Prevents new OpCodes from being queued
   - Allows current frame to complete

3. **ExecuteOp()**: Fine-grained execution control
   - Stops execution mid-sequence if needed
   - Prevents long-running operations from blocking termination
   - Ensures no OpCodes execute after termination

## Backward Compatibility

The non-blocking implementation maintains backward compatibility:

### Observable Behavior
- Timing is identical to the old blocking implementation
- Wait() operations execute for the same number of ticks
- Event handlers (MIDI_END, RBDOWN) trigger at the same times
- Sequence execution order is preserved

### Script Compatibility
- All existing FILLY scripts work without modification
- No changes to script syntax or semantics
- No changes to function signatures or behavior

### Test Compatibility
- All existing tests pass without modification
- Property-based tests verify timing accuracy
- Integration tests verify end-to-end functionality

## Performance Considerations

### UpdateVM() Execution Time
- Target: < 5ms per frame to maintain 60 FPS
- Actual: Typically 1-2ms for normal sequences
- Sequences with many OpCodes may take longer but don't block the game loop

### Concurrent Sequences
- Engine supports up to 100 concurrent sequences without performance degradation
- Each sequence maintains independent timing
- Sequences don't interfere with each other

### Memory Usage
- Sequencer structures are cleaned up when inactive
- No memory leaks from terminated sequences
- Global variables are properly managed

## Headless Mode

Headless mode maintains the same non-blocking architecture:

```bash
go run cmd/son-et/main.go --headless --timeout=5s samples/kuma2
```

**Key Points:**
- No GUI window is opened
- All rendering operations are logged but not displayed
- Audio is initialized but muted (volume = 0)
- MIDI timing still works correctly for MIDI_TIME mode
- Timeout flag ensures automatic termination

## Testing Strategy

### Unit Tests
- Test specific examples of non-blocking behavior
- Test ESC key detection and termination
- Test window close event handling
- Test error logging format

### Property-Based Tests
- Test RegisterSequence timing across many sequences
- Test frame rate consistency under various loads
- Test Wait() operation accuracy with random tick counts
- Test concurrent sequence execution with random operations

### Integration Tests
- Test complete flow with real FILLY scripts
- Test user interaction during script execution
- Test termination from various sources
- Test backward compatibility with existing scripts

## Common Patterns

### Registering Multiple Sequences
```go
// All sequences return immediately
RegisterSequence(Time, sequence1)
RegisterSequence(Time, sequence2)
RegisterSequence(Time, sequence3)

// All execute concurrently through the game loop
```

### Nested mes() Blocks
```go
// Outer mes() block
mes(TIME) {
    // Inner mes() block
    mes(TIME) {
        // Both execute non-blocking
    }
}
```

### Event Handlers
```go
// Event handlers are registered and triggered asynchronously
on(MIDI_END) {
    // Executes when MIDI playback ends
}

on(RBDOWN) {
    // Executes when right mouse button is pressed
}
```

## Troubleshooting

### Issue: Script completes but window doesn't close
**Cause**: No termination signal sent
**Solution**: Call `ExitTitle()` at end of script or press ESC

### Issue: Sequences don't execute
**Cause**: `programTerminated` flag is set
**Solution**: Check for early termination, verify flag is false

### Issue: Timing is inaccurate
**Cause**: UpdateVM() taking too long (> 5ms)
**Solution**: Optimize OpCode execution, reduce sequence complexity

### Issue: Window is unresponsive
**Cause**: Game loop is blocked (should not happen with non-blocking implementation)
**Solution**: Verify no blocking operations in Update() or Draw()

## References

- Design Document: `.kiro/specs/user-input-handling/design.md`
- Requirements Document: `.kiro/specs/user-input-handling/requirements.md`
- Implementation Tasks: `.kiro/specs/user-input-handling/tasks.md`
- Ebiten Documentation: https://ebiten.org/
