package graphics

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// TestNewPictureSpriteManager はPictureSpriteManagerの作成をテストする
func TestNewPictureSpriteManager(t *testing.T) {
	sm := NewSpriteManager()
	psm := NewPictureSpriteManager(sm)

	if psm == nil {
		t.Fatal("NewPictureSpriteManager returned nil")
	}
	if psm.spriteManager != sm {
		t.Error("SpriteManager not set correctly")
	}
	if len(psm.pictureSprites) != 0 {
		t.Errorf("Expected empty pictureSprites map, got %d", len(psm.pictureSprites))
	}
	if psm.Count() != 0 {
		t.Errorf("Expected count 0, got %d", psm.Count())
	}
}

// TestCreatePictureSprite はPictureSpriteの作成をテストする
// 要件 6.1: BMPファイルからスプライトを作成できる
// 要件 6.3: ピクチャの一部を切り出してスプライトにできる
func TestCreatePictureSprite(t *testing.T) {
	sm := NewSpriteManager()
	psm := NewPictureSpriteManager(sm)

	// テスト用の画像を作成
	srcImg := ebiten.NewImage(100, 100)

	ps := psm.CreatePictureSprite(
		srcImg,
		1,      // picID
		10, 20, // srcX, srcY
		50, 60, // width, height
		100, 150, // destX, destY
		5,     // zOrder
		false, // transparent
	)

	if ps == nil {
		t.Fatal("CreatePictureSprite returned nil")
	}

	// 属性を確認
	if ps.GetPicID() != 1 {
		t.Errorf("Expected picID 1, got %d", ps.GetPicID())
	}
	if ps.GetSrcX() != 10 {
		t.Errorf("Expected srcX 10, got %d", ps.GetSrcX())
	}
	if ps.GetSrcY() != 20 {
		t.Errorf("Expected srcY 20, got %d", ps.GetSrcY())
	}
	if ps.GetWidth() != 50 {
		t.Errorf("Expected width 50, got %d", ps.GetWidth())
	}
	if ps.GetHeight() != 60 {
		t.Errorf("Expected height 60, got %d", ps.GetHeight())
	}
	if ps.GetDestX() != 100 {
		t.Errorf("Expected destX 100, got %d", ps.GetDestX())
	}
	if ps.GetDestY() != 150 {
		t.Errorf("Expected destY 150, got %d", ps.GetDestY())
	}
	if ps.IsTransparent() {
		t.Error("Expected transparent to be false")
	}

	// スプライトの属性を確認
	sprite := ps.GetSprite()
	if sprite == nil {
		t.Fatal("GetSprite returned nil")
	}

	x, y := sprite.Position()
	if x != 100 || y != 150 {
		t.Errorf("Expected position (100, 150), got (%v, %v)", x, y)
	}

	// Z_Pathはまだ設定されていない（親スプライトなしで作成されたため）
	// zOrderパラメータは互換性のために残されているが、実際のZ順序はZ_Pathで管理される

	if !sprite.Visible() {
		t.Error("Expected sprite to be visible")
	}

	// カウントを確認
	if psm.Count() != 1 {
		t.Errorf("Expected count 1, got %d", psm.Count())
	}
}

// TestCreatePictureSpriteFromDrawingEntry はDrawingEntryからPictureSpriteを作成するテスト
func TestCreatePictureSpriteFromDrawingEntry(t *testing.T) {
	sm := NewSpriteManager()
	psm := NewPictureSpriteManager(sm)

	// テスト用のDrawingEntryを作成
	img := ebiten.NewImage(50, 60)
	entry := NewDrawingEntry(1, 2, img, 100, 150, 50, 60, 0)

	ps := psm.CreatePictureSpriteFromDrawingEntry(entry, 10)

	if ps == nil {
		t.Fatal("CreatePictureSpriteFromDrawingEntry returned nil")
	}

	// 属性を確認
	if ps.GetPicID() != 2 {
		t.Errorf("Expected picID 2, got %d", ps.GetPicID())
	}
	if ps.GetWidth() != 50 {
		t.Errorf("Expected width 50, got %d", ps.GetWidth())
	}
	if ps.GetHeight() != 60 {
		t.Errorf("Expected height 60, got %d", ps.GetHeight())
	}
	if ps.GetDestX() != 100 {
		t.Errorf("Expected destX 100, got %d", ps.GetDestX())
	}
	if ps.GetDestY() != 150 {
		t.Errorf("Expected destY 150, got %d", ps.GetDestY())
	}

	// Z_Pathはまだ設定されていない（親スプライトなしで作成されたため）
	// zOrderパラメータは互換性のために残されているが、実際のZ順序はZ_Pathで管理される
}

