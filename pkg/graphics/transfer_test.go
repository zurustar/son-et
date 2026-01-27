package graphics

import (
	"image/color"
	"testing"
)

// TestMovePicNormalMode tests MovePic with mode=0 (normal copy)
func TestMovePicNormalMode(t *testing.T) {
	gs := NewGraphicsSystem("")

	// Create source and destination pictures
	srcID, err := gs.CreatePic(100, 100)
	if err != nil {
		t.Fatalf("Failed to create source picture: %v", err)
	}

	dstID, err := gs.CreatePic(100, 100)
	if err != nil {
		t.Fatalf("Failed to create destination picture: %v", err)
	}

	// Transfer from source to destination
	err = gs.MovePic(srcID, 0, 0, 50, 50, dstID, 10, 10, 0)
	if err != nil {
		t.Errorf("MovePic failed: %v", err)
	}
}

// TestMovePicTransparentMode tests MovePic with mode=1 (transparent color exclusion)
func TestMovePicTransparentMode(t *testing.T) {
	gs := NewGraphicsSystem("")

	srcID, err := gs.CreatePic(100, 100)
	if err != nil {
		t.Fatalf("Failed to create source picture: %v", err)
	}

	dstID, err := gs.CreatePic(100, 100)
	if err != nil {
		t.Fatalf("Failed to create destination picture: %v", err)
	}

	// Transfer with transparency
	err = gs.MovePic(srcID, 0, 0, 50, 50, dstID, 10, 10, 1)
	if err != nil {
		t.Errorf("MovePic with transparency failed: %v", err)
	}
}

// TestMovePicInvalidSource tests MovePic with invalid source picture
func TestMovePicInvalidSource(t *testing.T) {
	gs := NewGraphicsSystem("")

	dstID, err := gs.CreatePic(100, 100)
	if err != nil {
		t.Fatalf("Failed to create destination picture: %v", err)
	}

	// Try to transfer from non-existent source
	err = gs.MovePic(999, 0, 0, 50, 50, dstID, 10, 10, 0)
	if err == nil {
		t.Error("Expected error for invalid source picture, got nil")
	}
}

// TestMovePicInvalidDestination tests MovePic with invalid destination picture
func TestMovePicInvalidDestination(t *testing.T) {
	gs := NewGraphicsSystem("")

	srcID, err := gs.CreatePic(100, 100)
	if err != nil {
		t.Fatalf("Failed to create source picture: %v", err)
	}

	// Try to transfer to non-existent destination
	err = gs.MovePic(srcID, 0, 0, 50, 50, 999, 10, 10, 0)
	if err == nil {
		t.Error("Expected error for invalid destination picture, got nil")
	}
}

// TestMovePicSamePicture tests MovePic when source and destination are the same
func TestMovePicSamePicture(t *testing.T) {
	gs := NewGraphicsSystem("")

	picID, err := gs.CreatePic(100, 100)
	if err != nil {
		t.Fatalf("Failed to create picture: %v", err)
	}

	// Try to transfer to same picture (should not error, just skip)
	err = gs.MovePic(picID, 0, 0, 50, 50, picID, 10, 10, 0)
	if err != nil {
		t.Errorf("MovePic to same picture should not error: %v", err)
	}
}

// TestTransPic tests TransPic with custom transparent color
func TestTransPic(t *testing.T) {
	gs := NewGraphicsSystem("")

	srcID, err := gs.CreatePic(100, 100)
	if err != nil {
		t.Fatalf("Failed to create source picture: %v", err)
	}

	dstID, err := gs.CreatePic(100, 100)
	if err != nil {
		t.Fatalf("Failed to create destination picture: %v", err)
	}

	// Transfer with custom transparent color (red)
	transColor := color.RGBA{255, 0, 0, 255}
	err = gs.TransPic(srcID, 0, 0, 50, 50, dstID, 10, 10, transColor)
	if err != nil {
		t.Errorf("TransPic failed: %v", err)
	}
}

// TestTransPicInvalidPictures tests TransPic with invalid pictures
func TestTransPicInvalidPictures(t *testing.T) {
	gs := NewGraphicsSystem("")

	dstID, err := gs.CreatePic(100, 100)
	if err != nil {
		t.Fatalf("Failed to create destination picture: %v", err)
	}

	// Invalid source
	err = gs.TransPic(999, 0, 0, 50, 50, dstID, 10, 10, color.Black)
	if err == nil {
		t.Error("Expected error for invalid source picture")
	}

	srcID, err := gs.CreatePic(100, 100)
	if err != nil {
		t.Fatalf("Failed to create source picture: %v", err)
	}

	// Invalid destination
	err = gs.TransPic(srcID, 0, 0, 50, 50, 999, 10, 10, color.Black)
	if err == nil {
		t.Error("Expected error for invalid destination picture")
	}
}

