package sprite

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// TestNewDebugOverlay はNewDebugOverlayをテストする
func TestNewDebugOverlay(t *testing.T) {
	do := NewDebugOverlay()

	if do == nil {
		t.Fatal("NewDebugOverlay returned nil")
	}

	// デフォルトでは無効
	if do.IsEnabled() {
		t.Error("DebugOverlay should be disabled by default")
	}

	// デフォルトオプションの確認
	options := do.GetOptions()
	if !options.ShowSpriteInfo {
		t.Error("ShowSpriteInfo should be true by default")
	}
	if !options.ShowBoundingBoxes {
		t.Error("ShowBoundingBoxes should be true by default")
	}
	if options.ShowHiddenSprites {
		t.Error("ShowHiddenSprites should be false by default")
	}
	if !options.ShowChildCount {
		t.Error("ShowChildCount should be true by default")
	}
}

// TestDebugOverlaySetEnabled はSetEnabled/IsEnabledをテストする
func TestDebugOverlaySetEnabled(t *testing.T) {
	do := NewDebugOverlay()

	// 有効化
	do.SetEnabled(true)
	if !do.IsEnabled() {
		t.Error("DebugOverlay should be enabled after SetEnabled(true)")
	}

	// 無効化
	do.SetEnabled(false)
	if do.IsEnabled() {
		t.Error("DebugOverlay should be disabled after SetEnabled(false)")
	}
}

// TestDebugOverlaySetOptions はSetOptions/GetOptionsをテストする
func TestDebugOverlaySetOptions(t *testing.T) {
	do := NewDebugOverlay()

	// カスタムオプションを設定
	customOptions := DebugOverlayOptions{
		ShowSpriteInfo:    false,
		ShowBoundingBoxes: true,
		ShowHiddenSprites: true,
		ShowChildCount:    false,
	}
	do.SetOptions(customOptions)

	// オプションを取得して確認
	options := do.GetOptions()
	if options.ShowSpriteInfo != customOptions.ShowSpriteInfo {
		t.Errorf("ShowSpriteInfo mismatch: got %v, want %v", options.ShowSpriteInfo, customOptions.ShowSpriteInfo)
	}
	if options.ShowBoundingBoxes != customOptions.ShowBoundingBoxes {
		t.Errorf("ShowBoundingBoxes mismatch: got %v, want %v", options.ShowBoundingBoxes, customOptions.ShowBoundingBoxes)
	}
	if options.ShowHiddenSprites != customOptions.ShowHiddenSprites {
		t.Errorf("ShowHiddenSprites mismatch: got %v, want %v", options.ShowHiddenSprites, customOptions.ShowHiddenSprites)
	}
	if options.ShowChildCount != customOptions.ShowChildCount {
		t.Errorf("ShowChildCount mismatch: got %v, want %v", options.ShowChildCount, customOptions.ShowChildCount)
	}
}

// TestDefaultDebugOverlayOptions はDefaultDebugOverlayOptionsをテストする
func TestDefaultDebugOverlayOptions(t *testing.T) {
	options := DefaultDebugOverlayOptions()

	if !options.ShowSpriteInfo {
		t.Error("ShowSpriteInfo should be true by default")
	}
	if !options.ShowBoundingBoxes {
		t.Error("ShowBoundingBoxes should be true by default")
	}
	if options.ShowHiddenSprites {
		t.Error("ShowHiddenSprites should be false by default")
	}
	if !options.ShowChildCount {
		t.Error("ShowChildCount should be true by default")
	}
}

// TestGlobalDebugOverlay はグローバルデバッグオーバーレイ関数をテストする
func TestGlobalDebugOverlay(t *testing.T) {
	// 初期状態を保存
	initialState := IsDebugMode()
	defer SetDebugMode(initialState)

	// グローバルオーバーレイの取得
	do := GetDebugOverlay()
	if do == nil {
		t.Fatal("GetDebugOverlay returned nil")
	}

	// デバッグモードの設定
	SetDebugMode(true)
	if !IsDebugMode() {
		t.Error("Debug mode should be enabled after SetDebugMode(true)")
	}

	SetDebugMode(false)
	if IsDebugMode() {
		t.Error("Debug mode should be disabled after SetDebugMode(false)")
	}
}

// TestDebugOverlayDrawDisabled はデバッグオーバーレイが無効時に描画しないことをテストする
func TestDebugOverlayDrawDisabled(t *testing.T) {
	do := NewDebugOverlay()
	sm := NewSpriteManager()

	// スプライトを作成
	img := ebiten.NewImage(100, 100)
	sm.CreateSprite(img, nil)

	// 無効時は描画しない（パニックしないことを確認）
	screen := ebiten.NewImage(800, 600)
	do.Draw(screen, sm) // パニックしなければOK
}

// TestDebugOverlayDrawEnabled はデバッグオーバーレイが有効時に描画することをテストする
func TestDebugOverlayDrawEnabled(t *testing.T) {
	do := NewDebugOverlay()
	do.SetEnabled(true)
	sm := NewSpriteManager()

	// スプライトを作成
	img := ebiten.NewImage(100, 100)
	root := sm.CreateSprite(img, nil)

	// 子スプライトを追加
	childImg := ebiten.NewImage(50, 50)
	child := sm.CreateSprite(childImg, root)
	child.SetPosition(10, 10)

	// 有効時は描画する（パニックしないことを確認）
	screen := ebiten.NewImage(800, 600)
	do.Draw(screen, sm) // パニックしなければOK
}

