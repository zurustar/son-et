// Package graphics provides text sprite creation with anti-aliasing removal.
package graphics

import (
	"image"
	"image/color"
	"image/draw"
	"sync"

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
	// BackMode は背景モード（0=背景あり/不透明, 1=透明）
	BackMode int
	// Width は画像の幅（0の場合は自動計算）
	Width int
	// Height は画像の高さ（0の場合は自動計算）
	Height int
	// X はテキストのX座標
	X int
	// Y はテキストのベースラインY座標
	Y int
}

// TextSpriteOptionsWithBackground はテキストスプライト作成のオプション（背景画像付き）
type TextSpriteOptionsWithBackground struct {
	TextSpriteOptions
	// BackgroundImage は背景画像（マスク方式で使用）
	BackgroundImage *image.RGBA
}

// CreateTextSpriteImage は差分抽出方式でテキストスプライト用の画像を作成する
// アンチエイリアスの影響を除去し、透過画像を返す
// 特殊ケース: TextColorとBgColorが同じ場合は「消しゴム」モードとして、
// 不透明な背景色で塗りつぶした画像を返す（下のレイヤーを隠すため）
func CreateTextSpriteImage(opts TextSpriteOptions) *image.RGBA {
	return CreateTextSpriteImageWithBackground(TextSpriteOptionsWithBackground{
		TextSpriteOptions: opts,
		BackgroundImage:   nil,
	})
}

// CreateTextSpriteImageWithBackground はマスク方式でテキストスプライト用の画像を作成する
// 背景画像が提供された場合、マスク方式を使用して完全に不透明なスプライトを作成する
// これにより、同じ位置に異なる色のテキストを重ねても、下のテキストが透けない
//
// BackMode対応:
// - BackMode=0（背景あり/不透明）: 背景色で塗りつぶした上にテキストを描画し、不透明なスプライトを作成
// - BackMode=1（透明）: 従来の差分抽出方式で透明背景のスプライトを作成
//
// マスク方式のアルゴリズム:
// 1. マスク作成: 黒字で白背景に描画（塗る場所を決定）
// 2. 透明度付き色画像: マスクから指定色で透明度付き画像を作成
// 3. 背景と合成: 透明度付き画像を背景と合成して不透明な画像にする
// 4. 結果: 不透明なスプライトなので、何度重ねても下が透けない
func CreateTextSpriteImageWithBackground(opts TextSpriteOptionsWithBackground) *image.RGBA {
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

	// 背景色とテキスト色を取得
	bgColor := opts.BgColor
	if bgColor == nil {
		bgColor = color.White
	}
	textColor := opts.TextColor
	if textColor == nil {
		textColor = color.Black
	}

	// BackMode=0（背景あり/不透明）の場合、背景色で塗りつぶした不透明なスプライトを作成
	if opts.BackMode == 0 {
		return createOpaqueTextSprite(opts.Face, opts.Text, opts.X, opts.Y, width, height, textColor, bgColor)
	}

	// BackMode=1（透明）の場合

	// 背景画像が提供された場合、マスク方式を使用
	if opts.BackgroundImage != nil {
		return createTextSpriteWithMask(opts.Face, opts.Text, opts.X, opts.Y, width, height, textColor, opts.BackgroundImage)
	}

	// 背景画像がない場合は従来の差分抽出方式を使用
	// TextColorとBgColorが同じかどうかをチェック（「消しゴム」モード判定）
	// 同じ場合は、不透明な背景色で塗りつぶした画像を返す
	if colorsEqual(textColor, bgColor) {
		// 消しゴムモード: 不透明な背景色で塗りつぶした画像を作成
		// これにより、下のレイヤーにあるテキストを「隠す」ことができる
		result := image.NewRGBA(image.Rect(0, 0, width, height))
		draw.Draw(result, result.Bounds(), image.NewUniform(bgColor), image.Point{}, draw.Src)
		return result
	}

	// 通常モード: 差分抽出方式

	// 1. 背景色で塗りつぶした画像を作成
	bgImg := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(bgImg, bgImg.Bounds(), image.NewUniform(bgColor), image.Point{}, draw.Src)

	// 2. 背景のコピーを保持
	bgCopy := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(bgCopy, bgCopy.Bounds(), bgImg, image.Point{}, draw.Src)

	// 3. テキストを描画
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

// createOpaqueTextSprite はBackMode=0（背景あり/不透明）用の不透明なテキストスプライトを作成する
// 背景色で塗りつぶした上にテキストを描画し、完全に不透明なスプライトを返す
func createOpaqueTextSprite(face font.Face, text string, x, y, width, height int, textColor, bgColor color.Color) *image.RGBA {
	// 背景色で塗りつぶした画像を作成
	result := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(result, result.Bounds(), image.NewUniform(bgColor), image.Point{}, draw.Src)

	// テキストを描画
	drawer := &font.Drawer{
		Dst:  result,
		Src:  image.NewUniform(textColor),
		Face: face,
		Dot:  fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)},
	}
	drawer.DrawString(text)

	return result
}

