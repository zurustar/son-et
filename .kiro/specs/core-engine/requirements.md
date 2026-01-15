# Requirements Document: Core Engine

## Introduction

This document defines the requirements for the son-et core engine, a source-to-source compiler (transpiler) that converts legacy FILLY scripts into modern Go source code. The engine provides a complete runtime environment including graphics rendering, audio playback, timing synchronization, and script execution.

## Glossary

- **Transpiler**: A source-to-source compiler that converts FILLY script syntax into Go code
- **FILLY_Script**: Legacy scripting language with C-like syntax used for creating interactive multimedia applications
- **Virtual_Desktop**: A fixed 1280x720 rendering canvas that hosts multiple virtual windows
- **Virtual_Window**: A sub-region within the Virtual_Desktop that displays game content (typically 640x480)
- **Picture**: An image buffer that can be loaded, manipulated, and displayed
- **Cast**: A sprite object with transparency support and z-ordering
- **MIDI_Sync_Mode**: Timing mode where script execution is synchronized to MIDI playback ticks
- **Time_Mode**: Timing mode where script execution is driven by the game loop at 60 FPS
- **Script_Goroutine**: A separate thread that executes the transpiled user script
- **Main_Thread**: The primary thread running the game loop (Update/Draw at 60 FPS)
- **Render_Mutex**: A synchronization lock protecting shared rendering state
- **Step**: A timing unit whose meaning depends on the current timing mode
- **mes_Block**: A message-driven execution block that responds to events
- **SoundFont**: A .sf2 file containing instrument samples for MIDI synthesis

## Requirements

## Part A: Runtime Engine Requirements

### Requirement 1: Script Transpilation

**User Story:** As a developer, I want to convert FILLY scripts to Go source code, so that I can build native executables from legacy scripts.

#### Acceptance Criteria

1. WHEN a FILLY script file is provided as input, THE Transpiler SHALL parse it and generate valid Go source code
2. WHEN the generated Go code is compiled, THE Transpiler SHALL produce a working executable without compilation errors
3. THE Transpiler SHALL support C-like syntax including function definitions, variable declarations, and function calls
4. THE Transpiler SHALL handle case-insensitive identifiers and convert them to lowercase in generated code
5. THE Transpiler SHALL support implicit type inference for variables (int, str, int[])

### Requirement 2: Asset Embedding

**User Story:** As a developer, I want all assets embedded in the executable, so that I can distribute a single binary without external dependencies.

#### Acceptance Criteria

1. WHEN the Transpiler detects LoadPic() calls, THE Transpiler SHALL identify referenced BMP files and embed them using go:embed
2. WHEN the Transpiler detects PlayMIDI() calls, THE Transpiler SHALL identify referenced MIDI files and embed them using go:embed
3. WHEN the Transpiler detects PlayWAVE() calls, THE Transpiler SHALL identify referenced WAV files and embed them using go:embed
4. THE Transpiler SHALL perform case-insensitive file matching for asset references (Windows 3.1 compatibility)
5. THE Transpiler SHALL generate a single executable containing all embedded assets

### Requirement 3: Virtual Display Architecture

**User Story:** As a user, I want legacy 640x480 content displayed within a modern 1280x720 window, so that I can run old games without scaling artifacts.

#### Acceptance Criteria

1. THE Virtual_Desktop SHALL initialize with a fixed resolution of 1280x720 pixels
2. WHEN OpenWin is called, THE Virtual_Desktop SHALL create a Virtual_Window at the specified position and size
3. WHEN multiple Virtual_Windows are created, THE Virtual_Desktop SHALL render them in creation order (first created = bottom layer)
4. THE Virtual_Desktop SHALL render legacy content at 1:1 pixel ratio to avoid scaling artifacts
5. WHEN a Virtual_Window is closed, THE Virtual_Desktop SHALL remove it from the rendering pipeline

### Requirement 4: Picture Management

**User Story:** As a script author, I want to load and manipulate images, so that I can create visual content for my application.

#### Acceptance Criteria

1. WHEN LoadPic is called with a filename, THE Picture_Manager SHALL load the BMP file and return a unique Picture ID
2. THE Picture_Manager SHALL assign Picture IDs sequentially starting from 0
3. WHEN CreatePic is called, THE Picture_Manager SHALL create an empty image buffer with the specified dimensions
4. WHEN MovePic is called, THE Picture_Manager SHALL copy pixels from source to destination with optional transparency
5. WHEN DelPic is called, THE Picture_Manager SHALL release the Picture resources
6. THE Picture_Manager SHALL support PicWidth and PicHeight queries for any loaded Picture

