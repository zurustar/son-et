# 設計書: OpCodeコンパイラ

## はじめに

このドキュメントは、FILLYスクリプト（.TFYファイル）をOpCodeに変換するコンパイラの設計を定義します。要件定義書（requirements.md）に基づき、字句解析、構文解析、OpCode生成の3フェーズからなるコンパイラパイプラインを設計します。

## 用語集

requirements.mdの用語集を参照してください。

---

## アーキテクチャ概要

### コンパイルパイプライン

```
ソースコード (UTF-8)
       │
       ▼
┌─────────────┐
│   Lexer     │  トークン化
└─────────────┘
       │
       ▼
   Token列
       │
       ▼
┌─────────────┐
│   Parser    │  構文解析
└─────────────┘
       │
       ▼
     AST
       │
       ▼
┌─────────────┐
│  Compiler   │  OpCode生成
└─────────────┘
       │
       ▼
   OpCode列
```

### パッケージ構成

```
pkg/compiler/
├── lexer/
│   ├── lexer.go      # Lexer実装
│   ├── token.go      # Token定義
│   └── lexer_test.go
├── parser/
│   ├── parser.go     # Parser実装
│   ├── ast.go        # AST定義
│   └── parser_test.go
├── compiler/
│   ├── compiler.go   # Compiler実装
│   ├── opcode.go     # OpCode定義
│   └── compiler_test.go
└── compiler.go       # 統合API
```

---

## Part 1: 字句解析（Lexer）

### Token定義

```go
type TokenType int

const (
    // 特殊トークン
    TOKEN_ILLEGAL TokenType = iota
    TOKEN_EOF
    TOKEN_COMMENT

    // リテラル
    TOKEN_IDENT      // 識別子
    TOKEN_INT        // 整数リテラル
    TOKEN_FLOAT      // 浮動小数点リテラル
    TOKEN_STRING     // 文字列リテラル

    // プリプロセッサディレクティブ
    TOKEN_DIRECTIVE  // #info, #include など
    TOKEN_INFO       // #info
    TOKEN_INCLUDE    // #include

    // 演算子
    TOKEN_PLUS       // +
    TOKEN_MINUS      // -
    TOKEN_ASTERISK   // *
    TOKEN_SLASH      // /
    TOKEN_PERCENT    // %
    TOKEN_ASSIGN     // =
    TOKEN_EQ         // ==
    TOKEN_NEQ        // !=
    TOKEN_LT         // <
    TOKEN_GT         // >
    TOKEN_LTE        // <=
    TOKEN_GTE        // >=
    TOKEN_AND        // &&
    TOKEN_OR         // ||
    TOKEN_NOT        // !

    // 区切り文字
    TOKEN_LPAREN     // (
    TOKEN_RPAREN     // )
    TOKEN_LBRACE     // {
    TOKEN_RBRACE     // }
    TOKEN_LBRACKET   // [
    TOKEN_RBRACKET   // ]
    TOKEN_COMMA      // ,
    TOKEN_SEMICOLON  // ;
    TOKEN_COLON      // : (for case labels)

    // キーワード
    TOKEN_INT_TYPE   // int
    TOKEN_STR_TYPE   // str
    TOKEN_IF         // if
    TOKEN_ELSE       // else
    TOKEN_FOR        // for
    TOKEN_WHILE      // while
    TOKEN_SWITCH     // switch
    TOKEN_CASE       // case
    TOKEN_DEFAULT    // default
    TOKEN_BREAK      // break
    TOKEN_CONTINUE   // continue
    TOKEN_RETURN     // return
    TOKEN_MES        // mes
    TOKEN_STEP       // step
    TOKEN_END_STEP   // end_step
    TOKEN_DEL_ME     // del_me
    TOKEN_DEL_US     // del_us
    TOKEN_DEL_ALL    // del_all
)

type Token struct {
    Type    TokenType
    Literal string
    Line    int
    Column  int
}
```

### キーワードマップ（大文字小文字を区別しない）

