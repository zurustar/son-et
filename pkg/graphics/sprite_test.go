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
	// Z_Pathは初期状態ではnil
	if s.GetZPath() != nil {
		t.Error("expected zPath nil")
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

	// 3つのスプライトを異なるZ_Pathで作成
	img1 := ebiten.NewImage(10, 10)
	img1.Fill(color.RGBA{255, 0, 0, 255})
	s1 := sm.CreateSprite(img1)
	s1.SetZPath(NewZPath(2))

	img2 := ebiten.NewImage(10, 10)
	img2.Fill(color.RGBA{0, 255, 0, 255})
	s2 := sm.CreateSprite(img2)
	s2.SetZPath(NewZPath(1))

	img3 := ebiten.NewImage(10, 10)
	img3.Fill(color.RGBA{0, 0, 255, 255})
	s3 := sm.CreateSprite(img3)
	s3.SetZPath(NewZPath(3))

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
	s.SetVisible(false)
	if !s.IsDirty() {
		t.Error("sprite should be dirty after SetVisible")
	}

	s.ClearDirty()
	s.SetAlpha(0.5)
	if !s.IsDirty() {
		t.Error("sprite should be dirty after SetAlpha")
	}

	s.ClearDirty()
	s.SetZPath(NewZPath(1, 2))
	if !s.IsDirty() {
		t.Error("sprite should be dirty after SetZPath")
	}
}

// TestSpriteChildManagement は子スプライト管理メソッドをテストする
// 要件 1.4: 子スプライトが作成されたとき、親のZ_Pathを継承し、自身のLocal_Z_Orderを追加する
// 要件 9.1: PictureSpriteは子スプライトを持てる
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
	if child2.Parent() != parent {
		t.Error("AddChild後、子のParent()は親を返すはず")
	}

	// RemoveChild: 子スプライトを削除
	parent.RemoveChild(child1.ID())
	if len(parent.GetChildren()) != 1 {
		t.Errorf("RemoveChild後、GetChildren()は1つの子を返すはず、got %d", len(parent.GetChildren()))
	}
	if child1.Parent() != nil {
		t.Error("RemoveChild後、削除された子のParent()はnilを返すはず")
	}
	// 残っている子はchild2
	if parent.GetChildren()[0] != child2 {
		t.Error("RemoveChild後、残っている子はchild2のはず")
	}

	// 最後の子を削除
	parent.RemoveChild(child2.ID())
	if parent.HasChildren() {
		t.Error("すべての子を削除後、HasChildren()はfalseを返すはず")
	}
	if len(parent.GetChildren()) != 0 {
		t.Errorf("すべての子を削除後、GetChildren()は空のスライスを返すはず、got %d", len(parent.GetChildren()))
	}
}

// TestSpriteAddChildNil はnilの子スプライトを追加しようとした場合のテスト
func TestSpriteAddChildNil(t *testing.T) {
	parent := NewSprite(1, nil)

	// nilを追加しても何も起こらない（パニックしない）
	parent.AddChild(nil)

	if parent.HasChildren() {
		t.Error("nilを追加してもHasChildren()はfalseを返すはず")
	}
}

// TestSpriteRemoveChildNotFound は存在しない子スプライトを削除しようとした場合のテスト
func TestSpriteRemoveChildNotFound(t *testing.T) {
	parent := NewSprite(1, nil)
	child := NewSprite(2, nil)
	parent.AddChild(child)

	// 存在しないIDで削除を試みる（パニックしない）
	parent.RemoveChild(999)

	// 子スプライトは削除されていない
	if len(parent.GetChildren()) != 1 {
		t.Errorf("存在しないIDで削除を試みても子スプライトは削除されないはず、got %d", len(parent.GetChildren()))
	}
}

// TestSpriteChildrenOrder は子スプライトの順序が保持されることをテストする
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

// TestSpriteManager_CreateRootSprite はルートスプライト作成をテストする
// 要件 1.3: Root_Spriteは単一要素のZ_Path（例: [0]）を持つ
// 要件 4.1: ウインドウをRoot_Spriteとして扱う
func TestSpriteManager_CreateRootSprite(t *testing.T) {
	sm := NewSpriteManager()

	// ウインドウ0を作成
	window0 := sm.CreateRootSprite(nil, 0)
	if window0 == nil {
		t.Fatal("CreateRootSpriteはnilを返さないはず")
	}
	if window0.GetZPath() == nil {
		t.Fatal("ルートスプライトはZ_Pathを持つはず")
	}
	if window0.GetZPath().Depth() != 1 {
		t.Errorf("ルートスプライトのZ_Pathの深さは1のはず、got %d", window0.GetZPath().Depth())
	}
	if window0.GetZPath().LocalZOrder() != 0 {
		t.Errorf("ウインドウ0のLocal_Z_Orderは0のはず、got %d", window0.GetZPath().LocalZOrder())
	}
	expectedPath := []int{0}
	if !equalIntSlice(window0.GetZPath().Path(), expectedPath) {
		t.Errorf("ウインドウ0のZ_Pathは[0]のはず、got %v", window0.GetZPath().Path())
	}

	// ウインドウ1を作成
	window1 := sm.CreateRootSprite(nil, 1)
	if window1.GetZPath().LocalZOrder() != 1 {
		t.Errorf("ウインドウ1のLocal_Z_Orderは1のはず、got %d", window1.GetZPath().LocalZOrder())
	}
	expectedPath = []int{1}
	if !equalIntSlice(window1.GetZPath().Path(), expectedPath) {
		t.Errorf("ウインドウ1のZ_Pathは[1]のはず、got %v", window1.GetZPath().Path())
	}

	// ウインドウ5を作成（任意のZ順序）
	window5 := sm.CreateRootSprite(nil, 5)
	expectedPath = []int{5}
	if !equalIntSlice(window5.GetZPath().Path(), expectedPath) {
		t.Errorf("ウインドウ5のZ_Pathは[5]のはず、got %v", window5.GetZPath().Path())
	}

	// スプライトマネージャーに登録されていることを確認
	if sm.Count() != 3 {
		t.Errorf("3つのスプライトが登録されているはず、got %d", sm.Count())
	}
}

// TestSpriteManager_CreateSpriteWithZPath は親子関係を持つスプライト作成をテストする
// 要件 1.2: Z_Pathは親のZ_Pathに自身のLocal_Z_Orderを追加した形式
// 要件 1.4: 子スプライトが作成されたとき、親のZ_Pathを継承し、自身のLocal_Z_Orderを追加する
// 要件 2.5: スプライトが作成されたとき、Z_Order_Counterをインクリメントする
func TestSpriteManager_CreateSpriteWithZPath(t *testing.T) {
	sm := NewSpriteManager()

	// ルートスプライト（ウインドウ）を作成
	window := sm.CreateRootSprite(nil, 0)

	// 子スプライト1を作成
	child1 := sm.CreateSpriteWithZPath(nil, window)
	if child1 == nil {
		t.Fatal("CreateSpriteWithZPathはnilを返さないはず")
	}
	if child1.GetZPath() == nil {
		t.Fatal("子スプライトはZ_Pathを持つはず")
	}
	if child1.GetZPath().Depth() != 2 {
		t.Errorf("子スプライトのZ_Pathの深さは2のはず、got %d", child1.GetZPath().Depth())
	}
	expectedPath := []int{0, 0}
	if !equalIntSlice(child1.GetZPath().Path(), expectedPath) {
		t.Errorf("child1のZ_Pathは[0, 0]のはず、got %v", child1.GetZPath().Path())
	}

	// 子スプライト2を作成（同じ親）
	child2 := sm.CreateSpriteWithZPath(nil, window)
	expectedPath = []int{0, 1}
	if !equalIntSlice(child2.GetZPath().Path(), expectedPath) {
		t.Errorf("child2のZ_Pathは[0, 1]のはず、got %v", child2.GetZPath().Path())
	}

	// 子スプライト3を作成（同じ親）
	child3 := sm.CreateSpriteWithZPath(nil, window)
	expectedPath = []int{0, 2}
	if !equalIntSlice(child3.GetZPath().Path(), expectedPath) {
		t.Errorf("child3のZ_Pathは[0, 2]のはず、got %v", child3.GetZPath().Path())
	}

	// 親子関係が設定されていることを確認
	if child1.Parent() != window {
		t.Error("child1の親はwindowのはず")
	}
	if child2.Parent() != window {
		t.Error("child2の親はwindowのはず")
	}
	if child3.Parent() != window {
		t.Error("child3の親はwindowのはず")
	}

	// windowの子スプライトリストに追加されていることを確認
	if len(window.GetChildren()) != 3 {
		t.Errorf("windowは3つの子スプライトを持つはず、got %d", len(window.GetChildren()))
	}
}

