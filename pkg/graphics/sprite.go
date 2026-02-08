// Package graphics provides sprite-based rendering system.
package graphics

import (
	"encoding/json"
	"fmt"
	"image"
	"sort"
	"strings"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

// Sprite は汎用スプライト（階層的Z順序対応）
// すべての描画要素（ウインドウ、ピクチャ、キャスト、文字、図形）の基盤となる
type Sprite struct {
	id      int
	image   *ebiten.Image
	x, y    float64
	visible bool
	alpha   float64
	parent  *Sprite
	dirty   bool // 再描画が必要かどうか

	// 階層的Z順序
	// 要件 1.1: スプライトはZ_Pathを持つ
	zPath *ZPath

	// 子スプライト管理
	// 要件 9.1: PictureSpriteは子スプライトを持てる
	children []*Sprite

	// ソートキャッシュ
	// 要件 5.4: 比較結果をキャッシュして再利用する
	sortKey string // Z_Pathの文字列表現（キャッシュ用）

	// 子スプライトの座標オフセット
	// 子スプライトの位置を計算する際に適用されるオフセット
	// 例: ウィンドウのコンテンツ領域オフセット、PicX/PicYオフセット
	childOffsetX, childOffsetY float64

	// カスタム描画関数
	// 透明色処理など、特殊な描画が必要な場合に使用
	// nilの場合は通常の描画を行う
	customDraw func(screen *ebiten.Image, x, y float64, alpha float32)
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
		zPath:    nil,
		children: nil,
		sortKey:  "",
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

// SetParent はスプライトの親を設定する
func (s *Sprite) SetParent(p *Sprite) {
	s.parent = p
	s.dirty = true
}

// GetZPath はスプライトのZ_Pathを返す
// 要件 1.1: スプライトはZ_Pathを持つ
func (s *Sprite) GetZPath() *ZPath {
	return s.zPath
}

// SetZPath はスプライトのZ_Pathを設定する
// 要件 8.2: Local_Z_Orderが変更されたとき、Z_Pathを再計算する
func (s *Sprite) SetZPath(zPath *ZPath) {
	s.zPath = zPath
	if zPath != nil {
		s.sortKey = zPath.String()
	} else {
		s.sortKey = ""
	}
	s.dirty = true
}

// SortKey はソートキャッシュを返す
// 要件 5.4: 比較結果をキャッシュして再利用する
func (s *Sprite) SortKey() string {
	return s.sortKey
}

// GetChildren は子スプライトのリストを返す
// 要件 9.1: PictureSpriteは子スプライトを持てる
func (s *Sprite) GetChildren() []*Sprite {
	return s.children
}

// AddChild は子スプライトを追加する
// 要件 1.4: 子スプライトが作成されたとき、親のZ_Pathを継承し、自身のLocal_Z_Orderを追加する
// 注意: 既に別の親がある場合は、その親から削除してから追加する
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

// AbsolutePosition は親を考慮した絶対位置を返す
// 親のchildOffsetX/childOffsetYも考慮する
func (s *Sprite) AbsolutePosition() (float64, float64) {
	x, y := s.x, s.y
	if s.parent != nil {
		px, py := s.parent.AbsolutePosition()
		// 親のchildOffsetを適用
		x += px + s.parent.childOffsetX
		y += py + s.parent.childOffsetY
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

// GetChildOffset は子スプライトの座標オフセットを返す
func (s *Sprite) GetChildOffset() (float64, float64) {
	return s.childOffsetX, s.childOffsetY
}

// SetChildOffset は子スプライトの座標オフセットを設定する
// このオフセットは子スプライトの絶対位置計算時に適用される
func (s *Sprite) SetChildOffset(x, y float64) {
	s.childOffsetX = x
	s.childOffsetY = y
	s.dirty = true
}

// SetCustomDraw はカスタム描画関数を設定する
// 透明色処理など、特殊な描画が必要な場合に使用
// nilを設定すると通常の描画に戻る
func (s *Sprite) SetCustomDraw(fn func(screen *ebiten.Image, x, y float64, alpha float32)) {
	s.customDraw = fn
	s.dirty = true
}

// GetCustomDraw はカスタム描画関数を返す
func (s *Sprite) GetCustomDraw() func(screen *ebiten.Image, x, y float64, alpha float32) {
	return s.customDraw
}

// SpriteManager はスプライトを管理する（階層的Z順序対応）
type SpriteManager struct {
	mu       sync.RWMutex
	sprites  map[int]*Sprite
	nextID   int
	sorted   []*Sprite // Z順序でソート済みのキャッシュ
	needSort bool

	// 階層的Z順序
	// 要件 2.1: 各親スプライトごとにZ_Order_Counterを管理する
	zOrderCounter *ZOrderCounter

	// デバッグ描画コールバック
	// 各スプライト描画後に呼び出される（デバッグオーバーレイ用）
	// 引数: screen, sprite, absX, absY
	debugDrawCallback func(screen *ebiten.Image, s *Sprite, absX, absY float64)
}

// NewSpriteManager は新しいSpriteManagerを作成する
func NewSpriteManager() *SpriteManager {
	return &SpriteManager{
		sprites:       make(map[int]*Sprite),
		nextID:        1,
		needSort:      true,
		zOrderCounter: NewZOrderCounter(),
	}
}

// SetDebugDrawCallback はデバッグ描画コールバックを設定する
// 各スプライト描画後に呼び出され、デバッグ情報を描画するために使用する
// nilを設定するとデバッグ描画を無効化する
func (sm *SpriteManager) SetDebugDrawCallback(callback func(screen *ebiten.Image, s *Sprite, absX, absY float64)) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.debugDrawCallback = callback
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

// CreateSpriteHidden は新しいスプライトを非表示状態で作成して登録する
// レースコンディション対策: スプライトが完全に初期化される前に描画されることを防ぐ
// Z_Pathを設定した後にSetVisible(true)を呼ぶ必要がある
func (sm *SpriteManager) CreateSpriteHidden(img *ebiten.Image) *Sprite {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	s := NewSprite(sm.nextID, img)
	s.visible = false // 最初から非表示で作成
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
// 要件 9.2: ウィンドウが閉じられたとき、関連するスプライト（キャスト、テキスト等）を削除する
//
// このメソッドは以下の処理を行います：
// 1. スプライトを親の子リストから削除
// 2. すべての子スプライトを再帰的に削除
// 3. スプライト自身をsprites mapから削除
func (sm *SpriteManager) RemoveSprite(id int) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sprite := sm.sprites[id]
	if sprite == nil {
		return
	}

	// 1. 親の子リストから削除
	if sprite.parent != nil {
		sprite.parent.RemoveChild(id)
	}

	// 2. すべての子スプライトを再帰的に削除
	// 子リストのコピーを作成してから削除（イテレーション中の変更を避ける）
	children := make([]*Sprite, len(sprite.children))
	copy(children, sprite.children)
	for _, child := range children {
		sm.removeSpriteLocked(child.id)
	}

	// 3. スプライト自身を削除
	delete(sm.sprites, id)
	sm.needSort = true
}

// removeSpriteLocked はロック済みの状態でスプライトを削除する（内部用）
// RemoveSpriteから再帰的に呼び出される
func (sm *SpriteManager) removeSpriteLocked(id int) {
	sprite := sm.sprites[id]
	if sprite == nil {
		return
	}

	// 親の子リストから削除（親がいる場合）
	if sprite.parent != nil {
		sprite.parent.RemoveChild(id)
	}

	// すべての子スプライトを再帰的に削除
	children := make([]*Sprite, len(sprite.children))
	copy(children, sprite.children)
	for _, child := range children {
		sm.removeSpriteLocked(child.id)
	}

	// スプライト自身を削除
	delete(sm.sprites, id)
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

// sortSprites はスプライトをZ_Pathの辞書順でソートする
// 要件 1.5: Z_Pathの辞書順比較でスプライトの描画順序を決定する
// 要件 7.1: Z_Pathのソート結果をキャッシュする
//
// Z_Pathがnilのスプライトは、Z_Pathを持つスプライトより先に描画されます（背面）。
func (sm *SpriteManager) sortSprites() {
	sm.sorted = make([]*Sprite, 0, len(sm.sprites))
	for _, s := range sm.sprites {
		sm.sorted = append(sm.sorted, s)
	}

	sort.Slice(sm.sorted, func(i, j int) bool {
		si := sm.sorted[i]
		sj := sm.sorted[j]

		// 両方ともZ_Pathを持つ場合は辞書順比較
		if si.zPath != nil && sj.zPath != nil {
			return si.zPath.Less(sj.zPath)
		}

		// 片方だけZ_Pathを持つ場合
		// Z_Pathを持たないスプライトを先に描画（背面）
		if si.zPath == nil && sj.zPath != nil {
			return true
		}
		if si.zPath != nil && sj.zPath == nil {
			return false
		}

		// 両方ともZ_Pathを持たない場合はIDで比較（安定ソート）
		return si.id < sj.id
	})

	sm.needSort = false
}

// Draw はすべての可視スプライトをZ_Path順で描画する
// 要件 3.1: 親スプライトを先に描画し、その後に子スプライトを描画する
// 要件 3.2: 同じ親を持つ子スプライトをLocal_Z_Order順で描画する
// 要件 15.1-15.8: デバッグオーバーレイの描画（各スプライト描画直後）
func (sm *SpriteManager) Draw(screen *ebiten.Image) {
	sm.mu.Lock()
	if sm.needSort {
		sm.sortSprites()
	}
	// ソート済みスライスのコピーを作成し、描画中のレースコンディションを防ぐ
	// 各スプライトの状態（visible, image, position等）も描画前にスナップショットを取る
	type drawItem struct {
		sprite     *Sprite
		visible    bool
		image      *ebiten.Image
		x, y       float64
		alpha      float64
		customDraw func(screen *ebiten.Image, x, y float64, alpha float32)
	}
	items := make([]drawItem, 0, len(sm.sorted))
	for _, s := range sm.sorted {
		// レースコンディション対策: zPathがnilのスプライトはスキップ
		// スプライトが完全に初期化される前（zPathが設定される前）に描画されることを防ぐ
		// これにより、新しく作成されたスプライトがzPathを設定する前に
		// 意図しない位置（背面）に描画されることを防ぐ
		if s.zPath == nil {
			continue
		}
		if !s.IsEffectivelyVisible() || s.image == nil {
			continue
		}
		x, y := s.AbsolutePosition()
		items = append(items, drawItem{
			sprite:     s,
			visible:    true,
			image:      s.image,
			x:          x,
			y:          y,
			alpha:      s.EffectiveAlpha(),
			customDraw: s.customDraw,
		})
	}
	debugCallback := sm.debugDrawCallback
	sm.mu.Unlock()

	for _, item := range items {
		// カスタム描画関数が設定されている場合はそれを使用
		// 透明色処理など、特殊な描画が必要なスプライトで使用
		if item.customDraw != nil {
			item.customDraw(screen, item.x, item.y, float32(item.alpha))
		} else {
			// 通常描画
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(item.x, item.y)

			if item.alpha < 1.0 {
				op.ColorScale.ScaleAlpha(float32(item.alpha))
			}

			screen.DrawImage(item.image, op)
		}

		// デバッグ描画コールバックを呼び出す（各スプライト描画直後）
		// これにより、後から描画されるスプライトによってデバッグ情報が隠れる
		if debugCallback != nil {
			debugCallback(screen, item.sprite, item.x, item.y)
		}
	}
}

// MarkNeedSort はソートが必要であることをマークする
func (sm *SpriteManager) MarkNeedSort() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.needSort = true
}

// GetZOrderCounter はZOrderCounterを返す
// 要件 2.1: 各親スプライトごとにZ_Order_Counterを管理する
// 外部からZOrderCounterにアクセスするために使用します（CastSpriteManager等）
func (sm *SpriteManager) GetZOrderCounter() *ZOrderCounter {
	return sm.zOrderCounter
}

// CreateSpriteWithZPath は新しいスプライトを作成してZ_Pathを設定する
// 要件 2.2, 2.3, 2.4: 操作時にZ_Order_Counterを使用してLocal_Z_Orderを割り当てる
// 要件 1.4: 子スプライトが作成されたとき、親のZ_Pathを継承し、自身のLocal_Z_Orderを追加する
//
// parentがnilの場合、ルートスプライトとして作成されます（Local_Z_Order=0から開始）。
// parentが指定された場合、親の子スプライトとして作成され、親のZ_Pathを継承します。
//
// 例:
//
//	// ルートスプライトの作成
//	root := sm.CreateSpriteWithZPath(img, nil) // Z_Path: [0]
//
//	// 子スプライトの作成
//	child1 := sm.CreateSpriteWithZPath(img, root) // Z_Path: [0, 0]
//	child2 := sm.CreateSpriteWithZPath(img, root) // Z_Path: [0, 1]
func (sm *SpriteManager) CreateSpriteWithZPath(img *ebiten.Image, parent *Sprite) *Sprite {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	s := NewSprite(sm.nextID, img)
	sm.sprites[s.id] = s
	sm.nextID++

	// 親のZ_Pathを継承してZ_Pathを設定
	var parentID int
	var parentZPath *ZPath
	if parent != nil {
		parentID = parent.id
		parentZPath = parent.zPath
		parent.AddChild(s)
	}

	// 要件 2.5: Z_Order_Counterをインクリメント
	localZOrder := sm.zOrderCounter.GetNext(parentID)
	s.SetZPath(NewZPathFromParent(parentZPath, localZOrder))

	// 要件 7.2: スプライトの変更時にソートが必要であることをマークする
	sm.needSort = true

	return s
}

// CreateRootSprite はルートスプライト（ウインドウ）を作成する
// 要件 1.3: Root_Spriteは単一要素のZ_Path（例: [0]）を持つ
// 要件 4.1: ウインドウをRoot_Spriteとして扱う
//
// windowZOrderはウインドウのZ順序を指定します。
// 前面のウインドウほど大きな値を持ちます。
//
// 例:
//
//	// ウインドウ0（背面）
//	window0 := sm.CreateRootSprite(img, 0) // Z_Path: [0]
//
//	// ウインドウ1（前面）
//	window1 := sm.CreateRootSprite(img, 1) // Z_Path: [1]
func (sm *SpriteManager) CreateRootSprite(img *ebiten.Image, windowZOrder int) *Sprite {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	s := NewSprite(sm.nextID, img)
	sm.sprites[s.id] = s
	sm.nextID++

	// ルートスプライトは単一要素のZ_Path
	s.SetZPath(NewZPath(windowZOrder))

	// 要件 7.2: スプライトの変更時にソートが必要であることをマークする
	sm.needSort = true

	return s
}

// BringToFront はスプライトを最前面に移動する
// 要件 8.4: スプライトを最前面に移動するメソッドを提供する
//
// 同じ親を持つ兄弟スプライトの中で、指定されたスプライトを最前面に移動します。
// スプライトのLocal_Z_Orderが更新され、Z_Pathが再計算されます。
// 子スプライトがある場合、それらのZ_Pathも再帰的に更新されます。
//
// 例:
//
//	// スプライトを最前面に移動
//	err := sm.BringToFront(spriteID)
//	if err != nil {
//	    // エラー処理
//	}
func (sm *SpriteManager) BringToFront(spriteID int) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	s := sm.sprites[spriteID]
	if s == nil {
		return fmt.Errorf("sprite not found: %d", spriteID)
	}

	// 親のIDを取得
	var parentID int
	if s.parent != nil {
		parentID = s.parent.id
	}

	// 新しいZ順序を取得
	newLocalZOrder := sm.zOrderCounter.GetNext(parentID)

	// Z_Pathを再計算
	// 要件 8.2: Local_Z_Orderが変更されたとき、Z_Pathを再計算する
	var parentZPath *ZPath
	if s.parent != nil {
		parentZPath = s.parent.zPath
	}
	s.SetZPath(NewZPathFromParent(parentZPath, newLocalZOrder))

	// 要件 8.3: 親スプライトが変更されたとき、子スプライトのZ_Pathを再計算する
	sm.updateChildrenZPaths(s)

	sm.needSort = true
	return nil
}

// updateChildrenZPaths は子スプライトのZ_Pathを再帰的に更新する
// 要件 8.3: 親スプライトが変更されたとき、子スプライトのZ_Pathを再計算する
//
// 親スプライトのZ_Pathが変更された場合、すべての子スプライトのZ_Pathを
// 親のZ_Pathを基に再計算します。この処理は再帰的に行われ、
// すべての子孫スプライトのZ_Pathが更新されます。
func (sm *SpriteManager) updateChildrenZPaths(parent *Sprite) {
	for _, child := range parent.children {
		localZOrder := child.zPath.LocalZOrder()
		child.SetZPath(NewZPathFromParent(parent.zPath, localZOrder))
		sm.updateChildrenZPaths(child)
	}
}

// UpdateChildrenZPathsForTest はテスト用にupdateChildrenZPathsを公開するメソッド
// 本番コードでは使用しないでください
func (sm *SpriteManager) UpdateChildrenZPathsForTest(parent *Sprite) {
	sm.updateChildrenZPaths(parent)
}

// UpdateChildrenZPaths は子スプライトのZ_Pathを再帰的に更新する（公開メソッド）
// 要件 4.3: ウインドウのZ順序変更時に、そのウインドウの子スプライトのZ_Pathを更新する
// 要件 8.3: 親スプライトが変更されたとき、子スプライトのZ_Pathを再計算する
//
// WindowSpriteなど外部からZ_Path更新が必要な場合に使用します。
func (sm *SpriteManager) UpdateChildrenZPaths(parent *Sprite) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.updateChildrenZPaths(parent)
}

// SendToBack はスプライトを最背面に移動する
// 要件 8.5: スプライトを最背面に移動するメソッドを提供する
//
// 同じ親を持つ兄弟スプライトの中で、指定されたスプライトを最背面に移動します。
// 兄弟スプライトの中で最小のLocal_Z_Orderを見つけ、それより1小さい値を
// 新しいLocal_Z_Orderとして設定します。
// 子スプライトがある場合、それらのZ_Pathも再帰的に更新されます。
//
// 例:
//
//	// スプライトを最背面に移動
//	err := sm.SendToBack(spriteID)
//	if err != nil {
//	    // エラー処理
//	}
func (sm *SpriteManager) SendToBack(spriteID int) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	s := sm.sprites[spriteID]
	if s == nil {
		return fmt.Errorf("sprite not found: %d", spriteID)
	}

	// 最小のZ順序を見つける
	minZOrder := 0
	foundSibling := false
	for _, other := range sm.sprites {
		if other.parent == s.parent && other.id != s.id {
			if other.zPath != nil {
				if !foundSibling || other.zPath.LocalZOrder() < minZOrder {
					minZOrder = other.zPath.LocalZOrder()
					foundSibling = true
				}
			}
		}
	}

	// 新しいZ順序を設定（最小値 - 1）
	newLocalZOrder := minZOrder - 1

	// Z_Pathを再計算
	// 要件 8.2: Local_Z_Orderが変更されたとき、Z_Pathを再計算する
	var parentZPath *ZPath
	if s.parent != nil {
		parentZPath = s.parent.zPath
	}
	s.SetZPath(NewZPathFromParent(parentZPath, newLocalZOrder))

	// 要件 8.3: 親スプライトが変更されたとき、子スプライトのZ_Pathを再計算する
	sm.updateChildrenZPaths(s)

	sm.needSort = true
	return nil
}

// ============================================================================
// Z_Pathの可視化 (Z_Path Visualization)
// ============================================================================

// ZPathString はスプライトのZ_Pathを文字列として取得する
// 要件 10.1: スプライトのZ_Pathを文字列として取得できる
//
// Z_Pathがnilの場合は"nil"を返します。
//
// 例:
//
//	str := sprite.ZPathString() // "[0, 1, 2]"
func (s *Sprite) ZPathString() string {
	if s.zPath == nil {
		return "nil"
	}
	return s.zPath.String()
}

// PrintHierarchy はスプライト階層をツリー形式で出力する
// 要件 10.2: スプライト階層をツリー形式で出力できる
//
// ルートスプライトから始まり、子スプライトをインデントして表示します。
// 各スプライトはID、Z_Path、可視性を表示します。
//
// 例:
//
//	hierarchy := sm.PrintHierarchy()
//	fmt.Println(hierarchy)
//	// 出力:
//	// - Sprite 1: [0] (visible)
//	//   - Sprite 2: [0, 0] (visible)
//	//   - Sprite 3: [0, 1] (hidden)
//	// - Sprite 4: [1] (visible)
func (sm *SpriteManager) PrintHierarchy() string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var sb strings.Builder

	// ルートスプライトを見つける
	roots := make([]*Sprite, 0)
	for _, s := range sm.sprites {
		if s.parent == nil {
			roots = append(roots, s)
		}
	}

	// Z_Path順でソート
	sort.Slice(roots, func(i, j int) bool {
		if roots[i].zPath == nil {
			return true
		}
		if roots[j].zPath == nil {
			return false
		}
		return roots[i].zPath.Less(roots[j].zPath)
	})

	// ツリー形式で出力
	for _, root := range roots {
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

	zPathStr := "nil"
	if s.zPath != nil {
		zPathStr = s.zPath.String()
	}

	sb.WriteString(fmt.Sprintf("%s- Sprite %d: %s (%s)\n", indent, s.id, zPathStr, visibility))

	// 子スプライトをZ_Path順でソート
	children := make([]*Sprite, len(s.children))
	copy(children, s.children)
	sort.Slice(children, func(i, j int) bool {
		if children[i].zPath == nil {
			return true
		}
		if children[j].zPath == nil {
			return false
		}
		return children[i].zPath.Less(children[j].zPath)
	})

	for _, child := range children {
		sm.printSpriteTree(sb, child, depth+1)
	}
}

// PrintDrawOrder は描画順序のリストを出力する
// 要件 10.3: 描画順序のリストを出力できる
//
// スプライトを描画順序（Z_Path順）で一覧表示します。
// 各スプライトはID、Z_Path、可視性を表示します。
//
// 例:
//
//	drawOrder := sm.PrintDrawOrder()
//	fmt.Println(drawOrder)
//	// 出力:
//	// Draw Order:
//	//   1. Sprite 1: [0] (visible)
//	//   2. Sprite 2: [0, 0] (visible)
//	//   3. Sprite 3: [0, 1] (hidden)
//	//   4. Sprite 4: [1] (visible)
func (sm *SpriteManager) PrintDrawOrder() string {
	sm.mu.Lock()
	if sm.needSort {
		sm.sortSprites()
	}
	sorted := sm.sorted
	sm.mu.Unlock()

	var sb strings.Builder
	sb.WriteString("Draw Order:\n")
	for i, s := range sorted {
		visibility := "visible"
		if !s.visible {
			visibility = "hidden"
		}

		zPathStr := "nil"
		if s.zPath != nil {
			zPathStr = s.zPath.String()
		}

		sb.WriteString(fmt.Sprintf("  %d. Sprite %d: %s (%s)\n", i+1, s.id, zPathStr, visibility))
	}
	return sb.String()
}

// SpriteStateJSON はスプライト状態のJSON表現
type SpriteStateJSON struct {
	TotalSprites int           `json:"total_sprites"`
	Sprites      []*SpriteJSON `json:"sprites"`
}

// SpriteJSON は個々のスプライトのJSON表現
type SpriteJSON struct {
	ID                 int           `json:"id"`
	ZPath              interface{}   `json:"z_path"` // []int or nil
	Position           [2]float64    `json:"position"`
	Size               [2]int        `json:"size"`
	Visible            bool          `json:"visible"`
	EffectivelyVisible bool          `json:"effectively_visible"`
	Alpha              float64       `json:"alpha"`
	ParentID           *int          `json:"parent_id,omitempty"`
	Children           []*SpriteJSON `json:"children,omitempty"`
}

// DumpSpriteState はスプライトの状態を詳細にダンプする（デバッグ用）
// 操作後のスプライト構成を確認するために使用
// JSON形式で出力する
func (sm *SpriteManager) DumpSpriteState() string {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	state := &SpriteStateJSON{
		TotalSprites: len(sm.sprites),
		Sprites:      make([]*SpriteJSON, 0),
	}

	// ルートスプライトを見つける
	roots := make([]*Sprite, 0)
	for _, s := range sm.sprites {
		if s.parent == nil {
			roots = append(roots, s)
		}
	}

	// Z_Path順でソート
	sort.Slice(roots, func(i, j int) bool {
		if roots[i].zPath == nil {
			return true
		}
		if roots[j].zPath == nil {
			return false
		}
		return roots[i].zPath.Less(roots[j].zPath)
	})

	// 各ルートスプライトとその子を出力
	for _, root := range roots {
		state.Sprites = append(state.Sprites, sm.spriteToJSON(root))
	}

	// JSON形式で出力（改行なし）
	jsonBytes, err := json.Marshal(state)
	if err != nil {
		return fmt.Sprintf(`{"error": "%s"}`, err.Error())
	}
	return string(jsonBytes)
}

// spriteToJSON はスプライトをJSON構造体に変換する
// 注意: 子スプライトはスライスの順序（描画順序）で出力する
func (sm *SpriteManager) spriteToJSON(s *Sprite) *SpriteJSON {
	sj := &SpriteJSON{
		ID:                 s.id,
		Position:           [2]float64{s.x, s.y},
		Visible:            s.visible,
		EffectivelyVisible: s.IsEffectivelyVisible(),
		Alpha:              s.alpha,
	}

	// Z_Path
	if s.zPath != nil {
		sj.ZPath = s.zPath.path
	} else {
		sj.ZPath = nil
	}

	// 画像サイズ
	if s.image != nil {
		bounds := s.image.Bounds()
		sj.Size = [2]int{bounds.Dx(), bounds.Dy()}
	} else {
		sj.Size = [2]int{0, 0}
	}

	// 親ID
	if s.parent != nil {
		parentID := s.parent.id
		sj.ParentID = &parentID
	}

	// 子スプライトをスライスの順序（描画順序）で出力
	// 注意: 以前はZ_Path順でソートしていたが、実際の描画順序を確認するためにスライス順序で出力する
	if len(s.children) > 0 {
		sj.Children = make([]*SpriteJSON, 0, len(s.children))
		for _, child := range s.children {
			sj.Children = append(sj.Children, sm.spriteToJSON(child))
		}
	}

	return sj
}

// DumpSpriteStateText はスプライトの状態をテキスト形式でダンプする（従来形式）
// 操作後のスプライト構成を確認するために使用
func (sm *SpriteManager) DumpSpriteStateText() string {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	var sb strings.Builder
	sb.WriteString("=== Sprite State Dump ===\n")
	sb.WriteString(fmt.Sprintf("Total sprites: %d\n", len(sm.sprites)))

	// ルートスプライトを見つける
	roots := make([]*Sprite, 0)
	for _, s := range sm.sprites {
		if s.parent == nil {
			roots = append(roots, s)
		}
	}

	// Z_Path順でソート
	sort.Slice(roots, func(i, j int) bool {
		if roots[i].zPath == nil {
			return true
		}
		if roots[j].zPath == nil {
			return false
		}
		return roots[i].zPath.Less(roots[j].zPath)
	})

	// 各ルートスプライトとその子を出力
	for _, root := range roots {
		sm.dumpSpriteRecursive(&sb, root, 0)
	}

	sb.WriteString("=========================\n")
	return sb.String()
}

// dumpSpriteRecursive はスプライトを再帰的にダンプする
func (sm *SpriteManager) dumpSpriteRecursive(sb *strings.Builder, s *Sprite, depth int) {
	indent := strings.Repeat("  ", depth)

	visibility := "V"
	if !s.visible {
		visibility = "H"
	}
	effectiveVisibility := "EV"
	if !s.IsEffectivelyVisible() {
		effectiveVisibility = "EH"
	}

	zPathStr := "nil"
	if s.zPath != nil {
		zPathStr = s.zPath.String()
	}

	// 画像サイズ
	imgSize := "nil"
	if s.image != nil {
		bounds := s.image.Bounds()
		imgSize = fmt.Sprintf("%dx%d", bounds.Dx(), bounds.Dy())
	}

	sb.WriteString(fmt.Sprintf("%sSprite[%d] zPath=%s pos=(%.0f,%.0f) size=%s %s/%s children=%d\n",
		indent, s.id, zPathStr, s.x, s.y, imgSize, visibility, effectiveVisibility, len(s.children)))

	// 子スプライトをZ_Path順でソート
	children := make([]*Sprite, len(s.children))
	copy(children, s.children)
	sort.Slice(children, func(i, j int) bool {
		if children[i].zPath == nil {
			return true
		}
		if children[j].zPath == nil {
			return false
		}
		return children[i].zPath.Less(children[j].zPath)
	})

	for _, child := range children {
		sm.dumpSpriteRecursive(sb, child, depth+1)
	}
}