### Requirement 5: Cast (Sprite) Management

**User Story:** As a script author, I want to display sprites with transparency and layering, so that I can create animated graphics.

#### Acceptance Criteria

1. WHEN PutCast is called, THE Cast_Manager SHALL create a new Cast with the specified source Picture and position
2. THE Cast_Manager SHALL assign Cast IDs sequentially and track creation order for z-ordering
3. WHEN MoveCast is called, THE Cast_Manager SHALL update the Cast position and re-render the destination Picture
4. THE Cast_Manager SHALL support SubImage clipping (srcX, srcY, width, height) for sprite sheets
5. WHEN rendering Casts, THE Cast_Manager SHALL draw them in creation order (first created = bottom, last created = top)
6. THE Cast_Manager SHALL use the top-left pixel color as the transparency key by default
7. WHEN DelCast is called, THE Cast_Manager SHALL remove the Cast and update the rendering order

### Requirement 6: Thread-Safe Rendering

**User Story:** As a system architect, I want thread-safe access to rendering state, so that the Script_Goroutine and Main_Thread do not cause race conditions.

#### Acceptance Criteria

1. WHEN the Script_Goroutine modifies Pictures or Casts, THE Render_Mutex SHALL be acquired before modification
2. WHEN the Main_Thread renders a frame, THE Render_Mutex SHALL be acquired for the duration of Draw()
3. THE Render_Mutex SHALL protect all shared state including Pictures, Casts, and Windows
4. WHEN MoveCast performs rendering, THE Render_Mutex SHALL be held during the entire operation
5. THE Render_Mutex SHALL prevent partial rendering artifacts and data races

### Requirement 7: Double Buffering

**User Story:** As a user, I want flicker-free rendering, so that animations appear smooth without visual artifacts.

#### Acceptance Criteria

1. WHEN MoveCast updates a Picture, THE Renderer SHALL create a temporary off-screen buffer
2. THE Renderer SHALL perform all drawing operations (clear, background, Casts) on the off-screen buffer
3. WHEN drawing is complete, THE Renderer SHALL atomically swap the buffer pointer inside the Render_Mutex
4. THE Renderer SHALL discard the old buffer for garbage collection
5. THE Main_Thread SHALL only display fully-rendered frames to prevent flickering

### Requirement 8: MIDI Playback and Synthesis

**User Story:** As a script author, I want to play MIDI files with software synthesis, so that my application has background music without requiring external MIDI hardware.

#### Acceptance Criteria

1. WHEN PlayMIDI is called with a filename, THE MIDI_Player SHALL load the Standard MIDI File (SMF)
2. THE MIDI_Player SHALL use a SoundFont (.sf2) file for instrument samples
3. THE MIDI_Player SHALL synthesize audio in real-time and output to the system audio device
4. THE MIDI_Player SHALL play the MIDI file once by default (no looping)
5. WHEN MIDI playback advances, THE MIDI_Player SHALL invoke NotifyTick callbacks to update timing state
6. THE MIDI_Player SHALL work natively on macOS using CoreAudio

### Requirement 9: WAVE Audio Playback

**User Story:** As a script author, I want to play WAV files, so that I can add sound effects to my application.

#### Acceptance Criteria

1. WHEN PlayWAVE is called with a filename, THE Audio_Player SHALL decode the WAV file
2. THE Audio_Player SHALL load the entire audio buffer into memory
3. THE Audio_Player SHALL play the audio through the Ebiten audio system
4. THE Audio_Player SHALL support standard WAV formats (PCM, various sample rates)
5. THE Audio_Player SHALL allow multiple WAV files to play concurrently

### Requirement 10: MIDI Sync Mode Timing

**User Story:** As a script author, I want to synchronize animations with MIDI playback, so that visual events align with musical beats.

#### Acceptance Criteria

