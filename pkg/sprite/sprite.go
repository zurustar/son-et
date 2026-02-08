// Package sprite provides sprite-based rendering system with slice-based draw ordering.
package sprite

import (
	"fmt"
	"image"
	"strings"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

// Sprite は汎用スプライト（スライスベースの描画順序）
// すべての描画要素（ウインドウ、ピクチャ、キャスト、文字、図形）の基盤となる
//
// 描画順序はchildrenスライスの順序で決定される:
// - スライスの先頭 = 最背面（最初に描画）
// - スライスの末尾 = 最前面（最後に描画）
type Sprite struct {
	id      int
	image   *ebiten.Image
	x, y    float64
	visible bool
	alpha   float64
	dirty   bool // 再描画が必要かどうか

	// 親子関係
	parent   *Sprite   // 親スプライトへのポインタ（nilの場合はルート）
	children []*Sprite // 子スプライトのスライス（順序 = 描画順序）
}

// NewSprite は新しいスプライトを作成する
func NewSprite(id int, img *ebiten.Image) *Sprite {
	return &Sprite{
		id:       id,
		image:    img,
		x:        0,
		y:        0,
		visible:  true,
		alpha:    1.0,
		parent:   nil,
		dirty:    true,
		children: make([]*Sprite, 0),
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

// GetChildren は子スプライトのリストを返す
func (s *Sprite) GetChildren() []*Sprite {
	return s.children
}

// HasChildren は子スプライトを持っているかどうかを返す
func (s *Sprite) HasChildren() bool {
	return len(s.children) > 0
}

// IsDirty は再描画が必要かどうかを返す
func (s *Sprite) IsDirty() bool {
	return s.dirty
}

// ClearDirty はdirtyフラグをクリアする
func (s *Sprite) ClearDirty() {
	s.dirty = false
}

// AddChild は子スプライトをスライスの末尾に追加する（最前面に配置）
// 要件 2.5: 子スプライトを追加するとき、スライスの末尾に追加する
func (s *Sprite) AddChild(child *Sprite) {
	if child == nil {
		return
	}
	// 既に別の親がある場合は削除
	if child.parent != nil && child.parent != s {
		child.parent.RemoveChild(child.id)
	}
	child.parent = s
	s.children = append(s.children, child)
}

// RemoveChild は子スプライトを削除する
func (s *Sprite) RemoveChild(childID int) {
	for i, child := range s.children {
		if child.id == childID {
			child.parent = nil
			s.children = append(s.children[:i], s.children[i+1:]...)
			return
		}
	}
}

// BringToFront はスプライトを最前面に移動する（スライス末尾に移動）
// 要件 12.1: スプライトを最前面に移動するメソッドを提供する
func (s *Sprite) BringToFront() {
	if s.parent == nil {
		return
	}
	parent := s.parent
	// 現在の位置を見つけて削除
	for i, child := range parent.children {
		if child.id == s.id {
			parent.children = append(parent.children[:i], parent.children[i+1:]...)
			break
		}
	}
	// 末尾に追加
	parent.children = append(parent.children, s)
}

// SendToBack はスプライトを最背面に移動する（スライス先頭に移動）
// 要件 12.2: スプライトを最背面に移動するメソッドを提供する
func (s *Sprite) SendToBack() {
	if s.parent == nil {
		return
	}
	parent := s.parent
	// 現在の位置を見つけて削除
	for i, child := range parent.children {
		if child.id == s.id {
			parent.children = append(parent.children[:i], parent.children[i+1:]...)
			break
		}
	}
	// 先頭に挿入
	parent.children = append([]*Sprite{s}, parent.children...)
}

// AbsolutePosition は親を考慮した絶対位置を返す
// 要件 2.1: 親の位置を加算して絶対位置を計算する
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
// 要件 2.2: 親の透明度を乗算して実効透明度を計算する
func (s *Sprite) EffectiveAlpha() float64 {
	alpha := s.alpha
	if s.parent != nil {
		alpha *= s.parent.EffectiveAlpha()
	}
	return alpha
}

// IsEffectivelyVisible は親を考慮した実効可視性を返す
// 要件 2.3: 親が非表示の場合、子も非表示として扱う
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

// ChildIndex は子スプライトのインデックスを返す（見つからない場合は-1）
func (s *Sprite) ChildIndex(childID int) int {
	for i, child := range s.children {
		if child.id == childID {
			return i
		}
	}
	return -1
}

// ============================================================================
// SpriteManager（スプライト管理）
// ============================================================================

// SpriteManager はスプライトを管理する（スライスベースの描画順序）
type SpriteManager struct {
	mu      sync.RWMutex
	sprites map[int]*Sprite // ID -> Sprite
	roots   []*Sprite       // ルートスプライト（ウインドウ）のスライス
	nextID  int
}

// NewSpriteManager は新しいSpriteManagerを作成する
func NewSpriteManager() *SpriteManager {
	return &SpriteManager{
		sprites: make(map[int]*Sprite),
		roots:   make([]*Sprite, 0),
		nextID:  1,
	}
}

// CreateSprite は新しいスプライトを作成する
// parentがnilの場合、ルートスプライトとして作成される
func (sm *SpriteManager) CreateSprite(img *ebiten.Image, parent *Sprite) *Sprite {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	s := NewSprite(sm.nextID, img)
	sm.sprites[s.id] = s
	sm.nextID++

	if parent != nil {
		parent.AddChild(s)
	} else {
		sm.roots = append(sm.roots, s)
	}

	return s
}

// CreateRootSprite はルートスプライト（ウインドウ）を作成する
func (sm *SpriteManager) CreateRootSprite(img *ebiten.Image) *Sprite {
	return sm.CreateSprite(img, nil)
}

// CreateSpriteWithSize は指定サイズの空のスプライトを作成する
func (sm *SpriteManager) CreateSpriteWithSize(width, height int, parent *Sprite) *Sprite {
	img := ebiten.NewImage(width, height)
	return sm.CreateSprite(img, parent)
}

// GetSprite はIDでスプライトを取得する
func (sm *SpriteManager) GetSprite(id int) *Sprite {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.sprites[id]
}

// DeleteSprite はスプライトを削除する
// 要件 3.4: スプライト削除時に子スプライトも削除する
func (sm *SpriteManager) DeleteSprite(id int) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	s := sm.sprites[id]
	if s == nil {
		return
	}

	// 子スプライトを再帰的に削除
	sm.deleteChildrenRecursive(s)

	// 親から削除
	if s.parent != nil {
		s.parent.RemoveChild(id)
	} else {
		// ルートスプライトから削除
		for i, root := range sm.roots {
			if root.id == id {
				sm.roots = append(sm.roots[:i], sm.roots[i+1:]...)
				break
			}
		}
	}

	delete(sm.sprites, id)
}

// deleteChildrenRecursive は子スプライトを再帰的に削除する（内部用）
func (sm *SpriteManager) deleteChildrenRecursive(s *Sprite) {
	for _, child := range s.children {
		sm.deleteChildrenRecursive(child)
		delete(sm.sprites, child.id)
	}
	s.children = nil
}

// Clear はすべてのスプライトを削除する
// 要件 3.5: すべてのスプライトをクリアする
func (sm *SpriteManager) Clear() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.sprites = make(map[int]*Sprite)
	sm.roots = make([]*Sprite, 0)
}

// Count は登録されているスプライトの数を返す
func (sm *SpriteManager) Count() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.sprites)
}

