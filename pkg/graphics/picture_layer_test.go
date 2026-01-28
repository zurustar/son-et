package graphics

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// TestPictureLayerImplementsLayer はPictureLayerがLayerインターフェースを実装していることを確認する
func TestPictureLayerImplementsLayer(t *testing.T) {
	// コンパイル時にLayerインターフェースを実装していることを確認
	var _ Layer = (*PictureLayer)(nil)
}

// TestNewPictureLayer は新しいPictureLayerの作成をテストする
func TestNewPictureLayer(t *testing.T) {
	layer := NewPictureLayer(1, 640, 480)

	if layer == nil {
		t.Fatal("NewPictureLayer returned nil")
	}

	// IDの確認
	if layer.GetID() != 1 {
		t.Errorf("Expected ID 1, got %d", layer.GetID())
	}

	// 境界の確認
	bounds := layer.GetBounds()
	if bounds.Dx() != 640 || bounds.Dy() != 480 {
		t.Errorf("Expected bounds 640x480, got %dx%d", bounds.Dx(), bounds.Dy())
	}

	// 可視性の確認
	if !layer.IsVisible() {
		t.Error("Expected layer to be visible")
	}

	// ダーティフラグの確認（新規作成時はtrue）
	if !layer.IsDirty() {
		t.Error("Expected layer to be dirty after creation")
	}

	// 焼き付け可能フラグの確認
	if !layer.IsBakeable() {
		t.Error("Expected layer to be bakeable")
	}

	// 画像の確認
	img := layer.GetImage()
	if img == nil {
		t.Error("Expected image to be non-nil")
	}

	// 画像サイズの確認
	width, height := layer.GetSize()
	if width != 640 || height != 480 {
		t.Errorf("Expected size 640x480, got %dx%d", width, height)
	}
}

// TestPictureLayerBake は焼き付け機能をテストする
func TestPictureLayerBake(t *testing.T) {
	layer := NewPictureLayer(1, 640, 480)

	// ダーティフラグをクリア
	layer.SetDirty(false)

	// 小さなソース画像を作成
	src := ebiten.NewImage(100, 100)

	// 焼き付け
	layer.Bake(src, 10, 20)

	// 焼き付け後、ダーティフラグが設定されていることを確認
	if !layer.IsDirty() {
		t.Error("Expected layer to be dirty after baking")
	}
}

// TestPictureLayerClear はクリア機能をテストする
func TestPictureLayerClear(t *testing.T) {
	layer := NewPictureLayer(1, 640, 480)

	// ダーティフラグをクリア
	layer.SetDirty(false)

	// クリア
	layer.Clear()

	// クリア後、ダーティフラグが設定されていることを確認
	if !layer.IsDirty() {
		t.Error("Expected layer to be dirty after clearing")
	}
}

// TestPictureLayerInvalidate はキャッシュ無効化をテストする
func TestPictureLayerInvalidate(t *testing.T) {
	layer := NewPictureLayer(1, 640, 480)

	// ダーティフラグをクリア
	layer.SetDirty(false)

	// 無効化
	layer.Invalidate()

	// 無効化後、ダーティフラグが設定されていることを確認
	if !layer.IsDirty() {
		t.Error("Expected layer to be dirty after invalidation")
	}
}

// TestPictureLayerZOrder はZ順序の設定をテストする
func TestPictureLayerZOrder(t *testing.T) {
	layer := NewPictureLayer(1, 640, 480)

	// 初期Z順序は0
	if layer.GetZOrder() != 0 {
		t.Errorf("Expected initial Z order 0, got %d", layer.GetZOrder())
	}

	// Z順序を設定
	layer.SetZOrder(100)

	if layer.GetZOrder() != 100 {
		t.Errorf("Expected Z order 100, got %d", layer.GetZOrder())
	}
}

// TestPictureLayerIsOpaque は不透明度の確認をテストする
func TestPictureLayerIsOpaque(t *testing.T) {
	layer := NewPictureLayer(1, 640, 480)

	// PictureLayerは透明画像なので、デフォルトは不透明ではない
	if layer.IsOpaque() {
		t.Error("Expected layer to not be opaque by default")
	}
}

// TestPictureLayerGetLayerType はレイヤータイプの取得をテストする
// 要件 2.4: レイヤーが作成されたとき、レイヤータイプを識別可能にする
func TestPictureLayerGetLayerType(t *testing.T) {
	layer := NewPictureLayer(1, 640, 480)

	// PictureLayerはLayerTypePictureを返す
	if layer.GetLayerType() != LayerTypePicture {
		t.Errorf("Expected LayerTypePicture, got %v", layer.GetLayerType())
	}
}

