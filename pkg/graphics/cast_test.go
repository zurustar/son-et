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

// TestCastManagerWithoutLayerManager tests that CastManager works independently without LayerManager
// スプライトシステム移行: LayerManagerは削除されました
func TestCastManagerWithoutLayerManager(t *testing.T) {
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

	// Verify position was updated
	cast, _ := cm.GetCast(castID)
	if cast.X != 100 || cast.Y != 200 {
		t.Errorf("expected position (100, 200), got (%d, %d)", cast.X, cast.Y)
	}

	// Delete the cast
	err = cm.DelCast(castID)
	if err != nil {
		t.Fatalf("DelCast failed: %v", err)
	}
}
