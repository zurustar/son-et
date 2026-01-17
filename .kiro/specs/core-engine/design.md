# Design Document: Core Engine

## Overview

The son-et core engine is a runtime library that provides the execution environment for FILLY scripts. The system consists of:

1. **Runtime Library** - An Ebitengine-based execution environment providing graphics, audio, timing, and script execution
2. **VM Execution Engine** - OpCode interpreter that executes FILLY scripts

The design prioritizes correctness, thread safety, and cross-platform compatibility while maintaining compatibility with legacy FILLY script semantics.

This specification focuses on the runtime engine functionality. Common design elements (OpCode structure, VM architecture, concurrency model, etc.) are defined in [COMMON_DESIGN.md](../COMMON_DESIGN.md). Interpreter architecture is defined in [interpreter-architecture/design.md](../interpreter-architecture/design.md).

## Common Design Elements

This specification depends on the following common design elements defined in [COMMON_DESIGN.md](../COMMON_DESIGN.md):
- OpCode Structure
- Variable Scope Model
- Concurrency Model
- Rendering Pipeline
- Timing & Synchronization
- Virtual Display Architecture
- Data Models
- Error Handling
- Critical Design Constraints

## Architecture

## Architecture

### System Components

#### The Runtime Library (`pkg/engine`)

The Go package that provides the game loop and runtime environment.

**Core Functions:**
- `engine.Init()`: Sets up the Ebitengine window (1280x720 Virtual Desktop) and initializes Audio
- `engine.Run()`: Starts the Game Loop
- `engine.UpdateVM()`: Executes the VM OpCodes for timing-critical sequences
- `engine.OpenWin/MoveCast`: Manages the Virtual Windows and Sprite rendering

### Directory Structure

```
/
├── cmd/
│   └── son-et/        # The Interpreter CLI Entrypoint
├── pkg/
│   ├── compiler/      # Interpreter Logic
│   │   ├── lexer/
│   │   ├── parser/
│   │   └── ast/
│   └── engine/        # Runtime Library (Ebitengine wrapper + VM)
├── samples/           # Example projects
└── docs/              # Documentation
```

## Components and Interfaces

### Picture Manager

**Responsibilities:**
- Load BMP files from embedded assets
- Create empty image buffers
- Copy pixels between pictures with optional transparency
- Track picture dimensions
- Release picture resources

**Key Functions:**
- `LoadPic(filename)` → Picture ID
- `CreatePic(width, height)` → Picture ID
- `MovePic(src, srcX, srcY, w, h, dst, dstX, dstY, mode)`
- `DelPic(picID)`
- `PicWidth(picID)`, `PicHeight(picID)`

### Cast Manager

**Responsibilities:**
- Create sprites with transparency
- Track creation order for z-ordering
- Update sprite positions and re-render
- Support sprite sheet clipping
- Manage cast lifecycle

**Key Functions:**
- `PutCast(srcPic, dstPic, x, y, transparentColor, ..., w, h, srcX, srcY)` → Cast ID
- `MoveCast(castID, pic, x, y, w, h, srcX, srcY)`
- `DelCast(castID)`

**Data Structure:**
```go
type Cast struct {
    SrcPicID int
    DstPicID int
    X, Y     int
    W, H     int
    SrcX, SrcY int
    TransparentColor color.Color
}

var casts map[int]*Cast
var castDrawOrder []int  // Maintains creation order
```

**CRITICAL: Cast Transparency Implementation**

**Design Principle: Separation of Concerns**
- **Picture (Pic)**: Raw image data loaded from files. Never modified after loading.
- **Cast**: A sprite that references a Picture, with optional transparency processing.

**Transparency Processing Strategy:**

When `PutCast` is called with a transparent color parameter:

1. **One-Time Processing at Cast Creation:**
   - Create a NEW transparency-processed image using `convertTransparentColor()`
   - Store this processed image as a NEW Picture with a new Picture ID
   - The Cast references this processed Picture ID, NOT the original
   - This happens ONCE when the Cast is created

