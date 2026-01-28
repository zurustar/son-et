package graphics

import (
	"testing"
)

func TestNewCastManager(t *testing.T) {
	cm := NewCastManager()
	if cm == nil {
		t.Fatal("NewCastManager returned nil")
	}
	if cm.Count() != 0 {
		t.Errorf("expected 0 casts, got %d", cm.Count())
	}
}

func TestPutCast(t *testing.T) {
	cm := NewCastManager()

	id, err := cm.PutCast(0, 1, 10, 20, 0, 0, 32, 32)
	if err != nil {
		t.Fatalf("PutCast failed: %v", err)
	}
	if id != 0 {
		t.Errorf("expected ID 0, got %d", id)
	}

	cast, err := cm.GetCast(id)
	if err != nil {
		t.Fatalf("GetCast failed: %v", err)
	}
	if cast.WinID != 0 {
		t.Errorf("expected WinID 0, got %d", cast.WinID)
	}
	if cast.PicID != 1 {
		t.Errorf("expected PicID 1, got %d", cast.PicID)
	}
	if cast.X != 10 || cast.Y != 20 {
		t.Errorf("expected position (10, 20), got (%d, %d)", cast.X, cast.Y)
	}
	if cast.Width != 32 || cast.Height != 32 {
		t.Errorf("expected size (32, 32), got (%d, %d)", cast.Width, cast.Height)
	}
	if !cast.Visible {
		t.Error("expected cast to be visible")
	}
}

func TestMoveCast(t *testing.T) {
	cm := NewCastManager()

	id, _ := cm.PutCast(0, 1, 10, 20, 0, 0, 32, 32)

	// Move position
	err := cm.MoveCast(id, WithCastPosition(100, 200))
	if err != nil {
		t.Fatalf("MoveCast failed: %v", err)
	}

	cast, _ := cm.GetCast(id)
	if cast.X != 100 || cast.Y != 200 {
		t.Errorf("expected position (100, 200), got (%d, %d)", cast.X, cast.Y)
	}

	// Move source
	err = cm.MoveCast(id, WithCastSource(10, 10, 64, 64))
	if err != nil {
		t.Fatalf("MoveCast failed: %v", err)
	}

	cast, _ = cm.GetCast(id)
	if cast.SrcX != 10 || cast.SrcY != 10 {
		t.Errorf("expected source (10, 10), got (%d, %d)", cast.SrcX, cast.SrcY)
	}
	if cast.Width != 64 || cast.Height != 64 {
		t.Errorf("expected size (64, 64), got (%d, %d)", cast.Width, cast.Height)
	}

	// Change picture ID
	err = cm.MoveCast(id, WithCastPicID(5))
	if err != nil {
		t.Fatalf("MoveCast failed: %v", err)
	}

	cast, _ = cm.GetCast(id)
	if cast.PicID != 5 {
		t.Errorf("expected PicID 5, got %d", cast.PicID)
	}
}

func TestMoveCastNotFound(t *testing.T) {
	cm := NewCastManager()

	err := cm.MoveCast(999, WithCastPosition(0, 0))
	if err == nil {
		t.Error("expected error for non-existent cast")
	}
}

func TestDelCast(t *testing.T) {
	cm := NewCastManager()

	id, _ := cm.PutCast(0, 1, 10, 20, 0, 0, 32, 32)

	err := cm.DelCast(id)
	if err != nil {
		t.Fatalf("DelCast failed: %v", err)
	}

	_, err = cm.GetCast(id)
	if err == nil {
		t.Error("expected error for deleted cast")
	}

	if cm.Count() != 0 {
		t.Errorf("expected 0 casts, got %d", cm.Count())
	}
}

func TestDelCastNotFound(t *testing.T) {
	cm := NewCastManager()

	err := cm.DelCast(999)
	if err == nil {
		t.Error("expected error for non-existent cast")
	}
}

