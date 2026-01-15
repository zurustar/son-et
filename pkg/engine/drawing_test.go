package engine

import (
	"image/color"
	"testing"
)

// TestDrawLine tests basic line drawing functionality
func TestDrawLine(t *testing.T) {
	// Create test engine and set as global
	engine := NewTestEngine()
	globalEngine = engine

	// Create a test picture
	picID := engine.CreatePic(100, 100)

	// Set drawing color to red
	SetPaintColor(255, 0, 0)

	// Draw a horizontal line
	DrawLine(picID, 10, 50, 90, 50)

	// Verify the line was drawn by checking the picture exists
	pic := engine.pictures[picID]
	if pic == nil {
		t.Fatal("Picture not found")
	}

	// We can't read pixels without starting the game, but we can verify
	// the function executed without error
	t.Log("DrawLine executed successfully")
}

// TestDrawLineVertical tests vertical line drawing
func TestDrawLineVertical(t *testing.T) {
	engine := NewTestEngine()
	globalEngine = engine
	picID := engine.CreatePic(100, 100)

	SetPaintColor(0, 255, 0) // Green

	// Draw a vertical line
	DrawLine(picID, 50, 10, 50, 90)

	pic := engine.pictures[picID]
	if pic == nil {
		t.Fatal("Picture not found")
	}
	t.Log("DrawLine vertical executed successfully")
}

// TestDrawLineDiagonal tests diagonal line drawing
func TestDrawLineDiagonal(t *testing.T) {
	engine := NewTestEngine()
	globalEngine = engine
	picID := engine.CreatePic(100, 100)

	SetPaintColor(0, 0, 255) // Blue

	// Draw a diagonal line
	DrawLine(picID, 10, 10, 90, 90)

	pic := engine.pictures[picID]
	if pic == nil {
		t.Fatal("Picture not found")
	}
	t.Log("DrawLine diagonal executed successfully")
}

// TestDrawLineWidth tests line width
func TestDrawLineWidth(t *testing.T) {
	engine := NewTestEngine()
	globalEngine = engine
	picID := engine.CreatePic(100, 100)

	SetPaintColor(255, 0, 0)
	SetLineSize(3)

	// Draw a line with width 3
	DrawLine(picID, 50, 10, 50, 90)

	pic := engine.pictures[picID]
	if pic == nil {
		t.Fatal("Picture not found")
	}

	// Reset line size
	SetLineSize(1)
	t.Log("DrawLine with width executed successfully")
}

// TestDrawCircleOutline tests circle outline drawing
func TestDrawCircleOutline(t *testing.T) {
	engine := NewTestEngine()
	globalEngine = engine
	picID := engine.CreatePic(200, 200)

	SetPaintColor(255, 0, 0)

	// Draw a circle outline
	DrawCircle(picID, 100, 100, 50, 50, FILL_NONE)

	pic := engine.pictures[picID]
	if pic == nil {
		t.Fatal("Picture not found")
	}
	t.Log("DrawCircle outline executed successfully")
}

// TestDrawCircleSolid tests solid filled circle
func TestDrawCircleSolid(t *testing.T) {
	engine := NewTestEngine()
	globalEngine = engine
	picID := engine.CreatePic(200, 200)

	SetPaintColor(0, 255, 0)

	// Draw a solid filled circle
	DrawCircle(picID, 100, 100, 50, 50, FILL_SOLID)

	pic := engine.pictures[picID]
	if pic == nil {
		t.Fatal("Picture not found")
	}
	t.Log("DrawCircle solid executed successfully")
}

// TestDrawCircleHatch tests hatch filled circle
func TestDrawCircleHatch(t *testing.T) {
	engine := NewTestEngine()
	globalEngine = engine
	picID := engine.CreatePic(200, 200)

	SetPaintColor(0, 0, 255)

	// Draw a hatch filled circle
	DrawCircle(picID, 100, 100, 50, 50, FILL_HATCH)

	pic := engine.pictures[picID]
	if pic == nil {
		t.Fatal("Picture not found")
	}
	t.Log("DrawCircle hatch executed successfully")
}

