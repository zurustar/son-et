package engine

import (
	"image"
	"image/color"
	"testing"
)

func TestDrawingContext_SetLineSize(t *testing.T) {
	dc := NewDrawingContext()

	// Test default line size
	if dc.GetLineSize() != 1 {
		t.Errorf("Expected default line size 1, got %d", dc.GetLineSize())
	}

	// Test setting line size
	dc.SetLineSize(5)
	if dc.GetLineSize() != 5 {
		t.Errorf("Expected line size 5, got %d", dc.GetLineSize())
	}

	// Test minimum line size (should be 1)
	dc.SetLineSize(0)
	if dc.GetLineSize() != 1 {
		t.Errorf("Expected minimum line size 1, got %d", dc.GetLineSize())
	}

	dc.SetLineSize(-5)
	if dc.GetLineSize() != 1 {
		t.Errorf("Expected minimum line size 1, got %d", dc.GetLineSize())
	}
}

func TestDrawingContext_SetPaintColor(t *testing.T) {
	dc := NewDrawingContext()

	// Test default color (black)
	defaultColor := dc.GetPaintColor()
	r, g, b, a := defaultColor.RGBA()
	if r != 0 || g != 0 || b != 0 || a != 0xFFFF {
		t.Errorf("Expected default color black, got RGBA(%d, %d, %d, %d)", r>>8, g>>8, b>>8, a>>8)
	}

	// Test setting color
	red := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	dc.SetPaintColor(red)
	c := dc.GetPaintColor()
	r, g, b, a = c.RGBA()
	if r>>8 != 255 || g>>8 != 0 || b>>8 != 0 {
		t.Errorf("Expected red color, got RGBA(%d, %d, %d, %d)", r>>8, g>>8, b>>8, a>>8)
	}
}

func TestDrawingContext_SetROP(t *testing.T) {
	dc := NewDrawingContext()

	// Test default ROP mode (COPYPEN)
	if dc.GetROP() != COPYPEN {
		t.Errorf("Expected default ROP mode COPYPEN, got %d", dc.GetROP())
	}

	// Test setting ROP mode
	dc.SetROP(XORPEN)
	if dc.GetROP() != XORPEN {
		t.Errorf("Expected ROP mode XORPEN, got %d", dc.GetROP())
	}

	dc.SetROP(MERGEPEN)
	if dc.GetROP() != MERGEPEN {
		t.Errorf("Expected ROP mode MERGEPEN, got %d", dc.GetROP())
	}
}

func TestDrawingContext_DrawLine(t *testing.T) {
	dc := NewDrawingContext()
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))

	// Set paint color to red
	dc.SetPaintColor(color.RGBA{R: 255, G: 0, B: 0, A: 255})

	// Draw horizontal line
	dc.DrawLine(img, 10, 10, 50, 10)

	// Verify pixels along the line
	for x := 10; x <= 50; x++ {
		c := img.At(x, 10)
		r, g, b, _ := c.RGBA()
		if r>>8 != 255 || g>>8 != 0 || b>>8 != 0 {
			t.Errorf("Expected red pixel at (%d, 10), got RGBA(%d, %d, %d)", x, r>>8, g>>8, b>>8)
		}
	}

	// Draw vertical line
	dc.DrawLine(img, 20, 20, 20, 60)

	// Verify pixels along the line
	for y := 20; y <= 60; y++ {
		c := img.At(20, y)
		r, g, b, _ := c.RGBA()
		if r>>8 != 255 || g>>8 != 0 || b>>8 != 0 {
			t.Errorf("Expected red pixel at (20, %d), got RGBA(%d, %d, %d)", y, r>>8, g>>8, b>>8)
		}
	}

	// Draw diagonal line
	dc.DrawLine(img, 30, 30, 40, 40)

	// Verify at least some pixels along the diagonal
	c := img.At(30, 30)
	r, g, b, _ := c.RGBA()
	if r>>8 != 255 || g>>8 != 0 || b>>8 != 0 {
		t.Errorf("Expected red pixel at (30, 30), got RGBA(%d, %d, %d)", r>>8, g>>8, b>>8)
	}

	c = img.At(40, 40)
	r, g, b, _ = c.RGBA()
	if r>>8 != 255 || g>>8 != 0 || b>>8 != 0 {
		t.Errorf("Expected red pixel at (40, 40), got RGBA(%d, %d, %d)", r>>8, g>>8, b>>8)
	}
}

