package engine

import (
	"image"
	"image/color"
	"testing"
)

// TestMoveCastWithBackBuffer tests that MoveCast uses double-buffering correctly
func TestMoveCastWithBackBuffer(t *testing.T) {
	state := NewEngineState(nil, nil, nil)

	// Create a destination picture
	destPic := &Picture{
		ID:     1,
		Image:  image.NewRGBA(image.Rect(0, 0, 200, 200)),
		Width:  200,
		Height: 200,
	}
	state.pictures[1] = destPic

	// Create a source picture (red square)
	srcPic := &Picture{
		ID:     2,
		Image:  createColoredImage(64, 64, color.RGBA{255, 0, 0, 255}),
		Width:  64,
		Height: 64,
	}
	state.pictures[2] = srcPic

	// Create a cast
	castID := state.PutCast(1, 2, 50, 50, 0, 0, 64, 64, -1)

	// Move the cast - this should create BackBuffer
	err := state.MoveCast(castID, 100, 100, -1, -1, -1, -1)
	if err != nil {
		t.Fatalf("MoveCast failed: %v", err)
	}

	// Verify BackBuffer was created after MoveCast
	if destPic.BackBuffer == nil {
		t.Error("Expected BackBuffer to be created after MoveCast")
	}

	// Verify cast position was updated
	cast := state.GetCast(castID)
	if cast.X != 100 || cast.Y != 100 {
		t.Errorf("Expected cast at (100,100), got (%d,%d)", cast.X, cast.Y)
	}
}

// TestMoveCastMultipleCasts tests that MoveCast redraws all casts
func TestMoveCastMultipleCasts(t *testing.T) {
	state := NewEngineState(nil, nil, nil)

	// Create a destination picture
	destPic := &Picture{
		ID:     1,
		Image:  image.NewRGBA(image.Rect(0, 0, 300, 300)),
		Width:  300,
		Height: 300,
	}
	state.pictures[1] = destPic

	// Create source pictures
	redPic := &Picture{
		ID:     2,
		Image:  createColoredImage(64, 64, color.RGBA{255, 0, 0, 255}),
		Width:  64,
		Height: 64,
	}
	state.pictures[2] = redPic

	bluePic := &Picture{
		ID:     3,
		Image:  createColoredImage(64, 64, color.RGBA{0, 0, 255, 255}),
		Width:  64,
		Height: 64,
	}
	state.pictures[3] = bluePic

	// Create two casts
	cast1 := state.PutCast(1, 2, 50, 50, 0, 0, 64, 64, -1)
	cast2 := state.PutCast(1, 3, 100, 100, 0, 0, 64, 64, -1)

	// Move cast1 - this should redraw both casts
	err := state.MoveCast(cast1, 150, 150, -1, -1, -1, -1)
	if err != nil {
		t.Fatalf("MoveCast failed: %v", err)
	}

	// Verify both casts still exist
	c1 := state.GetCast(cast1)
	c2 := state.GetCast(cast2)

	if c1 == nil || c2 == nil {
		t.Fatal("Expected both casts to still exist")
	}

	// Verify cast1 position was updated
	if c1.X != 150 || c1.Y != 150 {
		t.Errorf("Expected cast1 at (150,150), got (%d,%d)", c1.X, c1.Y)
	}

	// Verify cast2 position unchanged
	if c2.X != 100 || c2.Y != 100 {
		t.Errorf("Expected cast2 at (100,100), got (%d,%d)", c2.X, c2.Y)
	}
}

// Helper function to create a colored image
func createColoredImage(width, height int, c color.Color) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, c)
		}
	}
	return img
}
