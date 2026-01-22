# Design Document: Sample Game Fixes

## Overview

This design addresses four critical bugs in the son-et game engine that prevent sample games from executing correctly. The son-et engine is a Go-based interpreter for legacy FILLY/Toffy scripts, providing cross-platform execution of visual novel content. Through investigation, we have identified specific engine components responsible for each issue and designed targeted fixes.

The four issues are:
1. **sab2 termination**: Game does not exit properly after completion
2. **y_saru P14 hang**: Execution gets stuck at step P14 in mes() blocks
3. **yosemiya animation**: Curtain opening animation does not play
4. **yosemiya mojibake**: Virtual window text displays garbled Japanese characters

## Architecture

### Component Overview

The son-et engine consists of several key components:

```
┌─────────────────────────────────────────────────────────┐
│                        Engine                            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │   Compiler   │  │      VM      │  │   Renderer   │  │
│  │ (TFY→OpCode) │  │  (Executor)  │  │   (Display)  │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
│                                                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │  Sequencer   │  │ MIDI Player  │  │     Text     │  │
│  │   (Timing)   │  │   (Audio)    │  │  (Rendering) │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
└─────────────────────────────────────────────────────────┘
```

### Execution Flow

1. **Compilation**: TFY scripts are parsed and converted to OpCodes
2. **Registration**: mes() blocks are registered as Sequencers with timing modes (TIME or MIDI_TIME)
3. **Execution**: The VM executes OpCodes from active Sequencers
4. **Timing**: Sequencers manage wait states and step progression
5. **Termination**: Engine monitors completion and triggers shutdown

## Components and Interfaces

### Issue 1: sab2 Termination Problem

**Root Cause**: The sab2 sample contains unsupported legacy functions (Shell, GetIniStr, MCI) that cause parsing errors. The parser fails before execution begins, preventing proper termination logic from running.

**Component**: Compiler/Parser (pkg/compiler/)

**Current Behavior**:
- Parser encounters unsupported function calls
- Parsing fails with multiple errors
- Engine exits with error code but doesn't execute termination cleanup

**Fix Strategy**: Add stub implementations for legacy functions that are no longer supported. These stubs will:
- Parse successfully (preventing parse errors)
- Log warnings when called
- Return safe default values
- Allow scripts to continue execution

**Affected Functions**:
- `Shell()` - Launches external programs (Windows-specific)
- `GetIniStr()` - Reads INI files (legacy configuration)
- `MCI()` - Windows Media Control Interface commands
- `StrMCI()` - String variant of MCI commands

### Issue 2: y_saru P14 Hang

**Root Cause**: Investigation shows that y_saru uses MIDI_TIME mode with step() blocks. The sequencer correctly processes wait states, but there may be an issue with how the script is structured or how certain step indices are handled.

**Component**: Sequencer (pkg/engine/sequencer.go) and VM (pkg/engine/vm.go)

**Current Behavior**:
- mes(MIDI_TIME) blocks execute with step(8) timing
- Wait operations correctly decrement
- Execution appears to progress normally in logs

**Investigation Needed**: 
- Examine the actual TFY script to identify what happens at P14
- Check if there's a specific OpCode or pattern that causes the hang
- Verify MIDI synchronization is working correctly

**Potential Fix**: 
- Add bounds checking for step indices
- Improve logging around step transitions
- Ensure MIDI_TIME sequences properly advance their program counter

### Issue 3: yosemiya Animation

**Root Cause**: The curtain opening animation likely involves MovePic() commands that manipulate cast positions over time. The issue may be related to:
- Timing synchronization between animation frames
- Cast buffer management
- Rendering updates not being triggered

**Component**: Drawing Context (pkg/engine/drawing.go) and Cast Management

**Current Behavior**:
- MovePic() commands are issued
- Cast positions may not update visually
- Animation frames may be skipped

**Fix Strategy**:
- Verify MovePic() correctly updates cast positions
- Ensure rendering pipeline processes cast updates
- Check timing between animation frames
- Validate cast buffer state transitions

### Issue 4: yosemiya Mojibake

**Root Cause**: Japanese text in virtual windows displays as garbled characters. This is a character encoding issue where Shift-JIS encoded text is not being properly converted or rendered.

**Component**: Text Renderer (pkg/engine/text.go) and Preprocessor (pkg/compiler/preprocessor/)

**Current Behavior**:
- TFY scripts are preprocessed and converted from Shift-JIS to UTF-8
- TextWrite() function renders text to picture buffers
- Virtual windows display the rendered text
- Japanese characters appear corrupted

**Analysis**:
The preprocessor already handles Shift-JIS to UTF-8 conversion for TFY source files. However, there may be issues with:
1. String literals embedded in the compiled OpCodes
2. Text rendering font selection for Japanese characters
3. Virtual window text buffer encoding

