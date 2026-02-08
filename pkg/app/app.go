package app

import (
	"embed"
	"fmt"
	"log/slog"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	ebitenAudio "github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/zurustar/son-et/pkg/cli"
	"github.com/zurustar/son-et/pkg/compiler"
	"github.com/zurustar/son-et/pkg/fileutil"
	"github.com/zurustar/son-et/pkg/graphics"
	"github.com/zurustar/son-et/pkg/logger"
	"github.com/zurustar/son-et/pkg/script"
	"github.com/zurustar/son-et/pkg/title"
	"github.com/zurustar/son-et/pkg/vm"
	"github.com/zurustar/son-et/pkg/vm/audio"
	"github.com/zurustar/son-et/pkg/window"
)

// Application はアプリケーションのメインロジックを管理する
type Application struct {
	config        *cli.Config
	log           *slog.Logger
	titleReg      *title.FillyTitleRegistry
	embedFS       embed.FS
	opcodes       []compiler.OpCode // コンパイル済みOpCode
	selectedTitle *title.FillyTitle // 選択されたタイトル
	soundFontPath string            // SoundFontファイルのパス（後方互換性のため保持）

	// soundFontLocation はSoundFontファイルの場所情報
	// 埋め込みファイルと外部ファイルの両方に対応
	soundFontLocation *SoundFontLocation

	// sharedAudioCtx はEbitengineのオーディオコンテキスト（一度だけ作成可能）
	// タイトル切り替え時に再利用する
	sharedAudioCtx *ebitenAudio.Context
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

	// タイトル選択画面からデスクトップモードに遷移した場合は、
	// runWithSelection内で全て処理されているので終了
	if selectedTitle == nil {
		// ユーザーがキャンセルした場合
		app.log.Info("No title selected, exiting")
		return nil
	}

	// runWithSelectionで処理された場合（opcodesが既に設定されている）
	if app.opcodes != nil {
		app.log.Info("Application terminated normally (via selection mode)")
		return nil
	}

	app.log.Info("Title selected", "name", selectedTitle.Name, "path", selectedTitle.Path, "entryFile", selectedTitle.EntryFile)
	app.selectedTitle = selectedTitle

	// 4. スクリプトファイルの読み込み
	scripts, err := app.loadScripts(selectedTitle)
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

	// GUIモードの場合は選択画面を表示し、選択後にデスクトップモードに遷移
	// 単一のRunGame呼び出しで両方を処理する
	return app.runWithSelection(titles)
}

