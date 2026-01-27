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

// TextLayer はテキスト描画のレイヤーを表す
// 背景との差分を取って文字部分だけを透明背景で保持する
type TextLayer struct {
	Image  *image.RGBA // 透明背景 + 文字
	PicID  int         // 描画先のピクチャーID
	X, Y   int         // ピクチャー内の描画位置
	Width  int         // レイヤーの幅
	Height int         // レイヤーの高さ
}

// TextLayerManager はテキストレイヤーを管理する
// TextWriteごとにレイヤーを作成し、最終描画時に合成する
type TextLayerManager struct {
	layers map[int][]*TextLayer // ピクチャーIDごとのレイヤーリスト
	mu     sync.RWMutex
}

// NewTextLayerManager は新しい TextLayerManager を作成する
func NewTextLayerManager() *TextLayerManager {
	return &TextLayerManager{
		layers: make(map[int][]*TextLayer),
	}
}

// AddLayer はレイヤーを追加する
func (m *TextLayerManager) AddLayer(layer *TextLayer) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.layers[layer.PicID] == nil {
		m.layers[layer.PicID] = make([]*TextLayer, 0)
	}
	m.layers[layer.PicID] = append(m.layers[layer.PicID], layer)
}

// GetLayers は指定されたピクチャーのレイヤーを取得する
func (m *TextLayerManager) GetLayers(picID int) []*TextLayer {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.layers[picID]
}

// ClearLayers は指定されたピクチャーのレイヤーをクリアする
func (m *TextLayerManager) ClearLayers(picID int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.layers, picID)
}

// ClearAllLayers はすべてのレイヤーをクリアする
func (m *TextLayerManager) ClearAllLayers() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.layers = make(map[int][]*TextLayer)
}

// CompositeToImage はレイヤーを合成してEbitengine画像を返す
// 背景画像の上にレイヤーを順番に重ねる
func (m *TextLayerManager) CompositeToImage(picID int, background *ebiten.Image) *ebiten.Image {
	m.mu.RLock()
	layers := m.layers[picID]
	m.mu.RUnlock()

	if len(layers) == 0 {
		return background
	}

	// 背景をRGBAにコピー
	bounds := background.Bounds()
	result := image.NewRGBA(bounds)

	// 背景をコピー
	for py := bounds.Min.Y; py < bounds.Max.Y; py++ {
		for px := bounds.Min.X; px < bounds.Max.X; px++ {
			result.Set(px, py, background.At(px, py))
		}
	}

	// レイヤーを順番に合成（後のレイヤーが上）
	for _, layer := range layers {
		destRect := image.Rect(layer.X, layer.Y, layer.X+layer.Width, layer.Y+layer.Height)
		draw.Draw(result, destRect, layer.Image, image.Point{}, draw.Over)
	}

	return ebiten.NewImageFromImage(result)
}

// CreateTextLayer はテキストレイヤーを作成する
// 背景に文字を描画し、差分を取って文字部分だけを透明背景で抽出する
func CreateTextLayer(
	background *image.RGBA,
	face font.Face,
	text string,
	x, y int,
	fontSize int,
	textColor color.Color,
	picID int,
) *TextLayer {
	bounds := background.Bounds()

	// テキストの境界を計算
	textBounds, _ := font.BoundString(face, text)
	textWidth := (textBounds.Max.X - textBounds.Min.X).Ceil() + 10
	textHeight := fontSize + 10

	// 描画位置
	destX := x
	destY := y

	// 範囲チェック
	if destX < 0 {
		destX = 0
	}
	if destY < 0 {
		destY = 0
	}
	if destX+textWidth > bounds.Dx() {
		textWidth = bounds.Dx() - destX
	}
	if destY+textHeight > bounds.Dy() {
		textHeight = bounds.Dy() - destY
	}

	if textWidth <= 0 || textHeight <= 0 {
		return nil
	}

	// 1. 背景をコピーして一時画像を作成
	tempImg := image.NewRGBA(image.Rect(0, 0, textWidth, textHeight))
	for py := 0; py < textHeight; py++ {
		for px := 0; px < textWidth; px++ {
			tempImg.Set(px, py, background.At(destX+px, destY+py))
		}
	}

	// 背景のコピーを保持（差分計算用）
	bgCopy := image.NewRGBA(image.Rect(0, 0, textWidth, textHeight))
	for py := 0; py < textHeight; py++ {
		for px := 0; px < textWidth; px++ {
			bgCopy.Set(px, py, tempImg.At(px, py))
		}
	}

	// 2. 一時画像に文字を描画（アルファブレンディング）
	drawer := &font.Drawer{
		Dst:  tempImg,
		Src:  image.NewUniform(textColor),
		Face: face,
		Dot:  fixed.Point26_6{X: fixed.I(0), Y: fixed.I(fontSize)},
	}

	// 文字ごとに描画（パニック回復付き）
	for _, r := range text {
		func() {
			defer func() { recover() }()
			drawer.DrawString(string(r))
		}()
	}

	// 3. 背景との差分を取って、文字部分だけを透明背景で抽出
	layerImg := image.NewRGBA(image.Rect(0, 0, textWidth, textHeight))
	for py := 0; py < textHeight; py++ {
		for px := 0; px < textWidth; px++ {
			bgPixel := bgCopy.At(px, py)
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
		PicID:  picID,
		X:      destX,
		Y:      destY,
		Width:  textWidth,
		Height: textHeight,
	}
}
