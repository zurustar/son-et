# Dependency Injection in Engine

## Overview

The engine now supports dependency injection for external dependencies, making it easier to test and mock components. This document explains how to use the dependency injection system.

## Interfaces

### AssetLoader

The `AssetLoader` interface abstracts the embedded filesystem, allowing you to provide custom implementations for testing or alternative asset sources.

```go
type AssetLoader interface {
    ReadFile(name string) ([]byte, error)
    ReadDir(name string) ([]fs.DirEntry, error)
}
```

**Default Implementation**: `EmbedFSAssetLoader` - wraps `embed.FS`

### ImageDecoder

The `ImageDecoder` interface abstracts BMP decoding, allowing you to provide custom implementations for testing or supporting additional image formats.

```go
type ImageDecoder interface {
    Decode(data []byte) (image.Image, error)
}
```

**Default Implementation**: `BMPImageDecoder` - uses `github.com/jsummers/gobmp`

## Usage

### Creating an EngineState with Default Dependencies

```go
// Creates an EngineState with default dependencies
engine := NewEngineState()
```

### Creating an EngineState with Custom Dependencies

```go
// Create custom dependencies
assetLoader := NewCustomAssetLoader()
imageDecoder := NewCustomImageDecoder()

// Create EngineState with injected dependencies
engine := NewEngineState(
    WithAssetLoader(assetLoader),
    WithImageDecoder(imageDecoder),
)
```

### Creating an EngineState from embed.FS

```go
//go:embed *.bmp *.mid *.wav
var assets embed.FS

// Initialize EngineState with embedded assets
engine := InitEngineState(assets)
```

## Testing with Mocks

### Example: Mock AssetLoader

```go
type MockAssetLoader struct {
    files map[string][]byte
}

func (m *MockAssetLoader) ReadFile(name string) ([]byte, error) {
    if data, ok := m.files[name]; ok {
        return data, nil
    }
    return nil, errors.New("file not found")
}

func (m *MockAssetLoader) ReadDir(name string) ([]fs.DirEntry, error) {
    // Return mock directory entries
    return mockEntries, nil
}

// Usage in tests
func TestMyFeature(t *testing.T) {
    mockLoader := &MockAssetLoader{
        files: map[string][]byte{
            "test.bmp": []byte("mock bmp data"),
        },
    }
    
    engine := NewEngineState(WithAssetLoader(mockLoader))
    
    // Test your feature
    picID := engine.LoadPic("test.bmp")
    // ...
}
```

### Example: Mock ImageDecoder

```go
type MockImageDecoder struct {
    width  int
    height int
}

func (m *MockImageDecoder) Decode(data []byte) (image.Image, error) {
    // Return a mock image with specified dimensions
    img := image.NewRGBA(image.Rect(0, 0, m.width, m.height))
    return img, nil
}

// Usage in tests
func TestImageLoading(t *testing.T) {
    mockDecoder := &MockImageDecoder{width: 640, height: 480}
    
    engine := NewEngineState(WithImageDecoder(mockDecoder))
    
    // Test image loading
    // ...
}
```

## Benefits

1. **Testability**: Easy to mock dependencies for unit testing
2. **Flexibility**: Can swap implementations without changing core code
3. **Isolation**: Tests don't depend on actual file system or image files
4. **Performance**: Mock implementations can be faster for testing
5. **Extensibility**: Easy to add support for new asset sources or image formats

## Implementation Details

### EngineStateOption Pattern

The engine uses the functional options pattern for dependency injection:

```go
type EngineStateOption func(*EngineState)

func WithAssetLoader(loader AssetLoader) EngineStateOption {
    return func(e *EngineState) {
        e.assetLoader = loader
    }
}
```

This pattern allows:
- Optional configuration
- Composable options
- Clear, readable API
- Easy to extend with new options

### Default Dependencies

When creating an `EngineState` without options:
- `assetLoader` is `nil` (must be set via `InitEngineState` or options)
- `imageDecoder` defaults to `BMPImageDecoder`

### Thread Safety

All dependency injection happens during initialization. Once the `EngineState` is created, dependencies are immutable, ensuring thread safety.

## Migration Guide

### Before (Global State)

```go
// Old code using global assets variable
func LoadPic(path string) int {
    data, _ := assets.ReadFile(path)
    img, _ := gobmp.Decode(bytes.NewReader(data))
    // ...
}
```

### After (Dependency Injection)

```go
// New code using injected dependencies
func (e *EngineState) LoadPic(path string) int {
    data, _ := e.assetLoader.ReadFile(path)
    img, _ := e.imageDecoder.Decode(data)
    // ...
}
```

## Future Enhancements

Potential future interfaces for dependency injection:
- `AudioPlayer` - Abstract MIDI and WAV playback
- `FontLoader` - Abstract font loading
- `Renderer` - Abstract rendering operations
- `FileSystem` - Abstract file I/O operations

## See Also

- `engine_di_test.go` - Comprehensive test examples
- `interfaces.go` - Interface definitions
- `asset_loader.go` - Default implementations