2. **Drawing (PutCast and MoveCast):**
   - Simply draw the transparency-processed Picture
   - NO additional transparency processing needed
   - Ebitengine's native alpha blending handles the transparency automatically

3. **Performance Considerations:**
   - Transparency processing is expensive (loops through all pixels)
   - MUST be done only ONCE at Cast creation time
   - NEVER process transparency on every draw call
   - The processed image is reused for all subsequent draws

**Example Flow:**
```
LoadPic("sprite.bmp") → Picture ID 17 (original image)
PutCast(17, dest, x, y, 0xffffff, ...) → 
  1. convertTransparentColor(Picture 17, white) → new image
  2. Store as Picture ID 28 (transparency-processed)
  3. Create Cast ID 2 with Picture=28
  4. Draw Picture 28 to destination
MoveCast(2, ...) →
  1. Cast 2 references Picture 28 (already processed)
  2. Draw Picture 28 to destination (no processing needed)
```

**Common Mistakes to Avoid:**
- ❌ Processing transparency on every draw call (performance issue)
- ❌ Modifying the original Picture (violates separation of concerns)
- ❌ Storing transparency info in Cast but processing at draw time (inefficient)
- ✅ Process transparency ONCE at Cast creation, store as new Picture
- ✅ Keep original Picture unchanged
- ✅ Draw the processed Picture directly (no additional processing)

### Window Manager

**Responsibilities:**
- Create virtual windows within the desktop
- Update window properties (position, size, picture)
- Manage window lifecycle
- Set window captions

**Key Functions:**
- `OpenWin(pic, x, y, width, height, picX, picY, bgColor)` → Window ID
- `MoveWin(winID, pic, x, y, width, height, picX, picY)`
- `CloseWin(winID)`
- `CloseWinAll()`
- `CapTitle(winID, title)`

### Text Renderer

**Responsibilities:**
- Load fonts with specified attributes
- Render text to pictures
- Manage text color and background
- Support Japanese fonts

**Key Functions:**
- `SetFont(size, fontName, charset, ...)`
- `TextWrite(text, picID, x, y)`
- `TextColor(r, g, b)`
- `BgColor(r, g, b)`
- `BackMode(mode)` - 0: opaque, 1: transparent

### Timing System (VM)

**Responsibilities:**
- Execute timing-synchronized commands
- Manage MIDI sync and time mode
- Handle Wait operations
- Drive mes block execution

**Key Functions:**
- `RegisterSequence(mode, operations)` - Queue operations for VM
- `UpdateVM()` - Called each frame to advance VM state
- `SetStep(count)` - Set step resolution
- `Wait(steps)` - Wait for specified steps

**Data Structures:**
```go
type Operation struct {
    Command string
    Args    []interface{}
}

type Sequence struct {
    Operations []Operation
    PC         int  // Program Counter
    StepSize   int
    Mode       int  // TIME or MIDI_TIME
}
```

## Data Models

### Picture

```go
type Picture struct {
    ID    int
    Image *ebiten.Image
    Width int
    Height int
}
```

### Cast

```go
type Cast struct {
    ID               int
    SrcPicID         int
    DstPicID         int
    X, Y             int
    W, H             int
    SrcX, SrcY       int
    TransparentColor color.Color
}
```

### Window

```go
type Window struct {
    ID       int
    PicID    int
    X, Y     int
    Width    int
    Height   int
    PicX, PicY int
    BgColor  color.Color
    Caption  string
}
```

### VM Sequence

```go
type Sequence struct {
    ID         int
    Operations []Operation
    PC         int
    StepSize   int
    Mode       int
    TickCount  int
    WaitUntil  int
    Completed  bool
}
```

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system—essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: Sequential ID assignment

*For any* sequence of LoadPic or PutCast calls, the returned IDs should be sequential starting from 0 (0, 1, 2, ...).

