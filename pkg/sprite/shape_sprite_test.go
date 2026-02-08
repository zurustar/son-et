package sprite

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
	ss := ssm.CreateLineSprite(1, 10, 20, 100, 80, lineColor, 2, nil)

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
	ss := ssm.CreateRectSprite(2, 50, 50, 150, 100, rectColor, 3, nil)

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
	ss := ssm.CreateFillRectSprite(3, 0, 0, 100, 100, fillColor, nil)

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
	ss := ssm.CreateCircleSprite(4, 100, 100, 50, circleColor, 2, nil)

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
	ss := ssm.CreateFillCircleSprite(5, 200, 200, 30, fillColor, nil)

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

	ss := ssm.CreateCircleSprite(1, 100, 100, 0, color.White, 1, nil)
	if ss != nil {
		t.Error("Expected nil for zero radius circle")
	}

	ss = ssm.CreateFillCircleSprite(1, 100, 100, -5, color.White, nil)
	if ss != nil {
		t.Error("Expected nil for negative radius circle")
	}
}

func TestGetShapeSprites(t *testing.T) {
	sm := NewSpriteManager()
	ssm := NewShapeSpriteManager(sm)

	// 同じpicIDに複数の図形を作成
	ssm.CreateLineSprite(1, 0, 0, 10, 10, color.White, 1, nil)
	ssm.CreateRectSprite(1, 20, 20, 40, 40, color.White, 1, nil)
	ssm.CreateFillRectSprite(1, 50, 50, 70, 70, color.White, nil)

	// 別のpicIDに図形を作成
	ssm.CreateLineSprite(2, 0, 0, 10, 10, color.White, 1, nil)

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

	ss1 := ssm.CreateLineSprite(1, 0, 0, 10, 10, color.White, 1, nil)
	ss2 := ssm.CreateRectSprite(1, 20, 20, 40, 40, color.White, 1, nil)

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

	ssm.CreateLineSprite(1, 0, 0, 10, 10, color.White, 1, nil)
	ssm.CreateRectSprite(1, 20, 20, 40, 40, color.White, 1, nil)
	ssm.CreateLineSprite(2, 0, 0, 10, 10, color.White, 1, nil)

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

	ssm.CreateLineSprite(1, 0, 0, 10, 10, color.White, 1, nil)
	ssm.CreateRectSprite(2, 20, 20, 40, 40, color.White, 1, nil)
	ssm.CreateFillRectSprite(3, 50, 50, 70, 70, color.White, nil)

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

	ss := ssm.CreateFillRectSprite(1, 0, 0, 50, 50, color.White, nil)

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

	ss := ssm.CreateRectSprite(1, 0, 0, 50, 50, color.White, 1, nil)

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

	parent := sm.CreateSpriteWithSize(100, 100, nil)
	ss := ssm.CreateFillRectSprite(1, 10, 10, 30, 30, color.White, nil)

	ss.SetParent(parent)

	if ss.GetSprite().Parent() != parent {
		t.Error("Parent not set correctly")
	}
}

func TestRectCoordinateNormalization(t *testing.T) {
	sm := NewSpriteManager()
	ssm := NewShapeSpriteManager(sm)

	// 逆順の座標で矩形を作成
	ss := ssm.CreateRectSprite(1, 100, 80, 50, 20, color.White, 1, nil)

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

	ss := ssm.CreateLineSprite(42, 0, 0, 10, 10, color.White, 1, nil)

	if ss.GetPicID() != 42 {
		t.Errorf("Expected picID 42, got %d", ss.GetPicID())
	}
}

