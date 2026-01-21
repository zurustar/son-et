package engine

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
)

const (
	// Window border thickness for decorations
	BorderThickness = 4
)

var (
	// Window decoration colors
	titleBarColor  = color.RGBA{0, 0, 128, 255}     // Dark blue
	borderColor    = color.RGBA{192, 192, 192, 255} // Gray
	highlightColor = color.RGBA{255, 255, 255, 255} // White (for raised edge)
	shadowColor    = color.RGBA{0, 0, 0, 255}       // Black (for recessed edge)
	captionColor   = color.RGBA{255, 255, 255, 255} // White text
	captionFont    font.Face                        // Font for window captions
)

func init() {
	captionFont = basicfont.Face7x13
}

// EbitenRenderer implements the Renderer interface using Ebiten.
type EbitenRenderer struct {
	backgroundColor color.Color
	logger          *Logger
	frameCount      int // Frame counter for debug logging
}

// NewEbitenRenderer creates a new Ebiten-based renderer.
func NewEbitenRenderer() *EbitenRenderer {
	return &EbitenRenderer{
		backgroundColor: color.RGBA{0x1F, 0x7E, 0x7F, 0xff}, // Teal
		logger:          NewLogger(DebugLevelError),
	}
}

// SetLogger sets the logger for the renderer.
func (r *EbitenRenderer) SetLogger(logger *Logger) {
	r.logger = logger
}

// RenderFrame renders the current engine state to the screen.
// Rendering pipeline: desktop → windows → casts
func (r *EbitenRenderer) RenderFrame(screen image.Image, state *EngineState) {
	// Lock state for reading
	state.renderMutex.Lock()
	defer state.renderMutex.Unlock()

	r.frameCount++
	logThisFrame := (r.frameCount%60 == 1) // Log once per second (at 60fps)

	// Convert screen to Ebiten image
	ebitenScreen, ok := screen.(*ebiten.Image)
	if !ok {
		return
	}

	// Clear screen with background color
	ebitenScreen.Fill(r.backgroundColor)

	// Render all windows in z-order (creation order)
	windows := state.GetWindows()

	if logThisFrame && r.logger != nil {
		r.logger.LogDebug("RenderFrame: %d windows to render", len(windows))
	}

	for _, win := range windows {
		if !win.Visible {
			continue
		}

		r.renderWindow(ebitenScreen, state, win, logThisFrame)
	}
}

// renderWindow renders a single window with decorations and its casts.
func (r *EbitenRenderer) renderWindow(screen *ebiten.Image, state *EngineState, win *Window, logThisFrame bool) {
	// win.X and win.Y represent the top-left corner of the entire window (including decorations)
	// Content area starts at (win.X + BorderThickness, win.Y + TitleBarHeight + BorderThickness)

	winX := float32(win.X)
	winY := float32(win.Y)
	winW := float32(win.Width)
	winH := float32(win.Height)

	totalW := winW + float32(BorderThickness*2)
	totalH := winH + float32(BorderThickness*2) + float32(TitleBarHeight)

	// 1. Draw window frame background (gray)
	vector.DrawFilledRect(screen,
		winX,
		winY,
		totalW, totalH,
		borderColor, false)

	// 2. Draw 3D border effect
	// Top and left edges (highlight - raised effect)
	vector.StrokeLine(screen,
		winX, winY,
		winX+totalW, winY,
		1, highlightColor, false)
	vector.StrokeLine(screen,
		winX, winY,
		winX, winY+totalH,
		1, highlightColor, false)

	// Bottom and right edges (shadow - recessed effect)
	vector.StrokeLine(screen,
		winX, winY+totalH,
		winX+totalW, winY+totalH,
		1, shadowColor, false)
	vector.StrokeLine(screen,
		winX+totalW, winY,
		winX+totalW, winY+totalH,
		1, shadowColor, false)

	// 3. Draw title bar (blue)
	vector.DrawFilledRect(screen,
		winX+float32(BorderThickness),
		winY+float32(BorderThickness),
		winW, float32(TitleBarHeight),
		titleBarColor, false)

	// 4. Draw caption text if present
	if win.Caption != "" {
		// Draw caption text using basic font
		text.Draw(screen, win.Caption,
			captionFont,
			int(winX)+BorderThickness+4,
			int(winY)+BorderThickness+14, // Adjust Y for baseline
			captionColor)
	}

	// 5. Draw window content area
	r.renderWindowContent(screen, state, win, logThisFrame)

	// 6. Render casts for this window
	casts := state.GetCastsByWindow(win.ID)
	for _, cast := range casts {
		if !cast.Visible {
			continue
		}

		r.renderCast(screen, state, win, cast)
	}
}

