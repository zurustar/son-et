package graphics

import (
	"image"
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// TestNewDrawingLayer は新しい描画レイヤーの作成をテストする
func TestNewDrawingLayer(t *testing.T) {
	// ゼロサイズでの作成
	t.Run("zero size", func(t *testing.T) {
		layer := NewDrawingLayer(1, 100, 0, 0)

		if layer.GetID() != 1 {
			t.Errorf("expected ID 1, got %d", layer.GetID())
		}
		if layer.GetPicID() != 100 {
			t.Errorf("expected PicID 100, got %d", layer.GetPicID())
		}
		if layer.GetZOrder() != ZOrderDrawing {
			t.Errorf("expected ZOrder %d, got %d", ZOrderDrawing, layer.GetZOrder())
		}
		if !layer.IsVisible() {
			t.Error("expected layer to be visible")
		}
		if !layer.IsDirty() {
			t.Error("expected layer to be dirty on creation")
		}
		if layer.GetImage() != nil {
			t.Error("expected nil image for zero size")
		}
		if !layer.GetBounds().Empty() {
			t.Errorf("expected empty bounds, got %v", layer.GetBounds())
		}
	})

	// 有効なサイズでの作成
	t.Run("with valid size", func(t *testing.T) {
		layer := NewDrawingLayer(2, 200, 640, 480)

		if layer.GetID() != 2 {
			t.Errorf("expected ID 2, got %d", layer.GetID())
		}
		if layer.GetPicID() != 200 {
			t.Errorf("expected PicID 200, got %d", layer.GetPicID())
		}
		if layer.GetZOrder() != ZOrderDrawing {
			t.Errorf("expected ZOrder %d, got %d", ZOrderDrawing, layer.GetZOrder())
		}
		if layer.GetImage() == nil {
			t.Error("expected non-nil image")
		}

		expectedBounds := image.Rect(0, 0, 640, 480)
		if layer.GetBounds() != expectedBounds {
			t.Errorf("expected bounds %v, got %v", expectedBounds, layer.GetBounds())
		}
	})

	// 負のサイズでの作成
	t.Run("negative size", func(t *testing.T) {
		layer := NewDrawingLayer(3, 300, -100, -100)

		if layer.GetImage() != nil {
			t.Error("expected nil image for negative size")
		}
		if !layer.GetBounds().Empty() {
			t.Errorf("expected empty bounds for negative size, got %v", layer.GetBounds())
		}
	})
}

// TestNewDrawingLayerWithImage は既存の画像から描画レイヤーを作成するテスト
func TestNewDrawingLayerWithImage(t *testing.T) {
	// nilイメージでの作成
	t.Run("nil image", func(t *testing.T) {
		layer := NewDrawingLayerWithImage(1, 100, nil)

		if layer.GetID() != 1 {
			t.Errorf("expected ID 1, got %d", layer.GetID())
		}
		if layer.GetPicID() != 100 {
			t.Errorf("expected PicID 100, got %d", layer.GetPicID())
		}
		if layer.GetZOrder() != ZOrderDrawing {
			t.Errorf("expected ZOrder %d, got %d", ZOrderDrawing, layer.GetZOrder())
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
		img := ebiten.NewImage(800, 600)
		layer := NewDrawingLayerWithImage(2, 200, img)

		if layer.GetID() != 2 {
			t.Errorf("expected ID 2, got %d", layer.GetID())
		}
		if layer.GetImage() != img {
			t.Error("expected same image")
		}

		expectedBounds := image.Rect(0, 0, 800, 600)
		if layer.GetBounds() != expectedBounds {
			t.Errorf("expected bounds %v, got %v", expectedBounds, layer.GetBounds())
		}
	})
}

// TestDrawingLayerZOrderIsAlwaysOne はZ順序が常に1であることをテストする
func TestDrawingLayerZOrderIsAlwaysOne(t *testing.T) {
	layer := NewDrawingLayer(1, 100, 640, 480)

	// 初期値が1であることを確認
	if layer.GetZOrder() != 1 {
		t.Errorf("expected ZOrder 1, got %d", layer.GetZOrder())
	}

	// ZOrderDrawingが1であることを確認
	if ZOrderDrawing != 1 {
		t.Errorf("expected ZOrderDrawing to be 1, got %d", ZOrderDrawing)
	}

	// 背景より大きく、キャストより小さいことを確認
	if ZOrderDrawing <= ZOrderBackground {
		t.Errorf("expected ZOrderDrawing > ZOrderBackground, got %d <= %d", ZOrderDrawing, ZOrderBackground)
	}
	if ZOrderDrawing >= ZOrderCastBase {
		t.Errorf("expected ZOrderDrawing < ZOrderCastBase, got %d >= %d", ZOrderDrawing, ZOrderCastBase)
	}
}

// TestDrawingLayerSetImage は画像の設定をテストする
func TestDrawingLayerSetImage(t *testing.T) {
	layer := NewDrawingLayer(1, 100, 640, 480)

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

// TestDrawingLayerInvalidate はキャッシュ無効化をテストする
func TestDrawingLayerInvalidate(t *testing.T) {
	layer := NewDrawingLayer(1, 100, 640, 480)

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

// TestDrawingLayerVisibility は可視性の設定をテストする
func TestDrawingLayerVisibility(t *testing.T) {
	layer := NewDrawingLayer(1, 100, 640, 480)

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

// TestDrawingLayerClear はクリア機能をテストする
func TestDrawingLayerClear(t *testing.T) {
	layer := NewDrawingLayer(1, 100, 100, 100)

	// 画像に何か描画
	img := layer.GetImage()
	if img != nil {
		img.Fill(color.RGBA{255, 0, 0, 255})
	}

	// ダーティフラグをクリア
	layer.SetDirty(false)

	// クリアを呼び出す
	layer.Clear()

	if !layer.IsDirty() {
		t.Error("expected layer to be dirty after Clear")
	}

	// nilイメージの場合はパニックしないことを確認
	nilLayer := NewDrawingLayer(2, 200, 0, 0)
	nilLayer.Clear() // パニックしないことを確認
}

// TestDrawingLayerDrawImage は画像描画機能をテストする
func TestDrawingLayerDrawImage(t *testing.T) {
	layer := NewDrawingLayer(1, 100, 200, 200)

	// ソース画像を作成
	srcImg := ebiten.NewImage(50, 50)
	srcImg.Fill(color.RGBA{255, 0, 0, 255})

	// ダーティフラグをクリア
	layer.SetDirty(false)

	// 画像を描画
	layer.DrawImage(srcImg, 10, 20)

	if !layer.IsDirty() {
		t.Error("expected layer to be dirty after DrawImage")
	}

	// nilソースの場合はパニックしないことを確認
	layer.SetDirty(false)
	layer.DrawImage(nil, 0, 0)
	if layer.IsDirty() {
		t.Error("expected layer to not be dirty after DrawImage with nil source")
	}

	// nilレイヤー画像の場合はパニックしないことを確認
	nilLayer := NewDrawingLayer(2, 200, 0, 0)
	nilLayer.DrawImage(srcImg, 0, 0) // パニックしないことを確認
}

// TestDrawingLayerDrawSubImage は部分画像描画機能をテストする
func TestDrawingLayerDrawSubImage(t *testing.T) {
	layer := NewDrawingLayer(1, 100, 200, 200)

	// ソース画像を作成
	srcImg := ebiten.NewImage(100, 100)
	srcImg.Fill(color.RGBA{0, 255, 0, 255})

	// ダーティフラグをクリア
	layer.SetDirty(false)

	// 部分画像を描画
	layer.DrawSubImage(srcImg, 10, 20, 25, 25, 50, 50)

	if !layer.IsDirty() {
		t.Error("expected layer to be dirty after DrawSubImage")
	}

	// nilソースの場合はパニックしないことを確認
	layer.SetDirty(false)
	layer.DrawSubImage(nil, 0, 0, 0, 0, 10, 10)
	if layer.IsDirty() {
		t.Error("expected layer to not be dirty after DrawSubImage with nil source")
	}

	// nilレイヤー画像の場合はパニックしないことを確認
	nilLayer := NewDrawingLayer(2, 200, 0, 0)
	nilLayer.DrawSubImage(srcImg, 0, 0, 0, 0, 10, 10) // パニックしないことを確認
}

// TestDrawingLayerResize はリサイズ機能をテストする
func TestDrawingLayerResize(t *testing.T) {
	layer := NewDrawingLayer(1, 100, 100, 100)

	// ダーティフラグをクリア
	layer.SetDirty(false)

	// リサイズ
	layer.Resize(200, 150)

	if !layer.IsDirty() {
		t.Error("expected layer to be dirty after Resize")
	}

	expectedBounds := image.Rect(0, 0, 200, 150)
	if layer.GetBounds() != expectedBounds {
		t.Errorf("expected bounds %v, got %v", expectedBounds, layer.GetBounds())
	}

	if layer.GetImage() == nil {
		t.Error("expected non-nil image after Resize")
	}

	// ゼロサイズにリサイズ
	layer.SetDirty(false)
	layer.Resize(0, 0)

	if !layer.IsDirty() {
		t.Error("expected layer to be dirty after Resize to zero")
	}
	if layer.GetImage() != nil {
		t.Error("expected nil image after Resize to zero")
	}
	if !layer.GetBounds().Empty() {
		t.Errorf("expected empty bounds after Resize to zero, got %v", layer.GetBounds())
	}
}

// TestDrawingLayerCopyFrom はコピー機能をテストする
func TestDrawingLayerCopyFrom(t *testing.T) {
	layer := NewDrawingLayer(1, 100, 50, 50)

	// ソース画像を作成
	srcImg := ebiten.NewImage(100, 100)
	srcImg.Fill(color.RGBA{0, 0, 255, 255})

	// ダーティフラグをクリア
	layer.SetDirty(false)

	// コピー
	layer.CopyFrom(srcImg)

	if !layer.IsDirty() {
		t.Error("expected layer to be dirty after CopyFrom")
	}

	// サイズが調整されていることを確認
	expectedBounds := image.Rect(0, 0, 100, 100)
	if layer.GetBounds() != expectedBounds {
		t.Errorf("expected bounds %v, got %v", expectedBounds, layer.GetBounds())
	}

	// nilソースの場合はパニックしないことを確認
	layer.SetDirty(false)
	layer.CopyFrom(nil)
	if layer.IsDirty() {
		t.Error("expected layer to not be dirty after CopyFrom with nil source")
	}
}

// TestDrawingLayerImplementsLayerInterface はLayerインターフェースの実装をテストする
func TestDrawingLayerImplementsLayerInterface(t *testing.T) {
	var _ Layer = (*DrawingLayer)(nil)
}

// TestDrawingLayerSetPicID はピクチャーIDの設定をテストする
func TestDrawingLayerSetPicID(t *testing.T) {
	layer := NewDrawingLayer(1, 100, 640, 480)

	if layer.GetPicID() != 100 {
		t.Errorf("expected PicID 100, got %d", layer.GetPicID())
	}

	layer.SetPicID(200)
	if layer.GetPicID() != 200 {
		t.Errorf("expected PicID 200 after SetPicID, got %d", layer.GetPicID())
	}
}

// TestDrawingLayerZOrderRelationship はZ順序の関係をテストする
// 要件 1.6: 背景 → 描画 → キャスト → テキストの順序
func TestDrawingLayerZOrderRelationship(t *testing.T) {
	bgLayer := NewBackgroundLayer(1, 100, nil)
	drawingLayer := NewDrawingLayer(2, 100, 640, 480)

	// 描画レイヤーは背景レイヤーより前面にあることを確認
	if drawingLayer.GetZOrder() <= bgLayer.GetZOrder() {
		t.Errorf("expected DrawingLayer ZOrder > BackgroundLayer ZOrder, got %d <= %d",
			drawingLayer.GetZOrder(), bgLayer.GetZOrder())
	}

	// 描画レイヤーはキャストレイヤーより背面にあることを確認
	if drawingLayer.GetZOrder() >= ZOrderCastBase {
		t.Errorf("expected DrawingLayer ZOrder < ZOrderCastBase, got %d >= %d",
			drawingLayer.GetZOrder(), ZOrderCastBase)
	}
}
