package graphics

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestNewCastSpriteManager(t *testing.T) {
	sm := NewSpriteManager()
	csm := NewCastSpriteManager(sm)

	if csm == nil {
		t.Fatal("NewCastSpriteManager returned nil")
	}
	if csm.Count() != 0 {
		t.Errorf("expected 0 cast sprites, got %d", csm.Count())
	}
}

func TestCreateCastSprite(t *testing.T) {
	sm := NewSpriteManager()
	csm := NewCastSpriteManager(sm)

	// テスト用のキャストを作成
	cast := &Cast{
		ID:      0,
		WinID:   0,
		PicID:   1,
		X:       10,
		Y:       20,
		SrcX:    0,
		SrcY:    0,
		Width:   32,
		Height:  32,
		Visible: true,
		ZOrder:  0,
	}

	// テスト用のソース画像を作成
	srcImage := ebiten.NewImage(64, 64)

	// CastSpriteを作成
	cs := csm.CreateCastSprite(cast, srcImage, 100)

	if cs == nil {
		t.Fatal("CreateCastSprite returned nil")
	}
	if csm.Count() != 1 {
		t.Errorf("expected 1 cast sprite, got %d", csm.Count())
	}

	// スプライトの属性を確認
	sprite := cs.GetSprite()
	if sprite == nil {
		t.Fatal("GetSprite returned nil")
	}

	x, y := sprite.Position()
	if x != 10 || y != 20 {
		t.Errorf("expected position (10, 20), got (%f, %f)", x, y)
	}

	// Z_Pathはまだ設定されていない（親スプライトなしで作成されたため）
	// zOrderパラメータは互換性のために残されているが、実際のZ順序はZ_Pathで管理される

	// レースコンディション対策: 親スプライトなしで作成された場合、スプライトは非表示のまま
	// 親スプライトを設定してZ_Pathを設定した後にSetVisible(true)が呼ばれる
	if sprite.Visible() {
		t.Error("expected sprite to be hidden (no parent set)")
	}
}

func TestCreateCastSpriteWithTransColor(t *testing.T) {
	// このテストはEbitengineのゲームループ外では実行できないためスキップ
	// 透明色処理はsrc.At()を使用し、これはReadPixelsを内部で呼び出す
	t.Skip("Transparent color processing cannot be tested before the game starts")
}

func TestGetCastSprite(t *testing.T) {
	sm := NewSpriteManager()
	csm := NewCastSpriteManager(sm)

	// テスト用のキャストを作成
	cast := &Cast{
		ID:      5,
		WinID:   0,
		PicID:   1,
		X:       10,
		Y:       20,
		SrcX:    0,
		SrcY:    0,
		Width:   32,
		Height:  32,
		Visible: true,
		ZOrder:  0,
	}

	srcImage := ebiten.NewImage(64, 64)
	csm.CreateCastSprite(cast, srcImage, 100)

	// 存在するキャストを取得
	cs := csm.GetCastSprite(5)
	if cs == nil {
		t.Fatal("GetCastSprite returned nil for existing cast")
	}

	// 存在しないキャストを取得
	cs = csm.GetCastSprite(999)
	if cs != nil {
		t.Error("GetCastSprite should return nil for non-existing cast")
	}
}

func TestRemoveCastSprite(t *testing.T) {
	sm := NewSpriteManager()
	csm := NewCastSpriteManager(sm)

	// テスト用のキャストを作成
	cast := &Cast{
		ID:      0,
		WinID:   0,
		PicID:   1,
		X:       10,
		Y:       20,
		SrcX:    0,
		SrcY:    0,
		Width:   32,
		Height:  32,
		Visible: true,
		ZOrder:  0,
	}

	srcImage := ebiten.NewImage(64, 64)
	csm.CreateCastSprite(cast, srcImage, 100)

	if csm.Count() != 1 {
		t.Errorf("expected 1 cast sprite, got %d", csm.Count())
	}

	// CastSpriteを削除
	csm.RemoveCastSprite(0)

	if csm.Count() != 0 {
		t.Errorf("expected 0 cast sprites after removal, got %d", csm.Count())
	}

	// 削除後は取得できない
	cs := csm.GetCastSprite(0)
	if cs != nil {
		t.Error("GetCastSprite should return nil after removal")
	}
}

