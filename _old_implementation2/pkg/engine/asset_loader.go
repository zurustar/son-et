package engine

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// FilesystemAssetLoader loads assets from the filesystem with case-insensitive matching.
// This provides Windows 3.1 compatibility where filenames are case-insensitive.
type FilesystemAssetLoader struct {
	baseDir string
}

// NewFilesystemAssetLoader creates a new filesystem asset loader.
// baseDir is the root directory for asset loading.
func NewFilesystemAssetLoader(baseDir string) *FilesystemAssetLoader {
	return &FilesystemAssetLoader{
		baseDir: baseDir,
	}
}

// ReadFile reads a file with case-insensitive matching.
func (f *FilesystemAssetLoader) ReadFile(path string) ([]byte, error) {
	// Try exact match first
	fullPath := filepath.Join(f.baseDir, path)
	data, err := os.ReadFile(fullPath)
	if err == nil {
		return data, nil
	}

	// Try case-insensitive match
	actualPath, err := f.findCaseInsensitive(path)
	if err != nil {
		return nil, fmt.Errorf("file not found: %s", path)
	}

	return os.ReadFile(actualPath)
}

// Exists checks if a file exists with case-insensitive matching.
func (f *FilesystemAssetLoader) Exists(path string) bool {
	fullPath := filepath.Join(f.baseDir, path)
	if _, err := os.Stat(fullPath); err == nil {
		return true
	}

	_, err := f.findCaseInsensitive(path)
	return err == nil
}

// ListFiles lists files matching a pattern with case-insensitive matching.
func (f *FilesystemAssetLoader) ListFiles(pattern string) ([]string, error) {
	var matches []string

	err := filepath.Walk(f.baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(f.baseDir, path)
		if err != nil {
			return err
		}

		matched, err := filepath.Match(pattern, filepath.Base(relPath))
		if err != nil {
			return err
		}
		if matched {
			matches = append(matches, relPath)
		}

		return nil
	})

	return matches, err
}

// findCaseInsensitive finds a file with case-insensitive matching.
func (f *FilesystemAssetLoader) findCaseInsensitive(path string) (string, error) {
	parts := strings.Split(filepath.ToSlash(path), "/")
	currentPath := f.baseDir

	for _, part := range parts {
		entries, err := os.ReadDir(currentPath)
		if err != nil {
			return "", err
		}

		found := false
		for _, entry := range entries {
			if strings.EqualFold(entry.Name(), part) {
				currentPath = filepath.Join(currentPath, entry.Name())
				found = true
				break
			}
		}

		if !found {
			return "", fmt.Errorf("path component not found: %s", part)
		}
	}

	return currentPath, nil
}

// EmbedFSAssetLoader loads assets from an embedded filesystem.
type EmbedFSAssetLoader struct {
	fs      embed.FS
	baseDir string // Base directory within the embedded FS
}

// NewEmbedFSAssetLoader creates a new embedded filesystem asset loader.
func NewEmbedFSAssetLoader(embedFS embed.FS) *EmbedFSAssetLoader {
	return &EmbedFSAssetLoader{
		fs:      embedFS,
		baseDir: "",
	}
}

// NewEmbedFSAssetLoaderWithBaseDir creates a new embedded filesystem asset loader with a base directory.
func NewEmbedFSAssetLoaderWithBaseDir(embedFS embed.FS, baseDir string) *EmbedFSAssetLoader {
	return &EmbedFSAssetLoader{
		fs:      embedFS,
		baseDir: baseDir,
	}
}

// ReadFile reads a file from the embedded filesystem.
func (e *EmbedFSAssetLoader) ReadFile(path string) ([]byte, error) {
	fullPath := filepath.Join(e.baseDir, path)
	return e.fs.ReadFile(fullPath)
}

// Exists checks if a file exists in the embedded filesystem.
func (e *EmbedFSAssetLoader) Exists(path string) bool {
	fullPath := filepath.Join(e.baseDir, path)
	_, err := e.fs.Open(fullPath)
	return err == nil
}

// ListFiles lists files matching a pattern in the embedded filesystem.
func (e *EmbedFSAssetLoader) ListFiles(pattern string) ([]string, error) {
	var matches []string

	startDir := "."
	if e.baseDir != "" {
		startDir = e.baseDir
	}

	err := fs.WalkDir(e.fs, startDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		// Remove baseDir prefix from path for matching
		relPath := path
		if e.baseDir != "" && strings.HasPrefix(path, e.baseDir+"/") {
			relPath = strings.TrimPrefix(path, e.baseDir+"/")
		}

		matched, err := filepath.Match(pattern, filepath.Base(relPath))
		if err != nil {
			return err
		}
		if matched {
			matches = append(matches, relPath)
		}

		return nil
	})

	return matches, err
}
