package graphics

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
// 要件 7.1: 指定サイズ・背景色のウインドウスプライトを作成できる
func TestCreateWindowSprite(t *testing.T) {
	sm := NewSpriteManager()
	wsm := NewWindowSpriteManager(sm)

	// テスト用のウインドウとピクチャーを作成
	win := &Window{
		ID:      1,
		PicID:   1,
		X:       100,
		Y:       50,
		Width:   200,
		Height:  150,
		BgColor: color.RGBA{255, 255, 255, 255},
		Visible: true,
		ZOrder:  0,
	}

	pic := &Picture{
		ID:     1,
		Width:  200,
		Height: 150,
		Image:  ebiten.NewImage(200, 150),
	}

	ws := wsm.CreateWindowSprite(win, pic)

	if ws == nil {
		t.Fatal("CreateWindowSprite returned nil")
	}
	if ws.GetWindow() != win {
		t.Error("Window not set correctly")
	}
	if ws.GetSprite() == nil {
		t.Error("Sprite not created")
	}

	// スプライトの位置を確認
	x, y := ws.GetSprite().Position()
	if x != 100 || y != 50 {
		t.Errorf("Expected position (100, 50), got (%v, %v)", x, y)
	}

	// スプライトのZ順序を確認
	if ws.GetSprite().ZOrder() != 0 {
		t.Errorf("Expected ZOrder 0, got %d", ws.GetSprite().ZOrder())
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

	win := &Window{
		ID:      1,
		Width:   200,
		Height:  150,
		Visible: true,
	}

	pic := &Picture{
		ID:     1,
		Width:  200,
		Height: 150,
		Image:  ebiten.NewImage(200, 150),
	}

	wsm.CreateWindowSprite(win, pic)

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
// 要件 7.3: ウインドウが閉じられたときにウインドウとその子スプライトを削除する
func TestRemoveWindowSprite(t *testing.T) {
	sm := NewSpriteManager()
	wsm := NewWindowSpriteManager(sm)

	win := &Window{
		ID:      1,
		Width:   200,
		Height:  150,
		Visible: true,
	}

	pic := &Picture{
		ID:     1,
		Width:  200,
		Height: 150,
		Image:  ebiten.NewImage(200, 150),
	}

	ws := wsm.CreateWindowSprite(win, pic)
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
		win := &Window{
			ID:      i,
			Width:   200,
			Height:  150,
			Visible: true,
		}
		pic := &Picture{
			ID:     i,
			Width:  200,
			Height: 150,
			Image:  ebiten.NewImage(200, 150),
		}
		wsm.CreateWindowSprite(win, pic)
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
// 要件 7.2: ウインドウスプライトを親として子スプライトを追加できる
func TestWindowSpriteAddChild(t *testing.T) {
	sm := NewSpriteManager()
	wsm := NewWindowSpriteManager(sm)

	win := &Window{
		ID:      1,
		X:       100,
		Y:       50,
		Width:   200,
		Height:  150,
		Visible: true,
	}

	pic := &Picture{
		ID:     1,
		Width:  200,
		Height: 150,
		Image:  ebiten.NewImage(200, 150),
	}

	ws := wsm.CreateWindowSprite(win, pic)

	// 子スプライトを作成
	childImg := ebiten.NewImage(50, 50)
	child := sm.CreateSprite(childImg)
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

	win := &Window{
		ID:      1,
		Width:   200,
		Height:  150,
		Visible: true,
	}

	pic := &Picture{
		ID:     1,
		Width:  200,
		Height: 150,
		Image:  ebiten.NewImage(200, 150),
	}

	ws := wsm.CreateWindowSprite(win, pic)

	// 子スプライトを作成して追加
	childImg := ebiten.NewImage(50, 50)
	child := sm.CreateSprite(childImg)
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
// 要件 7.3: ウインドウが閉じられたときにウインドウとその子スプライトを削除する
func TestWindowSpriteRemoveWithChildren(t *testing.T) {
	sm := NewSpriteManager()
	wsm := NewWindowSpriteManager(sm)

	win := &Window{
		ID:      1,
		Width:   200,
		Height:  150,
		Visible: true,
	}

	pic := &Picture{
		ID:     1,
		Width:  200,
		Height: 150,
		Image:  ebiten.NewImage(200, 150),
	}

	ws := wsm.CreateWindowSprite(win, pic)

	// 子スプライトを作成して追加
	childImg := ebiten.NewImage(50, 50)
	child := sm.CreateSprite(childImg)
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

	win := &Window{
		ID:      1,
		X:       100,
		Y:       50,
		Width:   200,
		Height:  150,
		Visible: true,
	}

	pic := &Picture{
		ID:     1,
		Width:  200,
		Height: 150,
		Image:  ebiten.NewImage(200, 150),
	}

	ws := wsm.CreateWindowSprite(win, pic)

	// 位置を更新
	ws.UpdatePosition(200, 100)

	// ウインドウの位置が更新されたことを確認
	if win.X != 200 || win.Y != 100 {
		t.Errorf("Expected window position (200, 100), got (%d, %d)", win.X, win.Y)
	}

	// スプライトの位置が更新されたことを確認
	x, y := ws.GetSprite().Position()
	if x != 200 || y != 100 {
		t.Errorf("Expected sprite position (200, 100), got (%v, %v)", x, y)
	}
}

// TestWindowSpriteUpdateZOrder はZ順序の更新をテストする
func TestWindowSpriteUpdateZOrder(t *testing.T) {
	sm := NewSpriteManager()
	wsm := NewWindowSpriteManager(sm)

	win := &Window{
		ID:      1,
		Width:   200,
		Height:  150,
		Visible: true,
		ZOrder:  0,
	}

	pic := &Picture{
		ID:     1,
		Width:  200,
		Height: 150,
		Image:  ebiten.NewImage(200, 150),
	}

	ws := wsm.CreateWindowSprite(win, pic)

	// Z順序を更新
	ws.UpdateZOrder(10)

	// ウインドウのZ順序が更新されたことを確認
	if win.ZOrder != 10 {
		t.Errorf("Expected window ZOrder 10, got %d", win.ZOrder)
	}

	// スプライトのZ順序が更新されたことを確認
	// 要件 14.3: グローバルZ順序が使用される
	expectedGlobalZOrder := CalculateGlobalZOrder(10, ZOrderWindowBase)
	if ws.GetSprite().ZOrder() != expectedGlobalZOrder {
		t.Errorf("Expected sprite ZOrder %d (global), got %d", expectedGlobalZOrder, ws.GetSprite().ZOrder())
	}
}

// TestWindowSpriteUpdateVisible は可視性の更新をテストする
func TestWindowSpriteUpdateVisible(t *testing.T) {
	sm := NewSpriteManager()
	wsm := NewWindowSpriteManager(sm)

	win := &Window{
		ID:      1,
		Width:   200,
		Height:  150,
		Visible: true,
	}

	pic := &Picture{
		ID:     1,
		Width:  200,
		Height: 150,
		Image:  ebiten.NewImage(200, 150),
	}

	ws := wsm.CreateWindowSprite(win, pic)

	// 可視性を更新
	ws.UpdateVisible(false)

	// ウインドウの可視性が更新されたことを確認
	if win.Visible {
		t.Error("Expected window to be invisible")
	}

	// スプライトの可視性が更新されたことを確認
	if ws.GetSprite().Visible() {
		t.Error("Expected sprite to be invisible")
	}
}

// TestWindowSpriteGetContentOffset はコンテンツオフセットの取得をテストする
func TestWindowSpriteGetContentOffset(t *testing.T) {
	sm := NewSpriteManager()
	wsm := NewWindowSpriteManager(sm)

	win := &Window{
		ID:      1,
		Width:   200,
		Height:  150,
		Visible: true,
	}

	pic := &Picture{
		ID:     1,
		Width:  200,
		Height: 150,
		Image:  ebiten.NewImage(200, 150),
	}

	ws := wsm.CreateWindowSprite(win, pic)

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

// TestGraphicsSystemWindowSpriteIntegration はGraphicsSystemとの統合をテストする
func TestGraphicsSystemWindowSpriteIntegration(t *testing.T) {
	gs := NewGraphicsSystem("")

	// ピクチャーを作成
	picID, err := gs.CreatePic(200, 150)
	if err != nil {
		t.Fatalf("CreatePic failed: %v", err)
	}

	// ウインドウを開く
	winID, err := gs.OpenWin(picID, 100, 50, 200, 150, 0, 0, 0)
	if err != nil {
		t.Fatalf("OpenWin failed: %v", err)
	}

	// WindowSpriteが作成されたことを確認
	wsm := gs.GetWindowSpriteManager()
	if wsm == nil {
		t.Fatal("WindowSpriteManager is nil")
	}

	ws := wsm.GetWindowSprite(winID)
	if ws == nil {
		t.Error("WindowSprite should be created when window is opened")
	}

	// ウインドウを閉じる
	err = gs.CloseWin(winID)
	if err != nil {
		t.Fatalf("CloseWin failed: %v", err)
	}

	// WindowSpriteが削除されたことを確認
	ws = wsm.GetWindowSprite(winID)
	if ws != nil {
		t.Error("WindowSprite should be removed when window is closed")
	}
}

// TestGraphicsSystemCloseWinAllRemovesWindowSprites はCloseWinAllがすべてのWindowSpriteを削除することをテストする
func TestGraphicsSystemCloseWinAllRemovesWindowSprites(t *testing.T) {
	gs := NewGraphicsSystem("")

	// 複数のピクチャーとウインドウを作成
	winIDs := make([]int, 3)
	for i := 0; i < 3; i++ {
		picID, _ := gs.CreatePic(200, 150)
		winID, _ := gs.OpenWin(picID, i*100, i*50, 200, 150, 0, 0, 0)
		winIDs[i] = winID
	}

	// WindowSpriteが作成されたことを確認
	wsm := gs.GetWindowSpriteManager()
	for _, winID := range winIDs {
		if wsm.GetWindowSprite(winID) == nil {
			t.Errorf("WindowSprite for window %d should exist", winID)
		}
	}

	// すべてのウインドウを閉じる
	gs.CloseWinAll()

	// すべてのWindowSpriteが削除されたことを確認
	for _, winID := range winIDs {
		if wsm.GetWindowSprite(winID) != nil {
			t.Errorf("WindowSprite for window %d should be removed after CloseWinAll", winID)
		}
	}
}

// TestGetWindowSpriteSprite はGetWindowSpriteSpriteメソッドをテストする
// 要件 14.2: ウインドウ内のスプライトをウインドウの子スプライトとして管理する
func TestGetWindowSpriteSprite(t *testing.T) {
	sm := NewSpriteManager()
	wsm := NewWindowSpriteManager(sm)

	win := &Window{
		ID:      1,
		Width:   200,
		Height:  150,
		Visible: true,
	}

	pic := &Picture{
		ID:     1,
		Width:  200,
		Height: 150,
		Image:  ebiten.NewImage(200, 150),
	}

	ws := wsm.CreateWindowSprite(win, pic)

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

	win := &Window{
		ID:      1,
		Width:   200,
		Height:  150,
		PicX:    10,
		PicY:    20,
		Visible: true,
	}

	pic := &Picture{
		ID:     1,
		Width:  200,
		Height: 150,
		Image:  ebiten.NewImage(200, 150),
	}

	ws := wsm.CreateWindowSprite(win, pic)

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

	win := &Window{
		ID:      1,
		Width:   200,
		Height:  150,
		Visible: true,
	}

	pic := &Picture{
		ID:     1,
		Width:  200,
		Height: 150,
		Image:  ebiten.NewImage(200, 150),
	}

	ws := wsm.CreateWindowSprite(win, pic)

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

	win := &Window{
		ID:      1,
		X:       100,
		Y:       50,
		Width:   200,
		Height:  150,
		Visible: true,
	}

	pic := &Picture{
		ID:     1,
		Width:  200,
		Height: 150,
		Image:  ebiten.NewImage(200, 150),
	}

	ws := wsm.CreateWindowSprite(win, pic)

	// 親スプライトの透明度を設定
	ws.GetSprite().SetAlpha(0.5)

	// 子スプライトを作成して追加
	childImg := ebiten.NewImage(50, 50)
	child := sm.CreateSprite(childImg)
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
