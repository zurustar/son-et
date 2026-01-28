package graphics

import (
	"fmt"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// CastLayer はキャストを保持するレイヤー
// 要件 1.2: キャストレイヤー（Cast_Layer）をZ順序で管理する
// Z順序は100から開始（ZOrderCastBase = 100）
type CastLayer struct {
	BaseLayer
	castID        int
	picID         int
	srcPicID      int
	x, y          int
	srcX, srcY    int
	width, height int
	transColor    color.Color
	hasTransColor bool
	image         *ebiten.Image // キャッシュされた画像（透明色処理済み）
	sourceImage   *ebiten.Image // ソース画像への参照（キャッシュ生成用）
}

// NewCastLayer は新しいキャストレイヤーを作成する
// zOrderOffset はZOrderCastBaseからのオフセット値
// 要件 10.3: レイヤー作成に失敗したときにエラーをログに記録し、nilを返す
func NewCastLayer(id, castID, picID, srcPicID int, x, y, srcX, srcY, width, height int, zOrderOffset int) *CastLayer {
	// 要件 10.3: 無効なパラメータの場合はエラーをログに記録し、nilを返す
	// 要件 10.5: エラーメッセージに関数名と関連パラメータを含める
	if width <= 0 || height <= 0 {
		fmt.Printf("NewCastLayer: invalid size, id=%d, castID=%d, width=%d, height=%d\n", id, castID, width, height)
		return nil
	}

	bounds := image.Rect(x, y, x+width, y+height)

	layer := &CastLayer{
		BaseLayer: BaseLayer{
			id:      id,
			bounds:  bounds,
			zOrder:  ZOrderCastBase + zOrderOffset, // 100から開始
			visible: true,
			dirty:   true,
			opaque:  false, // キャストは透明色を持つ可能性があるため、デフォルトは透明
		},
		castID:        castID,
		picID:         picID,
		srcPicID:      srcPicID,
		x:             x,
		y:             y,
		srcX:          srcX,
		srcY:          srcY,
		width:         width,
		height:        height,
		transColor:    nil,
		hasTransColor: false,
		image:         nil,
		sourceImage:   nil,
	}

	return layer
}

// NewCastLayerWithTransColor は透明色付きで新しいキャストレイヤーを作成する
// 要件 10.3: レイヤー作成に失敗したときにエラーをログに記録し、nilを返す
func NewCastLayerWithTransColor(id, castID, picID, srcPicID int, x, y, srcX, srcY, width, height int, zOrderOffset int, transColor color.Color) *CastLayer {
	layer := NewCastLayer(id, castID, picID, srcPicID, x, y, srcX, srcY, width, height, zOrderOffset)
	if layer == nil {
		// NewCastLayerが既にエラーをログに記録しているので、ここでは追加のログは不要
		return nil
	}
	if transColor != nil {
		layer.transColor = transColor
		layer.hasTransColor = true
	}
	return layer
}

// GetImage はレイヤーの画像を返す（キャッシュがあればキャッシュを返す）
// 要件 5.1, 5.2: レイヤーキャッシュの使用
func (l *CastLayer) GetImage() *ebiten.Image {
	// キャッシュがあり、ダーティでなければキャッシュを返す
	if l.image != nil && !l.dirty {
		return l.image
	}

	// ソース画像がない場合はnilを返す
	if l.sourceImage == nil {
		return l.image
	}

	// キャッシュを再生成
	l.rebuildCache()
	return l.image
}

// Invalidate はキャッシュを無効化する
// 要件 5.3: キャッシュの無効化
func (l *CastLayer) Invalidate() {
	l.dirty = true
	l.image = nil
}

// SetSourceImage はソース画像を設定し、キャッシュを生成する
// 要件 3.2: 内容が変更されたときにダーティフラグを設定
func (l *CastLayer) SetSourceImage(src *ebiten.Image) {
	l.sourceImage = src
	l.dirty = true
	l.rebuildCache()
}

// rebuildCache はキャッシュを再構築する
// 透明色処理を含む
func (l *CastLayer) rebuildCache() {
	if l.sourceImage == nil || l.width <= 0 || l.height <= 0 {
		l.image = nil
		l.dirty = false
		return
	}

	// ソース画像から部分領域を取得
	srcBounds := l.sourceImage.Bounds()

	// ソース領域がソース画像の範囲内かチェック
	srcRect := image.Rect(l.srcX, l.srcY, l.srcX+l.width, l.srcY+l.height)
	srcRect = srcRect.Intersect(srcBounds)
	if srcRect.Empty() {
		l.image = nil
		l.dirty = false
		return
	}

	// 新しいキャッシュ画像を作成
	actualWidth := srcRect.Dx()
	actualHeight := srcRect.Dy()
	l.image = ebiten.NewImage(actualWidth, actualHeight)

	// ソース画像の部分領域を取得
	subImg := l.sourceImage.SubImage(srcRect).(*ebiten.Image)

	// 透明色処理なしでコピー（透明色処理は描画時に行う）
	op := &ebiten.DrawImageOptions{}
	l.image.DrawImage(subImg, op)

	l.dirty = false
}

// SetPosition は位置を設定し、ダーティフラグを設定する
// 要件 3.1: 位置が変更されたときにダーティフラグを設定
func (l *CastLayer) SetPosition(x, y int) {
	if l.x != x || l.y != y {
		l.x = x
		l.y = y
		l.bounds = image.Rect(x, y, x+l.width, y+l.height)
		l.dirty = true
	}
}

// SetTransColor は透明色を設定する
// 要件 3.2: 内容が変更されたときにダーティフラグを設定
func (l *CastLayer) SetTransColor(transColor color.Color) {
	l.transColor = transColor
	l.hasTransColor = transColor != nil
	l.dirty = true
	// キャッシュを無効化して再生成を促す
	l.image = nil
}

// GetCastID はキャストIDを返す
func (l *CastLayer) GetCastID() int {
	return l.castID
}

// GetPicID はピクチャーIDを返す
func (l *CastLayer) GetPicID() int {
	return l.picID
}

// SetPicID はピクチャーIDを設定する
func (l *CastLayer) SetPicID(picID int) {
	l.picID = picID
}

// GetSrcPicID はソースピクチャーIDを返す
func (l *CastLayer) GetSrcPicID() int {
	return l.srcPicID
}

// GetPosition は位置を返す
func (l *CastLayer) GetPosition() (int, int) {
	return l.x, l.y
}

// GetSourceRect はソース領域を返す
func (l *CastLayer) GetSourceRect() (srcX, srcY, width, height int) {
	return l.srcX, l.srcY, l.width, l.height
}

// GetSize はサイズを返す
func (l *CastLayer) GetSize() (width, height int) {
	return l.width, l.height
}

// HasTransColor は透明色が設定されているかを返す
func (l *CastLayer) HasTransColor() bool {
	return l.hasTransColor
}

// GetTransColor は透明色を返す
func (l *CastLayer) GetTransColor() color.Color {
	return l.transColor
}

// SetSourceRect はソース領域を設定する
// 要件 3.2: 内容が変更されたときにダーティフラグを設定
func (l *CastLayer) SetSourceRect(srcX, srcY, width, height int) {
	if l.srcX != srcX || l.srcY != srcY || l.width != width || l.height != height {
		l.srcX = srcX
		l.srcY = srcY
		l.width = width
		l.height = height
		l.bounds = image.Rect(l.x, l.y, l.x+width, l.y+height)
		l.dirty = true
		// キャッシュを無効化して再生成を促す
		l.image = nil
	}
}

// UpdateFromCast はCast構造体からレイヤーを更新する
func (l *CastLayer) UpdateFromCast(cast *Cast) {
	if cast == nil {
		return
	}

	posChanged := l.x != cast.X || l.y != cast.Y
	srcChanged := l.srcX != cast.SrcX || l.srcY != cast.SrcY ||
		l.width != cast.Width || l.height != cast.Height
	transChanged := l.hasTransColor != cast.HasTransColor ||
		(l.hasTransColor && !colorEqual(l.transColor, cast.TransColor))

	if posChanged {
		l.x = cast.X
		l.y = cast.Y
	}

	if srcChanged {
		l.srcX = cast.SrcX
		l.srcY = cast.SrcY
		l.width = cast.Width
		l.height = cast.Height
	}

	if transChanged {
		l.transColor = cast.TransColor
		l.hasTransColor = cast.HasTransColor
	}

	if posChanged || srcChanged {
		l.bounds = image.Rect(l.x, l.y, l.x+l.width, l.y+l.height)
	}

	if posChanged || srcChanged || transChanged {
		l.dirty = true
		if srcChanged || transChanged {
			l.image = nil // キャッシュを無効化
		}
	}

	l.visible = cast.Visible
}

// colorEqual は2つの色が等しいかを比較する
func colorEqual(c1, c2 color.Color) bool {
	if c1 == nil && c2 == nil {
		return true
	}
	if c1 == nil || c2 == nil {
		return false
	}
	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()
	return r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2
}

// SetCachedImage はキャッシュされた画像を直接設定する
// ベンチマークテスト用のメソッド
// 通常の使用ではSetSourceImageを使用すること
func (l *CastLayer) SetCachedImage(img *ebiten.Image) {
	l.image = img
	l.dirty = false
}

// GetLayerType はレイヤータイプを返す
// 要件 2.4: レイヤーが作成されたとき、レイヤータイプを識別可能にする
func (l *CastLayer) GetLayerType() LayerType {
	return LayerTypeCast
}
