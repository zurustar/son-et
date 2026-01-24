package engine

import (
	"testing"

	"github.com/zurustar/son-et/pkg/compiler/interpreter"
)

// TestFunctionReturnValueAssignment tests that function return values can be assigned to variables
func TestFunctionReturnValueAssignment(t *testing.T) {
	// Create test VM
	vm, engine, state, _ := newTestVM()

	// Create a sequencer
	seq := NewSequencer([]interpreter.OpCode{}, TIME, nil)

	// Test CreatePic return value
	t.Run("CreatePic return value", func(t *testing.T) {
		// Create OpCode: canvas = CreatePic(400, 300)
		op := interpreter.OpCode{
			Cmd: interpreter.OpAssign,
			Args: []any{
				interpreter.Variable("canvas"),
				interpreter.OpCode{
					Cmd:  interpreter.OpCall,
					Args: []any{"CreatePic", int64(400), int64(300)},
				},
			},
		}

		// Execute assignment
		err := vm.ExecuteOp(seq, op)
		if err != nil {
			t.Fatalf("ExecuteOp failed: %v", err)
		}

		// Verify variable was set to picture ID
		canvasID := seq.GetVariable("canvas")
		if canvasID == nil || canvasID == int64(0) {
			t.Errorf("Expected canvas to be set to picture ID, got %v", canvasID)
		}

		// Verify picture was created with correct dimensions
		canvasIDInt := int(vm.toInt(canvasID))
		pic := state.GetPicture(canvasIDInt)
		if pic == nil {
			t.Fatalf("Picture %d was not created", canvasIDInt)
		}
		if pic.Width != 400 || pic.Height != 300 {
			t.Errorf("Expected picture dimensions 400x300, got %dx%d", pic.Width, pic.Height)
		}
	})

	// Test OpenWin return value
	t.Run("OpenWin return value", func(t *testing.T) {
		// First create a picture
		picID := engine.CreatePic(100, 100)

		// Create OpCode: win = OpenWin(pic)
		op := interpreter.OpCode{
			Cmd: interpreter.OpAssign,
			Args: []any{
				interpreter.Variable("win"),
				interpreter.OpCode{
					Cmd:  interpreter.OpCall,
					Args: []any{"OpenWin", int64(picID)},
				},
			},
		}

		// Execute assignment
		err := vm.ExecuteOp(seq, op)
		if err != nil {
			t.Fatalf("ExecuteOp failed: %v", err)
		}

		// Verify variable was set to window ID
		winID := seq.GetVariable("win")
		if winID == nil || winID == int64(0) {
			t.Errorf("Expected win to be set to window ID, got %v", winID)
		}

		// Verify window was created
		winIDInt := int(vm.toInt(winID))
		if state.GetWindow(winIDInt) == nil {
			t.Errorf("Window %d was not created", winIDInt)
		}
	})

	// Test PutCast return value
	t.Run("PutCast return value", func(t *testing.T) {
		// Create picture and window
		picID := engine.CreatePic(100, 100)
		winID := engine.OpenWin(picID, 0, 0, 100, 100, 0, 0, 0)

		// Create OpCode: cast = PutCast(win, pic, 10, 10, 0, 0, 50, 50)
		op := interpreter.OpCode{
			Cmd: interpreter.OpAssign,
			Args: []any{
				interpreter.Variable("cast"),
				interpreter.OpCode{
					Cmd: interpreter.OpCall,
					Args: []any{
						"PutCast",
						int64(winID),
						int64(picID),
						int64(10),
						int64(10),
						int64(0),
						int64(0),
						int64(50),
						int64(50),
					},
				},
			},
		}

		// Execute assignment
		err := vm.ExecuteOp(seq, op)
		if err != nil {
			t.Fatalf("ExecuteOp failed: %v", err)
		}

		// Verify variable was set to cast ID
		castID := seq.GetVariable("cast")
		if castID == nil || castID == int64(0) {
			t.Errorf("Expected cast to be set to cast ID, got %v", castID)
		}

		// Verify cast was created
		castIDInt := int(vm.toInt(castID))
		if state.GetCast(castIDInt) == nil {
			t.Errorf("Cast %d was not created", castIDInt)
		}
	})
}

