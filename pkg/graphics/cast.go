package graphics

import (
	"sync"
)

// Cast はスプライトを表す
type Cast struct {
	ID      int
	WinID   int
	PicID   int
	X, Y    int
	SrcX    int
	SrcY    int
	Width   int
	Height  int
	Visible bool
	ZOrder  int
}

// CastManager はキャストを管理する
type CastManager struct {
	casts  map[int]*Cast
	nextID int
	maxID  int
	mu     sync.RWMutex
}

// NewCastManager は新しい CastManager を作成する
func NewCastManager() *CastManager {
	return &CastManager{
		casts:  make(map[int]*Cast),
		nextID: 0,
		maxID:  1024,
	}
}

// GetCastsByWindow は指定されたウィンドウに属するキャストを返す
func (cm *CastManager) GetCastsByWindow(winID int) []*Cast {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	casts := make([]*Cast, 0)
	for _, cast := range cm.casts {
		if cast.WinID == winID {
			casts = append(casts, cast)
		}
	}

	return casts
}
