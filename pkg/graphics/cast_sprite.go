// Package graphics provides sprite-based rendering system.
package graphics

import (
	"image"
	"image/color"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

// CastSprite はキャストとスプライトを組み合わせたラッパー構造体
// 要件 8.1: キャストをスプライトとして作成できる
// 要件 8.2: キャストの位置を移動できる（残像なし）
// 要件 8.3: キャストを削除できる
// 要件 8.4: 透明色処理をサポートする
type CastSprite struct {
	cast   *Cast   // 元のキャスト情報
	sprite *Sprite // スプライト（キャストの画像を保持）

	// ソース画像情報
	srcPicID int           // ソースピクチャーID
	srcImage *ebiten.Image // ソース画像への参照

	// 前回のソース領域（変更検出用）
	lastSrcX   int
	lastSrcY   int
	lastWidth  int
	lastHeight int

	// 透明色処理
	transColor    color.Color // 透明色
	hasTransColor bool        // 透明色が設定されているか

	// キャッシュ
	cachedImage *ebiten.Image // 透明色処理済みのキャッシュ画像
	dirty       bool          // キャッシュが無効かどうか

	mu sync.RWMutex
}

// CastSpriteManager はCastSpriteを管理する
type CastSpriteManager struct {
	castSprites   map[int]*CastSprite // castID -> CastSprite
	spriteManager *SpriteManager
	mu            sync.RWMutex
}

// NewCastSpriteManager は新しいCastSpriteManagerを作成する
func NewCastSpriteManager(sm *SpriteManager) *CastSpriteManager {
	return &CastSpriteManager{
		castSprites:   make(map[int]*CastSprite),
		spriteManager: sm,
	}
}

// CreateCastSprite はキャストからCastSpriteを作成する
// 要件 8.1: キャストをスプライトとして作成できる
func (csm *CastSpriteManager) CreateCastSprite(
	cast *Cast,
	srcImage *ebiten.Image,
	zOrder int,
) *CastSprite {
	csm.mu.Lock()
	defer csm.mu.Unlock()

	if cast == nil {
		return nil
	}

	// キャストのソース領域を切り出す
	var img *ebiten.Image
	if srcImage != nil && cast.Width > 0 && cast.Height > 0 {
		img = csm.extractCastImage(srcImage, cast)
	}

	// スプライトを作成（非表示状態で作成）
	// 注意: zOrderパラメータは互換性のために残されているが、
	// 実際のZ順序はZ_Pathで管理される
	// レースコンディション対策: CreateSpriteHiddenを使用して最初から非表示で作成
	// Z_Pathを設定した後にSetVisible(true)を呼ぶ必要がある
	sprite := csm.spriteManager.CreateSpriteHidden(img)
	sprite.SetPosition(float64(cast.X), float64(cast.Y))
	// 注意: visibleはZ_Path設定後に設定される（CreateCastSpriteWithParentで）

	cs := &CastSprite{
		cast:          cast,
		sprite:        sprite,
		srcPicID:      cast.PicID,
		srcImage:      srcImage,
		lastSrcX:      cast.SrcX,
		lastSrcY:      cast.SrcY,
		lastWidth:     cast.Width,
		lastHeight:    cast.Height,
		transColor:    cast.TransColor,
		hasTransColor: cast.HasTransColor,
		cachedImage:   img,
		dirty:         false,
	}

	csm.castSprites[cast.ID] = cs
	return cs
}

// CreateCastSpriteWithParent はキャストからCastSpriteを作成し、親スプライトを設定する
// 要件 8.1: キャストをスプライトとして作成できる
// 要件 14.2: ウインドウ内のスプライトをウインドウの子スプライトとして管理する
// 要件 2.2: PutCastが呼び出されたとき、現在のZ_Order_Counterを使用してLocal_Z_Orderを割り当てる
// 要件 1.4: 子スプライトが作成されたとき、親のZ_Pathを継承し、自身のLocal_Z_Orderを追加する
func (csm *CastSpriteManager) CreateCastSpriteWithParent(
	cast *Cast,
	srcImage *ebiten.Image,
	zOrder int,
	parent *Sprite,
) *CastSprite {
	cs := csm.CreateCastSprite(cast, srcImage, zOrder)
	if cs != nil && parent != nil {
		cs.SetParent(parent)

		// 要件 2.2, 2.6: 操作順序でLocal_Z_Orderを割り当てる
		// 要件 1.4: 親のZ_Pathを継承し、自身のLocal_Z_Orderを追加する
		if parent.GetZPath() != nil {
			// 親のIDを使用してZOrderCounterから次のLocal_Z_Orderを取得
			localZOrder := csm.spriteManager.GetZOrderCounter().GetNext(parent.ID())
			zPath := NewZPathFromParent(parent.GetZPath(), localZOrder)
			cs.sprite.SetZPath(zPath)

			// レースコンディション対策: Z_Pathを設定した後にSetVisibleを呼ぶ
			// これにより、Draw()がスナップショットを取る際に一貫した状態を保証する
			if cast != nil && cast.Visible {
				cs.sprite.SetVisible(true)
			}

			csm.spriteManager.MarkNeedSort()
		}
	}
	return cs
}

// CreateCastSpriteWithTransColor は透明色付きでCastSpriteを作成する
// 要件 8.4: 透明色処理をサポートする
// パフォーマンス最適化: 透明色処理済みの画像をキャッシュし、毎フレームの再処理を避ける
// 注意: 透明色処理はキャスト作成時に一度だけ行われ、結果がキャッシュされる
func (csm *CastSpriteManager) CreateCastSpriteWithTransColor(
	cast *Cast,
	srcImage *ebiten.Image,
	zOrder int,
	transColor color.Color,
) *CastSprite {
	csm.mu.Lock()
	defer csm.mu.Unlock()

	if cast == nil {
		return nil
	}

	// キャストのソース領域を切り出す（透明色処理付き）
	// パフォーマンス最適化: 透明色処理を一度だけ行い、結果をキャッシュする
	var img *ebiten.Image
	if srcImage != nil && cast.Width > 0 && cast.Height > 0 {
		if transColor != nil {
			// 透明色処理付きで切り出し
			img = csm.extractCastImageWithTransColorFromEbiten(srcImage, cast, transColor)
		} else {
			img = csm.extractCastImage(srcImage, cast)
		}
	}

	// スプライトを作成（非表示状態で作成）
	// 注意: zOrderパラメータは互換性のために残されているが、
	// 実際のZ順序はZ_Pathで管理される
	// レースコンディション対策: CreateSpriteHiddenを使用して最初から非表示で作成
	// Z_Pathを設定した後にSetVisible(true)を呼ぶ必要がある
	sprite := csm.spriteManager.CreateSpriteHidden(img)
	sprite.SetPosition(float64(cast.X), float64(cast.Y))
	// 注意: visibleはZ_Path設定後に設定される（CreateCastSpriteWithTransColorAndParentで）

	cs := &CastSprite{
		cast:          cast,
		sprite:        sprite,
		srcPicID:      cast.PicID,
		srcImage:      srcImage,
		lastSrcX:      cast.SrcX,
		lastSrcY:      cast.SrcY,
		lastWidth:     cast.Width,
		lastHeight:    cast.Height,
		transColor:    transColor,
		hasTransColor: transColor != nil,
		cachedImage:   img, // 透明色処理済みの画像をキャッシュ
		dirty:         false,
	}

	// 透明色が設定されている場合、customDraw関数を設定
	// パフォーマンス最適化: キャッシュ済みの画像を使用するため、毎フレームの透明色処理は不要
	if transColor != nil {
		cs.setupCustomDrawWithCache()
	}

	csm.castSprites[cast.ID] = cs
	return cs
}

// CreateCastSpriteWithTransColorAndParent は透明色付きでCastSpriteを作成し、親スプライトを設定する
// 要件 8.4: 透明色処理をサポートする
// 要件 14.2: ウインドウ内のスプライトをウインドウの子スプライトとして管理する
// 要件 2.2: PutCastが呼び出されたとき、現在のZ_Order_Counterを使用してLocal_Z_Orderを割り当てる
// 要件 1.4: 子スプライトが作成されたとき、親のZ_Pathを継承し、自身のLocal_Z_Orderを追加する
func (csm *CastSpriteManager) CreateCastSpriteWithTransColorAndParent(
	cast *Cast,
	srcImage *ebiten.Image,
	zOrder int,
	transColor color.Color,
	parent *Sprite,
) *CastSprite {
	cs := csm.CreateCastSpriteWithTransColor(cast, srcImage, zOrder, transColor)
	if cs != nil && parent != nil {
		cs.SetParent(parent)

		// 要件 2.2, 2.6: 操作順序でLocal_Z_Orderを割り当てる
		// 要件 1.4: 親のZ_Pathを継承し、自身のLocal_Z_Orderを追加する
		if parent.GetZPath() != nil {
			// 親のIDを使用してZOrderCounterから次のLocal_Z_Orderを取得
			localZOrder := csm.spriteManager.GetZOrderCounter().GetNext(parent.ID())
			zPath := NewZPathFromParent(parent.GetZPath(), localZOrder)
			cs.sprite.SetZPath(zPath)

			// レースコンディション対策: Z_Pathを設定した後にSetVisibleを呼ぶ
			// これにより、Draw()がスナップショットを取る際に一貫した状態を保証する
			if cs.cast != nil && cs.cast.Visible {
				cs.sprite.SetVisible(true)
			}

			csm.spriteManager.MarkNeedSort()
		}
	}
	return cs
}

// extractCastImage はソース画像からキャストの領域を切り出す
func (csm *CastSpriteManager) extractCastImage(srcImage *ebiten.Image, cast *Cast) *ebiten.Image {
	if srcImage == nil || cast.Width <= 0 || cast.Height <= 0 {
		return nil
	}

	srcBounds := srcImage.Bounds()

	// ソース領域のクリッピング
	srcX := cast.SrcX
	srcY := cast.SrcY
	srcW := cast.Width
	srcH := cast.Height

	if srcX < 0 {
		srcW += srcX
		srcX = 0
	}
	if srcY < 0 {
		srcH += srcY
		srcY = 0
	}
	if srcX+srcW > srcBounds.Dx() {
		srcW = srcBounds.Dx() - srcX
	}
	if srcY+srcH > srcBounds.Dy() {
		srcH = srcBounds.Dy() - srcY
	}

	if srcW <= 0 || srcH <= 0 {
		return nil
	}

	// ソース領域を切り出す
	srcRect := image.Rect(srcX, srcY, srcX+srcW, srcY+srcH)
	subImg := srcImage.SubImage(srcRect).(*ebiten.Image)

	// 新しい画像にコピー
	img := ebiten.NewImage(srcW, srcH)
	img.DrawImage(subImg, nil)

	return img
}

// extractCastImageWithTransColor はソース画像からキャストの領域を切り出し、透明色処理を適用する
// 注意: この関数はゲームループ外では動作しない（ebiten.Image.At()を使用するため）
// 代わりにextractCastImageWithTransColorFromEbitenを使用してください
func (csm *CastSpriteManager) extractCastImageWithTransColor(srcImage *ebiten.Image, cast *Cast, transColor color.Color) *ebiten.Image {
	if srcImage == nil || cast.Width <= 0 || cast.Height <= 0 {
		return nil
	}

	srcBounds := srcImage.Bounds()

	// ソース領域のクリッピング
	srcX := cast.SrcX
	srcY := cast.SrcY
	srcW := cast.Width
	srcH := cast.Height

	if srcX < 0 {
		srcW += srcX
		srcX = 0
	}
	if srcY < 0 {
		srcH += srcY
		srcY = 0
	}
	if srcX+srcW > srcBounds.Dx() {
		srcW = srcBounds.Dx() - srcX
	}
	if srcY+srcH > srcBounds.Dy() {
		srcH = srcBounds.Dy() - srcY
	}

	if srcW <= 0 || srcH <= 0 {
		return nil
	}

	// ソース領域を切り出す
	srcRect := image.Rect(srcX, srcY, srcX+srcW, srcY+srcH)
	subImg := srcImage.SubImage(srcRect).(*ebiten.Image)

	// 新しい画像にコピー
	img := ebiten.NewImage(srcW, srcH)

	if transColor == nil {
		// 透明色がない場合は通常コピー
		img.DrawImage(subImg, nil)
	} else {
		// 透明色処理を適用
		applyColorKeyToImage(img, subImg, transColor)
	}

	return img
}

// extractCastImageWithTransColorFromEbiten はソース画像からキャストの領域を切り出す
// 透明色処理は描画時にシェーダーで行うため、ここでは通常のコピーのみ行う
// パフォーマンス最適化: 透明色処理はGPUシェーダーで行う
func (csm *CastSpriteManager) extractCastImageWithTransColorFromEbiten(srcImage *ebiten.Image, cast *Cast, transColor color.Color) *ebiten.Image {
	// 通常の切り出しを行う（透明色処理は描画時に行う）
	return csm.extractCastImage(srcImage, cast)
}

// GetCastSprite はキャストIDからCastSpriteを取得する
func (csm *CastSpriteManager) GetCastSprite(castID int) *CastSprite {
	csm.mu.RLock()
	defer csm.mu.RUnlock()
	return csm.castSprites[castID]
}

// GetCastSpritesByWindow はウィンドウIDに属するすべてのCastSpriteを取得する
func (csm *CastSpriteManager) GetCastSpritesByWindow(winID int) []*CastSprite {
	csm.mu.RLock()
	defer csm.mu.RUnlock()

	result := make([]*CastSprite, 0)
	for _, cs := range csm.castSprites {
		if cs.cast != nil && cs.cast.WinID == winID {
			result = append(result, cs)
		}
	}
	return result
}

// GetCastSpritesByPicID はピクチャーIDに属するすべてのCastSpriteを取得する
func (csm *CastSpriteManager) GetCastSpritesByPicID(picID int) []*CastSprite {
	csm.mu.RLock()
	defer csm.mu.RUnlock()

	result := make([]*CastSprite, 0)
	for _, cs := range csm.castSprites {
		if cs.srcPicID == picID {
			result = append(result, cs)
		}
	}
	return result
}

// GetCastSpritesInRegion は指定されたピクチャーIDの指定領域内にあるCastSpriteを取得する
// MovePicで転送元の領域内にあるCastSpriteを取得するために使用
// 注意: CastSpriteの位置（cast.X, cast.Y）が転送領域内にあるかをチェック
func (csm *CastSpriteManager) GetCastSpritesInRegion(picID, srcX, srcY, width, height int) []*CastSprite {
	csm.mu.RLock()
	defer csm.mu.RUnlock()

	result := make([]*CastSprite, 0)
	for _, cs := range csm.castSprites {
		if cs.cast == nil {
			continue
		}
		// CastSpriteの描画先ピクチャーIDをチェック
		// 注意: srcPicIDはソース画像のピクチャーID、cast.PicIDは描画先のピクチャーID
		// MovePicでは描画先のピクチャーIDを使用する
		// しかし、CastはPicIDに描画されるのではなく、ウインドウに配置される
		// ここでは、CastのソースピクチャーIDではなく、親スプライトのピクチャーIDを確認する必要がある
		// 実際には、CastSpriteの親がpicIDのPictureSpriteかどうかを確認する必要がある
		// 簡易的に、CastのX, Y位置が転送領域内にあるかをチェック
		if cs.cast.X >= srcX && cs.cast.X < srcX+width && cs.cast.Y >= srcY && cs.cast.Y < srcY+height {
			// 親スプライトのピクチャーIDを確認
			if cs.sprite != nil && cs.sprite.Parent() != nil {
				// 親がpicIDのPictureSpriteかどうかを確認するのは複雑なので、
				// ここではsrcPicIDを使用する（ソース画像のピクチャーID）
				// 実際には、CastはウインドウにアタッチされたPictureSpriteの子として配置される
			}
			result = append(result, cs)
		}
	}
	return result
}

// RemoveCastSprite はCastSpriteを削除する
// 要件 8.3: キャストを削除できる
func (csm *CastSpriteManager) RemoveCastSprite(castID int) {
	csm.mu.Lock()
	defer csm.mu.Unlock()

	cs, exists := csm.castSprites[castID]
	if !exists {
		return
	}

	// スプライトを削除
	if cs.sprite != nil {
		csm.spriteManager.RemoveSprite(cs.sprite.ID())
	}

	delete(csm.castSprites, castID)
}

// RemoveCastSpritesByWindow はウィンドウIDに属するすべてのCastSpriteを削除する
func (csm *CastSpriteManager) RemoveCastSpritesByWindow(winID int) {
	csm.mu.Lock()
	defer csm.mu.Unlock()

	toDelete := make([]int, 0)
	for castID, cs := range csm.castSprites {
		if cs.cast != nil && cs.cast.WinID == winID {
			toDelete = append(toDelete, castID)
		}
	}

	for _, castID := range toDelete {
		cs := csm.castSprites[castID]
		if cs.sprite != nil {
			csm.spriteManager.RemoveSprite(cs.sprite.ID())
		}
		delete(csm.castSprites, castID)
	}
}

// Clear はすべてのCastSpriteを削除する
func (csm *CastSpriteManager) Clear() {
	csm.mu.Lock()
	defer csm.mu.Unlock()

	for castID, cs := range csm.castSprites {
		if cs.sprite != nil {
			csm.spriteManager.RemoveSprite(cs.sprite.ID())
		}
		delete(csm.castSprites, castID)
	}
}

// Count は登録されているCastSpriteの数を返す
func (csm *CastSpriteManager) Count() int {
	csm.mu.RLock()
	defer csm.mu.RUnlock()
	return len(csm.castSprites)
}

// CastSprite methods

// GetCast は元のキャストを返す
func (cs *CastSprite) GetCast() *Cast {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.cast
}

// GetSprite はスプライトを返す
func (cs *CastSprite) GetSprite() *Sprite {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.sprite
}

// GetSrcPicID はソースピクチャーIDを返す
func (cs *CastSprite) GetSrcPicID() int {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.srcPicID
}

// HasTransColor は透明色が設定されているかを返す
func (cs *CastSprite) HasTransColor() bool {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.hasTransColor
}

// GetTransColor は透明色を返す
func (cs *CastSprite) GetTransColor() color.Color {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.transColor
}

// UpdatePosition はキャストの位置を更新する
// 要件 8.2: キャストの位置を移動できる（残像なし）
func (cs *CastSprite) UpdatePosition(x, y int) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if cs.cast != nil {
		cs.cast.X = x
		cs.cast.Y = y
	}
	if cs.sprite != nil {
		cs.sprite.SetPosition(float64(x), float64(y))
	}
}

