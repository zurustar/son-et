# Design Document: son-et Core Engine

## Introduction

This document describes the architectural design for son-et, an interpreter and runtime for FILLY scripts. The design is based on the requirements defined in [requirements.md](requirements.md) and focuses on architectural principles and ideal implementation patterns.

**Design Philosophy**: This document describes how the system should be architected, not how it is currently implemented. It focuses on principles, abstractions, and clean architecture patterns that support the unique execution model of FILLY scripts.

## Glossary

See [GLOSSARY.md](GLOSSARY.md) for common terms used across all son-et specifications.

---

## Part 1: Core Architectural Principles

These principles form the foundation of the system architecture. Every design decision should align with these principles.

### Principle 1: Uniform OpCode-Based Execution

**Principle**: All FILLY code, regardless of complexity, must be represented and executed as OpCode sequences.

**Rationale**: A uniform execution model simplifies the VM, enables consistent variable scoping, and provides predictable timing behavior. This is not an optimization - it's the foundation that makes the step-based execution model possible.

**Design Implications**:
- The interpreter must convert all language constructs to OpCode during parsing
- The VM must have a single execution path: `ExecuteOp(opcode)`
- Expressions, statements, control flow, and code blocks are all OpCode
- No special-case execution paths for different language features

**OpCode Structure**:
```go
type OpCmd int  // Enum for type safety and performance

const (
    OpAssign OpCmd = iota
    OpCall
    OpIf
    OpFor
    OpRegisterSequence
    OpWait
    // ... other commands
)

type OpCode struct {
    Cmd  OpCmd      // Command type (enum, not string)
    Args []any      // Arguments (can contain nested OpCodes)
}

type Variable string  // Distinguishes variable references from literals
```

**Key Design Decisions**:
- Use enum types for commands (compile-time safety, better performance)
- Support nested OpCode in arguments (enables expression trees)
- Distinguish variables from literals using type wrappers
- Keep OpCode structure simple and uniform

### Principle 2: Event-Driven Step-Based Execution

**Principle**: Script execution advances in discrete steps synchronized with external events, not continuously.

**Rationale**: FILLY scripts are designed for multimedia applications where timing is critical. The step-based model allows precise synchronization with MIDI playback, frame updates, and user input. This is fundamentally different from traditional scripting languages.

**Design Implications**:
- The VM must be designed as a cooperative multitasking system
- Sequences yield control after each operation
- External events (ticks) drive execution forward
- Wait operations pause sequences for a specified number of steps
- Multiple sequences execute concurrently, each with independent state

**Execution Model**:
```
1. External event occurs (MIDI tick, frame update, user input)
2. VM increments tick counter
3. For each active sequence:
   a. Check if sequence is waiting (wait counter > 0)
   b. If waiting, decrement wait counter and continue to next sequence
   c. If not waiting, execute next operation
   d. If operation is Wait(n), set wait counter to n
   e. Advance program counter
4. Repeat on next event
```

**Key Design Decisions**:
- Sequences are independent execution units with their own state
- Each sequence has: program counter, wait counter, variable scope, active flag
- The VM never blocks waiting for a sequence to complete
- Sequences can be registered and forgotten (fire-and-forget pattern)

### Principle 3: Dual Timing Mode Architecture

**Principle**: The system must support two fundamentally different timing models with different execution strategies.

**Rationale**: FILLY supports both music-synchronized animations (MIDI_TIME) and procedural animations (TIME). These modes have different blocking behavior, tick sources, and step interpretations. Mixing their logic causes deadlocks and timing errors.

**MIDI_TIME Mode (Music Synchronization)**:
- **Tick Source**: MIDI playback callbacks
- **Blocking Behavior**: Non-blocking (RegisterSequence returns immediately)
- **Step Duration**: Musical time (32nd note duration, tempo-dependent)
- **Use Case**: Animations synchronized to music
- **Execution Order**: Allows PlayMIDI to be called after mes() block

**TIME Mode (Frame-Based Timing)**:
- **Tick Source**: 60 FPS game loop
- **Blocking Behavior**: Blocking (RegisterSequence waits for completion)
- **Game Loop Continuation**: Game loop continues during blocking (rendering, input handling)
- **Step Duration**: Real time (50ms units)
- **Use Case**: Procedural animations and sequential logic
- **Execution Order**: Ensures sequential execution (mes completes before next statement)

**Design Implications**:
- Separate code paths for MIDI_TIME and TIME modes
- Different tick generation strategies
- Different blocking semantics in RegisterSequence
- Never apply TIME logic to MIDI_TIME or vice versa

**Critical Rule**: Making MIDI_TIME blocking causes deadlock (PlayMIDI never executes). Making TIME non-blocking breaks sequential logic.

### Principle 4: Hierarchical Variable Scope

**Principle**: Variables must be resolved through a lexical scope chain with parent pointers, not a flat global namespace.

**Rationale**: FILLY uses lexical scoping where mes() blocks inherit the scope of their enclosing function. This requires a parent-pointer scope chain that allows variable lookup to walk up the hierarchy.

**Scope Chain Design**:
```
Root Scope (main function)
  ├─ Variable: x
  ├─ Variable: y
  └─ Child Scope (mes block in main)
       ├─ Variable: z
       └─ Parent pointer → Root Scope
```

**Variable Resolution Algorithm**:
```
1. Convert variable name to lowercase (case-insensitive)
2. Check current scope's variable map
3. If found, return value
4. If not found and parent exists, check parent scope
5. Repeat until found or root reached
6. If not found anywhere, return default value (0, "", [])
```

**Design Implications**:
- Each sequence must maintain a reference to its parent scope
- Variable assignment updates the scope where the variable was first declared
- mes() blocks created in a function inherit that function's scope
- Case-insensitive variable lookup throughout the chain

**Key Design Decisions**:
- Use parent pointers, not scope IDs or global lookups
- Lazy initialization (variables created on first assignment)
- Default values for undefined variables (no errors)
- Immutable parent references (set at sequence creation)

**Array Design**:
Arrays are dynamic integer arrays stored as Go slices in the variable map.

```go
type VariableValue interface{}  // Can be int, string, []int, or []string

// Variable storage in scope
vars := map[string]VariableValue{
    "x":      42,                      // int
    "name":   "FILLY",                 // string
    "scores": []int{85, 92, 78},       // integer array
    "names":  []string{"a", "b", "c"}, // string array
}
```

**Array Operations**:
- **Access**: `arr[index]` - auto-expand if index >= len(arr)
- **Assignment**: `arr[index] = value` - auto-expand if needed
- **Size**: `ArraySize(arr)` - returns current length
- **Insert**: `InsArrayAt(arr, index, value)` - insert at position
- **Delete**: `DelArrayAt(arr, index)` - remove at position
- **Clear**: `DelArrayAll(arr)` - remove all elements

**Auto-Expansion**:
```
When accessing arr[10] and len(arr) == 5:
1. Expand arr to length 11
2. Fill new elements [5..9] with 0
3. Set arr[10] = value (if assignment) or return 0 (if read)
```

**Key Design Decisions**:
- Arrays are Go slices (efficient, dynamic)
- Auto-expansion simplifies script authoring
- Zero-fill maintains predictable behavior
- Integer arrays store []int, string arrays store []string
- No multi-dimensional arrays

### Principle 5: Thread-Safe State Management

**Principle**: Shared state must be protected from concurrent access by the script goroutine and main thread.

**Rationale**: FILLY scripts execute in a separate goroutine while the main thread handles rendering and input. Graphics state (pictures, windows, casts) is accessed from both threads and must be protected.

**Concurrency Model**:
```
Main Thread:
  - Rendering (reads graphics state)
  - Input handling (reads/writes input state)
  - Game loop (drives tick generation)

Script Goroutine:
  - VM execution (reads/writes all state)
  - Asset loading (writes graphics state)
  - Function calls (reads/writes graphics state)
```

**Design Implications**:
- All graphics state must be protected by a mutex
- Rendering must acquire the mutex before reading state
- Script operations must acquire the mutex before modifying state
- Use double buffering to minimize lock contention
- Avoid holding locks during long operations

**Key Design Decisions**:
- Single render mutex for all graphics state
- Lock-free tick counter using atomic operations
- Separate mutexes for independent subsystems (audio, input)
- Minimize critical sections to reduce contention

### Principle 6: Non-Blocking Audio Architecture

**Principle**: Audio playback must run asynchronously without blocking the game loop or script execution.

**Rationale**: MIDI and WAV playback can take seconds or minutes. Blocking the game loop or script execution would freeze the entire application. Audio must run in background goroutines.

**Design Implications**:
- MIDI player runs in a separate goroutine
- WAV playback uses asynchronous audio APIs
- PlayMIDI and PlayWAVE return immediately
- MIDI tick callbacks invoke VM updates asynchronously
- Audio state is independent from script state

**Key Design Decisions**:
- MIDI player is a global background task
- MIDI ticks are delivered via callbacks, not polling
- Audio continues playing even if the sequence that started it terminates
- Graceful shutdown waits for audio to complete or times out

---

## Part 1.5: Concurrency and Threading Model

This section describes the threading architecture and synchronization mechanisms that ensure thread-safe execution.

### Threading Architecture Overview

son-etエンジンは、Ebitenゲームループとの統合において、スレッド安全性を確保するための特別な設計を採用しています。

### The Ebiten Constraint

**Critical Constraint**: Ebiten's `Image` type is **NOT thread-safe**. Concurrent access from multiple goroutines will corrupt internal state, causing rendering artifacts or crashes.

### Thread Structure

The engine uses three types of goroutines:

```
┌─────────────────────────────────────┐
│ Main Thread (Ebiten Game Loop)     │
├─────────────────────────────────────┤
│ Update()                            │
│  ├─ Process pending MIDI ticks     │
│  ├─ UpdateMIDISequences()          │
│  │   └─ MovePic, TextWrite, etc.   │
│  └─ UpdateVM()                     │
│                                     │
│ Draw()                              │
│  └─ RenderFrame()                  │
│      └─ Image.SubImage(), etc.     │
└─────────────────────────────────────┘

┌─────────────────────────────────────┐
│ MIDI Thread (Audio Callback)       │
├─────────────────────────────────────┤
│ MIDIStream.Read()                   │
│  ├─ Render audio samples           │
│  ├─ Calculate current tick         │
│  └─ Accumulate ticks               │
│     (NO image operations)           │
└─────────────────────────────────────┘

┌─────────────────────────────────────┐
│ WAV Threads (One per playback)     │
├─────────────────────────────────────┤
│ Audio playback only                 │
│ (No engine state access)            │
└─────────────────────────────────────┘
```

### The MIDI Tick Accumulation Pattern

**Problem**: MIDI callbacks run on the audio thread and need to trigger script execution that modifies images.

**Solution**: Accumulate ticks in the audio thread, process them on the main thread.

**Implementation**:

```go
// In MIDI callback (audio thread)
func (ms *MidiStream) Read(buf []byte) (int, error) {
    // Calculate ticks from wall-clock time
    currentTick := ms.calculateCurrentTick()
    ticksAdvanced := currentTick - ms.lastTick
    
    // Accumulate ticks (thread-safe)
    ms.engine.midiPlayer.midiTickMutex.Lock()
    ms.engine.midiPlayer.pendingMIDITicks += ticksAdvanced
    ms.engine.midiPlayer.midiTickMutex.Unlock()
    
    ms.lastTick = currentTick
    
    // Render audio samples
    ms.sequencer.Render(left, right)
    return n, nil
}

// In main thread (Update method)
func (e *Engine) Update() error {
    // Process accumulated MIDI ticks
    if e.midiPlayer != nil {
        e.midiPlayer.midiTickMutex.Lock()
        pendingTicks := e.midiPlayer.pendingMIDITicks
        e.midiPlayer.pendingMIDITicks = 0
        e.midiPlayer.midiTickMutex.Unlock()

        if pendingTicks > 0 {
            e.UpdateMIDISequences(pendingTicks)
        }
    }
    
    // ... rest of update logic
    return nil
}
```

**Key Design Decisions**:
- MIDI thread NEVER calls image operations directly
- Ticks are accumulated atomically using a mutex
- Main thread processes all accumulated ticks in batch
- Image operations (MovePic, TextWrite, etc.) only execute on main thread

### Synchronization Mechanisms

**renderMutex**:
Protects graphics state (pictures, windows, casts) from concurrent access:

```go
type EngineState struct {
    renderMutex sync.RWMutex
    pictures    map[int]*Picture
    windows     map[int]*Window
    casts       map[int]*Cast
    // ...
}

// In RenderFrame (main thread)
func (r *Renderer) RenderFrame(screen *ebiten.Image, state *EngineState) {
    state.renderMutex.RLock()
    defer state.renderMutex.RUnlock()
    // ... read graphics state for rendering
}

// In MovePic (main thread, called from UpdateMIDISequences)
func (e *Engine) MovePic(srcPicID, dstPicID int, ...) error {
    e.state.renderMutex.Lock()
    defer e.state.renderMutex.Unlock()
    // ... modify graphics state
}
```

**midiTickMutex**:
Protects the `pendingMIDITicks` counter:

```go
type MIDIPlayer struct {
    midiTickMutex    sync.Mutex
    pendingMIDITicks int64
    // ...
}
```