// runDesktop 仮想デスクトップを実行
// Requirement 13.1: Application integrates VM after compilation.
// Requirement 13.2: Application passes compiled OpCode to VM.
// Requirement 13.3: Application starts VM execution.
func (app *Application) runDesktop() error {
	app.log.Info("Starting virtual desktop")

	// ヘッドレスモードの場合はVMを実行
	if app.config.Headless {
		app.log.Info("Headless mode: running VM without GUI")
		return app.runVM()
	}

	// GUIモードの場合はEbitengineのゲームループでVMとGraphicsSystemを統合
	app.log.Info("GUI mode: running VM with Ebitengine")

	// VMオプションを設定
	opts := []vm.Option{
		vm.WithHeadless(false),
		vm.WithLogger(app.log),
		vm.WithTitlePath(app.selectedTitle.Path),
	}

	// タイムアウトが指定されている場合
	if app.config.Timeout > 0 {
		opts = append(opts, vm.WithTimeout(app.config.Timeout))
	}

	// SoundFontパスを設定（埋め込みファイルと外部ファイルの両方に対応）
	// Requirement 3.1, 3.2, 3.3: 優先順位に従ってSF2ファイルを検索
	if app.soundFontLocation == nil {
		app.soundFontLocation = findSoundFont(app.embedFS, app.selectedTitle.Path, app.selectedTitle.IsEmbedded)
	}

	if app.soundFontLocation != nil {
		app.soundFontPath = app.soundFontLocation.Path
		opts = append(opts, vm.WithSoundFont(app.soundFontPath))
		app.log.Info("SoundFont configured", "path", app.soundFontPath, "embedded", app.soundFontLocation.IsEmbedded)
	}

	// VMを作成
	vmInstance := vm.New(app.opcodes, opts...)

	// オーディオシステムを初期化
	// Requirement 2.1: FileSystemインターフェースを使用してSF2ファイルを読み込む
	if app.soundFontLocation != nil {
		var audioSys *audio.AudioSystem
		var err error

		// SoundFontのFileSystemを使用してオーディオシステムを初期化
		audioSys, err = audio.NewAudioSystemWithFS(
			app.soundFontLocation.Path,
			vmInstance.GetEventQueue(),
			nil, // audioCtx - 新規作成
			app.soundFontLocation.FileSystem,
		)
		if err != nil {
			app.log.Warn("Failed to initialize audio system", "error", err)
		} else {
			// 埋め込みタイトルの場合はMIDI/WAV用のFileSystemを設定
			if app.selectedTitle.IsEmbedded {
				embedFS := fileutil.NewEmbedFS(app.embedFS, app.selectedTitle.Path)
				audioSys.SetFileSystem(embedFS)
				app.log.Info("Audio system using embedded file system for MIDI/WAV", "basePath", app.selectedTitle.Path)
			}
			vmInstance.SetAudioSystem(audioSys)
			app.log.Info("Audio system initialized")

			defer func() {
				vmInstance.ShutdownAudio()
				app.log.Info("Audio system shut down")
			}()
		}
	}

	// グラフィックスシステムを初期化
	graphicsSys := graphics.NewGraphicsSystem(
		app.selectedTitle.Path,
		graphics.WithLogger(app.log),
	)
	// 埋め込みタイトルの場合はembed.FSを設定
	if app.selectedTitle.IsEmbedded {
		graphicsSys.SetEmbedFS(app.embedFS)
	}
	vmInstance.SetGraphicsSystem(graphicsSys)
	app.log.Info("Graphics system initialized")

	// ログレベルに基づいてデバッグオーバーレイを有効化
	graphicsSys.SetDebugOverlayFromLogLevelString(app.config.LogLevel)

	defer func() {
		graphicsSys.Shutdown()
		app.log.Info("Graphics system shut down")
	}()

	// Ebitengineのゲームを作成
	game := window.NewGame(window.ModeDesktop, nil, app.config.Timeout)

	// 単一タイトル実行時はタイトル選択画面がないことを明示的に設定
	// Requirements 3.1, 3.2: 単一タイトル実行中にESCキーを押すとプログラムが終了する
	game.SetHasTitleSelection(false)

	game.SetGraphicsSystem(graphicsSys)
	game.SetVMRunner(vmInstance)
	game.SetEventPusher(vmInstance) // マウスイベントをVMに伝達

	// VMを開始する関数を設定（Ebitengine初期化後に呼び出される）
	vmErrCh := make(chan error, 1)
	game.SetVMStartFunc(func() {
		app.log.Info("Starting VM execution in background (after Ebitengine init)")
		go func() {
			vmErrCh <- vmInstance.Run()
		}()
	}, vmErrCh)

	// Ebitengineのゲームループを実行
	app.log.Info("Starting Ebitengine game loop")
	// skelton要件 3.2: ウィンドウサイズは 1024x768 ピクセル
	ebiten.SetWindowSize(1024, 768)
	ebiten.SetWindowTitle("son-et - FILLY interpreter")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeDisabled)

	if err := ebiten.RunGame(game); err != nil {
		app.log.Error("Ebitengine game loop failed", "error", err)
		vmInstance.Stop()
		return fmt.Errorf("game loop failed: %w", err)
	}

	// VMの終了を待つ
	select {
	case vmErr := <-vmErrCh:
		if vmErr != nil {
			app.log.Error("VM execution failed", "error", vmErr)
			return vmErr
		}
	default:
		// VMがまだ実行中の場合は停止
		vmInstance.Stop()
	}

	app.log.Info("Desktop execution completed")
	return nil
}

