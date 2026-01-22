package main

import (
	"reflect"
	"testing"
)

func TestReorderArgs(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "flags before positional",
			input:    []string{"--headless", "--timeout", "5s", "file.tfy"},
			expected: []string{"--headless", "--timeout", "5s", "file.tfy"},
		},
		{
			name:     "positional before flags",
			input:    []string{"file.tfy", "--headless", "--timeout", "5s"},
			expected: []string{"--headless", "--timeout", "5s", "file.tfy"},
		},
		{
			name:     "mixed order",
			input:    []string{"--headless", "file.tfy", "--timeout", "5s"},
			expected: []string{"--headless", "--timeout", "5s", "file.tfy"},
		},
		{
			name:     "only positional",
			input:    []string{"file.tfy"},
			expected: []string{"file.tfy"},
		},
		{
			name:     "only flags",
			input:    []string{"--headless", "--timeout", "5s"},
			expected: []string{"--headless", "--timeout", "5s"},
		},
		{
			name:     "debug flag with value",
			input:    []string{"file.tfy", "--debug", "2"},
			expected: []string{"--debug", "2", "file.tfy"},
		},
		{
			name:     "multiple positional args",
			input:    []string{"--headless", "file1.tfy", "file2.tfy"},
			expected: []string{"--headless", "file1.tfy", "file2.tfy"},
		},
		{
			name:     "short flags",
			input:    []string{"file.tfy", "-timeout", "5s"},
			expected: []string{"-timeout", "5s", "file.tfy"},
		},
		{
			name:     "empty args",
			input:    []string{},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := reorderArgs(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("reorderArgs(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
