# 設計書: 実行エンジン (Execution Engine)

## 概要

実行エンジンは、コンパイラが生成したOpCodeを実行し、Toffyスクリプトのイベントドリブンな動作を実現します。本設計では、MIDIのテンポに同期したタイミング制御を中心に、サウンドシステムとイベント機構を実装します。

### 主要機能

1. **イベントシステム**: イベントキュー、ディスパッチャ、ハンドラ管理
2. **タイマーシステム**: 定期的なTIMEイベント生成
3. **MIDIシステム**: MIDI再生、テンポ抽出、MIDI_TIMEイベント生成
4. **WAVシステム**: WAV再生、複数ストリームのミキシング
5. **ステップ実行**: step()構文のサポート、待機制御
6. **OpCode実行**: 仮想マシンによる命令実行
7. **スコープ管理**: グローバル/ローカル変数スコープ
8. **エラーハンドリング**: 継続実行可能なエラー処理

### 設計原則

- **イベント駆動**: すべての処理はイベントループを中心に動作
- **非同期処理**: オーディオ再生は非同期で実行
- **テスト可能性**: ヘッドレスモードでGUIなしでテスト可能
- **エラー耐性**: 致命的でないエラーは記録して実行継続

## アーキテクチャ

### システム構成図

```
┌─────────────────────────────────────────────────────────────┐
│                        Application                          │
│                    (pkg/app/app.go)                         │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                     Virtual Machine                         │
│                      (pkg/vm/vm.go)                         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │ Event Loop   │  │ OpCode       │  │ Scope        │     │
│  │              │  │ Executor     │  │ Manager      │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
└────────────────────────┬────────────────────────────────────┘
                         │
         ┌───────────────┼───────────────┐
         │               │               │
         ▼               ▼               ▼
┌─────────────┐  ┌─────────────┐  ┌─────────────┐
│   Event     │  │   Audio     │  │   Input     │
│   System    │  │   System    │  │   System    │
│             │  │             │  │             │
│ ┌─────────┐ │  │ ┌─────────┐ │  │ ┌─────────┐ │
│ │ Queue   │ │  │ │  MIDI   │ │  │ │  Mouse  │ │
│ │         │ │  │ │ Player  │ │  │ │ Handler │ │
│ └─────────┘ │  │ └─────────┘ │  │ └─────────┘ │
│ ┌─────────┐ │  │ ┌─────────┐ │  │             │
│ │Dispatch │ │  │ │   WAV   │ │  │             │
│ │   er    │ │  │ │ Player  │ │  │             │
│ └─────────┘ │  │ └─────────┘ │  │             │
│ ┌─────────┐ │  │ ┌─────────┐ │  │             │
│ │Handler  │ │  │ │  Timer  │ │  │             │
│ │Registry │ │  │ │         │ │  │             │
│ └─────────┘ │  │ └─────────┘ │  │             │
└─────────────┘  └─────────────┘  └─────────────┘
```

### コンポーネント概要

#### 1. Virtual Machine (VM)
- OpCodeの実行を管理
- イベントループの制御
- スコープ管理
- 組み込み関数の実行

#### 2. Event System
- イベントキュー: イベントを時系列順に格納
- イベントディスパッチャ: イベントを適切なハンドラに配信
- ハンドラレジストリ: 登録されたハンドラを管理

#### 3. Audio System
- MIDI Player: MIDIファイルの再生とテンポ管理
- WAV Player: WAVファイルの再生とミキシング
- Timer: 定期的なTIMEイベント生成

#### 4. Input System
- Mouse Handler: マウスイベントの処理

## コンポーネントと
インターフェース

### 1. Virtual Machine (VM)

#### 構造体

```go
type VM struct {
    // OpCode実行
    opcodes []compiler.OpCode
    pc      int // プログラムカウンタ
    
    // スコープ管理
    globalScope *Scope
    localScope  *Scope
    callStack   []*StackFrame
    
    // イベントシステム
    eventSystem *EventSystem
    
    // オーディオシステム
    audioSystem *AudioSystem
    
    // 入力システム
    inputSystem *InputSystem
    
    // 実行制御
    running     bool
    headless    bool
    timeout     time.Duration
    
    // デバッグ
    logger      *logger.Logger
}
```