**Validates: Requirements 1.2 (Picture Management), 1.2 (Cast Management)**

### Property 2: Window creation and properties

*For any* valid OpenWin parameters (picture, position, size), a window should be created with exactly those properties.

**Validates: Requirements 1.2 (Virtual Display Architecture)**

### Property 3: Creation order rendering

*For any* sequence of window or cast creations, the rendering order should match the creation order (first created = bottom layer, last created = top layer).

**Validates: Requirements 1.3 (Virtual Display Architecture)**

### Property 4: Resource cleanup on deletion

*For any* picture, cast, or window, after calling the corresponding Del function, that resource ID should no longer be valid and should not appear in subsequent operations.

**Validates: Requirements 1.5 (Picture Management), 1.7 (Cast Management)**

### Property 5: Picture dimension queries

*For any* loaded or created picture, PicWidth and PicHeight should return the exact dimensions of that picture.

**Validates: Requirements 1.6 (Picture Management)**

### Property 6: Sprite clipping correctness

*For any* cast with SubImage clipping parameters (srcX, srcY, w, h), the rendered sprite should display only the specified region of the source picture.

**Validates: Requirements 1.4 (Cast Management)**

### Property 7: Transparency key behavior

*For any* sprite, pixels matching the top-left pixel color (or specified transparent color) should not be rendered (transparent).

**Validates: Requirements 1.6 (Cast Management)**

### Property 8: MIDI Sync Mode non-blocking

*For any* mes(MIDI_TIME) block, RegisterSequence should return immediately without blocking, allowing subsequent code (like PlayMIDI) to execute.

**Validates: Requirements 10.5 (MIDI Sync Mode Timing)**

### Property 9: Time Mode blocking

*For any* mes(TIME) block, RegisterSequence should block until the sequence completes, ensuring sequential execution.

**Validates: Requirements 11.5, 11.6 (Time Mode Timing)**

### Property 10: Step interpretation in MIDI Sync Mode

*For any* step(n) call in MIDI Sync Mode, the wait duration should be n * (32nd note duration based on MIDI tempo).

**Validates: Requirements 10.3 (MIDI Sync Mode Timing)**

### Property 11: Step interpretation in Time Mode

*For any* step(n) call in Time Mode, the wait duration should be approximately n * 50ms (±frame time tolerance).

**Validates: Requirements 11.3 (Time Mode Timing)**

### Property 12: String operation correctness

*For any* string s, StrLen(s) should return the length, SubStr(s, pos, len) should return the substring from pos with length len, and StrFind(s, search) should return the first index of search in s (or -1 if not found).

**Validates: Requirements 13.1, 13.2, 13.3 (String Operations)**

### Property 13: String formatting correctness

*For any* format string and values, StrPrint should produce output matching the format specifiers (%s for strings, %ld for decimal integers, %lx for hexadecimal integers).

**Validates: Requirements 13.4, 13.5 (String Operations)**

### Property 14: Random number range

*For any* positive integer max, Random(max) should return a value in the range [0, max).

**Validates: Requirements 16.4 (System Information)**

### Property 15: Message system state tracking

*For any* mes block execution, GetMesNo(0) should return the last executed block number, and GetMesNo(1) should return the currently executing block number.

**Validates: Requirements 15.1, 15.2 (Message System)**

### Property 16: Script termination cleanup

*For any* script instance, after calling DelMe, DelUs, or DelAll, all associated resources (pictures, casts, windows) should be released and the script should stop executing.

**Validates: Requirements 17.1, 17.2, 17.3, 17.4 (Script Lifecycle)**

### Property 17: Text rendering state persistence

*For any* sequence of TextColor, BgColor, and BackMode calls, subsequent TextWrite operations should use the most recently set values.

**Validates: Requirements 12.3, 12.4, 12.5 (Text Rendering)**

### Property 18: Window state updates

