# Requirements Document

## Introduction

This specification addresses timing accuracy issues in the son-et MIDI synchronization system. The current implementation experiences irregular tick delivery due to audio buffer processing, causing animations and script execution to be out of sync with MIDI playback. This feature will ensure smooth, accurate tick generation that maintains precise synchronization between MIDI tempo and script execution.

## Glossary

- **MIDI_Player**: The component responsible for MIDI file playback and audio synthesis
- **VM**: The Virtual Machine that executes FILLY script opcodes
- **Tick**: A discrete time unit used by the VM for timing synchronization
- **PPQ**: Pulses Per Quarter note - MIDI timing resolution (typically 480)
- **Step**: A user-defined musical subdivision (e.g., step(8) = eighth notes)
- **Audio_Buffer**: A block of audio samples processed at once (variable size)
- **Sample_Rate**: Audio samples per second (44100 Hz)
- **Tempo**: Musical tempo in beats per minute (BPM)
- **Tick_Generator**: Component that converts audio sample count to MIDI ticks

## Requirements

### Requirement 1: Accurate Tick Calculation

**User Story:** As a developer, I want MIDI ticks to be calculated accurately from audio sample count, so that timing synchronization is mathematically correct.

#### Acceptance Criteria

1. WHEN audio samples are processed, THE Tick_Generator SHALL calculate the current tick position based on sample count, sample rate, tempo, and PPQ
2. WHEN tempo changes occur in the MIDI file, THE Tick_Generator SHALL update tick calculations to reflect the new tempo
3. THE Tick_Generator SHALL maintain fractional tick precision internally to prevent cumulative rounding errors
4. WHEN calculating ticks from samples, THE Tick_Generator SHALL use the formula: `ticks = (samples * tempo * PPQ) / (sample_rate * 60)`

### Requirement 2: Smooth Tick Delivery

**User Story:** As a script author, I want ticks to be delivered smoothly to the VM, so that animations and timing-based logic execute at consistent intervals.

#### Acceptance Criteria

1. WHEN the MIDI_Player processes an audio buffer, THE Tick_Generator SHALL deliver only the net tick advancement, not individual tick increments
2. WHEN ticks advance by N positions, THE Tick_Generator SHALL call NotifyTick once with the new tick value, not N times
3. WHILE the MIDI_Player is playing, THE Tick_Generator SHALL ensure tick delivery happens at regular intervals corresponding to the audio buffer processing rate
4. THE Tick_Generator SHALL prevent delivering the same tick value multiple times

### Requirement 3: Timing Accuracy Verification

**User Story:** As a developer, I want to verify timing accuracy, so that I can confirm synchronization is within acceptable tolerances.

#### Acceptance Criteria

1. WHEN a Wait operation specifies N steps, THE VM SHALL wait for exactly N * (PPQ / step_divisor) ticks
2. WHEN measuring elapsed time for a Wait operation, THE actual duration SHALL be within 50ms of the expected duration
3. THE Tick_Generator SHALL log timing information including current tick, tempo, and sample position for debugging
4. WHEN running in headless mode with timeout, THE timing accuracy SHALL be equivalent to GUI mode

### Requirement 4: Tempo Change Handling

**User Story:** As a script author, I want tempo changes in MIDI files to be handled correctly, so that synchronization remains accurate throughout playback.

#### Acceptance Criteria

1. WHEN a MIDI tempo change event occurs, THE Tick_Generator SHALL update its tempo value immediately
2. WHEN calculating ticks after a tempo change, THE Tick_Generator SHALL use the new tempo for subsequent calculations
3. THE Tick_Generator SHALL maintain tick continuity across tempo changes without jumps or gaps
4. WHEN multiple tempo changes occur, THE Tick_Generator SHALL handle each change independently and accurately

### Requirement 5: Buffer Size Independence

**User Story:** As a developer, I want timing accuracy to be independent of audio buffer size, so that synchronization works consistently across different audio configurations.

#### Acceptance Criteria

1. WHEN audio buffer size varies, THE Tick_Generator SHALL produce the same tick values at the same sample positions
2. THE Tick_Generator SHALL not accumulate timing errors due to buffer size variations
3. WHEN processing large audio buffers, THE Tick_Generator SHALL maintain the same accuracy as with small buffers
4. THE timing accuracy SHALL be deterministic and reproducible regardless of buffer processing patterns

### Requirement 6: Integration with VM Wait Operations

**User Story:** As a script author, I want Wait operations to synchronize accurately with MIDI playback, so that animations align with musical events.

#### Acceptance Criteria

1. WHEN a Wait operation is executed in MIDI_TIME mode, THE VM SHALL calculate the target tick based on current step divisor and wait count
2. WHEN the target tick is reached, THE VM SHALL resume execution within one audio buffer processing cycle
3. THE VM SHALL handle Wait operations that span multiple audio buffer processing cycles
4. WHEN MIDI playback ends, THE VM SHALL handle pending Wait operations gracefully

### Requirement 7: Headless Mode Timing

**User Story:** As a developer, I want timing accuracy in headless mode, so that I can test MIDI synchronization without GUI.

#### Acceptance Criteria

1. WHEN running in headless mode, THE Tick_Generator SHALL produce the same tick values as GUI mode
2. WHEN running with timeout in headless mode, THE timing accuracy SHALL remain within acceptable tolerances
3. THE headless mode SHALL log tick advancement for verification
4. WHEN audio is muted in headless mode, THE MIDI timing SHALL continue to function correctly

### Requirement 8: Error Handling and Edge Cases

**User Story:** As a developer, I want robust error handling for timing edge cases, so that the system remains stable under unusual conditions.

#### Acceptance Criteria

1. WHEN MIDI file has invalid tempo values, THE Tick_Generator SHALL use a default tempo and log a warning
2. WHEN sample rate is zero or invalid, THE Tick_Generator SHALL return an error and prevent initialization
3. IF audio buffer processing is delayed, THEN THE Tick_Generator SHALL catch up smoothly without skipping ticks
4. WHEN MIDI playback is paused or stopped, THE Tick_Generator SHALL maintain its current tick position
