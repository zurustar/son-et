package graphics

import (
	"fmt"
	"image"
	"image/color"
	"log/slog"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// GraphicsSystem は描画システム全体を管理する
type GraphicsSystem struct {
	pictures     *PictureManager
	windows      *WindowManager
	casts        *CastManager
	textRenderer *TextRenderer
	cmdQueue     *CommandQueue
	sceneChanges *SceneChangeManager
	debugOverlay *DebugOverlay
	layerManager *LayerManager // 要件 8.1: LayerManagerを統合

	// 仮想デスクトップ
	virtualWidth  int
	virtualHeight int

	// 描画状態
	paintColor color.Color
	lineSize   int

	// ログ
	log *slog.Logger
	mu  sync.RWMutex
}

// Option は GraphicsSystem のオプションを設定する関数型
type Option func(*GraphicsSystem)

// WithLogger はロガーを設定する
func WithLogger(log *slog.Logger) Option {
	return func(gs *GraphicsSystem) {
		gs.log = log
	}
}

// WithVirtualSize は仮想デスクトップのサイズを設定する
func WithVirtualSize(width, height int) Option {
	return func(gs *GraphicsSystem) {
		gs.virtualWidth = width
		gs.virtualHeight = height
	}
}

// WithBasePath は画像ファイルの基準パスを設定する
func WithBasePath(basePath string) Option {
	return func(gs *GraphicsSystem) {
		gs.pictures.basePath = basePath
	}
}

// WithDebugOverlay はデバッグオーバーレイの有効/無効を設定する
// 要件 15.7, 15.8: ログレベルに基づいた表示/非表示の切り替え
func WithDebugOverlay(enabled bool) Option {
	return func(gs *GraphicsSystem) {
		if gs.debugOverlay != nil {
			gs.debugOverlay.SetEnabled(enabled)
		}
	}
}

// NewGraphicsSystem は新しい GraphicsSystem を作成する
func NewGraphicsSystem(basePath string, opts ...Option) *GraphicsSystem {
	gs := &GraphicsSystem{
		virtualWidth:  1024, // skelton要件に合わせて1024x768
		virtualHeight: 768,
		paintColor:    color.RGBA{255, 255, 255, 255}, // デフォルトは白
		lineSize:      1,
		log:           slog.Default(),
	}

	// サブシステムを初期化
	gs.pictures = NewPictureManager(basePath)
	gs.windows = NewWindowManager()
	gs.casts = NewCastManager()
	gs.textRenderer = NewTextRenderer()
	gs.cmdQueue = NewCommandQueue()
	gs.sceneChanges = NewSceneChangeManager()
	gs.debugOverlay = NewDebugOverlay()
	gs.layerManager = NewLayerManager() // 要件 8.1: LayerManagerを初期化

	// 要件 8.2: CastManagerとLayerManagerを統合
	gs.casts.SetLayerManager(gs.layerManager)

	// 要件 8.3: TextRendererとLayerManagerを統合
	gs.textRenderer.SetLayerManager(gs.layerManager)

	// オプションを適用
	for _, opt := range opts {
		opt(gs)
	}

	gs.log.Info("GraphicsSystem initialized",
		"virtualWidth", gs.virtualWidth,
		"virtualHeight", gs.virtualHeight,
		"basePath", basePath)

	return gs
}

// Update はゲームループから呼び出され、コマンドキューを処理する
// Ebitengineのメインスレッドで実行される
func (gs *GraphicsSystem) Update() error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	// コマンドキューからすべてのコマンドを取得
	commands := gs.cmdQueue.PopAll()

	// コマンドを順次実行
	for _, cmd := range commands {
		if err := gs.executeCommand(cmd); err != nil {
			// エラーをログに記録して継続（要件 7.7）
			gs.log.Error("Failed to execute command",
				"type", cmd.Type,
				"error", err)
		}
	}

	// シーンチェンジを更新（要件 13.11: 非同期実行）
	gs.sceneChanges.Update()

	return nil
}