// TestDebugOverlayDrawWithHiddenSprites は非表示スプライトの描画をテストする
func TestDebugOverlayDrawWithHiddenSprites(t *testing.T) {
	do := NewDebugOverlay()
	do.SetEnabled(true)
	sm := NewSpriteManager()

	// スプライトを作成
	img := ebiten.NewImage(100, 100)
	root := sm.CreateSprite(img, nil)

	// 非表示の子スプライトを追加
	childImg := ebiten.NewImage(50, 50)
	child := sm.CreateSprite(childImg, root)
	child.SetVisible(false)

	// ShowHiddenSprites = false の場合
	screen := ebiten.NewImage(800, 600)
	do.Draw(screen, sm) // パニックしなければOK

	// ShowHiddenSprites = true の場合
	options := do.GetOptions()
	options.ShowHiddenSprites = true
	do.SetOptions(options)
	do.Draw(screen, sm) // パニックしなければOK
}

// TestDebugOverlayDrawSingleSprite はDrawSingleSpriteをテストする
func TestDebugOverlayDrawSingleSprite(t *testing.T) {
	do := NewDebugOverlay()
	do.SetEnabled(true)

	// スプライトを作成
	img := ebiten.NewImage(100, 100)
	s := NewSprite(1, img)
	s.SetPosition(50, 50)

	// 単一スプライトの描画（パニックしないことを確認）
	screen := ebiten.NewImage(800, 600)
	do.DrawSingleSprite(screen, s)

	// nilスプライトの場合
	do.DrawSingleSprite(screen, nil) // パニックしなければOK
}

// TestSpriteManagerDrawDebugOverlay はSpriteManager.DrawDebugOverlayをテストする
func TestSpriteManagerDrawDebugOverlay(t *testing.T) {
	// 初期状態を保存
	initialState := IsDebugMode()
	defer SetDebugMode(initialState)

	sm := NewSpriteManager()

	// スプライトを作成
	img := ebiten.NewImage(100, 100)
	sm.CreateSprite(img, nil)

	// デバッグモードを有効化
	SetDebugMode(true)

	// 描画（パニックしないことを確認）
	screen := ebiten.NewImage(800, 600)
	sm.DrawDebugOverlay(screen)
}

// TestDebugOverlayOptionsGlobal はグローバルオプション関数をテストする
func TestDebugOverlayOptionsGlobal(t *testing.T) {
	// 初期状態を保存
	initialOptions := GetDebugOverlayOptions()
	defer SetDebugOverlayOptions(initialOptions)

	// カスタムオプションを設定
	customOptions := DebugOverlayOptions{
		ShowSpriteInfo:    false,
		ShowBoundingBoxes: false,
		ShowHiddenSprites: true,
		ShowChildCount:    false,
	}
	SetDebugOverlayOptions(customOptions)

	// オプションを取得して確認
	options := GetDebugOverlayOptions()
	if options.ShowSpriteInfo != customOptions.ShowSpriteInfo {
		t.Errorf("ShowSpriteInfo mismatch: got %v, want %v", options.ShowSpriteInfo, customOptions.ShowSpriteInfo)
	}
	if options.ShowBoundingBoxes != customOptions.ShowBoundingBoxes {
		t.Errorf("ShowBoundingBoxes mismatch: got %v, want %v", options.ShowBoundingBoxes, customOptions.ShowBoundingBoxes)
	}
	if options.ShowHiddenSprites != customOptions.ShowHiddenSprites {
		t.Errorf("ShowHiddenSprites mismatch: got %v, want %v", options.ShowHiddenSprites, customOptions.ShowHiddenSprites)
	}
	if options.ShowChildCount != customOptions.ShowChildCount {
		t.Errorf("ShowChildCount mismatch: got %v, want %v", options.ShowChildCount, customOptions.ShowChildCount)
	}
}

// TestDebugOverlayDrawWithNilImage は画像がnilのスプライトの描画をテストする
func TestDebugOverlayDrawWithNilImage(t *testing.T) {
	do := NewDebugOverlay()
	do.SetEnabled(true)
	sm := NewSpriteManager()

	// 画像なしのスプライトを作成
	root := sm.CreateSprite(nil, nil)
	root.SetPosition(100, 100)

	// 描画（パニックしないことを確認）
	screen := ebiten.NewImage(800, 600)
	do.Draw(screen, sm) // パニックしなければOK
}

// TestDebugOverlayDrawDeepHierarchy は深い階層のスプライトの描画をテストする
func TestDebugOverlayDrawDeepHierarchy(t *testing.T) {
	do := NewDebugOverlay()
	do.SetEnabled(true)
	sm := NewSpriteManager()

	// 深い階層を作成
	img := ebiten.NewImage(100, 100)
	root := sm.CreateSprite(img, nil)

	parent := root
	for i := 0; i < 10; i++ {
		childImg := ebiten.NewImage(50, 50)
		child := sm.CreateSprite(childImg, parent)
		child.SetPosition(float64(i*10), float64(i*10))
		parent = child
	}

	// 描画（パニックしないことを確認）
	screen := ebiten.NewImage(800, 600)
	do.Draw(screen, sm) // パニックしなければOK
}
