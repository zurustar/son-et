package logger

import (
	"fmt"
	"log/slog"
	"os"
)

var globalLogger *slog.Logger

// InitLogger ログレベルに応じてslogを初期化
func InitLogger(level string) error {
	var slogLevel slog.Level

	switch level {
	case "debug":
		slogLevel = slog.LevelDebug
	case "info":
		slogLevel = slog.LevelInfo
	case "warn":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		return fmt.Errorf("invalid log level: %s", level)
	}

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slogLevel,
	})

	globalLogger = slog.New(handler)
	slog.SetDefault(globalLogger)

	return nil
}

// GetLogger グローバルロガーを取得
func GetLogger() *slog.Logger {
	if globalLogger == nil {
		// デフォルトロガーを返す
		return slog.Default()
	}
	return globalLogger
}
