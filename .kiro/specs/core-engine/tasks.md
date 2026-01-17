# Implementation Plan: Core Engine

## Overview

This implementation plan covers the son-et core engine, including both the transpiler (compiler pipeline) and runtime library (execution environment). The tasks are organized to build incrementally, with testing integrated throughout.

**Implementation Status Summary:**
- ✅ Core transpiler (lexer, parser, codegen) - Complete
- ✅ Asset embedding system - Complete
- ✅ Graphics system (pictures, windows, casts) - Complete
- ✅ Audio system (MIDI, WAV) - Complete
- ✅ Timing system (TIME and MIDI_TIME modes) - Complete
- ✅ Thread-safe rendering with double buffering - Complete
- ✅ Text rendering (Japanese fonts on macOS) - Complete
- ⚠️ Advanced features (file I/O, drawing functions, etc.) - Pending
- ⚠️ Comprehensive test suite - Pending

## Tasks

- [x] 0. Improve testability through architectural refactoring
  - [x] 0.1 Create Engine state struct to encapsulate global state
    - Define EngineState struct containing all global variables
    - Include: pictures, windows, casts, ID counters, rendering state
    - Add NewEngineState() constructor for clean initialization
    - Add Reset() method for test cleanup
    - _Requirements: All (foundation for testing)_

  - [x] 0.2 Refactor functions to use EngineState receiver
    - Convert package-level functions to methods on EngineState
    - Examples: LoadPic, CreatePic, OpenWin, PutCast, MoveCast
    - Pass EngineState through call chain instead of using globals
    - Maintain backward compatibility with wrapper functions if needed
    - _Requirements: All (foundation for testing)_
    - _Note: AssetLoader interface moved to interpreter-architecture spec (Task 0.1)_

  - [x] 0.4 Separate rendering logic from state management
    - Extract rendering code into Renderer struct
    - Renderer should read from EngineState but not modify it
    - Allow headless testing without Ebitengine initialization
    - Create mock Renderer for unit tests
    - _Requirements: 3.1, 3.2, 3.3, 6.1_

  - [x] 0.5 Create test utilities and helpers
    - Add NewTestEngine() helper for test setup
    - Add assertion helpers for common checks
    - Create fixture data for test images and assets
    - Add helper to verify state consistency
    - _Requirements: All (testing infrastructure)_

  - [x] 0.6 Write baseline tests for refactored code
    - Test EngineState initialization and reset
    - Test basic operations (LoadPic, CreatePic, OpenWin)
    - Test state isolation between test runs
    - Verify no global state leakage
    - _Requirements: 4.1, 4.2, 5.1, 5.2, 14.1_

  - [x] 0.7 Update existing code to use new architecture
    - Migrate all package functions to use EngineState
    - Update Game.Update() and Game.Draw() to use EngineState
    - Ensure backward compatibility for generated code
    - Run existing sample projects to verify no regressions
    - _Requirements: All_

- [x] 0.8 Checkpoint - Verify refactoring maintains functionality
  - Run all existing samples to ensure no regressions
  - Verify test infrastructure is working
  - Confirm state isolation between tests
  - Ask user if questions arise

- [x] 1. Verify and document existing implementation
  - Review current codebase structure
  - Document existing functionality
  - Identify gaps between requirements and implementation
  - _Requirements: 1.1, 1.2, 1.3_
  - **Status: Complete - Core transpiler and runtime implemented and documented**

- [x] 1.1 Write property test for transpiler code generation
  - **Property 1: Transpiler generates valid Go code**
  - **Validates: Requirements 1.1, 1.2**

- [x] 1.2 Write property test for case-insensitive identifiers
  - **Property 2: Case-insensitive identifier transformation**
  - **Validates: Requirements 1.4**

- [x] 2. Enhance asset embedding system
  - [x] 2.1 Implement comprehensive asset detection
    - Scan for all LoadPic, PlayMIDI, PlayWAVE calls
    - Generate go:embed directives for all assets
    - _Requirements: 2.1, 2.2, 2.3_

- [x] 2.2 Write property test for asset embedding
  - **Property 3: Asset embedding completeness**
  - **Validates: Requirements 2.1, 2.2, 2.3**

- [x] 2.3 Write property test for case-insensitive asset matching
  - **Property 4: Case-insensitive asset matching**
  - **Validates: Requirements 2.4**

- [x] 2.4 Implement #include directive support
  - [x] 2.4.1 Implement file inclusion in lexer/parser
    - Parse #include "filename.TFY" directives
    - Recursively read and parse included files
    - Handle relative paths from including file's directory
    - Detect and prevent circular includes
    - Maintain correct line number tracking for error messages
    - _Requirements: 1.1, 1.2_

  - [x] 2.4.2 Write unit tests for #include functionality
    - Test simple file inclusion
    - Test nested includes (A includes B, B includes C)
    - Test circular include detection (A includes B, B includes A)
    - Test file not found errors
    - Test relative path resolution
    - _Requirements: 1.1, 1.2_

- [x] 3. Checkpoint - Ensure transpiler tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [x] 4. Implement control flow statements
  - [x] 4.1 Implement if-else statement parsing and code generation
    - Add lexer tokens for if, else
    - Add parser rules for conditional statements
    - Generate Go if-else code
    - _Requirements: 20.1, 20.2, 20.3, 20.4, 20.5_

- [x] 4.2 Write unit tests for if-else statements
  - Test simple conditions
  - Test nested if-else
  - Test comparison operators
  - _Requirements: 20.1, 20.2, 20.3_

