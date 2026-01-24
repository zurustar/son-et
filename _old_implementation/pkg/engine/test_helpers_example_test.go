package engine

import (
	"image/color"
	"testing"
)

// Example_basicTestSetup demonstrates basic test setup with NewTestEngine
func Example_basicTestSetup() {
	// Create a test engine with mock dependencies
	engine := NewTestEngine()

	// Create test resources
	picID := CreateTestPicture(engine, 640, 480)

	// Use the picture
	_ = engine.OpenWin(picID, 0, 0, 640, 480, 0, 0, 0xFFFFFF)

	// Clean up
	engine.Reset()
}

// Example_testWithAssets demonstrates testing with pre-loaded assets
func Example_testWithAssets() {
	// Create engine with test assets
	engine := NewTestEngineWithAssets(GetTestAssets())

	// Load an asset
	picID := engine.LoadPic("test.bmp")

	// Use the loaded picture
	_ = picID
}

// Example_assertionHelpers demonstrates using assertion helpers
func ExampleAssertPictureExists() {
	t := &testing.T{} // In real tests, this comes from the test function
	engine := NewTestEngine()

	// Create a picture
	picID := CreateTestPicture(engine, 100, 100)

	// Assert it exists and get it
	pic := AssertPictureExists(t, engine, picID)

	// Use the picture
	_ = pic
}

// Example_stateConsistency demonstrates state consistency checking
func ExampleAssertStateConsistency() {
	t := &testing.T{} // In real tests, this comes from the test function
	engine := NewTestEngine()

	// Create some resources
	picID := CreateTestPicture(engine, 100, 100)
	_ = engine.OpenWin(picID, 0, 0, 100, 100, 0, 0, 0xFFFFFF)

	// Verify state is consistent
	AssertStateConsistency(t, engine)
}

// Example_completeTest demonstrates a complete test workflow
func Example_completeTest() {
	t := &testing.T{} // In real tests, this comes from the test function

	// 1. Setup: Create test engine with assets
	engine := NewTestEngineWithAssets(GetTestAssets())

	// 2. Verify initial state
	AssertResourceCount(t, engine, 0, 0, 0)

	// 3. Load a picture
	picID := engine.LoadPic("test.bmp")
	pic := AssertPictureExists(t, engine, picID)
	AssertPictureDimensions(t, pic, 100, 100)

	// 4. Create a window
	winID := engine.OpenWin(picID, 10, 20, 640, 480, 0, 0, 0xFFFFFF)
	win := AssertWindowExists(t, engine, winID)
	AssertWindowProperties(t, win, picID, 14, 48, 640, 480) // Note: adjusted for borders

	// 5. Create a cast
	dstPicID := CreateTestPicture(engine, 800, 600)
	castID := engine.PutCast(picID, dstPicID, 50, 60, 0, 0, 0, 0, 100, 100, 0, 0)
	cast := AssertCastExists(t, engine, castID)
	AssertCastProperties(t, cast, picID, dstPicID, 50, 60)

	// 6. Verify final state
	AssertResourceCount(t, engine, 2, 1, 1)
	AssertStateConsistency(t, engine)

	// 7. Cleanup
	engine.Reset()
	AssertResourceCount(t, engine, 0, 0, 0)
}

// Example_customTestData demonstrates creating custom test data
func Example_customTestData() {
	engine := NewTestEngine()

	// Create a red picture
	red := color.RGBA{255, 0, 0, 255}
	redPicID := CreateTestPictureWithColor(engine, 100, 100, red)

	// Create a blue picture
	blue := color.RGBA{0, 0, 255, 255}
	bluePicID := CreateTestPictureWithColor(engine, 100, 100, blue)

	// Use the pictures
	_, _ = redPicID, bluePicID
}

// Example_fixtureData demonstrates using fixture data
func Example_fixtureData() {
	// Get pre-defined test assets
	assets := GetTestAssets()

	// Create engine with these assets
	engine := NewTestEngineWithAssets(assets)

	// Load various asset types
	_ = engine.LoadPic("test.bmp")
	_ = engine.LoadPic("small.bmp")
	_ = engine.LoadPic("medium.bmp")
	_ = engine.LoadPic("large.bmp")
}

// Example_testIsolation demonstrates test isolation
func Example_testIsolation() {
	// Each test gets its own engine
	engine1 := NewTestEngine()
	engine2 := NewTestEngine()

	// Modifications to engine1 don't affect engine2
	CreateTestPicture(engine1, 100, 100)
	CreateTestPicture(engine1, 200, 200)

	// engine2 is still empty
	_ = len(engine2.pictures) // 0

	// Each engine can be reset independently
	engine1.Reset()
	engine2.Reset()
}
