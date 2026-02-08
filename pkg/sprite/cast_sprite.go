// Package sprite provides sprite-based rendering system with slice-based draw ordering.
package sprite

import (
	"image"
	"image/color"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

// CastConfig はキャスト作成時の設定
type CastConfig struct {
	ID            int
	WinID         int
	PicID         int
	X, Y          int
	SrcX, SrcY    int
	Width, Height int
	Visible       bool
	TransColor    color.Color
	HasTransColor bool
}

// CastSprite はキャストとスプライトを組み合わせたラッパー構造体
// 要件 6.1: キャストをスプライトとして作成できる
// 要件 6.2: キャストの位置を移動できる（残像なし）
// 要件 6.3: キャストを削除できる
// 要件 6.4: 透明色処理をサポートする
type CastSprite struct {
	castID int     // キャストID
	sprite *Sprite // スプライト（キャストの画像を保持）

	// キャスト情報
	winID  int
	picID  int
	x, y   int
	srcX   int
	srcY   int
	width  int
	height int

	// ソース画像情報
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
// 要件 6.1: キャストをスプライトとして作成できる
func (csm *CastSpriteManager) CreateCastSprite(
	config CastConfig,
	srcImage *ebiten.Image,
	parent *Sprite,
) *CastSprite {
	csm.mu.Lock()
	defer csm.mu.Unlock()

	// キャストのソース領域を切り出す
	var img *ebiten.Image
	if srcImage != nil && config.Width > 0 && config.Height > 0 {
		img = extractCastImage(srcImage, config.SrcX, config.SrcY, config.Width, config.Height)
	}

	// スプライトを作成
	sprite := csm.spriteManager.CreateSprite(img, parent)
	sprite.SetPosition(float64(config.X), float64(config.Y))
	sprite.SetVisible(config.Visible)

	cs := &CastSprite{
		castID:        config.ID,
		sprite:        sprite,
		winID:         config.WinID,
		picID:         config.PicID,
		x:             config.X,
		y:             config.Y,
		srcX:          config.SrcX,
		srcY:          config.SrcY,
		width:         config.Width,
		height:        config.Height,
		srcImage:      srcImage,
		lastSrcX:      config.SrcX,
		lastSrcY:      config.SrcY,
		lastWidth:     config.Width,
		lastHeight:    config.Height,
		transColor:    config.TransColor,
		hasTransColor: config.HasTransColor,
		cachedImage:   img,
		dirty:         false,
	}

	csm.castSprites[config.ID] = cs
	return cs
}

// CreateCastSpriteWithTransColor は透明色付きでCastSpriteを作成する
// 要件 6.4: 透明色処理をサポートする
func (csm *CastSpriteManager) CreateCastSpriteWithTransColor(
	config CastConfig,
	srcImage *ebiten.Image,
	transColor color.Color,
	parent *Sprite,
) *CastSprite {
	config.TransColor = transColor
	config.HasTransColor = transColor != nil
	return csm.CreateCastSprite(config, srcImage, parent)
}

// extractCastImage はソース画像からキャストの領域を切り出す
func extractCastImage(srcImage *ebiten.Image, srcX, srcY, width, height int) *ebiten.Image {
	if srcImage == nil || width <= 0 || height <= 0 {
		return nil
	}

	srcBounds := srcImage.Bounds()

	// ソース領域のクリッピング
	if srcX < 0 {
		width += srcX
		srcX = 0
	}
	if srcY < 0 {
		height += srcY
		srcY = 0
	}
	if srcX+width > srcBounds.Dx() {
		width = srcBounds.Dx() - srcX
	}
	if srcY+height > srcBounds.Dy() {
		height = srcBounds.Dy() - srcY
	}

	if width <= 0 || height <= 0 {
		return nil
	}

	// ソース領域を切り出す
	srcRect := image.Rect(srcX, srcY, srcX+width, srcY+height)
	subImg := srcImage.SubImage(srcRect).(*ebiten.Image)

	// 新しい画像にコピー
	img := ebiten.NewImage(width, height)
	img.DrawImage(subImg, nil)

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
		if cs.winID == winID {
			result = append(result, cs)
		}
	}
	return result
}

// RemoveCastSprite はCastSpriteを削除する
// 要件 6.3: キャストを削除できる
func (csm *CastSpriteManager) RemoveCastSprite(castID int) {
	csm.mu.Lock()
	defer csm.mu.Unlock()

	cs, exists := csm.castSprites[castID]
	if !exists {
		return
	}

	// スプライトを削除
	if cs.sprite != nil {
		csm.spriteManager.DeleteSprite(cs.sprite.ID())
	}

	delete(csm.castSprites, castID)
}

// RemoveCastSpritesByWindow はウィンドウIDに属するすべてのCastSpriteを削除する
func (csm *CastSpriteManager) RemoveCastSpritesByWindow(winID int) {
	csm.mu.Lock()
	defer csm.mu.Unlock()

	toDelete := make([]int, 0)
	for castID, cs := range csm.castSprites {
		if cs.winID == winID {
			toDelete = append(toDelete, castID)
		}
	}

	for _, castID := range toDelete {
		cs := csm.castSprites[castID]
		if cs.sprite != nil {
			csm.spriteManager.DeleteSprite(cs.sprite.ID())
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
			csm.spriteManager.DeleteSprite(cs.sprite.ID())
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

// GetCastID はキャストIDを返す
func (cs *CastSprite) GetCastID() int {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.castID
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
	return cs.picID
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
// 要件 6.2: キャストの位置を移動できる（残像なし）
func (cs *CastSprite) UpdatePosition(x, y int) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	cs.x = x
	cs.y = y
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

	cs.srcX = srcX
	cs.srcY = srcY
	cs.width = width
	cs.height = height
}

// UpdatePicID はソースピクチャーIDを更新する
func (cs *CastSprite) UpdatePicID(picID int) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	cs.picID = picID
	cs.dirty = true
}

// UpdateVisible は可視性を更新する
func (cs *CastSprite) UpdateVisible(visible bool) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

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
	cs.dirty = true
}

// SetParent は親スプライトを設定する
// ウインドウ内のキャストで使用
func (cs *CastSprite) SetParent(parent *Sprite) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if cs.sprite != nil && parent != nil {
		parent.AddChild(cs.sprite)
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

	if srcImage == nil {
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
	if srcImage == nil || cs.width <= 0 || cs.height <= 0 {
		return nil
	}

	srcBounds := srcImage.Bounds()

	// ソース領域のクリッピング
	srcX := cs.srcX
	srcY := cs.srcY
	srcW := cs.width
	srcH := cs.height

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
	if srcImage == nil || cs.width <= 0 || cs.height <= 0 {
		return nil
	}

	srcBounds := srcImage.Bounds()

	// ソース領域のクリッピング
	srcX := cs.srcX
	srcY := cs.srcY
	srcW := cs.width
	srcH := cs.height

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
		ApplyColorKeyToImage(img, subImg, cs.transColor)
	}

	return img
}

// ApplyColorKeyToImage は透明色処理を適用して画像をコピーする
// 注意: この関数はゲームループ外でも動作するように、image.NewRGBAを使用する
func ApplyColorKeyToImage(dst, src *ebiten.Image, transColor color.Color) {
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
	for sy := range height {
		for sx := range width {
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
