package graphics

import (
	"image/color"
	"testing"
)

// TestBakeToPictureLayer_EmptyStack はレイヤースタックが空の場合のテスト
// 要件 3.4: レイヤースタックが空である場合、新しいPicture_Layerを作成
func TestBakeToPictureLayer_EmptyStack(t *testing.T) {
	gs := NewGraphicsSystem("")

	// ソースと転送先のピクチャーを作成
	srcID, err := gs.CreatePic(100, 100)
	if err != nil {
		t.Fatalf("Failed to create source picture: %v", err)
	}

	dstID, err := gs.CreatePic(200, 200)
	if err != nil {
		t.Fatalf("Failed to create destination picture: %v", err)
	}

	// ウィンドウを開く（転送先ピクチャーに関連付け）
	winID, err := gs.OpenWin(dstID, WithSize(200, 200))
	if err != nil {
		t.Fatalf("Failed to open window: %v", err)
	}

	// WindowLayerSetが空であることを確認
	wls := gs.layerManager.GetWindowLayerSet(winID)
	if wls != nil && wls.GetLayerCount() > 0 {
		t.Fatalf("Expected empty layer stack, got %d layers", wls.GetLayerCount())
	}

	// MovePicを実行
	err = gs.MovePic(srcID, 0, 0, 50, 50, dstID, 10, 10, 0)
	if err != nil {
		t.Fatalf("MovePic failed: %v", err)
	}

	// WindowLayerSetが作成され、PictureLayerが追加されていることを確認
	wls = gs.layerManager.GetWindowLayerSet(winID)
	if wls == nil {
		t.Fatal("WindowLayerSet should be created")
	}

	if wls.GetLayerCount() != 1 {
		t.Errorf("Expected 1 layer, got %d", wls.GetLayerCount())
	}

	// 最上位レイヤーがPictureLayerであることを確認
	topmost := wls.GetTopmostLayer()
	if topmost == nil {
		t.Fatal("Topmost layer should not be nil")
	}

	if topmost.GetLayerType() != LayerTypePicture {
		t.Errorf("Expected LayerTypePicture, got %v", topmost.GetLayerType())
	}
}

// TestBakeToPictureLayer_ExistingPictureLayer は最上位がPictureLayerの場合のテスト
// 要件 3.2: 最上位レイヤーがPicture_Layerである場合、そのレイヤーに画像を焼き付ける
func TestBakeToPictureLayer_ExistingPictureLayer(t *testing.T) {
	gs := NewGraphicsSystem("")

	// ソースと転送先のピクチャーを作成
	srcID, err := gs.CreatePic(100, 100)
	if err != nil {
		t.Fatalf("Failed to create source picture: %v", err)
	}

	dstID, err := gs.CreatePic(200, 200)
	if err != nil {
		t.Fatalf("Failed to create destination picture: %v", err)
	}

	// ウィンドウを開く
	winID, err := gs.OpenWin(dstID, WithSize(200, 200))
	if err != nil {
		t.Fatalf("Failed to open window: %v", err)
	}

	// 最初のMovePicでPictureLayerを作成
	err = gs.MovePic(srcID, 0, 0, 50, 50, dstID, 10, 10, 0)
	if err != nil {
		t.Fatalf("First MovePic failed: %v", err)
	}

	wls := gs.layerManager.GetWindowLayerSet(winID)
	if wls == nil {
		t.Fatal("WindowLayerSet should be created")
	}

	initialLayerCount := wls.GetLayerCount()
	if initialLayerCount != 1 {
		t.Errorf("Expected 1 layer after first MovePic, got %d", initialLayerCount)
	}

	// 2回目のMovePicで同じPictureLayerに焼き付け
	err = gs.MovePic(srcID, 0, 0, 30, 30, dstID, 50, 50, 0)
	if err != nil {
		t.Fatalf("Second MovePic failed: %v", err)
	}

	// レイヤー数が増えていないことを確認（同じPictureLayerに焼き付けられた）
	if wls.GetLayerCount() != initialLayerCount {
		t.Errorf("Expected layer count to remain %d, got %d", initialLayerCount, wls.GetLayerCount())
	}
}

