package graphics

import (
	"image/color"
	"testing"
)

func TestColorFromInt(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected color.RGBA
	}{
		{
			name:     "black",
			input:    0x000000,
			expected: color.RGBA{R: 0, G: 0, B: 0, A: 0xFF},
		},
		{
			name:     "white",
			input:    0xFFFFFF,
			expected: color.RGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF},
		},
		{
			name:     "red",
			input:    0xFF0000,
			expected: color.RGBA{R: 0xFF, G: 0, B: 0, A: 0xFF},
		},
		{
			name:     "green",
			input:    0x00FF00,
			expected: color.RGBA{R: 0, G: 0xFF, B: 0, A: 0xFF},
		},
		{
			name:     "blue",
			input:    0x0000FF,
			expected: color.RGBA{R: 0, G: 0, B: 0xFF, A: 0xFF},
		},
		{
			name:     "custom color",
			input:    0x123456,
			expected: color.RGBA{R: 0x12, G: 0x34, B: 0x56, A: 0xFF},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ColorFromInt(tt.input)
			rgba, ok := result.(color.RGBA)
			if !ok {
				t.Fatalf("expected color.RGBA, got %T", result)
			}
			if rgba != tt.expected {
				t.Errorf("ColorFromInt(0x%06X) = %+v, want %+v", tt.input, rgba, tt.expected)
			}
		})
	}
}

func TestColorToInt(t *testing.T) {
	tests := []struct {
		name     string
		input    color.Color
		expected int
	}{
		{
			name:     "black",
			input:    color.RGBA{R: 0, G: 0, B: 0, A: 0xFF},
			expected: 0x000000,
		},
		{
			name:     "white",
			input:    color.RGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF},
			expected: 0xFFFFFF,
		},
		{
			name:     "red",
			input:    color.RGBA{R: 0xFF, G: 0, B: 0, A: 0xFF},
			expected: 0xFF0000,
		},
		{
			name:     "green",
			input:    color.RGBA{R: 0, G: 0xFF, B: 0, A: 0xFF},
			expected: 0x00FF00,
		},
		{
			name:     "blue",
			input:    color.RGBA{R: 0, G: 0, B: 0xFF, A: 0xFF},
			expected: 0x0000FF,
		},
		{
			name:     "custom color",
			input:    color.RGBA{R: 0x12, G: 0x34, B: 0x56, A: 0xFF},
			expected: 0x123456,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ColorToInt(tt.input)
			if result != tt.expected {
				t.Errorf("ColorToInt(%+v) = 0x%06X, want 0x%06X", tt.input, result, tt.expected)
			}
		})
	}
}

func TestColorRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		color int
	}{
		{"black", 0x000000},
		{"white", 0xFFFFFF},
		{"red", 0xFF0000},
		{"green", 0x00FF00},
		{"blue", 0x0000FF},
		{"yellow", 0xFFFF00},
		{"cyan", 0x00FFFF},
		{"magenta", 0xFF00FF},
		{"custom1", 0x123456},
		{"custom2", 0xABCDEF},
		{"custom3", 0x7F7F7F},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert int -> Color -> int
			colorObj := ColorFromInt(tt.color)
			result := ColorToInt(colorObj)
			if result != tt.color {
				t.Errorf("Round trip failed: 0x%06X -> %+v -> 0x%06X", tt.color, colorObj, result)
			}
		})
	}
}

func TestTransparentColor(t *testing.T) {
	expected := color.RGBA{R: 0, G: 0, B: 0, A: 0xFF}
	if TransparentColor != expected {
		t.Errorf("TransparentColor = %+v, want %+v", TransparentColor, expected)
	}
}
