package graphics

import (
	"image/color"
	"testing"
)

// TestNewPictureLayerError はNewPictureLayerのエラー処理をテストする
// 要件 10.3: レイヤー作成に失敗したときにエラーをログに記録し、nilを返す
func TestNewPictureLayerError(t *testing.T) {
	tests := []struct {
		name      string
		id        int
		winWidth  int
		winHeight int
		wantNil   bool
	}{
		{
			name:      "valid parameters",
			id:        1,
			winWidth:  640,
			winHeight: 480,
			wantNil:   false,
		},
		{
			name:      "zero width",
			id:        2,
			winWidth:  0,
			winHeight: 480,
			wantNil:   true,
		},
		{
			name:      "zero height",
			id:        3,
			winWidth:  640,
			winHeight: 0,
			wantNil:   true,
		},
		{
			name:      "negative width",
			id:        4,
			winWidth:  -100,
			winHeight: 480,
			wantNil:   true,
		},
		{
			name:      "negative height",
			id:        5,
			winWidth:  640,
			winHeight: -100,
			wantNil:   true,
		},
		{
			name:      "both zero",
			id:        6,
			winWidth:  0,
			winHeight: 0,
			wantNil:   true,
		},
		{
			name:      "both negative",
			id:        7,
			winWidth:  -100,
			winHeight: -200,
			wantNil:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			layer := NewPictureLayer(tt.id, tt.winWidth, tt.winHeight)
			if tt.wantNil && layer != nil {
				t.Errorf("NewPictureLayer() = %v, want nil", layer)
			}
			if !tt.wantNil && layer == nil {
				t.Errorf("NewPictureLayer() = nil, want non-nil")
			}
		})
	}
}

// TestNewCastLayerError はNewCastLayerのエラー処理をテストする
// 要件 10.3: レイヤー作成に失敗したときにエラーをログに記録し、nilを返す
func TestNewCastLayerError(t *testing.T) {
	tests := []struct {
		name    string
		id      int
		castID  int
		width   int
		height  int
		wantNil bool
	}{
		{
			name:    "valid parameters",
			id:      1,
			castID:  1,
			width:   100,
			height:  100,
			wantNil: false,
		},
		{
			name:    "zero width",
			id:      2,
			castID:  2,
			width:   0,
			height:  100,
			wantNil: true,
		},
		{
			name:    "zero height",
			id:      3,
			castID:  3,
			width:   100,
			height:  0,
			wantNil: true,
		},
		{
			name:    "negative width",
			id:      4,
			castID:  4,
			width:   -50,
			height:  100,
			wantNil: true,
		},
		{
			name:    "negative height",
			id:      5,
			castID:  5,
			width:   100,
			height:  -50,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			layer := NewCastLayer(tt.id, tt.castID, 0, 0, 0, 0, 0, 0, tt.width, tt.height, 0)
			if tt.wantNil && layer != nil {
				t.Errorf("NewCastLayer() = %v, want nil", layer)
			}
			if !tt.wantNil && layer == nil {
				t.Errorf("NewCastLayer() = nil, want non-nil")
			}
		})
	}
}

// TestNewCastLayerWithTransColorError はNewCastLayerWithTransColorのエラー処理をテストする
// 要件 10.3: レイヤー作成に失敗したときにエラーをログに記録し、nilを返す
func TestNewCastLayerWithTransColorError(t *testing.T) {
	tests := []struct {
		name       string
		id         int
		castID     int
		width      int
		height     int
		transColor color.Color
		wantNil    bool
	}{
		{
			name:       "valid parameters with trans color",
			id:         1,
			castID:     1,
			width:      100,
			height:     100,
			transColor: color.RGBA{R: 255, G: 0, B: 255, A: 255},
			wantNil:    false,
		},
		{
			name:       "valid parameters without trans color",
			id:         2,
			castID:     2,
			width:      100,
			height:     100,
			transColor: nil,
			wantNil:    false,
		},
		{
			name:       "invalid size with trans color",
			id:         3,
			castID:     3,
			width:      0,
			height:     100,
			transColor: color.RGBA{R: 255, G: 0, B: 255, A: 255},
			wantNil:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			layer := NewCastLayerWithTransColor(tt.id, tt.castID, 0, 0, 0, 0, 0, 0, tt.width, tt.height, 0, tt.transColor)
			if tt.wantNil && layer != nil {
				t.Errorf("NewCastLayerWithTransColor() = %v, want nil", layer)
			}
			if !tt.wantNil && layer == nil {
				t.Errorf("NewCastLayerWithTransColor() = nil, want non-nil")
			}
		})
	}
}

