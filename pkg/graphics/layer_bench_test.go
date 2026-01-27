package graphics

import (
	"image"
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// ============================================================================
// LayerManager ベースのベンチマークテスト
// 要件 9.1: レイヤー合成が60fps（16.67ms以内）で完了する
// 要件 9.2: テキストレイヤーの作成が効率的に行われる
// ============================================================================

// BenchmarkLayerManagerComposite はLayerManagerを使用したレイヤー合成のパフォーマンスを測定
// 要件 9.1: 画像レイヤー100枚の合成を16.67ms（60fps）以内に完了する
func BenchmarkLayerManagerComposite10(b *testing.B) {
	benchmarkLayerManagerComposite(b, 10)
}

func BenchmarkLayerManagerComposite50(b *testing.B) {
	benchmarkLayerManagerComposite(b, 50)
}

func BenchmarkLayerManagerComposite100(b *testing.B) {
	benchmarkLayerManagerComposite(b, 100)
}

func benchmarkLayerManagerComposite(b *testing.B, layerCount int) {
	// LayerManagerを作成
	lm := NewLayerManager()

	// 640x480のピクチャーを作成
	baseWidth, baseHeight := 640, 480
	picID := 0

	// PictureLayerSetを取得
	pls := lm.GetOrCreatePictureLayerSet(picID)

	// 背景レイヤーを作成
	bgImg := ebiten.NewImage(baseWidth, baseHeight)
	bgImg.Fill(color.RGBA{255, 255, 255, 255})
	bgLayer := NewBackgroundLayer(lm.GetNextLayerID(), picID, bgImg)
	pls.SetBackground(bgLayer)

	// 描画レイヤーを作成
	drawLayer := NewDrawingLayer(lm.GetNextLayerID(), picID, baseWidth, baseHeight)
	pls.SetDrawing(drawLayer)

	// キャストレイヤーを作成（100x100のサイズ）
	layerWidth, layerHeight := 100, 100
	for i := 0; i < layerCount; i++ {
		castImg := ebiten.NewImage(layerWidth, layerHeight)
		r := uint8((i * 17) % 256)
		g := uint8((i * 31) % 256)
		bl := uint8((i * 47) % 256)
		castImg.Fill(color.RGBA{r, g, bl, 128})

		x := (i * 5) % (baseWidth - layerWidth)
		y := (i * 7) % (baseHeight - layerHeight)

		castLayer := NewCastLayer(
			lm.GetNextLayerID(),
			i, // castID
			picID,
			0, // srcPicID
			x, y,
			0, 0, // srcX, srcY
			layerWidth, layerHeight,
			pls.GetNextCastZOffset(),
		)
		castLayer.SetCachedImage(castImg)
		pls.AddCastLayer(castLayer)
	}

	// 可視領域
	visibleRect := image.Rect(0, 0, baseWidth, baseHeight)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 毎回ダーティフラグを設定して合成を強制
		pls.MarkFullDirty()
		_ = pls.Composite(visibleRect)
	}
}

// BenchmarkTextLayerEntryCreate はTextLayerEntryの作成パフォーマンスを測定
// 要件 9.2: テキストレイヤーの作成が効率的に行われる
func BenchmarkTextLayerEntryCreate1(b *testing.B) {
	benchmarkTextLayerEntryCreate(b, 1)
}

func BenchmarkTextLayerEntryCreate10(b *testing.B) {
	benchmarkTextLayerEntryCreate(b, 10)
}

func BenchmarkTextLayerEntryCreate50(b *testing.B) {
	benchmarkTextLayerEntryCreate(b, 50)
}

func BenchmarkTextLayerEntryCreate100(b *testing.B) {
	benchmarkTextLayerEntryCreate(b, 100)
}

func benchmarkTextLayerEntryCreate(b *testing.B, count int) {
	baseWidth, baseHeight := 640, 480
	picID := 0

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		lm := NewLayerManager()
		pls := lm.GetOrCreatePictureLayerSet(picID)

		for j := 0; j < count; j++ {
			x := (j * 20) % (baseWidth - 100)
			y := (j * 15) % (baseHeight - 20)

			entry := NewTextLayerEntry(
				lm.GetNextLayerID(),
				picID,
				x, y,
				"テスト文字列ABC123",
				pls.GetNextTextZOffset(),
			)

			// 画像を設定（実際のテキストレンダリングをシミュレート）
			textImg := ebiten.NewImage(100, 20)
			textImg.Fill(color.RGBA{0, 0, 0, 255})
			entry.SetImage(textImg)

			pls.AddTextLayer(entry)
		}
	}
}

