package graphics

import (
	"fmt"
	"image"
	"image/color"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

// リソース制限の定数
// 設計ドキュメントに基づく最大レイヤー数
const (
	// MaxCastLayers は最大キャストレイヤー数
	MaxCastLayers = 1024

	// MaxTextLayers は最大テキストレイヤー数
	MaxTextLayers = 256
)

// LayerManager はレイヤーを管理する
// 要件 1.1, 1.2, 1.3, 1.4: 各種レイヤーの管理
// スレッドセーフな実装（sync.RWMutex使用）
type LayerManager struct {
	// ウィンドウIDごとのレイヤーセット（新規）
	// 要件 1.1: レイヤーをWindowIDで管理する
	// 要件 1.5: WindowIDをキーとしてWindowLayerSetを検索する
	windowLayers map[int]*WindowLayerSet

	// ピクチャーIDごとのレイヤー（後方互換性のために残す）
	layers map[int]*PictureLayerSet

	// 次のレイヤーID
	nextLayerID int

	// ミューテックス
	mu sync.RWMutex
}

// PictureLayerSet はピクチャーに属するレイヤーのセット
// 要件 1.6: 背景 → 描画 → キャスト → テキストの順で合成
// 要件 10.1, 10.2: 操作順序に基づくZ順序管理
// スプライトシステム移行: BackgroundLayer, CastLayer, DrawingLayerは削除された
type PictureLayerSet struct {
	// ピクチャーID
	PicID int

	// テキストレイヤー
	// 要件 1.4: テキストレイヤー（Text_Layer）を管理する
	Texts []*TextLayerEntry

	// 描画エントリ（MovePicで作成）
	// 要件 10.3, 10.4: 各MovePic呼び出しで新しいDrawingEntryを作成
	DrawingEntries []*DrawingEntry

	// 合成バッファ
	// 要件 5.1: 各レイヤーの描画結果をキャッシュする
	CompositeBuffer *ebiten.Image

	// ダーティ領域
	// 要件 6.1: 変更があった領域（ダーティ領域）を追跡する
	DirtyRegion image.Rectangle

	// 全体がダーティかどうか
	// 要件 3.1, 3.2, 3.3: ダーティフラグによる最適化
	FullDirty bool

	// 次のZ順序カウンター（すべての操作で共有）
	// 要件 10.1, 10.2: 操作順序に基づくZ順序
	// 背景は常にZ=0なので、カウンターは1から開始
	nextZOrder int

	// 次のテキストZ順序オフセット（後方互換性のために残す）
	nextTextZOffset int
}

// NewLayerManager は新しいLayerManagerを作成する
func NewLayerManager() *LayerManager {
	return &LayerManager{
		windowLayers: make(map[int]*WindowLayerSet),
		layers:       make(map[int]*PictureLayerSet),
		nextLayerID:  1,
	}
}

// GetOrCreateWindowLayerSet は指定されたウィンドウIDのWindowLayerSetを取得または作成する
// 要件 1.2: ウィンドウが開かれたときにWindowLayerSetを作成する
// 要件 1.5: WindowIDをキーとしてWindowLayerSetを検索する
func (lm *LayerManager) GetOrCreateWindowLayerSet(winID, width, height int, bgColor color.Color) *WindowLayerSet {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	wls, exists := lm.windowLayers[winID]
	if !exists {
		wls = NewWindowLayerSet(winID, width, height, bgColor)
		lm.windowLayers[winID] = wls
	}

	return wls
}

// GetWindowLayerSet は指定されたウィンドウIDのWindowLayerSetを取得する
// 存在しない場合はnilを返す
// 要件 1.5: WindowIDをキーとしてWindowLayerSetを検索する
func (lm *LayerManager) GetWindowLayerSet(winID int) *WindowLayerSet {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	return lm.windowLayers[winID]
}

// DeleteWindowLayerSet は指定されたウィンドウIDのWindowLayerSetを削除する
// 要件 1.3: ウィンドウが閉じられたときにそのウィンドウに属するすべてのレイヤーを削除する
func (lm *LayerManager) DeleteWindowLayerSet(winID int) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	// WindowLayerSetが存在する場合、すべてのレイヤーをクリアしてから削除
	if wls, exists := lm.windowLayers[winID]; exists {
		wls.ClearLayers()
		delete(lm.windowLayers, winID)
	}
}

