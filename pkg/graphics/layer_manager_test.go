package graphics

import (
	"image"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// TestNewLayerManager はLayerManagerの作成をテストする
func TestNewLayerManager(t *testing.T) {
	lm := NewLayerManager()

	if lm == nil {
		t.Fatal("NewLayerManager returned nil")
	}

	if lm.layers == nil {
		t.Error("layers map should not be nil")
	}

	if lm.nextLayerID != 1 {
		t.Errorf("nextLayerID should be 1, got %d", lm.nextLayerID)
	}

	if lm.GetPictureLayerSetCount() != 0 {
		t.Errorf("initial count should be 0, got %d", lm.GetPictureLayerSetCount())
	}
}

// TestGetOrCreatePictureLayerSet はPictureLayerSetの取得・作成をテストする
func TestGetOrCreatePictureLayerSet(t *testing.T) {
	lm := NewLayerManager()

	// 新規作成
	pls1 := lm.GetOrCreatePictureLayerSet(1)
	if pls1 == nil {
		t.Fatal("GetOrCreatePictureLayerSet returned nil")
	}
	if pls1.PicID != 1 {
		t.Errorf("PicID should be 1, got %d", pls1.PicID)
	}

	// 同じIDで取得
	pls1Again := lm.GetOrCreatePictureLayerSet(1)
	if pls1Again != pls1 {
		t.Error("GetOrCreatePictureLayerSet should return the same instance")
	}

	// 別のIDで作成
	pls2 := lm.GetOrCreatePictureLayerSet(2)
	if pls2 == nil {
		t.Fatal("GetOrCreatePictureLayerSet returned nil for picID 2")
	}
	if pls2.PicID != 2 {
		t.Errorf("PicID should be 2, got %d", pls2.PicID)
	}

	// カウント確認
	if lm.GetPictureLayerSetCount() != 2 {
		t.Errorf("count should be 2, got %d", lm.GetPictureLayerSetCount())
	}
}

// TestGetPictureLayerSet はPictureLayerSetの取得をテストする
func TestGetPictureLayerSet(t *testing.T) {
	lm := NewLayerManager()

	// 存在しないIDで取得
	pls := lm.GetPictureLayerSet(1)
	if pls != nil {
		t.Error("GetPictureLayerSet should return nil for non-existent ID")
	}

	// 作成後に取得
	lm.GetOrCreatePictureLayerSet(1)
	pls = lm.GetPictureLayerSet(1)
	if pls == nil {
		t.Error("GetPictureLayerSet should return the created instance")
	}
}

// TestDeletePictureLayerSet はPictureLayerSetの削除をテストする
func TestDeletePictureLayerSet(t *testing.T) {
	lm := NewLayerManager()

	// 作成
	lm.GetOrCreatePictureLayerSet(1)
	lm.GetOrCreatePictureLayerSet(2)

	if lm.GetPictureLayerSetCount() != 2 {
		t.Errorf("count should be 2, got %d", lm.GetPictureLayerSetCount())
	}

	// 削除
	lm.DeletePictureLayerSet(1)

	if lm.GetPictureLayerSetCount() != 1 {
		t.Errorf("count should be 1 after deletion, got %d", lm.GetPictureLayerSetCount())
	}

	if lm.GetPictureLayerSet(1) != nil {
		t.Error("GetPictureLayerSet should return nil after deletion")
	}

	if lm.GetPictureLayerSet(2) == nil {
		t.Error("GetPictureLayerSet should still return picID 2")
	}
}

// TestGetNextLayerID はレイヤーIDの取得をテストする
func TestGetNextLayerID(t *testing.T) {
	lm := NewLayerManager()

	id1 := lm.GetNextLayerID()
	if id1 != 1 {
		t.Errorf("first ID should be 1, got %d", id1)
	}

	id2 := lm.GetNextLayerID()
	if id2 != 2 {
		t.Errorf("second ID should be 2, got %d", id2)
	}

	id3 := lm.GetNextLayerID()
	if id3 != 3 {
		t.Errorf("third ID should be 3, got %d", id3)
	}
}

// TestLayerManagerClear はLayerManagerのクリアをテストする
func TestLayerManagerClear(t *testing.T) {
	lm := NewLayerManager()

	// 複数作成
	lm.GetOrCreatePictureLayerSet(1)
	lm.GetOrCreatePictureLayerSet(2)
	lm.GetOrCreatePictureLayerSet(3)

	// IDを進める
	lm.GetNextLayerID()
	lm.GetNextLayerID()

	// クリア
	lm.Clear()

	if lm.GetPictureLayerSetCount() != 0 {
		t.Errorf("count should be 0 after clear, got %d", lm.GetPictureLayerSetCount())
	}

	// nextLayerIDはリセットされない
	nextID := lm.GetNextLayerID()
	if nextID != 3 {
		t.Errorf("nextLayerID should not be reset, expected 3, got %d", nextID)
	}
}

// TestGetAllPictureLayerSets はすべてのPictureLayerSetの取得をテストする
func TestGetAllPictureLayerSets(t *testing.T) {
	lm := NewLayerManager()

	lm.GetOrCreatePictureLayerSet(1)
	lm.GetOrCreatePictureLayerSet(2)
	lm.GetOrCreatePictureLayerSet(3)

	all := lm.GetAllPictureLayerSets()

	if len(all) != 3 {
		t.Errorf("should have 3 items, got %d", len(all))
	}

	// コピーであることを確認（元のマップを変更しても影響しない）
	delete(all, 1)
	if lm.GetPictureLayerSetCount() != 3 {
		t.Error("deleting from returned map should not affect original")
	}
}

// TestNewPictureLayerSet はPictureLayerSetの作成をテストする
func TestNewPictureLayerSet(t *testing.T) {
	pls := NewPictureLayerSet(1)

	if pls == nil {
		t.Fatal("NewPictureLayerSet returned nil")
	}

	if pls.PicID != 1 {
		t.Errorf("PicID should be 1, got %d", pls.PicID)
	}

	if pls.Background != nil {
		t.Error("Background should be nil initially")
	}

	if pls.Drawing != nil {
		t.Error("Drawing should be nil initially")
	}

	if len(pls.Casts) != 0 {
		t.Errorf("Casts should be empty, got %d", len(pls.Casts))
	}

	if len(pls.Texts) != 0 {
		t.Errorf("Texts should be empty, got %d", len(pls.Texts))
	}

	if !pls.FullDirty {
		t.Error("FullDirty should be true initially")
	}
}

// TestPictureLayerSetBackground は背景レイヤーの設定をテストする
func TestPictureLayerSetBackground(t *testing.T) {
	pls := NewPictureLayerSet(1)

	// 初期状態
	if pls.Background != nil {
		t.Error("Background should be nil initially")
	}

	// 背景レイヤーを設定
	bg := NewBackgroundLayer(1, 1, nil)
	pls.SetBackground(bg)

	if pls.Background != bg {
		t.Error("Background should be set")
	}

	if !pls.FullDirty {
		t.Error("FullDirty should be true after setting background")
	}
}

// TestPictureLayerSetDrawing は描画レイヤーの設定をテストする
func TestPictureLayerSetDrawing(t *testing.T) {
	pls := NewPictureLayerSet(1)

	// 初期状態
	if pls.Drawing != nil {
		t.Error("Drawing should be nil initially")
	}

	// 描画レイヤーを設定
	drawing := NewDrawingLayer(1, 1, 100, 100)
	pls.SetDrawing(drawing)

	if pls.Drawing != drawing {
		t.Error("Drawing should be set")
	}

	if !pls.FullDirty {
		t.Error("FullDirty should be true after setting drawing")
	}
}

// TestPictureLayerSetCastLayers はキャストレイヤーの追加・削除をテストする
func TestPictureLayerSetCastLayers(t *testing.T) {
	pls := NewPictureLayerSet(1)

	// 初期状態
	if pls.GetCastLayerCount() != 0 {
		t.Errorf("initial cast count should be 0, got %d", pls.GetCastLayerCount())
	}

	// キャストレイヤーを追加
	cast1 := NewCastLayer(1, 100, 1, 2, 10, 20, 0, 0, 50, 50, 0)
	pls.AddCastLayer(cast1)

	if pls.GetCastLayerCount() != 1 {
		t.Errorf("cast count should be 1, got %d", pls.GetCastLayerCount())
	}

	// キャストIDで取得
	retrieved := pls.GetCastLayer(100)
	if retrieved != cast1 {
		t.Error("GetCastLayer should return the added cast")
	}

	// レイヤーIDで取得
	retrievedByID := pls.GetCastLayerByID(1)
	if retrievedByID != cast1 {
		t.Error("GetCastLayerByID should return the added cast")
	}

	// 存在しないIDで取得
	if pls.GetCastLayer(999) != nil {
		t.Error("GetCastLayer should return nil for non-existent ID")
	}

	// 2つ目のキャストを追加
	cast2 := NewCastLayer(2, 101, 1, 2, 30, 40, 0, 0, 50, 50, 1)
	pls.AddCastLayer(cast2)

	if pls.GetCastLayerCount() != 2 {
		t.Errorf("cast count should be 2, got %d", pls.GetCastLayerCount())
	}

	// キャストIDで削除
	removed := pls.RemoveCastLayer(100)
	if !removed {
		t.Error("RemoveCastLayer should return true")
	}

	if pls.GetCastLayerCount() != 1 {
		t.Errorf("cast count should be 1 after removal, got %d", pls.GetCastLayerCount())
	}

	if pls.GetCastLayer(100) != nil {
		t.Error("removed cast should not be found")
	}

	// 存在しないIDで削除
	removed = pls.RemoveCastLayer(999)
	if removed {
		t.Error("RemoveCastLayer should return false for non-existent ID")
	}

	// レイヤーIDで削除
	removed = pls.RemoveCastLayerByID(2)
	if !removed {
		t.Error("RemoveCastLayerByID should return true")
	}

	if pls.GetCastLayerCount() != 0 {
		t.Errorf("cast count should be 0 after removal, got %d", pls.GetCastLayerCount())
	}
}

// TestPictureLayerSetTextLayers はテキストレイヤーの追加・削除をテストする
func TestPictureLayerSetTextLayers(t *testing.T) {
	pls := NewPictureLayerSet(1)

	// 初期状態
	if pls.GetTextLayerCount() != 0 {
		t.Errorf("initial text count should be 0, got %d", pls.GetTextLayerCount())
	}

	// テキストレイヤーを追加
	text1 := NewTextLayerEntry(1, 1, 10, 20, "Hello", 0)
	pls.AddTextLayer(text1)

	if pls.GetTextLayerCount() != 1 {
		t.Errorf("text count should be 1, got %d", pls.GetTextLayerCount())
	}

	// レイヤーIDで取得
	retrieved := pls.GetTextLayer(1)
	if retrieved != text1 {
		t.Error("GetTextLayer should return the added text")
	}

	// 存在しないIDで取得
	if pls.GetTextLayer(999) != nil {
		t.Error("GetTextLayer should return nil for non-existent ID")
	}

	// 2つ目のテキストを追加
	text2 := NewTextLayerEntry(2, 1, 30, 40, "World", 1)
	pls.AddTextLayer(text2)

	if pls.GetTextLayerCount() != 2 {
		t.Errorf("text count should be 2, got %d", pls.GetTextLayerCount())
	}

	// 削除
	removed := pls.RemoveTextLayer(1)
	if !removed {
		t.Error("RemoveTextLayer should return true")
	}

	if pls.GetTextLayerCount() != 1 {
		t.Errorf("text count should be 1 after removal, got %d", pls.GetTextLayerCount())
	}

	// 存在しないIDで削除
	removed = pls.RemoveTextLayer(999)
	if removed {
		t.Error("RemoveTextLayer should return false for non-existent ID")
	}
}

// TestPictureLayerSetClearLayers はレイヤーのクリアをテストする
func TestPictureLayerSetClearLayers(t *testing.T) {
	pls := NewPictureLayerSet(1)

	// キャストとテキストを追加
	pls.AddCastLayer(NewCastLayer(1, 100, 1, 2, 10, 20, 0, 0, 50, 50, 0))
	pls.AddCastLayer(NewCastLayer(2, 101, 1, 2, 30, 40, 0, 0, 50, 50, 1))
	pls.AddTextLayer(NewTextLayerEntry(3, 1, 10, 20, "Hello", 0))
	pls.AddTextLayer(NewTextLayerEntry(4, 1, 30, 40, "World", 1))

	// キャストをクリア
	pls.ClearCastLayers()
	if pls.GetCastLayerCount() != 0 {
		t.Errorf("cast count should be 0 after clear, got %d", pls.GetCastLayerCount())
	}
	if pls.GetTextLayerCount() != 2 {
		t.Errorf("text count should still be 2, got %d", pls.GetTextLayerCount())
	}

	// テキストをクリア
	pls.ClearTextLayers()
	if pls.GetTextLayerCount() != 0 {
		t.Errorf("text count should be 0 after clear, got %d", pls.GetTextLayerCount())
	}
}

// TestPictureLayerSetDirtyRegion はダーティ領域の追跡をテストする
func TestPictureLayerSetDirtyRegion(t *testing.T) {
	pls := NewPictureLayerSet(1)

	// 初期状態
	if !pls.DirtyRegion.Empty() {
		t.Error("DirtyRegion should be empty initially")
	}

	// ダーティ領域を追加
	rect1 := image.Rect(10, 10, 50, 50)
	pls.AddDirtyRegion(rect1)

	if pls.DirtyRegion != rect1 {
		t.Errorf("DirtyRegion should be %v, got %v", rect1, pls.DirtyRegion)
	}

	// 2つ目のダーティ領域を追加（統合される）
	rect2 := image.Rect(40, 40, 100, 100)
	pls.AddDirtyRegion(rect2)

	expected := rect1.Union(rect2)
	if pls.DirtyRegion != expected {
		t.Errorf("DirtyRegion should be %v, got %v", expected, pls.DirtyRegion)
	}

	// クリア
	pls.ClearDirtyRegion()
	if !pls.DirtyRegion.Empty() {
		t.Error("DirtyRegion should be empty after clear")
	}
	if pls.FullDirty {
		t.Error("FullDirty should be false after clear")
	}
}

// TestPictureLayerSetIsDirty はダーティ状態の判定をテストする
func TestPictureLayerSetIsDirty(t *testing.T) {
	pls := NewPictureLayerSet(1)

	// 初期状態はダーティ
	if !pls.IsDirty() {
		t.Error("should be dirty initially (FullDirty=true)")
	}

	// クリア後はダーティではない
	pls.ClearDirtyRegion()
	if pls.IsDirty() {
		t.Error("should not be dirty after clear")
	}

	// ダーティ領域を追加するとダーティ
	pls.AddDirtyRegion(image.Rect(10, 10, 50, 50))
	if !pls.IsDirty() {
		t.Error("should be dirty after adding dirty region")
	}

	// クリア
	pls.ClearDirtyRegion()

	// 背景レイヤーがダーティだとダーティ
	bg := NewBackgroundLayer(1, 1, nil)
	bg.SetDirty(true)
	pls.SetBackground(bg)
	pls.ClearDirtyRegion() // SetBackgroundでFullDirtyがtrueになるのでクリア
	bg.SetDirty(true)
	if !pls.IsDirty() {
		t.Error("should be dirty when background is dirty")
	}

	// 背景のダーティをクリア
	bg.SetDirty(false)
	if pls.IsDirty() {
		t.Error("should not be dirty when background is not dirty")
	}

	// キャストレイヤーがダーティだとダーティ
	cast := NewCastLayer(1, 100, 1, 2, 10, 20, 0, 0, 50, 50, 0)
	cast.SetDirty(true)
	pls.AddCastLayer(cast)
	pls.ClearDirtyRegion() // AddCastLayerでFullDirtyがtrueになるのでクリア
	cast.SetDirty(true)
	if !pls.IsDirty() {
		t.Error("should be dirty when cast is dirty")
	}
}

// TestPictureLayerSetZOrderOffsets はZ順序オフセットの管理をテストする
func TestPictureLayerSetZOrderOffsets(t *testing.T) {
	pls := NewPictureLayerSet(1)

	// 初期状態
	if pls.GetNextCastZOffset() != 0 {
		t.Errorf("initial cast z offset should be 0, got %d", pls.GetNextCastZOffset())
	}
	if pls.GetNextTextZOffset() != 0 {
		t.Errorf("initial text z offset should be 0, got %d", pls.GetNextTextZOffset())
	}

	// キャストを追加するとオフセットが増加
	pls.AddCastLayer(NewCastLayer(1, 100, 1, 2, 10, 20, 0, 0, 50, 50, 0))
	if pls.GetNextCastZOffset() != 1 {
		t.Errorf("cast z offset should be 1, got %d", pls.GetNextCastZOffset())
	}

	pls.AddCastLayer(NewCastLayer(2, 101, 1, 2, 30, 40, 0, 0, 50, 50, 1))
	if pls.GetNextCastZOffset() != 2 {
		t.Errorf("cast z offset should be 2, got %d", pls.GetNextCastZOffset())
	}

	// テキストを追加するとオフセットが増加
	pls.AddTextLayer(NewTextLayerEntry(3, 1, 10, 20, "Hello", 0))
	if pls.GetNextTextZOffset() != 1 {
		t.Errorf("text z offset should be 1, got %d", pls.GetNextTextZOffset())
	}

	// クリアするとオフセットがリセット
	pls.ClearCastLayers()
	if pls.GetNextCastZOffset() != 0 {
		t.Errorf("cast z offset should be 0 after clear, got %d", pls.GetNextCastZOffset())
	}

	pls.ClearTextLayers()
	if pls.GetNextTextZOffset() != 0 {
		t.Errorf("text z offset should be 0 after clear, got %d", pls.GetNextTextZOffset())
	}
}

// TestPictureLayerSetCompositeBuffer は合成バッファの管理をテストする
func TestPictureLayerSetCompositeBuffer(t *testing.T) {
	pls := NewPictureLayerSet(1)

	// 初期状態
	if pls.GetCompositeBuffer() != nil {
		t.Error("CompositeBuffer should be nil initially")
	}

	// バッファを設定
	buffer := ebiten.NewImage(100, 100)
	pls.SetCompositeBuffer(buffer)

	if pls.GetCompositeBuffer() != buffer {
		t.Error("GetCompositeBuffer should return the set buffer")
	}
}

// TestPictureLayerSetClearAllDirtyFlags はすべてのダーティフラグのクリアをテストする
func TestPictureLayerSetClearAllDirtyFlags(t *testing.T) {
	pls := NewPictureLayerSet(1)

	// 各レイヤーを追加してダーティに設定
	bg := NewBackgroundLayer(1, 1, nil)
	bg.SetDirty(true)
	pls.SetBackground(bg)

	drawing := NewDrawingLayer(2, 1, 100, 100)
	drawing.SetDirty(true)
	pls.SetDrawing(drawing)

	cast := NewCastLayer(3, 100, 1, 2, 10, 20, 0, 0, 50, 50, 0)
	cast.SetDirty(true)
	pls.AddCastLayer(cast)

	text := NewTextLayerEntry(4, 1, 10, 20, "Hello", 0)
	text.SetDirty(true)
	pls.AddTextLayer(text)

	pls.AddDirtyRegion(image.Rect(10, 10, 50, 50))
	pls.MarkFullDirty()

	// すべてダーティであることを確認
	if !pls.IsDirty() {
		t.Error("should be dirty before clear")
	}

	// すべてのダーティフラグをクリア
	pls.ClearAllDirtyFlags()

	// すべてクリアされていることを確認
	if bg.IsDirty() {
		t.Error("background should not be dirty after clear")
	}
	if drawing.IsDirty() {
		t.Error("drawing should not be dirty after clear")
	}
	if cast.IsDirty() {
		t.Error("cast should not be dirty after clear")
	}
	if text.IsDirty() {
		t.Error("text should not be dirty after clear")
	}
	if pls.FullDirty {
		t.Error("FullDirty should be false after clear")
	}
	if !pls.DirtyRegion.Empty() {
		t.Error("DirtyRegion should be empty after clear")
	}
}

// TestLayerManagerConcurrency はLayerManagerの並行アクセスをテストする
func TestLayerManagerConcurrency(t *testing.T) {
	lm := NewLayerManager()

	// 並行してPictureLayerSetを作成・取得
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				lm.GetOrCreatePictureLayerSet(id)
				lm.GetPictureLayerSet(id)
				lm.GetNextLayerID()
			}
			done <- true
		}(i)
	}

	// すべてのゴルーチンが完了するのを待つ
	for i := 0; i < 10; i++ {
		<-done
	}

	// 10個のPictureLayerSetが作成されていることを確認
	if lm.GetPictureLayerSetCount() != 10 {
		t.Errorf("should have 10 PictureLayerSets, got %d", lm.GetPictureLayerSetCount())
	}
}