// executeCommand は個別のコマンドを実行する
func (gs *GraphicsSystem) executeCommand(cmd Command) error {
	// コマンドタイプに応じて処理を分岐
	// 実際の実装は各フェーズで追加される
	switch cmd.Type {
	case CmdMovePic:
		// TODO: フェーズ5で実装
		gs.log.Debug("MovePic command", "args", cmd.Args)
	case CmdMoveSPic:
		// TODO: フェーズ5で実装
		gs.log.Debug("MoveSPic command", "args", cmd.Args)
	case CmdTransPic:
		// TODO: フェーズ5で実装
		gs.log.Debug("TransPic command", "args", cmd.Args)
	case CmdReversePic:
		// TODO: フェーズ5で実装
		gs.log.Debug("ReversePic command", "args", cmd.Args)
	case CmdOpenWin:
		// TODO: フェーズ3で実装
		gs.log.Debug("OpenWin command", "args", cmd.Args)
	case CmdMoveWin:
		// TODO: フェーズ3で実装
		gs.log.Debug("MoveWin command", "args", cmd.Args)
	case CmdCloseWin:
		// TODO: フェーズ3で実装
		gs.log.Debug("CloseWin command", "args", cmd.Args)
	case CmdPutCast:
		// TODO: フェーズ4で実装
		gs.log.Debug("PutCast command", "args", cmd.Args)
	case CmdMoveCast:
		// TODO: フェーズ4で実装
		gs.log.Debug("MoveCast command", "args", cmd.Args)
	case CmdDelCast:
		// TODO: フェーズ4で実装
		gs.log.Debug("DelCast command", "args", cmd.Args)
	case CmdTextWrite:
		// TODO: フェーズ6で実装
		gs.log.Debug("TextWrite command", "args", cmd.Args)
	case CmdDrawLine:
		// TODO: フェーズ7で実装
		gs.log.Debug("DrawLine command", "args", cmd.Args)
	case CmdDrawRect:
		// TODO: フェーズ7で実装
		gs.log.Debug("DrawRect command", "args", cmd.Args)
	case CmdFillRect:
		// TODO: フェーズ7で実装
		gs.log.Debug("FillRect command", "args", cmd.Args)
	case CmdDrawCircle:
		// TODO: フェーズ7で実装
		gs.log.Debug("DrawCircle command", "args", cmd.Args)
	default:
		gs.log.Warn("Unknown command type", "type", cmd.Type)
	}

	return nil
}

// Draw は画面に描画する
// Ebitengineのメインスレッドで実行される
// 要件 3.11: ウィンドウをZ順序で管理し、後から開いたウィンドウを前面に表示する
// 要件 4.8: キャストを透明色（黒 0x000000）を除いて描画する
// 要件 4.9: キャストをZ順序で管理し、後から配置したキャストを前面に表示する
// 要件 15.1-15.8: デバッグオーバーレイの描画
func (gs *GraphicsSystem) Draw(screen *ebiten.Image) {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	// 背景色は window.go の drawDesktop() で既に設定されているので、
	// ここでは塗りつぶさない

	// ウィンドウをZ順序で取得（要件 3.11）
	windows := gs.windows.GetWindowsOrdered()

	// 各ウィンドウを描画
	for _, win := range windows {
		if !win.Visible {
			continue
		}

		// ピクチャーを取得
		pic, err := gs.pictures.GetPicWithoutLock(win.PicID)
		if err != nil {
			gs.log.Warn("Failed to get picture for window",
				"windowID", win.ID,
				"pictureID", win.PicID,
				"error", err)
			continue
		}

		// ウィンドウ装飾を描画（Windows 3.1風）
		gs.drawWindowDecoration(screen, win, pic)

		// このウィンドウに属するキャストを描画（要件 4.9: Z順序でソート済み）
		gs.drawCastsForWindow(screen, win)

		// デバッグオーバーレイを描画（要件 15.1-15.8）
		gs.drawDebugOverlayForWindow(screen, win, pic)
	}
}

// drawCastsForWindow はウィンドウに属するキャストを描画する
// 要件 4.8: キャストを透明色（黒 0x000000）を除いて描画する
// 要件 4.9: キャストをZ順序で管理し、後から配置したキャストを前面に表示する
// 要件 4.10: キャストの位置をウィンドウ相対座標で管理する
func (gs *GraphicsSystem) drawCastsForWindow(screen *ebiten.Image, win *Window) {
	const (
		borderThickness = 4
		titleBarHeight  = 20
	)

	// コンテンツ領域の開始位置を計算
	contentX := win.X + borderThickness
	contentY := win.Y + borderThickness + titleBarHeight

	// キャストの位置はピクチャー座標系で指定される
	// PicX/PicYはピクチャーの表示オフセットなので、キャストの位置にも適用する
	// PicXが負の場合、ピクチャーは右にシフトされるので、キャストも同様にシフト
	castOffsetX := -win.PicX
	castOffsetY := -win.PicY

	// このウィンドウに属するキャストを取得（Z順序でソート済み）
	casts := gs.casts.GetCastsByWindow(win.ID)

	// デバッグログは頻繁すぎるので削除
	// gs.log.Debug("drawCastsForWindow", "winID", win.ID, "castCount", len(casts))

	for _, cast := range casts {
		if !cast.Visible {
			continue
		}

		// キャストのピクチャーを取得
		castPic, err := gs.pictures.GetPicWithoutLock(cast.PicID)
		if err != nil {
			gs.log.Warn("Failed to get picture for cast",
				"castID", cast.ID,
				"pictureID", cast.PicID,
				"error", err)
			continue
		}

		// キャストのソース領域を切り出す
		srcX := cast.SrcX
		srcY := cast.SrcY
		srcW := cast.Width
		srcH := cast.Height

		// ソース領域のクリッピング
		if srcX < 0 {
			srcW += srcX
			srcX = 0
		}
		if srcY < 0 {
			srcH += srcY
			srcY = 0
		}
		if srcX+srcW > castPic.Width {
			srcW = castPic.Width - srcX
		}
		if srcY+srcH > castPic.Height {
			srcH = castPic.Height - srcY
		}

		// サイズが0以下なら描画しない
		if srcW <= 0 || srcH <= 0 {
			continue
		}

		// ソース領域を切り出す
		srcRect := image.Rect(srcX, srcY, srcX+srcW, srcY+srcH)
		subImg := castPic.Image.SubImage(srcRect).(*ebiten.Image)

		// キャストの描画位置を計算（ピクチャー座標系 → スクリーン座標）
		// キャストの位置はピクチャー座標系で指定されるので、PicX/PicYオフセットを適用
		screenX := contentX + castOffsetX + cast.X
		screenY := contentY + castOffsetY + cast.Y

		// キャストを描画（要件 4.8: 透明色除外）
		gs.drawCastWithTransparency(screen, subImg, screenX, screenY, cast.TransColor, cast.HasTransColor)
	}
}