// TestSpriteManager_CreateSpriteWithZPath_NestedChildren は深い階層の子スプライト作成をテストする
// 要件 3.3: 任意の深さの親子関係に対応する
func TestSpriteManager_CreateSpriteWithZPath_NestedChildren(t *testing.T) {
	sm := NewSpriteManager()

	// ルートスプライト（ウインドウ）を作成
	window := sm.CreateRootSprite(nil, 0) // Z_Path: [0]

	// 子スプライトを作成
	child := sm.CreateSpriteWithZPath(nil, window) // Z_Path: [0, 0]

	// 孫スプライトを作成
	grandchild := sm.CreateSpriteWithZPath(nil, child) // Z_Path: [0, 0, 0]

	// 曾孫スプライトを作成
	greatGrandchild := sm.CreateSpriteWithZPath(nil, grandchild) // Z_Path: [0, 0, 0, 0]

	// Z_Pathの深さを確認
	if window.GetZPath().Depth() != 1 {
		t.Errorf("windowのZ_Pathの深さは1のはず、got %d", window.GetZPath().Depth())
	}
	if child.GetZPath().Depth() != 2 {
		t.Errorf("childのZ_Pathの深さは2のはず、got %d", child.GetZPath().Depth())
	}
	if grandchild.GetZPath().Depth() != 3 {
		t.Errorf("grandchildのZ_Pathの深さは3のはず、got %d", grandchild.GetZPath().Depth())
	}
	if greatGrandchild.GetZPath().Depth() != 4 {
		t.Errorf("greatGrandchildのZ_Pathの深さは4のはず、got %d", greatGrandchild.GetZPath().Depth())
	}

	// Z_Pathの値を確認
	expectedPath := []int{0, 0, 0, 0}
	if !equalIntSlice(greatGrandchild.GetZPath().Path(), expectedPath) {
		t.Errorf("greatGrandchildのZ_Pathは[0, 0, 0, 0]のはず、got %v", greatGrandchild.GetZPath().Path())
	}

	// 親のZ_Pathがプレフィックスであることを確認
	if !grandchild.GetZPath().IsPrefix(greatGrandchild.GetZPath()) {
		t.Error("grandchildのZ_PathはgreatGrandchildのZ_Pathのプレフィックスのはず")
	}
	if !child.GetZPath().IsPrefix(greatGrandchild.GetZPath()) {
		t.Error("childのZ_PathはgreatGrandchildのZ_Pathのプレフィックスのはず")
	}
	if !window.GetZPath().IsPrefix(greatGrandchild.GetZPath()) {
		t.Error("windowのZ_PathはgreatGrandchildのZ_Pathのプレフィックスのはず")
	}
}

// TestSpriteManager_CreateSpriteWithZPath_NilParent は親がnilの場合のテスト
func TestSpriteManager_CreateSpriteWithZPath_NilParent(t *testing.T) {
	sm := NewSpriteManager()

	// 親がnilの場合、ルートスプライトとして作成される
	sprite1 := sm.CreateSpriteWithZPath(nil, nil)
	expectedPath := []int{0}
	if !equalIntSlice(sprite1.GetZPath().Path(), expectedPath) {
		t.Errorf("親がnilの場合、Z_Pathは[0]のはず、got %v", sprite1.GetZPath().Path())
	}

	// 2つ目のルートスプライト
	sprite2 := sm.CreateSpriteWithZPath(nil, nil)
	expectedPath = []int{1}
	if !equalIntSlice(sprite2.GetZPath().Path(), expectedPath) {
		t.Errorf("2つ目の親がnilのスプライトのZ_Pathは[1]のはず、got %v", sprite2.GetZPath().Path())
	}
}

// TestSpriteManager_CreateSpriteWithZPath_MultipleWindows は複数ウインドウでの子スプライト作成をテストする
// 要件 4.2: ウインドウAがウインドウBより前面にあるとき、ウインドウAのすべての子スプライトをウインドウBのすべての子スプライトより前面に描画する
func TestSpriteManager_CreateSpriteWithZPath_MultipleWindows(t *testing.T) {
	sm := NewSpriteManager()

	// ウインドウ0を作成
	window0 := sm.CreateRootSprite(nil, 0) // Z_Path: [0]

	// ウインドウ1を作成
	window1 := sm.CreateRootSprite(nil, 1) // Z_Path: [1]

	// ウインドウ0の子スプライトを作成
	child0_1 := sm.CreateSpriteWithZPath(nil, window0) // Z_Path: [0, 0]
	child0_2 := sm.CreateSpriteWithZPath(nil, window0) // Z_Path: [0, 1]

	// ウインドウ1の子スプライトを作成
	child1_1 := sm.CreateSpriteWithZPath(nil, window1) // Z_Path: [1, 0]
	child1_2 := sm.CreateSpriteWithZPath(nil, window1) // Z_Path: [1, 1]

	// Z_Pathを確認
	expectedPath := []int{0, 0}
	if !equalIntSlice(child0_1.GetZPath().Path(), expectedPath) {
		t.Errorf("child0_1のZ_Pathは[0, 0]のはず、got %v", child0_1.GetZPath().Path())
	}
	expectedPath = []int{0, 1}
	if !equalIntSlice(child0_2.GetZPath().Path(), expectedPath) {
		t.Errorf("child0_2のZ_Pathは[0, 1]のはず、got %v", child0_2.GetZPath().Path())
	}
	expectedPath = []int{1, 0}
	if !equalIntSlice(child1_1.GetZPath().Path(), expectedPath) {
		t.Errorf("child1_1のZ_Pathは[1, 0]のはず、got %v", child1_1.GetZPath().Path())
	}
	expectedPath = []int{1, 1}
	if !equalIntSlice(child1_2.GetZPath().Path(), expectedPath) {
		t.Errorf("child1_2のZ_Pathは[1, 1]のはず、got %v", child1_2.GetZPath().Path())
	}

	// ウインドウ1の子スプライトはウインドウ0の子スプライトより前面（Z_Pathが大きい）
	if !child0_2.GetZPath().Less(child1_1.GetZPath()) {
		t.Error("ウインドウ0の子スプライトはウインドウ1の子スプライトより背面のはず")
	}
}

// TestSpriteManager_ZOrderCounter はZOrderCounterが正しく動作することをテストする
// 要件 2.1: 各親スプライトごとにZ_Order_Counterを管理する
func TestSpriteManager_ZOrderCounter(t *testing.T) {
	sm := NewSpriteManager()

	// ウインドウ0を作成
	window0 := sm.CreateRootSprite(nil, 0)

	// ウインドウ1を作成
	window1 := sm.CreateRootSprite(nil, 1)

	// ウインドウ0の子スプライトを作成
	child0_1 := sm.CreateSpriteWithZPath(nil, window0) // Local_Z_Order: 0
	child0_2 := sm.CreateSpriteWithZPath(nil, window0) // Local_Z_Order: 1

	// ウインドウ1の子スプライトを作成（別のカウンター）
	child1_1 := sm.CreateSpriteWithZPath(nil, window1) // Local_Z_Order: 0
	child1_2 := sm.CreateSpriteWithZPath(nil, window1) // Local_Z_Order: 1

	// Local_Z_Orderを確認
	if child0_1.GetZPath().LocalZOrder() != 0 {
		t.Errorf("child0_1のLocal_Z_Orderは0のはず、got %d", child0_1.GetZPath().LocalZOrder())
	}
	if child0_2.GetZPath().LocalZOrder() != 1 {
		t.Errorf("child0_2のLocal_Z_Orderは1のはず、got %d", child0_2.GetZPath().LocalZOrder())
	}
	if child1_1.GetZPath().LocalZOrder() != 0 {
		t.Errorf("child1_1のLocal_Z_Orderは0のはず、got %d", child1_1.GetZPath().LocalZOrder())
	}
	if child1_2.GetZPath().LocalZOrder() != 1 {
		t.Errorf("child1_2のLocal_Z_Orderは1のはず、got %d", child1_2.GetZPath().LocalZOrder())
	}
}

