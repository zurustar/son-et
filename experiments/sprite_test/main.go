// 実験: スプライト方式でのテキスト描画（アンチエイリアシング対応）

package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"
	"sort"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

const (
	screenWidth  = 800
	screenHeight = 600
)

// Sprite は汎用スプライト
type Sprite struct {
	ID       int
	Image    *ebiten.Image
	X, Y     float64
	ZOrder   int
	Visible  bool
	Alpha    float64
}

// SpriteManager はスプライトを管理
type SpriteManager struct {
	sprites  map[int]*Sprite
	nextID   int
}

func NewSpriteManager() *SpriteManager {
	return &SpriteManager{
		sprites: make(map[int]*Sprite),
		nextID:  1,
	}
}

func (sm *SpriteManager) CreateSprite(img *ebiten.Image, x, y float64, zOrder int) *Sprite {
	s := &Sprite{
		ID:      sm.nextID,
		Image:   img,
		X:       x,
		Y:       y,
		ZOrder:  zOrder,
		Visible: true,
		Alpha:   1.0,
	}
	sm.sprites[s.ID] = s
	sm.nextID++
	return s
}

func (sm *SpriteManager) Draw(screen *ebiten.Image) {
	var sorted []*Sprite
	for _, s := range sm.sprites {
		if s.Visible && s.Image != nil {
			sorted = append(sorted, s)
		}
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].ZOrder < sorted[j].ZOrder
	})

	for _, s := range sorted {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(s.X, s.Y)
		if s.Alpha < 1.0 {
			op.ColorScale.ScaleAlpha(float32(s.Alpha))
		}
		screen.DrawImage(s.Image, op)
	}
}

// Game
type Game struct {
	spriteManager   *SpriteManager
	face            font.Face
	initialized     bool
	screenshotTaken bool
	bgColor         color.RGBA
}

func NewGame() *Game {
	return &Game{
		spriteManager: NewSpriteManager(),
		bgColor:       color.RGBA{255, 255, 200, 255}, // 薄い黄色
	}
}

func (g *Game) initFonts() error {
	if g.initialized {
		return nil
	}

	fontPath := "/System/Library/Fonts/ヒラギノ角ゴシック W3.ttc"
	fontData, err := os.ReadFile(fontPath)
	if err != nil {
		return fmt.Errorf("failed to read font: %w", err)
	}

	collection, err := opentype.ParseCollection(fontData)
	if err != nil {
		return fmt.Errorf("failed to parse font collection: %w", err)
	}

	tt, err := collection.Font(0)
	if err != nil {
		return fmt.Errorf("failed to get font: %w", err)
	}

	g.face, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    24,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return fmt.Errorf("failed to create face: %w", err)
	}

	g.createTestSprites()

	g.initialized = true
	log.Println("Initialized successfully")
	return nil
}

func (g *Game) createTestSprites() {
	// 背景スプライト（Z=0）
	bgImg := ebiten.NewImage(screenWidth, screenHeight)
	bgImg.Fill(g.bgColor)
	g.spriteManager.CreateSprite(bgImg, 0, 0, 0)

	// テスト1: 黒→白の重ね書き
	g.createTextTest1()

	// テスト2: 複数テキストの重ね書き
	g.createTextTest2()
}

func (g *Game) createTextTest1() {
	text := "テスト Test"
	
	// ラベル（Y=50）
	labelImg := g.createLabelImage("テスト1: 黒→白 重ね書き")
	g.spriteManager.CreateSprite(labelImg, 30, 50, 100)

	// テスト文字列（Y=100、ラベルの下）
	bgImg := image.NewRGBA(image.Rect(0, 0, 200, 40))
	for py := 0; py < 40; py++ {
		for px := 0; px < 200; px++ {
			bgImg.Set(px, py, g.bgColor)
		}
	}

	// 黒文字のスプライト
	blackLayer := g.createTextLayerImage(bgImg, text, 0, 30, color.Black)
	g.spriteManager.CreateSprite(
		ebiten.NewImageFromImage(blackLayer),
		30, 100, 10,
	)

	// 白文字のスプライト（同じ位置、Z順序が上）
	whiteLayer := g.createTextLayerImage(bgImg, text, 0, 30, color.White)
	g.spriteManager.CreateSprite(
		ebiten.NewImageFromImage(whiteLayer),
		30, 100, 11,
	)
}

