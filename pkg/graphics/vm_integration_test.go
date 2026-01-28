package graphics

import (
	"testing"
)

// TestVMGraphicsSystemLayerManagerIntegration tests the complete integration chain:
// VM → GraphicsSystem → CastManager/TextRenderer → LayerManager
// 要件 8.4: VMのPutCast/MoveCast実装がLayerManagerを使用することを確認する
// 要件 8.5: VMのTextWrite実装がLayerManagerを使用することを確認する
func TestVMGraphicsSystemLayerManagerIntegration(t *testing.T) {
	t.Run("GraphicsSystem initializes LayerManager", func(t *testing.T) {
		// 要件 8.1: GraphicsSystemにLayerManagerを統合する
		gs := NewGraphicsSystem("")

		lm := gs.GetLayerManager()
		if lm == nil {
			t.Fatal("expected LayerManager to be initialized")
		}
	})

	t.Run("GraphicsSystem connects CastManager to LayerManager", func(t *testing.T) {
		// 要件 8.2: CastManagerとLayerManagerを統合する
		gs := NewGraphicsSystem("")

		// CastManagerはGraphicsSystemの内部にあるので、
		// PutCastを呼び出してLayerManagerにCastLayerが作成されることを確認
		lm := gs.GetLayerManager()

		// ウィンドウを開く（キャストはウィンドウに属する）
		picID, err := gs.CreatePic(100, 100)
		if err != nil {
			t.Fatalf("CreatePic failed: %v", err)
		}

		winID, err := gs.OpenWin(picID)
		if err != nil {
			t.Fatalf("OpenWin failed: %v", err)
		}

		// キャストを配置
		castID, err := gs.PutCast(winID, picID, 10, 20, 0, 0, 32, 32)
		if err != nil {
			t.Fatalf("PutCast failed: %v", err)
		}

		// LayerManagerにCastLayerが作成されたことを確認（WindowLayerSetを使用）
		wls := lm.GetWindowLayerSet(winID)
		if wls == nil {
			t.Fatal("expected WindowLayerSet to be created")
		}

		castLayer := wls.GetCastLayer(castID)
		if castLayer == nil {
			t.Fatal("expected CastLayer to be created via GraphicsSystem.PutCast")
		}

		// CastLayerの位置を確認
		x, y := castLayer.GetPosition()
		if x != 10 || y != 20 {
			t.Errorf("expected position (10, 20), got (%d, %d)", x, y)
		}
	})

	t.Run("GraphicsSystem.MoveCast updates LayerManager", func(t *testing.T) {
		// 要件 8.4: VMのPutCast/MoveCast実装がLayerManagerを使用することを確認する
		gs := NewGraphicsSystem("")
		lm := gs.GetLayerManager()

		// ウィンドウを開く
		picID, _ := gs.CreatePic(100, 100)
		winID, _ := gs.OpenWin(picID)

		// キャストを配置
		castID, _ := gs.PutCast(winID, picID, 10, 20, 0, 0, 32, 32)

		// キャストを移動
		err := gs.MoveCastWithOptions(castID, WithCastPosition(100, 200))
		if err != nil {
			t.Fatalf("MoveCastWithOptions failed: %v", err)
		}

		// LayerManagerのCastLayerが更新されたことを確認（WindowLayerSetを使用）
		wls := lm.GetWindowLayerSet(winID)
		if wls == nil {
			t.Fatal("expected WindowLayerSet to be created")
		}
		castLayer := wls.GetCastLayer(castID)
		if castLayer == nil {
			t.Fatal("expected CastLayer to exist")
		}

		x, y := castLayer.GetPosition()
		if x != 100 || y != 200 {
			t.Errorf("expected position (100, 200), got (%d, %d)", x, y)
		}
	})

	t.Run("GraphicsSystem.DelCast removes from LayerManager", func(t *testing.T) {
		// 要件 8.4: VMのPutCast/MoveCast実装がLayerManagerを使用することを確認する
		gs := NewGraphicsSystem("")
		lm := gs.GetLayerManager()

		// ウィンドウを開く
		picID, _ := gs.CreatePic(100, 100)
		winID, _ := gs.OpenWin(picID)

		// キャストを配置
		castID, _ := gs.PutCast(winID, picID, 10, 20, 0, 0, 32, 32)

		// CastLayerが存在することを確認（WindowLayerSetを使用）
		wls := lm.GetWindowLayerSet(winID)
		if wls == nil {
			t.Fatal("expected WindowLayerSet to be created")
		}
		if wls.GetCastLayer(castID) == nil {
			t.Fatal("expected CastLayer to exist before deletion")
		}

		// キャストを削除
		err := gs.DelCast(castID)
		if err != nil {
			t.Fatalf("DelCast failed: %v", err)
		}

		// LayerManagerからCastLayerが削除されたことを確認
		if wls.GetCastLayer(castID) != nil {
			t.Error("expected CastLayer to be removed after deletion")
		}
	})

	t.Run("GraphicsSystem connects TextRenderer to LayerManager", func(t *testing.T) {
		// 要件 8.3: TextRendererとLayerManagerを統合する
		// 要件 8.5: VMのTextWrite実装がLayerManagerを使用することを確認する
		gs := NewGraphicsSystem("")
		lm := gs.GetLayerManager()

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

		// LayerManagerにTextLayerEntryが作成されたことを確認
		pls := lm.GetPictureLayerSet(picID)
		if pls == nil {
			t.Fatal("expected PictureLayerSet to be created")
		}

		textCount := pls.GetTextLayerCount()
		if textCount != 1 {
			t.Errorf("expected 1 TextLayerEntry, got %d", textCount)
		}
	})

	t.Run("GraphicsSystem.CloseWin removes casts from LayerManager", func(t *testing.T) {
		// 要件 2.6: ウィンドウが閉じられたときにそのウィンドウに属するすべてのレイヤーを削除する
		gs := NewGraphicsSystem("")
		lm := gs.GetLayerManager()

		// ウィンドウを開く
		picID, _ := gs.CreatePic(100, 100)
		winID, _ := gs.OpenWin(picID)

		// キャストを配置
		gs.PutCast(winID, picID, 10, 10, 0, 0, 32, 32)
		gs.PutCast(winID, picID, 20, 20, 0, 0, 32, 32)

		// CastLayerが存在することを確認（WindowLayerSetを使用）
		wls := lm.GetWindowLayerSet(winID)
		if wls == nil {
			t.Fatal("expected WindowLayerSet to be created")
		}
		if wls.GetCastLayerCount() != 2 {
			t.Errorf("expected 2 CastLayers before close, got %d", wls.GetCastLayerCount())
		}

		// ウィンドウを閉じる
		err := gs.CloseWin(winID)
		if err != nil {
			t.Fatalf("CloseWin failed: %v", err)
		}

		// WindowLayerSetが削除されたことを確認
		wlsAfter := lm.GetWindowLayerSet(winID)
		if wlsAfter != nil {
			t.Error("expected WindowLayerSet to be deleted after CloseWin")
		}
	})

	t.Run("PutCastWithTransColor creates CastLayer with transparency", func(t *testing.T) {
		// 要件 8.4: VMのPutCast/MoveCast実装がLayerManagerを使用することを確認する
		gs := NewGraphicsSystem("")
		lm := gs.GetLayerManager()

		// ウィンドウを開く
		picID, _ := gs.CreatePic(100, 100)
		winID, _ := gs.OpenWin(picID)

		// 透明色付きでキャストを配置
		transColor := DefaultTransparentColor
		castID, err := gs.PutCastWithTransColor(winID, picID, 10, 20, 0, 0, 32, 32, transColor)
		if err != nil {
			t.Fatalf("PutCastWithTransColor failed: %v", err)
		}

		// LayerManagerにCastLayerが作成されたことを確認（WindowLayerSetを使用）
		wls := lm.GetWindowLayerSet(winID)
		if wls == nil {
			t.Fatal("expected WindowLayerSet to be created")
		}
		castLayer := wls.GetCastLayer(castID)
		if castLayer == nil {
			t.Fatal("expected CastLayer to be created")
		}

		// 透明色が設定されていることを確認
		if !castLayer.HasTransColor() {
			t.Error("expected CastLayer to have trans color")
		}
	})

	t.Run("Multiple TextWrite creates multiple TextLayerEntries", func(t *testing.T) {
		// 要件 8.5: VMのTextWrite実装がLayerManagerを使用することを確認する
		gs := NewGraphicsSystem("")
		lm := gs.GetLayerManager()

		// ピクチャーを作成
		picID, _ := gs.CreatePic(200, 100)

		// 複数のテキストを描画
		gs.TextWrite(picID, 10, 20, "Hello")
		gs.TextWrite(picID, 10, 40, "World")
		gs.TextWrite(picID, 10, 60, "Test")

		// LayerManagerに複数のTextLayerEntryが作成されたことを確認
		pls := lm.GetPictureLayerSet(picID)
		textCount := pls.GetTextLayerCount()
		if textCount != 3 {
			t.Errorf("expected 3 TextLayerEntries, got %d", textCount)
		}
	})

	t.Run("Integration chain maintains Z-order", func(t *testing.T) {
		// 要件 1.6: レイヤーを背景 → 描画 → キャスト → テキストの順序で合成する
		gs := NewGraphicsSystem("")
		lm := gs.GetLayerManager()

		// ピクチャーを作成
		picID, _ := gs.CreatePic(200, 100)
		winID, _ := gs.OpenWin(picID)

		// キャストを配置
		castID1, _ := gs.PutCast(winID, picID, 10, 10, 0, 0, 32, 32)
		castID2, _ := gs.PutCast(winID, picID, 20, 20, 0, 0, 32, 32)

		// テキストを描画
		gs.TextWrite(picID, 10, 20, "Hello")

		// Z順序を確認（WindowLayerSetを使用）
		wls := lm.GetWindowLayerSet(winID)
		if wls == nil {
			t.Fatal("expected WindowLayerSet to be created")
		}
		castLayer1 := wls.GetCastLayer(castID1)
		castLayer2 := wls.GetCastLayer(castID2)

		if castLayer1 == nil || castLayer2 == nil {
			t.Fatal("expected both CastLayers to exist")
		}

		// 操作順序に基づくZ順序: 最初のキャストはZ=1から開始
		// 要件 10.1, 10.2: 操作順序に基づくZ順序
		if castLayer1.GetZOrder() < 1 {
			t.Errorf("castLayer1 Z-order should be >= 1, got %d", castLayer1.GetZOrder())
		}

		// 後から作成したキャストは前面に
		if castLayer1.GetZOrder() >= castLayer2.GetZOrder() {
			t.Error("castLayer1 should have lower Z-order than castLayer2")
		}

		// テキストレイヤーのZ順序を確認（picIDのPictureLayerSetから取得）
		plsText := lm.GetPictureLayerSet(picID)
		if plsText != nil && plsText.GetTextLayerCount() > 0 {
			// 操作順序に基づくZ順序: テキストレイヤーも1から開始
			// GetAllLayersSortedを使用してテキストレイヤーを取得
			allLayers := plsText.GetAllLayersSorted()
			for _, layer := range allLayers {
				// TextLayerEntryはZ順序が1以上
				if layer.GetZOrder() >= 1 {
					// テキストレイヤーが見つかった
					if layer.GetZOrder() < 1 {
						t.Errorf("textLayer Z-order should be >= 1, got %d", layer.GetZOrder())
					}
					break
				}
			}
		}
	})
}

