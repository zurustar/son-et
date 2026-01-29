package sprite

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
	for y := range height {
		for x := range width {
			img.Set(x, y, c)
		}
	}
	return img
}

// TestTextSpriteManager tests
func TestTextSpriteManager_CreateTextSprite(t *testing.T) {
	sm := NewSpriteManager()
	tsm := NewTextSpriteManager(sm)

	ts := tsm.CreateTextSprite(
		1,      // picID
		10, 20, // x, y
		"Hello",            // text
		color.Black,        // textColor
		color.White,        // bgColor
		basicfont.Face7x13, // face
		nil,                // parent
	)

	if ts == nil {
		t.Fatal("expected non-nil TextSprite")
	}

	if ts.GetPicID() != 1 {
		t.Errorf("expected picID 1, got %d", ts.GetPicID())
	}

	if ts.GetText() != "Hello" {
		t.Errorf("expected text 'Hello', got '%s'", ts.GetText())
	}

	x, y := ts.GetPosition()
	if x != 10 || y != 20 {
		t.Errorf("expected position (10, 20), got (%d, %d)", x, y)
	}

	if tsm.Count() != 1 {
		t.Errorf("expected count 1, got %d", tsm.Count())
	}
}

func TestTextSpriteManager_CreateTextSprite_EmptyText(t *testing.T) {
	sm := NewSpriteManager()
	tsm := NewTextSpriteManager(sm)

	ts := tsm.CreateTextSprite(
		1,      // picID
		10, 20, // x, y
		"",                 // empty text
		color.Black,        // textColor
		color.White,        // bgColor
		basicfont.Face7x13, // face
		nil,                // parent
	)

	if ts != nil {
		t.Error("expected nil TextSprite for empty text")
	}

	if tsm.Count() != 0 {
		t.Errorf("expected count 0, got %d", tsm.Count())
	}
}

func TestTextSpriteManager_CreateTextSprite_NilFace(t *testing.T) {
	sm := NewSpriteManager()
	tsm := NewTextSpriteManager(sm)

	ts := tsm.CreateTextSprite(
		1,      // picID
		10, 20, // x, y
		"Hello",     // text
		color.Black, // textColor
		color.White, // bgColor
		nil,         // nil face
		nil,         // parent
	)

	if ts != nil {
		t.Error("expected nil TextSprite for nil face")
	}

	if tsm.Count() != 0 {
		t.Errorf("expected count 0, got %d", tsm.Count())
	}
}

func TestTextSpriteManager_GetTextSprites(t *testing.T) {
	sm := NewSpriteManager()
	tsm := NewTextSpriteManager(sm)

	// 同じピクチャに複数のテキストを追加
	tsm.CreateTextSprite(1, 10, 20, "Hello", color.Black, color.White, basicfont.Face7x13, nil)
	tsm.CreateTextSprite(1, 10, 40, "World", color.Black, color.White, basicfont.Face7x13, nil)
	tsm.CreateTextSprite(2, 10, 20, "Other", color.Black, color.White, basicfont.Face7x13, nil)

	sprites := tsm.GetTextSprites(1)
	if len(sprites) != 2 {
		t.Errorf("expected 2 sprites for picID 1, got %d", len(sprites))
	}

	sprites = tsm.GetTextSprites(2)
	if len(sprites) != 1 {
		t.Errorf("expected 1 sprite for picID 2, got %d", len(sprites))
	}

	sprites = tsm.GetTextSprites(999)
	if sprites != nil && len(sprites) != 0 {
		t.Errorf("expected 0 sprites for non-existent picID, got %d", len(sprites))
	}
}

func TestTextSpriteManager_RemoveTextSprite(t *testing.T) {
	sm := NewSpriteManager()
	tsm := NewTextSpriteManager(sm)

	ts := tsm.CreateTextSprite(1, 10, 20, "Hello", color.Black, color.White, basicfont.Face7x13, nil)

	if tsm.Count() != 1 {
		t.Errorf("expected count 1, got %d", tsm.Count())
	}

	tsm.RemoveTextSprite(ts)

	if tsm.Count() != 0 {
		t.Errorf("expected count 0 after removal, got %d", tsm.Count())
	}
}