// drawCastWithTransparency はキャストを透明色を除いて描画する
// 要件 4.8: キャストを透明色を除いて描画する
func (gs *GraphicsSystem) drawCastWithTransparency(screen *ebiten.Image, src *ebiten.Image, dstX, dstY int, transColor color.Color, hasTransColor bool) {
	if !hasTransColor {
		// 透明色が設定されていない場合は通常描画
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(float64(dstX), float64(dstY))
		screen.DrawImage(src, opts)
		return
	}

	// 透明色が設定されている場合、ピクセル単位で透明色処理
	if err := drawImageWithColorKey(screen, src, dstX, dstY, transColor); err != nil {
		// エラーの場合はフォールバック（通常描画）
		gs.log.Warn("Failed to draw with color key, falling back to normal draw",
			"error", err)
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(float64(dstX), float64(dstY))
		screen.DrawImage(src, opts)
	}
}

// drawDebugOverlayForWindow はウィンドウのデバッグオーバーレイを描画する
// 要件 15.1-15.8: デバッグオーバーレイの実装
func (gs *GraphicsSystem) drawDebugOverlayForWindow(screen *ebiten.Image, win *Window, pic *Picture) {
	if gs.debugOverlay == nil || !gs.debugOverlay.IsEnabled() {
		return
	}

	const (
		borderThickness = 4
		titleBarHeight  = 20
	)

	// ウィンドウの実際のサイズを計算
	winWidth := pic.Width
	if win.Width > 0 {
		winWidth = win.Width
	}

	// タイトルバーの位置とサイズ
	titleBarX := win.X + borderThickness
	titleBarY := win.Y + borderThickness
	titleBarWidth := winWidth

	// ウィンドウIDをタイトルバーに描画（要件 15.1）
	gs.debugOverlay.DrawWindowID(screen, win, titleBarX, titleBarY, titleBarWidth)

	// コンテンツ領域の開始位置を計算
	contentX := win.X + borderThickness
	contentY := win.Y + borderThickness + titleBarHeight

	// ピクチャーIDをコンテンツ領域の左上に描画（要件 15.2）
	gs.debugOverlay.DrawPictureID(screen, win.PicID, contentX+2, contentY+2)

	// キャストの位置はピクチャー座標系で指定される
	// PicX/PicYはピクチャーの表示オフセットなので、キャストの位置にも適用する
	castOffsetX := -win.PicX
	castOffsetY := -win.PicY

	// このウィンドウに属するキャストのデバッグ情報を描画（要件 15.3）
	casts := gs.casts.GetCastsByWindow(win.ID)
	for _, cast := range casts {
		if !cast.Visible {
			continue
		}

		// キャストの描画位置を計算（ピクチャー座標系 → スクリーン座標）
		// キャストの位置はピクチャー座標系で指定されるので、PicX/PicYオフセットを適用
		castScreenX := contentX + castOffsetX + cast.X
		castScreenY := contentY + castOffsetY + cast.Y

		// キャストIDを描画
		gs.debugOverlay.DrawCastID(screen, cast, castScreenX, castScreenY)
	}
}

// Shutdown はGraphicsSystemをシャットダウンし、すべてのリソースを解放する
func (gs *GraphicsSystem) Shutdown() {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	gs.log.Info("Shutting down GraphicsSystem")

	// すべてのウィンドウを閉じる（関連するキャストも削除される）
	gs.windows.CloseWinAll()

	// すべてのピクチャーを削除
	if gs.pictures != nil {
		for id := range gs.pictures.pictures {
			if err := gs.pictures.DelPic(id); err != nil {
				gs.log.Warn("Failed to delete picture during shutdown",
					"pictureID", id,
					"error", err)
			}
		}
	}

	// コマンドキューをクリア
	if gs.cmdQueue != nil {
		gs.cmdQueue.PopAll()
	}

	gs.log.Info("GraphicsSystem shutdown complete")
}

// SetDebugOverlayEnabled はデバッグオーバーレイの有効/無効を設定する
// 要件 15.7, 15.8: ログレベルに基づいた表示/非表示の切り替え
func (gs *GraphicsSystem) SetDebugOverlayEnabled(enabled bool) {
	if gs.debugOverlay != nil {
		gs.debugOverlay.SetEnabled(enabled)
	}
}

// SetDebugOverlayFromLogLevel はログレベルに基づいてデバッグオーバーレイの有効/無効を設定する
// 要件 15.1, 15.7: ログレベルがDebug以上のとき、デバッグオーバーレイを表示する
func (gs *GraphicsSystem) SetDebugOverlayFromLogLevel(level slog.Level) {
	if gs.debugOverlay != nil {
		gs.debugOverlay.SetEnabledFromLogLevel(level)
	}
}

