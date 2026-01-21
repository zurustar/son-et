# Implementation Tasks: son-et Core Engine

## Overview

This task list implements the requirements defined in [requirements.md](requirements.md) following the architecture described in [design.md](design.md).

**Strategy**: Incremental implementation from foundation to features, following dependency order.

**Principle**: Build from generic to specific, from lower layers to higher layers.

**Reference**: Existing code in `_old_implementation/` can be referenced but should not constrain the new design.

---

## Phase 0: Foundation Layer

### Task 0.1: Core Data Structures
**Goal**: Define clean, minimal data structures for OpCode and execution state.

**Subtasks**:
- [x] 0.1.1 Define OpCmd enum type with all command types
- [x] 0.1.2 Define OpCode struct (Cmd, Args)
- [x] 0.1.3 Define Variable type wrapper
- [x] 0.1.4 Add unit tests for OpCode construction

**Acceptance Criteria**:
- OpCmd uses enum (not strings) for type safety
- OpCode supports nested OpCode in Args
- Variable type distinguishes variables from literals

---

### Task 0.2: Sequencer Structure
**Goal**: Define the execution context for sequences with clean scope management.

**Subtasks**:
- [x] 0.2.1 Define Sequencer struct (commands, pc, active, mode, waitCount, stepSize, vars, parent)
- [x] 0.2.2 Implement NewSequencer constructor
- [x] 0.2.3 Implement variable resolution with scope chain walking
- [x] 0.2.4 Implement array storage and auto-expansion
- [x] 0.2.5 Add unit tests for variable scoping
- [x] 0.2.6 Add unit tests for array operations

**Acceptance Criteria**:
- Sequencer maintains parent pointer for scope chain
- Variable lookup walks up scope chain correctly
- Case-insensitive variable resolution
- Default values for undefined variables (0, "", [])
- Arrays stored as Go slices
- Array auto-expansion works correctly
- Zero-fill for new array elements

---

### Task 0.3: Interface Definitions
**Goal**: Define clean interfaces for dependency injection.

**Subtasks**:
- [x] 0.3.1 Define Renderer interface
- [x] 0.3.2 Define AssetLoader interface (ReadFile, Exists, ListFiles)
- [x] 0.3.3 Define ImageDecoder interface
- [x] 0.3.4 Define TickGenerator interface

**Acceptance Criteria**:
- Interfaces are minimal and focused
- Interfaces enable testing without platform dependencies
- Interfaces support both production and mock implementations
- AssetLoader supports both filesystem and embedded modes

---

### Task 0.4: AssetLoader Implementations
**Goal**: Implement asset loading for both direct and embedded modes.

**Subtasks**:
- [x] 0.4.1 Implement FilesystemAssetLoader (direct mode)
- [x] 0.4.2 Implement case-insensitive file matching (Windows 3.1 compatibility)
- [x] 0.4.3 Implement EmbedFSAssetLoader (embedded mode)
- [x] 0.4.4 Add unit tests for both implementations

**Acceptance Criteria**:
- FilesystemAssetLoader reads from filesystem with case-insensitive matching
- EmbedFSAssetLoader reads from embed.FS
- Both implementations satisfy AssetLoader interface
- Tests verify both modes work correctly

---

## Phase 1: System Infrastructure

### Task 1.1: Logging System
**Goal**: Implement timestamped logging with debug levels.

**Subtasks**:
- [x] 1.1.1 Implement timestamp formatting [HH:MM:SS.mmm]
- [x] 1.1.2 Implement DEBUG_LEVEL support (0, 1, 2)
- [x] 1.1.3 Implement log functions (LogError, LogInfo, LogDebug)
- [x] 1.1.4 Add tests for logging

**Acceptance Criteria**:
- Timestamps in all logs
- Debug levels control verbosity (0=errors, 1=info, 2=debug)
- Thread-safe logging

---

### Task 1.2: Headless Mode
**Goal**: Implement headless execution for testing.

**Subtasks**:
- [x] 1.2.1 Implement --headless flag parsing
- [x] 1.2.2 Implement MockRenderer for headless mode
- [x] 1.2.3 Implement headless execution loop (60 FPS)
- [x] 1.2.4 Implement rendering operation logging
- [x] 1.2.5 Add tests for headless mode

**Acceptance Criteria**:
- Scripts execute without GUI
- All logic executes normally
- Rendering operations logged
- 60 FPS tick generation

---

### Task 1.3: Program Termination
**Goal**: Implement clean program termination.

**Subtasks**:
- [x] 1.3.1 Implement --timeout flag parsing and timer
- [x] 1.3.2 Implement programTerminated flag (atomic)
- [x] 1.3.3 Implement ESC key detection and termination
- [x] 1.3.4 Implement graceful shutdown (cleanup resources)
- [x] 1.3.5 Add tests for termination

