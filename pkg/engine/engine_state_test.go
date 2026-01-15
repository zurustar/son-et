package engine

import (
	"image/color"
	"reflect"
	"sync/atomic"
	"testing"
)

// TestNewEngineState verifies that NewEngineState creates a properly initialized state
func TestNewEngineState(t *testing.T) {
	state := NewEngineState()

	// Verify maps are initialized
	if state.pictures == nil {
		t.Error("pictures map not initialized")
	}
	if state.windows == nil {
		t.Error("windows map not initialized")
	}
	if state.casts == nil {
		t.Error("casts map not initialized")
	}
	if state.userFuncs == nil {
		t.Error("userFuncs map not initialized")
	}

	// Verify slices are initialized
	if state.castDrawOrder == nil {
		t.Error("castDrawOrder slice not initialized")
	}
	if state.windowOrder == nil {
		t.Error("windowOrder slice not initialized")
	}

	// Verify ID counters start at correct values
	if state.nextPicID != 0 {
		t.Errorf("nextPicID should start at 0, got %d", state.nextPicID)
	}
	if state.nextWinID != 0 {
		t.Errorf("nextWinID should start at 0, got %d", state.nextWinID)
	}
	if state.nextCastID != 1 {
		t.Errorf("nextCastID should start at 1, got %d", state.nextCastID)
	}

	// Verify text rendering defaults
	if state.currentFontSize != 14 {
		t.Errorf("currentFontSize should be 14, got %d", state.currentFontSize)
	}
	if state.currentFontName != "sans-serif" {
		t.Errorf("currentFontName should be 'sans-serif', got %s", state.currentFontName)
	}
	expectedTextColor := color.RGBA{0, 0, 0, 255}
	if state.currentTextColor != expectedTextColor {
		t.Errorf("currentTextColor should be black, got %v", state.currentTextColor)
	}
	expectedBgColor := color.RGBA{255, 255, 255, 255}
	if state.currentBgColor != expectedBgColor {
		t.Errorf("currentBgColor should be white, got %v", state.currentBgColor)
	}
	if state.currentBackMode != 0 {
		t.Errorf("currentBackMode should be 0, got %d", state.currentBackMode)
	}

	// Verify window decoration defaults
	if state.defaultWindowTitle != "FILLY Window" {
		t.Errorf("defaultWindowTitle should be 'FILLY Window', got %s", state.defaultWindowTitle)
	}
	if state.globalWindowTitle != "FILLY Window" {
		t.Errorf("globalWindowTitle should be 'FILLY Window', got %s", state.globalWindowTitle)
	}

	// Verify VM defaults
	if state.ticksPerStep != 12 {
		t.Errorf("ticksPerStep should be 12, got %d", state.ticksPerStep)
	}
	if state.GlobalPPQ != 480 {
		t.Errorf("GlobalPPQ should be 480, got %d", state.GlobalPPQ)
	}
	if state.MidiTime != 1 {
		t.Errorf("MidiTime should be 1, got %d", state.MidiTime)
	}

	// Verify procedural defaults
	if state.procMode != 0 {
		t.Errorf("procMode should be 0, got %d", state.procMode)
	}
	if state.procStep != 6 {
		t.Errorf("procStep should be 6, got %d", state.procStep)
	}
}

