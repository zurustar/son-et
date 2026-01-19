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

## Phase 1: System Infrastructure

### Task 1.1: Logging System
**Goal**: Implement timestamped logging with debug levels.

**Subtasks**:
- [ ] 1.1.1 Implement timestamp formatting [HH:MM:SS.mmm]
- [ ] 1.1.2 Implement DEBUG_LEVEL support (0, 1, 2)
- [ ] 1.1.3 Implement log functions (LogError, LogInfo, LogDebug)
- [ ] 1.1.4 Add tests for logging

**Acceptance Criteria**:
- Timestamps in all logs
- Debug levels control verbosity (0=errors, 1=info, 2=debug)
- Thread-safe logging

---

### Task 1.2: Headless Mode
**Goal**: Implement headless execution for testing.

**Subtasks**:
- [ ] 1.2.1 Implement --headless flag parsing
- [ ] 1.2.2 Implement MockRenderer for headless mode
- [ ] 1.2.3 Implement headless execution loop (60 FPS)
- [ ] 1.2.4 Implement rendering operation logging
- [ ] 1.2.5 Add tests for headless mode

**Acceptance Criteria**:
- Scripts execute without GUI
- All logic executes normally
- Rendering operations logged
- 60 FPS tick generation

---

### Task 1.3: Program Termination
**Goal**: Implement clean program termination.

**Subtasks**:
- [ ] 1.3.1 Implement --timeout flag parsing and timer
- [ ] 1.3.2 Implement programTerminated flag (atomic)
- [ ] 1.3.3 Implement ESC key detection and termination
- [ ] 1.3.4 Implement graceful shutdown (cleanup resources)
- [ ] 1.3.5 Add tests for termination

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
- [ ] 1.4.1 Implement parsing error reporting (line, column)
- [ ] 1.4.2 Implement runtime error reporting (OpCode, args)
- [ ] 1.4.3 Implement asset loading error reporting (filename)
- [ ] 1.4.4 Add tests for error reporting

**Acceptance Criteria**:
- Parsing errors include line/column
- Runtime errors include OpCode context
- Asset errors include filename
- Errors logged with appropriate level

---

## Phase 2: Compilation Layer

### Task 2.1: Lexer
**Goal**: Clean, minimal lexer that tokenizes FILLY source code.

**Subtasks**:
- [ ] 2.1.1 Implement token types (keywords, operators, literals, identifiers)
- [ ] 2.1.2 Implement lexer with position tracking
- [ ] 2.1.3 Implement error reporting with line/column
- [ ] 2.1.4 Add comprehensive lexer tests

**Acceptance Criteria**:
- Lexer handles all FILLY syntax
- Clear error messages with line/column numbers
- Case-insensitive keywords

---

### Task 2.2: Parser
**Goal**: Clean parser that builds AST from tokens.

**Subtasks**:
- [ ] 2.2.1 Define AST node types
- [ ] 2.2.2 Implement expression parsing (precedence climbing)
- [ ] 2.2.3 Implement statement parsing
- [ ] 2.2.4 Implement control flow parsing (if, for, while, switch)
- [ ] 2.2.5 Implement function definition parsing
- [ ] 2.2.6 Implement mes() block parsing
- [ ] 2.2.7 Add comprehensive parser tests

**Acceptance Criteria**:
- Parser handles all FILLY syntax
- Clear error messages with line/column numbers
- AST structure is clean and minimal

---

### Task 2.3: OpCode Generation
**Goal**: Convert AST to OpCode sequences uniformly.

**Subtasks**:
- [ ] 2.3.1 Implement statement conversion (assignments, function calls)
- [ ] 2.3.2 Implement expression conversion (arithmetic, comparisons, nested)
- [ ] 2.3.3 Implement control flow conversion (if, for, while, switch)
- [ ] 2.3.4 Implement function definition conversion
- [ ] 2.3.5 Implement mes() block conversion
- [ ] 2.3.6 Add comprehensive OpCode generation tests

**Acceptance Criteria**:
- All FILLY constructs convert to OpCode
- Nested expressions become nested OpCode
- Control flow uses OpCode with nested blocks
- mes() blocks become OpCode sequences