**Acceptance Criteria**:
- Timeout formats supported (5s, 500ms, 2m)
- ESC key sets termination flag immediately
- Termination check happens before VM execution
- Graceful resource cleanup
- Exit code 0 on normal termination

---

### Task 1.4: Error Reporting
**Goal**: Implement clear error messages.

**Subtasks**:
- [x] 1.4.1 Implement parsing error reporting (line, column)
- [x] 1.4.2 Implement runtime error reporting (OpCode, args)
- [x] 1.4.3 Implement asset loading error reporting (filename)
- [x] 1.4.4 Add tests for error reporting

**Acceptance Criteria**:
- Parsing errors include line/column
- Runtime errors include OpCode context
- Asset errors include filename
- Errors logged with appropriate level

---

## Phase 2: Compilation Layer

### Task 2.1: Preprocessor
**Goal**: Implement #info and #include directive processing with character encoding support.

**Subtasks**:
- [x] 2.1.1 Implement character encoding detection (UTF-8 vs Shift-JIS)
- [x] 2.1.2 Implement Shift-JIS to UTF-8 conversion using golang.org/x/text
- [x] 2.1.3 Implement #info directive parsing and metadata storage
- [x] 2.1.4 Implement #include directive parsing
- [x] 2.1.5 Implement file path resolution (case-insensitive)
- [x] 2.1.6 Implement circular include detection
- [x] 2.1.7 Implement recursive include processing (with encoding conversion)
- [x] 2.1.8 Implement preprocessed source merging
- [x] 2.1.9 Add tests for preprocessor with Shift-JIS files
- [x] 2.1.10 Add tests for encoding conversion edge cases

**Acceptance Criteria**:
- Shift-JIS files automatically converted to UTF-8
- UTF-8 files processed without conversion
- Encoding errors reported clearly
- #info directives parsed and metadata stored
- #include directives load and merge files correctly
- Case-insensitive file matching works
- Circular includes detected and reported
- Recursive includes work correctly with proper encoding
- All standard #info tags supported

---

### Task 2.2: Lexer
**Goal**: Clean, minimal lexer that tokenizes FILLY source code.

**Subtasks**:
- [x] 2.2.1 Implement token types (keywords, operators, literals, identifiers)
- [x] 2.2.2 Implement lexer with position tracking
- [x] 2.2.3 Implement error reporting with line/column
- [x] 2.2.4 Add comprehensive lexer tests

**Acceptance Criteria**:
- Lexer handles all FILLY syntax
- Clear error messages with line/column numbers
- Case-insensitive keywords

---

### Task 2.3: Parser
**Goal**: Clean parser that builds AST from tokens.

**Subtasks**:
- [x] 2.3.1 Define AST node types
- [x] 2.3.2 Implement expression parsing (precedence climbing)
- [x] 2.3.3 Implement statement parsing
- [x] 2.3.4 Implement control flow parsing (if, for, while, switch)
- [x] 2.3.5 Implement function definition parsing
- [x] 2.3.6 Implement mes() block parsing
- [x] 2.3.7 Implement array syntax parsing (arr[index])
- [x] 2.3.8 Fix for-loop trailing semicolon handling (legacy compatibility)
- [x] 2.3.9 Add comprehensive parser tests

**Acceptance Criteria**:
- Parser handles all FILLY syntax including arrays
- Clear error messages with line/column numbers
- AST structure is clean and minimal
- for-loops accept optional trailing semicolon: `for(k=0; k<3; k=k+1;)` (legacy FILLY syntax)

**Implementation Notes**:
- Modified `parseForStatement()` to accept optional semicolon before closing paren
- This fixes parsing errors in legacy scripts like samples/yosemiya

---

### Task 2.4: OpCode Generation
**Goal**: Convert AST to OpCode sequences uniformly.

**Subtasks**:
- [x] 2.4.1 Implement statement conversion (assignments, function calls)
- [x] 2.4.2 Implement expression conversion (arithmetic, comparisons, nested)
- [x] 2.4.3 Implement control flow conversion (if, for, while, switch)
- [x] 2.4.4 Implement function definition conversion
- [x] 2.4.5 Implement mes() block conversion
- [x] 2.4.6 Implement array access conversion (arr[index])
- [x] 2.4.7 Add comprehensive OpCode generation tests

**Acceptance Criteria**:
- All FILLY constructs convert to OpCode
- Nested expressions become nested OpCode
- Control flow uses OpCode with nested blocks
- mes() blocks become OpCode sequences
- Array operations convert correctly

---

### Task 2.5: OpCode Serialization (for Embedded Mode)
**Goal**: Generate Go source code from OpCode for build-time compilation.

