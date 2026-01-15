# Design Document: Core Engine

## Overview

The son-et core engine is a source-to-source compiler (transpiler) that converts legacy FILLY scripts into modern Go source code. The system consists of two major components:

1. **Compiler Pipeline** - Lexer, Parser, AST, and Code Generator that transforms FILLY syntax into Go code
2. **Runtime Library** - An Ebitengine-based execution environment providing graphics, audio, timing, and script execution

The design prioritizes correctness, thread safety, and cross-platform compatibility while maintaining compatibility with legacy FILLY script semantics.

## Architecture

### System Components

#### 1. The Compiler (`cmd/son-et`)

The command-line tool responsible for converting user code.

**Components:**
- **Recursive Reader**: Handles `#include` directives to merge multiple source files
- **Lexer** (`pkg/compiler/lexer`): Tokenizes the input FILLY source code (`.tfy`, `.fil`)
- **Parser** (`pkg/compiler/parser`): Constructs an Abstract Syntax Tree (AST) from tokens
- **Code Generator** (`pkg/compiler/codegen`): Traverses the AST and emits Go source code
  - **Resource Embedding**: Scans usage of `LoadPic` and `PlayMIDI` to generate `//go:embed` directives
  - **VM Compilation**: Converts `mes(TIME)` blocks into bytecode/OpCodes for the VM

#### 2. The Runtime Library (`pkg/engine`)

The Go package that the generated code imports. It provides the game loop and runtime environment.

**Core Functions:**
- `engine.Init()`: Sets up the Ebitengine window (1280x720 Virtual Desktop) and initializes Audio
- `engine.Run()`: Starts the Game Loop
- `engine.UpdateVM()`: Executes the VM OpCodes for timing-critical sequences
- `engine.OpenWin/MoveCast`: Manages the Virtual Windows and Sprite rendering

### Data Flow

```
User writes game.tfy
        ↓
son-et game.tfy (Compile)
        ↓
Lexer → Parser → AST → Code Generator
        ↓
game_game.go (with //go:embed directives)
        ↓
go build game_game.go
        ↓
Standalone executable with embedded assets
```

### Directory Structure

```
/
├── cmd/
│   └── son-et/        # The Transpiler CLI Entrypoint
├── pkg/
│   ├── compiler/      # Transpiler Logic
│   │   ├── lexer/
│   │   ├── parser/
│   │   ├── ast/
│   │   └── codegen/
│   └── engine/        # Runtime Library (Ebitengine wrapper + VM)
├── samples/           # Example projects
└── docs/              # Documentation
```

## Concurrency Model

The engine operates using a **Producer-Consumer** concurrency model to bridge the gap between the procedural, blocking nature of legacy FILLY scripts and the event-driven, non-blocking nature of Ebitengine.

### Dual-Thread Architecture

**Main Thread (Ebitengine Loop):**
- Runs `Game.Update()` and `Game.Draw()` at 60 FPS
- Responsible for input handling and final rendering to the screen
- **Constraint**: Cannot be blocked. Must return quickly.

**Script Goroutine:**
- Runs the converted user code (`script()`)
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

## Variable Scope & VM Architecture

### The Problem: Dual Execution Model

FILLY scripts have a unified variable scope where variables defined anywhere in a function are accessible from `mes()` blocks. However, the current transpiler generates two different execution models:

1. **Outside `mes()` blocks**: Direct Go code execution
   - Variables are Go local variables in the generated `main()` function
   - Functions are called directly (e.g., `engine.LoadPic()`)
   - Example: `winw := engine.WinInfo(0)` creates a Go local variable
   
2. **Inside `mes()` blocks**: VM OpCode execution
   - Variables are stored in `Sequencer.vars` map
   - Functions are executed via `ExecuteOp()`
   - Example: `{Cmd: "Assign", Args: []any{"winw", ...}}` stores in VM

This creates a **scope isolation problem**: Variables defined outside `mes()` blocks (as Go locals) are not accessible inside them (VM variables).

**Real-World Example:**
```filly
main() {
    winW = WinInfo(0)  // Transpiled to: winw := engine.WinInfo(0)
    winH = WinInfo(1)  // Transpiled to: winh := engine.WinInfo(1)
    
    mes(MIDI_TIME) {
        // Tries to use winW, winH but they're not in VM scope!
        OpenWin(p39, winW-320, winH-240, 640, 480, 0, 0, 0)
    }
}
```