// SetDebugOverlayFromLogLevelString はログレベル文字列に基づいてデバッグオーバーレイの有効/無効を設定する
// 要件 15.1, 15.7: ログレベルがDebug以上のとき、デバッグオーバーレイを表示する
func (gs *GraphicsSystem) SetDebugOverlayFromLogLevelString(level string) {
	if gs.debugOverlay != nil {
		gs.debugOverlay.SetEnabledFromLogLevelString(level)
	}
}

// IsDebugOverlayEnabled はデバッグオーバーレイが有効かどうかを返す
func (gs *GraphicsSystem) IsDebugOverlayEnabled() bool {
	if gs.debugOverlay != nil {
		return gs.debugOverlay.IsEnabled()
	}
	return false
}

// GetLayerManager はLayerManagerを返す
// 要件 8.1: GraphicsSystemにLayerManagerを統合する
func (gs *GraphicsSystem) GetLayerManager() *LayerManager {
	return gs.layerManager
}

// VM Interface Implementation
// These methods implement the GraphicsSystemInterface for VM integration

// LoadPic loads a picture from a file
func (gs *GraphicsSystem) LoadPic(filename string) (int, error) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	return gs.pictures.LoadPic(filename)
}

// CreatePic creates a new empty picture
func (gs *GraphicsSystem) CreatePic(width, height int) (int, error) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	return gs.pictures.CreatePic(width, height)
}

// CreatePicFrom creates a new picture from an existing picture
func (gs *GraphicsSystem) CreatePicFrom(srcID int) (int, error) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	return gs.pictures.CreatePicFrom(srcID)
}

// CreatePicWithSize は指定されたサイズの空のピクチャーを生成する
// srcID: 参照用のソースピクチャーID（存在確認のみ）
// width, height: 新しいピクチャーのサイズ
// 戻り値: 新しいピクチャーID、エラー
func (gs *GraphicsSystem) CreatePicWithSize(srcID, width, height int) (int, error) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	return gs.pictures.CreatePicWithSize(srcID, width, height)
}

// DelPic deletes a picture
func (gs *GraphicsSystem) DelPic(id int) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	return gs.pictures.DelPic(id)
}

// PicWidth returns the width of a picture
func (gs *GraphicsSystem) PicWidth(id int) int {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	return gs.pictures.PicWidth(id)
}

// PicHeight returns the height of a picture
func (gs *GraphicsSystem) PicHeight(id int) int {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	return gs.pictures.PicHeight(id)
}

// GetVirtualWidth returns the virtual desktop width
func (gs *GraphicsSystem) GetVirtualWidth() int {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	return gs.virtualWidth
}

// GetVirtualHeight returns the virtual desktop height
func (gs *GraphicsSystem) GetVirtualHeight() int {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	return gs.virtualHeight
}

// OpenWin opens a window
func (gs *GraphicsSystem) OpenWin(picID int, opts ...any) (int, error) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	// Convert any options to WinOption
	winOpts := gs.parseWinOptions(opts)

	return gs.windows.OpenWin(picID, winOpts...)
}

// parseWinOptions converts any slice to WinOption slice
// Supports: x, y, width, height, picX, picY, bgColor
func (gs *GraphicsSystem) parseWinOptions(opts []any) []WinOption {
	winOpts := make([]WinOption, 0)

	// OpenWin(pic, x, y, width, height, pic_x, pic_y, color)
	if len(opts) >= 2 {
		if x, ok := toIntFromAny(opts[0]); ok {
			if y, ok := toIntFromAny(opts[1]); ok {
				winOpts = append(winOpts, WithPosition(x, y))
			}
		}
	}
	if len(opts) >= 4 {
		if w, ok := toIntFromAny(opts[2]); ok {
			if h, ok := toIntFromAny(opts[3]); ok {
				winOpts = append(winOpts, WithSize(w, h))
			}
		}
	}
	if len(opts) >= 6 {
		if picX, ok := toIntFromAny(opts[4]); ok {
			if picY, ok := toIntFromAny(opts[5]); ok {
				winOpts = append(winOpts, WithPicOffset(picX, picY))
			}
		}
	}
	if len(opts) >= 7 {
		if colorInt, ok := toIntFromAny(opts[6]); ok {
			winOpts = append(winOpts, WithBgColor(ColorFromInt(colorInt)))
		}
	}

	return winOpts
}