**Subtasks**:
- [x] 2.5.1 Implement OpCodeGenerator (serialize OpCode to Go source)
- [x] 2.5.2 Implement function OpCode serialization
- [x] 2.5.3 Implement variable declaration serialization
- [x] 2.5.4 Implement asset reference tracking
- [x] 2.5.5 Implement #info metadata serialization
- [x] 2.5.6 Add tests for OpCode serialization

**Acceptance Criteria**:
- OpCode serializes to valid Go source code
- Generated code is type-safe
- Generated code is human-readable
- All OpCode types supported
- Metadata preserved in embedded mode
- Generated code is type-safe
- Generated code is human-readable
- All OpCode types supported

---

### Task 2.6: step(n) { ... } Block Syntax Support
**Goal**: Implement block-style step syntax for timed loops.

**Subtasks**:
- [x] 2.6.1 Add StepBlockStatement to AST
- [x] 2.6.2 Implement parser support for step(n) { ... } syntax
- [x] 2.6.3 Implement CodeGen for StepBlockStatement (convert to flat sequence with waits)
- [x] 2.6.4 Add tests for step block parsing
- [x] 2.6.5 Add tests for step block code generation

**Acceptance Criteria**:
- step(n) { ... } syntax parses correctly
- Block body executes as flat sequence with waits between commands
- Empty steps (from commas) generate additional waits
- end_step stops code generation
- Compatible with both TIME and MIDI_TIME modes
- KUMA2 sample script works correctly

**Note**: This syntax is used in legacy FILLY scripts like KUMA2. The block generates a flat sequence: cmd1, wait(n), cmd2, wait(n), ... instead of a loop.

**Implementation Notes**:
- Changed from loop-based to flat sequence approach
- Modified executeWait() to not decrement PC (allows natural advancement)
- Modified IsComplete() to check IsWaiting() first
- All tests passing

---

## Phase 3: Execution Layer

### Task 3.1: VM Core
**Goal**: Implement clean VM that executes OpCode uniformly.

**Subtasks**:
- [x] 3.1.1 Implement ExecuteOp dispatcher (switch on OpCmd)
- [x] 3.1.2 Implement OpAssign handler
- [x] 3.1.3 Implement OpCall handler (stub for now)
- [x] 3.1.4 Implement OpIf handler
- [x] 3.1.5 Implement OpFor handler
- [x] 3.1.6 Implement OpWhile handler
- [x] 3.1.7 Implement OpWait handler
- [x] 3.1.8 Add unit tests for each OpCode handler

**Acceptance Criteria**:
- Single execution path through ExecuteOp
- All OpCode types handled correctly
- Clean error handling with context

---

### Task 3.2: Sequence Management
**Goal**: Implement sequence lifecycle and concurrent execution.

**Subtasks**:
- [x] 3.2.1 Implement RegisterSequence (create sequencer, link parent scope)
- [x] 3.2.2 Implement sequence activation/deactivation
- [x] 3.2.3 Implement del_me (deactivate current sequence)
- [x] 3.2.4 Implement del_us (deactivate group)
- [x] 3.2.5 Implement del_all (cleanup all sequences)
- [x] 3.2.6 Add tests for sequence lifecycle

**Acceptance Criteria**:
- Sequences register without blocking (except TIME mode)
- Multiple sequences execute concurrently
- Sequence termination doesn't affect other sequences
- Clean resource cleanup

---

### Task 3.3: mes() Block Support
**Goal**: Implement all mes() event types.

**Subtasks**:
- [x] 3.3.1 Implement mes(TIME) - blocking, frame-driven
- [x] 3.3.2 Implement mes(MIDI_TIME) - non-blocking, MIDI-driven
- [x] 3.3.3 Implement mes(MIDI_END) - MIDI completion event
- [x] 3.3.4 Implement mes(KEY) - keyboard input event
- [x] 3.3.5 Implement mes(CLICK) - mouse click event
- [x] 3.3.6 Implement mes(RBDOWN) - right button down event
- [x] 3.3.7 Implement mes(RBDBLCLK) - right button double-click event
- [x] 3.3.8 Implement mes(USER) - custom message event
- [x] 3.3.9 Add tests for all mes() types

**Acceptance Criteria**:
- TIME mode blocks until completion
- MIDI_TIME mode returns immediately
- Event handlers trigger on appropriate events
- Multiple handlers can respond to same event
- Event parameters passed via MesP1-MesP4

---

### Task 3.4: Tick-Driven Execution
**Goal**: Implement step-based execution model.

**Subtasks**:
- [x] 3.4.1 Implement UpdateVM (process one tick for all sequences)
- [x] 3.4.2 Implement wait counter decrement
- [x] 3.4.3 Implement program counter advancement
- [x] 3.4.4 Implement sequence completion detection
- [x] 3.4.5 Add tests for tick processing

