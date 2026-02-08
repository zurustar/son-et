package graphics

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// DrawLine は指定されたピクチャーに直線を描画する
// 要件 6.1
// スプライトシステム要件 9.1: 線を描画したスプライトを作成できる
//
// パラメータ:
//   - picID: 描画先ピクチャーID
//   - x1, y1: 始点座標
//   - x2, y2: 終点座標
func (gs *GraphicsSystem) drawLineInternal(picID, x1, y1, x2, y2 int) error {
	// ピクチャーを取得
	pic, err := gs.pictures.GetPicWithoutLock(picID)
	if err != nil {
		gs.log.Error("DrawLine: picture not found",
			"picID", picID,
			"error", err)
		return fmt.Errorf("picture not found: %d", picID)
	}

	// 線を描画
	vector.StrokeLine(
		pic.Image,
		float32(x1), float32(y1),
		float32(x2), float32(y2),
		float32(gs.lineSize),
		gs.paintColor,
		false, // アンチエイリアスなし
	)

	// 親スプライトを取得（TextWriteと同様にウインドウ内のスプライトとして管理）
	var parentSprite *Sprite
	if gs.pictureSpriteManager != nil {
		ps := gs.pictureSpriteManager.GetPictureSpriteByPictureID(picID)
		if ps != nil {
			parentSprite = ps.GetSprite()
		} else {
			// フォールバック: 従来の方法
			parentSprite = gs.pictureSpriteManager.GetBackgroundPictureSpriteSprite(picID)
		}
	}

	// スプライトシステム要件 9.1: ShapeSpriteを作成する
	if gs.shapeSpriteManager != nil {
		var ss *ShapeSprite
		if parentSprite != nil {
			ss = gs.shapeSpriteManager.CreateLineSpriteWithParent(
				picID,
				x1, y1, x2, y2,
				gs.paintColor,
				gs.lineSize,
				0, // Z順序はスプライトシステムで自動管理
				parentSprite,
			)
			if ss != nil {
				parentSprite.AddChild(ss.GetSprite())
			}
		} else {
			ss = gs.shapeSpriteManager.CreateLineSprite(
				picID,
				x1, y1, x2, y2,
				gs.paintColor,
				gs.lineSize,
				0, // Z順序はスプライトシステムで自動管理
			)
		}
		gs.log.Debug("DrawLine: created ShapeSprite", "picID", picID, "hasParent", parentSprite != nil)
	}

	gs.log.Debug("DrawLine: drew line",
		"picID", picID,
		"x1", x1, "y1", y1,
		"x2", x2, "y2", y2,
		"lineSize", gs.lineSize)

	return nil
}

// DrawLineOnPic は指定されたピクチャーに直線を描画する（公開API）
func (gs *GraphicsSystem) DrawLineOnPic(picID, x1, y1, x2, y2 int) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	return gs.drawLineInternal(picID, x1, y1, x2, y2)
}

