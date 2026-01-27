package graphics

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

// Layer は描画レイヤーの基本インターフェース
// 要件 1.1, 1.2, 1.3, 1.4: レイヤー構造の基本定義
type Layer interface {
	// GetID はレイヤーIDを返す
	GetID() int

	// GetBounds はレイヤーの境界ボックスを返す
	// 要件 4.3: 各レイヤーの境界ボックスを計算する
	GetBounds() image.Rectangle

	// GetZOrder はZ順序を返す
	// 要件 1.5, 1.6: Z順序による描画順序の管理
	GetZOrder() int

	// IsVisible は可視性を返す
	// 要件 3.3: 可視性の変更を追跡
	IsVisible() bool

	// IsOpaque はレイヤーが不透明かどうかを返す
	// 要件 7.2, 7.3: 上書きスキップの判定に使用
	// 不透明なレイヤーは下のレイヤーを完全に覆い隠す
	IsOpaque() bool

	// IsDirty はダーティフラグを返す
	// 要件 3.1, 3.2, 3.3: ダーティフラグによる最適化
	IsDirty() bool

	// SetDirty はダーティフラグを設定する
	// 要件 3.1, 3.2, 3.3: ダーティフラグの設定
	SetDirty(dirty bool)

	// GetImage はレイヤーの画像を返す（キャッシュがあればキャッシュを返す）
	// 要件 5.1, 5.2: レイヤーキャッシュの使用
	GetImage() *ebiten.Image

	// Invalidate はキャッシュを無効化する
	// 要件 5.3: キャッシュの無効化
	Invalidate()
}

// Z順序の定数
// 設計ドキュメントに基づくZ順序の割り当て
const (
	// ZOrderBackground は背景レイヤーのZ順序（常に最背面）
	ZOrderBackground = 0

	// ZOrderDrawing は描画レイヤー（MovePic）のZ順序
	ZOrderDrawing = 1

	// ZOrderCastBase はキャストレイヤーのZ順序の開始値
	ZOrderCastBase = 100

	// ZOrderCastMax はキャストレイヤーのZ順序の最大値
	ZOrderCastMax = 999

	// ZOrderTextBase はテキストレイヤーのZ順序の開始値
	ZOrderTextBase = 1000
)

// BaseLayer はレイヤーの基本実装を提供する構造体
// 各レイヤータイプはこの構造体を埋め込んで使用する
type BaseLayer struct {
	id      int
	bounds  image.Rectangle
	zOrder  int
	visible bool
	dirty   bool
	opaque  bool // 要件 7.2: 不透明度の追跡
}

// GetID はレイヤーIDを返す
func (l *BaseLayer) GetID() int {
	return l.id
}

// GetBounds はレイヤーの境界ボックスを返す
func (l *BaseLayer) GetBounds() image.Rectangle {
	return l.bounds
}

// GetZOrder はZ順序を返す
func (l *BaseLayer) GetZOrder() int {
	return l.zOrder
}

// IsVisible は可視性を返す
func (l *BaseLayer) IsVisible() bool {
	return l.visible
}

// IsDirty はダーティフラグを返す
func (l *BaseLayer) IsDirty() bool {
	return l.dirty
}

// SetDirty はダーティフラグを設定する
func (l *BaseLayer) SetDirty(dirty bool) {
	l.dirty = dirty
}

// SetVisible は可視性を設定し、ダーティフラグを設定する
// 要件 3.3: 可視性が変更されたときにダーティフラグを設定
func (l *BaseLayer) SetVisible(visible bool) {
	if l.visible != visible {
		l.visible = visible
		l.dirty = true
	}
}

// SetBounds は境界ボックスを設定し、ダーティフラグを設定する
// 要件 3.1: 位置が変更されたときにダーティフラグを設定
func (l *BaseLayer) SetBounds(bounds image.Rectangle) {
	if l.bounds != bounds {
		l.bounds = bounds
		l.dirty = true
	}
}

// SetZOrder はZ順序を設定する
func (l *BaseLayer) SetZOrder(zOrder int) {
	l.zOrder = zOrder
}

// SetID はレイヤーIDを設定する
func (l *BaseLayer) SetID(id int) {
	l.id = id
}

// IsOpaque はレイヤーが不透明かどうかを返す
// 要件 7.2, 7.3: 上書きスキップの判定に使用
func (l *BaseLayer) IsOpaque() bool {
	return l.opaque
}

// SetOpaque は不透明度を設定する
// 要件 7.2: 各レイヤーの不透明度を追跡する
func (l *BaseLayer) SetOpaque(opaque bool) {
	l.opaque = opaque
}