// renderWindowContent renders the content area of a window (picture).
func (r *EbitenRenderer) renderWindowContent(screen *ebiten.Image, state *EngineState, win *Window, logThisFrame bool) {
	// Get window's picture
	pic := state.GetPicture(win.PictureID)
	if pic == nil {
		if logThisFrame && r.logger != nil {
			r.logger.LogDebug("renderWindowContent: window %d has no picture (picID=%d)", win.ID, win.PictureID)
		}
		return
	}

	if logThisFrame && r.logger != nil {
		r.logger.LogDebug("renderWindowContent: win=%d pic=%d picSize=(%dx%d) winPos=(%d,%d) winSize=(%dx%d) picOffset=(%d,%d)",
			win.ID, win.PictureID, pic.Width, pic.Height, win.X, win.Y, win.Width, win.Height, win.PicX, win.PicY)
	}

	// Convert picture to Ebiten image
	var ebitenPic *ebiten.Image
	switch img := pic.Image.(type) {
	case *image.RGBA:
		ebitenPic = ebiten.NewImageFromImage(img)
	default:
		// Convert to RGBA if needed
		bounds := pic.Image.Bounds()
		rgba := image.NewRGBA(bounds)
		draw.Draw(rgba, bounds, pic.Image, bounds.Min, draw.Src)
		ebitenPic = ebiten.NewImageFromImage(rgba)
	}

	// Content area starts at (win.X + BorderThickness, win.Y + TitleBarHeight + BorderThickness)
	contentX := win.X + BorderThickness
	contentY := win.Y + TitleBarHeight + BorderThickness

	// Window rectangle (content area in screen coordinates)
	winRect := image.Rect(contentX, contentY, contentX+win.Width, contentY+win.Height)

	// Image rectangle (where the full image would be drawn if positioned by PicX/PicY offsets)
	// PicX and PicY act as offsets relative to the content area's top-left
	// Negative offsets shift the image left/up, showing the center portion
	imgAbsX := contentX + win.PicX
	imgAbsY := contentY + win.PicY
	imgRect := image.Rect(imgAbsX, imgAbsY, imgAbsX+pic.Width, imgAbsY+pic.Height)

	if logThisFrame && r.logger != nil {
		r.logger.LogDebug("  contentPos=(%d,%d) winRect=%v imgRect=%v", contentX, contentY, winRect, imgRect)
	}

	// Calculate intersection: the visible part of the image
	drawRect := winRect.Intersect(imgRect)

	// If intersection is empty, nothing to draw
	if drawRect.Empty() {
		if logThisFrame && r.logger != nil {
			r.logger.LogDebug("  intersection is EMPTY - nothing to draw!")
		}
		return
	}

	// Calculate source rectangle in the picture
	// The top-left of the image is at (imgAbsX, imgAbsY)
	// The visible part starts at (drawRect.Min.X, drawRect.Min.Y)
	// So source coordinates are relative to the image origin
	srcX := drawRect.Min.X - imgAbsX
	srcY := drawRect.Min.Y - imgAbsY
	srcW := drawRect.Dx()
	srcH := drawRect.Dy()

	if logThisFrame && r.logger != nil {
		r.logger.LogDebug("  drawRect=%v srcRect=(%d,%d,%d,%d)", drawRect, srcX, srcY, srcW, srcH)
	}

	// Create subimage from the visible portion
	srcRect := image.Rect(srcX, srcY, srcX+srcW, srcY+srcH)
	subImg := ebitenPic.SubImage(srcRect).(*ebiten.Image)

	// Draw at the intersection point on screen
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(drawRect.Min.X), float64(drawRect.Min.Y))
	screen.DrawImage(subImg, opts)

	if logThisFrame && r.logger != nil {
		r.logger.LogDebug("  drew image at screen pos (%d,%d)", drawRect.Min.X, drawRect.Min.Y)
	}
}

// renderCast renders a single cast (sprite).
func (r *EbitenRenderer) renderCast(screen *ebiten.Image, state *EngineState, win *Window, cast *Cast) {
	// Get cast's picture
	pic := state.GetPicture(cast.PictureID)
	if pic == nil {
		return
	}

	// Convert picture to Ebiten image
	var ebitenPic *ebiten.Image
	switch img := pic.Image.(type) {
	case *image.RGBA:
		ebitenPic = ebiten.NewImageFromImage(img)
	default:
		// Convert to RGBA if needed
		bounds := pic.Image.Bounds()
		rgba := image.NewRGBA(bounds)
		draw.Draw(rgba, bounds, pic.Image, bounds.Min, draw.Src)
		ebitenPic = ebiten.NewImageFromImage(rgba)
	}

	// Clip the picture to the cast's source rectangle
	subImg := ebitenPic.SubImage(image.Rect(cast.SrcX, cast.SrcY, cast.SrcX+cast.Width, cast.SrcY+cast.Height)).(*ebiten.Image)

	// Content area starts at (win.X + BorderThickness, win.Y + TitleBarHeight + BorderThickness)
	contentX := win.X + BorderThickness
	contentY := win.Y + TitleBarHeight + BorderThickness

	// Create draw options for cast (position relative to content area)
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(contentX+cast.X), float64(contentY+cast.Y))

	// Draw cast
	screen.DrawImage(subImg, opts)
}

// Clear clears the screen with the specified color.
func (r *EbitenRenderer) Clear(colorValue uint32) {
	r.backgroundColor = color.RGBA{
		R: uint8((colorValue >> 16) & 0xFF),
		G: uint8((colorValue >> 8) & 0xFF),
		B: uint8(colorValue & 0xFF),
		A: 0xFF,
	}
}
