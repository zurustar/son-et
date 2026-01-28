// Package graphics provides sprite-based rendering system.
package graphics

import (
	"image/color"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// ShapeType は図形の種類を表す
type ShapeType int

const (
	// ShapeTypeLine は線を表す
	ShapeTypeLine ShapeType = iota
	// ShapeTypeRect は矩形（輪郭のみ）を表す
	ShapeTypeRect
	// ShapeTypeFillRect は塗りつぶし矩形を表す
	ShapeTypeFillRect
	// ShapeTypeCircle は円（輪郭のみ）を表す
	ShapeTypeCircle
	// ShapeTypeFillCircle は塗りつぶし円を表す
	ShapeTypeFillCircle
)

// ShapeSprite は図形とスプライトを組み合わせたラッパー構造体
// 要件 9.1: 線を描画したスプライトを作成できる
// 要件 9.2: 矩形を描画したスプライトを作成できる
// 要件 9.3: 塗りつぶし矩形を描画したスプライトを作成できる
type ShapeSprite struct {
	sprite *Sprite // 基盤となるスプライト

	// 図形情報
	shapeType ShapeType   // 図形の種類
	color     color.Color // 描画色
	lineSize  int         // 線の太さ

	// 図形固有のパラメータ
	// 線: (x1, y1) -> (x2, y2)
	// 矩形: (x1, y1) が左上、(x2, y2) が右下
	// 円: (x1, y1) が中心、radius が半径
	x1, y1   int
	x2, y2   int
	radius   int
	fillMode int // 0=輪郭のみ, 2=塗りつぶし

	// 描画先情報
	picID int // 描画先ピクチャーID
	destX int // 描画先X座標
	destY int // 描画先Y座標

	mu sync.RWMutex
}

// ShapeSpriteManager はShapeSpriteを管理する
type ShapeSpriteManager struct {
	shapeSprites  map[int][]*ShapeSprite // picID -> ShapeSprites
	spriteManager *SpriteManager
	mu            sync.RWMutex
	nextID        int
}

// NewShapeSpriteManager は新しいShapeSpriteManagerを作成する
func NewShapeSpriteManager(sm *SpriteManager) *ShapeSpriteManager {
	return &ShapeSpriteManager{
		shapeSprites:  make(map[int][]*ShapeSprite),
		spriteManager: sm,
		nextID:        1,
	}
}

// CreateLineSprite は線を描画したスプライトを作成する
// 要件 9.1: 線を描画したスプライトを作成できる
func (ssm *ShapeSpriteManager) CreateLineSprite(
	picID int,
	x1, y1, x2, y2 int,
	lineColor color.Color,
	lineSize int,
	zOrder int,
) *ShapeSprite {
	ssm.mu.Lock()
	defer ssm.mu.Unlock()

	// 線の境界ボックスを計算
	minX, maxX := x1, x2
	if x1 > x2 {
		minX, maxX = x2, x1
	}
	minY, maxY := y1, y2
	if y1 > y2 {
		minY, maxY = y2, y1
	}

	// 線の太さを考慮したサイズ
	halfLine := lineSize / 2
	if halfLine < 1 {
		halfLine = 1
	}
	width := maxX - minX + lineSize*2
	height := maxY - minY + lineSize*2
	if width < 1 {
		width = lineSize * 2
	}
	if height < 1 {
		height = lineSize * 2
	}

	// スプライト用の画像を作成
	img := ebiten.NewImage(width, height)

	// 線を描画（ローカル座標に変換）
	localX1 := float32(x1 - minX + halfLine)
	localY1 := float32(y1 - minY + halfLine)
	localX2 := float32(x2 - minX + halfLine)
	localY2 := float32(y2 - minY + halfLine)

	vector.StrokeLine(img, localX1, localY1, localX2, localY2, float32(lineSize), lineColor, false)

	// スプライトを作成
	sprite := ssm.spriteManager.CreateSprite(img)
	sprite.SetPosition(float64(minX-halfLine), float64(minY-halfLine))
	sprite.SetZOrder(zOrder)
	sprite.SetVisible(true)

	ss := &ShapeSprite{
		sprite:    sprite,
		shapeType: ShapeTypeLine,
		color:     lineColor,
		lineSize:  lineSize,
		x1:        x1,
		y1:        y1,
		x2:        x2,
		y2:        y2,
		picID:     picID,
		destX:     minX - halfLine,
		destY:     minY - halfLine,
	}

	ssm.shapeSprites[picID] = append(ssm.shapeSprites[picID], ss)
	ssm.nextID++

	return ss
}

// CreateLineSpriteWithParent は線を描画したスプライトを作成し、親スプライトを設定する
// 要件 9.1: 線を描画したスプライトを作成できる
// 要件 14.2: ウインドウ内のスプライトをウインドウの子スプライトとして管理する
func (ssm *ShapeSpriteManager) CreateLineSpriteWithParent(
	picID int,
	x1, y1, x2, y2 int,
	lineColor color.Color,
	lineSize int,
	zOrder int,
	parent *Sprite,
) *ShapeSprite {
	ss := ssm.CreateLineSprite(picID, x1, y1, x2, y2, lineColor, lineSize, zOrder)
	if ss != nil && parent != nil {
		ss.SetParent(parent)
	}
	return ss
}

// CreateRectSprite は矩形を描画したスプライトを作成する
// 要件 9.2: 矩形を描画したスプライトを作成できる
func (ssm *ShapeSpriteManager) CreateRectSprite(
	picID int,
	x1, y1, x2, y2 int,
	rectColor color.Color,
	lineSize int,
	zOrder int,
) *ShapeSprite {
	ssm.mu.Lock()
	defer ssm.mu.Unlock()

	// 座標を正規化
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}

	// 線の太さを考慮したサイズ
	halfLine := lineSize / 2
	if halfLine < 1 {
		halfLine = 1
	}
	width := x2 - x1 + lineSize*2
	height := y2 - y1 + lineSize*2

	// スプライト用の画像を作成
	img := ebiten.NewImage(width, height)

	// 矩形を描画（ローカル座標）
	localX := float32(halfLine)
	localY := float32(halfLine)
	localW := float32(x2 - x1)
	localH := float32(y2 - y1)

	vector.StrokeRect(img, localX, localY, localW, localH, float32(lineSize), rectColor, false)

	// スプライトを作成
	sprite := ssm.spriteManager.CreateSprite(img)
	sprite.SetPosition(float64(x1-halfLine), float64(y1-halfLine))
	sprite.SetZOrder(zOrder)
	sprite.SetVisible(true)

	ss := &ShapeSprite{
		sprite:    sprite,
		shapeType: ShapeTypeRect,
		color:     rectColor,
		lineSize:  lineSize,
		x1:        x1,
		y1:        y1,
		x2:        x2,
		y2:        y2,
		fillMode:  0,
		picID:     picID,
		destX:     x1 - halfLine,
		destY:     y1 - halfLine,
	}

	ssm.shapeSprites[picID] = append(ssm.shapeSprites[picID], ss)
	ssm.nextID++

	return ss
}