**Current Transpiler Output (Broken):**
```go
func main() {
    engine.Init(assets, func() {
        // Variables declared as Go locals (NOT accessible in VM)
        var winw int
        var winh int
        _ = winw
        _ = winh
        
        winw = engine.WinInfo(0)  // Go local variable
        winh = engine.WinInfo(1)  // Go local variable
        
        // mes() block registers OpCodes with VM
        engine.RegisterSequence(engine.MidiTime, []engine.OpCode{
            {Cmd: "OpenWin", Args: []any{
                engine.Variable("p39"),
                // Tries to read "winw" from VM, but it's not there!
                engine.OpCode{Cmd: "Infix", Args: []any{"-", engine.Variable("winw"), 320}},
                engine.OpCode{Cmd: "Infix", Args: []any{"-", engine.Variable("winh"), 240}},
                // ...
            }},
        }, map[string]any{})  // Empty initial vars map!
    })
}
```

**Why This Fails:**
1. `winw` and `winh` are Go local variables in the closure
2. `mes()` block tries to reference them via `engine.Variable("winw")`
3. `ResolveArg` looks for "winw" in `Sequencer.vars` map
4. Variable not found → returns 0 → window appears at wrong position

### The Solution: Hierarchical Scope Chain

**Design Principle**: Implement lexical scoping with parent scope lookup, similar to JavaScript's prototype chain or Python's LEGB rule.

**Architecture:**

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

**Key Insight**: The root Sequencer represents the `main()` function's variable scope, and child Sequencers (mes blocks) inherit from it.

### Implementation Details

**Phase 1: VM Parent Scope Support** ✅ (COMPLETED)

1. **Sequencer Structure** (pkg/engine/engine.go:1332):
```go
type Sequencer struct {
    commands     []OpCode
    pc           int
    vars         map[string]any
    parent       *Sequencer  // Parent scope reference for variable lookup
    mode         int
    onComplete   func()
    // ...
}
```

2. **Variable Resolution** (pkg/engine/engine.go:1490):
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
    // ...
    }
}
```

3. **Scope Creation** (pkg/engine/engine.go:1380):
```go
func RegisterSequence(mode int, ops []OpCode, initialVars ...map[string]any) {
    vmLock.Lock()
    
    // Initialize vars map
    vars := make(map[string]any)
    // Copy initial variables if provided
    if len(initialVars) > 0 {
        for k, v := range initialVars[0] {
            vars[strings.ToLower(k)] = v // Case-insensitive
        }
    }
    
    // Save current sequencer as parent
    parentSeq := mainSequencer
    
    mainSequencer = &Sequencer{
        commands:     ops,
        vars:         vars,
        parent:       parentSeq,  // Link to parent scope
        // ...
    }
    vmLock.Unlock()
}
```

4. **Helper Functions** (pkg/engine/engine.go:1381-1405):
```go
// SetVMVar sets a variable in the VM for use in mes() blocks
func SetVMVar(name string, value any) {
    vmLock.Lock()
    defer vmLock.Unlock()
    
    if mainSequencer == nil {
        // Create a root sequencer to hold variables
        mainSequencer = &Sequencer{
            vars:   make(map[string]any),
            parent: nil,
        }
    }
    
    mainSequencer.vars[strings.ToLower(name)] = value
}

// Assign is a helper function for transpiled code to set variables
// that need to be accessible in mes() blocks
func Assign(name string, value any) any {
    SetVMVar(name, value)
    return value
}
```

**Phase 2: Transpiler Variable Registration** (IN PROGRESS)

**Goal**: Modify the transpiler to register variables in the root Sequencer so they're accessible in mes() blocks.

**Current Transpiler Behavior** (pkg/compiler/codegen/codegen.go:626):
```go
case *ast.MesBlockStatement:
    // Collect variables used in the mes block
    usedVars := g.collectVariablesInBlock(s.Body)
    
    g.buf.WriteString("\tengine.RegisterSequence(")
    g.genExpression(s.Time)
    g.buf.WriteString(", []engine.OpCode{\n")
    g.genOpCodes(s.Body)
    g.buf.WriteString("\t}, map[string]any{")
    
    // Add used variables to the map
    first := true
    for _, varName := range usedVars {
        if !first {
            g.buf.WriteString(", ")
        }
        first = false
        g.buf.WriteString(fmt.Sprintf("%q: %s", varName, varName))
    }
    g.buf.WriteString("})\n")
