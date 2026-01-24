# FILLY Execution Model

## Overview

This document explains the execution model of FILLY scripts in son-et, focusing on the relationship between sequences, MIDI playback, and program termination. Understanding this model is critical for correctly implementing FILLY scripts.

## Core Concepts

### 1. Sequences (Threads)

A **sequence** is an independent execution unit in FILLY. Each sequence runs concurrently with others.

**Types of sequences:**
- `main()` function - The initial sequence
- `mes(TIME)` blocks - Real-time synchronized sequences
- `mes(MIDI_TIME)` blocks - MIDI-synchronized sequences
- `mes(KEY)` blocks - Keyboard event sequences
- `mes(CLICK)` blocks - Mouse click event sequences
- `mes(MIDI_END)` blocks - MIDI completion event sequences

**Key properties:**
- Sequences are **non-blocking** - they register and return immediately
- Sequences run **concurrently** - multiple sequences execute in parallel
- Sequences are **independent** - terminating one doesn't affect others

### 2. Global MIDI Player

The **MIDI player** is a global background task, completely independent from sequences.

**Key properties:**
- Started by `PlayMIDI()` function
- Runs in the background until:
  - MIDI file finishes playing
  - `StopMIDI()` is called (if implemented)
  - Program terminates
- Continues running even if the sequence that called `PlayMIDI()` terminates
- Sends MIDI ticks to all `mes(MIDI_TIME)` sequences

### 3. Termination Commands

**`del_me`** - Terminates the **current sequence only**
- Does NOT affect other sequences
- Does NOT stop MIDI playback
- Does NOT terminate the program

**`del_all`** - Terminates **all sequences**
- Stops all `mes()` blocks
- Does NOT stop MIDI playback
- Does NOT terminate the program

**`ExitTitle()`** - Terminates the **entire program**
- Stops all sequences
- Stops MIDI playback
- Closes all windows
- Exits the application

## Execution Flow Examples

### Example 1: yosemiya (Multiple TIME sequences)

```filly
main(){
    // Variable initialization
    fMAKU_OPEN=0;
    fMES_ON=0;
    
    // Sequence 0: Main animation with step() block
    mes(TIME){
        step(5){
            OpenWin(...);
            ,,,,,,,,
            fMAKU_OPEN=1;   // Set flag
            ,,,,,,,,,,,,,,,
            fMES_ON=1;      // Set flag
            ...
            del_all; del_me;  // Terminate when done
        }
    }
    
    // Sequence 1: Curtain animation (loops continuously)
    mes(TIME){
        if(fMAKU_OPEN==1 && i<=321/nMAKUMOVE){
            // Curtain opening animation
            MovePic(...);
            i=i+1;
        }
        
        if(fMES_ON==1 && j<300 && i>=321/nMAKUMOVE){
            // Display animation
            for(k=0; k<3; k=k+1;){
                X=Random(640);
                Y=Random(2)*75;
                MovePic(...);
                j=j+1;
            }
        }
        
        if(j>nMES_MOVE+180){
            del_me;  // Terminate when animation complete
        }
    }
    
    PlayMIDI("YOSEMIYA.MID");  // Start MIDI playback (global)
}
```

**Execution flow:**
1. `main()` starts executing
2. First `mes(TIME)` registers Sequence 0 and returns immediately
3. Second `mes(TIME)` registers Sequence 1 and returns immediately
4. `PlayMIDI()` starts global MIDI player and returns immediately
5. `main()` function completes
6. **Sequence 0 and 1 continue running** (independent from main)
7. **MIDI player continues running** (independent from all sequences)
8. Sequence 0 sets flags, Sequence 1 checks flags and animates
9. Both sequences eventually call `del_me` and terminate
10. MIDI continues playing until file ends

### Example 2: y-saru (MIDI_TIME sequence)

```filly
main(){
    LoadPic(...);
    
    // Sequence 0: MIDI-synchronized animation
    mes(MIDI_TIME){
        step(8){
            // Animation commands synchronized to MIDI ticks
            MovePic(...);,
            MovePic(...);,
            ...
        }
    }
    
    PlayMIDI("FLYINSKY.mid");  // Start MIDI playback (global)
    del_me;  // Terminate main() sequence
}
```

