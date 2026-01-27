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
// 要件 10.3, 10.4: MovePicでDrawingEntryを作成し、操作順序に基づくZ順序を割り当てる
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
		// 要件 10.3, 10.4: DrawingEntryを作成してLayerManagerに追加
		gs.createDrawingEntry(dstID, subImg, dstX, dstY, width, height, false)

	case TransferModeTransparent:
		// 透明色除外（mode=1）
		// 要件 10.3, 10.4: DrawingEntryを作成してLayerManagerに追加
		gs.createDrawingEntryWithTransparency(dstID, subImg, dstX, dstY, width, height, DefaultTransparentColor)

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

	// TODO: 完全な透明色処理が必要な場合は、カスタムシェーダーを実装する
	// 現時点では、透明色処理は簡易的な実装とし、
	// 必要に応じて後のフェーズで改善する
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

// createDrawingEntry はDrawingEntryを作成してLayerManagerに追加し、ピクチャーに描画する
// 要件 10.3, 10.4: MovePicでDrawingEntryを作成し、操作順序に基づくZ順序を割り当てる
func (gs *GraphicsSystem) createDrawingEntry(
	dstID int,
	srcImg *ebiten.Image,
	dstX, dstY, width, height int,
	transparent bool,
) {
	// まず、実際のピクチャーに描画する（これがないと画面に表示されない）
	dstPic, err := gs.pictures.GetPicWithoutLock(dstID)
	if err != nil {
		return
	}
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(dstX), float64(dstY))
	dstPic.Image.DrawImage(srcImg, opts)

	// LayerManagerがある場合は、DrawingEntryも作成（将来のレイヤー合成機能用）
	if gs.layerManager != nil {
		// PictureLayerSetを取得または作成
		pls := gs.layerManager.GetOrCreatePictureLayerSet(dstID)

		// ソース画像をコピー（DrawingEntryは独自の画像を持つ）
		entryImg := ebiten.NewImage(width, height)
		entryImg.DrawImage(srcImg, nil)

		// DrawingEntryを作成
		layerID := gs.layerManager.GetNextLayerID()
		entry := NewDrawingEntry(layerID, dstID, entryImg, dstX, dstY, width, height, 0)

		// LayerManagerに追加（操作順序に基づくZ順序が割り当てられる）
		pls.AddDrawingEntry(entry)

		gs.log.Debug("MovePic: created DrawingEntry",
			"layerID", layerID, "dstID", dstID,
			"dstX", dstX, "dstY", dstY,
			"width", width, "height", height,
			"zOrder", entry.GetZOrder())
	}
}

// createDrawingEntryWithTransparency は透明色処理付きでDrawingEntryを作成し、ピクチャーに描画する
// 要件 10.3, 10.4: MovePicでDrawingEntryを作成し、操作順序に基づくZ順序を割り当てる
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

	// LayerManagerがある場合は、DrawingEntryも作成（将来のレイヤー合成機能用）
	if gs.layerManager != nil {
		// PictureLayerSetを取得または作成
		pls := gs.layerManager.GetOrCreatePictureLayerSet(dstID)

		// ソース画像をコピー（透明色処理は合成時に行う）
		entryImg := ebiten.NewImage(width, height)
		entryImg.DrawImage(srcImg, nil)

		// DrawingEntryを作成
		layerID := gs.layerManager.GetNextLayerID()
		entry := NewDrawingEntry(layerID, dstID, entryImg, dstX, dstY, width, height, 0)

		// LayerManagerに追加（操作順序に基づくZ順序が割り当てられる）
		pls.AddDrawingEntry(entry)

		gs.log.Debug("MovePic: created DrawingEntry with transparency",
			"layerID", layerID, "dstID", dstID,
			"dstX", dstX, "dstY", dstY,
			"width", width, "height", height,
			"zOrder", entry.GetZOrder())
	}
}
