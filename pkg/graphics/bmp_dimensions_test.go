package graphics

import (
	"encoding/binary"
	"testing"
)

// buildBMPHeader builds a minimal BMP byte stream with the given dimensions.
func buildBMPHeader(width, height int32, bitCount uint16, compression uint32, colorsUsed uint32) []byte {
	buf := make([]byte, 0, 64)
	// File header (14 bytes)
	buf = append(buf, 'B', 'M')
	buf = append(buf, 0, 0, 0, 0) // file size (ignored)
	buf = append(buf, 0, 0, 0, 0) // reserved
	dataOffset := uint32(54 + colorsUsed*4)
	off := make([]byte, 4)
	binary.LittleEndian.PutUint32(off, dataOffset)
	buf = append(buf, off...)
	// Info header (40 bytes)
	hdr := make([]byte, 40)
	binary.LittleEndian.PutUint32(hdr[0:], 40)
	binary.LittleEndian.PutUint32(hdr[4:], uint32(width))
	binary.LittleEndian.PutUint32(hdr[8:], uint32(height))
	binary.LittleEndian.PutUint16(hdr[12:], 1)
	binary.LittleEndian.PutUint16(hdr[14:], bitCount)
	binary.LittleEndian.PutUint32(hdr[16:], compression)
	binary.LittleEndian.PutUint32(hdr[32:], colorsUsed)
	buf = append(buf, hdr...)
	for i := uint32(0); i < colorsUsed; i++ {
		buf = append(buf, 0, 0, 0, 0)
	}
	for i := 0; i < 16; i++ {
		buf = append(buf, 0)
	}
	return buf
}

// TestDecodeBMPRejectsBadDimensions is a regression test for the bug where a
// negative width caused makeslice panic and huge dimensions caused OOM.
// See docs/bug-hunt-findings.md finding E.
func TestDecodeBMPRejectsBadDimensions(t *testing.T) {
	cases := []struct {
		name          string
		width, height int32
	}{
		{"negative width", -4, 1},
		{"negative height handled as topdown, zero width", 0, -1},
		{"zero dimensions", 0, 0},
		{"huge width", 1 << 20, 1},
		{"huge height", 1, 1 << 20},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data := buildBMPHeader(tc.width, tc.height, 8, biRGB, 1)
			img, err := DecodeBMPFromBytes(data) // must return an error, not panic
			if err == nil {
				t.Errorf("expected error for dimensions %dx%d, got image %v", tc.width, tc.height, img != nil)
			}
		})
	}
}

// TestDecodeBMPAcceptsValidDimensions ensures a small valid BMP still decodes.
func TestDecodeBMPAcceptsValidDimensions(t *testing.T) {
	data := buildBMPHeader(2, 2, 8, biRGB, 1)
	img, err := DecodeBMPFromBytes(data)
	if err != nil {
		t.Fatalf("unexpected error for valid 2x2 BMP: %v", err)
	}
	if img == nil {
		t.Fatal("expected a decoded image, got nil")
	}
	if b := img.Bounds(); b.Dx() != 2 || b.Dy() != 2 {
		t.Errorf("bounds = %v, want 2x2", b)
	}
}
