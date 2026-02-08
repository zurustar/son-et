package graphics

import (
	"image/color"
)

// ColorFromInt converts a FILLY color format (0xRRGGBB) to color.Color.
// The input is a 24-bit RGB value where:
// - Bits 16-23: Red component
// - Bits 8-15: Green component
// - Bits 0-7: Blue component
func ColorFromInt(c int) color.Color {
	return color.RGBA{
		R: uint8((c >> 16) & 0xFF),
		G: uint8((c >> 8) & 0xFF),
		B: uint8(c & 0xFF),
		A: 0xFF,
	}
}

// ColorToInt converts a color.Color to FILLY color format (0xRRGGBB).
// Returns a 24-bit RGB value.
func ColorToInt(c color.Color) int {
	r, g, b, _ := c.RGBA()
	// RGBA() returns 16-bit values, so shift right by 8 to get 8-bit values
	return int(r>>8)<<16 | int(g>>8)<<8 | int(b>>8)
}

