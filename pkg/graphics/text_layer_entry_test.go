package graphics

import (
	"image"
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// TestNewTextLayerEntry は新しいテキストレイヤーエントリの作成をテストする
func TestNewTextLayerEntry(t *testing.T) {
	t.Run("basic creation", func(t *testing.T) {
		entry := NewTextLayerEntry(1, 100, 50, 60, "Hello", 0)

		if entry.GetID() != 1 {
			t.Errorf("expected ID 1, got %d", entry.GetID())
		}
		if entry.GetPicID() != 100 {
			t.Errorf("expected PicID 100, got %d", entry.GetPicID())
		}
		if entry.GetZOrder() != ZOrderTextBase {
			t.Errorf("expected ZOrder %d, got %d", ZOrderTextBase, entry.GetZOrder())
		}
		if !entry.IsVisible() {
			t.Error("expected entry to be visible")
		}
		if !entry.IsDirty() {
			t.Error("expected entry to be dirty on creation")
		}

		x, y := entry.GetPosition()
		if x != 50 || y != 60 {
			t.Errorf("expected position (50, 60), got (%d, %d)", x, y)
		}

		if entry.GetText() != "Hello" {
			t.Errorf("expected text 'Hello', got '%s'", entry.GetText())
		}

		// 画像がない場合、境界は空
		if !entry.GetBounds().Empty() {
			t.Errorf("expected empty bounds without image, got %v", entry.GetBounds())
		}
	})

	t.Run("with zOrder offset", func(t *testing.T) {
		entry := NewTextLayerEntry(2, 100, 0, 0, "Test", 5)

		expectedZOrder := ZOrderTextBase + 5
		if entry.GetZOrder() != expectedZOrder {
			t.Errorf("expected ZOrder %d, got %d", expectedZOrder, entry.GetZOrder())
		}
	})
}

// TestNewTextLayerEntryWithImage は画像付きテキストレイヤーエントリの作成をテストする
func TestNewTextLayerEntryWithImage(t *testing.T) {
	t.Run("with image", func(t *testing.T) {
		img := ebiten.NewImage(100, 30)
		entry := NewTextLayerEntryWithImage(1, 100, 50, 60, "Hello", img, 0)

		if entry.GetID() != 1 {
			t.Errorf("expected ID 1, got %d", entry.GetID())
		}
		if entry.GetPicID() != 100 {
			t.Errorf("expected PicID 100, got %d", entry.GetPicID())
		}
		if entry.IsDirty() {
			t.Error("expected entry to not be dirty when created with image")
		}

		// 境界が正しく設定されていることを確認
		expectedBounds := image.Rect(50, 60, 150, 90)
		if entry.GetBounds() != expectedBounds {
			t.Errorf("expected bounds %v, got %v", expectedBounds, entry.GetBounds())
		}

		// 画像が設定されていることを確認
		if entry.GetImage() != img {
			t.Error("expected same image instance")
		}
	})

	t.Run("with nil image", func(t *testing.T) {
		entry := NewTextLayerEntryWithImage(1, 100, 50, 60, "Hello", nil, 0)

		if entry.GetBounds().Empty() != true {
			t.Errorf("expected empty bounds with nil image, got %v", entry.GetBounds())
		}
		if entry.GetImage() != nil {
			t.Error("expected nil image")
		}
	})
}

// TestTextLayerEntryZOrderStartsAt1000 はZ順序が1000から開始することをテストする
func TestTextLayerEntryZOrderStartsAt1000(t *testing.T) {
	entry := NewTextLayerEntry(1, 100, 0, 0, "Test", 0)

	// Z順序が1000であることを確認
	if entry.GetZOrder() != 1000 {
		t.Errorf("expected ZOrder 1000, got %d", entry.GetZOrder())
	}

	// ZOrderTextBaseが1000であることを確認
	if ZOrderTextBase != 1000 {
		t.Errorf("expected ZOrderTextBase to be 1000, got %d", ZOrderTextBase)
	}

	// キャストレイヤーより大きいことを確認
	if entry.GetZOrder() <= ZOrderCastMax {
		t.Errorf("expected TextLayerEntry ZOrder > ZOrderCastMax, got %d <= %d", entry.GetZOrder(), ZOrderCastMax)
	}

	// 描画レイヤーより大きいことを確認
	if entry.GetZOrder() <= ZOrderDrawing {
		t.Errorf("expected TextLayerEntry ZOrder > ZOrderDrawing, got %d <= %d", entry.GetZOrder(), ZOrderDrawing)
	}
}

// TestTextLayerEntrySetPosition は位置設定をテストする
func TestTextLayerEntrySetPosition(t *testing.T) {
	entry := NewTextLayerEntry(1, 100, 0, 0, "Test", 0)

	// 画像を設定
	img := ebiten.NewImage(100, 30)
	entry.SetImage(img)

	// ダーティフラグをクリア
	entry.SetDirty(false)

	// 位置を変更
	entry.SetPosition(100, 150)

	x, y := entry.GetPosition()
	if x != 100 || y != 150 {
		t.Errorf("expected position (100, 150), got (%d, %d)", x, y)
	}

	if !entry.IsDirty() {
		t.Error("expected entry to be dirty after SetPosition")
	}

	// 境界ボックスも更新されていることを確認
	expectedBounds := image.Rect(100, 150, 200, 180)
	if entry.GetBounds() != expectedBounds {
		t.Errorf("expected bounds %v, got %v", expectedBounds, entry.GetBounds())
	}

	// 同じ位置を設定（変更なし）
	entry.SetDirty(false)
	entry.SetPosition(100, 150)
	if entry.IsDirty() {
		t.Error("expected entry to not be dirty when position unchanged")
	}
}

// TestTextLayerEntrySetText はテキスト設定をテストする
func TestTextLayerEntrySetText(t *testing.T) {
	entry := NewTextLayerEntry(1, 100, 0, 0, "Hello", 0)

	// 画像を設定
	img := ebiten.NewImage(100, 30)
	entry.SetImage(img)

	// ダーティフラグをクリア
	entry.SetDirty(false)

	// テキストを変更
	entry.SetText("World")

	if entry.GetText() != "World" {
		t.Errorf("expected text 'World', got '%s'", entry.GetText())
	}

	if !entry.IsDirty() {
		t.Error("expected entry to be dirty after SetText")
	}

	// キャッシュが無効化されていることを確認
	if entry.GetImage() != nil {
		t.Error("expected image to be nil after SetText")
	}

	// 同じテキストを設定（変更なし）
	entry.SetDirty(false)
	entry.SetText("World")
	if entry.IsDirty() {
		t.Error("expected entry to not be dirty when text unchanged")
	}
}

// TestTextLayerEntrySetImage は画像設定をテストする
func TestTextLayerEntrySetImage(t *testing.T) {
	entry := NewTextLayerEntry(1, 100, 50, 60, "Test", 0)

	// 初期状態は画像なし
	if entry.GetImage() != nil {
		t.Error("expected nil image initially")
	}

	// 画像を設定
	img := ebiten.NewImage(100, 30)
	entry.SetImage(img)

	if entry.GetImage() != img {
		t.Error("expected same image instance")
	}

	// 境界が更新されていることを確認
	expectedBounds := image.Rect(50, 60, 150, 90)
	if entry.GetBounds() != expectedBounds {
		t.Errorf("expected bounds %v, got %v", expectedBounds, entry.GetBounds())
	}

	// ダーティフラグがクリアされていることを確認
	if entry.IsDirty() {
		t.Error("expected entry to not be dirty after SetImage")
	}

	// nilを設定
	entry.SetImage(nil)
	if entry.GetImage() != nil {
		t.Error("expected nil image after SetImage(nil)")
	}
	if !entry.GetBounds().Empty() {
		t.Errorf("expected empty bounds after SetImage(nil), got %v", entry.GetBounds())
	}
}

// TestTextLayerEntryInvalidate はキャッシュ無効化をテストする
func TestTextLayerEntryInvalidate(t *testing.T) {
	entry := NewTextLayerEntry(1, 100, 0, 0, "Test", 0)

	// 画像を設定
	img := ebiten.NewImage(100, 30)
	entry.SetImage(img)

	// ダーティフラグをクリア
	entry.SetDirty(false)
	if entry.IsDirty() {
		t.Error("expected entry to not be dirty after SetDirty(false)")
	}

	// Invalidateを呼び出す
	entry.Invalidate()
	if !entry.IsDirty() {
		t.Error("expected entry to be dirty after Invalidate")
	}
	if entry.GetImage() != nil {
		t.Error("expected image to be nil after Invalidate")
	}
}

// TestTextLayerEntryVisibility は可視性設定をテストする
func TestTextLayerEntryVisibility(t *testing.T) {
	entry := NewTextLayerEntry(1, 100, 0, 0, "Test", 0)

	// 初期状態は可視
	if !entry.IsVisible() {
		t.Error("expected entry to be visible initially")
	}

	// ダーティフラグをクリア
	entry.SetDirty(false)

	// 非表示に設定
	entry.SetVisible(false)
	if entry.IsVisible() {
		t.Error("expected entry to be invisible after SetVisible(false)")
	}
	if !entry.IsDirty() {
		t.Error("expected entry to be dirty after visibility change")
	}

	// ダーティフラグをクリア
	entry.SetDirty(false)

	// 同じ値を設定（変更なし）
	entry.SetVisible(false)
	if entry.IsDirty() {
		t.Error("expected entry to not be dirty when visibility unchanged")
	}
}

// TestTextLayerEntryGetSize はサイズ取得をテストする
func TestTextLayerEntryGetSize(t *testing.T) {
	t.Run("without image", func(t *testing.T) {
		entry := NewTextLayerEntry(1, 100, 0, 0, "Test", 0)

		width, height := entry.GetSize()
		if width != 0 || height != 0 {
			t.Errorf("expected size (0, 0) without image, got (%d, %d)", width, height)
		}
	})

	t.Run("with image", func(t *testing.T) {
		entry := NewTextLayerEntry(1, 100, 0, 0, "Test", 0)

		img := ebiten.NewImage(100, 30)
		entry.SetImage(img)

		width, height := entry.GetSize()
		if width != 100 || height != 30 {
			t.Errorf("expected size (100, 30), got (%d, %d)", width, height)
		}
	})
}

// TestTextLayerEntryHasImage はHasImageをテストする
func TestTextLayerEntryHasImage(t *testing.T) {
	entry := NewTextLayerEntry(1, 100, 0, 0, "Test", 0)

	if entry.HasImage() {
		t.Error("expected HasImage to be false initially")
	}

	img := ebiten.NewImage(100, 30)
	entry.SetImage(img)

	if !entry.HasImage() {
		t.Error("expected HasImage to be true after SetImage")
	}

	entry.SetImage(nil)
	if entry.HasImage() {
		t.Error("expected HasImage to be false after SetImage(nil)")
	}
}

// TestTextLayerEntrySetPicID はピクチャーID設定をテストする
func TestTextLayerEntrySetPicID(t *testing.T) {
	entry := NewTextLayerEntry(1, 100, 0, 0, "Test", 0)

	if entry.GetPicID() != 100 {
		t.Errorf("expected PicID 100, got %d", entry.GetPicID())
	}

	entry.SetPicID(300)
	if entry.GetPicID() != 300 {
		t.Errorf("expected PicID 300 after SetPicID, got %d", entry.GetPicID())
	}
}

// TestTextLayerEntryImplementsLayerInterface はLayerインターフェースの実装をテストする
func TestTextLayerEntryImplementsLayerInterface(t *testing.T) {
	var _ Layer = (*TextLayerEntry)(nil)
}

// TestTextLayerEntryZOrderRelationship はZ順序の関係をテストする
// 要件 1.6: 背景 → 描画 → キャスト → テキストの順序
func TestTextLayerEntryZOrderRelationship(t *testing.T) {
	textEntry := NewTextLayerEntry(1, 100, 0, 0, "Test", 0)

	// テキストレイヤーは背景レイヤー（ZOrderBackground=0）より前面にあることを確認
	if textEntry.GetZOrder() <= ZOrderBackground {
		t.Errorf("expected TextLayerEntry ZOrder > ZOrderBackground, got %d <= %d",
			textEntry.GetZOrder(), ZOrderBackground)
	}

	// テキストレイヤーは描画レイヤー（ZOrderDrawing=100）より前面にあることを確認
	if textEntry.GetZOrder() <= ZOrderDrawing {
		t.Errorf("expected TextLayerEntry ZOrder > ZOrderDrawing, got %d <= %d",
			textEntry.GetZOrder(), ZOrderDrawing)
	}

	// テキストレイヤーはキャストレイヤー（ZOrderCastMax=999）より前面にあることを確認
	if textEntry.GetZOrder() <= ZOrderCastMax {
		t.Errorf("expected TextLayerEntry ZOrder > ZOrderCastMax, got %d <= %d",
			textEntry.GetZOrder(), ZOrderCastMax)
	}
}

// TestNewTextLayerEntryFromTextLayer は既存のTextLayerからの変換をテストする
func TestNewTextLayerEntryFromTextLayer(t *testing.T) {
	t.Run("with valid TextLayer", func(t *testing.T) {
		// TextLayerを作成
		textLayer := &TextLayer{
			Image:  image.NewRGBA(image.Rect(0, 0, 100, 30)),
			PicID:  100,
			X:      50,
			Y:      60,
			Width:  100,
			Height: 30,
		}

		entry := NewTextLayerEntryFromTextLayer(1, textLayer, 0)

		if entry == nil {
			t.Fatal("expected non-nil entry")
		}
		if entry.GetID() != 1 {
			t.Errorf("expected ID 1, got %d", entry.GetID())
		}
		if entry.GetPicID() != 100 {
			t.Errorf("expected PicID 100, got %d", entry.GetPicID())
		}

		x, y := entry.GetPosition()
		if x != 50 || y != 60 {
			t.Errorf("expected position (50, 60), got (%d, %d)", x, y)
		}

		if entry.GetImage() == nil {
			t.Error("expected non-nil image")
		}
	})

	t.Run("with nil TextLayer", func(t *testing.T) {
		entry := NewTextLayerEntryFromTextLayer(1, nil, 0)

		if entry != nil {
			t.Error("expected nil entry for nil TextLayer")
		}
	})

	t.Run("with nil image in TextLayer", func(t *testing.T) {
		textLayer := &TextLayer{
			Image:  nil,
			PicID:  100,
			X:      50,
			Y:      60,
			Width:  100,
			Height: 30,
		}

		entry := NewTextLayerEntryFromTextLayer(1, textLayer, 0)

		if entry == nil {
			t.Fatal("expected non-nil entry")
		}
		if entry.GetImage() != nil {
			t.Error("expected nil image")
		}
	})
}

// TestTextLayerEntryCacheReuse はキャッシュの再利用をテストする
func TestTextLayerEntryCacheReuse(t *testing.T) {
	entry := NewTextLayerEntry(1, 100, 0, 0, "Test", 0)

	// 画像を設定
	img := ebiten.NewImage(100, 30)
	img.Fill(color.RGBA{255, 0, 0, 255})
	entry.SetImage(img)

	// 最初のGetImage呼び出し
	img1 := entry.GetImage()
	if img1 == nil {
		t.Fatal("expected non-nil image")
	}

	// 2回目のGetImage呼び出し（キャッシュが再利用されるべき）
	img2 := entry.GetImage()
	if img2 != img1 {
		t.Error("expected same image instance (cache reuse)")
	}
}

// TestTextLayerEntryPositionWithoutImage は画像なしでの位置変更をテストする
func TestTextLayerEntryPositionWithoutImage(t *testing.T) {
	entry := NewTextLayerEntry(1, 100, 0, 0, "Test", 0)

	// ダーティフラグをクリア
	entry.SetDirty(false)

	// 画像なしで位置を変更
	entry.SetPosition(100, 150)

	x, y := entry.GetPosition()
	if x != 100 || y != 150 {
		t.Errorf("expected position (100, 150), got (%d, %d)", x, y)
	}

	if !entry.IsDirty() {
		t.Error("expected entry to be dirty after SetPosition")
	}

	// 画像がないので境界は空のまま
	// （境界は画像のサイズに依存するため）
}

// TestTextLayerEntryMultipleZOrderOffsets は複数のZ順序オフセットをテストする
func TestTextLayerEntryMultipleZOrderOffsets(t *testing.T) {
	entries := make([]*TextLayerEntry, 5)
	for i := 0; i < 5; i++ {
		entries[i] = NewTextLayerEntry(i, 100, 0, 0, "Test", i)
	}

	// 各エントリのZ順序が正しいことを確認
	for i, entry := range entries {
		expectedZOrder := ZOrderTextBase + i
		if entry.GetZOrder() != expectedZOrder {
			t.Errorf("entry %d: expected ZOrder %d, got %d", i, expectedZOrder, entry.GetZOrder())
		}
	}

	// Z順序が昇順であることを確認
	for i := 1; i < len(entries); i++ {
		if entries[i].GetZOrder() <= entries[i-1].GetZOrder() {
			t.Errorf("expected entries[%d].ZOrder > entries[%d].ZOrder, got %d <= %d",
				i, i-1, entries[i].GetZOrder(), entries[i-1].GetZOrder())
		}
	}
}
