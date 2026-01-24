package engine

import (
	"image"
	"image/color"
	"testing"
)

// TestMoveCastDebug is a debug version of TestMoveCastDoubleBuffering
// with detailed logging to understand what's happening.
func TestMoveCastDebug(t *testing.T) {
	state := NewEngineState(nil, nil, nil)
	state.SetDebugLevel(2) // Enable debug logging

	// Create source picture (10x10 red square)
	srcPic := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			srcPic.Set(x, y, color.RGBA{255, 0, 0, 255}) // Red
		}
	}
	state.pictures[1] = &Picture{
		ID:     1,
		Image:  srcPic,
		Width:  10,
		Height: 10,
	}

	// Create destination picture (50x50 white background)
	dstPic := image.NewRGBA(image.Rect(0, 0, 50, 50))
	for y := 0; y < 50; y++ {
		for x := 0; x < 50; x++ {
			dstPic.Set(x, y, color.RGBA{255, 255, 255, 255}) // White
		}
	}
	state.pictures[2] = &Picture{
		ID:     2,
		Image:  dstPic,
		Width:  50,
		Height: 50,
	}

	t.Log("=== Initial state ===")
	t.Logf("Source picture 1: %dx%d", state.pictures[1].Width, state.pictures[1].Height)
	t.Logf("Dest picture 2: %dx%d", state.pictures[2].Width, state.pictures[2].Height)

	// PutCast at position (10, 10)
	t.Log("\n=== PutCast at (10, 10) ===")
	castID := state.PutCast(2, 1, 10, 10, 0, 0, 10, 10, -1)
	t.Logf("Created cast ID: %d", castID)

	// Check cast properties
	cast := state.GetCast(castID)
	t.Logf("Cast: WindowID=%d, PictureID=%d, X=%d, Y=%d, SrcX=%d, SrcY=%d, Width=%d, Height=%d",
		cast.WindowID, cast.PictureID, cast.X, cast.Y, cast.SrcX, cast.SrcY, cast.Width, cast.Height)

	// Verify red square at (10, 10)
	destImg := state.pictures[2].Image.(*image.RGBA)
	r, g, b, _ := destImg.At(10, 10).RGBA()
	t.Logf("Pixel at (10,10) after PutCast: RGB(%d,%d,%d)", r>>8, g>>8, b>>8)
	if r>>8 != 255 || g>>8 != 0 || b>>8 != 0 {
		t.Errorf("Expected red at (10,10), got RGB(%d,%d,%d)", r>>8, g>>8, b>>8)
	}

	// Check BackBuffer
	if state.pictures[2].BackBuffer != nil {
		t.Logf("BackBuffer exists: %T", state.pictures[2].BackBuffer)
	} else {
		t.Log("BackBuffer is nil")
	}

	// MoveCast to position (20, 20)
	t.Log("\n=== MoveCast to (20, 20) ===")
	err := state.MoveCast(castID, 20, 20, -1, -1, -1, -1)
	if err != nil {
		t.Fatalf("MoveCast failed: %v", err)
	}

	// Check cast properties after move
	cast = state.GetCast(castID)
	t.Logf("Cast after move: X=%d, Y=%d", cast.X, cast.Y)

	// Check all casts
	allCasts := state.GetCasts()
	t.Logf("Total casts: %d", len(allCasts))
	for i, c := range allCasts {
		t.Logf("  Cast[%d]: ID=%d, WindowID=%d, X=%d, Y=%d, Visible=%v",
			i, c.ID, c.WindowID, c.X, c.Y, c.Visible)
	}

	// Check BackBuffer after MoveCast
	// NOTE: After MoveCast, Image and BackBuffer are swapped!
	// So the "new" Image is what was the BackBuffer, and vice versa
	if state.pictures[2].BackBuffer != nil {
		t.Logf("BackBuffer exists after MoveCast: %T", state.pictures[2].BackBuffer)
		// BackBuffer now contains the OLD image (before the move)
		backBuf := state.pictures[2].BackBuffer.(*image.RGBA)
		r, g, b, _ := backBuf.At(20, 20).RGBA()
		t.Logf("BackBuffer pixel at (20,20): RGB(%d,%d,%d)", r>>8, g>>8, b>>8)
		// BackBuffer should have the old content (white or red at old position)
	}

	// Verify red square is now at (20, 20) in the MAIN IMAGE (which was swapped from BackBuffer)
	// IMPORTANT: Get the image again after MoveCast because it was swapped!
	destImg = state.pictures[2].Image.(*image.RGBA)
	r, g, b, _ = destImg.At(20, 20).RGBA()
	t.Logf("Main Image pixel at (20,20) after MoveCast: RGB(%d,%d,%d)", r>>8, g>>8, b>>8)
	if r>>8 != 255 || g>>8 != 0 || b>>8 != 0 {
		t.Errorf("Expected red at new position (20,20), got RGB(%d,%d,%d)", r>>8, g>>8, b>>8)
	}

	// Check old position (10, 10) - should be transparent (cleared by double buffering)
	r, g, b, a := destImg.At(10, 10).RGBA()
	t.Logf("Main Image pixel at (10,10) after MoveCast: RGBA(%d,%d,%d,%d)", r>>8, g>>8, b>>8, a>>8)
	// After MoveCast with double buffering, old position should be transparent (alpha=0)
	// or the background color if there was one
	if a>>8 != 0 && (r>>8 == 255 && g>>8 == 0 && b>>8 == 0) {
		t.Errorf("Expected transparent or non-red at old position (10,10) after MoveCast, got RGBA(%d,%d,%d,%d)", r>>8, g>>8, b>>8, a>>8)
		t.Errorf("This indicates cast accumulation - double buffering is not working correctly")
	} else {
		t.Logf("✓ Old position correctly cleared (transparent or non-red)")
	}
}
