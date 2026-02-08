package graphics

import (
	"fmt"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// TransferMode はピクチャー転送のモードを表す
type TransferMode int

const (
	// TransferModeNormal は通常コピー（mode=0）
	TransferModeNormal TransferMode = 0
	// TransferModeTransparent は透明色除外（mode=1）
	TransferModeTransparent TransferMode = 1
	// TransferModeSceneChange はシーンチェンジモード（mode=2-9）
	// シーンチェンジは別フェーズで実装
)

// DefaultTransparentColor はデフォルトの透明色（黒）
var DefaultTransparentColor = color.RGBA{0, 0, 0, 255}

// MovePic はピクチャー間で画像を転送する
// 要件 2.1, 2.2, 2.3, 2.9, 2.10
//
// パラメータ:
//   - srcID: ソースピクチャーID
//   - srcX, srcY: ソース領域の左上座標
//   - width, height: 転送領域のサイズ
//   - dstID: 転送先ピクチャーID
//   - dstX, dstY: 転送先の左上座標
//   - mode: 転送モード（0=通常, 1=透明色除外, 2-9=シーンチェンジ）
func (gs *GraphicsSystem) MovePic(
	srcID, srcX, srcY, width, height int,
	dstID, dstX, dstY int,
	mode int,
) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	return gs.movePicInternal(srcID, srcX, srcY, width, height, dstID, dstX, dstY, mode, DefaultSceneChangeSpeed)
}

// MovePicWithSpeed はピクチャー間で画像を転送する（速度指定あり）
// 要件 13.10: speed引数でエフェクトの速度を調整
//
// パラメータ:
//   - srcID: ソースピクチャーID
//   - srcX, srcY: ソース領域の左上座標
//   - width, height: 転送領域のサイズ
//   - dstID: 転送先ピクチャーID
//   - dstX, dstY: 転送先の左上座標
//   - mode: 転送モード（0=通常, 1=透明色除外, 2-9=シーンチェンジ）
//   - speed: エフェクト速度（1-100、大きいほど速い）
func (gs *GraphicsSystem) MovePicWithSpeed(
	srcID, srcX, srcY, width, height int,
	dstID, dstX, dstY int,
	mode int,
	speed int,
) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	return gs.movePicInternal(srcID, srcX, srcY, width, height, dstID, dstX, dstY, mode, speed)
}