func (g *Game) createTextTest2() {
	// ラベル（Y=180）
	labelImg := g.createLabelImage("テスト2: 複数色テキストの重ね書き")
	g.spriteManager.CreateSprite(labelImg, 30, 180, 100)

	// テスト文字列（Y=220）
	bgImg := image.NewRGBA(image.Rect(0, 0, 300, 40))
	for py := 0; py < 40; py++ {
		for px := 0; px < 300; px++ {
			bgImg.Set(px, py, g.bgColor)
		}
	}

	// 赤文字
	redLayer := g.createTextLayerImage(bgImg, "赤文字", 0, 30, color.RGBA{255, 0, 0, 255})
	g.spriteManager.CreateSprite(
		ebiten.NewImageFromImage(redLayer),
		30, 220, 20,
	)

	// 青文字
	blueLayer := g.createTextLayerImage(bgImg, "青文字", 0, 30, color.RGBA{0, 0, 255, 255})
	g.spriteManager.CreateSprite(
		ebiten.NewImageFromImage(blueLayer),
		100, 220, 21,
	)

	// 緑文字
	greenLayer := g.createTextLayerImage(bgImg, "緑文字", 0, 30, color.RGBA{0, 128, 0, 255})
	g.spriteManager.CreateSprite(
		ebiten.NewImageFromImage(greenLayer),
		170, 220, 22,
	)
}

func (g *Game) createTextLayerImage(background *image.RGBA, text string, x, y int, textColor color.Color) *image.RGBA {
	bounds := background.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	tempImg := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(tempImg, tempImg.Bounds(), background, image.Point{}, draw.Src)

	bgCopy := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(bgCopy, bgCopy.Bounds(), background, image.Point{}, draw.Src)

	drawer := &font.Drawer{
		Dst:  tempImg,
		Src:  image.NewUniform(textColor),
		Face: g.face,
		Dot:  fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)},
	}
	drawer.DrawString(text)

	layerImg := image.NewRGBA(image.Rect(0, 0, width, height))
	for py := 0; py < height; py++ {
		for px := 0; px < width; px++ {
			bgPixel := bgCopy.At(px, py)
			textPixel := tempImg.At(px, py)

			bgR, bgG, bgB, _ := bgPixel.RGBA()
			txR, txG, txB, _ := textPixel.RGBA()

			if bgR != txR || bgG != txG || bgB != txB {
				layerImg.Set(px, py, textPixel)
			} else {
				layerImg.Set(px, py, color.RGBA{0, 0, 0, 0})
			}
		}
	}

	return layerImg
}

func (g *Game) createLabelImage(text string) *ebiten.Image {
	width := len(text) * 15
	height := 30

	img := image.NewRGBA(image.Rect(0, 0, width, height))

	drawer := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.RGBA{100, 100, 100, 255}),
		Face: g.face,
		Dot:  fixed.Point26_6{X: fixed.I(0), Y: fixed.I(24)},
	}
	drawer.DrawString(text)

	return ebiten.NewImageFromImage(img)
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
	if !g.initialized {
		screen.Fill(color.White)
		return
	}

	g.spriteManager.Draw(screen)

	// 操作説明
	if g.face != nil {
		infoImg := image.NewRGBA(image.Rect(0, 0, 600, 60))
		drawer := &font.Drawer{
			Dst:  infoImg,
			Src:  image.NewUniform(color.RGBA{100, 100, 100, 255}),
			Face: g.face,
			Dot:  fixed.Point26_6{X: fixed.I(0), Y: fixed.I(24)},
		}
		drawer.DrawString("[S] スクリーンショット  [Esc] 終了")
		drawer.Dot = fixed.Point26_6{X: fixed.I(0), Y: fixed.I(50)}
		drawer.Src = image.NewUniform(color.RGBA{180, 0, 0, 255})
		drawer.DrawString("※ テスト1で白文字だけ見えれば成功（黒い影がなければOK）")

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(30, 500)
		screen.DrawImage(ebiten.NewImageFromImage(infoImg), op)
	}

	if g.screenshotTaken {
		g.saveScreenshot(screen)
		g.screenshotTaken = false
	}
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
	ebiten.SetWindowTitle("Sprite + Text Layer Test")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game := NewGame()
	if err := ebiten.RunGame(game); err != nil && err != ebiten.Termination {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
