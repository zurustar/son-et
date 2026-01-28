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
// スプライトシステム移行: LayerManagerは不要になった（CastSpriteで管理）
type CastManager struct {
	casts      map[int]*Cast
	nextID     int
	maxID      int // 最大1024
	nextZOrder int
	mu         sync.RWMutex
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
		casts:      make(map[int]*Cast),
		nextID:     0,
		maxID:      1024, // 要件 9.7
		nextZOrder: 0,
	}
}

// SetLayerManager はLayerManagerを設定する（後方互換性のために残す、何もしない）
// Deprecated: スプライトシステム移行により不要になった
func (cm *CastManager) SetLayerManager(lm *LayerManager) {
	// スプライトシステム移行により、LayerManagerは不要になった
	// この関数は後方互換性のために残すが、何もしない
}

// GetLayerManager はLayerManagerを取得する（後方互換性のために残す、nilを返す）
// Deprecated: スプライトシステム移行により不要になった
func (cm *CastManager) GetLayerManager() *LayerManager {
	// スプライトシステム移行により、LayerManagerは不要になった
	return nil
}

// PutCast はキャストを配置する
// 要件 4.1, 4.2
func (cm *CastManager) PutCast(winID, picID, x, y, srcX, srcY, width, height int) (int, error) {
	return cm.PutCastWithTransColor(winID, picID, x, y, srcX, srcY, width, height, nil)
}

// PutCastWithTransColor は透明色付きでキャストを配置する
// スプライトシステム: CastSpriteはGraphicsSystem.PutCast()で作成される
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

	// スプライトシステム: CastSpriteはGraphicsSystem.PutCast()で作成される
	// LayerManagerのCastLayerは不要になった

	return cast.ID, nil
}

// MoveCast はキャストの位置やソース領域を変更する
// 要件 4.3, 4.4, 4.5
// スプライトシステム: CastSpriteはGraphicsSystem.MoveCast()で更新される
func (cm *CastManager) MoveCast(id int, opts ...CastOption) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// キャストを取得
	cast, exists := cm.casts[id]
	if !exists {
		return fmt.Errorf("cast not found: %d", id)
	}

	// オプションを適用
	for _, opt := range opts {
		opt(cast)
	}

	// スプライトシステム: CastSpriteはGraphicsSystem.MoveCast()で更新される
	// LayerManagerのCastLayerは不要になった

	return nil
}

// DelCast は指定されたキャストを削除する
// 要件 4.6
// スプライトシステム: CastSpriteはGraphicsSystem.DelCast()で削除される
func (cm *CastManager) DelCast(id int) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// キャストが存在するか確認
	_, exists := cm.casts[id]
	if !exists {
		return fmt.Errorf("cast not found: %d", id)
	}

	// スプライトシステム: CastSpriteはGraphicsSystem.DelCast()で削除される
	// LayerManagerのCastLayerは不要になった

	// キャストを削除
	delete(cm.casts, id)

	return nil
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
// スプライトシステム: CastSpriteはGraphicsSystem.CloseWin()で削除される
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

	// スプライトシステム: CastSpriteはGraphicsSystem.CloseWin()で削除される
	// LayerManagerのCastLayerは不要になった

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
