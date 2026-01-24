package engine

import (
	"fmt"
	"image"
	"image/color"
	"io"
)

// BMPImageDecoder implements ImageDecoder for BMP files.
// Supports uncompressed and RLE-compressed BMP formats.
type BMPImageDecoder struct{}

// NewBMPImageDecoder creates a new BMP image decoder.
func NewBMPImageDecoder() *BMPImageDecoder {
	return &BMPImageDecoder{}
}

// Decode decodes a BMP image from a reader.
// Supports BI_RGB (uncompressed), BI_RLE8, and BI_RLE4 compression.
func (d *BMPImageDecoder) Decode(r io.Reader) (image.Image, string, error) {
	// Read BMP file header (14 bytes)
	fileHeader := make([]byte, 14)
	if _, err := io.ReadFull(r, fileHeader); err != nil {
		return nil, "", fmt.Errorf("failed to read BMP file header: %w", err)
	}

	// Check BMP signature
	if fileHeader[0] != 'B' || fileHeader[1] != 'M' {
		return nil, "", fmt.Errorf("invalid BMP signature")
	}

	// Read DIB header size
	dibHeaderSize := make([]byte, 4)
	if _, err := io.ReadFull(r, dibHeaderSize); err != nil {
		return nil, "", fmt.Errorf("failed to read DIB header size: %w", err)
	}

	headerSize := uint32(dibHeaderSize[0]) | uint32(dibHeaderSize[1])<<8 |
		uint32(dibHeaderSize[2])<<16 | uint32(dibHeaderSize[3])<<24

	// Read rest of DIB header
	dibHeader := make([]byte, headerSize-4)
	if _, err := io.ReadFull(r, dibHeader); err != nil {
		return nil, "", fmt.Errorf("failed to read DIB header: %w", err)
	}

	// Parse BITMAPINFOHEADER (most common format)
	if headerSize < 40 {
		return nil, "", fmt.Errorf("unsupported BMP header size: %d", headerSize)
	}

	width := int(int32(dibHeader[0]) | int32(dibHeader[1])<<8 | int32(dibHeader[2])<<16 | int32(dibHeader[3])<<24)
	height := int(int32(dibHeader[4]) | int32(dibHeader[5])<<8 | int32(dibHeader[6])<<16 | int32(dibHeader[7])<<24)
	bitsPerPixel := uint16(dibHeader[10]) | uint16(dibHeader[11])<<8
	compression := uint32(dibHeader[12]) | uint32(dibHeader[13])<<8 | uint32(dibHeader[14])<<16 | uint32(dibHeader[15])<<24

	// Handle negative height (top-down bitmap)
	topDown := false
	if height < 0 {
		height = -height
		topDown = true
	}

	// Read color palette if present
	var palette []color.RGBA
	if bitsPerPixel <= 8 {
		numColors := 1 << bitsPerPixel
		palette = make([]color.RGBA, numColors)
		for i := 0; i < numColors; i++ {
			colorData := make([]byte, 4)
			if _, err := io.ReadFull(r, colorData); err != nil {
				return nil, "", fmt.Errorf("failed to read palette: %w", err)
			}
			palette[i] = color.RGBA{
				R: colorData[2],
				G: colorData[1],
				B: colorData[0],
				A: 255,
			}
		}
	}

	// Create image
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Decode based on compression type
	switch compression {
	case 0: // BI_RGB (uncompressed)
		if err := d.decodeUncompressed(r, img, width, height, bitsPerPixel, palette, topDown); err != nil {
			return nil, "", err
		}
	case 1: // BI_RLE8
		if bitsPerPixel != 8 {
			return nil, "", fmt.Errorf("BI_RLE8 requires 8 bits per pixel, got %d", bitsPerPixel)
		}
		if err := d.decodeRLE8(r, img, width, height, palette, topDown); err != nil {
			return nil, "", err
		}
	case 2: // BI_RLE4
		if bitsPerPixel != 4 {
			return nil, "", fmt.Errorf("BI_RLE4 requires 4 bits per pixel, got %d", bitsPerPixel)
		}
		if err := d.decodeRLE4(r, img, width, height, palette, topDown); err != nil {
			return nil, "", err
		}
	default:
		return nil, "", fmt.Errorf("unsupported BMP compression: %d", compression)
	}

	return img, "bmp", nil
}