// CreateRectSpriteWithParent は矩形を描画したスプライトを作成し、親スプライトを設定する
// 要件 9.2: 矩形を描画したスプライトを作成できる
// 要件 14.2: ウインドウ内のスプライトをウインドウの子スプライトとして管理する
func (ssm *ShapeSpriteManager) CreateRectSpriteWithParent(
	picID int,
	x1, y1, x2, y2 int,
	rectColor color.Color,
	lineSize int,
	zOrder int,
	parent *Sprite,
) *ShapeSprite {
	ss := ssm.CreateRectSprite(picID, x1, y1, x2, y2, rectColor, lineSize, zOrder)
	if ss != nil && parent != nil {
		ss.SetParent(parent)
	}
	return ss
}

// CreateFillRectSprite は塗りつぶし矩形を描画したスプライトを作成する
// 要件 9.3: 塗りつぶし矩形を描画したスプライトを作成できる
func (ssm *ShapeSpriteManager) CreateFillRectSprite(
	picID int,
	x1, y1, x2, y2 int,
	fillColor color.Color,
	zOrder int,
) *ShapeSprite {
	ssm.mu.Lock()
	defer ssm.mu.Unlock()

	// 座標を正規化
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}

	width := x2 - x1
	height := y2 - y1
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}

	// スプライト用の画像を作成
	img := ebiten.NewImage(width, height)

	// 塗りつぶし矩形を描画
	vector.DrawFilledRect(img, 0, 0, float32(width), float32(height), fillColor, false)

	// スプライトを作成
	sprite := ssm.spriteManager.CreateSprite(img)
	sprite.SetPosition(float64(x1), float64(y1))
	sprite.SetZOrder(zOrder)
	sprite.SetVisible(true)

	ss := &ShapeSprite{
		sprite:    sprite,
		shapeType: ShapeTypeFillRect,
		color:     fillColor,
		x1:        x1,
		y1:        y1,
		x2:        x2,
		y2:        y2,
		fillMode:  2,
		picID:     picID,
		destX:     x1,
		destY:     y1,
	}

	ssm.shapeSprites[picID] = append(ssm.shapeSprites[picID], ss)
	ssm.nextID++

	return ss
}