// GetOrCreatePictureLayerSet は指定されたピクチャーIDのPictureLayerSetを取得または作成する
// 要件 1.1, 1.2, 1.3, 1.4: 各種レイヤーの管理
func (lm *LayerManager) GetOrCreatePictureLayerSet(picID int) *PictureLayerSet {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	pls, exists := lm.layers[picID]
	if !exists {
		pls = NewPictureLayerSet(picID)
		lm.layers[picID] = pls
	}

	return pls
}

// GetPictureLayerSet は指定されたピクチャーIDのPictureLayerSetを取得する
// 存在しない場合はnilを返す
func (lm *LayerManager) GetPictureLayerSet(picID int) *PictureLayerSet {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	return lm.layers[picID]
}

// DeletePictureLayerSet は指定されたピクチャーIDのPictureLayerSetを削除する
// 要件 2.6: ウィンドウが閉じられたときにそのウィンドウに属するすべてのレイヤーを削除する
func (lm *LayerManager) DeletePictureLayerSet(picID int) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	delete(lm.layers, picID)
}

// GetNextLayerID は次のレイヤーIDを取得してインクリメントする
// 要件 1.5: レイヤーが追加されたときに自動的にZ順序を割り当てる
func (lm *LayerManager) GetNextLayerID() int {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	id := lm.nextLayerID
	lm.nextLayerID++
	return id
}

// GetAllPictureLayerSets はすべてのPictureLayerSetを取得する
func (lm *LayerManager) GetAllPictureLayerSets() map[int]*PictureLayerSet {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	// コピーを返す
	result := make(map[int]*PictureLayerSet, len(lm.layers))
	for k, v := range lm.layers {
		result[k] = v
	}
	return result
}

// GetPictureLayerSetCount はPictureLayerSetの数を返す
func (lm *LayerManager) GetPictureLayerSetCount() int {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	return len(lm.layers)
}

// Clear はすべてのPictureLayerSetとWindowLayerSetを削除する
func (lm *LayerManager) Clear() {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	lm.windowLayers = make(map[int]*WindowLayerSet)
	lm.layers = make(map[int]*PictureLayerSet)
	// nextLayerIDはリセットしない（一意性を保つため）
}

// NewPictureLayerSet は新しいPictureLayerSetを作成する
// スプライトシステム移行: BackgroundLayer, CastLayer, DrawingLayerは削除された
func NewPictureLayerSet(picID int) *PictureLayerSet {
	return &PictureLayerSet{
		PicID:           picID,
		Texts:           make([]*TextLayerEntry, 0),
		DrawingEntries:  make([]*DrawingEntry, 0),
		CompositeBuffer: nil,
		DirtyRegion:     image.Rectangle{},
		FullDirty:       true, // 初期状態はダーティ
		nextZOrder:      1,    // 背景はZ=0なので、1から開始
		nextTextZOffset: 0,
	}
}

// SetBackground は背景レイヤーを設定する（後方互換性のために残す、何もしない）
// Deprecated: スプライトシステム移行により不要になった
func (pls *PictureLayerSet) SetBackground(layer interface{}) {
	// スプライトシステム移行により、BackgroundLayerは不要になった
	pls.FullDirty = true
}

// SetDrawing は描画レイヤーを設定する（後方互換性のために残す、何もしない）
// Deprecated: スプライトシステム移行により不要になった
func (pls *PictureLayerSet) SetDrawing(layer interface{}) {
	// スプライトシステム移行により、DrawingLayerは不要になった
	pls.FullDirty = true
}

// AddCastLayer はキャストレイヤーを追加する（後方互換性のために残す、何もしない）
// Deprecated: スプライトシステム移行により不要になった
func (pls *PictureLayerSet) AddCastLayer(layer interface{}) {
	// スプライトシステム移行により、CastLayerは不要になった
	pls.nextZOrder++
	pls.FullDirty = true
}

