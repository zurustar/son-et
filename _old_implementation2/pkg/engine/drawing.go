package engine

import (
	"image"
	"image/color"
	"image/draw"
	"math"
)

// DrawingContext manages drawing state (line size, paint color, ROP mode).
type DrawingContext struct {
	lineSize   int
	paintColor color.Color
	ropMode    ROPMode
}

// ROPMode represents raster operation modes.
type ROPMode int

const (
	COPYPEN    ROPMode = iota // Normal copy (default)
	XORPEN                    // XOR operation
	MERGEPEN                  // OR operation
	NOTCOPYPEN                // NOT source
	MASKPEN                   // AND operation
)

// NewDrawingContext creates a new drawing context with default values.
func NewDrawingContext() *DrawingContext {
	return &DrawingContext{
		lineSize:   1,
		paintColor: color.Black,
		ropMode:    COPYPEN,
	}
}

// SetLineSize sets the line width for drawing operations.
func (dc *DrawingContext) SetLineSize(size int) {
	if size < 1 {
		size = 1
	}
	dc.lineSize = size
}

// GetLineSize returns the current line size.
func (dc *DrawingContext) GetLineSize() int {
	return dc.lineSize
}

// SetPaintColor sets the drawing color.
func (dc *DrawingContext) SetPaintColor(c color.Color) {
	dc.paintColor = c
}

// GetPaintColor returns the current paint color.
func (dc *DrawingContext) GetPaintColor() color.Color {
	return dc.paintColor
}

// SetROP sets the raster operation mode.
func (dc *DrawingContext) SetROP(mode ROPMode) {
	dc.ropMode = mode
}

// GetROP returns the current raster operation mode.
func (dc *DrawingContext) GetROP() ROPMode {
	return dc.ropMode
}

// DrawLine draws a line from (x1, y1) to (x2, y2) on the given image.
// Uses Bresenham's line algorithm.
func (dc *DrawingContext) DrawLine(img *image.RGBA, x1, y1, x2, y2 int) {
	// Bresenham's line algorithm
	dx := abs(x2 - x1)
	dy := abs(y2 - y1)
	sx := 1
	if x1 > x2 {
		sx = -1
	}
	sy := 1
	if y1 > y2 {
		sy = -1
	}
	err := dx - dy

	x, y := x1, y1
	for {
		// Draw pixel with line size
		dc.drawThickPixel(img, x, y)

		if x == x2 && y == y2 {
			break
		}

		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x += sx
		}
		if e2 < dx {
			err += dx
			y += sy
		}
	}
}

// DrawCircle draws a circle at (cx, cy) with the given radius.
// fillMode: 0=outline, 1=hatch, 2=solid
func (dc *DrawingContext) DrawCircle(img *image.RGBA, cx, cy, radius int, fillMode int) {
	if fillMode == 2 {
		// Solid fill
		dc.drawFilledCircle(img, cx, cy, radius)
	} else if fillMode == 1 {
		// Hatch fill
		dc.drawHatchedCircle(img, cx, cy, radius)
	} else {
		// Outline only
		dc.drawCircleOutline(img, cx, cy, radius)
	}
}

// DrawRect draws a rectangle from (x1, y1) to (x2, y2).
// fillMode: 0=outline, 1=hatch, 2=solid
func (dc *DrawingContext) DrawRect(img *image.RGBA, x1, y1, x2, y2 int, fillMode int) {
	// Ensure x1 <= x2 and y1 <= y2
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}

	if fillMode == 2 {
		// Solid fill
		dc.fillRect(img, x1, y1, x2, y2)
	} else if fillMode == 1 {
		// Hatch fill
		dc.hatchRect(img, x1, y1, x2, y2)
	} else {
		// Outline only
		dc.DrawLine(img, x1, y1, x2, y1) // Top
		dc.DrawLine(img, x2, y1, x2, y2) // Right
		dc.DrawLine(img, x2, y2, x1, y2) // Bottom
		dc.DrawLine(img, x1, y2, x1, y1) // Left
	}
}

// GetColor returns the color of the pixel at (x, y).
func GetColor(img *image.RGBA, x, y int) color.Color {
	bounds := img.Bounds()
	if x < bounds.Min.X || x >= bounds.Max.X || y < bounds.Min.Y || y >= bounds.Max.Y {
		return color.Transparent
	}
	return img.At(x, y)
}

// Helper functions