```go
var keywords = map[string]TokenType{
    "int":      TOKEN_INT_TYPE,
    "str":      TOKEN_STR_TYPE,
    "if":       TOKEN_IF,
    "else":     TOKEN_ELSE,
    "for":      TOKEN_FOR,
    "while":    TOKEN_WHILE,
    "switch":   TOKEN_SWITCH,
    "case":     TOKEN_CASE,
    "default":  TOKEN_DEFAULT,
    "break":    TOKEN_BREAK,
    "continue": TOKEN_CONTINUE,
    "return":   TOKEN_RETURN,
    "mes":      TOKEN_MES,
    "step":     TOKEN_STEP,
    "end_step": TOKEN_END_STEP,
    "del_me":   TOKEN_DEL_ME,
    "del_us":   TOKEN_DEL_US,
    "del_all":  TOKEN_DEL_ALL,
}

// LookupIdent は識別子がキーワードかどうかを判定（大文字小文字を区別しない）
func LookupIdent(ident string) TokenType {
    if tok, ok := keywords[strings.ToLower(ident)]; ok {
        return tok
    }
    return TOKEN_IDENT
}
```

### Lexer構造体

```go
type Lexer struct {
    input        string
    position     int  // 現在の位置
    readPosition int  // 次に読む位置
    ch           byte // 現在の文字
    line         int  // 現在の行番号
    column       int  // 現在の列番号
}

func New(input string) *Lexer
func (l *Lexer) NextToken() Token
func (l *Lexer) Tokenize() ([]Token, error)
```

### 数値リテラルの解析

```go
// 整数リテラル: 10進数または16進数（0x/0Xプレフィックス）
func (l *Lexer) readNumber() Token {
    // 0xで始まる場合は16進数
    if l.ch == '0' && (l.peekChar() == 'x' || l.peekChar() == 'X') {
        return l.readHexNumber()
    }
    // それ以外は10進数（小数点があればfloat）
    return l.readDecimalNumber()
}
```

---

## Part 2: 構文解析（Parser）

### AST定義

```go
// ノードインターフェース
type Node interface {
    TokenLiteral() string
}

type Statement interface {
    Node
    statementNode()
}

type Expression interface {
    Node
    expressionNode()
}

// プログラム（ルートノード）
type Program struct {
    Statements []Statement
}

// 変数宣言
type VarDeclaration struct {
    Token    Token      // int または str
    Type     string     // "int" または "str"
    Names    []string   // 変数名のリスト
    IsArray  []bool     // 各変数が配列かどうか
    Sizes    []Expression // 配列サイズ（指定がある場合）
}

// プリプロセッサディレクティブ: #info
type InfoDirective struct {
    Token   Token
    Key     string // INAM, ISBJ, IART, etc.
    Value   string // ディレクティブの値
}

// プリプロセッサディレクティブ: #include
type IncludeDirective struct {
    Token    Token
    FileName string // インクルードするファイル名
}

// 配列参照（関数引数として配列全体を渡す）
// 例: func(arr[])
type ArrayReference struct {
    Token Token
    Name  string // 配列名
}

// 関数定義
type FunctionStatement struct {
    Token      Token
    Name       string
    Parameters []Parameter
    Body       *BlockStatement
}

type Parameter struct {
    Name         string
    Type         string     // "int", "str", または ""（型指定なし）
    IsArray      bool
    DefaultValue Expression // デフォルト値（オプション）
}

// ブロック文
type BlockStatement struct {
    Token      Token
    Statements []Statement
}

// 代入文
type AssignStatement struct {
    Token Token
    Name  Expression // Identifier または IndexExpression
    Value Expression
}

// 式文
type ExpressionStatement struct {
    Token      Token
    Expression Expression
}

// if文
type IfStatement struct {
    Token       Token
    Condition   Expression
    Consequence *BlockStatement
    Alternative Statement // *BlockStatement または *IfStatement（else if）
}

// forループ
type ForStatement struct {
    Token     Token
    Init      Statement
    Condition Expression
    Post      Statement
    Body      *BlockStatement
}

// whileループ
type WhileStatement struct {
    Token     Token
    Condition Expression
    Body      *BlockStatement
}

// switch文
type SwitchStatement struct {
    Token   Token
    Value   Expression
    Cases   []*CaseClause
    Default *BlockStatement
}

type CaseClause struct {
    Token Token
    Value Expression
    Body  []Statement
}

// mes文
type MesStatement struct {
    Token     Token
    EventType string // TIME, MIDI_TIME, MIDI_END, KEY, CLICK, etc.
    Body      *BlockStatement
}

// step文
type StepStatement struct {
    Token Token
    Count Expression // step(n)のn、省略時はnil
    Body  *StepBody
}

type StepBody struct {
    Commands []StepCommand
}

type StepCommand struct {
    Statement Statement // 実行する文（nilの場合はwait）
    WaitCount int       // 連続するカンマの数
}

// 式ノード
type Identifier struct {
    Token Token
    Value string
}

type IntegerLiteral struct {
    Token Token
    Value int64
}

type FloatLiteral struct {
    Token Token
    Value float64
}

type StringLiteral struct {
    Token Token
    Value string
}

type BinaryExpression struct {
    Token    Token
    Left     Expression
    Operator string
    Right    Expression
}

type UnaryExpression struct {
    Token    Token
    Operator string
    Right    Expression
}

type CallExpression struct {
    Token     Token
    Function  string
    Arguments []Expression
}

type IndexExpression struct {
    Token Token
    Left  Expression // 配列
    Index Expression // インデックス
}
```