func TestGetCastsByWindow(t *testing.T) {
	cm := NewCastManager()

	// Create casts for window 0
	cm.PutCast(0, 1, 10, 10, 0, 0, 32, 32)
	cm.PutCast(0, 2, 20, 20, 0, 0, 32, 32)

	// Create casts for window 1
	cm.PutCast(1, 3, 30, 30, 0, 0, 32, 32)

	casts := cm.GetCastsByWindow(0)
	if len(casts) != 2 {
		t.Errorf("expected 2 casts for window 0, got %d", len(casts))
	}

	casts = cm.GetCastsByWindow(1)
	if len(casts) != 1 {
		t.Errorf("expected 1 cast for window 1, got %d", len(casts))
	}

	casts = cm.GetCastsByWindow(2)
	if len(casts) != 0 {
		t.Errorf("expected 0 casts for window 2, got %d", len(casts))
	}
}

func TestDeleteCastsByWindow(t *testing.T) {
	cm := NewCastManager()

	// Create casts for window 0
	cm.PutCast(0, 1, 10, 10, 0, 0, 32, 32)
	cm.PutCast(0, 2, 20, 20, 0, 0, 32, 32)

	// Create casts for window 1
	cm.PutCast(1, 3, 30, 30, 0, 0, 32, 32)

	cm.DeleteCastsByWindow(0)

	casts := cm.GetCastsByWindow(0)
	if len(casts) != 0 {
		t.Errorf("expected 0 casts for window 0 after delete, got %d", len(casts))
	}

	casts = cm.GetCastsByWindow(1)
	if len(casts) != 1 {
		t.Errorf("expected 1 cast for window 1 after delete, got %d", len(casts))
	}

	if cm.Count() != 1 {
		t.Errorf("expected 1 total cast, got %d", cm.Count())
	}
}

func TestCastZOrder(t *testing.T) {
	cm := NewCastManager()

	id1, _ := cm.PutCast(0, 1, 10, 10, 0, 0, 32, 32)
	id2, _ := cm.PutCast(0, 2, 20, 20, 0, 0, 32, 32)
	id3, _ := cm.PutCast(0, 3, 30, 30, 0, 0, 32, 32)

	cast1, _ := cm.GetCast(id1)
	cast2, _ := cm.GetCast(id2)
	cast3, _ := cm.GetCast(id3)

	if cast1.ZOrder >= cast2.ZOrder {
		t.Error("cast1 should have lower ZOrder than cast2")
	}
	if cast2.ZOrder >= cast3.ZOrder {
		t.Error("cast2 should have lower ZOrder than cast3")
	}

	// Verify GetCastsByWindow returns sorted by ZOrder
	casts := cm.GetCastsByWindow(0)
	for i := 0; i < len(casts)-1; i++ {
		if casts[i].ZOrder >= casts[i+1].ZOrder {
			t.Error("casts should be sorted by ZOrder")
		}
	}
}

func TestCastResourceLimit(t *testing.T) {
	cm := NewCastManager()

	// Create 1024 casts
	for i := 0; i < 1024; i++ {
		_, err := cm.PutCast(0, i%256, i*10, i*10, 0, 0, 32, 32)
		if err != nil {
			t.Fatalf("PutCast failed at %d: %v", i, err)
		}
	}

	// Try to create one more
	_, err := cm.PutCast(0, 0, 0, 0, 0, 0, 32, 32)
	if err == nil {
		t.Error("expected error when exceeding resource limit")
	}

	if cm.Count() != 1024 {
		t.Errorf("expected 1024 casts, got %d", cm.Count())
	}
}

