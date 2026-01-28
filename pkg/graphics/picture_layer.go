package graphics

import (
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

// PictureLayer はMovePicで作成されるレイヤー
// 要件 2.1: Picture_Layerを定義する（MovePicで作成、焼き付け可能）
// 要件 2.5: Picture_Layerは焼き付け対象として機能する
type PictureLayer struct {
	BaseLayer

	// ウィンドウサイズの透明画像
	// 要件 3.5: Picture_Layerはウィンドウサイズの透明画像として初期化される
	image *ebiten.Image

	// 焼き付け可能フラグ
	// 要件 2.5: Picture_Layerは焼き付け対象として機能する
	bakeable bool
}

// NewPictureLayer は新しいPictureLayerを作成する
// ウィンドウサイズの透明画像として初期化される
// 要件 3.5: Picture_Layerはウィンドウサイズの透明画像として初期化される
// 要件 10.3: レイヤー作成に失敗したときにエラーをログに記録し、nilを返す
func NewPictureLayer(id, winWidth, winHeight int) *PictureLayer {
	// 要件 10.3: 無効なパラメータの場合はエラーをログに記録し、nilを返す
	// 要件 10.5: エラーメッセージに関数名と関連パラメータを含める
	if winWidth <= 0 || winHeight <= 0 {
		fmt.Printf("NewPictureLayer: invalid size, id=%d, width=%d, height=%d\n", id, winWidth, winHeight)
		return nil
	}

	bounds := image.Rect(0, 0, winWidth, winHeight)

	// ウィンドウサイズの透明画像を作成
	img := ebiten.NewImage(winWidth, winHeight)
	// ebiten.NewImageは自動的に透明（RGBA(0,0,0,0)）で初期化される

	return &PictureLayer{
		BaseLayer: BaseLayer{
			id:      id,
			bounds:  bounds,
			zOrder:  0, // Z順序は後で設定される
			visible: true,
			dirty:   true,
			opaque:  false, // 透明画像なので不透明ではない
		},
		image:    img,
		bakeable: true, // Picture_Layerは焼き付け可能
	}
}

// GetImage はレイヤーの画像を返す
// 要件 5.1, 5.2: レイヤーキャッシュの使用
func (l *PictureLayer) GetImage() *ebiten.Image {
	return l.image
}

// Invalidate はキャッシュを無効化する
// 要件 5.3: キャッシュの無効化
func (l *PictureLayer) Invalidate() {
	l.dirty = true
}

// IsBakeable は焼き付け可能かどうかを返す
// 要件 2.5: Picture_Layerは焼き付け対象として機能する
func (l *PictureLayer) IsBakeable() bool {
	return l.bakeable
}

// Bake は画像をこのレイヤーに焼き付ける
// 要件 3.2: 最上位レイヤーがPicture_Layerである場合、そのレイヤーに画像を焼き付ける
// 要件 3.6: 焼き付けが行われたとき、焼き付け先レイヤーをダーティとしてマークする
func (l *PictureLayer) Bake(src *ebiten.Image, destX, destY int) {
	if l.image == nil || src == nil {
		return
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(destX), float64(destY))
	l.image.DrawImage(src, op)

	// 焼き付け後、ダーティフラグを設定
	l.dirty = true
}

// BakeWithOptions は画像をこのレイヤーに焼き付ける（オプション付き）
// 透明色処理などの追加オプションをサポート
func (l *PictureLayer) BakeWithOptions(src *ebiten.Image, op *ebiten.DrawImageOptions) {
	if l.image == nil || src == nil {
		return
	}

	l.image.DrawImage(src, op)

	// 焼き付け後、ダーティフラグを設定
	l.dirty = true
}

// Clear はレイヤーの画像をクリアする（透明にする）
func (l *PictureLayer) Clear() {
	if l.image != nil {
		l.image.Clear()
		l.dirty = true
	}
}

// GetSize はレイヤーのサイズを返す
func (l *PictureLayer) GetSize() (width, height int) {
	if l.image == nil {
		return 0, 0
	}
	bounds := l.image.Bounds()
	return bounds.Dx(), bounds.Dy()
}

// GetLayerType はレイヤータイプを返す
// 要件 2.4: レイヤーが作成されたとき、レイヤータイプを識別可能にする
func (l *PictureLayer) GetLayerType() LayerType {
	return LayerTypePicture
}