// decodeUncompressed decodes an uncompressed BMP image.
func (d *BMPImageDecoder) decodeUncompressed(r io.Reader, img *image.RGBA, width, height int, bitsPerPixel uint16, palette []color.RGBA, topDown bool) error {
	// Calculate row size with padding
	rowSize := ((int(bitsPerPixel)*width + 31) / 32) * 4

	for y := 0; y < height; y++ {
		rowData := make([]byte, rowSize)
		if _, err := io.ReadFull(r, rowData); err != nil {
			return fmt.Errorf("failed to read row %d: %w", y, err)
		}

		// Determine actual y coordinate (BMP is bottom-up by default)
		actualY := height - 1 - y
		if topDown {
			actualY = y
		}

		// Decode pixels based on bits per pixel
		switch bitsPerPixel {
		case 24:
			for x := 0; x < width; x++ {
				offset := x * 3
				img.SetRGBA(x, actualY, color.RGBA{
					R: rowData[offset+2],
					G: rowData[offset+1],
					B: rowData[offset],
					A: 255,
				})
			}
		case 8:
			for x := 0; x < width; x++ {
				paletteIndex := rowData[x]
				if int(paletteIndex) < len(palette) {
					img.SetRGBA(x, actualY, palette[paletteIndex])
				}
			}
		case 4:
			for x := 0; x < width; x++ {
				byteIndex := x / 2
				shift := uint(4 * (1 - (x % 2)))
				paletteIndex := (rowData[byteIndex] >> shift) & 0x0F
				if int(paletteIndex) < len(palette) {
					img.SetRGBA(x, actualY, palette[paletteIndex])
				}
			}
		case 1:
			for x := 0; x < width; x++ {
				byteIndex := x / 8
				bitIndex := uint(7 - (x % 8))
				paletteIndex := (rowData[byteIndex] >> bitIndex) & 0x01
				if int(paletteIndex) < len(palette) {
					img.SetRGBA(x, actualY, palette[paletteIndex])
				}
			}
		default:
			return fmt.Errorf("unsupported bits per pixel: %d", bitsPerPixel)
		}
	}

	return nil
}

// decodeRLE8 decodes an RLE8-compressed BMP image.
func (d *BMPImageDecoder) decodeRLE8(r io.Reader, img *image.RGBA, width, height int, palette []color.RGBA, topDown bool) error {
	x, y := 0, 0
	if topDown {
		y = 0
	} else {
		y = height - 1
	}

	for {
		// Read two bytes
		code := make([]byte, 2)
		if _, err := io.ReadFull(r, code); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to read RLE8 code: %w", err)
		}

		count := int(code[0])
		value := code[1]

		if count == 0 {
			// Escape code
			switch value {
			case 0: // End of line
				x = 0
				if topDown {
					y++
				} else {
					y--
				}
			case 1: // End of bitmap
				return nil
			case 2: // Delta
				delta := make([]byte, 2)
				if _, err := io.ReadFull(r, delta); err != nil {
					return fmt.Errorf("failed to read RLE8 delta: %w", err)
				}
				x += int(delta[0])
				if topDown {
					y += int(delta[1])
				} else {
					y -= int(delta[1])
				}
			default: // Absolute mode
				absCount := int(value)
				absData := make([]byte, absCount)
				if _, err := io.ReadFull(r, absData); err != nil {
					return fmt.Errorf("failed to read RLE8 absolute data: %w", err)
				}
				// Word-align
				if absCount%2 == 1 {
					io.ReadFull(r, make([]byte, 1))
				}
				for i := 0; i < absCount && x < width; i++ {
					if y >= 0 && y < height && int(absData[i]) < len(palette) {
						img.SetRGBA(x, y, palette[absData[i]])
					}
					x++
				}
			}
		} else {
			// Encoded mode
			for i := 0; i < count && x < width; i++ {
				if y >= 0 && y < height && int(value) < len(palette) {
					img.SetRGBA(x, y, palette[value])
				}
				x++
			}
		}
	}

	return nil
}

