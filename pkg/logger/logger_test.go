package logger

import (
	"log/slog"
	"testing"
)

func TestInitLogger_ValidLevels(t *testing.T) {
	tests := []struct {
		name  string
		level string
	}{
		{"debug", "debug"},
		{"info", "info"},
		{"warn", "warn"},
		{"error", "error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := InitLogger(tt.level)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			logger := GetLogger()
			if logger == nil {
				t.Fatal("GetLogger() returned nil")
			}
		})
	}
}

func TestInitLogger_InvalidLevel(t *testing.T) {
	err := InitLogger("invalid")
	if err == nil {
		t.Error("expected error for invalid log level, got nil")
	}
}

func TestGetLogger_BeforeInit(t *testing.T) {
	// globalLoggerをリセット
	globalLogger = nil

	logger := GetLogger()
	if logger == nil {
		t.Error("GetLogger() should return default logger when not initialized")
	}

	// デフォルトロガーが返されることを確認
	if logger != slog.Default() {
		t.Error("GetLogger() should return slog.Default() when not initialized")
	}
}

func TestGetLogger_AfterInit(t *testing.T) {
	err := InitLogger("info")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	logger := GetLogger()
	if logger == nil {
		t.Error("GetLogger() returned nil after initialization")
	}

	if logger != globalLogger {
		t.Error("GetLogger() should return the initialized logger")
	}
}