// TestCastManagerWindowLayerSetIntegration tests the integration between CastManager and WindowLayerSet
// 要件 7.1: Cast_LayerをウィンドウIDで登録する
func TestCastManagerWindowLayerSetIntegration(t *testing.T) {
	t.Run("PutCast creates CastLayer in WindowLayerSet", func(t *testing.T) {
		// 要件 7.1: PutCastが呼び出されたときにCast_LayerをウィンドウIDで登録する
		cm := NewCastManager()
		lm := NewLayerManager()
		cm.SetLayerManager(lm)

		// WindowLayerSetを事前に作成
		winID := 0
		wls := lm.GetOrCreateWindowLayerSet(winID, 640, 480, nil)
		if wls == nil {
			t.Fatal("expected WindowLayerSet to be created")
		}

		// Create a cast
		castID, err := cm.PutCast(winID, 1, 10, 20, 0, 0, 32, 32)
		if err != nil {
			t.Fatalf("PutCast failed: %v", err)
		}

		// Verify CastLayer was created in WindowLayerSet
		if wls.GetLayerCount() != 1 {
			t.Errorf("expected 1 layer in WindowLayerSet, got %d", wls.GetLayerCount())
		}

		// Verify the layer is a CastLayer with correct properties
		layers := wls.GetLayers()
		if len(layers) != 1 {
			t.Fatalf("expected 1 layer, got %d", len(layers))
		}

		castLayer, ok := layers[0].(*CastLayer)
		if !ok {
			t.Fatal("expected layer to be CastLayer")
		}

		if castLayer.GetCastID() != castID {
			t.Errorf("expected castID %d, got %d", castID, castLayer.GetCastID())
		}

		x, y := castLayer.GetPosition()
		if x != 10 || y != 20 {
			t.Errorf("expected position (10, 20), got (%d, %d)", x, y)
		}
	})

	t.Run("PutCast with WindowLayerSet assigns Z-order correctly", func(t *testing.T) {
		// 要件 6.3: 新しいレイヤーが作成されたときに現在のZ順序カウンターを割り当て
		cm := NewCastManager()
		lm := NewLayerManager()
		cm.SetLayerManager(lm)

		// WindowLayerSetを事前に作成
		winID := 0
		wls := lm.GetOrCreateWindowLayerSet(winID, 640, 480, nil)

		// Create multiple casts
		castID1, _ := cm.PutCast(winID, 1, 10, 10, 0, 0, 32, 32)
		castID2, _ := cm.PutCast(winID, 2, 20, 20, 0, 0, 32, 32)
		castID3, _ := cm.PutCast(winID, 3, 30, 30, 0, 0, 32, 32)

		// Verify Z-order is assigned correctly
		if wls.GetLayerCount() != 3 {
			t.Errorf("expected 3 layers, got %d", wls.GetLayerCount())
		}

		layers := wls.GetLayers()
		var layer1, layer2, layer3 *CastLayer
		for _, l := range layers {
			cl := l.(*CastLayer)
			switch cl.GetCastID() {
			case castID1:
				layer1 = cl
			case castID2:
				layer2 = cl
			case castID3:
				layer3 = cl
			}
		}

		if layer1 == nil || layer2 == nil || layer3 == nil {
			t.Fatal("expected all layers to be found")
		}

		// Z-order should be increasing
		if layer1.GetZOrder() >= layer2.GetZOrder() {
			t.Error("layer1 should have lower Z-order than layer2")
		}
		if layer2.GetZOrder() >= layer3.GetZOrder() {
			t.Error("layer2 should have lower Z-order than layer3")
		}
	})

	t.Run("PutCast falls back to PictureLayerSet when WindowLayerSet not exists", func(t *testing.T) {
		// 後方互換性: WindowLayerSetが存在しない場合はPictureLayerSetを使用
		cm := NewCastManager()
		lm := NewLayerManager()
		cm.SetLayerManager(lm)

		// WindowLayerSetを作成しない
		winID := 0

		// Create a cast
		castID, err := cm.PutCast(winID, 1, 10, 20, 0, 0, 32, 32)
		if err != nil {
			t.Fatalf("PutCast failed: %v", err)
		}

		// Verify CastLayer was created in PictureLayerSet (fallback)
		pls := lm.GetPictureLayerSet(winID)
		if pls == nil {
			t.Fatal("expected PictureLayerSet to be created as fallback")
		}

		castLayer := pls.GetCastLayer(castID)
		if castLayer == nil {
			t.Fatal("expected CastLayer to be created in PictureLayerSet")
		}
	})

	t.Run("PutCastWithTransColor creates CastLayer in WindowLayerSet", func(t *testing.T) {
		// 要件 7.1: 透明色付きでもWindowLayerSetに登録
		cm := NewCastManager()
		lm := NewLayerManager()
		cm.SetLayerManager(lm)

		// WindowLayerSetを事前に作成
		winID := 0
		wls := lm.GetOrCreateWindowLayerSet(winID, 640, 480, nil)

		transColor := DefaultTransparentColor

		// Create a cast with transparent color
		castID, err := cm.PutCastWithTransColor(winID, 1, 10, 20, 0, 0, 32, 32, transColor)
		if err != nil {
			t.Fatalf("PutCastWithTransColor failed: %v", err)
		}

		// Verify CastLayer was created in WindowLayerSet
		if wls.GetLayerCount() != 1 {
			t.Errorf("expected 1 layer in WindowLayerSet, got %d", wls.GetLayerCount())
		}

		layers := wls.GetLayers()
		castLayer, ok := layers[0].(*CastLayer)
		if !ok {
			t.Fatal("expected layer to be CastLayer")
		}

		if castLayer.GetCastID() != castID {
			t.Errorf("expected castID %d, got %d", castID, castLayer.GetCastID())
		}

		if !castLayer.HasTransColor() {
			t.Error("expected CastLayer to have trans color")
		}
	})

	t.Run("Multiple windows have separate WindowLayerSets", func(t *testing.T) {
		// 要件 7.1: 各ウィンドウに独立したレイヤー管理
		cm := NewCastManager()
		lm := NewLayerManager()
		cm.SetLayerManager(lm)

		// Create WindowLayerSets for two windows
		wls0 := lm.GetOrCreateWindowLayerSet(0, 640, 480, nil)
		wls1 := lm.GetOrCreateWindowLayerSet(1, 640, 480, nil)

		// Create casts for window 0
		cm.PutCast(0, 1, 10, 10, 0, 0, 32, 32)
		cm.PutCast(0, 2, 20, 20, 0, 0, 32, 32)

		// Create casts for window 1
		cm.PutCast(1, 3, 30, 30, 0, 0, 32, 32)

		// Verify each WindowLayerSet has correct number of layers
		if wls0.GetLayerCount() != 2 {
			t.Errorf("expected 2 layers in window 0, got %d", wls0.GetLayerCount())
		}
		if wls1.GetLayerCount() != 1 {
			t.Errorf("expected 1 layer in window 1, got %d", wls1.GetLayerCount())
		}
	})
}

