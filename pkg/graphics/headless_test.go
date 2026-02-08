package graphics

import (
	"log/slog"
	"os"
	"testing"
)

func TestNewHeadlessGraphicsSystem(t *testing.T) {
	hgs := NewHeadlessGraphicsSystem()
	if hgs == nil {
		t.Fatal("NewHeadlessGraphicsSystem returned nil")
	}

	// デフォルト値の確認
	if hgs.virtualWidth != 1024 {
		t.Errorf("expected virtualWidth 1024, got %d", hgs.virtualWidth)
	}
	if hgs.virtualHeight != 768 {
		t.Errorf("expected virtualHeight 768, got %d", hgs.virtualHeight)
	}
	if hgs.maxPictures != 256 {
		t.Errorf("expected maxPictures 256, got %d", hgs.maxPictures)
	}
	if hgs.maxWindows != 64 {
		t.Errorf("expected maxWindows 64, got %d", hgs.maxWindows)
	}
	if hgs.maxCasts != 1024 {
		t.Errorf("expected maxCasts 1024, got %d", hgs.maxCasts)
	}
}

func TestHeadlessGraphicsSystem_Options(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	hgs := NewHeadlessGraphicsSystem(
		WithHeadlessLogger(logger),
		WithHeadlessVirtualSize(800, 600),
		WithLogOperations(false),
	)

	if hgs.virtualWidth != 800 {
		t.Errorf("expected virtualWidth 800, got %d", hgs.virtualWidth)
	}
	if hgs.virtualHeight != 600 {
		t.Errorf("expected virtualHeight 600, got %d", hgs.virtualHeight)
	}
	if hgs.logOperations {
		t.Error("expected logOperations to be false")
	}
}

func TestHeadlessGraphicsSystem_PictureManagement(t *testing.T) {
	hgs := NewHeadlessGraphicsSystem(WithLogOperations(false))

	// LoadPic
	picID, err := hgs.LoadPic("test.bmp")
	if err != nil {
		t.Fatalf("LoadPic failed: %v", err)
	}
	if picID != 0 {
		t.Errorf("expected picID 0, got %d", picID)
	}

	// CreatePic
	picID2, err := hgs.CreatePic(100, 200)
	if err != nil {
		t.Fatalf("CreatePic failed: %v", err)
	}
	if picID2 != 1 {
		t.Errorf("expected picID 1, got %d", picID2)
	}

	// PicWidth/PicHeight
	if w := hgs.PicWidth(picID2); w != 100 {
		t.Errorf("expected width 100, got %d", w)
	}
	if h := hgs.PicHeight(picID2); h != 200 {
		t.Errorf("expected height 200, got %d", h)
	}

	// CreatePicFrom
	picID3, err := hgs.CreatePicFrom(picID2)
	if err != nil {
		t.Fatalf("CreatePicFrom failed: %v", err)
	}
	if hgs.PicWidth(picID3) != 100 || hgs.PicHeight(picID3) != 200 {
		t.Error("CreatePicFrom did not copy dimensions correctly")
	}

	// DelPic
	if err := hgs.DelPic(picID); err != nil {
		t.Fatalf("DelPic failed: %v", err)
	}

	// 削除後のアクセス
	if w := hgs.PicWidth(picID); w != 0 {
		t.Errorf("expected width 0 for deleted picture, got %d", w)
	}
}

func TestHeadlessGraphicsSystem_WindowManagement(t *testing.T) {
	hgs := NewHeadlessGraphicsSystem(WithLogOperations(false))

	// ピクチャーを作成
	picID, _ := hgs.CreatePic(640, 480)

	// OpenWin
	winID, err := hgs.OpenWin(picID, 100, 200, 320, 240, 0, 0)
	if err != nil {
		t.Fatalf("OpenWin failed: %v", err)
	}
	if winID != 0 {
		t.Errorf("expected winID 0, got %d", winID)
	}

	// GetPicNo
	gotPicID, err := hgs.GetPicNo(winID)
	if err != nil {
		t.Fatalf("GetPicNo failed: %v", err)
	}
	if gotPicID != picID {
		t.Errorf("expected picID %d, got %d", picID, gotPicID)
	}

	// CapTitle
	if err := hgs.CapTitle(winID, "Test Window"); err != nil {
		t.Fatalf("CapTitle failed: %v", err)
	}

	// MoveWin
	if err := hgs.MoveWin(winID, picID, 50, 60, 400, 300, 10, 20); err != nil {
		t.Fatalf("MoveWin failed: %v", err)
	}

	// CloseWin
	if err := hgs.CloseWin(winID); err != nil {
		t.Fatalf("CloseWin failed: %v", err)
	}

	// 削除後のアクセス
	_, err = hgs.GetPicNo(winID)
	if err == nil {
		t.Error("expected error for deleted window")
	}
}