**Execution flow:**
1. `main()` starts executing
2. `mes(MIDI_TIME)` registers Sequence 0 and returns immediately
3. `PlayMIDI()` starts global MIDI player and returns immediately
4. `del_me` terminates `main()` sequence
5. **Sequence 0 continues running** (independent from main)
6. **MIDI player continues running** (independent from main)
7. MIDI player sends ticks to Sequence 0
8. Sequence 0 executes animation synchronized to MIDI ticks
9. When `step()` block completes, Sequence 0 terminates
10. MIDI continues playing until file ends

### Example 3: sabo2 (PlayMIDI inside step)

```filly
main(){
    LoadPic(...);
    
    mes(TIME) {
        step(2){
            PlayMIDI("SAMPLE.MID");  // Start MIDI inside step
            
            OpenWin( 0 );
            ,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
            MoveWin(0, 1);
            ,,,,,
            MoveWin(0, 2);
            ...
            end_step;
        }
    }
}
```

**Execution flow:**
1. `main()` starts executing
2. `mes(TIME)` registers Sequence 0 and returns immediately
3. `main()` function completes
4. **Sequence 0 continues running** (independent from main)
5. `step()` block starts executing
6. `PlayMIDI()` starts global MIDI player and returns immediately
7. Animation commands execute
8. `end_step` terminates the step block
9. Sequence 0 terminates
10. **MIDI player continues running** (independent from Sequence 0)

### Example 4: robot (MIDI_END event)

```filly
main(){
    mes(TIME){step(20){,start();end_step;del_me;}}
}

start(){
    mes(MIDI_TIME){step{
        // MIDI-synchronized animation
        ...
        end_step; del_me;
    }}
    
    mes(TIME){step(10){
        ...
        PlayMIDI("JMK021.MID");  // Start MIDI
        end_step; del_me;
    }}
    
    mes(MIDI_END){  // Triggered when MIDI finishes
        mes(TIME){step(5){
            message();
            ...
            CloseWinAll();
            end_step; del_all; del_me;
        }}
        del_me;
    }
}
```

**Execution flow:**
1. `main()` registers a TIME sequence that calls `start()`
2. `start()` registers multiple sequences:
   - MIDI_TIME sequence for animation
   - TIME sequence that starts MIDI
   - MIDI_END sequence (waits for MIDI completion)
3. All sequences run concurrently
4. MIDI player starts and runs in background
5. When MIDI finishes, `mes(MIDI_END)` sequence is triggered
6. MIDI_END sequence displays message and calls `del_all`
7. All sequences terminate, program ends

### Example 5: TOKYO.TFY (Sequential MIDI playback)

```filly
main(){
    MIDIFile[1]="saba_n2.mid";
    MIDIFile[2]="saba_n1.mid";
    
    MIDINum=1;
    PlayMIDI(MIDIFile[1]);  // Start first MIDI
    
    mes(MIDI_END){
        if(MIDINum==1){
            MIDINum=2;
            PlayMIDI(MIDIFile[2]);  // Start second MIDI
        }
        else if(MIDINum==2){
            // Launch another program or continue
            ...
        }
    }
}
```

**Execution flow:**
1. First MIDI starts playing
2. `mes(MIDI_END)` sequence waits for MIDI completion
3. When first MIDI finishes, MIDI_END sequence triggers
4. Second MIDI starts playing
5. MIDI_END sequence waits again
6. When second MIDI finishes, MIDI_END sequence triggers again
7. Program continues or terminates

## mes(TIME) and mes(MIDI_TIME) Execution Model

**CRITICAL**: mes() blocks do NOT loop. They execute commands sequentially until reaching the end.

**Execution Model:**
- Each mes() block is a sequence with a program counter (pc)
- Every frame (TIME mode) or MIDI tick (MIDI_TIME mode), the VM executes ONE command from each active sequence
- The pc advances to the next command
- When pc reaches the end of commands, the sequence terminates
- `del_me` can terminate the sequence early

**Example:**
```filly
mes(TIME){
    if(condition1){      // Frame 1: Check condition, execute if true
        MovePic(...);     // Same frame
        i=i+1;            // Same frame
    }                     // pc advances to next command
    
    if(condition2){      // Frame 2: Check next condition
        MovePic(...);
    }
    
    if(done){            // Frame N: Final check
        del_me;           // Terminate if done
    }
    // Frame N+1: pc reaches end, sequence terminates
}
```