#### 主要メソッド

```go
// New creates a new VM instance
func New(opcodes []compiler.OpCode, opts ...Option) *VM

// Run starts the VM execution loop
func (vm *VM) Run() error

// Stop stops the VM execution
func (vm *VM) Stop()

// Execute executes a single OpCode
func (vm *VM) Execute(op compiler.OpCode) (any, error)

// RegisterBuiltinFunction registers a built-in function
func (vm *VM) RegisterBuiltinFunction(name string, fn BuiltinFunc)
```

### 2. Event System

#### イベントタイプ

```go
type EventType string

const (
    EventTIME        EventType = "TIME"
    EventMIDI_TIME   EventType = "MIDI_TIME"
    EventMIDI_END    EventType = "MIDI_END"
    EventLBDOWN      EventType = "LBDOWN"
    EventRBDOWN      EventType = "RBDOWN"
    EventRBDBLCLK    EventType = "RBDBLCLK"
)
```

#### イベント構造体

```go
type Event struct {
    Type      EventType
    Timestamp time.Time
    Params    map[string]any // MesP1, MesP2, MesP3など
}
```

#### イベントキュー

```go
type EventQueue struct {
    events []Event
    mu     sync.Mutex
}

func (eq *EventQueue) Push(event Event)
func (eq *EventQueue) Pop() (Event, bool)
func (eq *EventQueue) Peek() (Event, bool)
func (eq *EventQueue) Len() int
```

#### イベントハンドラ

```go
type EventHandler struct {
    ID        string
    EventType EventType
    OpCodes   []compiler.OpCode
    VM        *VM
    Active    bool
    
    // ステップ実行用
    StepCounter int
    WaitCounter int
}

func (eh *EventHandler) Execute() error
func (eh *EventHandler) Remove()
```

#### ハンドラレジストリ

```go
type HandlerRegistry struct {
    handlers map[EventType][]*EventHandler
    mu       sync.RWMutex
}

func (hr *HandlerRegistry) Register(eventType EventType, handler *EventHandler) string
func (hr *HandlerRegistry) Unregister(id string)
func (hr *HandlerRegistry) UnregisterAll()
func (hr *HandlerRegistry) GetHandlers(eventType EventType) []*EventHandler
```

#### イベントディスパッチャ

```go
type EventDispatcher struct {
    queue    *EventQueue
    registry *HandlerRegistry
    logger   *logger.Logger
}

func (ed *EventDispatcher) Dispatch(event Event) error
func (ed *EventDispatcher) ProcessQueue() error
```

### 3. Audio System

#### MIDI Player

```go
type MIDIPlayer struct {
    // go-meltysynth
    soundFont  *meltysynth.SoundFont
    synth      *meltysynth.Synthesizer
    sequencer  *meltysynth.MidiFileSequencer
    
    // Ebitengine/audio
    audioCtx   *audio.Context
    player     *audio.Player
    stream     *MIDIStream
    
    // テンポ管理
    tickCalc   *TickCalculator
    
    // イベント生成
    eventSystem *EventSystem
    lastTick    int
    
    // 状態
    playing    bool
    muted      bool
    duration   time.Duration
    mu         sync.RWMutex
}

// MIDIStream implements io.Reader for Ebitengine/audio
type MIDIStream struct {
    sequencer *meltysynth.MidiFileSequencer
    mu        sync.Mutex
}

func (s *MIDIStream) Read(p []byte) (int, error) {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    samples := len(p) / 4 // 16-bit stereo
    left := make([]float32, samples)
    right := make([]float32, samples)
    s.sequencer.Render(left, right)
    
    // Convert float32 to int16 interleaved
    for i := 0; i < samples; i++ {
        l := int16(clamp(left[i], -1, 1) * 32767)
        r := int16(clamp(right[i], -1, 1) * 32767)
        binary.LittleEndian.PutUint16(p[i*4:], uint16(l))
        binary.LittleEndian.PutUint16(p[i*4+2:], uint16(r))
    }
    return len(p), nil
}

// TickCalculator calculates ticks from audio position considering tempo changes
type TickCalculator struct {
    ppq           int
    tempoMap      []TempoEvent
    sampleAtTempo []int64 // pre-calculated sample count at each tempo change
}

func NewMIDIPlayer(soundFontPath string, eventSystem *EventSystem) (*MIDIPlayer, error)
func (mp *MIDIPlayer) Play(filename string) error
func (mp *MIDIPlayer) Stop()
func (mp *MIDIPlayer) SetMuted(muted bool)
func (mp *MIDIPlayer) Update() // Called from game loop to generate MIDI_TIME events
func (mp *MIDIPlayer) IsPlaying() bool
```

