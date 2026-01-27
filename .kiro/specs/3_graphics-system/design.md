# 設計書: 描画システム (Graphics System)

## 概要

描画システムは、FILLYスクリプトの描画機能を実装します。Ebitengineを使用し、ピクチャー管理、ウィンドウ管理、キャスト管理、テキスト描画、描画プリミティブを提供します。

### 主要機能

1. **ピクチャーシステム**: 画像データの読み込み、生成、転送
2. **ウィンドウシステム**: 仮想ウィンドウの管理
3. **キャストシステム**: スプライトの管理
4. **テキストシステム**: 文字列描画
5. **描画プリミティブ**: 基本図形描画
6. **描画コマンドキュー**: メインスレッド制約への対応

### 設計原則

- **メインスレッド制約**: Ebitengineの描画APIはメインスレッドでのみ呼び出し可能
- **コマンドキュー**: イベントハンドラからの描画はキューイングして実行
- **リソース管理**: 明示的な解放とID再利用
- **エラー耐性**: 致命的でないエラーは記録して継続

## アーキテクチャ

### Ebitengineのメインスレッド制約

**重要**: Ebitengineの描画API（`ebiten.Image`への描画操作）はメインスレッドでのみ呼び出し可能です。VMのイベントハンドラ（mes()ブロック）は別のコンテキストで実行されるため、直接描画APIを呼び出すことはできません。

この制約に対応するため、以下のアーキテクチャを採用します：

```
┌─────────────────────────────────────────────────────────────────┐
│                     メインスレッド (Ebitengine)                   │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐         │
│  │  Update()   │ -> │   Draw()    │ -> │  Layout()   │         │
│  │             │    │             │    │             │         │
│  │ 1. VMイベント│    │ 1. キュー   │    │ スケーリング │         │
│  │    処理     │    │    処理     │    │             │         │
│  │ 2. オーディオ│    │ 2. ウィンドウ│    │             │         │
│  │    更新     │    │    描画     │    │             │         │
│  │ 3. 入力処理 │    │ 3. キャスト │    │             │         │
│  │             │    │    描画     │    │             │         │
│  └─────────────┘    └─────────────┘    └─────────────┘         │
└─────────────────────────────────────────────────────────────────┘
         ↑                    ↑
         │                    │
    コマンド追加          コマンド実行
         │                    │
┌────────┴────────────────────┴───────────────────────────────────┐
│                    描画コマンドキュー                             │
│                   (スレッドセーフ)                               │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐              │
│  │MovePic  │→│PutCast  │→│TextWrite│→│DrawRect │→ ...         │
│  └─────────┘ └─────────┘ └─────────┘ └─────────┘              │
└─────────────────────────────────────────────────────────────────┘
         ↑
         │ Push (どのスレッドからでも可)
         │
┌────────┴────────────────────────────────────────────────────────┐
│                    VMイベントハンドラ                            │
│                   (mes()ブロック)                               │
│                                                                 │
│  mes(MIDI_TIME) {                                               │
│      MovePic(...);  // → キューにPush                           │
│      PutCast(...);  // → キューにPush                           │
│  }                                                              │
└─────────────────────────────────────────────────────────────────┘
```

### 処理フロー

1. **VMイベントハンドラ**が描画関数（MovePic、PutCast等）を呼び出す
2. 描画関数は**実際の描画を行わず**、コマンドをキューにPushする
3. Ebitengineの**Draw()メソッド**（メインスレッド）でキューからコマンドを取り出し実行
4. 実際の`ebiten.Image`への描画はDraw()内でのみ行われる

### システム構成図

```
┌─────────────────────────────────────────────────────────────────┐
│                        Application                               │
│                      (pkg/app/app.go)                           │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Game Loop (Ebitengine)                      │
│                    (pkg/window/window.go)                        │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │   Update()   │  │    Draw()    │  │   Layout()   │          │
│  │  - VM Event  │  │  - Render    │  │  - Scaling   │          │
│  │  - Audio     │  │  - Windows   │  │              │          │
│  │  - Input     │  │  - Casts     │  │              │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
└────────────────────────────┬────────────────────────────────────┘
                             │
         ┌───────────────────┼───────────────────┐
         │                   │                   │
         ▼                   ▼                   ▼
┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐
│  Graphics       │  │  Virtual        │  │  Command        │
│  System         │  │  Machine        │  │  Queue          │
│                 │  │                 │  │                 │
│ ┌─────────────┐ │  │ ┌─────────────┐ │  │ ┌─────────────┐ │
│ │  Picture    │ │  │ │  Event      │ │  │ │  Draw Cmd   │ │
│ │  Manager    │ │  │ │  Dispatcher │ │  │ │  (FIFO)     │ │
│ └─────────────┘ │  │ └─────────────┘ │  │ └─────────────┘ │
│ ┌─────────────┐ │  │ ┌─────────────┐ │  │                 │
│ │  Window     │ │  │ │  Handler    │ │  │                 │
│ │  Manager    │ │  │ │  Registry   │ │  │                 │
│ └─────────────┘ │  │ └─────────────┘ │  │                 │
│ ┌─────────────┐ │  │ ┌─────────────┐ │  │                 │
│ │  Cast       │ │  │ │  Audio      │ │  │                 │
│ │  Manager    │ │  │ │  System     │ │  │                 │
│ └─────────────┘ │  │ └─────────────┘ │  │                 │
│ ┌─────────────┐ │  │                 │  │                 │
│ │  Text       │ │  │                 │  │                 │
│ │  Renderer   │ │  │                 │  │                 │
│ └─────────────┘ │  │                 │  │                 │
└─────────────────┘  └─────────────────┘  └─────────────────┘
```