// GetRoots はルートスプライトのリストを返す
func (sm *SpriteManager) GetRoots() []*Sprite {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	result := make([]*Sprite, len(sm.roots))
	copy(result, sm.roots)
	return result
}

// ============================================================================
// 描画（Drawing）
// ============================================================================

// Draw はすべての可視スプライトを描画する
// 要件 9.1, 9.2: スライス順序で描画（先頭が最背面、末尾が最前面）
// 要件 10.1: 親スプライトを先に描画し、その後に子スプライトを描画する
func (sm *SpriteManager) Draw(screen *ebiten.Image) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// デバッグモードが有効かどうかを取得
	debugEnabled := IsDebugMode()

	// ルートスプライトを順番に描画
	for _, root := range sm.roots {
		sm.drawSprite(screen, root, 0, 0, 1.0, debugEnabled, true)
	}
}

// drawSprite はスプライトとその子を再帰的に描画する
// 要件 10.4: 親が非表示の場合は子も描画しない
// デバッグモードが有効な場合、各スプライトの描画直後にデバッグ情報を描画する
func (sm *SpriteManager) drawSprite(screen *ebiten.Image, s *Sprite, parentX, parentY, parentAlpha float64, debugEnabled bool, isRoot bool) {
	if !s.visible {
		return
	}

	// 絶対位置と実効透明度を計算
	absX := parentX + s.x
	absY := parentY + s.y
	effectiveAlpha := parentAlpha * s.alpha

	// スプライトを描画（画像がある場合のみ）
	if s.image != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(absX, absY)
		if effectiveAlpha < 1.0 {
			op.ColorScale.ScaleAlpha(float32(effectiveAlpha))
		}
		screen.DrawImage(s.image, op)
	}

	// デバッグオーバーレイを描画（スプライトと同じ階層で描画）
	if debugEnabled {
		globalDebugOverlay.drawSpriteOverlayInline(screen, s, absX, absY, isRoot)
	}

	// 子スプライトを順番に描画（スライス順序 = 描画順序）
	for _, child := range s.children {
		sm.drawSprite(screen, child, absX, absY, effectiveAlpha, debugEnabled, false)
	}
}

