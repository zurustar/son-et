package engine

import (
	"testing"

	"github.com/zurustar/son-et/pkg/compiler/interpreter"
)

// TestSpritePositioningWithScopedVariables tests that sprites (casts) can be
// positioned and moved using variables from parent scope
func TestSpritePositioningWithScopedVariables(t *testing.T) {
	// Test 1: Create cast with position from variables
	t.Run("CreateCastWithVariablePosition", func(t *testing.T) {
		engine := NewTestEngine()

		// Create source and destination pictures
		srcPicID := CreateTestPicture(engine, 100, 100)
		dstPicID := CreateTestPicture(engine, 640, 480)

		// Clear VM
		vmLock.Lock()
		mainSequencer = nil
		vmLock.Unlock()

		// Set position variables
		SetVMVar("spriteX", 50)
		SetVMVar("spriteY", 60)

		// Create cast using variables
		x := Assign("spriteX", 50).(int)
		y := Assign("spriteY", 60).(int)

		castID := engine.PutCast(srcPicID, dstPicID, x, y, 0, 0, 0, 0, 100, 100, 0, 0)

		// Verify cast was created with correct position
		cast := AssertCastExists(t, engine, castID)
		// Note: PutCast creates a transparency-processed picture, so Picture ID will be different
		// We verify position and destination instead
		if cast.X != x {
			t.Errorf("Expected cast X=%d, got %d", x, cast.X)
		}
		if cast.Y != y {
			t.Errorf("Expected cast Y=%d, got %d", y, cast.Y)
		}
		if cast.DestPicture != dstPicID {
			t.Errorf("Expected cast DestPicture=%d, got %d", dstPicID, cast.DestPicture)
		}
	})

	// Test 2: Move cast with calculated position from variables
	t.Run("MoveCastWithCalculatedPosition", func(t *testing.T) {
		engine := NewTestEngine()

		// Create source and destination pictures
		srcPicID := CreateTestPicture(engine, 100, 100)
		dstPicID := CreateTestPicture(engine, 640, 480)

		// Clear VM
		vmLock.Lock()
		mainSequencer = nil
		vmLock.Unlock()

		// Set base variables
		SetVMVar("baseX", 100)
		SetVMVar("baseY", 150)
		SetVMVar("offsetX", 20)
		SetVMVar("offsetY", 30)

		// Create a cast
		castID := engine.PutCast(srcPicID, dstPicID, 0, 0, 0, 0, 0, 0, 100, 100, 0, 0)

		// Calculate new position
		baseX := Assign("baseX", 100).(int)
		baseY := Assign("baseY", 150).(int)
		offsetX := Assign("offsetX", 20).(int)
		offsetY := Assign("offsetY", 30).(int)

		newX := baseX + offsetX // 120
		newY := baseY + offsetY // 180

		// Move cast to calculated position
		engine.MoveCast(castID, dstPicID, newX, newY, 100, 100, 0, 0)

		// Verify cast moved to correct position
		cast := AssertCastExists(t, engine, castID)
		if cast.X != newX {
			t.Errorf("Expected cast X=%d, got %d", newX, cast.X)
		}
		if cast.Y != newY {
			t.Errorf("Expected cast Y=%d, got %d", newY, cast.Y)
		}
	})

	// Test 3: Sprite animation with variable-based positions
	t.Run("SpriteAnimationWithVariables", func(t *testing.T) {
		engine := NewTestEngine()

		// Create source and destination pictures
		srcPicID := CreateTestPicture(engine, 100, 100)
		dstPicID := CreateTestPicture(engine, 640, 480)

		// Clear VM
		vmLock.Lock()
		mainSequencer = nil
		vmLock.Unlock()

		// Create a cast for animation
		castID := engine.PutCast(srcPicID, dstPicID, 0, 0, 0, 0, 0, 0, 100, 100, 0, 0)

		// Simulate animation frames with variable positions
		positions := []struct{ x, y int }{
			{10, 10},
			{20, 15},
			{30, 20},
			{40, 25},
		}

		for i, pos := range positions {
			// Set variables for this frame
			SetVMVar("frameX", pos.x)
			SetVMVar("frameY", pos.y)

			// Move cast to frame position
			x := Assign("frameX", pos.x).(int)
			y := Assign("frameY", pos.y).(int)
			engine.MoveCast(castID, dstPicID, x, y, 100, 100, 0, 0)

			// Verify position
			cast := AssertCastExists(t, engine, castID)
			if cast.X != pos.x {
				t.Errorf("Frame %d: Expected X=%d, got %d", i, pos.x, cast.X)
			}
			if cast.Y != pos.y {
				t.Errorf("Frame %d: Expected Y=%d, got %d", i, pos.y, cast.Y)
			}
		}
	})

	// Test 4: Multiple sprites with relative positioning
	t.Run("MultipleSpriteRelativePositioning", func(t *testing.T) {
		engine := NewTestEngine()

		// Create source and destination pictures
		srcPicID := CreateTestPicture(engine, 100, 100)
		dstPicID := CreateTestPicture(engine, 640, 480)

		// Clear VM
		vmLock.Lock()
		mainSequencer = nil
		vmLock.Unlock()

		// Set base position
		SetVMVar("leaderX", 100)
		SetVMVar("leaderY", 100)

		leaderX := Assign("leaderX", 100).(int)
		leaderY := Assign("leaderY", 100).(int)

		// Create leader sprite
		leaderCast := engine.PutCast(srcPicID, dstPicID, leaderX, leaderY, 0, 0, 0, 0, 100, 100, 0, 0)

		// Create follower sprites at relative positions
		follower1Cast := engine.PutCast(srcPicID, dstPicID, leaderX+50, leaderY, 0, 0, 0, 0, 100, 100, 0, 0)
		follower2Cast := engine.PutCast(srcPicID, dstPicID, leaderX, leaderY+50, 0, 0, 0, 0, 100, 100, 0, 0)

		// Verify positions
		leader := AssertCastExists(t, engine, leaderCast)
		follower1 := AssertCastExists(t, engine, follower1Cast)
		follower2 := AssertCastExists(t, engine, follower2Cast)

		if leader.X != 100 || leader.Y != 100 {
			t.Errorf("Leader position incorrect: (%d, %d)", leader.X, leader.Y)
		}
		if follower1.X != 150 || follower1.Y != 100 {
			t.Errorf("Follower1 position incorrect: (%d, %d)", follower1.X, follower1.Y)
		}
		if follower2.X != 100 || follower2.Y != 150 {
			t.Errorf("Follower2 position incorrect: (%d, %d)", follower2.X, follower2.Y)
		}
	})
}

