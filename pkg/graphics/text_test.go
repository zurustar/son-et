package graphics

import (
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestNewTextRenderer(t *testing.T) {
	tr := NewTextRenderer()
	if tr == nil {
		t.Fatal("NewTextRenderer returned nil")
	}

	// デフォルト設定の確認
	fontSettings := tr.GetFontSettings()
	if fontSettings.Name != "default" {
		t.Errorf("expected font name 'default', got '%s'", fontSettings.Name)
	}
	if fontSettings.Size != 12 {
		t.Errorf("expected font size 12, got %d", fontSettings.Size)
	}
	if fontSettings.Weight != 400 {
		t.Errorf("expected font weight 400, got %d", fontSettings.Weight)
	}

	textSettings := tr.GetTextSettings()
	if textSettings.BackMode != 0 {
		t.Errorf("expected back mode 0, got %d", textSettings.BackMode)
	}
}

func TestSetFont(t *testing.T) {
	tr := NewTextRenderer()

	// 存在しないフォントを設定（フォールバックが動作するはず）
	err := tr.SetFont("NonExistentFont", 16)
	if err != nil {
		t.Errorf("SetFont should not return error for non-existent font: %v", err)
	}

	fontSettings := tr.GetFontSettings()
	if fontSettings.Name != "NonExistentFont" {
		t.Errorf("expected font name 'NonExistentFont', got '%s'", fontSettings.Name)
	}
	if fontSettings.Size != 16 {
		t.Errorf("expected font size 16, got %d", fontSettings.Size)
	}
}

func TestSetFontWithOptions(t *testing.T) {
	tr := NewTextRenderer()

	err := tr.SetFont("TestFont", 14,
		WithWeight(700),
		WithItalic(true),
		WithUnderline(true),
		WithStrikeout(true),
		WithCharset(128),
	)
	if err != nil {
		t.Errorf("SetFont should not return error: %v", err)
	}

	fontSettings := tr.GetFontSettings()
	if fontSettings.Weight != 700 {
		t.Errorf("expected weight 700, got %d", fontSettings.Weight)
	}
	if !fontSettings.Italic {
		t.Error("expected italic to be true")
	}
	if !fontSettings.Underline {
		t.Error("expected underline to be true")
	}
	if !fontSettings.Strikeout {
		t.Error("expected strikeout to be true")
	}
	if fontSettings.Charset != 128 {
		t.Errorf("expected charset 128, got %d", fontSettings.Charset)
	}
}

func TestSetFontSizeValidation(t *testing.T) {
	tr := NewTextRenderer()

	// 負のサイズはデフォルトに
	err := tr.SetFont("Test", -5)
	if err != nil {
		t.Errorf("SetFont should not return error: %v", err)
	}
	fontSettings := tr.GetFontSettings()
	if fontSettings.Size != 12 {
		t.Errorf("expected default size 12 for negative input, got %d", fontSettings.Size)
	}

	// 極端に大きいサイズは元のサイズを保持する（背景塗りつぶし計算用）
	// ただし、実際のフォント描画には制限されたサイズが使用される
	err = tr.SetFont("Test", 500)
	if err != nil {
		t.Errorf("SetFont should not return error: %v", err)
	}
	fontSettings = tr.GetFontSettings()
	if fontSettings.Size != 500 {
		t.Errorf("expected original size 500 for large input, got %d", fontSettings.Size)
	}
}

func TestSetTextColor(t *testing.T) {
	tr := NewTextRenderer()

	testColor := color.RGBA{255, 128, 64, 255}
	tr.SetTextColor(testColor)

	textSettings := tr.GetTextSettings()
	r, g, b, a := textSettings.TextColor.RGBA()
	expectedR, expectedG, expectedB, expectedA := testColor.RGBA()

	if r != expectedR || g != expectedG || b != expectedB || a != expectedA {
		t.Errorf("text color mismatch: expected %v, got %v", testColor, textSettings.TextColor)
	}
}

func TestSetBgColor(t *testing.T) {
	tr := NewTextRenderer()

	testColor := color.RGBA{32, 64, 128, 255}
	tr.SetBgColor(testColor)

	textSettings := tr.GetTextSettings()
	r, g, b, a := textSettings.BgColor.RGBA()
	expectedR, expectedG, expectedB, expectedA := testColor.RGBA()

	if r != expectedR || g != expectedG || b != expectedB || a != expectedA {
		t.Errorf("bg color mismatch: expected %v, got %v", testColor, textSettings.BgColor)
	}
}

func TestSetBackMode(t *testing.T) {
	tr := NewTextRenderer()

	// 透明モード
	tr.SetBackMode(0)
	textSettings := tr.GetTextSettings()
	if textSettings.BackMode != 0 {
		t.Errorf("expected back mode 0, got %d", textSettings.BackMode)
	}

	// 不透明モード
	tr.SetBackMode(1)
	textSettings = tr.GetTextSettings()
	if textSettings.BackMode != 1 {
		t.Errorf("expected back mode 1, got %d", textSettings.BackMode)
	}
}

func TestTextWriteNilPicture(t *testing.T) {
	tr := NewTextRenderer()

	err := tr.TextWrite(nil, 0, 0, "test")
	if err != ErrPictureNotFound {
		t.Errorf("expected ErrPictureNotFound, got %v", err)
	}
}

func TestTextWriteNilImage(t *testing.T) {
	tr := NewTextRenderer()

	pic := &Picture{
		ID:     0,
		Image:  nil,
		Width:  100,
		Height: 100,
	}

	err := tr.TextWrite(pic, 0, 0, "test")
	if err == nil {
		t.Error("expected error for nil image")
	}
}

func TestTextWrite(t *testing.T) {
	tr := NewTextRenderer()

	// テスト用のピクチャーを作成
	img := ebiten.NewImage(200, 100)
	pic := &Picture{
		ID:     0,
		Image:  img,
		Width:  200,
		Height: 100,
	}

	// テキストを描画
	err := tr.TextWrite(pic, 10, 10, "Hello")
	if err != nil {
		t.Errorf("TextWrite failed: %v", err)
	}

	// ピクチャーが更新されていることを確認
	if pic.Image == nil {
		t.Error("picture image should not be nil after TextWrite")
	}
}

func TestMeasureText(t *testing.T) {
	tr := NewTextRenderer()

	width, height := tr.MeasureText("Hello")
	if width <= 0 {
		t.Errorf("expected positive width, got %d", width)
	}
	if height <= 0 {
		t.Errorf("expected positive height, got %d", height)
	}

	// 空文字列
	width, height = tr.MeasureText("")
	if width != 0 {
		t.Errorf("expected width 0 for empty string, got %d", width)
	}
}

func TestFontFallback(t *testing.T) {
	tr := NewTextRenderer()

	// MSゴシックを設定（Windowsフォント）
	// macOS/Linuxではフォールバックが動作するはず
	err := tr.SetFont("ＭＳ ゴシック", 12)
	if err != nil {
		t.Errorf("SetFont should not return error: %v", err)
	}

	// フォントが設定されていることを確認（フォールバックでも可）
	fontSettings := tr.GetFontSettings()
	if fontSettings.Name != "ＭＳ ゴシック" {
		t.Errorf("expected font name 'ＭＳ ゴシック', got '%s'", fontSettings.Name)
	}

	// テキスト描画が動作することを確認
	img := ebiten.NewImage(200, 100)
	pic := &Picture{
		ID:     0,
		Image:  img,
		Width:  200,
		Height: 100,
	}

	err = tr.TextWrite(pic, 10, 10, "テスト")
	if err != nil {
		t.Errorf("TextWrite failed with fallback font: %v", err)
	}
}

func TestFontMappingLowerCase(t *testing.T) {
	tr := NewTextRenderer()

	// 小文字のフォント名でも動作することを確認
	err := tr.SetFont("ms gothic", 12)
	if err != nil {
		t.Errorf("SetFont should not return error: %v", err)
	}

	fontSettings := tr.GetFontSettings()
	if fontSettings.Name != "ms gothic" {
		t.Errorf("expected font name 'ms gothic', got '%s'", fontSettings.Name)
	}
}

func TestConcurrentAccess(t *testing.T) {
	tr := NewTextRenderer()

	// 並行アクセスのテスト
	done := make(chan bool)

	// 複数のゴルーチンから同時にアクセス
	for i := 0; i < 10; i++ {
		go func(id int) {
			tr.SetTextColor(color.RGBA{uint8(id * 25), 0, 0, 255})
			tr.SetBgColor(color.RGBA{0, uint8(id * 25), 0, 255})
			tr.SetBackMode(id % 2)
			_ = tr.GetFontSettings()
			_ = tr.GetTextSettings()
			done <- true
		}(i)
	}

	// すべてのゴルーチンが完了するのを待つ
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestGetSystemFontPaths(t *testing.T) {
	paths := getSystemFontPaths()
	if len(paths) == 0 {
		t.Error("expected at least one system font path")
	}
}

func TestGetJapaneseFontPaths(t *testing.T) {
	paths := getJapaneseFontPaths()
	if len(paths) == 0 {
		t.Error("expected at least one Japanese font path")
	}
}

// TestTextRendererLayerManagerIntegration はTextRendererとLayerManagerの統合をテストする
// 要件 8.3: TextRendererとLayerManagerを統合する
// 要件 2.5: TextWriteが呼び出されたときに対応するText_Layerを作成する
// 注意: このテストは旧レイヤーシステム（LayerManager）を使用しているため、
// スプライトシステムへの移行に伴いスキップされます。
// スプライトシステムでのテキスト描画は text_sprite_test.go でテストされています。
func TestTextRendererLayerManagerIntegration(t *testing.T) {
	t.Skip("Skipped: This test uses the old LayerManager which has been replaced by the sprite system. See text_sprite_test.go for text rendering tests.")
}

// TestGraphicsSystem_TextRendererLayerManagerIntegration はGraphicsSystemでのTextRendererとLayerManagerの統合をテストする
// 注意: このテストは旧レイヤーシステム（LayerManager）を使用しているため、
// スプライトシステムへの移行に伴いスキップされます。
// スプライトシステムでのテキスト描画は text_sprite_test.go でテストされています。
func TestGraphicsSystem_TextRendererLayerManagerIntegration(t *testing.T) {
	t.Skip("Skipped: This test uses the old LayerManager which has been replaced by the sprite system. See text_sprite_test.go for text rendering tests.")
}