// CreateFillRectSpriteWithParent は塗りつぶし矩形を描画したスプライトを作成し、親スプライトを設定する
// 要件 9.3: 塗りつぶし矩形を描画したスプライトを作成できる
// 要件 14.2: ウインドウ内のスプライトをウインドウの子スプライトとして管理する
func (ssm *ShapeSpriteManager) CreateFillRectSpriteWithParent(
	picID int,
	x1, y1, x2, y2 int,
	fillColor color.Color,
	zOrder int,
	parent *Sprite,
) *ShapeSprite {
	ss := ssm.CreateFillRectSprite(picID, x1, y1, x2, y2, fillColor, zOrder)
	if ss != nil && parent != nil {
		ss.SetParent(parent)
	}
	return ss
}

// CreateCircleSprite は円を描画したスプライトを作成する
func (ssm *ShapeSpriteManager) CreateCircleSprite(
	picID int,
	cx, cy, radius int,
	circleColor color.Color,
	lineSize int,
	zOrder int,
) *ShapeSprite {
	ssm.mu.Lock()
	defer ssm.mu.Unlock()

	if radius <= 0 {
		return nil
	}

	// 線の太さを考慮したサイズ
	halfLine := lineSize / 2
	if halfLine < 1 {
		halfLine = 1
	}
	size := (radius + halfLine) * 2

	// スプライト用の画像を作成
	img := ebiten.NewImage(size, size)

	// 円を描画（中心をローカル座標に変換）
	localCX := float32(radius + halfLine)
	localCY := float32(radius + halfLine)

	vector.StrokeCircle(img, localCX, localCY, float32(radius), float32(lineSize), circleColor, false)

	// スプライトを作成
	sprite := ssm.spriteManager.CreateSprite(img)
	sprite.SetPosition(float64(cx-radius-halfLine), float64(cy-radius-halfLine))
	sprite.SetZOrder(zOrder)
	sprite.SetVisible(true)

	ss := &ShapeSprite{
		sprite:    sprite,
		shapeType: ShapeTypeCircle,
		color:     circleColor,
		lineSize:  lineSize,
		x1:        cx,
		y1:        cy,
		radius:    radius,
		fillMode:  0,
		picID:     picID,
		destX:     cx - radius - halfLine,
		destY:     cy - radius - halfLine,
	}

	ssm.shapeSprites[picID] = append(ssm.shapeSprites[picID], ss)
	ssm.nextID++

	return ss
}