*For any* window, after calling MoveWin with new parameters, the window should have exactly those new properties.

**Validates: Requirements 14.3 (Window Management)**

### Property 19: Window caption updates

*For any* window, after calling CapTitle with a string, the window's caption should be that exact string.

**Validates: Requirements 14.6 (Window Management)**

### Property 20: Multiple WAV concurrent playback

*For any* set of WAV files, calling PlayWAVE on each should result in all of them playing simultaneously without interference.

**Validates: Requirements 9.5 (WAVE Audio Playback)**

### Property 21: MIDI playback single iteration

*For any* MIDI file, calling PlayMIDI should play the file exactly once and then stop (no looping).

**Validates: Requirements 8.4 (MIDI Playback and Synthesis)**

## Error Handling

See [COMMON_DESIGN.md](../COMMON_DESIGN.md) for common error handling strategies.

### Runtime Errors Specific to Core Engine

### Runtime Errors Specific to Core Engine

**Audio Initialization Failures:**
- When MIDI player or audio system fails to initialize
- Action: Log error, disable audio features, continue execution

**Font Loading Failures:**
- When SetFont fails to load the specified font
- Action: Log error, use default font, continue execution

**Picture Format Errors:**
- When LoadPic encounters an unsupported or corrupted image format
- Action: Log error, return invalid ID (-1), continue execution

## Testing Strategy

See [COMMON_DESIGN.md](../COMMON_DESIGN.md) for common testing strategies.

### Unit Testing Focus Areas Specific to Core Engine

**Runtime:**
- Picture loading and manipulation
- Cast creation and z-ordering
- Window management lifecycle
- Text rendering with various fonts
- String operations edge cases

**Audio:**
- MIDI file loading and playback
- WAV file decoding and playback
- Concurrent audio playback

**Platform Testing:**
- Verify macOS CoreAudio integration
- Test on different macOS versions

### Performance Testing

**Benchmarks:**
- Rendering performance with many casts
- Audio synthesis latency
- VM execution overhead

**Profiling:**
- CPU profiling for hot paths
- Memory profiling for resource usage
- Goroutine profiling for concurrency

## Implementation Notes

See [COMMON_DESIGN.md](../COMMON_DESIGN.md) for critical design constraints and common mistakes to avoid.

### Development Workflow

**Debugging Checklist:**
- [ ] Enable DEBUG_LEVEL=2 for detailed logging
- [ ] Use timestamped logs for timing issues
- [ ] Check for mutex double-locking patterns
- [ ] Verify timing mode (MIDI_TIME vs TIME)
- [ ] Run with `-race` flag for concurrency issues

### Future Enhancements

**Potential Improvements:**
- Support for additional image formats (PNG, JPEG)
- Hardware-accelerated rendering
- More sophisticated audio mixing
- Visual editor for sprite placement
- Hot-reload for faster development iteration


## Critical Implementation Details

### MoveWin with Variable Arguments

**IMPORTANT:** `MoveWin` can be called with different numbers of arguments, and the behavior must adapt accordingly.

**2-Argument Form: MoveWin(winID, picID)**

When called with only 2 arguments, the function should:
1. Keep the current window position
2. Use the NEW picture's dimensions (not hardcoded values)
3. Reset source offsets to (0, 0)

**Correct Implementation:**
```go
case "movewin":
    if len(args) >= 2 {
        winID, ok1 := args[0].(int)
        picID, ok2 := args[1].(int)
        if ok1 && ok2 {
            // Get current window and new picture
            var win *Window
            var pic *Picture
            if globalEngine != nil {
                globalEngine.renderMutex.Lock()
                win, ok = globalEngine.windows[winID]
                if ok {
                    pic, _ = globalEngine.pictures[picID]
                }
                globalEngine.renderMutex.Unlock()
            }
            
            if ok && pic != nil {
                x := win.X - BorderThickness
                y := win.Y - TitleBarHeight - BorderThickness
                // Use NEW picture's size (CRITICAL)
                w := pic.Width
                h := pic.Height
                srcX := 0
                srcY := 0
                MoveWin(winID, picID, x, y, w, h, srcX, srcY)
            }
        }
    }
```