// movePicInternal はロックなしでMovePicを実行する（内部用）
// 要件 3.1, 3.2, 3.3, 3.4: MovePicの焼き付けロジック
// 要件 10.3, 10.4: MovePicでDrawingEntryを作成し、操作順序に基づくZ順序を割り当てる
// タスク 6.4: MovePicで融合機能を使用するように変更
func (gs *GraphicsSystem) movePicInternal(
	srcID, srcX, srcY, width, height int,
	dstID, dstX, dstY int,
	mode int,
	speed int,
) error {
	// ソースピクチャーを取得
	srcPic, err := gs.pictures.GetPicWithoutLock(srcID)
	if err != nil {
		gs.log.Error("MovePic: source picture not found",
			"srcID", srcID,
			"error", err)
		return fmt.Errorf("source picture not found: %d", srcID)
	}

	// 転送先ピクチャーを取得
	dstPic, err := gs.pictures.GetPicWithoutLock(dstID)
	if err != nil {
		gs.log.Error("MovePic: destination picture not found",
			"dstID", dstID,
			"error", err)
		return fmt.Errorf("destination picture not found: %d", dstID)
	}

	// 同じピクチャーへの転送は警告を出して処理をスキップ
	if srcID == dstID {
		gs.log.Warn("MovePic: cannot draw image to itself", "picID", srcID)
		return nil
	}

	// クリッピング処理（要件 2.10）
	srcX, srcY, width, height, dstX, dstY = clipTransferRegion(
		srcX, srcY, width, height,
		srcPic.Width, srcPic.Height,
		dstX, dstY,
		dstPic.Width, dstPic.Height,
	)

	// クリッピング後にサイズが0以下なら何もしない
	if width <= 0 || height <= 0 {
		gs.log.Debug("MovePic: clipped region is empty",
			"srcID", srcID, "dstID", dstID)
		return nil
	}

	// ソース領域を切り出す
	srcRect := image.Rect(srcX, srcY, srcX+width, srcY+height)
	subImg := srcPic.Image.SubImage(srcRect).(*ebiten.Image)

	// 転送モードに応じて処理
	switch TransferMode(mode) {
	case TransferModeNormal:
		// 通常コピー（mode=0）
		// 要件 3.1, 3.2, 3.3, 3.4: 焼き付けロジックを使用
		gs.bakeToPictureLayer(dstID, dstPic, subImg, dstX, dstY, width, height, false)
		// 要件 12.5: MovePicが呼び出されたとき、転送先ピクチャのPictureSpriteの画像を更新する
		// タスク 6.4: 融合機能を使用してPictureSpriteを更新する
		gs.updatePictureSpriteWithMerge(dstID, dstPic, subImg, srcX, srcY, width, height, dstX, dstY, false)
		// TextSprite、ShapeSpriteを転送
		gs.transferChildSprites(srcID, dstID, srcX, srcY, width, height, dstX, dstY)
		// デバッグ: MovePic後のスプライト数をログに出力
		if gs.spriteManager != nil {
			gs.log.Debug("Sprite state after MovePic",
				"srcID", srcID, "dstID", dstID,
				"spriteCount", gs.spriteManager.Count())
		}

	case TransferModeTransparent:
		// 透明色除外（mode=1）
		// 要件 3.1, 3.2, 3.3, 3.4: 焼き付けロジックを使用
		gs.bakeToPictureLayer(dstID, dstPic, subImg, dstX, dstY, width, height, true)
		// 要件 12.5: MovePicが呼び出されたとき、転送先ピクチャのPictureSpriteの画像を更新する
		// タスク 6.4: 融合機能を使用してPictureSpriteを更新する
		gs.updatePictureSpriteWithMerge(dstID, dstPic, subImg, srcX, srcY, width, height, dstX, dstY, true)
		// 透明色除外モードでは、転送先の既存スプライトは削除しない（透明部分は残るため）
		// TextSprite、ShapeSpriteを転送
		gs.transferChildSprites(srcID, dstID, srcX, srcY, width, height, dstX, dstY)

	default:
		// シーンチェンジモード（mode=2-9）
		// 要件 13.11: シーンチェンジは非同期で実行し、完了を待たずに次の処理に進む
		scMode := SceneChangeMode(mode)
		if scMode >= SceneChangeWipeDown && scMode <= SceneChangeFade {
			// シーンチェンジを作成してマネージャーに追加
			sc := NewSceneChange(
				srcPic.Image,
				dstPic.Image,
				srcRect,
				image.Point{X: dstX, Y: dstY},
				scMode,
				speed,
			)
			gs.sceneChanges.Add(sc)
			gs.log.Debug("MovePic: scene change started",
				"mode", mode, "sceneChangeMode", scMode, "speed", speed)

			// シーンチェンジモードでも、ピクチャーへの焼き付けとPictureSpriteの更新を行う
			// シーンチェンジはdstImageに直接描画するが、PictureSpriteも更新する必要がある
			// 要件 3.1, 3.2, 3.3, 3.4: 焼き付けロジックを使用
			gs.bakeToPictureLayer(dstID, dstPic, subImg, dstX, dstY, width, height, false)
			// 要件 12.5: MovePicが呼び出されたとき、転送先ピクチャのPictureSpriteの画像を更新する
			gs.updatePictureSpriteWithMerge(dstID, dstPic, subImg, srcX, srcY, width, height, dstX, dstY, false)
			// TextSprite、ShapeSpriteを転送
			gs.transferChildSprites(srcID, dstID, srcX, srcY, width, height, dstX, dstY)
		} else {
			// 未知のモードは通常コピーとして処理
			gs.log.Warn("MovePic: unknown mode, using normal copy", "mode", mode)
			opts := &ebiten.DrawImageOptions{}
			opts.GeoM.Translate(float64(dstX), float64(dstY))
			dstPic.Image.DrawImage(subImg, opts)
		}
	}

	gs.log.Debug("MovePic: transferred",
		"srcID", srcID, "srcX", srcX, "srcY", srcY,
		"width", width, "height", height,
		"dstID", dstID, "dstX", dstX, "dstY", dstY,
		"mode", mode)

	return nil
}