## コンポーネントとインターフェース

### 1. Graphics System (pkg/graphics/graphics.go)

#### 構造体

```go
// GraphicsSystem は描画システム全体を管理する
type GraphicsSystem struct {
    pictures     *PictureManager
    windows      *WindowManager
    casts        *CastManager
    textRenderer *TextRenderer
    cmdQueue     *CommandQueue
    
    // 仮想デスクトップ
    virtualWidth  int
    virtualHeight int
    
    // 描画状態
    paintColor color.Color
    lineSize   int
    
    // ログ
    log *slog.Logger
    mu  sync.RWMutex
}

func NewGraphicsSystem(opts ...Option) *GraphicsSystem
func (gs *GraphicsSystem) Update()  // ゲームループから呼び出し
func (gs *GraphicsSystem) Draw(screen *ebiten.Image)  // 描画
func (gs *GraphicsSystem) Shutdown()
```

#### Draw関数の実装

Draw関数は以下の順序で描画を行います：

1. **背景色**: `pkg/window/window.go`で既に設定されているため、GraphicsSystemでは塗りつぶさない
2. **ウィンドウ**: Z順序でソートされたウィンドウを順次描画
   - 各ウィンドウに対して`drawWindowDecoration()`を呼び出し
   - ウィンドウ装飾（枠、タイトルバー）とコンテンツを描画
3. **キャスト**: 各ウィンドウに属するキャストを描画

```go
func (gs *GraphicsSystem) Draw(screen *ebiten.Image) {
    gs.mu.RLock()
    defer gs.mu.RUnlock()

    // 背景色は window.go の drawDesktop() で既に設定されている

    // ウィンドウをZ順序で取得
    windows := gs.windows.GetWindowsOrdered()

    // 各ウィンドウを描画
    for _, win := range windows {
        if !win.Visible {
            continue
        }

        // ピクチャーを取得
        pic, err := gs.pictures.GetPic(win.PicID)
        if err != nil {
            continue
        }

        // ウィンドウ装飾を描画（Windows 3.1風）
        gs.drawWindowDecoration(screen, win, pic)

        // キャストを描画
        // ...
    }
}
```

### 2. Picture Manager (pkg/graphics/picture.go)

#### 構造体

```go
// Picture はメモリ上の画像データを表す
type Picture struct {
    ID            int
    Image         *ebiten.Image  // 現在の画像（テキスト描画後）
    OriginalImage *image.RGBA    // 元の背景画像（テキスト描画前）
    Width         int
    Height        int
}
```

**OriginalImageフィールドについて**:
- テキスト描画時のアンチエイリアス影問題を解決するために追加
- LoadPic/CreatePic時に元の画像をRGBAとして保存
- TextWriteでは、このOriginalImageを基準に差分を取ることで、同じ位置に別の色でテキストを描画しても前のテキストの影が残らない

// PictureManager はピクチャーを管理する
type PictureManager struct {
    pictures map[int]*Picture
    nextID   int
    maxID    int  // 最大256
    basePath string  // 画像ファイルの基準パス
    mu       sync.RWMutex
}