**Atomic Operations**:
Used for simple flags that don't require mutex protection:

```go
type Engine struct {
    programTerminated atomic.Bool
    // ...
}

// Can be safely accessed from any thread
if e.programTerminated.Load() {
    return ErrTerminated
}
```

### Thread Safety Rules

**Rule 1: All Image Operations on Main Thread**
Any function that modifies `ebiten.Image` MUST be called from the main thread (Update method).

**Functions that require main thread**:
- `MovePic` - draws one image onto another
- `TextWrite` - renders text onto an image
- `DrawLine`, `DrawCircle`, `DrawRect` - drawing primitives
- `PutCast`, `MoveCast` - sprite operations
- `LoadPicture` - creates new images

**Rule 2: Lock Before Accessing Graphics State**
Any function that reads or writes graphics state MUST acquire `renderMutex`:

```go
func (e *Engine) NewDrawingFunction(picID int, ...) error {
    // ALWAYS acquire lock first
    e.state.renderMutex.Lock()
    defer e.state.renderMutex.Unlock()

    pic := e.state.GetPicture(picID)
    // ... modify image ...
    
    return nil
}
```

**Rule 3: Never Call Image Operations from MIDI Thread**
MIDI callbacks must ONLY accumulate ticks, never call engine methods that modify images.

**Rule 4: Minimize Lock Duration**
Hold locks for the shortest time possible to reduce contention:

```go
// GOOD: Lock only during state access
e.state.renderMutex.Lock()
pic := e.state.GetPicture(picID)
e.state.renderMutex.Unlock()

// Process data without lock
processedData := expensiveOperation(pic)

e.state.renderMutex.Lock()
e.state.UpdatePicture(picID, processedData)
e.state.renderMutex.Unlock()

// BAD: Lock held during expensive operation
e.state.renderMutex.Lock()
pic := e.state.GetPicture(picID)
processedData := expensiveOperation(pic)  // Lock held too long!
e.state.UpdatePicture(picID, processedData)
e.state.renderMutex.Unlock()
```

### Why This Architecture Works

**Prevents Rendering Corruption**:
- All image operations execute on main thread
- No concurrent access to `ebiten.Image` internals
- Rendering and modification are properly serialized

**Maintains Timing Accuracy**:
- MIDI ticks calculated from wall-clock time (no drift)
- Tick accumulation ensures no ticks are lost
- Sequential processing maintains animation frame order

**Minimizes Lock Contention**:
- MIDI thread only holds lock briefly to update counter
- Main thread processes ticks in batch
- Rendering uses read lock (allows concurrent reads)

### Historical Context: The gomidi Failure

During development, we attempted to use separate goroutines for MIDI message scheduling and audio rendering. This approach failed catastrophically:

**Problem**: Notes were dropping during playback (音符が抜けている)
- MIDI messages scheduled in separate goroutine
- Audio samples generated in audio thread
- Asynchronous execution caused timing drift
- Result: Missing notes, broken melodies

**Solution**: Return to synchronous processing
- Use `MidiFileSequencer` for unified playback + synthesis
- MIDI messages processed atomically with audio rendering
- Perfect synchronization between events and audio

**Lesson**: For MIDI playback, message processing MUST be synchronous with audio rendering. Any asynchronous approach introduces timing drift.

### Guidelines for Adding New Features

When implementing new drawing or graphics functions:

1. **Ensure main thread execution**: Function must be called from `Update()` or a method called by `Update()`
2. **Acquire renderMutex**: Always lock before accessing graphics state
3. **Never call from MIDI thread**: MIDI callbacks must only accumulate ticks
4. **Test for corruption**: Run with MIDI playback to verify no rendering artifacts

**Example Template**:

```go
func (e *Engine) NewGraphicsFunction(picID int, ...) error {
    // 1. Acquire lock
    e.state.renderMutex.Lock()
    defer e.state.renderMutex.Unlock()

    // 2. Access graphics state
    pic := e.state.GetPicture(picID)
    if pic == nil {
        return fmt.Errorf("picture %d not found", picID)
    }

    // 3. Modify image (safe because we're on main thread with lock)
    // ... image operations ...
    
    return nil
}
```

### Performance Characteristics

**Lock Contention**:
- Minimal: MIDI thread holds lock for ~1μs per tick
- Main thread processes ticks in batch (one lock acquisition per frame)
- Rendering uses read lock (allows concurrent reads if needed)

**Tick Processing Latency**:
- Ticks processed within one frame (16.67ms at 60 FPS)
- Acceptable for music synchronization (humans perceive <50ms as simultaneous)
- Sequential tick delivery prevents animation frame skipping

**Memory Overhead**:
- `pendingMIDITicks`: 8 bytes (int64)
- Mutexes: ~8 bytes each
- Negligible compared to image data

---

## Part 2: System Architecture

This section describes the high-level architecture and component boundaries.

### Layered Architecture

The system is organized into distinct layers with clear responsibilities:

```
┌─────────────────────────────────────────────────────┐
│                  Application Layer                   │
│  (CLI, Embedded Executables, Test Harnesses)        │
└─────────────────────────────────────────────────────┘
                         │
┌─────────────────────────────────────────────────────┐
│                   Runtime Layer                      │
│  (Game Loop, Rendering, Input, Audio Playback)      │
└─────────────────────────────────────────────────────┘
                         │
┌─────────────────────────────────────────────────────┐
│                  Execution Layer                     │
│  (VM, Sequence Management, Tick Generation)          │
└─────────────────────────────────────────────────────┘
                         │
┌─────────────────────────────────────────────────────┐
│                 Compilation Layer                    │
│  (Lexer, Parser, AST, OpCode Generation)            │
└─────────────────────────────────────────────────────┘
                         │
┌─────────────────────────────────────────────────────┐
│                  Foundation Layer                    │
│  (State Management, Asset Loading, Interfaces)       │
└─────────────────────────────────────────────────────┘
```

**Layer Responsibilities**:

1. **Application Layer**: Entry points, command-line parsing, embedded project configuration
2. **Runtime Layer**: Platform integration, rendering, audio, input handling
3. **Execution Layer**: OpCode execution, sequence lifecycle, timing control
4. **Compilation Layer**: TFY parsing, AST construction, OpCode generation
5. **Foundation Layer**: Core data structures, dependency injection, interfaces

**Dependency Rules**:
- Higher layers depend on lower layers
- Lower layers never depend on higher layers
- Each layer exposes interfaces, not implementations
- Cross-layer communication uses dependency injection

### Component Boundaries

**Compiler Package** (`pkg/compiler/`):
- Preprocessor: Handles #info and #include directives
- Lexer: Tokenizes TFY source code
- Parser: Builds AST from tokens
- AST: Represents program structure
- Interpreter: Converts AST to OpCode sequences

**Preprocessor Design**:
```
1. Read main TFY file (raw bytes)
2. Detect and convert character encoding:
   a. Try to decode as UTF-8
   b. If UTF-8 decode fails, assume Shift-JIS
   c. Convert Shift-JIS to UTF-8 using golang.org/x/text/encoding/japanese
   d. If conversion fails, report error
3. Process #info directives (store metadata)
4. Process #include directives:
   a. Resolve file path (case-insensitive)
   b. Strip trailing comments from the directive (// or /* */)
   c. Check for circular includes
   d. Recursively preprocess included file (with encoding conversion)
   e. Insert processed content at #include location
5. Output preprocessed UTF-8 source to lexer
```

**Character Encoding Strategy**:
- **Detection**: Try UTF-8 first, fall back to Shift-JIS
- **Conversion**: Use `golang.org/x/text/transform` package
- **Rationale**: Most legacy FILLY scripts are Shift-JIS, but modern scripts may be UTF-8
- **Error Handling**: Clear error messages for unsupported encodings

**Key Design Decisions**:
- Preprocessing happens before lexing
- All internal processing uses UTF-8
- #info metadata stored separately (not in AST)
- #include creates a single merged source file
- #include supports trailing comments (stripped during parsing)
- Circular include detection uses file path tracking
- Case-insensitive file matching for Windows 3.1 compatibility
- Encoding conversion is transparent to the rest of the compiler

**Lexer Design**:
```
Input: Preprocessed UTF-8 source code
Output: Token stream with position tracking

Token Types:
- Keywords: if, for, while, mes, step, etc. (case-insensitive)
- Identifiers: variable names, function names
- Literals: integers, floats, strings
- Operators: +, -, *, /, ==, !=, <, >, etc.
- Delimiters: (, ), {, }, [, ], ,, ;
- Comments: // single-line, /* multi-line */

Position Tracking:
- Each token stores line and column numbers
- Enables precise error reporting
- Preserved through parsing for error messages
```

**Key Lexer Decisions**:
- Case-insensitive keyword matching (FILLY legacy)
- Position tracking for all tokens
- Single-pass tokenization
- No backtracking needed

**Parser Design**:
```
Input: Token stream from lexer
Output: Abstract Syntax Tree (AST)

Parsing Strategy:
- Recursive descent parser
- Precedence climbing for expressions
- Error recovery at statement boundaries

AST Node Types:
- Statements: Assignment, FunctionCall, If, For, While, Mes, VarDecl
- Expressions: BinaryOp, UnaryOp, Literal, Variable, ArrayAccess
- Declarations: FunctionDef, VarDecl

Variable Declaration Syntax:
- int x;              // Single integer variable
- int x, y, z;        // Multiple integer variables (comma-separated)
- int arr[];          // Integer array declaration
- str s;              // String variable
- str s1, s2;         // Multiple string variables
- str arr[];          // String array declaration

Control Flow Syntax:
- if (cond) { ... }
- if (cond) { ... } else { ... }
- if (cond) { ... } else if (cond) { ... } else { ... }
- Whitespace/newlines between if and else/else-if are allowed

Function Definition Syntax:
- name() { ... }                    // No 'function' keyword required
- name(int x) { ... }               // With typed parameters
- name(int x=0) { ... }             // With default values
- name(int arr[]) { ... }           // With typed array parameters
- name(p[], c[]) { ... }            // With untyped array parameters
- name(c, p[], x, y, l=10) { ... }  // Mixed: normal + array + default

Expression Precedence (highest to lowest):
1. Primary: literals, variables, array access, parentheses
2. Unary: -, !
3. Multiplicative: *, /, %
4. Additive: +, -
5. Relational: <, >, <=, >=
6. Equality: ==, !=
7. Logical AND: &&
8. Logical OR: ||
```

**Key Parser Decisions**:
- Precedence climbing for clean expression parsing
- Error messages include line/column from tokens
- AST is minimal and clean (no unnecessary nodes)
- Array syntax `arr[index]` parsed as ArrayAccess node
- Function definitions without 'function' keyword (FILLY legacy syntax)
- Comma-separated variable declarations in single statement
- String array declarations supported (`str arr[];`)
- else/else-if associated with nearest preceding if regardless of whitespace
- Array parameters in function declarations: `int arr[]`, `p[]`
- Default parameter values: `param=value`
- Mixed parameter types: normal, array, and default values can be combined
- Function call vs declaration disambiguation: checks for typed parameters or `{` after `)`

**OpCode Generation Design**:
```
Input: AST from parser
Output: OpCode sequences

Conversion Strategy:
- Recursive traversal of AST
- Each AST node converts to one or more OpCodes
- Nested expressions become nested OpCodes
- Control flow uses OpCode with nested blocks

Example Conversions:
- Assignment: x = 5 + 3
  → OpCode{Cmd: OpAssign, Args: [Variable("x"), OpCode{Cmd: OpBinaryOp, Args: ["+", 5, 3]}]}

- If statement: if (x > 5) { y = 10 }
  → OpCode{Cmd: OpIf, Args: [
      OpCode{Cmd: OpBinaryOp, Args: [">", Variable("x"), 5]},
      []OpCode{{Cmd: OpAssign, Args: [Variable("y"), 10]}},
      []OpCode{}  // empty else block
    ]}

- mes() block: mes(TIME) { step(10); }
  → OpCode{Cmd: OpRegisterSequence, Args: [
      "TIME",
      []OpCode{{Cmd: OpWait, Args: [10]}}
    ]}
```

**Key CodeGen Decisions**:
- All FILLY constructs convert to OpCode uniformly
- Nested OpCode structure preserves expression trees
- Control flow blocks are OpCode slices
- No special cases or optimizations (simplicity over performance)
- mes() blocks become OpRegisterSequence with nested OpCodes

**Engine Package** (`pkg/engine/`):
- State: Manages all runtime state (graphics, audio, variables)
- VM: Executes OpCode sequences
- Sequencer: Manages sequence lifecycle and scope
- Renderer: Abstracts rendering operations
- Audio: Manages MIDI and WAV playback
- Assets: Loads and decodes resources

**Interfaces**:
- `Renderer`: Abstracts rendering (enables headless mode)
- `AssetLoader`: Abstracts asset loading (enables testing)
- `ImageDecoder`: Abstracts image decoding (enables mocking)
- `TickGenerator`: Abstracts tick generation (enables testing)

---

## Part 3: Execution Model Design

This section describes the detailed design of the execution model.

### Sequence Lifecycle

**Sequence States**:
```
Created → Active → Waiting → Active → ... → Completed
                      ↓
                  Terminated (del_me)
```

