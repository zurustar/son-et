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
type VariableValue interface{}  // Can be int, string, or []int

// Variable storage in scope
vars := map[string]VariableValue{
    "x":      42,           // int
    "name":   "FILLY",      // string
    "scores": []int{85, 92, 78},  // array
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
- Integer-only (no string arrays)
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
   b. Check for circular includes
   c. Recursively preprocess included file (with encoding conversion)
   d. Insert processed content at #include location
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
- Circular include detection uses file path tracking
- Case-insensitive file matching for Windows 3.1 compatibility
- Encoding conversion is transparent to the rest of the compiler

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
- Casts support transparency and alpha blending
- Casts are positioned relative to their containing window

**Window Management**:
- Windows display pictures with optional decorations
- Windows maintain creation order for z-ordering
- Windows support captions and resizing
- Windows clip casts to their boundaries

**Key Design Decisions**:
- Sequential ID assignment (simple, predictable)
- Creation order determines z-order (no explicit z-index)
- Immutable resources (transformations create new resources)
- Reference counting for memory management

### Rendering Pipeline

**Rendering Stages**:
```
1. Clear virtual desktop
2. For each window (in creation order):
   a. Draw window background (picture)
   b. Draw window decorations (caption, border)
   c. For each cast in window (in creation order):
      - Apply clipping region
      - Apply transparency
      - Draw cast to window
3. Scale virtual desktop to screen
4. Present to display
```

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

**MIDI Player Design**:
- Global singleton (one MIDI file plays at a time)
- Runs in separate goroutine (audio thread)
- Uses MeltySynth for MIDI synthesis with SoundFont (.sf2)
- Implements io.Reader interface (MidiStream) for Ebiten audio integration
- Generates ticks based on wall-clock time and tempo map
- Invokes tick callbacks to drive MIDI_TIME sequences

**MIDI Tick Calculation**:
```
Tick Interval = (60 / tempo) / PPQ / 8
  where:
    tempo = beats per minute (from MIDI tempo events)
    PPQ = pulses per quarter note (from MIDI file header)
    8 = 32nd notes per quarter note (FILLY's tick resolution)
```

**Wall-Clock Time Based Timing**:
```
currentTick = CalculateTickFromTime(elapsed_seconds)
  where:
    elapsed_seconds = time.Since(startTime).Seconds()
    
Algorithm:
  1. Get elapsed time since playback started
  2. Find current tempo from tempo map
  3. Calculate ticks: elapsed * (tempo / 60) * PPQ
  4. Adjust for tempo changes throughout the file
```

**Rationale for Wall-Clock Time**:
- **Accuracy**: No cumulative drift from audio buffer processing delays
- **Determinism**: Same elapsed time always produces same tick
- **Tempo-awareness**: Properly handles tempo changes via tempo map
- **Buffer independence**: Works correctly regardless of buffer size

**MIDI Playback Lifecycle**:
```
1. PlayMIDI(filename) called
2. Load MIDI file and parse tempo/PPQ using MeltySynth
3. Load SoundFont (.sf2) file for synthesis
4. Create MidiStream (implements io.Reader):
   a. Wraps MeltySynth sequencer
   b. Stores tempo map and PPQ
   c. Records start time (wall-clock)
   d. Calculates total ticks for end detection
5. Create Ebiten audio player with MidiStream
6. Start playback in goroutine (non-blocking)
7. MidiStream.Read() called by audio thread:
   a. Render audio samples via MeltySynth
   b. Calculate current tick from elapsed time
   c. Deliver all ticks from lastTick+1 to currentTick sequentially
   d. Detect MIDI end (currentTick >= totalTicks)
   e. Trigger MIDI_END event when complete
8. On completion:
   a. Set midiFinished flag
   b. Trigger MIDI_END event
   c. Stop tick generation
```

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
- MIDI player is independent from sequences
- Tick delivery uses wall-clock time (not sample counting) for accuracy
- Sequential tick delivery prevents frame skipping
- MidiStream implements io.Reader for Ebiten audio integration
- MeltySynth provides accurate MIDI synthesis with SoundFont support
- MIDI continues playing even if starting sequence terminates
- Only one MIDI file plays at a time (matches original behavior)
- MIDI end detection based on tick count comparison

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

## Part 7: Testing Strategy

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
