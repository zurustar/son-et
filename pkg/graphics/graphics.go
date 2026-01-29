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
// スプライトシステム移行: LayerManagerは不要になった
type GraphicsSystem struct {
	pictures             *PictureManager
	windows              *WindowManager
	casts                *CastManager
	textRenderer         *TextRenderer
	cmdQueue             *CommandQueue
	sceneChanges         *SceneChangeManager
	debugOverlay         *DebugOverlay
	spriteManager        *SpriteManager        // スプライトシステム要件 3.1〜3.6: SpriteManagerを統合
	windowSpriteManager  *WindowSpriteManager  // スプライトシステム要件 7.1〜7.3: WindowSpriteManagerを統合
	pictureSpriteManager *PictureSpriteManager // スプライトシステム要件 6.1〜6.3: PictureSpriteManagerを統合
	castSpriteManager    *CastSpriteManager    // スプライトシステム要件 8.1〜8.4: CastSpriteManagerを統合
	textSpriteManager    *TextSpriteManager    // スプライトシステム要件 5.1〜5.5: TextSpriteManagerを統合
	shapeSpriteManager   *ShapeSpriteManager   // スプライトシステム要件 9.1〜9.3: ShapeSpriteManagerを統合

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
	gs.spriteManager = NewSpriteManager()                               // スプライトシステム要件 3.1〜3.6: SpriteManagerを初期化
	gs.windowSpriteManager = NewWindowSpriteManager(gs.spriteManager)   // スプライトシステム要件 7.1〜7.3: WindowSpriteManagerを初期化
	gs.pictureSpriteManager = NewPictureSpriteManager(gs.spriteManager) // スプライトシステム要件 6.1〜6.3: PictureSpriteManagerを初期化
	gs.castSpriteManager = NewCastSpriteManager(gs.spriteManager)       // スプライトシステム要件 8.1〜8.4: CastSpriteManagerを初期化
	gs.textSpriteManager = NewTextSpriteManager(gs.spriteManager)       // スプライトシステム要件 5.1〜5.5: TextSpriteManagerを初期化
	gs.shapeSpriteManager = NewShapeSpriteManager(gs.spriteManager)     // スプライトシステム要件 9.1〜9.3: ShapeSpriteManagerを初期化

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
// 要件 10.1, 10.2: PutCast、MovePic、TextWriteの操作順序に基づくZ順序で描画
// スプライトシステム要件 7.1〜7.3: WindowSpriteを使用した描画
// スプライトシステム要件 6.1〜6.3: PictureSpriteを使用した描画
// スプライトシステム要件 14.1: SpriteManager.Draw()ベースの描画
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

		// スプライトシステム要件 7.1〜7.3: WindowSpriteを使用した描画
		// WindowSpriteが存在する場合はスプライトベースで描画
		if gs.windowSpriteManager != nil {
			ws := gs.windowSpriteManager.GetWindowSprite(win.ID)
			if ws != nil {
				gs.drawWindowSpriteDecoration(screen, ws, pic)
			} else {
				// WindowSpriteが存在しない場合は従来の方法で描画
				gs.drawWindowDecoration(screen, win, pic)
			}
		} else {
			// WindowSpriteManagerが存在しない場合は従来の方法で描画
			gs.drawWindowDecoration(screen, win, pic)
		}

		// このウィンドウに属するすべてのレイヤーをZ順序で描画
		// 要件 10.1, 10.2: キャスト、描画エントリ（MovePic）、テキストを操作順序で描画
		gs.drawLayersForWindow(screen, win)

		// デバッグオーバーレイを描画（要件 15.1-15.8）
		gs.drawDebugOverlayForWindow(screen, win, pic)
	}

	// スプライトシステム要件 14.1: SpriteManager.Draw()ベースの描画
	// 注意: 現在は移行期間中なので、以下の描画は行わない
	// - WindowSprite: 上記のループで個別に描画済み
	// - CastSprite: drawLayersForWindow()で描画済み（透明色処理のため）
	// - TextSprite: 従来のTextRendererで描画済み
	// - ShapeSprite: 従来のプリミティブ描画で描画済み
	//
	// 将来的には、すべてのスプライトの親子関係を適切に設定し、
	// SpriteManager.Draw()のみで描画を行う予定
	// 現時点では、スプライトシステムへの移行を段階的に行うため、
	// 従来の描画ロジックを維持している
	//
	// 移行完了後は以下のコードのみで描画が完了する:
	// gs.spriteManager.Draw(screen)
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

// drawLayersForWindow はウィンドウに属するすべての描画要素をZ順序で描画する
// スプライトシステム: すべてのスプライトタイプを統一的に描画する
// 要件 14.3: Z順序の統一（ウインドウ間、ウインドウ内）
func (gs *GraphicsSystem) drawLayersForWindow(screen *ebiten.Image, win *Window) {
	const (
		borderThickness = 4
		titleBarHeight  = 20
	)

	// コンテンツ領域の開始位置を計算
	contentX := win.X + borderThickness
	contentY := win.Y + borderThickness + titleBarHeight

	// 位置オフセット（ピクチャー座標系）
	offsetX := -win.PicX
	offsetY := -win.PicY

	// すべてのスプライトを収集してZ順序でソート
	allSprites := gs.collectAllSpritesForWindow(win)

	// Z順序でソート
	sortSpritesByZOrder(allSprites)

	// 描画
	for _, item := range allSprites {
		gs.drawSpriteItem(screen, item, contentX, contentY, offsetX, offsetY)
	}
}

// spriteItem はスプライトとそのタイプを保持する
type spriteItem struct {
	sprite     *Sprite
	spriteType string
	castSprite *CastSprite // キャストスプライトの場合のみ設定（透明色処理用）
	textSprite *TextSprite // テキストスプライトの場合のみ設定（背景ブレンド用）
}

