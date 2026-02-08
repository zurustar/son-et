package graphics

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

// FontSettings はフォント設定を保持する
type FontSettings struct {
	Name      string // フォント名
	Size      int    // フォントサイズ（ピクセル）
	Charset   int    // 文字セット（Windows互換用）
	Weight    int    // フォントの太さ（400=通常、700=太字）
	Italic    bool   // イタリック
	Underline bool   // 下線
	Strikeout bool   // 取り消し線
}

// TextSettings はテキスト描画設定を保持する
type TextSettings struct {
	TextColor color.Color // 文字色
	BgColor   color.Color // 背景色
	BackMode  int         // 背景モード（0=背景あり/不透明, 1=透明）
}

// TextRenderer はテキスト描画を管理する
// スプライトシステム移行: LayerManagerは不要になった（TextSpriteで管理）
type TextRenderer struct {
	font     *FontSettings // 現在のフォント設定
	settings *TextSettings // 現在のテキスト設定
	face     font.Face     // 現在のフォントフェイス
	log      *slog.Logger  // ロガー
	mu       sync.RWMutex  // 排他制御
}

// フォントマッピング（Windows → クロスプラットフォーム）
// 要件 5.8: 指定されたフォントが見つからないとき、デフォルトフォントを使用する
var fontMapping = map[string][]string{
	// 全角フォント名
	"ＭＳ ゴシック":  {"Hiragino Kaku Gothic Pro", "Hiragino Sans", "Noto Sans JP", "IPAGothic"},
	"ＭＳ Ｐゴシック": {"Hiragino Kaku Gothic Pro", "Hiragino Sans", "Noto Sans JP", "IPAGothic"},
	"ＭＳ 明朝":    {"Hiragino Mincho Pro", "Hiragino Mincho ProN", "Noto Serif JP", "IPAMincho"},
	"ＭＳ Ｐ明朝":   {"Hiragino Mincho Pro", "Hiragino Mincho ProN", "Noto Serif JP", "IPAMincho"},
	// 半角フォント名（小文字）
	"ms gothic":  {"Hiragino Kaku Gothic Pro", "Hiragino Sans", "Noto Sans JP", "IPAGothic"},
	"ms mincho":  {"Hiragino Mincho Pro", "Hiragino Mincho ProN", "Noto Serif JP", "IPAMincho"},
	"ms pgothic": {"Hiragino Kaku Gothic Pro", "Hiragino Sans", "Noto Sans JP", "IPAGothic"},
	"ms pmincho": {"Hiragino Mincho Pro", "Hiragino Mincho ProN", "Noto Serif JP", "IPAMincho"},
}

// NewTextRenderer は新しい TextRenderer を作成する
func NewTextRenderer() *TextRenderer {
	tr := &TextRenderer{
		font: &FontSettings{
			Name:   "default",
			Size:   12,
			Weight: 400, // 通常の太さ
		},
		settings: &TextSettings{
			TextColor: color.RGBA{0, 0, 0, 255},       // デフォルトは黒
			BgColor:   color.RGBA{255, 255, 255, 255}, // デフォルトは白
			BackMode:  0,                              // 背景あり/不透明 (0=背景あり, 1=透明)
		},
		face: basicfont.Face7x13, // デフォルトフォント
		log:  slog.Default(),
	}
	return tr
}

// NewTextRendererWithLogger は新しい TextRenderer をロガー付きで作成する
func NewTextRendererWithLogger(log *slog.Logger) *TextRenderer {
	tr := NewTextRenderer()
	tr.log = log
	return tr
}

// SetLayerManager は削除されました
// Deprecated: スプライトシステム移行により不要になった