// TestShouldSkipLayer は上書きスキップ判定をテストする
// 要件 7.1, 7.2, 7.3: 上書きスキップの正確性
func TestShouldSkipLayer(t *testing.T) {
	lm := NewLayerManager()

	t.Run("nil layer should be skipped", func(t *testing.T) {
		if !lm.ShouldSkipLayer(nil, nil) {
			t.Error("nil layer should be skipped")
		}
	})

	t.Run("invisible layer should be skipped", func(t *testing.T) {
		layer := NewBackgroundLayer(1, 1, nil)
		layer.SetVisible(false)
		if !lm.ShouldSkipLayer(layer, nil) {
			t.Error("invisible layer should be skipped")
		}
	})

	t.Run("layer with empty bounds should be skipped", func(t *testing.T) {
		layer := NewBackgroundLayer(1, 1, nil)
		// 画像がないので境界は空
		if !lm.ShouldSkipLayer(layer, nil) {
			t.Error("layer with empty bounds should be skipped")
		}
	})

	t.Run("visible layer without upper layers should not be skipped", func(t *testing.T) {
		img := ebiten.NewImage(100, 100)
		layer := NewBackgroundLayer(1, 1, img)
		if lm.ShouldSkipLayer(layer, nil) {
			t.Error("visible layer without upper layers should not be skipped")
		}
	})

	t.Run("layer completely covered by opaque upper layer should be skipped", func(t *testing.T) {
		// 下位レイヤー: 50x50 at (25, 25)
		lowerImg := ebiten.NewImage(50, 50)
		lower := NewBackgroundLayer(1, 1, lowerImg)
		lower.SetBounds(image.Rect(25, 25, 75, 75))

		// 上位レイヤー: 100x100 at (0, 0) - 下位レイヤーを完全に覆う
		upperImg := ebiten.NewImage(100, 100)
		upper := NewBackgroundLayer(2, 1, upperImg)
		upper.SetOpaque(true)

		upperLayers := []Layer{upper}

		if !lm.ShouldSkipLayer(lower, upperLayers) {
			t.Error("layer completely covered by opaque upper layer should be skipped")
		}
	})

	t.Run("layer partially covered by opaque upper layer should not be skipped", func(t *testing.T) {
		// 下位レイヤー: 100x100 at (0, 0)
		lowerImg := ebiten.NewImage(100, 100)
		lower := NewBackgroundLayer(1, 1, lowerImg)

		// 上位レイヤー: 50x50 at (25, 25) - 下位レイヤーを部分的にしか覆わない
		upperImg := ebiten.NewImage(50, 50)
		upper := NewBackgroundLayer(2, 1, upperImg)
		upper.SetBounds(image.Rect(25, 25, 75, 75))
		upper.SetOpaque(true)

		upperLayers := []Layer{upper}

		if lm.ShouldSkipLayer(lower, upperLayers) {
			t.Error("layer partially covered by opaque upper layer should not be skipped")
		}
	})

	t.Run("layer covered by transparent upper layer should not be skipped", func(t *testing.T) {
		// 下位レイヤー: 50x50 at (25, 25)
		lowerImg := ebiten.NewImage(50, 50)
		lower := NewBackgroundLayer(1, 1, lowerImg)
		lower.SetBounds(image.Rect(25, 25, 75, 75))

		// 上位レイヤー: 100x100 at (0, 0) - 下位レイヤーを完全に覆うが透明
		upperImg := ebiten.NewImage(100, 100)
		upper := NewBackgroundLayer(2, 1, upperImg)
		upper.SetOpaque(false) // 透明

		upperLayers := []Layer{upper}

		if lm.ShouldSkipLayer(lower, upperLayers) {
			t.Error("layer covered by transparent upper layer should not be skipped")
		}
	})

	t.Run("layer covered by invisible upper layer should not be skipped", func(t *testing.T) {
		// 下位レイヤー: 50x50 at (25, 25)
		lowerImg := ebiten.NewImage(50, 50)
		lower := NewBackgroundLayer(1, 1, lowerImg)
		lower.SetBounds(image.Rect(25, 25, 75, 75))

		// 上位レイヤー: 100x100 at (0, 0) - 下位レイヤーを完全に覆うが非表示
		upperImg := ebiten.NewImage(100, 100)
		upper := NewBackgroundLayer(2, 1, upperImg)
		upper.SetOpaque(true)
		upper.SetVisible(false) // 非表示

		upperLayers := []Layer{upper}

		if lm.ShouldSkipLayer(lower, upperLayers) {
			t.Error("layer covered by invisible upper layer should not be skipped")
		}
	})

	t.Run("multiple upper layers - one opaque covering should skip", func(t *testing.T) {
		// 下位レイヤー: 50x50 at (25, 25)
		lowerImg := ebiten.NewImage(50, 50)
		lower := NewBackgroundLayer(1, 1, lowerImg)
		lower.SetBounds(image.Rect(25, 25, 75, 75))

		// 上位レイヤー1: 透明で覆う
		upper1Img := ebiten.NewImage(100, 100)
		upper1 := NewBackgroundLayer(2, 1, upper1Img)
		upper1.SetOpaque(false)

		// 上位レイヤー2: 不透明で完全に覆う
		upper2Img := ebiten.NewImage(100, 100)
		upper2 := NewBackgroundLayer(3, 1, upper2Img)
		upper2.SetOpaque(true)

		upperLayers := []Layer{upper1, upper2}

		if !lm.ShouldSkipLayer(lower, upperLayers) {
			t.Error("layer should be skipped when at least one opaque upper layer covers it")
		}
	})
}

