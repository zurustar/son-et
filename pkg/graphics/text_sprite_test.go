package graphics

import (
	"image"
	"image/color"
	"testing"

	"golang.org/x/image/font/basicfont"
)

func TestCreateTextSpriteImage_Basic(t *testing.T) {
	opts := TextSpriteOptions{
		Text:      "Hello",
		TextColor: color.Black,
		Face:      basicfont.Face7x13,
		BgColor:   color.White,
		Width:     100,
		Height:    30,
		X:         5,
		Y:         20,
	}

	img := CreateTextSpriteImage(opts)
	if img == nil {
		t.Fatal("expected non-nil image")
	}

	bounds := img.Bounds()
	if bounds.Dx() != 100 || bounds.Dy() != 30 {
		t.Errorf("expected size (100,30), got (%d,%d)", bounds.Dx(), bounds.Dy())
	}
}

func TestCreateTextSpriteImage_NilFace(t *testing.T) {
	opts := TextSpriteOptions{
		Text:      "Hello",
		TextColor: color.Black,
		Face:      nil,
		BgColor:   color.White,
	}

	img := CreateTextSpriteImage(opts)
	if img != nil {
		t.Error("expected nil image for nil face")
	}
}

func TestCreateTextSpriteImage_EmptyText(t *testing.T) {
	opts := TextSpriteOptions{
		Text:      "",
		TextColor: color.Black,
		Face:      basicfont.Face7x13,
		BgColor:   color.White,
	}

	img := CreateTextSpriteImage(opts)
	if img != nil {
		t.Error("expected nil image for empty text")
	}
}

func TestCreateTextSpriteImage_AutoSize(t *testing.T) {
	opts := TextSpriteOptions{
		Text:      "Test",
		TextColor: color.Black,
		Face:      basicfont.Face7x13,
		BgColor:   color.White,
		X:         0,
		Y:         13,
	}

	img := CreateTextSpriteImage(opts)
	if img == nil {
		t.Fatal("expected non-nil image")
	}

	bounds := img.Bounds()
	if bounds.Dx() <= 0 || bounds.Dy() <= 0 {
		t.Errorf("expected positive size, got (%d,%d)", bounds.Dx(), bounds.Dy())
	}
}

func TestCreateTextSpriteImage_DifferenceExtraction(t *testing.T) {
	bgColor := color.RGBA{255, 255, 200, 255} // 薄い黄色
	textColor := color.RGBA{0, 0, 0, 255}     // 黒

	opts := TextSpriteOptions{
		Text:      "A",
		TextColor: textColor,
		Face:      basicfont.Face7x13,
		BgColor:   bgColor,
		Width:     20,
		Height:    20,
		X:         5,
		Y:         15,
	}

	img := CreateTextSpriteImage(opts)
	if img == nil {
		t.Fatal("expected non-nil image")
	}

	// 背景色のピクセルは透明になっているはず
	hasTransparent := false
	hasOpaque := false

	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			_, _, _, a := img.At(x, y).RGBA()
			if a == 0 {
				hasTransparent = true
			} else {
				hasOpaque = true
			}
		}
	}

	if !hasTransparent {
		t.Error("expected some transparent pixels (background)")
	}
	if !hasOpaque {
		t.Error("expected some opaque pixels (text)")
	}
}

func TestCreateTextSpriteImage_DefaultColors(t *testing.T) {
	opts := TextSpriteOptions{
		Text:   "Test",
		Face:   basicfont.Face7x13,
		Width:  50,
		Height: 20,
		Y:      15,
		// TextColor と BgColor は nil（デフォルト値を使用）
	}

	img := CreateTextSpriteImage(opts)
	if img == nil {
		t.Fatal("expected non-nil image with default colors")
	}
}

func TestMeasureText_Sprite(t *testing.T) {
	face := basicfont.Face7x13
	bounds := measureText(face, "Hello")

	if bounds.Dx() <= 0 {
		t.Errorf("expected positive width, got %d", bounds.Dx())
	}
}

func TestSpriteManager_CreateTextSprite(t *testing.T) {
	sm := NewSpriteManager()

	opts := TextSpriteOptions{
		Text:      "Hello",
		TextColor: color.Black,
		Face:      basicfont.Face7x13,
		BgColor:   color.White,
		Width:     100,
		Height:    30,
		Y:         20,
	}

	sprite := sm.CreateTextSprite(opts)
	if sprite == nil {
		t.Fatal("expected non-nil sprite")
	}

	if sm.Count() != 1 {
		t.Errorf("expected count 1, got %d", sm.Count())
	}

	// IDで取得できることを確認
	got := sm.GetSprite(sprite.ID())
	if got != sprite {
		t.Error("GetSprite returned wrong sprite")
	}
}

func TestSpriteManager_CreateTextSprite_Invalid(t *testing.T) {
	sm := NewSpriteManager()

	// 無効なオプション（空のテキスト）
	opts := TextSpriteOptions{
		Text: "",
		Face: basicfont.Face7x13,
	}

	sprite := sm.CreateTextSprite(opts)
	if sprite != nil {
		t.Error("expected nil sprite for invalid options")
	}

	if sm.Count() != 0 {
		t.Errorf("expected count 0, got %d", sm.Count())
	}
}

// TestExtractDifference は差分抽出関数を直接テストする
func TestExtractDifference(t *testing.T) {
	// テスト用の小さな画像を作成
	bg := createTestImage(10, 10, color.RGBA{255, 255, 255, 255})
	text := createTestImage(10, 10, color.RGBA{255, 255, 255, 255})
	result := createTestImage(10, 10, color.RGBA{0, 0, 0, 0})

	// テキスト画像の一部を変更
	text.Set(5, 5, color.RGBA{0, 0, 0, 255})

	extractDifference(bg, text, result)

	// 変更されたピクセルは残る
	r, g, b, a := result.At(5, 5).RGBA()
	if a == 0 {
		t.Error("expected opaque pixel at changed position")
	}
	if r != 0 || g != 0 || b != 0 {
		t.Error("expected black pixel at changed position")
	}

	// 変更されていないピクセルは透明
	_, _, _, a = result.At(0, 0).RGBA()
	if a != 0 {
		t.Error("expected transparent pixel at unchanged position")
	}
}

func createTestImage(width, height int, c color.Color) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, c)
		}
	}
	return img
}