func NewPictureManager(basePath string) *PictureManager
func (pm *PictureManager) LoadPic(filename string) (int, error)
func (pm *PictureManager) CreatePic(width, height int) (int, error)
func (pm *PictureManager) CreatePicFrom(srcID, width, height int) (int, error)
func (pm *PictureManager) DelPic(id int) error
func (pm *PictureManager) GetPic(id int) (*Picture, error)
func (pm *PictureManager) PicWidth(id int) int
func (pm *PictureManager) PicHeight(id int) int
```

#### ファイル検索

ファイル検索は既存の `pkg/fileutil/fileutil.go` を使用します。大文字小文字非依存の検索機能が実装済みです。

```go
// LoadPicでのファイル検索
func (pm *PictureManager) LoadPic(filename string) (int, error) {
    // fileutil.FindFileを使用して大文字小文字非依存で検索
    fullPath, err := fileutil.FindFile(pm.basePath, filename)
    if err != nil {
        return -1, fmt.Errorf("file not found: %s", filename)
    }
    // ...
}
```

#### RLE圧縮BMPデコーダー (pkg/graphics/bmp.go)

Go標準ライブラリの`image/bmp`はRLE圧縮をサポートしていないため、カスタムデコーダーを実装しています。

**サポートする圧縮方式**:
- BI_RGB (0): 非圧縮
- BI_RLE8 (1): 8ビットRLE圧縮
- BI_RLE4 (2): 4ビットRLE圧縮

**RLE8デコードアルゴリズム**:
```
2バイトペアを読み取る:
- 最初のバイトが0でない場合: 2番目のバイトを最初のバイト回繰り返す
- 最初のバイトが0の場合:
  - 2番目のバイトが0: 行末 (End of Line)
  - 2番目のバイトが1: ビットマップ終了 (End of Bitmap)
  - 2番目のバイトが2: デルタ（位置移動）
  - それ以外: 絶対モード（2番目のバイト個のピクセルをそのまま読み取る）
```

**RLE4デコードアルゴリズム**:
RLE8と同様だが、1バイトに2ピクセル（上位4ビットと下位4ビット）が格納される。

**LoadPicでの使用**:
```go
func (pm *PictureManager) LoadPic(filename string) (int, error) {
    // BMPファイルの場合、RLE圧縮かどうかを確認
    if isBMPFile(fullPath) {
        isRLE, err := IsBMPRLECompressed(file)
        if isRLE {
            // RLE圧縮BMPの場合、カスタムデコーダーを使用
            img, err = DecodeBMP(file)
        } else {
            // 非圧縮BMPの場合、標準デコーダーを使用
            img, _, err = image.Decode(file)
        }
    }
    // ...
}
```

### 3. Picture Transfer (pkg/graphics/transfer.go)

#### 転送関数

```go
// MovePic はピクチャー間で画像を転送する
// mode: 0=通常, 1=透明色除外, 2-9=シーンチェンジ
func (gs *GraphicsSystem) MovePic(
    srcID, srcX, srcY, width, height int,
    dstID, dstX, dstY int,
    mode int, speed int,
) error

// MoveSPic は拡大縮小して転送する
func (gs *GraphicsSystem) MoveSPic(
    srcID, srcX, srcY, srcW, srcH int,
    dstID, dstX, dstY, dstW, dstH int,
) error

// TransPic は指定した透明色を除いて転送する
func (gs *GraphicsSystem) TransPic(
    srcID, srcX, srcY, width, height int,
    dstID, dstX, dstY int,
    transColor color.Color,
) error

// ReversePic は左右反転して転送する
func (gs *GraphicsSystem) ReversePic(
    srcID, srcX, srcY, width, height int,
    dstID, dstX, dstY int,
) error
```

#### シーンチェンジ実装

```go
// SceneChangeMode はシーンチェンジのモードを表す
type SceneChangeMode int

const (
    SceneChangeNone       SceneChangeMode = 0  // 通常コピー
    SceneChangeTransparent SceneChangeMode = 1  // 透明色除外
    SceneChangeWipeDown   SceneChangeMode = 2  // 上から下
    SceneChangeWipeRight  SceneChangeMode = 3  // 左から右
    SceneChangeWipeLeft   SceneChangeMode = 4  // 右から左
    SceneChangeWipeUp     SceneChangeMode = 5  // 下から上
    SceneChangeWipeOut    SceneChangeMode = 6  // 中央から外側
    SceneChangeWipeIn     SceneChangeMode = 7  // 外側から中央
    SceneChangeRandom     SceneChangeMode = 8  // ランダムブロック
    SceneChangeFade       SceneChangeMode = 9  // フェード
)

// SceneChange はシーンチェンジを管理する
type SceneChange struct {
    srcPic    *Picture
    dstPic    *Picture
    srcRect   image.Rectangle
    dstPoint  image.Point
    mode      SceneChangeMode
    speed     int
    progress  float64  // 0.0 - 1.0
    completed bool
}

func (sc *SceneChange) Update() bool  // 完了したらtrue
func (sc *SceneChange) Apply()        // 現在の進捗を適用
```

### 4. Window Manager (pkg/graphics/window.go)

#### 構造体

```go
// Window は仮想ウィンドウを表す
type Window struct {
    ID       int
    PicID    int           // 関連付けられたピクチャー
    X, Y     int           // 仮想デスクトップ上の位置
    Width    int           // 表示幅
    Height   int           // 表示高さ
    PicX     int           // ピクチャー内の参照X
    PicY     int           // ピクチャー内の参照Y
    BgColor  color.Color   // 背景色
    Caption  string        // キャプション
    Visible  bool
    ZOrder   int           // Z順序（大きいほど前面）
    Casts    []int         // このウィンドウに属するキャストID
}