func TestTextSpriteManager_RemoveTextSpritesByPicID(t *testing.T) {
	sm := NewSpriteManager()
	tsm := NewTextSpriteManager(sm)

	tsm.CreateTextSprite(1, 10, 20, "Hello", color.Black, color.White, basicfont.Face7x13, nil)
	tsm.CreateTextSprite(1, 10, 40, "World", color.Black, color.White, basicfont.Face7x13, nil)
	tsm.CreateTextSprite(2, 10, 20, "Other", color.Black, color.White, basicfont.Face7x13, nil)

	if tsm.Count() != 3 {
		t.Errorf("expected count 3, got %d", tsm.Count())
	}

	tsm.RemoveTextSpritesByPicID(1)

	if tsm.Count() != 1 {
		t.Errorf("expected count 1 after removal, got %d", tsm.Count())
	}

	sprites := tsm.GetTextSprites(1)
	if sprites != nil && len(sprites) != 0 {
		t.Errorf("expected 0 sprites for picID 1 after removal, got %d", len(sprites))
	}
}

func TestTextSpriteManager_Clear(t *testing.T) {
	sm := NewSpriteManager()
	tsm := NewTextSpriteManager(sm)

	tsm.CreateTextSprite(1, 10, 20, "Hello", color.Black, color.White, basicfont.Face7x13, nil)
	tsm.CreateTextSprite(2, 10, 20, "World", color.Black, color.White, basicfont.Face7x13, nil)

	if tsm.Count() != 2 {
		t.Errorf("expected count 2, got %d", tsm.Count())
	}

	tsm.Clear()

	if tsm.Count() != 0 {
		t.Errorf("expected count 0 after clear, got %d", tsm.Count())
	}
}

func TestTextSprite_SetPosition(t *testing.T) {
	sm := NewSpriteManager()
	tsm := NewTextSpriteManager(sm)

	ts := tsm.CreateTextSprite(1, 10, 20, "Hello", color.Black, color.White, basicfont.Face7x13, nil)

	ts.SetPosition(50, 60)

	x, y := ts.GetPosition()
	if x != 50 || y != 60 {
		t.Errorf("expected position (50, 60), got (%d, %d)", x, y)
	}

	// スプライトの位置も更新されていることを確認
	sprite := ts.GetSprite()
	sx, sy := sprite.Position()
	if sx != 50 || sy != 60 {
		t.Errorf("expected sprite position (50, 60), got (%f, %f)", sx, sy)
	}
}

func TestTextSprite_SetVisible(t *testing.T) {
	sm := NewSpriteManager()
	tsm := NewTextSpriteManager(sm)

	ts := tsm.CreateTextSprite(1, 10, 20, "Hello", color.Black, color.White, basicfont.Face7x13, nil)

	ts.SetVisible(false)

	sprite := ts.GetSprite()
	if sprite.Visible() {
		t.Error("expected sprite to be invisible")
	}

	ts.SetVisible(true)

	if !sprite.Visible() {
		t.Error("expected sprite to be visible")
	}
}

func TestTextSprite_SetParent(t *testing.T) {
	sm := NewSpriteManager()
	tsm := NewTextSpriteManager(sm)

	ts := tsm.CreateTextSprite(1, 10, 20, "Hello", color.Black, color.White, basicfont.Face7x13, nil)

	// 親スプライトを作成
	parent := sm.CreateSpriteWithSize(100, 100, nil)
	parent.SetPosition(50, 50)

	ts.SetParent(parent)

	sprite := ts.GetSprite()
	if sprite.Parent() != parent {
		t.Error("expected parent to be set")
	}

	// 絶対位置が親の位置を考慮していることを確認
	absX, absY := sprite.AbsolutePosition()
	if absX != 60 || absY != 70 { // 10+50, 20+50
		t.Errorf("expected absolute position (60, 70), got (%f, %f)", absX, absY)
	}
}

func TestGetFontHeight(t *testing.T) {
	height := getFontHeight(basicfont.Face7x13)
	if height <= 0 {
		t.Errorf("expected positive font height, got %d", height)
	}

	// nilフェイスの場合はデフォルト値
	height = getFontHeight(nil)
	if height != 13 {
		t.Errorf("expected default font height 13, got %d", height)
	}
}
