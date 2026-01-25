package app

import (
	"embed"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/zurustar/son-et/pkg/cli"
	"github.com/zurustar/son-et/pkg/compiler"
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
	opcodes  []compiler.OpCode // コンパイル済みOpCode
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

	app.log.Info("Title selected", "name", selectedTitle.Name, "path", selectedTitle.Path, "entryFile", selectedTitle.EntryFile)

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

	// 5. スクリプトのコンパイル
	opcodes, err := app.compileScripts(scripts, selectedTitle)
	if err != nil {
		return fmt.Errorf("failed to compile scripts: %w", err)
	}
	app.opcodes = opcodes

	app.log.Info("Scripts compiled successfully", "opcode_count", len(opcodes))
	app.log.Debug("OpCodes generated", "opcodes", formatOpCodesPreview(opcodes, 10))

	// 6. 仮想デスクトップの実行
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
		// EntryFileが指定されている場合はそれを使用
		if err := app.titleReg.LoadExternalTitleWithEntry(app.config.TitlePath, app.config.EntryFile); err != nil {
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

// compileScripts スクリプトをコンパイルしてOpCodeを生成
// Requirement 13.1: Application calls compiler after loading scripts to generate OpCode.
// Requirement 13.3: When compilation fails, display error message and terminate application.
// Requirement 13.4: When multiple script files exist, identify file containing main function as entry point.
func (app *Application) compileScripts(scripts []script.Script, selectedTitle *title.FillyTitle) ([]compiler.OpCode, error) {
	// エントリーポイントが明示的に指定されている場合
	if selectedTitle.EntryFile != "" {
		app.log.Info("Using explicit entry point with preprocessor", "file", selectedTitle.EntryFile)

		// プリプロセッサを使用してエントリーポイントからコンパイル
		// Requirement 16.1: Preprocessor starts processing from entry point file.
		opcodes, result, err := compiler.CompileWithPreprocessor(selectedTitle.Path, selectedTitle.EntryFile)
		if err != nil {
			app.log.Error("Compilation with preprocessor failed", "file", selectedTitle.EntryFile, "error", err)
			return nil, err
		}

		app.log.Info("Preprocessor completed", "included_files", result.IncludedFiles)
		return opcodes, nil
	}

	// mainエントリーポイントを探してコンパイル
	// Requirement 14.1: System scans all TFY files to identify the file containing main function.
	mainInfo, err := compiler.FindMainScript(scripts)
	if err != nil {
		// Requirement 13.5: When main function is not found, display error message.
		app.log.Error("Failed to find main entry point", "error", err)
		return nil, err
	}

	app.log.Info("Main entry point found, using preprocessor", "file", mainInfo.FileName)

	// プリプロセッサを使用してmainエントリーポイントからコンパイル
	opcodes, result, err := compiler.CompileWithPreprocessor(selectedTitle.Path, mainInfo.FileName)
	if err != nil {
		// Requirement 13.3: When compilation fails, display error message.
		app.log.Error("Compilation with preprocessor failed", "error", err)
		return nil, err
	}

	app.log.Info("Preprocessor completed", "included_files", result.IncludedFiles)
	return opcodes, nil
}

// truncate 文字列を指定した長さで切り詰める
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// formatOpCodesPreview OpCodeのプレビューを生成（デバッグ用）
// Requirement 13.2: When compilation succeeds, output generated OpCode to log (debug level).
func formatOpCodesPreview(opcodes []compiler.OpCode, maxCount int) string {
	if len(opcodes) == 0 {
		return "[]"
	}

	count := len(opcodes)
	if count > maxCount {
		count = maxCount
	}

	var result string
	for i := 0; i < count; i++ {
		if i > 0 {
			result += ", "
		}
		result += fmt.Sprintf("{Cmd: %s}", opcodes[i].Cmd)
	}

	if len(opcodes) > maxCount {
		result += fmt.Sprintf(", ... (%d more)", len(opcodes)-maxCount)
	}

	return "[" + result + "]"
}
