package engine

import (
	"testing"
)

func TestNewWAVPlayer(t *testing.T) {
	renderer := &MockRenderer{}
	assetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
	imageDecoder := &MockImageDecoder{Width: 640, Height: 480}
	engine := NewEngine(renderer, assetLoader, imageDecoder)

	if engine.wavPlayer == nil {
		t.Fatal("WAV player not created")
	}

	if engine.wavPlayer.players == nil {
		t.Error("Players map not initialized")
	}

	if engine.wavPlayer.resources == nil {
		t.Error("Resources map not initialized")
	}
}

func TestPlayWAVE_FileNotFound(t *testing.T) {
	renderer := &MockRenderer{}
	assetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
	imageDecoder := &MockImageDecoder{Width: 640, Height: 480}
	engine := NewEngine(renderer, assetLoader, imageDecoder)

	err := engine.PlayWAVE("nonexistent.wav")
	if err == nil {
		t.Error("Expected error for nonexistent WAV file")
	}
}

func TestPlayWAVE_InvalidWAV(t *testing.T) {
	renderer := &MockRenderer{}
	assetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
	imageDecoder := &MockImageDecoder{Width: 640, Height: 480}
	engine := NewEngine(renderer, assetLoader, imageDecoder)

	// Add invalid WAV data
	assetLoader.Files["invalid.wav"] = []byte("not a wav file")

	err := engine.PlayWAVE("invalid.wav")
	if err == nil {
		t.Error("Expected error for invalid WAV file")
	}
}

func TestLoadRsc_FileNotFound(t *testing.T) {
	renderer := &MockRenderer{}
	assetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
	imageDecoder := &MockImageDecoder{Width: 640, Height: 480}
	engine := NewEngine(renderer, assetLoader, imageDecoder)

	_, err := engine.LoadRsc("nonexistent.wav")
	if err == nil {
		t.Error("Expected error for nonexistent WAV file")
	}
}

func TestLoadRsc_Success(t *testing.T) {
	renderer := &MockRenderer{}
	assetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
	imageDecoder := &MockImageDecoder{Width: 640, Height: 480}
	engine := NewEngine(renderer, assetLoader, imageDecoder)

	// Add dummy WAV data (just needs to be loadable, not playable for this test)
	assetLoader.Files["test.wav"] = []byte("dummy wav data")

	resourceID, err := engine.LoadRsc("test.wav")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if resourceID < 0 {
		t.Errorf("Expected non-negative resource ID, got %d", resourceID)
	}
}

func TestLoadRsc_MultipleResources(t *testing.T) {
	renderer := &MockRenderer{}
	assetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
	imageDecoder := &MockImageDecoder{Width: 640, Height: 480}
	engine := NewEngine(renderer, assetLoader, imageDecoder)

	// Add dummy WAV data
	assetLoader.Files["test1.wav"] = []byte("dummy wav data 1")
	assetLoader.Files["test2.wav"] = []byte("dummy wav data 2")

	id1, err1 := engine.LoadRsc("test1.wav")
	if err1 != nil {
		t.Errorf("Unexpected error loading resource 1: %v", err1)
	}

	id2, err2 := engine.LoadRsc("test2.wav")
	if err2 != nil {
		t.Errorf("Unexpected error loading resource 2: %v", err2)
	}

	if id1 == id2 {
		t.Error("Expected different resource IDs for different files")
	}
}

func TestPlayRsc_ResourceNotFound(t *testing.T) {
	renderer := &MockRenderer{}
	assetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
	imageDecoder := &MockImageDecoder{Width: 640, Height: 480}
	engine := NewEngine(renderer, assetLoader, imageDecoder)

	err := engine.PlayRsc(999)
	if err == nil {
		t.Error("Expected error for nonexistent resource ID")
	}
}