1. WHEN mes(MIDI_TIME) is called, THE Timing_System SHALL enter MIDI Sync Mode
2. THE Timing_System SHALL drive execution using NotifyTick callbacks from the MIDI_Player
3. WHEN step(n) is called in MIDI Sync Mode, THE Timing_System SHALL interpret n as 32nd note multiples
4. WHEN Wait(n) is called in MIDI Sync Mode, THE Timing_System SHALL wait for n Steps in musical time
5. THE Timing_System SHALL NOT block RegisterSequence in MIDI Sync Mode (must return immediately)
6. WHEN PlayMIDI is called after mes(MIDI_TIME), THE MIDI_Player SHALL start and drive the timing system

### Requirement 11: Time Mode Timing

**User Story:** As a script author, I want frame-based timing independent of MIDI, so that I can create procedural animations.

#### Acceptance Criteria

1. WHEN mes(TIME) is called, THE Timing_System SHALL enter Time Mode
2. THE Timing_System SHALL drive execution using the Main_Thread game loop at 60 FPS
3. WHEN step(n) is called in Time Mode, THE Timing_System SHALL interpret n as milliseconds (n * 50ms)
4. WHEN Wait(n) is called in Time Mode, THE Timing_System SHALL wait for n Steps in frame time
5. THE Timing_System SHALL block RegisterSequence in Time Mode until the sequence completes
6. THE Timing_System SHALL ensure sequential execution (mes block completes before subsequent code)

### Requirement 12: Text Rendering

**User Story:** As a script author, I want to draw text on Pictures, so that I can display messages and UI elements.

#### Acceptance Criteria

1. WHEN SetFont is called, THE Text_Renderer SHALL load the specified font with the given size and attributes
2. WHEN TextWrite is called, THE Text_Renderer SHALL draw the text string on the specified Picture at the given coordinates
3. WHEN TextColor is called, THE Text_Renderer SHALL set the foreground color for subsequent text rendering
4. WHEN BgColor is called, THE Text_Renderer SHALL set the background color for subsequent text rendering
5. WHEN BackMode is called with 1, THE Text_Renderer SHALL render text with transparent background
6. THE Text_Renderer SHALL support Japanese fonts (MS Gothic, MS Mincho) with charset 128

### Requirement 13: String Operations

**User Story:** As a script author, I want to manipulate strings, so that I can process text data in my scripts.

#### Acceptance Criteria

1. WHEN StrLen is called, THE String_Library SHALL return the length of the string
2. WHEN SubStr is called, THE String_Library SHALL extract a substring starting at the specified position with the given length
3. WHEN StrFind is called, THE String_Library SHALL return the index of the first occurrence of the search string
4. WHEN StrPrint is called, THE String_Library SHALL format integers and strings according to the format specifier
5. THE String_Library SHALL support format specifiers: %s (string), %ld (decimal), %lx (hexadecimal)

### Requirement 14: Window Management

**User Story:** As a script author, I want to create and manage virtual windows, so that I can organize visual content.

#### Acceptance Criteria

1. WHEN OpenWin is called with a Picture, THE Window_Manager SHALL create a Virtual_Window displaying that Picture
2. WHEN OpenWin is called with full parameters, THE Window_Manager SHALL create a Virtual_Window with specified position, size, and viewport
3. WHEN MoveWin is called, THE Window_Manager SHALL update the Virtual_Window properties
4. WHEN CloseWin is called, THE Window_Manager SHALL remove the specified Virtual_Window
5. WHEN CloseWinAll is called, THE Window_Manager SHALL remove all Virtual_Windows
6. WHEN CapTitle is called, THE Window_Manager SHALL set the window caption text

### Requirement 15: Message System

**User Story:** As a script author, I want to manage message-driven execution blocks, so that I can control script flow.

#### Acceptance Criteria

1. WHEN GetMesNo is called with 0, THE Message_System SHALL return the last executed mes block number
2. WHEN GetMesNo is called with 1, THE Message_System SHALL return the currently executing mes block number
3. WHEN DelMes is called, THE Message_System SHALL terminate the specified mes block
4. THE Message_System SHALL support mes blocks for MIDI_TIME, TIME, USER, MIDI_END, and other event types
5. THE Message_System SHALL execute mes blocks asynchronously or synchronously based on the message type

### Requirement 16: System Information

**User Story:** As a script author, I want to query system information, so that I can adapt my script to the runtime environment.

#### Acceptance Criteria

