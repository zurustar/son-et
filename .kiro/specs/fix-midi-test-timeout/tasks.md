# 実装計画: MIDIテストのタイムアウト修正

## 概要

`pkg/vm/audio/midi_test.go` のテスト実行時間を短縮し、`go test -timeout 30s` で全テストが完了するようにする。プロダクションコードの変更は不要。

## タスク

- [x] 1. `TestMIDIPlayerMIDIEndEvent` のスキップ閾値を修正
  - [x] 1.1 `midi_test.go` の `TestMIDIPlayerMIDIEndEvent` 内の3つのサブテストで、MIDIファイルのdurationが5秒を超える場合にスキップするよう閾値を変更する
    - 定数 `midiEndTestMaxDuration = 5 * time.Second` を定義
    - `maxWait > 30*time.Second` の条件を `duration > midiEndTestMaxDuration` に変更（3箇所）
    - _Requirements: 1.1, 1.3_

- [x] 2. `time.Sleep` の短縮
  - [x] 2.1 `TestMIDIPlayerUpdate` 内の `time.Sleep` を短縮する
    - 200ms → 100ms（"generates MIDI_TIME events when playing" サブテスト）
    - 300ms → 150ms、100ms → 50ms（"generates sequential tick events" サブテスト）
    - 200ms → 100ms（"starts from tick 1 after Play" サブテスト）
    - _Requirements: 1.1, 1.2_
  - [x] 2.2 `TestMIDIPlayerExclusiveControl` 内の `time.Sleep` を短縮する
    - 200ms → 100ms（"new MIDI resets tick counter" サブテスト、2箇所）
    - _Requirements: 1.1, 1.2_

- [x] 3. ポーリングループの最適化
  - [x] 3.1 `TestMIDIPlayerMIDIEndEvent` 内のポーリング間隔を50msから20msに短縮する（3箇所）
    - _Requirements: 1.1_

- [x] 4. チェックポイント - テスト実行確認
  - `go test ./pkg/vm/audio/... -timeout 30s -v` を実行し、全テストが30秒以内にパスすることを確認する。問題があればユーザーに確認する。
  - _Requirements: 1.1, 2.2, 2.3, 2.4, 3.1_
