package engine

import (
	"testing"
)

// TestVirtualDesktopDimensions verifies that the virtual desktop is always 1280×720.
func TestVirtualDesktopDimensions(t *testing.T) {
	state := NewEngineState(nil, nil, nil)

	width := state.GetDesktopWidth()
	height := state.GetDesktopHeight()

	if width != 1280 {
		t.Errorf("Expected desktop width 1280, got %d", width)
	}

	if height != 720 {
		t.Errorf("Expected desktop height 720, got %d", height)
	}
}

// TestWinInfo_Width verifies that WinInfo(0) returns desktop width.
func TestWinInfo_Width(t *testing.T) {
	engine := NewEngine(nil, nil, nil)

	width := engine.WinInfo(0)

	if width != 1280 {
		t.Errorf("Expected WinInfo(0) = 1280, got %d", width)
	}
}

// TestWinInfo_Height verifies that WinInfo(1) returns desktop height.
func TestWinInfo_Height(t *testing.T) {
	engine := NewEngine(nil, nil, nil)

	height := engine.WinInfo(1)

	if height != 720 {
		t.Errorf("Expected WinInfo(1) = 720, got %d", height)
	}
}

// TestWinInfo_InvalidIndex verifies that WinInfo returns 0 for invalid indices.
func TestWinInfo_InvalidIndex(t *testing.T) {
	engine := NewEngine(nil, nil, nil)

	tests := []struct {
		index    int
		expected int
	}{
		{-1, 0},
		{2, 0},
		{100, 0},
	}

	for _, tt := range tests {
		result := engine.WinInfo(tt.index)
		if result != tt.expected {
			t.Errorf("WinInfo(%d) = %d, expected %d", tt.index, result, tt.expected)
		}
	}
}