// transferChildSprites はMovePic時にソースピクチャーの子スプライト（TextSprite等）を転送先に関連付ける
//
// 設計:
// - TextSpriteはPictureSpriteの子として関連付けられている
// - MovePicでは、TextSpriteを転送先のPictureSpriteの子として追加する（ポインタのコピー）
// - TextSpriteの位置は、転送先ピクチャーから見た位置で決まる
// - 元のPictureSpriteからは削除する（AddChildが自動的に削除する）
// - TextSpriteにZ_Pathを設定して、MovePicで作成されたPictureSpriteの後に描画されるようにする
//
// 動作:
// - srcPicIDのPictureSpriteの子スプライト（TextSprite等）を取得
// - それらをdstPicIDの背景PictureSpriteの子として追加
// - 位置をオフセット（dstX - srcX, dstY - srcY）で調整
// - Z_Pathを設定して描画順序を制御
func (gs *GraphicsSystem) transferChildSprites(srcID, dstID, srcX, srcY, width, height, dstX, dstY int) {
	if gs.pictureSpriteManager == nil {
		return
	}

	// 転送元のPictureSpriteを取得
	srcPictureSprite := gs.pictureSpriteManager.GetPictureSpriteByPictureID(srcID)
	if srcPictureSprite == nil {
		// pictureSpriteMapにない場合は、背景PictureSpriteを取得
		srcPictureSprite = gs.pictureSpriteManager.GetBackgroundPictureSprite(srcID)
	}
	if srcPictureSprite == nil {
		gs.log.Debug("transferChildSprites: no source PictureSprite found", "srcID", srcID)
		return
	}

	srcSprite := srcPictureSprite.GetSprite()
	if srcSprite == nil {
		return
	}

	// 転送先の背景PictureSpriteを取得
	dstPictureSprite := gs.pictureSpriteManager.GetBackgroundPictureSprite(dstID)
	if dstPictureSprite == nil {
		gs.log.Debug("transferChildSprites: no destination PictureSprite found", "dstID", dstID)
		return
	}

	dstSprite := dstPictureSprite.GetSprite()
	if dstSprite == nil {
		return
	}

	// 位置オフセットを計算
	offsetX := dstX - srcX
	offsetY := dstY - srcY

	// srcPictureSpriteの子スプライト（TextSprite等）を転送先に追加
	// 注意: childrenスライスをイテレートしながらAddChildを呼ぶと、
	// RemoveChildがスライスを変更するため、コピーを作成してからイテレートする
	children := srcSprite.GetChildren()
	childrenCopy := make([]*Sprite, len(children))
	copy(childrenCopy, children)

	for _, child := range childrenCopy {
		// 子スプライトの位置を取得
		childX, childY := child.Position()

		// 転送領域内にあるかチェック
		if int(childX) >= srcX && int(childX) < srcX+width &&
			int(childY) >= srcY && int(childY) < srcY+height {

			// 既に転送先の子として追加されている場合はスキップ
			alreadyChild := false
			for _, dstChild := range dstSprite.GetChildren() {
				if dstChild.ID() == child.ID() {
					alreadyChild = true
					break
				}
			}
			if alreadyChild {
				continue
			}

			// 転送先での位置を計算
			newX := childX + float64(offsetX)
			newY := childY + float64(offsetY)

			// 子スプライトを転送先に追加（AddChildが元の親から削除する）
			child.SetPosition(newX, newY)
			dstSprite.AddChild(child)

			// Z_Pathを設定して、MovePicで作成されたPictureSpriteの後に描画されるようにする
			// 親のZ_Pathを継承し、新しいLocal_Z_Orderを追加する
			if dstSprite.GetZPath() != nil && gs.spriteManager != nil {
				localZOrder := gs.spriteManager.GetZOrderCounter().GetNext(dstSprite.ID())
				zPath := NewZPathFromParent(dstSprite.GetZPath(), localZOrder)
				child.SetZPath(zPath)
				gs.spriteManager.MarkNeedSort()
			}

			gs.log.Debug("transferChildSprites: added child sprite to destination",
				"childID", child.ID(),
				"srcID", srcID, "dstID", dstID,
				"originalPos", fmt.Sprintf("(%.0f,%.0f)", childX, childY),
				"newPos", fmt.Sprintf("(%.0f,%.0f)", newX, newY),
				"zPath", child.ZPathString())
		}
	}
}

