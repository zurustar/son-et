package engine

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// TestMockRenderer verifies that MockRenderer can be used for headless testing
func TestMockRenderer(t *testing.T) {
	// Create an EngineState with a mock renderer
	mockRenderer := NewMockRenderer()
	state := NewEngineState(WithRenderer(mockRenderer))

	// Verify initial state
	if mockRenderer.RenderCount != 0 {
		t.Errorf("Expected RenderCount to be 0, got %d", mockRenderer.RenderCount)
	}

	// Create a dummy screen (nil is fine for mock renderer)
	var screen *ebiten.Image

	// Call RenderFrame
	mockRenderer.RenderFrame(screen, state)

	// Verify render was called
	if mockRenderer.RenderCount != 1 {
		t.Errorf("Expected RenderCount to be 1, got %d", mockRenderer.RenderCount)
	}

	// Verify state was captured
	if mockRenderer.LastState != state {
		t.Error("Expected LastState to match the provided state")
	}
}

// TestRendererSeparation verifies that rendering logic is separated from state management
func TestRendererSeparation(t *testing.T) {
	// Create an EngineState with a mock renderer
	mockRenderer := NewMockRenderer()
	state := NewEngineState(WithRenderer(mockRenderer))

	// Modify state (this should not trigger rendering)
	state.renderMutex.Lock()
	state.pictures[0] = &Picture{
		ID:     0,
		Image:  ebiten.NewImage(100, 100),
		Width:  100,
		Height: 100,
	}
	state.renderMutex.Unlock()

	// Verify no rendering occurred
	if mockRenderer.RenderCount != 0 {
		t.Errorf("Expected RenderCount to be 0 after state modification, got %d", mockRenderer.RenderCount)
	}

	// Now explicitly render
	mockRenderer.RenderFrame(nil, state)

	// Verify rendering occurred
	if mockRenderer.RenderCount != 1 {
		t.Errorf("Expected RenderCount to be 1 after explicit render, got %d", mockRenderer.RenderCount)
	}
}

// TestHeadlessTesting verifies that we can test without Ebitengine initialization
func TestHeadlessTesting(t *testing.T) {
	// This test runs without calling ebiten.RunGame or initializing Ebitengine
	// This demonstrates headless testing capability

	mockRenderer := NewMockRenderer()
	state := NewEngineState(WithRenderer(mockRenderer))

	// Perform state operations
	state.renderMutex.Lock()
	pic := &Picture{
		ID:     0,
		Image:  ebiten.NewImage(640, 480),
		Width:  640,
		Height: 480,
	}
	state.pictures[0] = pic
	state.renderMutex.Unlock()

	// Verify state
	if len(state.pictures) != 1 {
		t.Errorf("Expected 1 picture, got %d", len(state.pictures))
	}

	// Simulate rendering without actually rendering to screen
	mockRenderer.RenderFrame(nil, state)

	// Verify mock captured the state
	if mockRenderer.LastState == nil {
		t.Error("Expected LastState to be set")
	}

	if len(mockRenderer.LastState.pictures) != 1 {
		t.Errorf("Expected LastState to have 1 picture, got %d", len(mockRenderer.LastState.pictures))
	}
}

// TestEbitenRendererCreation verifies that EbitenRenderer can be created
func TestEbitenRendererCreation(t *testing.T) {
	renderer := NewEbitenRenderer()
	if renderer == nil {
		t.Error("Expected NewEbitenRenderer to return non-nil renderer")
	}

	if renderer.tickCount != 0 {
		t.Errorf("Expected initial tickCount to be 0, got %d", renderer.tickCount)
	}
}

// TestRendererInterface verifies that both renderers implement the Renderer interface
func TestRendererInterface(t *testing.T) {
	var _ Renderer = NewEbitenRenderer()
	var _ Renderer = NewMockRenderer()
}
