package engine

import (
	"testing"

	"github.com/zurustar/son-et/pkg/compiler/interpreter"
)

func TestNewSequencer(t *testing.T) {
	commands := []interpreter.OpCode{
		{Cmd: interpreter.OpLiteral, Args: []any{42}},
	}

	seq := NewSequencer(commands, TIME, nil)

	if seq == nil {
		t.Fatal("NewSequencer returned nil")
	}
	if len(seq.commands) != 1 {
		t.Errorf("Expected 1 command, got %d", len(seq.commands))
	}
	if seq.pc != 0 {
		t.Errorf("Expected pc=0, got %d", seq.pc)
	}
	if !seq.active {
		t.Error("Expected sequence to be active")
	}
	if seq.mode != TIME {
		t.Errorf("Expected TIME mode, got %v", seq.mode)
	}
	if seq.vars == nil {
		t.Error("Expected vars map to be initialized")
	}
}

func TestSequencer_VariableScoping(t *testing.T) {
	t.Run("Get/Set in current scope", func(t *testing.T) {
		seq := NewSequencer(nil, TIME, nil)

		seq.SetVariable("x", 42)
		val := seq.GetVariable("x")

		if val != 42 {
			t.Errorf("Expected 42, got %v", val)
		}
	})

	t.Run("Case-insensitive variable names", func(t *testing.T) {
		seq := NewSequencer(nil, TIME, nil)

		seq.SetVariable("MyVar", 100)
		val1 := seq.GetVariable("myvar")
		val2 := seq.GetVariable("MYVAR")
		val3 := seq.GetVariable("MyVar")

		if val1 != 100 || val2 != 100 || val3 != 100 {
			t.Errorf("Case-insensitive lookup failed: %v, %v, %v", val1, val2, val3)
		}
	})

	t.Run("Parent scope lookup", func(t *testing.T) {
		parent := NewSequencer(nil, TIME, nil)
		parent.SetVariable("parentVar", 200)

		child := NewSequencer(nil, TIME, parent)

		val := child.GetVariable("parentVar")
		if val != 200 {
			t.Errorf("Expected 200 from parent scope, got %v", val)
		}
	})

	t.Run("Child scope shadows parent", func(t *testing.T) {
		parent := NewSequencer(nil, TIME, nil)
		parent.SetVariable("x", 100)

		child := NewSequencer(nil, TIME, parent)
		// In FILLY, setting a variable that exists in parent updates the parent
		// (no shadowing - variables are shared across scope chain)
		child.SetVariable("x", 200)

		parentVal := parent.GetVariable("x")
		childVal := child.GetVariable("x")

		// Both should see the updated value
		if parentVal != 200 {
			t.Errorf("Parent x should be updated to 200, got %v", parentVal)
		}
		if childVal != 200 {
			t.Errorf("Child x should be 200, got %v", childVal)
		}
	})

	t.Run("Update parent variable from child", func(t *testing.T) {
		parent := NewSequencer(nil, TIME, nil)
		parent.SetVariable("shared", 100)

		child := NewSequencer(nil, TIME, parent)
		// Setting a variable that exists in parent should update parent
		child.SetVariable("shared", 300)

		parentVal := parent.GetVariable("shared")
		if parentVal != 300 {
			t.Errorf("Parent shared should be updated to 300, got %v", parentVal)
		}
	})

	t.Run("Default value for undefined variable", func(t *testing.T) {
		seq := NewSequencer(nil, TIME, nil)

		val := seq.GetVariable("undefined")
		if val != 0 {
			t.Errorf("Expected default value 0, got %v", val)
		}
	})
}

func TestSequencer_ArrayOperations(t *testing.T) {
	t.Run("Create and access array", func(t *testing.T) {
		seq := NewSequencer(nil, TIME, nil)

		seq.SetArrayElement("arr", 0, 10)
		seq.SetArrayElement("arr", 1, 20)
		seq.SetArrayElement("arr", 2, 30)

		if seq.GetArrayElement("arr", 0) != 10 {
			t.Errorf("Expected arr[0]=10, got %d", seq.GetArrayElement("arr", 0))
		}
		if seq.GetArrayElement("arr", 1) != 20 {
			t.Errorf("Expected arr[1]=20, got %d", seq.GetArrayElement("arr", 1))
		}
		if seq.GetArrayElement("arr", 2) != 30 {
			t.Errorf("Expected arr[2]=30, got %d", seq.GetArrayElement("arr", 2))
		}
	})

	t.Run("Auto-expansion on set", func(t *testing.T) {
		seq := NewSequencer(nil, TIME, nil)

		// Set element at index 10 (array doesn't exist yet)
		seq.SetArrayElement("arr", 10, 100)

		// Array should be expanded to size 11, zero-filled
		for i := 0; i < 10; i++ {
			if seq.GetArrayElement("arr", i) != 0 {
				t.Errorf("Expected arr[%d]=0 (zero-fill), got %d", i, seq.GetArrayElement("arr", i))
			}
		}
		if seq.GetArrayElement("arr", 10) != 100 {
			t.Errorf("Expected arr[10]=100, got %d", seq.GetArrayElement("arr", 10))
		}
	})

	t.Run("Auto-expansion on get", func(t *testing.T) {
		seq := NewSequencer(nil, TIME, nil)

		// Create small array
		seq.SetArrayElement("arr", 0, 5)

		// Access beyond current size
		val := seq.GetArrayElement("arr", 10)

		// Should return 0 and expand array
		if val != 0 {
			t.Errorf("Expected 0 for out-of-bounds access, got %d", val)
		}

		// Array should now be expanded
		if seq.GetArrayElement("arr", 5) != 0 {
			t.Error("Array should be zero-filled after expansion")
		}
	})

	t.Run("Case-insensitive array names", func(t *testing.T) {
		seq := NewSequencer(nil, TIME, nil)

		seq.SetArrayElement("MyArray", 0, 42)
		val1 := seq.GetArrayElement("myarray", 0)
		val2 := seq.GetArrayElement("MYARRAY", 0)

		if val1 != 42 || val2 != 42 {
			t.Errorf("Case-insensitive array access failed: %d, %d", val1, val2)
		}
	})

	t.Run("Array in parent scope", func(t *testing.T) {
		parent := NewSequencer(nil, TIME, nil)
		parent.SetArrayElement("parentArr", 0, 100)

		child := NewSequencer(nil, TIME, parent)

		val := child.GetArrayElement("parentArr", 0)
		if val != 100 {
			t.Errorf("Expected 100 from parent array, got %d", val)
		}
	})
}