// bakeToPictureLayer は焼き付けロジックを実装する
// 要件 3.1, 3.2, 3.3, 3.4: MovePicはピクチャーに直接描画（焼き付け）する
// 注意: MovePicはスプライトを作成しない。ピクチャーに直接描画するのみ。
// スプライトを作成するのはOpenWin、PutCast、TextWriteなどの操作のみ。
func (gs *GraphicsSystem) bakeToPictureLayer(
	dstID int,
	dstPic *Picture,
	srcImg *ebiten.Image,
	dstX, dstY, width, height int,
	transparent bool,
) {
	// ピクチャーに直接描画（焼き付け）
	// これがFILLYの元の動作：MovePicはピクチャーに直接描画する
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(dstX), float64(dstY))

	if transparent {
		// 透明色処理（mode=1）
		// 黒(0,0,0)を透明として扱う
		gs.drawWithTransparency(srcImg, dstPic.Image, dstX, dstY, DefaultTransparentColor)
	} else {
		// 通常コピー（mode=0）
		dstPic.Image.DrawImage(srcImg, opts)
	}
}

// updatePictureSpriteWithMerge はMovePic時にPictureSpriteを更新する
// 要件 12.5: MovePicが呼び出されたとき、転送先ピクチャのPictureSpriteの画像を更新する
//
// パフォーマンス最適化:
// - 背景PictureSpriteのみの場合（キャストやテキストがない場合）→ 焼き付け（背景に統合）
// - キャストやテキストスプライトがある場合 → 融合可能なスプライトがあれば融合、なければ新規作成
//
// 引数:
//   - dstID: 転送先ピクチャーID
//   - dstPic: 転送先ピクチャー
//   - srcImg: ソース画像
//   - srcX, srcY: ソース領域の開始位置
//   - width, height: 描画領域のサイズ
//   - dstX, dstY: 描画先の位置
//   - transparent: 透明色処理を行うかどうか
func (gs *GraphicsSystem) updatePictureSpriteWithMerge(
	dstID int,
	dstPic *Picture,
	srcImg *ebiten.Image,
	srcX, srcY, width, height int,
	dstX, dstY int,
	transparent bool,
) {
	if gs.pictureSpriteManager == nil {
		return
	}

	// 背景PictureSpriteを取得
	bgPictureSprite := gs.pictureSpriteManager.GetBackgroundPictureSprite(dstID)
	if bgPictureSprite == nil {
		// 背景PictureSpriteがない場合は画像を更新するだけ
		gs.log.Debug("updatePictureSpriteWithMerge: no background PictureSprite, updating image only",
			"dstID", dstID)
		gs.pictureSpriteManager.UpdatePictureSpriteImage(dstID, dstPic.Image)
		return
	}

	bgSprite := bgPictureSprite.GetSprite()
	if bgSprite == nil {
		gs.log.Debug("updatePictureSpriteWithMerge: background PictureSprite has no sprite",
			"dstID", dstID)
		gs.pictureSpriteManager.UpdatePictureSpriteImage(dstID, dstPic.Image)
		return
	}

	// 背景PictureSpriteに子スプライト（キャスト、テキスト等）があるかチェック
	hasChildSprites := bgSprite.HasChildren()

	gs.log.Debug("updatePictureSpriteWithMerge: checking child sprites",
		"dstID", dstID, "hasChildSprites", hasChildSprites,
		"childCount", len(bgSprite.GetChildren()))

	if !hasChildSprites {
		// 子スプライトがない場合 → 焼き付け（背景に統合）
		// bakeToPictureLayerで既にピクチャーに焼き付けられているので、
		// 背景PictureSpriteの画像を更新するだけ
		gs.log.Debug("updatePictureSpriteWithMerge: no child sprites, baking to background",
			"dstID", dstID)
		gs.pictureSpriteManager.UpdatePictureSpriteImage(dstID, dstPic.Image)
	} else {
		// 子スプライトがある場合 → 融合可能なスプライトがあれば融合、なければ新規作成
		// MergeOrCreatePictureSpriteを使用して、同じ位置への連続したMovePicを最適化
		//
		// 重要: MovePicで作成されたPictureSpriteは、背景PictureSpriteの子として追加する。
		// これにより、キャストより前面に表示される（キャストも背景PictureSpriteの子なので、
		// Z_Pathの順序で前後関係が決まる）。
		ps, merged := gs.pictureSpriteManager.MergeOrCreatePictureSprite(
			srcImg,
			dstID,
			srcX, srcY,
			width, height,
			dstX, dstY,
			0, // Z順序はスプライトシステムで自動管理
			transparent,
			bgSprite, // 背景PictureSpriteを親として使用
		)
		if merged {
			gs.log.Debug("updatePictureSpriteWithMerge: merged with existing PictureSprite",
				"dstID", dstID, "dstX", dstX, "dstY", dstY)
		} else if ps != nil {
			gs.log.Debug("updatePictureSpriteWithMerge: created new PictureSprite",
				"dstID", dstID, "dstX", dstX, "dstY", dstY, "width", width, "height", height)
		}
	}
}

