package engine

import (
	"fmt"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// EbitenRenderer implements Renderer using Ebitengine
type EbitenRenderer struct {
	tickCount int
}

// NewEbitenRenderer creates a new Ebitengine-based renderer
func NewEbitenRenderer() *EbitenRenderer {
	return &EbitenRenderer{
		tickCount: 0,
	}
}

// RenderFrame renders the complete frame to the screen
func (r *EbitenRenderer) RenderFrame(screen *ebiten.Image, state *EngineState) {
	// Lock state for reading during rendering
	state.renderMutex.Lock()
	defer state.renderMutex.Unlock()

	r.tickCount++

	// Virtual desktop background color (teal)
	screen.Fill(color.RGBA{0x1F, 0x7E, 0x7F, 0xff})

	// Render all windows in order
	for _, winID := range state.windowOrder {
		win, ok := state.windows[winID]
		if !ok || !win.Visible {
			continue
		}

		pic, ok := state.pictures[win.Picture]
		if !ok {
			continue
		}

		// Debug: log which window is being drawn (very infrequent)
		if r.tickCount == 600 {
			fmt.Printf("Drawing Window ID=%d (Pic=%d) at (%d,%d)\n", winID, win.Picture, win.X, win.Y)
		}

		// Render window with decorations
		r.renderWindow(screen, win, pic, state)
	}
}

// renderWindow renders a single window with decorations
func (r *EbitenRenderer) renderWindow(screen *ebiten.Image, win *Window, pic *Picture, state *EngineState) {
	// Window Geometry using Global Constants
	winX := float32(win.X)
	winY := float32(win.Y)
	winW := float32(win.W)
	winH := float32(win.H)

	totalW := winW + float32(BorderThickness*2)
	totalH := winH + float32(BorderThickness*2) + float32(TitleBarHeight)

	// Draw Window Frame (Windows 3.1 Style)
	// 1. Main background (Gray)
	vector.DrawFilledRect(screen, winX-float32(BorderThickness), winY-float32(TitleBarHeight)-float32(BorderThickness), totalW, totalH, color.RGBA{192, 192, 192, 255}, true)

	// 2. Borders (3D effect)
	// Top/Left Highlight (White)
	vector.StrokeLine(screen, winX-float32(BorderThickness), winY-float32(TitleBarHeight)-float32(BorderThickness), winX+totalW-float32(BorderThickness), winY-float32(TitleBarHeight)-float32(BorderThickness), 1, color.White, true)
	vector.StrokeLine(screen, winX-float32(BorderThickness), winY-float32(TitleBarHeight)-float32(BorderThickness), winX-float32(BorderThickness), winY+winH+float32(BorderThickness), 1, color.White, true)

	// Bottom/Right Shadow (Black/Dark Gray)
	vector.StrokeLine(screen, winX-float32(BorderThickness), winY+winH+float32(BorderThickness), winX+winW+float32(BorderThickness), winY+winH+float32(BorderThickness), 1, color.Black, true)
	vector.StrokeLine(screen, winX+winW+float32(BorderThickness), winY-float32(TitleBarHeight)-float32(BorderThickness), winX+winW+float32(BorderThickness), winY+winH+float32(BorderThickness), 1, color.Black, true)

	// 3. Title Bar (Blue)
	vector.DrawFilledRect(screen, winX, winY-float32(TitleBarHeight), winW, float32(TitleBarHeight), color.RGBA{0, 0, 128, 255}, true)

	// 4. Title Text (White)
	if state.currentFont != nil && win.Title != "" {
		text.Draw(screen, win.Title, state.currentFont, int(winX)+4, int(winY)-6, color.White)
	}

	// Draw window content
	r.renderWindowContent(screen, win, pic, state)
}

// renderWindowContent renders the content area of a window
func (r *EbitenRenderer) renderWindowContent(screen *ebiten.Image, win *Window, pic *Picture, state *EngineState) {
	winX := float32(win.X)
	winY := float32(win.Y)
	winW := float32(win.W)
	winH := float32(win.H)

	// 1. Draw Window Background (solid color)
	bgCol := win.Color
	if bgCol == nil {
		bgCol = color.White
	}
	vector.DrawFilledRect(screen, winX, winY, winW, winH, bgCol, true)

	// 2. Draw Image with Clipping
	// Calculate the geometric intersection of the Window and the Image (positioned by offsets).

	// Window Rectangle (Absolute Screen Coords)
	winRect := image.Rect(win.X, win.Y, win.X+win.W, win.Y+win.H)

	// Image Rectangle (Absolute Screen Coords if full image were drawn)
	// win.SrcX, win.SrcY act as offsets relative to the window's top-left (win.X, win.Y)
	imgAbsX := win.X + win.SrcX
	imgAbsY := win.Y + win.SrcY
	picW, picH := pic.Width, pic.Height
	imgRect := image.Rect(imgAbsX, imgAbsY, imgAbsX+picW, imgAbsY+picH)

	// Intersection: The visible part of the image on screen
	drawRect := winRect.Intersect(imgRect)

	// DEBUG: Print image drawing info (very infrequent)
	if r.tickCount%600 == 0 || (drawRect.Empty() && r.tickCount%60 == 0) {
		fmt.Printf("DEBUG: WinID=%d Pic=%d ImgRect=%v DrawRect=%v Empty=%v\n",
			win.ID, win.Picture, imgRect, drawRect, drawRect.Empty())
	}

	// Debug base_pic (usually 25)
	if win.Picture == 25 && r.tickCount%60 == 0 {
		fmt.Printf("DEBUG: Drawing Window with BasePic (P25). DrawRect Empty? %v. SrcX=%d, SrcY=%d\n", drawRect.Empty(), win.SrcX, win.SrcY)
	}

	// If intersection is empty, the image is not visible in this window
	if !drawRect.Empty() {
		// Calculate Source Rectangle in the texture (SubImage)
		// The top-left of the image is at (imgAbsX, imgAbsY).
		// The visible part starts at (drawRect.Min.X, drawRect.Min.Y).
		// So source X = drawRect.Min.X - imgAbsX.
		srcX := drawRect.Min.X - imgAbsX
		srcY := drawRect.Min.Y - imgAbsY
		srcW := drawRect.Dx()
		srcH := drawRect.Dy()

		srcRect := image.Rect(srcX, srcY, srcX+srcW, srcY+srcH)
		subImg := pic.Image.SubImage(srcRect).(*ebiten.Image)

		// Draw SubImage at the intersection point on screen
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(float64(drawRect.Min.X), float64(drawRect.Min.Y))
		screen.DrawImage(subImg, opts)

		// DEBUG: Show Picture ID if debug overlay enabled
		if debugLevel >= 2 && state.currentFont != nil {
			picLabel := fmt.Sprintf("P%d", win.Picture)
			text.Draw(screen, picLabel, state.currentFont, drawRect.Min.X+5, drawRect.Min.Y+15, color.RGBA{0, 255, 0, 255})
		}
	}
}

// MockRenderer is a no-op renderer for testing
type MockRenderer struct {
	RenderCount int
	LastState   *EngineState
}

// NewMockRenderer creates a new mock renderer for testing
func NewMockRenderer() *MockRenderer {
	return &MockRenderer{
		RenderCount: 0,
		LastState:   nil,
	}
}

// RenderFrame records the render call but doesn't actually render
func (m *MockRenderer) RenderFrame(screen *ebiten.Image, state *EngineState) {
	m.RenderCount++
	m.LastState = state
	// No actual rendering - this is for headless testing
}
