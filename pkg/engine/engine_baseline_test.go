package engine

import (
	"image/color"
	"testing"
)

// TestBaselineEngineStateInitialization tests that EngineState initializes correctly
// Requirements: 4.1, 4.2, 5.1, 5.2, 14.1
func TestBaselineEngineStateInitialization(t *testing.T) {
	engine := NewTestEngine()

	// Verify all maps are initialized
	if engine.pictures == nil {
		t.Error("pictures map not initialized")
	}
	if engine.windows == nil {
		t.Error("windows map not initialized")
	}
	if engine.casts == nil {
		t.Error("casts map not initialized")
	}

	// Verify all maps are empty
	if len(engine.pictures) != 0 {
		t.Errorf("pictures should be empty on init, got %d", len(engine.pictures))
	}
	if len(engine.windows) != 0 {
		t.Errorf("windows should be empty on init, got %d", len(engine.windows))
	}
	if len(engine.casts) != 0 {
		t.Errorf("casts should be empty on init, got %d", len(engine.casts))
	}

	// Verify ID counters start at correct values (Requirement 4.2, 5.2)
	AssertIDCounters(t, engine, 0, 0, 1)

	// Verify state consistency
	AssertStateConsistency(t, engine)
}

// TestBaselineEngineStateReset tests that Reset clears all state
// Requirements: 4.1, 4.2, 5.1, 5.2, 14.1
func TestBaselineEngineStateReset(t *testing.T) {
	engine := NewTestEngineWithAssets(GetTestAssets())

	// Create some state
	picID := engine.LoadPic("test.bmp")
	if picID < 0 {
		t.Fatal("Failed to load picture")
	}

	winID := engine.OpenWin(picID, 100, 100, 640, 480, 0, 0, 0xFFFFFF)
	if winID < 0 {
		t.Fatal("Failed to open window")
	}

	// Verify state was created
	AssertResourceCount(t, engine, 1, 1, 0)

	// Reset the engine
	engine.Reset()

	// Verify all resources are cleared
	AssertResourceCount(t, engine, 0, 0, 0)

	// Verify ID counters are reset (Requirement 4.2, 5.2)
	AssertIDCounters(t, engine, 0, 0, 1)

	// Verify state consistency after reset
	AssertStateConsistency(t, engine)
}

// TestLoadPicBasic tests basic LoadPic functionality
// Requirements: 4.1, 4.2
func TestLoadPicBasic(t *testing.T) {
	engine := NewTestEngineWithAssets(GetTestAssets())

	// Load a picture
	picID := engine.LoadPic("test.bmp")

	// Verify picture was loaded with sequential ID (Requirement 4.2)
	if picID != 0 {
		t.Errorf("Expected first picture ID to be 0, got %d", picID)
	}

	// Verify picture exists
	pic := AssertPictureExists(t, engine, picID)

	// Verify picture has valid dimensions (Requirement 4.1)
	if pic.Width <= 0 || pic.Height <= 0 {
		t.Errorf("Picture dimensions invalid: %dx%d", pic.Width, pic.Height)
	}

	// Verify state consistency
	AssertStateConsistency(t, engine)
}

// TestLoadPicSequentialIDs tests that LoadPic assigns sequential IDs
// Requirements: 4.2
func TestLoadPicSequentialIDs(t *testing.T) {
	engine := NewTestEngineWithAssets(GetTestAssets())

	// Load multiple pictures
	pic1 := engine.LoadPic("test.bmp")
	pic2 := engine.LoadPic("small.bmp")
	pic3 := engine.LoadPic("medium.bmp")

	// Verify sequential IDs starting from 0 (Requirement 4.2)
	if pic1 != 0 {
		t.Errorf("Expected first picture ID to be 0, got %d", pic1)
	}
	if pic2 != 1 {
		t.Errorf("Expected second picture ID to be 1, got %d", pic2)
	}
	if pic3 != 2 {
		t.Errorf("Expected third picture ID to be 2, got %d", pic3)
	}

	// Verify all pictures exist
	AssertPictureExists(t, engine, pic1)
	AssertPictureExists(t, engine, pic2)
	AssertPictureExists(t, engine, pic3)

	// Verify resource count
	AssertResourceCount(t, engine, 3, 0, 0)

	// Verify state consistency
	AssertStateConsistency(t, engine)
}