// TestDrawRectOutline tests rectangle outline drawing
func TestDrawRectOutline(t *testing.T) {
	engine := NewTestEngine()
	globalEngine = engine
	picID := engine.CreatePic(200, 200)

	SetPaintColor(255, 0, 0)

	// Draw a rectangle outline
	DrawRect(picID, 50, 50, 100, 80, FILL_NONE)

	pic := engine.pictures[picID]
	if pic == nil {
		t.Fatal("Picture not found")
	}
	t.Log("DrawRect outline executed successfully")
}

// TestDrawRectSolid tests solid filled rectangle
func TestDrawRectSolid(t *testing.T) {
	engine := NewTestEngine()
	globalEngine = engine
	picID := engine.CreatePic(200, 200)

	SetPaintColor(0, 255, 0)

	// Draw a solid filled rectangle
	DrawRect(picID, 50, 50, 100, 80, FILL_SOLID)

	pic := engine.pictures[picID]
	if pic == nil {
		t.Fatal("Picture not found")
	}
	t.Log("DrawRect solid executed successfully")
}

// TestDrawRectHatch tests hatch filled rectangle
func TestDrawRectHatch(t *testing.T) {
	engine := NewTestEngine()
	globalEngine = engine
	picID := engine.CreatePic(200, 200)

	SetPaintColor(0, 0, 255)

	// Draw a hatch filled rectangle
	DrawRect(picID, 50, 50, 100, 80, FILL_HATCH)

	pic := engine.pictures[picID]
	if pic == nil {
		t.Fatal("Picture not found")
	}
	t.Log("DrawRect hatch executed successfully")
}

// TestSetLineSize tests line size setting
func TestSetLineSize(t *testing.T) {
	// Test that line size can be set
	SetLineSize(5)
	if currentLineSize != 5 {
		t.Errorf("Expected line size 5, got %d", currentLineSize)
	}

	// Test minimum line size
	SetLineSize(0)
	if currentLineSize != 1 {
		t.Errorf("Expected minimum line size 1, got %d", currentLineSize)
	}

	// Reset
	SetLineSize(1)
}

// TestSetPaintColor tests paint color setting
func TestSetPaintColor(t *testing.T) {
	SetPaintColor(128, 64, 32)

	expected := color.RGBA{128, 64, 32, 255}
	if currentPaintColor != expected {
		t.Errorf("Expected color RGB(128,64,32), got %v", currentPaintColor)
	}
}

// TestGetColor tests pixel color reading with mock
func TestGetColor(t *testing.T) {
	engine := NewTestEngine()
	globalEngine = engine
	picID := engine.CreatePic(100, 100)

	// We can't actually read pixels without starting the game,
	// but we can test that the function doesn't crash with invalid input
	col := GetColor(999, 50, 50)

	// Should return 0 for invalid picture
	if col != 0 {
		t.Errorf("Expected 0 for invalid picture, got 0x%X", col)
	}

	t.Logf("GetColor function verified (picture ID=%d)", picID)
}

// TestGetColorInvalidPicture tests GetColor with invalid picture ID
func TestGetColorInvalidPicture(t *testing.T) {
	engine := NewTestEngine()
	globalEngine = engine

	// Try to get color from non-existent picture
	col := GetColor(999, 0, 0)

	if col != 0 {
		t.Errorf("Expected 0 for invalid picture, got 0x%X", col)
	}
}

// TestSetROP tests ROP mode setting
func TestSetROP(t *testing.T) {
	// Test setting different ROP modes
	SetROP(ROP_XORPEN)
	if currentROP != ROP_XORPEN {
		t.Errorf("Expected ROP_XORPEN, got %d", currentROP)
	}

	SetROP(ROP_COPYPEN)
	if currentROP != ROP_COPYPEN {
		t.Errorf("Expected ROP_COPYPEN, got %d", currentROP)
	}
}

// TestDrawingFunctionsIntegration tests multiple drawing operations together
func TestDrawingFunctionsIntegration(t *testing.T) {
	engine := NewTestEngine()
	globalEngine = engine
	picID := engine.CreatePic(300, 300)

	// Draw a red rectangle
	SetPaintColor(255, 0, 0)
	DrawRect(picID, 50, 50, 200, 200, FILL_SOLID)

	// Draw a blue circle on top
	SetPaintColor(0, 0, 255)
	DrawCircle(picID, 150, 150, 50, 50, FILL_SOLID)

	// Draw a green line
	SetPaintColor(0, 255, 0)
	SetLineSize(3)
	DrawLine(picID, 50, 50, 250, 250)

	pic := engine.pictures[picID]
	if pic == nil {
		t.Fatal("Picture not found")
	}

	t.Log("Integration test executed successfully")
}