// CreateCircleSpriteWithParent は円を描画したスプライトを作成し、親スプライトを設定する
// 要件 14.2: ウインドウ内のスプライトをウインドウの子スプライトとして管理する
func (ssm *ShapeSpriteManager) CreateCircleSpriteWithParent(
	picID int,
	cx, cy, radius int,
	circleColor color.Color,
	lineSize int,
	zOrder int,
	parent *Sprite,
) *ShapeSprite {
	ss := ssm.CreateCircleSprite(picID, cx, cy, radius, circleColor, lineSize, zOrder)
	if ss != nil && parent != nil {
		ss.SetParent(parent)
	}
	return ss
}

// CreateFillCircleSprite は塗りつぶし円を描画したスプライトを作成する
func (ssm *ShapeSpriteManager) CreateFillCircleSprite(
	picID int,
	cx, cy, radius int,
	fillColor color.Color,
	zOrder int,
) *ShapeSprite {
	ssm.mu.Lock()
	defer ssm.mu.Unlock()

	if radius <= 0 {
		return nil
	}

	size := radius * 2

	// スプライト用の画像を作成
	img := ebiten.NewImage(size, size)

	// 塗りつぶし円を描画
	localCX := float32(radius)
	localCY := float32(radius)

	vector.DrawFilledCircle(img, localCX, localCY, float32(radius), fillColor, false)

	// スプライトを作成
	sprite := ssm.spriteManager.CreateSprite(img)
	sprite.SetPosition(float64(cx-radius), float64(cy-radius))
	sprite.SetZOrder(zOrder)
	sprite.SetVisible(true)

	ss := &ShapeSprite{
		sprite:    sprite,
		shapeType: ShapeTypeFillCircle,
		color:     fillColor,
		x1:        cx,
		y1:        cy,
		radius:    radius,
		fillMode:  2,
		picID:     picID,
		destX:     cx - radius,
		destY:     cy - radius,
	}

	ssm.shapeSprites[picID] = append(ssm.shapeSprites[picID], ss)
	ssm.nextID++

	return ss
}

// CreateFillCircleSpriteWithParent は塗りつぶし円を描画したスプライトを作成し、親スプライトを設定する
// 要件 14.2: ウインドウ内のスプライトをウインドウの子スプライトとして管理する
func (ssm *ShapeSpriteManager) CreateFillCircleSpriteWithParent(
	picID int,
	cx, cy, radius int,
	fillColor color.Color,
	zOrder int,
	parent *Sprite,
) *ShapeSprite {
	ss := ssm.CreateFillCircleSprite(picID, cx, cy, radius, fillColor, zOrder)
	if ss != nil && parent != nil {
		ss.SetParent(parent)
	}
	return ss
}

// GetShapeSprites はピクチャIDに関連するすべてのShapeSpriteを取得する
func (ssm *ShapeSpriteManager) GetShapeSprites(picID int) []*ShapeSprite {
	ssm.mu.RLock()
	defer ssm.mu.RUnlock()

	sprites := ssm.shapeSprites[picID]
	if sprites == nil {
		return nil
	}

	result := make([]*ShapeSprite, len(sprites))
	copy(result, sprites)
	return result
}

// RemoveShapeSprite は指定されたShapeSpriteを削除する
func (ssm *ShapeSpriteManager) RemoveShapeSprite(ss *ShapeSprite) {
	if ss == nil {
		return
	}

	ssm.mu.Lock()
	defer ssm.mu.Unlock()

	// スプライトを削除
	if ss.sprite != nil {
		ssm.spriteManager.RemoveSprite(ss.sprite.ID())
	}

	// リストから削除
	sprites := ssm.shapeSprites[ss.picID]
	for i, s := range sprites {
		if s == ss {
			ssm.shapeSprites[ss.picID] = append(sprites[:i], sprites[i+1:]...)
			break
		}
	}

	// リストが空になったら削除
	if len(ssm.shapeSprites[ss.picID]) == 0 {
		delete(ssm.shapeSprites, ss.picID)
	}
}