// SetFont はフォント設定を変更する
// 要件 5.1: SetFont(font_name, size, charset, weight, italic, underline, strikeout)が呼ばれたとき、フォント設定を変更する
func (tr *TextRenderer) SetFont(name string, size int, opts ...FontOption) error {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	// サイズの妥当性チェック
	if size <= 0 {
		size = 12 // デフォルトサイズ
	}
	// 大きなフォントサイズは許可する（FILLYスクリプトでは640などの大きなサイズが使われる）
	// これは「ピクチャー全体を背景色で塗りつぶす」テクニックとして使用される
	// ただし、実際のフォント描画では最大サイズを制限する
	actualSize := size
	if size > 200 {
		tr.log.Debug("SetFont: Large size detected, will use for background fill calculation",
			"requestedSize", size)
		// 実際のフォント描画には制限されたサイズを使用
		actualSize = 72 // 大きめのサイズを使用
	}

	// フォント設定を更新
	tr.font.Name = name
	tr.font.Size = size // 元のサイズを保持（背景塗りつぶし計算用）

	// オプションを適用
	for _, opt := range opts {
		opt(tr.font)
	}

	// フォントを読み込む（実際の描画用サイズを使用）
	face, err := tr.loadFont(name, actualSize)
	if err != nil {
		tr.log.Warn("Failed to load font, using fallback",
			"fontName", name,
			"error", err)
		// フォールバックフォントを使用
		tr.face = basicfont.Face7x13
		return nil // エラーは返さない（要件 5.8）
	}

	tr.face = face
	tr.log.Debug("Font set successfully",
		"name", name,
		"size", size,
		"weight", tr.font.Weight,
		"italic", tr.font.Italic)

	return nil
}

// FontOption はフォント設定のオプション
type FontOption func(*FontSettings)

// WithCharset は文字セットを設定する
func WithCharset(charset int) FontOption {
	return func(fs *FontSettings) {
		fs.Charset = charset
	}
}

// WithWeight はフォントの太さを設定する
func WithWeight(weight int) FontOption {
	return func(fs *FontSettings) {
		fs.Weight = weight
	}
}

// WithItalic はイタリックを設定する
func WithItalic(italic bool) FontOption {
	return func(fs *FontSettings) {
		fs.Italic = italic
	}
}

// WithUnderline は下線を設定する
func WithUnderline(underline bool) FontOption {
	return func(fs *FontSettings) {
		fs.Underline = underline
	}
}

// WithStrikeout は取り消し線を設定する
func WithStrikeout(strikeout bool) FontOption {
	return func(fs *FontSettings) {
		fs.Strikeout = strikeout
	}
}

// SetTextColor は文字色を設定する
// 要件 5.3: TextColor(color)が呼ばれたとき、文字色を設定する
func (tr *TextRenderer) SetTextColor(c color.Color) {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	tr.settings.TextColor = c
}

// SetBgColor は背景色を設定する
// 要件 5.4: BgColor(color)が呼ばれたとき、背景色を設定する
func (tr *TextRenderer) SetBgColor(c color.Color) {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	tr.settings.BgColor = c
}

// SetBackMode は背景モードを設定する
// 要件 5.5: BackMode(mode)が呼ばれたとき、背景モードを設定する（0=背景あり/不透明, 1=透明）
func (tr *TextRenderer) SetBackMode(mode int) {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	tr.settings.BackMode = mode
}

