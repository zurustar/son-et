# Implementation Tasks: son-et Core Engine

## Overview

This task list implements the requirements defined in [requirements.md](requirements.md) following the architecture described in [design.md](design.md).

**Strategy**: Incremental refactoring - build clean modules alongside existing code, then replace module by module.

**Reference**: Existing code in `main` branch can be referenced but should not constrain the new design.

---

## Phase 0: Foundation Layer

### Task 0.1: Core Data Structures
**Goal**: Define clean, minimal data structures for OpCode and execution state.

**Subtasks**:
- [ ] 0.1.1 Define OpCmd enum type with all command types
- [ ] 0.1.2 Define OpCode struct (Cmd, Args)
- [ ] 0.1.3 Define Variable type wrapper
- [ ] 0.1.4 Add unit tests for OpCode construction

**Acceptance Criteria**:
- OpCmd uses enum (not strings) for type safety
- OpCode supports nested OpCode in Args
- Variable type distinguishes variables from literals

---

### Task 0.2: Sequencer Structure
**Goal**: Define the execution context for sequences with clean scope management.

**Subtasks**:
- [ ] 0.2.1 Define Sequencer struct (commands, pc, active, mode, waitCount, stepSize, vars, parent)
- [ ] 0.2.2 Implement NewSequencer constructor
- [ ] 0.2.3 Implement variable resolution with scope chain walking
- [ ] 0.2.4 Add unit tests for variable scoping

**Acceptance Criteria**:
- Sequencer maintains parent pointer for scope chain
- Variable lookup walks up scope chain correctly
- Case-insensitive variable resolution
- Default values for undefined variables (0, "", [])

---

### Task 0.3: Interface Definitions
**Goal**: Define clean interfaces for dependency injection.

**Subtasks**:
- [ ] 0.3.1 Define Renderer interface
- [ ] 0.3.2 Define AssetLoader interface
- [ ] 0.3.3 Define ImageDecoder interface
- [ ] 0.3.4 Define TickGenerator interface

**Acceptance Criteria**:
- Interfaces are minimal and focused
- Interfaces enable testing without platform dependencies
- Interfaces support both production and mock implementations

---

## Phase 1: Compilation Layer

### Task 1.1: Lexer Refactoring
**Goal**: Clean, minimal lexer that tokenizes FILLY source code.

**Subtasks**:
- [ ] 1.1.1 Review existing lexer for simplification opportunities
- [ ] 1.1.2 Ensure all FILLY tokens are supported
- [ ] 1.1.3 Add comprehensive lexer tests
- [ ] 1.1.4 Document token types and lexing rules

**Acceptance Criteria**:
- Lexer handles all FILLY syntax
- Clear error messages with line/column numbers
- No unnecessary complexity

---


### Task 1.2: Parser Refactoring
**Goal**: Clean parser that builds AST from tokens.

**Subtasks**:
- [ ] 1.2.1 Review existing parser for simplification opportunities
- [ ] 1.2.2 Ensure all FILLY constructs are parsed correctly
- [ ] 1.2.3 Add comprehensive parser tests
- [ ] 1.2.4 Document AST structure

**Acceptance Criteria**:
- Parser handles all FILLY syntax
- Clear error messages with line/column numbers
- AST structure is clean and minimal

---

### Task 1.3: OpCode Generation (Interpreter)
**Goal**: Convert AST to OpCode sequences uniformly.

**Subtasks**:
- [ ] 1.3.1 Implement statement conversion (assignments, function calls)
- [ ] 1.3.2 Implement expression conversion (arithmetic, comparisons, nested expressions)
- [ ] 1.3.3 Implement control flow conversion (if, for, while, switch)
- [ ] 1.3.4 Implement function definition conversion
- [ ] 1.3.5 Implement mes() block conversion
- [ ] 1.3.6 Add comprehensive OpCode generation tests

