package engine

import (
	"image"
	"image/color"
	"io/fs"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// NewTestEngine creates a new EngineState configured for testing
// It uses mock dependencies to avoid external dependencies and enable headless testing
func NewTestEngine() *EngineState {
	mockLoader := NewMockAssetLoader()
	mockDecoder := NewMockImageDecoder(100, 100)
	mockRenderer := NewMockRenderer()

	return NewEngineState(
		WithAssetLoader(mockLoader),
		WithImageDecoder(mockDecoder),
		WithRenderer(mockRenderer),
	)
}

// NewTestEngineWithAssets creates a test engine with pre-loaded mock assets
// This is useful for tests that need specific assets available
func NewTestEngineWithAssets(assets map[string][]byte) *EngineState {
	mockLoader := NewMockAssetLoader()
	mockDecoder := NewMockImageDecoder(100, 100)
	mockRenderer := NewMockRenderer()

	// Add assets to mock loader
	for name, data := range assets {
		mockLoader.AddFile(name, data)
	}

	// Create directory listing
	entries := make([]fs.DirEntry, 0, len(assets))
	for name := range assets {
		entries = append(entries, &MockDirEntry{name: name, isDir: false})
	}
	mockLoader.AddDir(".", entries)

	return NewEngineState(
		WithAssetLoader(mockLoader),
		WithImageDecoder(mockDecoder),
		WithRenderer(mockRenderer),
	)
}

// TestAsset represents a test asset with name and data
type TestAsset struct {
	Name string
	Data []byte
}

// NewTestEngineWithAssetList creates a test engine with a list of assets
func NewTestEngineWithAssetList(assets []TestAsset) *EngineState {
	assetMap := make(map[string][]byte)
	for _, asset := range assets {
		assetMap[asset.Name] = asset.Data
	}
	return NewTestEngineWithAssets(assetMap)
}

// AssertPictureExists checks that a picture with the given ID exists
func AssertPictureExists(t *testing.T, engine *EngineState, picID int) *Picture {
	t.Helper()
	pic := engine.GetPicture(picID)
	if pic == nil {
		t.Fatalf("Expected picture %d to exist, but it does not", picID)
	}
	return pic
}

// AssertPictureNotExists checks that a picture with the given ID does not exist
func AssertPictureNotExists(t *testing.T, engine *EngineState, picID int) {
	t.Helper()
	pic := engine.GetPicture(picID)
	if pic != nil {
		t.Fatalf("Expected picture %d to not exist, but it does", picID)
	}
}

// AssertPictureDimensions checks that a picture has the expected dimensions
func AssertPictureDimensions(t *testing.T, pic *Picture, expectedWidth, expectedHeight int) {
	t.Helper()
	if pic.Width != expectedWidth || pic.Height != expectedHeight {
		t.Errorf("Expected picture dimensions %dx%d, got %dx%d",
			expectedWidth, expectedHeight, pic.Width, pic.Height)
	}
}

// AssertWindowExists checks that a window with the given ID exists
func AssertWindowExists(t *testing.T, engine *EngineState, winID int) *Window {
	t.Helper()
	win := engine.GetWindow(winID)
	if win == nil {
		t.Fatalf("Expected window %d to exist, but it does not", winID)
	}
	return win
}

// AssertWindowNotExists checks that a window with the given ID does not exist
func AssertWindowNotExists(t *testing.T, engine *EngineState, winID int) {
	t.Helper()
	win := engine.GetWindow(winID)
	if win != nil {
		t.Fatalf("Expected window %d to not exist, but it does", winID)
	}
}

// AssertWindowProperties checks that a window has the expected properties
func AssertWindowProperties(t *testing.T, win *Window, expectedPicID, expectedX, expectedY, expectedW, expectedH int) {
	t.Helper()
	if win.Picture != expectedPicID {
		t.Errorf("Expected window picture ID %d, got %d", expectedPicID, win.Picture)
	}
	if win.X != expectedX || win.Y != expectedY {
		t.Errorf("Expected window position (%d, %d), got (%d, %d)",
			expectedX, expectedY, win.X, win.Y)
	}
	if win.W != expectedW || win.H != expectedH {
		t.Errorf("Expected window size %dx%d, got %dx%d",
			expectedW, expectedH, win.W, win.H)
	}
}

// AssertCastExists checks that a cast with the given ID exists
func AssertCastExists(t *testing.T, engine *EngineState, castID int) *Cast {
	t.Helper()
	cast := engine.GetCast(castID)
	if cast == nil {
		t.Fatalf("Expected cast %d to exist, but it does not", castID)
	}
	return cast
}

// AssertCastNotExists checks that a cast with the given ID does not exist
func AssertCastNotExists(t *testing.T, engine *EngineState, castID int) {
	t.Helper()
	cast := engine.GetCast(castID)
	if cast != nil {
		t.Fatalf("Expected cast %d to not exist, but it does", castID)
	}
}

// AssertCastProperties checks that a cast has the expected properties
func AssertCastProperties(t *testing.T, cast *Cast, expectedPicID, expectedDestPicID, expectedX, expectedY int) {
	t.Helper()
	if cast.Picture != expectedPicID {
		t.Errorf("Expected cast picture ID %d, got %d", expectedPicID, cast.Picture)
	}
	if cast.DestPicture != expectedDestPicID {
		t.Errorf("Expected cast destination picture ID %d, got %d", expectedDestPicID, cast.DestPicture)
	}
	if cast.X != expectedX || cast.Y != expectedY {
		t.Errorf("Expected cast position (%d, %d), got (%d, %d)",
			expectedX, expectedY, cast.X, cast.Y)
	}
}

// AssertIDCounters checks that ID counters have the expected values
func AssertIDCounters(t *testing.T, engine *EngineState, expectedPicID, expectedWinID, expectedCastID int) {
	t.Helper()
	if engine.nextPicID != expectedPicID {
		t.Errorf("Expected nextPicID %d, got %d", expectedPicID, engine.nextPicID)
	}
	if engine.nextWinID != expectedWinID {
		t.Errorf("Expected nextWinID %d, got %d", expectedWinID, engine.nextWinID)
	}
	if engine.nextCastID != expectedCastID {
		t.Errorf("Expected nextCastID %d, got %d", expectedCastID, engine.nextCastID)
	}
}

// AssertResourceCount checks that the engine has the expected number of resources
func AssertResourceCount(t *testing.T, engine *EngineState, expectedPics, expectedWins, expectedCasts int) {
	t.Helper()
	if len(engine.pictures) != expectedPics {
		t.Errorf("Expected %d pictures, got %d", expectedPics, len(engine.pictures))
	}
	if len(engine.windows) != expectedWins {
		t.Errorf("Expected %d windows, got %d", expectedWins, len(engine.windows))
	}
	if len(engine.casts) != expectedCasts {
		t.Errorf("Expected %d casts, got %d", expectedCasts, len(engine.casts))
	}
}

// AssertStateConsistency verifies that the engine state is internally consistent
// This checks for common invariants that should always hold
func AssertStateConsistency(t *testing.T, engine *EngineState) {
	t.Helper()

	// Check that all pictures referenced by windows exist
	for winID, win := range engine.windows {
		if _, exists := engine.pictures[win.Picture]; !exists {
			t.Errorf("Window %d references non-existent picture %d", winID, win.Picture)
		}
	}

	// Check that all pictures referenced by casts exist
	for castID, cast := range engine.casts {
		if _, exists := engine.pictures[cast.Picture]; !exists {
			t.Errorf("Cast %d references non-existent source picture %d", castID, cast.Picture)
		}
		if _, exists := engine.pictures[cast.DestPicture]; !exists {
			t.Errorf("Cast %d references non-existent destination picture %d", castID, cast.DestPicture)
		}
	}

	// Check that castDrawOrder contains only valid cast IDs
	for _, castID := range engine.castDrawOrder {
		if _, exists := engine.casts[castID]; !exists {
			t.Errorf("castDrawOrder contains non-existent cast ID %d", castID)
		}
	}

	// Check that windowOrder contains only valid window IDs
	for _, winID := range engine.windowOrder {
		if _, exists := engine.windows[winID]; !exists {
			t.Errorf("windowOrder contains non-existent window ID %d", winID)
		}
	}

	// Check that all casts in the map are in castDrawOrder
	if len(engine.casts) != len(engine.castDrawOrder) {
		t.Errorf("Cast count mismatch: %d casts in map, %d in draw order",
			len(engine.casts), len(engine.castDrawOrder))
	}

	// Check that all windows in the map are in windowOrder
	if len(engine.windows) != len(engine.windowOrder) {
		t.Errorf("Window count mismatch: %d windows in map, %d in order",
			len(engine.windows), len(engine.windowOrder))
	}

	// Check that ID counters are >= the highest ID in use
	maxPicID := -1
	for picID := range engine.pictures {
		if picID > maxPicID {
			maxPicID = picID
		}
	}
	if maxPicID >= 0 && engine.nextPicID <= maxPicID {
		t.Errorf("nextPicID (%d) should be > highest picture ID (%d)", engine.nextPicID, maxPicID)
	}

	maxWinID := -1
	for winID := range engine.windows {
		if winID > maxWinID {
			maxWinID = winID
		}
	}
	if maxWinID >= 0 && engine.nextWinID <= maxWinID {
		t.Errorf("nextWinID (%d) should be > highest window ID (%d)", engine.nextWinID, maxWinID)
	}

	maxCastID := 0
	for castID := range engine.casts {
		if castID > maxCastID {
			maxCastID = castID
		}
	}
	if maxCastID > 0 && engine.nextCastID <= maxCastID {
		t.Errorf("nextCastID (%d) should be > highest cast ID (%d)", engine.nextCastID, maxCastID)
	}
}

// CreateTestImage creates a simple test image with the specified dimensions and color
func CreateTestImage(width, height int, col color.Color) *ebiten.Image {
	img := ebiten.NewImage(width, height)
	img.Fill(col)
	return img
}

// CreateTestImageWithPattern creates a test image with a simple pattern
// This is useful for visual verification in tests
func CreateTestImageWithPattern(width, height int) *ebiten.Image {
	img := ebiten.NewImage(width, height)

	// Create a checkerboard pattern
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if (x/10+y/10)%2 == 0 {
				img.Set(x, y, color.RGBA{255, 0, 0, 255}) // Red
			} else {
				img.Set(x, y, color.RGBA{0, 0, 255, 255}) // Blue
			}
		}
	}

	return img
}

