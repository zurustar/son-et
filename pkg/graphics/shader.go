package graphics

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// drawImageWithColorKey は透明色を指定して画像を描画する
// transColor: 透明にする色
// 前の実装（_old_implementation2）と同じアプローチを使用：
// 1. image.NewRGBAで新しい画像を作成
// 2. ピクセル単位で透明色を処理
// 3. ebiten.NewImageFromImageで変換
func drawImageWithColorKey(dst, src *ebiten.Image, x, y int, transColor color.Color) error {
	bounds := src.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	if w <= 0 || h <= 0 {
		return nil
	}

	// 透明色のRGBA値を取得（8bit）
	tr, tg, tb, _ := transColor.RGBA()
	tr8 := uint8(tr >> 8)
	tg8 := uint8(tg >> 8)
	tb8 := uint8(tb >> 8)

	// 新しいRGBA画像を作成（前の実装と同じアプローチ）
	processedImg := image.NewRGBA(image.Rect(0, 0, w, h))

	// ピクセル単位で透明色を処理
	for sy := 0; sy < h; sy++ {
		for sx := 0; sx < w; sx++ {
			// ソース画像からピクセルを取得
			// bounds.Min を考慮してSubImageの場合も正しく動作するようにする
			c := src.At(bounds.Min.X+sx, bounds.Min.Y+sy)
			r, g, b, a := c.RGBA()

			// 16bitから8bitに変換
			r8 := uint8(r >> 8)
			g8 := uint8(g >> 8)
			b8 := uint8(b >> 8)

			// 透明色と一致する場合、完全に透明にする
			if r8 == tr8 && g8 == tg8 && b8 == tb8 {
				processedImg.Set(sx, sy, color.RGBA{0, 0, 0, 0})
			} else {
				// 元の色を保持
				processedImg.Set(sx, sy, color.RGBA{r8, g8, b8, uint8(a >> 8)})
			}
		}
	}

	// Ebiten画像に変換して描画
	tmpImg := ebiten.NewImageFromImage(processedImg)

	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(x), float64(y))
	dst.DrawImage(tmpImg, opts)

	return nil
}
