// Package sprite provides sprite-based rendering system with slice-based draw ordering.
package sprite

import (
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

// PictureSpriteState はPictureSpriteの状態を表す
// 要件 14.1: PictureSpriteは「未関連付け」「関連付け済み」の状態を持つ
type PictureSpriteState int

const (
	// PictureSpriteUnattached は未関連付け状態
	// LoadPic時に作成され、まだウインドウに関連付けられていない状態
	// 要件 14.2: この状態ではスプライトを描画しない
	PictureSpriteUnattached PictureSpriteState = iota

	// PictureSpriteAttached は関連付け済み状態
	// SetPic/OpenWinでウインドウに関連付けられた状態
	// 要件 14.3: この状態では親ウインドウの可視性に従って描画する
	PictureSpriteAttached
)

// String はPictureSpriteStateの文字列表現を返す
func (s PictureSpriteState) String() string {
	switch s {
	case PictureSpriteUnattached:
		return "Unattached"
	case PictureSpriteAttached:
		return "Attached"
	default:
		return "Unknown"
	}
}

// PictureSprite はピクチャ描画をスプライトとして表現するラッパー構造体
// 要件 5.1: BMPファイルからスプライトを作成できる
// 要件 5.2: 透明色を指定できる
// 要件 5.3: ピクチャの一部を切り出してスプライトにできる
// 要件 13.1: LoadPicが呼び出されたとき、非表示のPictureSpriteを作成する
// 要件 14.1: PictureSpriteは「未関連付け」「関連付け済み」の状態を持つ
type PictureSprite struct {
	sprite *Sprite // 基盤となるスプライト

	// 元の情報
	picID  int // ソースピクチャーID
	srcX   int // ソース領域の左上X座標
	srcY   int // ソース領域の左上Y座標
	width  int // 描画幅
	height int // 描画高さ
	destX  int // 描画先X座標
	destY  int // 描画先Y座標

	// 透明色処理
	transparent bool // 透明色処理が有効かどうか

	// 状態管理
	// 要件 14.1: PictureSpriteは「未関連付け」「関連付け済み」の状態を持つ
	state    PictureSpriteState // 現在の状態
	windowID int                // 関連付けられたウインドウID（-1 = 未関連付け）
}

// PictureSpriteManager はPictureSpriteを管理する
type PictureSpriteManager struct {
	pictureSprites   map[int][]*PictureSprite // picID -> PictureSprites（同じピクチャに複数のスプライトがある場合）
	pictureSpriteMap map[int]*PictureSprite   // picID -> 背景PictureSprite（LoadPic時に作成される主要なスプライト）
	spriteManager    *SpriteManager
	mu               sync.RWMutex
	nextID           int // 内部ID管理
}

// NewPictureSpriteManager は新しいPictureSpriteManagerを作成する
func NewPictureSpriteManager(sm *SpriteManager) *PictureSpriteManager {
	return &PictureSpriteManager{
		pictureSprites:   make(map[int][]*PictureSprite),
		pictureSpriteMap: make(map[int]*PictureSprite),
		spriteManager:    sm,
		nextID:           1,
	}
}

// CreatePictureSprite はPictureSpriteを作成する
// 要件 5.1: BMPファイルからスプライトを作成できる
// 要件 5.3: ピクチャの一部を切り出してスプライトにできる
func (psm *PictureSpriteManager) CreatePictureSprite(
	srcImg *ebiten.Image,
	picID int,
	srcX, srcY, width, height int,
	destX, destY int,
	transparent bool,
	parent *Sprite,
) *PictureSprite {
	psm.mu.Lock()
	defer psm.mu.Unlock()

	// スプライト用の画像を作成（ソース画像をコピー）
	img := ebiten.NewImage(width, height)
	if srcImg != nil {
		img.DrawImage(srcImg, nil)
	}

	// スプライトを作成
	sprite := psm.spriteManager.CreateSprite(img, parent)
	sprite.SetPosition(float64(destX), float64(destY))
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
		state:       PictureSpriteAttached, // 親が指定されている場合は関連付け済み
		windowID:    -1,                    // 後で設定される
	}

	// ピクチャIDごとにスプライトを管理
	psm.pictureSprites[picID] = append(psm.pictureSprites[picID], ps)
	psm.nextID++

	return ps
}

// CreateBackgroundPictureSprite は背景用のPictureSpriteを作成する
// 背景PictureSpriteはピクチャーの画像への参照を保持し、コピーしない
// これにより、MovePicで更新されたピクチャー画像が反映される
func (psm *PictureSpriteManager) CreateBackgroundPictureSprite(
	srcImg *ebiten.Image,
	picID int,
	width, height int,
	destX, destY int,
	parent *Sprite,
) *PictureSprite {
	psm.mu.Lock()
	defer psm.mu.Unlock()

	// 背景PictureSpriteはピクチャーの画像への参照を保持（コピーしない）
	sprite := psm.spriteManager.CreateSprite(srcImg, parent)
	sprite.SetPosition(float64(destX), float64(destY))
	sprite.SetVisible(true)

	ps := &PictureSprite{
		sprite:      sprite,
		picID:       picID,
		srcX:        0,
		srcY:        0,
		width:       width,
		height:      height,
		destX:       destX,
		destY:       destY,
		transparent: false,
		state:       PictureSpriteAttached, // OpenWin経由で作成されるので関連付け済み
		windowID:    -1,                    // 後でAttachPictureSpriteToWindowで設定される
	}

	// ピクチャIDごとにスプライトを管理
	psm.pictureSprites[picID] = append(psm.pictureSprites[picID], ps)
	psm.nextID++

	return ps
}

