package graphics

import (
	"image/color"
	"testing"
)

func TestNewShapeSpriteManager(t *testing.T) {
	sm := NewSpriteManager()
	ssm := NewShapeSpriteManager(sm)

	if ssm == nil {
		t.Fatal("NewShapeSpriteManager returned nil")
	}
	if ssm.spriteManager != sm {
		t.Error("SpriteManager not set correctly")
	}
	if ssm.Count() != 0 {
		t.Errorf("Expected count 0, got %d", ssm.Count())
	}
}

func TestCreateLineSprite(t *testing.T) {
	sm := NewSpriteManager()
	ssm := NewShapeSpriteManager(sm)

	lineColor := color.RGBA{255, 0, 0, 255}
	ss := ssm.CreateLineSprite(1, 10, 20, 100, 80, lineColor, 2, 5)

	if ss == nil {
		t.Fatal("CreateLineSprite returned nil")
	}

	// 図形タイプの確認
	if ss.GetShapeType() != ShapeTypeLine {
		t.Errorf("Expected ShapeTypeLine, got %v", ss.GetShapeType())
	}

	// 色の確認
	if ss.GetColor() != lineColor {
		t.Error("Color not set correctly")
	}

	// 線の太さの確認
	if ss.GetLineSize() != 2 {
		t.Errorf("Expected line size 2, got %d", ss.GetLineSize())
	}

	// 座標の確認
	x1, y1, x2, y2 := ss.GetLineCoords()
	if x1 != 10 || y1 != 20 || x2 != 100 || y2 != 80 {
		t.Errorf("Line coords incorrect: (%d,%d)->(%d,%d)", x1, y1, x2, y2)
	}

	// スプライトの確認
	sprite := ss.GetSprite()
	if sprite == nil {
		t.Fatal("Sprite is nil")
	}
	// Z_Pathはまだ設定されていない（親スプライトなしで作成されたため）
	// zOrderパラメータは互換性のために残されているが、実際のZ順序はZ_Pathで管理される
	if !sprite.Visible() {
		t.Error("Sprite should be visible")
	}

	// カウントの確認
	if ssm.Count() != 1 {
		t.Errorf("Expected count 1, got %d", ssm.Count())
	}
}

func TestCreateRectSprite(t *testing.T) {
	sm := NewSpriteManager()
	ssm := NewShapeSpriteManager(sm)

	rectColor := color.RGBA{0, 255, 0, 255}
	ss := ssm.CreateRectSprite(2, 50, 50, 150, 100, rectColor, 3, 10)

	if ss == nil {
		t.Fatal("CreateRectSprite returned nil")
	}

	if ss.GetShapeType() != ShapeTypeRect {
		t.Errorf("Expected ShapeTypeRect, got %v", ss.GetShapeType())
	}

	if ss.GetFillMode() != 0 {
		t.Errorf("Expected fill mode 0, got %d", ss.GetFillMode())
	}

	x1, y1, x2, y2 := ss.GetRectCoords()
	if x1 != 50 || y1 != 50 || x2 != 150 || y2 != 100 {
		t.Errorf("Rect coords incorrect: (%d,%d)->(%d,%d)", x1, y1, x2, y2)
	}
}

func TestCreateFillRectSprite(t *testing.T) {
	sm := NewSpriteManager()
	ssm := NewShapeSpriteManager(sm)

	fillColor := color.RGBA{0, 0, 255, 255}
	ss := ssm.CreateFillRectSprite(3, 0, 0, 100, 100, fillColor, 15)

	if ss == nil {
		t.Fatal("CreateFillRectSprite returned nil")
	}

	if ss.GetShapeType() != ShapeTypeFillRect {
		t.Errorf("Expected ShapeTypeFillRect, got %v", ss.GetShapeType())
	}

	if ss.GetFillMode() != 2 {
		t.Errorf("Expected fill mode 2, got %d", ss.GetFillMode())
	}

	sprite := ss.GetSprite()
	if sprite == nil {
		t.Fatal("Sprite is nil")
	}

	// 位置の確認
	x, y := sprite.Position()
	if x != 0 || y != 0 {
		t.Errorf("Expected position (0,0), got (%f,%f)", x, y)
	}
}