// WindowManager はウィンドウを管理する
type WindowManager struct {
    windows    map[int]*Window
    nextID     int
    maxID      int  // 最大64
    nextZOrder int
    mu         sync.RWMutex
}

func NewWindowManager() *WindowManager
func (wm *WindowManager) OpenWin(picID int, opts ...WinOption) (int, error)
func (wm *WindowManager) MoveWin(id int, opts ...WinOption) error
func (wm *WindowManager) CloseWin(id int) error
func (wm *WindowManager) CloseWinAll()
func (wm *WindowManager) GetWin(id int) (*Window, error)
func (wm *WindowManager) GetWindowsOrdered() []*Window  // Z順序でソート
```

#### ウィンドウオプション

```go
type WinOption func(*Window)

func WithPosition(x, y int) WinOption
func WithSize(width, height int) WinOption
func WithPicOffset(picX, picY int) WinOption
func WithBgColor(c color.Color) WinOption
func WithPicID(picID int) WinOption
```

#### ウィンドウ装飾

仮想ウィンドウはWindows 3.1風の装飾で描画されます：

```go
const (
    BorderThickness = 4  // 外枠の幅
    TitleBarHeight  = 20 // タイトルバーの高さ
)

// ウィンドウ装飾の色
var (
    titleBarColor  = color.RGBA{0, 0, 128, 255}     // 濃い青
    borderColor    = color.RGBA{192, 192, 192, 255} // グレー
    highlightColor = color.RGBA{255, 255, 255, 255} // 白（立体効果のハイライト）
    shadowColor    = color.RGBA{0, 0, 0, 255}       // 黒（立体効果の影）
)
```

**装飾の構成**:
1. **外枠**: グレーの背景に3D効果（上と左に白いハイライト、下と右に黒い影）
2. **タイトルバー**: 濃い青（#000080）の矩形、高さ20ピクセル
3. **コンテンツ領域**: ピクチャーが表示される領域

**座標計算**:
- ウィンドウ全体のサイズ = コンテンツ幅 + BorderThickness × 2
- ウィンドウ全体の高さ = コンテンツ高さ + BorderThickness × 2 + TitleBarHeight
- コンテンツ領域の開始位置 = (win.X + BorderThickness, win.Y + BorderThickness + TitleBarHeight)

**描画順序**:
1. グレーの背景を全体に描画
2. 3D枠線効果を描画（vector.StrokeLineを使用）
3. タイトルバーを描画
4. コンテンツ領域にピクチャーを描画

**実装参考**: `_old_implementation2/pkg/engine/renderer.go`のrenderWindow関数

### 5. Cast Manager (pkg/graphics/cast.go)

#### 概要

キャストはスプライト（動くキャラクター）として機能します。`PutCast`で配置し、`MoveCast`で位置やソース領域を更新します。キャストは毎フレーム`drawCastsForWindow`で描画され、背景画像に焼き付けられることはありません。

#### 動作原理

1. **PutCast**: キャストを作成し、ウィンドウに配置
2. **MoveCast**: キャストの位置/ソース領域を更新（描画は行わない）
3. **描画**: 毎フレーム`Draw()`内で`drawCastsForWindow()`が呼ばれ、すべての可視キャストが描画される
4. **DelCast**: キャストを削除

この設計により、キャストを移動してもアニメーションが正しく表示され、残像が発生しません。

#### 構造体

```go
// Cast はスプライトを表す
type Cast struct {
    ID       int
    WinID    int           // 所属するウィンドウ
    PicID    int           // ソースピクチャー
    X, Y     int           // ウィンドウ内の位置
    SrcX     int           // ピクチャー内のソースX
    SrcY     int           // ピクチャー内のソースY
    Width    int           // 幅
    Height   int           // 高さ
    Visible  bool
    ZOrder   int           // Z順序
}

// CastManager はキャストを管理する
type CastManager struct {
    casts   map[int]*Cast
    nextID  int
    maxID   int  // 最大1024
    mu      sync.RWMutex
}

func NewCastManager() *CastManager
func (cm *CastManager) PutCast(winID, picID, x, y, srcX, srcY, w, h int) (int, error)
func (cm *CastManager) MoveCast(id int, opts ...CastOption) error
func (cm *CastManager) DelCast(id int) error
func (cm *CastManager) GetCast(id int) (*Cast, error)
func (cm *CastManager) GetCastsByWindow(winID int) []*Cast
func (cm *CastManager) DeleteCastsByWindow(winID int)
```

### 6. Text Renderer (pkg/graphics/text.go)

#### 構造体

```go
// FontSettings はフォント設定を保持する
type FontSettings struct {
    Name      string
    Size      int
    Weight    int
    Italic    bool
    Underline bool
    Strikeout bool
}

