package cli

import (
	"flag"
	"fmt"
	"os"
	"time"
)

// Config はコマンドライン引数から解析された設定を保持する
type Config struct {
	TitlePath string        // FILLYタイトルのパス（コマンドライン引数から）
	Timeout   time.Duration // タイムアウト時間（0は無制限）
	LogLevel  string        // ログレベル（debug, info, warn, error）
	Headless  bool          // ヘッドレスモード
	ShowHelp  bool          // ヘルプ表示フラグ
}

// ParseArgs コマンドライン引数を解析してConfigを返す
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
		config.TitlePath = fs.Arg(0)
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
  title-path    FILLYタイトルのディレクトリパス（省略可）

Options:
  -t, --timeout <seconds>     指定秒数後にプログラムを終了（デフォルト: 無制限）
  -l, --log-level <level>     ログレベル: debug, info, warn, error（デフォルト: info）
  --headless                  ヘッドレスモード（GUIなし）
  -h, --help                  このヘルプを表示

Examples:
  son-et /path/to/title       外部タイトルを実行
  son-et --timeout 10         10秒後に自動終了
  son-et --headless           ヘッドレスモードで実行
  son-et --log-level debug    デバッグログを有効化
`)
}
