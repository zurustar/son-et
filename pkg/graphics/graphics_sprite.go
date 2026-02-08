// graphics_sprite.go はスプライト管理メソッドを提供する
// PutCast, MoveCast, DelCast、スプライト収集・更新メソッドを含む
package graphics

import (
	"fmt"
	"image/color"
)

// PutCast places a cast on a picture
// スプライトシステム要件 8.1: キャストをスプライトとして作成する
// 要件 9.2: ピクチャ内にキャストが配置されたとき、キャストをピクチャの子スプライトとして管理する
func (gs *GraphicsSystem) PutCast(srcPicID, dstPicID, x, y, srcX, srcY, w, h int) (int, error) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	// 配置先ピクチャーからウインドウIDを逆引き
	winID, err := gs.windows.GetWinByPicID(dstPicID)
	if err != nil {
		winID = -1
		gs.log.Debug("PutCast: no window found for dstPicID, creating unattached cast", "dstPicID", dstPicID)
	}

	castID, err := gs.casts.PutCast(winID, srcPicID, x, y, srcX, srcY, w, h)
	if err != nil {
		return -1, err
	}

	if gs.castSpriteManager != nil {
		cast, err := gs.casts.GetCast(castID)
		if err == nil && cast != nil {
			srcPic, err := gs.pictures.GetPicWithoutLock(srcPicID)
			if err == nil && srcPic != nil && srcPic.Image != nil {
				var parentSprite *Sprite
				if gs.pictureSpriteManager != nil {
					ps := gs.pictureSpriteManager.GetPictureSpriteByPictureID(dstPicID)
					if ps != nil {
						parentSprite = ps.GetSprite()
					} else {
						parentSprite = gs.pictureSpriteManager.GetBackgroundPictureSpriteSprite(dstPicID)
					}
				}

				zOrder := 0

				cs := gs.castSpriteManager.CreateCastSpriteWithParent(cast, srcPic.Image, zOrder, parentSprite)
				if cs != nil && parentSprite != nil {
					parentSprite.AddChild(cs.GetSprite())
				}
				gs.log.Debug("PutCast: created CastSprite", "castID", castID, "srcPicID", srcPicID, "dstPicID", dstPicID, "winID", winID, "hasParent", parentSprite != nil)
			}
		}
	}

	gs.dumpSpriteState(fmt.Sprintf("PutCast(castID=%d, srcPicID=%d, dstPicID=%d)", castID, srcPicID, dstPicID))

	return castID, nil
}

// PutCastWithTransColor places a cast on a picture with transparent color
// スプライトシステム要件 8.1, 8.4: キャストをスプライトとして作成し、透明色処理をサポートする
func (gs *GraphicsSystem) PutCastWithTransColor(srcPicID, dstPicID, x, y, srcX, srcY, w, h int, transColor color.Color) (int, error) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	winID, err := gs.windows.GetWinByPicID(dstPicID)
	if err != nil {
		winID = -1
		gs.log.Debug("PutCastWithTransColor: no window found for dstPicID, creating unattached cast", "dstPicID", dstPicID)
	}

	castID, err := gs.casts.PutCastWithTransColor(winID, srcPicID, x, y, srcX, srcY, w, h, transColor)
	if err != nil {
		return -1, err
	}

	if gs.castSpriteManager != nil {
		cast, err := gs.casts.GetCast(castID)
		if err == nil && cast != nil {
			srcPic, err := gs.pictures.GetPicWithoutLock(srcPicID)
			if err == nil && srcPic != nil && srcPic.Image != nil {
				var parentSprite *Sprite
				if gs.pictureSpriteManager != nil {
					ps := gs.pictureSpriteManager.GetPictureSpriteByPictureID(dstPicID)
					if ps != nil {
						parentSprite = ps.GetSprite()
					} else {
						parentSprite = gs.pictureSpriteManager.GetBackgroundPictureSpriteSprite(dstPicID)
					}
				}

				zOrder := 0

				cs := gs.castSpriteManager.CreateCastSpriteWithTransColorAndParent(cast, srcPic.Image, zOrder, transColor, parentSprite)
				if cs != nil && parentSprite != nil {
					parentSprite.AddChild(cs.GetSprite())
				}
				gs.log.Debug("PutCastWithTransColor: created CastSprite", "castID", castID, "srcPicID", srcPicID, "dstPicID", dstPicID, "winID", winID, "hasParent", parentSprite != nil)
			}
		}
	}

	gs.dumpSpriteState(fmt.Sprintf("PutCastWithTransColor(castID=%d, srcPicID=%d, dstPicID=%d)", castID, srcPicID, dstPicID))

	return castID, nil
}

// MoveCast moves a cast
// スプライトシステム要件 8.2: キャストの位置を移動できる（残像なし）
func (gs *GraphicsSystem) MoveCast(id int, opts ...any) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	castOpts := make([]CastOption, 0)

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

	if err := gs.casts.MoveCast(id, castOpts...); err != nil {
		return err
	}

	gs.updateCastSprite(id)

	return nil
}

// MoveCastWithOptions moves a cast with explicit options
func (gs *GraphicsSystem) MoveCastWithOptions(id int, opts ...CastOption) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	if err := gs.casts.MoveCast(id, opts...); err != nil {
		return err
	}

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

	cs.UpdatePosition(cast.X, cast.Y)

	cs.UpdateSource(cast.SrcX, cast.SrcY, cast.Width, cast.Height)

	if cs.GetSrcPicID() != cast.PicID {
		cs.UpdatePicID(cast.PicID)
	}

	if cs.IsDirty() {
		srcPic, err := gs.pictures.GetPicWithoutLock(cast.PicID)
		if err == nil && srcPic != nil && srcPic.Image != nil {
			cs.RebuildCache(srcPic.Image)
		}
	}

	cs.UpdateVisible(cast.Visible)
}

// DelCast deletes a cast
// スプライトシステム要件 8.3: キャストを削除できる
func (gs *GraphicsSystem) DelCast(id int) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

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
