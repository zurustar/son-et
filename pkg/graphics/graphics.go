package graphics

import (
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

// NewGraphicsSystem は新しい GraphicsSystem を作成する
func NewGraphicsSystem(basePath string, opts ...Option) *GraphicsSystem {
	gs := &GraphicsSystem{
		virtualWidth:  1280, // skelton要件に合わせて1280x720
		virtualHeight: 720,
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
func (gs *GraphicsSystem) Draw(screen *ebiten.Image) {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	// 背景色は window.go の drawDesktop() で既に設定されているので、
	// ここでは塗りつぶさない

	// ウィンドウをZ順序で取得
	windows := gs.windows.GetWindowsOrdered()

	// 各ウィンドウを描画
	for _, win := range windows {
		if !win.Visible {
			continue
		}

		// ピクチャーを取得
		pic, err := gs.pictures.GetPic(win.PicID)
		if err != nil {
			gs.log.Warn("Failed to get picture for window",
				"windowID", win.ID,
				"pictureID", win.PicID,
				"error", err)
			continue
		}

		// ウィンドウ装飾を描画（Windows 3.1風）
		gs.drawWindowDecoration(screen, win, pic)

		// このウィンドウに属するキャストを描画
		casts := gs.casts.GetCastsByWindow(win.ID)
		for _, cast := range casts {
			if !cast.Visible {
				continue
			}

			// キャストのピクチャーを取得
			castPic, err := gs.pictures.GetPic(cast.PicID)
			if err != nil {
				gs.log.Warn("Failed to get picture for cast",
					"castID", cast.ID,
					"pictureID", cast.PicID,
					"error", err)
				continue
			}

			// キャストの描画領域を計算
			castOpts := &ebiten.DrawImageOptions{}
			castOpts.GeoM.Translate(float64(win.X+cast.X), float64(win.Y+cast.Y))

			// キャストを描画（透明色除外は後のフェーズで実装）
			screen.DrawImage(castPic.Image, castOpts)
		}
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

// OpenWin opens a window
func (gs *GraphicsSystem) OpenWin(picID int, opts ...any) (int, error) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	// Convert any options to WinOption
	winOpts := make([]WinOption, 0)
	// For now, just open with default options
	// TODO: Parse opts and convert to WinOption when needed

	return gs.windows.OpenWin(picID, winOpts...)
}

// MoveWin moves or modifies a window
func (gs *GraphicsSystem) MoveWin(id int, opts ...any) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	// Convert any options to WinOption
	winOpts := make([]WinOption, 0)
	// TODO: Parse opts and convert to WinOption when needed

	return gs.windows.MoveWin(id, winOpts...)
}

// CloseWin closes a window
func (gs *GraphicsSystem) CloseWin(id int) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	return gs.windows.CloseWin(id)
}

// CloseWinAll closes all windows
func (gs *GraphicsSystem) CloseWinAll() {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.windows.CloseWinAll()
}

// CapTitle sets the caption of a window
func (gs *GraphicsSystem) CapTitle(id int, title string) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	return gs.windows.CapTitle(id, title)
}

// GetPicNo returns the picture number associated with a window
func (gs *GraphicsSystem) GetPicNo(id int) (int, error) {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	return gs.windows.GetPicNo(id)
}

// Cast management (placeholder implementations)

// PutCast places a cast on a window
func (gs *GraphicsSystem) PutCast(winID, picID, x, y, srcX, srcY, w, h int) (int, error) {
	gs.log.Debug("PutCast not yet implemented")
	return -1, nil
}

// MoveCast moves a cast
func (gs *GraphicsSystem) MoveCast(id int, opts ...any) error {
	gs.log.Debug("MoveCast not yet implemented")
	return nil
}

// DelCast deletes a cast
func (gs *GraphicsSystem) DelCast(id int) error {
	gs.log.Debug("DelCast not yet implemented")
	return nil
}

// Text rendering (placeholder implementations)

// TextWrite writes text to a picture
func (gs *GraphicsSystem) TextWrite(picID, x, y int, text string) error {
	gs.log.Debug("TextWrite not yet implemented")
	return nil
}

// SetFont sets the font
func (gs *GraphicsSystem) SetFont(name string, size int, opts ...any) error {
	gs.log.Debug("SetFont not yet implemented")
	return nil
}

// SetTextColor sets the text color
func (gs *GraphicsSystem) SetTextColor(c any) error {
	gs.log.Debug("SetTextColor not yet implemented")
	return nil
}

// SetBgColor sets the background color
func (gs *GraphicsSystem) SetBgColor(c any) error {
	gs.log.Debug("SetBgColor not yet implemented")
	return nil
}

// SetBackMode sets the background mode
func (gs *GraphicsSystem) SetBackMode(mode int) error {
	gs.log.Debug("SetBackMode not yet implemented")
	return nil
}

// Drawing primitives (placeholder implementations)

// DrawLine draws a line
func (gs *GraphicsSystem) DrawLine(picID, x1, y1, x2, y2 int) error {
	gs.log.Debug("DrawLine not yet implemented")
	return nil
}

// DrawRect draws a rectangle
func (gs *GraphicsSystem) DrawRect(picID, x1, y1, x2, y2, fillMode int) error {
	gs.log.Debug("DrawRect not yet implemented")
	return nil
}

// FillRect fills a rectangle
func (gs *GraphicsSystem) FillRect(picID, x1, y1, x2, y2 int, c any) error {
	gs.log.Debug("FillRect not yet implemented")
	return nil
}

// SetPaintColor sets the paint color
func (gs *GraphicsSystem) SetPaintColor(c any) error {
	gs.log.Debug("SetPaintColor not yet implemented")
	return nil
}

// drawWindowDecoration はWindows 3.1風のウィンドウ装飾を描画する
// _old_implementation2/pkg/engine/renderer.goを参考に実装
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

	// ピクチャーを描画（コンテンツ領域内）
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(contentX), float64(contentY))

	// PicX, PicYオフセットを適用
	if win.PicX != 0 || win.PicY != 0 {
		// サブイメージを作成してオフセットを適用
		srcX := win.PicX
		srcY := win.PicY
		srcW := winWidth
		srcH := winHeight

		// 範囲チェック
		if srcX < 0 {
			srcX = 0
		}
		if srcY < 0 {
			srcY = 0
		}
		if srcX+srcW > pic.Width {
			srcW = pic.Width - srcX
		}
		if srcY+srcH > pic.Height {
			srcH = pic.Height - srcY
		}

		if srcW > 0 && srcH > 0 {
			subImg := pic.Image.SubImage(image.Rect(srcX, srcY, srcX+srcW, srcY+srcH)).(*ebiten.Image)
			screen.DrawImage(subImg, opts)
		}
	} else {
		screen.DrawImage(pic.Image, opts)
	}
}

// drawText はテキストを描画する（内部ヘルパー関数）
func (gs *GraphicsSystem) drawText(screen *ebiten.Image, text string, x, y int, c color.Color) {
	// 簡易的なテキスト描画
	// TODO: 後でTextRendererを使用するように改善

	// 現時点では何もしない（テキスト描画は後のフェーズで実装）
	// ウィンドウの装飾だけでも十分視覚的に改善される
}