// BenchmarkDirtyRegionTracking はダーティ領域追跡のパフォーマンスを測定
// 要件 6.1, 6.3: ダーティ領域の追跡と統合
func BenchmarkDirtyRegionTracking10(b *testing.B) {
	benchmarkDirtyRegionTracking(b, 10)
}

func BenchmarkDirtyRegionTracking50(b *testing.B) {
	benchmarkDirtyRegionTracking(b, 50)
}

func BenchmarkDirtyRegionTracking100(b *testing.B) {
	benchmarkDirtyRegionTracking(b, 100)
}

func benchmarkDirtyRegionTracking(b *testing.B, updateCount int) {
	picID := 0
	baseWidth, baseHeight := 640, 480

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		pls := NewPictureLayerSet(picID)

		// 複数のダーティ領域を追加
		for j := 0; j < updateCount; j++ {
			x := (j * 13) % (baseWidth - 100)
			y := (j * 17) % (baseHeight - 100)
			rect := image.Rect(x, y, x+100, y+100)
			pls.AddDirtyRegion(rect)
		}
	}
}

// BenchmarkPartialVsFullComposite は部分更新と全体更新のパフォーマンス比較
// 要件 6.2: ダーティ領域のみを再合成する
func BenchmarkPartialComposite(b *testing.B) {
	lm := NewLayerManager()
	baseWidth, baseHeight := 640, 480
	picID := 0

	pls := lm.GetOrCreatePictureLayerSet(picID)

	// 背景レイヤーを作成
	bgImg := ebiten.NewImage(baseWidth, baseHeight)
	bgImg.Fill(color.RGBA{255, 255, 255, 255})
	bgLayer := NewBackgroundLayer(lm.GetNextLayerID(), picID, bgImg)
	pls.SetBackground(bgLayer)

	// 50個のキャストレイヤーを作成
	layerWidth, layerHeight := 100, 100
	for i := 0; i < 50; i++ {
		castImg := ebiten.NewImage(layerWidth, layerHeight)
		castImg.Fill(color.RGBA{uint8(i * 5), uint8(i * 3), uint8(i * 7), 255})

		x := (i * 5) % (baseWidth - layerWidth)
		y := (i * 7) % (baseHeight - layerHeight)

		castLayer := NewCastLayer(
			lm.GetNextLayerID(),
			i, picID, 0,
			x, y, 0, 0,
			layerWidth, layerHeight,
			pls.GetNextCastZOffset(),
		)
		castLayer.SetCachedImage(castImg)
		pls.AddCastLayer(castLayer)
	}

	visibleRect := image.Rect(0, 0, baseWidth, baseHeight)

	// 初回合成
	pls.Composite(visibleRect)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 1つのレイヤーのみダーティにする（部分更新）
		if len(pls.Casts) > 0 {
			pls.Casts[0].SetDirty(true)
			pls.AddDirtyRegion(pls.Casts[0].GetBounds())
		}
		_ = pls.Composite(visibleRect)
	}
}

func BenchmarkFullComposite(b *testing.B) {
	lm := NewLayerManager()
	baseWidth, baseHeight := 640, 480
	picID := 0

	pls := lm.GetOrCreatePictureLayerSet(picID)

	// 背景レイヤーを作成
	bgImg := ebiten.NewImage(baseWidth, baseHeight)
	bgImg.Fill(color.RGBA{255, 255, 255, 255})
	bgLayer := NewBackgroundLayer(lm.GetNextLayerID(), picID, bgImg)
	pls.SetBackground(bgLayer)

	// 50個のキャストレイヤーを作成
	layerWidth, layerHeight := 100, 100
	for i := 0; i < 50; i++ {
		castImg := ebiten.NewImage(layerWidth, layerHeight)
		castImg.Fill(color.RGBA{uint8(i * 5), uint8(i * 3), uint8(i * 7), 255})

		x := (i * 5) % (baseWidth - layerWidth)
		y := (i * 7) % (baseHeight - layerHeight)

		castLayer := NewCastLayer(
			lm.GetNextLayerID(),
			i, picID, 0,
			x, y, 0, 0,
			layerWidth, layerHeight,
			pls.GetNextCastZOffset(),
		)
		castLayer.SetCachedImage(castImg)
		pls.AddCastLayer(castLayer)
	}

	visibleRect := image.Rect(0, 0, baseWidth, baseHeight)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 全体をダーティにする（全体更新）
		pls.MarkFullDirty()
		_ = pls.Composite(visibleRect)
	}
}