// TestCreatePictureSpriteFromDrawingEntryNil はnilのDrawingEntryを渡した場合のテスト
func TestCreatePictureSpriteFromDrawingEntryNil(t *testing.T) {
	sm := NewSpriteManager()
	psm := NewPictureSpriteManager(sm)

	ps := psm.CreatePictureSpriteFromDrawingEntry(nil, 10)

	if ps != nil {
		t.Error("Expected nil for nil DrawingEntry")
	}
}

// TestGetPictureSprites はピクチャIDに関連するPictureSpriteの取得をテストする
func TestGetPictureSprites(t *testing.T) {
	sm := NewSpriteManager()
	psm := NewPictureSpriteManager(sm)

	// 同じピクチャIDで複数のPictureSpriteを作成
	srcImg := ebiten.NewImage(50, 50)
	psm.CreatePictureSprite(srcImg, 1, 0, 0, 50, 50, 0, 0, 0, false)
	psm.CreatePictureSprite(srcImg, 1, 0, 0, 50, 50, 100, 0, 1, false)
	psm.CreatePictureSprite(srcImg, 2, 0, 0, 50, 50, 0, 100, 2, false)

	// ピクチャID 1のスプライトを取得
	sprites := psm.GetPictureSprites(1)
	if len(sprites) != 2 {
		t.Errorf("Expected 2 sprites for picID 1, got %d", len(sprites))
	}

	// ピクチャID 2のスプライトを取得
	sprites = psm.GetPictureSprites(2)
	if len(sprites) != 1 {
		t.Errorf("Expected 1 sprite for picID 2, got %d", len(sprites))
	}

	// 存在しないピクチャIDのスプライトを取得
	sprites = psm.GetPictureSprites(999)
	if sprites != nil {
		t.Error("Expected nil for non-existing picID")
	}
}

// TestRemovePictureSprite はPictureSpriteの削除をテストする
func TestRemovePictureSprite(t *testing.T) {
	sm := NewSpriteManager()
	psm := NewPictureSpriteManager(sm)

	srcImg := ebiten.NewImage(50, 50)
	ps := psm.CreatePictureSprite(srcImg, 1, 0, 0, 50, 50, 0, 0, 0, false)
	spriteID := ps.GetSprite().ID()

	// スプライトが存在することを確認
	if sm.GetSprite(spriteID) == nil {
		t.Error("Sprite should exist before removal")
	}

	// PictureSpriteを削除
	psm.RemovePictureSprite(ps)

	// PictureSpriteが削除されたことを確認
	sprites := psm.GetPictureSprites(1)
	if len(sprites) != 0 {
		t.Errorf("Expected 0 sprites after removal, got %d", len(sprites))
	}

	// スプライトも削除されたことを確認
	if sm.GetSprite(spriteID) != nil {
		t.Error("Sprite should be removed from SpriteManager")
	}

	// カウントを確認
	if psm.Count() != 0 {
		t.Errorf("Expected count 0, got %d", psm.Count())
	}
}

// TestRemovePictureSpriteNil はnilのPictureSpriteを削除した場合のテスト
func TestRemovePictureSpriteNil(t *testing.T) {
	sm := NewSpriteManager()
	psm := NewPictureSpriteManager(sm)

	// nilを削除してもパニックしないことを確認
	psm.RemovePictureSprite(nil)
}

// TestRemovePictureSpritesByPicID はピクチャIDに関連するすべてのPictureSpriteの削除をテストする
func TestRemovePictureSpritesByPicID(t *testing.T) {
	sm := NewSpriteManager()
	psm := NewPictureSpriteManager(sm)

	srcImg := ebiten.NewImage(50, 50)
	psm.CreatePictureSprite(srcImg, 1, 0, 0, 50, 50, 0, 0, 0, false)
	psm.CreatePictureSprite(srcImg, 1, 0, 0, 50, 50, 100, 0, 1, false)
	psm.CreatePictureSprite(srcImg, 2, 0, 0, 50, 50, 0, 100, 2, false)

	// ピクチャID 1のスプライトを削除
	psm.RemovePictureSpritesByPicID(1)

	// ピクチャID 1のスプライトが削除されたことを確認
	sprites := psm.GetPictureSprites(1)
	if len(sprites) != 0 {
		t.Errorf("Expected 0 sprites for picID 1, got %d", len(sprites))
	}

	// ピクチャID 2のスプライトは残っていることを確認
	sprites = psm.GetPictureSprites(2)
	if len(sprites) != 1 {
		t.Errorf("Expected 1 sprite for picID 2, got %d", len(sprites))
	}

	// カウントを確認
	if psm.Count() != 1 {
		t.Errorf("Expected count 1, got %d", psm.Count())
	}
}

