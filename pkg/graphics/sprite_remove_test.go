package graphics

import (
	"testing"
)

// TestRemoveSprite_RemovesFromParentChildren tests that RemoveSprite removes the sprite from its parent's children list
// 要件 9.2: ウィンドウが閉じられたとき、関連するスプライト（キャスト、テキスト等）を削除する
func TestRemoveSprite_RemovesFromParentChildren(t *testing.T) {
	sm := NewSpriteManager()

	// 親スプライトを作成
	parent := sm.CreateRootSprite(nil, 0)

	// 子スプライトを作成
	child1 := sm.CreateSpriteWithZPath(nil, parent)
	child2 := sm.CreateSpriteWithZPath(nil, parent)

	// 初期状態: 親は2つの子を持つ
	if len(parent.GetChildren()) != 2 {
		t.Fatalf("親は2つの子を持つはず、got %d", len(parent.GetChildren()))
	}

	// child1を削除
	sm.RemoveSprite(child1.ID())

	// 親の子リストからchild1が削除されていることを確認
	if len(parent.GetChildren()) != 1 {
		t.Errorf("RemoveSprite後、親は1つの子を持つはず、got %d", len(parent.GetChildren()))
	}

	// 残っている子はchild2のはず
	if len(parent.GetChildren()) > 0 && parent.GetChildren()[0].ID() != child2.ID() {
		t.Errorf("残っている子はchild2のはず、got ID %d", parent.GetChildren()[0].ID())
	}

	// child1はsprites mapから削除されている
	if sm.GetSprite(child1.ID()) != nil {
		t.Error("child1はsprites mapから削除されているはず")
	}

	// child2はまだ存在する
	if sm.GetSprite(child2.ID()) == nil {
		t.Error("child2はまだsprites mapに存在するはず")
	}
}

// TestRemoveSprite_RecursivelyDeletesChildren tests that RemoveSprite recursively deletes all child sprites
// 要件 9.2: ウィンドウが閉じられたとき、関連するスプライト（キャスト、テキスト等）を削除する
func TestRemoveSprite_RecursivelyDeletesChildren(t *testing.T) {
	sm := NewSpriteManager()

	// 親スプライトを作成
	parent := sm.CreateRootSprite(nil, 0)

	// 子スプライトを作成
	child := sm.CreateSpriteWithZPath(nil, parent)

	// 孫スプライトを作成
	grandchild1 := sm.CreateSpriteWithZPath(nil, child)
	grandchild2 := sm.CreateSpriteWithZPath(nil, child)

	// 曾孫スプライトを作成
	greatGrandchild := sm.CreateSpriteWithZPath(nil, grandchild1)

	// 初期状態: 5つのスプライトが存在
	if sm.Count() != 5 {
		t.Fatalf("初期状態で5つのスプライトが存在するはず、got %d", sm.Count())
	}

	// 子スプライトを削除（孫、曾孫も再帰的に削除されるはず）
	sm.RemoveSprite(child.ID())

	// 子、孫、曾孫がすべて削除されている
	if sm.GetSprite(child.ID()) != nil {
		t.Error("childはsprites mapから削除されているはず")
	}
	if sm.GetSprite(grandchild1.ID()) != nil {
		t.Error("grandchild1はsprites mapから削除されているはず")
	}
	if sm.GetSprite(grandchild2.ID()) != nil {
		t.Error("grandchild2はsprites mapから削除されているはず")
	}
	if sm.GetSprite(greatGrandchild.ID()) != nil {
		t.Error("greatGrandchildはsprites mapから削除されているはず")
	}

	// 親だけが残っている
	if sm.Count() != 1 {
		t.Errorf("親だけが残っているはず、got %d sprites", sm.Count())
	}
	if sm.GetSprite(parent.ID()) == nil {
		t.Error("親はまだsprites mapに存在するはず")
	}

	// 親の子リストは空
	if len(parent.GetChildren()) != 0 {
		t.Errorf("親の子リストは空のはず、got %d children", len(parent.GetChildren()))
	}
}

// TestRemoveSprite_WithNoParent tests that RemoveSprite works correctly for sprites without a parent
func TestRemoveSprite_WithNoParent(t *testing.T) {
	sm := NewSpriteManager()

	// 親なしのスプライトを作成
	sprite := sm.CreateRootSprite(nil, 0)

	// 初期状態
	if sm.Count() != 1 {
		t.Fatalf("初期状態で1つのスプライトが存在するはず、got %d", sm.Count())
	}

	// スプライトを削除
	sm.RemoveSprite(sprite.ID())

	// スプライトが削除されている
	if sm.GetSprite(sprite.ID()) != nil {
		t.Error("spriteはsprites mapから削除されているはず")
	}
	if sm.Count() != 0 {
		t.Errorf("すべてのスプライトが削除されているはず、got %d sprites", sm.Count())
	}
}

// TestRemoveSprite_WithMultipleSiblings tests that RemoveSprite only removes the specified sprite, not its siblings
func TestRemoveSprite_WithMultipleSiblings(t *testing.T) {
	sm := NewSpriteManager()

	// 親スプライトを作成
	parent := sm.CreateRootSprite(nil, 0)

	// 3つの子スプライトを作成
	child1 := sm.CreateSpriteWithZPath(nil, parent)
	child2 := sm.CreateSpriteWithZPath(nil, parent)
	child3 := sm.CreateSpriteWithZPath(nil, parent)

	// 初期状態: 4つのスプライトが存在
	if sm.Count() != 4 {
		t.Fatalf("初期状態で4つのスプライトが存在するはず、got %d", sm.Count())
	}

	// child2を削除
	sm.RemoveSprite(child2.ID())

	// child2だけが削除されている
	if sm.GetSprite(child2.ID()) != nil {
		t.Error("child2はsprites mapから削除されているはず")
	}

	// child1とchild3は残っている
	if sm.GetSprite(child1.ID()) == nil {
		t.Error("child1はまだsprites mapに存在するはず")
	}
	if sm.GetSprite(child3.ID()) == nil {
		t.Error("child3はまだsprites mapに存在するはず")
	}

	// 親も残っている
	if sm.GetSprite(parent.ID()) == nil {
		t.Error("親はまだsprites mapに存在するはず")
	}

	// 親の子リストには2つの子が残っている
	if len(parent.GetChildren()) != 2 {
		t.Errorf("親の子リストには2つの子が残っているはず、got %d", len(parent.GetChildren()))
	}

	// 残っている子はchild1とchild3
	foundChild1 := false
	foundChild3 := false
	for _, child := range parent.GetChildren() {
		if child.ID() == child1.ID() {
			foundChild1 = true
		}
		if child.ID() == child3.ID() {
			foundChild3 = true
		}
	}
	if !foundChild1 {
		t.Error("child1が親の子リストに残っているはず")
	}
	if !foundChild3 {
		t.Error("child3が親の子リストに残っているはず")
	}
}
