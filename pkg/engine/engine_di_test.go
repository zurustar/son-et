package engine

import (
	"io/fs"
	"testing"
	"time"
)

// TestEngineStateWithDependencyInjection tests that EngineState can be created with custom dependencies
func TestEngineStateWithDependencyInjection(t *testing.T) {
	// Create mock dependencies
	mockLoader := NewMockAssetLoader()
	mockDecoder := NewMockImageDecoder(100, 100)

	// Create EngineState with injected dependencies
	engine := NewEngineState(
		WithAssetLoader(mockLoader),
		WithImageDecoder(mockDecoder),
	)

	// Verify dependencies were set
	if engine.assetLoader == nil {
		t.Error("AssetLoader was not set")
	}

	if engine.imageDecoder == nil {
		t.Error("ImageDecoder was not set")
	}

	// Verify it's using our mock implementations
	if _, ok := engine.assetLoader.(*MockAssetLoader); !ok {
		t.Error("AssetLoader is not the mock implementation")
	}

	if _, ok := engine.imageDecoder.(*MockImageDecoder); !ok {
		t.Error("ImageDecoder is not the mock implementation")
	}
}

// TestLoadPicWithMockDependencies tests LoadPic with mock dependencies
func TestLoadPicWithMockDependencies(t *testing.T) {
	// Create mock dependencies
	mockLoader := NewMockAssetLoader()
	mockDecoder := NewMockImageDecoder(64, 48)

	// Add a mock file
	mockLoader.AddFile("test.bmp", []byte("mock bmp data"))
	mockLoader.AddDir(".", []fs.DirEntry{
		&MockDirEntry{name: "test.bmp", isDir: false},
	})

	// Create EngineState with mocks
	engine := NewEngineState(
		WithAssetLoader(mockLoader),
		WithImageDecoder(mockDecoder),
	)

	// Test LoadPic
	picID := engine.LoadPic("test.bmp")

	if picID < 0 {
		t.Errorf("LoadPic failed, got ID: %d", picID)
	}

	// Verify picture was created
	pic := engine.GetPicture(picID)
	if pic == nil {
		t.Error("Picture was not created")
	}

	if pic.Width != 64 || pic.Height != 48 {
		t.Errorf("Picture dimensions incorrect: got %dx%d, want 64x48", pic.Width, pic.Height)
	}
}

// TestLoadPicCaseInsensitive tests case-insensitive file matching with mocks
func TestLoadPicCaseInsensitive(t *testing.T) {
	mockLoader := NewMockAssetLoader()
	mockDecoder := NewMockImageDecoder(32, 32)

	// Add file with uppercase name
	mockLoader.AddFile("IMAGE.BMP", []byte("mock data"))
	mockLoader.AddDir(".", []fs.DirEntry{
		&MockDirEntry{name: "IMAGE.BMP", isDir: false},
	})

	engine := NewEngineState(
		WithAssetLoader(mockLoader),
		WithImageDecoder(mockDecoder),
	)

	// Try to load with lowercase name
	picID := engine.LoadPic("image.bmp")

	if picID < 0 {
		t.Error("Case-insensitive matching failed")
	}
}

// TestLoadPicWithoutAssetLoader tests that LoadPic fails gracefully without an asset loader
func TestLoadPicWithoutAssetLoader(t *testing.T) {
	// Create EngineState without asset loader
	engine := NewEngineState()

	// LoadPic should fail gracefully
	picID := engine.LoadPic("test.bmp")

	if picID != -1 {
		t.Errorf("Expected LoadPic to return -1 without asset loader, got %d", picID)
	}
}

// TestEngineStateResetWithDI tests that Reset clears all state with dependency injection
func TestEngineStateResetWithDI(t *testing.T) {
	mockLoader := NewMockAssetLoader()
	mockDecoder := NewMockImageDecoder(100, 100)

	mockLoader.AddFile("test.bmp", []byte("data"))
	mockLoader.AddDir(".", []fs.DirEntry{
		&MockDirEntry{name: "test.bmp", isDir: false},
	})

	engine := NewEngineState(
		WithAssetLoader(mockLoader),
		WithImageDecoder(mockDecoder),
	)

	// Create some state
	picID := engine.LoadPic("test.bmp")
	if picID < 0 {
		t.Fatal("Failed to load picture")
	}

	winID := engine.OpenWin(picID, 0, 0, 640, 480, 0, 0, 0xFFFFFF)
	if winID < 0 {
		t.Fatal("Failed to open window")
	}

	// Reset
	engine.Reset()

	// Verify state was cleared
	if len(engine.pictures) != 0 {
		t.Error("Pictures were not cleared")
	}

	if len(engine.windows) != 0 {
		t.Error("Windows were not cleared")
	}

	if engine.nextPicID != 0 {
		t.Error("Picture ID counter was not reset")
	}

	if engine.nextWinID != 0 {
		t.Error("Window ID counter was not reset")
	}
}

// TestDefaultDependencies tests that default dependencies are set
func TestDefaultDependencies(t *testing.T) {
	engine := NewEngineState()

	// ImageDecoder should have a default
	if engine.imageDecoder == nil {
		t.Error("Default ImageDecoder was not set")
	}

	// Should be BMPImageDecoder
	if _, ok := engine.imageDecoder.(*BMPImageDecoder); !ok {
		t.Error("Default ImageDecoder is not BMPImageDecoder")
	}
}

// BenchmarkLoadPicWithMocks benchmarks LoadPic with mock dependencies
func BenchmarkLoadPicWithMocks(b *testing.B) {
	mockLoader := NewMockAssetLoader()
	mockDecoder := NewMockImageDecoder(640, 480)

	mockLoader.AddFile("test.bmp", make([]byte, 1024))
	mockLoader.AddDir(".", []fs.DirEntry{
		&MockDirEntry{name: "test.bmp", isDir: false},
	})

	engine := NewEngineState(
		WithAssetLoader(mockLoader),
		WithImageDecoder(mockDecoder),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset state between iterations
		engine.Reset()
		engine.nextPicID = 0

		_ = engine.LoadPic("test.bmp")
	}
}

// TestConcurrentLoadPic tests thread safety of LoadPic
func TestConcurrentLoadPic(t *testing.T) {
	mockLoader := NewMockAssetLoader()
	mockDecoder := NewMockImageDecoder(100, 100)

	// Add multiple files
	for i := 0; i < 10; i++ {
		name := "test" + string(rune('0'+i)) + ".bmp"
		mockLoader.AddFile(name, []byte("data"))
	}

	entries := make([]fs.DirEntry, 10)
	for i := 0; i < 10; i++ {
		name := "test" + string(rune('0'+i)) + ".bmp"
		entries[i] = &MockDirEntry{name: name, isDir: false}
	}
	mockLoader.AddDir(".", entries)

	engine := NewEngineState(
		WithAssetLoader(mockLoader),
		WithImageDecoder(mockDecoder),
	)

	// Load pictures concurrently
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(idx int) {
			name := "test" + string(rune('0'+idx)) + ".bmp"
			picID := engine.LoadPic(name)
			if picID < 0 {
				t.Errorf("Failed to load %s", name)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines with timeout
	timeout := time.After(5 * time.Second)
	for i := 0; i < 10; i++ {
		select {
		case <-done:
			// Success
		case <-timeout:
			t.Fatal("Test timed out")
		}
	}

	// Verify all pictures were loaded
	if len(engine.pictures) != 10 {
		t.Errorf("Expected 10 pictures, got %d", len(engine.pictures))
	}
}