// createTextSpriteWithMask はマスク方式でテキストスプライト画像を作成する
// この方式では、テキストを背景と合成して不透明なスプライトを作成する
// これにより、同じ位置に異なる色のテキストを重ねても、下のテキストが透けない
func createTextSpriteWithMask(face font.Face, text string, x, y, width, height int, textColor color.Color, background *image.RGBA) *image.RGBA {
	// Step 1: マスク作成（黒字で白背景に描画）
	maskImg := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(maskImg, maskImg.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)

	drawer := &font.Drawer{
		Dst:  maskImg,
		Src:  image.NewUniform(color.Black),
		Face: face,
		Dot:  fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)},
	}
	drawer.DrawString(text)

	// Step 2: 透明度付き色画像を作成
	// マスクの黒い部分 → 指定色（不透明）
	// マスクのグレー部分 → 指定色（半透明）
	// マスクの白い部分 → 透明
	alphaImg := createAlphaColorImage(maskImg, textColor)

	// Step 3: 背景と合成して不透明な画像にする
	result := blendWithBackground(alphaImg, background, width, height)

	return result
}

// createAlphaColorImage はマスクから透明度付きの色画像を作成する
// - マスクの黒い部分 → 指定色（不透明）
// - マスクのグレー部分 → 指定色（半透明）
// - マスクの白い部分 → 透明
func createAlphaColorImage(mask *image.RGBA, textColor color.Color) *image.RGBA {
	bounds := mask.Bounds()
	result := image.NewRGBA(bounds)

	tr, tg, tb, _ := textColor.RGBA()
	tr8 := uint8(tr >> 8)
	tg8 := uint8(tg >> 8)
	tb8 := uint8(tb >> 8)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			maskPixel := mask.At(x, y)
			r, _, _, _ := maskPixel.RGBA()

			// マスク値: 白(65535)=透明、黒(0)=不透明
			alpha := uint8((65535 - r) >> 8)

			if alpha > 0 {
				// Premultiplied alpha
				premulR := uint8(uint16(tr8) * uint16(alpha) / 255)
				premulG := uint8(uint16(tg8) * uint16(alpha) / 255)
				premulB := uint8(uint16(tb8) * uint16(alpha) / 255)
				result.Set(x, y, color.RGBA{R: premulR, G: premulG, B: premulB, A: alpha})
			} else {
				result.Set(x, y, color.RGBA{0, 0, 0, 0})
			}
		}
	}

	return result
}

// blendWithBackground は透明度付き画像を背景と合成して不透明な画像にする
// 結果は文字部分のみ不透明、それ以外は透明
func blendWithBackground(alphaImg, background *image.RGBA, width, height int) *image.RGBA {
	result := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			_, _, _, a := alphaImg.At(x, y).RGBA()
			alpha := a >> 8

			if alpha > 0 {
				// 透明度付き画像のピクセル（premultiplied）
				sr, sg, sb, sa := alphaImg.At(x, y).RGBA()
				// 背景のピクセル
				br, bg, bb, _ := background.At(x, y).RGBA()

				// Porter-Duff Over: result = src + dst * (1 - srcAlpha)
				// srcはpremultipliedなのでそのまま使用
				invAlpha := 65535 - sa
				finalR := uint8((sr + br*invAlpha/65535) >> 8)
				finalG := uint8((sg + bg*invAlpha/65535) >> 8)
				finalB := uint8((sb + bb*invAlpha/65535) >> 8)

				// 不透明なピクセルとして設定
				result.Set(x, y, color.RGBA{R: finalR, G: finalG, B: finalB, A: 255})
			} else {
				// 透明部分はそのまま透明
				result.Set(x, y, color.RGBA{0, 0, 0, 0})
			}
		}
	}

	return result
}

