package engine

import (
	"errors"
	"image"
	"image/color"
	"io/fs"
)

// MockAssetLoader is a mock implementation of AssetLoader for testing
type MockAssetLoader struct {
	files map[string][]byte
	dirs  map[string][]fs.DirEntry
}

// NewMockAssetLoader creates a new mock asset loader
func NewMockAssetLoader() *MockAssetLoader {
	return &MockAssetLoader{
		files: make(map[string][]byte),
		dirs:  make(map[string][]fs.DirEntry),
	}
}

// AddFile adds a file to the mock asset loader
func (m *MockAssetLoader) AddFile(name string, data []byte) {
	m.files[name] = data
}

// AddDir adds a directory listing to the mock asset loader
func (m *MockAssetLoader) AddDir(name string, entries []fs.DirEntry) {
	m.dirs[name] = entries
}

// ReadFile reads a file from the mock
func (m *MockAssetLoader) ReadFile(name string) ([]byte, error) {
	if data, ok := m.files[name]; ok {
		return data, nil
	}
	return nil, errors.New("file not found")
}

// ReadDir reads a directory from the mock
func (m *MockAssetLoader) ReadDir(name string) ([]fs.DirEntry, error) {
	if entries, ok := m.dirs[name]; ok {
		return entries, nil
	}
	return nil, errors.New("directory not found")
}

// MockDirEntry is a mock implementation of fs.DirEntry
type MockDirEntry struct {
	name  string
	isDir bool
}

func (m *MockDirEntry) Name() string               { return m.name }
func (m *MockDirEntry) IsDir() bool                { return m.isDir }
func (m *MockDirEntry) Type() fs.FileMode          { return 0 }
func (m *MockDirEntry) Info() (fs.FileInfo, error) { return nil, nil }

// MockImageDecoder is a mock implementation of ImageDecoder for testing
type MockImageDecoder struct {
	shouldFail bool
	width      int
	height     int
}

// NewMockImageDecoder creates a new mock image decoder
func NewMockImageDecoder(width, height int) *MockImageDecoder {
	return &MockImageDecoder{
		shouldFail: false,
		width:      width,
		height:     height,
	}
}

// Decode returns a mock image
func (m *MockImageDecoder) Decode(data []byte) (image.Image, error) {
	if m.shouldFail {
		return nil, errors.New("mock decode error")
	}

	// Create a simple test image
	img := image.NewRGBA(image.Rect(0, 0, m.width, m.height))
	// Fill with a test color
	for y := 0; y < m.height; y++ {
		for x := 0; x < m.width; x++ {
			img.Set(x, y, color.RGBA{255, 0, 0, 255}) // Red
		}
	}
	return img, nil
}