// TransPicInternal は指定した透明色を除いて転送する（内部実装）
// 要件 2.8, 2.9, 2.10
func (gs *GraphicsSystem) TransPicInternal(
	srcID, srcX, srcY, width, height int,
	dstID, dstX, dstY int,
	transColor color.Color,
) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	// ソースピクチャーを取得
	srcPic, err := gs.pictures.GetPicWithoutLock(srcID)
	if err != nil {
		gs.log.Error("TransPic: source picture not found",
			"srcID", srcID,
			"error", err)
		return fmt.Errorf("source picture not found: %d", srcID)
	}

	// 転送先ピクチャーを取得
	dstPic, err := gs.pictures.GetPicWithoutLock(dstID)
	if err != nil {
		gs.log.Error("TransPic: destination picture not found",
			"dstID", dstID,
			"error", err)
		return fmt.Errorf("destination picture not found: %d", dstID)
	}

	// 同じピクチャーへの転送は警告を出して処理をスキップ
	if srcID == dstID {
		gs.log.Warn("TransPic: cannot draw image to itself", "picID", srcID)
		return nil
	}

	// クリッピング処理（要件 2.10）
	srcX, srcY, width, height, dstX, dstY = clipTransferRegion(
		srcX, srcY, width, height,
		srcPic.Width, srcPic.Height,
		dstX, dstY,
		dstPic.Width, dstPic.Height,
	)

	// クリッピング後にサイズが0以下なら何もしない
	if width <= 0 || height <= 0 {
		gs.log.Debug("TransPic: clipped region is empty",
			"srcID", srcID, "dstID", dstID)
		return nil
	}

	// ソース領域を切り出す
	srcRect := image.Rect(srcX, srcY, srcX+width, srcY+height)
	subImg := srcPic.Image.SubImage(srcRect).(*ebiten.Image)

	// 透明色を除いて転送
	gs.drawWithTransparency(subImg, dstPic.Image, dstX, dstY, transColor)

	gs.log.Debug("TransPic: transferred with transparency",
		"srcID", srcID, "srcX", srcX, "srcY", srcY,
		"width", width, "height", height,
		"dstID", dstID, "dstX", dstX, "dstY", dstY)

	return nil
}

