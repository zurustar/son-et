// Package graphics provides the graphics system for FILLY script execution.
// This file implements the headless mode graphics system.
package graphics

import (
	"fmt"
	"image/color"
	"log/slog"
	"sync"
)

// OperationRecord は描画操作の記録を表す
type OperationRecord struct {
	Operation string
	Args      map[string]any
}

// HeadlessGraphicsSystem はヘッドレスモード用のダミー描画システム
// 要件 10.4: ヘッドレスモードが有効のとき、描画操作をログに記録するのみで実際の描画を行わない
type HeadlessGraphicsSystem struct {
	// ピクチャー管理（メモリ上のデータのみ）
	pictures    map[int]*HeadlessPicture
	nextPicID   int
	maxPictures int
	pictureMu   sync.RWMutex

	// ウィンドウ管理
	windows    map[int]*HeadlessWindow
	nextWinID  int
	maxWindows int
	nextZOrder int
	windowMu   sync.RWMutex

	// キャスト管理
	casts      map[int]*HeadlessCast
	nextCastID int
	maxCasts   int
	castMu     sync.RWMutex

	// 描画状態
	paintColor color.Color
	lineSize   int
	textColor  color.Color
	bgColor    color.Color
	backMode   int
	fontName   string
	fontSize   int

	// 仮想デスクトップ
	virtualWidth  int
	virtualHeight int

	// ログ
	log              *slog.Logger
	logOperations    bool // 描画操作をログに記録するかどうか
	recordHistory    bool // 操作履歴を保持するかどうか
	operationHistory []OperationRecord
	historyMu        sync.RWMutex
}

// HeadlessPicture はヘッドレスモード用のピクチャー
type HeadlessPicture struct {
	ID     int
	Width  int
	Height int
}

// HeadlessWindow はヘッドレスモード用のウィンドウ
type HeadlessWindow struct {
	ID      int
	PicID   int
	X, Y    int
	Width   int
	Height  int
	PicX    int
	PicY    int
	BgColor color.Color
	Caption string
	Visible bool
	ZOrder  int
}