**Acceptance Criteria**:
- All FILLY constructs convert to OpCode
- Nested expressions become nested OpCode
- Control flow uses OpCode with nested blocks
- mes() blocks become OpCode sequences

---

## Phase 2: Execution Layer

### Task 2.1: VM Core
**Goal**: Implement clean VM that executes OpCode uniformly.

**Subtasks**:
- [ ] 2.1.1 Implement ExecuteOp dispatcher (switch on OpCmd)
- [ ] 2.1.2 Implement OpAssign handler
- [ ] 2.1.3 Implement OpCall handler
- [ ] 2.1.4 Implement OpIf handler
- [ ] 2.1.5 Implement OpFor handler
- [ ] 2.1.6 Implement OpWhile handler
- [ ] 2.1.7 Implement OpWait handler
- [ ] 2.1.8 Add unit tests for each OpCode handler

**Acceptance Criteria**:
- Single execution path through ExecuteOp
- All OpCode types handled correctly
- Clean error handling with context

---

### Task 2.2: Sequence Management
**Goal**: Implement sequence lifecycle and concurrent execution.

**Subtasks**:
- [ ] 2.2.1 Implement RegisterSequence (create sequencer, link parent scope)
- [ ] 2.2.2 Implement sequence activation/deactivation
- [ ] 2.2.3 Implement del_me (deactivate current sequence)
- [ ] 2.2.4 Implement del_us (deactivate group)
- [ ] 2.2.5 Implement del_all (cleanup all sequences)
- [ ] 2.2.6 Add tests for sequence lifecycle

**Acceptance Criteria**:
- Sequences register without blocking (except TIME mode)
- Multiple sequences execute concurrently
- Sequence termination doesn't affect other sequences
- Clean resource cleanup

---

### Task 2.3: Tick-Driven Execution
**Goal**: Implement step-based execution model.

**Subtasks**:
- [ ] 2.3.1 Implement UpdateVM (process one tick for all sequences)
- [ ] 2.3.2 Implement wait counter decrement
- [ ] 2.3.3 Implement program counter advancement
- [ ] 2.3.4 Implement sequence completion detection
- [ ] 2.3.5 Add tests for tick processing

**Acceptance Criteria**:
- Each tick advances all active sequences by one step
- Wait operations pause sequences correctly
- Sequences complete when pc reaches end
- Independent execution for each sequence

---

### Task 2.4: Timing Modes
**Goal**: Implement TIME and MIDI_TIME modes correctly.

**Subtasks**:
- [ ] 2.4.1 Implement TIME mode (blocking RegisterSequence)
- [ ] 2.4.2 Implement MIDI_TIME mode (non-blocking RegisterSequence)
- [ ] 2.4.3 Implement step(n) interpretation for each mode
- [ ] 2.4.4 Implement tick generation for TIME mode (60 FPS)
- [ ] 2.4.5 Implement tick generation for MIDI_TIME mode (MIDI callbacks)
- [ ] 2.4.6 Add tests for both timing modes

**Acceptance Criteria**:
- TIME mode blocks until sequence completes
- MIDI_TIME mode returns immediately
- Step duration calculated correctly for each mode
- No mixing of timing mode logic

---

## Phase 3: Graphics System

### Task 3.1: State Management
**Goal**: Implement clean graphics state with thread safety.

**Subtasks**:
- [ ] 3.1.1 Define Picture struct (ID, Image, Width, Height)
- [ ] 3.1.2 Define Cast struct (ID, PictureID, Position, Clipping, Transparency)
- [ ] 3.1.3 Define Window struct (ID, PictureID, Position, Size, Caption)
- [ ] 3.1.4 Implement render mutex for thread safety
- [ ] 3.1.5 Add tests for state management

**Acceptance Criteria**:
- Clean data structures without unnecessary fields
- Thread-safe access to graphics state
- Sequential ID assignment

---

### Task 3.2: Picture Operations
**Goal**: Implement picture loading, creation, and management.

