# Common Design

This document defines design elements that are shared across multiple son-et specifications.

## Glossary

See [GLOSSARY.md](GLOSSARY.md) for common terms.

## OpCode Structure

All FILLY code is executed as OpCode sequences by the virtual machine.

### OpCode Definition

```go
type OpCode struct {
    Cmd  string  // Command name (e.g., "Assign", "Call", "If", "For")
    Args []any   // Command arguments (can contain nested OpCodes)
}
```

### OpCode Representation Examples

**Variable Assignment:**
```go
// x = 10
OpCode{Cmd: "Assign", Args: []any{"x", 10}}
```

**Function Call:**
```go
// MovePic(0, 100, 200)
OpCode{Cmd: "Call", Args: []any{"MovePic", 0, 100, 200}}
```

**Control Flow - If Statement:**
```go
// if (x > 10) { ... } else { ... }
OpCode{Cmd: "If", Args: []any{
    OpCode{Cmd: ">", Args: []any{Variable("x"), 10}},
    []OpCode{...}, // then branch
    []OpCode{...}, // else branch (optional)
}}
```

**Control Flow - For Loop:**
```go
// for (i=0; i<10; i=i+1) { ... }
OpCode{Cmd: "For", Args: []any{
    OpCode{Cmd: "Assign", Args: []any{"i", 0}},           // init
    OpCode{Cmd: "<", Args: []any{Variable("i"), 10}},     // condition
    OpCode{Cmd: "Assign", Args: []any{"i", OpCode{Cmd: "+", Args: []any{Variable("i"), 1}}}}, // post
    []OpCode{...}, // body
}}
```

**mes() Blocks:**
```go
// mes(TIME) { ... }
OpCode{Cmd: "RegisterSequence", Args: []any{
    0, // TIME mode
    []OpCode{...}, // body
    map[string]any{...}, // captured variables
}}
```

## Variable Scope Model

Variables are resolved through a hierarchical scope chain, similar to lexical scoping in modern languages.

### Scope Chain Architecture

```
Root Sequencer (main function scope)
  vars: {winw: 640, winh: 480, p39: 39, ...}
  parent: nil
  ↓
Child Sequencer (mes block scope)
  vars: {localVar: 123}  // Variables defined inside mes() block
  parent: → Root Sequencer
  ↓
Variable lookup: localVar → found in current scope
Variable lookup: winw → not in current, check parent → found in parent
Variable lookup: p39 → not in current, check parent → found in parent
```

### Sequencer Structure

```go
type Sequencer struct {
    commands     []OpCode
    pc           int
    waitTicks    int
    active       bool
    ticksPerStep int
    vars         map[string]any
    parent       *Sequencer  // Parent scope reference for variable lookup
    mode         int
    onComplete   func()
}
```

### Variable Resolution

```go
func ResolveArg(arg any, seq *Sequencer) any {
    switch v := arg.(type) {
    case Variable:
        // Case-insensitive variable lookup (FILLY is case-insensitive)
        varName := strings.ToLower(string(v))
        
        // Walk up the scope chain
        currentSeq := seq
        for currentSeq != nil {
            if val, ok := currentSeq.vars[varName]; ok {
                return val  // Found in this scope
            }
            currentSeq = currentSeq.parent  // Check parent scope
        }
        
        // Variable not found in any scope
        return 0
    // ... other cases
    }
}
```

### Case Sensitivity

FILLY is case-insensitive for identifiers. The VM implements this by:
- Converting all variable names to lowercase on storage: `vars[strings.ToLower(name)] = value`
- Converting all variable lookups to lowercase: `varName := strings.ToLower(string(v))`
- This ensures `winW`, `winw`, `WINW` all refer to the same variable

## Concurrency Model

The engine operates using a **Producer-Consumer** concurrency model to bridge the gap between the procedural, blocking nature of legacy FILLY scripts and the event-driven, non-blocking nature of Ebitengine.

### Dual-Thread Architecture

**Main Thread (Ebitengine Loop):**
- Runs `Game.Update()` and `Game.Draw()` at 60 FPS
- Responsible for input handling and final rendering to the screen
- **Constraint**: Cannot be blocked. Must return quickly.

**Script Goroutine:**
- Runs the interpreted user code
- Executes blocking commands like `Wait()`, `EnterMes()`, and long-running loops
- **Constraint**: Modifies game state (Cast positions, Picture contents) asynchronously