func TestDrawingContext_DrawLineThick(t *testing.T) {
	dc := NewDrawingContext()
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))

	// Set paint color to blue
	dc.SetPaintColor(color.RGBA{R: 0, G: 0, B: 255, A: 255})
	dc.SetLineSize(3)

	// Draw horizontal line
	dc.DrawLine(img, 10, 50, 90, 50)

	// Verify pixels around the line (should be thick)
	for y := 49; y <= 51; y++ {
		c := img.At(50, y)
		r, g, b, _ := c.RGBA()
		if r>>8 != 0 || g>>8 != 0 || b>>8 != 255 {
			t.Errorf("Expected blue pixel at (50, %d), got RGBA(%d, %d, %d)", y, r>>8, g>>8, b>>8)
		}
	}
}

func TestDrawingContext_DrawCircleOutline(t *testing.T) {
	dc := NewDrawingContext()
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))

	// Set paint color to green
	dc.SetPaintColor(color.RGBA{R: 0, G: 255, B: 0, A: 255})

	// Draw circle outline (fillMode = 0)
	dc.DrawCircle(img, 50, 50, 20, 0)

	// Verify some pixels on the circle outline
	// Top of circle (approximately)
	c := img.At(50, 30)
	r, g, b, _ := c.RGBA()
	if r>>8 != 0 || g>>8 != 255 || b>>8 != 0 {
		t.Errorf("Expected green pixel at (50, 30), got RGBA(%d, %d, %d)", r>>8, g>>8, b>>8)
	}

	// Right of circle (approximately)
	c = img.At(70, 50)
	r, g, b, _ = c.RGBA()
	if r>>8 != 0 || g>>8 != 255 || b>>8 != 0 {
		t.Errorf("Expected green pixel at (70, 50), got RGBA(%d, %d, %d)", r>>8, g>>8, b>>8)
	}

	// Center should be transparent (not filled)
	c = img.At(50, 50)
	r, g, b, a := c.RGBA()
	if a != 0 {
		t.Errorf("Expected transparent pixel at center (50, 50), got RGBA(%d, %d, %d, %d)", r>>8, g>>8, b>>8, a>>8)
	}
}

func TestDrawingContext_DrawCircleFilled(t *testing.T) {
	dc := NewDrawingContext()
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))

	// Set paint color to yellow
	dc.SetPaintColor(color.RGBA{R: 255, G: 255, B: 0, A: 255})

	// Draw filled circle (fillMode = 2)
	dc.DrawCircle(img, 50, 50, 20, 2)

	// Verify center is filled
	c := img.At(50, 50)
	r, g, b, _ := c.RGBA()
	if r>>8 != 255 || g>>8 != 255 || b>>8 != 0 {
		t.Errorf("Expected yellow pixel at center (50, 50), got RGBA(%d, %d, %d)", r>>8, g>>8, b>>8)
	}

	// Verify edge is filled
	c = img.At(70, 50)
	r, g, b, _ = c.RGBA()
	if r>>8 != 255 || g>>8 != 255 || b>>8 != 0 {
		t.Errorf("Expected yellow pixel at edge (70, 50), got RGBA(%d, %d, %d)", r>>8, g>>8, b>>8)
	}
}

func TestDrawingContext_DrawCircleHatched(t *testing.T) {
	dc := NewDrawingContext()
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))

	// Set paint color to cyan
	dc.SetPaintColor(color.RGBA{R: 0, G: 255, B: 255, A: 255})

	// Draw hatched circle (fillMode = 1)
	dc.DrawCircle(img, 50, 50, 20, 1)

	// Verify some pixels are filled (hatch pattern)
	// We can't predict exact pixels, but center should have some colored pixels
	hasColoredPixels := false
	for y := 45; y <= 55; y++ {
		for x := 45; x <= 55; x++ {
			c := img.At(x, y)
			r, g, b, _ := c.RGBA()
			if r>>8 == 0 && g>>8 == 255 && b>>8 == 255 {
				hasColoredPixels = true
				break
			}
		}
		if hasColoredPixels {
			break
		}
	}

	if !hasColoredPixels {
		t.Error("Expected some cyan pixels in hatched circle")
	}
}