// TestCreatePicBasic tests basic CreatePic functionality
// Requirements: 4.1, 4.2
func TestCreatePicBasic(t *testing.T) {
	engine := NewTestEngine()

	// Create a picture with specific dimensions
	width, height := 320, 240
	picID := engine.CreatePic(width, height)

	// Verify picture was created with sequential ID (Requirement 4.2)
	if picID != 0 {
		t.Errorf("Expected first picture ID to be 0, got %d", picID)
	}

	// Verify picture exists
	pic := AssertPictureExists(t, engine, picID)

	// Verify picture dimensions (Requirement 4.1)
	AssertPictureDimensions(t, pic, width, height)

	// Verify state consistency
	AssertStateConsistency(t, engine)
}

// TestCreatePicSequentialIDs tests that CreatePic assigns sequential IDs
// Requirements: 4.2
func TestCreatePicSequentialIDs(t *testing.T) {
	engine := NewTestEngine()

	// Create multiple pictures
	pic1 := engine.CreatePic(100, 100)
	pic2 := engine.CreatePic(200, 200)
	pic3 := engine.CreatePic(300, 300)

	// Verify sequential IDs starting from 0 (Requirement 4.2)
	if pic1 != 0 {
		t.Errorf("Expected first picture ID to be 0, got %d", pic1)
	}
	if pic2 != 1 {
		t.Errorf("Expected second picture ID to be 1, got %d", pic2)
	}
	if pic3 != 2 {
		t.Errorf("Expected third picture ID to be 2, got %d", pic3)
	}

	// Verify all pictures exist
	AssertPictureExists(t, engine, pic1)
	AssertPictureExists(t, engine, pic2)
	AssertPictureExists(t, engine, pic3)

	// Verify resource count
	AssertResourceCount(t, engine, 3, 0, 0)

	// Verify state consistency
	AssertStateConsistency(t, engine)
}

// TestOpenWinBasic tests basic OpenWin functionality
// Requirements: 14.1
func TestOpenWinBasic(t *testing.T) {
	engine := NewTestEngine()

	// Create a picture first
	picID := engine.CreatePic(640, 480)

	// Open a window
	x, y, w, h := 100, 100, 640, 480
	picX, picY := 0, 0
	bgColor := 0xFFFFFF
	winID := engine.OpenWin(picID, x, y, w, h, picX, picY, bgColor)

	// Verify window was created with sequential ID
	if winID != 0 {
		t.Errorf("Expected first window ID to be 0, got %d", winID)
	}

	// Verify window exists
	win := AssertWindowExists(t, engine, winID)

	// Verify window properties (Requirement 14.1)
	// Note: OpenWin adjusts position by BorderThickness (4) and TitleBarHeight (24)
	// So x=100 becomes 104, y=100 becomes 128 (100 + 24 + 4)
	expectedX := x + 4      // BorderThickness
	expectedY := y + 24 + 4 // TitleBarHeight + BorderThickness
	AssertWindowProperties(t, win, picID, expectedX, expectedY, w, h)

	// Verify state consistency
	AssertStateConsistency(t, engine)
}

// TestOpenWinSequentialIDs tests that OpenWin assigns sequential IDs
// Requirements: 14.1
func TestOpenWinSequentialIDs(t *testing.T) {
	engine := NewTestEngine()

	// Create pictures
	pic1 := engine.CreatePic(640, 480)
	pic2 := engine.CreatePic(320, 240)
	pic3 := engine.CreatePic(800, 600)

	// Open multiple windows
	win1 := engine.OpenWin(pic1, 0, 0, 640, 480, 0, 0, 0xFFFFFF)
	win2 := engine.OpenWin(pic2, 100, 100, 320, 240, 0, 0, 0x000000)
	win3 := engine.OpenWin(pic3, 200, 200, 800, 600, 0, 0, 0xFF0000)

	// Verify sequential IDs starting from 0
	if win1 != 0 {
		t.Errorf("Expected first window ID to be 0, got %d", win1)
	}
	if win2 != 1 {
		t.Errorf("Expected second window ID to be 1, got %d", win2)
	}
	if win3 != 2 {
		t.Errorf("Expected third window ID to be 2, got %d", win3)
	}

	// Verify all windows exist
	AssertWindowExists(t, engine, win1)
	AssertWindowExists(t, engine, win2)
	AssertWindowExists(t, engine, win3)

	// Verify resource count
	AssertResourceCount(t, engine, 3, 3, 0)

	// Verify state consistency
	AssertStateConsistency(t, engine)
}