**State Transitions**:
- **Created**: Sequence registered but not yet executed
- **Active**: Sequence executing operations
- **Waiting**: Sequence paused for n steps
- **Completed**: Sequence reached end of operations
- **Terminated**: Sequence explicitly terminated via del_me

**Sequence Structure**:
```go
type Sequencer struct {
    // Execution state
    commands []OpCode    // OpCode sequence to execute
    pc       int         // Program counter
    active   bool        // Is sequence active?
    
    // Timing state
    mode       int       // TIME or MIDI_TIME
    waitCount  int       // Steps remaining in current wait
    stepSize   int       // Duration of one step
    
    // Scope state
    vars     map[string]any  // Variables in this scope
    parent   *Sequencer      // Parent scope (for variable lookup)
    
    // Metadata
    id       int         // Unique sequence ID
    groupID  int         // Group ID (for del_us)
}
```

**Key Design Decisions**:
- Sequences are immutable after creation (commands, parent, mode)
- Mutable state is limited to: pc, active, waitCount, vars
- Parent pointer enables lexical scoping
- Group ID enables bulk termination (del_us)

### Tick-Driven Execution

**Tick Generation**:
- **TIME Mode**: Game loop generates ticks at 60 FPS
- **MIDI_TIME Mode**: MIDI player generates ticks based on tempo and PPQ

**Tick Processing**:
```
On each tick:
  1. Increment global tick counter
  2. For each active sequence:
     a. If sequence.waitCount > 0:
        - Decrement waitCount
        - Continue to next sequence
     b. If sequence.waitCount == 0:
        - Execute sequence.commands[sequence.pc]
        - Handle operation result:
          * Wait(n): Set waitCount = n
          * del_me: Set active = false
          * Normal: Advance pc
     c. If pc >= len(commands):
        - Set active = false (sequence complete)
  3. Check for program termination
```

**Key Design Decisions**:
- Tick counter is global and monotonically increasing
- Each sequence processes independently
- Wait operations are counted in ticks, not time
- Sequences can complete naturally or via del_me

### Concurrency Model

**Goroutine Structure**:
```
Main Goroutine:
  - Ebiten game loop (Update, Draw)
  - Rendering
  - Input handling

Script Goroutine:
  - main() function execution
  - Function calls
  - Asset loading

MIDI Goroutine:
  - MIDI playback
  - Tick callbacks to VM

WAV Goroutines:
  - One per concurrent WAV playback
```

**Synchronization**:
- Render mutex protects graphics state
- Atomic operations for tick counter
- Channels for MIDI tick delivery
- No locks during OpCode execution (single-threaded VM)

**Key Design Decisions**:
- VM execution is single-threaded (no locks needed)
- Only graphics state requires synchronization
- MIDI ticks are delivered asynchronously via atomic updates
- Script goroutine terminates when main() completes

---

## Part 4: Graphics System Design

This section describes the design of the graphics subsystem.

### Virtual Display Architecture

**Concept**: A fixed 1280×720 pixel virtual desktop that contains all windows.

**Design Rationale**:
- Provides consistent coordinate system across platforms
- Simplifies window positioning and clipping
- Matches original FILLY behavior
- Enables pixel-perfect rendering

**Virtual Desktop Structure**:
```
┌─────────────────────────────────────────┐
│     Virtual Desktop (1280×720)          │
│  ┌──────────────┐  ┌──────────────┐    │
│  │  Window 0    │  │  Window 1    │    │
│  │              │  │              │    │
│  │  ┌────────┐  │  │              │    │
│  │  │ Cast 0 │  │  │              │    │
│  │  └────────┘  │  │              │    │
│  └──────────────┘  └──────────────┘    │
└─────────────────────────────────────────┘
```

**Key Design Decisions**:
- Virtual desktop is always 1280×720 (not configurable)
- Windows are positioned within the virtual desktop
- Casts (sprites) are positioned within windows
- Rendering scales virtual desktop to actual screen size

### Resource Management

**Picture Management**:
- Pictures are loaded from files or created programmatically
- Each picture has a unique ID (sequential assignment)
- Pictures are reference-counted (shared between windows/casts)
- Pictures are immutable after creation (transformations create new pictures)

