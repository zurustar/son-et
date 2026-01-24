package engine

import (
	"image"
	"image/color"
	"testing"
)

// TestCastTransparency tests that PutCast correctly handles transparent colors.
// This reproduces the issue where C1 appears white and C2 shows incorrect images.
func TestCastTransparency(t *testing.T) {
	state := NewEngineState(nil, nil, nil)

	// Create source picture with white background and red content
	// White (0xFFFFFF) should be transparent
	srcPic := image.NewRGBA(image.Rect(0, 0, 20, 20))
	for y := 0; y < 20; y++ {
		for x := 0; x < 20; x++ {
			if x >= 5 && x < 15 && y >= 5 && y < 15 {
				// Red content in center
				srcPic.Set(x, y, color.RGBA{255, 0, 0, 255})
			} else {
				// White background (should be transparent)
				srcPic.Set(x, y, color.RGBA{255, 255, 255, 255})
			}
		}
	}
	state.pictures[1] = &Picture{
		ID:     1,
		Image:  srcPic,
		Width:  20,
		Height: 20,
	}

	// Create destination picture with blue background
	dstPic := image.NewRGBA(image.Rect(0, 0, 50, 50))
	for y := 0; y < 50; y++ {
		for x := 0; x < 50; x++ {
			dstPic.Set(x, y, color.RGBA{0, 0, 255, 255}) // Blue
		}
	}
	state.pictures[2] = &Picture{
		ID:     2,
		Image:  dstPic,
		Width:  50,
		Height: 50,
	}

	// PutCast with white (0xFFFFFF) as transparent color
	castID := state.PutCast(2, 1, 10, 10, 0, 0, 20, 20, 0xFFFFFF)

	// Verify cast was created
	cast := state.GetCast(castID)
	if cast == nil {
		t.Fatal("Cast not found")
	}

	// Check destination picture
	destImg := state.pictures[2].Image.(*image.RGBA)

	// Center of cast (15, 15) should be red (from source)
	r, g, b, _ := destImg.At(15, 15).RGBA()
	if r>>8 != 255 || g>>8 != 0 || b>>8 != 0 {
		t.Errorf("Expected red at cast center (15,15), got RGB(%d,%d,%d)", r>>8, g>>8, b>>8)
	}

	// Corner of cast (10, 10) should be blue (transparent white from source, showing background)
	r, g, b, _ = destImg.At(10, 10).RGBA()
	if r>>8 != 0 || g>>8 != 0 || b>>8 != 255 {
		t.Errorf("Expected blue at cast corner (10,10) due to transparency, got RGB(%d,%d,%d)", r>>8, g>>8, b>>8)
	}

	// Outside cast (5, 5) should be blue (original background)
	r, g, b, _ = destImg.At(5, 5).RGBA()
	if r>>8 != 0 || g>>8 != 0 || b>>8 != 255 {
		t.Errorf("Expected blue outside cast (5,5), got RGB(%d,%d,%d)", r>>8, g>>8, b>>8)
	}
}

// TestMoveCastDoubleBuffering tests that MoveCast correctly uses double buffering
// to prevent cast accumulation.
func TestMoveCastDoubleBuffering(t *testing.T) {
	state := NewEngineState(nil, nil, nil)

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

	// PutCast at position (10, 10)
	castID := state.PutCast(2, 1, 10, 10, 0, 0, 10, 10, -1)

	// Verify red square at (10, 10)
	destImg := state.pictures[2].Image.(*image.RGBA)
	r, g, b, _ := destImg.At(10, 10).RGBA()
	if r>>8 != 255 || g>>8 != 0 || b>>8 != 0 {
		t.Errorf("Expected red at (10,10), got RGB(%d,%d,%d)", r>>8, g>>8, b>>8)
	}

	// MoveCast to position (20, 20)
	err := state.MoveCast(castID, 20, 20, -1, -1, -1, -1)
	if err != nil {
		t.Fatalf("MoveCast failed: %v", err)
	}

	// Verify red square is now at (20, 20)
	// IMPORTANT: Get the image again after MoveCast because it was swapped!
	destImg = state.pictures[2].Image.(*image.RGBA)
	r, g, b, _ = destImg.At(20, 20).RGBA()
	if r>>8 != 255 || g>>8 != 0 || b>>8 != 0 {
		t.Errorf("Expected red at new position (20,20), got RGB(%d,%d,%d)", r>>8, g>>8, b>>8)
	}

	// CRITICAL: Old position (10, 10) should be TRANSPARENT (cleared by double buffering)
	// NOT red (which would indicate accumulation)
	r, g, b, a := destImg.At(10, 10).RGBA()
	// Check that it's either transparent (alpha=0) or not red
	if a>>8 != 0 && r>>8 == 255 && g>>8 == 0 && b>>8 == 0 {
		t.Errorf("Expected transparent or non-red at old position (10,10) after MoveCast, got RGBA(%d,%d,%d,%d)", r>>8, g>>8, b>>8, a>>8)
		t.Errorf("This indicates cast accumulation - double buffering is not working correctly")
	}
}

