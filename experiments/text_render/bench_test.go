package main

import (
	"image"
	"image/color"
	"image/draw"
	"testing"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// BenchmarkLayerMethod はレイヤー方式のベンチマーク
func BenchmarkLayerMethod(b *testing.B) {
	face := basicfont.Face7x13
	width, height := 400, 100
	background := image.NewRGBA(image.Rect(0, 0, width, height))
	bgColor := color.RGBA{255, 255, 200, 255}
	draw.Draw(background, background.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)

	text := "テスト"
	x, y := 10, 50
	fontSize := 13
	textWidth, textHeight := 100, 20

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// レイヤー作成
		_ = createTextLayerBench(background, face, text, x, y, fontSize, textWidth, textHeight, color.Black)
	}
}

// BenchmarkComposite は合成のベンチマーク
func BenchmarkComposite(b *testing.B) {
	face := basicfont.Face7x13
	width, height := 400, 100
	background := image.NewRGBA(image.Rect(0, 0, width, height))
	bgColor := color.RGBA{255, 255, 200, 255}
	draw.Draw(background, background.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)

	text := "テスト"
	x, y := 10, 50
	fontSize := 13
	textWidth, textHeight := 100, 20

	// 10個のレイヤーを事前に作成
	layers := make([]*TextLayerBench, 10)
	for i := 0; i < 10; i++ {
		layers[i] = createTextLayerBench(background, face, text, x, y, fontSize, textWidth, textHeight, color.Black)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 最終合成
		finalImg := image.NewRGBA(image.Rect(0, 0, width, height))
		draw.Draw(finalImg, finalImg.Bounds(), background, image.Point{}, draw.Src)
		for _, layer := range layers {
			destRect := image.Rect(layer.X, layer.Y, layer.X+layer.Width, layer.Y+layer.Height)
			draw.Draw(finalImg, destRect, layer.Image, image.Point{}, draw.Over)
		}
	}
}

type TextLayerBench struct {
	Image  *image.RGBA
	X, Y   int
	Width  int
	Height int
}

func createTextLayerBench(background *image.RGBA, face font.Face, text string, x, y, fontSize, textWidth, textHeight int, textColor color.Color) *TextLayerBench {
	destX := x
	destY := y - fontSize

	tempImg := image.NewRGBA(image.Rect(0, 0, textWidth, textHeight))
	for py := 0; py < textHeight; py++ {
		for px := 0; px < textWidth; px++ {
			tempImg.Set(px, py, background.At(destX+px, destY+py))
		}
	}

	drawer := &font.Drawer{
		Dst:  tempImg,
		Src:  image.NewUniform(textColor),
		Face: face,
		Dot:  fixed.Point26_6{X: fixed.I(0), Y: fixed.I(fontSize)},
	}
	drawer.DrawString(text)

	layerImg := image.NewRGBA(image.Rect(0, 0, textWidth, textHeight))
	for py := 0; py < textHeight; py++ {
		for px := 0; px < textWidth; px++ {
			bgPixel := background.At(destX+px, destY+py)
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

	return &TextLayerBench{
		Image:  layerImg,
		X:      destX,
		Y:      destY,
		Width:  textWidth,
		Height: textHeight,
	}
}