func TestSequencer_ExecutionState(t *testing.T) {
	t.Run("Active/Deactivate", func(t *testing.T) {
		seq := NewSequencer(nil, TIME, nil)

		if !seq.IsActive() {
			t.Error("New sequence should be active")
		}

		seq.Deactivate()

		if seq.IsActive() {
			t.Error("Sequence should be inactive after Deactivate()")
		}
	})

	t.Run("Wait counter", func(t *testing.T) {
		seq := NewSequencer(nil, TIME, nil)

		if seq.IsWaiting() {
			t.Error("New sequence should not be waiting")
		}

		seq.SetWait(5)

		if !seq.IsWaiting() {
			t.Error("Sequence should be waiting after SetWait(5)")
		}

		seq.DecrementWait()
		seq.DecrementWait()
		seq.DecrementWait()

		if seq.waitCount != 2 {
			t.Errorf("Expected waitCount=2, got %d", seq.waitCount)
		}

		seq.DecrementWait()
		seq.DecrementWait()

		if seq.IsWaiting() {
			t.Error("Sequence should not be waiting after count reaches 0")
		}
	})

	t.Run("Program counter", func(t *testing.T) {
		commands := []interpreter.OpCode{
			{Cmd: interpreter.OpLiteral, Args: []any{1}},
			{Cmd: interpreter.OpLiteral, Args: []any{2}},
			{Cmd: interpreter.OpLiteral, Args: []any{3}},
		}
		seq := NewSequencer(commands, TIME, nil)

		if seq.GetPC() != 0 {
			t.Errorf("Expected initial PC=0, got %d", seq.GetPC())
		}

		seq.IncrementPC()
		if seq.GetPC() != 1 {
			t.Errorf("Expected PC=1, got %d", seq.GetPC())
		}

		seq.IncrementPC()
		seq.IncrementPC()

		if !seq.IsComplete() {
			t.Error("Sequence should be complete when PC >= len(commands)")
		}
	})

	t.Run("GetCurrentCommand", func(t *testing.T) {
		commands := []interpreter.OpCode{
			{Cmd: interpreter.OpLiteral, Args: []any{42}},
			{Cmd: interpreter.OpLiteral, Args: []any{100}},
		}
		seq := NewSequencer(commands, TIME, nil)

		cmd := seq.GetCurrentCommand()
		if cmd == nil {
			t.Fatal("GetCurrentCommand returned nil")
		}
		if cmd.Cmd != interpreter.OpLiteral {
			t.Errorf("Expected OpLiteral, got %v", cmd.Cmd)
		}
		if cmd.Args[0] != 42 {
			t.Errorf("Expected arg 42, got %v", cmd.Args[0])
		}

		seq.IncrementPC()
		cmd = seq.GetCurrentCommand()
		if cmd.Args[0] != 100 {
			t.Errorf("Expected arg 100, got %v", cmd.Args[0])
		}

		seq.IncrementPC()
		cmd = seq.GetCurrentCommand()
		if cmd != nil {
			t.Error("GetCurrentCommand should return nil when complete")
		}
	})
}

func TestSequencer_Metadata(t *testing.T) {
	seq := NewSequencer(nil, TIME, nil)

	seq.SetID(42)
	if seq.GetID() != 42 {
		t.Errorf("Expected ID=42, got %d", seq.GetID())
	}

	seq.SetGroupID(10)
	if seq.GetGroupID() != 10 {
		t.Errorf("Expected GroupID=10, got %d", seq.GetGroupID())
	}

	if seq.GetMode() != TIME {
		t.Errorf("Expected TIME mode, got %v", seq.GetMode())
	}
}
