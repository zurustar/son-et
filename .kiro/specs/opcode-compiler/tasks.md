# 実装タスク: OpCodeコンパイラ

## タスク概要

FILLYスクリプト（.TFYファイル）をOpCodeに変換するコンパイラを実装します。
トライアンドエラーの方針で、サンプルファイルを使って動作確認しながら進めます。

## タスク

### 1. プロジェクト構造のセットアップ

- [ ] 1.1 pkg/compiler/ディレクトリ構造を作成
- [ ] 1.2 基本的なパッケージファイルを作成（lexer, parser, compiler）

### 2. Token定義とLexer実装

- [ ] 2.1 Token型とTokenType定数を定義（pkg/compiler/lexer/token.go）
- [ ] 2.2 キーワードマップを定義（大文字小文字非依存）
- [ ] 2.3 Lexer構造体と基本メソッドを実装（pkg/compiler/lexer/lexer.go）
- [ ] 2.4 数値リテラル（10進数、16進数）の解析を実装
- [ ] 2.5 文字列リテラルの解析を実装
- [ ] 2.6 演算子と区切り文字の解析を実装
- [ ] 2.7 コメント（単一行、複数行）のスキップを実装
- [ ] 2.8 サンプルファイル（samples/robot/ROBOT.TFY）でLexerをテスト

### 3. AST定義とParser実装

- [ ] 3.1 ASTノード型を定義（pkg/compiler/parser/ast.go）
- [ ] 3.2 Parser構造体と基本メソッドを実装（pkg/compiler/parser/parser.go）
- [ ] 3.3 変数宣言（int, str, 配列）の解析を実装
- [ ] 3.4 関数定義（functionキーワードなし）の解析を実装
- [ ] 3.5 式の解析を実装（演算子優先順位、関数呼び出し、配列アクセス）
- [ ] 3.6 制御構文（if, for, while）の解析を実装
- [ ] 3.7 mes文の解析を実装
- [ ] 3.8 step文の解析を実装（カンマを待機命令として処理）
- [ ] 3.9 サンプルファイルでParserをテスト

### 4. OpCode定義とCompiler実装

- [ ] 4.1 OpCode型を定義（pkg/compiler/compiler/opcode.go）
- [ ] 4.2 Compiler構造体と基本メソッドを実装（pkg/compiler/compiler/compiler.go）
- [ ] 4.3 変数代入のOpCode生成を実装
- [ ] 4.4 関数呼び出しのOpCode生成を実装
- [ ] 4.5 制御構文のOpCode生成を実装
- [ ] 4.6 mes文のOpCode生成を実装
- [ ] 4.7 step文のOpCode生成を実装
- [ ] 4.8 サンプルファイルでCompilerをテスト

### 5. 統合とAPI

- [ ] 5.1 統合API（Compile, CompileFile）を実装（pkg/compiler/compiler.go）
- [ ] 5.2 既存のscript.Loaderとの統合
- [ ] 5.3 複数のサンプルファイルで統合テスト

### 6. エラーハンドリング

- [ ] 6.1 CompileError型を定義
- [ ] 6.2 エラーコンテキスト（ソースコード周辺）の生成を実装
- [ ] 6.3 各フェーズでのエラー収集と報告を実装