### Synchronization Strategy

Because the Script Goroutine modifies resources (`pictures`, `casts`) that the Main Thread reads for drawing, a race condition exists.

**Solution: Global `renderMutex`**

**Lock Scope:**
- **Write Access**: All script functions that modify state (`PutCast`, `MoveCast`, `DelCast`, `OpenWin`, `TextWrite`, `LoadPic`) must acquire the lock
- **Read Access**: The `Game.Draw()` loop acquires the lock for the duration of frame rendering

**Critical Rule: Avoid Double-Locking**

Many engine functions acquire `renderMutex` at function entry:

```go
func MoveWin(winID, pic, x, y, w, h, picX, picY int) {
    renderMutex.Lock()
    defer renderMutex.Unlock()
    // ... implementation
}
```

**Fatal mistake**: Acquiring the same lock in `ExecuteOp` before calling these functions causes deadlock.

**Correct Pattern**: Don't lock in `ExecuteOp` - let the function handle it:

```go
case "MoveWin":
    // Resolve arguments
    rArgs := make([]any, len(op.Args))
    for i, a := range op.Args {
        rArgs[i] = ResolveArg(a, seq)
    }
    
    // Call directly - MoveWin will handle locking
    MoveWin(rArgs[0].(int), rArgs[1].(int), ...)
    return nil, false
```

## Rendering Pipeline

### Double Buffering (Flicker Prevention)

Directly drawing to the destination image from the Script Goroutine causes flickering because `Game.Draw()` might display the image while it is being cleared or partially drawn.

**Solution:**
1. `MoveCast` creates a temporary, off-screen buffer (`newImg := ebiten.NewImage(...)`)
2. It performs all drawing operations (Clear → Draw Background → Draw Casts) on `newImg`
3. **Atomic Swap**: Inside the mutex lock, it swaps the pointer: `destPic.Image = newImg`
4. The old image is discarded (garbage collected)

### Z-Ordering (Layering)

To ensure consistent layering of sprites:

**Policy**: Creation Order (Painter's Algorithm)

**Implementation:**
- A global slice `castDrawOrder []int` tracks Cast IDs in the order they were added via `PutCast`
- `MoveCast` iterates through `castDrawOrder` rather than the `casts` map (which has random iteration order)
- This guarantees that a cast created later is always drawn on top of a cast created earlier

### Sprite Management

**Clipping**: The engine supports `SubImage` rendering. `Cast` structs store `SrcX`, `SrcY`, `W`, `H` to define the region of interest in the source texture.

**Transparency**: Per-pixel transparency is handled via `drawWithColorKey` which checks the top-left pixel (or a specified color) to determine the key color.

## Timing & Synchronization

The engine utilizes a **Dual Timing Architecture** to support both musical synchronization and time-based scripting.

### MIDI Sync Mode (`mes(MIDI_TIME)`)

**Use Case**: MIDI-synchronized animations

**Characteristics:**
- `RegisterSequence()` **immediately returns** (non-blocking)
- Timing driven by MIDI player (`NotifyTick` callbacks)
- `step(n)` = `n * 32nd note` (musical time)
- Script execution **continues** immediately after registering sequence

**Driver**: Audio Thread (`NotifyTick` callback from MIDI synthesizer)

**Resolution**: High precision, tied to MIDI ticks (PPQ)

**Step Logic**: `step(n)` defines the wait unit in terms of musical time
- Logic: `1 step = 32nd note * n`
- Example: `step(8)` sets the wait unit to a 4th note (quarter note)

**Critical Requirement**: `PlayMIDI()` **must** be callable after `mes` block
- If `RegisterSequence` blocked, execution would never reach `PlayMIDI()`
- → MIDI player never starts → `targetTick` never updates → VM never executes → **deadlock**

**Implementation Detail:**
- `RegisterSequence(mode=1, ops)` with `mode == MidiTime` does NOT create `WaitGroup`
- Returns immediately, allowing `PlayMIDI()` to execute
- VM execution starts only when `NotifyTick` advances `targetTick`

### Time Mode (`mes(TIME)`)

**Use Case**: Procedural animations

**Characteristics:**
- `RegisterSequence()` **blocks** (via `WaitGroup.Wait()`) until sequence completes
- Timing driven by main game loop (60 FPS)
- `step(n)` = `n * 50ms` (3 frames per step)
- Script execution **pauses** at `mes` block until all commands finish

**Driver**: Main Game Loop (Frame-based, 60 FPS)

**Resolution**: 50ms base unit

**Behavior:**
- Decoupled from MIDI clock to ensure consistent execution speed
- One "tick" in this mode corresponds to 1 frame (1/60s)

**Step Logic**: `step(n)` defines the wait unit in milliseconds
- Logic: `1 step = n * 50ms` (approx. 3 frames)
- Example: `step(20)` waits for 1 second (1000ms)

**Implementation Detail:**
- `RegisterSequence(mode=0, ops)` with `mode != MidiTime` creates `WaitGroup`
- Caller waits for `onComplete` callback
- Ensures sequential execution (mes → post-mes code)

### Synchronization Mechanism Comparison

| Aspect | mes(TIME) | mes(MIDI_TIME) |
|--------|-----------|----------------|
| **RegisterSequence Blocking** | Yes (WaitGroup) | No (immediate return) |
| **Tick Driver** | `Game.Update()` (60 FPS) | `NotifyTick()` (MIDI player) |
| **targetTick Update** | Frame-based increment | Audio thread callback |
| **Step Unit** | `n * 50ms` | `n * 32nd note` |
| **Execution Order** | Sequential (mes → post-mes) | Concurrent (mes \|\| post-mes) |
| **Primary Use** | Procedural scripts | Music-synchronized scripts |

### Common Pitfalls

1. **Making MIDI_TIME blocking**: Causes deadlock (PlayMIDI never executes)
2. **Making TIME non-blocking**: Breaks sequential logic (CloseWin executes before mes finishes)
3. **Applying TIME logic to MIDI_TIME**: Causes 60FPS execution instead of MIDI sync
4. **Bootstrap targetTick in MIDI mode**: Breaks synchronization (executes too fast)

## Virtual Display Architecture

### Virtual Desktop

The application runs in a fixed **1280x720** window (Virtual Desktop).

**Purpose**: Provides a modern canvas for legacy content without scaling artifacts

### Virtual Windows

The original legacy games ("scenarios") typically run at 640x480. These are rendered as "Virtual Windows" inside the 1280x720 desktop.

**Behavior:**
- `OpenWin` commands in the script create these virtual windows
- The engine manages the desktop background and window positioning
- This approach avoids scaling artifacts by rendering the legacy content at 1:1 pixel ratio within the larger modern canvas

## Data Models

### Picture

```go
type Picture struct {
    ID    int
    Image *ebiten.Image
    Width int
    Height int
}
```

### Cast

```go
type Cast struct {
    ID               int
    SrcPicID         int
    DstPicID         int
    X, Y             int
    W, H             int
    SrcX, SrcY       int
    TransparentColor color.Color
}
```

### Window

```go
type Window struct {
    ID       int
    PicID    int
    X, Y     int
    Width    int
    Height   int
    PicX, PicY int
    BgColor  color.Color
    Caption  string
}
```

### VM Sequence

```go
type Sequence struct {
    ID         int
    Operations []OpCode
    PC         int
    StepSize   int
    Mode       int
    TickCount  int
    WaitUntil  int
    Completed  bool
}
```

## Error Handling

### Runtime Errors

**Invalid Resource IDs:**
- When a function is called with a non-existent picture/cast/window ID
- Action: Log error, return early without crashing

**File Loading Failures:**
- When LoadPic/PlayMIDI/PlayWAVE fails to load asset
- Action: Log error, return invalid ID (-1), continue execution

**Audio Initialization Failures:**
- When MIDI player or audio system fails to initialize
- Action: Log error, disable audio features, continue execution

**Mutex Deadlock Prevention:**
- When potential double-locking is detected
- Action: Use defer patterns and careful lock ordering to prevent deadlocks

**Resource Exhaustion:**
- When too many pictures/casts/windows are created
- Action: Log warning, allow creation but monitor memory usage

### Debugging Support

**Debug Levels:**
- `DEBUG_LEVEL=0`: Errors only
- `DEBUG_LEVEL=1`: Important operations (default)
- `DEBUG_LEVEL=2`: All debug output (VM execution, tick updates, render details)

**Logging Format:**
```bash
DEBUG_LEVEL=2 ./executable 2>&1 | while IFS= read -r line; do 
  echo "$(date '+%H:%M:%S.%3N') $line"
done | tee debug.log
```

**Key Log Indicators:**
- `VM: Executing [PC] CommandName (Tick N)` - Command execution trace
- `RegisterSequence: MIDI Sync Mode ON/OFF` - Confirms execution mode
- `PERF: [PC] Command took Xms` - Performance monitoring

## Critical Design Constraints

**Thread Safety:**
- All functions that modify shared state MUST acquire `renderMutex`
- NEVER acquire `renderMutex` in `ExecuteOp` before calling functions that already acquire it
- Use `defer` for lock release to prevent deadlocks

**Timing Modes:**
- MIDI_TIME mode: RegisterSequence MUST NOT block
- TIME mode: RegisterSequence MUST block until completion
- Never mix timing mode logic

**Memory Management:**
- Use double buffering for flicker-free rendering
- Discard old buffers for garbage collection
- Clean up resources on script termination

## Common Mistakes to Avoid

1. Applying TIME logic to MIDI_TIME → Breaks MIDI sync
2. Making MIDI_TIME blocking → Deadlocks (PlayMIDI never runs)
3. Double-locking mutexes → Deadlock
4. Bootstrapping targetTick in MIDI mode → Fast-forward execution
5. Assuming understanding without verification → Test both modes!


## Known Issues and Solutions (Lessons Learned)

### Issue 1: step() Block Misinterpretation

**Problem:** Initially interpreted `step(n)` as a loop count (repeat n times), but the correct interpretation is that `step(n)` sets the time duration for each Wait(1) operation.

**Correct Behavior:**
- `step(65)` in TIME mode → Each Wait(1) = 65 * 50ms = 3.25 seconds
- `step(8)` in MIDI_TIME mode → Each Wait(1) = 8 * (32nd note duration)
- The block body executes ONCE, not n times

**Solution:** Emit `SetStep(n)` followed by the block body statements, not wrapped in a loop.

### Issue 2: main() Function Execution and Nested Sequences

**Problem:** Wrapping the entire `main()` function in `RegisterSequence` caused nested sequence deadlock. When `mes()` blocks called `RegisterSequence` internally, the outer sequence would block waiting for completion, but the inner sequence couldn't execute because the outer one was still active.

**Correct Behavior:**
- `main()` function body should execute directly (not in a sequence)
- Only `mes()` blocks should call `RegisterSequence`
- This allows `mes()` blocks to register their own sequences without nesting issues

**Solution:** Execute `main()` function OpCodes directly using `ExecuteOpDirect`, not wrapped in `RegisterSequence`.

### Issue 3: vmLock Deadlock in PlayMIDI

**Problem:** `PlayMIDI` was called from `ExecuteOp`, which is called from `UpdateVM`. `UpdateVM` holds `vmLock`, but `PlayMIDI` tried to acquire `vmLock` again, causing a deadlock.

**Correct Behavior:**
- Functions called from `ExecuteOp` should NOT acquire `vmLock`
- `vmLock` is already held by `UpdateVM`
- Only top-level entry points should acquire `vmLock`

**Solution:** Remove `vmLock.Lock()` calls from `PlayMIDI` since it's called from within `UpdateVM`.

### Issue 4: MIDI Player Blocking

**Problem:** `midiPlayer.Play()` was blocking the main thread, preventing the game loop from continuing.

**Correct Behavior:**
- Audio playback should be asynchronous
- `midiPlayer.Play()` should not block the caller

**Solution:** Call `midiPlayer.Play()` in a goroutine to avoid blocking.

### Issue 5: MoveWin Hardcoded Size

**Problem:** `CallEngineFunction` for "movewin" with 2 arguments used hardcoded 640x480 size instead of the new picture's actual size.

**Correct Behavior:**
- `MoveWin(winID, picID)` should use the new picture's dimensions
- Window size should match the picture being displayed

**Solution:** Look up the new picture's size and use it for the window dimensions.

## Design Principles (Updated)

1. **Avoid Nested Locking:** Functions should document whether they expect locks to be held
2. **Direct Execution for Top-Level:** Only wrap code in sequences when timing control is needed
3. **Goroutines for Blocking Operations:** Use goroutines for operations that might block (audio, I/O)
4. **Dynamic Size Calculation:** Never hardcode dimensions; always query actual sizes
5. **Test with Real Samples:** Design assumptions should be validated with actual FILLY scripts