**Subtasks**:
- [ ] 3.2.1 Implement LoadPic (load BMP, assign ID)
- [ ] 3.2.2 Implement CreatePic (create empty buffer)
- [ ] 3.2.3 Implement MovePic (copy pixels with transparency)
- [ ] 3.2.4 Implement DelPic (release resources)
- [ ] 3.2.5 Implement PicWidth, PicHeight queries
- [ ] 3.2.6 Add tests for picture operations

**Acceptance Criteria**:
- BMP loading works correctly
- Sequential ID assignment
- Transparency handled correctly
- Resource cleanup on deletion

---

### Task 3.3: Cast (Sprite) Operations
**Goal**: Implement sprite creation and management.

**Subtasks**:
- [ ] 3.3.1 Implement PutCast (create sprite with clipping)
- [ ] 3.3.2 Implement MoveCast (update position)
- [ ] 3.3.3 Implement DelCast (remove sprite)
- [ ] 3.3.4 Implement z-ordering (creation order)
- [ ] 3.3.5 Add tests for cast operations

**Acceptance Criteria**:
- Sprites created with transparency and clipping
- Creation order determines z-order
- Position updates work correctly

---

### Task 3.4: Window Operations
**Goal**: Implement window creation and management.

**Subtasks**:
- [ ] 3.4.1 Implement OpenWin (create window)
- [ ] 3.4.2 Implement MoveWin (update properties)
- [ ] 3.4.3 Implement CloseWin (close window)
- [ ] 3.4.4 Implement WinInfo (query desktop dimensions)
- [ ] 3.4.5 Add tests for window operations

**Acceptance Criteria**:
- Windows display pictures correctly
- Window properties update correctly
- Virtual desktop is 1280×720

---

### Task 3.5: Renderer Implementation
**Goal**: Implement clean renderer with mock support.

**Subtasks**:
- [ ] 3.5.1 Implement EbitenRenderer (production)
- [ ] 3.5.2 Implement MockRenderer (testing)
- [ ] 3.5.3 Implement rendering pipeline (windows → casts → scale)
- [ ] 3.5.4 Implement double buffering
- [ ] 3.5.5 Add tests for rendering

**Acceptance Criteria**:
- Renderer reads state but doesn't modify it
- MockRenderer enables headless testing
- Rendering is stateless (no frame dependencies)

---

## Phase 4: Audio System

### Task 4.1: MIDI Player
**Goal**: Implement MIDI playback with tick generation.

**Subtasks**:
- [ ] 4.1.1 Implement PlayMIDI (load and start playback)
- [ ] 4.1.2 Implement MIDI tick calculation (tempo, PPQ)
- [ ] 4.1.3 Implement tick callbacks to VM
- [ ] 4.1.4 Implement MIDI_END event
- [ ] 4.1.5 Add tests for MIDI playback

**Acceptance Criteria**:
- MIDI playback runs in background goroutine
- Tick callbacks drive MIDI_TIME sequences
- Accurate tick timing based on tempo
- MIDI continues after starting sequence terminates

---

### Task 4.2: WAV Player
**Goal**: Implement WAV playback.

**Subtasks**:
- [ ] 4.2.1 Implement PlayWAVE (decode and play)
- [ ] 4.2.2 Implement concurrent playback support
- [ ] 4.2.3 Implement resource preloading (LoadRsc, PlayRsc, DelRsc)
- [ ] 4.2.4 Add tests for WAV playback

**Acceptance Criteria**:
- WAV playback runs asynchronously
- Multiple WAV files can play concurrently
- Preloading enables fast playback start

---

## Phase 5: Runtime Features

### Task 5.1: Text Rendering
**Goal**: Implement text drawing on pictures.

**Subtasks**:
- [ ] 5.1.1 Implement SetFont (load font)
- [ ] 5.1.2 Implement TextWrite (draw text)
- [ ] 5.1.3 Implement TextColor, BgColor (set colors)
- [ ] 5.1.4 Implement BackMode (transparent background)
- [ ] 5.1.5 Add tests for text rendering

