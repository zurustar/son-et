package graphics

import (
	"image/color"
	"testing"
)

func TestNewWindowManager(t *testing.T) {
	wm := NewWindowManager()
	if wm == nil {
		t.Fatal("NewWindowManager returned nil")
	}
	if wm.nextID != 0 {
		t.Errorf("Expected nextID to be 0, got %d", wm.nextID)
	}
	if wm.maxID != 64 {
		t.Errorf("Expected maxID to be 64, got %d", wm.maxID)
	}
	if len(wm.windows) != 0 {
		t.Errorf("Expected empty windows map, got %d windows", len(wm.windows))
	}
}

func TestOpenWin(t *testing.T) {
	wm := NewWindowManager()

	// Test opening a window with default options
	winID, err := wm.OpenWin(1)
	if err != nil {
		t.Fatalf("OpenWin failed: %v", err)
	}
	if winID != 0 {
		t.Errorf("Expected first window ID to be 0, got %d", winID)
	}

	// Verify window was created
	win, err := wm.GetWin(winID)
	if err != nil {
		t.Fatalf("GetWin failed: %v", err)
	}
	if win.PicID != 1 {
		t.Errorf("Expected PicID to be 1, got %d", win.PicID)
	}
	if !win.Visible {
		t.Error("Expected window to be visible")
	}
	if win.ZOrder != 0 {
		t.Errorf("Expected ZOrder to be 0, got %d", win.ZOrder)
	}
}

func TestOpenWinWithOptions(t *testing.T) {
	wm := NewWindowManager()

	// Test opening a window with custom options
	winID, err := wm.OpenWin(1,
		WithPosition(10, 20),
		WithSize(100, 200),
		WithPicOffset(5, 10),
		WithBgColor(color.RGBA{255, 0, 0, 255}),
		WithCaption("Test Window"),
	)
	if err != nil {
		t.Fatalf("OpenWin failed: %v", err)
	}

	win, err := wm.GetWin(winID)
	if err != nil {
		t.Fatalf("GetWin failed: %v", err)
	}

	if win.X != 10 || win.Y != 20 {
		t.Errorf("Expected position (10, 20), got (%d, %d)", win.X, win.Y)
	}
	if win.Width != 100 || win.Height != 200 {
		t.Errorf("Expected size (100, 200), got (%d, %d)", win.Width, win.Height)
	}
	if win.PicX != 5 || win.PicY != 10 {
		t.Errorf("Expected pic offset (5, 10), got (%d, %d)", win.PicX, win.PicY)
	}
	if win.Caption != "Test Window" {
		t.Errorf("Expected caption 'Test Window', got '%s'", win.Caption)
	}
}

func TestOpenWinResourceLimit(t *testing.T) {
	wm := NewWindowManager()

	// Open maximum number of windows
	for i := 0; i < 64; i++ {
		_, err := wm.OpenWin(i)
		if err != nil {
			t.Fatalf("Failed to open window %d: %v", i, err)
		}
	}

	// Try to open one more window (should fail)
	_, err := wm.OpenWin(64)
	if err == nil {
		t.Error("Expected error when exceeding window limit, got nil")
	}
}

func TestMoveWin(t *testing.T) {
	wm := NewWindowManager()

	// Create a window
	winID, err := wm.OpenWin(1)
	if err != nil {
		t.Fatalf("OpenWin failed: %v", err)
	}

	// Move the window
	err = wm.MoveWin(winID,
		WithPosition(50, 60),
		WithSize(150, 250),
		WithPicID(2),
	)
	if err != nil {
		t.Fatalf("MoveWin failed: %v", err)
	}

	// Verify changes
	win, err := wm.GetWin(winID)
	if err != nil {
		t.Fatalf("GetWin failed: %v", err)
	}

	if win.X != 50 || win.Y != 60 {
		t.Errorf("Expected position (50, 60), got (%d, %d)", win.X, win.Y)
	}
	if win.Width != 150 || win.Height != 250 {
		t.Errorf("Expected size (150, 250), got (%d, %d)", win.Width, win.Height)
	}
	if win.PicID != 2 {
		t.Errorf("Expected PicID to be 2, got %d", win.PicID)
	}
}

