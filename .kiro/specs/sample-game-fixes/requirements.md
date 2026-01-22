# Requirements Document

## Introduction

This specification addresses critical bugs in the son-et game engine that prevent sample games from running correctly. The son-et engine is a Go-based interpreter for legacy FILLY/Toffy scripts (TFY files) that provides cross-platform execution of visual novel and adventure game content. Four specific issues have been identified in sample games that indicate underlying engine defects requiring investigation and resolution.

## Glossary

- **Engine**: The son-et interpreter that executes TFY script files
- **Sample_Game**: A demonstration game located in the samples/ directory that uses TFY scripts
- **TFY_Script**: A FILLY/Toffy script file containing game logic and commands
- **VM**: The Virtual Machine component that executes compiled opcodes
- **Sequencer**: The component responsible for managing execution timing and mes() blocks
- **mes()_Block**: A timing-controlled code block that executes based on TIME or MIDI_TIME events
- **Termination**: The proper shutdown of a game when execution completes
- **Mojibake**: Character encoding corruption resulting in garbled text display
- **Cast**: A graphical element that can be moved and displayed on screen
- **Virtual_Window**: A text display window created by OpenWin() for showing text content

## Requirements

### Requirement 1: Game Termination

**User Story:** As a developer, I want sample games to terminate properly when execution completes, so that the engine exits cleanly without hanging.

#### Acceptance Criteria

1. WHEN a TFY script completes all execution paths, THE Engine SHALL terminate within 1 second
2. WHEN the main() function returns, THE Engine SHALL clean up all resources and exit
3. WHEN a game reaches its final instruction, THE Engine SHALL not enter an infinite wait state
4. IF a mes() block has no more scheduled events, THEN THE Engine SHALL recognize completion and allow termination
5. WHEN termination occurs, THE Engine SHALL return exit code 0 for successful completion

### Requirement 2: mes() Block Progression

**User Story:** As a developer, I want mes() blocks to progress through all scheduled steps, so that games execute their full logic without getting stuck.

#### Acceptance Criteria

1. WHEN a mes() block contains step() sequences, THE Sequencer SHALL execute each step in order
2. WHEN a step is scheduled for a specific tick or time, THE Sequencer SHALL execute it at the correct moment
3. IF execution reaches a specific step number (e.g., P14), THEN THE Sequencer SHALL continue to subsequent steps
4. WHEN all steps in a mes() block are complete, THE Sequencer SHALL mark the block as finished
5. THE Sequencer SHALL not skip or hang on any valid step instruction

### Requirement 3: Animation Playback

**User Story:** As a player, I want animations to play correctly, so that visual effects display as intended by the game designer.

#### Acceptance Criteria

1. WHEN an animation sequence is triggered, THE Engine SHALL execute all animation frames
2. WHEN MovePic() commands are issued in sequence, THE Engine SHALL display each movement
3. IF an animation involves multiple cast movements, THEN THE Engine SHALL coordinate timing correctly
4. WHEN a curtain opening animation is defined, THE Engine SHALL render the visual transition
5. THE Engine SHALL not skip animation frames due to timing or rendering issues

### Requirement 4: Text Encoding

**User Story:** As a player, I want text to display correctly in the proper character encoding, so that Japanese text is readable without corruption.

#### Acceptance Criteria

1. WHEN a Virtual_Window displays text, THE Engine SHALL use the correct character encoding
2. WHEN Japanese text is rendered, THE Engine SHALL interpret Shift-JIS encoding correctly
3. IF text contains multi-byte characters, THEN THE Engine SHALL not corrupt or misinterpret them
4. WHEN TextWrite() is called with Japanese strings, THE Engine SHALL display readable characters
5. THE Engine SHALL not produce mojibake (garbled characters) in text output

### Requirement 5: Issue Investigation

**User Story:** As a developer, I want to systematically investigate each reported issue, so that I can identify the root cause before implementing fixes.

#### Acceptance Criteria

1. WHEN investigating an issue, THE Developer SHALL run the affected sample game to reproduce the problem
2. WHEN a bug is reproduced, THE Developer SHALL examine relevant engine logs and execution traces
3. THE Developer SHALL identify which engine component is responsible for the malfunction
4. WHEN comparing behavior, THE Developer SHALL reference the old implementation if needed for clarification
5. THE Developer SHALL document the root cause before proceeding to implementation

### Requirement 6: Verification Testing

**User Story:** As a developer, I want to verify that fixes resolve the reported issues, so that I can confirm sample games work correctly.

#### Acceptance Criteria

1. WHEN a fix is implemented, THE Developer SHALL test the affected sample game
2. WHEN testing sab2, THE Game SHALL terminate properly after execution completes
3. WHEN testing y_saru, THE Game SHALL progress past P14 and complete all steps
4. WHEN testing yosemiya, THE Game SHALL display the curtain opening animation
5. WHEN testing yosemiya, THE Virtual_Window text SHALL display without mojibake
6. THE Developer SHALL run each sample game to completion to verify correct behavior

### Requirement 7: Regression Prevention

**User Story:** As a developer, I want to ensure fixes don't break existing functionality, so that other sample games continue to work correctly.

#### Acceptance Criteria

1. WHEN engine changes are made, THE Developer SHALL test multiple sample games
2. WHEN a fix is applied, THE Developer SHALL verify that working samples remain functional
3. IF a change affects core engine behavior, THEN THE Developer SHALL run the full test suite
4. THE Developer SHALL not introduce new bugs while fixing reported issues
5. WHEN all fixes are complete, THE Engine SHALL pass all existing tests
