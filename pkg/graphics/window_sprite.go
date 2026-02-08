// Package graphics provides sprite-based rendering system.
package graphics

import (
	"image"
	"image/color"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
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
	fontFace      font.Face // キャプション描画用フォント
	mu            sync.RWMutex
}

// NewWindowSpriteManager は新しいWindowSpriteManagerを作成する
func NewWindowSpriteManager(sm *SpriteManager) *WindowSpriteManager {
	return &WindowSpriteManager{
		windowSprites: make(map[int]*WindowSprite),
		spriteManager: sm,
	}
}

// SetFontFace はキャプション描画用のフォントを設定する
func (wsm *WindowSpriteManager) SetFontFace(face font.Face) {
	wsm.mu.Lock()
	defer wsm.mu.Unlock()
	wsm.fontFace = face
}

// CreateWindowSprite はウインドウからWindowSpriteを作成する
// 要件 7.1: 指定サイズ・背景色のウインドウスプライトを作成できる
// 要件 14.3: Z順序の統一（ウインドウ間、ウインドウ内）
// 要件 4.1: ウインドウをRoot_Spriteとして扱う
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

	// ウインドウ装飾を描画（キャプションを含む）
	drawWindowDecorationOnImage(img, win.BgColor, winWidth, winHeight, borderThickness, titleBarHeight, win.Caption)

	// スプライトを作成
	sprite := wsm.spriteManager.CreateSprite(img)
	sprite.SetPosition(float64(win.X), float64(win.Y))
	sprite.SetVisible(win.Visible)

	// 子スプライトの座標オフセットを設定
	// コンテンツ領域オフセット + PicX/PicYオフセット（負の値）
	childOffsetX := float64(borderThickness) - float64(win.PicX)
	childOffsetY := float64(borderThickness+titleBarHeight) - float64(win.PicY)
	sprite.SetChildOffset(childOffsetX, childOffsetY)

	// 要件 4.1: ウインドウをRoot_Spriteとして扱う
	// 要件 1.3: Root_Spriteは単一要素のZ_Path（例: [0]）を持つ
	// ウインドウスプライトにZ_Pathを設定（ルートスプライトとして）
	sprite.SetZPath(NewZPath(win.ZOrder))

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

