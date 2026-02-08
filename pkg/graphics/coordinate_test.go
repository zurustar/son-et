package graphics

import (
	"testing"
)

// TestCoordinateConverterGetContentOffset はGetContentOffsetのテスト
func TestCoordinateConverterGetContentOffset(t *testing.T) {
	cc := NewCoordinateConverter()

	offsetX, offsetY := cc.GetContentOffset()

	// borderThickness = 4, titleBarHeight = 20
	if offsetX != BorderThickness {
		t.Errorf("Expected offsetX %d, got %d", BorderThickness, offsetX)
	}
	if offsetY != BorderThickness+TitleBarHeight {
		t.Errorf("Expected offsetY %d, got %d", BorderThickness+TitleBarHeight, offsetY)
	}
}

// TestCoordinateConverterWindowToContent はWindowToContentのテスト
func TestCoordinateConverterWindowToContent(t *testing.T) {
	cc := NewCoordinateConverter()

	tests := []struct {
		name      string
		winX      int
		winY      int
		expectedX int
		expectedY int
	}{
		{
			name:      "origin",
			winX:      0,
			winY:      0,
			expectedX: BorderThickness,
			expectedY: BorderThickness + TitleBarHeight,
		},
		{
			name:      "positive position",
			winX:      100,
			winY:      50,
			expectedX: 100 + BorderThickness,
			expectedY: 50 + BorderThickness + TitleBarHeight,
		},
		{
			name:      "negative position",
			winX:      -10,
			winY:      -5,
			expectedX: -10 + BorderThickness,
			expectedY: -5 + BorderThickness + TitleBarHeight,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contentX, contentY := cc.WindowToContent(tt.winX, tt.winY)
			if contentX != tt.expectedX {
				t.Errorf("Expected contentX %d, got %d", tt.expectedX, contentX)
			}
			if contentY != tt.expectedY {
				t.Errorf("Expected contentY %d, got %d", tt.expectedY, contentY)
			}
		})
	}
}

// TestCoordinateConverterPictureToScreen はPictureToScreenのテスト
func TestCoordinateConverterPictureToScreen(t *testing.T) {
	cc := NewCoordinateConverter()

	tests := []struct {
		name       string
		picX       int
		picY       int
		contentX   int
		contentY   int
		picOffsetX int
		picOffsetY int
		expectedX  int
		expectedY  int
	}{
		{
			name:       "no offset",
			picX:       10,
			picY:       20,
			contentX:   100,
			contentY:   50,
			picOffsetX: 0,
			picOffsetY: 0,
			expectedX:  110, // 100 + 10 - 0
			expectedY:  70,  // 50 + 20 - 0
		},
		{
			name:       "positive offset (picture shifts left)",
			picX:       10,
			picY:       20,
			contentX:   100,
			contentY:   50,
			picOffsetX: 5,
			picOffsetY: 10,
			expectedX:  105, // 100 + 10 - 5
			expectedY:  60,  // 50 + 20 - 10
		},
		{
			name:       "negative offset (picture shifts right)",
			picX:       10,
			picY:       20,
			contentX:   100,
			contentY:   50,
			picOffsetX: -5,
			picOffsetY: -10,
			expectedX:  115, // 100 + 10 - (-5)
			expectedY:  80,  // 50 + 20 - (-10)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			screenX, screenY := cc.PictureToScreen(tt.picX, tt.picY, tt.contentX, tt.contentY, tt.picOffsetX, tt.picOffsetY)
			if screenX != tt.expectedX {
				t.Errorf("Expected screenX %d, got %d", tt.expectedX, screenX)
			}
			if screenY != tt.expectedY {
				t.Errorf("Expected screenY %d, got %d", tt.expectedY, screenY)
			}
		})
	}
}

// TestCoordinateConverterScreenToPicture はScreenToPictureのテスト（逆変換）
func TestCoordinateConverterScreenToPicture(t *testing.T) {
	cc := NewCoordinateConverter()

	tests := []struct {
		name       string
		screenX    int
		screenY    int
		contentX   int
		contentY   int
		picOffsetX int
		picOffsetY int
		expectedX  int
		expectedY  int
	}{
		{
			name:       "no offset",
			screenX:    110,
			screenY:    70,
			contentX:   100,
			contentY:   50,
			picOffsetX: 0,
			picOffsetY: 0,
			expectedX:  10, // 110 - 100 + 0
			expectedY:  20, // 70 - 50 + 0
		},
		{
			name:       "positive offset",
			screenX:    105,
			screenY:    60,
			contentX:   100,
			contentY:   50,
			picOffsetX: 5,
			picOffsetY: 10,
			expectedX:  10, // 105 - 100 + 5
			expectedY:  20, // 60 - 50 + 10
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			picX, picY := cc.ScreenToPicture(tt.screenX, tt.screenY, tt.contentX, tt.contentY, tt.picOffsetX, tt.picOffsetY)
			if picX != tt.expectedX {
				t.Errorf("Expected picX %d, got %d", tt.expectedX, picX)
			}
			if picY != tt.expectedY {
				t.Errorf("Expected picY %d, got %d", tt.expectedY, picY)
			}
		})
	}
}

