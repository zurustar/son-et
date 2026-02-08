// graphics_window.go はウィンドウ管理メソッドを提供する
// OpenWin, CloseWin, MoveWin, CloseWinAll、
// ウィンドウオプション解析、ウィンドウ情報取得、タイトル設定を含む
package graphics

import (
	"fmt"
)

// OpenWin opens a window
// スプライトシステム要件 7.1: ウィンドウが開かれたときにWindowSpriteを作成する
// 要件 3.1.1: 位置指定なしの場合はセンタリングする
func (gs *GraphicsSystem) OpenWin(picID int, opts ...any) (int, error) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	// Convert any options to WinOption
	winOpts, hasPosition := gs.parseWinOptions(opts)

	// ウィンドウを開く
	winID, err := gs.windows.OpenWin(picID, winOpts...)
	if err != nil {
		return -1, err
	}

	// ウィンドウの情報を取得
	win, err := gs.windows.GetWin(winID)
	if err != nil {
		gs.log.Error("OpenWin: failed to get window after creation", "winID", winID, "error", err)
		return winID, nil
	}

	// ウィンドウのサイズを取得（設定されていない場合はピクチャーのサイズを使用）
	width := win.Width
	height := win.Height
	var pic *Picture
	if width <= 0 || height <= 0 {
		pic, err = gs.pictures.GetPicWithoutLock(picID)
		if err == nil {
			width = pic.Width
			height = pic.Height
		} else {
			width = 640
			height = 480
		}
	} else {
		pic, _ = gs.pictures.GetPicWithoutLock(picID)
	}

	// 要件 3.1.1: 位置指定がない場合はセンタリングする
	if !hasPosition {
		totalWidth := width + BorderThickness*2
		totalHeight := height + BorderThickness*2 + TitleBarHeight

		centerX := (gs.virtualWidth - totalWidth) / 2
		centerY := (gs.virtualHeight - totalHeight) / 2

		win.X = centerX
		win.Y = centerY

		gs.log.Debug("OpenWin: centering window",
			"winID", winID,
			"virtualSize", fmt.Sprintf("%dx%d", gs.virtualWidth, gs.virtualHeight),
			"windowSize", fmt.Sprintf("%dx%d", totalWidth, totalHeight),
			"centerPos", fmt.Sprintf("(%d, %d)", centerX, centerY))
	}

	// スプライトシステム要件 7.1: WindowSpriteを作成する
	var ws *WindowSprite
	if pic != nil && gs.windowSpriteManager != nil {
		ws = gs.windowSpriteManager.CreateWindowSprite(win, pic)
		gs.log.Debug("OpenWin: created WindowSprite", "winID", winID)
	}

	// 要件 11.3, 11.4: PictureSpriteをウインドウに関連付ける
	if pic != nil && pic.Image != nil && gs.pictureSpriteManager != nil && ws != nil {
		existingPs := gs.pictureSpriteManager.GetPictureSpriteByPictureID(picID)
		if existingPs != nil {
			gs.pictureSpriteManager.AttachPictureSpriteToWindow(picID, ws.GetSprite(), winID)
			gs.log.Debug("OpenWin: attached existing PictureSprite to WindowSprite",
				"winID", winID, "picID", picID,
				"zPath", existingPs.GetSprite().GetZPath())
		} else {
			destX := 0
			destY := 0

			ps := gs.pictureSpriteManager.CreateBackgroundPictureSprite(
				pic.Image,
				picID,
				pic.Width, pic.Height,
				destX, destY,
				0,
			)

			if ps != nil {
				ps.GetSprite().SetParent(ws.GetSprite())
				ws.AddChild(ps.GetSprite())

				if ws.GetSprite().GetZPath() != nil {
					localZOrder := gs.spriteManager.GetZOrderCounter().GetNext(ws.GetSprite().ID())
					zPath := NewZPathFromParent(ws.GetSprite().GetZPath(), localZOrder)
					ps.GetSprite().SetZPath(zPath)

					ps.GetSprite().SetVisible(true)

					gs.spriteManager.MarkNeedSort()
				}

				gs.log.Debug("OpenWin: created new background PictureSprite as child of WindowSprite",
					"winID", winID, "picID", picID, "destX", destX, "destY", destY,
					"zPath", ps.GetSprite().GetZPath())
			}
		}
	}

	gs.log.Debug("OpenWin: window opened", "winID", winID, "width", width, "height", height)

	gs.dumpSpriteState(fmt.Sprintf("OpenWin(winID=%d, picID=%d)", winID, picID))

	return winID, nil
}