// TestBaselineStateIsolationBetweenTests tests that state is isolated between test runs
// Requirements: 4.1, 4.2, 5.1, 5.2, 14.1
func TestBaselineStateIsolationBetweenTests(t *testing.T) {
	// Create first engine and populate it
	engine1 := NewTestEngineWithAssets(GetTestAssets())
	pic1 := engine1.LoadPic("test.bmp")
	win1 := engine1.OpenWin(pic1, 0, 0, 640, 480, 0, 0, 0xFFFFFF)

	// Verify first engine has resources
	AssertResourceCount(t, engine1, 1, 1, 0)
	AssertPictureExists(t, engine1, pic1)
	AssertWindowExists(t, engine1, win1)

	// Create second engine - should be completely independent
	engine2 := NewTestEngineWithAssets(GetTestAssets())

	// Verify second engine is empty
	AssertResourceCount(t, engine2, 0, 0, 0)

	// Verify ID counters are independent
	AssertIDCounters(t, engine2, 0, 0, 1)

	// Load resources in second engine
	pic2 := engine2.LoadPic("test.bmp")
	win2 := engine2.OpenWin(pic2, 100, 100, 320, 240, 0, 0, 0x000000)

	// Verify second engine has its own resources
	AssertResourceCount(t, engine2, 1, 1, 0)
	AssertPictureExists(t, engine2, pic2)
	AssertWindowExists(t, engine2, win2)

	// Verify first engine is unaffected
	AssertResourceCount(t, engine1, 1, 1, 0)
	AssertPictureExists(t, engine1, pic1)
	AssertWindowExists(t, engine1, win1)

	// Verify both engines have consistent state
	AssertStateConsistency(t, engine1)
	AssertStateConsistency(t, engine2)
}

// TestNoGlobalStateLeakage tests that there is no global state leakage
// Requirements: 4.1, 4.2, 5.1, 5.2, 14.1
func TestNoGlobalStateLeakage(t *testing.T) {
	// Create and populate first engine
	engine1 := NewTestEngine()
	pic1 := CreateTestPicture(engine1, 100, 100)
	win1 := engine1.OpenWin(pic1, 0, 0, 100, 100, 0, 0, 0xFFFFFF)

	// Modify some state
	engine1.globalWindowTitle = "Engine 1 Title"
	engine1.currentFontSize = 24
	engine1.currentTextColor = color.RGBA{255, 0, 0, 255}

	// Create second engine
	engine2 := NewTestEngine()

	// Verify second engine has default values, not values from engine1
	if engine2.globalWindowTitle == "Engine 1 Title" {
		t.Error("Global window title leaked from engine1 to engine2")
	}
	if engine2.currentFontSize == 24 {
		t.Error("Font size leaked from engine1 to engine2")
	}
	if engine2.currentTextColor == (color.RGBA{255, 0, 0, 255}) {
		t.Error("Text color leaked from engine1 to engine2")
	}

	// Verify second engine has no resources from engine1
	AssertResourceCount(t, engine2, 0, 0, 0)
	AssertPictureNotExists(t, engine2, pic1)
	AssertWindowNotExists(t, engine2, win1)

	// Verify both engines are independent
	AssertStateConsistency(t, engine1)
	AssertStateConsistency(t, engine2)
}

// TestResetClearsAllState tests that Reset clears all types of state
// Requirements: 4.1, 4.2, 5.1, 5.2, 14.1
func TestResetClearsAllState(t *testing.T) {
	engine := NewTestEngineWithAssets(GetTestAssets())

	// Create various resources
	pic1 := engine.LoadPic("test.bmp")
	pic2 := engine.CreatePic(200, 200)
	win1 := engine.OpenWin(pic1, 0, 0, 640, 480, 0, 0, 0xFFFFFF)
	win2 := engine.OpenWin(pic2, 100, 100, 200, 200, 0, 0, 0x000000)

	// Modify various state
	engine.globalWindowTitle = "Custom Title"
	engine.currentFontSize = 32
	engine.currentTextColor = color.RGBA{128, 128, 128, 255}
	engine.tickCount = 1000

	// Verify resources exist
	AssertResourceCount(t, engine, 2, 2, 0)

	// Reset
	engine.Reset()

	// Verify all resources are cleared
	AssertResourceCount(t, engine, 0, 0, 0)
	AssertPictureNotExists(t, engine, pic1)
	AssertPictureNotExists(t, engine, pic2)
	AssertWindowNotExists(t, engine, win1)
	AssertWindowNotExists(t, engine, win2)

	// Verify ID counters are reset
	AssertIDCounters(t, engine, 0, 0, 1)

	// Verify other state is reset to defaults
	if engine.globalWindowTitle != engine.defaultWindowTitle {
		t.Errorf("globalWindowTitle not reset, got %s", engine.globalWindowTitle)
	}
	if engine.currentFontSize != 14 {
		t.Errorf("currentFontSize not reset, got %d", engine.currentFontSize)
	}
	if engine.tickCount != 0 {
		t.Errorf("tickCount not reset, got %d", engine.tickCount)
	}

	// Verify state consistency
	AssertStateConsistency(t, engine)
}