func (dc *DrawingContext) drawThickPixel(img *image.RGBA, x, y int) {
	if dc.lineSize == 1 {
		dc.setPixel(img, x, y)
	} else {
		// Draw a square of pixels for thick lines
		halfSize := dc.lineSize / 2
		for dy := -halfSize; dy <= halfSize; dy++ {
			for dx := -halfSize; dx <= halfSize; dx++ {
				dc.setPixel(img, x+dx, y+dy)
			}
		}
	}
}

func (dc *DrawingContext) setPixel(img *image.RGBA, x, y int) {
	bounds := img.Bounds()
	if x < bounds.Min.X || x >= bounds.Max.X || y < bounds.Min.Y || y >= bounds.Max.Y {
		return
	}

	switch dc.ropMode {
	case COPYPEN:
		img.Set(x, y, dc.paintColor)
	case XORPEN:
		existing := img.RGBAAt(x, y)
		r, g, b, a := dc.paintColor.RGBA()
		newColor := color.RGBA{
			R: existing.R ^ uint8(r>>8),
			G: existing.G ^ uint8(g>>8),
			B: existing.B ^ uint8(b>>8),
			A: uint8(a >> 8),
		}
		img.Set(x, y, newColor)
	case MERGEPEN:
		existing := img.RGBAAt(x, y)
		r, g, b, a := dc.paintColor.RGBA()
		newColor := color.RGBA{
			R: existing.R | uint8(r>>8),
			G: existing.G | uint8(g>>8),
			B: existing.B | uint8(b>>8),
			A: uint8(a >> 8),
		}
		img.Set(x, y, newColor)
	case NOTCOPYPEN:
		r, g, b, a := dc.paintColor.RGBA()
		newColor := color.RGBA{
			R: ^uint8(r >> 8),
			G: ^uint8(g >> 8),
			B: ^uint8(b >> 8),
			A: uint8(a >> 8),
		}
		img.Set(x, y, newColor)
	case MASKPEN:
		existing := img.RGBAAt(x, y)
		r, g, b, a := dc.paintColor.RGBA()
		newColor := color.RGBA{
			R: existing.R & uint8(r>>8),
			G: existing.G & uint8(g>>8),
			B: existing.B & uint8(b>>8),
			A: uint8(a >> 8),
		}
		img.Set(x, y, newColor)
	}
}

func (dc *DrawingContext) drawCircleOutline(img *image.RGBA, cx, cy, radius int) {
	// Midpoint circle algorithm
	x := radius
	y := 0
	err := 0

	for x >= y {
		dc.drawThickPixel(img, cx+x, cy+y)
		dc.drawThickPixel(img, cx+y, cy+x)
		dc.drawThickPixel(img, cx-y, cy+x)
		dc.drawThickPixel(img, cx-x, cy+y)
		dc.drawThickPixel(img, cx-x, cy-y)
		dc.drawThickPixel(img, cx-y, cy-x)
		dc.drawThickPixel(img, cx+y, cy-x)
		dc.drawThickPixel(img, cx+x, cy-y)

		if err <= 0 {
			y++
			err += 2*y + 1
		}
		if err > 0 {
			x--
			err -= 2*x + 1
		}
	}
}

func (dc *DrawingContext) drawFilledCircle(img *image.RGBA, cx, cy, radius int) {
	// Draw filled circle by scanning horizontally
	for y := -radius; y <= radius; y++ {
		x := int(math.Sqrt(float64(radius*radius - y*y)))
		for dx := -x; dx <= x; dx++ {
			dc.setPixel(img, cx+dx, cy+y)
		}
	}
}

func (dc *DrawingContext) drawHatchedCircle(img *image.RGBA, cx, cy, radius int) {
	// Draw hatch pattern (diagonal lines)
	for y := -radius; y <= radius; y++ {
		x := int(math.Sqrt(float64(radius*radius - y*y)))
		for dx := -x; dx <= x; dx++ {
			// Hatch pattern: diagonal lines every 4 pixels
			if (dx+y)%4 == 0 {
				dc.setPixel(img, cx+dx, cy+y)
			}
		}
	}
}

func (dc *DrawingContext) fillRect(img *image.RGBA, x1, y1, x2, y2 int) {
	// Use image/draw for efficient solid fill
	rect := image.Rect(x1, y1, x2+1, y2+1)
	draw.Draw(img, rect, &image.Uniform{dc.paintColor}, image.Point{}, draw.Over)
}

func (dc *DrawingContext) hatchRect(img *image.RGBA, x1, y1, x2, y2 int) {
	// Draw hatch pattern (diagonal lines)
	for y := y1; y <= y2; y++ {
		for x := x1; x <= x2; x++ {
			// Hatch pattern: diagonal lines every 4 pixels
			if (x+y)%4 == 0 {
				dc.setPixel(img, x, y)
			}
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
