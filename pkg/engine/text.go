package engine

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// TextRenderer manages text rendering state and operations.
type TextRenderer struct {
	currentFont     font.Face
	currentFontSize int
	currentFontName string
	textColor       color.Color
	bgColor         color.Color
	backMode        int // 0=transparent, 1=opaque
	engine          *Engine
}

// NewTextRenderer creates a new text renderer.
func NewTextRenderer(engine *Engine) *TextRenderer {
	return &TextRenderer{
		currentFont:     basicfont.Face7x13,
		currentFontSize: 13,
		currentFontName: "default",
		textColor:       color.RGBA{0, 0, 0, 255},       // Black
		bgColor:         color.RGBA{255, 255, 255, 255}, // White
		backMode:        0,                              // Transparent
		engine:          engine,
	}
}

// SetFont sets the current font for text rendering.
// For now, we only support the basic font. In the future, this could load TrueType fonts.
func (tr *TextRenderer) SetFont(size int, name string, charset int) {
	tr.engine.logger.LogDebug("SetFont: size=%d, name=%s, charset=%d", size, name, charset)

	// Legacy support: Some scripts pass unreasonable sizes
	// If size > 200, it might be a legacy parameter order issue
	if size > 200 {
		tr.engine.logger.LogDebug("SetFont: Large size detected (%d), using default size 13", size)
		size = 13
	}

	tr.currentFontSize = size
	tr.currentFontName = name

	// For now, we use basicfont regardless of the requested font
	// In the future, we could load TrueType fonts based on name
	tr.currentFont = basicfont.Face7x13

	tr.engine.logger.LogInfo("Font set: %s (size %d)", name, size)
}

// SetTextColor sets the text color.
func (tr *TextRenderer) SetTextColor(r, g, b int) {
	tr.textColor = color.RGBA{uint8(r), uint8(g), uint8(b), 255}
	tr.engine.logger.LogDebug("Text color set: RGB(%d, %d, %d)", r, g, b)
}

// SetBgColor sets the background color for text.
func (tr *TextRenderer) SetBgColor(r, g, b int) {
	tr.bgColor = color.RGBA{uint8(r), uint8(g), uint8(b), 255}
	tr.engine.logger.LogDebug("Background color set: RGB(%d, %d, %d)", r, g, b)
}

// SetBackMode sets the background mode (0=transparent, 1=opaque).
func (tr *TextRenderer) SetBackMode(mode int) {
	tr.backMode = mode
	if mode == 0 {
		tr.engine.logger.LogDebug("Background mode: transparent")
	} else {
		tr.engine.logger.LogDebug("Background mode: opaque")
	}
}

// TextWrite draws text on a picture at the specified position.
func (tr *TextRenderer) TextWrite(text string, picID, x, y int) error {
	// Get the picture
	pic := tr.engine.state.GetPicture(picID)
	if pic == nil {
		return fmt.Errorf("picture %d not found", picID)
	}

	// Convert picture to RGBA if needed
	var rgba *image.RGBA
	switch img := pic.Image.(type) {
	case *image.RGBA:
		rgba = img
	default:
		// Convert to RGBA
		bounds := pic.Image.Bounds()
		rgba = image.NewRGBA(bounds)
		draw.Draw(rgba, bounds, pic.Image, bounds.Min, draw.Src)
		pic.Image = rgba
	}

	// Draw background if opaque mode
	if tr.backMode == 1 {
		// Measure text to get background rectangle
		bounds, _ := font.BoundString(tr.currentFont, text)
		width := (bounds.Max.X - bounds.Min.X).Ceil()
		height := (bounds.Max.Y - bounds.Min.Y).Ceil()

		// Draw background rectangle
		bgRect := image.Rect(x, y, x+width, y+height)
		draw.Draw(rgba, bgRect, &image.Uniform{tr.bgColor}, image.Point{}, draw.Src)
	}

	// Create a drawer for text rendering
	drawer := &font.Drawer{
		Dst:  rgba,
		Src:  image.NewUniform(tr.textColor),
		Face: tr.currentFont,
		Dot:  fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y + tr.currentFontSize)},
	}

	// Draw the text
	drawer.DrawString(text)

	tr.engine.logger.LogDebug("Text written: %q at (%d, %d) on picture %d", text, x, y, picID)
	return nil
}

// MeasureText returns the width and height of the text in pixels.
func (tr *TextRenderer) MeasureText(text string) (int, int) {
	bounds, advance := font.BoundString(tr.currentFont, text)
	width := advance.Ceil()
	height := (bounds.Max.Y - bounds.Min.Y).Ceil()
	return width, height
}