func TestMoveWinInvalidID(t *testing.T) {
	wm := NewWindowManager()

	// Try to move non-existent window
	err := wm.MoveWin(999, WithPosition(10, 20))
	if err == nil {
		t.Error("Expected error when moving non-existent window, got nil")
	}
}

func TestCloseWin(t *testing.T) {
	wm := NewWindowManager()

	// Create a window
	winID, err := wm.OpenWin(1)
	if err != nil {
		t.Fatalf("OpenWin failed: %v", err)
	}

	// Close the window
	err = wm.CloseWin(winID)
	if err != nil {
		t.Fatalf("CloseWin failed: %v", err)
	}

	// Verify window is gone
	_, err = wm.GetWin(winID)
	if err == nil {
		t.Error("Expected error when getting closed window, got nil")
	}
}

func TestCloseWinInvalidID(t *testing.T) {
	wm := NewWindowManager()

	// Try to close non-existent window
	err := wm.CloseWin(999)
	if err == nil {
		t.Error("Expected error when closing non-existent window, got nil")
	}
}

func TestCloseWinAll(t *testing.T) {
	wm := NewWindowManager()

	// Create multiple windows
	for i := 0; i < 5; i++ {
		_, err := wm.OpenWin(i)
		if err != nil {
			t.Fatalf("OpenWin failed: %v", err)
		}
	}

	// Close all windows
	wm.CloseWinAll()

	// Verify all windows are gone
	windows := wm.GetWindowsOrdered()
	if len(windows) != 0 {
		t.Errorf("Expected 0 windows after CloseWinAll, got %d", len(windows))
	}
}

func TestGetWin(t *testing.T) {
	wm := NewWindowManager()

	// Create a window
	winID, err := wm.OpenWin(1)
	if err != nil {
		t.Fatalf("OpenWin failed: %v", err)
	}

	// Get the window
	win, err := wm.GetWin(winID)
	if err != nil {
		t.Fatalf("GetWin failed: %v", err)
	}
	if win.ID != winID {
		t.Errorf("Expected window ID %d, got %d", winID, win.ID)
	}
}

func TestGetWinInvalidID(t *testing.T) {
	wm := NewWindowManager()

	// Try to get non-existent window
	_, err := wm.GetWin(999)
	if err == nil {
		t.Error("Expected error when getting non-existent window, got nil")
	}
}

func TestGetWindowsOrdered(t *testing.T) {
	wm := NewWindowManager()

	// Create multiple windows
	ids := make([]int, 5)
	for i := 0; i < 5; i++ {
		id, err := wm.OpenWin(i)
		if err != nil {
			t.Fatalf("OpenWin failed: %v", err)
		}
		ids[i] = id
	}

	// Get windows in Z order
	windows := wm.GetWindowsOrdered()
	if len(windows) != 5 {
		t.Fatalf("Expected 5 windows, got %d", len(windows))
	}

	// Verify Z order (should be ascending)
	for i := 0; i < len(windows)-1; i++ {
		if windows[i].ZOrder >= windows[i+1].ZOrder {
			t.Errorf("Windows not in Z order: window[%d].ZOrder=%d >= window[%d].ZOrder=%d",
				i, windows[i].ZOrder, i+1, windows[i+1].ZOrder)
		}
	}

	// Verify IDs match creation order
	for i, win := range windows {
		if win.ID != ids[i] {
			t.Errorf("Expected window ID %d at position %d, got %d", ids[i], i, win.ID)
		}
	}
}