// TextSettings はテキスト描画設定を保持する
type TextSettings struct {
    TextColor color.Color
    BgColor   color.Color
    BackMode  int  // 0=透明, 1=不透明
}

// TextRenderer はテキスト描画を管理する
type TextRenderer struct {
    font     *FontSettings
    settings *TextSettings
    face     text.Face
    mu       sync.RWMutex
}

func NewTextRenderer() *TextRenderer
func (tr *TextRenderer) SetFont(name string, size int, opts ...FontOption) error
func (tr *TextRenderer) SetTextColor(c color.Color)
func (tr *TextRenderer) SetBgColor(c color.Color)
func (tr *TextRenderer) SetBackMode(mode int)
func (tr *TextRenderer) TextWrite(pic *Picture, x, y int, text string) error
```

#### レイヤー方式によるテキスト描画（アンチエイリアス影問題の解決）

FILLYスクリプトでは、同じ位置に異なる色でテキストを描画することがあります（例：黒で描画後、白で上書き）。
通常のアルファブレンディングでは、前のテキストのアンチエイリアス部分が「影」として残ってしまいます。

この問題を解決するため、レイヤー方式を採用しています：

```
┌─────────────────────────────────────────────────────────────────┐
│                    レイヤー方式の処理フロー                       │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  1. 元の背景画像（OriginalImage）を取得                          │
│     ┌─────────────┐                                             │
│     │ 背景画像    │  ← LoadPic時に保存されたオリジナル           │
│     │ (文字なし)  │                                             │
│     └─────────────┘                                             │
│                                                                 │
│  2. 背景のコピーにテキストを描画                                  │
│     ┌─────────────┐                                             │
│     │ 背景 + 文字 │  ← アルファブレンディングで描画              │
│     │ (黒い文字)  │                                             │
│     └─────────────┘                                             │
│                                                                 │
│  3. 背景との差分を取り、文字部分だけを抽出                        │
│     ┌─────────────┐                                             │
│     │ 透明背景    │  ← 色が変わった部分 = 文字                   │
│     │ + 文字のみ  │     色が同じ部分 = 透明                      │
│     └─────────────┘                                             │
│                                                                 │
│  4. 抽出したレイヤーを元の背景に合成                              │
│     ┌─────────────┐                                             │
│     │ 最終画像    │  ← draw.Over でアルファブレンディング        │
│     │             │                                             │
│     └─────────────┘                                             │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

**重要なポイント**:
- 毎回「元の背景画像（OriginalImage）」を基準に差分を取る
- これにより、同じ位置に別の色で描画しても、前のテキストの影が残らない
- 後から描画したテキストが完全に前のテキストを上書きする

**実装（pkg/graphics/text_layer.go）**:

```go
// TextLayer はテキスト描画のレイヤーを表す
type TextLayer struct {
    Image  *image.RGBA // 透明背景 + 文字
    PicID  int         // 描画先のピクチャーID
    X, Y   int         // ピクチャー内の描画位置
    Width  int         // レイヤーの幅
    Height int         // レイヤーの高さ
}

// CreateTextLayer はテキストレイヤーを作成する
// 背景に文字を描画し、差分を取って文字部分だけを透明背景で抽出する
func CreateTextLayer(
    background *image.RGBA,  // 元の背景画像（OriginalImage）
    face font.Face,
    text string,
    x, y int,
    fontSize int,
    textColor color.Color,
    picID int,
) *TextLayer
```

**パフォーマンス**:
- ベンチマーク結果: レイヤー作成 ~66μs、10レイヤー合成 ~50μs
- 60fps（16.7ms/フレーム）で十分な余裕がある

#### アンチエイリアス無効化

オリジナルのFILLYと同じ見た目を再現するため、テキスト描画時にアンチエイリアス（スムージング）を無効にします。

```go
// TextWriteではアンチエイリアスを無効にして描画
func (tr *TextRenderer) TextWrite(pic *Picture, x, y int, text string) error {
    // Ebitengineのtext.Drawはデフォルトでアンチエイリアスが有効
    // ビットマップフォントを使用するか、描画後に二値化処理を行う
    
    // 方法1: ビットマップフォントを使用
    // 方法2: 描画後にピクセルを二値化（閾値処理）
    // 方法3: opentype.FaceOptions で Hinting を設定
}
```

#### 日本語フォント対応

FILLYスクリプトではWindows用フォント（MSゴシック、MS明朝等）が指定されますが、macOSやLinuxには存在しないため、フォールバック機能を実装します。