// TestPictureSpriteManagerClear はすべてのPictureSpriteの削除をテストする
func TestPictureSpriteManagerClear(t *testing.T) {
	sm := NewSpriteManager()
	psm := NewPictureSpriteManager(sm)

	srcImg := ebiten.NewImage(50, 50)
	for i := 1; i <= 3; i++ {
		psm.CreatePictureSprite(srcImg, i, 0, 0, 50, 50, i*100, 0, i, false)
	}

	// すべてのPictureSpriteをクリア
	psm.Clear()

	// すべてのPictureSpriteが削除されたことを確認
	for i := 1; i <= 3; i++ {
		sprites := psm.GetPictureSprites(i)
		if len(sprites) != 0 {
			t.Errorf("Expected 0 sprites for picID %d after Clear, got %d", i, len(sprites))
		}
	}

	// カウントを確認
	if psm.Count() != 0 {
		t.Errorf("Expected count 0, got %d", psm.Count())
	}
}

// TestPictureSpriteSetPosition は位置の更新をテストする
func TestPictureSpriteSetPosition(t *testing.T) {
	sm := NewSpriteManager()
	psm := NewPictureSpriteManager(sm)

	srcImg := ebiten.NewImage(50, 50)
	ps := psm.CreatePictureSprite(srcImg, 1, 0, 0, 50, 50, 0, 0, 0, false)

	// 位置を更新
	ps.SetPosition(200, 300)

	// PictureSpriteの位置が更新されたことを確認
	if ps.GetDestX() != 200 || ps.GetDestY() != 300 {
		t.Errorf("Expected position (200, 300), got (%d, %d)", ps.GetDestX(), ps.GetDestY())
	}

	// スプライトの位置が更新されたことを確認
	x, y := ps.GetSprite().Position()
	if x != 200 || y != 300 {
		t.Errorf("Expected sprite position (200, 300), got (%v, %v)", x, y)
	}
}

// TestPictureSpriteSetVisible は可視性の更新をテストする
func TestPictureSpriteSetVisible(t *testing.T) {
	sm := NewSpriteManager()
	psm := NewPictureSpriteManager(sm)

	srcImg := ebiten.NewImage(50, 50)
	ps := psm.CreatePictureSprite(srcImg, 1, 0, 0, 50, 50, 0, 0, 0, false)

	// 可視性を更新
	ps.SetVisible(false)

	// スプライトの可視性が更新されたことを確認
	if ps.GetSprite().Visible() {
		t.Error("Expected sprite to be invisible")
	}
}

// TestPictureSpriteSetParent は親スプライトの設定をテストする
func TestPictureSpriteSetParent(t *testing.T) {
	sm := NewSpriteManager()
	psm := NewPictureSpriteManager(sm)

	// 親スプライトを作成
	parentImg := ebiten.NewImage(200, 200)
	parent := sm.CreateSprite(parentImg)
	parent.SetPosition(100, 50)

	// PictureSpriteを作成
	srcImg := ebiten.NewImage(50, 50)
	ps := psm.CreatePictureSprite(srcImg, 1, 0, 0, 50, 50, 10, 20, 0, false)

	// 親を設定
	ps.SetParent(parent)

	// 親が設定されていることを確認
	if ps.GetSprite().Parent() != parent {
		t.Error("Parent should be set")
	}

	// 絶対位置を確認
	// 親の位置(100, 50) + 子の相対位置(10, 20) = (110, 70)
	absX, absY := ps.GetSprite().AbsolutePosition()
	if absX != 110 || absY != 70 {
		t.Errorf("Expected absolute position (110, 70), got (%v, %v)", absX, absY)
	}
}

// TestPictureSpriteUpdateImage は画像の更新をテストする
func TestPictureSpriteUpdateImage(t *testing.T) {
	sm := NewSpriteManager()
	psm := NewPictureSpriteManager(sm)

	srcImg := ebiten.NewImage(50, 50)
	ps := psm.CreatePictureSprite(srcImg, 1, 0, 0, 50, 50, 0, 0, 0, false)

	// 新しい画像を作成
	newImg := ebiten.NewImage(100, 80)

	// 画像を更新
	ps.UpdateImage(newImg)

	// スプライトの画像が更新されたことを確認
	if ps.GetSprite().Image() != newImg {
		t.Error("Image should be updated")
	}

	// サイズが更新されたことを確認
	if ps.GetWidth() != 100 || ps.GetHeight() != 80 {
		t.Errorf("Expected size (100, 80), got (%d, %d)", ps.GetWidth(), ps.GetHeight())
	}
}

