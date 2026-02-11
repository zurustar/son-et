package vm

import (
	"os"
	"testing"
)

// helperCreateTempFile creates a temporary file for testing and registers cleanup.
func helperCreateTempFile(t *testing.T) *os.File {
	t.Helper()
	f, err := os.CreateTemp("", "fht-test-*")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	t.Cleanup(func() {
		// Close if still open (ignore error since it may already be closed)
		f.Close()
		os.Remove(f.Name())
	})
	return f
}

// TestFileHandleTable_Open tests the Open method of FileHandleTable.
// Requirements: 3.1, 3.2
func TestFileHandleTable_Open(t *testing.T) {
	t.Run("returns handle >= 1", func(t *testing.T) {
		// Requirement 3.2: Handles start from 1.
		fht := NewFileHandleTable()
		f := helperCreateTempFile(t)

		handle := fht.Open(f)
		if handle < minHandleID {
			t.Errorf("handle should be >= %d, got %d", minHandleID, handle)
		}
	})

	t.Run("assigns sequential handles", func(t *testing.T) {
		// Requirement 3.2: Assigns smallest unused handle.
		fht := NewFileHandleTable()
		f1 := helperCreateTempFile(t)
		f2 := helperCreateTempFile(t)
		f3 := helperCreateTempFile(t)

		h1 := fht.Open(f1)
		h2 := fht.Open(f2)
		h3 := fht.Open(f3)

		if h1 != 1 {
			t.Errorf("first handle should be 1, got %d", h1)
		}
		if h2 != 2 {
			t.Errorf("second handle should be 2, got %d", h2)
		}
		if h3 != 3 {
			t.Errorf("third handle should be 3, got %d", h3)
		}
	})
}

// TestFileHandleTable_Get tests the Get method of FileHandleTable.
// Requirements: 3.1, 3.5
func TestFileHandleTable_Get(t *testing.T) {
	t.Run("returns correct file entry", func(t *testing.T) {
		// Requirement 3.1: Manages mapping from integer handle to *os.File.
		fht := NewFileHandleTable()
		f := helperCreateTempFile(t)

		handle := fht.Open(f)
		entry, err := fht.Get(handle)
		if err != nil {
			t.Fatalf("Get returned error: %v", err)
		}
		if entry == nil {
			t.Fatal("Get returned nil entry")
		}
		if entry.file != f {
			t.Error("Get returned entry with wrong file")
		}
	})

	t.Run("returns error for invalid handle", func(t *testing.T) {
		// Requirement 3.5: Invalid handle returns error.
		fht := NewFileHandleTable()

		_, err := fht.Get(999)
		if err == nil {
			t.Error("Get should return error for invalid handle")
		}
	})

	t.Run("returns error for handle 0", func(t *testing.T) {
		// Handles start from 1, so 0 is always invalid.
		fht := NewFileHandleTable()

		_, err := fht.Get(0)
		if err == nil {
			t.Error("Get should return error for handle 0")
		}
	})

	t.Run("returns error for negative handle", func(t *testing.T) {
		fht := NewFileHandleTable()

		_, err := fht.Get(-1)
		if err == nil {
			t.Error("Get should return error for negative handle")
		}
	})
}

// TestFileHandleTable_Close tests the Close method of FileHandleTable.
// Requirements: 3.3, 3.5
func TestFileHandleTable_Close(t *testing.T) {
	t.Run("releases handle and closes file", func(t *testing.T) {
		// Requirement 3.3: Close releases handle.
		fht := NewFileHandleTable()
		f := helperCreateTempFile(t)

		handle := fht.Open(f)
		err := fht.Close(handle)
		if err != nil {
			t.Fatalf("Close returned error: %v", err)
		}

		// Handle should now be invalid
		_, err = fht.Get(handle)
		if err == nil {
			t.Error("Get should return error after Close")
		}
	})

	t.Run("returns error for invalid handle", func(t *testing.T) {
		// Requirement 3.5: Invalid handle returns error.
		fht := NewFileHandleTable()

		err := fht.Close(999)
		if err == nil {
			t.Error("Close should return error for invalid handle")
		}
	})

	t.Run("returns error for already closed handle", func(t *testing.T) {
		fht := NewFileHandleTable()
		f := helperCreateTempFile(t)

		handle := fht.Open(f)
		err := fht.Close(handle)
		if err != nil {
			t.Fatalf("first Close returned error: %v", err)
		}

		// Second close should fail
		err = fht.Close(handle)
		if err == nil {
			t.Error("Close should return error for already closed handle")
		}
	})
}

