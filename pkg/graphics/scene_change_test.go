package graphics

import (
	"image"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// TestSceneChangeModeConstants tests that scene change mode constants are defined correctly
func TestSceneChangeModeConstants(t *testing.T) {
	tests := []struct {
		mode     SceneChangeMode
		expected int
	}{
		{SceneChangeNone, 0},
		{SceneChangeTransparent, 1},
		{SceneChangeWipeDown, 2},
		{SceneChangeWipeRight, 3},
		{SceneChangeWipeLeft, 4},
		{SceneChangeWipeUp, 5},
		{SceneChangeWipeOut, 6},
		{SceneChangeWipeIn, 7},
		{SceneChangeRandom, 8},
		{SceneChangeFade, 9},
	}

	for _, tt := range tests {
		if int(tt.mode) != tt.expected {
			t.Errorf("SceneChangeMode %d expected %d, got %d", tt.mode, tt.expected, int(tt.mode))
		}
	}
}

// TestNewSceneChange tests SceneChange creation
func TestNewSceneChange(t *testing.T) {
	srcImg := ebiten.NewImage(100, 100)
	dstImg := ebiten.NewImage(100, 100)
	srcRect := image.Rect(0, 0, 50, 50)
	dstPoint := image.Point{X: 10, Y: 10}

	tests := []struct {
		name          string
		mode          SceneChangeMode
		speed         int
		expectedSpeed int
	}{
		{"wipe down default speed", SceneChangeWipeDown, 0, DefaultSceneChangeSpeed},
		{"wipe right custom speed", SceneChangeWipeRight, 10, 10},
		{"fade max speed", SceneChangeFade, 150, 100}, // Should be capped at 100
		{"random negative speed", SceneChangeRandom, -5, DefaultSceneChangeSpeed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := NewSceneChange(srcImg, dstImg, srcRect, dstPoint, tt.mode, tt.speed)

			if sc.mode != tt.mode {
				t.Errorf("mode = %v, want %v", sc.mode, tt.mode)
			}
			if sc.speed != tt.expectedSpeed {
				t.Errorf("speed = %v, want %v", sc.speed, tt.expectedSpeed)
			}
			if sc.progress != 0.0 {
				t.Errorf("initial progress = %v, want 0.0", sc.progress)
			}
			if sc.completed {
				t.Error("initial completed = true, want false")
			}
		})
	}
}

// TestSceneChangeUpdate tests the Update method
func TestSceneChangeUpdate(t *testing.T) {
	srcImg := ebiten.NewImage(100, 100)
	dstImg := ebiten.NewImage(100, 100)
	srcRect := image.Rect(0, 0, 50, 50)
	dstPoint := image.Point{X: 0, Y: 0}

	// Test with speed=100 (should complete in 1 update)
	sc := NewSceneChange(srcImg, dstImg, srcRect, dstPoint, SceneChangeWipeDown, 100)

	completed := sc.Update()
	if !completed {
		t.Error("Update with speed=100 should complete in 1 frame")
	}
	if sc.GetProgress() != 1.0 {
		t.Errorf("progress after completion = %v, want 1.0", sc.GetProgress())
	}

	// Test with speed=50 (should take 2 updates)
	sc2 := NewSceneChange(srcImg, dstImg, srcRect, dstPoint, SceneChangeWipeRight, 50)

	completed = sc2.Update()
	if completed {
		t.Error("Update with speed=50 should not complete in 1 frame")
	}
	if sc2.GetProgress() != 0.5 {
		t.Errorf("progress after 1 update = %v, want 0.5", sc2.GetProgress())
	}

	completed = sc2.Update()
	if !completed {
		t.Error("Update with speed=50 should complete in 2 frames")
	}
}

// TestSceneChangeApply tests the Apply method for each mode
func TestSceneChangeApply(t *testing.T) {
	modes := []SceneChangeMode{
		SceneChangeWipeDown,
		SceneChangeWipeRight,
		SceneChangeWipeLeft,
		SceneChangeWipeUp,
		SceneChangeWipeOut,
		SceneChangeWipeIn,
		SceneChangeRandom,
		SceneChangeFade,
	}

	for _, mode := range modes {
		t.Run(mode.String(), func(t *testing.T) {
			srcImg := ebiten.NewImage(100, 100)
			dstImg := ebiten.NewImage(100, 100)
			srcRect := image.Rect(0, 0, 50, 50)
			dstPoint := image.Point{X: 0, Y: 0}

			sc := NewSceneChange(srcImg, dstImg, srcRect, dstPoint, mode, 50)

			// Apply at 50% progress
			sc.Update()
			sc.Apply()

			// Should not panic and should not be completed yet
			if sc.IsCompleted() {
				t.Error("should not be completed at 50% progress")
			}

			// Complete the scene change
			sc.Update()
			sc.Apply()

			if !sc.IsCompleted() {
				t.Error("should be completed after 100% progress")
			}
		})
	}
}