// TestCastOperationsWithVMScope tests cast operations within VM sequences
// using variables from parent scope
func TestCastOperationsWithVMScope(t *testing.T) {
	// Test 1: PutCast in VM sequence with parent scope variables
	t.Run("PutCastInVMSequence", func(t *testing.T) {
		engine := NewTestEngine()

		// Create pictures
		srcPicID := CreateTestPicture(engine, 100, 100)
		dstPicID := CreateTestPicture(engine, 640, 480)

		// Clear VM
		vmLock.Lock()
		mainSequencer = nil
		vmLock.Unlock()

		// Set variables in parent scope
		SetVMVar("srcPic", srcPicID)
		SetVMVar("dstPic", dstPicID)
		SetVMVar("posX", 75)
		SetVMVar("posY", 85)

		// Register sequence that creates a cast
		ops := []OpCode{
			{Cmd: interpreter.OpPutCast, Args: []any{
				Variable("srcPic"),
				Variable("dstPic"),
				Variable("posX"),
				Variable("posY"),
				0, 0, 0, 0, 100, 100, 0, 0,
			}},
		}

		// Use MIDI_TIME mode to avoid blocking
		RegisterSequence(MidiTime, ops)

		// Verify variables are accessible
		vmLock.Lock()
		if mainSequencer == nil {
			vmLock.Unlock()
			t.Fatal("Sequencer not created")
		}

		srcPic := ResolveArg(Variable("srcPic"), mainSequencer)
		dstPic := ResolveArg(Variable("dstPic"), mainSequencer)
		posX := ResolveArg(Variable("posX"), mainSequencer)
		posY := ResolveArg(Variable("posY"), mainSequencer)
		vmLock.Unlock()

		if srcPic != srcPicID {
			t.Errorf("Expected srcPic=%d, got %v", srcPicID, srcPic)
		}
		if dstPic != dstPicID {
			t.Errorf("Expected dstPic=%d, got %v", dstPicID, dstPic)
		}
		if posX != 75 {
			t.Errorf("Expected posX=75, got %v", posX)
		}
		if posY != 85 {
			t.Errorf("Expected posY=85, got %v", posY)
		}
	})

	// Test 2: MoveCast in VM sequence with calculated expressions
	t.Run("MoveCastInVMSequenceWithCalculations", func(t *testing.T) {
		engine := NewTestEngine()

		// Create pictures
		srcPicID := CreateTestPicture(engine, 100, 100)
		dstPicID := CreateTestPicture(engine, 640, 480)

		// Clear VM
		vmLock.Lock()
		mainSequencer = nil
		vmLock.Unlock()

		// Create a cast first
		castID := engine.PutCast(srcPicID, dstPicID, 0, 0, 0, 0, 0, 0, 100, 100, 0, 0)

		// Set variables for calculation
		SetVMVar("castID", castID)
		SetVMVar("screenW", 640)
		SetVMVar("screenH", 480)
		SetVMVar("spriteW", 100)
		SetVMVar("spriteH", 100)

		// Register sequence that moves cast to center
		ops := []OpCode{
			{Cmd: interpreter.OpMoveCast, Args: []any{
				Variable("castID"),
				dstPicID,
				OpCode{Cmd: interpreter.OpInfix, Args: []any{"/", OpCode{Cmd: interpreter.OpInfix, Args: []any{"-", Variable("screenW"), Variable("spriteW")}}, 2}},
				OpCode{Cmd: interpreter.OpInfix, Args: []any{"/", OpCode{Cmd: interpreter.OpInfix, Args: []any{"-", Variable("screenH"), Variable("spriteH")}}, 2}},
				100, 100, 0, 0,
			}},
		}

		// Use MIDI_TIME mode
		RegisterSequence(MidiTime, ops)

		// Verify calculation: (640-100)/2 = 270, (480-100)/2 = 190
		vmLock.Lock()
		if mainSequencer == nil {
			vmLock.Unlock()
			t.Fatal("Sequencer not created")
		}

		// Resolve the complex expression
		xExpr := ops[0].Args[2].(OpCode)
		yExpr := ops[0].Args[3].(OpCode)

		x := ResolveArg(xExpr, mainSequencer)
		y := ResolveArg(yExpr, mainSequencer)
		vmLock.Unlock()

		expectedX := (640 - 100) / 2 // 270
		expectedY := (480 - 100) / 2 // 190

		if x != expectedX {
			t.Errorf("Expected calculated X=%d, got %v", expectedX, x)
		}
		if y != expectedY {
			t.Errorf("Expected calculated Y=%d, got %v", expectedY, y)
		}
	})

	// Test 3: Case-insensitive cast variable lookup
	t.Run("CaseInsensitiveCastVariables", func(t *testing.T) {
		engine := NewTestEngine()

		// Create pictures
		srcPicID := CreateTestPicture(engine, 100, 100)
		dstPicID := CreateTestPicture(engine, 640, 480)

		// Clear VM
		vmLock.Lock()
		mainSequencer = nil
		vmLock.Unlock()

		// Create a cast
		castID := engine.PutCast(srcPicID, dstPicID, 0, 0, 0, 0, 0, 0, 100, 100, 0, 0)

		// Set variables with mixed case
		SetVMVar("CastID", castID)
		SetVMVar("SpriteX", 123)
		SetVMVar("SpriteY", 456)

		// Use MIDI_TIME mode
		RegisterSequence(MidiTime, []OpCode{})

		// Try to access with different cases
		vmLock.Lock()
		if mainSequencer == nil {
			vmLock.Unlock()
			t.Fatal("Sequencer not created")
		}

		val1 := ResolveArg(Variable("castid"), mainSequencer)
		val2 := ResolveArg(Variable("CASTID"), mainSequencer)
		val3 := ResolveArg(Variable("CastID"), mainSequencer)

		x1 := ResolveArg(Variable("spritex"), mainSequencer)
		x2 := ResolveArg(Variable("SPRITEX"), mainSequencer)

		y1 := ResolveArg(Variable("spritey"), mainSequencer)
		y2 := ResolveArg(Variable("SpriteY"), mainSequencer)
		vmLock.Unlock()

		// All should resolve to the same value
		if val1 != castID || val2 != castID || val3 != castID {
			t.Errorf("Case-insensitive castID lookup failed: %v, %v, %v", val1, val2, val3)
		}
		if x1 != 123 || x2 != 123 {
			t.Errorf("Case-insensitive spriteX lookup failed: %v, %v", x1, x2)
		}
		if y1 != 456 || y2 != 456 {
			t.Errorf("Case-insensitive spriteY lookup failed: %v, %v", y1, y2)
		}
	})
}