// CreateTestPicture creates a test picture and adds it to the engine
func CreateTestPicture(engine *EngineState, width, height int) int {
	engine.renderMutex.Lock()
	defer engine.renderMutex.Unlock()

	picID := engine.nextPicID
	engine.nextPicID++

	pic := &Picture{
		ID:     picID,
		Image:  CreateTestImage(width, height, color.RGBA{255, 255, 255, 255}),
		Width:  width,
		Height: height,
	}
	engine.pictures[picID] = pic

	return picID
}

// CreateTestPictureWithColor creates a test picture with a specific color
func CreateTestPictureWithColor(engine *EngineState, width, height int, col color.Color) int {
	engine.renderMutex.Lock()
	defer engine.renderMutex.Unlock()

	picID := engine.nextPicID
	engine.nextPicID++

	pic := &Picture{
		ID:     picID,
		Image:  CreateTestImage(width, height, col),
		Width:  width,
		Height: height,
	}
	engine.pictures[picID] = pic

	return picID
}

// GetPixelColor gets the color of a pixel in a picture
// This is useful for verifying rendering operations
func GetPixelColor(pic *Picture, x, y int) color.Color {
	if pic.Image == nil {
		return nil
	}
	return pic.Image.At(x, y)
}

// AssertPixelColor checks that a pixel has the expected color
func AssertPixelColor(t *testing.T, pic *Picture, x, y int, expected color.Color) {
	t.Helper()
	actual := GetPixelColor(pic, x, y)
	if actual == nil {
		t.Fatalf("Could not get pixel color at (%d, %d)", x, y)
	}

	// Convert to RGBA for comparison
	r1, g1, b1, a1 := expected.RGBA()
	r2, g2, b2, a2 := actual.RGBA()

	if r1 != r2 || g1 != g2 || b1 != b2 || a1 != a2 {
		t.Errorf("Expected pixel at (%d, %d) to be %v, got %v", x, y, expected, actual)
	}
}

