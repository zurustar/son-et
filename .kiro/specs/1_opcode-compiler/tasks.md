# 実装タスク: OpCodeコンパイラ

## タスク概要

FILLYスクリプト（.TFYファイル）をOpCodeに変換するコンパイラを実装します。
トライアンドエラーの方針で、サンプルファイルを使って動作確認しながら進めます。

## タスク

### 1. プロジェクト構造のセットアップ

- [x] 1.1 pkg/compiler/ディレクトリ構造を作成
- [x] 1.2 基本的なパッケージファイルを作成（lexer, parser, compiler）

### 2. Token定義とLexer実装

- [x] 2.1 Token型とTokenType定数を定義（pkg/compiler/lexer/token.go）
- [x] 2.2 キーワードマップを定義（大文字小文字非依存）
- [x] 2.3 Lexer構造体と基本メソッドを実装（pkg/compiler/lexer/lexer.go）
- [x] 2.4 数値リテラル（10進数、16進数）の解析を実装
- [x] 2.5 文字列リテラルの解析を実装
- [x] 2.6 演算子と区切り文字の解析を実装
- [x] 2.7 コメント（単一行、複数行）のスキップを実装
- [x] 2.8 サンプルファイル（samples/robot/ROBOT.TFY）でLexerをテスト

### 3. AST定義とParser実装

- [x] 3.1 ASTノード型を定義（pkg/compiler/parser/ast.go）
- [x] 3.2 Parser構造体と基本メソッドを実装（pkg/compiler/parser/parser.go）
- [x] 3.3 変数宣言（int, str, 配列）の解析を実装
- [x] 3.4 関数定義（functionキーワードなし）の解析を実装
- [x] 3.5 式の解析を実装（演算子優先順位、関数呼び出し、配列アクセス）
- [x] 3.6 制御構文（if, for, while）の解析を実装
- [x] 3.7 mes文の解析を実装
- [x] 3.8 step文の解析を実装（カンマを待機命令として処理）
- [x] 3.9 サンプルファイルでParserをテスト

### 4. OpCode定義とCompiler実装

- [x] 4.1 OpCode型を定義（pkg/compiler/compiler/opcode.go）
- [x] 4.2 Compiler構造体と基本メソッドを実装（pkg/compiler/compiler/compiler.go）
- [x] 4.3 変数代入のOpCode生成を実装
- [x] 4.4 関数呼び出しのOpCode生成を実装
- [x] 4.5 制御構文のOpCode生成を実装
- [x] 4.6 mes文のOpCode生成を実装
- [x] 4.7 step文のOpCode生成を実装
- [x] 4.8 サンプルファイルでCompilerをテスト

### 5. 統合とAPI

- [x] 5.1 統合API（Compile, CompileFile）を実装（pkg/compiler/compiler.go）
- [x] 5.2 既存のscript.Loaderとの統合
- [x] 5.3 複数のサンプルファイルで統合テスト

### 6. エラーハンドリング

- [x] 6.1 CompileError型を定義
- [x] 6.2 エラーコンテキスト（ソースコード周辺）の生成を実装
- [x] 6.3 各フェーズでのエラー収集と報告を実装

### 7. アプリケーション統合

- [x] 7.1 mainエントリーポイント検出機能を実装（pkg/script/script.go）
- [x] 7.2 app.goにコンパイラ呼び出しを追加
- [x] 7.3 コンパイル結果のログ出力を実装
- [x] 7.4 コンパイルエラー時のアプリケーション終了処理を実装
- [x] 7.5 複数サンプルディレクトリでの統合テスト

### 8. プリプロセッサディレクティブ対応

- [x] 8.1 Lexerに#info, #includeディレクティブのトークン化を追加
- [x] 8.2 ParserにInfoDirective, IncludeDirectiveノードを追加
- [x] 8.3 #infoディレクティブの解析を実装（行末まで読み込み）
- [x] 8.4 #includeディレクティブの解析を実装（ファイル名抽出）
- [x] 8.5 サンプルファイルでディレクティブ解析をテスト

### 9. 配列参照構文対応

- [x] 9.1 ParserにArrayReferenceノードを追加
- [x] 9.2 関数呼び出し引数での配列参照（arr[]）の解析を実装
- [x] 9.3 サンプルファイルで配列参照構文をテスト

### 10. タイトルメタデータ抽出

- [x] 10.1 TitleMetadata構造体を定義（pkg/title/title.go）
- [x] 10.2 ExtractMetadata関数を実装（Lexerのみ使用）
- [x] 10.3 ExtractMetadataFromDirectory関数を実装
- [x] 10.4 FillyTitleにMetadataフィールドを追加
- [x] 10.5 DisplayName()メソッドを実装
- [x] 10.6 サンプルディレクトリでメタデータ抽出をテスト

### 11. プリプロセッサ実装

- [x] 11.1 Preprocessor構造体を定義（pkg/compiler/preprocessor/preprocessor.go）
- [x] 11.2 #includeディレクティブの検出と展開を実装
- [x] 11.3 循環参照検出を実装
- [x] 11.4 インクルードガード（重複防止）を実装
- [x] 11.5 エントリーポイントからの依存ファイル解決を実装
- [x] 11.6 PreprocessFile関数を実装（ファイルパスから展開済みソースを返す）
- [x] 11.7 サンプルファイルでプリプロセッサをテスト

### 12. エントリーポイントベースのコンパイル

- [x] 12.1 CompileWithEntryPointを修正（プリプロセッサを使用）
- [x] 12.2 エントリーポイントから到達可能なファイルのみコンパイル
- [x] 12.3 sab2ディレクトリでエントリーポイント指定コンパイルをテスト
