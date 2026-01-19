package engine

import (
	"bytes"
	"image"
	"image/color"
	"io"
	"testing"
)

// MockRenderer is a test implementation of Renderer
type MockRenderer struct {
	RenderCount int
	ClearCount  int
	LastColor   uint32
}

func (m *MockRenderer) RenderFrame(screen image.Image, state *EngineState) {
	m.RenderCount++
}

func (m *MockRenderer) Clear(c uint32) {
	m.ClearCount++
	m.LastColor = c
}

// MockAssetLoader is a test implementation of AssetLoader
type MockAssetLoader struct {
	Files map[string][]byte
}

func (m *MockAssetLoader) ReadFile(path string) ([]byte, error) {
	if data, ok := m.Files[path]; ok {
		return data, nil
	}
	return nil, io.EOF
}

func (m *MockAssetLoader) Exists(path string) bool {
	_, ok := m.Files[path]
	return ok
}

func (m *MockAssetLoader) ListFiles(pattern string) ([]string, error) {
	// Simple implementation - return all files
	files := make([]string, 0, len(m.Files))
	for path := range m.Files {
		files = append(files, path)
	}
	return files, nil
}

// MockImageDecoder is a test implementation of ImageDecoder
type MockImageDecoder struct {
	Width  int
	Height int
}

func (m *MockImageDecoder) Decode(r io.Reader) (image.Image, string, error) {
	// Return a simple test image
	img := image.NewRGBA(image.Rect(0, 0, m.Width, m.Height))
	return img, "test", nil
}

func (m *MockImageDecoder) DecodeConfig(r io.Reader) (image.Config, string, error) {
	return image.Config{
		ColorModel: color.RGBAModel,
		Width:      m.Width,
		Height:     m.Height,
	}, "test", nil
}

// MockTickGenerator is a test implementation of TickGenerator
type MockTickGenerator struct {
	CurrentTick       int
	LastDeliveredTick int
}

func (m *MockTickGenerator) CalculateTickFromTime(elapsed float64) int {
	return m.CurrentTick
}

func (m *MockTickGenerator) GetLastDeliveredTick() int {
	return m.LastDeliveredTick
}

func (m *MockTickGenerator) SetLastDeliveredTick(tick int) {
	m.LastDeliveredTick = tick
}

func TestMockRenderer(t *testing.T) {
	mock := &MockRenderer{}

	// Test RenderFrame
	mock.RenderFrame(nil, nil)
	if mock.RenderCount != 1 {
		t.Errorf("Expected RenderCount=1, got %d", mock.RenderCount)
	}

	// Test Clear
	mock.Clear(0xFF0000)
	if mock.ClearCount != 1 {
		t.Errorf("Expected ClearCount=1, got %d", mock.ClearCount)
	}
	if mock.LastColor != 0xFF0000 {
		t.Errorf("Expected LastColor=0xFF0000, got %X", mock.LastColor)
	}
}

func TestMockAssetLoader(t *testing.T) {
	mock := &MockAssetLoader{
		Files: map[string][]byte{
			"test.txt": []byte("hello"),
			"data.bin": []byte{0x01, 0x02, 0x03},
		},
	}

	// Test ReadFile
	data, err := mock.ReadFile("test.txt")
	if err != nil {
		t.Errorf("ReadFile failed: %v", err)
	}
	if string(data) != "hello" {
		t.Errorf("Expected 'hello', got %s", string(data))
	}

	// Test Exists
	if !mock.Exists("test.txt") {
		t.Error("Expected test.txt to exist")
	}
	if mock.Exists("missing.txt") {
		t.Error("Expected missing.txt to not exist")
	}

	// Test ListFiles
	files, err := mock.ListFiles("*")
	if err != nil {
		t.Errorf("ListFiles failed: %v", err)
	}
	if len(files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(files))
	}
}

func TestMockImageDecoder(t *testing.T) {
	mock := &MockImageDecoder{
		Width:  100,
		Height: 200,
	}

	// Test Decode
	img, format, err := mock.Decode(bytes.NewReader(nil))
	if err != nil {
		t.Errorf("Decode failed: %v", err)
	}
	if format != "test" {
		t.Errorf("Expected format 'test', got %s", format)
	}
	bounds := img.Bounds()
	if bounds.Dx() != 100 || bounds.Dy() != 200 {
		t.Errorf("Expected 100x200 image, got %dx%d", bounds.Dx(), bounds.Dy())
	}

	// Test DecodeConfig
	config, format, err := mock.DecodeConfig(bytes.NewReader(nil))
	if err != nil {
		t.Errorf("DecodeConfig failed: %v", err)
	}
	if config.Width != 100 || config.Height != 200 {
		t.Errorf("Expected 100x200 config, got %dx%d", config.Width, config.Height)
	}
}

func TestMockTickGenerator(t *testing.T) {
	mock := &MockTickGenerator{
		CurrentTick: 100,
	}

	// Test CalculateTickFromTime
	tick := mock.CalculateTickFromTime(1.5)
	if tick != 100 {
		t.Errorf("Expected tick=100, got %d", tick)
	}

	// Test Get/SetLastDeliveredTick
	mock.SetLastDeliveredTick(50)
	if mock.GetLastDeliveredTick() != 50 {
		t.Errorf("Expected LastDeliveredTick=50, got %d", mock.GetLastDeliveredTick())
	}
}

func TestEngineState_Creation(t *testing.T) {
	renderer := &MockRenderer{}
	assetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
	imageDecoder := &MockImageDecoder{Width: 640, Height: 480}

	state := NewEngineState(renderer, assetLoader, imageDecoder)

	if state == nil {
		t.Fatal("NewEngineState returned nil")
	}
	if state.renderer == nil {
		t.Error("Renderer not set")
	}
	if state.assetLoader == nil {
		t.Error("AssetLoader not set")
	}
	if state.imageDecoder == nil {
		t.Error("ImageDecoder not set")
	}
	if state.pictures == nil {
		t.Error("Pictures map not initialized")
	}
	if state.windows == nil {
		t.Error("Windows map not initialized")
	}
	if state.casts == nil {
		t.Error("Casts map not initialized")
	}
}

func TestEngineState_Configuration(t *testing.T) {
	state := NewEngineState(nil, nil, nil)

	// Test headless mode
	state.SetHeadlessMode(true)
	if !state.headlessMode {
		t.Error("Headless mode not set")
	}

	// Test debug level
	state.SetDebugLevel(2)
	if state.debugLevel != 2 {
		t.Errorf("Expected debug level 2, got %d", state.debugLevel)
	}
}

func TestEngineState_TickCounter(t *testing.T) {
	state := NewEngineState(nil, nil, nil)

	if state.GetTickCount() != 0 {
		t.Errorf("Expected initial tick count 0, got %d", state.GetTickCount())
	}

	state.IncrementTick()
	if state.GetTickCount() != 1 {
		t.Errorf("Expected tick count 1, got %d", state.GetTickCount())
	}

	state.IncrementTick()
	state.IncrementTick()
	if state.GetTickCount() != 3 {
		t.Errorf("Expected tick count 3, got %d", state.GetTickCount())
	}
}
