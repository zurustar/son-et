# オーディオシステム

## 概要

son-etのオーディオシステムは、MIDI再生とWAV再生の2つの機能を提供します。
特にMIDI再生は、FILLYスクリプトの `mes(MIDI_TIME)` イベントハンドラと連携し、音楽に同期したアニメーション制御の基盤となります。

### システム構成

```
┌─────────────────────────────────────────────────────────────┐
│                      AudioSystem                            │
│                                                             │
│  ┌─────────────────┐  ┌─────────────────┐  ┌───────────┐  │
│  │   MIDIPlayer    │  │    WAVPlayer    │  │   Timer   │  │
│  │                 │  │                 │  │           │  │
│  │ ┌─────────────┐ │  │  Ebitengine/    │  │  TIMEイベ │  │
│  │ │go-meltysynth│ │  │  audio          │  │  ント生成  │  │
│  │ │ (合成)      │ │  │                 │  │           │  │
│  │ └─────────────┘ │  └─────────────────┘  └───────────┘  │
│  │ ┌─────────────┐ │                                       │
│  │ │TickCalcula- │ │                                       │
│  │ │tor (テンポ) │ │                                       │
│  │ └─────────────┘ │                                       │
│  │ ┌─────────────┐ │                                       │
│  │ │Ebitengine/  │ │                                       │
│  │ │audio (出力) │ │                                       │
│  │ └─────────────┘ │                                       │
│  └─────────────────┘                                       │
└─────────────────────────────────────────────────────────────┘
```

---

## 1. MIDI再生の技術選択と理由

### 検討したアプローチ

MIDI再生の実装にあたり、以下の3つのアプローチを検討・実験しました。

#### アプローチ1: fluidsynth + 自前のタイミング管理（以前の実装）

- fluidsynthでMIDI合成
- MIDI終了イベントを自前で検出
- テンポ変更の追跡が複雑

**評価**: 実装が複雑、CGO依存、メンテナンスコストが高い

#### アプローチ2: gomidi/midi/player + fluidsynth

- gomidiでMIDIファイル解析とタイミング管理
- fluidsynthで音声合成

**評価**: CGO依存が問題

#### アプローチ3: go-meltysynth + Ebitengine/audio（採用）

