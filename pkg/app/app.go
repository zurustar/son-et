package app

import (
	"embed"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/zurustar/son-et/pkg/cli"
	"github.com/zurustar/son-et/pkg/logger"
	"github.com/zurustar/son-et/pkg/script"
	"github.com/zurustar/son-et/pkg/title"
	"github.com/zurustar/son-et/pkg/window"
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

	// 5. 仮想デスクトップの実行
	if err := app.runDesktop(); err != nil {
		return fmt.Errorf("failed to run desktop: %w", err)
	}

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
	if len(titles) == 0 {
		return nil, fmt.Errorf("no titles available for selection")
	}

	app.log.Info("Multiple titles available, showing selection screen", "count", len(titles))

	// ヘッドレスモードの場合は標準入出力で選択
	if app.config.Headless {
		return window.RunHeadless(titles, app.config.Timeout, os.Stdin, os.Stdout)
	}

	// GUIモードの場合はウィンドウで選択
	return window.Run(window.ModeSelection, titles, app.config.Timeout)
}

// runDesktop 仮想デスクトップを実行
func (app *Application) runDesktop() error {
	app.log.Info("Starting virtual desktop")

	// ヘッドレスモードの場合
	if app.config.Headless {
		app.log.Info("Headless mode: skipping desktop display")

		// タイムアウトが指定されている場合は、その時間だけ待機
		if app.config.Timeout > 0 {
			app.log.Info("Waiting for timeout", "duration", app.config.Timeout)
			time.Sleep(app.config.Timeout)
			app.log.Info("Timeout reached, terminating")
		}

		return nil
	}

	// GUIモードの場合は仮想デスクトップを表示
	_, err := window.Run(window.ModeDesktop, nil, app.config.Timeout)
	return err
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