func TestGetCastSpritesByWindow(t *testing.T) {
	sm := NewSpriteManager()
	csm := NewCastSpriteManager(sm)

	srcImage := ebiten.NewImage(64, 64)

	// ウィンドウ0に2つのキャストを作成
	cast1 := &Cast{ID: 0, WinID: 0, PicID: 1, X: 10, Y: 10, Width: 32, Height: 32, Visible: true}
	cast2 := &Cast{ID: 1, WinID: 0, PicID: 1, X: 20, Y: 20, Width: 32, Height: 32, Visible: true}

	// ウィンドウ1に1つのキャストを作成
	cast3 := &Cast{ID: 2, WinID: 1, PicID: 1, X: 30, Y: 30, Width: 32, Height: 32, Visible: true}

	csm.CreateCastSprite(cast1, srcImage, 100)
	csm.CreateCastSprite(cast2, srcImage, 101)
	csm.CreateCastSprite(cast3, srcImage, 102)

	// ウィンドウ0のキャストを取得
	sprites := csm.GetCastSpritesByWindow(0)
	if len(sprites) != 2 {
		t.Errorf("expected 2 cast sprites for window 0, got %d", len(sprites))
	}

	// ウィンドウ1のキャストを取得
	sprites = csm.GetCastSpritesByWindow(1)
	if len(sprites) != 1 {
		t.Errorf("expected 1 cast sprite for window 1, got %d", len(sprites))
	}

	// 存在しないウィンドウのキャストを取得
	sprites = csm.GetCastSpritesByWindow(999)
	if len(sprites) != 0 {
		t.Errorf("expected 0 cast sprites for window 999, got %d", len(sprites))
	}
}

func TestRemoveCastSpritesByWindow(t *testing.T) {
	sm := NewSpriteManager()
	csm := NewCastSpriteManager(sm)

	srcImage := ebiten.NewImage(64, 64)

	// ウィンドウ0に2つのキャストを作成
	cast1 := &Cast{ID: 0, WinID: 0, PicID: 1, X: 10, Y: 10, Width: 32, Height: 32, Visible: true}
	cast2 := &Cast{ID: 1, WinID: 0, PicID: 1, X: 20, Y: 20, Width: 32, Height: 32, Visible: true}

	// ウィンドウ1に1つのキャストを作成
	cast3 := &Cast{ID: 2, WinID: 1, PicID: 1, X: 30, Y: 30, Width: 32, Height: 32, Visible: true}

	csm.CreateCastSprite(cast1, srcImage, 100)
	csm.CreateCastSprite(cast2, srcImage, 101)
	csm.CreateCastSprite(cast3, srcImage, 102)

	if csm.Count() != 3 {
		t.Errorf("expected 3 cast sprites, got %d", csm.Count())
	}

	// ウィンドウ0のキャストを削除
	csm.RemoveCastSpritesByWindow(0)

	if csm.Count() != 1 {
		t.Errorf("expected 1 cast sprite after removal, got %d", csm.Count())
	}

	// ウィンドウ0のキャストは取得できない
	sprites := csm.GetCastSpritesByWindow(0)
	if len(sprites) != 0 {
		t.Errorf("expected 0 cast sprites for window 0 after removal, got %d", len(sprites))
	}

	// ウィンドウ1のキャストは残っている
	sprites = csm.GetCastSpritesByWindow(1)
	if len(sprites) != 1 {
		t.Errorf("expected 1 cast sprite for window 1 after removal, got %d", len(sprites))
	}
}

func TestCastSpriteClear(t *testing.T) {
	sm := NewSpriteManager()
	csm := NewCastSpriteManager(sm)

	srcImage := ebiten.NewImage(64, 64)

	// 複数のキャストを作成
	for i := 0; i < 5; i++ {
		cast := &Cast{ID: i, WinID: 0, PicID: 1, X: i * 10, Y: i * 10, Width: 32, Height: 32, Visible: true}
		csm.CreateCastSprite(cast, srcImage, 100+i)
	}

	if csm.Count() != 5 {
		t.Errorf("expected 5 cast sprites, got %d", csm.Count())
	}

	// すべてクリア
	csm.Clear()

	if csm.Count() != 0 {
		t.Errorf("expected 0 cast sprites after clear, got %d", csm.Count())
	}
}

