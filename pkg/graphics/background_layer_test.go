package graphics

import (
	"image"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// TestNewBackgroundLayer は新しい背景レイヤーの作成をテストする
func TestNewBackgroundLayer(t *testing.T) {
	// nilイメージでの作成
	t.Run("nil image", func(t *testing.T) {
		layer := NewBackgroundLayer(1, 100, nil)

		if layer.GetID() != 1 {
			t.Errorf("expected ID 1, got %d", layer.GetID())
		}
		if layer.GetPicID() != 100 {
			t.Errorf("expected PicID 100, got %d", layer.GetPicID())
		}
		if layer.GetZOrder() != ZOrderBackground {
			t.Errorf("expected ZOrder %d, got %d", ZOrderBackground, layer.GetZOrder())
		}
		if !layer.IsVisible() {
			t.Error("expected layer to be visible")
		}
		if !layer.IsDirty() {
			t.Error("expected layer to be dirty on creation")
		}
		if layer.GetImage() != nil {
			t.Error("expected nil image")
		}
		if !layer.GetBounds().Empty() {
			t.Errorf("expected empty bounds, got %v", layer.GetBounds())
		}
	})

	// 有効なイメージでの作成
	t.Run("with image", func(t *testing.T) {
		img := ebiten.NewImage(640, 480)
		layer := NewBackgroundLayer(2, 200, img)

		if layer.GetID() != 2 {
			t.Errorf("expected ID 2, got %d", layer.GetID())
		}
		if layer.GetPicID() != 200 {
			t.Errorf("expected PicID 200, got %d", layer.GetPicID())
		}
		if layer.GetZOrder() != ZOrderBackground {
			t.Errorf("expected ZOrder %d, got %d", ZOrderBackground, layer.GetZOrder())
		}
		if layer.GetImage() != img {
			t.Error("expected same image")
		}

		expectedBounds := image.Rect(0, 0, 640, 480)
		if layer.GetBounds() != expectedBounds {
			t.Errorf("expected bounds %v, got %v", expectedBounds, layer.GetBounds())
		}
	})
}

// TestBackgroundLayerZOrderIsAlwaysZero はZ順序が常に0であることをテストする
func TestBackgroundLayerZOrderIsAlwaysZero(t *testing.T) {
	layer := NewBackgroundLayer(1, 100, nil)

	// 初期値が0であることを確認
	if layer.GetZOrder() != 0 {
		t.Errorf("expected ZOrder 0, got %d", layer.GetZOrder())
	}

	// ZOrderBackgroundが0であることを確認
	if ZOrderBackground != 0 {
		t.Errorf("expected ZOrderBackground to be 0, got %d", ZOrderBackground)
	}
}

// TestBackgroundLayerSetImage は画像の設定をテストする
func TestBackgroundLayerSetImage(t *testing.T) {
	layer := NewBackgroundLayer(1, 100, nil)

	// ダーティフラグをクリア
	layer.SetDirty(false)

	// 新しい画像を設定
	img := ebiten.NewImage(800, 600)
	layer.SetImage(img)

	if layer.GetImage() != img {
		t.Error("expected same image after SetImage")
	}
	if !layer.IsDirty() {
		t.Error("expected layer to be dirty after SetImage")
	}

	expectedBounds := image.Rect(0, 0, 800, 600)
	if layer.GetBounds() != expectedBounds {
		t.Errorf("expected bounds %v, got %v", expectedBounds, layer.GetBounds())
	}

	// nilに設定
	layer.SetDirty(false)
	layer.SetImage(nil)

	if layer.GetImage() != nil {
		t.Error("expected nil image after SetImage(nil)")
	}
	if !layer.IsDirty() {
		t.Error("expected layer to be dirty after SetImage(nil)")
	}
	if !layer.GetBounds().Empty() {
		t.Errorf("expected empty bounds after SetImage(nil), got %v", layer.GetBounds())
	}
}

// TestBackgroundLayerInvalidate はキャッシュ無効化をテストする
func TestBackgroundLayerInvalidate(t *testing.T) {
	layer := NewBackgroundLayer(1, 100, nil)

	// ダーティフラグをクリア
	layer.SetDirty(false)
	if layer.IsDirty() {
		t.Error("expected layer to not be dirty after SetDirty(false)")
	}

	// Invalidateを呼び出す
	layer.Invalidate()
	if !layer.IsDirty() {
		t.Error("expected layer to be dirty after Invalidate")
	}
}

// TestBackgroundLayerVisibility は可視性の設定をテストする
func TestBackgroundLayerVisibility(t *testing.T) {
	layer := NewBackgroundLayer(1, 100, nil)

	// 初期状態は可視
	if !layer.IsVisible() {
		t.Error("expected layer to be visible initially")
	}

	// ダーティフラグをクリア
	layer.SetDirty(false)

	// 非表示に設定
	layer.SetVisible(false)
	if layer.IsVisible() {
		t.Error("expected layer to be invisible after SetVisible(false)")
	}
	if !layer.IsDirty() {
		t.Error("expected layer to be dirty after visibility change")
	}

	// ダーティフラグをクリア
	layer.SetDirty(false)

	// 同じ値を設定（変更なし）
	layer.SetVisible(false)
	if layer.IsDirty() {
		t.Error("expected layer to not be dirty when visibility unchanged")
	}
}

// TestBackgroundLayerImplementsLayerInterface はLayerインターフェースの実装をテストする
func TestBackgroundLayerImplementsLayerInterface(t *testing.T) {
	var _ Layer = (*BackgroundLayer)(nil)
}

// TestBackgroundLayerSetPicID はピクチャーIDの設定をテストする
func TestBackgroundLayerSetPicID(t *testing.T) {
	layer := NewBackgroundLayer(1, 100, nil)

	if layer.GetPicID() != 100 {
		t.Errorf("expected PicID 100, got %d", layer.GetPicID())
	}

	layer.SetPicID(200)
	if layer.GetPicID() != 200 {
		t.Errorf("expected PicID 200 after SetPicID, got %d", layer.GetPicID())
	}
}