```

**Problem**: This passes variables by value at the time of `RegisterSequence` call, but doesn't handle variables defined BEFORE the mes() block in the main function.

**Solution**: Use `engine.Assign()` for variable assignments that need to be accessible in mes() blocks.

**Target Transpiler Output:**
```go
func main() {
    engine.Init(assets, func() {
        // Variables declared as Go locals (for type safety)
        var winw int
        var winh int
        var p39 int
        _ = winw
        _ = winh
        _ = p39
        
        // Use Assign() to register in VM AND assign to Go local
        winw = engine.Assign("winW", engine.WinInfo(0)).(int)
        winh = engine.Assign("winH", engine.WinInfo(1)).(int)
        p39 = engine.Assign("p39", engine.LoadPic("P39.BMP")).(int)
        
        // mes() block can now access these variables via parent scope
        engine.RegisterSequence(engine.MidiTime, []engine.OpCode{
            {Cmd: "OpenWin", Args: []any{
                engine.Variable("p39"),  // Found in parent scope
                engine.OpCode{Cmd: "Infix", Args: []any{"-", engine.Variable("winw"), 320}},
                engine.OpCode{Cmd: "Infix", Args: []any{"-", engine.Variable("winh"), 240}},
                // ...
            }},
        }, map[string]any{})
    })
}
```

**Implementation Strategy:**

1. **Detect Variable Usage in mes() Blocks**:
   - During code generation, scan all mes() blocks in the function
   - Collect all variables referenced in mes() blocks
   - Mark these variables as "needs VM registration"

2. **Generate Assign() Calls**:
   - For marked variables, generate: `varname = engine.Assign("varname", value).(type)`
   - For unmarked variables, generate normal: `varname = value`

3. **Maintain Type Safety**:
   - Keep Go local variable declarations for type checking
   - Use type assertions on `Assign()` return value: `.(int)`, `.(string)`, etc.

**Phase 3: Full OpCode Generation** (FUTURE)

**Long-term Vision**: All FILLY code should be executed as OpCodes for a unified execution model.

**Target Architecture:**
```go
func main() {
    engine.Init(assets, func() {
        // NO Go local variables - everything is OpCodes
        engine.ExecuteScript([]engine.OpCode{
            {Cmd: "Assign", Args: []any{"winw", engine.OpCode{Cmd: "WinInfo", Args: []any{0}}}},
            {Cmd: "Assign", Args: []any{"winh", engine.OpCode{Cmd: "WinInfo", Args: []any{1}}}},
            {Cmd: "Assign", Args: []any{"p39", engine.OpCode{Cmd: "LoadPic", Args: []any{"P39.BMP"}}}},
            
            // mes() block is just a nested OpCode sequence
            {Cmd: "RegisterSequence", Args: []any{
                engine.MidiTime,
                []engine.OpCode{
                    {Cmd: "OpenWin", Args: []any{
                        engine.Variable("p39"),
                        engine.OpCode{Cmd: "Infix", Args: []any{"-", engine.Variable("winw"), 320}},
                        // ...
                    }},
                },
            }},
        })
    })
}
```

**Benefits:**
1. **Unified execution model**: No distinction between mes() and non-mes() code
2. **Proper scoping**: Variables accessible according to lexical scope rules
3. **Debuggability**: Can pause, inspect, and step through all code
4. **Consistency**: Same variable resolution logic everywhere
5. **Future-proof**: Supports nested scopes (if/for/while), closures, etc.

**Challenges:**
- Requires significant transpiler rewrite
- All control flow (if/for/while) must be OpCode-based
- Performance overhead of VM execution vs. native Go code
- Type safety is lost (everything is `any`)

**Decision**: Implement Phase 2 first (Assign() helper) as a pragmatic solution, defer Phase 3 until needed.

### Case Sensitivity

FILLY is case-insensitive for identifiers. The VM implements this by:
- Converting all variable names to lowercase on storage: `vars[strings.ToLower(name)] = value`
- Converting all variable lookups to lowercase: `varName := strings.ToLower(string(v))`
- This ensures `winW`, `winw`, `WINW` all refer to the same variable

**Example:**
```filly
winW = 640  // Stored as "winw"
mes(TIME) {
    OpenWin(0, WINW, 0, 640, 480, 0, 0, 0)  // Looks up "winw"
}
```

### Current Status

**✅ Phase 1 Complete**: VM supports parent scope lookup
- `Sequencer.parent` field added
- `ResolveArg` walks scope chain
- `RegisterSequence` links parent scope
- `SetVMVar` and `Assign` helper functions added

**🚧 Phase 2 In Progress**: Transpiler needs modification
- Need to detect variables used in mes() blocks
- Need to generate `engine.Assign()` calls for those variables
- Need to maintain type safety with type assertions

**📋 Phase 3 Planned**: Full OpCode generation (future enhancement)
- Requires major transpiler rewrite
- Deferred until Phase 2 proves insufficient

## Audio & MIDI Strategy

### MIDI Handling

**Library**: `gitlab.com/gomidi/midi/v2` - Used to parse Standard MIDI Files

**Synthesis**: Software Synthesizer using `github.com/sinshu/go-meltysynth` (or similar)

**SoundFont**: Loads `.sf2` files (defaulting to `default.sf2`) to generate audio from MIDI events

**Integration**: The synth generates a PCM stream which is fed into Ebitengine's audio player

**Playback:**
- Plays Standard MIDI Files (.mid)
- **Looping**: Disabled by default (One-shot playback)
- **Sync**: Updates the global `targetTick` for MIDI Sync Mode consumers
- Works natively on **macOS** using CoreAudio

### WAVE Audio Playback

**Implementation**: Decodes full audio buffer and plays via Ebiten

**Characteristics:**
- Loads entire WAV file into memory
- Supports standard WAV formats (PCM, various sample rates)
- Allows multiple concurrent playbacks

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

## Components and Interfaces

### Picture Manager

**Responsibilities:**
- Load BMP files from embedded assets
- Create empty image buffers
- Copy pixels between pictures with optional transparency
- Track picture dimensions
- Release picture resources

**Key Functions:**
- `LoadPic(filename)` → Picture ID
- `CreatePic(width, height)` → Picture ID
- `MovePic(src, srcX, srcY, w, h, dst, dstX, dstY, mode)`
- `DelPic(picID)`
- `PicWidth(picID)`, `PicHeight(picID)`

### Cast Manager

**Responsibilities:**
- Create sprites with transparency
- Track creation order for z-ordering
- Update sprite positions and re-render
- Support sprite sheet clipping
- Manage cast lifecycle

**Key Functions:**
- `PutCast(srcPic, dstPic, x, y, transparentColor, ..., w, h, srcX, srcY)` → Cast ID
- `MoveCast(castID, pic, x, y, w, h, srcX, srcY)`
- `DelCast(castID)`

**Data Structure:**
```go
type Cast struct {
    SrcPicID int
    DstPicID int
    X, Y     int
    W, H     int
    SrcX, SrcY int
    TransparentColor color.Color
}

