// Package sprite provides sprite-based rendering system with slice-based draw ordering.
package sprite

import (
	"image"
	"image/color"
	"image/draw"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// TextSpriteOptions はテキストスプライト作成のオプション
type TextSpriteOptions struct {
	// Text は描画するテキスト
	Text string
	// TextColor はテキストの色
	TextColor color.Color
	// Face はフォントフェイス
	Face font.Face
	// BgColor は差分抽出用の背景色
	BgColor color.Color
	// Width は画像の幅（0の場合は自動計算）
	Width int
	// Height は画像の高さ（0の場合は自動計算）
	Height int
	// X はテキストのX座標
	X int
	// Y はテキストのベースラインY座標
	Y int
}

// CreateTextSpriteImage は差分抽出方式でテキストスプライト用の画像を作成する
// アンチエイリアスの影響を除去し、透過画像を返す
// 要件 7.1: 背景色の上にテキストを描画し、差分を抽出する
// 要件 7.2: アンチエイリアスの影響を除去する
func CreateTextSpriteImage(opts TextSpriteOptions) *image.RGBA {
	if opts.Face == nil || opts.Text == "" {
		return nil
	}

	// サイズの自動計算
	width := opts.Width
	height := opts.Height
	if width == 0 || height == 0 {
		bounds := measureText(opts.Face, opts.Text)
		if width == 0 {
			width = bounds.Dx() + opts.X + 10 // 余白を追加
		}
		if height == 0 {
			height = bounds.Dy() + 10 // 余白を追加
		}
	}

	if width <= 0 || height <= 0 {
		return nil
	}

	// 1. 背景色で塗りつぶした画像を作成
	bgImg := image.NewRGBA(image.Rect(0, 0, width, height))
	bgColor := opts.BgColor
	if bgColor == nil {
		bgColor = color.White
	}
	draw.Draw(bgImg, bgImg.Bounds(), image.NewUniform(bgColor), image.Point{}, draw.Src)

	// 2. 背景のコピーを保持
	bgCopy := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(bgCopy, bgCopy.Bounds(), bgImg, image.Point{}, draw.Src)

	// 3. テキストを描画
	textColor := opts.TextColor
	if textColor == nil {
		textColor = color.Black
	}
	drawer := &font.Drawer{
		Dst:  bgImg,
		Src:  image.NewUniform(textColor),
		Face: opts.Face,
		Dot:  fixed.Point26_6{X: fixed.I(opts.X), Y: fixed.I(opts.Y)},
	}
	drawer.DrawString(opts.Text)

	// 4. 差分を抽出（背景と異なるピクセルのみを残す）
	result := image.NewRGBA(image.Rect(0, 0, width, height))
	extractDifference(bgCopy, bgImg, result)

	return result
}

// extractDifference は2つの画像の差分を抽出し、結果画像に書き込む
// 背景と同じピクセルは透明に、異なるピクセルはそのまま残す
func extractDifference(bg, text, result *image.RGBA) {
	bounds := bg.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			bgPixel := bg.At(x, y)
			textPixel := text.At(x, y)

			bgR, bgG, bgB, _ := bgPixel.RGBA()
			txR, txG, txB, _ := textPixel.RGBA()

			if bgR != txR || bgG != txG || bgB != txB {
				// 差分があるピクセルはそのまま残す
				result.Set(x, y, textPixel)
			} else {
				// 背景と同じピクセルは透明にする
				result.Set(x, y, color.RGBA{0, 0, 0, 0})
			}
		}
	}
}

// measureText はテキストの境界ボックスを計算する
func measureText(face font.Face, text string) image.Rectangle {
	bounds, _ := font.BoundString(face, text)
	return image.Rect(
		bounds.Min.X.Floor(),
		bounds.Min.Y.Floor(),
		bounds.Max.X.Ceil(),
		bounds.Max.Y.Ceil(),
	)
}

// getFontHeight はフォントの高さを取得する
func getFontHeight(face font.Face) int {
	if face == nil {
		return 13 // デフォルト値
	}
	metrics := face.Metrics()
	return (metrics.Ascent + metrics.Descent).Ceil()
}

// TextSprite はテキストとスプライトを組み合わせたラッパー構造体
// 要件 7.1: 背景色の上にテキストを描画し、差分を抽出する
// 要件 7.2: アンチエイリアスの影響を除去する
type TextSprite struct {
	sprite *Sprite // 基盤となるスプライト

	// テキスト情報
	picID int    // 描画先ピクチャーID
	text  string // テキスト内容
	x, y  int    // 描画位置

	// テキスト設定
	textColor color.Color // テキスト色
	bgColor   color.Color // 背景色（差分抽出用）
	face      font.Face   // フォントフェイス

	mu sync.RWMutex
}

// TextSpriteManager はTextSpriteを管理する
type TextSpriteManager struct {
	textSprites   map[int][]*TextSprite // picID -> TextSprites（同じピクチャに複数のテキストがある場合）
	spriteManager *SpriteManager
	mu            sync.RWMutex
	nextID        int // 内部ID管理
}

// NewTextSpriteManager は新しいTextSpriteManagerを作成する
func NewTextSpriteManager(sm *SpriteManager) *TextSpriteManager {
	return &TextSpriteManager{
		textSprites:   make(map[int][]*TextSprite),
		spriteManager: sm,
		nextID:        1,
	}
}

