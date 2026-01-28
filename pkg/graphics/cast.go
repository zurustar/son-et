package graphics

import (
	"fmt"
	"image/color"
	"sort"
	"sync"
)

// Cast はスプライトを表す
// 要件 4.1, 4.2
type Cast struct {
	ID            int // キャストID（0から始まる連番）
	WinID         int // 所属するウィンドウ
	PicID         int // ソースピクチャー
	X, Y          int // ウィンドウ内の位置
	SrcX          int // ピクチャー内のソースX
	SrcY          int // ピクチャー内のソースY
	Width         int // 幅
	Height        int // 高さ
	Visible       bool
	ZOrder        int         // Z順序（大きいほど前面）
	TransColor    color.Color // 透明色（nilの場合は透明色なし）
	HasTransColor bool        // 透明色が設定されているか
}

// CastManager はキャストを管理する
// 要件 9.7: 最大1024キャスト
// 要件 8.2: LayerManagerとの統合
type CastManager struct {
	casts        map[int]*Cast
	nextID       int
	maxID        int // 最大1024
	nextZOrder   int
	mu           sync.RWMutex
	layerManager *LayerManager // 要件 8.2: LayerManagerとの統合（オプション）
}

// CastOption はキャストのオプションを設定する関数型
type CastOption func(*Cast)

// WithCastPosition はキャストの位置を設定する
func WithCastPosition(x, y int) CastOption {
	return func(c *Cast) {
		c.X = x
		c.Y = y
	}
}

// WithCastSource はキャストのソース領域を設定する
func WithCastSource(srcX, srcY, width, height int) CastOption {
	return func(c *Cast) {
		c.SrcX = srcX
		c.SrcY = srcY
		c.Width = width
		c.Height = height
	}
}

// WithCastPicID はキャストのピクチャーIDを設定する
func WithCastPicID(picID int) CastOption {
	return func(c *Cast) {
		c.PicID = picID
	}
}

// WithCastTransColor はキャストの透明色を設定する
func WithCastTransColor(transColor color.Color) CastOption {
	return func(c *Cast) {
		c.TransColor = transColor
		c.HasTransColor = true
	}
}

// NewCastManager は新しい CastManager を作成する
func NewCastManager() *CastManager {
	return &CastManager{
		casts:        make(map[int]*Cast),
		nextID:       0,
		maxID:        1024, // 要件 9.7
		nextZOrder:   0,
		layerManager: nil, // デフォルトはLayerManager統合なし
	}
}

// SetLayerManager はLayerManagerを設定する
// 要件 8.2: CastManagerとLayerManagerを統合する
func (cm *CastManager) SetLayerManager(lm *LayerManager) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.layerManager = lm
}

// GetLayerManager はLayerManagerを取得する
func (cm *CastManager) GetLayerManager() *LayerManager {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.layerManager
}

// PutCast はキャストを配置する
// 要件 4.1, 4.2
func (cm *CastManager) PutCast(winID, picID, x, y, srcX, srcY, width, height int) (int, error) {
	return cm.PutCastWithTransColor(winID, picID, x, y, srcX, srcY, width, height, nil)
}

// PutCastWithTransColor は透明色付きでキャストを配置する
// 要件 2.1: PutCastが呼び出されたときに対応するCast_Layerを作成する
func (cm *CastManager) PutCastWithTransColor(winID, picID, x, y, srcX, srcY, width, height int, transColor color.Color) (int, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// リソース制限チェック（要件 9.7, 9.8）
	if len(cm.casts) >= cm.maxID {
		return -1, fmt.Errorf("cast limit reached (max: %d)", cm.maxID)
	}

	// 新しいキャストを作成
	cast := &Cast{
		ID:            cm.nextID,
		WinID:         winID,
		PicID:         picID,
		X:             x,
		Y:             y,
		SrcX:          srcX,
		SrcY:          srcY,
		Width:         width,
		Height:        height,
		Visible:       true,
		ZOrder:        cm.nextZOrder,
		TransColor:    transColor,
		HasTransColor: transColor != nil,
	}

	// キャストを登録
	cm.casts[cast.ID] = cast
	cm.nextID++
	cm.nextZOrder++

	// 要件 2.1: LayerManagerが設定されている場合、CastLayerを作成
	if cm.layerManager != nil {
		cm.createCastLayer(cast)
	}

	return cast.ID, nil
}