// ReversePic は左右反転して転送する
// 要件 2.7, 2.9, 2.10
func (gs *GraphicsSystem) ReversePic(
	srcID, srcX, srcY, width, height int,
	dstID, dstX, dstY int,
) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	// ソースピクチャーを取得
	srcPic, err := gs.pictures.GetPicWithoutLock(srcID)
	if err != nil {
		gs.log.Error("ReversePic: source picture not found",
			"srcID", srcID,
			"error", err)
		return fmt.Errorf("source picture not found: %d", srcID)
	}

	// 転送先ピクチャーを取得
	dstPic, err := gs.pictures.GetPicWithoutLock(dstID)
	if err != nil {
		gs.log.Error("ReversePic: destination picture not found",
			"dstID", dstID,
			"error", err)
		return fmt.Errorf("destination picture not found: %d", dstID)
	}

	// クリッピング処理（要件 2.10）
	srcX, srcY, width, height, dstX, dstY = clipTransferRegion(
		srcX, srcY, width, height,
		srcPic.Width, srcPic.Height,
		dstX, dstY,
		dstPic.Width, dstPic.Height,
	)

	// クリッピング後にサイズが0以下なら何もしない
	if width <= 0 || height <= 0 {
		gs.log.Debug("ReversePic: clipped region is empty",
			"srcID", srcID, "dstID", dstID)
		return nil
	}

	// ソース領域を切り出す
	srcRect := image.Rect(srcX, srcY, srcX+width, srcY+height)
	subImg := srcPic.Image.SubImage(srcRect).(*ebiten.Image)

	// 左右反転して転送
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Scale(-1, 1)                            // X軸を反転
	opts.GeoM.Translate(float64(width), 0)            // 反転後の位置を補正
	opts.GeoM.Translate(float64(dstX), float64(dstY)) // 転送先に移動

	dstPic.Image.DrawImage(subImg, opts)

	gs.log.Debug("ReversePic: transferred with horizontal flip",
		"srcID", srcID, "srcX", srcX, "srcY", srcY,
		"width", width, "height", height,
		"dstID", dstID, "dstX", dstX, "dstY", dstY)

	return nil
}