// RemoveCastLayer はキャストレイヤーを削除する（後方互換性のために残す、常にfalseを返す）
// Deprecated: スプライトシステム移行により不要になった
func (pls *PictureLayerSet) RemoveCastLayer(castID int) bool {
	// スプライトシステム移行により、CastLayerは不要になった
	return false
}

// RemoveCastLayerByID はレイヤーIDでキャストレイヤーを削除する（後方互換性のために残す、常にfalseを返す）
// Deprecated: スプライトシステム移行により不要になった
func (pls *PictureLayerSet) RemoveCastLayerByID(layerID int) bool {
	// スプライトシステム移行により、CastLayerは不要になった
	return false
}

// GetCastLayer はキャストIDでキャストレイヤーを取得する（後方互換性のために残す、常にnilを返す）
// Deprecated: スプライトシステム移行により不要になった
func (pls *PictureLayerSet) GetCastLayer(castID int) interface{} {
	// スプライトシステム移行により、CastLayerは不要になった
	return nil
}

// GetCastLayerByID はレイヤーIDでキャストレイヤーを取得する（後方互換性のために残す、常にnilを返す）
// Deprecated: スプライトシステム移行により不要になった
func (pls *PictureLayerSet) GetCastLayerByID(layerID int) interface{} {
	// スプライトシステム移行により、CastLayerは不要になった
	return nil
}

// GetCastLayerCount はキャストレイヤーの数を返す（後方互換性のために残す、常に0を返す）
// Deprecated: スプライトシステム移行により不要になった
func (pls *PictureLayerSet) GetCastLayerCount() int {
	// スプライトシステム移行により、CastLayerは不要になった
	return 0
}

// AddTextLayer はテキストレイヤーを追加する
// 要件 1.4: テキストレイヤー（Text_Layer）を管理する
// 要件 1.5: レイヤーが追加されたときに自動的にZ順序を割り当てる
// 要件 10.1, 10.2: 操作順序に基づくZ順序
func (pls *PictureLayerSet) AddTextLayer(layer *TextLayerEntry) {
	// 操作順序に基づくZ順序を割り当て
	layer.SetZOrder(pls.nextZOrder)
	pls.nextZOrder++

	pls.Texts = append(pls.Texts, layer)
	pls.nextTextZOffset++
	pls.FullDirty = true
}

// AddDrawingEntry は描画エントリを追加する
// 要件 10.3, 10.4: MovePicが呼び出されたときにDrawingEntryを作成する
func (pls *PictureLayerSet) AddDrawingEntry(entry *DrawingEntry) {
	// 操作順序に基づくZ順序を割り当て
	entry.SetZOrder(pls.nextZOrder)
	pls.nextZOrder++

	pls.DrawingEntries = append(pls.DrawingEntries, entry)
	pls.FullDirty = true
}

// GetDrawingEntryCount は描画エントリの数を返す
func (pls *PictureLayerSet) GetDrawingEntryCount() int {
	return len(pls.DrawingEntries)
}

// ClearDrawingEntries はすべての描画エントリをクリアする
func (pls *PictureLayerSet) ClearDrawingEntries() {
	pls.DrawingEntries = make([]*DrawingEntry, 0)
	pls.FullDirty = true
}

// GetNextZOrder は次のZ順序を返す（インクリメントしない）
func (pls *PictureLayerSet) GetNextZOrder() int {
	return pls.nextZOrder
}

