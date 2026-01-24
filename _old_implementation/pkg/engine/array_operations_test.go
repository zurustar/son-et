package engine

import (
	"testing"
)

// TestArraySize tests the ArraySize function
// Requirement 31.1: WHEN ArraySize is called, THE Runtime SHALL return the number of elements in the array
func TestArraySize(t *testing.T) {
	tests := []struct {
		name     string
		arr      []int
		expected int
	}{
		{
			name:     "empty array",
			arr:      []int{},
			expected: 0,
		},
		{
			name:     "single element",
			arr:      []int{42},
			expected: 1,
		},
		{
			name:     "multiple elements",
			arr:      []int{1, 2, 3, 4, 5},
			expected: 5,
		},
		{
			name:     "array with negative values",
			arr:      []int{-1, -2, -3},
			expected: 3,
		},
		{
			name:     "array with zeros",
			arr:      []int{0, 0, 0, 0},
			expected: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ArraySize(tt.arr)
			if result != tt.expected {
				t.Errorf("ArraySize() = %d, want %d", result, tt.expected)
			}
		})
	}
}

// TestDelArrayAll tests the DelArrayAll function
// Requirement 31.2: WHEN DelArrayAll is called, THE Runtime SHALL remove all elements from the array
func TestDelArrayAll(t *testing.T) {
	tests := []struct {
		name string
		arr  []int
	}{
		{
			name: "empty array",
			arr:  []int{},
		},
		{
			name: "single element",
			arr:  []int{42},
		},
		{
			name: "multiple elements",
			arr:  []int{1, 2, 3, 4, 5},
		},
		{
			name: "large array",
			arr:  make([]int, 1000),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DelArrayAll(tt.arr)
			if len(result) != 0 {
				t.Errorf("DelArrayAll() returned array with %d elements, want 0", len(result))
			}
			// Verify it's a valid empty slice, not nil
			if result == nil {
				t.Error("DelArrayAll() returned nil, want empty slice")
			}
		})
	}
}

// TestDelArrayAt tests the DelArrayAt function
// Requirement 31.3: WHEN DelArrayAt is called, THE Runtime SHALL remove the element at the specified index
func TestDelArrayAt(t *testing.T) {
	tests := []struct {
		name     string
		arr      []int
		index    int
		expected []int
		wantErr  bool
	}{
		{
			name:     "remove first element",
			arr:      []int{1, 2, 3, 4, 5},
			index:    0,
			expected: []int{2, 3, 4, 5},
			wantErr:  false,
		},
		{
			name:     "remove middle element",
			arr:      []int{1, 2, 3, 4, 5},
			index:    2,
			expected: []int{1, 2, 4, 5},
			wantErr:  false,
		},
		{
			name:     "remove last element",
			arr:      []int{1, 2, 3, 4, 5},
			index:    4,
			expected: []int{1, 2, 3, 4},
			wantErr:  false,
		},
		{
			name:     "remove from single element array",
			arr:      []int{42},
			index:    0,
			expected: []int{},
			wantErr:  false,
		},
		{
			name:     "negative index (out of bounds)",
			arr:      []int{1, 2, 3},
			index:    -1,
			expected: []int{1, 2, 3}, // Should return unchanged
			wantErr:  true,
		},
		{
			name:     "index too large (out of bounds)",
			arr:      []int{1, 2, 3},
			index:    5,
			expected: []int{1, 2, 3}, // Should return unchanged
			wantErr:  true,
		},
		{
			name:     "empty array",
			arr:      []int{},
			index:    0,
			expected: []int{}, // Should return unchanged
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DelArrayAt(tt.arr, tt.index)

			// Check length
			if len(result) != len(tt.expected) {
				t.Errorf("DelArrayAt() returned array with %d elements, want %d", len(result), len(tt.expected))
				return
			}

			// Check contents
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("DelArrayAt() result[%d] = %d, want %d", i, result[i], tt.expected[i])
				}
			}
		})
	}
}