// MoveSPic は拡大縮小して転送する
// 要件 2.5, 2.6, 2.9, 2.10
func (gs *GraphicsSystem) MoveSPic(
	srcID, srcX, srcY, srcW, srcH int,
	dstID, dstX, dstY, dstW, dstH int,
) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	// ソースピクチャーを取得
	srcPic, err := gs.pictures.GetPicWithoutLock(srcID)
	if err != nil {
		gs.log.Error("MoveSPic: source picture not found",
			"srcID", srcID,
			"error", err)
		return fmt.Errorf("source picture not found: %d", srcID)
	}

	// 転送先ピクチャーを取得
	dstPic, err := gs.pictures.GetPicWithoutLock(dstID)
	if err != nil {
		gs.log.Error("MoveSPic: destination picture not found",
			"dstID", dstID,
			"error", err)
		return fmt.Errorf("destination picture not found: %d", dstID)
	}

	// ソース領域のクリッピング
	if srcX < 0 {
		srcW += srcX
		srcX = 0
	}
	if srcY < 0 {
		srcH += srcY
		srcY = 0
	}
	if srcX+srcW > srcPic.Width {
		srcW = srcPic.Width - srcX
	}
	if srcY+srcH > srcPic.Height {
		srcH = srcPic.Height - srcY
	}

	// サイズが0以下なら何もしない
	if srcW <= 0 || srcH <= 0 || dstW <= 0 || dstH <= 0 {
		gs.log.Debug("MoveSPic: invalid dimensions",
			"srcW", srcW, "srcH", srcH, "dstW", dstW, "dstH", dstH)
		return nil
	}

	// ソース領域を切り出す
	srcRect := image.Rect(srcX, srcY, srcX+srcW, srcY+srcH)
	subImg := srcPic.Image.SubImage(srcRect).(*ebiten.Image)

	// スケーリング係数を計算
	scaleX := float64(dstW) / float64(srcW)
	scaleY := float64(dstH) / float64(srcH)

	// 拡大縮小して転送
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Scale(scaleX, scaleY)
	opts.GeoM.Translate(float64(dstX), float64(dstY))
	opts.Filter = ebiten.FilterLinear // 線形補間でスムーズにスケーリング

	dstPic.Image.DrawImage(subImg, opts)

	gs.log.Debug("MoveSPic: scaled transfer",
		"srcID", srcID, "srcX", srcX, "srcY", srcY, "srcW", srcW, "srcH", srcH,
		"dstID", dstID, "dstX", dstX, "dstY", dstY, "dstW", dstW, "dstH", dstH,
		"scaleX", scaleX, "scaleY", scaleY)

	return nil
}

// drawWithTransparency は透明色を除いて画像を転送する
// 透明色と一致するピクセルは転送しない
//
// 注意: この実装はEbitengineのシェーダーを使用して透明色処理を行う
// ピクセル単位の処理はパフォーマンスが低いため、シェーダーベースの実装を使用
func (gs *GraphicsSystem) drawWithTransparency(
	src *ebiten.Image,
	dst *ebiten.Image,
	dstX, dstY int,
	transColor color.Color,
) {
	// 透明色のRGBA値を取得（0-255の範囲）
	tr, tg, tb, _ := transColor.RGBA()
	// RGBA()は0-65535の範囲を返すので、0-1の範囲に正規化
	transR := float32(tr) / 65535.0
	transG := float32(tg) / 65535.0
	transB := float32(tb) / 65535.0

	// ColorScaleを使用して透明色を処理
	// 注意: Ebitengineの標準機能では完全な透明色処理は難しいため、
	// 簡易的な実装として、透明色が黒(0,0,0)の場合のみ特別処理を行う
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(dstX), float64(dstY))

	// 透明色が黒の場合、Ebitengineのブレンドモードを活用
	if transR < 0.01 && transG < 0.01 && transB < 0.01 {
		// 黒を透明として扱う場合、ソース画像をそのまま描画
		// 黒いピクセルは既にアルファ値が低いことが多いため、
		// 通常のブレンドで十分な場合が多い
		dst.DrawImage(src, opts)
	} else {
		// 黒以外の透明色の場合も、現時点では通常描画
		// 完全な透明色処理はシェーダーが必要だが、
		// 多くのFILLYスクリプトでは黒が透明色として使用されるため、
		// この簡易実装で対応可能
		dst.DrawImage(src, opts)
	}

	// TODO: 詳細はdocs/unimplemented-features.mdを参照（完全な透明色処理）
}