// RemoveTextLayer はテキストレイヤーを削除する
// 要件 10.2: 存在しないレイヤーIDが指定されたときにエラーをログに記録し、処理をスキップする
func (pls *PictureLayerSet) RemoveTextLayer(layerID int) bool {
	for i, text := range pls.Texts {
		if text.GetID() == layerID {
			// 削除前の位置をダーティ領域に追加
			pls.AddDirtyRegion(text.GetBounds())

			// スライスから削除
			pls.Texts = append(pls.Texts[:i], pls.Texts[i+1:]...)
			pls.FullDirty = true
			return true
		}
	}
	// 要件 10.2: 存在しないレイヤーIDが指定されたときにエラーをログに記録
	// 要件 10.5: エラーメッセージに関数名、ピクチャーID、レイヤーIDを含める
	fmt.Printf("RemoveTextLayer: layer not found, picID=%d, layerID=%d\n", pls.PicID, layerID)
	return false
}

// GetTextLayer はレイヤーIDでテキストレイヤーを取得する
// 要件 10.2: 存在しないレイヤーIDが指定されたときにエラーをログに記録し、処理をスキップする
func (pls *PictureLayerSet) GetTextLayer(layerID int) *TextLayerEntry {
	for _, text := range pls.Texts {
		if text.GetID() == layerID {
			return text
		}
	}
	// 要件 10.2: 存在しないレイヤーIDが指定されたときにエラーをログに記録
	// 要件 10.5: エラーメッセージに関数名、ピクチャーID、レイヤーIDを含める
	fmt.Printf("GetTextLayer: layer not found, picID=%d, layerID=%d\n", pls.PicID, layerID)
	return nil
}

// GetTextLayerCount はテキストレイヤーの数を返す
func (pls *PictureLayerSet) GetTextLayerCount() int {
	return len(pls.Texts)
}

// ClearTextLayers はすべてのテキストレイヤーをクリアする
func (pls *PictureLayerSet) ClearTextLayers() {
	pls.Texts = make([]*TextLayerEntry, 0)
	pls.nextTextZOffset = 0
	pls.FullDirty = true
}

// ClearCastLayers はすべてのキャストレイヤーをクリアする（後方互換性のために残す、何もしない）
// Deprecated: スプライトシステム移行により不要になった
func (pls *PictureLayerSet) ClearCastLayers() {
	// スプライトシステム移行により、CastLayerは不要になった
	pls.FullDirty = true
}

// GetNextCastZOffset は次のキャストZ順序オフセットを返す（後方互換性のために残す、常に0を返す）
// Deprecated: スプライトシステム移行により不要になった
func (pls *PictureLayerSet) GetNextCastZOffset() int {
	// スプライトシステム移行により、CastLayerは不要になった
	return 0
}

// GetNextTextZOffset は次のテキストZ順序オフセットを返す
// 要件 1.5: レイヤーが追加されたときに自動的にZ順序を割り当てる
func (pls *PictureLayerSet) GetNextTextZOffset() int {
	return pls.nextTextZOffset
}

// AddDirtyRegion はダーティ領域を追加する
// 要件 6.1: 変更があった領域（ダーティ領域）を追跡する
// 要件 6.3: 複数のダーティ領域があるときにそれらを統合して処理する
func (pls *PictureLayerSet) AddDirtyRegion(rect image.Rectangle) {
	if rect.Empty() {
		return
	}

	if pls.DirtyRegion.Empty() {
		pls.DirtyRegion = rect
	} else {
		// 要件 6.3: 複数のダーティ領域を統合
		pls.DirtyRegion = pls.DirtyRegion.Union(rect)
	}
}

// ClearDirtyRegion はダーティ領域をクリアする
// 要件 3.4: 合成処理が完了したときにすべてのDirty_Flagをクリアする
func (pls *PictureLayerSet) ClearDirtyRegion() {
	pls.DirtyRegion = image.Rectangle{}
	pls.FullDirty = false
}

// IsDirty はダーティかどうかを返す
// スプライトシステム移行: BackgroundLayer, CastLayer, DrawingLayerは削除された
func (pls *PictureLayerSet) IsDirty() bool {
	if pls.FullDirty {
		return true
	}
	if !pls.DirtyRegion.Empty() {
		return true
	}

	// 各レイヤーのダーティフラグをチェック
	for _, text := range pls.Texts {
		if text.IsDirty() {
			return true
		}
	}
	for _, entry := range pls.DrawingEntries {
		if entry.IsDirty() {
			return true
		}
	}

	return false
}