// TestVisualElementsRendering tests that all visual elements appear correctly
// when using scoped variables
func TestVisualElementsRendering(t *testing.T) {
	// Test 1: Window and sprite coordination
	t.Run("WindowAndSpriteCoordination", func(t *testing.T) {
		engine := NewTestEngine()

		// Clear VM
		vmLock.Lock()
		mainSequencer = nil
		vmLock.Unlock()

		// Set screen dimensions
		SetVMVar("screenW", 640)
		SetVMVar("screenH", 480)

		screenW := Assign("screenW", 640).(int)
		screenH := Assign("screenH", 480).(int)

		// Create a picture for the window
		winPicID := CreateTestPicture(engine, screenW, screenH)

		// Open window at calculated position
		winX := (1280 - screenW) / 2 // Center in 1280x720 desktop
		winY := (720 - screenH) / 2

		winID := engine.OpenWin(winPicID, winX, winY, screenW, screenH, 0, 0, 0xFFFFFF)

		// Verify window was created
		AssertWindowExists(t, engine, winID)

		// Create sprites in the window
		spritePicID := CreateTestPicture(engine, 50, 50)

		// Position sprites relative to window
		sprite1X := 100
		sprite1Y := 100
		sprite2X := 200
		sprite2Y := 150

		cast1 := engine.PutCast(spritePicID, winPicID, sprite1X, sprite1Y, 0, 0, 0, 0, 50, 50, 0, 0)
		cast2 := engine.PutCast(spritePicID, winPicID, sprite2X, sprite2Y, 0, 0, 0, 0, 50, 50, 0, 0)

		// Verify casts were created
		AssertCastExists(t, engine, cast1)
		AssertCastExists(t, engine, cast2)

		// Verify resource counts
		// Note: PutCast creates additional processed pictures for transparency
		AssertResourceCount(t, engine, 4, 1, 2) // 2 original + 2 processed, 1 window, 2 casts
	})

	// Test 2: Z-ordering with multiple sprites
	t.Run("SpriteZOrdering", func(t *testing.T) {
		engine := NewTestEngine() // Fresh engine for this test

		// Create pictures
		bgPicID := CreateTestPicture(engine, 640, 480)
		spritePicID := CreateTestPicture(engine, 50, 50)

		// Create sprites in specific order
		// First created should be bottom layer
		cast1 := engine.PutCast(spritePicID, bgPicID, 100, 100, 0, 0, 0, 0, 50, 50, 0, 0)
		cast2 := engine.PutCast(spritePicID, bgPicID, 110, 110, 0, 0, 0, 0, 50, 50, 0, 0)
		cast3 := engine.PutCast(spritePicID, bgPicID, 120, 120, 0, 0, 0, 0, 50, 50, 0, 0)

		// Verify draw order matches creation order
		if len(engine.castDrawOrder) != 3 {
			t.Fatalf("Expected 3 casts in draw order, got %d", len(engine.castDrawOrder))
		}

		if engine.castDrawOrder[0] != cast1 {
			t.Errorf("Expected first cast in draw order to be %d, got %d", cast1, engine.castDrawOrder[0])
		}
		if engine.castDrawOrder[1] != cast2 {
			t.Errorf("Expected second cast in draw order to be %d, got %d", cast2, engine.castDrawOrder[1])
		}
		if engine.castDrawOrder[2] != cast3 {
			t.Errorf("Expected third cast in draw order to be %d, got %d", cast3, engine.castDrawOrder[2])
		}
	})

	// Test 3: Sprite clipping with variables
	t.Run("SpriteClippingWithVariables", func(t *testing.T) {
		engine := NewTestEngine()

		// Clear VM
		vmLock.Lock()
		mainSequencer = nil
		vmLock.Unlock()

		// Create a sprite sheet
		sheetPicID := CreateTestPicture(engine, 200, 200)
		dstPicID := CreateTestPicture(engine, 640, 480)

		// Set clipping variables
		SetVMVar("srcX", 50)
		SetVMVar("srcY", 50)
		SetVMVar("clipW", 50)
		SetVMVar("clipH", 50)

		srcX := Assign("srcX", 50).(int)
		srcY := Assign("srcY", 50).(int)
		clipW := Assign("clipW", 50).(int)
		clipH := Assign("clipH", 50).(int)

		// Create cast with clipping
		castID := engine.PutCast(sheetPicID, dstPicID, 100, 100, 0, 0, 0, 0, clipW, clipH, srcX, srcY)

		// Verify cast has correct clipping parameters
		cast := AssertCastExists(t, engine, castID)
		if cast.SrcX != srcX {
			t.Errorf("Expected SrcX=%d, got %d", srcX, cast.SrcX)
		}
		if cast.SrcY != srcY {
			t.Errorf("Expected SrcY=%d, got %d", srcY, cast.SrcY)
		}
		if cast.W != clipW {
			t.Errorf("Expected W=%d, got %d", clipW, cast.W)
		}
		if cast.H != clipH {
			t.Errorf("Expected H=%d, got %d", clipH, cast.H)
		}
	})
}
