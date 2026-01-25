package graphics

import (
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/image/bmp"
)

// createTestBMP creates a test BMP file
func createTestBMP(t *testing.T, path string, width, height int) {
	t.Helper()

	// Create a simple test image
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{uint8(x % 256), uint8(y % 256), 128, 255})
		}
	}

	// Write BMP file
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("Failed to create test BMP: %v", err)
	}
	defer file.Close()

	if err := bmp.Encode(file, img); err != nil {
		t.Fatalf("Failed to encode BMP: %v", err)
	}
}

func TestNewPictureManager(t *testing.T) {
	pm := NewPictureManager("/test/path")

	if pm == nil {
		t.Fatal("NewPictureManager returned nil")
	}

	if pm.basePath != "/test/path" {
		t.Errorf("Expected basePath '/test/path', got '%s'", pm.basePath)
	}

	if pm.maxID != 256 {
		t.Errorf("Expected maxID 256, got %d", pm.maxID)
	}

	if pm.nextID != 0 {
		t.Errorf("Expected nextID 0, got %d", pm.nextID)
	}

	if len(pm.pictures) != 0 {
		t.Errorf("Expected empty pictures map, got %d entries", len(pm.pictures))
	}
}

func TestCreatePic(t *testing.T) {
	pm := NewPictureManager("")

	// Create a picture
	id, err := pm.CreatePic(100, 200)
	if err != nil {
		t.Fatalf("CreatePic failed: %v", err)
	}

	if id != 0 {
		t.Errorf("Expected first picture ID to be 0, got %d", id)
	}

	// Verify picture exists
	pic, err := pm.GetPic(id)
	if err != nil {
		t.Fatalf("GetPic failed: %v", err)
	}

	if pic.Width != 100 {
		t.Errorf("Expected width 100, got %d", pic.Width)
	}

	if pic.Height != 200 {
		t.Errorf("Expected height 200, got %d", pic.Height)
	}

	if pic.Image == nil {
		t.Error("Expected non-nil Image")
	}
}

func TestCreatePicMultiple(t *testing.T) {
	pm := NewPictureManager("")

	// Create multiple pictures
	id1, err := pm.CreatePic(50, 50)
	if err != nil {
		t.Fatalf("CreatePic 1 failed: %v", err)
	}

	id2, err := pm.CreatePic(100, 100)
	if err != nil {
		t.Fatalf("CreatePic 2 failed: %v", err)
	}

	// IDs should be unique and sequential
	if id1 == id2 {
		t.Errorf("Picture IDs should be unique, both are %d", id1)
	}

	if id2 != id1+1 {
		t.Errorf("Expected sequential IDs, got %d and %d", id1, id2)
	}
}

func TestDelPic(t *testing.T) {
	pm := NewPictureManager("")

	// Create a picture
	id, err := pm.CreatePic(100, 100)
	if err != nil {
		t.Fatalf("CreatePic failed: %v", err)
	}

	// Delete the picture
	err = pm.DelPic(id)
	if err != nil {
		t.Fatalf("DelPic failed: %v", err)
	}

	// Verify picture no longer exists
	_, err = pm.GetPic(id)
	if err == nil {
		t.Error("Expected error when getting deleted picture, got nil")
	}
}

func TestDelPicNonExistent(t *testing.T) {
	pm := NewPictureManager("")

	// Try to delete non-existent picture
	err := pm.DelPic(999)
	if err == nil {
		t.Error("Expected error when deleting non-existent picture, got nil")
	}
}

func TestPicWidth(t *testing.T) {
	pm := NewPictureManager("")

	// Create a picture
	id, err := pm.CreatePic(123, 456)
	if err != nil {
		t.Fatalf("CreatePic failed: %v", err)
	}

	// Get width
	width := pm.PicWidth(id)
	if width != 123 {
		t.Errorf("Expected width 123, got %d", width)
	}
}

func TestPicHeight(t *testing.T) {
	pm := NewPictureManager("")

	// Create a picture
	id, err := pm.CreatePic(123, 456)
	if err != nil {
		t.Fatalf("CreatePic failed: %v", err)
	}

	// Get height
	height := pm.PicHeight(id)
	if height != 456 {
		t.Errorf("Expected height 456, got %d", height)
	}
}

func TestPicWidthNonExistent(t *testing.T) {
	pm := NewPictureManager("")

	// Get width of non-existent picture
	width := pm.PicWidth(999)
	if width != 0 {
		t.Errorf("Expected width 0 for non-existent picture, got %d", width)
	}
}

func TestPicHeightNonExistent(t *testing.T) {
	pm := NewPictureManager("")

	// Get height of non-existent picture
	height := pm.PicHeight(999)
	if height != 0 {
		t.Errorf("Expected height 0 for non-existent picture, got %d", height)
	}
}