**Cast (Sprite) Management**:
- Casts reference pictures with optional clipping regions
- Casts maintain creation order for z-ordering (painter's algorithm)
- Casts support transparency via color key (TransparentColor field)
- Casts are positioned relative to their destination picture (WindowID field stores destPicID)

**Cast Double Buffering (MoveCast)**:
When `MoveCast` is called, the system uses double buffering to prevent cast accumulation artifacts:

```
1. Initialize BackBuffer if not exists (same size as destination picture)
2. Clear BackBuffer to transparent (RGBA 0,0,0,0)
3. Redraw ALL casts belonging to the destination picture onto BackBuffer
4. Swap BackBuffer with main Image (atomic pointer swap)
```

This ensures that casts are redrawn cleanly each frame without accumulating previous positions.

**Cast Data Structure**:
```go
type Cast struct {
    ID               int   // Unique cast ID
    PictureID        int   // Source picture ID
    WindowID         int   // Destination picture ID (legacy name, actually destPicID)
    X                int   // Position X relative to destination
    Y                int   // Position Y relative to destination
    SrcX             int   // Source clipping X
    SrcY             int   // Source clipping Y
    Width            int   // Clipping width
    Height           int   // Clipping height
    TransparentColor int   // Color key for transparency (0xRRGGBB, -1 = no transparency)
    Visible          bool  // Is cast visible
}
```

**Picture BackBuffer**:
```go
type Picture struct {
    ID         int         // Unique picture ID
    Image      image.Image // The actual image data (front buffer)
    BackBuffer image.Image // Double buffer for cast rendering
    Width      int         // Image width
    Height     int         // Image height
}
```

**Key Design Decisions**:
- BackBuffer is lazily initialized on first MoveCast call
- All casts on the destination picture are redrawn (not just the moved cast)
- Atomic buffer swap prevents visual artifacts
- Transparent clear ensures clean redraw without ghost images

**Window Management**:
- Windows display pictures with optional decorations
- Windows maintain creation order for z-ordering
- Windows support captions and resizing
- Windows clip casts to their boundaries

**Window Positioning with Picture Offsets (PicX/PicY)**:
- Windows can display a portion of a picture using PicX/PicY offsets
- **Legacy Compatibility**: Offsets are inverted (`-picX, -picY`) for compatibility with original FILLY
- This allows centering images larger than the window by using negative offsets
- Example: To center a 640×480 image in a 320×240 window, use PicX=-160, PicY=-120
- The renderer implements proper rectangle intersection to handle negative offsets correctly

**Critical Implementation Detail**:
```go
// In OpenWindow() - offsets must be inverted for legacy compatibility
window.SrcX = -picX  // Note the negation
window.SrcY = -picY  // Note the negation
```

This inversion is essential for correct image display in legacy FILLY scripts.

**Key Design Decisions**:
- Sequential ID assignment (simple, predictable)
- Creation order determines z-order (no explicit z-index)
- Immutable resources (transformations create new resources)
- Reference counting for memory management

### Rendering Pipeline

**Rendering Stages**:
```
1. Fill virtual desktop with background color (teal: RGB 31, 126, 127 / 0x1F7E7F)
2. For each window (in creation order):
   a. Draw window frame (classic desktop style):
      - Gray background (RGB 192, 192, 192)
      - Raised 3D border effect (light highlight on top/left, dark shadow on bottom/right)
      - Border thickness: 4px
   b. Draw title bar (if caption exists):
      - Blue background (RGB 0, 0, 128)
      - Height: 24px (TitleBarHeight constant)
      - White caption text
      - [DEBUG] Window ID label at right side of title bar (yellow text)
   c. Draw window content area:
      - First: Fill with window background color (from Window.Color field)
      - Then: Draw picture with clipping and transparency
      - [DEBUG] Picture ID label at top-left of content area (green text)
   d. For each cast in window (in creation order):
      - Apply clipping region
      - Apply transparency
      - Draw cast to window
      - [DEBUG] Cast ID label at cast position (yellow text)
3. Scale virtual desktop to screen
4. Present to display
```

**Window Background Color**:
The window background color is stored in the `Window.Color` field and drawn by the renderer BEFORE the picture content. This ensures that transparent areas in the picture show the background color instead of the desktop.

```go
// Window struct includes Color field
type Window struct {
    // ... other fields ...
    Color     color.Color // Background color (drawn behind picture content)
}

// In OpenWindow(), convert bgColor parameter to color.Color
if bgColor >= 0 {
    r := uint8((bgColor >> 16) & 0xFF)
    g := uint8((bgColor >> 8) & 0xFF)
    b := uint8(bgColor & 0xFF)
    win.Color = color.RGBA{R: r, G: g, B: b, A: 255}
} else {
    win.Color = color.RGBA{R: 255, G: 255, B: 255, A: 255} // Default white
}

// In renderer, draw background color before picture
if win.Color != nil {
    vector.DrawFilledRect(screen, contentX, contentY, win.Width, win.Height, win.Color)
}
// Then draw picture content on top
```

**Key Design Decision**: The background color is NOT applied to the picture itself (which would overwrite picture content). Instead, it's drawn by the renderer as a separate layer behind the picture.

**Window Decoration Constants**:
- `TitleBarHeight = 20` pixels
- `BorderThickness = 4` pixels
- Title bar color: RGB(0, 0, 128) - Dark blue
- Border color: RGB(192, 192, 192) - Gray
- Highlight color: RGB(255, 255, 255) - White (for raised edge effect)
- Shadow color: RGB(0, 0, 0) - Black (for recessed edge effect)
- Caption text color: RGB(255, 255, 255) - White

**Desktop Background Color**:
- The virtual desktop uses a distinctive teal color (0x1F7E7F) as its background
- This color is visible in areas not covered by windows
- Provides visual feedback that the engine is running
- Matches the original FILLY implementation aesthetic

**Debug Overlay (DEBUG_LEVEL >= 2)**:
When debug level is set to 2 or higher, the renderer displays ID labels for debugging:

| Element | Label Format | Position | Color |
|---------|-------------|----------|-------|
| Window | `[W{id}]` | Title bar, right side | Yellow (255, 255, 0) |
| Picture | `P{id}` | Content area, top-left | Green (0, 255, 0) |
| Cast | `C{id}(P{picID})` | Cast position | Yellow (255, 255, 0) |

This layout prevents overlap when pictures are drawn at position (0,0) - Window ID is in the title bar, while Picture ID is in the content area.

**Renderer Interface**:
```go
type Renderer interface {
    RenderFrame(screen *ebiten.Image, state *EngineState)
}
```

**Key Design Decisions**:
- Renderer reads state but never modifies it
- Rendering is stateless (no frame-to-frame dependencies)
- Double buffering prevents flicker
- Renderer is swappable (enables headless mode)
- Desktop background is rendered first, before any windows
- Window background color drawn before picture content (layered rendering)

---

## Part 5: Audio System Design

This section describes the design of the audio subsystem.

### Technology Stack

**Graphics and Game Loop**:
- **Ebitengine (github.com/hajimehoshi/ebiten/v2)**: 2D game engine
  - Provides game loop (Update/Draw at 60 FPS)
  - Cross-platform rendering
  - Input handling
  - Audio context for MIDI and WAV playback
  - **Environment Variable**: Set `EBITENGINE_GRAPHICS_LIBRARY=opengl` to use OpenGL mode and avoid Metal deprecation warnings on macOS

**MIDI Synthesis**:
- **MeltySynth (github.com/sinshu/go-meltysynth/meltysynth)**: Software MIDI synthesizer
  - Parses Standard MIDI Files (SMF) and extracts tempo, PPQ
  - Loads SoundFont (.sf2) files for instrument samples
  - Renders audio samples in real-time
  - Provides accurate MIDI file sequencing

**WAV Playback**:
- **Ebiten Audio/WAV (github.com/hajimehoshi/ebiten/v2/audio/wav)**: WAV decoding and playback
  - Decodes WAV files with sample rate conversion
  - Supports concurrent playback of multiple WAV files
  - Process MIDI events sequentially
  - Provides timing information for synchronization
- **FluidSynth Go bindings or alternative synthesizer**: MIDI synthesis with SF2 soundfonts
  - Converts MIDI events to audio
  - Supports General MIDI (GM) soundfonts
  - Real-time synthesis

**WAV Playback**:
- **Ebitengine audio (github.com/hajimehoshi/ebiten/v2/audio)**: WAV decoding and playback
  - Concurrent playback support
  - Cross-platform audio output

### MIDI Architecture

**MIDI Player Design (MeltySynth MidiFileSequencer)**:
- Global singleton (one MIDI file plays at a time)
- Runs in separate goroutine (audio thread)
- **Uses meltysynth.MidiFileSequencer for unified MIDI playback and synthesis**
- Implements io.Reader interface (MidiStream) for Ebiten audio integration
- Generates ticks based on wall-clock time and tempo map
- Invokes tick callbacks to drive MIDI_TIME sequences
- **Detects playback completion by comparing current tick to total ticks**

**Architecture Overview**:
```
┌─────────────────────────────────────────────────────────────┐
│                        MIDIPlayer                            │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │     meltysynth.MidiFileSequencer                      │  │
│  │  (Unified MIDI playback + synthesis)                  │  │
│  │                                                        │  │
│  │  • Parses MIDI file internally                        │  │
│  │  • Processes MIDI messages atomically with audio      │  │
│  │  • Generates audio samples                            │  │
│  └──────────────────────────────────────────────────────┘  │
│                           │                                  │
│                           v                                  │
│                  ┌─────────────┐                             │
│                  │ Audio       │                             │
│                  │ Samples     │                             │
│                  └─────────────┘                             │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │         WallClockTickGenerator                        │  │
│  │  (Converts wall-clock time → FILLY ticks)            │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Manual Tempo Map Parsing (parseMIDITempo)           │  │
│  │  • Extracts tempo events from MIDI file              │  │
│  │  • Calculates PPQ from time division                 │  │
│  │  • Adds default 120 BPM if no tempo events           │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

**Key Architectural Decisions**:
- **Uses**: `meltysynth.MidiFileSequencer` (combined playback + synthesis)
- **Uses**: `calculateMIDILength()` function (manual parsing for end detection)
- **Uses**: `parseMIDITempo()` function (manual parsing for tick generation)
- **Removed**: `gomidi` dependency (eliminated timing drift issues)
- **Removed**: `MIDIBridge` component (no longer needed)
- **Removed**: `playMIDIMessages()` goroutine (sequencer handles timing)
- **Removed**: `finishedChan` (completion detected via tick comparison)

**MIDI Tick Calculation**:
```
step(n) in MIDI_TIME mode = n × (PPQ / 8) MIDI ticks
  where:
    n = step count from script
    PPQ = pulses per quarter note (from MIDI file header)
    8 = 32nd notes per quarter note (FILLY's tick resolution)
    
Example: step(10) with PPQ=480 → 10 × (480/8) = 600 MIDI ticks
```

**Wall-Clock Time Based Timing**:
```
currentTick = CalculateTickFromTime(elapsed_seconds)
  where:
    elapsed_seconds = time.Since(startTime).Seconds()
    
Algorithm:
  1. Get elapsed time since playback started
  2. Find current tempo from tempo map
  3. Calculate MIDI ticks: elapsed * (tempo / 60) * PPQ
  4. Adjust for tempo changes throughout the file
  
Note: This calculates MIDI ticks directly, not 32nd notes.
The conversion to 32nd notes happens in step(n) calculation.
```

**Tempo Map Parsing**:
The `parseMIDITempo()` function manually parses MIDI files to extract tempo events:

```go
func parseMIDITempo(data []byte) ([]TempoEvent, int, error) {
    // Extract PPQ from MIDI header
    timeDivision := int(data[12])<<8 | int(data[13])
    ppq := 480 // default
    if timeDivision&0x8000 == 0 {
        ppq = timeDivision
    }
    
    // Scan all tracks for tempo meta events (0xFF 0x51)
    var events []TempoEvent
    // ... parse track chunks and extract tempo events ...
    
    // Add default 120 BPM at tick 0 if no tempo events found
    if len(events) == 0 {
        events = []TempoEvent{{Tick: 0, MicrosPerBeat: 500000}}
    } else if events[0].Tick > 0 {
        // Prepend default tempo if first event is not at tick 0
        events = append([]TempoEvent{{Tick: 0, MicrosPerBeat: 500000}}, events...)
    }
    
    return events, ppq, nil
}
```

**Critical Tempo Map Logic**:
- **Default tempo**: Only added if NO tempo events exist in the file
- **Prepend logic**: If first tempo event is at tick > 0, prepend default 120 BPM at tick 0
- **Rationale**: Prevents incorrect tempo for the first segment of the song
- **Bug fix**: Previous implementation always added default tempo, causing first segment to play at wrong speed

**Critical Implementation Details**:

**mes(MIDI_TIME) Execution**:
- `mes(MIDI_TIME)` blocks execute immediately (not registered as event handlers)
- This allows `PlayMIDI()` to be called inside the mes() block
- The sequence is registered and returns immediately (non-blocking)
- MIDI playback starts and drives tick generation

**MIDI Sequence Updates**:
- MIDI_TIME sequences have their wait counters decremented by MIDI ticks
- Separate `UpdateMIDISequences()` method handles MIDI tick updates
- TIME sequences use frame ticks, MIDI_TIME sequences use MIDI ticks
- Never mix timing modes (causes incorrect wait behavior)

**Rationale for Wall-Clock Time**:
- **Accuracy**: No cumulative drift from audio buffer processing delays
- **Determinism**: Same elapsed time always produces same tick
- **Tempo-awareness**: Properly handles tempo changes via tempo map
- **Buffer independence**: Works correctly regardless of buffer size

**MIDI Playback Lifecycle**:
```
0. Engine initialization:
   a. Check project directory for SoundFont files
   b. Search order: default.sf2, GeneralUser-GS.sf2, *.sf2
   c. If found, automatically load SoundFont
   d. If not found, MIDI playback will fail until LoadSoundFont() is called
1. PlayMIDI(filename) called
2. Load MIDI file and parse using meltysynth.NewMidiFile()
3. Parse tempo map and PPQ manually using parseMIDITempo()
4. Calculate total ticks using calculateMIDILength()
5. Verify SoundFont is loaded (error if not)
6. Create meltysynth.Synthesizer with SoundFont
7. Create meltysynth.MidiFileSequencer with Synthesizer
8. Call sequencer.Play(midiFile, false) to start playback
9. Create WallClockTickGenerator with tempo map and PPQ
10. Create MidiStream (implements io.Reader):
    a. Wraps meltysynth.MidiFileSequencer
    b. Stores tempo map, PPQ, and totalTicks
    c. Records start time (wall-clock)
    d. Tracks lastTick for sequential delivery
11. Create Ebiten audio player with MidiStream
12. Start playback in goroutine (non-blocking)
13. MidiStream.Read() called by audio thread:
    a. Render audio samples via sequencer.Render(left, right)
    b. Calculate current tick from elapsed time
    c. Deliver all ticks from lastTick+1 to currentTick sequentially
    d. Check if currentTick >= totalTicks for end detection
    e. If ended, trigger MIDI_END event and set isFinished flag
14. On completion:
    a. MIDI_END event triggered
    b. isFinished flag set to true
    c. isPlaying flag set to false
    d. Stop tick generation
```

**Unified Playback and Synthesis**:
The `MidiFileSequencer` processes MIDI messages atomically with audio sample generation:

```go
// In MidiStream.Read()
left := make([]float32, sampleCount)
right := make([]float32, sampleCount)
ms.sequencer.Render(left, right)  // Processes MIDI messages + generates audio
```

This ensures perfect synchronization between MIDI message processing and audio rendering, eliminating the timing drift that occurred with separate goroutines.

**Termination Detection**:
The engine uses an `isFinished` flag to track MIDI completion:

```go
// In MidiStream.Read()
if int64(currentMIDITick) >= ms.totalTicks && !ms.endReported {
    ms.endReported = true
    
    // Mark MIDI as finished in the player
    if ms.engine.midiPlayer != nil {
        ms.engine.midiPlayer.mutex.Lock()
        ms.engine.midiPlayer.isFinished = true
        ms.engine.midiPlayer.isPlaying = false
        ms.engine.midiPlayer.mutex.Unlock()
    }
    
    // Trigger MIDI_END event
    ms.engine.TriggerEvent(EventMIDI_END, &EventData{})
}

// In Engine.AllSequencesComplete()
if e.midiPlayer != nil && e.midiPlayer.IsPlaying() && !e.midiPlayer.IsFinished() {
    return false  // MIDI still playing, don't terminate
}
```

**Critical Termination Logic**:
- Check `IsFinished()` BEFORE `IsPlaying()` to avoid race condition
- `IsPlaying()` may return true while audio buffer drains after MIDI ends
- `IsFinished()` provides definitive completion status
- Prevents premature termination while MIDI is still active

**SoundFont Auto-Loading**:
The engine automatically searches for and loads a SoundFont file during initialization:
- **Search locations**: Project directory (where TFY file is located)
- **Search order**: 
  1. `default.sf2` (preferred name)
  2. `GeneralUser-GS.sf2` (common GM soundfont)
  3. Any `.sf2` file (first match)
- **Fallback**: If no SoundFont found, scripts can call `LoadSoundFont(filename)` explicitly
- **Error handling**: PlayMIDI() will fail with clear error if no SoundFont is loaded

**Sequential Tick Delivery**:
To prevent animation frame skipping, MidiStream.Read() delivers ALL ticks from lastDeliveredTick+1 to currentTick sequentially. This ensures that even if processing is delayed, no ticks are skipped.

Example: If lastDeliveredTick=100 and currentTick=105, Read() will call:
```
NotifyTick(101)
NotifyTick(102)
NotifyTick(103)
NotifyTick(104)
NotifyTick(105)
```

**Key Design Decisions**:
- **MidiFileSequencer for unified playback**: Eliminates timing drift from separate goroutines
- **Atomic message processing**: MIDI messages processed synchronously with audio rendering
- **Manual tempo parsing**: Provides control over tempo map construction
- **Wall-clock time for tick generation**: Prevents cumulative drift from buffer delays
- **isFinished flag for termination**: Reliable completion detection independent of audio buffer state
- MIDI player is independent from sequences
- Tick delivery uses wall-clock time (not sample counting) for accuracy
- Sequential tick delivery prevents frame skipping
- MidiStream implements io.Reader for Ebiten audio integration
- MIDI continues playing even if starting sequence terminates
- Only one MIDI file plays at a time (matches original behavior)
- SoundFont auto-loading improves user experience (no manual setup required)
- **Context-based cancellation**: Proper cleanup when engine context is cancelled

**Lessons Learned from gomidi Integration Attempt**:

During development, we attempted to use the `gomidi` library for MIDI playback control with `meltysynth` for synthesis only. This approach failed due to fundamental timing issues:

**Problem**: Notes were dropping during playback (音符が抜けている)
- MIDI messages were scheduled in a separate goroutine (`playMIDIMessages`)
- Audio samples were generated in the audio thread (`MidiStream.Read`)
- These two processes ran asynchronously with no synchronization
- Timing drift accumulated, causing MIDI messages to arrive late or be skipped
- Result: Audible gaps in music, missing notes, broken melodies

**Root Cause**: Asynchronous message scheduling
```
Time →
Audio Thread:    [Render samples] [Render samples] [Render samples]
Message Thread:  [Schedule msg]      [Schedule msg]    [Schedule msg]
                      ↓ drift ↓           ↓ drift ↓         ↓ drift ↓
```

**Solution**: Return to `MidiFileSequencer`
- MIDI message processing happens atomically with audio sample generation
- `sequencer.Render(left, right)` processes messages AND generates audio in one call
- No separate goroutine for message scheduling
- Perfect synchronization between MIDI events and audio output

**Why MidiFileSequencer Works**:
```
Time →
Audio Thread:    [Process MIDI + Render] [Process MIDI + Render] [Process MIDI + Render]
                      ↓ synchronized ↓         ↓ synchronized ↓         ↓ synchronized ↓
```

**Key Insight**: For MIDI playback, message processing MUST be synchronous with audio rendering. Any asynchronous approach introduces timing drift that manifests as dropped notes or incorrect timing.

**Time Investment**: ~10 hours of debugging to identify and resolve this issue.

**Architectural Principle**: When integrating MIDI playback, always prefer libraries that provide unified playback+synthesis over separate parsing+synthesis components.

### WAV Architecture

**WAV Player Design**:
- Multiple concurrent playback (one goroutine per WAV)
- Fire-and-forget (no lifecycle management)
- Uses platform audio APIs (Ebiten audio)
- No synchronization with VM execution

**WAV Playback Lifecycle**:
```
1. PlayWAVE(filename) called
2. Load and decode WAV file
3. Create audio player
4. Start playback (returns immediately)
5. Playback continues in background
6. Cleanup on completion
```

**Resource Management**:
- Preload: LoadRsc(id, filename) - loads WAV into memory
- Play: PlayRsc(id) - plays preloaded WAV
- Release: DelRsc(id) - frees preloaded WAV

**Key Design Decisions**:
- No limit on concurrent WAV playback
- No synchronization with script execution
- Preloading enables fast playback start
- Automatic cleanup on completion

---

## Part 6: User Input System Design

This section describes the design of the user input handling subsystem.

### Input Event Architecture

**Event Types**:
- **Keyboard Events**: `mes(KEY)` - triggered on any key press
- **Mouse Click Events**: `mes(CLICK)` - triggered on mouse click
- **Right Button Events**: `mes(RBDOWN)`, `mes(RBDBLCLK)` - right mouse button specific
- **Custom Events**: `mes(USER)` - triggered by `PostMes()`

**Event Delivery Model**:
```
1. Input event occurs (keyboard, mouse)
2. System identifies event type
3. System finds all matching mes() blocks
4. System triggers each matching sequence
5. Sequences execute asynchronously
```

**Key Design Decisions**:
- Events are delivered to all matching sequences (broadcast model)
- Event delivery is non-blocking (doesn't wait for sequences to complete)
- Multiple sequences can respond to the same event
- Event parameters are passed via global variables (MesP1, MesP2, etc.)

### Keyboard Input Design

**KEY Event Handling**:
- Triggered on any key press
- Key code available via system variable
- Multiple KEY sequences can be active
- No key-specific filtering (all keys trigger all KEY sequences)

**Key Design Decisions**:
- Simple broadcast model (no key filtering)
- Sequences check key code if needed
- Non-blocking delivery

### Mouse Input Design

**Mouse Event Types**:
- `CLICK`: General mouse click (typically left button)
- `RBDOWN`: Right button down
- `RBDBLCLK`: Right button double-click

**Mouse Position**:
- Mouse coordinates available via system variables
- Coordinates relative to virtual desktop (1280×720)

**Key Design Decisions**:
- Separate event types for different mouse actions
- Position information always available
- No button state tracking (event-driven only)

### ESC Key Special Handling

**ESC Key Behavior**:
The ESC key receives special treatment and is not delivered as a KEY event.

**Termination Flow**:
```
1. ESC key pressed
2. Set programTerminated flag (atomic)
3. On next Update() call:
   a. Check programTerminated flag BEFORE VM execution
   b. If set, return termination signal immediately
   c. Skip all OpCode execution
4. Game loop receives termination signal
5. Cleanup and exit
```

**Key Design Decisions**:
- ESC key bypasses normal event system
- Termination check happens before any OpCode execution
- Atomic flag prevents race conditions
- Immediate termination (no cleanup sequences)
- Cannot be overridden by scripts

**Rationale**: ESC key provides a guaranteed way for users to exit hung or misbehaving scripts. This is a safety feature that must always work.

### Reliable Timeout Architecture

**Problem Statement**:
Simple timeout implementations that only check elapsed time in the main Update() loop are unreliable. If the loop is blocked by long-running operations, infinite loops in scripts, or hung mes() blocks, the timeout check never executes. Additionally, background goroutines (MIDI/WAV players) may continue running after timeout.

**Solution**: Context-based cancellation architecture that propagates timeout signals to all components.

**Architecture Overview**:
```
┌─────────────────────────────────────────────────────────────┐
│                    Engine.Start()                            │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  context.WithTimeout(timeout)                        │    │
│  │         │                                            │    │
│  │         ├──► Timeout Monitor Goroutine               │    │
│  │         │         │                                  │    │
│  │         │         └──► Sets programTerminated        │    │
│  │         │                                            │    │
│  │         ├──► VM Execution (checks every 100 ops)     │    │
│  │         │                                            │    │
│  │         ├──► Loop Execution (checks every 100 iters) │    │
│  │         │                                            │    │
│  │         ├──► MIDI Player (monitors context.Done())   │    │
│  │         │                                            │    │
│  │         └──► WAV Players (monitor context.Done())    │    │
│  └─────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
```

**Key Components**:

1. **Context Management**:
```go
// Engine.Start() creates timeout context
func (e *Engine) Start() {
    if e.timeout > 0 {
        e.ctx, e.cancel = context.WithTimeout(context.Background(), e.timeout)
    } else {
        e.ctx, e.cancel = context.WithCancel(context.Background())
    }
    go e.monitorTimeout()
}

// Context accessible to all components
func (e *Engine) GetContext() context.Context {
    return e.ctx
}
```

2. **Timeout Monitor Goroutine**:
```go
func (e *Engine) monitorTimeout() {
    <-e.ctx.Done()
    if e.ctx.Err() == context.DeadlineExceeded {
        e.Log(1, "Timeout reached, terminating...")
        e.programTerminated.Store(true)
    }
}
```

3. **Periodic Termination Checks**:
```go
// In VM execution loop
func (vm *VM) ExecuteTick(seq *Sequencer) error {
    for iterations := 0; iterations < maxOperations; iterations++ {
        // Check termination every 100 operations
        if iterations%100 == 0 && vm.engine.CheckTermination() {
            return ErrTerminated
        }
        // ... execute operation
    }
    return nil
}

// In while/for loops
func (vm *VM) executeWhile(seq *Sequencer, op OpCode) error {
    for iterations := 0; iterations < maxLoopIterations; iterations++ {
        // Check termination every 100 iterations
        if iterations%100 == 0 && vm.engine.CheckTermination() {
            return ErrTerminated
        }
        // ... execute loop body
    }
    return ErrMaxIterations  // Safety limit reached
}
```

4. **Goroutine Cancellation**:
```go
// MIDI player monitors context
func (mp *MIDIPlayer) playbackLoop() {
    for {
        select {
        case <-mp.engine.GetContext().Done():
            mp.cleanup()
            return
        default:
            // Continue playback
        }
    }
}

// MIDI stream returns EOF on cancellation
func (ms *MidiStream) Read(buf []byte) (int, error) {
    select {
    case <-ms.engine.GetContext().Done():
        return 0, io.EOF
    default:
        // Continue reading
    }
}
```

**Safety Limits**:
| Component | Limit | Purpose |
|-----------|-------|---------|
| While loops | 100,000 iterations | Prevent infinite loops |
| For loops | 100,000 iterations | Prevent infinite loops |
| Top-level execution | 10,000 operations/tick | Prevent runaway scripts |
| VM per-tick | 1,000 operations | Ensure responsive updates |

**Timeout Guarantees**:
- Timeout triggers within specified duration ± 500ms
- All goroutines terminate when timeout occurs
- Resources are properly cleaned up
- Exit code 0 for normal termination, error for timeout

**Key Design Decisions**:
- Use Go's `context.WithTimeout()` for reliable cancellation
- Timeout signal propagates to all goroutines via context
- Periodic checks in execution loops catch hung scripts
- Safety limits prevent infinite loops even without timeout
- Graceful shutdown sequence ensures resource cleanup

### Event Sequence Lifecycle

**Registration**:
```go
// mes(KEY) { ... } becomes:
RegisterEventHandler(KEY, []OpCode{...})
```

**Triggering**:
```go
// On key press:
TriggerEvent(KEY, keyCode)
  -> Find all KEY handlers
  -> Create new sequence for each handler
  -> Execute sequences asynchronously
```

**Key Design Decisions**:
- Event handlers are registered once at startup
- Each event trigger creates a new sequence instance
- Sequences are independent (one can terminate without affecting others)
- No limit on concurrent event sequences

### Custom Message System

**PostMes Design**:
```filly
PostMes(messageType, p1, p2, p3, p4)
```

**Message Delivery**:
```
1. PostMes() called with parameters
2. System sets MesP1-MesP4 global variables
3. System finds all mes(USER) blocks matching messageType
4. System triggers each matching sequence
5. Sequences read parameters from MesP1-MesP4
```

**Key Design Decisions**:
- Parameters passed via global variables (legacy compatibility)
- Message type used for filtering
- Broadcast delivery to all matching handlers
- Non-blocking delivery

---

## Part 7: Text Rendering System Design

This section describes the design of the text rendering subsystem.

### Text Rendering Architecture

**Font Loading Strategy**:
The engine supports TrueType fonts with automatic fallback to system fonts for Japanese text rendering.

**Font Search Order (macOS)**:
1. `/System/Library/Fonts/ヒラギノ明朝 ProN.ttc` - Hiragino Mincho (serif)
2. `/System/Library/Fonts/ヒラギノ角ゴシック W3.ttc` - Hiragino Kaku Gothic (sans-serif, light)
3. `/System/Library/Fonts/ヒラギノ角ゴシック W4.ttc` - Hiragino Kaku Gothic (sans-serif, regular)
4. `/Library/Fonts/Arial Unicode.ttf` - Arial Unicode (fallback)
5. `/System/Library/Fonts/Supplemental/Arial Unicode.ttf` - Arial Unicode (system fallback)
6. `basicfont.Face7x13` - Built-in bitmap font (final fallback)

**Font Loading Implementation**:
```go
func (tr *TextRenderer) loadFont(path string, size float64) font.Face {
    // Read font file
    fontData, err := os.ReadFile(path)
    if err != nil {
        return nil
    }
    
    // Try single font (.ttf) first
    tt, err := opentype.Parse(fontData)
    if err != nil {
        // Try font collection (.ttc)
        collection, err := opentype.ParseCollection(fontData)
        if err != nil {
            return nil
        }
        // Use first font in collection
        tt, _ = collection.Font(0)
    }
    
    // Create font face with specified size
    face, err := opentype.NewFace(tt, &opentype.FaceOptions{
        Size:    size,
        DPI:     72,
        Hinting: font.HintingFull,
    })
    return face
}
```

**Anti-Aliasing Artifact Prevention**:
Text rendering uses alpha blending for smooth edges (anti-aliasing). When drawing text multiple times on the same area, semi-transparent pixels from the old text can create shadow artifacts.

**Solution**: Clear the text area with the current background color (`tr.bgColor`) before drawing new text. This respects the `BgColor()` setting from the script.

```go
func (tr *TextRenderer) TextWrite(text string, picID, x, y int) error {
    // Clear text area first (prevents anti-aliasing artifacts)
    // Use the background color set by BgColor() function
    textWidth := len(text) * tr.currentFontSize
    textHeight := tr.currentFontSize + 4
    clearColor := tr.bgColor  // Respects BgColor() setting
    
    for py := 0; py < textHeight; py++ {
        for px := 0; px < textWidth; px++ {
            if x+px >= 0 && x+px < pic.Width && y+py >= 0 && y+py < pic.Height {
                rgba.Set(x+px, y+py, clearColor)
            }
        }
    }
    
    // Now draw text (no artifacts from previous text)
    drawer := &font.Drawer{
        Dst:  rgba,
        Src:  image.NewUniform(tr.textColor),
        Face: tr.currentFont,
        Dot:  fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y + tr.currentFontSize)},
    }
    drawer.DrawString(text)
    
    return nil
}
```

**Key Design Decisions**:
- **System font integration**: Use native fonts for proper Japanese rendering
- **Font collection support**: Handle both .ttf (single font) and .ttc (font collection) files
- **Graceful fallback**: If system fonts unavailable, use built-in bitmap font
- **Anti-aliasing handling**: Clear area before drawing to prevent shadow artifacts
- **Background color respect**: Use `tr.bgColor` (not hardcoded white) for clearing, respecting script's `BgColor()` setting
- **Modern API usage**: Use `os.ReadFile` instead of deprecated `ioutil.ReadFile`

**Text Rendering State**:
```go
type TextRenderer struct {
    currentFont     font.Face      // Loaded TrueType font or basicfont
    currentFontSize int            // Font size in pixels
    currentFontName string         // Font name from script
    textColor       color.Color    // Text foreground color
    bgColor         color.Color    // Background color (for opaque mode)
    backMode        int            // 0=transparent, 1=opaque
    engine          *Engine        // Reference to engine for logging
}
```

**Font Size Handling**:
- Legacy scripts may pass unreasonable sizes (e.g., 640 instead of 14)
- If size > 200, treat as legacy parameter order issue and use default size (13)
- This maintains compatibility with older FILLY scripts

---

## Part 8: Testing Strategy

This section describes the testing approach and testability design.

### Testability Principles

**Principle 1: Dependency Injection**
- All external dependencies are injected via interfaces
- Enables mocking for unit tests
- Supports headless testing without GUI

**Principle 2: State Isolation**
- State is encapsulated in structs, not global variables
- Each test creates fresh state
- No test pollution or ordering dependencies

**Principle 3: Interface-Based Design**
- Core functionality exposed through interfaces
- Multiple implementations (production, mock, test)
- Enables testing without platform dependencies

### Test Levels

**Unit Tests**:
- Test individual functions and methods
- Use mocks for external dependencies
- Fast execution (milliseconds)
- No GUI or audio initialization

**Property-Based Tests**:
- Test universal properties across many inputs
- Use generators for test data
- Verify invariants and correctness properties
- Catch edge cases missed by example-based tests

**Integration Tests**:
- Test component interactions
- Use real implementations where possible
- May require GUI or audio initialization
- Slower execution (seconds)

**End-to-End Tests**:
- Test complete sample scripts
- Verify timing accuracy and behavior
- Use headless mode for automation
- Slowest execution (seconds to minutes)

### Mock Strategy

**Mockable Interfaces**:
- `Renderer`: Mock rendering without Ebiten
- `AssetLoader`: Mock asset loading without filesystem
- `ImageDecoder`: Mock image decoding without actual images
- `TickGenerator`: Mock tick generation for deterministic tests

**Mock Implementations**:
```go
type MockRenderer struct {
    RenderCount int
    LastState   *EngineState
}