// BenchmarkCastLayerMovement はキャストレイヤーの移動パフォーマンスを測定
// 要件 2.2: MoveCastが呼び出されたときに対応するCast_Layerの位置を更新する
func BenchmarkCastLayerMovement(b *testing.B) {
	lm := NewLayerManager()
	baseWidth, baseHeight := 640, 480
	picID := 0

	pls := lm.GetOrCreatePictureLayerSet(picID)

	// キャストレイヤーを作成
	layerWidth, layerHeight := 100, 100
	castImg := ebiten.NewImage(layerWidth, layerHeight)
	castImg.Fill(color.RGBA{255, 0, 0, 255})

	castLayer := NewCastLayer(
		lm.GetNextLayerID(),
		0, picID, 0,
		0, 0, 0, 0,
		layerWidth, layerHeight,
		pls.GetNextCastZOffset(),
	)
	castLayer.SetCachedImage(castImg)
	pls.AddCastLayer(castLayer)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 位置を更新
		x := (i * 5) % (baseWidth - layerWidth)
		y := (i * 7) % (baseHeight - layerHeight)
		castLayer.SetPosition(x, y)
	}
}

// BenchmarkVisibilityClipping は可視領域クリッピングのパフォーマンスを測定
// 要件 4.1, 4.4: 可視領域クリッピング
func BenchmarkVisibilityClipping(b *testing.B) {
	lm := NewLayerManager()
	picID := 0

	pls := lm.GetOrCreatePictureLayerSet(picID)

	// 100個のキャストレイヤーを作成（一部は可視領域外）
	layerWidth, layerHeight := 100, 100
	for i := 0; i < 100; i++ {
		castImg := ebiten.NewImage(layerWidth, layerHeight)
		castImg.Fill(color.RGBA{uint8(i * 5), uint8(i * 3), uint8(i * 7), 255})

		// 一部のレイヤーを可視領域外に配置
		x := (i * 10) - 50 // -50 から 950
		y := (i * 8) - 40  // -40 から 760

		castLayer := NewCastLayer(
			lm.GetNextLayerID(),
			i, picID, 0,
			x, y, 0, 0,
			layerWidth, layerHeight,
			pls.GetNextCastZOffset(),
		)
		castLayer.SetCachedImage(castImg)
		pls.AddCastLayer(castLayer)
	}

	// 可視領域（画面の一部のみ）
	visibleRect := image.Rect(100, 100, 500, 400)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		pls.MarkFullDirty()
		_ = pls.Composite(visibleRect)
	}
}

// ============================================================================
// 以下は既存のベンチマークテスト（直接描画方式）
// ============================================================================

// BenchmarkLayerComposite はレイヤー合成のパフォーマンスを測定する
// 100枚のレイヤーを重ねた場合のコストを確認

func BenchmarkLayerComposite10(b *testing.B) {
	benchmarkLayerComposite(b, 10)
}

func BenchmarkLayerComposite50(b *testing.B) {
	benchmarkLayerComposite(b, 50)
}

func BenchmarkLayerComposite100(b *testing.B) {
	benchmarkLayerComposite(b, 100)
}

func BenchmarkLayerComposite200(b *testing.B) {
	benchmarkLayerComposite(b, 200)
}

func benchmarkLayerComposite(b *testing.B, layerCount int) {
	// 640x480のベース画像を作成
	baseWidth, baseHeight := 640, 480
	baseImg := ebiten.NewImage(baseWidth, baseHeight)
	baseImg.Fill(color.RGBA{255, 255, 255, 255})

	// レイヤー画像を作成（100x100のサイズ）
	layerWidth, layerHeight := 100, 100
	layers := make([]*ebiten.Image, layerCount)
	for i := 0; i < layerCount; i++ {
		layers[i] = ebiten.NewImage(layerWidth, layerHeight)
		// 各レイヤーに異なる色を設定
		r := uint8((i * 17) % 256)
		g := uint8((i * 31) % 256)
		b := uint8((i * 47) % 256)
		layers[i].Fill(color.RGBA{r, g, b, 128}) // 半透明
	}

	// 描画先
	dst := ebiten.NewImage(baseWidth, baseHeight)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// ベース画像をコピー
		dst.Clear()
		dst.DrawImage(baseImg, nil)

		// 各レイヤーを合成
		for j, layer := range layers {
			opts := &ebiten.DrawImageOptions{}
			// レイヤーを散らばらせて配置
			x := (j * 5) % (baseWidth - layerWidth)
			y := (j * 7) % (baseHeight - layerHeight)
			opts.GeoM.Translate(float64(x), float64(y))
			dst.DrawImage(layer, opts)
		}
	}
}