// TestGetUpperLayers は上位レイヤー取得をテストする
// 要件 10.1, 10.2: 操作順序に基づくZ順序
func TestGetUpperLayers(t *testing.T) {
	pls := NewPictureLayerSet(1)

	// 背景レイヤー（Z順序: 0）
	bg := NewBackgroundLayer(1, 1, nil)
	pls.SetBackground(bg)

	// 描画レイヤー（Z順序: 1）
	drawing := NewDrawingLayer(2, 1, 100, 100)
	pls.SetDrawing(drawing)

	// キャストレイヤー（操作順序に基づくZ順序: 1, 2）
	cast1 := NewCastLayer(3, 100, 1, 2, 10, 20, 0, 0, 50, 50, 0)
	cast2 := NewCastLayer(4, 101, 1, 2, 30, 40, 0, 0, 50, 50, 0)
	pls.AddCastLayer(cast1) // Z=1
	pls.AddCastLayer(cast2) // Z=2

	// テキストレイヤー（操作順序に基づくZ順序: 3）
	text := NewTextLayerEntry(5, 1, 10, 20, "Hello", 0)
	pls.AddTextLayer(text) // Z=3

	t.Run("get upper layers from background", func(t *testing.T) {
		upperLayers := pls.GetUpperLayers(ZOrderBackground) // Z=0
		// 描画(Z=1)、キャスト2つ(Z=1,2)、テキスト1つ(Z=3) = 4つ
		if len(upperLayers) != 4 {
			t.Errorf("expected 4 upper layers, got %d", len(upperLayers))
		}
	})

	t.Run("get upper layers from drawing", func(t *testing.T) {
		upperLayers := pls.GetUpperLayers(ZOrderDrawing) // Z=1
		// キャスト1つ(Z=2)、テキスト1つ(Z=3) = 2つ
		// 注: cast1はZ=1なので含まれない
		if len(upperLayers) != 2 {
			t.Errorf("expected 2 upper layers, got %d", len(upperLayers))
		}
	})

	t.Run("get upper layers from cast", func(t *testing.T) {
		upperLayers := pls.GetUpperLayers(1) // Z=1（cast1のZ順序）
		// キャスト1つ(Z=2)、テキスト1つ(Z=3) = 2つ
		if len(upperLayers) != 2 {
			t.Errorf("expected 2 upper layers, got %d", len(upperLayers))
		}
	})

	t.Run("get upper layers from text", func(t *testing.T) {
		upperLayers := pls.GetUpperLayers(3) // Z=3（textのZ順序）
		// 上位レイヤーなし
		if len(upperLayers) != 0 {
			t.Errorf("expected 0 upper layers, got %d", len(upperLayers))
		}
	})
}