1. WHEN WinInfo is called with 0, THE System_Info SHALL return the Virtual_Desktop width (1280)
2. WHEN WinInfo is called with 1, THE System_Info SHALL return the Virtual_Desktop height (720)
3. THE System_Info SHALL support queries for screen dimensions
4. WHEN Random is called with max, THE System_Info SHALL return a random integer in the range [0, max)

### Requirement 17: Script Lifecycle

**User Story:** As a script author, I want to control script execution lifecycle, so that I can terminate or manage script instances.

#### Acceptance Criteria

1. WHEN DelMe is called, THE Script_Runtime SHALL terminate the current script instance
2. WHEN DelUs is called, THE Script_Runtime SHALL terminate all scripts in the current group
3. WHEN DelAll is called, THE Script_Runtime SHALL terminate all active script instances
4. THE Script_Runtime SHALL clean up all resources (Pictures, Casts, Windows) when a script terminates
5. THE Script_Runtime SHALL support graceful shutdown without resource leaks

---

## Part B: Transpiler Language Support Requirements

### Requirement 18: Function Definition Support

**User Story:** As a script author, I want to define custom functions with parameters, so that I can organize my code into reusable components.

#### Acceptance Criteria

1. WHEN a function is defined with typed parameters, THE Transpiler SHALL generate a corresponding Go function
2. THE Transpiler SHALL support parameter types: int, str, and int[] (array)
3. WHEN a function has default parameter values, THE Transpiler SHALL generate Go code that handles optional parameters
4. THE Transpiler SHALL support function calls with positional arguments
5. THE Transpiler SHALL convert all function names to lowercase in generated code

### Requirement 19: Variable Declaration Support

**User Story:** As a script author, I want to declare variables with implicit typing, so that I can store and manipulate data.

#### Acceptance Criteria

1. WHEN a variable is declared without explicit type, THE Transpiler SHALL infer the type from usage
2. THE Transpiler SHALL infer int[] type when a variable is used with array indexing
3. THE Transpiler SHALL infer string type when a variable is assigned a string literal or string function result
4. THE Transpiler SHALL default to int type for all other variables
5. THE Transpiler SHALL support global and local variable scopes

### Requirement 20: Control Flow - Conditional Statements

**User Story:** As a script author, I want to use if-else statements, so that I can implement conditional logic.

#### Acceptance Criteria

1. WHEN an if statement is encountered, THE Transpiler SHALL generate equivalent Go if statement
2. THE Transpiler SHALL support else clauses
3. THE Transpiler SHALL support nested if-else statements
4. THE Transpiler SHALL evaluate boolean expressions in conditions
5. THE Transpiler SHALL support comparison operators (==, !=, <, >, <=, >=)

### Requirement 21: Control Flow - Loop Statements

**User Story:** As a script author, I want to use for, while, and do-while loops, so that I can implement repetitive operations.

#### Acceptance Criteria

1. WHEN a for loop is encountered, THE Transpiler SHALL generate equivalent Go for loop
2. WHEN a while loop is encountered, THE Transpiler SHALL generate equivalent Go for loop with condition
3. WHEN a do-while loop is encountered, THE Transpiler SHALL generate Go code that executes the body at least once
4. THE Transpiler SHALL support break statements to exit loops
5. THE Transpiler SHALL support continue statements to skip to next iteration

### Requirement 22: Control Flow - Switch Statements

**User Story:** As a script author, I want to use switch-case statements, so that I can implement multi-way branching.

#### Acceptance Criteria

1. WHEN a switch statement is encountered, THE Transpiler SHALL generate equivalent Go switch statement
2. THE Transpiler SHALL support multiple case clauses
3. THE Transpiler SHALL support default clause
4. THE Transpiler SHALL support break statements in case clauses
5. THE Transpiler SHALL handle fall-through behavior correctly

### Requirement 23: Drawing Functions - Lines and Shapes

**User Story:** As a script author, I want to draw lines, circles, and rectangles, so that I can create vector graphics.

#### Acceptance Criteria

1. WHEN DrawLine is called, THE Runtime SHALL draw a line between two points on the specified Picture
2. WHEN DrawCircle is called, THE Runtime SHALL draw a circle or ellipse with optional fill modes
3. WHEN DrawRect is called, THE Runtime SHALL draw a rectangle with optional fill modes
4. WHEN SetLineSize is called, THE Runtime SHALL set the line width for subsequent drawing operations
5. WHEN SetPaintColor is called, THE Runtime SHALL set the drawing color for subsequent operations
6. THE Runtime SHALL support fill modes: none, hatch, and solid fill

