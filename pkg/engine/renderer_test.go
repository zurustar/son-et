package engine

import (
	"image"
	"image/color"
	"testing"
)

func TestMockRenderer_RenderFrame(t *testing.T) {
	renderer := &MockRenderer{}
	state := NewEngineState(nil, nil, nil)

	// Create some windows
	state.OpenWindow(1, 100, 100, 200, 150, 0, 0, 0)
	state.OpenWindow(2, 200, 200, 300, 250, 0, 0, 0)

	// Create a dummy screen
	screen := image.NewRGBA(image.Rect(0, 0, 1280, 720))

	// Render
	renderer.RenderFrame(screen, state)

	// Verify render was called
	if renderer.RenderCount != 1 {
		t.Errorf("Expected RenderCount 1, got %d", renderer.RenderCount)
	}

	// Verify windows were recorded
	if len(renderer.RenderedWindows) != 2 {
		t.Errorf("Expected 2 rendered windows, got %d", len(renderer.RenderedWindows))
	}
}

func TestMockRenderer_Clear(t *testing.T) {
	renderer := &MockRenderer{}

	// Clear with color
	renderer.Clear(0xFF0000) // Red

	// Verify clear was called
	if renderer.ClearCount != 1 {
		t.Errorf("Expected ClearCount 1, got %d", renderer.ClearCount)
	}

	// Verify color was recorded
	if renderer.LastColor != 0xFF0000 {
		t.Errorf("Expected LastColor 0xFF0000, got 0x%X", renderer.LastColor)
	}
}

func TestMockRenderer_RendersOnlyVisibleWindows(t *testing.T) {
	renderer := &MockRenderer{}
	state := NewEngineState(nil, nil, nil)

	// Create windows
	win1 := state.OpenWindow(1, 100, 100, 200, 150, 0, 0, 0)
	win2 := state.OpenWindow(2, 200, 200, 300, 250, 0, 0, 0)

	// Hide window 2
	state.GetWindow(win2).Visible = false

	// Create a dummy screen
	screen := image.NewRGBA(image.Rect(0, 0, 1280, 720))

	// Render
	renderer.RenderFrame(screen, state)

	// Verify only visible window was recorded
	if len(renderer.RenderedWindows) != 1 {
		t.Errorf("Expected 1 rendered window, got %d", len(renderer.RenderedWindows))
	}

	if renderer.RenderedWindows[0] != win1 {
		t.Errorf("Expected window %d to be rendered, got %d", win1, renderer.RenderedWindows[0])
	}
}

func TestMockRenderer_RendersCasts(t *testing.T) {
	renderer := &MockRenderer{}
	state := NewEngineState(nil, nil, nil)

	// Create window
	state.OpenWindow(1, 100, 100, 200, 150, 0, 0, 0)

	// Create casts
	state.PutCast(1, 2, 10, 10, 0, 0, 64, 64)
	state.PutCast(1, 3, 20, 20, 0, 0, 64, 64)

	// Create a dummy screen
	screen := image.NewRGBA(image.Rect(0, 0, 1280, 720))

	// Render
	renderer.RenderFrame(screen, state)

	// Verify casts were recorded
	if len(renderer.RenderedCasts) != 2 {
		t.Errorf("Expected 2 rendered casts, got %d", len(renderer.RenderedCasts))
	}
}

func TestMockRenderer_RendersOnlyVisibleCasts(t *testing.T) {
	renderer := &MockRenderer{}
	state := NewEngineState(nil, nil, nil)

	// Create window
	state.OpenWindow(1, 100, 100, 200, 150, 0, 0, 0)

	// Create casts
	cast1 := state.PutCast(1, 2, 10, 10, 0, 0, 64, 64)
	cast2 := state.PutCast(1, 3, 20, 20, 0, 0, 64, 64)

	// Hide cast 2
	state.GetCast(cast2).Visible = false

	// Create a dummy screen
	screen := image.NewRGBA(image.Rect(0, 0, 1280, 720))

	// Render
	renderer.RenderFrame(screen, state)

	// Verify only visible cast was recorded
	if len(renderer.RenderedCasts) != 1 {
		t.Errorf("Expected 1 rendered cast, got %d", len(renderer.RenderedCasts))
	}

	if renderer.RenderedCasts[0] != cast1 {
		t.Errorf("Expected cast %d to be rendered, got %d", cast1, renderer.RenderedCasts[0])
	}
}

func TestMockRenderer_Reset(t *testing.T) {
	renderer := &MockRenderer{}
	state := NewEngineState(nil, nil, nil)

	// Create window
	state.OpenWindow(1, 100, 100, 200, 150, 0, 0, 0)

	// Create a dummy screen
	screen := image.NewRGBA(image.Rect(0, 0, 1280, 720))

	// Render and clear
	renderer.RenderFrame(screen, state)
	renderer.Clear(0xFF0000)

	// Verify counts
	if renderer.RenderCount != 1 {
		t.Errorf("Expected RenderCount 1, got %d", renderer.RenderCount)
	}
	if renderer.ClearCount != 1 {
		t.Errorf("Expected ClearCount 1, got %d", renderer.ClearCount)
	}

	// Reset
	renderer.Reset()

	// Verify reset
	if renderer.RenderCount != 0 {
		t.Errorf("Expected RenderCount 0 after reset, got %d", renderer.RenderCount)
	}
	if renderer.ClearCount != 0 {
		t.Errorf("Expected ClearCount 0 after reset, got %d", renderer.ClearCount)
	}
	if renderer.LastColor != 0 {
		t.Errorf("Expected LastColor 0 after reset, got 0x%X", renderer.LastColor)
	}
	if len(renderer.RenderedWindows) != 0 {
		t.Errorf("Expected 0 rendered windows after reset, got %d", len(renderer.RenderedWindows))
	}
	if len(renderer.RenderedCasts) != 0 {
		t.Errorf("Expected 0 rendered casts after reset, got %d", len(renderer.RenderedCasts))
	}
}

func TestMockRenderer_ThreadSafety(t *testing.T) {
	renderer := &MockRenderer{}
	state := NewEngineState(nil, nil, nil)

	// Create window
	state.OpenWindow(1, 100, 100, 200, 150, 0, 0, 0)

	// Create a dummy screen
	screen := image.NewRGBA(image.Rect(0, 0, 1280, 720))

	// Render from multiple goroutines
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			renderer.RenderFrame(screen, state)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all renders completed
	if renderer.RenderCount != 10 {
		t.Errorf("Expected RenderCount 10, got %d", renderer.RenderCount)
	}
}

func TestEbitenRenderer_Clear(t *testing.T) {
	renderer := NewEbitenRenderer()

	// Clear with red
	renderer.Clear(0xFF0000)

	// Verify background color was set
	expected := uint8(0xFF)
	if renderer.backgroundColor.(color.RGBA).R != expected {
		t.Errorf("Expected R %d, got %d", expected, renderer.backgroundColor.(color.RGBA).R)
	}
	if renderer.backgroundColor.(color.RGBA).G != 0 {
		t.Errorf("Expected G 0, got %d", renderer.backgroundColor.(color.RGBA).G)
	}
	if renderer.backgroundColor.(color.RGBA).B != 0 {
		t.Errorf("Expected B 0, got %d", renderer.backgroundColor.(color.RGBA).B)
	}
}
