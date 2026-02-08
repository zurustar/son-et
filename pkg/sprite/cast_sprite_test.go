package sprite

import (
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestNewCastSpriteManager(t *testing.T) {
	sm := NewSpriteManager()
	csm := NewCastSpriteManager(sm)

	if csm == nil {
		t.Fatal("NewCastSpriteManager returned nil")
	}
	if csm.Count() != 0 {
		t.Errorf("expected 0 cast sprites, got %d", csm.Count())
	}
}

func TestCreateCastSprite(t *testing.T) {
	sm := NewSpriteManager()
	csm := NewCastSpriteManager(sm)

	// テスト用のソース画像を作成
	srcImage := ebiten.NewImage(64, 64)

	// CastSpriteを作成
	config := CastConfig{
		ID:            0,
		WinID:         0,
		PicID:         1,
		X:             10,
		Y:             20,
		SrcX:          0,
		SrcY:          0,
		Width:         32,
		Height:        32,
		Visible:       true,
		TransColor:    nil,
		HasTransColor: false,
	}
	cs := csm.CreateCastSprite(config, srcImage, nil)

	if cs == nil {
		t.Fatal("CreateCastSprite returned nil")
	}
	if csm.Count() != 1 {
		t.Errorf("expected 1 cast sprite, got %d", csm.Count())
	}

	// スプライトの属性を確認
	sprite := cs.GetSprite()
	if sprite == nil {
		t.Fatal("GetSprite returned nil")
	}

	x, y := sprite.Position()
	if x != 10 || y != 20 {
		t.Errorf("expected position (10, 20), got (%f, %f)", x, y)
	}

	if !sprite.Visible() {
		t.Error("expected sprite to be visible")
	}
}

func TestGetCastSprite(t *testing.T) {
	sm := NewSpriteManager()
	csm := NewCastSpriteManager(sm)

	srcImage := ebiten.NewImage(64, 64)
	config := CastConfig{
		ID:      5,
		WinID:   0,
		PicID:   1,
		X:       10,
		Y:       20,
		SrcX:    0,
		SrcY:    0,
		Width:   32,
		Height:  32,
		Visible: true,
	}
	csm.CreateCastSprite(config, srcImage, nil)

	// 存在するキャストを取得
	cs := csm.GetCastSprite(5)
	if cs == nil {
		t.Fatal("GetCastSprite returned nil for existing cast")
	}

	// 存在しないキャストを取得
	cs = csm.GetCastSprite(999)
	if cs != nil {
		t.Error("GetCastSprite should return nil for non-existing cast")
	}
}

func TestRemoveCastSprite(t *testing.T) {
	sm := NewSpriteManager()
	csm := NewCastSpriteManager(sm)

	srcImage := ebiten.NewImage(64, 64)
	config := CastConfig{
		ID:      0,
		WinID:   0,
		PicID:   1,
		X:       10,
		Y:       20,
		SrcX:    0,
		SrcY:    0,
		Width:   32,
		Height:  32,
		Visible: true,
	}
	csm.CreateCastSprite(config, srcImage, nil)

	if csm.Count() != 1 {
		t.Errorf("expected 1 cast sprite, got %d", csm.Count())
	}

	// CastSpriteを削除
	csm.RemoveCastSprite(0)

	if csm.Count() != 0 {
		t.Errorf("expected 0 cast sprites after removal, got %d", csm.Count())
	}

	// 削除後は取得できない
	cs := csm.GetCastSprite(0)
	if cs != nil {
		t.Error("GetCastSprite should return nil after removal")
	}
}

func TestGetCastSpritesByWindow(t *testing.T) {
	sm := NewSpriteManager()
	csm := NewCastSpriteManager(sm)

	srcImage := ebiten.NewImage(64, 64)

	// ウィンドウ0に2つのキャストを作成
	config1 := CastConfig{ID: 0, WinID: 0, PicID: 1, X: 10, Y: 10, Width: 32, Height: 32, Visible: true}
	config2 := CastConfig{ID: 1, WinID: 0, PicID: 1, X: 20, Y: 20, Width: 32, Height: 32, Visible: true}

	// ウィンドウ1に1つのキャストを作成
	config3 := CastConfig{ID: 2, WinID: 1, PicID: 1, X: 30, Y: 30, Width: 32, Height: 32, Visible: true}

	csm.CreateCastSprite(config1, srcImage, nil)
	csm.CreateCastSprite(config2, srcImage, nil)
	csm.CreateCastSprite(config3, srcImage, nil)

	// ウィンドウ0のキャストを取得
	sprites := csm.GetCastSpritesByWindow(0)
	if len(sprites) != 2 {
		t.Errorf("expected 2 cast sprites for window 0, got %d", len(sprites))
	}

	// ウィンドウ1のキャストを取得
	sprites = csm.GetCastSpritesByWindow(1)
	if len(sprites) != 1 {
		t.Errorf("expected 1 cast sprite for window 1, got %d", len(sprites))
	}

	// 存在しないウィンドウのキャストを取得
	sprites = csm.GetCastSpritesByWindow(999)
	if len(sprites) != 0 {
		t.Errorf("expected 0 cast sprites for window 999, got %d", len(sprites))
	}
}

func TestRemoveCastSpritesByWindow(t *testing.T) {
	sm := NewSpriteManager()
	csm := NewCastSpriteManager(sm)

	srcImage := ebiten.NewImage(64, 64)

	// ウィンドウ0に2つのキャストを作成
	config1 := CastConfig{ID: 0, WinID: 0, PicID: 1, X: 10, Y: 10, Width: 32, Height: 32, Visible: true}
	config2 := CastConfig{ID: 1, WinID: 0, PicID: 1, X: 20, Y: 20, Width: 32, Height: 32, Visible: true}

	// ウィンドウ1に1つのキャストを作成
	config3 := CastConfig{ID: 2, WinID: 1, PicID: 1, X: 30, Y: 30, Width: 32, Height: 32, Visible: true}

	csm.CreateCastSprite(config1, srcImage, nil)
	csm.CreateCastSprite(config2, srcImage, nil)
	csm.CreateCastSprite(config3, srcImage, nil)

	if csm.Count() != 3 {
		t.Errorf("expected 3 cast sprites, got %d", csm.Count())
	}

	// ウィンドウ0のキャストを削除
	csm.RemoveCastSpritesByWindow(0)

	if csm.Count() != 1 {
		t.Errorf("expected 1 cast sprite after removal, got %d", csm.Count())
	}

	// ウィンドウ0のキャストは取得できない
	sprites := csm.GetCastSpritesByWindow(0)
	if len(sprites) != 0 {
		t.Errorf("expected 0 cast sprites for window 0 after removal, got %d", len(sprites))
	}

	// ウィンドウ1のキャストは残っている
	sprites = csm.GetCastSpritesByWindow(1)
	if len(sprites) != 1 {
		t.Errorf("expected 1 cast sprite for window 1 after removal, got %d", len(sprites))
	}
}

func TestCastSpriteClear(t *testing.T) {
	sm := NewSpriteManager()
	csm := NewCastSpriteManager(sm)

	srcImage := ebiten.NewImage(64, 64)

	// 複数のキャストを作成
	for i := range 5 {
		config := CastConfig{ID: i, WinID: 0, PicID: 1, X: i * 10, Y: i * 10, Width: 32, Height: 32, Visible: true}
		csm.CreateCastSprite(config, srcImage, nil)
	}

	if csm.Count() != 5 {
		t.Errorf("expected 5 cast sprites, got %d", csm.Count())
	}

	// すべてクリア
	csm.Clear()

	if csm.Count() != 0 {
		t.Errorf("expected 0 cast sprites after clear, got %d", csm.Count())
	}
}

func TestCastSpriteUpdatePosition(t *testing.T) {
	sm := NewSpriteManager()
	csm := NewCastSpriteManager(sm)

	srcImage := ebiten.NewImage(64, 64)
	config := CastConfig{ID: 0, WinID: 0, PicID: 1, X: 10, Y: 20, Width: 32, Height: 32, Visible: true}
	cs := csm.CreateCastSprite(config, srcImage, nil)

	// 位置を更新
	cs.UpdatePosition(100, 200)

	// スプライトの位置が更新されていることを確認
	x, y := cs.GetSprite().Position()
	if x != 100 || y != 200 {
		t.Errorf("expected sprite position (100, 200), got (%f, %f)", x, y)
	}
}

func TestCastSpriteUpdateVisible(t *testing.T) {
	sm := NewSpriteManager()
	csm := NewCastSpriteManager(sm)

	srcImage := ebiten.NewImage(64, 64)
	config := CastConfig{ID: 0, WinID: 0, PicID: 1, X: 10, Y: 20, Width: 32, Height: 32, Visible: true}
	cs := csm.CreateCastSprite(config, srcImage, nil)

	// 可視性を更新
	cs.UpdateVisible(false)

	// スプライトの可視性が更新されていることを確認
	if cs.GetSprite().Visible() {
		t.Error("expected sprite to be invisible")
	}
}

func TestCastSpriteSetParent(t *testing.T) {
	sm := NewSpriteManager()
	csm := NewCastSpriteManager(sm)

	srcImage := ebiten.NewImage(64, 64)
	config := CastConfig{ID: 0, WinID: 0, PicID: 1, X: 10, Y: 20, Width: 32, Height: 32, Visible: true}
	cs := csm.CreateCastSprite(config, srcImage, nil)

	// 親スプライトを作成
	parentSprite := sm.CreateSpriteWithSize(100, 100, nil)
	parentSprite.SetPosition(50, 50)

	// 親を設定
	cs.SetParent(parentSprite)

	// 親が設定されていることを確認
	if cs.GetSprite().Parent() != parentSprite {
		t.Error("expected parent to be set")
	}

	// 絶対位置が親の位置を考慮していることを確認
	absX, absY := cs.GetSprite().AbsolutePosition()
	if absX != 60 || absY != 70 { // 10+50, 20+50
		t.Errorf("expected absolute position (60, 70), got (%f, %f)", absX, absY)
	}
}

func TestCastSpriteNilSourceImage(t *testing.T) {
	sm := NewSpriteManager()
	csm := NewCastSpriteManager(sm)

	config := CastConfig{ID: 0, WinID: 0, PicID: 1, X: 10, Y: 20, Width: 32, Height: 32, Visible: true}

	// nilソース画像でCastSpriteを作成
	cs := csm.CreateCastSprite(config, nil, nil)

	if cs == nil {
		t.Fatal("CreateCastSprite should not return nil for nil source image")
	}

	// スプライトの画像はnilになる
	if cs.GetSprite().Image() != nil {
		t.Error("expected sprite image to be nil for nil source image")
	}
}

// TestCastSpriteParentVisibilityInheritance は親の可視性が子に継承されることをテストする
// 要件 2.3: 親が非表示のとき子も非表示として扱う
func TestCastSpriteParentVisibilityInheritance(t *testing.T) {
	sm := NewSpriteManager()
	csm := NewCastSpriteManager(sm)

	// 親スプライトを作成
	parentSprite := sm.CreateSpriteWithSize(200, 150, nil)
	parentSprite.SetVisible(true)

	srcImage := ebiten.NewImage(64, 64)
	config := CastConfig{ID: 0, WinID: 0, PicID: 1, X: 10, Y: 20, Width: 32, Height: 32, Visible: true}

	// 親スプライト付きでCastSpriteを作成
	cs := csm.CreateCastSprite(config, srcImage, parentSprite)

	// 親が表示中の場合、子も実効的に表示
	if !cs.GetSprite().IsEffectivelyVisible() {
		t.Error("expected child to be effectively visible when parent is visible")
	}

	// 親を非表示にする
	parentSprite.SetVisible(false)

	// 親が非表示の場合、子も実効的に非表示
	if cs.GetSprite().IsEffectivelyVisible() {
		t.Error("expected child to be effectively invisible when parent is invisible")
	}
}

// TestCastSpriteWithTransColor は透明色付きでCastSpriteを作成するテスト
func TestCastSpriteWithTransColor(t *testing.T) {
	sm := NewSpriteManager()
	csm := NewCastSpriteManager(sm)

	srcImage := ebiten.NewImage(64, 64)
	transColor := color.RGBA{255, 0, 255, 255} // マゼンタ

	config := CastConfig{
		ID:            0,
		WinID:         0,
		PicID:         1,
		X:             10,
		Y:             20,
		Width:         32,
		Height:        32,
		Visible:       true,
		TransColor:    transColor,
		HasTransColor: true,
	}

	cs := csm.CreateCastSpriteWithTransColor(config, srcImage, transColor, nil)

	if cs == nil {
		t.Fatal("CreateCastSpriteWithTransColor returned nil")
	}

	// 透明色が設定されていることを確認
	if !cs.HasTransColor() {
		t.Error("expected HasTransColor to be true")
	}

	if cs.GetTransColor() != transColor {
		t.Error("expected TransColor to be set correctly")
	}
}

// TestCastSpriteGetSrcPicID はソースピクチャーIDの取得をテストする
func TestCastSpriteGetSrcPicID(t *testing.T) {
	sm := NewSpriteManager()
	csm := NewCastSpriteManager(sm)

	srcImage := ebiten.NewImage(64, 64)
	config := CastConfig{ID: 0, WinID: 0, PicID: 42, X: 10, Y: 20, Width: 32, Height: 32, Visible: true}
	cs := csm.CreateCastSprite(config, srcImage, nil)

	if cs.GetSrcPicID() != 42 {
		t.Errorf("expected SrcPicID 42, got %d", cs.GetSrcPicID())
	}
}

// TestCastSpriteUpdateSource はソース領域の更新をテストする
func TestCastSpriteUpdateSource(t *testing.T) {
	sm := NewSpriteManager()
	csm := NewCastSpriteManager(sm)

	srcImage := ebiten.NewImage(64, 64)
	config := CastConfig{ID: 0, WinID: 0, PicID: 1, X: 10, Y: 20, SrcX: 0, SrcY: 0, Width: 32, Height: 32, Visible: true}
	cs := csm.CreateCastSprite(config, srcImage, nil)

	// ソース領域を更新
	cs.UpdateSource(10, 20, 40, 50)

	// dirtyフラグが設定されていることを確認
	if !cs.IsDirty() {
		t.Error("expected dirty flag to be set after UpdateSource")
	}
}

// TestCastSpriteRebuildCache はキャッシュの再構築をテストする
func TestCastSpriteRebuildCache(t *testing.T) {
	sm := NewSpriteManager()
	csm := NewCastSpriteManager(sm)

	srcImage := ebiten.NewImage(64, 64)
	config := CastConfig{ID: 0, WinID: 0, PicID: 1, X: 10, Y: 20, SrcX: 0, SrcY: 0, Width: 32, Height: 32, Visible: true}
	cs := csm.CreateCastSprite(config, srcImage, nil)

	// ソース領域を更新してdirtyにする
	cs.UpdateSource(10, 20, 40, 50)

	// キャッシュを再構築
	newSrcImage := ebiten.NewImage(100, 100)
	cs.RebuildCache(newSrcImage)

	// dirtyフラグがクリアされていることを確認
	if cs.IsDirty() {
		t.Error("expected dirty flag to be cleared after RebuildCache")
	}
}
