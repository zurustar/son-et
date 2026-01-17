package engine

import (
	"testing"
)

// TestPutCastCreatesAndDraws tests that PutCast creates a cast and draws it to the destination picture
// Validates: Requirements 5.1, 5.2
func TestPutCastCreatesAndDraws(t *testing.T) {
	engine := NewTestEngine()

	// Create source and destination pictures
	srcPicID := CreateTestPicture(engine, 100, 100)
	dstPicID := CreateTestPicture(engine, 640, 480)

	// Verify initial state
	AssertResourceCount(t, engine, 2, 0, 0) // 2 pictures, 0 windows, 0 casts

	// Create a cast
	castID := engine.PutCast(srcPicID, dstPicID, 50, 60, 0, 0, 0, 0, 100, 100, 0, 0)

	// Verify cast was created
	if castID < 1 {
		t.Fatalf("PutCast failed, returned ID %d", castID)
	}

	// Verify cast exists
	cast := AssertCastExists(t, engine, castID)

	// Verify cast properties
	if cast.DestPicture != dstPicID {
		t.Errorf("Expected cast destination picture ID %d, got %d", dstPicID, cast.DestPicture)
	}
	if cast.X != 50 {
		t.Errorf("Expected cast X=50, got %d", cast.X)
	}
	if cast.Y != 60 {
		t.Errorf("Expected cast Y=60, got %d", cast.Y)
	}
	if cast.W != 100 {
		t.Errorf("Expected cast W=100, got %d", cast.W)
	}
	if cast.H != 100 {
		t.Errorf("Expected cast H=100, got %d", cast.H)
	}
	if !cast.Visible {
		t.Error("Expected cast to be visible")
	}

	// Verify cast was added to draw order
	if len(engine.castDrawOrder) != 1 {
		t.Errorf("Expected 1 cast in draw order, got %d", len(engine.castDrawOrder))
	}
	if engine.castDrawOrder[0] != castID {
		t.Errorf("Expected cast ID %d in draw order, got %d", castID, engine.castDrawOrder[0])
	}

	// Verify resource count
	// Note: PutCast creates an additional processed picture for transparency
	AssertResourceCount(t, engine, 3, 0, 1) // 3 pictures (2 original + 1 processed), 0 windows, 1 cast
}

// TestMoveCastUpdatesPosition tests that MoveCast updates the cast position and re-renders
// Validates: Requirements 5.3
func TestMoveCastUpdatesPosition(t *testing.T) {
	engine := NewTestEngine()

	// Create source and destination pictures
	srcPicID := CreateTestPicture(engine, 100, 100)
	dstPicID := CreateTestPicture(engine, 640, 480)

	// Create a cast at initial position
	castID := engine.PutCast(srcPicID, dstPicID, 50, 60, 0, 0, 0, 0, 100, 100, 0, 0)

	// Verify initial position
	cast := AssertCastExists(t, engine, castID)
	if cast.X != 50 || cast.Y != 60 {
		t.Errorf("Expected initial position (50, 60), got (%d, %d)", cast.X, cast.Y)
	}

	// Move the cast to a new position
	engine.MoveCast(castID, dstPicID, 150, 200, 100, 100, 0, 0)

	// Verify position was updated
	cast = AssertCastExists(t, engine, castID)
	if cast.X != 150 {
		t.Errorf("Expected cast X=150 after move, got %d", cast.X)
	}
	if cast.Y != 200 {
		t.Errorf("Expected cast Y=200 after move, got %d", cast.Y)
	}

	// Verify cast is still visible
	if !cast.Visible {
		t.Error("Expected cast to remain visible after move")
	}
}

// TestMoveCastWithDifferentSizes tests that MoveCast can update cast dimensions
// Validates: Requirements 5.3
func TestMoveCastWithDifferentSizes(t *testing.T) {
	engine := NewTestEngine()

	// Create source and destination pictures
	srcPicID := CreateTestPicture(engine, 200, 200)
	dstPicID := CreateTestPicture(engine, 640, 480)

	// Create a cast with initial size
	castID := engine.PutCast(srcPicID, dstPicID, 50, 60, 0, 0, 0, 0, 100, 100, 0, 0)

	// Verify initial size
	cast := AssertCastExists(t, engine, castID)
	if cast.W != 100 || cast.H != 100 {
		t.Errorf("Expected initial size (100, 100), got (%d, %d)", cast.W, cast.H)
	}

	// Move the cast with different size
	engine.MoveCast(castID, dstPicID, 150, 200, 0, 80, 60, 0, 0)

	// Verify size was updated
	cast = AssertCastExists(t, engine, castID)
	if cast.W != 80 {
		t.Errorf("Expected cast W=80 after move, got %d", cast.W)
	}
	if cast.H != 60 {
		t.Errorf("Expected cast H=60 after move, got %d", cast.H)
	}
}

