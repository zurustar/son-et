package graphics

import (
	"image"
	"image/color"
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"
)

// ============================================================
// WindowLayerSetのプロパティベーステスト
// **Validates: Requirements 1.1, 1.2, 1.3, 1.5**
// ============================================================

// ============================================================
// Property 1: レイヤーのWindowID管理
// *任意の*ウィンドウIDとレイヤーに対して、そのウィンドウIDでレイヤーを登録した場合、
// 同じウィンドウIDで検索するとそのレイヤーが見つかる。
// **Validates: Requirements 1.1, 1.5**
// ============================================================

// TestProperty1_LayerWindowIDManagement_Registration はレイヤー登録と検索の一貫性をテストする
// Property 1: レイヤーのWindowID管理
// **Validates: Requirements 1.1, 1.5**
func TestProperty1_LayerWindowIDManagement_Registration(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: 任意のウィンドウIDでWindowLayerSetを作成した場合、
	// 同じウィンドウIDで検索するとそのWindowLayerSetが見つかる
	property := func(winID int, width, height uint16) bool {
		// 負のウィンドウIDは許可しない（実際のシステムでは正のIDのみ使用）
		if winID < 0 {
			winID = -winID
		}

		// サイズは最小1以上
		w := int(width%1000) + 1
		h := int(height%1000) + 1

		lm := NewLayerManager()
		bgColor := color.RGBA{0, 0, 0, 255}

		// WindowLayerSetを作成
		created := lm.GetOrCreateWindowLayerSet(winID, w, h, bgColor)
		if created == nil {
			return false
		}

		// 同じウィンドウIDで検索
		found := lm.GetWindowLayerSet(winID)
		if found == nil {
			return false
		}

		// 同じインスタンスであることを確認
		if found != created {
			return false
		}

		// ウィンドウIDが正しいことを確認
		if found.GetWinID() != winID {
			return false
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 1 (Registration) failed: %v", err)
	}
}

// TestProperty1_LayerWindowIDManagement_LayerRetrieval はレイヤー追加と検索の一貫性をテストする
// Property 1: レイヤーのWindowID管理
// **Validates: Requirements 1.1, 1.5**
func TestProperty1_LayerWindowIDManagement_LayerRetrieval(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: 任意のウィンドウIDにレイヤーを追加した場合、
	// そのウィンドウIDで検索するとレイヤーが見つかる
	property := func(winID int, layerID int) bool {
		// 負のIDは正に変換
		if winID < 0 {
			winID = -winID
		}
		if layerID < 0 {
			layerID = -layerID
		}

		lm := NewLayerManager()
		bgColor := color.RGBA{0, 0, 0, 255}

		// WindowLayerSetを作成
		wls := lm.GetOrCreateWindowLayerSet(winID, 640, 480, bgColor)
		if wls == nil {
			return false
		}

		// レイヤーを作成して追加
		layer := &mockLayer{}
		layer.SetID(layerID)
		layer.SetBounds(image.Rect(0, 0, 100, 100))
		layer.SetVisible(true)

		wls.AddLayer(layer)

		// レイヤーが見つかることを確認
		found := wls.GetLayer(layerID)
		if found == nil {
			return false
		}

		// 同じレイヤーIDであることを確認
		if found.GetID() != layerID {
			return false
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 1 (Layer Retrieval) failed: %v", err)
	}
}

// TestProperty1_LayerWindowIDManagement_MultipleWindows は複数ウィンドウでのレイヤー管理をテストする
// Property 1: レイヤーのWindowID管理
// **Validates: Requirements 1.1, 1.5**
func TestProperty1_LayerWindowIDManagement_MultipleWindows(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: 複数のウィンドウIDに対して、各ウィンドウのレイヤーは独立して管理される
	property := func(winID1, winID2 int, layerID1, layerID2 int) bool {
		// 負のIDは正に変換
		if winID1 < 0 {
			winID1 = -winID1
		}
		if winID2 < 0 {
			winID2 = -winID2
		}
		if layerID1 < 0 {
			layerID1 = -layerID1
		}
		if layerID2 < 0 {
			layerID2 = -layerID2
		}

		// 同じウィンドウIDの場合はスキップ
		if winID1 == winID2 {
			return true
		}

		lm := NewLayerManager()
		bgColor := color.RGBA{0, 0, 0, 255}

		// 2つのWindowLayerSetを作成
		wls1 := lm.GetOrCreateWindowLayerSet(winID1, 640, 480, bgColor)
		wls2 := lm.GetOrCreateWindowLayerSet(winID2, 800, 600, bgColor)

		if wls1 == nil || wls2 == nil {
			return false
		}

		// 各ウィンドウにレイヤーを追加
		layer1 := &mockLayer{}
		layer1.SetID(layerID1)
		layer1.SetBounds(image.Rect(0, 0, 100, 100))
		wls1.AddLayer(layer1)

		layer2 := &mockLayer{}
		layer2.SetID(layerID2)
		layer2.SetBounds(image.Rect(0, 0, 100, 100))
		wls2.AddLayer(layer2)

		// 各ウィンドウのレイヤーが独立していることを確認
		// wls1にはlayer1のみ
		if wls1.GetLayer(layerID1) == nil {
			return false
		}
		// wls2にはlayer2のみ
		if wls2.GetLayer(layerID2) == nil {
			return false
		}

		// レイヤーIDが異なる場合、クロスチェック
		if layerID1 != layerID2 {
			// wls1にはlayer2がない
			if wls1.GetLayer(layerID2) != nil {
				return false
			}
			// wls2にはlayer1がない
			if wls2.GetLayer(layerID1) != nil {
				return false
			}
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 1 (Multiple Windows) failed: %v", err)
	}
}

// ============================================================
// Property 2: ウィンドウ開閉時のレイヤーセット管理
// *任意の*ウィンドウに対して、ウィンドウを開いた後はWindowLayerSetが存在し、
// ウィンドウを閉じた後はそのウィンドウに属するすべてのレイヤーが削除される。
// **Validates: Requirements 1.2, 1.3**
// ============================================================

// TestProperty2_WindowLifecycle_OpenClose はウィンドウ開閉時のWindowLayerSet管理をテストする
// Property 2: ウィンドウ開閉時のレイヤーセット管理
// **Validates: Requirements 1.2, 1.3**
func TestProperty2_WindowLifecycle_OpenClose(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: ウィンドウを開いた後はWindowLayerSetが存在し、
	// ウィンドウを閉じた後はWindowLayerSetが削除される
	property := func(winID int) bool {
		// 負のIDは正に変換
		if winID < 0 {
			winID = -winID
		}

		lm := NewLayerManager()
		bgColor := color.RGBA{0, 0, 0, 255}

		// 開く前は存在しない
		if lm.GetWindowLayerSet(winID) != nil {
			return false
		}

		// ウィンドウを開く（WindowLayerSetを作成）
		wls := lm.GetOrCreateWindowLayerSet(winID, 640, 480, bgColor)
		if wls == nil {
			return false
		}

		// 開いた後は存在する
		if lm.GetWindowLayerSet(winID) == nil {
			return false
		}

		// ウィンドウを閉じる（WindowLayerSetを削除）
		lm.DeleteWindowLayerSet(winID)

		// 閉じた後は存在しない
		if lm.GetWindowLayerSet(winID) != nil {
			return false
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 2 (Open/Close) failed: %v", err)
	}
}

// TestProperty2_WindowLifecycle_LayerDeletion はウィンドウ閉鎖時のレイヤー削除をテストする
// Property 2: ウィンドウ開閉時のレイヤーセット管理
// **Validates: Requirements 1.2, 1.3**
func TestProperty2_WindowLifecycle_LayerDeletion(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: ウィンドウを閉じた後は、そのウィンドウに属するすべてのレイヤーが削除される
	property := func(winID int, numLayers uint8) bool {
		// 負のIDは正に変換
		if winID < 0 {
			winID = -winID
		}

		// レイヤー数を制限（0-10）
		layerCount := int(numLayers % 11)

		lm := NewLayerManager()
		bgColor := color.RGBA{0, 0, 0, 255}

		// ウィンドウを開く
		wls := lm.GetOrCreateWindowLayerSet(winID, 640, 480, bgColor)
		if wls == nil {
			return false
		}

		// レイヤーを追加
		for i := 0; i < layerCount; i++ {
			layer := &mockLayer{}
			layer.SetID(i + 1)
			layer.SetBounds(image.Rect(0, 0, 100, 100))
			wls.AddLayer(layer)
		}

		// レイヤーが追加されていることを確認
		if wls.GetLayerCount() != layerCount {
			return false
		}

		// ウィンドウを閉じる
		lm.DeleteWindowLayerSet(winID)

		// WindowLayerSetが削除されていることを確認
		if lm.GetWindowLayerSet(winID) != nil {
			return false
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 2 (Layer Deletion) failed: %v", err)
	}
}

// TestProperty2_WindowLifecycle_IndependentWindows は複数ウィンドウの独立性をテストする
// Property 2: ウィンドウ開閉時のレイヤーセット管理
// **Validates: Requirements 1.2, 1.3**
func TestProperty2_WindowLifecycle_IndependentWindows(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: 1つのウィンドウを閉じても、他のウィンドウのレイヤーは影響を受けない
	property := func(winID1, winID2 int, numLayers1, numLayers2 uint8) bool {
		// 負のIDは正に変換
		if winID1 < 0 {
			winID1 = -winID1
		}
		if winID2 < 0 {
			winID2 = -winID2
		}

		// 同じウィンドウIDの場合はスキップ
		if winID1 == winID2 {
			return true
		}

		// レイヤー数を制限（0-10）
		layerCount1 := int(numLayers1 % 11)
		layerCount2 := int(numLayers2 % 11)

		lm := NewLayerManager()
		bgColor := color.RGBA{0, 0, 0, 255}

		// 2つのウィンドウを開く
		wls1 := lm.GetOrCreateWindowLayerSet(winID1, 640, 480, bgColor)
		wls2 := lm.GetOrCreateWindowLayerSet(winID2, 800, 600, bgColor)

		if wls1 == nil || wls2 == nil {
			return false
		}

		// 各ウィンドウにレイヤーを追加
		for i := 0; i < layerCount1; i++ {
			layer := &mockLayer{}
			layer.SetID(i + 1)
			layer.SetBounds(image.Rect(0, 0, 100, 100))
			wls1.AddLayer(layer)
		}

		for i := 0; i < layerCount2; i++ {
			layer := &mockLayer{}
			layer.SetID(i + 1000) // 異なるIDを使用
			layer.SetBounds(image.Rect(0, 0, 100, 100))
			wls2.AddLayer(layer)
		}

		// ウィンドウ1を閉じる
		lm.DeleteWindowLayerSet(winID1)

		// ウィンドウ1は削除されている
		if lm.GetWindowLayerSet(winID1) != nil {
			return false
		}

		// ウィンドウ2は影響を受けない
		wls2After := lm.GetWindowLayerSet(winID2)
		if wls2After == nil {
			return false
		}

		// ウィンドウ2のレイヤー数は変わらない
		if wls2After.GetLayerCount() != layerCount2 {
			return false
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 2 (Independent Windows) failed: %v", err)
	}
}

// ============================================================
// 追加のプロパティテスト: Z順序の一貫性
// **Validates: Requirements 1.1, 1.5**
// ============================================================

// TestProperty_ZOrderConsistency はZ順序の一貫性をテストする
// **Validates: Requirements 1.1, 1.5**
func TestProperty_ZOrderConsistency(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: レイヤーを追加するたびにZ順序が増加する
	property := func(winID int, numLayers uint8) bool {
		// 負のIDは正に変換
		if winID < 0 {
			winID = -winID
		}

		// レイヤー数を制限（1-20）
		layerCount := int(numLayers%20) + 1

		lm := NewLayerManager()
		bgColor := color.RGBA{0, 0, 0, 255}

		wls := lm.GetOrCreateWindowLayerSet(winID, 640, 480, bgColor)
		if wls == nil {
			return false
		}

		// レイヤーを追加
		layers := make([]*mockLayer, layerCount)
		for i := 0; i < layerCount; i++ {
			layer := &mockLayer{}
			layer.SetID(i + 1)
			layer.SetBounds(image.Rect(0, 0, 100, 100))
			layers[i] = layer
			wls.AddLayer(layer)
		}

		// Z順序が昇順であることを確認
		for i := 1; i < layerCount; i++ {
			if layers[i-1].GetZOrder() >= layers[i].GetZOrder() {
				return false
			}
		}

		// GetLayersSortedがZ順序でソートされていることを確認
		sorted := wls.GetLayersSorted()
		for i := 1; i < len(sorted); i++ {
			if sorted[i-1].GetZOrder() > sorted[i].GetZOrder() {
				return false
			}
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property (Z-Order Consistency) failed: %v", err)
	}
}

// ============================================================
// カスタムジェネレータを使用したプロパティテスト
// ============================================================

// WindowOperation はウィンドウ操作を表す
type WindowOperation struct {
	OpType int // 0: Open, 1: Close, 2: AddLayer
	WinID  int
}

// Generate はWindowOperationのランダム生成を実装する
func (WindowOperation) Generate(rand *rand.Rand, size int) reflect.Value {
	op := WindowOperation{
		OpType: rand.Intn(3),
		WinID:  rand.Intn(10), // 0-9のウィンドウID
	}
	return reflect.ValueOf(op)
}

// TestProperty_OperationSequence は操作シーケンスの一貫性をテストする
// **Validates: Requirements 1.1, 1.2, 1.3, 1.5**
func TestProperty_OperationSequence(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: 任意の操作シーケンスに対して、システムは一貫した状態を維持する
	property := func(ops []WindowOperation) bool {
		if len(ops) == 0 {
			return true
		}

		// 操作数を制限
		if len(ops) > 50 {
			ops = ops[:50]
		}

		lm := NewLayerManager()
		bgColor := color.RGBA{0, 0, 0, 255}

		// 期待される状態を追跡
		expectedWindows := make(map[int]bool)
		layerCounter := 0

		for _, op := range ops {
			switch op.OpType {
			case 0: // Open
				lm.GetOrCreateWindowLayerSet(op.WinID, 640, 480, bgColor)
				expectedWindows[op.WinID] = true

			case 1: // Close
				lm.DeleteWindowLayerSet(op.WinID)
				delete(expectedWindows, op.WinID)

			case 2: // AddLayer
				wls := lm.GetWindowLayerSet(op.WinID)
				if wls != nil {
					layer := &mockLayer{}
					layerCounter++
					layer.SetID(layerCounter)
					layer.SetBounds(image.Rect(0, 0, 100, 100))
					wls.AddLayer(layer)
				}
			}
		}

		// 最終状態を検証
		for winID, expected := range expectedWindows {
			wls := lm.GetWindowLayerSet(winID)
			if expected && wls == nil {
				return false
			}
			if !expected && wls != nil {
				return false
			}
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property (Operation Sequence) failed: %v", err)
	}
}

// TestProperty_ClearAllWindows はClearがすべてのWindowLayerSetを削除することをテストする
// **Validates: Requirements 1.3**
func TestProperty_ClearAllWindows(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: Clearを呼び出すと、すべてのWindowLayerSetが削除される
	property := func(winIDs []int) bool {
		if len(winIDs) == 0 {
			return true
		}

		// ウィンドウ数を制限
		if len(winIDs) > 20 {
			winIDs = winIDs[:20]
		}

		lm := NewLayerManager()
		bgColor := color.RGBA{0, 0, 0, 255}

		// 複数のウィンドウを作成
		uniqueWinIDs := make(map[int]bool)
		for _, winID := range winIDs {
			if winID < 0 {
				winID = -winID
			}
			lm.GetOrCreateWindowLayerSet(winID, 640, 480, bgColor)
			uniqueWinIDs[winID] = true
		}

		// Clearを呼び出す
		lm.Clear()

		// すべてのウィンドウが削除されていることを確認
		for winID := range uniqueWinIDs {
			if lm.GetWindowLayerSet(winID) != nil {
				return false
			}
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property (Clear All Windows) failed: %v", err)
	}
}

// ============================================================
// Property 7: Z順序の管理
// *任意の*レイヤー追加操作に対して:
// - 新しいレイヤーには現在のZ順序カウンターが割り当てられる
// - カウンターは操作ごとに増加する
// - すべてのレイヤータイプで共通のカウンターが使用される
// - レイヤーはZ順序（小さい順）でソートされる
// **Validates: Requirements 6.2, 6.3, 6.4**
// ============================================================

// TestProperty7_ZOrder_CounterIncrementsWithEachAddition はZ順序カウンターが各レイヤー追加で増加することをテストする
// Property 7: Z順序の管理
// **Validates: Requirements 6.3**
func TestProperty7_ZOrder_CounterIncrementsWithEachAddition(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: 任意のレイヤー追加操作に対して、Z順序カウンターは1ずつ増加する
	property := func(winID int, numLayers uint8) bool {
		// 負のIDは正に変換
		if winID < 0 {
			winID = -winID
		}

		// レイヤー数を制限（1-30）
		layerCount := int(numLayers%30) + 1

		wls := NewWindowLayerSet(winID, 640, 480, color.RGBA{0, 0, 0, 255})
		if wls == nil {
			return false
		}

		// 初期Z順序カウンターを記録
		initialNextZOrder := wls.GetNextZOrder()

		// レイヤーを追加
		for i := 0; i < layerCount; i++ {
			layer := &mockLayer{}
			layer.SetID(i + 1)
			layer.SetBounds(image.Rect(0, 0, 100, 100))

			// 追加前のZ順序カウンターを記録
			beforeZOrder := wls.GetNextZOrder()

			wls.AddLayer(layer)

			// 追加後のZ順序カウンターが1増加していることを確認
			afterZOrder := wls.GetNextZOrder()
			if afterZOrder != beforeZOrder+1 {
				return false
			}

			// レイヤーに割り当てられたZ順序が追加前のカウンター値であることを確認
			if layer.GetZOrder() != beforeZOrder {
				return false
			}
		}

		// 最終的なZ順序カウンターが初期値 + レイヤー数であることを確認
		if wls.GetNextZOrder() != initialNextZOrder+layerCount {
			return false
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 7 (Counter Increments) failed: %v", err)
	}
}

// TestProperty7_ZOrder_AllLayerTypesShareCounter はすべてのレイヤータイプで共通のZ順序カウンターが使用されることをテストする
// Property 7: Z順序の管理
// **Validates: Requirements 6.4**
func TestProperty7_ZOrder_AllLayerTypesShareCounter(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: Picture、Text、Castの各レイヤータイプを追加しても、
	// 共通のZ順序カウンターが使用される
	property := func(winID int, sequence []uint8) bool {
		// 負のIDは正に変換
		if winID < 0 {
			winID = -winID
		}

		// シーケンス長を制限（1-20）
		if len(sequence) == 0 {
			return true
		}
		if len(sequence) > 20 {
			sequence = sequence[:20]
		}

		lm := NewLayerManager()
		bgColor := color.RGBA{0, 0, 0, 255}

		wls := lm.GetOrCreateWindowLayerSet(winID, 640, 480, bgColor)
		if wls == nil {
			return false
		}

		// 追加されたレイヤーのZ順序を記録
		zOrders := make([]int, 0)

		for i, s := range sequence {
			layerType := s % 3 // 0: Picture, 1: Text, 2: Cast

			var layer Layer
			switch layerType {
			case 0:
				// PictureLayer
				pl := NewPictureLayer(lm.GetNextLayerID(), 640, 480)
				layer = pl
			case 1:
				// TextLayerEntry
				tl := NewTextLayerEntry(lm.GetNextLayerID(), 1, 0, 0, "test", 0)
				layer = tl
			case 2:
				// mockLayer (Cast相当)
				ml := &mockLayer{}
				ml.SetID(lm.GetNextLayerID())
				ml.SetBounds(image.Rect(0, 0, 32, 32))
				ml.layerType = LayerTypeCast
				layer = ml
			}

			wls.AddLayer(layer)
			zOrders = append(zOrders, layer.GetZOrder())

			// Z順序が連続して増加していることを確認
			if i > 0 && zOrders[i] != zOrders[i-1]+1 {
				return false
			}
		}

		// すべてのZ順序が一意であることを確認
		seen := make(map[int]bool)
		for _, z := range zOrders {
			if seen[z] {
				return false
			}
			seen[z] = true
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 7 (All Layer Types Share Counter) failed: %v", err)
	}
}

// TestProperty7_ZOrder_GetLayersSortedReturnsAscendingOrder はGetLayersSortedがZ順序の昇順でレイヤーを返すことをテストする
// Property 7: Z順序の管理
// **Validates: Requirements 6.2**
func TestProperty7_ZOrder_GetLayersSortedReturnsAscendingOrder(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: GetLayersSorted()は常にZ順序の昇順でレイヤーを返す
	property := func(winID int, numLayers uint8) bool {
		// 負のIDは正に変換
		if winID < 0 {
			winID = -winID
		}

		// レイヤー数を制限（1-30）
		layerCount := int(numLayers%30) + 1

		wls := NewWindowLayerSet(winID, 640, 480, color.RGBA{0, 0, 0, 255})
		if wls == nil {
			return false
		}

		// レイヤーを追加
		for i := 0; i < layerCount; i++ {
			layer := &mockLayer{}
			layer.SetID(i + 1)
			layer.SetBounds(image.Rect(0, 0, 100, 100))
			wls.AddLayer(layer)
		}

		// GetLayersSortedを呼び出し
		sorted := wls.GetLayersSorted()

		// レイヤー数が正しいことを確認
		if len(sorted) != layerCount {
			return false
		}

		// Z順序が昇順であることを確認
		for i := 1; i < len(sorted); i++ {
			if sorted[i-1].GetZOrder() >= sorted[i].GetZOrder() {
				return false
			}
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 7 (GetLayersSorted Ascending Order) failed: %v", err)
	}
}

// TestProperty7_ZOrder_MixedLayerTypesPreserveOrder は異なるレイヤータイプを混在させてもZ順序が保持されることをテストする
// Property 7: Z順序の管理
// **Validates: Requirements 6.2, 6.3, 6.4**
func TestProperty7_ZOrder_MixedLayerTypesPreserveOrder(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: Picture、Text、Castを任意の順序で追加しても、
	// 追加順序に基づくZ順序が保持される
	property := func(winID int, layerTypes []uint8) bool {
		// 負のIDは正に変換
		if winID < 0 {
			winID = -winID
		}

		// レイヤー数を制限（1-20）
		if len(layerTypes) == 0 {
			return true
		}
		if len(layerTypes) > 20 {
			layerTypes = layerTypes[:20]
		}

		lm := NewLayerManager()
		bgColor := color.RGBA{0, 0, 0, 255}

		wls := lm.GetOrCreateWindowLayerSet(winID, 640, 480, bgColor)
		if wls == nil {
			return false
		}

		// レイヤーを追加順序で記録
		addedLayers := make([]Layer, 0)

		for _, lt := range layerTypes {
			layerType := lt % 3

			var layer Layer
			switch layerType {
			case 0:
				pl := NewPictureLayer(lm.GetNextLayerID(), 640, 480)
				layer = pl
			case 1:
				tl := NewTextLayerEntry(lm.GetNextLayerID(), 1, 0, 0, "test", 0)
				layer = tl
			case 2:
				ml := &mockLayer{}
				ml.SetID(lm.GetNextLayerID())
				ml.SetBounds(image.Rect(0, 0, 32, 32))
				ml.layerType = LayerTypeCast
				layer = ml
			}

			wls.AddLayer(layer)
			addedLayers = append(addedLayers, layer)
		}

		// GetLayersSortedを呼び出し
		sorted := wls.GetLayersSorted()

		// レイヤー数が正しいことを確認
		if len(sorted) != len(addedLayers) {
			return false
		}

		// ソートされたレイヤーが追加順序と一致することを確認
		// （Z順序は追加順序に基づいて割り当てられるため）
		for i, layer := range sorted {
			if layer.GetID() != addedLayers[i].GetID() {
				return false
			}
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 7 (Mixed Layer Types Preserve Order) failed: %v", err)
	}
}

// TestProperty7_ZOrder_UniqueZOrdersForAllLayers はすべてのレイヤーが一意のZ順序を持つことをテストする
// Property 7: Z順序の管理
// **Validates: Requirements 6.3**
func TestProperty7_ZOrder_UniqueZOrdersForAllLayers(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: 任意のレイヤー追加操作に対して、すべてのレイヤーは一意のZ順序を持つ
	property := func(winID int, numLayers uint8) bool {
		// 負のIDは正に変換
		if winID < 0 {
			winID = -winID
		}

		// レイヤー数を制限（1-50）
		layerCount := int(numLayers%50) + 1

		wls := NewWindowLayerSet(winID, 640, 480, color.RGBA{0, 0, 0, 255})
		if wls == nil {
			return false
		}

		// レイヤーを追加
		for i := 0; i < layerCount; i++ {
			layer := &mockLayer{}
			layer.SetID(i + 1)
			layer.SetBounds(image.Rect(0, 0, 100, 100))
			wls.AddLayer(layer)
		}

		// すべてのレイヤーのZ順序を収集
		layers := wls.GetLayers()
		zOrders := make(map[int]bool)

		for _, layer := range layers {
			z := layer.GetZOrder()
			if zOrders[z] {
				// 重複するZ順序が見つかった
				return false
			}
			zOrders[z] = true
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 7 (Unique Z-Orders) failed: %v", err)
	}
}

// TestProperty7_ZOrder_IntegratedBehavior はZ順序管理の統合動作をテストする
// Property 7: Z順序の管理
// **Validates: Requirements 6.2, 6.3, 6.4**
func TestProperty7_ZOrder_IntegratedBehavior(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: Z順序管理の全体的な動作が正しいことを確認
	// - 新しいレイヤーには現在のZ順序カウンターが割り当てられる
	// - カウンターは操作ごとに増加する
	// - すべてのレイヤータイプで共通のカウンターが使用される
	// - レイヤーはZ順序（小さい順）でソートされる
	property := func(winID int, operations []uint8) bool {
		// 負のIDは正に変換
		if winID < 0 {
			winID = -winID
		}

		// 操作数を制限（1-30）
		if len(operations) == 0 {
			return true
		}
		if len(operations) > 30 {
			operations = operations[:30]
		}

		lm := NewLayerManager()
		bgColor := color.RGBA{0, 0, 0, 255}

		wls := lm.GetOrCreateWindowLayerSet(winID, 640, 480, bgColor)
		if wls == nil {
			return false
		}

		// 追加されたレイヤーを記録
		addedLayers := make([]Layer, 0)
		expectedZOrders := make([]int, 0)

		for _, op := range operations {
			layerType := op % 3

			// 追加前のZ順序カウンターを記録
			expectedZ := wls.GetNextZOrder()

			var layer Layer
			switch layerType {
			case 0:
				pl := NewPictureLayer(lm.GetNextLayerID(), 640, 480)
				layer = pl
			case 1:
				tl := NewTextLayerEntry(lm.GetNextLayerID(), 1, 0, 0, "test", 0)
				layer = tl
			case 2:
				ml := &mockLayer{}
				ml.SetID(lm.GetNextLayerID())
				ml.SetBounds(image.Rect(0, 0, 32, 32))
				ml.layerType = LayerTypeCast
				layer = ml
			}

			wls.AddLayer(layer)
			addedLayers = append(addedLayers, layer)
			expectedZOrders = append(expectedZOrders, expectedZ)

			// 要件 6.3: 新しいレイヤーには現在のZ順序カウンターが割り当てられる
			if layer.GetZOrder() != expectedZ {
				return false
			}
		}

		// 要件 6.2: レイヤースタックをZ順序（小さい順）で描画する
		sorted := wls.GetLayersSorted()
		for i := 1; i < len(sorted); i++ {
			if sorted[i-1].GetZOrder() >= sorted[i].GetZOrder() {
				return false
			}
		}

		// 要件 6.4: すべてのレイヤータイプで共通のカウンターが使用される
		// Z順序が連続していることを確認
		for i := 1; i < len(expectedZOrders); i++ {
			if expectedZOrders[i] != expectedZOrders[i-1]+1 {
				return false
			}
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 7 (Integrated Behavior) failed: %v", err)
	}
}

// ============================================================
// Property 9: ダーティフラグの動作
// *任意の*レイヤー変更操作に対して:
// - 位置変更時にダーティフラグが設定される
// - 内容変更時にダーティフラグが設定される
// - 合成処理完了後にダーティフラグがクリアされる
// **Validates: Requirements 9.1, 9.2**
// ============================================================

// TestProperty9_DirtyFlag_PositionChangeSetsDirty は位置変更時にダーティフラグが設定されることをテストする
// Property 9: ダーティフラグの動作
// **Validates: Requirements 9.1, 9.2**
func TestProperty9_DirtyFlag_PositionChangeSetsDirty(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: 任意のレイヤーの位置を変更すると、ダーティフラグが設定される
	property := func(winID int, x1, y1, x2, y2 int16, w, h uint16) bool {
		// 負のIDは正に変換
		if winID < 0 {
			winID = -winID
		}

		// サイズは最小1以上
		width := int(w%100) + 1
		height := int(h%100) + 1

		wls := NewWindowLayerSet(winID, 640, 480, color.RGBA{0, 0, 0, 255})
		if wls == nil {
			return false
		}

		// レイヤーを作成して追加
		layer := &mockLayer{}
		layer.SetID(1)
		layer.SetBounds(image.Rect(int(x1), int(y1), int(x1)+width, int(y1)+height))
		layer.SetVisible(true)
		wls.AddLayer(layer)

		// ダーティフラグをクリア
		wls.ClearDirty()
		layer.SetDirty(false)

		// 位置が同じ場合はスキップ
		if x1 == x2 && y1 == y2 {
			return true
		}

		// 位置を変更
		newBounds := image.Rect(int(x2), int(y2), int(x2)+width, int(y2)+height)
		layer.SetBounds(newBounds)

		// レイヤーのダーティフラグが設定されていることを確認
		if !layer.IsDirty() {
			return false
		}

		// WindowLayerSetがダーティであることを確認
		if !wls.IsDirty() {
			return false
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 9 (Position Change Sets Dirty) failed: %v", err)
	}
}

// TestProperty9_DirtyFlag_ContentChangeSetsDirty は内容変更時にダーティフラグが設定されることをテストする
// Property 9: ダーティフラグの動作
// **Validates: Requirements 9.1, 9.2**
func TestProperty9_DirtyFlag_ContentChangeSetsDirty(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: 任意のレイヤーの内容を変更（Invalidate）すると、ダーティフラグが設定される
	property := func(winID int, numLayers uint8) bool {
		// 負のIDは正に変換
		if winID < 0 {
			winID = -winID
		}

		// レイヤー数を制限（1-10）
		layerCount := int(numLayers%10) + 1

		wls := NewWindowLayerSet(winID, 640, 480, color.RGBA{0, 0, 0, 255})
		if wls == nil {
			return false
		}

		// レイヤーを作成して追加
		layers := make([]*mockLayer, layerCount)
		for i := 0; i < layerCount; i++ {
			layer := &mockLayer{}
			layer.SetID(i + 1)
			layer.SetBounds(image.Rect(i*10, i*10, i*10+50, i*10+50))
			layer.SetVisible(true)
			layers[i] = layer
			wls.AddLayer(layer)
		}

		// ダーティフラグをクリア
		wls.ClearDirty()
		for _, layer := range layers {
			layer.SetDirty(false)
		}

		// 各レイヤーの内容を変更（Invalidate）
		for i, layer := range layers {
			layer.Invalidate()

			// レイヤーのダーティフラグが設定されていることを確認
			if !layer.IsDirty() {
				t.Logf("Layer %d dirty flag not set after Invalidate", i)
				return false
			}

			// WindowLayerSetがダーティであることを確認
			if !wls.IsDirty() {
				t.Logf("WindowLayerSet not dirty after layer %d Invalidate", i)
				return false
			}
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 9 (Content Change Sets Dirty) failed: %v", err)
	}
}

// TestProperty9_DirtyFlag_ClearDirtyClearsAllFlags は合成処理完了後にダーティフラグがクリアされることをテストする
// Property 9: ダーティフラグの動作
// **Validates: Requirements 9.1, 9.2**
func TestProperty9_DirtyFlag_ClearDirtyClearsAllFlags(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: ClearDirty()を呼び出すと、すべてのダーティフラグがクリアされる
	property := func(winID int, numLayers uint8) bool {
		// 負のIDは正に変換
		if winID < 0 {
			winID = -winID
		}

		// レイヤー数を制限（1-10）
		layerCount := int(numLayers%10) + 1

		wls := NewWindowLayerSet(winID, 640, 480, color.RGBA{0, 0, 0, 255})
		if wls == nil {
			return false
		}

		// レイヤーを作成して追加
		layers := make([]*mockLayer, layerCount)
		for i := 0; i < layerCount; i++ {
			layer := &mockLayer{}
			layer.SetID(i + 1)
			layer.SetBounds(image.Rect(i*10, i*10, i*10+50, i*10+50))
			layer.SetVisible(true)
			layer.SetDirty(true) // ダーティに設定
			layers[i] = layer
			wls.AddLayer(layer)
		}

		// WindowLayerSetがダーティであることを確認
		if !wls.IsDirty() {
			return false
		}

		// ClearDirty()を呼び出す（合成処理完了をシミュレート）
		wls.ClearDirty()

		// WindowLayerSetのダーティフラグがクリアされていることを確認
		if wls.IsFullDirty() {
			return false
		}

		// ダーティ領域が空であることを確認
		if !wls.GetDirtyRegion().Empty() {
			return false
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 9 (ClearDirty Clears All Flags) failed: %v", err)
	}
}

// TestProperty9_DirtyFlag_ClearAllDirtyFlagsClearsLayerFlags はClearAllDirtyFlagsがすべてのレイヤーのダーティフラグをクリアすることをテストする
// Property 9: ダーティフラグの動作
// **Validates: Requirements 9.1, 9.2**
func TestProperty9_DirtyFlag_ClearAllDirtyFlagsClearsLayerFlags(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: ClearAllDirtyFlags()を呼び出すと、すべてのレイヤーのダーティフラグがクリアされる
	property := func(winID int, numLayers uint8) bool {
		// 負のIDは正に変換
		if winID < 0 {
			winID = -winID
		}

		// レイヤー数を制限（1-10）
		layerCount := int(numLayers%10) + 1

		wls := NewWindowLayerSet(winID, 640, 480, color.RGBA{0, 0, 0, 255})
		if wls == nil {
			return false
		}

		// レイヤーを作成して追加
		layers := make([]*mockLayer, layerCount)
		for i := 0; i < layerCount; i++ {
			layer := &mockLayer{}
			layer.SetID(i + 1)
			layer.SetBounds(image.Rect(i*10, i*10, i*10+50, i*10+50))
			layer.SetVisible(true)
			layer.SetDirty(true) // ダーティに設定
			layers[i] = layer
			wls.AddLayer(layer)
		}

		// すべてのレイヤーがダーティであることを確認
		for _, layer := range layers {
			if !layer.IsDirty() {
				return false
			}
		}

		// ClearAllDirtyFlags()を呼び出す
		wls.ClearAllDirtyFlags()

		// すべてのレイヤーのダーティフラグがクリアされていることを確認
		for _, layer := range wls.GetLayers() {
			if layer.IsDirty() {
				return false
			}
		}

		// WindowLayerSetのダーティフラグもクリアされていることを確認
		if wls.IsFullDirty() {
			return false
		}

		if !wls.GetDirtyRegion().Empty() {
			return false
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 9 (ClearAllDirtyFlags Clears Layer Flags) failed: %v", err)
	}
}

// TestProperty9_DirtyFlag_AddLayerSetsDirty はレイヤー追加時にダーティフラグが設定されることをテストする
// Property 9: ダーティフラグの動作
// **Validates: Requirements 9.1, 9.2**
func TestProperty9_DirtyFlag_AddLayerSetsDirty(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: 任意のレイヤーを追加すると、WindowLayerSetがダーティになる
	property := func(winID int, numLayers uint8) bool {
		// 負のIDは正に変換
		if winID < 0 {
			winID = -winID
		}

		// レイヤー数を制限（1-10）
		layerCount := int(numLayers%10) + 1

		wls := NewWindowLayerSet(winID, 640, 480, color.RGBA{0, 0, 0, 255})
		if wls == nil {
			return false
		}

		// 初期状態でダーティフラグをクリア
		wls.ClearDirty()

		// 各レイヤーを追加するたびにダーティフラグが設定されることを確認
		for i := 0; i < layerCount; i++ {
			// ダーティフラグをクリア
			wls.ClearDirty()

			// ダーティでないことを確認
			if wls.IsFullDirty() {
				return false
			}

			// レイヤーを追加
			layer := &mockLayer{}
			layer.SetID(i + 1)
			layer.SetBounds(image.Rect(i*10, i*10, i*10+50, i*10+50))
			layer.SetVisible(true)
			wls.AddLayer(layer)

			// ダーティフラグが設定されていることを確認
			if !wls.IsFullDirty() {
				return false
			}
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 9 (Add Layer Sets Dirty) failed: %v", err)
	}
}

// TestProperty9_DirtyFlag_RemoveLayerSetsDirty はレイヤー削除時にダーティフラグが設定されることをテストする
// Property 9: ダーティフラグの動作
// **Validates: Requirements 9.1, 9.2**
func TestProperty9_DirtyFlag_RemoveLayerSetsDirty(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: 任意のレイヤーを削除すると、WindowLayerSetがダーティになる
	property := func(winID int, numLayers uint8) bool {
		// 負のIDは正に変換
		if winID < 0 {
			winID = -winID
		}

		// レイヤー数を制限（2-10）
		layerCount := int(numLayers%9) + 2

		wls := NewWindowLayerSet(winID, 640, 480, color.RGBA{0, 0, 0, 255})
		if wls == nil {
			return false
		}

		// レイヤーを追加
		for i := 0; i < layerCount; i++ {
			layer := &mockLayer{}
			layer.SetID(i + 1)
			layer.SetBounds(image.Rect(i*10, i*10, i*10+50, i*10+50))
			layer.SetVisible(true)
			wls.AddLayer(layer)
		}

		// ダーティフラグをクリア
		wls.ClearDirty()

		// 各レイヤーを削除するたびにダーティフラグが設定されることを確認
		for i := 0; i < layerCount; i++ {
			// ダーティフラグをクリア
			wls.ClearDirty()

			// ダーティでないことを確認
			if wls.IsFullDirty() {
				return false
			}

			// レイヤーを削除
			removed := wls.RemoveLayer(i + 1)
			if !removed {
				// レイヤーが見つからない場合はスキップ
				continue
			}

			// ダーティフラグが設定されていることを確認
			if !wls.IsFullDirty() {
				return false
			}
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 9 (Remove Layer Sets Dirty) failed: %v", err)
	}
}

// TestProperty9_DirtyFlag_VisibilityChangeSetsDirty は可視性変更時にダーティフラグが設定されることをテストする
// Property 9: ダーティフラグの動作
// **Validates: Requirements 9.1, 9.2**
func TestProperty9_DirtyFlag_VisibilityChangeSetsDirty(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: 任意のレイヤーの可視性を変更すると、ダーティフラグが設定される
	property := func(winID int, initialVisible bool) bool {
		// 負のIDは正に変換
		if winID < 0 {
			winID = -winID
		}

		wls := NewWindowLayerSet(winID, 640, 480, color.RGBA{0, 0, 0, 255})
		if wls == nil {
			return false
		}

		// レイヤーを作成して追加
		layer := &mockLayer{}
		layer.SetID(1)
		layer.SetBounds(image.Rect(0, 0, 100, 100))
		layer.SetVisible(initialVisible)
		wls.AddLayer(layer)

		// ダーティフラグをクリア
		wls.ClearDirty()
		layer.SetDirty(false)

		// 可視性を反転
		layer.SetVisible(!initialVisible)

		// レイヤーのダーティフラグが設定されていることを確認
		if !layer.IsDirty() {
			return false
		}

		// WindowLayerSetがダーティであることを確認
		if !wls.IsDirty() {
			return false
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 9 (Visibility Change Sets Dirty) failed: %v", err)
	}
}

// TestProperty9_DirtyFlag_DirtyRegionTracking はダーティ領域の追跡が正しく動作することをテストする
// Property 9: ダーティフラグの動作
// **Validates: Requirements 9.1, 9.2**
func TestProperty9_DirtyFlag_DirtyRegionTracking(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: ダーティ領域を追加すると、WindowLayerSetがダーティになり、
	// ClearDirty()でダーティ領域がクリアされる
	property := func(winID int, x, y int16, w, h uint16) bool {
		// 負のIDは正に変換
		if winID < 0 {
			winID = -winID
		}

		// サイズは最小1以上
		width := int(w%100) + 1
		height := int(h%100) + 1

		wls := NewWindowLayerSet(winID, 640, 480, color.RGBA{0, 0, 0, 255})
		if wls == nil {
			return false
		}

		// ダーティフラグをクリア
		wls.ClearDirty()

		// ダーティ領域を追加
		rect := image.Rect(int(x), int(y), int(x)+width, int(y)+height)
		wls.AddDirtyRegion(rect)

		// ダーティ領域が設定されていることを確認
		if wls.GetDirtyRegion().Empty() {
			return false
		}

		// WindowLayerSetがダーティであることを確認
		if !wls.IsDirty() {
			return false
		}

		// ClearDirty()を呼び出す
		wls.ClearDirty()

		// ダーティ領域がクリアされていることを確認
		if !wls.GetDirtyRegion().Empty() {
			return false
		}

		// WindowLayerSetがダーティでないことを確認
		if wls.IsDirty() {
			return false
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 9 (Dirty Region Tracking) failed: %v", err)
	}
}

// TestProperty9_DirtyFlag_MultipleDirtyRegionsUnion は複数のダーティ領域が統合されることをテストする
// Property 9: ダーティフラグの動作
// **Validates: Requirements 9.1, 9.2**
func TestProperty9_DirtyFlag_MultipleDirtyRegionsUnion(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: 複数のダーティ領域を追加すると、それらが統合される
	property := func(winID int, x1, y1, x2, y2 int16, w1, h1, w2, h2 uint16) bool {
		// 負のIDは正に変換
		if winID < 0 {
			winID = -winID
		}

		// サイズは最小1以上
		width1 := int(w1%100) + 1
		height1 := int(h1%100) + 1
		width2 := int(w2%100) + 1
		height2 := int(h2%100) + 1

		wls := NewWindowLayerSet(winID, 640, 480, color.RGBA{0, 0, 0, 255})
		if wls == nil {
			return false
		}

		// ダーティフラグをクリア
		wls.ClearDirty()

		// 2つのダーティ領域を追加
		rect1 := image.Rect(int(x1), int(y1), int(x1)+width1, int(y1)+height1)
		rect2 := image.Rect(int(x2), int(y2), int(x2)+width2, int(y2)+height2)

		wls.AddDirtyRegion(rect1)
		wls.AddDirtyRegion(rect2)

		// ダーティ領域が統合されていることを確認
		expectedUnion := rect1.Union(rect2)
		actualRegion := wls.GetDirtyRegion()

		if actualRegion != expectedUnion {
			return false
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 9 (Multiple Dirty Regions Union) failed: %v", err)
	}
}

// TestProperty9_DirtyFlag_IntegratedBehavior はダーティフラグの統合動作をテストする
// Property 9: ダーティフラグの動作
// **Validates: Requirements 9.1, 9.2**
func TestProperty9_DirtyFlag_IntegratedBehavior(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: ダーティフラグの全体的な動作が正しいことを確認
	// - 位置変更時にダーティフラグが設定される
	// - 内容変更時にダーティフラグが設定される
	// - 合成処理完了後にダーティフラグがクリアされる
	property := func(winID int, operations []uint8) bool {
		// 負のIDは正に変換
		if winID < 0 {
			winID = -winID
		}

		// 操作数を制限（1-20）
		if len(operations) == 0 {
			return true
		}
		if len(operations) > 20 {
			operations = operations[:20]
		}

		wls := NewWindowLayerSet(winID, 640, 480, color.RGBA{0, 0, 0, 255})
		if wls == nil {
			return false
		}

		// レイヤーを作成して追加
		layer := &mockLayer{}
		layer.SetID(1)
		layer.SetBounds(image.Rect(0, 0, 100, 100))
		layer.SetVisible(true)
		wls.AddLayer(layer)

		// ダーティフラグをクリア
		wls.ClearDirty()
		layer.SetDirty(false)

		for _, op := range operations {
			opType := op % 4

			switch opType {
			case 0:
				// 位置変更
				newX := int(op % 50)
				newY := int((op * 2) % 50)
				oldBounds := layer.GetBounds()
				newBounds := image.Rect(newX, newY, newX+oldBounds.Dx(), newY+oldBounds.Dy())

				if oldBounds != newBounds {
					layer.SetBounds(newBounds)

					// 要件 9.1: 位置変更時にダーティフラグが設定される
					if !layer.IsDirty() {
						return false
					}
					if !wls.IsDirty() {
						return false
					}
				}

			case 1:
				// 内容変更（Invalidate）
				layer.Invalidate()

				// 要件 9.1: 内容変更時にダーティフラグが設定される
				if !layer.IsDirty() {
					return false
				}
				if !wls.IsDirty() {
					return false
				}

			case 2:
				// 可視性変更
				currentVisible := layer.IsVisible()
				layer.SetVisible(!currentVisible)

				// 要件 9.1: 可視性変更時にダーティフラグが設定される
				if !layer.IsDirty() {
					return false
				}
				if !wls.IsDirty() {
					return false
				}

			case 3:
				// 合成処理完了（ClearDirty）
				wls.ClearDirty()
				layer.SetDirty(false)

				// 要件 9.2: 合成処理完了後にダーティフラグがクリアされる
				if wls.IsFullDirty() {
					return false
				}
				if !wls.GetDirtyRegion().Empty() {
					return false
				}
			}
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 9 (Integrated Behavior) failed: %v", err)
	}
}

// TestProperty9_DirtyFlag_BgColorChangeSetsDirty は背景色変更時にダーティフラグが設定されることをテストする
// Property 9: ダーティフラグの動作
// **Validates: Requirements 9.1, 9.2**
func TestProperty9_DirtyFlag_BgColorChangeSetsDirty(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: 背景色を変更すると、WindowLayerSetがダーティになる
	property := func(winID int, r1, g1, b1, r2, g2, b2 uint8) bool {
		// 負のIDは正に変換
		if winID < 0 {
			winID = -winID
		}

		bgColor1 := color.RGBA{r1, g1, b1, 255}
		bgColor2 := color.RGBA{r2, g2, b2, 255}

		wls := NewWindowLayerSet(winID, 640, 480, bgColor1)
		if wls == nil {
			return false
		}

		// ダーティフラグをクリア
		wls.ClearDirty()

		// 背景色を変更
		wls.SetBgColor(bgColor2)

		// ダーティフラグが設定されていることを確認
		if !wls.IsFullDirty() {
			return false
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 9 (BgColor Change Sets Dirty) failed: %v", err)
	}
}

// TestProperty9_DirtyFlag_SizeChangeSetsDirty はサイズ変更時にダーティフラグが設定されることをテストする
// Property 9: ダーティフラグの動作
// **Validates: Requirements 9.1, 9.2**
func TestProperty9_DirtyFlag_SizeChangeSetsDirty(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: サイズを変更すると、WindowLayerSetがダーティになる
	property := func(winID int, w1, h1, w2, h2 uint16) bool {
		// 負のIDは正に変換
		if winID < 0 {
			winID = -winID
		}

		// サイズは最小1以上
		width1 := int(w1%1000) + 1
		height1 := int(h1%1000) + 1
		width2 := int(w2%1000) + 1
		height2 := int(h2%1000) + 1

		wls := NewWindowLayerSet(winID, width1, height1, color.RGBA{0, 0, 0, 255})
		if wls == nil {
			return false
		}

		// ダーティフラグをクリア
		wls.ClearDirty()

		// サイズが同じ場合はスキップ
		if width1 == width2 && height1 == height2 {
			return true
		}

		// サイズを変更
		wls.SetSize(width2, height2)

		// ダーティフラグが設定されていることを確認
		if !wls.IsFullDirty() {
			return false
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 9 (Size Change Sets Dirty) failed: %v", err)
	}
}
