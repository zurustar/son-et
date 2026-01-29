// Package graphics provides sprite-based rendering system.
package graphics

import (
	"fmt"
	"sync"
)

// ZPath は階層的Z順序を表す
// 要件 1.1: スプライトはZ_Pathを持つ（整数配列として表現）
//
// Z_Pathは親子関係に基づいた多次元のZ順序を表現します。
// 例:
//   - ウインドウ0: [0]
//   - ウインドウ0の背景: [0, 0]
//   - ウインドウ0のキャスト1: [0, 1]
//   - ウインドウ0のテキスト1: [0, 2]
type ZPath struct {
	path []int
}

// NewZPath は新しいZPathを作成する
// 可変長引数で初期化できます。
//
// 例:
//
//	zpath := NewZPath(0, 1, 2) // [0, 1, 2]
//	root := NewZPath(0)        // [0]
//	empty := NewZPath()        // []
func NewZPath(path ...int) *ZPath {
	return &ZPath{path: append([]int{}, path...)}
}

// NewZPathFromParent は親のZPathに子のローカルZ順序を追加したZPathを作成する
// 要件 1.2: Z_Pathは親のZ_Pathに自身のLocal_Z_Orderを追加した形式
//
// 親がnilの場合は、単一要素のZPathを作成します（ルートスプライト用）。
//
// 例:
//
//	parent := NewZPath(0)
//	child := NewZPathFromParent(parent, 1) // [0, 1]
//	root := NewZPathFromParent(nil, 0)     // [0]
func NewZPathFromParent(parent *ZPath, localZOrder int) *ZPath {
	if parent == nil {
		return NewZPath(localZOrder)
	}
	newPath := make([]int, len(parent.path)+1)
	copy(newPath, parent.path)
	newPath[len(parent.path)] = localZOrder
	return &ZPath{path: newPath}
}

// Path はZ_Pathの配列のコピーを返す
// 返される配列は元のZPathとは独立しています。
func (z *ZPath) Path() []int {
	if z == nil {
		return nil
	}
	result := make([]int, len(z.path))
	copy(result, z.path)
	return result
}

// Depth はZ_Pathの深さ（要素数）を返す
// ルートスプライトの深さは1、その子の深さは2、というように増加します。
func (z *ZPath) Depth() int {
	if z == nil {
		return 0
	}
	return len(z.path)
}

// LocalZOrder は最後の要素（ローカルZ順序）を返す
// 空のZPathの場合は0を返します。
func (z *ZPath) LocalZOrder() int {
	if z == nil || len(z.path) == 0 {
		return 0
	}
	return z.path[len(z.path)-1]
}

// Parent は親のZ_Pathを返す
// ルートスプライト（深さ1）または空のZPathの場合はnilを返します。
//
// 例:
//
//	zpath := NewZPath(0, 1, 2)
//	parent := zpath.Parent() // [0, 1]
//	grandparent := parent.Parent() // [0]
//	root := grandparent.Parent() // nil
func (z *ZPath) Parent() *ZPath {
	if z == nil || len(z.path) <= 1 {
		return nil
	}
	parentPath := make([]int, len(z.path)-1)
	copy(parentPath, z.path[:len(z.path)-1])
	return &ZPath{path: parentPath}
}

// String はZ_Pathの文字列表現を返す
// 要件 10.1: スプライトのZ_Pathを文字列として取得できる
//
// 例:
//
//	zpath := NewZPath(0, 1, 2)
//	fmt.Println(zpath.String()) // "[0 1 2]"
func (z *ZPath) String() string {
	if z == nil {
		return "nil"
	}
	return fmt.Sprintf("%v", z.path)
}

// Compare はZ_Pathを辞書順で比較する
// 要件 5.1: Z_Pathを辞書順（lexicographic order）で比較する
// 戻り値: -1 (z < other), 0 (z == other), 1 (z > other)
//
// 比較ルール:
//   - 要件 5.2: Z_Path Aの先頭がZ_Path Bの先頭と一致するとき、次の要素を比較する
//   - 要件 5.3: Z_Path AがZ_Path Bのプレフィックスであるとき、AをBより前（背面）と判定する
//
// 例:
//
//	[0, 1].Compare([0, 2]) // -1 (0,1 < 0,2)
//	[0, 2].Compare([0, 1]) // 1  (0,2 > 0,1)
//	[0, 1].Compare([0, 1]) // 0  (等しい)
//	[0].Compare([0, 1])    // -1 (プレフィックスは背面)
//	[0, 1].Compare([0])    // 1  (プレフィックスより前面)
func (z *ZPath) Compare(other *ZPath) int {
	// nilの処理
	if z == nil && other == nil {
		return 0
	}
	if z == nil {
		return -1
	}
	if other == nil {
		return 1
	}

	// 比較する長さを決定
	minLen := len(z.path)
	if len(other.path) < minLen {
		minLen = len(other.path)
	}

	// 要件 5.2: Z_Path Aの先頭がZ_Path Bの先頭と一致するとき、次の要素を比較する
	for i := 0; i < minLen; i++ {
		if z.path[i] < other.path[i] {
			return -1
		}
		if z.path[i] > other.path[i] {
			return 1
		}
	}

	// 要件 5.3: Z_Path AがZ_Path Bのプレフィックスであるとき、AをBより前（背面）と判定する
	if len(z.path) < len(other.path) {
		return -1
	}
	if len(z.path) > len(other.path) {
		return 1
	}

	return 0
}