// parseWinOptions converts any slice to WinOption slice
// Returns: WinOption slice and whether position was explicitly specified
func (gs *GraphicsSystem) parseWinOptions(opts []any) ([]WinOption, bool) {
	winOpts := make([]WinOption, 0)
	hasPosition := false

	// OpenWin(pic, x, y, width, height, pic_x, pic_y, color)
	if len(opts) >= 2 {
		if x, ok := toIntFromAny(opts[0]); ok {
			if y, ok := toIntFromAny(opts[1]); ok {
				winOpts = append(winOpts, WithPosition(x, y))
				hasPosition = true
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

	return winOpts, hasPosition
}

// MoveWin moves or modifies a window
func (gs *GraphicsSystem) MoveWin(id int, opts ...any) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	winOpts := make([]WinOption, 0)

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

				if gs.windowSpriteManager != nil {
					ws := gs.windowSpriteManager.GetWindowSprite(id)
					if ws != nil {
						ws.UpdatePicOffset(picX, picY)
					}
				}
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

	ws := gs.windowSpriteManager.GetWindowSprite(winID)
	if ws == nil {
		gs.log.Debug("MoveWin: WindowSprite not found", "winID", winID)
		return
	}

	newPic, err := gs.pictures.GetPicWithoutLock(newPicID)
	if err != nil {
		gs.log.Warn("MoveWin: new picture not found", "picID", newPicID, "error", err)
		return
	}

	windowSprite := ws.GetSprite()
	if windowSprite == nil {
		return
	}

	// 現在表示中のすべての子PictureSpriteを非表示にする
	children := windowSprite.GetChildren()
	for _, child := range children {
		if child != nil {
			child.SetVisible(false)
		}
	}

	// 方法1: pictureSpriteMapから未関連付けのPictureSpriteを取得してAttach
	existingPs := gs.pictureSpriteManager.GetPictureSpriteByPictureID(newPicID)
	if existingPs != nil {
		gs.pictureSpriteManager.AttachPictureSpriteToWindow(newPicID, windowSprite, winID)
		gs.log.Debug("MoveWin: attached existing PictureSprite to WindowSprite",
			"winID", winID, "picID", newPicID,
			"zPath", existingPs.GetSprite().GetZPath())
	} else {
		// 方法2: 既に関連付け済みのPictureSpriteを探して表示する
		found := false
		for _, child := range children {
			if child != nil && child.Image() == newPic.Image {
				child.SetVisible(true)
				found = true
				gs.log.Debug("MoveWin: reactivated existing PictureSprite",
					"winID", winID, "picID", newPicID,
					"spriteID", child.ID())
				break
			}
		}

		if !found {
			for _, child := range children {
				if child != nil {
					child.SetImage(newPic.Image)
					child.SetVisible(true)
					gs.log.Debug("MoveWin: updated PictureSprite image",
						"winID", winID, "picID", newPicID,
						"spriteID", child.ID())
					break
				}
			}
		}
	}

	gs.dumpSpriteState(fmt.Sprintf("MoveWin(winID=%d, newPicID=%d)", winID, newPicID))
}

// CloseWin closes a window
// スプライトシステム要件 7.3: ウィンドウが閉じられたときにWindowSpriteを削除する
func (gs *GraphicsSystem) CloseWin(id int) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	if gs.castSpriteManager != nil {
		gs.castSpriteManager.RemoveCastSpritesByWindow(id)
		gs.log.Debug("CloseWin: deleted CastSprites", "winID", id)
	}

	gs.casts.DeleteCastsByWindow(id)

	if gs.windowSpriteManager != nil {
		gs.windowSpriteManager.RemoveWindowSprite(id)
		gs.log.Debug("CloseWin: deleted WindowSprite", "winID", id)
	}

	return gs.windows.CloseWin(id)
}

// CloseWinAll closes all windows
func (gs *GraphicsSystem) CloseWinAll() {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	windows := gs.windows.GetWindowsOrdered()
	for _, win := range windows {
		if gs.castSpriteManager != nil {
			gs.castSpriteManager.RemoveCastSpritesByWindow(win.ID)
		}
		gs.casts.DeleteCastsByWindow(win.ID)
	}

	if gs.windowSpriteManager != nil {
		gs.windowSpriteManager.Clear()
		gs.log.Debug("CloseWinAll: deleted all WindowSprites")
	}

	if gs.castSpriteManager != nil {
		gs.castSpriteManager.Clear()
		gs.log.Debug("CloseWinAll: deleted all CastSprites")
	}

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