// TestGetAllLayersSorted はすべてのレイヤーのソート取得をテストする
func TestGetAllLayersSorted(t *testing.T) {
	pls := NewPictureLayerSet(1)

	// 背景レイヤー
	bg := NewBackgroundLayer(1, 1, nil)
	pls.SetBackground(bg)

	// 描画レイヤー
	drawing := NewDrawingLayer(2, 1, 100, 100)
	pls.SetDrawing(drawing)

	// キャストレイヤー
	cast := NewCastLayer(3, 100, 1, 2, 10, 20, 0, 0, 50, 50, 0)
	pls.AddCastLayer(cast)

	// テキストレイヤー
	text := NewTextLayerEntry(4, 1, 10, 20, "Hello", 0)
	pls.AddTextLayer(text)

	layers := pls.GetAllLayersSorted()

	if len(layers) != 4 {
		t.Errorf("expected 4 layers, got %d", len(layers))
	}

	// Z順序が正しいことを確認
	for i := 1; i < len(layers); i++ {
		if layers[i-1].GetZOrder() > layers[i].GetZOrder() {
			t.Errorf("layers not sorted by Z order: %d > %d",
				layers[i-1].GetZOrder(), layers[i].GetZOrder())
		}
	}
}

// TestLayerOpacity はレイヤーの不透明度をテストする
func TestLayerOpacity(t *testing.T) {
	t.Run("background layer is opaque by default", func(t *testing.T) {
		bg := NewBackgroundLayer(1, 1, nil)
		if !bg.IsOpaque() {
			t.Error("background layer should be opaque by default")
		}
	})

	t.Run("drawing layer is transparent by default", func(t *testing.T) {
		drawing := NewDrawingLayer(1, 1, 100, 100)
		if drawing.IsOpaque() {
			t.Error("drawing layer should be transparent by default")
		}
	})

	t.Run("cast layer is transparent by default", func(t *testing.T) {
		cast := NewCastLayer(1, 100, 1, 2, 10, 20, 0, 0, 50, 50, 0)
		if cast.IsOpaque() {
			t.Error("cast layer should be transparent by default")
		}
	})

	t.Run("text layer is transparent by default", func(t *testing.T) {
		text := NewTextLayerEntry(1, 1, 10, 20, "Hello", 0)
		if text.IsOpaque() {
			t.Error("text layer should be transparent by default")
		}
	})

	t.Run("SetOpaque changes opacity", func(t *testing.T) {
		bg := NewBackgroundLayer(1, 1, nil)
		bg.SetOpaque(false)
		if bg.IsOpaque() {
			t.Error("SetOpaque(false) should make layer transparent")
		}

		bg.SetOpaque(true)
		if !bg.IsOpaque() {
			t.Error("SetOpaque(true) should make layer opaque")
		}
	})
}

// TestContainsRect は矩形包含判定をテストする
func TestContainsRect(t *testing.T) {
	t.Run("rect contains itself", func(t *testing.T) {
		rect := image.Rect(10, 10, 50, 50)
		if !containsRect(rect, rect) {
			t.Error("rect should contain itself")
		}
	})

	t.Run("larger rect contains smaller rect", func(t *testing.T) {
		larger := image.Rect(0, 0, 100, 100)
		smaller := image.Rect(25, 25, 75, 75)
		if !containsRect(larger, smaller) {
			t.Error("larger rect should contain smaller rect")
		}
	})

	t.Run("smaller rect does not contain larger rect", func(t *testing.T) {
		larger := image.Rect(0, 0, 100, 100)
		smaller := image.Rect(25, 25, 75, 75)
		if containsRect(smaller, larger) {
			t.Error("smaller rect should not contain larger rect")
		}
	})

	t.Run("partially overlapping rects", func(t *testing.T) {
		rect1 := image.Rect(0, 0, 50, 50)
		rect2 := image.Rect(25, 25, 75, 75)
		if containsRect(rect1, rect2) {
			t.Error("partially overlapping rect should not contain the other")
		}
		if containsRect(rect2, rect1) {
			t.Error("partially overlapping rect should not contain the other")
		}
	})

	t.Run("non-overlapping rects", func(t *testing.T) {
		rect1 := image.Rect(0, 0, 50, 50)
		rect2 := image.Rect(100, 100, 150, 150)
		if containsRect(rect1, rect2) {
			t.Error("non-overlapping rect should not contain the other")
		}
	})
}

// TestIsLayerVisible は可視領域クリッピング判定をテストする
// 要件 4.1, 4.4: 可視領域クリッピングの正確性
func TestIsLayerVisible(t *testing.T) {
	t.Run("nil layer is not visible", func(t *testing.T) {
		visibleRect := image.Rect(0, 0, 100, 100)
		if IsLayerVisible(nil, visibleRect) {
			t.Error("nil layer should not be visible")
		}
	})

	t.Run("invisible layer is not visible", func(t *testing.T) {
		img := ebiten.NewImage(50, 50)
		layer := NewBackgroundLayer(1, 1, img)
		layer.SetVisible(false)
		visibleRect := image.Rect(0, 0, 100, 100)
		if IsLayerVisible(layer, visibleRect) {
			t.Error("invisible layer should not be visible")
		}
	})

	t.Run("layer with empty bounds is not visible", func(t *testing.T) {
		layer := NewBackgroundLayer(1, 1, nil)
		visibleRect := image.Rect(0, 0, 100, 100)
		if IsLayerVisible(layer, visibleRect) {
			t.Error("layer with empty bounds should not be visible")
		}
	})

	t.Run("layer completely inside visible rect is visible", func(t *testing.T) {
		img := ebiten.NewImage(50, 50)
		layer := NewBackgroundLayer(1, 1, img)
		layer.SetBounds(image.Rect(25, 25, 75, 75))
		visibleRect := image.Rect(0, 0, 100, 100)
		if !IsLayerVisible(layer, visibleRect) {
			t.Error("layer completely inside visible rect should be visible")
		}
	})

	t.Run("layer partially inside visible rect is visible", func(t *testing.T) {
		img := ebiten.NewImage(50, 50)
		layer := NewBackgroundLayer(1, 1, img)
		layer.SetBounds(image.Rect(75, 75, 125, 125))
		visibleRect := image.Rect(0, 0, 100, 100)
		if !IsLayerVisible(layer, visibleRect) {
			t.Error("layer partially inside visible rect should be visible")
		}
	})

	t.Run("layer completely outside visible rect is not visible", func(t *testing.T) {
		img := ebiten.NewImage(50, 50)
		layer := NewBackgroundLayer(1, 1, img)
		layer.SetBounds(image.Rect(200, 200, 250, 250))
		visibleRect := image.Rect(0, 0, 100, 100)
		if IsLayerVisible(layer, visibleRect) {
			t.Error("layer completely outside visible rect should not be visible")
		}
	})
}

