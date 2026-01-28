package graphics

import (
	"image"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// TestBaseLayerGetID はBaseLayerのGetIDメソッドをテストする
func TestBaseLayerGetID(t *testing.T) {
	layer := &BaseLayer{}
	layer.SetID(42)

	if got := layer.GetID(); got != 42 {
		t.Errorf("GetID() = %d, want 42", got)
	}
}

// TestBaseLayerGetBounds はBaseLayerのGetBoundsメソッドをテストする
func TestBaseLayerGetBounds(t *testing.T) {
	layer := &BaseLayer{}
	bounds := image.Rect(10, 20, 100, 200)
	layer.SetBounds(bounds)

	if got := layer.GetBounds(); got != bounds {
		t.Errorf("GetBounds() = %v, want %v", got, bounds)
	}
}

// TestBaseLayerGetZOrder はBaseLayerのGetZOrderメソッドをテストする
func TestBaseLayerGetZOrder(t *testing.T) {
	layer := &BaseLayer{}
	layer.SetZOrder(ZOrderCastBase + 5)

	if got := layer.GetZOrder(); got != ZOrderCastBase+5 {
		t.Errorf("GetZOrder() = %d, want %d", got, ZOrderCastBase+5)
	}
}

// TestBaseLayerIsVisible はBaseLayerのIsVisibleメソッドをテストする
func TestBaseLayerIsVisible(t *testing.T) {
	layer := &BaseLayer{}

	// デフォルトはfalse
	if layer.IsVisible() {
		t.Error("IsVisible() should be false by default")
	}

	layer.SetVisible(true)
	if !layer.IsVisible() {
		t.Error("IsVisible() should be true after SetVisible(true)")
	}

	layer.SetVisible(false)
	if layer.IsVisible() {
		t.Error("IsVisible() should be false after SetVisible(false)")
	}
}

// TestBaseLayerIsDirty はBaseLayerのIsDirtyメソッドをテストする
func TestBaseLayerIsDirty(t *testing.T) {
	layer := &BaseLayer{}

	// デフォルトはfalse
	if layer.IsDirty() {
		t.Error("IsDirty() should be false by default")
	}

	layer.SetDirty(true)
	if !layer.IsDirty() {
		t.Error("IsDirty() should be true after SetDirty(true)")
	}

	layer.SetDirty(false)
	if layer.IsDirty() {
		t.Error("IsDirty() should be false after SetDirty(false)")
	}
}

// TestBaseLayerSetVisibleSetsDirty は可視性変更時にダーティフラグが設定されることをテストする
// 要件 3.3: 可視性が変更されたときにダーティフラグを設定
func TestBaseLayerSetVisibleSetsDirty(t *testing.T) {
	layer := &BaseLayer{}
	layer.SetDirty(false)

	// 可視性を変更するとダーティフラグが設定される
	layer.SetVisible(true)
	if !layer.IsDirty() {
		t.Error("SetVisible should set dirty flag when visibility changes")
	}

	// ダーティフラグをクリア
	layer.SetDirty(false)

	// 同じ値を設定してもダーティフラグは設定されない
	layer.SetVisible(true)
	if layer.IsDirty() {
		t.Error("SetVisible should not set dirty flag when visibility doesn't change")
	}
}

// TestBaseLayerSetBoundsSetsDirty は境界変更時にダーティフラグが設定されることをテストする
// 要件 3.1: 位置が変更されたときにダーティフラグを設定
func TestBaseLayerSetBoundsSetsDirty(t *testing.T) {
	layer := &BaseLayer{}
	layer.SetDirty(false)

	// 境界を変更するとダーティフラグが設定される
	bounds := image.Rect(10, 20, 100, 200)
	layer.SetBounds(bounds)
	if !layer.IsDirty() {
		t.Error("SetBounds should set dirty flag when bounds change")
	}

	// ダーティフラグをクリア
	layer.SetDirty(false)

	// 同じ値を設定してもダーティフラグは設定されない
	layer.SetBounds(bounds)
	if layer.IsDirty() {
		t.Error("SetBounds should not set dirty flag when bounds don't change")
	}
}

// TestZOrderConstants はZ順序の定数が正しい順序であることをテストする
// 要件 1.6: 背景 → 描画 → キャスト → テキストの順序
func TestZOrderConstants(t *testing.T) {
	// 背景 < 描画 < キャスト < テキスト
	if ZOrderBackground >= ZOrderDrawing {
		t.Errorf("ZOrderBackground (%d) should be less than ZOrderDrawing (%d)",
			ZOrderBackground, ZOrderDrawing)
	}
	if ZOrderDrawing >= ZOrderCastBase {
		t.Errorf("ZOrderDrawing (%d) should be less than ZOrderCastBase (%d)",
			ZOrderDrawing, ZOrderCastBase)
	}
	if ZOrderCastMax >= ZOrderTextBase {
		t.Errorf("ZOrderCastMax (%d) should be less than ZOrderTextBase (%d)",
			ZOrderCastMax, ZOrderTextBase)
	}
}

// mockLayer はLayerインターフェースのモック実装
type mockLayer struct {
	BaseLayer
	image     *ebiten.Image
	layerType LayerType
}

func (m *mockLayer) GetImage() *ebiten.Image {
	return m.image
}

func (m *mockLayer) Invalidate() {
	m.image = nil
	m.dirty = true
}

func (m *mockLayer) GetLayerType() LayerType {
	return m.layerType
}

// TestLayerInterface はLayerインターフェースが正しく実装されていることをテストする
func TestLayerInterface(t *testing.T) {
	// mockLayerがLayerインターフェースを実装していることを確認
	var _ Layer = &mockLayer{}

	mock := &mockLayer{}
	mock.SetID(1)
	mock.SetBounds(image.Rect(0, 0, 100, 100))
	mock.SetZOrder(ZOrderCastBase)
	mock.SetVisible(true)

	// インターフェースメソッドのテスト
	if mock.GetID() != 1 {
		t.Errorf("GetID() = %d, want 1", mock.GetID())
	}
	if mock.GetBounds() != image.Rect(0, 0, 100, 100) {
		t.Errorf("GetBounds() = %v, want %v", mock.GetBounds(), image.Rect(0, 0, 100, 100))
	}
	if mock.GetZOrder() != ZOrderCastBase {
		t.Errorf("GetZOrder() = %d, want %d", mock.GetZOrder(), ZOrderCastBase)
	}
	if !mock.IsVisible() {
		t.Error("IsVisible() should be true")
	}

	// Invalidateのテスト
	mock.SetDirty(false)
	mock.Invalidate()
	if !mock.IsDirty() {
		t.Error("Invalidate should set dirty flag")
	}
}
