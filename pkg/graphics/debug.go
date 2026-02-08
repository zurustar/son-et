package graphics

import (
	"fmt"
	"image/color"
	"log/slog"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// デバッグオーバーレイの色定義
// 要件 15.5: ウィンドウIDを黄色、ピクチャーIDを緑色、キャストIDを黄色で表示する
var (
	debugWindowIDColor  = color.RGBA{255, 255, 0, 255} // 黄色
	debugPictureIDColor = color.RGBA{0, 255, 0, 255}   // 緑色
	debugCastIDColor    = color.RGBA{255, 255, 0, 255} // 黄色
	debugBgColor        = color.RGBA{0, 0, 0, 200}     // 半透明黒
)

// DebugOverlay はデバッグ情報の描画を管理する
// 要件 15.1-15.8: デバッグオーバーレイの実装
type DebugOverlay struct {
	enabled  bool             // デバッグオーバーレイの有効/無効
	fontFace *text.GoTextFace // デバッグ用フォント
	log      *slog.Logger
	mu       sync.RWMutex
}

// NewDebugOverlay は新しい DebugOverlay を作成する
func NewDebugOverlay() *DebugOverlay {
	do := &DebugOverlay{
		enabled: false,
		log:     slog.Default(),
	}
	return do
}

// NewDebugOverlayWithLogger は新しい DebugOverlay をロガー付きで作成する
func NewDebugOverlayWithLogger(log *slog.Logger) *DebugOverlay {
	do := &DebugOverlay{
		enabled: false,
		log:     log,
	}
	return do
}

// SetEnabled はデバッグオーバーレイの有効/無効を設定する
// 要件 15.7, 15.8: ログレベルがDebug未満のとき、デバッグオーバーレイを表示しない
func (do *DebugOverlay) SetEnabled(enabled bool) {
	do.mu.Lock()
	defer do.mu.Unlock()
	do.enabled = enabled
	if do.log != nil {
		do.log.Debug("DebugOverlay enabled state changed", "enabled", enabled)
	}
}

// IsEnabled はデバッグオーバーレイが有効かどうかを返す
func (do *DebugOverlay) IsEnabled() bool {
	do.mu.RLock()
	defer do.mu.RUnlock()
	return do.enabled
}

// DrawWindowID はウィンドウIDをタイトルバーに描画する
// 要件 15.1: ログレベルがDebug以上のとき、ウィンドウIDをタイトルバーに表示する
// 要件 15.6: ウィンドウIDの表示形式は `[W1]`
func (do *DebugOverlay) DrawWindowID(screen *ebiten.Image, win *Window, titleBarX, titleBarY, titleBarWidth int) {
	do.mu.RLock()
	defer do.mu.RUnlock()

	if !do.enabled {
		return
	}

	label := fmt.Sprintf("[W%d]", win.ID)

	// ラベルのサイズを計算（basicfontは7x13ピクセル）
	labelWidth := len(label) * 7

	// タイトルバーの右側に配置
	x := titleBarX + titleBarWidth - labelWidth - 4
	y := titleBarY + 3 // タイトルバー内で少し下にオフセット

	// 黄色でテキストを描画
	do.drawDebugText(screen, label, x, y, debugWindowIDColor, false)
}

// DrawPictureID はピクチャーIDをウィンドウ内容の左上に描画する
// 要件 15.2: ログレベルがDebug以上のとき、ピクチャーIDをウィンドウ内容の左上に表示する
// 要件 15.4: デバッグラベルを半透明の背景付きで表示し、視認性を確保する
// 要件 15.6: ピクチャーIDの表示形式は `P1`
func (do *DebugOverlay) DrawPictureID(screen *ebiten.Image, picID int, x, y int) {
	do.mu.RLock()
	defer do.mu.RUnlock()

	if !do.enabled {
		return
	}

	label := fmt.Sprintf("P%d", picID)

	// 半透明黒背景 + 緑色テキスト
	do.drawDebugText(screen, label, x, y, debugPictureIDColor, true)
}

// DrawCastID はキャストIDをキャスト位置に描画する
// 要件 15.3: ログレベルがDebug以上のとき、キャストIDとソースピクチャーIDをキャスト位置に表示する
// 要件 15.4: デバッグラベルを半透明の背景付きで表示し、視認性を確保する
// 要件 15.6: キャストIDの表示形式は `C1(P2)`
func (do *DebugOverlay) DrawCastID(screen *ebiten.Image, cast *Cast, x, y int) {
	do.mu.RLock()
	defer do.mu.RUnlock()

	if !do.enabled {
		return
	}

	label := fmt.Sprintf("C%d(P%d)", cast.ID, cast.PicID)

	// 半透明黒背景 + 黄色テキスト
	do.drawDebugText(screen, label, x, y, debugCastIDColor, true)
}

// drawDebugText はデバッグテキストを描画する
// withBackground が true の場合、半透明の黒背景を描画する
func (do *DebugOverlay) drawDebugText(screen *ebiten.Image, label string, x, y int, textColor color.Color, withBackground bool) {
	// ラベルのサイズを計算（ebitenutil.DebugPrintAtは6x16ピクセル）
	labelWidth := len(label) * 6
	labelHeight := 16
	padding := 2

	if withBackground {
		// 半透明黒背景を描画
		vector.FillRect(screen,
			float32(x-padding),
			float32(y-padding),
			float32(labelWidth+padding*2),
			float32(labelHeight+padding*2),
			debugBgColor,
			false)
	}

	// テキストを描画
	// Ebitengine v2のtext/v2パッケージを使用
	if do.fontFace != nil {
		op := &text.DrawOptions{}
		op.GeoM.Translate(float64(x), float64(y))
		op.ColorScale.ScaleWithColor(textColor)
		text.Draw(screen, label, do.fontFace, op)
	} else {
		// フォールバック: ebitenutil.DebugPrintAtを使用
		do.drawTextWithBasicFont(screen, label, x, y, textColor)
	}
}

// drawTextWithBasicFont はbasicfontを使用してテキストを描画する
// ebitenutil.DebugPrintAtを使用して効率的に描画
func (do *DebugOverlay) drawTextWithBasicFont(screen *ebiten.Image, label string, x, y int, textColor color.Color) {
	// ebitenutil.DebugPrintAtは白色のテキストを描画する
	// 直接スクリーンに描画（色は背景色で区別）
	ebitenutil.DebugPrintAt(screen, label, x, y)
}

// SetEnabledFromLogLevel はログレベルに基づいてデバッグオーバーレイの有効/無効を設定する
// 要件 15.1, 15.7: ログレベルがDebug（レベル2）以上のとき、デバッグオーバーレイを表示する
// slog.LevelDebug = -4, slog.LevelInfo = 0, slog.LevelWarn = 4, slog.LevelError = 8
// Debug以上 = LevelDebug以下の値
func (do *DebugOverlay) SetEnabledFromLogLevel(level slog.Level) {
	// slogではレベルが低いほど詳細（Debug = -4, Info = 0, Warn = 4, Error = 8）
	// Debug以上 = level <= slog.LevelDebug
	enabled := level <= slog.LevelDebug
	do.SetEnabled(enabled)
}

// SetEnabledFromLogLevelString はログレベル文字列に基づいてデバッグオーバーレイの有効/無効を設定する
// 要件 15.1, 15.7: ログレベルがDebug以上のとき、デバッグオーバーレイを表示する
func (do *DebugOverlay) SetEnabledFromLogLevelString(level string) {
	enabled := level == "debug"
	do.SetEnabled(enabled)
}

// DrawSpriteDebugInfo はスプライトのデバッグ情報を描画する
// SpriteManager.Drawから各スプライト描画直後に呼び出される
// 要件 15.1-15.8: デバッグオーバーレイの実装
//
// スプライトのIDと位置を表示します。
// 半透明の黒背景に黄色のテキストで表示されます。
//
// 例:
//
//	do.DrawSpriteDebugInfo(screen, sprite, absX, absY)
func (do *DebugOverlay) DrawSpriteDebugInfo(screen *ebiten.Image, s *Sprite, absX, absY float64) {
	do.mu.RLock()
	defer do.mu.RUnlock()

	if !do.enabled || s == nil {
		return
	}

	// スプライトIDを表示
	label := fmt.Sprintf("S%d", s.ID())

	// Z_Pathがある場合は追加
	if s.zPath != nil {
		label = fmt.Sprintf("S%d %s", s.ID(), s.zPath.String())
	}

	// 半透明黒背景 + 黄色テキストでスプライト情報を表示
	do.drawDebugText(screen, label, int(absX), int(absY), debugWindowIDColor, true)
}