// TestInsArrayAt tests the InsArrayAt function
// Requirement 31.4: WHEN InsArrayAt is called, THE Runtime SHALL insert an element at the specified index
func TestInsArrayAt(t *testing.T) {
	tests := []struct {
		name     string
		arr      []int
		index    int
		value    int
		expected []int
		wantErr  bool
	}{
		{
			name:     "insert at beginning",
			arr:      []int{2, 3, 4, 5},
			index:    0,
			value:    1,
			expected: []int{1, 2, 3, 4, 5},
			wantErr:  false,
		},
		{
			name:     "insert in middle",
			arr:      []int{1, 2, 4, 5},
			index:    2,
			value:    3,
			expected: []int{1, 2, 3, 4, 5},
			wantErr:  false,
		},
		{
			name:     "insert at end (append)",
			arr:      []int{1, 2, 3, 4},
			index:    4,
			value:    5,
			expected: []int{1, 2, 3, 4, 5},
			wantErr:  false,
		},
		{
			name:     "insert into empty array",
			arr:      []int{},
			index:    0,
			value:    42,
			expected: []int{42},
			wantErr:  false,
		},
		{
			name:     "insert negative value",
			arr:      []int{1, 2, 3},
			index:    1,
			value:    -99,
			expected: []int{1, -99, 2, 3},
			wantErr:  false,
		},
		{
			name:     "negative index (out of bounds)",
			arr:      []int{1, 2, 3},
			index:    -1,
			value:    99,
			expected: []int{1, 2, 3}, // Should return unchanged
			wantErr:  true,
		},
		{
			name:     "index too large (out of bounds)",
			arr:      []int{1, 2, 3},
			index:    5,
			value:    99,
			expected: []int{1, 2, 3}, // Should return unchanged
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := InsArrayAt(tt.arr, tt.index, tt.value)

			// Check length
			if len(result) != len(tt.expected) {
				t.Errorf("InsArrayAt() returned array with %d elements, want %d", len(result), len(tt.expected))
				return
			}

			// Check contents
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("InsArrayAt() result[%d] = %d, want %d", i, result[i], tt.expected[i])
				}
			}
		})
	}
}

// TestArrayAutomaticResizing tests that arrays automatically resize during operations
// Requirement 31.5: THE Runtime SHALL automatically resize arrays as needed during insertion and deletion
func TestArrayAutomaticResizing(t *testing.T) {
	t.Run("multiple insertions grow array", func(t *testing.T) {
		arr := []int{}

		// Insert multiple elements
		arr = InsArrayAt(arr, 0, 1)
		arr = InsArrayAt(arr, 1, 2)
		arr = InsArrayAt(arr, 2, 3)
		arr = InsArrayAt(arr, 3, 4)
		arr = InsArrayAt(arr, 4, 5)

		if len(arr) != 5 {
			t.Errorf("After 5 insertions, array size = %d, want 5", len(arr))
		}

		expected := []int{1, 2, 3, 4, 5}
		for i := range arr {
			if arr[i] != expected[i] {
				t.Errorf("arr[%d] = %d, want %d", i, arr[i], expected[i])
			}
		}
	})

	t.Run("multiple deletions shrink array", func(t *testing.T) {
		arr := []int{1, 2, 3, 4, 5}

		// Delete multiple elements
		arr = DelArrayAt(arr, 4) // Remove 5
		arr = DelArrayAt(arr, 3) // Remove 4
		arr = DelArrayAt(arr, 2) // Remove 3

		if len(arr) != 2 {
			t.Errorf("After 3 deletions, array size = %d, want 2", len(arr))
		}

		expected := []int{1, 2}
		for i := range arr {
			if arr[i] != expected[i] {
				t.Errorf("arr[%d] = %d, want %d", i, arr[i], expected[i])
			}
		}
	})

	t.Run("insert and delete operations", func(t *testing.T) {
		arr := []int{1, 2, 3}

		// Insert at middle
		arr = InsArrayAt(arr, 1, 99)
		if len(arr) != 4 {
			t.Errorf("After insertion, array size = %d, want 4", len(arr))
		}

		// Delete from middle
		arr = DelArrayAt(arr, 1)
		if len(arr) != 3 {
			t.Errorf("After deletion, array size = %d, want 3", len(arr))
		}

		// Verify contents
		expected := []int{1, 2, 3}
		for i := range arr {
			if arr[i] != expected[i] {
				t.Errorf("arr[%d] = %d, want %d", i, arr[i], expected[i])
			}
		}
	})
}

// TestArrayOperationsEdgeCases tests edge cases for array operations
func TestArrayOperationsEdgeCases(t *testing.T) {
	t.Run("DelArrayAll on already empty array", func(t *testing.T) {
		arr := []int{}
		result := DelArrayAll(arr)
		if len(result) != 0 {
			t.Errorf("DelArrayAll on empty array returned %d elements, want 0", len(result))
		}
	})

	t.Run("ArraySize on nil slice", func(t *testing.T) {
		var arr []int
		result := ArraySize(arr)
		if result != 0 {
			t.Errorf("ArraySize on nil slice = %d, want 0", result)
		}
	})

	t.Run("operations preserve array independence", func(t *testing.T) {
		original := []int{1, 2, 3}

		// DelArrayAt should not modify original
		modified := DelArrayAt(original, 1)
		if len(original) != 3 {
			t.Errorf("DelArrayAt modified original array, len = %d, want 3", len(original))
		}
		if len(modified) != 2 {
			t.Errorf("DelArrayAt result len = %d, want 2", len(modified))
		}

		// InsArrayAt should not modify original
		modified2 := InsArrayAt(original, 1, 99)
		if len(original) != 3 {
			t.Errorf("InsArrayAt modified original array, len = %d, want 3", len(original))
		}
		if len(modified2) != 4 {
			t.Errorf("InsArrayAt result len = %d, want 4", len(modified2))
		}
	})
}
