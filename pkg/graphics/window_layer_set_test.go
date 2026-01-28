package graphics

import (
	"image"
	"image/color"
	"testing"
)

// TestNewWindowLayerSet はWindowLayerSetの作成をテストする
func TestNewWindowLayerSet(t *testing.T) {
	winID := 1
	width := 640
	height := 480
	bgColor := color.RGBA{0, 0, 128, 255}

	wls := NewWindowLayerSet(winID, width, height, bgColor)

	if wls == nil {
		t.Fatal("NewWindowLayerSet returned nil")
	}

	if wls.WinID != winID {
		t.Errorf("WinID = %d, want %d", wls.WinID, winID)
	}

	if wls.Width != width {
		t.Errorf("Width = %d, want %d", wls.Width, width)
	}

	if wls.Height != height {
		t.Errorf("Height = %d, want %d", wls.Height, height)
	}

	if wls.BgColor != bgColor {
		t.Errorf("BgColor = %v, want %v", wls.BgColor, bgColor)
	}

	if len(wls.Layers) != 0 {
		t.Errorf("Layers length = %d, want 0", len(wls.Layers))
	}

	if wls.nextZOrder != 1 {
		t.Errorf("nextZOrder = %d, want 1", wls.nextZOrder)
	}

	if !wls.FullDirty {
		t.Error("FullDirty should be true initially")
	}
}

// TestWindowLayerSetGetters はゲッターメソッドをテストする
func TestWindowLayerSetGetters(t *testing.T) {
	winID := 2
	width := 800
	height := 600
	bgColor := color.RGBA{255, 255, 255, 255}

	wls := NewWindowLayerSet(winID, width, height, bgColor)

	if got := wls.GetWinID(); got != winID {
		t.Errorf("GetWinID() = %d, want %d", got, winID)
	}

	if got := wls.GetBgColor(); got != bgColor {
		t.Errorf("GetBgColor() = %v, want %v", got, bgColor)
	}

	gotWidth, gotHeight := wls.GetSize()
	if gotWidth != width || gotHeight != height {
		t.Errorf("GetSize() = (%d, %d), want (%d, %d)", gotWidth, gotHeight, width, height)
	}

	if got := wls.GetNextZOrder(); got != 1 {
		t.Errorf("GetNextZOrder() = %d, want 1", got)
	}

	if got := wls.GetLayerCount(); got != 0 {
		t.Errorf("GetLayerCount() = %d, want 0", got)
	}
}

// TestWindowLayerSetSetBgColor は背景色の設定をテストする
func TestWindowLayerSetSetBgColor(t *testing.T) {
	wls := NewWindowLayerSet(1, 640, 480, color.RGBA{0, 0, 0, 255})
	wls.FullDirty = false // リセット

	newColor := color.RGBA{128, 128, 128, 255}
	wls.SetBgColor(newColor)

	if wls.BgColor != newColor {
		t.Errorf("BgColor = %v, want %v", wls.BgColor, newColor)
	}

	if !wls.FullDirty {
		t.Error("FullDirty should be true after SetBgColor")
	}
}

// TestWindowLayerSetSetSize はサイズの設定をテストする
func TestWindowLayerSetSetSize(t *testing.T) {
	wls := NewWindowLayerSet(1, 640, 480, color.RGBA{0, 0, 0, 255})
	wls.FullDirty = false // リセット

	wls.SetSize(800, 600)

	width, height := wls.GetSize()
	if width != 800 || height != 600 {
		t.Errorf("GetSize() = (%d, %d), want (800, 600)", width, height)
	}

	if !wls.FullDirty {
		t.Error("FullDirty should be true after SetSize")
	}
}

// TestWindowLayerSetSetSizeSameValue は同じサイズを設定した場合をテストする
func TestWindowLayerSetSetSizeSameValue(t *testing.T) {
	wls := NewWindowLayerSet(1, 640, 480, color.RGBA{0, 0, 0, 255})
	wls.FullDirty = false // リセット

	wls.SetSize(640, 480) // 同じサイズ

	if wls.FullDirty {
		t.Error("FullDirty should remain false when size doesn't change")
	}
}

// newTestLayer はテスト用のモックレイヤーを作成する
// 既存のmockLayerを使用（layer_test.goで定義）
func newTestLayer(id int) *mockLayer {
	layer := &mockLayer{}
	layer.SetID(id)
	layer.SetBounds(image.Rect(0, 0, 100, 100))
	layer.SetVisible(true)
	return layer
}