**Acceptance Criteria**:
- Each tick advances all active sequences by one step
- Wait operations pause sequences correctly
- Sequences complete when pc reaches end
- Independent execution for each sequence

---

### Task 3.5: Timing Modes
**Goal**: Implement TIME and MIDI_TIME modes correctly.

**Subtasks**:
- [x] 3.5.1 Implement step(n) interpretation for TIME mode (n × 50ms)
- [x] 3.5.2 Implement step(n) interpretation for MIDI_TIME mode (n × 32nd note)
- [x] 3.5.3 Implement non-blocking behavior for MIDI_TIME mode
- [x] 3.5.4 Add tests for both timing modes

**Acceptance Criteria**:
- Step duration calculated correctly for each mode
- No mixing of timing mode logic
- MIDI_TIME mode is non-blocking

**Note**: TIME mode blocking behavior requires main loop integration and will be verified in Task 7.3.

---

### Task 3.6: Function Return Values
**Goal**: Implement function return value handling for variable assignment.

**Subtasks**:
- [x] 3.6.1 Modify built-in functions to return values via special variable
- [x] 3.6.2 Implement return value capture in OpCall handler
- [x] 3.6.3 Implement return value assignment in OpAssign handler (already working via evaluateValue)
- [x] 3.6.4 Support function calls in expressions (e.g., x = CreatePic(w, h) + 1)
- [x] 3.6.5 Add tests for return value assignment
- [x] 3.6.6 Add tests for return values in expressions

**Acceptance Criteria**:
- Built-in functions return values correctly (LoadPic, CreatePic, OpenWin, etc.) ✅
- Return values can be assigned to variables ✅
- Return values can be used in expressions ✅
- User-defined functions can return values (not yet implemented, but mechanism is ready)
- test_drawing sample works correctly with canvas = CreatePic(400, 300) (blocked by preprocessor issue)

**Implementation Notes**:
- Used special variable `__return__` to pass return values
- Modified `evaluateExpression` to capture return values from `OpCall`
- Return value is cleared after reading to prevent stale values
- All tests passing in `pkg/engine/return_value_test.go`
- Functions modified: LoadPic, CreatePic, PicWidth, PicHeight, OpenWin, GetPicNo, PutCast, LoadRsc, GetColor

**Note**: This enables `canvas = CreatePic(400, 300)` and similar patterns used in sample scripts.

---

## Phase 4: Graphics Foundation

### Task 4.1: Virtual Desktop
**Goal**: Implement fixed 1280×720 virtual desktop.

**Subtasks**:
- [x] 4.1.1 Define virtual desktop dimensions (1280×720)
- [x] 4.1.2 Implement WinInfo(0) - return desktop width
- [x] 4.1.3 Implement WinInfo(1) - return desktop height
- [x] 4.1.4 Add tests for virtual desktop

**Acceptance Criteria**:
- Virtual desktop is always 1280×720
- WinInfo returns correct dimensions

---

### Task 4.2: Picture Management
**Goal**: Implement picture loading, creation, and management.

**Subtasks**:
- [x] 4.2.1 Define Picture struct (ID, Image, Width, Height)
- [x] 4.2.2 Implement LoadPic using AssetLoader (supports both filesystem and embedded)
- [x] 4.2.3 Implement CreatePic (create empty buffer)
- [x] 4.2.4 Implement MovePic (copy pixels with transparency)
- [x] 4.2.5 Implement DelPic (release resources)
- [x] 4.2.6 Implement PicWidth, PicHeight queries
- [x] 4.2.7 Implement MoveSPic (scale and copy)
- [x] 4.2.8 Implement ReversePic (horizontal flip)
- [x] 4.2.9 Add tests for picture operations with both AssetLoader implementations

**Acceptance Criteria**:
- BMP loading works via AssetLoader (both modes)
- Sequential ID assignment
- Transparency handled correctly
- Resource cleanup on deletion
- Scaling and flipping preserve transparency
- Tests verify both filesystem and embedded loading

---

### Task 4.3: Window Management
**Goal**: Implement window creation and management with drag support.

**Subtasks**:
- [x] 4.3.1 Define Window struct (ID, PictureID, Position, Size, Caption, SrcX, SrcY)
- [x] 4.3.2 Implement OpenWin (create window with PicX/PicY offset support)
- [x] 4.3.3 Fix PicX/PicY offset inversion for legacy compatibility (-picX, -picY)
- [x] 4.3.4 Implement MoveWin (update properties)
- [x] 4.3.5 Implement CloseWin (close window)
- [x] 4.3.6 Implement CloseWinAll (close all windows)
- [x] 4.3.7 Implement CapTitle (set caption)
- [x] 4.3.8 Implement GetPicNo (query picture ID)
- [x] 4.3.9 Implement window drag detection (mouse down on title bar)
- [x] 4.3.10 Implement window drag update (mouse move while dragging)
- [x] 4.3.11 Implement window drag constraints (keep within virtual desktop)
- [x] 4.3.12 Add tests for window operations including drag

