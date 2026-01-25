package cli

import (
	"os"
	"testing"
	"time"
)

func TestParseArgs_ValidArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected Config
	}{
		{
			name: "デフォルト設定",
			args: []string{},
			expected: Config{
				TitlePath: "",
				Timeout:   0,
				LogLevel:  "info",
				Headless:  false,
				ShowHelp:  false,
			},
		},
		{
			name: "タイトルパス指定",
			args: []string{"/path/to/title"},
			expected: Config{
				TitlePath: "/path/to/title",
				Timeout:   0,
				LogLevel:  "info",
				Headless:  false,
				ShowHelp:  false,
			},
		},
		{
			name: "タイムアウト指定",
			args: []string{"--timeout", "10"},
			expected: Config{
				TitlePath: "",
				Timeout:   10 * time.Second,
				LogLevel:  "info",
				Headless:  false,
				ShowHelp:  false,
			},
		},
		{
			name: "タイムアウト指定（短縮形）",
			args: []string{"-t", "5"},
			expected: Config{
				TitlePath: "",
				Timeout:   5 * time.Second,
				LogLevel:  "info",
				Headless:  false,
				ShowHelp:  false,
			},
		},
		{
			name: "ログレベル指定",
			args: []string{"--log-level", "debug"},
			expected: Config{
				TitlePath: "",
				Timeout:   0,
				LogLevel:  "debug",
				Headless:  false,
				ShowHelp:  false,
			},
		},
		{
			name: "ログレベル指定（短縮形）",
			args: []string{"-l", "error"},
			expected: Config{
				TitlePath: "",
				Timeout:   0,
				LogLevel:  "error",
				Headless:  false,
				ShowHelp:  false,
			},
		},
		{
			name: "ヘッドレスモード",
			args: []string{"--headless"},
			expected: Config{
				TitlePath: "",
				Timeout:   0,
				LogLevel:  "info",
				Headless:  true,
				ShowHelp:  false,
			},
		},
		{
			name: "ヘルプ表示",
			args: []string{"--help"},
			expected: Config{
				TitlePath: "",
				Timeout:   0,
				LogLevel:  "info",
				Headless:  false,
				ShowHelp:  true,
			},
		},
		{
			name: "ヘルプ表示（短縮形）",
			args: []string{"-h"},
			expected: Config{
				TitlePath: "",
				Timeout:   0,
				LogLevel:  "info",
				Headless:  false,
				ShowHelp:  true,
			},
		},
		{
			name: "複数オプション",
			args: []string{"--timeout", "30", "--log-level", "warn", "--headless", "/path/to/title"},
			expected: Config{
				TitlePath: "/path/to/title",
				Timeout:   30 * time.Second,
				LogLevel:  "warn",
				Headless:  true,
				ShowHelp:  false,
			},
		},
		{
			name: "位置引数の後にフラグ（順序に関係なく動作）",
			args: []string{"-log-level", "debug", "./samples/kuma2", "--timeout", "5"},
			expected: Config{
				TitlePath: "./samples/kuma2",
				Timeout:   5 * time.Second,
				LogLevel:  "debug",
				Headless:  false,
				ShowHelp:  false,
			},
		},
		{
			name: "位置引数が最初（順序に関係なく動作）",
			args: []string{"/path/to/title", "--timeout", "10", "--headless"},
			expected: Config{
				TitlePath: "/path/to/title",
				Timeout:   10 * time.Second,
				LogLevel:  "info",
				Headless:  true,
				ShowHelp:  false,
			},
		},
		{
			name: "TFYファイルパス指定",
			args: []string{"/path/to/title/MAIN.TFY"},
			expected: Config{
				TitlePath: "/path/to/title",
				EntryFile: "MAIN.TFY",
				Timeout:   0,
				LogLevel:  "info",
				Headless:  false,
				ShowHelp:  false,
			},
		},
		{
			name: "TFYファイルパス指定（小文字）",
			args: []string{"samples/kuma2/kuma2.tfy"},
			expected: Config{
				TitlePath: "samples/kuma2",
				EntryFile: "kuma2.tfy",
				Timeout:   0,
				LogLevel:  "info",
				Headless:  false,
				ShowHelp:  false,
			},
		},
		{
			name: "TFYファイルパス指定とオプション",
			args: []string{"--headless", "samples/sab2/TOKYO.TFY", "--timeout", "5"},
			expected: Config{
				TitlePath: "samples/sab2",
				EntryFile: "TOKYO.TFY",
				Timeout:   5 * time.Second,
				LogLevel:  "info",
				Headless:  true,
				ShowHelp:  false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ParseArgs(tt.args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if config.TitlePath != tt.expected.TitlePath {
				t.Errorf("TitlePath = %q, want %q", config.TitlePath, tt.expected.TitlePath)
			}
			if config.EntryFile != tt.expected.EntryFile {
				t.Errorf("EntryFile = %q, want %q", config.EntryFile, tt.expected.EntryFile)
			}
			if config.Timeout != tt.expected.Timeout {
				t.Errorf("Timeout = %v, want %v", config.Timeout, tt.expected.Timeout)
			}
			if config.LogLevel != tt.expected.LogLevel {
				t.Errorf("LogLevel = %q, want %q", config.LogLevel, tt.expected.LogLevel)
			}
			if config.Headless != tt.expected.Headless {
				t.Errorf("Headless = %v, want %v", config.Headless, tt.expected.Headless)
			}
			if config.ShowHelp != tt.expected.ShowHelp {
				t.Errorf("ShowHelp = %v, want %v", config.ShowHelp, tt.expected.ShowHelp)
			}
		})
	}
}

