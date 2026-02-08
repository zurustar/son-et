// graphics_draw.go は描画ロジックを提供する
// Draw()メイン描画メソッド、描画ヘルパー、デバッグオーバーレイ、
// 座標変換、テキスト・図形描画、転送メソッドを含む
package graphics

import (
	"fmt"
	"image"
	"image/color"
	"log/slog"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Draw は画面に描画する
// Ebitengineのメインスレッドで実行される
// スプライトシステム要件 14.1: SpriteManager.Draw()ベースの描画
func (gs *GraphicsSystem) Draw(screen *ebiten.Image) {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	// スプライトシステム要件 14.1: SpriteManager.Draw()ベースの描画
	// すべてのスプライトをZ_Path順で描画する
	if gs.spriteManager != nil {
		gs.spriteManager.Draw(screen)
	}
}

// drawCastsForWindow はウィンドウに属するキャストを描画する
// 要件 4.8: キャストを透明色（黒 0x000000）を除いて描画する
// 要件 4.9: キャストをZ順序で管理し、後から配置したキャストを前面に表示する
// 要件 4.10: キャストの位置をウィンドウ相対座標で管理する
func (gs *GraphicsSystem) drawCastsForWindow(screen *ebiten.Image, win *Window) {
	// CoordinateConverterを使用して座標変換
	cc := GetDefaultCoordinateConverter()

	// コンテンツ領域の開始位置を計算
	contentX, contentY := cc.WindowToContent(win.X, win.Y)

	// ピクチャーオフセットを取得
	picOffsetX, picOffsetY := GetPicOffset(win)

	// このウィンドウに属するキャストを取得（Z順序でソート済み）
	casts := gs.casts.GetCastsByWindow(win.ID)

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
		screenX, screenY := cc.PictureToScreen(cast.X, cast.Y, contentX, contentY, picOffsetX, picOffsetY)

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
func (gs *GraphicsSystem) drawLayersForWindow(screen *ebiten.Image, win *Window) {
	// CoordinateConverterを使用して座標変換
	cc := GetDefaultCoordinateConverter()

	// コンテンツ領域の開始位置を計算
	contentX, contentY := cc.WindowToContent(win.X, win.Y)

	// ピクチャーオフセットを取得
	picOffsetX, picOffsetY := GetPicOffset(win)

	// すべてのスプライトを収集してZ順序でソート
	allSprites := gs.collectAllSpritesForWindow(win)

	// Z順序でソート
	sortSpritesByZOrder(allSprites)

	// 描画
	for _, item := range allSprites {
		gs.drawSpriteItem(screen, item, contentX, contentY, picOffsetX, picOffsetY)
	}
}

// drawSpriteItem はスプライトアイテムを描画する
func (gs *GraphicsSystem) drawSpriteItem(screen *ebiten.Image, item spriteItem, contentX, contentY, picOffsetX, picOffsetY int) {
	sprite := item.sprite
	if sprite == nil || sprite.Image() == nil {
		return
	}

	// 可視性チェック
	if !sprite.IsEffectivelyVisible() {
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
	cc := GetDefaultCoordinateConverter()
	x, y := sprite.Position()
	screenX, screenY := cc.PictureToScreenFloat(x, y, contentX, contentY, picOffsetX, picOffsetY)

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
func (gs *GraphicsSystem) drawTextSpriteWithBackground(screen *ebiten.Image, ts *TextSprite, screenX, screenY int) {
	sprite := ts.GetSprite()
	if sprite == nil || sprite.Image() == nil {
		return
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(screenX), float64(screenY))

	alpha := sprite.EffectiveAlpha()
	if alpha < 1.0 {
		op.ColorScale.ScaleAlpha(float32(alpha))
	}

	screen.DrawImage(sprite.Image(), op)
}

// drawCastSpritesForWindow はウィンドウに属するすべてのCastSpriteを描画する
func (gs *GraphicsSystem) drawCastSpritesForWindow(screen *ebiten.Image, win *Window, contentX, contentY, picOffsetX, picOffsetY int) {
	if gs.castSpriteManager == nil {
		return
	}

	castSprites := gs.castSpriteManager.GetCastSpritesByWindow(win.ID)
	if len(castSprites) == 0 {
		return
	}

	sortCastSpritesByZPath(castSprites)

	for _, cs := range castSprites {
		gs.drawCastSpriteOnScreen(screen, cs, contentX, contentY, picOffsetX, picOffsetY)
	}
}

// drawCastSpriteOnScreen はCastSpriteをスクリーンに描画する
func (gs *GraphicsSystem) drawCastSpriteOnScreen(screen *ebiten.Image, cs *CastSprite, contentX, contentY, picOffsetX, picOffsetY int) {
	sprite := cs.GetSprite()
	if sprite == nil || sprite.Image() == nil {
		return
	}

	if !sprite.IsEffectivelyVisible() {
		return
	}

	cc := GetDefaultCoordinateConverter()
	x, y := sprite.Position()
	screenX, screenY := cc.PictureToScreenFloat(x, y, contentX, contentY, picOffsetX, picOffsetY)

	if cs.HasTransColor() {
		gs.drawCastWithTransparency(screen, sprite.Image(), int(screenX), int(screenY), cs.GetTransColor(), true)
		return
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(screenX, screenY)

	alpha := sprite.EffectiveAlpha()
	if alpha < 1.0 {
		op.ColorScale.ScaleAlpha(float32(alpha))
	}

	screen.DrawImage(sprite.Image(), op)
}

// drawTextSpriteOnScreen はTextSpriteをスクリーンに描画する
func (gs *GraphicsSystem) drawTextSpriteOnScreen(screen *ebiten.Image, ts *TextSprite, contentX, contentY, picOffsetX, picOffsetY int) {
	sprite := ts.GetSprite()
	if sprite == nil || sprite.Image() == nil {
		return
	}

	if !sprite.IsEffectivelyVisible() {
		return
	}

	cc := GetDefaultCoordinateConverter()
	x, y := sprite.Position()
	screenX, screenY := cc.PictureToScreenFloat(x, y, contentX, contentY, picOffsetX, picOffsetY)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(screenX, screenY)

	alpha := sprite.EffectiveAlpha()
	if alpha < 1.0 {
		op.ColorScale.ScaleAlpha(float32(alpha))
	}

	screen.DrawImage(sprite.Image(), op)
}

// drawTextSpritesForWindow はウィンドウに属するすべてのTextSpriteを描画する
func (gs *GraphicsSystem) drawTextSpritesForWindow(screen *ebiten.Image, win *Window, contentX, contentY, picOffsetX, picOffsetY int) {
	if gs.textSpriteManager == nil {
		return
	}

	picID := win.PicID

	textSprites := gs.textSpriteManager.GetTextSprites(picID)
	if len(textSprites) == 0 {
		return
	}

	sortTextSpritesByZPath(textSprites)

	for _, ts := range textSprites {
		gs.drawTextSpriteOnScreen(screen, ts, contentX, contentY, picOffsetX, picOffsetY)
	}
}

// drawDebugOverlayForWindow はウィンドウのデバッグオーバーレイを描画する
// 要件 15.1-15.8: デバッグオーバーレイの実装
func (gs *GraphicsSystem) drawDebugOverlayForWindow(screen *ebiten.Image, win *Window, pic *Picture) {
	if gs.debugOverlay == nil || !gs.debugOverlay.IsEnabled() {
		return
	}

	// CoordinateConverterを使用して座標変換
	cc := GetDefaultCoordinateConverter()

	// ウィンドウの実際のサイズを計算
	winWidth := pic.Width
	if win.Width > 0 {
		winWidth = win.Width
	}

	// タイトルバーの位置とサイズ
	titleBarX := win.X + BorderThickness
	titleBarY := win.Y + BorderThickness
	titleBarWidth := winWidth

	// ウィンドウIDをタイトルバーに描画（要件 15.1）
	gs.debugOverlay.DrawWindowID(screen, win, titleBarX, titleBarY, titleBarWidth)

	// コンテンツ領域の開始位置を計算
	contentX, contentY := cc.WindowToContent(win.X, win.Y)

	// ピクチャーオフセットを取得
	picOffsetX, picOffsetY := GetPicOffset(win)

	// ピクチャーIDをピクチャーの左上に描画（要件 15.2）
	picScreenX, picScreenY := cc.PictureToScreen(0, 0, contentX, contentY, picOffsetX, picOffsetY)
	gs.debugOverlay.DrawPictureID(screen, win.PicID, picScreenX+2, picScreenY+2)

	// このウィンドウに属するキャストのデバッグ情報を描画（要件 15.3）
	casts := gs.casts.GetCastsByWindow(win.ID)
	for _, cast := range casts {
		if !cast.Visible {
			continue
		}

		castCenterX, castCenterY := cc.PictureToScreen(
			cast.X+cast.Width/2,
			cast.Y+cast.Height/2,
			contentX, contentY,
			picOffsetX, picOffsetY,
		)

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
		}
	}
}

// SetDebugOverlayEnabled はデバッグオーバーレイの有効/無効を設定する
// 要件 15.7, 15.8: ログレベルに基づいた表示/非表示の切り替え
func (gs *GraphicsSystem) SetDebugOverlayEnabled(enabled bool) {
	if gs.debugOverlay != nil {
		gs.debugOverlay.SetEnabled(enabled)
	}
	gs.updateDebugDrawCallback(enabled)
}

// updateDebugDrawCallback はデバッグ描画コールバックを更新する
func (gs *GraphicsSystem) updateDebugDrawCallback(enabled bool) {
	if gs.spriteManager == nil {
		return
	}
	if enabled && gs.debugOverlay != nil {
		gs.spriteManager.SetDebugDrawCallback(func(screen *ebiten.Image, s *Sprite, absX, absY float64) {
			gs.debugOverlay.DrawSpriteDebugInfo(screen, s, absX, absY)
		})
	} else {
		gs.spriteManager.SetDebugDrawCallback(nil)
	}
}

// SetDebugOverlayFromLogLevel はログレベルに基づいてデバッグオーバーレイの有効/無効を設定する
func (gs *GraphicsSystem) SetDebugOverlayFromLogLevel(level slog.Level) {
	if gs.debugOverlay != nil {
		gs.debugOverlay.SetEnabledFromLogLevel(level)
		gs.updateDebugDrawCallback(gs.debugOverlay.IsEnabled())
	}
}

// SetDebugOverlayFromLogLevelString はログレベル文字列に基づいてデバッグオーバーレイの有効/無効を設定する
func (gs *GraphicsSystem) SetDebugOverlayFromLogLevelString(level string) {
	if gs.debugOverlay != nil {
		gs.debugOverlay.SetEnabledFromLogLevelString(level)
		gs.updateDebugDrawCallback(gs.debugOverlay.IsEnabled())
	}
}

// IsDebugOverlayEnabled はデバッグオーバーレイが有効かどうかを返す
func (gs *GraphicsSystem) IsDebugOverlayEnabled() bool {
	if gs.debugOverlay != nil {
		return gs.debugOverlay.IsEnabled()
	}
	return false
}

// drawWindowSpriteDecoration はWindowSpriteを使用してウィンドウ装飾を描画する
func (gs *GraphicsSystem) drawWindowSpriteDecoration(screen *ebiten.Image, ws *WindowSprite, pic *Picture) {
	sprite := ws.GetSprite()
	if sprite == nil || sprite.Image() == nil {
		gs.drawWindowDecoration(screen, ws.GetWindow(), pic)
		return
	}

	ws.RedrawDecoration(pic)

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
// 注意: この関数はWindowSpriteが存在しない場合のフォールバックとして使用される
func (gs *GraphicsSystem) drawWindowDecoration(screen *ebiten.Image, win *Window, pic *Picture) {
	var (
		titleBarColor  = color.RGBA{0, 0, 128, 255}
		borderColor    = color.RGBA{192, 192, 192, 255}
		highlightColor = color.RGBA{255, 255, 255, 255}
		shadowColor    = color.RGBA{0, 0, 0, 255}
	)

	winWidth := pic.Width
	winHeight := pic.Height
	if win.Width > 0 {
		winWidth = win.Width
	}
	if win.Height > 0 {
		winHeight = win.Height
	}

	winX := float32(win.X)
	winY := float32(win.Y)
	winW := float32(winWidth)
	winH := float32(winHeight)
	totalW := winW + float32(BorderThickness*2)
	totalH := winH + float32(BorderThickness*2) + float32(TitleBarHeight)

	// 1. ウィンドウフレームの背景を描画（グレー）
	vector.FillRect(screen, winX, winY, totalW, totalH, borderColor, false)

	// 2. 3D枠線効果を描画
	vector.StrokeLine(screen, winX, winY, winX+totalW, winY, 1, highlightColor, false)
	vector.StrokeLine(screen, winX, winY, winX, winY+totalH, 1, highlightColor, false)
	vector.StrokeLine(screen, winX, winY+totalH, winX+totalW, winY+totalH, 1, shadowColor, false)
	vector.StrokeLine(screen, winX+totalW, winY, winX+totalW, winY+totalH, 1, shadowColor, false)

	// 3. タイトルバーを描画（濃い青）
	vector.FillRect(screen,
		winX+float32(BorderThickness),
		winY+float32(BorderThickness),
		winW, float32(TitleBarHeight),
		titleBarColor, false)

	// 4. キャプションテキストの描画はWindowSprite経由（drawWindowDecorationOnImage）で実装済み
	// このフォールバックパスでは未対応

	// 5. コンテンツ領域の背景色を描画（要件 3.12）
	cc := GetDefaultCoordinateConverter()
	contentX, contentY := cc.WindowToContent(win.X, win.Y)

	if win.BgColor != nil {
		vector.FillRect(screen,
			float32(contentX), float32(contentY),
			float32(winWidth), float32(winHeight),
			win.BgColor, false)
	}
}

// VirtualToScreen は仮想デスクトップ座標をスクリーン座標に変換する
func (gs *GraphicsSystem) VirtualToScreen(vx, vy int, screenW, screenH int) (int, int) {
	scaleX := float64(screenW) / float64(gs.virtualWidth)
	scaleY := float64(screenH) / float64(gs.virtualHeight)
	scale := min(scaleX, scaleY)

	offsetX := (float64(screenW) - float64(gs.virtualWidth)*scale) / 2
	offsetY := (float64(screenH) - float64(gs.virtualHeight)*scale) / 2

	return int(float64(vx)*scale + offsetX), int(float64(vy)*scale + offsetY)
}

// ScreenToVirtual はスクリーン座標を仮想デスクトップ座標に変換する
func (gs *GraphicsSystem) ScreenToVirtual(sx, sy int, screenW, screenH int) (int, int) {
	scaleX := float64(screenW) / float64(gs.virtualWidth)
	scaleY := float64(screenH) / float64(gs.virtualHeight)
	scale := min(scaleX, scaleY)

	offsetX := (float64(screenW) - float64(gs.virtualWidth)*scale) / 2
	offsetY := (float64(screenH) - float64(gs.virtualHeight)*scale) / 2

	vx := int((float64(sx) - offsetX) / scale)
	vy := int((float64(sy) - offsetY) / scale)

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
func (gs *GraphicsSystem) GetScaleAndOffset(screenW, screenH int) (scale, offsetX, offsetY float64) {
	scaleX := float64(screenW) / float64(gs.virtualWidth)
	scaleY := float64(screenH) / float64(gs.virtualHeight)
	scale = min(scaleX, scaleY)

	offsetX = (float64(screenW) - float64(gs.virtualWidth)*scale) / 2
	offsetY = (float64(screenH) - float64(gs.virtualHeight)*scale) / 2

	return scale, offsetX, offsetY
}

// DrawWithSpriteManager はSpriteManager.Draw()を使用して描画する
func (gs *GraphicsSystem) DrawWithSpriteManager(screen *ebiten.Image) {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	if gs.spriteManager == nil {
		return
	}

	gs.spriteManager.Draw(screen)
}

// DrawScaled は仮想デスクトップをスケーリングして描画する
func (gs *GraphicsSystem) DrawScaled(screen *ebiten.Image, virtualScreen *ebiten.Image) {
	screenW := screen.Bounds().Dx()
	screenH := screen.Bounds().Dy()

	scale, offsetX, offsetY := gs.GetScaleAndOffset(screenW, screenH)

	screen.Fill(color.Black)

	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Scale(scale, scale)
	opts.GeoM.Translate(offsetX, offsetY)
	opts.Filter = ebiten.FilterLinear

	screen.DrawImage(virtualScreen, opts)
}

// ============================================================================
// テキスト描画メソッド
// ============================================================================

// TextWrite writes text to a picture
func (gs *GraphicsSystem) TextWrite(picID, x, y int, text string) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	pic, err := gs.pictures.GetPicWithoutLock(picID)
	if err != nil {
		gs.log.Error("TextWrite: picture not found", "picID", picID, "error", err)
		return err
	}

	if y >= pic.Height || y < -100 || x >= pic.Width || x < -1000 {
		gs.log.Debug("TextWrite: text position out of bounds, skipping",
			"picID", picID, "x", x, "y", y, "picWidth", pic.Width, "picHeight", pic.Height)
		return nil
	}

	if gs.textSpriteManager != nil {
		textSettings := gs.textRenderer.GetTextSettings()
		face := gs.textRenderer.GetFace()

		var parentSprite *Sprite
		if gs.pictureSpriteManager != nil {
			ps := gs.pictureSpriteManager.GetPictureSpriteByPictureID(picID)
			if ps != nil {
				parentSprite = ps.GetSprite()
			} else {
				parentSprite = gs.pictureSpriteManager.GetBackgroundPictureSpriteSprite(picID)
			}
		}

		zOrder := 0

		ts := gs.textSpriteManager.CreateTextSpriteWithParent(
			picID,
			x, y,
			text,
			textSettings.TextColor,
			textSettings.BgColor,
			face,
			zOrder,
			parentSprite,
			textSettings.BackMode,
		)
		if ts != nil && parentSprite != nil {
			parentSprite.AddChild(ts.GetSprite())
		}
		gs.log.Debug("TextWrite: created TextSprite", "picID", picID, "text", text, "x", x, "y", y, "hasParent", parentSprite != nil)
	} else {
		if err := gs.textRenderer.TextWrite(pic, x, y, text); err != nil {
			return err
		}

		if gs.pictureSpriteManager != nil {
			gs.pictureSpriteManager.UpdatePictureSpriteImage(picID, pic.Image)
		}
	}

	gs.dumpSpriteState(fmt.Sprintf("TextWrite(picID=%d, text=%q)", picID, text))

	return nil
}

// SetFont sets the font
func (gs *GraphicsSystem) SetFont(name string, size int, opts ...any) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	fontOpts := make([]FontOption, 0)

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

// ============================================================================
// 図形描画メソッド
// ============================================================================

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

// ============================================================================
// ピクチャー転送メソッド
// ============================================================================

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