func TestCastSpriteUpdatePosition(t *testing.T) {
	sm := NewSpriteManager()
	csm := NewCastSpriteManager(sm)

	cast := &Cast{ID: 0, WinID: 0, PicID: 1, X: 10, Y: 20, Width: 32, Height: 32, Visible: true}
	srcImage := ebiten.NewImage(64, 64)

	cs := csm.CreateCastSprite(cast, srcImage, 100)

	// 位置を更新
	cs.UpdatePosition(100, 200)

	// キャストの位置が更新されていることを確認
	if cs.GetCast().X != 100 || cs.GetCast().Y != 200 {
		t.Errorf("expected cast position (100, 200), got (%d, %d)", cs.GetCast().X, cs.GetCast().Y)
	}

	// スプライトの位置が更新されていることを確認
	x, y := cs.GetSprite().Position()
	if x != 100 || y != 200 {
		t.Errorf("expected sprite position (100, 200), got (%f, %f)", x, y)
	}
}

func TestCastSpriteUpdateVisible(t *testing.T) {
	sm := NewSpriteManager()
	csm := NewCastSpriteManager(sm)

	cast := &Cast{ID: 0, WinID: 0, PicID: 1, X: 10, Y: 20, Width: 32, Height: 32, Visible: true}
	srcImage := ebiten.NewImage(64, 64)

	cs := csm.CreateCastSprite(cast, srcImage, 100)

	// 可視性を更新
	cs.UpdateVisible(false)

	// キャストの可視性が更新されていることを確認
	if cs.GetCast().Visible {
		t.Error("expected cast to be invisible")
	}

	// スプライトの可視性が更新されていることを確認
	if cs.GetSprite().Visible() {
		t.Error("expected sprite to be invisible")
	}
}

func TestCastSpriteSetParent(t *testing.T) {
	sm := NewSpriteManager()
	csm := NewCastSpriteManager(sm)

	cast := &Cast{ID: 0, WinID: 0, PicID: 1, X: 10, Y: 20, Width: 32, Height: 32, Visible: true}
	srcImage := ebiten.NewImage(64, 64)

	cs := csm.CreateCastSprite(cast, srcImage, 100)

	// 親スプライトを作成
	parentSprite := sm.CreateSpriteWithSize(100, 100)
	parentSprite.SetPosition(50, 50)

	// 親を設定
	cs.SetParent(parentSprite)

	// 親が設定されていることを確認
	if cs.GetSprite().Parent() != parentSprite {
		t.Error("expected parent to be set")
	}

	// 絶対位置が親の位置を考慮していることを確認
	absX, absY := cs.GetSprite().AbsolutePosition()
	if absX != 60 || absY != 70 { // 10+50, 20+50
		t.Errorf("expected absolute position (60, 70), got (%f, %f)", absX, absY)
	}
}

func TestCastSpriteNilCast(t *testing.T) {
	sm := NewSpriteManager()
	csm := NewCastSpriteManager(sm)

	srcImage := ebiten.NewImage(64, 64)

	// nilキャストでCastSpriteを作成
	cs := csm.CreateCastSprite(nil, srcImage, 100)

	if cs != nil {
		t.Error("CreateCastSprite should return nil for nil cast")
	}
}

func TestCastSpriteNilSourceImage(t *testing.T) {
	sm := NewSpriteManager()
	csm := NewCastSpriteManager(sm)

	cast := &Cast{ID: 0, WinID: 0, PicID: 1, X: 10, Y: 20, Width: 32, Height: 32, Visible: true}

	// nilソース画像でCastSpriteを作成
	cs := csm.CreateCastSprite(cast, nil, 100)

	if cs == nil {
		t.Fatal("CreateCastSprite should not return nil for nil source image")
	}

	// スプライトの画像はnilになる
	if cs.GetSprite().Image() != nil {
		t.Error("expected sprite image to be nil for nil source image")
	}
}

func TestApplyColorKeyToImage(t *testing.T) {
	// このテストはEbitengineのゲームループ外では実行できないためスキップ
	// ReadPixels/WritePixelsはゲームが開始した後でないと使用できない
	t.Skip("ReadPixels/WritePixels cannot be called before the game starts")
}

