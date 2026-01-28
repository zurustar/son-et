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

	// SetZOrder はZ順序を設定する
	// 要件 6.3: 新しいレイヤーが作成されたときに現在のZ順序カウンターを割り当てる
	SetZOrder(zOrder int)

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

	// GetLayerType はレイヤータイプを返す
	// 要件 2.4: レイヤーが作成されたとき、レイヤータイプを識別可能にする
	GetLayerType() LayerType
}

// LayerType はレイヤーの種類を表す
// 要件 2.4: レイヤーが作成されたとき、レイヤータイプを識別可能にする
type LayerType int

const (
	// LayerTypePicture はMovePicで作成されるレイヤー（焼き付け可能）
	// 要件 2.1: Picture_Layerを定義する
	LayerTypePicture LayerType = iota

	// LayerTypeText はTextWriteで作成されるレイヤー（常に新規作成）
	// 要件 2.2: Text_Layerを定義する
	LayerTypeText

	// LayerTypeCast はPutCastで作成されるスプライトレイヤー
	// 要件 2.3: Cast_Layerを定義する
	LayerTypeCast
)

// String はLayerTypeの文字列表現を返す
func (lt LayerType) String() string {
	switch lt {
	case LayerTypePicture:
		return "Picture"
	case LayerTypeText:
		return "Text"
	case LayerTypeCast:
		return "Cast"
	default:
		return "Unknown"
	}
}

// Z順序の定数
// 設計ドキュメントに基づくZ順序の割り当て
//
// ウインドウ内のZ順序（相対Z順序）:
//   - ZOrderBackground (0): 背景レイヤー
//   - ZOrderDrawing (1): 描画レイヤー（MovePic）
//   - ZOrderCastBase (100) - ZOrderCastMax (999): キャストレイヤー
//   - ZOrderTextBase (1000): テキストレイヤー
//
// グローバルZ順序（ウインドウ間の統一）:
//   - 各ウインドウにZOrderWindowRangeの範囲を割り当て
//   - ウインドウ内のスプライトはその範囲内でZ順序を持つ
//   - 例: ウインドウ0は0-9999、ウインドウ1は10000-19999
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

	// ZOrderWindowRange はウインドウごとのZ順序の範囲
	// 各ウインドウはこの範囲内でスプライトのZ順序を管理する
	// ウインドウ0: 0 - 9999
	// ウインドウ1: 10000 - 19999
	// ...
	ZOrderWindowRange = 10000

	// ZOrderWindowBase はウインドウスプライト自体のZ順序オフセット
	// ウインドウスプライトはウインドウ範囲の先頭に配置される
	ZOrderWindowBase = 0
)

// CalculateGlobalZOrder はウインドウのZ順序とウインドウ内の相対Z順序からグローバルZ順序を計算する
// windowZOrder: ウインドウのZ順序（0, 1, 2, ...）
// localZOrder: ウインドウ内の相対Z順序（ZOrderCastBase + offset など）
// 戻り値: グローバルZ順序
//
// 例:
//   - ウインドウ0のキャスト（localZOrder=100）: 0 * 10000 + 100 = 100
//   - ウインドウ1のキャスト（localZOrder=100）: 1 * 10000 + 100 = 10100
//   - ウインドウ0のテキスト（localZOrder=1000）: 0 * 10000 + 1000 = 1000
//   - ウインドウ1のテキスト（localZOrder=1000）: 1 * 10000 + 1000 = 11000
func CalculateGlobalZOrder(windowZOrder, localZOrder int) int {
	return windowZOrder*ZOrderWindowRange + localZOrder
}

// CalculateWindowZOrderFromGlobal はグローバルZ順序からウインドウのZ順序を計算する
// globalZOrder: グローバルZ順序
// 戻り値: ウインドウのZ順序
func CalculateWindowZOrderFromGlobal(globalZOrder int) int {
	return globalZOrder / ZOrderWindowRange
}

// CalculateLocalZOrderFromGlobal はグローバルZ順序からウインドウ内の相対Z順序を計算する
// globalZOrder: グローバルZ順序
// 戻り値: ウインドウ内の相対Z順序
func CalculateLocalZOrderFromGlobal(globalZOrder int) int {
	return globalZOrder % ZOrderWindowRange
}

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
