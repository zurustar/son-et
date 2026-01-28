// 実験: アンチエイリアシング検証 - システムフォント使用版
// ヒラギノ角ゴシックを使用してMSゴシックに近い見た目で検証

package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 800
	screenHeight = 600
)

type Game struct {
	hiraginoFace *text.GoTextFace
	initialized  bool
	screenshotTaken bool
	fontError    string
}

func NewGame() *Game {
	return &Game{}
}

func (g *Game) initFonts() error {
	if g.initialized {
		return nil
	}

	// ヒラギノ角ゴシックを読み込む
	fontPaths := []string{
		"/System/Library/Fonts/ヒラギノ角ゴシック W3.ttc",
		"/System/Library/Fonts/ヒラギノ角ゴシック W4.ttc",
		"/Library/Fonts/ヒラギノ角ゴ ProN W3.otf",
	}

	var src *text.GoTextFaceSource
	var err error

	for _, fontPath := range fontPaths {
		f, ferr := os.Open(fontPath)
		if ferr != nil {
			continue
		}
		defer f.Close()

		src, err = text.NewGoTextFaceSource(f)
		if err == nil {
			log.Printf("Loaded font: %s", fontPath)
			break
		}
	}

	if src == nil {
		g.fontError = "ヒラギノフォントが見つかりません"
		g.initialized = true
		return nil
	}

	g.hiraginoFace = &text.GoTextFace{
		Source: src,
		Size:   24,
	}

	g.initialized = true
	log.Println("Fonts initialized successfully")
	return nil
}

func (g *Game) Update() error {
	if err := g.initFonts(); err != nil {
		return err
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyS) {
		g.screenshotTaken = true
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	bgColor := color.RGBA{255, 200, 200, 255}
	screen.Fill(bgColor)

	if !g.initialized {
		return
	}

	if g.fontError != "" {
		vector.DrawFilledRect(screen, 10, 10, 400, 30, color.White, false)
		return
	}

	testText := "テスト Test 漢字"
	y := float64(30)

	// タイトル
	g.drawText(screen, "=== ヒラギノ角ゴシック + Vector描画テスト ===", 10, y, color.Black)
	y += 50

	// 1. 通常のtext.Draw（アンチエイリアスあり）
	g.drawText(screen, "1. text.Draw (通常 - AA有り):", 10, y, color.Black)
	y += 30
	g.drawText(screen, testText, 30, y, color.Black)
	g.drawText(screen, testText, 30, y, color.White)
	y += 50

	// 2. Vector Path (AntiAlias: false)
	g.drawText(screen, "2. Vector Path (AntiAlias: false):", 10, y, color.Black)
	y += 30
	g.drawVectorText(screen, testText, 30, y, color.Black, false)
	g.drawVectorText(screen, testText, 30, y, color.White, false)
	y += 50

	// 3. Vector Path (AntiAlias: true)
	g.drawText(screen, "3. Vector Path (AntiAlias: true):", 10, y, color.Black)
	y += 30
	g.drawVectorText(screen, testText, 30, y, color.Black, true)
	g.drawVectorText(screen, testText, 30, y, color.White, true)
	y += 60

	// 拡大表示
	g.drawText(screen, "=== 拡大表示 (4x) ===", 10, y, color.Black)
	y += 30

	// text.Draw 拡大
	smallImg1 := ebiten.NewImage(120, 35)
	smallImg1.Fill(bgColor)
	g.drawTextToImage(smallImg1, "Aa漢字", 5, 25, color.Black)
	g.drawTextToImage(smallImg1, "Aa漢字", 5, 25, color.White)

	g.drawText(screen, "text.Draw:", 30, y, color.Black)
	opScale := &ebiten.DrawImageOptions{}
	opScale.GeoM.Scale(4, 4)
	opScale.GeoM.Translate(130, y-10)
	opScale.Filter = ebiten.FilterNearest
	screen.DrawImage(smallImg1, opScale)

	// Vector(AA:off) 拡大
	smallImg2 := ebiten.NewImage(120, 35)
	smallImg2.Fill(bgColor)
	g.drawVectorText(smallImg2, "Aa漢字", 5, 25, color.Black, false)
	g.drawVectorText(smallImg2, "Aa漢字", 5, 25, color.White, false)

	g.drawText(screen, "Vector(AA:off):", 30, y+150, color.Black)
	opScale2 := &ebiten.DrawImageOptions{}
	opScale2.GeoM.Scale(4, 4)
	opScale2.GeoM.Translate(160, y+140)
	opScale2.Filter = ebiten.FilterNearest
	screen.DrawImage(smallImg2, opScale2)

	y += 310

	// 操作説明
	g.drawText(screen, "[S] スクリーンショット  [Esc] 終了", 10, y, color.RGBA{100, 100, 100, 255})

	if g.screenshotTaken {
		g.saveScreenshot(screen)
		g.screenshotTaken = false
	}
}

func (g *Game) drawText(screen *ebiten.Image, str string, x, y float64, clr color.Color) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(clr)
	text.Draw(screen, str, g.hiraginoFace, op)
}

func (g *Game) drawTextToImage(dst *ebiten.Image, str string, x, y float64, clr color.Color) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(clr)
	text.Draw(dst, str, g.hiraginoFace, op)
}

func (g *Game) drawVectorText(dst *ebiten.Image, str string, x, y float64, clr color.Color, antiAlias bool) {
	path := &vector.Path{}
	op := &text.LayoutOptions{}
	text.AppendVectorPath(path, str, g.hiraginoFace, op)

	r, gr, b, a := clr.RGBA()
	fr := float32(r) / 0xffff
	fg := float32(gr) / 0xffff
	fb := float32(b) / 0xffff
	fa := float32(a) / 0xffff

	vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)

	for i := range vs {
		vs[i].DstX += float32(x)
		vs[i].DstY += float32(y)
		vs[i].ColorR = fr
		vs[i].ColorG = fg
		vs[i].ColorB = fb
		vs[i].ColorA = fa
	}

	drawOp := &ebiten.DrawTrianglesOptions{
		AntiAlias: antiAlias,
	}

	whiteImg := ebiten.NewImage(1, 1)
	whiteImg.Fill(color.White)

	dst.DrawTriangles(vs, is, whiteImg, drawOp)
}

func (g *Game) saveScreenshot(screen *ebiten.Image) {
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("screenshot_%s.png", timestamp)

	bounds := screen.Bounds()
	img := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			img.Set(x, y, screen.At(x, y))
		}
	}

	file, err := os.Create(filename)
	if err != nil {
		log.Printf("Failed to create screenshot file: %v", err)
		return
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		log.Printf("Failed to encode screenshot: %v", err)
		return
	}

	log.Printf("Screenshot saved: %s", filename)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Anti-aliasing Test - Hiragino Gothic")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game := NewGame()
	if err := ebiten.RunGame(game); err != nil && err != ebiten.Termination {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