// UpdateSource はキャストのソース領域を更新する
// 値が実際に変更された場合のみdirtyフラグを設定する
func (cs *CastSprite) UpdateSource(srcX, srcY, width, height int) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	// 前回の値と比較して変更があった場合のみdirtyフラグを設定
	if cs.lastSrcX != srcX || cs.lastSrcY != srcY ||
		cs.lastWidth != width || cs.lastHeight != height {
		cs.lastSrcX = srcX
		cs.lastSrcY = srcY
		cs.lastWidth = width
		cs.lastHeight = height
		cs.dirty = true
	}

	// castオブジェクトも更新（CastManagerと同期）
	if cs.cast != nil {
		cs.cast.SrcX = srcX
		cs.cast.SrcY = srcY
		cs.cast.Width = width
		cs.cast.Height = height
	}
}

// UpdatePicID はソースピクチャーIDを更新する
func (cs *CastSprite) UpdatePicID(picID int) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if cs.cast != nil {
		cs.cast.PicID = picID
	}
	cs.srcPicID = picID
	cs.dirty = true
}

// UpdateVisible は可視性を更新する
func (cs *CastSprite) UpdateVisible(visible bool) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if cs.cast != nil {
		cs.cast.Visible = visible
	}
	if cs.sprite != nil {
		cs.sprite.SetVisible(visible)
	}
}

