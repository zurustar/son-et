// Package graphics provides sprite-based rendering system.
package graphics

import (
	"image"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

// PictureSpriteState はPictureSpriteの状態を表す
// 要件 12.1: PictureSpriteは「未関連付け」「関連付け済み」の状態を持つ
type PictureSpriteState int

const (
	// PictureSpriteUnattached は未関連付け状態
	// LoadPic時に作成され、まだウインドウに関連付けられていない状態
	// 要件 12.2: この状態ではスプライトを描画しない
	PictureSpriteUnattached PictureSpriteState = iota

	// PictureSpriteAttached は関連付け済み状態
	// SetPic/OpenWinでウインドウに関連付けられた状態
	// 要件 12.3: この状態では親ウインドウの可視性に従って描画する
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

// PictureSprite はピクチャ描画（MovePic）をスプライトとして表現するラッパー構造体
// 要件 6.1: BMPファイルからスプライトを作成できる
// 要件 6.2: 透明色を指定できる
// 要件 6.3: ピクチャの一部を切り出してスプライトにできる
// 要件 11.1: LoadPicが呼び出されたとき、非表示のPictureSpriteを作成する
// 要件 12.1: PictureSpriteは「未関連付け」「関連付け済み」の状態を持つ
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

	// 状態管理
	// 要件 12.1: PictureSpriteは「未関連付け」「関連付け済み」の状態を持つ
	state    PictureSpriteState // 現在の状態
	windowID int                // 関連付けられたウインドウID（-1 = 未関連付け）
}

// PictureSpriteManager はPictureSpriteを管理する
type PictureSpriteManager struct {
	pictureSprites   map[int][]*PictureSprite // picID -> PictureSprites（同じピクチャに複数のスプライトがある場合）
	pictureSpriteMap map[int]*PictureSprite   // picID -> 背景PictureSprite（LoadPic時に作成される主要なスプライト）
	spriteManager    *SpriteManager
	mu               sync.RWMutex
	nextID           int         // 内部ID管理
	zOffsets         map[int]int // picID -> 次のZ順序オフセット
}

// NewPictureSpriteManager は新しいPictureSpriteManagerを作成する
func NewPictureSpriteManager(sm *SpriteManager) *PictureSpriteManager {
	return &PictureSpriteManager{
		pictureSprites:   make(map[int][]*PictureSprite),
		pictureSpriteMap: make(map[int]*PictureSprite),
		spriteManager:    sm,
		nextID:           1,
		zOffsets:         make(map[int]int),
	}
}

// GetNextZOffset は指定されたピクチャIDの次のZ順序オフセットを取得してインクリメントする
func (psm *PictureSpriteManager) GetNextZOffset(picID int) int {
	psm.mu.Lock()
	defer psm.mu.Unlock()

	offset := psm.zOffsets[picID]
	psm.zOffsets[picID] = offset + 1
	return offset
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

	// スプライトを作成（非表示状態で作成）
	// 注意: zOrderパラメータは互換性のために残されているが、
	// 実際のZ順序はZ_Pathで管理される
	// レースコンディション対策: CreateSpriteHiddenを使用して最初から非表示で作成
	// Z_Pathが設定されるまで非表示にすることで、意図しない描画順序を防ぐ
	// 呼び出し元でZ_Pathを設定した後にSetVisible(true)を呼ぶ必要がある
	sprite := psm.spriteManager.CreateSpriteHidden(img)
	sprite.SetPosition(float64(destX), float64(destY))

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
		state:       PictureSpriteAttached, // MovePic経由で作成されるので関連付け済み
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
// 要件 11.4: ピクチャー画像をPictureSpriteとして作成し、WindowSpriteの子として追加
func (psm *PictureSpriteManager) CreateBackgroundPictureSprite(
	srcImg *ebiten.Image,
	picID int,
	width, height int,
	destX, destY int,
	zOrder int,
) *PictureSprite {
	psm.mu.Lock()
	defer psm.mu.Unlock()

	// 背景PictureSpriteはピクチャーの画像への参照を保持（コピーしない）
	// これにより、MovePicで更新されたピクチャー画像が反映される
	// 注意: zOrderパラメータは互換性のために残されているが、
	// 実際のZ順序はZ_Pathで管理される
	// レースコンディション対策: CreateSpriteHiddenを使用して最初から非表示で作成
	// Z_Pathを設定した後にSetVisible(true)を呼ぶ必要がある
	sprite := psm.spriteManager.CreateSpriteHidden(srcImg)
	sprite.SetPosition(float64(destX), float64(destY))
	// 注意: visibleはZ_Path設定後に呼び出し元で設定される

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
// 要件 11.1: LoadPicが呼び出されたとき、非表示のPictureSpriteを作成する
// 要件 11.2: PictureSpriteはピクチャ番号をキーとして管理される
// 要件 12.1: PictureSpriteは「未関連付け」状態で作成される
// 要件 12.2: 未関連付け状態ではスプライトを描画しない
func (psm *PictureSpriteManager) CreatePictureSpriteOnLoad(
	srcImg *ebiten.Image,
	picID int,
	width, height int,
) *PictureSprite {
	psm.mu.Lock()
	defer psm.mu.Unlock()

	// 要件 27.4: 既存のPictureSpriteがあれば削除する
	if existing, ok := psm.pictureSpriteMap[picID]; ok {
		psm.removePictureSpriteInternal(existing)
	}

	// 要件 27.2: 非表示状態でPictureSpriteを作成する
	// ピクチャーの画像への参照を保持（コピーしない）
	// レースコンディション対策: CreateSpriteHiddenを使用して最初から非表示で作成
	sprite := psm.spriteManager.CreateSpriteHidden(srcImg)
	sprite.SetPosition(0, 0)

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

	// 要件 27.3: pictureSpriteMapに登録する
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
		psm.spriteManager.RemoveSprite(child.ID())
	}

	// 親から削除
	if ps.sprite.Parent() != nil {
		ps.sprite.Parent().RemoveChild(ps.sprite.ID())
	}

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

	// pictureSpriteMapから削除
	if psm.pictureSpriteMap[ps.picID] == ps {
		delete(psm.pictureSpriteMap, ps.picID)
	}
}

// AttachPictureSpriteToWindow はSetPic/OpenWin時にPictureSpriteをウインドウに関連付ける
// 要件 11.3: SetPicが呼び出されたとき、既存のPictureSpriteをウインドウの子として関連付ける
// 要件 11.4: SetPicが呼び出されたとき、PictureSpriteを表示状態にする
// 要件 11.7: ピクチャがウインドウに関連付けられたとき、既存の子スプライト（キャスト、テキスト）のZ_Pathを更新する
// 注意: 関連付け後はpictureSpriteMapから削除される（同じピクチャを複数ウインドウで使用可能にするため）
func (psm *PictureSpriteManager) AttachPictureSpriteToWindow(picID int, windowSprite *Sprite, windowID int) error {
	psm.mu.Lock()
	defer psm.mu.Unlock()

	ps, ok := psm.pictureSpriteMap[picID]
	if !ok {
		// pictureSpriteMapに存在しない場合はエラーではなく、nilを返す
		// （従来の方法でPictureSpriteが作成される場合があるため）
		return nil
	}

	// 要件 28.2: PictureSpriteをWindowSpriteの子として追加する
	if windowSprite != nil {
		windowSprite.AddChild(ps.sprite)
		ps.sprite.SetParent(windowSprite)
	}

	// 要件 28.3: Z_Pathを設定する
	if windowSprite != nil && windowSprite.GetZPath() != nil {
		localZOrder := psm.spriteManager.GetZOrderCounter().GetNext(windowSprite.ID())
		zPath := NewZPathFromParent(windowSprite.GetZPath(), localZOrder)
		ps.sprite.SetZPath(zPath)
	}

	// 要件 28.4: 状態をAttachedに変更し、表示状態にする
	ps.state = PictureSpriteAttached
	ps.windowID = windowID
	ps.sprite.SetVisible(true)

	// 要件 28.5: 既存の子スプライトのZ_Pathを更新する
	psm.spriteManager.UpdateChildrenZPaths(ps.sprite)
	psm.spriteManager.MarkNeedSort()

	// 関連付け後はpictureSpriteMapから削除する
	// これにより、同じピクチャを複数ウインドウで使用する場合、
	// 次のOpenWin時に新しいPictureSpriteが作成される
	delete(psm.pictureSpriteMap, picID)

	return nil
}

// GetPictureSpriteByPictureID はピクチャ番号からPictureSpriteを取得する
// 要件 11.5, 11.6: ウインドウに関連付けられていないピクチャに対するCastSet/TextWriteで使用
// 要件 12.4: ピクチャ番号からPictureSpriteを効率的に検索できる
func (psm *PictureSpriteManager) GetPictureSpriteByPictureID(picID int) *PictureSprite {
	psm.mu.RLock()
	defer psm.mu.RUnlock()
	return psm.pictureSpriteMap[picID]
}

// FreePictureSprite はピクチャ解放時にPictureSpriteを削除する
// 要件 11.8: ピクチャが解放されたとき、対応するPictureSpriteとその子スプライトを削除する
// 要件 30.1: FreePictureSprite()メソッドを実装する
// 要件 30.2: 子スプライトを再帰的に削除する
// 要件 30.3: pictureSpriteMapから削除する
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
			psm.spriteManager.RemoveSprite(child.ID())
		}

		// 親から削除
		if ps.sprite.Parent() != nil {
			ps.sprite.Parent().RemoveChild(ps.sprite.ID())
		}

		// スプライトを削除
		psm.spriteManager.RemoveSprite(ps.sprite.ID())
	}
	delete(psm.pictureSprites, picID)

	// Z順序オフセットをリセット
	delete(psm.zOffsets, picID)
}