// MarkFullDirty は全体をダーティとしてマークする
func (pls *PictureLayerSet) MarkFullDirty() {
	pls.FullDirty = true
}

// SetCompositeBuffer は合成バッファを設定する
func (pls *PictureLayerSet) SetCompositeBuffer(buffer *ebiten.Image) {
	pls.CompositeBuffer = buffer
}

// GetCompositeBuffer は合成バッファを取得する
func (pls *PictureLayerSet) GetCompositeBuffer() *ebiten.Image {
	return pls.CompositeBuffer
}

// ClearAllDirtyFlags はすべてのレイヤーのダーティフラグをクリアする
// 要件 3.4: 合成処理が完了したときにすべてのDirty_Flagをクリアする
// スプライトシステム移行: BackgroundLayer, CastLayer, DrawingLayerは削除された
func (pls *PictureLayerSet) ClearAllDirtyFlags() {
	for _, text := range pls.Texts {
		text.SetDirty(false)
	}
	for _, entry := range pls.DrawingEntries {
		entry.SetDirty(false)
	}
	pls.ClearDirtyRegion()
}

// ShouldSkipLayer は上書きスキップ判定を行う
// 要件 7.1: 不透明なレイヤーが別のレイヤーを完全に覆っているときにそのレイヤーの描画をスキップする
// 要件 7.2: 部分的に覆われているレイヤーは描画する
// 要件 7.3: 透明なレイヤーは上書きスキップの対象としない
func (lm *LayerManager) ShouldSkipLayer(layer Layer, upperLayers []Layer) bool {
	if layer == nil {
		return true
	}

	// 非表示のレイヤーはスキップ
	if !layer.IsVisible() {
		return true
	}

	layerBounds := layer.GetBounds()
	if layerBounds.Empty() {
		return true
	}

	for _, upper := range upperLayers {
		if upper == nil {
			continue
		}

		// 要件 7.3: 透明なレイヤーは上書きスキップの対象としない
		if !upper.IsOpaque() {
			continue
		}

		// 上位レイヤーが非表示の場合はスキップ対象としない
		if !upper.IsVisible() {
			continue
		}

		upperBounds := upper.GetBounds()
		if upperBounds.Empty() {
			continue
		}

		// 要件 7.1: 不透明なレイヤーが別のレイヤーを完全に覆っているときにスキップ
		// 上位レイヤーの境界が下位レイヤーの境界を完全に含んでいるかチェック
		if containsRect(upperBounds, layerBounds) {
			return true
		}
	}

	// 要件 7.2: 部分的に覆われているレイヤーは描画する
	return false
}

// containsRect は rect1 が rect2 を完全に含んでいるかを判定する
func containsRect(rect1, rect2 image.Rectangle) bool {
	return rect1.Min.X <= rect2.Min.X &&
		rect1.Min.Y <= rect2.Min.Y &&
		rect1.Max.X >= rect2.Max.X &&
		rect1.Max.Y >= rect2.Max.Y
}

// GetUpperLayers は指定されたレイヤーより上位のレイヤーを取得する
// 合成処理で上書きスキップ判定に使用
// スプライトシステム移行: BackgroundLayer, CastLayer, DrawingLayerは削除された
func (pls *PictureLayerSet) GetUpperLayers(targetZOrder int) []Layer {
	var upperLayers []Layer

	// 描画エントリ
	for _, entry := range pls.DrawingEntries {
		if entry.GetZOrder() > targetZOrder {
			upperLayers = append(upperLayers, entry)
		}
	}

	// テキストレイヤー
	for _, text := range pls.Texts {
		if text.GetZOrder() > targetZOrder {
			upperLayers = append(upperLayers, text)
		}
	}

	return upperLayers
}

