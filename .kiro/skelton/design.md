# アプリケーションスケルトン - 設計書

## 1. アーキテクチャ概要

### 1.1 全体構成
```
┌─────────────────────────────────────────┐
│           cmd/son-et/main.go            │
│         (エントリーポイント)              │
└─────────────────┬───────────────────────┘
                  │
    ┌─────────────┼─────────────┐
    │             │             │
    ▼             ▼             ▼
┌────────┐   ┌─────────┐   ┌──────────┐
│pkg/cli │   │pkg/     │   │pkg/      │
│        │   │logger   │   │title     │
└────────┘   └─────────┘   └──────────┘
                                │
                    ┌───────────┼───────────┐
                    ▼           ▼           ▼
              ┌──────────┐ ┌────────┐ ┌─────────┐
              │pkg/      │ │pkg/    │ │embed.FS │
              │window    │ │script  │ │         │
              └──────────┘ └────────┘ └─────────┘
                    │
                    ▼
              ┌──────────┐
              │Ebitengine│
              └──────────┘
```

### 1.2 責務の分離
- **cmd/son-et**: アプリケーションの起動とライフサイクル管理
- **pkg/cli**: コマンドライン引数の解析
- **pkg/logger**: ログ出力の管理
- **pkg/title**: FILLYタイトルの検出と管理
- **pkg/script**: スクリプトファイルの読み込みとエンコーディング変換
- **pkg/window**: Ebitengineウィンドウの管理（タイトル選択画面、仮想デスクトップ）

## 2. パッケージ設計

### 2.1 pkg/cli

#### 2.1.1 構造体
```go
type Config struct {
    TitlePath   string        // FILLYタイトルのパス（コマンドライン引数から）
    Timeout     time.Duration // タイムアウト時間（0は無制限）
    LogLevel    string        // ログレベル（debug, info, warn, error）
    Headless    bool          // ヘッドレスモード
    ShowHelp    bool          // ヘルプ表示フラグ
}
```

#### 2.1.2 関数
```go
// ParseArgs コマンドライン引数を解析してConfigを返す
func ParseArgs(args []string) (*Config, error)

// PrintHelp ヘルプメッセージを表示
func PrintHelp()
```

#### 2.1.3 エラーハンドリング
- 無効なフラグ値: わかりやすいエラーメッセージを返す
- ヘルプ表示: `ShowHelp=true` を設定してエラーなしで返す

### 2.2 pkg/logger

#### 2.2.1 関数
```go
// InitLogger ログレベルに応じてslogを初期化
func InitLogger(level string) error

// GetLogger グローバルロガーを取得
func GetLogger() *slog.Logger
```

#### 2.2.2 実装詳細
- `log/slog` の `TextHandler` を使用
- 標準出力に出力
- タイムスタンプ、レベル、メッセージを含む人間が読みやすい形式

### 2.3 pkg/title

#### 2.3.1 構造体
```go
type FillyTitle struct {
    Name string      // タイトル名
    Path string      // タイトルのパス（embedの場合は仮想パス）
    IsEmbedded bool  // embedされたタイトルかどうか
}

type FillyTitleRegistry struct {
    embeddedTitles []FillyTitle  // embedされたタイトル一覧
    externalTitle  *FillyTitle   // 外部から指定されたタイトル
}
```

#### 2.3.2 関数
```go
// NewFillyTitleRegistry FillyTitleRegistryを作成
func NewFillyTitleRegistry(embedFS embed.FS) *FillyTitleRegistry

// LoadExternalTitle 外部ディレクトリからタイトルを読み込む
func (r *FillyTitleRegistry) LoadExternalTitle(path string) error

// GetAvailableTitles 利用可能なタイトル一覧を取得
func (r *FillyTitleRegistry) GetAvailableTitles() []FillyTitle

// SelectTitle タイトルを選択（単一の場合は自動選択）
func (r *FillyTitleRegistry) SelectTitle() (*FillyTitle, bool, error)
// 戻り値: (選択されたタイトル, 選択画面が必要か, エラー)
```

#### 2.3.3 タイトル選択ロジック
1. 外部タイトルが指定されている → そのタイトルを返す（選択画面不要）
2. embedタイトルが1つ → そのタイトルを返す（選択画面不要）
3. embedタイトルが複数 → 選択画面が必要
4. タイトルが0個 → エラー

