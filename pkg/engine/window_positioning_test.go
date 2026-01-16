package engine

import (
	"testing"
)

// TestWindowPositioningWithCalculatedValues tests that windows can be positioned
// using calculated values from variables (e.g., winW-320, winH/2)
func TestWindowPositioningWithCalculatedValues(t *testing.T) {
	// Note: WinInfo returns 640x480 (legacy window size), not 1280x720 (virtual desktop)

	// Test 1: Center positioning using calculated values
	t.Run("CenterPositioning", func(t *testing.T) {
		winW := WinInfo(0) // Returns 640
		winH := WinInfo(1) // Returns 480

		// Calculate center position (like in Y-SARU sample)
		centerX := winW / 2 // 320
		centerY := winH / 2 // 240

		// Verify calculations
		if centerX != 320 {
			t.Errorf("Expected centerX=320, got %d", centerX)
		}
		if centerY != 240 {
			t.Errorf("Expected centerY=240, got %d", centerY)
		}

		// Test that calculated positions work correctly
		// Position a 200x150 window at center
		x := centerX - 100
		y := centerY - 75

		if x != 220 {
			t.Errorf("Expected x=220, got %d", x)
		}
		if y != 165 {
			t.Errorf("Expected y=165, got %d", y)
		}
	})

	// Test 2: Offset positioning using calculated values
	t.Run("OffsetPositioning", func(t *testing.T) {
		winW := WinInfo(0) // 640
		winH := WinInfo(1) // 480

		picW := 300
		picH := 200

		// Calculate offset position (like in Y-SARU sample)
		// winX = 0 - (winW - picW) / 2
		// winY = 0 - (winH - picH) / 2
		winX := 0 - ((winW - picW) / 2)
		winY := 0 - ((winH - picH) / 2)

		// Verify calculations
		expectedWinX := 0 - ((640 - 300) / 2) // 0 - 170 = -170
		expectedWinY := 0 - ((480 - 200) / 2) // 0 - 140 = -140

		if winX != expectedWinX {
			t.Errorf("Expected winX=%d, got %d", expectedWinX, winX)
		}
		if winY != expectedWinY {
			t.Errorf("Expected winY=%d, got %d", expectedWinY, winY)
		}
	})

	// Test 3: Edge positioning (bottom-right corner)
	t.Run("EdgePositioning", func(t *testing.T) {
		winW := WinInfo(0) // 640
		winH := WinInfo(1) // 480

		windowWidth := 200
		windowHeight := 150

		// Position at bottom-right corner
		x := winW - windowWidth
		y := winH - windowHeight

		// Verify calculations
		if x != 440 { // 640 - 200
			t.Errorf("Expected x=440, got %d", x)
		}
		if y != 330 { // 480 - 150
			t.Errorf("Expected y=330, got %d", y)
		}
	})

	// Test 4: Complex calculated expression
	t.Run("ComplexCalculation", func(t *testing.T) {
		winW := WinInfo(0) // 640
		winH := WinInfo(1) // 480

		// Test expression like: winW - 320
		x := winW - 320
		if x != 320 {
			t.Errorf("Expected x=320, got %d", x)
		}

		// Test expression like: winH / 2 - 100
		y := winH/2 - 100
		if y != 140 {
			t.Errorf("Expected y=140, got %d", y)
		}

		// Test expression like: (winW - 100) / 2
		z := (winW - 100) / 2
		if z != 270 {
			t.Errorf("Expected z=270, got %d", z)
		}
	})
}

