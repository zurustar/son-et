package graphics

import (
	"testing"
)

// TestNewZPath はNewZPath関数のテスト
func TestNewZPath(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected []int
	}{
		{
			name:     "空のZPath",
			input:    []int{},
			expected: []int{},
		},
		{
			name:     "単一要素のZPath（ルートスプライト）",
			input:    []int{0},
			expected: []int{0},
		},
		{
			name:     "複数要素のZPath",
			input:    []int{0, 1, 2},
			expected: []int{0, 1, 2},
		},
		{
			name:     "負の値を含むZPath",
			input:    []int{-1, 0, 1},
			expected: []int{-1, 0, 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zpath := NewZPath(tt.input...)
			result := zpath.Path()

			if len(result) != len(tt.expected) {
				t.Errorf("Path() length = %d, want %d", len(result), len(tt.expected))
				return
			}

			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("Path()[%d] = %d, want %d", i, v, tt.expected[i])
				}
			}
		})
	}
}

// TestNewZPath_Immutability はNewZPathが元の配列から独立していることを確認
func TestNewZPath_Immutability(t *testing.T) {
	original := []int{0, 1, 2}
	zpath := NewZPath(original...)

	// 元の配列を変更
	original[0] = 999

	// ZPathは影響を受けないはず
	path := zpath.Path()
	if path[0] != 0 {
		t.Errorf("ZPath was affected by original array modification: got %d, want 0", path[0])
	}
}

// TestNewZPathFromParent はNewZPathFromParent関数のテスト
func TestNewZPathFromParent(t *testing.T) {
	tests := []struct {
		name        string
		parent      *ZPath
		localZOrder int
		expected    []int
	}{
		{
			name:        "親がnilの場合（ルートスプライト）",
			parent:      nil,
			localZOrder: 0,
			expected:    []int{0},
		},
		{
			name:        "親がルートスプライトの場合",
			parent:      NewZPath(0),
			localZOrder: 1,
			expected:    []int{0, 1},
		},
		{
			name:        "深い階層の場合",
			parent:      NewZPath(0, 1),
			localZOrder: 2,
			expected:    []int{0, 1, 2},
		},
		{
			name:        "負のローカルZ順序",
			parent:      NewZPath(0),
			localZOrder: -1,
			expected:    []int{0, -1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zpath := NewZPathFromParent(tt.parent, tt.localZOrder)
			result := zpath.Path()

			if len(result) != len(tt.expected) {
				t.Errorf("Path() length = %d, want %d", len(result), len(tt.expected))
				return
			}

			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("Path()[%d] = %d, want %d", i, v, tt.expected[i])
				}
			}
		})
	}
}

// TestZPath_Depth はDepth()メソッドのテスト
func TestZPath_Depth(t *testing.T) {
	tests := []struct {
		name     string
		zpath    *ZPath
		expected int
	}{
		{
			name:     "nilのZPath",
			zpath:    nil,
			expected: 0,
		},
		{
			name:     "空のZPath",
			zpath:    NewZPath(),
			expected: 0,
		},
		{
			name:     "深さ1（ルートスプライト）",
			zpath:    NewZPath(0),
			expected: 1,
		},
		{
			name:     "深さ3",
			zpath:    NewZPath(0, 1, 2),
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.zpath.Depth()
			if result != tt.expected {
				t.Errorf("Depth() = %d, want %d", result, tt.expected)
			}
		})
	}
}

// TestZPath_LocalZOrder はLocalZOrder()メソッドのテスト
func TestZPath_LocalZOrder(t *testing.T) {
	tests := []struct {
		name     string
		zpath    *ZPath
		expected int
	}{
		{
			name:     "nilのZPath",
			zpath:    nil,
			expected: 0,
		},
		{
			name:     "空のZPath",
			zpath:    NewZPath(),
			expected: 0,
		},
		{
			name:     "単一要素",
			zpath:    NewZPath(5),
			expected: 5,
		},
		{
			name:     "複数要素",
			zpath:    NewZPath(0, 1, 7),
			expected: 7,
		},
		{
			name:     "負の値",
			zpath:    NewZPath(0, -3),
			expected: -3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.zpath.LocalZOrder()
			if result != tt.expected {
				t.Errorf("LocalZOrder() = %d, want %d", result, tt.expected)
			}
		})
	}
}