// TestMultipleResets tests that Reset can be called multiple times
// Requirements: 4.1, 4.2, 5.1, 5.2, 14.1
func TestMultipleResets(t *testing.T) {
	engine := NewTestEngineWithAssets(GetTestAssets())

	// First cycle: create, reset
	pic1 := engine.LoadPic("test.bmp")
	AssertPictureExists(t, engine, pic1)
	engine.Reset()
	AssertResourceCount(t, engine, 0, 0, 0)

	// Second cycle: create, reset
	pic2 := engine.LoadPic("test.bmp")
	AssertPictureExists(t, engine, pic2)
	engine.Reset()
	AssertResourceCount(t, engine, 0, 0, 0)

	// Third cycle: create, reset
	pic3 := engine.LoadPic("test.bmp")
	AssertPictureExists(t, engine, pic3)
	engine.Reset()
	AssertResourceCount(t, engine, 0, 0, 0)

	// Verify IDs restart from 0 after each reset
	if pic1 != 0 || pic2 != 0 || pic3 != 0 {
		t.Errorf("Picture IDs should restart from 0 after reset, got %d, %d, %d", pic1, pic2, pic3)
	}

	// Verify final state is clean
	AssertStateConsistency(t, engine)
}

// TestMixedOperations tests a mix of LoadPic, CreatePic, and OpenWin
// Requirements: 4.1, 4.2, 5.1, 5.2, 14.1
func TestMixedOperations(t *testing.T) {
	engine := NewTestEngineWithAssets(GetTestAssets())

	// Mix of operations
	pic1 := engine.LoadPic("test.bmp")  // ID 0
	pic2 := engine.CreatePic(320, 240)  // ID 1
	pic3 := engine.LoadPic("small.bmp") // ID 2
	pic4 := engine.CreatePic(800, 600)  // ID 3

	// Verify sequential IDs
	if pic1 != 0 || pic2 != 1 || pic3 != 2 || pic4 != 3 {
		t.Errorf("Expected sequential IDs 0,1,2,3, got %d,%d,%d,%d", pic1, pic2, pic3, pic4)
	}

	// Open windows
	win1 := engine.OpenWin(pic1, 0, 0, 640, 480, 0, 0, 0xFFFFFF)     // ID 0
	win2 := engine.OpenWin(pic2, 100, 100, 320, 240, 0, 0, 0x000000) // ID 1
	win3 := engine.OpenWin(pic3, 200, 200, 100, 100, 0, 0, 0xFF0000) // ID 2

	// Verify sequential window IDs
	if win1 != 0 || win2 != 1 || win3 != 2 {
		t.Errorf("Expected sequential window IDs 0,1,2, got %d,%d,%d", win1, win2, win3)
	}

	// Verify resource counts
	AssertResourceCount(t, engine, 4, 3, 0)

	// Verify all resources exist
	AssertPictureExists(t, engine, pic1)
	AssertPictureExists(t, engine, pic2)
	AssertPictureExists(t, engine, pic3)
	AssertPictureExists(t, engine, pic4)
	AssertWindowExists(t, engine, win1)
	AssertWindowExists(t, engine, win2)
	AssertWindowExists(t, engine, win3)

	// Verify state consistency
	AssertStateConsistency(t, engine)
}

// TestStateConsistencyAfterOperations tests that state remains consistent
// Requirements: 4.1, 4.2, 5.1, 5.2, 14.1
func TestStateConsistencyAfterOperations(t *testing.T) {
	engine := NewTestEngineWithAssets(GetTestAssets())

	// Perform various operations
	for i := 0; i < 5; i++ {
		pic := engine.CreatePic(100*(i+1), 100*(i+1))
		win := engine.OpenWin(pic, i*50, i*50, 100*(i+1), 100*(i+1), 0, 0, 0xFFFFFF)

		// Verify state consistency after each operation
		AssertStateConsistency(t, engine)

		// Verify resources exist
		AssertPictureExists(t, engine, pic)
		AssertWindowExists(t, engine, win)
	}

	// Final consistency check
	AssertResourceCount(t, engine, 5, 5, 0)
	AssertStateConsistency(t, engine)
}
