package sprite

import (
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// TestNewWindowSpriteManager はWindowSpriteManagerの作成をテストする
func TestNewWindowSpriteManager(t *testing.T) {
	sm := NewSpriteManager()
	wsm := NewWindowSpriteManager(sm)

	if wsm == nil {
		t.Fatal("NewWindowSpriteManager returned nil")
	}
	if wsm.spriteManager != sm {
		t.Error("SpriteManager not set correctly")
	}
	if len(wsm.windowSprites) != 0 {
		t.Errorf("Expected empty windowSprites map, got %d", len(wsm.windowSprites))
	}
}

// TestCreateWindowSprite はWindowSpriteの作成をテストする
// 要件 4.1: 指定サイズ・背景色のウインドウスプライトを作成できる
func TestCreateWindowSprite(t *testing.T) {
	sm := NewSpriteManager()
	wsm := NewWindowSpriteManager(sm)

	config := WindowSpriteConfig{
		ID:      1,
		X:       100,
		Y:       50,
		Width:   200,
		Height:  150,
		PicX:    0,
		PicY:    0,
		BgColor: color.RGBA{255, 255, 255, 255},
		Visible: true,
	}

	ws := wsm.CreateWindowSprite(config, 200, 150)

	if ws == nil {
		t.Fatal("CreateWindowSprite returned nil")
	}
	if ws.GetWindowID() != 1 {
		t.Errorf("Expected windowID 1, got %d", ws.GetWindowID())
	}
	if ws.GetSprite() == nil {
		t.Error("Sprite not created")
	}

	// スプライトの位置を確認
	x, y := ws.GetSprite().Position()
	if x != 100 || y != 50 {
		t.Errorf("Expected position (100, 50), got (%v, %v)", x, y)
	}

	// スプライトの可視性を確認
	if !ws.GetSprite().Visible() {
		t.Error("Expected sprite to be visible")
	}
}

// TestGetWindowSprite はWindowSpriteの取得をテストする
func TestGetWindowSprite(t *testing.T) {
	sm := NewSpriteManager()
	wsm := NewWindowSpriteManager(sm)

	config := WindowSpriteConfig{
		ID:      1,
		Width:   200,
		Height:  150,
		Visible: true,
	}

	wsm.CreateWindowSprite(config, 200, 150)

	// 存在するWindowSpriteを取得
	ws := wsm.GetWindowSprite(1)
	if ws == nil {
		t.Error("GetWindowSprite returned nil for existing window")
	}

	// 存在しないWindowSpriteを取得
	ws = wsm.GetWindowSprite(999)
	if ws != nil {
		t.Error("GetWindowSprite should return nil for non-existing window")
	}
}

// TestRemoveWindowSprite はWindowSpriteの削除をテストする
// 要件 4.3: ウインドウが閉じられたときにウインドウとその子スプライトを削除する
func TestRemoveWindowSprite(t *testing.T) {
	sm := NewSpriteManager()
	wsm := NewWindowSpriteManager(sm)

	config := WindowSpriteConfig{
		ID:      1,
		Width:   200,
		Height:  150,
		Visible: true,
	}

	ws := wsm.CreateWindowSprite(config, 200, 150)
	spriteID := ws.GetSprite().ID()

	// スプライトが存在することを確認
	if sm.GetSprite(spriteID) == nil {
		t.Error("Sprite should exist before removal")
	}

	// WindowSpriteを削除
	wsm.RemoveWindowSprite(1)

	// WindowSpriteが削除されたことを確認
	if wsm.GetWindowSprite(1) != nil {
		t.Error("WindowSprite should be removed")
	}

	// スプライトも削除されたことを確認
	if sm.GetSprite(spriteID) != nil {
		t.Error("Sprite should be removed from SpriteManager")
	}
}

// TestWindowSpriteClear はすべてのWindowSpriteの削除をテストする
func TestWindowSpriteClear(t *testing.T) {
	sm := NewSpriteManager()
	wsm := NewWindowSpriteManager(sm)

	// 複数のWindowSpriteを作成
	for i := 1; i <= 3; i++ {
		config := WindowSpriteConfig{
			ID:      i,
			Width:   200,
			Height:  150,
			Visible: true,
		}
		wsm.CreateWindowSprite(config, 200, 150)
	}

	// すべてのWindowSpriteをクリア
	wsm.Clear()

	// すべてのWindowSpriteが削除されたことを確認
	for i := 1; i <= 3; i++ {
		if wsm.GetWindowSprite(i) != nil {
			t.Errorf("WindowSprite %d should be removed after Clear", i)
		}
	}
}

// TestWindowSpriteAddChild は子スプライトの追加をテストする
// 要件 4.2: ウインドウスプライトを親として子スプライトを追加できる
func TestWindowSpriteAddChild(t *testing.T) {
	sm := NewSpriteManager()
	wsm := NewWindowSpriteManager(sm)

	config := WindowSpriteConfig{
		ID:      1,
		X:       100,
		Y:       50,
		Width:   200,
		Height:  150,
		Visible: true,
	}

	ws := wsm.CreateWindowSprite(config, 200, 150)

	// 子スプライトを作成
	childImg := ebiten.NewImage(50, 50)
	child := sm.CreateSprite(childImg, nil)
	child.SetPosition(10, 20)

	// 子スプライトを追加
	ws.AddChild(child)

	// 子スプライトの親が設定されていることを確認
	if child.Parent() != ws.GetSprite() {
		t.Error("Child's parent should be the window sprite")
	}

	// 子スプライトの絶対位置を確認
	// 親の位置(100, 50) + 子の相対位置(10, 20) = (110, 70)
	absX, absY := child.AbsolutePosition()
	if absX != 110 || absY != 70 {
		t.Errorf("Expected absolute position (110, 70), got (%v, %v)", absX, absY)
	}

	// 子スプライトのリストを確認
	children := ws.GetChildren()
	if len(children) != 1 {
		t.Errorf("Expected 1 child, got %d", len(children))
	}
}

// TestWindowSpriteRemoveChild は子スプライトの削除をテストする
func TestWindowSpriteRemoveChild(t *testing.T) {
	sm := NewSpriteManager()
	wsm := NewWindowSpriteManager(sm)

	config := WindowSpriteConfig{
		ID:      1,
		Width:   200,
		Height:  150,
		Visible: true,
	}

	ws := wsm.CreateWindowSprite(config, 200, 150)

	// 子スプライトを作成して追加
	childImg := ebiten.NewImage(50, 50)
	child := sm.CreateSprite(childImg, nil)
	ws.AddChild(child)

	// 子スプライトを削除
	ws.RemoveChild(child.ID())

	// 子スプライトの親がnilになっていることを確認
	if child.Parent() != nil {
		t.Error("Child's parent should be nil after removal")
	}

	// 子スプライトのリストが空になっていることを確認
	children := ws.GetChildren()
	if len(children) != 0 {
		t.Errorf("Expected 0 children, got %d", len(children))
	}
}

// TestWindowSpriteRemoveWithChildren は子スプライトを持つWindowSpriteの削除をテストする
// 要件 4.3: ウインドウが閉じられたときにウインドウとその子スプライトを削除する
func TestWindowSpriteRemoveWithChildren(t *testing.T) {
	sm := NewSpriteManager()
	wsm := NewWindowSpriteManager(sm)

	config := WindowSpriteConfig{
		ID:      1,
		Width:   200,
		Height:  150,
		Visible: true,
	}

	ws := wsm.CreateWindowSprite(config, 200, 150)

	// 子スプライトを作成して追加
	childImg := ebiten.NewImage(50, 50)
	child := sm.CreateSprite(childImg, nil)
	childID := child.ID()
	ws.AddChild(child)

	// WindowSpriteを削除
	wsm.RemoveWindowSprite(1)

	// 子スプライトも削除されたことを確認
	if sm.GetSprite(childID) != nil {
		t.Error("Child sprite should be removed when WindowSprite is removed")
	}
}

// TestWindowSpriteUpdatePosition は位置の更新をテストする
func TestWindowSpriteUpdatePosition(t *testing.T) {
	sm := NewSpriteManager()
	wsm := NewWindowSpriteManager(sm)

	config := WindowSpriteConfig{
		ID:      1,
		X:       100,
		Y:       50,
		Width:   200,
		Height:  150,
		Visible: true,
	}

	ws := wsm.CreateWindowSprite(config, 200, 150)

	// 位置を更新
	ws.UpdatePosition(200, 100)

	// スプライトの位置が更新されたことを確認
	x, y := ws.GetSprite().Position()
	if x != 200 || y != 100 {
		t.Errorf("Expected sprite position (200, 100), got (%v, %v)", x, y)
	}
}

// TestWindowSpriteUpdateVisible は可視性の更新をテストする
func TestWindowSpriteUpdateVisible(t *testing.T) {
	sm := NewSpriteManager()
	wsm := NewWindowSpriteManager(sm)

	config := WindowSpriteConfig{
		ID:      1,
		Width:   200,
		Height:  150,
		Visible: true,
	}

	ws := wsm.CreateWindowSprite(config, 200, 150)

	// 可視性を更新
	ws.UpdateVisible(false)

	// スプライトの可視性が更新されたことを確認
	if ws.GetSprite().Visible() {
		t.Error("Expected sprite to be invisible")
	}
}

// TestWindowSpriteGetContentOffset はコンテンツオフセットの取得をテストする
func TestWindowSpriteGetContentOffset(t *testing.T) {
	sm := NewSpriteManager()
	wsm := NewWindowSpriteManager(sm)

	config := WindowSpriteConfig{
		ID:      1,
		Width:   200,
		Height:  150,
		Visible: true,
	}

	ws := wsm.CreateWindowSprite(config, 200, 150)

	// コンテンツオフセットを取得
	offsetX, offsetY := ws.GetContentOffset()

	// borderThickness = 4, titleBarHeight = 20
	if offsetX != 4 {
		t.Errorf("Expected offsetX 4, got %d", offsetX)
	}
	if offsetY != 24 { // 4 + 20
		t.Errorf("Expected offsetY 24, got %d", offsetY)
	}
}

// TestGetWindowSpriteSprite はGetWindowSpriteSpriteメソッドをテストする
func TestGetWindowSpriteSprite(t *testing.T) {
	sm := NewSpriteManager()
	wsm := NewWindowSpriteManager(sm)

	config := WindowSpriteConfig{
		ID:      1,
		Width:   200,
		Height:  150,
		Visible: true,
	}

	ws := wsm.CreateWindowSprite(config, 200, 150)

	// GetWindowSpriteSpriteで基盤スプライトを取得
	sprite := wsm.GetWindowSpriteSprite(1)
	if sprite == nil {
		t.Error("GetWindowSpriteSprite returned nil for existing window")
	}
	if sprite != ws.GetSprite() {
		t.Error("GetWindowSpriteSprite should return the same sprite as GetSprite()")
	}

	// 存在しないウインドウの場合
	sprite = wsm.GetWindowSpriteSprite(999)
	if sprite != nil {
		t.Error("GetWindowSpriteSprite should return nil for non-existing window")
	}
}

// TestGetPicOffset はGetPicOffsetメソッドをテストする
func TestGetPicOffset(t *testing.T) {
	sm := NewSpriteManager()
	wsm := NewWindowSpriteManager(sm)

	config := WindowSpriteConfig{
		ID:      1,
		Width:   200,
		Height:  150,
		PicX:    10,
		PicY:    20,
		Visible: true,
	}

	ws := wsm.CreateWindowSprite(config, 200, 150)

	// GetPicOffsetでピクチャーオフセットを取得
	picX, picY := ws.GetPicOffset()
	if picX != 10 || picY != 20 {
		t.Errorf("Expected PicOffset (10, 20), got (%d, %d)", picX, picY)
	}
}

// TestGetContentSprite はGetContentSpriteメソッドをテストする
func TestGetContentSprite(t *testing.T) {
	sm := NewSpriteManager()
	wsm := NewWindowSpriteManager(sm)

	config := WindowSpriteConfig{
		ID:      1,
		Width:   200,
		Height:  150,
		Visible: true,
	}

	ws := wsm.CreateWindowSprite(config, 200, 150)

	// GetContentSpriteでコンテンツスプライトを取得
	contentSprite := ws.GetContentSprite()
	if contentSprite == nil {
		t.Error("GetContentSprite returned nil")
	}
	if contentSprite != ws.GetSprite() {
		t.Error("GetContentSprite should return the same sprite as GetSprite()")
	}
}

// TestChildSpriteInheritance は子スプライトが親の属性を継承することをテストする
// 要件 2.1, 2.2, 2.3: 親子関係の位置、透明度、可視性の継承
func TestChildSpriteInheritance(t *testing.T) {
	sm := NewSpriteManager()
	wsm := NewWindowSpriteManager(sm)

	config := WindowSpriteConfig{
		ID:      1,
		X:       100,
		Y:       50,
		Width:   200,
		Height:  150,
		Visible: true,
	}

	ws := wsm.CreateWindowSprite(config, 200, 150)

	// 親スプライトの透明度を設定
	ws.GetSprite().SetAlpha(0.5)

	// 子スプライトを作成して追加
	childImg := ebiten.NewImage(50, 50)
	child := sm.CreateSprite(childImg, nil)
	child.SetPosition(10, 20)
	child.SetAlpha(0.8)
	ws.AddChild(child)

	// 子スプライトの実効透明度を確認（0.5 * 0.8 = 0.4）
	effectiveAlpha := child.EffectiveAlpha()
	if effectiveAlpha != 0.4 {
		t.Errorf("Expected effective alpha 0.4, got %v", effectiveAlpha)
	}

	// 親を非表示にする
	ws.UpdateVisible(false)

	// 子スプライトの実効可視性を確認
	if child.IsEffectivelyVisible() {
		t.Error("Child should be effectively invisible when parent is invisible")
	}
}

// TestWindowSpriteRegisteredAsRoot はウインドウがルートスプライトとして登録されることをテストする
// 要件 11.1: ウインドウをルートスプライトとして扱う
func TestWindowSpriteRegisteredAsRoot(t *testing.T) {
	sm := NewSpriteManager()
	wsm := NewWindowSpriteManager(sm)

	config := WindowSpriteConfig{
		ID:      1,
		Width:   200,
		Height:  150,
		Visible: true,
	}

	ws := wsm.CreateWindowSprite(config, 200, 150)

	// ルートスプライトとして登録されていることを確認
	roots := sm.GetRoots()
	if len(roots) != 1 {
		t.Errorf("Expected 1 root sprite, got %d", len(roots))
	}

	// ルートスプライトがウインドウのスプライトであることを確認
	if roots[0] != ws.GetSprite() {
		t.Error("Root sprite should be the window sprite")
	}

	// ウインドウスプライトの親がnilであることを確認（ルートスプライトの条件）
	if ws.GetSprite().Parent() != nil {
		t.Error("Window sprite should have no parent (root sprite)")
	}
}

// TestMultipleWindowsAsRoots は複数のウインドウがルートスプライトとして登録されることをテストする
// 要件 11.3: ルートスプライトのスライス順序でウインドウの描画順序を決定する
func TestMultipleWindowsAsRoots(t *testing.T) {
	sm := NewSpriteManager()
	wsm := NewWindowSpriteManager(sm)

	// 複数のウインドウを作成
	var windowSprites []*WindowSprite
	for i := 1; i <= 3; i++ {
		config := WindowSpriteConfig{
			ID:      i,
			Width:   200,
			Height:  150,
			Visible: true,
		}
		ws := wsm.CreateWindowSprite(config, 200, 150)
		windowSprites = append(windowSprites, ws)
	}

	// ルートスプライトの数を確認
	roots := sm.GetRoots()
	if len(roots) != 3 {
		t.Errorf("Expected 3 root sprites, got %d", len(roots))
	}

	// ルートスプライトの順序を確認（作成順）
	for i, ws := range windowSprites {
		if roots[i] != ws.GetSprite() {
			t.Errorf("Root sprite %d should be window sprite %d", i, ws.GetWindowID())
		}
	}
}

// TestWindowSpriteRemoveWithNestedChildren はネストした子スプライトを持つWindowSpriteの削除をテストする
// 要件 4.3: ウインドウが閉じられたときにウインドウとその子スプライトを削除する
func TestWindowSpriteRemoveWithNestedChildren(t *testing.T) {
	sm := NewSpriteManager()
	wsm := NewWindowSpriteManager(sm)

	config := WindowSpriteConfig{
		ID:      1,
		Width:   200,
		Height:  150,
		Visible: true,
	}

	ws := wsm.CreateWindowSprite(config, 200, 150)

	// 子スプライトを作成して追加
	childImg := ebiten.NewImage(50, 50)
	child := sm.CreateSprite(childImg, nil)
	childID := child.ID()
	ws.AddChild(child)

	// 孫スプライトを作成して追加
	grandchildImg := ebiten.NewImage(25, 25)
	grandchild := sm.CreateSprite(grandchildImg, child)
	grandchildID := grandchild.ID()

	// スプライトが存在することを確認
	if sm.GetSprite(childID) == nil {
		t.Error("Child sprite should exist before removal")
	}
	if sm.GetSprite(grandchildID) == nil {
		t.Error("Grandchild sprite should exist before removal")
	}

	// WindowSpriteを削除
	wsm.RemoveWindowSprite(1)

	// 子スプライトも削除されたことを確認
	if sm.GetSprite(childID) != nil {
		t.Error("Child sprite should be removed when WindowSprite is removed")
	}

	// 孫スプライトも削除されたことを確認
	if sm.GetSprite(grandchildID) != nil {
		t.Error("Grandchild sprite should be removed when WindowSprite is removed")
	}
}