// UpdateTransColor は透明色を更新する
// パフォーマンス最適化: 透明色が変更された場合、キャッシュを再構築する
func (cs *CastSprite) UpdateTransColor(transColor color.Color) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	cs.transColor = transColor
	cs.hasTransColor = transColor != nil
	if cs.cast != nil {
		cs.cast.TransColor = transColor
		cs.cast.HasTransColor = transColor != nil
	}
	cs.dirty = true

	// 透明色が設定された場合、customDraw関数を設定してキャッシュを再構築
	if transColor != nil {
		cs.setupCustomDrawLocked()
	} else {
		// 透明色が解除された場合、customDraw関数をクリアしてキャッシュを更新
		if cs.sprite != nil {
			cs.sprite.SetCustomDraw(nil)
			// 元の画像をキャッシュとして使用
			cs.cachedImage = cs.sprite.Image()
		}
	}
}

// setupCustomDraw はカスタム描画関数を設定する（ロック取得版）
func (cs *CastSprite) setupCustomDraw() {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.setupCustomDrawLocked()
}

// setupCustomDrawWithCache はキャッシュを使用するカスタム描画関数を設定する
// パフォーマンス最適化: 透明色処理を描画時に行うが、結果をキャッシュして再利用する
// 注意: この関数はキャスト作成時に呼び出され、透明色処理は描画時に行われる
func (cs *CastSprite) setupCustomDrawWithCache() {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if cs.sprite == nil || !cs.hasTransColor {
		return
	}

	// クロージャでCastSpriteへの参照を保持
	// 透明色処理は描画時に行われる（drawImageWithColorKey使用）
	cs.sprite.SetCustomDraw(func(screen *ebiten.Image, x, y float64, alpha float32) {
		cs.mu.RLock()
		img := cs.sprite.Image()
		transColor := cs.transColor
		hasTransColor := cs.hasTransColor
		cs.mu.RUnlock()

		if img == nil {
			return
		}

		if hasTransColor && transColor != nil {
			// 透明色処理を適用して描画
			drawImageWithColorKey(screen, img, int(x), int(y), transColor)
		} else {
			// 通常描画
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(x, y)
			if alpha < 1.0 {
				op.ColorScale.ScaleAlpha(alpha)
			}
			screen.DrawImage(img, op)
		}
	})
}