// equalIntSlice は2つのint配列が等しいかどうかを比較するヘルパー関数
func equalIntSlice(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TestSpriteManager_SortByZPath はZ_Pathによるソートが正しく動作することをテストする
// 要件 1.5: Z_Pathの辞書順比較でスプライトの描画順序を決定する
// 要件 7.1: Z_Pathのソート結果をキャッシュする
func TestSpriteManager_SortByZPath(t *testing.T) {
	sm := NewSpriteManager()

	// ウインドウ1を先に作成（Z_Path: [1]）
	window1 := sm.CreateRootSprite(nil, 1)

	// ウインドウ0を後に作成（Z_Path: [0]）
	window0 := sm.CreateRootSprite(nil, 0)

	// ウインドウ1の子スプライトを作成
	child1_0 := sm.CreateSpriteWithZPath(nil, window1) // Z_Path: [1, 0]
	child1_1 := sm.CreateSpriteWithZPath(nil, window1) // Z_Path: [1, 1]

	// ウインドウ0の子スプライトを作成
	child0_0 := sm.CreateSpriteWithZPath(nil, window0) // Z_Path: [0, 0]
	child0_1 := sm.CreateSpriteWithZPath(nil, window0) // Z_Path: [0, 1]

	// 描画を実行してソートをトリガー
	screen := ebiten.NewImage(100, 100)
	sm.Draw(screen)

	// ソート結果を確認（内部のsortedスライスにアクセスできないため、
	// Z_Pathの比較で正しい順序になることを確認）

	// 期待される描画順序:
	// 1. window0 [0]
	// 2. child0_0 [0, 0]
	// 3. child0_1 [0, 1]
	// 4. window1 [1]
	// 5. child1_0 [1, 0]
	// 6. child1_1 [1, 1]

	// Z_Pathの順序を確認
	if !window0.GetZPath().Less(child0_0.GetZPath()) {
		t.Error("window0はchild0_0より前に描画されるはず")
	}
	if !child0_0.GetZPath().Less(child0_1.GetZPath()) {
		t.Error("child0_0はchild0_1より前に描画されるはず")
	}
	if !child0_1.GetZPath().Less(window1.GetZPath()) {
		t.Error("child0_1はwindow1より前に描画されるはず")
	}
	if !window1.GetZPath().Less(child1_0.GetZPath()) {
		t.Error("window1はchild1_0より前に描画されるはず")
	}
	if !child1_0.GetZPath().Less(child1_1.GetZPath()) {
		t.Error("child1_0はchild1_1より前に描画されるはず")
	}
}

// TestSpriteManager_SortCacheValidity はソートキャッシュの有効性をテストする
// 要件 7.1: Z_Pathのソート結果をキャッシュする
// 要件 7.2: スプライトの変更時にソートが必要であることをマークする
func TestSpriteManager_SortCacheValidity(t *testing.T) {
	sm := NewSpriteManager()

	// スプライトを作成
	window := sm.CreateRootSprite(nil, 0)
	sm.CreateSpriteWithZPath(nil, window)
	sm.CreateSpriteWithZPath(nil, window)

	// 最初の描画でソートが実行される
	screen := ebiten.NewImage(100, 100)
	sm.Draw(screen)

	// 2回目の描画ではソートは実行されない（キャッシュが有効）
	// これはneedSortフラグがfalseになっていることで確認できる
	// 直接フラグにアクセスできないため、描画が正常に完了することを確認
	sm.Draw(screen)

	// 新しいスプライトを追加するとソートが必要になる
	sm.CreateSpriteWithZPath(nil, window)

	// 3回目の描画でソートが再実行される
	sm.Draw(screen)

	// スプライトを削除するとソートが必要になる
	sm.RemoveSprite(window.ID())

	// 4回目の描画でソートが再実行される
	sm.Draw(screen)
}

// TestSpriteManager_SortMixedZPathAndZOrder はZ_Pathを持つスプライトと持たないスプライトの混在をテストする
func TestSpriteManager_SortMixedZPathAndNoZPath(t *testing.T) {
	sm := NewSpriteManager()

	// Z_Pathを持たないスプライトを作成
	spriteNoZPath1 := sm.CreateSprite(nil)
	spriteNoZPath2 := sm.CreateSprite(nil)

	// Z_Pathを持つスプライトを作成
	window := sm.CreateRootSprite(nil, 0)
	child := sm.CreateSpriteWithZPath(nil, window)

	// 描画を実行してソートをトリガー
	screen := ebiten.NewImage(100, 100)
	sm.Draw(screen)

	// Z_Pathを持たないスプライトはZ_Pathを持つスプライトより先に描画される
	// spriteNoZPath1, spriteNoZPath2 (IDで比較) < window [0] < child [0, 0]

	// Z_Pathを持たないスプライト同士はIDで比較（安定ソート）
	if spriteNoZPath1.ID() >= spriteNoZPath2.ID() {
		t.Error("spriteNoZPath1はspriteNoZPath2より小さいIDを持つはず")
	}

	// Z_Pathを持つスプライトの順序を確認
	if !window.GetZPath().Less(child.GetZPath()) {
		t.Error("windowはchildより前に描画されるはず")
	}
}

// TestSpriteManager_MarkNeedSort はMarkNeedSortメソッドをテストする
// 要件 7.2: スプライトの変更時にソートが必要であることをマークする
func TestSpriteManager_MarkNeedSort(t *testing.T) {
	sm := NewSpriteManager()

	// スプライトを作成
	window := sm.CreateRootSprite(nil, 0)
	sm.CreateSpriteWithZPath(nil, window)

	// 描画を実行してソートを完了
	screen := ebiten.NewImage(100, 100)
	sm.Draw(screen)

	// MarkNeedSortを呼び出す
	sm.MarkNeedSort()

	// 次の描画でソートが再実行される（エラーなく完了することを確認）
	sm.Draw(screen)
}

// TestIsEffectivelyVisible_ParentVisibility は親スプライトの可視性が子スプライトに影響することをテストする
// 要件 3.4: 親スプライトが非表示のとき、子スプライトも描画しない
func TestIsEffectivelyVisible_ParentVisibility(t *testing.T) {
	// 親スプライトを作成
	parent := NewSprite(1, nil)
	parent.SetVisible(true)

	// 子スプライトを作成
	child := NewSprite(2, nil)
	child.SetVisible(true)
	child.SetParent(parent)

	// 孫スプライトを作成
	grandchild := NewSprite(3, nil)
	grandchild.SetVisible(true)
	grandchild.SetParent(child)

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

	// 親が非表示なので、子と孫も実効的に非表示
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

	// すべて可視に戻る
	if !parent.IsEffectivelyVisible() {
		t.Error("親スプライトは可視のはず")
	}
	if !child.IsEffectivelyVisible() {
		t.Error("子スプライトは可視のはず")
	}
	if !grandchild.IsEffectivelyVisible() {
		t.Error("孫スプライトは可視のはず")
	}

	// 中間の子を非表示にする
	child.SetVisible(false)

	// 親は可視、子と孫は非表示
	if !parent.IsEffectivelyVisible() {
		t.Error("親スプライトは可視のはず")
	}
	if child.IsEffectivelyVisible() {
		t.Error("子スプライトは非表示のはず")
	}
	if grandchild.IsEffectivelyVisible() {
		t.Error("子が非表示なので、孫スプライトは実効的に非表示のはず")
	}
}

// TestIsEffectivelyVisible_SelfVisibility は自身の可視性のみが影響する場合をテストする
func TestIsEffectivelyVisible_SelfVisibility(t *testing.T) {
	// 親なしのスプライト
	sprite := NewSprite(1, nil)
	sprite.SetVisible(true)

	if !sprite.IsEffectivelyVisible() {
		t.Error("可視スプライトはIsEffectivelyVisible()がtrueを返すはず")
	}

	sprite.SetVisible(false)
	if sprite.IsEffectivelyVisible() {
		t.Error("非表示スプライトはIsEffectivelyVisible()がfalseを返すはず")
	}
}

// TestDraw_ParentVisibilityAffectsChildren は親の可視性が子の描画に影響することをテストする
// 要件 3.4: 親スプライトが非表示のとき、子スプライトも描画しない
func TestDraw_ParentVisibilityAffectsChildren(t *testing.T) {
	sm := NewSpriteManager()

	// 親ウインドウを作成
	parentImg := ebiten.NewImage(100, 100)
	parent := sm.CreateRootSprite(parentImg, 0)

	// 子スプライトを作成
	childImg := ebiten.NewImage(50, 50)
	child := sm.CreateSpriteWithZPath(childImg, parent)

	// 孫スプライトを作成
	grandchildImg := ebiten.NewImage(25, 25)
	grandchild := sm.CreateSpriteWithZPath(grandchildImg, child)

	// 描画先のスクリーン
	screen := ebiten.NewImage(200, 200)

	// 初期状態: すべて可視で描画される
	if !parent.IsEffectivelyVisible() {
		t.Error("親スプライトは可視のはず")
	}
	if !child.IsEffectivelyVisible() {
		t.Error("子スプライトは可視のはず")
	}
	if !grandchild.IsEffectivelyVisible() {
		t.Error("孫スプライトは可視のはず")
	}

	// 描画を実行（エラーなく完了することを確認）
	sm.Draw(screen)

	// 親を非表示にする
	parent.SetVisible(false)
	sm.MarkNeedSort()

	// 親が非表示なので、子と孫も実効的に非表示
	if parent.IsEffectivelyVisible() {
		t.Error("親スプライトは非表示のはず")
	}
	if child.IsEffectivelyVisible() {
		t.Error("親が非表示なので、子スプライトは実効的に非表示のはず")
	}
	if grandchild.IsEffectivelyVisible() {
		t.Error("親が非表示なので、孫スプライトは実効的に非表示のはず")
	}

	// 描画を実行（非表示スプライトはスキップされる）
	sm.Draw(screen)

	// 親を再び可視にする
	parent.SetVisible(true)
	sm.MarkNeedSort()

	// すべて可視に戻る
	if !parent.IsEffectivelyVisible() {
		t.Error("親スプライトは可視のはず")
	}
	if !child.IsEffectivelyVisible() {
		t.Error("子スプライトは可視のはず")
	}
	if !grandchild.IsEffectivelyVisible() {
		t.Error("孫スプライトは可視のはず")
	}

	// 描画を実行
	sm.Draw(screen)
}

// TestDraw_ZPathOrderWithVisibility はZ_Path順描画と可視性の組み合わせをテストする
// 要件 3.1: 親スプライトを先に描画し、その後に子スプライトを描画する
// 要件 3.2: 同じ親を持つ子スプライトをLocal_Z_Order順で描画する
// 要件 3.4: 親スプライトが非表示のとき、子スプライトも描画しない
func TestDraw_ZPathOrderWithVisibility(t *testing.T) {
	sm := NewSpriteManager()

	// ウインドウ0を作成（背面）
	window0Img := ebiten.NewImage(100, 100)
	window0 := sm.CreateRootSprite(window0Img, 0)

	// ウインドウ1を作成（前面）
	window1Img := ebiten.NewImage(100, 100)
	window1 := sm.CreateRootSprite(window1Img, 1)

	// ウインドウ0の子スプライトを作成
	child0_0Img := ebiten.NewImage(50, 50)
	child0_0 := sm.CreateSpriteWithZPath(child0_0Img, window0)

	child0_1Img := ebiten.NewImage(50, 50)
	child0_1 := sm.CreateSpriteWithZPath(child0_1Img, window0)

	// ウインドウ1の子スプライトを作成
	child1_0Img := ebiten.NewImage(50, 50)
	child1_0 := sm.CreateSpriteWithZPath(child1_0Img, window1)

	// 描画先のスクリーン
	screen := ebiten.NewImage(200, 200)

	// 初期状態: すべて可視
	sm.Draw(screen)

	// Z_Pathの順序を確認
	// window0 [0] < child0_0 [0, 0] < child0_1 [0, 1] < window1 [1] < child1_0 [1, 0]
	if !window0.GetZPath().Less(child0_0.GetZPath()) {
		t.Error("window0はchild0_0より前に描画されるはず")
	}
	if !child0_0.GetZPath().Less(child0_1.GetZPath()) {
		t.Error("child0_0はchild0_1より前に描画されるはず")
	}
	if !child0_1.GetZPath().Less(window1.GetZPath()) {
		t.Error("child0_1はwindow1より前に描画されるはず")
	}
	if !window1.GetZPath().Less(child1_0.GetZPath()) {
		t.Error("window1はchild1_0より前に描画されるはず")
	}

	// ウインドウ0を非表示にする
	window0.SetVisible(false)

	// ウインドウ0とその子は実効的に非表示
	if window0.IsEffectivelyVisible() {
		t.Error("window0は非表示のはず")
	}
	if child0_0.IsEffectivelyVisible() {
		t.Error("child0_0は実効的に非表示のはず")
	}
	if child0_1.IsEffectivelyVisible() {
		t.Error("child0_1は実効的に非表示のはず")
	}

	// ウインドウ1とその子は可視のまま
	if !window1.IsEffectivelyVisible() {
		t.Error("window1は可視のはず")
	}
	if !child1_0.IsEffectivelyVisible() {
		t.Error("child1_0は可視のはず")
	}

	// 描画を実行（非表示スプライトはスキップされる）
	sm.Draw(screen)
}

// TestSpriteManager_BringToFront はBringToFrontメソッドをテストする
// 要件 8.4: スプライトを最前面に移動するメソッドを提供する
func TestSpriteManager_BringToFront(t *testing.T) {
	sm := NewSpriteManager()

	// ウインドウを作成
	window := sm.CreateRootSprite(nil, 0) // Z_Path: [0]

	// 子スプライトを作成
	child1 := sm.CreateSpriteWithZPath(nil, window) // Z_Path: [0, 0]
	child2 := sm.CreateSpriteWithZPath(nil, window) // Z_Path: [0, 1]
	child3 := sm.CreateSpriteWithZPath(nil, window) // Z_Path: [0, 2]

	// 初期状態を確認
	if child1.GetZPath().LocalZOrder() != 0 {
		t.Errorf("child1のLocal_Z_Orderは0のはず、got %d", child1.GetZPath().LocalZOrder())
	}
	if child2.GetZPath().LocalZOrder() != 1 {
		t.Errorf("child2のLocal_Z_Orderは1のはず、got %d", child2.GetZPath().LocalZOrder())
	}
	if child3.GetZPath().LocalZOrder() != 2 {
		t.Errorf("child3のLocal_Z_Orderは2のはず、got %d", child3.GetZPath().LocalZOrder())
	}

	// child1を最前面に移動
	err := sm.BringToFront(child1.ID())
	if err != nil {
		t.Fatalf("BringToFrontがエラーを返した: %v", err)
	}

	// child1のLocal_Z_Orderが最大になっていることを確認
	if child1.GetZPath().LocalZOrder() <= child2.GetZPath().LocalZOrder() {
		t.Errorf("BringToFront後、child1のLocal_Z_Orderはchild2より大きいはず: child1=%d, child2=%d",
			child1.GetZPath().LocalZOrder(), child2.GetZPath().LocalZOrder())
	}
	if child1.GetZPath().LocalZOrder() <= child3.GetZPath().LocalZOrder() {
		t.Errorf("BringToFront後、child1のLocal_Z_Orderはchild3より大きいはず: child1=%d, child3=%d",
			child1.GetZPath().LocalZOrder(), child3.GetZPath().LocalZOrder())
	}

	// Z_Pathの順序を確認（child1が最後に描画される）
	if !child2.GetZPath().Less(child1.GetZPath()) {
		t.Error("BringToFront後、child2はchild1より前に描画されるはず")
	}
	if !child3.GetZPath().Less(child1.GetZPath()) {
		t.Error("BringToFront後、child3はchild1より前に描画されるはず")
	}
}

// TestSpriteManager_BringToFront_NotFound は存在しないスプライトIDでBringToFrontを呼び出した場合のテスト
func TestSpriteManager_BringToFront_NotFound(t *testing.T) {
	sm := NewSpriteManager()

	// 存在しないIDでBringToFrontを呼び出す
	err := sm.BringToFront(999)
	if err == nil {
		t.Error("存在しないスプライトIDでBringToFrontを呼び出した場合、エラーを返すはず")
	}
}

// TestSpriteManager_BringToFront_RootSprite はルートスプライトでBringToFrontを呼び出した場合のテスト
// 注意: CreateRootSpriteはwindowZOrderを直接Z_Pathとして設定するため、
// ZOrderCounterを使用しません。BringToFrontはZOrderCounterを使用するため、
// 親がnilのスプライト（parentID=0）のカウンターから新しいLocal_Z_Orderを取得します。
func TestSpriteManager_BringToFront_RootSprite(t *testing.T) {
	sm := NewSpriteManager()

	// CreateSpriteWithZPathで親がnilのスプライトを作成（ZOrderCounterを使用）
	// これにより、BringToFrontと同じカウンターを使用する
	window0 := sm.CreateSpriteWithZPath(nil, nil) // Z_Path: [0]
	window1 := sm.CreateSpriteWithZPath(nil, nil) // Z_Path: [1]

	// 初期状態を確認
	if window0.GetZPath().LocalZOrder() != 0 {
		t.Errorf("window0のLocal_Z_Orderは0のはず、got %d", window0.GetZPath().LocalZOrder())
	}
	if window1.GetZPath().LocalZOrder() != 1 {
		t.Errorf("window1のLocal_Z_Orderは1のはず、got %d", window1.GetZPath().LocalZOrder())
	}

	// window0を最前面に移動
	err := sm.BringToFront(window0.ID())
	if err != nil {
		t.Fatalf("BringToFrontがエラーを返した: %v", err)
	}

	// window0のLocal_Z_Orderがwindow1より大きくなっていることを確認
	if window0.GetZPath().LocalZOrder() <= window1.GetZPath().LocalZOrder() {
		t.Errorf("BringToFront後、window0のLocal_Z_Orderはwindow1より大きいはず: window0=%d, window1=%d",
			window0.GetZPath().LocalZOrder(), window1.GetZPath().LocalZOrder())
	}

	// Z_Pathの順序を確認（window0が最後に描画される）
	if !window1.GetZPath().Less(window0.GetZPath()) {
		t.Error("BringToFront後、window1はwindow0より前に描画されるはず")
	}
}

// TestSpriteManager_BringToFront_WithChildren は子スプライトを持つスプライトでBringToFrontを呼び出した場合のテスト
// 要件 8.3: 親スプライトが変更されたとき、子スプライトのZ_Pathを再計算する
func TestSpriteManager_BringToFront_WithChildren(t *testing.T) {
	sm := NewSpriteManager()

	// ウインドウを作成
	window := sm.CreateRootSprite(nil, 0) // Z_Path: [0]

	// 親スプライトを作成
	parent1 := sm.CreateSpriteWithZPath(nil, window) // Z_Path: [0, 0]
	parent2 := sm.CreateSpriteWithZPath(nil, window) // Z_Path: [0, 1]

	// parent1の子スプライトを作成
	child1_1 := sm.CreateSpriteWithZPath(nil, parent1) // Z_Path: [0, 0, 0]
	child1_2 := sm.CreateSpriteWithZPath(nil, parent1) // Z_Path: [0, 0, 1]

	// 初期状態を確認
	expectedPath := []int{0, 0, 0}
	if !equalIntSlice(child1_1.GetZPath().Path(), expectedPath) {
		t.Errorf("child1_1のZ_Pathは[0, 0, 0]のはず、got %v", child1_1.GetZPath().Path())
	}
	expectedPath = []int{0, 0, 1}
	if !equalIntSlice(child1_2.GetZPath().Path(), expectedPath) {
		t.Errorf("child1_2のZ_Pathは[0, 0, 1]のはず、got %v", child1_2.GetZPath().Path())
	}

	// parent1を最前面に移動
	err := sm.BringToFront(parent1.ID())
	if err != nil {
		t.Fatalf("BringToFrontがエラーを返した: %v", err)
	}

	// parent1のLocal_Z_Orderがparent2より大きくなっていることを確認
	if parent1.GetZPath().LocalZOrder() <= parent2.GetZPath().LocalZOrder() {
		t.Errorf("BringToFront後、parent1のLocal_Z_Orderはparent2より大きいはず: parent1=%d, parent2=%d",
			parent1.GetZPath().LocalZOrder(), parent2.GetZPath().LocalZOrder())
	}

	// 子スプライトのZ_Pathが更新されていることを確認
	// parent1のZ_Pathが[0, 2]になったとすると、child1_1は[0, 2, 0]、child1_2は[0, 2, 1]になるはず
	newParentLocalZOrder := parent1.GetZPath().LocalZOrder()
	expectedPath = []int{0, newParentLocalZOrder, 0}
	if !equalIntSlice(child1_1.GetZPath().Path(), expectedPath) {
		t.Errorf("BringToFront後、child1_1のZ_Pathは%vのはず、got %v", expectedPath, child1_1.GetZPath().Path())
	}
	expectedPath = []int{0, newParentLocalZOrder, 1}
	if !equalIntSlice(child1_2.GetZPath().Path(), expectedPath) {
		t.Errorf("BringToFront後、child1_2のZ_Pathは%vのはず、got %v", expectedPath, child1_2.GetZPath().Path())
	}

	// 子スプライトのLocal_Z_Orderは変わらないことを確認
	if child1_1.GetZPath().LocalZOrder() != 0 {
		t.Errorf("BringToFront後、child1_1のLocal_Z_Orderは0のままのはず、got %d", child1_1.GetZPath().LocalZOrder())
	}
	if child1_2.GetZPath().LocalZOrder() != 1 {
		t.Errorf("BringToFront後、child1_2のLocal_Z_Orderは1のままのはず、got %d", child1_2.GetZPath().LocalZOrder())
	}

	// Z_Pathの順序を確認（parent1とその子がparent2より後に描画される）
	if !parent2.GetZPath().Less(parent1.GetZPath()) {
		t.Error("BringToFront後、parent2はparent1より前に描画されるはず")
	}
	if !parent2.GetZPath().Less(child1_1.GetZPath()) {
		t.Error("BringToFront後、parent2はchild1_1より前に描画されるはず")
	}
	if !parent2.GetZPath().Less(child1_2.GetZPath()) {
		t.Error("BringToFront後、parent2はchild1_2より前に描画されるはず")
	}
}

// TestSpriteManager_BringToFront_DeepHierarchy は深い階層でBringToFrontを呼び出した場合のテスト
// 要件 8.3: 親スプライトが変更されたとき、子スプライトのZ_Pathを再計算する
func TestSpriteManager_BringToFront_DeepHierarchy(t *testing.T) {
	sm := NewSpriteManager()

	// ウインドウを作成
	window := sm.CreateRootSprite(nil, 0) // Z_Path: [0]

	// 深い階層を作成
	level1_1 := sm.CreateSpriteWithZPath(nil, window) // Z_Path: [0, 0]
	level1_2 := sm.CreateSpriteWithZPath(nil, window) // Z_Path: [0, 1]
	level2 := sm.CreateSpriteWithZPath(nil, level1_1) // Z_Path: [0, 0, 0]
	level3 := sm.CreateSpriteWithZPath(nil, level2)   // Z_Path: [0, 0, 0, 0]
	level4 := sm.CreateSpriteWithZPath(nil, level3)   // Z_Path: [0, 0, 0, 0, 0]

	// 初期状態を確認
	if level4.GetZPath().Depth() != 5 {
		t.Errorf("level4のZ_Pathの深さは5のはず、got %d", level4.GetZPath().Depth())
	}

	// level1_1を最前面に移動
	err := sm.BringToFront(level1_1.ID())
	if err != nil {
		t.Fatalf("BringToFrontがエラーを返した: %v", err)
	}

	// level1_1のLocal_Z_Orderがlevel1_2より大きくなっていることを確認
	if level1_1.GetZPath().LocalZOrder() <= level1_2.GetZPath().LocalZOrder() {
		t.Errorf("BringToFront後、level1_1のLocal_Z_Orderはlevel1_2より大きいはず")
	}

	// すべての子孫スプライトのZ_Pathが更新されていることを確認
	newLevel1LocalZOrder := level1_1.GetZPath().LocalZOrder()

	// level2のZ_Pathを確認
	expectedPath := []int{0, newLevel1LocalZOrder, 0}
	if !equalIntSlice(level2.GetZPath().Path(), expectedPath) {
		t.Errorf("BringToFront後、level2のZ_Pathは%vのはず、got %v", expectedPath, level2.GetZPath().Path())
	}

	// level3のZ_Pathを確認
	expectedPath = []int{0, newLevel1LocalZOrder, 0, 0}
	if !equalIntSlice(level3.GetZPath().Path(), expectedPath) {
		t.Errorf("BringToFront後、level3のZ_Pathは%vのはず、got %v", expectedPath, level3.GetZPath().Path())
	}

	// level4のZ_Pathを確認
	expectedPath = []int{0, newLevel1LocalZOrder, 0, 0, 0}
	if !equalIntSlice(level4.GetZPath().Path(), expectedPath) {
		t.Errorf("BringToFront後、level4のZ_Pathは%vのはず、got %v", expectedPath, level4.GetZPath().Path())
	}

	// 深さは変わらないことを確認
	if level4.GetZPath().Depth() != 5 {
		t.Errorf("BringToFront後、level4のZ_Pathの深さは5のままのはず、got %d", level4.GetZPath().Depth())
	}
}

// TestSpriteManager_BringToFront_MultipleCalls は複数回BringToFrontを呼び出した場合のテスト
func TestSpriteManager_BringToFront_MultipleCalls(t *testing.T) {
	sm := NewSpriteManager()

	// ウインドウを作成
	window := sm.CreateRootSprite(nil, 0) // Z_Path: [0]

	// 子スプライトを作成
	child1 := sm.CreateSpriteWithZPath(nil, window) // Z_Path: [0, 0]
	child2 := sm.CreateSpriteWithZPath(nil, window) // Z_Path: [0, 1]
	child3 := sm.CreateSpriteWithZPath(nil, window) // Z_Path: [0, 2]

	// child1を最前面に移動
	err := sm.BringToFront(child1.ID())
	if err != nil {
		t.Fatalf("BringToFrontがエラーを返した: %v", err)
	}

	// child2を最前面に移動
	err = sm.BringToFront(child2.ID())
	if err != nil {
		t.Fatalf("BringToFrontがエラーを返した: %v", err)
	}

	// child2が最前面になっていることを確認
	if child2.GetZPath().LocalZOrder() <= child1.GetZPath().LocalZOrder() {
		t.Errorf("2回目のBringToFront後、child2のLocal_Z_Orderはchild1より大きいはず")
	}
	if child2.GetZPath().LocalZOrder() <= child3.GetZPath().LocalZOrder() {
		t.Errorf("2回目のBringToFront後、child2のLocal_Z_Orderはchild3より大きいはず")
	}

	// child3を最前面に移動
	err = sm.BringToFront(child3.ID())
	if err != nil {
		t.Fatalf("BringToFrontがエラーを返した: %v", err)
	}

	// child3が最前面になっていることを確認
	if child3.GetZPath().LocalZOrder() <= child1.GetZPath().LocalZOrder() {
		t.Errorf("3回目のBringToFront後、child3のLocal_Z_Orderはchild1より大きいはず")
	}
	if child3.GetZPath().LocalZOrder() <= child2.GetZPath().LocalZOrder() {
		t.Errorf("3回目のBringToFront後、child3のLocal_Z_Orderはchild2より大きいはず")
	}

	// Z_Pathの順序を確認
	// child1 < child2 < child3 の順で描画されるはず
	if !child1.GetZPath().Less(child2.GetZPath()) {
		t.Error("child1はchild2より前に描画されるはず")
	}
	if !child2.GetZPath().Less(child3.GetZPath()) {
		t.Error("child2はchild3より前に描画されるはず")
	}
}

// TestSpriteManager_SendToBack はSendToBackメソッドをテストする
// 要件 8.5: スプライトを最背面に移動するメソッドを提供する
func TestSpriteManager_SendToBack(t *testing.T) {
	sm := NewSpriteManager()

	// ウインドウを作成
	window := sm.CreateRootSprite(nil, 0) // Z_Path: [0]

	// 子スプライトを作成
	child1 := sm.CreateSpriteWithZPath(nil, window) // Z_Path: [0, 0]
	child2 := sm.CreateSpriteWithZPath(nil, window) // Z_Path: [0, 1]
	child3 := sm.CreateSpriteWithZPath(nil, window) // Z_Path: [0, 2]

	// 初期状態を確認
	if child1.GetZPath().LocalZOrder() != 0 {
		t.Errorf("child1のLocal_Z_Orderは0のはず、got %d", child1.GetZPath().LocalZOrder())
	}
	if child2.GetZPath().LocalZOrder() != 1 {
		t.Errorf("child2のLocal_Z_Orderは1のはず、got %d", child2.GetZPath().LocalZOrder())
	}
	if child3.GetZPath().LocalZOrder() != 2 {
		t.Errorf("child3のLocal_Z_Orderは2のはず、got %d", child3.GetZPath().LocalZOrder())
	}

	// child3を最背面に移動
	err := sm.SendToBack(child3.ID())
	if err != nil {
		t.Fatalf("SendToBackがエラーを返した: %v", err)
	}

	// child3のLocal_Z_Orderが最小になっていることを確認
	if child3.GetZPath().LocalZOrder() >= child1.GetZPath().LocalZOrder() {
		t.Errorf("SendToBack後、child3のLocal_Z_Orderはchild1より小さいはず: child3=%d, child1=%d",
			child3.GetZPath().LocalZOrder(), child1.GetZPath().LocalZOrder())
	}
	if child3.GetZPath().LocalZOrder() >= child2.GetZPath().LocalZOrder() {
		t.Errorf("SendToBack後、child3のLocal_Z_Orderはchild2より小さいはず: child3=%d, child2=%d",
			child3.GetZPath().LocalZOrder(), child2.GetZPath().LocalZOrder())
	}

	// Z_Pathの順序を確認（child3が最初に描画される）
	if !child3.GetZPath().Less(child1.GetZPath()) {
		t.Error("SendToBack後、child3はchild1より前に描画されるはず")
	}
	if !child3.GetZPath().Less(child2.GetZPath()) {
		t.Error("SendToBack後、child3はchild2より前に描画されるはず")
	}
}

// TestSpriteManager_SendToBack_NotFound は存在しないスプライトIDでSendToBackを呼び出した場合のテスト
func TestSpriteManager_SendToBack_NotFound(t *testing.T) {
	sm := NewSpriteManager()

	// 存在しないIDでSendToBackを呼び出す
	err := sm.SendToBack(999)
	if err == nil {
		t.Error("存在しないスプライトIDでSendToBackを呼び出した場合、エラーを返すはず")
	}
}

// TestSpriteManager_SendToBack_RootSprite はルートスプライトでSendToBackを呼び出した場合のテスト
func TestSpriteManager_SendToBack_RootSprite(t *testing.T) {
	sm := NewSpriteManager()

	// CreateSpriteWithZPathで親がnilのスプライトを作成（ZOrderCounterを使用）
	window0 := sm.CreateSpriteWithZPath(nil, nil) // Z_Path: [0]
	window1 := sm.CreateSpriteWithZPath(nil, nil) // Z_Path: [1]

	// 初期状態を確認
	if window0.GetZPath().LocalZOrder() != 0 {
		t.Errorf("window0のLocal_Z_Orderは0のはず、got %d", window0.GetZPath().LocalZOrder())
	}
	if window1.GetZPath().LocalZOrder() != 1 {
		t.Errorf("window1のLocal_Z_Orderは1のはず、got %d", window1.GetZPath().LocalZOrder())
	}

	// window1を最背面に移動
	err := sm.SendToBack(window1.ID())
	if err != nil {
		t.Fatalf("SendToBackがエラーを返した: %v", err)
	}

	// window1のLocal_Z_Orderがwindow0より小さくなっていることを確認
	if window1.GetZPath().LocalZOrder() >= window0.GetZPath().LocalZOrder() {
		t.Errorf("SendToBack後、window1のLocal_Z_Orderはwindow0より小さいはず: window1=%d, window0=%d",
			window1.GetZPath().LocalZOrder(), window0.GetZPath().LocalZOrder())
	}

	// Z_Pathの順序を確認（window1が最初に描画される）
	if !window1.GetZPath().Less(window0.GetZPath()) {
		t.Error("SendToBack後、window1はwindow0より前に描画されるはず")
	}
}

// TestSpriteManager_SendToBack_WithChildren は子スプライトを持つスプライトでSendToBackを呼び出した場合のテスト
// 要件 8.3: 親スプライトが変更されたとき、子スプライトのZ_Pathを再計算する
func TestSpriteManager_SendToBack_WithChildren(t *testing.T) {
	sm := NewSpriteManager()

	// ウインドウを作成
	window := sm.CreateRootSprite(nil, 0) // Z_Path: [0]

	// 親スプライトを作成
	parent1 := sm.CreateSpriteWithZPath(nil, window) // Z_Path: [0, 0]
	parent2 := sm.CreateSpriteWithZPath(nil, window) // Z_Path: [0, 1]

	// parent2の子スプライトを作成
	child2_1 := sm.CreateSpriteWithZPath(nil, parent2) // Z_Path: [0, 1, 0]
	child2_2 := sm.CreateSpriteWithZPath(nil, parent2) // Z_Path: [0, 1, 1]

	// 初期状態を確認
	expectedPath := []int{0, 1, 0}
	if !equalIntSlice(child2_1.GetZPath().Path(), expectedPath) {
		t.Errorf("child2_1のZ_Pathは[0, 1, 0]のはず、got %v", child2_1.GetZPath().Path())
	}
	expectedPath = []int{0, 1, 1}
	if !equalIntSlice(child2_2.GetZPath().Path(), expectedPath) {
		t.Errorf("child2_2のZ_Pathは[0, 1, 1]のはず、got %v", child2_2.GetZPath().Path())
	}

	// parent2を最背面に移動
	err := sm.SendToBack(parent2.ID())
	if err != nil {
		t.Fatalf("SendToBackがエラーを返した: %v", err)
	}

	// parent2のLocal_Z_Orderがparent1より小さくなっていることを確認
	if parent2.GetZPath().LocalZOrder() >= parent1.GetZPath().LocalZOrder() {
		t.Errorf("SendToBack後、parent2のLocal_Z_Orderはparent1より小さいはず: parent2=%d, parent1=%d",
			parent2.GetZPath().LocalZOrder(), parent1.GetZPath().LocalZOrder())
	}

	// 子スプライトのZ_Pathが更新されていることを確認
	newParentLocalZOrder := parent2.GetZPath().LocalZOrder()
	expectedPath = []int{0, newParentLocalZOrder, 0}
	if !equalIntSlice(child2_1.GetZPath().Path(), expectedPath) {
		t.Errorf("SendToBack後、child2_1のZ_Pathは%vのはず、got %v", expectedPath, child2_1.GetZPath().Path())
	}
	expectedPath = []int{0, newParentLocalZOrder, 1}
	if !equalIntSlice(child2_2.GetZPath().Path(), expectedPath) {
		t.Errorf("SendToBack後、child2_2のZ_Pathは%vのはず、got %v", expectedPath, child2_2.GetZPath().Path())
	}

	// 子スプライトのLocal_Z_Orderは変わらないことを確認
	if child2_1.GetZPath().LocalZOrder() != 0 {
		t.Errorf("SendToBack後、child2_1のLocal_Z_Orderは0のままのはず、got %d", child2_1.GetZPath().LocalZOrder())
	}
	if child2_2.GetZPath().LocalZOrder() != 1 {
		t.Errorf("SendToBack後、child2_2のLocal_Z_Orderは1のままのはず、got %d", child2_2.GetZPath().LocalZOrder())
	}

	// Z_Pathの順序を確認（parent2とその子がparent1より前に描画される）
	if !parent2.GetZPath().Less(parent1.GetZPath()) {
		t.Error("SendToBack後、parent2はparent1より前に描画されるはず")
	}
	if !child2_1.GetZPath().Less(parent1.GetZPath()) {
		t.Error("SendToBack後、child2_1はparent1より前に描画されるはず")
	}
	if !child2_2.GetZPath().Less(parent1.GetZPath()) {
		t.Error("SendToBack後、child2_2はparent1より前に描画されるはず")
	}
}

// TestSpriteManager_SendToBack_DeepHierarchy は深い階層でSendToBackを呼び出した場合のテスト
// 要件 8.3: 親スプライトが変更されたとき、子スプライトのZ_Pathを再計算する
func TestSpriteManager_SendToBack_DeepHierarchy(t *testing.T) {
	sm := NewSpriteManager()

	// ウインドウを作成
	window := sm.CreateRootSprite(nil, 0) // Z_Path: [0]

	// 深い階層を作成
	level1_1 := sm.CreateSpriteWithZPath(nil, window) // Z_Path: [0, 0]
	level1_2 := sm.CreateSpriteWithZPath(nil, window) // Z_Path: [0, 1]
	level2 := sm.CreateSpriteWithZPath(nil, level1_2) // Z_Path: [0, 1, 0]
	level3 := sm.CreateSpriteWithZPath(nil, level2)   // Z_Path: [0, 1, 0, 0]
	level4 := sm.CreateSpriteWithZPath(nil, level3)   // Z_Path: [0, 1, 0, 0, 0]

	// 初期状態を確認
	if level4.GetZPath().Depth() != 5 {
		t.Errorf("level4のZ_Pathの深さは5のはず、got %d", level4.GetZPath().Depth())
	}

	// level1_2を最背面に移動
	err := sm.SendToBack(level1_2.ID())
	if err != nil {
		t.Fatalf("SendToBackがエラーを返した: %v", err)
	}

	// level1_2のLocal_Z_Orderがlevel1_1より小さくなっていることを確認
	if level1_2.GetZPath().LocalZOrder() >= level1_1.GetZPath().LocalZOrder() {
		t.Errorf("SendToBack後、level1_2のLocal_Z_Orderはlevel1_1より小さいはず")
	}

	// すべての子孫スプライトのZ_Pathが更新されていることを確認
	newLevel1LocalZOrder := level1_2.GetZPath().LocalZOrder()

	// level2のZ_Pathを確認
	expectedPath := []int{0, newLevel1LocalZOrder, 0}
	if !equalIntSlice(level2.GetZPath().Path(), expectedPath) {
		t.Errorf("SendToBack後、level2のZ_Pathは%vのはず、got %v", expectedPath, level2.GetZPath().Path())
	}

	// level3のZ_Pathを確認
	expectedPath = []int{0, newLevel1LocalZOrder, 0, 0}
	if !equalIntSlice(level3.GetZPath().Path(), expectedPath) {
		t.Errorf("SendToBack後、level3のZ_Pathは%vのはず、got %v", expectedPath, level3.GetZPath().Path())
	}

	// level4のZ_Pathを確認
	expectedPath = []int{0, newLevel1LocalZOrder, 0, 0, 0}
	if !equalIntSlice(level4.GetZPath().Path(), expectedPath) {
		t.Errorf("SendToBack後、level4のZ_Pathは%vのはず、got %v", expectedPath, level4.GetZPath().Path())
	}

	// 深さは変わらないことを確認
	if level4.GetZPath().Depth() != 5 {
		t.Errorf("SendToBack後、level4のZ_Pathの深さは5のままのはず、got %d", level4.GetZPath().Depth())
	}
}

// TestSpriteManager_SendToBack_MultipleCalls は複数回SendToBackを呼び出した場合のテスト
func TestSpriteManager_SendToBack_MultipleCalls(t *testing.T) {
	sm := NewSpriteManager()

	// ウインドウを作成
	window := sm.CreateRootSprite(nil, 0) // Z_Path: [0]

	// 子スプライトを作成
	child1 := sm.CreateSpriteWithZPath(nil, window) // Z_Path: [0, 0]
	child2 := sm.CreateSpriteWithZPath(nil, window) // Z_Path: [0, 1]
	child3 := sm.CreateSpriteWithZPath(nil, window) // Z_Path: [0, 2]

	// child3を最背面に移動
	err := sm.SendToBack(child3.ID())
	if err != nil {
		t.Fatalf("SendToBackがエラーを返した: %v", err)
	}

	// child2を最背面に移動
	err = sm.SendToBack(child2.ID())
	if err != nil {
		t.Fatalf("SendToBackがエラーを返した: %v", err)
	}

	// child2が最背面になっていることを確認
	if child2.GetZPath().LocalZOrder() >= child3.GetZPath().LocalZOrder() {
		t.Errorf("2回目のSendToBack後、child2のLocal_Z_Orderはchild3より小さいはず")
	}
	if child2.GetZPath().LocalZOrder() >= child1.GetZPath().LocalZOrder() {
		t.Errorf("2回目のSendToBack後、child2のLocal_Z_Orderはchild1より小さいはず")
	}

	// child1を最背面に移動
	err = sm.SendToBack(child1.ID())
	if err != nil {
		t.Fatalf("SendToBackがエラーを返した: %v", err)
	}

	// child1が最背面になっていることを確認
	if child1.GetZPath().LocalZOrder() >= child2.GetZPath().LocalZOrder() {
		t.Errorf("3回目のSendToBack後、child1のLocal_Z_Orderはchild2より小さいはず")
	}
	if child1.GetZPath().LocalZOrder() >= child3.GetZPath().LocalZOrder() {
		t.Errorf("3回目のSendToBack後、child1のLocal_Z_Orderはchild3より小さいはず")
	}

	// Z_Pathの順序を確認
	// child1 < child2 < child3 の順で描画されるはず
	if !child1.GetZPath().Less(child2.GetZPath()) {
		t.Error("child1はchild2より前に描画されるはず")
	}
	if !child2.GetZPath().Less(child3.GetZPath()) {
		t.Error("child2はchild3より前に描画されるはず")
	}
}

// TestSpriteManager_SendToBack_SingleSprite は兄弟がいない場合のSendToBackをテストする
func TestSpriteManager_SendToBack_SingleSprite(t *testing.T) {
	sm := NewSpriteManager()

	// ウインドウを作成
	window := sm.CreateRootSprite(nil, 0) // Z_Path: [0]

	// 子スプライトを1つだけ作成
	child := sm.CreateSpriteWithZPath(nil, window) // Z_Path: [0, 0]

	// 初期状態を確認
	initialLocalZOrder := child.GetZPath().LocalZOrder()

	// childを最背面に移動（兄弟がいないので、最小値-1になる）
	err := sm.SendToBack(child.ID())
	if err != nil {
		t.Fatalf("SendToBackがエラーを返した: %v", err)
	}

	// Local_Z_Orderが変更されていることを確認（0 - 1 = -1）
	if child.GetZPath().LocalZOrder() >= initialLocalZOrder {
		t.Errorf("SendToBack後、childのLocal_Z_Orderは初期値より小さいはず: got %d, initial %d",
			child.GetZPath().LocalZOrder(), initialLocalZOrder)
	}
}

// TestSpriteManager_BringToFront_And_SendToBack はBringToFrontとSendToBackを組み合わせたテスト
func TestSpriteManager_BringToFront_And_SendToBack(t *testing.T) {
	sm := NewSpriteManager()

	// ウインドウを作成
	window := sm.CreateRootSprite(nil, 0) // Z_Path: [0]

	// 子スプライトを作成
	child1 := sm.CreateSpriteWithZPath(nil, window) // Z_Path: [0, 0]
	child2 := sm.CreateSpriteWithZPath(nil, window) // Z_Path: [0, 1]
	child3 := sm.CreateSpriteWithZPath(nil, window) // Z_Path: [0, 2]

	// child1を最前面に移動
	err := sm.BringToFront(child1.ID())
	if err != nil {
		t.Fatalf("BringToFrontがエラーを返した: %v", err)
	}

	// child1が最前面になっていることを確認
	if !child2.GetZPath().Less(child1.GetZPath()) {
		t.Error("BringToFront後、child2はchild1より前に描画されるはず")
	}
	if !child3.GetZPath().Less(child1.GetZPath()) {
		t.Error("BringToFront後、child3はchild1より前に描画されるはず")
	}

	// child1を最背面に移動
	err = sm.SendToBack(child1.ID())
	if err != nil {
		t.Fatalf("SendToBackがエラーを返した: %v", err)
	}

	// child1が最背面になっていることを確認
	if !child1.GetZPath().Less(child2.GetZPath()) {
		t.Error("SendToBack後、child1はchild2より前に描画されるはず")
	}
	if !child1.GetZPath().Less(child3.GetZPath()) {
		t.Error("SendToBack後、child1はchild3より前に描画されるはず")
	}
}

// TestUpdateChildrenZPaths_DirectCall はupdateChildrenZPathsの直接呼び出しをテストする
// 要件 8.3: 親スプライトが変更されたとき、子スプライトのZ_Pathを再計算する
func TestUpdateChildrenZPaths_DirectCall(t *testing.T) {
	sm := NewSpriteManager()

	// ウインドウを作成
	window := sm.CreateRootSprite(nil, 0) // Z_Path: [0]

	// 親スプライトを作成
	parent := sm.CreateSpriteWithZPath(nil, window) // Z_Path: [0, 0]

	// 子スプライトを作成
	child1 := sm.CreateSpriteWithZPath(nil, parent) // Z_Path: [0, 0, 0]
	child2 := sm.CreateSpriteWithZPath(nil, parent) // Z_Path: [0, 0, 1]

	// 初期状態を確認
	expectedPath := []int{0, 0, 0}
	if !equalIntSlice(child1.GetZPath().Path(), expectedPath) {
		t.Errorf("child1のZ_Pathは[0, 0, 0]のはず、got %v", child1.GetZPath().Path())
	}
	expectedPath = []int{0, 0, 1}
	if !equalIntSlice(child2.GetZPath().Path(), expectedPath) {
		t.Errorf("child2のZ_Pathは[0, 0, 1]のはず、got %v", child2.GetZPath().Path())
	}

	// 親のZ_Pathを直接変更
	parent.SetZPath(NewZPath(0, 5))

	// updateChildrenZPathsを呼び出す（BringToFrontやSendToBackを経由せずに直接テスト）
	sm.UpdateChildrenZPathsForTest(parent)

	// 子スプライトのZ_Pathが更新されていることを確認
	expectedPath = []int{0, 5, 0}
	if !equalIntSlice(child1.GetZPath().Path(), expectedPath) {
		t.Errorf("親のZ_Path変更後、child1のZ_Pathは[0, 5, 0]のはず、got %v", child1.GetZPath().Path())
	}
	expectedPath = []int{0, 5, 1}
	if !equalIntSlice(child2.GetZPath().Path(), expectedPath) {
		t.Errorf("親のZ_Path変更後、child2のZ_Pathは[0, 5, 1]のはず、got %v", child2.GetZPath().Path())
	}

	// Local_Z_Orderは保持されていることを確認
	if child1.GetZPath().LocalZOrder() != 0 {
		t.Errorf("child1のLocal_Z_Orderは0のままのはず、got %d", child1.GetZPath().LocalZOrder())
	}
	if child2.GetZPath().LocalZOrder() != 1 {
		t.Errorf("child2のLocal_Z_Orderは1のままのはず、got %d", child2.GetZPath().LocalZOrder())
	}
}

// TestUpdateChildrenZPaths_RecursiveUpdate は再帰的な更新をテストする
// 要件 8.3: 親スプライトが変更されたとき、子スプライトのZ_Pathを再計算する
func TestUpdateChildrenZPaths_RecursiveUpdate(t *testing.T) {
	sm := NewSpriteManager()

	// ウインドウを作成
	window := sm.CreateRootSprite(nil, 0) // Z_Path: [0]

	// 深い階層を作成
	level1 := sm.CreateSpriteWithZPath(nil, window) // Z_Path: [0, 0]
	level2 := sm.CreateSpriteWithZPath(nil, level1) // Z_Path: [0, 0, 0]
	level3 := sm.CreateSpriteWithZPath(nil, level2) // Z_Path: [0, 0, 0, 0]
	level4 := sm.CreateSpriteWithZPath(nil, level3) // Z_Path: [0, 0, 0, 0, 0]

	// 初期状態を確認
	expectedPath := []int{0, 0, 0, 0, 0}
	if !equalIntSlice(level4.GetZPath().Path(), expectedPath) {
		t.Errorf("level4のZ_Pathは[0, 0, 0, 0, 0]のはず、got %v", level4.GetZPath().Path())
	}

	// level1のZ_Pathを直接変更
	level1.SetZPath(NewZPath(0, 10))

	// updateChildrenZPathsを呼び出す
	sm.UpdateChildrenZPathsForTest(level1)

	// すべての子孫スプライトのZ_Pathが更新されていることを確認
	expectedPath = []int{0, 10, 0}
	if !equalIntSlice(level2.GetZPath().Path(), expectedPath) {
		t.Errorf("level2のZ_Pathは[0, 10, 0]のはず、got %v", level2.GetZPath().Path())
	}
	expectedPath = []int{0, 10, 0, 0}
	if !equalIntSlice(level3.GetZPath().Path(), expectedPath) {
		t.Errorf("level3のZ_Pathは[0, 10, 0, 0]のはず、got %v", level3.GetZPath().Path())
	}
	expectedPath = []int{0, 10, 0, 0, 0}
	if !equalIntSlice(level4.GetZPath().Path(), expectedPath) {
		t.Errorf("level4のZ_Pathは[0, 10, 0, 0, 0]のはず、got %v", level4.GetZPath().Path())
	}

	// 深さは変わらないことを確認
	if level4.GetZPath().Depth() != 5 {
		t.Errorf("level4のZ_Pathの深さは5のままのはず、got %d", level4.GetZPath().Depth())
	}
}

// TestUpdateChildrenZPaths_PreservesLocalZOrder はLocal_Z_Orderが保持されることをテストする
// 要件 8.3: 親スプライトが変更されたとき、子スプライトのZ_Pathを再計算する
func TestUpdateChildrenZPaths_PreservesLocalZOrder(t *testing.T) {
	sm := NewSpriteManager()

	// ウインドウを作成
	window := sm.CreateRootSprite(nil, 0) // Z_Path: [0]

	// 親スプライトを作成
	parent := sm.CreateSpriteWithZPath(nil, window) // Z_Path: [0, 0]

	// 複数の子スプライトを作成
	child1 := sm.CreateSpriteWithZPath(nil, parent) // Z_Path: [0, 0, 0], Local_Z_Order: 0
	child2 := sm.CreateSpriteWithZPath(nil, parent) // Z_Path: [0, 0, 1], Local_Z_Order: 1
	child3 := sm.CreateSpriteWithZPath(nil, parent) // Z_Path: [0, 0, 2], Local_Z_Order: 2

	// 初期のLocal_Z_Orderを記録
	initialLocalZOrder1 := child1.GetZPath().LocalZOrder()
	initialLocalZOrder2 := child2.GetZPath().LocalZOrder()
	initialLocalZOrder3 := child3.GetZPath().LocalZOrder()

	// 親のZ_Pathを変更
	parent.SetZPath(NewZPath(0, 99))
	sm.UpdateChildrenZPathsForTest(parent)

	// Local_Z_Orderが保持されていることを確認
	if child1.GetZPath().LocalZOrder() != initialLocalZOrder1 {
		t.Errorf("child1のLocal_Z_Orderは%dのままのはず、got %d", initialLocalZOrder1, child1.GetZPath().LocalZOrder())
	}
	if child2.GetZPath().LocalZOrder() != initialLocalZOrder2 {
		t.Errorf("child2のLocal_Z_Orderは%dのままのはず、got %d", initialLocalZOrder2, child2.GetZPath().LocalZOrder())
	}
	if child3.GetZPath().LocalZOrder() != initialLocalZOrder3 {
		t.Errorf("child3のLocal_Z_Orderは%dのままのはず、got %d", initialLocalZOrder3, child3.GetZPath().LocalZOrder())
	}

	// Z_Pathの親部分が更新されていることを確認
	expectedPath := []int{0, 99, 0}
	if !equalIntSlice(child1.GetZPath().Path(), expectedPath) {
		t.Errorf("child1のZ_Pathは[0, 99, 0]のはず、got %v", child1.GetZPath().Path())
	}
	expectedPath = []int{0, 99, 1}
	if !equalIntSlice(child2.GetZPath().Path(), expectedPath) {
		t.Errorf("child2のZ_Pathは[0, 99, 1]のはず、got %v", child2.GetZPath().Path())
	}
	expectedPath = []int{0, 99, 2}
	if !equalIntSlice(child3.GetZPath().Path(), expectedPath) {
		t.Errorf("child3のZ_Pathは[0, 99, 2]のはず、got %v", child3.GetZPath().Path())
	}
}

// TestUpdateChildrenZPaths_NoChildren は子スプライトがない場合のテスト
func TestUpdateChildrenZPaths_NoChildren(t *testing.T) {
	sm := NewSpriteManager()

	// ウインドウを作成
	window := sm.CreateRootSprite(nil, 0) // Z_Path: [0]

	// 子スプライトを持たないスプライトを作成
	sprite := sm.CreateSpriteWithZPath(nil, window) // Z_Path: [0, 0]

	// 子スプライトがないことを確認
	if sprite.HasChildren() {
		t.Error("spriteは子スプライトを持たないはず")
	}

	// Z_Pathを変更
	sprite.SetZPath(NewZPath(0, 5))

	// updateChildrenZPathsを呼び出す（パニックしないことを確認）
	sm.UpdateChildrenZPathsForTest(sprite)

	// Z_Pathが変更されていることを確認
	expectedPath := []int{0, 5}
	if !equalIntSlice(sprite.GetZPath().Path(), expectedPath) {
		t.Errorf("spriteのZ_Pathは[0, 5]のはず、got %v", sprite.GetZPath().Path())
	}
}

// TestUpdateChildrenZPaths_MultipleBranches は複数の分岐を持つ階層のテスト
func TestUpdateChildrenZPaths_MultipleBranches(t *testing.T) {
	sm := NewSpriteManager()

	// ウインドウを作成
	window := sm.CreateRootSprite(nil, 0) // Z_Path: [0]

	// 親スプライトを作成
	parent := sm.CreateSpriteWithZPath(nil, window) // Z_Path: [0, 0]

	// 複数の子スプライトを作成（それぞれが孫を持つ）
	child1 := sm.CreateSpriteWithZPath(nil, parent)        // Z_Path: [0, 0, 0]
	grandchild1_1 := sm.CreateSpriteWithZPath(nil, child1) // Z_Path: [0, 0, 0, 0]
	grandchild1_2 := sm.CreateSpriteWithZPath(nil, child1) // Z_Path: [0, 0, 0, 1]

	child2 := sm.CreateSpriteWithZPath(nil, parent)        // Z_Path: [0, 0, 1]
	grandchild2_1 := sm.CreateSpriteWithZPath(nil, child2) // Z_Path: [0, 0, 1, 0]

	// 親のZ_Pathを変更
	parent.SetZPath(NewZPath(0, 7))
	sm.UpdateChildrenZPathsForTest(parent)

	// すべての子孫スプライトのZ_Pathが更新されていることを確認
	expectedPath := []int{0, 7, 0}
	if !equalIntSlice(child1.GetZPath().Path(), expectedPath) {
		t.Errorf("child1のZ_Pathは[0, 7, 0]のはず、got %v", child1.GetZPath().Path())
	}
	expectedPath = []int{0, 7, 0, 0}
	if !equalIntSlice(grandchild1_1.GetZPath().Path(), expectedPath) {
		t.Errorf("grandchild1_1のZ_Pathは[0, 7, 0, 0]のはず、got %v", grandchild1_1.GetZPath().Path())
	}
	expectedPath = []int{0, 7, 0, 1}
	if !equalIntSlice(grandchild1_2.GetZPath().Path(), expectedPath) {
		t.Errorf("grandchild1_2のZ_Pathは[0, 7, 0, 1]のはず、got %v", grandchild1_2.GetZPath().Path())
	}
	expectedPath = []int{0, 7, 1}
	if !equalIntSlice(child2.GetZPath().Path(), expectedPath) {
		t.Errorf("child2のZ_Pathは[0, 7, 1]のはず、got %v", child2.GetZPath().Path())
	}
	expectedPath = []int{0, 7, 1, 0}
	if !equalIntSlice(grandchild2_1.GetZPath().Path(), expectedPath) {
		t.Errorf("grandchild2_1のZ_Pathは[0, 7, 1, 0]のはず、got %v", grandchild2_1.GetZPath().Path())
	}
}