// BenchmarkLayerCompositeWithTransparency は透明色処理付きのレイヤー合成を測定
func BenchmarkLayerCompositeWithTransparency10(b *testing.B) {
	benchmarkLayerCompositeWithTransparency(b, 10)
}

func BenchmarkLayerCompositeWithTransparency50(b *testing.B) {
	benchmarkLayerCompositeWithTransparency(b, 50)
}

func BenchmarkLayerCompositeWithTransparency100(b *testing.B) {
	benchmarkLayerCompositeWithTransparency(b, 100)
}

func benchmarkLayerCompositeWithTransparency(b *testing.B, layerCount int) {
	// 640x480のベース画像を作成
	baseWidth, baseHeight := 640, 480
	baseImg := ebiten.NewImage(baseWidth, baseHeight)
	baseImg.Fill(color.RGBA{255, 255, 255, 255})

	// レイヤー画像を作成（100x100のサイズ、透明色処理済み）
	layerWidth, layerHeight := 100, 100
	layers := make([]*ebiten.Image, layerCount)
	for i := 0; i < layerCount; i++ {
		// RGBAイメージを作成して透明色処理
		rgba := image.NewRGBA(image.Rect(0, 0, layerWidth, layerHeight))
		r := uint8((i * 17) % 256)
		g := uint8((i * 31) % 256)
		bl := uint8((i * 47) % 256)
		transColor := color.RGBA{255, 255, 255, 255} // 白を透明色とする

		for y := 0; y < layerHeight; y++ {
			for x := 0; x < layerWidth; x++ {
				// 市松模様で透明色と通常色を交互に
				if (x+y)%2 == 0 {
					rgba.Set(x, y, transColor)
				} else {
					rgba.Set(x, y, color.RGBA{r, g, bl, 255})
				}
			}
		}

		// 透明色を実際の透明に変換
		for y := 0; y < layerHeight; y++ {
			for x := 0; x < layerWidth; x++ {
				c := rgba.At(x, y)
				cr, cg, cb, _ := c.RGBA()
				tr, tg, tb, _ := transColor.RGBA()
				if cr == tr && cg == tg && cb == tb {
					rgba.Set(x, y, color.RGBA{0, 0, 0, 0})
				}
			}
		}

		layers[i] = ebiten.NewImageFromImage(rgba)
	}

	// 描画先
	dst := ebiten.NewImage(baseWidth, baseHeight)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// ベース画像をコピー
		dst.Clear()
		dst.DrawImage(baseImg, nil)

		// 各レイヤーを合成
		for j, layer := range layers {
			opts := &ebiten.DrawImageOptions{}
			x := (j * 5) % (baseWidth - layerWidth)
			y := (j * 7) % (baseHeight - layerHeight)
			opts.GeoM.Translate(float64(x), float64(y))
			dst.DrawImage(layer, opts)
		}
	}
}

// BenchmarkDirtyRectUpdate はダーティ領域のみ更新する場合のコストを測定
func BenchmarkDirtyRectUpdate(b *testing.B) {
	baseWidth, baseHeight := 640, 480
	baseImg := ebiten.NewImage(baseWidth, baseHeight)
	baseImg.Fill(color.RGBA{255, 255, 255, 255})

	// 小さな更新領域（100x100）
	updateWidth, updateHeight := 100, 100
	updateImg := ebiten.NewImage(updateWidth, updateHeight)
	updateImg.Fill(color.RGBA{255, 0, 0, 255})

	dst := ebiten.NewImage(baseWidth, baseHeight)
	dst.DrawImage(baseImg, nil)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 小さな領域のみ更新
		opts := &ebiten.DrawImageOptions{}
		x := (i * 5) % (baseWidth - updateWidth)
		y := (i * 7) % (baseHeight - updateHeight)
		opts.GeoM.Translate(float64(x), float64(y))
		dst.DrawImage(updateImg, opts)
	}
}