// TestWindowLayerSetAddLayer はレイヤーの追加をテストする
func TestWindowLayerSetAddLayer(t *testing.T) {
	wls := NewWindowLayerSet(1, 640, 480, color.RGBA{0, 0, 0, 255})
	wls.FullDirty = false // リセット

	layer1 := newTestLayer(1)
	wls.AddLayer(layer1)

	if wls.GetLayerCount() != 1 {
		t.Errorf("GetLayerCount() = %d, want 1", wls.GetLayerCount())
	}

	if layer1.GetZOrder() != 1 {
		t.Errorf("layer1.GetZOrder() = %d, want 1", layer1.GetZOrder())
	}

	if wls.GetNextZOrder() != 2 {
		t.Errorf("GetNextZOrder() = %d, want 2", wls.GetNextZOrder())
	}

	if !wls.FullDirty {
		t.Error("FullDirty should be true after AddLayer")
	}

	// 2つ目のレイヤーを追加
	layer2 := newTestLayer(2)
	wls.AddLayer(layer2)

	if wls.GetLayerCount() != 2 {
		t.Errorf("GetLayerCount() = %d, want 2", wls.GetLayerCount())
	}

	if layer2.GetZOrder() != 2 {
		t.Errorf("layer2.GetZOrder() = %d, want 2", layer2.GetZOrder())
	}

	if wls.GetNextZOrder() != 3 {
		t.Errorf("GetNextZOrder() = %d, want 3", wls.GetNextZOrder())
	}
}

// TestWindowLayerSetAddLayerNil はnilレイヤーの追加をテストする
func TestWindowLayerSetAddLayerNil(t *testing.T) {
	wls := NewWindowLayerSet(1, 640, 480, color.RGBA{0, 0, 0, 255})
	initialCount := wls.GetLayerCount()

	wls.AddLayer(nil)

	if wls.GetLayerCount() != initialCount {
		t.Errorf("GetLayerCount() = %d, want %d (nil should not be added)", wls.GetLayerCount(), initialCount)
	}
}

// TestWindowLayerSetRemoveLayer はレイヤーの削除をテストする
func TestWindowLayerSetRemoveLayer(t *testing.T) {
	wls := NewWindowLayerSet(1, 640, 480, color.RGBA{0, 0, 0, 255})

	layer1 := newTestLayer(1)
	layer2 := newTestLayer(2)
	wls.AddLayer(layer1)
	wls.AddLayer(layer2)
	wls.FullDirty = false // リセット

	// layer1を削除
	removed := wls.RemoveLayer(1)

	if !removed {
		t.Error("RemoveLayer should return true")
	}

	if wls.GetLayerCount() != 1 {
		t.Errorf("GetLayerCount() = %d, want 1", wls.GetLayerCount())
	}

	if wls.GetLayer(1) != nil {
		t.Error("GetLayer(1) should return nil after removal")
	}

	if wls.GetLayer(2) == nil {
		t.Error("GetLayer(2) should not be nil")
	}

	if !wls.FullDirty {
		t.Error("FullDirty should be true after RemoveLayer")
	}
}

// TestWindowLayerSetRemoveLayerNotFound は存在しないレイヤーの削除をテストする
func TestWindowLayerSetRemoveLayerNotFound(t *testing.T) {
	wls := NewWindowLayerSet(1, 640, 480, color.RGBA{0, 0, 0, 255})

	layer1 := newTestLayer(1)
	wls.AddLayer(layer1)

	removed := wls.RemoveLayer(999) // 存在しないID

	if removed {
		t.Error("RemoveLayer should return false for non-existent layer")
	}

	if wls.GetLayerCount() != 1 {
		t.Errorf("GetLayerCount() = %d, want 1", wls.GetLayerCount())
	}
}

// TestWindowLayerSetGetLayer はレイヤーの取得をテストする
func TestWindowLayerSetGetLayer(t *testing.T) {
	wls := NewWindowLayerSet(1, 640, 480, color.RGBA{0, 0, 0, 255})

	layer1 := newTestLayer(1)
	layer2 := newTestLayer(2)
	wls.AddLayer(layer1)
	wls.AddLayer(layer2)

	got := wls.GetLayer(1)
	if got == nil {
		t.Error("GetLayer(1) should not return nil")
	}
	if got.GetID() != 1 {
		t.Errorf("GetLayer(1).GetID() = %d, want 1", got.GetID())
	}

	got = wls.GetLayer(2)
	if got == nil {
		t.Error("GetLayer(2) should not return nil")
	}
	if got.GetID() != 2 {
		t.Errorf("GetLayer(2).GetID() = %d, want 2", got.GetID())
	}

	got = wls.GetLayer(999)
	if got != nil {
		t.Error("GetLayer(999) should return nil")
	}
}