**WRONG Approach (DO NOT USE):**
```go
// WRONG: Hardcoded dimensions
MoveWin(winID, picID, 0, 0, 640, 480, 0, 0)  // Causes incorrect window sizes
```

**Why This Matters:**
- Different pictures have different sizes
- Hardcoded sizes cause visual glitches
- Window should always match the picture being displayed

### Thread Safety in Audio Functions

**IMPORTANT:** Functions called from `ExecuteOp` must not acquire `vmLock` as it's already held by `UpdateVM`.

**PlayMIDI Correct Implementation:**
```go
func PlayMIDI(args ...any) {
    // ... validation ...
    
    PlayMidiFile(path)
    
    // NOTE: We are already inside vmLock from UpdateVM, so don't lock again
    tickCount = 0
    atomic.StoreInt64(&targetTick, 0)
    
    if mainSequencer != nil {
        mainSequencer.active = true
    }
    
    StartQueuedCallback()
}
```

**WRONG Approach (DO NOT USE):**
```go
// WRONG: Causes deadlock
func PlayMIDI(args ...any) {
    PlayMidiFile(path)
    
    vmLock.Lock()  // DEADLOCK: UpdateVM already holds this lock
    tickCount = 0
    vmLock.Unlock()
}
```

### Asynchronous Audio Playback

**IMPORTANT:** Audio playback operations should not block the main thread.

**Correct Implementation:**
```go
midiPlayer, err = audioContext.NewPlayer(stream)
if err != nil {
    return
}

// Start playback in a goroutine to avoid blocking
go func() {
    midiPlayer.Play()
    fmt.Println("PlayMIDI: Playback started")
}()
```

**Why This Matters:**
- Blocking audio operations freeze the game loop
- Ebiten's main thread must remain responsive
- Audio should run independently of the VM

### ExecuteOpDirect for Top-Level Execution

**Purpose:** Execute OpCodes directly without sequence wrapping.

**When to Use:**
- Executing `main()` function body
- Top-level script initialization
- Any code that should run immediately without timing control

**Implementation:**
```go
func ExecuteOpDirect(op OpCode) {
    // Create a dummy sequencer for variable scope
    dummySeq := &Sequencer{
        vars: make(map[string]any),
        mode: Time,
    }
    ExecuteOp(op, dummySeq)
}
```

**Why This Exists:**
- Avoids nested sequence deadlocks
- Allows `mes()` blocks to register their own sequences
- Provides variable scope without sequence overhead

## Debugging Common Issues

**Note:** For high-level overview of common issues, see [COMMON_DESIGN.md](../COMMON_DESIGN.md#lessons-learned-from-implementation)

### Issue: Window Not Displaying

**Symptoms:**
- Audio plays but no window appears
- Program seems to run but nothing visible

**Likely Causes:**
1. Game loop blocked (check for deadlocks)
2. `RegisterSequence` blocking in wrong context
3. `main()` wrapped in sequence (should use `ExecuteOpDirect`)

**Solution:** Ensure `main()` executes directly, not in a sequence.

### Issue: Images Wrong Size

**Symptoms:**
- First image correct, subsequent images wrong size
- Window size doesn't match picture

**Likely Cause:** `MoveWin` using hardcoded dimensions instead of picture size.

**Solution:** Look up new picture dimensions dynamically.

### Issue: Program Freezes After PlayMIDI

**Symptoms:**
- MIDI starts playing
- Game loop stops
- No further updates

**Likely Causes:**
1. `PlayMIDI` acquiring `vmLock` (deadlock)
2. `midiPlayer.Play()` blocking main thread

**Solution:** Remove `vmLock` from `PlayMIDI`, use goroutine for playback.
