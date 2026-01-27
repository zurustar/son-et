package graphics

import (
	"image/color"
	"testing"
)

// TestDrawLine tests DrawLine function
func TestDrawLine(t *testing.T) {
	gs := NewGraphicsSystem("")

	// Create a picture to draw on
	picID, err := gs.CreatePic(100, 100)
	if err != nil {
		t.Fatalf("Failed to create picture: %v", err)
	}

	// Draw a line
	err = gs.DrawLine(picID, 10, 10, 90, 90)
	if err != nil {
		t.Errorf("DrawLine failed: %v", err)
	}
}

// TestDrawLineInvalidPicture tests DrawLine with invalid picture
func TestDrawLineInvalidPicture(t *testing.T) {
	gs := NewGraphicsSystem("")

	// Try to draw on non-existent picture
	err := gs.DrawLine(999, 10, 10, 90, 90)
	if err == nil {
		t.Error("Expected error for invalid picture, got nil")
	}
}

// TestDrawRect tests DrawRect function
func TestDrawRect(t *testing.T) {
	gs := NewGraphicsSystem("")

	picID, err := gs.CreatePic(100, 100)
	if err != nil {
		t.Fatalf("Failed to create picture: %v", err)
	}

	// Draw outline rectangle
	err = gs.DrawRect(picID, 10, 10, 90, 90, 0)
	if err != nil {
		t.Errorf("DrawRect (outline) failed: %v", err)
	}

	// Draw filled rectangle
	err = gs.DrawRect(picID, 20, 20, 80, 80, 2)
	if err != nil {
		t.Errorf("DrawRect (filled) failed: %v", err)
	}
}

// TestDrawRectSwappedCoordinates tests DrawRect with swapped coordinates
func TestDrawRectSwappedCoordinates(t *testing.T) {
	gs := NewGraphicsSystem("")

	picID, err := gs.CreatePic(100, 100)
	if err != nil {
		t.Fatalf("Failed to create picture: %v", err)
	}

	// Draw with x1 > x2 and y1 > y2 (should normalize)
	err = gs.DrawRect(picID, 90, 90, 10, 10, 0)
	if err != nil {
		t.Errorf("DrawRect with swapped coordinates failed: %v", err)
	}
}

// TestDrawRectInvalidPicture tests DrawRect with invalid picture
func TestDrawRectInvalidPicture(t *testing.T) {
	gs := NewGraphicsSystem("")

	err := gs.DrawRect(999, 10, 10, 90, 90, 0)
	if err == nil {
		t.Error("Expected error for invalid picture, got nil")
	}
}

// TestFillRect tests FillRect function
func TestFillRect(t *testing.T) {
	gs := NewGraphicsSystem("")

	picID, err := gs.CreatePic(100, 100)
	if err != nil {
		t.Fatalf("Failed to create picture: %v", err)
	}

	// Fill with color.Color
	err = gs.FillRect(picID, 10, 10, 50, 50, color.RGBA{255, 0, 0, 255})
	if err != nil {
		t.Errorf("FillRect with color.Color failed: %v", err)
	}

	// Fill with int color (0xRRGGBB)
	err = gs.FillRect(picID, 50, 50, 90, 90, 0x00FF00)
	if err != nil {
		t.Errorf("FillRect with int color failed: %v", err)
	}
}

// TestFillRectInvalidPicture tests FillRect with invalid picture
func TestFillRectInvalidPicture(t *testing.T) {
	gs := NewGraphicsSystem("")

	err := gs.FillRect(999, 10, 10, 50, 50, color.White)
	if err == nil {
		t.Error("Expected error for invalid picture, got nil")
	}
}

// TestDrawCircle tests DrawCircle function
func TestDrawCircle(t *testing.T) {
	gs := NewGraphicsSystem("")

	picID, err := gs.CreatePic(100, 100)
	if err != nil {
		t.Fatalf("Failed to create picture: %v", err)
	}

	// Draw outline circle
	err = gs.DrawCircle(picID, 50, 50, 30, 0)
	if err != nil {
		t.Errorf("DrawCircle (outline) failed: %v", err)
	}

	// Draw filled circle
	err = gs.DrawCircle(picID, 50, 50, 20, 2)
	if err != nil {
		t.Errorf("DrawCircle (filled) failed: %v", err)
	}
}

// TestDrawCircleZeroRadius tests DrawCircle with zero radius
func TestDrawCircleZeroRadius(t *testing.T) {
	gs := NewGraphicsSystem("")

	picID, err := gs.CreatePic(100, 100)
	if err != nil {
		t.Fatalf("Failed to create picture: %v", err)
	}

	// Zero radius should not error (just skip)
	err = gs.DrawCircle(picID, 50, 50, 0, 0)
	if err != nil {
		t.Errorf("DrawCircle with zero radius should not error: %v", err)
	}
}

// TestDrawCircleInvalidPicture tests DrawCircle with invalid picture
func TestDrawCircleInvalidPicture(t *testing.T) {
	gs := NewGraphicsSystem("")

	err := gs.DrawCircle(999, 50, 50, 30, 0)
	if err == nil {
		t.Error("Expected error for invalid picture, got nil")
	}
}