// TestWindowLayerSetGetLayersSorted はZ順序でソートされたレイヤーの取得をテストする
func TestWindowLayerSetGetLayersSorted(t *testing.T) {
	wls := NewWindowLayerSet(1, 640, 480, color.RGBA{0, 0, 0, 255})

	// 順番に追加
	layer1 := newTestLayer(1)
	layer2 := newTestLayer(2)
	layer3 := newTestLayer(3)
	wls.AddLayer(layer1)
	wls.AddLayer(layer2)
	wls.AddLayer(layer3)

	sorted := wls.GetLayersSorted()

	if len(sorted) != 3 {
		t.Fatalf("GetLayersSorted() length = %d, want 3", len(sorted))
	}

	// Z順序が昇順であることを確認
	for i := 1; i < len(sorted); i++ {
		if sorted[i-1].GetZOrder() > sorted[i].GetZOrder() {
			t.Errorf("Layers not sorted: Z[%d]=%d > Z[%d]=%d",
				i-1, sorted[i-1].GetZOrder(), i, sorted[i].GetZOrder())
		}
	}
}

// TestWindowLayerSetClearLayers はレイヤーのクリアをテストする
func TestWindowLayerSetClearLayers(t *testing.T) {
	wls := NewWindowLayerSet(1, 640, 480, color.RGBA{0, 0, 0, 255})

	layer1 := newTestLayer(1)
	layer2 := newTestLayer(2)
	wls.AddLayer(layer1)
	wls.AddLayer(layer2)
	wls.FullDirty = false // リセット

	wls.ClearLayers()

	if wls.GetLayerCount() != 0 {
		t.Errorf("GetLayerCount() = %d, want 0", wls.GetLayerCount())
	}

	if wls.GetNextZOrder() != 1 {
		t.Errorf("GetNextZOrder() = %d, want 1", wls.GetNextZOrder())
	}

	if !wls.FullDirty {
		t.Error("FullDirty should be true after ClearLayers")
	}
}

// TestWindowLayerSetDirtyRegion はダーティ領域の管理をテストする
func TestWindowLayerSetDirtyRegion(t *testing.T) {
	wls := NewWindowLayerSet(1, 640, 480, color.RGBA{0, 0, 0, 255})
	wls.FullDirty = false
	wls.DirtyRegion = image.Rectangle{}

	// 最初のダーティ領域を追加
	rect1 := image.Rect(10, 10, 50, 50)
	wls.AddDirtyRegion(rect1)

	if wls.DirtyRegion != rect1 {
		t.Errorf("DirtyRegion = %v, want %v", wls.DirtyRegion, rect1)
	}

	// 2つ目のダーティ領域を追加（統合される）
	rect2 := image.Rect(40, 40, 100, 100)
	wls.AddDirtyRegion(rect2)

	expected := rect1.Union(rect2)
	if wls.DirtyRegion != expected {
		t.Errorf("DirtyRegion = %v, want %v", wls.DirtyRegion, expected)
	}

	// ダーティ領域をクリア
	wls.ClearDirtyRegion()

	if !wls.DirtyRegion.Empty() {
		t.Error("DirtyRegion should be empty after ClearDirtyRegion")
	}

	if wls.FullDirty {
		t.Error("FullDirty should be false after ClearDirtyRegion")
	}
}

// TestWindowLayerSetAddDirtyRegionEmpty は空のダーティ領域の追加をテストする
func TestWindowLayerSetAddDirtyRegionEmpty(t *testing.T) {
	wls := NewWindowLayerSet(1, 640, 480, color.RGBA{0, 0, 0, 255})
	wls.FullDirty = false
	wls.DirtyRegion = image.Rectangle{}

	// 空の領域を追加
	wls.AddDirtyRegion(image.Rectangle{})

	if !wls.DirtyRegion.Empty() {
		t.Error("DirtyRegion should remain empty when adding empty rect")
	}
}

// TestWindowLayerSetIsDirty はダーティ判定をテストする
func TestWindowLayerSetIsDirty(t *testing.T) {
	wls := NewWindowLayerSet(1, 640, 480, color.RGBA{0, 0, 0, 255})

	// 初期状態はダーティ
	if !wls.IsDirty() {
		t.Error("IsDirty() should return true initially")
	}

	// ダーティフラグをクリア
	wls.FullDirty = false
	wls.DirtyRegion = image.Rectangle{}

	if wls.IsDirty() {
		t.Error("IsDirty() should return false after clearing flags")
	}

	// ダーティ領域を追加
	wls.AddDirtyRegion(image.Rect(0, 0, 10, 10))

	if !wls.IsDirty() {
		t.Error("IsDirty() should return true after adding dirty region")
	}

	// ダーティ領域をクリアしてレイヤーをダーティにする
	wls.ClearDirtyRegion()
	layer := newTestLayer(1)
	layer.SetDirty(true)
	wls.AddLayer(layer)
	wls.FullDirty = false

	if !wls.IsDirty() {
		t.Error("IsDirty() should return true when a layer is dirty")
	}
}