// TestNewDrawingLayerError はNewDrawingLayerのエラー処理をテストする
// 要件 10.3: レイヤー作成に失敗したときにエラーをログに記録し、nilを返す
func TestNewDrawingLayerError(t *testing.T) {
	tests := []struct {
		name    string
		id      int
		picID   int
		width   int
		height  int
		wantNil bool
	}{
		{
			name:    "valid parameters",
			id:      1,
			picID:   1,
			width:   640,
			height:  480,
			wantNil: false,
		},
		{
			name:    "zero width",
			id:      2,
			picID:   2,
			width:   0,
			height:  480,
			wantNil: true,
		},
		{
			name:    "zero height",
			id:      3,
			picID:   3,
			width:   640,
			height:  0,
			wantNil: true,
		},
		{
			name:    "negative width",
			id:      4,
			picID:   4,
			width:   -100,
			height:  480,
			wantNil: true,
		},
		{
			name:    "negative height",
			id:      5,
			picID:   5,
			width:   640,
			height:  -100,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			layer := NewDrawingLayer(tt.id, tt.picID, tt.width, tt.height)
			if tt.wantNil && layer != nil {
				t.Errorf("NewDrawingLayer() = %v, want nil", layer)
			}
			if !tt.wantNil && layer == nil {
				t.Errorf("NewDrawingLayer() = nil, want non-nil")
			}
		})
	}
}

// TestNewDrawingEntryError はNewDrawingEntryのエラー処理をテストする
// 要件 10.3: レイヤー作成に失敗したときにエラーをログに記録し、nilを返す
func TestNewDrawingEntryError(t *testing.T) {
	tests := []struct {
		name    string
		id      int
		picID   int
		width   int
		height  int
		wantNil bool
	}{
		{
			name:    "valid parameters",
			id:      1,
			picID:   1,
			width:   100,
			height:  100,
			wantNil: false,
		},
		{
			name:    "zero width",
			id:      2,
			picID:   2,
			width:   0,
			height:  100,
			wantNil: true,
		},
		{
			name:    "zero height",
			id:      3,
			picID:   3,
			width:   100,
			height:  0,
			wantNil: true,
		},
		{
			name:    "negative width",
			id:      4,
			picID:   4,
			width:   -50,
			height:  100,
			wantNil: true,
		},
		{
			name:    "negative height",
			id:      5,
			picID:   5,
			width:   100,
			height:  -50,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := NewDrawingEntry(tt.id, tt.picID, nil, 0, 0, tt.width, tt.height, 0)
			if tt.wantNil && entry != nil {
				t.Errorf("NewDrawingEntry() = %v, want nil", entry)
			}
			if !tt.wantNil && entry == nil {
				t.Errorf("NewDrawingEntry() = nil, want non-nil")
			}
		})
	}
}