// RemoveShapeSpritesByPicID はピクチャIDに関連するすべてのShapeSpriteを削除する
func (ssm *ShapeSpriteManager) RemoveShapeSpritesByPicID(picID int) {
	ssm.mu.Lock()
	defer ssm.mu.Unlock()

	sprites := ssm.shapeSprites[picID]
	for _, ss := range sprites {
		if ss.sprite != nil {
			ssm.spriteManager.RemoveSprite(ss.sprite.ID())
		}
	}
	delete(ssm.shapeSprites, picID)
}

// Clear はすべてのShapeSpriteを削除する
func (ssm *ShapeSpriteManager) Clear() {
	ssm.mu.Lock()
	defer ssm.mu.Unlock()

	for picID, sprites := range ssm.shapeSprites {
		for _, ss := range sprites {
			if ss.sprite != nil {
				ssm.spriteManager.RemoveSprite(ss.sprite.ID())
			}
		}
		delete(ssm.shapeSprites, picID)
	}
}

// Count は登録されているShapeSpriteの総数を返す
func (ssm *ShapeSpriteManager) Count() int {
	ssm.mu.RLock()
	defer ssm.mu.RUnlock()

	count := 0
	for _, sprites := range ssm.shapeSprites {
		count += len(sprites)
	}
	return count
}

// ShapeSprite methods

// GetSprite は基盤となるスプライトを返す
func (ss *ShapeSprite) GetSprite() *Sprite {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return ss.sprite
}

// GetShapeType は図形の種類を返す
func (ss *ShapeSprite) GetShapeType() ShapeType {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return ss.shapeType
}

// GetColor は描画色を返す
func (ss *ShapeSprite) GetColor() color.Color {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return ss.color
}

// GetLineSize は線の太さを返す
func (ss *ShapeSprite) GetLineSize() int {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return ss.lineSize
}

// GetPicID は描画先ピクチャーIDを返す
func (ss *ShapeSprite) GetPicID() int {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return ss.picID
}

// GetLineCoords は線の座標を返す（線の場合のみ有効）
func (ss *ShapeSprite) GetLineCoords() (x1, y1, x2, y2 int) {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return ss.x1, ss.y1, ss.x2, ss.y2
}

// GetRectCoords は矩形の座標を返す（矩形の場合のみ有効）
func (ss *ShapeSprite) GetRectCoords() (x1, y1, x2, y2 int) {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return ss.x1, ss.y1, ss.x2, ss.y2
}

// GetCircleParams は円のパラメータを返す（円の場合のみ有効）
func (ss *ShapeSprite) GetCircleParams() (cx, cy, radius int) {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return ss.x1, ss.y1, ss.radius
}

// GetFillMode は塗りつぶしモードを返す
func (ss *ShapeSprite) GetFillMode() int {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return ss.fillMode
}

// SetPosition は描画位置を更新する
func (ss *ShapeSprite) SetPosition(x, y int) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	ss.destX = x
	ss.destY = y
	if ss.sprite != nil {
		ss.sprite.SetPosition(float64(x), float64(y))
	}
}

// SetZOrder はZ順序を更新する
func (ss *ShapeSprite) SetZOrder(z int) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	if ss.sprite != nil {
		ss.sprite.SetZOrder(z)
	}
}

// SetVisible は可視性を更新する
func (ss *ShapeSprite) SetVisible(visible bool) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	if ss.sprite != nil {
		ss.sprite.SetVisible(visible)
	}
}

// SetParent は親スプライトを設定する
// ウインドウ内の図形描画で使用
func (ss *ShapeSprite) SetParent(parent *Sprite) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	if ss.sprite != nil {
		ss.sprite.SetParent(parent)
	}
}