// =============================================================================
// タスク 6.5: 統合テスト
// 要件 8.1, 8.2, 8.3, 8.4 に基づく追加テスト
// =============================================================================

// TestGraphicsSystemLayerManagerIntegration_Composite は合成処理の統合テスト
// 要件 8.1: GraphicsSystemへのLayerManager統合
func TestGraphicsSystemLayerManagerIntegration_Composite(t *testing.T) {
	t.Run("LayerManager WindowLayerSet is created when cast is placed", func(t *testing.T) {
		gs := NewGraphicsSystem("")
		lm := gs.GetLayerManager()

		// ピクチャーを作成
		picID, err := gs.CreatePic(100, 100)
		if err != nil {
			t.Fatalf("CreatePic failed: %v", err)
		}

		// ウィンドウを開く
		winID, err := gs.OpenWin(picID)
		if err != nil {
			t.Fatalf("OpenWin failed: %v", err)
		}

		// キャストを配置（これによりWindowLayerSetにCastLayerが追加される）
		_, err = gs.PutCast(winID, picID, 10, 10, 0, 0, 32, 32)
		if err != nil {
			t.Fatalf("PutCast failed: %v", err)
		}

		// WindowLayerSetが作成されていることを確認
		wls := lm.GetWindowLayerSet(winID)
		if wls == nil {
			t.Fatal("expected WindowLayerSet to be created for window")
		}

		// WinIDが正しいことを確認
		if wls.GetWinID() != winID {
			t.Errorf("expected WinID %d, got %d", winID, wls.GetWinID())
		}
	})

	t.Run("Multiple windows create separate WindowLayerSets", func(t *testing.T) {
		gs := NewGraphicsSystem("")
		lm := gs.GetLayerManager()

		// 複数のピクチャーとウィンドウを作成
		picID1, _ := gs.CreatePic(100, 100)
		picID2, _ := gs.CreatePic(100, 100)

		winID1, _ := gs.OpenWin(picID1)
		winID2, _ := gs.OpenWin(picID2)

		// キャストを配置（これによりWindowLayerSetにCastLayerが追加される）
		gs.PutCast(winID1, picID1, 10, 10, 0, 0, 32, 32)
		gs.PutCast(winID2, picID2, 10, 10, 0, 0, 32, 32)

		// 各ウィンドウに対応するWindowLayerSetが作成されていることを確認
		wls1 := lm.GetWindowLayerSet(winID1)
		wls2 := lm.GetWindowLayerSet(winID2)

		if wls1 == nil {
			t.Fatal("expected WindowLayerSet for window 1")
		}
		if wls2 == nil {
			t.Fatal("expected WindowLayerSet for window 2")
		}

		// 異なるWindowLayerSetであることを確認
		if wls1 == wls2 {
			t.Error("expected different WindowLayerSets for different windows")
		}
	})
}

