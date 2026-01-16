package engine

import (
	"bytes"
	"embed"
	"image"
	"io/fs"
	"os"
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

// FilesystemAssetLoader implements AssetLoader for filesystem access (direct mode)
type FilesystemAssetLoader struct {
	baseDir string
}

// NewFilesystemAssetLoader creates a new AssetLoader for filesystem access
func NewFilesystemAssetLoader(baseDir string) AssetLoader {
	return &FilesystemAssetLoader{baseDir: baseDir}
}

// ReadFile reads the named file from the filesystem
// Performs case-insensitive matching for Windows 3.1 compatibility
func (f *FilesystemAssetLoader) ReadFile(name string) ([]byte, error) {
	// First try exact match
	fullPath := name
	if !strings.HasPrefix(name, "/") && !strings.HasPrefix(name, f.baseDir) {
		fullPath = strings.TrimPrefix(name, "./")
		fullPath = strings.TrimPrefix(fullPath, f.baseDir+"/")
		fullPath = strings.Join([]string{f.baseDir, fullPath}, "/")
	}

	data, err := os.ReadFile(fullPath)
	if err == nil {
		return data, nil
	}

	// Try case-insensitive search in base directory
	entries, err := os.ReadDir(f.baseDir)
	if err != nil {
		return nil, err
	}

	lowerName := strings.ToLower(name)
	for _, entry := range entries {
		if strings.ToLower(entry.Name()) == lowerName {
			return os.ReadFile(strings.Join([]string{f.baseDir, entry.Name()}, "/"))
		}
	}

	// Return original error if not found
	return os.ReadFile(fullPath)
}

// ReadDir reads the named directory from the filesystem
func (f *FilesystemAssetLoader) ReadDir(name string) ([]fs.DirEntry, error) {
	fullPath := name
	if !strings.HasPrefix(name, "/") && !strings.HasPrefix(name, f.baseDir) {
		fullPath = strings.TrimPrefix(name, "./")
		fullPath = strings.TrimPrefix(fullPath, f.baseDir+"/")
		fullPath = strings.Join([]string{f.baseDir, fullPath}, "/")
	}
	return os.ReadDir(fullPath)
}