- [x] 4.3 Implement for loop parsing and code generation
    - Add lexer tokens for for loop syntax
    - Add parser rules for for loops
    - Generate Go for loop code
    - _Requirements: 21.1_

- [x] 4.4 Write unit tests for for loops
  - Test basic for loops
  - Test loop with break
  - Test loop with continue
  - _Requirements: 21.1, 21.4, 21.5_

- [x] 4.5 Implement while and do-while loops
    - Add parser rules for while loops
    - Add parser rules for do-while loops
    - Generate appropriate Go code
    - _Requirements: 21.2, 21.3_

- [x] 4.6 Write unit tests for while loops
  - Test while loop execution
  - Test do-while executes at least once
  - _Requirements: 21.2, 21.3_

- [x] 4.7 Implement switch-case statements
    - Add lexer tokens for switch, case, default
    - Add parser rules for switch statements
    - Generate Go switch code with proper fall-through
    - _Requirements: 22.1, 22.2, 22.3, 22.4, 22.5_

- [x] 4.8 Write unit tests for switch-case
  - Test multiple cases
  - Test default clause
  - Test break behavior
  - _Requirements: 22.1, 22.2, 22.3, 22.4_

- [x] 4.9 Implement control flow statements in VM mode (mes blocks)
  - [x] 4.9.1 Add support for if-else statements in genOpCodes
    - Generate OpCode sequences for conditional execution
    - Support nested if-else within mes blocks
    - _Requirements: 20.1, 20.2, 20.3_

  - [x] 4.9.2 Add support for for loops in genOpCodes
    - Generate OpCode sequences for loop initialization, condition, and increment
    - Support break and continue within loops
    - _Requirements: 21.1, 21.4, 21.5_

  - [x] 4.9.3 Add support for while loops in genOpCodes
    - Generate OpCode sequences for while loop execution
    - Support break and continue within loops
    - _Requirements: 21.2_

  - [x] 4.9.4 Add support for switch-case statements in genOpCodes
    - Generate OpCode sequences for switch-case execution
    - Support break statements in cases
    - _Requirements: 22.1, 22.2, 22.3, 22.4_

  - [x] 4.9.5 Update VM executor to handle control flow OpCodes
    - Implement conditional branching in ExecuteOp
    - Implement loop control (break, continue) in VM
    - Support nested control structures
    - _Requirements: 20.1-20.5, 21.1-21.5, 22.1-22.5_

  - [x] 4.9.6 Write unit tests for VM mode control flow
    - Test if-else execution in mes blocks
    - Test loop execution in mes blocks
    - Test switch-case execution in mes blocks
    - _Requirements: 20.1-20.5, 21.1-21.5, 22.1-22.5_

- [x] 5. Checkpoint - Ensure control flow tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [x] 6. Implement drawing functions
  - [x] 6.1 Implement DrawLine function
    - Add runtime function for line drawing
    - Support line width and color
    - _Requirements: 23.1, 23.4, 23.5_

- [x] 6.2 Implement DrawCircle function
    - Add runtime function for circle/ellipse drawing
    - Support fill modes (none, hatch, solid)
    - _Requirements: 23.2, 23.6_

- [x] 6.3 Implement DrawRect function
    - Add runtime function for rectangle drawing
    - Support fill modes
    - _Requirements: 23.3, 23.6_

- [x] 6.4 Write unit tests for drawing functions
  - Test line drawing
  - Test circle drawing with different fill modes
  - Test rectangle drawing
  - _Requirements: 23.1, 23.2, 23.3_

- [x] 6.5 Implement pixel operations
    - Implement GetColor function
    - Implement SetROP function with ROP codes
    - _Requirements: 24.1, 24.2, 24.3, 24.4, 24.5_

- [x] 6.6 Write unit tests for pixel operations
  - Test GetColor returns correct RGB values
  - Test SetROP affects drawing operations
  - _Requirements: 24.1, 24.2_

- [x] 7. Implement picture transformation functions
  - [x] 7.1 Implement MoveSPic (scaling) function
    - Add scaling support with interpolation
    - Support transparency during scaling
    - _Requirements: 25.1, 25.2, 25.3, 25.4, 25.5_

- [x] 7.2 Write property test for image scaling
  - Test scaling preserves aspect ratio when appropriate
  - Test transparency is preserved
  - _Requirements: 25.1, 25.2, 25.3_

- [x] 7.3 Complete ReversePic (flipping) implementation
    - Implement horizontal flip logic
    - Preserve transparency during flip
    - _Requirements: 26.1, 26.2, 26.3, 26.4_

- [x] 7.4 Write unit tests for image flipping
  - Test horizontal flip correctness
  - Test transparency preservation
  - _Requirements: 26.1, 26.2, 26.3_

- [x] 7.5 Implement GetPicNo function
    - Return Picture ID associated with Window
    - _Requirements: 26.5_

- [x] 7.6 Fix array identifier case conversion in transpiler
    - Fix code generation to convert array identifiers to lowercase
    - Ensure LPic[] becomes lpic[], BirdPic[] becomes birdpic[], etc.
    - Test with sample scenarios that use multiple arrays
    - Verify generated code compiles without undefined identifier errors
    - _Requirements: 1.4_

