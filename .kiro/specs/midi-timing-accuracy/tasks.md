# Implementation Plan: MIDI Timing Accuracy

## Overview

This implementation plan breaks down the MIDI timing accuracy improvements into discrete coding tasks. The approach focuses on creating a new `TickGenerator` component that maintains fractional precision and delivers smooth tick updates to the VM, replacing the current inefficient loop-based notification system.

## Implementation Summary

The implementation has been completed with the following key changes:

1. **Wall-Clock Based Timing**: Switched from sample-count-based tick calculation to wall-clock time measurement using `time.Since(startTime)`. This eliminates cumulative drift from audio buffer processing delays.

2. **Tempo-Aware Tick Calculation**: Implemented `CalculateTickFromTime()` method in `TickGenerator` that properly accounts for tempo changes by traversing the tempo map and calculating ticks for each tempo segment.

3. **Sequential Tick Delivery**: Modified `MidiStream.Read()` to deliver all ticks from `lastDeliveredTick+1` to `currentTick` sequentially, preventing animation frame skipping even when processing is delayed.

4. **Wait Operation Fix**: Fixed off-by-one error in Wait operation by setting `seq.waitTicks = totalTicks - 1`, ensuring Wait(N steps) waits exactly N steps instead of N+1.

**Timing Accuracy Results** (y_saru sample, 60 seconds):
- Expected: 57.62 seconds (59,040 ticks at 128.07 BPM)
- Actual: 58.02 seconds
- Drift: +0.40 seconds (0.69% too slow)
- Animation skipping: None (all MoveCast operations at correct 480-tick intervals)

## Tasks

- [x] 1. Create TickGenerator component with core data structures
  - Create new file `pkg/engine/tick_generator.go`
  - Define `TickGenerator` struct with fields: sampleRate, ppq, tempoMap, currentSamples, fractionalTick, lastDeliveredTick, tempoMapIndex, currentTempo
  - Define `TempoEvent` struct (if not already defined)
  - Add constructor `NewTickGenerator(sampleRate, ppq int, tempoMap []TempoEvent) (*TickGenerator, error)`
  - Validate inputs (sample rate > 0, ppq > 0)
  - Initialize state fields to zero/default values
  - _Requirements: 1.1, 1.3, 8.2_

- [ ] 1.1 Write property test for TickGenerator initialization
  - **Property 1: Tick Calculation Formula Accuracy**
  - **Validates: Requirements 1.1, 1.4**

- [x]* 1.2 Write unit tests for invalid inputs
  - Test zero sample rate returns error
  - Test negative sample rate returns error
  - Test zero PPQ returns error
  - _Requirements: 8.2_

- [ ] 2. Implement tick calculation with fractional precision
  - [x] 2.1 Implement `ProcessSamples(numSamples int) int` method
    - Add numSamples to currentSamples
    - Calculate tick advancement using formula: `(samples * tempo * PPQ) / (sample_rate * 60)`
    - Update fractionalTick (maintain float64 precision)
    - Check if integer part changed
    - If changed, update lastDeliveredTick and return new tick
    - If unchanged, return -1
    - _Requirements: 1.1, 1.3, 1.4, 2.1_

  - [x]* 2.2 Write property test for fractional precision preservation
    - **Property 2: Fractional Precision Preservation**
    - **Validates: Requirements 1.3, 5.2**

  - [x]* 2.3 Write property test for tick calculation formula
    - **Property 1: Tick Calculation Formula Accuracy**
    - **Validates: Requirements 1.1, 1.4**

- [ ] 3. Implement tempo map handling
  - [x] 3.1 Add tempo map traversal logic to ProcessSamples
    - Check if current sample position crosses next tempo boundary
    - If yes, calculate ticks for segment before boundary using old tempo
    - Update tempoMapIndex and currentTempo
    - Calculate remaining ticks using new tempo
    - Maintain tick continuity across boundaries
    - _Requirements: 1.2, 4.1, 4.2, 4.3, 4.4_

  - [x]* 3.2 Write property test for tempo change correctness
    - **Property 3: Tempo Change Correctness**
    - **Validates: Requirements 1.2, 4.1, 4.2, 4.3, 4.4**

  - [~]* 3.3 Write unit tests for tempo change edge cases
    - Test tempo change at tick 0
    - Test multiple tempo changes in one buffer
    - Test tempo change at exact buffer boundary
    - _Requirements: 4.4_

- [x] 4. Add helper methods to TickGenerator
  - Implement `GetCurrentTick() int` - returns lastDeliveredTick
  - Implement `GetFractionalTick() float64` - returns fractionalTick
  - Implement `Reset()` - resets all state to initial values
  - _Requirements: 8.4_

- [~]* 4.1 Write unit tests for helper methods
  - Test GetCurrentTick returns correct value
  - Test GetFractionalTick returns precise value
  - Test Reset clears all state
  - _Requirements: 8.4_