// TestReversePic tests ReversePic (horizontal flip)
func TestReversePic(t *testing.T) {
	gs := NewGraphicsSystem("")

	srcID, err := gs.CreatePic(100, 100)
	if err != nil {
		t.Fatalf("Failed to create source picture: %v", err)
	}

	dstID, err := gs.CreatePic(100, 100)
	if err != nil {
		t.Fatalf("Failed to create destination picture: %v", err)
	}

	// Flip and transfer
	err = gs.ReversePic(srcID, 0, 0, 50, 50, dstID, 10, 10)
	if err != nil {
		t.Errorf("ReversePic failed: %v", err)
	}
}

// TestReversePicInvalidPictures tests ReversePic with invalid pictures
func TestReversePicInvalidPictures(t *testing.T) {
	gs := NewGraphicsSystem("")

	dstID, err := gs.CreatePic(100, 100)
	if err != nil {
		t.Fatalf("Failed to create destination picture: %v", err)
	}

	// Invalid source
	err = gs.ReversePic(999, 0, 0, 50, 50, dstID, 10, 10)
	if err == nil {
		t.Error("Expected error for invalid source picture")
	}

	srcID, err := gs.CreatePic(100, 100)
	if err != nil {
		t.Fatalf("Failed to create source picture: %v", err)
	}

	// Invalid destination
	err = gs.ReversePic(srcID, 0, 0, 50, 50, 999, 10, 10)
	if err == nil {
		t.Error("Expected error for invalid destination picture")
	}
}

// TestMoveSPic tests MoveSPic (scaled transfer)
func TestMoveSPic(t *testing.T) {
	gs := NewGraphicsSystem("")

	srcID, err := gs.CreatePic(100, 100)
	if err != nil {
		t.Fatalf("Failed to create source picture: %v", err)
	}

	dstID, err := gs.CreatePic(200, 200)
	if err != nil {
		t.Fatalf("Failed to create destination picture: %v", err)
	}

	// Scale up 2x
	err = gs.MoveSPic(srcID, 0, 0, 50, 50, dstID, 0, 0, 100, 100)
	if err != nil {
		t.Errorf("MoveSPic (scale up) failed: %v", err)
	}

	// Scale down 0.5x
	err = gs.MoveSPic(srcID, 0, 0, 100, 100, dstID, 0, 0, 50, 50)
	if err != nil {
		t.Errorf("MoveSPic (scale down) failed: %v", err)
	}
}

// TestMoveSPicInvalidPictures tests MoveSPic with invalid pictures
func TestMoveSPicInvalidPictures(t *testing.T) {
	gs := NewGraphicsSystem("")

	dstID, err := gs.CreatePic(100, 100)
	if err != nil {
		t.Fatalf("Failed to create destination picture: %v", err)
	}

	// Invalid source
	err = gs.MoveSPic(999, 0, 0, 50, 50, dstID, 0, 0, 100, 100)
	if err == nil {
		t.Error("Expected error for invalid source picture")
	}

	srcID, err := gs.CreatePic(100, 100)
	if err != nil {
		t.Fatalf("Failed to create source picture: %v", err)
	}

	// Invalid destination
	err = gs.MoveSPic(srcID, 0, 0, 50, 50, 999, 0, 0, 100, 100)
	if err == nil {
		t.Error("Expected error for invalid destination picture")
	}
}

