# Testing Strategy for son-et

## Overview

This document describes the testing strategy for son-et, particularly regarding the challenges of testing Ebiten-based graphics code.

## Testing Challenges

### Ebiten Initialization Requirement

Ebiten requires a display/window context to be initialized before creating images. This makes traditional unit testing difficult:

```go
// ❌ This will panic in test environment
func TestSomething(t *testing.T) {
    img := ebiten.NewImage(100, 100)  // Panic: no display context
}
```

**Error**: `panic: runtime error: invalid memory address or nil pointer dereference`

This occurs because:
1. Ebiten needs to initialize OpenGL/Metal context
2. Tests run in headless CI environments without displays
3. `ebiten.NewImage()` requires the game loop to be running

### Current Testing Approach

We use **integration tests** instead of unit tests for graphics code:

1. **Integration Tests** (samples/*/):
   - Test complete TFY scripts (kuma2, y_saru, etc.)
   - Run actual game loop with real Ebiten context
   - Verify end-to-end functionality
   - Can run in GUI mode for visual verification

2. **Unit Tests** (for non-graphics code):
   - Lexer, parser, compiler tests
   - VM execution tests (without graphics)
   - Array operations, string operations
   - Sequencer, event system

## Test Coverage

### ✅ Well-Tested Components

- **Compiler Pipeline**: lexer, parser, codegen, preprocessor
- **VM Execution**: opcode execution, variable scoping, control flow
- **Timing System**: step-based execution, dual timing modes
- **Event System**: mes() blocks, event handlers
- **Audio System**: MIDI playback, WAV playback
- **Asset Loading**: filesystem and embedded FS loaders

### ⚠️ Integration-Only Testing

These components are tested only through integration tests:

- **Picture Operations**: LoadPicture, CreatePicture, MovePicture, MoveSPicture, ReversePicture
- **Cast Operations**: PutCast, MoveCast, cast transparency, double buffering
- **Window Operations**: OpenWin, MoveWin, window dragging
- **Drawing Operations**: Line, Circle, Rectangle drawing
- **Text Rendering**: TextWrite, font loading
- **Renderer**: Frame rendering, window decorations

### ❌ Not Tested

- Performance benchmarks (frame rate, memory usage)
- Stress tests (many windows, many casts)
- Edge cases in graphics operations

## Running Tests

### Unit Tests Only

```bash
# Run all unit tests (excludes integration tests)
go test ./pkg/... -short

# Run specific package
go test ./pkg/compiler/parser -v
```

### Integration Tests

```bash
# Run integration tests (requires display)
go test ./pkg/engine -run Integration -v

# Run specific integration test
go test ./pkg/engine -run TestKuma2Integration -v
```

### All Tests

```bash
# Run everything (may take several minutes)
go test ./...
```

## Future Improvements

### Option 1: Headless Ebiten Testing

Ebiten may add headless testing support in the future. When available:

```go
func TestWithHeadlessEbiten(t *testing.T) {
    ebiten.SetHeadless(true)  // Hypothetical API
    img := ebiten.NewImage(100, 100)
    // Test graphics operations
}
```

### Option 2: Mock Ebiten Images

Create a mock implementation of `*ebiten.Image` for testing:

```go
type MockEbitenImage struct {
    width, height int
    pixels []color.Color
}

func (m *MockEbitenImage) Bounds() image.Rectangle { ... }
func (m *MockEbitenImage) At(x, y int) color.Color { ... }
// ... implement other methods
```

**Challenges**:
- `*ebiten.Image` is a concrete type, not an interface
- Would require refactoring to use an interface
- May not catch real Ebiten-specific bugs

### Option 3: Visual Regression Testing

Use screenshot comparison for visual tests:

```go
func TestVisualRegression(t *testing.T) {
    // Render scene
    screenshot := captureScreen()
    
    // Compare with golden image
    if !imagesEqual(screenshot, goldenImage) {
        t.Error("Visual regression detected")
    }
}
```

**Challenges**:
- Requires display/CI setup
- Platform-specific rendering differences
- Large binary assets (screenshots)

## Recommendations

### For New Features

1. **Write integration tests** for graphics features
2. **Write unit tests** for non-graphics logic
3. **Test manually** in GUI mode for visual correctness
4. **Document** expected behavior in requirements.md

### For Bug Fixes

1. **Reproduce** the bug in an integration test
2. **Fix** the bug
3. **Verify** the integration test passes
4. **Test manually** in GUI mode

### For Performance

1. **Profile** with real workloads (y_saru, etc.)
2. **Measure** frame times, memory usage
3. **Document** performance requirements in requirements.md
4. **Verify** with manual testing

## Test File Organization

```
pkg/engine/
├── *_test.go              # Unit tests (non-graphics)
├── *_integration_test.go  # Integration tests (with graphics)
└── testdata/              # Test assets
```

## Continuous Integration

### GitHub Actions / CI Setup

```yaml
# Run unit tests only (no display required)
- name: Run unit tests
  run: go test ./... -short

# Integration tests require display
# Skip in CI or use xvfb (Linux)
- name: Run integration tests
  run: xvfb-run go test ./pkg/engine -run Integration
```

## Conclusion

Due to Ebiten's display requirements, we rely heavily on integration tests for graphics code. This is acceptable because:

1. Integration tests provide end-to-end coverage
2. Real TFY scripts exercise all graphics operations
3. Manual testing catches visual bugs
4. Performance is verified with real workloads

The trade-off is longer test execution time and less granular failure messages, but this is necessary given Ebiten's architecture.
