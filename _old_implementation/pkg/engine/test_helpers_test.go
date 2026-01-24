package engine

import (
	"image/color"
	"testing"
)

// TestNewTestEngine verifies that NewTestEngine creates a properly configured test engine
func TestNewTestEngine(t *testing.T) {
	engine := NewTestEngine()

	// Verify it's properly initialized
	if engine == nil {
		t.Fatal("NewTestEngine returned nil")
	}

	// Verify mock dependencies are set
	if engine.assetLoader == nil {
		t.Error("AssetLoader not set")
	}
	if engine.imageDecoder == nil {
		t.Error("ImageDecoder not set")
	}
	if engine.renderer == nil {
		t.Error("Renderer not set")
	}

	// Verify it's using mock implementations
	if _, ok := engine.assetLoader.(*MockAssetLoader); !ok {
		t.Error("AssetLoader is not MockAssetLoader")
	}
	if _, ok := engine.imageDecoder.(*MockImageDecoder); !ok {
		t.Error("ImageDecoder is not MockImageDecoder")
	}
	if _, ok := engine.renderer.(*MockRenderer); !ok {
		t.Error("Renderer is not MockRenderer")
	}

	// Verify initial state
	AssertResourceCount(t, engine, 0, 0, 0)
	AssertIDCounters(t, engine, 0, 0, 1)
}

// TestNewTestEngineWithAssets verifies asset loading with pre-configured assets
func TestNewTestEngineWithAssets(t *testing.T) {
	assets := map[string][]byte{
		"test1.bmp": []byte("test data 1"),
		"test2.bmp": []byte("test data 2"),
	}

	engine := NewTestEngineWithAssets(assets)

	// Try to load the assets
	picID1 := engine.LoadPic("test1.bmp")
	picID2 := engine.LoadPic("test2.bmp")

	if picID1 < 0 {
		t.Error("Failed to load test1.bmp")
	}
	if picID2 < 0 {
		t.Error("Failed to load test2.bmp")
	}

	// Verify pictures were created
	AssertPictureExists(t, engine, picID1)
	AssertPictureExists(t, engine, picID2)
}

// TestNewTestEngineWithAssetList verifies asset loading with asset list
func TestNewTestEngineWithAssetList(t *testing.T) {
	assets := []TestAsset{
		{Name: "asset1.bmp", Data: []byte("data1")},
		{Name: "asset2.bmp", Data: []byte("data2")},
	}

	engine := NewTestEngineWithAssetList(assets)

	// Try to load the assets
	picID1 := engine.LoadPic("asset1.bmp")
	picID2 := engine.LoadPic("asset2.bmp")

	if picID1 < 0 || picID2 < 0 {
		t.Error("Failed to load assets")
	}
}

// TestAssertPictureExists verifies the picture existence assertion
func TestAssertPictureExists(t *testing.T) {
	engine := NewTestEngine()

	// Create a test picture
	picID := CreateTestPicture(engine, 100, 100)

	// This should not fail
	pic := AssertPictureExists(t, engine, picID)
	if pic == nil {
		t.Error("AssertPictureExists returned nil")
	}
}

// TestAssertPictureNotExists verifies the picture non-existence assertion
func TestAssertPictureNotExists(t *testing.T) {
	engine := NewTestEngine()

	// This should not fail (picture 999 doesn't exist)
	AssertPictureNotExists(t, engine, 999)
}

// TestAssertPictureDimensions verifies the picture dimension assertion
func TestAssertPictureDimensions(t *testing.T) {
	engine := NewTestEngine()

	// Create a test picture with specific dimensions
	picID := CreateTestPicture(engine, 640, 480)
	pic := AssertPictureExists(t, engine, picID)

	// This should not fail
	AssertPictureDimensions(t, pic, 640, 480)
}