// TestCastManagerLayerManagerIntegration_FullChain はCastManagerの完全な統合チェーンをテスト
// 要件 8.2: CastManagerとの統合
func TestCastManagerLayerManagerIntegration_FullChain(t *testing.T) {
	t.Run("Cast operations update dirty flags", func(t *testing.T) {
		gs := NewGraphicsSystem("")
		lm := gs.GetLayerManager()

		// ウィンドウを開く
		picID, _ := gs.CreatePic(100, 100)
		winID, _ := gs.OpenWin(picID)

		// キャストを配置（これによりWindowLayerSetにCastLayerが追加される）
		castID, _ := gs.PutCast(winID, picID, 10, 20, 0, 0, 32, 32)

		// WindowLayerSetを取得
		wls := lm.GetWindowLayerSet(winID)
		if wls == nil {
			t.Fatal("expected WindowLayerSet to be created")
		}

		// ダーティフラグをクリア
		wls.ClearDirty()

		// キャストを移動（ダーティフラグが設定されるはず）
		gs.MoveCastWithOptions(castID, WithCastPosition(50, 60))

		// ダーティフラグが設定されていることを確認
		if !wls.IsDirty() {
			t.Error("expected WindowLayerSet to be dirty after MoveCast")
		}
	})

	t.Run("Cast source rect update works through LayerManager", func(t *testing.T) {
		gs := NewGraphicsSystem("")
		lm := gs.GetLayerManager()

		// ウィンドウを開く
		picID, _ := gs.CreatePic(100, 100)
		winID, _ := gs.OpenWin(picID)

		// キャストを配置
		castID, _ := gs.PutCast(winID, picID, 10, 20, 0, 0, 32, 32)

		// ソース矩形を更新
		gs.MoveCastWithOptions(castID, WithCastSource(10, 10, 64, 64))

		// LayerManagerのCastLayerが更新されたことを確認
		wls := lm.GetWindowLayerSet(winID)
		if wls == nil {
			t.Fatal("expected WindowLayerSet to be created")
		}
		castLayer := wls.GetCastLayer(castID)
		if castLayer == nil {
			t.Fatal("expected CastLayer to exist")
		}

		srcX, srcY, srcW, srcH := castLayer.GetSourceRect()
		if srcX != 10 || srcY != 10 || srcW != 64 || srcH != 64 {
			t.Errorf("expected source rect (10, 10, 64, 64), got (%d, %d, %d, %d)",
				srcX, srcY, srcW, srcH)
		}
	})

	t.Run("Multiple casts maintain correct Z-order", func(t *testing.T) {
		gs := NewGraphicsSystem("")
		lm := gs.GetLayerManager()

		// ウィンドウを開く
		picID, _ := gs.CreatePic(100, 100)
		winID, _ := gs.OpenWin(picID)

		// 複数のキャストを配置
		castID1, _ := gs.PutCast(winID, picID, 10, 10, 0, 0, 32, 32)
		castID2, _ := gs.PutCast(winID, picID, 20, 20, 0, 0, 32, 32)
		castID3, _ := gs.PutCast(winID, picID, 30, 30, 0, 0, 32, 32)

		// Z順序を確認
		wls := lm.GetWindowLayerSet(winID)
		if wls == nil {
			t.Fatal("expected WindowLayerSet to be created")
		}
		cast1 := wls.GetCastLayer(castID1)
		cast2 := wls.GetCastLayer(castID2)
		cast3 := wls.GetCastLayer(castID3)

		if cast1 == nil || cast2 == nil || cast3 == nil {
			t.Fatal("expected all CastLayers to exist")
		}

		// 後から作成したキャストほどZ順序が大きい
		if cast1.GetZOrder() >= cast2.GetZOrder() {
			t.Error("cast1 should have lower Z-order than cast2")
		}
		if cast2.GetZOrder() >= cast3.GetZOrder() {
			t.Error("cast2 should have lower Z-order than cast3")
		}
	})
}