- [go-meltysynth](https://github.com/sinshu/go-meltysynth) で純粋GoによるMIDI合成
- Ebitengine/audio でオーディオ出力

**評価**: **純粋Go、CGO不要、Ebitengineと統合済み**

### 採用理由

**go-meltysynth + Ebitengine/audio** を採用した理由は以下の通りです。

| 観点 | 詳細 |
|---|---|
| **純粋Go実装** | CGO不要のため、クロスコンパイルが容易 |
| **Ebitengine統合** | 描画とオーディオが同じフレームワークで統一 |
| **再生位置の正確な取得** | `player.Position()` でオーディオの実再生位置を取得可能 |
| **テンポ変更対応** | TickCalculatorによりテンポマップを考慮した正確なティック計算 |
| **長時間再生の安定性** | オーディオとティックがずれない |

### 実装アーキテクチャ

MIDI再生は以下の流れで動作します。

```
1. SoundFont (.sf2) + MIDIファイル (.mid) を読み込み
2. go-meltysynth の Synthesizer + MidiFileSequencer を作成
3. MIDIStream (io.Reader実装) を Ebitengine/audio に渡す
4. Ebitengine/audio の Player で再生
5. Update() 内で player.Position() から現在のティックを計算
6. ティックが進んでいれば MIDI_TIME イベントを生成
```

#### MIDIStream

`MIDIStream` は `io.Reader` インターフェースを実装し、go-meltysynth のシーケンサーから生成されたオーディオデータを Ebitengine/audio に渡します。

```go
type MIDIStream struct {
    sequencer   *meltysynth.MidiFileSequencer
    sampleCount int64
    mu          sync.Mutex
}

func (s *MIDIStream) Read(p []byte) (int, error) {
    s.mu.Lock()
    defer s.mu.Unlock()

    samples := len(p) / 4  // 16-bit stereo = 4 bytes per sample
    left := make([]float32, samples)
    right := make([]float32, samples)

    s.sequencer.Render(left, right)
    s.sampleCount += int64(samples)

    // float32 → int16 に変換してインターリーブ
    for i := 0; i < samples; i++ {
        l := int16(clamp(left[i], -1, 1) * 32767)
        r := int16(clamp(right[i], -1, 1) * 32767)
        binary.LittleEndian.PutUint16(p[i*4:], uint16(l))
        binary.LittleEndian.PutUint16(p[i*4+2:], uint16(r))
    }

    return len(p), nil
}
```

#### MIDI_TIMEイベントの生成

`Update()` メソッドが定期的に呼び出され、オーディオの実再生位置からティックを計算してイベントを生成します。

```go
// Update() 内での MIDI_TIME イベント生成
position := player.Position()
currentTick := tickCalculator.FillyTickFromSamples(position)
if currentTick > lastTick {
    // MIDI_TIME イベントを生成
}
```

**重要**: `Update()` は以下の2箇所で呼び出す必要があります。

| 呼び出し場所 | 用途 |
|---|---|
| Ebitengineのゲームループ内 | GUIモードでの更新 |
| VMのイベントループ内 | ヘッドレスモードを含む全モードでの更新 |

### MIDI再生の排他性

同時に再生できるMIDIファイルは1つだけです。新しい `PlayMIDI()` が呼ばれると、再生中のMIDIは停止されます。

### 使用ライブラリ

| ライブラリ | 用途 |
|---|---|
| [go-meltysynth](https://github.com/sinshu/go-meltysynth) | 純粋GoのMIDIシンセサイザー |
| [Ebitengine/audio](https://github.com/hajimehoshi/ebiten) | オーディオ出力 |

---

## 2. SoundFont検索優先順位

MIDI再生にはSoundFont (.sf2) ファイルが必要です。アプリケーションは以下の優先順位でSoundFontファイルを検索します。

### 検索順序

```
┌─────────────────────────────────────────────────────────────┐
│                    SoundFont検索順序                         │
├─────────────────────────────────────────────────────────────┤
│ 1. 埋め込み: soundfonts/GeneralUser-GS.sf2                  │
│    └─ ビルド時に --soundfont オプションで埋め込み            │
├─────────────────────────────────────────────────────────────┤
│ 2. 埋め込み: {title_path}/GeneralUser-GS.sf2                │
│    └─ タイトルと一緒に埋め込まれたSF2                        │
├─────────────────────────────────────────────────────────────┤
│ 3. 外部: ./GeneralUser-GS.sf2                               │
│    └─ カレントディレクトリのSF2                              │
├─────────────────────────────────────────────────────────────┤
│ 4. 外部: {title_path}/GeneralUser-GS.sf2                    │
│    └─ タイトルディレクトリ内のSF2                            │
└─────────────────────────────────────────────────────────────┘
```

### 設計方針

- **埋め込みファイルが優先**: 埋め込みSF2ファイルが見つかった場合、外部ファイルより優先して使用
- **フォールバック**: どのパスにもSF2ファイルが見つからない場合、MIDI再生を無効化してログに警告を出力
- **FileSystemインターフェース**: 埋め込みファイルシステムと外部ファイルシステムの両方に対応

### FileSystem経由の読み込み

```go
// ReadSoundFontFS は FileSystem インターフェースを使用してSF2ファイルを読み込む。
// fs が nil の場合、os.ReadFile にフォールバックする。
func ReadSoundFontFS(path string, fs fileutil.FileSystem) ([]byte, error)

// LoadSoundFontFS は FileSystem からSF2ファイルを読み込み、パースする。
func LoadSoundFontFS(path string, fs fileutil.FileSystem) (*meltysynth.SoundFont, error)
```

### ビルドスクリプトでの埋め込み

`scripts/build-embedded.sh` の `--soundfont` オプションでSF2ファイルをバイナリに埋め込めます。

```bash
# 使用例
./scripts/build-embedded.sh --soundfont GeneralUser-GS.sf2 samples/my_project my_project
```

処理フロー:

1. `--soundfont` オプションが指定された場合、SF2ファイルの存在を確認
2. SF2ファイルを `cmd/son-et/soundfonts/` ディレクトリにコピー
3. ビルド実行
4. ビルド完了後、コピーしたSF2ファイルをクリーンアップ

---

## 3. MIDIテンポ同期の仕組み（TickCalculator）

### 背景

FILLYスクリプトでは `mes(MIDI_TIME)` ハンドラがMIDIのテンポに同期して呼び出されます。MIDIファイルには途中でテンポが変わる曲もあるため、テンポ変更を正確に追跡する仕組みが必要です。

### TickCalculatorの役割

`TickCalculator` は、オーディオの再生位置（サンプル数）からMIDIティックを計算します。テンポマップ（テンポ変更イベントのリスト）を事前に解析し、テンポ変更を跨いだ正確なティック計算を実現します。

### データ構造

```go
// TempoEvent はMIDIファイル内のテンポ変更を表す
type TempoEvent struct {
    Tick          int
    MicrosPerBeat int // マイクロ秒/四分音符（500000 = 120 BPM）
}

// TickCalculator はテンポ変更を考慮してサンプル数からティックを計算する
type TickCalculator struct {
    ppq           int           // ticks per quarter note（MIDIヘッダから取得）
    tempoMap      []TempoEvent  // テンポ変更イベントのリスト
    sampleAtTempo []int64       // 各テンポ変更時点でのサンプル数（事前計算）
}
```

### 計算アルゴリズム

#### 事前計算（precalculate）

各テンポ変更時点でのサンプル数を事前に計算します。

```
テンポ区間ごとのサンプル数:
  samplesPerTick = sampleRate × microsPerBeat / ppq / 1,000,000
  区間のサンプル数 = ticksInSegment × samplesPerTick
```

#### サンプル数→ティック変換（TickFromSamples）

1. 現在のサンプル数がどのテンポ区間に属するかを特定
2. その区間内でのオフセットサンプル数を計算
3. オフセットをティックに変換
4. 区間の開始ティック + オフセットティック = 現在のMIDIティック

#### MIDIティック→FILLYティック変換（FillyTickFromSamples）

FILLYは32分音符を基本単位として使用します（1四分音符 = 8 FILLYティック）。

```go
// MIDIティックからFILLYティックへの変換
fillyTick = midiTick * 8 / ppq
```

### テンポマップの解析

MIDIファイルからテンポイベント（メタイベント 0x51: Set Tempo）を抽出します。テンポイベントがない場合はデフォルト120 BPM（500,000 microseconds per beat）を使用します。

### 精度の保証

| 項目 | 詳細 |
|---|---|
| テンポ変更対応 | テンポマップにより複数のテンポ変更を正確に追跡 |
| 長時間再生 | サンプル数ベースの計算により、累積誤差が発生しない |
| 再生位置の取得 | `player.Position()` によるオーディオの実再生位置を使用 |

---

## 4. WAV再生

### 基本仕様

WAV再生は効果音の再生に使用されます。MIDI再生とは異なり、複数のWAVファイルを同時に再生できます。

### 機能

| 機能 | 詳細 |
|---|---|
| 同時再生 | 複数の `PlayWAVE()` 呼び出しで全WAVファイルを同時再生 |
| ミキシング | 複数のWAVストリームを単一のオーディオ出力にミックス |
| 対応フォーマット | 標準WAVファイル（PCM、8ビット、16ビット） |
| エラー耐性 | ファイル未検出・破損時はエラーをログに記録して実行継続 |

### 実装構造

```go
type WAVPlayer struct {
    audioCtx *audio.Context   // Ebitengine/audio（MIDIPlayerと共有）
    players  []*audio.Player  // 再生中のプレイヤー
    mu       sync.Mutex
    muted    bool
}
```

### MIDIとの共存

WAVPlayerとMIDIPlayerは同じ `audio.Context` を共有します。Ebitengine/audio が内部でミキシングを行うため、MIDIとWAVの同時再生が可能です。

### エラーハンドリング

WAV再生のエラーは致命的ではありません。

| エラー | 動作 |
|---|---|
| ファイルが見つからない | エラーをログに記録して実行継続 |
| ファイルが破損している | エラーをログに記録して実行継続 |

---

## ヘッドレスモードでのオーディオ

ヘッドレスモード（`--headless`）では、オーディオシステムは以下のように動作します。

| 項目 | 動作 |
|---|---|
| オーディオシステム | 初期化される |
| オーディオ出力 | ミュート |
| MIDI_TIMEイベント | 通常通り生成 |
| テンポ同期 | 正常に動作 |

これにより、CI/CD環境やテスト実行時でも、MIDI_TIMEイベントに依存するスクリプトの動作を検証できます。

---

## タイミング精度

| イベント | 間隔 | 備考 |
|---|---|---|
| TIME | 50ms（デフォルト） | タイマーベース |
| MIDI_TIME | MIDIテンポに依存（通常1-2ms） | オーディオ再生位置ベース |

タイマーの精度はOSに依存しますが、MIDI_TIMEイベントはオーディオの実再生位置から計算されるため、高い精度を維持します。