var casts map[int]*Cast
var castDrawOrder []int  // Maintains creation order
```

**CRITICAL: Cast Transparency Implementation**

**Design Principle: Separation of Concerns**
- **Picture (Pic)**: Raw image data loaded from files. Never modified after loading.
- **Cast**: A sprite that references a Picture, with optional transparency processing.

**Transparency Processing Strategy:**

When `PutCast` is called with a transparent color parameter:

1. **One-Time Processing at Cast Creation:**
   - Create a NEW transparency-processed image using `convertTransparentColor()`
   - Store this processed image as a NEW Picture with a new Picture ID
   - The Cast references this processed Picture ID, NOT the original
   - This happens ONCE when the Cast is created

2. **Drawing (PutCast and MoveCast):**
   - Simply draw the transparency-processed Picture
   - NO additional transparency processing needed
   - Ebitengine's native alpha blending handles the transparency automatically

3. **Performance Considerations:**
   - Transparency processing is expensive (loops through all pixels)
   - MUST be done only ONCE at Cast creation time
   - NEVER process transparency on every draw call
   - The processed image is reused for all subsequent draws

**Example Flow:**
```
LoadPic("sprite.bmp") → Picture ID 17 (original image)
PutCast(17, dest, x, y, 0xffffff, ...) → 
  1. convertTransparentColor(Picture 17, white) → new image
  2. Store as Picture ID 28 (transparency-processed)
  3. Create Cast ID 2 with Picture=28
  4. Draw Picture 28 to destination