func TestCreateVerticalLine(t *testing.T) {
	sm := NewSpriteManager()
	ssm := NewShapeSpriteManager(sm)

	// 垂直線（x1 == x2）を作成
	ss := ssm.CreateLineSprite(1, 50, 10, 50, 100, color.White, 2, nil)

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
	ss := ssm.CreateLineSprite(1, 10, 50, 100, 50, color.White, 2, nil)

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
	ss := ssm.CreateLineSprite(1, 100, 80, 10, 20, color.White, 2, nil)

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
	ss := ssm.CreateFillRectSprite(1, 100, 80, 50, 20, color.White, nil)

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
	ss := ssm.CreateFillRectSprite(1, 50, 50, 50, 50, color.White, nil)

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
	ss := ssm.CreateLineSprite(1, 0, 0, 10, 10, color.White, 1, nil)

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

// TestShapeSpriteAsChildOfPicture は図形が対象ピクチャーの子として追加されることをテストする
// 要件8対応: 図形を対象ピクチャーの子として追加する
func TestShapeSpriteAsChildOfPicture(t *testing.T) {
	sm := NewSpriteManager()
	ssm := NewShapeSpriteManager(sm)

	// 親となるピクチャースプライトを作成
	parent := sm.CreateSpriteWithSize(200, 200, nil)
	parent.SetPosition(100, 100)

	// 線を親の子として作成
	lineSS := ssm.CreateLineSprite(1, 10, 10, 50, 50, color.White, 2, parent)
	if lineSS == nil {
		t.Fatal("CreateLineSprite returned nil")
	}

	// 親子関係の確認
	lineSprite := lineSS.GetSprite()
	if lineSprite.Parent() != parent {
		t.Error("Line sprite should have parent as its parent")
	}

	// 親の子リストに含まれていることを確認
	children := parent.GetChildren()
	found := false
	for _, child := range children {
		if child == lineSprite {
			found = true
			break
		}
	}
	if !found {
		t.Error("Line sprite should be in parent's children list")
	}
}

// TestMultipleShapesAsChildrenOfPicture は複数の図形が対象ピクチャーの子として追加されることをテストする
func TestMultipleShapesAsChildrenOfPicture(t *testing.T) {
	sm := NewSpriteManager()
	ssm := NewShapeSpriteManager(sm)

	// 親となるピクチャースプライトを作成
	parent := sm.CreateSpriteWithSize(300, 300, nil)

	// 複数の図形を親の子として作成
	lineSS := ssm.CreateLineSprite(1, 10, 10, 50, 50, color.White, 2, parent)
	rectSS := ssm.CreateRectSprite(1, 60, 60, 100, 100, color.White, 2, parent)
	fillRectSS := ssm.CreateFillRectSprite(1, 110, 110, 150, 150, color.White, parent)
	circleSS := ssm.CreateCircleSprite(1, 200, 200, 30, color.White, 2, parent)
	fillCircleSS := ssm.CreateFillCircleSprite(1, 250, 250, 20, color.White, parent)

	// すべての図形が親の子として追加されていることを確認
	children := parent.GetChildren()
	if len(children) != 5 {
		t.Errorf("Expected 5 children, got %d", len(children))
	}

	// 各図形の親が正しいことを確認
	sprites := []*ShapeSprite{lineSS, rectSS, fillRectSS, circleSS, fillCircleSS}
	for i, ss := range sprites {
		if ss.GetSprite().Parent() != parent {
			t.Errorf("Shape %d should have parent as its parent", i)
		}
	}

	// 描画順序が追加順であることを確認（スライスの順序）
	if children[0] != lineSS.GetSprite() {
		t.Error("Line should be first child")
	}
	if children[1] != rectSS.GetSprite() {
		t.Error("Rect should be second child")
	}
	if children[2] != fillRectSS.GetSprite() {
		t.Error("FillRect should be third child")
	}
	if children[3] != circleSS.GetSprite() {
		t.Error("Circle should be fourth child")
	}
	if children[4] != fillCircleSS.GetSprite() {
		t.Error("FillCircle should be fifth child")
	}
}

// TestShapeSpriteInheritedVisibility は親の可視性が子の図形に継承されることをテストする
func TestShapeSpriteInheritedVisibility(t *testing.T) {
	sm := NewSpriteManager()
	ssm := NewShapeSpriteManager(sm)

	// 親となるスプライトを作成
	parent := sm.CreateSpriteWithSize(200, 200, nil)
	parent.SetVisible(true)

	// 図形を親の子として作成
	ss := ssm.CreateFillRectSprite(1, 10, 10, 50, 50, color.White, parent)

	// 初期状態では可視
	if !ss.GetSprite().IsEffectivelyVisible() {
		t.Error("Shape should be effectively visible initially")
	}

	// 親を非表示にする
	parent.SetVisible(false)

	// 子も実効的に非表示になる
	if ss.GetSprite().IsEffectivelyVisible() {
		t.Error("Shape should be effectively invisible when parent is invisible")
	}

	// 親を再表示
	parent.SetVisible(true)

	// 子も実効的に表示される
	if !ss.GetSprite().IsEffectivelyVisible() {
		t.Error("Shape should be effectively visible when parent is visible")
	}
}

// TestShapeSpriteAbsolutePosition は親の位置が子の絶対位置に反映されることをテストする
func TestShapeSpriteAbsolutePosition(t *testing.T) {
	sm := NewSpriteManager()
	ssm := NewShapeSpriteManager(sm)

	// 親となるスプライトを作成
	parent := sm.CreateSpriteWithSize(200, 200, nil)
	parent.SetPosition(100, 50)

	// 図形を親の子として作成（位置は(10, 20)）
	ss := ssm.CreateFillRectSprite(1, 10, 20, 50, 60, color.White, parent)

	// 絶対位置を確認
	absX, absY := ss.GetSprite().AbsolutePosition()
	expectedX := 100.0 + 10.0 // 親のX + 子のX
	expectedY := 50.0 + 20.0  // 親のY + 子のY

	if absX != expectedX || absY != expectedY {
		t.Errorf("Expected absolute position (%f,%f), got (%f,%f)", expectedX, expectedY, absX, absY)
	}
}