**Acceptance Criteria**:
- Text draws correctly on pictures
- Japanese fonts supported
- Transparent background works

---

### Task 5.2: Drawing Functions
**Goal**: Implement vector drawing primitives.

**Subtasks**:
- [ ] 5.2.1 Implement DrawLine
- [ ] 5.2.2 Implement DrawCircle
- [ ] 5.2.3 Implement DrawRect
- [ ] 5.2.4 Implement SetLineSize, SetPaintColor
- [ ] 5.2.5 Implement raster operations (ROP)
- [ ] 5.2.6 Add tests for drawing functions

**Acceptance Criteria**:
- Drawing primitives work correctly
- Fill modes supported
- Raster operations work

---

### Task 5.3: File Operations
**Goal**: Implement file I/O.

**Subtasks**:
- [ ] 5.3.1 Implement INI file operations (WriteIniInt, GetIniInt, WriteIniStr, GetIniStr)
- [ ] 5.3.2 Implement binary file I/O (OpenF, CloseF, ReadF, WriteF, SeekF)
- [ ] 5.3.3 Implement file management (CopyFile, DelFile, IsExist, MkDir, RmDir, ChDir, GetCwd)
- [ ] 5.3.4 Add tests for file operations

**Acceptance Criteria**:
- INI files read/write correctly
- Binary I/O works correctly
- File management operations work

---

### Task 5.4: String Operations
**Goal**: Implement string manipulation functions.

**Subtasks**:
- [ ] 5.4.1 Implement StrLen, SubStr, StrFind
- [ ] 5.4.2 Implement StrPrint (printf-style)
- [ ] 5.4.3 Implement StrUp, StrLow (case conversion)
- [ ] 5.4.4 Implement CharCode, StrCode
- [ ] 5.4.5 Add tests for string operations

**Acceptance Criteria**:
- String operations work correctly
- Printf-style formatting works
- Case conversion works

---

### Task 5.5: Array Operations
**Goal**: Implement dynamic array manipulation.

**Subtasks**:
- [ ] 5.5.1 Implement ArraySize
- [ ] 5.5.2 Implement DelArrayAll
- [ ] 5.5.3 Implement DelArrayAt, InsArrayAt
- [ ] 5.5.4 Add tests for array operations

**Acceptance Criteria**:
- Arrays resize automatically
- Insert/delete operations work correctly

---

### Task 5.6: User Input Handling
**Goal**: Implement keyboard and mouse input event handling.

**Subtasks**:
- [ ] 5.6.1 Implement mes(KEY) event registration and triggering
- [ ] 5.6.2 Implement mes(CLICK) event registration and triggering
- [ ] 5.6.3 Implement mes(RBDOWN) event registration and triggering
- [ ] 5.6.4 Implement mes(RBDBLCLK) event registration and triggering
- [ ] 5.6.5 Implement ESC key special handling (programTerminated flag)
- [ ] 5.6.6 Implement event parameter passing (MesP1-MesP4)
- [ ] 5.6.7 Implement PostMes for custom messages
- [ ] 5.6.8 Add tests for input event handling

**Acceptance Criteria**:
- Keyboard events trigger KEY sequences
- Mouse events trigger appropriate sequences
- ESC key immediately sets termination flag
- Termination check happens before VM execution
- Multiple sequences can respond to same event
- Event parameters passed correctly
- PostMes delivers custom messages

---

## Phase 6: Development Features

### Task 6.1: Headless Mode
**Goal**: Implement headless execution for testing.

**Subtasks**:
- [ ] 6.1.1 Implement --headless flag parsing
- [ ] 6.1.2 Implement headless execution loop (60 FPS)
- [ ] 6.1.3 Implement rendering operation logging
- [ ] 6.1.4 Add tests for headless mode

**Acceptance Criteria**:
- Scripts execute without GUI
- All logic executes normally
- Rendering operations logged

