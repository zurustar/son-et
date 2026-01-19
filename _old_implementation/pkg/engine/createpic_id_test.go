package engine

import (
	"testing"
)

// TestCreatePicWithSourceID tests CreatePic(sourcePicID) ID assignment
// This test mimics the y_saru scenario where many pictures are loaded
// and then CreatePic is called with a source picture ID
func TestCreatePicWithSourceID(t *testing.T) {
	engine := NewTestEngine()

	// Simulate loading 26 pictures (IDs 0-25)
	for i := 0; i < 26; i++ {
		picID := engine.CreatePic(100, 100)
		if picID != i {
			t.Errorf("Expected picture ID %d, got %d", i, picID)
		}
	}

	// Now call CreatePic(25) - should create a new picture with ID 26
	// This mimics: base_pic = CreatePic(25)
	newPicID := engine.CreatePic(25)

	// The new picture should have ID 26, not 0
	if newPicID != 26 {
		t.Errorf("Expected CreatePic(25) to return ID 26, got %d", newPicID)
	}

	// Verify the picture exists
	pic := AssertPictureExists(t, engine, newPicID)

	// Verify it has the same dimensions as picture 25
	srcPic := engine.pictures[25]
	if pic.Width != srcPic.Width || pic.Height != srcPic.Height {
		t.Errorf("Expected dimensions %dx%d, got %dx%d",
			srcPic.Width, srcPic.Height, pic.Width, pic.Height)
	}

	// Verify we now have 27 pictures total
	AssertResourceCount(t, engine, 27, 0, 0)
}

// TestCreatePicAfterLoadPic tests that CreatePic increments nextPicID correctly
// after LoadPic operations
func TestCreatePicAfterLoadPic(t *testing.T) {
	engine := NewTestEngineWithAssets(map[string][]byte{
		"test.bmp": []byte{}, // Empty data is fine for mock decoder
	})

	// Load a picture - should get ID 0
	pic1 := engine.LoadPic("test.bmp")
	if pic1 != 0 {
		t.Errorf("Expected LoadPic to return ID 0, got %d", pic1)
	}

	// Create a picture - should get ID 1
	pic2 := engine.CreatePic(200, 200)
	if pic2 != 1 {
		t.Errorf("Expected CreatePic to return ID 1, got %d", pic2)
	}

	// Load another picture - should get ID 2
	pic3 := engine.LoadPic("test.bmp")
	if pic3 != 2 {
		t.Errorf("Expected LoadPic to return ID 2, got %d", pic3)
	}

	// Create with source ID - should get ID 3
	pic4 := engine.CreatePic(0)
	if pic4 != 3 {
		t.Errorf("Expected CreatePic(0) to return ID 3, got %d", pic4)
	}

	// Verify all pictures exist
	AssertPictureExists(t, engine, 0)
	AssertPictureExists(t, engine, 1)
	AssertPictureExists(t, engine, 2)
	AssertPictureExists(t, engine, 3)

	// Verify resource count
	AssertResourceCount(t, engine, 4, 0, 0)
}

// TestCreatePicIDAssignment tests all CreatePic ID assignment behaviors
// This test explicitly validates Requirements 4.2 and 4.3:
// - CreatePic(width, height) returns sequential IDs
// - CreatePic(sourcePicID) returns new sequential ID (not source ID)
// - LoadPic and CreatePic share the same ID counter
// - nextPicID increments correctly across mixed operations
func TestCreatePicIDAssignment(t *testing.T) {
	engine := NewTestEngineWithAssets(map[string][]byte{
		"test.bmp": []byte{},
	})

	// Test 1: CreatePic(width, height) returns sequential IDs starting from 0
	pic0 := engine.CreatePic(100, 100)
	if pic0 != 0 {
		t.Errorf("Expected first CreatePic(width, height) to return ID 0, got %d", pic0)
	}

	pic1 := engine.CreatePic(200, 200)
	if pic1 != 1 {
		t.Errorf("Expected second CreatePic(width, height) to return ID 1, got %d", pic1)
	}

	pic2 := engine.CreatePic(150, 150)
	if pic2 != 2 {
		t.Errorf("Expected third CreatePic(width, height) to return ID 2, got %d", pic2)
	}

	// Test 2: CreatePic(sourcePicID) returns new sequential ID, NOT the source ID
	pic3 := engine.CreatePic(0) // Copy from picture 0
	if pic3 != 3 {
		t.Errorf("Expected CreatePic(0) to return new ID 3, got %d (should NOT return source ID 0)", pic3)
	}

	pic4 := engine.CreatePic(1) // Copy from picture 1
	if pic4 != 4 {
		t.Errorf("Expected CreatePic(1) to return new ID 4, got %d (should NOT return source ID 1)", pic4)
	}

	// Test 3: LoadPic and CreatePic share the same ID counter
	pic5 := engine.LoadPic("test.bmp")
	if pic5 != 5 {
		t.Errorf("Expected LoadPic to return ID 5 (continuing from CreatePic counter), got %d", pic5)
	}

	pic6 := engine.CreatePic(300, 300)
	if pic6 != 6 {
		t.Errorf("Expected CreatePic to return ID 6 (continuing from LoadPic counter), got %d", pic6)
	}

	// Test 4: Verify nextPicID increments correctly across mixed operations
	// Pattern: CreatePic -> LoadPic -> CreatePic(source) -> LoadPic -> CreatePic
	pic7 := engine.CreatePic(50, 50)
	pic8 := engine.LoadPic("test.bmp")
	pic9 := engine.CreatePic(2) // Copy from picture 2
	pic10 := engine.LoadPic("test.bmp")
	pic11 := engine.CreatePic(75, 75)

	expectedIDs := []int{7, 8, 9, 10, 11}
	actualIDs := []int{pic7, pic8, pic9, pic10, pic11}

	for i, expected := range expectedIDs {
		if actualIDs[i] != expected {
			t.Errorf("Mixed operation %d: expected ID %d, got %d", i, expected, actualIDs[i])
		}
	}

	// Verify all pictures exist
	for i := 0; i <= 11; i++ {
		AssertPictureExists(t, engine, i)
	}

	// Verify total count
	AssertResourceCount(t, engine, 12, 0, 0)

	// Verify that CreatePic(sourcePicID) creates pictures with correct dimensions
	srcPic0 := engine.pictures[0]
	copyPic3 := engine.pictures[3]
	if copyPic3.Width != srcPic0.Width || copyPic3.Height != srcPic0.Height {
		t.Errorf("CreatePic(0) should copy dimensions: expected %dx%d, got %dx%d",
			srcPic0.Width, srcPic0.Height, copyPic3.Width, copyPic3.Height)
	}

	srcPic1 := engine.pictures[1]
	copyPic4 := engine.pictures[4]
	if copyPic4.Width != srcPic1.Width || copyPic4.Height != srcPic1.Height {
		t.Errorf("CreatePic(1) should copy dimensions: expected %dx%d, got %dx%d",
			srcPic1.Width, srcPic1.Height, copyPic4.Width, copyPic4.Height)
	}
}
