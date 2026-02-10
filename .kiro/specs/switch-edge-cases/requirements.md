# 要件定義書

## はじめに

FILLY言語のswitch-case文の実装について、エッジケースの網羅的なテストと必要に応じた修正を行う。現在の実装はパーサー・コンパイラ・VMの各層で基本的なswitch-case/default/breakに対応しているが、以下のエッジケースについてテストカバレッジが不足している可能性がある。

### 現状の実装分析

- **パーサー**: `parseSwitchStatement` / `parseCaseClause` で基本構文をパース
- **コンパイラ**: `compileSwitchStatement` でOpSwitchにコンパイル
- **VM**: `executeSwitch` で最初にマッチしたcaseのみ実行（フォールスルーなし）
- **break**: `breakSignal` を返すが、`executeSwitch` はこれをキャッチせず上位に伝播させる

## 用語集

- **Switch文**: FILLY言語の多分岐制御構文
- **CaseClause**: switch文内の個別の分岐条件と本体
- **Default**: どのcaseにもマッチしない場合に実行されるブロック
- **フォールスルー**: C言語のようにbreakなしで次のcaseに処理が流れる動作（FILLY言語では発生しない）
- **breakSignal**: VM内部でbreak文の実行を通知するシグナル
- **Parser**: FILLY言語のソースコードを抽象構文木（AST）に変換するコンポーネント
- **Compiler**: ASTをOpCodeに変換するコンポーネント
- **VM**: OpCodeを実行する仮想マシン

## 要件

### 要件 1: switch文のbreak動作の正確性

**ユーザーストーリー:** 開発者として、switch文内のbreakがswitch文のみを終了し、外側のループに影響しないことを保証したい。

#### 受け入れ基準

1. WHEN switch文のcase本体にbreak文がある場合、THE VM SHALL そのcase本体の残りの文をスキップしてswitch文を終了する
2. WHEN switch文がループ内にあり、case本体にbreak文がある場合、THE VM SHALL switch文のみを終了し、外側のループの実行を継続する
3. WHEN switch文のcase本体にbreak文がない場合、THE VM SHALL そのcase本体の全文を実行してswitch文を終了する（フォールスルーは発生しない）

### 要件 2: switch文のcase値マッチングの正確性

**ユーザーストーリー:** 開発者として、switch文のcase値が正しく評価・比較されることを保証したい。

#### 受け入れ基準

1. WHEN switch値とcase値が整数で一致する場合、THE VM SHALL そのcaseの本体を実行する
2. WHEN switch値とcase値が文字列で一致する場合、THE VM SHALL そのcaseの本体を実行する
3. WHEN switch値がどのcase値とも一致せずdefaultがある場合、THE VM SHALL defaultブロックを実行する
4. WHEN switch値がどのcase値とも一致せずdefaultがない場合、THE VM SHALL 何も実行せずswitch文を終了する
5. WHEN case値が式（変数参照や演算）の場合、THE VM SHALL 式を評価した結果でマッチングを行う

### 要件 3: switch文のパース正確性

**ユーザーストーリー:** 開発者として、switch文の様々な構文パターンが正しくパースされることを保証したい。

#### 受け入れ基準

1. WHEN switch文に空のcase本体がある場合、THE Parser SHALL エラーなくパースを完了する
2. WHEN switch文がネストされている場合、THE Parser SHALL 内側と外側のswitch文を正しく区別してパースする
3. WHEN switch文のcase本体に複数の文がある場合、THE Parser SHALL 全ての文を正しくパースする
4. WHEN switch文のdefaultの後にcaseがある場合、THE Parser SHALL 正しくパースする（またはエラーを報告する）