func TestCreatePicFrom(t *testing.T) {
	pm := NewPictureManager("")

	// Create source picture
	srcID, err := pm.CreatePic(100, 100)
	if err != nil {
		t.Fatalf("CreatePic failed: %v", err)
	}

	// Draw something on the source picture
	srcPic, _ := pm.GetPic(srcID)
	srcPic.Image.Fill(color.RGBA{255, 0, 0, 255})

	// Create copy
	copyID, err := pm.CreatePicFrom(srcID)
	if err != nil {
		t.Fatalf("CreatePicFrom failed: %v", err)
	}

	// Verify copy exists and has same dimensions
	copyPic, err := pm.GetPic(copyID)
	if err != nil {
		t.Fatalf("GetPic failed: %v", err)
	}

	if copyPic.Width != 100 {
		t.Errorf("Expected copy width 100, got %d", copyPic.Width)
	}

	if copyPic.Height != 100 {
		t.Errorf("Expected copy height 100, got %d", copyPic.Height)
	}

	// IDs should be different
	if copyID == srcID {
		t.Error("Copy should have different ID from source")
	}
}

func TestCreatePicFromNonExistent(t *testing.T) {
	pm := NewPictureManager("")

	// Try to create from non-existent picture
	_, err := pm.CreatePicFrom(999)
	if err == nil {
		t.Error("Expected error when creating from non-existent picture, got nil")
	}
}

func TestLoadPic(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Create test BMP file
	testFile := filepath.Join(tmpDir, "test.bmp")
	createTestBMP(t, testFile, 50, 60)

	pm := NewPictureManager(tmpDir)

	// Load the picture
	id, err := pm.LoadPic("test.bmp")
	if err != nil {
		t.Fatalf("LoadPic failed: %v", err)
	}

	// Verify picture was loaded
	pic, err := pm.GetPic(id)
	if err != nil {
		t.Fatalf("GetPic failed: %v", err)
	}

	if pic.Width != 50 {
		t.Errorf("Expected width 50, got %d", pic.Width)
	}

	if pic.Height != 60 {
		t.Errorf("Expected height 60, got %d", pic.Height)
	}
}

func TestLoadPicCaseInsensitive(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Create test BMP file with lowercase name
	testFile := filepath.Join(tmpDir, "test.bmp")
	createTestBMP(t, testFile, 50, 60)

	pm := NewPictureManager(tmpDir)

	// Try to load with different case
	id, err := pm.LoadPic("TEST.BMP")
	if err != nil {
		t.Fatalf("LoadPic with different case failed: %v", err)
	}

	// Verify picture was loaded
	pic, err := pm.GetPic(id)
	if err != nil {
		t.Fatalf("GetPic failed: %v", err)
	}

	if pic.Width != 50 {
		t.Errorf("Expected width 50, got %d", pic.Width)
	}
}

func TestLoadPicNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	pm := NewPictureManager(tmpDir)

	// Try to load non-existent file
	id, err := pm.LoadPic("nonexistent.bmp")
	if err == nil {
		t.Error("Expected error when loading non-existent file, got nil")
	}

	if id != -1 {
		t.Errorf("Expected ID -1 for failed load, got %d", id)
	}
}

func TestResourceLimit(t *testing.T) {
	pm := NewPictureManager("")
	pm.maxID = 3 // Set low limit for testing

	// Create pictures up to limit
	for i := 0; i < 3; i++ {
		_, err := pm.CreatePic(10, 10)
		if err != nil {
			t.Fatalf("CreatePic %d failed: %v", i, err)
		}
	}

	// Try to create one more (should fail)
	id, err := pm.CreatePic(10, 10)
	if err == nil {
		t.Error("Expected error when exceeding resource limit, got nil")
	}

	if id != -1 {
		t.Errorf("Expected ID -1 when exceeding limit, got %d", id)
	}
}

func TestIDReuse(t *testing.T) {
	pm := NewPictureManager("")

	// Create a picture
	id1, err := pm.CreatePic(10, 10)
	if err != nil {
		t.Fatalf("CreatePic failed: %v", err)
	}

	// Delete it
	err = pm.DelPic(id1)
	if err != nil {
		t.Fatalf("DelPic failed: %v", err)
	}

	// Create another picture (ID should be reusable)
	id2, err := pm.CreatePic(20, 20)
	if err != nil {
		t.Fatalf("CreatePic after delete failed: %v", err)
	}

	// New ID should be different (sequential)
	if id2 == id1 {
		t.Logf("Note: ID was reused (id1=%d, id2=%d)", id1, id2)
	}

	// But we should be able to use the slot
	if len(pm.pictures) != 1 {
		t.Errorf("Expected 1 picture after delete and create, got %d", len(pm.pictures))
	}
}