### Requirement 24: Drawing Functions - Pixel Operations

**User Story:** As a script author, I want to read and manipulate individual pixels, so that I can implement custom graphics effects.

#### Acceptance Criteria

1. WHEN GetColor is called, THE Runtime SHALL return the RGB color value of the specified pixel
2. WHEN SetROP is called, THE Runtime SHALL set the raster operation mode for subsequent drawing
3. THE Runtime SHALL support standard ROP codes (COPYPEN, XORPEN, MERGEPEN, etc.)
4. THE Runtime SHALL apply the selected ROP mode to all drawing operations
5. THE Runtime SHALL preserve pixel data integrity during operations

### Requirement 25: Picture Functions - Scaling

**User Story:** As a script author, I want to scale images, so that I can resize graphics dynamically.

#### Acceptance Criteria

1. WHEN MoveSPic is called, THE Runtime SHALL copy and scale the source region to the destination region
2. THE Runtime SHALL support arbitrary scaling ratios (both upscaling and downscaling)
3. WHEN transparency color is specified, THE Runtime SHALL apply transparent color keying during scaling
4. THE Runtime SHALL use appropriate interpolation for scaled images
5. THE Runtime SHALL preserve aspect ratio when source and destination ratios match

### Requirement 26: Picture Functions - Transformation

**User Story:** As a script author, I want to flip and transform images, so that I can create mirrored graphics.

#### Acceptance Criteria

1. WHEN ReversePic is called, THE Runtime SHALL horizontally flip the source image region
2. THE Runtime SHALL copy the flipped image to the destination Picture
3. THE Runtime SHALL preserve transparency during flipping operations
4. THE Runtime SHALL support arbitrary source and destination regions
5. WHEN GetPicNo is called, THE Runtime SHALL return the Picture ID associated with the specified Window

### Requirement 27: File Operations - INI Files

**User Story:** As a script author, I want to read and write INI configuration files, so that I can persist application settings.

#### Acceptance Criteria

1. WHEN WriteIniInt is called, THE Runtime SHALL write an integer value to the specified INI section and entry
2. WHEN GetIniInt is called, THE Runtime SHALL read an integer value from the specified INI section and entry
3. WHEN WriteIniStr is called, THE Runtime SHALL write a string value to the specified INI section and entry
4. WHEN GetIniStr is called, THE Runtime SHALL read a string value from the specified INI section and entry
5. THE Runtime SHALL create INI files if they do not exist

### Requirement 28: File Operations - File Management

**User Story:** As a script author, I want to manage files and directories, so that I can organize application data.

#### Acceptance Criteria

1. WHEN CopyFile is called, THE Runtime SHALL copy a file from source to destination path
2. WHEN DelFile is called, THE Runtime SHALL delete the specified file
3. WHEN IsExist is called, THE Runtime SHALL return whether the specified file exists
4. WHEN MkDir is called, THE Runtime SHALL create the specified directory
5. WHEN RmDir is called, THE Runtime SHALL remove the specified directory
6. WHEN ChDir is called, THE Runtime SHALL change the current working directory
7. WHEN GetCWD is called, THE Runtime SHALL return the current working directory path

### Requirement 29: File Operations - Binary I/O

**User Story:** As a script author, I want to read and write binary files, so that I can implement custom file formats.

#### Acceptance Criteria

1. WHEN OpenF is called, THE Runtime SHALL open a file and return a file handle
2. WHEN CloseF is called, THE Runtime SHALL close the specified file handle
3. WHEN SeekF is called, THE Runtime SHALL move the file pointer to the specified position
4. WHEN ReadF is called, THE Runtime SHALL read 1-4 bytes and return as an integer
5. WHEN WriteF is called, THE Runtime SHALL write an integer value as 1-4 bytes
6. WHEN StrReadF is called, THE Runtime SHALL read a null-terminated string from the file
7. WHEN StrWriteF is called, THE Runtime SHALL write a null-terminated string to the file

### Requirement 30: String Functions - Advanced Operations

**User Story:** As a script author, I want advanced string manipulation functions, so that I can process text data effectively.

#### Acceptance Criteria

