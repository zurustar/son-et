# Implementation Plan: Sample Game Fixes

## Overview

This implementation plan addresses four critical bugs in the son-et game engine by systematically investigating each issue, implementing targeted fixes, and verifying correct behavior. The approach follows a test-driven methodology with property-based testing to ensure comprehensive coverage.

## Tasks

- [x] 1. Investigate and fix sab2 termination issue
  - [x] 1.1 Run sab2 sample and capture parse errors
    - Execute `./son-et --headless --timeout=10s --debug=2 samples/sab2 2>&1 > sab2_errors.log`
    - Identify all unsupported function calls causing parse failures
    - _Requirements: 5.1, 5.2, 5.5_
  
  - [x] 1.2 Add stub implementations for legacy functions
    - Add Shell() stub to pkg/engine/vm.go in executeCall function
    - Add GetIniStr() stub to pkg/engine/vm.go
    - Add MCI() stub to pkg/engine/vm.go
    - Add StrMCI() stub to pkg/engine/vm.go
    - Each stub should log a warning and return safe default values
    - _Requirements: 1.1, 1.2_
  
  - [x] 1.3 Write unit tests for legacy function stubs
    - Test that Shell() parses and returns without error
    - Test that GetIniStr() returns empty string default
    - Test that MCI() and StrMCI() log warnings
    - _Requirements: 1.1_
  
  - [x] 1.4 Verify sab2 parses and terminates correctly
    - Run sab2 sample again with stubs in place
    - Verify no parse errors occur
    - Verify game terminates with exit code 0
    - _Requirements: 6.2, 1.5_

- [x] 2. Investigate and fix y_saru P14 progression issue
  - [x] 2.1 Analyze y_saru TFY script structure
    - Read samples/y_saru/Y-SARU.TFY and identify mes(MIDI_TIME) block structure
    - Locate step P14 in the script
    - Document what operations occur at P14
    - _Requirements: 5.1, 5.2, 5.3_
  
  - [x] 2.2 Run y_saru with detailed logging
    - Execute with `--debug=2` to capture sequencer state
    - Monitor step progression through P14
    - Check for wait count anomalies or PC stalls
    - _Requirements: 5.1, 5.2_
  
  - [x] 2.3 Implement fix for step progression
    - Based on investigation, fix identified issue in pkg/engine/sequencer.go or pkg/engine/vm.go
    - Ensure program counter advances correctly after each step
    - Add bounds checking for step indices if needed
    - _Requirements: 2.1, 2.3, 2.5_
  
  - [x] 2.4 Write property test for step progression
    - **Property 2: Sequencer Step Progression**
    - **Validates: Requirements 2.1, 2.3, 2.5**
    - Generate random mes() blocks with various step counts
    - Verify sequencer executes all steps without hanging
    - Test with step indices from 0 to 100
  
  - [x] 2.5 Verify y_saru completes successfully
    - Run y_saru sample to completion
    - Verify it progresses past P14
    - Verify game terminates normally
    - _Requirements: 6.3_

- [x] 3. Investigate and fix yosemiya curtain animation
  - [x] 3.1 Analyze yosemiya animation code
    - Read samples/yosemiya/YOSEMIYA.TFY
    - Identify curtain opening animation sequence
    - Document MovePic() commands and timing
    - _Requirements: 5.1, 5.2, 5.3_
  
  - [x] 3.2 Run yosemiya with animation logging
    - Execute with `--debug=2` to capture MovePic() calls
    - Monitor cast position updates
    - Check if animation frames are being executed
    - _Requirements: 5.1, 5.2_
  
  - [x] 3.3 Implement fix for animation rendering
    - Based on investigation, fix identified issue in pkg/engine/drawing.go or cast management
    - Ensure MovePic() updates cast positions correctly
    - Verify rendering pipeline processes cast updates
    - _Requirements: 3.1, 3.2, 3.5_
  
  - [x] 3.4 Write property test for animation execution
    - **Property 5: Animation Frame Completeness**
    - **Validates: Requirements 3.1, 3.2, 3.5**
    - Generate random sequences of MovePic() commands
    - Verify all frames execute in correct order
    - Count executed frames vs expected frames
  
  - [x] 3.5 Verify yosemiya animation plays correctly
    - Run yosemiya sample
    - Verify curtain opening animation executes
    - Check logs for all expected MovePic() calls
    - _Requirements: 6.4_