// TestClipTransferRegion tests the clipping function
func TestClipTransferRegion(t *testing.T) {
	tests := []struct {
		name                                           string
		srcX, srcY, width, height, srcWidth, srcHeight int
		dstX, dstY, dstWidth, dstHeight                int
		wantSrcX, wantSrcY, wantWidth, wantHeight      int
		wantDstX, wantDstY                             int
	}{
		{
			name: "no clipping needed",
			srcX: 0, srcY: 0, width: 50, height: 50,
			srcWidth: 100, srcHeight: 100,
			dstX: 0, dstY: 0, dstWidth: 100, dstHeight: 100,
			wantSrcX: 0, wantSrcY: 0, wantWidth: 50, wantHeight: 50,
			wantDstX: 0, wantDstY: 0,
		},
		{
			name: "source exceeds bounds",
			srcX: 80, srcY: 80, width: 50, height: 50,
			srcWidth: 100, srcHeight: 100,
			dstX: 0, dstY: 0, dstWidth: 100, dstHeight: 100,
			wantSrcX: 80, wantSrcY: 80, wantWidth: 20, wantHeight: 20,
			wantDstX: 0, wantDstY: 0,
		},
		{
			name: "destination exceeds bounds",
			srcX: 0, srcY: 0, width: 50, height: 50,
			srcWidth: 100, srcHeight: 100,
			dstX: 80, dstY: 80, dstWidth: 100, dstHeight: 100,
			wantSrcX: 0, wantSrcY: 0, wantWidth: 20, wantHeight: 20,
			wantDstX: 80, wantDstY: 80,
		},
		{
			name: "negative source coordinates",
			srcX: -10, srcY: -10, width: 50, height: 50,
			srcWidth: 100, srcHeight: 100,
			dstX: 0, dstY: 0, dstWidth: 100, dstHeight: 100,
			wantSrcX: 0, wantSrcY: 0, wantWidth: 40, wantHeight: 40,
			wantDstX: 10, wantDstY: 10,
		},
		{
			name: "negative destination coordinates",
			srcX: 0, srcY: 0, width: 50, height: 50,
			srcWidth: 100, srcHeight: 100,
			dstX: -10, dstY: -10, dstWidth: 100, dstHeight: 100,
			wantSrcX: 10, wantSrcY: 10, wantWidth: 40, wantHeight: 40,
			wantDstX: 0, wantDstY: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSrcX, gotSrcY, gotWidth, gotHeight, gotDstX, gotDstY := clipTransferRegion(
				tt.srcX, tt.srcY, tt.width, tt.height,
				tt.srcWidth, tt.srcHeight,
				tt.dstX, tt.dstY,
				tt.dstWidth, tt.dstHeight,
			)

			if gotSrcX != tt.wantSrcX || gotSrcY != tt.wantSrcY ||
				gotWidth != tt.wantWidth || gotHeight != tt.wantHeight ||
				gotDstX != tt.wantDstX || gotDstY != tt.wantDstY {
				t.Errorf("clipTransferRegion() = (%d, %d, %d, %d, %d, %d), want (%d, %d, %d, %d, %d, %d)",
					gotSrcX, gotSrcY, gotWidth, gotHeight, gotDstX, gotDstY,
					tt.wantSrcX, tt.wantSrcY, tt.wantWidth, tt.wantHeight, tt.wantDstX, tt.wantDstY)
			}
		})
	}
}

// TestMovePicClipping tests that MovePic properly clips regions
func TestMovePicClipping(t *testing.T) {
	gs := NewGraphicsSystem("")

	srcID, err := gs.CreatePic(100, 100)
	if err != nil {
		t.Fatalf("Failed to create source picture: %v", err)
	}

	dstID, err := gs.CreatePic(100, 100)
	if err != nil {
		t.Fatalf("Failed to create destination picture: %v", err)
	}

	// Transfer region that exceeds source bounds
	err = gs.MovePic(srcID, 80, 80, 50, 50, dstID, 0, 0, 0)
	if err != nil {
		t.Errorf("MovePic with clipping failed: %v", err)
	}

	// Transfer region that exceeds destination bounds
	err = gs.MovePic(srcID, 0, 0, 50, 50, dstID, 80, 80, 0)
	if err != nil {
		t.Errorf("MovePic with destination clipping failed: %v", err)
	}

	// Transfer with negative source coordinates
	err = gs.MovePic(srcID, -10, -10, 50, 50, dstID, 0, 0, 0)
	if err != nil {
		t.Errorf("MovePic with negative source coords failed: %v", err)
	}

	// Transfer with negative destination coordinates
	err = gs.MovePic(srcID, 0, 0, 50, 50, dstID, -10, -10, 0)
	if err != nil {
		t.Errorf("MovePic with negative destination coords failed: %v", err)
	}
}

// TestMoveSPicZeroSize tests MoveSPic with zero or negative sizes
func TestMoveSPicZeroSize(t *testing.T) {
	gs := NewGraphicsSystem("")

	srcID, err := gs.CreatePic(100, 100)
	if err != nil {
		t.Fatalf("Failed to create source picture: %v", err)
	}

	dstID, err := gs.CreatePic(100, 100)
	if err != nil {
		t.Fatalf("Failed to create destination picture: %v", err)
	}

	// Zero source size should not error (just skip)
	err = gs.MoveSPic(srcID, 0, 0, 0, 0, dstID, 0, 0, 50, 50)
	if err != nil {
		t.Errorf("MoveSPic with zero source size should not error: %v", err)
	}

	// Zero destination size should not error (just skip)
	err = gs.MoveSPic(srcID, 0, 0, 50, 50, dstID, 0, 0, 0, 0)
	if err != nil {
		t.Errorf("MoveSPic with zero destination size should not error: %v", err)
	}
}
