package engine

import (
	"image"
	"image/color"
	"testing"
)

// TestCastAnimation tests that MoveCast correctly updates srcY parameter
// to show different frames from a vertically-stacked sprite sheet.
// This reproduces the y_saru animation pattern where srcY alternates
// between 0 and 89 to show different animation frames.
func TestCastAnimation(t *testing.T) {
	state := NewEngineState(nil, nil, nil)
	state.SetDebugLevel(2) // Enable debug logging

	// Create source picture with two frames stacked vertically
	// Frame 1 (y=0-88): Red
	// Frame 2 (y=89-177): Blue
	srcPic := image.NewRGBA(image.Rect(0, 0, 77, 178))

	// Frame 1: Red (y=0 to 88)
	for y := 0; y < 89; y++ {
		for x := 0; x < 77; x++ {
			srcPic.Set(x, y, color.RGBA{255, 0, 0, 255}) // Red
		}
	}

	// Frame 2: Blue (y=89 to 177)
	for y := 89; y < 178; y++ {
		for x := 0; x < 77; x++ {
			srcPic.Set(x, y, color.RGBA{0, 0, 255, 255}) // Blue
		}
	}

	state.pictures[17] = &Picture{
		ID:     17,
		Image:  srcPic,
		Width:  77,
		Height: 178,
	}

	// Create destination picture (base_pic)
	dstPic := image.NewRGBA(image.Rect(0, 0, 300, 300))
	for y := 0; y < 300; y++ {
		for x := 0; x < 300; x++ {
			dstPic.Set(x, y, color.RGBA{255, 255, 255, 255}) // White background
		}
	}
	state.pictures[25] = &Picture{
		ID:     25,
		Image:  dstPic,
		Width:  300,
		Height: 300,
	}

	t.Log("=== Initial PutCast with srcY=0 (Frame 1: Red) ===")
	// PutCast(17, base_pic, 113, 0, 0xffffff, 0, 0, 0, 89, 77, 0, 0)
	// This creates a cast showing frame 1 (srcY=0, height=77, width=89)
	castID := state.PutCast(25, 17, 113, 0, 0, 0, 89, 77, 0xffffff)

	cast := state.GetCast(castID)
	t.Logf("Cast created: ID=%d, X=%d, Y=%d, SrcX=%d, SrcY=%d, Width=%d, Height=%d",
		cast.ID, cast.X, cast.Y, cast.SrcX, cast.SrcY, cast.Width, cast.Height)

	// Verify frame 1 (red) is displayed
	destImg := state.pictures[25].Image.(*image.RGBA)
	r, g, b, _ := destImg.At(113, 0).RGBA()
	t.Logf("Pixel at (113,0) after PutCast: RGB(%d,%d,%d)", r>>8, g>>8, b>>8)
	if r>>8 != 255 || g>>8 != 0 || b>>8 != 0 {
		t.Errorf("Expected red (frame 1) at (113,0), got RGB(%d,%d,%d)", r>>8, g>>8, b>>8)
	}

	t.Log("\n=== MoveCast with srcY=89 (Frame 2: Blue) ===")
	// MoveCast(cast, 17, 105, 10, 0, 89, 77, 89, 0)
	// This should switch to frame 2 (srcY=89) and move to position (105, 10)
	err := state.MoveCast(castID, 105, 10, 0, 89, 89, 77)
	if err != nil {
		t.Fatalf("MoveCast failed: %v", err)
	}

	cast = state.GetCast(castID)
	t.Logf("Cast after MoveCast: X=%d, Y=%d, SrcX=%d, SrcY=%d, Width=%d, Height=%d",
		cast.X, cast.Y, cast.SrcX, cast.SrcY, cast.Width, cast.Height)

	// Verify srcY was updated to 89
	if cast.SrcY != 89 {
		t.Errorf("Expected SrcY=89 after MoveCast, got SrcY=%d", cast.SrcY)
	}

	// Verify frame 2 (blue) is now displayed at new position
	destImg = state.pictures[25].Image.(*image.RGBA)
	r, g, b, _ = destImg.At(105, 10).RGBA()
	t.Logf("Pixel at (105,10) after MoveCast: RGB(%d,%d,%d)", r>>8, g>>8, b>>8)
	if r>>8 != 0 || g>>8 != 0 || b>>8 != 255 {
		t.Errorf("Expected blue (frame 2) at (105,10), got RGB(%d,%d,%d)", r>>8, g>>8, b>>8)
		t.Errorf("This indicates srcY parameter is not being used correctly in animation")
	}

	t.Log("\n=== MoveCast back to srcY=0 (Frame 1: Red) ===")
	// MoveCast(cast, 17, 101, 20, 0, 89, 77, 0, 0)
	// This should switch back to frame 1 (srcY=0) and move to position (101, 20)
	err = state.MoveCast(castID, 101, 20, 0, 0, 89, 77)
	if err != nil {
		t.Fatalf("MoveCast failed: %v", err)
	}

	cast = state.GetCast(castID)
	t.Logf("Cast after second MoveCast: X=%d, Y=%d, SrcX=%d, SrcY=%d, Width=%d, Height=%d",
		cast.X, cast.Y, cast.SrcX, cast.SrcY, cast.Width, cast.Height)

	// Verify srcY was updated back to 0
	if cast.SrcY != 0 {
		t.Errorf("Expected SrcY=0 after second MoveCast, got SrcY=%d", cast.SrcY)
	}

	// Verify frame 1 (red) is displayed again at new position
	destImg = state.pictures[25].Image.(*image.RGBA)
	r, g, b, _ = destImg.At(101, 20).RGBA()
	t.Logf("Pixel at (101,20) after second MoveCast: RGB(%d,%d,%d)", r>>8, g>>8, b>>8)
	if r>>8 != 255 || g>>8 != 0 || b>>8 != 0 {
		t.Errorf("Expected red (frame 1) at (101,20), got RGB(%d,%d,%d)", r>>8, g>>8, b>>8)
		t.Errorf("This indicates animation is not alternating between frames correctly")
	}

	// Verify old position (105, 10) is cleared (no accumulation)
	r, g, b, a := destImg.At(105, 10).RGBA()
	t.Logf("Pixel at old position (105,10): RGBA(%d,%d,%d,%d)", r>>8, g>>8, b>>8, a>>8)
	if a>>8 != 0 && (r>>8 == 0 && g>>8 == 0 && b>>8 == 255) {
		t.Errorf("Old position still shows blue - cast accumulation detected")
	}
}