// TestWindowLayerSetMarkFullDirty はMarkFullDirtyをテストする
func TestWindowLayerSetMarkFullDirty(t *testing.T) {
	wls := NewWindowLayerSet(1, 640, 480, color.RGBA{0, 0, 0, 255})
	wls.FullDirty = false

	wls.MarkFullDirty()

	if !wls.FullDirty {
		t.Error("FullDirty should be true after MarkFullDirty")
	}
}

// TestWindowLayerSetGetTopmostLayer は最上位レイヤーの取得をテストする
func TestWindowLayerSetGetTopmostLayer(t *testing.T) {
	wls := NewWindowLayerSet(1, 640, 480, color.RGBA{0, 0, 0, 255})

	// 空の場合
	if wls.GetTopmostLayer() != nil {
		t.Error("GetTopmostLayer() should return nil for empty layer set")
	}

	// レイヤーを追加
	layer1 := newTestLayer(1)
	layer2 := newTestLayer(2)
	layer3 := newTestLayer(3)
	wls.AddLayer(layer1) // Z=1
	wls.AddLayer(layer2) // Z=2
	wls.AddLayer(layer3) // Z=3

	topmost := wls.GetTopmostLayer()
	if topmost == nil {
		t.Fatal("GetTopmostLayer() should not return nil")
	}

	if topmost.GetID() != 3 {
		t.Errorf("GetTopmostLayer().GetID() = %d, want 3", topmost.GetID())
	}

	if topmost.GetZOrder() != 3 {
		t.Errorf("GetTopmostLayer().GetZOrder() = %d, want 3", topmost.GetZOrder())
	}
}

// TestWindowLayerSetClearAllDirtyFlags はすべてのダーティフラグのクリアをテストする
func TestWindowLayerSetClearAllDirtyFlags(t *testing.T) {
	wls := NewWindowLayerSet(1, 640, 480, color.RGBA{0, 0, 0, 255})

	layer1 := newTestLayer(1)
	layer1.SetDirty(true)
	layer2 := newTestLayer(2)
	layer2.SetDirty(true)
	wls.AddLayer(layer1)
	wls.AddLayer(layer2)
	wls.AddDirtyRegion(image.Rect(0, 0, 100, 100))

	wls.ClearAllDirtyFlags()

	if layer1.IsDirty() {
		t.Error("layer1 should not be dirty after ClearAllDirtyFlags")
	}

	if layer2.IsDirty() {
		t.Error("layer2 should not be dirty after ClearAllDirtyFlags")
	}

	if !wls.DirtyRegion.Empty() {
		t.Error("DirtyRegion should be empty after ClearAllDirtyFlags")
	}

	if wls.FullDirty {
		t.Error("FullDirty should be false after ClearAllDirtyFlags")
	}
}

// TestWindowLayerSetCompositeBuffer は合成バッファの管理をテストする
func TestWindowLayerSetCompositeBuffer(t *testing.T) {
	wls := NewWindowLayerSet(1, 640, 480, color.RGBA{0, 0, 0, 255})

	if wls.GetCompositeBuffer() != nil {
		t.Error("GetCompositeBuffer() should return nil initially")
	}

	// 注: ebiten.Imageはテスト環境では作成できないため、nilのテストのみ
}

// ============================================================
// LayerManager WindowLayerSet CRUD操作のテスト
// 要件 1.2, 1.3: ウィンドウ開閉時のWindowLayerSet管理
// ============================================================

// TestLayerManagerGetOrCreateWindowLayerSet はGetOrCreateWindowLayerSetをテストする
// 要件 1.2: ウィンドウが開かれたときにWindowLayerSetを作成する
func TestLayerManagerGetOrCreateWindowLayerSet(t *testing.T) {
	lm := NewLayerManager()

	winID := 1
	width := 640
	height := 480
	bgColor := color.RGBA{0, 0, 128, 255}

	// 最初の呼び出しで新規作成
	wls := lm.GetOrCreateWindowLayerSet(winID, width, height, bgColor)

	if wls == nil {
		t.Fatal("GetOrCreateWindowLayerSet returned nil")
	}

	if wls.WinID != winID {
		t.Errorf("WinID = %d, want %d", wls.WinID, winID)
	}

	if wls.Width != width {
		t.Errorf("Width = %d, want %d", wls.Width, width)
	}

	if wls.Height != height {
		t.Errorf("Height = %d, want %d", wls.Height, height)
	}

	if wls.BgColor != bgColor {
		t.Errorf("BgColor = %v, want %v", wls.BgColor, bgColor)
	}

	// 2回目の呼び出しで同じインスタンスを返す
	wls2 := lm.GetOrCreateWindowLayerSet(winID, 800, 600, color.RGBA{255, 255, 255, 255})

	if wls2 != wls {
		t.Error("GetOrCreateWindowLayerSet should return the same instance for the same winID")
	}

	// 元のサイズと色が維持されていることを確認
	if wls2.Width != width {
		t.Errorf("Width should remain %d, got %d", width, wls2.Width)
	}

	if wls2.BgColor != bgColor {
		t.Errorf("BgColor should remain %v, got %v", bgColor, wls2.BgColor)
	}
}

