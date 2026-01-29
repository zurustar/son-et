package graphics

import (
	"image/color"
	"math"
	"testing"
	"testing/quick"

	"golang.org/x/image/font/basicfont"
)

// Property 1: スプライトID管理
// 任意のスプライト作成に対して、作成されたスプライトは一意のIDを持ち、そのIDで取得できる
// **Validates: Requirements 1.1, 3.1, 3.3**
func TestProperty_SpriteIDManagement(t *testing.T) {
	f := func(count uint8) bool {
		if count == 0 {
			return true
		}
		// 最大100個に制限
		n := int(count)
		if n > 100 {
			n = 100
		}

		sm := NewSpriteManager()
		ids := make(map[int]bool)
		sprites := make([]*Sprite, 0, n)

		// n個のスプライトを作成
		for i := 0; i < n; i++ {
			s := sm.CreateSpriteWithSize(10, 10)
			if s == nil {
				return false
			}
			// IDが一意であることを確認
			if ids[s.ID()] {
				return false // 重複ID
			}
			ids[s.ID()] = true
			sprites = append(sprites, s)
		}

		// すべてのスプライトがIDで取得できることを確認
		for _, s := range sprites {
			got := sm.GetSprite(s.ID())
			if got != s {
				return false
			}
		}

		return true
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

// Property 2: 親子関係の位置計算
// 任意の親子関係を持つスプライトに対して、子の絶対位置は親の位置と子の相対位置の和である
// **Validates: Requirements 2.1**
func TestProperty_ParentChildPosition(t *testing.T) {
	f := func(parentX, parentY, childX, childY int16) bool {
		// int16を使用して値の範囲を制限（-32768〜32767）
		px := float64(parentX)
		py := float64(parentY)
		cx := float64(childX)
		cy := float64(childY)

		parent := NewSprite(1, nil)
		parent.SetPosition(px, py)

		child := NewSprite(2, nil)
		child.SetPosition(cx, cy)
		child.SetParent(parent)

		absX, absY := child.AbsolutePosition()
		expectedX := px + cx
		expectedY := py + cy

		// 浮動小数点の誤差を考慮
		const epsilon = 1e-9
		return math.Abs(absX-expectedX) < epsilon && math.Abs(absY-expectedY) < epsilon
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

// Property 3: 親子関係の透明度計算
// 任意の親子関係を持つスプライトに対して、子の実効透明度は親の透明度と子の透明度の積である
// **Validates: Requirements 2.2**
func TestProperty_ParentChildAlpha(t *testing.T) {
	f := func(parentAlpha, childAlpha float64) bool {
		// 0.0〜1.0の範囲に正規化
		parentAlpha = math.Abs(math.Mod(parentAlpha, 1.0))
		childAlpha = math.Abs(math.Mod(childAlpha, 1.0))

		parent := NewSprite(1, nil)
		parent.SetAlpha(parentAlpha)

		child := NewSprite(2, nil)
		child.SetAlpha(childAlpha)
		child.SetParent(parent)

		effectiveAlpha := child.EffectiveAlpha()
		expectedAlpha := parentAlpha * childAlpha

		// 浮動小数点の誤差を考慮
		const epsilon = 1e-9
		return math.Abs(effectiveAlpha-expectedAlpha) < epsilon
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

// Property 4: 親子関係の可視性
// 任意の親子関係を持つスプライトに対して、親が非表示なら子も非表示として扱われる
// **Validates: Requirements 2.3, 4.3**
func TestProperty_ParentChildVisibility(t *testing.T) {
	f := func(parentVisible, childVisible bool) bool {
		parent := NewSprite(1, nil)
		parent.SetVisible(parentVisible)

		child := NewSprite(2, nil)
		child.SetVisible(childVisible)
		child.SetParent(parent)

		effectivelyVisible := child.IsEffectivelyVisible()

		// 親が非表示なら子も非表示
		if !parentVisible {
			return !effectivelyVisible
		}
		// 親が表示なら子の可視性に依存
		return effectivelyVisible == childVisible
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

// Property 5: Z_Pathによる描画順
// 任意のスプライト集合に対して、描画はZ_Pathの辞書順で行われる
// **Validates: Requirements 4.1**
func TestProperty_ZPathDrawing(t *testing.T) {
	f := func(zOrders []int8) bool {
		if len(zOrders) == 0 {
			return true
		}
		// 最大50個に制限
		if len(zOrders) > 50 {
			zOrders = zOrders[:50]
		}

		sm := NewSpriteManager()

		// スプライトを作成してZ_Pathを設定
		for _, z := range zOrders {
			s := sm.CreateSpriteWithSize(10, 10)
			s.SetZPath(NewZPath(int(z)))
		}

		// ソートを実行
		sm.mu.Lock()
		sm.sortSprites()
		sorted := sm.sorted
		sm.mu.Unlock()

		// Z_Pathが辞書順であることを確認
		for i := 1; i < len(sorted); i++ {
			prev := sorted[i-1].GetZPath()
			curr := sorted[i].GetZPath()
			if prev != nil && curr != nil {
				if !prev.Less(curr) && prev.Compare(curr) != 0 {
					return false
				}
			}
		}

		return true
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

// Property 6: テキスト差分抽出
// 任意のテキスト描画に対して、差分抽出後の画像は背景色のピクセルを含まない（透明になる）
// **Validates: Requirements 5.1, 5.2**
func TestProperty_TextDifferenceExtraction(t *testing.T) {
	// テスト用の背景色を複数用意
	bgColors := []color.RGBA{
		{255, 255, 255, 255}, // 白
		{255, 255, 200, 255}, // 薄い黄色
		{200, 200, 255, 255}, // 薄い青
		{255, 200, 200, 255}, // 薄いピンク
	}

	textColors := []color.RGBA{
		{0, 0, 0, 255},       // 黒
		{255, 0, 0, 255},     // 赤
		{0, 0, 255, 255},     // 青
		{255, 255, 255, 255}, // 白
	}

	for _, bgColor := range bgColors {
		for _, textColor := range textColors {
			// 背景色とテキスト色が同じ場合はスキップ
			if bgColor == textColor {
				continue
			}

			opts := TextSpriteOptions{
				Text:      "Test",
				TextColor: textColor,
				Face:      basicfont.Face7x13,
				BgColor:   bgColor,
				Width:     50,
				Height:    20,
				X:         5,
				Y:         15,
			}

			img := CreateTextSpriteImage(opts)
			if img == nil {
				t.Errorf("failed to create text sprite with bg=%v, text=%v", bgColor, textColor)
				continue
			}

			// 背景色のピクセルが残っていないことを確認
			bounds := img.Bounds()
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					pixel := img.At(x, y)
					r, g, b, a := pixel.RGBA()

					// 透明でないピクセルが背景色と一致していないことを確認
					if a > 0 {
						bgR, bgG, bgB, _ := bgColor.RGBA()
						if r == bgR && g == bgG && b == bgB {
							t.Errorf("found background color pixel at (%d,%d) with bg=%v", x, y, bgColor)
						}
					}
				}
			}
		}
	}
}

// Property 7: スプライト削除
// 任意のスプライト削除に対して、削除後はそのIDでスプライトを取得できない
// **Validates: Requirements 3.4**
func TestProperty_SpriteRemoval(t *testing.T) {
	f := func(count uint8) bool {
		if count == 0 {
			return true
		}
		n := int(count)
		if n > 50 {
			n = 50
		}

		sm := NewSpriteManager()
		ids := make([]int, 0, n)

		// スプライトを作成
		for i := 0; i < n; i++ {
			s := sm.CreateSpriteWithSize(10, 10)
			ids = append(ids, s.ID())
		}

		// 半分を削除
		for i := 0; i < n/2; i++ {
			sm.RemoveSprite(ids[i])
		}

		// 削除されたスプライトは取得できない
		for i := 0; i < n/2; i++ {
			if sm.GetSprite(ids[i]) != nil {
				return false
			}
		}

		// 削除されていないスプライトは取得できる
		for i := n / 2; i < n; i++ {
			if sm.GetSprite(ids[i]) == nil {
				return false
			}
		}

		return true
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}