type MockAssetLoader struct {
    Files map[string][]byte
}

type MockImageDecoder struct {
    Width, Height int
}
```

**Key Design Decisions**:
- Mocks record calls for verification
- Mocks provide deterministic behavior
- Mocks are simple (no complex logic)
- Mocks are reusable across tests

### Headless Testing

**Headless Mode Design**:
- Executes scripts without GUI
- Logs rendering operations instead of drawing
- Enables automated testing in CI/CD
- Supports timeout for automatic termination

**Headless Execution**:
```
1. Parse command-line flags (--headless, --timeout)
2. Initialize engine with MockRenderer
3. Execute script in background goroutine
4. Run tick loop at 60 FPS
5. Log all operations
6. Terminate on timeout or completion
```

**Key Design Decisions**:
- Headless mode uses same execution path as GUI mode
- Only rendering is mocked (logic is identical)
- Timestamped logging enables timing verification
- Automatic termination prevents orphaned processes

### Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system—essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

**Property 1: Engine Termination Completeness**

*For any* TFY script that completes all execution paths (including mes() blocks with no more scheduled events), the engine should terminate within 1 second and return exit code 0, without entering an infinite wait state.

**Validates: Requirements A7.1, A7.3, A7.4, A7.5**

**Property 2: Sequencer Step Progression**

*For any* mes() block containing step() sequences, the sequencer should execute each step in order without skipping or hanging on any valid step index.

**Validates: Requirements A2.3, A2.4, A2.5**

**Property 3: Step Timing Accuracy**

*For any* step scheduled for a specific tick or time in a mes() block, the sequencer should execute it at the correct moment according to the timing mode (TIME or MIDI_TIME).

**Validates: Requirements A3.4, A3.9**

**Property 4: Step Block Completion Detection**

*For any* mes() block, when all steps are complete, the sequencer should mark the block as finished and allow the engine to proceed with termination checks.

**Validates: Requirements A2.5, A7.4**

**Property 5: Animation Frame Completeness**

*For any* sequence of MovePic() or MoveCast() commands in an animation, the engine should execute all animation frames in the correct order without skipping frames due to timing or rendering issues.

**Validates: Requirements B1.12, B1.16, B1.17**

**Property 6: Text Encoding Preservation**

*For any* Japanese text string in a TFY script, the text should be correctly converted from Shift-JIS to UTF-8 during preprocessing, preserved through compilation, and rendered without mojibake (garbled characters) when displayed via TextWrite().

**Validates: Requirements A1.17, A1.18, A1.19, A1.20, B5.5**

**Property 7: Timeout Reliability**

*For any* execution with a specified timeout duration T, the engine should terminate within T ± 500ms, regardless of whether the script contains infinite loops, hung mes() blocks, or long-running MIDI operations. All goroutines should be cancelled and resources cleaned up.

**Validates: Requirements A7.6, A7.7, A7.8, A7.9**

**Property-Based Testing Configuration**:
- Use Go's testing/quick package or a PBT library like gopter
- Minimum 100 iterations per property test
- Each test tagged with: **Feature: core-engine, Property {number}: {property_text}**

**Test Files**:
- `pkg/engine/termination_property_test.go` - Property 1
- `pkg/engine/step_progression_property_test.go` - Property 2
- `pkg/engine/step_timing_property_test.go` - Property 3
- `pkg/engine/step_completion_property_test.go` - Property 4
- `pkg/engine/animation_property_test.go` - Property 5
- `pkg/engine/text_test.go` - Property 6 (encoding tests)
- `pkg/engine/timeout_reliability_test.go` - Property 7

---

## Part 8: Error Handling Philosophy

This section describes the error handling approach.

### Error Handling Principles

**Principle 1: Fail Fast**
- Detect errors as early as possible
- Report errors with clear context
- Don't continue execution with invalid state

**Principle 2: Graceful Degradation**
- Undefined variables return default values (0, "", [])
- Missing assets log warnings but don't crash
- Invalid operations log errors but continue execution

**Principle 3: Clear Error Messages**
- Include file names and line numbers for parsing errors
- Include OpCode command and arguments for runtime errors
- Include asset names for loading errors

### Error Categories

**Parsing Errors** (Fail Fast):
- Syntax errors in TFY files
- Invalid function definitions
- Malformed expressions
- Report: filename, line, column, error message

**Runtime Errors** (Graceful Degradation):
- Undefined variables (return default value)
- Invalid function calls (log warning, skip operation)
- Asset loading failures (log error, use placeholder)
- Report: OpCode command, arguments, error message

**Resource Errors** (Graceful Degradation):
- Missing image files (log error, create empty image)
- Missing MIDI files (log error, skip playback)
- Missing WAV files (log error, skip playback)
- Report: asset name, error message

**Key Design Decisions**:
- Parsing errors prevent execution (fail fast)
- Runtime errors allow continued execution (graceful degradation)
- All errors are logged with context
- No silent failures (always log)

### Legacy Function Stubs

**Problem Statement**:
Legacy FILLY scripts may contain calls to Windows-specific functions that cannot be implemented on modern cross-platform systems. Without stub implementations, these scripts fail to parse, preventing execution of otherwise valid content.

**Solution**: Provide stub implementations that allow scripts to parse and execute while logging warnings for debugging.

**Stubbed Functions**:

| Function | Original Purpose | Stub Behavior |
|----------|-----------------|---------------|
| `Shell(cmd)` | Launch external Windows programs | Log warning, return 0 |
| `MCI(cmd)` | Windows Media Control Interface | Log warning, return 0 |
| `StrMCI(cmd)` | MCI with string return | Log warning, return "" |
| `GetIniStr(sec, key, def, file)` | Read INI/Registry | Return default value |

**Implementation Strategy**:

```go
// In VM.executeBuiltinFunction()
case "shell":
    vm.engine.Log(1, "WARNING: Shell() not supported on this platform")
    return 0, nil

