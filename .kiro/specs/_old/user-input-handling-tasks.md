# Implementation Plan: User Input Handling

## Overview

This implementation plan converts the design for non-blocking mes(TIME) execution into discrete coding tasks. The approach is to incrementally modify the engine to remove blocking behavior while maintaining timing accuracy and adding input event handling.

## Tasks

- [x] 1. Remove blocking behavior from RegisterSequence
  - Modify `RegisterSequence` function in `pkg/engine/engine.go`
  - Remove `WaitGroup` creation and blocking for TIME mode
  - Remove `onComplete` callback usage for TIME mode
  - Keep MIDI_TIME mode behavior unchanged
  - Add logging to confirm non-blocking operation
  - _Requirements: 1.1, 1.4_

- [x] 1.1 Write property test for RegisterSequence timing
  - **Property 1: RegisterSequence Non-blocking**
  - **Validates: Requirements 1.1**
  - Test that RegisterSequence returns within 10ms for TIME mode
  - Use `testing/quick` to generate random OpCode sequences
  - Measure time from call to return
  - _Requirements: 1.1_

- [x] 2. Add termination flag and ESC key handling
  - [x] 2.1 Add programTerminated global variable
    - Add `var programTerminated bool` to global variables section
    - Initialize to false
    - _Requirements: 6.1_
  
  - [x] 2.2 Add ESC key detection to Game.Update()
    - Check `ebiten.IsKeyPressed(ebiten.KeyEscape)` at start of Update()
    - Set `programTerminated = true` when ESC is pressed
    - Return `ebiten.Termination` when ESC is pressed
    - Add logging for ESC key detection
    - _Requirements: 3.1, 3.2_
  
  - [x] 2.3 Add termination check at start of Update()
    - Check `programTerminated` flag before processing
    - Return `ebiten.Termination` if flag is set
    - Add logging for termination detection
    - _Requirements: 2.5, 4.3, 6.4_

- [x] 2.4 Write unit tests for ESC key handling
  - Test ESC key sets programTerminated flag
  - Test Update() returns ebiten.Termination when flag is set
  - Test termination check happens before VM execution
  - _Requirements: 3.1, 3.2_

- [x] 3. Add termination handling to UpdateVM
  - [x] 3.1 Check programTerminated flag in UpdateVM
    - Add check at start of UpdateVM function
    - Return early if flag is set
    - Mark all active sequences as inactive
    - Add logging for termination in UpdateVM
    - _Requirements: 6.1, 6.2_
  
  - [x] 3.2 Check termination flag in ExecuteOp
    - Add check before executing each OpCode
    - Return ebiten.Termination if flag is set
    - Propagate termination signal up to UpdateVM
    - _Requirements: 6.1, 6.2_

- [x] 3.3 Write property test for termination stops execution
  - **Property 9: Termination Stops Execution**
  - **Validates: Requirements 6.1, 6.2**
  - Generate random sequences with Wait() operations
  - Set programTerminated flag mid-execution
  - Verify no new OpCodes execute after flag is set
  - _Requirements: 6.1, 6.2_

- [x] 4. Checkpoint - Test with kuma2 sample
  - Run kuma2 sample with new non-blocking implementation
  - Verify window is responsive (no busy cursor)
  - Verify ESC key closes the window
  - Verify audio plays correctly
  - Verify images display correctly
  - Ensure all tests pass, ask the user if questions arise.

- [x] 5. Add timing accuracy tests
  - [x] 5.1 Write property test for frame-accurate timing
    - **Property 6: Frame-Accurate Timing**
    - **Validates: Requirements 5.1**
    - Measure tick count increments over time
    - Verify 60 ticks per second (±5% tolerance)
    - Test under various sequence loads
    - _Requirements: 5.1_
  
  - [x] 5.2 Write property test for Wait() accuracy
    - **Property 7: Wait Operation Accuracy**
    - **Validates: Requirements 5.2, 8.3**
    - Generate random Wait(N) operations
    - Count ticks until next OpCode executes
    - Verify exactly N ticks passed
    - _Requirements: 5.2, 8.3_
  
  - [x] 5.3 Write property test for concurrent timing independence
    - **Property 8: Concurrent Timing Independence**
    - **Validates: Requirements 5.3**
    - Register multiple sequences with different Wait() patterns
    - Verify each sequence maintains its own timing
    - Verify sequences don't interfere with each other
    - _Requirements: 5.3_