// TestAllLayerTypesGetLayerType は全レイヤータイプのGetLayerType()をテストする
// 要件 2.4: レイヤーが作成されたとき、レイヤータイプを識別可能にする
func TestAllLayerTypesGetLayerType(t *testing.T) {
	tests := []struct {
		name     string
		layer    Layer
		expected LayerType
	}{
		{
			name:     "PictureLayer",
			layer:    NewPictureLayer(1, 640, 480),
			expected: LayerTypePicture,
		},
		{
			name:     "CastLayer",
			layer:    NewCastLayer(2, 1, 1, 1, 0, 0, 0, 0, 100, 100, 0),
			expected: LayerTypeCast,
		},
		{
			name:     "TextLayerEntry",
			layer:    NewTextLayerEntry(3, 1, 0, 0, "test", 0),
			expected: LayerTypeText,
		},
		{
			name:     "BackgroundLayer",
			layer:    NewBackgroundLayer(4, 1, nil),
			expected: LayerTypePicture, // BackgroundLayerはPictureLayerの一種
		},
		{
			name:     "DrawingEntry",
			layer:    NewDrawingEntry(5, 1, nil, 0, 0, 100, 100, 0),
			expected: LayerTypePicture, // DrawingEntryはPictureLayerの一種
		},
		{
			name:     "DrawingLayer",
			layer:    NewDrawingLayer(6, 1, 100, 100),
			expected: LayerTypePicture, // DrawingLayerはPictureLayerの一種
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.layer.GetLayerType(); got != tt.expected {
				t.Errorf("%s.GetLayerType() = %v, want %v", tt.name, got, tt.expected)
			}
		})
	}
}

// TestLayerTypeString はLayerTypeのString()メソッドをテストする
func TestLayerTypeString(t *testing.T) {
	tests := []struct {
		layerType LayerType
		expected  string
	}{
		{LayerTypePicture, "Picture"},
		{LayerTypeText, "Text"},
		{LayerTypeCast, "Cast"},
		{LayerType(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.layerType.String(); got != tt.expected {
				t.Errorf("LayerType(%d).String() = %v, want %v", tt.layerType, got, tt.expected)
			}
		})
	}
}

// TestPictureLayerBakeWithOptions はBakeWithOptionsメソッドをテストする
// 要件 2.5: Picture_Layerは焼き付け対象として機能する
func TestPictureLayerBakeWithOptions(t *testing.T) {
	layer := NewPictureLayer(1, 640, 480)

	// ダーティフラグをクリア
	layer.SetDirty(false)

	// 小さなソース画像を作成
	src := ebiten.NewImage(100, 100)

	// オプションを作成
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(50, 50)

	// 焼き付け
	layer.BakeWithOptions(src, op)

	// 焼き付け後、ダーティフラグが設定されていることを確認
	if !layer.IsDirty() {
		t.Error("Expected layer to be dirty after baking with options")
	}
}

// TestPictureLayerMultipleBakes は同じレイヤーへの複数回焼き付けをテストする
// 要件 2.5: Picture_Layerは焼き付け対象として機能する
func TestPictureLayerMultipleBakes(t *testing.T) {
	layer := NewPictureLayer(1, 640, 480)

	// 複数のソース画像を作成
	src1 := ebiten.NewImage(100, 100)
	src2 := ebiten.NewImage(50, 50)
	src3 := ebiten.NewImage(200, 200)

	// 1回目の焼き付け
	layer.SetDirty(false)
	layer.Bake(src1, 0, 0)
	if !layer.IsDirty() {
		t.Error("Expected layer to be dirty after first bake")
	}

	// 2回目の焼き付け
	layer.SetDirty(false)
	layer.Bake(src2, 100, 100)
	if !layer.IsDirty() {
		t.Error("Expected layer to be dirty after second bake")
	}

	// 3回目の焼き付け
	layer.SetDirty(false)
	layer.Bake(src3, 200, 200)
	if !layer.IsDirty() {
		t.Error("Expected layer to be dirty after third bake")
	}

	// レイヤーの画像が有効であることを確認
	img := layer.GetImage()
	if img == nil {
		t.Error("Expected image to be non-nil after multiple bakes")
	}
}