// TestCreateCastSpriteWithParent はCreateCastSpriteWithParentメソッドをテストする
// 要件 14.2: ウインドウ内のスプライトをウインドウの子スプライトとして管理する
func TestCreateCastSpriteWithParent(t *testing.T) {
	sm := NewSpriteManager()
	csm := NewCastSpriteManager(sm)

	// 親スプライトを作成
	parentSprite := sm.CreateSpriteWithSize(200, 150)
	parentSprite.SetPosition(100, 50)

	// テスト用のキャストを作成
	cast := &Cast{
		ID:      0,
		WinID:   0,
		PicID:   1,
		X:       10,
		Y:       20,
		SrcX:    0,
		SrcY:    0,
		Width:   32,
		Height:  32,
		Visible: true,
		ZOrder:  0,
	}

	srcImage := ebiten.NewImage(64, 64)

	// 親スプライト付きでCastSpriteを作成
	cs := csm.CreateCastSpriteWithParent(cast, srcImage, 100, parentSprite)

	if cs == nil {
		t.Fatal("CreateCastSpriteWithParent returned nil")
	}

	// 親が設定されていることを確認
	if cs.GetSprite().Parent() != parentSprite {
		t.Error("expected parent to be set")
	}

	// 絶対位置が親の位置を考慮していることを確認
	absX, absY := cs.GetSprite().AbsolutePosition()
	if absX != 110 || absY != 70 { // 10+100, 20+50
		t.Errorf("expected absolute position (110, 70), got (%f, %f)", absX, absY)
	}
}

// TestCreateCastSpriteWithTransColorAndParent はCreateCastSpriteWithTransColorAndParentメソッドをテストする
// 要件 8.4, 14.2: 透明色処理と親子関係の管理
func TestCreateCastSpriteWithTransColorAndParent(t *testing.T) {
	sm := NewSpriteManager()
	csm := NewCastSpriteManager(sm)

	// 親スプライトを作成
	parentSprite := sm.CreateSpriteWithSize(200, 150)
	parentSprite.SetPosition(100, 50)

	// テスト用のキャストを作成
	cast := &Cast{
		ID:      0,
		WinID:   0,
		PicID:   1,
		X:       10,
		Y:       20,
		SrcX:    0,
		SrcY:    0,
		Width:   32,
		Height:  32,
		Visible: true,
		ZOrder:  0,
	}

	srcImage := ebiten.NewImage(64, 64)

	// 透明色と親スプライト付きでCastSpriteを作成
	cs := csm.CreateCastSpriteWithTransColorAndParent(cast, srcImage, 100, nil, parentSprite)

	if cs == nil {
		t.Fatal("CreateCastSpriteWithTransColorAndParent returned nil")
	}

	// 親が設定されていることを確認
	if cs.GetSprite().Parent() != parentSprite {
		t.Error("expected parent to be set")
	}

	// 絶対位置が親の位置を考慮していることを確認
	absX, absY := cs.GetSprite().AbsolutePosition()
	if absX != 110 || absY != 70 { // 10+100, 20+50
		t.Errorf("expected absolute position (110, 70), got (%f, %f)", absX, absY)
	}
}

// TestCastSpriteParentVisibilityInheritance は親の可視性が子に継承されることをテストする
// 要件 2.3: 親が非表示のとき子も非表示として扱う
func TestCastSpriteParentVisibilityInheritance(t *testing.T) {
	sm := NewSpriteManager()
	csm := NewCastSpriteManager(sm)

	// 親スプライトを作成（Z_Pathを設定）
	parentSprite := sm.CreateSpriteWithSize(200, 150)
	parentSprite.SetZPath(NewZPath(0)) // Z_Pathを設定
	parentSprite.SetVisible(true)

	// テスト用のキャストを作成
	cast := &Cast{
		ID:      0,
		WinID:   0,
		PicID:   1,
		X:       10,
		Y:       20,
		SrcX:    0,
		SrcY:    0,
		Width:   32,
		Height:  32,
		Visible: true,
		ZOrder:  0,
	}

	srcImage := ebiten.NewImage(64, 64)

	// 親スプライト付きでCastSpriteを作成
	cs := csm.CreateCastSpriteWithParent(cast, srcImage, 100, parentSprite)

	// 親が表示中の場合、子も実効的に表示
	if !cs.GetSprite().IsEffectivelyVisible() {
		t.Error("expected child to be effectively visible when parent is visible")
	}

	// 親を非表示にする
	parentSprite.SetVisible(false)

	// 親が非表示の場合、子も実効的に非表示
	if cs.GetSprite().IsEffectivelyVisible() {
		t.Error("expected child to be effectively invisible when parent is invisible")
	}
}