MoveCast(2, ...) →
  1. Cast 2 references Picture 28 (already processed)
  2. Draw Picture 28 to destination (no processing needed)
```

**Common Mistakes to Avoid:**
- ❌ Processing transparency on every draw call (performance issue)
- ❌ Modifying the original Picture (violates separation of concerns)
- ❌ Storing transparency info in Cast but processing at draw time (inefficient)
- ✅ Process transparency ONCE at Cast creation, store as new Picture
- ✅ Keep original Picture unchanged
- ✅ Draw the processed Picture directly (no additional processing)

### Window Manager

**Responsibilities:**
- Create virtual windows within the desktop
- Update window properties (position, size, picture)
- Manage window lifecycle
- Set window captions

**Key Functions:**
- `OpenWin(pic, x, y, width, height, picX, picY, bgColor)` → Window ID
- `MoveWin(winID, pic, x, y, width, height, picX, picY)`
- `CloseWin(winID)`
- `CloseWinAll()`
- `CapTitle(winID, title)`

### Text Renderer

**Responsibilities:**
- Load fonts with specified attributes
- Render text to pictures
- Manage text color and background
- Support Japanese fonts

**Key Functions:**
- `SetFont(size, fontName, charset, ...)`
- `TextWrite(text, picID, x, y)`
- `TextColor(r, g, b)`
- `BgColor(r, g, b)`
- `BackMode(mode)` - 0: opaque, 1: transparent

### Timing System (VM)

**Responsibilities:**
- Execute timing-synchronized commands
- Manage MIDI sync and time mode
- Handle Wait operations
- Drive mes block execution

**Key Functions:**
- `RegisterSequence(mode, operations)` - Queue operations for VM
- `UpdateVM()` - Called each frame to advance VM state
- `SetStep(count)` - Set step resolution
- `Wait(steps)` - Wait for specified steps

**Data Structures:**
```go
type Operation struct {
    Command string
    Args    []interface{}
}

type Sequence struct {
    Operations []Operation
    PC         int  // Program Counter
    StepSize   int
    Mode       int  // TIME or MIDI_TIME
}
```

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
    Operations []Operation
    PC         int
    StepSize   int
    Mode       int
    TickCount  int
    WaitUntil  int
    Completed  bool
}
```

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system—essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*


### Property 1: Transpiler generates valid Go code

*For any* valid FILLY script, transpiling and compiling the generated Go code should succeed without compilation errors.

**Validates: Requirements 1.1, 1.2**

### Property 2: Case-insensitive identifier transformation

*For any* FILLY identifier with mixed case, the generated Go code should contain the lowercase version of that identifier.

**Validates: Requirements 1.4**

### Property 3: Asset embedding completeness

*For any* FILLY script containing LoadPic(), PlayMIDI(), or PlayWAVE() calls, the generated Go code should contain go:embed directives for all referenced asset files.

**Validates: Requirements 2.1, 2.2, 2.3**

### Property 4: Case-insensitive asset matching

*For any* asset filename reference with different casing than the actual file, the transpiler should correctly match and embed the asset.

**Validates: Requirements 2.4**

### Property 5: Sequential ID assignment

*For any* sequence of LoadPic or PutCast calls, the returned IDs should be sequential starting from 0 (0, 1, 2, ...).

**Validates: Requirements 4.2, 5.2**

### Property 6: Window creation and properties

*For any* valid OpenWin parameters (picture, position, size), a window should be created with exactly those properties.

**Validates: Requirements 3.2, 14.1, 14.2**

### Property 7: Creation order rendering

*For any* sequence of window or cast creations, the rendering order should match the creation order (first created = bottom layer, last created = top layer).

**Validates: Requirements 3.3, 5.5**

