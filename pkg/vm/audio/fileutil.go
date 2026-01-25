// Package audio provides audio-related components for the FILLY virtual machine.
// This file implements case-insensitive file search utilities for Windows 3.1 compatibility.
package audio

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FindFileInsensitive searches for a file with case-insensitive matching.
// This is necessary for Windows 3.1 era programs where filenames may have
// inconsistent casing.
//
// Parameters:
//   - filename: The filename to search for (can be absolute or relative path)
//
// Returns:
//   - string: The actual path to the file with correct casing
//   - error: Error if the file cannot be found
//
// Example:
//   - FindFileInsensitive("KUMA.MID") might return "kuma.mid"
//   - FindFileInsensitive("samples/kuma2/TITLE.BMP") might return "samples/kuma2/title.bmp"
func FindFileInsensitive(filename string) (string, error) {
	// First try exact match (fast path)
	if _, err := os.Stat(filename); err == nil {
		return filename, nil
	}

	// Split into directory and filename
	dir := filepath.Dir(filename)
	base := filepath.Base(filename)

	// If directory doesn't exist, try to find it case-insensitively
	if dir != "." && dir != "/" {
		actualDir, err := findDirInsensitive(dir)
		if err != nil {
			return "", fmt.Errorf("directory not found: %s", dir)
		}
		dir = actualDir
	}

	// Read directory entries
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	// Search for matching filename (case-insensitive)
	baseLower := strings.ToLower(base)
	for _, entry := range entries {
		if strings.ToLower(entry.Name()) == baseLower {
			return filepath.Join(dir, entry.Name()), nil
		}
	}

	return "", fmt.Errorf("file not found: %s", filename)
}

// findDirInsensitive finds a directory with case-insensitive matching.
// It recursively searches from the root to find each directory component.
func findDirInsensitive(path string) (string, error) {
	// If path is absolute, handle it differently
	if filepath.IsAbs(path) {
		return findAbsDirInsensitive(path)
	}

	// For relative paths, split into components
	components := strings.Split(filepath.ToSlash(path), "/")
	currentPath := "."

	for _, component := range components {
		if component == "." || component == "" {
			continue
		}

		// Read current directory
		entries, err := os.ReadDir(currentPath)
		if err != nil {
			return "", err
		}

		// Find matching component (case-insensitive)
		componentLower := strings.ToLower(component)
		found := false
		for _, entry := range entries {
			if entry.IsDir() && strings.ToLower(entry.Name()) == componentLower {
				currentPath = filepath.Join(currentPath, entry.Name())
				found = true
				break
			}
		}

		if !found {
			return "", fmt.Errorf("directory component not found: %s in %s", component, currentPath)
		}
	}

	return currentPath, nil
}

// findAbsDirInsensitive finds an absolute directory path with case-insensitive matching.
func findAbsDirInsensitive(path string) (string, error) {
	// For absolute paths, we need to handle the root differently
	// This is a simplified version that works for most cases
	components := strings.Split(filepath.ToSlash(path), "/")
	currentPath := "/"

	for i, component := range components {
		if component == "" {
			continue
		}

		// For the first component on Windows (drive letter), keep as-is
		if i == 0 && len(component) == 2 && component[1] == ':' {
			currentPath = component + "/"
			continue
		}

		// Read current directory
		entries, err := os.ReadDir(currentPath)
		if err != nil {
			return "", err
		}

		// Find matching component (case-insensitive)
		componentLower := strings.ToLower(component)
		found := false
		for _, entry := range entries {
			if entry.IsDir() && strings.ToLower(entry.Name()) == componentLower {
				currentPath = filepath.Join(currentPath, entry.Name())
				found = true
				break
			}
		}

		if !found {
			return "", fmt.Errorf("directory component not found: %s", component)
		}
	}

	return currentPath, nil
}