// CreateTextSprite はテキストからTextSpriteを作成する
// 要件 7.1: 背景色の上にテキストを描画し、差分を抽出する
// 要件 7.2: アンチエイリアスの影響を除去する
func (tsm *TextSpriteManager) CreateTextSprite(
	picID int,
	x, y int,
	text string,
	textColor, bgColor color.Color,
	face font.Face,
	parent *Sprite,
) *TextSprite {
	tsm.mu.Lock()
	defer tsm.mu.Unlock()

	if text == "" || face == nil {
		return nil
	}

	// テキストスプライト用の画像を作成（差分抽出方式）
	opts := TextSpriteOptions{
		Text:      text,
		TextColor: textColor,
		Face:      face,
		BgColor:   bgColor,
		X:         0,
		Y:         getFontHeight(face),
	}

	img := CreateTextSpriteImage(opts)
	if img == nil {
		return nil
	}

	// スプライトを作成
	sprite := tsm.spriteManager.CreateSprite(ebiten.NewImageFromImage(img), parent)
	sprite.SetPosition(float64(x), float64(y))
	sprite.SetVisible(true)

	ts := &TextSprite{
		sprite:    sprite,
		picID:     picID,
		text:      text,
		x:         x,
		y:         y,
		textColor: textColor,
		bgColor:   bgColor,
		face:      face,
	}

	// ピクチャIDごとにスプライトを管理
	tsm.textSprites[picID] = append(tsm.textSprites[picID], ts)
	tsm.nextID++

	return ts
}

// GetTextSprites はピクチャIDに関連するすべてのTextSpriteを取得する
func (tsm *TextSpriteManager) GetTextSprites(picID int) []*TextSprite {
	tsm.mu.RLock()
	defer tsm.mu.RUnlock()

	sprites := tsm.textSprites[picID]
	if sprites == nil {
		return nil
	}

	// コピーを返す
	result := make([]*TextSprite, len(sprites))
	copy(result, sprites)
	return result
}

// RemoveTextSprite は指定されたTextSpriteを削除する
func (tsm *TextSpriteManager) RemoveTextSprite(ts *TextSprite) {
	if ts == nil {
		return
	}

	tsm.mu.Lock()
	defer tsm.mu.Unlock()

	// スプライトを削除
	if ts.sprite != nil {
		tsm.spriteManager.DeleteSprite(ts.sprite.ID())
	}

	// リストから削除
	sprites := tsm.textSprites[ts.picID]
	for i, s := range sprites {
		if s == ts {
			tsm.textSprites[ts.picID] = append(sprites[:i], sprites[i+1:]...)
			break
		}
	}

	// リストが空になったら削除
	if len(tsm.textSprites[ts.picID]) == 0 {
		delete(tsm.textSprites, ts.picID)
	}
}

// RemoveTextSpritesByPicID はピクチャIDに関連するすべてのTextSpriteを削除する
func (tsm *TextSpriteManager) RemoveTextSpritesByPicID(picID int) {
	tsm.mu.Lock()
	defer tsm.mu.Unlock()

	sprites := tsm.textSprites[picID]
	for _, ts := range sprites {
		if ts.sprite != nil {
			tsm.spriteManager.DeleteSprite(ts.sprite.ID())
		}
	}
	delete(tsm.textSprites, picID)
}

// Clear はすべてのTextSpriteを削除する
func (tsm *TextSpriteManager) Clear() {
	tsm.mu.Lock()
	defer tsm.mu.Unlock()

	for picID, sprites := range tsm.textSprites {
		for _, ts := range sprites {
			if ts.sprite != nil {
				tsm.spriteManager.DeleteSprite(ts.sprite.ID())
			}
		}
		delete(tsm.textSprites, picID)
	}
}

// Count は登録されているTextSpriteの総数を返す
func (tsm *TextSpriteManager) Count() int {
	tsm.mu.RLock()
	defer tsm.mu.RUnlock()

	count := 0
	for _, sprites := range tsm.textSprites {
		count += len(sprites)
	}
	return count
}

// TextSprite methods

// GetSprite は基盤となるスプライトを返す
func (ts *TextSprite) GetSprite() *Sprite {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.sprite
}

// GetPicID は描画先ピクチャーIDを返す
func (ts *TextSprite) GetPicID() int {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.picID
}

// GetText はテキスト内容を返す
func (ts *TextSprite) GetText() string {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.text
}

// GetPosition は描画位置を返す
func (ts *TextSprite) GetPosition() (int, int) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.x, ts.y
}

// SetPosition は描画位置を更新する
func (ts *TextSprite) SetPosition(x, y int) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.x = x
	ts.y = y
	if ts.sprite != nil {
		ts.sprite.SetPosition(float64(x), float64(y))
	}
}

// SetVisible は可視性を更新する
func (ts *TextSprite) SetVisible(visible bool) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.sprite != nil {
		ts.sprite.SetVisible(visible)
	}
}

// SetParent は親スプライトを設定する
// ウインドウ内のテキスト描画で使用
func (ts *TextSprite) SetParent(parent *Sprite) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.sprite != nil && parent != nil {
		parent.AddChild(ts.sprite)
	}
}

// UpdateText はテキストを更新し、画像を再生成する
func (ts *TextSprite) UpdateText(text string) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.text == text {
		return
	}

	ts.text = text

	// 画像を再生成
	if ts.face != nil {
		opts := TextSpriteOptions{
			Text:      text,
			TextColor: ts.textColor,
			Face:      ts.face,
			BgColor:   ts.bgColor,
			X:         0,
			Y:         getFontHeight(ts.face),
		}

		img := CreateTextSpriteImage(opts)
		if img != nil && ts.sprite != nil {
			ts.sprite.SetImage(ebiten.NewImageFromImage(img))
		}
	}
}
