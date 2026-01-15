# Task 13: ExitTitle and Event Handlers Implementation

## Date
2026-01-15

## Objective
Implement ExitTitle function and event handlers for mes(MIDI_END), mes(RBDOWN), and mes(RBDBLCLK) to support more sample scenarios.

## Changes Made

### 1. Engine Changes (pkg/engine/engine.go)

#### Added ExitTitle to ExecuteOp
- Added case "ExitTitle" in the ExecuteOp switch statement (around line 1710)
- Calls the existing ExitTitle() function which terminates the program with os.Exit(0)

#### Event Handler Infrastructure (Already Implemented)
The following were already implemented in previous work:
- Global event handler variables: `midiEndHandler`, `rbDownHandler`, `rbDblClkHandler`
- Registration functions: `RegisterMidiEndHandler()`, `RegisterRBDownHandler()`, `RegisterRBDblClkHandler()`
- Trigger functions: `TriggerMidiEnd()`, `TriggerRBDown()`, `TriggerRBDblClk()`
- Mouse event handling in `Game.Update()` for right button clicks

### 2. Code Generator Changes (pkg/compiler/codegen/codegen.go)

#### Event Handler Code Generation (Already Implemented)
The genStatement function already handles mes() blocks for event handlers:
- mes(MIDI_END) generates `engine.RegisterMidiEndHandler(func() { ... })`
- mes(RBDOWN) generates `engine.RegisterRBDownHandler(func() { ... })`
- mes(RBDBLCLK) generates `engine.RegisterRBDblClkHandler(func() { ... })`

### 3. MIDI Player Changes (pkg/engine/midi_player.go)

#### MIDI_END Detection (Partial Implementation)
- Added TODO comment for MIDI_END detection in ProcessSamples()
- The meltysynth library doesn't expose an EndOfSequence() method
- Removed the non-working `s.sequencer.EndOfSequence()` call
- Future work needed: Track MIDI file length and compare with rendered samples

## Testing

### Test File Created
Created a simple test file to verify:
- ExitTitle function call
- mes(MIDI_END) handler registration
- mes(RBDOWN) handler registration
- mes(RBDBLCLK) handler registration

### Test Results
- ✅ Code generation works correctly
- ✅ Generated Go code compiles without errors
- ✅ Event handlers are registered properly
- ✅ ExitTitle is called as an OpCode in sequences

## Known Limitations

### MIDI_END Event
- MIDI_END events are not automatically triggered when MIDI playback finishes
- The meltysynth library (v0.0.0-20230205031334-05d311382fc4) doesn't provide an EndOfSequence() method
- Manual triggering of MIDI_END is possible via TriggerMidiEnd() but automatic detection needs implementation

### Double-Click Detection
- Right button click detection is implemented
- Full double-click timing logic needs refinement
- Currently only basic click detection is in place

## Sample Compatibility

These implementations enable the following functionality:
- Event handlers for MIDI_END, RBDOWN, RBDBLCLK
- ExitTitle function for window management

Note: Some samples may also use Shell() and MCI() functions which are not yet implemented.

## Future Work

1. **MIDI_END Auto-Detection**
   - Research meltysynth API for sequence completion detection
   - Implement sample counting to detect when MIDI file has finished
   - Consider tracking MIDI file length from MidiFile.GetLength()

2. **Double-Click Timing**
   - Implement proper double-click detection with timing window
   - Track time between clicks to distinguish single vs double clicks

3. **Shell() Function**
   - Implement Shell() for launching external programs
   - Platform-specific implementation (Windows vs macOS/Linux)

4. **MCI() Function**
   - Implement MCI() for Windows Media Control Interface
   - May need platform-specific handling or alternative for non-Windows

## Files Modified

- pkg/engine/engine.go (added ExitTitle case to ExecuteOp)
- pkg/engine/midi_player.go (commented out non-working EndOfSequence call)

## Files Already Containing Related Code

- pkg/engine/engine.go (event handler infrastructure)
- pkg/compiler/codegen/codegen.go (event handler code generation)

## Verification Commands

```bash
# Test compilation
go build ./pkg/engine/...

# Test with sample (requires assets)
go run cmd/son-et/main.go samples/xxx/SCRIPT.TFY > samples/xxx/game.go
cd samples/xxx
go build -o game game.go
```