// TestFileHandleTable_CloseAll tests the CloseAll method of FileHandleTable.
// Requirements: 3.4
func TestFileHandleTable_CloseAll(t *testing.T) {
	t.Run("closes all files and invalidates all handles", func(t *testing.T) {
		// Requirement 3.4: CloseAll closes all open files.
		fht := NewFileHandleTable()
		f1 := helperCreateTempFile(t)
		f2 := helperCreateTempFile(t)
		f3 := helperCreateTempFile(t)

		h1 := fht.Open(f1)
		h2 := fht.Open(f2)
		h3 := fht.Open(f3)

		fht.CloseAll()

		// All handles should now be invalid
		for _, h := range []int{h1, h2, h3} {
			_, err := fht.Get(h)
			if err == nil {
				t.Errorf("Get(handle=%d) should return error after CloseAll", h)
			}
		}
	})

	t.Run("works on empty table", func(t *testing.T) {
		// CloseAll on empty table should not panic.
		fht := NewFileHandleTable()
		fht.CloseAll() // should not panic
	})
}

// TestFileHandleTable_HandleReuse tests that closed handles are reused.
// Requirements: 3.2, 3.3
func TestFileHandleTable_HandleReuse(t *testing.T) {
	t.Run("reuses smallest unused handle after Close", func(t *testing.T) {
		// Requirement 3.2: Assigns smallest unused handle.
		// Requirement 3.3: Released handles are reusable.
		fht := NewFileHandleTable()
		f1 := helperCreateTempFile(t)
		f2 := helperCreateTempFile(t)
		f3 := helperCreateTempFile(t)
		f4 := helperCreateTempFile(t)

		h1 := fht.Open(f1) // handle 1
		h2 := fht.Open(f2) // handle 2
		_ = fht.Open(f3)   // handle 3

		// Close handle 1
		err := fht.Close(h1)
		if err != nil {
			t.Fatalf("Close returned error: %v", err)
		}

		// Next open should reuse handle 1 (smallest unused)
		h4 := fht.Open(f4)
		if h4 != h1 {
			t.Errorf("expected handle %d to be reused, got %d", h1, h4)
		}

		// Close handle 2
		f5 := helperCreateTempFile(t)
		err = fht.Close(h2)
		if err != nil {
			t.Fatalf("Close returned error: %v", err)
		}

		// Next open should reuse handle 2
		h5 := fht.Open(f5)
		if h5 != h2 {
			t.Errorf("expected handle %d to be reused, got %d", h2, h5)
		}
	})
}

// TestFileHandleTable_ResetReader tests the ResetReader method.
func TestFileHandleTable_ResetReader(t *testing.T) {
	t.Run("resets bufio.Reader for existing handle", func(t *testing.T) {
		fht := NewFileHandleTable()
		f := helperCreateTempFile(t)

		// Write some data
		_, err := f.WriteString("hello\nworld\n")
		if err != nil {
			t.Fatalf("failed to write: %v", err)
		}
		_, err = f.Seek(0, 0)
		if err != nil {
			t.Fatalf("failed to seek: %v", err)
		}

		handle := fht.Open(f)

		// Get entry and initialize reader
		entry, err := fht.Get(handle)
		if err != nil {
			t.Fatalf("Get returned error: %v", err)
		}

		// Reader should be nil initially (lazy initialization)
		if entry.reader != nil {
			t.Error("reader should be nil initially (lazy initialization)")
		}

		// ResetReader on nil reader should not panic
		fht.ResetReader(handle)
	})

	t.Run("no-op for invalid handle", func(t *testing.T) {
		fht := NewFileHandleTable()
		// Should not panic
		fht.ResetReader(999)
	})
}