**Why this works:**
- Conditional checks happen every frame
- Variables are updated within the same frame
- Next frame checks the updated variables
- Animation progresses frame by frame

## Program Termination

**Automatic Termination:**
When all sequences finish (reach end or call `del_me`) AND all MIDI/audio playback completes, the program automatically terminates:
- In GUI mode: Ebiten window closes
- In headless mode: Execution loop exits
- All resources are cleaned up

**Termination Conditions:**
The program terminates when BOTH conditions are met:
1. All sequences have finished (no active sequences remaining)
2. No MIDI is currently playing (midiPlayer is nil or not playing)

This ensures that:
- MIDI playback is not cut off abruptly
- Programs like y-saru that play MIDI after sequences finish will play the full MIDI file
- Programs like kuma2 that have sequences running during MIDI playback will terminate immediately after sequences finish

**Manual Termination:**
- `del_me` - Terminates the current sequence only
- `del_all` - Terminates all sequences (triggers automatic termination check)
- `ExitTitle()` - Immediately terminates the entire program
- ESC key (GUI mode) - User-initiated termination

**Example 1: Sequences finish before MIDI**
```filly
main(){
    mes(MIDI_TIME){
        step(8){
            // Animation synchronized to MIDI
            ...
            end_step; del_me;  // Sequence finishes
        }
    }
    
    PlayMIDI("FLYINSKY.mid");  // MIDI continues playing
    del_me;  // Main sequence finishes
    
    // Program waits for MIDI to finish before terminating
}
```

**Example 2: MIDI finishes before sequences**
```filly
main(){
    mes(TIME){
        step(65){
            PlayMIDI("kuma.mid");,,  // MIDI starts and finishes
            ...
            CloseWin(0);,
            del_all; del_me;  // Sequences finish after MIDI
        }
    }
    // Program terminates immediately (no MIDI playing)
}
```

## Common Patterns

### Pattern 1: Conditional Animation Loop

```filly
mes(TIME){
    if(flag==1 && counter<100){
        // Perform animation
        counter=counter+1;
    }
    
    if(counter>=100){
        del_me;  // Terminate when done
    }
    
    // Loops back to check conditions again
}
```

### Pattern 2: MIDI-Synchronized Animation

```filly
mes(MIDI_TIME){
    step(8){
        // Commands execute synchronized to MIDI ticks
        MovePic(...);,
        MovePic(...);,
        ...
        end_step;  // Terminates when step completes
    }
}
PlayMIDI("music.mid");
```

### Pattern 3: Event Handler

```filly
mes(MIDI_END){
    // Triggered when MIDI finishes
    ShowMessage("MIDI finished!");
    del_me;
}
```

### Pattern 4: Multiple Concurrent Sequences

```filly
// All three sequences run concurrently
mes(TIME){
    // Sequence 1: Background animation
    ...
}

mes(TIME){
    // Sequence 2: User interface updates
    ...
}

mes(MIDI_TIME){
    // Sequence 3: MIDI-synchronized effects
    ...
}

PlayMIDI("music.mid");
```

## Debugging Tips

### Issue: Sequence terminates too early

**Symptom**: Animation stops before expected

**Cause**: Sequence reached end without explicit termination

**Solution**: 
- mes() blocks execute commands sequentially and terminate when pc reaches end
- Use `del_me` to explicitly terminate when done
- Check for `end_step` in `step()` blocks

### Issue: Program doesn't terminate

**Symptom**: Program continues running after all animations finish

**Cause**: One or more sequences are still active (not terminated)

**Solution**:
- Ensure all sequences call `del_me` when done
- Check that all `step()` blocks have `end_step`
- Use `del_all` to terminate all sequences at once
- Program automatically terminates when all sequences finish

### Issue: MIDI continues after sequences finish

**Symptom**: MIDI playback continues even though all sequences have terminated

**Cause**: This is expected behavior - MIDI player is global and independent from sequences

**Solution**:
- Program will automatically wait for MIDI to finish before terminating
- This ensures MIDI playback is not cut off abruptly
- If you want to terminate immediately, use `ExitTitle()` or ESC key
- Use `StopMIDI()` if you need to stop MIDI before program termination (if implemented)

### Issue: Infinite loop in mes(TIME)

**Symptom**: Program never terminates, continues running indefinitely

**Cause**: mes() block doesn't reach end and doesn't call `del_me`