```go
// フォントマッピング（Windows → クロスプラットフォーム）
var fontMapping = map[string][]string{
    "ＭＳ ゴシック":   {"Hiragino Kaku Gothic Pro", "Noto Sans JP", "IPAGothic"},
    "ＭＳ Ｐゴシック": {"Hiragino Kaku Gothic Pro", "Noto Sans JP", "IPAGothic"},
    "ＭＳ 明朝":       {"Hiragino Mincho Pro", "Noto Serif JP", "IPAMincho"},
    "ＭＳ Ｐ明朝":     {"Hiragino Mincho Pro", "Noto Serif JP", "IPAMincho"},
    "ms gothic":       {"Hiragino Kaku Gothic Pro", "Noto Sans JP", "IPAGothic"},
    "ms mincho":       {"Hiragino Mincho Pro", "Noto Serif JP", "IPAMincho"},
}

// loadFont はフォントを読み込む
func (tr *TextRenderer) loadFont(name string, size int) (text.Face, error) {
    // 1. フォントマッピングでフォールバック候補を取得
    candidates := []string{name}
    if mapped, ok := fontMapping[strings.ToLower(name)]; ok {
        candidates = append(candidates, mapped...)
    }
    
    // 2. システムフォントを順番に検索
    for _, fontName := range candidates {
        if face, err := tr.loadSystemFont(fontName, size); err == nil {
            return face, nil
        }
    }
    
    // 3. 埋め込みフォントを使用（フォールバック）
    return tr.loadEmbeddedFont(size)
}

// システムフォントのパス（OS別）
func getSystemFontPaths() []string {
    switch runtime.GOOS {
    case "darwin":
        return []string{
            "/System/Library/Fonts",
            "/Library/Fonts",
            os.ExpandEnv("$HOME/Library/Fonts"),
        }
    case "linux":
        return []string{
            "/usr/share/fonts",
            "/usr/local/share/fonts",
            os.ExpandEnv("$HOME/.fonts"),
        }
    case "windows":
        return []string{
            os.ExpandEnv("$WINDIR/Fonts"),
        }
    default:
        return nil
    }
}

// 埋め込みフォント（日本語対応）
//go:embed fonts/NotoSansJP-Regular.ttf
var defaultFontData []byte
```

### 7. Command Queue (pkg/graphics/queue.go)

#### 構造体

```go
// CommandType は描画コマンドの種類を表す
type CommandType int

const (
    CmdMovePic CommandType = iota
    CmdMoveSPic
    CmdTransPic
    CmdReversePic
    CmdOpenWin
    CmdMoveWin
    CmdCloseWin
    CmdPutCast
    CmdMoveCast
    CmdDelCast
    CmdTextWrite
    CmdDrawLine
    CmdDrawRect
    CmdFillRect
    CmdDrawCircle
)

// Command は描画コマンドを表す
type Command struct {
    Type CommandType
    Args []any
}

// CommandQueue はスレッドセーフな描画コマンドキュー
type CommandQueue struct {
    commands []Command
    mu       sync.Mutex
}

func NewCommandQueue() *CommandQueue
func (cq *CommandQueue) Push(cmd Command)
func (cq *CommandQueue) PopAll() []Command
func (cq *CommandQueue) Len() int
```

### 8. Drawing Primitives (pkg/graphics/primitives.go)

#### 描画関数

```go
// DrawLine は直線を描画する
func (gs *GraphicsSystem) DrawLine(picID, x1, y1, x2, y2 int) error

// DrawRect は矩形を描画する
func (gs *GraphicsSystem) DrawRect(picID, x1, y1, x2, y2, fillMode int) error

// FillRect は矩形を塗りつぶす
func (gs *GraphicsSystem) FillRect(picID, x1, y1, x2, y2 int, c color.Color) error

// DrawCircle は円を描画する
func (gs *GraphicsSystem) DrawCircle(picID, x, y, radius, fillMode int) error

// SetLineSize は線の太さを設定する
func (gs *GraphicsSystem) SetLineSize(size int)

// SetPaintColor は描画色を設定する
func (gs *GraphicsSystem) SetPaintColor(c color.Color)

// GetColor は指定座標のピクセル色を取得する
func (gs *GraphicsSystem) GetColor(picID, x, y int) (color.Color, error)
```

## データモデル

### 色の表現

```go
// FILLYの色形式（0xRRGGBB）をcolor.Colorに変換
func ColorFromInt(c int) color.Color {
    return color.RGBA{
        R: uint8((c >> 16) & 0xFF),
        G: uint8((c >> 8) & 0xFF),
        B: uint8(c & 0xFF),
        A: 0xFF,
    }
}

// color.ColorをFILLYの色形式に変換
func ColorToInt(c color.Color) int {
    r, g, b, _ := c.RGBA()
    return int(r>>8)<<16 | int(g>>8)<<8 | int(b>>8)
}

// 透明色（黒）
var TransparentColor = color.RGBA{0, 0, 0, 0xFF}
```