// GetAllLayersSorted はすべてのレイヤーをZ順序でソートして返す
// 要件 10.5: 合成時にすべてのレイヤーをZ順序でソートして描画
// スプライトシステム移行: BackgroundLayer, CastLayer, DrawingLayerは削除された
func (pls *PictureLayerSet) GetAllLayersSorted() []Layer {
	var layers []Layer

	// 描画エントリ
	for _, entry := range pls.DrawingEntries {
		layers = append(layers, entry)
	}

	// テキストレイヤー
	for _, text := range pls.Texts {
		layers = append(layers, text)
	}

	// Z順序でソート
	sortLayersByZOrder(layers)

	return layers
}

// sortLayersByZOrder はレイヤーをZ順序でソートする（挿入ソート）
func sortLayersByZOrder(layers []Layer) {
	for i := 1; i < len(layers); i++ {
		key := layers[i]
		j := i - 1
		for j >= 0 && layers[j].GetZOrder() > key.GetZOrder() {
			layers[j+1] = layers[j]
			j--
		}
		layers[j+1] = key
	}
}

// IsLayerVisible は可視領域クリッピング判定を行う
// 要件 4.1: レイヤーがウィンドウの可視領域外にあるときに描画をスキップする
// 要件 4.4: 可視領域との交差判定を行う
func IsLayerVisible(layer Layer, visibleRect image.Rectangle) bool {
	if layer == nil {
		return false
	}

	// 非表示のレイヤーは可視ではない
	if !layer.IsVisible() {
		return false
	}

	layerBounds := layer.GetBounds()
	if layerBounds.Empty() {
		return false
	}

	// 要件 4.4: 可視領域との交差判定
	// 交差領域が空でなければ可視
	return !layerBounds.Intersect(visibleRect).Empty()
}

// GetVisibleRegion はレイヤーの可視部分を返す
// 要件 4.2: レイヤーが部分的に可視領域内にあるときに可視部分のみを描画する
func GetVisibleRegion(layer Layer, visibleRect image.Rectangle) image.Rectangle {
	if layer == nil {
		return image.Rectangle{}
	}

	layerBounds := layer.GetBounds()
	return layerBounds.Intersect(visibleRect)
}

// Composite は可視領域内のレイヤーを合成して結果画像を返す
// 要件 1.6: 背景 → 描画 → キャスト → テキストの順で合成する
// 要件 6.2: ダーティ領域のみを再合成する
// 要件 3.4: 合成処理が完了したときにすべてのDirty_Flagをクリアする
func (pls *PictureLayerSet) Composite(visibleRect image.Rectangle) *ebiten.Image {
	// 可視領域が空の場合は何もしない
	if visibleRect.Empty() {
		return pls.CompositeBuffer
	}

	// ダーティでない場合はキャッシュを返す
	if !pls.IsDirty() && pls.CompositeBuffer != nil {
		return pls.CompositeBuffer
	}

	// 合成バッファの初期化または再利用
	bufferWidth := visibleRect.Dx()
	bufferHeight := visibleRect.Dy()

	if pls.CompositeBuffer == nil {
		pls.CompositeBuffer = ebiten.NewImage(bufferWidth, bufferHeight)
	} else {
		// バッファサイズが異なる場合は再作成
		currentBounds := pls.CompositeBuffer.Bounds()
		if currentBounds.Dx() != bufferWidth || currentBounds.Dy() != bufferHeight {
			pls.CompositeBuffer = ebiten.NewImage(bufferWidth, bufferHeight)
		}
	}

	// 要件 6.2: ダーティ領域のみを再合成
	// FullDirtyの場合は全体を再合成
	if pls.FullDirty {
		pls.CompositeBuffer.Clear()
		pls.compositeAllLayers(visibleRect)
	} else if !pls.DirtyRegion.Empty() {
		// ダーティ領域のみを再合成
		// 注: 現在の実装では簡略化のため、ダーティ領域がある場合も全体を再合成
		// 将来的にはダーティ領域のみを更新する最適化が可能
		pls.CompositeBuffer.Clear()
		pls.compositeAllLayers(visibleRect)
	} else {
		// 個別のレイヤーがダーティな場合
		pls.CompositeBuffer.Clear()
		pls.compositeAllLayers(visibleRect)
	}

	// 要件 3.4: 合成処理が完了したときにすべてのDirty_Flagをクリアする
	pls.ClearAllDirtyFlags()

	return pls.CompositeBuffer
}