// TestLayerManagerGetWindowLayerSet はGetWindowLayerSetをテストする
// 要件 1.5: WindowIDをキーとしてWindowLayerSetを検索する
func TestLayerManagerGetWindowLayerSet(t *testing.T) {
	lm := NewLayerManager()

	// 存在しないウィンドウIDの場合はnilを返す
	wls := lm.GetWindowLayerSet(999)
	if wls != nil {
		t.Error("GetWindowLayerSet should return nil for non-existent winID")
	}

	// WindowLayerSetを作成
	winID := 1
	created := lm.GetOrCreateWindowLayerSet(winID, 640, 480, color.RGBA{0, 0, 0, 255})

	// 作成後は取得できる
	wls = lm.GetWindowLayerSet(winID)
	if wls == nil {
		t.Fatal("GetWindowLayerSet should not return nil after creation")
	}

	if wls != created {
		t.Error("GetWindowLayerSet should return the same instance")
	}
}

// TestLayerManagerDeleteWindowLayerSet はDeleteWindowLayerSetをテストする
// 要件 1.3: ウィンドウが閉じられたときにそのウィンドウに属するすべてのレイヤーを削除する
func TestLayerManagerDeleteWindowLayerSet(t *testing.T) {
	lm := NewLayerManager()

	winID := 1
	wls := lm.GetOrCreateWindowLayerSet(winID, 640, 480, color.RGBA{0, 0, 0, 255})

	// レイヤーを追加
	layer1 := newTestLayer(1)
	layer2 := newTestLayer(2)
	wls.AddLayer(layer1)
	wls.AddLayer(layer2)

	if wls.GetLayerCount() != 2 {
		t.Errorf("Layer count before delete = %d, want 2", wls.GetLayerCount())
	}

	// 削除
	lm.DeleteWindowLayerSet(winID)

	// 削除後は取得できない
	if lm.GetWindowLayerSet(winID) != nil {
		t.Error("GetWindowLayerSet should return nil after deletion")
	}
}

// TestLayerManagerDeleteWindowLayerSetNonExistent は存在しないウィンドウIDの削除をテストする
func TestLayerManagerDeleteWindowLayerSetNonExistent(t *testing.T) {
	lm := NewLayerManager()

	// 存在しないウィンドウIDの削除はパニックしない
	lm.DeleteWindowLayerSet(999) // パニックしなければOK
}

// TestLayerManagerMultipleWindowLayerSets は複数のWindowLayerSetの管理をテストする
func TestLayerManagerMultipleWindowLayerSets(t *testing.T) {
	lm := NewLayerManager()

	// 複数のウィンドウを作成
	wls1 := lm.GetOrCreateWindowLayerSet(1, 640, 480, color.RGBA{255, 0, 0, 255})
	wls2 := lm.GetOrCreateWindowLayerSet(2, 800, 600, color.RGBA{0, 255, 0, 255})
	wls3 := lm.GetOrCreateWindowLayerSet(3, 1024, 768, color.RGBA{0, 0, 255, 255})

	// 各ウィンドウが独立していることを確認
	if wls1.WinID != 1 || wls2.WinID != 2 || wls3.WinID != 3 {
		t.Error("Each WindowLayerSet should have its own WinID")
	}

	// 各ウィンドウにレイヤーを追加
	wls1.AddLayer(newTestLayer(1))
	wls2.AddLayer(newTestLayer(2))
	wls2.AddLayer(newTestLayer(3))
	wls3.AddLayer(newTestLayer(4))
	wls3.AddLayer(newTestLayer(5))
	wls3.AddLayer(newTestLayer(6))

	if wls1.GetLayerCount() != 1 {
		t.Errorf("wls1 layer count = %d, want 1", wls1.GetLayerCount())
	}
	if wls2.GetLayerCount() != 2 {
		t.Errorf("wls2 layer count = %d, want 2", wls2.GetLayerCount())
	}
	if wls3.GetLayerCount() != 3 {
		t.Errorf("wls3 layer count = %d, want 3", wls3.GetLayerCount())
	}

	// ウィンドウ2を削除
	lm.DeleteWindowLayerSet(2)

	// ウィンドウ1と3は影響を受けない
	if lm.GetWindowLayerSet(1) == nil {
		t.Error("wls1 should still exist")
	}
	if lm.GetWindowLayerSet(2) != nil {
		t.Error("wls2 should be deleted")
	}
	if lm.GetWindowLayerSet(3) == nil {
		t.Error("wls3 should still exist")
	}

	// レイヤー数も維持されている
	if lm.GetWindowLayerSet(1).GetLayerCount() != 1 {
		t.Errorf("wls1 layer count after delete = %d, want 1", lm.GetWindowLayerSet(1).GetLayerCount())
	}
	if lm.GetWindowLayerSet(3).GetLayerCount() != 3 {
		t.Errorf("wls3 layer count after delete = %d, want 3", lm.GetWindowLayerSet(3).GetLayerCount())
	}
}

