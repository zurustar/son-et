package graphics

import (
	"image"
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// TestNewCastLayer は新しいキャストレイヤーの作成をテストする
func TestNewCastLayer(t *testing.T) {
	t.Run("basic creation", func(t *testing.T) {
		layer := NewCastLayer(1, 10, 100, 200, 50, 60, 0, 0, 32, 32, 0)

		if layer.GetID() != 1 {
			t.Errorf("expected ID 1, got %d", layer.GetID())
		}
		if layer.GetCastID() != 10 {
			t.Errorf("expected CastID 10, got %d", layer.GetCastID())
		}
		if layer.GetPicID() != 100 {
			t.Errorf("expected PicID 100, got %d", layer.GetPicID())
		}
		if layer.GetSrcPicID() != 200 {
			t.Errorf("expected SrcPicID 200, got %d", layer.GetSrcPicID())
		}
		if layer.GetZOrder() != ZOrderCastBase {
			t.Errorf("expected ZOrder %d, got %d", ZOrderCastBase, layer.GetZOrder())
		}
		if !layer.IsVisible() {
			t.Error("expected layer to be visible")
		}
		if !layer.IsDirty() {
			t.Error("expected layer to be dirty on creation")
		}

		x, y := layer.GetPosition()
		if x != 50 || y != 60 {
			t.Errorf("expected position (50, 60), got (%d, %d)", x, y)
		}

		srcX, srcY, width, height := layer.GetSourceRect()
		if srcX != 0 || srcY != 0 || width != 32 || height != 32 {
			t.Errorf("expected source rect (0, 0, 32, 32), got (%d, %d, %d, %d)", srcX, srcY, width, height)
		}

		expectedBounds := image.Rect(50, 60, 82, 92)
		if layer.GetBounds() != expectedBounds {
			t.Errorf("expected bounds %v, got %v", expectedBounds, layer.GetBounds())
		}
	})

	t.Run("with zOrder offset", func(t *testing.T) {
		layer := NewCastLayer(2, 20, 100, 200, 0, 0, 0, 0, 32, 32, 5)

		expectedZOrder := ZOrderCastBase + 5
		if layer.GetZOrder() != expectedZOrder {
			t.Errorf("expected ZOrder %d, got %d", expectedZOrder, layer.GetZOrder())
		}
	})
}

// TestNewCastLayerWithTransColor は透明色付きキャストレイヤーの作成をテストする
func TestNewCastLayerWithTransColor(t *testing.T) {
	t.Run("with trans color", func(t *testing.T) {
		transColor := color.RGBA{255, 0, 255, 255} // マゼンタ
		layer := NewCastLayerWithTransColor(1, 10, 100, 200, 0, 0, 0, 0, 32, 32, 0, transColor)

		if !layer.HasTransColor() {
			t.Error("expected layer to have trans color")
		}

		gotColor := layer.GetTransColor()
		if gotColor == nil {
			t.Error("expected non-nil trans color")
		}

		r1, g1, b1, _ := transColor.RGBA()
		r2, g2, b2, _ := gotColor.RGBA()
		if r1 != r2 || g1 != g2 || b1 != b2 {
			t.Errorf("expected trans color %v, got %v", transColor, gotColor)
		}
	})

	t.Run("with nil trans color", func(t *testing.T) {
		layer := NewCastLayerWithTransColor(1, 10, 100, 200, 0, 0, 0, 0, 32, 32, 0, nil)

		if layer.HasTransColor() {
			t.Error("expected layer to not have trans color")
		}
		if layer.GetTransColor() != nil {
			t.Error("expected nil trans color")
		}
	})
}

// TestCastLayerZOrderStartsAt100 はZ順序が100から開始することをテストする
func TestCastLayerZOrderStartsAt100(t *testing.T) {
	layer := NewCastLayer(1, 10, 100, 200, 0, 0, 0, 0, 32, 32, 0)

	// Z順序が100であることを確認
	if layer.GetZOrder() != 100 {
		t.Errorf("expected ZOrder 100, got %d", layer.GetZOrder())
	}

	// ZOrderCastBaseが100であることを確認
	if ZOrderCastBase != 100 {
		t.Errorf("expected ZOrderCastBase to be 100, got %d", ZOrderCastBase)
	}

	// 描画レイヤーより大きいことを確認
	if layer.GetZOrder() <= ZOrderDrawing {
		t.Errorf("expected CastLayer ZOrder > ZOrderDrawing, got %d <= %d", layer.GetZOrder(), ZOrderDrawing)
	}

	// テキストレイヤーより小さいことを確認
	if layer.GetZOrder() >= ZOrderTextBase {
		t.Errorf("expected CastLayer ZOrder < ZOrderTextBase, got %d >= %d", layer.GetZOrder(), ZOrderTextBase)
	}
}

// TestCastLayerSetPosition は位置設定をテストする
func TestCastLayerSetPosition(t *testing.T) {
	layer := NewCastLayer(1, 10, 100, 200, 0, 0, 0, 0, 32, 32, 0)

	// ダーティフラグをクリア
	layer.SetDirty(false)

	// 位置を変更
	layer.SetPosition(100, 150)

	x, y := layer.GetPosition()
	if x != 100 || y != 150 {
		t.Errorf("expected position (100, 150), got (%d, %d)", x, y)
	}

	if !layer.IsDirty() {
		t.Error("expected layer to be dirty after SetPosition")
	}

	expectedBounds := image.Rect(100, 150, 132, 182)
	if layer.GetBounds() != expectedBounds {
		t.Errorf("expected bounds %v, got %v", expectedBounds, layer.GetBounds())
	}

	// 同じ位置を設定（変更なし）
	layer.SetDirty(false)
	layer.SetPosition(100, 150)
	if layer.IsDirty() {
		t.Error("expected layer to not be dirty when position unchanged")
	}
}

// TestCastLayerSetTransColor は透明色設定をテストする
func TestCastLayerSetTransColor(t *testing.T) {
	layer := NewCastLayer(1, 10, 100, 200, 0, 0, 0, 0, 32, 32, 0)

	// 初期状態は透明色なし
	if layer.HasTransColor() {
		t.Error("expected layer to not have trans color initially")
	}

	// ダーティフラグをクリア
	layer.SetDirty(false)

	// 透明色を設定
	transColor := color.RGBA{0, 255, 0, 255} // 緑
	layer.SetTransColor(transColor)

	if !layer.HasTransColor() {
		t.Error("expected layer to have trans color after SetTransColor")
	}
	if !layer.IsDirty() {
		t.Error("expected layer to be dirty after SetTransColor")
	}

	// nilに設定
	layer.SetDirty(false)
	layer.SetTransColor(nil)

	if layer.HasTransColor() {
		t.Error("expected layer to not have trans color after SetTransColor(nil)")
	}
	if !layer.IsDirty() {
		t.Error("expected layer to be dirty after SetTransColor(nil)")
	}
}

// TestCastLayerSetSourceRect はソース領域設定をテストする
func TestCastLayerSetSourceRect(t *testing.T) {
	layer := NewCastLayer(1, 10, 100, 200, 50, 60, 0, 0, 32, 32, 0)

	// ダーティフラグをクリア
	layer.SetDirty(false)

	// ソース領域を変更
	layer.SetSourceRect(10, 20, 64, 64)

	srcX, srcY, width, height := layer.GetSourceRect()
	if srcX != 10 || srcY != 20 || width != 64 || height != 64 {
		t.Errorf("expected source rect (10, 20, 64, 64), got (%d, %d, %d, %d)", srcX, srcY, width, height)
	}

	if !layer.IsDirty() {
		t.Error("expected layer to be dirty after SetSourceRect")
	}

	// 境界ボックスも更新されていることを確認
	expectedBounds := image.Rect(50, 60, 114, 124)
	if layer.GetBounds() != expectedBounds {
		t.Errorf("expected bounds %v, got %v", expectedBounds, layer.GetBounds())
	}

	// 同じ値を設定（変更なし）
	layer.SetDirty(false)
	layer.SetSourceRect(10, 20, 64, 64)
	if layer.IsDirty() {
		t.Error("expected layer to not be dirty when source rect unchanged")
	}
}

// TestCastLayerInvalidate はキャッシュ無効化をテストする
func TestCastLayerInvalidate(t *testing.T) {
	layer := NewCastLayer(1, 10, 100, 200, 0, 0, 0, 0, 32, 32, 0)

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

// TestCastLayerVisibility は可視性設定をテストする
func TestCastLayerVisibility(t *testing.T) {
	layer := NewCastLayer(1, 10, 100, 200, 0, 0, 0, 0, 32, 32, 0)

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

// TestCastLayerGetImage はGetImageをテストする
func TestCastLayerGetImage(t *testing.T) {
	t.Run("without source image", func(t *testing.T) {
		layer := NewCastLayer(1, 10, 100, 200, 0, 0, 0, 0, 32, 32, 0)

		// ソース画像がない場合はnilを返す
		img := layer.GetImage()
		if img != nil {
			t.Error("expected nil image without source image")
		}
	})

	t.Run("with source image", func(t *testing.T) {
		layer := NewCastLayer(1, 10, 100, 200, 0, 0, 0, 0, 32, 32, 0)

		// ソース画像を設定
		srcImg := ebiten.NewImage(64, 64)
		srcImg.Fill(color.RGBA{255, 0, 0, 255})
		layer.SetSourceImage(srcImg)

		// キャッシュが生成されていることを確認
		img := layer.GetImage()
		if img == nil {
			t.Error("expected non-nil image with source image")
		}

		// サイズが正しいことを確認
		bounds := img.Bounds()
		if bounds.Dx() != 32 || bounds.Dy() != 32 {
			t.Errorf("expected image size (32, 32), got (%d, %d)", bounds.Dx(), bounds.Dy())
		}
	})
}

// TestCastLayerTransparencyProcessing は透明色処理をテストする
// Note: EbitenのReadPixelsはゲームループが開始する前には呼び出せないため、
// ピクセル値の検証は統合テストで行う
func TestCastLayerTransparencyProcessing(t *testing.T) {
	// ソース画像を作成
	srcImg := ebiten.NewImage(32, 32)

	// 緑を透明色として設定
	transColor := color.RGBA{0, 255, 0, 255}
	layer := NewCastLayerWithTransColor(1, 10, 100, 200, 0, 0, 0, 0, 32, 32, 0, transColor)
	layer.SetSourceImage(srcImg)

	// キャッシュが生成されることを確認
	img := layer.GetImage()
	if img == nil {
		t.Fatal("expected non-nil image")
	}

	// 画像サイズが正しいことを確認
	bounds := img.Bounds()
	if bounds.Dx() != 32 || bounds.Dy() != 32 {
		t.Errorf("expected image size (32, 32), got (%d, %d)", bounds.Dx(), bounds.Dy())
	}

	// Note: 実際のピクセル値の検証はゲームループが必要なため、
	// 統合テストで行う
	t.Log("Transparency pixel value test skipped - requires running game loop")
}

// TestCastLayerCacheReuse はキャッシュの再利用をテストする
func TestCastLayerCacheReuse(t *testing.T) {
	layer := NewCastLayer(1, 10, 100, 200, 0, 0, 0, 0, 32, 32, 0)

	// ソース画像を設定
	srcImg := ebiten.NewImage(64, 64)
	srcImg.Fill(color.RGBA{255, 0, 0, 255})
	layer.SetSourceImage(srcImg)

	// 最初のGetImage呼び出し
	img1 := layer.GetImage()
	if img1 == nil {
		t.Fatal("expected non-nil image")
	}

	// ダーティフラグをクリア
	layer.SetDirty(false)

	// 2回目のGetImage呼び出し（キャッシュが再利用されるべき）
	img2 := layer.GetImage()
	if img2 != img1 {
		t.Error("expected same image instance (cache reuse)")
	}
}

// TestCastLayerUpdateFromCast はCast構造体からの更新をテストする
func TestCastLayerUpdateFromCast(t *testing.T) {
	layer := NewCastLayer(1, 10, 100, 200, 0, 0, 0, 0, 32, 32, 0)

	// ダーティフラグをクリア
	layer.SetDirty(false)

	// Cast構造体を作成
	cast := &Cast{
		ID:            10,
		WinID:         100,
		PicID:         200,
		X:             50,
		Y:             60,
		SrcX:          10,
		SrcY:          20,
		Width:         64,
		Height:        64,
		Visible:       true,
		ZOrder:        0,
		TransColor:    color.RGBA{255, 0, 255, 255},
		HasTransColor: true,
	}

	// 更新
	layer.UpdateFromCast(cast)

	// 位置が更新されていることを確認
	x, y := layer.GetPosition()
	if x != 50 || y != 60 {
		t.Errorf("expected position (50, 60), got (%d, %d)", x, y)
	}

	// ソース領域が更新されていることを確認
	srcX, srcY, width, height := layer.GetSourceRect()
	if srcX != 10 || srcY != 20 || width != 64 || height != 64 {
		t.Errorf("expected source rect (10, 20, 64, 64), got (%d, %d, %d, %d)", srcX, srcY, width, height)
	}

	// 透明色が更新されていることを確認
	if !layer.HasTransColor() {
		t.Error("expected layer to have trans color")
	}

	// ダーティフラグが設定されていることを確認
	if !layer.IsDirty() {
		t.Error("expected layer to be dirty after UpdateFromCast")
	}

	// nilのCastでパニックしないことを確認
	layer.UpdateFromCast(nil)
}

// TestCastLayerImplementsLayerInterface はLayerインターフェースの実装をテストする
func TestCastLayerImplementsLayerInterface(t *testing.T) {
	var _ Layer = (*CastLayer)(nil)
}

// TestCastLayerZOrderRelationship はZ順序の関係をテストする
// 要件 1.6: 背景 → 描画 → キャスト → テキストの順序
func TestCastLayerZOrderRelationship(t *testing.T) {
	bgLayer := NewBackgroundLayer(1, 100, nil)
	drawingLayer := NewDrawingLayer(2, 100, 640, 480)
	castLayer := NewCastLayer(3, 10, 100, 200, 0, 0, 0, 0, 32, 32, 0)

	// キャストレイヤーは背景レイヤーより前面にあることを確認
	if castLayer.GetZOrder() <= bgLayer.GetZOrder() {
		t.Errorf("expected CastLayer ZOrder > BackgroundLayer ZOrder, got %d <= %d",
			castLayer.GetZOrder(), bgLayer.GetZOrder())
	}

	// キャストレイヤーは描画レイヤーより前面にあることを確認
	if castLayer.GetZOrder() <= drawingLayer.GetZOrder() {
		t.Errorf("expected CastLayer ZOrder > DrawingLayer ZOrder, got %d <= %d",
			castLayer.GetZOrder(), drawingLayer.GetZOrder())
	}

	// キャストレイヤーはテキストレイヤーより背面にあることを確認
	if castLayer.GetZOrder() >= ZOrderTextBase {
		t.Errorf("expected CastLayer ZOrder < ZOrderTextBase, got %d >= %d",
			castLayer.GetZOrder(), ZOrderTextBase)
	}
}

// TestCastLayerGetSize はサイズ取得をテストする
func TestCastLayerGetSize(t *testing.T) {
	layer := NewCastLayer(1, 10, 100, 200, 0, 0, 0, 0, 64, 48, 0)

	width, height := layer.GetSize()
	if width != 64 || height != 48 {
		t.Errorf("expected size (64, 48), got (%d, %d)", width, height)
	}
}

// TestCastLayerSetPicID はピクチャーID設定をテストする
func TestCastLayerSetPicID(t *testing.T) {
	layer := NewCastLayer(1, 10, 100, 200, 0, 0, 0, 0, 32, 32, 0)

	if layer.GetPicID() != 100 {
		t.Errorf("expected PicID 100, got %d", layer.GetPicID())
	}

	layer.SetPicID(300)
	if layer.GetPicID() != 300 {
		t.Errorf("expected PicID 300 after SetPicID, got %d", layer.GetPicID())
	}
}

// TestColorEqual はcolorEqual関数をテストする
func TestColorEqual(t *testing.T) {
	t.Run("both nil", func(t *testing.T) {
		if !colorEqual(nil, nil) {
			t.Error("expected nil == nil")
		}
	})

	t.Run("one nil", func(t *testing.T) {
		c := color.RGBA{255, 0, 0, 255}
		if colorEqual(c, nil) {
			t.Error("expected color != nil")
		}
		if colorEqual(nil, c) {
			t.Error("expected nil != color")
		}
	})

	t.Run("same color", func(t *testing.T) {
		c1 := color.RGBA{255, 128, 64, 255}
		c2 := color.RGBA{255, 128, 64, 255}
		if !colorEqual(c1, c2) {
			t.Error("expected same colors to be equal")
		}
	})

	t.Run("different color", func(t *testing.T) {
		c1 := color.RGBA{255, 0, 0, 255}
		c2 := color.RGBA{0, 255, 0, 255}
		if colorEqual(c1, c2) {
			t.Error("expected different colors to not be equal")
		}
	})
}

// TestCastLayerSourceImageOutOfBounds はソース領域がソース画像の範囲外の場合をテストする
func TestCastLayerSourceImageOutOfBounds(t *testing.T) {
	// 小さいソース画像を作成
	srcImg := ebiten.NewImage(16, 16)
	srcImg.Fill(color.RGBA{255, 0, 0, 255})

	// ソース領域がソース画像より大きい場合
	layer := NewCastLayer(1, 10, 100, 200, 0, 0, 0, 0, 32, 32, 0)
	layer.SetSourceImage(srcImg)

	img := layer.GetImage()
	if img == nil {
		t.Fatal("expected non-nil image")
	}

	// 実際のサイズはソース画像のサイズに制限される
	bounds := img.Bounds()
	if bounds.Dx() != 16 || bounds.Dy() != 16 {
		t.Errorf("expected image size (16, 16), got (%d, %d)", bounds.Dx(), bounds.Dy())
	}
}

// TestCastLayerZeroSize はゼロサイズのレイヤーをテストする
// 要件 10.3: レイヤー作成に失敗したときにエラーをログに記録し、nilを返す
func TestCastLayerZeroSize(t *testing.T) {
	// ゼロサイズの場合はnilが返される（要件 10.3）
	layer := NewCastLayer(1, 10, 100, 200, 0, 0, 0, 0, 0, 0, 0)
	if layer != nil {
		t.Error("expected nil for zero size layer")
	}
}