// MoveWin moves or modifies a window
func (gs *GraphicsSystem) MoveWin(id int, opts ...any) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	winOpts := make([]WinOption, 0)

	// MoveWin(win, pic) - change picture only
	// MoveWin(win, pic, x, y, width, height, pic_x, pic_y) - full update
	if len(opts) >= 1 {
		if picID, ok := toIntFromAny(opts[0]); ok {
			winOpts = append(winOpts, WithPicID(picID))
		}
	}
	if len(opts) >= 3 {
		if x, ok := toIntFromAny(opts[1]); ok {
			if y, ok := toIntFromAny(opts[2]); ok {
				winOpts = append(winOpts, WithPosition(x, y))
			}
		}
	}
	if len(opts) >= 5 {
		if w, ok := toIntFromAny(opts[3]); ok {
			if h, ok := toIntFromAny(opts[4]); ok {
				winOpts = append(winOpts, WithSize(w, h))
			}
		}
	}
	if len(opts) >= 7 {
		if picX, ok := toIntFromAny(opts[5]); ok {
			if picY, ok := toIntFromAny(opts[6]); ok {
				winOpts = append(winOpts, WithPicOffset(picX, picY))
			}
		}
	}

	return gs.windows.MoveWin(id, winOpts...)
}

// CloseWin closes a window
func (gs *GraphicsSystem) CloseWin(id int) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	// Delete casts belonging to this window (要件 9.2)
	gs.casts.DeleteCastsByWindow(id)

	return gs.windows.CloseWin(id)
}

// CloseWinAll closes all windows
func (gs *GraphicsSystem) CloseWinAll() {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	// Get all windows and delete their casts
	windows := gs.windows.GetWindowsOrdered()
	for _, win := range windows {
		gs.casts.DeleteCastsByWindow(win.ID)
	}

	gs.windows.CloseWinAll()
}

// CapTitle sets the caption of a window
func (gs *GraphicsSystem) CapTitle(id int, title string) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	return gs.windows.CapTitle(id, title)
}

// CapTitleAll は全てのウィンドウのキャプションを設定する
// title: 設定するキャプション
// 受け入れ基準 3.1, 3.2
func (gs *GraphicsSystem) CapTitleAll(title string) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.windows.CapTitleAll(title)
}

// GetPicNo returns the picture number associated with a window
func (gs *GraphicsSystem) GetPicNo(id int) (int, error) {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	return gs.windows.GetPicNo(id)
}

// GetWinByPicID returns the window ID associated with a picture ID
func (gs *GraphicsSystem) GetWinByPicID(picID int) (int, error) {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	return gs.windows.GetWinByPicID(picID)
}

// Cast management

// PutCast places a cast on a window
func (gs *GraphicsSystem) PutCast(winID, picID, x, y, srcX, srcY, w, h int) (int, error) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	return gs.casts.PutCast(winID, picID, x, y, srcX, srcY, w, h)
}

// PutCastWithTransColor places a cast on a window with transparent color
func (gs *GraphicsSystem) PutCastWithTransColor(winID, picID, x, y, srcX, srcY, w, h int, transColor color.Color) (int, error) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	return gs.casts.PutCastWithTransColor(winID, picID, x, y, srcX, srcY, w, h, transColor)
}

// MoveCast moves a cast
func (gs *GraphicsSystem) MoveCast(id int, opts ...any) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	castOpts := make([]CastOption, 0)

	// MoveCast(cast_no, x, y) - move position only
	// MoveCast(cast_no, x, y, src_x, src_y, width, height) - move and change source
	// MoveCast(cast_no, pic_no, x, y) - change picture and position
	if len(opts) >= 2 {
		if x, ok := toIntFromAny(opts[0]); ok {
			if y, ok := toIntFromAny(opts[1]); ok {
				castOpts = append(castOpts, WithCastPosition(x, y))
			}
		}
	}
	if len(opts) >= 6 {
		if srcX, ok := toIntFromAny(opts[2]); ok {
			if srcY, ok := toIntFromAny(opts[3]); ok {
				if w, ok := toIntFromAny(opts[4]); ok {
					if h, ok := toIntFromAny(opts[5]); ok {
						castOpts = append(castOpts, WithCastSource(srcX, srcY, w, h))
					}
				}
			}
		}
	}
	// Check for pic_no, x, y pattern (3 args where first is pic)
	if len(opts) == 3 {
		if picID, ok := toIntFromAny(opts[0]); ok {
			if x, ok := toIntFromAny(opts[1]); ok {
				if y, ok := toIntFromAny(opts[2]); ok {
					castOpts = []CastOption{
						WithCastPicID(picID),
						WithCastPosition(x, y),
					}
				}
			}
		}
	}

	return gs.casts.MoveCast(id, castOpts...)
}

// MoveCastWithOptions moves a cast with explicit options
// キャストはスプライトとして動作し、位置/ソースの更新のみを行う
// 実際の描画は毎フレームdrawCastsForWindowで行われる
func (gs *GraphicsSystem) MoveCastWithOptions(id int, opts ...CastOption) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	// キャストの位置/ソースを更新
	if err := gs.casts.MoveCast(id, opts...); err != nil {
		return err
	}

	return nil
}

// DelCast deletes a cast
func (gs *GraphicsSystem) DelCast(id int) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	return gs.casts.DelCast(id)
}

// Text rendering

// TextWrite writes text to a picture
func (gs *GraphicsSystem) TextWrite(picID, x, y int, text string) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	pic, err := gs.pictures.GetPicWithoutLock(picID)
	if err != nil {
		gs.log.Error("TextWrite: picture not found", "picID", picID, "error", err)
		return err
	}

	return gs.textRenderer.TextWrite(pic, x, y, text)
}