// DrawRect は指定されたピクチャーに矩形を描画する
// 要件 6.5
// スプライトシステム要件 9.2, 9.3: 矩形を描画したスプライトを作成できる
//
// パラメータ:
//   - picID: 描画先ピクチャーID
//   - x1, y1: 左上座標
//   - x2, y2: 右下座標
//   - fillMode: 0=塗りつぶし, 1=輪郭のみ（サンプルの動作に基づく）
func (gs *GraphicsSystem) drawRectInternal(picID, x1, y1, x2, y2, fillMode int) error {
	// ピクチャーを取得
	pic, err := gs.pictures.GetPicWithoutLock(picID)
	if err != nil {
		gs.log.Error("DrawRect: picture not found",
			"picID", picID,
			"error", err)
		return fmt.Errorf("picture not found: %d", picID)
	}

	// 座標を正規化（x1 < x2, y1 < y2 を保証）
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}

	width := float32(x2 - x1)
	height := float32(y2 - y1)

	// 親スプライトを取得（TextWriteと同様にウインドウ内のスプライトとして管理）
	var parentSprite *Sprite
	if gs.pictureSpriteManager != nil {
		ps := gs.pictureSpriteManager.GetPictureSpriteByPictureID(picID)
		if ps != nil {
			parentSprite = ps.GetSprite()
		} else {
			// フォールバック: 従来の方法
			parentSprite = gs.pictureSpriteManager.GetBackgroundPictureSpriteSprite(picID)
		}
	}

	if fillMode == 1 {
		// fillMode=1: 輪郭のみ（4本の線で描画）
		vector.StrokeRect(
			pic.Image,
			float32(x1), float32(y1),
			width, height,
			float32(gs.lineSize),
			gs.paintColor,
			false, // アンチエイリアスなし
		)

		// スプライトシステム要件 9.2: 矩形のShapeSpriteを作成する
		if gs.shapeSpriteManager != nil {
			var ss *ShapeSprite
			if parentSprite != nil {
				ss = gs.shapeSpriteManager.CreateRectSpriteWithParent(
					picID,
					x1, y1, x2, y2,
					gs.paintColor,
					gs.lineSize,
					0, // Z順序はスプライトシステムで自動管理
					parentSprite,
				)
				if ss != nil {
					parentSprite.AddChild(ss.GetSprite())
				}
			} else {
				ss = gs.shapeSpriteManager.CreateRectSprite(
					picID,
					x1, y1, x2, y2,
					gs.paintColor,
					gs.lineSize,
					0, // Z順序はスプライトシステムで自動管理
				)
			}
			gs.log.Debug("DrawRect: created RectSprite", "picID", picID, "hasParent", parentSprite != nil)
		}
	} else {
		// fillMode=0（デフォルト）: 塗りつぶし
		vector.FillRect(
			pic.Image,
			float32(x1), float32(y1),
			width, height,
			gs.paintColor,
			false, // アンチエイリアスなし
		)

		// スプライトシステム要件 9.3: 塗りつぶし矩形のShapeSpriteを作成する
		if gs.shapeSpriteManager != nil {
			var ss *ShapeSprite
			if parentSprite != nil {
				ss = gs.shapeSpriteManager.CreateFillRectSpriteWithParent(
					picID,
					x1, y1, x2, y2,
					gs.paintColor,
					0, // Z順序はスプライトシステムで自動管理
					parentSprite,
				)
				if ss != nil {
					parentSprite.AddChild(ss.GetSprite())
				}
			} else {
				ss = gs.shapeSpriteManager.CreateFillRectSprite(
					picID,
					x1, y1, x2, y2,
					gs.paintColor,
					0, // Z順序はスプライトシステムで自動管理
				)
			}
			gs.log.Debug("DrawRect: created FillRectSprite", "picID", picID, "hasParent", parentSprite != nil)
		}
	}

	gs.log.Debug("DrawRect: drew rectangle",
		"picID", picID,
		"x1", x1, "y1", y1,
		"x2", x2, "y2", y2,
		"fillMode", fillMode)

	return nil
}

// DrawRectOnPic は指定されたピクチャーに矩形を描画する（公開API）
func (gs *GraphicsSystem) DrawRectOnPic(picID, x1, y1, x2, y2, fillMode int) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	return gs.drawRectInternal(picID, x1, y1, x2, y2, fillMode)
}