// TestZPath_Parent はParent()メソッドのテスト
func TestZPath_Parent(t *testing.T) {
	tests := []struct {
		name           string
		zpath          *ZPath
		expectedNil    bool
		expectedParent []int
	}{
		{
			name:        "nilのZPath",
			zpath:       nil,
			expectedNil: true,
		},
		{
			name:        "空のZPath",
			zpath:       NewZPath(),
			expectedNil: true,
		},
		{
			name:        "ルートスプライト（深さ1）",
			zpath:       NewZPath(0),
			expectedNil: true,
		},
		{
			name:           "深さ2",
			zpath:          NewZPath(0, 1),
			expectedNil:    false,
			expectedParent: []int{0},
		},
		{
			name:           "深さ3",
			zpath:          NewZPath(0, 1, 2),
			expectedNil:    false,
			expectedParent: []int{0, 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parent := tt.zpath.Parent()

			if tt.expectedNil {
				if parent != nil {
					t.Errorf("Parent() = %v, want nil", parent.Path())
				}
				return
			}

			if parent == nil {
				t.Errorf("Parent() = nil, want %v", tt.expectedParent)
				return
			}

			result := parent.Path()
			if len(result) != len(tt.expectedParent) {
				t.Errorf("Parent().Path() length = %d, want %d", len(result), len(tt.expectedParent))
				return
			}

			for i, v := range result {
				if v != tt.expectedParent[i] {
					t.Errorf("Parent().Path()[%d] = %d, want %d", i, v, tt.expectedParent[i])
				}
			}
		})
	}
}

// TestZPath_Parent_Immutability はParent()が元のZPathから独立していることを確認
func TestZPath_Parent_Immutability(t *testing.T) {
	zpath := NewZPath(0, 1, 2)
	parent := zpath.Parent()

	// 親のパスを取得して変更
	parentPath := parent.Path()
	parentPath[0] = 999

	// 元のZPathは影響を受けないはず
	originalPath := zpath.Path()
	if originalPath[0] != 0 {
		t.Errorf("Original ZPath was affected by parent modification: got %d, want 0", originalPath[0])
	}
}