func TestParseArgs_InvalidArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "負のタイムアウト",
			args: []string{"--timeout", "-10"},
		},
		{
			name: "無効なログレベル",
			args: []string{"--log-level", "invalid"},
		},
		{
			name: "無効なログレベル（短縮形）",
			args: []string{"-l", "trace"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseArgs(tt.args)
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestParseArgs_EnvironmentVariables(t *testing.T) {
	// Save original environment variables
	origHeadless := os.Getenv("HEADLESS")
	origTimeout := os.Getenv("TIMEOUT")
	origLogLevel := os.Getenv("LOG_LEVEL")

	// Restore environment variables after test
	defer func() {
		os.Setenv("HEADLESS", origHeadless)
		os.Setenv("TIMEOUT", origTimeout)
		os.Setenv("LOG_LEVEL", origLogLevel)
	}()

	tests := []struct {
		name     string
		args     []string
		envVars  map[string]string
		expected Config
	}{
		{
			name: "HEADLESS=1 enables headless mode",
			args: []string{},
			envVars: map[string]string{
				"HEADLESS": "1",
			},
			expected: Config{
				Headless: true,
				LogLevel: "info",
			},
		},
		{
			name: "HEADLESS=true enables headless mode",
			args: []string{},
			envVars: map[string]string{
				"HEADLESS": "true",
			},
			expected: Config{
				Headless: true,
				LogLevel: "info",
			},
		},
		{
			name: "HEADLESS=TRUE enables headless mode (case insensitive)",
			args: []string{},
			envVars: map[string]string{
				"HEADLESS": "TRUE",
			},
			expected: Config{
				Headless: true,
				LogLevel: "info",
			},
		},
		{
			name: "TIMEOUT sets timeout",
			args: []string{},
			envVars: map[string]string{
				"TIMEOUT": "30",
			},
			expected: Config{
				Timeout:  30 * time.Second,
				LogLevel: "info",
			},
		},
		{
			name: "LOG_LEVEL sets log level",
			args: []string{},
			envVars: map[string]string{
				"LOG_LEVEL": "debug",
			},
			expected: Config{
				LogLevel: "debug",
			},
		},
		{
			name: "command line flag overrides HEADLESS env var",
			args: []string{"--headless"},
			envVars: map[string]string{
				"HEADLESS": "0",
			},
			expected: Config{
				Headless: true,
				LogLevel: "info",
			},
		},
		{
			name: "command line flag overrides TIMEOUT env var",
			args: []string{"--timeout", "10"},
			envVars: map[string]string{
				"TIMEOUT": "30",
			},
			expected: Config{
				Timeout:  10 * time.Second,
				LogLevel: "info",
			},
		},
		{
			name: "command line flag overrides LOG_LEVEL env var",
			args: []string{"--log-level", "error"},
			envVars: map[string]string{
				"LOG_LEVEL": "debug",
			},
			expected: Config{
				LogLevel: "error",
			},
		},
		{
			name: "multiple env vars",
			args: []string{},
			envVars: map[string]string{
				"HEADLESS":  "1",
				"TIMEOUT":   "60",
				"LOG_LEVEL": "warn",
			},
			expected: Config{
				Headless: true,
				Timeout:  60 * time.Second,
				LogLevel: "warn",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment variables
			os.Unsetenv("HEADLESS")
			os.Unsetenv("TIMEOUT")
			os.Unsetenv("LOG_LEVEL")

			// Set test environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			config, err := ParseArgs(tt.args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if config.Headless != tt.expected.Headless {
				t.Errorf("Headless = %v, want %v", config.Headless, tt.expected.Headless)
			}
			if config.Timeout != tt.expected.Timeout {
				t.Errorf("Timeout = %v, want %v", config.Timeout, tt.expected.Timeout)
			}
			if config.LogLevel != tt.expected.LogLevel {
				t.Errorf("LogLevel = %q, want %q", config.LogLevel, tt.expected.LogLevel)
			}
		})
	}
}