// TestMultipleCastsDoubleBuffering tests that MoveCast correctly redraws
// ALL casts when one cast moves.
func TestMultipleCastsDoubleBuffering(t *testing.T) {
	state := NewEngineState(nil, nil, nil)

	// Create red source picture
	redPic := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			redPic.Set(x, y, color.RGBA{255, 0, 0, 255}) // Red
		}
	}
	state.pictures[1] = &Picture{
		ID:     1,
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
	state.pictures[2] = &Picture{
		ID:     2,
		Image:  bluePic,
		Width:  10,
		Height: 10,
	}

	// Create destination picture (white background)
	dstPic := image.NewRGBA(image.Rect(0, 0, 50, 50))
	for y := 0; y < 50; y++ {
		for x := 0; x < 50; x++ {
			dstPic.Set(x, y, color.RGBA{255, 255, 255, 255}) // White
		}
	}
	state.pictures[3] = &Picture{
		ID:     3,
		Image:  dstPic,
		Width:  50,
		Height: 50,
	}

	// Create two casts
	cast1 := state.PutCast(3, 1, 10, 10, 0, 0, 10, 10, -1) // Red at (10, 10)
	cast2 := state.PutCast(3, 2, 25, 25, 0, 0, 10, 10, -1) // Blue at (25, 25)

	// Verify both casts are visible
	destImg := state.pictures[3].Image.(*image.RGBA)
	r, g, b, _ := destImg.At(10, 10).RGBA()
	if r>>8 != 255 || g>>8 != 0 || b>>8 != 0 {
		t.Errorf("Expected red cast1 at (10,10), got RGB(%d,%d,%d)", r>>8, g>>8, b>>8)
	}
	r, g, b, _ = destImg.At(25, 25).RGBA()
	if r>>8 != 0 || g>>8 != 0 || b>>8 != 255 {
		t.Errorf("Expected blue cast2 at (25,25), got RGB(%d,%d,%d)", r>>8, g>>8, b>>8)
	}

	// Move cast1 to (15, 15)
	err := state.MoveCast(cast1, 15, 15, -1, -1, -1, -1)
	if err != nil {
		t.Fatalf("MoveCast failed: %v", err)
	}

	// IMPORTANT: Get the image again after MoveCast because it was swapped!
	destImg = state.pictures[3].Image.(*image.RGBA)

	// Verify cast1 is at new position
	r, g, b, _ = destImg.At(15, 15).RGBA()
	if r>>8 != 255 || g>>8 != 0 || b>>8 != 0 {
		t.Errorf("Expected red cast1 at new position (15,15), got RGB(%d,%d,%d)", r>>8, g>>8, b>>8)
	}

	// CRITICAL: cast2 should STILL be visible at (25, 25)
	// Double buffering should redraw ALL casts, not just the moved one
	r, g, b, _ = destImg.At(25, 25).RGBA()
	if r>>8 != 0 || g>>8 != 0 || b>>8 != 255 {
		t.Errorf("Expected blue cast2 to remain at (25,25), got RGB(%d,%d,%d)", r>>8, g>>8, b>>8)
		t.Errorf("This indicates MoveCast is not redrawing all casts correctly")
	}

	// Old position (10, 10) should be transparent (cleared)
	r, g, b, a := destImg.At(10, 10).RGBA()
	// Check that it's either transparent (alpha=0) or not red
	if a>>8 != 0 && r>>8 == 255 && g>>8 == 0 && b>>8 == 0 {
		t.Errorf("Expected transparent or non-red at old position (10,10), got RGBA(%d,%d,%d,%d)", r>>8, g>>8, b>>8, a>>8)
	}

	// Verify cast2 is still registered
	c2 := state.GetCast(cast2)
	if c2 == nil {
		t.Error("Cast2 should still exist")
	}
}