### Parser構造体

```go
type Parser struct {
    lexer     *lexer.Lexer
    tokens    []lexer.Token
    pos       int
    errors    []string
    
    // 演算子優先順位
    prefixParseFns map[lexer.TokenType]prefixParseFn
    infixParseFns  map[lexer.TokenType]infixParseFn
}

func New(l *lexer.Lexer) *Parser
func (p *Parser) ParseProgram() (*Program, []error)
```

### 演算子優先順位

```go
const (
    _ int = iota
    LOWEST
    OR          // ||
    AND         // &&
    EQUALS      // ==, !=
    LESSGREATER // <, >, <=, >=
    SUM         // +, -
    PRODUCT     // *, /, %
    PREFIX      // -x, !x
    CALL        // func(x)
    INDEX       // array[index]
)

var precedences = map[lexer.TokenType]int{
    lexer.TOKEN_OR:       OR,
    lexer.TOKEN_AND:      AND,
    lexer.TOKEN_EQ:       EQUALS,
    lexer.TOKEN_NEQ:      EQUALS,
    lexer.TOKEN_LT:       LESSGREATER,
    lexer.TOKEN_GT:       LESSGREATER,
    lexer.TOKEN_LTE:      LESSGREATER,
    lexer.TOKEN_GTE:      LESSGREATER,
    lexer.TOKEN_PLUS:     SUM,
    lexer.TOKEN_MINUS:    SUM,
    lexer.TOKEN_ASTERISK: PRODUCT,
    lexer.TOKEN_SLASH:    PRODUCT,
    lexer.TOKEN_PERCENT:  PRODUCT,
    lexer.TOKEN_LPAREN:   CALL,
    lexer.TOKEN_LBRACKET: INDEX,
}
```

### 関数定義の解析

FILLYでは`function`キーワードなしで関数を定義できます：

```go
// 関数定義の判定
// name(params){body} または name(){body}
func (p *Parser) isFunctionDefinition() bool {
    // 識別子の後に ( があり、その後に ) { が続くかチェック
    // パラメータリストの中身は型指定やデフォルト値を含む可能性がある
}

// パラメータの解析
// 例: int x, y[], str s, p[], c[], x=10, int time=1
func (p *Parser) parseParameters() []Parameter {
    // 型指定あり: int x, str s
    // 配列: arr[], int arr[]
    // デフォルト値: x=10, int time=1
    // 型指定なし: p, c
}
```

### step文の解析

step文は特殊な構文を持ちます：