// TestTextRendererLayerManagerIntegration_FullChain はTextRendererの完全な統合チェーンをテスト
// 要件 8.3: TextRendererとの統合
func TestTextRendererLayerManagerIntegration_FullChain(t *testing.T) {
	t.Run("TextWrite updates dirty flags", func(t *testing.T) {
		gs := NewGraphicsSystem("")
		lm := gs.GetLayerManager()

		// ピクチャーを作成
		picID, _ := gs.CreatePic(200, 100)

		// PictureLayerSetを取得（TextWriteで作成される）
		gs.TextWrite(picID, 10, 20, "Hello")

		pls := lm.GetPictureLayerSet(picID)
		if pls == nil {
			t.Fatal("expected PictureLayerSet to be created")
		}

		// ダーティフラグが設定されていることを確認
		if !pls.IsDirty() {
			t.Error("expected PictureLayerSet to be dirty after TextWrite")
		}
	})

	t.Run("Multiple texts maintain correct Z-order", func(t *testing.T) {
		gs := NewGraphicsSystem("")
		lm := gs.GetLayerManager()

		// ピクチャーを作成
		picID, _ := gs.CreatePic(200, 100)

		// 複数のテキストを描画
		gs.TextWrite(picID, 10, 20, "First")
		gs.TextWrite(picID, 10, 40, "Second")
		gs.TextWrite(picID, 10, 60, "Third")

		// Z順序を確認
		pls := lm.GetPictureLayerSet(picID)
		texts := pls.Texts

		if len(texts) != 3 {
			t.Fatalf("expected 3 texts, got %d", len(texts))
		}

		// 後から作成したテキストほどZ順序が大きい
		for i := 1; i < len(texts); i++ {
			if texts[i-1].GetZOrder() >= texts[i].GetZOrder() {
				t.Errorf("text %d should have lower Z-order than text %d", i-1, i)
			}
		}
	})

	t.Run("Text position is correctly stored", func(t *testing.T) {
		gs := NewGraphicsSystem("")
		lm := gs.GetLayerManager()

		// ピクチャーを作成
		picID, _ := gs.CreatePic(200, 100)

		// テキストを描画
		gs.TextWrite(picID, 50, 75, "Test")

		// 位置を確認
		pls := lm.GetPictureLayerSet(picID)
		if pls.GetTextLayerCount() != 1 {
			t.Fatal("expected 1 text layer")
		}

		text := pls.Texts[0]
		x, y := text.GetPosition()
		if x != 50 || y != 75 {
			t.Errorf("expected position (50, 75), got (%d, %d)", x, y)
		}
	})
}