- [x] 5. Integrate TickGenerator with MidiStream
  - [x] 5.1 Add tickGenerator field to MidiStream struct
    - Add `tickGenerator *TickGenerator` field to MidiStream in `pkg/engine/midi_player.go`
    - Add `startTime time.Time` field to track playback start
    - Initialize in `PlayMidiFile()` after parsing tempo map
    - Pass sampleRate, ppq, and tempoMap to NewTickGenerator
    - Handle initialization errors
    - _Requirements: 2.1, 2.2_

  - [x] 5.2 Modify MidiStream.Read() to use TickGenerator
    - Use `time.Since(startTime)` to get elapsed time
    - Call `CalculateTickFromTime()` to get current tick
    - Deliver all ticks from `lastDeliveredTick+1` to `currentTick` sequentially
    - Update `lastDeliveredTick` after delivery
    - _Requirements: 2.1, 2.2, 2.4_

  - [~]* 5.3 Write property test for single tick delivery
    - **Property 4: Single Tick Delivery**
    - **Validates: Requirements 2.1, 2.2**

  - [~]* 5.4 Write property test for monotonic tick progression
    - **Property 5: Monotonic Tick Progression**
    - **Validates: Requirements 2.4, 4.3**

- [x] 6. Checkpoint - Ensure all tests pass
  - Run all unit tests: `go test -timeout=30s ./pkg/engine/...`
  - Run all property tests with 100+ iterations
  - Verify no compilation errors
  - Check for race conditions: `go test -race -timeout=30s ./pkg/engine/...`
  - Test with y_saru sample to verify timing accuracy and no animation skipping
  - Ask the user if questions arise

- [x] 6.1 Fix Wait operation timing
  - Identified off-by-one error in Wait operation
  - Modified `seq.waitTicks = totalTicks - 1` to account for decrement on next tick
  - Verified Wait(N steps) now waits exactly N steps
  - Tested with y_saru sample: all MoveCast operations execute at correct 480-tick intervals
  - _Requirements: 3.1, 6.1_

- [ ] 7. Add timing accuracy verification
  - [ ] 7.1 Add logging to TickGenerator
    - Log tick advancement every 100 ticks with timestamp
    - Include current tick, tempo, sample position in log
    - Use format: `[HH:MM:SS.mmm] TickGenerator: tick=%d, tempo=%.2f BPM, samples=%d`
    - _Requirements: 3.3, 7.3_

  - [~]* 7.2 Write property test for timing information logging
    - **Property 14: Timing Information Logging**
    - **Validates: Requirements 3.3, 7.3**

  - [~]* 7.3 Write property test for wait operation timing accuracy
    - **Property 8: Wait Operation Timing Accuracy**
    - **Validates: Requirements 3.2**

- [ ] 8. Implement buffer size determinism
  - [~]* 8.1 Write property test for buffer size determinism
    - **Property 11: Buffer Size Determinism**
    - **Validates: Requirements 5.1, 5.2, 5.3, 5.4**

  - [~]* 8.2 Write property test for regular tick delivery intervals
    - **Property 6: Regular Tick Delivery Intervals**
    - **Validates: Requirements 2.3**

- [ ] 9. Add headless mode verification
  - [~]* 9.1 Write property test for headless mode equivalence
    - **Property 12: Headless Mode Equivalence**
    - **Validates: Requirements 3.4, 7.1, 7.4**

  - [~]* 9.2 Write property test for headless mode timing accuracy
    - **Property 13: Headless Mode Timing Accuracy**
    - **Validates: Requirements 7.2**

- [ ] 10. Implement error handling
  - [ ] 10.1 Add invalid tempo handling
    - Detect invalid tempo values during tempo map parsing
    - Replace with default 120 BPM (500000 microseconds per beat)
    - Log warning: `"Warning: Invalid tempo value %d at tick %d, using default 120 BPM"`
    - _Requirements: 8.1_

  - [ ] 10.2 Add MIDI end handling during wait
    - Add flag to track MIDI end state
    - Set flag when sequencer finishes in MidiStream.Read()
    - Check flag in VM wait operations
    - Resume execution or mark sequence complete if MIDI ended
    - Log: `"MIDI playback ended during wait, resuming execution"`
    - _Requirements: 6.4_

  - [~]* 10.3 Write unit tests for error handling
    - Test invalid tempo handling
    - Test MIDI end during wait
    - _Requirements: 8.1, 6.4_

- [ ] 11. Add VM integration tests
  - [~]* 11.1 Write property test for wait operation tick calculation
    - **Property 7: Wait Operation Tick Calculation**
    - **Validates: Requirements 3.1, 6.1**

  - [~]* 11.2 Write property test for wait resume latency
    - **Property 9: Wait Resume Latency**
    - **Validates: Requirements 6.2**

  - [~]* 11.3 Write property test for multi-buffer wait handling
    - **Property 10: Multi-Buffer Wait Handling**
    - **Validates: Requirements 6.3**

- [ ] 12. Add edge case handling
  - [~]* 12.1 Write property test for delayed processing catch-up
    - **Property 15: Delayed Processing Catch-Up**
    - **Validates: Requirements 8.3**

  - [~]* 12.2 Write property test for pause state preservation
    - **Property 16: Pause State Preservation**
    - **Validates: Requirements 8.4**

  - [~]* 12.3 Write unit tests for edge cases
    - Test zero-length audio buffers
    - Test first buffer (initialization)
    - Test very large buffer sizes
    - _Requirements: 8.3, 8.4_

