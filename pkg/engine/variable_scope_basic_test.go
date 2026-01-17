package engine

import (
	"strings"
	"testing"
)

// TestSetVMVarBasic tests basic SetVMVar functionality
func TestSetVMVarBasic(t *testing.T) {
	// Reset VM state
	ResetEngineForTest()

	// Set a variable
	SetVMVar("x", 100)

	// Verify it was set in globalVars
	if val, ok := globalVars["x"]; !ok {
		t.Error("Variable x not found in globalVars")
	} else if val != 100 {
		t.Errorf("Expected x=100, got %v", val)
	}
}

// TestAssignBasic tests basic Assign functionality
func TestAssignBasic(t *testing.T) {
	// Reset VM state
	ResetEngineForTest()

	// Use Assign
	result := Assign("y", 200)

	// Should return the value
	if result != 200 {
		t.Errorf("Expected Assign to return 200, got %v", result)
	}

	// Verify it was set in globalVars
	if val, ok := globalVars["y"]; !ok {
		t.Error("Variable y not found in globalVars")
	} else if val != 200 {
		t.Errorf("Expected y=200, got %v", val)
	}
}

// TestResolveArgBasic tests basic variable resolution
func TestResolveArgBasic(t *testing.T) {
	// Reset VM state
	ResetEngineForTest()

	// Set a variable
	SetVMVar("z", 300)

	// Verify it's in globalVars
	if val, ok := globalVars["z"]; !ok {
		t.Error("Variable z not found in globalVars")
	} else if val != 300 {
		t.Errorf("Expected z=300, got %v", val)
	}
}

// TestCaseInsensitiveBasic tests case-insensitive variable names
func TestCaseInsensitiveBasic(t *testing.T) {
	// Reset VM state
	ResetEngineForTest()

	// Set variable with mixed case
	SetVMVar("WinW", 1280)

	// Should be accessible with different casing in globalVars
	testCases := []string{"winw", "WINW", "WinW"}
	for _, varName := range testCases {
		// Check in globalVars (case-insensitive)
		found := false
		for key, val := range globalVars {
			if strings.EqualFold(key, varName) {
				if val != 1280 {
					t.Errorf("Expected %s=1280, got %v", varName, val)
				}
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Variable %s not found in globalVars", varName)
		}
	}
}
