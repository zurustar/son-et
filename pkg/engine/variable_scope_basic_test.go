package engine

import (
	"testing"
)

// TestSetVMVarBasic tests basic SetVMVar functionality
func TestSetVMVarBasic(t *testing.T) {
	// Reset VM state
	vmLock.Lock()
	mainSequencer = nil
	vmLock.Unlock()

	// Set a variable
	SetVMVar("x", 100)

	// Verify it was set
	vmLock.Lock()
	defer vmLock.Unlock()

	if mainSequencer == nil {
		t.Fatal("mainSequencer is nil after SetVMVar")
	}

	if val, ok := mainSequencer.vars["x"]; !ok {
		t.Error("Variable x not found in mainSequencer.vars")
	} else if val != 100 {
		t.Errorf("Expected x=100, got %v", val)
	}
}

// TestAssignBasic tests basic Assign functionality
func TestAssignBasic(t *testing.T) {
	// Reset VM state
	vmLock.Lock()
	mainSequencer = nil
	vmLock.Unlock()

	// Use Assign
	result := Assign("y", 200)

	// Should return the value
	if result != 200 {
		t.Errorf("Expected Assign to return 200, got %v", result)
	}

	// Verify it was set in VM
	vmLock.Lock()
	defer vmLock.Unlock()

	if mainSequencer == nil {
		t.Fatal("mainSequencer is nil after Assign")
	}

	if val, ok := mainSequencer.vars["y"]; !ok {
		t.Error("Variable y not found in mainSequencer.vars")
	} else if val != 200 {
		t.Errorf("Expected y=200, got %v", val)
	}
}

// TestResolveArgBasic tests basic variable resolution
func TestResolveArgBasic(t *testing.T) {
	// Reset VM state
	vmLock.Lock()
	mainSequencer = nil
	vmLock.Unlock()

	// Set a variable
	SetVMVar("z", 300)

	vmLock.Lock()
	defer vmLock.Unlock()

	if mainSequencer == nil {
		t.Fatal("mainSequencer is nil")
	}

	// Resolve the variable
	val := ResolveArg(Variable("z"), mainSequencer)
	if val != 300 {
		t.Errorf("Expected z=300, got %v", val)
	}
}

// TestCaseInsensitiveBasic tests case-insensitive variable names
func TestCaseInsensitiveBasic(t *testing.T) {
	// Reset VM state
	vmLock.Lock()
	mainSequencer = nil
	vmLock.Unlock()

	// Set variable with mixed case
	SetVMVar("WinW", 1280)

	vmLock.Lock()
	defer vmLock.Unlock()

	if mainSequencer == nil {
		t.Fatal("mainSequencer is nil")
	}

	// Should be accessible with different casing
	testCases := []string{"winw", "WINW", "WinW"}
	for _, varName := range testCases {
		val := ResolveArg(Variable(varName), mainSequencer)
		if val != 1280 {
			t.Errorf("Expected %s=1280, got %v", varName, val)
		}
	}
}