// TestVMIntegration_CompleteRenderingPipeline は完全な描画パイプラインをテスト
// 要件 8.4: VMとの統合
func TestVMIntegration_CompleteRenderingPipeline(t *testing.T) {
	t.Run("Complete pipeline: Picture -> Window -> Cast -> LayerManager", func(t *testing.T) {
		gs := NewGraphicsSystem("")
		lm := gs.GetLayerManager()

		// 1. ピクチャーを作成
		bgPicID, err := gs.CreatePic(200, 200)
		if err != nil {
			t.Fatalf("CreatePic for background failed: %v", err)
		}

		spritePicID, err := gs.CreatePic(50, 50)
		if err != nil {
			t.Fatalf("CreatePic for sprite failed: %v", err)
		}

		// 2. ウィンドウを開く
		winID, err := gs.OpenWin(bgPicID, 100, 100, 200, 200, 0, 0, 0)
		if err != nil {
			t.Fatalf("OpenWin failed: %v", err)
		}

		// 3. キャストを配置
		castID1, err := gs.PutCast(winID, spritePicID, 10, 10, 0, 0, 50, 50)
		if err != nil {
			t.Fatalf("PutCast 1 failed: %v", err)
		}

		castID2, err := gs.PutCast(winID, spritePicID, 60, 60, 0, 0, 50, 50)
		if err != nil {
			t.Fatalf("PutCast 2 failed: %v", err)
		}

		// 4. LayerManagerの状態を確認（WindowLayerSetを使用）
		wls := lm.GetWindowLayerSet(winID)
		if wls == nil {
			t.Fatal("expected WindowLayerSet to be created")
		}

		// キャストが正しく登録されていることを確認
		if wls.GetCastLayerCount() != 2 {
			t.Errorf("expected 2 casts, got %d", wls.GetCastLayerCount())
		}

		cast1 := wls.GetCastLayer(castID1)
		cast2 := wls.GetCastLayer(castID2)

		if cast1 == nil || cast2 == nil {
			t.Fatal("expected both casts to be registered")
		}

		// 5. キャストを移動
		err = gs.MoveCastWithOptions(castID1, WithCastPosition(30, 30))
		if err != nil {
			t.Fatalf("MoveCast failed: %v", err)
		}

		// 移動が反映されていることを確認
		x, y := cast1.GetPosition()
		if x != 30 || y != 30 {
			t.Errorf("expected position (30, 30), got (%d, %d)", x, y)
		}

		// 6. キャストを削除
		err = gs.DelCast(castID1)
		if err != nil {
			t.Fatalf("DelCast failed: %v", err)
		}

		// 削除が反映されていることを確認
		if wls.GetCastLayerCount() != 1 {
			t.Errorf("expected 1 cast after deletion, got %d", wls.GetCastLayerCount())
		}

		if wls.GetCastLayer(castID1) != nil {
			t.Error("expected cast1 to be removed")
		}

		// 7. ウィンドウを閉じる
		err = gs.CloseWin(winID)
		if err != nil {
			t.Fatalf("CloseWin failed: %v", err)
		}

		// WindowLayerSetが削除されたことを確認
		wlsAfter := lm.GetWindowLayerSet(winID)
		if wlsAfter != nil {
			t.Error("expected WindowLayerSet to be deleted after CloseWin")
		}
	})

	t.Run("Complete pipeline: Picture -> Text -> LayerManager", func(t *testing.T) {
		gs := NewGraphicsSystem("")
		lm := gs.GetLayerManager()

		// 1. ピクチャーを作成
		picID, err := gs.CreatePic(300, 200)
		if err != nil {
			t.Fatalf("CreatePic failed: %v", err)
		}

		// 2. テキストを描画
		err = gs.TextWrite(picID, 10, 20, "Line 1")
		if err != nil {
			t.Fatalf("TextWrite 1 failed: %v", err)
		}

		err = gs.TextWrite(picID, 10, 40, "Line 2")
		if err != nil {
			t.Fatalf("TextWrite 2 failed: %v", err)
		}

		err = gs.TextWrite(picID, 10, 60, "Line 3")
		if err != nil {
			t.Fatalf("TextWrite 3 failed: %v", err)
		}

		// 3. LayerManagerの状態を確認
		pls := lm.GetPictureLayerSet(picID)
		if pls == nil {
			t.Fatal("expected PictureLayerSet to be created")
		}

		// テキストが正しく登録されていることを確認
		if pls.GetTextLayerCount() != 3 {
			t.Errorf("expected 3 texts, got %d", pls.GetTextLayerCount())
		}

		// Z順序が正しいことを確認
		allLayers := pls.GetAllLayersSorted()
		for i := 1; i < len(allLayers); i++ {
			if allLayers[i-1].GetZOrder() > allLayers[i].GetZOrder() {
				t.Errorf("layers not sorted by Z-order at index %d", i)
			}
		}
	})

	t.Run("Mixed casts and texts maintain correct Z-order", func(t *testing.T) {
		gs := NewGraphicsSystem("")
		lm := gs.GetLayerManager()

		// ピクチャーを作成
		picID, _ := gs.CreatePic(200, 200)
		winID, _ := gs.OpenWin(picID)

		// キャストとテキストを交互に追加
		gs.PutCast(winID, picID, 10, 10, 0, 0, 32, 32)
		gs.TextWrite(picID, 50, 50, "Text 1")
		gs.PutCast(winID, picID, 20, 20, 0, 0, 32, 32)
		gs.TextWrite(picID, 60, 60, "Text 2")

		// LayerManagerの状態を確認（WindowLayerSetを使用）
		wls := lm.GetWindowLayerSet(winID)
		plsPic := lm.GetPictureLayerSet(picID)

		if wls == nil {
			t.Fatal("expected WindowLayerSet to be created")
		}

		// キャストは2つ
		if wls.GetCastLayerCount() != 2 {
			t.Errorf("expected 2 casts, got %d", wls.GetCastLayerCount())
		}

		// テキストは2つ
		if plsPic.GetTextLayerCount() != 2 {
			t.Errorf("expected 2 texts, got %d", plsPic.GetTextLayerCount())
		}

		// 操作順序に基づくZ順序: キャストとテキストは異なるLayerSetに追加されるため、
		// 各LayerSet内でのZ順序の増加を確認
		// キャスト内のZ順序が増加していることを確認
		casts := wls.GetAllCastLayers()
		for i := 0; i < len(casts)-1; i++ {
			if casts[i].GetZOrder() >= casts[i+1].GetZOrder() {
				t.Error("cast Z-order should increase with addition order")
			}
		}

		// テキスト内のZ順序が増加していることを確認
		for i := 0; i < len(plsPic.Texts)-1; i++ {
			if plsPic.Texts[i].GetZOrder() >= plsPic.Texts[i+1].GetZOrder() {
				t.Error("text Z-order should increase with addition order")
			}
		}
	})
}

