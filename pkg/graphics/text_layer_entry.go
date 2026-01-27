package graphics

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

// TextLayerEntry はテキストレイヤーのエントリ
// 要件 1.4: テキストレイヤー（Text_Layer）を管理する
// Z順序は1000から開始（ZOrderTextBase = 1000）
// 要件 5.4: テキストレイヤーのキャッシュを特に重視する（作成コストが高いため）
type TextLayerEntry struct {
	BaseLayer
	picID int
	x, y  int
	text  string
	image *ebiten.Image // キャッシュされた画像
}

// NewTextLayerEntry は新しいテキストレイヤーエントリを作成する
// zOrderOffset はZOrderTextBaseからのオフセット値
func NewTextLayerEntry(id, picID int, x, y int, text string, zOrderOffset int) *TextLayerEntry {
	entry := &TextLayerEntry{
		BaseLayer: BaseLayer{
			id:      id,
			bounds:  image.Rectangle{},             // 画像が設定されるまで空
			zOrder:  ZOrderTextBase + zOrderOffset, // 1000から開始
			visible: true,
			dirty:   true,
			opaque:  false, // テキストレイヤーは透明部分を含む
		},
		picID: picID,
		x:     x,
		y:     y,
		text:  text,
		image: nil,
	}

	return entry
}

// NewTextLayerEntryWithImage は既存の画像から新しいテキストレイヤーエントリを作成する
// 既存のTextLayerから変換する際に使用
func NewTextLayerEntryWithImage(id, picID int, x, y int, text string, img *ebiten.Image, zOrderOffset int) *TextLayerEntry {
	var bounds image.Rectangle
	if img != nil {
		// 画像の境界を位置に合わせて設定
		imgBounds := img.Bounds()
		bounds = image.Rect(x, y, x+imgBounds.Dx(), y+imgBounds.Dy())
	}

	entry := &TextLayerEntry{
		BaseLayer: BaseLayer{
			id:      id,
			bounds:  bounds,
			zOrder:  ZOrderTextBase + zOrderOffset, // 1000から開始
			visible: true,
			dirty:   false, // 画像が既に設定されているのでダーティではない
			opaque:  false, // テキストレイヤーは透明部分を含む
		},
		picID: picID,
		x:     x,
		y:     y,
		text:  text,
		image: img,
	}

	return entry
}

// NewTextLayerEntryFromTextLayer は既存のTextLayerからTextLayerEntryを作成する
// 既存のTextLayerManagerとの互換性のため
func NewTextLayerEntryFromTextLayer(id int, textLayer *TextLayer, zOrderOffset int) *TextLayerEntry {
	if textLayer == nil {
		return nil
	}

	var img *ebiten.Image
	if textLayer.Image != nil {
		img = ebiten.NewImageFromImage(textLayer.Image)
	}

	return NewTextLayerEntryWithImage(
		id,
		textLayer.PicID,
		textLayer.X,
		textLayer.Y,
		"", // TextLayerはテキスト文字列を保持していない
		img,
		zOrderOffset,
	)
}

// GetImage はレイヤーの画像を返す（キャッシュがあればキャッシュを返す）
// 要件 5.1, 5.2: レイヤーキャッシュの使用
// 要件 5.4: テキストレイヤーのキャッシュを特に重視する（作成コストが高いため）
func (l *TextLayerEntry) GetImage() *ebiten.Image {
	return l.image
}

// Invalidate はキャッシュを無効化する
// 要件 5.3: キャッシュの無効化
func (l *TextLayerEntry) Invalidate() {
	l.dirty = true
	l.image = nil
}

// SetImage はキャッシュされた画像を設定する
// 要件 3.2: 内容が変更されたときにダーティフラグを設定
func (l *TextLayerEntry) SetImage(img *ebiten.Image) {
	l.image = img
	if img != nil {
		imgBounds := img.Bounds()
		l.bounds = image.Rect(l.x, l.y, l.x+imgBounds.Dx(), l.y+imgBounds.Dy())
	} else {
		l.bounds = image.Rectangle{}
	}
	l.dirty = false // 画像が設定されたのでダーティではない
}

// SetText はテキストを設定し、ダーティフラグを設定する
// 要件 3.2: 内容が変更されたときにダーティフラグを設定
func (l *TextLayerEntry) SetText(text string) {
	if l.text != text {
		l.text = text
		l.dirty = true
		l.image = nil // キャッシュを無効化
	}
}

// GetText はテキストを返す
func (l *TextLayerEntry) GetText() string {
	return l.text
}

// SetPosition は位置を設定し、ダーティフラグを設定する
// 要件 3.1: 位置が変更されたときにダーティフラグを設定
func (l *TextLayerEntry) SetPosition(x, y int) {
	if l.x != x || l.y != y {
		l.x = x
		l.y = y
		// 境界ボックスを更新
		if l.image != nil {
			imgBounds := l.image.Bounds()
			l.bounds = image.Rect(x, y, x+imgBounds.Dx(), y+imgBounds.Dy())
		}
		l.dirty = true
	}
}

// GetPosition は位置を返す
func (l *TextLayerEntry) GetPosition() (int, int) {
	return l.x, l.y
}

// GetPicID はピクチャーIDを返す
func (l *TextLayerEntry) GetPicID() int {
	return l.picID
}

// SetPicID はピクチャーIDを設定する
func (l *TextLayerEntry) SetPicID(picID int) {
	l.picID = picID
}

// GetSize はサイズを返す
func (l *TextLayerEntry) GetSize() (width, height int) {
	if l.image != nil {
		bounds := l.image.Bounds()
		return bounds.Dx(), bounds.Dy()
	}
	return 0, 0
}

// HasImage はキャッシュされた画像があるかを返す
func (l *TextLayerEntry) HasImage() bool {
	return l.image != nil
}
