// graphics_core.go はGraphicsSystemのコア機能を提供する
// 構造体定義、コンストラクタ、更新ループ、シャットダウン、
// オプション関数、ゲッターメソッド、ピクチャー操作を含む
package graphics

import (
	"fmt"
	"image/color"
	"io/fs"
	"log/slog"
	"sync"

	"github.com/zurustar/son-et/pkg/fileutil"
)

// GraphicsSystem は描画システム全体を管理する
// スプライトシステム移行: LayerManagerは不要になった
type GraphicsSystem struct {
	pictures             *PictureManager
	windows              *WindowManager
	casts                *CastManager
	textRenderer         *TextRenderer
	sceneChanges         *SceneChangeManager
	debugOverlay         *DebugOverlay
	spriteManager        *SpriteManager        // スプライトシステム要件 3.1〜3.6: SpriteManagerを統合
	windowSpriteManager  *WindowSpriteManager  // スプライトシステム要件 7.1〜7.3: WindowSpriteManagerを統合
	pictureSpriteManager *PictureSpriteManager // スプライトシステム要件 6.1〜6.3: PictureSpriteManagerを統合
	castSpriteManager    *CastSpriteManager    // スプライトシステム要件 8.1〜8.4: CastSpriteManagerを統合
	textSpriteManager    *TextSpriteManager    // スプライトシステム要件 5.1〜5.5: TextSpriteManagerを統合
	shapeSpriteManager   *ShapeSpriteManager   // スプライトシステム要件 9.1〜9.3: ShapeSpriteManagerを統合

	// パフォーマンス測定（タスク 7.1, 7.2, 7.3）
	fpsCounter     *FPSCounter          // FPS測定
	statsCollector *SpriteStatsCollector // スプライト統計収集

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
		gs.pictures.SetFileSystem(fileutil.NewRealFS(basePath))
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

// WithPerformanceMonitoring はパフォーマンス測定の有効/無効を設定する
// タスク 7.3: パフォーマンスボトルネックの特定と改善
func WithPerformanceMonitoring(enabled bool) Option {
	return func(gs *GraphicsSystem) {
		if gs.fpsCounter != nil {
			gs.fpsCounter.SetEnabled(enabled)
		}
		if gs.statsCollector != nil {
			gs.statsCollector.SetEnabled(enabled)
		}
	}
}

// NewGraphicsSystem は新しい GraphicsSystem を作成する
func NewGraphicsSystem(basePath string, opts ...Option) *GraphicsSystem {
	gs := &GraphicsSystem{
		virtualWidth:  1024, // skelton要件に合わせて1024x768
		virtualHeight: 768,
		paintColor:    color.RGBA{0, 0, 0, 255}, // デフォルトは黒（オリジナルFILLY互換）
		lineSize:      1,
		log:           slog.Default(),
	}

	// サブシステムを初期化
	gs.pictures = NewPictureManager(basePath)
	gs.windows = NewWindowManager()
	gs.casts = NewCastManager()
	gs.textRenderer = NewTextRenderer()
	gs.sceneChanges = NewSceneChangeManager()
	gs.debugOverlay = NewDebugOverlay()
	gs.spriteManager = NewSpriteManager()                               // スプライトシステム要件 3.1〜3.6: SpriteManagerを初期化
	gs.windowSpriteManager = NewWindowSpriteManager(gs.spriteManager)   // スプライトシステム要件 7.1〜7.3: WindowSpriteManagerを初期化
	gs.pictureSpriteManager = NewPictureSpriteManager(gs.spriteManager) // スプライトシステム要件 6.1〜6.3: PictureSpriteManagerを初期化
	gs.castSpriteManager = NewCastSpriteManager(gs.spriteManager)       // スプライトシステム要件 8.1〜8.4: CastSpriteManagerを初期化
	gs.textSpriteManager = NewTextSpriteManager(gs.spriteManager)       // スプライトシステム要件 5.1〜5.5: TextSpriteManagerを初期化
	gs.shapeSpriteManager = NewShapeSpriteManager(gs.spriteManager)     // スプライトシステム要件 9.1〜9.3: ShapeSpriteManagerを初期化

	// パフォーマンス測定（タスク 7.1, 7.2, 7.3）
	gs.fpsCounter = NewFPSCounter()
	gs.statsCollector = NewSpriteStatsCollector(gs)

	// スプライトシステム移行: LayerManagerは不要になった
	// CastManagerとTextRendererへのLayerManager設定は不要

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

// SetEmbedFS は埋め込みファイルシステムを設定する
// 埋め込みタイトルを実行する場合に使用
func (gs *GraphicsSystem) SetEmbedFS(fsys fs.FS) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.pictures.SetEmbedFS(fsys)
}

// dumpSpriteState はスプライト構成をログに出力する（デバッグ用）
// 操作後のスプライト階層を確認するために使用
func (gs *GraphicsSystem) dumpSpriteState(operation string) {
	if gs.spriteManager == nil {
		return
	}

	// DEBUGレベルでのみ出力
	// JSON形式で改行を保持するため、直接fmt.Printfで出力
	dump := gs.spriteManager.DumpSpriteState()
	gs.log.Debug("Sprite state after " + operation)
	fmt.Printf("=== Sprite State (%s) ===\n%s\n", operation, dump)
}

// Update はゲームループから呼び出され、コマンドキューを処理する
// Ebitengineのメインスレッドで実行される
func (gs *GraphicsSystem) Update() error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	// シーンチェンジを更新（要件 13.11: 非同期実行）
	gs.sceneChanges.Update()

	return nil
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

	gs.log.Info("GraphicsSystem shutdown complete")
}