case "mci":
    vm.engine.Log(1, "WARNING: MCI() not supported on this platform")
    return 0, nil

case "strmci":
    vm.engine.Log(1, "WARNING: StrMCI() not supported on this platform")
    return "", nil

case "getinistr":
    // Return the default value (3rd argument)
    if len(args) >= 3 {
        return args[2], nil
    }
    return "", nil
```

**Key Design Decisions**:
- Stubs are implemented in the VM, not the parser
- All stub calls log warnings at debug level 1
- Return values are safe defaults (0, "", or provided default)
- Scripts continue execution after stub calls
- No special parsing required (functions are treated as normal calls)

**Rationale**: This approach maximizes compatibility with legacy content while providing clear feedback to developers about unsupported functionality.

---

## Part 9: Implementation Guidelines

This section provides guidance for implementing the design.

### Code Organization

**Package Structure**:
```
son-et/
├── cmd/
│   └── son-et/          # Application entry point
├── pkg/
│   ├── compiler/        # Compilation layer
│   │   ├── lexer/       # Tokenization
│   │   ├── parser/      # AST construction
│   │   ├── ast/         # AST definitions
│   │   └── interpreter/ # OpCode generation
│   └── engine/          # Runtime layer
│       ├── state.go     # State management
│       ├── vm.go        # VM execution
│       ├── sequencer.go # Sequence management
│       ├── renderer.go  # Rendering abstraction
│       ├── audio.go     # Audio playback
│       └── interfaces.go # Interface definitions
└── samples/             # Test scripts
```

**Key Design Decisions**:
- Clear separation between compilation and runtime
- Interfaces defined in separate file
- State management centralized
- Test files colocated with implementation

### Naming Conventions

**Types**:
- Interfaces: `Renderer`, `AssetLoader`, `ImageDecoder`
- Structs: `EngineState`, `Sequencer`, `OpCode`
- Enums: `OpCmd`, `TimingMode`

**Functions**:
- Public: `ExecuteOp`, `RegisterSequence`, `LoadPic`
- Private: `executeAssign`, `resolveVariable`, `loadAsset`
- Constructors: `NewEngineState`, `NewSequencer`

**Variables**:
- Public: `Pictures`, `Windows`, `Casts`
- Private: `tickCount`, `waitCounter`, `programCounter`

**Key Design Decisions**:
- Use Go naming conventions (PascalCase for public, camelCase for private)
- Descriptive names over abbreviations
- Consistent naming across codebase

### Dependency Injection Pattern

**Functional Options Pattern**:
```go
type EngineStateOption func(*EngineState)

func WithRenderer(r Renderer) EngineStateOption {
    return func(e *EngineState) {
        e.renderer = r
    }
}

func NewEngineState(opts ...EngineStateOption) *EngineState {
    e := &EngineState{
        // Default initialization
    }
    for _, opt := range opts {
        opt(e)
    }
    return e
}
```

**Key Design Decisions**:
- Optional configuration via functional options
- Composable options
- Clear, readable API
- Easy to extend with new options

### Thread Safety Guidelines

**Mutex Usage**:
- Acquire mutex before reading/writing shared state
- Release mutex as soon as possible
- Never hold mutex during long operations
- Use defer for automatic release

**Atomic Operations**:
- Use atomic operations for simple counters
- Avoid locks for performance-critical paths
- Ensure memory ordering guarantees

**Goroutine Management**:
- Always clean up goroutines on shutdown
- Use context for cancellation
- Avoid goroutine leaks

**Key Design Decisions**:
- Minimize lock contention
- Use atomic operations where possible
- Clear ownership of shared state

---

## Part 10: Performance Considerations

This section describes performance design considerations.

### Execution Performance

**OpCode Execution**:
- Use enum types for fast command dispatch
- Minimize allocations during execution
- Cache frequently accessed data
- Avoid reflection in hot paths

**Variable Lookup**:
- Use map for O(1) lookup
- Cache parent scope references
- Minimize scope chain traversal
- Use lowercase keys for case-insensitive lookup

**Tick Processing**:
- Process all sequences in single pass
- Minimize work per tick
- Use early exit for waiting sequences
- Batch state updates

### Memory Management

**Resource Pooling**:
- Reuse OpCode slices where possible
- Pool temporary buffers
- Avoid allocations in hot paths

**Reference Counting**:
- Track picture references
- Release unused resources
- Avoid memory leaks

**Garbage Collection**:
- Minimize allocations
- Reuse objects where possible
- Profile and optimize hot paths

### Rendering Performance

**Batching**:
- Batch draw calls where possible
- Minimize state changes
- Use texture atlases for sprites

**Culling**:
- Skip rendering of off-screen windows
- Skip rendering of invisible casts
- Use dirty flags to avoid redundant rendering

**Double Buffering**:
- Render to back buffer
- Swap buffers atomically
- Minimize lock duration

---

## Part 11: Future Extensibility

This section describes how the design supports future enhancements.

### Extension Points

**New OpCode Commands**:
- Add new OpCmd enum value
- Implement handler in ExecuteOp
- Update interpreter to generate new OpCode
- No changes to VM structure required

**New Timing Modes**:
- Add new mode constant
- Implement tick generation strategy
- Update RegisterSequence blocking logic
- No changes to sequence structure required

**New Asset Types**:
- Implement new AssetLoader interface
- Add new decoder interface
- Update asset loading logic
- No changes to core engine required

**New Rendering Backends**:
- Implement Renderer interface
- Provide alternative implementation
- Inject via dependency injection
- No changes to state management required

### Backward Compatibility

**Versioning Strategy**:
- Maintain OpCode format compatibility
- Support legacy TFY syntax
- Provide migration tools for breaking changes
- Document compatibility guarantees

**Deprecation Policy**:
- Mark deprecated features clearly
- Provide migration path
- Maintain deprecated features for at least one major version
- Remove only after sufficient notice

---

## Part 12: Embedded Executable Architecture

This section describes the design for creating standalone executables with embedded projects.

### Execution Modes

**son-et supports three execution modes**:

1. **Direct Mode** (Development):
   - Loads TFY scripts from filesystem at runtime
   - Parses and compiles TFY to OpCode on startup
   - Loads assets from filesystem
   - Fast iteration cycle for development

2. **Single-Title Embedded Mode** (Distribution):
   - Single TFY title pre-compiled to OpCode at build time
   - OpCode embedded in executable binary
   - Assets embedded in executable using Go's embed.FS
   - Single-file distribution, no external dependencies
   - Launches directly into the title

3. **Multi-Title Embedded Mode** (Distribution):
   - Multiple TFY titles pre-compiled to OpCode at build time
   - Each title's directory embedded separately to avoid asset conflicts
   - Menu system for title selection
   - Single executable containing multiple titles
   - Ideal for collections or compilations

### Build-Time Compilation

**Compilation Strategy**:
```
Build Time:
  1. Read TFY files from project directory(ies)
  2. Parse TFY to AST
  3. Convert AST to OpCode sequences
  4. Serialize OpCode to Go source code
  5. Embed assets using //go:embed directive (per-directory for multi-title)
  6. Compile to executable with build tags

Runtime (Single-Title Embedded Mode):
  1. Deserialize embedded OpCode
  2. Initialize engine with embedded AssetLoader
  3. Execute OpCode directly (no parsing)

Runtime (Multi-Title Embedded Mode):
  1. Display title selection menu
  2. User selects title
  3. Load selected title's OpCode and AssetLoader
  4. Execute title
  5. Return to menu on completion or ESC
```

**Key Design Decisions**:
- Build-time compilation eliminates parsing overhead
- OpCode serialization format is Go source code (type-safe)
- Assets embedded using standard Go embed.FS
- Each title gets its own embed.FS to avoid asset conflicts
- No custom binary format needed

### Asset Loading Abstraction

**AssetLoader Interface**:
```go
type AssetLoader interface {
    // ReadFile reads a file from the asset source
    ReadFile(path string) ([]byte, error)
    
    // Exists checks if a file exists
    Exists(path string) bool
    
    // ListFiles lists files matching a pattern
    ListFiles(pattern string) ([]string, error)
}
```

**Implementations**:

1. **FilesystemAssetLoader** (Direct Mode):
```go
type FilesystemAssetLoader struct {
    basePath string  // Project directory
}

