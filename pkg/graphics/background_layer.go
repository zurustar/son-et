package graphics

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

// BackgroundLayer は背景レイヤー
// 要件 1.1: 背景レイヤー（Background_Layer）を管理する
// Z順序は常に0（最背面）
type BackgroundLayer struct {
	BaseLayer
	picID int
	image *ebiten.Image
}

// NewBackgroundLayer は新しい背景レイヤーを作成する
// 要件 10.3: レイヤー作成に失敗したときにエラーをログに記録し、nilを返す
// 注意: imgがnilの場合は許容する（後で設定される場合がある）
func NewBackgroundLayer(id, picID int, img *ebiten.Image) *BackgroundLayer {
	var bounds image.Rectangle
	if img != nil {
		bounds = img.Bounds()
	}

	layer := &BackgroundLayer{
		BaseLayer: BaseLayer{
			id:      id,
			bounds:  bounds,
			zOrder:  ZOrderBackground, // 常に0（最背面）
			visible: true,
			dirty:   true,
			opaque:  true, // 背景レイヤーは通常不透明
		},
		picID: picID,
		image: img,
	}

	return layer
}

// GetImage はレイヤーの画像を返す
// 要件 5.1, 5.2: レイヤーキャッシュの使用
func (l *BackgroundLayer) GetImage() *ebiten.Image {
	return l.image
}

// Invalidate はキャッシュを無効化する
// 要件 5.3: キャッシュの無効化
func (l *BackgroundLayer) Invalidate() {
	l.dirty = true
}

// SetImage は背景画像を設定する
// 要件 3.2: 内容が変更されたときにダーティフラグを設定
func (l *BackgroundLayer) SetImage(img *ebiten.Image) {
	l.image = img
	if img != nil {
		l.bounds = img.Bounds()
	} else {
		l.bounds = image.Rectangle{}
	}
	l.dirty = true
}

// GetPicID はピクチャーIDを返す
func (l *BackgroundLayer) GetPicID() int {
	return l.picID
}

// SetPicID はピクチャーIDを設定する
func (l *BackgroundLayer) SetPicID(picID int) {
	l.picID = picID
}

// GetLayerType はレイヤータイプを返す
// 要件 2.4: レイヤーが作成されたとき、レイヤータイプを識別可能にする
// BackgroundLayerはPictureLayerの一種として扱う
func (l *BackgroundLayer) GetLayerType() LayerType {
	return LayerTypePicture
}
