# 実装計画: missing-builtin-functions

## 概要

ROBOTサンプル実行時のエラーを修正するため、StrPrint関数、CreatePic 3引数パターン、CapTitle 1引数パターンを実装する。

## タスク

- [x] 1. StrPrint 関数の実装
  - [x] 1.1 pkg/vm/vm.go に StrPrint 組み込み関数を登録
    - フォーマット文字列の変換処理（%ld→%d, %lx→%x）
    - エスケープシーケンスの変換（\n, \t, \r）
    - fmt.Sprintf を使用したフォーマット実行
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 1.7, 1.8_
  - [x] 1.2 StrPrint のユニットテストを作成
    - 各フォーマット指定子のテスト
    - エスケープシーケンスのテスト
    - エッジケース（引数不足・過剰）のテスト
    - _Requirements: 1.1-1.8_
  - [x] 1.3 StrPrint のプロパティテストを作成
    - **Property 1: StrPrint フォーマット変換の正確性**
    - **Validates: Requirements 1.1, 1.2, 1.3, 1.4, 1.5**

- [x] 2. CreatePic 3引数パターンの実装
  - [x] 2.1 pkg/graphics/picture.go に CreatePicWithSize メソッドを追加
    - ソースピクチャーIDの存在確認
    - 幅・高さのバリデーション
    - 空のピクチャー作成
    - _Requirements: 2.1, 2.2, 2.3, 2.4_
  - [x] 2.2 pkg/graphics/graphics.go に CreatePicWithSize を追加
    - GraphicsSystem のラッパーメソッド
    - _Requirements: 2.1_
  - [x] 2.3 pkg/vm/vm.go の GraphicsSystemInterface に CreatePicWithSize を追加
    - インターフェース定義の更新
    - _Requirements: 2.1_
  - [x] 2.4 pkg/vm/vm.go の CreatePic 組み込み関数を3引数パターンに対応
    - 引数の数に応じた分岐処理
    - _Requirements: 2.1, 2.5_
  - [x] 2.5 CreatePic 3引数パターンのユニットテストを作成
    - 正常系テスト
    - 異常系テスト（存在しないID、不正なサイズ）
    - _Requirements: 2.1-2.5_
  - [x] 2.6 CreatePic 3引数パターンのプロパティテストを作成
    - **Property 2: CreatePic 3引数パターンのサイズと内容**
    - **Validates: Requirements 2.1, 2.2**

- [x] 3. CapTitle 1引数パターンの修正
  - [x] 3.1 pkg/graphics/window.go に CapTitleAll メソッドを追加
    - 全ウィンドウのキャプション設定
    - ウィンドウが存在しない場合の処理
    - _Requirements: 3.1, 3.2_
  - [x] 3.2 pkg/graphics/graphics.go に CapTitleAll を追加
    - GraphicsSystem のラッパーメソッド
    - _Requirements: 3.1_
  - [x] 3.3 pkg/vm/vm.go の GraphicsSystemInterface に CapTitleAll を追加
    - インターフェース定義の更新
    - _Requirements: 3.1_
  - [x] 3.4 pkg/vm/vm.go の CapTitle 組み込み関数を1引数パターンに対応
    - 引数の数に応じた分岐処理
    - 1引数の場合は CapTitleAll を呼び出し
    - 2引数の場合は既存の CapTitle を呼び出し（エラー無視）
    - _Requirements: 3.1, 3.3, 3.4, 3.5_
  - [x] 3.5 CapTitle のユニットテストを作成
    - 1引数パターンのテスト
    - 2引数パターンのテスト
    - エッジケース（ウィンドウなし、存在しないID）のテスト
    - _Requirements: 3.1-3.5_
  - [x] 3.6 CapTitle のプロパティテストを作成
    - **Property 3: CapTitle 1引数パターンで全ウィンドウ更新**
    - **Property 4: CapTitle 2引数パターンで特定ウィンドウのみ更新**
    - **Validates: Requirements 3.1, 3.3**

- [x] 4. チェックポイント - 全テストの実行
  - 全てのテストが通ることを確認し、問題があればユーザーに確認する。

- [x] 5. ROBOTサンプルでの動作確認
  - [x] 5.1 samples/robot/ROBOT.TFY を実行して動作確認
    - StrPrint によるファイル名生成が正しく動作することを確認
    - CreatePic 3引数パターンが正しく動作することを確認
    - CapTitle が正しく動作することを確認
    - _Requirements: 1.1, 2.1, 3.1_

## 備考

- 全てのタスクは必須です
- 各プロパティテストは最低100回のイテレーションを実行します
- テストには `testing/quick` パッケージを使用します
- go test コマンドには `-timeout` オプションを使用して実行します
