package graphics

import (
	"log/slog"
	"os"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestNewGraphicsSystem(t *testing.T) {
	// デフォルトのGraphicsSystemを作成
	gs := NewGraphicsSystem("")

	if gs == nil {
		t.Fatal("NewGraphicsSystem returned nil")
	}

	if gs.virtualWidth != 1024 {
		t.Errorf("Expected virtualWidth 1024, got %d", gs.virtualWidth)
	}

	if gs.virtualHeight != 768 {
		t.Errorf("Expected virtualHeight 768, got %d", gs.virtualHeight)
	}

	if gs.pictures == nil {
		t.Error("PictureManager not initialized")
	}

	if gs.windows == nil {
		t.Error("WindowManager not initialized")
	}

	if gs.casts == nil {
		t.Error("CastManager not initialized")
	}

	if gs.textRenderer == nil {
		t.Error("TextRenderer not initialized")
	}

	if gs.cmdQueue == nil {
		t.Error("CommandQueue not initialized")
	}

	// 要件 8.1: LayerManagerが初期化されていることを確認
	if gs.layerManager == nil {
		t.Error("LayerManager not initialized")
	}
}

func TestNewGraphicsSystemWithOptions(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	gs := NewGraphicsSystem("",
		WithLogger(logger),
		WithVirtualSize(800, 600),
	)

	if gs.virtualWidth != 800 {
		t.Errorf("Expected virtualWidth 800, got %d", gs.virtualWidth)
	}

	if gs.virtualHeight != 600 {
		t.Errorf("Expected virtualHeight 600, got %d", gs.virtualHeight)
	}

	if gs.log != logger {
		t.Error("Logger not set correctly")
	}
}

func TestGraphicsSystemUpdate(t *testing.T) {
	gs := NewGraphicsSystem("")

	// コマンドをキューに追加
	gs.cmdQueue.Push(Command{Type: CmdMovePic, Args: []any{1, 2, 3}})
	gs.cmdQueue.Push(Command{Type: CmdOpenWin, Args: []any{1}})

	if gs.cmdQueue.Len() != 2 {
		t.Errorf("Expected 2 commands in queue, got %d", gs.cmdQueue.Len())
	}

	// Update を呼び出してコマンドを処理
	err := gs.Update()
	if err != nil {
		t.Errorf("Update returned error: %v", err)
	}

	// キューが空になっているはず
	if gs.cmdQueue.Len() != 0 {
		t.Errorf("Expected empty queue after Update, got %d commands", gs.cmdQueue.Len())
	}
}

func TestGraphicsSystemDraw(t *testing.T) {
	gs := NewGraphicsSystem("")

	// ダミーのスクリーンを作成
	screen := ebiten.NewImage(1024, 768)

	// Draw を呼び出す（エラーが発生しないことを確認）
	gs.Draw(screen)
}

func TestGraphicsSystemShutdown(t *testing.T) {
	gs := NewGraphicsSystem("")

	// コマンドをキューに追加
	gs.cmdQueue.Push(Command{Type: CmdMovePic, Args: []any{1, 2, 3}})

	// Shutdown を呼び出す
	gs.Shutdown()

	// キューが空になっているはず
	if gs.cmdQueue.Len() != 0 {
		t.Errorf("Expected empty queue after Shutdown, got %d commands", gs.cmdQueue.Len())
	}
}

// TestGraphicsSystem_GetLayerManager はGetLayerManagerメソッドをテストする
// 要件 8.1: GraphicsSystemにLayerManagerを統合する
func TestGraphicsSystem_GetLayerManager(t *testing.T) {
	gs := NewGraphicsSystem("")

	// GetLayerManagerがnilでないことを確認
	lm := gs.GetLayerManager()
	if lm == nil {
		t.Fatal("GetLayerManager returned nil")
	}

	// LayerManagerが正しく動作することを確認
	// PictureLayerSetを作成
	pls := lm.GetOrCreatePictureLayerSet(1)
	if pls == nil {
		t.Fatal("GetOrCreatePictureLayerSet returned nil")
	}

	if pls.PicID != 1 {
		t.Errorf("Expected PicID 1, got %d", pls.PicID)
	}

	// 同じPictureLayerSetを取得できることを確認
	pls2 := lm.GetPictureLayerSet(1)
	if pls2 != pls {
		t.Error("GetPictureLayerSet returned different instance")
	}

	// LayerManagerのカウントを確認
	if lm.GetPictureLayerSetCount() != 1 {
		t.Errorf("Expected 1 PictureLayerSet, got %d", lm.GetPictureLayerSetCount())
	}
}

// TestGraphicsSystem_CastManagerLayerManagerIntegration はCastManagerとLayerManagerの統合をテストする
// 要件 8.2: CastManagerとLayerManagerを統合する
func TestGraphicsSystem_CastManagerLayerManagerIntegration(t *testing.T) {
	gs := NewGraphicsSystem("")

	// CastManagerがLayerManagerと統合されていることを確認
	lm := gs.GetLayerManager()
	if lm == nil {
		t.Fatal("LayerManager is nil")
	}

	// PutCastを呼び出してCastLayerが作成されることを確認
	// 要件 2.1: PutCastが呼び出されたときに対応するCast_Layerを作成する
	castID, err := gs.PutCast(0, 1, 10, 20, 0, 0, 32, 32)
	if err != nil {
		t.Fatalf("PutCast failed: %v", err)
	}

	// CastLayerが作成されていることを確認
	pls := lm.GetPictureLayerSet(0)
	if pls == nil {
		t.Fatal("PictureLayerSet not created")
	}

	castLayer := pls.GetCastLayer(castID)
	if castLayer == nil {
		t.Fatal("CastLayer not created")
	}

	// CastLayerの位置を確認
	x, y := castLayer.GetPosition()
	if x != 10 || y != 20 {
		t.Errorf("Expected position (10, 20), got (%d, %d)", x, y)
	}

	// MoveCastを呼び出してCastLayerが更新されることを確認
	// 要件 2.2: MoveCastが呼び出されたときに対応するCast_Layerの位置を更新する
	err = gs.MoveCast(castID, 100, 200)
	if err != nil {
		t.Fatalf("MoveCast failed: %v", err)
	}

	x, y = castLayer.GetPosition()
	if x != 100 || y != 200 {
		t.Errorf("Expected position (100, 200), got (%d, %d)", x, y)
	}

	// DelCastを呼び出してCastLayerが削除されることを確認
	// 要件 2.3: DelCastが呼び出されたときに対応するCast_Layerを削除する
	err = gs.DelCast(castID)
	if err != nil {
		t.Fatalf("DelCast failed: %v", err)
	}

	if pls.GetCastLayer(castID) != nil {
		t.Error("CastLayer should be deleted")
	}
}
