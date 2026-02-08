package graphics

import (
	"log/slog"
	"os"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestNewGraphicsSystem(t *testing.T) {
	// デフォルトのGraphicsSystemを作成
	gs := NewGraphicsSystem("")

	if gs == nil {
		t.Fatal("NewGraphicsSystem returned nil")
	}

	if gs.virtualWidth != 1024 {
		t.Errorf("Expected virtualWidth 1024, got %d", gs.virtualWidth)
	}

	if gs.virtualHeight != 768 {
		t.Errorf("Expected virtualHeight 768, got %d", gs.virtualHeight)
	}

	if gs.pictures == nil {
		t.Error("PictureManager not initialized")
	}

	if gs.windows == nil {
		t.Error("WindowManager not initialized")
	}

	if gs.casts == nil {
		t.Error("CastManager not initialized")
	}

	if gs.textRenderer == nil {
		t.Error("TextRenderer not initialized")
	}

	// スプライトシステム要件 3.1〜3.6: SpriteManagerが初期化されていることを確認
	if gs.spriteManager == nil {
		t.Error("SpriteManager not initialized")
	}
}

func TestNewGraphicsSystemWithOptions(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	gs := NewGraphicsSystem("",
		WithLogger(logger),
		WithVirtualSize(800, 600),
	)

	if gs.virtualWidth != 800 {
		t.Errorf("Expected virtualWidth 800, got %d", gs.virtualWidth)
	}

	if gs.virtualHeight != 600 {
		t.Errorf("Expected virtualHeight 600, got %d", gs.virtualHeight)
	}

	if gs.log != logger {
		t.Error("Logger not set correctly")
	}
}

func TestGraphicsSystemUpdate(t *testing.T) {
	gs := NewGraphicsSystem("")

	// Update を呼び出してエラーが発生しないことを確認
	err := gs.Update()
	if err != nil {
		t.Errorf("Update returned error: %v", err)
	}
}

func TestGraphicsSystemDraw(t *testing.T) {
	gs := NewGraphicsSystem("")

	// ダミーのスクリーンを作成
	screen := ebiten.NewImage(1024, 768)

	// Draw を呼び出す（エラーが発生しないことを確認）
	gs.Draw(screen)
}

func TestGraphicsSystemShutdown(t *testing.T) {
	gs := NewGraphicsSystem("")

	// Shutdown を呼び出す
	gs.Shutdown()
}

// TestGraphicsSystem_GetSpriteManager はGetSpriteManagerメソッドをテストする
// スプライトシステム要件 3.1〜3.6: GraphicsSystemにSpriteManagerを統合する
func TestGraphicsSystem_GetSpriteManager(t *testing.T) {
	gs := NewGraphicsSystem("")

	// GetSpriteManagerがnilでないことを確認
	sm := gs.GetSpriteManager()
	if sm == nil {
		t.Fatal("GetSpriteManager returned nil")
	}

	// SpriteManagerが正しく動作することを確認
	// スプライトを作成
	sprite := sm.CreateSpriteWithSize(100, 100)
	if sprite == nil {
		t.Fatal("CreateSpriteWithSize returned nil")
	}

	// スプライトIDが正しいことを確認
	if sprite.ID() <= 0 {
		t.Errorf("Expected positive sprite ID, got %d", sprite.ID())
	}

	// 同じスプライトを取得できることを確認
	sprite2 := sm.GetSprite(sprite.ID())
	if sprite2 != sprite {
		t.Error("GetSprite returned different instance")
	}

	// SpriteManagerのカウントを確認
	if sm.Count() != 1 {
		t.Errorf("Expected 1 sprite, got %d", sm.Count())
	}

	// スプライトを削除
	sm.RemoveSprite(sprite.ID())
	if sm.Count() != 0 {
		t.Errorf("Expected 0 sprites after removal, got %d", sm.Count())
	}

	// 削除後は取得できないことを確認
	if sm.GetSprite(sprite.ID()) != nil {
		t.Error("GetSprite should return nil after removal")
	}
}

// TestGraphicsSystem_SpriteManagerInitialization はSpriteManagerの初期化をテストする
func TestGraphicsSystem_SpriteManagerInitialization(t *testing.T) {
	gs := NewGraphicsSystem("")

	sm := gs.GetSpriteManager()
	if sm == nil {
		t.Fatal("SpriteManager not initialized")
	}

	// 初期状態ではスプライトがないことを確認
	if sm.Count() != 0 {
		t.Errorf("Expected 0 sprites initially, got %d", sm.Count())
	}

	// 複数のスプライトを作成
	sprite1 := sm.CreateSpriteWithSize(50, 50)
	sprite2 := sm.CreateSpriteWithSize(100, 100)
	sprite3 := sm.CreateSpriteWithSize(150, 150)

	if sm.Count() != 3 {
		t.Errorf("Expected 3 sprites, got %d", sm.Count())
	}

	// 各スプライトが異なるIDを持つことを確認
	if sprite1.ID() == sprite2.ID() || sprite2.ID() == sprite3.ID() || sprite1.ID() == sprite3.ID() {
		t.Error("Sprites should have unique IDs")
	}

	// Clearで全スプライトを削除
	sm.Clear()
	if sm.Count() != 0 {
		t.Errorf("Expected 0 sprites after Clear, got %d", sm.Count())
	}
}

// TestGraphicsSystem_GetShapeSpriteManager はGetShapeSpriteManagerメソッドをテストする
// スプライトシステム要件 9.1〜9.3: GraphicsSystemにShapeSpriteManagerを統合する
func TestGraphicsSystem_GetShapeSpriteManager(t *testing.T) {
	gs := NewGraphicsSystem("")

	// GetShapeSpriteManagerがnilでないことを確認
	ssm := gs.GetShapeSpriteManager()
	if ssm == nil {
		t.Fatal("GetShapeSpriteManager returned nil")
	}

	// 初期状態ではShapeSpriteがないことを確認
	if ssm.Count() != 0 {
		t.Errorf("Expected 0 shape sprites initially, got %d", ssm.Count())
	}
}

// TestGraphicsSystem_DrawLine_CreatesShapeSprite はDrawLineがShapeSpriteを作成することをテストする
// スプライトシステム要件 9.1: 線を描画したスプライトを作成できる
func TestGraphicsSystem_DrawLine_CreatesShapeSprite(t *testing.T) {
	gs := NewGraphicsSystem("")

	// ピクチャーを作成
	picID, err := gs.CreatePic(200, 200)
	if err != nil {
		t.Fatalf("CreatePic failed: %v", err)
	}

	// 線を描画
	err = gs.DrawLine(picID, 10, 10, 100, 100)
	if err != nil {
		t.Fatalf("DrawLine failed: %v", err)
	}

	// ShapeSpriteが作成されていることを確認
	ssm := gs.GetShapeSpriteManager()
	if ssm.Count() != 1 {
		t.Errorf("Expected 1 shape sprite, got %d", ssm.Count())
	}

	// ShapeSpriteの種類を確認
	sprites := ssm.GetShapeSprites(picID)
	if len(sprites) != 1 {
		t.Fatalf("Expected 1 shape sprite for picID %d, got %d", picID, len(sprites))
	}

	if sprites[0].GetShapeType() != ShapeTypeLine {
		t.Errorf("Expected ShapeTypeLine, got %v", sprites[0].GetShapeType())
	}
}

// TestGraphicsSystem_DrawRect_CreatesShapeSprite はDrawRectがShapeSpriteを作成することをテストする
// スプライトシステム要件 9.2: 矩形を描画したスプライトを作成できる
func TestGraphicsSystem_DrawRect_CreatesShapeSprite(t *testing.T) {
	gs := NewGraphicsSystem("")

	// ピクチャーを作成
	picID, err := gs.CreatePic(200, 200)
	if err != nil {
		t.Fatalf("CreatePic failed: %v", err)
	}

	// 矩形（輪郭のみ）を描画 - fillMode=1は輪郭のみ
	err = gs.DrawRect(picID, 10, 10, 100, 100, 1)
	if err != nil {
		t.Fatalf("DrawRect failed: %v", err)
	}

	// ShapeSpriteが作成されていることを確認
	ssm := gs.GetShapeSpriteManager()
	if ssm.Count() != 1 {
		t.Errorf("Expected 1 shape sprite, got %d", ssm.Count())
	}

	// ShapeSpriteの種類を確認
	sprites := ssm.GetShapeSprites(picID)
	if len(sprites) != 1 {
		t.Fatalf("Expected 1 shape sprite for picID %d, got %d", picID, len(sprites))
	}

	if sprites[0].GetShapeType() != ShapeTypeRect {
		t.Errorf("Expected ShapeTypeRect, got %v", sprites[0].GetShapeType())
	}
}

// TestGraphicsSystem_DrawRect_FillMode_CreatesShapeSprite はDrawRectの塗りつぶしモードがShapeSpriteを作成することをテストする
// スプライトシステム要件 9.3: 塗りつぶし矩形を描画したスプライトを作成できる
func TestGraphicsSystem_DrawRect_FillMode_CreatesShapeSprite(t *testing.T) {
	gs := NewGraphicsSystem("")

	// ピクチャーを作成
	picID, err := gs.CreatePic(200, 200)
	if err != nil {
		t.Fatalf("CreatePic failed: %v", err)
	}

	// 塗りつぶし矩形を描画 - fillMode=0は塗りつぶし（サンプルの動作に基づく）
	err = gs.DrawRect(picID, 10, 10, 100, 100, 0)
	if err != nil {
		t.Fatalf("DrawRect failed: %v", err)
	}

	// ShapeSpriteが作成されていることを確認
	ssm := gs.GetShapeSpriteManager()
	if ssm.Count() != 1 {
		t.Errorf("Expected 1 shape sprite, got %d", ssm.Count())
	}

	// ShapeSpriteの種類を確認
	sprites := ssm.GetShapeSprites(picID)
	if len(sprites) != 1 {
		t.Fatalf("Expected 1 shape sprite for picID %d, got %d", picID, len(sprites))
	}

	if sprites[0].GetShapeType() != ShapeTypeFillRect {
		t.Errorf("Expected ShapeTypeFillRect, got %v", sprites[0].GetShapeType())
	}
}

// TestGraphicsSystem_FillRect_CreatesShapeSprite はFillRectがShapeSpriteを作成することをテストする
// スプライトシステム要件 9.3: 塗りつぶし矩形を描画したスプライトを作成できる
func TestGraphicsSystem_FillRect_CreatesShapeSprite(t *testing.T) {
	gs := NewGraphicsSystem("")

	// ピクチャーを作成
	picID, err := gs.CreatePic(200, 200)
	if err != nil {
		t.Fatalf("CreatePic failed: %v", err)
	}

	// 塗りつぶし矩形を描画
	err = gs.FillRect(picID, 10, 10, 100, 100, 0xFF0000)
	if err != nil {
		t.Fatalf("FillRect failed: %v", err)
	}

	// ShapeSpriteが作成されていることを確認
	ssm := gs.GetShapeSpriteManager()
	if ssm.Count() != 1 {
		t.Errorf("Expected 1 shape sprite, got %d", ssm.Count())
	}

	// ShapeSpriteの種類を確認
	sprites := ssm.GetShapeSprites(picID)
	if len(sprites) != 1 {
		t.Fatalf("Expected 1 shape sprite for picID %d, got %d", picID, len(sprites))
	}

	if sprites[0].GetShapeType() != ShapeTypeFillRect {
		t.Errorf("Expected ShapeTypeFillRect, got %v", sprites[0].GetShapeType())
	}
}

// TestGraphicsSystem_DrawCircle_CreatesShapeSprite はDrawCircleがShapeSpriteを作成することをテストする
// スプライトシステム要件 9: 円を描画したスプライトを作成できる
func TestGraphicsSystem_DrawCircle_CreatesShapeSprite(t *testing.T) {
	gs := NewGraphicsSystem("")

	// ピクチャーを作成
	picID, err := gs.CreatePic(200, 200)
	if err != nil {
		t.Fatalf("CreatePic failed: %v", err)
	}

	// 円（輪郭のみ）を描画
	err = gs.DrawCircle(picID, 100, 100, 50, 0)
	if err != nil {
		t.Fatalf("DrawCircle failed: %v", err)
	}

	// ShapeSpriteが作成されていることを確認
	ssm := gs.GetShapeSpriteManager()
	if ssm.Count() != 1 {
		t.Errorf("Expected 1 shape sprite, got %d", ssm.Count())
	}

	// ShapeSpriteの種類を確認
	sprites := ssm.GetShapeSprites(picID)
	if len(sprites) != 1 {
		t.Fatalf("Expected 1 shape sprite for picID %d, got %d", picID, len(sprites))
	}

	if sprites[0].GetShapeType() != ShapeTypeCircle {
		t.Errorf("Expected ShapeTypeCircle, got %v", sprites[0].GetShapeType())
	}
}

// TestGraphicsSystem_DrawCircle_FillMode_CreatesShapeSprite はDrawCircleの塗りつぶしモードがShapeSpriteを作成することをテストする
// スプライトシステム要件 9: 塗りつぶし円を描画したスプライトを作成できる
func TestGraphicsSystem_DrawCircle_FillMode_CreatesShapeSprite(t *testing.T) {
	gs := NewGraphicsSystem("")

	// ピクチャーを作成
	picID, err := gs.CreatePic(200, 200)
	if err != nil {
		t.Fatalf("CreatePic failed: %v", err)
	}

	// 塗りつぶし円を描画
	err = gs.DrawCircle(picID, 100, 100, 50, 2)
	if err != nil {
		t.Fatalf("DrawCircle failed: %v", err)
	}

	// ShapeSpriteが作成されていることを確認
	ssm := gs.GetShapeSpriteManager()
	if ssm.Count() != 1 {
		t.Errorf("Expected 1 shape sprite, got %d", ssm.Count())
	}

	// ShapeSpriteの種類を確認
	sprites := ssm.GetShapeSprites(picID)
	if len(sprites) != 1 {
		t.Fatalf("Expected 1 shape sprite for picID %d, got %d", picID, len(sprites))
	}

	if sprites[0].GetShapeType() != ShapeTypeFillCircle {
		t.Errorf("Expected ShapeTypeFillCircle, got %v", sprites[0].GetShapeType())
	}
}

// TestGraphicsSystem_MultipleShapes_CreatesMultipleShapeSprites は複数の図形描画が複数のShapeSpriteを作成することをテストする
func TestGraphicsSystem_MultipleShapes_CreatesMultipleShapeSprites(t *testing.T) {
	gs := NewGraphicsSystem("")

	// ピクチャーを作成
	picID, err := gs.CreatePic(200, 200)
	if err != nil {
		t.Fatalf("CreatePic failed: %v", err)
	}

	// 複数の図形を描画
	gs.DrawLine(picID, 10, 10, 50, 50)
	gs.DrawRect(picID, 60, 60, 100, 100, 0)
	gs.FillRect(picID, 110, 110, 150, 150, 0x00FF00)
	gs.DrawCircle(picID, 175, 175, 20, 0)

	// ShapeSpriteが4つ作成されていることを確認
	ssm := gs.GetShapeSpriteManager()
	if ssm.Count() != 4 {
		t.Errorf("Expected 4 shape sprites, got %d", ssm.Count())
	}

	// すべてのShapeSpriteが同じpicIDに関連付けられていることを確認
	sprites := ssm.GetShapeSprites(picID)
	if len(sprites) != 4 {
		t.Errorf("Expected 4 shape sprites for picID %d, got %d", picID, len(sprites))
	}
}

// TestGraphicsSystem_DrawWithSpriteManager はDrawWithSpriteManagerメソッドをテストする
// スプライトシステム要件 14.1: SpriteManager.Draw()ベースの描画
func TestGraphicsSystem_DrawWithSpriteManager(t *testing.T) {
	gs := NewGraphicsSystem("")

	// ダミーのスクリーンを作成
	screen := ebiten.NewImage(1024, 768)

	// DrawWithSpriteManagerを呼び出す（エラーが発生しないことを確認）
	gs.DrawWithSpriteManager(screen)

	// SpriteManagerにスプライトを追加
	sm := gs.GetSpriteManager()
	sprite := sm.CreateSpriteWithSize(100, 100)
	sprite.SetPosition(50, 50)
	sprite.SetVisible(true)

	// 再度DrawWithSpriteManagerを呼び出す
	gs.DrawWithSpriteManager(screen)

	// スプライトが描画されたことを確認（エラーが発生しないことを確認）
	if sm.Count() != 1 {
		t.Errorf("Expected 1 sprite, got %d", sm.Count())
	}
}

// TestGraphicsSystem_DrawWithSpriteManager_NilSpriteManager はSpriteManagerがnilの場合のテスト
func TestGraphicsSystem_DrawWithSpriteManager_NilSpriteManager(t *testing.T) {
	gs := NewGraphicsSystem("")

	// SpriteManagerをnilに設定（テスト用）
	gs.spriteManager = nil

	// ダミーのスクリーンを作成
	screen := ebiten.NewImage(1024, 768)

	// DrawWithSpriteManagerを呼び出す（パニックしないことを確認）
	gs.DrawWithSpriteManager(screen)
}

// TestGraphicsSystem_DrawWithSpriteManager_MultipleSprites は複数のスプライトの描画をテストする
func TestGraphicsSystem_DrawWithSpriteManager_MultipleSprites(t *testing.T) {
	gs := NewGraphicsSystem("")

	// ダミーのスクリーンを作成
	screen := ebiten.NewImage(1024, 768)

	// SpriteManagerに複数のスプライトを追加
	sm := gs.GetSpriteManager()

	// 異なるZ_Pathでスプライトを作成
	sprite1 := sm.CreateSpriteWithSize(50, 50)
	sprite1.SetPosition(10, 10)
	sprite1.SetZPath(NewZPath(100))
	sprite1.SetVisible(true)

	sprite2 := sm.CreateSpriteWithSize(50, 50)
	sprite2.SetPosition(20, 20)
	sprite2.SetZPath(NewZPath(50)) // sprite1より背面
	sprite2.SetVisible(true)

	sprite3 := sm.CreateSpriteWithSize(50, 50)
	sprite3.SetPosition(30, 30)
	sprite3.SetZPath(NewZPath(150)) // sprite1より前面
	sprite3.SetVisible(true)

	// DrawWithSpriteManagerを呼び出す
	gs.DrawWithSpriteManager(screen)

	// スプライトが3つあることを確認
	if sm.Count() != 3 {
		t.Errorf("Expected 3 sprites, got %d", sm.Count())
	}
}

// TestGraphicsSystem_DrawWithSpriteManager_InvisibleSprites は非表示スプライトの描画をテストする
func TestGraphicsSystem_DrawWithSpriteManager_InvisibleSprites(t *testing.T) {
	gs := NewGraphicsSystem("")

	// ダミーのスクリーンを作成
	screen := ebiten.NewImage(1024, 768)

	// SpriteManagerにスプライトを追加
	sm := gs.GetSpriteManager()

	// 可視スプライト
	visibleSprite := sm.CreateSpriteWithSize(50, 50)
	visibleSprite.SetPosition(10, 10)
	visibleSprite.SetVisible(true)

	// 非表示スプライト
	invisibleSprite := sm.CreateSpriteWithSize(50, 50)
	invisibleSprite.SetPosition(20, 20)
	invisibleSprite.SetVisible(false)

	// DrawWithSpriteManagerを呼び出す（非表示スプライトは描画されない）
	gs.DrawWithSpriteManager(screen)

	// スプライトが2つあることを確認
	if sm.Count() != 2 {
		t.Errorf("Expected 2 sprites, got %d", sm.Count())
	}

	// 可視性を確認
	if !visibleSprite.Visible() {
		t.Error("visibleSprite should be visible")
	}
	if invisibleSprite.Visible() {
		t.Error("invisibleSprite should not be visible")
	}
}