func TestCapTitle(t *testing.T) {
	wm := NewWindowManager()

	// Create a window
	winID, err := wm.OpenWin(1)
	if err != nil {
		t.Fatalf("OpenWin failed: %v", err)
	}

	// Set caption
	err = wm.CapTitle(winID, "My Window")
	if err != nil {
		t.Fatalf("CapTitle failed: %v", err)
	}

	// Verify caption
	win, err := wm.GetWin(winID)
	if err != nil {
		t.Fatalf("GetWin failed: %v", err)
	}
	if win.Caption != "My Window" {
		t.Errorf("Expected caption 'My Window', got '%s'", win.Caption)
	}
}

func TestCapTitleInvalidID(t *testing.T) {
	wm := NewWindowManager()

	// Try to set caption on non-existent window
	err := wm.CapTitle(999, "Test")
	if err == nil {
		t.Error("Expected error when setting caption on non-existent window, got nil")
	}
}

func TestGetPicNo(t *testing.T) {
	wm := NewWindowManager()

	// Create a window
	winID, err := wm.OpenWin(42)
	if err != nil {
		t.Fatalf("OpenWin failed: %v", err)
	}

	// Get picture number
	picNo, err := wm.GetPicNo(winID)
	if err != nil {
		t.Fatalf("GetPicNo failed: %v", err)
	}
	if picNo != 42 {
		t.Errorf("Expected picture number 42, got %d", picNo)
	}
}

func TestGetPicNoInvalidID(t *testing.T) {
	wm := NewWindowManager()

	// Try to get picture number from non-existent window
	_, err := wm.GetPicNo(999)
	if err == nil {
		t.Error("Expected error when getting picture number from non-existent window, got nil")
	}
}

func TestCapTitleAll(t *testing.T) {
	wm := NewWindowManager()

	// Create multiple windows
	winIDs := make([]int, 3)
	for i := 0; i < 3; i++ {
		id, err := wm.OpenWin(i)
		if err != nil {
			t.Fatalf("OpenWin failed: %v", err)
		}
		winIDs[i] = id
	}

	// Set caption on all windows
	wm.CapTitleAll("All Windows Title")

	// Verify all windows have the same caption
	for _, winID := range winIDs {
		win, err := wm.GetWin(winID)
		if err != nil {
			t.Fatalf("GetWin failed: %v", err)
		}
		if win.Caption != "All Windows Title" {
			t.Errorf("Expected caption 'All Windows Title', got '%s' for window %d", win.Caption, winID)
		}
	}
}

func TestCapTitleAllNoWindows(t *testing.T) {
	wm := NewWindowManager()

	// Call CapTitleAll with no windows (should not panic or error)
	wm.CapTitleAll("Test Title")

	// Verify no windows exist
	windows := wm.GetWindowsOrdered()
	if len(windows) != 0 {
		t.Errorf("Expected 0 windows, got %d", len(windows))
	}
}

func TestCapTitleAllEmptyString(t *testing.T) {
	wm := NewWindowManager()

	// Create a window with a caption
	winID, err := wm.OpenWin(1, WithCaption("Initial Caption"))
	if err != nil {
		t.Fatalf("OpenWin failed: %v", err)
	}

	// Clear caption using empty string
	wm.CapTitleAll("")

	// Verify caption is cleared
	win, err := wm.GetWin(winID)
	if err != nil {
		t.Fatalf("GetWin failed: %v", err)
	}
	if win.Caption != "" {
		t.Errorf("Expected empty caption, got '%s'", win.Caption)
	}
}

// TestGetWinByPicID tests the GetWinByPicID method
// Validates: Requirements 7.5 - ピクチャーIDからウィンドウIDへの逆引きをサポートする
func TestGetWinByPicID(t *testing.T) {
	wm := NewWindowManager()

	// Create a window with a specific picture ID
	winID, err := wm.OpenWin(42)
	if err != nil {
		t.Fatalf("OpenWin failed: %v", err)
	}

	// Get window ID by picture ID
	foundWinID, err := wm.GetWinByPicID(42)
	if err != nil {
		t.Fatalf("GetWinByPicID failed: %v", err)
	}
	if foundWinID != winID {
		t.Errorf("Expected window ID %d, got %d", winID, foundWinID)
	}
}