// createCastLayer はキャストに対応するCastLayerを作成する
// 要件 2.1: PutCastが呼び出されたときに対応するCast_Layerを作成する
// 要件 7.1: Cast_LayerをウィンドウIDで登録する
// 要件 10.1: 存在しないウィンドウIDが指定されたときにエラーをログに記録し、処理をスキップする
func (cm *CastManager) createCastLayer(cast *Cast) {
	if cm.layerManager == nil || cast == nil {
		return
	}

	// レイヤーIDを取得
	layerID := cm.layerManager.GetNextLayerID()

	// WindowLayerSetを取得（要件 7.1: ウィンドウIDで登録）
	wls := cm.layerManager.GetWindowLayerSet(cast.WinID)
	if wls != nil {
		// WindowLayerSetが存在する場合、そこにCastLayerを追加
		// Z順序はWindowLayerSetのAddLayerで自動的に割り当てられる
		var castLayer *CastLayer
		if cast.HasTransColor {
			castLayer = NewCastLayerWithTransColor(
				layerID,
				cast.ID,
				cast.WinID, // destPicID（ウィンドウID）
				cast.PicID, // srcPicID
				cast.X,
				cast.Y,
				cast.SrcX,
				cast.SrcY,
				cast.Width,
				cast.Height,
				0, // zOrderOffsetはWindowLayerSet.AddLayerで設定されるため0
				cast.TransColor,
			)
		} else {
			castLayer = NewCastLayer(
				layerID,
				cast.ID,
				cast.WinID, // destPicID（ウィンドウID）
				cast.PicID, // srcPicID
				cast.X,
				cast.Y,
				cast.SrcX,
				cast.SrcY,
				cast.Width,
				cast.Height,
				0, // zOrderOffsetはWindowLayerSet.AddLayerで設定されるため0
			)
		}

		// WindowLayerSetにCastLayerを追加
		wls.AddLayer(castLayer)
		return
	}

	// 要件 10.1: 存在しないウィンドウIDが指定されたときにエラーをログに記録
	// 後方互換性のためPictureLayerSetにフォールバックするが、警告をログに記録
	// Note: slogを使用するためにはCastManagerにloggerを追加する必要があるが、
	// 現在の設計ではCastManagerはloggerを持っていないため、fmtでログを出力
	// 将来的にはCastManagerにloggerを追加することを検討
	fmt.Printf("createCastLayer: WindowLayerSet not found, windowID=%d, castID=%d (using PictureLayerSet fallback)\n", cast.WinID, cast.ID)

	// 後方互換性: WindowLayerSetが存在しない場合はPictureLayerSetを使用
	pls := cm.layerManager.GetOrCreatePictureLayerSet(cast.WinID)

	// Z順序オフセットを取得
	zOrderOffset := pls.GetNextCastZOffset()

	// CastLayerを作成
	var castLayer *CastLayer
	if cast.HasTransColor {
		castLayer = NewCastLayerWithTransColor(
			layerID,
			cast.ID,
			cast.WinID, // destPicID（ウィンドウID）
			cast.PicID, // srcPicID
			cast.X,
			cast.Y,
			cast.SrcX,
			cast.SrcY,
			cast.Width,
			cast.Height,
			zOrderOffset,
			cast.TransColor,
		)
	} else {
		castLayer = NewCastLayer(
			layerID,
			cast.ID,
			cast.WinID, // destPicID（ウィンドウID）
			cast.PicID, // srcPicID
			cast.X,
			cast.Y,
			cast.SrcX,
			cast.SrcY,
			cast.Width,
			cast.Height,
			zOrderOffset,
		)
	}

	// PictureLayerSetにCastLayerを追加
	pls.AddCastLayer(castLayer)
}

// MoveCast はキャストの位置やソース領域を変更する
// 要件 4.3, 4.4, 4.5
// 要件 2.2: MoveCastが呼び出されたときに対応するCast_Layerの位置を更新する
func (cm *CastManager) MoveCast(id int, opts ...CastOption) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// キャストを取得
	cast, exists := cm.casts[id]
	if !exists {
		return fmt.Errorf("cast not found: %d", id)
	}

	// 変更前の位置を記録（ダーティ領域追跡用）
	oldX, oldY := cast.X, cast.Y

	// オプションを適用
	for _, opt := range opts {
		opt(cast)
	}

	// 要件 2.2: LayerManagerが設定されている場合、CastLayerを更新
	if cm.layerManager != nil {
		cm.updateCastLayer(cast, oldX, oldY)
	}

	return nil
}