// TestCoordinateConverterRoundTrip は座標変換の往復テスト
// PictureToScreen -> ScreenToPicture で元の座標に戻ることを確認
func TestCoordinateConverterRoundTrip(t *testing.T) {
	cc := NewCoordinateConverter()

	tests := []struct {
		name       string
		picX       int
		picY       int
		contentX   int
		contentY   int
		picOffsetX int
		picOffsetY int
	}{
		{
			name:       "no offset",
			picX:       10,
			picY:       20,
			contentX:   100,
			contentY:   50,
			picOffsetX: 0,
			picOffsetY: 0,
		},
		{
			name:       "positive offset",
			picX:       10,
			picY:       20,
			contentX:   100,
			contentY:   50,
			picOffsetX: 5,
			picOffsetY: 10,
		},
		{
			name:       "negative offset",
			picX:       10,
			picY:       20,
			contentX:   100,
			contentY:   50,
			picOffsetX: -5,
			picOffsetY: -10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Picture -> Screen
			screenX, screenY := cc.PictureToScreen(tt.picX, tt.picY, tt.contentX, tt.contentY, tt.picOffsetX, tt.picOffsetY)

			// Screen -> Picture
			resultX, resultY := cc.ScreenToPicture(screenX, screenY, tt.contentX, tt.contentY, tt.picOffsetX, tt.picOffsetY)

			if resultX != tt.picX {
				t.Errorf("Round trip failed for X: expected %d, got %d", tt.picX, resultX)
			}
			if resultY != tt.picY {
				t.Errorf("Round trip failed for Y: expected %d, got %d", tt.picY, resultY)
			}
		})
	}
}

// TestGetPicOffset はGetPicOffsetのテスト
func TestGetPicOffset(t *testing.T) {
	tests := []struct {
		name      string
		win       *Window
		expectedX int
		expectedY int
	}{
		{
			name:      "nil window",
			win:       nil,
			expectedX: 0,
			expectedY: 0,
		},
		{
			name: "zero offset",
			win: &Window{
				PicX: 0,
				PicY: 0,
			},
			expectedX: 0,
			expectedY: 0,
		},
		{
			name: "positive offset",
			win: &Window{
				PicX: 10,
				PicY: 20,
			},
			expectedX: 10,
			expectedY: 20,
		},
		{
			name: "negative offset",
			win: &Window{
				PicX: -100,
				PicY: -75,
			},
			expectedX: -100,
			expectedY: -75,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			offsetX, offsetY := GetPicOffset(tt.win)
			if offsetX != tt.expectedX {
				t.Errorf("Expected offsetX %d, got %d", tt.expectedX, offsetX)
			}
			if offsetY != tt.expectedY {
				t.Errorf("Expected offsetY %d, got %d", tt.expectedY, offsetY)
			}
		})
	}
}

// TestCalculateDrawPosition はCalculateDrawPositionのテスト
func TestCalculateDrawPosition(t *testing.T) {
	cc := NewCoordinateConverter()

	tests := []struct {
		name      string
		win       *Window
		spriteX   float64
		spriteY   float64
		expectedX float64
		expectedY float64
	}{
		{
			name: "no offset",
			win: &Window{
				X:    100,
				Y:    50,
				PicX: 0,
				PicY: 0,
			},
			spriteX:   10,
			spriteY:   20,
			expectedX: float64(100 + BorderThickness + 10),
			expectedY: float64(50 + BorderThickness + TitleBarHeight + 20),
		},
		{
			name: "positive offset",
			win: &Window{
				X:    100,
				Y:    50,
				PicX: 5,
				PicY: 10,
			},
			spriteX:   10,
			spriteY:   20,
			expectedX: float64(100 + BorderThickness + 10 - 5),
			expectedY: float64(50 + BorderThickness + TitleBarHeight + 20 - 10),
		},
		{
			name: "negative offset",
			win: &Window{
				X:    100,
				Y:    50,
				PicX: -100,
				PicY: -75,
			},
			spriteX:   10,
			spriteY:   20,
			expectedX: float64(100 + BorderThickness + 10 + 100),
			expectedY: float64(50 + BorderThickness + TitleBarHeight + 20 + 75),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			screenX, screenY := cc.CalculateDrawPosition(tt.win, tt.spriteX, tt.spriteY)
			if screenX != tt.expectedX {
				t.Errorf("Expected screenX %f, got %f", tt.expectedX, screenX)
			}
			if screenY != tt.expectedY {
				t.Errorf("Expected screenY %f, got %f", tt.expectedY, screenY)
			}
		})
	}
}

// TestDefaultCoordinateConverter はデフォルトのCoordinateConverterのテスト
func TestDefaultCoordinateConverter(t *testing.T) {
	cc := GetDefaultCoordinateConverter()
	if cc == nil {
		t.Error("GetDefaultCoordinateConverter returned nil")
	}

	// デフォルトの定数が正しく設定されていることを確認
	offsetX, offsetY := cc.GetContentOffset()
	if offsetX != BorderThickness {
		t.Errorf("Expected offsetX %d, got %d", BorderThickness, offsetX)
	}
	if offsetY != BorderThickness+TitleBarHeight {
		t.Errorf("Expected offsetY %d, got %d", BorderThickness+TitleBarHeight, offsetY)
	}
}
