// Package graphics provides sprite-based rendering system.
package graphics

import (
	"image/color"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// WindowSprite はウインドウとスプライトを組み合わせたラッパー構造体
// 要件 7.1: 指定サイズ・背景色のウインドウスプライトを作成できる
// 要件 7.2: ウインドウスプライトを親として子スプライトを追加できる
// 要件 7.3: ウインドウが閉じられたときにウインドウとその子スプライトを削除する
type WindowSprite struct {
	window *Window // 元のウインドウ情報
	sprite *Sprite // スプライト（ウインドウ全体の画像を保持）

	// ウインドウ装飾の定数
	borderThickness int
	titleBarHeight  int

	// 子スプライト管理
	children []*Sprite
	mu       sync.RWMutex
}

// WindowSpriteManager はWindowSpriteを管理する
type WindowSpriteManager struct {
	windowSprites map[int]*WindowSprite // windowID -> WindowSprite
	spriteManager *SpriteManager
	mu            sync.RWMutex
}

// NewWindowSpriteManager は新しいWindowSpriteManagerを作成する
func NewWindowSpriteManager(sm *SpriteManager) *WindowSpriteManager {
	return &WindowSpriteManager{
		windowSprites: make(map[int]*WindowSprite),
		spriteManager: sm,
	}
}

// CreateWindowSprite はウインドウからWindowSpriteを作成する
// 要件 7.1: 指定サイズ・背景色のウインドウスプライトを作成できる
// 要件 14.3: Z順序の統一（ウインドウ間、ウインドウ内）
func (wsm *WindowSpriteManager) CreateWindowSprite(win *Window, pic *Picture) *WindowSprite {
	wsm.mu.Lock()
	defer wsm.mu.Unlock()

	const (
		borderThickness = 4
		titleBarHeight  = 20
	)

	// ウインドウの実際のサイズを計算
	winWidth := pic.Width
	winHeight := pic.Height
	if win.Width > 0 {
		winWidth = win.Width
	}
	if win.Height > 0 {
		winHeight = win.Height
	}

	// 全体のウインドウサイズ（装飾を含む）
	totalW := winWidth + borderThickness*2
	totalH := winHeight + borderThickness*2 + titleBarHeight

	// スプライト用の画像を作成
	img := ebiten.NewImage(totalW, totalH)

	// ウインドウ装飾を描画
	drawWindowDecorationOnImage(img, win, pic, winWidth, winHeight, borderThickness, titleBarHeight)

	// スプライトを作成
	sprite := wsm.spriteManager.CreateSprite(img)
	sprite.SetPosition(float64(win.X), float64(win.Y))
	// 要件 14.3: グローバルZ順序を使用
	// ウインドウスプライト自体はウインドウ範囲の先頭に配置
	globalZOrder := CalculateGlobalZOrder(win.ZOrder, ZOrderWindowBase)
	sprite.SetZOrder(globalZOrder)
	sprite.SetVisible(win.Visible)

	ws := &WindowSprite{
		window:          win,
		sprite:          sprite,
		borderThickness: borderThickness,
		titleBarHeight:  titleBarHeight,
		children:        make([]*Sprite, 0),
	}

	wsm.windowSprites[win.ID] = ws
	return ws
}

// GetWindowSprite はウインドウIDからWindowSpriteを取得する
func (wsm *WindowSpriteManager) GetWindowSprite(winID int) *WindowSprite {
	wsm.mu.RLock()
	defer wsm.mu.RUnlock()
	return wsm.windowSprites[winID]
}

// GetWindowSpriteSprite はウインドウIDからWindowSpriteの基盤スプライトを取得する
// 子スプライトの親として使用する
func (wsm *WindowSpriteManager) GetWindowSpriteSprite(winID int) *Sprite {
	wsm.mu.RLock()
	defer wsm.mu.RUnlock()
	ws := wsm.windowSprites[winID]
	if ws == nil {
		return nil
	}
	return ws.sprite
}

// RemoveWindowSprite はWindowSpriteを削除する
// 要件 7.3: ウインドウが閉じられたときにウインドウとその子スプライトを削除する
func (wsm *WindowSpriteManager) RemoveWindowSprite(winID int) {
	wsm.mu.Lock()
	defer wsm.mu.Unlock()

	ws, exists := wsm.windowSprites[winID]
	if !exists {
		return
	}

	// 子スプライトを削除
	for _, child := range ws.children {
		wsm.spriteManager.RemoveSprite(child.ID())
	}

	// ウインドウスプライト自体を削除
	wsm.spriteManager.RemoveSprite(ws.sprite.ID())

	delete(wsm.windowSprites, winID)
}

// Clear はすべてのWindowSpriteを削除する
func (wsm *WindowSpriteManager) Clear() {
	wsm.mu.Lock()
	defer wsm.mu.Unlock()

	for winID, ws := range wsm.windowSprites {
		// 子スプライトを削除
		for _, child := range ws.children {
			wsm.spriteManager.RemoveSprite(child.ID())
		}
		// ウインドウスプライト自体を削除
		wsm.spriteManager.RemoveSprite(ws.sprite.ID())
		delete(wsm.windowSprites, winID)
	}
}

// WindowSprite methods

// GetWindow は元のウインドウを返す
func (ws *WindowSprite) GetWindow() *Window {
	return ws.window
}

// GetSprite はスプライトを返す
func (ws *WindowSprite) GetSprite() *Sprite {
	return ws.sprite
}

// GetContentOffset はコンテンツ領域のオフセットを返す
func (ws *WindowSprite) GetContentOffset() (int, int) {
	return ws.borderThickness, ws.borderThickness + ws.titleBarHeight
}

// GetContentSprite はコンテンツ領域用の仮想スプライトを返す
// 子スプライトはこのスプライトを親として設定することで、
// コンテンツ領域内の相対位置で描画される
// 注意: この関数は将来の完全移行のための準備として実装されている
func (ws *WindowSprite) GetContentSprite() *Sprite {
	return ws.sprite
}

// GetPicOffset はピクチャーのオフセット（PicX, PicY）を返す
func (ws *WindowSprite) GetPicOffset() (int, int) {
	if ws.window == nil {
		return 0, 0
	}
	return ws.window.PicX, ws.window.PicY
}

// AddChild は子スプライトを追加する
// 要件 7.2: ウインドウスプライトを親として子スプライトを追加できる
func (ws *WindowSprite) AddChild(child *Sprite) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	child.SetParent(ws.sprite)
	ws.children = append(ws.children, child)
}