```go
// step(n){commands} または step{commands}
// commands内のカンマは待機命令として解釈
func (p *Parser) parseStepStatement() *StepStatement {
    // step(10){...} - 10ステップ単位
    // step{...} - デフォルト（16分音符）
    
    // ボディ内の解析:
    // - 文の後のカンマは1ステップ待機
    // - 連続するカンマは複数ステップ待機
    // - end_stepでブロック終了
}
```

---

## Part 3: OpCode生成（Compiler）

### OpCode定義

```go
type OpCmd string

const (
    OpAssign              OpCmd = "Assign"
    OpArrayAssign         OpCmd = "ArrayAssign"
    OpCall                OpCmd = "Call"
    OpBinaryOp            OpCmd = "BinaryOp"
    OpUnaryOp             OpCmd = "UnaryOp"
    OpArrayAccess         OpCmd = "ArrayAccess"
    OpIf                  OpCmd = "If"
    OpFor                 OpCmd = "For"
    OpWhile               OpCmd = "While"
    OpSwitch              OpCmd = "Switch"
    OpBreak               OpCmd = "Break"
    OpContinue            OpCmd = "Continue"
    OpRegisterEventHandler OpCmd = "RegisterEventHandler"
    OpWait                OpCmd = "Wait"
    OpSetStep             OpCmd = "SetStep"
)

type OpCode struct {
    Cmd  OpCmd
    Args []any
}

// 変数参照を区別するための型
type Variable string
```

### Compiler構造体

```go
type Compiler struct {
    errors []string
}

func New() *Compiler
func (c *Compiler) Compile(program *ast.Program) ([]OpCode, []error)
func (c *Compiler) compileStatement(stmt ast.Statement) []OpCode
func (c *Compiler) compileExpression(expr ast.Expression) any
```

### OpCode生成例

#### 変数代入

```go
// x = 5 + 3
OpCode{
    Cmd: OpAssign,
    Args: []any{
        Variable("x"),
        OpCode{
            Cmd: OpBinaryOp,
            Args: []any{"+", 5, 3},
        },
    },
}
```

#### 配列代入

```go
// arr[i] = value
OpCode{
    Cmd: OpArrayAssign,
    Args: []any{
        Variable("arr"),
        Variable("i"),
        Variable("value"),
    },
}
```

#### 関数呼び出し

```go
// LoadPic("image.bmp")
OpCode{
    Cmd: OpCall,
    Args: []any{
        "LoadPic",
        "image.bmp",
    },
}

// MovePic(src, 0, 0, 100, 100, dst, 0, 0)
OpCode{
    Cmd: OpCall,
    Args: []any{
        "MovePic",
        Variable("src"),
        0, 0, 100, 100,
        Variable("dst"),
        0, 0,
    },
}
```

#### if文

```go
// if (x > 5) { y = 10 } else { y = 0 }
OpCode{
    Cmd: OpIf,
    Args: []any{
        // 条件
        OpCode{Cmd: OpBinaryOp, Args: []any{">", Variable("x"), 5}},
        // 結果（then節）
        []OpCode{
            {Cmd: OpAssign, Args: []any{Variable("y"), 10}},
        },
        // 代替（else節）
        []OpCode{
            {Cmd: OpAssign, Args: []any{Variable("y"), 0}},
        },
    },
}
```

#### forループ

```go
// for(i=0; i<10; i=i+1) { ... }
OpCode{
    Cmd: OpFor,
    Args: []any{
        // 初期化
        []OpCode{{Cmd: OpAssign, Args: []any{Variable("i"), 0}}},
        // 条件
        OpCode{Cmd: OpBinaryOp, Args: []any{"<", Variable("i"), 10}},
        // 後処理
        []OpCode{{Cmd: OpAssign, Args: []any{
            Variable("i"),
            OpCode{Cmd: OpBinaryOp, Args: []any{"+", Variable("i"), 1}},
        }}},
        // ボディ
        []OpCode{...},
    },
}
```

#### mes文