1. WHEN StrInput is called, THE Runtime SHALL display a dialog box and return user input as a string
2. WHEN CharCode is called, THE Runtime SHALL return the character code of the first character
3. WHEN StrUp is called, THE Runtime SHALL convert all lowercase letters to uppercase
4. WHEN StrLow is called, THE Runtime SHALL convert all uppercase letters to lowercase
5. WHEN StrCode is called, THE Runtime SHALL convert a character code to a single-character string

### Requirement 31: Array Operations

**User Story:** As a script author, I want to manipulate dynamic arrays, so that I can manage collections of data.

#### Acceptance Criteria

1. WHEN ArraySize is called, THE Runtime SHALL return the number of elements in the array
2. WHEN DelArrayAll is called, THE Runtime SHALL remove all elements from the array
3. WHEN DelArrayAt is called, THE Runtime SHALL remove the element at the specified index
4. WHEN InsArrayAt is called, THE Runtime SHALL insert an element at the specified index
5. THE Runtime SHALL automatically resize arrays as needed during insertion and deletion

### Requirement 32: Integer Functions - Bit Operations

**User Story:** As a script author, I want to pack and unpack integer values, so that I can efficiently store multiple values.

#### Acceptance Criteria

1. WHEN MakeLong is called, THE Runtime SHALL combine two 16-bit values into a single 32-bit value
2. WHEN GetHiWord is called, THE Runtime SHALL extract the upper 16 bits of a 32-bit value
3. WHEN GetLowWord is called, THE Runtime SHALL extract the lower 16 bits of a 32-bit value
4. THE Runtime SHALL preserve bit patterns during pack and unpack operations
5. THE Runtime SHALL handle signed and unsigned values correctly

### Requirement 33: Multimedia - Video Playback

**User Story:** As a script author, I want to play AVI video files, so that I can display video content.

#### Acceptance Criteria

1. WHEN PlayAVI is called with only a filename, THE Runtime SHALL play the video in a default window
2. WHEN PlayAVI is called with position and size parameters, THE Runtime SHALL play the video in the specified region
3. THE Runtime SHALL support standard AVI codecs
4. THE Runtime SHALL synchronize video playback with audio
5. THE Runtime SHALL generate AVI_START and AVI_END events

### Requirement 34: Multimedia - CD Audio (NOT IMPLEMENTED - Legacy Hardware)

**User Story:** As a script author, I want to play CD audio tracks, so that I can use CD-based music.

**Implementation Status:** NOT IMPLEMENTED - CD audio playback relies on physical CD-ROM drives which are obsolete. Modern systems should use digital audio files (MIDI, WAV) instead.

#### Acceptance Criteria (Reference Only)

1. ~~WHEN PlayCD is called, THE Runtime SHALL play the specified CD audio track~~
2. ~~THE Runtime SHALL support CD audio control commands (play, pause, stop)~~
3. ~~THE Runtime SHALL generate CD_START and CD_END events~~
4. ~~THE Runtime SHALL handle CD drive errors gracefully~~
5. ~~THE Runtime SHALL support multiple CD drives~~

**Note:** Scripts using PlayCD will need to be refactored to use PlayMIDI or PlayWAVE with digital audio files.

### Requirement 35: Multimedia - MCI Commands (NOT IMPLEMENTED - Windows Only)

**User Story:** As a script author, I want to execute MCI commands, so that I can control multimedia devices directly.

**Implementation Status:** NOT IMPLEMENTED - MCI (Media Control Interface) is a Windows-specific API and cannot be supported in a cross-platform implementation targeting macOS and other platforms. This feature is documented for completeness but will not be implemented.

#### Acceptance Criteria (Reference Only)

1. ~~WHEN MCI is called, THE Runtime SHALL execute the MCI command string and return an integer result~~
2. ~~WHEN StrMCI is called, THE Runtime SHALL execute the MCI command string and return a string result~~
3. ~~THE Runtime SHALL support standard MCI command syntax~~
4. ~~THE Runtime SHALL handle MCI errors and return appropriate error codes~~
5. ~~THE Runtime SHALL support device aliases and compound devices~~

**Note:** Scripts using MCI commands will need to be refactored to use cross-platform alternatives (PlayMIDI, PlayWAVE, PlayAVI).

### Requirement 36: Multimedia - Resource Management

**User Story:** As a script author, I want to preload audio resources, so that I can play sounds without loading delays.

