package preprocessor

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// Metadata holds #info directive values.
type Metadata struct {
	Title       string
	Author      string
	Version     string
	Description string
	Custom      map[string]string
}

// Preprocessor handles #info and #include directives with encoding conversion.
type Preprocessor struct {
	baseDir       string
	metadata      *Metadata
	includedFiles map[string]bool // Track included files for circular detection
	assetLoader   AssetLoader
}

// AssetLoader interface for loading files (supports both filesystem and embedded).
type AssetLoader interface {
	ReadFile(path string) ([]byte, error)
	Exists(path string) bool
}

// NewPreprocessor creates a new preprocessor.
func NewPreprocessor(baseDir string, assetLoader AssetLoader) *Preprocessor {
	return &Preprocessor{
		baseDir:       baseDir,
		metadata:      &Metadata{Custom: make(map[string]string)},
		includedFiles: make(map[string]bool),
		assetLoader:   assetLoader,
	}
}

// Process processes a source file with encoding conversion and directive handling.
func (p *Preprocessor) Process(filename string) (string, error) {
	// Reset state
	p.includedFiles = make(map[string]bool)

	// Process the main file
	return p.processFile(filename)
}

// GetMetadata returns the collected metadata.
func (p *Preprocessor) GetMetadata() *Metadata {
	return p.metadata
}

// processFile processes a single file recursively.
func (p *Preprocessor) processFile(filename string) (string, error) {
	// Construct full path for circular include tracking
	fullPath := filepath.Join(p.baseDir, filename)

	// Check for circular includes
	if p.includedFiles[fullPath] {
		return "", fmt.Errorf("circular include detected: %s", filename)
	}
	p.includedFiles[fullPath] = true

	// Read file (AssetLoader handles baseDir internally for FilesystemAssetLoader)
	data, err := p.assetLoader.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	// Detect and convert encoding
	text, err := p.convertEncoding(data)
	if err != nil {
		return "", fmt.Errorf("encoding error in %s: %w", filename, err)
	}

	// Process directives
	return p.processDirectives(text, filename)
}

// convertEncoding detects and converts Shift-JIS to UTF-8.
func (p *Preprocessor) convertEncoding(data []byte) (string, error) {
	// Try UTF-8 first (check if valid)
	if isValidUTF8(data) {
		return string(data), nil
	}

	// Try Shift-JIS conversion
	decoder := japanese.ShiftJIS.NewDecoder()
	utf8Data, err := io.ReadAll(transform.NewReader(bytes.NewReader(data), decoder))
	if err != nil {
		return "", fmt.Errorf("failed to decode Shift-JIS: %w", err)
	}

	return string(utf8Data), nil
}

// isValidUTF8 checks if data is valid UTF-8.
func isValidUTF8(data []byte) bool {
	// Simple heuristic: check for common UTF-8 patterns
	// If it contains null bytes or invalid sequences, it's likely not UTF-8
	for i := 0; i < len(data); i++ {
		if data[i] == 0 {
			return false
		}
		// Check for valid UTF-8 multi-byte sequences
		if data[i] >= 0x80 {
			if data[i] < 0xC0 {
				return false // Invalid start byte
			}
			// Count continuation bytes
			var count int
			if data[i] < 0xE0 {
				count = 1
			} else if data[i] < 0xF0 {
				count = 2
			} else {
				count = 3
			}
			// Check continuation bytes
			for j := 1; j <= count; j++ {
				if i+j >= len(data) || (data[i+j]&0xC0) != 0x80 {
					return false
				}
			}
			i += count
		}
	}
	return true
}

// processDirectives processes #info and #include directives.
func (p *Preprocessor) processDirectives(text string, currentFile string) (string, error) {
	var result strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(text))

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Handle #info directive
		if strings.HasPrefix(trimmed, "#info") {
			if err := p.parseInfoDirective(trimmed); err != nil {
				return "", err
			}
			continue // Don't include #info in output
		}

		// Handle #include directive
		if strings.HasPrefix(trimmed, "#include") {
			included, err := p.parseIncludeDirective(trimmed, currentFile)
			if err != nil {
				return "", err
			}
			result.WriteString(included)
			result.WriteString("\n")
			continue
		}

		// Regular line
		result.WriteString(line)
		result.WriteString("\n")
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return result.String(), nil
}

// parseInfoDirective parses a #info directive.
func (p *Preprocessor) parseInfoDirective(line string) error {
	// Format: #info key value
	parts := strings.SplitN(line, " ", 3)
	if len(parts) < 3 {
		return fmt.Errorf("invalid #info directive: %s", line)
	}

	key := strings.ToLower(parts[1])
	value := strings.Trim(parts[2], "\"")

	// Store in metadata
	switch key {
	case "title":
		p.metadata.Title = value
	case "author":
		p.metadata.Author = value
	case "version":
		p.metadata.Version = value
	case "description":
		p.metadata.Description = value
	default:
		p.metadata.Custom[key] = value
	}

	return nil
}

// parseIncludeDirective parses a #include directive and processes the included file.
func (p *Preprocessor) parseIncludeDirective(line string, currentFile string) (string, error) {
	// Format: #include "filename"
	parts := strings.SplitN(line, " ", 2)
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid #include directive: %s", line)
	}

	filename := strings.Trim(parts[1], "\"")

	// Resolve path relative to current file
	currentDir := filepath.Dir(currentFile)
	includePath := filepath.Join(currentDir, filename)

	// Check if file exists (AssetLoader handles baseDir internally)
	if !p.assetLoader.Exists(includePath) {
		return "", fmt.Errorf("included file not found: %s", includePath)
	}

	// Process the included file recursively
	return p.processFile(includePath)
}