// TestGetWinByPicIDNotFound tests GetWinByPicID with non-existent picture ID
// Validates: Requirements 7.5
func TestGetWinByPicIDNotFound(t *testing.T) {
	wm := NewWindowManager()

	// Create a window with a specific picture ID
	_, err := wm.OpenWin(42)
	if err != nil {
		t.Fatalf("OpenWin failed: %v", err)
	}

	// Try to get window ID by non-existent picture ID
	_, err = wm.GetWinByPicID(999)
	if err == nil {
		t.Error("Expected error when getting window by non-existent picture ID, got nil")
	}
}

// TestGetWinByPicIDMultipleWindows tests GetWinByPicID when multiple windows use the same picture
// Validates: Requirements 7.5
func TestGetWinByPicIDMultipleWindows(t *testing.T) {
	wm := NewWindowManager()

	// Create multiple windows with the same picture ID
	// The implementation should return the window with the highest ZOrder (last opened)
	winID1, err := wm.OpenWin(42)
	if err != nil {
		t.Fatalf("OpenWin failed: %v", err)
	}

	winID2, err := wm.OpenWin(42)
	if err != nil {
		t.Fatalf("OpenWin failed: %v", err)
	}

	// Get window ID by picture ID - should return the last opened window (highest ZOrder)
	foundWinID, err := wm.GetWinByPicID(42)
	if err != nil {
		t.Fatalf("GetWinByPicID failed: %v", err)
	}

	// Verify that the second window (higher ZOrder) is returned
	if foundWinID != winID2 {
		t.Errorf("Expected window ID %d (last opened), got %d (first opened: %d)", winID2, foundWinID, winID1)
	}
}

// TestGetWinByPicIDAfterClose tests GetWinByPicID after closing a window
// Validates: Requirements 7.5
func TestGetWinByPicIDAfterClose(t *testing.T) {
	wm := NewWindowManager()

	// Create a window
	winID, err := wm.OpenWin(42)
	if err != nil {
		t.Fatalf("OpenWin failed: %v", err)
	}

	// Verify we can find it
	foundWinID, err := wm.GetWinByPicID(42)
	if err != nil {
		t.Fatalf("GetWinByPicID failed: %v", err)
	}
	if foundWinID != winID {
		t.Errorf("Expected window ID %d, got %d", winID, foundWinID)
	}

	// Close the window
	err = wm.CloseWin(winID)
	if err != nil {
		t.Fatalf("CloseWin failed: %v", err)
	}

	// Now GetWinByPicID should fail
	_, err = wm.GetWinByPicID(42)
	if err == nil {
		t.Error("Expected error when getting window by picture ID after close, got nil")
	}
}

// TestGetWinByPicIDMultipleWindowsAfterClose tests GetWinByPicID when one of multiple windows is closed
// Validates: Requirements 7.5
func TestGetWinByPicIDMultipleWindowsAfterClose(t *testing.T) {
	wm := NewWindowManager()

	// Create multiple windows with the same picture ID
	winID1, err := wm.OpenWin(42)
	if err != nil {
		t.Fatalf("OpenWin failed: %v", err)
	}

	winID2, err := wm.OpenWin(42)
	if err != nil {
		t.Fatalf("OpenWin failed: %v", err)
	}

	// Close the second window (higher ZOrder)
	err = wm.CloseWin(winID2)
	if err != nil {
		t.Fatalf("CloseWin failed: %v", err)
	}

	// Now GetWinByPicID should return the first window
	foundWinID, err := wm.GetWinByPicID(42)
	if err != nil {
		t.Fatalf("GetWinByPicID failed: %v", err)
	}
	if foundWinID != winID1 {
		t.Errorf("Expected window ID %d (remaining window), got %d", winID1, foundWinID)
	}
}