func TestHeadlessGraphicsSystem_CastManagement(t *testing.T) {
	hgs := NewHeadlessGraphicsSystem(WithLogOperations(false))

	// ピクチャーとウィンドウを作成
	picID, _ := hgs.CreatePic(640, 480)
	_, _ = hgs.OpenWin(picID)

	// PutCast
	// 新API: PutCast(srcPicID, dstPicID, x, y, srcX, srcY, w, h)
	castID, err := hgs.PutCast(picID, picID, 10, 20, 0, 0, 32, 32)
	if err != nil {
		t.Fatalf("PutCast failed: %v", err)
	}
	if castID != 0 {
		t.Errorf("expected castID 0, got %d", castID)
	}

	// MoveCast
	if err := hgs.MoveCast(castID, 50, 60); err != nil {
		t.Fatalf("MoveCast failed: %v", err)
	}

	// DelCast
	if err := hgs.DelCast(castID); err != nil {
		t.Fatalf("DelCast failed: %v", err)
	}

	// 削除後のアクセス
	if err := hgs.MoveCast(castID, 0, 0); err == nil {
		t.Error("expected error for deleted cast")
	}
}

func TestHeadlessGraphicsSystem_CloseWinAll(t *testing.T) {
	hgs := NewHeadlessGraphicsSystem(WithLogOperations(false))

	// 複数のウィンドウとキャストを作成
	picID, _ := hgs.CreatePic(640, 480)
	winID1, _ := hgs.OpenWin(picID)
	winID2, _ := hgs.OpenWin(picID)
	// 新API: PutCast(srcPicID, dstPicID, x, y, srcX, srcY, w, h)
	hgs.PutCast(picID, picID, 0, 0, 0, 0, 32, 32)
	hgs.PutCast(picID, picID, 0, 0, 0, 0, 32, 32)

	// CloseWinAll
	hgs.CloseWinAll()

	// すべてのウィンドウが削除されていることを確認
	_, err := hgs.GetPicNo(winID1)
	if err == nil {
		t.Error("expected error for deleted window 1")
	}
	_, err = hgs.GetPicNo(winID2)
	if err == nil {
		t.Error("expected error for deleted window 2")
	}
}

func TestHeadlessGraphicsSystem_ResourceLimits(t *testing.T) {
	hgs := NewHeadlessGraphicsSystem(WithLogOperations(false))
	hgs.maxPictures = 3 // テスト用に制限を小さくする

	// 制限まで作成
	for i := 0; i < 3; i++ {
		_, err := hgs.CreatePic(100, 100)
		if err != nil {
			t.Fatalf("CreatePic %d failed: %v", i, err)
		}
	}

	// 制限を超えて作成
	_, err := hgs.CreatePic(100, 100)
	if err == nil {
		t.Error("expected error when exceeding resource limit")
	}
}

func TestHeadlessGraphicsSystem_DrawingOperations(t *testing.T) {
	hgs := NewHeadlessGraphicsSystem(WithLogOperations(false))

	picID, _ := hgs.CreatePic(640, 480)

	// 描画操作（エラーが発生しないことを確認）
	if err := hgs.DrawLine(picID, 0, 0, 100, 100); err != nil {
		t.Errorf("DrawLine failed: %v", err)
	}
	if err := hgs.DrawRect(picID, 10, 10, 50, 50, 0); err != nil {
		t.Errorf("DrawRect failed: %v", err)
	}
	if err := hgs.FillRect(picID, 10, 10, 50, 50, 0xFF0000); err != nil {
		t.Errorf("FillRect failed: %v", err)
	}
	if err := hgs.DrawCircle(picID, 100, 100, 50, 0); err != nil {
		t.Errorf("DrawCircle failed: %v", err)
	}

	hgs.SetLineSize(2)
	if err := hgs.SetPaintColor(0x00FF00); err != nil {
		t.Errorf("SetPaintColor failed: %v", err)
	}

	_, err := hgs.GetColor(picID, 50, 50)
	if err != nil {
		t.Errorf("GetColor failed: %v", err)
	}
}