// runWithSelection タイトル選択画面からデスクトップモードまでを単一のRunGameで実行
// Ebitengineは一度RunGameが終了すると再利用できないため、
// タイトル選択とデスクトップ実行を同じRunGame内で行う必要がある
func (app *Application) runWithSelection(titles []title.FillyTitle) (*title.FillyTitle, error) {
	// Gameを選択モードで作成
	game := window.NewGame(window.ModeSelection, titles, app.config.Timeout)

	// 複数タイトル環境であることを設定
	// Requirements 2.1, 3.1, 5.1: タイトル選択画面があることを示す
	game.SetHasTitleSelection(true)

	// タイトル選択時のコールバックを設定
	// このコールバック内でVM/GraphicsSystemをセットアップし、デスクトップモードに遷移する
	var vmInstance *vm.VM
	var graphicsSys *graphics.GraphicsSystem
	var audioSys *audio.AudioSystem
	vmErrCh := make(chan error, 1)

	// タイトル終了時のリソースクリーンアップコールバックを設定
	// Requirements 2.2, 2.3, 2.4, 4.1, 4.2, 4.3: リソースのクリーンアップ
	game.SetOnTitleExit(func() error {
		app.log.Info("Cleaning up resources for title exit")

		// VM停止 (Requirement 4.1: VMのすべてのゴルーチンを停止)
		if vmInstance != nil {
			vmInstance.Stop()
			app.log.Info("VM stopped")

			// AudioSystem停止 (Requirement 4.3: すべての再生中の音声を停止)
			// AudioSystemはVMを通じてシャットダウンする
			if audioSys != nil {
				vmInstance.ShutdownAudio()
				app.log.Info("Audio system shut down")
			}
		}

		// GraphicsSystem停止 (Requirement 4.2: すべてのスプライトとテクスチャを解放)
		if graphicsSys != nil {
			graphicsSys.Shutdown()
			app.log.Info("Graphics system shut down")
		}

		// リソース参照をクリア
		vmInstance = nil
		graphicsSys = nil
		audioSys = nil

		return nil
	})

	game.SetOnTitleSelected(func(selectedTitle *title.FillyTitle) error {
		app.log.Info("Title selected, setting up VM and graphics", "name", selectedTitle.Name)
		app.selectedTitle = selectedTitle

		// スクリプトの読み込みとコンパイル
		scripts, err := app.loadScripts(selectedTitle)
		if err != nil {
			return fmt.Errorf("failed to load scripts: %w", err)
		}
		app.log.Info("Scripts loaded", "count", len(scripts))

		opcodes, err := app.compileScripts(scripts, selectedTitle)
		if err != nil {
			return fmt.Errorf("failed to compile scripts: %w", err)
		}
		app.opcodes = opcodes
		app.log.Info("Scripts compiled", "opcode_count", len(opcodes))

		// VMオプションを設定
		opts := []vm.Option{
			vm.WithHeadless(false),
			vm.WithLogger(app.log),
			vm.WithTitlePath(selectedTitle.Path),
		}

		if app.config.Timeout > 0 {
			opts = append(opts, vm.WithTimeout(app.config.Timeout))
		}

		// SoundFontパスを設定（埋め込みファイルと外部ファイルの両方に対応）
		// Requirement 3.1, 3.2, 3.3: 優先順位に従ってSF2ファイルを検索
		app.soundFontLocation = findSoundFont(app.embedFS, selectedTitle.Path, selectedTitle.IsEmbedded)

		if app.soundFontLocation != nil {
			app.soundFontPath = app.soundFontLocation.Path
			opts = append(opts, vm.WithSoundFont(app.soundFontPath))
			app.log.Info("SoundFont configured", "path", app.soundFontPath, "embedded", app.soundFontLocation.IsEmbedded)
		}

		// VMを作成
		vmInstance = vm.New(opcodes, opts...)

		// オーディオシステムを初期化
		// Ebitengineのオーディオコンテキストは一度しか作成できないため、
		// アプリケーションレベルで保持して再利用する
		// Requirement 2.1: FileSystemインターフェースを使用してSF2ファイルを読み込む
		if app.soundFontLocation != nil {
			var err error
			// 共有オーディオコンテキストがなければ作成
			if app.sharedAudioCtx == nil {
				app.sharedAudioCtx = ebitenAudio.NewContext(audio.SampleRate)
				app.log.Info("Created shared audio context")
			}
			// SoundFontのFileSystemを使用してオーディオシステムを作成
			audioSys, err = audio.NewAudioSystemWithFS(
				app.soundFontLocation.Path,
				vmInstance.GetEventQueue(),
				app.sharedAudioCtx,
				app.soundFontLocation.FileSystem,
			)
			if err != nil {
				app.log.Warn("Failed to initialize audio system", "error", err)
			} else {
				// 埋め込みタイトルの場合はMIDI/WAV用のFileSystemを設定
				if selectedTitle.IsEmbedded {
					embedFS := fileutil.NewEmbedFS(app.embedFS, selectedTitle.Path)
					audioSys.SetFileSystem(embedFS)
					app.log.Info("Audio system using embedded file system for MIDI/WAV", "basePath", selectedTitle.Path)
				}
				vmInstance.SetAudioSystem(audioSys)
				app.log.Info("Audio system initialized")
			}
		}

		// グラフィックスシステムを初期化
		graphicsSys = graphics.NewGraphicsSystem(
			selectedTitle.Path,
			graphics.WithLogger(app.log),
		)
		if selectedTitle.IsEmbedded {
			graphicsSys.SetEmbedFS(app.embedFS)
		}
		vmInstance.SetGraphicsSystem(graphicsSys)
		app.log.Info("Graphics system initialized")

		graphicsSys.SetDebugOverlayFromLogLevelString(app.config.LogLevel)

		// GameにVM/GraphicsSystemを設定
		game.SetGraphicsSystem(graphicsSys)
		game.SetVMRunner(vmInstance)
		game.SetEventPusher(vmInstance)

		// VMを開始する関数を設定
		game.SetVMStartFunc(func() {
			app.log.Info("Starting VM execution in background (after mode transition)")
			go func() {
				vmErrCh <- vmInstance.Run()
			}()
		}, vmErrCh)

		return nil
	})

	// ウィンドウ設定
	ebiten.SetWindowSize(1024, 768)
	ebiten.SetWindowTitle("son-et - FILLY interpreter")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	// ゲームを実行（選択画面 -> デスクトップモードまで）
	app.log.Info("Starting Ebitengine game loop (selection mode)")
	if err := ebiten.RunGame(game); err != nil {
		app.log.Error("Ebitengine game loop failed", "error", err)
		if vmInstance != nil {
			vmInstance.Stop()
		}
		return nil, fmt.Errorf("game loop failed: %w", err)
	}

	// モード遷移時のエラーをチェック
	if err := game.GetTransitionError(); err != nil {
		return nil, err
	}

	// クリーンアップ
	if graphicsSys != nil {
		graphicsSys.Shutdown()
		app.log.Info("Graphics system shut down")
	}
	if audioSys != nil {
		vmInstance.ShutdownAudio()
		app.log.Info("Audio system shut down")
	}

	// VMの終了を待つ
	select {
	case vmErr := <-vmErrCh:
		if vmErr != nil {
			app.log.Error("VM execution failed", "error", vmErr)
			return game.GetSelectedTitle(), vmErr
		}
	default:
		if vmInstance != nil {
			vmInstance.Stop()
		}
	}

	return game.GetSelectedTitle(), nil
}