// ============================================================================
// ゲッターメソッド
// ============================================================================

// GetSpriteManager はSpriteManagerを返す
// スプライトシステム要件 3.1〜3.6: GraphicsSystemにSpriteManagerを統合する
func (gs *GraphicsSystem) GetSpriteManager() *SpriteManager {
	return gs.spriteManager
}

// GetWindowSpriteManager はWindowSpriteManagerを返す
// スプライトシステム要件 7.1〜7.3: GraphicsSystemにWindowSpriteManagerを統合する
func (gs *GraphicsSystem) GetWindowSpriteManager() *WindowSpriteManager {
	return gs.windowSpriteManager
}

// GetPictureSpriteManager はPictureSpriteManagerを返す
// スプライトシステム要件 6.1〜6.3: GraphicsSystemにPictureSpriteManagerを統合する
func (gs *GraphicsSystem) GetPictureSpriteManager() *PictureSpriteManager {
	return gs.pictureSpriteManager
}

// GetCastSpriteManager はCastSpriteManagerを返す
// スプライトシステム要件 8.1〜8.4: GraphicsSystemにCastSpriteManagerを統合する
func (gs *GraphicsSystem) GetCastSpriteManager() *CastSpriteManager {
	return gs.castSpriteManager
}

// GetTextSpriteManager はTextSpriteManagerを返す
// スプライトシステム要件 5.1〜5.5: GraphicsSystemにTextSpriteManagerを統合する
func (gs *GraphicsSystem) GetTextSpriteManager() *TextSpriteManager {
	return gs.textSpriteManager
}

// GetShapeSpriteManager はShapeSpriteManagerを返す
// スプライトシステム要件 9.1〜9.3: GraphicsSystemにShapeSpriteManagerを統合する
func (gs *GraphicsSystem) GetShapeSpriteManager() *ShapeSpriteManager {
	return gs.shapeSpriteManager
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

// GetWindowCount returns the number of open windows
func (gs *GraphicsSystem) GetWindowCount() int {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	if gs.windows == nil {
		return 0
	}
	return gs.windows.Count()
}

// ============================================================================
// ピクチャー操作メソッド
// ============================================================================

// LoadPic loads a picture from a file
func (gs *GraphicsSystem) LoadPic(filename string) (int, error) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	// ピクチャーをロード
	picID, err := gs.pictures.LoadPic(filename)
	if err != nil {
		return picID, err
	}

	// 要件 11.1: LoadPicが呼び出されたとき、非表示のPictureSpriteを作成する
	// これにより、ウインドウに関連付けられる前でもキャストやテキストの親として機能できる
	if gs.pictureSpriteManager != nil {
		pic, picErr := gs.pictures.GetPicWithoutLock(picID)
		if picErr == nil && pic != nil && pic.Image != nil {
			gs.pictureSpriteManager.CreatePictureSpriteOnLoad(pic.Image, picID, pic.Width, pic.Height)
			gs.log.Debug("LoadPic: created PictureSprite on load", "picID", picID, "filename", filename)
		}
	}

	return picID, nil
}

// CreatePic creates a new empty picture
func (gs *GraphicsSystem) CreatePic(width, height int) (int, error) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	picID, err := gs.pictures.CreatePic(width, height)
	if err != nil {
		return picID, err
	}

	// 要件 11.1: CreatePicが呼び出されたとき、非表示のPictureSpriteを作成する
	if gs.pictureSpriteManager != nil {
		pic, picErr := gs.pictures.GetPicWithoutLock(picID)
		if picErr == nil && pic != nil && pic.Image != nil {
			gs.pictureSpriteManager.CreatePictureSpriteOnLoad(pic.Image, picID, pic.Width, pic.Height)
			gs.log.Debug("CreatePic: created PictureSprite on create", "picID", picID)
		}
	}

	return picID, nil
}

// CreatePicFrom creates a new picture from an existing picture
func (gs *GraphicsSystem) CreatePicFrom(srcID int) (int, error) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	picID, err := gs.pictures.CreatePicFrom(srcID)
	if err != nil {
		return picID, err
	}

	// 要件 11.1: CreatePicFromが呼び出されたとき、非表示のPictureSpriteを作成する
	if gs.pictureSpriteManager != nil {
		pic, picErr := gs.pictures.GetPicWithoutLock(picID)
		if picErr == nil && pic != nil && pic.Image != nil {
			gs.pictureSpriteManager.CreatePictureSpriteOnLoad(pic.Image, picID, pic.Width, pic.Height)
			gs.log.Debug("CreatePicFrom: created PictureSprite on create", "picID", picID)
		}
	}

	return picID, nil
}