// TestMoveCastWithSourceOffset tests that MoveCast can update source clipping offset
// Validates: Requirements 5.4
func TestMoveCastWithSourceOffset(t *testing.T) {
	engine := NewTestEngine()

	// Create a sprite sheet (source) and destination picture
	srcPicID := CreateTestPicture(engine, 200, 200)
	dstPicID := CreateTestPicture(engine, 640, 480)

	// Create a cast with initial source offset
	castID := engine.PutCast(srcPicID, dstPicID, 50, 60, 0, 0, 0, 0, 50, 50, 0, 0)

	// Verify initial source offset
	cast := AssertCastExists(t, engine, castID)
	if cast.SrcX != 0 || cast.SrcY != 0 {
		t.Errorf("Expected initial source offset (0, 0), got (%d, %d)", cast.SrcX, cast.SrcY)
	}

	// Move the cast with different source offset (sprite sheet animation)
	engine.MoveCast(castID, dstPicID, 50, 60, 0, 50, 50, 50, 50)

	// Verify source offset was updated
	cast = AssertCastExists(t, engine, castID)
	if cast.SrcX != 50 {
		t.Errorf("Expected cast SrcX=50 after move, got %d", cast.SrcX)
	}
	if cast.SrcY != 50 {
		t.Errorf("Expected cast SrcY=50 after move, got %d", cast.SrcY)
	}
}

// TestCastVisibilityWithValidDestination tests that casts are visible when destination picture is valid
// Validates: Requirements 5.1, 5.2
func TestCastVisibilityWithValidDestination(t *testing.T) {
	engine := NewTestEngine()

	// Create source and destination pictures
	srcPicID := CreateTestPicture(engine, 100, 100)
	dstPicID := CreateTestPicture(engine, 640, 480)

	// Create a cast
	castID := engine.PutCast(srcPicID, dstPicID, 50, 60, 0, 0, 0, 0, 100, 100, 0, 0)

	// Verify cast is visible
	cast := AssertCastExists(t, engine, castID)
	if !cast.Visible {
		t.Error("Expected cast to be visible with valid destination picture")
	}

	// Verify destination picture exists
	AssertPictureExists(t, engine, dstPicID)
}

// TestCastRenderingOrder tests that casts are rendered in creation order (z-ordering)
// Validates: Requirements 5.5
func TestCastRenderingOrder(t *testing.T) {
	engine := NewTestEngine()

	// Create source and destination pictures
	srcPicID := CreateTestPicture(engine, 100, 100)
	dstPicID := CreateTestPicture(engine, 640, 480)

	// Create multiple casts in specific order
	cast1 := engine.PutCast(srcPicID, dstPicID, 100, 100, 0, 0, 0, 0, 100, 100, 0, 0)
	cast2 := engine.PutCast(srcPicID, dstPicID, 110, 110, 0, 0, 0, 0, 100, 100, 0, 0)
	cast3 := engine.PutCast(srcPicID, dstPicID, 120, 120, 0, 0, 0, 0, 100, 100, 0, 0)

	// Verify all casts were created
	AssertCastExists(t, engine, cast1)
	AssertCastExists(t, engine, cast2)
	AssertCastExists(t, engine, cast3)

	// Verify draw order matches creation order
	if len(engine.castDrawOrder) != 3 {
		t.Fatalf("Expected 3 casts in draw order, got %d", len(engine.castDrawOrder))
	}

	// First created should be first in draw order (bottom layer)
	if engine.castDrawOrder[0] != cast1 {
		t.Errorf("Expected first cast in draw order to be %d, got %d", cast1, engine.castDrawOrder[0])
	}
	if engine.castDrawOrder[1] != cast2 {
		t.Errorf("Expected second cast in draw order to be %d, got %d", cast2, engine.castDrawOrder[1])
	}
	if engine.castDrawOrder[2] != cast3 {
		t.Errorf("Expected third cast in draw order to be %d, got %d", cast3, engine.castDrawOrder[2])
	}
}