// updateCastLayer はキャストに対応するCastLayerを更新する
// 要件 2.2: MoveCastが呼び出されたときに対応するCast_Layerの位置を更新する
// 要件 4.2: Cast_Layerの位置を更新する（残像なし）
// 要件 10.1: 存在しないウィンドウIDが指定されたときにエラーをログに記録し、処理をスキップする
func (cm *CastManager) updateCastLayer(cast *Cast, oldX, oldY int) {
	if cm.layerManager == nil || cast == nil {
		return
	}

	// まずWindowLayerSetを試す（要件 7.1: ウィンドウIDで登録）
	wls := cm.layerManager.GetWindowLayerSet(cast.WinID)
	if wls != nil {
		// WindowLayerSetからCastLayerを取得
		castLayer := wls.GetCastLayer(cast.ID)
		if castLayer != nil {
			// 古い位置をダーティ領域に追加（要件 4.6: 古い位置をダーティ領域としてマーク）
			if oldX != cast.X || oldY != cast.Y {
				wls.AddDirtyRegion(castLayer.GetBounds())
			}

			// CastLayerを更新
			castLayer.UpdateFromCast(cast)

			// 新しい位置をダーティ領域に追加（要件 4.6: 新しい位置をダーティ領域としてマーク）
			wls.AddDirtyRegion(castLayer.GetBounds())
			return
		}
		// 要件 10.1: CastLayerが見つからない場合はエラーをログに記録
		fmt.Printf("updateCastLayer: CastLayer not found in WindowLayerSet, windowID=%d, castID=%d\n", cast.WinID, cast.ID)
		return
	}

	// 要件 10.1: WindowLayerSetが見つからない場合はエラーをログに記録
	// 後方互換性のためPictureLayerSetにフォールバック
	fmt.Printf("updateCastLayer: WindowLayerSet not found, windowID=%d, castID=%d (using PictureLayerSet fallback)\n", cast.WinID, cast.ID)

	// 後方互換性: WindowLayerSetが存在しない場合はPictureLayerSetを使用
	pls := cm.layerManager.GetPictureLayerSet(cast.WinID)
	if pls == nil {
		// 要件 10.1: PictureLayerSetも見つからない場合はエラーをログに記録し、処理をスキップ
		fmt.Printf("updateCastLayer: PictureLayerSet not found, windowID=%d, castID=%d (skipping)\n", cast.WinID, cast.ID)
		return
	}

	// CastLayerを取得
	castLayer := pls.GetCastLayer(cast.ID)
	if castLayer == nil {
		// 要件 10.1: CastLayerが見つからない場合はエラーをログに記録し、処理をスキップ
		fmt.Printf("updateCastLayer: CastLayer not found in PictureLayerSet, windowID=%d, castID=%d (skipping)\n", cast.WinID, cast.ID)
		return
	}

	// 古い位置をダーティ領域に追加
	if oldX != cast.X || oldY != cast.Y {
		pls.AddDirtyRegion(castLayer.GetBounds())
	}

	// CastLayerを更新
	castLayer.UpdateFromCast(cast)

	// 新しい位置をダーティ領域に追加
	pls.AddDirtyRegion(castLayer.GetBounds())
}

// DelCast は指定されたキャストを削除する
// 要件 4.6
// 要件 2.3: DelCastが呼び出されたときに対応するCast_Layerを削除する
func (cm *CastManager) DelCast(id int) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// キャストが存在するか確認
	cast, exists := cm.casts[id]
	if !exists {
		return fmt.Errorf("cast not found: %d", id)
	}

	// 要件 2.3: LayerManagerが設定されている場合、CastLayerを削除
	if cm.layerManager != nil {
		cm.deleteCastLayer(cast)
	}

	// キャストを削除
	delete(cm.casts, id)

	return nil
}