### Property 8: Resource cleanup on deletion

*For any* picture, cast, or window, after calling the corresponding Del function, that resource ID should no longer be valid and should not appear in subsequent operations.

**Validates: Requirements 4.5, 5.7, 14.4**

### Property 9: Picture dimension queries

*For any* loaded or created picture, PicWidth and PicHeight should return the exact dimensions of that picture.

**Validates: Requirements 4.6**

### Property 10: Sprite clipping correctness

*For any* cast with SubImage clipping parameters (srcX, srcY, w, h), the rendered sprite should display only the specified region of the source picture.

**Validates: Requirements 5.4**

### Property 11: Transparency key behavior

*For any* sprite, pixels matching the top-left pixel color (or specified transparent color) should not be rendered (transparent).

**Validates: Requirements 5.6**

### Property 12: MIDI Sync Mode non-blocking

*For any* mes(MIDI_TIME) block, RegisterSequence should return immediately without blocking, allowing subsequent code (like PlayMIDI) to execute.

**Validates: Requirements 10.5**

### Property 13: Time Mode blocking

*For any* mes(TIME) block, RegisterSequence should block until the sequence completes, ensuring sequential execution.

**Validates: Requirements 11.5, 11.6**

### Property 14: Step interpretation in MIDI Sync Mode

*For any* step(n) call in MIDI Sync Mode, the wait duration should be n * (32nd note duration based on MIDI tempo).

**Validates: Requirements 10.3**

### Property 15: Step interpretation in Time Mode

*For any* step(n) call in Time Mode, the wait duration should be approximately n * 50ms (±frame time tolerance).

**Validates: Requirements 11.3**

### Property 16: String operation correctness

*For any* string s, StrLen(s) should return the length, SubStr(s, pos, len) should return the substring from pos with length len, and StrFind(s, search) should return the first index of search in s (or -1 if not found).

**Validates: Requirements 13.1, 13.2, 13.3**

### Property 17: String formatting correctness

*For any* format string and values, StrPrint should produce output matching the format specifiers (%s for strings, %ld for decimal integers, %lx for hexadecimal integers).

**Validates: Requirements 13.4, 13.5**

### Property 18: Random number range

*For any* positive integer max, Random(max) should return a value in the range [0, max).

**Validates: Requirements 16.4**

### Property 19: Message system state tracking

*For any* mes block execution, GetMesNo(0) should return the last executed block number, and GetMesNo(1) should return the currently executing block number.

**Validates: Requirements 15.1, 15.2**

### Property 20: Script termination cleanup

*For any* script instance, after calling DelMe, DelUs, or DelAll, all associated resources (pictures, casts, windows) should be released and the script should stop executing.

**Validates: Requirements 17.1, 17.2, 17.3, 17.4**

### Property 21: Text rendering state persistence

*For any* sequence of TextColor, BgColor, and BackMode calls, subsequent TextWrite operations should use the most recently set values.

**Validates: Requirements 12.3, 12.4, 12.5**

### Property 22: Window state updates

*For any* window, after calling MoveWin with new parameters, the window should have exactly those new properties.

**Validates: Requirements 14.3**

### Property 23: Window caption updates

*For any* window, after calling CapTitle with a string, the window's caption should be that exact string.

**Validates: Requirements 14.6**

### Property 24: Multiple WAV concurrent playback

*For any* set of WAV files, calling PlayWAVE on each should result in all of them playing simultaneously without interference.

**Validates: Requirements 9.5**

### Property 25: MIDI playback single iteration

*For any* MIDI file, calling PlayMIDI should play the file exactly once and then stop (no looping).

**Validates: Requirements 8.4**

## Error Handling

### Transpiler Errors

**File Not Found:**
- When an asset file referenced in LoadPic/PlayMIDI/PlayWAVE cannot be found
- Action: Report error with filename and line number, halt compilation

**Syntax Errors:**
- When FILLY script contains invalid syntax
- Action: Report error with line number and context, halt compilation

**Type Inference Failures:**
- When variable usage is ambiguous or contradictory
- Action: Report error with variable name and conflicting usages, halt compilation

### Runtime Errors

**Invalid Resource IDs:**
- When a function is called with a non-existent picture/cast/window ID
- Action: Log error, return early without crashing