// TestMultipleCastsOnSameDestination tests that multiple casts can be drawn to the same destination
// Validates: Requirements 5.1, 5.5
func TestMultipleCastsOnSameDestination(t *testing.T) {
	engine := NewTestEngine()

	// Create source and destination pictures
	srcPicID := CreateTestPicture(engine, 100, 100)
	dstPicID := CreateTestPicture(engine, 640, 480)

	// Create multiple casts on the same destination
	cast1 := engine.PutCast(srcPicID, dstPicID, 50, 50, 0, 0, 0, 0, 100, 100, 0, 0)
	cast2 := engine.PutCast(srcPicID, dstPicID, 150, 150, 0, 0, 0, 0, 100, 100, 0, 0)
	cast3 := engine.PutCast(srcPicID, dstPicID, 250, 250, 0, 0, 0, 0, 100, 100, 0, 0)

	// Verify all casts exist
	AssertCastExists(t, engine, cast1)
	AssertCastExists(t, engine, cast2)
	AssertCastExists(t, engine, cast3)

	// Verify all casts reference the same destination
	c1 := engine.casts[cast1]
	c2 := engine.casts[cast2]
	c3 := engine.casts[cast3]

	if c1.DestPicture != dstPicID || c2.DestPicture != dstPicID || c3.DestPicture != dstPicID {
		t.Error("Not all casts reference the same destination picture")
	}

	// Verify resource count
	// Note: Each PutCast creates a processed picture for transparency
	AssertResourceCount(t, engine, 5, 0, 3) // 2 original + 3 processed, 0 windows, 3 casts
}

// TestMoveCastRerendersAllCasts tests that MoveCast re-renders all casts on the destination
// Validates: Requirements 5.3, 5.5
func TestMoveCastRerendersAllCasts(t *testing.T) {
	engine := NewTestEngine()

	// Create source and destination pictures
	srcPicID := CreateTestPicture(engine, 100, 100)
	dstPicID := CreateTestPicture(engine, 640, 480)

	// Create multiple casts
	cast1 := engine.PutCast(srcPicID, dstPicID, 50, 50, 0, 0, 0, 0, 100, 100, 0, 0)
	cast2 := engine.PutCast(srcPicID, dstPicID, 150, 150, 0, 0, 0, 0, 100, 100, 0, 0)

	// Move one cast - this should trigger re-rendering of all casts on the destination
	engine.MoveCast(cast1, dstPicID, 100, 100, 100, 100, 0, 0)

	// Verify both casts still exist and are visible
	c1 := AssertCastExists(t, engine, cast1)
	c2 := AssertCastExists(t, engine, cast2)

	if !c1.Visible || !c2.Visible {
		t.Error("Expected all casts to remain visible after MoveCast")
	}

	// Verify cast1 position was updated
	if c1.X != 100 || c1.Y != 100 {
		t.Errorf("Expected cast1 position (100, 100), got (%d, %d)", c1.X, c1.Y)
	}

	// Verify cast2 position remained unchanged
	if c2.X != 150 || c2.Y != 150 {
		t.Errorf("Expected cast2 position (150, 150), got (%d, %d)", c2.X, c2.Y)
	}
}

// TestCastAppearsOnDestinationAfterPutCast tests that a cast is immediately visible on the destination picture
// Validates: Requirements 5.1, 5.2
func TestCastAppearsOnDestinationAfterPutCast(t *testing.T) {
	engine := NewTestEngine()

	// Create source and destination pictures
	srcPicID := CreateTestPicture(engine, 100, 100)
	dstPicID := CreateTestPicture(engine, 640, 480)

	// Verify destination picture exists before creating cast
	dstPic := AssertPictureExists(t, engine, dstPicID)
	if dstPic.Image == nil {
		t.Fatal("Destination picture image is nil")
	}

	// Create a cast
	castID := engine.PutCast(srcPicID, dstPicID, 50, 60, 0, 0, 0, 0, 100, 100, 0, 0)

	// Verify cast was created
	cast := AssertCastExists(t, engine, castID)

	// Verify cast references the destination picture
	if cast.DestPicture != dstPicID {
		t.Errorf("Expected cast to reference destination picture %d, got %d", dstPicID, cast.DestPicture)
	}

	// Verify cast is visible
	if !cast.Visible {
		t.Error("Expected cast to be visible immediately after PutCast")
	}

	// Verify destination picture still exists and is valid
	dstPic = AssertPictureExists(t, engine, dstPicID)
	if dstPic.Image == nil {
		t.Error("Destination picture image became nil after PutCast")
	}
}