// setupCustomDrawLocked はカスタム描画関数を設定する（ロック済み版）
// 透明色処理を描画時に行うためのカスタム描画関数を設定する
// パフォーマンス最適化: 透明色処理済みの画像をキャッシュし、毎フレームの再処理を避ける
func (cs *CastSprite) setupCustomDrawLocked() {
	if cs.sprite == nil || !cs.hasTransColor {
		return
	}

	// 透明色処理済みの画像をキャッシュとして作成
	// これにより毎フレームの透明色処理を避ける
	cs.rebuildTransparentCacheLocked()

	// クロージャでCastSpriteへの参照を保持
	cs.sprite.SetCustomDraw(func(screen *ebiten.Image, x, y float64, alpha float32) {
		cs.mu.RLock()
		cachedImg := cs.cachedImage
		cs.mu.RUnlock()

		if cachedImg == nil {
			return
		}

		// キャッシュ済みの透明色処理画像を描画
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(x, y)
		if alpha < 1.0 {
			op.ColorScale.ScaleAlpha(alpha)
		}
		screen.DrawImage(cachedImg, op)
	})
}

// rebuildTransparentCacheLocked は透明色処理済みのキャッシュ画像を再構築する（ロック済み版）
// パフォーマンス最適化: 透明色処理を一度だけ行い、結果をキャッシュする
// 注意: この関数はソース画像（srcImage）から直接処理を行う
// Ebitengineのゲームループ外でも動作するように、image.NewRGBAを使用する
func (cs *CastSprite) rebuildTransparentCacheLocked() {
	if cs.srcImage == nil || cs.cast == nil {
		return
	}

	// ソース画像からキャストの領域を切り出して透明色処理を適用
	srcBounds := cs.srcImage.Bounds()

	// ソース領域のクリッピング
	srcX := cs.cast.SrcX
	srcY := cs.cast.SrcY
	srcW := cs.cast.Width
	srcH := cs.cast.Height

	if srcX < 0 {
		srcW += srcX
		srcX = 0
	}
	if srcY < 0 {
		srcH += srcY
		srcY = 0
	}
	if srcX+srcW > srcBounds.Dx() {
		srcW = srcBounds.Dx() - srcX
	}
	if srcY+srcH > srcBounds.Dy() {
		srcH = srcBounds.Dy() - srcY
	}

	if srcW <= 0 || srcH <= 0 {
		return
	}

	// 透明色のRGBA値を取得（8bit）
	tr, tg, tb, _ := cs.transColor.RGBA()
	tr8 := uint8(tr >> 8)
	tg8 := uint8(tg >> 8)
	tb8 := uint8(tb >> 8)

	// 新しいRGBA画像を作成
	processedImg := image.NewRGBA(image.Rect(0, 0, srcW, srcH))

	// ピクセル単位で透明色を処理
	for sy := range srcH {
		for sx := range srcW {
			c := cs.srcImage.At(srcBounds.Min.X+srcX+sx, srcBounds.Min.Y+srcY+sy)
			r, g, b, a := c.RGBA()

			r8 := uint8(r >> 8)
			g8 := uint8(g >> 8)
			b8 := uint8(b >> 8)

			if r8 == tr8 && g8 == tg8 && b8 == tb8 {
				processedImg.Set(sx, sy, color.RGBA{0, 0, 0, 0})
			} else {
				processedImg.Set(sx, sy, color.RGBA{r8, g8, b8, uint8(a >> 8)})
			}
		}
	}

	// キャッシュ画像を更新
	cs.cachedImage = ebiten.NewImageFromImage(processedImg)
}