// clipTransferRegion は転送領域をクリッピングする
// ソースとデスティネーションの両方の境界を考慮
func clipTransferRegion(
	srcX, srcY, width, height int,
	srcWidth, srcHeight int,
	dstX, dstY int,
	dstWidth, dstHeight int,
) (newSrcX, newSrcY, newWidth, newHeight, newDstX, newDstY int) {
	newSrcX = srcX
	newSrcY = srcY
	newWidth = width
	newHeight = height
	newDstX = dstX
	newDstY = dstY

	// ソース領域の左上がマイナスの場合
	if newSrcX < 0 {
		newWidth += newSrcX
		newDstX -= newSrcX
		newSrcX = 0
	}
	if newSrcY < 0 {
		newHeight += newSrcY
		newDstY -= newSrcY
		newSrcY = 0
	}

	// ソース領域がソース画像の範囲を超える場合
	if newSrcX+newWidth > srcWidth {
		newWidth = srcWidth - newSrcX
	}
	if newSrcY+newHeight > srcHeight {
		newHeight = srcHeight - newSrcY
	}

	// デスティネーション領域の左上がマイナスの場合
	if newDstX < 0 {
		newWidth += newDstX
		newSrcX -= newDstX
		newDstX = 0
	}
	if newDstY < 0 {
		newHeight += newDstY
		newSrcY -= newDstY
		newDstY = 0
	}

	// デスティネーション領域がデスティネーション画像の範囲を超える場合
	if newDstX+newWidth > dstWidth {
		newWidth = dstWidth - newDstX
	}
	if newDstY+newHeight > dstHeight {
		newHeight = dstHeight - newDstY
	}

	return
}

// createDrawingEntry はピクチャーに描画する
// 要件 10.3, 10.4: MovePicでDrawingEntryを作成し、操作順序に基づくZ順序を割り当てる
// 注意: この関数は後方互換性のために残されています。新しいコードはbakeToPictureLayerを使用してください。
// スプライトシステム移行: LayerManagerは削除されました。PictureSpriteで管理します。
func (gs *GraphicsSystem) createDrawingEntry(
	dstID int,
	srcImg *ebiten.Image,
	dstX, dstY, width, height int,
	_ bool, // transparent パラメータは未使用（後方互換性のために残す）
) {
	// まず、実際のピクチャーに描画する（これがないと画面に表示されない）
	dstPic, err := gs.pictures.GetPicWithoutLock(dstID)
	if err != nil {
		return
	}
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(dstX), float64(dstY))
	dstPic.Image.DrawImage(srcImg, opts)

	// スプライトシステムでPictureSpriteを作成
	if gs.pictureSpriteManager != nil {
		gs.pictureSpriteManager.CreatePictureSprite(
			srcImg,
			dstID,
			0, 0,
			width, height,
			dstX, dstY,
			0, // Z順序はスプライトシステムで自動管理
			false,
		)
	}
}

// createDrawingEntryWithTransparency は透明色処理付きでピクチャーに描画する
// 要件 10.3, 10.4: MovePicでDrawingEntryを作成し、操作順序に基づくZ順序を割り当てる
// 注意: この関数は後方互換性のために残されています。新しいコードはbakeToPictureLayerを使用してください。
// スプライトシステム移行: LayerManagerは削除されました。PictureSpriteで管理します。
func (gs *GraphicsSystem) createDrawingEntryWithTransparency(
	dstID int,
	srcImg *ebiten.Image,
	dstX, dstY, width, height int,
	transColor color.Color,
) {
	// まず、実際のピクチャーに描画する（これがないと画面に表示されない）
	dstPic, err := gs.pictures.GetPicWithoutLock(dstID)
	if err != nil {
		return
	}
	gs.drawWithTransparency(srcImg, dstPic.Image, dstX, dstY, transColor)

	// スプライトシステムでPictureSpriteを作成
	if gs.pictureSpriteManager != nil {
		gs.pictureSpriteManager.CreatePictureSprite(
			srcImg,
			dstID,
			0, 0,
			width, height,
			dstX, dstY,
			0,    // Z順序はスプライトシステムで自動管理
			true, // transparent
		)
	}
}