// TestNewWindowLayerSetError はNewWindowLayerSetのエラー処理をテストする
// 要件 10.3: レイヤー作成に失敗したときにエラーをログに記録し、nilを返す
func TestNewWindowLayerSetError(t *testing.T) {
	tests := []struct {
		name    string
		winID   int
		width   int
		height  int
		wantNil bool
	}{
		{
			name:    "valid parameters",
			winID:   1,
			width:   640,
			height:  480,
			wantNil: false,
		},
		{
			name:    "zero width",
			winID:   2,
			width:   0,
			height:  480,
			wantNil: true,
		},
		{
			name:    "zero height",
			winID:   3,
			width:   640,
			height:  0,
			wantNil: true,
		},
		{
			name:    "negative width",
			winID:   4,
			width:   -100,
			height:  480,
			wantNil: true,
		},
		{
			name:    "negative height",
			winID:   5,
			width:   640,
			height:  -100,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wls := NewWindowLayerSet(tt.winID, tt.width, tt.height, color.White)
			if tt.wantNil && wls != nil {
				t.Errorf("NewWindowLayerSet() = %v, want nil", wls)
			}
			if !tt.wantNil && wls == nil {
				t.Errorf("NewWindowLayerSet() = nil, want non-nil")
			}
		})
	}
}

// TestNewTextLayerEntryFromTextLayerError はNewTextLayerEntryFromTextLayerのエラー処理をテストする
// 要件 10.3: レイヤー作成に失敗したときにエラーをログに記録し、nilを返す
func TestNewTextLayerEntryFromTextLayerError(t *testing.T) {
	t.Run("nil textLayer", func(t *testing.T) {
		entry := NewTextLayerEntryFromTextLayer(1, nil, 0)
		if entry != nil {
			t.Errorf("NewTextLayerEntryFromTextLayer(nil) = %v, want nil", entry)
		}
	})

	t.Run("valid textLayer", func(t *testing.T) {
		textLayer := &TextLayer{
			PicID: 1,
			X:     10,
			Y:     20,
			Image: nil,
		}
		entry := NewTextLayerEntryFromTextLayer(1, textLayer, 0)
		if entry == nil {
			t.Errorf("NewTextLayerEntryFromTextLayer(valid) = nil, want non-nil")
		}
	})
}

// TestLayerCreationErrorGracefulHandling はレイヤー作成失敗後も実行が継続されることをテストする
// 要件 10.4: 致命的でないエラーの後も実行を継続する
func TestLayerCreationErrorGracefulHandling(t *testing.T) {
	// 無効なパラメータでレイヤー作成を試みる
	pictureLayer := NewPictureLayer(1, -100, -100)
	castLayer := NewCastLayer(2, 1, 0, 0, 0, 0, 0, 0, -50, -50, 0)
	drawingLayer := NewDrawingLayer(3, 1, -100, -100)
	windowLayerSet := NewWindowLayerSet(1, -100, -100, color.White)

	// すべてnilが返されることを確認
	if pictureLayer != nil {
		t.Errorf("Expected nil PictureLayer for invalid params")
	}
	if castLayer != nil {
		t.Errorf("Expected nil CastLayer for invalid params")
	}
	if drawingLayer != nil {
		t.Errorf("Expected nil DrawingLayer for invalid params")
	}
	if windowLayerSet != nil {
		t.Errorf("Expected nil WindowLayerSet for invalid params")
	}

	// 有効なパラメータでレイヤー作成が成功することを確認
	// （エラー後も実行が継続されることを確認）
	validPictureLayer := NewPictureLayer(10, 640, 480)
	if validPictureLayer == nil {
		t.Errorf("Expected non-nil PictureLayer for valid params after error")
	}

	validCastLayer := NewCastLayer(11, 1, 0, 0, 0, 0, 0, 0, 100, 100, 0)
	if validCastLayer == nil {
		t.Errorf("Expected non-nil CastLayer for valid params after error")
	}

	validDrawingLayer := NewDrawingLayer(12, 1, 640, 480)
	if validDrawingLayer == nil {
		t.Errorf("Expected non-nil DrawingLayer for valid params after error")
	}

	validWindowLayerSet := NewWindowLayerSet(2, 640, 480, color.White)
	if validWindowLayerSet == nil {
		t.Errorf("Expected non-nil WindowLayerSet for valid params after error")
	}
}