func TestHeadlessGraphicsSystem_TextOperations(t *testing.T) {
	hgs := NewHeadlessGraphicsSystem(WithLogOperations(false))

	picID, _ := hgs.CreatePic(640, 480)

	// テキスト操作（エラーが発生しないことを確認）
	if err := hgs.SetFont("Arial", 12); err != nil {
		t.Errorf("SetFont failed: %v", err)
	}
	if err := hgs.SetTextColor(0xFFFFFF); err != nil {
		t.Errorf("SetTextColor failed: %v", err)
	}
	if err := hgs.SetBgColor(0x000000); err != nil {
		t.Errorf("SetBgColor failed: %v", err)
	}
	if err := hgs.SetBackMode(1); err != nil {
		t.Errorf("SetBackMode failed: %v", err)
	}
	if err := hgs.TextWrite(picID, 10, 10, "Hello, World!"); err != nil {
		t.Errorf("TextWrite failed: %v", err)
	}
}

func TestHeadlessGraphicsSystem_PictureTransfer(t *testing.T) {
	hgs := NewHeadlessGraphicsSystem(WithLogOperations(false))

	srcPic, _ := hgs.CreatePic(640, 480)
	dstPic, _ := hgs.CreatePic(640, 480)

	// 転送操作（エラーが発生しないことを確認）
	if err := hgs.MovePic(srcPic, 0, 0, 100, 100, dstPic, 0, 0, 0); err != nil {
		t.Errorf("MovePic failed: %v", err)
	}
	if err := hgs.MovePicWithSpeed(srcPic, 0, 0, 100, 100, dstPic, 0, 0, 2, 50); err != nil {
		t.Errorf("MovePicWithSpeed failed: %v", err)
	}
	if err := hgs.MoveSPic(srcPic, 0, 0, 100, 100, dstPic, 0, 0, 200, 200); err != nil {
		t.Errorf("MoveSPic failed: %v", err)
	}
	if err := hgs.TransPic(srcPic, 0, 0, 100, 100, dstPic, 0, 0, 0x000000); err != nil {
		t.Errorf("TransPic failed: %v", err)
	}
	if err := hgs.ReversePic(srcPic, 0, 0, 100, 100, dstPic, 0, 0); err != nil {
		t.Errorf("ReversePic failed: %v", err)
	}
}

func TestHeadlessGraphicsSystem_VirtualDesktop(t *testing.T) {
	hgs := NewHeadlessGraphicsSystem(
		WithHeadlessVirtualSize(1920, 1080),
	)

	if w := hgs.GetVirtualWidth(); w != 1920 {
		t.Errorf("expected virtualWidth 1920, got %d", w)
	}
	if h := hgs.GetVirtualHeight(); h != 1080 {
		t.Errorf("expected virtualHeight 1080, got %d", h)
	}
}

func TestHeadlessGraphicsSystem_OperationHistory(t *testing.T) {
	hgs := NewHeadlessGraphicsSystem(
		WithLogOperations(false),
		WithRecordHistory(true),
	)

	// 操作を実行
	hgs.CreatePic(100, 100)
	hgs.DrawLine(0, 0, 0, 100, 100)
	hgs.TextWrite(0, 10, 10, "Hello")

	// 履歴を確認
	history := hgs.GetOperationHistory()
	if len(history) != 3 {
		t.Errorf("expected 3 operations, got %d", len(history))
	}

	// 操作名を確認
	expectedOps := []string{"CreatePic", "DrawLine", "TextWrite"}
	for i, expected := range expectedOps {
		if history[i].Operation != expected {
			t.Errorf("operation %d: expected %s, got %s", i, expected, history[i].Operation)
		}
	}

	// 引数を確認
	if history[0].Args["width"] != 100 {
		t.Errorf("expected width 100, got %v", history[0].Args["width"])
	}

	// 件数を確認
	if count := hgs.GetOperationCount(); count != 3 {
		t.Errorf("expected count 3, got %d", count)
	}

	// 履歴をクリア
	hgs.ClearOperationHistory()
	if count := hgs.GetOperationCount(); count != 0 {
		t.Errorf("expected count 0 after clear, got %d", count)
	}
}

func TestHeadlessGraphicsSystem_OperationHistoryDisabled(t *testing.T) {
	hgs := NewHeadlessGraphicsSystem(
		WithLogOperations(false),
		WithRecordHistory(false), // 履歴記録を無効
	)

	// 操作を実行
	hgs.CreatePic(100, 100)
	hgs.DrawLine(0, 0, 0, 100, 100)

	// 履歴は空であることを確認
	history := hgs.GetOperationHistory()
	if len(history) != 0 {
		t.Errorf("expected 0 operations when history disabled, got %d", len(history))
	}
}