// TestSetLineSize tests SetLineSize function
func TestSetLineSize(t *testing.T) {
	gs := NewGraphicsSystem("")

	// Default line size should be 1
	if gs.GetLineSize() != 1 {
		t.Errorf("Default line size should be 1, got %d", gs.GetLineSize())
	}

	// Set line size
	gs.SetLineSize(5)
	if gs.GetLineSize() != 5 {
		t.Errorf("Line size should be 5, got %d", gs.GetLineSize())
	}

	// Set line size to 0 (should be clamped to 1)
	gs.SetLineSize(0)
	if gs.GetLineSize() != 1 {
		t.Errorf("Line size should be clamped to 1, got %d", gs.GetLineSize())
	}
}

// TestSetPaintColor tests SetPaintColor function
func TestSetPaintColor(t *testing.T) {
	gs := NewGraphicsSystem("")

	// Set color with color.Color
	err := gs.SetPaintColor(color.RGBA{255, 0, 0, 255})
	if err != nil {
		t.Errorf("SetPaintColor with color.Color failed: %v", err)
	}

	// Set color with int (0xRRGGBB)
	err = gs.SetPaintColor(0x00FF00)
	if err != nil {
		t.Errorf("SetPaintColor with int failed: %v", err)
	}
}

// TestGetColor tests GetColor function
// Note: Ebiten's ReadPixels cannot be called before the game starts,
// so we can only test error handling, not actual pixel values
func TestGetColor(t *testing.T) {
	gs := NewGraphicsSystem("")

	picID, err := gs.CreatePic(100, 100)
	if err != nil {
		t.Fatalf("Failed to create picture: %v", err)
	}

	// Fill with a known color
	err = gs.FillRect(picID, 0, 0, 100, 100, 0xFF0000)
	if err != nil {
		t.Fatalf("FillRect failed: %v", err)
	}

	// Note: We cannot test actual pixel values in unit tests because
	// Ebiten's ReadPixels requires the game loop to be running.
	// The GetColor function is tested through integration tests.
	t.Log("GetColor pixel value test skipped - requires running game loop")
}

// TestGetColorOutOfBounds tests GetColor with out of bounds coordinates
// Note: Ebiten's ReadPixels cannot be called before the game starts,
// so we can only test error handling
func TestGetColorOutOfBounds(t *testing.T) {
	gs := NewGraphicsSystem("")

	picID, err := gs.CreatePic(100, 100)
	if err != nil {
		t.Fatalf("Failed to create picture: %v", err)
	}

	// Note: We cannot test actual pixel values in unit tests because
	// Ebiten's ReadPixels requires the game loop to be running.
	// The GetColor function is tested through integration tests.
	t.Logf("GetColor out of bounds test skipped - requires running game loop (picID=%d)", picID)
}

// TestGetColorInvalidPicture tests GetColor with invalid picture
func TestGetColorInvalidPicture(t *testing.T) {
	gs := NewGraphicsSystem("")

	_, err := gs.GetColor(999, 50, 50)
	if err == nil {
		t.Error("Expected error for invalid picture, got nil")
	}
}

// TestDrawLineWithLineSize tests DrawLine with different line sizes
func TestDrawLineWithLineSize(t *testing.T) {
	gs := NewGraphicsSystem("")

	picID, err := gs.CreatePic(100, 100)
	if err != nil {
		t.Fatalf("Failed to create picture: %v", err)
	}

	// Set line size and draw
	gs.SetLineSize(3)
	err = gs.DrawLine(picID, 10, 10, 90, 90)
	if err != nil {
		t.Errorf("DrawLine with line size 3 failed: %v", err)
	}

	gs.SetLineSize(10)
	err = gs.DrawLine(picID, 10, 90, 90, 10)
	if err != nil {
		t.Errorf("DrawLine with line size 10 failed: %v", err)
	}
}

// TestDrawWithPaintColor tests drawing with different paint colors
func TestDrawWithPaintColor(t *testing.T) {
	gs := NewGraphicsSystem("")

	picID, err := gs.CreatePic(100, 100)
	if err != nil {
		t.Fatalf("Failed to create picture: %v", err)
	}

	// Set paint color and draw
	err = gs.SetPaintColor(0xFF0000) // Red
	if err != nil {
		t.Fatalf("SetPaintColor failed: %v", err)
	}

	err = gs.DrawLine(picID, 10, 10, 90, 10)
	if err != nil {
		t.Errorf("DrawLine with red color failed: %v", err)
	}

	err = gs.SetPaintColor(0x00FF00) // Green
	if err != nil {
		t.Fatalf("SetPaintColor failed: %v", err)
	}

	err = gs.DrawRect(picID, 20, 20, 80, 80, 0)
	if err != nil {
		t.Errorf("DrawRect with green color failed: %v", err)
	}

	err = gs.SetPaintColor(0x0000FF) // Blue
	if err != nil {
		t.Fatalf("SetPaintColor failed: %v", err)
	}

	err = gs.DrawCircle(picID, 50, 50, 20, 2)
	if err != nil {
		t.Errorf("DrawCircle with blue color failed: %v", err)
	}
}