// collectAllSpritesForWindow はウィンドウに属するすべてのスプライトを収集する
func (gs *GraphicsSystem) collectAllSpritesForWindow(win *Window) []spriteItem {
	var items []spriteItem

	// キャストスプライトを収集
	if gs.castSpriteManager != nil {
		castSprites := gs.castSpriteManager.GetCastSpritesByWindow(win.ID)
		for _, cs := range castSprites {
			if cs.GetSprite() != nil {
				items = append(items, spriteItem{
					sprite:     cs.GetSprite(),
					spriteType: "cast",
					castSprite: cs,
				})
			}
		}
	}

	// ピクチャースプライトを収集
	// 注意: PictureSpriteはWindowSpriteの子として追加されているが、
	// ここでも収集して描画する（MovePicで更新された画像を表示するため）
	if gs.pictureSpriteManager != nil {
		pictureSprites := gs.pictureSpriteManager.GetPictureSprites(win.PicID)
		for _, ps := range pictureSprites {
			if ps != nil && ps.GetSprite() != nil {
				items = append(items, spriteItem{
					sprite:     ps.GetSprite(),
					spriteType: "picture",
				})
			}
		}
	}

	// テキストスプライトを収集
	if gs.textSpriteManager != nil {
		textSprites := gs.textSpriteManager.GetTextSprites(win.PicID)
		for _, ts := range textSprites {
			if ts.GetSprite() != nil {
				items = append(items, spriteItem{
					sprite:     ts.GetSprite(),
					spriteType: "text",
					textSprite: ts,
				})
			}
		}
	}

	// 図形スプライトを収集
	if gs.shapeSpriteManager != nil {
		shapeSprites := gs.shapeSpriteManager.GetShapeSprites(win.PicID)
		for _, ss := range shapeSprites {
			if ss.GetSprite() != nil {
				items = append(items, spriteItem{
					sprite:     ss.GetSprite(),
					spriteType: "shape",
				})
			}
		}
	}

	return items
}

// sortSpritesByZOrder はスプライトをZ順序でソートする
// 階層的Z順序システム: Z_Pathを使用してソートする
// Z_Pathがない場合はzOrderフィールドを使用（フォールバック）
func sortSpritesByZOrder(items []spriteItem) {
	for i := 1; i < len(items); i++ {
		key := items[i]
		j := i - 1
		for j >= 0 {
			if !compareSpritesForSort(items[j], key) {
				break
			}
			items[j+1] = items[j]
			j--
		}
		items[j+1] = key
	}
}

// compareSpritesForSort は2つのスプライトを比較する（ソート用）
// aがbより後に描画されるべき場合（aがbより大きい場合）trueを返す
func compareSpritesForSort(a, b spriteItem) bool {
	// 両方ともZ_Pathを持つ場合は辞書順比較
	if a.sprite != nil && b.sprite != nil {
		aZPath := a.sprite.GetZPath()
		bZPath := b.sprite.GetZPath()

		if aZPath != nil && bZPath != nil {
			// aがbより大きい（後に描画される）場合true
			return aZPath.Compare(bZPath) > 0
		}

		// 片方だけZ_Pathを持つ場合
		// Z_Pathを持たないスプライトを先に描画（背面）
		if aZPath == nil && bZPath != nil {
			return false // aは先に描画される
		}
		if aZPath != nil && bZPath == nil {
			return true // aは後に描画される
		}
	}

	// 両方ともZ_Pathを持たない場合はIDで比較（安定ソート）
	aID := 0
	bID := 0
	if a.sprite != nil {
		aID = a.sprite.ID()
	}
	if b.sprite != nil {
		bID = b.sprite.ID()
	}
	return aID > bID
}

// drawSpriteItem はスプライトアイテムを描画する
func (gs *GraphicsSystem) drawSpriteItem(screen *ebiten.Image, item spriteItem, contentX, contentY, offsetX, offsetY int) {
	sprite := item.sprite
	if sprite == nil || sprite.Image() == nil {
		return
	}

	// 可視性チェック
	if !sprite.IsEffectivelyVisible() {
		// デバッグ: 可視性チェックで除外されたスプライトをログ出力
		parent := sprite.Parent()
		parentVisible := true
		parentID := -1
		if parent != nil {
			parentVisible = parent.Visible()
			parentID = parent.ID()
		}
		gs.log.Debug("drawSpriteItem: sprite not visible",
			"spriteID", sprite.ID(),
			"spriteType", item.spriteType,
			"visible", sprite.Visible(),
			"parentID", parentID,
			"parentVisible", parentVisible,
		)
		return
	}

	// 描画位置を計算（ピクチャー座標系 → スクリーン座標）
	x, y := sprite.Position()
	screenX := float64(contentX+offsetX) + x
	screenY := float64(contentY+offsetY) + y

	// キャストスプライトの場合は透明色処理が必要
	if item.spriteType == "cast" && item.castSprite != nil && item.castSprite.HasTransColor() {
		gs.drawCastWithTransparency(screen, sprite.Image(), int(screenX), int(screenY), item.castSprite.GetTransColor(), true)
		return
	}

	// テキストスプライトの場合は背景とブレンドして描画
	if item.spriteType == "text" && item.textSprite != nil {
		gs.drawTextSpriteWithBackground(screen, item.textSprite, int(screenX), int(screenY))
		return
	}

	// 通常描画
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(screenX, screenY)

	alpha := sprite.EffectiveAlpha()
	if alpha < 1.0 {
		op.ColorScale.ScaleAlpha(float32(alpha))
	}

	screen.DrawImage(sprite.Image(), op)
}

// drawTextSpriteWithBackground はテキストスプライトを背景とブレンドして描画する
// 元のスプライト画像（差分抽出済み）を背景の上にアルファブレンディングで描画する
func (gs *GraphicsSystem) drawTextSpriteWithBackground(screen *ebiten.Image, ts *TextSprite, screenX, screenY int) {
	sprite := ts.GetSprite()
	if sprite == nil || sprite.Image() == nil {
		return
	}

	// 通常のアルファブレンディング描画
	// TextSpriteの画像は既に差分抽出されており、透明背景 + 文字になっている
	// Ebitengineのデフォルトのアルファブレンディングで正しく描画される
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(screenX), float64(screenY))

	alpha := sprite.EffectiveAlpha()
	if alpha < 1.0 {
		op.ColorScale.ScaleAlpha(float32(alpha))
	}

	screen.DrawImage(sprite.Image(), op)
}