// decodeRLE4 decodes an RLE4-compressed BMP image.
func (d *BMPImageDecoder) decodeRLE4(r io.Reader, img *image.RGBA, width, height int, palette []color.RGBA, topDown bool) error {
	x, y := 0, 0
	if topDown {
		y = 0
	} else {
		y = height - 1
	}

	for {
		// Read two bytes
		code := make([]byte, 2)
		if _, err := io.ReadFull(r, code); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to read RLE4 code: %w", err)
		}

		count := int(code[0])
		value := code[1]

		if count == 0 {
			// Escape code
			switch value {
			case 0: // End of line
				x = 0
				if topDown {
					y++
				} else {
					y--
				}
			case 1: // End of bitmap
				return nil
			case 2: // Delta
				delta := make([]byte, 2)
				if _, err := io.ReadFull(r, delta); err != nil {
					return fmt.Errorf("failed to read RLE4 delta: %w", err)
				}
				x += int(delta[0])
				if topDown {
					y += int(delta[1])
				} else {
					y -= int(delta[1])
				}
			default: // Absolute mode
				absCount := int(value)
				byteCount := (absCount + 1) / 2
				absData := make([]byte, byteCount)
				if _, err := io.ReadFull(r, absData); err != nil {
					return fmt.Errorf("failed to read RLE4 absolute data: %w", err)
				}
				// Word-align
				if byteCount%2 == 1 {
					io.ReadFull(r, make([]byte, 1))
				}
				for i := 0; i < absCount && x < width; i++ {
					byteIndex := i / 2
					shift := uint(4 * (1 - (i % 2)))
					paletteIndex := (absData[byteIndex] >> shift) & 0x0F
					if y >= 0 && y < height && int(paletteIndex) < len(palette) {
						img.SetRGBA(x, y, palette[paletteIndex])
					}
					x++
				}
			}
		} else {
			// Encoded mode - alternate between two 4-bit values
			value1 := (value >> 4) & 0x0F
			value2 := value & 0x0F
			for i := 0; i < count && x < width; i++ {
				paletteIndex := value1
				if i%2 == 1 {
					paletteIndex = value2
				}
				if y >= 0 && y < height && int(paletteIndex) < len(palette) {
					img.SetRGBA(x, y, palette[paletteIndex])
				}
				x++
			}
		}
	}

	return nil
}

// DecodeConfig decodes only the image configuration without full decoding.
func (d *BMPImageDecoder) DecodeConfig(r io.Reader) (image.Config, string, error) {
	// Read BMP file header (14 bytes)
	fileHeader := make([]byte, 14)
	if _, err := io.ReadFull(r, fileHeader); err != nil {
		return image.Config{}, "", fmt.Errorf("failed to read BMP file header: %w", err)
	}

	// Check BMP signature
	if fileHeader[0] != 'B' || fileHeader[1] != 'M' {
		return image.Config{}, "", fmt.Errorf("invalid BMP signature")
	}

	// Read DIB header size
	dibHeaderSize := make([]byte, 4)
	if _, err := io.ReadFull(r, dibHeaderSize); err != nil {
		return image.Config{}, "", fmt.Errorf("failed to read DIB header size: %w", err)
	}

	headerSize := uint32(dibHeaderSize[0]) | uint32(dibHeaderSize[1])<<8 |
		uint32(dibHeaderSize[2])<<16 | uint32(dibHeaderSize[3])<<24

	// Read rest of DIB header
	dibHeader := make([]byte, headerSize-4)
	if _, err := io.ReadFull(r, dibHeader); err != nil {
		return image.Config{}, "", fmt.Errorf("failed to read DIB header: %w", err)
	}

	// Parse dimensions
	if headerSize < 40 {
		return image.Config{}, "", fmt.Errorf("unsupported BMP header size: %d", headerSize)
	}

	width := int(int32(dibHeader[0]) | int32(dibHeader[1])<<8 | int32(dibHeader[2])<<16 | int32(dibHeader[3])<<24)
	height := int(int32(dibHeader[4]) | int32(dibHeader[5])<<8 | int32(dibHeader[6])<<16 | int32(dibHeader[7])<<24)

	// Handle negative height
	if height < 0 {
		height = -height
	}

	return image.Config{
		ColorModel: color.RGBAModel,
		Width:      width,
		Height:     height,
	}, "bmp", nil
}
