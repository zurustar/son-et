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

	if gs.virtualWidth != 1280 {
		t.Errorf("Expected virtualWidth 1280, got %d", gs.virtualWidth)
	}

	if gs.virtualHeight != 720 {
		t.Errorf("Expected virtualHeight 720, got %d", gs.virtualHeight)
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
	screen := ebiten.NewImage(1280, 720)

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