- [x] 8. Checkpoint - Ensure graphics tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [x] 8.1 Optimize transparency handling to use Ebitengine native alpha channel
  - [x] 8.1.1 Add convertTransparentColor helper function
    - Create function to convert specified color to alpha=0 in image
    - Process entire image once during loading/cast creation
    - Return new ebiten.Image with transparency in alpha channel
    - _Requirements: 5.6_

  - [x] 8.1.2 Modify LoadPic to support optional transparency conversion
    - Add optional parameter for transparent color
    - Call convertTransparentColor if transparency specified
    - Store processed image with native alpha channel
    - _Requirements: 4.1, 5.6_

  - [x] 8.1.3 Modify PutCast to pre-process transparency
    - Get transparent color (from args or top-left pixel)
    - Call convertTransparentColor on source image
    - Store processed image in Cast or Picture
    - Remove per-pixel transparency checking from draw operations
    - _Requirements: 5.1, 5.6_

  - [x] 8.1.4 Update MoveCast to use native transparency
    - Remove drawWithColorKey function calls
    - Use standard ebiten.DrawImage with pre-processed images
    - Leverage Ebitengine's native alpha blending
    - Maintain double buffering for flicker-free rendering
    - _Requirements: 5.3, 5.5, 5.6_

  - [x] 8.1.5 Test transparency optimization with verified samples
    - Test with existing sample projects (sprite animations)
    - Test sprite movements and clipping
    - Verify transparency still works correctly
    - Verify no visual regressions
    - _Requirements: 5.6_

  - [x] 8.1.6 Update architecture documentation
    - Update Transparency Handling section in architecture.md
    - Document new approach: pre-process vs per-draw checking
    - Explain performance benefits of native alpha channel
    - Update code examples to reflect new implementation
    - _Requirements: 5.6_

- [x] 8.2 Fix transpiler array identifier case conversion
  - [x] 8.2.1 Analyze array identifier handling in codegen
    - Review how array identifiers are processed in genExpression
    - Identify why LPic[], BirdPic[], OPPic[] are not converted to lowercase
    - Check if issue is in identifier lookup or code generation
    - _Requirements: 1.4_

  - [x] 8.2.2 Fix array identifier case conversion
    - Ensure all array identifiers are converted to lowercase
    - Update identifier processing to handle array subscript expressions
    - Maintain consistency with other identifier conversions
    - _Requirements: 1.4_

  - [x] 8.2.3 Test with sample scenarios
    - Transpile sample scenarios that use arrays
    - Verify generated code compiles without undefined identifier errors
    - Build and run executables
    - Confirm ReversePic and other functions work correctly
    - _Requirements: 1.4, 26.1, 26.2, 26.3_

- [x] 9. Implement file operations
  - [x] 9.1 Implement INI file operations
    - Implement WriteIniInt, GetIniInt
    - Implement WriteIniStr, GetIniStr
    - Create INI files if they don't exist
    - _Requirements: 27.1, 27.2, 27.3, 27.4, 27.5_

- [x] 9.2 Write unit tests for INI operations
  - Test reading and writing integers
  - Test reading and writing strings
  - Test file creation
  - _Requirements: 27.1, 27.2, 27.3, 27.4_

- [x] 9.3 Implement file management functions
    - Implement CopyFile, DelFile, IsExist
    - Implement MkDir, RmDir, ChDir, GetCWD
    - _Requirements: 28.1, 28.2, 28.3, 28.4, 28.5, 28.6, 28.7_

- [x] 9.4 Write unit tests for file management
  - Test file copy and delete
  - Test directory operations
  - Test path operations
  - _Requirements: 28.1, 28.2, 28.3, 28.4, 28.5_

- [x] 9.5 Implement binary file I/O
    - Implement OpenF, CloseF, SeekF
    - Implement ReadF, WriteF
    - Implement StrReadF, StrWriteF
    - _Requirements: 29.1, 29.2, 29.3, 29.4, 29.5, 29.6, 29.7_

- [x] 9.6 Write unit tests for binary I/O
  - Test file open/close
  - Test read/write operations
  - Test string I/O
  - _Requirements: 29.1, 29.2, 29.3, 29.4, 29.5, 29.6, 29.7_

- [x] 10. Implement advanced string functions
  - [x] 10.1 Implement StrInput dialog function
    - Display dialog box for user input
    - Return user-entered string
    - _Requirements: 30.1_

- [x] 10.2 Implement string case conversion
    - Implement StrUp (to uppercase)
    - Implement StrLow (to lowercase)
    - _Requirements: 30.3, 30.4_

- [x] 10.3 Implement character code functions
    - Implement CharCode (get character code)
    - Complete StrCode implementation (code to character)
    - _Requirements: 30.2, 30.5_

- [x] 10.4 Write property tests for string operations
  - **Property 16: String operation correctness**
  - **Validates: Requirements 13.1, 13.2, 13.3**

- [x] 10.5 Write property test for string formatting
  - **Property 17: String formatting correctness**
  - **Validates: Requirements 13.4, 13.5**

- [x] 11. Implement array operations
  - [x] 11.1 Implement array size and clear functions
    - Implement ArraySize
    - Implement DelArrayAll
    - _Requirements: 31.1, 31.2_

- [x] 11.2 Implement array element operations
    - Implement DelArrayAt
    - Implement InsArrayAt
    - Support automatic resizing
    - _Requirements: 31.3, 31.4, 31.5_

- [x] 11.3 Write unit tests for array operations
  - Test array size queries
  - Test element insertion and deletion
  - Test automatic resizing
  - _Requirements: 31.1, 31.2, 31.3, 31.4_

- [x] 12. Implement integer bit operations
  - [x] 12.1 Implement bit packing functions
    - Implement MakeLong
    - Implement GetHiWord
    - Implement GetLowWord
    - _Requirements: 32.1, 32.2, 32.3, 32.4, 32.5_