### 座標系

```go
// 仮想デスクトップ座標（1024x768）
// 左上が(0, 0)、右下が(1023, 767)

// 実際のウィンドウ座標への変換
func (gs *GraphicsSystem) VirtualToScreen(vx, vy int, screenW, screenH int) (int, int) {
    scaleX := float64(screenW) / float64(gs.virtualWidth)
    scaleY := float64(screenH) / float64(gs.virtualHeight)
    scale := min(scaleX, scaleY)
    
    offsetX := (float64(screenW) - float64(gs.virtualWidth)*scale) / 2
    offsetY := (float64(screenH) - float64(gs.virtualHeight)*scale) / 2
    
    return int(float64(vx)*scale + offsetX), int(float64(vy)*scale + offsetY)
}

// スクリーン座標から仮想デスクトップ座標への変換
func (gs *GraphicsSystem) ScreenToVirtual(sx, sy int, screenW, screenH int) (int, int) {
    // 逆変換
}
```

## 正確性プロパティ

### ピクチャーシステムのプロパティ

**プロパティ1: ピクチャーIDの一意性**
*任意の*2つのピクチャーについて、それらのIDは異なる
**検証: 要件 1.2**

**プロパティ2: ピクチャーサイズの正確性**
*任意の*ピクチャーについて、PicWidth/PicHeightは実際の画像サイズと一致する
**検証: 要件 1.7, 1.8**

**プロパティ3: ピクチャー削除後のアクセス**
*任意の*削除されたピクチャーIDについて、アクセス時にエラーが返される
**検証: 要件 1.9**

### ウィンドウシステムのプロパティ

**プロパティ4: ウィンドウZ順序**
*任意の*2つのウィンドウについて、後から開いたウィンドウのZOrderは大きい
**検証: 要件 3.11**

**プロパティ5: ウィンドウ削除時のキャスト削除**
*任意の*ウィンドウについて、CloseWin後にそのウィンドウに属するキャストは存在しない
**検証: 要件 9.2**

### キャストシステムのプロパティ

**プロパティ6: キャストIDの一意性**
*任意の*2つのキャストについて、それらのIDは異なる
**検証: 要件 4.2**

**プロパティ7: キャスト位置の更新**
*任意の*キャストについて、MoveCast後の位置は指定された値と一致する
**検証: 要件 4.3**

### コマンドキューのプロパティ

**プロパティ8: コマンド実行順序**
*任意の*コマンド列について、実行順序はキューへの追加順序と一致する
**検証: 要件 7.4**

**プロパティ9: スレッドセーフ性**
*任意の*並行アクセスについて、データ競合が発生しない
**検証: 要件 7.1**

### リソース管理のプロパティ

**プロパティ10: リソース制限**
*任意の*時点で、ピクチャー数は256以下、ウィンドウ数は64以下、キャスト数は1024以下
**検証: 要件 9.5, 9.6, 9.7**

## エラーハンドリング

### エラーの分類

#### 致命的エラー
- グラフィックスシステムの初期化失敗
- Ebitengineの初期化失敗

#### 非致命的エラー
- ファイルが見つからない
- 無効なピクチャーID
- 無効なウィンドウID
- 無効なキャストID
- リソース制限超過

### エラーログフォーマット

```
[timestamp] [level] [component] function_name: message (args)
```

例:
```
[19:43:20.577] [ERROR] [Graphics] LoadPic: file not found (filename=image.bmp)
[19:43:20.578] [WARN] [Graphics] MovePic: invalid picture ID (srcID=999)
```

## テスト戦略

### ユニットテスト

- PictureManager: 読み込み、生成、削除、サイズ取得
- WindowManager: 開く、移動、閉じる、Z順序
- CastManager: 配置、移動、削除
- TextRenderer: フォント設定、描画
- CommandQueue: Push、PopAll、並行アクセス

### プロパティベーステスト

- プロパティ1-10（上記参照）
- テストライブラリ: gopter

### 統合テスト

- サンプルスクリプトの実行
- 描画結果のスクリーンショット比較（将来）

## 実装の優先順位

### フェーズ1: 基盤
1. GraphicsSystem構造体
2. PictureManager（LoadPic、CreatePic、DelPic）
3. CommandQueue

### フェーズ2: ウィンドウとキャスト
1. WindowManager（OpenWin、MoveWin、CloseWin）
2. CastManager（PutCast、MoveCast、DelCast）
3. 描画ループ統合

### フェーズ3: ピクチャー転送
1. MovePic（mode=0, 1）
2. TransPic、ReversePic
3. MoveSPic

### フェーズ4: テキストと図形
1. TextRenderer（SetFont、TextWrite）
2. 描画プリミティブ（DrawLine、DrawRect、FillRect）