// SetParent は親スプライトを設定する
// ウインドウ内のキャストで使用
func (cs *CastSprite) SetParent(parent *Sprite) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if cs.sprite != nil {
		cs.sprite.SetParent(parent)
	}
}

// IsDirty はキャッシュが無効かどうかを返す
func (cs *CastSprite) IsDirty() bool {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.dirty
}

// RebuildCache はキャッシュを再構築する
// パフォーマンス最適化: 透明色処理済みの画像をキャッシュし、毎フレームの再処理を避ける
func (cs *CastSprite) RebuildCache(srcImage *ebiten.Image) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if cs.cast == nil || srcImage == nil {
		return
	}

	cs.srcImage = srcImage

	// キャストの領域を切り出し（透明色処理なし）
	img := cs.extractImage(srcImage)
	if img == nil {
		return
	}

	// スプライトの画像を更新
	if cs.sprite != nil {
		cs.sprite.SetImage(img)
	}

	// 透明色が設定されている場合、透明色処理済みのキャッシュを再構築
	if cs.hasTransColor && cs.transColor != nil {
		cs.rebuildTransparentCacheLocked()
	} else {
		cs.cachedImage = img
	}

	cs.dirty = false
}

// extractImage はソース画像からキャストの領域を切り出す（内部用）
func (cs *CastSprite) extractImage(srcImage *ebiten.Image) *ebiten.Image {
	if srcImage == nil || cs.cast == nil || cs.cast.Width <= 0 || cs.cast.Height <= 0 {
		return nil
	}

	srcBounds := srcImage.Bounds()

	// ソース領域のクリッピング
	srcX := cs.cast.SrcX
	srcY := cs.cast.SrcY
	srcW := cs.cast.Width
	srcH := cs.cast.Height

	if srcX < 0 {
		srcW += srcX
		srcX = 0
	}
	if srcY < 0 {
		srcH += srcY
		srcY = 0
	}
	if srcX+srcW > srcBounds.Dx() {
		srcW = srcBounds.Dx() - srcX
	}
	if srcY+srcH > srcBounds.Dy() {
		srcH = srcBounds.Dy() - srcY
	}

	if srcW <= 0 || srcH <= 0 {
		return nil
	}

	// ソース領域を切り出す
	srcRect := image.Rect(srcX, srcY, srcX+srcW, srcY+srcH)
	subImg := srcImage.SubImage(srcRect).(*ebiten.Image)

	// 新しい画像にコピー
	img := ebiten.NewImage(srcW, srcH)
	img.DrawImage(subImg, nil)

	return img
}

