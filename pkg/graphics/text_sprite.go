// Package graphics provides text sprite creation with anti-aliasing removal.
package graphics

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// TextSpriteOptions はテキストスプライト作成のオプション
type TextSpriteOptions struct {
	// Text は描画するテキスト
	Text string
	// TextColor はテキストの色
	TextColor color.Color
	// Face はフォントフェイス
	Face font.Face
	// BgColor は差分抽出用の背景色
	BgColor color.Color
	// Width は画像の幅（0の場合は自動計算）
	Width int
	// Height は画像の高さ（0の場合は自動計算）
	Height int
	// X はテキストのX座標
	X int
	// Y はテキストのベースラインY座標
	Y int
}

// CreateTextSpriteImage は差分抽出方式でテキストスプライト用の画像を作成する
// アンチエイリアスの影響を除去し、透過画像を返す
func CreateTextSpriteImage(opts TextSpriteOptions) *image.RGBA {
	if opts.Face == nil || opts.Text == "" {
		return nil
	}

	// サイズの自動計算
	width := opts.Width
	height := opts.Height
	if width == 0 || height == 0 {
		bounds := measureText(opts.Face, opts.Text)
		if width == 0 {
			width = bounds.Dx() + opts.X + 10 // 余白を追加
		}
		if height == 0 {
			height = bounds.Dy() + 10 // 余白を追加
		}
	}

	if width <= 0 || height <= 0 {
		return nil
	}

	// 1. 背景色で塗りつぶした画像を作成
	bgImg := image.NewRGBA(image.Rect(0, 0, width, height))
	bgColor := opts.BgColor
	if bgColor == nil {
		bgColor = color.White
	}
	draw.Draw(bgImg, bgImg.Bounds(), image.NewUniform(bgColor), image.Point{}, draw.Src)

	// 2. 背景のコピーを保持
	bgCopy := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(bgCopy, bgCopy.Bounds(), bgImg, image.Point{}, draw.Src)

	// 3. テキストを描画
	textColor := opts.TextColor
	if textColor == nil {
		textColor = color.Black
	}
	drawer := &font.Drawer{
		Dst:  bgImg,
		Src:  image.NewUniform(textColor),
		Face: opts.Face,
		Dot:  fixed.Point26_6{X: fixed.I(opts.X), Y: fixed.I(opts.Y)},
	}
	drawer.DrawString(opts.Text)

	// 4. 差分を抽出（背景と異なるピクセルのみを残す）
	result := image.NewRGBA(image.Rect(0, 0, width, height))
	extractDifference(bgCopy, bgImg, result)

	return result
}

// extractDifference は2つの画像の差分を抽出し、結果画像に書き込む
// 背景と同じピクセルは透明に、異なるピクセルはそのまま残す
func extractDifference(bg, text, result *image.RGBA) {
	bounds := bg.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			bgPixel := bg.At(x, y)
			textPixel := text.At(x, y)

			bgR, bgG, bgB, _ := bgPixel.RGBA()
			txR, txG, txB, _ := textPixel.RGBA()

			if bgR != txR || bgG != txG || bgB != txB {
				// 差分があるピクセルはそのまま残す
				result.Set(x, y, textPixel)
			} else {
				// 背景と同じピクセルは透明にする
				result.Set(x, y, color.RGBA{0, 0, 0, 0})
			}
		}
	}
}

// measureText はテキストの境界ボックスを計算する
func measureText(face font.Face, text string) image.Rectangle {
	bounds, _ := font.BoundString(face, text)
	return image.Rect(
		bounds.Min.X.Floor(),
		bounds.Min.Y.Floor(),
		bounds.Max.X.Ceil(),
		bounds.Max.Y.Ceil(),
	)
}

// CreateTextSprite はテキストスプライトを作成してSpriteManagerに登録する
func (sm *SpriteManager) CreateTextSprite(opts TextSpriteOptions) *Sprite {
	img := CreateTextSpriteImage(opts)
	if img == nil {
		return nil
	}
	return sm.CreateSprite(ebiten.NewImageFromImage(img))
}
