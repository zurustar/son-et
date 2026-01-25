package cli

import (
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