// compositeAllLayers はすべてのレイヤーを合成する
// 要件 10.5: すべてのレイヤーをZ順序でソートして描画する
func (pls *PictureLayerSet) compositeAllLayers(visibleRect image.Rectangle) {
	lm := &LayerManager{} // 上書きスキップ判定用

	// すべてのレイヤーをZ順序でソートして取得
	layers := pls.GetAllLayersSorted()

	// Z順序順に描画
	for _, layer := range layers {
		upperLayers := pls.GetUpperLayers(layer.GetZOrder())
		pls.compositeLayer(layer, visibleRect, lm, upperLayers)
	}
}

// compositeLayer は単一のレイヤーを合成バッファに描画する
// 要件 4.1: レイヤーがウィンドウの可視領域外にあるときに描画をスキップする
// 要件 7.1: 不透明なレイヤーが別のレイヤーを完全に覆っているときにスキップする
func (pls *PictureLayerSet) compositeLayer(layer Layer, visibleRect image.Rectangle, lm *LayerManager, upperLayers []Layer) {
	if layer == nil {
		return
	}

	// 要件 4.1: 可視領域クリッピング
	if !IsLayerVisible(layer, visibleRect) {
		return
	}

	// 要件 7.1: 上書きスキップ判定
	if lm.ShouldSkipLayer(layer, upperLayers) {
		return
	}

	// レイヤーの画像を取得
	img := layer.GetImage()
	if img == nil {
		return
	}

	// レイヤーの位置を取得
	bounds := layer.GetBounds()

	// 可視領域の原点からの相対位置を計算
	destX := bounds.Min.X - visibleRect.Min.X
	destY := bounds.Min.Y - visibleRect.Min.Y

	// 描画オプションを設定
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(destX), float64(destY))

	// 合成バッファに描画
	pls.CompositeBuffer.DrawImage(img, op)
}

// CompositeWithLayerManager は指定されたLayerManagerを使用して合成する
// テスト用に公開されたメソッド
func (pls *PictureLayerSet) CompositeWithLayerManager(visibleRect image.Rectangle, lm *LayerManager) *ebiten.Image {
	// 可視領域が空の場合は何もしない
	if visibleRect.Empty() {
		return pls.CompositeBuffer
	}

	// ダーティでない場合はキャッシュを返す
	if !pls.IsDirty() && pls.CompositeBuffer != nil {
		return pls.CompositeBuffer
	}

	// 合成バッファの初期化または再利用
	bufferWidth := visibleRect.Dx()
	bufferHeight := visibleRect.Dy()

	if pls.CompositeBuffer == nil {
		pls.CompositeBuffer = ebiten.NewImage(bufferWidth, bufferHeight)
	} else {
		currentBounds := pls.CompositeBuffer.Bounds()
		if currentBounds.Dx() != bufferWidth || currentBounds.Dy() != bufferHeight {
			pls.CompositeBuffer = ebiten.NewImage(bufferWidth, bufferHeight)
		}
	}

	pls.CompositeBuffer.Clear()
	pls.compositeAllLayersWithLM(visibleRect, lm)

	// ダーティフラグをクリア
	pls.ClearAllDirtyFlags()

	return pls.CompositeBuffer
}

// compositeAllLayersWithLM は指定されたLayerManagerを使用してすべてのレイヤーを合成する
// 要件 10.5: すべてのレイヤーをZ順序でソートして描画する
func (pls *PictureLayerSet) compositeAllLayersWithLM(visibleRect image.Rectangle, lm *LayerManager) {
	// すべてのレイヤーをZ順序でソートして取得
	layers := pls.GetAllLayersSorted()

	// Z順序順に描画
	for _, layer := range layers {
		upperLayers := pls.GetUpperLayers(layer.GetZOrder())
		pls.compositeLayer(layer, visibleRect, lm, upperLayers)
	}
}