### 2.4 pkg/script

#### 2.4.1 構造体
```go
type Script struct {
    FileName string  // ファイル名
    Content  string  // UTF-8に変換された内容
    Size     int64   // ファイルサイズ
}

type Loader struct {
    titlePath string
}
```

#### 2.4.2 関数
```go
// NewLoader Loaderを作成
func NewLoader(titlePath string) *Loader

// LoadAllScripts すべての.TFYファイルを読み込む
func (l *Loader) LoadAllScripts() ([]Script, error)

// findScriptFiles .TFYファイルを検出（case-insensitive）
func (l *Loader) findScriptFiles() ([]string, error)

// loadScript 単一のスクリプトファイルを読み込む
func (l *Loader) loadScript(path string) (*Script, error)

// convertShiftJISToUTF8 Shift-JISからUTF-8に変換
func convertShiftJISToUTF8(data []byte) (string, error)
```

#### 2.4.3 実装詳細
- `filepath.Walk` でディレクトリを走査
- 拡張子の比較は `strings.EqualFold` を使用（case-insensitive）
- `golang.org/x/text/encoding/japanese` の `ShiftJIS.NewDecoder()` を使用
- エラー時は詳細なログを出力

### 2.5 pkg/window

#### 2.5.1 構造体
```go
type Mode int

const (
    ModeSelection Mode = iota  // タイトル選択画面
    ModeDesktop                // 仮想デスクトップ
)

type Game struct {
    mode          Mode
    titles        []title.Title
    selectedIndex int
    timeout       time.Duration
    startTime     time.Time
}
```

#### 2.5.2 Ebitengineインターフェース実装
```go
// Update ゲームロジックの更新（Ebitengineが毎フレーム呼び出す）
func (g *Game) Update() error

// Draw 画面描画（Ebitengineが毎フレーム呼び出す）
func (g *Game) Draw(screen *ebiten.Image)

// Layout 画面サイズを返す
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int)
```

#### 2.5.3 関数
```go
// NewGame Gameを作成
func NewGame(mode Mode, titles []title.Title, timeout time.Duration) *Game

// Run Ebitengineを起動
func Run(game *Game) error
```

#### 2.5.4 タイトル選択画面の実装
- 背景色: #0087C8
- タイトル一覧をテキストで縦に表示
- 選択中の項目をハイライト表示
- 上下矢印キーで選択移動
- Enterキーで決定 → `ModeDesktop` に遷移
- Escキーで終了

#### 2.5.5 仮想デスクトップの実装
- 背景色: #0087C8
- ウィンドウサイズ: 1280x720
- ウィンドウタイトル: "FILLY Interpreter"
- タイムアウト処理: `startTime` からの経過時間をチェック

#### 2.5.6 ヘッドレスモード対応
```go
// RunHeadless ヘッドレスモードでタイトル選択を実行
func RunHeadless(titles []title.FillyTitle, timeout time.Duration, reader io.Reader, writer io.Writer) (*title.FillyTitle, error)
```

**設計原則**:
- GUIウィンドウを一切作成しない
- 標準入出力のみを使用して対話する
- `window.Run()`関数は呼び出さない
- Ebitenゲームエンジンを初期化しない

**実装詳細**:
- タイトルが1つの場合は自動選択
- 複数のタイトルがある場合は標準入出力で選択
- タイムアウト処理はコンテキストを使用
- 無効な入力に対してはエラーメッセージを表示して再プロンプト
- 'q'または'Q'入力で終了

## 3. メイン処理フロー

### 3.1 pkg/app パッケージ

#### 3.1.1 構造体
```go
type Application struct {
    config      *cli.Config
    log         *slog.Logger
    titleReg    *title.FillyTitleRegistry
    embedFS     embed.FS
}
```

#### 3.1.2 関数
```go
// New Applicationを作成
func New(embedFS embed.FS) *Application

// Run アプリケーションを実行
func (app *Application) Run(args []string) error

// parseArgs コマンドライン引数を解析
func (app *Application) parseArgs(args []string) error

// initLogger ロガーを初期化
func (app *Application) initLogger() error

// loadTitle タイトルを読み込む
func (app *Application) loadTitle() (*title.FillyTitle, error)

// selectTitle タイトルを選択（選択画面が必要な場合は表示）
func (app *Application) selectTitle(titles []title.FillyTitle) (*title.FillyTitle, error)

// loadScripts スクリプトファイルを読み込む
func (app *Application) loadScripts(titlePath string) ([]script.Script, error)

// runDesktop 仮想デスクトップを実行
func (app *Application) runDesktop() error
```