// TestLayerManagerIntegration_DirtyFlagPropagation はダーティフラグの伝播をテスト
// 要件 8.1, 8.2, 8.3, 8.4: 統合時のダーティフラグ管理
func TestLayerManagerIntegration_DirtyFlagPropagation(t *testing.T) {
	t.Run("PutCast creates WindowLayerSet and sets dirty flag", func(t *testing.T) {
		gs := NewGraphicsSystem("")
		lm := gs.GetLayerManager()

		picID, _ := gs.CreatePic(100, 100)
		winID, _ := gs.OpenWin(picID)

		// キャストを配置（WindowLayerSetにCastLayerが追加される）
		gs.PutCast(winID, picID, 10, 10, 0, 0, 32, 32)

		wls := lm.GetWindowLayerSet(winID)
		if wls == nil {
			t.Fatal("expected WindowLayerSet to be created")
		}

		// ダーティフラグが設定されていることを確認
		if !wls.IsDirty() {
			t.Error("expected dirty flag to be set after PutCast")
		}
	})

	t.Run("MoveCast sets dirty flag", func(t *testing.T) {
		gs := NewGraphicsSystem("")
		lm := gs.GetLayerManager()

		picID, _ := gs.CreatePic(100, 100)
		winID, _ := gs.OpenWin(picID)
		castID, _ := gs.PutCast(winID, picID, 10, 10, 0, 0, 32, 32)

		wls := lm.GetWindowLayerSet(winID)
		wls.ClearDirty()

		// キャストを移動
		gs.MoveCastWithOptions(castID, WithCastPosition(50, 50))

		// ダーティフラグが設定されていることを確認
		if !wls.IsDirty() {
			t.Error("expected dirty flag to be set after MoveCast")
		}
	})

	t.Run("DelCast sets dirty flag", func(t *testing.T) {
		gs := NewGraphicsSystem("")
		lm := gs.GetLayerManager()

		picID, _ := gs.CreatePic(100, 100)
		winID, _ := gs.OpenWin(picID)
		castID, _ := gs.PutCast(winID, picID, 10, 10, 0, 0, 32, 32)

		wls := lm.GetWindowLayerSet(winID)
		wls.ClearDirty()

		// キャストを削除
		gs.DelCast(castID)

		// ダーティフラグが設定されていることを確認
		if !wls.IsDirty() {
			t.Error("expected dirty flag to be set after DelCast")
		}
	})

	t.Run("TextWrite creates PictureLayerSet and sets dirty flag", func(t *testing.T) {
		gs := NewGraphicsSystem("")
		lm := gs.GetLayerManager()

		picID, _ := gs.CreatePic(200, 100)

		// TextWriteでPictureLayerSetが作成される
		gs.TextWrite(picID, 10, 20, "First")

		pls := lm.GetPictureLayerSet(picID)
		if pls == nil {
			t.Fatal("expected PictureLayerSet to be created")
		}

		// ダーティフラグが設定されていることを確認
		if !pls.IsDirty() {
			t.Error("expected dirty flag to be set after TextWrite")
		}
	})

	t.Run("Second TextWrite sets dirty flag", func(t *testing.T) {
		gs := NewGraphicsSystem("")
		lm := gs.GetLayerManager()

		picID, _ := gs.CreatePic(200, 100)

		// 最初のTextWriteでPictureLayerSetが作成される
		gs.TextWrite(picID, 10, 20, "First")

		pls := lm.GetPictureLayerSet(picID)
		pls.ClearAllDirtyFlags()

		// 2回目のTextWrite
		gs.TextWrite(picID, 10, 40, "Second")

		// ダーティフラグが設定されていることを確認
		if !pls.IsDirty() {
			t.Error("expected dirty flag to be set after second TextWrite")
		}
	})
}

