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

	// レースコンディション対策: CreatePictureSpriteは非表示でスプライトを作成する
	// Z_Pathが設定された後にSetVisible(true)を呼ぶ必要がある
	if sprite.Visible() {
		t.Error("Expected sprite to be hidden (race condition prevention)")
	}

	// カウントを確認
	if psm.Count() != 1 {
		t.Errorf("Expected count 1, got %d", psm.Count())
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

// TestFindMergeableSprite は融合可能なPictureSpriteの検索をテストする
func TestFindMergeableSprite(t *testing.T) {
	sm := NewSpriteManager()
	psm := NewPictureSpriteManager(sm)

	// 親スプライトを作成
	parentImg := ebiten.NewImage(640, 480)
	parent := sm.CreateSprite(parentImg)

	// テスト用の画像を作成
	srcImg := ebiten.NewImage(100, 100)

	// PictureSpriteを作成（位置: 100, 100、サイズ: 50x50）
	ps := psm.CreatePictureSprite(srcImg, 1, 0, 0, 50, 50, 100, 100, 0, false)
	ps.SetParent(parent)

	t.Run("重なる領域で融合可能なスプライトを見つける", func(t *testing.T) {
		// 重なる領域（120, 120）で検索
		found := psm.FindMergeableSprite(1, 120, 120, 30, 30, parent)
		if found != ps {
			t.Error("Expected to find the existing PictureSprite for overlapping region")
		}
	})

	t.Run("隣接する領域で融合可能なスプライトを見つける", func(t *testing.T) {
		// 隣接する領域（150, 100）で検索（右隣）
		found := psm.FindMergeableSprite(1, 150, 100, 30, 30, parent)
		if found != ps {
			t.Error("Expected to find the existing PictureSprite for adjacent region")
		}
	})

	t.Run("離れた領域では融合可能なスプライトを見つけない", func(t *testing.T) {
		// 離れた領域（300, 300）で検索
		found := psm.FindMergeableSprite(1, 300, 300, 30, 30, parent)
		if found != nil {
			t.Error("Expected not to find any PictureSprite for distant region")
		}
	})

	t.Run("異なるピクチャIDでは融合可能なスプライトを見つけない", func(t *testing.T) {
		// 異なるピクチャID（2）で検索
		found := psm.FindMergeableSprite(2, 100, 100, 30, 30, parent)
		if found != nil {
			t.Error("Expected not to find any PictureSprite for different picID")
		}
	})

	t.Run("異なる親スプライトでは融合可能なスプライトを見つけない", func(t *testing.T) {
		// 別の親スプライトを作成
		otherParentImg := ebiten.NewImage(640, 480)
		otherParent := sm.CreateSprite(otherParentImg)

		// 異なる親で検索
		found := psm.FindMergeableSprite(1, 100, 100, 30, 30, otherParent)
		if found != nil {
			t.Error("Expected not to find any PictureSprite for different parent")
		}
	})

	t.Run("親がnilの場合は親を考慮しない", func(t *testing.T) {
		// 親がnilで検索（親を考慮しない）
		found := psm.FindMergeableSprite(1, 100, 100, 30, 30, nil)
		if found != ps {
			t.Error("Expected to find the existing PictureSprite when parent is nil")
		}
	})
}

// TestFindMergeableSpriteEdgeCases はFindMergeableSpriteのエッジケースをテストする
func TestFindMergeableSpriteEdgeCases(t *testing.T) {
	sm := NewSpriteManager()
	psm := NewPictureSpriteManager(sm)

	t.Run("空のマネージャーでは何も見つからない", func(t *testing.T) {
		found := psm.FindMergeableSprite(1, 0, 0, 50, 50, nil)
		if found != nil {
			t.Error("Expected nil for empty manager")
		}
	})

	t.Run("存在しないピクチャIDでは何も見つからない", func(t *testing.T) {
		// PictureSpriteを作成
		srcImg := ebiten.NewImage(100, 100)
		psm.CreatePictureSprite(srcImg, 1, 0, 0, 50, 50, 0, 0, 0, false)

		// 存在しないピクチャIDで検索
		found := psm.FindMergeableSprite(999, 0, 0, 50, 50, nil)
		if found != nil {
			t.Error("Expected nil for non-existing picID")
		}
	})

	t.Run("完全に一致する領域で融合可能なスプライトを見つける", func(t *testing.T) {
		srcImg := ebiten.NewImage(100, 100)
		ps := psm.CreatePictureSprite(srcImg, 2, 0, 0, 50, 50, 200, 200, 0, false)

		// 完全に一致する領域で検索
		found := psm.FindMergeableSprite(2, 200, 200, 50, 50, nil)
		if found != ps {
			t.Error("Expected to find the existing PictureSprite for exact match")
		}
	})

	t.Run("境界ぎりぎりで隣接する領域", func(t *testing.T) {
		srcImg := ebiten.NewImage(100, 100)
		ps := psm.CreatePictureSprite(srcImg, 3, 0, 0, 50, 50, 0, 0, 0, false)

		// 境界ぎりぎりで隣接（1ピクセル離れている）
		found := psm.FindMergeableSprite(3, 51, 0, 50, 50, nil)
		if found != ps {
			t.Error("Expected to find the existing PictureSprite for boundary adjacent region")
		}
	})

	t.Run("境界を超えて離れている領域", func(t *testing.T) {
		srcImg := ebiten.NewImage(100, 100)
		psm.CreatePictureSprite(srcImg, 4, 0, 0, 50, 50, 0, 0, 0, false)

		// 境界を超えて離れている（2ピクセル以上離れている）
		found := psm.FindMergeableSprite(4, 52, 0, 50, 50, nil)
		if found != nil {
			t.Error("Expected nil for region beyond adjacency tolerance")
		}
	})
}

// TestPictureSpriteMergeImage はMergeImage()メソッドをテストする
// タスク 6.2: PictureSpriteにMergeImage()を追加する
func TestPictureSpriteMergeImage(t *testing.T) {
	sm := NewSpriteManager()
	psm := NewPictureSpriteManager(sm)

	t.Run("基本的な画像合成", func(t *testing.T) {
		// 50x50の画像を持つPictureSpriteを作成
		srcImg := ebiten.NewImage(50, 50)
		ps := psm.CreatePictureSprite(srcImg, 1, 0, 0, 50, 50, 0, 0, 0, false)

		// 30x30の画像を位置(10, 10)に合成
		mergeImg := ebiten.NewImage(30, 30)
		ps.MergeImage(mergeImg, 10, 10, false)

		// サイズは変わらない（合成領域が元の領域内に収まる）
		if ps.GetWidth() != 50 || ps.GetHeight() != 50 {
			t.Errorf("Expected size (50, 50), got (%d, %d)", ps.GetWidth(), ps.GetHeight())
		}
	})

	t.Run("領域拡張を伴う画像合成（右下方向）", func(t *testing.T) {
		// 50x50の画像を持つPictureSpriteを作成
		srcImg := ebiten.NewImage(50, 50)
		ps := psm.CreatePictureSprite(srcImg, 2, 0, 0, 50, 50, 100, 100, 0, false)

		// 30x30の画像を位置(40, 40)に合成（右下に拡張）
		mergeImg := ebiten.NewImage(30, 30)
		ps.MergeImage(mergeImg, 40, 40, false)

		// サイズが拡張される（40+30=70）
		if ps.GetWidth() != 70 || ps.GetHeight() != 70 {
			t.Errorf("Expected size (70, 70), got (%d, %d)", ps.GetWidth(), ps.GetHeight())
		}

		// 位置は変わらない
		if ps.GetDestX() != 100 || ps.GetDestY() != 100 {
			t.Errorf("Expected position (100, 100), got (%d, %d)", ps.GetDestX(), ps.GetDestY())
		}
	})

	t.Run("領域拡張を伴う画像合成（左上方向）", func(t *testing.T) {
		// 50x50の画像を持つPictureSpriteを作成
		srcImg := ebiten.NewImage(50, 50)
		ps := psm.CreatePictureSprite(srcImg, 3, 0, 0, 50, 50, 100, 100, 0, false)

		// 30x30の画像を位置(-10, -10)に合成（左上に拡張）
		mergeImg := ebiten.NewImage(30, 30)
		ps.MergeImage(mergeImg, -10, -10, false)

		// サイズが拡張される（10+50=60）
		if ps.GetWidth() != 60 || ps.GetHeight() != 60 {
			t.Errorf("Expected size (60, 60), got (%d, %d)", ps.GetWidth(), ps.GetHeight())
		}

		// 位置が調整される（100-10=90）
		if ps.GetDestX() != 90 || ps.GetDestY() != 90 {
			t.Errorf("Expected position (90, 90), got (%d, %d)", ps.GetDestX(), ps.GetDestY())
		}
	})

	t.Run("nil画像の合成は何もしない", func(t *testing.T) {
		srcImg := ebiten.NewImage(50, 50)
		ps := psm.CreatePictureSprite(srcImg, 4, 0, 0, 50, 50, 0, 0, 0, false)

		// nil画像を合成
		ps.MergeImage(nil, 10, 10, false)

		// サイズは変わらない
		if ps.GetWidth() != 50 || ps.GetHeight() != 50 {
			t.Errorf("Expected size (50, 50), got (%d, %d)", ps.GetWidth(), ps.GetHeight())
		}
	})

	t.Run("透明色処理付きの画像合成", func(t *testing.T) {
		srcImg := ebiten.NewImage(50, 50)
		ps := psm.CreatePictureSprite(srcImg, 5, 0, 0, 50, 50, 0, 0, 0, false)

		// 透明色処理付きで合成
		mergeImg := ebiten.NewImage(30, 30)
		ps.MergeImage(mergeImg, 10, 10, true)

		// サイズは変わらない
		if ps.GetWidth() != 50 || ps.GetHeight() != 50 {
			t.Errorf("Expected size (50, 50), got (%d, %d)", ps.GetWidth(), ps.GetHeight())
		}
	})

	t.Run("現在の画像がnilの場合", func(t *testing.T) {
		// 画像なしでPictureSpriteを作成（CreatePictureSpriteは内部で画像を作成するため、
		// 直接PictureSpriteを構築してテストする）
		sprite := sm.CreateSprite(nil)
		ps := &PictureSprite{
			sprite:      sprite,
			picID:       6,
			srcX:        0,
			srcY:        0,
			width:       0,
			height:      0,
			destX:       50,
			destY:       50,
			transparent: false,
			state:       PictureSpriteAttached,
			windowID:    -1,
		}

		// 30x30の画像を合成
		mergeImg := ebiten.NewImage(30, 30)
		ps.MergeImage(mergeImg, 10, 20, false)

		// サイズが設定される
		if ps.GetWidth() != 30 || ps.GetHeight() != 30 {
			t.Errorf("Expected size (30, 30), got (%d, %d)", ps.GetWidth(), ps.GetHeight())
		}

		// 位置が設定される
		if ps.GetDestX() != 10 || ps.GetDestY() != 20 {
			t.Errorf("Expected position (10, 20), got (%d, %d)", ps.GetDestX(), ps.GetDestY())
		}
	})
}

// TestPictureSpriteMergeImageMultiple は複数回のMergeImageをテストする
func TestPictureSpriteMergeImageMultiple(t *testing.T) {
	sm := NewSpriteManager()
	psm := NewPictureSpriteManager(sm)

	// 50x50の画像を持つPictureSpriteを作成
	srcImg := ebiten.NewImage(50, 50)
	ps := psm.CreatePictureSprite(srcImg, 1, 0, 0, 50, 50, 0, 0, 0, false)

	// 1回目の合成: 右下に拡張
	mergeImg1 := ebiten.NewImage(30, 30)
	ps.MergeImage(mergeImg1, 40, 40, false)

	if ps.GetWidth() != 70 || ps.GetHeight() != 70 {
		t.Errorf("After first merge: Expected size (70, 70), got (%d, %d)", ps.GetWidth(), ps.GetHeight())
	}

	// 2回目の合成: さらに右下に拡張
	mergeImg2 := ebiten.NewImage(20, 20)
	ps.MergeImage(mergeImg2, 60, 60, false)

	if ps.GetWidth() != 80 || ps.GetHeight() != 80 {
		t.Errorf("After second merge: Expected size (80, 80), got (%d, %d)", ps.GetWidth(), ps.GetHeight())
	}

	// 3回目の合成: 領域内に収まる
	mergeImg3 := ebiten.NewImage(10, 10)
	ps.MergeImage(mergeImg3, 30, 30, false)

	// サイズは変わらない
	if ps.GetWidth() != 80 || ps.GetHeight() != 80 {
		t.Errorf("After third merge: Expected size (80, 80), got (%d, %d)", ps.GetWidth(), ps.GetHeight())
	}
}

// TestMergeOrCreatePictureSprite はMergeOrCreatePictureSprite()メソッドをテストする
// タスク 6.3: PictureSpriteManagerにMergeOrCreatePictureSprite()を追加する
func TestMergeOrCreatePictureSprite(t *testing.T) {
	sm := NewSpriteManager()
	psm := NewPictureSpriteManager(sm)

	// 親スプライトを作成
	parentImg := ebiten.NewImage(640, 480)
	parent := sm.CreateSprite(parentImg)

	t.Run("融合可能なスプライトがない場合は新規作成", func(t *testing.T) {
		srcImg := ebiten.NewImage(50, 50)
		ps, merged := psm.MergeOrCreatePictureSprite(
			srcImg,
			10,   // picID
			0, 0, // srcX, srcY
			50, 50, // width, height
			100, 100, // destX, destY
			0,     // zOrder
			false, // transparent
			parent,
		)

		if ps == nil {
			t.Fatal("MergeOrCreatePictureSprite returned nil")
		}
		if merged {
			t.Error("Expected merged to be false for new sprite creation")
		}
		if ps.GetDestX() != 100 || ps.GetDestY() != 100 {
			t.Errorf("Expected position (100, 100), got (%d, %d)", ps.GetDestX(), ps.GetDestY())
		}
		if ps.GetWidth() != 50 || ps.GetHeight() != 50 {
			t.Errorf("Expected size (50, 50), got (%d, %d)", ps.GetWidth(), ps.GetHeight())
		}
	})

	t.Run("融合可能なスプライトがある場合は融合", func(t *testing.T) {
		// 最初のスプライトを作成
		srcImg1 := ebiten.NewImage(50, 50)
		ps1, _ := psm.MergeOrCreatePictureSprite(
			srcImg1,
			20,   // picID
			0, 0, // srcX, srcY
			50, 50, // width, height
			200, 200, // destX, destY
			0,     // zOrder
			false, // transparent
			parent,
		)

		// 重なる領域に2番目のスプライトを作成（融合されるはず）
		srcImg2 := ebiten.NewImage(30, 30)
		ps2, merged := psm.MergeOrCreatePictureSprite(
			srcImg2,
			20,   // 同じpicID
			0, 0, // srcX, srcY
			30, 30, // width, height
			220, 220, // destX, destY（重なる位置）
			0,     // zOrder
			false, // transparent
			parent,
		)

		if ps2 == nil {
			t.Fatal("MergeOrCreatePictureSprite returned nil")
		}
		if !merged {
			t.Error("Expected merged to be true for overlapping region")
		}
		if ps2 != ps1 {
			t.Error("Expected to return the same PictureSprite after merge")
		}
		// 融合後、サイズが拡張される（200+50=250, 220+30=250）
		// 元のサイズ50x50、位置200,200
		// 新しい領域は220,220から30x30なので、右下が250,250
		// 結果: 幅=250-200=50, 高さ=250-200=50 → 変わらない（元の領域内に収まる）
		// 実際には220+30=250 > 200+50=250なので同じ
		// 正確には: max(200+50, 220+30) - 200 = max(250, 250) - 200 = 50
		if ps2.GetWidth() != 50 || ps2.GetHeight() != 50 {
			t.Errorf("Expected size (50, 50), got (%d, %d)", ps2.GetWidth(), ps2.GetHeight())
		}
	})

	t.Run("異なるpicIDでは融合しない", func(t *testing.T) {
		// 最初のスプライトを作成
		srcImg1 := ebiten.NewImage(50, 50)
		psm.MergeOrCreatePictureSprite(
			srcImg1,
			30,   // picID
			0, 0, // srcX, srcY
			50, 50, // width, height
			300, 300, // destX, destY
			0,     // zOrder
			false, // transparent
			parent,
		)

		// 異なるpicIDで同じ位置にスプライトを作成
		srcImg2 := ebiten.NewImage(30, 30)
		ps2, merged := psm.MergeOrCreatePictureSprite(
			srcImg2,
			31,   // 異なるpicID
			0, 0, // srcX, srcY
			30, 30, // width, height
			300, 300, // destX, destY（同じ位置）
			0,     // zOrder
			false, // transparent
			parent,
		)

		if ps2 == nil {
			t.Fatal("MergeOrCreatePictureSprite returned nil")
		}
		if merged {
			t.Error("Expected merged to be false for different picID")
		}
	})

	t.Run("異なる親スプライトでは融合しない", func(t *testing.T) {
		// 別の親スプライトを作成
		otherParentImg := ebiten.NewImage(640, 480)
		otherParent := sm.CreateSprite(otherParentImg)

		// 最初のスプライトを作成
		srcImg1 := ebiten.NewImage(50, 50)
		psm.MergeOrCreatePictureSprite(
			srcImg1,
			40,   // picID
			0, 0, // srcX, srcY
			50, 50, // width, height
			400, 400, // destX, destY
			0,     // zOrder
			false, // transparent
			parent,
		)

		// 異なる親で同じ位置にスプライトを作成
		srcImg2 := ebiten.NewImage(30, 30)
		ps2, merged := psm.MergeOrCreatePictureSprite(
			srcImg2,
			40,   // 同じpicID
			0, 0, // srcX, srcY
			30, 30, // width, height
			400, 400, // destX, destY（同じ位置）
			0,           // zOrder
			false,       // transparent
			otherParent, // 異なる親
		)

		if ps2 == nil {
			t.Fatal("MergeOrCreatePictureSprite returned nil")
		}
		if merged {
			t.Error("Expected merged to be false for different parent")
		}
	})

	t.Run("離れた領域では融合しない", func(t *testing.T) {
		// 最初のスプライトを作成
		srcImg1 := ebiten.NewImage(50, 50)
		psm.MergeOrCreatePictureSprite(
			srcImg1,
			50,   // picID
			0, 0, // srcX, srcY
			50, 50, // width, height
			0, 0, // destX, destY
			0,     // zOrder
			false, // transparent
			parent,
		)

		// 離れた位置にスプライトを作成
		srcImg2 := ebiten.NewImage(30, 30)
		ps2, merged := psm.MergeOrCreatePictureSprite(
			srcImg2,
			50,   // 同じpicID
			0, 0, // srcX, srcY
			30, 30, // width, height
			500, 500, // destX, destY（離れた位置）
			0,     // zOrder
			false, // transparent
			parent,
		)

		if ps2 == nil {
			t.Fatal("MergeOrCreatePictureSprite returned nil")
		}
		if merged {
			t.Error("Expected merged to be false for distant region")
		}
	})

	t.Run("親がnilの場合でも動作する", func(t *testing.T) {
		srcImg := ebiten.NewImage(50, 50)
		ps, merged := psm.MergeOrCreatePictureSprite(
			srcImg,
			60,   // picID
			0, 0, // srcX, srcY
			50, 50, // width, height
			600, 600, // destX, destY
			0,     // zOrder
			false, // transparent
			nil,   // 親なし
		)

		if ps == nil {
			t.Fatal("MergeOrCreatePictureSprite returned nil")
		}
		if merged {
			t.Error("Expected merged to be false for new sprite with nil parent")
		}
	})
}

// TestMergeOrCreatePictureSpriteWithExpansion は領域拡張を伴う融合をテストする
func TestMergeOrCreatePictureSpriteWithExpansion(t *testing.T) {
	sm := NewSpriteManager()
	psm := NewPictureSpriteManager(sm)

	// 親スプライトを作成
	parentImg := ebiten.NewImage(640, 480)
	parent := sm.CreateSprite(parentImg)

	t.Run("右下方向への領域拡張", func(t *testing.T) {
		// 最初のスプライトを作成（位置: 100,100、サイズ: 50x50）
		srcImg1 := ebiten.NewImage(50, 50)
		ps1, _ := psm.MergeOrCreatePictureSprite(
			srcImg1,
			100,  // picID
			0, 0, // srcX, srcY
			50, 50, // width, height
			100, 100, // destX, destY
			0,     // zOrder
			false, // transparent
			parent,
		)

		// 右下に拡張する位置にスプライトを作成（位置: 130,130、サイズ: 40x40）
		// 融合後の領域: 100,100 から 170,170 → サイズ: 70x70
		srcImg2 := ebiten.NewImage(40, 40)
		ps2, merged := psm.MergeOrCreatePictureSprite(
			srcImg2,
			100,  // 同じpicID
			0, 0, // srcX, srcY
			40, 40, // width, height
			130, 130, // destX, destY
			0,     // zOrder
			false, // transparent
			parent,
		)

		if !merged {
			t.Error("Expected merged to be true")
		}
		if ps2 != ps1 {
			t.Error("Expected to return the same PictureSprite")
		}
		// 融合後のサイズを確認
		// 元: 100,100 から 150,150 (50x50)
		// 新: 130,130 から 170,170 (40x40)
		// 結果: 100,100 から 170,170 → サイズ: 70x70
		if ps2.GetWidth() != 70 || ps2.GetHeight() != 70 {
			t.Errorf("Expected size (70, 70), got (%d, %d)", ps2.GetWidth(), ps2.GetHeight())
		}
	})

	t.Run("左上方向への領域拡張", func(t *testing.T) {
		// 最初のスプライトを作成（位置: 200,200、サイズ: 50x50）
		srcImg1 := ebiten.NewImage(50, 50)
		ps1, _ := psm.MergeOrCreatePictureSprite(
			srcImg1,
			101,  // picID
			0, 0, // srcX, srcY
			50, 50, // width, height
			200, 200, // destX, destY
			0,     // zOrder
			false, // transparent
			parent,
		)

		// 左上に拡張する位置にスプライトを作成（位置: 180,180、サイズ: 40x40）
		// 融合後の領域: 180,180 から 250,250 → サイズ: 70x70
		srcImg2 := ebiten.NewImage(40, 40)
		ps2, merged := psm.MergeOrCreatePictureSprite(
			srcImg2,
			101,  // 同じpicID
			0, 0, // srcX, srcY
			40, 40, // width, height
			180, 180, // destX, destY
			0,     // zOrder
			false, // transparent
			parent,
		)

		if !merged {
			t.Error("Expected merged to be true")
		}
		if ps2 != ps1 {
			t.Error("Expected to return the same PictureSprite")
		}
		// 融合後のサイズを確認
		// 元: 200,200 から 250,250 (50x50)
		// 新: 180,180 から 220,220 (40x40)
		// 結果: 180,180 から 250,250 → サイズ: 70x70
		// 位置も調整される: 200 → 180
		if ps2.GetWidth() != 70 || ps2.GetHeight() != 70 {
			t.Errorf("Expected size (70, 70), got (%d, %d)", ps2.GetWidth(), ps2.GetHeight())
		}
		if ps2.GetDestX() != 180 || ps2.GetDestY() != 180 {
			t.Errorf("Expected position (180, 180), got (%d, %d)", ps2.GetDestX(), ps2.GetDestY())
		}
	})
}