---

### Task 6.2: Auto-Termination
**Goal**: Implement timeout for automated testing.

**Subtasks**:
- [ ] 6.2.1 Implement --timeout flag parsing
- [ ] 6.2.2 Implement timeout timer
- [ ] 6.2.3 Implement graceful shutdown on timeout
- [ ] 6.2.4 Add tests for timeout

**Acceptance Criteria**:
- Timeout formats supported (5s, 500ms, 2m)
- Graceful shutdown on timeout
- Exit code 0 on timeout

---

### Task 6.3: Logging System
**Goal**: Implement timestamped logging with debug levels.

**Subtasks**:
- [ ] 6.3.1 Implement timestamp formatting [HH:MM:SS.mmm]
- [ ] 6.3.2 Implement DEBUG_LEVEL support (0, 1, 2)
- [ ] 6.3.3 Implement logging for VM execution
- [ ] 6.3.4 Implement logging for asset loading
- [ ] 6.3.5 Implement logging for rendering

**Acceptance Criteria**:
- Timestamps in all logs
- Debug levels control verbosity
- Logs help verify timing accuracy

---

### Task 6.4: Error Reporting
**Goal**: Implement clear error messages.

**Subtasks**:
- [ ] 6.4.1 Implement parsing error reporting (line, column)
- [ ] 6.4.2 Implement runtime error reporting (OpCode, args)
- [ ] 6.4.3 Implement asset loading error reporting (filename)
- [ ] 6.4.4 Add tests for error reporting

**Acceptance Criteria**:
- Parsing errors include line/column
- Runtime errors include OpCode context
- Asset errors include filename

---

## Phase 7: Integration and Testing

### Task 7.1: Property-Based Tests
**Goal**: Implement PBT for correctness properties.

**Subtasks**:
- [ ] 7.1.1 Write property tests for OpCode execution
- [ ] 7.1.2 Write property tests for variable scoping
- [ ] 7.1.3 Write property tests for timing accuracy
- [ ] 7.1.4 Write property tests for concurrent execution
- [ ] 7.1.5 Write property tests for sequence lifecycle

**Acceptance Criteria**:
- Properties verify universal invariants
- Tests catch edge cases
- Properties linked to requirements

---

### Task 7.2: Integration Tests
**Goal**: Test complete sample scripts.

**Subtasks**:
- [ ] 7.2.1 Test kuma2 sample (TIME mode)
- [ ] 7.2.2 Test y-saru sample (MIDI_TIME mode)
- [ ] 7.2.3 Test yosemiya sample (multiple sequences)
- [ ] 7.2.4 Test robot sample (MIDI_END event)
- [ ] 7.2.5 Add headless integration tests

**Acceptance Criteria**:
- All sample scripts execute correctly
- Timing behavior matches expectations
- Headless tests pass in CI/CD

---

### Task 7.3: Backward Compatibility Tests
**Goal**: Ensure existing scripts work.

**Subtasks**:
- [ ] 7.3.1 Test all sample scripts from main branch
- [ ] 7.3.2 Compare behavior with old implementation
- [ ] 7.3.3 Document any intentional behavior changes
- [ ] 7.3.4 Fix compatibility issues

**Acceptance Criteria**:
- All existing scripts work without modification
- Timing behavior is compatible
- No regressions

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

**Acceptance Criteria**:
- Design document reflects actual implementation
- Design decisions explained
- Testing strategy documented

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

**Testing Strategy**:
- Write tests before or alongside implementation
- Use property-based tests for universal properties
- Use unit tests for specific examples
- Use integration tests for end-to-end behavior

**Reference Code**:
- Existing code in `main` branch can be referenced
- Don't copy implementation details blindly
- Focus on simplicity and correctness
- Follow design.md principles

**Incremental Approach**:
- Complete one phase before moving to next
- Each phase should have passing tests
- Commit frequently with clear messages
- Review and refactor as needed