// CreatePicWithSize は指定されたサイズの空のピクチャーを生成する
// srcID: 参照用のソースピクチャーID（存在確認のみ）
// width, height: 新しいピクチャーのサイズ
// 戻り値: 新しいピクチャーID、エラー
func (gs *GraphicsSystem) CreatePicWithSize(srcID, width, height int) (int, error) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	picID, err := gs.pictures.CreatePicWithSize(srcID, width, height)
	if err != nil {
		return picID, err
	}

	// 要件 11.1: CreatePicWithSizeが呼び出されたとき、非表示のPictureSpriteを作成する
	if gs.pictureSpriteManager != nil {
		pic, picErr := gs.pictures.GetPicWithoutLock(picID)
		if picErr == nil && pic != nil && pic.Image != nil {
			gs.pictureSpriteManager.CreatePictureSpriteOnLoad(pic.Image, picID, pic.Width, pic.Height)
			gs.log.Debug("CreatePicWithSize: created PictureSprite on create", "picID", picID)
		}
	}

	return picID, nil
}

// DelPic deletes a picture
// 要件 11.8: ピクチャが解放されたとき、対応するPictureSpriteとその子スプライトを削除する
func (gs *GraphicsSystem) DelPic(id int) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	// 要件 30.1-30.3: PictureSpriteを削除する
	if gs.pictureSpriteManager != nil {
		gs.pictureSpriteManager.FreePictureSprite(id)
	}

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

// ============================================================================
// パフォーマンス測定機能（タスク 7.1, 7.2, 7.3）
// ============================================================================

// GetFPSCounter はFPSCounterを返す
// タスク 7.2: FPS測定機能
func (gs *GraphicsSystem) GetFPSCounter() *FPSCounter {
	return gs.fpsCounter
}

// GetStatsCollector はSpriteStatsCollectorを返す
// タスク 7.1: スプライト数の測定機能
func (gs *GraphicsSystem) GetStatsCollector() *SpriteStatsCollector {
	return gs.statsCollector
}

// SetPerformanceMonitoringEnabled はパフォーマンス測定の有効/無効を設定する
// タスク 7.3: パフォーマンスボトルネックの特定と改善
func (gs *GraphicsSystem) SetPerformanceMonitoringEnabled(enabled bool) {
	if gs.fpsCounter != nil {
		gs.fpsCounter.SetEnabled(enabled)
	}
	if gs.statsCollector != nil {
		gs.statsCollector.SetEnabled(enabled)
	}
}

// IsPerformanceMonitoringEnabled はパフォーマンス測定が有効かどうかを返す
func (gs *GraphicsSystem) IsPerformanceMonitoringEnabled() bool {
	if gs.fpsCounter != nil && gs.fpsCounter.IsEnabled() {
		return true
	}
	if gs.statsCollector != nil && gs.statsCollector.IsEnabled() {
		return true
	}
	return false
}

// UpdatePerformanceStats はパフォーマンス統計を更新する
// Draw()から呼び出すことを想定
// タスク 7.3: パフォーマンスボトルネックの特定と改善
func (gs *GraphicsSystem) UpdatePerformanceStats() {
	if gs.fpsCounter != nil {
		gs.fpsCounter.Update()
	}
	if gs.statsCollector != nil {
		gs.statsCollector.Update()
	}
}

// GetPerformanceSummary はパフォーマンスサマリーを返す
// タスク 7.3: パフォーマンスボトルネックの特定と改善
func (gs *GraphicsSystem) GetPerformanceSummary() string {
	var summary string

	if gs.fpsCounter != nil && gs.fpsCounter.IsEnabled() {
		summary += gs.fpsCounter.Compact() + " "
	}

	stats := gs.GetSpriteStats()
	if stats != nil {
		summary += stats.Compact()
	}

	return summary
}

// LogPerformanceStats はパフォーマンス統計をログに出力する
// タスク 7.3: パフォーマンスボトルネックの特定と改善
func (gs *GraphicsSystem) LogPerformanceStats() {
	if gs.fpsCounter != nil && gs.fpsCounter.IsEnabled() {
		fpsStats := gs.fpsCounter.GetStats()
		gs.log.Info("FPS Statistics",
			"currentFPS", fpsStats.CurrentFPS,
			"averageFPS", fpsStats.AverageFPS,
			"minFPS", fpsStats.MinFPS,
			"maxFPS", fpsStats.MaxFPS,
			"frameTime", fpsStats.FrameTime,
		)
	}

	stats := gs.GetSpriteStats()
	if stats != nil {
		gs.log.Info("Sprite Statistics",
			"totalSprites", stats.TotalSprites,
			"windowSprites", stats.WindowSprites,
			"pictureSprites", stats.PictureSprites,
			"castSprites", stats.CastSprites,
			"textSprites", stats.TextSprites,
			"shapeSprites", stats.ShapeSprites,
			"pictures", stats.Pictures,
			"windows", stats.Windows,
			"casts", stats.Casts,
		)
	}
}