**Acceptance Criteria**:
- Windows display pictures correctly with proper offset handling
- PicX/PicY offsets inverted for legacy compatibility (window.SrcX = -picX, window.SrcY = -picY)
- Images larger than windows display correctly (centered using negative offsets)
- Window properties update correctly
- Windows render in creation order
- All windows within virtual desktop
- Windows with captions can be dragged by title bar
- Dragged windows constrain to desktop bounds
- Smooth drag interaction with real-time updates

**Critical Implementation Note**:
- In `OpenWindow()`, offsets must be inverted: `window.SrcX = -picX`, `window.SrcY = -picY`
- This is essential for legacy FILLY script compatibility
- Allows centering large images in small windows using negative offsets

---

### Task 4.4: Cast (Sprite) Management
**Goal**: Implement sprite creation and management.

**Subtasks**:
- [x] 4.4.1 Define Cast struct (ID, PictureID, Position, Clipping, Transparency)
- [x] 4.4.2 Implement PutCast (create sprite with clipping)
- [x] 4.4.3 Implement MoveCast (update position)
- [x] 4.4.4 Implement DelCast (remove sprite)
- [x] 4.4.5 Implement z-ordering (creation order)
- [x] 4.4.6 Add tests for cast operations

**Acceptance Criteria**:
- Sprites created with transparency and clipping
- Creation order determines z-order
- Position updates work correctly
- Sprites render within windows

---

### Task 4.5: Renderer Implementation
**Goal**: Implement clean renderer with mock support.

**Subtasks**:
- [x] 4.5.1 Implement EbitenRenderer (production)
- [x] 4.5.1 Implement EbitenRenderer (production)
- [x] 4.5.2 Enhance MockRenderer (testing)
- [x] 4.5.3 Implement rendering pipeline (desktop → windows → casts)
- [x] 4.5.4 Implement double buffering
- [x] 4.5.5 Implement render mutex for thread safety
- [x] 4.5.6 Implement classic desktop-style window decorations (title bar, 3D borders)
- [x] 4.5.7 Add tests for rendering

**Acceptance Criteria**:
- Renderer reads state but doesn't modify it
- MockRenderer enables headless testing
- Rendering is stateless
- Thread-safe rendering
- Windows display with classic desktop-style decorations (blue title bar, gray 3D borders)

---

## Phase 5: Audio System

### Task 5.1: MIDI Player
**Goal**: Implement MIDI playback with tick generation using MeltySynth.

**Subtasks**:
- [x] 5.1.1 Integrate MeltySynth library (github.com/sinshu/go-meltysynth/meltysynth)
- [x] 5.1.2 Implement SoundFont (.sf2) loading
- [x] 5.1.2.1 Implement automatic SoundFont discovery and loading on engine initialization
- [x] 5.1.3 Implement PlayMIDI using AssetLoader (supports both filesystem and embedded)
- [x] 5.1.4 Implement MIDI file parsing (tempo map, PPQ extraction)
- [x] 5.1.5 Implement MidiStream (io.Reader) with MeltySynth sequencer
- [x] 5.1.6 Implement wall-clock time based tick calculation (MIDI ticks, not 32nd notes)
- [x] 5.1.7 Fix CalculateTickFromTime() to calculate MIDI ticks directly
- [x] 5.1.8 Implement sequential tick delivery (no skipping)
- [x] 5.1.9 Implement MIDI end detection (currentTick >= totalTicks)
- [x] 5.1.10 Implement MIDI_END event triggering
- [x] 5.1.11 Fix mes(MIDI_TIME) to execute immediately (not as event handler)
- [x] 5.1.12 Implement UpdateMIDISequences() for MIDI tick-based wait updates
- [x] 5.1.13 Fix step(n) calculation for MIDI_TIME mode (n × PPQ/8 ticks)
- [x] 5.1.14 Add tests for MIDI playback with both AssetLoader implementations
- [x] 5.1.15 Add tests for tempo changes and tick accuracy