// TestZPath_String はString()メソッドのテスト
func TestZPath_String(t *testing.T) {
	tests := []struct {
		name     string
		zpath    *ZPath
		expected string
	}{
		{
			name:     "nilのZPath",
			zpath:    nil,
			expected: "nil",
		},
		{
			name:     "空のZPath",
			zpath:    NewZPath(),
			expected: "[]",
		},
		{
			name:     "単一要素",
			zpath:    NewZPath(0),
			expected: "[0]",
		},
		{
			name:     "複数要素",
			zpath:    NewZPath(0, 1, 2),
			expected: "[0 1 2]",
		},
		{
			name:     "負の値を含む",
			zpath:    NewZPath(-1, 0, 1),
			expected: "[-1 0 1]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.zpath.String()
			if result != tt.expected {
				t.Errorf("String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestZPath_Path_Immutability はPath()が返す配列が元のZPathから独立していることを確認
func TestZPath_Path_Immutability(t *testing.T) {
	zpath := NewZPath(0, 1, 2)
	path := zpath.Path()

	// 返された配列を変更
	path[0] = 999

	// 元のZPathは影響を受けないはず
	originalPath := zpath.Path()
	if originalPath[0] != 0 {
		t.Errorf("Original ZPath was affected by returned path modification: got %d, want 0", originalPath[0])
	}
}

// TestZPath_Path_NilReceiver はnilレシーバーでPath()を呼び出した場合のテスト
func TestZPath_Path_NilReceiver(t *testing.T) {
	var zpath *ZPath = nil
	result := zpath.Path()
	if result != nil {
		t.Errorf("Path() on nil receiver = %v, want nil", result)
	}
}

// TestZPath_Compare はCompare()メソッドのテスト
// 要件 5.1: Z_Pathを辞書順（lexicographic order）で比較する
// 要件 5.2: Z_Path Aの先頭がZ_Path Bの先頭と一致するとき、次の要素を比較する
// 要件 5.3: Z_Path AがZ_Path Bのプレフィックスであるとき、AをBより前（背面）と判定する
func TestZPath_Compare(t *testing.T) {
	tests := []struct {
		name     string
		z        *ZPath
		other    *ZPath
		expected int
	}{
		// 基本的な比較
		{
			name:     "等しいZPath",
			z:        NewZPath(0, 1, 2),
			other:    NewZPath(0, 1, 2),
			expected: 0,
		},
		{
			name:     "zが小さい（最初の要素で決定）",
			z:        NewZPath(0, 1),
			other:    NewZPath(1, 0),
			expected: -1,
		},
		{
			name:     "zが大きい（最初の要素で決定）",
			z:        NewZPath(1, 0),
			other:    NewZPath(0, 1),
			expected: 1,
		},
		{
			name:     "zが小さい（2番目の要素で決定）",
			z:        NewZPath(0, 1),
			other:    NewZPath(0, 2),
			expected: -1,
		},
		{
			name:     "zが大きい（2番目の要素で決定）",
			z:        NewZPath(0, 2),
			other:    NewZPath(0, 1),
			expected: 1,
		},
		// プレフィックスの比較（要件 5.3）
		{
			name:     "zがotherのプレフィックス（背面）",
			z:        NewZPath(0),
			other:    NewZPath(0, 1),
			expected: -1,
		},
		{
			name:     "otherがzのプレフィックス（前面）",
			z:        NewZPath(0, 1),
			other:    NewZPath(0),
			expected: 1,
		},
		{
			name:     "深いプレフィックス",
			z:        NewZPath(0, 1),
			other:    NewZPath(0, 1, 2),
			expected: -1,
		},
		// 空のZPath
		{
			name:     "両方空",
			z:        NewZPath(),
			other:    NewZPath(),
			expected: 0,
		},
		{
			name:     "zが空、otherが非空",
			z:        NewZPath(),
			other:    NewZPath(0),
			expected: -1,
		},
		{
			name:     "zが非空、otherが空",
			z:        NewZPath(0),
			other:    NewZPath(),
			expected: 1,
		},
		// nilの処理
		{
			name:     "両方nil",
			z:        nil,
			other:    nil,
			expected: 0,
		},
		{
			name:     "zがnil、otherが非nil",
			z:        nil,
			other:    NewZPath(0),
			expected: -1,
		},
		{
			name:     "zが非nil、otherがnil",
			z:        NewZPath(0),
			other:    nil,
			expected: 1,
		},
		// 負の値を含む比較
		{
			name:     "負の値を含む（zが小さい）",
			z:        NewZPath(-1, 0),
			other:    NewZPath(0, 0),
			expected: -1,
		},
		{
			name:     "負の値を含む（zが大きい）",
			z:        NewZPath(0, 0),
			other:    NewZPath(-1, 0),
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.z.Compare(tt.other)
			if result != tt.expected {
				t.Errorf("Compare() = %d, want %d", result, tt.expected)
			}
		})
	}
}

// TestZPath_Less はLess()メソッドのテスト
func TestZPath_Less(t *testing.T) {
	tests := []struct {
		name     string
		z        *ZPath
		other    *ZPath
		expected bool
	}{
		{
			name:     "zが小さい",
			z:        NewZPath(0, 1),
			other:    NewZPath(0, 2),
			expected: true,
		},
		{
			name:     "zが大きい",
			z:        NewZPath(0, 2),
			other:    NewZPath(0, 1),
			expected: false,
		},
		{
			name:     "等しい",
			z:        NewZPath(0, 1),
			other:    NewZPath(0, 1),
			expected: false,
		},
		{
			name:     "プレフィックス（背面）",
			z:        NewZPath(0),
			other:    NewZPath(0, 1),
			expected: true,
		},
		{
			name:     "nilのz",
			z:        nil,
			other:    NewZPath(0),
			expected: true,
		},
		{
			name:     "nilのother",
			z:        NewZPath(0),
			other:    nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.z.Less(tt.other)
			if result != tt.expected {
				t.Errorf("Less() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestZPath_IsPrefix はIsPrefix()メソッドのテスト
// 要件 5.3: Z_Path AがZ_Path Bのプレフィックスであるとき、AをBより前（背面）と判定する
func TestZPath_IsPrefix(t *testing.T) {
	tests := []struct {
		name     string
		z        *ZPath
		other    *ZPath
		expected bool
	}{
		{
			name:     "完全なプレフィックス",
			z:        NewZPath(0),
			other:    NewZPath(0, 1),
			expected: true,
		},
		{
			name:     "深いプレフィックス",
			z:        NewZPath(0, 1),
			other:    NewZPath(0, 1, 2),
			expected: true,
		},
		{
			name:     "自身もプレフィックス",
			z:        NewZPath(0, 1),
			other:    NewZPath(0, 1),
			expected: true,
		},
		{
			name:     "プレフィックスではない（長さが長い）",
			z:        NewZPath(0, 1, 2),
			other:    NewZPath(0, 1),
			expected: false,
		},
		{
			name:     "プレフィックスではない（値が異なる）",
			z:        NewZPath(0, 2),
			other:    NewZPath(0, 1, 2),
			expected: false,
		},
		{
			name:     "空のZPathは任意のプレフィックス",
			z:        NewZPath(),
			other:    NewZPath(0, 1),
			expected: true,
		},
		{
			name:     "空のZPathは空のZPathのプレフィックス",
			z:        NewZPath(),
			other:    NewZPath(),
			expected: true,
		},
		{
			name:     "nilは任意のプレフィックス",
			z:        nil,
			other:    NewZPath(0, 1),
			expected: true,
		},
		{
			name:     "nilはnilのプレフィックス",
			z:        nil,
			other:    nil,
			expected: true,
		},
		{
			name:     "非空はnilのプレフィックスではない",
			z:        NewZPath(0),
			other:    nil,
			expected: false,
		},
		{
			name:     "空はnilのプレフィックス",
			z:        NewZPath(),
			other:    nil,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.z.IsPrefix(tt.other)
			if result != tt.expected {
				t.Errorf("IsPrefix() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestZPath_Equal はEqual()メソッドのテスト
func TestZPath_Equal(t *testing.T) {
	tests := []struct {
		name     string
		z        *ZPath
		other    *ZPath
		expected bool
	}{
		{
			name:     "等しい",
			z:        NewZPath(0, 1, 2),
			other:    NewZPath(0, 1, 2),
			expected: true,
		},
		{
			name:     "異なる（値が違う）",
			z:        NewZPath(0, 1),
			other:    NewZPath(0, 2),
			expected: false,
		},
		{
			name:     "異なる（長さが違う）",
			z:        NewZPath(0),
			other:    NewZPath(0, 1),
			expected: false,
		},
		{
			name:     "両方空",
			z:        NewZPath(),
			other:    NewZPath(),
			expected: true,
		},
		{
			name:     "両方nil",
			z:        nil,
			other:    nil,
			expected: true,
		},
		{
			name:     "一方がnil",
			z:        NewZPath(0),
			other:    nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.z.Equal(tt.other)
			if result != tt.expected {
				t.Errorf("Equal() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestZPath_Compare_Transitivity は比較の推移性をテスト
// a < b かつ b < c ならば a < c
func TestZPath_Compare_Transitivity(t *testing.T) {
	a := NewZPath(0, 1)
	b := NewZPath(0, 2)
	c := NewZPath(0, 3)

	if a.Compare(b) >= 0 {
		t.Errorf("Expected a < b, but got %d", a.Compare(b))
	}
	if b.Compare(c) >= 0 {
		t.Errorf("Expected b < c, but got %d", b.Compare(c))
	}
	if a.Compare(c) >= 0 {
		t.Errorf("Expected a < c (transitivity), but got %d", a.Compare(c))
	}
}

// TestZPath_Compare_Antisymmetry は比較の反対称性をテスト
// a < b ならば b > a
func TestZPath_Compare_Antisymmetry(t *testing.T) {
	a := NewZPath(0, 1)
	b := NewZPath(0, 2)

	if a.Compare(b) != -1 {
		t.Errorf("Expected a < b, but got %d", a.Compare(b))
	}
	if b.Compare(a) != 1 {
		t.Errorf("Expected b > a (antisymmetry), but got %d", b.Compare(a))
	}
}

// TestZPath_Compare_Reflexivity は比較の反射性をテスト
// a == a
func TestZPath_Compare_Reflexivity(t *testing.T) {
	a := NewZPath(0, 1, 2)

	if a.Compare(a) != 0 {
		t.Errorf("Expected a == a (reflexivity), but got %d", a.Compare(a))
	}
}

// =============================================================================
// ZOrderCounter のテスト
// =============================================================================

// TestNewZOrderCounter はNewZOrderCounter関数のテスト
func TestNewZOrderCounter(t *testing.T) {
	counter := NewZOrderCounter()

	if counter == nil {
		t.Fatal("NewZOrderCounter() returned nil")
	}

	if counter.counters == nil {
		t.Error("counters map is nil")
	}

	if len(counter.counters) != 0 {
		t.Errorf("counters map should be empty, got %d entries", len(counter.counters))
	}
}

// TestZOrderCounter_GetNext はGetNext関数のテスト
// 要件 2.5: スプライトが作成されたとき、Z_Order_Counterをインクリメントする
func TestZOrderCounter_GetNext(t *testing.T) {
	counter := NewZOrderCounter()

	// 最初の呼び出しは0を返す
	result := counter.GetNext(0)
	if result != 0 {
		t.Errorf("GetNext(0) first call = %d, want 0", result)
	}

	// 2回目の呼び出しは1を返す
	result = counter.GetNext(0)
	if result != 1 {
		t.Errorf("GetNext(0) second call = %d, want 1", result)
	}

	// 3回目の呼び出しは2を返す
	result = counter.GetNext(0)
	if result != 2 {
		t.Errorf("GetNext(0) third call = %d, want 2", result)
	}
}

// TestZOrderCounter_GetNext_MultipleParents は複数の親IDでのGetNextのテスト
// 要件 2.1: 各親スプライトごとにZ_Order_Counterを管理する
func TestZOrderCounter_GetNext_MultipleParents(t *testing.T) {
	counter := NewZOrderCounter()

	// 親ID 0 の最初の子
	result := counter.GetNext(0)
	if result != 0 {
		t.Errorf("GetNext(0) = %d, want 0", result)
	}

	// 親ID 1 の最初の子（別のカウンター）
	result = counter.GetNext(1)
	if result != 0 {
		t.Errorf("GetNext(1) = %d, want 0", result)
	}

	// 親ID 0 の2番目の子
	result = counter.GetNext(0)
	if result != 1 {
		t.Errorf("GetNext(0) second call = %d, want 1", result)
	}

	// 親ID 1 の2番目の子
	result = counter.GetNext(1)
	if result != 1 {
		t.Errorf("GetNext(1) second call = %d, want 1", result)
	}

	// 親ID 2 の最初の子
	result = counter.GetNext(2)
	if result != 0 {
		t.Errorf("GetNext(2) = %d, want 0", result)
	}
}

// TestZOrderCounter_Reset はReset関数のテスト
func TestZOrderCounter_Reset(t *testing.T) {
	counter := NewZOrderCounter()

	// カウンターを進める
	counter.GetNext(0) // 0
	counter.GetNext(0) // 1
	counter.GetNext(0) // 2

	// リセット
	counter.Reset(0)

	// リセット後は0から開始
	result := counter.GetNext(0)
	if result != 0 {
		t.Errorf("GetNext(0) after Reset = %d, want 0", result)
	}
}

// TestZOrderCounter_Reset_OnlyAffectsSpecifiedParent はResetが指定された親のみに影響することを確認
func TestZOrderCounter_Reset_OnlyAffectsSpecifiedParent(t *testing.T) {
	counter := NewZOrderCounter()

	// 複数の親のカウンターを進める
	counter.GetNext(0) // 0
	counter.GetNext(0) // 1
	counter.GetNext(1) // 0
	counter.GetNext(1) // 1

	// 親ID 0 のみリセット
	counter.Reset(0)

	// 親ID 0 は0から開始
	result := counter.GetNext(0)
	if result != 0 {
		t.Errorf("GetNext(0) after Reset = %d, want 0", result)
	}

	// 親ID 1 は影響を受けない
	result = counter.GetNext(1)
	if result != 2 {
		t.Errorf("GetNext(1) after Reset(0) = %d, want 2", result)
	}
}

// TestZOrderCounter_ResetAll はResetAll関数のテスト
func TestZOrderCounter_ResetAll(t *testing.T) {
	counter := NewZOrderCounter()

	// 複数の親のカウンターを進める
	counter.GetNext(0) // 0
	counter.GetNext(0) // 1
	counter.GetNext(1) // 0
	counter.GetNext(1) // 1
	counter.GetNext(2) // 0

	// すべてリセット
	counter.ResetAll()

	// すべての親が0から開始
	tests := []int{0, 1, 2}
	for _, parentID := range tests {
		result := counter.GetNext(parentID)
		if result != 0 {
			t.Errorf("GetNext(%d) after ResetAll = %d, want 0", parentID, result)
		}
	}
}

// TestZOrderCounter_GetCurrentCount はGetCurrentCount関数のテスト
func TestZOrderCounter_GetCurrentCount(t *testing.T) {
	counter := NewZOrderCounter()

	// 初期状態は0
	result := counter.GetCurrentCount(0)
	if result != 0 {
		t.Errorf("GetCurrentCount(0) initial = %d, want 0", result)
	}

	// GetNextを呼び出すとカウンターが増加
	counter.GetNext(0) // 0を返し、カウンターを1に
	result = counter.GetCurrentCount(0)
	if result != 1 {
		t.Errorf("GetCurrentCount(0) after GetNext = %d, want 1", result)
	}

	// GetCurrentCountはカウンターを変更しない
	result = counter.GetCurrentCount(0)
	if result != 1 {
		t.Errorf("GetCurrentCount(0) second call = %d, want 1", result)
	}

	// 存在しない親IDは0を返す
	result = counter.GetCurrentCount(999)
	if result != 0 {
		t.Errorf("GetCurrentCount(999) = %d, want 0", result)
	}
}

// TestZOrderCounter_NegativeParentID は負の親IDでの動作テスト
func TestZOrderCounter_NegativeParentID(t *testing.T) {
	counter := NewZOrderCounter()

	// 負の親IDも正常に動作する
	result := counter.GetNext(-1)
	if result != 0 {
		t.Errorf("GetNext(-1) = %d, want 0", result)
	}

	result = counter.GetNext(-1)
	if result != 1 {
		t.Errorf("GetNext(-1) second call = %d, want 1", result)
	}

	// リセットも正常に動作
	counter.Reset(-1)
	result = counter.GetNext(-1)
	if result != 0 {
		t.Errorf("GetNext(-1) after Reset = %d, want 0", result)
	}
}

// TestZOrderCounter_Reset_NonExistentParent は存在しない親IDのリセットテスト
func TestZOrderCounter_Reset_NonExistentParent(t *testing.T) {
	counter := NewZOrderCounter()

	// 存在しない親IDのリセットはエラーにならない
	counter.Reset(999) // パニックしないことを確認

	// その後も正常に動作
	result := counter.GetNext(999)
	if result != 0 {
		t.Errorf("GetNext(999) after Reset = %d, want 0", result)
	}
}