// drawCastSpritesForWindow はウィンドウに属するすべてのCastSpriteを描画する
func (gs *GraphicsSystem) drawCastSpritesForWindow(screen *ebiten.Image, win *Window, contentX, contentY, offsetX, offsetY int) {
	if gs.castSpriteManager == nil {
		return
	}

	// このウィンドウに属するCastSpriteを取得
	castSprites := gs.castSpriteManager.GetCastSpritesByWindow(win.ID)
	if len(castSprites) == 0 {
		return
	}

	// Z_Pathでソート
	sortCastSpritesByZPath(castSprites)

	// 描画
	for _, cs := range castSprites {
		gs.drawCastSpriteOnScreen(screen, cs, contentX, contentY, offsetX, offsetY)
	}
}

// sortCastSpritesByZPath はCastSpriteをZ_Pathでソートする
func sortCastSpritesByZPath(sprites []*CastSprite) {
	for i := 1; i < len(sprites); i++ {
		key := sprites[i]
		keyZPath := (*ZPath)(nil)
		if key.GetSprite() != nil {
			keyZPath = key.GetSprite().GetZPath()
		}
		j := i - 1
		for j >= 0 {
			jZPath := (*ZPath)(nil)
			if sprites[j].GetSprite() != nil {
				jZPath = sprites[j].GetSprite().GetZPath()
			}
			// Z_Pathで比較（nilは先に描画）
			if jZPath == nil || (keyZPath != nil && !keyZPath.Less(jZPath)) {
				break
			}
			sprites[j+1] = sprites[j]
			j--
		}
		sprites[j+1] = key
	}
}

// drawCastSpriteOnScreen はCastSpriteをスクリーンに描画する
// スプライトシステム要件 8.1〜8.4: CastSpriteを使用した描画
func (gs *GraphicsSystem) drawCastSpriteOnScreen(screen *ebiten.Image, cs *CastSprite, contentX, contentY, offsetX, offsetY int) {
	sprite := cs.GetSprite()
	if sprite == nil || sprite.Image() == nil {
		return
	}

	// 可視性チェック
	if !sprite.IsEffectivelyVisible() {
		return
	}

	// キャストの描画位置を計算（ピクチャー座標系 → スクリーン座標）
	x, y := sprite.Position()
	screenX := float64(contentX+offsetX) + x
	screenY := float64(contentY+offsetY) + y

	// 透明色処理が必要な場合
	if cs.HasTransColor() {
		// 透明色処理を適用して描画
		gs.drawCastWithTransparency(screen, sprite.Image(), int(screenX), int(screenY), cs.GetTransColor(), true)
		return
	}

	// 透明色処理が不要な場合は通常描画
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(screenX, screenY)

	alpha := sprite.EffectiveAlpha()
	if alpha < 1.0 {
		op.ColorScale.ScaleAlpha(float32(alpha))
	}

	screen.DrawImage(sprite.Image(), op)
}

// drawDrawingEntryOnScreen は描画エントリをスクリーンに描画する
// スプライトシステム要件 6.1〜6.3: PictureSpriteが存在する場合はスプライトベースで描画
// 注意: スプライトシステムが完全に統合された後は、この関数は不要になる
func (gs *GraphicsSystem) drawDrawingEntryOnScreen(screen *ebiten.Image, entry *DrawingEntry, contentX, contentY, offsetX, offsetY int) {
	// スプライトシステムが有効な場合、PictureSpriteで描画されるためスキップ
	// 注意: 現在は移行期間中なので、スプライトシステムが有効でも従来の描画を行う
	// 将来的には、スプライトシステムが完全に統合された後、この関数は削除される

	img := entry.GetImage()
	if img == nil {
		return
	}

	// 描画位置を計算（ピクチャー座標系 → スクリーン座標）
	screenX := contentX + offsetX + entry.GetDestX()
	screenY := contentY + offsetY + entry.GetDestY()

	// 描画
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(screenX), float64(screenY))
	screen.DrawImage(img, opts)
}

// drawTextLayerOnScreen はテキストレイヤーをスクリーンに描画する
// スプライトシステム要件 5.1〜5.5: TextSpriteを使用した描画
func (gs *GraphicsSystem) drawTextLayerOnScreen(screen *ebiten.Image, textLayer *TextLayerEntry, contentX, contentY, offsetX, offsetY int) {
	img := textLayer.GetImage()
	if img == nil {
		return
	}

	// 描画位置を計算（ピクチャー座標系 → スクリーン座標）
	x, y := textLayer.GetPosition()
	screenX := contentX + offsetX + x
	screenY := contentY + offsetY + y

	// 描画
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(screenX), float64(screenY))
	screen.DrawImage(img, opts)
}

// drawTextSpriteOnScreen はTextSpriteをスクリーンに描画する
// スプライトシステム要件 5.1〜5.5: TextSpriteを使用した描画
func (gs *GraphicsSystem) drawTextSpriteOnScreen(screen *ebiten.Image, ts *TextSprite, contentX, contentY, offsetX, offsetY int) {
	sprite := ts.GetSprite()
	if sprite == nil || sprite.Image() == nil {
		return
	}

	// 可視性チェック
	if !sprite.IsEffectivelyVisible() {
		return
	}

	// テキストの描画位置を計算（ピクチャー座標系 → スクリーン座標）
	x, y := sprite.Position()
	screenX := float64(contentX+offsetX) + x
	screenY := float64(contentY+offsetY) + y

	// 描画
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(screenX, screenY)

	alpha := sprite.EffectiveAlpha()
	if alpha < 1.0 {
		op.ColorScale.ScaleAlpha(float32(alpha))
	}

	screen.DrawImage(sprite.Image(), op)
}