func TestCreateCircleSprite(t *testing.T) {
	sm := NewSpriteManager()
	ssm := NewShapeSpriteManager(sm)

	circleColor := color.RGBA{255, 255, 0, 255}
	ss := ssm.CreateCircleSprite(4, 100, 100, 50, circleColor, 2, 20)

	if ss == nil {
		t.Fatal("CreateCircleSprite returned nil")
	}

	if ss.GetShapeType() != ShapeTypeCircle {
		t.Errorf("Expected ShapeTypeCircle, got %v", ss.GetShapeType())
	}

	cx, cy, radius := ss.GetCircleParams()
	if cx != 100 || cy != 100 || radius != 50 {
		t.Errorf("Circle params incorrect: center=(%d,%d), radius=%d", cx, cy, radius)
	}
}

func TestCreateFillCircleSprite(t *testing.T) {
	sm := NewSpriteManager()
	ssm := NewShapeSpriteManager(sm)

	fillColor := color.RGBA{255, 0, 255, 255}
	ss := ssm.CreateFillCircleSprite(5, 200, 200, 30, fillColor, 25)

	if ss == nil {
		t.Fatal("CreateFillCircleSprite returned nil")
	}

	if ss.GetShapeType() != ShapeTypeFillCircle {
		t.Errorf("Expected ShapeTypeFillCircle, got %v", ss.GetShapeType())
	}

	if ss.GetFillMode() != 2 {
		t.Errorf("Expected fill mode 2, got %d", ss.GetFillMode())
	}
}

func TestCreateCircleSpriteWithZeroRadius(t *testing.T) {
	sm := NewSpriteManager()
	ssm := NewShapeSpriteManager(sm)

	ss := ssm.CreateCircleSprite(1, 100, 100, 0, color.White, 1, 0)
	if ss != nil {
		t.Error("Expected nil for zero radius circle")
	}

	ss = ssm.CreateFillCircleSprite(1, 100, 100, -5, color.White, 0)
	if ss != nil {
		t.Error("Expected nil for negative radius circle")
	}
}

func TestGetShapeSprites(t *testing.T) {
	sm := NewSpriteManager()
	ssm := NewShapeSpriteManager(sm)

	// 同じpicIDに複数の図形を作成
	ssm.CreateLineSprite(1, 0, 0, 10, 10, color.White, 1, 0)
	ssm.CreateRectSprite(1, 20, 20, 40, 40, color.White, 1, 0)
	ssm.CreateFillRectSprite(1, 50, 50, 70, 70, color.White, 0)

	// 別のpicIDに図形を作成
	ssm.CreateLineSprite(2, 0, 0, 10, 10, color.White, 1, 0)

	sprites := ssm.GetShapeSprites(1)
	if len(sprites) != 3 {
		t.Errorf("Expected 3 sprites for picID 1, got %d", len(sprites))
	}

	sprites = ssm.GetShapeSprites(2)
	if len(sprites) != 1 {
		t.Errorf("Expected 1 sprite for picID 2, got %d", len(sprites))
	}

	sprites = ssm.GetShapeSprites(999)
	if sprites != nil {
		t.Error("Expected nil for non-existent picID")
	}
}

func TestRemoveShapeSprite(t *testing.T) {
	sm := NewSpriteManager()
	ssm := NewShapeSpriteManager(sm)

	ss1 := ssm.CreateLineSprite(1, 0, 0, 10, 10, color.White, 1, 0)
	ss2 := ssm.CreateRectSprite(1, 20, 20, 40, 40, color.White, 1, 0)

	if ssm.Count() != 2 {
		t.Errorf("Expected count 2, got %d", ssm.Count())
	}

	ssm.RemoveShapeSprite(ss1)

	if ssm.Count() != 1 {
		t.Errorf("Expected count 1 after removal, got %d", ssm.Count())
	}

	sprites := ssm.GetShapeSprites(1)
	if len(sprites) != 1 {
		t.Errorf("Expected 1 sprite remaining, got %d", len(sprites))
	}
	if sprites[0] != ss2 {
		t.Error("Wrong sprite remaining")
	}
}

func TestRemoveShapeSpritesByPicID(t *testing.T) {
	sm := NewSpriteManager()
	ssm := NewShapeSpriteManager(sm)

	ssm.CreateLineSprite(1, 0, 0, 10, 10, color.White, 1, 0)
	ssm.CreateRectSprite(1, 20, 20, 40, 40, color.White, 1, 0)
	ssm.CreateLineSprite(2, 0, 0, 10, 10, color.White, 1, 0)

	if ssm.Count() != 3 {
		t.Errorf("Expected count 3, got %d", ssm.Count())
	}

	ssm.RemoveShapeSpritesByPicID(1)

	if ssm.Count() != 1 {
		t.Errorf("Expected count 1 after removal, got %d", ssm.Count())
	}

	sprites := ssm.GetShapeSprites(1)
	if sprites != nil {
		t.Error("Expected nil for removed picID")
	}

	sprites = ssm.GetShapeSprites(2)
	if len(sprites) != 1 {
		t.Errorf("Expected 1 sprite for picID 2, got %d", len(sprites))
	}
}