- [ ] 13. Final checkpoint - Comprehensive testing
  - Run all tests with race detector: `go test -race -timeout=30s ./pkg/engine/...`
  - Test with existing samples: `go run cmd/son-et/main.go --headless --timeout=10s samples/kuma2`
  - Verify timing accuracy in logs
  - Compare headless vs GUI mode behavior
  - Ensure all property tests pass with 100+ iterations
  - Ask the user if questions arise

- [ ] 14. Clean up and documentation
  - Remove old tick calculation code (CalculateTickFromTime, ProcessSamples loop)
  - Remove unused global variables (currentSamples, lastTick)
  - Add code comments explaining tick calculation algorithm
  - Update any relevant documentation
  - _Requirements: All_

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- Each task references specific requirements for traceability
- Checkpoints ensure incremental validation
- Property tests validate universal correctness properties with 100+ iterations
- Unit tests validate specific examples and edge cases
- Integration tests verify end-to-end behavior with real MIDI files


## Additional Tasks - Logging and Termination Issues

- [x] 15. Reduce excessive logging during normal execution
  - [x] 15.1 Reduce VM execution logging
    - Modified "VM: Executing" log to only output at debugLevel >= 3
    - Location: pkg/engine/engine.go:2155
    - _Status: Completed and verified with kuma2_

  - [x] 15.2 Reduce NotifyTick logging
    - Modified "NotifyTick" log to only output at debugLevel >= 2
    - Location: pkg/engine/engine.go:1738
    - _Status: Completed and verified with kuma2_

  - [x] 15.3 Reduce MIDI wait logging frequency
    - Changed "Program terminated, waiting for MIDI" log from 1/second to 1/10seconds
    - Modified from `currentTick%60` to `currentTick%600`
    - Location: pkg/engine/engine.go:2078
    - _Status: Completed and verified with kuma2_

- [ ] 16. Fix yosemiya MIDI termination hang
  - **Problem**: yosemiya continues waiting for MIDI to finish for 5+ minutes after all sequences complete
  - **Expected**: MIDI should finish in ~64 seconds (46,210 ticks at 120 BPM, 48 PPQ)
  - **Symptoms**:
    - "All sequences finished, terminating program" appears at 23:59:42
    - "Program terminated, waiting for MIDI to finish" continues every 10 seconds indefinitely
    - Program never terminates naturally
  
  - [ ] 16.1 Add debug logging to investigate midiFinished flag
    - Add logging in MidiStream.Read() to show:
      - currentTick vs totalTicks comparison
      - midiFinished flag state
      - endReported flag state
    - Add logging in UpdateVM/Game.Update to show:
      - midiPlayer != nil check result
      - midiFinished value when waiting
      - Current tick count
    - Location: pkg/engine/midi_player.go:480, pkg/engine/engine.go:2077
    - _Requirements: Debug and fix termination logic_

  - [ ] 16.2 Verify MidiStream.Read() is being called
    - Confirm audio processing continues after program termination
    - Check if Read() reaches the totalTicks check
    - Verify CalculateTickFromTime() returns correct values
    - _Requirements: Ensure audio thread continues processing_

  - [ ] 16.3 Check for multiple MIDI player instances
    - Verify only one midiPlayer exists at termination time
    - Check if PlayMIDI creates new player without cleaning up old one
    - Ensure midiFinished flag applies to the correct player
    - _Requirements: Single MIDI player instance management_

  - [ ] 16.4 Test with debug output
    - Run: `DEBUG_LEVEL=2 go run ./cmd/son-et/main.go samples/yosemiya > yosemiya_debug.log 2>&1 & PID=$!; sleep 120; kill -9 -$PID 2>/dev/null; wait $PID 2>/dev/null`
    - Analyze logs for tick progression and midiFinished state
    - Compare expected vs actual MIDI duration
    - _Requirements: Reproduce and diagnose issue_

  - [ ] 16.5 Implement fix based on findings
    - Fix identified issue with midiFinished flag or tick calculation
    - Ensure proper cleanup when MIDI ends
    - Test with both kuma2 and yosemiya samples
    - _Requirements: Correct MIDI termination behavior_

## Debug Information

### yosemiya MIDI Details
- File: YOSEMIYA.MID
- Total ticks: 46,210
- PPQ: 48
- Initial BPM: 120
- Expected duration: ~64 seconds
- Script calls PlayMIDI() at end of main()

### Related Files
- pkg/engine/engine.go (UpdateVM, Game.Update, runHeadless termination logic)
- pkg/engine/midi_player.go (MidiStream.Read, PlayMIDI, midiFinished flag)
- samples/yosemiya/YOSEMIYA.TFY (script structure)

### Verified Working
- kuma2 terminates correctly after MIDI finishes
- Logging is now clean and minimal
- midiFinished flag is set in MidiStream.Read() at line 481
- midiFinished flag is reset in PlayMIDI() at line 194