// TestFunctionReturnValueInExpression tests that function return values can be used in expressions
func TestFunctionReturnValueInExpression(t *testing.T) {
	// Create test VM
	vm, _, _, _ := newTestVM()

	// Create a sequencer
	seq := NewSequencer([]interpreter.OpCode{}, TIME, nil)

	// Test return value in arithmetic expression
	t.Run("Return value in arithmetic expression", func(t *testing.T) {
		// Create OpCode: x = CreatePic(100, 100) + 5
		op := interpreter.OpCode{
			Cmd: interpreter.OpAssign,
			Args: []any{
				interpreter.Variable("x"),
				interpreter.OpCode{
					Cmd: interpreter.OpBinaryOp,
					Args: []any{
						"+",
						interpreter.OpCode{
							Cmd:  interpreter.OpCall,
							Args: []any{"CreatePic", int64(100), int64(100)},
						},
						int64(5),
					},
				},
			},
		}

		// Execute assignment
		err := vm.ExecuteOp(seq, op)
		if err != nil {
			t.Fatalf("ExecuteOp failed: %v", err)
		}

		// Verify variable was set to picture ID + 5
		x := seq.GetVariable("x")
		if x == nil {
			t.Fatalf("Expected x to be set, got nil")
		}

		// The result should be picID + 5
		// Since we don't know the exact picID, just verify it's greater than 5
		xInt := vm.toInt(x)
		if xInt <= 5 {
			t.Errorf("Expected x > 5, got %d", xInt)
		}
	})

	// Test multiple return values in expression
	t.Run("Multiple return values in expression", func(t *testing.T) {
		// Create OpCode: result = CreatePic(50, 50) + CreatePic(60, 60)
		op := interpreter.OpCode{
			Cmd: interpreter.OpAssign,
			Args: []any{
				interpreter.Variable("result"),
				interpreter.OpCode{
					Cmd: interpreter.OpBinaryOp,
					Args: []any{
						"+",
						interpreter.OpCode{
							Cmd:  interpreter.OpCall,
							Args: []any{"CreatePic", int64(50), int64(50)},
						},
						interpreter.OpCode{
							Cmd:  interpreter.OpCall,
							Args: []any{"CreatePic", int64(60), int64(60)},
						},
					},
				},
			},
		}

		// Execute assignment
		err := vm.ExecuteOp(seq, op)
		if err != nil {
			t.Fatalf("ExecuteOp failed: %v", err)
		}

		// Verify variable was set
		result := seq.GetVariable("result")
		if result == nil {
			t.Fatalf("Expected result to be set, got nil")
		}

		// The result should be sum of two picture IDs
		resultInt := vm.toInt(result)
		if resultInt == 0 {
			t.Errorf("Expected result > 0, got %d", resultInt)
		}
	})
}

// TestReturnValueCleanup tests that return values are cleaned up after use
func TestReturnValueCleanup(t *testing.T) {
	// Create test VM
	vm, _, _, _ := newTestVM()

	// Create a sequencer
	seq := NewSequencer([]interpreter.OpCode{}, TIME, nil)

	// Execute first function call
	op1 := interpreter.OpCode{
		Cmd: interpreter.OpAssign,
		Args: []any{
			interpreter.Variable("pic1"),
			interpreter.OpCode{
				Cmd:  interpreter.OpCall,
				Args: []any{"CreatePic", int64(100), int64(100)},
			},
		},
	}

	err := vm.ExecuteOp(seq, op1)
	if err != nil {
		t.Fatalf("First ExecuteOp failed: %v", err)
	}

	pic1 := seq.GetVariable("pic1")
	pic1Int := vm.toInt(pic1)

	// Verify __return__ was cleared
	returnValue := seq.GetVariable("__return__")
	if returnValue != int64(0) {
		t.Errorf("Expected __return__ to be cleared (0), got %v", returnValue)
	}

	// Execute second function call
	op2 := interpreter.OpCode{
		Cmd: interpreter.OpAssign,
		Args: []any{
			interpreter.Variable("pic2"),
			interpreter.OpCode{
				Cmd:  interpreter.OpCall,
				Args: []any{"CreatePic", int64(200), int64(200)},
			},
		},
	}

	err = vm.ExecuteOp(seq, op2)
	if err != nil {
		t.Fatalf("Second ExecuteOp failed: %v", err)
	}

	pic2 := seq.GetVariable("pic2")
	pic2Int := vm.toInt(pic2)

	// Verify second call got a different ID
	if pic1Int == pic2Int {
		t.Errorf("Expected different picture IDs, got pic1=%d, pic2=%d", pic1Int, pic2Int)
	}

	// Verify __return__ was cleared again
	returnValue = seq.GetVariable("__return__")
	if returnValue != int64(0) {
		t.Errorf("Expected __return__ to be cleared (0) after second call, got %v", returnValue)
	}
}

// TestFunctionsWithoutReturnValues tests that functions without return values still work
func TestFunctionsWithoutReturnValues(t *testing.T) {
	// Create test VM
	vm, engine, state, _ := newTestVM()

	// Create a sequencer
	seq := NewSequencer([]interpreter.OpCode{}, TIME, nil)

	// Create a picture first
	picID := engine.CreatePic(100, 100)

	// Test DelPic (no return value)
	op := interpreter.OpCode{
		Cmd:  interpreter.OpCall,
		Args: []any{"DelPic", int64(picID)},
	}

	err := vm.ExecuteOp(seq, op)
	if err != nil {
		t.Fatalf("ExecuteOp failed: %v", err)
	}

	// Verify picture was deleted
	if state.GetPicture(picID) != nil {
		t.Errorf("Picture %d should have been deleted", picID)
	}

	// Verify __return__ is 0 (default) or not set
	returnValue := seq.GetVariable("__return__")
	// GetVariable returns 0 for undefined variables, so we expect 0 or int64(0)
	returnInt := vm.toInt(returnValue)
	if returnInt != 0 {
		t.Errorf("Expected __return__ to be 0 for function without return value, got %d", returnInt)
	}
}
