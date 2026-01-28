// Package graphics provides sprite-based rendering system.
package graphics

import (
	"image"
	"sort"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

// Sprite は汎用スプライト
// すべての描画要素（ウインドウ、ピクチャ、キャスト、文字、図形）の基盤となる
type Sprite struct {
	id      int
	image   *ebiten.Image
	x, y    float64
	zOrder  int
	visible bool
	alpha   float64
	parent  *Sprite
	dirty   bool // 再描画が必要かどうか
}

// NewSprite は新しいスプライトを作成する
func NewSprite(id int, img *ebiten.Image) *Sprite {
	return &Sprite{
		id:      id,
		image:   img,
		x:       0,
		y:       0,
		zOrder:  0,
		visible: true,
		alpha:   1.0,
		parent:  nil,
		dirty:   true,
	}
}

// ID はスプライトのIDを返す
func (s *Sprite) ID() int {
	return s.id
}

// Image はスプライトの画像を返す
func (s *Sprite) Image() *ebiten.Image {
	return s.image
}

// SetImage はスプライトの画像を設定する
func (s *Sprite) SetImage(img *ebiten.Image) {
	s.image = img
	s.dirty = true
}

// Position はスプライトの位置を返す
func (s *Sprite) Position() (float64, float64) {
	return s.x, s.y
}

// SetPosition はスプライトの位置を設定する
func (s *Sprite) SetPosition(x, y float64) {
	s.x = x
	s.y = y
	s.dirty = true
}

// ZOrder はスプライトのZ順序を返す
func (s *Sprite) ZOrder() int {
	return s.zOrder
}

// SetZOrder はスプライトのZ順序を設定する
func (s *Sprite) SetZOrder(z int) {
	s.zOrder = z
	s.dirty = true
}

// Visible はスプライトの可視性を返す
func (s *Sprite) Visible() bool {
	return s.visible
}

// SetVisible はスプライトの可視性を設定する
func (s *Sprite) SetVisible(v bool) {
	s.visible = v
	s.dirty = true
}

// Alpha はスプライトの透明度を返す（0.0〜1.0）
func (s *Sprite) Alpha() float64 {
	return s.alpha
}

// SetAlpha はスプライトの透明度を設定する（0.0〜1.0）
func (s *Sprite) SetAlpha(a float64) {
	if a < 0 {
		a = 0
	}
	if a > 1 {
		a = 1
	}
	s.alpha = a
	s.dirty = true
}

// Parent はスプライトの親を返す
func (s *Sprite) Parent() *Sprite {
	return s.parent
}

// SetParent はスプライトの親を設定する
func (s *Sprite) SetParent(p *Sprite) {
	s.parent = p
	s.dirty = true
}

// IsDirty は再描画が必要かどうかを返す
func (s *Sprite) IsDirty() bool {
	return s.dirty
}

// ClearDirty はdirtyフラグをクリアする
func (s *Sprite) ClearDirty() {
	s.dirty = false
}

// AbsolutePosition は親を考慮した絶対位置を返す
func (s *Sprite) AbsolutePosition() (float64, float64) {
	x, y := s.x, s.y
	if s.parent != nil {
		px, py := s.parent.AbsolutePosition()
		x += px
		y += py
	}
	return x, y
}

// EffectiveAlpha は親を考慮した実効透明度を返す
func (s *Sprite) EffectiveAlpha() float64 {
	alpha := s.alpha
	if s.parent != nil {
		alpha *= s.parent.EffectiveAlpha()
	}
	return alpha
}

// IsEffectivelyVisible は親を考慮した実効可視性を返す
func (s *Sprite) IsEffectivelyVisible() bool {
	if !s.visible {
		return false
	}
	if s.parent != nil {
		return s.parent.IsEffectivelyVisible()
	}
	return true
}

// Bounds はスプライトの境界を返す
func (s *Sprite) Bounds() image.Rectangle {
	if s.image == nil {
		return image.Rectangle{}
	}
	return s.image.Bounds()
}

// Size はスプライトのサイズを返す
func (s *Sprite) Size() (int, int) {
	if s.image == nil {
		return 0, 0
	}
	bounds := s.image.Bounds()
	return bounds.Dx(), bounds.Dy()
}

// SpriteManager はスプライトを管理する
type SpriteManager struct {
	mu       sync.RWMutex
	sprites  map[int]*Sprite
	nextID   int
	sorted   []*Sprite // Z順序でソート済みのキャッシュ
	needSort bool
}

// NewSpriteManager は新しいSpriteManagerを作成する
func NewSpriteManager() *SpriteManager {
	return &SpriteManager{
		sprites:  make(map[int]*Sprite),
		nextID:   1,
		needSort: true,
	}
}

// CreateSprite は新しいスプライトを作成して登録する
func (sm *SpriteManager) CreateSprite(img *ebiten.Image) *Sprite {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	s := NewSprite(sm.nextID, img)
	sm.sprites[s.id] = s
	sm.nextID++
	sm.needSort = true
	return s
}

// CreateSpriteWithSize は指定サイズの空のスプライトを作成する
func (sm *SpriteManager) CreateSpriteWithSize(width, height int) *Sprite {
	img := ebiten.NewImage(width, height)
	return sm.CreateSprite(img)
}

// GetSprite はIDでスプライトを取得する
func (sm *SpriteManager) GetSprite(id int) *Sprite {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.sprites[id]
}

// RemoveSprite はスプライトを削除する
func (sm *SpriteManager) RemoveSprite(id int) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.sprites, id)
	sm.needSort = true
}

// Clear はすべてのスプライトを削除する
func (sm *SpriteManager) Clear() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.sprites = make(map[int]*Sprite)
	sm.sorted = nil
	sm.needSort = true
}

// Count は登録されているスプライトの数を返す
func (sm *SpriteManager) Count() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.sprites)
}

// sortSprites はスプライトをZ順序でソートする
func (sm *SpriteManager) sortSprites() {
	sm.sorted = make([]*Sprite, 0, len(sm.sprites))
	for _, s := range sm.sprites {
		sm.sorted = append(sm.sorted, s)
	}
	sort.Slice(sm.sorted, func(i, j int) bool {
		return sm.sorted[i].zOrder < sm.sorted[j].zOrder
	})
	sm.needSort = false
}

// Draw はすべての可視スプライトをZ順序で描画する
func (sm *SpriteManager) Draw(screen *ebiten.Image) {
	sm.mu.Lock()
	if sm.needSort {
		sm.sortSprites()
	}
	sorted := sm.sorted
	sm.mu.Unlock()

	for _, s := range sorted {
		if !s.IsEffectivelyVisible() || s.image == nil {
			continue
		}

		op := &ebiten.DrawImageOptions{}
		x, y := s.AbsolutePosition()
		op.GeoM.Translate(x, y)

		alpha := s.EffectiveAlpha()
		if alpha < 1.0 {
			op.ColorScale.ScaleAlpha(float32(alpha))
		}

		screen.DrawImage(s.image, op)
	}
}

// MarkNeedSort はソートが必要であることをマークする
func (sm *SpriteManager) MarkNeedSort() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.needSort = true
}