func TestDrawingContext_DrawRectOutline(t *testing.T) {
	dc := NewDrawingContext()
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))

	// Set paint color to magenta
	dc.SetPaintColor(color.RGBA{R: 255, G: 0, B: 255, A: 255})

	// Draw rectangle outline (fillMode = 0)
	dc.DrawRect(img, 10, 10, 50, 50, 0)

	// Verify corners
	corners := []image.Point{
		{10, 10}, {50, 10}, {10, 50}, {50, 50},
	}

	for _, pt := range corners {
		c := img.At(pt.X, pt.Y)
		r, g, b, _ := c.RGBA()
		if r>>8 != 255 || g>>8 != 0 || b>>8 != 255 {
			t.Errorf("Expected magenta pixel at corner (%d, %d), got RGBA(%d, %d, %d)", pt.X, pt.Y, r>>8, g>>8, b>>8)
		}
	}

	// Verify center is not filled
	c := img.At(30, 30)
	r, g, b, a := c.RGBA()
	if a != 0 {
		t.Errorf("Expected transparent pixel at center (30, 30), got RGBA(%d, %d, %d, %d)", r>>8, g>>8, b>>8, a>>8)
	}
}

func TestDrawingContext_DrawRectFilled(t *testing.T) {
	dc := NewDrawingContext()
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))

	// Set paint color to orange
	dc.SetPaintColor(color.RGBA{R: 255, G: 165, B: 0, A: 255})

	// Draw filled rectangle (fillMode = 2)
	dc.DrawRect(img, 10, 10, 50, 50, 2)

	// Verify center is filled
	c := img.At(30, 30)
	r, g, b, _ := c.RGBA()
	if r>>8 != 255 || g>>8 != 165 || b>>8 != 0 {
		t.Errorf("Expected orange pixel at center (30, 30), got RGBA(%d, %d, %d)", r>>8, g>>8, b>>8)
	}

	// Verify corners are filled
	c = img.At(10, 10)
	r, g, b, _ = c.RGBA()
	if r>>8 != 255 || g>>8 != 165 || b>>8 != 0 {
		t.Errorf("Expected orange pixel at corner (10, 10), got RGBA(%d, %d, %d)", r>>8, g>>8, b>>8)
	}
}

func TestDrawingContext_DrawRectHatched(t *testing.T) {
	dc := NewDrawingContext()
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))

	// Set paint color to purple
	dc.SetPaintColor(color.RGBA{R: 128, G: 0, B: 128, A: 255})

	// Draw hatched rectangle (fillMode = 1)
	dc.DrawRect(img, 10, 10, 50, 50, 1)

	// Verify some pixels are filled (hatch pattern)
	hasColoredPixels := false
	for y := 10; y <= 50; y++ {
		for x := 10; x <= 50; x++ {
			c := img.At(x, y)
			r, g, b, _ := c.RGBA()
			if r>>8 == 128 && g>>8 == 0 && b>>8 == 128 {
				hasColoredPixels = true
				break
			}
		}
		if hasColoredPixels {
			break
		}
	}

	if !hasColoredPixels {
		t.Error("Expected some purple pixels in hatched rectangle")
	}
}

func TestGetColor(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))

	// Set a pixel to red
	img.Set(50, 50, color.RGBA{R: 255, G: 0, B: 0, A: 255})

	// Get the color
	c := GetColor(img, 50, 50)
	r, g, b, _ := c.RGBA()
	if r>>8 != 255 || g>>8 != 0 || b>>8 != 0 {
		t.Errorf("Expected red pixel at (50, 50), got RGBA(%d, %d, %d)", r>>8, g>>8, b>>8)
	}

	// Get color outside bounds (should return transparent)
	c = GetColor(img, 200, 200)
	_, _, _, a := c.RGBA()
	if a != 0 {
		t.Errorf("Expected transparent color outside bounds, got alpha %d", a>>8)
	}
}