// FillRectOnPic は指定されたピクチャーに指定色で矩形を塗りつぶす
// 要件 6.6
// スプライトシステム要件 9.3: 塗りつぶし矩形を描画したスプライトを作成できる
//
// パラメータ:
//   - picID: 描画先ピクチャーID
//   - x1, y1: 左上座標
//   - x2, y2: 右下座標
//   - c: 塗りつぶし色
func (gs *GraphicsSystem) fillRectInternal(picID, x1, y1, x2, y2 int, c color.Color) error {
	// ピクチャーを取得
	pic, err := gs.pictures.GetPicWithoutLock(picID)
	if err != nil {
		gs.log.Error("FillRect: picture not found",
			"picID", picID,
			"error", err)
		return fmt.Errorf("picture not found: %d", picID)
	}

	// 座標を正規化（x1 < x2, y1 < y2 を保証）
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}

	width := float32(x2 - x1)
	height := float32(y2 - y1)

	// 塗りつぶし
	vector.FillRect(
		pic.Image,
		float32(x1), float32(y1),
		width, height,
		c,
		false, // アンチエイリアスなし
	)

	// 親スプライトを取得（TextWriteと同様にウインドウ内のスプライトとして管理）
	var parentSprite *Sprite
	if gs.pictureSpriteManager != nil {
		ps := gs.pictureSpriteManager.GetPictureSpriteByPictureID(picID)
		if ps != nil {
			parentSprite = ps.GetSprite()
		} else {
			// フォールバック: 従来の方法
			parentSprite = gs.pictureSpriteManager.GetBackgroundPictureSpriteSprite(picID)
		}
	}

	// スプライトシステム要件 9.3: 塗りつぶし矩形のShapeSpriteを作成する
	if gs.shapeSpriteManager != nil {
		var ss *ShapeSprite
		if parentSprite != nil {
			ss = gs.shapeSpriteManager.CreateFillRectSpriteWithParent(
				picID,
				x1, y1, x2, y2,
				c,
				0, // Z順序はスプライトシステムで自動管理
				parentSprite,
			)
			if ss != nil {
				parentSprite.AddChild(ss.GetSprite())
			}
		} else {
			ss = gs.shapeSpriteManager.CreateFillRectSprite(
				picID,
				x1, y1, x2, y2,
				c,
				0, // Z順序はスプライトシステムで自動管理
			)
		}
		gs.log.Debug("FillRect: created FillRectSprite", "picID", picID, "hasParent", parentSprite != nil)
	}

	gs.log.Debug("FillRect: filled rectangle",
		"picID", picID,
		"x1", x1, "y1", y1,
		"x2", x2, "y2", y2)

	return nil
}

// FillRectOnPic は指定されたピクチャーに指定色で矩形を塗りつぶす（公開API）
func (gs *GraphicsSystem) FillRectOnPic(picID, x1, y1, x2, y2 int, c color.Color) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	return gs.fillRectInternal(picID, x1, y1, x2, y2, c)
}