// CreatePictureSpriteOnLoad はLoadPic時に非表示のPictureSpriteを作成する
// 要件 13.1: LoadPicが呼び出されたとき、非表示のPictureSpriteを作成する
// 要件 14.1: PictureSpriteは「未関連付け」状態で作成される
// 要件 14.2: 未関連付け状態ではスプライトを描画しない
func (psm *PictureSpriteManager) CreatePictureSpriteOnLoad(
	srcImg *ebiten.Image,
	picID int,
	width, height int,
) *PictureSprite {
	psm.mu.Lock()
	defer psm.mu.Unlock()

	// 既存のPictureSpriteがあれば削除する
	if existing, ok := psm.pictureSpriteMap[picID]; ok {
		psm.removePictureSpriteInternal(existing)
	}

	// 非表示状態でPictureSpriteを作成する
	// ピクチャーの画像への参照を保持（コピーしない）
	sprite := psm.spriteManager.CreateSprite(srcImg, nil)
	sprite.SetPosition(0, 0)
	sprite.SetVisible(false) // 未関連付けなので非表示

	ps := &PictureSprite{
		sprite:      sprite,
		picID:       picID,
		srcX:        0,
		srcY:        0,
		width:       width,
		height:      height,
		destX:       0,
		destY:       0,
		transparent: false,
		state:       PictureSpriteUnattached, // 未関連付け状態
		windowID:    -1,                      // 未関連付け
	}

	// pictureSpriteMapに登録する
	psm.pictureSpriteMap[picID] = ps

	// ピクチャIDごとにスプライトを管理（既存のリストにも追加）
	psm.pictureSprites[picID] = append(psm.pictureSprites[picID], ps)
	psm.nextID++

	return ps
}

// removePictureSpriteInternal は内部用のPictureSprite削除メソッド（ロック不要）
func (psm *PictureSpriteManager) removePictureSpriteInternal(ps *PictureSprite) {
	if ps == nil {
		return
	}

	// 子スプライトを再帰的に削除
	for _, child := range ps.sprite.GetChildren() {
		psm.spriteManager.DeleteSprite(child.ID())
	}

	// 親から削除
	if ps.sprite.Parent() != nil {
		ps.sprite.Parent().RemoveChild(ps.sprite.ID())
	}

	// スプライトを削除
	psm.spriteManager.DeleteSprite(ps.sprite.ID())

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

	// pictureSpriteMapから削除
	if psm.pictureSpriteMap[ps.picID] == ps {
		delete(psm.pictureSpriteMap, ps.picID)
	}
}

// AttachPictureSpriteToWindow はSetPic/OpenWin時にPictureSpriteをウインドウに関連付ける
// 要件 13.3: SetPicが呼び出されたとき、既存のPictureSpriteをウインドウの子として関連付ける
// 要件 13.4: SetPicが呼び出されたとき、PictureSpriteを表示状態にする
func (psm *PictureSpriteManager) AttachPictureSpriteToWindow(picID int, windowSprite *Sprite, windowID int) error {
	psm.mu.Lock()
	defer psm.mu.Unlock()

	ps, ok := psm.pictureSpriteMap[picID]
	if !ok {
		// pictureSpriteMapに存在しない場合はエラーではなく、nilを返す
		// （従来の方法でPictureSpriteが作成される場合があるため）
		return nil
	}

	// PictureSpriteをWindowSpriteの子として追加する
	if windowSprite != nil {
		windowSprite.AddChild(ps.sprite)
	}

	// 状態をAttachedに変更し、表示状態にする
	ps.state = PictureSpriteAttached
	ps.windowID = windowID
	ps.sprite.SetVisible(true)

	// 関連付け後はpictureSpriteMapから削除する
	// これにより、同じピクチャを複数ウインドウで使用する場合、
	// 次のOpenWin時に新しいPictureSpriteが作成される
	delete(psm.pictureSpriteMap, picID)

	return nil
}

// GetPictureSpriteByPictureID はピクチャ番号からPictureSpriteを取得する
// 要件 14.4: ピクチャ番号からPictureSpriteを効率的に検索できる
func (psm *PictureSpriteManager) GetPictureSpriteByPictureID(picID int) *PictureSprite {
	psm.mu.RLock()
	defer psm.mu.RUnlock()
	return psm.pictureSpriteMap[picID]
}

