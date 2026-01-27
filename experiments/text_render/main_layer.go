//go:build ignore
// +build ignore

package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

// TextLayer はテキスト描画のレイヤーを表す
type TextLayer struct {
	Image  *image.RGBA // 透明背景 + 文字
	X, Y   int         // 描画位置
	Width  int
	Height int
}

func main() {
	// 日本語フォントを読み込む
	fontPath := "/System/Library/Fonts/ヒラギノ角ゴシック W3.ttc"
	fontData, err := os.ReadFile(fontPath)
	if err != nil {
		log.Fatalf("Failed to read font: %v", err)
	}

	collection, err := opentype.ParseCollection(fontData)
	if err != nil {
		log.Fatalf("Failed to parse font collection: %v", err)
	}

	tt, err := collection.Font(0)
	if err != nil {
		log.Fatalf("Failed to get font: %v", err)
	}

	face, err := opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    24,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatalf("Failed to create face: %v", err)
	}

	// 背景画像（文字なし）
	width, height := 400, 100
	background := image.NewRGBA(image.Rect(0, 0, width, height))
	bgColor := color.RGBA{255, 255, 200, 255} // 黄色
	draw.Draw(background, background.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)

	text := "テスト"
	x, y := 10, 50
	fontSize := 24

	// テキストの境界を計算
	bounds, _ := font.BoundString(face, text)
	textWidth := (bounds.Max.X - bounds.Min.X).Ceil() + 10
	textHeight := fontSize + 10

	var layers []*TextLayer

	// === TextWrite 1回目: 黒で描画 ===
	layer1 := createTextLayer(background, face, text, x, y, fontSize, textWidth, textHeight, color.Black)
	layers = append(layers, layer1)
	log.Println("Layer 1 (黒) created")

	// === TextWrite 2回目: 白で描画（同じ位置） ===
	layer2 := createTextLayer(background, face, text, x, y, fontSize, textWidth, textHeight, color.White)
	layers = append(layers, layer2)
	log.Println("Layer 2 (白) created")

	// === 最終合成 ===
	// 背景をコピー
	finalImg := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(finalImg, finalImg.Bounds(), background, image.Point{}, draw.Src)

	// レイヤーを順番に重ねる（後のレイヤーが上）
	for i, layer := range layers {
		log.Printf("Compositing layer %d at (%d, %d)", i, layer.X, layer.Y)
		// レイヤーを描画（アルファブレンディング）
		destRect := image.Rect(layer.X, layer.Y, layer.X+layer.Width, layer.Y+layer.Height)
		draw.Draw(finalImg, destRect, layer.Image, image.Point{}, draw.Over)
	}

	// 画像を保存
	outFile, err := os.Create("output_layer.png")
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer outFile.Close()

	if err := png.Encode(outFile, finalImg); err != nil {
		log.Fatalf("Failed to encode PNG: %v", err)
	}

	log.Println("Output saved to output_layer.png")
	log.Println("レイヤー方式: 黒レイヤー → 白レイヤー の順で合成")
}

// createTextLayer は背景に文字を描画し、差分を取って透明背景のレイヤーを作成する
func createTextLayer(background *image.RGBA, face font.Face, text string, x, y, fontSize, textWidth, textHeight int, textColor color.Color) *TextLayer {
	destX := x
	destY := y - fontSize

	// 1. 背景をコピーして一時画像を作成
	tempImg := image.NewRGBA(image.Rect(0, 0, textWidth, textHeight))
	for py := 0; py < textHeight; py++ {
		for px := 0; px < textWidth; px++ {
			tempImg.Set(px, py, background.At(destX+px, destY+py))
		}
	}

	// 2. 一時画像に文字を描画（アルファブレンディング）
	drawer := &font.Drawer{
		Dst:  tempImg,
		Src:  image.NewUniform(textColor),
		Face: face,
		Dot:  fixed.Point26_6{X: fixed.I(0), Y: fixed.I(fontSize)},
	}
	drawer.DrawString(text)

	// 3. 背景との差分を取って、文字部分だけを透明背景で抽出
	layerImg := image.NewRGBA(image.Rect(0, 0, textWidth, textHeight))
	for py := 0; py < textHeight; py++ {
		for px := 0; px < textWidth; px++ {
			bgPixel := background.At(destX+px, destY+py)
			textPixel := tempImg.At(px, py)

			bgR, bgG, bgB, _ := bgPixel.RGBA()
			txR, txG, txB, _ := textPixel.RGBA()

			// 色が変わっていれば文字部分
			if bgR != txR || bgG != txG || bgB != txB {
				// 文字部分はそのまま（不透明）
				layerImg.Set(px, py, textPixel)
			} else {
				// 背景と同じなら透明
				layerImg.Set(px, py, color.RGBA{0, 0, 0, 0})
			}
		}
	}

	return &TextLayer{
		Image:  layerImg,
		X:      destX,
		Y:      destY,
		Width:  textWidth,
		Height: textHeight,
	}
}