- [x] 12.2 Write unit tests for bit operations
  - Test packing and unpacking
  - Test bit pattern preservation
  - Test signed/unsigned handling
  - _Requirements: 32.1, 32.2, 32.3, 32.4_

- [x] 13. Checkpoint - Ensure utility function tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [x] 14. Implement Variable Scope & VM Architecture (Phase 2)
  - [x] 14.1 Analyze variable usage in mes() blocks
    - Scan all mes() blocks in a function during code generation
    - Collect all variables referenced inside mes() blocks
    - Mark these variables as "needs VM registration"
    - Store marked variables for use during code generation
    - _Requirements: 1.1, 1.2, 1.4_

  - [x] 14.2 Modify transpiler to generate engine.Assign() calls
    - For variables marked as "needs VM registration", generate: `varname = engine.Assign("varname", value).(type)`
    - For unmarked variables, generate normal: `varname = value`
    - Maintain type safety with appropriate type assertions (.(int), .(string), etc.)
    - Ensure case-insensitive variable names (lowercase in Assign calls)
    - _Requirements: 1.1, 1.2, 1.4_

  - [x] 14.3 Update collectVariablesInBlock to detect all variable references
    - Enhance variable collection to handle nested expressions
    - Detect variables in infix expressions (e.g., winW-320)
    - Detect variables in function call arguments
    - Detect variables in array subscripts
    - _Requirements: 1.1, 1.2, 1.4_

  - [x] 14.4 Test with sample scenario
    - Transpile a sample scenario with new variable registration
    - Verify generated code uses engine.Assign() for variables used in mes() blocks
    - Build and run the executable
    - Verify windows appear at correct positions (not off-screen)
    - Verify all visual elements render correctly
    - _Requirements: 1.1, 1.2, 1.4_

  - [x] 14.5 Write unit tests for variable scope
    - Test that variables defined outside mes() blocks are accessible inside
    - Test that variables defined inside mes() blocks are local to that block
    - Test case-insensitive variable lookup (winW, winw, WINW all refer to same variable)
    - Test parent scope chain lookup (nested mes() blocks)
    - _Requirements: 1.1, 1.2, 1.4_

  - [x] 14.6 Update documentation
    - Update design.md with Phase 2 completion status
    - Document the Assign() helper function usage
    - Add examples of generated code with variable registration
    - Note any limitations or edge cases discovered
    - _Requirements: 1.1, 1.2, 1.4_

- [x] 15. Checkpoint - Verify variable scope implementation
  - [x] 15.1 Run all existing unit tests
    - Execute all test suites to ensure no regressions
    - Verify all property-based tests pass
    - Check test coverage for variable scope functionality
    - _Requirements: 1.1, 1.2, 1.4_

  - [x] 15.2 Test with sample scenarios using variable scope
    - Transpile sample scenarios that use variables in mes() blocks
    - Verify generated code includes proper engine.Assign() calls
    - Build executables and verify they compile without errors
    - _Requirements: 1.1, 1.2, 1.4_

  - [x] 15.3 Verify window positioning correctness
    - Run sample scenarios and verify windows appear at correct positions
    - Test scenarios with calculated positions (e.g., winW-320, winH/2)
    - Ensure no windows appear off-screen due to variable scope issues
    - _Requirements: 1.1, 1.2, 1.4_

  - [x] 15.4 Verify visual elements render correctly
    - Test sprite positioning and movement
    - Verify cast operations work with scoped variables
    - Check that all visual elements appear as expected
    - _Requirements: 1.1, 1.2, 1.4_

  - [x] 15.5 Run regression tests on other samples
    - Test samples that don't use variable scope features
    - Verify no regressions in basic functionality
    - Ensure backward compatibility is maintained
    - _Requirements: All_

  - [x] 15.6 Ask user for verification if issues arise
    - Report any test failures or unexpected behavior
    - Request user guidance on any ambiguous results
    - Document any limitations or edge cases discovered
    - _Requirements: 1.1, 1.2, 1.4_

- [ ] 16. Implement advanced message system
  - [ ] 16.1 Implement message pause/resume
    - Implement FreezeMes
    - Implement ActivateMes
    - Queue messages while frozen
    - _Requirements: 37.1, 37.2, 37.3, 37.4, 37.5_

  - [ ] 16.2 Write unit tests for message control
    - Test message freezing
    - Test message queueing
    - Test message activation
    - _Requirements: 37.1, 37.2, 37.3, 37.4_

  - [ ] 16.3 Complete PostMes implementation
    - Support all message types
    - Support message parameters (MesP1-MesP4)
    - Deliver messages asynchronously
    - _Requirements: 38.1, 38.2, 38.3, 38.4, 38.5_

  - [ ] 16.4 Write unit tests for message generation
    - Test message delivery
    - Test message parameters
    - Test async delivery
    - _Requirements: 38.1, 38.2, 38.3, 38.4_