// TestEngineStateReset verifies that Reset clears all state
func TestEngineStateReset(t *testing.T) {
	state := NewEngineState()

	// Populate state with some data
	state.pictures[0] = &Picture{ID: 0, Width: 100, Height: 100}
	state.pictures[1] = &Picture{ID: 1, Width: 200, Height: 200}
	state.windows[0] = &Window{ID: 0, Picture: 0}
	state.casts[1] = &Cast{ID: 1, Picture: 0}
	state.castDrawOrder = append(state.castDrawOrder, 1)
	state.windowOrder = append(state.windowOrder, 0)
	state.nextPicID = 5
	state.nextWinID = 3
	state.nextCastID = 10
	state.currentFontSize = 24
	state.globalWindowTitle = "Custom Title"
	state.tickCount = 100
	atomic.StoreInt64(&state.targetTick, 50)
	state.userFuncs["test"] = reflect.ValueOf(func() {})
	state.procMode = 1
	state.procStep = 20

	// Reset state
	state.Reset()

	// Verify all resources are cleared
	if len(state.pictures) != 0 {
		t.Errorf("pictures should be empty after reset, got %d items", len(state.pictures))
	}
	if len(state.windows) != 0 {
		t.Errorf("windows should be empty after reset, got %d items", len(state.windows))
	}
	if len(state.casts) != 0 {
		t.Errorf("casts should be empty after reset, got %d items", len(state.casts))
	}
	if len(state.castDrawOrder) != 0 {
		t.Errorf("castDrawOrder should be empty after reset, got %d items", len(state.castDrawOrder))
	}
	if len(state.windowOrder) != 0 {
		t.Errorf("windowOrder should be empty after reset, got %d items", len(state.windowOrder))
	}
	if len(state.userFuncs) != 0 {
		t.Errorf("userFuncs should be empty after reset, got %d items", len(state.userFuncs))
	}

	// Verify ID counters are reset
	if state.nextPicID != 0 {
		t.Errorf("nextPicID should be reset to 0, got %d", state.nextPicID)
	}
	if state.nextWinID != 0 {
		t.Errorf("nextWinID should be reset to 0, got %d", state.nextWinID)
	}
	if state.nextCastID != 1 {
		t.Errorf("nextCastID should be reset to 1, got %d", state.nextCastID)
	}

	// Verify text rendering state is reset
	if state.currentFontSize != 14 {
		t.Errorf("currentFontSize should be reset to 14, got %d", state.currentFontSize)
	}
	if state.currentFont != nil {
		t.Error("currentFont should be reset to nil")
	}

	// Verify window decoration is reset
	if state.globalWindowTitle != state.defaultWindowTitle {
		t.Errorf("globalWindowTitle should be reset to defaultWindowTitle, got %s", state.globalWindowTitle)
	}

	// Verify VM state is reset
	if state.mainSequencer != nil {
		t.Error("mainSequencer should be reset to nil")
	}
	if state.tickCount != 0 {
		t.Errorf("tickCount should be reset to 0, got %d", state.tickCount)
	}
	if state.ticksPerStep != 12 {
		t.Errorf("ticksPerStep should be reset to 12, got %d", state.ticksPerStep)
	}
	if state.midiSyncMode != false {
		t.Error("midiSyncMode should be reset to false")
	}
	if atomic.LoadInt64(&state.targetTick) != 0 {
		t.Errorf("targetTick should be reset to 0, got %d", atomic.LoadInt64(&state.targetTick))
	}

	// Verify procedural state is reset
	if state.procMode != 0 {
		t.Errorf("procMode should be reset to 0, got %d", state.procMode)
	}
	if state.procStep != 6 {
		t.Errorf("procStep should be reset to 6, got %d", state.procStep)
	}
	if state.queuedCallback != nil {
		t.Error("queuedCallback should be reset to nil")
	}
}

// TestEngineStateIsolation verifies that multiple EngineState instances are isolated
func TestEngineStateIsolation(t *testing.T) {
	state1 := NewEngineState()
	state2 := NewEngineState()

	// Modify state1
	state1.pictures[0] = &Picture{ID: 0, Width: 100, Height: 100}
	state1.nextPicID = 5
	state1.globalWindowTitle = "State 1 Title"

	// Verify state2 is unaffected
	if len(state2.pictures) != 0 {
		t.Error("state2 pictures should be empty")
	}
	if state2.nextPicID != 0 {
		t.Errorf("state2 nextPicID should be 0, got %d", state2.nextPicID)
	}
	if state2.globalWindowTitle != "FILLY Window" {
		t.Errorf("state2 globalWindowTitle should be default, got %s", state2.globalWindowTitle)
	}

	// Verify state1 has the modifications
	if len(state1.pictures) != 1 {
		t.Errorf("state1 should have 1 picture, got %d", len(state1.pictures))
	}
	if state1.nextPicID != 5 {
		t.Errorf("state1 nextPicID should be 5, got %d", state1.nextPicID)
	}
	if state1.globalWindowTitle != "State 1 Title" {
		t.Errorf("state1 globalWindowTitle should be 'State 1 Title', got %s", state1.globalWindowTitle)
	}
}