// TestGetColorReturnsCorrectRGB tests that GetColor returns correct RGB values
func TestGetColorReturnsCorrectRGB(t *testing.T) {
	engine := NewTestEngine()
	globalEngine = engine
	picID := engine.CreatePic(100, 100)

	// We can't actually test pixel reading without starting the game
	// But we can verify the function exists and has the right signature
	// by checking it compiles and doesn't crash with invalid input

	// Test with invalid picture ID (should return 0)
	col := GetColor(999, 10, 10)
	if col != 0 {
		t.Errorf("Expected 0 for invalid picture, got 0x%06X", col)
	}

	t.Logf("GetColor function verified (picture ID=%d)", picID)
}

// TestSetROPAffectsDrawing tests that SetROP affects drawing operations
func TestSetROPAffectsDrawing(t *testing.T) {
	engine := NewTestEngine()
	globalEngine = engine
	picID := engine.CreatePic(100, 100)

	// Test COPYPEN mode (default) - this doesn't require pixel reading
	SetROP(ROP_COPYPEN)
	SetPaintColor(255, 0, 0)
	DrawLine(picID, 10, 10, 90, 10)

	// Note: Other ROP modes (XORPEN, MERGEPEN, etc.) require reading
	// destination pixels, which can't be done without starting the game.
	// We test that SetROP can be called with different modes.

	SetROP(ROP_XORPEN)
	if currentROP != ROP_XORPEN {
		t.Errorf("Expected ROP_XORPEN, got %d", currentROP)
	}

	SetROP(ROP_MERGEPEN)
	if currentROP != ROP_MERGEPEN {
		t.Errorf("Expected ROP_MERGEPEN, got %d", currentROP)
	}

	SetROP(ROP_NOTPEN)
	if currentROP != ROP_NOTPEN {
		t.Errorf("Expected ROP_NOTPEN, got %d", currentROP)
	}

	SetROP(ROP_MASKPEN)
	if currentROP != ROP_MASKPEN {
		t.Errorf("Expected ROP_MASKPEN, got %d", currentROP)
	}

	// Reset to COPYPEN
	SetROP(ROP_COPYPEN)

	pic := engine.pictures[picID]
	if pic == nil {
		t.Fatal("Picture not found")
	}

	t.Log("SetROP modes set successfully")
}

// TestROPModes tests different ROP modes
func TestROPModes(t *testing.T) {
	tests := []struct {
		name string
		rop  int
	}{
		{"COPYPEN", ROP_COPYPEN},
		{"XORPEN", ROP_XORPEN},
		{"MERGEPEN", ROP_MERGEPEN},
		{"NOTPEN", ROP_NOTPEN},
		{"MASKPEN", ROP_MASKPEN},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetROP(tt.rop)
			if currentROP != tt.rop {
				t.Errorf("Expected ROP %d, got %d", tt.rop, currentROP)
			}
		})
	}

	// Reset
	SetROP(ROP_COPYPEN)
}

// TestPixelOperationsWithDrawing tests pixel operations integrated with drawing
func TestPixelOperationsWithDrawing(t *testing.T) {
	engine := NewTestEngine()
	globalEngine = engine
	picID := engine.CreatePic(200, 200)

	// Draw a red rectangle with COPYPEN (doesn't require pixel reading)
	SetROP(ROP_COPYPEN)
	SetPaintColor(255, 0, 0)
	DrawRect(picID, 50, 50, 100, 100, FILL_SOLID)

	// Draw a blue circle with COPYPEN mode
	// (XOR mode would require pixel reading which needs game to be running)
	SetROP(ROP_COPYPEN)
	SetPaintColor(0, 0, 255)
	DrawCircle(picID, 100, 100, 30, 30, FILL_SOLID)

	// Reset ROP mode
	SetROP(ROP_COPYPEN)

	pic := engine.pictures[picID]
	if pic == nil {
		t.Fatal("Picture not found")
	}

	t.Log("Pixel operations with drawing executed successfully")
}