// runVM VMを実行
// Requirement 13.1: Application integrates VM after compilation.
// Requirement 13.2: Application passes compiled OpCode to VM.
// Requirement 13.3: Application starts VM execution.
func (app *Application) runVM() error {
	app.log.Info("Creating VM", "opcode_count", len(app.opcodes))

	// VMオプションを設定
	opts := []vm.Option{
		vm.WithHeadless(app.config.Headless),
		vm.WithLogger(app.log),
		vm.WithTitlePath(app.selectedTitle.Path),
	}

	// タイムアウトが指定されている場合
	if app.config.Timeout > 0 {
		opts = append(opts, vm.WithTimeout(app.config.Timeout))
	}

	// SoundFontパスを設定（埋め込みファイルと外部ファイルの両方に対応）
	// Requirement 3.1, 3.2, 3.3: 優先順位に従ってSF2ファイルを検索
	if app.soundFontLocation == nil {
		app.soundFontLocation = findSoundFont(app.embedFS, app.selectedTitle.Path, app.selectedTitle.IsEmbedded)
	}

	if app.soundFontLocation != nil {
		app.soundFontPath = app.soundFontLocation.Path
		opts = append(opts, vm.WithSoundFont(app.soundFontPath))
		app.log.Info("SoundFont configured", "path", app.soundFontPath, "embedded", app.soundFontLocation.IsEmbedded)
	}

	// VMを作成
	vmInstance := vm.New(app.opcodes, opts...)

	// オーディオシステムを初期化（SoundFontが設定されている場合）
	// Requirement 2.1: FileSystemインターフェースを使用してSF2ファイルを読み込む
	if app.soundFontLocation != nil {
		audioSys, err := audio.NewAudioSystemWithFS(
			app.soundFontLocation.Path,
			vmInstance.GetEventQueue(),
			nil, // audioCtx - 新規作成
			app.soundFontLocation.FileSystem,
		)
		if err != nil {
			app.log.Warn("Failed to initialize audio system", "error", err)
			// オーディオシステムの初期化に失敗しても続行
		} else {
			// 埋め込みタイトルの場合はMIDI/WAV用のFileSystemを設定
			if app.selectedTitle.IsEmbedded {
				embedFS := fileutil.NewEmbedFS(app.embedFS, app.selectedTitle.Path)
				audioSys.SetFileSystem(embedFS)
				app.log.Info("Audio system using embedded file system for MIDI/WAV", "basePath", app.selectedTitle.Path)
			}
			vmInstance.SetAudioSystem(audioSys)
			app.log.Info("Audio system initialized")

			// クリーンアップを設定
			defer func() {
				vmInstance.ShutdownAudio()
				app.log.Info("Audio system shut down")
			}()
		}
	}

	// グラフィックスシステムを初期化
	// 要件 10.4: ヘッドレスモードが有効のとき、描画操作をログに記録するのみで実際の描画を行わない
	if app.config.Headless {
		// ヘッドレスモード用のダミーGraphicsSystemを使用
		headlessGS := graphics.NewHeadlessGraphicsSystem(
			graphics.WithHeadlessLogger(app.log),
			graphics.WithLogOperations(true),
		)
		vmInstance.SetGraphicsSystem(headlessGS)
		app.log.Info("Headless graphics system initialized")

		defer func() {
			headlessGS.Shutdown()
			app.log.Info("Headless graphics system shut down")
		}()
	} else {
		// 通常のGraphicsSystemを使用
		graphicsSys := graphics.NewGraphicsSystem(
			app.selectedTitle.Path,
			graphics.WithLogger(app.log),
		)
		// 埋め込みタイトルの場合はembed.FSを設定
		if app.selectedTitle.IsEmbedded {
			graphicsSys.SetEmbedFS(app.embedFS)
		}
		vmInstance.SetGraphicsSystem(graphicsSys)
		app.log.Info("Graphics system initialized")

		// ログレベルに基づいてデバッグオーバーレイを有効化
		graphicsSys.SetDebugOverlayFromLogLevelString(app.config.LogLevel)

		// クリーンアップを設定
		defer func() {
			graphicsSys.Shutdown()
			app.log.Info("Graphics system shut down")
		}()
	}

	// VMを実行
	app.log.Info("Starting VM execution")
	if err := vmInstance.Run(); err != nil {
		app.log.Error("VM execution failed", "error", err)
		return fmt.Errorf("VM execution failed: %w", err)
	}

	app.log.Info("VM execution completed")
	return nil
}

