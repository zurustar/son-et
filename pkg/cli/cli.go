package cli

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Config はコマンドライン引数から解析された設定を保持する
type Config struct {
	TitlePath string        // FILLYタイトルのパス（ディレクトリ）
	EntryFile string        // エントリーポイントファイル名（TFYファイル指定時）
	Timeout   time.Duration // タイムアウト時間（0は無制限）
	LogLevel  string        // ログレベル（debug, info, warn, error）
	Headless  bool          // ヘッドレスモード
	ShowHelp  bool          // ヘルプ表示フラグ
}

// ParseArgs コマンドライン引数を解析してConfigを返す
// Requirement 12.7: System supports enabling headless mode via command line flag.
// Requirement 12.8: System supports enabling headless mode via environment variable.
// Requirement 13.5: System supports timeout specification via command line flag.
func ParseArgs(args []string) (*Config, error) {
	// 引数を並べ替え：フラグを前に、位置引数を後ろに
	reorderedArgs := reorderArgs(args)

	fs := flag.NewFlagSet("son-et", flag.ContinueOnError)

	config := &Config{}

	var timeoutSec int
	fs.IntVar(&timeoutSec, "timeout", 0, "タイムアウト時間（秒）")
	fs.IntVar(&timeoutSec, "t", 0, "タイムアウト時間（秒）（短縮形）")
	fs.StringVar(&config.LogLevel, "log-level", "info", "ログレベル（debug, info, warn, error）")
	fs.StringVar(&config.LogLevel, "l", "info", "ログレベル（短縮形）")
	fs.BoolVar(&config.Headless, "headless", false, "ヘッドレスモード")
	fs.BoolVar(&config.ShowHelp, "help", false, "ヘルプを表示")
	fs.BoolVar(&config.ShowHelp, "h", false, "ヘルプを表示（短縮形）")

	if err := fs.Parse(reorderedArgs); err != nil {
		return nil, err
	}

	// 環境変数からの設定（コマンドラインフラグが優先）
	// Requirement 12.8: System supports enabling headless mode via environment variable.
	if !config.Headless {
		if headlessEnv := os.Getenv("HEADLESS"); headlessEnv != "" {
			config.Headless = headlessEnv == "1" || strings.ToLower(headlessEnv) == "true"
		}
	}

	// 環境変数からタイムアウトを取得（コマンドラインフラグが優先）
	if timeoutSec == 0 {
		if timeoutEnv := os.Getenv("TIMEOUT"); timeoutEnv != "" {
			if t, err := strconv.Atoi(timeoutEnv); err == nil && t > 0 {
				timeoutSec = t
			}
		}
	}

	// 環境変数からログレベルを取得（コマンドラインフラグが優先）
	if config.LogLevel == "info" {
		if logLevelEnv := os.Getenv("LOG_LEVEL"); logLevelEnv != "" {
			config.LogLevel = strings.ToLower(logLevelEnv)
		}
	}

	// タイムアウトの検証
	if timeoutSec < 0 {
		return nil, fmt.Errorf("timeout must be non-negative, got %d", timeoutSec)
	}
	config.Timeout = time.Duration(timeoutSec) * time.Second

	// ログレベルの検証
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLogLevels[config.LogLevel] {
		return nil, fmt.Errorf("invalid log level: %s (must be debug, info, warn, or error)", config.LogLevel)
	}

	// 位置引数（FILLYタイトルのパス）
	if fs.NArg() > 0 {
		path := fs.Arg(0)

		// TFYファイルが指定された場合、ディレクトリとエントリーファイルに分離
		if strings.HasSuffix(strings.ToLower(path), ".tfy") {
			config.TitlePath = filepath.Dir(path)
			config.EntryFile = filepath.Base(path)
		} else {
			config.TitlePath = path
		}
	}

	return config, nil
}

// reorderArgs 引数を並べ替えて、フラグを前に、位置引数を後ろに配置する
func reorderArgs(args []string) []string {
	var flags []string
	var positional []string

	for i := 0; i < len(args); i++ {
		arg := args[i]

		// フラグかどうかを判定（-または--で始まる）
		if len(arg) > 0 && arg[0] == '-' {
			flags = append(flags, arg)

			// 次の引数が値である可能性をチェック
			// （-t 5 のような場合）
			if i+1 < len(args) && len(args[i+1]) > 0 && args[i+1][0] != '-' {
				// ブール型フラグでない場合は次の引数も追加
				if arg != "-h" && arg != "--help" && arg != "--headless" {
					i++
					flags = append(flags, args[i])
				}
			}
		} else {
			// 位置引数
			positional = append(positional, arg)
		}
	}

	// フラグを前に、位置引数を後ろに配置
	return append(flags, positional...)
}

// PrintHelp ヘルプメッセージを表示
func PrintHelp() {
	fmt.Fprintf(os.Stdout, `son-et - FILLY Interpreter

Usage:
  son-et [options] [title-path]

Arguments:
  title-path    FILLYタイトルのディレクトリパス、またはエントリーTFYファイルのパス（省略可）
                ディレクトリを指定した場合、main関数を含むファイルを自動検出
                TFYファイルを指定した場合、そのファイルをエントリーポイントとして使用

Options:
  -t, --timeout <seconds>     指定秒数後にプログラムを終了（デフォルト: 無制限）
  -l, --log-level <level>     ログレベル: debug, info, warn, error（デフォルト: info）
  --headless                  ヘッドレスモード（GUIなし）
  -h, --help                  このヘルプを表示

Environment Variables:
  HEADLESS=1                  ヘッドレスモードを有効化
  TIMEOUT=<seconds>           タイムアウト時間（秒）
  LOG_LEVEL=<level>           ログレベル

Examples:
  son-et /path/to/title           ディレクトリを指定（main関数を自動検出）
  son-et /path/to/title/MAIN.TFY  エントリーファイルを明示的に指定
  son-et --timeout 10             10秒後に自動終了
  son-et --headless               ヘッドレスモードで実行
  son-et --log-level debug        デバッグログを有効化
  HEADLESS=1 son-et /path/to/title  環境変数でヘッドレスモード
`)
}