// TestBakeToPictureLayer_TopmostIsCastLayer は最上位がCastLayerの場合のテスト
// 要件 3.3: 最上位レイヤーがCast_Layerである場合、新しいPicture_Layerを作成
func TestBakeToPictureLayer_TopmostIsCastLayer(t *testing.T) {
	gs := NewGraphicsSystem("")

	// ソースと転送先のピクチャーを作成
	srcID, err := gs.CreatePic(100, 100)
	if err != nil {
		t.Fatalf("Failed to create source picture: %v", err)
	}

	dstID, err := gs.CreatePic(200, 200)
	if err != nil {
		t.Fatalf("Failed to create destination picture: %v", err)
	}

	// ウィンドウを開く
	winID, err := gs.OpenWin(dstID, WithSize(200, 200))
	if err != nil {
		t.Fatalf("Failed to open window: %v", err)
	}

	// WindowLayerSetを取得または作成
	wls := gs.layerManager.GetOrCreateWindowLayerSet(winID, 200, 200, color.Black)

	// CastLayerを追加
	// NewCastLayer(id, castID, picID, srcPicID, x, y, srcX, srcY, width, height, zOrderOffset)
	castLayer := NewCastLayer(gs.layerManager.GetNextLayerID(), 1, dstID, srcID, 10, 10, 0, 0, 50, 50, 0)
	wls.AddLayer(castLayer)

	initialLayerCount := wls.GetLayerCount()
	if initialLayerCount != 1 {
		t.Errorf("Expected 1 layer after adding CastLayer, got %d", initialLayerCount)
	}

	// 最上位がCastLayerであることを確認
	topmost := wls.GetTopmostLayer()
	if topmost.GetLayerType() != LayerTypeCast {
		t.Errorf("Expected topmost to be CastLayer, got %v", topmost.GetLayerType())
	}

	// MovePicを実行
	err = gs.MovePic(srcID, 0, 0, 50, 50, dstID, 10, 10, 0)
	if err != nil {
		t.Fatalf("MovePic failed: %v", err)
	}

	// 新しいPictureLayerが追加されていることを確認
	if wls.GetLayerCount() != initialLayerCount+1 {
		t.Errorf("Expected layer count to be %d, got %d", initialLayerCount+1, wls.GetLayerCount())
	}

	// 最上位がPictureLayerであることを確認
	topmost = wls.GetTopmostLayer()
	if topmost.GetLayerType() != LayerTypePicture {
		t.Errorf("Expected topmost to be PictureLayer, got %v", topmost.GetLayerType())
	}
}

// TestBakeToPictureLayer_TopmostIsTextLayer は最上位がTextLayerの場合のテスト
// 要件 3.3: 最上位レイヤーがText_Layerである場合、新しいPicture_Layerを作成
func TestBakeToPictureLayer_TopmostIsTextLayer(t *testing.T) {
	gs := NewGraphicsSystem("")

	// ソースと転送先のピクチャーを作成
	srcID, err := gs.CreatePic(100, 100)
	if err != nil {
		t.Fatalf("Failed to create source picture: %v", err)
	}

	dstID, err := gs.CreatePic(200, 200)
	if err != nil {
		t.Fatalf("Failed to create destination picture: %v", err)
	}

	// ウィンドウを開く
	winID, err := gs.OpenWin(dstID, WithSize(200, 200))
	if err != nil {
		t.Fatalf("Failed to open window: %v", err)
	}

	// WindowLayerSetを取得または作成
	wls := gs.layerManager.GetOrCreateWindowLayerSet(winID, 200, 200, color.Black)

	// TextLayerEntryを追加
	textLayer := NewTextLayerEntry(gs.layerManager.GetNextLayerID(), dstID, 10, 10, "test", 0)
	wls.AddLayer(textLayer)

	initialLayerCount := wls.GetLayerCount()
	if initialLayerCount != 1 {
		t.Errorf("Expected 1 layer after adding TextLayer, got %d", initialLayerCount)
	}

	// 最上位がTextLayerであることを確認
	topmost := wls.GetTopmostLayer()
	if topmost.GetLayerType() != LayerTypeText {
		t.Errorf("Expected topmost to be TextLayer, got %v", topmost.GetLayerType())
	}

	// MovePicを実行
	err = gs.MovePic(srcID, 0, 0, 50, 50, dstID, 10, 10, 0)
	if err != nil {
		t.Fatalf("MovePic failed: %v", err)
	}

	// 新しいPictureLayerが追加されていることを確認
	if wls.GetLayerCount() != initialLayerCount+1 {
		t.Errorf("Expected layer count to be %d, got %d", initialLayerCount+1, wls.GetLayerCount())
	}

	// 最上位がPictureLayerであることを確認
	topmost = wls.GetTopmostLayer()
	if topmost.GetLayerType() != LayerTypePicture {
		t.Errorf("Expected topmost to be PictureLayer, got %v", topmost.GetLayerType())
	}
}