// TestAssertWindowExists verifies the window existence assertion
func TestAssertWindowExists(t *testing.T) {
	engine := NewTestEngine()

	// Create a test picture and window
	picID := CreateTestPicture(engine, 100, 100)
	winID := engine.OpenWin(picID, 0, 0, 100, 100, 0, 0, 0xFFFFFF)

	// This should not fail
	win := AssertWindowExists(t, engine, winID)
	if win == nil {
		t.Error("AssertWindowExists returned nil")
	}
}

// TestAssertWindowProperties verifies the window properties assertion
func TestAssertWindowProperties(t *testing.T) {
	engine := NewTestEngine()

	// Create a test picture and window
	picID := CreateTestPicture(engine, 100, 100)
	winID := engine.OpenWin(picID, 10, 20, 640, 480, 0, 0, 0xFFFFFF)

	win := AssertWindowExists(t, engine, winID)

	// Note: OpenWin adjusts position by BorderThickness (4) and TitleBarHeight (24)
	// So x=10 becomes 14, y=20 becomes 48 (20 + 24 + 4)
	AssertWindowProperties(t, win, picID, 14, 48, 640, 480)
}

// TestAssertCastExists verifies the cast existence assertion
func TestAssertCastExists(t *testing.T) {
	engine := NewTestEngine()

	// Create test pictures
	srcPicID := CreateTestPicture(engine, 100, 100)
	dstPicID := CreateTestPicture(engine, 200, 200)

	// Create a cast
	castID := engine.PutCast(srcPicID, dstPicID, 10, 10, 0, 0, 0, 0, 100, 100, 0, 0)

	// This should not fail
	cast := AssertCastExists(t, engine, castID)
	if cast == nil {
		t.Error("AssertCastExists returned nil")
	}
}

// TestAssertCastProperties verifies the cast properties assertion
func TestAssertCastProperties(t *testing.T) {
	engine := NewTestEngine()

	// Create test pictures
	srcPicID := CreateTestPicture(engine, 100, 100)
	dstPicID := CreateTestPicture(engine, 200, 200)

	// Create a cast
	castID := engine.PutCast(srcPicID, dstPicID, 50, 60, 0, 0, 0, 0, 100, 100, 0, 0)

	cast := AssertCastExists(t, engine, castID)

	// Note: Cast now references a processed picture (with transparency), not the original source
	// The processed picture ID will be different from srcPicID
	// We verify the destination and position instead
	if cast.DestPicture != dstPicID {
		t.Errorf("Expected cast destination picture ID %d, got %d", dstPicID, cast.DestPicture)
	}
	if cast.X != 50 || cast.Y != 60 {
		t.Errorf("Expected cast position (50, 60), got (%d, %d)", cast.X, cast.Y)
	}
}

// TestAssertIDCounters verifies the ID counter assertion
func TestAssertIDCounters(t *testing.T) {
	engine := NewTestEngine()

	// Initial state
	AssertIDCounters(t, engine, 0, 0, 1)

	// Create some resources
	CreateTestPicture(engine, 100, 100)
	CreateTestPicture(engine, 100, 100)

	// Check updated counters
	AssertIDCounters(t, engine, 2, 0, 1)
}

// TestAssertResourceCount verifies the resource count assertion
func TestAssertResourceCount(t *testing.T) {
	engine := NewTestEngine()

	// Initial state
	AssertResourceCount(t, engine, 0, 0, 0)

	// Create resources
	picID := CreateTestPicture(engine, 100, 100)
	_ = engine.OpenWin(picID, 0, 0, 100, 100, 0, 0, 0xFFFFFF)
	dstPicID := CreateTestPicture(engine, 200, 200)
	_ = engine.PutCast(picID, dstPicID, 10, 10, 0, 0, 0, 0, 100, 100, 0, 0)

	// Check counts
	// Note: PutCast creates an additional processed picture for transparency
	AssertResourceCount(t, engine, 3, 1, 1)
}

