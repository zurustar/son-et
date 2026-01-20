package engine

import (
	"image/color"
	"testing"
)

func TestNewTextRenderer(t *testing.T) {
	renderer := &MockRenderer{}
	assetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
	imageDecoder := &MockImageDecoder{Width: 640, Height: 480}
	engine := NewEngine(renderer, assetLoader, imageDecoder)

	if engine.textRenderer == nil {
		t.Fatal("Text renderer not created")
	}

	if engine.textRenderer.currentFont == nil {
		t.Error("Default font not set")
	}

	if engine.textRenderer.currentFontSize != 13 {
		t.Errorf("Expected default font size 13, got %d", engine.textRenderer.currentFontSize)
	}

	expectedColor := color.RGBA{0, 0, 0, 255}
	if engine.textRenderer.textColor != expectedColor {
		t.Errorf("Expected black text color, got %v", engine.textRenderer.textColor)
	}
}

func TestSetFont(t *testing.T) {
	renderer := &MockRenderer{}
	assetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
	imageDecoder := &MockImageDecoder{Width: 640, Height: 480}
	engine := NewEngine(renderer, assetLoader, imageDecoder)

	engine.SetFont(16, "Arial", 0)

	if engine.textRenderer.currentFontSize != 16 {
		t.Errorf("Expected font size 16, got %d", engine.textRenderer.currentFontSize)
	}

	if engine.textRenderer.currentFontName != "Arial" {
		t.Errorf("Expected font name 'Arial', got %s", engine.textRenderer.currentFontName)
	}
}

func TestSetFont_LegacySize(t *testing.T) {
	renderer := &MockRenderer{}
	assetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
	imageDecoder := &MockImageDecoder{Width: 640, Height: 480}
	engine := NewEngine(renderer, assetLoader, imageDecoder)

	// Test legacy mode with unreasonable size
	engine.SetFont(640, "MS UI Gothic", 0)

	// Should use default size instead of 640
	if engine.textRenderer.currentFontSize == 640 {
		t.Error("Should not accept unreasonably large font size")
	}
}

func TestTextColor(t *testing.T) {
	renderer := &MockRenderer{}
	assetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
	imageDecoder := &MockImageDecoder{Width: 640, Height: 480}
	engine := NewEngine(renderer, assetLoader, imageDecoder)

	engine.TextColor(255, 0, 0) // Red

	expected := color.RGBA{255, 0, 0, 255}
	if engine.textRenderer.textColor != expected {
		t.Errorf("Expected red color, got %v", engine.textRenderer.textColor)
	}
}

func TestBgColor(t *testing.T) {
	renderer := &MockRenderer{}
	assetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
	imageDecoder := &MockImageDecoder{Width: 640, Height: 480}
	engine := NewEngine(renderer, assetLoader, imageDecoder)

	engine.BgColor(0, 255, 0) // Green

	expected := color.RGBA{0, 255, 0, 255}
	if engine.textRenderer.bgColor != expected {
		t.Errorf("Expected green color, got %v", engine.textRenderer.bgColor)
	}
}

func TestBackMode(t *testing.T) {
	renderer := &MockRenderer{}
	assetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
	imageDecoder := &MockImageDecoder{Width: 640, Height: 480}
	engine := NewEngine(renderer, assetLoader, imageDecoder)

	// Default should be transparent
	if engine.textRenderer.backMode != 0 {
		t.Errorf("Expected default back mode 0 (transparent), got %d", engine.textRenderer.backMode)
	}

	// Set to opaque
	engine.BackMode(1)
	if engine.textRenderer.backMode != 1 {
		t.Errorf("Expected back mode 1 (opaque), got %d", engine.textRenderer.backMode)
	}

	// Set back to transparent
	engine.BackMode(0)
	if engine.textRenderer.backMode != 0 {
		t.Errorf("Expected back mode 0 (transparent), got %d", engine.textRenderer.backMode)
	}
}

func TestTextWrite_PictureNotFound(t *testing.T) {
	renderer := &MockRenderer{}
	assetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
	imageDecoder := &MockImageDecoder{Width: 640, Height: 480}
	engine := NewEngine(renderer, assetLoader, imageDecoder)

	err := engine.TextWrite("Hello", 999, 10, 10)
	if err == nil {
		t.Error("Expected error for nonexistent picture")
	}
}

