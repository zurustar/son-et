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
