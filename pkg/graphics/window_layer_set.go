package graphics

import (
	"fmt"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// WindowLayerSet はウィンドウに属するレイヤーの集合
// 要件 1.1: レイヤーをWindowIDで管理する
// 要件 1.4: 背景色とレイヤースタックを保持する
type WindowLayerSet struct {
	// ウィンドウID
	// 要件 1.1: レイヤーをWindowIDで管理する
	WinID int

	// 背景色
	// 要件 1.4: 背景色を保持する
	BgColor color.Color

	// ウィンドウサイズ
	Width, Height int

	// レイヤースタック（Z順序でソート）
	// 要件 1.4: レイヤースタックを保持する
	// 要件 6.2: レイヤースタックをZ順序（小さい順）で描画する
	Layers []Layer

	// 次のZ順序カウンター
	// 要件 6.3: 新しいレイヤーが作成されたときに現在のZ順序カウンターを割り当て、カウンターを増加させる
	// 要件 6.4: すべてのレイヤータイプで共通のカウンターを使用する
	nextZOrder int

	// 合成バッファ
	// 要件 9.2: 変更のないレイヤーのキャッシュを使用する
	CompositeBuffer *ebiten.Image

	// ダーティフラグ
	// 要件 9.1: ダーティフラグによる部分更新をサポートする
	FullDirty bool

	// ダーティ領域
	// 要件 9.5: レイヤーが変更されたときにダーティ領域のみを再合成する
	DirtyRegion image.Rectangle
}

// NewWindowLayerSet は新しいWindowLayerSetを作成する
// 要件 1.2: ウィンドウが開かれたときにWindowLayerSetを作成する
// 要件 10.3: レイヤー作成に失敗したときにエラーをログに記録し、nilを返す
func NewWindowLayerSet(winID int, width, height int, bgColor color.Color) *WindowLayerSet {
	// 要件 10.3: 無効なパラメータの場合はエラーをログに記録し、nilを返す
	// 要件 10.5: エラーメッセージに関数名と関連パラメータを含める
	if width <= 0 || height <= 0 {
		fmt.Printf("NewWindowLayerSet: invalid size, winID=%d, width=%d, height=%d\n", winID, width, height)
		return nil
	}

	return &WindowLayerSet{
		WinID:           winID,
		BgColor:         bgColor,
		Width:           width,
		Height:          height,
		Layers:          make([]Layer, 0),
		nextZOrder:      1, // 背景はZ=0なので、1から開始
		CompositeBuffer: nil,
		FullDirty:       true, // 初期状態はダーティ
		DirtyRegion:     image.Rectangle{},
	}
}

// GetWinID はウィンドウIDを返す
func (wls *WindowLayerSet) GetWinID() int {
	return wls.WinID
}

// GetBgColor は背景色を返す
func (wls *WindowLayerSet) GetBgColor() color.Color {
	return wls.BgColor
}

// SetBgColor は背景色を設定する
func (wls *WindowLayerSet) SetBgColor(bgColor color.Color) {
	wls.BgColor = bgColor
	wls.FullDirty = true
}

// GetSize はウィンドウサイズを返す
func (wls *WindowLayerSet) GetSize() (int, int) {
	return wls.Width, wls.Height
}

// SetSize はウィンドウサイズを設定する
func (wls *WindowLayerSet) SetSize(width, height int) {
	if wls.Width != width || wls.Height != height {
		wls.Width = width
		wls.Height = height
		wls.FullDirty = true
		// 合成バッファをリセット（サイズが変わったため）
		wls.CompositeBuffer = nil
	}
}

// AddLayer はレイヤーをスタックに追加する
// 要件 6.3: 新しいレイヤーが作成されたときに現在のZ順序カウンターを割り当て、カウンターを増加させる
func (wls *WindowLayerSet) AddLayer(layer Layer) {
	if layer == nil {
		return
	}

	// Z順序を割り当て
	layer.SetZOrder(wls.nextZOrder)
	wls.nextZOrder++

	wls.Layers = append(wls.Layers, layer)
	wls.FullDirty = true
}

// RemoveLayer はレイヤーをスタックから削除する
// 要件 10.2: 存在しないレイヤーIDが指定されたときにエラーをログに記録し、処理をスキップする
func (wls *WindowLayerSet) RemoveLayer(layerID int) bool {
	for i, layer := range wls.Layers {
		if layer.GetID() == layerID {
			// 削除前の位置をダーティ領域に追加
			wls.AddDirtyRegion(layer.GetBounds())

			// スライスから削除
			wls.Layers = append(wls.Layers[:i], wls.Layers[i+1:]...)
			wls.FullDirty = true
			return true
		}
	}
	// 要件 10.2: 存在しないレイヤーIDが指定されたときにエラーをログに記録
	// 要件 10.5: エラーメッセージに関数名、ウィンドウID、レイヤーIDを含める
	fmt.Printf("RemoveLayer: layer not found, windowID=%d, layerID=%d\n", wls.WinID, layerID)
	return false
}

// GetLayer はレイヤーIDでレイヤーを取得する
// 要件 10.2: 存在しないレイヤーIDが指定されたときにエラーをログに記録し、処理をスキップする
func (wls *WindowLayerSet) GetLayer(layerID int) Layer {
	for _, layer := range wls.Layers {
		if layer.GetID() == layerID {
			return layer
		}
	}
	// 要件 10.2: 存在しないレイヤーIDが指定されたときにエラーをログに記録
	// 要件 10.5: エラーメッセージに関数名、ウィンドウID、レイヤーIDを含める
	fmt.Printf("GetLayer: layer not found, windowID=%d, layerID=%d\n", wls.WinID, layerID)
	return nil
}

// GetLayerCount はレイヤーの数を返す
func (wls *WindowLayerSet) GetLayerCount() int {
	return len(wls.Layers)
}

// GetLayers はすべてのレイヤーを返す
func (wls *WindowLayerSet) GetLayers() []Layer {
	return wls.Layers
}

// GetLayersSorted はすべてのレイヤーをZ順序でソートして返す
// 要件 6.2: レイヤースタックをZ順序（小さい順）で描画する
func (wls *WindowLayerSet) GetLayersSorted() []Layer {
	// コピーを作成
	sorted := make([]Layer, len(wls.Layers))
	copy(sorted, wls.Layers)

	// Z順序でソート（挿入ソート）
	for i := 1; i < len(sorted); i++ {
		key := sorted[i]
		j := i - 1
		for j >= 0 && sorted[j].GetZOrder() > key.GetZOrder() {
			sorted[j+1] = sorted[j]
			j--
		}
		sorted[j+1] = key
	}

	return sorted
}

// GetNextZOrder は次のZ順序を返す（インクリメントしない）
func (wls *WindowLayerSet) GetNextZOrder() int {
	return wls.nextZOrder
}

// ClearLayers はすべてのレイヤーをクリアする
// 要件 1.3: ウィンドウが閉じられたときにすべてのレイヤーを削除する
func (wls *WindowLayerSet) ClearLayers() {
	wls.Layers = make([]Layer, 0)
	wls.nextZOrder = 1
	wls.FullDirty = true
}

// AddDirtyRegion はダーティ領域を追加する
// 要件 9.5: レイヤーが変更されたときにダーティ領域のみを再合成する
func (wls *WindowLayerSet) AddDirtyRegion(rect image.Rectangle) {
	if rect.Empty() {
		return
	}

	if wls.DirtyRegion.Empty() {
		wls.DirtyRegion = rect
	} else {
		// 複数のダーティ領域を統合
		wls.DirtyRegion = wls.DirtyRegion.Union(rect)
	}
}

// ClearDirtyRegion はダーティ領域をクリアする
func (wls *WindowLayerSet) ClearDirtyRegion() {
	wls.DirtyRegion = image.Rectangle{}
	wls.FullDirty = false
}

// IsDirty はダーティかどうかを返す
func (wls *WindowLayerSet) IsDirty() bool {
	if wls.FullDirty {
		return true
	}
	if !wls.DirtyRegion.Empty() {
		return true
	}

	// 各レイヤーのダーティフラグをチェック
	for _, layer := range wls.Layers {
		if layer.IsDirty() {
			return true
		}
	}

	return false
}

// MarkFullDirty は全体をダーティとしてマークする
func (wls *WindowLayerSet) MarkFullDirty() {
	wls.FullDirty = true
}

// SetCompositeBuffer は合成バッファを設定する
func (wls *WindowLayerSet) SetCompositeBuffer(buffer *ebiten.Image) {
	wls.CompositeBuffer = buffer
}

// GetCompositeBuffer は合成バッファを取得する
func (wls *WindowLayerSet) GetCompositeBuffer() *ebiten.Image {
	return wls.CompositeBuffer
}

// ClearAllDirtyFlags はすべてのレイヤーのダーティフラグをクリアする
func (wls *WindowLayerSet) ClearAllDirtyFlags() {
	for _, layer := range wls.Layers {
		layer.SetDirty(false)
	}
	wls.ClearDirtyRegion()
}

// GetTopmostLayer は最上位のレイヤーを返す
// 要件 3.1: MovePicが呼び出されたときに最上位レイヤーのタイプを確認する
func (wls *WindowLayerSet) GetTopmostLayer() Layer {
	if len(wls.Layers) == 0 {
		return nil
	}

	// Z順序が最大のレイヤーを探す
	var topmost Layer
	maxZOrder := -1

	for _, layer := range wls.Layers {
		if layer.GetZOrder() > maxZOrder {
			maxZOrder = layer.GetZOrder()
			topmost = layer
		}
	}

	return topmost
}

// GetDirtyRegion はダーティ領域を返す
func (wls *WindowLayerSet) GetDirtyRegion() image.Rectangle {
	return wls.DirtyRegion
}

// IsFullDirty は全体がダーティかどうかを返す
func (wls *WindowLayerSet) IsFullDirty() bool {
	return wls.FullDirty
}

// GetCastLayer はキャストIDでCastLayerを取得する
// 要件 4.2: MoveCastが呼び出されたときにCast_Layerの位置を更新する
func (wls *WindowLayerSet) GetCastLayer(castID int) *CastLayer {
	for _, layer := range wls.Layers {
		if castLayer, ok := layer.(*CastLayer); ok {
			if castLayer.GetCastID() == castID {
				return castLayer
			}
		}
	}
	return nil
}

// RemoveCastLayer はキャストIDでCastLayerを削除する
// 要件 4.3: DelCastが呼び出されたときにCast_Layerを削除する
func (wls *WindowLayerSet) RemoveCastLayer(castID int) bool {
	for i, layer := range wls.Layers {
		if castLayer, ok := layer.(*CastLayer); ok {
			if castLayer.GetCastID() == castID {
				// 削除前の位置をダーティ領域に追加
				wls.AddDirtyRegion(layer.GetBounds())

				// スライスから削除
				wls.Layers = append(wls.Layers[:i], wls.Layers[i+1:]...)
				wls.FullDirty = true
				return true
			}
		}
	}
	return false
}

// GetCastLayerCount はCastLayerの数を返す
func (wls *WindowLayerSet) GetCastLayerCount() int {
	count := 0
	for _, layer := range wls.Layers {
		if _, ok := layer.(*CastLayer); ok {
			count++
		}
	}
	return count
}

// GetAllCastLayers はすべてのCastLayerを返す
func (wls *WindowLayerSet) GetAllCastLayers() []*CastLayer {
	var casts []*CastLayer
	for _, layer := range wls.Layers {
		if castLayer, ok := layer.(*CastLayer); ok {
			casts = append(casts, castLayer)
		}
	}
	return casts
}

// ClearDirty はダーティフラグをクリアする
func (wls *WindowLayerSet) ClearDirty() {
	wls.FullDirty = false
	wls.DirtyRegion = image.Rectangle{}
}

// MarkDirty は特定の領域をダーティとしてマークする
// 要件 9.1: ダーティフラグによる部分更新をサポートする
// 要件 9.5: レイヤーが変更されたときにダーティ領域のみを再合成する
func (wls *WindowLayerSet) MarkDirty(rect image.Rectangle) {
	wls.AddDirtyRegion(rect)
}