```go
// mes(MIDI_TIME) { step { ... } }
OpCode{
    Cmd: OpRegisterEventHandler,
    Args: []any{
        "MIDI_TIME",
        []OpCode{...}, // ボディのOpCode
    },
}
```

#### step文

```go
// step(10) { func1();, func2();,, end_step; del_me; }
// カンマは待機命令に変換
[]OpCode{
    {Cmd: OpSetStep, Args: []any{10}},
    {Cmd: OpCall, Args: []any{"func1"}},
    {Cmd: OpWait, Args: []any{1}},  // カンマ1つ
    {Cmd: OpCall, Args: []any{"func2"}},
    {Cmd: OpWait, Args: []any{2}},  // カンマ2つ
    {Cmd: OpCall, Args: []any{"del_me"}},
}
```

---

## Part 4: 統合API

### コンパイラAPI

```go
package compiler

// Compile はソースコードをOpCodeにコンパイルする
func Compile(source string) ([]OpCode, []error) {
    l := lexer.New(source)
    p := parser.New(l)
    program, errs := p.ParseProgram()
    if len(errs) > 0 {
        return nil, errs
    }
    
    c := compiler.New()
    return c.Compile(program)
}

// CompileFile はファイルからソースコードを読み込んでコンパイルする
func CompileFile(path string) ([]OpCode, []error) {
    loader := script.NewLoader(filepath.Dir(path))
    scripts, err := loader.LoadAllScripts()
    if err != nil {
        return nil, []error{err}
    }
    
    // mainを含むスクリプトを探す
    // ...
}

// CompileWithOptions はオプション付きでコンパイルする
type CompileOptions struct {
    Debug bool // デバッグ情報を含める
}

func CompileWithOptions(source string, opts CompileOptions) ([]OpCode, []error)
```

---

## Part 5: エラーハンドリング

### エラー形式

```go
type CompileError struct {
    Phase   string // "lexer", "parser", "compiler"
    Message string
    Line    int
    Column  int
    Context string // エラー周辺のソースコード
}

func (e *CompileError) Error() string {
    return fmt.Sprintf("%s error at line %d, column %d: %s\n%s",
        e.Phase, e.Line, e.Column, e.Message, e.Context)
}
```

### エラーコンテキスト

```go
// エラー行の前後2行を含むコンテキストを生成
func generateErrorContext(source string, line, column int) string {
    lines := strings.Split(source, "\n")
    start := max(0, line-3)
    end := min(len(lines), line+2)
    
    var buf strings.Builder
    for i := start; i < end; i++ {
        if i == line-1 {
            buf.WriteString(fmt.Sprintf("> %4d | %s\n", i+1, lines[i]))
            buf.WriteString(fmt.Sprintf("       | %s^\n", strings.Repeat(" ", column-1)))
        } else {
            buf.WriteString(fmt.Sprintf("  %4d | %s\n", i+1, lines[i]))
        }
    }
    return buf.String()
}
```

---

## 正確性プロパティ

### P1: Lexerの完全性

すべての有効なFILLYトークンは、Lexerによって正しいTokenTypeで識別される。

### P2: キーワードの大文字小文字非依存

キーワードは大文字小文字に関係なく同じTokenTypeとして識別される。
例: `MES`, `mes`, `Mes` → すべて `TOKEN_MES`

### P3: Parserの構文木整合性

有効なFILLYソースコードに対して、Parserは構文的に正しいASTを生成する。

### P4: OpCodeの完全性

ASTのすべてのノードは、対応するOpCodeに変換される。

### P5: step文のカンマ変換

step文内の連続するカンマは、その数に等しいWaitカウントを持つOpWait命令に変換される。

### P6: エラー位置の正確性

すべてのコンパイルエラーは、正確な行番号と列番号を含む。

---

## Part 6: アプリケーション統合

### app.goでのコンパイラ呼び出し