// TestCreateCastSpriteWithParentZPath はCreateCastSpriteWithParentでZ_Pathが正しく設定されることをテストする
// 要件 1.4: 子スプライトが作成されたとき、親のZ_Pathを継承し、自身のLocal_Z_Orderを追加する
// 要件 2.2: PutCastが呼び出されたとき、現在のZ_Order_Counterを使用してLocal_Z_Orderを割り当てる
func TestCreateCastSpriteWithParentZPath(t *testing.T) {
	sm := NewSpriteManager()
	csm := NewCastSpriteManager(sm)

	// 親スプライトを作成（ウインドウスプライトを模倣）
	parentSprite := sm.CreateSpriteWithSize(200, 150)
	parentSprite.SetPosition(100, 50)
	// 親にZ_Pathを設定（ウインドウのZ_Path）
	parentSprite.SetZPath(NewZPath(0))

	srcImage := ebiten.NewImage(64, 64)

	// 最初のキャストを作成
	cast1 := &Cast{
		ID:      0,
		WinID:   0,
		PicID:   1,
		X:       10,
		Y:       20,
		SrcX:    0,
		SrcY:    0,
		Width:   32,
		Height:  32,
		Visible: true,
		ZOrder:  0,
	}
	cs1 := csm.CreateCastSpriteWithParent(cast1, srcImage, 100, parentSprite)

	if cs1 == nil {
		t.Fatal("CreateCastSpriteWithParent returned nil")
	}

	// Z_Pathが設定されていることを確認
	zPath1 := cs1.GetSprite().GetZPath()
	if zPath1 == nil {
		t.Fatal("Z_Path should be set for cast sprite with parent")
	}

	// Z_Pathが親のZ_Pathを継承していることを確認
	// 親のZ_Path: [0], 子のZ_Path: [0, 0]
	expectedPath1 := []int{0, 0}
	actualPath1 := zPath1.Path()
	if len(actualPath1) != len(expectedPath1) {
		t.Errorf("expected Z_Path length %d, got %d", len(expectedPath1), len(actualPath1))
	}
	for i, v := range expectedPath1 {
		if actualPath1[i] != v {
			t.Errorf("expected Z_Path[%d] = %d, got %d", i, v, actualPath1[i])
		}
	}

	// 2番目のキャストを作成
	cast2 := &Cast{
		ID:      1,
		WinID:   0,
		PicID:   1,
		X:       30,
		Y:       40,
		SrcX:    0,
		SrcY:    0,
		Width:   32,
		Height:  32,
		Visible: true,
		ZOrder:  0,
	}
	cs2 := csm.CreateCastSpriteWithParent(cast2, srcImage, 101, parentSprite)

	if cs2 == nil {
		t.Fatal("CreateCastSpriteWithParent returned nil for second cast")
	}

	// 2番目のキャストのZ_Pathを確認
	zPath2 := cs2.GetSprite().GetZPath()
	if zPath2 == nil {
		t.Fatal("Z_Path should be set for second cast sprite")
	}

	// 2番目のキャストのZ_Path: [0, 1]（操作順序でインクリメント）
	expectedPath2 := []int{0, 1}
	actualPath2 := zPath2.Path()
	if len(actualPath2) != len(expectedPath2) {
		t.Errorf("expected Z_Path length %d, got %d", len(expectedPath2), len(actualPath2))
	}
	for i, v := range expectedPath2 {
		if actualPath2[i] != v {
			t.Errorf("expected Z_Path[%d] = %d, got %d", i, v, actualPath2[i])
		}
	}

	// Z_Pathの比較: cs1 < cs2（操作順序）
	if !zPath1.Less(zPath2) {
		t.Error("expected first cast Z_Path to be less than second cast Z_Path")
	}
}