#### WAV Player

```go
type WAVPlayer struct {
    // Ebitengine/audio
    audioCtx *audio.Context
    
    // 再生中のプレイヤー
    players []*audio.Player
    mu      sync.Mutex
    
    // 状態
    muted   bool
}

func NewWAVPlayer(audioCtx *audio.Context) *WAVPlayer
func (wp *WAVPlayer) Play(filename string) error
func (wp *WAVPlayer) SetMuted(muted bool)
func (wp *WAVPlayer) StopAll()
```

#### Timer

```go
type Timer struct {
    interval    time.Duration
    eventSystem *EventSystem
    ticker      *time.Ticker
    running     bool
    mu          sync.Mutex
}

func NewTimer(interval time.Duration, eventSystem *EventSystem) *Timer
func (t *Timer) Start()
func (t *Timer) Stop()
```

#### Audio System統合

```go
type AudioSystem struct {
    midiPlayer *MIDIPlayer
    wavPlayer  *WAVPlayer
    timer      *Timer
    audioCtx   *audio.Context // Ebitengine audio context (shared)
}

func NewAudioSystem(soundFontPath string, eventSystem *EventSystem) (*AudioSystem, error)
func (as *AudioSystem) PlayMIDI(filename string) error
func (as *AudioSystem) PlayWAVE(filename string) error
func (as *AudioSystem) SetMuted(muted bool)
func (as *AudioSystem) Update() // Called from game loop
func (as *AudioSystem) Shutdown()
```

**重要: Update()の呼び出し**

MIDI_TIMEイベントの生成は、Ebitengineのゲームループ内で`Update()`を呼び出すことで行います:

```go
func (g *Game) Update() error {
    // オーディオシステムの更新（MIDI_TIMEイベント生成）
    g.audioSystem.Update()
    
    // イベント処理
    g.eventSystem.ProcessQueue()
    
    // ...
    return nil
}
```

これにより、`player.Position()`でオーディオの実再生位置を取得し、正確なタイミングでMIDI_TIMEイベントを生成できます。

### 4. Scope Management

#### スコープ構造体

```go
type Scope struct {
    variables map[string]any
    parent    *Scope
    mu        sync.RWMutex
}

func NewScope(parent *Scope) *Scope
func (s *Scope) Get(name string) (any, bool)
func (s *Scope) Set(name string, value any)
func (s *Scope) GetLocal(name string) (any, bool)
func (s *Scope) SetLocal(name string, value any)
```

#### スタックフレーム

```go
type StackFrame struct {
    functionName string
    localScope   *Scope
    returnPC     int
}
```

### 5. Input System

#### マウスハンドラ

```go
type MouseHandler struct {
    eventSystem *EventSystem
    logger      *logger.Logger
}

func NewMouseHandler(eventSystem *EventSystem) *MouseHandler
func (mh *MouseHandler) HandleMouseEvent(button MouseButton, x, y, windowID int)
```

## データモデル

### OpCode実行状態

```go
type ExecutionState struct {
    PC          int
    Scope       *Scope
    CallStack   []*StackFrame
    BreakFlag   bool
    ContinueFlag bool
}
```

