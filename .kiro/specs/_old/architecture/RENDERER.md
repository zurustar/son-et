# Renderer Architecture

## Overview

The rendering logic has been separated from state management to enable:
- **Headless testing** without Ebitengine initialization
- **Mock implementations** for unit tests
- **Clear separation of concerns** between state and rendering

## Architecture

### Renderer Interface

```go
type Renderer interface {
    RenderFrame(screen *ebiten.Image, state *EngineState)
}
```

The `Renderer` interface abstracts all rendering operations. It reads from `EngineState` but does not modify it.

### Implementations

#### EbitenRenderer

The production renderer that uses Ebitengine to draw to the screen:

```go
renderer := NewEbitenRenderer()
```

Features:
- Renders windows with Windows 3.1 style decorations
- Handles clipping and viewport transformations
- Supports debug overlays
- Thread-safe (acquires renderMutex during rendering)

#### MockRenderer

A no-op renderer for testing:

```go
mockRenderer := NewMockRenderer()
```

Features:
- Records render calls without actually rendering
- Captures state for verification
- Enables headless testing
- No Ebitengine initialization required

## Usage

### Production Code

```go
// Create engine state with default Ebitengine renderer
state := NewEngineState()

// The renderer is automatically set to EbitenRenderer
// Rendering happens in Game.Draw()
```

### Testing Code

```go
// Create engine state with mock renderer
mockRenderer := NewMockRenderer()
state := NewEngineState(WithRenderer(mockRenderer))

// Perform state operations
state.pictures[0] = &Picture{...}

// Simulate rendering (no actual drawing)
mockRenderer.RenderFrame(nil, state)

// Verify rendering occurred
if mockRenderer.RenderCount != 1 {
    t.Error("Expected render to be called")
}
```

## Benefits

### 1. Headless Testing

Tests can run without initializing Ebitengine or creating a window:

```go
func TestStateOperations(t *testing.T) {
    mockRenderer := NewMockRenderer()
    state := NewEngineState(WithRenderer(mockRenderer))
    
    // Test state operations without rendering
    state.LoadPic("test.bmp")
    state.CreatePic(640, 480)
    
    // No Ebitengine initialization needed!
}
```

### 2. Clear Separation of Concerns

- **EngineState**: Manages game state (pictures, windows, casts)
- **Renderer**: Reads state and draws to screen
- **Game**: Coordinates updates and rendering

### 3. Testability

Mock renderer allows verification of:
- Render call count
- State at render time
- Rendering behavior without visual inspection

## Implementation Details

### Thread Safety

The renderer acquires `renderMutex` when reading state:

```go
func (r *EbitenRenderer) RenderFrame(screen *ebiten.Image, state *EngineState) {
    state.renderMutex.Lock()
    defer state.renderMutex.Unlock()
    
    // Read state and render...
}
```

This ensures thread-safe access when the script goroutine modifies state.

### State Immutability During Rendering

The renderer only reads from `EngineState` - it never modifies it. All state modifications happen through EngineState methods that acquire the mutex.

## Migration Path

The current implementation maintains backward compatibility:
- Global state still exists for legacy code
- New code should use EngineState methods
- Renderer works with EngineState, not global state

Future work will complete the migration to EngineState-only architecture.

## Examples

### Example 1: Basic Rendering Test

```go
func TestBasicRendering(t *testing.T) {
    mockRenderer := NewMockRenderer()
    state := NewEngineState(WithRenderer(mockRenderer))
    
    // Create a picture
    pic := &Picture{
        ID:     0,
        Image:  ebiten.NewImage(100, 100),
        Width:  100,
        Height: 100,
    }
    state.pictures[0] = pic
    
    // Render
    mockRenderer.RenderFrame(nil, state)
    
    // Verify
    if mockRenderer.RenderCount != 1 {
        t.Error("Expected one render call")
    }
}
```

### Example 2: State Isolation Test

```go
func TestStateIsolation(t *testing.T) {
    mockRenderer := NewMockRenderer()
    state := NewEngineState(WithRenderer(mockRenderer))
    
    // Modify state
    state.pictures[0] = &Picture{...}
    
    // Verify no rendering occurred
    if mockRenderer.RenderCount != 0 {
        t.Error("State modification should not trigger rendering")
    }
    
    // Explicit render
    mockRenderer.RenderFrame(nil, state)
    
    // Now rendering occurred
    if mockRenderer.RenderCount != 1 {
        t.Error("Expected render after explicit call")
    }
}
```

## Future Enhancements

Potential improvements:
- Add more renderer implementations (e.g., headless image capture)
- Support render-to-texture for testing
- Add performance profiling hooks
- Support custom rendering pipelines