// Less は z < other かどうかを返す（sort.Interface用）
// 要件 5.1: Z_Pathを辞書順（lexicographic order）で比較する
//
// 例:
//
//	[0, 1].Less([0, 2]) // true
//	[0, 2].Less([0, 1]) // false
//	[0, 1].Less([0, 1]) // false
func (z *ZPath) Less(other *ZPath) bool {
	return z.Compare(other) < 0
}

// IsPrefix は z が other のプレフィックスかどうかを返す
// 要件 5.3: Z_Path AがZ_Path Bのプレフィックスであるとき、AをBより前（背面）と判定する
//
// 例:
//
//	[0].IsPrefix([0, 1])    // true
//	[0, 1].IsPrefix([0, 1]) // true (自身もプレフィックス)
//	[0, 1].IsPrefix([0])    // false
//	[0, 2].IsPrefix([0, 1]) // false
func (z *ZPath) IsPrefix(other *ZPath) bool {
	// nilの処理
	if z == nil {
		return true // 空のパスは任意のパスのプレフィックス
	}
	if other == nil {
		return len(z.path) == 0 // 空のパスのみがnilのプレフィックス
	}

	// zがotherより長い場合はプレフィックスではない
	if len(z.path) > len(other.path) {
		return false
	}

	// 各要素を比較
	for i := 0; i < len(z.path); i++ {
		if z.path[i] != other.path[i] {
			return false
		}
	}

	return true
}

// Equal は z と other が等しいかどうかを返す
// 要件 5.1: Z_Pathを辞書順（lexicographic order）で比較する
//
// 例:
//
//	[0, 1].Equal([0, 1]) // true
//	[0, 1].Equal([0, 2]) // false
//	[0].Equal([0, 1])    // false
func (z *ZPath) Equal(other *ZPath) bool {
	return z.Compare(other) == 0
}

// ZOrderCounter は操作順序を追跡するカウンター
// 要件 2.1: 各親スプライトごとにZ_Order_Counterを管理する
//
// ZOrderCounterは、スプライトが作成される際にLocal_Z_Orderを割り当てるために使用されます。
// 各親スプライトIDごとに独立したカウンターを管理し、スレッドセーフな操作を提供します。
//
// 使用例:
//
//	counter := NewZOrderCounter()
//	localZ1 := counter.GetNext(0) // 親ID 0 の最初の子: 0
//	localZ2 := counter.GetNext(0) // 親ID 0 の2番目の子: 1
//	localZ3 := counter.GetNext(1) // 親ID 1 の最初の子: 0
type ZOrderCounter struct {
	counters map[int]int  // parentSpriteID -> counter
	mu       sync.RWMutex // スレッドセーフな操作のためのミューテックス
}

// NewZOrderCounter は新しいZOrderCounterを作成する
//
// 返されるZOrderCounterは空のカウンターマップを持ち、
// すぐに使用可能な状態です。
func NewZOrderCounter() *ZOrderCounter {
	return &ZOrderCounter{
		counters: make(map[int]int),
	}
}

// GetNext は指定された親スプライトの次のZ順序を取得し、カウンターをインクリメントする
// 要件 2.5: スプライトが作成されたとき、Z_Order_Counterをインクリメントする
//
// この関数はアトミックに以下の操作を行います:
//  1. 現在のカウンター値を取得（存在しない場合は0）
//  2. カウンターをインクリメント
//  3. 取得した値を返す
//
// parentIDが初めて使用される場合、カウンターは0から開始します。
//
// 例:
//
//	counter := NewZOrderCounter()
//	counter.GetNext(0) // 0を返し、カウンターを1に
//	counter.GetNext(0) // 1を返し、カウンターを2に
//	counter.GetNext(1) // 0を返し（別の親）、カウンターを1に
func (c *ZOrderCounter) GetNext(parentID int) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	current := c.counters[parentID]
	c.counters[parentID] = current + 1
	return current
}

// Reset は指定された親スプライトのカウンターをリセットする
//
// 指定されたparentIDのカウンターを削除します。
// 次にGetNextが呼ばれた際は、再び0から開始します。
//
// 例:
//
//	counter := NewZOrderCounter()
//	counter.GetNext(0) // 0
//	counter.GetNext(0) // 1
//	counter.Reset(0)
//	counter.GetNext(0) // 0（リセットされた）
func (c *ZOrderCounter) Reset(parentID int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.counters, parentID)
}

// ResetAll はすべてのカウンターをリセットする
//
// すべての親スプライトのカウンターを削除します。
// 通常、シーンの切り替えやウインドウのクリア時に使用されます。
func (c *ZOrderCounter) ResetAll() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.counters = make(map[int]int)
}

// GetCurrentCount は指定された親スプライトの現在のカウンター値を取得する（インクリメントしない）
//
// デバッグや状態確認のために使用します。
// カウンターが存在しない場合は0を返します。
func (c *ZOrderCounter) GetCurrentCount(parentID int) int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.counters[parentID]
}
