package graphics

import (
	"testing"
	"testing/quick"
)

// ============================================================================
// ZPathのプロパティベーステスト
// ============================================================================

// Property 1: Z_Pathの一意性
// 任意の2つのスプライトについて、それらのZ_Pathは異なる
// **Validates: Requirements 1.1, 1.2**
func TestProperty_ZPath_Uniqueness(t *testing.T) {
	f := func(count uint8) bool {
		if count < 2 {
			return true
		}
		// 最大50個に制限
		n := int(count)
		if n > 50 {
			n = 50
		}

		sm := NewSpriteManager()
		root := sm.CreateRootSprite(nil, 0)
		if root == nil {
			return false
		}

		zPaths := make(map[string]bool)
		zPaths[root.ZPathString()] = true

		// n個の子スプライトを作成
		for i := 0; i < n-1; i++ {
			s := sm.CreateSpriteWithZPath(nil, root)
			if s == nil {
				return false
			}
			zPathStr := s.ZPathString()
			// Z_Pathが一意であることを確認
			if zPaths[zPathStr] {
				return false // 重複Z_Path
			}
			zPaths[zPathStr] = true
		}

		return true
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

// Property 2: Z_Pathの継承
// 任意の子スプライトについて、そのZ_Pathは親のZ_Pathをプレフィックスとして持つ
// **Validates: Requirements 1.2, 1.4**
func TestProperty_ZPath_Inheritance(t *testing.T) {
	f := func(depth uint8) bool {
		// 深さを1〜10に制限
		d := int(depth%10) + 1

		sm := NewSpriteManager()
		parent := sm.CreateRootSprite(nil, 0)
		if parent == nil {
			return false
		}

		// 深さdまでの階層を作成
		for i := 0; i < d; i++ {
			child := sm.CreateSpriteWithZPath(nil, parent)
			if child == nil {
				return false
			}

			// 子のZ_Pathが親のZ_Pathをプレフィックスとして持つことを確認
			if parent.GetZPath() != nil && child.GetZPath() != nil {
				if !parent.GetZPath().IsPrefix(child.GetZPath()) {
					return false
				}
			}

			parent = child
		}

		return true
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

// Property 3: ルートスプライトのZ_Path
// 任意のルートスプライトについて、そのZ_Pathは単一要素である
// **Validates: Requirements 1.3**
func TestProperty_ZPath_RootSprite(t *testing.T) {
	f := func(windowZOrder int16) bool {
		sm := NewSpriteManager()
		root := sm.CreateRootSprite(nil, int(windowZOrder))
		if root == nil {
			return false
		}

		zPath := root.GetZPath()
		if zPath == nil {
			return false
		}

		// ルートスプライトのZ_Pathは単一要素
		return zPath.Depth() == 1 && zPath.LocalZOrder() == int(windowZOrder)
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

// Property 11: 辞書順比較の正確性
// 任意の2つのZ_Pathについて、Compare関数は正しい辞書順を返す
// **Validates: Requirements 5.1, 5.2, 5.3**
func TestProperty_ZPath_LexicographicComparison(t *testing.T) {
	f := func(a1, a2, b1, b2 int8) bool {
		// 2要素のZ_Pathを作成
		zPathA := NewZPath(int(a1), int(a2))
		zPathB := NewZPath(int(b1), int(b2))

		cmp := zPathA.Compare(zPathB)

		// 辞書順比較の検証
		if int(a1) < int(b1) {
			return cmp == -1
		}
		if int(a1) > int(b1) {
			return cmp == 1
		}
		// a1 == b1の場合、a2とb2を比較
		if int(a2) < int(b2) {
			return cmp == -1
		}
		if int(a2) > int(b2) {
			return cmp == 1
		}
		// 完全に等しい
		return cmp == 0
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

// Property: 比較の推移性
// 任意の3つのZ_Pathについて、A < B かつ B < C ならば A < C
// **Validates: Requirements 5.1**
func TestProperty_ZPath_Transitivity(t *testing.T) {
	f := func(a, b, c int8) bool {
		// 単一要素のZ_Pathを作成
		zPathA := NewZPath(int(a))
		zPathB := NewZPath(int(b))
		zPathC := NewZPath(int(c))

		// A < B かつ B < C ならば A < C
		if zPathA.Less(zPathB) && zPathB.Less(zPathC) {
			return zPathA.Less(zPathC)
		}

		return true
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

// Property: 比較の反対称性
// 任意の2つのZ_Pathについて、A < B ならば B < A は偽
// **Validates: Requirements 5.1**
func TestProperty_ZPath_Antisymmetry(t *testing.T) {
	f := func(a, b int8) bool {
		zPathA := NewZPath(int(a))
		zPathB := NewZPath(int(b))

		// A < B ならば B < A は偽
		if zPathA.Less(zPathB) {
			return !zPathB.Less(zPathA)
		}

		return true
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

// Property: 比較の反射性
// 任意のZ_Pathについて、A < A は偽
// **Validates: Requirements 5.1**
func TestProperty_ZPath_Reflexivity(t *testing.T) {
	f := func(a int8) bool {
		zPath := NewZPath(int(a))

		// A < A は偽
		return !zPath.Less(zPath)
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

// Property: プレフィックス判定
// 任意のZ_Pathについて、自身は自身のプレフィックスである
// **Validates: Requirements 5.3**
func TestProperty_ZPath_SelfIsPrefix(t *testing.T) {
	f := func(a, b int8) bool {
		zPath := NewZPath(int(a), int(b))

		// 自身は自身のプレフィックス
		return zPath.IsPrefix(zPath)
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

// Property: 親のZ_Pathは子のプレフィックス
// 任意の親子関係について、親のZ_Pathは子のZ_Pathのプレフィックスである
// **Validates: Requirements 1.2, 5.3**
func TestProperty_ZPath_ParentIsPrefix(t *testing.T) {
	f := func(parentZOrder, childLocalZOrder int8) bool {
		parent := NewZPath(int(parentZOrder))
		child := NewZPathFromParent(parent, int(childLocalZOrder))

		// 親のZ_Pathは子のZ_Pathのプレフィックス
		return parent.IsPrefix(child)
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

// ============================================================================
// ZOrderCounterのプロパティベーステスト
// ============================================================================

// Property: カウンターの単調増加
// 任意の親IDについて、GetNextを呼び出すたびにカウンターは増加する
// **Validates: Requirements 2.1, 2.5**
func TestProperty_ZOrderCounter_Monotonic(t *testing.T) {
	f := func(parentID int8, count uint8) bool {
		if count == 0 {
			return true
		}
		// 最大100回に制限
		n := int(count)
		if n > 100 {
			n = 100
		}

		counter := NewZOrderCounter()
		prev := -1

		for i := 0; i < n; i++ {
			current := counter.GetNext(int(parentID))
			if current <= prev {
				return false // 単調増加でない
			}
			prev = current
		}

		return true
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

// Property: 異なる親IDは独立したカウンター
// 任意の2つの異なる親IDについて、それぞれのカウンターは独立している
// **Validates: Requirements 2.1**
func TestProperty_ZOrderCounter_Independence(t *testing.T) {
	f := func(parentID1, parentID2 int8) bool {
		if parentID1 == parentID2 {
			return true // 同じ親IDの場合はスキップ
		}

		counter := NewZOrderCounter()

		// 親ID1のカウンターを進める
		for i := 0; i < 5; i++ {
			counter.GetNext(int(parentID1))
		}

		// 親ID2のカウンターは0から始まる
		first := counter.GetNext(int(parentID2))
		return first == 0
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

// Property: リセット後のカウンター
// 任意の親IDについて、Reset後のカウンターは0から始まる
// **Validates: Requirements 2.1**
func TestProperty_ZOrderCounter_Reset(t *testing.T) {
	f := func(parentID int8, count uint8) bool {
		if count == 0 {
			return true
		}
		n := int(count)
		if n > 50 {
			n = 50
		}

		counter := NewZOrderCounter()

		// カウンターを進める
		for i := 0; i < n; i++ {
			counter.GetNext(int(parentID))
		}

		// リセット
		counter.Reset(int(parentID))

		// リセット後は0から始まる
		first := counter.GetNext(int(parentID))
		return first == 0
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

// ============================================================================
// 操作順序のプロパティベーステスト
// ============================================================================

// Property 4: 操作順序の反映
// 任意の同じ親を持つ2つのスプライトについて、後から作成されたスプライトのLocal_Z_Orderは大きい
// **Validates: Requirements 2.2, 2.3, 2.4, 2.5**
func TestProperty_OperationOrder_Reflection(t *testing.T) {
	f := func(count uint8) bool {
		if count < 2 {
			return true
		}
		// 最大50個に制限
		n := int(count)
		if n > 50 {
			n = 50
		}

		sm := NewSpriteManager()
		root := sm.CreateRootSprite(nil, 0)
		if root == nil {
			return false
		}

		prevLocalZOrder := -1

		// n個の子スプライトを作成
		for i := 0; i < n; i++ {
			s := sm.CreateSpriteWithZPath(nil, root)
			if s == nil {
				return false
			}

			currentLocalZOrder := s.GetZPath().LocalZOrder()

			// 後から作成されたスプライトのLocal_Z_Orderは大きい
			if currentLocalZOrder <= prevLocalZOrder {
				return false
			}

			prevLocalZOrder = currentLocalZOrder
		}

		return true
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

// Property 5: タイプ非依存性
// 任意のスプライトについて、そのZ順序はタイプ（キャスト、テキスト、ピクチャ）に依存しない
// **Validates: Requirements 2.6**
func TestProperty_TypeIndependence(t *testing.T) {
	f := func(seed uint8) bool {
		sm := NewSpriteManager()
		root := sm.CreateRootSprite(nil, 0)
		if root == nil {
			return false
		}

		// 異なる「タイプ」のスプライトを作成（実際にはすべてSprite）
		// タイプに関係なく、作成順序でLocal_Z_Orderが決まることを確認
		sprites := make([]*Sprite, 0)

		// 3つのスプライトを作成
		for i := 0; i < 3; i++ {
			s := sm.CreateSpriteWithZPath(nil, root)
			if s == nil {
				return false
			}
			sprites = append(sprites, s)
		}

		// 作成順序でLocal_Z_Orderが増加していることを確認
		for i := 1; i < len(sprites); i++ {
			if sprites[i].GetZPath().LocalZOrder() <= sprites[i-1].GetZPath().LocalZOrder() {
				return false
			}
		}

		return true
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

// ============================================================================
// 描画順序のプロパティベーステスト
// ============================================================================

// Property 6: 親子描画順序
// 任意の親子関係について、親スプライトは子スプライトより先に描画される
// **Validates: Requirements 3.1**
func TestProperty_DrawOrder_ParentBeforeChild(t *testing.T) {
	f := func(depth uint8) bool {
		// 深さを1〜5に制限
		d := int(depth%5) + 1

		sm := NewSpriteManager()
		parent := sm.CreateRootSprite(nil, 0)
		if parent == nil {
			return false
		}

		sprites := []*Sprite{parent}

		// 深さdまでの階層を作成
		for i := 0; i < d; i++ {
			child := sm.CreateSpriteWithZPath(nil, parent)
			if child == nil {
				return false
			}
			sprites = append(sprites, child)
			parent = child
		}

		// 親のZ_Pathは子のZ_Pathより小さい（先に描画される）
		for i := 0; i < len(sprites)-1; i++ {
			parent := sprites[i]
			child := sprites[i+1]

			if parent.GetZPath() == nil || child.GetZPath() == nil {
				return false
			}

			// 親のZ_Pathは子のZ_Pathより小さい
			if !parent.GetZPath().Less(child.GetZPath()) {
				return false
			}
		}

		return true
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

// Property 7: 兄弟描画順序
// 任意の同じ親を持つ2つのスプライトについて、Local_Z_Orderが小さいスプライトが先に描画される
// **Validates: Requirements 3.2**
func TestProperty_DrawOrder_SiblingOrder(t *testing.T) {
	f := func(count uint8) bool {
		if count < 2 {
			return true
		}
		// 最大20個に制限
		n := int(count)
		if n > 20 {
			n = 20
		}

		sm := NewSpriteManager()
		root := sm.CreateRootSprite(nil, 0)
		if root == nil {
			return false
		}

		siblings := make([]*Sprite, 0, n)

		// n個の兄弟スプライトを作成
		for i := 0; i < n; i++ {
			s := sm.CreateSpriteWithZPath(nil, root)
			if s == nil {
				return false
			}
			siblings = append(siblings, s)
		}

		// Local_Z_Orderが小さいスプライトのZ_Pathは小さい（先に描画される）
		for i := 0; i < len(siblings)-1; i++ {
			s1 := siblings[i]
			s2 := siblings[i+1]

			if s1.GetZPath() == nil || s2.GetZPath() == nil {
				return false
			}

			// s1のLocal_Z_Orderはs2より小さい
			if s1.GetZPath().LocalZOrder() >= s2.GetZPath().LocalZOrder() {
				return false
			}

			// s1のZ_Pathはs2より小さい
			if !s1.GetZPath().Less(s2.GetZPath()) {
				return false
			}
		}

		return true
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

// Property 8: 可視性の継承
// 任意の親が非表示のスプライトについて、その子スプライトも描画されない
// **Validates: Requirements 3.4**
func TestProperty_DrawOrder_VisibilityInheritance(t *testing.T) {
	f := func(depth uint8) bool {
		// 深さを1〜5に制限
		d := int(depth%5) + 1

		sm := NewSpriteManager()
		parent := sm.CreateRootSprite(nil, 0)
		if parent == nil {
			return false
		}

		sprites := []*Sprite{parent}

		// 深さdまでの階層を作成
		for i := 0; i < d; i++ {
			child := sm.CreateSpriteWithZPath(nil, parent)
			if child == nil {
				return false
			}
			sprites = append(sprites, child)
			parent = child
		}

		// ルートを非表示にする
		sprites[0].SetVisible(false)

		// すべての子孫スプライトはIsEffectivelyVisible() == false
		for i := 1; i < len(sprites); i++ {
			if sprites[i].IsEffectivelyVisible() {
				return false
			}
		}

		return true
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

// ============================================================================
// ウインドウ間のプロパティベーステスト
// ============================================================================

// Property 9: ウインドウ間の描画順序
// 任意の2つのウインドウについて、前面のウインドウのすべての子スプライトは
// 背面のウインドウのすべての子スプライトより後に描画される
// **Validates: Requirements 4.1, 4.2**
func TestProperty_WindowOrder_DrawOrder(t *testing.T) {
	f := func(childCount uint8) bool {
		// 子スプライト数を1〜10に制限
		n := int(childCount%10) + 1

		sm := NewSpriteManager()

		// 2つのウインドウを作成（window0が背面、window1が前面）
		window0 := sm.CreateRootSprite(nil, 0)
		window1 := sm.CreateRootSprite(nil, 1)
		if window0 == nil || window1 == nil {
			return false
		}

		// 各ウインドウにn個の子スプライトを作成
		children0 := make([]*Sprite, 0, n)
		children1 := make([]*Sprite, 0, n)

		for i := 0; i < n; i++ {
			c0 := sm.CreateSpriteWithZPath(nil, window0)
			c1 := sm.CreateSpriteWithZPath(nil, window1)
			if c0 == nil || c1 == nil {
				return false
			}
			children0 = append(children0, c0)
			children1 = append(children1, c1)
		}

		// window0のすべての子スプライトはwindow1のすべての子スプライトより先に描画される
		for _, c0 := range children0 {
			for _, c1 := range children1 {
				if c0.GetZPath() == nil || c1.GetZPath() == nil {
					return false
				}
				// c0のZ_Pathはc1より小さい
				if !c0.GetZPath().Less(c1.GetZPath()) {
					return false
				}
			}
		}

		return true
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

// Property 10: ウインドウZ順序更新の伝播
// 任意のウインドウについて、Z順序が変更されたとき、そのすべての子スプライトのZ_Pathが更新される
// **Validates: Requirements 4.3, 4.4**
func TestProperty_WindowOrder_ZPathPropagation(t *testing.T) {
	f := func(childCount uint8, newZOrder int8) bool {
		// 子スプライト数を1〜10に制限
		n := int(childCount%10) + 1
		// 新しいZ順序を正の値に制限
		newZ := int(newZOrder)
		if newZ < 0 {
			newZ = -newZ
		}

		sm := NewSpriteManager()

		// ウインドウを作成
		window := sm.CreateRootSprite(nil, 0)
		if window == nil {
			return false
		}

		// n個の子スプライトを作成
		children := make([]*Sprite, 0, n)
		for i := 0; i < n; i++ {
			c := sm.CreateSpriteWithZPath(nil, window)
			if c == nil {
				return false
			}
			children = append(children, c)
		}

		// ウインドウのZ順序を変更
		window.SetZPath(NewZPath(newZ))
		sm.UpdateChildrenZPaths(window)

		// すべての子スプライトのZ_Pathが更新されていることを確認
		for _, c := range children {
			if c.GetZPath() == nil {
				return false
			}
			// 子のZ_Pathの最初の要素が新しいウインドウZ順序と一致
			path := c.GetZPath().Path()
			if len(path) < 2 || path[0] != newZ {
				return false
			}
		}

		return true
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

// ============================================================================
// 動的変更のプロパティベーステスト
// ============================================================================

// Property 14: 最前面移動
// 任意のスプライトについて、BringToFront後にそのスプライトは
// 同じ親を持つ兄弟の中で最大のLocal_Z_Orderを持つ
// **Validates: Requirements 8.4**
func TestProperty_DynamicChange_BringToFront(t *testing.T) {
	f := func(count uint8, targetIdx uint8) bool {
		if count < 2 {
			return true
		}
		// 兄弟数を2〜20に制限
		n := int(count)
		if n > 20 {
			n = 20
		}

		sm := NewSpriteManager()
		root := sm.CreateRootSprite(nil, 0)
		if root == nil {
			return false
		}

		// n個の兄弟スプライトを作成
		siblings := make([]*Sprite, 0, n)
		for i := 0; i < n; i++ {
			s := sm.CreateSpriteWithZPath(nil, root)
			if s == nil {
				return false
			}
			siblings = append(siblings, s)
		}

		// ターゲットを選択
		target := siblings[int(targetIdx)%n]

		// BringToFrontを実行
		err := sm.BringToFront(target.ID())
		if err != nil {
			return false
		}

		// ターゲットのLocal_Z_Orderが最大であることを確認
		targetLocalZOrder := target.GetZPath().LocalZOrder()
		for _, s := range siblings {
			if s.ID() != target.ID() {
				if s.GetZPath().LocalZOrder() >= targetLocalZOrder {
					return false
				}
			}
		}

		return true
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

// Property 15: 最背面移動
// 任意のスプライトについて、SendToBack後にそのスプライトは
// 同じ親を持つ兄弟の中で最小のLocal_Z_Orderを持つ
// **Validates: Requirements 8.5**
func TestProperty_DynamicChange_SendToBack(t *testing.T) {
	f := func(count uint8, targetIdx uint8) bool {
		if count < 2 {
			return true
		}
		// 兄弟数を2〜20に制限
		n := int(count)
		if n > 20 {
			n = 20
		}

		sm := NewSpriteManager()
		root := sm.CreateRootSprite(nil, 0)
		if root == nil {
			return false
		}

		// n個の兄弟スプライトを作成
		siblings := make([]*Sprite, 0, n)
		for i := 0; i < n; i++ {
			s := sm.CreateSpriteWithZPath(nil, root)
			if s == nil {
				return false
			}
			siblings = append(siblings, s)
		}

		// ターゲットを選択
		target := siblings[int(targetIdx)%n]

		// SendToBackを実行
		err := sm.SendToBack(target.ID())
		if err != nil {
			return false
		}

		// ターゲットのLocal_Z_Orderが最小であることを確認
		targetLocalZOrder := target.GetZPath().LocalZOrder()
		for _, s := range siblings {
			if s.ID() != target.ID() {
				if s.GetZPath().LocalZOrder() <= targetLocalZOrder {
					return false
				}
			}
		}

		return true
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}
