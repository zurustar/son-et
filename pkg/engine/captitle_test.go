package engine

import (
	"testing"
)

func TestCapTitle_EmptyString(t *testing.T) {
	// Initialize engine state
	engine := NewEngineState()
	globalEngine = engine

	// Test 1: CapTitle("") should set empty title
	CapTitle("")
	if engine.globalWindowTitle != "" {
		t.Errorf("Expected empty title, got '%s'", engine.globalWindowTitle)
	}

	// Test 2: OpenWin should use the empty title
	// Load a test picture first
	pic := &Picture{
		ID:     0,
		Width:  100,
		Height: 100,
	}
	engine.pictures[0] = pic

	winID := engine.OpenWin(0, 0, 0, 100, 100, 0, 0, 0xFFFFFF)
	win := engine.windows[winID]

	if win.Title != "" {
		t.Errorf("Expected window title to be empty, got '%s'", win.Title)
	}
}

func TestCapTitle_NonEmptyString(t *testing.T) {
	// Initialize engine state
	engine := NewEngineState()
	globalEngine = engine

	// Test: CapTitle("Test Title") should set title
	CapTitle("Test Title")
	if engine.globalWindowTitle != "Test Title" {
		t.Errorf("Expected 'Test Title', got '%s'", engine.globalWindowTitle)
	}

	// OpenWin should use the title
	pic := &Picture{
		ID:     0,
		Width:  100,
		Height: 100,
	}
	engine.pictures[0] = pic

	winID := engine.OpenWin(0, 0, 0, 100, 100, 0, 0, 0xFFFFFF)
	win := engine.windows[winID]

	if win.Title != "Test Title" {
		t.Errorf("Expected window title to be 'Test Title', got '%s'", win.Title)
	}
}

func TestCapTitle_SpecificWindow(t *testing.T) {
	// Initialize engine state
	engine := NewEngineState()
	globalEngine = engine

	// Create a window first
	pic := &Picture{
		ID:     0,
		Width:  100,
		Height: 100,
	}
	engine.pictures[0] = pic

	winID := engine.OpenWin(0, 0, 0, 100, 100, 0, 0, 0xFFFFFF)

	// Test: CapTitle(winID, "New Title") should update specific window
	CapTitle(winID, "New Title")
	win := engine.windows[winID]

	if win.Title != "New Title" {
		t.Errorf("Expected window title to be 'New Title', got '%s'", win.Title)
	}
}