// TestSceneChangeComplete tests the Complete method
func TestSceneChangeComplete(t *testing.T) {
	srcImg := ebiten.NewImage(100, 100)
	dstImg := ebiten.NewImage(100, 100)
	srcRect := image.Rect(0, 0, 50, 50)
	dstPoint := image.Point{X: 0, Y: 0}

	sc := NewSceneChange(srcImg, dstImg, srcRect, dstPoint, SceneChangeWipeDown, 1)

	// Force complete
	sc.Complete()

	if !sc.IsCompleted() {
		t.Error("IsCompleted should return true after Complete()")
	}
	if sc.GetProgress() != 1.0 {
		t.Errorf("progress after Complete() = %v, want 1.0", sc.GetProgress())
	}
}

// TestSceneChangeManager tests the SceneChangeManager
func TestSceneChangeManager(t *testing.T) {
	scm := NewSceneChangeManager()

	if scm.HasActiveChanges() {
		t.Error("new manager should have no active changes")
	}
	if scm.GetActiveCount() != 0 {
		t.Errorf("new manager active count = %d, want 0", scm.GetActiveCount())
	}

	// Add a scene change
	srcImg := ebiten.NewImage(100, 100)
	dstImg := ebiten.NewImage(100, 100)
	srcRect := image.Rect(0, 0, 50, 50)
	dstPoint := image.Point{X: 0, Y: 0}

	sc := NewSceneChange(srcImg, dstImg, srcRect, dstPoint, SceneChangeWipeDown, 100)
	scm.Add(sc)

	if !scm.HasActiveChanges() {
		t.Error("manager should have active changes after Add")
	}
	if scm.GetActiveCount() != 1 {
		t.Errorf("active count = %d, want 1", scm.GetActiveCount())
	}

	// Update should complete and remove the scene change
	scm.Update()

	if scm.HasActiveChanges() {
		t.Error("manager should have no active changes after completion")
	}
	if scm.GetActiveCount() != 0 {
		t.Errorf("active count after completion = %d, want 0", scm.GetActiveCount())
	}
}

// TestSceneChangeManagerMultiple tests managing multiple scene changes
func TestSceneChangeManagerMultiple(t *testing.T) {
	scm := NewSceneChangeManager()

	srcImg := ebiten.NewImage(100, 100)
	dstImg := ebiten.NewImage(100, 100)
	srcRect := image.Rect(0, 0, 50, 50)
	dstPoint := image.Point{X: 0, Y: 0}

	// Add multiple scene changes with different speeds
	sc1 := NewSceneChange(srcImg, dstImg, srcRect, dstPoint, SceneChangeWipeDown, 100) // Completes in 1 frame
	sc2 := NewSceneChange(srcImg, dstImg, srcRect, dstPoint, SceneChangeWipeRight, 50) // Completes in 2 frames
	sc3 := NewSceneChange(srcImg, dstImg, srcRect, dstPoint, SceneChangeFade, 34)      // Completes in 3 frames

	scm.Add(sc1)
	scm.Add(sc2)
	scm.Add(sc3)

	if scm.GetActiveCount() != 3 {
		t.Errorf("active count = %d, want 3", scm.GetActiveCount())
	}

	// First update: sc1 completes
	scm.Update()
	if scm.GetActiveCount() != 2 {
		t.Errorf("active count after 1st update = %d, want 2", scm.GetActiveCount())
	}

	// Second update: sc2 completes
	scm.Update()
	if scm.GetActiveCount() != 1 {
		t.Errorf("active count after 2nd update = %d, want 1", scm.GetActiveCount())
	}

	// Third update: sc3 completes
	scm.Update()
	if scm.GetActiveCount() != 0 {
		t.Errorf("active count after 3rd update = %d, want 0", scm.GetActiveCount())
	}
}

// TestSceneChangeManagerClear tests the Clear method
func TestSceneChangeManagerClear(t *testing.T) {
	scm := NewSceneChangeManager()

	srcImg := ebiten.NewImage(100, 100)
	dstImg := ebiten.NewImage(100, 100)
	srcRect := image.Rect(0, 0, 50, 50)
	dstPoint := image.Point{X: 0, Y: 0}

	// Add some scene changes
	scm.Add(NewSceneChange(srcImg, dstImg, srcRect, dstPoint, SceneChangeWipeDown, 1))
	scm.Add(NewSceneChange(srcImg, dstImg, srcRect, dstPoint, SceneChangeWipeRight, 1))

	if scm.GetActiveCount() != 2 {
		t.Errorf("active count = %d, want 2", scm.GetActiveCount())
	}

	// Clear all
	scm.Clear()

	if scm.HasActiveChanges() {
		t.Error("manager should have no active changes after Clear")
	}
	if scm.GetActiveCount() != 0 {
		t.Errorf("active count after Clear = %d, want 0", scm.GetActiveCount())
	}
}