func TestTextWrite_Success(t *testing.T) {
	renderer := &MockRenderer{}
	assetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
	imageDecoder := &MockImageDecoder{Width: 640, Height: 480}
	engine := NewEngine(renderer, assetLoader, imageDecoder)

	// Create a picture
	picID := engine.CreatePic(200, 100)

	// Write text
	err := engine.TextWrite("Hello World", picID, 10, 10)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify picture still exists
	pic := engine.state.GetPicture(picID)
	if pic == nil {
		t.Error("Picture should still exist after text write")
	}
}

func TestTextWrite_TransparentBackground(t *testing.T) {
	renderer := &MockRenderer{}
	assetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
	imageDecoder := &MockImageDecoder{Width: 640, Height: 480}
	engine := NewEngine(renderer, assetLoader, imageDecoder)

	// Create a picture
	picID := engine.CreatePic(200, 100)

	// Set transparent background
	engine.BackMode(0)
	engine.TextColor(255, 0, 0) // Red text

	// Write text
	err := engine.TextWrite("Test", picID, 10, 10)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestTextWrite_OpaqueBackground(t *testing.T) {
	renderer := &MockRenderer{}
	assetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
	imageDecoder := &MockImageDecoder{Width: 640, Height: 480}
	engine := NewEngine(renderer, assetLoader, imageDecoder)

	// Create a picture
	picID := engine.CreatePic(200, 100)

	// Set opaque background
	engine.BackMode(1)
	engine.TextColor(0, 0, 255) // Blue text
	engine.BgColor(255, 255, 0) // Yellow background

	// Write text
	err := engine.TextWrite("Test", picID, 10, 10)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestMeasureText(t *testing.T) {
	renderer := &MockRenderer{}
	assetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
	imageDecoder := &MockImageDecoder{Width: 640, Height: 480}
	engine := NewEngine(renderer, assetLoader, imageDecoder)

	width, height := engine.MeasureText("Hello")

	if width <= 0 {
		t.Errorf("Expected positive width, got %d", width)
	}

	if height <= 0 {
		t.Errorf("Expected positive height, got %d", height)
	}

	// Longer text should have greater width
	width2, _ := engine.MeasureText("Hello World")
	if width2 <= width {
		t.Errorf("Expected longer text to have greater width: %d vs %d", width2, width)
	}
}

func TestTextWrite_MultipleLines(t *testing.T) {
	renderer := &MockRenderer{}
	assetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
	imageDecoder := &MockImageDecoder{Width: 640, Height: 480}
	engine := NewEngine(renderer, assetLoader, imageDecoder)

	// Create a picture
	picID := engine.CreatePic(300, 200)

	// Write multiple lines of text
	err1 := engine.TextWrite("Line 1", picID, 10, 10)
	if err1 != nil {
		t.Errorf("Unexpected error on line 1: %v", err1)
	}

	err2 := engine.TextWrite("Line 2", picID, 10, 30)
	if err2 != nil {
		t.Errorf("Unexpected error on line 2: %v", err2)
	}

	err3 := engine.TextWrite("Line 3", picID, 10, 50)
	if err3 != nil {
		t.Errorf("Unexpected error on line 3: %v", err3)
	}
}

func TestTextWrite_DifferentColors(t *testing.T) {
	renderer := &MockRenderer{}
	assetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
	imageDecoder := &MockImageDecoder{Width: 640, Height: 480}
	engine := NewEngine(renderer, assetLoader, imageDecoder)

	// Create a picture
	picID := engine.CreatePic(300, 200)

	// Write text in different colors
	engine.TextColor(255, 0, 0) // Red
	err1 := engine.TextWrite("Red", picID, 10, 10)
	if err1 != nil {
		t.Errorf("Unexpected error: %v", err1)
	}

	engine.TextColor(0, 255, 0) // Green
	err2 := engine.TextWrite("Green", picID, 10, 30)
	if err2 != nil {
		t.Errorf("Unexpected error: %v", err2)
	}

	engine.TextColor(0, 0, 255) // Blue
	err3 := engine.TextWrite("Blue", picID, 10, 50)
	if err3 != nil {
		t.Errorf("Unexpected error: %v", err3)
	}
}

func TestTextWrite_EmptyString(t *testing.T) {
	renderer := &MockRenderer{}
	assetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
	imageDecoder := &MockImageDecoder{Width: 640, Height: 480}
	engine := NewEngine(renderer, assetLoader, imageDecoder)

	// Create a picture
	picID := engine.CreatePic(200, 100)

	// Write empty string (should not error)
	err := engine.TextWrite("", picID, 10, 10)
	if err != nil {
		t.Errorf("Unexpected error for empty string: %v", err)
	}
}