// Count は登録されているWindowSpriteの数を返す
// タスク 7.1: スプライト数の測定機能
func (wsm *WindowSpriteManager) Count() int {
	wsm.mu.RLock()
	defer wsm.mu.RUnlock()
	return len(wsm.windowSprites)
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

// GetPicID は関連付けられたピクチャーIDを返す
func (ws *WindowSprite) GetPicID() int {
	if ws.window == nil {
		return -1
	}
	return ws.window.PicID
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
// 要件 1.4: 子スプライトが作成されたとき、親のZ_Pathを継承し、自身のLocal_Z_Orderを追加する
func (ws *WindowSprite) AddChild(child *Sprite) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	child.SetParent(ws.sprite)
	ws.children = append(ws.children, child)

	// 子スプライトにZ_Pathが設定されていない場合、親のZ_Pathを継承して設定
	// 注意: 通常、子スプライトのZ_Pathは作成時（CastSprite, TextSprite等）に設定されるべき
	// このコードは、Z_Pathが設定されていない場合のフォールバックとして機能する
	if child.GetZPath() == nil && ws.sprite.GetZPath() != nil {
		// 子スプライトの数をLocal_Z_Orderとして使用（簡易的な実装）
		localZOrder := len(ws.children) - 1
		child.SetZPath(NewZPathFromParent(ws.sprite.GetZPath(), localZOrder))
	}
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
// 要件 4.4: ウインドウが前面に移動したとき、そのウインドウのZ_Pathを更新する
// 注意: このメソッドは互換性のために維持されています。
// 子スプライトのZ_Path更新が必要な場合は UpdateWindowZOrder() を使用してください。
func (ws *WindowSprite) UpdateZOrder(z int) {
	ws.window.ZOrder = z

	// 要件 4.4: ウインドウが前面に移動したとき、そのウインドウのZ_Pathを更新する
	ws.sprite.SetZPath(NewZPath(z))
}

// UpdatePicOffset はPicX/PicYオフセットを更新する
// MoveWinでPicX/PicYが変更された場合に呼び出す
func (ws *WindowSprite) UpdatePicOffset(picX, picY int) {
	ws.window.PicX = picX
	ws.window.PicY = picY

	// 子スプライトの座標オフセットを再計算
	childOffsetX := float64(ws.borderThickness) - float64(picX)
	childOffsetY := float64(ws.borderThickness+ws.titleBarHeight) - float64(picY)
	ws.sprite.SetChildOffset(childOffsetX, childOffsetY)
}

// UpdateWindowZOrder はウインドウのZ順序を更新し、子スプライトのZ_Pathも更新する
// 要件 4.3: ウインドウのZ順序変更時に、そのウインドウの子スプライトのZ_Pathを更新する
// 要件 4.4: ウインドウが前面に移動したとき、そのウインドウのZ_Pathを更新する
func (ws *WindowSprite) UpdateWindowZOrder(newZOrder int, sm *SpriteManager) {
	ws.window.ZOrder = newZOrder

	// ウインドウスプライトのZ_Pathを更新
	ws.sprite.SetZPath(NewZPath(newZOrder))

	// 子スプライトのZ_Pathを再帰的に更新
	sm.UpdateChildrenZPaths(ws.sprite)

	sm.MarkNeedSort()
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
		drawWindowDecorationOnImage(img, ws.window.BgColor, winWidth, winHeight, ws.borderThickness, ws.titleBarHeight, ws.window.Caption)
	}
}

// drawWindowDecorationOnImage はウインドウ装飾を画像に描画する
// Windows 3.1風のウインドウ装飾を描画
// 要件 11.2: ピクチャー画像の直接描画は廃止（スプライトシステムで描画）
// ピクチャー画像はPictureSpriteとして別途描画される
func drawWindowDecorationOnImage(img *ebiten.Image, bgColor color.Color, winWidth, winHeight, borderThickness, titleBarHeight int, caption string) {
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
	vector.FillRect(img, 0, 0, totalW, totalH, borderColor, false)

	// 2. 3D枠線効果を描画
	// 上と左の縁（ハイライト）
	vector.StrokeLine(img, 0, 0, totalW, 0, 1, highlightColor, false)
	vector.StrokeLine(img, 0, 0, 0, totalH, 1, highlightColor, false)

	// 下と右の縁（影）
	vector.StrokeLine(img, 0, totalH-1, totalW, totalH-1, 1, shadowColor, false)
	vector.StrokeLine(img, totalW-1, 0, totalW-1, totalH, 1, shadowColor, false)

	// 3. タイトルバーを描画（濃い青）
	vector.FillRect(img,
		float32(borderThickness),
		float32(borderThickness),
		float32(winWidth), float32(titleBarHeight),
		titleBarColor, false)

	// 3.1 キャプションを描画（白色テキスト）
	if caption != "" {
		// タイトルバーの中央にキャプションを描画
		textX := borderThickness + 4 // 左端から少しオフセット
		textY := borderThickness + 2 // タイトルバー内で少し下にオフセット
		drawCaptionText(img, caption, textX, textY)
	}

	// 4. コンテンツ領域を描画
	contentX := borderThickness
	contentY := borderThickness + titleBarHeight

	// 4.1 背景色を描画
	if bgColor != nil {
		vector.FillRect(img,
			float32(contentX), float32(contentY),
			float32(winWidth), float32(winHeight),
			bgColor, false)
	}

	// 要件 11.2: ピクチャー画像はスプライトシステムで描画する
	// ここではウィンドウ装飾（枠、タイトルバー、背景色）のみを描画
	// ピクチャー画像はOpenWin時にPictureSpriteとして作成され、
	// drawLayersForWindow()で描画される
}

// drawCaptionText はキャプションテキストをebiten.Imageに描画する
// 日本語を含むテキストを描画するため、font.Drawerを使用する
func drawCaptionText(img *ebiten.Image, caption string, x, y int) {
	// ebiten.Imageのサイズを取得
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// 一時的なRGBA画像を作成
	rgba := image.NewRGBA(image.Rect(0, 0, width, height))

	// 白色でテキストを描画
	textColor := color.RGBA{255, 255, 255, 255}

	// フォントフェイスを取得（日本語フォントがあれば使用、なければbasicfont）
	face := getCaptionFont()

	drawer := &font.Drawer{
		Dst:  rgba,
		Src:  image.NewUniform(textColor),
		Face: face,
		Dot:  fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y + 12)}, // ベースライン調整
	}
	drawer.DrawString(caption)

	// RGBA画像をebiten.Imageに描画
	textImg := ebiten.NewImageFromImage(rgba)
	op := &ebiten.DrawImageOptions{}
	img.DrawImage(textImg, op)
}

// captionFont はキャプション用のフォントをキャッシュする
var (
	captionFont     font.Face
	captionFontOnce sync.Once
)

// getCaptionFont はキャプション用のフォントを返す
// 日本語フォントが利用可能であればそれを使用し、なければbasicfontを使用する
func getCaptionFont() font.Face {
	captionFontOnce.Do(func() {
		// TextRendererを使用して日本語フォントを読み込む
		// TextRendererは内部でシステムフォントを検索し、日本語フォントを見つける
		tr := NewTextRenderer()
		// SetFontで日本語フォントを読み込む（サイズ12）
		_ = tr.SetFont("ＭＳ ゴシック", 12)

		// TextRendererのフォントフェイスを取得
		face := tr.GetFace()
		if face != nil && face != basicfont.Face7x13 {
			captionFont = face
		} else {
			// フォールバック: basicfont（日本語は表示できない）
			captionFont = basicfont.Face7x13
		}
	})
	return captionFont
}
