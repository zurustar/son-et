package engine

import (
	"image"
	"testing"
)

func TestPutCast(t *testing.T) {
	state := NewEngineState(nil, nil, nil)

	// Create a cast
	castID := state.PutCast(1, 2, 100, 200, 0, 0, 64, 64, -1)

	if castID != 1 {
		t.Errorf("Expected cast ID 1, got %d", castID)
	}

	// Verify cast was created
	cast := state.GetCast(castID)
	if cast == nil {
		t.Fatal("Cast not found")
	}

	if cast.WindowID != 1 {
		t.Errorf("Expected WindowID 1, got %d", cast.WindowID)
	}
	if cast.PictureID != 2 {
		t.Errorf("Expected PictureID 2, got %d", cast.PictureID)
	}
	if cast.X != 100 {
		t.Errorf("Expected X 100, got %d", cast.X)
	}
	if cast.Y != 200 {
		t.Errorf("Expected Y 200, got %d", cast.Y)
	}
	if cast.SrcX != 0 {
		t.Errorf("Expected SrcX 0, got %d", cast.SrcX)
	}
	if cast.SrcY != 0 {
		t.Errorf("Expected SrcY 0, got %d", cast.SrcY)
	}
	if cast.Width != 64 {
		t.Errorf("Expected Width 64, got %d", cast.Width)
	}
	if cast.Height != 64 {
		t.Errorf("Expected Height 64, got %d", cast.Height)
	}
	if !cast.Visible {
		t.Error("Expected cast to be visible")
	}
}

func TestMoveCastPosition(t *testing.T) {
	state := NewEngineState(nil, nil, nil)

	// Create a destination picture (required for MoveCast double buffering)
	destPic := &Picture{
		ID:     1,
		Image:  image.NewRGBA(image.Rect(0, 0, 200, 200)),
		Width:  200,
		Height: 200,
	}
	state.pictures[1] = destPic

	// Create a cast
	castID := state.PutCast(1, 2, 100, 200, 0, 0, 64, 64, -1)

	// Move cast position only (clipping unchanged)
	err := state.MoveCast(castID, 150, 250, -1, -1, -1, -1)
	if err != nil {
		t.Fatalf("MoveCast failed: %v", err)
	}

	// Verify position was updated
	cast := state.GetCast(castID)
	if cast.X != 150 {
		t.Errorf("Expected X 150, got %d", cast.X)
	}
	if cast.Y != 250 {
		t.Errorf("Expected Y 250, got %d", cast.Y)
	}

	// Verify clipping unchanged
	if cast.SrcX != 0 {
		t.Errorf("Expected SrcX 0, got %d", cast.SrcX)
	}
	if cast.SrcY != 0 {
		t.Errorf("Expected SrcY 0, got %d", cast.SrcY)
	}
	if cast.Width != 64 {
		t.Errorf("Expected Width 64, got %d", cast.Width)
	}
	if cast.Height != 64 {
		t.Errorf("Expected Height 64, got %d", cast.Height)
	}
}

func TestMoveCastPositionAndClipping(t *testing.T) {
	state := NewEngineState(nil, nil, nil)

	// Create a destination picture (required for MoveCast double buffering)
	destPic := &Picture{
		ID:     1,
		Image:  image.NewRGBA(image.Rect(0, 0, 200, 200)),
		Width:  200,
		Height: 200,
	}
	state.pictures[1] = destPic

	// Create a cast
	castID := state.PutCast(1, 2, 100, 200, 0, 0, 64, 64, -1)

	// Move cast position and clipping
	err := state.MoveCast(castID, 150, 250, 32, 32, 128, 128)
	if err != nil {
		t.Fatalf("MoveCast failed: %v", err)
	}

	// Verify position was updated
	cast := state.GetCast(castID)
	if cast.X != 150 {
		t.Errorf("Expected X 150, got %d", cast.X)
	}
	if cast.Y != 250 {
		t.Errorf("Expected Y 250, got %d", cast.Y)
	}

	// Verify clipping was updated
	if cast.SrcX != 32 {
		t.Errorf("Expected SrcX 32, got %d", cast.SrcX)
	}
	if cast.SrcY != 32 {
		t.Errorf("Expected SrcY 32, got %d", cast.SrcY)
	}
	if cast.Width != 128 {
		t.Errorf("Expected Width 128, got %d", cast.Width)
	}
	if cast.Height != 128 {
		t.Errorf("Expected Height 128, got %d", cast.Height)
	}
}

