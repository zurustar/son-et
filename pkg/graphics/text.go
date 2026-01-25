package graphics

import (
	"image/color"
	"sync"
)

// FontSettings はフォント設定を保持する
type FontSettings struct {
	Name      string
	Size      int
	Weight    int
	Italic    bool
	Underline bool
	Strikeout bool
}

// TextSettings はテキスト描画設定を保持する
type TextSettings struct {
	TextColor color.Color
	BgColor   color.Color
	BackMode  int
}

// TextRenderer はテキスト描画を管理する
type TextRenderer struct {
	font     *FontSettings
	settings *TextSettings
	mu       sync.RWMutex
}

// NewTextRenderer は新しい TextRenderer を作成する
func NewTextRenderer() *TextRenderer {
	return &TextRenderer{
		font: &FontSettings{
			Name:   "default",
			Size:   12,
			Weight: 400,
		},
		settings: &TextSettings{
			TextColor: color.RGBA{255, 255, 255, 255},
			BgColor:   color.RGBA{0, 0, 0, 255},
			BackMode:  0,
		},
	}
}