---

## Phase 3: Execution Layer

### Task 3.1: VM Core
**Goal**: Implement clean VM that executes OpCode uniformly.

**Subtasks**:
- [ ] 3.1.1 Implement ExecuteOp dispatcher (switch on OpCmd)
- [ ] 3.1.2 Implement OpAssign handler
- [ ] 3.1.3 Implement OpCall handler (stub for now)
- [ ] 3.1.4 Implement OpIf handler
- [ ] 3.1.5 Implement OpFor handler
- [ ] 3.1.6 Implement OpWhile handler
- [ ] 3.1.7 Implement OpWait handler
- [ ] 3.1.8 Add unit tests for each OpCode handler

**Acceptance Criteria**:
- Single execution path through ExecuteOp
- All OpCode types handled correctly
- Clean error handling with context

---

### Task 3.2: Sequence Management
**Goal**: Implement sequence lifecycle and concurrent execution.

**Subtasks**:
- [ ] 3.2.1 Implement RegisterSequence (create sequencer, link parent scope)
- [ ] 3.2.2 Implement sequence activation/deactivation
- [ ] 3.2.3 Implement del_me (deactivate current sequence)
- [ ] 3.2.4 Implement del_us (deactivate group)
- [ ] 3.2.5 Implement del_all (cleanup all sequences)
- [ ] 3.2.6 Add tests for sequence lifecycle

**Acceptance Criteria**:
- Sequences register without blocking (except TIME mode)
- Multiple sequences execute concurrently
- Sequence termination doesn't affect other sequences
- Clean resource cleanup

---

### Task 3.3: mes() Block Support
**Goal**: Implement all mes() event types.

**Subtasks**:
- [ ] 3.3.1 Implement mes(TIME) - blocking, frame-driven
- [ ] 3.3.2 Implement mes(MIDI_TIME) - non-blocking, MIDI-driven
- [ ] 3.3.3 Implement mes(MIDI_END) - MIDI completion event
- [ ] 3.3.4 Implement mes(KEY) - keyboard input event
- [ ] 3.3.5 Implement mes(CLICK) - mouse click event
- [ ] 3.3.6 Implement mes(RBDOWN) - right button down event
- [ ] 3.3.7 Implement mes(RBDBLCLK) - right button double-click event
- [ ] 3.3.8 Implement mes(USER) - custom message event
- [ ] 3.3.9 Add tests for all mes() types

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
- [ ] 3.4.1 Implement UpdateVM (process one tick for all sequences)
- [ ] 3.4.2 Implement wait counter decrement
- [ ] 3.4.3 Implement program counter advancement
- [ ] 3.4.4 Implement sequence completion detection
- [ ] 3.4.5 Implement tick generation for TIME mode (60 FPS)
- [ ] 3.4.6 Add tests for tick processing

**Acceptance Criteria**:
- Each tick advances all active sequences by one step
- Wait operations pause sequences correctly
- Sequences complete when pc reaches end
- Independent execution for each sequence

---

### Task 3.5: Timing Modes
**Goal**: Implement TIME and MIDI_TIME modes correctly.

**Subtasks**:
- [ ] 3.5.1 Implement step(n) interpretation for TIME mode (n × 50ms)
- [ ] 3.5.2 Implement step(n) interpretation for MIDI_TIME mode (n × 32nd note)
- [ ] 3.5.3 Implement blocking behavior for TIME mode
- [ ] 3.5.4 Implement non-blocking behavior for MIDI_TIME mode
- [ ] 3.5.5 Add tests for both timing modes

**Acceptance Criteria**:
- Step duration calculated correctly for each mode
- No mixing of timing mode logic
- TIME mode blocks, MIDI_TIME doesn't

---

## Phase 4: Graphics Foundation

### Task 4.1: Virtual Desktop
**Goal**: Implement fixed 1280×720 virtual desktop.

**Subtasks**:
- [ ] 4.1.1 Define virtual desktop dimensions (1280×720)
- [ ] 4.1.2 Implement WinInfo(0) - return desktop width
- [ ] 4.1.3 Implement WinInfo(1) - return desktop height
- [ ] 4.1.4 Add tests for virtual desktop