### ステップ実行状態

```go
type StepState struct {
    StepCount   int
    CurrentStep int
    WaitCount   int
    OpCodes     []compiler.OpCode
}
```

### イベントパラメータ

```go
// マウスイベント用
type MouseEventParams struct {
    WindowID int
    X        int
    Y        int
}

// MIDI_TIME用
type MIDITimeParams struct {
    Tick     int
    Tempo    float64
}
```

## 正確性プロパティ

プロパティとは、システムのすべての有効な実行において真であるべき特性や振る舞いの形式的な記述です。プロパティは、人間が読める仕様と機械で検証可能な正確性保証の橋渡しとなります。


### イベントシステムのプロパティ

**プロパティ1: イベントキューの時系列順序保証**
*任意の*イベント列について、キューに追加した後に取り出す順序は、タイムスタンプの昇順である
**検証: 要件 1.1, 1.3**

**プロパティ2: イベントタイムスタンプの自動割り当て**
*任意の*イベントについて、キューに追加された後は必ずタイムスタンプが設定されている
**検証: 要件 1.2**

**プロパティ3: ハンドラの完全呼び出し**
*任意の*イベントタイプと任意の数の登録済みハンドラについて、イベントディスパッチ後にすべてのハンドラが呼び出される
**検証: 要件 1.4**

**プロパティ4: ハンドラの登録順実行**
*任意の*数のハンドラについて、実行順序は登録順序と一致する
**検証: 要件 1.5**

**プロパティ5: イベントパラメータのアクセス可能性**
*任意の*イベントパラメータ（MesP1、MesP2、MesP3）について、ハンドラ実行中にアクセス可能である
**検証: 要件 1.6**

**プロパティ6: ハンドラ登録の成功**
*任意の*イベントタイプについて、OpRegisterEventHandler実行後にハンドラが登録されている
**検証: 要件 2.1**

**プロパティ7: del_meによるハンドラ削除**
*任意の*ハンドラについて、そのハンドラ内でdel_meを呼び出した後、そのハンドラは削除されている
**検証: 要件 2.9**

**プロパティ8: del_allによる全ハンドラ削除**
*任意の*数の登録済みハンドラについて、del_all呼び出し後にすべてのハンドラが削除されている
**検証: 要件 2.10**

### ステップ実行のプロパティ

**プロパティ9: ステップカウンタの初期化**
*任意の*ステップカウント値について、OpSetStep実行後のステップカウンタはその値に等しい
**検証: 要件 6.1**

**プロパティ10: イベントごとのステップ進行**
*任意の*ステップ数について、ステップ実行中にイベントが発生するたびに現在のステップが1つ進む
**検証: 要件 6.3**

**プロパティ11: 連続カンマの待機**
*任意の*数nの連続カンマについて、n回のイベント発生後に次のステップに進む
**検証: 要件 6.6**

**プロパティ12: Wait(n)の待機**
*任意の*正の整数nについて、Wait(n)呼び出し後、n回のイベント発生後に実行が再開される
**検証: 要件 17.1**

### OpCode実行のプロパティ

**プロパティ13: OpCode順次実行**
*任意の*OpCodeシーケンスについて、実行順序はシーケンスの順序と一致する
**検証: 要件 8.1**

**プロパティ14: 変数代入の正確性**
*任意の*変数名と値について、OpAssign実行後にその変数の値は指定された値に等しい
**検証: 要件 8.2**

**プロパティ15: 二項演算の正確性**
*任意の*演算子と2つのオペランドについて、OpBinaryOp実行結果は数学的に正しい演算結果に等しい
**検証: 要件 8.11**

### スコープ管理のプロパティ

**プロパティ16: グローバル変数のスコープ**
*任意の*トップレベルで宣言された変数について、その変数はグローバルスコープに存在する
**検証: 要件 9.1**