// FreePictureSprite はピクチャ解放時にPictureSpriteを削除する
// 要件 13.7: ピクチャが解放されたとき、対応するPictureSpriteとその子スプライトを削除する
func (psm *PictureSpriteManager) FreePictureSprite(picID int) {
	psm.mu.Lock()
	defer psm.mu.Unlock()

	// pictureSpriteMapから削除
	if ps, ok := psm.pictureSpriteMap[picID]; ok {
		psm.removePictureSpriteInternal(ps)
		delete(psm.pictureSpriteMap, picID)
	}

	// pictureSpritesからも削除（関連付け済みのPictureSpriteも含む）
	sprites := psm.pictureSprites[picID]
	for _, ps := range sprites {
		// 子スプライトを再帰的に削除
		for _, child := range ps.sprite.GetChildren() {
			psm.spriteManager.DeleteSprite(child.ID())
		}

		// 親から削除
		if ps.sprite.Parent() != nil {
			ps.sprite.Parent().RemoveChild(ps.sprite.ID())
		}

		// スプライトを削除
		psm.spriteManager.DeleteSprite(ps.sprite.ID())
	}
	delete(psm.pictureSprites, picID)
}

// UpdatePictureSpriteImage はMovePic時にPictureSpriteの画像を更新する
// 要件 14.5: MovePicが呼び出されたとき、転送先ピクチャのPictureSpriteの画像を更新する
func (psm *PictureSpriteManager) UpdatePictureSpriteImage(picID int, img *ebiten.Image) {
	psm.mu.Lock()
	defer psm.mu.Unlock()

	// pictureSpriteMapから取得
	if ps, ok := psm.pictureSpriteMap[picID]; ok {
		ps.sprite.SetImage(img)
		if img != nil {
			bounds := img.Bounds()
			ps.width = bounds.Dx()
			ps.height = bounds.Dy()
		}
	}

	// pictureSpritesからも更新（関連付け済みのPictureSpriteも含む）
	sprites := psm.pictureSprites[picID]
	for _, ps := range sprites {
		ps.sprite.SetImage(img)
		if img != nil {
			bounds := img.Bounds()
			ps.width = bounds.Dx()
			ps.height = bounds.Dy()
		}
	}
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

// GetBackgroundPictureSprite はピクチャIDに関連する背景PictureSpriteを取得する
// 背景PictureSpriteはOpenWin時に最初に作成されるPictureSprite
// CastSpriteやTextSpriteの親として使用される
// 要件 15.1: PictureSpriteは子スプライトを持てる
func (psm *PictureSpriteManager) GetBackgroundPictureSprite(picID int) *PictureSprite {
	psm.mu.RLock()
	defer psm.mu.RUnlock()

	sprites := psm.pictureSprites[picID]
	if len(sprites) == 0 {
		return nil
	}

	// 最初のPictureSpriteが背景
	return sprites[0]
}

// GetBackgroundPictureSpriteSprite はピクチャIDに関連する背景PictureSpriteの基盤スプライトを取得する
// CastSpriteやTextSpriteの親として使用する
func (psm *PictureSpriteManager) GetBackgroundPictureSpriteSprite(picID int) *Sprite {
	ps := psm.GetBackgroundPictureSprite(picID)
	if ps == nil {
		return nil
	}
	return ps.GetSprite()
}

// RemovePictureSprite は指定されたPictureSpriteを削除する
func (psm *PictureSpriteManager) RemovePictureSprite(ps *PictureSprite) {
	if ps == nil {
		return
	}

	psm.mu.Lock()
	defer psm.mu.Unlock()

	// スプライトを削除
	psm.spriteManager.DeleteSprite(ps.sprite.ID())

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
		psm.spriteManager.DeleteSprite(ps.sprite.ID())
	}
	delete(psm.pictureSprites, picID)
}

// Clear はすべてのPictureSpriteを削除する
func (psm *PictureSpriteManager) Clear() {
	psm.mu.Lock()
	defer psm.mu.Unlock()

	for picID, sprites := range psm.pictureSprites {
		for _, ps := range sprites {
			psm.spriteManager.DeleteSprite(ps.sprite.ID())
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

// SetVisible は可視性を更新する
func (ps *PictureSprite) SetVisible(visible bool) {
	ps.sprite.SetVisible(visible)
}

// SetParent は親スプライトを設定する
// ウインドウ内のピクチャ描画で使用
func (ps *PictureSprite) SetParent(parent *Sprite) {
	if parent != nil {
		parent.AddChild(ps.sprite)
	}
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

// IsEffectivelyVisible は実効的な可視性を返す
// 要件 14.2: PictureSpriteが「未関連付け」状態のとき、そのスプライトを描画しない
// 要件 14.3: PictureSpriteが「関連付け済み」状態のとき、親ウインドウの可視性に従って描画する
func (ps *PictureSprite) IsEffectivelyVisible() bool {
	// 未関連付け状態では描画しない
	if ps.state == PictureSpriteUnattached {
		return false
	}
	// 関連付け済み状態では親の可視性に従う
	return ps.sprite.IsEffectivelyVisible()
}

// GetState は現在の状態を返す
func (ps *PictureSprite) GetState() PictureSpriteState {
	return ps.state
}

// GetWindowID は関連付けられたウインドウIDを返す（-1 = 未関連付け）
func (ps *PictureSprite) GetWindowID() int {
	return ps.windowID
}