func TestShapeSpriteManagerClear(t *testing.T) {
	sm := NewSpriteManager()
	ssm := NewShapeSpriteManager(sm)

	ssm.CreateLineSprite(1, 0, 0, 10, 10, color.White, 1, 0)
	ssm.CreateRectSprite(2, 20, 20, 40, 40, color.White, 1, 0)
	ssm.CreateFillRectSprite(3, 50, 50, 70, 70, color.White, 0)

	if ssm.Count() != 3 {
		t.Errorf("Expected count 3, got %d", ssm.Count())
	}

	ssm.Clear()

	if ssm.Count() != 0 {
		t.Errorf("Expected count 0 after clear, got %d", ssm.Count())
	}
}

func TestShapeSpriteSetPosition(t *testing.T) {
	sm := NewSpriteManager()
	ssm := NewShapeSpriteManager(sm)

	ss := ssm.CreateFillRectSprite(1, 0, 0, 50, 50, color.White, 0)

	ss.SetPosition(100, 200)

	sprite := ss.GetSprite()
	x, y := sprite.Position()
	if x != 100 || y != 200 {
		t.Errorf("Expected position (100,200), got (%f,%f)", x, y)
	}
}

func TestShapeSpriteSetVisible(t *testing.T) {
	sm := NewSpriteManager()
	ssm := NewShapeSpriteManager(sm)

	ss := ssm.CreateRectSprite(1, 0, 0, 50, 50, color.White, 1, 0)

	if !ss.GetSprite().Visible() {
		t.Error("Sprite should be visible initially")
	}

	ss.SetVisible(false)

	if ss.GetSprite().Visible() {
		t.Error("Sprite should be invisible after SetVisible(false)")
	}
}

func TestShapeSpriteSetParent(t *testing.T) {
	sm := NewSpriteManager()
	ssm := NewShapeSpriteManager(sm)

	parent := sm.CreateSpriteWithSize(100, 100)
	ss := ssm.CreateFillRectSprite(1, 10, 10, 30, 30, color.White, 0)

	ss.SetParent(parent)

	if ss.GetSprite().Parent() != parent {
		t.Error("Parent not set correctly")
	}
}

func TestRectCoordinateNormalization(t *testing.T) {
	sm := NewSpriteManager()
	ssm := NewShapeSpriteManager(sm)

	// 逆順の座標で矩形を作成
	ss := ssm.CreateRectSprite(1, 100, 80, 50, 20, color.White, 1, 0)

	x1, y1, x2, y2 := ss.GetRectCoords()
	// 座標は正規化されているはず
	if x1 != 50 || y1 != 20 || x2 != 100 || y2 != 80 {
		t.Errorf("Coords not normalized: (%d,%d)->(%d,%d)", x1, y1, x2, y2)
	}
}

func TestRemoveNilShapeSprite(t *testing.T) {
	sm := NewSpriteManager()
	ssm := NewShapeSpriteManager(sm)

	// nilを削除してもパニックしないことを確認
	ssm.RemoveShapeSprite(nil)
}

func TestGetPicID(t *testing.T) {
	sm := NewSpriteManager()
	ssm := NewShapeSpriteManager(sm)

	ss := ssm.CreateLineSprite(42, 0, 0, 10, 10, color.White, 1, 0)

	if ss.GetPicID() != 42 {
		t.Errorf("Expected picID 42, got %d", ss.GetPicID())
	}
}

func TestCreateVerticalLine(t *testing.T) {
	sm := NewSpriteManager()
	ssm := NewShapeSpriteManager(sm)

	// 垂直線（x1 == x2）を作成
	ss := ssm.CreateLineSprite(1, 50, 10, 50, 100, color.White, 2, 0)

	if ss == nil {
		t.Fatal("CreateLineSprite returned nil for vertical line")
	}

	x1, y1, x2, y2 := ss.GetLineCoords()
	if x1 != 50 || y1 != 10 || x2 != 50 || y2 != 100 {
		t.Errorf("Vertical line coords incorrect: (%d,%d)->(%d,%d)", x1, y1, x2, y2)
	}

	// スプライトが正しく作成されていることを確認
	sprite := ss.GetSprite()
	if sprite == nil {
		t.Fatal("Sprite is nil")
	}
	if sprite.Image() == nil {
		t.Fatal("Sprite image is nil")
	}
}

