package engine

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"os"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/font/opentype"
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
// Attempts to load TrueType fonts from system paths, falling back to basicfont if unavailable.
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

	// Try to load system font
	// Common Japanese font paths on macOS
	fontPaths := []string{
		"/System/Library/Fonts/ヒラギノ明朝 ProN.ttc",
		"/System/Library/Fonts/ヒラギノ角ゴシック W3.ttc",
		"/System/Library/Fonts/ヒラギノ角ゴシック W4.ttc",
		"/Library/Fonts/Arial Unicode.ttf",
		"/System/Library/Fonts/Supplemental/Arial Unicode.ttf",
	}

	// Try each font path
	for _, fontPath := range fontPaths {
		if _, err := os.Stat(fontPath); err == nil {
			if face := tr.loadFont(fontPath, float64(size)); face != nil {
				tr.currentFont = face
				tr.engine.logger.LogInfo("Loaded font: %s (size %d)", fontPath, size)
				return
			}
		}
	}

	// Fall back to basicfont if no system font could be loaded
	tr.currentFont = basicfont.Face7x13
	tr.engine.logger.LogInfo("Warning: Could not load system font, using basicfont for: %s (size %d)", name, size)
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

	tr.engine.logger.LogDebug("TextWrite: before drawing, rgba.Bounds()=%v, pic.Width=%d, pic.Height=%d",
		rgba.Bounds(), pic.Width, pic.Height)

	// Clear the text area first to prevent alpha blending artifacts
	// This is needed because text rendering uses alpha blending, so drawing text
	// multiple times on the same area causes the old text to show through
	textWidth := len(text) * tr.currentFontSize // Rough estimate
	textHeight := tr.currentFontSize + 4        // Add some padding

	// Use opaque white background to completely erase previous text
	// (transparent background doesn't work well with antialiased text)
	clearColor := color.RGBA{255, 255, 255, 255} // Opaque white
	for py := 0; py < textHeight; py++ {
		for px := 0; px < textWidth; px++ {
			if x+px >= 0 && x+px < pic.Width && y+py >= 0 && y+py < pic.Height {
				rgba.Set(x+px, y+py, clearColor)
			}
		}
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

	// Update picture dimensions if image bounds changed
	newBounds := rgba.Bounds()
	pic.Width = newBounds.Dx()
	pic.Height = newBounds.Dy()

	tr.engine.logger.LogDebug("Text written: %q at (%d, %d) on picture %d (new size: %dx%d)",
		text, x, y, picID, pic.Width, pic.Height)
	return nil
}

// MeasureText returns the width and height of the text in pixels.
func (tr *TextRenderer) MeasureText(text string) (int, int) {
	bounds, advance := font.BoundString(tr.currentFont, text)
	width := advance.Ceil()
	height := (bounds.Max.Y - bounds.Min.Y).Ceil()
	return width, height
}

// loadFont loads a TrueType font from file.
// Supports both single fonts (.ttf) and font collections (.ttc).
func (tr *TextRenderer) loadFont(path string, size float64) font.Face {
	fontData, err := os.ReadFile(path)
	if err != nil {
		tr.engine.logger.LogDebug("Failed to read font file %s: %v", path, err)
		return nil
	}

	// Try to parse as a single font first
	tt, err := opentype.Parse(fontData)
	if err != nil {
		// If that fails, try as a font collection (.ttc)
		collection, err := opentype.ParseCollection(fontData)
		if err != nil {
			tr.engine.logger.LogDebug("Failed to parse font %s: %v", path, err)
			return nil
		}
		// Use the first font in the collection
		if collection.NumFonts() > 0 {
			tt, err = collection.Font(0)
			if err != nil {
				tr.engine.logger.LogDebug("Failed to get font from collection %s: %v", path, err)
				return nil
			}
		} else {
			tr.engine.logger.LogDebug("Font collection %s is empty", path)
			return nil
		}
	}

	face, err := opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		tr.engine.logger.LogDebug("Failed to create font face for %s: %v", path, err)
		return nil
	}

	return face
}
