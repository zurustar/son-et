# コンパイラ

## 概要

FILLYスクリプト（.TFYファイル）をOpCodeに変換するコンパイラの設計について説明します。
コンパイラは、プリプロセッサ、字句解析（Lexer）、構文解析（Parser）、OpCode生成（Compiler）の4フェーズからなるパイプラインで構成されています。

---

## 1. コンパイラパイプライン

### パイプライン全体図

```
TFYファイル (Shift-JIS)
       │
       ▼
┌──────────────────┐
│  script.Loader   │  Shift-JIS → UTF-8 変換 & ファイル読み込み
└──────────────────┘
       │
       ▼
ソースコード (UTF-8)
       │
       ▼
┌──────────────────┐
│  Preprocessor    │  #include 展開、#info 抽出
└──────────────────┘
       │
       ▼
展開済みソースコード
       │
       ▼
┌──────────────────┐
│     Lexer        │  トークン化（字句解析）
└──────────────────┘
       │
       ▼
    Token列
       │
       ▼
┌──────────────────┐
│     Parser       │  構文解析 → AST構築
└──────────────────┘
       │
       ▼
      AST
       │
       ▼
┌──────────────────┐
│    Compiler      │  OpCode生成
└──────────────────┘
       │
       ▼
   OpCode列
```

### 各フェーズの役割

| フェーズ | 入力 | 出力 | 役割 |
|---|---|---|---|
| **script.Loader** | TFYファイル（Shift-JIS） | UTF-8ソースコード | ファイル読み込みとエンコーディング変換 |
| **Preprocessor** | UTF-8ソースコード | 展開済みソースコード | `#include` の再帰展開、`#info` メタデータ抽出 |
| **Lexer** | 展開済みソースコード | Token列 | ソースコードをトークンに分解 |
| **Parser** | Token列 | AST（抽象構文木） | トークン列を構造化された木構造に変換 |
| **Compiler** | AST | OpCode列 | ASTからVM実行可能な命令列を生成 |

### エントリーポイントの解決

複数のTFYファイルが存在する場合、以下の手順でエントリーポイントを決定します。

1. ユーザーが明示的にエントリーポイントファイルを指定した場合、そのファイルを使用
2. 指定がない場合、`main()` 関数を含むファイルを自動検出
3. `main()` が複数ファイルに存在する場合、エラーを報告
4. `main()` が見つからない場合、エラーを報告
5. パースエラーが発生したファイルはスキップして検出を続行

エントリーポイントから `#include` されるファイルのみがコンパイル対象となり、インクルードされていないファイルは無視されます。

### プリプロセッサの詳細

プリプロセッサは以下の処理を行います。

| 機能 | 説明 |
|---|---|
| `#include "filename"` 展開 | 指定ファイルの内容を再帰的に展開 |
| `#info` 抽出 | INAM（タイトル名）、ICOP（著作権）、ISBJ（説明）、ICMT（コメント）を抽出 |
| 循環参照検出 | `#include` の循環参照を検出してエラー報告 |
| インクルードガード | 同じファイルの重複インクルードを防止 |

### パッケージ構成

```
pkg/compiler/
├── preprocessor.go    # プリプロセッサ（#include展開）
├── lexer/
│   ├── lexer.go       # Lexer実装
│   └── token.go       # Token定義
├── parser/
│   ├── parser.go      # Parser実装
│   └── ast.go         # AST定義
├── compiler/
│   ├── compiler.go    # Compiler実装
│   └── opcode.go      # OpCode定義
└── compiler.go        # 統合API（Compile, CompileString, CompileWithOptions）
```

---

## 2. パース上の特殊な判断

FILLYの構文仕様については [言語仕様](language-spec.md) を参照してください。
ここでは、パーサーが特殊な判断を必要とする箇所について説明します。

### 2.1 大文字小文字の処理