// TestLayerManagerClearIncludesWindowLayerSets はClearがWindowLayerSetも削除することをテストする
func TestLayerManagerClearIncludesWindowLayerSets(t *testing.T) {
	lm := NewLayerManager()

	// WindowLayerSetを作成
	lm.GetOrCreateWindowLayerSet(1, 640, 480, color.RGBA{0, 0, 0, 255})
	lm.GetOrCreateWindowLayerSet(2, 800, 600, color.RGBA{255, 255, 255, 255})

	// PictureLayerSetも作成
	lm.GetOrCreatePictureLayerSet(100)

	// Clear
	lm.Clear()

	// すべて削除されている
	if lm.GetWindowLayerSet(1) != nil {
		t.Error("WindowLayerSet 1 should be cleared")
	}
	if lm.GetWindowLayerSet(2) != nil {
		t.Error("WindowLayerSet 2 should be cleared")
	}
	if lm.GetPictureLayerSet(100) != nil {
		t.Error("PictureLayerSet 100 should be cleared")
	}
}

// TestWindowLayerSetMarkDirty はMarkDirtyメソッドをテストする
// 要件 9.1: ダーティフラグによる部分更新をサポートする
func TestWindowLayerSetMarkDirty(t *testing.T) {
	wls := NewWindowLayerSet(1, 640, 480, color.RGBA{0, 0, 0, 255})
	wls.FullDirty = false
	wls.DirtyRegion = image.Rectangle{}

	// MarkDirtyで領域をダーティとしてマーク
	rect1 := image.Rect(10, 10, 50, 50)
	wls.MarkDirty(rect1)

	if wls.DirtyRegion != rect1 {
		t.Errorf("DirtyRegion = %v, want %v", wls.DirtyRegion, rect1)
	}

	if !wls.IsDirty() {
		t.Error("IsDirty() should return true after MarkDirty")
	}

	// 2つ目の領域をマーク（統合される）
	rect2 := image.Rect(40, 40, 100, 100)
	wls.MarkDirty(rect2)

	expected := rect1.Union(rect2)
	if wls.DirtyRegion != expected {
		t.Errorf("DirtyRegion = %v, want %v", wls.DirtyRegion, expected)
	}
}

// TestWindowLayerSetMarkDirtyEmpty は空の領域でのMarkDirtyをテストする
func TestWindowLayerSetMarkDirtyEmpty(t *testing.T) {
	wls := NewWindowLayerSet(1, 640, 480, color.RGBA{0, 0, 0, 255})
	wls.FullDirty = false
	wls.DirtyRegion = image.Rectangle{}

	// 空の領域をマーク
	wls.MarkDirty(image.Rectangle{})

	if !wls.DirtyRegion.Empty() {
		t.Error("DirtyRegion should remain empty when marking empty rect")
	}
}

// TestWindowLayerSetDirtyTrackingIntegration はダーティ領域追跡の統合テスト
// 要件 9.1, 9.2: ダーティフラグによる部分更新とキャッシュ使用
func TestWindowLayerSetDirtyTrackingIntegration(t *testing.T) {
	wls := NewWindowLayerSet(1, 640, 480, color.RGBA{0, 0, 0, 255})

	// 初期状態はFullDirty
	if !wls.IsFullDirty() {
		t.Error("IsFullDirty() should return true initially")
	}

	// ダーティフラグをクリア
	wls.ClearDirty()

	if wls.IsFullDirty() {
		t.Error("IsFullDirty() should return false after ClearDirty")
	}

	if wls.IsDirty() {
		t.Error("IsDirty() should return false after ClearDirty")
	}

	// 特定の領域をダーティとしてマーク
	rect := image.Rect(100, 100, 200, 200)
	wls.MarkDirty(rect)

	if !wls.IsDirty() {
		t.Error("IsDirty() should return true after MarkDirty")
	}

	if wls.IsFullDirty() {
		t.Error("IsFullDirty() should still be false after MarkDirty")
	}

	dirtyRegion := wls.GetDirtyRegion()
	if dirtyRegion != rect {
		t.Errorf("GetDirtyRegion() = %v, want %v", dirtyRegion, rect)
	}

	// 全体をダーティとしてマーク
	wls.MarkFullDirty()

	if !wls.IsFullDirty() {
		t.Error("IsFullDirty() should return true after MarkFullDirty")
	}

	// ClearDirtyですべてクリア
	wls.ClearDirty()

	if wls.IsFullDirty() {
		t.Error("IsFullDirty() should return false after ClearDirty")
	}

	if !wls.GetDirtyRegion().Empty() {
		t.Error("GetDirtyRegion() should be empty after ClearDirty")
	}
}