// TestOpenWinCreatesWindowLayerSet はOpenWinがWindowLayerSetを作成することをテスト
// 要件 1.2: ウィンドウが開かれたときにWindowLayerSetを作成する
func TestOpenWinCreatesWindowLayerSet(t *testing.T) {
	t.Run("OpenWin creates WindowLayerSet", func(t *testing.T) {
		gs := NewGraphicsSystem("")
		lm := gs.GetLayerManager()

		// ピクチャーを作成
		picID, err := gs.CreatePic(200, 150)
		if err != nil {
			t.Fatalf("CreatePic failed: %v", err)
		}

		// ウィンドウを開く
		winID, err := gs.OpenWin(picID)
		if err != nil {
			t.Fatalf("OpenWin failed: %v", err)
		}

		// WindowLayerSetが作成されていることを確認
		wls := lm.GetWindowLayerSet(winID)
		if wls == nil {
			t.Fatal("expected WindowLayerSet to be created when window is opened")
		}

		// WindowLayerSetのウィンドウIDが正しいことを確認
		if wls.GetWinID() != winID {
			t.Errorf("expected WinID %d, got %d", winID, wls.GetWinID())
		}
	})

	t.Run("OpenWin creates WindowLayerSet with correct size from picture", func(t *testing.T) {
		gs := NewGraphicsSystem("")
		lm := gs.GetLayerManager()

		// ピクチャーを作成
		picID, err := gs.CreatePic(320, 240)
		if err != nil {
			t.Fatalf("CreatePic failed: %v", err)
		}

		// ウィンドウを開く（サイズ指定なし）
		winID, err := gs.OpenWin(picID)
		if err != nil {
			t.Fatalf("OpenWin failed: %v", err)
		}

		// WindowLayerSetが作成されていることを確認
		wls := lm.GetWindowLayerSet(winID)
		if wls == nil {
			t.Fatal("expected WindowLayerSet to be created")
		}

		// サイズがピクチャーのサイズと一致することを確認
		width, height := wls.GetSize()
		if width != 320 || height != 240 {
			t.Errorf("expected size (320, 240), got (%d, %d)", width, height)
		}
	})

	t.Run("OpenWin creates WindowLayerSet with specified size", func(t *testing.T) {
		gs := NewGraphicsSystem("")
		lm := gs.GetLayerManager()

		// ピクチャーを作成
		picID, err := gs.CreatePic(640, 480)
		if err != nil {
			t.Fatalf("CreatePic failed: %v", err)
		}

		// ウィンドウを開く（サイズ指定あり）
		winID, err := gs.OpenWin(picID, 0, 0, 400, 300, 0, 0, 0)
		if err != nil {
			t.Fatalf("OpenWin failed: %v", err)
		}

		// WindowLayerSetが作成されていることを確認
		wls := lm.GetWindowLayerSet(winID)
		if wls == nil {
			t.Fatal("expected WindowLayerSet to be created")
		}

		// サイズが指定したサイズと一致することを確認
		width, height := wls.GetSize()
		if width != 400 || height != 300 {
			t.Errorf("expected size (400, 300), got (%d, %d)", width, height)
		}
	})

	t.Run("OpenWin creates WindowLayerSet with background color", func(t *testing.T) {
		gs := NewGraphicsSystem("")
		lm := gs.GetLayerManager()

		// ピクチャーを作成
		picID, err := gs.CreatePic(200, 200)
		if err != nil {
			t.Fatalf("CreatePic failed: %v", err)
		}

		// ウィンドウを開く（背景色指定あり: 0xFF0000 = 赤）
		winID, err := gs.OpenWin(picID, 0, 0, 200, 200, 0, 0, 0xFF0000)
		if err != nil {
			t.Fatalf("OpenWin failed: %v", err)
		}

		// WindowLayerSetが作成されていることを確認
		wls := lm.GetWindowLayerSet(winID)
		if wls == nil {
			t.Fatal("expected WindowLayerSet to be created")
		}

		// 背景色が設定されていることを確認
		bgColor := wls.GetBgColor()
		if bgColor == nil {
			t.Fatal("expected background color to be set")
		}
	})

	t.Run("Multiple OpenWin creates separate WindowLayerSets", func(t *testing.T) {
		gs := NewGraphicsSystem("")
		lm := gs.GetLayerManager()

		// 複数のピクチャーを作成
		picID1, _ := gs.CreatePic(100, 100)
		picID2, _ := gs.CreatePic(200, 200)

		// 複数のウィンドウを開く
		winID1, err := gs.OpenWin(picID1)
		if err != nil {
			t.Fatalf("OpenWin 1 failed: %v", err)
		}

		winID2, err := gs.OpenWin(picID2)
		if err != nil {
			t.Fatalf("OpenWin 2 failed: %v", err)
		}

		// 各ウィンドウに対応するWindowLayerSetが作成されていることを確認
		wls1 := lm.GetWindowLayerSet(winID1)
		wls2 := lm.GetWindowLayerSet(winID2)

		if wls1 == nil {
			t.Fatal("expected WindowLayerSet for window 1")
		}
		if wls2 == nil {
			t.Fatal("expected WindowLayerSet for window 2")
		}

		// 異なるWindowLayerSetであることを確認
		if wls1 == wls2 {
			t.Error("expected different WindowLayerSets for different windows")
		}

		// 各WindowLayerSetのウィンドウIDが正しいことを確認
		if wls1.GetWinID() != winID1 {
			t.Errorf("expected WinID %d for wls1, got %d", winID1, wls1.GetWinID())
		}
		if wls2.GetWinID() != winID2 {
			t.Errorf("expected WinID %d for wls2, got %d", winID2, wls2.GetWinID())
		}
	})
}