- [ ] 17. Implement system integration functions
  - [ ] 17.1 Implement Shell function
    - Execute external applications
    - Set working directory
    - Handle errors gracefully
    - _Requirements: 39.1, 39.2, 39.3, 39.4, 39.5_

  - [ ] 17.2 Write unit tests for Shell function
    - Test process execution
    - Test working directory
    - Test error handling
    - _Requirements: 39.1, 39.2, 39.4_

  - [ ] 17.3 Implement system time functions
    - Implement GetSysTime
    - Implement WhatDay
    - Implement WhatTime
    - _Requirements: 40.1, 40.2, 40.3, 40.4, 40.5_

  - [ ] 17.4 Write unit tests for time functions
    - Test time retrieval
    - Test date retrieval
    - Test timezone handling
    - _Requirements: 40.1, 40.2, 40.3, 40.4_

  - [ ] 17.5 Implement GetCmdLine function
    - Return complete command line string
    - Preserve argument quoting
    - Handle special characters
    - _Requirements: 42.1, 42.2, 42.3, 42.4, 42.5_

  - [ ] 17.6 Write unit tests for command line access
    - Test argument retrieval
    - Test quoting preservation
    - Test special characters
    - _Requirements: 42.1, 42.2, 42.3, 42.5_

- [ ] 18. Checkpoint - Ensure all new features pass tests
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 19. Write comprehensive property-based tests
  - [ ] 19.1 Write property test for sequential ID assignment
    - **Property 5: Sequential ID assignment**
    - **Validates: Requirements 4.2, 5.2**

  - [ ] 19.2 Write property test for window creation
    - **Property 6: Window creation and properties**
    - **Validates: Requirements 3.2, 14.1, 14.2**

  - [ ] 19.3 Write property test for creation order rendering
    - **Property 7: Creation order rendering**
    - **Validates: Requirements 3.3, 5.5**

  - [ ] 19.4 Write property test for resource cleanup
    - **Property 8: Resource cleanup on deletion**
    - **Validates: Requirements 4.5, 5.7, 14.4**

  - [ ] 19.5 Write property test for picture dimensions
    - **Property 9: Picture dimension queries**
    - **Validates: Requirements 4.6**

  - [ ] 19.6 Write property test for sprite clipping
    - **Property 10: Sprite clipping correctness**
    - **Validates: Requirements 5.4**

  - [ ] 19.7 Write property test for transparency
    - **Property 11: Transparency key behavior**
    - **Validates: Requirements 5.6**

  - [ ] 19.8 Write property test for MIDI sync non-blocking
    - **Property 12: MIDI Sync Mode non-blocking**
    - **Validates: Requirements 10.5**

- [ ] 18.9 Write property test for Time mode blocking
    - **Property 13: Time Mode blocking**
    - **Validates: Requirements 11.5, 11.6**

- [ ] 18.10 Write property test for MIDI step interpretation
    - **Property 14: Step interpretation in MIDI Sync Mode**
    - **Validates: Requirements 10.3**

- [ ] 18.11 Write property test for Time step interpretation
    - **Property 15: Step interpretation in Time Mode**
    - **Validates: Requirements 11.3**

- [ ] 18.12 Write property test for random number range
    - **Property 18: Random number range**
    - **Validates: Requirements 16.4**

- [ ] 18.13 Write property test for message system state
    - **Property 19: Message system state tracking**
    - **Validates: Requirements 15.1, 15.2**

- [ ] 18.14 Write property test for script termination
    - **Property 20: Script termination cleanup**
    - **Validates: Requirements 17.1, 17.2, 17.3, 17.4**

- [ ] 18.15 Write property test for text rendering state
    - **Property 21: Text rendering state persistence**
    - **Validates: Requirements 12.3, 12.4, 12.5**

- [ ] 18.16 Write property test for window updates
    - **Property 22: Window state updates**
    - **Validates: Requirements 14.3**

- [ ] 18.17 Write property test for window captions
    - **Property 23: Window caption updates**
    - **Validates: Requirements 14.6**

- [ ] 18.18 Write property test for concurrent WAV playback
    - **Property 24: Multiple WAV concurrent playback**
    - **Validates: Requirements 9.5**

- [ ] 18.19 Write property test for MIDI single iteration
    - **Property 25: MIDI playback single iteration**
    - **Validates: Requirements 8.4**

- [ ] 19. Integration testing and validation
  - [ ] 19.1 Test with existing sample projects
    - Run transpiler on existing sample projects
    - Verify executables run correctly
    - _Requirements: All_

- [ ] 19.2 Verify thread safety with race detector
    - Run all tests with -race flag
    - Fix any detected race conditions
    - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5_

- [ ] 19.3 Performance benchmarking
    - Benchmark transpilation speed
    - Benchmark rendering performance
    - Benchmark audio synthesis
    - Profile and optimize hot paths

- [ ] 19.4 Cross-platform testing
    - Test on macOS (primary platform)
    - Verify CoreAudio integration
    - Test asset embedding in built executables

- [ ] 20. Implement multimedia functions (Future Implementation)
  - [ ] 20.1 Implement AVI video playback
    - Add AVI decoder support
    - Implement PlayAVI with position/size parameters
    - Generate AVI_START and AVI_END events
    - _Requirements: 33.1, 33.2, 33.3, 33.4, 33.5_

  - [ ] 20.2 Write unit tests for AVI playback
    - Test video loading
    - Test playback events
    - _Requirements: 33.1, 33.5_

  - [ ] 20.3 Implement audio resource management
    - Implement LoadRsc (preload WAV)
    - Implement PlayRsc (play preloaded)
    - Implement DelRsc (release resource)
    - _Requirements: 36.1, 36.2, 36.3, 36.4, 36.5_

  - [ ] 20.4 Write unit tests for resource management
    - Test resource loading and playback
    - Test resource cleanup
    - Test concurrent playback
    - _Requirements: 36.1, 36.2, 36.3, 36.4_

- [ ] 21. Final checkpoint - Complete system validation
  - Ensure all tests pass, ask the user if questions arise.
  - Verify all requirements are implemented
  - Review code coverage (aim for >80% on critical paths)