// extractImageWithTransColor はソース画像からキャストの領域を切り出し、透明色処理を適用する（内部用）
func (cs *CastSprite) extractImageWithTransColor(srcImage *ebiten.Image) *ebiten.Image {
	if srcImage == nil || cs.cast == nil || cs.cast.Width <= 0 || cs.cast.Height <= 0 {
		return nil
	}

	srcBounds := srcImage.Bounds()

	// ソース領域のクリッピング
	srcX := cs.cast.SrcX
	srcY := cs.cast.SrcY
	srcW := cs.cast.Width
	srcH := cs.cast.Height

	if srcX < 0 {
		srcW += srcX
		srcX = 0
	}
	if srcY < 0 {
		srcH += srcY
		srcY = 0
	}
	if srcX+srcW > srcBounds.Dx() {
		srcW = srcBounds.Dx() - srcX
	}
	if srcY+srcH > srcBounds.Dy() {
		srcH = srcBounds.Dy() - srcY
	}

	if srcW <= 0 || srcH <= 0 {
		return nil
	}

	// ソース領域を切り出す
	srcRect := image.Rect(srcX, srcY, srcX+srcW, srcY+srcH)
	subImg := srcImage.SubImage(srcRect).(*ebiten.Image)

	// 新しい画像にコピー
	img := ebiten.NewImage(srcW, srcH)

	if cs.transColor == nil {
		// 透明色がない場合は通常コピー
		img.DrawImage(subImg, nil)
	} else {
		// 透明色処理を適用
		applyColorKeyToImage(img, subImg, cs.transColor)
	}

	return img
}

