package graphics

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

// DrawingEntry はMovePicで描画された内容を保持するエントリ
// 要件 10.3, 10.4: 各MovePic呼び出しで新しいDrawingEntryを作成
// Z順序は操作順序に基づいて動的に割り当てられる
type DrawingEntry struct {
	BaseLayer
	picID  int
	image  *ebiten.Image
	destX  int // 描画先X座標
	destY  int // 描画先Y座標
	width  int // 描画幅
	height int // 描画高さ
}

// NewDrawingEntry は新しいDrawingEntryを作成する
// 要件 10.3: MovePicが呼び出されたときにDrawingEntryを作成する
func NewDrawingEntry(id, picID int, img *ebiten.Image, destX, destY, width, height, zOrder int) *DrawingEntry {
	bounds := image.Rect(destX, destY, destX+width, destY+height)

	entry := &DrawingEntry{
		BaseLayer: BaseLayer{
			id:      id,
			bounds:  bounds,
			zOrder:  zOrder, // 操作順序に基づくZ順序
			visible: true,
			dirty:   true,
			opaque:  false, // 描画エントリは透明部分を含む可能性がある
		},
		picID:  picID,
		image:  img,
		destX:  destX,
		destY:  destY,
		width:  width,
		height: height,
	}

	return entry
}

// GetImage はレイヤーの画像を返す
func (e *DrawingEntry) GetImage() *ebiten.Image {
	return e.image
}

// Invalidate はキャッシュを無効化する
func (e *DrawingEntry) Invalidate() {
	e.dirty = true
}

// GetPicID はピクチャーIDを返す
func (e *DrawingEntry) GetPicID() int {
	return e.picID
}

// GetDestX は描画先X座標を返す
func (e *DrawingEntry) GetDestX() int {
	return e.destX
}

// GetDestY は描画先Y座標を返す
func (e *DrawingEntry) GetDestY() int {
	return e.destY
}

// GetWidth は描画幅を返す
func (e *DrawingEntry) GetWidth() int {
	return e.width
}

// GetHeight は描画高さを返す
func (e *DrawingEntry) GetHeight() int {
	return e.height
}

// DrawingLayer はMovePicで描画された内容を保持するレイヤー
// 要件 1.3: 描画レイヤー（Drawing_Layer）を管理する
// Z順序は常に1（背景の上、キャストの下）
// 注意: 新しい実装ではDrawingEntryを使用するが、後方互換性のために残す
type DrawingLayer struct {
	BaseLayer
	picID int
	image *ebiten.Image
}

// NewDrawingLayer は新しい描画レイヤーを作成する
func NewDrawingLayer(id, picID int, width, height int) *DrawingLayer {
	var img *ebiten.Image
	var bounds image.Rectangle

	if width > 0 && height > 0 {
		img = ebiten.NewImage(width, height)
		bounds = image.Rect(0, 0, width, height)
	}

	layer := &DrawingLayer{
		BaseLayer: BaseLayer{
			id:      id,
			bounds:  bounds,
			zOrder:  ZOrderDrawing, // 常に1（背景の上、キャストの下）
			visible: true,
			dirty:   true,
			opaque:  false, // 描画レイヤーは透明部分を含む可能性がある
		},
		picID: picID,
		image: img,
	}

	return layer
}

// NewDrawingLayerWithImage は既存の画像から新しい描画レイヤーを作成する
func NewDrawingLayerWithImage(id, picID int, img *ebiten.Image) *DrawingLayer {
	var bounds image.Rectangle
	if img != nil {
		bounds = img.Bounds()
	}

	layer := &DrawingLayer{
		BaseLayer: BaseLayer{
			id:      id,
			bounds:  bounds,
			zOrder:  ZOrderDrawing, // 常に1（背景の上、キャストの下）
			visible: true,
			dirty:   true,
			opaque:  false, // 描画レイヤーは透明部分を含む可能性がある
		},
		picID: picID,
		image: img,
	}

	return layer
}

// GetImage はレイヤーの画像を返す
// 要件 5.1, 5.2: レイヤーキャッシュの使用
func (l *DrawingLayer) GetImage() *ebiten.Image {
	return l.image
}

// Invalidate はキャッシュを無効化する
// 要件 5.3: キャッシュの無効化
func (l *DrawingLayer) Invalidate() {
	l.dirty = true
}

// SetImage は描画画像を設定する
// 要件 3.2: 内容が変更されたときにダーティフラグを設定
func (l *DrawingLayer) SetImage(img *ebiten.Image) {
	l.image = img
	if img != nil {
		l.bounds = img.Bounds()
	} else {
		l.bounds = image.Rectangle{}
	}
	l.dirty = true
}

// GetPicID はピクチャーIDを返す
func (l *DrawingLayer) GetPicID() int {
	return l.picID
}

// SetPicID はピクチャーIDを設定する
func (l *DrawingLayer) SetPicID(picID int) {
	l.picID = picID
}

// Clear は描画レイヤーの内容をクリアする
// 要件 3.2: 内容が変更されたときにダーティフラグを設定
func (l *DrawingLayer) Clear() {
	if l.image != nil {
		l.image.Clear()
		l.dirty = true
	}
}

// DrawImage は指定された位置に画像を描画する
// 要件 2.4: MovePicが呼び出されたときにDrawing_Layerに描画内容を追加する
func (l *DrawingLayer) DrawImage(src *ebiten.Image, x, y int) {
	if l.image == nil || src == nil {
		return
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	l.image.DrawImage(src, op)
	l.dirty = true
}

// DrawSubImage は指定された位置にソース画像の一部を描画する
// 要件 2.4: MovePicが呼び出されたときにDrawing_Layerに描画内容を追加する
func (l *DrawingLayer) DrawSubImage(src *ebiten.Image, destX, destY, srcX, srcY, width, height int) {
	if l.image == nil || src == nil {
		return
	}

	// ソース画像の部分領域を取得
	srcRect := image.Rect(srcX, srcY, srcX+width, srcY+height)
	subImg := src.SubImage(srcRect).(*ebiten.Image)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(destX), float64(destY))
	l.image.DrawImage(subImg, op)
	l.dirty = true
}

// Resize は描画レイヤーのサイズを変更する
// 既存の内容は保持されない
func (l *DrawingLayer) Resize(width, height int) {
	if width > 0 && height > 0 {
		l.image = ebiten.NewImage(width, height)
		l.bounds = image.Rect(0, 0, width, height)
	} else {
		l.image = nil
		l.bounds = image.Rectangle{}
	}
	l.dirty = true
}

// CopyFrom は別の画像から内容をコピーする
// 要件 3.2: 内容が変更されたときにダーティフラグを設定
func (l *DrawingLayer) CopyFrom(src *ebiten.Image) {
	if src == nil {
		return
	}

	srcBounds := src.Bounds()
	width := srcBounds.Dx()
	height := srcBounds.Dy()

	// 必要に応じてサイズを調整
	if l.image == nil || l.bounds.Dx() != width || l.bounds.Dy() != height {
		l.image = ebiten.NewImage(width, height)
		l.bounds = image.Rect(0, 0, width, height)
	}

	l.image.Clear()
	op := &ebiten.DrawImageOptions{}
	l.image.DrawImage(src, op)
	l.dirty = true
}
