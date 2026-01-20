package engine

import (
	"testing"
)

func TestOpenWindow(t *testing.T) {
	state := NewEngineState(nil, nil, nil)

	// Create a window
	winID := state.OpenWindow(1, 100, 200, 640, 480, 0, 0, "Test Window")

	if winID != 1 {
		t.Errorf("Expected window ID 1, got %d", winID)
	}

	// Verify window was created
	win := state.GetWindow(winID)
	if win == nil {
		t.Fatal("Window not found")
	}

	if win.PictureID != 1 {
		t.Errorf("Expected PictureID 1, got %d", win.PictureID)
	}
	if win.X != 100 {
		t.Errorf("Expected X 100, got %d", win.X)
	}
	if win.Y != 200 {
		t.Errorf("Expected Y 200, got %d", win.Y)
	}
	if win.Width != 640 {
		t.Errorf("Expected Width 640, got %d", win.Width)
	}
	if win.Height != 480 {
		t.Errorf("Expected Height 480, got %d", win.Height)
	}
	if win.Caption != "Test Window" {
		t.Errorf("Expected Caption 'Test Window', got %q", win.Caption)
	}
	if !win.Visible {
		t.Error("Expected window to be visible")
	}
}

func TestMoveWindow(t *testing.T) {
	state := NewEngineState(nil, nil, nil)

	// Create a window
	winID := state.OpenWindow(1, 100, 200, 640, 480, 0, 0, "Test")

	// Move the window
	err := state.MoveWindow(winID, 150, 250, 800, 600, 10, 20)
	if err != nil {
		t.Fatalf("MoveWindow failed: %v", err)
	}

	// Verify window was moved
	win := state.GetWindow(winID)
	if win.X != 150 {
		t.Errorf("Expected X 150, got %d", win.X)
	}
	if win.Y != 250 {
		t.Errorf("Expected Y 250, got %d", win.Y)
	}
	if win.Width != 800 {
		t.Errorf("Expected Width 800, got %d", win.Width)
	}
	if win.Height != 600 {
		t.Errorf("Expected Height 600, got %d", win.Height)
	}
	if win.PicX != 10 {
		t.Errorf("Expected PicX 10, got %d", win.PicX)
	}
	if win.PicY != 20 {
		t.Errorf("Expected PicY 20, got %d", win.PicY)
	}
}

func TestCloseWindow(t *testing.T) {
	state := NewEngineState(nil, nil, nil)

	// Create a window
	winID := state.OpenWindow(1, 100, 200, 640, 480, 0, 0, "Test")

	// Close the window
	state.CloseWindow(winID)

	// Verify window was closed
	win := state.GetWindow(winID)
	if win != nil {
		t.Error("Expected window to be closed")
	}
}

func TestCloseAllWindows(t *testing.T) {
	state := NewEngineState(nil, nil, nil)

	// Create multiple windows
	state.OpenWindow(1, 100, 200, 640, 480, 0, 0, "Window 1")
	state.OpenWindow(2, 150, 250, 800, 600, 0, 0, "Window 2")
	state.OpenWindow(3, 200, 300, 1024, 768, 0, 0, "Window 3")

	// Close all windows
	state.CloseAllWindows()

	// Verify all windows were closed
	windows := state.GetWindows()
	if len(windows) != 0 {
		t.Errorf("Expected 0 windows, got %d", len(windows))
	}
}

func TestSetWindowCaption(t *testing.T) {
	state := NewEngineState(nil, nil, nil)

	// Create a window
	winID := state.OpenWindow(1, 100, 200, 640, 480, 0, 0, "Original Caption")

	// Set new caption
	err := state.SetWindowCaption(winID, "New Caption")
	if err != nil {
		t.Fatalf("SetWindowCaption failed: %v", err)
	}

	// Verify caption was changed
	win := state.GetWindow(winID)
	if win.Caption != "New Caption" {
		t.Errorf("Expected Caption 'New Caption', got %q", win.Caption)
	}
}

func TestGetWindowPictureID(t *testing.T) {
	state := NewEngineState(nil, nil, nil)

	// Create a window
	winID := state.OpenWindow(42, 100, 200, 640, 480, 0, 0, "Test")

	// Get picture ID
	picID := state.GetWindowPictureID(winID)
	if picID != 42 {
		t.Errorf("Expected PictureID 42, got %d", picID)
	}

	// Test non-existent window
	picID = state.GetWindowPictureID(999)
	if picID != 0 {
		t.Errorf("Expected PictureID 0 for non-existent window, got %d", picID)
	}
}

func TestGetWindowsOrder(t *testing.T) {
	state := NewEngineState(nil, nil, nil)

	// Create windows in specific order
	id1 := state.OpenWindow(1, 100, 100, 100, 100, 0, 0, "Window 1")
	id2 := state.OpenWindow(2, 200, 200, 100, 100, 0, 0, "Window 2")
	id3 := state.OpenWindow(3, 300, 300, 100, 100, 0, 0, "Window 3")

	// Get windows
	windows := state.GetWindows()

	// Verify order (should be in creation order)
	if len(windows) != 3 {
		t.Fatalf("Expected 3 windows, got %d", len(windows))
	}

	if windows[0].ID != id1 {
		t.Errorf("Expected first window ID %d, got %d", id1, windows[0].ID)
	}
	if windows[1].ID != id2 {
		t.Errorf("Expected second window ID %d, got %d", id2, windows[1].ID)
	}
	if windows[2].ID != id3 {
		t.Errorf("Expected third window ID %d, got %d", id3, windows[2].ID)
	}
}
