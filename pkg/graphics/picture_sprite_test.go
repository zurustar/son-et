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

	if sprite.ZOrder() != 5 {
		t.Errorf("Expected ZOrder 5, got %d", sprite.ZOrder())
	}

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

	// スプライトのZ順序を確認
	if ps.GetSprite().ZOrder() != 10 {
		t.Errorf("Expected ZOrder 10, got %d", ps.GetSprite().ZOrder())
	}
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

// TestPictureSpriteSetZOrder はZ順序の更新をテストする
func TestPictureSpriteSetZOrder(t *testing.T) {
	sm := NewSpriteManager()
	psm := NewPictureSpriteManager(sm)

	srcImg := ebiten.NewImage(50, 50)
	ps := psm.CreatePictureSprite(srcImg, 1, 0, 0, 50, 50, 0, 0, 0, false)

	// Z順序を更新
	ps.SetZOrder(100)

	// スプライトのZ順序が更新されたことを確認
	if ps.GetSprite().ZOrder() != 100 {
		t.Errorf("Expected ZOrder 100, got %d", ps.GetSprite().ZOrder())
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
