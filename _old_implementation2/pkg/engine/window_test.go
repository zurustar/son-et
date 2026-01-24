package engine

import (
	"testing"
)

func TestOpenWindow(t *testing.T) {
	state := NewEngineState(nil, nil, nil)

	// Create a window
	winID := state.OpenWindow(1, 100, 200, 640, 480, 0, 0, 0)

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
	if win.Caption != "" {
		t.Errorf("Expected empty Caption, got %q", win.Caption)
	}
	if !win.Visible {
		t.Error("Expected window to be visible")
	}
}

func TestMoveWindow(t *testing.T) {
	state := NewEngineState(nil, nil, nil)

	// Create a window
	winID := state.OpenWindow(1, 100, 200, 640, 480, 0, 0, 0)

	// Move the window (8 args: winID, picID, x, y, width, height, picX, picY)
	err := state.MoveWindow(winID, 2, 150, 250, 800, 600, 10, 20)
	if err != nil {
		t.Fatalf("MoveWindow failed: %v", err)
	}

	// Verify window was moved
	win := state.GetWindow(winID)
	if win.PictureID != 2 {
		t.Errorf("Expected PictureID 2, got %d", win.PictureID)
	}
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
	winID := state.OpenWindow(1, 100, 200, 640, 480, 0, 0, 0)

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
	state.OpenWindow(1, 100, 200, 640, 480, 0, 0, 0)
	state.OpenWindow(2, 150, 250, 800, 600, 0, 0, 0)
	state.OpenWindow(3, 200, 300, 1024, 768, 0, 0, 0)

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
	winID := state.OpenWindow(1, 100, 200, 640, 480, 0, 0, 0)

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
	winID := state.OpenWindow(42, 100, 200, 640, 480, 0, 0, 0)

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
	id1 := state.OpenWindow(1, 100, 100, 100, 100, 0, 0, 0)
	id2 := state.OpenWindow(2, 200, 200, 100, 100, 0, 0, 0)
	id3 := state.OpenWindow(3, 300, 300, 100, 100, 0, 0, 0)

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

func TestStartWindowDrag(t *testing.T) {
	state := NewEngineState(nil, nil, nil)

	// Create a window with a caption (draggable)
	winID := state.OpenWindow(1, 100, 100, 200, 150, 0, 0, 0)
	// Set caption to make it draggable
	state.SetWindowCaption(winID, "Draggable Window")

	// Click on the title bar
	draggedID := state.StartWindowDrag(150, 110)

	if draggedID != winID {
		t.Errorf("Expected to start dragging window %d, got %d", winID, draggedID)
	}

	// Verify window is in dragging state
	win := state.GetWindow(winID)
	if !win.IsDragging {
		t.Error("Expected window to be in dragging state")
	}

	// Verify drag offsets are set correctly
	if win.DragOffsetX != 50 { // 150 - 100
		t.Errorf("Expected DragOffsetX 50, got %d", win.DragOffsetX)
	}
	if win.DragOffsetY != 10 { // 110 - 100
		t.Errorf("Expected DragOffsetY 10, got %d", win.DragOffsetY)
	}
}

func TestStartWindowDragNoCaptionIgnored(t *testing.T) {
	state := NewEngineState(nil, nil, nil)

	// Create a window without a caption (not draggable)
	state.OpenWindow(1, 100, 100, 200, 150, 0, 0, 0)

	// Try to click on where the title bar would be
	draggedID := state.StartWindowDrag(150, 110)

	if draggedID != 0 {
		t.Errorf("Expected no window to be dragged (no caption), got window %d", draggedID)
	}
}

func TestStartWindowDragOutsideTitleBar(t *testing.T) {
	state := NewEngineState(nil, nil, nil)

	// Create a window with a caption
	winID := state.OpenWindow(1, 100, 100, 200, 150, 0, 0, 0)
	state.SetWindowCaption(winID, "Window")

	// Click below the title bar (in the window content area)
	draggedID := state.StartWindowDrag(150, 130)

	if draggedID != 0 {
		t.Errorf("Expected no window to be dragged (clicked outside title bar), got window %d", draggedID)
	}
}

func TestStartWindowDragTopmost(t *testing.T) {
	state := NewEngineState(nil, nil, nil)

	// Create overlapping windows
	win1 := state.OpenWindow(1, 100, 100, 200, 150, 0, 0, 0)
	state.SetWindowCaption(win1, "Window 1")
	win2 := state.OpenWindow(2, 150, 150, 200, 150, 0, 0, 0)
	state.SetWindowCaption(win2, "Window 2")

	// Click on overlapping area (should select the topmost window = win2)
	draggedID := state.StartWindowDrag(180, 160)

	if draggedID != win2 {
		t.Errorf("Expected to drag topmost window %d, got %d", win2, draggedID)
	}

	// Verify only win2 is dragging
	if state.GetWindow(win1).IsDragging {
		t.Error("Expected window 1 to not be dragging")
	}
	if !state.GetWindow(win2).IsDragging {
		t.Error("Expected window 2 to be dragging")
	}
}

func TestUpdateWindowDrag(t *testing.T) {
	state := NewEngineState(nil, nil, nil)

	// Create and start dragging a window
	winID := state.OpenWindow(1, 100, 100, 200, 150, 0, 0, 0)
	state.SetWindowCaption(winID, "Window")
	state.StartWindowDrag(150, 110) // Click at (150, 110), offset = (50, 10)

	// Move mouse to new position
	updated := state.UpdateWindowDrag(200, 150)

	if !updated {
		t.Error("Expected window drag to be updated")
	}

	// Verify window moved to new position (accounting for offset)
	win := state.GetWindow(winID)
	expectedX := 200 - 50 // 150
	expectedY := 150 - 10 // 140
	if win.X != expectedX {
		t.Errorf("Expected X %d, got %d", expectedX, win.X)
	}
	if win.Y != expectedY {
		t.Errorf("Expected Y %d, got %d", expectedY, win.Y)
	}
}

func TestUpdateWindowDragNoDrag(t *testing.T) {
	state := NewEngineState(nil, nil, nil)

	// Create a window but don't start dragging
	state.OpenWindow(1, 100, 100, 200, 150, 0, 0, 0)

	// Try to update drag
	updated := state.UpdateWindowDrag(200, 150)

	if updated {
		t.Error("Expected no update when no window is being dragged")
	}
}

func TestUpdateWindowDragConstraints(t *testing.T) {
	state := NewEngineState(nil, nil, nil)

	// Create a window
	winID := state.OpenWindow(1, 100, 100, 200, 150, 0, 0, 0)
	state.SetWindowCaption(winID, "Window")
	state.StartWindowDrag(150, 110) // offset = (50, 10)

	tests := []struct {
		name      string
		mouseX    int
		mouseY    int
		expectedX int
		expectedY int
	}{
		{
			name:      "constrain top",
			mouseX:    150,
			mouseY:    5, // Would put window at Y=-5
			expectedX: 100,
			expectedY: 0, // Constrained to 0
		},
		{
			name:      "constrain bottom",
			mouseX:    150,
			mouseY:    800, // Would put window at Y=790
			expectedX: 100,
			expectedY: VirtualDesktopHeight - TitleBarHeight, // Constrained
		},
		{
			name:      "constrain left (partial)",
			mouseX:    -110, // With offset 50, would put window at X=-160
			mouseY:    110,
			expectedX: -(200 - 50), // Constrained to -150
			expectedY: 100,
		},
		{
			name:      "constrain right (partial)",
			mouseX:    VirtualDesktopWidth + 100,
			mouseY:    110,
			expectedX: VirtualDesktopWidth - 50, // Keep 50px visible
			expectedY: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset window position
			win := state.GetWindow(winID)
			win.X = 100
			win.Y = 100
			win.IsDragging = true
			win.DragOffsetX = 50
			win.DragOffsetY = 10
			state.draggedWindowID = winID

			// Update drag
			state.UpdateWindowDrag(tt.mouseX, tt.mouseY)

			// Check constraints
			if win.X != tt.expectedX {
				t.Errorf("Expected X %d, got %d", tt.expectedX, win.X)
			}
			if win.Y != tt.expectedY {
				t.Errorf("Expected Y %d, got %d", tt.expectedY, win.Y)
			}
		})
	}
}