// ============================================================================
// デバッグ支援（Debug Support）
// ============================================================================

// PrintHierarchy はスプライト階層をツリー形式で出力する
// 要件 20.1: スプライト階層をツリー形式で出力できる
func (sm *SpriteManager) PrintHierarchy() string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var sb strings.Builder

	// ルートスプライトをツリー形式で出力
	for _, root := range sm.roots {
		sm.printSpriteTree(&sb, root, 0)
	}

	return sb.String()
}

// printSpriteTree はスプライトツリーを再帰的に出力する（内部用）
func (sm *SpriteManager) printSpriteTree(sb *strings.Builder, s *Sprite, depth int) {
	indent := strings.Repeat("  ", depth)
	visibility := "visible"
	if !s.visible {
		visibility = "hidden"
	}

	// 画像サイズ
	imgSize := "nil"
	if s.image != nil {
		bounds := s.image.Bounds()
		imgSize = fmt.Sprintf("%dx%d", bounds.Dx(), bounds.Dy())
	}

	fmt.Fprintf(sb, "%s- Sprite %d: pos=(%.0f,%.0f) size=%s (%s) children=%d\n",
		indent, s.id, s.x, s.y, imgSize, visibility, len(s.children))

	for _, child := range s.children {
		sm.printSpriteTree(sb, child, depth+1)
	}
}

// PrintDrawOrder は描画順序のリストを出力する
// 要件 20.2: 描画順序のリストを出力できる
func (sm *SpriteManager) PrintDrawOrder() string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var sb strings.Builder
	sb.WriteString("Draw Order:\n")

	order := 1
	for _, root := range sm.roots {
		order = sm.printDrawOrderRecursive(&sb, root, order)
	}

	return sb.String()
}

// printDrawOrderRecursive は描画順序を再帰的に出力する（内部用）
func (sm *SpriteManager) printDrawOrderRecursive(sb *strings.Builder, s *Sprite, order int) int {
	visibility := "visible"
	if !s.visible {
		visibility = "hidden"
	}

	fmt.Fprintf(sb, "  %d. Sprite %d (%s)\n", order, s.id, visibility)
	order++

	for _, child := range s.children {
		order = sm.printDrawOrderRecursive(sb, child, order)
	}

	return order
}

// BringRootToFront はルートスプライトを最前面に移動する
func (sm *SpriteManager) BringRootToFront(spriteID int) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// ルートスプライトを見つける
	idx := -1
	for i, root := range sm.roots {
		if root.id == spriteID {
			idx = i
			break
		}
	}

	if idx == -1 {
		return fmt.Errorf("root sprite not found: %d", spriteID)
	}

	// 現在の位置から削除
	s := sm.roots[idx]
	sm.roots = append(sm.roots[:idx], sm.roots[idx+1:]...)

	// 末尾に追加
	sm.roots = append(sm.roots, s)

	return nil
}

// SendRootToBack はルートスプライトを最背面に移動する
func (sm *SpriteManager) SendRootToBack(spriteID int) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// ルートスプライトを見つける
	idx := -1
	for i, root := range sm.roots {
		if root.id == spriteID {
			idx = i
			break
		}
	}

	if idx == -1 {
		return fmt.Errorf("root sprite not found: %d", spriteID)
	}

	// 現在の位置から削除
	s := sm.roots[idx]
	sm.roots = append(sm.roots[:idx], sm.roots[idx+1:]...)

	// 先頭に挿入
	sm.roots = append([]*Sprite{s}, sm.roots...)

	return nil
}