func TestDeleteCast(t *testing.T) {
	state := NewEngineState(nil, nil, nil)

	// Create a cast
	castID := state.PutCast(1, 2, 100, 200, 0, 0, 64, 64, -1)

	// Delete the cast
	state.DeleteCast(castID)

	// Verify cast was deleted
	cast := state.GetCast(castID)
	if cast != nil {
		t.Error("Expected cast to be deleted")
	}
}

func TestGetCastsOrder(t *testing.T) {
	state := NewEngineState(nil, nil, nil)

	// Create casts in specific order
	id1 := state.PutCast(1, 2, 100, 100, 0, 0, 64, 64, -1)
	id2 := state.PutCast(1, 3, 200, 200, 0, 0, 64, 64, -1)
	id3 := state.PutCast(1, 4, 300, 300, 0, 0, 64, 64, -1)

	// Get casts
	casts := state.GetCasts()

	// Verify order (should be in creation order for z-ordering)
	if len(casts) != 3 {
		t.Fatalf("Expected 3 casts, got %d", len(casts))
	}

	if casts[0].ID != id1 {
		t.Errorf("Expected first cast ID %d, got %d", id1, casts[0].ID)
	}
	if casts[1].ID != id2 {
		t.Errorf("Expected second cast ID %d, got %d", id2, casts[1].ID)
	}
	if casts[2].ID != id3 {
		t.Errorf("Expected third cast ID %d, got %d", id3, casts[2].ID)
	}
}

func TestGetCastsByWindow(t *testing.T) {
	state := NewEngineState(nil, nil, nil)

	// Create casts in different windows
	win1cast1 := state.PutCast(1, 2, 100, 100, 0, 0, 64, 64, -1)
	win2cast1 := state.PutCast(2, 3, 200, 200, 0, 0, 64, 64, -1)
	win1cast2 := state.PutCast(1, 4, 300, 300, 0, 0, 64, 64, -1)

	// Get casts for window 1
	casts := state.GetCastsByWindow(1)

	// Verify only window 1 casts returned
	if len(casts) != 2 {
		t.Fatalf("Expected 2 casts for window 1, got %d", len(casts))
	}

	// Verify order (creation order)
	if casts[0].ID != win1cast1 {
		t.Errorf("Expected first cast ID %d, got %d", win1cast1, casts[0].ID)
	}
	if casts[1].ID != win1cast2 {
		t.Errorf("Expected second cast ID %d, got %d", win1cast2, casts[1].ID)
	}

	// Get casts for window 2
	casts = state.GetCastsByWindow(2)

	// Verify only window 2 casts returned
	if len(casts) != 1 {
		t.Fatalf("Expected 1 cast for window 2, got %d", len(casts))
	}

	if casts[0].ID != win2cast1 {
		t.Errorf("Expected cast ID %d, got %d", win2cast1, casts[0].ID)
	}
}

func TestMoveCastNonExistent(t *testing.T) {
	state := NewEngineState(nil, nil, nil)

	// Try to move non-existent cast
	err := state.MoveCast(999, 100, 200, -1, -1, -1, -1)
	if err == nil {
		t.Error("Expected error when moving non-existent cast")
	}
}

func TestMultipleCastsZOrder(t *testing.T) {
	state := NewEngineState(nil, nil, nil)

	// Create multiple casts
	id1 := state.PutCast(1, 2, 0, 0, 0, 0, 64, 64, -1)
	_ = state.PutCast(1, 3, 10, 10, 0, 0, 64, 64, -1)
	id3 := state.PutCast(1, 4, 20, 20, 0, 0, 64, 64, -1)

	// Get all casts
	casts := state.GetCasts()

	// Verify z-order (creation order = rendering order)
	if len(casts) != 3 {
		t.Fatalf("Expected 3 casts, got %d", len(casts))
	}

	// First created should be rendered first (bottom)
	if casts[0].ID != id1 {
		t.Errorf("Expected bottom cast ID %d, got %d", id1, casts[0].ID)
	}

	// Last created should be rendered last (top)
	if casts[2].ID != id3 {
		t.Errorf("Expected top cast ID %d, got %d", id3, casts[2].ID)
	}
}