**プロパティ17: スコープのラウンドトリップ**
*任意の*関数について、関数呼び出し時にローカルスコープが作成され、関数から戻った後にそのスコープは破棄されている
**検証: 要件 9.3, 9.4**

**プロパティ18: 変数解決の優先順位**
*任意の*変数名について、ローカルスコープとグローバルスコープの両方に同名の変数が存在する場合、ローカルスコープの値が返される
**検証: 要件 9.5**

### エラーハンドリングのプロパティ

**プロパティ19: エラー後の実行継続**
*任意の*致命的でないエラー（ゼロ除算、範囲外アクセス等）について、エラー発生後も実行が継続される
**検証: 要件 11.8**

### オーディオシステムのプロパティ

**プロパティ20: MIDI再生の排他性**
*任意の*2つのMIDIファイルについて、2つ目のPlayMIDI呼び出し時に1つ目のMIDI再生が停止している
**検証: 要件 4.6**

### 配列のプロパティ

**プロパティ21: 配列の自動拡張**
*任意の*配列と任意のインデックスiについて、i番目の要素への代入後、配列のサイズはi+1以上である
**検証: 要件 19.5**

**プロパティ22: 配列の参照渡し**
*任意の*配列について、関数に渡して関数内で変更した後、呼び出し元でも変更が反映されている
**検証: 要件 19.8**

### スタック管理のプロパティ

**プロパティ23: スタックフレームのラウンドトリップ**
*任意の*関数について、関数呼び出し前のスタックサイズをnとすると、呼び出し中はn+1、戻った後はnである
**検証: 要件 20.1, 20.2**

## エラーハンドリング

### エラーの分類

#### 致命的エラー
- スタックオーバーフロー
- メモリ不足
- システムリソースの枯渇

これらのエラーが発生した場合、VMは実行を停止し、エラーメッセージをログに記録します。

#### 非致命的エラー
- ファイルが見つからない
- ゼロ除算
- 配列インデックス範囲外
- 未定義の変数アクセス
- 未定義の関数呼び出し

これらのエラーが発生した場合、VMはエラーをログに記録し、デフォルト値（ゼロまたは空文字列）を返して実行を継続します。

### エラーログフォーマット

```
[timestamp] [level] [component] message
```

例:
```
[19:43:20.577] [ERROR] [VM] Division by zero at line 42
[19:43:20.578] [WARN] [AudioSystem] WAV file not found: sound.wav
```

## テスト戦略

### 二重テストアプローチ

実行エンジンの正確性を保証するため、ユニットテストとプロパティベーステストの両方を使用します。

#### ユニットテスト
- 特定の例とエッジケースを検証
- 統合ポイントのテスト
- エラー条件のテスト

**対象:**
- イベントキューの基本操作
- ハンドラ登録と削除
- スコープの作成と破棄
- ゼロ除算などのエッジケース
- スタックオーバーフローの検出

#### プロパティベーステスト
- 普遍的なプロパティを検証
- ランダム化による包括的な入力カバレッジ

**設定:**
- 最小100イテレーション/テスト
- 各テストは設計書のプロパティを参照
- タグフォーマット: **Feature: execution-engine, Property {number}: {property_text}**

**対象プロパティ:**
- プロパティ1-23（上記参照）

### テストライブラリ

Goの標準的なプロパティベーステストライブラリを使用:
- **gopter**: Go用のプロパティベーステストフレームワーク
- **testing/quick**: Go標準ライブラリのクイックチェック

### ヘッドレスモードでのテスト

ヘッドレスモードを使用することで、GUIなしで実行エンジンをテストできます:

```bash
son-et --headless --timeout=5s <project_directory>
```

**利点:**
- CI/CD環境での自動テスト
- タイミング検証（タイムスタンプ付きログ）
- MIDI_TIME同期の検証（オーディオはミュートだがイベントは生成）
- ログ出力による動作確認

### テスト例

#### ユニットテスト例

