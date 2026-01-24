package engine

import (
	"fmt"
	"sync"
	"time"
)

// DebugLevel represents the logging verbosity level.
type DebugLevel int

const (
	// DebugLevelError logs only errors (level 0)
	DebugLevelError DebugLevel = 0
	// DebugLevelInfo logs errors and info messages (level 1)
	DebugLevelInfo DebugLevel = 1
	// DebugLevelDebug logs everything including debug messages (level 2)
	DebugLevelDebug DebugLevel = 2
)

// Logger provides timestamped logging with debug levels.
type Logger struct {
	level DebugLevel
	mu    sync.Mutex
}

// NewLogger creates a new logger with the specified debug level.
func NewLogger(level DebugLevel) *Logger {
	return &Logger{
		level: level,
	}
}

// SetLevel sets the debug level.
func (l *Logger) SetLevel(level DebugLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// GetLevel returns the current debug level.
func (l *Logger) GetLevel() DebugLevel {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.level
}

// LogError logs an error message (always shown).
func (l *Logger) LogError(format string, args ...interface{}) {
	l.log("ERROR", format, args...)
}

// LogInfo logs an info message (shown at level 1+).
func (l *Logger) LogInfo(format string, args ...interface{}) {
	l.mu.Lock()
	level := l.level
	l.mu.Unlock()

	if level >= DebugLevelInfo {
		l.log("INFO", format, args...)
	}
}

// LogDebug logs a debug message (shown at level 2+).
func (l *Logger) LogDebug(format string, args ...interface{}) {
	l.mu.Lock()
	level := l.level
	l.mu.Unlock()

	if level >= DebugLevelDebug {
		l.log("DEBUG", format, args...)
	}
}

// log formats and prints a log message with timestamp.
func (l *Logger) log(levelStr string, format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := formatTimestamp(time.Now())
	message := fmt.Sprintf(format, args...)
	fmt.Printf("[%s] %s: %s\n", timestamp, levelStr, message)
}

// formatTimestamp formats a time as [HH:MM:SS.mmm]
func formatTimestamp(t time.Time) string {
	hour := t.Hour()
	minute := t.Minute()
	second := t.Second()
	millisecond := t.Nanosecond() / 1000000

	return fmt.Sprintf("%02d:%02d:%02d.%03d", hour, minute, second, millisecond)
}