// BenchmarkFullRedraw は全画面再描画のコストを測定
func BenchmarkFullRedraw(b *testing.B) {
	baseWidth, baseHeight := 640, 480
	baseImg := ebiten.NewImage(baseWidth, baseHeight)
	baseImg.Fill(color.RGBA{255, 255, 255, 255})

	dst := ebiten.NewImage(baseWidth, baseHeight)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		dst.Clear()
		dst.DrawImage(baseImg, nil)
	}
}

// BenchmarkTextLayerCreate はテキストレイヤー作成のコストを測定
func BenchmarkTextLayerCreate1(b *testing.B) {
	benchmarkTextLayerCreate(b, 1)
}

func BenchmarkTextLayerCreate10(b *testing.B) {
	benchmarkTextLayerCreate(b, 10)
}

func BenchmarkTextLayerCreate50(b *testing.B) {
	benchmarkTextLayerCreate(b, 50)
}

func BenchmarkTextLayerCreate100(b *testing.B) {
	benchmarkTextLayerCreate(b, 100)
}

func benchmarkTextLayerCreate(b *testing.B, textCount int) {
	// TextRendererを作成
	tr := NewTextRenderer()
	tr.SetFont("", 14)
	tr.SetTextColor(color.RGBA{0, 0, 0, 255})

	// 640x480のピクチャーを作成
	baseWidth, baseHeight := 640, 480

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 毎回新しいピクチャーを作成
		ebitenImg := ebiten.NewImage(baseWidth, baseHeight)
		ebitenImg.Fill(color.RGBA{255, 255, 255, 255})

		originalRGBA := image.NewRGBA(image.Rect(0, 0, baseWidth, baseHeight))
		for y := 0; y < baseHeight; y++ {
			for x := 0; x < baseWidth; x++ {
				originalRGBA.Set(x, y, color.RGBA{255, 255, 255, 255})
			}
		}

		pic := &Picture{
			ID:            0,
			Image:         ebitenImg,
			OriginalImage: originalRGBA,
			Width:         baseWidth,
			Height:        baseHeight,
		}

		// テキストを描画
		for j := 0; j < textCount; j++ {
			x := (j * 20) % (baseWidth - 100)
			y := (j * 15) % (baseHeight - 20)
			tr.TextWrite(pic, x, y, "テスト文字列ABC123")
		}
	}
}

// BenchmarkTextLayerComposite はテキストレイヤー合成のコストを測定
// （レイヤーは事前に作成済みの場合）
func BenchmarkTextLayerComposite10(b *testing.B) {
	benchmarkTextLayerComposite(b, 10)
}

func BenchmarkTextLayerComposite50(b *testing.B) {
	benchmarkTextLayerComposite(b, 50)
}

func BenchmarkTextLayerComposite100(b *testing.B) {
	benchmarkTextLayerComposite(b, 100)
}

func benchmarkTextLayerComposite(b *testing.B, layerCount int) {
	// 背景画像を作成
	baseWidth, baseHeight := 640, 480
	background := image.NewRGBA(image.Rect(0, 0, baseWidth, baseHeight))
	for y := 0; y < baseHeight; y++ {
		for x := 0; x < baseWidth; x++ {
			background.Set(x, y, color.RGBA{255, 255, 255, 255})
		}
	}

	// TextRendererを作成してフォントを設定
	tr := NewTextRenderer()
	tr.SetFont("", 14)
	tr.SetTextColor(color.RGBA{0, 0, 0, 255})

	// テキストレイヤーを事前に作成
	layers := make([]*TextLayer, layerCount)
	for i := 0; i < layerCount; i++ {
		x := (i * 20) % (baseWidth - 100)
		y := (i * 15) % (baseHeight - 20)
		layers[i] = CreateTextLayer(
			background,
			tr.face,
			"テスト文字列ABC123",
			x, y,
			14,
			color.RGBA{0, 0, 0, 255},
			0,
		)
	}

	// 描画先
	dst := ebiten.NewImage(baseWidth, baseHeight)
	baseEbiten := ebiten.NewImageFromImage(background)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// ベース画像をコピー
		dst.Clear()
		dst.DrawImage(baseEbiten, nil)

		// 各テキストレイヤーを合成
		for _, layer := range layers {
			if layer == nil {
				continue
			}
			layerEbiten := ebiten.NewImageFromImage(layer.Image)
			opts := &ebiten.DrawImageOptions{}
			opts.GeoM.Translate(float64(layer.X), float64(layer.Y))
			dst.DrawImage(layerEbiten, opts)
		}
	}
}