// TextWrite はピクチャーに文字列を描画する
// 要件 5.2: TextWrite(pic_no, x, y, text)が呼ばれたとき、指定されたピクチャーに文字列を描画する
// スプライトシステム: TextSpriteはGraphicsSystem.TextWrite()で作成される
// レイヤー方式: 背景に文字を描画し、差分を取って文字部分だけを抽出
// これにより、同じ位置に別の色で描画しても前の文字の影が残らない
func (tr *TextRenderer) TextWrite(pic *Picture, x, y int, text string) error {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	if pic == nil {
		return ErrPictureNotFound
	}

	if pic.Image == nil {
		return fmt.Errorf("picture image is nil")
	}

	bounds := pic.Image.Bounds()

	// 元の背景画像を使用（テキスト描画前の状態）
	// これにより、同じ位置に別の色で描画しても前の文字の影が残らない
	var background *image.RGBA
	if pic.OriginalImage != nil {
		background = pic.OriginalImage
	} else {
		// OriginalImageがない場合は白で初期化
		background = image.NewRGBA(bounds)
		draw.Draw(background, bounds, &image.Uniform{color.RGBA{255, 255, 255, 255}}, image.Point{}, draw.Src)
	}

	// 背景モードが不透明(BackMode=0)の場合、先に背景を描画
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, background, image.Point{}, draw.Src)

	if tr.settings.BackMode == 0 {
		// BackMode=0: 背景あり/不透明
		// テキストの境界を計算
		textBounds, _ := font.BoundString(tr.face, text)
		width := (textBounds.Max.X - textBounds.Min.X).Ceil()
		height := (textBounds.Max.Y - textBounds.Min.Y).Ceil()

		// 大きなフォントサイズの場合、ピクチャー全体を背景色で塗りつぶす
		// これはFILLYスクリプトで「ピクチャーを白で塗りつぶす」テクニックとして使用される
		// 例: SetFont(640,...); TextWrite("  ", pic, 0, 0);
		if tr.font.Size > 200 {
			// フォントサイズが大きい場合、ピクチャー全体を背景色で塗りつぶす
			draw.Draw(rgba, bounds, &image.Uniform{tr.settings.BgColor}, image.Point{}, draw.Src)
			tr.log.Debug("TextWrite: Large font size, filling entire picture with background color",
				"fontSize", tr.font.Size,
				"bgColor", tr.settings.BgColor)
		} else {
			// 通常のフォントサイズの場合、テキストの境界だけを塗りつぶす
			bgRect := image.Rect(x, y, x+width, y+height+tr.font.Size)
			draw.Draw(rgba, bgRect, &image.Uniform{tr.settings.BgColor}, image.Point{}, draw.Src)
		}
	}

	// テキストを直接描画
	drawer := &font.Drawer{
		Dst:  rgba,
		Src:  image.NewUniform(tr.settings.TextColor),
		Face: tr.face,
		Dot:  fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y + tr.font.Size)},
	}
	drawer.DrawString(text)

	// Ebitengine画像に変換して戻す
	pic.Image = ebiten.NewImageFromImage(rgba)

	tr.log.Debug("TextWrite completed",
		"text", text,
		"x", x,
		"y", y,
		"picWidth", pic.Width,
		"picHeight", pic.Height)

	return nil
}

// MeasureText はテキストの幅と高さを返す
func (tr *TextRenderer) MeasureText(text string) (int, int) {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	bounds, advance := font.BoundString(tr.face, text)
	width := advance.Ceil()
	height := (bounds.Max.Y - bounds.Min.Y).Ceil()
	return width, height
}

// GetFontSettings は現在のフォント設定を返す
func (tr *TextRenderer) GetFontSettings() FontSettings {
	tr.mu.RLock()
	defer tr.mu.RUnlock()
	return *tr.font
}

// GetTextSettings は現在のテキスト設定を返す
func (tr *TextRenderer) GetTextSettings() TextSettings {
	tr.mu.RLock()
	defer tr.mu.RUnlock()
	return *tr.settings
}

// GetFace は現在のフォントフェイスを返す
func (tr *TextRenderer) GetFace() font.Face {
	tr.mu.RLock()
	defer tr.mu.RUnlock()
	return tr.face
}

// loadFont はフォントを読み込む
// 要件 5.8: 指定されたフォントが見つからないとき、デフォルトフォントを使用する
func (tr *TextRenderer) loadFont(name string, size int) (font.Face, error) {
	// 1. フォントマッピングでフォールバック候補を取得
	candidates := []string{name}
	if mapped, ok := fontMapping[strings.ToLower(name)]; ok {
		candidates = append(candidates, mapped...)
	}
	// 全角名でもチェック
	if mapped, ok := fontMapping[name]; ok {
		candidates = append(candidates, mapped...)
	}

	// 2. システムフォントを順番に検索
	fontPaths := getSystemFontPaths()
	for _, fontName := range candidates {
		for _, basePath := range fontPaths {
			// フォントファイルを検索
			face, err := tr.tryLoadFontFromPath(basePath, fontName, size)
			if err == nil && face != nil {
				return face, nil
			}
		}
	}

	// 3. 日本語フォントのパスを直接試す
	japaneseFontPaths := getJapaneseFontPaths()
	for _, fontPath := range japaneseFontPaths {
		if _, err := os.Stat(fontPath); err == nil {
			face, err := tr.loadFontFromFile(fontPath, float64(size))
			if err == nil && face != nil {
				return face, nil
			}
		}
	}

	// 4. フォールバック: basicfontを使用
	return nil, fmt.Errorf("font not found: %s", name)
}