#### 3.1.3 実装例
```go
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
    
    // 3. タイトルの読み込みと選択
    selectedTitle, err := app.loadTitle()
    if err != nil {
        return fmt.Errorf("failed to load title: %w", err)
    }
    
    // 4. スクリプトファイルの読み込み
    scripts, err := app.loadScripts(selectedTitle.Path)
    if err != nil {
        return fmt.Errorf("failed to load scripts: %w", err)
    }
    
    app.log.Info("Loaded scripts", "count", len(scripts))
    for _, s := range scripts {
        app.log.Info("Script loaded", "file", s.FileName, "size", s.Size)
        app.log.Debug("Script content preview", "file", s.FileName, "preview", truncate(s.Content, 100))
    }
    
    // 5. 仮想デスクトップの実行
    if err := app.runDesktop(); err != nil {
        return fmt.Errorf("failed to run desktop: %w", err)
    }
    
    app.log.Info("Application terminated normally")
    return nil
}

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

func (app *Application) runDesktop() error {
    app.log.Info("Starting virtual desktop")
    
    // ヘッドレスモードの場合は何もしない（将来的にスクリプト実行）
    if app.config.Headless {
        app.log.Info("Headless mode: skipping desktop display")
        return nil
    }
    
    // GUIモードの場合は仮想デスクトップを表示
    _, err := window.Run(window.ModeDesktop, nil, app.config.Timeout)
    return err
}
```

### 3.2 cmd/son-et/main.go
```go
//go:embed titles/*
var embeddedTitles embed.FS

func main() {
    app := app.New(embeddedTitles)
    if err := app.Run(os.Args[1:]); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

## 4. エラーハンドリング戦略

### 4.1 エラーの分類
1. **ユーザーエラー**: コマンドライン引数の誤り、存在しないパス指定
   - わかりやすいメッセージを表示
   - 終了コード: 1

2. **システムエラー**: ファイル読み込み失敗、Ebitengine初期化失敗
   - エラーメッセージとスタックトレースをログ出力
   - 終了コード: 1

3. **正常終了**: ヘルプ表示、タイムアウト、ユーザーによる終了
   - 終了コード: 0

### 4.2 ログ出力レベル
- **Debug**: スクリプト内容のプレビュー、詳細な処理フロー
- **Info**: 起動、タイトル選択、スクリプト読み込み完了
- **Warn**: 非推奨機能の使用、リカバリ可能なエラー
- **Error**: 致命的なエラー、プログラム終了

## 5. テスト戦略

### 5.1 ユニットテスト
各パッケージごとに以下をテスト：
- **pkg/app**: アプリケーションのメインフロー
- **pkg/cli**: 引数解析の正常系・異常系
- **pkg/logger**: ログレベルの設定と出力
- **pkg/title**: タイトル検出、選択ロジック
- **pkg/script**: ファイル検出、Shift-JIS変換
- **pkg/window**: キー入力処理、タイムアウト処理

### 5.2 統合テスト
- サンプルFILLYタイトルを用意して、起動から終了までの一連の流れをテスト
- ヘッドレスモードでの自動テスト

### 5.3 テストデータ
- `.kiro/skelton/testdata/` にサンプルタイトルを配置
- Shift-JISとUTF-8の両方のスクリプトファイルを用意

## 6. 依存関係

### 6.1 外部ライブラリ
```go
require (
    github.com/hajimehoshi/ebiten/v2 v2.6.x
    golang.org/x/text v0.14.x
)
```

### 6.2 標準ライブラリ
- `embed`: タイトルのembed
- `flag`: コマンドライン引数解析
- `log/slog`: ログ出力
- `os`: ファイルシステム操作
- `path/filepath`: パス操作
- `time`: タイムアウト処理

## 7. 将来の拡張ポイント

### 7.1 後続の要件で追加予定
- **pkg/lexer**: スクリプトの字句解析
- **pkg/parser**: スクリプトの構文解析
- **pkg/ast**: 抽象構文木の定義
- **pkg/interpreter**: スクリプトの実行エンジン

### 7.2 仮実装から本実装への移行
- タイトル選択画面: アイコン、サムネイル表示
- 仮想デスクトップ: 実際のレンダリング機能
- スクリプト読み込み: 解析と実行機能の追加

## 8. 正当性プロパティ

### 8.1 コマンドライン引数解析の正当性
**プロパティ 1.1**: すべての有効な引数の組み合わせは、エラーなく解析される
```
∀ valid_args: ParseArgs(valid_args) returns Config without error
```

**プロパティ 1.2**: 無効な引数は、必ずエラーを返す
```
∀ invalid_args: ParseArgs(invalid_args) returns error
```

### 8.2 タイトル選択の正当性
**プロパティ 2.1**: 利用可能なタイトルが存在する場合、必ず1つのタイトルが選択される
```
∀ titles where len(titles) > 0: SelectTitle() returns exactly one title
```

**プロパティ 2.2**: タイトルが存在しない場合、エラーを返す
```
len(titles) == 0 → SelectTitle() returns error
```

### 8.3 スクリプト読み込みの正当性
**プロパティ 3.1**: すべての.TFYファイルが検出される（case-insensitive）
```
∀ file in directory where extension matches ".tfy" (case-insensitive):
    file ∈ LoadAllScripts().files
