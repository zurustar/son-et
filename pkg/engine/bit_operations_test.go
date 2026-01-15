package engine

import (
	"testing"
)

// TestMakeLong tests the MakeLong function
// Requirement 32.1: WHEN MakeLong is called, THE Runtime SHALL combine two 16-bit values into a single 32-bit value
// Requirement 32.4: THE Runtime SHALL preserve bit patterns during pack and unpack operations
func TestMakeLong(t *testing.T) {
	tests := []struct {
		name     string
		lowWord  int
		hiWord   int
		expected int
	}{
		{
			name:     "zero values",
			lowWord:  0x0000,
			hiWord:   0x0000,
			expected: 0x00000000,
		},
		{
			name:     "low word only",
			lowWord:  0x1234,
			hiWord:   0x0000,
			expected: 0x00001234,
		},
		{
			name:     "high word only",
			lowWord:  0x0000,
			hiWord:   0x5678,
			expected: 0x56780000,
		},
		{
			name:     "both words",
			lowWord:  0x1234,
			hiWord:   0x5678,
			expected: 0x56781234,
		},
		{
			name:     "all bits set",
			lowWord:  0xFFFF,
			hiWord:   0xFFFF,
			expected: int(0xFFFFFFFF), // Cast to handle signed representation
		},
		{
			name:     "alternating bits",
			lowWord:  0xAAAA,
			hiWord:   0x5555,
			expected: 0x5555AAAA,
		},
		{
			name:     "values exceeding 16 bits (should be masked)",
			lowWord:  0x12345, // Will be masked to 0x2345
			hiWord:   0x67890, // Will be masked to 0x7890
			expected: 0x78902345,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MakeLong(tt.lowWord, tt.hiWord)
			if result != tt.expected {
				t.Errorf("MakeLong(0x%04X, 0x%04X) = 0x%08X, want 0x%08X",
					tt.lowWord, tt.hiWord, result, tt.expected)
			}
		})
	}
}

// TestGetHiWord tests the GetHiWord function
// Requirement 32.2: WHEN GetHiWord is called, THE Runtime SHALL extract the upper 16 bits of a 32-bit value
// Requirement 32.4: THE Runtime SHALL preserve bit patterns during pack and unpack operations
func TestGetHiWord(t *testing.T) {
	tests := []struct {
		name     string
		value    int
		expected int
	}{
		{
			name:     "zero value",
			value:    0x00000000,
			expected: 0x0000,
		},
		{
			name:     "low word only",
			value:    0x00001234,
			expected: 0x0000,
		},
		{
			name:     "high word only",
			value:    0x56780000,
			expected: 0x5678,
		},
		{
			name:     "both words",
			value:    0x56781234,
			expected: 0x5678,
		},
		{
			name:     "all bits set",
			value:    -1, // 0xFFFFFFFF
			expected: 0xFFFF,
		},
		{
			name:     "alternating bits",
			value:    0x5555AAAA,
			expected: 0x5555,
		},
		{
			name:     "negative value",
			value:    -256, // 0xFFFFFF00
			expected: 0xFFFF,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetHiWord(tt.value)
			if result != tt.expected {
				t.Errorf("GetHiWord(0x%08X) = 0x%04X, want 0x%04X",
					tt.value, result, tt.expected)
			}
		})
	}
}

// TestGetLowWord tests the GetLowWord function
// Requirement 32.3: WHEN GetLowWord is called, THE Runtime SHALL extract the lower 16 bits of a 32-bit value
// Requirement 32.4: THE Runtime SHALL preserve bit patterns during pack and unpack operations
func TestGetLowWord(t *testing.T) {
	tests := []struct {
		name     string
		value    int
		expected int
	}{
		{
			name:     "zero value",
			value:    0x00000000,
			expected: 0x0000,
		},
		{
			name:     "low word only",
			value:    0x00001234,
			expected: 0x1234,
		},
		{
			name:     "high word only",
			value:    0x56780000,
			expected: 0x0000,
		},
		{
			name:     "both words",
			value:    0x56781234,
			expected: 0x1234,
		},
		{
			name:     "all bits set",
			value:    -1, // 0xFFFFFFFF
			expected: 0xFFFF,
		},
		{
			name:     "alternating bits",
			value:    0x5555AAAA,
			expected: 0xAAAA,
		},
		{
			name:     "negative value",
			value:    -256, // 0xFFFFFF00
			expected: 0xFF00,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetLowWord(tt.value)
			if result != tt.expected {
				t.Errorf("GetLowWord(0x%08X) = 0x%04X, want 0x%04X",
					tt.value, result, tt.expected)
			}
		})
	}
}

// TestBitOperationsRoundTrip tests packing and unpacking
// Requirement 32.4: THE Runtime SHALL preserve bit patterns during pack and unpack operations
func TestBitOperationsRoundTrip(t *testing.T) {
	tests := []struct {
		name    string
		lowWord int
		hiWord  int
	}{
		{
			name:    "zero values",
			lowWord: 0x0000,
			hiWord:  0x0000,
		},
		{
			name:    "typical values",
			lowWord: 0x1234,
			hiWord:  0x5678,
		},
		{
			name:    "all bits set",
			lowWord: 0xFFFF,
			hiWord:  0xFFFF,
		},
		{
			name:    "alternating bits",
			lowWord: 0xAAAA,
			hiWord:  0x5555,
		},
		{
			name:    "single bit patterns",
			lowWord: 0x0001,
			hiWord:  0x8000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Pack the values
			packed := MakeLong(tt.lowWord, tt.hiWord)

			// Unpack and verify
			extractedLow := GetLowWord(packed)
			extractedHi := GetHiWord(packed)

			// Mask input values to 16 bits for comparison
			expectedLow := tt.lowWord & 0xFFFF
			expectedHi := tt.hiWord & 0xFFFF

			if extractedLow != expectedLow {
				t.Errorf("Round trip failed for low word: got 0x%04X, want 0x%04X",
					extractedLow, expectedLow)
			}
			if extractedHi != expectedHi {
				t.Errorf("Round trip failed for high word: got 0x%04X, want 0x%04X",
					extractedHi, expectedHi)
			}
		})
	}
}

// TestBitOperationsSignedUnsigned tests signed and unsigned handling
// Requirement 32.5: THE Runtime SHALL handle signed and unsigned values correctly
func TestBitOperationsSignedUnsigned(t *testing.T) {
	tests := []struct {
		name        string
		value       int
		description string
	}{
		{
			name:        "positive value",
			value:       0x12345678,
			description: "positive 32-bit value",
		},
		{
			name:        "negative value",
			value:       -1,
			description: "all bits set (0xFFFFFFFF)",
		},
		{
			name:        "small negative",
			value:       -256,
			description: "0xFFFFFF00",
		},
		{
			name:        "large positive",
			value:       0x7FFFFFFF,
			description: "maximum positive 32-bit signed int",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Extract words
			hi := GetHiWord(tt.value)
			low := GetLowWord(tt.value)

			// Reconstruct
			reconstructed := MakeLong(low, hi)

			// For negative values, we need to compare the bit patterns, not the signed values
			// Convert both to uint32 for comparison to avoid signed integer issues
			originalBits := uint32(tt.value)
			reconstructedBits := uint32(reconstructed)

			// Verify reconstruction matches original bit pattern
			if reconstructedBits != originalBits {
				t.Errorf("Signed/unsigned handling failed for %s: original=0x%08X, reconstructed=0x%08X",
					tt.description, originalBits, reconstructedBits)
			}
		})
	}
}
