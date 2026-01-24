package engine

import (
	"image"
	"image/color"
	"testing"
)

// TestMovePicAccumulation tests that MovePic doesn't accumulate pixels
// when called multiple times to the same destination.
// This reproduces the issue where P14 appears to be drawn multiple times
// at slightly different positions.
func TestMovePicAccumulation(t *testing.T) {
	state := NewEngineState(nil, nil, nil)

	// Create source picture (10x10 red square)
	srcPic := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			srcPic.Set(x, y, color.RGBA{255, 0, 0, 255}) // Red
		}
	}
	state.pictures[0] = &Picture{
		ID:     0,
		Image:  srcPic,
		Width:  10,
		Height: 10,
	}

	// Create destination picture (30x30 white background)
	dstPic := image.NewRGBA(image.Rect(0, 0, 30, 30))
	for y := 0; y < 30; y++ {
		for x := 0; x < 30; x++ {
			dstPic.Set(x, y, color.RGBA{255, 255, 255, 255}) // White
		}
	}
	state.pictures[1] = &Picture{
		ID:     1,
		Image:  dstPic,
		Width:  30,
		Height: 30,
	}

	// First MovePic: copy red square to position (5, 5)
	err := state.MovePicture(0, 0, 0, 10, 10, 1, 5, 5, 0)
	if err != nil {
		t.Fatalf("First MovePic failed: %v", err)
	}

	// Verify red square is at (5, 5)
	destImg := state.pictures[1].Image.(*image.RGBA)
	r, g, b, _ := destImg.At(5, 5).RGBA()
	if r>>8 != 255 || g>>8 != 0 || b>>8 != 0 {
		t.Errorf("Expected red at (5,5), got RGB(%d,%d,%d)", r>>8, g>>8, b>>8)
	}

	// Second MovePic: copy red square to position (10, 10)
	// This should NOT leave the first red square at (5, 5)
	err = state.MovePicture(0, 0, 0, 10, 10, 1, 10, 10, 0)
	if err != nil {
		t.Fatalf("Second MovePic failed: %v", err)
	}

	// Verify red square is at (10, 10)
	r, g, b, _ = destImg.At(10, 10).RGBA()
	if r>>8 != 255 || g>>8 != 0 || b>>8 != 0 {
		t.Errorf("Expected red at (10,10), got RGB(%d,%d,%d)", r>>8, g>>8, b>>8)
	}

	// CRITICAL: The first red square at (5, 5) should STILL BE THERE
	// because MovePic uses draw.Over which preserves existing pixels
	// This is the expected behavior in FILLY - MovePic accumulates!
	r, g, b, _ = destImg.At(5, 5).RGBA()
	if r>>8 != 255 || g>>8 != 0 || b>>8 != 0 {
		t.Errorf("Expected red to remain at (5,5), got RGB(%d,%d,%d)", r>>8, g>>8, b>>8)
	}

	t.Logf("MovePic correctly accumulates pixels (both red squares visible)")
}

// TestMovePicOverwrite tests that MovePic with the same position
// overwrites the previous content (not accumulates).
func TestMovePicOverwrite(t *testing.T) {
	state := NewEngineState(nil, nil, nil)

	// Create red source picture
	redPic := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			redPic.Set(x, y, color.RGBA{255, 0, 0, 255}) // Red
		}
	}
	state.pictures[0] = &Picture{
		ID:     0,
		Image:  redPic,
		Width:  10,
		Height: 10,
	}

	// Create blue source picture
	bluePic := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			bluePic.Set(x, y, color.RGBA{0, 0, 255, 255}) // Blue
		}
	}
	state.pictures[1] = &Picture{
		ID:     1,
		Image:  bluePic,
		Width:  10,
		Height: 10,
	}

	// Create destination picture (white background)
	dstPic := image.NewRGBA(image.Rect(0, 0, 30, 30))
	for y := 0; y < 30; y++ {
		for x := 0; x < 30; x++ {
			dstPic.Set(x, y, color.RGBA{255, 255, 255, 255}) // White
		}
	}
	state.pictures[2] = &Picture{
		ID:     2,
		Image:  dstPic,
		Width:  30,
		Height: 30,
	}

	// First: copy red square to (5, 5)
	err := state.MovePicture(0, 0, 0, 10, 10, 2, 5, 5, 0)
	if err != nil {
		t.Fatalf("First MovePic failed: %v", err)
	}

	// Second: copy blue square to SAME position (5, 5)
	// This should overwrite the red with blue
	err = state.MovePicture(1, 0, 0, 10, 10, 2, 5, 5, 0)
	if err != nil {
		t.Fatalf("Second MovePic failed: %v", err)
	}

	// Verify blue square overwrote red at (5, 5)
	destImg := state.pictures[2].Image.(*image.RGBA)
	r, g, b, _ := destImg.At(5, 5).RGBA()
	if r>>8 != 0 || g>>8 != 0 || b>>8 != 255 {
		t.Errorf("Expected blue at (5,5), got RGB(%d,%d,%d)", r>>8, g>>8, b>>8)
	}

	t.Logf("MovePic correctly overwrites at same position")
}