// TestCreateCastSpriteWithTransColorAndParentZPath は透明色付きでZ_Pathが正しく設定されることをテストする
// 要件 1.4, 2.2, 2.6: 操作順序でZ順序を決定
func TestCreateCastSpriteWithTransColorAndParentZPath(t *testing.T) {
	sm := NewSpriteManager()
	csm := NewCastSpriteManager(sm)

	// 親スプライトを作成（ウインドウスプライトを模倣）
	parentSprite := sm.CreateSpriteWithSize(200, 150)
	parentSprite.SetPosition(100, 50)
	// 親にZ_Pathを設定（ウインドウのZ_Path）
	parentSprite.SetZPath(NewZPath(1))

	srcImage := ebiten.NewImage(64, 64)

	// キャストを作成
	cast := &Cast{
		ID:      0,
		WinID:   0,
		PicID:   1,
		X:       10,
		Y:       20,
		SrcX:    0,
		SrcY:    0,
		Width:   32,
		Height:  32,
		Visible: true,
		ZOrder:  0,
	}
	cs := csm.CreateCastSpriteWithTransColorAndParent(cast, srcImage, 100, nil, parentSprite)

	if cs == nil {
		t.Fatal("CreateCastSpriteWithTransColorAndParent returned nil")
	}

	// Z_Pathが設定されていることを確認
	zPath := cs.GetSprite().GetZPath()
	if zPath == nil {
		t.Fatal("Z_Path should be set for cast sprite with parent")
	}

	// Z_Pathが親のZ_Pathを継承していることを確認
	// 親のZ_Path: [1], 子のZ_Path: [1, 0]
	expectedPath := []int{1, 0}
	actualPath := zPath.Path()
	if len(actualPath) != len(expectedPath) {
		t.Errorf("expected Z_Path length %d, got %d", len(expectedPath), len(actualPath))
	}
	for i, v := range expectedPath {
		if actualPath[i] != v {
			t.Errorf("expected Z_Path[%d] = %d, got %d", i, v, actualPath[i])
		}
	}
}

// TestCastSpriteZPathOperationOrder は操作順序でZ順序が決定されることをテストする
// 要件 2.6: タイプ（キャスト、テキスト、ピクチャ）に関係なく、操作順でZ順序を決定する
func TestCastSpriteZPathOperationOrder(t *testing.T) {
	sm := NewSpriteManager()
	csm := NewCastSpriteManager(sm)

	// 親スプライトを作成
	parentSprite := sm.CreateSpriteWithSize(200, 150)
	parentSprite.SetZPath(NewZPath(0))

	srcImage := ebiten.NewImage(64, 64)

	// 複数のキャストを順番に作成
	casts := make([]*CastSprite, 5)
	for i := 0; i < 5; i++ {
		cast := &Cast{
			ID:      i,
			WinID:   0,
			PicID:   1,
			X:       i * 10,
			Y:       i * 10,
			SrcX:    0,
			SrcY:    0,
			Width:   32,
			Height:  32,
			Visible: true,
			ZOrder:  0,
		}
		casts[i] = csm.CreateCastSpriteWithParent(cast, srcImage, 100+i, parentSprite)
	}

	// 各キャストのZ_Pathが操作順序を反映していることを確認
	for i := 0; i < 5; i++ {
		zPath := casts[i].GetSprite().GetZPath()
		if zPath == nil {
			t.Fatalf("Z_Path should be set for cast %d", i)
		}

		expectedPath := []int{0, i}
		actualPath := zPath.Path()
		if len(actualPath) != len(expectedPath) {
			t.Errorf("cast %d: expected Z_Path length %d, got %d", i, len(expectedPath), len(actualPath))
		}
		for j, v := range expectedPath {
			if actualPath[j] != v {
				t.Errorf("cast %d: expected Z_Path[%d] = %d, got %d", i, j, v, actualPath[j])
			}
		}
	}

	// 後から作成されたキャストが前面にあることを確認
	for i := 0; i < 4; i++ {
		zPath1 := casts[i].GetSprite().GetZPath()
		zPath2 := casts[i+1].GetSprite().GetZPath()
		if !zPath1.Less(zPath2) {
			t.Errorf("expected cast %d Z_Path to be less than cast %d Z_Path", i, i+1)
		}
	}
}