// drawTextSpritesForWindow はウィンドウに属するすべてのTextSpriteを描画する
// スプライトシステム要件 5.1〜5.5: TextSpriteを使用した描画
func (gs *GraphicsSystem) drawTextSpritesForWindow(screen *ebiten.Image, win *Window, contentX, contentY, offsetX, offsetY int) {
	if gs.textSpriteManager == nil {
		return
	}

	// ウィンドウに関連するピクチャーIDを取得
	picID := win.PicID

	// このピクチャーに属するTextSpriteを取得
	textSprites := gs.textSpriteManager.GetTextSprites(picID)
	if len(textSprites) == 0 {
		return
	}

	// Z_Pathでソート
	sortTextSpritesByZPath(textSprites)

	// 描画
	for _, ts := range textSprites {
		gs.drawTextSpriteOnScreen(screen, ts, contentX, contentY, offsetX, offsetY)
	}
}

// sortTextSpritesByZPath はTextSpriteをZ_Pathでソートする
func sortTextSpritesByZPath(sprites []*TextSprite) {
	for i := 1; i < len(sprites); i++ {
		key := sprites[i]
		keyZPath := (*ZPath)(nil)
		if key.GetSprite() != nil {
			keyZPath = key.GetSprite().GetZPath()
		}
		j := i - 1
		for j >= 0 {
			jZPath := (*ZPath)(nil)
			if sprites[j].GetSprite() != nil {
				jZPath = sprites[j].GetSprite().GetZPath()
			}
			// Z_Pathで比較（nilは先に描画）
			if jZPath == nil || (keyZPath != nil && !keyZPath.Less(jZPath)) {
				break
			}
			sprites[j+1] = sprites[j]
			j--
		}
		sprites[j+1] = key
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

	// PicX/PicYオフセット（ピクチャー座標系 → スクリーン座標）
	offsetX := -win.PicX
	offsetY := -win.PicY

	// ピクチャーIDをピクチャーの左上に描画（要件 15.2）
	// ピクチャー座標(0,0)をスクリーン座標に変換
	picScreenX := contentX + offsetX
	picScreenY := contentY + offsetY
	gs.debugOverlay.DrawPictureID(screen, win.PicID, picScreenX+2, picScreenY+2)

	// このウィンドウに属するキャストのデバッグ情報を描画（要件 15.3）
	casts := gs.casts.GetCastsByWindow(win.ID)
	for _, cast := range casts {
		if !cast.Visible {
			continue
		}

		// キャストの描画位置を計算（ピクチャー座標系 → スクリーン座標）
		// キャストの中央に表示
		castCenterX := contentX + offsetX + cast.X + cast.Width/2
		castCenterY := contentY + offsetY + cast.Y + cast.Height/2

		// キャストIDを描画
		gs.debugOverlay.DrawCastID(screen, cast, castCenterX, castCenterY)
	}

	// このウィンドウに属するテキストスプライトのデバッグ情報を描画
	if gs.textSpriteManager != nil {
		textSprites := gs.textSpriteManager.GetTextSprites(win.PicID)
		for _, ts := range textSprites {
			if ts == nil || ts.GetSprite() == nil {
				continue
			}
			sprite := ts.GetSprite()
			if !sprite.IsEffectivelyVisible() {
				continue
			}

			// テキストスプライトの描画位置を計算（ピクチャー座標系 → スクリーン座標）
			x, y := sprite.Position()
			textScreenX := contentX + offsetX + int(x)
			textScreenY := contentY + offsetY + int(y)

			// テキストスプライトIDを描画
			gs.debugOverlay.DrawTextSpriteID(screen, sprite.ID(), ts.GetText(), textScreenX, textScreenY)
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

// GetLayerManager はLayerManagerを返す（後方互換性のために残す、nilを返す）
// Deprecated: スプライトシステム移行により不要になった
func (gs *GraphicsSystem) GetLayerManager() *LayerManager {
	// スプライトシステム移行により、LayerManagerは不要になった
	return nil
}

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

// VM Interface Implementation
// These methods implement the GraphicsSystemInterface for VM integration

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
// スプライトシステム要件 7.1: ウィンドウが開かれたときにWindowSpriteを作成する
// 要件 11.4: すべての描画要素をスプライトとして管理する
func (gs *GraphicsSystem) OpenWin(picID int, opts ...any) (int, error) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	// Convert any options to WinOption
	winOpts := gs.parseWinOptions(opts)

	// ウィンドウを開く
	winID, err := gs.windows.OpenWin(picID, winOpts...)
	if err != nil {
		return -1, err
	}

	// ウィンドウの情報を取得
	win, err := gs.windows.GetWin(winID)
	if err != nil {
		// ウィンドウが作成されたが取得できない場合はエラー
		gs.log.Error("OpenWin: failed to get window after creation", "winID", winID, "error", err)
		return winID, nil // ウィンドウは作成されているので、IDは返す
	}

	// ウィンドウのサイズを取得（設定されていない場合はピクチャーのサイズを使用）
	width := win.Width
	height := win.Height
	var pic *Picture
	if width <= 0 || height <= 0 {
		// ピクチャーのサイズを取得
		pic, err = gs.pictures.GetPicWithoutLock(picID)
		if err == nil {
			width = pic.Width
			height = pic.Height
		} else {
			// デフォルトサイズ
			width = 640
			height = 480
		}
	} else {
		// ピクチャーを取得（WindowSprite作成用）
		pic, _ = gs.pictures.GetPicWithoutLock(picID)
	}

	// スプライトシステム要件 7.1: WindowSpriteを作成する
	var ws *WindowSprite
	if pic != nil && gs.windowSpriteManager != nil {
		ws = gs.windowSpriteManager.CreateWindowSprite(win, pic)
		gs.log.Debug("OpenWin: created WindowSprite", "winID", winID)
	}

	// 要件 11.3, 11.4: PictureSpriteをウインドウに関連付ける
	// LoadPic時に作成されたPictureSpriteがあれば、それをWindowSpriteの子として関連付ける
	// なければ、新しいPictureSpriteを作成する
	if pic != nil && pic.Image != nil && gs.pictureSpriteManager != nil && ws != nil {
		// まず、LoadPic時に作成されたPictureSpriteを関連付けを試みる
		existingPs := gs.pictureSpriteManager.GetPictureSpriteByPictureID(picID)
		if existingPs != nil {
			// 既存のPictureSpriteをWindowSpriteの子として関連付ける
			// AttachPictureSpriteToWindowは関連付け後にpictureSpriteMapから削除する
			gs.pictureSpriteManager.AttachPictureSpriteToWindow(picID, ws.GetSprite(), winID)
			gs.log.Debug("OpenWin: attached existing PictureSprite to WindowSprite",
				"winID", winID, "picID", picID,
				"zPath", existingPs.GetSprite().GetZPath())
		} else {
			// PictureSpriteの位置は(0, 0)に設定
			// PicX/PicYオフセットはdrawLayersForWindowで適用される
			destX := 0
			destY := 0

			// Z順序を計算（背景レイヤー）
			zOrder := CalculateGlobalZOrder(win.ZOrder, ZOrderBackground)

			// ピクチャー画像全体をPictureSpriteとして作成
			// 注意: 背景PictureSpriteはピクチャーの画像への参照を保持する
			// MovePicで更新されたピクチャー画像が反映されるようにするため
			ps := gs.pictureSpriteManager.CreateBackgroundPictureSprite(
				pic.Image,
				picID,
				pic.Width, pic.Height,
				destX, destY,
				zOrder,
			)

			// WindowSpriteの子として追加
			if ps != nil {
				ps.GetSprite().SetParent(ws.GetSprite())
				ws.AddChild(ps.GetSprite())

				// 要件 1.4: 背景PictureSpriteにZ_Pathを設定
				// WindowSpriteのZ_Pathを継承し、Local_Z_Order=0を追加
				// これにより、CastSpriteやTextSpriteの親として使用できる
				if ws.GetSprite().GetZPath() != nil {
					localZOrder := gs.spriteManager.GetZOrderCounter().GetNext(ws.GetSprite().ID())
					zPath := NewZPathFromParent(ws.GetSprite().GetZPath(), localZOrder)
					ps.GetSprite().SetZPath(zPath)
					gs.spriteManager.MarkNeedSort()
				}

				gs.log.Debug("OpenWin: created new background PictureSprite as child of WindowSprite",
					"winID", winID, "picID", picID, "destX", destX, "destY", destY,
					"zPath", ps.GetSprite().GetZPath())
			}
		}
	}

	gs.log.Debug("OpenWin: window opened", "winID", winID, "width", width, "height", height)

	// スプライト構成をダンプ
	gs.dumpSpriteState(fmt.Sprintf("OpenWin(winID=%d, picID=%d)", winID, picID))

	return winID, nil
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

	// ピクチャーIDが変更される場合、スプライトシステムを更新
	if len(opts) >= 1 {
		if newPicID, ok := toIntFromAny(opts[0]); ok {
			gs.updateWindowSpriteForMoveWin(id, newPicID)
		}
	}

	return gs.windows.MoveWin(id, winOpts...)
}

// updateWindowSpriteForMoveWin はMoveWin時にWindowSpriteの子PictureSpriteを更新する
func (gs *GraphicsSystem) updateWindowSpriteForMoveWin(winID, newPicID int) {
	if gs.windowSpriteManager == nil || gs.pictureSpriteManager == nil {
		return
	}

	// WindowSpriteを取得
	ws := gs.windowSpriteManager.GetWindowSprite(winID)
	if ws == nil {
		gs.log.Debug("MoveWin: WindowSprite not found", "winID", winID)
		return
	}

	// 新しいピクチャーを取得
	newPic, err := gs.pictures.GetPicWithoutLock(newPicID)
	if err != nil {
		gs.log.Warn("MoveWin: new picture not found", "picID", newPicID, "error", err)
		return
	}

	// 既存のPictureSpriteを探す（WindowSpriteの子の中から）
	windowSprite := ws.GetSprite()
	if windowSprite == nil {
		return
	}

	// 方法1: pictureSpriteMapから未関連付けのPictureSpriteを取得してAttach
	existingPs := gs.pictureSpriteManager.GetPictureSpriteByPictureID(newPicID)
	if existingPs != nil {
		// 未関連付けのPictureSpriteがある場合、それをWindowSpriteの子として関連付け
		gs.pictureSpriteManager.AttachPictureSpriteToWindow(newPicID, windowSprite, winID)
		gs.log.Debug("MoveWin: attached existing PictureSprite to WindowSprite",
			"winID", winID, "picID", newPicID,
			"zPath", existingPs.GetSprite().GetZPath())
	} else {
		// 方法2: 既存のPictureSpriteの画像を更新
		// WindowSpriteの最初の子（背景PictureSprite）の画像を更新
		children := windowSprite.GetChildren()
		for _, child := range children {
			// 最初に見つかったPictureSpriteの画像を更新
			if child != nil && child.Image() != nil {
				child.SetImage(newPic.Image)
				gs.log.Debug("MoveWin: updated PictureSprite image",
					"winID", winID, "picID", newPicID,
					"spriteID", child.ID())
				break
			}
		}
	}

	// スプライト構成をダンプ
	gs.dumpSpriteState(fmt.Sprintf("MoveWin(winID=%d, newPicID=%d)", winID, newPicID))
}

// CloseWin closes a window
// CloseWin closes a window
// スプライトシステム要件 7.3: ウィンドウが閉じられたときにWindowSpriteを削除する
// スプライトシステム要件 8.3: ウィンドウが閉じられたときにCastSpriteを削除する
func (gs *GraphicsSystem) CloseWin(id int) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	// スプライトシステム要件 8.3: CastSpriteを削除する
	if gs.castSpriteManager != nil {
		gs.castSpriteManager.RemoveCastSpritesByWindow(id)
		gs.log.Debug("CloseWin: deleted CastSprites", "winID", id)
	}

	// Delete casts belonging to this window (要件 9.2)
	gs.casts.DeleteCastsByWindow(id)

	// スプライトシステム要件 7.3: WindowSpriteを削除する
	if gs.windowSpriteManager != nil {
		gs.windowSpriteManager.RemoveWindowSprite(id)
		gs.log.Debug("CloseWin: deleted WindowSprite", "winID", id)
	}

	return gs.windows.CloseWin(id)
}

// CloseWinAll closes all windows
// スプライトシステム要件 7.3: ウィンドウが閉じられたときにWindowSpriteを削除する
// スプライトシステム要件 8.3: ウィンドウが閉じられたときにCastSpriteを削除する
func (gs *GraphicsSystem) CloseWinAll() {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	// Get all windows and delete their casts
	windows := gs.windows.GetWindowsOrdered()
	for _, win := range windows {
		// スプライトシステム要件 8.3: CastSpriteを削除する
		if gs.castSpriteManager != nil {
			gs.castSpriteManager.RemoveCastSpritesByWindow(win.ID)
		}

		gs.casts.DeleteCastsByWindow(win.ID)
	}

	// スプライトシステム要件 7.3: すべてのWindowSpriteを削除する
	if gs.windowSpriteManager != nil {
		gs.windowSpriteManager.Clear()
		gs.log.Debug("CloseWinAll: deleted all WindowSprites")
	}

	// スプライトシステム要件 8.3: すべてのCastSpriteを削除する
	if gs.castSpriteManager != nil {
		gs.castSpriteManager.Clear()
		gs.log.Debug("CloseWinAll: deleted all CastSprites")
	}

	// スプライトシステム要件 5.1〜5.5: すべてのTextSpriteを削除する
	if gs.textSpriteManager != nil {
		gs.textSpriteManager.Clear()
		gs.log.Debug("CloseWinAll: deleted all TextSprites")
	}

	gs.windows.CloseWinAll()
	gs.log.Debug("CloseWinAll: deleted all WindowLayerSets", "windowCount", len(windows))
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
// スプライトシステム要件 8.1: キャストをスプライトとして作成する
// 要件 9.2: ピクチャ内にキャストが配置されたとき、キャストをピクチャの子スプライトとして管理する
// 要件 11.5: ウインドウに関連付けられていないピクチャに対するCastSetでも、PictureSpriteを親として設定する
// 要件 14.3: Z順序の統一（ウインドウ間、ウインドウ内）
func (gs *GraphicsSystem) PutCast(winID, picID, x, y, srcX, srcY, w, h int) (int, error) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	// キャストを作成
	castID, err := gs.casts.PutCast(winID, picID, x, y, srcX, srcY, w, h)
	if err != nil {
		return -1, err
	}

	// スプライトシステム要件 8.1: CastSpriteを作成する
	// 要件 9.2, 11.5: PictureSpriteを親として設定する（キャストはピクチャーに配置される）
	// 要件 14.3: グローバルZ順序を使用
	if gs.castSpriteManager != nil {
		cast, err := gs.casts.GetCast(castID)
		if err == nil && cast != nil {
			// ソース画像を取得
			srcPic, err := gs.pictures.GetPicWithoutLock(picID)
			if err == nil && srcPic != nil && srcPic.Image != nil {
				// ウインドウのZ順序を取得
				win, winErr := gs.windows.GetWin(winID)
				windowZOrder := 0
				var winPicID int
				if winErr == nil && win != nil {
					windowZOrder = win.ZOrder
					winPicID = win.PicID
				}

				// Z順序を計算（グローバルZ順序）
				localZOrder := ZOrderCastBase + cast.ZOrder
				zOrder := CalculateGlobalZOrder(windowZOrder, localZOrder)

				// 要件 9.2, 11.5: PictureSpriteを親として取得（キャストはピクチャーに配置される）
				// まずpictureSpriteMapから取得を試みる（LoadPic時に作成されたPictureSprite）
				// 見つからない場合は従来の方法（GetBackgroundPictureSpriteSprite）を使用
				var parentSprite *Sprite
				if gs.pictureSpriteManager != nil {
					ps := gs.pictureSpriteManager.GetPictureSpriteByPictureID(winPicID)
					if ps != nil {
						parentSprite = ps.GetSprite()
					} else {
						// フォールバック: 従来の方法
						parentSprite = gs.pictureSpriteManager.GetBackgroundPictureSpriteSprite(winPicID)
					}
				}

				// CastSpriteを作成（親スプライト付き）
				cs := gs.castSpriteManager.CreateCastSpriteWithParent(cast, srcPic.Image, zOrder, parentSprite)
				if cs != nil && parentSprite != nil {
					// PictureSpriteの子として登録
					parentSprite.AddChild(cs.GetSprite())
				}
				gs.log.Debug("PutCast: created CastSprite", "castID", castID, "winID", winID, "winPicID", winPicID, "hasParent", parentSprite != nil, "globalZOrder", zOrder)
			}
		}
	}

	// スプライト構成をダンプ
	gs.dumpSpriteState(fmt.Sprintf("PutCast(castID=%d, winID=%d, picID=%d)", castID, winID, picID))

	return castID, nil
}

// PutCastWithTransColor places a cast on a window with transparent color
// スプライトシステム要件 8.1, 8.4: キャストをスプライトとして作成し、透明色処理をサポートする
// 要件 9.2: ピクチャ内にキャストが配置されたとき、キャストをピクチャの子スプライトとして管理する
// 要件 11.5: ウインドウに関連付けられていないピクチャに対するCastSetでも、PictureSpriteを親として設定する
// 要件 14.3: Z順序の統一（ウインドウ間、ウインドウ内）
func (gs *GraphicsSystem) PutCastWithTransColor(winID, picID, x, y, srcX, srcY, w, h int, transColor color.Color) (int, error) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	// キャストを作成
	castID, err := gs.casts.PutCastWithTransColor(winID, picID, x, y, srcX, srcY, w, h, transColor)
	if err != nil {
		return -1, err
	}

	// スプライトシステム要件 8.1, 8.4: CastSpriteを作成する（透明色付き）
	// 要件 9.2, 11.5: PictureSpriteを親として設定する（キャストはピクチャーに配置される）
	// 要件 14.3: グローバルZ順序を使用
	if gs.castSpriteManager != nil {
		cast, err := gs.casts.GetCast(castID)
		if err == nil && cast != nil {
			// ソース画像を取得
			srcPic, err := gs.pictures.GetPicWithoutLock(picID)
			if err == nil && srcPic != nil && srcPic.Image != nil {
				// ウインドウのZ順序を取得
				win, winErr := gs.windows.GetWin(winID)
				windowZOrder := 0
				var winPicID int
				if winErr == nil && win != nil {
					windowZOrder = win.ZOrder
					winPicID = win.PicID
				}

				// Z順序を計算（グローバルZ順序）
				localZOrder := ZOrderCastBase + cast.ZOrder
				zOrder := CalculateGlobalZOrder(windowZOrder, localZOrder)

				// 要件 9.2, 11.5: PictureSpriteを親として取得（キャストはピクチャーに配置される）
				// まずpictureSpriteMapから取得を試みる（LoadPic時に作成されたPictureSprite）
				// 見つからない場合は従来の方法（GetBackgroundPictureSpriteSprite）を使用
				var parentSprite *Sprite
				if gs.pictureSpriteManager != nil {
					ps := gs.pictureSpriteManager.GetPictureSpriteByPictureID(winPicID)
					if ps != nil {
						parentSprite = ps.GetSprite()
					} else {
						// フォールバック: 従来の方法
						parentSprite = gs.pictureSpriteManager.GetBackgroundPictureSpriteSprite(winPicID)
					}
				}

				// CastSpriteを作成（親スプライト付き、透明色付き）
				cs := gs.castSpriteManager.CreateCastSpriteWithTransColorAndParent(cast, srcPic.Image, zOrder, transColor, parentSprite)
				if cs != nil && parentSprite != nil {
					// PictureSpriteの子として登録
					parentSprite.AddChild(cs.GetSprite())
				}
				gs.log.Debug("PutCastWithTransColor: created CastSprite", "castID", castID, "winID", winID, "winPicID", winPicID, "hasParent", parentSprite != nil, "globalZOrder", zOrder)
			}
		}
	}

	// スプライト構成をダンプ
	gs.dumpSpriteState(fmt.Sprintf("PutCastWithTransColor(castID=%d, winID=%d, picID=%d)", castID, winID, picID))

	return castID, nil
}

// MoveCast moves a cast
// スプライトシステム要件 8.2: キャストの位置を移動できる（残像なし）
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

	// キャストを更新
	if err := gs.casts.MoveCast(id, castOpts...); err != nil {
		return err
	}

	// スプライトシステム要件 8.2: CastSpriteを更新する
	gs.updateCastSprite(id)

	return nil
}