**Acceptance Criteria**:
- Virtual desktop is always 1280×720
- WinInfo returns correct dimensions

---

### Task 4.2: Picture Management
**Goal**: Implement picture loading, creation, and management.

**Subtasks**:
- [ ] 4.2.1 Define Picture struct (ID, Image, Width, Height)
- [ ] 4.2.2 Implement LoadPic (load BMP, assign ID)
- [ ] 4.2.3 Implement CreatePic (create empty buffer)
- [ ] 4.2.4 Implement MovePic (copy pixels with transparency)
- [ ] 4.2.5 Implement DelPic (release resources)
- [ ] 4.2.6 Implement PicWidth, PicHeight queries
- [ ] 4.2.7 Implement MoveSPic (scale and copy)
- [ ] 4.2.8 Implement ReversePic (horizontal flip)
- [ ] 4.2.9 Add tests for picture operations

**Acceptance Criteria**:
- BMP loading works correctly
- Sequential ID assignment
- Transparency handled correctly
- Resource cleanup on deletion
- Scaling and flipping preserve transparency

---

### Task 4.3: Window Management
**Goal**: Implement window creation and management.

**Subtasks**:
- [ ] 4.3.1 Define Window struct (ID, PictureID, Position, Size, Caption)
- [ ] 4.3.2 Implement OpenWin (create window)
- [ ] 4.3.3 Implement MoveWin (update properties)
- [ ] 4.3.4 Implement CloseWin (close window)
- [ ] 4.3.5 Implement CloseWinAll (close all windows)
- [ ] 4.3.6 Implement CapTitle (set caption)
- [ ] 4.3.7 Implement GetPicNo (query picture ID)
- [ ] 4.3.8 Add tests for window operations

**Acceptance Criteria**:
- Windows display pictures correctly
- Window properties update correctly
- Windows render in creation order
- All windows within virtual desktop

---

### Task 4.4: Cast (Sprite) Management
**Goal**: Implement sprite creation and management.

**Subtasks**:
- [ ] 4.4.1 Define Cast struct (ID, PictureID, Position, Clipping, Transparency)
- [ ] 4.4.2 Implement PutCast (create sprite with clipping)
- [ ] 4.4.3 Implement MoveCast (update position)
- [ ] 4.4.4 Implement DelCast (remove sprite)
- [ ] 4.4.5 Implement z-ordering (creation order)
- [ ] 4.4.6 Add tests for cast operations

**Acceptance Criteria**:
- Sprites created with transparency and clipping
- Creation order determines z-order
- Position updates work correctly
- Sprites render within windows

---

### Task 4.5: Renderer Implementation
**Goal**: Implement clean renderer with mock support.

**Subtasks**:
- [ ] 4.5.1 Implement EbitenRenderer (production)
- [ ] 4.5.2 Enhance MockRenderer (testing)
- [ ] 4.5.3 Implement rendering pipeline (desktop → windows → casts)
- [ ] 4.5.4 Implement double buffering
- [ ] 4.5.5 Implement render mutex for thread safety
- [ ] 4.5.6 Add tests for rendering

**Acceptance Criteria**:
- Renderer reads state but doesn't modify it
- MockRenderer enables headless testing
- Rendering is stateless
- Thread-safe rendering

---

## Phase 5: Audio System

### Task 5.1: MIDI Player
**Goal**: Implement MIDI playback with tick generation.

**Subtasks**:
- [ ] 5.1.1 Implement PlayMIDI (load and start playback)
- [ ] 5.1.2 Implement MIDI tick calculation (tempo, PPQ)
- [ ] 5.1.3 Implement tick callbacks to VM
- [ ] 5.1.4 Implement MIDI_END event triggering
- [ ] 5.1.5 Add tests for MIDI playback

**Acceptance Criteria**:
- MIDI playback runs in background goroutine
- Tick callbacks drive MIDI_TIME sequences
- Accurate tick timing based on tempo
- MIDI continues after starting sequence terminates

---

