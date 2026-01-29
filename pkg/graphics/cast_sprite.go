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

	// スプライトを作成
	// 注意: zOrderパラメータは互換性のために残されているが、
	// 実際のZ順序はZ_Pathで管理される
	sprite := csm.spriteManager.CreateSprite(img)
	sprite.SetPosition(float64(cast.X), float64(cast.Y))
	sprite.SetVisible(cast.Visible)

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
			csm.spriteManager.MarkNeedSort()
		}
	}
	return cs
}

// CreateCastSpriteWithTransColor は透明色付きでCastSpriteを作成する
// 要件 8.4: 透明色処理をサポートする
// 注意: 透明色処理は描画時に行われるため、ここでは透明色情報のみを保存する
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

	// キャストのソース領域を切り出す（透明色処理なし）
	// 透明色処理は描画時に行われる
	var img *ebiten.Image
	if srcImage != nil && cast.Width > 0 && cast.Height > 0 {
		img = csm.extractCastImage(srcImage, cast)
	}

	// スプライトを作成
	// 注意: zOrderパラメータは互換性のために残されているが、
	// 実際のZ順序はZ_Pathで管理される
	sprite := csm.spriteManager.CreateSprite(img)
	sprite.SetPosition(float64(cast.X), float64(cast.Y))
	sprite.SetVisible(cast.Visible)

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
		cachedImage:   img,
		dirty:         false,
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
func (cs *CastSprite) RebuildCache(srcImage *ebiten.Image) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if cs.cast == nil || srcImage == nil {
		return
	}

	cs.srcImage = srcImage

	// キャストの領域を切り出し
	var img *ebiten.Image
	if cs.hasTransColor {
		img = cs.extractImageWithTransColor(srcImage)
	} else {
		img = cs.extractImage(srcImage)
	}

	cs.cachedImage = img
	if cs.sprite != nil {
		cs.sprite.SetImage(img)
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
