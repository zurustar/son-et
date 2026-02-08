// Package preprocessor provides preprocessing functionality for FILLY scripts.
// It handles #include directives, resolving file dependencies from an entry point.
package preprocessor

import (
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/zurustar/son-et/pkg/compiler/lexer"
	"github.com/zurustar/son-et/pkg/fileutil"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// Preprocessor handles #include directive expansion and dependency resolution.
type Preprocessor struct {
	fs             fileutil.FileSystem // ファイルシステムインターフェース
	includedFiles  map[string]bool     // Set of already included files (include guard)
	includeStack   []string            // Stack for circular reference detection
	processedFiles []string            // List of processed files in order
}

// PreprocessResult contains the result of preprocessing.
type PreprocessResult struct {
	// Source is the preprocessed source code with all #include directives expanded
	Source string
	// IncludedFiles is the list of files that were included (in order of inclusion)
	IncludedFiles []string
}

// New creates a new Preprocessor with the given base directory.
func New(baseDir string) *Preprocessor {
	return &Preprocessor{
		fs:             fileutil.NewRealFS(baseDir),
		includedFiles:  make(map[string]bool),
		includeStack:   []string{},
		processedFiles: []string{},
	}
}

// NewWithFS creates a new Preprocessor with the given base directory and file system.
func NewWithFS(baseDir string, embedFS fs.FS) *Preprocessor {
	return &Preprocessor{
		fs:             fileutil.NewEmbedFS(embedFS, baseDir),
		includedFiles:  make(map[string]bool),
		includeStack:   []string{},
		processedFiles: []string{},
	}
}

// NewWithFileSystem creates a new Preprocessor with the given FileSystem interface.
func NewWithFileSystem(fsys fileutil.FileSystem) *Preprocessor {
	return &Preprocessor{
		fs:             fsys,
		includedFiles:  make(map[string]bool),
		includeStack:   []string{},
		processedFiles: []string{},
	}
}

// PreprocessFile preprocesses a file starting from the given entry point.
// It expands all #include directives recursively.
//
// Parameters:
//   - entryFile: The entry point file name (relative to baseDir)
//
// Returns:
//   - *PreprocessResult: The preprocessing result
//   - error: Any error that occurred during preprocessing
//
// Requirement 16.1: Preprocessor starts processing from entry point file.
// Requirement 16.2: Preprocessor expands #include directives.
// Requirement 16.3: Preprocessor processes included files recursively.
// Requirement 16.4: Preprocessor detects circular references.
// Requirement 16.5: Preprocessor prevents duplicate includes (include guard).
func (p *Preprocessor) PreprocessFile(entryFile string) (*PreprocessResult, error) {
	// Reset state
	p.includedFiles = make(map[string]bool)
	p.includeStack = []string{}
	p.processedFiles = []string{}

	// Process the entry file
	source, err := p.processFile(entryFile)
	if err != nil {
		return nil, err
	}

	return &PreprocessResult{
		Source:        source,
		IncludedFiles: p.processedFiles,
	}, nil
}

// processFile processes a single file, expanding #include directives.
func (p *Preprocessor) processFile(filename string) (string, error) {
	// Normalize the filename
	normalizedName := normalizeFilename(filename)

	// Check for circular reference
	// Requirement 16.4: Preprocessor detects circular references.
	for _, stackFile := range p.includeStack {
		if normalizeFilename(stackFile) == normalizedName {
			return "", fmt.Errorf("circular include detected: %s -> %s",
				strings.Join(p.includeStack, " -> "), filename)
		}
	}

	// Check include guard
	// Requirement 16.5: Preprocessor prevents duplicate includes.
	if p.includedFiles[normalizedName] {
		return "", nil // Already included, skip
	}

	// Mark as included
	p.includedFiles[normalizedName] = true
	p.includeStack = append(p.includeStack, filename)
	defer func() {
		p.includeStack = p.includeStack[:len(p.includeStack)-1]
	}()

	// Read the file using FileSystem interface
	content, err := p.readFileWithEncoding(filename)
	if err != nil {
		// Requirement 16.9: Preprocessor reports error if file not found.
		return "", fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	// Record the processed file
	p.processedFiles = append(p.processedFiles, filename)

	// Process #include directives
	// Requirement 16.2: Preprocessor expands #include directives.
	result, err := p.expandIncludes(content)
	if err != nil {
		return "", err
	}

	return result, nil
}

// expandIncludes expands #include directives in the source code.
func (p *Preprocessor) expandIncludes(source string) (string, error) {
	// Use lexer to find #include directives
	l := lexer.New(source)

	var result strings.Builder
	lastPos := 0
	sourceBytes := []byte(source)

	for {
		tok := l.NextToken()
		if tok.Type == lexer.TOKEN_EOF {
			break
		}

		if tok.Type == lexer.TOKEN_INCLUDE {
			// Extract the include filename from the token literal
			// Format: #include "filename" or #include <filename>
			includeFile := extractIncludeFilename(tok.Literal)
			if includeFile == "" {
				continue // Invalid include directive, skip
			}

			// Calculate the position of this directive in the source
			// We need to find and replace the entire #include line
			directiveStart := findDirectiveStart(source, lastPos)
			if directiveStart >= 0 {
				// Add content before the directive
				result.Write(sourceBytes[lastPos:directiveStart])

				// Process the included file
				// Requirement 16.3: Preprocessor processes included files recursively.
				includedContent, err := p.processFile(includeFile)
				if err != nil {
					return "", err
				}

				// Add the included content
				result.WriteString(includedContent)

				// Find the end of the directive line
				directiveEnd := findLineEnd(source, directiveStart)
				lastPos = directiveEnd
			}
		}
	}

	// Add remaining content
	if lastPos < len(sourceBytes) {
		result.Write(sourceBytes[lastPos:])
	}

	return result.String(), nil
}

// extractIncludeFilename extracts the filename from an #include directive.
// Supports both #include "filename" and #include <filename> formats.
func extractIncludeFilename(literal string) string {
	// Remove #include prefix
	rest := strings.TrimPrefix(literal, "#include")
	rest = strings.TrimSpace(rest)

	if len(rest) < 2 {
		return ""
	}

	// Check for quoted filename
	if rest[0] == '"' {
		end := strings.Index(rest[1:], "\"")
		if end >= 0 {
			return rest[1 : end+1]
		}
	}

	// Check for angle bracket filename
	if rest[0] == '<' {
		end := strings.Index(rest[1:], ">")
		if end >= 0 {
			return rest[1 : end+1]
		}
	}

	// Try without quotes (some FILLY files might use this)
	return strings.TrimSpace(rest)
}

// findDirectiveStart finds the start position of a directive in the source.
func findDirectiveStart(source string, startPos int) int {
	// Look for #include starting from startPos
	idx := strings.Index(source[startPos:], "#include")
	if idx >= 0 {
		return startPos + idx
	}
	return -1
}

// findLineEnd finds the end of the line (including newline character).
func findLineEnd(source string, startPos int) int {
	for i := startPos; i < len(source); i++ {
		if source[i] == '\n' {
			return i + 1
		}
		if source[i] == '\r' {
			if i+1 < len(source) && source[i+1] == '\n' {
				return i + 2
			}
			return i + 1
		}
	}
	return len(source)
}

// normalizeFilename normalizes a filename for comparison (case-insensitive).
func normalizeFilename(filename string) string {
	return strings.ToUpper(filepath.Clean(filename))
}

// readFileWithEncoding reads a file and converts from Shift-JIS to UTF-8 if needed.
func (p *Preprocessor) readFileWithEncoding(filename string) (string, error) {
	// FileSystemインターフェースを使用してファイルを読み込む
	data, err := p.fs.ReadFile(filename)
	if err != nil {
		return "", err
	}

	// Try to convert from Shift-JIS to UTF-8
	decoder := japanese.ShiftJIS.NewDecoder()
	reader := transform.NewReader(strings.NewReader(string(data)), decoder)
	utf8Data, err := io.ReadAll(reader)
	if err != nil {
		// If conversion fails, return original data
		return string(data), nil
	}

	return string(utf8Data), nil
}