// TestGraphicsSystemPictureSpriteIntegration はGraphicsSystemとの統合をテストする
func TestGraphicsSystemPictureSpriteIntegration(t *testing.T) {
	gs := NewGraphicsSystem("")

	// PictureSpriteManagerが初期化されていることを確認
	psm := gs.GetPictureSpriteManager()
	if psm == nil {
		t.Fatal("PictureSpriteManager is nil")
	}

	// 初期状態ではPictureSpriteがないことを確認
	if psm.Count() != 0 {
		t.Errorf("Expected count 0, got %d", psm.Count())
	}
}

// TestCreatePictureSpriteOnLoad はLoadPic時のPictureSprite作成をテストする
// 要件 11.1: LoadPicが呼び出されたとき、非表示のPictureSpriteを作成する
// 要件 11.2: PictureSpriteはピクチャ番号をキーとして管理される
// 要件 12.1: PictureSpriteは「未関連付け」状態で作成される
// 要件 12.2: 未関連付け状態ではスプライトを描画しない
func TestCreatePictureSpriteOnLoad(t *testing.T) {
	sm := NewSpriteManager()
	psm := NewPictureSpriteManager(sm)

	// テスト用の画像を作成
	srcImg := ebiten.NewImage(100, 100)

	// LoadPic時にPictureSpriteを作成
	ps := psm.CreatePictureSpriteOnLoad(srcImg, 35, 100, 100)

	if ps == nil {
		t.Fatal("CreatePictureSpriteOnLoad returned nil")
	}

	// 要件 11.2: ピクチャ番号をキーとして管理される
	retrievedPS := psm.GetPictureSpriteByPictureID(35)
	if retrievedPS != ps {
		t.Error("GetPictureSpriteByPictureID should return the created PictureSprite")
	}

	// 要件 12.1: 未関連付け状態で作成される
	if ps.GetState() != PictureSpriteUnattached {
		t.Errorf("Expected state Unattached, got %v", ps.GetState())
	}

	// 要件 12.2: 未関連付け状態では描画しない
	if ps.IsEffectivelyVisible() {
		t.Error("Unattached PictureSprite should not be effectively visible")
	}

	// スプライトが非表示であることを確認
	if ps.GetSprite().Visible() {
		t.Error("Sprite should be invisible when unattached")
	}

	// ウインドウIDが-1（未関連付け）であることを確認
	if ps.GetWindowID() != -1 {
		t.Errorf("Expected windowID -1, got %d", ps.GetWindowID())
	}
}

// TestAttachPictureSpriteToWindow はSetPic時の関連付けをテストする
// 要件 11.3: SetPicが呼び出されたとき、既存のPictureSpriteをウインドウの子として関連付ける
// 要件 11.4: SetPicが呼び出されたとき、PictureSpriteを表示状態にする
func TestAttachPictureSpriteToWindow(t *testing.T) {
	sm := NewSpriteManager()
	psm := NewPictureSpriteManager(sm)

	// テスト用の画像を作成
	srcImg := ebiten.NewImage(100, 100)

	// LoadPic時にPictureSpriteを作成
	ps := psm.CreatePictureSpriteOnLoad(srcImg, 35, 100, 100)

	// ウインドウスプライトを作成（親として使用）
	windowImg := ebiten.NewImage(640, 480)
	windowSprite := sm.CreateSprite(windowImg)
	windowSprite.SetZPath(NewZPath(0))
	windowSprite.SetVisible(true)

	// PictureSpriteをウインドウに関連付け
	err := psm.AttachPictureSpriteToWindow(35, windowSprite, 0)
	if err != nil {
		t.Fatalf("AttachPictureSpriteToWindow failed: %v", err)
	}

	// 要件 11.3: ウインドウの子として関連付けられる
	if ps.GetSprite().Parent() != windowSprite {
		t.Error("PictureSprite should be a child of WindowSprite")
	}

	// 要件 11.4: 表示状態になる
	if !ps.GetSprite().Visible() {
		t.Error("PictureSprite should be visible after attachment")
	}

	// 状態がAttachedに変更される
	if ps.GetState() != PictureSpriteAttached {
		t.Errorf("Expected state Attached, got %v", ps.GetState())
	}

	// ウインドウIDが設定される
	if ps.GetWindowID() != 0 {
		t.Errorf("Expected windowID 0, got %d", ps.GetWindowID())
	}

	// Z_Pathが設定される
	if ps.GetSprite().GetZPath() == nil {
		t.Error("Z_Path should be set after attachment")
	}

	// 関連付け後はpictureSpriteMapから削除される
	if psm.GetPictureSpriteByPictureID(35) != nil {
		t.Error("PictureSprite should be removed from pictureSpriteMap after attachment")
	}
}