- [x] 6. Add game loop continuity tests
  - [x] 6.1 Write property test for game loop continuity
    - **Property 2: Game Loop Continuity**
    - **Validates: Requirements 1.2, 3.3, 4.4, 7.4**
    - Register long-running sequence
    - Measure time between Update() calls
    - Verify approximately 16.67ms per frame (±2ms)
    - _Requirements: 1.2, 3.3, 4.4, 7.4_
  
  - [x] 6.2 Write property test for rendering frame rate
    - **Property 3: Rendering Frame Rate**
    - **Validates: Requirements 1.3**
    - Measure time between Draw() calls
    - Verify approximately 16.67ms per frame
    - Test under various sequence loads
    - _Requirements: 1.3_
  
  - [x] 6.3 Write property test for input responsiveness during Wait
    - **Property 5: Input Responsiveness During Wait**
    - **Validates: Requirements 3.4**
    - Execute Wait(100) operation
    - Simulate input events during wait
    - Verify Update() continues to be called
    - Verify input events are processed
    - _Requirements: 3.4_

- [x] 7. Add concurrent execution tests
  - [x] 7.1 Write property test for concurrent sequence execution
    - **Property 4: Concurrent Sequence Execution**
    - **Validates: Requirements 1.4**
    - Register multiple mes(TIME) blocks concurrently
    - Verify all sequences execute in parallel
    - Verify no blocking between sequences
    - _Requirements: 1.4_
  
  - [x] 7.2 Write property test for Update advances VM
    - **Property 10: Update Advances VM**
    - **Validates: Requirements 7.1**
    - Call Game.Update() multiple times
    - Verify tick count increments by 1 each time
    - Test in TIME mode only
    - _Requirements: 7.1_
  
  - [x] 7.3 Write property test for UpdateVM non-blocking
    - **Property 11: UpdateVM Non-blocking**
    - **Validates: Requirements 7.3**
    - Measure UpdateVM() execution time
    - Verify < 5ms for typical sequences
    - Test with various sequence complexities
    - _Requirements: 7.3_

- [x] 8. Add error handling tests
  - [x] 8.1 Write unit test for OpCode execution errors
    - Trigger OpCode execution error
    - Verify error is logged with context
    - Verify sequence is marked inactive
    - Verify other sequences continue
    - _Requirements: 10.1, 10.2_
  
  - [x] 8.2 Write property test for error recovery responsiveness
    - **Property 12: Error Recovery Responsiveness**
    - **Validates: Requirements 10.4**
    - Trigger OpCode execution error
    - Verify Update() continues to be called
    - Verify input events are still processed
    - _Requirements: 10.4_

- [x] 9. Add backward compatibility tests
  - [x] 9.1 Write unit tests for existing scripts
    - Test kuma2 sample behavior
    - Test robot sample behavior (if available)
    - Verify timing is preserved
    - Verify output matches expected behavior
    - _Requirements: 8.1, 8.2, 8.3, 8.4_
  
  - [x] 9.2 Write unit test for MIDI_TIME mode unchanged
    - Verify MIDI_TIME mode is still non-blocking
    - Verify MIDI synchronization works correctly
    - Verify event handlers trigger correctly
    - _Requirements: 8.2, 8.4_

- [x] 10. Add headless mode tests
  - [x] 10.1 Write unit test for headless execution
    - Run sequence in headless mode
    - Verify execution completes without GUI
    - Verify timing accuracy maintained
    - _Requirements: 9.1, 9.2_
  
  - [x] 10.2 Write unit test for headless timeout
    - Run with --timeout flag
    - Verify program terminates after timeout
    - Verify exit code is 0
    - _Requirements: 9.3_
  
  - [x] 10.3 Write unit test for headless logging
    - Verify log messages contain timestamps
    - Verify execution progress is logged
    - _Requirements: 9.4_

- [x] 11. Final checkpoint - Integration testing
  - Run all samples (kuma2, robot, etc.)
  - Verify window responsiveness in all cases
  - Verify ESC key works in all cases
  - Verify timing accuracy is maintained
  - Verify no regressions in existing functionality
  - Ensure all tests pass, ask the user if questions arise.

- [x] 12. Documentation and cleanup
  - Update code comments to reflect non-blocking behavior
  - Remove unused WaitGroup code
  - Update any documentation that mentions blocking behavior
  - Add comments explaining termination flow
  - _Requirements: All_

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- Each task references specific requirements for traceability
- Checkpoints ensure incremental validation
- Property tests validate universal correctness properties
- Unit tests validate specific examples and edge cases
- Integration tests verify end-to-end functionality
- The implementation should maintain backward compatibility with existing scripts
- All timing behavior should be preserved
- The game loop should remain responsive at all times