// MoveCastWithOptions moves a cast with explicit options
// キャストはスプライトとして動作し、位置/ソースの更新のみを行う
// 実際の描画は毎フレームdrawCastsForWindowで行われる
// スプライトシステム要件 8.2: キャストの位置を移動できる（残像なし）
func (gs *GraphicsSystem) MoveCastWithOptions(id int, opts ...CastOption) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	// キャストの位置/ソースを更新
	if err := gs.casts.MoveCast(id, opts...); err != nil {
		return err
	}

	// スプライトシステム要件 8.2: CastSpriteを更新する
	gs.updateCastSprite(id)

	return nil
}

// updateCastSprite はCastSpriteを更新する（内部用）
func (gs *GraphicsSystem) updateCastSprite(castID int) {
	if gs.castSpriteManager == nil {
		return
	}

	cs := gs.castSpriteManager.GetCastSprite(castID)
	if cs == nil {
		return
	}

	cast, err := gs.casts.GetCast(castID)
	if err != nil || cast == nil {
		return
	}

	// 位置を更新
	cs.UpdatePosition(cast.X, cast.Y)

	// ソース領域を更新（常に更新してdirtyフラグを設定）
	// MoveCastでソース領域が変更された場合、キャッシュを再構築する必要がある
	cs.UpdateSource(cast.SrcX, cast.SrcY, cast.Width, cast.Height)

	// ピクチャーIDが変更された場合も更新
	if cs.GetSrcPicID() != cast.PicID {
		cs.UpdatePicID(cast.PicID)
	}

	// ソース領域が変更された場合はキャッシュを再構築
	if cs.IsDirty() {
		srcPic, err := gs.pictures.GetPicWithoutLock(cast.PicID)
		if err == nil && srcPic != nil && srcPic.Image != nil {
			cs.RebuildCache(srcPic.Image)
		}
	}

	// 可視性を更新
	cs.UpdateVisible(cast.Visible)
}

