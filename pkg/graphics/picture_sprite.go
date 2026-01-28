// Package graphics provides sprite-based rendering system.
package graphics

import (
	"image"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

// PictureSprite はピクチャ描画（MovePic）をスプライトとして表現するラッパー構造体
// 要件 6.1: BMPファイルからスプライトを作成できる
// 要件 6.2: 透明色を指定できる
// 要件 6.3: ピクチャの一部を切り出してスプライトにできる
type PictureSprite struct {
	sprite *Sprite // 基盤となるスプライト

	// 元のDrawingEntry情報
	picID  int // ソースピクチャーID
	srcX   int // ソース領域の左上X座標
	srcY   int // ソース領域の左上Y座標
	width  int // 描画幅
	height int // 描画高さ
	destX  int // 描画先X座標
	destY  int // 描画先Y座標

	// 透明色処理
	transparent      bool        // 透明色処理が有効かどうか
	transparentColor image.Point // 透明色（未使用、将来の拡張用）
}

// PictureSpriteManager はPictureSpriteを管理する
type PictureSpriteManager struct {
	pictureSprites map[int][]*PictureSprite // picID -> PictureSprites（同じピクチャに複数のスプライトがある場合）
	spriteManager  *SpriteManager
	mu             sync.RWMutex
	nextID         int // 内部ID管理
}

// NewPictureSpriteManager は新しいPictureSpriteManagerを作成する
func NewPictureSpriteManager(sm *SpriteManager) *PictureSpriteManager {
	return &PictureSpriteManager{
		pictureSprites: make(map[int][]*PictureSprite),
		spriteManager:  sm,
		nextID:         1,
	}
}

// CreatePictureSprite はMovePicの結果からPictureSpriteを作成する
// 要件 6.1: BMPファイルからスプライトを作成できる
// 要件 6.3: ピクチャの一部を切り出してスプライトにできる
func (psm *PictureSpriteManager) CreatePictureSprite(
	srcImg *ebiten.Image,
	picID int,
	srcX, srcY, width, height int,
	destX, destY int,
	zOrder int,
	transparent bool,
) *PictureSprite {
	psm.mu.Lock()
	defer psm.mu.Unlock()

	// スプライト用の画像を作成（ソース画像をコピー）
	img := ebiten.NewImage(width, height)
	if srcImg != nil {
		img.DrawImage(srcImg, nil)
	}

	// スプライトを作成
	sprite := psm.spriteManager.CreateSprite(img)
	sprite.SetPosition(float64(destX), float64(destY))
	sprite.SetZOrder(zOrder)
	sprite.SetVisible(true)

	ps := &PictureSprite{
		sprite:      sprite,
		picID:       picID,
		srcX:        srcX,
		srcY:        srcY,
		width:       width,
		height:      height,
		destX:       destX,
		destY:       destY,
		transparent: transparent,
	}

	// ピクチャIDごとにスプライトを管理
	psm.pictureSprites[picID] = append(psm.pictureSprites[picID], ps)
	psm.nextID++

	return ps
}

// CreatePictureSpriteWithParent はMovePicの結果からPictureSpriteを作成し、親スプライトを設定する
// 要件 6.1: BMPファイルからスプライトを作成できる
// 要件 6.3: ピクチャの一部を切り出してスプライトにできる
// 要件 14.2: ウインドウ内のスプライトをウインドウの子スプライトとして管理する
func (psm *PictureSpriteManager) CreatePictureSpriteWithParent(
	srcImg *ebiten.Image,
	picID int,
	srcX, srcY, width, height int,
	destX, destY int,
	zOrder int,
	transparent bool,
	parent *Sprite,
) *PictureSprite {
	ps := psm.CreatePictureSprite(srcImg, picID, srcX, srcY, width, height, destX, destY, zOrder, transparent)
	if ps != nil && parent != nil {
		ps.SetParent(parent)
	}
	return ps
}

// CreatePictureSpriteFromDrawingEntry はDrawingEntryからPictureSpriteを作成する
// 既存のDrawingEntryをスプライトベースに変換するアダプタ
func (psm *PictureSpriteManager) CreatePictureSpriteFromDrawingEntry(entry *DrawingEntry, zOrder int) *PictureSprite {
	if entry == nil {
		return nil
	}

	return psm.CreatePictureSprite(
		entry.GetImage(),
		entry.GetPicID(),
		0, 0, // DrawingEntryはすでに切り出し済みなのでsrcX, srcYは0
		entry.GetWidth(),
		entry.GetHeight(),
		entry.GetDestX(),
		entry.GetDestY(),
		zOrder,
		false, // 透明色処理はDrawingEntry作成時に適用済み
	)
}

// GetPictureSprites はピクチャIDに関連するすべてのPictureSpriteを取得する
func (psm *PictureSpriteManager) GetPictureSprites(picID int) []*PictureSprite {
	psm.mu.RLock()
	defer psm.mu.RUnlock()

	sprites := psm.pictureSprites[picID]
	if sprites == nil {
		return nil
	}

	// コピーを返す
	result := make([]*PictureSprite, len(sprites))
	copy(result, sprites)
	return result
}

// RemovePictureSprite は指定されたPictureSpriteを削除する
func (psm *PictureSpriteManager) RemovePictureSprite(ps *PictureSprite) {
	if ps == nil {
		return
	}

	psm.mu.Lock()
	defer psm.mu.Unlock()

	// スプライトを削除
	psm.spriteManager.RemoveSprite(ps.sprite.ID())

	// リストから削除
	sprites := psm.pictureSprites[ps.picID]
	for i, s := range sprites {
		if s == ps {
			psm.pictureSprites[ps.picID] = append(sprites[:i], sprites[i+1:]...)
			break
		}
	}

	// リストが空になったら削除
	if len(psm.pictureSprites[ps.picID]) == 0 {
		delete(psm.pictureSprites, ps.picID)
	}
}

// RemovePictureSpritesByPicID はピクチャIDに関連するすべてのPictureSpriteを削除する
func (psm *PictureSpriteManager) RemovePictureSpritesByPicID(picID int) {
	psm.mu.Lock()
	defer psm.mu.Unlock()

	sprites := psm.pictureSprites[picID]
	for _, ps := range sprites {
		psm.spriteManager.RemoveSprite(ps.sprite.ID())
	}
	delete(psm.pictureSprites, picID)
}

// Clear はすべてのPictureSpriteを削除する
func (psm *PictureSpriteManager) Clear() {
	psm.mu.Lock()
	defer psm.mu.Unlock()

	for picID, sprites := range psm.pictureSprites {
		for _, ps := range sprites {
			psm.spriteManager.RemoveSprite(ps.sprite.ID())
		}
		delete(psm.pictureSprites, picID)
	}
}

// Count は登録されているPictureSpriteの総数を返す
func (psm *PictureSpriteManager) Count() int {
	psm.mu.RLock()
	defer psm.mu.RUnlock()

	count := 0
	for _, sprites := range psm.pictureSprites {
		count += len(sprites)
	}
	return count
}

// PictureSprite methods

// GetSprite は基盤となるスプライトを返す
func (ps *PictureSprite) GetSprite() *Sprite {
	return ps.sprite
}

// GetPicID はソースピクチャーIDを返す
func (ps *PictureSprite) GetPicID() int {
	return ps.picID
}

// GetSrcX はソース領域の左上X座標を返す
func (ps *PictureSprite) GetSrcX() int {
	return ps.srcX
}

// GetSrcY はソース領域の左上Y座標を返す
func (ps *PictureSprite) GetSrcY() int {
	return ps.srcY
}

// GetWidth は描画幅を返す
func (ps *PictureSprite) GetWidth() int {
	return ps.width
}

// GetHeight は描画高さを返す
func (ps *PictureSprite) GetHeight() int {
	return ps.height
}

// GetDestX は描画先X座標を返す
func (ps *PictureSprite) GetDestX() int {
	return ps.destX
}

// GetDestY は描画先Y座標を返す
func (ps *PictureSprite) GetDestY() int {
	return ps.destY
}

// IsTransparent は透明色処理が有効かどうかを返す
func (ps *PictureSprite) IsTransparent() bool {
	return ps.transparent
}

// SetPosition は描画位置を更新する
func (ps *PictureSprite) SetPosition(x, y int) {
	ps.destX = x
	ps.destY = y
	ps.sprite.SetPosition(float64(x), float64(y))
}

// SetZOrder はZ順序を更新する
func (ps *PictureSprite) SetZOrder(z int) {
	ps.sprite.SetZOrder(z)
}

// SetVisible は可視性を更新する
func (ps *PictureSprite) SetVisible(visible bool) {
	ps.sprite.SetVisible(visible)
}

// SetParent は親スプライトを設定する
// ウインドウ内のピクチャ描画で使用
func (ps *PictureSprite) SetParent(parent *Sprite) {
	ps.sprite.SetParent(parent)
}

// UpdateImage はスプライトの画像を更新する
func (ps *PictureSprite) UpdateImage(img *ebiten.Image) {
	ps.sprite.SetImage(img)
	if img != nil {
		bounds := img.Bounds()
		ps.width = bounds.Dx()
		ps.height = bounds.Dy()
	}
}
