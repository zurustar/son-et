package app

import (
	"embed"
	"fmt"
	"log/slog"

	"github.com/zurustar/son-et/pkg/cli"
	"github.com/zurustar/son-et/pkg/logger"
)

// Application はアプリケーションのメインロジックを管理する
type Application struct {
	config  *cli.Config
	log     *slog.Logger
	embedFS embed.FS
}

// New Applicationを作成
func New(embedFS embed.FS) *Application {
	return &Application{
		embedFS: embedFS,
	}
}

// Run アプリケーションを実行
func (app *Application) Run(args []string) error {
	// 1. コマンドライン引数の解析
	if err := app.parseArgs(args); err != nil {
		return fmt.Errorf("failed to parse args: %w", err)
	}

	if app.config.ShowHelp {
		cli.PrintHelp()
		return nil
	}

	// 2. ロガーの初期化
	if err := app.initLogger(); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	app.log.Info("Application started")

	// TODO: 3. タイトルの読み込みと選択
	// TODO: 4. スクリプトファイルの読み込み
	// TODO: 5. 仮想デスクトップの実行

	app.log.Info("Application terminated normally")
	return nil
}

// parseArgs コマンドライン引数を解析
func (app *Application) parseArgs(args []string) error {
	config, err := cli.ParseArgs(args)
	if err != nil {
		return err
	}
	app.config = config
	return nil
}

// initLogger ロガーを初期化
func (app *Application) initLogger() error {
	if err := logger.InitLogger(app.config.LogLevel); err != nil {
		return err
	}
	app.log = logger.GetLogger()
	return nil
}