// DelCast deletes a cast
// スプライトシステム要件 8.3: キャストを削除できる
// 要件 14.2: ウインドウ内のスプライトをウインドウの子スプライトとして管理する
func (gs *GraphicsSystem) DelCast(id int) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	// スプライトシステム要件 8.3: CastSpriteを削除する
	// 要件 14.2: WindowSpriteの子リストからも削除する
	if gs.castSpriteManager != nil {
		cs := gs.castSpriteManager.GetCastSprite(id)
		if cs != nil && gs.windowSpriteManager != nil {
			cast := cs.GetCast()
			if cast != nil {
				ws := gs.windowSpriteManager.GetWindowSprite(cast.WinID)
				if ws != nil {
					ws.RemoveChild(cs.GetSprite().ID())
				}
			}
		}
		gs.castSpriteManager.RemoveCastSprite(id)
	}

	return gs.casts.DelCast(id)
}

// Text rendering

// TextWrite writes text to a picture
// スプライトシステム要件 5.1〜5.5: TextSpriteを作成する
// 要件 11.6: ウインドウに関連付けられていないピクチャに対するTextWriteでも、PictureSpriteを親として設定する
// 要件 14.2: ウインドウ内のスプライトをウインドウの子スプライトとして管理する
// 要件 14.3: Z順序の統一（ウインドウ間、ウインドウ内）
func (gs *GraphicsSystem) TextWrite(picID, x, y int, text string) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	pic, err := gs.pictures.GetPicWithoutLock(picID)
	if err != nil {
		gs.log.Error("TextWrite: picture not found", "picID", picID, "error", err)
		return err
	}

	// 範囲チェック: テキストがピクチャーの範囲外の場合はスキップ
	// y座標がピクチャーの高さを超えている場合、テキストは表示されない
	if y >= pic.Height || y < -100 || x >= pic.Width || x < -1000 {
		gs.log.Debug("TextWrite: text position out of bounds, skipping",
			"picID", picID, "x", x, "y", y, "picWidth", pic.Width, "picHeight", pic.Height)
		return nil
	}

	// 従来のTextRendererでテキストを描画（レガシー互換性のため）
	if err := gs.textRenderer.TextWrite(pic, x, y, text); err != nil {
		return err
	}

	// 要件: TextWriteでpic.Imageが新しい画像に置き換えられるため、
	// PictureSpriteの画像も更新する必要がある
	if gs.pictureSpriteManager != nil {
		gs.pictureSpriteManager.UpdatePictureSpriteImage(picID, pic.Image)
	}

	// スプライトシステム要件 5.1〜5.5: TextSpriteを作成する
	// 要件 9.3, 11.6: PictureSpriteを親として設定する（テキストはピクチャーに配置される）
	// 要件 14.3: グローバルZ順序を使用
	if gs.textSpriteManager != nil {
		// テキスト設定を取得
		textSettings := gs.textRenderer.GetTextSettings()

		// ローカルZ順序を計算（TextSpriteManagerのカウンターを使用）
		localZOrder := ZOrderTextBase + gs.textSpriteManager.GetNextZOffset(picID)

		// フォントフェイスを取得（TextRendererから）
		face := gs.textRenderer.GetFace()

		// 要件 9.3, 11.6: PictureSpriteを親として取得（テキストはピクチャーに配置される）
		// まずpictureSpriteMapから取得を試みる（LoadPic時に作成されたPictureSprite）
		// 見つからない場合は従来の方法（GetBackgroundPictureSpriteSprite）を使用
		var parentSprite *Sprite
		windowZOrder := 0
		if gs.pictureSpriteManager != nil {
			ps := gs.pictureSpriteManager.GetPictureSpriteByPictureID(picID)
			if ps != nil {
				parentSprite = ps.GetSprite()
			} else {
				// フォールバック: 従来の方法
				parentSprite = gs.pictureSpriteManager.GetBackgroundPictureSpriteSprite(picID)
			}
			// ウインドウのZ順序を取得
			winID, _ := gs.windows.GetWinByPicID(picID)
			if winID >= 0 {
				win, winErr := gs.windows.GetWin(winID)
				if winErr == nil && win != nil {
					windowZOrder = win.ZOrder
				}
			}
		}

		// グローバルZ順序を計算
		zOrder := CalculateGlobalZOrder(windowZOrder, localZOrder)

		// TextSpriteを作成（親スプライト付き）
		ts := gs.textSpriteManager.CreateTextSpriteWithParent(
			picID,
			x, y,
			text,
			textSettings.TextColor,
			textSettings.BgColor,
			face,
			zOrder,
			parentSprite,
		)
		if ts != nil && parentSprite != nil {
			// PictureSpriteの子として登録
			parentSprite.AddChild(ts.GetSprite())
		}
		gs.log.Debug("TextWrite: created TextSprite", "picID", picID, "text", text, "x", x, "y", y, "hasParent", parentSprite != nil, "globalZOrder", zOrder)
	}

	// スプライト構成をダンプ
	gs.dumpSpriteState(fmt.Sprintf("TextWrite(picID=%d, text=%q)", picID, text))

	return nil
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