func (f *FilesystemAssetLoader) ReadFile(path string) ([]byte, error) {
    // Case-insensitive file matching (Windows 3.1 compatibility)
    fullPath := filepath.Join(f.basePath, path)
    return os.ReadFile(fullPath)
}
```

2. **EmbedFSAssetLoader** (Embedded Mode):
```go
type EmbedFSAssetLoader struct {
    fs embed.FS  // Embedded filesystem
}

func (e *EmbedFSAssetLoader) ReadFile(path string) ([]byte, error) {
    // Read from embedded filesystem
    return e.fs.ReadFile(path)
}
```

**Key Design Decisions**:
- Single interface for both modes
- Case-insensitive file matching in both modes
- No mode-specific code in engine
- AssetLoader injected via dependency injection

### Build Configuration

**Build Tag Strategy**:
```go
// build_kuma2.go
//go:build embed_kuma2

package main

import _ "embed"

//go:embed samples/kuma2/*
var embeddedFS embed.FS

var embeddedProject = &EmbeddedProject{
    Name:    "kuma2",
    Assets:  embeddedFS,
    OpCodes: kuma2OpCodes,  // Generated at build time
}
```

**Build Commands**:
```bash
# Direct mode (development)
go run ./cmd/son-et samples/kuma2

# Embedded mode (distribution)
go build -tags embed_kuma2 -o kuma2 ./cmd/son-et
./kuma2  # Runs embedded project
```

**Key Design Decisions**:
- Use Go build tags for mode selection
- One build configuration file per project
- Embedded projects registered at package init time
- No runtime mode detection needed

### OpCode Serialization

**Serialization Format**:
OpCode sequences are serialized as Go source code for type safety and simplicity.

```go
// Generated code example
var kuma2OpCodes = map[string][]OpCode{
    "main": {
        {Cmd: OpCall, Args: []any{"LoadPic", 0, "TITLE.BMP"}},
        {Cmd: OpCall, Args: []any{"OpenWin", 0, 0, 0, 640, 480}},
        {Cmd: OpCall, Args: []any{"Wait", 100}},
        // ... more opcodes
    },
    "animate": {
        // ... function opcodes
    },
}
```

**Generator Design**:
```go
type OpCodeGenerator struct {
    writer io.Writer
}

func (g *OpCodeGenerator) Generate(script *Script) error {
    // 1. Write package declaration
    // 2. Write imports
    // 3. For each function, serialize OpCode sequence
    // 4. Write variable declarations
    // 5. Write initialization code
}
```

**Key Design Decisions**:
- Use Go source code, not binary format
- Type-safe at compile time
- Human-readable for debugging
- No custom serialization format needed

### Multi-Title Architecture

**Directory-Based Embed Strategy**:

Each FILLY title resides in its own directory with all assets:
```
samples/
├── kuma2/              # Title 1
│   ├── KUMA2.TFY      # Entry point
│   ├── KUMA-1.BMP     # Assets
│   ├── KUMA.MID
│   └── ...
├── robot/              # Title 2
│   ├── ROBOT.TFY
│   ├── ROBOT000.BMP
│   └── ...
└── y-saru/             # Title 3
    ├── Y-SARU.TFY
    └── ...
```

**Separate Embed Per Title**:
To avoid asset filename conflicts between titles, each title's directory is embedded separately:

```go
// Generated multi-title code
//go:embed samples/kuma2
var kuma2FS embed.FS

//go:embed samples/robot
var robotFS embed.FS

//go:embed samples/y-saru
var ysaruFS embed.FS

type TitleInfo struct {
    Name        string
    Title       string                      // Display name from #info
    Description string                      // From #info
    Directory   string                      // Source directory
    GetOpCodes  func() []interpreter.OpCode // Returns compiled OpCodes
    GetFS       func() embed.FS             // Returns title's filesystem
}

func GetTitles() []TitleInfo {
    return []TitleInfo{
        {
            Name:        "kuma2",
            Title:       "Kuma Game",
            Description: "A bear adventure",
            Directory:   "samples/kuma2",
            GetOpCodes:  GetKuma2OpCodes,
            GetFS:       GetKuma2FS,
        },
        {
            Name:        "robot",
            Title:       "Robot Story",
            Description: "An animated story",
            Directory:   "samples/robot",
            GetOpCodes:  GetRobotOpCodes,
            GetFS:       GetRobotFS,
        },
        // ... more titles
    }
}
```

**Menu System**:
```go
func DisplayMenu(titles []TitleInfo) int {
    fmt.Println("=================================")
    fmt.Println("  FILLY Title Launcher")
    fmt.Println("=================================")
    fmt.Println()
    
    for i, title := range titles {
        fmt.Printf("%d. %s\n", i+1, title.Title)
        if title.Description != "" {
            fmt.Printf("   %s\n", title.Description)
        }
    }
    
    fmt.Println()
    fmt.Println("0. Exit")
    fmt.Println()
    fmt.Print("Select title: ")
    
    var choice int
    fmt.Scanf("%d", &choice)
    return choice
}

func main() {
    titles := GetTitles()
    
    for {
        choice := DisplayMenu(titles)
        
        if choice == 0 {
            os.Exit(0)
        }
        
        if choice < 1 || choice > len(titles) {
            fmt.Println("Invalid choice. Please try again.")
            continue
        }
        
        selectedTitle := titles[choice-1]
        fmt.Printf("\nLaunching: %s\n\n", selectedTitle.Title)
        
        // Get title's OpCodes and filesystem
        opcodes := selectedTitle.GetOpCodes()
        titleFS := selectedTitle.GetFS()
        
        // Create AssetLoader for this title
        assetLoader := NewEmbedFSAssetLoader(titleFS, selectedTitle.Directory)
        
        // Execute title
        engine := NewEngine(assetLoader)
        engine.Execute(opcodes)
        
        fmt.Println("\nTitle completed. Press Enter to return to menu...")
        fmt.Scanln()
    }
}
```

**Key Design Decisions**:
- Each title gets its own `embed.FS` variable
- Prevents asset filename conflicts (e.g., multiple titles with "TITLE.BMP")
- Each title's AssetLoader scoped to its directory
- Menu-driven selection for user experience
- ESC in title returns to menu (multi-title mode)
- ESC in menu exits program
- Clean state reset between title executions

**Single-Title vs Multi-Title Generation**:

The serializer provides two functions:
- `SerializeSingleTitle(title)` - Generates code for one title, launches directly
- `SerializeMultiTitle(titles)` - Generates code for multiple titles with menu

Build tool decides which to use based on configuration.

### Mode Detection and Initialization

**Application Entry Point**:
```go
func main() {
    // Check if embedded project is registered
    if embeddedProject != nil {
        // Embedded mode
        runEmbedded(embeddedProject)
    } else if len(os.Args) > 1 {
        // Direct mode with project path
        runDirect(os.Args[1])
    } else {
        // No project specified
        printUsage()
    }
}
```

**Initialization Flow**:
```
Direct Mode:
  1. Parse command-line arguments
  2. Load TFY files from directory
  3. Parse TFY to AST
  4. Convert AST to OpCode
  5. Create FilesystemAssetLoader
  6. Initialize engine
  7. Execute OpCode

Embedded Mode:
  1. Load embedded OpCode
  2. Create EmbedFSAssetLoader
  3. Initialize engine
  4. Execute OpCode
```

**Key Design Decisions**:
- Mode determined by presence of embedded project
- Same engine initialization for both modes
- Only AssetLoader differs between modes
- No conditional logic in engine code

### Headless Mode Integration

**Headless mode works in both execution modes**:

```bash
# Direct mode + headless
go run ./cmd/son-et --headless --timeout=10s samples/kuma2

# Embedded mode + headless
./kuma2 --headless --timeout=10s
```

**Implementation**:
- Headless flag parsed before mode detection
- MockRenderer injected regardless of mode
- Logging works identically in both modes
- Timeout applies to both modes

**Key Design Decisions**:
- Headless is orthogonal to execution mode
- Same command-line interface for both modes
- No mode-specific headless logic

### Distribution Workflow

**Creating Standalone Executables**:

1. **Create build configuration**:
```go
// build_myproject.go
//go:build embed_myproject

package main

import _ "embed"

//go:embed samples/myproject/*
var embeddedFS embed.FS

var embeddedProject = &EmbeddedProject{
    Name:   "myproject",
    Assets: embeddedFS,
}

func init() {
    // Generate OpCode at build time
    embeddedProject.OpCodes = generateOpCodes("samples/myproject")
}
```

2. **Build executable**:
```bash
go build -tags embed_myproject -o myproject ./cmd/son-et
```

3. **Distribute**:
```bash
# Single executable, no dependencies
./myproject
```

**Key Design Decisions**:
- One-step build process
- No external tools required
- Standard Go toolchain
- Cross-platform support via Go's build system

### Error Handling in Embedded Mode

**Build-Time Errors**:
- TFY parsing errors fail the build
- Missing assets fail the build
- OpCode generation errors fail the build
- Clear error messages with file/line numbers

**Runtime Errors**:
- Same error handling as direct mode
- Embedded mode has no parsing errors (pre-compiled)
- Asset loading errors still possible (corrupted embed)
- Error messages reference embedded paths

**Key Design Decisions**:
- Fail fast at build time
- Catch errors before distribution
- Runtime errors are rare in embedded mode
- Same error reporting format for both modes

### Performance Characteristics

**Direct Mode**:
- Startup time: ~100ms (parsing + compilation)
- Memory: Higher (AST + OpCode in memory)
- Iteration: Fast (no rebuild needed)

**Embedded Mode**:
- Startup time: ~10ms (no parsing)
- Memory: Lower (only OpCode in memory)
- Distribution: Single file, smaller size

**Key Design Decisions**:
- Embedded mode optimized for startup time
- Direct mode optimized for iteration speed
- Both modes have identical runtime performance
- No performance penalty for abstraction

### Window Drag Interaction

**User Interaction Design**:
Windows with captions can be dragged by the user to reposition them within the virtual desktop.

**Drag Detection**:
```
1. Mouse down event occurs
2. Check if mouse position is within any window's title bar region
3. If yes, enter drag mode for that window
4. Store initial mouse position and window position
```

**Drag Update**:
```
1. Mouse move event occurs while in drag mode
2. Calculate delta: (currentMouseX - initialMouseX, currentMouseY - initialMouseY)
3. Update window position: (initialWindowX + deltaX, initialWindowY + deltaY)
4. Constrain window to virtual desktop bounds
5. Trigger re-render
```

**Drag End**:
```
1. Mouse up event occurs
2. Exit drag mode
3. Finalize window position
```

**Title Bar Region**:
- Height: Typically 20-30 pixels (configurable)
- Width: Full window width
- Position: Top of window

**Boundary Constraints**:
- Window must remain at least partially visible
- Minimum visible area: Title bar must be within desktop
- Prevents windows from being dragged completely off-screen

**Key Design Decisions**:
- Drag only by title bar (not entire window)
- Real-time position updates during drag
- Smooth visual feedback
- Boundary constraints prevent lost windows
- No script API needed (automatic behavior)

### Multi-Project Menu System

**Multi-Project Mode**:
When multiple projects are embedded, the executable provides a menu-driven interface for project selection.

**Application State Machine**:
```
┌─────────────┐
│   Startup   │
└──────┬──────┘
       │
       ▼
┌─────────────────┐
│  Project Menu   │◄─────────┐
│  (List projects)│          │
└────┬────────┬───┘          │
     │        │              │
     │ Select │         ESC pressed
     │        │         or project
     ▼        │         completes
┌─────────────┐│              │
│   Execute   ││              │
│   Project   ├┘              │
└─────────────┴───────────────┘
     │
     │ ESC in menu
     ▼
┌─────────────┐
│  Terminate  │
└─────────────┘
```

**Menu UI Design**:
```
┌────────────────────────────────────┐
│     son-et Project Selector        │
│                                    │
│  1. Kuma2 Adventure                │
│  2. Y-Saru Game                    │
│  3. Robot Demo                     │
│                                    │
│  Select project (1-3) or ESC to exit│
└────────────────────────────────────┘
```

**Menu Implementation**:
- Simple text-based menu rendered on virtual desktop
- Keyboard input for selection (1-9 for projects, ESC to exit)
- Mouse click support for selection
- Clear visual feedback for selection

**ESC Key Behavior Context**:

**Single Project Mode**:
- ESC → Terminate program immediately

**Multi-Project Mode - In Menu**:
- ESC → Terminate program

**Multi-Project Mode - In Project**:
- ESC → Return to menu (not terminate)

**Project Completion**:
- When main() completes → Return to menu (multi-project mode)
- When main() completes → Terminate (single project mode)

**State Management**:
```go
type ApplicationMode int

const (
    SingleProject ApplicationMode = iota
    MultiProjectMenu
    MultiProjectRunning
)

type ApplicationState struct {
    mode            ApplicationMode
    projects        []EmbeddedProject
    currentProject  *EmbeddedProject
    menuSelection   int
}