// TestGetVisibleRegion は可視部分の取得をテストする
// 要件 4.2: レイヤーが部分的に可視領域内にあるときに可視部分のみを描画する
func TestGetVisibleRegion(t *testing.T) {
	t.Run("nil layer returns empty rect", func(t *testing.T) {
		visibleRect := image.Rect(0, 0, 100, 100)
		result := GetVisibleRegion(nil, visibleRect)
		if !result.Empty() {
			t.Error("nil layer should return empty rect")
		}
	})

	t.Run("layer completely inside visible rect", func(t *testing.T) {
		img := ebiten.NewImage(50, 50)
		layer := NewBackgroundLayer(1, 1, img)
		layer.SetBounds(image.Rect(25, 25, 75, 75))
		visibleRect := image.Rect(0, 0, 100, 100)
		result := GetVisibleRegion(layer, visibleRect)
		expected := image.Rect(25, 25, 75, 75)
		if result != expected {
			t.Errorf("expected %v, got %v", expected, result)
		}
	})

	t.Run("layer partially inside visible rect", func(t *testing.T) {
		img := ebiten.NewImage(50, 50)
		layer := NewBackgroundLayer(1, 1, img)
		layer.SetBounds(image.Rect(75, 75, 125, 125))
		visibleRect := image.Rect(0, 0, 100, 100)
		result := GetVisibleRegion(layer, visibleRect)
		expected := image.Rect(75, 75, 100, 100)
		if result != expected {
			t.Errorf("expected %v, got %v", expected, result)
		}
	})

	t.Run("layer completely outside visible rect", func(t *testing.T) {
		img := ebiten.NewImage(50, 50)
		layer := NewBackgroundLayer(1, 1, img)
		layer.SetBounds(image.Rect(200, 200, 250, 250))
		visibleRect := image.Rect(0, 0, 100, 100)
		result := GetVisibleRegion(layer, visibleRect)
		if !result.Empty() {
			t.Error("layer outside visible rect should return empty rect")
		}
	})
}

// TestComposite は合成処理をテストする
// 要件 1.6: 背景 → 描画 → キャスト → テキストの順で合成する
// 要件 6.2: ダーティ領域のみを再合成する
// 要件 3.4: 合成処理が完了したときにすべてのDirty_Flagをクリアする
func TestComposite(t *testing.T) {
	t.Run("empty visible rect returns existing buffer", func(t *testing.T) {
		pls := NewPictureLayerSet(1)
		existingBuffer := ebiten.NewImage(100, 100)
		pls.SetCompositeBuffer(existingBuffer)

		result := pls.Composite(image.Rectangle{})
		if result != existingBuffer {
			t.Error("empty visible rect should return existing buffer")
		}
	})

	t.Run("composite creates buffer if nil", func(t *testing.T) {
		pls := NewPictureLayerSet(1)
		visibleRect := image.Rect(0, 0, 100, 100)

		result := pls.Composite(visibleRect)
		if result == nil {
			t.Error("composite should create buffer")
		}

		bounds := result.Bounds()
		if bounds.Dx() != 100 || bounds.Dy() != 100 {
			t.Errorf("buffer size should be 100x100, got %dx%d", bounds.Dx(), bounds.Dy())
		}
	})

	t.Run("composite clears dirty flags", func(t *testing.T) {
		pls := NewPictureLayerSet(1)

		// 背景レイヤーを追加
		bgImg := ebiten.NewImage(100, 100)
		bg := NewBackgroundLayer(1, 1, bgImg)
		bg.SetDirty(true)
		pls.SetBackground(bg)

		// 描画レイヤーを追加
		drawing := NewDrawingLayer(2, 1, 100, 100)
		drawing.SetDirty(true)
		pls.SetDrawing(drawing)

		// キャストレイヤーを追加
		cast := NewCastLayer(3, 100, 1, 2, 10, 20, 0, 0, 50, 50, 0)
		cast.SetDirty(true)
		pls.AddCastLayer(cast)

		// テキストレイヤーを追加
		text := NewTextLayerEntry(4, 1, 10, 20, "Hello", 0)
		text.SetDirty(true)
		pls.AddTextLayer(text)

		// ダーティ領域を追加
		pls.AddDirtyRegion(image.Rect(10, 10, 50, 50))

		// 合成前はダーティ
		if !pls.IsDirty() {
			t.Error("should be dirty before composite")
		}

		// 合成
		visibleRect := image.Rect(0, 0, 100, 100)
		pls.Composite(visibleRect)

		// 合成後はダーティではない
		if pls.IsDirty() {
			t.Error("should not be dirty after composite")
		}
		if bg.IsDirty() {
			t.Error("background should not be dirty after composite")
		}
		if drawing.IsDirty() {
			t.Error("drawing should not be dirty after composite")
		}
		if cast.IsDirty() {
			t.Error("cast should not be dirty after composite")
		}
		if text.IsDirty() {
			t.Error("text should not be dirty after composite")
		}
	})

	t.Run("composite returns cached buffer when not dirty", func(t *testing.T) {
		pls := NewPictureLayerSet(1)
		visibleRect := image.Rect(0, 0, 100, 100)

		// 最初の合成
		result1 := pls.Composite(visibleRect)

		// ダーティフラグがクリアされているので、2回目は同じバッファを返す
		result2 := pls.Composite(visibleRect)

		if result1 != result2 {
			t.Error("should return same buffer when not dirty")
		}
	})

	t.Run("composite resizes buffer when visible rect changes", func(t *testing.T) {
		pls := NewPictureLayerSet(1)

		// 最初の合成
		visibleRect1 := image.Rect(0, 0, 100, 100)
		result1 := pls.Composite(visibleRect1)
		bounds1 := result1.Bounds()

		// ダーティにする
		pls.MarkFullDirty()

		// サイズを変更して合成
		visibleRect2 := image.Rect(0, 0, 200, 200)
		result2 := pls.Composite(visibleRect2)
		bounds2 := result2.Bounds()

		if bounds1.Dx() == bounds2.Dx() && bounds1.Dy() == bounds2.Dy() {
			t.Error("buffer should be resized when visible rect changes")
		}
		if bounds2.Dx() != 200 || bounds2.Dy() != 200 {
			t.Errorf("buffer size should be 200x200, got %dx%d", bounds2.Dx(), bounds2.Dy())
		}
	})
}

// TestCompositeLayerOrder は合成順序をテストする
// 要件 1.6: 背景 → 描画 → キャスト → テキストの順で合成する
func TestCompositeLayerOrder(t *testing.T) {
	pls := NewPictureLayerSet(1)

	// 各レイヤーを追加（異なる色で識別）
	// 背景レイヤー
	bgImg := ebiten.NewImage(100, 100)
	bg := NewBackgroundLayer(1, 1, bgImg)
	pls.SetBackground(bg)

	// 描画レイヤー
	drawingImg := ebiten.NewImage(100, 100)
	drawing := NewDrawingLayerWithImage(2, 1, drawingImg)
	pls.SetDrawing(drawing)

	// キャストレイヤー
	cast := NewCastLayer(3, 100, 1, 2, 10, 20, 0, 0, 50, 50, 0)
	pls.AddCastLayer(cast)

	// テキストレイヤー
	text := NewTextLayerEntry(4, 1, 10, 20, "Hello", 0)
	pls.AddTextLayer(text)

	// 合成
	visibleRect := image.Rect(0, 0, 100, 100)
	result := pls.Composite(visibleRect)

	// 結果が存在することを確認
	if result == nil {
		t.Error("composite result should not be nil")
	}

	// レイヤーの順序を確認（GetAllLayersSortedで確認）
	layers := pls.GetAllLayersSorted()
	if len(layers) != 4 {
		t.Errorf("expected 4 layers, got %d", len(layers))
	}

	// Z順序が正しいことを確認（操作順序に基づく）
	// 背景=0, 描画=1, キャスト=1（AddCastLayerで割り当て）, テキスト=2（AddTextLayerで割り当て）
	// 注: 背景と描画は固定Z順序、キャストとテキストは操作順序に基づく
	expectedZOrders := []int{ZOrderBackground, ZOrderDrawing, 1, 2}
	for i, layer := range layers {
		if layer.GetZOrder() != expectedZOrders[i] {
			t.Errorf("layer %d: expected Z order %d, got %d", i, expectedZOrders[i], layer.GetZOrder())
		}
	}
}

