# Requirements Document: son-et Core Engine

## Introduction

son-et is an interpreter and runtime for FILLY scripts - a legacy scripting language for multimedia applications from the Windows 3.1 era. This document defines **what** the system must do, focusing on functional requirements, user needs, and acceptance criteria.

**Scope**: This document describes requirements (WHAT), not design (HOW). For architectural design and implementation patterns, see [design.md](design.md).

**Note:** For detailed FILLY language syntax and function reference, see [language_spec.md](../../reference/language_spec.md).

## Glossary

See [GLOSSARY.md](../GLOSSARY.md) for common terms used across all son-et specifications.

---

## Part 1: Core Execution Requirements

These requirements define the fundamental behavior of the FILLY execution model.

### Requirement A1: Uniform Code Execution

**User Story:** As a script author, I want all FILLY code to execute consistently, so that I can predict program behavior.

**Rationale:** FILLY scripts must execute with consistent semantics regardless of code complexity. All language constructs (statements, expressions, control flow, functions, mes() blocks) must be processed uniformly.

#### Acceptance Criteria

1. WHEN the system processes any FILLY code, THE System SHALL execute it with consistent semantics
2. WHEN the system encounters statements, THE System SHALL execute them correctly
3. WHEN the system encounters expressions, THE System SHALL evaluate them correctly
4. WHEN the system encounters control flow, THE System SHALL execute branches and loops correctly
5. WHEN the system encounters mes() blocks, THE System SHALL register and execute them correctly
6. WHEN the system encounters function definitions, THE System SHALL execute function bodies correctly
7. THE System SHALL maintain consistent variable scoping across all code constructs
8. WHEN the system encounters preprocessor directives (#info, #include), THE System SHALL process them before execution
9. WHEN the system encounters array operations, THE System SHALL handle dynamic arrays correctly

**Preprocessor Directives:**
10. WHEN a #info directive is encountered, THE System SHALL parse and store the metadata
11. WHEN a #include directive is encountered, THE System SHALL load and parse the specified TFY file
12. WHEN processing #include, THE System SHALL resolve file paths relative to the project directory
13. WHEN processing #include, THE System SHALL use case-insensitive file matching (Windows 3.1 compatibility)
14. WHEN a circular #include is detected, THE System SHALL report an error
15. THE System SHALL process #include directives recursively (included files can include other files)
16. THE System SHALL support all standard #info tags (INAM, IART, ICOP, GENR, WRIT, GRPC, COMP, PROD, CONT, MDFY, TRNS, JINT, VIDO, INST, ISBJ, ICMT)

**Character Encoding:**
17. WHEN reading TFY files, THE System SHALL detect if the file is Shift-JIS encoded
18. WHEN a TFY file is Shift-JIS encoded, THE System SHALL convert it to UTF-8 before parsing
19. WHEN conversion fails, THE System SHALL report a clear error message
20. THE System SHALL support both Shift-JIS and UTF-8 encoded TFY files

**Arrays:**
21. THE System SHALL support dynamic integer arrays with automatic resizing
22. THE System SHALL support 0-based array indexing
23. WHEN an array element is accessed beyond current size, THE System SHALL expand the array automatically
24. WHEN an uninitialized array element is read, THE System SHALL return 0
25. THE System SHALL support array manipulation functions (ArraySize, DelArrayAll, DelArrayAt, InsArrayAt)

### Requirement A2: Step-Based Execution Model

**User Story:** As a script author, I want my animations and interactions to advance in discrete steps synchronized with external events, so that timing-critical multimedia applications work correctly.

**Rationale:** FILLY scripts are designed for multimedia applications where precise timing is critical. Unlike traditional scripting languages that run continuously, FILLY scripts must advance one step at a time, driven by:
- MIDI playback ticks (for music-synchronized animations)
- Frame updates at 60 FPS (for time-based animations)
- User input events (for interactive applications)

#### Acceptance Criteria

1. WHEN a mes() block is registered, THE System SHALL NOT execute it immediately
2. WHEN an external event occurs (tick, frame, input), THE System SHALL advance execution by one step
3. WHEN a Wait(n) operation is executed, THE System SHALL pause that sequence for n steps
4. WHEN a step(n) operation is executed, THE System SHALL set the step duration for subsequent operations
5. THE System SHALL maintain separate execution state for each active sequence
6. THE System SHALL allow multiple sequences to execute concurrently
7. THE System SHALL advance each sequence independently based on its timing mode

### Requirement A3: Dual Timing Modes

**User Story:** As a script author, I want to synchronize animations either with MIDI music or with real time, so that I can create both music-synchronized and procedural animations.

**Rationale:** FILLY supports two fundamentally different timing models:
- **MIDI_TIME Mode**: Animations synchronized to music (tempo-dependent)
- **TIME Mode**: Animations synchronized to real time (frame-based)

These modes have different execution characteristics and must be supported correctly.

#### Acceptance Criteria

**MIDI_TIME Mode:**
1. WHEN mes(MIDI_TIME) is called, THE System SHALL register the sequence without blocking
2. WHEN mes(MIDI_TIME) is called, THE System SHALL return immediately to allow subsequent code execution
3. WHEN PlayMIDI is called, THE System SHALL drive MIDI_TIME sequences via MIDI tick callbacks
4. WHEN step(n) is used in MIDI_TIME mode, THE System SHALL interpret it as n × (32nd note duration)
5. THE System SHALL allow PlayMIDI() to be called after the mes(MIDI_TIME) block

**TIME Mode:**
6. WHEN mes(TIME) is called, THE System SHALL register the sequence and block until completion
7. WHEN mes(TIME) is called, THE System SHALL ensure sequential execution (mes completes before next statement)
8. WHEN the game loop updates, THE System SHALL drive TIME sequences via 60 FPS frame updates
9. WHEN step(n) is used in TIME mode, THE System SHALL interpret it as n × 50ms
10. THE System SHALL NOT mix timing mode logic (applying TIME logic to MIDI_TIME or vice versa)

### Requirement A4: Lexical Variable Scoping

**User Story:** As a script author, I want variables to be accessible across function boundaries and mes() blocks according to lexical scoping rules, so that I can share state between different parts of my script.

**Rationale:** FILLY uses lexical scoping where:
- Global variables are accessible everywhere
- Function-local variables are scoped to that function
- mes() blocks inherit the scope of their enclosing function
- Variable lookup walks up the scope hierarchy

#### Acceptance Criteria

1. WHEN a variable is declared in the main function, THE System SHALL make it accessible to all mes() blocks defined in main
2. WHEN a variable is declared in a user function, THE System SHALL scope it to that function and its nested mes() blocks
3. WHEN a mes() block references a variable, THE System SHALL resolve it by walking up the scope chain
4. WHEN a variable is assigned, THE System SHALL update it in the scope where it was first declared
5. WHEN a variable is not found in any scope, THE System SHALL return a default value (0 for int, "" for string, [] for array)
6. THE System SHALL implement case-insensitive variable lookup (FILLY is case-insensitive)

### Requirement A5: Thread-Safe Concurrent Execution

**User Story:** As a developer, I want the system to handle concurrent access to shared state safely, so that scripts execute correctly without race conditions.

**Rationale:** FILLY scripts execute in a separate thread while the main thread handles rendering and input. Shared state must be protected from concurrent access.

#### Acceptance Criteria

1. WHEN the script thread modifies graphics state, THE System SHALL ensure thread-safe access
2. WHEN the main thread renders graphics, THE System SHALL ensure thread-safe access to graphics state
3. WHEN concurrent access occurs, THE System SHALL prevent race conditions
4. WHEN rendering occurs, THE System SHALL prevent flicker and visual artifacts
5. THE System SHALL allow the script thread and main thread to execute concurrently without deadlocks

### Requirement A6: Non-Blocking Audio Playback

**User Story:** As a script author, I want audio playback to run in the background without blocking my script, so that my application remains responsive.

**Rationale:** MIDI and WAV playback can take seconds or minutes. Blocking the script or game loop would freeze the entire application.

#### Acceptance Criteria

1. WHEN PlayMIDI is called, THE System SHALL start playback and return immediately
2. WHEN PlayWAVE is called, THE System SHALL start playback and return immediately
3. WHEN audio is playing, THE System SHALL continue script execution concurrently
4. WHEN audio is playing, THE System SHALL continue rendering and input handling
5. WHEN a sequence that started audio terminates, THE System SHALL continue audio playback
6. THE System SHALL support concurrent playback of multiple WAV files

---

## Part 2: Graphics and Multimedia Requirements

These requirements define the multimedia capabilities that scripts can use.

### Requirement B1: Graphics Rendering System

**User Story:** As a script author, I want to load images, create sprites, and display them in windows, so that I can create visual applications.

**Function Categories:** Picture management, Cast/Sprite management, Window management, Drawing functions

**See:** [language_spec.md](../../reference/language_spec.md) for detailed function specifications.

#### Acceptance Criteria

**Virtual Desktop:**
1. THE System SHALL provide a fixed 1280×720 pixel virtual desktop
2. WHEN win_info(0) is called, THE System SHALL return 1280 (desktop width)
3. WHEN win_info(1) is called, THE System SHALL return 720 (desktop height)
4. THE System SHALL render all virtual windows within this 1280×720 canvas

**Picture Management:**
5. THE System SHALL support loading BMP images and assigning sequential Picture IDs
6. THE System SHALL support creating empty image buffers with specified dimensions
7. THE System SHALL support copying pixels between pictures with optional transparency
8. THE System SHALL support querying picture dimensions (pic_width, pic_height)
9. THE System SHALL support releasing picture resources

**Cast (Sprite) Management:**
10. THE System SHALL support creating sprites with transparency and clipping regions
11. THE System SHALL maintain creation order for z-ordering (painter's algorithm)
12. THE System SHALL support updating sprite positions and re-rendering
13. THE System SHALL support sprite sheet clipping (srcX, srcY, width, height)
14. THE System SHALL support removing sprites

**Window Management:**
15. THE System SHALL support creating virtual windows displaying pictures
16. THE System SHALL support updating window properties (position, size, picture)
17. THE System SHALL support window captions
18. THE System SHALL render windows in creation order within the virtual desktop
19. WHEN a window has a caption, THE System SHALL allow the user to drag the window by its title bar
20. WHEN the user drags a window, THE System SHALL update the window position in real-time
21. THE System SHALL constrain dragged windows to remain within the virtual desktop bounds

### Requirement B2: Audio Playback System

**User Story:** As a script author, I want to play MIDI music and WAV sound effects, so that my application has audio.

**Function Categories:** MIDI playback, WAV playback, Resource management

**See:** [language_spec.md](../../reference/language_spec.md) for detailed function specifications.

#### Acceptance Criteria

**MIDI Playback:**
1. THE System SHALL support loading and playing Standard MIDI Files (SMF)
2. THE System SHALL use SoundFont (.sf2) files for software synthesis
3. THE System SHALL generate tick callbacks at 32nd note resolution to drive MIDI_TIME mode execution
4. THE System SHALL calculate ticks accurately based on elapsed time, tempo, and PPQ (Pulses Per Quarter note)
5. WHEN MIDI tempo changes occur, THE System SHALL adjust tick timing accordingly
6. THE System SHALL play MIDI files once (no looping by default)
7. THE System SHALL trigger MIDI_END event when playback completes
8. THE System SHALL deliver all ticks sequentially without skipping (even if processing is delayed)

**WAV Playback:**
6. THE System SHALL support decoding and playing WAV files
7. THE System SHALL support concurrent playback of multiple WAV files

**Resource Management:**
8. THE System SHALL support preloading WAV files with resource IDs
9. THE System SHALL support playing preloaded resources
10. THE System SHALL support releasing preloaded resources

### Requirement B3: User Input Handling

**User Story:** As a script author, I want to respond to user input events (keyboard, mouse), so that I can create interactive applications.

**Rationale:** FILLY supports event-driven programming through `mes()` blocks that respond to user input. This enables interactive multimedia applications.

#### Acceptance Criteria

**Keyboard Input:**
1. WHEN mes(KEY) is registered, THE System SHALL trigger the sequence when any key is pressed
2. THE System SHALL provide key code information to the sequence
3. THE System SHALL support multiple KEY sequences concurrently

**Mouse Input:**
4. WHEN mes(CLICK) is registered, THE System SHALL trigger the sequence when the mouse is clicked
5. WHEN mes(RBDOWN) is registered, THE System SHALL trigger the sequence when the right mouse button is pressed
6. WHEN mes(RBDBLCLK) is registered, THE System SHALL trigger the sequence when the right mouse button is double-clicked
7. THE System SHALL provide mouse position information to the sequence
8. THE System SHALL support multiple mouse event sequences concurrently

**ESC Key Handling:**
9. WHEN the ESC key is pressed, THE System SHALL set a termination flag
10. WHEN the termination flag is set, THE System SHALL stop all VM execution immediately
11. WHEN the termination flag is set, THE System SHALL return termination signal to the game loop
12. THE System SHALL check the termination flag before executing any OpCode
13. THE System SHALL allow graceful cleanup before termination

**Event Delivery:**
14. THE System SHALL deliver events to all matching mes() blocks
15. THE System SHALL not block the main thread while delivering events
16. THE System SHALL queue events if sequences are busy

### Requirement B5: Text Rendering

**User Story:** As a script author, I want to draw text on images, so that I can display messages and UI elements.

**See:** [language_spec.md](../../reference/language_spec.md) for detailed function specifications.

#### Acceptance Criteria

1. THE System SHALL support loading fonts with specified size and attributes
2. THE System SHALL support drawing text on pictures at specified coordinates
3. THE System SHALL support setting foreground and background colors
4. THE System SHALL support transparent background mode
5. THE System SHALL support Japanese fonts (MS Gothic, MS Mincho) with charset 128

### Requirement B4: String Operations

**User Story:** As a script author, I want to manipulate strings, so that I can process text data.

**See:** [language_spec.md](../../reference/language_spec.md) for detailed function specifications.

#### Acceptance Criteria

1. THE System SHALL support querying string length
2. THE System SHALL support extracting substrings
3. THE System SHALL support searching within strings
4. THE System SHALL support formatted string output (printf-style)
5. THE System SHALL support case conversion and character code operations

### Requirement B5: Drawing Functions

**User Story:** As a script author, I want to draw lines, shapes, and manipulate pixels, so that I can create vector graphics.

**See:** [language_spec.md](../../reference/language_spec.md) for detailed function specifications.

#### Acceptance Criteria

1. THE System SHALL support drawing lines, circles, and rectangles
2. THE System SHALL support setting line width and drawing color
3. THE System SHALL support fill modes (none, hatch, solid)
4. THE System SHALL support raster operations (COPYPEN, XORPEN, MERGEPEN, etc.)
5. THE System SHALL support reading pixel colors

### Requirement B6: Picture Transformation

**User Story:** As a script author, I want to scale and flip images, so that I can create transformed graphics.

**See:** [language_spec.md](../../reference/language_spec.md) for detailed function specifications.

#### Acceptance Criteria

1. THE System SHALL support scaling images with arbitrary ratios
2. THE System SHALL support horizontal image flipping
3. THE System SHALL preserve transparency during transformations
4. THE System SHALL support querying picture IDs associated with windows

### Requirement B7: File Operations

**User Story:** As a script author, I want to read and write files, so that I can persist data.

**See:** [language_spec.md](../../reference/language_spec.md) for detailed function specifications.

#### Acceptance Criteria

1. THE System SHALL support reading and writing INI configuration files
2. THE System SHALL support binary file I/O with file handles
3. THE System SHALL support file and directory management operations
4. THE System SHALL create files and directories if they do not exist

### Requirement B8: Array Operations

**User Story:** As a script author, I want to manipulate dynamic arrays, so that I can manage collections of data.

**See:** [language_spec.md](../../reference/language_spec.md) for detailed function specifications.

#### Acceptance Criteria

1. THE System SHALL support querying array size
2. THE System SHALL support clearing all array elements
3. THE System SHALL support inserting and deleting elements at specific indices
4. THE System SHALL automatically resize arrays during operations

### Requirement B9: Control Flow Constructs

**User Story:** As a script author, I want to use if, for, while, and switch statements, so that I can implement logic.

**See:** [language_spec.md](../../reference/language_spec.md) for detailed function specifications.

#### Acceptance Criteria

1. THE System SHALL support if-else conditionals
2. THE System SHALL support for, while, and do-while loops
3. THE System SHALL support switch-case multi-way branching
4. THE System SHALL support break and continue statements in loops
5. THE System SHALL support nested control structures
6. THE System SHALL support comparison and logical operators

### Requirement B10: Function Definitions

**User Story:** As a script author, I want to define custom functions with parameters, so that I can organize code.

**See:** [language_spec.md](../../reference/language_spec.md) for detailed function specifications.

#### Acceptance Criteria

1. THE System SHALL support function definitions with typed parameters (int, str, int[])
2. THE System SHALL support default parameter values
3. THE System SHALL use case-insensitive function names
4. THE System SHALL create a new scope for each function call
5. THE System SHALL support recursive function calls

### Requirement B11: Message System

**User Story:** As a script author, I want to manage mes() blocks and generate custom messages, so that I can control script flow.

**See:** [language_spec.md](../../reference/language_spec.md) for detailed function specifications.

#### Acceptance Criteria

1. THE System SHALL support querying current and last executed mes block numbers
2. THE System SHALL support terminating specific mes blocks
3. THE System SHALL support pausing and resuming mes blocks
4. THE System SHALL support generating custom messages with parameters
5. THE System SHALL deliver messages to all matching mes blocks

### Requirement B12: System Information

**User Story:** As a script author, I want to query system information, so that I can adapt my script.

**See:** [language_spec.md](../../reference/language_spec.md) for detailed function specifications.

#### Acceptance Criteria

1. THE System SHALL return virtual desktop dimensions (1280×720) via win_info
2. THE System SHALL generate random integers in specified ranges
3. THE System SHALL provide current system time and date
4. THE System SHALL provide access to command line arguments

### Requirement B13: Sequence Lifecycle Control

**User Story:** As a script author, I want to control sequence execution lifecycle, so that I can terminate sequences and clean up resources.

**Rationale:** FILLY provides commands (del_me, del_us, del_all) to control sequence execution. These commands operate on sequences (mes() blocks), not on the entire program.

#### Acceptance Criteria

1. WHEN del_me is called, THE System SHALL deactivate the current sequence only
2. WHEN del_me is called, THE System SHALL NOT terminate other sequences or the program
3. WHEN del_us is called, THE System SHALL deactivate all sequences in the current group
4. WHEN del_all is called, THE System SHALL close all windows and clean up graphics resources
5. WHEN del_all is called, THE System SHALL NOT terminate MIDI playback or other sequences
6. THE System SHALL continue running other active sequences after del_me is called
7. THE System SHALL support graceful sequence termination without resource leaks
8. THE System SHALL allow these commands to be called without parentheses (e.g., `del_me` not `del_me()`)

### Requirement B14: Integer Operations

**User Story:** As a script author, I want to pack and unpack integer values, so that I can efficiently store data.

**See:** [language_spec.md](../../reference/language_spec.md) for detailed function specifications.

#### Acceptance Criteria

1. THE System SHALL support combining two 16-bit values into a 32-bit value
2. THE System SHALL support extracting upper and lower 16 bits
3. THE System SHALL preserve bit patterns during operations

### Requirement B15: Process Execution

**User Story:** As a script author, I want to launch external applications, so that I can integrate with other programs.

**See:** [language_spec.md](../../reference/language_spec.md) for detailed function specifications.

#### Acceptance Criteria

1. THE System SHALL support executing external applications with parameters
2. THE System SHALL set the working directory for launched processes
3. THE System SHALL return immediately without waiting for process completion
4. THE System SHALL handle process launch errors gracefully

---

## Part 3: Development and Testing Requirements

These requirements support development, testing, and debugging workflows.

### Requirement C1: Headless Mode

**User Story:** As a developer, I want to run scripts without GUI, so that I can test and debug in automated environments.

#### Acceptance Criteria

1. WHEN the `--headless` flag is provided, THE System SHALL execute scripts without creating a GUI window
2. WHEN running in headless mode, THE System SHALL execute all script logic (timing, audio, state changes) normally
3. WHEN running in headless mode, THE System SHALL skip rendering operations but log them
4. WHEN running in headless mode, THE System SHALL output all logs to stdout/stderr
5. THE System SHALL support both direct mode and embedded mode in headless operation
6. WHEN running in headless mode, THE System SHALL exit cleanly when the script completes

### Requirement C2: Auto-Termination

**User Story:** As a developer, I want scripts to automatically terminate after a duration, so that I can test without manual process management.

#### Acceptance Criteria

1. WHEN the `--timeout` flag is provided, THE System SHALL automatically terminate after that duration
2. THE System SHALL support timeout formats: seconds (5s), milliseconds (500ms), minutes (2m)
3. WHEN the timeout is reached, THE System SHALL gracefully shut down all resources
4. WHEN the timeout is reached, THE System SHALL exit with status code 0
5. THE System SHALL support timeout in both GUI and headless modes

### Requirement C3: Timestamped Logging

**User Story:** As a developer, I want timestamped logs, so that I can verify timing accuracy.

#### Acceptance Criteria

1. WHEN running with debug logging, THE System SHALL include timestamps in [HH:MM:SS.mmm] format
2. THE System SHALL log VM execution (operation execution, tick updates, wait operations)
3. THE System SHALL log asset loading (pictures, MIDI, WAV)
4. THE System SHALL log rendering operations in headless mode

### Requirement C4: Debug Levels

**User Story:** As a developer, I want to control logging verbosity, so that I can focus on relevant information during debugging.

#### Acceptance Criteria

1. THE System SHALL support multiple debug levels (0=errors only, 1=important operations, 2=all debug output)
2. WHEN DEBUG_LEVEL=0 is set, THE System SHALL log errors only
3. WHEN DEBUG_LEVEL=1 is set, THE System SHALL log important operations (default level)
4. WHEN DEBUG_LEVEL=2 is set, THE System SHALL log all debug output including execution details, timing, rendering, and asset loading
5. WHEN DEBUG_LEVEL is not set, THE System SHALL default to level 1

### Requirement C5: Direct Execution Mode

**User Story:** As a developer, I want to run TFY projects directly, so that I can iterate quickly.

#### Acceptance Criteria

1. WHEN a developer runs `son-et <directory>`, THE System SHALL execute the project immediately
2. WHEN executing in direct mode, THE System SHALL locate the main function in TFY files
3. WHEN executing in direct mode, THE System SHALL convert TFY to executable form at runtime
4. WHEN executing in direct mode, THE System SHALL load assets from the project directory
5. WHEN a runtime error occurs, THE System SHALL report the error with TFY script line numbers

### Requirement C6: Embedded Executable Generation

**User Story:** As a developer, I want to create standalone executables with one or more FILLY titles embedded, so that I can distribute applications or collections.

#### Acceptance Criteria

**Single-Title Mode:**
1. WHEN building with a title directory specified, THE System SHALL create an executable with the title embedded
2. WHEN the embedded executable runs, THE System SHALL execute the embedded title without external TFY files
3. WHEN the embedded executable runs, THE System SHALL load assets from embedded data
4. WHEN building with embedded title, THE System SHALL convert TFY to OpCode at build time
5. WHEN the embedded executable runs without arguments, THE System SHALL execute the embedded title directly
6. WHEN ESC is pressed during single-title execution, THE System SHALL terminate the program
7. THE System SHALL embed the title's directory using Go's embed.FS
8. THE System SHALL generate type-safe Go source code from OpCode sequences

**Multi-Title Mode:**
9. WHEN building with multiple titles specified, THE System SHALL create an executable with all titles embedded
10. THE System SHALL embed each title's directory separately to avoid asset filename conflicts
11. WHEN the multi-title executable runs, THE System SHALL display a title selection menu
12. THE System SHALL display title names and descriptions from #info metadata in the menu
13. WHEN the user selects a title from the menu, THE System SHALL execute that title
14. WHEN ESC is pressed during title execution in multi-title mode, THE System SHALL return to the title selection menu
15. WHEN a title completes normally in multi-title mode, THE System SHALL return to the title selection menu
16. WHEN ESC is pressed in the title selection menu, THE System SHALL terminate the program
17. THE System SHALL allow the user to select and run different titles without restarting the executable
18. THE System SHALL reset engine state between title executions
19. THE System SHALL create a separate AssetLoader for each title scoped to its directory

### Requirement C7: Error Reporting

**User Story:** As a developer, I want clear error messages, so that I can quickly identify and fix issues.

#### Acceptance Criteria

1. WHEN a parsing error occurs, THE System SHALL report the error with TFY script line and column numbers
2. WHEN a runtime error occurs, THE System SHALL report the error with the operation that failed
3. WHEN an asset loading error occurs, THE System SHALL report which asset file could not be loaded
4. WHEN a variable resolution error occurs, THE System SHALL report which variable was not found
5. WHEN a function call error occurs, THE System SHALL report which function was called incorrectly

---

## Part 4: Compatibility and Platform Requirements

### Requirement D1: Backward Compatibility

**User Story:** As a developer, I want existing TFY scripts to work without modification, so that I don't need to rewrite applications.

#### Acceptance Criteria

1. WHEN the interpreter processes an existing TFY script, THE System SHALL execute it correctly without modifications
2. WHEN the interpreter encounters FILLY language constructs, THE System SHALL support all existing syntax
3. WHEN the interpreter processes mes() blocks, THE System SHALL maintain timing behavior compatibility
4. WHEN the interpreter processes user-defined functions, THE System SHALL maintain calling convention compatibility
5. WHEN the interpreter processes variable declarations, THE System SHALL maintain scope behavior compatibility
6. THE System SHALL perform case-insensitive file matching for asset references (Windows 3.1 compatibility)

### Requirement D2: Cross-Platform Support

**User Story:** As a developer, I want son-et to run on modern platforms, so that legacy applications can be preserved.

#### Acceptance Criteria

1. THE System SHALL run natively on macOS
2. THE System SHALL support BMP image format (legacy format)
3. THE System SHALL support Standard MIDI File (SMF) format
4. THE System SHALL support WAV audio format
5. THE System SHALL use modern cross-platform libraries for graphics and audio

### Requirement D3: Legacy Feature Exclusions

**User Story:** As a developer, I want to know which legacy features are not supported, so that I can plan migrations.

#### Acceptance Criteria

1. THE System SHALL NOT implement CD audio playback (PlayCD) - obsolete hardware
2. THE System SHALL NOT implement MCI commands (MCI, StrMCI) - Windows-specific API
3. THE System SHALL NOT implement Windows Registry access (SetRegStr, GetRegStr) - Windows-specific
4. THE System SHALL NOT implement AVI video playback (PlayAVI) - complex codec support
5. THE Documentation SHALL clearly list unsupported features and suggest alternatives

**Migration Guidance:**
- CD audio → Use PlayMIDI or PlayWAVE with digital audio files
- MCI commands → Use PlayMIDI, PlayWAVE, or platform-specific alternatives
- Registry access → Use INI files (WriteIniInt, GetIniInt, WriteIniStr, GetIniStr)
- AVI playback → Use modern video formats with external players

---

## Summary

This requirements document defines **what** the son-et system must do to correctly execute FILLY scripts. The key requirements are:

1. **Uniform Code Execution** - All FILLY code executes with consistent semantics
2. **Step-Based Execution** - Scripts advance in discrete steps driven by external events
3. **Dual Timing Modes** - Support both MIDI-synchronized and real-time animations
4. **Lexical Variable Scoping** - Variables follow lexical scoping rules with scope chain resolution
5. **Thread-Safe Execution** - Concurrent access to shared state is safe
6. **Non-Blocking Audio** - Audio playback runs in the background without blocking

For architectural design and implementation patterns that satisfy these requirements, see [design.md](design.md).
