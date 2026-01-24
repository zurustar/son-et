# Requirements Document: User Input Handling

## Introduction

This specification addresses the issue where the son-et engine blocks user interaction during `mes(TIME)` block execution. Currently, when a FILLY script executes a `mes(TIME)` block, the `RegisterSequence` function blocks the calling goroutine until the sequence completes, preventing the window from responding to user input events (close, move, keyboard, mouse). This results in a busy cursor (砂時計) and an unresponsive application.

The goal is to enable non-blocking execution of `mes(TIME)` blocks while maintaining timing accuracy and allowing users to interact with the application window at all times.

## Glossary

- **Engine**: The son-et runtime system that executes FILLY scripts
- **VM**: Virtual Machine component that executes OpCode sequences
- **Sequencer**: A structure that manages the execution state of a mes() block
- **mes(TIME)**: A FILLY language construct for time-based procedural execution
- **mes(MIDI_TIME)**: A FILLY language construct for MIDI-synchronized execution
- **RegisterSequence**: The function that registers a mes() block for execution
- **UpdateVM**: The function called each frame to advance VM execution
- **Game Loop**: The Ebiten game loop that calls Update() and Draw() at 60 FPS
- **OpCode**: A single instruction in the VM's instruction set
- **Blocking Execution**: Execution that prevents the calling goroutine from continuing
- **Non-blocking Execution**: Execution that allows the calling goroutine to continue immediately

## Requirements

### Requirement 1: Non-blocking mes(TIME) Execution

**User Story:** As a user, I want FILLY scripts to execute without blocking the window, so that I can interact with the application while scripts are running.

#### Acceptance Criteria

1. WHEN a script calls `mes(TIME)`, THE Engine SHALL register the sequence without blocking the calling goroutine
2. WHEN a mes(TIME) block is executing, THE Engine SHALL continue processing window events
3. WHEN a mes(TIME) block is executing, THE Engine SHALL maintain 60 FPS rendering
4. WHEN multiple mes(TIME) blocks are registered, THE Engine SHALL execute them concurrently without blocking

### Requirement 2: Window Interaction During Script Execution

**User Story:** As a user, I want to interact with the application window at all times, so that I can close, move, or resize the window even while scripts are running.

#### Acceptance Criteria

1. WHEN a script is executing, THE Engine SHALL process window close events
2. WHEN a script is executing, THE Engine SHALL process window move events
3. WHEN a script is executing, THE Engine SHALL process window resize events
4. WHEN a script is executing, THE Engine SHALL display a normal cursor (not busy cursor)
5. WHEN the user closes the window, THE Engine SHALL terminate gracefully

### Requirement 3: Keyboard Input Handling

**User Story:** As a user, I want to use keyboard shortcuts to control the application, so that I can terminate scripts or trigger actions without using the mouse.

#### Acceptance Criteria

1. WHEN the user presses ESC, THE Engine SHALL terminate the current script execution
2. WHEN the user presses ESC, THE Engine SHALL close the application window
3. WHEN keyboard events occur, THE Engine SHALL process them without blocking script execution
4. WHEN a script is waiting, THE Engine SHALL remain responsive to keyboard input

### Requirement 4: Mouse Input Handling

**User Story:** As a user, I want to use mouse input to interact with the application, so that I can trigger events or close windows.

#### Acceptance Criteria

1. WHEN the user clicks the right mouse button, THE Engine SHALL trigger the RBDOWN event handler if registered
2. WHEN the user double-clicks the right mouse button, THE Engine SHALL trigger the RBDBLCLK event handler if registered
3. WHEN the user clicks the window close button, THE Engine SHALL terminate gracefully
4. WHEN mouse events occur, THE Engine SHALL process them without blocking script execution

### Requirement 5: Timing Accuracy Preservation

**User Story:** As a developer, I want mes(TIME) blocks to maintain accurate timing, so that animations and sequences execute at the correct speed.

#### Acceptance Criteria

1. WHEN a mes(TIME) block executes, THE Engine SHALL maintain frame-accurate timing (60 FPS)
2. WHEN a Wait() operation is executed, THE Engine SHALL wait for the exact number of ticks specified
3. WHEN multiple sequences execute concurrently, THE Engine SHALL maintain timing accuracy for each sequence
4. WHEN the system is under load, THE Engine SHALL maintain timing accuracy within acceptable tolerance (±5%)

### Requirement 6: Script Termination Control

**User Story:** As a user, I want to terminate running scripts gracefully, so that I can exit the application without forcing a kill.

#### Acceptance Criteria

1. WHEN the user closes the window, THE Engine SHALL signal all active sequences to terminate
2. WHEN a termination signal is received, THE Engine SHALL stop executing new OpCodes
3. WHEN a termination signal is received, THE Engine SHALL clean up resources (close files, stop audio)
4. WHEN termination is complete, THE Engine SHALL exit with status code 0

### Requirement 7: Event Loop Integration

**User Story:** As a developer, I want the VM to integrate properly with the Ebiten game loop, so that script execution and rendering happen in sync.

#### Acceptance Criteria

1. WHEN the game loop calls Update(), THE Engine SHALL advance VM execution by one tick
2. WHEN the game loop calls Draw(), THE Engine SHALL render the current frame state
3. WHEN VM execution takes longer than one frame, THE Engine SHALL not block the game loop
4. WHEN the game loop is running, THE Engine SHALL process input events each frame

### Requirement 8: Backward Compatibility

**User Story:** As a developer, I want existing FILLY scripts to continue working, so that I don't have to rewrite scripts for the new execution model.

#### Acceptance Criteria

1. WHEN a script uses mes(TIME), THE Engine SHALL execute it with the same observable behavior as before
2. WHEN a script uses mes(MIDI_TIME), THE Engine SHALL continue to execute it non-blocking as before
3. WHEN a script uses Wait() operations, THE Engine SHALL maintain the same timing behavior
4. WHEN a script uses event handlers (MIDI_END, RBDOWN), THE Engine SHALL trigger them at the correct times

### Requirement 9: Headless Mode Compatibility

**User Story:** As a developer, I want headless mode to continue working, so that I can run automated tests without a GUI.

#### Acceptance Criteria

1. WHEN running in headless mode, THE Engine SHALL execute mes(TIME) blocks without requiring window events
2. WHEN running in headless mode, THE Engine SHALL maintain timing accuracy
3. WHEN running in headless mode with --timeout flag, THE Engine SHALL terminate after the specified duration
4. WHEN running in headless mode, THE Engine SHALL log execution progress with timestamps

### Requirement 10: Error Handling During Execution

**User Story:** As a developer, I want clear error messages when execution fails, so that I can debug issues quickly.

#### Acceptance Criteria

1. WHEN an OpCode execution fails, THE Engine SHALL log the error with context (sequence ID, PC, OpCode)
2. WHEN an OpCode execution fails, THE Engine SHALL terminate the sequence gracefully
3. WHEN a critical error occurs, THE Engine SHALL display an error message to the user
4. WHEN an error occurs, THE Engine SHALL not leave the application in an unresponsive state