// deleteCastLayer はキャストに対応するCastLayerを削除する
// 要件 2.3: DelCastが呼び出されたときに対応するCast_Layerを削除する
// 要件 4.3: DelCastが呼び出されたときにCast_Layerを削除する
// 要件 10.1: 存在しないウィンドウIDが指定されたときにエラーをログに記録し、処理をスキップする
func (cm *CastManager) deleteCastLayer(cast *Cast) {
	if cm.layerManager == nil || cast == nil {
		return
	}

	// まずWindowLayerSetを試す（要件 7.1: ウィンドウIDで登録）
	wls := cm.layerManager.GetWindowLayerSet(cast.WinID)
	if wls != nil {
		// WindowLayerSetからCastLayerを削除
		// RemoveCastLayerは内部でダーティ領域を追加する
		if wls.RemoveCastLayer(cast.ID) {
			return
		}
		// 要件 10.1: CastLayerが見つからない場合はエラーをログに記録
		fmt.Printf("deleteCastLayer: CastLayer not found in WindowLayerSet, windowID=%d, castID=%d\n", cast.WinID, cast.ID)
		return
	}

	// 要件 10.1: WindowLayerSetが見つからない場合はエラーをログに記録
	// 後方互換性のためPictureLayerSetにフォールバック
	fmt.Printf("deleteCastLayer: WindowLayerSet not found, windowID=%d, castID=%d (using PictureLayerSet fallback)\n", cast.WinID, cast.ID)

	// 後方互換性: WindowLayerSetが存在しない場合はPictureLayerSetを使用
	pls := cm.layerManager.GetPictureLayerSet(cast.WinID)
	if pls == nil {
		// 要件 10.1: PictureLayerSetも見つからない場合はエラーをログに記録し、処理をスキップ
		fmt.Printf("deleteCastLayer: PictureLayerSet not found, windowID=%d, castID=%d (skipping)\n", cast.WinID, cast.ID)
		return
	}

	// CastLayerを削除（RemoveCastLayerは内部でダーティ領域を追加する）
	if !pls.RemoveCastLayer(cast.ID) {
		// 要件 10.1: CastLayerが見つからない場合はエラーをログに記録
		fmt.Printf("deleteCastLayer: CastLayer not found in PictureLayerSet, windowID=%d, castID=%d (skipping)\n", cast.WinID, cast.ID)
	}
}

// GetCast は指定されたキャストを取得する
// 要件 4.7
func (cm *CastManager) GetCast(id int) (*Cast, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	cast, exists := cm.casts[id]
	if !exists {
		return nil, fmt.Errorf("cast not found: %d", id)
	}

	return cast, nil
}

// GetCastsByWindow は指定されたウィンドウに属するキャストをZ順序でソートして返す
// 要件 4.9
func (cm *CastManager) GetCastsByWindow(winID int) []*Cast {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	casts := make([]*Cast, 0)
	for _, cast := range cm.casts {
		if cast.WinID == winID {
			casts = append(casts, cast)
		}
	}

	// Z順序でソート（小さい順 = 奥から手前）
	sort.Slice(casts, func(i, j int) bool {
		return casts[i].ZOrder < casts[j].ZOrder
	})

	return casts
}

// DeleteCastsByWindow は指定されたウィンドウに属するすべてのキャストを削除する
// 要件 9.2
// 要件 2.6: ウィンドウが閉じられたときにそのウィンドウに属するすべてのレイヤーを削除する
func (cm *CastManager) DeleteCastsByWindow(winID int) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 削除対象のキャストIDを収集
	toDelete := make([]int, 0)
	for id, cast := range cm.casts {
		if cast.WinID == winID {
			toDelete = append(toDelete, id)
		}
	}

	// 要件 2.6: LayerManagerが設定されている場合、CastLayerも削除
	if cm.layerManager != nil {
		pls := cm.layerManager.GetPictureLayerSet(winID)
		if pls != nil {
			// すべてのキャストレイヤーをクリア
			pls.ClearCastLayers()
		}
	}

	// キャストを削除
	for _, id := range toDelete {
		delete(cm.casts, id)
	}
}

// GetCastsOrdered はすべてのキャストをZ順序でソートして返す
func (cm *CastManager) GetCastsOrdered() []*Cast {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	casts := make([]*Cast, 0, len(cm.casts))
	for _, cast := range cm.casts {
		casts = append(casts, cast)
	}

	// Z順序でソート（小さい順 = 奥から手前）
	sort.Slice(casts, func(i, j int) bool {
		return casts[i].ZOrder < casts[j].ZOrder
	})

	return casts
}

// Count は現在のキャスト数を返す
func (cm *CastManager) Count() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.casts)
}
