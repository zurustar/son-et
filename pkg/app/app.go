package app

import (
	"embed"
	"fmt"
	"log/slog"

	"github.com/zurustar/son-et/pkg/cli"
	"github.com/zurustar/son-et/pkg/logger"
	"github.com/zurustar/son-et/pkg/script"
	"github.com/zurustar/son-et/pkg/title"
)

// Application はアプリケーションのメインロジックを管理する
type Application struct {
	config   *cli.Config
	log      *slog.Logger
	titleReg *title.FillyTitleRegistry
	embedFS  embed.FS
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

	// 3. タイトルの読み込みと選択
	selectedTitle, err := app.loadTitle()
	if err != nil {
		return fmt.Errorf("failed to load title: %w", err)
	}

	app.log.Info("Title selected", "name", selectedTitle.Name, "path", selectedTitle.Path)

	// 4. スクリプトファイルの読み込み
	scripts, err := app.loadScripts(selectedTitle.Path)
	if err != nil {
		return fmt.Errorf("failed to load scripts: %w", err)
	}

	app.log.Info("Scripts loaded", "count", len(scripts))
	for _, s := range scripts {
		app.log.Info("Script file", "name", s.FileName, "size", s.Size)
		app.log.Debug("Script content preview", "name", s.FileName, "preview", truncate(s.Content, 100))
	}

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

// loadTitle タイトルを読み込む
func (app *Application) loadTitle() (*title.FillyTitle, error) {
	app.titleReg = title.NewFillyTitleRegistry(app.embedFS)

	// 外部タイトルの読み込み（指定されている場合）
	if app.config.TitlePath != "" {
		if err := app.titleReg.LoadExternalTitle(app.config.TitlePath); err != nil {
			return nil, fmt.Errorf("failed to load external title: %w", err)
		}
	}

	// タイトルの選択
	selectedTitle, needsSelection, err := app.titleReg.SelectTitle()
	if err != nil {
		return nil, fmt.Errorf("failed to select title: %w", err)
	}

	// タイトル選択画面の表示（必要な場合）
	if needsSelection {
		selectedTitle, err = app.selectTitle(app.titleReg.GetAvailableTitles())
		if err != nil {
			return nil, fmt.Errorf("failed to select title from menu: %w", err)
		}
	}

	return selectedTitle, nil
}

// selectTitle タイトルを選択（選択画面が必要な場合は表示）
func (app *Application) selectTitle(titles []title.FillyTitle) (*title.FillyTitle, error) {
	// TODO: 選択画面の実装（現在は仮実装：最初のタイトルを返す）
	if len(titles) == 0 {
		return nil, fmt.Errorf("no titles available for selection")
	}

	app.log.Info("Multiple titles available, selecting first one (temporary implementation)", "count", len(titles))
	return &titles[0], nil
}

// loadScripts スクリプトファイルを読み込む
func (app *Application) loadScripts(titlePath string) ([]script.Script, error) {
	loader := script.NewLoader(titlePath)
	scripts, err := loader.LoadAllScripts()
	if err != nil {
		return nil, err
	}
	return scripts, nil
}

// truncate 文字列を指定した長さで切り詰める
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