**Acceptance Criteria**:
- MIDI playback runs in background goroutine (audio thread)
- MIDI loading works via AssetLoader (both modes)
- SoundFont (.sf2) files load correctly
- SoundFont auto-loads from project directory (default.sf2, GeneralUser-GS.sf2, or *.sf2)
- Tick calculation uses wall-clock time and calculates MIDI ticks directly (not 32nd notes)
- step(n) in MIDI_TIME mode correctly calculates n × (PPQ/8) MIDI ticks
- mes(MIDI_TIME) blocks execute immediately, allowing PlayMIDI() to be called inside
- MIDI_TIME sequences have wait counters decremented by MIDI ticks (not frame ticks)
- Ticks delivered sequentially without skipping
- Tempo changes handled correctly via tempo map
- Accurate tick timing based on tempo and PPQ
- MIDI_END event triggers when playback completes
- MIDI continues after starting sequence terminates

**Critical Implementation Notes**:
- `CalculateTickFromTime()` must calculate MIDI ticks directly: `elapsed * (tempo/60) * PPQ`
- `step(n)` in MIDI_TIME mode: `n × (PPQ / 8)` MIDI ticks (not just n)
- `mes(MIDI_TIME)` executes immediately (calls RegisterSequence directly, not as event handler)
- Separate `UpdateMIDISequences()` method updates MIDI_TIME sequences with MIDI ticks
- Never mix TIME and MIDI_TIME tick updates (causes incorrect wait behavior)

---

### Task 5.2: WAV Player
**Goal**: Implement WAV playback.

**Subtasks**:
- [x] 5.2.1 Implement PlayWAVE using AssetLoader (supports both filesystem and embedded)
- [x] 5.2.2 Implement concurrent playback support
- [x] 5.2.3 Implement resource preloading (LoadRsc, PlayRsc, DelRsc)
- [x] 5.2.4 Add tests for WAV playback with both AssetLoader implementations

**Acceptance Criteria**:
- WAV playback runs asynchronously
- WAV loading works via AssetLoader (both modes)
- Multiple WAV files can play concurrently
- Preloading enables fast playback start

---

## Phase 6: Runtime Features

### Task 6.1: Text Rendering
**Goal**: Implement text drawing on pictures with Japanese font support.

**Subtasks**:
- [x] 6.1.1 Implement SetFont (load TrueType fonts from system paths)
- [x] 6.1.2 Implement font loading for .ttf and .ttc files
- [x] 6.1.3 Implement font fallback chain (Hiragino → Arial Unicode → basicfont)
- [x] 6.1.4 Implement TextWrite (draw text with anti-aliasing artifact prevention)
- [x] 6.1.5 Implement text area clearing before drawing (opaque white fill)
- [x] 6.1.6 Implement TextColor, BgColor (set colors)
- [x] 6.1.7 Implement BackMode (transparent/opaque background)
- [x] 6.1.8 Update to modern API (os.ReadFile instead of ioutil.ReadFile)
- [x] 6.1.9 Add tests for text rendering

**Acceptance Criteria**:
- Text draws correctly on pictures
- Japanese fonts supported (Hiragino on macOS)
- TrueType font loading works for both .ttf and .ttc files
- Font fallback chain works correctly
- Anti-aliasing artifacts prevented (no shadow from previous text)
- Transparent background works
- Modern Go APIs used (no deprecated functions)

**Implementation Notes**:
- System fonts searched in order: Hiragino Mincho → Hiragino Kaku Gothic → Arial Unicode → basicfont
- Text area cleared with opaque white before drawing to prevent alpha blending artifacts
- Font collections (.ttc) supported by extracting first font
- Legacy font size handling (size > 200 treated as parameter order issue)

---

### Task 6.2: Drawing Functions
**Goal**: Implement vector drawing primitives.

**Subtasks**:
- [x] 6.2.1 Implement DrawLine
- [x] 6.2.2 Implement DrawCircle
- [x] 6.2.3 Implement DrawRect
- [x] 6.2.4 Implement SetLineSize, SetPaintColor
- [x] 6.2.5 Implement raster operations (ROP)
- [x] 6.2.6 Implement GetColor (pixel query)
- [x] 6.2.7 Add tests for drawing functions

**Acceptance Criteria**:
- Drawing primitives work correctly
- Fill modes supported (outline, hatch, solid)
- Raster operations work (COPYPEN, XORPEN, MERGEPEN, NOTCOPYPEN, MASKPEN)
- All 13 drawing tests pass

**Note**: Function return values (Task 3.6) required for full test_drawing sample compatibility.

---

### Task 6.3: File Operations
**Goal**: Implement file I/O.

**Subtasks**:
- [ ] 6.3.1 Implement INI file operations (WriteIniInt, GetIniInt, WriteIniStr, GetIniStr)
- [ ] 6.3.2 Implement binary file I/O (OpenF, CloseF, ReadF, WriteF, SeekF)
- [ ] 6.3.3 Implement file management (CopyFile, DelFile, IsExist, MkDir, RmDir, ChDir, GetCwd)
- [ ] 6.3.4 Add tests for file operations

**Acceptance Criteria**:
- INI files read/write correctly
- Binary I/O works correctly
- File management operations work