// DrawCircle は指定されたピクチャーに円を描画する
// 要件 6.2, 6.3, 6.4
// スプライトシステム要件 9: 円を描画したスプライトを作成できる
//
// パラメータ:
//   - picID: 描画先ピクチャーID
//   - x, y: 中心座標
//   - radius: 半径
//   - fillMode: 0=輪郭のみ, 2=塗りつぶし
func (gs *GraphicsSystem) drawCircleInternal(picID, x, y, radius, fillMode int) error {
	// ピクチャーを取得
	pic, err := gs.pictures.GetPicWithoutLock(picID)
	if err != nil {
		gs.log.Error("DrawCircle: picture not found",
			"picID", picID,
			"error", err)
		return fmt.Errorf("picture not found: %d", picID)
	}

	if radius <= 0 {
		gs.log.Debug("DrawCircle: invalid radius", "radius", radius)
		return nil
	}

	// 親スプライトを取得（TextWriteと同様にウインドウ内のスプライトとして管理）
	var parentSprite *Sprite
	if gs.pictureSpriteManager != nil {
		ps := gs.pictureSpriteManager.GetPictureSpriteByPictureID(picID)
		if ps != nil {
			parentSprite = ps.GetSprite()
		} else {
			// フォールバック: 従来の方法
			parentSprite = gs.pictureSpriteManager.GetBackgroundPictureSpriteSprite(picID)
		}
	}

	if fillMode == 2 {
		// 塗りつぶし円
		vector.DrawFilledCircle(
			pic.Image,
			float32(x), float32(y),
			float32(radius),
			gs.paintColor,
			false, // アンチエイリアスなし
		)

		// スプライトシステム要件 9: 塗りつぶし円のShapeSpriteを作成する
		if gs.shapeSpriteManager != nil {
			var ss *ShapeSprite
			if parentSprite != nil {
				ss = gs.shapeSpriteManager.CreateFillCircleSpriteWithParent(
					picID,
					x, y, radius,
					gs.paintColor,
					0, // Z順序はスプライトシステムで自動管理
					parentSprite,
				)
				if ss != nil {
					parentSprite.AddChild(ss.GetSprite())
				}
			} else {
				ss = gs.shapeSpriteManager.CreateFillCircleSprite(
					picID,
					x, y, radius,
					gs.paintColor,
					0, // Z順序はスプライトシステムで自動管理
				)
			}
			gs.log.Debug("DrawCircle: created FillCircleSprite", "picID", picID, "hasParent", parentSprite != nil)
		}
	} else {
		// 輪郭のみ
		vector.StrokeCircle(
			pic.Image,
			float32(x), float32(y),
			float32(radius),
			float32(gs.lineSize),
			gs.paintColor,
			false, // アンチエイリアスなし
		)

		// スプライトシステム要件 9: 円のShapeSpriteを作成する
		if gs.shapeSpriteManager != nil {
			var ss *ShapeSprite
			if parentSprite != nil {
				ss = gs.shapeSpriteManager.CreateCircleSpriteWithParent(
					picID,
					x, y, radius,
					gs.paintColor,
					gs.lineSize,
					0, // Z順序はスプライトシステムで自動管理
					parentSprite,
				)
				if ss != nil {
					parentSprite.AddChild(ss.GetSprite())
				}
			} else {
				ss = gs.shapeSpriteManager.CreateCircleSprite(
					picID,
					x, y, radius,
					gs.paintColor,
					gs.lineSize,
					0, // Z順序はスプライトシステムで自動管理
				)
			}
			gs.log.Debug("DrawCircle: created CircleSprite", "picID", picID, "hasParent", parentSprite != nil)
		}
	}

	gs.log.Debug("DrawCircle: drew circle",
		"picID", picID,
		"x", x, "y", y,
		"radius", radius,
		"fillMode", fillMode)

	return nil
}

// DrawCircleOnPic は指定されたピクチャーに円を描画する（公開API）
func (gs *GraphicsSystem) DrawCircleOnPic(picID, x, y, radius, fillMode int) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	return gs.drawCircleInternal(picID, x, y, radius, fillMode)
}

// SetLineSizeValue は線の太さを設定する
// 要件 6.7
func (gs *GraphicsSystem) SetLineSizeValue(size int) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	if size < 1 {
		size = 1
	}
	gs.lineSize = size

	gs.log.Debug("SetLineSize: set line size", "size", size)
}

// SetPaintColorValue は描画色を設定する
// 要件 6.8
func (gs *GraphicsSystem) SetPaintColorValue(c color.Color) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	gs.paintColor = c

	gs.log.Debug("SetPaintColor: set paint color")
}