// SetFont sets the font
func (gs *GraphicsSystem) SetFont(name string, size int, opts ...any) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	fontOpts := make([]FontOption, 0)

	// SetFont(size, name, charset, italic, underline, strikeout, weight)
	// Note: The order in FILLY is different from our internal API
	if len(opts) >= 1 {
		if charset, ok := toIntFromAny(opts[0]); ok {
			fontOpts = append(fontOpts, WithCharset(charset))
		}
	}
	if len(opts) >= 2 {
		if italic, ok := toIntFromAny(opts[1]); ok {
			fontOpts = append(fontOpts, WithItalic(italic != 0))
		}
	}
	if len(opts) >= 3 {
		if underline, ok := toIntFromAny(opts[2]); ok {
			fontOpts = append(fontOpts, WithUnderline(underline != 0))
		}
	}
	if len(opts) >= 4 {
		if strikeout, ok := toIntFromAny(opts[3]); ok {
			fontOpts = append(fontOpts, WithStrikeout(strikeout != 0))
		}
	}
	if len(opts) >= 5 {
		if weight, ok := toIntFromAny(opts[4]); ok {
			fontOpts = append(fontOpts, WithWeight(weight))
		}
	}

	return gs.textRenderer.SetFont(name, size, fontOpts...)
}

// SetTextColor sets the text color
func (gs *GraphicsSystem) SetTextColor(c any) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	var textColor color.Color
	switch v := c.(type) {
	case int:
		textColor = ColorFromInt(v)
	case color.Color:
		textColor = v
	default:
		return fmt.Errorf("invalid color type: %T", c)
	}
	gs.textRenderer.SetTextColor(textColor)
	return nil
}

// SetBgColor sets the background color
func (gs *GraphicsSystem) SetBgColor(c any) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	var bgColor color.Color
	switch v := c.(type) {
	case int:
		bgColor = ColorFromInt(v)
	case color.Color:
		bgColor = v
	default:
		return fmt.Errorf("invalid color type: %T", c)
	}
	gs.textRenderer.SetBgColor(bgColor)
	return nil
}

// SetBackMode sets the background mode
func (gs *GraphicsSystem) SetBackMode(mode int) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.textRenderer.SetBackMode(mode)
	return nil
}

// Drawing primitives

// DrawLine draws a line
func (gs *GraphicsSystem) DrawLine(picID, x1, y1, x2, y2 int) error {
	return gs.DrawLineOnPic(picID, x1, y1, x2, y2)
}

// DrawRect draws a rectangle
func (gs *GraphicsSystem) DrawRect(picID, x1, y1, x2, y2, fillMode int) error {
	return gs.DrawRectOnPic(picID, x1, y1, x2, y2, fillMode)
}

// FillRect fills a rectangle
func (gs *GraphicsSystem) FillRect(picID, x1, y1, x2, y2 int, c any) error {
	var fillColor color.Color
	switch v := c.(type) {
	case int:
		fillColor = ColorFromInt(v)
	case color.Color:
		fillColor = v
	default:
		fillColor = gs.paintColor
	}
	return gs.FillRectOnPic(picID, x1, y1, x2, y2, fillColor)
}

// DrawCircle draws a circle
func (gs *GraphicsSystem) DrawCircle(picID, x, y, radius, fillMode int) error {
	return gs.DrawCircleOnPic(picID, x, y, radius, fillMode)
}

// SetLineSize sets the line size
func (gs *GraphicsSystem) SetLineSize(size int) {
	gs.SetLineSizeValue(size)
}

// SetPaintColor sets the paint color
func (gs *GraphicsSystem) SetPaintColor(c any) error {
	var paintColor color.Color
	switch v := c.(type) {
	case int:
		paintColor = ColorFromInt(v)
	case color.Color:
		paintColor = v
	default:
		return fmt.Errorf("invalid color type: %T", c)
	}
	gs.SetPaintColorValue(paintColor)
	return nil
}

// GetColor gets the color at a specific pixel
func (gs *GraphicsSystem) GetColor(picID, x, y int) (int, error) {
	return gs.GetColorAt(picID, x, y)
}

// Picture transfer methods

// MovePicTransfer transfers a picture region (wrapper for internal MovePic)
func (gs *GraphicsSystem) MovePicTransfer(srcID, srcX, srcY, width, height, dstID, dstX, dstY, mode int) error {
	return gs.MovePic(srcID, srcX, srcY, width, height, dstID, dstX, dstY, mode)
}

// MovePicWithSpeedTransfer transfers a picture region with speed (wrapper)
func (gs *GraphicsSystem) MovePicWithSpeedTransfer(srcID, srcX, srcY, width, height, dstID, dstX, dstY, mode, speed int) error {
	return gs.MovePicWithSpeed(srcID, srcX, srcY, width, height, dstID, dstX, dstY, mode, speed)
}

// MoveSPicTransfer scales and transfers a picture region (wrapper)
func (gs *GraphicsSystem) MoveSPicTransfer(srcID, srcX, srcY, srcW, srcH, dstID, dstX, dstY, dstW, dstH int) error {
	return gs.MoveSPic(srcID, srcX, srcY, srcW, srcH, dstID, dstX, dstY, dstW, dstH)
}

