package preprocessor

import (
	"strings"
	"testing"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// MockAssetLoader for testing
type MockAssetLoader struct {
	Files map[string][]byte
}

func (m *MockAssetLoader) ReadFile(path string) ([]byte, error) {
	if data, ok := m.Files[path]; ok {
		return data, nil
	}
	return nil, &FileNotFoundError{Path: path}
}

func (m *MockAssetLoader) Exists(path string) bool {
	_, ok := m.Files[path]
	return ok
}

type FileNotFoundError struct {
	Path string
}

func (e *FileNotFoundError) Error() string {
	return "file not found: " + e.Path
}

func TestPreprocessor_InfoDirective(t *testing.T) {
	loader := &MockAssetLoader{
		Files: map[string][]byte{
			"test.tfy": []byte(`#info title "Test Game"
#info author "Test Author"
#info version "1.0"
#info description "A test game"
#info custom_key "custom value"

// Game code here
`),
		},
	}

	p := NewPreprocessor(".", loader)
	result, err := p.Process("test.tfy")
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Check metadata
	meta := p.GetMetadata()
	if meta.Title != "Test Game" {
		t.Errorf("Expected title 'Test Game', got '%s'", meta.Title)
	}
	if meta.Author != "Test Author" {
		t.Errorf("Expected author 'Test Author', got '%s'", meta.Author)
	}
	if meta.Version != "1.0" {
		t.Errorf("Expected version '1.0', got '%s'", meta.Version)
	}
	if meta.Description != "A test game" {
		t.Errorf("Expected description 'A test game', got '%s'", meta.Description)
	}
	if meta.Custom["custom_key"] != "custom value" {
		t.Errorf("Expected custom_key 'custom value', got '%s'", meta.Custom["custom_key"])
	}

	// Check that #info directives are removed from output
	if strings.Contains(result, "#info") {
		t.Error("#info directives should be removed from output")
	}

	// Check that code is preserved
	if !strings.Contains(result, "// Game code here") {
		t.Error("Game code should be preserved")
	}
}

func TestPreprocessor_IncludeDirective(t *testing.T) {
	loader := &MockAssetLoader{
		Files: map[string][]byte{
			"main.tfy": []byte(`// Main file
#include "lib.tfy"
// After include
`),
			"lib.tfy": []byte(`// Library file
function test() {}
`),
		},
	}

	p := NewPreprocessor(".", loader)
	result, err := p.Process("main.tfy")
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Check that include is expanded
	if !strings.Contains(result, "// Library file") {
		t.Error("Included file content should be present")
	}
	if !strings.Contains(result, "function test()") {
		t.Error("Included function should be present")
	}
	if !strings.Contains(result, "// Main file") {
		t.Error("Main file content should be preserved")
	}
	if !strings.Contains(result, "// After include") {
		t.Error("Content after include should be preserved")
	}

	// Check that #include directive is removed
	if strings.Contains(result, "#include") {
		t.Error("#include directive should be removed from output")
	}
}

func TestPreprocessor_CircularInclude(t *testing.T) {
	loader := &MockAssetLoader{
		Files: map[string][]byte{
			"a.tfy": []byte(`#include "b.tfy"`),
			"b.tfy": []byte(`#include "a.tfy"`),
		},
	}

	p := NewPreprocessor(".", loader)
	_, err := p.Process("a.tfy")
	if err == nil {
		t.Error("Expected error for circular include")
	}
	if !strings.Contains(err.Error(), "circular include") {
		t.Errorf("Expected circular include error, got: %v", err)
	}
}

func TestPreprocessor_ShiftJISConversion(t *testing.T) {
	// Create Shift-JIS encoded text
	utf8Text := "こんにちは世界"
	encoder := japanese.ShiftJIS.NewEncoder()
	shiftJISData, _, err := transform.Bytes(encoder, []byte(utf8Text))
	if err != nil {
		t.Fatalf("Failed to encode to Shift-JIS: %v", err)
	}

	loader := &MockAssetLoader{
		Files: map[string][]byte{
			"sjis.tfy": shiftJISData,
		},
	}

	p := NewPreprocessor(".", loader)
	result, err := p.Process("sjis.tfy")
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	if !strings.Contains(result, utf8Text) {
		t.Errorf("Expected UTF-8 text '%s' in result", utf8Text)
	}
}

func TestPreprocessor_UTF8Passthrough(t *testing.T) {
	utf8Text := "Hello World こんにちは"
	loader := &MockAssetLoader{
		Files: map[string][]byte{
			"utf8.tfy": []byte(utf8Text),
		},
	}

	p := NewPreprocessor(".", loader)
	result, err := p.Process("utf8.tfy")
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	if !strings.Contains(result, utf8Text) {
		t.Errorf("Expected UTF-8 text '%s' in result", utf8Text)
	}
}

func TestPreprocessor_NestedIncludes(t *testing.T) {
	loader := &MockAssetLoader{
		Files: map[string][]byte{
			"main.tfy": []byte(`// Main
#include "a.tfy"
`),
			"a.tfy": []byte(`// A
#include "b.tfy"
`),
			"b.tfy": []byte(`// B
function nested() {}
`),
		},
	}

	p := NewPreprocessor(".", loader)
	result, err := p.Process("main.tfy")
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	if !strings.Contains(result, "// Main") {
		t.Error("Main content missing")
	}
	if !strings.Contains(result, "// A") {
		t.Error("A content missing")
	}
	if !strings.Contains(result, "// B") {
		t.Error("B content missing")
	}
	if !strings.Contains(result, "function nested()") {
		t.Error("Nested function missing")
	}
}

func TestPreprocessor_MissingInclude(t *testing.T) {
	loader := &MockAssetLoader{
		Files: map[string][]byte{
			"main.tfy": []byte(`#include "missing.tfy"`),
		},
	}

	p := NewPreprocessor(".", loader)
	_, err := p.Process("main.tfy")
	if err == nil {
		t.Error("Expected error for missing include")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' error, got: %v", err)
	}
}

func TestPreprocessor_InvalidInfoDirective(t *testing.T) {
	loader := &MockAssetLoader{
		Files: map[string][]byte{
			"test.tfy": []byte(`#info invalid`),
		},
	}

	p := NewPreprocessor(".", loader)
	_, err := p.Process("test.tfy")
	if err == nil {
		t.Error("Expected error for invalid #info directive")
	}
}

func TestPreprocessor_InvalidIncludeDirective(t *testing.T) {
	loader := &MockAssetLoader{
		Files: map[string][]byte{
			"test.tfy": []byte(`#include`),
		},
	}

	p := NewPreprocessor(".", loader)
	_, err := p.Process("test.tfy")
	if err == nil {
		t.Error("Expected error for invalid #include directive")
	}
}

func TestPreprocessor_MixedDirectives(t *testing.T) {
	loader := &MockAssetLoader{
		Files: map[string][]byte{
			"main.tfy": []byte(`#info title "Mixed Test"
// Code before include
#include "lib.tfy"
// Code after include
#info author "Test"
`),
			"lib.tfy": []byte(`// Library
#info version "1.0"
function lib() {}
`),
		},
	}

	p := NewPreprocessor(".", loader)
	result, err := p.Process("main.tfy")
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Check metadata from both files
	meta := p.GetMetadata()
	if meta.Title != "Mixed Test" {
		t.Errorf("Expected title 'Mixed Test', got '%s'", meta.Title)
	}
	if meta.Author != "Test" {
		t.Errorf("Expected author 'Test', got '%s'", meta.Author)
	}
	if meta.Version != "1.0" {
		t.Errorf("Expected version '1.0', got '%s'", meta.Version)
	}

	// Check code is preserved
	if !strings.Contains(result, "// Code before include") {
		t.Error("Code before include missing")
	}
	if !strings.Contains(result, "// Code after include") {
		t.Error("Code after include missing")
	}
	if !strings.Contains(result, "function lib()") {
		t.Error("Library function missing")
	}
}

func TestIsValidUTF8(t *testing.T) {
	tests := []struct {
		name  string
		data  []byte
		valid bool
	}{
		{"ASCII", []byte("Hello World"), true},
		{"UTF-8 Japanese", []byte("こんにちは"), true},
		{"Invalid with null", []byte{0x00, 0x41}, false},
		{"Invalid sequence", []byte{0xFF, 0xFE}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidUTF8(tt.data)
			if result != tt.valid {
				t.Errorf("Expected %v, got %v for %s", tt.valid, result, tt.name)
			}
		})
	}
}
