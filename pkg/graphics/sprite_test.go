package graphics

import (
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestNewSprite(t *testing.T) {
	img := ebiten.NewImage(100, 100)
	s := NewSprite(1, img)

	if s.ID() != 1 {
		t.Errorf("expected ID 1, got %d", s.ID())
	}
	if s.Image() != img {
		t.Error("image mismatch")
	}
	x, y := s.Position()
	if x != 0 || y != 0 {
		t.Errorf("expected position (0,0), got (%f,%f)", x, y)
	}
	if s.ZOrder() != 0 {
		t.Errorf("expected zOrder 0, got %d", s.ZOrder())
	}
	if !s.Visible() {
		t.Error("expected visible true")
	}
	if s.Alpha() != 1.0 {
		t.Errorf("expected alpha 1.0, got %f", s.Alpha())
	}
	if s.Parent() != nil {
		t.Error("expected parent nil")
	}
}

func TestSpriteSetters(t *testing.T) {
	s := NewSprite(1, nil)

	s.SetPosition(10, 20)
	x, y := s.Position()
	if x != 10 || y != 20 {
		t.Errorf("expected position (10,20), got (%f,%f)", x, y)
	}

	s.SetZOrder(5)
	if s.ZOrder() != 5 {
		t.Errorf("expected zOrder 5, got %d", s.ZOrder())
	}

	s.SetVisible(false)
	if s.Visible() {
		t.Error("expected visible false")
	}

	s.SetAlpha(0.5)
	if s.Alpha() != 0.5 {
		t.Errorf("expected alpha 0.5, got %f", s.Alpha())
	}

	// Alpha clamping
	s.SetAlpha(-0.5)
	if s.Alpha() != 0 {
		t.Errorf("expected alpha 0, got %f", s.Alpha())
	}
	s.SetAlpha(1.5)
	if s.Alpha() != 1 {
		t.Errorf("expected alpha 1, got %f", s.Alpha())
	}
}

func TestSpriteParentChild(t *testing.T) {
	parent := NewSprite(1, nil)
	parent.SetPosition(100, 50)
	parent.SetAlpha(0.8)

	child := NewSprite(2, nil)
	child.SetPosition(10, 20)
	child.SetAlpha(0.5)
	child.SetParent(parent)

	// AbsolutePosition
	x, y := child.AbsolutePosition()
	if x != 110 || y != 70 {
		t.Errorf("expected absolute position (110,70), got (%f,%f)", x, y)
	}

	// EffectiveAlpha
	alpha := child.EffectiveAlpha()
	expected := 0.8 * 0.5
	if alpha != expected {
		t.Errorf("expected effective alpha %f, got %f", expected, alpha)
	}

	// IsEffectivelyVisible
	if !child.IsEffectivelyVisible() {
		t.Error("expected effectively visible true")
	}

	parent.SetVisible(false)
	if child.IsEffectivelyVisible() {
		t.Error("expected effectively visible false when parent hidden")
	}
}

func TestSpriteSize(t *testing.T) {
	img := ebiten.NewImage(200, 150)
	s := NewSprite(1, img)

	w, h := s.Size()
	if w != 200 || h != 150 {
		t.Errorf("expected size (200,150), got (%d,%d)", w, h)
	}

	// nil image
	s2 := NewSprite(2, nil)
	w, h = s2.Size()
	if w != 0 || h != 0 {
		t.Errorf("expected size (0,0) for nil image, got (%d,%d)", w, h)
	}
}

func TestSpriteManager(t *testing.T) {
	sm := NewSpriteManager()

	if sm.Count() != 0 {
		t.Errorf("expected count 0, got %d", sm.Count())
	}

	s1 := sm.CreateSprite(nil)
	s2 := sm.CreateSprite(nil)

	if sm.Count() != 2 {
		t.Errorf("expected count 2, got %d", sm.Count())
	}

	if s1.ID() == s2.ID() {
		t.Error("sprites should have different IDs")
	}

	// GetSprite
	got := sm.GetSprite(s1.ID())
	if got != s1 {
		t.Error("GetSprite returned wrong sprite")
	}

	// RemoveSprite
	sm.RemoveSprite(s1.ID())
	if sm.Count() != 1 {
		t.Errorf("expected count 1 after remove, got %d", sm.Count())
	}
	if sm.GetSprite(s1.ID()) != nil {
		t.Error("removed sprite should not be found")
	}

	// Clear
	sm.Clear()
	if sm.Count() != 0 {
		t.Errorf("expected count 0 after clear, got %d", sm.Count())
	}
}

func TestSpriteManagerCreateWithSize(t *testing.T) {
	sm := NewSpriteManager()
	s := sm.CreateSpriteWithSize(320, 240)

	w, h := s.Size()
	if w != 320 || h != 240 {
		t.Errorf("expected size (320,240), got (%d,%d)", w, h)
	}
}

func TestSpriteManagerDraw(t *testing.T) {
	sm := NewSpriteManager()

	// 3つのスプライトを異なるZ順序で作成
	img1 := ebiten.NewImage(10, 10)
	img1.Fill(color.RGBA{255, 0, 0, 255})
	s1 := sm.CreateSprite(img1)
	s1.SetZOrder(2)

	img2 := ebiten.NewImage(10, 10)
	img2.Fill(color.RGBA{0, 255, 0, 255})
	s2 := sm.CreateSprite(img2)
	s2.SetZOrder(1)

	img3 := ebiten.NewImage(10, 10)
	img3.Fill(color.RGBA{0, 0, 255, 255})
	s3 := sm.CreateSprite(img3)
	s3.SetZOrder(3)

	// 描画テスト（エラーが出なければOK）
	screen := ebiten.NewImage(100, 100)
	sm.Draw(screen)

	// 非表示スプライトは描画されない
	s2.SetVisible(false)
	sm.Draw(screen)
}

func TestSpriteDirtyFlag(t *testing.T) {
	s := NewSprite(1, nil)

	if !s.IsDirty() {
		t.Error("new sprite should be dirty")
	}

	s.ClearDirty()
	if s.IsDirty() {
		t.Error("sprite should not be dirty after clear")
	}

	s.SetPosition(10, 10)
	if !s.IsDirty() {
		t.Error("sprite should be dirty after SetPosition")
	}

	s.ClearDirty()
	s.SetZOrder(5)
	if !s.IsDirty() {
		t.Error("sprite should be dirty after SetZOrder")
	}

	s.ClearDirty()
	s.SetVisible(false)
	if !s.IsDirty() {
		t.Error("sprite should be dirty after SetVisible")
	}

	s.ClearDirty()
	s.SetAlpha(0.5)
	if !s.IsDirty() {
		t.Error("sprite should be dirty after SetAlpha")
	}
}