// TransPic transfers with transparency (interface method)
// Accepts any type for transColor and converts to color.Color
func (gs *GraphicsSystem) TransPic(srcID, srcX, srcY, width, height, dstID, dstX, dstY int, transColor any) error {
	var tc color.Color
	switch v := transColor.(type) {
	case int:
		tc = ColorFromInt(v)
	case color.Color:
		tc = v
	default:
		tc = DefaultTransparentColor
	}
	return gs.TransPicInternal(srcID, srcX, srcY, width, height, dstID, dstX, dstY, tc)
}

// ReversePicTransfer transfers with horizontal flip (wrapper)
func (gs *GraphicsSystem) ReversePicTransfer(srcID, srcX, srcY, width, height, dstID, dstX, dstY int) error {
	return gs.ReversePic(srcID, srcX, srcY, width, height, dstID, dstX, dstY)
}

// toIntFromAny converts any to int
func toIntFromAny(v any) (int, bool) {
	switch val := v.(type) {
	case int:
		return val, true
	case int64:
		return int(val), true
	case float64:
		return int(val), true
	case float32:
		return int(val), true
	default:
		return 0, false
	}
}

// drawWindowDecoration はWindows 3.1風のウィンドウ装飾を描画する
// _old_implementation2/pkg/engine/renderer.goを参考に実装
// 要件 3.11: ウィンドウをZ順序で管理し、後から開いたウィンドウを前面に表示する
// 要件 3.12: ウィンドウの背景色（color引数）を適用する
func (gs *GraphicsSystem) drawWindowDecoration(screen *ebiten.Image, win *Window, pic *Picture) {
	const (
		borderThickness = 4  // 外枠の幅
		titleBarHeight  = 20 // タイトルバーの高さ
	)

	// Windows 3.1風の色（_old_implementation2を参考）
	var (
		titleBarColor  = color.RGBA{0, 0, 128, 255}     // 濃い青
		borderColor    = color.RGBA{192, 192, 192, 255} // グレー
		highlightColor = color.RGBA{255, 255, 255, 255} // 白（立体効果のハイライト）
		shadowColor    = color.RGBA{0, 0, 0, 255}       // 黒（立体効果の影）
	)

	// ウィンドウの実際のサイズを計算
	winWidth := pic.Width
	winHeight := pic.Height
	if win.Width > 0 {
		winWidth = win.Width
	}
	if win.Height > 0 {
		winHeight = win.Height
	}

	// 全体のウィンドウサイズ（装飾を含む）
	winX := float32(win.X)
	winY := float32(win.Y)
	winW := float32(winWidth)
	winH := float32(winHeight)
	totalW := winW + float32(borderThickness*2)
	totalH := winH + float32(borderThickness*2) + float32(titleBarHeight)

	// 1. ウィンドウフレームの背景を描画（グレー）
	vector.DrawFilledRect(screen,
		winX, winY,
		totalW, totalH,
		borderColor, false)

	// 2. 3D枠線効果を描画
	// 上と左の縁（ハイライト - 立体的に浮き上がって見える）
	vector.StrokeLine(screen,
		winX, winY,
		winX+totalW, winY,
		1, highlightColor, false)
	vector.StrokeLine(screen,
		winX, winY,
		winX, winY+totalH,
		1, highlightColor, false)

	// 下と右の縁（影 - 立体的にへこんで見える）
	vector.StrokeLine(screen,
		winX, winY+totalH,
		winX+totalW, winY+totalH,
		1, shadowColor, false)
	vector.StrokeLine(screen,
		winX+totalW, winY,
		winX+totalW, winY+totalH,
		1, shadowColor, false)

	// 3. タイトルバーを描画（濃い青）
	vector.DrawFilledRect(screen,
		winX+float32(borderThickness),
		winY+float32(borderThickness),
		winW, float32(titleBarHeight),
		titleBarColor, false)

	// 4. キャプションテキストを描画（後のフェーズで実装）
	// TODO: win.Captionがある場合、白色でテキストを描画

	// 5. コンテンツ領域を描画
	contentX := win.X + borderThickness
	contentY := win.Y + borderThickness + titleBarHeight

	// 5.1 背景色を描画（要件 3.12）
	if win.BgColor != nil {
		vector.DrawFilledRect(screen,
			float32(contentX), float32(contentY),
			float32(winWidth), float32(winHeight),
			win.BgColor, false)
	}

	// 5.2 ピクチャーを描画（コンテンツ領域内）
	// ウィンドウ矩形（スクリーン座標でのコンテンツ領域）
	winRect := image.Rect(contentX, contentY, contentX+winWidth, contentY+winHeight)

	// PicXとPicYは「ピクチャー内の参照位置」を指定する
	// 正の値: ピクチャーの(PicX, PicY)がウィンドウの左上に表示される
	// 負の値: ピクチャーの左上がウィンドウの(-PicX, -PicY)に表示される
	//
	// 例: PicX=-490, PicY=-235 の場合
	// ピクチャーの左上はウィンドウの(490, 235)に配置される
	// つまり、ピクチャーはウィンドウの中央付近に表示される

	// 画像の描画位置を計算
	// PicXが負の場合、画像はウィンドウ内で右にシフト
	// PicYが負の場合、画像はウィンドウ内で下にシフト
	imgAbsX := contentX - win.PicX // PicXが負なら右にシフト
	imgAbsY := contentY - win.PicY // PicYが負なら下にシフト
	imgRect := image.Rect(imgAbsX, imgAbsY, imgAbsX+pic.Width, imgAbsY+pic.Height)

	// 交差領域を計算（画像の可視部分）
	drawRect := winRect.Intersect(imgRect)

	// 交差領域が空なら描画しない
	if drawRect.Empty() {
		return
	}

	// ソース矩形を計算（ピクチャー内の座標）
	// 画像の左上は(imgAbsX, imgAbsY)にある
	// 可視部分は(drawRect.Min.X, drawRect.Min.Y)から始まる
	// ソース座標は画像原点からの相対座標
	srcX := drawRect.Min.X - imgAbsX
	srcY := drawRect.Min.Y - imgAbsY
	srcW := drawRect.Dx()
	srcH := drawRect.Dy()

	// ソース領域を切り出す
	srcRect := image.Rect(srcX, srcY, srcX+srcW, srcY+srcH)
	subImg := pic.Image.SubImage(srcRect).(*ebiten.Image)

	// 交差点でスクリーンに描画
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(drawRect.Min.X), float64(drawRect.Min.Y))
	screen.DrawImage(subImg, opts)
}