// TestWindowLayerSetLayerOperationsSetDirty はレイヤー操作がダーティフラグを設定することをテストする
// 要件 9.1: ダーティフラグによる部分更新をサポートする
func TestWindowLayerSetLayerOperationsSetDirty(t *testing.T) {
	t.Run("AddLayer sets FullDirty", func(t *testing.T) {
		wls := NewWindowLayerSet(1, 640, 480, color.RGBA{0, 0, 0, 255})
		wls.FullDirty = false

		layer := newTestLayer(1)
		wls.AddLayer(layer)

		if !wls.FullDirty {
			t.Error("AddLayer should set FullDirty to true")
		}
	})

	t.Run("RemoveLayer sets FullDirty and adds dirty region", func(t *testing.T) {
		wls := NewWindowLayerSet(1, 640, 480, color.RGBA{0, 0, 0, 255})
		layer := newTestLayer(1)
		layer.SetBounds(image.Rect(10, 10, 100, 100))
		wls.AddLayer(layer)
		wls.FullDirty = false
		wls.DirtyRegion = image.Rectangle{}

		wls.RemoveLayer(1)

		if !wls.FullDirty {
			t.Error("RemoveLayer should set FullDirty to true")
		}
	})

	t.Run("SetBgColor sets FullDirty", func(t *testing.T) {
		wls := NewWindowLayerSet(1, 640, 480, color.RGBA{0, 0, 0, 255})
		wls.FullDirty = false

		wls.SetBgColor(color.RGBA{255, 255, 255, 255})

		if !wls.FullDirty {
			t.Error("SetBgColor should set FullDirty to true")
		}
	})

	t.Run("SetSize sets FullDirty", func(t *testing.T) {
		wls := NewWindowLayerSet(1, 640, 480, color.RGBA{0, 0, 0, 255})
		wls.FullDirty = false

		wls.SetSize(800, 600)

		if !wls.FullDirty {
			t.Error("SetSize should set FullDirty to true")
		}
	})

	t.Run("ClearLayers sets FullDirty", func(t *testing.T) {
		wls := NewWindowLayerSet(1, 640, 480, color.RGBA{0, 0, 0, 255})
		layer := newTestLayer(1)
		wls.AddLayer(layer)
		wls.FullDirty = false

		wls.ClearLayers()

		if !wls.FullDirty {
			t.Error("ClearLayers should set FullDirty to true")
		}
	})
}

// ============================================================
// エラーハンドリングのテスト
// 要件 10.2: 存在しないレイヤーIDが指定されたときにエラーをログに記録し、処理をスキップする
// ============================================================

// TestWindowLayerSetGetLayerErrorHandling は存在しないレイヤーIDのエラー処理をテストする
// 要件 10.2: 存在しないレイヤーIDが指定されたときにエラーをログに記録し、処理をスキップする
func TestWindowLayerSetGetLayerErrorHandling(t *testing.T) {
	wls := NewWindowLayerSet(1, 640, 480, color.RGBA{0, 0, 0, 255})

	// レイヤーを追加
	layer1 := newTestLayer(1)
	wls.AddLayer(layer1)

	// 存在しないレイヤーIDを指定
	// エラーがログに記録され、nilが返される
	result := wls.GetLayer(999)

	if result != nil {
		t.Error("GetLayer should return nil for non-existent layer ID")
	}

	// 存在するレイヤーIDは正常に取得できる
	result = wls.GetLayer(1)
	if result == nil {
		t.Error("GetLayer should return layer for existing layer ID")
	}
}

// TestWindowLayerSetRemoveLayerErrorHandling は存在しないレイヤーIDの削除エラー処理をテストする
// 要件 10.2: 存在しないレイヤーIDが指定されたときにエラーをログに記録し、処理をスキップする
func TestWindowLayerSetRemoveLayerErrorHandling(t *testing.T) {
	wls := NewWindowLayerSet(1, 640, 480, color.RGBA{0, 0, 0, 255})

	// レイヤーを追加
	layer1 := newTestLayer(1)
	wls.AddLayer(layer1)

	initialCount := wls.GetLayerCount()

	// 存在しないレイヤーIDを指定
	// エラーがログに記録され、falseが返される
	removed := wls.RemoveLayer(999)

	if removed {
		t.Error("RemoveLayer should return false for non-existent layer ID")
	}

	// レイヤー数は変わらない
	if wls.GetLayerCount() != initialCount {
		t.Errorf("Layer count should remain %d, got %d", initialCount, wls.GetLayerCount())
	}

	// 存在するレイヤーIDは正常に削除できる
	removed = wls.RemoveLayer(1)
	if !removed {
		t.Error("RemoveLayer should return true for existing layer ID")
	}
}

