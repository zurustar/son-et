// Package fileutil provides file system utility functions.
package fileutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FindFileCaseInsensitive searches for a file with the given name in the specified directory.
// The search is case-insensitive, which is useful for cross-platform compatibility.
//
// Parameters:
//   - dir: The directory to search in
//   - filename: The filename to search for (case-insensitive)
//
// Returns:
//   - string: The actual path to the file if found
//   - error: Error if the file is not found or if there's an I/O error
//
// Example:
//
//	path, err := FindFileCaseInsensitive("/path/to/dir", "MyFile.TXT")
//	// Will find "myfile.txt", "MYFILE.TXT", "MyFile.txt", etc.
func FindFileCaseInsensitive(dir, filename string) (string, error) {
	// Normalize the search filename to lowercase for comparison
	searchName := strings.ToLower(filename)

	// Read directory entries
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	// Search for matching file (case-insensitive)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Compare lowercase versions
		if strings.ToLower(entry.Name()) == searchName {
			return filepath.Join(dir, entry.Name()), nil
		}
	}

	return "", fmt.Errorf("file not found: %s (searched in %s)", filename, dir)
}

// FindFileInPaths searches for a file in multiple directories.
// The search is case-insensitive.
//
// Parameters:
//   - paths: List of directories to search in
//   - filename: The filename to search for (case-insensitive)
//
// Returns:
//   - string: The actual path to the file if found
//   - error: Error if the file is not found in any of the paths
//
// Example:
//
//	path, err := FindFileInPaths([]string{"/path1", "/path2"}, "file.wav")
func FindFileInPaths(paths []string, filename string) (string, error) {
	var searchedPaths []string

	for _, dir := range paths {
		path, err := FindFileCaseInsensitive(dir, filename)
		if err == nil {
			return path, nil
		}
		searchedPaths = append(searchedPaths, dir)
	}

	return "", fmt.Errorf("file not found: %s (searched in %v)", filename, searchedPaths)
}

// ResolveFilePath resolves a file path, handling both absolute and relative paths.
// If the path is relative, it searches in the provided base directories.
// The search is case-insensitive.
//
// Parameters:
//   - filename: The filename or path to resolve
//   - baseDirs: List of base directories to search in (for relative paths)
//
// Returns:
//   - string: The resolved absolute path
//   - error: Error if the file cannot be found
//
// Example:
//
//	path, err := ResolveFilePath("sound.wav", []string{".", "/assets"})
func ResolveFilePath(filename string, baseDirs []string) (string, error) {
	// If absolute path, check if it exists
	if filepath.IsAbs(filename) {
		if _, err := os.Stat(filename); err == nil {
			return filename, nil
		}
		return "", fmt.Errorf("file not found: %s", filename)
	}

	// For relative paths, search in base directories
	return FindFileInPaths(baseDirs, filename)
}