// AssertImageDimensions checks that an image has the expected dimensions
func AssertImageDimensions(t *testing.T, img image.Image, expectedWidth, expectedHeight int) {
	t.Helper()
	bounds := img.Bounds()
	actualWidth := bounds.Dx()
	actualHeight := bounds.Dy()

	if actualWidth != expectedWidth || actualHeight != expectedHeight {
		t.Errorf("Expected image dimensions %dx%d, got %dx%d",
			expectedWidth, expectedHeight, actualWidth, actualHeight)
	}
}

// FixtureData provides common test data
type FixtureData struct {
	// Common test images
	SmallImage  []byte
	MediumImage []byte
	LargeImage  []byte

	// Common test assets
	TestBMP  []byte
	TestMIDI []byte
	TestWAV  []byte
}

// NewFixtureData creates a new FixtureData with common test data
func NewFixtureData() *FixtureData {
	return &FixtureData{
		SmallImage:  []byte("small image data"),
		MediumImage: []byte("medium image data"),
		LargeImage:  []byte("large image data"),
		TestBMP:     []byte("BM test bmp data"),
		TestMIDI:    []byte("MThd test midi data"),
		TestWAV:     []byte("RIFF test wav data"),
	}
}

// GetTestAssets returns a map of common test assets
func GetTestAssets() map[string][]byte {
	fixtures := NewFixtureData()
	return map[string][]byte{
		"test.bmp":   fixtures.TestBMP,
		"test.mid":   fixtures.TestMIDI,
		"test.wav":   fixtures.TestWAV,
		"small.bmp":  fixtures.SmallImage,
		"medium.bmp": fixtures.MediumImage,
		"large.bmp":  fixtures.LargeImage,
	}
}

// GetTestAssetList returns a list of common test assets
func GetTestAssetList() []TestAsset {
	fixtures := NewFixtureData()
	return []TestAsset{
		{Name: "test.bmp", Data: fixtures.TestBMP},
		{Name: "test.mid", Data: fixtures.TestMIDI},
		{Name: "test.wav", Data: fixtures.TestWAV},
		{Name: "small.bmp", Data: fixtures.SmallImage},
		{Name: "medium.bmp", Data: fixtures.MediumImage},
		{Name: "large.bmp", Data: fixtures.LargeImage},
	}
}

// ResetEngineForTest resets all global engine state for testing
// This should be called at the beginning of each test that uses global state
func ResetEngineForTest() {
	vmLock.Lock()
	defer vmLock.Unlock()

	// Reset sequencers
	mainSequencer = nil
	sequencers = nil

	// Reset global variables
	globalVars = make(map[string]any)

	// Reset timing state
	tickCount = 0
	ticksPerStep = 12
	midiSyncMode = false
	GlobalPPQ = 480

	// Reset program termination flag
	programTerminated = false

	// Reset global engine
	if globalEngine != nil {
		globalEngine.Reset()
	}
}