// TestPictureLayerSetGetCastLayerByIDErrorHandling は存在しないレイヤーIDのエラー処理をテストする
// 要件 10.2: 存在しないレイヤーIDが指定されたときにエラーをログに記録し、処理をスキップする
func TestPictureLayerSetGetCastLayerByIDErrorHandling(t *testing.T) {
	pls := NewPictureLayerSet(1)

	// キャストレイヤーを追加
	castLayer := NewCastLayer(1, 0, 1, 0, 0, 0, 0, 0, 50, 50, 0)
	pls.AddCastLayer(castLayer)

	// 存在しないレイヤーIDを指定
	// エラーがログに記録され、nilが返される
	result := pls.GetCastLayerByID(999)

	if result != nil {
		t.Error("GetCastLayerByID should return nil for non-existent layer ID")
	}

	// 存在するレイヤーIDは正常に取得できる
	result = pls.GetCastLayerByID(1)
	if result == nil {
		t.Error("GetCastLayerByID should return layer for existing layer ID")
	}
}

// TestPictureLayerSetRemoveCastLayerByIDErrorHandling は存在しないレイヤーIDの削除エラー処理をテストする
// 要件 10.2: 存在しないレイヤーIDが指定されたときにエラーをログに記録し、処理をスキップする
func TestPictureLayerSetRemoveCastLayerByIDErrorHandling(t *testing.T) {
	pls := NewPictureLayerSet(1)

	// キャストレイヤーを追加
	castLayer := NewCastLayer(1, 0, 1, 0, 0, 0, 0, 0, 50, 50, 0)
	pls.AddCastLayer(castLayer)

	initialCount := pls.GetCastLayerCount()

	// 存在しないレイヤーIDを指定
	// エラーがログに記録され、falseが返される
	removed := pls.RemoveCastLayerByID(999)

	if removed {
		t.Error("RemoveCastLayerByID should return false for non-existent layer ID")
	}

	// レイヤー数は変わらない
	if pls.GetCastLayerCount() != initialCount {
		t.Errorf("Cast layer count should remain %d, got %d", initialCount, pls.GetCastLayerCount())
	}

	// 存在するレイヤーIDは正常に削除できる
	removed = pls.RemoveCastLayerByID(1)
	if !removed {
		t.Error("RemoveCastLayerByID should return true for existing layer ID")
	}
}

// TestPictureLayerSetGetTextLayerErrorHandling は存在しないレイヤーIDのエラー処理をテストする
// 要件 10.2: 存在しないレイヤーIDが指定されたときにエラーをログに記録し、処理をスキップする
func TestPictureLayerSetGetTextLayerErrorHandling(t *testing.T) {
	pls := NewPictureLayerSet(1)

	// テキストレイヤーを追加
	textLayer := NewTextLayerEntry(1, 1, 0, 0, "test", 0)
	pls.AddTextLayer(textLayer)

	// 存在しないレイヤーIDを指定
	// エラーがログに記録され、nilが返される
	result := pls.GetTextLayer(999)

	if result != nil {
		t.Error("GetTextLayer should return nil for non-existent layer ID")
	}

	// 存在するレイヤーIDは正常に取得できる
	result = pls.GetTextLayer(1)
	if result == nil {
		t.Error("GetTextLayer should return layer for existing layer ID")
	}
}

// TestPictureLayerSetRemoveTextLayerErrorHandling は存在しないレイヤーIDの削除エラー処理をテストする
// 要件 10.2: 存在しないレイヤーIDが指定されたときにエラーをログに記録し、処理をスキップする
func TestPictureLayerSetRemoveTextLayerErrorHandling(t *testing.T) {
	pls := NewPictureLayerSet(1)

	// テキストレイヤーを追加
	textLayer := NewTextLayerEntry(1, 1, 0, 0, "test", 0)
	pls.AddTextLayer(textLayer)

	initialCount := pls.GetTextLayerCount()

	// 存在しないレイヤーIDを指定
	// エラーがログに記録され、falseが返される
	removed := pls.RemoveTextLayer(999)

	if removed {
		t.Error("RemoveTextLayer should return false for non-existent layer ID")
	}

	// レイヤー数は変わらない
	if pls.GetTextLayerCount() != initialCount {
		t.Errorf("Text layer count should remain %d, got %d", initialCount, pls.GetTextLayerCount())
	}

	// 存在するレイヤーIDは正常に削除できる
	removed = pls.RemoveTextLayer(1)
	if !removed {
		t.Error("RemoveTextLayer should return true for existing layer ID")
	}
}