func TestStopWindowDrag(t *testing.T) {
	state := NewEngineState(nil, nil, nil)

	// Create and start dragging a window
	winID := state.OpenWindow(1, 100, 100, 200, 150, 0, 0, 0)
	state.SetWindowCaption(winID, "Window")
	state.StartWindowDrag(150, 110)

	// Verify dragging started
	if state.GetDraggedWindowID() != winID {
		t.Error("Expected window to be dragging")
	}

	// Stop dragging
	state.StopWindowDrag()

	// Verify dragging stopped
	if state.GetDraggedWindowID() != 0 {
		t.Error("Expected no window to be dragging")
	}

	win := state.GetWindow(winID)
	if win.IsDragging {
		t.Error("Expected window IsDragging to be false")
	}
}

func TestGetDraggedWindowID(t *testing.T) {
	state := NewEngineState(nil, nil, nil)

	// Initially no window is dragged
	if state.GetDraggedWindowID() != 0 {
		t.Error("Expected no window to be dragged initially")
	}

	// Start dragging
	winID := state.OpenWindow(1, 100, 100, 200, 150, 0, 0, 0)
	state.SetWindowCaption(winID, "Window")
	state.StartWindowDrag(150, 110)

	// Verify correct window ID is returned
	if state.GetDraggedWindowID() != winID {
		t.Errorf("Expected dragged window ID %d, got %d", winID, state.GetDraggedWindowID())
	}

	// Stop dragging
	state.StopWindowDrag()

	// Verify no window is dragged
	if state.GetDraggedWindowID() != 0 {
		t.Error("Expected no window to be dragged after stop")
	}
}