// TestMovePicSceneChangeMode tests MovePic with scene change modes (2-9)
func TestMovePicSceneChangeMode(t *testing.T) {
	gs := NewGraphicsSystem("")

	srcID, err := gs.CreatePic(100, 100)
	if err != nil {
		t.Fatalf("Failed to create source picture: %v", err)
	}

	dstID, err := gs.CreatePic(100, 100)
	if err != nil {
		t.Fatalf("Failed to create destination picture: %v", err)
	}

	// Test each scene change mode
	modes := []int{2, 3, 4, 5, 6, 7, 8, 9}
	for _, mode := range modes {
		t.Run("mode_"+string(rune('0'+mode)), func(t *testing.T) {
			err := gs.MovePic(srcID, 0, 0, 50, 50, dstID, 0, 0, mode)
			if err != nil {
				t.Errorf("MovePic with mode %d failed: %v", mode, err)
			}

			// Scene change should be added to the manager
			if !gs.sceneChanges.HasActiveChanges() {
				t.Errorf("Scene change manager should have active changes for mode %d", mode)
			}

			// Clear for next test
			gs.sceneChanges.Clear()
		})
	}
}

// TestMovePicWithSpeed tests MovePicWithSpeed
func TestMovePicWithSpeed(t *testing.T) {
	gs := NewGraphicsSystem("")

	srcID, err := gs.CreatePic(100, 100)
	if err != nil {
		t.Fatalf("Failed to create source picture: %v", err)
	}

	dstID, err := gs.CreatePic(100, 100)
	if err != nil {
		t.Fatalf("Failed to create destination picture: %v", err)
	}

	// Test with custom speed
	err = gs.MovePicWithSpeed(srcID, 0, 0, 50, 50, dstID, 0, 0, 2, 50)
	if err != nil {
		t.Errorf("MovePicWithSpeed failed: %v", err)
	}

	if !gs.sceneChanges.HasActiveChanges() {
		t.Error("Scene change manager should have active changes")
	}
}

// TestApplyImmediate tests the ApplyImmediate function
func TestApplyImmediate(t *testing.T) {
	srcImg := ebiten.NewImage(100, 100)
	dstImg := ebiten.NewImage(100, 100)
	srcRect := image.Rect(0, 0, 50, 50)
	dstPoint := image.Point{X: 0, Y: 0}

	// Test each mode - should not panic
	modes := []SceneChangeMode{
		SceneChangeNone,
		SceneChangeTransparent,
		SceneChangeWipeDown,
		SceneChangeFade,
	}

	for _, mode := range modes {
		t.Run(mode.String(), func(t *testing.T) {
			// Should not panic
			ApplyImmediate(srcImg, dstImg, srcRect, dstPoint, mode, nil)
		})
	}
}

// TestRandomBlocksInitialization tests that random blocks are properly initialized
func TestRandomBlocksInitialization(t *testing.T) {
	srcImg := ebiten.NewImage(100, 100)
	dstImg := ebiten.NewImage(100, 100)
	srcRect := image.Rect(0, 0, 100, 100)
	dstPoint := image.Point{X: 0, Y: 0}

	sc := NewSceneChange(srcImg, dstImg, srcRect, dstPoint, SceneChangeRandom, 50)

	// Check that blocks are initialized
	expectedBlocksX := (100 + RandomBlockSize - 1) / RandomBlockSize
	expectedBlocksY := (100 + RandomBlockSize - 1) / RandomBlockSize
	expectedBlockCount := expectedBlocksX * expectedBlocksY

	if sc.blockCount != expectedBlockCount {
		t.Errorf("blockCount = %d, want %d", sc.blockCount, expectedBlockCount)
	}

	if len(sc.blockOrder) != expectedBlockCount {
		t.Errorf("len(blockOrder) = %d, want %d", len(sc.blockOrder), expectedBlockCount)
	}

	// Verify all block indices are present (shuffled but complete)
	seen := make(map[int]bool)
	for _, idx := range sc.blockOrder {
		if idx < 0 || idx >= expectedBlockCount {
			t.Errorf("invalid block index: %d", idx)
		}
		if seen[idx] {
			t.Errorf("duplicate block index: %d", idx)
		}
		seen[idx] = true
	}
}

// String returns a string representation of SceneChangeMode for test output
func (m SceneChangeMode) String() string {
	switch m {
	case SceneChangeNone:
		return "None"
	case SceneChangeTransparent:
		return "Transparent"
	case SceneChangeWipeDown:
		return "WipeDown"
	case SceneChangeWipeRight:
		return "WipeRight"
	case SceneChangeWipeLeft:
		return "WipeLeft"
	case SceneChangeWipeUp:
		return "WipeUp"
	case SceneChangeWipeOut:
		return "WipeOut"
	case SceneChangeWipeIn:
		return "WipeIn"
	case SceneChangeRandom:
		return "Random"
	case SceneChangeFade:
		return "Fade"
	default:
		return "Unknown"
	}
}