### Task 5.2: WAV Player
**Goal**: Implement WAV playback.

**Subtasks**:
- [ ] 5.2.1 Implement PlayWAVE (decode and play)
- [ ] 5.2.2 Implement concurrent playback support
- [ ] 5.2.3 Implement resource preloading (LoadRsc, PlayRsc, DelRsc)
- [ ] 5.2.4 Add tests for WAV playback

**Acceptance Criteria**:
- WAV playback runs asynchronously
- Multiple WAV files can play concurrently
- Preloading enables fast playback start

---

## Phase 6: Runtime Features

### Task 6.1: Text Rendering
**Goal**: Implement text drawing on pictures.

**Subtasks**:
- [ ] 6.1.1 Implement SetFont (load font)
- [ ] 6.1.2 Implement TextWrite (draw text)
- [ ] 6.1.3 Implement TextColor, BgColor (set colors)
- [ ] 6.1.4 Implement BackMode (transparent background)
- [ ] 6.1.5 Add tests for text rendering

**Acceptance Criteria**:
- Text draws correctly on pictures
- Japanese fonts supported
- Transparent background works

---

### Task 6.2: Drawing Functions
**Goal**: Implement vector drawing primitives.

**Subtasks**:
- [ ] 6.2.1 Implement DrawLine
- [ ] 6.2.2 Implement DrawCircle
- [ ] 6.2.3 Implement DrawRect
- [ ] 6.2.4 Implement SetLineSize, SetPaintColor
- [ ] 6.2.5 Implement raster operations (ROP)
- [ ] 6.2.6 Implement GetColor (pixel query)
- [ ] 6.2.7 Add tests for drawing functions

**Acceptance Criteria**:
- Drawing primitives work correctly
- Fill modes supported
- Raster operations work

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
- [ ] 6.4.1 Implement StrLen, SubStr, StrFind
- [ ] 6.4.2 Implement StrPrint (printf-style)
- [ ] 6.4.3 Implement StrUp, StrLow (case conversion)
- [ ] 6.4.4 Implement CharCode, StrCode
- [ ] 6.4.5 Implement StrInput (user input)
- [ ] 6.4.6 Add tests for string operations

**Acceptance Criteria**:
- String operations work correctly
- Printf-style formatting works
- Case conversion works

---

### Task 6.5: Array Operations
**Goal**: Implement dynamic array manipulation.

**Subtasks**:
- [ ] 6.5.1 Implement ArraySize
- [ ] 6.5.2 Implement DelArrayAll
- [ ] 6.5.3 Implement DelArrayAt, InsArrayAt
- [ ] 6.5.4 Add tests for array operations

**Acceptance Criteria**:
- Arrays resize automatically
- Insert/delete operations work correctly

---

### Task 6.6: System Functions
**Goal**: Implement system information and utility functions.

**Subtasks**:
- [ ] 6.6.1 Implement Random (random number generation)
- [ ] 6.6.2 Implement GetSysTime, WhatDay, WhatTime (time/date)
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
- [ ] 6.7.1 Implement GetMesNo (query message number)
- [ ] 6.7.2 Implement DelMes (terminate specific message)
- [ ] 6.7.3 Implement FreezeMes, ActivateMes (pause/resume)
- [ ] 6.7.4 Implement PostMes (send custom message)
- [ ] 6.7.5 Add tests for message system

**Acceptance Criteria**:
- Message queries work correctly
- Message control functions work
- Custom messages delivered correctly

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
- [ ] 7.3.1 Test all sample scripts from _old_implementation
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

**Implementation Order Rationale**:
1. **Foundation first**: Core data structures enable everything else
2. **Infrastructure early**: Logging and testing infrastructure needed throughout
3. **Compilation before execution**: Need OpCode before VM can run
4. **Execution core early**: mes() is central to FILLY, implement early
5. **Graphics layer by layer**: Desktop → Picture → Window → Cast (dependency order)
6. **Audio after graphics**: Audio is independent but benefits from working graphics
7. **Runtime features last**: These depend on working execution and graphics
8. **Integration testing**: Verify everything works together
9. **Documentation**: Document the final, working system

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