// RemoveChild は子スプライトを削除する
func (ws *WindowSprite) RemoveChild(childID int) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	for i, child := range ws.children {
		if child.ID() == childID {
			child.SetParent(nil)
			ws.children = append(ws.children[:i], ws.children[i+1:]...)
			return
		}
	}
}

// GetChildren は子スプライトのリストを返す
func (ws *WindowSprite) GetChildren() []*Sprite {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	result := make([]*Sprite, len(ws.children))
	copy(result, ws.children)
	return result
}

// UpdatePosition はウインドウの位置を更新する
func (ws *WindowSprite) UpdatePosition(x, y int) {
	ws.window.X = x
	ws.window.Y = y
	ws.sprite.SetPosition(float64(x), float64(y))
}

// UpdateZOrder はZ順序を更新する
// 要件 14.3: グローバルZ順序を使用
func (ws *WindowSprite) UpdateZOrder(z int) {
	ws.window.ZOrder = z
	// グローバルZ順序を計算
	globalZOrder := CalculateGlobalZOrder(z, ZOrderWindowBase)
	ws.sprite.SetZOrder(globalZOrder)
}

// UpdateVisible は可視性を更新する
func (ws *WindowSprite) UpdateVisible(visible bool) {
	ws.window.Visible = visible
	ws.sprite.SetVisible(visible)
}

// RedrawDecoration はウインドウ装飾を再描画する
func (ws *WindowSprite) RedrawDecoration(pic *Picture) {
	winWidth := pic.Width
	winHeight := pic.Height
	if ws.window.Width > 0 {
		winWidth = ws.window.Width
	}
	if ws.window.Height > 0 {
		winHeight = ws.window.Height
	}

	// 画像をクリアして再描画
	img := ws.sprite.Image()
	if img != nil {
		img.Clear()
		drawWindowDecorationOnImage(img, ws.window, pic, winWidth, winHeight, ws.borderThickness, ws.titleBarHeight)
	}
}