func TestCreateHorizontalLine(t *testing.T) {
	sm := NewSpriteManager()
	ssm := NewShapeSpriteManager(sm)

	// 水平線（y1 == y2）を作成
	ss := ssm.CreateLineSprite(1, 10, 50, 100, 50, color.White, 2, 0)

	if ss == nil {
		t.Fatal("CreateLineSprite returned nil for horizontal line")
	}

	x1, y1, x2, y2 := ss.GetLineCoords()
	if x1 != 10 || y1 != 50 || x2 != 100 || y2 != 50 {
		t.Errorf("Horizontal line coords incorrect: (%d,%d)->(%d,%d)", x1, y1, x2, y2)
	}
}

func TestCreateLineWithReversedCoords(t *testing.T) {
	sm := NewSpriteManager()
	ssm := NewShapeSpriteManager(sm)

	// 逆順の座標で線を作成（x2 < x1, y2 < y1）
	ss := ssm.CreateLineSprite(1, 100, 80, 10, 20, color.White, 2, 0)

	if ss == nil {
		t.Fatal("CreateLineSprite returned nil")
	}

	// 座標は元のまま保持される
	x1, y1, x2, y2 := ss.GetLineCoords()
	if x1 != 100 || y1 != 80 || x2 != 10 || y2 != 20 {
		t.Errorf("Line coords should be preserved: (%d,%d)->(%d,%d)", x1, y1, x2, y2)
	}
}

func TestCreateFillRectWithReversedCoords(t *testing.T) {
	sm := NewSpriteManager()
	ssm := NewShapeSpriteManager(sm)

	// 逆順の座標で塗りつぶし矩形を作成
	ss := ssm.CreateFillRectSprite(1, 100, 80, 50, 20, color.White, 0)

	if ss == nil {
		t.Fatal("CreateFillRectSprite returned nil")
	}

	// 座標は正規化されているはず
	x1, y1, x2, y2 := ss.GetRectCoords()
	if x1 != 50 || y1 != 20 || x2 != 100 || y2 != 80 {
		t.Errorf("FillRect coords not normalized: (%d,%d)->(%d,%d)", x1, y1, x2, y2)
	}
}

func TestCreateFillRectWithMinimalSize(t *testing.T) {
	sm := NewSpriteManager()
	ssm := NewShapeSpriteManager(sm)

	// 同じ座標で塗りつぶし矩形を作成（width=0, height=0）
	ss := ssm.CreateFillRectSprite(1, 50, 50, 50, 50, color.White, 0)

	if ss == nil {
		t.Fatal("CreateFillRectSprite returned nil for minimal size")
	}

	// スプライトが正しく作成されていることを確認
	sprite := ss.GetSprite()
	if sprite == nil {
		t.Fatal("Sprite is nil")
	}
	if sprite.Image() == nil {
		t.Fatal("Sprite image is nil")
	}

	// 最小サイズ（1x1）になっているはず
	bounds := sprite.Image().Bounds()
	if bounds.Dx() < 1 || bounds.Dy() < 1 {
		t.Errorf("Image size should be at least 1x1, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestRemoveShapeSpriteFromEmptyList(t *testing.T) {
	sm := NewSpriteManager()
	ssm := NewShapeSpriteManager(sm)

	// 存在しないpicIDのShapeSpriteを作成して削除
	ss := ssm.CreateLineSprite(1, 0, 0, 10, 10, color.White, 1, 0)

	// 一度削除
	ssm.RemoveShapeSprite(ss)

	// 同じものを再度削除してもパニックしないことを確認
	ssm.RemoveShapeSprite(ss)
}

func TestShapeSpriteWithNilSprite(t *testing.T) {
	// nilスプライトを持つShapeSpriteのメソッドがパニックしないことを確認
	ss := &ShapeSprite{
		sprite: nil,
	}

	// これらのメソッドはnilスプライトでもパニックしない
	ss.SetPosition(10, 20)
	ss.SetVisible(true)
	ss.SetParent(nil)
}