```go
func TestEventQueueOrder(t *testing.T) {
    queue := NewEventQueue()
    
    // イベントを逆順で追加
    event1 := Event{Type: EventTIME, Timestamp: time.Now().Add(2 * time.Second)}
    event2 := Event{Type: EventTIME, Timestamp: time.Now().Add(1 * time.Second)}
    event3 := Event{Type: EventTIME, Timestamp: time.Now()}
    
    queue.Push(event1)
    queue.Push(event2)
    queue.Push(event3)
    
    // 時系列順に取り出されることを確認
    e1, _ := queue.Pop()
    e2, _ := queue.Pop()
    e3, _ := queue.Pop()
    
    assert.True(t, e1.Timestamp.Before(e2.Timestamp))
    assert.True(t, e2.Timestamp.Before(e3.Timestamp))
}
```

#### プロパティベーステスト例

```go
// Feature: execution-engine, Property 1: イベントキューの時系列順序保証
func TestEventQueueOrderProperty(t *testing.T) {
    properties := gopter.NewProperties(nil)
    
    properties.Property("events are dequeued in chronological order", 
        prop.ForAll(
            func(events []Event) bool {
                queue := NewEventQueue()
                
                // すべてのイベントをキューに追加
                for _, e := range events {
                    queue.Push(e)
                }
                
                // 取り出して順序を確認
                var prev *Event
                for queue.Len() > 0 {
                    e, _ := queue.Pop()
                    if prev != nil && e.Timestamp.Before(prev.Timestamp) {
                        return false
                    }
                    prev = &e
                }
                
                return true
            },
            genEventList(),
        ))
    
    properties.TestingRun(t, gopter.ConsoleReporter(false))
}
```

## 実装の優先順位

### フェーズ1: コアシステム（最優先）
1. VM基本構造
2. イベントシステム（キュー、ディスパッチャ、ハンドラレジストリ）
3. スコープ管理
4. 基本的なOpCode実行（Assign、Call、BinaryOp等）

### フェーズ2: オーディオシステム
1. Timer（TIMEイベント生成）
2. MIDI Player（MIDI再生、テンポ抽出、MIDI_TIMEイベント生成）
3. WAV Player（WAV再生、ミキシング）
4. Audio System統合

### フェーズ3: ステップ実行
1. OpSetStep、OpWait実装
2. ステップカウンタ管理
3. Wait()関数実装

### フェーズ4: 入力システム
1. マウスイベント処理
2. イベントパラメータ（MesP1、MesP2、MesP3）設定

### フェーズ5: エラーハンドリングとデバッグ
1. エラーログ
2. デバッグモード
3. ヘッドレスモード
4. タイムアウト機能

## 技術的な考慮事項

### 並行性

- イベントキューはスレッドセーフ（sync.Mutex使用）
- オーディオ再生は別ゴルーチンで実行
- タイマーは別ゴルーチンでイベント生成

**重要: Ebitengineの描画制約**
- Ebitengineの描画APIはメインスレッドでのみ呼び出し可能
- 描画コマンド（MovePic、PutCast等）はメインスレッドのキューに積む必要がある
- イベントハンドラから描画を行う場合、描画コマンドをキューイングし、メインループで実行する
- リアルタイム性を維持するため、描画キューの処理は高優先度で行う

### パフォーマンス

- イベントキューのサイズ制限（デフォルト1000イベント）
- スタック深度制限（デフォルト1000フレーム）
- 配列の動的拡張は2倍ずつ

### メモリ管理

- 使用済みイベントは即座に破棄
- ハンドラ削除時はOpCodeも解放
- スコープ破棄時は変数も解放

### タイミング精度

- TIMEイベント: 50ms間隔（デフォルト）
- MIDI_TIMEイベント: MIDIテンポに依存（通常1-2ms）
- タイマーの精度はOSに依存

## 依存関係

### 外部ライブラリ

**MIDI再生の技術選択:**

以下のアプローチを検討し、実験により検証しました:

1. **fluidsynth + 自前のタイミング管理** (以前の実装)
   - fluidsynthでMIDI合成
   - MIDI終了イベントを自前で検出
   - テンポ変更の追跡が複雑
   - 評価: 実装が複雑、CGO依存、メンテナンスコストが高い