// TestPictureLayerBakeNilSource はnilソース画像での焼き付けをテストする
// エッジケース: nilソースでクラッシュしないことを確認
func TestPictureLayerBakeNilSource(t *testing.T) {
	layer := NewPictureLayer(1, 640, 480)

	// ダーティフラグをクリア
	layer.SetDirty(false)

	// nilソースで焼き付け（クラッシュしないことを確認）
	layer.Bake(nil, 10, 20)

	// nilソースの場合、ダーティフラグは変更されない
	if layer.IsDirty() {
		t.Error("Expected layer to not be dirty after baking nil source")
	}
}

// TestPictureLayerBakeWithOptionsNilSource はnilソース画像でのBakeWithOptionsをテストする
// エッジケース: nilソースでクラッシュしないことを確認
func TestPictureLayerBakeWithOptionsNilSource(t *testing.T) {
	layer := NewPictureLayer(1, 640, 480)

	// ダーティフラグをクリア
	layer.SetDirty(false)

	// オプションを作成
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(50, 50)

	// nilソースで焼き付け（クラッシュしないことを確認）
	layer.BakeWithOptions(nil, op)

	// nilソースの場合、ダーティフラグは変更されない
	if layer.IsDirty() {
		t.Error("Expected layer to not be dirty after baking nil source with options")
	}
}

// TestPictureLayerBakeOutOfBounds は境界外への焼き付けをテストする
// エッジケース: 境界外への焼き付けでクラッシュしないことを確認
func TestPictureLayerBakeOutOfBounds(t *testing.T) {
	layer := NewPictureLayer(1, 640, 480)

	// 小さなソース画像を作成
	src := ebiten.NewImage(100, 100)

	// ダーティフラグをクリア
	layer.SetDirty(false)

	// 負の座標への焼き付け
	layer.Bake(src, -50, -50)
	if !layer.IsDirty() {
		t.Error("Expected layer to be dirty after baking at negative coordinates")
	}

	// ダーティフラグをクリア
	layer.SetDirty(false)

	// 境界外への焼き付け（右下）
	layer.Bake(src, 700, 500)
	if !layer.IsDirty() {
		t.Error("Expected layer to be dirty after baking outside bounds")
	}
}

// TestPictureLayerBakePreservesLayerProperties は焼き付け後もレイヤープロパティが保持されることをテストする
// 要件 2.1: Picture_Layerを定義する
func TestPictureLayerBakePreservesLayerProperties(t *testing.T) {
	layer := NewPictureLayer(1, 640, 480)

	// プロパティを設定
	layer.SetZOrder(50)
	layer.SetVisible(true)

	// 焼き付け
	src := ebiten.NewImage(100, 100)
	layer.Bake(src, 10, 20)

	// プロパティが保持されていることを確認
	if layer.GetID() != 1 {
		t.Errorf("Expected ID 1, got %d", layer.GetID())
	}
	if layer.GetZOrder() != 50 {
		t.Errorf("Expected Z order 50, got %d", layer.GetZOrder())
	}
	if !layer.IsVisible() {
		t.Error("Expected layer to be visible")
	}
	if !layer.IsBakeable() {
		t.Error("Expected layer to be bakeable")
	}
	if layer.GetLayerType() != LayerTypePicture {
		t.Errorf("Expected LayerTypePicture, got %v", layer.GetLayerType())
	}
}

// TestPictureLayerBakeSequence は焼き付けシーケンスをテストする
// 要件 2.5: Picture_Layerは焼き付け対象として機能する
func TestPictureLayerBakeSequence(t *testing.T) {
	layer := NewPictureLayer(1, 640, 480)

	// 複数の焼き付け操作を実行
	for i := 0; i < 10; i++ {
		src := ebiten.NewImage(50, 50)
		layer.SetDirty(false)
		layer.Bake(src, i*50, i*40)

		// 各焼き付け後にダーティフラグが設定されていることを確認
		if !layer.IsDirty() {
			t.Errorf("Expected layer to be dirty after bake %d", i)
		}
	}

	// レイヤーの画像が有効であることを確認
	img := layer.GetImage()
	if img == nil {
		t.Error("Expected image to be non-nil after bake sequence")
	}

	// サイズが変わっていないことを確認
	width, height := layer.GetSize()
	if width != 640 || height != 480 {
		t.Errorf("Expected size 640x480, got %dx%d", width, height)
	}
}