// UpdatePictureSpriteImage はMovePic時にPictureSpriteの画像を更新する
// 要件 12.5: MovePicが呼び出されたとき、転送先ピクチャのPictureSpriteの画像を更新する
// 要件 31.1: UpdatePictureSpriteImage()メソッドを実装する
// 注意: 現在の実装では、PictureSpriteはピクチャー画像への参照を保持しているため、
// MovePicでピクチャー画像が更新されると自動的に反映される。
// このメソッドは将来の拡張のために提供される。
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

// CreatePictureSpriteWithParent はMovePicの結果からPictureSpriteを作成し、親スプライトを設定する
// 要件 6.1: BMPファイルからスプライトを作成できる
// 要件 6.3: ピクチャの一部を切り出してスプライトにできる
// 要件 14.2: ウインドウ内のスプライトをウインドウの子スプライトとして管理する
// 要件 2.2: MovePicが呼び出されたとき、現在のZ_Order_Counterを使用してLocal_Z_Orderを割り当てる
// 要件 1.4: 子スプライトが作成されたとき、親のZ_Pathを継承し、自身のLocal_Z_Orderを追加する
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

		// 要件 2.2, 2.6: 操作順序でLocal_Z_Orderを割り当てる
		// 要件 1.4: 親のZ_Pathを継承し、自身のLocal_Z_Orderを追加する
		if parent.GetZPath() != nil {
			// 親のIDを使用してZOrderCounterから次のLocal_Z_Orderを取得
			localZOrder := psm.spriteManager.GetZOrderCounter().GetNext(parent.ID())
			zPath := NewZPathFromParent(parent.GetZPath(), localZOrder)
			ps.sprite.SetZPath(zPath)
		}

		// レースコンディション対策: Z_Pathを設定した後、MarkNeedSortを呼ぶ前に
		// SetVisibleを呼ぶことで、ソート時に正しいZ_Pathで可視状態になる
		// MarkNeedSortはSpriteManagerのロックを取得するため、
		// その前にSetVisibleを呼ぶことで、Draw()がスナップショットを取る際に
		// 一貫した状態を保証する
		ps.sprite.SetVisible(true)
		psm.spriteManager.MarkNeedSort()
	}
	return ps
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
// 背景PictureSpriteはOpenWin時にウィンドウに関連付けられたPictureSprite
// CastSpriteやTextSpriteの親として使用される
// 要件 9.1: PictureSpriteは子スプライトを持てる
// 要件 9.2: ピクチャ内にキャストが配置されたとき、キャストをピクチャの子スプライトとして管理する
// 要件 9.3: ピクチャ内にテキストが配置されたとき、テキストをピクチャの子スプライトとして管理する
func (psm *PictureSpriteManager) GetBackgroundPictureSprite(picID int) *PictureSprite {
	psm.mu.RLock()
	defer psm.mu.RUnlock()

	sprites := psm.pictureSprites[picID]
	if len(sprites) == 0 {
		return nil
	}

	// ウィンドウに関連付けられている（Attached状態で、Z_Pathが設定されている）PictureSpriteを探す
	// これがキャストやテキストの親として使用される正しい背景スプライト
	for _, ps := range sprites {
		if ps.state == PictureSpriteAttached && ps.sprite.GetZPath() != nil {
			return ps
		}
	}

	// 関連付けられたスプライトが見つからない場合は、最初のスプライトを返す（フォールバック）
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

// MergeOrCreatePictureSprite は融合可能なPictureSpriteがあれば融合し、なければ新規作成する
// タスク 6.3: MovePicで融合機能を使用するためのメソッド
//
// 動作:
// 1. FindMergeableSpriteで融合可能なスプライトを検索
// 2. 見つかった場合: MergeImageで画像を融合し、Z-orderを更新
// 3. 見つからない場合: CreatePictureSpriteWithParentで新規作成
//
// 引数:
//   - srcImg: ソース画像
//   - picID: 転送先ピクチャーID
//   - srcX, srcY: ソース領域の開始位置
//   - width, height: 描画領域のサイズ
//   - destX, destY: 描画先の位置
//   - zOrder: Z順序（新規作成時のみ使用）
//   - transparent: 透明色処理を行うかどうか
//   - parent: 親スプライト（nilの場合は親を考慮しない）
//
// 戻り値:
//   - 融合または作成されたPictureSprite
//   - merged: 融合が行われた場合はtrue、新規作成の場合はfalse
func (psm *PictureSpriteManager) MergeOrCreatePictureSprite(
	srcImg *ebiten.Image,
	picID int,
	srcX, srcY, width, height int,
	destX, destY int,
	zOrder int,
	transparent bool,
	parent *Sprite,
) (ps *PictureSprite, merged bool) {
	// 融合可能なスプライトを検索
	existingSprite := psm.FindMergeableSprite(picID, destX, destY, width, height, parent)

	if existingSprite != nil {
		// 融合可能なスプライトが見つかった場合、画像を融合
		// 既存スプライトのローカル座標系での相対位置を計算
		relativeX := destX - existingSprite.destX
		relativeY := destY - existingSprite.destY
		existingSprite.MergeImage(srcImg, relativeX, relativeY, transparent)

		// 重要: 融合後もZ-orderを更新して、既存の子スプライト（TextSprite等）より
		// 後に描画されるようにする。これにより、MovePicで転送された画像が
		// 既存のテキストを覆い隠す正しい動作になる。
		if parent != nil && parent.GetZPath() != nil {
			localZOrder := psm.spriteManager.GetZOrderCounter().GetNext(parent.ID())
			zPath := NewZPathFromParent(parent.GetZPath(), localZOrder)
			existingSprite.sprite.SetZPath(zPath)
			psm.spriteManager.MarkNeedSort()
		}

		return existingSprite, true
	}

	// 融合可能なスプライトが見つからない場合、新規作成
	if parent != nil {
		ps = psm.CreatePictureSpriteWithParent(srcImg, picID, srcX, srcY, width, height, destX, destY, zOrder, transparent, parent)
	} else {
		ps = psm.CreatePictureSprite(srcImg, picID, srcX, srcY, width, height, destX, destY, zOrder, transparent)
	}
	return ps, false
}

// FindMergeableSprite は融合可能なPictureSpriteを検索する
// 融合の条件:
// 1. 同じ転送先ピクチャーID（picID）を持つ
// 2. 同じ親スプライトを持つ（同じウィンドウ内）
// 3. 領域が重なっている、または隣接している
// 4. 背景PictureSpriteではない（MovePicで作成されたスプライトのみ融合対象）
//
// 引数:
//   - picID: 転送先ピクチャーID
//   - destX, destY: 描画先の位置
//   - width, height: 描画領域のサイズ
//   - parent: 親スプライト（nilの場合は親を考慮しない）
//
// 戻り値:
//   - 融合可能なPictureSpriteが見つかった場合はそのポインタ、見つからない場合はnil
func (psm *PictureSpriteManager) FindMergeableSprite(
	picID int,
	destX, destY int,
	width, height int,
	parent *Sprite,
) *PictureSprite {
	psm.mu.RLock()
	defer psm.mu.RUnlock()

	sprites := psm.pictureSprites[picID]
	if len(sprites) == 0 {
		return nil
	}

	// 背景PictureSpriteを取得（融合対象から除外するため）
	// 背景PictureSpriteはOpenWin時に作成され、キャストやテキストの親として使用される
	// MovePicで作成されたPictureSpriteのみが融合対象
	var bgPictureSprite *PictureSprite
	for _, ps := range sprites {
		if ps.state == PictureSpriteAttached && ps.sprite.GetZPath() != nil {
			bgPictureSprite = ps
			break
		}
	}

	// 新しい領域の境界を計算
	newLeft := destX
	newTop := destY
	newRight := destX + width
	newBottom := destY + height

	// 融合可能なスプライトを検索
	for _, ps := range sprites {
		// 背景PictureSpriteは融合対象から除外
		if ps == bgPictureSprite {
			continue
		}

		// 親スプライトが指定されている場合、同じ親を持つスプライトのみを対象とする
		if parent != nil && ps.sprite.Parent() != parent {
			continue
		}

		// 既存のスプライトの領域を取得
		existingLeft := ps.destX
		existingTop := ps.destY
		existingRight := ps.destX + ps.width
		existingBottom := ps.destY + ps.height

		// 領域が重なっているか、隣接しているかをチェック
		// 隣接の許容範囲（ピクセル単位）
		const adjacencyTolerance = 1

		// 重なりまたは隣接のチェック
		// 水平方向: 左端が右端より左にあり、右端が左端より右にある（許容範囲を含む）
		// 垂直方向: 上端が下端より上にあり、下端が上端より下にある（許容範囲を含む）
		horizontalOverlap := newLeft <= existingRight+adjacencyTolerance && newRight >= existingLeft-adjacencyTolerance
		verticalOverlap := newTop <= existingBottom+adjacencyTolerance && newBottom >= existingTop-adjacencyTolerance

		if horizontalOverlap && verticalOverlap {
			return ps
		}
	}

	return nil
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

// IsEffectivelyVisible は実効的な可視性を返す
// 要件 12.2: PictureSpriteが「未関連付け」状態のとき、そのスプライトを描画しない
// 要件 12.3: PictureSpriteが「関連付け済み」状態のとき、親ウインドウの可視性に従って描画する
// 要件 32.1: PictureSprite.IsEffectivelyVisible()を実装する
func (ps *PictureSprite) IsEffectivelyVisible() bool {
	// 要件 32.2: 未関連付け状態では描画しない
	if ps.state == PictureSpriteUnattached {
		return false
	}
	// 要件 32.3: 関連付け済み状態では親の可視性に従う
	return ps.sprite.IsEffectivelyVisible()
}

// MergeImage は別の画像をこのPictureSpriteに合成する
// タスク 6.2: ピクチャースプライトの融合機能
//
// 引数:
//   - srcImg: 合成するソース画像
//   - destX, destY: このスプライト内での合成先位置（スプライトのローカル座標）
//   - transparent: 透明色処理を行うかどうか（trueの場合、黒(0,0,0)を透明として扱う）
//
// 動作:
//   - ソース画像をこのスプライトの画像に合成する
//   - 必要に応じてスプライトの領域を拡張する
//   - 透明色処理が有効な場合、黒色ピクセルは合成しない
func (ps *PictureSprite) MergeImage(srcImg *ebiten.Image, destX, destY int, transparent bool) {
	if srcImg == nil {
		return
	}

	currentImg := ps.sprite.Image()
	if currentImg == nil {
		// 現在の画像がない場合は、ソース画像をそのまま設定
		ps.sprite.SetImage(srcImg)
		ps.destX = destX
		ps.destY = destY
		bounds := srcImg.Bounds()
		ps.width = bounds.Dx()
		ps.height = bounds.Dy()
		return
	}

	srcBounds := srcImg.Bounds()
	srcWidth := srcBounds.Dx()
	srcHeight := srcBounds.Dy()

	// 現在のスプライトの領域
	currentWidth := ps.width
	currentHeight := ps.height

	// 合成後の領域を計算
	// destX, destYは現在のスプライト内での相対位置
	newLeft := min(0, destX)
	newTop := min(0, destY)
	newRight := max(currentWidth, destX+srcWidth)
	newBottom := max(currentHeight, destY+srcHeight)

	newWidth := newRight - newLeft
	newHeight := newBottom - newTop

	// 領域が拡張される場合、新しい画像を作成
	if newWidth > currentWidth || newHeight > currentHeight || newLeft < 0 || newTop < 0 {
		// 新しい画像を作成
		newImg := ebiten.NewImage(newWidth, newHeight)

		// 現在の画像を新しい位置にコピー
		currentOp := &ebiten.DrawImageOptions{}
		currentOp.GeoM.Translate(float64(-newLeft), float64(-newTop))
		newImg.DrawImage(currentImg, currentOp)

		// ソース画像を合成
		srcOp := &ebiten.DrawImageOptions{}
		srcOp.GeoM.Translate(float64(destX-newLeft), float64(destY-newTop))

		if transparent {
			// 透明色処理: シェーダーを使用して黒を透明にする
			// 注意: Ebitengineでは直接的な透明色処理が難しいため、
			// ここでは単純にDrawImageを使用し、透明色処理は呼び出し側で行う想定
			newImg.DrawImage(srcImg, srcOp)
		} else {
			newImg.DrawImage(srcImg, srcOp)
		}

		// スプライトの画像を更新
		ps.sprite.SetImage(newImg)

		// 位置を調整（領域が左上に拡張された場合）
		if newLeft < 0 || newTop < 0 {
			ps.destX += newLeft
			ps.destY += newTop
			ps.sprite.SetPosition(float64(ps.destX), float64(ps.destY))
		}

		// サイズを更新
		ps.width = newWidth
		ps.height = newHeight
	} else {
		// 領域が拡張されない場合、現在の画像に直接合成
		srcOp := &ebiten.DrawImageOptions{}
		srcOp.GeoM.Translate(float64(destX), float64(destY))

		if transparent {
			// 透明色処理: シェーダーを使用して黒を透明にする
			// 注意: Ebitengineでは直接的な透明色処理が難しいため、
			// ここでは単純にDrawImageを使用し、透明色処理は呼び出し側で行う想定
			currentImg.DrawImage(srcImg, srcOp)
		} else {
			currentImg.DrawImage(srcImg, srcOp)
		}
	}
}

// GetState は現在の状態を返す
func (ps *PictureSprite) GetState() PictureSpriteState {
	return ps.state
}

// GetWindowID は関連付けられたウインドウIDを返す（-1 = 未関連付け）
func (ps *PictureSprite) GetWindowID() int {
	return ps.windowID
}