- [x] 4. Investigate and fix yosemiya text mojibake
  - [x] 4.1 Analyze text encoding flow
    - Trace text from TFY source through preprocessor to TextWrite()
    - Verify Shift-JIS to UTF-8 conversion in pkg/compiler/preprocessor/preprocessor.go
    - Check string literal encoding in compiled OpCodes
    - _Requirements: 5.1, 5.2, 5.3_
  
  - [x] 4.2 Run yosemiya with text rendering logging
    - Execute with `--debug=2` to capture TextWrite() calls
    - Log text strings being rendered
    - Check for encoding corruption in logs
    - _Requirements: 5.1, 5.2_
  
  - [x] 4.3 Implement fix for text encoding
    - Based on investigation, fix identified issue in pkg/engine/text.go or preprocessor
    - Ensure UTF-8 encoding is preserved through compilation
    - Verify TextWrite() handles multi-byte characters correctly
    - Ensure font loading supports Japanese character sets
    - _Requirements: 4.1, 4.2, 4.3, 4.4_
  
  - [x] 4.4 Write property test for text encoding
    - **Property 6: Text Encoding Preservation**
    - **Validates: Requirements 4.1, 4.2, 4.3, 4.4, 4.5**
    - Generate random Japanese text strings
    - Verify Shift-JIS to UTF-8 conversion is correct
    - Verify TextWrite() renders without mojibake
    - Test with various multi-byte character combinations
  
  - [x] 4.5 Verify yosemiya text displays correctly
    - Run yosemiya sample
    - Verify virtual window text is readable
    - Check that Japanese characters display without corruption
    - _Requirements: 6.5_

- [-] 5. Implement comprehensive termination property test
  - [x] 5.1 Write property test for engine termination
    - **Property 1: Engine Termination Completeness**
    - **Validates: Requirements 1.1, 1.3, 1.4, 1.5**
    - Generate random TFY scripts with various completion patterns
    - Verify engine terminates within 1 second after completion
    - Verify exit code is 0 for successful completion
    - Test with mes() blocks that have no more scheduled events

- [ ] 6. Implement step timing and completion tests
  - [ ] 6.1 Write property test for step timing accuracy
    - **Property 3: Step Timing Accuracy**
    - **Validates: Requirements 2.2**
    - Generate random step sequences with specific tick timings
    - Verify steps execute at correct moments in both TIME and MIDI_TIME modes
    - Test with various ticksPerStep values
  
  - [ ] 6.2 Write property test for step completion detection
    - **Property 4: Step Block Completion Detection**
    - **Validates: Requirements 2.4**
    - Generate random mes() blocks with various step counts
    - Verify sequencer marks blocks as finished when all steps complete
    - Verify engine proceeds with termination checks after completion

- [ ] 7. Checkpoint - Run all sample games and verify fixes
  - Run sab2 and verify it parses and terminates
  - Run y_saru and verify it progresses past P14
  - Run yosemiya and verify animation and text work correctly
  - Ensure all tests pass, ask the user if questions arise.
  - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5_

- [ ] 8. Regression testing
  - [ ] 8.1 Test other sample games
    - Run samples/test_minimal
    - Run samples/test_loop
    - Run samples/test_drawing
    - Run samples/mes_test
    - Verify all complete successfully without new errors
    - _Requirements: 7.2_
  
  - [ ] 8.2 Run existing test suite
    - Execute `go test ./pkg/engine/... -v`
    - Execute `go test ./pkg/compiler/... -v`
    - Verify all existing tests pass
    - _Requirements: 7.3, 7.5_

- [ ] 9. Final checkpoint - Verify all requirements met
  - Verify all four sample game issues are resolved
  - Verify all property tests pass
  - Verify no regressions in existing functionality
  - Ensure all tests pass, ask the user if questions arise.
  - _Requirements: 6.6, 7.4, 7.5_

## Notes

- Each task references specific requirements for traceability
- Investigation tasks should be completed before implementing fixes
- Property tests validate universal correctness properties
- Unit tests validate specific examples and edge cases
- Checkpoints ensure incremental validation