- [-] 22. Implement headless mode and auto-termination
  - [x] 22.1 Add command line flag parsing
    - Add --headless flag for GUI-less execution
    - Add --timeout flag for auto-termination (e.g., "5s", "500ms", "2m")
    - Add HEADLESS environment variable support
    - _Requirements: 43.1, 44.1, 44.2_

  - [ ] 22.2 Implement headless execution mode
    - Create runHeadless() function that simulates game loop without Ebiten
    - Implement 60 FPS ticker for VM updates
    - Stub out rendering operations (OpenWin, PutCast, MoveCast) with logging
    - Ensure script logic executes normally (timing, audio, state changes)
    - _Requirements: 43.1, 43.2, 43.3, 43.4, 43.5_

  - [ ] 22.3 Implement timeout management
    - Parse timeout duration from command line flag
    - Create timeout channel with time.After()
    - Integrate timeout check into game loop (both GUI and headless)
    - Log timeout message when auto-termination occurs
    - _Requirements: 44.1, 44.2, 44.3, 44.7_

  - [ ] 22.4 Implement Exit() function
    - Add Exit() runtime function for immediate termination
    - Cleanup resources (audio, windows, files) before exit
    - Exit with status code 0
    - Override timeout when Exit() is called
    - _Requirements: 44.6_

  - [ ] 22.5 Implement graceful shutdown
    - Ensure all resources are cleaned up on timeout
    - Close audio players, file handles, windows
    - Exit with status code 0 for normal termination
    - _Requirements: 44.3, 44.4_

  - [ ] 22.6 Write unit tests for headless mode
    - Test headless execution without Ebiten initialization
    - Test rendering operation stubs (verify logging, no crashes)
    - Test script logic execution in headless mode
    - Verify timing system works in headless mode
    - _Requirements: 43.1, 43.2, 43.3, 43.4_

  - [ ] 22.7 Write unit tests for auto-termination
    - Test timeout parsing (various formats: "5s", "500ms", "2m")
    - Test timeout triggers termination
    - Test Exit() overrides timeout
    - Test graceful resource cleanup on timeout
    - _Requirements: 44.1, 44.2, 44.3, 44.4, 44.6_

  - [ ] 22.8 Write property test for headless execution equivalence
    - **Property 22: Headless mode execution equivalence**
    - **Validates: Requirements 43.3, 43.4**

  - [ ] 22.9 Write property test for timeout termination
    - **Property 23: Timeout termination**
    - **Validates: Requirements 44.1, 44.3, 44.4**

  - [ ] 22.10 Write property test for Exit immediate termination
    - **Property 24: Exit immediate termination**
    - **Validates: Requirements 44.6**

  - [ ] 22.11 Integration test with sample scenarios
    - Test sample scenarios in headless mode with timeout
    - Verify logs are captured correctly
    - Verify no orphaned processes remain
    - Test both direct mode and embedded mode
    - _Requirements: 43.1, 43.5, 43.6, 44.5_

  - [ ] 22.12 Update documentation
    - Update README.md with headless mode usage examples
    - Update build-workflow.md with new execution commands
    - Document timeout format options
    - Add troubleshooting section for headless mode
    - _Requirements: 43.1, 44.1_

- [ ] 23. Checkpoint - Verify headless mode and auto-termination
  - Run all tests to ensure no regressions
  - Test with sample scenarios in both GUI and headless modes
  - Verify timeout works correctly in both modes
  - Ensure no orphaned processes after timeout
  - Ask user for verification if issues arise

- [ ] 24. Fix Picture and Cast rendering issues (y_saru sample bugs)
  - [ ] 24.1 Investigate and fix CreatePic ID assignment issue
    - Analyze why CreatePic(sourcePicID) returns unexpected IDs
    - Review nextPicID counter initialization and increment logic
    - Verify that LoadPic increments nextPicID correctly
    - Test with y_saru sample: verify base_pic = CreatePic(25) returns correct ID
    - Ensure OpenWin(base_pic, ...) displays the correct picture
    - _Requirements: 4.2, 4.3_
    - _Bug: P14の後にP0が表示される (P0 appears after P14)_

  - [ ] 24.2 Investigate and fix Cast visibility issue
    - Analyze why PutCast creates casts but they don't appear on screen
    - Verify Cast is drawn to correct destination picture
    - Check if destination picture exists and is valid
    - Verify MoveCast updates cast position and re-renders
    - Test with y_saru sample シーン3: verify flying plane cast appears
    - Test with y_saru sample シーン5.5: verify monkey cast appears
    - Test with y_saru sample シーン7: verify monkey cast appears
    - _Requirements: 5.1, 5.2, 5.3, 5.5_
    - _Bug: Castが表示されない (Casts are not visible)_

  - [ ] 24.3 Add Window ID debug display
    - When debugLevel >= 2, draw Window ID on each window
    - Display format: "W{windowID}" in yellow color at top-left corner
    - Similar to existing Picture ID display ("P{picID}")
    - Use existing font (e.currentFont) if available
    - Position: offset by 5 pixels from top-left to avoid overlap
    - _Requirements: 14.1, 14.2_
    - _Enhancement: ウィンドウにもIDを表示 (Display Window IDs for debugging)_

  - [ ] 24.4 Write unit tests for CreatePic ID assignment
    - Test CreatePic(width, height) returns sequential IDs
    - Test CreatePic(sourcePicID) returns new sequential ID (not source ID)
    - Test that LoadPic and CreatePic share the same ID counter
    - Verify nextPicID increments correctly across mixed LoadPic/CreatePic calls
    - _Requirements: 4.2, 4.3_

  - [ ] 24.5 Write unit tests for Cast rendering
    - Test PutCast creates cast and draws to destination picture
    - Test MoveCast updates cast position and re-renders
    - Test cast visibility with valid destination picture
    - Test cast rendering order (z-ordering)
    - Verify cast appears on destination picture after PutCast
    - _Requirements: 5.1, 5.2, 5.3, 5.5_

  - [ ] 24.6 Integration test with y_saru sample
    - Run y_saru sample with DEBUG_LEVEL=2
    - Verify Window IDs are displayed on each window
    - Verify Picture IDs are displayed on each picture
    - Verify Casts are visible in シーン3, シーン5.5, シーン7
    - Verify correct picture sequence (no unexpected P0 after P14)
    - Capture logs and verify no errors
    - _Requirements: 4.2, 4.3, 5.1, 5.2, 5.3, 5.5, 14.1, 14.2_

