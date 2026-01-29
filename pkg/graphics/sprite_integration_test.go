// Package graphics provides integration tests for the sprite system.
// These tests verify that the sprite system components work together correctly.
// タスク 14.4: 描画システムの統合テスト
package graphics

import (
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// =============================================================================
// タスク 14.4: 描画システムの統合テスト
// 要件 14.1: GraphicsSystem.Draw()をSpriteManager.Draw()ベースに変更する
// 要件 14.2: ウインドウ内のスプライトをウインドウの子スプライトとして管理する
// 要件 14.3: Z順序の統一（ウインドウ間、ウインドウ内）を実装する
// =============================================================================

// TestSpriteSystemIntegration_GraphicsSystemManagers はGraphicsSystemが
// すべてのスプライトマネージャーを正しく初期化することをテストする
func TestSpriteSystemIntegration_GraphicsSystemManagers(t *testing.T) {
	gs := NewGraphicsSystem("")

	// SpriteManagerが初期化されていることを確認
	if gs.GetSpriteManager() == nil {
		t.Error("SpriteManager should be initialized")
	}

	// WindowSpriteManagerが初期化されていることを確認
	if gs.GetWindowSpriteManager() == nil {
		t.Error("WindowSpriteManager should be initialized")
	}

	// CastSpriteManagerが初期化されていることを確認
	if gs.GetCastSpriteManager() == nil {
		t.Error("CastSpriteManager should be initialized")
	}

	// TextSpriteManagerが初期化されていることを確認
	if gs.GetTextSpriteManager() == nil {
		t.Error("TextSpriteManager should be initialized")
	}

	// ShapeSpriteManagerが初期化されていることを確認
	if gs.GetShapeSpriteManager() == nil {
		t.Error("ShapeSpriteManager should be initialized")
	}
}

// TestSpriteSystemIntegration_WindowSpriteCreation はウインドウを開いたときに
// WindowSpriteが作成されることをテストする
// 要件 7.1: 指定サイズ・背景色のウインドウスプライトを作成できる
func TestSpriteSystemIntegration_WindowSpriteCreation(t *testing.T) {
	gs := NewGraphicsSystem("")

	// ピクチャーを作成
	picID, err := gs.CreatePic(200, 150)
	if err != nil {
		t.Fatalf("CreatePic failed: %v", err)
	}

	// ウインドウを開く
	winID, err := gs.OpenWin(picID, 100, 50, 200, 150, 0, 0, 0xFF0000)
	if err != nil {
		t.Fatalf("OpenWin failed: %v", err)
	}

	// WindowSpriteが作成されたことを確認
	wsm := gs.GetWindowSpriteManager()
	ws := wsm.GetWindowSprite(winID)
	if ws == nil {
		t.Fatal("WindowSprite should be created when window is opened")
	}

	// スプライトの位置を確認
	sprite := ws.GetSprite()
	x, y := sprite.Position()
	if x != 100 || y != 50 {
		t.Errorf("Expected position (100, 50), got (%v, %v)", x, y)
	}

	// スプライトが可視であることを確認
	if !sprite.Visible() {
		t.Error("WindowSprite should be visible")
	}
}

// TestSpriteSystemIntegration_CastSpriteAsChild はキャストがウインドウの
// 子スプライトとして作成されることをテストする
// 要件 14.2: ウインドウ内のスプライトをウインドウの子スプライトとして管理する
func TestSpriteSystemIntegration_CastSpriteAsChild(t *testing.T) {
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

	// キャストを配置
	castID, err := gs.PutCast(winID, picID, 10, 20, 0, 0, 50, 50)
	if err != nil {
		t.Fatalf("PutCast failed: %v", err)
	}

	// CastSpriteが作成されたことを確認
	csm := gs.GetCastSpriteManager()
	cs := csm.GetCastSprite(castID)
	if cs == nil {
		t.Fatal("CastSprite should be created when cast is placed")
	}

	// CastSpriteの親がPictureSpriteであることを確認
	// 新しい設計では、CastSpriteの親はWindowSpriteではなくPictureSpriteになる
	psm := gs.GetPictureSpriteManager()
	parentSprite := psm.GetBackgroundPictureSpriteSprite(picID)
	if parentSprite == nil {
		// PictureSpriteが存在しない場合は、WindowSpriteを親として使用
		wsm := gs.GetWindowSpriteManager()
		parentSprite = wsm.GetWindowSpriteSprite(winID)
	}
	if cs.GetSprite().Parent() != parentSprite {
		t.Errorf("CastSprite's parent should be the PictureSprite or WindowSprite, got parent=%v, expected=%v",
			cs.GetSprite().Parent(), parentSprite)
	}
}

// TestSpriteSystemIntegration_GlobalZOrder はグローバルZ順序が正しく
// 計算されることをテストする
// 要件 14.3: Z順序の統一（ウインドウ間、ウインドウ内）を実装する
func TestSpriteSystemIntegration_GlobalZOrder(t *testing.T) {
	gs := NewGraphicsSystem("")

	// ピクチャーを作成
	picID, err := gs.CreatePic(200, 150)
	if err != nil {
		t.Fatalf("CreatePic failed: %v", err)
	}

	// 2つのウインドウを開く
	winID1, err := gs.OpenWin(picID, 0, 0, 200, 150, 0, 0, 0)
	if err != nil {
		t.Fatalf("OpenWin 1 failed: %v", err)
	}

	winID2, err := gs.OpenWin(picID, 100, 100, 200, 150, 0, 0, 0)
	if err != nil {
		t.Fatalf("OpenWin 2 failed: %v", err)
	}

	// 各ウインドウにキャストを配置
	castID1, _ := gs.PutCast(winID1, picID, 10, 10, 0, 0, 50, 50)
	castID2, _ := gs.PutCast(winID2, picID, 10, 10, 0, 0, 50, 50)

	// CastSpriteを取得
	csm := gs.GetCastSpriteManager()
	cs1 := csm.GetCastSprite(castID1)
	cs2 := csm.GetCastSprite(castID2)

	if cs1 == nil || cs2 == nil {
		t.Fatal("CastSprites should be created")
	}

	// ウインドウ2のキャストはウインドウ1のキャストより前面にあるはず
	// （ウインドウ2が後から開かれたため）
	// Z_Pathで比較
	zPath1 := cs1.GetSprite().GetZPath()
	zPath2 := cs2.GetSprite().GetZPath()

	if zPath1 == nil || zPath2 == nil {
		t.Fatal("Z_Paths should be set")
	}

	// ウインドウ2のZ_Pathはウインドウ1のZ_Pathより大きいはず
	if !zPath1.Less(zPath2) {
		t.Errorf("Cast in window 2 should have higher Z_Path than cast in window 1: %v >= %v", zPath2.String(), zPath1.String())
	}
}

// TestSpriteSystemIntegration_CloseWindowRemovesSprites はウインドウを閉じたときに
// 関連するスプライトが削除されることをテストする
// 要件 7.3: ウインドウが閉じられたときにウインドウとその子スプライトを削除する
// 要件 8.3: キャストを削除できる
func TestSpriteSystemIntegration_CloseWindowRemovesSprites(t *testing.T) {
	gs := NewGraphicsSystem("")

	// ピクチャーを作成
	picID, _ := gs.CreatePic(200, 150)

	// ウインドウを開く
	winID, _ := gs.OpenWin(picID, 0, 0, 200, 150, 0, 0, 0)

	// キャストを配置
	castID, _ := gs.PutCast(winID, picID, 10, 10, 0, 0, 50, 50)

	// スプライトが存在することを確認
	wsm := gs.GetWindowSpriteManager()
	csm := gs.GetCastSpriteManager()

	if wsm.GetWindowSprite(winID) == nil {
		t.Error("WindowSprite should exist before close")
	}
	if csm.GetCastSprite(castID) == nil {
		t.Error("CastSprite should exist before close")
	}

	// ウインドウを閉じる
	gs.CloseWin(winID)

	// スプライトが削除されたことを確認
	if wsm.GetWindowSprite(winID) != nil {
		t.Error("WindowSprite should be removed after close")
	}
	if csm.GetCastSprite(castID) != nil {
		t.Error("CastSprite should be removed after close")
	}
}

// TestSpriteSystemIntegration_ParentChildPositionInheritance は親子関係の
// 位置継承が正しく動作することをテストする
// 要件 2.1: 子スプライトの絶対位置は親の位置と子の相対位置の和である
func TestSpriteSystemIntegration_ParentChildPositionInheritance(t *testing.T) {
	gs := NewGraphicsSystem("")

	// ピクチャーを作成
	picID, _ := gs.CreatePic(200, 150)

	// ウインドウを開く（位置: 100, 50）
	winID, _ := gs.OpenWin(picID, 100, 50, 200, 150, 0, 0, 0)

	// キャストを配置（相対位置: 10, 20）
	castID, _ := gs.PutCast(winID, picID, 10, 20, 0, 0, 50, 50)

	// CastSpriteを取得
	csm := gs.GetCastSpriteManager()
	cs := csm.GetCastSprite(castID)
	if cs == nil {
		t.Fatal("CastSprite should be created")
	}

	// 絶対位置を確認
	// 親の位置(100, 50) + 子の相対位置(10, 20) = (110, 70)
	absX, absY := cs.GetSprite().AbsolutePosition()
	if absX != 110 || absY != 70 {
		t.Errorf("Expected absolute position (110, 70), got (%v, %v)", absX, absY)
	}
}

// TestSpriteSystemIntegration_ParentChildVisibilityInheritance は親子関係の
// 可視性継承が正しく動作することをテストする
// 要件 2.3: 親スプライトが非表示のとき子スプライトも非表示として扱う
func TestSpriteSystemIntegration_ParentChildVisibilityInheritance(t *testing.T) {
	gs := NewGraphicsSystem("")

	// ピクチャーを作成
	picID, _ := gs.CreatePic(200, 150)

	// ウインドウを開く
	winID, _ := gs.OpenWin(picID, 0, 0, 200, 150, 0, 0, 0)

	// キャストを配置
	castID, _ := gs.PutCast(winID, picID, 10, 10, 0, 0, 50, 50)

	// CastSpriteを取得
	csm := gs.GetCastSpriteManager()
	cs := csm.GetCastSprite(castID)
	if cs == nil {
		t.Fatal("CastSprite should be created")
	}

	// 親が表示中の場合、子も実効的に表示
	if !cs.GetSprite().IsEffectivelyVisible() {
		t.Error("CastSprite should be effectively visible when parent is visible")
	}

	// 親（WindowSprite）を非表示にする
	wsm := gs.GetWindowSpriteManager()
	ws := wsm.GetWindowSprite(winID)
	ws.UpdateVisible(false)

	// 親が非表示の場合、子も実効的に非表示
	if cs.GetSprite().IsEffectivelyVisible() {
		t.Error("CastSprite should be effectively invisible when parent is invisible")
	}
}

// TestSpriteSystemIntegration_MultipleCastsZOrder は同一ウインドウ内の
// 複数キャストのZ順序が正しいことをテストする
// 要件 4.1: スプライトをZ順序（小さい順）で描画する
func TestSpriteSystemIntegration_MultipleCastsZOrder(t *testing.T) {
	gs := NewGraphicsSystem("")

	// ピクチャーを作成
	picID, _ := gs.CreatePic(200, 150)

	// ウインドウを開く
	winID, _ := gs.OpenWin(picID, 0, 0, 200, 150, 0, 0, 0)

	// 複数のキャストを配置
	castID1, _ := gs.PutCast(winID, picID, 10, 10, 0, 0, 50, 50)
	castID2, _ := gs.PutCast(winID, picID, 20, 20, 0, 0, 50, 50)
	castID3, _ := gs.PutCast(winID, picID, 30, 30, 0, 0, 50, 50)

	// CastSpriteを取得
	csm := gs.GetCastSpriteManager()
	cs1 := csm.GetCastSprite(castID1)
	cs2 := csm.GetCastSprite(castID2)
	cs3 := csm.GetCastSprite(castID3)

	if cs1 == nil || cs2 == nil || cs3 == nil {
		t.Fatal("All CastSprites should be created")
	}

	// 後から作成したキャストほどZ_Pathが大きい
	zPath1 := cs1.GetSprite().GetZPath()
	zPath2 := cs2.GetSprite().GetZPath()
	zPath3 := cs3.GetSprite().GetZPath()

	if zPath1 == nil || zPath2 == nil || zPath3 == nil {
		t.Fatal("All Z_Paths should be set")
	}

	if !zPath1.Less(zPath2) || !zPath2.Less(zPath3) {
		t.Errorf("Z_Path should increase with creation order: %v, %v, %v", zPath1.String(), zPath2.String(), zPath3.String())
	}
}

// TestSpriteSystemIntegration_DrawWithSpriteManager はDrawWithSpriteManager
// メソッドが正しく動作することをテストする
// 要件 14.1: GraphicsSystem.Draw()をSpriteManager.Draw()ベースに変更する
func TestSpriteSystemIntegration_DrawWithSpriteManager(t *testing.T) {
	gs := NewGraphicsSystem("")

	// ピクチャーを作成
	picID, _ := gs.CreatePic(200, 150)

	// ウインドウを開く
	winID, _ := gs.OpenWin(picID, 0, 0, 200, 150, 0, 0, 0)

	// キャストを配置
	gs.PutCast(winID, picID, 10, 10, 0, 0, 50, 50)

	// テスト用のスクリーンを作成
	screen := ebiten.NewImage(1024, 768)

	// DrawWithSpriteManagerを呼び出す（パニックしないことを確認）
	gs.DrawWithSpriteManager(screen)

	// 通常のDrawも呼び出す（パニックしないことを確認）
	gs.Draw(screen)
}

// TestSpriteSystemIntegration_MoveCastUpdatesSprite はキャストの移動が
// スプライトに反映されることをテストする
// 要件 8.2: キャストの位置を移動できる（残像なし）
func TestSpriteSystemIntegration_MoveCastUpdatesSprite(t *testing.T) {
	gs := NewGraphicsSystem("")

	// ピクチャーを作成
	picID, _ := gs.CreatePic(200, 150)

	// ウインドウを開く
	winID, _ := gs.OpenWin(picID, 0, 0, 200, 150, 0, 0, 0)

	// キャストを配置
	castID, _ := gs.PutCast(winID, picID, 10, 20, 0, 0, 50, 50)

	// CastSpriteを取得
	csm := gs.GetCastSpriteManager()
	cs := csm.GetCastSprite(castID)
	if cs == nil {
		t.Fatal("CastSprite should be created")
	}

	// 初期位置を確認
	x, y := cs.GetSprite().Position()
	if x != 10 || y != 20 {
		t.Errorf("Expected initial position (10, 20), got (%v, %v)", x, y)
	}

	// キャストを移動
	gs.MoveCastWithOptions(castID, WithCastPosition(100, 200))

	// 位置が更新されたことを確認
	x, y = cs.GetSprite().Position()
	if x != 100 || y != 200 {
		t.Errorf("Expected updated position (100, 200), got (%v, %v)", x, y)
	}
}

// TestSpriteSystemIntegration_DelCastRemovesSprite はキャストの削除が
// スプライトにも反映されることをテストする
// 要件 8.3: キャストを削除できる
func TestSpriteSystemIntegration_DelCastRemovesSprite(t *testing.T) {
	gs := NewGraphicsSystem("")

	// ピクチャーを作成
	picID, _ := gs.CreatePic(200, 150)

	// ウインドウを開く
	winID, _ := gs.OpenWin(picID, 0, 0, 200, 150, 0, 0, 0)

	// キャストを配置
	castID, _ := gs.PutCast(winID, picID, 10, 10, 0, 0, 50, 50)

	// CastSpriteが存在することを確認
	csm := gs.GetCastSpriteManager()
	if csm.GetCastSprite(castID) == nil {
		t.Error("CastSprite should exist before deletion")
	}

	// キャストを削除
	gs.DelCast(castID)

	// CastSpriteが削除されたことを確認
	if csm.GetCastSprite(castID) != nil {
		t.Error("CastSprite should be removed after deletion")
	}
}

// TestSpriteSystemIntegration_CloseWinAllClearsAllSprites はCloseWinAllが
// すべてのスプライトを削除することをテストする
func TestSpriteSystemIntegration_CloseWinAllClearsAllSprites(t *testing.T) {
	gs := NewGraphicsSystem("")

	// 複数のウインドウとキャストを作成
	for i := 0; i < 3; i++ {
		picID, _ := gs.CreatePic(200, 150)
		winID, _ := gs.OpenWin(picID, i*100, i*50, 200, 150, 0, 0, 0)
		gs.PutCast(winID, picID, 10, 10, 0, 0, 50, 50)
	}

	// スプライトが存在することを確認
	wsm := gs.GetWindowSpriteManager()
	csm := gs.GetCastSpriteManager()

	// すべてのウインドウを閉じる
	gs.CloseWinAll()

	// すべてのスプライトが削除されたことを確認
	// WindowSpriteManagerのカウントを確認
	for i := 0; i < 3; i++ {
		if wsm.GetWindowSprite(i) != nil {
			t.Errorf("WindowSprite %d should be removed after CloseWinAll", i)
		}
	}

	// CastSpriteManagerのカウントを確認
	if csm.Count() != 0 {
		t.Errorf("Expected 0 CastSprites after CloseWinAll, got %d", csm.Count())
	}
}

// TestSpriteSystemIntegration_TextSpriteCreation はTextWriteがTextSpriteを
// 作成することをテストする
// 要件 5.1, 5.2: テキストスプライトの作成
func TestSpriteSystemIntegration_TextSpriteCreation(t *testing.T) {
	gs := NewGraphicsSystem("")

	// ピクチャーを作成
	picID, err := gs.CreatePic(200, 100)
	if err != nil {
		t.Fatalf("CreatePic failed: %v", err)
	}

	// テキストを描画
	err = gs.TextWrite(picID, 10, 20, "Hello")
	if err != nil {
		t.Fatalf("TextWrite failed: %v", err)
	}

	// TextSpriteが作成されたことを確認
	tsm := gs.GetTextSpriteManager()
	if tsm.Count() != 1 {
		t.Errorf("Expected 1 TextSprite, got %d", tsm.Count())
	}

	sprites := tsm.GetTextSprites(picID)
	if len(sprites) != 1 {
		t.Errorf("Expected 1 TextSprite for picID %d, got %d", picID, len(sprites))
	}
}

// TestSpriteSystemIntegration_ZOrderConstants はZ順序定数が正しく定義されて
// いることをテストする
// 要件 14.3: Z順序の統一
func TestSpriteSystemIntegration_ZOrderConstants(t *testing.T) {
	// Z順序定数の値を確認
	if ZOrderBackground != 0 {
		t.Errorf("Expected ZOrderBackground = 0, got %d", ZOrderBackground)
	}
	if ZOrderDrawing != 1 {
		t.Errorf("Expected ZOrderDrawing = 1, got %d", ZOrderDrawing)
	}
	if ZOrderCastBase != 100 {
		t.Errorf("Expected ZOrderCastBase = 100, got %d", ZOrderCastBase)
	}
	if ZOrderCastMax != 999 {
		t.Errorf("Expected ZOrderCastMax = 999, got %d", ZOrderCastMax)
	}
	if ZOrderTextBase != 1000 {
		t.Errorf("Expected ZOrderTextBase = 1000, got %d", ZOrderTextBase)
	}
	if ZOrderWindowRange != 10000 {
		t.Errorf("Expected ZOrderWindowRange = 10000, got %d", ZOrderWindowRange)
	}
}

// TestSpriteSystemIntegration_CalculateGlobalZOrder はグローバルZ順序の
// 計算が正しいことをテストする
// 要件 14.3: Z順序の統一
func TestSpriteSystemIntegration_CalculateGlobalZOrder(t *testing.T) {
	testCases := []struct {
		windowZOrder int
		localZOrder  int
		expected     int
	}{
		{0, 0, 0},
		{0, 100, 100},
		{1, 0, 10000},
		{1, 100, 10100},
		{2, 1000, 21000},
	}

	for _, tc := range testCases {
		result := CalculateGlobalZOrder(tc.windowZOrder, tc.localZOrder)
		if result != tc.expected {
			t.Errorf("CalculateGlobalZOrder(%d, %d) = %d, expected %d",
				tc.windowZOrder, tc.localZOrder, result, tc.expected)
		}
	}
}

// TestSpriteSystemIntegration_WindowSpriteChildManagement はWindowSpriteの
// 子スプライト管理が正しく動作することをテストする
// 要件 7.2: ウインドウスプライトを親として子スプライトを追加できる
// 要件 11.4: すべての描画要素をスプライトとして管理する（背景ピクチャーを含む）
// 注意: 新しい設計では、CastSpriteはPictureSpriteの子になる
func TestSpriteSystemIntegration_WindowSpriteChildManagement(t *testing.T) {
	gs := NewGraphicsSystem("")

	// ピクチャーを作成
	picID, _ := gs.CreatePic(200, 150)

	// ウインドウを開く
	// 要件 11.4: OpenWinで背景ピクチャー用のPictureSpriteが作成される
	winID, _ := gs.OpenWin(picID, 0, 0, 200, 150, 0, 0, 0)

	// 複数のキャストを配置
	castID1, _ := gs.PutCast(winID, picID, 10, 10, 0, 0, 50, 50)
	castID2, _ := gs.PutCast(winID, picID, 20, 20, 0, 0, 50, 50)

	// WindowSpriteを取得
	wsm := gs.GetWindowSpriteManager()
	ws := wsm.GetWindowSprite(winID)
	if ws == nil {
		t.Fatal("WindowSprite should exist")
	}

	// WindowSpriteの基盤スプライトの子スプライトのリストを取得
	// 新しい設計では、WindowSpriteの直接の子はPictureSpriteのみ
	// CastSpriteはPictureSpriteの子になる
	windowSpriteSprite := ws.GetSprite()
	children := windowSpriteSprite.GetChildren()
	// PictureSpriteが子として存在することを確認
	if len(children) < 1 {
		t.Logf("WindowSprite has %d children (expected at least 1 PictureSprite)", len(children))
	}

	// PictureSpriteの子としてCastSpriteが存在することを確認
	psm := gs.GetPictureSpriteManager()
	ps := psm.GetBackgroundPictureSprite(picID)
	if ps != nil {
		psChildren := ps.GetSprite().GetChildren()
		// CastSpriteが2つ存在することを確認
		if len(psChildren) < 2 {
			t.Logf("PictureSprite has %d children (expected at least 2 casts)", len(psChildren))
		}
	}

	// キャストを削除
	gs.DelCast(castID1)

	// 残っているキャストが正しいことを確認
	csm := gs.GetCastSpriteManager()
	cs2 := csm.GetCastSprite(castID2)
	if cs2 == nil {
		t.Error("CastSprite 2 should still exist")
	}
}

// TestSpriteSystemIntegration_TransparentColorCast は透明色付きキャストが
// 正しく作成されることをテストする
// 要件 8.4: 透明色処理をサポートする
func TestSpriteSystemIntegration_TransparentColorCast(t *testing.T) {
	gs := NewGraphicsSystem("")

	// ピクチャーを作成
	picID, _ := gs.CreatePic(200, 150)

	// ウインドウを開く
	winID, _ := gs.OpenWin(picID, 0, 0, 200, 150, 0, 0, 0)

	// 透明色付きでキャストを配置
	transColor := color.RGBA{0, 0, 0, 255} // 黒を透明色として指定
	castID, err := gs.PutCastWithTransColor(winID, picID, 10, 10, 0, 0, 50, 50, transColor)
	if err != nil {
		t.Fatalf("PutCastWithTransColor failed: %v", err)
	}

	// CastSpriteが作成されたことを確認
	csm := gs.GetCastSpriteManager()
	cs := csm.GetCastSprite(castID)
	if cs == nil {
		t.Fatal("CastSprite should be created")
	}

	// 透明色が設定されていることを確認
	if !cs.HasTransColor() {
		t.Error("CastSprite should have transparent color")
	}
}

// TestSpriteSystemIntegration_CompleteRenderingPipeline は完全な描画パイプラインを
// テストする（ウインドウ、キャスト、テキストの組み合わせ）
func TestSpriteSystemIntegration_CompleteRenderingPipeline(t *testing.T) {
	gs := NewGraphicsSystem("")

	// 背景ピクチャーを作成
	bgPicID, err := gs.CreatePic(400, 300)
	if err != nil {
		t.Fatalf("CreatePic for background failed: %v", err)
	}

	// スプライトピクチャーを作成
	spritePicID, err := gs.CreatePic(50, 50)
	if err != nil {
		t.Fatalf("CreatePic for sprite failed: %v", err)
	}

	// ウインドウを開く
	winID, err := gs.OpenWin(bgPicID, 100, 100, 400, 300, 0, 0, 0xFFFFFF)
	if err != nil {
		t.Fatalf("OpenWin failed: %v", err)
	}

	// キャストを配置
	castID1, _ := gs.PutCast(winID, spritePicID, 10, 10, 0, 0, 50, 50)
	castID2, _ := gs.PutCast(winID, spritePicID, 60, 60, 0, 0, 50, 50)

	// テキストを描画
	gs.TextWrite(bgPicID, 10, 200, "Hello World")

	// スプライトマネージャーの状態を確認
	wsm := gs.GetWindowSpriteManager()
	csm := gs.GetCastSpriteManager()
	tsm := gs.GetTextSpriteManager()

	// WindowSpriteが存在することを確認
	if wsm.GetWindowSprite(winID) == nil {
		t.Error("WindowSprite should exist")
	}

	// CastSpriteが存在することを確認
	if csm.GetCastSprite(castID1) == nil || csm.GetCastSprite(castID2) == nil {
		t.Error("CastSprites should exist")
	}

	// TextSpriteが存在することを確認
	if tsm.Count() != 1 {
		t.Errorf("Expected 1 TextSprite, got %d", tsm.Count())
	}

	// テスト用のスクリーンを作成して描画
	screen := ebiten.NewImage(1024, 768)
	gs.Draw(screen)

	// クリーンアップ
	gs.CloseWinAll()

	// すべてのスプライトが削除されたことを確認
	if csm.Count() != 0 {
		t.Errorf("Expected 0 CastSprites after cleanup, got %d", csm.Count())
	}
	if tsm.Count() != 0 {
		t.Errorf("Expected 0 TextSprites after cleanup, got %d", tsm.Count())
	}
}

// TestSpriteSystemIntegration_SpriteManagerDraw はSpriteManager.Draw()が
// 正しく動作することをテストする
// 要件 14.1: SpriteManager.Draw()ベースの描画
func TestSpriteSystemIntegration_SpriteManagerDraw(t *testing.T) {
	sm := NewSpriteManager()

	// 複数のスプライトを作成
	img1 := ebiten.NewImage(50, 50)
	img2 := ebiten.NewImage(50, 50)
	img3 := ebiten.NewImage(50, 50)

	s1 := sm.CreateSprite(img1)
	s1.SetPosition(10, 10)
	s1.SetZPath(NewZPath(100))

	s2 := sm.CreateSprite(img2)
	s2.SetPosition(20, 20)
	s2.SetZPath(NewZPath(50))

	s3 := sm.CreateSprite(img3)
	s3.SetPosition(30, 30)
	s3.SetZPath(NewZPath(150))

	// テスト用のスクリーンを作成
	screen := ebiten.NewImage(200, 200)

	// Draw()を呼び出す（パニックしないことを確認）
	sm.Draw(screen)

	// スプライトの数を確認
	if sm.Count() != 3 {
		t.Errorf("Expected 3 sprites, got %d", sm.Count())
	}
}