// drawWindowDecorationOnImage はウインドウ装飾を画像に描画する
// Windows 3.1風のウインドウ装飾を描画
func drawWindowDecorationOnImage(img *ebiten.Image, win *Window, pic *Picture, winWidth, winHeight, borderThickness, titleBarHeight int) {
	// Windows 3.1風の色
	var (
		titleBarColor  = color.RGBA{0, 0, 128, 255}     // 濃い青
		borderColor    = color.RGBA{192, 192, 192, 255} // グレー
		highlightColor = color.RGBA{255, 255, 255, 255} // 白（立体効果のハイライト）
		shadowColor    = color.RGBA{0, 0, 0, 255}       // 黒（立体効果の影）
	)

	totalW := float32(winWidth + borderThickness*2)
	totalH := float32(winHeight + borderThickness*2 + titleBarHeight)

	// 1. ウィンドウフレームの背景を描画（グレー）
	vector.DrawFilledRect(img, 0, 0, totalW, totalH, borderColor, false)

	// 2. 3D枠線効果を描画
	// 上と左の縁（ハイライト）
	vector.StrokeLine(img, 0, 0, totalW, 0, 1, highlightColor, false)
	vector.StrokeLine(img, 0, 0, 0, totalH, 1, highlightColor, false)

	// 下と右の縁（影）
	vector.StrokeLine(img, 0, totalH-1, totalW, totalH-1, 1, shadowColor, false)
	vector.StrokeLine(img, totalW-1, 0, totalW-1, totalH, 1, shadowColor, false)

	// 3. タイトルバーを描画（濃い青）
	vector.DrawFilledRect(img,
		float32(borderThickness),
		float32(borderThickness),
		float32(winWidth), float32(titleBarHeight),
		titleBarColor, false)

	// 4. コンテンツ領域を描画
	contentX := borderThickness
	contentY := borderThickness + titleBarHeight

	// 4.1 背景色を描画
	if win.BgColor != nil {
		vector.DrawFilledRect(img,
			float32(contentX), float32(contentY),
			float32(winWidth), float32(winHeight),
			win.BgColor, false)
	}

	// 4.2 ピクチャーを描画
	if pic != nil && pic.Image != nil {
		// PicX/PicYオフセットを考慮した描画位置
		imgX := contentX - win.PicX
		imgY := contentY - win.PicY

		// クリッピング計算
		srcX := 0
		srcY := 0
		srcW := pic.Width
		srcH := pic.Height

		// 描画位置がコンテンツ領域外の場合の調整
		if imgX < contentX {
			srcX = contentX - imgX
			srcW -= srcX
			imgX = contentX
		}
		if imgY < contentY {
			srcY = contentY - imgY
			srcH -= srcY
			imgY = contentY
		}

		// コンテンツ領域を超える部分のクリッピング
		if imgX+srcW > contentX+winWidth {
			srcW = contentX + winWidth - imgX
		}
		if imgY+srcH > contentY+winHeight {
			srcH = contentY + winHeight - imgY
		}

		// 有効な領域がある場合のみ描画
		if srcW > 0 && srcH > 0 {
			subImg := pic.Image.SubImage(pic.Image.Bounds()).(*ebiten.Image)
			if srcX > 0 || srcY > 0 || srcW < pic.Width || srcH < pic.Height {
				// サブイメージを切り出す
				subRect := pic.Image.Bounds()
				subRect.Min.X = srcX
				subRect.Min.Y = srcY
				subRect.Max.X = srcX + srcW
				subRect.Max.Y = srcY + srcH
				subImg = pic.Image.SubImage(subRect).(*ebiten.Image)
			}

			opts := &ebiten.DrawImageOptions{}
			opts.GeoM.Translate(float64(imgX), float64(imgY))
			img.DrawImage(subImg, opts)
		}
	}
}