// TestCastManagerLayerManagerIntegration tests the integration between CastManager and LayerManager
// 要件 8.2: CastManagerとLayerManagerを統合する
func TestCastManagerLayerManagerIntegration(t *testing.T) {
	t.Run("SetLayerManager", func(t *testing.T) {
		cm := NewCastManager()
		lm := NewLayerManager()

		// Initially nil
		if cm.GetLayerManager() != nil {
			t.Error("expected nil LayerManager initially")
		}

		// Set LayerManager
		cm.SetLayerManager(lm)
		if cm.GetLayerManager() != lm {
			t.Error("expected LayerManager to be set")
		}
	})

	t.Run("PutCast creates CastLayer", func(t *testing.T) {
		// 要件 2.1: PutCastが呼び出されたときに対応するCast_Layerを作成する
		cm := NewCastManager()
		lm := NewLayerManager()
		cm.SetLayerManager(lm)

		// Create a cast
		castID, err := cm.PutCast(0, 1, 10, 20, 0, 0, 32, 32)
		if err != nil {
			t.Fatalf("PutCast failed: %v", err)
		}

		// Verify CastLayer was created
		pls := lm.GetPictureLayerSet(0)
		if pls == nil {
			t.Fatal("expected PictureLayerSet to be created")
		}

		castLayer := pls.GetCastLayer(castID)
		if castLayer == nil {
			t.Fatal("expected CastLayer to be created")
		}

		// Verify CastLayer properties
		x, y := castLayer.GetPosition()
		if x != 10 || y != 20 {
			t.Errorf("expected position (10, 20), got (%d, %d)", x, y)
		}

		srcX, srcY, w, h := castLayer.GetSourceRect()
		if srcX != 0 || srcY != 0 || w != 32 || h != 32 {
			t.Errorf("expected source rect (0, 0, 32, 32), got (%d, %d, %d, %d)", srcX, srcY, w, h)
		}
	})

	t.Run("PutCastWithTransColor creates CastLayer with trans color", func(t *testing.T) {
		// 要件 2.1: PutCastが呼び出されたときに対応するCast_Layerを作成する
		cm := NewCastManager()
		lm := NewLayerManager()
		cm.SetLayerManager(lm)

		transColor := DefaultTransparentColor

		// Create a cast with transparent color
		castID, err := cm.PutCastWithTransColor(0, 1, 10, 20, 0, 0, 32, 32, transColor)
		if err != nil {
			t.Fatalf("PutCastWithTransColor failed: %v", err)
		}

		// Verify CastLayer was created with trans color
		pls := lm.GetPictureLayerSet(0)
		castLayer := pls.GetCastLayer(castID)
		if castLayer == nil {
			t.Fatal("expected CastLayer to be created")
		}

		if !castLayer.HasTransColor() {
			t.Error("expected CastLayer to have trans color")
		}
	})

	t.Run("MoveCast updates CastLayer position", func(t *testing.T) {
		// 要件 2.2: MoveCastが呼び出されたときに対応するCast_Layerの位置を更新する
		cm := NewCastManager()
		lm := NewLayerManager()
		cm.SetLayerManager(lm)

		// Create a cast
		castID, _ := cm.PutCast(0, 1, 10, 20, 0, 0, 32, 32)

		// Move the cast
		err := cm.MoveCast(castID, WithCastPosition(100, 200))
		if err != nil {
			t.Fatalf("MoveCast failed: %v", err)
		}

		// Verify CastLayer position was updated
		pls := lm.GetPictureLayerSet(0)
		castLayer := pls.GetCastLayer(castID)

		x, y := castLayer.GetPosition()
		if x != 100 || y != 200 {
			t.Errorf("expected position (100, 200), got (%d, %d)", x, y)
		}
	})

	t.Run("MoveCast updates CastLayer source rect", func(t *testing.T) {
		// 要件 2.2: MoveCastが呼び出されたときに対応するCast_Layerの位置を更新する
		cm := NewCastManager()
		lm := NewLayerManager()
		cm.SetLayerManager(lm)

		// Create a cast
		castID, _ := cm.PutCast(0, 1, 10, 20, 0, 0, 32, 32)

		// Move the cast source
		err := cm.MoveCast(castID, WithCastSource(10, 10, 64, 64))
		if err != nil {
			t.Fatalf("MoveCast failed: %v", err)
		}

		// Verify CastLayer source rect was updated
		pls := lm.GetPictureLayerSet(0)
		castLayer := pls.GetCastLayer(castID)

		srcX, srcY, w, h := castLayer.GetSourceRect()
		if srcX != 10 || srcY != 10 || w != 64 || h != 64 {
			t.Errorf("expected source rect (10, 10, 64, 64), got (%d, %d, %d, %d)", srcX, srcY, w, h)
		}
	})

	t.Run("DelCast removes CastLayer", func(t *testing.T) {
		// 要件 2.3: DelCastが呼び出されたときに対応するCast_Layerを削除する
		cm := NewCastManager()
		lm := NewLayerManager()
		cm.SetLayerManager(lm)

		// Create a cast
		castID, _ := cm.PutCast(0, 1, 10, 20, 0, 0, 32, 32)

		// Verify CastLayer exists
		pls := lm.GetPictureLayerSet(0)
		if pls.GetCastLayer(castID) == nil {
			t.Fatal("expected CastLayer to exist before deletion")
		}

		// Delete the cast
		err := cm.DelCast(castID)
		if err != nil {
			t.Fatalf("DelCast failed: %v", err)
		}

		// Verify CastLayer was removed
		if pls.GetCastLayer(castID) != nil {
			t.Error("expected CastLayer to be removed after deletion")
		}
	})

	t.Run("DeleteCastsByWindow removes all CastLayers", func(t *testing.T) {
		// 要件 2.6: ウィンドウが閉じられたときにそのウィンドウに属するすべてのレイヤーを削除する
		cm := NewCastManager()
		lm := NewLayerManager()
		cm.SetLayerManager(lm)

		// Create casts for window 0
		cm.PutCast(0, 1, 10, 10, 0, 0, 32, 32)
		cm.PutCast(0, 2, 20, 20, 0, 0, 32, 32)

		// Create casts for window 1
		cm.PutCast(1, 3, 30, 30, 0, 0, 32, 32)

		// Verify CastLayers exist
		pls0 := lm.GetPictureLayerSet(0)
		if pls0.GetCastLayerCount() != 2 {
			t.Errorf("expected 2 CastLayers for window 0, got %d", pls0.GetCastLayerCount())
		}

		pls1 := lm.GetPictureLayerSet(1)
		if pls1.GetCastLayerCount() != 1 {
			t.Errorf("expected 1 CastLayer for window 1, got %d", pls1.GetCastLayerCount())
		}

		// Delete casts for window 0
		cm.DeleteCastsByWindow(0)

		// Verify CastLayers for window 0 were removed
		if pls0.GetCastLayerCount() != 0 {
			t.Errorf("expected 0 CastLayers for window 0 after delete, got %d", pls0.GetCastLayerCount())
		}

		// Verify CastLayers for window 1 still exist
		if pls1.GetCastLayerCount() != 1 {
			t.Errorf("expected 1 CastLayer for window 1 after delete, got %d", pls1.GetCastLayerCount())
		}
	})

	t.Run("Without LayerManager still works", func(t *testing.T) {
		// CastManager should work independently without LayerManager
		cm := NewCastManager()

		// Create a cast without LayerManager
		castID, err := cm.PutCast(0, 1, 10, 20, 0, 0, 32, 32)
		if err != nil {
			t.Fatalf("PutCast failed: %v", err)
		}

		// Move the cast
		err = cm.MoveCast(castID, WithCastPosition(100, 200))
		if err != nil {
			t.Fatalf("MoveCast failed: %v", err)
		}

		// Delete the cast
		err = cm.DelCast(castID)
		if err != nil {
			t.Fatalf("DelCast failed: %v", err)
		}
	})

	t.Run("Multiple casts create multiple CastLayers", func(t *testing.T) {
		cm := NewCastManager()
		lm := NewLayerManager()
		cm.SetLayerManager(lm)

		// Create multiple casts
		castID1, _ := cm.PutCast(0, 1, 10, 10, 0, 0, 32, 32)
		castID2, _ := cm.PutCast(0, 2, 20, 20, 0, 0, 32, 32)
		castID3, _ := cm.PutCast(0, 3, 30, 30, 0, 0, 32, 32)

		// Verify all CastLayers were created
		pls := lm.GetPictureLayerSet(0)
		if pls.GetCastLayerCount() != 3 {
			t.Errorf("expected 3 CastLayers, got %d", pls.GetCastLayerCount())
		}

		// Verify each CastLayer exists
		if pls.GetCastLayer(castID1) == nil {
			t.Error("expected CastLayer for castID1")
		}
		if pls.GetCastLayer(castID2) == nil {
			t.Error("expected CastLayer for castID2")
		}
		if pls.GetCastLayer(castID3) == nil {
			t.Error("expected CastLayer for castID3")
		}
	})

	t.Run("CastLayer Z-order is correct", func(t *testing.T) {
		cm := NewCastManager()
		lm := NewLayerManager()
		cm.SetLayerManager(lm)

		// Create multiple casts
		castID1, _ := cm.PutCast(0, 1, 10, 10, 0, 0, 32, 32)
		castID2, _ := cm.PutCast(0, 2, 20, 20, 0, 0, 32, 32)
		castID3, _ := cm.PutCast(0, 3, 30, 30, 0, 0, 32, 32)

		// Verify Z-order
		pls := lm.GetPictureLayerSet(0)
		layer1 := pls.GetCastLayer(castID1)
		layer2 := pls.GetCastLayer(castID2)
		layer3 := pls.GetCastLayer(castID3)

		if layer1.GetZOrder() >= layer2.GetZOrder() {
			t.Error("layer1 should have lower Z-order than layer2")
		}
		if layer2.GetZOrder() >= layer3.GetZOrder() {
			t.Error("layer2 should have lower Z-order than layer3")
		}

		// 操作順序に基づくZ順序: 最初のキャストはZ=1から開始
		// 要件 10.1, 10.2: 操作順序に基づくZ順序
		if layer1.GetZOrder() < 1 {
			t.Errorf("layer1 Z-order should be >= 1, got %d", layer1.GetZOrder())
		}
	})
}

