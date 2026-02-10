# 要件定義書

## はじめに

`pkg/vm/audio` パッケージの `TestMIDIPlayerGetCurrentTick` テストが `go test ./... -timeout 30s` 実行時にタイムアウトで失敗する問題を修正する。テストファイル `midi_test.go` 内の複数テストが累積的に長時間かかり、30秒のタイムアウトを超過している。

## 用語集

- **MIDIPlayer**: MIDI ファイルの再生を管理するコンポーネント
- **TestSuite**: `midi_test.go` 内の全テスト関数群
- **タイムアウト**: `go test -timeout` で指定されるテスト全体の実行制限時間
- **time.Sleep**: テスト内でオーディオレンダリングの待機に使用される固定待機呼び出し
- **ポーリングループ**: MIDI 再生完了を待つ `for` ループ（`TestMIDIPlayerMIDIEndEvent` 等で使用）

## 要件

### 要件 1: テスト実行時間の短縮

**ユーザーストーリー:** 開発者として、`go test ./pkg/vm/audio/... -timeout 30s` でテストスイート全体が30秒以内に完了してほしい。CI/CDパイプラインが安定して動作するようにしたい。

#### 受け入れ基準

1. WHEN `go test ./pkg/vm/audio/... -timeout 30s` を実行した場合、THE TestSuite SHALL 30秒以内に全テストを完了する
2. WHEN テストが MIDI 再生の待機を行う場合、THE TestSuite SHALL 必要最小限の待機時間のみ使用する
3. WHEN `TestMIDIPlayerMIDIEndEvent` が長い MIDI ファイルを検出した場合、THE TestSuite SHALL 適切にスキップまたは短時間で完了する

### 要件 2: テストの正確性の維持

**ユーザーストーリー:** 開発者として、テスト実行時間を短縮しても、各テストが検証すべき動作を正しく検証し続けてほしい。

#### 受け入れ基準

1. WHEN テストの待機時間を短縮した場合、THE TestSuite SHALL 各テストの検証ロジックを変更せずに維持する
2. WHEN `TestMIDIPlayerGetCurrentTick` を実行した場合、THE TestSuite SHALL 再生中のティック値が非負であることを検証する
3. WHEN `TestMIDIPlayerUpdate` を実行した場合、THE TestSuite SHALL MIDI_TIME イベントの生成を正しく検証する
4. WHEN `TestMIDIPlayerMIDIEndEvent` を実行した場合、THE TestSuite SHALL MIDI_END イベントの生成を正しく検証する

### 要件 3: テストの安定性

**ユーザーストーリー:** 開発者として、テストが環境やタイミングに依存せず安定して成功してほしい。

#### 受け入れ基準

1. WHEN テストを繰り返し実行した場合、THE TestSuite SHALL 一貫した結果を返す
2. IF テスト環境で MIDI ファイルの再生が遅い場合、THEN THE TestSuite SHALL タイムアウトせずに適切に処理する