- [ ] 25. Checkpoint - Verify y_saru sample bug fixes
  - Run y_saru sample and verify all three issues are resolved
  - Verify Window IDs appear in debug mode
  - Verify Casts are visible in all scenes
  - Verify correct picture display sequence
  - Run all existing tests to ensure no regressions
  - Ask user for verification if issues arise

- [ ] 26. Fix TIME mode execution issues (kuma2 sample bugs)
  - [ ] 26.1 Investigate and fix mes(TIME) block looping issue
    - Analyze why mes(TIME) block executes repeatedly instead of once
    - Review RegisterSequence blocking behavior in TIME mode
    - Verify sequence completion detection and cleanup
    - Check if sequence is being re-registered after completion
    - Ensure mes(TIME) block completes and allows subsequent code to run
    - Test with kuma2: verify mes(TIME) executes once, then del_all/del_me run
    - _Requirements: 11.5, 11.6_
    - _Bug: mes(TIME)ブロックが繰り返し実行される (mes(TIME) block loops instead of executing once)_

  - [ ] 26.2 Investigate and fix step() timing in TIME mode
    - Analyze step(n) interpretation in TIME mode
    - Current implementation: step(n) = n * 50ms per comma
    - Verify against sabo2 sample: step(2) should sync with MIDI duration
    - Check if step(n) means "n ticks per comma" vs "n * 50ms per comma"
    - Review original FILLY specification for step() semantics
    - Test with multiple samples: kuma2 (step 65), sabo2 (step 2)
    - Ensure audio and animation stay synchronized
    - _Requirements: 11.3, 11.4_
    - _Bug: step(n)の解釈が間違っている - 音楽とアニメーションが同期しない_

  - [ ] 26.3 Add detailed logging for TIME mode execution
    - Log when mes(TIME) block starts and completes
    - Log RegisterSequence blocking/unblocking in TIME mode
    - Log sequence completion and cleanup
    - Log step(n) value when entering step block
    - Log calculated wait duration for each comma
    - Log actual elapsed time between operations
    - Help diagnose timing and looping issues
    - _Requirements: 11.3, 11.5, 11.6_

  - [ ] 26.4 Write unit tests for mes(TIME) execution
    - Test mes(TIME) block executes exactly once
    - Test RegisterSequence blocks until sequence completes in TIME mode
    - Test subsequent code runs after mes(TIME) completes
    - Test del_all/del_me execute after mes(TIME) block
    - Verify no sequence re-registration after completion
    - _Requirements: 11.5, 11.6_

  - [ ] 26.5 Write unit tests for step timing
    - Test step(n) calculates correct wait duration in TIME mode
    - Test comma expansion: `,` = 1 step, `,,` = 2 steps, etc.
    - Test various step values: step(2), step(65), step(100)
    - Verify total execution time matches expected duration
    - Test audio/animation synchronization with different step values
    - _Requirements: 11.3, 11.4_

  - [ ] 26.6 Integration test with multiple samples
    - Test kuma2 sample (step 65): verify timing and synchronization
    - Test sabo2 sample (step 2): verify audio matches animation duration
    - Measure total execution time for each sample
    - Verify mes(TIME) executes once and completes
    - Verify del_all/del_me execute after mes(TIME) completes
    - Verify program terminates cleanly
    - Capture logs and verify execution flow
    - _Requirements: 11.3, 11.4, 11.5, 11.6_

- [ ] 27. Checkpoint - Verify kuma2 and sabo2 sample bug fixes
  - Run kuma2 sample and verify issues are resolved
  - Run sabo2 sample and verify audio/animation synchronization
  - Verify mes(TIME) executes once without looping
  - Verify step(n) timing is correct for different values
  - Verify program terminates cleanly after completion
  - Run all existing tests to ensure no regressions
  - Ask user for verification if issues arise