**Solution**:
- mes() blocks execute sequentially and terminate when pc reaches end
- Add `del_me` when animation/logic is complete
- Use conditional checks to determine when to terminate
- Consider using `step()` blocks with `end_step`
- Program automatically terminates when all sequences finish

### Issue: mes(MIDI_TIME) doesn't execute

**Symptom**: MIDI-synchronized animation doesn't run

**Cause**: `PlayMIDI()` not called, or called after sequence terminates

**Solution**:
- Ensure `PlayMIDI()` is called before or during sequence execution
- Remember: MIDI player is global, can be started from anywhere
- Check that MIDI file exists and loads correctly

## Summary

**Key Takeaways:**

1. **Sequences are independent threads** - they run concurrently and don't block each other

2. **MIDI player is global** - started by `PlayMIDI()`, runs independently from sequences

3. **`del_me` terminates current sequence only** - doesn't affect other sequences or MIDI

4. **mes() blocks execute sequentially** - commands execute one per frame/tick until reaching end

5. **mes(MIDI_TIME) synchronizes to MIDI ticks** - requires MIDI player to be running

6. **mes(MIDI_END) triggers on MIDI completion** - useful for sequential playback

7. **Program auto-terminates** - when all sequences finish AND MIDI playback completes, program automatically terminates

Understanding these concepts is essential for correctly implementing FILLY scripts and avoiding common pitfalls.

## Implementation Details

### Sequence Management

**Internal Structure:**
- `mainSequencer` - The main() function's sequencer (never overwritten)
- `sequencers[]` - Array of all active sequences (including mes() blocks)
- Each sequence has its own program counter (pc), variables, and state

**RegisterSequence() Behavior:**
```go
func RegisterSequence(mode int, ops []OpCode, initialVars ...map[string]any) {
    // Save current sequencer as parent
    parentSeq := mainSequencer
    
    // Create new sequencer for this mes() block
    newSeq := &Sequencer{
        commands: ops,
        active:   true,
        parent:   parentSeq,  // Link to parent for variable scope
        mode:     mode,
    }
    
    // Add to sequencers list (DO NOT overwrite mainSequencer!)
    sequencers = append(sequencers, newSeq)
    
    // Return immediately (non-blocking)
}
```

**CRITICAL**: `RegisterSequence` must NOT overwrite `mainSequencer`. The main() function's sequencer must remain stable throughout execution, otherwise `del_me` calls in main() will incorrectly deactivate mes() blocks instead of main().

### UpdateVM() Behavior

**TIME Mode:**
```go
// Called at 60 FPS
tickCount++
UpdateVM(tickCount)
```

**MIDI_TIME Mode:**
```go
// Called at 60 FPS, but only processes when targetTick advances
currentTarget := atomic.LoadInt64(&targetTick)
for tickCount < currentTarget {
    tickCount++
    UpdateVM(tickCount)
}
```

**Sequence Looping:**
```go
if seq.pc >= len(seq.commands) {
    if seq.mode == Time || seq.mode == MidiTime {
        // Loop back to beginning
        seq.pc = 0
    } else {
        // Other modes - mark as complete
        seq.active = false
    }
}
```

### Headless Mode

In headless mode, the execution loop must replicate the same behavior as the Ebiten game loop:

```go
func runHeadless() {
    ticker := time.NewTicker(time.Second / 60)
    
    for {
        <-ticker.C
        
        // Check for program termination
        if programTerminated {
            break
        }
        
        if !midiSyncMode {
            // TIME MODE
            tickCount++
            UpdateVM(tickCount)
        } else {
            // MIDI_TIME MODE
            currentTarget := atomic.LoadInt64(&targetTick)
            for tickCount < currentTarget {
                tickCount++
                UpdateVM(tickCount)
            }
        }
        
        // Check for termination after VM update
        if programTerminated {
            break
        }
    }
}
```

This ensures:
- MIDI_TIME sequences execute correctly in both GUI and headless modes
- Program terminates automatically when all sequences finish
- No orphaned processes or infinite loops

## References

- Non-Blocking Architecture: `pkg/engine/NON_BLOCKING_ARCHITECTURE.md`
- MIDI Timing: `.kiro/specs/midi-timing-accuracy/`
- Sample Scripts: `samples/yosemiya/`, `samples/y_saru/`, `samples/sabo2/`, `samples/robot/`
