package engine

import (
	"bytes"
	"embed"
	"image"
	"io/fs"
	"strings"

	"github.com/jsummers/gobmp"
)

// EmbedFSAssetLoader implements AssetLoader using embed.FS
type EmbedFSAssetLoader struct {
	fs embed.FS
}

// NewEmbedFSAssetLoader creates a new AssetLoader from an embed.FS
func NewEmbedFSAssetLoader(fs embed.FS) AssetLoader {
	return &EmbedFSAssetLoader{fs: fs}
}

// ReadFile reads the named file from the embedded assets
// Performs case-insensitive matching for Windows 3.1 compatibility
func (e *EmbedFSAssetLoader) ReadFile(name string) ([]byte, error) {
	// First try exact match
	data, err := e.fs.ReadFile(name)
	if err == nil {
		return data, nil
	}

	// Try case-insensitive search
	entries, err := e.fs.ReadDir(".")
	if err != nil {
		return nil, err
	}

	lowerName := strings.ToLower(name)
	for _, entry := range entries {
		if strings.ToLower(entry.Name()) == lowerName {
			return e.fs.ReadFile(entry.Name())
		}
	}

	// Return original error if not found
	return e.fs.ReadFile(name)
}

// ReadDir reads the named directory from the embedded assets
func (e *EmbedFSAssetLoader) ReadDir(name string) ([]fs.DirEntry, error) {
	return e.fs.ReadDir(name)
}

// BMPImageDecoder implements ImageDecoder for BMP files
type BMPImageDecoder struct{}

// NewBMPImageDecoder creates a new BMP image decoder
func NewBMPImageDecoder() ImageDecoder {
	return &BMPImageDecoder{}
}

// Decode decodes a BMP image from the provided byte data
func (b *BMPImageDecoder) Decode(data []byte) (image.Image, error) {
	return gobmp.Decode(bytes.NewReader(data))
}