2. **gomidi/midi/player + fluidsynth**
   - gomidiでMIDIファイル解析とタイミング管理
   - fluidsynthで音声合成
   - 評価: CGO依存が問題

3. **go-meltysynth + Ebitengine/audio** (採用)
   - [go-meltysynth](https://github.com/sinshu/go-meltysynth)で純粋GoによるMIDI合成
   - Ebitengine/audioでオーディオ出力
   - 評価: **純粋Go、CGO不要、Ebitengineと統合済み**

**採用アプローチ: go-meltysynth + Ebitengine/audio**

実験コード（`experiments/midi_test/ebiten_midi_v2.go`）で検証済み。

```go
// MIDI再生の実装概要

// 1. SoundFontとMIDIファイルの読み込み
soundFont, _ := meltysynth.NewSoundFont(sf2Reader)
midiFile, _ := meltysynth.NewMidiFile(midiReader)

// 2. シンセサイザーとシーケンサーの作成
settings := meltysynth.NewSynthesizerSettings(44100)
synth, _ := meltysynth.NewSynthesizer(soundFont, settings)
sequencer := meltysynth.NewMidiFileSequencer(synth)
sequencer.Play(midiFile, false)

// 3. MIDIStreamでio.Readerを実装（Ebitengine/audioに渡す）
type MIDIStream struct {
    sequencer *meltysynth.MidiFileSequencer
}

func (s *MIDIStream) Read(p []byte) (int, error) {
    samples := len(p) / 4
    left := make([]float32, samples)
    right := make([]float32, samples)
    s.sequencer.Render(left, right)
    // float32 → int16に変換してインターリーブ
    // ...
    return len(p), nil
}

// 4. Ebitengine/audioで再生
audioCtx := audio.NewContext(44100)
player, _ := audioCtx.NewPlayer(stream)
player.Play()

// 5. MIDI_TIMEイベント生成（Update()内で）
// player.Position()でオーディオの実再生位置を取得
// TickCalculatorでテンポマップを考慮してティックに変換
position := player.Position()
currentTick := tickCalculator.FillyTickFromSamples(position)
if currentTick > lastTick {
    // MIDI_TIMEイベントを生成
}
```

**テンポ変更対応:**

MIDIファイルからテンポマップを抽出し、`TickCalculator`で経過時間からティックを計算:

```go
type TickCalculator struct {
    ppq      int           // ticks per quarter note
    tempoMap []TempoEvent  // テンポ変更イベントのリスト
}

// サンプル数からFILLYティック（32分音符単位）を計算
func (tc *TickCalculator) FillyTickFromSamples(samples int64) int {
    midiTick := tc.TickFromSamples(samples)
    return midiTick * 8 / tc.ppq  // 1 quarter note = 8 FILLY ticks
}
```

**利点:**
- 純粋Go実装（CGO不要）
- Ebitengineと統合済み（描画とオーディオが同じフレームワーク）
- `player.Position()`でオーディオ再生位置を正確に取得可能
- テンポ変更に対応
- 長時間再生でもオーディオとティックがずれない

**必要なライブラリ:**
- **github.com/hajimehoshi/ebiten/v2**: ゲームエンジン（描画、オーディオ、入力）
- **github.com/hajimehoshi/ebiten/v2/audio**: オーディオ再生
- **github.com/sinshu/go-meltysynth**: 純粋GoのMIDIシンセサイザー
- **gopter**: プロパティベーステスト

### 内部パッケージ

- **pkg/compiler**: OpCode定義
- **pkg/logger**: ログ出力
- **pkg/script**: スクリプト読み込み

## 今後の拡張

### 描画系機能（別スペック）
- LoadPic、CreatePic、MovePic
- OpenWin、CloseWin、MoveWin
- PutCast、MoveCast、DelCast
- TextWrite、DrawRect、DrawLine

### その他の機能
- キーボードイベント
- ファイルI/O
- ネットワーク通信