// GetColorAt は指定座標のピクセル色を取得する
// 要件 6.9
//
// パラメータ:
//   - picID: ピクチャーID
//   - x, y: 座標
//
// 戻り値:
//   - 色（0xRRGGBB形式）
//   - エラー
func (gs *GraphicsSystem) GetColorAt(picID, x, y int) (int, error) {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	// ピクチャーを取得
	pic, err := gs.pictures.GetPicWithoutLock(picID)
	if err != nil {
		gs.log.Error("GetColor: picture not found",
			"picID", picID,
			"error", err)
		return 0, fmt.Errorf("picture not found: %d", picID)
	}

	// 座標が範囲内かチェック
	if x < 0 || x >= pic.Width || y < 0 || y >= pic.Height {
		gs.log.Warn("GetColor: coordinates out of bounds",
			"picID", picID,
			"x", x, "y", y,
			"width", pic.Width, "height", pic.Height)
		return 0, nil
	}

	// ピクセル色を取得
	// OriginalImageから取得することで、Ebitengineのゲームループ開始前でも動作する
	// 要件 6.9: GetColorは指定された座標のピクセル色を返す
	var c color.Color
	if pic.OriginalImage != nil {
		c = pic.OriginalImage.At(x, y)
	} else {
		// OriginalImageがない場合はEbitengine Imageから取得
		// （ゲームループ開始後のみ動作）
		c = pic.Image.At(x, y)
	}
	colorInt := ColorToInt(c)

	gs.log.Debug("GetColor: got color",
		"picID", picID,
		"x", x, "y", y,
		"color", fmt.Sprintf("0x%06X", colorInt))

	return colorInt, nil
}

// GetLineSize は現在の線の太さを返す
func (gs *GraphicsSystem) GetLineSize() int {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	return gs.lineSize
}

// GetPaintColor は現在の描画色を返す
func (gs *GraphicsSystem) GetPaintColor() color.Color {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	return gs.paintColor
}

// drawEllipse は楕円を描画するヘルパー関数（将来の拡張用）
func drawEllipse(dst *ebiten.Image, cx, cy, rx, ry float32, c color.Color, filled bool) {
	if filled {
		// 塗りつぶし楕円
		// 楕円を多角形として近似
		segments := int(math.Max(float64(rx), float64(ry)) * 2)
		if segments < 16 {
			segments = 16
		}
		if segments > 360 {
			segments = 360
		}

		var path vector.Path
		for i := 0; i <= segments; i++ {
			angle := float64(i) * 2 * math.Pi / float64(segments)
			x := cx + rx*float32(math.Cos(angle))
			y := cy + ry*float32(math.Sin(angle))
			if i == 0 {
				path.MoveTo(x, y)
			} else {
				path.LineTo(x, y)
			}
		}
		path.Close()

		vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)
		r, g, b, a := c.RGBA()
		for i := range vs {
			vs[i].ColorR = float32(r) / 65535.0
			vs[i].ColorG = float32(g) / 65535.0
			vs[i].ColorB = float32(b) / 65535.0
			vs[i].ColorA = float32(a) / 65535.0
		}
		dst.DrawTriangles(vs, is, emptyImage, nil)
	} else {
		// 輪郭のみ
		segments := int(math.Max(float64(rx), float64(ry)) * 2)
		if segments < 16 {
			segments = 16
		}
		if segments > 360 {
			segments = 360
		}

		var path vector.Path
		for i := 0; i <= segments; i++ {
			angle := float64(i) * 2 * math.Pi / float64(segments)
			x := cx + rx*float32(math.Cos(angle))
			y := cy + ry*float32(math.Sin(angle))
			if i == 0 {
				path.MoveTo(x, y)
			} else {
				path.LineTo(x, y)
			}
		}
		path.Close()

		op := &vector.StrokeOptions{
			Width: 1,
		}
		vs, is := path.AppendVerticesAndIndicesForStroke(nil, nil, op)
		r, g, b, a := c.RGBA()
		for i := range vs {
			vs[i].ColorR = float32(r) / 65535.0
			vs[i].ColorG = float32(g) / 65535.0
			vs[i].ColorB = float32(b) / 65535.0
			vs[i].ColorA = float32(a) / 65535.0
		}
		dst.DrawTriangles(vs, is, emptyImage, nil)
	}
}

// emptyImage は描画用の空の画像（シェーダー用）
var emptyImage = func() *ebiten.Image {
	img := ebiten.NewImage(1, 1)
	img.Fill(color.White)
	return img
}()
