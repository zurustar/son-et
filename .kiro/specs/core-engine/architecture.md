# Architecture Document: son-et Core Engine

**Date:** 2026-01-15  
**Status:** Living Document  
**Verified:** Multiple sample projects

## Table of Contents

1. [System Overview](#system-overview)
2. [Compiler Architecture](#compiler-architecture)
3. [Runtime Architecture](#runtime-architecture)
4. [Concurrency Model](#concurrency-model)
5. [Timing System](#timing-system)
6. [Rendering Pipeline](#rendering-pipeline)
7. [Audio System](#audio-system)
8. [Critical Design Patterns](#critical-design-patterns)
9. [Debugging Guide](#debugging-guide)

## System Overview

The son-et core engine is a source-to-source compiler (transpiler) that converts legacy FILLY scripts into modern Go executables. The system consists of two major components:

1. **Compiler Pipeline:** Transforms FILLY source code into Go source code
2. **Runtime Library:** Provides execution environment for generated code

### High-Level Data Flow

```
User writes game.tfy
        ↓
son-et game.tfy (Transpile)
        ↓
Lexer → Parser → AST → Code Generator
        ↓
game.go (with //go:embed directives)
        ↓
go build game.go
        ↓
Standalone executable with embedded assets
        ↓
User runs ./game
        ↓
Runtime Library (Ebitengine-based)
```

### Directory Structure

```
/
├── cmd/
│   └── son-et/           # Transpiler CLI
│       └── main.go       # Entry point
├── pkg/
│   ├── compiler/         # Compiler Pipeline
│   │   ├── lexer/        # Tokenization
│   │   ├── parser/       # AST Construction
│   │   ├── ast/          # AST Node Definitions
│   │   ├── token/        # Token Definitions
│   │   └── codegen/      # Go Code Generation
│   └── engine/           # Runtime Library
│       ├── engine.go     # Core runtime + graphics
│       └── midi_player.go # Audio system
├── samples/              # Example FILLY projects
└── .kiro/specs/          # Specification documents
```

## Compiler Architecture

### Overview

The compiler follows a traditional multi-pass architecture:

```
Source Code → Lexer → Tokens → Parser → AST → Code Generator → Go Code
```


### Lexer (pkg/compiler/lexer/)

**Responsibility:** Convert source text into tokens

**Key Features:**
- UTF-8 support for Japanese characters
- Comment handling (// and /* */)
- Hexadecimal number literals (0x...)
- String literals with escape sequences
- Line number tracking for error reporting

**Token Types:**
```go
IDENT, NUMBER, STRING
LPAREN, RPAREN, LBRACE, RBRACE, LBRACKET, RBRACKET
PLUS, MINUS, ASTERISK, SLASH
ASSIGN, EQ, NOT_EQ, LT, GT, LTE, GTE
AND, OR, BANG
SEMICOLON, COMMA, COLON
Keywords: IF, ELSE, FOR, MES, STEP, INT, STR
```

**Implementation Notes:**
- Uses rune-based scanning for Unicode support
- Snapshot/Reset mechanism for lookahead
- Whitespace includes ideographic space (0x3000)

### Parser (pkg/compiler/parser/)

**Responsibility:** Build Abstract Syntax Tree from tokens

**Key Features:**
- Pratt parser for expressions (precedence climbing)
- Infinite lookahead for function vs. call disambiguation
- Support for default parameter values
- Implicit type inference for variables

**AST Node Types:**
```go
// Statements
FunctionStatement, LetStatement, AssignStatement
IfStatement, ForStatement, BlockStatement
MesBlockStatement, StepBlockStatement
ExpressionStatement, WaitStatement

// Expressions
Identifier, IntegerLiteral, StringLiteral
CallExpression, InfixExpression, PrefixExpression
IndexExpression
```

**Critical Design Decision: Function vs. Call Disambiguation**

FILLY syntax allows both:
```
Scene1(p1, p2) { ... }  // Function definition
Scene1(p1, p2);         // Function call
```

The parser uses infinite lookahead to distinguish:
```go
func (p *Parser) isFunctionDefinition() bool {
    snapshot := p.l.Snapshot()
    defer p.l.Reset(snapshot)
    
    // Scan past parameter list
    parenCount := 1
    for tok := p.l.NextToken(); tok.Type != EOF; tok = p.l.NextToken() {
        if tok.Type == LPAREN { parenCount++ }
        if tok.Type == RPAREN {
            parenCount--
            if parenCount == 0 {
                nextToken := p.l.NextToken()
                return nextToken.Type == LBRACE  // { means function
            }
        }
    }
    return false
}
```


### Code Generator (pkg/compiler/codegen/)

**Responsibility:** Emit Go source code from AST

**Key Features:**
- Asset detection and go:embed generation
- VM OpCode generation for mes() blocks
- Variable pre-declaration to avoid := issues
- Case-insensitive identifier handling (lowercase conversion)
- User function registration

**Generated Code Structure:**
```go
package main

import (
    "embed"
    "github.com/zurustar/filly2exe/pkg/engine"
)

//go:embed ASSET1.BMP ASSET2.MID
var assets embed.FS

const (
    // Constants from #define
)

var (
    // Global variables
)

func UserFunction(params...) {
    var localVar1 int
    var localVar2 string
    // ... function body
}

func main() {
    engine.RegisterUserFunc("UserFunction", UserFunction)
    engine.Init(assets, func() {
        var localVar1 int
        // ... main script body
    })
    defer engine.Close()
    engine.Run()
}
```

**Critical Design Decision: Variable Pre-Declaration**

Go requires variables to be declared before use. FILLY allows implicit declaration. The code generator scans the AST to find all local variables and pre-declares them:

```go
func (g *Generator) scanLocals(block *ast.BlockStatement) map[string]string {
    locals := map[string]string{}
    // Walk AST and collect all assigned variables
    // Infer types from usage (string literals → string, else → int)
    // Arrays detected by index usage
    return locals
}
```

This avoids Go's `:=` operator which would create shadowed variables.

**Asset Embedding:**

The generator scans for `LoadPic()`, `PlayMIDI()`, and `PlayWAVE()` calls:

```go
func (g *Generator) scanResources(program *ast.Program) []string {
    resources := []string{}
    // Walk AST and find all string literals in asset-loading calls
    // Collect unique filenames
    return resources
}
```

These are emitted as `//go:embed` directives, creating a single-file executable.


## Runtime Architecture

### Overview

The runtime library provides a complete execution environment based on Ebitengine (a 2D game engine for Go). It implements:

1. Graphics system (pictures, windows, casts)
2. Audio system (MIDI, WAV)
3. Timing system (VM sequencer)
4. Concurrency management

### Core Data Structures

```go
// Picture: Image buffer
type Picture struct {
    ID         int
    Image      *ebiten.Image
    BackBuffer *ebiten.Image  // For double buffering
    Width      int
    Height     int
}

// Window: Display region
type Window struct {
    ID         int
    Picture    int  // Picture ID to display
    X, Y       int  // Position on virtual desktop
    W, H       int  // Size
    SrcX, SrcY int  // Picture offset (viewport)
    Visible    bool
    Title      string
    Color      color.Color
}

// Cast: Sprite with transparency
type Cast struct {
    ID          int
    Picture     int  // Source picture
    DestPicture int  // Destination picture
    X, Y        int  // Position
    W, H        int  // Size
    SrcX, SrcY  int  // Source offset (for sprite sheets)
    Transparent color.Color
    Visible     bool
}

// VM Sequencer
type Sequencer struct {
    commands     []OpCode
    pc           int  // Program counter
    waitTicks    int  // Remaining wait ticks
    active       bool
    ticksPerStep int  // Ticks per step() unit
    vars         map[string]any  // VM variables
    mode         int  // 0=TIME, 1=MIDI_TIME
    onComplete   func()
}

type OpCode struct {
    Cmd  string
    Args []any
}
```

### Global State

```go
var (
    // Resource Maps
    pictures      map[int]*Picture
    windows       map[int]*Window
    casts         map[int]*Cast
    
    // Ordering
    windowOrder   []int  // Z-order for windows
    castDrawOrder []int  // Creation order for casts
    
    // ID Counters
    nextPicID  int
    nextWinID  int
    nextCastID int
    
    // Synchronization
    renderMutex sync.Mutex  // Protects all shared state
    
    // VM State
    mainSequencer *Sequencer
    vmLock        sync.Mutex
    tickCount     int64
    targetTick    int64  // Atomic
    
    // Timing Mode
    midiSyncMode bool
    GlobalPPQ    int  // Pulses per quarter note
)
```


## Concurrency Model

### Thread Architecture

The engine uses a **dual-thread architecture** to bridge the gap between FILLY's procedural, blocking nature and Ebitengine's event-driven, non-blocking nature:

```
┌─────────────────────────────────┐     ┌──────────────────────────────┐
│     Main Thread (Ebitengine)    │     │    Script Goroutine          │
├─────────────────────────────────┤     ├──────────────────────────────┤
│ Game.Update() @ 60 FPS          │     │ script() function            │
│   ├─ UpdateVM(currentTick)      │     │   ├─ RegisterSequence()     │
│   └─ Handle input               │     │   ├─ Engine API calls        │
│                                 │     │   └─ Wait() / blocking ops   │
│ Game.Draw() @ 60 FPS            │     │                              │
│   ├─ Acquire renderMutex        │     │ Modifies shared state:       │
│   ├─ Render windows             │     │   ├─ pictures                │
│   ├─ Render casts               │     │   ├─ windows                 │
│   └─ Release renderMutex        │     │   └─ casts                   │
└─────────────────────────────────┘     └──────────────────────────────┘
              │                                      │
              └──────────── renderMutex ────────────┘
```

### Synchronization Strategy

**Problem:** Script goroutine modifies resources (pictures, casts, windows) that the main thread reads for rendering. This creates a race condition.

**Solution:** Global `renderMutex` protects all shared state.

**Lock Acquisition Rules:**

1. **Write Access (Script Goroutine):**
   - All functions that modify state acquire `renderMutex` at entry
   - Examples: `PutCast()`, `MoveCast()`, `OpenWin()`, `TextWrite()`, `LoadPic()`

2. **Read Access (Main Thread):**
   - `Game.Draw()` acquires `renderMutex` for entire frame rendering
   - Ensures consistent snapshot of state

3. **Critical Rule: Avoid Double-Locking**
   - **NEVER** acquire `renderMutex` in `ExecuteOp()` before calling functions
   - Each function handles its own locking
   - Violating this causes deadlock

**Example of Correct Pattern:**

```go
// CORRECT: Let function handle locking
case "MoveCast":
    rArgs := make([]any, len(op.Args))
    for i, a := range op.Args {
        rArgs[i] = ResolveArg(a, seq)
    }
    MoveCast(rArgs...)  // MoveCast acquires renderMutex internally
    return nil, false

// WRONG: Double-locking causes deadlock
case "MoveCast":
    renderMutex.Lock()  // ❌ DON'T DO THIS
    defer renderMutex.Unlock()
    
    rArgs := make([]any, len(op.Args))
    for i, a := range op.Args {
        rArgs[i] = ResolveArg(a, seq)
    }
    MoveCast(rArgs...)  // MoveCast tries to acquire renderMutex → DEADLOCK
    return nil, false
```

### Race Condition Prevention

The engine is designed to be race-free. Verify with:

```bash
go build -race -o game game.go
./game
```

If races are detected, they indicate bugs in lock acquisition patterns.


## Timing System

### Dual-Mode Architecture

The timing system supports two fundamentally different execution modes:

```
┌──────────────────────────────────────────────────────────────┐
│                     Timing System                            │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌─────────────────────┐      ┌──────────────────────────┐ │
│  │   TIME Mode         │      │   MIDI_TIME Mode         │ │
│  │   (Procedural)      │      │   (Music-Synchronized)   │ │
│  ├─────────────────────┤      ├──────────────────────────┤ │
│  │ Driver: Game Loop   │      │ Driver: MIDI Player      │ │
│  │ Rate: 60 FPS        │      │ Rate: Variable (tempo)   │ │
│  │ step(n) = n×50ms    │      │ step(n) = n×32nd note    │ │
│  │ Blocking: YES       │      │ Blocking: NO             │ │
│  └─────────────────────┘      └──────────────────────────┘ │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

### TIME Mode (Frame-Based)

**Use Case:** Procedural animations, UI interactions

**Characteristics:**
- Driven by main game loop at 60 FPS
- `step(n)` defines wait unit as `n × 50ms` (approximately 3 frames)
- `RegisterSequence()` **blocks** until sequence completes
- Sequential execution: mes block → post-mes code

**Implementation:**

```go
func RegisterSequence(mode int, ops []OpCode) {
    if mode != MidiTime {
        wg := &sync.WaitGroup{}
        wg.Add(1)
        
        mainSequencer = &Sequencer{
            commands:   ops,
            mode:       mode,
            onComplete: func() { wg.Done() },
        }
        
        wg.Wait()  // Block until sequence completes
    }
}

func (g *Game) Update() error {
    if !midiSyncMode {
        tickLock.Lock()
        tickCount++
        currentTick := int(tickCount)
        tickLock.Unlock()
        
        UpdateVM(currentTick)  // Execute VM
    }
    return nil
}
```

**Step Calculation:**

```go
case "SetStep":
    if seq.mode == 0 {  // TIME mode
        // 60 FPS → 50ms = 3 ticks
        seq.ticksPerStep = count * 3
    }
```

**Example:**
```
mes(TIME) {
    step(20);  // 20 × 50ms = 1000ms = 1 second
    MoveCast(...);
    ,;  // Wait 1 step = 1 second
    MoveCast(...);
}
// Execution blocks here until mes completes
CloseWin(...);  // Executes after mes finishes
```


### MIDI_TIME Mode (Music-Synchronized)

**Use Case:** Music-synchronized animations, rhythm games

**Characteristics:**
- Driven by MIDI player via `NotifyTick()` callbacks
- `step(n)` defines wait unit as `n × 32nd note` (musical time)
- `RegisterSequence()` **returns immediately** (non-blocking)
- Concurrent execution: mes block || post-mes code

**Critical Requirement:** Must be non-blocking to allow `PlayMIDI()` to execute.

**Why Non-Blocking is Essential:**

```
mes(MIDI_TIME) {
    step(8);
    MoveCast(...);
    ,;  // Wait 1 step
}
PlayMIDI("music.mid");  // ← Must execute to start MIDI player

If RegisterSequence() blocks:
  → PlayMIDI() never executes
  → MIDI player never starts
  → targetTick never updates
  → VM never executes
  → DEADLOCK
```

**Implementation:**

```go
func RegisterSequence(mode int, ops []OpCode) {
    if mode == MidiTime {
        // NO WaitGroup - return immediately
        mainSequencer = &Sequencer{
            commands:   ops,
            mode:       mode,
            onComplete: nil,  // No callback
        }
        // Returns immediately, allowing PlayMIDI() to execute
    }
}

func (g *Game) Update() error {
    if midiSyncMode {
        currentTarget := atomic.LoadInt64(&targetTick)
        
        // Catch up to MIDI time
        tickLock.Lock()
        for tickCount < currentTarget && loops < 10 {
            tickCount++
            currentTick := int(tickCount)
            tickLock.Unlock()
            
            UpdateVM(currentTick)
            
            tickLock.Lock()
            loops++
        }
        tickLock.Unlock()
    }
    return nil
}
```

**Step Calculation:**

```go
case "SetStep":
    if seq.mode == MidiTime {
        // Musical time: 32nd note × n
        // PPQ = 480 (typical)
        // 32nd note = PPQ / 8 = 60 ticks
        // step(8) = 8 × 60 = 480 ticks = quarter note
        seq.ticksPerStep = (GlobalPPQ / 8) * count
    }
```

**Audio Thread Synchronization:**

```go
// Audio thread (MIDI player)
func (s *MidiStream) ProcessSamples(n int) {
    currentSamples += int64(n)
    timeSec := float64(currentSamples) / float64(sampleRate)
    tick := CalculateTickFromTime(timeSec)
    
    if tick > lastTick {
        delta := tick - lastTick
        for i := 0; i < delta; i++ {
            NotifyTick(lastTick + i + 1)
        }
        lastTick = tick
    }
}

// Atomic update (thread-safe)
func NotifyTick(tick int) {
    atomic.StoreInt64(&targetTick, int64(tick))
}
```

**Example:**
```
mes(MIDI_TIME) {
    step(8);  // 8 × 32nd note = quarter note
    MoveCast(...);
    ,;  // Wait 1 step = quarter note
    MoveCast(...);
}
PlayMIDI("music.mid");  // Executes immediately
CloseWin(...);  // Executes immediately (concurrent with mes)
```


### VM Execution Flow

The VM executes OpCodes sequentially until it hits a Wait:

```go
func UpdateVM(currentTick int) {
    vmLock.Lock()
    defer vmLock.Unlock()
    
    if mainSequencer == nil || !mainSequencer.active {
        return
    }
    
    seq := mainSequencer
    
    // Handle Wait
    if seq.waitTicks > 0 {
        seq.waitTicks--
        return  // Yield execution
    }
    
    // Execute instructions until Wait or End
    for seq.pc < len(seq.commands) {
        op := seq.commands[seq.pc]
        seq.pc++
        
        result, yield := ExecuteOp(op, seq)
        
        if yield {
            // Wait command sets seq.waitTicks and returns true
            break
        }
    }
    
    // Check for completion
    if seq.pc >= len(seq.commands) {
        seq.active = false
        if seq.onComplete != nil {
            seq.onComplete()
        }
    }
}
```

**OpCode Execution:**

```go
func ExecuteOp(op OpCode, seq *Sequencer) (any, bool) {
    switch op.Cmd {
    case "Wait":
        steps := ResolveArg(op.Args[0], seq).(int)
        totalTicks := steps * seq.ticksPerStep
        seq.waitTicks = totalTicks
        return nil, true  // Yield
        
    case "SetStep":
        count := ResolveArg(op.Args[0], seq).(int)
        if seq.mode == 0 {  // TIME
            seq.ticksPerStep = count * 3
        } else {  // MIDI_TIME
            seq.ticksPerStep = (GlobalPPQ / 8) * count
        }
        return nil, false  // Continue
        
    case "MoveCast":
        rArgs := resolveArgs(op.Args, seq)
        MoveCast(rArgs...)
        return nil, false  // Continue
        
    // ... other commands
    }
}
```

### Timing Mode Comparison

| Aspect | TIME Mode | MIDI_TIME Mode |
|--------|-----------|----------------|
| **Driver** | Game loop (60 FPS) | MIDI player (audio thread) |
| **Tick Source** | Frame counter | MIDI tick counter |
| **Step Unit** | n × 50ms (3 frames) | n × 32nd note |
| **RegisterSequence** | Blocks (WaitGroup) | Returns immediately |
| **Execution Order** | Sequential (mes → post) | Concurrent (mes \|\| post) |
| **Use Case** | Procedural scripts | Music-synchronized scripts |
| **targetTick Update** | Frame-based increment | Audio callback |

### Common Pitfalls

1. **Making MIDI_TIME blocking** → Deadlock (PlayMIDI never executes)
2. **Making TIME non-blocking** → Breaks sequential logic
3. **Applying TIME logic to MIDI_TIME** → Executes at 60 FPS instead of music tempo
4. **Bootstrapping targetTick in MIDI mode** → Breaks synchronization


## Rendering Pipeline

### Double Buffering

To prevent flickering when casts are redrawn, the engine uses double buffering:

```
┌─────────────────────────────────────────────────────────┐
│                  MoveCast() Flow                        │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  1. Acquire renderMutex                                 │
│  2. Create temporary off-screen buffer (newImg)         │
│  3. Clear buffer (white background)                     │
│  4. Redraw ALL casts in creation order:                 │
│     for _, cID := range castDrawOrder {                 │
│         if cast.DestPicture == targetPic {              │
│             drawWithColorKey(newImg, cast)              │
│         }                                               │
│     }                                                   │
│  5. Atomic swap:                                        │
│     temp = destPic.Image                                │
│     destPic.Image = destPic.BackBuffer                  │
│     destPic.BackBuffer = temp                           │
│  6. Release renderMutex                                 │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

**Why This Works:**

- `Game.Draw()` only sees fully-rendered frames
- No partial rendering artifacts
- Old buffer is reused (no allocation overhead)
- Thread-safe via `renderMutex`

### Z-Ordering (Layering)

**Policy:** Creation Order (Painter's Algorithm)

```go
var castDrawOrder []int  // Global draw order

// PutCast: Add to end of order
func PutCast(args ...any) int {
    castID := nextCastID
    nextCastID++
    
    castDrawOrder = append(castDrawOrder, castID)
    casts[castID] = &Cast{...}
    
    return castID
}

// MoveCast: Iterate in creation order
func MoveCast(args ...any) {
    for _, cID := range castDrawOrder {
        cast := casts[cID]
        if cast.DestPicture == targetPic {
            drawWithColorKey(buffer, cast)
        }
    }
}
```

**Guarantee:** First created = bottom layer, last created = top layer

### Transparency Handling

**Optimized Implementation (Task 8.1 - Completed):**

The engine now uses Ebitengine's native alpha channel for transparency, providing better performance than per-pixel checking during draw operations.

**Pre-Processing Approach:**

Images are processed once during cast creation to convert the transparent color to alpha=0:

```go
func convertTransparentColor(src *ebiten.Image, transparentColor color.Color) *ebiten.Image {
    bounds := src.Bounds()
    width := bounds.Dx()
    height := bounds.Dy()
    
    // Create new image with transparency
    result := ebiten.NewImage(width, height)
    
    // Get RGBA values of transparent color
    tr, tg, tb, _ := transparentColor.RGBA()
    
    // Process pixels: convert matching color to transparent, copy others
    for py := 0; py < height; py++ {
        for px := 0; px < width; px++ {
            c := src.At(px+bounds.Min.X, py+bounds.Min.Y)
            r, g, b, a := c.RGBA()
            
            // If pixel matches transparent color, skip it (leave as transparent)
            if r == tr && g == tg && b == tb {
                continue
            }
            
            // Copy non-transparent pixel
            result.Set(px, py, color.RGBA{
                R: uint8(r >> 8),
                G: uint8(g >> 8),
                B: uint8(b >> 8),
                A: uint8(a >> 8),
            })
        }
    }
    
    return result
}
```

**PutCast Integration:**

When a cast is created, the source image is pre-processed and stored:

```go
func (e *EngineState) PutCast(args ...any) int {
    // ... parse arguments ...
    
    // Get transparent color (from args or top-left pixel)
    var transparentColor color.Color
    if len(args) > 4 {
        // Explicit color from arguments
        transparentColor = parseColor(args[4])
    } else {
        // Default to top-left pixel
        transparentColor = srcPic.Image.At(0, 0)
    }
    
    // Pre-process transparency once
    processedImage := convertTransparentColor(srcPic.Image, transparentColor)
    
    // Store processed image in a new Picture entry
    processedPicID := e.nextPicID
    e.nextPicID++
    e.pictures[processedPicID] = &Picture{
        ID:     processedPicID,
        Image:  processedImage,
        Width:  srcPic.Width,
        Height: srcPic.Height,
    }
    
    // Cast references the processed image
    e.casts[castID] = &Cast{
        ID:          castID,
        Picture:     processedPicID,  // Pre-processed image
        // ...
    }
}
```

**MoveCast Rendering:**

Drawing now uses standard Ebitengine alpha blending:

```go
func (e *EngineState) MoveCast(args ...any) {
    // ... update cast position ...
    
    // Redraw all casts
    for _, cID := range e.castDrawOrder {
        cast := e.casts[cID]
        if cast.Visible && cast.DestPicture == targetPic {
            castSrc := e.pictures[cast.Picture]
            
            // Use native alpha blending (no per-pixel checking)
            opts := &ebiten.DrawImageOptions{}
            opts.GeoM.Translate(float64(cast.X), float64(cast.Y))
            targetImg.DrawImage(imgToDraw, opts)
        }
    }
}
```

**Performance Benefits:**

1. **One-time Processing:** Transparency conversion happens once during cast creation, not on every frame
2. **Native Alpha Blending:** Leverages Ebitengine's optimized GPU-accelerated alpha blending
3. **Reduced CPU Usage:** No per-pixel color comparison during rendering
4. **Better Cache Locality:** Pre-processed images have better memory layout for rendering

**Transparent Color:** Determined by top-left pixel of source image (or explicitly specified in PutCast arguments)

**Legacy drawWithColorKey Function:**

The old `drawWithColorKey` function is retained for reference but no longer used in the rendering pipeline.


### Window Rendering

Windows are rendered with Windows 3.1-style decorations:

```go
func (g *Game) Draw(screen *ebiten.Image) {
    renderMutex.Lock()
    defer renderMutex.Unlock()
    
    // Virtual desktop background (teal)
    screen.Fill(color.RGBA{0x1F, 0x7E, 0x7F, 0xff})
    
    // Render windows in order
    for _, winID := range windowOrder {
        win := windows[winID]
        pic := pictures[win.Picture]
        
        // Draw window frame (gray, 3D borders)
        // Draw title bar (blue)
        // Draw title text (white)
        // Draw window content with clipping
        
        // Calculate intersection of window and picture
        winRect := image.Rect(win.X, win.Y, win.X+win.W, win.Y+win.H)
        imgRect := image.Rect(win.X+win.SrcX, win.Y+win.SrcY, ...)
        drawRect := winRect.Intersect(imgRect)
        
        if !drawRect.Empty() {
            // Draw visible portion
            srcRect := image.Rect(...)
            subImg := pic.Image.SubImage(srcRect)
            screen.DrawImage(subImg, opts)
        }
    }
}
```

**Window Geometry:**
- Content area: User-specified size
- Title bar: 24 pixels
- Border: 4 pixels
- Total size: (W + 8) × (H + 28)

## Audio System

### MIDI Playback

**Components:**
1. **MidiFileSequencer:** Plays MIDI events
2. **Synthesizer:** Converts MIDI to audio (SoundFont-based)
3. **MidiStream:** Pipes audio to Ebitengine

**Flow:**

```
MIDI File (.mid)
    ↓
MidiFileSequencer (meltysynth)
    ↓
Synthesizer (SoundFont .sf2)
    ↓
MidiStream.Read() → PCM samples
    ↓
Ebiten Audio Player
    ↓
System Audio Output
```

**Synchronization:**

```go
func (s *MidiStream) Read(p []byte) (n int, err error) {
    numSamples := len(p) / 4
    
    // Render audio samples
    s.sequencer.Render(s.leftBuf[:numSamples], s.rightBuf[:numSamples])
    
    // Update clock
    s.ProcessSamples(numSamples)
    
    // Convert float32 to int16 bytes
    // ...
}

func (s *MidiStream) ProcessSamples(n int) {
    currentSamples += int64(n)
    timeSec := float64(currentSamples) / float64(sampleRate)
    tick := CalculateTickFromTime(timeSec)
    
    if tick > lastTick {
        delta := tick - lastTick
        for i := 0; i < delta; i++ {
            NotifyTick(lastTick + i + 1)
        }
        lastTick = tick
    }
}
```

**Tempo Map:**

MIDI files can have tempo changes. The engine parses the tempo map:

```go
type TempoEvent struct {
    Tick          int
    MicrosPerBeat int
}

func parseMidiTempo(data []byte) ([]TempoEvent, int, error) {
    // Parse MIDI file
    // Find all Set Tempo meta events (FF 51 03)
    // Build tempo map
    return events, ppq, nil
}
```

**Tick Calculation:**

```go
func CalculateTickFromTime(t float64) int {
    currentTime := 0.0
    currentTick := 0
    
    for i, ev := range globalTempoMap {
        secPerTick := (ev.MicrosPerBeat / 1000000.0) / float64(globalPPQ)
        
        nextEvTick := ...
        durationTicks := nextEvTick - ev.Tick
        segmentTime := float64(durationTicks) * secPerTick
        
        if currentTime + segmentTime >= t {
            remainTime := t - currentTime
            remainTicks := remainTime / secPerTick
            return ev.Tick + int(remainTicks)
        }
        
        currentTime += segmentTime
        currentTick = nextEvTick
    }
    
    return currentTick
}
```

### WAV Playback

Simple PCM audio playback:

```go
func PlayWAVE(path string) {
    data := findAsset(path)
    stream := wav.DecodeWithSampleRate(sampleRate, bytes.NewReader(data))
    player := audioContext.NewPlayer(stream)
    player.Play()
}
```

Multiple WAV files can play concurrently.


## Critical Design Patterns

### 1. Avoid Double-Locking

**Problem:** Acquiring `renderMutex` twice causes deadlock.

**Pattern:**
```go
// ✅ CORRECT
case "MoveCast":
    rArgs := resolveArgs(op.Args, seq)
    MoveCast(rArgs...)  // Function handles locking

// ❌ WRONG
case "MoveCast":
    renderMutex.Lock()  // First lock
    defer renderMutex.Unlock()
    rArgs := resolveArgs(op.Args, seq)
    MoveCast(rArgs...)  // Second lock → DEADLOCK
```

**Rule:** Let each function handle its own locking. Never lock in `ExecuteOp()`.

### 2. Non-Blocking MIDI_TIME

**Problem:** Blocking in MIDI_TIME mode prevents `PlayMIDI()` from executing.

**Pattern:**
```go
// ✅ CORRECT
func RegisterSequence(mode int, ops []OpCode) {
    if mode == MidiTime {
        mainSequencer = &Sequencer{...}
        // Return immediately
    } else {
        wg := &sync.WaitGroup{}
        wg.Add(1)
        mainSequencer = &Sequencer{
            onComplete: func() { wg.Done() },
        }
        wg.Wait()  // Block for TIME mode only
    }
}

// ❌ WRONG
func RegisterSequence(mode int, ops []OpCode) {
    wg := &sync.WaitGroup{}
    wg.Add(1)
    mainSequencer = &Sequencer{
        onComplete: func() { wg.Done() },
    }
    wg.Wait()  // Blocks for all modes → DEADLOCK in MIDI_TIME
}
```

### 3. Creation Order Z-Ordering

**Problem:** Map iteration order is random in Go.

**Pattern:**
```go
// ✅ CORRECT
var castDrawOrder []int  // Explicit order

func PutCast(...) int {
    castDrawOrder = append(castDrawOrder, castID)
    return castID
}

func MoveCast(...) {
    for _, cID := range castDrawOrder {  // Iterate in order
        cast := casts[cID]
        // Draw cast
    }
}

// ❌ WRONG
func MoveCast(...) {
    for cID, cast := range casts {  // Random order
        // Draw cast
    }
}
```

### 4. Double Buffering

**Problem:** Drawing directly to displayed image causes flickering.

**Pattern:**
```go
// ✅ CORRECT
func MoveCast(...) {
    renderMutex.Lock()
    defer renderMutex.Unlock()
    
    // Draw to back buffer
    targetImg := destPic.BackBuffer
    targetImg.Clear()
    for _, cast := range casts {
        drawWithColorKey(targetImg, cast)
    }
    
    // Atomic swap
    temp := destPic.Image
    destPic.Image = destPic.BackBuffer
    destPic.BackBuffer = temp
}

// ❌ WRONG
func MoveCast(...) {
    // Draw directly to displayed image
    destPic.Image.Clear()  // User sees this!
    for _, cast := range casts {
        drawWithColorKey(destPic.Image, cast)  // Partial rendering visible
    }
}
```

## Debugging Guide

### Debug Levels

Set via environment variable:

```bash
DEBUG_LEVEL=0  # Errors only
DEBUG_LEVEL=1  # Important operations (default)
DEBUG_LEVEL=2  # All debug output (verbose)
```

**Example:**
```bash
DEBUG_LEVEL=2 ./game 2>&1 | tee debug.log
```

### Common Issues

#### 1. Images Not Loading

**Symptoms:** Black screen, "Picture ID not found" errors

**Checks:**
- Assets in same directory as generated Go file?
- `//go:embed` directives in generated code?
- Case-insensitive filename matching working?

**Debug:**
```bash
grep "//go:embed" game.go
ls -la *.BMP *.bmp
```

#### 2. MIDI Not Playing

**Symptoms:** No audio, "No SoundFont loaded" error

**Checks:**
- SoundFont file (`.sf2`) exists?
- MIDI file embedded correctly?
- `PlayMIDI()` called after `mes(MIDI_TIME)` block?

**Debug:**
```bash
ls -la *.sf2
./game -sf path/to/soundfont.sf2
```

#### 3. Deadlock or Freeze

**Symptoms:** Application hangs, no response

**Checks:**
- Mutex double-locking? (see pattern above)
- MIDI_TIME mode is non-blocking?
- TIME mode is blocking correctly?

**Debug:**
```bash
# Run with race detector
go build -race -o game game.go
./game

# Check goroutine stack traces
kill -QUIT <pid>  # Sends SIGQUIT, prints stack traces
```

#### 4. Flickering Graphics

**Symptoms:** Sprites flicker, partial rendering visible

**Checks:**
- Double buffering implemented?
- `renderMutex` held during entire draw operation?

**Debug:**
```bash
# Enable verbose logging
DEBUG_LEVEL=2 ./game 2>&1 | grep "MoveCast\|Draw"
```

#### 5. Timing Issues

**Symptoms:** Animations too fast/slow, not synchronized with music

**Checks:**
- Correct timing mode (TIME vs MIDI_TIME)?
- `step()` value appropriate?
- `targetTick` updating correctly?

**Debug:**
```bash
DEBUG_LEVEL=2 ./game 2>&1 | grep "VM:\|Tick\|SetStep"
```

### Performance Profiling

**CPU Profiling:**
```bash
go build -o game game.go
./game &
PID=$!
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/profile?seconds=30
```

**Memory Profiling:**
```bash
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/heap
```

**Race Detection:**
```bash
go build -race -o game game.go
./game
# Any races will be reported to stderr
```

### Logging Best Practices

**Add Timestamps:**
```bash
DEBUG_LEVEL=2 ./game 2>&1 | while IFS= read -r line; do 
  echo "$(date '+%H:%M:%S.%3N') $line"
done | tee debug.log
```

**Filter Specific Operations:**
```bash
DEBUG_LEVEL=2 ./game 2>&1 | grep "MoveCast"
DEBUG_LEVEL=2 ./game 2>&1 | grep "VM:"
DEBUG_LEVEL=2 ./game 2>&1 | grep "PERF:"
```

## Conclusion

The son-et core engine is a well-architected system with clear separation of concerns, proper thread safety, and efficient rendering. The dual-mode timing system correctly handles both procedural and music-synchronized execution.

Key architectural decisions:
- **Dual-thread model** bridges procedural FILLY and event-driven Ebitengine
- **Render mutex** ensures thread-safe access to shared state
- **Double buffering** prevents flickering
- **Creation order z-ordering** ensures consistent layering
- **Non-blocking MIDI_TIME** allows music synchronization

The implementation is production-ready for verified sample projects and provides a solid foundation for supporting additional FILLY features and samples.