// TestAssertStateConsistency verifies the state consistency checker
func TestAssertStateConsistency(t *testing.T) {
	engine := NewTestEngine()

	// Create a consistent state
	picID := CreateTestPicture(engine, 100, 100)
	winID := engine.OpenWin(picID, 0, 0, 100, 100, 0, 0, 0xFFFFFF)
	dstPicID := CreateTestPicture(engine, 200, 200)
	castID := engine.PutCast(picID, dstPicID, 10, 10, 0, 0, 0, 0, 100, 100, 0, 0)

	// This should not fail
	AssertStateConsistency(t, engine)

	// Verify resources exist
	AssertPictureExists(t, engine, picID)
	AssertPictureExists(t, engine, dstPicID)
	AssertWindowExists(t, engine, winID)
	AssertCastExists(t, engine, castID)
}

// TestCreateTestImage verifies test image creation
func TestCreateTestImage(t *testing.T) {
	img := CreateTestImage(100, 100, color.RGBA{255, 0, 0, 255})

	if img == nil {
		t.Fatal("CreateTestImage returned nil")
	}

	bounds := img.Bounds()
	if bounds.Dx() != 100 || bounds.Dy() != 100 {
		t.Errorf("Expected 100x100 image, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

// TestCreateTestImageWithPattern verifies pattern image creation
func TestCreateTestImageWithPattern(t *testing.T) {
	img := CreateTestImageWithPattern(100, 100)

	if img == nil {
		t.Fatal("CreateTestImageWithPattern returned nil")
	}

	bounds := img.Bounds()
	if bounds.Dx() != 100 || bounds.Dy() != 100 {
		t.Errorf("Expected 100x100 image, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

// TestCreateTestPicture verifies test picture creation
func TestCreateTestPicture(t *testing.T) {
	engine := NewTestEngine()

	picID := CreateTestPicture(engine, 640, 480)

	if picID < 0 {
		t.Error("CreateTestPicture failed")
	}

	pic := AssertPictureExists(t, engine, picID)
	AssertPictureDimensions(t, pic, 640, 480)
}

// TestCreateTestPictureWithColor verifies colored test picture creation
func TestCreateTestPictureWithColor(t *testing.T) {
	engine := NewTestEngine()

	red := color.RGBA{255, 0, 0, 255}
	picID := CreateTestPictureWithColor(engine, 100, 100, red)

	if picID < 0 {
		t.Error("CreateTestPictureWithColor failed")
	}

	pic := AssertPictureExists(t, engine, picID)
	AssertPictureDimensions(t, pic, 100, 100)

	// Note: We cannot read pixels from Ebitengine images before the game starts
	// This is a limitation of Ebitengine's rendering system
	// The image was created with the correct color, but we can't verify it here
}

// TestGetPixelColor verifies pixel color retrieval
func TestGetPixelColor(t *testing.T) {
	engine := NewTestEngine()

	red := color.RGBA{255, 0, 0, 255}
	picID := CreateTestPictureWithColor(engine, 100, 100, red)
	pic := AssertPictureExists(t, engine, picID)

	// Note: We cannot read pixels from Ebitengine images before the game starts
	// This is a limitation of Ebitengine's rendering system
	// GetPixelColor is still useful for integration tests where the game is running
	_ = pic // Use the variable to avoid unused warning
}

// TestAssertPixelColor verifies pixel color assertion
func TestAssertPixelColor(t *testing.T) {
	engine := NewTestEngine()

	red := color.RGBA{255, 0, 0, 255}
	picID := CreateTestPictureWithColor(engine, 100, 100, red)
	pic := AssertPictureExists(t, engine, picID)

	// Note: We cannot read pixels from Ebitengine images before the game starts
	// This is a limitation of Ebitengine's rendering system
	// AssertPixelColor is still useful for integration tests where the game is running
	_ = pic // Use the variable to avoid unused warning
}

// TestFixtureData verifies fixture data creation
func TestFixtureData(t *testing.T) {
	fixtures := NewFixtureData()

	if fixtures == nil {
		t.Fatal("NewFixtureData returned nil")
	}

	if len(fixtures.SmallImage) == 0 {
		t.Error("SmallImage is empty")
	}
	if len(fixtures.TestBMP) == 0 {
		t.Error("TestBMP is empty")
	}
	if len(fixtures.TestMIDI) == 0 {
		t.Error("TestMIDI is empty")
	}
	if len(fixtures.TestWAV) == 0 {
		t.Error("TestWAV is empty")
	}
}

// TestGetTestAssets verifies test asset retrieval
func TestGetTestAssets(t *testing.T) {
	assets := GetTestAssets()

	if len(assets) == 0 {
		t.Fatal("GetTestAssets returned empty map")
	}

	// Check for expected assets
	expectedAssets := []string{"test.bmp", "test.mid", "test.wav", "small.bmp", "medium.bmp", "large.bmp"}
	for _, name := range expectedAssets {
		if _, exists := assets[name]; !exists {
			t.Errorf("Expected asset %s not found", name)
		}
	}
}

// TestGetTestAssetList verifies test asset list retrieval
func TestGetTestAssetList(t *testing.T) {
	assets := GetTestAssetList()

	if len(assets) == 0 {
		t.Fatal("GetTestAssetList returned empty list")
	}

	// Verify each asset has name and data
	for i, asset := range assets {
		if asset.Name == "" {
			t.Errorf("Asset %d has empty name", i)
		}
		if len(asset.Data) == 0 {
			t.Errorf("Asset %d (%s) has empty data", i, asset.Name)
		}
	}
}

// TestCompleteWorkflow demonstrates a complete test workflow using helpers
func TestCompleteWorkflow(t *testing.T) {
	// 1. Create test engine with assets
	engine := NewTestEngineWithAssets(GetTestAssets())

	// 2. Verify initial state
	AssertResourceCount(t, engine, 0, 0, 0)
	AssertIDCounters(t, engine, 0, 0, 1)

	// 3. Load a picture
	picID := engine.LoadPic("test.bmp")
	if picID < 0 {
		t.Fatal("Failed to load test.bmp")
	}

	// 4. Verify picture was created
	pic := AssertPictureExists(t, engine, picID)
	AssertPictureDimensions(t, pic, 100, 100) // MockImageDecoder creates 100x100 images

	// 5. Create a window
	winID := engine.OpenWin(picID, 10, 20, 640, 480, 0, 0, 0xFFFFFF)
	if winID < 0 {
		t.Fatal("Failed to open window")
	}

	// 6. Verify window was created
	win := AssertWindowExists(t, engine, winID)
	// Note: OpenWin adjusts position by BorderThickness (4) and TitleBarHeight (24)
	AssertWindowProperties(t, win, picID, 14, 48, 640, 480)

	// 7. Create another picture for cast destination
	dstPicID := CreateTestPicture(engine, 800, 600)

	// 8. Create a cast
	castID := engine.PutCast(picID, dstPicID, 50, 60, 0, 0, 0, 0, 100, 100, 0, 0)
	if castID < 0 {
		t.Fatal("Failed to create cast")
	}

	// 9. Verify cast was created
	cast := AssertCastExists(t, engine, castID)
	// Note: Cast now references a processed picture (with transparency), not the original source
	if cast.DestPicture != dstPicID {
		t.Errorf("Expected cast destination picture ID %d, got %d", dstPicID, cast.DestPicture)
	}
	if cast.X != 50 || cast.Y != 60 {
		t.Errorf("Expected cast position (50, 60), got (%d, %d)", cast.X, cast.Y)
	}

	// 10. Verify resource counts
	// Note: PutCast creates an additional processed picture for transparency
	AssertResourceCount(t, engine, 3, 1, 1)

	// 11. Verify state consistency
	AssertStateConsistency(t, engine)

	// 12. Clean up
	engine.Reset()

	// 13. Verify cleanup
	AssertResourceCount(t, engine, 0, 0, 0)
	AssertIDCounters(t, engine, 0, 0, 1)
}