// TestVMVariableScope tests that variables defined outside mes() blocks
// are accessible inside them via the VM variable scope chain
func TestVMVariableScope(t *testing.T) {
	// Initialize VM
	vmLock.Lock()
	mainSequencer = nil
	vmLock.Unlock()

	// Test 1: Set variables before RegisterSequence
	t.Run("VariablesAccessibleInMesBlock", func(t *testing.T) {
		// Clear VM first
		vmLock.Lock()
		mainSequencer = nil
		vmLock.Unlock()

		// Simulate variables set before mes() block
		SetVMVar("winW", 640)
		SetVMVar("winH", 480)
		SetVMVar("centerX", 320)
		SetVMVar("centerY", 240)

		// Register a sequence that uses these variables
		ops := []OpCode{
			{Cmd: "OpenWin", Args: []any{
				0,
				Variable("centerX"),
				Variable("centerY"),
				Variable("winW"),
				Variable("winH"),
				0, 0, 0xffffff,
			}},
		}

		// Use MIDI_TIME mode to avoid blocking (Time mode blocks until sequence completes)
		RegisterSequence(MidiTime, ops)

		// Verify variables are accessible in the sequencer
		vmLock.Lock()
		if mainSequencer == nil {
			vmLock.Unlock()
			t.Fatal("Sequencer not created")
		}

		// Check that variables can be resolved
		winW := ResolveArg(Variable("winW"), mainSequencer)
		winH := ResolveArg(Variable("winH"), mainSequencer)
		centerX := ResolveArg(Variable("centerX"), mainSequencer)
		centerY := ResolveArg(Variable("centerY"), mainSequencer)
		vmLock.Unlock()

		if winW != 640 {
			t.Errorf("Expected winW=640, got %v", winW)
		}
		if winH != 480 {
			t.Errorf("Expected winH=480, got %v", winH)
		}
		if centerX != 320 {
			t.Errorf("Expected centerX=320, got %v", centerX)
		}
		if centerY != 240 {
			t.Errorf("Expected centerY=240, got %v", centerY)
		}
	})

	// Test 2: Case-insensitive variable lookup
	t.Run("CaseInsensitiveVariableLookup", func(t *testing.T) {
		// Clear VM
		vmLock.Lock()
		mainSequencer = nil
		vmLock.Unlock()

		// Set variable with mixed case
		SetVMVar("WinW", 640)

		// Use MIDI_TIME mode to avoid blocking
		RegisterSequence(MidiTime, []OpCode{})

		// Try to access with different cases
		vmLock.Lock()
		if mainSequencer == nil {
			vmLock.Unlock()
			t.Fatal("Sequencer not created")
		}

		val1 := ResolveArg(Variable("winw"), mainSequencer)
		val2 := ResolveArg(Variable("WINW"), mainSequencer)
		val3 := ResolveArg(Variable("WinW"), mainSequencer)
		vmLock.Unlock()

		// All should resolve to the same value
		if val1 != 640 || val2 != 640 || val3 != 640 {
			t.Errorf("Case-insensitive lookup failed: winw=%v, WINW=%v, WinW=%v",
				val1, val2, val3)
		}
	})

	// Test 3: Calculated expressions in mes() blocks
	t.Run("CalculatedExpressionsInMesBlock", func(t *testing.T) {
		// Clear VM
		vmLock.Lock()
		mainSequencer = nil
		vmLock.Unlock()

		// Set base variables
		SetVMVar("winW", 640)
		SetVMVar("winH", 480)

		// Register sequence with calculated expressions
		ops := []OpCode{
			{Cmd: "OpenWin", Args: []any{
				0,
				OpCode{Cmd: "Infix", Args: []any{"-", Variable("winW"), 320}},
				OpCode{Cmd: "Infix", Args: []any{"-", Variable("winH"), 240}},
				640, 480, 0, 0, 0xffffff,
			}},
		}

		// Use MIDI_TIME mode to avoid blocking
		RegisterSequence(MidiTime, ops)

		// Execute the OpCode to verify calculation works
		vmLock.Lock()
		if mainSequencer == nil {
			vmLock.Unlock()
			t.Fatal("Sequencer not created")
		}

		// Resolve the calculated expressions
		xExpr := ops[0].Args[1].(OpCode)
		yExpr := ops[0].Args[2].(OpCode)

		x := ResolveArg(xExpr, mainSequencer)
		y := ResolveArg(yExpr, mainSequencer)
		vmLock.Unlock()

		expectedX := 640 - 320 // 320
		expectedY := 480 - 240 // 240

		if x != expectedX {
			t.Errorf("Expected calculated X=%d, got %v", expectedX, x)
		}
		if y != expectedY {
			t.Errorf("Expected calculated Y=%d, got %v", expectedY, y)
		}
	})
}