func (s *ApplicationState) HandleESC() {
    switch s.mode {
    case SingleProject:
        // Terminate immediately
        terminate()
    case MultiProjectMenu:
        // Terminate from menu
        terminate()
    case MultiProjectRunning:
        // Return to menu
        s.mode = MultiProjectMenu
        cleanupProject()
    }
}
```

**Project Lifecycle in Multi-Project Mode**:
```
1. Display menu
2. User selects project
3. Initialize engine with project's AssetLoader
4. Load project's OpCode
5. Execute project
6. On completion or ESC:
   a. Cleanup engine state
   b. Reset graphics/audio
   c. Return to menu
7. Repeat from step 1
```

**Build Configuration for Multi-Project**:
```go
// build_collection.go
//go:build embed_collection

package main

import _ "embed"

//go:embed samples/kuma2/*
var kuma2FS embed.FS

//go:embed samples/y-saru/*
var ysaruFS embed.FS

//go:embed samples/robot/*
var robotFS embed.FS

var embeddedProjects = []EmbeddedProject{
    {Name: "Kuma2 Adventure", Assets: kuma2FS, OpCodes: kuma2OpCodes},
    {Name: "Y-Saru Game", Assets: ysaruFS, OpCodes: ysaruOpCodes},
    {Name: "Robot Demo", Assets: robotFS, OpCodes: robotOpCodes},
}
```

**Key Design Decisions**:
- Menu is part of the application layer, not engine
- ESC behavior is context-aware (menu vs project)
- Clean state reset between projects
- No cross-project state contamination
- Simple, intuitive navigation
- Supports up to 9 projects (keyboard 1-9)

---

## Part 13: Function Return Value Mechanism

This section describes the design for capturing and using function return values.

### Problem Statement

FILLY functions can return values that need to be captured by the caller:
```filly
pic = LoadPic("image.bmp");    // Capture picture ID
win = OpenWin(pic);            // Capture window ID
cast = PutCast(win, pic, ...); // Capture cast ID
```

Currently, return values are ignored with `_ = picID` in the VM implementation.

### Design Approach: Special Return Variable

**Mechanism**: Use a special variable `__return__` to pass return values from functions to assignment statements.

**Execution Flow**:
```
1. Assignment with function call: x = LoadPic("file.bmp")
2. Parser generates: OpAssign(Variable("x"), OpCall("LoadPic", "file.bmp"))
3. VM evaluates right-hand side (OpCall):
   a. Execute LoadPic function
   b. Function stores result in __return__ variable
   c. VM reads __return__ and returns it as expression value
4. VM assigns returned value to variable x
```

**Implementation Strategy**:

**Step 1: Modify VM Function Call Handling**
```go
// In executeBuiltinFunction:
case "loadpic":
    filename := fmt.Sprintf("%v", evaluatedArgs[0])
    picID := vm.engine.LoadPic(filename)
    // Store in special return variable
    seq.SetVariable("__return__", picID)
    return nil

case "openwin":
    // ... execute OpenWin ...
    winID := vm.engine.OpenWin(...)
    seq.SetVariable("__return__", winID)
    return nil
```

**Step 2: Modify Expression Evaluation**
```go
// In evaluateExpression:
case interpreter.OpCall:
    // Execute function call
    err := vm.executeCall(seq, op)
    if err != nil {
        return nil, err
    }
    // Read return value from special variable
    returnValue := seq.GetVariable("__return__")
    // Clear return variable for next call
    seq.SetVariable("__return__", 0)
    return returnValue, nil
```

**Step 3: Update All Functions That Return Values**

Functions that return values:
- `LoadPic(filename)` → picture ID
- `CreatePic(width, height)` → picture ID
- `OpenWin(...)` → window ID
- `PutCast(...)` → cast ID
- `GetPicNo(winID)` → picture ID
- `PicWidth(picID)` → width
- `PicHeight(picID)` → height
- `GetColor(pic, x, y)` → color value
- `LoadRsc(filename)` → resource ID
- String functions: `StrLen`, `SubStr`, `StrFind`, `StrPrint`, etc.
- Array functions: `ArraySize`
- Math functions: `Random`, `MakeLong`, `GetHiWord`, `GetLowWord`
- File functions: `OpenF`, `ReadF`, `IsExist`, `GetCwd`
- System functions: `WinInfo`, `GetSysTime`, `WhatDay`, `WhatTime`, `GetCmdLine`

### Key Design Decisions

- **Special variable name**: `__return__` is unlikely to conflict with user variables
- **Automatic cleanup**: Clear `__return__` after reading to prevent stale values
- **Scope-local**: Return value stored in current sequence's scope
- **No stack needed**: Single return value sufficient (no nested calls in expressions)
- **Backward compatible**: Functions without return values work unchanged

### Error Handling

**Missing return value**:
- If function doesn't set `__return__`, default value (0) is used
- No error thrown (graceful degradation)

**Nested function calls**:
- Not supported in current parser (expressions are flat)
- If added later, would need return value stack

### Testing Strategy

**Unit tests**:
- Test each function's return value storage
- Test assignment with function call
- Test return value cleanup
- Test functions without return values

**Integration tests**:
- Test sample scripts using return values
- Verify correct IDs are captured and used

---

## Part 14: Step Block Syntax

This section describes the design for `step(n) { ... }` block syntax.

### Problem Statement

FILLY supports a block-style step syntax used in some sample scripts:
```filly
mes(MIDI_TIME) {
    step(65) {
        command1;,
        command2;,,
        command3;,
        end_step;
    }
}
```

This differs from the simple `step(n);` statement and requires special parsing and execution.

### Syntax Semantics

**Block structure**:
```filly
step(n) {
    command1;,    // Execute command1, then wait 1 step
    command2;,,   // Execute command2, then wait 2 steps
    command3;,    // Execute command3, then wait 1 step
    end_step;     // Exit step block (optional)
}
```

**Comma semantics**:
- Each comma after a semicolon represents one wait step
- `;,` → Execute command, then Wait(1)
- `;,,` → Execute command, then Wait(2)
- `;,,,` → Execute command, then Wait(3)
- The actual wait duration is: comma_count × step_duration (where step_duration = n from step(n))

**Execution model**:
1. SetStep(n) - Set the step duration
2. Execute `command1`
3. Wait(1) - Wait 1 step (from the single comma)
4. Execute `command2`
5. Wait(2) - Wait 2 steps (from the two commas)
6. Execute `command3`
7. Wait(1) - Wait 1 step
8. Encounter `end_step` → exit block

**Key difference from simple step**:
- Simple: `step(10);` → Wait(10) - wait 10 steps once
- Block: `step(10) { cmd1;, cmd2;,, }` → SetStep(10), cmd1, Wait(1), cmd2, Wait(2)

### Design Approach: Flat Sequence Generation

**Strategy**: Transform step block into a flat sequence with SetStep and Wait operations.

**AST Representation**:
```go
type StepStatement struct {
    Token token.Token      // 'step' token
    Count Expression       // Step count (e.g., 65)
    Body  *BlockStatement  // Optional: step(n) { ... } block
}
```

**OpCode Generation**:
```
step(n) {
    cmd1;,
    cmd2;,,
    end_step;
}

Transforms to:

[
    OpSetStep(n),      // Set step duration
    OpCode for cmd1,   // Execute command
    OpWait(1),         // Wait 1 step (from single comma)
    OpCode for cmd2,   // Execute command
    OpWait(2),         // Wait 2 steps (from two commas)
    // end_step terminates generation
]
```

**Parser Changes**:

**Step 1: StepStatement in AST**
```go
// In pkg/compiler/ast/ast.go
type StepStatement struct {
    Token token.Token
    Count Expression
    Body  *BlockStatement  // nil for simple step(n);
}

func (s *StepStatement) statementNode() {}
func (s *StepStatement) TokenLiteral() string { return s.Token.Literal }
```

**Step 2: Parser Recognizes Block Syntax**
```go
// In pkg/compiler/parser/parser.go
func (p *Parser) parseStepStatement() *ast.StepStatement {
    stmt := &ast.StepStatement{Token: p.curToken}
    
    // Expect '('
    if !p.expectPeek(token.LPAREN) {
        return nil
    }
    
    // Parse count expression
    p.nextToken()
    stmt.Count = p.parseExpression(LOWEST)
    
    // Expect ')'
    if !p.expectPeek(token.RPAREN) {
        return nil
    }
    
    // Check for block syntax
    if p.peekTokenIs(token.LBRACE) {
        p.nextToken()
        stmt.Body = p.parseStepBlock()
    }
    
    return stmt
}

// parseStepBlock parses a step block where commas represent wait steps
func (p *Parser) parseStepBlock() *ast.BlockStatement {
    block := &ast.BlockStatement{Token: p.curToken}
    block.Statements = []ast.Statement{}
    
    p.nextToken()
    for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
        stmt := p.parseStatement()
        if stmt != nil {
            block.Statements = append(block.Statements, stmt)
        }
        
        // In step blocks, semicolons and commas are both statement terminators
        // Commas represent empty steps (Wait operations)
        for p.peekTokenIs(token.SEMICOLON) || p.peekTokenIs(token.COMMA) {
            if p.peekTokenIs(token.COMMA) {
                // Add empty statement for each comma (represents Wait(1))
                block.Statements = append(block.Statements, &ast.ExpressionStatement{
                    Token:      p.curToken,
                    Expression: nil,
                })
            }
            p.nextToken()
        }
        p.nextToken()
    }
    
    return block
}
```

**Step 3: CodeGen Generates Flat Sequence**
```go
// In pkg/compiler/codegen/codegen.go
func (g *Generator) generateStepStatement(stmt *ast.StepStatement) []interpreter.OpCode {
    count := g.generateExpression(stmt.Count)
    
    // Simple form: step(n);
    if stmt.Body == nil {
        return []interpreter.OpCode{{
            Cmd:  interpreter.OpWait,
            Args: []any{count},
        }}
    }
    
    // Block form: step(n) { ... }
    var opcodes []interpreter.OpCode
    
    // First, generate SetStep to set the step duration
    opcodes = append(opcodes, interpreter.OpCode{
        Cmd:  interpreter.OpSetStep,
        Args: []any{count},
    })
    
    i := 0
    for i < len(stmt.Body.Statements) {
        cmd := stmt.Body.Statements[i]
        
        // Check for empty statement (from commas) - count consecutive commas
        if exprStmt, ok := cmd.(*ast.ExpressionStatement); ok && exprStmt.Expression == nil {
            waitCount := 0
            for i < len(stmt.Body.Statements) {
                if exprStmt, ok := stmt.Body.Statements[i].(*ast.ExpressionStatement); ok && exprStmt.Expression == nil {
                    waitCount++
                    i++
                } else {
                    break
                }
            }
            
            // Generate Wait with the count
            opcodes = append(opcodes, interpreter.OpCode{
                Cmd:  interpreter.OpWait,
                Args: []any{int64(waitCount)},
            })
            continue
        }
        
        // Check for end_step - terminates code generation
        if callExpr, ok := cmd.(*ast.ExpressionStatement); ok {
            if call, ok := callExpr.Expression.(*ast.CallExpression); ok {
                if ident, ok := call.Function.(*ast.Identifier); ok && ident.Value == "end_step" {
                    break
                }
            }
        }
        
        // Generate command OpCode
        cmdOps := g.generateStatement(cmd)
        opcodes = append(opcodes, cmdOps...)
        
        i++
    }
    
    return opcodes
}
```

### Comma Syntax Handling

**Trailing comma**: Commands ending with `,` indicate "wait after execution"
```filly
command1;,   // Execute command1, then wait 1 step
command2;,,  // Execute command2, then wait 2 steps
```

**Parser strategy**:
- Each comma creates an empty statement in the AST
- CodeGen counts consecutive empty statements
- Generates Wait(count) for consecutive commas

### Key Design Decisions

- **Flat sequence generation**: No loop transformation, generates linear OpCode sequence
- **SetStep operation**: Sets step duration at the start of the block
- **Comma counting**: Consecutive commas become Wait(count)
- **end_step terminates generation**: No OpCode generated, just stops processing
- **No automatic wait insertion**: Commas are required for waits

### Error Handling

**Missing commas**:
- Commands without trailing commas execute without waiting
- This is valid syntax (immediate execution)

**Invalid step count**:
- Runtime error if count evaluates to non-integer
- Graceful degradation: treat as 0 (no wait)

### Testing Strategy

**Unit tests**:
- Test parser recognizes step block syntax
- Test comma counting in parser
- Test OpCode generation for step blocks
- Test execution with multiple commands and varying comma counts
- Test end_step termination

**Integration tests**:
- Test KUMA2 sample (uses step blocks)
- Verify timing accuracy
- Verify command execution order

---

## Conclusion

This design document describes the ideal architecture for son-et based on the requirements and architectural principles. The design emphasizes:

1. **Uniform OpCode-based execution** for consistency and simplicity
2. **Event-driven step-based execution** for precise timing control
3. **Dual timing mode architecture** for music and time synchronization
4. **Hierarchical variable scope** for lexical scoping
5. **Thread-safe state management** for concurrent access
6. **Non-blocking audio architecture** for responsive playback
7. **Dual execution modes** for development and distribution

The design is organized into clear layers with well-defined boundaries, uses dependency injection for testability, and provides extension points for future enhancements. All design decisions align with the core architectural principles and support the unique execution model of FILLY scripts.

**Next Steps**: Implement the design incrementally, starting with the foundation layer and working up through the compilation and execution layers. Use property-based testing to verify correctness properties at each stage.