### フェーズ5: シーンチェンジ
1. MovePic（mode=2-9）
2. アニメーション制御

### フェーズ6: VM統合
1. 組み込み関数の実装
2. ヘッドレスモード対応
3. ゲームループ統合

## 技術的な考慮事項

### 並行性

- CommandQueueはsync.Mutexで保護
- PictureManager、WindowManager、CastManagerはsync.RWMutexで保護
- Ebitengineの描画APIはDraw()内でのみ呼び出し

### パフォーマンス

- ピクチャー転送はEbitengineのDrawImage()を使用（GPU加速）
- 変更のないウィンドウは再描画をスキップ
- キャストはバッチ描画で最適化

### メモリ管理

- ピクチャー削除時にebiten.Imageを解放
- ウィンドウ削除時に関連キャストを削除
- シャットダウン時にすべてのリソースを解放

## 依存関係

### 外部ライブラリ

- **github.com/hajimehoshi/ebiten/v2**: 描画エンジン
- **github.com/hajimehoshi/ebiten/v2/vector**: ベクター描画（ウィンドウ装飾用）
- **github.com/hajimehoshi/ebiten/v2/text/v2**: テキスト描画
- **golang.org/x/image/font**: フォント処理
- **golang.org/x/image/bmp**: BMP読み込み

### 内部パッケージ

- **pkg/vm**: 仮想マシン（組み込み関数登録）
- **pkg/logger**: ログ出力
- **pkg/window**: 既存のウィンドウスケルトン

## パッケージ構成

```
pkg/graphics/
├── graphics.go      # GraphicsSystem
├── picture.go       # PictureManager
├── bmp.go           # RLE圧縮BMPデコーダー
├── transfer.go      # ピクチャー転送
├── window.go        # WindowManager
├── cast.go          # CastManager
├── text.go          # TextRenderer
├── text_layer.go    # テキストレイヤー（アンチエイリアス影問題対策）
├── primitives.go    # 描画プリミティブ
├── queue.go         # CommandQueue
├── color.go         # 色変換ユーティリティ
├── scene_change.go  # シーンチェンジ
├── debug.go         # デバッグオーバーレイ
└── fonts/
    └── NotoSansJP-Regular.ttf  # 埋め込みフォント
```

## デバッグオーバーレイ

### 概要

デバッグ時に描画要素のIDを画面上に表示することで、問題の特定を容易にします。ログレベルがDebug（レベル2）以上の場合にのみ表示されます。

### 表示内容

```
┌─────────────────────────────────────────────────────────────────┐
│  ウィンドウ (タイトルバー)                              [W1]    │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────┐                                                    │
│  │ P1      │  ← ピクチャーID（緑色、半透明黒背景）              │
│  │         │                                                    │
│  │  ┌────────────┐                                              │
│  │  │ C1(P2)     │  ← キャストID（黄色、半透明黒背景）          │
│  │  │            │     C1=キャストID、P2=ソースピクチャーID     │
│  │  └────────────┘                                              │
│  └─────────┘                                                    │
└─────────────────────────────────────────────────────────────────┘
```

### 実装

```go
// DebugOverlay はデバッグ情報の描画を管理する
type DebugOverlay struct {
    enabled bool
    font    text.Face
}

// DrawWindowID はウィンドウIDをタイトルバーに描画する
func (do *DebugOverlay) DrawWindowID(screen *ebiten.Image, win *Window) {
    if !do.enabled {
        return
    }
    label := fmt.Sprintf("[W%d]", win.ID)
    // タイトルバーの右側に黄色で描画
    // ...
}

// DrawPictureID はピクチャーIDをウィンドウ内容の左上に描画する
func (do *DebugOverlay) DrawPictureID(screen *ebiten.Image, picID int, x, y int) {
    if !do.enabled {
        return
    }
    label := fmt.Sprintf("P%d", picID)
    // 半透明黒背景 + 緑色テキスト
    // ...
}

// DrawCastID はキャストIDをキャスト位置に描画する
func (do *DebugOverlay) DrawCastID(screen *ebiten.Image, cast *Cast, x, y int) {
    if !do.enabled {
        return
    }
    label := fmt.Sprintf("C%d(P%d)", cast.ID, cast.PicID)
    // 半透明黒背景 + 黄色テキスト
    // ...
}
```

### 色の定義

```go
var (
    debugWindowIDColor  = color.RGBA{255, 255, 0, 255}   // 黄色
    debugPictureIDColor = color.RGBA{0, 255, 0, 255}     // 緑色
    debugCastIDColor    = color.RGBA{255, 255, 0, 255}   // 黄色
    debugBgColor        = color.RGBA{0, 0, 0, 200}       // 半透明黒
)
```
