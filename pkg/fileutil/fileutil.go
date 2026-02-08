// Package fileutil provides file system utility functions.
package fileutil

import (
	"fmt"
	"io/fs"
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

// FindFileCaseInsensitiveFS searches for a file with the given name in the specified directory
// using the provided file system (can be embed.FS or os.DirFS).
// The search is case-insensitive.
//
// Parameters:
//   - fsys: The file system to search in
//   - dir: The directory to search in
//   - filename: The filename to search for (case-insensitive)
//
// Returns:
//   - string: The actual path to the file if found
//   - error: Error if the file is not found or if there's an I/O error
func FindFileCaseInsensitiveFS(fsys fs.FS, dir, filename string) (string, error) {
	// Normalize the search filename to lowercase for comparison
	searchName := strings.ToLower(filename)

	// Read directory entries
	entries, err := fs.ReadDir(fsys, dir)
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
			// fs.FS uses forward slashes
			return dir + "/" + entry.Name(), nil
		}
	}

	return "", fmt.Errorf("file not found: %s (searched in %s)", filename, dir)
}