// TestMoveCastWithWindowLayerSet tests MoveCast with WindowLayerSet
// 要件 4.2: MoveCastが呼び出されたときにCast_Layerの位置を更新する（残像なし）
func TestMoveCastWithWindowLayerSet(t *testing.T) {
	t.Run("MoveCast updates CastLayer position in WindowLayerSet", func(t *testing.T) {
		// 要件 4.2: Cast_Layerの位置を更新する
		cm := NewCastManager()
		lm := NewLayerManager()
		cm.SetLayerManager(lm)

		// WindowLayerSetを事前に作成
		winID := 0
		wls := lm.GetOrCreateWindowLayerSet(winID, 640, 480, nil)

		// Create a cast
		castID, err := cm.PutCast(winID, 1, 10, 20, 0, 0, 32, 32)
		if err != nil {
			t.Fatalf("PutCast failed: %v", err)
		}

		// Verify initial position
		castLayer := wls.GetCastLayer(castID)
		if castLayer == nil {
			t.Fatal("expected CastLayer to be created in WindowLayerSet")
		}

		x, y := castLayer.GetPosition()
		if x != 10 || y != 20 {
			t.Errorf("expected initial position (10, 20), got (%d, %d)", x, y)
		}

		// Move the cast
		err = cm.MoveCast(castID, WithCastPosition(100, 200))
		if err != nil {
			t.Fatalf("MoveCast failed: %v", err)
		}

		// Verify position was updated
		x, y = castLayer.GetPosition()
		if x != 100 || y != 200 {
			t.Errorf("expected position (100, 200), got (%d, %d)", x, y)
		}
	})

	t.Run("MoveCast updates dirty region in WindowLayerSet", func(t *testing.T) {
		// 要件 4.6: 古い位置と新しい位置をダーティ領域としてマークする
		cm := NewCastManager()
		lm := NewLayerManager()
		cm.SetLayerManager(lm)

		// WindowLayerSetを事前に作成
		winID := 0
		wls := lm.GetOrCreateWindowLayerSet(winID, 640, 480, nil)

		// Clear initial dirty state
		wls.ClearDirtyRegion()

		// Create a cast
		castID, _ := cm.PutCast(winID, 1, 10, 20, 0, 0, 32, 32)

		// Clear dirty state after creation
		wls.ClearDirtyRegion()

		// Move the cast
		err := cm.MoveCast(castID, WithCastPosition(100, 200))
		if err != nil {
			t.Fatalf("MoveCast failed: %v", err)
		}

		// Verify dirty region was updated
		dirtyRegion := wls.GetDirtyRegion()
		if dirtyRegion.Empty() {
			t.Error("expected dirty region to be set after MoveCast")
		}

		// Dirty region should include both old and new positions
		// Old position: (10, 20) with size (32, 32) -> (10, 20, 42, 52)
		// New position: (100, 200) with size (32, 32) -> (100, 200, 132, 232)
		// Union should be approximately (10, 20, 132, 232)
		if dirtyRegion.Min.X > 10 || dirtyRegion.Min.Y > 20 {
			t.Errorf("dirty region should include old position, got min (%d, %d)", dirtyRegion.Min.X, dirtyRegion.Min.Y)
		}
		if dirtyRegion.Max.X < 132 || dirtyRegion.Max.Y < 232 {
			t.Errorf("dirty region should include new position, got max (%d, %d)", dirtyRegion.Max.X, dirtyRegion.Max.Y)
		}
	})

	t.Run("MoveCast updates source rect in WindowLayerSet", func(t *testing.T) {
		// 要件 4.2: Cast_Layerの位置を更新する
		cm := NewCastManager()
		lm := NewLayerManager()
		cm.SetLayerManager(lm)

		// WindowLayerSetを事前に作成
		winID := 0
		wls := lm.GetOrCreateWindowLayerSet(winID, 640, 480, nil)

		// Create a cast
		castID, _ := cm.PutCast(winID, 1, 10, 20, 0, 0, 32, 32)

		// Move the cast source
		err := cm.MoveCast(castID, WithCastSource(10, 10, 64, 64))
		if err != nil {
			t.Fatalf("MoveCast failed: %v", err)
		}

		// Verify source rect was updated
		castLayer := wls.GetCastLayer(castID)
		srcX, srcY, w, h := castLayer.GetSourceRect()
		if srcX != 10 || srcY != 10 || w != 64 || h != 64 {
			t.Errorf("expected source rect (10, 10, 64, 64), got (%d, %d, %d, %d)", srcX, srcY, w, h)
		}
	})

	t.Run("MoveCast falls back to PictureLayerSet when WindowLayerSet not exists", func(t *testing.T) {
		// 後方互換性: WindowLayerSetが存在しない場合はPictureLayerSetを使用
		cm := NewCastManager()
		lm := NewLayerManager()
		cm.SetLayerManager(lm)

		// WindowLayerSetを作成しない
		winID := 0

		// Create a cast (will use PictureLayerSet)
		castID, _ := cm.PutCast(winID, 1, 10, 20, 0, 0, 32, 32)

		// Move the cast
		err := cm.MoveCast(castID, WithCastPosition(100, 200))
		if err != nil {
			t.Fatalf("MoveCast failed: %v", err)
		}

		// Verify position was updated in PictureLayerSet
		pls := lm.GetPictureLayerSet(winID)
		castLayer := pls.GetCastLayer(castID)
		if castLayer == nil {
			t.Fatal("expected CastLayer to exist in PictureLayerSet")
		}

		x, y := castLayer.GetPosition()
		if x != 100 || y != 200 {
			t.Errorf("expected position (100, 200), got (%d, %d)", x, y)
		}
	})

	t.Run("MoveCast with multiple casts in WindowLayerSet", func(t *testing.T) {
		// 複数のキャストがある場合のテスト
		cm := NewCastManager()
		lm := NewLayerManager()
		cm.SetLayerManager(lm)

		// WindowLayerSetを事前に作成
		winID := 0
		wls := lm.GetOrCreateWindowLayerSet(winID, 640, 480, nil)

		// Create multiple casts
		castID1, _ := cm.PutCast(winID, 1, 10, 10, 0, 0, 32, 32)
		castID2, _ := cm.PutCast(winID, 2, 50, 50, 0, 0, 32, 32)
		castID3, _ := cm.PutCast(winID, 3, 100, 100, 0, 0, 32, 32)

		// Move only the second cast
		err := cm.MoveCast(castID2, WithCastPosition(200, 200))
		if err != nil {
			t.Fatalf("MoveCast failed: %v", err)
		}

		// Verify only the second cast was moved
		layer1 := wls.GetCastLayer(castID1)
		layer2 := wls.GetCastLayer(castID2)
		layer3 := wls.GetCastLayer(castID3)

		x1, y1 := layer1.GetPosition()
		if x1 != 10 || y1 != 10 {
			t.Errorf("cast1 should not have moved, got (%d, %d)", x1, y1)
		}

		x2, y2 := layer2.GetPosition()
		if x2 != 200 || y2 != 200 {
			t.Errorf("cast2 should have moved to (200, 200), got (%d, %d)", x2, y2)
		}

		x3, y3 := layer3.GetPosition()
		if x3 != 100 || y3 != 100 {
			t.Errorf("cast3 should not have moved, got (%d, %d)", x3, y3)
		}
	})
}
