package graphics

import (
	"image"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Feature: layer-based-rendering, Property 1: レイヤー管理の一貫性
// **Validates: Requirements 1.1, 1.2, 1.3, 1.4, 1.5, 1.6**
//
// 任意のレイヤーマネージャーに対して、背景レイヤー、描画レイヤー、キャストレイヤー、
// テキストレイヤーを追加した場合、それらは正しいZ順序（背景 < 描画 < キャスト < テキスト）で管理される。

// TestProperty1_LayerZOrderConsistency はZ順序の一貫性をテストする
// 背景(0) < 描画(1) < キャスト(100+) < テキスト(1000+)
func TestProperty1_LayerZOrderConsistency(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 1.1: 背景レイヤーのZ順序は常に0
	properties.Property("背景レイヤーのZ順序は常に0（最背面）", prop.ForAll(
		func(picID int) bool {
			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()
			bg := NewBackgroundLayer(layerID, picID, nil)

			return bg.GetZOrder() == ZOrderBackground && bg.GetZOrder() == 0
		},
		gen.IntRange(0, 255),
	))

	// Property 1.2: 描画レイヤーのZ順序は常に1
	properties.Property("描画レイヤーのZ順序は常に1（背景の上、キャストの下）", prop.ForAll(
		func(picID int) bool {
			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()
			dl := NewDrawingLayer(layerID, picID, 640, 480)

			return dl.GetZOrder() == ZOrderDrawing && dl.GetZOrder() == 1
		},
		gen.IntRange(0, 255),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty1_CastLayerZOrder はキャストレイヤーのZ順序をテストする
func TestProperty1_CastLayerZOrder(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 1.3: キャストレイヤーのZ順序は操作順序に基づく
	// 要件 10.1, 10.2: 操作順序に基づくZ順序
	properties.Property("キャストレイヤーのZ順序は操作順序に基づく", prop.ForAll(
		func(zOrderOffset int) bool {
			if zOrderOffset < 0 || zOrderOffset > 899 {
				return true // 範囲外はスキップ
			}

			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()
			// NewCastLayerは初期Z順序を設定するが、AddCastLayerで上書きされる
			cl := NewCastLayer(layerID, 0, 0, 0, 0, 0, 0, 0, 32, 32, zOrderOffset)

			// NewCastLayerで設定されたZ順序を確認
			expectedZOrder := ZOrderCastBase + zOrderOffset
			return cl.GetZOrder() == expectedZOrder
		},
		gen.IntRange(0, 899),
	))

	// Property 1.4: 複数のキャストレイヤーを追加しても順序が正しい
	// 要件 10.1, 10.2: 操作順序に基づくZ順序
	properties.Property("複数のキャストレイヤーを追加しても順序が正しい", prop.ForAll(
		func(castCount int) bool {
			if castCount <= 0 || castCount > 100 {
				return true
			}

			lm := NewLayerManager()
			pls := lm.GetOrCreatePictureLayerSet(0)

			// キャストレイヤーを追加
			for i := 0; i < castCount; i++ {
				layerID := lm.GetNextLayerID()
				// AddCastLayerが操作順序に基づくZ順序を割り当てる
				cl := NewCastLayer(layerID, i, 0, 0, i*10, i*10, 0, 0, 32, 32, 0)
				pls.AddCastLayer(cl)
			}

			// Z順序が正しいことを確認（追加順に増加）
			for i := 0; i < len(pls.Casts)-1; i++ {
				if pls.Casts[i].GetZOrder() >= pls.Casts[i+1].GetZOrder() {
					return false
				}
			}

			// すべてのキャストレイヤーのZ順序が1以上であることを確認
			// 要件 10.1: 操作順序に基づくZ順序は1から開始
			for _, cast := range pls.Casts {
				if cast.GetZOrder() < 1 {
					return false
				}
			}

			return true
		},
		gen.IntRange(1, 100),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty1_TextLayerZOrder はテキストレイヤーのZ順序をテストする
func TestProperty1_TextLayerZOrder(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 1.5: テキストレイヤーのZ順序は操作順序に基づく
	// 要件 10.1, 10.2: 操作順序に基づくZ順序
	properties.Property("テキストレイヤーのZ順序は操作順序に基づく", prop.ForAll(
		func(zOrderOffset int) bool {
			if zOrderOffset < 0 || zOrderOffset > 255 {
				return true // 範囲外はスキップ
			}

			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()
			// NewTextLayerEntryは初期Z順序を設定するが、AddTextLayerで上書きされる
			tl := NewTextLayerEntry(layerID, 0, 0, 0, "test", zOrderOffset)

			// NewTextLayerEntryで設定されたZ順序を確認
			expectedZOrder := ZOrderTextBase + zOrderOffset
			return tl.GetZOrder() == expectedZOrder
		},
		gen.IntRange(0, 255),
	))

	// Property 1.6: 複数のテキストレイヤーを追加しても順序が正しい
	// 要件 10.1, 10.2: 操作順序に基づくZ順序
	properties.Property("複数のテキストレイヤーを追加しても順序が正しい", prop.ForAll(
		func(textCount int) bool {
			if textCount <= 0 || textCount > 100 {
				return true
			}

			lm := NewLayerManager()
			pls := lm.GetOrCreatePictureLayerSet(0)

			// テキストレイヤーを追加
			for i := 0; i < textCount; i++ {
				layerID := lm.GetNextLayerID()
				// AddTextLayerが操作順序に基づくZ順序を割り当てる
				tl := NewTextLayerEntry(layerID, 0, i*10, i*10, "test", 0)
				pls.AddTextLayer(tl)
			}

			// Z順序が正しいことを確認（追加順に増加）
			for i := 0; i < len(pls.Texts)-1; i++ {
				if pls.Texts[i].GetZOrder() >= pls.Texts[i+1].GetZOrder() {
					return false
				}
			}

			// すべてのテキストレイヤーのZ順序が1以上であることを確認
			// 要件 10.1: 操作順序に基づくZ順序は1から開始
			for _, text := range pls.Texts {
				if text.GetZOrder() < 1 {
					return false
				}
			}

			return true
		},
		gen.IntRange(1, 100),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty1_LayerZOrderHierarchy はレイヤー階層全体のZ順序をテストする
// 要件 10.1, 10.2: 操作順序に基づくZ順序
// 背景は常にZ=0、その他のレイヤーは操作順序に基づく
func TestProperty1_LayerZOrderHierarchy(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 1.7: 操作順序に基づくZ順序が維持される
	// 背景は常にZ=0、その他は操作順序に基づく
	properties.Property("操作順序に基づくZ順序が維持される", prop.ForAll(
		func(castCount, textCount int) bool {
			if castCount <= 0 || castCount > 50 || textCount <= 0 || textCount > 50 {
				return true
			}

			lm := NewLayerManager()
			pls := lm.GetOrCreatePictureLayerSet(0)

			// 背景レイヤーを設定（Z=0）
			bgID := lm.GetNextLayerID()
			bg := NewBackgroundLayer(bgID, 0, nil)
			pls.SetBackground(bg)

			// 描画レイヤーを設定（Z=1、後方互換性）
			dlID := lm.GetNextLayerID()
			dl := NewDrawingLayer(dlID, 0, 640, 480)
			pls.SetDrawing(dl)

			// キャストレイヤーを追加（操作順序に基づくZ順序）
			for i := 0; i < castCount; i++ {
				layerID := lm.GetNextLayerID()
				cl := NewCastLayer(layerID, i, 0, 0, i*10, i*10, 0, 0, 32, 32, 0)
				pls.AddCastLayer(cl)
			}

			// テキストレイヤーを追加（操作順序に基づくZ順序）
			for i := 0; i < textCount; i++ {
				layerID := lm.GetNextLayerID()
				tl := NewTextLayerEntry(layerID, 0, i*10, i*10, "test", 0)
				pls.AddTextLayer(tl)
			}

			// 背景は常にZ=0
			if pls.Background.GetZOrder() != 0 {
				return false
			}

			// 描画レイヤーはZ=1（後方互換性）
			if pls.Drawing.GetZOrder() != 1 {
				return false
			}

			// キャスト内のZ順序が正しい（追加順に増加）
			for i := 0; i < len(pls.Casts)-1; i++ {
				if pls.Casts[i].GetZOrder() >= pls.Casts[i+1].GetZOrder() {
					return false
				}
			}

			// テキスト内のZ順序が正しい（追加順に増加）
			for i := 0; i < len(pls.Texts)-1; i++ {
				if pls.Texts[i].GetZOrder() >= pls.Texts[i+1].GetZOrder() {
					return false
				}
			}

			// キャストはテキストより先に追加されたので、キャストのZ順序 < テキストのZ順序
			// （この特定のテストケースでは）
			if len(pls.Casts) > 0 && len(pls.Texts) > 0 {
				lastCastZ := pls.Casts[len(pls.Casts)-1].GetZOrder()
				firstTextZ := pls.Texts[0].GetZOrder()
				if lastCastZ >= firstTextZ {
					return false
				}
			}

			return true
		},
		gen.IntRange(1, 50),
		gen.IntRange(1, 50),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty1_LayerAdditionMaintainsZOrder はレイヤー追加後もZ順序が維持されることをテストする
// 要件 10.1, 10.2: 操作順序に基づくZ順序
func TestProperty1_LayerAdditionMaintainsZOrder(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 1.8: レイヤー追加後もZ順序が維持される（操作順序に基づく）
	properties.Property("レイヤー追加後もZ順序が維持される（操作順序に基づく）", prop.ForAll(
		func(operations []int) bool {
			if len(operations) == 0 || len(operations) > 100 {
				return true
			}

			lm := NewLayerManager()
			pls := lm.GetOrCreatePictureLayerSet(0)

			// 背景と描画レイヤーを設定
			bgID := lm.GetNextLayerID()
			bg := NewBackgroundLayer(bgID, 0, nil)
			pls.SetBackground(bg)

			dlID := lm.GetNextLayerID()
			dl := NewDrawingLayer(dlID, 0, 640, 480)
			pls.SetDrawing(dl)

			// 追加順序を記録
			var addedLayers []Layer

			// ランダムな順序でキャストとテキストを追加
			for _, op := range operations {
				if op%2 == 0 {
					// キャストを追加
					layerID := lm.GetNextLayerID()
					cl := NewCastLayer(layerID, pls.GetCastLayerCount(), 0, 0, 0, 0, 0, 0, 32, 32, 0)
					pls.AddCastLayer(cl)
					addedLayers = append(addedLayers, cl)
				} else {
					// テキストを追加
					layerID := lm.GetNextLayerID()
					tl := NewTextLayerEntry(layerID, 0, 0, 0, "test", 0)
					pls.AddTextLayer(tl)
					addedLayers = append(addedLayers, tl)
				}
			}

			// 背景は常にZ=0
			if pls.Background.GetZOrder() != 0 {
				return false
			}

			// 描画レイヤーはZ=1（後方互換性）
			if pls.Drawing.GetZOrder() != 1 {
				return false
			}

			// 追加順序に基づくZ順序が正しいことを確認
			for i := 0; i < len(addedLayers)-1; i++ {
				if addedLayers[i].GetZOrder() >= addedLayers[i+1].GetZOrder() {
					return false
				}
			}

			// キャスト内のZ順序が正しい（追加順に増加）
			for i := 0; i < len(pls.Casts)-1; i++ {
				if pls.Casts[i].GetZOrder() >= pls.Casts[i+1].GetZOrder() {
					return false
				}
			}

			// テキスト内のZ順序が正しい
			for i := 0; i < len(pls.Texts)-1; i++ {
				if pls.Texts[i].GetZOrder() >= pls.Texts[i+1].GetZOrder() {
					return false
				}
			}

			return true
		},
		gen.SliceOfN(100, gen.IntRange(0, 100)),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty1_ZOrderConstants はZ順序の定数が正しいことをテストする
func TestProperty1_ZOrderConstants(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 1.9: Z順序の定数が正しい階層を形成する
	properties.Property("Z順序の定数が正しい階層を形成する", prop.ForAll(
		func(_ int) bool {
			// 背景(0) < 描画(1) < キャストベース(100) < テキストベース(1000)
			return ZOrderBackground < ZOrderDrawing &&
				ZOrderDrawing < ZOrderCastBase &&
				ZOrderCastBase < ZOrderTextBase
		},
		gen.IntRange(0, 100),
	))

	// Property 1.10: キャストのZ順序範囲がテキストと重ならない
	properties.Property("キャストのZ順序範囲がテキストと重ならない", prop.ForAll(
		func(castOffset int) bool {
			if castOffset < 0 || castOffset > 899 {
				return true
			}

			castZOrder := ZOrderCastBase + castOffset
			return castZOrder < ZOrderTextBase
		},
		gen.IntRange(0, 899),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty1_MultiplePictureLayers は複数のピクチャーでレイヤー管理が独立していることをテストする
func TestProperty1_MultiplePictureLayers(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 1.11: 複数のピクチャーでレイヤー管理が独立している
	// 要件 10.1, 10.2: 操作順序に基づくZ順序
	properties.Property("複数のピクチャーでレイヤー管理が独立している", prop.ForAll(
		func(picCount, layerCount int) bool {
			if picCount <= 0 || picCount > 10 || layerCount <= 0 || layerCount > 20 {
				return true
			}

			lm := NewLayerManager()

			// 各ピクチャーにレイヤーを追加
			for picID := 0; picID < picCount; picID++ {
				pls := lm.GetOrCreatePictureLayerSet(picID)

				// 背景と描画レイヤーを設定
				bgID := lm.GetNextLayerID()
				bg := NewBackgroundLayer(bgID, picID, nil)
				pls.SetBackground(bg)

				dlID := lm.GetNextLayerID()
				dl := NewDrawingLayer(dlID, picID, 640, 480)
				pls.SetDrawing(dl)

				// キャストとテキストを追加（操作順序に基づくZ順序）
				for i := 0; i < layerCount; i++ {
					castID := lm.GetNextLayerID()
					cl := NewCastLayer(castID, i, picID, 0, i*10, i*10, 0, 0, 32, 32, 0)
					pls.AddCastLayer(cl)

					textID := lm.GetNextLayerID()
					tl := NewTextLayerEntry(textID, picID, i*10, i*10, "test", 0)
					pls.AddTextLayer(tl)
				}
			}

			// 各ピクチャーのZ順序が正しいことを確認
			for picID := 0; picID < picCount; picID++ {
				pls := lm.GetPictureLayerSet(picID)
				if pls == nil {
					return false
				}

				// 背景は常にZ=0
				if pls.Background.GetZOrder() != 0 {
					return false
				}

				// 描画レイヤーはZ=1（後方互換性）
				if pls.Drawing.GetZOrder() != 1 {
					return false
				}

				// キャスト内のZ順序が正しい（追加順に増加）
				for i := 0; i < len(pls.Casts)-1; i++ {
					if pls.Casts[i].GetZOrder() >= pls.Casts[i+1].GetZOrder() {
						return false
					}
				}

				// テキスト内のZ順序が正しい（追加順に増加）
				for i := 0; i < len(pls.Texts)-1; i++ {
					if pls.Texts[i].GetZOrder() >= pls.Texts[i+1].GetZOrder() {
						return false
					}
				}

				// このテストでは、キャスト→テキストの順で交互に追加しているので、
				// 各キャストの次のテキストはキャストより大きいZ順序を持つ
				for i := 0; i < len(pls.Casts) && i < len(pls.Texts); i++ {
					if pls.Casts[i].GetZOrder() >= pls.Texts[i].GetZOrder() {
						return false
					}
				}
			}

			return true
		},
		gen.IntRange(1, 10),
		gen.IntRange(1, 20),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// ============================================================================
// Feature: layer-based-rendering, Property 3: ダーティフラグの正確性
// **Validates: Requirements 3.1, 3.2, 3.3, 3.4, 3.5**
//
// 任意のレイヤーに対して、位置、内容、または可視性が変更された場合、
// ダーティフラグが設定され、合成処理後にクリアされる。
// ============================================================================

// TestProperty3_PositionChangeSetsDirectyFlag は位置変更時にダーティフラグが設定されることをテストする
// 要件 3.1: 位置が変更されたときにダーティフラグを設定する
func TestProperty3_PositionChangeSetsDirectyFlag(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 3.1.1: CastLayerの位置変更時にダーティフラグが設定される
	properties.Property("CastLayerの位置変更時にダーティフラグが設定される", prop.ForAll(
		func(x1, y1, x2, y2 int) bool {
			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()
			cl := NewCastLayer(layerID, 0, 0, 0, x1, y1, 0, 0, 32, 32, 0)

			// 初期状態はダーティ
			if !cl.IsDirty() {
				return false
			}

			// ダーティフラグをクリア
			cl.SetDirty(false)
			if cl.IsDirty() {
				return false
			}

			// 位置を変更
			cl.SetPosition(x2, y2)

			// 位置が変わった場合のみダーティフラグが設定される
			if x1 != x2 || y1 != y2 {
				return cl.IsDirty()
			}
			// 位置が同じ場合はダーティフラグは設定されない
			return !cl.IsDirty()
		},
		gen.IntRange(0, 640),
		gen.IntRange(0, 480),
		gen.IntRange(0, 640),
		gen.IntRange(0, 480),
	))

	// Property 3.1.2: TextLayerEntryの位置変更時にダーティフラグが設定される
	properties.Property("TextLayerEntryの位置変更時にダーティフラグが設定される", prop.ForAll(
		func(x1, y1, x2, y2 int) bool {
			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()
			tl := NewTextLayerEntry(layerID, 0, x1, y1, "test", 0)

			// 初期状態はダーティ
			if !tl.IsDirty() {
				return false
			}

			// ダーティフラグをクリア
			tl.SetDirty(false)
			if tl.IsDirty() {
				return false
			}

			// 位置を変更
			tl.SetPosition(x2, y2)

			// 位置が変わった場合のみダーティフラグが設定される
			if x1 != x2 || y1 != y2 {
				return tl.IsDirty()
			}
			// 位置が同じ場合はダーティフラグは設定されない
			return !tl.IsDirty()
		},
		gen.IntRange(0, 640),
		gen.IntRange(0, 480),
		gen.IntRange(0, 640),
		gen.IntRange(0, 480),
	))

	// Property 3.1.3: BaseLayerのSetBoundsでダーティフラグが設定される
	properties.Property("BaseLayerのSetBoundsでダーティフラグが設定される", prop.ForAll(
		func(x1, y1, w1, h1, x2, y2, w2, h2 int) bool {
			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()
			bg := NewBackgroundLayer(layerID, 0, nil)

			// ダーティフラグをクリア
			bg.SetDirty(false)

			// 境界を設定
			bounds1 := image.Rect(x1, y1, x1+w1, y1+h1)
			bounds2 := image.Rect(x2, y2, x2+w2, y2+h2)
			bg.SetBounds(bounds1)
			bg.SetDirty(false)
			bg.SetBounds(bounds2)

			// 境界が変わった場合のみダーティフラグが設定される
			if bounds1 != bounds2 {
				return bg.IsDirty()
			}
			return !bg.IsDirty()
		},
		gen.IntRange(0, 100), gen.IntRange(0, 100), gen.IntRange(1, 100), gen.IntRange(1, 100),
		gen.IntRange(0, 100), gen.IntRange(0, 100), gen.IntRange(1, 100), gen.IntRange(1, 100),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty3_ContentChangeSetsDirectyFlag は内容変更時にダーティフラグが設定されることをテストする
// 要件 3.2: 内容が変更されたときにダーティフラグを設定する
func TestProperty3_ContentChangeSetsDirectyFlag(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 3.2.1: CastLayerのソース領域変更時にダーティフラグが設定される
	properties.Property("CastLayerのソース領域変更時にダーティフラグが設定される", prop.ForAll(
		func(srcX1, srcY1, w1, h1, srcX2, srcY2, w2, h2 int) bool {
			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()
			cl := NewCastLayer(layerID, 0, 0, 0, 0, 0, srcX1, srcY1, w1, h1, 0)

			// ダーティフラグをクリア
			cl.SetDirty(false)

			// ソース領域を変更
			cl.SetSourceRect(srcX2, srcY2, w2, h2)

			// ソース領域が変わった場合のみダーティフラグが設定される
			if srcX1 != srcX2 || srcY1 != srcY2 || w1 != w2 || h1 != h2 {
				return cl.IsDirty()
			}
			return !cl.IsDirty()
		},
		gen.IntRange(0, 100), gen.IntRange(0, 100), gen.IntRange(1, 64), gen.IntRange(1, 64),
		gen.IntRange(0, 100), gen.IntRange(0, 100), gen.IntRange(1, 64), gen.IntRange(1, 64),
	))

	// Property 3.2.2: TextLayerEntryのテキスト変更時にダーティフラグが設定される
	properties.Property("TextLayerEntryのテキスト変更時にダーティフラグが設定される", prop.ForAll(
		func(textIndex1, textIndex2 int) bool {
			texts := []string{"test1", "test2", "test3", "hello", "world"}
			text1 := texts[textIndex1%len(texts)]
			text2 := texts[textIndex2%len(texts)]

			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()
			tl := NewTextLayerEntry(layerID, 0, 0, 0, text1, 0)

			// ダーティフラグをクリア
			tl.SetDirty(false)

			// テキストを変更
			tl.SetText(text2)

			// テキストが変わった場合のみダーティフラグが設定される
			if text1 != text2 {
				return tl.IsDirty()
			}
			return !tl.IsDirty()
		},
		gen.IntRange(0, 4),
		gen.IntRange(0, 4),
	))

	// Property 3.2.3: CastLayerのInvalidateでダーティフラグが設定される
	properties.Property("CastLayerのInvalidateでダーティフラグが設定される", prop.ForAll(
		func(x, y int) bool {
			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()
			cl := NewCastLayer(layerID, 0, 0, 0, x, y, 0, 0, 32, 32, 0)

			// ダーティフラグをクリア
			cl.SetDirty(false)
			if cl.IsDirty() {
				return false
			}

			// Invalidateを呼び出す
			cl.Invalidate()

			// ダーティフラグが設定される
			return cl.IsDirty()
		},
		gen.IntRange(0, 640),
		gen.IntRange(0, 480),
	))

	// Property 3.2.4: TextLayerEntryのInvalidateでダーティフラグが設定される
	properties.Property("TextLayerEntryのInvalidateでダーティフラグが設定される", prop.ForAll(
		func(x, y int) bool {
			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()
			tl := NewTextLayerEntry(layerID, 0, x, y, "test", 0)

			// ダーティフラグをクリア
			tl.SetDirty(false)
			if tl.IsDirty() {
				return false
			}

			// Invalidateを呼び出す
			tl.Invalidate()

			// ダーティフラグが設定される
			return tl.IsDirty()
		},
		gen.IntRange(0, 640),
		gen.IntRange(0, 480),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty3_VisibilityChangeSetsDirectyFlag は可視性変更時にダーティフラグが設定されることをテストする
// 要件 3.3: 可視性が変更されたときにダーティフラグを設定する
func TestProperty3_VisibilityChangeSetsDirectyFlag(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 3.3.1: BaseLayerのSetVisibleでダーティフラグが設定される
	properties.Property("BaseLayerのSetVisibleでダーティフラグが設定される", prop.ForAll(
		func(initialVisible, newVisible bool) bool {
			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()
			bg := NewBackgroundLayer(layerID, 0, nil)

			// 初期可視性を設定
			bg.SetVisible(initialVisible)
			bg.SetDirty(false)

			// 可視性を変更
			bg.SetVisible(newVisible)

			// 可視性が変わった場合のみダーティフラグが設定される
			if initialVisible != newVisible {
				return bg.IsDirty()
			}
			return !bg.IsDirty()
		},
		gen.Bool(),
		gen.Bool(),
	))

	// Property 3.3.2: CastLayerのSetVisibleでダーティフラグが設定される
	properties.Property("CastLayerのSetVisibleでダーティフラグが設定される", prop.ForAll(
		func(initialVisible, newVisible bool) bool {
			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()
			cl := NewCastLayer(layerID, 0, 0, 0, 0, 0, 0, 0, 32, 32, 0)

			// 初期可視性を設定
			cl.SetVisible(initialVisible)
			cl.SetDirty(false)

			// 可視性を変更
			cl.SetVisible(newVisible)

			// 可視性が変わった場合のみダーティフラグが設定される
			if initialVisible != newVisible {
				return cl.IsDirty()
			}
			return !cl.IsDirty()
		},
		gen.Bool(),
		gen.Bool(),
	))

	// Property 3.3.3: TextLayerEntryのSetVisibleでダーティフラグが設定される
	properties.Property("TextLayerEntryのSetVisibleでダーティフラグが設定される", prop.ForAll(
		func(initialVisible, newVisible bool) bool {
			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()
			tl := NewTextLayerEntry(layerID, 0, 0, 0, "test", 0)

			// 初期可視性を設定
			tl.SetVisible(initialVisible)
			tl.SetDirty(false)

			// 可視性を変更
			tl.SetVisible(newVisible)

			// 可視性が変わった場合のみダーティフラグが設定される
			if initialVisible != newVisible {
				return tl.IsDirty()
			}
			return !tl.IsDirty()
		},
		gen.Bool(),
		gen.Bool(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty3_DirtyFlagClearedAfterComposite は合成処理後にダーティフラグがクリアされることをテストする
// 要件 3.4: 合成処理が完了したときにすべてのDirty_Flagをクリアする
func TestProperty3_DirtyFlagClearedAfterComposite(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 3.4.1: ClearAllDirtyFlagsですべてのレイヤーのダーティフラグがクリアされる
	properties.Property("ClearAllDirtyFlagsですべてのレイヤーのダーティフラグがクリアされる", prop.ForAll(
		func(castCount, textCount int) bool {
			if castCount < 0 || castCount > 50 || textCount < 0 || textCount > 50 {
				return true
			}

			lm := NewLayerManager()
			pls := lm.GetOrCreatePictureLayerSet(0)

			// 背景レイヤーを設定
			bgID := lm.GetNextLayerID()
			bg := NewBackgroundLayer(bgID, 0, nil)
			pls.SetBackground(bg)

			// 描画レイヤーを設定
			dlID := lm.GetNextLayerID()
			dl := NewDrawingLayer(dlID, 0, 640, 480)
			pls.SetDrawing(dl)

			// キャストレイヤーを追加
			for i := 0; i < castCount; i++ {
				layerID := lm.GetNextLayerID()
				zOrderOffset := pls.GetNextCastZOffset()
				cl := NewCastLayer(layerID, i, 0, 0, i*10, i*10, 0, 0, 32, 32, zOrderOffset)
				pls.AddCastLayer(cl)
			}

			// テキストレイヤーを追加
			for i := 0; i < textCount; i++ {
				layerID := lm.GetNextLayerID()
				zOrderOffset := pls.GetNextTextZOffset()
				tl := NewTextLayerEntry(layerID, 0, i*10, i*10, "test", zOrderOffset)
				pls.AddTextLayer(tl)
			}

			// すべてのレイヤーがダーティであることを確認
			if pls.Background != nil && !pls.Background.IsDirty() {
				return false
			}
			if pls.Drawing != nil && !pls.Drawing.IsDirty() {
				return false
			}

			// ClearAllDirtyFlagsを呼び出す
			pls.ClearAllDirtyFlags()

			// すべてのレイヤーのダーティフラグがクリアされていることを確認
			if pls.Background != nil && pls.Background.IsDirty() {
				return false
			}
			if pls.Drawing != nil && pls.Drawing.IsDirty() {
				return false
			}
			for _, cast := range pls.Casts {
				if cast.IsDirty() {
					return false
				}
			}
			for _, text := range pls.Texts {
				if text.IsDirty() {
					return false
				}
			}

			// PictureLayerSetのFullDirtyもクリアされていることを確認
			if pls.FullDirty {
				return false
			}

			// DirtyRegionもクリアされていることを確認
			if !pls.DirtyRegion.Empty() {
				return false
			}

			return true
		},
		gen.IntRange(0, 50),
		gen.IntRange(0, 50),
	))

	// Property 3.4.2: 個別のSetDirty(false)でダーティフラグがクリアされる
	properties.Property("個別のSetDirty(false)でダーティフラグがクリアされる", prop.ForAll(
		func(layerType int) bool {
			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()

			var layer Layer
			switch layerType % 3 {
			case 0:
				layer = NewBackgroundLayer(layerID, 0, nil)
			case 1:
				layer = NewCastLayer(layerID, 0, 0, 0, 0, 0, 0, 0, 32, 32, 0)
			case 2:
				layer = NewTextLayerEntry(layerID, 0, 0, 0, "test", 0)
			}

			// 初期状態はダーティ
			if !layer.IsDirty() {
				return false
			}

			// SetDirty(false)でクリア
			layer.SetDirty(false)

			// ダーティフラグがクリアされていることを確認
			return !layer.IsDirty()
		},
		gen.IntRange(0, 100),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty3_NonDirtyLayerUsesCache はダーティでないレイヤーがキャッシュを使用することをテストする
// 要件 3.5: Dirty_Flagが設定されていないレイヤーがあるときにそのレイヤーのキャッシュを使用する
func TestProperty3_NonDirtyLayerUsesCache(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 3.5.1: PictureLayerSetのIsDirtyがレイヤーのダーティ状態を正しく反映する
	properties.Property("PictureLayerSetのIsDirtyがレイヤーのダーティ状態を正しく反映する", prop.ForAll(
		func(castCount int) bool {
			if castCount <= 0 || castCount > 20 {
				return true
			}

			lm := NewLayerManager()
			pls := lm.GetOrCreatePictureLayerSet(0)

			// 背景レイヤーを設定
			bgID := lm.GetNextLayerID()
			bg := NewBackgroundLayer(bgID, 0, nil)
			pls.SetBackground(bg)

			// キャストレイヤーを追加
			for i := 0; i < castCount; i++ {
				layerID := lm.GetNextLayerID()
				zOrderOffset := pls.GetNextCastZOffset()
				cl := NewCastLayer(layerID, i, 0, 0, i*10, i*10, 0, 0, 32, 32, zOrderOffset)
				pls.AddCastLayer(cl)
			}

			// 初期状態はダーティ
			if !pls.IsDirty() {
				return false
			}

			// すべてのダーティフラグをクリア
			pls.ClearAllDirtyFlags()

			// ダーティでないことを確認
			if pls.IsDirty() {
				return false
			}

			// 1つのレイヤーをダーティにする
			if len(pls.Casts) > 0 {
				pls.Casts[0].SetDirty(true)
				// PictureLayerSetがダーティになることを確認
				if !pls.IsDirty() {
					return false
				}
			}

			return true
		},
		gen.IntRange(1, 20),
	))

	// Property 3.5.2: ダーティでないレイヤーのGetImageはキャッシュを返す
	properties.Property("ダーティでないレイヤーのGetImageはキャッシュを返す（TextLayerEntry）", prop.ForAll(
		func(x, y int) bool {
			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()
			tl := NewTextLayerEntry(layerID, 0, x, y, "test", 0)

			// 初期状態では画像はnil
			if tl.GetImage() != nil {
				return false
			}

			// ダーティフラグをクリア（画像なしでもクリア可能）
			tl.SetDirty(false)

			// GetImageを呼び出してもダーティフラグは変わらない
			_ = tl.GetImage()
			return !tl.IsDirty()
		},
		gen.IntRange(0, 640),
		gen.IntRange(0, 480),
	))

	// Property 3.5.3: FullDirtyフラグがPictureLayerSetのダーティ状態に影響する
	properties.Property("FullDirtyフラグがPictureLayerSetのダーティ状態に影響する", prop.ForAll(
		func(setFullDirty bool) bool {
			lm := NewLayerManager()
			pls := lm.GetOrCreatePictureLayerSet(0)

			// 背景レイヤーを設定
			bgID := lm.GetNextLayerID()
			bg := NewBackgroundLayer(bgID, 0, nil)
			pls.SetBackground(bg)

			// すべてのダーティフラグをクリア
			pls.ClearAllDirtyFlags()

			// ダーティでないことを確認
			if pls.IsDirty() {
				return false
			}

			// FullDirtyを設定
			if setFullDirty {
				pls.MarkFullDirty()
				return pls.IsDirty() && pls.FullDirty
			}

			return !pls.IsDirty()
		},
		gen.Bool(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty3_DirtyFlagSequence はダーティフラグの一連の操作をテストする
// 複合的なシナリオでダーティフラグが正しく動作することを確認
func TestProperty3_DirtyFlagSequence(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 3.6: 複数の変更操作後にダーティフラグが正しく設定される
	properties.Property("複数の変更操作後にダーティフラグが正しく設定される", prop.ForAll(
		func(operations []int) bool {
			if len(operations) == 0 || len(operations) > 50 {
				return true
			}

			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()
			cl := NewCastLayer(layerID, 0, 0, 0, 0, 0, 0, 0, 32, 32, 0)

			// 各操作を実行
			for _, op := range operations {
				switch op % 4 {
				case 0:
					// 位置変更
					cl.SetPosition(op%640, op%480)
				case 1:
					// 可視性変更
					cl.SetVisible(op%2 == 0)
				case 2:
					// ダーティフラグクリア
					cl.SetDirty(false)
				case 3:
					// Invalidate
					cl.Invalidate()
				}
			}

			// 最後の操作がクリアでなければダーティであるべき
			lastOp := operations[len(operations)-1] % 4
			if lastOp == 2 {
				// 最後がクリアなら、その前の操作に依存
				// クリア後は必ず非ダーティ
				return !cl.IsDirty()
			}

			// 最後がInvalidateなら必ずダーティ
			if lastOp == 3 {
				return cl.IsDirty()
			}

			// 位置変更や可視性変更は、値が変わった場合のみダーティ
			// この場合は常にダーティになる可能性がある
			return true
		},
		gen.SliceOfN(50, gen.IntRange(0, 100)),
	))

	// Property 3.7: DirtyRegionの追加と統合が正しく動作する
	properties.Property("DirtyRegionの追加と統合が正しく動作する", prop.ForAll(
		func(rects []int) bool {
			if len(rects) < 4 || len(rects) > 40 {
				return true
			}

			lm := NewLayerManager()
			pls := lm.GetOrCreatePictureLayerSet(0)

			// ダーティ領域をクリア
			pls.ClearDirtyRegion()

			// 複数の矩形を追加
			var expectedUnion image.Rectangle
			for i := 0; i+3 < len(rects); i += 4 {
				x := rects[i] % 640
				y := rects[i+1] % 480
				w := (rects[i+2] % 100) + 1
				h := (rects[i+3] % 100) + 1
				rect := image.Rect(x, y, x+w, y+h)

				pls.AddDirtyRegion(rect)

				if expectedUnion.Empty() {
					expectedUnion = rect
				} else {
					expectedUnion = expectedUnion.Union(rect)
				}
			}

			// 統合された領域が期待通りであることを確認
			return pls.DirtyRegion == expectedUnion
		},
		gen.SliceOfN(40, gen.IntRange(0, 640)),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// ============================================================================
// Feature: layer-based-rendering, Property 5: キャッシュ管理の正確性
// **Validates: Requirements 5.1, 5.2, 5.3**
//
// 任意のレイヤーに対して、内容が変更されていない場合はキャッシュが使用され、
// 変更された場合はキャッシュが無効化される。
// ============================================================================

// TestProperty5_LayerCacheStorage は各レイヤーの描画結果をキャッシュすることをテストする
// 要件 5.1: 各レイヤーの描画結果をキャッシュする
func TestProperty5_LayerCacheStorage(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 5.1.1: BackgroundLayerは画像をキャッシュする
	properties.Property("BackgroundLayerは画像をキャッシュする", prop.ForAll(
		func(picID int) bool {
			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()

			// 画像なしで作成
			bg := NewBackgroundLayer(layerID, picID, nil)
			if bg.GetImage() != nil {
				return false
			}

			// 画像を設定
			img := ebiten.NewImage(64, 64)
			bg.SetImage(img)

			// キャッシュされた画像が返される
			cachedImg := bg.GetImage()
			return cachedImg == img
		},
		gen.IntRange(0, 255),
	))

	// Property 5.1.2: DrawingLayerは画像をキャッシュする
	properties.Property("DrawingLayerは画像をキャッシュする", prop.ForAll(
		func(width, height int) bool {
			if width <= 0 || height <= 0 || width > 1024 || height > 1024 {
				return true
			}

			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()

			// 描画レイヤーを作成（内部で画像が作成される）
			dl := NewDrawingLayer(layerID, 0, width, height)

			// キャッシュされた画像が返される
			cachedImg := dl.GetImage()
			if cachedImg == nil {
				return false
			}

			// サイズが正しいことを確認
			bounds := cachedImg.Bounds()
			return bounds.Dx() == width && bounds.Dy() == height
		},
		gen.IntRange(1, 256),
		gen.IntRange(1, 256),
	))

	// Property 5.1.3: TextLayerEntryは画像をキャッシュする
	properties.Property("TextLayerEntryは画像をキャッシュする", prop.ForAll(
		func(x, y int) bool {
			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()

			// テキストレイヤーを作成（初期状態では画像なし）
			tl := NewTextLayerEntry(layerID, 0, x, y, "test", 0)
			if tl.GetImage() != nil {
				return false
			}

			// 画像を設定
			img := ebiten.NewImage(100, 20)
			tl.SetImage(img)

			// キャッシュされた画像が返される
			cachedImg := tl.GetImage()
			return cachedImg == img
		},
		gen.IntRange(0, 640),
		gen.IntRange(0, 480),
	))

	// Property 5.1.4: CastLayerはソース画像からキャッシュを生成する
	properties.Property("CastLayerはソース画像からキャッシュを生成する", prop.ForAll(
		func(srcX, srcY, width, height int) bool {
			if width <= 0 || height <= 0 || width > 64 || height > 64 {
				return true
			}
			if srcX < 0 || srcY < 0 || srcX > 100 || srcY > 100 {
				return true
			}

			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()

			// キャストレイヤーを作成
			cl := NewCastLayer(layerID, 0, 0, 0, 0, 0, srcX, srcY, width, height, 0)

			// ソース画像なしではキャッシュはnil
			if cl.GetImage() != nil {
				return false
			}

			// ソース画像を設定（十分なサイズ）
			srcImg := ebiten.NewImage(256, 256)
			cl.SetSourceImage(srcImg)

			// キャッシュが生成される
			cachedImg := cl.GetImage()
			return cachedImg != nil
		},
		gen.IntRange(0, 100),
		gen.IntRange(0, 100),
		gen.IntRange(1, 64),
		gen.IntRange(1, 64),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty5_CacheUsedWhenNotDirty は内容が変更されていないときにキャッシュが使用されることをテストする
// 要件 5.2: 内容が変更されていないときはキャッシュされた画像を使用する
func TestProperty5_CacheUsedWhenNotDirty(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 5.2.1: BackgroundLayerはダーティでないときに同じ画像を返す
	properties.Property("BackgroundLayerはダーティでないときに同じ画像を返す", prop.ForAll(
		func(picID int) bool {
			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()

			img := ebiten.NewImage(64, 64)
			bg := NewBackgroundLayer(layerID, picID, img)

			// ダーティフラグをクリア
			bg.SetDirty(false)

			// 複数回GetImageを呼び出しても同じ画像が返される
			img1 := bg.GetImage()
			img2 := bg.GetImage()
			img3 := bg.GetImage()

			return img1 == img && img2 == img && img3 == img
		},
		gen.IntRange(0, 255),
	))

	// Property 5.2.2: DrawingLayerはダーティでないときに同じ画像を返す
	properties.Property("DrawingLayerはダーティでないときに同じ画像を返す", prop.ForAll(
		func(width, height int) bool {
			if width <= 0 || height <= 0 || width > 256 || height > 256 {
				return true
			}

			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()

			dl := NewDrawingLayer(layerID, 0, width, height)

			// ダーティフラグをクリア
			dl.SetDirty(false)

			// 複数回GetImageを呼び出しても同じ画像が返される
			img1 := dl.GetImage()
			img2 := dl.GetImage()
			img3 := dl.GetImage()

			return img1 == img2 && img2 == img3 && img1 != nil
		},
		gen.IntRange(1, 256),
		gen.IntRange(1, 256),
	))

	// Property 5.2.3: TextLayerEntryはダーティでないときに同じ画像を返す
	properties.Property("TextLayerEntryはダーティでないときに同じ画像を返す", prop.ForAll(
		func(x, y int) bool {
			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()

			tl := NewTextLayerEntry(layerID, 0, x, y, "test", 0)

			// 画像を設定
			img := ebiten.NewImage(100, 20)
			tl.SetImage(img)

			// SetImageはダーティフラグをクリアする
			if tl.IsDirty() {
				return false
			}

			// 複数回GetImageを呼び出しても同じ画像が返される
			img1 := tl.GetImage()
			img2 := tl.GetImage()
			img3 := tl.GetImage()

			return img1 == img && img2 == img && img3 == img
		},
		gen.IntRange(0, 640),
		gen.IntRange(0, 480),
	))

	// Property 5.2.4: CastLayerはダーティでないときにキャッシュを再生成しない
	properties.Property("CastLayerはダーティでないときにキャッシュを再生成しない", prop.ForAll(
		func(width, height int) bool {
			if width <= 0 || height <= 0 || width > 64 || height > 64 {
				return true
			}

			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()

			cl := NewCastLayer(layerID, 0, 0, 0, 0, 0, 0, 0, width, height, 0)

			// ソース画像を設定
			srcImg := ebiten.NewImage(128, 128)
			cl.SetSourceImage(srcImg)

			// ダーティフラグをクリア
			cl.SetDirty(false)

			// 複数回GetImageを呼び出しても同じ画像が返される
			img1 := cl.GetImage()
			img2 := cl.GetImage()
			img3 := cl.GetImage()

			return img1 == img2 && img2 == img3 && img1 != nil
		},
		gen.IntRange(1, 64),
		gen.IntRange(1, 64),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty5_CacheInvalidatedOnChange は内容が変更されたときにキャッシュが無効化されることをテストする
// 要件 5.3: 内容が変更されたときはキャッシュを無効化する
func TestProperty5_CacheInvalidatedOnChange(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 5.3.1: BackgroundLayerのSetImageでキャッシュが更新される
	properties.Property("BackgroundLayerのSetImageでキャッシュが更新される", prop.ForAll(
		func(picID int) bool {
			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()

			img1 := ebiten.NewImage(64, 64)
			bg := NewBackgroundLayer(layerID, picID, img1)

			// 最初の画像を確認
			if bg.GetImage() != img1 {
				return false
			}

			// 新しい画像を設定
			img2 := ebiten.NewImage(128, 128)
			bg.SetImage(img2)

			// キャッシュが更新される
			if bg.GetImage() != img2 {
				return false
			}

			// ダーティフラグが設定される
			return bg.IsDirty()
		},
		gen.IntRange(0, 255),
	))

	// Property 5.3.2: DrawingLayerのSetImageでキャッシュが更新される
	properties.Property("DrawingLayerのSetImageでキャッシュが更新される", prop.ForAll(
		func(width, height int) bool {
			if width <= 0 || height <= 0 || width > 256 || height > 256 {
				return true
			}

			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()

			dl := NewDrawingLayer(layerID, 0, width, height)
			originalImg := dl.GetImage()

			// 新しい画像を設定
			newImg := ebiten.NewImage(width*2, height*2)
			dl.SetImage(newImg)

			// キャッシュが更新される
			if dl.GetImage() != newImg {
				return false
			}

			// 元の画像とは異なる
			if dl.GetImage() == originalImg {
				return false
			}

			// ダーティフラグが設定される
			return dl.IsDirty()
		},
		gen.IntRange(1, 128),
		gen.IntRange(1, 128),
	))

	// Property 5.3.3: TextLayerEntryのSetTextでキャッシュが無効化される
	properties.Property("TextLayerEntryのSetTextでキャッシュが無効化される", prop.ForAll(
		func(textIndex1, textIndex2 int) bool {
			texts := []string{"test1", "test2", "test3", "hello", "world"}
			text1 := texts[textIndex1%len(texts)]
			text2 := texts[textIndex2%len(texts)]

			if text1 == text2 {
				return true // 同じテキストの場合はスキップ
			}

			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()

			tl := NewTextLayerEntry(layerID, 0, 0, 0, text1, 0)

			// 画像を設定
			img := ebiten.NewImage(100, 20)
			tl.SetImage(img)

			// ダーティフラグをクリア
			if tl.IsDirty() {
				return false
			}

			// テキストを変更
			tl.SetText(text2)

			// キャッシュが無効化される（nilになる）
			if tl.GetImage() != nil {
				return false
			}

			// ダーティフラグが設定される
			return tl.IsDirty()
		},
		gen.IntRange(0, 4),
		gen.IntRange(0, 4),
	))

	// Property 5.3.4: CastLayerのSetSourceRectでダーティフラグが設定される
	// 注: CastLayerのGetImageはソース画像がある場合にキャッシュを再生成するため、
	// キャッシュの無効化はダーティフラグで確認する
	properties.Property("CastLayerのSetSourceRectでダーティフラグが設定される", prop.ForAll(
		func(srcX1, srcY1, w1, h1, srcX2, srcY2, w2, h2 int) bool {
			if w1 <= 0 || h1 <= 0 || w2 <= 0 || h2 <= 0 {
				return true
			}
			if w1 > 64 || h1 > 64 || w2 > 64 || h2 > 64 {
				return true
			}

			// 同じ値の場合はスキップ
			if srcX1 == srcX2 && srcY1 == srcY2 && w1 == w2 && h1 == h2 {
				return true
			}

			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()

			cl := NewCastLayer(layerID, 0, 0, 0, 0, 0, srcX1, srcY1, w1, h1, 0)

			// ソース画像を設定
			srcImg := ebiten.NewImage(256, 256)
			cl.SetSourceImage(srcImg)

			// キャッシュが生成されていることを確認
			originalCache := cl.GetImage()
			if originalCache == nil {
				return false
			}

			// ダーティフラグをクリア
			cl.SetDirty(false)

			// ソース領域を変更
			cl.SetSourceRect(srcX2, srcY2, w2, h2)

			// ダーティフラグが設定される（キャッシュ無効化の指標）
			return cl.IsDirty()
		},
		gen.IntRange(0, 50), gen.IntRange(0, 50), gen.IntRange(1, 64), gen.IntRange(1, 64),
		gen.IntRange(0, 50), gen.IntRange(0, 50), gen.IntRange(1, 64), gen.IntRange(1, 64),
	))

	// Property 5.3.5: Invalidateでキャッシュが無効化される
	properties.Property("Invalidateでキャッシュが無効化される", prop.ForAll(
		func(layerType int) bool {
			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()

			switch layerType % 4 {
			case 0:
				// BackgroundLayer
				img := ebiten.NewImage(64, 64)
				bg := NewBackgroundLayer(layerID, 0, img)
				bg.SetDirty(false)
				bg.Invalidate()
				// BackgroundLayerのInvalidateはダーティフラグを設定するが画像は保持
				return bg.IsDirty()

			case 1:
				// DrawingLayer
				dl := NewDrawingLayer(layerID, 0, 64, 64)
				dl.SetDirty(false)
				dl.Invalidate()
				// DrawingLayerのInvalidateはダーティフラグを設定するが画像は保持
				return dl.IsDirty()

			case 2:
				// TextLayerEntry
				tl := NewTextLayerEntry(layerID, 0, 0, 0, "test", 0)
				img := ebiten.NewImage(100, 20)
				tl.SetImage(img)
				tl.Invalidate()
				// TextLayerEntryのInvalidateはキャッシュをnilにする
				return tl.IsDirty() && tl.GetImage() == nil

			case 3:
				// CastLayer
				cl := NewCastLayer(layerID, 0, 0, 0, 0, 0, 0, 0, 32, 32, 0)
				srcImg := ebiten.NewImage(64, 64)
				cl.SetSourceImage(srcImg)
				cl.SetDirty(false)
				cl.Invalidate()
				// CastLayerのInvalidateはキャッシュをnilにする
				return cl.IsDirty()
			}

			return true
		},
		gen.IntRange(0, 100),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty5_CacheConsistencyAcrossOperations はキャッシュの一貫性をテストする
// 複合的なシナリオでキャッシュが正しく管理されることを確認
func TestProperty5_CacheConsistencyAcrossOperations(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 5.4: 複数の操作後もキャッシュの一貫性が保たれる
	properties.Property("複数の操作後もキャッシュの一貫性が保たれる（BackgroundLayer）", prop.ForAll(
		func(operations []int) bool {
			if len(operations) == 0 || len(operations) > 20 {
				return true
			}

			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()

			img := ebiten.NewImage(64, 64)
			bg := NewBackgroundLayer(layerID, 0, img)

			var currentImg *ebiten.Image = img

			for _, op := range operations {
				switch op % 3 {
				case 0:
					// 新しい画像を設定
					newImg := ebiten.NewImage(64, 64)
					bg.SetImage(newImg)
					currentImg = newImg
				case 1:
					// ダーティフラグをクリア
					bg.SetDirty(false)
				case 2:
					// Invalidate
					bg.Invalidate()
				}
			}

			// 現在の画像が正しいことを確認
			return bg.GetImage() == currentImg
		},
		gen.SliceOfN(20, gen.IntRange(0, 100)),
	))

	// Property 5.5: TextLayerEntryのキャッシュ管理が正しい
	// 注: SetTextは同じテキストの場合はキャッシュを無効化しない
	properties.Property("TextLayerEntryのキャッシュ管理が正しい", prop.ForAll(
		func(operations []int) bool {
			if len(operations) == 0 || len(operations) > 20 {
				return true
			}

			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()

			tl := NewTextLayerEntry(layerID, 0, 0, 0, "test", 0)

			var hasImage bool = false // 初期状態では画像はnil
			currentText := "test"

			for _, op := range operations {
				switch op % 4 {
				case 0:
					// 画像を設定
					img := ebiten.NewImage(100, 20)
					tl.SetImage(img)
					hasImage = true
				case 1:
					// テキストを変更（異なるテキストの場合のみキャッシュが無効化される）
					newText := "new text"
					if currentText != newText {
						tl.SetText(newText)
						currentText = newText
						hasImage = false // キャッシュが無効化される
					}
				case 2:
					// Invalidate（キャッシュが無効化される）
					tl.Invalidate()
					hasImage = false
				case 3:
					// ダーティフラグをクリア（キャッシュには影響しない）
					tl.SetDirty(false)
				}
			}

			// キャッシュの状態が期待通りであることを確認
			if hasImage {
				return tl.GetImage() != nil
			}
			return tl.GetImage() == nil
		},
		gen.SliceOfN(20, gen.IntRange(0, 100)),
	))

	// Property 5.6: CastLayerのキャッシュ再生成が正しく動作する
	properties.Property("CastLayerのキャッシュ再生成が正しく動作する", prop.ForAll(
		func(width, height int) bool {
			if width <= 0 || height <= 0 || width > 64 || height > 64 {
				return true
			}

			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()

			cl := NewCastLayer(layerID, 0, 0, 0, 0, 0, 0, 0, width, height, 0)

			// ソース画像を設定
			srcImg := ebiten.NewImage(128, 128)
			cl.SetSourceImage(srcImg)

			// キャッシュが生成される
			cache1 := cl.GetImage()
			if cache1 == nil {
				return false
			}

			// ダーティフラグをクリア
			cl.SetDirty(false)

			// 同じ画像が返される
			cache2 := cl.GetImage()
			if cache1 != cache2 {
				return false
			}

			// Invalidateでキャッシュを無効化
			cl.Invalidate()

			// GetImageで新しいキャッシュが生成される
			cache3 := cl.GetImage()
			if cache3 == nil {
				return false
			}

			// ダーティフラグがクリアされる
			return !cl.IsDirty()
		},
		gen.IntRange(1, 64),
		gen.IntRange(1, 64),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty5_PictureLayerSetCacheManagement はPictureLayerSetのキャッシュ管理をテストする
func TestProperty5_PictureLayerSetCacheManagement(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 5.7: CompositeBufferのキャッシュ管理が正しい
	properties.Property("CompositeBufferのキャッシュ管理が正しい", prop.ForAll(
		func(width, height int) bool {
			if width <= 0 || height <= 0 || width > 256 || height > 256 {
				return true
			}

			lm := NewLayerManager()
			pls := lm.GetOrCreatePictureLayerSet(0)

			// 初期状態ではCompositeBufferはnil
			if pls.GetCompositeBuffer() != nil {
				return false
			}

			// CompositeBufferを設定
			buffer := ebiten.NewImage(width, height)
			pls.SetCompositeBuffer(buffer)

			// 設定した画像が返される
			if pls.GetCompositeBuffer() != buffer {
				return false
			}

			// 新しいバッファを設定
			newBuffer := ebiten.NewImage(width*2, height*2)
			pls.SetCompositeBuffer(newBuffer)

			// 新しい画像が返される
			return pls.GetCompositeBuffer() == newBuffer
		},
		gen.IntRange(1, 256),
		gen.IntRange(1, 256),
	))

	// Property 5.8: レイヤー追加時にFullDirtyが設定される
	properties.Property("レイヤー追加時にFullDirtyが設定される", prop.ForAll(
		func(castCount, textCount int) bool {
			if castCount < 0 || castCount > 20 || textCount < 0 || textCount > 20 {
				return true
			}

			lm := NewLayerManager()
			pls := lm.GetOrCreatePictureLayerSet(0)

			// 初期状態はFullDirty
			if !pls.FullDirty {
				return false
			}

			// ダーティフラグをクリア
			pls.ClearAllDirtyFlags()
			if pls.FullDirty {
				return false
			}

			// キャストレイヤーを追加
			for i := range castCount {
				layerID := lm.GetNextLayerID()
				zOrderOffset := pls.GetNextCastZOffset()
				cl := NewCastLayer(layerID, i, 0, 0, i*10, i*10, 0, 0, 32, 32, zOrderOffset)
				pls.AddCastLayer(cl)

				// 追加後はFullDirtyが設定される
				if !pls.FullDirty {
					return false
				}

				pls.ClearAllDirtyFlags()
			}

			// テキストレイヤーを追加
			for i := range textCount {
				layerID := lm.GetNextLayerID()
				zOrderOffset := pls.GetNextTextZOffset()
				tl := NewTextLayerEntry(layerID, 0, i*10, i*10, "test", zOrderOffset)
				pls.AddTextLayer(tl)

				// 追加後はFullDirtyが設定される
				if !pls.FullDirty {
					return false
				}

				pls.ClearAllDirtyFlags()
			}

			return true
		},
		gen.IntRange(0, 20),
		gen.IntRange(0, 20),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty5_CacheInvalidationOnLayerRemoval はレイヤー削除時のキャッシュ無効化をテストする
func TestProperty5_CacheInvalidationOnLayerRemoval(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 5.9: キャストレイヤー削除時にFullDirtyが設定される
	properties.Property("キャストレイヤー削除時にFullDirtyが設定される", prop.ForAll(
		func(castCount int) bool {
			if castCount <= 0 || castCount > 20 {
				return true
			}

			lm := NewLayerManager()
			pls := lm.GetOrCreatePictureLayerSet(0)

			// キャストレイヤーを追加
			var castIDs []int
			for i := range castCount {
				layerID := lm.GetNextLayerID()
				zOrderOffset := pls.GetNextCastZOffset()
				cl := NewCastLayer(layerID, i, 0, 0, i*10, i*10, 0, 0, 32, 32, zOrderOffset)
				pls.AddCastLayer(cl)
				castIDs = append(castIDs, i)
			}

			// ダーティフラグをクリア
			pls.ClearAllDirtyFlags()

			// キャストレイヤーを削除
			for _, castID := range castIDs {
				if !pls.RemoveCastLayer(castID) {
					continue // 既に削除されている場合
				}

				// 削除後はFullDirtyが設定される
				if !pls.FullDirty {
					return false
				}

				pls.ClearAllDirtyFlags()
			}

			return true
		},
		gen.IntRange(1, 20),
	))

	// Property 5.10: テキストレイヤー削除時にFullDirtyが設定される
	properties.Property("テキストレイヤー削除時にFullDirtyが設定される", prop.ForAll(
		func(textCount int) bool {
			if textCount <= 0 || textCount > 20 {
				return true
			}

			lm := NewLayerManager()
			pls := lm.GetOrCreatePictureLayerSet(0)

			// テキストレイヤーを追加
			var layerIDs []int
			for i := range textCount {
				layerID := lm.GetNextLayerID()
				zOrderOffset := pls.GetNextTextZOffset()
				tl := NewTextLayerEntry(layerID, 0, i*10, i*10, "test", zOrderOffset)
				pls.AddTextLayer(tl)
				layerIDs = append(layerIDs, layerID)
			}

			// ダーティフラグをクリア
			pls.ClearAllDirtyFlags()

			// テキストレイヤーを削除
			for _, layerID := range layerIDs {
				if !pls.RemoveTextLayer(layerID) {
					continue // 既に削除されている場合
				}

				// 削除後はFullDirtyが設定される
				if !pls.FullDirty {
					return false
				}

				pls.ClearAllDirtyFlags()
			}

			return true
		},
		gen.IntRange(1, 20),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// ============================================================================
// Feature: layer-based-rendering, Property 4: 可視領域クリッピングの正確性
// **Validates: Requirements 4.1, 4.2, 4.3, 4.4**
//
// 任意のレイヤーと可視領域に対して、レイヤーが可視領域外にある場合は描画がスキップされ、
// 部分的に可視な場合は可視部分のみが描画される。
// ============================================================================

// TestProperty4_LayerCompletelyOutsideVisibleRegion はレイヤーが可視領域外にある場合をテストする
// 要件 4.1: レイヤーがウィンドウの可視領域外にあるときに描画をスキップする
func TestProperty4_LayerCompletelyOutsideVisibleRegion(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 4.1.1: レイヤーが可視領域の左側にある場合、IsLayerVisibleはfalseを返す
	properties.Property("レイヤーが可視領域の左側にある場合、IsLayerVisibleはfalseを返す", prop.ForAll(
		func(layerX, layerY, layerW, layerH, visibleX, visibleW, visibleH int) bool {
			if layerW <= 0 || layerH <= 0 || visibleW <= 0 || visibleH <= 0 {
				return true
			}

			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()

			// レイヤーを可視領域の左側に配置（レイヤーの右端が可視領域の左端より左）
			actualLayerX := visibleX - layerW - 10 // 可視領域の左側に配置
			cl := NewCastLayer(layerID, 0, 0, 0, actualLayerX, layerY, 0, 0, layerW, layerH, 0)

			visibleRect := image.Rect(visibleX, 0, visibleX+visibleW, visibleH)

			return !IsLayerVisible(cl, visibleRect)
		},
		gen.IntRange(0, 100),
		gen.IntRange(0, 100),
		gen.IntRange(1, 64),
		gen.IntRange(1, 64),
		gen.IntRange(100, 200),
		gen.IntRange(100, 300),
		gen.IntRange(100, 300),
	))

	// Property 4.1.2: レイヤーが可視領域の右側にある場合、IsLayerVisibleはfalseを返す
	properties.Property("レイヤーが可視領域の右側にある場合、IsLayerVisibleはfalseを返す", prop.ForAll(
		func(layerY, layerW, layerH, visibleX, visibleW, visibleH int) bool {
			if layerW <= 0 || layerH <= 0 || visibleW <= 0 || visibleH <= 0 {
				return true
			}

			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()

			// レイヤーを可視領域の右側に配置（レイヤーの左端が可視領域の右端より右）
			actualLayerX := visibleX + visibleW + 10 // 可視領域の右側に配置
			cl := NewCastLayer(layerID, 0, 0, 0, actualLayerX, layerY, 0, 0, layerW, layerH, 0)

			visibleRect := image.Rect(visibleX, 0, visibleX+visibleW, visibleH)

			return !IsLayerVisible(cl, visibleRect)
		},
		gen.IntRange(0, 100),
		gen.IntRange(1, 64),
		gen.IntRange(1, 64),
		gen.IntRange(0, 100),
		gen.IntRange(100, 300),
		gen.IntRange(100, 300),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty4_LayerCompletelyOutsideVisibleRegion_TopBottom はレイヤーが上下に可視領域外にある場合をテストする
func TestProperty4_LayerCompletelyOutsideVisibleRegion_TopBottom(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 4.1.3: レイヤーが可視領域の上側にある場合、IsLayerVisibleはfalseを返す
	properties.Property("レイヤーが可視領域の上側にある場合、IsLayerVisibleはfalseを返す", prop.ForAll(
		func(layerX, layerW, layerH, visibleY, visibleW, visibleH int) bool {
			if layerW <= 0 || layerH <= 0 || visibleW <= 0 || visibleH <= 0 {
				return true
			}

			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()

			// レイヤーを可視領域の上側に配置（レイヤーの下端が可視領域の上端より上）
			actualLayerY := visibleY - layerH - 10 // 可視領域の上側に配置
			cl := NewCastLayer(layerID, 0, 0, 0, layerX, actualLayerY, 0, 0, layerW, layerH, 0)

			visibleRect := image.Rect(0, visibleY, visibleW, visibleY+visibleH)

			return !IsLayerVisible(cl, visibleRect)
		},
		gen.IntRange(0, 100),
		gen.IntRange(1, 64),
		gen.IntRange(1, 64),
		gen.IntRange(100, 200),
		gen.IntRange(100, 300),
		gen.IntRange(100, 300),
	))

	// Property 4.1.4: レイヤーが可視領域の下側にある場合、IsLayerVisibleはfalseを返す
	properties.Property("レイヤーが可視領域の下側にある場合、IsLayerVisibleはfalseを返す", prop.ForAll(
		func(layerX, layerW, layerH, visibleY, visibleW, visibleH int) bool {
			if layerW <= 0 || layerH <= 0 || visibleW <= 0 || visibleH <= 0 {
				return true
			}

			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()

			// レイヤーを可視領域の下側に配置（レイヤーの上端が可視領域の下端より下）
			actualLayerY := visibleY + visibleH + 10 // 可視領域の下側に配置
			cl := NewCastLayer(layerID, 0, 0, 0, layerX, actualLayerY, 0, 0, layerW, layerH, 0)

			visibleRect := image.Rect(0, visibleY, visibleW, visibleY+visibleH)

			return !IsLayerVisible(cl, visibleRect)
		},
		gen.IntRange(0, 100),
		gen.IntRange(1, 64),
		gen.IntRange(1, 64),
		gen.IntRange(0, 100),
		gen.IntRange(100, 300),
		gen.IntRange(100, 300),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty4_LayerCompletelyInsideVisibleRegion はレイヤーが可視領域内に完全に含まれる場合をテストする
// 要件 4.2: レイヤーが部分的に可視領域内にあるときに可視部分のみを描画する（完全に内側の場合）
func TestProperty4_LayerCompletelyInsideVisibleRegion(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 4.2.1: レイヤーが可視領域内に完全に含まれる場合、IsLayerVisibleはtrueを返す
	properties.Property("レイヤーが可視領域内に完全に含まれる場合、IsLayerVisibleはtrueを返す", prop.ForAll(
		func(layerX, layerY, layerW, layerH, margin int) bool {
			if layerW <= 0 || layerH <= 0 || margin < 0 {
				return true
			}

			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()

			// レイヤーを作成
			cl := NewCastLayer(layerID, 0, 0, 0, layerX, layerY, 0, 0, layerW, layerH, 0)

			// 可視領域をレイヤーより大きく設定（レイヤーを完全に含む）
			visibleRect := image.Rect(
				layerX-margin,
				layerY-margin,
				layerX+layerW+margin,
				layerY+layerH+margin,
			)

			return IsLayerVisible(cl, visibleRect)
		},
		gen.IntRange(0, 100),
		gen.IntRange(0, 100),
		gen.IntRange(1, 64),
		gen.IntRange(1, 64),
		gen.IntRange(1, 50),
	))

	// Property 4.2.2: レイヤーが可視領域内に完全に含まれる場合、GetVisibleRegionはレイヤーの境界を返す
	properties.Property("レイヤーが可視領域内に完全に含まれる場合、GetVisibleRegionはレイヤーの境界を返す", prop.ForAll(
		func(layerX, layerY, layerW, layerH, margin int) bool {
			if layerW <= 0 || layerH <= 0 || margin < 0 {
				return true
			}

			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()

			// レイヤーを作成
			cl := NewCastLayer(layerID, 0, 0, 0, layerX, layerY, 0, 0, layerW, layerH, 0)

			// 可視領域をレイヤーより大きく設定（レイヤーを完全に含む）
			visibleRect := image.Rect(
				layerX-margin,
				layerY-margin,
				layerX+layerW+margin,
				layerY+layerH+margin,
			)

			visibleRegion := GetVisibleRegion(cl, visibleRect)
			layerBounds := cl.GetBounds()

			// 可視領域はレイヤーの境界と一致するはず
			return visibleRegion == layerBounds
		},
		gen.IntRange(0, 100),
		gen.IntRange(0, 100),
		gen.IntRange(1, 64),
		gen.IntRange(1, 64),
		gen.IntRange(1, 50),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty4_LayerPartiallyInsideVisibleRegion はレイヤーが部分的に可視領域内にある場合をテストする
// 要件 4.2: レイヤーが部分的に可視領域内にあるときに可視部分のみを描画する
func TestProperty4_LayerPartiallyInsideVisibleRegion(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 4.3.1: レイヤーが左側から部分的にはみ出している場合
	properties.Property("レイヤーが左側から部分的にはみ出している場合、IsLayerVisibleはtrueを返す", prop.ForAll(
		func(layerW, layerH, visibleW, visibleH, overlap int) bool {
			if layerW <= 0 || layerH <= 0 || visibleW <= 0 || visibleH <= 0 || overlap <= 0 {
				return true
			}
			if overlap >= layerW {
				return true // 完全に内側の場合はスキップ
			}

			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()

			// レイヤーを可視領域の左側からはみ出すように配置
			layerX := -layerW + overlap // overlap分だけ可視領域内に入る
			cl := NewCastLayer(layerID, 0, 0, 0, layerX, 0, 0, 0, layerW, layerH, 0)

			visibleRect := image.Rect(0, 0, visibleW, visibleH)

			return IsLayerVisible(cl, visibleRect)
		},
		gen.IntRange(10, 100),
		gen.IntRange(10, 100),
		gen.IntRange(100, 300),
		gen.IntRange(100, 300),
		gen.IntRange(1, 50),
	))

	// Property 4.3.2: レイヤーが左側から部分的にはみ出している場合、GetVisibleRegionは交差部分を返す
	properties.Property("レイヤーが左側から部分的にはみ出している場合、GetVisibleRegionは交差部分を返す", prop.ForAll(
		func(layerW, layerH, visibleW, visibleH, overlap int) bool {
			if layerW <= 0 || layerH <= 0 || visibleW <= 0 || visibleH <= 0 || overlap <= 0 {
				return true
			}
			if overlap >= layerW {
				return true
			}

			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()

			layerX := -layerW + overlap
			cl := NewCastLayer(layerID, 0, 0, 0, layerX, 0, 0, 0, layerW, layerH, 0)

			visibleRect := image.Rect(0, 0, visibleW, visibleH)
			visibleRegion := GetVisibleRegion(cl, visibleRect)

			// 期待される交差領域
			expectedMinY := 0
			expectedMaxY := layerH
			if expectedMaxY > visibleH {
				expectedMaxY = visibleH
			}
			expected := image.Rect(0, expectedMinY, overlap, expectedMaxY)

			return visibleRegion == expected
		},
		gen.IntRange(10, 100),
		gen.IntRange(10, 100),
		gen.IntRange(100, 300),
		gen.IntRange(100, 300),
		gen.IntRange(1, 50),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty4_LayerPartiallyInsideVisibleRegion_RightBottom は右側・下側からはみ出す場合をテストする
func TestProperty4_LayerPartiallyInsideVisibleRegion_RightBottom(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 4.3.3: レイヤーが右側から部分的にはみ出している場合
	properties.Property("レイヤーが右側から部分的にはみ出している場合、IsLayerVisibleはtrueを返す", prop.ForAll(
		func(layerW, layerH, visibleW, visibleH, overlap int) bool {
			if layerW <= 0 || layerH <= 0 || visibleW <= 0 || visibleH <= 0 || overlap <= 0 {
				return true
			}
			if overlap >= layerW {
				return true
			}

			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()

			// レイヤーを可視領域の右側からはみ出すように配置
			layerX := visibleW - overlap // overlap分だけ可視領域内に入る
			cl := NewCastLayer(layerID, 0, 0, 0, layerX, 0, 0, 0, layerW, layerH, 0)

			visibleRect := image.Rect(0, 0, visibleW, visibleH)

			return IsLayerVisible(cl, visibleRect)
		},
		gen.IntRange(10, 100),
		gen.IntRange(10, 100),
		gen.IntRange(100, 300),
		gen.IntRange(100, 300),
		gen.IntRange(1, 50),
	))

	// Property 4.3.4: レイヤーが下側から部分的にはみ出している場合
	properties.Property("レイヤーが下側から部分的にはみ出している場合、IsLayerVisibleはtrueを返す", prop.ForAll(
		func(layerW, layerH, visibleW, visibleH, overlap int) bool {
			if layerW <= 0 || layerH <= 0 || visibleW <= 0 || visibleH <= 0 || overlap <= 0 {
				return true
			}
			if overlap >= layerH {
				return true
			}

			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()

			// レイヤーを可視領域の下側からはみ出すように配置
			layerY := visibleH - overlap
			cl := NewCastLayer(layerID, 0, 0, 0, 0, layerY, 0, 0, layerW, layerH, 0)

			visibleRect := image.Rect(0, 0, visibleW, visibleH)

			return IsLayerVisible(cl, visibleRect)
		},
		gen.IntRange(10, 100),
		gen.IntRange(10, 100),
		gen.IntRange(100, 300),
		gen.IntRange(100, 300),
		gen.IntRange(1, 50),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty4_EmptyVisibleRegion は空の可視領域の場合をテストする
// 要件 4.4: 可視領域との交差判定を行う
func TestProperty4_EmptyVisibleRegion(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 4.4.1: 空の可視領域の場合、IsLayerVisibleはfalseを返す
	properties.Property("空の可視領域の場合、IsLayerVisibleはfalseを返す", prop.ForAll(
		func(layerX, layerY, layerW, layerH int) bool {
			if layerW <= 0 || layerH <= 0 {
				return true
			}

			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()

			cl := NewCastLayer(layerID, 0, 0, 0, layerX, layerY, 0, 0, layerW, layerH, 0)

			// 空の可視領域
			emptyRect := image.Rectangle{}

			return !IsLayerVisible(cl, emptyRect)
		},
		gen.IntRange(0, 100),
		gen.IntRange(0, 100),
		gen.IntRange(1, 64),
		gen.IntRange(1, 64),
	))

	// Property 4.4.2: 空の可視領域の場合、GetVisibleRegionは空の矩形を返す
	properties.Property("空の可視領域の場合、GetVisibleRegionは空の矩形を返す", prop.ForAll(
		func(layerX, layerY, layerW, layerH int) bool {
			if layerW <= 0 || layerH <= 0 {
				return true
			}

			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()

			cl := NewCastLayer(layerID, 0, 0, 0, layerX, layerY, 0, 0, layerW, layerH, 0)

			emptyRect := image.Rectangle{}
			visibleRegion := GetVisibleRegion(cl, emptyRect)

			return visibleRegion.Empty()
		},
		gen.IntRange(0, 100),
		gen.IntRange(0, 100),
		gen.IntRange(1, 64),
		gen.IntRange(1, 64),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty4_EmptyLayerBounds は空のレイヤー境界の場合をテストする
// 要件 4.3: 各レイヤーの境界ボックスを計算する
func TestProperty4_EmptyLayerBounds(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 4.5.1: 幅が0のレイヤーの場合、IsLayerVisibleはfalseを返す
	properties.Property("幅が0のレイヤーの場合、IsLayerVisibleはfalseを返す", prop.ForAll(
		func(layerX, layerY, layerH, visibleW, visibleH int) bool {
			if layerH <= 0 || visibleW <= 0 || visibleH <= 0 {
				return true
			}

			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()

			// 幅が0のレイヤー
			cl := NewCastLayer(layerID, 0, 0, 0, layerX, layerY, 0, 0, 0, layerH, 0)

			visibleRect := image.Rect(0, 0, visibleW, visibleH)

			return !IsLayerVisible(cl, visibleRect)
		},
		gen.IntRange(0, 100),
		gen.IntRange(0, 100),
		gen.IntRange(1, 64),
		gen.IntRange(100, 300),
		gen.IntRange(100, 300),
	))

	// Property 4.5.2: 高さが0のレイヤーの場合、IsLayerVisibleはfalseを返す
	properties.Property("高さが0のレイヤーの場合、IsLayerVisibleはfalseを返す", prop.ForAll(
		func(layerX, layerY, layerW, visibleW, visibleH int) bool {
			if layerW <= 0 || visibleW <= 0 || visibleH <= 0 {
				return true
			}

			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()

			// 高さが0のレイヤー
			cl := NewCastLayer(layerID, 0, 0, 0, layerX, layerY, 0, 0, layerW, 0, 0)

			visibleRect := image.Rect(0, 0, visibleW, visibleH)

			return !IsLayerVisible(cl, visibleRect)
		},
		gen.IntRange(0, 100),
		gen.IntRange(0, 100),
		gen.IntRange(1, 64),
		gen.IntRange(100, 300),
		gen.IntRange(100, 300),
	))

	// Property 4.5.3: nilレイヤーの場合、IsLayerVisibleはfalseを返す
	properties.Property("nilレイヤーの場合、IsLayerVisibleはfalseを返す", prop.ForAll(
		func(visibleW, visibleH int) bool {
			if visibleW <= 0 || visibleH <= 0 {
				return true
			}

			visibleRect := image.Rect(0, 0, visibleW, visibleH)

			return !IsLayerVisible(nil, visibleRect)
		},
		gen.IntRange(100, 300),
		gen.IntRange(100, 300),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty4_InvisibleLayer は非表示レイヤーの場合をテストする
// 要件 4.1: レイヤーがウィンドウの可視領域外にあるときに描画をスキップする
func TestProperty4_InvisibleLayer(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 4.6.1: 非表示のレイヤーの場合、IsLayerVisibleはfalseを返す
	properties.Property("非表示のレイヤーの場合、IsLayerVisibleはfalseを返す", prop.ForAll(
		func(layerX, layerY, layerW, layerH, visibleW, visibleH int) bool {
			if layerW <= 0 || layerH <= 0 || visibleW <= 0 || visibleH <= 0 {
				return true
			}

			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()

			// レイヤーを可視領域内に配置
			cl := NewCastLayer(layerID, 0, 0, 0, layerX%visibleW, layerY%visibleH, 0, 0, layerW, layerH, 0)

			// レイヤーを非表示に設定
			cl.SetVisible(false)

			visibleRect := image.Rect(0, 0, visibleW, visibleH)

			return !IsLayerVisible(cl, visibleRect)
		},
		gen.IntRange(0, 100),
		gen.IntRange(0, 100),
		gen.IntRange(1, 64),
		gen.IntRange(1, 64),
		gen.IntRange(100, 300),
		gen.IntRange(100, 300),
	))

	// Property 4.6.2: 表示状態のレイヤーが可視領域内にある場合、IsLayerVisibleはtrueを返す
	properties.Property("表示状態のレイヤーが可視領域内にある場合、IsLayerVisibleはtrueを返す", prop.ForAll(
		func(layerX, layerY, layerW, layerH, margin int) bool {
			if layerW <= 0 || layerH <= 0 || margin < 0 {
				return true
			}

			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()

			cl := NewCastLayer(layerID, 0, 0, 0, layerX, layerY, 0, 0, layerW, layerH, 0)

			// レイヤーは初期状態で表示
			if !cl.IsVisible() {
				return false
			}

			// 可視領域をレイヤーを含むように設定
			visibleRect := image.Rect(
				layerX-margin,
				layerY-margin,
				layerX+layerW+margin,
				layerY+layerH+margin,
			)

			return IsLayerVisible(cl, visibleRect)
		},
		gen.IntRange(0, 100),
		gen.IntRange(0, 100),
		gen.IntRange(1, 64),
		gen.IntRange(1, 64),
		gen.IntRange(1, 50),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty4_BoundaryBoxCalculation はレイヤーの境界ボックス計算をテストする
// 要件 4.3: 各レイヤーの境界ボックスを計算する
func TestProperty4_BoundaryBoxCalculation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 4.7.1: CastLayerの境界ボックスが正しく計算される
	properties.Property("CastLayerの境界ボックスが正しく計算される", prop.ForAll(
		func(layerX, layerY, layerW, layerH int) bool {
			if layerW <= 0 || layerH <= 0 {
				return true
			}

			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()

			cl := NewCastLayer(layerID, 0, 0, 0, layerX, layerY, 0, 0, layerW, layerH, 0)

			bounds := cl.GetBounds()
			expected := image.Rect(layerX, layerY, layerX+layerW, layerY+layerH)

			return bounds == expected
		},
		gen.IntRange(-100, 100),
		gen.IntRange(-100, 100),
		gen.IntRange(1, 128),
		gen.IntRange(1, 128),
	))

	// Property 4.7.2: TextLayerEntryの境界ボックスが正しく計算される（画像設定後）
	properties.Property("TextLayerEntryの境界ボックスが正しく計算される（画像設定後）", prop.ForAll(
		func(layerX, layerY, imgW, imgH int) bool {
			if imgW <= 0 || imgH <= 0 {
				return true
			}

			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()

			tl := NewTextLayerEntry(layerID, 0, layerX, layerY, "test", 0)

			// 画像を設定（TextLayerEntryは画像設定後に境界が計算される）
			img := ebiten.NewImage(imgW, imgH)
			tl.SetImage(img)

			bounds := tl.GetBounds()

			// 境界の位置とサイズが正しいことを確認
			return bounds.Min.X == layerX && bounds.Min.Y == layerY &&
				bounds.Dx() == imgW && bounds.Dy() == imgH
		},
		gen.IntRange(-100, 100),
		gen.IntRange(-100, 100),
		gen.IntRange(1, 128),
		gen.IntRange(1, 128),
	))

	// Property 4.7.3: BackgroundLayerの境界ボックスが正しく計算される
	properties.Property("BackgroundLayerの境界ボックスが正しく計算される", prop.ForAll(
		func(width, height int) bool {
			if width <= 0 || height <= 0 {
				return true
			}

			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()

			img := ebiten.NewImage(width, height)
			bg := NewBackgroundLayer(layerID, 0, img)

			bounds := bg.GetBounds()

			// 背景レイヤーの境界は画像サイズと一致
			return bounds.Dx() == width && bounds.Dy() == height
		},
		gen.IntRange(1, 256),
		gen.IntRange(1, 256),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty4_IntersectionCalculation は交差領域の計算をテストする
// 要件 4.4: 可視領域との交差判定を行う
func TestProperty4_IntersectionCalculation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 4.8.1: 交差領域は常にレイヤー境界と可視領域の両方に含まれる
	properties.Property("交差領域は常にレイヤー境界と可視領域の両方に含まれる", prop.ForAll(
		func(lx, ly, lw, lh, vx, vy, vw, vh int) bool {
			if lw <= 0 || lh <= 0 || vw <= 0 || vh <= 0 {
				return true
			}

			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()

			cl := NewCastLayer(layerID, 0, 0, 0, lx, ly, 0, 0, lw, lh, 0)

			visibleRect := image.Rect(vx, vy, vx+vw, vy+vh)
			visibleRegion := GetVisibleRegion(cl, visibleRect)

			if visibleRegion.Empty() {
				return true // 交差がない場合はOK
			}

			layerBounds := cl.GetBounds()

			// 交差領域がレイヤー境界に含まれることを確認
			if !containsRect(layerBounds, visibleRegion) {
				return false
			}

			// 交差領域が可視領域に含まれることを確認
			if !containsRect(visibleRect, visibleRegion) {
				return false
			}

			return true
		},
		gen.IntRange(-50, 150),
		gen.IntRange(-50, 150),
		gen.IntRange(1, 100),
		gen.IntRange(1, 100),
		gen.IntRange(0, 100),
		gen.IntRange(0, 100),
		gen.IntRange(50, 200),
		gen.IntRange(50, 200),
	))

	// Property 4.8.2: 交差領域が空でない場合、IsLayerVisibleはtrueを返す
	properties.Property("交差領域が空でない場合、IsLayerVisibleはtrueを返す", prop.ForAll(
		func(lx, ly, lw, lh, vx, vy, vw, vh int) bool {
			if lw <= 0 || lh <= 0 || vw <= 0 || vh <= 0 {
				return true
			}

			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()

			cl := NewCastLayer(layerID, 0, 0, 0, lx, ly, 0, 0, lw, lh, 0)

			visibleRect := image.Rect(vx, vy, vx+vw, vy+vh)
			visibleRegion := GetVisibleRegion(cl, visibleRect)
			isVisible := IsLayerVisible(cl, visibleRect)

			// 交差領域が空でない場合、IsLayerVisibleはtrueを返すべき
			if !visibleRegion.Empty() {
				return isVisible
			}

			// 交差領域が空の場合、IsLayerVisibleはfalseを返すべき
			return !isVisible
		},
		gen.IntRange(-50, 150),
		gen.IntRange(-50, 150),
		gen.IntRange(1, 100),
		gen.IntRange(1, 100),
		gen.IntRange(0, 100),
		gen.IntRange(0, 100),
		gen.IntRange(50, 200),
		gen.IntRange(50, 200),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty4_MultipleLayerTypes は複数のレイヤータイプでの可視領域クリッピングをテストする
func TestProperty4_MultipleLayerTypes(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 4.9.1: BackgroundLayerの可視領域クリッピングが正しく動作する
	properties.Property("BackgroundLayerの可視領域クリッピングが正しく動作する", prop.ForAll(
		func(imgW, imgH, vx, vy, vw, vh int) bool {
			if imgW <= 0 || imgH <= 0 || vw <= 0 || vh <= 0 {
				return true
			}

			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()

			img := ebiten.NewImage(imgW, imgH)
			bg := NewBackgroundLayer(layerID, 0, img)

			visibleRect := image.Rect(vx, vy, vx+vw, vy+vh)

			isVisible := IsLayerVisible(bg, visibleRect)
			visibleRegion := GetVisibleRegion(bg, visibleRect)

			// 可視性と交差領域の整合性を確認
			if isVisible {
				return !visibleRegion.Empty()
			}
			return visibleRegion.Empty()
		},
		gen.IntRange(1, 128),
		gen.IntRange(1, 128),
		gen.IntRange(-50, 100),
		gen.IntRange(-50, 100),
		gen.IntRange(50, 200),
		gen.IntRange(50, 200),
	))

	// Property 4.9.2: DrawingLayerの可視領域クリッピングが正しく動作する
	properties.Property("DrawingLayerの可視領域クリッピングが正しく動作する", prop.ForAll(
		func(dlW, dlH, vx, vy, vw, vh int) bool {
			if dlW <= 0 || dlH <= 0 || vw <= 0 || vh <= 0 {
				return true
			}

			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()

			dl := NewDrawingLayer(layerID, 0, dlW, dlH)

			visibleRect := image.Rect(vx, vy, vx+vw, vy+vh)

			isVisible := IsLayerVisible(dl, visibleRect)
			visibleRegion := GetVisibleRegion(dl, visibleRect)

			if isVisible {
				return !visibleRegion.Empty()
			}
			return visibleRegion.Empty()
		},
		gen.IntRange(1, 128),
		gen.IntRange(1, 128),
		gen.IntRange(-50, 100),
		gen.IntRange(-50, 100),
		gen.IntRange(50, 200),
		gen.IntRange(50, 200),
	))

	// Property 4.9.3: TextLayerEntryの可視領域クリッピングが正しく動作する
	properties.Property("TextLayerEntryの可視領域クリッピングが正しく動作する", prop.ForAll(
		func(tx, ty, vx, vy, vw, vh int) bool {
			if vw <= 0 || vh <= 0 {
				return true
			}

			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()

			tl := NewTextLayerEntry(layerID, 0, tx, ty, "test text", 0)

			visibleRect := image.Rect(vx, vy, vx+vw, vy+vh)

			isVisible := IsLayerVisible(tl, visibleRect)
			visibleRegion := GetVisibleRegion(tl, visibleRect)

			if isVisible {
				return !visibleRegion.Empty()
			}
			return visibleRegion.Empty()
		},
		gen.IntRange(-50, 150),
		gen.IntRange(-50, 150),
		gen.IntRange(0, 100),
		gen.IntRange(0, 100),
		gen.IntRange(50, 200),
		gen.IntRange(50, 200),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// ============================================================================
// Feature: layer-based-rendering, Property 6: ダーティ領域追跡の正確性
// **Validates: Requirements 6.1, 6.2, 6.3, 6.4**
//
// 任意のレイヤー変更に対して、ダーティ領域が正しく追跡され、
// 複数のダーティ領域は統合される。
// ============================================================================

// TestProperty6_AddDirtyRegionUpdatesField はダーティ領域の追加をテストする
// 要件 6.1: 変更があった領域（ダーティ領域）を追跡する
func TestProperty6_AddDirtyRegionUpdatesField(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 6.1.1: AddDirtyRegionでダーティ領域が設定される
	properties.Property("AddDirtyRegionでダーティ領域が設定される", prop.ForAll(
		func(x, y, w, h int) bool {
			if w <= 0 || h <= 0 {
				return true
			}

			lm := NewLayerManager()
			pls := lm.GetOrCreatePictureLayerSet(0)

			// ダーティ領域をクリア
			pls.ClearDirtyRegion()

			// ダーティ領域を追加
			rect := image.Rect(x, y, x+w, y+h)
			pls.AddDirtyRegion(rect)

			// ダーティ領域が設定されていることを確認
			return pls.DirtyRegion == rect
		},
		gen.IntRange(0, 640),
		gen.IntRange(0, 480),
		gen.IntRange(1, 100),
		gen.IntRange(1, 100),
	))

	// Property 6.1.2: 空の矩形を追加してもダーティ領域は変わらない
	properties.Property("空の矩形を追加してもダーティ領域は変わらない", prop.ForAll(
		func(x, y, w, h int) bool {
			if w <= 0 || h <= 0 {
				return true
			}

			lm := NewLayerManager()
			pls := lm.GetOrCreatePictureLayerSet(0)

			// 初期ダーティ領域を設定
			initialRect := image.Rect(x, y, x+w, y+h)
			pls.ClearDirtyRegion()
			pls.AddDirtyRegion(initialRect)

			// 空の矩形を追加
			emptyRect := image.Rectangle{}
			pls.AddDirtyRegion(emptyRect)

			// ダーティ領域が変わっていないことを確認
			return pls.DirtyRegion == initialRect
		},
		gen.IntRange(0, 640),
		gen.IntRange(0, 480),
		gen.IntRange(1, 100),
		gen.IntRange(1, 100),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty6_MultipleDirtyRegionsUnified は複数のダーティ領域の統合をテストする
// 要件 6.3: 複数のダーティ領域があるときにそれらを統合して処理する
func TestProperty6_MultipleDirtyRegionsUnified(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 6.3.1: 2つのダーティ領域がUnionで統合される
	properties.Property("2つのダーティ領域がUnionで統合される", prop.ForAll(
		func(x1, y1, w1, h1, x2, y2, w2, h2 int) bool {
			if w1 <= 0 || h1 <= 0 || w2 <= 0 || h2 <= 0 {
				return true
			}

			lm := NewLayerManager()
			pls := lm.GetOrCreatePictureLayerSet(0)

			// ダーティ領域をクリア
			pls.ClearDirtyRegion()

			// 2つのダーティ領域を追加
			rect1 := image.Rect(x1, y1, x1+w1, y1+h1)
			rect2 := image.Rect(x2, y2, x2+w2, y2+h2)
			pls.AddDirtyRegion(rect1)
			pls.AddDirtyRegion(rect2)

			// 期待される統合領域
			expectedUnion := rect1.Union(rect2)

			// ダーティ領域が統合されていることを確認
			return pls.DirtyRegion == expectedUnion
		},
		gen.IntRange(0, 300),
		gen.IntRange(0, 200),
		gen.IntRange(1, 100),
		gen.IntRange(1, 100),
		gen.IntRange(0, 300),
		gen.IntRange(0, 200),
		gen.IntRange(1, 100),
		gen.IntRange(1, 100),
	))

	// Property 6.3.2: 複数のダーティ領域が順次統合される
	properties.Property("複数のダーティ領域が順次統合される", prop.ForAll(
		func(rects []int) bool {
			if len(rects) < 4 || len(rects) > 40 {
				return true
			}

			lm := NewLayerManager()
			pls := lm.GetOrCreatePictureLayerSet(0)

			// ダーティ領域をクリア
			pls.ClearDirtyRegion()

			// 複数の矩形を追加し、期待される統合領域を計算
			var expectedUnion image.Rectangle
			for i := 0; i+3 < len(rects); i += 4 {
				x := rects[i] % 640
				y := rects[i+1] % 480
				w := (rects[i+2] % 100) + 1
				h := (rects[i+3] % 100) + 1
				rect := image.Rect(x, y, x+w, y+h)

				pls.AddDirtyRegion(rect)

				if expectedUnion.Empty() {
					expectedUnion = rect
				} else {
					expectedUnion = expectedUnion.Union(rect)
				}
			}

			// 統合された領域が期待通りであることを確認
			return pls.DirtyRegion == expectedUnion
		},
		gen.SliceOfN(40, gen.IntRange(0, 640)),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty6_EmptyDirtyRegionHandling は空のダーティ領域の処理をテストする
// 要件 6.4: ダーティ領域が空のときに再合成をスキップする
func TestProperty6_EmptyDirtyRegionHandling(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 6.4.1: ClearDirtyRegionでダーティ領域が空になる
	properties.Property("ClearDirtyRegionでダーティ領域が空になる", prop.ForAll(
		func(x, y, w, h int) bool {
			if w <= 0 || h <= 0 {
				return true
			}

			lm := NewLayerManager()
			pls := lm.GetOrCreatePictureLayerSet(0)

			// ダーティ領域を設定
			rect := image.Rect(x, y, x+w, y+h)
			pls.AddDirtyRegion(rect)

			// ダーティ領域が設定されていることを確認
			if pls.DirtyRegion.Empty() {
				return false
			}

			// ダーティ領域をクリア
			pls.ClearDirtyRegion()

			// ダーティ領域が空になっていることを確認
			return pls.DirtyRegion.Empty()
		},
		gen.IntRange(0, 640),
		gen.IntRange(0, 480),
		gen.IntRange(1, 100),
		gen.IntRange(1, 100),
	))

	// Property 6.4.2: 初期状態ではダーティ領域は空
	properties.Property("初期状態ではダーティ領域は空", prop.ForAll(
		func(picID int) bool {
			lm := NewLayerManager()
			pls := lm.GetOrCreatePictureLayerSet(picID)

			// 初期状態ではダーティ領域は空
			return pls.DirtyRegion.Empty()
		},
		gen.IntRange(0, 255),
	))

	// Property 6.4.3: ClearDirtyRegionでFullDirtyもクリアされる
	properties.Property("ClearDirtyRegionでFullDirtyもクリアされる", prop.ForAll(
		func(picID int) bool {
			lm := NewLayerManager()
			pls := lm.GetOrCreatePictureLayerSet(picID)

			// 初期状態はFullDirty
			if !pls.FullDirty {
				return false
			}

			// ダーティ領域をクリア
			pls.ClearDirtyRegion()

			// FullDirtyもクリアされていることを確認
			return !pls.FullDirty
		},
		gen.IntRange(0, 255),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty6_LayerPositionChangeAddsDirtyRegion はレイヤー位置変更時のダーティ領域追加をテストする
// 要件 6.1: 変更があった領域（ダーティ領域）を追跡する
func TestProperty6_LayerPositionChangeAddsDirtyRegion(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 6.5.1: CastLayerの位置変更でダーティフラグが設定される
	properties.Property("CastLayerの位置変更でダーティフラグが設定される", prop.ForAll(
		func(x1, y1, x2, y2, w, h int) bool {
			if w <= 0 || h <= 0 {
				return true
			}
			if x1 == x2 && y1 == y2 {
				return true // 位置が同じ場合はスキップ
			}

			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()
			cl := NewCastLayer(layerID, 0, 0, 0, x1, y1, 0, 0, w, h, 0)

			// ダーティフラグをクリア
			cl.SetDirty(false)

			// 位置を変更
			cl.SetPosition(x2, y2)

			// ダーティフラグが設定されていることを確認
			return cl.IsDirty()
		},
		gen.IntRange(0, 300),
		gen.IntRange(0, 200),
		gen.IntRange(0, 300),
		gen.IntRange(0, 200),
		gen.IntRange(1, 64),
		gen.IntRange(1, 64),
	))

	// Property 6.5.2: TextLayerEntryの位置変更でダーティフラグが設定される
	properties.Property("TextLayerEntryの位置変更でダーティフラグが設定される", prop.ForAll(
		func(x1, y1, x2, y2 int) bool {
			if x1 == x2 && y1 == y2 {
				return true // 位置が同じ場合はスキップ
			}

			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()
			tl := NewTextLayerEntry(layerID, 0, x1, y1, "test", 0)

			// ダーティフラグをクリア
			tl.SetDirty(false)

			// 位置を変更
			tl.SetPosition(x2, y2)

			// ダーティフラグが設定されていることを確認
			return tl.IsDirty()
		},
		gen.IntRange(0, 300),
		gen.IntRange(0, 200),
		gen.IntRange(0, 300),
		gen.IntRange(0, 200),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty6_LayerRemovalAddsDirtyRegion はレイヤー削除時のダーティ領域追加をテストする
// 要件 6.1: 変更があった領域（ダーティ領域）を追跡する
func TestProperty6_LayerRemovalAddsDirtyRegion(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 6.6.1: CastLayer削除時にダーティ領域が追加される
	properties.Property("CastLayer削除時にダーティ領域が追加される", prop.ForAll(
		func(x, y, w, h int) bool {
			if w <= 0 || h <= 0 {
				return true
			}

			lm := NewLayerManager()
			pls := lm.GetOrCreatePictureLayerSet(0)

			// キャストレイヤーを追加
			layerID := lm.GetNextLayerID()
			zOrderOffset := pls.GetNextCastZOffset()
			cl := NewCastLayer(layerID, 0, 0, 0, x, y, 0, 0, w, h, zOrderOffset)
			pls.AddCastLayer(cl)

			// ダーティ領域をクリア
			pls.ClearDirtyRegion()

			// キャストレイヤーを削除
			pls.RemoveCastLayer(0)

			// ダーティ領域が追加されていることを確認
			// 削除されたレイヤーの境界がダーティ領域に含まれている
			expectedBounds := image.Rect(x, y, x+w, y+h)
			return !pls.DirtyRegion.Empty() && containsRect(pls.DirtyRegion, expectedBounds)
		},
		gen.IntRange(0, 300),
		gen.IntRange(0, 200),
		gen.IntRange(1, 64),
		gen.IntRange(1, 64),
	))

	// Property 6.6.2: TextLayer削除時にダーティ領域が追加される
	properties.Property("TextLayer削除時にダーティ領域が追加される", prop.ForAll(
		func(x, y, imgW, imgH int) bool {
			if imgW <= 0 || imgH <= 0 {
				return true
			}

			lm := NewLayerManager()
			pls := lm.GetOrCreatePictureLayerSet(0)

			// テキストレイヤーを追加
			layerID := lm.GetNextLayerID()
			zOrderOffset := pls.GetNextTextZOffset()
			tl := NewTextLayerEntry(layerID, 0, x, y, "test", zOrderOffset)
			// 画像を設定して境界を確定
			img := ebiten.NewImage(imgW, imgH)
			tl.SetImage(img)
			pls.AddTextLayer(tl)

			// ダーティ領域をクリア
			pls.ClearDirtyRegion()

			// テキストレイヤーを削除
			pls.RemoveTextLayer(layerID)

			// ダーティ領域が追加されていることを確認
			expectedBounds := image.Rect(x, y, x+imgW, y+imgH)
			return !pls.DirtyRegion.Empty() && containsRect(pls.DirtyRegion, expectedBounds)
		},
		gen.IntRange(0, 300),
		gen.IntRange(0, 200),
		gen.IntRange(1, 64),
		gen.IntRange(1, 64),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty6_DirtyRegionUnionProperties はダーティ領域統合のプロパティをテストする
// 要件 6.3: 複数のダーティ領域があるときにそれらを統合して処理する
func TestProperty6_DirtyRegionUnionProperties(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 6.7.1: 統合された領域は元の領域をすべて含む
	properties.Property("統合された領域は元の領域をすべて含む", prop.ForAll(
		func(x1, y1, w1, h1, x2, y2, w2, h2 int) bool {
			if w1 <= 0 || h1 <= 0 || w2 <= 0 || h2 <= 0 {
				return true
			}

			lm := NewLayerManager()
			pls := lm.GetOrCreatePictureLayerSet(0)

			// ダーティ領域をクリア
			pls.ClearDirtyRegion()

			// 2つのダーティ領域を追加
			rect1 := image.Rect(x1, y1, x1+w1, y1+h1)
			rect2 := image.Rect(x2, y2, x2+w2, y2+h2)
			pls.AddDirtyRegion(rect1)
			pls.AddDirtyRegion(rect2)

			// 統合された領域が両方の元の領域を含むことを確認
			return containsRect(pls.DirtyRegion, rect1) && containsRect(pls.DirtyRegion, rect2)
		},
		gen.IntRange(0, 300),
		gen.IntRange(0, 200),
		gen.IntRange(1, 100),
		gen.IntRange(1, 100),
		gen.IntRange(0, 300),
		gen.IntRange(0, 200),
		gen.IntRange(1, 100),
		gen.IntRange(1, 100),
	))

	// Property 6.7.2: 重なる領域の統合は正しく動作する
	properties.Property("重なる領域の統合は正しく動作する", prop.ForAll(
		func(x, y, w, h, overlap int) bool {
			if w <= 0 || h <= 0 || overlap <= 0 {
				return true
			}
			if overlap >= w || overlap >= h {
				return true // 完全に重なる場合はスキップ
			}

			lm := NewLayerManager()
			pls := lm.GetOrCreatePictureLayerSet(0)

			// ダーティ領域をクリア
			pls.ClearDirtyRegion()

			// 重なる2つの領域を追加
			rect1 := image.Rect(x, y, x+w, y+h)
			rect2 := image.Rect(x+w-overlap, y+h-overlap, x+w*2-overlap, y+h*2-overlap)
			pls.AddDirtyRegion(rect1)
			pls.AddDirtyRegion(rect2)

			// 期待される統合領域
			expectedUnion := rect1.Union(rect2)

			return pls.DirtyRegion == expectedUnion
		},
		gen.IntRange(0, 200),
		gen.IntRange(0, 150),
		gen.IntRange(10, 50),
		gen.IntRange(10, 50),
		gen.IntRange(1, 20),
	))

	// Property 6.7.3: 離れた領域の統合は両方を含む最小の矩形になる
	properties.Property("離れた領域の統合は両方を含む最小の矩形になる", prop.ForAll(
		func(x1, y1, w1, h1, gap, w2, h2 int) bool {
			if w1 <= 0 || h1 <= 0 || w2 <= 0 || h2 <= 0 || gap <= 0 {
				return true
			}

			lm := NewLayerManager()
			pls := lm.GetOrCreatePictureLayerSet(0)

			// ダーティ領域をクリア
			pls.ClearDirtyRegion()

			// 離れた2つの領域を追加
			rect1 := image.Rect(x1, y1, x1+w1, y1+h1)
			rect2 := image.Rect(x1+w1+gap, y1+h1+gap, x1+w1+gap+w2, y1+h1+gap+h2)
			pls.AddDirtyRegion(rect1)
			pls.AddDirtyRegion(rect2)

			// 期待される統合領域
			expectedUnion := rect1.Union(rect2)

			return pls.DirtyRegion == expectedUnion
		},
		gen.IntRange(0, 100),
		gen.IntRange(0, 100),
		gen.IntRange(10, 50),
		gen.IntRange(10, 50),
		gen.IntRange(10, 50),
		gen.IntRange(10, 50),
		gen.IntRange(10, 50),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty6_DirtyRegionAndIsDirty はダーティ領域とIsDirtyの関係をテストする
// 要件 6.2: ダーティ領域のみを再合成する
func TestProperty6_DirtyRegionAndIsDirty(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 6.8.1: ダーティ領域が空でない場合、IsDirtyはtrueを返す
	properties.Property("ダーティ領域が空でない場合、IsDirtyはtrueを返す", prop.ForAll(
		func(x, y, w, h int) bool {
			if w <= 0 || h <= 0 {
				return true
			}

			lm := NewLayerManager()
			pls := lm.GetOrCreatePictureLayerSet(0)

			// すべてのダーティフラグをクリア
			pls.ClearAllDirtyFlags()

			// ダーティ領域を追加
			rect := image.Rect(x, y, x+w, y+h)
			pls.AddDirtyRegion(rect)

			// IsDirtyがtrueを返すことを確認
			return pls.IsDirty()
		},
		gen.IntRange(0, 640),
		gen.IntRange(0, 480),
		gen.IntRange(1, 100),
		gen.IntRange(1, 100),
	))

	// Property 6.8.2: ダーティ領域が空でFullDirtyがfalseの場合、レイヤーのダーティ状態に依存
	properties.Property("ダーティ領域が空でFullDirtyがfalseの場合、レイヤーのダーティ状態に依存", prop.ForAll(
		func(setLayerDirty bool) bool {
			lm := NewLayerManager()
			pls := lm.GetOrCreatePictureLayerSet(0)

			// 背景レイヤーを設定
			bgID := lm.GetNextLayerID()
			bg := NewBackgroundLayer(bgID, 0, nil)
			pls.SetBackground(bg)

			// すべてのダーティフラグをクリア
			pls.ClearAllDirtyFlags()

			// レイヤーのダーティ状態を設定
			if setLayerDirty {
				bg.SetDirty(true)
			}

			// IsDirtyがレイヤーのダーティ状態と一致することを確認
			return pls.IsDirty() == setLayerDirty
		},
		gen.Bool(),
	))

	// Property 6.8.3: ClearAllDirtyFlagsでダーティ領域もクリアされる
	properties.Property("ClearAllDirtyFlagsでダーティ領域もクリアされる", prop.ForAll(
		func(x, y, w, h int) bool {
			if w <= 0 || h <= 0 {
				return true
			}

			lm := NewLayerManager()
			pls := lm.GetOrCreatePictureLayerSet(0)

			// ダーティ領域を追加
			rect := image.Rect(x, y, x+w, y+h)
			pls.AddDirtyRegion(rect)

			// すべてのダーティフラグをクリア
			pls.ClearAllDirtyFlags()

			// ダーティ領域が空になっていることを確認
			return pls.DirtyRegion.Empty() && !pls.FullDirty
		},
		gen.IntRange(0, 640),
		gen.IntRange(0, 480),
		gen.IntRange(1, 100),
		gen.IntRange(1, 100),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty6_DirtyRegionSequence はダーティ領域の一連の操作をテストする
// 複合的なシナリオでダーティ領域が正しく管理されることを確認
func TestProperty6_DirtyRegionSequence(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 6.9.1: 複数の操作後もダーティ領域の一貫性が保たれる
	properties.Property("複数の操作後もダーティ領域の一貫性が保たれる", prop.ForAll(
		func(operations []int) bool {
			if len(operations) == 0 || len(operations) > 50 {
				return true
			}

			lm := NewLayerManager()
			pls := lm.GetOrCreatePictureLayerSet(0)

			// ダーティ領域をクリア
			pls.ClearDirtyRegion()

			var expectedUnion image.Rectangle

			for _, op := range operations {
				switch op % 3 {
				case 0:
					// ダーティ領域を追加
					x := (op * 7) % 640
					y := (op * 11) % 480
					w := (op % 50) + 10
					h := (op % 50) + 10
					rect := image.Rect(x, y, x+w, y+h)
					pls.AddDirtyRegion(rect)
					if expectedUnion.Empty() {
						expectedUnion = rect
					} else {
						expectedUnion = expectedUnion.Union(rect)
					}
				case 1:
					// ダーティ領域をクリア
					pls.ClearDirtyRegion()
					expectedUnion = image.Rectangle{}
				case 2:
					// 何もしない（状態を維持）
				}
			}

			return pls.DirtyRegion == expectedUnion
		},
		gen.SliceOfN(50, gen.IntRange(0, 1000)),
	))

	// Property 6.9.2: レイヤー追加・削除とダーティ領域の整合性
	properties.Property("レイヤー追加・削除とダーティ領域の整合性", prop.ForAll(
		func(castCount int) bool {
			if castCount <= 0 || castCount > 20 {
				return true
			}

			lm := NewLayerManager()
			pls := lm.GetOrCreatePictureLayerSet(0)

			// キャストレイヤーを追加
			for i := 0; i < castCount; i++ {
				layerID := lm.GetNextLayerID()
				zOrderOffset := pls.GetNextCastZOffset()
				cl := NewCastLayer(layerID, i, 0, 0, i*20, i*20, 0, 0, 32, 32, zOrderOffset)
				pls.AddCastLayer(cl)
			}

			// ダーティ領域をクリア
			pls.ClearDirtyRegion()

			// 最初のキャストレイヤーを削除
			if len(pls.Casts) > 0 {
				firstCast := pls.Casts[0]
				expectedBounds := firstCast.GetBounds()
				pls.RemoveCastLayer(firstCast.GetCastID())

				// 削除されたレイヤーの境界がダーティ領域に含まれていることを確認
				return containsRect(pls.DirtyRegion, expectedBounds)
			}

			return true
		},
		gen.IntRange(1, 20),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// ============================================================================
// Feature: layer-based-rendering, Property 7: 上書きスキップの正確性
// **Validates: Requirements 7.1, 7.2, 7.3**
//
// 任意の不透明なレイヤーが別のレイヤーを完全に覆っている場合、
// 覆われたレイヤーの描画がスキップされる。
// ============================================================================

// TestProperty7_OpaqueLayerCompletelyCoversAnother は不透明なレイヤーが別のレイヤーを完全に覆う場合をテストする
// 要件 7.1: 不透明なレイヤーが別のレイヤーを完全に覆っているときにそのレイヤーの描画をスキップする
func TestProperty7_OpaqueLayerCompletelyCoversAnother(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 7.1.1: 不透明なレイヤーが下位レイヤーを完全に覆う場合、ShouldSkipLayerはtrueを返す
	properties.Property("不透明なレイヤーが下位レイヤーを完全に覆う場合、ShouldSkipLayerはtrueを返す", prop.ForAll(
		func(lowerX, lowerY, lowerW, lowerH, margin int) bool {
			if lowerW <= 0 || lowerH <= 0 || margin < 0 {
				return true
			}

			lm := NewLayerManager()

			// 下位レイヤーを作成
			lowerID := lm.GetNextLayerID()
			lowerLayer := NewCastLayer(lowerID, 0, 0, 0, lowerX, lowerY, 0, 0, lowerW, lowerH, 0)

			// 上位レイヤーを作成（下位レイヤーを完全に覆う）
			upperID := lm.GetNextLayerID()
			upperLayer := NewCastLayer(upperID, 1, 0, 0,
				lowerX-margin, lowerY-margin, 0, 0,
				lowerW+margin*2, lowerH+margin*2, 1)
			upperLayer.SetOpaque(true)

			upperLayers := []Layer{upperLayer}

			return lm.ShouldSkipLayer(lowerLayer, upperLayers)
		},
		gen.IntRange(0, 100),
		gen.IntRange(0, 100),
		gen.IntRange(10, 64),
		gen.IntRange(10, 64),
		gen.IntRange(0, 20),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty7_OpaqueLayerExactlySameSize は不透明なレイヤーが同じサイズで覆う場合をテストする
func TestProperty7_OpaqueLayerExactlySameSize(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 7.1.2: 不透明なレイヤーが同じサイズで覆う場合、ShouldSkipLayerはtrueを返す
	properties.Property("不透明なレイヤーが同じサイズで覆う場合、ShouldSkipLayerはtrueを返す", prop.ForAll(
		func(x, y, w, h int) bool {
			if w <= 0 || h <= 0 {
				return true
			}

			lm := NewLayerManager()

			// 下位レイヤーを作成
			lowerID := lm.GetNextLayerID()
			lowerLayer := NewCastLayer(lowerID, 0, 0, 0, x, y, 0, 0, w, h, 0)

			// 上位レイヤーを作成（同じ位置・サイズ）
			upperID := lm.GetNextLayerID()
			upperLayer := NewCastLayer(upperID, 1, 0, 0, x, y, 0, 0, w, h, 1)
			upperLayer.SetOpaque(true)

			upperLayers := []Layer{upperLayer}

			return lm.ShouldSkipLayer(lowerLayer, upperLayers)
		},
		gen.IntRange(0, 200),
		gen.IntRange(0, 200),
		gen.IntRange(10, 100),
		gen.IntRange(10, 100),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty7_OpaqueLayerPartiallyCovers は不透明なレイヤーが部分的に覆う場合をテストする
// 要件 7.2: 部分的に覆われているレイヤーは描画する
func TestProperty7_OpaqueLayerPartiallyCovers(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 7.2.1: 不透明なレイヤーが部分的に覆う場合、ShouldSkipLayerはfalseを返す
	properties.Property("不透明なレイヤーが部分的に覆う場合、ShouldSkipLayerはfalseを返す", prop.ForAll(
		func(lowerX, lowerY, lowerW, lowerH, offset int) bool {
			if lowerW <= 0 || lowerH <= 0 || offset <= 0 {
				return true
			}
			if offset >= lowerW || offset >= lowerH {
				return true // 完全に覆わない場合のみテスト
			}

			lm := NewLayerManager()

			// 下位レイヤーを作成
			lowerID := lm.GetNextLayerID()
			lowerLayer := NewCastLayer(lowerID, 0, 0, 0, lowerX, lowerY, 0, 0, lowerW, lowerH, 0)

			// 上位レイヤーを作成（部分的に覆う - 右下にずれている）
			upperID := lm.GetNextLayerID()
			upperLayer := NewCastLayer(upperID, 1, 0, 0,
				lowerX+offset, lowerY+offset, 0, 0,
				lowerW, lowerH, 1)
			upperLayer.SetOpaque(true)

			upperLayers := []Layer{upperLayer}

			// 部分的に覆う場合はスキップしない
			return !lm.ShouldSkipLayer(lowerLayer, upperLayers)
		},
		gen.IntRange(0, 100),
		gen.IntRange(0, 100),
		gen.IntRange(20, 64),
		gen.IntRange(20, 64),
		gen.IntRange(1, 19),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty7_OpaqueLayerSmallerThanLower は上位レイヤーが下位レイヤーより小さい場合をテストする
func TestProperty7_OpaqueLayerSmallerThanLower(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 7.2.2: 上位レイヤーが下位レイヤーより小さい場合、ShouldSkipLayerはfalseを返す
	properties.Property("上位レイヤーが下位レイヤーより小さい場合、ShouldSkipLayerはfalseを返す", prop.ForAll(
		func(x, y, lowerW, lowerH, shrink int) bool {
			if lowerW <= 0 || lowerH <= 0 || shrink <= 0 {
				return true
			}
			upperW := lowerW - shrink
			upperH := lowerH - shrink
			if upperW <= 0 || upperH <= 0 {
				return true
			}

			lm := NewLayerManager()

			// 下位レイヤーを作成
			lowerID := lm.GetNextLayerID()
			lowerLayer := NewCastLayer(lowerID, 0, 0, 0, x, y, 0, 0, lowerW, lowerH, 0)

			// 上位レイヤーを作成（下位レイヤーより小さい）
			upperID := lm.GetNextLayerID()
			upperLayer := NewCastLayer(upperID, 1, 0, 0, x, y, 0, 0, upperW, upperH, 1)
			upperLayer.SetOpaque(true)

			upperLayers := []Layer{upperLayer}

			// 小さいレイヤーでは完全に覆えないのでスキップしない
			return !lm.ShouldSkipLayer(lowerLayer, upperLayers)
		},
		gen.IntRange(0, 100),
		gen.IntRange(0, 100),
		gen.IntRange(20, 100),
		gen.IntRange(20, 100),
		gen.IntRange(1, 19),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty7_TransparentLayerDoesNotCauseSkip は透明なレイヤーが覆っても描画をスキップしないことをテストする
// 要件 7.3: 透明なレイヤーは上書きスキップの対象としない
func TestProperty7_TransparentLayerDoesNotCauseSkip(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 7.3.1: 透明なレイヤーが完全に覆っても、ShouldSkipLayerはfalseを返す
	properties.Property("透明なレイヤーが完全に覆っても、ShouldSkipLayerはfalseを返す", prop.ForAll(
		func(lowerX, lowerY, lowerW, lowerH, margin int) bool {
			if lowerW <= 0 || lowerH <= 0 || margin < 0 {
				return true
			}

			lm := NewLayerManager()

			// 下位レイヤーを作成
			lowerID := lm.GetNextLayerID()
			lowerLayer := NewCastLayer(lowerID, 0, 0, 0, lowerX, lowerY, 0, 0, lowerW, lowerH, 0)

			// 上位レイヤーを作成（下位レイヤーを完全に覆うが透明）
			upperID := lm.GetNextLayerID()
			upperLayer := NewCastLayer(upperID, 1, 0, 0,
				lowerX-margin, lowerY-margin, 0, 0,
				lowerW+margin*2, lowerH+margin*2, 1)
			upperLayer.SetOpaque(false) // 透明

			upperLayers := []Layer{upperLayer}

			// 透明なレイヤーではスキップしない
			return !lm.ShouldSkipLayer(lowerLayer, upperLayers)
		},
		gen.IntRange(0, 100),
		gen.IntRange(0, 100),
		gen.IntRange(10, 64),
		gen.IntRange(10, 64),
		gen.IntRange(0, 20),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty7_TransparentLayerDefaultState はレイヤーのデフォルト透明状態をテストする
func TestProperty7_TransparentLayerDefaultState(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 7.3.2: CastLayerのデフォルト状態は透明（不透明ではない）
	properties.Property("CastLayerのデフォルト状態は透明（不透明ではない）", prop.ForAll(
		func(x, y, w, h int) bool {
			if w <= 0 || h <= 0 {
				return true
			}

			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()
			cl := NewCastLayer(layerID, 0, 0, 0, x, y, 0, 0, w, h, 0)

			// デフォルトは透明（不透明ではない）
			return !cl.IsOpaque()
		},
		gen.IntRange(0, 200),
		gen.IntRange(0, 200),
		gen.IntRange(1, 100),
		gen.IntRange(1, 100),
	))

	// Property 7.3.3: SetOpaqueで不透明度を変更できる
	properties.Property("SetOpaqueで不透明度を変更できる", prop.ForAll(
		func(x, y, w, h int, setOpaque bool) bool {
			if w <= 0 || h <= 0 {
				return true
			}

			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()
			cl := NewCastLayer(layerID, 0, 0, 0, x, y, 0, 0, w, h, 0)

			cl.SetOpaque(setOpaque)

			return cl.IsOpaque() == setOpaque
		},
		gen.IntRange(0, 200),
		gen.IntRange(0, 200),
		gen.IntRange(1, 100),
		gen.IntRange(1, 100),
		gen.Bool(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty7_MultipleOpaqueLayersInChain は複数の不透明レイヤーのチェーンをテストする
func TestProperty7_MultipleOpaqueLayersInChain(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 7.4.1: 複数の不透明レイヤーのうち1つでも完全に覆えばスキップする
	properties.Property("複数の不透明レイヤーのうち1つでも完全に覆えばスキップする", prop.ForAll(
		func(x, y, w, h, upperCount int) bool {
			if w <= 0 || h <= 0 || upperCount <= 0 || upperCount > 10 {
				return true
			}

			lm := NewLayerManager()

			// 下位レイヤーを作成
			lowerID := lm.GetNextLayerID()
			lowerLayer := NewCastLayer(lowerID, 0, 0, 0, x, y, 0, 0, w, h, 0)

			// 複数の上位レイヤーを作成（最後の1つだけが完全に覆う）
			var upperLayers []Layer
			for i := 0; i < upperCount-1; i++ {
				upperID := lm.GetNextLayerID()
				// 部分的に覆うレイヤー
				upperLayer := NewCastLayer(upperID, i+1, 0, 0,
					x+10, y+10, 0, 0, w/2, h/2, i+1)
				upperLayer.SetOpaque(true)
				upperLayers = append(upperLayers, upperLayer)
			}

			// 最後のレイヤーは完全に覆う
			lastID := lm.GetNextLayerID()
			lastLayer := NewCastLayer(lastID, upperCount, 0, 0,
				x-5, y-5, 0, 0, w+10, h+10, upperCount)
			lastLayer.SetOpaque(true)
			upperLayers = append(upperLayers, lastLayer)

			// 1つでも完全に覆えばスキップする
			return lm.ShouldSkipLayer(lowerLayer, upperLayers)
		},
		gen.IntRange(10, 100),
		gen.IntRange(10, 100),
		gen.IntRange(20, 64),
		gen.IntRange(20, 64),
		gen.IntRange(1, 10),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty7_NoOpaqueLayerCovers は不透明レイヤーがない場合をテストする
func TestProperty7_NoOpaqueLayerCovers(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 7.4.2: 複数の透明レイヤーがあっても、どれも完全に覆わなければスキップしない
	properties.Property("複数の透明レイヤーがあっても、どれも完全に覆わなければスキップしない", prop.ForAll(
		func(x, y, w, h, upperCount int) bool {
			if w <= 0 || h <= 0 || upperCount <= 0 || upperCount > 10 {
				return true
			}

			lm := NewLayerManager()

			// 下位レイヤーを作成
			lowerID := lm.GetNextLayerID()
			lowerLayer := NewCastLayer(lowerID, 0, 0, 0, x, y, 0, 0, w, h, 0)

			// 複数の上位レイヤーを作成（すべて透明）
			var upperLayers []Layer
			for i := 0; i < upperCount; i++ {
				upperID := lm.GetNextLayerID()
				// 完全に覆うが透明
				upperLayer := NewCastLayer(upperID, i+1, 0, 0,
					x-5, y-5, 0, 0, w+10, h+10, i+1)
				upperLayer.SetOpaque(false) // 透明
				upperLayers = append(upperLayers, upperLayer)
			}

			// すべて透明なのでスキップしない
			return !lm.ShouldSkipLayer(lowerLayer, upperLayers)
		},
		gen.IntRange(10, 100),
		gen.IntRange(10, 100),
		gen.IntRange(20, 64),
		gen.IntRange(20, 64),
		gen.IntRange(1, 10),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty7_InvisibleUpperLayerDoesNotCauseSkip は非表示の上位レイヤーがスキップを引き起こさないことをテストする
func TestProperty7_InvisibleUpperLayerDoesNotCauseSkip(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 7.5.1: 非表示の不透明レイヤーが完全に覆っても、ShouldSkipLayerはfalseを返す
	properties.Property("非表示の不透明レイヤーが完全に覆っても、ShouldSkipLayerはfalseを返す", prop.ForAll(
		func(lowerX, lowerY, lowerW, lowerH, margin int) bool {
			if lowerW <= 0 || lowerH <= 0 || margin < 0 {
				return true
			}

			lm := NewLayerManager()

			// 下位レイヤーを作成
			lowerID := lm.GetNextLayerID()
			lowerLayer := NewCastLayer(lowerID, 0, 0, 0, lowerX, lowerY, 0, 0, lowerW, lowerH, 0)

			// 上位レイヤーを作成（完全に覆うが非表示）
			upperID := lm.GetNextLayerID()
			upperLayer := NewCastLayer(upperID, 1, 0, 0,
				lowerX-margin, lowerY-margin, 0, 0,
				lowerW+margin*2, lowerH+margin*2, 1)
			upperLayer.SetOpaque(true)
			upperLayer.SetVisible(false) // 非表示

			upperLayers := []Layer{upperLayer}

			// 非表示のレイヤーではスキップしない
			return !lm.ShouldSkipLayer(lowerLayer, upperLayers)
		},
		gen.IntRange(0, 100),
		gen.IntRange(0, 100),
		gen.IntRange(10, 64),
		gen.IntRange(10, 64),
		gen.IntRange(0, 20),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty7_EmptyBoundsHandling は空の境界を持つレイヤーの処理をテストする
func TestProperty7_EmptyBoundsHandling(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 7.6.1: 下位レイヤーの境界が空の場合、ShouldSkipLayerはtrueを返す
	properties.Property("下位レイヤーの境界が空の場合、ShouldSkipLayerはtrueを返す", prop.ForAll(
		func(x, y int) bool {
			lm := NewLayerManager()

			// 境界が空の下位レイヤーを作成（幅または高さが0）
			lowerID := lm.GetNextLayerID()
			lowerLayer := NewCastLayer(lowerID, 0, 0, 0, x, y, 0, 0, 0, 0, 0)

			// 上位レイヤーを作成
			upperID := lm.GetNextLayerID()
			upperLayer := NewCastLayer(upperID, 1, 0, 0, 0, 0, 0, 0, 100, 100, 1)
			upperLayer.SetOpaque(true)

			upperLayers := []Layer{upperLayer}

			// 空の境界を持つレイヤーはスキップ
			return lm.ShouldSkipLayer(lowerLayer, upperLayers)
		},
		gen.IntRange(0, 200),
		gen.IntRange(0, 200),
	))

	// Property 7.6.2: 上位レイヤーの境界が空の場合、スキップを引き起こさない
	properties.Property("上位レイヤーの境界が空の場合、スキップを引き起こさない", prop.ForAll(
		func(x, y, w, h int) bool {
			if w <= 0 || h <= 0 {
				return true
			}

			lm := NewLayerManager()

			// 下位レイヤーを作成
			lowerID := lm.GetNextLayerID()
			lowerLayer := NewCastLayer(lowerID, 0, 0, 0, x, y, 0, 0, w, h, 0)

			// 境界が空の上位レイヤーを作成
			upperID := lm.GetNextLayerID()
			upperLayer := NewCastLayer(upperID, 1, 0, 0, 0, 0, 0, 0, 0, 0, 1)
			upperLayer.SetOpaque(true)

			upperLayers := []Layer{upperLayer}

			// 空の境界を持つ上位レイヤーはスキップを引き起こさない
			return !lm.ShouldSkipLayer(lowerLayer, upperLayers)
		},
		gen.IntRange(0, 200),
		gen.IntRange(0, 200),
		gen.IntRange(10, 100),
		gen.IntRange(10, 100),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty7_NilLayerHandling はnilレイヤーの処理をテストする
func TestProperty7_NilLayerHandling(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 7.6.3: nilの下位レイヤーの場合、ShouldSkipLayerはtrueを返す
	properties.Property("nilの下位レイヤーの場合、ShouldSkipLayerはtrueを返す", prop.ForAll(
		func(x, y, w, h int) bool {
			if w <= 0 || h <= 0 {
				return true
			}

			lm := NewLayerManager()

			// 上位レイヤーを作成
			upperID := lm.GetNextLayerID()
			upperLayer := NewCastLayer(upperID, 1, 0, 0, x, y, 0, 0, w, h, 1)
			upperLayer.SetOpaque(true)

			upperLayers := []Layer{upperLayer}

			// nilレイヤーはスキップ
			return lm.ShouldSkipLayer(nil, upperLayers)
		},
		gen.IntRange(0, 200),
		gen.IntRange(0, 200),
		gen.IntRange(10, 100),
		gen.IntRange(10, 100),
	))

	// Property 7.6.4: 上位レイヤーリストにnilが含まれていても正しく処理される
	properties.Property("上位レイヤーリストにnilが含まれていても正しく処理される", prop.ForAll(
		func(x, y, w, h int) bool {
			if w <= 0 || h <= 0 {
				return true
			}

			lm := NewLayerManager()

			// 下位レイヤーを作成
			lowerID := lm.GetNextLayerID()
			lowerLayer := NewCastLayer(lowerID, 0, 0, 0, x, y, 0, 0, w, h, 0)

			// nilを含む上位レイヤーリスト
			upperLayers := []Layer{nil, nil, nil}

			// nilのみの場合はスキップしない
			return !lm.ShouldSkipLayer(lowerLayer, upperLayers)
		},
		gen.IntRange(0, 200),
		gen.IntRange(0, 200),
		gen.IntRange(10, 100),
		gen.IntRange(10, 100),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty7_InvisibleLowerLayerSkipped は非表示の下位レイヤーがスキップされることをテストする
func TestProperty7_InvisibleLowerLayerSkipped(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 7.7.1: 非表示の下位レイヤーはスキップされる
	properties.Property("非表示の下位レイヤーはスキップされる", prop.ForAll(
		func(x, y, w, h int) bool {
			if w <= 0 || h <= 0 {
				return true
			}

			lm := NewLayerManager()

			// 非表示の下位レイヤーを作成
			lowerID := lm.GetNextLayerID()
			lowerLayer := NewCastLayer(lowerID, 0, 0, 0, x, y, 0, 0, w, h, 0)
			lowerLayer.SetVisible(false)

			// 上位レイヤーなし
			var upperLayers []Layer

			// 非表示のレイヤーはスキップ
			return lm.ShouldSkipLayer(lowerLayer, upperLayers)
		},
		gen.IntRange(0, 200),
		gen.IntRange(0, 200),
		gen.IntRange(10, 100),
		gen.IntRange(10, 100),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty7_ContainsRectHelper はcontainsRectヘルパー関数をテストする
func TestProperty7_ContainsRectHelper(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 7.8.1: 同じ矩形はcontainsRectでtrueを返す
	properties.Property("同じ矩形はcontainsRectでtrueを返す", prop.ForAll(
		func(x, y, w, h int) bool {
			if w <= 0 || h <= 0 {
				return true
			}

			rect := image.Rect(x, y, x+w, y+h)
			return containsRect(rect, rect)
		},
		gen.IntRange(0, 200),
		gen.IntRange(0, 200),
		gen.IntRange(1, 100),
		gen.IntRange(1, 100),
	))

	// Property 7.8.2: 大きい矩形は小さい矩形を含む
	properties.Property("大きい矩形は小さい矩形を含む", prop.ForAll(
		func(x, y, w, h, margin int) bool {
			if w <= 0 || h <= 0 || margin < 0 {
				return true
			}

			inner := image.Rect(x, y, x+w, y+h)
			outer := image.Rect(x-margin, y-margin, x+w+margin, y+h+margin)

			return containsRect(outer, inner)
		},
		gen.IntRange(10, 100),
		gen.IntRange(10, 100),
		gen.IntRange(10, 50),
		gen.IntRange(10, 50),
		gen.IntRange(1, 20),
	))

	// Property 7.8.3: 小さい矩形は大きい矩形を含まない
	properties.Property("小さい矩形は大きい矩形を含まない", prop.ForAll(
		func(x, y, w, h, margin int) bool {
			if w <= 0 || h <= 0 || margin <= 0 {
				return true
			}

			inner := image.Rect(x, y, x+w, y+h)
			outer := image.Rect(x-margin, y-margin, x+w+margin, y+h+margin)

			return !containsRect(inner, outer)
		},
		gen.IntRange(10, 100),
		gen.IntRange(10, 100),
		gen.IntRange(10, 50),
		gen.IntRange(10, 50),
		gen.IntRange(1, 20),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty7_GetUpperLayersIntegration はGetUpperLayersとShouldSkipLayerの統合をテストする
func TestProperty7_GetUpperLayersIntegration(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 7.9.1: GetUpperLayersで取得したレイヤーでShouldSkipLayerが正しく動作する
	properties.Property("GetUpperLayersで取得したレイヤーでShouldSkipLayerが正しく動作する", prop.ForAll(
		func(castCount int) bool {
			if castCount <= 1 || castCount > 10 {
				return true
			}

			lm := NewLayerManager()
			pls := lm.GetOrCreatePictureLayerSet(0)

			// 複数のキャストレイヤーを追加
			for i := 0; i < castCount; i++ {
				layerID := lm.GetNextLayerID()
				zOrderOffset := pls.GetNextCastZOffset()
				cl := NewCastLayer(layerID, i, 0, 0, 0, 0, 0, 0, 100, 100, zOrderOffset)
				if i == castCount-1 {
					// 最後のレイヤーを不透明に設定
					cl.SetOpaque(true)
				}
				pls.AddCastLayer(cl)
			}

			// 最初のレイヤーの上位レイヤーを取得
			firstLayer := pls.Casts[0]
			upperLayers := pls.GetUpperLayers(firstLayer.GetZOrder())

			// 最後のレイヤーが不透明で同じ位置・サイズなのでスキップされるべき
			return lm.ShouldSkipLayer(firstLayer, upperLayers)
		},
		gen.IntRange(2, 10),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// ============================================================================
// Feature: layer-based-rendering, Property 2: レイヤー操作の整合性
// **Validates: Requirements 2.1, 2.2, 2.3, 2.4, 2.5, 2.6**
//
// 任意のレイヤー操作（PutCast、MoveCast、DelCast、MovePic、TextWrite）に対して、
// 対応するレイヤーが正しく作成、更新、または削除される。
// ============================================================================

// TestProperty2_PutCastCreatesCastLayer はPutCastでCastLayerが作成されることをテストする
// 要件 2.1: PutCastが呼び出されたときに新しいCast_Layerを作成する
func TestProperty2_PutCastCreatesCastLayer(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 2.1.1: AddCastLayerでキャストレイヤーが追加される
	properties.Property("AddCastLayerでキャストレイヤーが追加される", prop.ForAll(
		func(picID, castID, x, y, w, h int) bool {
			if w <= 0 || h <= 0 {
				return true
			}

			lm := NewLayerManager()
			pls := lm.GetOrCreatePictureLayerSet(picID)

			initialCount := pls.GetCastLayerCount()

			// キャストレイヤーを追加
			layerID := lm.GetNextLayerID()
			zOrderOffset := pls.GetNextCastZOffset()
			cl := NewCastLayer(layerID, castID, picID, 0, x, y, 0, 0, w, h, zOrderOffset)
			pls.AddCastLayer(cl)

			// レイヤー数が増加していることを確認
			return pls.GetCastLayerCount() == initialCount+1
		},
		gen.IntRange(0, 255),
		gen.IntRange(0, 100),
		gen.IntRange(0, 640),
		gen.IntRange(0, 480),
		gen.IntRange(1, 64),
		gen.IntRange(1, 64),
	))

	// Property 2.1.2: 追加されたキャストレイヤーが正しい属性を持つ
	properties.Property("追加されたキャストレイヤーが正しい属性を持つ", prop.ForAll(
		func(picID, castID, x, y, w, h int) bool {
			if w <= 0 || h <= 0 {
				return true
			}

			lm := NewLayerManager()
			pls := lm.GetOrCreatePictureLayerSet(picID)

			// キャストレイヤーを追加
			layerID := lm.GetNextLayerID()
			zOrderOffset := pls.GetNextCastZOffset()
			cl := NewCastLayer(layerID, castID, picID, 0, x, y, 0, 0, w, h, zOrderOffset)
			pls.AddCastLayer(cl)

			// 追加されたレイヤーを取得
			addedLayer := pls.GetCastLayer(castID)
			if addedLayer == nil {
				return false
			}

			// 属性が正しいことを確認
			posX, posY := addedLayer.GetPosition()
			width, height := addedLayer.GetSize()

			return posX == x && posY == y && width == w && height == h &&
				addedLayer.GetCastID() == castID && addedLayer.GetPicID() == picID
		},
		gen.IntRange(0, 255),
		gen.IntRange(0, 100),
		gen.IntRange(0, 640),
		gen.IntRange(0, 480),
		gen.IntRange(1, 64),
		gen.IntRange(1, 64),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty2_PutCastMultipleCasts は複数のキャストを追加できることをテストする
func TestProperty2_PutCastMultipleCasts(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 2.1.3: 複数のキャストレイヤーを追加できる
	properties.Property("複数のキャストレイヤーを追加できる", prop.ForAll(
		func(castCount int) bool {
			if castCount <= 0 || castCount > 50 {
				return true
			}

			lm := NewLayerManager()
			pls := lm.GetOrCreatePictureLayerSet(0)

			// 複数のキャストレイヤーを追加
			for i := 0; i < castCount; i++ {
				layerID := lm.GetNextLayerID()
				zOrderOffset := pls.GetNextCastZOffset()
				cl := NewCastLayer(layerID, i, 0, 0, i*10, i*10, 0, 0, 32, 32, zOrderOffset)
				pls.AddCastLayer(cl)
			}

			// すべてのレイヤーが追加されていることを確認
			if pls.GetCastLayerCount() != castCount {
				return false
			}

			// 各レイヤーが取得できることを確認
			for i := 0; i < castCount; i++ {
				if pls.GetCastLayer(i) == nil {
					return false
				}
			}

			return true
		},
		gen.IntRange(1, 50),
	))

	// Property 2.1.4: 追加後にFullDirtyが設定される
	properties.Property("キャストレイヤー追加後にFullDirtyが設定される", prop.ForAll(
		func(picID, castID, x, y int) bool {
			lm := NewLayerManager()
			pls := lm.GetOrCreatePictureLayerSet(picID)

			// ダーティフラグをクリア
			pls.ClearAllDirtyFlags()

			// キャストレイヤーを追加
			layerID := lm.GetNextLayerID()
			zOrderOffset := pls.GetNextCastZOffset()
			cl := NewCastLayer(layerID, castID, picID, 0, x, y, 0, 0, 32, 32, zOrderOffset)
			pls.AddCastLayer(cl)

			// FullDirtyが設定されていることを確認
			return pls.FullDirty
		},
		gen.IntRange(0, 255),
		gen.IntRange(0, 100),
		gen.IntRange(0, 640),
		gen.IntRange(0, 480),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty2_MoveCastUpdatesPosition はMoveCastでCastLayerの位置が更新されることをテストする
// 要件 2.2: MoveCastが呼び出されたときに対応するCast_Layerの位置を更新する
func TestProperty2_MoveCastUpdatesPosition(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 2.2.1: SetPositionでキャストレイヤーの位置が更新される
	properties.Property("SetPositionでキャストレイヤーの位置が更新される", prop.ForAll(
		func(x1, y1, x2, y2, w, h int) bool {
			if w <= 0 || h <= 0 {
				return true
			}

			lm := NewLayerManager()
			pls := lm.GetOrCreatePictureLayerSet(0)

			// キャストレイヤーを追加
			layerID := lm.GetNextLayerID()
			zOrderOffset := pls.GetNextCastZOffset()
			cl := NewCastLayer(layerID, 0, 0, 0, x1, y1, 0, 0, w, h, zOrderOffset)
			pls.AddCastLayer(cl)

			// 位置を更新
			cl.SetPosition(x2, y2)

			// 位置が更新されていることを確認
			posX, posY := cl.GetPosition()
			return posX == x2 && posY == y2
		},
		gen.IntRange(0, 300),
		gen.IntRange(0, 200),
		gen.IntRange(0, 300),
		gen.IntRange(0, 200),
		gen.IntRange(1, 64),
		gen.IntRange(1, 64),
	))

	// Property 2.2.2: 位置更新後に境界ボックスが正しく更新される
	properties.Property("位置更新後に境界ボックスが正しく更新される", prop.ForAll(
		func(x1, y1, x2, y2, w, h int) bool {
			if w <= 0 || h <= 0 {
				return true
			}

			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()
			cl := NewCastLayer(layerID, 0, 0, 0, x1, y1, 0, 0, w, h, 0)

			// 位置を更新
			cl.SetPosition(x2, y2)

			// 境界ボックスが正しいことを確認
			bounds := cl.GetBounds()
			expectedBounds := image.Rect(x2, y2, x2+w, y2+h)

			return bounds == expectedBounds
		},
		gen.IntRange(0, 300),
		gen.IntRange(0, 200),
		gen.IntRange(0, 300),
		gen.IntRange(0, 200),
		gen.IntRange(1, 64),
		gen.IntRange(1, 64),
	))

	// Property 2.2.3: 位置更新後にダーティフラグが設定される
	properties.Property("位置更新後にダーティフラグが設定される", prop.ForAll(
		func(x1, y1, x2, y2, w, h int) bool {
			if w <= 0 || h <= 0 {
				return true
			}
			if x1 == x2 && y1 == y2 {
				return true // 位置が同じ場合はスキップ
			}

			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()
			cl := NewCastLayer(layerID, 0, 0, 0, x1, y1, 0, 0, w, h, 0)

			// ダーティフラグをクリア
			cl.SetDirty(false)

			// 位置を更新
			cl.SetPosition(x2, y2)

			// ダーティフラグが設定されていることを確認
			return cl.IsDirty()
		},
		gen.IntRange(0, 300),
		gen.IntRange(0, 200),
		gen.IntRange(0, 300),
		gen.IntRange(0, 200),
		gen.IntRange(1, 64),
		gen.IntRange(1, 64),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty2_DelCastRemovesCastLayer はDelCastでCastLayerが削除されることをテストする
// 要件 2.3: DelCastが呼び出されたときに対応するCast_Layerを削除する
func TestProperty2_DelCastRemovesCastLayer(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 2.3.1: RemoveCastLayerでキャストレイヤーが削除される
	properties.Property("RemoveCastLayerでキャストレイヤーが削除される", prop.ForAll(
		func(castID, x, y, w, h int) bool {
			if w <= 0 || h <= 0 {
				return true
			}

			lm := NewLayerManager()
			pls := lm.GetOrCreatePictureLayerSet(0)

			// キャストレイヤーを追加
			layerID := lm.GetNextLayerID()
			zOrderOffset := pls.GetNextCastZOffset()
			cl := NewCastLayer(layerID, castID, 0, 0, x, y, 0, 0, w, h, zOrderOffset)
			pls.AddCastLayer(cl)

			// レイヤーが追加されていることを確認
			if pls.GetCastLayer(castID) == nil {
				return false
			}

			// レイヤーを削除
			result := pls.RemoveCastLayer(castID)

			// 削除が成功し、レイヤーが存在しないことを確認
			return result && pls.GetCastLayer(castID) == nil
		},
		gen.IntRange(0, 100),
		gen.IntRange(0, 640),
		gen.IntRange(0, 480),
		gen.IntRange(1, 64),
		gen.IntRange(1, 64),
	))

	// Property 2.3.2: 削除後にレイヤー数が減少する
	properties.Property("削除後にレイヤー数が減少する", prop.ForAll(
		func(castCount int) bool {
			if castCount <= 0 || castCount > 20 {
				return true
			}

			lm := NewLayerManager()
			pls := lm.GetOrCreatePictureLayerSet(0)

			// 複数のキャストレイヤーを追加
			for i := 0; i < castCount; i++ {
				layerID := lm.GetNextLayerID()
				zOrderOffset := pls.GetNextCastZOffset()
				cl := NewCastLayer(layerID, i, 0, 0, i*10, i*10, 0, 0, 32, 32, zOrderOffset)
				pls.AddCastLayer(cl)
			}

			// 最初のレイヤーを削除
			countBefore := pls.GetCastLayerCount()
			pls.RemoveCastLayer(0)
			countAfter := pls.GetCastLayerCount()

			return countAfter == countBefore-1
		},
		gen.IntRange(1, 20),
	))

	// Property 2.3.3: 存在しないキャストIDの削除はfalseを返す
	properties.Property("存在しないキャストIDの削除はfalseを返す", prop.ForAll(
		func(castID, nonExistentID int) bool {
			if castID == nonExistentID {
				return true // 同じIDの場合はスキップ
			}

			lm := NewLayerManager()
			pls := lm.GetOrCreatePictureLayerSet(0)

			// キャストレイヤーを追加
			layerID := lm.GetNextLayerID()
			zOrderOffset := pls.GetNextCastZOffset()
			cl := NewCastLayer(layerID, castID, 0, 0, 0, 0, 0, 0, 32, 32, zOrderOffset)
			pls.AddCastLayer(cl)

			// 存在しないIDで削除を試みる
			result := pls.RemoveCastLayer(nonExistentID)

			// 削除が失敗することを確認
			return !result
		},
		gen.IntRange(0, 50),
		gen.IntRange(51, 100),
	))

	// Property 2.3.4: 削除後にFullDirtyが設定される
	properties.Property("キャストレイヤー削除後にFullDirtyが設定される", prop.ForAll(
		func(castID, x, y int) bool {
			lm := NewLayerManager()
			pls := lm.GetOrCreatePictureLayerSet(0)

			// キャストレイヤーを追加
			layerID := lm.GetNextLayerID()
			zOrderOffset := pls.GetNextCastZOffset()
			cl := NewCastLayer(layerID, castID, 0, 0, x, y, 0, 0, 32, 32, zOrderOffset)
			pls.AddCastLayer(cl)

			// ダーティフラグをクリア
			pls.ClearAllDirtyFlags()

			// レイヤーを削除
			pls.RemoveCastLayer(castID)

			// FullDirtyが設定されていることを確認
			return pls.FullDirty
		},
		gen.IntRange(0, 100),
		gen.IntRange(0, 640),
		gen.IntRange(0, 480),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty2_MovePicUpdatesDrawingLayer はMovePicでDrawingLayerが更新されることをテストする
// 要件 2.4: MovePicが呼び出されたときにDrawing_Layerに描画内容を追加する
func TestProperty2_MovePicUpdatesDrawingLayer(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 2.4.1: DrawingLayerのDrawImageで描画内容が追加される
	properties.Property("DrawingLayerのDrawImageで描画内容が追加される", prop.ForAll(
		func(destX, destY, srcW, srcH int) bool {
			if srcW <= 0 || srcH <= 0 {
				return true
			}

			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()
			dl := NewDrawingLayer(layerID, 0, 640, 480)

			// ソース画像を作成
			srcImg := ebiten.NewImage(srcW, srcH)

			// ダーティフラグをクリア
			dl.SetDirty(false)

			// 描画
			dl.DrawImage(srcImg, destX, destY)

			// ダーティフラグが設定されていることを確認
			return dl.IsDirty()
		},
		gen.IntRange(0, 600),
		gen.IntRange(0, 400),
		gen.IntRange(1, 64),
		gen.IntRange(1, 64),
	))

	// Property 2.4.2: DrawingLayerのDrawSubImageで部分描画ができる
	properties.Property("DrawingLayerのDrawSubImageで部分描画ができる", prop.ForAll(
		func(destX, destY, srcX, srcY, w, h int) bool {
			if w <= 0 || h <= 0 {
				return true
			}

			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()
			dl := NewDrawingLayer(layerID, 0, 640, 480)

			// ソース画像を作成（十分なサイズ）
			srcImg := ebiten.NewImage(256, 256)

			// ダーティフラグをクリア
			dl.SetDirty(false)

			// 部分描画
			dl.DrawSubImage(srcImg, destX, destY, srcX, srcY, w, h)

			// ダーティフラグが設定されていることを確認
			return dl.IsDirty()
		},
		gen.IntRange(0, 600),
		gen.IntRange(0, 400),
		gen.IntRange(0, 100),
		gen.IntRange(0, 100),
		gen.IntRange(1, 64),
		gen.IntRange(1, 64),
	))

	// Property 2.4.3: DrawingLayerのClearで内容がクリアされる
	properties.Property("DrawingLayerのClearで内容がクリアされる", prop.ForAll(
		func(w, h int) bool {
			if w <= 0 || h <= 0 || w > 256 || h > 256 {
				return true
			}

			lm := NewLayerManager()
			layerID := lm.GetNextLayerID()
			dl := NewDrawingLayer(layerID, 0, w, h)

			// ダーティフラグをクリア
			dl.SetDirty(false)

			// クリア
			dl.Clear()

			// ダーティフラグが設定されていることを確認
			return dl.IsDirty()
		},
		gen.IntRange(1, 256),
		gen.IntRange(1, 256),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty2_TextWriteCreatesTextLayer はTextWriteでTextLayerEntryが作成されることをテストする
// 要件 2.5: TextWriteが呼び出されたときに新しいText_Layerを作成する
func TestProperty2_TextWriteCreatesTextLayer(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 2.5.1: AddTextLayerでテキストレイヤーが追加される
	properties.Property("AddTextLayerでテキストレイヤーが追加される", prop.ForAll(
		func(picID, x, y int) bool {
			lm := NewLayerManager()
			pls := lm.GetOrCreatePictureLayerSet(picID)

			initialCount := pls.GetTextLayerCount()

			// テキストレイヤーを追加
			layerID := lm.GetNextLayerID()
			zOrderOffset := pls.GetNextTextZOffset()
			tl := NewTextLayerEntry(layerID, picID, x, y, "test text", zOrderOffset)
			pls.AddTextLayer(tl)

			// レイヤー数が増加していることを確認
			return pls.GetTextLayerCount() == initialCount+1
		},
		gen.IntRange(0, 255),
		gen.IntRange(0, 640),
		gen.IntRange(0, 480),
	))

	// Property 2.5.2: 追加されたテキストレイヤーが正しい属性を持つ
	properties.Property("追加されたテキストレイヤーが正しい属性を持つ", prop.ForAll(
		func(picID, x, y, textIndex int) bool {
			texts := []string{"Hello", "World", "Test", "テスト", "日本語"}
			text := texts[textIndex%len(texts)]

			lm := NewLayerManager()
			pls := lm.GetOrCreatePictureLayerSet(picID)

			// テキストレイヤーを追加
			layerID := lm.GetNextLayerID()
			zOrderOffset := pls.GetNextTextZOffset()
			tl := NewTextLayerEntry(layerID, picID, x, y, text, zOrderOffset)
			pls.AddTextLayer(tl)

			// 追加されたレイヤーを取得
			addedLayer := pls.GetTextLayer(layerID)
			if addedLayer == nil {
				return false
			}

			// 属性が正しいことを確認
			posX, posY := addedLayer.GetPosition()

			return posX == x && posY == y &&
				addedLayer.GetText() == text && addedLayer.GetPicID() == picID
		},
		gen.IntRange(0, 255),
		gen.IntRange(0, 640),
		gen.IntRange(0, 480),
		gen.IntRange(0, 4),
	))

	// Property 2.5.3: 複数のテキストレイヤーを追加できる
	properties.Property("複数のテキストレイヤーを追加できる", prop.ForAll(
		func(textCount int) bool {
			if textCount <= 0 || textCount > 50 {
				return true
			}

			lm := NewLayerManager()
			pls := lm.GetOrCreatePictureLayerSet(0)

			var layerIDs []int

			// 複数のテキストレイヤーを追加
			for i := 0; i < textCount; i++ {
				layerID := lm.GetNextLayerID()
				zOrderOffset := pls.GetNextTextZOffset()
				tl := NewTextLayerEntry(layerID, 0, i*10, i*10, "text", zOrderOffset)
				pls.AddTextLayer(tl)
				layerIDs = append(layerIDs, layerID)
			}

			// すべてのレイヤーが追加されていることを確認
			if pls.GetTextLayerCount() != textCount {
				return false
			}

			// 各レイヤーが取得できることを確認
			for _, id := range layerIDs {
				if pls.GetTextLayer(id) == nil {
					return false
				}
			}

			return true
		},
		gen.IntRange(1, 50),
	))

	// Property 2.5.4: テキストレイヤー追加後にFullDirtyが設定される
	properties.Property("テキストレイヤー追加後にFullDirtyが設定される", prop.ForAll(
		func(picID, x, y int) bool {
			lm := NewLayerManager()
			pls := lm.GetOrCreatePictureLayerSet(picID)

			// ダーティフラグをクリア
			pls.ClearAllDirtyFlags()

			// テキストレイヤーを追加
			layerID := lm.GetNextLayerID()
			zOrderOffset := pls.GetNextTextZOffset()
			tl := NewTextLayerEntry(layerID, picID, x, y, "test", zOrderOffset)
			pls.AddTextLayer(tl)

			// FullDirtyが設定されていることを確認
			return pls.FullDirty
		},
		gen.IntRange(0, 255),
		gen.IntRange(0, 640),
		gen.IntRange(0, 480),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty2_WindowCloseDeletesAllLayers はウィンドウが閉じられたときにすべてのレイヤーが削除されることをテストする
// 要件 2.6: ウィンドウが閉じられたときにそのウィンドウに属するすべてのレイヤーを削除する
func TestProperty2_WindowCloseDeletesAllLayers(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 2.6.1: DeletePictureLayerSetでピクチャーのすべてのレイヤーが削除される
	properties.Property("DeletePictureLayerSetでピクチャーのすべてのレイヤーが削除される", prop.ForAll(
		func(picID, castCount, textCount int) bool {
			if castCount < 0 || castCount > 20 || textCount < 0 || textCount > 20 {
				return true
			}

			lm := NewLayerManager()
			pls := lm.GetOrCreatePictureLayerSet(picID)

			// 背景レイヤーを設定
			bgID := lm.GetNextLayerID()
			bg := NewBackgroundLayer(bgID, picID, nil)
			pls.SetBackground(bg)

			// 描画レイヤーを設定
			dlID := lm.GetNextLayerID()
			dl := NewDrawingLayer(dlID, picID, 640, 480)
			pls.SetDrawing(dl)

			// キャストレイヤーを追加
			for i := 0; i < castCount; i++ {
				layerID := lm.GetNextLayerID()
				zOrderOffset := pls.GetNextCastZOffset()
				cl := NewCastLayer(layerID, i, picID, 0, i*10, i*10, 0, 0, 32, 32, zOrderOffset)
				pls.AddCastLayer(cl)
			}

			// テキストレイヤーを追加
			for i := 0; i < textCount; i++ {
				layerID := lm.GetNextLayerID()
				zOrderOffset := pls.GetNextTextZOffset()
				tl := NewTextLayerEntry(layerID, picID, i*10, i*10, "test", zOrderOffset)
				pls.AddTextLayer(tl)
			}

			// PictureLayerSetが存在することを確認
			if lm.GetPictureLayerSet(picID) == nil {
				return false
			}

			// PictureLayerSetを削除
			lm.DeletePictureLayerSet(picID)

			// PictureLayerSetが存在しないことを確認
			return lm.GetPictureLayerSet(picID) == nil
		},
		gen.IntRange(0, 255),
		gen.IntRange(0, 20),
		gen.IntRange(0, 20),
	))

	// Property 2.6.2: 削除後に他のピクチャーのレイヤーは影響を受けない
	properties.Property("削除後に他のピクチャーのレイヤーは影響を受けない", prop.ForAll(
		func(picID1, picID2, castCount int) bool {
			if picID1 == picID2 || castCount <= 0 || castCount > 10 {
				return true
			}

			lm := NewLayerManager()

			// 2つのピクチャーにレイヤーを追加
			for _, picID := range []int{picID1, picID2} {
				pls := lm.GetOrCreatePictureLayerSet(picID)
				for i := 0; i < castCount; i++ {
					layerID := lm.GetNextLayerID()
					zOrderOffset := pls.GetNextCastZOffset()
					cl := NewCastLayer(layerID, i, picID, 0, i*10, i*10, 0, 0, 32, 32, zOrderOffset)
					pls.AddCastLayer(cl)
				}
			}

			// 最初のピクチャーを削除
			lm.DeletePictureLayerSet(picID1)

			// 2番目のピクチャーが影響を受けていないことを確認
			pls2 := lm.GetPictureLayerSet(picID2)
			return pls2 != nil && pls2.GetCastLayerCount() == castCount
		},
		gen.IntRange(0, 127),
		gen.IntRange(128, 255),
		gen.IntRange(1, 10),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty2_ClearTextLayersRemovesAllTextLayers はClearTextLayersですべてのテキストレイヤーが削除されることをテストする
func TestProperty2_ClearTextLayersRemovesAllTextLayers(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 2.6.3: ClearTextLayersですべてのテキストレイヤーが削除される
	properties.Property("ClearTextLayersですべてのテキストレイヤーが削除される", prop.ForAll(
		func(textCount int) bool {
			if textCount <= 0 || textCount > 50 {
				return true
			}

			lm := NewLayerManager()
			pls := lm.GetOrCreatePictureLayerSet(0)

			// テキストレイヤーを追加
			for i := 0; i < textCount; i++ {
				layerID := lm.GetNextLayerID()
				zOrderOffset := pls.GetNextTextZOffset()
				tl := NewTextLayerEntry(layerID, 0, i*10, i*10, "test", zOrderOffset)
				pls.AddTextLayer(tl)
			}

			// テキストレイヤーが追加されていることを確認
			if pls.GetTextLayerCount() != textCount {
				return false
			}

			// すべてのテキストレイヤーをクリア
			pls.ClearTextLayers()

			// テキストレイヤーがすべて削除されていることを確認
			return pls.GetTextLayerCount() == 0
		},
		gen.IntRange(1, 50),
	))

	// Property 2.6.4: ClearTextLayers後もキャストレイヤーは影響を受けない
	properties.Property("ClearTextLayers後もキャストレイヤーは影響を受けない", prop.ForAll(
		func(castCount, textCount int) bool {
			if castCount <= 0 || castCount > 20 || textCount <= 0 || textCount > 20 {
				return true
			}

			lm := NewLayerManager()
			pls := lm.GetOrCreatePictureLayerSet(0)

			// キャストレイヤーを追加
			for i := 0; i < castCount; i++ {
				layerID := lm.GetNextLayerID()
				zOrderOffset := pls.GetNextCastZOffset()
				cl := NewCastLayer(layerID, i, 0, 0, i*10, i*10, 0, 0, 32, 32, zOrderOffset)
				pls.AddCastLayer(cl)
			}

			// テキストレイヤーを追加
			for i := 0; i < textCount; i++ {
				layerID := lm.GetNextLayerID()
				zOrderOffset := pls.GetNextTextZOffset()
				tl := NewTextLayerEntry(layerID, 0, i*10, i*10, "test", zOrderOffset)
				pls.AddTextLayer(tl)
			}

			// テキストレイヤーをクリア
			pls.ClearTextLayers()

			// キャストレイヤーが影響を受けていないことを確認
			return pls.GetCastLayerCount() == castCount
		},
		gen.IntRange(1, 20),
		gen.IntRange(1, 20),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty2_ClearCastLayersRemovesAllCastLayers はClearCastLayersですべてのキャストレイヤーが削除されることをテストする
func TestProperty2_ClearCastLayersRemovesAllCastLayers(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 2.6.5: ClearCastLayersですべてのキャストレイヤーが削除される
	properties.Property("ClearCastLayersですべてのキャストレイヤーが削除される", prop.ForAll(
		func(castCount int) bool {
			if castCount <= 0 || castCount > 50 {
				return true
			}

			lm := NewLayerManager()
			pls := lm.GetOrCreatePictureLayerSet(0)

			// キャストレイヤーを追加
			for i := 0; i < castCount; i++ {
				layerID := lm.GetNextLayerID()
				zOrderOffset := pls.GetNextCastZOffset()
				cl := NewCastLayer(layerID, i, 0, 0, i*10, i*10, 0, 0, 32, 32, zOrderOffset)
				pls.AddCastLayer(cl)
			}

			// キャストレイヤーが追加されていることを確認
			if pls.GetCastLayerCount() != castCount {
				return false
			}

			// すべてのキャストレイヤーをクリア
			pls.ClearCastLayers()

			// キャストレイヤーがすべて削除されていることを確認
			return pls.GetCastLayerCount() == 0
		},
		gen.IntRange(1, 50),
	))

	// Property 2.6.6: ClearCastLayers後もテキストレイヤーは影響を受けない
	properties.Property("ClearCastLayers後もテキストレイヤーは影響を受けない", prop.ForAll(
		func(castCount, textCount int) bool {
			if castCount <= 0 || castCount > 20 || textCount <= 0 || textCount > 20 {
				return true
			}

			lm := NewLayerManager()
			pls := lm.GetOrCreatePictureLayerSet(0)

			// キャストレイヤーを追加
			for i := 0; i < castCount; i++ {
				layerID := lm.GetNextLayerID()
				zOrderOffset := pls.GetNextCastZOffset()
				cl := NewCastLayer(layerID, i, 0, 0, i*10, i*10, 0, 0, 32, 32, zOrderOffset)
				pls.AddCastLayer(cl)
			}

			// テキストレイヤーを追加
			var textLayerIDs []int
			for i := 0; i < textCount; i++ {
				layerID := lm.GetNextLayerID()
				zOrderOffset := pls.GetNextTextZOffset()
				tl := NewTextLayerEntry(layerID, 0, i*10, i*10, "test", zOrderOffset)
				pls.AddTextLayer(tl)
				textLayerIDs = append(textLayerIDs, layerID)
			}

			// キャストレイヤーをクリア
			pls.ClearCastLayers()

			// テキストレイヤーが影響を受けていないことを確認
			return pls.GetTextLayerCount() == textCount
		},
		gen.IntRange(1, 20),
		gen.IntRange(1, 20),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty2_LayerOperationSequence はレイヤー操作のシーケンスをテストする
// 複合的なシナリオでレイヤー操作が正しく動作することを確認
func TestProperty2_LayerOperationSequence(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 2.7.1: 複数の操作後もレイヤー状態が一貫している
	properties.Property("複数の操作後もレイヤー状態が一貫している", prop.ForAll(
		func(operations []int) bool {
			if len(operations) == 0 || len(operations) > 50 {
				return true
			}

			lm := NewLayerManager()
			pls := lm.GetOrCreatePictureLayerSet(0)

			expectedCastCount := 0
			expectedTextCount := 0
			nextCastID := 0

			for _, op := range operations {
				switch op % 6 {
				case 0:
					// キャストレイヤーを追加
					if expectedCastCount < 100 {
						layerID := lm.GetNextLayerID()
						zOrderOffset := pls.GetNextCastZOffset()
						cl := NewCastLayer(layerID, nextCastID, 0, 0, 0, 0, 0, 0, 32, 32, zOrderOffset)
						pls.AddCastLayer(cl)
						nextCastID++
						expectedCastCount++
					}
				case 1:
					// テキストレイヤーを追加
					if expectedTextCount < 100 {
						layerID := lm.GetNextLayerID()
						zOrderOffset := pls.GetNextTextZOffset()
						tl := NewTextLayerEntry(layerID, 0, 0, 0, "test", zOrderOffset)
						pls.AddTextLayer(tl)
						expectedTextCount++
					}
				case 2:
					// キャストレイヤーを削除
					if expectedCastCount > 0 && len(pls.Casts) > 0 {
						castID := pls.Casts[0].GetCastID()
						if pls.RemoveCastLayer(castID) {
							expectedCastCount--
						}
					}
				case 3:
					// テキストレイヤーを削除
					if expectedTextCount > 0 && len(pls.Texts) > 0 {
						layerID := pls.Texts[0].GetID()
						if pls.RemoveTextLayer(layerID) {
							expectedTextCount--
						}
					}
				case 4:
					// キャストレイヤーの位置を更新
					if len(pls.Casts) > 0 {
						pls.Casts[0].SetPosition(op%640, op%480)
					}
				case 5:
					// テキストレイヤーの位置を更新
					if len(pls.Texts) > 0 {
						pls.Texts[0].SetPosition(op%640, op%480)
					}
				}
			}

			// レイヤー数が期待通りであることを確認
			return pls.GetCastLayerCount() == expectedCastCount &&
				pls.GetTextLayerCount() == expectedTextCount
		},
		gen.SliceOfN(50, gen.IntRange(0, 1000)),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty2_RemoveTextLayerByID はRemoveTextLayerがレイヤーIDで正しく削除することをテストする
func TestProperty2_RemoveTextLayerByID(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 2.8.1: RemoveTextLayerでテキストレイヤーが削除される
	properties.Property("RemoveTextLayerでテキストレイヤーが削除される", prop.ForAll(
		func(x, y int) bool {
			lm := NewLayerManager()
			pls := lm.GetOrCreatePictureLayerSet(0)

			// テキストレイヤーを追加
			layerID := lm.GetNextLayerID()
			zOrderOffset := pls.GetNextTextZOffset()
			tl := NewTextLayerEntry(layerID, 0, x, y, "test", zOrderOffset)
			pls.AddTextLayer(tl)

			// レイヤーが追加されていることを確認
			if pls.GetTextLayer(layerID) == nil {
				return false
			}

			// レイヤーを削除
			result := pls.RemoveTextLayer(layerID)

			// 削除が成功し、レイヤーが存在しないことを確認
			return result && pls.GetTextLayer(layerID) == nil
		},
		gen.IntRange(0, 640),
		gen.IntRange(0, 480),
	))

	// Property 2.8.2: 存在しないレイヤーIDの削除はfalseを返す
	properties.Property("存在しないテキストレイヤーIDの削除はfalseを返す", prop.ForAll(
		func(layerID, nonExistentID int) bool {
			if layerID == nonExistentID {
				return true
			}

			lm := NewLayerManager()
			pls := lm.GetOrCreatePictureLayerSet(0)

			// テキストレイヤーを追加
			zOrderOffset := pls.GetNextTextZOffset()
			tl := NewTextLayerEntry(layerID, 0, 0, 0, "test", zOrderOffset)
			pls.AddTextLayer(tl)

			// 存在しないIDで削除を試みる
			result := pls.RemoveTextLayer(nonExistentID)

			// 削除が失敗することを確認
			return !result
		},
		gen.IntRange(0, 50),
		gen.IntRange(51, 100),
	))

	// Property 2.8.3: テキストレイヤー削除後にFullDirtyが設定される
	properties.Property("テキストレイヤー削除後にFullDirtyが設定される", prop.ForAll(
		func(x, y int) bool {
			lm := NewLayerManager()
			pls := lm.GetOrCreatePictureLayerSet(0)

			// テキストレイヤーを追加
			layerID := lm.GetNextLayerID()
			zOrderOffset := pls.GetNextTextZOffset()
			tl := NewTextLayerEntry(layerID, 0, x, y, "test", zOrderOffset)
			// 画像を設定して境界を確定
			img := ebiten.NewImage(100, 20)
			tl.SetImage(img)
			pls.AddTextLayer(tl)

			// ダーティフラグをクリア
			pls.ClearAllDirtyFlags()

			// レイヤーを削除
			pls.RemoveTextLayer(layerID)

			// FullDirtyが設定されていることを確認
			return pls.FullDirty
		},
		gen.IntRange(0, 640),
		gen.IntRange(0, 480),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