- [ ] 28. Fix yosemiya sample issues (multiple mes blocks and missing functions)
  - [ ] 28.1 Implement missing functions
    - Implement CloseWinAll() function (currently shows "Unknown function closewinall")
    - Verify CapTitle() function mapping (currently shows "Unknown function captitle")
    - Check function name case conversion in interpreter
    - Test with yosemiya sample to verify functions are called correctly
    - _Requirements: 14.5, 14.6_
    - _Bug: CloseWinAllとCapTitleが未実装または関数名マッピングが間違っている_

  - [ ] 28.2 Investigate multiple mes(TIME) blocks execution
    - Analyze why second mes(TIME) block completes immediately
    - Review how multiple mes() blocks are registered and executed
    - Check if mes() blocks should run in parallel or sequentially
    - Verify variable scope between different mes() blocks
    - Test with yosemiya: verify second mes(TIME) block executes animations
    - _Requirements: 11.5, 11.6, 15.4, 15.5_
    - _Bug: 2つ目のmes(TIME)ブロックが即座に終了する (Second mes(TIME) block finishes immediately)_

  - [ ] 28.3 Fix variable initialization and scope in mes() blocks
    - Analyze why variable 'i' is not initialized in second mes(TIME) block
    - Review variable scope between different mes() blocks
    - Ensure variables declared in main() are accessible in all mes() blocks
    - Check if variables need explicit initialization before mes() blocks
    - Test with yosemiya: verify 'i', 'j', 'k' variables work correctly
    - _Requirements: 19.1, 19.4, 19.5_
    - _Bug: mes()ブロック間で変数が共有されていない (Variables not shared between mes() blocks)_

  - [ ] 28.4 Add detailed logging for multiple mes() blocks
    - Log when each mes() block is registered
    - Log when each mes() block starts and completes
    - Log variable values at the start of each mes() block
    - Log if/for condition evaluations in mes() blocks
    - Help diagnose why second mes() block doesn't execute properly
    - _Requirements: 11.5, 11.6, 15.4_

  - [ ] 28.5 Write unit tests for multiple mes() blocks
    - Test multiple mes(TIME) blocks in same script
    - Test variable sharing between mes() blocks
    - Test if/for statements inside mes() blocks
    - Test mes() blocks with different timing modes
    - Verify all mes() blocks execute as expected
    - _Requirements: 11.5, 11.6, 15.4, 15.5_

  - [ ] 28.6 Integration test with yosemiya sample
    - Run yosemiya sample and verify both mes(TIME) blocks execute
    - Verify curtain opening animation (second mes block)
    - Verify message display animation (second mes block)
    - Verify CloseWinAll() and CapTitle() work correctly
    - Capture logs and verify execution flow
    - _Requirements: 11.5, 11.6, 14.5, 14.6, 15.4, 15.5_

- [ ] 29. Checkpoint - Verify yosemiya sample bug fixes
  - Run yosemiya sample and verify all issues are resolved
  - Verify both mes(TIME) blocks execute properly
  - Verify animations run at correct speed
  - Verify CloseWinAll() and CapTitle() functions work
  - Run all existing tests to ensure no regressions
  - Ask user for verification if issues arise

- [ ] 30. Fix robot sample infinite loop issue (for loop not terminating)
  - [x] 30.1 Investigate infinite LoadPic loop
    - Analyze why `for(i=0;i<=1;i=i+1)` loops infinitely
    - Review for loop code generation in transpiler
    - Check if loop variable 'i' is being incremented correctly
    - Check if loop condition `i<=1` is being evaluated correctly
    - Verify loop initialization `i=0` executes before loop starts
    - Test with robot sample: verify loop loads only ROBOT000.BMP and ROBOT001.BMP
    - _Requirements: 21.1_
    - _Bug: for(i=0;i<=1;i=i+1)が無限ループになり、ROBOT000.BMPを67,081回ロードする_

  - [x] 30.2 Analyze loop variable scope and initialization
    - Check if loop variable 'i' is properly scoped to the loop
    - Verify loop variable is not being reset on each iteration
    - Check if loop variable conflicts with other variables named 'i'
    - Review variable registration for loop variables in mes() blocks
    - Test with simple for loop examples to isolate the issue
    - _Requirements: 21.1, 19.1_

  - [x] 30.3 Add detailed logging for for loop execution
    - Log loop initialization: variable name and initial value
    - Log loop condition evaluation: variable value and condition result
    - Log loop increment: variable value before and after increment
    - Log loop iteration count to detect infinite loops
    - Help diagnose why loop doesn't terminate
    - _Requirements: 21.1_

  - [x] 30.4 Write unit tests for for loop execution
    - Test simple for loop: `for(i=0;i<5;i=i+1)`
    - Test for loop with <= condition: `for(i=0;i<=1;i=i+1)`
    - Test for loop with different increments: `i=i+2`, `i=i-1`
    - Test nested for loops
    - Verify loop terminates after correct number of iterations
    - Verify loop variable has correct value after loop
    - _Requirements: 21.1_

  - [x] 30.5 Write property test for for loop termination
    - **Property 26: For loop termination**
    - Generate random loop bounds and increments
    - Verify loop terminates within expected iterations
    - Verify loop variable has correct final value
    - **Validates: Requirements 21.1**

  - [x] 30.6 Integration test with robot sample
    - Run robot sample with timeout (3-5 seconds)
    - Verify loop loads exactly 2 images: ROBOT000.BMP and ROBOT001.BMP
    - Verify loop terminates and script continues
    - Verify no infinite loop (check log line count < 1000)
    - Capture logs and verify loop execution flow
    - _Requirements: 21.1_

- [ ] 31. Checkpoint - Verify robot sample bug fix
  - Run robot sample and verify infinite loop is fixed
  - Verify for loop terminates correctly
  - Verify only 2 images are loaded (ROBOT000.BMP, ROBOT001.BMP)
  - Verify script continues after loop completes
  - Run all existing tests to ensure no regressions
  - Ask user for verification if issues arise

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- Each task references specific requirements for traceability
- Checkpoints ensure incremental validation
- Property tests validate universal correctness properties
- Unit tests validate specific examples and edge cases
- Integration tests verify end-to-end functionality