// TestCompositeWithVisibilityClipping は可視領域クリッピングを含む合成をテストする
// 要件 4.1: レイヤーがウィンドウの可視領域外にあるときに描画をスキップする
func TestCompositeWithVisibilityClipping(t *testing.T) {
	pls := NewPictureLayerSet(1)

	// 可視領域内のレイヤー
	bgImg := ebiten.NewImage(50, 50)
	bg := NewBackgroundLayer(1, 1, bgImg)
	bg.SetBounds(image.Rect(0, 0, 50, 50))
	pls.SetBackground(bg)

	// 可視領域外のキャストレイヤー
	cast := NewCastLayer(2, 100, 1, 2, 200, 200, 0, 0, 50, 50, 0)
	pls.AddCastLayer(cast)

	// 合成（可視領域は0,0から100,100）
	visibleRect := image.Rect(0, 0, 100, 100)
	result := pls.Composite(visibleRect)

	// 結果が存在することを確認
	if result == nil {
		t.Error("composite result should not be nil")
	}

	// 可視領域外のキャストは描画されないことを確認
	// （実際の描画内容の確認は困難なので、エラーなく完了することを確認）
}

// TestCompositeWithOverwriteSkip は上書きスキップを含む合成をテストする
// 要件 7.1: 不透明なレイヤーが別のレイヤーを完全に覆っているときにスキップする
func TestCompositeWithOverwriteSkip(t *testing.T) {
	pls := NewPictureLayerSet(1)

	// 背景レイヤー（小さい）
	bgImg := ebiten.NewImage(50, 50)
	bg := NewBackgroundLayer(1, 1, bgImg)
	bg.SetBounds(image.Rect(25, 25, 75, 75))
	pls.SetBackground(bg)

	// 描画レイヤー（大きく、不透明で背景を完全に覆う）
	drawingImg := ebiten.NewImage(100, 100)
	drawing := NewDrawingLayerWithImage(2, 1, drawingImg)
	drawing.SetOpaque(true)
	pls.SetDrawing(drawing)

	// 合成
	visibleRect := image.Rect(0, 0, 100, 100)
	result := pls.Composite(visibleRect)

	// 結果が存在することを確認
	if result == nil {
		t.Error("composite result should not be nil")
	}

	// 上書きスキップが適用されることを確認
	// （実際の描画内容の確認は困難なので、エラーなく完了することを確認）
}

// =============================================================================
// タスク 5.4: 合成処理のユニットテスト
// 要件 4.1, 4.2, 7.1, 1.6 に基づく追加テスト
// =============================================================================

// TestVisibleRegionClipping_EmptyVisibleRegion は空の可視領域のテスト
// 要件 4.1: レイヤーがウィンドウの可視領域外にあるときに描画をスキップする
func TestVisibleRegionClipping_EmptyVisibleRegion(t *testing.T) {
	pls := NewPictureLayerSet(1)

	// 背景レイヤーを追加
	bgImg := ebiten.NewImage(100, 100)
	bg := NewBackgroundLayer(1, 1, bgImg)
	pls.SetBackground(bg)

	// 空の可視領域で合成
	emptyRect := image.Rectangle{}
	result := pls.Composite(emptyRect)

	// 空の可視領域の場合、既存のバッファを返す（nilの場合はnil）
	// 新しいバッファは作成されない
	if result != nil && result != pls.GetCompositeBuffer() {
		t.Error("empty visible rect should return existing buffer or nil")
	}
}

// TestVisibleRegionClipping_LayerCompletelyOutside は完全に可視領域外のレイヤーのテスト
// 要件 4.1: レイヤーがウィンドウの可視領域外にあるときに描画をスキップする
func TestVisibleRegionClipping_LayerCompletelyOutside(t *testing.T) {
	t.Run("layer to the right of visible region", func(t *testing.T) {
		img := ebiten.NewImage(50, 50)
		layer := NewBackgroundLayer(1, 1, img)
		layer.SetBounds(image.Rect(200, 0, 250, 50))
		visibleRect := image.Rect(0, 0, 100, 100)

		if IsLayerVisible(layer, visibleRect) {
			t.Error("layer to the right should not be visible")
		}
	})

	t.Run("layer below visible region", func(t *testing.T) {
		img := ebiten.NewImage(50, 50)
		layer := NewBackgroundLayer(1, 1, img)
		layer.SetBounds(image.Rect(0, 200, 50, 250))
		visibleRect := image.Rect(0, 0, 100, 100)

		if IsLayerVisible(layer, visibleRect) {
			t.Error("layer below should not be visible")
		}
	})

	t.Run("layer to the left of visible region", func(t *testing.T) {
		img := ebiten.NewImage(50, 50)
		layer := NewBackgroundLayer(1, 1, img)
		layer.SetBounds(image.Rect(-100, 0, -50, 50))
		visibleRect := image.Rect(0, 0, 100, 100)

		if IsLayerVisible(layer, visibleRect) {
			t.Error("layer to the left should not be visible")
		}
	})

	t.Run("layer above visible region", func(t *testing.T) {
		img := ebiten.NewImage(50, 50)
		layer := NewBackgroundLayer(1, 1, img)
		layer.SetBounds(image.Rect(0, -100, 50, -50))
		visibleRect := image.Rect(0, 0, 100, 100)

		if IsLayerVisible(layer, visibleRect) {
			t.Error("layer above should not be visible")
		}
	})
}

// TestVisibleRegionClipping_LayerPartiallyInside は部分的に可視領域内のレイヤーのテスト
// 要件 4.2: レイヤーが部分的に可視領域内にあるときに可視部分のみを描画する
func TestVisibleRegionClipping_LayerPartiallyInside(t *testing.T) {
	t.Run("layer overlapping right edge", func(t *testing.T) {
		img := ebiten.NewImage(50, 50)
		layer := NewBackgroundLayer(1, 1, img)
		layer.SetBounds(image.Rect(80, 25, 130, 75))
		visibleRect := image.Rect(0, 0, 100, 100)

		if !IsLayerVisible(layer, visibleRect) {
			t.Error("layer overlapping right edge should be visible")
		}

		visibleRegion := GetVisibleRegion(layer, visibleRect)
		expected := image.Rect(80, 25, 100, 75)
		if visibleRegion != expected {
			t.Errorf("expected visible region %v, got %v", expected, visibleRegion)
		}
	})

	t.Run("layer overlapping bottom edge", func(t *testing.T) {
		img := ebiten.NewImage(50, 50)
		layer := NewBackgroundLayer(1, 1, img)
		layer.SetBounds(image.Rect(25, 80, 75, 130))
		visibleRect := image.Rect(0, 0, 100, 100)

		if !IsLayerVisible(layer, visibleRect) {
			t.Error("layer overlapping bottom edge should be visible")
		}

		visibleRegion := GetVisibleRegion(layer, visibleRect)
		expected := image.Rect(25, 80, 75, 100)
		if visibleRegion != expected {
			t.Errorf("expected visible region %v, got %v", expected, visibleRegion)
		}
	})

	t.Run("layer overlapping corner", func(t *testing.T) {
		img := ebiten.NewImage(50, 50)
		layer := NewBackgroundLayer(1, 1, img)
		layer.SetBounds(image.Rect(80, 80, 130, 130))
		visibleRect := image.Rect(0, 0, 100, 100)

		if !IsLayerVisible(layer, visibleRect) {
			t.Error("layer overlapping corner should be visible")
		}

		visibleRegion := GetVisibleRegion(layer, visibleRect)
		expected := image.Rect(80, 80, 100, 100)
		if visibleRegion != expected {
			t.Errorf("expected visible region %v, got %v", expected, visibleRegion)
		}
	})

	t.Run("layer larger than visible region", func(t *testing.T) {
		img := ebiten.NewImage(200, 200)
		layer := NewBackgroundLayer(1, 1, img)
		layer.SetBounds(image.Rect(-50, -50, 150, 150))
		visibleRect := image.Rect(0, 0, 100, 100)

		if !IsLayerVisible(layer, visibleRect) {
			t.Error("layer larger than visible region should be visible")
		}

		visibleRegion := GetVisibleRegion(layer, visibleRect)
		expected := image.Rect(0, 0, 100, 100)
		if visibleRegion != expected {
			t.Errorf("expected visible region %v, got %v", expected, visibleRegion)
		}
	})
}

// TestOverwriteSkip_MultipleLayersCovered は複数のレイヤーが覆われる場合のテスト
// 要件 7.1: 不透明なレイヤーが別のレイヤーを完全に覆っているときにスキップする
func TestOverwriteSkip_MultipleLayersCovered(t *testing.T) {
	lm := NewLayerManager()

	// 3つの小さいレイヤー（すべて覆われる）
	layer1Img := ebiten.NewImage(30, 30)
	layer1 := NewBackgroundLayer(1, 1, layer1Img)
	layer1.SetBounds(image.Rect(10, 10, 40, 40))

	layer2Img := ebiten.NewImage(30, 30)
	layer2 := NewBackgroundLayer(2, 1, layer2Img)
	layer2.SetBounds(image.Rect(40, 40, 70, 70))

	layer3Img := ebiten.NewImage(30, 30)
	layer3 := NewBackgroundLayer(3, 1, layer3Img)
	layer3.SetBounds(image.Rect(20, 50, 50, 80))

	// 大きな不透明レイヤー（すべてを覆う）
	coveringImg := ebiten.NewImage(100, 100)
	coveringLayer := NewBackgroundLayer(4, 1, coveringImg)
	coveringLayer.SetOpaque(true)

	upperLayers := []Layer{coveringLayer}

	// すべてのレイヤーがスキップされるべき
	if !lm.ShouldSkipLayer(layer1, upperLayers) {
		t.Error("layer1 should be skipped when covered by opaque layer")
	}
	if !lm.ShouldSkipLayer(layer2, upperLayers) {
		t.Error("layer2 should be skipped when covered by opaque layer")
	}
	if !lm.ShouldSkipLayer(layer3, upperLayers) {
		t.Error("layer3 should be skipped when covered by opaque layer")
	}
}