---

### Task 6.4: String Operations
**Goal**: Implement string manipulation functions.

**Subtasks**:
- [x] 6.4.1 Implement StrLen, SubStr, StrFind
- [x] 6.4.2 Implement StrPrint (printf-style)
- [x] 6.4.3 Implement StrUp, StrLow (case conversion)
- [x] 6.4.4 Implement CharCode, StrCode
- [ ] 6.4.5 Implement StrInput (user input)
- [x] 6.4.6 Add tests for string operations

**Acceptance Criteria**:
- String operations work correctly
- Printf-style formatting works
- Case conversion works

---

### Task 6.5: Array Operations
**Goal**: Implement dynamic array manipulation.

**Subtasks**:
- [x] 6.5.1 Implement ArraySize
- [x] 6.5.2 Implement DelArrayAll
- [x] 6.5.3 Implement DelArrayAt, InsArrayAt
- [x] 6.5.4 Add tests for array operations

**Acceptance Criteria**:
- Arrays resize automatically
- Insert/delete operations work correctly

---

### Task 6.6: System Functions
**Goal**: Implement system information and utility functions.

**Subtasks**:
- [x] 6.6.1 Implement Random (random number generation)
- [x] 6.6.2 Implement GetSysTime, WhatDay, WhatTime (time/date)
- [ ] 6.6.3 Implement GetCmdLine (command line args)
- [ ] 6.6.4 Implement Shell (external process execution)
- [ ] 6.6.5 Implement MakeLong, GetHiWord, GetLowWord (bit operations)
- [ ] 6.6.6 Add tests for system functions

**Acceptance Criteria**:
- Random numbers generated correctly
- Time/date functions work
- External processes launch correctly
- Bit operations preserve patterns

---

### Task 6.7: Message System
**Goal**: Implement message management functions.

**Subtasks**:
- [x] 6.7.1 Implement GetMesNo (query message number)
- [x] 6.7.2 Implement DelMes (terminate specific message)
- [x] 6.7.3 Implement FreezeMes, ActivateMes (pause/resume)
- [x] 6.7.4 Implement PostMes (send custom message)
- [x] 6.7.5 Add tests for message system

**Acceptance Criteria**:
- Message queries work correctly
- Message control functions work
- Custom messages delivered correctly

---

## Phase 7: Integration and Testing

### Task 7.1: Embedded Executable Generation
**Goal**: Implement build-time compilation and embedded mode with multi-project support.

**Subtasks**:
- [ ] 7.1.1 Implement CLI mode detection (direct vs single-project vs multi-project)
- [ ] 7.1.2 Implement build configuration structure (EmbeddedProject)
- [ ] 7.1.3 Create build script/tool for generating embedded executables
- [ ] 7.1.4 Implement embedded project registration (build tags)
- [ ] 7.1.5 Implement multi-project menu UI
- [ ] 7.1.6 Implement project selection (keyboard and mouse)
- [ ] 7.1.7 Implement context-aware ESC handling (menu vs project)
- [ ] 7.1.8 Implement project lifecycle management (init, run, cleanup, return to menu)
- [ ] 7.1.9 Implement state reset between projects
- [ ] 7.1.10 Test single-project embedded executable generation
- [ ] 7.1.11 Test multi-project embedded executable generation
- [ ] 7.1.12 Test embedded executable execution (single and multi-project)
- [ ] 7.1.13 Test headless mode with embedded executables
- [ ] 7.1.14 Test ESC behavior in all contexts (menu, single-project, multi-project)

**Acceptance Criteria**:
- Build command creates standalone executable (single or multi-project)
- Embedded executable runs without external TFY files
- Assets load from embedded data
- Multi-project mode displays menu on startup
- User can select projects from menu
- ESC in project returns to menu (multi-project mode)
- ESC in menu terminates program
- ESC in single-project mode terminates program
- Project completion returns to menu (multi-project mode)
- Clean state reset between project runs
- Headless mode works with embedded executables
- Build tags control which projects are embedded
- Generated executables are cross-platform

---

### Task 7.2: Property-Based Tests
**Goal**: Implement PBT for correctness properties.

**Subtasks**:
- [ ] 7.2.1 Write property tests for OpCode execution
- [ ] 7.2.2 Write property tests for variable scoping
- [ ] 7.2.3 Write property tests for timing accuracy
- [ ] 7.2.4 Write property tests for concurrent execution
- [ ] 7.2.5 Write property tests for sequence lifecycle
- [ ] 7.2.6 Write property tests for AssetLoader implementations

**Acceptance Criteria**:
- Properties verify universal invariants
- Tests catch edge cases
- Properties linked to requirements
- Both AssetLoader modes tested

---