// HeadlessCast はヘッドレスモード用のキャスト
type HeadlessCast struct {
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

// HeadlessOption は HeadlessGraphicsSystem のオプションを設定する関数型
type HeadlessOption func(*HeadlessGraphicsSystem)

// WithHeadlessLogger はロガーを設定する
func WithHeadlessLogger(log *slog.Logger) HeadlessOption {
	return func(hgs *HeadlessGraphicsSystem) {
		hgs.log = log
	}
}

// WithHeadlessVirtualSize は仮想デスクトップのサイズを設定する
func WithHeadlessVirtualSize(width, height int) HeadlessOption {
	return func(hgs *HeadlessGraphicsSystem) {
		hgs.virtualWidth = width
		hgs.virtualHeight = height
	}
}

// WithLogOperations は描画操作のログ記録を有効/無効にする
func WithLogOperations(enabled bool) HeadlessOption {
	return func(hgs *HeadlessGraphicsSystem) {
		hgs.logOperations = enabled
	}
}

// WithRecordHistory は操作履歴の記録を有効/無効にする
func WithRecordHistory(enabled bool) HeadlessOption {
	return func(hgs *HeadlessGraphicsSystem) {
		hgs.recordHistory = enabled
	}
}

// NewHeadlessGraphicsSystem は新しいヘッドレスモード用GraphicsSystemを作成する
func NewHeadlessGraphicsSystem(opts ...HeadlessOption) *HeadlessGraphicsSystem {
	hgs := &HeadlessGraphicsSystem{
		pictures:         make(map[int]*HeadlessPicture),
		nextPicID:        0,
		maxPictures:      256, // 要件 9.5
		windows:          make(map[int]*HeadlessWindow),
		nextWinID:        0,
		maxWindows:       64, // 要件 9.6
		nextZOrder:       0,
		casts:            make(map[int]*HeadlessCast),
		nextCastID:       0,
		maxCasts:         1024, // 要件 9.7
		paintColor:       color.RGBA{255, 255, 255, 255},
		lineSize:         1,
		textColor:        color.RGBA{255, 255, 255, 255},
		bgColor:          color.RGBA{0, 0, 0, 255},
		backMode:         0,
		fontName:         "default",
		fontSize:         12,
		virtualWidth:     1024,
		virtualHeight:    768,
		log:              slog.Default(),
		logOperations:    true,
		recordHistory:    false,
		operationHistory: make([]OperationRecord, 0),
	}

	// オプションを適用
	for _, opt := range opts {
		opt(hgs)
	}

	hgs.log.Info("HeadlessGraphicsSystem initialized",
		"virtualWidth", hgs.virtualWidth,
		"virtualHeight", hgs.virtualHeight)

	return hgs
}

// logOperation は描画操作をログに記録する
func (hgs *HeadlessGraphicsSystem) logOperation(operation string, args ...any) {
	if hgs.logOperations {
		hgs.log.Debug(fmt.Sprintf("[Headless] %s", operation), args...)
	}

	// 操作履歴を記録
	if hgs.recordHistory {
		record := OperationRecord{
			Operation: operation,
			Args:      make(map[string]any),
		}
		// argsをkey-valueペアとして解析
		for i := 0; i < len(args)-1; i += 2 {
			if key, ok := args[i].(string); ok {
				record.Args[key] = args[i+1]
			}
		}
		hgs.historyMu.Lock()
		hgs.operationHistory = append(hgs.operationHistory, record)
		hgs.historyMu.Unlock()
	}
}

// GetOperationHistory は操作履歴を返す
func (hgs *HeadlessGraphicsSystem) GetOperationHistory() []OperationRecord {
	hgs.historyMu.RLock()
	defer hgs.historyMu.RUnlock()
	// コピーを返す
	result := make([]OperationRecord, len(hgs.operationHistory))
	copy(result, hgs.operationHistory)
	return result
}

// ClearOperationHistory は操作履歴をクリアする
func (hgs *HeadlessGraphicsSystem) ClearOperationHistory() {
	hgs.historyMu.Lock()
	defer hgs.historyMu.Unlock()
	hgs.operationHistory = make([]OperationRecord, 0)
}

// GetOperationCount は操作履歴の件数を返す
func (hgs *HeadlessGraphicsSystem) GetOperationCount() int {
	hgs.historyMu.RLock()
	defer hgs.historyMu.RUnlock()
	return len(hgs.operationHistory)
}

// Update はゲームループから呼び出される（ヘッドレスモードでは何もしない）
func (hgs *HeadlessGraphicsSystem) Update() error {
	return nil
}

// Draw は画面に描画する（ヘッドレスモードでは何もしない）
func (hgs *HeadlessGraphicsSystem) Draw(screen any) {
	// ヘッドレスモードでは描画しない
}

// Shutdown はGraphicsSystemをシャットダウンする
func (hgs *HeadlessGraphicsSystem) Shutdown() {
	hgs.log.Info("HeadlessGraphicsSystem shutdown")
}

// GetVirtualWidth は仮想デスクトップの幅を返す
func (hgs *HeadlessGraphicsSystem) GetVirtualWidth() int {
	return hgs.virtualWidth
}

// GetVirtualHeight は仮想デスクトップの高さを返す
func (hgs *HeadlessGraphicsSystem) GetVirtualHeight() int {
	return hgs.virtualHeight
}

// ===== Picture Management =====

// LoadPic はピクチャーを読み込む（ヘッドレスモードではダミーピクチャーを作成）
func (hgs *HeadlessGraphicsSystem) LoadPic(filename string) (int, error) {
	hgs.pictureMu.Lock()
	defer hgs.pictureMu.Unlock()

	// リソース制限チェック
	if len(hgs.pictures) >= hgs.maxPictures {
		hgs.log.Error("LoadPic: resource limit reached", "max", hgs.maxPictures)
		return -1, fmt.Errorf("resource limit reached: max %d pictures", hgs.maxPictures)
	}

	// ダミーピクチャーを作成（デフォルトサイズ）
	id := hgs.nextPicID
	hgs.nextPicID++

	pic := &HeadlessPicture{
		ID:     id,
		Width:  640, // デフォルトサイズ
		Height: 480,
	}
	hgs.pictures[id] = pic

	hgs.logOperation("LoadPic", "filename", filename, "picID", id)
	return id, nil
}

// CreatePic は空のピクチャーを作成する
func (hgs *HeadlessGraphicsSystem) CreatePic(width, height int) (int, error) {
	hgs.pictureMu.Lock()
	defer hgs.pictureMu.Unlock()

	// リソース制限チェック
	if len(hgs.pictures) >= hgs.maxPictures {
		hgs.log.Error("CreatePic: resource limit reached", "max", hgs.maxPictures)
		return -1, fmt.Errorf("resource limit reached: max %d pictures", hgs.maxPictures)
	}

	id := hgs.nextPicID
	hgs.nextPicID++

	pic := &HeadlessPicture{
		ID:     id,
		Width:  width,
		Height: height,
	}
	hgs.pictures[id] = pic

	hgs.logOperation("CreatePic", "width", width, "height", height, "picID", id)
	return id, nil
}

// CreatePicFrom は既存のピクチャーからコピーを作成する
func (hgs *HeadlessGraphicsSystem) CreatePicFrom(srcID int) (int, error) {
	hgs.pictureMu.Lock()
	defer hgs.pictureMu.Unlock()

	// リソース制限チェック
	if len(hgs.pictures) >= hgs.maxPictures {
		hgs.log.Error("CreatePicFrom: resource limit reached", "max", hgs.maxPictures)
		return -1, fmt.Errorf("resource limit reached: max %d pictures", hgs.maxPictures)
	}

	// ソースピクチャーを取得
	srcPic, ok := hgs.pictures[srcID]
	if !ok {
		hgs.log.Error("CreatePicFrom: source picture not found", "srcID", srcID)
		return -1, fmt.Errorf("source picture not found: %d", srcID)
	}

	id := hgs.nextPicID
	hgs.nextPicID++

	pic := &HeadlessPicture{
		ID:     id,
		Width:  srcPic.Width,
		Height: srcPic.Height,
	}
	hgs.pictures[id] = pic

	hgs.logOperation("CreatePicFrom", "srcID", srcID, "picID", id)
	return id, nil
}

// CreatePicWithSize は指定されたサイズの空のピクチャーを生成する
// srcID: 参照用のソースピクチャーID（存在確認のみ）
// width, height: 新しいピクチャーのサイズ
// 戻り値: 新しいピクチャーID、エラー
func (hgs *HeadlessGraphicsSystem) CreatePicWithSize(srcID, width, height int) (int, error) {
	hgs.pictureMu.Lock()
	defer hgs.pictureMu.Unlock()

	// リソース制限チェック
	if len(hgs.pictures) >= hgs.maxPictures {
		hgs.log.Error("CreatePicWithSize: resource limit reached", "max", hgs.maxPictures)
		return -1, fmt.Errorf("resource limit reached: max %d pictures", hgs.maxPictures)
	}

	// ソースピクチャーの存在確認
	if _, ok := hgs.pictures[srcID]; !ok {
		hgs.log.Error("CreatePicWithSize: source picture not found", "srcID", srcID)
		return -1, fmt.Errorf("source picture not found: %d", srcID)
	}

	// サイズのバリデーション
	if width <= 0 || height <= 0 {
		hgs.log.Error("CreatePicWithSize: invalid size", "width", width, "height", height)
		return -1, fmt.Errorf("invalid size: width=%d, height=%d", width, height)
	}

	id := hgs.nextPicID
	hgs.nextPicID++

	pic := &HeadlessPicture{
		ID:     id,
		Width:  width,
		Height: height,
	}
	hgs.pictures[id] = pic

	hgs.logOperation("CreatePicWithSize", "srcID", srcID, "width", width, "height", height, "picID", id)
	return id, nil
}

// DelPic はピクチャーを削除する
func (hgs *HeadlessGraphicsSystem) DelPic(id int) error {
	hgs.pictureMu.Lock()
	defer hgs.pictureMu.Unlock()

	if _, ok := hgs.pictures[id]; !ok {
		hgs.log.Warn("DelPic: picture not found", "picID", id)
		return fmt.Errorf("picture not found: %d", id)
	}

	delete(hgs.pictures, id)
	hgs.logOperation("DelPic", "picID", id)
	return nil
}

// PicWidth はピクチャーの幅を返す
func (hgs *HeadlessGraphicsSystem) PicWidth(id int) int {
	hgs.pictureMu.RLock()
	defer hgs.pictureMu.RUnlock()

	if pic, ok := hgs.pictures[id]; ok {
		return pic.Width
	}
	hgs.log.Warn("PicWidth: picture not found", "picID", id)
	return 0
}

// PicHeight はピクチャーの高さを返す
func (hgs *HeadlessGraphicsSystem) PicHeight(id int) int {
	hgs.pictureMu.RLock()
	defer hgs.pictureMu.RUnlock()

	if pic, ok := hgs.pictures[id]; ok {
		return pic.Height
	}
	hgs.log.Warn("PicHeight: picture not found", "picID", id)
	return 0
}

// ===== Picture Transfer =====

// MovePic はピクチャー間で画像を転送する（ヘッドレスモードではログのみ）
func (hgs *HeadlessGraphicsSystem) MovePic(srcID, srcX, srcY, width, height, dstID, dstX, dstY, mode int) error {
	hgs.logOperation("MovePic",
		"srcID", srcID, "srcX", srcX, "srcY", srcY,
		"width", width, "height", height,
		"dstID", dstID, "dstX", dstX, "dstY", dstY,
		"mode", mode)
	return nil
}

// MovePicWithSpeed はピクチャー間で画像を転送する（速度指定付き）
func (hgs *HeadlessGraphicsSystem) MovePicWithSpeed(srcID, srcX, srcY, width, height, dstID, dstX, dstY, mode, speed int) error {
	hgs.logOperation("MovePicWithSpeed",
		"srcID", srcID, "srcX", srcX, "srcY", srcY,
		"width", width, "height", height,
		"dstID", dstID, "dstX", dstX, "dstY", dstY,
		"mode", mode, "speed", speed)
	return nil
}

// MoveSPic は拡大縮小して転送する
func (hgs *HeadlessGraphicsSystem) MoveSPic(srcID, srcX, srcY, srcW, srcH, dstID, dstX, dstY, dstW, dstH int) error {
	hgs.logOperation("MoveSPic",
		"srcID", srcID, "srcX", srcX, "srcY", srcY,
		"srcW", srcW, "srcH", srcH,
		"dstID", dstID, "dstX", dstX, "dstY", dstY,
		"dstW", dstW, "dstH", dstH)
	return nil
}

// TransPic は指定した透明色を除いて転送する
func (hgs *HeadlessGraphicsSystem) TransPic(srcID, srcX, srcY, width, height, dstID, dstX, dstY int, transColor any) error {
	hgs.logOperation("TransPic",
		"srcID", srcID, "srcX", srcX, "srcY", srcY,
		"width", width, "height", height,
		"dstID", dstID, "dstX", dstX, "dstY", dstY,
		"transColor", transColor)
	return nil
}

// ReversePic は左右反転して転送する
func (hgs *HeadlessGraphicsSystem) ReversePic(srcID, srcX, srcY, width, height, dstID, dstX, dstY int) error {
	hgs.logOperation("ReversePic",
		"srcID", srcID, "srcX", srcX, "srcY", srcY,
		"width", width, "height", height,
		"dstID", dstID, "dstX", dstX, "dstY", dstY)
	return nil
}

// ===== Window Management =====

// OpenWin はウィンドウを開く
func (hgs *HeadlessGraphicsSystem) OpenWin(picID int, opts ...any) (int, error) {
	hgs.windowMu.Lock()
	defer hgs.windowMu.Unlock()

	// リソース制限チェック
	if len(hgs.windows) >= hgs.maxWindows {
		hgs.log.Error("OpenWin: resource limit reached", "max", hgs.maxWindows)
		return -1, fmt.Errorf("resource limit reached: max %d windows", hgs.maxWindows)
	}

	id := hgs.nextWinID
	hgs.nextWinID++

	win := &HeadlessWindow{
		ID:      id,
		PicID:   picID,
		X:       0,
		Y:       0,
		Width:   640,
		Height:  480,
		PicX:    0,
		PicY:    0,
		BgColor: color.RGBA{0, 0, 0, 255},
		Caption: "",
		Visible: true,
		ZOrder:  hgs.nextZOrder,
	}
	hgs.nextZOrder++

	// オプションを解析
	if len(opts) >= 2 {
		if x, ok := toIntFromAny(opts[0]); ok {
			if y, ok := toIntFromAny(opts[1]); ok {
				win.X = x
				win.Y = y
			}
		}
	}
	if len(opts) >= 4 {
		if w, ok := toIntFromAny(opts[2]); ok {
			if h, ok := toIntFromAny(opts[3]); ok {
				win.Width = w
				win.Height = h
			}
		}
	}
	if len(opts) >= 6 {
		if picX, ok := toIntFromAny(opts[4]); ok {
			if picY, ok := toIntFromAny(opts[5]); ok {
				win.PicX = picX
				win.PicY = picY
			}
		}
	}

	hgs.windows[id] = win

	hgs.logOperation("OpenWin", "picID", picID, "winID", id, "opts", opts)
	return id, nil
}

// MoveWin はウィンドウを移動/変更する
func (hgs *HeadlessGraphicsSystem) MoveWin(id int, opts ...any) error {
	hgs.windowMu.Lock()
	defer hgs.windowMu.Unlock()

	win, ok := hgs.windows[id]
	if !ok {
		hgs.log.Warn("MoveWin: window not found", "winID", id)
		return fmt.Errorf("window not found: %d", id)
	}

	// オプションを解析して適用
	if len(opts) >= 1 {
		if picID, ok := toIntFromAny(opts[0]); ok {
			win.PicID = picID
		}
	}
	if len(opts) >= 3 {
		if x, ok := toIntFromAny(opts[1]); ok {
			if y, ok := toIntFromAny(opts[2]); ok {
				win.X = x
				win.Y = y
			}
		}
	}
	if len(opts) >= 5 {
		if w, ok := toIntFromAny(opts[3]); ok {
			if h, ok := toIntFromAny(opts[4]); ok {
				win.Width = w
				win.Height = h
			}
		}
	}
	if len(opts) >= 7 {
		if picX, ok := toIntFromAny(opts[5]); ok {
			if picY, ok := toIntFromAny(opts[6]); ok {
				win.PicX = picX
				win.PicY = picY
			}
		}
	}

	hgs.logOperation("MoveWin", "winID", id, "opts", opts)
	return nil
}

// CloseWin はウィンドウを閉じる
func (hgs *HeadlessGraphicsSystem) CloseWin(id int) error {
	hgs.windowMu.Lock()
	defer hgs.windowMu.Unlock()

	if _, ok := hgs.windows[id]; !ok {
		hgs.log.Warn("CloseWin: window not found", "winID", id)
		return fmt.Errorf("window not found: %d", id)
	}

	delete(hgs.windows, id)

	// このウィンドウに属するキャストを削除
	hgs.castMu.Lock()
	for castID, cast := range hgs.casts {
		if cast.WinID == id {
			delete(hgs.casts, castID)
		}
	}
	hgs.castMu.Unlock()

	hgs.logOperation("CloseWin", "winID", id)
	return nil
}

// CloseWinAll はすべてのウィンドウを閉じる
func (hgs *HeadlessGraphicsSystem) CloseWinAll() {
	hgs.windowMu.Lock()
	defer hgs.windowMu.Unlock()

	hgs.windows = make(map[int]*HeadlessWindow)
	hgs.nextWinID = 0
	hgs.nextZOrder = 0

	// すべてのキャストも削除
	hgs.castMu.Lock()
	hgs.casts = make(map[int]*HeadlessCast)
	hgs.nextCastID = 0
	hgs.castMu.Unlock()

	hgs.logOperation("CloseWinAll")
}

// CapTitle はウィンドウのキャプションを設定する
func (hgs *HeadlessGraphicsSystem) CapTitle(id int, title string) error {
	hgs.windowMu.Lock()
	defer hgs.windowMu.Unlock()

	win, ok := hgs.windows[id]
	if !ok {
		hgs.log.Warn("CapTitle: window not found", "winID", id)
		return fmt.Errorf("window not found: %d", id)
	}

	win.Caption = title
	hgs.logOperation("CapTitle", "winID", id, "title", title)
	return nil
}

// CapTitleAll は全てのウィンドウのキャプションを設定する
// ウィンドウが存在しない場合は何もしない（エラーなし）
// 受け入れ基準 3.1, 3.2
func (hgs *HeadlessGraphicsSystem) CapTitleAll(title string) {
	hgs.windowMu.Lock()
	defer hgs.windowMu.Unlock()

	for _, win := range hgs.windows {
		win.Caption = title
	}
	hgs.logOperation("CapTitleAll", "title", title, "windowCount", len(hgs.windows))
}

// GetPicNo はウィンドウに関連付けられたピクチャー番号を返す
func (hgs *HeadlessGraphicsSystem) GetPicNo(id int) (int, error) {
	hgs.windowMu.RLock()
	defer hgs.windowMu.RUnlock()

	win, ok := hgs.windows[id]
	if !ok {
		hgs.log.Warn("GetPicNo: window not found", "winID", id)
		return -1, fmt.Errorf("window not found: %d", id)
	}

	return win.PicID, nil
}

// GetWinByPicID はピクチャーIDに関連付けられたウィンドウIDを返す
func (hgs *HeadlessGraphicsSystem) GetWinByPicID(picID int) (int, error) {
	hgs.windowMu.RLock()
	defer hgs.windowMu.RUnlock()

	var foundWin *HeadlessWindow
	for _, win := range hgs.windows {
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

// ===== Cast Management =====

// PutCast はキャストを配置する
func (hgs *HeadlessGraphicsSystem) PutCast(winID, picID, x, y, srcX, srcY, w, h int) (int, error) {
	return hgs.PutCastWithTransColor(winID, picID, x, y, srcX, srcY, w, h, nil)
}

// PutCastWithTransColor は透明色付きでキャストを配置する
func (hgs *HeadlessGraphicsSystem) PutCastWithTransColor(winID, picID, x, y, srcX, srcY, w, h int, transColor color.Color) (int, error) {
	hgs.castMu.Lock()
	defer hgs.castMu.Unlock()

	// リソース制限チェック
	if len(hgs.casts) >= hgs.maxCasts {
		hgs.log.Error("PutCast: resource limit reached", "max", hgs.maxCasts)
		return -1, fmt.Errorf("resource limit reached: max %d casts", hgs.maxCasts)
	}

	id := hgs.nextCastID
	hgs.nextCastID++

	cast := &HeadlessCast{
		ID:      id,
		WinID:   winID,
		PicID:   picID,
		X:       x,
		Y:       y,
		SrcX:    srcX,
		SrcY:    srcY,
		Width:   w,
		Height:  h,
		Visible: true,
		ZOrder:  id, // 簡易的にIDをZOrderとして使用
	}
	hgs.casts[id] = cast

	hgs.logOperation("PutCast",
		"winID", winID, "picID", picID,
		"x", x, "y", y,
		"srcX", srcX, "srcY", srcY,
		"w", w, "h", h,
		"transColor", transColor,
		"castID", id)
	return id, nil
}

// MoveCast はキャストを移動する
func (hgs *HeadlessGraphicsSystem) MoveCast(id int, opts ...any) error {
	hgs.castMu.Lock()
	defer hgs.castMu.Unlock()

	cast, ok := hgs.casts[id]
	if !ok {
		hgs.log.Warn("MoveCast: cast not found", "castID", id)
		return fmt.Errorf("cast not found: %d", id)
	}

	// オプションを解析して適用
	if len(opts) >= 2 {
		if x, ok := toIntFromAny(opts[0]); ok {
			if y, ok := toIntFromAny(opts[1]); ok {
				cast.X = x
				cast.Y = y
			}
		}
	}
	if len(opts) >= 6 {
		if srcX, ok := toIntFromAny(opts[2]); ok {
			if srcY, ok := toIntFromAny(opts[3]); ok {
				if w, ok := toIntFromAny(opts[4]); ok {
					if h, ok := toIntFromAny(opts[5]); ok {
						cast.SrcX = srcX
						cast.SrcY = srcY
						cast.Width = w
						cast.Height = h
					}
				}
			}
		}
	}
	// 3引数パターン: pic_no, x, y
	if len(opts) == 3 {
		if picID, ok := toIntFromAny(opts[0]); ok {
			if x, ok := toIntFromAny(opts[1]); ok {
				if y, ok := toIntFromAny(opts[2]); ok {
					cast.PicID = picID
					cast.X = x
					cast.Y = y
				}
			}
		}
	}

	hgs.logOperation("MoveCast", "castID", id, "opts", opts)
	return nil
}

// MoveCastWithOptions はキャストを移動する（CastOptionを使用）
func (hgs *HeadlessGraphicsSystem) MoveCastWithOptions(id int, opts ...CastOption) error {
	hgs.castMu.Lock()
	defer hgs.castMu.Unlock()

	cast, ok := hgs.casts[id]
	if !ok {
		hgs.log.Warn("MoveCastWithOptions: cast not found", "castID", id)
		return fmt.Errorf("cast not found: %d", id)
	}

	// CastOptionを適用
	tempCast := &Cast{
		ID:      cast.ID,
		WinID:   cast.WinID,
		PicID:   cast.PicID,
		X:       cast.X,
		Y:       cast.Y,
		SrcX:    cast.SrcX,
		SrcY:    cast.SrcY,
		Width:   cast.Width,
		Height:  cast.Height,
		Visible: cast.Visible,
		ZOrder:  cast.ZOrder,
	}
	for _, opt := range opts {
		opt(tempCast)
	}

	// 更新を適用
	cast.PicID = tempCast.PicID
	cast.X = tempCast.X
	cast.Y = tempCast.Y
	cast.SrcX = tempCast.SrcX
	cast.SrcY = tempCast.SrcY
	cast.Width = tempCast.Width
	cast.Height = tempCast.Height

	hgs.logOperation("MoveCastWithOptions", "castID", id)
	return nil
}

// DelCast はキャストを削除する
func (hgs *HeadlessGraphicsSystem) DelCast(id int) error {
	hgs.castMu.Lock()
	defer hgs.castMu.Unlock()

	if _, ok := hgs.casts[id]; !ok {
		hgs.log.Warn("DelCast: cast not found", "castID", id)
		return fmt.Errorf("cast not found: %d", id)
	}

	delete(hgs.casts, id)
	hgs.logOperation("DelCast", "castID", id)
	return nil
}

// ===== Text Rendering =====

// TextWrite はテキストを描画する（ヘッドレスモードではログのみ）
func (hgs *HeadlessGraphicsSystem) TextWrite(picID, x, y int, text string) error {
	hgs.logOperation("TextWrite", "picID", picID, "x", x, "y", y, "text", text)
	return nil
}

// SetFont はフォントを設定する
func (hgs *HeadlessGraphicsSystem) SetFont(name string, size int, opts ...any) error {
	hgs.fontName = name
	hgs.fontSize = size
	hgs.logOperation("SetFont", "name", name, "size", size, "opts", opts)
	return nil
}

// SetTextColor はテキスト色を設定する
func (hgs *HeadlessGraphicsSystem) SetTextColor(c any) error {
	switch v := c.(type) {
	case int:
		hgs.textColor = ColorFromInt(v)
	case color.Color:
		hgs.textColor = v
	}
	hgs.logOperation("SetTextColor", "color", c)
	return nil
}

// SetBgColor は背景色を設定する
func (hgs *HeadlessGraphicsSystem) SetBgColor(c any) error {
	switch v := c.(type) {
	case int:
		hgs.bgColor = ColorFromInt(v)
	case color.Color:
		hgs.bgColor = v
	}
	hgs.logOperation("SetBgColor", "color", c)
	return nil
}

// SetBackMode は背景モードを設定する
func (hgs *HeadlessGraphicsSystem) SetBackMode(mode int) error {
	hgs.backMode = mode
	hgs.logOperation("SetBackMode", "mode", mode)
	return nil
}

// ===== Drawing Primitives =====

// DrawLine は直線を描画する（ヘッドレスモードではログのみ）
func (hgs *HeadlessGraphicsSystem) DrawLine(picID, x1, y1, x2, y2 int) error {
	hgs.logOperation("DrawLine", "picID", picID, "x1", x1, "y1", y1, "x2", x2, "y2", y2)
	return nil
}

// DrawRect は矩形を描画する（ヘッドレスモードではログのみ）
func (hgs *HeadlessGraphicsSystem) DrawRect(picID, x1, y1, x2, y2, fillMode int) error {
	hgs.logOperation("DrawRect", "picID", picID, "x1", x1, "y1", y1, "x2", x2, "y2", y2, "fillMode", fillMode)
	return nil
}

// FillRect は矩形を塗りつぶす（ヘッドレスモードではログのみ）
func (hgs *HeadlessGraphicsSystem) FillRect(picID, x1, y1, x2, y2 int, c any) error {
	hgs.logOperation("FillRect", "picID", picID, "x1", x1, "y1", y1, "x2", x2, "y2", y2, "color", c)
	return nil
}

// DrawCircle は円を描画する（ヘッドレスモードではログのみ）
func (hgs *HeadlessGraphicsSystem) DrawCircle(picID, x, y, radius, fillMode int) error {
	hgs.logOperation("DrawCircle", "picID", picID, "x", x, "y", y, "radius", radius, "fillMode", fillMode)
	return nil
}

// SetLineSize は線の太さを設定する
func (hgs *HeadlessGraphicsSystem) SetLineSize(size int) {
	hgs.lineSize = size
	hgs.logOperation("SetLineSize", "size", size)
}

// SetPaintColor は描画色を設定する
func (hgs *HeadlessGraphicsSystem) SetPaintColor(c any) error {
	switch v := c.(type) {
	case int:
		hgs.paintColor = ColorFromInt(v)
	case color.Color:
		hgs.paintColor = v
	}
	hgs.logOperation("SetPaintColor", "color", c)
	return nil
}

// GetColor は指定座標のピクセル色を取得する（ヘッドレスモードでは0を返す）
func (hgs *HeadlessGraphicsSystem) GetColor(picID, x, y int) (int, error) {
	hgs.logOperation("GetColor", "picID", picID, "x", x, "y", y)
	return 0, nil
}