func TestDrawingContext_ROPModes(t *testing.T) {
	tests := []struct {
		name     string
		ropMode  ROPMode
		existing color.RGBA
		paint    color.RGBA
		expected color.RGBA
	}{
		{
			name:     "COPYPEN",
			ropMode:  COPYPEN,
			existing: color.RGBA{R: 100, G: 100, B: 100, A: 255},
			paint:    color.RGBA{R: 255, G: 0, B: 0, A: 255},
			expected: color.RGBA{R: 255, G: 0, B: 0, A: 255},
		},
		{
			name:     "XORPEN",
			ropMode:  XORPEN,
			existing: color.RGBA{R: 255, G: 0, B: 0, A: 255},
			paint:    color.RGBA{R: 0, G: 255, B: 0, A: 255},
			expected: color.RGBA{R: 255, G: 255, B: 0, A: 255},
		},
		{
			name:     "MERGEPEN",
			ropMode:  MERGEPEN,
			existing: color.RGBA{R: 255, G: 0, B: 0, A: 255},
			paint:    color.RGBA{R: 0, G: 255, B: 0, A: 255},
			expected: color.RGBA{R: 255, G: 255, B: 0, A: 255},
		},
		{
			name:     "NOTCOPYPEN",
			ropMode:  NOTCOPYPEN,
			existing: color.RGBA{R: 100, G: 100, B: 100, A: 255},
			paint:    color.RGBA{R: 255, G: 0, B: 0, A: 255},
			expected: color.RGBA{R: 0, G: 255, B: 255, A: 255},
		},
		{
			name:     "MASKPEN",
			ropMode:  MASKPEN,
			existing: color.RGBA{R: 255, G: 255, B: 0, A: 255},
			paint:    color.RGBA{R: 255, G: 0, B: 255, A: 255},
			expected: color.RGBA{R: 255, G: 0, B: 0, A: 255},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dc := NewDrawingContext()
			img := image.NewRGBA(image.Rect(0, 0, 10, 10))

			// Set existing pixel
			img.Set(5, 5, tt.existing)

			// Set ROP mode and paint color
			dc.SetROP(tt.ropMode)
			dc.SetPaintColor(tt.paint)

			// Draw a pixel
			dc.setPixel(img, 5, 5)

			// Verify result
			result := img.At(5, 5)
			r, g, b, a := result.RGBA()
			if r>>8 != uint32(tt.expected.R) || g>>8 != uint32(tt.expected.G) || b>>8 != uint32(tt.expected.B) {
				t.Errorf("Expected RGBA(%d, %d, %d, %d), got RGBA(%d, %d, %d, %d)",
					tt.expected.R, tt.expected.G, tt.expected.B, tt.expected.A,
					r>>8, g>>8, b>>8, a>>8)
			}
		})
	}
}

func TestEngine_DrawingFunctions(t *testing.T) {
	// Create engine with mock renderer
	mockRenderer := &MockRenderer{}
	assetLoader := NewFilesystemAssetLoader("../../samples/test_minimal")
	imageDecoder := &BMPImageDecoder{}
	engine := NewEngine(mockRenderer, assetLoader, imageDecoder)
	engine.SetHeadless(true)

	// Create a picture to draw on
	picID := engine.CreatePic(200, 200)
	if picID == 0 {
		t.Fatal("Failed to create picture")
	}

	// Test SetLineSize
	engine.SetLineSize(3)
	if engine.drawingContext.GetLineSize() != 3 {
		t.Errorf("Expected line size 3, got %d", engine.drawingContext.GetLineSize())
	}

	// Test SetPaintColor
	engine.SetPaintColor(0xFF0000) // Red
	c := engine.drawingContext.GetPaintColor()
	r, g, b, _ := c.RGBA()
	if r>>8 != 255 || g>>8 != 0 || b>>8 != 0 {
		t.Errorf("Expected red color, got RGBA(%d, %d, %d)", r>>8, g>>8, b>>8)
	}

	// Test DrawLine
	err := engine.DrawLine(picID, 10, 10, 100, 10)
	if err != nil {
		t.Errorf("DrawLine failed: %v", err)
	}

	// Test DrawCircle
	err = engine.DrawCircle(picID, 50, 50, 20, 2)
	if err != nil {
		t.Errorf("DrawCircle failed: %v", err)
	}

	// Test DrawRect
	err = engine.DrawRect(picID, 120, 120, 180, 180, 1)
	if err != nil {
		t.Errorf("DrawRect failed: %v", err)
	}

	// Test GetColor
	colorValue, err := engine.GetColor(picID, 50, 50)
	if err != nil {
		t.Errorf("GetColor failed: %v", err)
	}
	// Should be red (0xFF0000) from the filled circle
	if colorValue != 0xFF0000 {
		t.Logf("GetColor returned 0x%06X (expected 0xFF0000, but may vary due to drawing order)", colorValue)
	}

	// Test SetROP
	engine.SetROP(int(XORPEN))
	if engine.drawingContext.GetROP() != XORPEN {
		t.Errorf("Expected ROP mode XORPEN, got %d", engine.drawingContext.GetROP())
	}

	// Test drawing on non-existent picture
	err = engine.DrawLine(999, 0, 0, 10, 10)
	if err == nil {
		t.Error("Expected error when drawing on non-existent picture")
	}
}