### Task 7.3: Integration Tests
**Goal**: Test complete sample scripts in both modes.

**Subtasks**:
- [ ] 7.3.1 Test kuma2 sample in direct mode (TIME mode)
- [ ] 7.3.2 Test y-saru sample in direct mode (MIDI_TIME mode)
- [ ] 7.3.3 Test yosemiya sample in direct mode (multiple sequences)
- [ ] 7.3.4 Test robot sample in direct mode (MIDI_END event)
- [ ] 7.3.5 Test kuma2 sample in embedded mode
- [ ] 7.3.6 Test y-saru sample in embedded mode
- [ ] 7.3.7 Verify 60 FPS tick generation for TIME mode
- [ ] 7.3.8 Verify TIME mode blocking behavior (mes(TIME) blocks until completion)
- [ ] 7.3.9 Add headless integration tests for both modes

**Acceptance Criteria**:
- All sample scripts execute correctly in direct mode
- All sample scripts execute correctly in embedded mode
- Behavior is identical between modes
- Timing behavior matches expectations (60 FPS for TIME mode)
- TIME mode mes() blocks until sequence completes
- Headless tests pass in CI/CD for both modes

---

### Task 7.4: Backward Compatibility Tests
**Goal**: Ensure existing scripts work in both modes.

**Subtasks**:
- [ ] 7.4.1 Test all sample scripts from _old_implementation in direct mode
- [ ] 7.4.2 Test all sample scripts from _old_implementation in embedded mode
- [ ] 7.4.3 Compare behavior with old implementation
- [ ] 7.4.4 Document any intentional behavior changes
- [ ] 7.4.5 Fix compatibility issues

**Acceptance Criteria**:
- All existing scripts work without modification in both modes
- Timing behavior is compatible
- No regressions
- Embedded mode produces identical behavior to direct mode

---

## Phase 8: Documentation and Cleanup

### Task 8.1: Code Documentation
**Goal**: Document all public APIs.

**Subtasks**:
- [ ] 8.1.1 Add godoc comments to all public types
- [ ] 8.1.2 Add godoc comments to all public functions
- [ ] 8.1.3 Add package-level documentation
- [ ] 8.1.4 Add examples in documentation

**Acceptance Criteria**:
- All public APIs documented
- Examples provided
- Documentation is clear and helpful

---

### Task 8.2: Architecture Documentation
**Goal**: Update architecture documentation.

**Subtasks**:
- [ ] 8.2.1 Update design.md with final architecture
- [ ] 8.2.2 Document key design decisions
- [ ] 8.2.3 Document testing strategy
- [ ] 8.2.4 Document deployment process
- [ ] 8.2.5 Document embedded executable generation workflow
- [ ] 8.2.6 Document build configuration examples

**Acceptance Criteria**:
- Design document reflects actual implementation
- Design decisions explained
- Testing strategy documented
- Embedded mode workflow documented with examples

---

### Task 8.3: Migration Guide
**Goal**: Document migration from old implementation.

**Subtasks**:
- [ ] 8.3.1 Document API changes
- [ ] 8.3.2 Document behavior changes
- [ ] 8.3.3 Provide migration examples
- [ ] 8.3.4 Document breaking changes

**Acceptance Criteria**:
- Migration guide is complete
- Examples provided
- Breaking changes documented

---

## Notes

**Implementation Order Rationale**:
1. **Foundation first**: Core data structures, AssetLoader, and array support enable everything else
2. **Infrastructure early**: Logging and testing infrastructure needed throughout
3. **Compilation before execution**: Preprocessor → Lexer → Parser → OpCode; OpCode serialization enables embedded mode; step(n) { ... } syntax support
4. **Execution core early**: mes() is central to FILLY, implement early; function return values enable variable assignment
5. **Graphics layer by layer**: Desktop → Picture → Window → Cast (dependency order); LoadPic uses AssetLoader
6. **Audio after graphics**: Audio is independent but benefits from working graphics; uses AssetLoader
7. **Runtime features last**: These depend on working execution and graphics; includes array operations
8. **Integration testing**: Verify everything works together; test both direct and embedded modes
9. **Documentation**: Document the final, working system including embedded mode workflow and preprocessor directives

**Testing Strategy**:
- Write tests before or alongside implementation
- Use property-based tests for universal properties
- Use unit tests for specific examples
- Use integration tests for end-to-end behavior

**Reference Code**:
- Existing code in `_old_implementation/` can be referenced
- Don't copy implementation details blindly
- Focus on simplicity and correctness
- Follow design.md principles

**Recent Additions**:
- **Task 2.6**: step(n) { ... } block syntax support (required for KUMA2 sample)
- **Task 3.6**: Function return value handling (required for test_drawing sample)
- **Task 6.2**: Completed drawing functions implementation