// TestBakeToPictureLayer_NoWindow はウィンドウがない場合のフォールバックテスト
// ウィンドウが見つからない場合は従来のDrawingEntry方式にフォールバック
func TestBakeToPictureLayer_NoWindow(t *testing.T) {
	gs := NewGraphicsSystem("")

	// ソースと転送先のピクチャーを作成（ウィンドウは開かない）
	srcID, err := gs.CreatePic(100, 100)
	if err != nil {
		t.Fatalf("Failed to create source picture: %v", err)
	}

	dstID, err := gs.CreatePic(200, 200)
	if err != nil {
		t.Fatalf("Failed to create destination picture: %v", err)
	}

	// MovePicを実行（ウィンドウがないのでフォールバック）
	err = gs.MovePic(srcID, 0, 0, 50, 50, dstID, 10, 10, 0)
	if err != nil {
		t.Fatalf("MovePic failed: %v", err)
	}

	// PictureLayerSetにDrawingEntryが追加されていることを確認
	pls := gs.layerManager.GetPictureLayerSet(dstID)
	if pls == nil {
		t.Fatal("PictureLayerSet should be created for fallback")
	}

	if pls.GetDrawingEntryCount() != 1 {
		t.Errorf("Expected 1 DrawingEntry, got %d", pls.GetDrawingEntryCount())
	}
}

// TestBakeToPictureLayer_DirtyFlag は焼き付け後のダーティフラグテスト
// 要件 3.6: 焼き付けが行われたとき、焼き付け先レイヤーをダーティとしてマークする
func TestBakeToPictureLayer_DirtyFlag(t *testing.T) {
	gs := NewGraphicsSystem("")

	// ソースと転送先のピクチャーを作成
	srcID, err := gs.CreatePic(100, 100)
	if err != nil {
		t.Fatalf("Failed to create source picture: %v", err)
	}

	dstID, err := gs.CreatePic(200, 200)
	if err != nil {
		t.Fatalf("Failed to create destination picture: %v", err)
	}

	// ウィンドウを開く
	winID, err := gs.OpenWin(dstID, WithSize(200, 200))
	if err != nil {
		t.Fatalf("Failed to open window: %v", err)
	}

	// MovePicを実行
	err = gs.MovePic(srcID, 0, 0, 50, 50, dstID, 10, 10, 0)
	if err != nil {
		t.Fatalf("MovePic failed: %v", err)
	}

	// WindowLayerSetを取得
	wls := gs.layerManager.GetWindowLayerSet(winID)
	if wls == nil {
		t.Fatal("WindowLayerSet should be created")
	}

	// 最上位レイヤーがダーティであることを確認
	topmost := wls.GetTopmostLayer()
	if topmost == nil {
		t.Fatal("Topmost layer should not be nil")
	}

	if !topmost.IsDirty() {
		t.Error("Topmost layer should be dirty after baking")
	}
}

// TestBakeToPictureLayer_TransparentMode は透明色モードのテスト
func TestBakeToPictureLayer_TransparentMode(t *testing.T) {
	gs := NewGraphicsSystem("")

	// ソースと転送先のピクチャーを作成
	srcID, err := gs.CreatePic(100, 100)
	if err != nil {
		t.Fatalf("Failed to create source picture: %v", err)
	}

	dstID, err := gs.CreatePic(200, 200)
	if err != nil {
		t.Fatalf("Failed to create destination picture: %v", err)
	}

	// ウィンドウを開く
	winID, err := gs.OpenWin(dstID, WithSize(200, 200))
	if err != nil {
		t.Fatalf("Failed to open window: %v", err)
	}

	// 透明色モードでMovePicを実行
	err = gs.MovePic(srcID, 0, 0, 50, 50, dstID, 10, 10, 1)
	if err != nil {
		t.Fatalf("MovePic with transparent mode failed: %v", err)
	}

	// WindowLayerSetが作成されていることを確認
	wls := gs.layerManager.GetWindowLayerSet(winID)
	if wls == nil {
		t.Fatal("WindowLayerSet should be created")
	}

	if wls.GetLayerCount() != 1 {
		t.Errorf("Expected 1 layer, got %d", wls.GetLayerCount())
	}
}