func TestDelRsc(t *testing.T) {
	renderer := &MockRenderer{}
	assetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
	imageDecoder := &MockImageDecoder{Width: 640, Height: 480}
	engine := NewEngine(renderer, assetLoader, imageDecoder)

	// Add dummy WAV data
	assetLoader.Files["test.wav"] = []byte("dummy wav data")

	// Load resource
	resourceID, err := engine.LoadRsc("test.wav")
	if err != nil {
		t.Fatalf("Failed to load resource: %v", err)
	}

	// Delete resource
	engine.DelRsc(resourceID)

	// Try to play deleted resource
	err = engine.PlayRsc(resourceID)
	if err == nil {
		t.Error("Expected error when playing deleted resource")
	}
}

func TestStopAllWAV(t *testing.T) {
	renderer := &MockRenderer{}
	assetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
	imageDecoder := &MockImageDecoder{Width: 640, Height: 480}
	engine := NewEngine(renderer, assetLoader, imageDecoder)

	// Should not panic even if no players are active
	engine.StopAllWAV()
}

func TestCleanupWAV(t *testing.T) {
	renderer := &MockRenderer{}
	assetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
	imageDecoder := &MockImageDecoder{Width: 640, Height: 480}
	engine := NewEngine(renderer, assetLoader, imageDecoder)

	// Should not panic even if no players are active
	engine.CleanupWAV()
}

func TestGetActivePlayerCount(t *testing.T) {
	renderer := &MockRenderer{}
	assetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
	imageDecoder := &MockImageDecoder{Width: 640, Height: 480}
	engine := NewEngine(renderer, assetLoader, imageDecoder)

	count := engine.wavPlayer.GetActivePlayerCount()
	if count != 0 {
		t.Errorf("Expected 0 active players initially, got %d", count)
	}
}

func TestWAVStream_Read(t *testing.T) {
	data := []byte("hello world")
	stream := NewWAVStream(data)

	// Read first 5 bytes
	buf := make([]byte, 5)
	n, err := stream.Read(buf)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if n != 5 {
		t.Errorf("Expected to read 5 bytes, got %d", n)
	}
	if string(buf) != "hello" {
		t.Errorf("Expected 'hello', got '%s'", string(buf))
	}

	// Read next 6 bytes
	buf = make([]byte, 6)
	n, err = stream.Read(buf)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if n != 6 {
		t.Errorf("Expected to read 6 bytes, got %d", n)
	}
	if string(buf) != " world" {
		t.Errorf("Expected ' world', got '%s'", string(buf))
	}

	// Read past end
	buf = make([]byte, 10)
	n, err = stream.Read(buf)
	if err == nil {
		t.Error("Expected EOF error")
	}
	if n != 0 {
		t.Errorf("Expected 0 bytes at EOF, got %d", n)
	}
}

func TestWAVStream_Seek(t *testing.T) {
	data := []byte("hello world")
	stream := NewWAVStream(data)

	// Seek to position 6
	pos, err := stream.Seek(6, 0) // SeekStart
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if pos != 6 {
		t.Errorf("Expected position 6, got %d", pos)
	}

	// Read from new position
	buf := make([]byte, 5)
	_, err = stream.Read(buf)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if string(buf) != "world" {
		t.Errorf("Expected 'world', got '%s'", string(buf))
	}

	// Seek relative
	pos, err = stream.Seek(-5, 1) // SeekCurrent
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if pos != 6 {
		t.Errorf("Expected position 6, got %d", pos)
	}

	// Seek from end
	pos, err = stream.Seek(-5, 2) // SeekEnd
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if pos != 6 {
		t.Errorf("Expected position 6, got %d", pos)
	}
}

func TestWAVStream_SeekNegative(t *testing.T) {
	data := []byte("hello world")
	stream := NewWAVStream(data)

	// Try to seek to negative position
	_, err := stream.Seek(-1, 0) // SeekStart
	if err == nil {
		t.Error("Expected error for negative position")
	}
}

func TestWAVStream_SeekInvalidWhence(t *testing.T) {
	data := []byte("hello world")
	stream := NewWAVStream(data)

	// Try invalid whence
	_, err := stream.Seek(0, 99)
	if err == nil {
		t.Error("Expected error for invalid whence")
	}
}