// applyColorKeyToImage は透明色処理を適用して画像をコピーする
// 注意: この関数はゲームループ外でも動作するように、image.NewRGBAを使用する
func applyColorKeyToImage(dst, src *ebiten.Image, transColor color.Color) {
	if dst == nil || src == nil {
		return
	}

	srcBounds := src.Bounds()
	width := srcBounds.Dx()
	height := srcBounds.Dy()

	if width <= 0 || height <= 0 {
		return
	}

	// 透明色のRGBA値を取得（8bit）
	tr, tg, tb, _ := transColor.RGBA()
	tr8 := uint8(tr >> 8)
	tg8 := uint8(tg >> 8)
	tb8 := uint8(tb >> 8)

	// 新しいRGBA画像を作成
	processedImg := image.NewRGBA(image.Rect(0, 0, width, height))

	// ピクセル単位で透明色を処理
	for sy := 0; sy < height; sy++ {
		for sx := 0; sx < width; sx++ {
			// ソース画像からピクセルを取得
			c := src.At(srcBounds.Min.X+sx, srcBounds.Min.Y+sy)
			r, g, b, a := c.RGBA()

			// 16bitから8bitに変換
			r8 := uint8(r >> 8)
			g8 := uint8(g >> 8)
			b8 := uint8(b >> 8)

			// 透明色と一致する場合、完全に透明にする
			if r8 == tr8 && g8 == tg8 && b8 == tb8 {
				processedImg.Set(sx, sy, color.RGBA{0, 0, 0, 0})
			} else {
				// 元の色を保持
				processedImg.Set(sx, sy, color.RGBA{r8, g8, b8, uint8(a >> 8)})
			}
		}
	}

	// Ebiten画像に変換して描画
	tmpImg := ebiten.NewImageFromImage(processedImg)
	dst.DrawImage(tmpImg, nil)
}