// colorsEqual は2つの色が等しいかどうかを判定する
func colorsEqual(c1, c2 color.Color) bool {
	r1, g1, b1, _ := c1.RGBA()
	r2, g2, b2, _ := c2.RGBA()
	return r1 == r2 && g1 == g2 && b1 == b2
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

// TextSprite はテキストとスプライトを組み合わせたラッパー構造体
// 要件 5.1〜5.5: テキスト描画をスプライトとして表現する
type TextSprite struct {
	sprite *Sprite // スプライト（テキストの画像を保持）

	// テキスト情報
	picID int    // 描画先ピクチャーID
	text  string // テキスト内容
	x, y  int    // 描画位置

	// テキスト設定
	textColor color.Color // テキスト色
	bgColor   color.Color // 背景色（差分抽出用）
	face      font.Face   // フォントフェイス

	mu sync.RWMutex
}

// TextSpriteManager はTextSpriteを管理する
type TextSpriteManager struct {
	textSprites   map[int][]*TextSprite // picID -> TextSprites（同じピクチャに複数のテキストがある場合）
	spriteManager *SpriteManager
	mu            sync.RWMutex
	nextID        int         // 内部ID管理
	zOffsets      map[int]int // picID -> 次のZ順序オフセット
}

// NewTextSpriteManager は新しいTextSpriteManagerを作成する
func NewTextSpriteManager(sm *SpriteManager) *TextSpriteManager {
	return &TextSpriteManager{
		textSprites:   make(map[int][]*TextSprite),
		spriteManager: sm,
		nextID:        1,
		zOffsets:      make(map[int]int),
	}
}

// GetNextZOffset は指定されたピクチャIDの次のZ順序オフセットを取得してインクリメントする
func (tsm *TextSpriteManager) GetNextZOffset(picID int) int {
	tsm.mu.Lock()
	defer tsm.mu.Unlock()

	offset := tsm.zOffsets[picID]
	tsm.zOffsets[picID] = offset + 1
	return offset
}

// CreateTextSprite はテキストからTextSpriteを作成する
// 要件 5.1, 5.2: 背景色の上にテキストを描画し、差分を抽出する
func (tsm *TextSpriteManager) CreateTextSprite(
	picID int,
	x, y int,
	text string,
	textColor, bgColor color.Color,
	face font.Face,
	zOrder int,
) *TextSprite {
	tsm.mu.Lock()
	defer tsm.mu.Unlock()

	if text == "" || face == nil {
		return nil
	}

	// テキストスプライト用の画像を作成（差分抽出方式）
	opts := TextSpriteOptions{
		Text:      text,
		TextColor: textColor,
		Face:      face,
		BgColor:   bgColor,
		X:         0,
		Y:         getFontHeight(face),
	}

	img := CreateTextSpriteImage(opts)
	if img == nil {
		return nil
	}

	// スプライトを作成
	// 注意: zOrderパラメータは互換性のために残されているが、
	// 実際のZ順序はZ_Pathで管理される
	sprite := tsm.spriteManager.CreateSprite(ebiten.NewImageFromImage(img))
	sprite.SetPosition(float64(x), float64(y))
	sprite.SetVisible(true)

	ts := &TextSprite{
		sprite:    sprite,
		picID:     picID,
		text:      text,
		x:         x,
		y:         y,
		textColor: textColor,
		bgColor:   bgColor,
		face:      face,
	}

	// ピクチャIDごとにスプライトを管理
	tsm.textSprites[picID] = append(tsm.textSprites[picID], ts)
	tsm.nextID++

	return ts
}

// CreateTextSpriteWithParent はテキストからTextSpriteを作成し、親スプライトを設定する
// 要件 5.1, 5.2: 背景色の上にテキストを描画し、差分を抽出する
// 要件 14.2: ウインドウ内のスプライトをウインドウの子スプライトとして管理する
// 要件 2.2: PutTextが呼び出されたとき、現在のZ_Order_Counterを使用してLocal_Z_Orderを割り当てる
// 要件 1.4: 子スプライトが作成されたとき、親のZ_Pathを継承し、自身のLocal_Z_Orderを追加する
//
// BackMode対応:
// - backMode=0（背景あり/不透明）: 背景色で塗りつぶした不透明なスプライトを作成
// - backMode=1（透明）: マスク方式または差分抽出方式で透明背景のスプライトを作成
//
// マスク方式:
// 親スプライトから背景画像を取得し、マスク方式でテキストスプライトを作成する
// これにより、同じ位置に異なる色のテキストを重ねても、下のテキストが透けない
//
// 同じ位置への再描画:
// 同じピクチャーの同じ位置（または近い位置）に既存のTextSpriteがある場合、
// 古いTextSpriteを削除してから新しいTextSpriteを作成する
// これにより、オリジナルFILLYの「焼き付け」動作をエミュレートする
func (tsm *TextSpriteManager) CreateTextSpriteWithParent(
	picID int,
	x, y int,
	text string,
	textColor, bgColor color.Color,
	face font.Face,
	zOrder int,
	parent *Sprite,
	backMode int,
) *TextSprite {
	tsm.mu.Lock()
	defer tsm.mu.Unlock()

	if text == "" || face == nil {
		return nil
	}

	// 同じピクチャーの同じ位置（または近い位置）にある既存のTextSpriteを削除
	// これにより、オリジナルFILLYの「焼き付け」動作をエミュレートする
	// 許容範囲: 10ピクセル以内
	const positionTolerance = 10
	existingSprites := tsm.textSprites[picID]
	var spritesToRemove []*TextSprite
	for _, ts := range existingSprites {
		tsX, tsY := ts.x, ts.y
		if abs(tsX-x) <= positionTolerance && abs(tsY-y) <= positionTolerance {
			spritesToRemove = append(spritesToRemove, ts)
		}
	}
	// 削除対象のスプライトを削除（ロック内で直接削除）
	for _, ts := range spritesToRemove {
		if ts.sprite != nil {
			// 親から子を削除
			if ts.sprite.Parent() != nil {
				ts.sprite.Parent().RemoveChild(ts.sprite.ID())
			}
			tsm.spriteManager.RemoveSprite(ts.sprite.ID())
		}
		// リストから削除
		for i, s := range tsm.textSprites[picID] {
			if s == ts {
				tsm.textSprites[picID] = append(tsm.textSprites[picID][:i], tsm.textSprites[picID][i+1:]...)
				break
			}
		}
	}

	// テキストのサイズを計算
	bounds := measureText(face, text)
	width := bounds.Dx() + 10  // 余白を追加
	height := bounds.Dy() + 10 // 余白を追加
	fontHeight := getFontHeight(face)

	// 親スプライトから背景画像を抽出
	var backgroundImg *image.RGBA
	if parent != nil && parent.Image() != nil {
		backgroundImg = extractBackgroundRegion(parent.Image(), x, y, width, height)
	}

	// テキストスプライト用の画像を作成（マスク方式）
	opts := TextSpriteOptionsWithBackground{
		TextSpriteOptions: TextSpriteOptions{
			Text:      text,
			TextColor: textColor,
			Face:      face,
			BgColor:   bgColor,
			BackMode:  backMode,
			X:         0,
			Y:         fontHeight,
			Width:     width,
			Height:    height,
		},
		BackgroundImage: backgroundImg,
	}

	img := CreateTextSpriteImageWithBackground(opts)
	if img == nil {
		return nil
	}

	// スプライトを作成
	sprite := tsm.spriteManager.CreateSprite(ebiten.NewImageFromImage(img))
	sprite.SetPosition(float64(x), float64(y))
	sprite.SetVisible(true)

	ts := &TextSprite{
		sprite:    sprite,
		picID:     picID,
		text:      text,
		x:         x,
		y:         y,
		textColor: textColor,
		bgColor:   bgColor,
		face:      face,
	}

	// ピクチャIDごとにスプライトを管理
	tsm.textSprites[picID] = append(tsm.textSprites[picID], ts)
	tsm.nextID++

	// 親スプライトを設定
	if parent != nil {
		ts.sprite.SetParent(parent)
		parent.AddChild(ts.sprite)

		// 要件 2.2, 2.6: 操作順序でLocal_Z_Orderを割り当てる
		// 要件 1.4: 親のZ_Pathを継承し、自身のLocal_Z_Orderを追加する
		if parent.GetZPath() != nil {
			// 親のIDを使用してZOrderCounterから次のLocal_Z_Orderを取得
			localZOrder := tsm.spriteManager.GetZOrderCounter().GetNext(parent.ID())
			zPath := NewZPathFromParent(parent.GetZPath(), localZOrder)
			ts.sprite.SetZPath(zPath)
			tsm.spriteManager.MarkNeedSort()
		}
	}

	return ts
}

// extractBackgroundRegion は親スプライトの画像から指定領域を抽出する
// Ebitengineの制限により、ゲームループ開始前はピクセルを読み取れないため、
// その場合はnilを返す（呼び出し側で差分抽出方式にフォールバックする）
func extractBackgroundRegion(parentImg *ebiten.Image, x, y, width, height int) *image.RGBA {
	if parentImg == nil {
		return nil
	}

	// Ebitengineの制限: ゲームループ開始前はAt()を呼び出せない
	// その場合はpanicが発生するので、recoverで捕捉してnilを返す
	result, ok := tryExtractBackgroundRegion(parentImg, x, y, width, height)
	if !ok {
		return nil
	}
	return result
}

// tryExtractBackgroundRegion は親スプライトの画像から指定領域を抽出する（panic対応版）
func tryExtractBackgroundRegion(parentImg *ebiten.Image, x, y, width, height int) (result *image.RGBA, ok bool) {
	defer func() {
		if r := recover(); r != nil {
			// ゲームループ開始前の場合、nilを返す
			result = nil
			ok = false
		}
	}()

	parentBounds := parentImg.Bounds()

	// 領域が親画像の範囲外の場合は背景色で埋める
	result = image.NewRGBA(image.Rect(0, 0, width, height))

	// 親画像からピクセルをコピー
	for dy := 0; dy < height; dy++ {
		for dx := 0; dx < width; dx++ {
			srcX := x + dx
			srcY := y + dy

			if srcX >= parentBounds.Min.X && srcX < parentBounds.Max.X &&
				srcY >= parentBounds.Min.Y && srcY < parentBounds.Max.Y {
				// 親画像の範囲内
				r, g, b, a := parentImg.At(srcX, srcY).RGBA()
				result.Set(dx, dy, color.RGBA{
					R: uint8(r >> 8),
					G: uint8(g >> 8),
					B: uint8(b >> 8),
					A: uint8(a >> 8),
				})
			} else {
				// 範囲外は白で埋める
				result.Set(dx, dy, color.RGBA{255, 255, 255, 255})
			}
		}
	}

	return result, true
}

// GetTextSprites はピクチャIDに関連するすべてのTextSpriteを取得する
func (tsm *TextSpriteManager) GetTextSprites(picID int) []*TextSprite {
	tsm.mu.RLock()
	defer tsm.mu.RUnlock()

	sprites := tsm.textSprites[picID]
	if sprites == nil {
		return nil
	}

	// コピーを返す
	result := make([]*TextSprite, len(sprites))
	copy(result, sprites)
	return result
}

// RemoveTextSprite は指定されたTextSpriteを削除する
func (tsm *TextSpriteManager) RemoveTextSprite(ts *TextSprite) {
	if ts == nil {
		return
	}

	tsm.mu.Lock()
	defer tsm.mu.Unlock()

	// スプライトを削除
	if ts.sprite != nil {
		tsm.spriteManager.RemoveSprite(ts.sprite.ID())
	}

	// リストから削除
	sprites := tsm.textSprites[ts.picID]
	for i, s := range sprites {
		if s == ts {
			tsm.textSprites[ts.picID] = append(sprites[:i], sprites[i+1:]...)
			break
		}
	}

	// リストが空になったら削除
	if len(tsm.textSprites[ts.picID]) == 0 {
		delete(tsm.textSprites, ts.picID)
	}
}

// RemoveTextSpritesByPicID はピクチャIDに関連するすべてのTextSpriteを削除する
func (tsm *TextSpriteManager) RemoveTextSpritesByPicID(picID int) {
	tsm.mu.Lock()
	defer tsm.mu.Unlock()

	sprites := tsm.textSprites[picID]
	for _, ts := range sprites {
		if ts.sprite != nil {
			tsm.spriteManager.RemoveSprite(ts.sprite.ID())
		}
	}
	delete(tsm.textSprites, picID)
}

// Clear はすべてのTextSpriteを削除する
func (tsm *TextSpriteManager) Clear() {
	tsm.mu.Lock()
	defer tsm.mu.Unlock()

	for picID, sprites := range tsm.textSprites {
		for _, ts := range sprites {
			if ts.sprite != nil {
				tsm.spriteManager.RemoveSprite(ts.sprite.ID())
			}
		}
		delete(tsm.textSprites, picID)
	}
	// Z順序オフセットもクリア
	tsm.zOffsets = make(map[int]int)
}

// Count は登録されているTextSpriteの総数を返す
func (tsm *TextSpriteManager) Count() int {
	tsm.mu.RLock()
	defer tsm.mu.RUnlock()

	count := 0
	for _, sprites := range tsm.textSprites {
		count += len(sprites)
	}
	return count
}

// GetTextSpritesInRegion は指定されたピクチャーIDの指定領域内にあるTextSpriteを取得する
// MovePicで転送元の領域内にあるTextSpriteを取得するために使用
func (tsm *TextSpriteManager) GetTextSpritesInRegion(picID, srcX, srcY, width, height int) []*TextSprite {
	tsm.mu.RLock()
	defer tsm.mu.RUnlock()

	sprites := tsm.textSprites[picID]
	if sprites == nil {
		return nil
	}

	result := make([]*TextSprite, 0)
	for _, ts := range sprites {
		// TextSpriteの位置が転送領域内にあるかチェック
		tsX, tsY := ts.GetPosition()
		if tsX >= srcX && tsX < srcX+width && tsY >= srcY && tsY < srcY+height {
			result = append(result, ts)
		}
	}
	return result
}

// TextSprite methods

// GetSprite は基盤となるスプライトを返す
func (ts *TextSprite) GetSprite() *Sprite {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.sprite
}

// GetPicID は描画先ピクチャーIDを返す
func (ts *TextSprite) GetPicID() int {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.picID
}

// GetText はテキスト内容を返す
func (ts *TextSprite) GetText() string {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.text
}

// GetPosition は描画位置を返す
func (ts *TextSprite) GetPosition() (int, int) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.x, ts.y
}