// TestFreePictureSprite はFreePic時の削除をテストする
// 要件 11.8: ピクチャが解放されたとき、対応するPictureSpriteとその子スプライトを削除する
func TestFreePictureSprite(t *testing.T) {
	sm := NewSpriteManager()
	psm := NewPictureSpriteManager(sm)

	// テスト用の画像を作成
	srcImg := ebiten.NewImage(100, 100)

	// LoadPic時にPictureSpriteを作成
	ps := psm.CreatePictureSpriteOnLoad(srcImg, 35, 100, 100)
	spriteID := ps.GetSprite().ID()

	// スプライトが存在することを確認
	if sm.GetSprite(spriteID) == nil {
		t.Error("Sprite should exist before FreePictureSprite")
	}

	// PictureSpriteを削除
	psm.FreePictureSprite(35)

	// pictureSpriteMapから削除されたことを確認
	if psm.GetPictureSpriteByPictureID(35) != nil {
		t.Error("PictureSprite should be removed from pictureSpriteMap")
	}

	// スプライトも削除されたことを確認
	if sm.GetSprite(spriteID) != nil {
		t.Error("Sprite should be removed from SpriteManager")
	}
}

// TestPictureSpriteIsEffectivelyVisible は実効的な可視性をテストする
// 要件 12.2: PictureSpriteが「未関連付け」状態のとき、そのスプライトを描画しない
// 要件 12.3: PictureSpriteが「関連付け済み」状態のとき、親ウインドウの可視性に従って描画する
func TestPictureSpriteIsEffectivelyVisible(t *testing.T) {
	sm := NewSpriteManager()
	psm := NewPictureSpriteManager(sm)

	// テスト用の画像を作成
	srcImg := ebiten.NewImage(100, 100)

	// LoadPic時にPictureSpriteを作成（未関連付け状態）
	ps := psm.CreatePictureSpriteOnLoad(srcImg, 35, 100, 100)

	// 要件 12.2: 未関連付け状態では描画しない
	if ps.IsEffectivelyVisible() {
		t.Error("Unattached PictureSprite should not be effectively visible")
	}

	// ウインドウスプライトを作成（親として使用）
	windowImg := ebiten.NewImage(640, 480)
	windowSprite := sm.CreateSprite(windowImg)
	windowSprite.SetZPath(NewZPath(0))
	windowSprite.SetVisible(true)

	// PictureSpriteをウインドウに関連付け
	psm.AttachPictureSpriteToWindow(35, windowSprite, 0)

	// 要件 12.3: 関連付け済み状態では親の可視性に従う
	if !ps.IsEffectivelyVisible() {
		t.Error("Attached PictureSprite with visible parent should be effectively visible")
	}

	// 親を非表示にする
	windowSprite.SetVisible(false)

	// 親が非表示なので、子も非表示になる
	if ps.IsEffectivelyVisible() {
		t.Error("PictureSprite with invisible parent should not be effectively visible")
	}
}

// TestPictureSpriteWithChildren は子スプライトを持つPictureSpriteの削除をテストする
func TestPictureSpriteWithChildren(t *testing.T) {
	sm := NewSpriteManager()
	psm := NewPictureSpriteManager(sm)

	// テスト用の画像を作成
	srcImg := ebiten.NewImage(100, 100)

	// LoadPic時にPictureSpriteを作成
	ps := psm.CreatePictureSpriteOnLoad(srcImg, 35, 100, 100)

	// 子スプライトを追加
	childImg := ebiten.NewImage(50, 50)
	child := sm.CreateSprite(childImg)
	ps.GetSprite().AddChild(child)

	childID := child.ID()

	// 子スプライトが存在することを確認
	if sm.GetSprite(childID) == nil {
		t.Error("Child sprite should exist before FreePictureSprite")
	}

	// PictureSpriteを削除（子スプライトも削除される）
	psm.FreePictureSprite(35)

	// 子スプライトも削除されたことを確認
	if sm.GetSprite(childID) != nil {
		t.Error("Child sprite should be removed when parent PictureSprite is freed")
	}
}