// tryLoadFontFromPath はパスからフォントを読み込もうとする
func (tr *TextRenderer) tryLoadFontFromPath(basePath, fontName string, size int) (font.Face, error) {
	// 一般的なフォントファイル拡張子
	extensions := []string{".ttf", ".ttc", ".otf"}

	for _, ext := range extensions {
		// フォント名をファイル名に変換（スペースを除去）
		fileName := strings.ReplaceAll(fontName, " ", "") + ext
		fullPath := basePath + "/" + fileName

		if _, err := os.Stat(fullPath); err == nil {
			face, err := tr.loadFontFromFile(fullPath, float64(size))
			if err == nil {
				return face, nil
			}
		}

		// スペースを含むファイル名も試す
		fileName = fontName + ext
		fullPath = basePath + "/" + fileName

		if _, err := os.Stat(fullPath); err == nil {
			face, err := tr.loadFontFromFile(fullPath, float64(size))
			if err == nil {
				return face, nil
			}
		}
	}

	return nil, fmt.Errorf("font not found in path: %s/%s", basePath, fontName)
}

// loadFontFromFile はファイルからフォントを読み込む
func (tr *TextRenderer) loadFontFromFile(path string, size float64) (font.Face, error) {
	fontData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read font file: %w", err)
	}

	// 単一フォントとして解析を試みる
	tt, err := opentype.Parse(fontData)
	if err != nil {
		// フォントコレクション（.ttc）として解析を試みる
		collection, err := opentype.ParseCollection(fontData)
		if err != nil {
			return nil, fmt.Errorf("failed to parse font: %w", err)
		}
		// コレクションの最初のフォントを使用
		if collection.NumFonts() > 0 {
			tt, err = collection.Font(0)
			if err != nil {
				return nil, fmt.Errorf("failed to get font from collection: %w", err)
			}
		} else {
			return nil, fmt.Errorf("font collection is empty")
		}
	}

	// フォントフェイスを作成
	// Hinting: font.HintingFull でアンチエイリアスを最小化
	face, err := opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create font face: %w", err)
	}

	return face, nil
}

// getSystemFontPaths はシステムフォントのパスを返す
func getSystemFontPaths() []string {
	switch runtime.GOOS {
	case "darwin":
		return []string{
			"/System/Library/Fonts",
			"/Library/Fonts",
			os.ExpandEnv("$HOME/Library/Fonts"),
		}
	case "linux":
		return []string{
			"/usr/share/fonts",
			"/usr/local/share/fonts",
			os.ExpandEnv("$HOME/.fonts"),
			os.ExpandEnv("$HOME/.local/share/fonts"),
		}
	case "windows":
		return []string{
			os.ExpandEnv("$WINDIR/Fonts"),
			os.ExpandEnv("$LOCALAPPDATA/Microsoft/Windows/Fonts"),
		}
	default:
		return nil
	}
}

// getJapaneseFontPaths は日本語フォントの直接パスを返す
func getJapaneseFontPaths() []string {
	switch runtime.GOOS {
	case "darwin":
		return []string{
			"/System/Library/Fonts/ヒラギノ角ゴシック W3.ttc",
			"/System/Library/Fonts/ヒラギノ角ゴシック W4.ttc",
			"/System/Library/Fonts/ヒラギノ明朝 ProN.ttc",
			"/Library/Fonts/Arial Unicode.ttf",
			"/System/Library/Fonts/Supplemental/Arial Unicode.ttf",
			"/System/Library/Fonts/ヒラギノ丸ゴ ProN W4.ttc",
		}
	case "linux":
		return []string{
			"/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc",
			"/usr/share/fonts/truetype/noto/NotoSansCJK-Regular.ttc",
			"/usr/share/fonts/opentype/ipafont-gothic/ipag.ttf",
			"/usr/share/fonts/truetype/fonts-japanese-gothic.ttf",
		}
	case "windows":
		return []string{
			os.ExpandEnv("$WINDIR/Fonts/msgothic.ttc"),
			os.ExpandEnv("$WINDIR/Fonts/msmincho.ttc"),
			os.ExpandEnv("$WINDIR/Fonts/meiryo.ttc"),
			os.ExpandEnv("$WINDIR/Fonts/YuGothM.ttc"),
		}
	default:
		return nil
	}
}