```

**プロパティ 3.2**: Shift-JISからUTF-8への変換は可逆である
```
∀ valid_shiftjis_data:
    UTF8ToShiftJIS(ShiftJISToUTF8(data)) == data
```

### 8.4 タイムアウトの正当性
**プロパティ 4.1**: タイムアウトが指定された場合、指定時間後に終了する
```
timeout > 0 → program terminates within (timeout + ε) seconds
where ε is a small tolerance for processing time
```

**プロパティ 4.2**: タイムアウトが0の場合、無期限に実行される
```
timeout == 0 → program runs until user termination
```

### 8.5 ヘッドレスモードの正当性

**プロパティ 5.1**: ヘッドレスモードでは、GUIウィンドウが一切作成されない
```
∀ execution where headless == true:
    no GUI_Window instance is created
    AND no Ebiten window creation function is called
```

**プロパティ 5.2**: ヘッドレスモードでタイトル選択が不要な場合、Virtual_Desktopを開かずに処理を続行する
```
headless == true AND needsSelection == false
    → runDesktop() returns immediately without creating window
```

**プロパティ 5.3**: ヘッドレスモードでタイトル選択が必要な場合、標準入出力を使用する
```
headless == true AND needsSelection == true
    → window.RunHeadless() is called
    AND window.Run() is NOT called
```

**プロパティ 5.4**: runDesktop()はヘッドレスモードでウィンドウを作成しない
```
headless == true AND runDesktop() is called
    → function returns immediately
    AND no window is created
    AND action is logged
```

**プロパティ 5.5**: window.Run()とwindow.RunHeadless()は相互排他的に呼び出される
```
∀ execution:
    (window.Run() is called XOR window.RunHeadless() is called)
    OR (neither is called)
```

**プロパティ 5.6**: ヘッドレスモードでも、すべての非GUI機能は正常に動作する
```
headless == true → script loading, logging, timeout work correctly
```

## 9. 実装の優先順位

1. **フェーズ1**: 基本構造
   - pkg/cli, pkg/logger の実装
   - pkg/app の基本構造
   - main.go の基本フロー

2. **フェーズ2**: タイトル管理
   - pkg/title の実装
   - 外部ディレクトリからの読み込み
   - pkg/app へのタイトル管理機能の統合

3. **フェーズ3**: スクリプト読み込み
   - pkg/script の実装
   - Shift-JIS対応
   - pkg/app へのスクリプト読み込み機能の統合

4. **フェーズ4**: GUI
   - pkg/window の実装
   - タイトル選択画面、仮想デスクトップ
   - pkg/app へのGUI機能の統合

5. **フェーズ5**: 統合とテスト
   - 全体の統合
   - テストの作成と実行

## 10. ディレクトリ構造

```
cmd/son-et/         # エントリーポイント
  main.go
pkg/
  app/              # アプリケーションのメインロジック
  cli/              # コマンドライン引数解析
  logger/           # ロガー
  window/           # Ebitengineウィンドウ管理
  title/            # FILLYタイトル管理
  script/           # スクリプトファイル読み込み
```
