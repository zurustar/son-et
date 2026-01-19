package engine

import (
	"os"
	"testing"
)

// TestWriteIniInt_ReadBack tests writing and reading integer values
// Requirement 27.1: WHEN WriteIniInt is called, THE Runtime SHALL write an integer value to the specified INI section and entry
// Requirement 27.2: WHEN GetIniInt is called, THE Runtime SHALL read an integer value from the specified INI section and entry
func TestWriteIniInt_ReadBack(t *testing.T) {
	testFile := "test_ini_int.ini"
	defer os.Remove(testFile)

	// Write a value
	WriteIniInt(testFile, "Section1", "Key1", 42)

	// Read it back
	value := GetIniInt(testFile, "Section1", "Key1", 0)
	if value != 42 {
		t.Errorf("Expected 42, got %d", value)
	}
}

// TestWriteIniStr_ReadBack tests writing and reading string values
// Requirement 27.3: WHEN WriteIniStr is called, THE Runtime SHALL write a string value to the specified INI section and entry
// Requirement 27.4: WHEN GetIniStr is called, THE Runtime SHALL read a string value from the specified INI section and entry
func TestWriteIniStr_ReadBack(t *testing.T) {
	testFile := "test_ini_str.ini"
	defer os.Remove(testFile)

	// Write a value
	WriteIniStr(testFile, "Section1", "Key1", "Hello World")

	// Read it back
	value := GetIniStr(testFile, "Section1", "Key1", "")
	if value != "Hello World" {
		t.Errorf("Expected 'Hello World', got '%s'", value)
	}
}

// TestGetIniInt_DefaultValue tests that default value is returned when key doesn't exist
// Requirement 27.2: WHEN GetIniInt is called, THE Runtime SHALL read an integer value from the specified INI section and entry
func TestGetIniInt_DefaultValue(t *testing.T) {
	testFile := "nonexistent.ini"

	// Try to read from non-existent file
	value := GetIniInt(testFile, "Section1", "Key1", 99)
	if value != 99 {
		t.Errorf("Expected default value 99, got %d", value)
	}
}

// TestGetIniStr_DefaultValue tests that default value is returned when key doesn't exist
// Requirement 27.4: WHEN GetIniStr is called, THE Runtime SHALL read a string value from the specified INI section and entry
func TestGetIniStr_DefaultValue(t *testing.T) {
	testFile := "nonexistent.ini"

	// Try to read from non-existent file
	value := GetIniStr(testFile, "Section1", "Key1", "default")
	if value != "default" {
		t.Errorf("Expected default value 'default', got '%s'", value)
	}
}

// TestIniFileCreation tests that INI files are created if they don't exist
// Requirement 27.5: THE Runtime SHALL create INI files if they do not exist
func TestIniFileCreation(t *testing.T) {
	testFile := "test_create.ini"
	defer os.Remove(testFile)

	// Write to non-existent file
	WriteIniInt(testFile, "NewSection", "NewKey", 123)

	// Verify file was created
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("INI file was not created")
	}

	// Verify content
	value := GetIniInt(testFile, "NewSection", "NewKey", 0)
	if value != 123 {
		t.Errorf("Expected 123, got %d", value)
	}
}

// TestIniMultipleSections tests writing to multiple sections
// Requirement 27.1, 27.3: Test multiple sections and entries
func TestIniMultipleSections(t *testing.T) {
	testFile := "test_multi.ini"
	defer os.Remove(testFile)

	// Write to multiple sections
	WriteIniInt(testFile, "Section1", "IntKey", 100)
	WriteIniStr(testFile, "Section1", "StrKey", "value1")
	WriteIniInt(testFile, "Section2", "IntKey", 200)
	WriteIniStr(testFile, "Section2", "StrKey", "value2")

	// Read back and verify
	if val := GetIniInt(testFile, "Section1", "IntKey", 0); val != 100 {
		t.Errorf("Section1.IntKey: expected 100, got %d", val)
	}
	if val := GetIniStr(testFile, "Section1", "StrKey", ""); val != "value1" {
		t.Errorf("Section1.StrKey: expected 'value1', got '%s'", val)
	}
	if val := GetIniInt(testFile, "Section2", "IntKey", 0); val != 200 {
		t.Errorf("Section2.IntKey: expected 200, got %d", val)
	}
	if val := GetIniStr(testFile, "Section2", "StrKey", ""); val != "value2" {
		t.Errorf("Section2.StrKey: expected 'value2', got '%s'", val)
	}
}

// TestIniOverwrite tests overwriting existing values
// Requirement 27.1, 27.3: Test updating existing entries
func TestIniOverwrite(t *testing.T) {
	testFile := "test_overwrite.ini"
	defer os.Remove(testFile)

	// Write initial value
	WriteIniInt(testFile, "Section1", "Key1", 42)

	// Overwrite with new value
	WriteIniInt(testFile, "Section1", "Key1", 84)

	// Verify new value
	value := GetIniInt(testFile, "Section1", "Key1", 0)
	if value != 84 {
		t.Errorf("Expected 84, got %d", value)
	}
}