構文の詳細は [言語仕様 - 基本構文](language-spec.md#基本構文) を参照。

Lexerでは識別子を元の形式で保持しつつ、キーワード判定時にのみ小文字に変換します。
これにより、エラーメッセージ等でユーザーが記述した元の表記を表示できます。

### 2.2 関数定義と関数呼び出しの判別

FILLYでは `function` キーワードなしで関数を定義するため（[言語仕様 - 関数定義](language-spec.md#関数定義) 参照）、パーサーは `識別子(` の形式を見たときに関数定義か関数呼び出しかを判別する必要があります。

判別基準:
- パラメータリスト内に型指定（`int`, `str`）やデフォルト値（`=`）がある場合 → 関数定義
- `)` の後に `{` が続く場合 → 関数定義
- それ以外 → 関数呼び出し

### 2.3 mes() / step() のパース

構文の詳細は [言語仕様 - mes ブロック](language-spec.md#mes-ブロック) および [言語仕様 - step ブロック](language-spec.md#step-ブロック) を参照。

`mes()` と `step()` はキーワードとして認識され、通常の関数呼び出しとは異なるパスで処理されます。
`step()` ブロック内のカンマ（`;,`）はLexerレベルで `Wait` トークンとして認識され、OpCode生成時に `Wait` 命令に変換されます（§3 の [step文のOpCode変換](#step文のopcode変換) を参照）。

### 2.4 変数宣言と代入の判別

構文の詳細は [言語仕様 - 変数宣言](language-spec.md#変数宣言) を参照。

パーサーは以下の基準で変数宣言と代入を判別します:
- 型キーワード（`int`, `str`）で始まる場合 → 変数宣言
- `識別子 = 値` の形式 → 型指定なしの代入（動的に変数を作成）

### 2.5 関数パラメータの判別

構文の詳細は [言語仕様 - 関数定義](language-spec.md#関数定義) を参照。

パーサーは関数定義のパラメータリストで以下を判別します:
- デフォルト値の検出: パラメータ名の後に `=` がある場合
- 配列パラメータの判別: パラメータ名の後に `[]` がある場合
- 配列参照渡し（呼び出し側）: 引数に `識別子[]` の形式がある場合

### 2.6 プリプロセッサディレクティブ

[§1 のプリプロセッサの詳細](#プリプロセッサの詳細) を参照。
`#info` のフィールド一覧は [言語仕様 - プリプロセッサディレクティブ](language-spec.md#プリプロセッサディレクティブ) を参照。

### 2.7 del_me / del_us / del_all のパース

動作の詳細は [言語仕様 - del_me](language-spec.md#del_me) および [実行モデル - del_me / del_us / del_all の挙動](execution-model.md#del_me--del_us--del_all-の挙動) を参照。

これらはキーワードですが、パーサーでは関数呼び出し（`Call` OpCode）として扱います。
引数なしの `Call` として処理され、VM側で特殊な動作を実行します。

---

## 3. OpCode一覧と引数仕様

### OpCode構造体

```go
type OpCode struct {
    Cmd  OpCmd    // コマンドタイプ（文字列）
    Args []any   // 引数（任意の型: 整数、文字列、Variable、ネストされたOpCode）
}

// 変数参照を区別するための型
type Variable string
```

### OpCode一覧

| OpCode | 説明 | Args仕様 |
|---|---|---|
| `Assign` | 変数代入 | `[Variable(変数名), 値]` |
| `ArrayAssign` | 配列要素代入 | `[Variable(配列名), インデックス, 値]` |
| `ArrayAccess` | 配列要素アクセス | `[Variable(配列名), インデックス]` |
| `Call` | 関数呼び出し | `[関数名(string), 引数1, 引数2, ...]` |
| `BinaryOp` | 二項演算 | `[演算子(string), 左辺, 右辺]` |
| `UnaryOp` | 単項演算 | `[演算子(string), オペランド]` |
| `If` | 条件分岐 | `[条件(OpCode), then節([]OpCode), else節([]OpCode)]` |
| `For` | forループ | `[初期化([]OpCode), 条件(OpCode), 後処理([]OpCode), 本体([]OpCode)]` |
| `While` | whileループ | `[条件(OpCode), 本体([]OpCode)]` |
| `Switch` | switch文 | `[値(OpCode), case群, default([]OpCode)]` |
| `Break` | ループ脱出 | `[]`（引数なし） |
| `Continue` | 次のループ反復 | `[]`（引数なし） |
| `RegisterEventHandler` | イベントハンドラ登録 | `[イベントタイプ(string), 本体([]OpCode)]` |
| `Wait` | 待機命令 | `[待機カウント(int)]` |
| `SetStep` | ステップ期間設定 | `[ステップカウント(int)]` |

### 引数の型

OpCodeの引数には以下の型が使用されます。

| 型 | 説明 | 例 |
|---|---|---|
| `int` / `int64` | 整数リテラル | `5`, `0xFF` |
| `float64` | 浮動小数点リテラル | `3.14` |
| `string` | 文字列リテラル / 関数名 / 演算子 | `"image.bmp"`, `"LoadPic"`, `"+"` |
| `Variable` | 変数参照 | `Variable("x")`, `Variable("arr")` |
| `OpCode` | ネストされた命令 | 式の評価結果 |
| `[]OpCode` | 命令シーケンス | ブロック本体 |

### サポートされる演算子

#### 二項演算子（BinaryOp）

| 演算子 | 説明 | 優先順位 |
|---|---|---|
| `\|\|` | 論理OR | 1（最低） |
| `&&` | 論理AND | 2 |
| `==`, `!=` | 等値比較 | 3 |
| `<`, `>`, `<=`, `>=` | 大小比較 | 4 |
| `+`, `-` | 加算・減算 | 5 |
| `*`, `/`, `%` | 乗算・除算・剰余 | 6 |

#### 単項演算子（UnaryOp）

| 演算子 | 説明 |
|---|---|
| `-` | 符号反転 |
| `!` | 論理否定 |

### OpCode生成例

#### 変数代入

```toffy
x = 5 + 3
```

```go
OpCode{
    Cmd: "Assign",
    Args: []any{
        Variable("x"),
        OpCode{Cmd: "BinaryOp", Args: []any{"+", 5, 3}},
    },
}
```

#### 配列代入

```toffy
arr[i] = value
```

```go
OpCode{
    Cmd: "ArrayAssign",
    Args: []any{Variable("arr"), Variable("i"), Variable("value")},
}
```

#### 関数呼び出し

```toffy
LoadPic("image.bmp")
```

```go
OpCode{
    Cmd: "Call",
    Args: []any{"LoadPic", "image.bmp"},
}
```

#### if文

```toffy
if (x > 5) { y = 10 } else { y = 0 }
```

```go
OpCode{
    Cmd: "If",
    Args: []any{
        OpCode{Cmd: "BinaryOp", Args: []any{">", Variable("x"), 5}},
        []OpCode{{Cmd: "Assign", Args: []any{Variable("y"), 10}}},
        []OpCode{{Cmd: "Assign", Args: []any{Variable("y"), 0}}},
    },
}
```

#### step文のOpCode変換

```toffy
step(10) {
    func1();,
    func2();,,
    end_step;
    del_me;
}
```

```go
// フラットなOpCodeシーケンスに変換される
[]OpCode{
    {Cmd: "SetStep", Args: []any{10}},
    {Cmd: "Call", Args: []any{"func1"}},
    {Cmd: "Wait", Args: []any{1}},     // カンマ1つ → 1ステップ待機
    {Cmd: "Call", Args: []any{"func2"}},
    {Cmd: "Wait", Args: []any{2}},     // カンマ2つ → 2ステップ待機
    {Cmd: "Call", Args: []any{"del_me"}},
}
```

#### mes文

```toffy
mes(MIDI_TIME) { step(8) { ... } }
```

```go
OpCode{
    Cmd: "RegisterEventHandler",
    Args: []any{
        "MIDI_TIME",
        []OpCode{...},  // ボディのOpCode
    },
}
```

---

## 4. エラーハンドリング方針

### エラー構造体

すべてのコンパイルエラーは統一されたフォーマットで報告されます。

```go
type CompileError struct {
    Phase   string // "lexer", "parser", "compiler"
    Message string // エラーメッセージ
    Line    int    // 行番号
    Column  int    // 列番号
    Context string // エラー周辺のソースコード
}
```

### フェーズ別エラー

| フェーズ | エラー種別 | 報告内容 |
|---|---|---|
| **Lexer** | 不正な文字 | 文字、行番号、列番号 |
| **Parser** | 構文エラー | 期待されるトークン、実際のトークン、行番号、列番号 |
| **Compiler** | 未知のASTノード | ノードタイプ、エラーメッセージ |
| **Preprocessor** | ファイル未検出 | ファイルパス |
| **Preprocessor** | 循環参照 | 循環しているファイル名 |

### エラーコンテキスト表示

構文エラーが発生した場合、エラー行の前後2行のソースコードコンテキストと、エラー位置を示すポインタ（`^`）が表示されます。

```
parser error at line 5, column 12: expected ')', got ';'
    3 |
    4 | main() {
  > 5 |     func(x;y)
       |            ^
    6 | }
    7 |
```

### エラー収集方針

- 各フェーズはエラーを検出しても可能な限り処理を継続し、複数のエラーを収集する
- いずれかのフェーズが失敗した場合、パイプラインは停止し蓄積されたすべてのエラーを返す
- コンパイルが成功した場合、空のエラーリストを返す
- すべてのトークンにはエラー報告のために行番号と列番号が記録される

### 統合API

```go
// ソースコード文字列からコンパイル
func CompileString(source string) ([]OpCode, []error)

// ファイルパスからコンパイル（script.Loaderと統合）
func Compile(path string) ([]OpCode, []error)

// オプション付きコンパイル
func CompileWithOptions(source string, opts CompileOptions) ([]OpCode, []error)
```

パイプラインは常に Lexer → Parser → Compiler の順序で実行され、前のフェーズが失敗した場合は後続のフェーズは実行されません。