#### Acceptance Criteria

1. WHEN LoadRsc is called, THE Runtime SHALL load a WAV file into memory with the specified resource ID
2. WHEN PlayRsc is called, THE Runtime SHALL play the preloaded WAV resource
3. WHEN DelRsc is called, THE Runtime SHALL release the memory used by the resource
4. THE Runtime SHALL support multiple simultaneous resource playbacks
5. THE Runtime SHALL manage memory efficiently for large audio files

### Requirement 37: Message System - Advanced Control

**User Story:** As a script author, I want to pause and resume message handlers, so that I can control script execution flow.

#### Acceptance Criteria

1. WHEN FreezeMes is called, THE Runtime SHALL pause message delivery to the specified mes block
2. WHEN ActivateMes is called, THE Runtime SHALL resume message delivery to the specified mes block
3. THE Runtime SHALL queue messages received while a mes block is frozen
4. WHEN a frozen mes block is activated, THE Runtime SHALL deliver queued messages
5. THE Runtime SHALL support freezing and activating multiple mes blocks independently

### Requirement 38: Message System - Event Generation

**User Story:** As a script author, I want to generate custom messages, so that I can trigger event handlers programmatically.

#### Acceptance Criteria

1. WHEN PostMes is called, THE Runtime SHALL generate a message of the specified type
2. THE Runtime SHALL deliver the message to all matching mes blocks
3. THE Runtime SHALL support message parameters (MesP1, MesP2, MesP3, MesP4)
4. THE Runtime SHALL support all message types (USER, TIME, MIDI_TIME, KEY_DOWN, MOUSEMOVE, etc.)
5. THE Runtime SHALL deliver messages asynchronously without blocking the caller

### Requirement 39: System Integration - Process Execution

**User Story:** As a script author, I want to launch external applications, so that I can integrate with other programs.

#### Acceptance Criteria

1. WHEN Shell is called, THE Runtime SHALL execute the specified application with parameters
2. THE Runtime SHALL set the working directory for the launched process
3. THE Runtime SHALL return immediately without waiting for process completion
4. THE Runtime SHALL handle process launch errors gracefully
5. THE Runtime SHALL support both absolute and relative paths

### Requirement 40: System Integration - System Information

**User Story:** As a script author, I want to query system time and date, so that I can implement time-based features.

#### Acceptance Criteria

1. WHEN GetSysTime is called, THE Runtime SHALL return the current system time as a timestamp
2. WHEN WhatDay is called, THE Runtime SHALL return the current date (year, month, day)
3. WHEN WhatTime is called, THE Runtime SHALL return the current time (hour, minute, second)
4. THE Runtime SHALL use the system's local timezone
5. THE Runtime SHALL handle date/time format conversions correctly

### Requirement 41: System Integration - Registry Access (NOT IMPLEMENTED - Windows Only)

**User Story:** As a script author, I want to read and write Windows registry values, so that I can store persistent configuration.

**Implementation Status:** NOT IMPLEMENTED - Windows Registry is a Windows-specific feature and cannot be supported in a cross-platform implementation. Use INI files (Requirement 27) or other cross-platform configuration storage instead.

#### Acceptance Criteria (Reference Only)

1. ~~WHEN SetRegStr is called, THE Runtime SHALL write a string value to the specified registry key~~
2. ~~WHEN GetRegStr is called, THE Runtime SHALL read a string value from the specified registry key~~
3. ~~THE Runtime SHALL support standard registry hives (HKEY_CURRENT_USER, HKEY_LOCAL_MACHINE, etc.)~~
4. ~~THE Runtime SHALL create registry keys if they do not exist~~
5. ~~THE Runtime SHALL handle registry access errors gracefully~~

**Note:** Scripts using registry functions will need to be refactored to use INI files or JSON configuration files.

### Requirement 42: System Integration - Command Line

**User Story:** As a script author, I want to access command line arguments, so that I can customize application behavior at launch.

#### Acceptance Criteria

1. WHEN GetCmdLine is called, THE Runtime SHALL return the complete command line string
2. THE Runtime SHALL include all arguments passed to the executable
3. THE Runtime SHALL preserve argument quoting and spacing
4. THE Runtime SHALL return an empty string if no arguments were provided
5. THE Runtime SHALL handle special characters in arguments correctly