// drawWindowSpriteDecoration はWindowSpriteを使用してウィンドウ装飾を描画する
// スプライトシステム要件 7.1〜7.3: WindowSpriteを使用した描画
func (gs *GraphicsSystem) drawWindowSpriteDecoration(screen *ebiten.Image, ws *WindowSprite, pic *Picture) {
	sprite := ws.GetSprite()
	if sprite == nil || sprite.Image() == nil {
		// スプライトが無効な場合は従来の方法で描画
		gs.drawWindowDecoration(screen, ws.GetWindow(), pic)
		return
	}

	// スプライトの画像を更新（ピクチャーが変更された場合など）
	// 注意: 毎フレーム再描画するのは非効率なので、将来的にはダーティフラグで制御する
	ws.RedrawDecoration(pic)

	// スプライトを描画
	op := &ebiten.DrawImageOptions{}
	x, y := sprite.AbsolutePosition()
	op.GeoM.Translate(x, y)

	alpha := sprite.EffectiveAlpha()
	if alpha < 1.0 {
		op.ColorScale.ScaleAlpha(float32(alpha))
	}

	screen.DrawImage(sprite.Image(), op)
}

// drawWindowDecoration はWindows 3.1風のウィンドウ装飾を描画する
// _old_implementation2/pkg/engine/renderer.goを参考に実装
// 要件 3.11: ウィンドウをZ順序で管理し、後から開いたウィンドウを前面に表示する
// 要件 3.12: ウィンドウの背景色（color引数）を適用する
// 要件 11.2: ピクチャー画像の直接描画は廃止（スプライトシステムで描画）
// 注意: この関数はWindowSpriteが存在しない場合のフォールバックとして使用される
// ピクチャー画像はOpenWin時にPictureSpriteとして作成されるため、ここでは描画しない
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

	// 5. コンテンツ領域の背景色を描画（要件 3.12）
	contentX := win.X + borderThickness
	contentY := win.Y + borderThickness + titleBarHeight

	if win.BgColor != nil {
		vector.DrawFilledRect(screen,
			float32(contentX), float32(contentY),
			float32(winWidth), float32(winHeight),
			win.BgColor, false)
	}

	// 要件 11.2: ピクチャー画像はスプライトシステムで描画する
	// OpenWin時にPictureSpriteとして作成されているため、ここでは描画しない
	// drawLayersForWindow()でPictureSpriteが描画される
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

// DrawWithSpriteManager はSpriteManager.Draw()を使用して描画する
// スプライトシステム要件 14.1: SpriteManager.Draw()ベースの描画
// 注意: この関数は将来の完全移行のための準備として実装されている
// 現在は、すべてのスプライトの親子関係が適切に設定されていないため、
// 直接使用すると描画位置がずれる可能性がある
// 完全移行後は、Draw()メソッドからこの関数を呼び出すようになる
func (gs *GraphicsSystem) DrawWithSpriteManager(screen *ebiten.Image) {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	if gs.spriteManager == nil {
		return
	}

	// SpriteManagerに登録されているすべてのスプライトをZ順序で描画
	gs.spriteManager.Draw(screen)
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
