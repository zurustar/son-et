package sprite

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
	if !s.Visible() {
		t.Error("expected visible true")
	}
	if s.Alpha() != 1.0 {
		t.Errorf("expected alpha 1.0, got %f", s.Alpha())
	}
	if s.Parent() != nil {
		t.Error("expected parent nil")
	}
	if s.HasChildren() {
		t.Error("expected no children")
	}
}

func TestSpriteSetters(t *testing.T) {
	s := NewSprite(1, nil)

	s.SetPosition(10, 20)
	x, y := s.Position()
	if x != 10 || y != 20 {
		t.Errorf("expected position (10,20), got (%f,%f)", x, y)
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
	parent.AddChild(child)

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

// TestSpriteChildManagement は子スプライト管理メソッドをテストする
func TestSpriteChildManagement(t *testing.T) {
	parent := NewSprite(1, nil)
	child1 := NewSprite(2, nil)
	child2 := NewSprite(3, nil)

	// 初期状態: 子スプライトなし
	if parent.HasChildren() {
		t.Error("新しいスプライトは子スプライトを持たないはず")
	}
	if len(parent.GetChildren()) != 0 {
		t.Errorf("GetChildren()は空のスライスを返すはず、got %d", len(parent.GetChildren()))
	}

	// AddChild: 子スプライトを追加
	parent.AddChild(child1)
	if !parent.HasChildren() {
		t.Error("AddChild後、HasChildren()はtrueを返すはず")
	}
	if len(parent.GetChildren()) != 1 {
		t.Errorf("AddChild後、GetChildren()は1つの子を返すはず、got %d", len(parent.GetChildren()))
	}
	if child1.Parent() != parent {
		t.Error("AddChild後、子のParent()は親を返すはず")
	}

	// 2つ目の子を追加
	parent.AddChild(child2)
	if len(parent.GetChildren()) != 2 {
		t.Errorf("2つ目のAddChild後、GetChildren()は2つの子を返すはず、got %d", len(parent.GetChildren()))
	}

	// RemoveChild: 子スプライトを削除
	parent.RemoveChild(child1.ID())
	if len(parent.GetChildren()) != 1 {
		t.Errorf("RemoveChild後、GetChildren()は1つの子を返すはず、got %d", len(parent.GetChildren()))
	}
	if child1.Parent() != nil {
		t.Error("RemoveChild後、削除された子のParent()はnilを返すはず")
	}

	// 最後の子を削除
	parent.RemoveChild(child2.ID())
	if parent.HasChildren() {
		t.Error("すべての子を削除後、HasChildren()はfalseを返すはず")
	}
}

// TestSpriteAddChildNil はnilの子スプライトを追加しようとした場合のテスト
func TestSpriteAddChildNil(t *testing.T) {
	parent := NewSprite(1, nil)
	parent.AddChild(nil)
	if parent.HasChildren() {
		t.Error("nilを追加してもHasChildren()はfalseを返すはず")
	}
}

// TestSpriteChildrenOrder は子スプライトの順序が保持されることをテストする
// 要件 9.3: 後から追加されたスプライトはスライスの後方に位置する
func TestSpriteChildrenOrder(t *testing.T) {
	parent := NewSprite(1, nil)
	child1 := NewSprite(2, nil)
	child2 := NewSprite(3, nil)
	child3 := NewSprite(4, nil)

	parent.AddChild(child1)
	parent.AddChild(child2)
	parent.AddChild(child3)

	children := parent.GetChildren()
	if len(children) != 3 {
		t.Fatalf("3つの子スプライトがあるはず、got %d", len(children))
	}
	if children[0] != child1 || children[1] != child2 || children[2] != child3 {
		t.Error("子スプライトは追加順に保持されるはず")
	}

	// 中間の子を削除
	parent.RemoveChild(child2.ID())
	children = parent.GetChildren()
	if len(children) != 2 {
		t.Fatalf("削除後、2つの子スプライトがあるはず、got %d", len(children))
	}
	if children[0] != child1 || children[1] != child3 {
		t.Error("削除後、残りの子スプライトの順序は保持されるはず")
	}
}

// TestSpriteBringToFront はBringToFrontメソッドをテストする
// 要件 12.1: スプライトを最前面に移動する（スライス末尾に移動）
func TestSpriteBringToFront(t *testing.T) {
	parent := NewSprite(1, nil)
	child1 := NewSprite(2, nil)
	child2 := NewSprite(3, nil)
	child3 := NewSprite(4, nil)

	parent.AddChild(child1)
	parent.AddChild(child2)
	parent.AddChild(child3)

	// child1を最前面に移動
	child1.BringToFront()

	children := parent.GetChildren()
	if len(children) != 3 {
		t.Fatalf("3つの子スプライトがあるはず、got %d", len(children))
	}
	// 順序: child2, child3, child1
	if children[0] != child2 || children[1] != child3 || children[2] != child1 {
		t.Errorf("BringToFront後、child1は末尾にあるはず、got %v", []int{children[0].ID(), children[1].ID(), children[2].ID()})
	}
}

// TestSpriteSendToBack はSendToBackメソッドをテストする
// 要件 12.2: スプライトを最背面に移動する（スライス先頭に移動）
func TestSpriteSendToBack(t *testing.T) {
	parent := NewSprite(1, nil)
	child1 := NewSprite(2, nil)
	child2 := NewSprite(3, nil)
	child3 := NewSprite(4, nil)

	parent.AddChild(child1)
	parent.AddChild(child2)
	parent.AddChild(child3)

	// child3を最背面に移動
	child3.SendToBack()

	children := parent.GetChildren()
	if len(children) != 3 {
		t.Fatalf("3つの子スプライトがあるはず、got %d", len(children))
	}
	// 順序: child3, child1, child2
	if children[0] != child3 || children[1] != child1 || children[2] != child2 {
		t.Errorf("SendToBack後、child3は先頭にあるはず、got %v", []int{children[0].ID(), children[1].ID(), children[2].ID()})
	}
}

// TestSpriteBringToFrontNoParent は親がないスプライトでBringToFrontを呼び出した場合のテスト
func TestSpriteBringToFrontNoParent(t *testing.T) {
	s := NewSprite(1, nil)
	// パニックしないことを確認
	s.BringToFront()
}

// TestSpriteSendToBackNoParent は親がないスプライトでSendToBackを呼び出した場合のテスト
func TestSpriteSendToBackNoParent(t *testing.T) {
	s := NewSprite(1, nil)
	// パニックしないことを確認
	s.SendToBack()
}

// TestSpriteManager はSpriteManagerの基本機能をテストする
func TestSpriteManager(t *testing.T) {
	sm := NewSpriteManager()

	if sm.Count() != 0 {
		t.Errorf("expected count 0, got %d", sm.Count())
	}

	s1 := sm.CreateRootSprite(nil)
	s2 := sm.CreateRootSprite(nil)

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

	// DeleteSprite
	sm.DeleteSprite(s1.ID())
	if sm.Count() != 1 {
		t.Errorf("expected count 1 after delete, got %d", sm.Count())
	}
	if sm.GetSprite(s1.ID()) != nil {
		t.Error("deleted sprite should not be found")
	}

	// Clear
	sm.Clear()
	if sm.Count() != 0 {
		t.Errorf("expected count 0 after clear, got %d", sm.Count())
	}
}

// TestSpriteManagerCreateWithParent は親子関係を持つスプライト作成をテストする
func TestSpriteManagerCreateWithParent(t *testing.T) {
	sm := NewSpriteManager()

	// ルートスプライトを作成
	root := sm.CreateRootSprite(nil)

	// 子スプライトを作成
	child1 := sm.CreateSprite(nil, root)
	child2 := sm.CreateSprite(nil, root)

	if child1.Parent() != root {
		t.Error("child1の親はrootのはず")
	}
	if child2.Parent() != root {
		t.Error("child2の親はrootのはず")
	}

	children := root.GetChildren()
	if len(children) != 2 {
		t.Errorf("rootは2つの子を持つはず、got %d", len(children))
	}
	if children[0] != child1 || children[1] != child2 {
		t.Error("子スプライトは追加順に保持されるはず")
	}
}

// TestSpriteManagerDeleteWithChildren は子スプライトを持つスプライトの削除をテストする
// 要件 3.4: スプライト削除時に子スプライトも削除する
func TestSpriteManagerDeleteWithChildren(t *testing.T) {
	sm := NewSpriteManager()

	root := sm.CreateRootSprite(nil)
	child := sm.CreateSprite(nil, root)
	grandchild := sm.CreateSprite(nil, child)

	if sm.Count() != 3 {
		t.Errorf("3つのスプライトがあるはず、got %d", sm.Count())
	}

	// rootを削除すると、child, grandchildも削除される
	sm.DeleteSprite(root.ID())

	if sm.Count() != 0 {
		t.Errorf("すべてのスプライトが削除されるはず、got %d", sm.Count())
	}
	if sm.GetSprite(child.ID()) != nil {
		t.Error("childも削除されるはず")
	}
	if sm.GetSprite(grandchild.ID()) != nil {
		t.Error("grandchildも削除されるはず")
	}
}

// TestSpriteManagerGetRoots はGetRootsメソッドをテストする
func TestSpriteManagerGetRoots(t *testing.T) {
	sm := NewSpriteManager()

	root1 := sm.CreateRootSprite(nil)
	root2 := sm.CreateRootSprite(nil)
	sm.CreateSprite(nil, root1) // 子スプライト

	roots := sm.GetRoots()
	if len(roots) != 2 {
		t.Errorf("2つのルートスプライトがあるはず、got %d", len(roots))
	}
	if roots[0] != root1 || roots[1] != root2 {
		t.Error("ルートスプライトは追加順に保持されるはず")
	}
}

// TestSpriteManagerDraw は描画機能をテストする
func TestSpriteManagerDraw(t *testing.T) {
	sm := NewSpriteManager()

	// スプライトを作成
	img1 := ebiten.NewImage(10, 10)
	img1.Fill(color.RGBA{255, 0, 0, 255})
	root := sm.CreateRootSprite(img1)

	img2 := ebiten.NewImage(10, 10)
	img2.Fill(color.RGBA{0, 255, 0, 255})
	sm.CreateSprite(img2, root)

	// 描画テスト（エラーが出なければOK）
	screen := ebiten.NewImage(100, 100)
	sm.Draw(screen)

	// 非表示スプライトは描画されない
	root.SetVisible(false)
	sm.Draw(screen)
}

// TestSpriteManagerBringRootToFront はルートスプライトの最前面移動をテストする
func TestSpriteManagerBringRootToFront(t *testing.T) {
	sm := NewSpriteManager()

	root1 := sm.CreateRootSprite(nil)
	root2 := sm.CreateRootSprite(nil)
	root3 := sm.CreateRootSprite(nil)

	// root1を最前面に移動
	err := sm.BringRootToFront(root1.ID())
	if err != nil {
		t.Fatalf("BringRootToFrontがエラーを返した: %v", err)
	}

	roots := sm.GetRoots()
	// 順序: root2, root3, root1
	if roots[0] != root2 || roots[1] != root3 || roots[2] != root1 {
		t.Errorf("BringRootToFront後、root1は末尾にあるはず、got %v", []int{roots[0].ID(), roots[1].ID(), roots[2].ID()})
	}
}

// TestSpriteManagerSendRootToBack はルートスプライトの最背面移動をテストする
func TestSpriteManagerSendRootToBack(t *testing.T) {
	sm := NewSpriteManager()

	root1 := sm.CreateRootSprite(nil)
	root2 := sm.CreateRootSprite(nil)
	root3 := sm.CreateRootSprite(nil)

	// root3を最背面に移動
	err := sm.SendRootToBack(root3.ID())
	if err != nil {
		t.Fatalf("SendRootToBackがエラーを返した: %v", err)
	}

	roots := sm.GetRoots()
	// 順序: root3, root1, root2
	if roots[0] != root3 || roots[1] != root1 || roots[2] != root2 {
		t.Errorf("SendRootToBack後、root3は先頭にあるはず、got %v", []int{roots[0].ID(), roots[1].ID(), roots[2].ID()})
	}
}

// TestSpriteManagerBringRootToFrontNotFound は存在しないルートスプライトの最前面移動をテストする
func TestSpriteManagerBringRootToFrontNotFound(t *testing.T) {
	sm := NewSpriteManager()
	sm.CreateRootSprite(nil)

	err := sm.BringRootToFront(999)
	if err == nil {
		t.Error("存在しないIDでBringRootToFrontを呼び出した場合、エラーを返すはず")
	}
}

// TestSpriteDirtyFlag はdirtyフラグをテストする
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

// TestIsEffectivelyVisible_ParentVisibility は親スプライトの可視性が子スプライトに影響することをテストする
// 要件 10.4: 親が非表示の場合は子も描画しない
func TestIsEffectivelyVisible_ParentVisibility(t *testing.T) {
	parent := NewSprite(1, nil)
	parent.SetVisible(true)

	child := NewSprite(2, nil)
	child.SetVisible(true)
	parent.AddChild(child)

	grandchild := NewSprite(3, nil)
	grandchild.SetVisible(true)
	child.AddChild(grandchild)

	// 初期状態: すべて可視
	if !parent.IsEffectivelyVisible() {
		t.Error("親スプライトは可視のはず")
	}
	if !child.IsEffectivelyVisible() {
		t.Error("子スプライトは可視のはず")
	}
	if !grandchild.IsEffectivelyVisible() {
		t.Error("孫スプライトは可視のはず")
	}

	// 親を非表示にする
	parent.SetVisible(false)

	if parent.IsEffectivelyVisible() {
		t.Error("親スプライトは非表示のはず")
	}
	if child.IsEffectivelyVisible() {
		t.Error("親が非表示なので、子スプライトは実効的に非表示のはず")
	}
	if grandchild.IsEffectivelyVisible() {
		t.Error("親が非表示なので、孫スプライトは実効的に非表示のはず")
	}

	// 親を再び可視にする
	parent.SetVisible(true)

	if !child.IsEffectivelyVisible() {
		t.Error("子スプライトは可視のはず")
	}
	if !grandchild.IsEffectivelyVisible() {
		t.Error("孫スプライトは可視のはず")
	}
}

// TestSpriteChildIndex はChildIndexメソッドをテストする
func TestSpriteChildIndex(t *testing.T) {
	parent := NewSprite(1, nil)
	child1 := NewSprite(2, nil)
	child2 := NewSprite(3, nil)
	child3 := NewSprite(4, nil)

	parent.AddChild(child1)
	parent.AddChild(child2)
	parent.AddChild(child3)

	if parent.ChildIndex(child1.ID()) != 0 {
		t.Errorf("child1のインデックスは0のはず、got %d", parent.ChildIndex(child1.ID()))
	}
	if parent.ChildIndex(child2.ID()) != 1 {
		t.Errorf("child2のインデックスは1のはず、got %d", parent.ChildIndex(child2.ID()))
	}
	if parent.ChildIndex(child3.ID()) != 2 {
		t.Errorf("child3のインデックスは2のはず、got %d", parent.ChildIndex(child3.ID()))
	}
	if parent.ChildIndex(999) != -1 {
		t.Errorf("存在しないIDのインデックスは-1のはず、got %d", parent.ChildIndex(999))
	}
}

// TestPrintHierarchy はPrintHierarchyメソッドをテストする
func TestPrintHierarchy(t *testing.T) {
	sm := NewSpriteManager()

	root := sm.CreateRootSprite(nil)
	child := sm.CreateSprite(nil, root)
	sm.CreateSprite(nil, child)

	output := sm.PrintHierarchy()
	if output == "" {
		t.Error("PrintHierarchyは空でない文字列を返すはず")
	}
}

// TestPrintDrawOrder はPrintDrawOrderメソッドをテストする
func TestPrintDrawOrder(t *testing.T) {
	sm := NewSpriteManager()

	root := sm.CreateRootSprite(nil)
	sm.CreateSprite(nil, root)

	output := sm.PrintDrawOrder()
	if output == "" {
		t.Error("PrintDrawOrderは空でない文字列を返すはず")
	}
}