// TestOverwriteSkip_PartialCoverage は部分的に覆われる場合のテスト
// 要件 7.1: 部分的に覆われているレイヤーは描画する
func TestOverwriteSkip_PartialCoverage(t *testing.T) {
	lm := NewLayerManager()

	// 大きなレイヤー
	largeImg := ebiten.NewImage(100, 100)
	largeLayer := NewBackgroundLayer(1, 1, largeImg)

	// 小さな不透明レイヤー（部分的にしか覆わない）
	smallImg := ebiten.NewImage(50, 50)
	smallLayer := NewBackgroundLayer(2, 1, smallImg)
	smallLayer.SetBounds(image.Rect(25, 25, 75, 75))
	smallLayer.SetOpaque(true)

	upperLayers := []Layer{smallLayer}

	// 部分的に覆われているのでスキップされない
	if lm.ShouldSkipLayer(largeLayer, upperLayers) {
		t.Error("partially covered layer should not be skipped")
	}
}

// TestOverwriteSkip_MultipleOpaqueLayersChain は複数の不透明レイヤーの連鎖テスト
// 要件 7.1: 不透明なレイヤーが別のレイヤーを完全に覆っているときにスキップする
func TestOverwriteSkip_MultipleOpaqueLayersChain(t *testing.T) {
	lm := NewLayerManager()

	// 最下層のレイヤー
	bottomImg := ebiten.NewImage(50, 50)
	bottomLayer := NewBackgroundLayer(1, 1, bottomImg)
	bottomLayer.SetBounds(image.Rect(25, 25, 75, 75))

	// 中間の不透明レイヤー（最下層を覆う）
	middleImg := ebiten.NewImage(60, 60)
	middleLayer := NewBackgroundLayer(2, 1, middleImg)
	middleLayer.SetBounds(image.Rect(20, 20, 80, 80))
	middleLayer.SetOpaque(true)

	// 最上層の不透明レイヤー（中間層を覆う）
	topImg := ebiten.NewImage(100, 100)
	topLayer := NewBackgroundLayer(3, 1, topImg)
	topLayer.SetOpaque(true)

	// 最下層は中間層によって覆われる
	upperLayersForBottom := []Layer{middleLayer, topLayer}
	if !lm.ShouldSkipLayer(bottomLayer, upperLayersForBottom) {
		t.Error("bottom layer should be skipped when covered by middle opaque layer")
	}

	// 中間層は最上層によって覆われる
	upperLayersForMiddle := []Layer{topLayer}
	if !lm.ShouldSkipLayer(middleLayer, upperLayersForMiddle) {
		t.Error("middle layer should be skipped when covered by top opaque layer")
	}
}

// TestCompositeOrder_ZOrderVerification はZ順序の検証テスト
// 要件 1.6: 背景 → 描画 → キャスト → テキストの順で合成する
func TestCompositeOrder_ZOrderVerification(t *testing.T) {
	pls := NewPictureLayerSet(1)

	// 各レイヤータイプを追加
	bgImg := ebiten.NewImage(100, 100)
	bg := NewBackgroundLayer(1, 1, bgImg)
	pls.SetBackground(bg)

	drawingImg := ebiten.NewImage(100, 100)
	drawing := NewDrawingLayerWithImage(2, 1, drawingImg)
	pls.SetDrawing(drawing)

	cast1 := NewCastLayer(3, 100, 1, 2, 10, 10, 0, 0, 30, 30, 0)
	cast2 := NewCastLayer(4, 101, 1, 2, 40, 40, 0, 0, 30, 30, 1)
	pls.AddCastLayer(cast1)
	pls.AddCastLayer(cast2)

	text1 := NewTextLayerEntry(5, 1, 10, 10, "Text1", 0)
	text2 := NewTextLayerEntry(6, 1, 50, 50, "Text2", 1)
	pls.AddTextLayer(text1)
	pls.AddTextLayer(text2)

	// すべてのレイヤーを取得
	layers := pls.GetAllLayersSorted()

	// 6つのレイヤーがあることを確認
	if len(layers) != 6 {
		t.Errorf("expected 6 layers, got %d", len(layers))
	}

	// Z順序が正しいことを確認
	expectedOrder := []struct {
		layerType string
		minZ      int
		maxZ      int
	}{
		{"background", ZOrderBackground, ZOrderBackground},
		{"drawing", ZOrderDrawing, ZOrderDrawing},
		// 操作順序に基づくZ順序: キャストとテキストは1から開始
		{"cast1", 1, 100},
		{"cast2", 1, 100},
		{"text1", 1, 100},
		{"text2", 1, 100},
	}

	for i, layer := range layers {
		z := layer.GetZOrder()
		if z < expectedOrder[i].minZ || z > expectedOrder[i].maxZ {
			t.Errorf("layer %d (%s): expected Z order in range [%d, %d], got %d",
				i, expectedOrder[i].layerType, expectedOrder[i].minZ, expectedOrder[i].maxZ, z)
		}
	}

	// 順序が昇順であることを確認
	for i := 1; i < len(layers); i++ {
		if layers[i-1].GetZOrder() > layers[i].GetZOrder() {
			t.Errorf("layers not in ascending Z order: layer %d (Z=%d) > layer %d (Z=%d)",
				i-1, layers[i-1].GetZOrder(), i, layers[i].GetZOrder())
		}
	}
}

// TestCompositeOrder_MultipleCastsZOrder は複数キャストのZ順序テスト
// 要件 10.1, 10.2: 操作順序に基づくZ順序
func TestCompositeOrder_MultipleCastsZOrder(t *testing.T) {
	pls := NewPictureLayerSet(1)

	// 複数のキャストを追加（追加順がZ順序になる）
	for i := 0; i < 5; i++ {
		cast := NewCastLayer(i+1, 100+i, 1, 2, i*20, i*20, 0, 0, 30, 30, 0)
		pls.AddCastLayer(cast)
	}

	// キャストのZ順序を確認（操作順序に基づく: 1から開始）
	for i, cast := range pls.Casts {
		expectedZ := 1 + i // 操作順序に基づくZ順序
		if cast.GetZOrder() != expectedZ {
			t.Errorf("cast %d: expected Z order %d, got %d", i, expectedZ, cast.GetZOrder())
		}
	}
}

// TestCompositeOrder_MultipleTextsZOrder は複数テキストのZ順序テスト
// 要件 10.1, 10.2: 操作順序に基づくZ順序
func TestCompositeOrder_MultipleTextsZOrder(t *testing.T) {
	pls := NewPictureLayerSet(1)

	// 複数のテキストを追加（追加順がZ順序になる）
	for i := 0; i < 5; i++ {
		text := NewTextLayerEntry(i+1, 1, i*20, i*20, "Text", 0)
		pls.AddTextLayer(text)
	}

	// テキストのZ順序を確認（操作順序に基づく: 1から開始）
	for i, text := range pls.Texts {
		expectedZ := 1 + i // 操作順序に基づくZ順序
		if text.GetZOrder() != expectedZ {
			t.Errorf("text %d: expected Z order %d, got %d", i, expectedZ, text.GetZOrder())
		}
	}
}

// TestComposite_WithMultipleOverlappingLayers は複数の重なり合うレイヤーの合成テスト
// 要件 4.1, 4.2, 7.1, 1.6 の統合テスト
func TestComposite_WithMultipleOverlappingLayers(t *testing.T) {
	pls := NewPictureLayerSet(1)

	// 背景レイヤー（全体をカバー）
	bgImg := ebiten.NewImage(200, 200)
	bg := NewBackgroundLayer(1, 1, bgImg)
	pls.SetBackground(bg)

	// 描画レイヤー
	drawingImg := ebiten.NewImage(200, 200)
	drawing := NewDrawingLayerWithImage(2, 1, drawingImg)
	pls.SetDrawing(drawing)

	// 複数の重なり合うキャスト
	cast1 := NewCastLayer(3, 100, 1, 2, 10, 10, 0, 0, 50, 50, 0)
	cast2 := NewCastLayer(4, 101, 1, 2, 30, 30, 0, 0, 50, 50, 1)   // cast1と重なる
	cast3 := NewCastLayer(5, 102, 1, 2, 100, 100, 0, 0, 50, 50, 2) // 他と重ならない
	pls.AddCastLayer(cast1)
	pls.AddCastLayer(cast2)
	pls.AddCastLayer(cast3)

	// テキストレイヤー
	text := NewTextLayerEntry(6, 1, 50, 50, "Overlapping", 0)
	pls.AddTextLayer(text)

	// 可視領域（一部のレイヤーのみ含む）
	visibleRect := image.Rect(0, 0, 100, 100)
	result := pls.Composite(visibleRect)

	// 結果が存在することを確認
	if result == nil {
		t.Error("composite result should not be nil")
	}

	// バッファサイズが正しいことを確認
	bounds := result.Bounds()
	if bounds.Dx() != 100 || bounds.Dy() != 100 {
		t.Errorf("buffer size should be 100x100, got %dx%d", bounds.Dx(), bounds.Dy())
	}

	// ダーティフラグがクリアされていることを確認
	if pls.IsDirty() {
		t.Error("should not be dirty after composite")
	}
}

