// Package sprite provides sprite-based rendering system with slice-based draw ordering.
package sprite

import (
	"fmt"
	"image/color"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// デバッグオーバーレイの色定義
var (
	// スプライト情報の色
	debugSpriteIDColor     = color.RGBA{255, 255, 0, 255}   // 黄色
	debugSpriteHiddenColor = color.RGBA{128, 128, 128, 255} // グレー（非表示スプライト）
	debugBgColor           = color.RGBA{0, 0, 0, 180}       // 半透明黒

	// バウンディングボックスの色
	debugBoundingBoxColor       = color.RGBA{0, 255, 0, 255}   // 緑（表示スプライト）
	debugBoundingBoxHiddenColor = color.RGBA{255, 0, 0, 128}   // 半透明赤（非表示スプライト）
	debugRootBoundingBoxColor   = color.RGBA{0, 128, 255, 255} // 青（ルートスプライト）
)

// デバッグオーバーレイの定数
const (
	debugLabelCharWidth  = 6  // ebitenutil.DebugPrintAtの文字幅
	debugLabelCharHeight = 16 // ebitenutil.DebugPrintAtの文字高さ
	debugLabelPadding    = 2  // ラベルのパディング
	debugBoundingBoxLine = 1  // バウンディングボックスの線幅
)

// DebugOverlayOptions はデバッグオーバーレイの表示オプション
type DebugOverlayOptions struct {
	ShowSpriteInfo    bool // スプライト情報（ID、位置）を表示
	ShowBoundingBoxes bool // バウンディングボックスを表示
	ShowHiddenSprites bool // 非表示スプライトも表示
	ShowChildCount    bool // 子スプライト数を表示
}

// DefaultDebugOverlayOptions はデフォルトのデバッグオーバーレイオプションを返す
func DefaultDebugOverlayOptions() DebugOverlayOptions {
	return DebugOverlayOptions{
		ShowSpriteInfo:    true,
		ShowBoundingBoxes: true,
		ShowHiddenSprites: false,
		ShowChildCount:    true,
	}
}

// DebugOverlay はスプライトシステムのデバッグオーバーレイを管理する
// 要件 20.3: デバッグモードが有効なとき、スプライト情報をオーバーレイ表示できる
type DebugOverlay struct {
	enabled bool
	options DebugOverlayOptions
	mu      sync.RWMutex
}

// NewDebugOverlay は新しいDebugOverlayを作成する
func NewDebugOverlay() *DebugOverlay {
	return &DebugOverlay{
		enabled: false,
		options: DefaultDebugOverlayOptions(),
	}
}

// SetEnabled はデバッグオーバーレイの有効/無効を設定する
func (do *DebugOverlay) SetEnabled(enabled bool) {
	do.mu.Lock()
	defer do.mu.Unlock()
	do.enabled = enabled
}

// IsEnabled はデバッグオーバーレイが有効かどうかを返す
func (do *DebugOverlay) IsEnabled() bool {
	do.mu.RLock()
	defer do.mu.RUnlock()
	return do.enabled
}

// SetOptions はデバッグオーバーレイのオプションを設定する
func (do *DebugOverlay) SetOptions(options DebugOverlayOptions) {
	do.mu.Lock()
	defer do.mu.Unlock()
	do.options = options
}

// GetOptions はデバッグオーバーレイのオプションを取得する
func (do *DebugOverlay) GetOptions() DebugOverlayOptions {
	do.mu.RLock()
	defer do.mu.RUnlock()
	return do.options
}

// Draw はデバッグオーバーレイを描画する
// SpriteManagerのすべてのスプライトに対してデバッグ情報を表示する
func (do *DebugOverlay) Draw(screen *ebiten.Image, sm *SpriteManager) {
	do.mu.RLock()
	defer do.mu.RUnlock()

	if !do.enabled {
		return
	}

	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// ルートスプライトから再帰的に描画
	for _, root := range sm.roots {
		do.drawSpriteOverlay(screen, root, 0, 0, true)
	}
}

// drawSpriteOverlay はスプライトとその子のデバッグオーバーレイを再帰的に描画する
func (do *DebugOverlay) drawSpriteOverlay(screen *ebiten.Image, s *Sprite, parentX, parentY float64, isRoot bool) {
	// 絶対位置を計算
	absX := parentX + s.x
	absY := parentY + s.y

	// 非表示スプライトの処理
	if !s.visible && !do.options.ShowHiddenSprites {
		return
	}

	// バウンディングボックスを描画
	if do.options.ShowBoundingBoxes {
		do.drawBoundingBox(screen, s, absX, absY, isRoot)
	}

	// スプライト情報を描画
	if do.options.ShowSpriteInfo {
		do.drawSpriteInfo(screen, s, absX, absY)
	}

	// 子スプライトを再帰的に処理
	for _, child := range s.children {
		do.drawSpriteOverlay(screen, child, absX, absY, false)
	}
}

// drawBoundingBox はスプライトのバウンディングボックスを描画する
func (do *DebugOverlay) drawBoundingBox(screen *ebiten.Image, s *Sprite, absX, absY float64, isRoot bool) {
	// 画像がない場合はスキップ
	if s.image == nil {
		return
	}

	bounds := s.image.Bounds()
	width := float32(bounds.Dx())
	height := float32(bounds.Dy())

	// 色を選択
	var boxColor color.Color
	if isRoot {
		boxColor = debugRootBoundingBoxColor
	} else if !s.visible {
		boxColor = debugBoundingBoxHiddenColor
	} else {
		boxColor = debugBoundingBoxColor
	}

	// バウンディングボックスを描画（線のみ）
	x := float32(absX)
	y := float32(absY)
	lineWidth := float32(debugBoundingBoxLine)

	// 上辺
	vector.FillRect(screen, x, y, width, lineWidth, boxColor, false)
	// 下辺
	vector.FillRect(screen, x, y+height-lineWidth, width, lineWidth, boxColor, false)
	// 左辺
	vector.FillRect(screen, x, y, lineWidth, height, boxColor, false)
	// 右辺
	vector.FillRect(screen, x+width-lineWidth, y, lineWidth, height, boxColor, false)
}

// drawSpriteInfo はスプライト情報のラベルを描画する
func (do *DebugOverlay) drawSpriteInfo(screen *ebiten.Image, s *Sprite, absX, absY float64) {
	// ラベルを作成
	var label string
	if do.options.ShowChildCount && len(s.children) > 0 {
		label = fmt.Sprintf("S%d (%.0f,%.0f) [%d]", s.id, s.x, s.y, len(s.children))
	} else {
		label = fmt.Sprintf("S%d (%.0f,%.0f)", s.id, s.x, s.y)
	}

	// 非表示の場合は印を追加
	if !s.visible {
		label += " H"
	}

	// ラベルのサイズを計算
	labelWidth := len(label) * debugLabelCharWidth
	labelHeight := debugLabelCharHeight

	// 背景を描画
	bgX := float32(absX) - float32(debugLabelPadding)
	bgY := float32(absY) - float32(debugLabelPadding)
	bgWidth := float32(labelWidth + debugLabelPadding*2)
	bgHeight := float32(labelHeight + debugLabelPadding*2)
	vector.FillRect(screen, bgX, bgY, bgWidth, bgHeight, debugBgColor, false)

	// テキストを描画（ebitenutil.DebugPrintAtは白色固定のため、色の区別は背景色で行う）
	do.drawColoredText(screen, label, int(absX), int(absY))
}

// drawSpriteOverlayInline はスプライトのデバッグオーバーレイを描画する（インライン版）
// SpriteManager.drawSprite()から呼び出され、各スプライトと同じ階層で描画される
// これにより、後から描画されるスプライトによってデバッグ情報が隠れる
func (do *DebugOverlay) drawSpriteOverlayInline(screen *ebiten.Image, s *Sprite, absX, absY float64, isRoot bool) {
	do.mu.RLock()
	defer do.mu.RUnlock()

	if !do.enabled {
		return
	}

	// バウンディングボックスを描画
	if do.options.ShowBoundingBoxes {
		do.drawBoundingBox(screen, s, absX, absY, isRoot)
	}

	// スプライト情報を描画
	if do.options.ShowSpriteInfo {
		do.drawSpriteInfo(screen, s, absX, absY)
	}
}

// drawColoredText はデバッグテキストを描画する
// ebitenutil.DebugPrintAtは白色固定のため、色の区別は背景色で行う
func (do *DebugOverlay) drawColoredText(screen *ebiten.Image, text string, x, y int) {
	ebitenutil.DebugPrintAt(screen, text, x, y)
}

// DrawSingleSprite は単一スプライトのデバッグオーバーレイを描画する
func (do *DebugOverlay) DrawSingleSprite(screen *ebiten.Image, s *Sprite) {
	do.mu.RLock()
	defer do.mu.RUnlock()

	if !do.enabled || s == nil {
		return
	}

	absX, absY := s.AbsolutePosition()

	// バウンディングボックスを描画
	if do.options.ShowBoundingBoxes {
		do.drawBoundingBox(screen, s, absX, absY, s.parent == nil)
	}

	// スプライト情報を描画
	if do.options.ShowSpriteInfo {
		do.drawSpriteInfo(screen, s, absX, absY)
	}
}

// ============================================================================
// SpriteManager へのデバッグオーバーレイ統合
// ============================================================================

// debugOverlay はSpriteManagerのデバッグオーバーレイ
var globalDebugOverlay *DebugOverlay

// init はグローバルデバッグオーバーレイを初期化する
func init() {
	globalDebugOverlay = NewDebugOverlay()
}

// GetDebugOverlay はグローバルデバッグオーバーレイを取得する
func GetDebugOverlay() *DebugOverlay {
	return globalDebugOverlay
}

// SetDebugMode はデバッグモードを設定する
// 要件 20.3: デバッグモードが有効なとき、スプライト情報をオーバーレイ表示できる
func SetDebugMode(enabled bool) {
	globalDebugOverlay.SetEnabled(enabled)
}

// IsDebugMode はデバッグモードが有効かどうかを返す
func IsDebugMode() bool {
	return globalDebugOverlay.IsEnabled()
}

// DrawDebugOverlay はSpriteManagerのデバッグオーバーレイを描画する
// 要件 20.3: デバッグモードが有効なとき、スプライト情報をオーバーレイ表示できる
func (sm *SpriteManager) DrawDebugOverlay(screen *ebiten.Image) {
	globalDebugOverlay.Draw(screen, sm)
}

// SetDebugOverlayOptions はデバッグオーバーレイのオプションを設定する
func SetDebugOverlayOptions(options DebugOverlayOptions) {
	globalDebugOverlay.SetOptions(options)
}

// GetDebugOverlayOptions はデバッグオーバーレイのオプションを取得する
func GetDebugOverlayOptions() DebugOverlayOptions {
	return globalDebugOverlay.GetOptions()
}
