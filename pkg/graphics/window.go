package graphics

import (
	"fmt"
	"image/color"
	"sort"
	"sync"
)

// Window は仮想ウィンドウを表す
type Window struct {
	ID      int
	PicID   int         // 関連付けられたピクチャー
	X, Y    int         // 仮想デスクトップ上の位置
	Width   int         // 表示幅
	Height  int         // 表示高さ
	PicX    int         // ピクチャー内の参照X
	PicY    int         // ピクチャー内の参照Y
	BgColor color.Color // 背景色
	Caption string      // キャプション
	Visible bool
	ZOrder  int   // Z順序（大きいほど前面）
	Casts   []int // このウィンドウに属するキャストID
}

// WindowManager はウィンドウを管理する
type WindowManager struct {
	windows    map[int]*Window
	nextID     int
	maxID      int // 最大64
	nextZOrder int
	mu         sync.RWMutex
}

// WinOption はウィンドウのオプションを設定する関数型
type WinOption func(*Window)

// WithPosition はウィンドウの位置を設定する
func WithPosition(x, y int) WinOption {
	return func(w *Window) {
		w.X = x
		w.Y = y
	}
}

// WithSize はウィンドウのサイズを設定する
func WithSize(width, height int) WinOption {
	return func(w *Window) {
		w.Width = width
		w.Height = height
	}
}

// WithPicOffset はピクチャー内の参照位置を設定する
func WithPicOffset(picX, picY int) WinOption {
	return func(w *Window) {
		w.PicX = picX
		w.PicY = picY
	}
}

// WithBgColor は背景色を設定する
func WithBgColor(c color.Color) WinOption {
	return func(w *Window) {
		w.BgColor = c
	}
}

// WithPicID はピクチャーIDを設定する
func WithPicID(picID int) WinOption {
	return func(w *Window) {
		w.PicID = picID
	}
}

// WithCaption はキャプションを設定する
func WithCaption(caption string) WinOption {
	return func(w *Window) {
		w.Caption = caption
	}
}

// NewWindowManager は新しい WindowManager を作成する
func NewWindowManager() *WindowManager {
	return &WindowManager{
		windows:    make(map[int]*Window),
		nextID:     0,
		maxID:      64, // 要件 9.6
		nextZOrder: 0,
	}
}

// OpenWin はウィンドウを開く
// 受け入れ基準 3.1, 3.2, 3.3
func (wm *WindowManager) OpenWin(picID int, opts ...WinOption) (int, error) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	// リソース制限チェック（要件 9.6）
	if len(wm.windows) >= wm.maxID {
		return -1, fmt.Errorf("window limit reached (max: %d)", wm.maxID)
	}

	// 新しいウィンドウを作成
	win := &Window{
		ID:      wm.nextID,
		PicID:   picID,
		X:       0,
		Y:       0,
		Width:   0, // デフォルトはピクチャー全体
		Height:  0,
		PicX:    0,
		PicY:    0,
		BgColor: color.RGBA{0, 0, 0, 255}, // デフォルトは黒
		Caption: "",
		Visible: true,
		ZOrder:  wm.nextZOrder,
		Casts:   make([]int, 0),
	}

	// オプションを適用
	for _, opt := range opts {
		opt(win)
	}

	// ウィンドウを登録
	wm.windows[win.ID] = win
	wm.nextID++
	wm.nextZOrder++

	return win.ID, nil
}

// MoveWin はウィンドウの位置、サイズ、ピクチャーを変更する
// 受け入れ基準 3.4, 3.5
func (wm *WindowManager) MoveWin(id int, opts ...WinOption) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	// ウィンドウを取得
	win, exists := wm.windows[id]
	if !exists {
		return fmt.Errorf("window not found: %d", id)
	}

	// オプションを適用
	for _, opt := range opts {
		opt(win)
	}

	return nil
}

// CloseWin は指定されたウィンドウを閉じる
// 受け入れ基準 3.6
func (wm *WindowManager) CloseWin(id int) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	// ウィンドウが存在するか確認
	if _, exists := wm.windows[id]; !exists {
		return fmt.Errorf("window not found: %d", id)
	}

	// ウィンドウを削除
	delete(wm.windows, id)

	return nil
}

// CloseWinAll はすべてのウィンドウを閉じる
// 受け入れ基準 3.7
func (wm *WindowManager) CloseWinAll() {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	// すべてのウィンドウを削除
	wm.windows = make(map[int]*Window)
}

// GetWin は指定されたウィンドウを取得する
// 受け入れ基準 3.10
func (wm *WindowManager) GetWin(id int) (*Window, error) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	win, exists := wm.windows[id]
	if !exists {
		return nil, fmt.Errorf("window not found: %d", id)
	}

	return win, nil
}

// GetWindowsOrdered はウィンドウをZ順序でソートして返す
// 受け入れ基準 3.11
func (wm *WindowManager) GetWindowsOrdered() []*Window {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	// ウィンドウをスライスに変換
	windows := make([]*Window, 0, len(wm.windows))
	for _, win := range wm.windows {
		windows = append(windows, win)
	}

	// Z順序でソート（小さい順 = 奥から手前）
	sort.Slice(windows, func(i, j int) bool {
		return windows[i].ZOrder < windows[j].ZOrder
	})

	return windows
}

// CapTitle はウィンドウのキャプションを設定する
// 受け入れ基準 3.8
func (wm *WindowManager) CapTitle(id int, title string) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	win, exists := wm.windows[id]
	if !exists {
		return fmt.Errorf("window not found: %d", id)
	}

	win.Caption = title
	return nil
}

// GetPicNo はウィンドウに関連付けられたピクチャー番号を返す
// 受け入れ基準 3.9
func (wm *WindowManager) GetPicNo(id int) (int, error) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	win, exists := wm.windows[id]
	if !exists {
		return -1, fmt.Errorf("window not found: %d", id)
	}

	return win.PicID, nil
}

// CapTitleAll は全てのウィンドウのキャプションを設定する
// title: 設定するキャプション
// ウィンドウが存在しない場合は何もしない（エラーなし）
// 受け入れ基準 3.1, 3.2
func (wm *WindowManager) CapTitleAll(title string) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	// 全てのウィンドウのキャプションを設定
	for _, win := range wm.windows {
		win.Caption = title
	}
}

// GetWinByPicID はピクチャーIDに関連付けられたウィンドウIDを返す
// 複数のウィンドウが同じピクチャーを使用している場合、最後に開かれたウィンドウを返す
func (wm *WindowManager) GetWinByPicID(picID int) (int, error) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	var foundWin *Window
	for _, win := range wm.windows {
		if win.PicID == picID {
			if foundWin == nil || win.ZOrder > foundWin.ZOrder {
				foundWin = win
			}
		}
	}

	if foundWin == nil {
		return -1, fmt.Errorf("no window found for picture: %d", picID)
	}

	return foundWin.ID, nil
}