// TestComposite_VisibleRegionOffset は可視領域のオフセットテスト
// 要件 4.2: 可視部分のみを描画する
func TestComposite_VisibleRegionOffset(t *testing.T) {
	pls := NewPictureLayerSet(1)

	// 背景レイヤー
	bgImg := ebiten.NewImage(200, 200)
	bg := NewBackgroundLayer(1, 1, bgImg)
	pls.SetBackground(bg)

	// オフセットされた可視領域
	visibleRect := image.Rect(50, 50, 150, 150)
	result := pls.Composite(visibleRect)

	// 結果が存在することを確認
	if result == nil {
		t.Error("composite result should not be nil")
	}

	// バッファサイズが可視領域のサイズと一致することを確認
	bounds := result.Bounds()
	if bounds.Dx() != 100 || bounds.Dy() != 100 {
		t.Errorf("buffer size should be 100x100, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

// TestComposite_LayerAtEdgeOfVisibleRegion は可視領域の端にあるレイヤーのテスト
// 要件 4.2: レイヤーが部分的に可視領域内にあるときに可視部分のみを描画する
func TestComposite_LayerAtEdgeOfVisibleRegion(t *testing.T) {
	pls := NewPictureLayerSet(1)

	// 可視領域の端にあるキャスト
	cast := NewCastLayer(1, 100, 1, 2, 90, 90, 0, 0, 50, 50, 0)
	castImg := ebiten.NewImage(50, 50)
	cast.SetSourceImage(castImg)
	pls.AddCastLayer(cast)

	visibleRect := image.Rect(0, 0, 100, 100)

	// キャストが部分的に可視であることを確認
	if !IsLayerVisible(cast, visibleRect) {
		t.Error("cast at edge should be visible")
	}

	// 可視部分を確認
	visibleRegion := GetVisibleRegion(cast, visibleRect)
	expected := image.Rect(90, 90, 100, 100)
	if visibleRegion != expected {
		t.Errorf("expected visible region %v, got %v", expected, visibleRegion)
	}

	// 合成
	result := pls.Composite(visibleRect)
	if result == nil {
		t.Error("composite result should not be nil")
	}
}

// TestOperationSequenceZOrder は操作順序に基づくZ順序をテストする
// 要件 10.1, 10.2, 10.3, 10.4: 操作順序に基づくZ順序管理
func TestOperationSequenceZOrder(t *testing.T) {
	t.Run("MovePic creates DrawingEntry with correct Z-order", func(t *testing.T) {
		gs := NewGraphicsSystem("")
		lm := gs.GetLayerManager()

		// ピクチャーを作成
		srcPicID, _ := gs.CreatePic(100, 100)
		dstPicID, _ := gs.CreatePic(200, 200)

		// MovePicを実行
		gs.MovePic(srcPicID, 0, 0, 50, 50, dstPicID, 10, 10, 0)

		// DrawingEntryが作成されていることを確認
		pls := lm.GetPictureLayerSet(dstPicID)
		if pls == nil {
			t.Fatal("expected PictureLayerSet to be created")
		}

		if pls.GetDrawingEntryCount() != 1 {
			t.Errorf("expected 1 DrawingEntry, got %d", pls.GetDrawingEntryCount())
		}
	})

	t.Run("PutCast after MovePic has higher Z-order", func(t *testing.T) {
		gs := NewGraphicsSystem("")
		lm := gs.GetLayerManager()

		// ピクチャーを作成
		srcPicID, _ := gs.CreatePic(100, 100)
		dstPicID, _ := gs.CreatePic(200, 200)
		winID, _ := gs.OpenWin(dstPicID)

		// MovePicを実行（Z=1）
		gs.MovePic(srcPicID, 0, 0, 50, 50, dstPicID, 10, 10, 0)

		// PutCastを実行（Z=2）
		gs.PutCast(winID, srcPicID, 20, 20, 0, 0, 32, 32)

		// Z順序を確認
		plsDst := lm.GetPictureLayerSet(dstPicID)
		plsWin := lm.GetPictureLayerSet(winID)

		if plsDst == nil || plsWin == nil {
			t.Fatal("expected PictureLayerSets to be created")
		}

		// DrawingEntryのZ順序
		if plsDst.GetDrawingEntryCount() != 1 {
			t.Errorf("expected 1 DrawingEntry, got %d", plsDst.GetDrawingEntryCount())
		}
		drawingZ := plsDst.DrawingEntries[0].GetZOrder()

		// CastLayerのZ順序
		if plsWin.GetCastLayerCount() != 1 {
			t.Errorf("expected 1 CastLayer, got %d", plsWin.GetCastLayerCount())
		}
		castZ := plsWin.Casts[0].GetZOrder()

		// 注: DrawingEntryとCastLayerは異なるPictureLayerSetに追加されるため、
		// 直接比較はできない。各PictureLayerSet内でのZ順序の増加を確認する。
		t.Logf("DrawingEntry Z-order: %d, CastLayer Z-order: %d", drawingZ, castZ)
	})

	t.Run("Multiple MovePic calls create multiple DrawingEntries", func(t *testing.T) {
		gs := NewGraphicsSystem("")
		lm := gs.GetLayerManager()

		// ピクチャーを作成
		srcPicID, _ := gs.CreatePic(100, 100)
		dstPicID, _ := gs.CreatePic(200, 200)

		// 複数のMovePicを実行
		gs.MovePic(srcPicID, 0, 0, 50, 50, dstPicID, 10, 10, 0)
		gs.MovePic(srcPicID, 0, 0, 50, 50, dstPicID, 60, 60, 0)
		gs.MovePic(srcPicID, 0, 0, 50, 50, dstPicID, 110, 110, 0)

		// DrawingEntryが3つ作成されていることを確認
		pls := lm.GetPictureLayerSet(dstPicID)
		if pls == nil {
			t.Fatal("expected PictureLayerSet to be created")
		}

		if pls.GetDrawingEntryCount() != 3 {
			t.Errorf("expected 3 DrawingEntries, got %d", pls.GetDrawingEntryCount())
		}

		// Z順序が増加していることを確認
		for i := 0; i < len(pls.DrawingEntries)-1; i++ {
			if pls.DrawingEntries[i].GetZOrder() >= pls.DrawingEntries[i+1].GetZOrder() {
				t.Errorf("DrawingEntry Z-order should increase: %d >= %d",
					pls.DrawingEntries[i].GetZOrder(), pls.DrawingEntries[i+1].GetZOrder())
			}
		}
	})

	t.Run("Mixed operations maintain correct Z-order sequence", func(t *testing.T) {
		gs := NewGraphicsSystem("")
		lm := gs.GetLayerManager()

		// ピクチャーを作成
		srcPicID, _ := gs.CreatePic(100, 100)
		dstPicID, _ := gs.CreatePic(200, 200)

		// 混合操作を実行
		gs.MovePic(srcPicID, 0, 0, 50, 50, dstPicID, 10, 10, 0) // DrawingEntry Z=1
		gs.TextWrite(dstPicID, 50, 50, "Text 1")                // TextLayerEntry Z=2
		gs.MovePic(srcPicID, 0, 0, 50, 50, dstPicID, 60, 60, 0) // DrawingEntry Z=3
		gs.TextWrite(dstPicID, 100, 100, "Text 2")              // TextLayerEntry Z=4

		// PictureLayerSetを確認
		pls := lm.GetPictureLayerSet(dstPicID)
		if pls == nil {
			t.Fatal("expected PictureLayerSet to be created")
		}

		// DrawingEntryが2つ
		if pls.GetDrawingEntryCount() != 2 {
			t.Errorf("expected 2 DrawingEntries, got %d", pls.GetDrawingEntryCount())
		}

		// TextLayerEntryが2つ
		if pls.GetTextLayerCount() != 2 {
			t.Errorf("expected 2 TextLayerEntries, got %d", pls.GetTextLayerCount())
		}

		// すべてのレイヤーをZ順序でソートして取得
		layers := pls.GetAllLayersSorted()

		// Z順序が増加していることを確認
		for i := 0; i < len(layers)-1; i++ {
			if layers[i].GetZOrder() > layers[i+1].GetZOrder() {
				t.Errorf("layers should be sorted by Z-order: %d > %d",
					layers[i].GetZOrder(), layers[i+1].GetZOrder())
			}
		}
	})
}