// loadScripts スクリプトファイルを読み込む
func (app *Application) loadScripts(selectedTitle *title.FillyTitle) ([]script.Script, error) {
	var loader *script.Loader
	if selectedTitle.IsEmbedded {
		// 埋め込みタイトルの場合はembed.FSを使用
		loader = script.NewEmbeddedLoader(selectedTitle.Path, app.embedFS)
	} else {
		// 外部タイトルの場合は実ファイルシステムを使用
		loader = script.NewLoader(selectedTitle.Path)
	}
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
		var opcodes []compiler.OpCode
		var result *compiler.PreprocessResult
		var err error

		if selectedTitle.IsEmbedded {
			// 埋め込みタイトルの場合はembed.FSを使用
			opcodes, result, err = compiler.CompileWithPreprocessorFS(selectedTitle.Path, selectedTitle.EntryFile, app.embedFS)
		} else {
			opcodes, result, err = compiler.CompileWithPreprocessor(selectedTitle.Path, selectedTitle.EntryFile)
		}
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
	var opcodes []compiler.OpCode
	var result *compiler.PreprocessResult

	if selectedTitle.IsEmbedded {
		// 埋め込みタイトルの場合はembed.FSを使用
		opcodes, result, err = compiler.CompileWithPreprocessorFS(selectedTitle.Path, mainInfo.FileName, app.embedFS)
	} else {
		opcodes, result, err = compiler.CompileWithPreprocessor(selectedTitle.Path, mainInfo.FileName)
	}
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