// SetPosition は描画位置を更新する
func (ts *TextSprite) SetPosition(x, y int) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.x = x
	ts.y = y
	if ts.sprite != nil {
		ts.sprite.SetPosition(float64(x), float64(y))
	}
}

// SetVisible は可視性を更新する
func (ts *TextSprite) SetVisible(visible bool) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.sprite != nil {
		ts.sprite.SetVisible(visible)
	}
}

// SetParent は親スプライトを設定する
// ウインドウ内のテキスト描画で使用
func (ts *TextSprite) SetParent(parent *Sprite) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.sprite != nil {
		ts.sprite.SetParent(parent)
	}
}

// UpdateText はテキストを更新し、画像を再生成する
func (ts *TextSprite) UpdateText(text string) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.text == text {
		return
	}

	ts.text = text

	// 画像を再生成
	if ts.face != nil {
		opts := TextSpriteOptions{
			Text:      text,
			TextColor: ts.textColor,
			Face:      ts.face,
			BgColor:   ts.bgColor,
			X:         0,
			Y:         getFontHeight(ts.face),
		}

		img := CreateTextSpriteImage(opts)
		if img != nil && ts.sprite != nil {
			ts.sprite.SetImage(ebiten.NewImageFromImage(img))
		}
	}
}

// getFontHeight はフォントの高さを取得する
func getFontHeight(face font.Face) int {
	if face == nil {
		return 13 // デフォルト値
	}
	metrics := face.Metrics()
	return (metrics.Ascent + metrics.Descent).Ceil()
}

// abs は整数の絶対値を返す
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
