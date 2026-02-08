package sprite

import (
	"math"
	"testing"
	"testing/quick"
)

// Property 1: スプライトID管理
// 任意のスプライト作成に対して、作成されたスプライトは一意のIDを持ち、そのIDで取得できる
// **Validates: Requirements 3.1**
func TestProperty_SpriteIDManagement(t *testing.T) {
	f := func(count uint8) bool {
		if count == 0 {
			return true
		}
		// 最大100個に制限
		n := int(count)
		if n > 100 {
			n = 100
		}

		sm := NewSpriteManager()
		ids := make(map[int]bool)
		sprites := make([]*Sprite, 0, n)

		// n個のスプライトを作成
		for i := 0; i < n; i++ {
			s := sm.CreateSpriteWithSize(10, 10, nil)
			if s == nil {
				return false
			}
			// IDが一意であることを確認
			if ids[s.ID()] {
				return false // 重複ID
			}
			ids[s.ID()] = true
			sprites = append(sprites, s)
		}

		// すべてのスプライトがIDで取得できることを確認
		for _, s := range sprites {
			got := sm.GetSprite(s.ID())
			if got != s {
				return false
			}
		}

		return true
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

// Property 2: 親子関係の位置計算
// 任意の親子関係を持つスプライトに対して、子の絶対位置は親の位置と子の相対位置の和である
// **Validates: Requirements 2.1**
func TestProperty_ParentChildPosition(t *testing.T) {
	f := func(parentX, parentY, childX, childY int16) bool {
		// int16を使用して値の範囲を制限（-32768〜32767）
		px := float64(parentX)
		py := float64(parentY)
		cx := float64(childX)
		cy := float64(childY)

		parent := NewSprite(1, nil)
		parent.SetPosition(px, py)

		child := NewSprite(2, nil)
		child.SetPosition(cx, cy)
		parent.AddChild(child)

		absX, absY := child.AbsolutePosition()
		expectedX := px + cx
		expectedY := py + cy

		// 浮動小数点の誤差を考慮
		const epsilon = 1e-9
		return math.Abs(absX-expectedX) < epsilon && math.Abs(absY-expectedY) < epsilon
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

// Property 3: 親子関係の透明度計算
// 任意の親子関係を持つスプライトに対して、子の実効透明度は親の透明度と子の透明度の積である
// **Validates: Requirements 2.2**
func TestProperty_ParentChildAlpha(t *testing.T) {
	f := func(parentAlpha, childAlpha float64) bool {
		// 0.0〜1.0の範囲に正規化
		parentAlpha = math.Abs(math.Mod(parentAlpha, 1.0))
		childAlpha = math.Abs(math.Mod(childAlpha, 1.0))

		parent := NewSprite(1, nil)
		parent.SetAlpha(parentAlpha)

		child := NewSprite(2, nil)
		child.SetAlpha(childAlpha)
		parent.AddChild(child)

		effectiveAlpha := child.EffectiveAlpha()
		expectedAlpha := parentAlpha * childAlpha

		// 浮動小数点の誤差を考慮
		const epsilon = 1e-9
		return math.Abs(effectiveAlpha-expectedAlpha) < epsilon
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

// Property 4: 親子関係の可視性
// 任意の親子関係を持つスプライトに対して、親が非表示なら子も非表示として扱われる
// **Validates: Requirements 2.3, 10.4**
func TestProperty_ParentChildVisibility(t *testing.T) {
	f := func(parentVisible, childVisible bool) bool {
		parent := NewSprite(1, nil)
		parent.SetVisible(parentVisible)

		child := NewSprite(2, nil)
		child.SetVisible(childVisible)
		parent.AddChild(child)

		effectivelyVisible := child.IsEffectivelyVisible()

		// 親が非表示なら子も非表示
		if !parentVisible {
			return !effectivelyVisible
		}
		// 親が表示なら子の可視性に依存
		return effectivelyVisible == childVisible
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

// Property 5: 追加順序の保持
// 任意の子スプライトについて、後から追加されたスプライトはスライスの後方に位置する
// **Validates: Requirements 9.3**
func TestProperty_AdditionOrderPreserved(t *testing.T) {
	f := func(count uint8) bool {
		if count < 2 {
			return true
		}
		n := int(count)
		if n > 50 {
			n = 50
		}

		parent := NewSprite(1, nil)
		children := make([]*Sprite, 0, n)

		// n個の子スプライトを追加
		for i := 0; i < n; i++ {
			child := NewSprite(i+2, nil)
			parent.AddChild(child)
			children = append(children, child)
		}

		// 追加順序が保持されていることを確認
		gotChildren := parent.GetChildren()
		if len(gotChildren) != n {
			return false
		}
		for i, child := range gotChildren {
			if child != children[i] {
				return false
			}
		}

		return true
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

// Property 6: BringToFront後のスライス末尾
// 任意のスプライトについて、BringToFront後にそのスプライトはスライスの末尾に位置する
// **Validates: Requirements 12.1**
func TestProperty_BringToFrontMovesToEnd(t *testing.T) {
	f := func(count uint8, targetIdx uint8) bool {
		if count < 2 {
			return true
		}
		n := int(count)
		if n > 50 {
			n = 50
		}
		idx := int(targetIdx) % n

		parent := NewSprite(1, nil)
		children := make([]*Sprite, 0, n)

		// n個の子スプライトを追加
		for i := 0; i < n; i++ {
			child := NewSprite(i+2, nil)
			parent.AddChild(child)
			children = append(children, child)
		}

		// 指定されたインデックスの子を最前面に移動
		target := children[idx]
		target.BringToFront()

		// 最前面に移動したスプライトがスライスの末尾にあることを確認
		gotChildren := parent.GetChildren()
		if len(gotChildren) != n {
			return false
		}
		if gotChildren[n-1] != target {
			return false
		}

		return true
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

// Property 7: SendToBack後のスライス先頭
// 任意のスプライトについて、SendToBack後にそのスプライトはスライスの先頭に位置する
// **Validates: Requirements 12.2**
func TestProperty_SendToBackMovesToFront(t *testing.T) {
	f := func(count uint8, targetIdx uint8) bool {
		if count < 2 {
			return true
		}
		n := int(count)
		if n > 50 {
			n = 50
		}
		idx := int(targetIdx) % n

		parent := NewSprite(1, nil)
		children := make([]*Sprite, 0, n)

		// n個の子スプライトを追加
		for i := 0; i < n; i++ {
			child := NewSprite(i+2, nil)
			parent.AddChild(child)
			children = append(children, child)
		}

		// 指定されたインデックスの子を最背面に移動
		target := children[idx]
		target.SendToBack()

		// 最背面に移動したスプライトがスライスの先頭にあることを確認
		gotChildren := parent.GetChildren()
		if len(gotChildren) != n {
			return false
		}
		if gotChildren[0] != target {
			return false
		}

		return true
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

// Property 8: スプライト削除
// 任意のスプライト削除に対して、削除後はそのIDでスプライトを取得できない
// **Validates: Requirements 3.4**
func TestProperty_SpriteRemoval(t *testing.T) {
	f := func(count uint8) bool {
		if count == 0 {
			return true
		}
		n := int(count)
		if n > 50 {
			n = 50
		}

		sm := NewSpriteManager()
		ids := make([]int, 0, n)

		// スプライトを作成
		for i := 0; i < n; i++ {
			s := sm.CreateSpriteWithSize(10, 10, nil)
			ids = append(ids, s.ID())
		}

		// 半分を削除
		for i := 0; i < n/2; i++ {
			sm.DeleteSprite(ids[i])
		}

		// 削除されたスプライトは取得できない
		for i := 0; i < n/2; i++ {
			if sm.GetSprite(ids[i]) != nil {
				return false
			}
		}

		// 削除されていないスプライトは取得できる
		for i := n / 2; i < n; i++ {
			if sm.GetSprite(ids[i]) == nil {
				return false
			}
		}

		return true
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

// Property 9: 子スプライトの削除時に親からも削除される
// 任意の親子関係を持つスプライトに対して、子を削除すると親のchildrenからも削除される
// **Validates: Requirements 3.4**
func TestProperty_ChildRemovalFromParent(t *testing.T) {
	f := func(count uint8, deleteIdx uint8) bool {
		if count < 2 {
			return true
		}
		n := int(count)
		if n > 50 {
			n = 50
		}
		idx := int(deleteIdx) % n

		sm := NewSpriteManager()
		parent := sm.CreateRootSprite(nil)
		children := make([]*Sprite, 0, n)

		// n個の子スプライトを追加
		for i := 0; i < n; i++ {
			child := sm.CreateSprite(nil, parent)
			children = append(children, child)
		}

		// 指定されたインデックスの子を削除
		target := children[idx]
		targetID := target.ID()
		sm.DeleteSprite(targetID)

		// 親のchildrenから削除されていることを確認
		for _, child := range parent.GetChildren() {
			if child.ID() == targetID {
				return false
			}
		}

		// SpriteManagerからも削除されていることを確認
		if sm.GetSprite(targetID) != nil {
			return false
		}

		return true
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}
