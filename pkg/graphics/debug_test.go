package graphics

import (
	"log/slog"
	"os"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestNewDebugOverlay(t *testing.T) {
	do := NewDebugOverlay()

	if do == nil {
		t.Fatal("NewDebugOverlay returned nil")
	}

	// デフォルトでは無効
	if do.IsEnabled() {
		t.Error("Expected debug overlay to be disabled by default")
	}
}

func TestNewDebugOverlayWithLogger(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	do := NewDebugOverlayWithLogger(logger)

	if do == nil {
		t.Fatal("NewDebugOverlayWithLogger returned nil")
	}

	if do.log != logger {
		t.Error("Logger not set correctly")
	}
}

func TestDebugOverlay_SetEnabled(t *testing.T) {
	do := NewDebugOverlay()

	// 有効にする
	do.SetEnabled(true)
	if !do.IsEnabled() {
		t.Error("Expected debug overlay to be enabled")
	}

	// 無効にする
	do.SetEnabled(false)
	if do.IsEnabled() {
		t.Error("Expected debug overlay to be disabled")
	}
}

func TestDebugOverlay_SetEnabledFromLogLevel(t *testing.T) {
	do := NewDebugOverlay()

	// Debug レベル → 有効
	do.SetEnabledFromLogLevel(slog.LevelDebug)
	if !do.IsEnabled() {
		t.Error("Expected debug overlay to be enabled for Debug level")
	}

	// Info レベル → 無効
	do.SetEnabledFromLogLevel(slog.LevelInfo)
	if do.IsEnabled() {
		t.Error("Expected debug overlay to be disabled for Info level")
	}

	// Warn レベル → 無効
	do.SetEnabledFromLogLevel(slog.LevelWarn)
	if do.IsEnabled() {
		t.Error("Expected debug overlay to be disabled for Warn level")
	}

	// Error レベル → 無効
	do.SetEnabledFromLogLevel(slog.LevelError)
	if do.IsEnabled() {
		t.Error("Expected debug overlay to be disabled for Error level")
	}
}

func TestDebugOverlay_SetEnabledFromLogLevelString(t *testing.T) {
	do := NewDebugOverlay()

	// "debug" → 有効
	do.SetEnabledFromLogLevelString("debug")
	if !do.IsEnabled() {
		t.Error("Expected debug overlay to be enabled for 'debug' level")
	}

	// "info" → 無効
	do.SetEnabledFromLogLevelString("info")
	if do.IsEnabled() {
		t.Error("Expected debug overlay to be disabled for 'info' level")
	}

	// "warn" → 無効
	do.SetEnabledFromLogLevelString("warn")
	if do.IsEnabled() {
		t.Error("Expected debug overlay to be disabled for 'warn' level")
	}

	// "error" → 無効
	do.SetEnabledFromLogLevelString("error")
	if do.IsEnabled() {
		t.Error("Expected debug overlay to be disabled for 'error' level")
	}
}

func TestDebugOverlay_DrawWindowID_Disabled(t *testing.T) {
	do := NewDebugOverlay()
	screen := ebiten.NewImage(100, 100)
	win := &Window{ID: 1}

	// 無効時は描画しない（パニックしないことを確認）
	do.DrawWindowID(screen, win, 0, 0, 100)
}

func TestDebugOverlay_DrawWindowID_Enabled(t *testing.T) {
	do := NewDebugOverlay()
	do.SetEnabled(true)
	screen := ebiten.NewImage(100, 100)
	win := &Window{ID: 1}

	// 有効時は描画する（パニックしないことを確認）
	do.DrawWindowID(screen, win, 0, 0, 100)
}

func TestDebugOverlay_DrawPictureID_Disabled(t *testing.T) {
	do := NewDebugOverlay()
	screen := ebiten.NewImage(100, 100)

	// 無効時は描画しない（パニックしないことを確認）
	do.DrawPictureID(screen, 1, 10, 10)
}

func TestDebugOverlay_DrawPictureID_Enabled(t *testing.T) {
	do := NewDebugOverlay()
	do.SetEnabled(true)
	screen := ebiten.NewImage(100, 100)

	// 有効時は描画する（パニックしないことを確認）
	do.DrawPictureID(screen, 1, 10, 10)
}

func TestDebugOverlay_DrawCastID_Disabled(t *testing.T) {
	do := NewDebugOverlay()
	screen := ebiten.NewImage(100, 100)
	cast := &Cast{ID: 1, PicID: 2}

	// 無効時は描画しない（パニックしないことを確認）
	do.DrawCastID(screen, cast, 10, 10)
}

func TestDebugOverlay_DrawCastID_Enabled(t *testing.T) {
	do := NewDebugOverlay()
	do.SetEnabled(true)
	screen := ebiten.NewImage(100, 100)
	cast := &Cast{ID: 1, PicID: 2}

	// 有効時は描画する（パニックしないことを確認）
	do.DrawCastID(screen, cast, 10, 10)
}

func TestGraphicsSystem_DebugOverlay(t *testing.T) {
	gs := NewGraphicsSystem("")

	// デフォルトでは無効
	if gs.IsDebugOverlayEnabled() {
		t.Error("Expected debug overlay to be disabled by default")
	}

	// 有効にする
	gs.SetDebugOverlayEnabled(true)
	if !gs.IsDebugOverlayEnabled() {
		t.Error("Expected debug overlay to be enabled")
	}

	// 無効にする
	gs.SetDebugOverlayEnabled(false)
	if gs.IsDebugOverlayEnabled() {
		t.Error("Expected debug overlay to be disabled")
	}
}

func TestGraphicsSystem_DebugOverlayFromLogLevel(t *testing.T) {
	gs := NewGraphicsSystem("")

	// Debug レベル → 有効
	gs.SetDebugOverlayFromLogLevel(slog.LevelDebug)
	if !gs.IsDebugOverlayEnabled() {
		t.Error("Expected debug overlay to be enabled for Debug level")
	}

	// Info レベル → 無効
	gs.SetDebugOverlayFromLogLevel(slog.LevelInfo)
	if gs.IsDebugOverlayEnabled() {
		t.Error("Expected debug overlay to be disabled for Info level")
	}
}

func TestGraphicsSystem_DebugOverlayFromLogLevelString(t *testing.T) {
	gs := NewGraphicsSystem("")

	// "debug" → 有効
	gs.SetDebugOverlayFromLogLevelString("debug")
	if !gs.IsDebugOverlayEnabled() {
		t.Error("Expected debug overlay to be enabled for 'debug' level")
	}

	// "info" → 無効
	gs.SetDebugOverlayFromLogLevelString("info")
	if gs.IsDebugOverlayEnabled() {
		t.Error("Expected debug overlay to be disabled for 'info' level")
	}
}

func TestGraphicsSystem_WithDebugOverlayOption(t *testing.T) {
	// WithDebugOverlay(true) オプションでGraphicsSystemを作成
	gs := NewGraphicsSystem("", WithDebugOverlay(true))

	if !gs.IsDebugOverlayEnabled() {
		t.Error("Expected debug overlay to be enabled with WithDebugOverlay(true)")
	}

	// WithDebugOverlay(false) オプションでGraphicsSystemを作成
	gs2 := NewGraphicsSystem("", WithDebugOverlay(false))

	if gs2.IsDebugOverlayEnabled() {
		t.Error("Expected debug overlay to be disabled with WithDebugOverlay(false)")
	}
}
