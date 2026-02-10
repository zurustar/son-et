# 実装計画: switch文エッジケーステスト

## 概要

FILLY言語のswitch-case文のエッジケーステストを追加し、`executeSwitch` の `breakSignal` 伝播問題を修正する。

## タスク

- [x] 1. `executeSwitch` の breakSignal キャッチ修正
  - [x] 1.1 `pkg/vm/vm.go` の `executeSwitch` メソッドで、case本体の `executeBlock` 結果から `breakSignal` をキャッチして消費するよう修正する
    - マッチしたcaseの `executeBlock` 結果が `breakSignal` の場合、`nil, nil` を返す
    - defaultブロックの `executeBlock` 結果も同様に処理する
    - _Requirements: 1.1, 1.2_
  - [x] 1.2 ループ内switch-breakのユニットテストを `pkg/vm/executor_test.go` に追加する
    - forループ内にswitchを配置し、case内のbreakがループを壊さないことを検証
    - _Requirements: 1.2_

- [x] 2. VM層のエッジケーステスト追加
  - [x] 2.1 プロパティテスト: フォールスルーなし
    - **Property 1: フォールスルーなし**
    - ランダムなswitch値とcase値リストを生成し、マッチしたcaseのみが実行されることを検証
    - **Validates: Requirements 1.1, 1.3**
  - [x] 2.2 プロパティテスト: ループ内breakの正確性
    - **Property 2: ループ内breakの正確性**
    - ランダムなループ回数とswitch値を生成し、ループ内switchのbreakがループを壊さないことを検証
    - **Validates: Requirements 1.2**
  - [x] 2.3 プロパティテスト: case値マッチングの正確性
    - **Property 3: case値マッチングの正確性**
    - ランダムな整数/文字列値を生成し、正しいcaseが選択されることを検証
    - **Validates: Requirements 2.1, 2.2**
  - [x] 2.4 プロパティテスト: defaultフォールバックの正確性
    - **Property 4: defaultフォールバックの正確性**
    - マッチしないswitch値を生成し、default有無に応じた動作を検証
    - **Validates: Requirements 2.3, 2.4**

- [x] 3. パーサー層のエッジケーステスト追加
  - [x] 3.1 パーサーのエッジケーステストを `pkg/compiler/parser/parser_test.go` に追加する
    - 空のcase本体のパース
    - ネストされたswitch文のパース
    - defaultの後にcaseがある場合のパース
    - case値が式の場合のパース
    - _Requirements: 3.1, 3.2, 3.3, 3.4_

- [x] 4. チェックポイント - テスト実行確認
  - `go test ./pkg/vm/... -timeout 30s -v` と `go test ./pkg/compiler/... -timeout 30s -v` を実行し、全テストがパスすることを確認する。問題があればユーザーに確認する。
  - _Requirements: 1.1, 1.2, 1.3, 2.1, 2.2, 2.3, 2.4, 2.5, 3.1, 3.2, 3.3, 3.4_

## 備考

- タスク `*` はオプション（スキップ可能）
- プロパティテストは `testing/quick` パッケージを使用
- 各プロパティテストは最低100回のイテレーションで実行
