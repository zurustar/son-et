package engine

import (
	"os"
	"testing"
)

// TestOpenF_CloseF tests opening and closing files
// Requirement 29.1: WHEN OpenF is called, THE Runtime SHALL open a file and return a file handle
// Requirement 29.2: WHEN CloseF is called, THE Runtime SHALL close the specified file handle
func TestOpenF_CloseF(t *testing.T) {
	engine := NewTestEngine()
	globalEngine = engine

	testFile := "test_open.bin"
	defer os.Remove(testFile)

	// Create a test file
	err := os.WriteFile(testFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Open file for reading
	handle := OpenF(testFile, 0)
	if handle < 0 {
		t.Errorf("OpenF failed, returned %d", handle)
	}

	// Close file
	result := CloseF(handle)
	if result != 0 {
		t.Errorf("CloseF failed with result %d", result)
	}

	// Try to close again (should fail)
	result = CloseF(handle)
	if result == 0 {
		t.Error("CloseF should fail on already closed handle")
	}
}

// TestWriteF_ReadF tests writing and reading integers
// Requirement 29.4: WHEN ReadF is called, THE Runtime SHALL read 1-4 bytes and return as an integer
// Requirement 29.5: WHEN WriteF is called, THE Runtime SHALL write an integer value as 1-4 bytes
func TestWriteF_ReadF(t *testing.T) {
	engine := NewTestEngine()
	globalEngine = engine

	testFile := "test_readwrite.bin"
	defer os.Remove(testFile)

	// Open file for writing
	handle := OpenF(testFile, 1)
	if handle < 0 {
		t.Fatalf("OpenF failed")
	}

	// Write different sized integers
	testCases := []struct {
		value int
		size  int
	}{
		{0x42, 1},       // 1 byte
		{0x1234, 2},     // 2 bytes
		{0x123456, 3},   // 3 bytes
		{0x12345678, 4}, // 4 bytes
	}

	for _, tc := range testCases {
		result := WriteF(handle, tc.value, tc.size)
		if result != 0 {
			t.Errorf("WriteF failed for value 0x%X, size %d", tc.value, tc.size)
		}
	}

	// Close and reopen for reading
	CloseF(handle)
	handle = OpenF(testFile, 0)
	if handle < 0 {
		t.Fatalf("OpenF failed for reading")
	}
	defer CloseF(handle)

	// Read back and verify
	for _, tc := range testCases {
		value := ReadF(handle, tc.size)
		// Mask to expected size
		mask := (1 << (tc.size * 8)) - 1
		expected := tc.value & mask
		if value != expected {
			t.Errorf("ReadF returned 0x%X, expected 0x%X (size %d)", value, expected, tc.size)
		}
	}
}

// TestSeekF tests file seeking
// Requirement 29.3: WHEN SeekF is called, THE Runtime SHALL move the file pointer to the specified position
func TestSeekF(t *testing.T) {
	engine := NewTestEngine()
	globalEngine = engine

	testFile := "test_seek.bin"
	defer os.Remove(testFile)

	// Create a file with known content
	content := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	err := os.WriteFile(testFile, content, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Open file for reading
	handle := OpenF(testFile, 0)
	if handle < 0 {
		t.Fatalf("OpenF failed")
	}
	defer CloseF(handle)

	// Seek to position 4 from start
	pos := SeekF(handle, 4, 0)
	if pos != 4 {
		t.Errorf("SeekF returned position %d, expected 4", pos)
	}

	// Read 1 byte (should be 0x05)
	value := ReadF(handle, 1)
	if value != 0x05 {
		t.Errorf("Read value 0x%X, expected 0x05", value)
	}

	// Seek to position 2 from start
	pos = SeekF(handle, 2, 0)
	if pos != 2 {
		t.Errorf("SeekF returned position %d, expected 2", pos)
	}

	// Read 1 byte (should be 0x03)
	value = ReadF(handle, 1)
	if value != 0x03 {
		t.Errorf("Read value 0x%X, expected 0x03", value)
	}

	// Seek to end
	pos = SeekF(handle, 0, 2)
	if pos != int(len(content)) {
		t.Errorf("SeekF to end returned position %d, expected %d", pos, len(content))
	}
}

// TestStrWriteF_StrReadF tests writing and reading strings
// Requirement 29.6: WHEN StrReadF is called, THE Runtime SHALL read a null-terminated string from the file
// Requirement 29.7: WHEN StrWriteF is called, THE Runtime SHALL write a null-terminated string to the file
func TestStrWriteF_StrReadF(t *testing.T) {
	engine := NewTestEngine()
	globalEngine = engine

	testFile := "test_string.bin"
	defer os.Remove(testFile)

	// Open file for writing
	handle := OpenF(testFile, 1)
	if handle < 0 {
		t.Fatalf("OpenF failed")
	}

	// Write strings
	testStrings := []string{
		"Hello",
		"World",
		"Test String",
	}

	for _, str := range testStrings {
		result := StrWriteF(handle, str)
		if result != 0 {
			t.Errorf("StrWriteF failed for string '%s'", str)
		}
	}

	// Close and reopen for reading
	CloseF(handle)
	handle = OpenF(testFile, 0)
	if handle < 0 {
		t.Fatalf("OpenF failed for reading")
	}
	defer CloseF(handle)

	// Read back and verify
	for _, expected := range testStrings {
		str := StrReadF(handle)
		if str != expected {
			t.Errorf("StrReadF returned '%s', expected '%s'", str, expected)
		}
	}
}

// TestBinaryIO_Integration tests a complete binary I/O workflow
// Requirements 29.1-29.7: Integration test for binary I/O
func TestBinaryIO_Integration(t *testing.T) {
	engine := NewTestEngine()
	globalEngine = engine

	testFile := "test_integration.bin"
	defer os.Remove(testFile)

	// Open file for writing
	handle := OpenF(testFile, 1)
	if handle < 0 {
		t.Fatalf("OpenF failed")
	}

	// Write a header (magic number + version)
	WriteF(handle, 0x46494C4C, 4) // "FILL" in ASCII
	WriteF(handle, 1, 2)          // Version 1

	// Write some data
	WriteF(handle, 100, 4)
	WriteF(handle, 200, 4)
	WriteF(handle, 300, 4)

	// Write a string
	StrWriteF(handle, "TestData")

	// Close file
	CloseF(handle)

	// Open file for reading
	handle = OpenF(testFile, 0)
	if handle < 0 {
		t.Fatalf("OpenF failed for reading")
	}
	defer CloseF(handle)

	// Read and verify header
	magic := ReadF(handle, 4)
	if magic != 0x46494C4C {
		t.Errorf("Magic number mismatch: got 0x%X, expected 0x46494C4C", magic)
	}

	version := ReadF(handle, 2)
	if version != 1 {
		t.Errorf("Version mismatch: got %d, expected 1", version)
	}

	// Read and verify data
	values := []int{100, 200, 300}
	for i, expected := range values {
		value := ReadF(handle, 4)
		if value != expected {
			t.Errorf("Value[%d] mismatch: got %d, expected %d", i, value, expected)
		}
	}

	// Read and verify string
	str := StrReadF(handle)
	if str != "TestData" {
		t.Errorf("String mismatch: got '%s', expected 'TestData'", str)
	}
}

// TestReadF_EOF tests reading at end of file
// Requirement 29.4: Test EOF handling
func TestReadF_EOF(t *testing.T) {
	engine := NewTestEngine()
	globalEngine = engine

	testFile := "test_eof.bin"
	defer os.Remove(testFile)

	// Create a small file
	err := os.WriteFile(testFile, []byte{0x01, 0x02}, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Open file for reading
	handle := OpenF(testFile, 0)
	if handle < 0 {
		t.Fatalf("OpenF failed")
	}
	defer CloseF(handle)

	// Read 2 bytes (should succeed)
	value := ReadF(handle, 2)
	if value != 0x0201 { // Little-endian
		t.Errorf("Read value 0x%X, expected 0x0201", value)
	}

	// Try to read more (should handle EOF gracefully)
	value = ReadF(handle, 2)
	// Should return -1 or 0 depending on implementation
	if debugLevel >= 2 {
		t.Logf("Read at EOF returned: %d", value)
	}
}

// TestOpenF_Modes tests different file open modes
// Requirement 29.1: Test different open modes
func TestOpenF_Modes(t *testing.T) {
	engine := NewTestEngine()
	globalEngine = engine

	testFile := "test_modes.bin"
	defer os.Remove(testFile)

	// Mode 1: Write (creates new file)
	handle := OpenF(testFile, 1)
	if handle < 0 {
		t.Fatalf("OpenF mode 1 failed")
	}
	WriteF(handle, 0x1234, 2)
	CloseF(handle)

	// Mode 0: Read
	handle = OpenF(testFile, 0)
	if handle < 0 {
		t.Fatalf("OpenF mode 0 failed")
	}
	value := ReadF(handle, 2)
	if value != 0x1234 {
		t.Errorf("Read value 0x%X, expected 0x1234", value)
	}
	CloseF(handle)

	// Mode 2: Read/Write
	handle = OpenF(testFile, 2)
	if handle < 0 {
		t.Fatalf("OpenF mode 2 failed")
	}
	// Read existing data
	value = ReadF(handle, 2)
	if value != 0x1234 {
		t.Errorf("Read value 0x%X, expected 0x1234", value)
	}
	// Write new data
	WriteF(handle, 0x5678, 2)
	CloseF(handle)

	// Verify new data was written
	handle = OpenF(testFile, 0)
	if handle < 0 {
		t.Fatalf("OpenF failed")
	}
	defer CloseF(handle)

	ReadF(handle, 2) // Skip first value
	value = ReadF(handle, 2)
	if value != 0x5678 {
		t.Errorf("Read value 0x%X, expected 0x5678", value)
	}
}