**File Loading Failures:**
- When LoadPic/PlayMIDI/PlayWAVE fails to load embedded asset
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

## Testing Strategy

### Dual Testing Approach

The core engine requires both **unit testing** and **property-based testing** to ensure correctness:

**Unit Tests:**
- Verify specific examples and edge cases
- Test integration points between components
- Validate error handling paths
- Test platform-specific behavior (macOS audio)

**Property-Based Tests:**
- Verify universal properties across all inputs
- Use randomized input generation for comprehensive coverage
- Run minimum 100 iterations per property test
- Each property test must reference its design document property

### Property-Based Testing Configuration

**Framework**: Use `testing/quick` (Go standard library) or `gopter` for more advanced features

**Test Structure:**
```go
func TestProperty1_TranspilerGeneratesValidGoCode(t *testing.T) {
    // Feature: core-engine, Property 1: Transpiler generates valid Go code
    // Validates: Requirements 1.1, 1.2
    
    property := func(script FILLYScript) bool {
        goCode := transpile(script)
        return compilesSuccessfully(goCode)
    }
    
    if err := quick.Check(property, &quick.Config{MaxCount: 100}); err != nil {
        t.Error(err)
    }
}
```

**Minimum Iterations**: 100 per property test (due to randomization)

**Tag Format**: Each test must include a comment:
```go
// Feature: core-engine, Property N: [property description]
// Validates: Requirements X.Y, X.Z
```

### Unit Testing Focus Areas

**Transpiler:**
- Lexer tokenization for various syntax elements
- Parser AST construction for complex expressions
- Code generator output for specific constructs
- Asset detection and embedding logic

**Runtime:**
- Picture loading and manipulation
- Cast creation and z-ordering
- Window management lifecycle
- Timing system mode switching
- Text rendering with various fonts
- String operations edge cases

**Concurrency:**
- Mutex acquisition patterns (use `-race` flag)
- Double-locking prevention
- Thread-safe resource access

**Audio:**
- MIDI file loading and playback
- WAV file decoding and playback
- Concurrent audio playback

### Integration Testing

**End-to-End Scenarios:**
- Transpile sample FILLY scripts and verify execution
- Test MIDI-synchronized animations
- Test time-based animations
- Test resource cleanup on script termination

**Platform Testing:**
- Verify macOS CoreAudio integration
- Test on different macOS versions
- Verify asset embedding in built executables

### Performance Testing

**Benchmarks:**
- Transpilation speed for large scripts
- Rendering performance with many casts
- Audio synthesis latency
- VM execution overhead

**Profiling:**
- CPU profiling for hot paths
- Memory profiling for resource usage
- Goroutine profiling for concurrency

### Regression Testing

**Test Suite Maintenance:**
- Add test for each bug fix
- Maintain sample FILLY scripts as test cases
- Track test coverage (aim for >80% for critical paths)

**Continuous Integration:**
- Run full test suite on every commit
- Run with `-race` flag to detect data races
- Verify builds on macOS

## Implementation Notes

### Critical Design Constraints

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

**Asset Embedding:**
- Perform case-insensitive file matching
- Generate go:embed directives for all referenced assets
- Verify assets exist at compile time

### Development Workflow

**Debugging Checklist:**
- [ ] Enable DEBUG_LEVEL=2 for detailed logging
- [ ] Use timestamped logs for timing issues
- [ ] Check for mutex double-locking patterns
- [ ] Verify timing mode (MIDI_TIME vs TIME)
- [ ] Run with `-race` flag for concurrency issues

**Common Mistakes to Avoid:**
1. Applying TIME logic to MIDI_TIME → Breaks MIDI sync
2. Making MIDI_TIME blocking → Deadlocks (PlayMIDI never runs)
3. Double-locking mutexes → Deadlock
4. Bootstrapping targetTick in MIDI mode → Fast-forward execution
5. Assuming understanding without verification → Test both modes!

### Future Enhancements

**Potential Improvements:**
- Support for additional image formats (PNG, JPEG)
- Hardware-accelerated rendering
- More sophisticated audio mixing
- Debugger integration for FILLY scripts
- Visual editor for sprite placement
- Hot-reload for faster development iteration