// VirtualToScreen は仮想デスクトップ座標をスクリーン座標に変換する
// 要件 8.4: 描画領域を実際のウィンドウサイズに合わせてスケーリングする
// 要件 8.5: アスペクト比を維持してスケーリングする
func (gs *GraphicsSystem) VirtualToScreen(vx, vy int, screenW, screenH int) (int, int) {
	scaleX := float64(screenW) / float64(gs.virtualWidth)
	scaleY := float64(screenH) / float64(gs.virtualHeight)
	scale := min(scaleX, scaleY)

	offsetX := (float64(screenW) - float64(gs.virtualWidth)*scale) / 2
	offsetY := (float64(screenH) - float64(gs.virtualHeight)*scale) / 2

	return int(float64(vx)*scale + offsetX), int(float64(vy)*scale + offsetY)
}

// ScreenToVirtual はスクリーン座標を仮想デスクトップ座標に変換する
// 要件 8.7: マウスイベントが発生したとき、描画領域座標に変換してMesP2、MesP3に設定する
func (gs *GraphicsSystem) ScreenToVirtual(sx, sy int, screenW, screenH int) (int, int) {
	scaleX := float64(screenW) / float64(gs.virtualWidth)
	scaleY := float64(screenH) / float64(gs.virtualHeight)
	scale := min(scaleX, scaleY)

	offsetX := (float64(screenW) - float64(gs.virtualWidth)*scale) / 2
	offsetY := (float64(screenH) - float64(gs.virtualHeight)*scale) / 2

	vx := int((float64(sx) - offsetX) / scale)
	vy := int((float64(sy) - offsetY) / scale)

	// 範囲チェック
	if vx < 0 {
		vx = 0
	}
	if vx >= gs.virtualWidth {
		vx = gs.virtualWidth - 1
	}
	if vy < 0 {
		vy = 0
	}
	if vy >= gs.virtualHeight {
		vy = gs.virtualHeight - 1
	}

	return vx, vy
}

// GetScaleAndOffset はスケーリング係数とオフセットを計算する
// 要件 8.4, 8.5, 8.6: スケーリングとレターボックス
func (gs *GraphicsSystem) GetScaleAndOffset(screenW, screenH int) (scale, offsetX, offsetY float64) {
	scaleX := float64(screenW) / float64(gs.virtualWidth)
	scaleY := float64(screenH) / float64(gs.virtualHeight)
	scale = min(scaleX, scaleY)

	offsetX = (float64(screenW) - float64(gs.virtualWidth)*scale) / 2
	offsetY = (float64(screenH) - float64(gs.virtualHeight)*scale) / 2

	return scale, offsetX, offsetY
}

// DrawScaled は仮想デスクトップをスケーリングして描画する
// 要件 8.4: 描画領域を実際のウィンドウサイズに合わせてスケーリングする
// 要件 8.5: アスペクト比を維持してスケーリングする
// 要件 8.6: スケーリング時にレターボックス（黒帯）を表示する
func (gs *GraphicsSystem) DrawScaled(screen *ebiten.Image, virtualScreen *ebiten.Image) {
	screenW := screen.Bounds().Dx()
	screenH := screen.Bounds().Dy()

	scale, offsetX, offsetY := gs.GetScaleAndOffset(screenW, screenH)

	// レターボックス（黒帯）を描画（要件 8.6）
	// 画面全体を黒で塗りつぶす（レターボックス部分）
	screen.Fill(color.Black)

	// 仮想デスクトップをスケーリングして描画
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Scale(scale, scale)
	opts.GeoM.Translate(offsetX, offsetY)
	opts.Filter = ebiten.FilterLinear // 線形補間でスムーズにスケーリング

	screen.DrawImage(virtualScreen, opts)
}