```go
// pkg/app/app.go

func (app *Application) Run(args []string) error {
    // ... 既存の処理 ...

    // 4. スクリプトファイルの読み込み
    scripts, err := app.loadScripts(selectedTitle.Path)
    if err != nil {
        return fmt.Errorf("failed to load scripts: %w", err)
    }

    // 5. スクリプトのコンパイル（新規追加）
    opcodes, err := app.compileScripts(scripts)
    if err != nil {
        return fmt.Errorf("failed to compile scripts: %w", err)
    }

    app.log.Info("Scripts compiled", "opcode_count", len(opcodes))

    // 6. 仮想デスクトップの実行
    // ...
}

// compileScripts スクリプトをコンパイルしてOpCodeを生成
func (app *Application) compileScripts(scripts []script.Script) ([]compiler.OpCode, error) {
    // mainエントリーポイントを探す
    mainScript, err := findMainScript(scripts)
    if err != nil {
        return nil, err
    }

    // コンパイル実行
    opcodes, errs := compiler.Compile(mainScript.Content)
    if len(errs) > 0 {
        for _, e := range errs {
            app.log.Error("Compile error", "error", e)
        }
        return nil, fmt.Errorf("compilation failed with %d errors", len(errs))
    }

    return opcodes, nil
}
```

### mainエントリーポイントの検出

```go
// pkg/script/script.go に追加

// FindMainScript main関数を含むスクリプトを探す
func FindMainScript(scripts []Script) (*Script, error) {
    var mainScripts []*Script
    
    for i := range scripts {
        if containsMainFunction(scripts[i].Content) {
            mainScripts = append(mainScripts, &scripts[i])
        }
    }
    
    if len(mainScripts) == 0 {
        return nil, fmt.Errorf("no main function found in any script file")
    }
    
    if len(mainScripts) > 1 {
        names := make([]string, len(mainScripts))
        for i, s := range mainScripts {
            names[i] = s.FileName
        }
        return nil, fmt.Errorf("multiple main functions found in: %v", names)
    }
    
    return mainScripts[0], nil
}

// containsMainFunction main関数が含まれているかチェック
// 簡易的な正規表現マッチングで判定
func containsMainFunction(content string) bool {
    // main() または main(params) の形式を検出
    // 大文字小文字を区別しない
    pattern := regexp.MustCompile(`(?i)\bmain\s*\([^)]*\)\s*\{`)
    return pattern.MatchString(content)
}
```

### 依存関係の解決（#include対応）

```go
// pkg/compiler/preprocessor.go

type Preprocessor struct {
    basePath string
    loaded   map[string]bool // 循環参照検出用
}

func NewPreprocessor(basePath string) *Preprocessor {
    return &Preprocessor{
        basePath: basePath,
        loaded:   make(map[string]bool),
    }
}

// Process #includeディレクティブを処理してソースを結合
func (p *Preprocessor) Process(source string, filename string) (string, error) {
    if p.loaded[filename] {
        return "", fmt.Errorf("circular include detected: %s", filename)
    }
    p.loaded[filename] = true
    
    // #include "filename" を検出して置換
    pattern := regexp.MustCompile(`#include\s+"([^"]+)"`)
    
    result := pattern.ReplaceAllStringFunc(source, func(match string) string {
        // インクルードファイルを読み込んで再帰処理
        // ...
    })
    
    return result, nil
}
```

---

## 実装上の注意

### 1. 大文字小文字の扱い

FILLYはキーワードと識別子の両方で大文字小文字を区別しません。Lexerでは識別子を元の形式で保持しつつ、キーワード判定時にのみ小文字に変換します。

### 2. step文の特殊構文

step文内では、カンマが待機命令として機能します。これは通常の文法とは異なるため、Parser内で特別な処理が必要です。

### 3. 関数定義の判別

`name(params){body}`形式の関数定義と、`name(args)`形式の関数呼び出しを区別する必要があります。パラメータリスト内に型指定やデフォルト値がある場合、または`)`の後に`{`が続く場合は関数定義です。

### 4. mes文のイベントタイプ

mes文のイベントタイプ（TIME, MIDI_TIME等）は、キーワードとしてではなく識別子として解析し、有効なイベントタイプかどうかを検証します。