// TestCastWithClipping tests that casts can be created with source clipping parameters
// Validates: Requirements 5.4
func TestCastWithClipping(t *testing.T) {
	engine := NewTestEngine()

	// Create a sprite sheet (200x200) and destination picture
	srcPicID := CreateTestPicture(engine, 200, 200)
	dstPicID := CreateTestPicture(engine, 640, 480)

	// Create a cast with clipping (extract 50x50 region from position 25,25)
	castID := engine.PutCast(srcPicID, dstPicID, 100, 100, 0, 0, 0, 0, 50, 50, 25, 25)

	// Verify cast was created with correct clipping parameters
	cast := AssertCastExists(t, engine, castID)

	if cast.W != 50 {
		t.Errorf("Expected cast width 50, got %d", cast.W)
	}
	if cast.H != 50 {
		t.Errorf("Expected cast height 50, got %d", cast.H)
	}
	if cast.SrcX != 25 {
		t.Errorf("Expected cast SrcX 25, got %d", cast.SrcX)
	}
	if cast.SrcY != 25 {
		t.Errorf("Expected cast SrcY 25, got %d", cast.SrcY)
	}
}

// TestCastZOrderingWithOverlap tests that overlapping casts render in correct z-order
// Validates: Requirements 5.5
func TestCastZOrderingWithOverlap(t *testing.T) {
	engine := NewTestEngine()

	// Create source and destination pictures
	srcPicID := CreateTestPicture(engine, 100, 100)
	dstPicID := CreateTestPicture(engine, 640, 480)

	// Create overlapping casts (each offset by 10 pixels)
	cast1 := engine.PutCast(srcPicID, dstPicID, 100, 100, 0, 0, 0, 0, 100, 100, 0, 0)
	cast2 := engine.PutCast(srcPicID, dstPicID, 110, 110, 0, 0, 0, 0, 100, 100, 0, 0)
	cast3 := engine.PutCast(srcPicID, dstPicID, 120, 120, 0, 0, 0, 0, 100, 100, 0, 0)
	cast4 := engine.PutCast(srcPicID, dstPicID, 130, 130, 0, 0, 0, 0, 100, 100, 0, 0)

	// Verify draw order is maintained
	if len(engine.castDrawOrder) != 4 {
		t.Fatalf("Expected 4 casts in draw order, got %d", len(engine.castDrawOrder))
	}

	// Verify order: first created = bottom layer, last created = top layer
	expectedOrder := []int{cast1, cast2, cast3, cast4}
	for i, expectedID := range expectedOrder {
		if engine.castDrawOrder[i] != expectedID {
			t.Errorf("Expected cast %d at position %d in draw order, got %d", expectedID, i, engine.castDrawOrder[i])
		}
	}
}

// TestMoveCastPreservesZOrder tests that moving a cast doesn't change its z-order
// Validates: Requirements 5.3, 5.5
func TestMoveCastPreservesZOrder(t *testing.T) {
	engine := NewTestEngine()

	// Create source and destination pictures
	srcPicID := CreateTestPicture(engine, 100, 100)
	dstPicID := CreateTestPicture(engine, 640, 480)

	// Create multiple casts
	_ = engine.PutCast(srcPicID, dstPicID, 100, 100, 0, 0, 0, 0, 100, 100, 0, 0)
	cast2 := engine.PutCast(srcPicID, dstPicID, 150, 150, 0, 0, 0, 0, 100, 100, 0, 0)
	_ = engine.PutCast(srcPicID, dstPicID, 200, 200, 0, 0, 0, 0, 100, 100, 0, 0)

	// Record initial draw order
	initialOrder := make([]int, len(engine.castDrawOrder))
	copy(initialOrder, engine.castDrawOrder)

	// Move the middle cast
	engine.MoveCast(cast2, dstPicID, 300, 300, 100, 100, 0, 0)

	// Verify draw order hasn't changed
	if len(engine.castDrawOrder) != len(initialOrder) {
		t.Fatalf("Draw order length changed after MoveCast")
	}

	for i := range initialOrder {
		if engine.castDrawOrder[i] != initialOrder[i] {
			t.Errorf("Draw order changed at position %d: expected %d, got %d", i, initialOrder[i], engine.castDrawOrder[i])
		}
	}

	// Verify cast2 position was updated
	c2 := AssertCastExists(t, engine, cast2)
	if c2.X != 300 || c2.Y != 300 {
		t.Errorf("Expected cast2 position (300, 300), got (%d, %d)", c2.X, c2.Y)
	}
}