**Fix Strategy**:
1. Verify string literals preserve UTF-8 encoding through compilation
2. Ensure TextWrite() correctly handles multi-byte UTF-8 characters
3. Confirm font loading supports Japanese character sets
4. Check virtual window text buffer encoding

## Data Models

### Sequencer State

```go
type Sequencer struct {
    commands     []interpreter.OpCode  // OpCode sequence
    pc           int                   // Program counter
    active       bool                  // Is sequence active?
    mode         TimingMode            // TIME or MIDI_TIME
    waitCount    int                   // Ticks remaining in wait
    ticksPerStep int                   // Ticks per step
    vars         map[string]any        // Variable scope
    parent       *Sequencer            // Parent scope
    id           int                   // Unique ID
    groupID      int                   // Group ID
}
```

### Termination State

```go
type Engine struct {
    programTerminated atomic.Bool  // Termination flag
    timeout           time.Duration // Execution timeout
    startTime         time.Time     // Start timestamp
    // ... other fields
}
```

### Text Rendering State

```go
type TextRenderer struct {
    currentFont     font.Face  // Active font
    currentFontSize int        // Font size
    currentFontName string     // Font name
    textColor       color.Color // Text color
    bgColor         color.Color // Background color
    backMode        int         // 0=transparent, 1=opaque
}
```

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system—essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: Engine Termination Completeness

*For any* TFY script that completes all execution paths (including mes() blocks with no more scheduled events), the engine should terminate within 1 second and return exit code 0, without entering an infinite wait state.

**Validates: Requirements 1.1, 1.3, 1.4, 1.5**

### Property 2: Sequencer Step Progression

*For any* mes() block containing step() sequences, the sequencer should execute each step in order without skipping or hanging on any valid step index (including edge cases like P14).

**Validates: Requirements 2.1, 2.3, 2.5**

### Property 3: Step Timing Accuracy

*For any* step scheduled for a specific tick or time in a mes() block, the sequencer should execute it at the correct moment according to the timing mode (TIME or MIDI_TIME).

**Validates: Requirements 2.2**

### Property 4: Step Block Completion Detection

*For any* mes() block, when all steps are complete, the sequencer should mark the block as finished and allow the engine to proceed with termination checks.

**Validates: Requirements 2.4**

### Property 5: Animation Frame Completeness

*For any* sequence of MovePic() commands in an animation, the engine should execute all animation frames in the correct order without skipping frames due to timing or rendering issues.

**Validates: Requirements 3.1, 3.2, 3.5**

### Property 6: Text Encoding Preservation

*For any* Japanese text string in a TFY script, the text should be correctly converted from Shift-JIS to UTF-8 during preprocessing, preserved through compilation, and rendered without mojibake (garbled characters) when displayed via TextWrite().

**Validates: Requirements 4.1, 4.2, 4.3, 4.4, 4.5**

## Error Handling

### Parse Errors

**Current**: Parser fails on unsupported functions, preventing execution
**Improved**: Parser accepts legacy functions with warnings, allows execution to continue

### Runtime Errors

**Current**: VM errors may not provide sufficient context for debugging
**Improved**: Enhanced logging with sequencer ID, PC, and OpCode details

### Encoding Errors

**Current**: Mojibake occurs silently without error reporting
**Improved**: Log warnings when character encoding issues are detected

### Termination Errors

**Current**: Games may hang indefinitely if termination logic fails
**Improved**: Timeout mechanism ensures eventual termination

## Testing Strategy

### Dual Testing Approach

This feature requires both unit tests and property-based tests:

**Unit Tests**: Verify specific examples, edge cases, and error conditions
- Test legacy function stub implementations
- Test specific step indices (including P14)
- Test animation sequences with known inputs
- Test specific Japanese text strings

**Property Tests**: Verify universal properties across all inputs
- Test termination across various script structures
- Test sequencer progression with random step counts
- Test text encoding with random Japanese strings
- Test animation timing with random frame counts

### Property-Based Testing Configuration

- Use Go's testing/quick package or a PBT library like gopter
- Minimum 100 iterations per property test
- Each test tagged with: **Feature: sample-game-fixes, Property {number}: {property_text}**

### Integration Testing

**Test Samples**:
1. Run sab2 and verify it parses and terminates
2. Run y_saru and verify it progresses past P14
3. Run yosemiya and verify curtain animation plays
4. Run yosemiya and verify text displays correctly

**Verification**:
- Use headless mode with timeout for automated testing
- Check exit codes (should be 0 for success)
- Parse logs for error messages
- Verify expected log patterns appear

### Regression Testing

**Existing Samples**: Test other sample games to ensure fixes don't break working functionality
- samples/test_minimal
- samples/test_loop
- samples/test_drawing
- samples/mes_test

**Test Suite**: Run existing unit tests to catch regressions
```bash
go test ./pkg/engine/... -v
go test ./pkg/compiler/... -v
```
