// Package sprite provides sprite-based rendering system with slice-based draw ordering.
package sprite

import (
	"image/color"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// WindowSpriteConfig はウインドウスプライト作成時の設定
// 要件 9.4: 数値によるZ順序管理を使用しない（スライスベースの描画順序）
type WindowSpriteConfig struct {
	ID      int
	X, Y    int
	Width   int
	Height  int
	PicX    int
	PicY    int
	BgColor color.Color
	Visible bool
}

// WindowSprite はウインドウとスプライトを組み合わせたラッパー構造体
// 要件 4.1: 指定サイズ・背景色のウインドウスプライトを作成できる
// 要件 4.2: ウインドウスプライトを親として子スプライトを追加できる
// 要件 4.3: ウインドウが閉じられたときにウインドウとその子スプライトを削除する
type WindowSprite struct {
	windowID int     // ウインドウID
	sprite   *Sprite // スプライト（ウインドウ全体の画像を保持）

	// ウインドウ情報
	x, y    int
	width   int
	height  int
	picX    int
	picY    int
	bgColor color.Color
	visible bool

	// ウインドウ装飾の定数
	borderThickness int
	titleBarHeight  int

	mu sync.RWMutex
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
// 要件 4.1: 指定サイズ・背景色のウインドウスプライトを作成できる
// 要件 11.1: ウインドウをルートスプライトとして扱う
func (wsm *WindowSpriteManager) CreateWindowSprite(config WindowSpriteConfig, picWidth, picHeight int) *WindowSprite {
	wsm.mu.Lock()
	defer wsm.mu.Unlock()

	const (
		borderThickness = 4
		titleBarHeight  = 20
	)

	// ウインドウの実際のサイズを計算
	winWidth := picWidth
	winHeight := picHeight
	if config.Width > 0 {
		winWidth = config.Width
	}
	if config.Height > 0 {
		winHeight = config.Height
	}

	// 全体のウインドウサイズ（装飾を含む）
	totalW := winWidth + borderThickness*2
	totalH := winHeight + borderThickness*2 + titleBarHeight

	// スプライト用の画像を作成
	img := ebiten.NewImage(totalW, totalH)

	// ウインドウ装飾を描画
	drawWindowDecorationOnImage(img, config.BgColor, winWidth, winHeight, borderThickness, titleBarHeight)

	// スプライトを作成（ルートスプライトとして）
	sprite := wsm.spriteManager.CreateRootSprite(img)
	sprite.SetPosition(float64(config.X), float64(config.Y))
	sprite.SetVisible(config.Visible)

	ws := &WindowSprite{
		windowID:        config.ID,
		sprite:          sprite,
		x:               config.X,
		y:               config.Y,
		width:           winWidth,
		height:          winHeight,
		picX:            config.PicX,
		picY:            config.PicY,
		bgColor:         config.BgColor,
		visible:         config.Visible,
		borderThickness: borderThickness,
		titleBarHeight:  titleBarHeight,
	}

	wsm.windowSprites[config.ID] = ws
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
// 要件 4.3: ウインドウが閉じられたときにウインドウとその子スプライトを削除する
func (wsm *WindowSpriteManager) RemoveWindowSprite(winID int) {
	wsm.mu.Lock()
	defer wsm.mu.Unlock()

	ws, exists := wsm.windowSprites[winID]
	if !exists {
		return
	}

	// 子スプライトを削除
	for _, child := range ws.sprite.GetChildren() {
		wsm.spriteManager.DeleteSprite(child.ID())
	}

	// ウインドウスプライト自体を削除
	wsm.spriteManager.DeleteSprite(ws.sprite.ID())

	delete(wsm.windowSprites, winID)
}

// Clear はすべてのWindowSpriteを削除する
func (wsm *WindowSpriteManager) Clear() {
	wsm.mu.Lock()
	defer wsm.mu.Unlock()

	for winID, ws := range wsm.windowSprites {
		// 子スプライトを削除
		for _, child := range ws.sprite.GetChildren() {
			wsm.spriteManager.DeleteSprite(child.ID())
		}
		// ウインドウスプライト自体を削除
		wsm.spriteManager.DeleteSprite(ws.sprite.ID())
		delete(wsm.windowSprites, winID)
	}
}

// WindowSprite methods

// GetWindowID はウインドウIDを返す
func (ws *WindowSprite) GetWindowID() int {
	return ws.windowID
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
func (ws *WindowSprite) GetContentSprite() *Sprite {
	return ws.sprite
}

// GetPicOffset はピクチャーのオフセット（PicX, PicY）を返す
func (ws *WindowSprite) GetPicOffset() (int, int) {
	return ws.picX, ws.picY
}

// AddChild は子スプライトを追加する
// 要件 4.2: ウインドウスプライトを親として子スプライトを追加できる
func (ws *WindowSprite) AddChild(child *Sprite) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	ws.sprite.AddChild(child)
}

// RemoveChild は子スプライトを削除する
func (ws *WindowSprite) RemoveChild(childID int) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	ws.sprite.RemoveChild(childID)
}

// GetChildren は子スプライトのリストを返す
func (ws *WindowSprite) GetChildren() []*Sprite {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	return ws.sprite.GetChildren()
}

// UpdatePosition はウインドウの位置を更新する
func (ws *WindowSprite) UpdatePosition(x, y int) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	ws.x = x
	ws.y = y
	ws.sprite.SetPosition(float64(x), float64(y))
}

// UpdateVisible は可視性を更新する
func (ws *WindowSprite) UpdateVisible(visible bool) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	ws.visible = visible
	ws.sprite.SetVisible(visible)
}

// RedrawDecoration はウインドウ装飾を再描画する
func (ws *WindowSprite) RedrawDecoration() {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	// 画像をクリアして再描画
	img := ws.sprite.Image()
	if img != nil {
		img.Clear()
		drawWindowDecorationOnImage(img, ws.bgColor, ws.width, ws.height, ws.borderThickness, ws.titleBarHeight)
	}
}

// drawWindowDecorationOnImage はウインドウ装飾を画像に描画する
// Windows 3.1風のウインドウ装飾を描画
func drawWindowDecorationOnImage(img *ebiten.Image, bgColor color.Color, winWidth, winHeight, borderThickness, titleBarHeight int) {
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
	if bgColor != nil {
		vector.DrawFilledRect(img,
			float32(contentX), float32(contentY),
			float32(winWidth), float32(winHeight),
			bgColor, false)
	}
}
