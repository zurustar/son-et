# FILLY Script Language Specification

## 概要

このドキュメントは、FILLY スクリプト言語の完全なリファレンスです。FILLY は Windows 3.1 時代のマルチメディアアプリケーション用スクリプト言語で、son-et はそのモダンな実装です。

**参考資料**: 秀丸エディタ用TOFFYライターマクロ「らくらくTOFFYライター for 秀丸エディタ Ver.1.30」

**文字エンコーディング**: FILLYスクリプト（.TFYファイル）は通常Shift-JISエンコーディングで保存されています。son-etは読み込み時に自動的にShift-JISからUTF-8に変換します。これにより、日本語を含むスクリプトが正しく処理されます。

---

## 構文概要

### 基本構文
- **スタイル**: C言語スタイルの手続き型言語
- **関数宣言**: `function`キーワードなし（例: `main() { ... }`, `MyFunc(int x, str s) { ... }`）
- **変数宣言**: 型キーワード付き（例: `int x, y, z;`, `str name;`, `int arr[];`）
- **エントリーポイント**: `main() { ... }` 関数が必須
- **コメント**: `//` で単一行コメント、`/* ... */` で複数行コメント
- **文の終端**: `;` (セミコロン)
- **大文字小文字**: 
  - **ファイル名**: 大文字小文字を区別しない (Windows 3.1 互換性)
  - **識別子**: 大文字小文字を区別しない (変数名、関数名)

### プリプロセッサディレクティブ

**#info - メタデータ定義**:
作品情報（タイトル、作者、著作権など）を定義します。

```filly
#info INAM "作品タイトル"
#info IART "作者名"
#info ICOP "著作権情報"
#info GENR "ジャンル"
#info WRIT "脚本担当,メールアドレス"
#info GRPC "グラフィック担当,メールアドレス"
#info COMP "作曲担当,メールアドレス"
#info PROD "制作者,メールアドレス"
#info CONT "連絡先,メールアドレス"
#info MDFY "改変可否 (YES/NO/Ask me)"
#info TRNS "転載可否 (YES/NO/Ask me)"
#info JINT "二次創作可否 (YES/NO/Ask me)"
#info VIDO "画面解像度と色数 (例: 640x480, 256)"
#info INST "MIDI音源 (例: GM, GS, SC-88)"
#info ISBJ "サブタイトル"
#info ICMT "コメント"
```

**注意**: `#info`ディレクティブは実行時には無視されますが、作品情報として保持されます。

**#include - ファイルインクルード**:
他のTFYファイルを現在のファイルに取り込みます。

```filly
#include "SCENE1.TFY"
#include "SCENE2.TFY"
#include "COMMON.TFY"
```

**動作**:
- インクルードされたファイルの内容が、`#include`の位置に展開されます
- 相対パスで指定し、プロジェクトディレクトリからの相対パスとして解決されます
- ファイル名は大文字小文字を区別しません（Windows 3.1互換性）
- 循環インクルードは検出されエラーとなります
- インクルードは再帰的に処理されます（インクルードされたファイルが別のファイルをインクルード可能）

### ウェイト構文
FILLY の最も特徴的な機能は、ステップベースの実行モデルです。

**MIDI同期モード (`mes(MIDI_TIME)`)**:
- `step(n)` で32分音符の倍数を指定
- 例: `step(8)` = 4分音符

**タイムモード (`mes(TIME)`)**:
- `step(n)` でウェイト単位をミリ秒で指定
- 計算式: `1 step = n × 50ms`
- 例: `step(20)` = 1秒 (1000ms)

---

## データ型

### 変数宣言
FILLYはC言語スタイルの型付き変数宣言を使用します。

**基本型**:
- `int`: 整数値、ピクチャーハンドル、ウィンドウハンドル
- `str`: 文字列値
- `int[]`: 整数配列（動的サイズ）

**変数宣言の構文**:
```filly
// 単一変数の宣言
int x;
str name;

// 複数変数の同時宣言（カンマ区切り）
int x, y, z;
str firstName, lastName;

// 配列の宣言
int arr[];           // 空配列
int scores[10];      // サイズ指定（実際は動的に拡張）

// グローバル変数（関数外で宣言）
int globalCounter;
str globalMessage;
```

**変数の初期化**:
```filly
// 宣言と同時に初期化はできない（C言語と異なる）
int x;
x = 10;  // 別の文で代入

// 配列への代入
int arr[];
arr[0] = 100;
arr[1] = 200;
```

**スコープ**:
- グローバル変数: 関数外で宣言、すべての関数とmes()ブロックからアクセス可能
- ローカル変数: 関数内で宣言、その関数とネストされたmes()ブロックからアクセス可能
- 変数名は大文字小文字を区別しない

### 配列

**配列の特徴**:
- 動的サイズ: 配列は自動的に拡張されます
- 0ベースインデックス: 最初の要素は `arr[0]`
- 整数のみ: 配列要素は整数値のみ
- 初期値: 未初期化要素は0

**配列の使用例**:
```filly
// 配列への代入
scores[0] = 85
scores[1] = 92
scores[2] = 78

// 配列からの読み取り
total = scores[0] + scores[1] + scores[2]

// 配列サイズの取得
size = ArraySize(scores)  // 3

// 配列の操作
InsArrayAt(scores, 1, 90)  // インデックス1に90を挿入
DelArrayAt(scores, 2)      // インデックス2の要素を削除
DelArrayAll(scores)        // 全要素を削除
```

**配列の制限**:
- 多次元配列は非サポート
- 文字列配列は非サポート（整数のみ）
- 配列のコピーは要素ごとに行う必要があります

### 関数定義
FILLYはC言語スタイルの関数定義を使用します。`function`キーワードは不要です。

**基本構文**:
```filly
// パラメータなしの関数
main() {
    // 関数本体
}

// パラメータ付きの関数（型指定あり）
MyFunction(int x, str name) {
    // 関数本体
}

// 配列パラメータ（型指定あり）
ProcessArray(int arr[], int size) {
    // arr は整数配列
}

// 配列パラメータ（型指定なし）
UpdateSprites(p[], c[]) {
    // p と c は配列（型は実行時に決定）
}

// 混合パラメータ（通常 + 配列 + デフォルト値）
DrawSprite(int cast_id, int positions[], int x, int y, int color=0xFFFFFF) {
    // cast_id: 通常の整数パラメータ
    // positions: 配列パラメータ
    // x, y: 通常の整数パラメータ
    // color: デフォルト値付きパラメータ
}

// デフォルトパラメータ値
DrawText(int x, int y, str text, int color=0xFFFFFF) {
    // color が指定されない場合は白色（0xFFFFFF）
}
```

**関数の特徴**:
- 関数名は大文字小文字を区別しない
- パラメータには型を指定する（`int`, `str`）
- 配列パラメータは `[]` を付けて宣言（`int arr[]`, `p[]`）
- 型なし配列パラメータもサポート（`p[]`, `c[]`）
- デフォルトパラメータ値をサポート（`=`で指定）
- 戻り値の型指定は不要（暗黙的）
- 再帰呼び出しをサポート

**配列パラメータの使用例**:
```filly
// グローバル配列
int sprites[];
int colors[];

// メイン関数
main() {
    sprites[0] = 1;
    sprites[1] = 2;
    colors[0] = 0xFF0000;
    colors[1] = 0x00FF00;
    
    // 配列を関数に渡す
    InitializeSprites(sprites, colors);
}

// 型付き配列パラメータ
InitializeSprites(int sprite_ids[], int color_values[]) {
    // 配列要素にアクセス
    MoveCast(sprite_ids[0], 100, 100);
    SetPaintColor(color_values[0]);
}

// 型なし配列パラメータ（レガシー構文）
Scene1ON(p[], c[]) {
    // p と c は配列として扱われる
    p[0] = LoadPic("image1.bmp");
    c[0] = PutCast(p[0], 0, 0);
}
```

**実際のサンプルからの例**:
```filly
// robot サンプルより
Chap1ON(int p[], int c[]) {
    // 型付き配列パラメータ
    p[0] = LoadPic("ROBOT002.bmp");
    c[1] = PutCast(p[2], p[0], 0, 200);
}

// robot サンプルより（デフォルト値付き）
OP_walk(c, p[], x, y, w, h, l=10) {
    // c: 通常パラメータ
    // p[]: 配列パラメータ
    // l: デフォルト値10
    if (l != 0) {
        MoveCast(c, p[1], x, y);
    }
}
```

### 暗黙的定数
- `TIME` / `time`: タイムモード定数 (値: 0)
- `USER` / `user`: ユーザーイベント定数 (値: 1)
- `MIDI_END` / `midi_end`: MIDI終了定数 (値: 0)
- `MIDI_TIME` / `MidiTime`: MIDI同期モード定数 (値: 1)

### 暗黙的グローバル変数
- `MidiTime`: 現在のMIDI時刻
- `MesP1`, `MesP2`, `MesP3`, `MesP4`: メッセージパラメータ

---

## ウィンドウ関連関数

### OpenWin
仮想デスクトップ上に仮想ウインドウを開く

**シンプル版**:
```filly
win_id = OpenWin(pic)
```

**詳細版**:
```filly
win_id = OpenWin(pic, x, y, width, height, pic_x, pic_y, color)
```

**引数**:
- `pic`: ピクチャー番号
- `x`: ウィンドウの位置 X座標
- `y`: ウィンドウの位置 Y座標
- `width`: ウィンドウの幅
- `height`: ウィンドウの高さ
- `pic_x`: ピクチャーの左上 X座標
- `pic_y`: ピクチャーの左上 Y座標
- `color`: ピクチャーが無い位置の色(16進数)

**戻り値**: ウィンドウID（0から始まる連番）

### MoveWin
仮想ウインドウの設定を変更

```filly
MoveWin(win, pic, x, y, width, height, pic_x, pic_y)
MoveWin(win, pic) // 短縮形（ピクチャー変更のみ）
```

**引数**:
- `win`: ウィンドウ番号
- `pic`: ピクチャー番号
- `x`: ウィンドウの位置 X座標
- `y`: ウィンドウの位置 Y座標
- `width`: ウィンドウの幅
- `height`: ウィンドウの高さ
- `pic_x`: ピクチャーの左上 X座標
- `pic_y`: ピクチャーの左上 Y座標

### CloseWin
仮想ウインドウを閉じる

```filly
CloseWin(win_no)
```

### CloseWinAll
すべての仮想ウインドウを閉じる

```filly
CloseWinAll()
```

### CapTitle
ウィンドウのキャプションの文字を指定

```filly
CapTitle(win_no, title)
```

### GetPicNo
ウィンドウに関連付けされたピクチャー番号を得る

```filly
pic_no = GetPicNo(win_no)
```

---

## ピクチャー関連関数

### LoadPic
画像ファイルの読み込み

```filly
pic_id = LoadPic(filename)
```

**引数**:
- `filename`: ファイル名(文字列)

**戻り値**: ピクチャーID（0から始まる連番）

**使用例**:
```filly
pic0 = LoadPic("image1.bmp");  // ID=0 が返される
pic1 = LoadPic("image2.bmp");  // ID=1 が返される
pic2 = LoadPic("image3.bmp");  // ID=2 が返される
```

### MovePic
画像データの転送

```filly
MovePic(src_pic, src_x, src_y, width, height, dst_pic, dst_x, dst_y, mode)
```

**引数**:
- `src_pic`: 移動元ピクチャー番号
- `src_x`: 移動元の始点 X座標
- `src_y`: 移動元の始点 Y座標
- `width`: サイズ 幅
- `height`: サイズ 高さ
- `dst_pic`: 移動先ピクチャー番号
- `dst_x`: 移動先の始点 X座標
- `dst_y`: 移動先の始点 Y座標
- `mode`: 転送モード
  - `0`: 通常
  - `1`: 透明色転送モード
  - `2`: シーンチェンジモード

### MoveSPic
画像データを拡大縮小して転送

```filly
MoveSPic(src_pic, src_x, src_y, src_w, src_h, dst_pic, dst_x, dst_y, dst_w, dst_h)
MoveSPic(src_pic, src_x, src_y, src_w, src_h, dst_pic, dst_x, dst_y, dst_w, dst_h, 0, trans_color)
```

**引数**:
- `src_pic`: 移動元ピクチャー番号
- `src_x`: 移動元の始点 X座標
- `src_y`: 移動元の始点 Y座標
- `src_w`: 移動元のサイズ 幅
- `src_h`: 移動元のサイズ 高さ
- `dst_pic`: 移動先ピクチャー番号
- `dst_x`: 移動先の始点 X座標
- `dst_y`: 移動先の始点 Y座標
- `dst_w`: 移動先のサイズ 幅
- `dst_h`: 移動先のサイズ 高さ
- `trans_color`: 透明色(16進数) - 透明色転送をする場合

### DelPic
画像データの破棄

```filly
DelPic(pic_no)
```

### CreatePic
ピクチャーの生成

```filly
pic_id = CreatePic(pic_no, width, height)
pic_id = CreatePic(pic_no, width, height, 0)  // デスクトップの取得
```

**引数**:
- `pic_no`: 基準ピクチャー番号
- `width`: サイズ 幅
- `height`: サイズ 高さ
- 第4引数に `0` を指定するとデスクトップの取得

### PicWidth
ピクチャーの幅の取得

```filly
width = PicWidth(pic_no)
```

### PicHeight
ピクチャーの高さの取得

```filly
height = PicHeight(pic_no)
```

### ReversePic
左右反転イメージの転写

```filly
ReversePic(src_pic, src_x, src_y, width, height, dst_pic, dst_x, dst_y)
```

**引数**:
- `src_pic`: 移動元ピクチャー番号
- `src_x`: 移動元の始点 X座標
- `src_y`: 移動元の始点 Y座標
- `width`: サイズ 幅
- `height`: サイズ 高さ
- `dst_pic`: 移動先ピクチャー番号
- `dst_x`: 移動先の始点 X座標
- `dst_y`: 移動先の始点 Y座標

---

## キャスト（スプライト）関連関数

### PutCast
キャストの配置

```filly
cast_id = PutCast(src_pic_no, dst_pic_no, x, y, trans_color, src_x, src_y, width, height)
```

**引数**:
- `src_pic_no`: ソースピクチャー番号（画像の取得元）
- `dst_pic_no`: 配置先ピクチャー番号（キャストを配置する先）
- `x`: 配置先ピクチャー内での配置位置 X座標
- `y`: 配置先ピクチャー内での配置位置 Y座標
- `trans_color`: 透明色（16進数、省略時は黒 0x000000）
- `src_x`: ソースピクチャーの切り出し始点 X座標
- `src_y`: ソースピクチャーの切り出し始点 Y座標
- `width`: 切り出しサイズ 幅
- `height`: 切り出しサイズ 高さ

**戻り値**: キャストID（0から始まる連番）

**使用例**:
```filly
// y_saruサンプルより
base_pic = CreatePic(25);
OpenWin(base_pic, 0, 0, winW, winH, winX, winY, 0xffffff);
PutCast(25, base_pic, 0, 0);  // ピクチャー25からbase_picに背景を配置

// キャストを配置（ソース=17、配置先=base_pic）
cast = PutCast(17, base_pic, 113, 0, 0xffffff, 0, 0, 89, 77);
```

**注意**: 
- 第1引数はソースピクチャー番号（画像の取得元）
- 第2引数は配置先ピクチャー番号（キャストを配置する先）
- ウインドウIDではなくピクチャーIDを指定する

### MoveCast
キャストの移動

```filly
MoveCast(cast_no, x, y)
MoveCast(cast_no, src_pic_no, x, y, src_x, src_y, width, height)
```

**引数**:
- `cast_no`: キャスト番号
- `src_pic_no`: ソースピクチャー番号（画像の取得元、省略時は元のピクチャーを使用）
- `x`: 配置位置 X座標
- `y`: 配置位置 Y座標
- `src_x`: ソースピクチャーの切り出し始点 X座標
- `src_y`: ソースピクチャーの切り出し始点 Y座標
- `width`: 切り出しサイズ 幅
- `height`: 切り出しサイズ 高さ

**使用例**:
```filly
// y_saruサンプルより
cast = PutCast(17, base_pic, 113, 0, 0xffffff, 0, 0, 89, 77);
MoveCast(cast, 17, 109, 0, 0, 89, 77, 0, 0);  // ソース=17、位置=(109,0)
MoveCast(cast, 17, 105, 10, 0, 89, 77, 89, 0);  // ソース=17、位置=(105,10)
```

### DelCast
キャストの削除

```filly
DelCast(cast_no)
```

---

## 文字表示関連関数

### SetFont
フォントの設定

```filly
SetFont(font_name, size, charset, weight, italic, underline, strikeout)
```

**引数**:
- `font_name`: フォント名
- `size`: フォントサイズ
- `charset`: 文字セット (128=日本語)
- `weight`: 太さ (400=標準, 700=太字)
- `italic`: イタリック (0=なし, 1=あり)
- `underline`: 下線 (0=なし, 1=あり)
- `strikeout`: 取り消し線 (0=なし, 1=あり)

### TextWrite
文字列の描画

```filly
TextWrite(pic_no, x, y, text)
```

**引数**:
- `pic_no`: ピクチャー番号
- `x`: 描画位置 X座標
- `y`: 描画位置 Y座標
- `text`: 描画する文字列

### TextColor
文字色の設定

```filly
TextColor(color)
```

**引数**:
- `color`: 色(16進数)

### BgColor
背景色の設定

```filly
BgColor(color)
```

**引数**:
- `color`: 色(16進数)

### BackMode
背景モードの設定

```filly
BackMode(mode)
```

**引数**:
- `mode`: 背景モード (0=透明, 1=不透明)

---

## 描画関連関数

### DrawLine
直線の描画

```filly
DrawLine(pic_no, x1, y1, x2, y2)
```

### DrawCircle
円の描画

```filly
DrawCircle(pic_no, x, y, radius, fill_mode)
```

**引数**:
- `fill_mode`: 塗りつぶしモード (0=なし, 1=ハッチ, 2=ソリッド)

### DrawRect
矩形の描画

```filly
DrawRect(pic_no, x1, y1, x2, y2, fill_mode)
```

### SetLineSize
線の太さの設定

```filly
SetLineSize(size)
```

### SetPaintColor
描画色の設定

```filly
SetPaintColor(color)
```

### GetColor
ピクセルの色の取得

```filly
color = GetColor(pic_no, x, y)
```

### SetROP
ラスタオペレーションの設定

```filly
SetROP(rop_mode)
```

**ROP モード**:
- `COPYPEN`: 通常コピー
- `XORPEN`: XOR
- `MERGEPEN`: OR
- その他のラスタオペレーション

---

## 文字列関連関数

### StrLen
文字列の長さを取得

```filly
length = StrLen(str)
```

### SubStr
部分文字列の取得

```filly
substr = SubStr(str, start, length)
```

### StrFind
文字列の検索

```filly
pos = StrFind(str, search_str)
```

**戻り値**: 見つかった位置（0から始まる）、見つからない場合は-1

### StrPrint
書式付き文字列の生成

```filly
result = StrPrint(format, arg1, arg2, ...)
```

**使用例**:
```filly
msg = StrPrint("Score: %d", score);
```

### StrInput
文字列の入力

```filly
input = StrInput(prompt, default_value)
```

### CharCode
文字コードの取得

```filly
code = CharCode(str, index)
```

### StrCode
文字コードから文字列を生成

```filly
str = StrCode(code)
```

### StrUp
大文字に変換

```filly
upper = StrUp(str)
```

### StrLow
小文字に変換

```filly
lower = StrLow(str)
```

---

## ファイル操作関連関数

### INIファイル操作

#### WriteIniInt
INIファイルに整数を書き込み

```filly
WriteIniInt(filename, section, key, value)
```

#### GetIniInt
INIファイルから整数を読み込み

```filly
value = GetIniInt(filename, section, key, default_value)
```

#### WriteIniStr
INIファイルに文字列を書き込み

```filly
WriteIniStr(filename, section, key, value)
```

#### GetIniStr
INIファイルから文字列を読み込み

```filly
value = GetIniStr(filename, section, key, default_value)
```

### バイナリファイルI/O

#### OpenF
ファイルを開く

```filly
handle = OpenF(filename, mode)
```

**モード**:
- `"r"`: 読み込み
- `"w"`: 書き込み
- `"a"`: 追記

#### CloseF
ファイルを閉じる

```filly
CloseF(handle)
```

#### ReadF
ファイルから読み込み

```filly
value = ReadF(handle, size)
```

#### WriteF
ファイルに書き込み

```filly
WriteF(handle, value)
```

#### SeekF
ファイルポインタの移動

```filly
SeekF(handle, offset, origin)
```

#### StrReadF
ファイルから文字列を読み込み

```filly
str = StrReadF(handle)
```

#### StrWriteF
ファイルに文字列を書き込み

```filly
StrWriteF(handle, str)
```

### ファイル管理

#### CopyFile
ファイルのコピー

```filly
CopyFile(src_file, dst_file)
```

#### DelFile
ファイルの削除

```filly
DelFile(filename)
```

#### IsExist
ファイルの存在確認

```filly
exists = IsExist(filename)
```

#### MkDir
ディレクトリの作成

```filly
MkDir(dirname)
```

#### RmDir
ディレクトリの削除

```filly
RmDir(dirname)
```

#### ChDir
カレントディレクトリの変更

```filly
ChDir(dirname)
```

#### GetCwd
カレントディレクトリの取得

```filly
dir = GetCwd()
```

---

## 配列操作関連関数

### ArraySize
配列のサイズを取得

```filly
size = ArraySize(array)
```

### DelArrayAll
配列の全要素を削除

```filly
DelArrayAll(array)
```

### DelArrayAt
配列の指定位置の要素を削除

```filly
DelArrayAt(array, index)
```

### InsArrayAt
配列の指定位置に要素を挿入

```filly
InsArrayAt(array, index, value)
```

---

## 整数関連関数

### Random
乱数の生成

```filly
value = Random(max)
```

**戻り値**: 0 から max-1 までの乱数

### MakeLong
2つの16ビット値を32ビット値に結合

```filly
long_value = MakeLong(low_word, high_word)
```

### GetHiWord
32ビット値の上位16ビットを取得

```filly
high_word = GetHiWord(long_value)
```

### GetLowWord
32ビット値の下位16ビットを取得

```filly
low_word = GetLowWord(long_value)
```

---

## オーディオ関連関数

### PlayMIDI
MIDIファイルの再生

```filly
PlayMIDI(filename)
```

**引数**:
- `filename`: MIDIファイル名

**注意**: 
- 再生は非同期（バックグラウンド）で行われる
- `mes(MIDI_TIME)` ブロックと組み合わせて使用

### PlayWAVE
WAVファイルの再生

```filly
PlayWAVE(filename)
```

**引数**:
- `filename`: WAVファイル名

### リソース管理

#### LoadRsc
リソースの読み込み

```filly
LoadRsc(id, filename)
```

#### PlayRsc
リソースの再生

```filly
PlayRsc(id)
```

#### DelRsc
リソースの削除

```filly
DelRsc(id)
```

---

## メッセージ関連関数

### GetMesNo
現在のメッセージ番号を取得

```filly
mes_no = GetMesNo()
```

### DelMes
指定したメッセージブロックを削除

```filly
DelMes(mes_no)
```

### FreezeMes
メッセージブロックを一時停止

```filly
FreezeMes(mes_no)
```

### ActivateMes
メッセージブロックを再開

```filly
ActivateMes(mes_no)
```

### PostMes
カスタムメッセージの送信

```filly
PostMes(mes_type, p1, p2, p3, p4)
```

**引数**:
- `mes_type`: メッセージタイプ
- `p1`, `p2`, `p3`, `p4`: メッセージパラメータ

---

## システム関連関数

### WinInfo
ウィンドウ情報の取得

```filly
width = WinInfo(0)   // デスクトップ幅 (1280)
height = WinInfo(1)  // デスクトップ高さ (720)
```

### GetSysTime
システム時刻の取得

```filly
time = GetSysTime()
```

### WhatDay
日付の取得

```filly
day = WhatDay()
```

### WhatTime
時刻の取得

```filly
time = WhatTime()
```

### GetCmdLine
コマンドライン引数の取得

```filly
cmdline = GetCmdLine()
```

### Shell
外部プログラムの実行

```filly
Shell(command, working_dir)
```

---

## 制御構文

### mes ブロック
イベント駆動の実行ブロック

```filly
mes(TIME) {
    // タイムモードのコード
}

mes(MIDI_TIME) {
    // MIDI同期モードのコード
}

mes(MIDI_END) {
    // MIDI終了時のコード
}

mes(KEY) {
    // キーボード入力時のコード
    // 任意のキーが押されたときに実行
}

mes(CLICK) {
    // マウスクリック時のコード
    // マウスがクリックされたときに実行
}

mes(RBDOWN) {
    // 右ボタンダウン時のコード
    // 右マウスボタンが押されたときに実行
}

mes(RBDBLCLK) {
    // 右ボタンダブルクリック時のコード
    // 右マウスボタンがダブルクリックされたときに実行
}

mes(USER) {
    // ユーザーイベントのコード
    // PostMes()で送信されたカスタムメッセージを受信
}
```

**イベントタイプ**:
- `TIME`: 60 FPSのフレーム更新で駆動（ブロッキング）
- `MIDI_TIME`: MIDI再生のティックで駆動（ノンブロッキング）
- `MIDI_END`: MIDI再生終了時に1回実行
- `KEY`: キーボード入力時に実行
- `CLICK`: マウスクリック時に実行
- `RBDOWN`: 右マウスボタンダウン時に実行
- `RBDBLCLK`: 右マウスボタンダブルクリック時に実行
- `USER`: カスタムメッセージ受信時に実行

### step ブロック
ステップ単位の実行

**シンプル形式**:
```filly
step(n);  // n ステップ待機
```

**ブロック形式**:
```filly
step(n) {
    command1;,      // command1を実行し、1ステップ待機
    command2;,,     // command2を実行し、2ステップ待機
    command3;,      // command3を実行し、1ステップ待機
    end_step;       // ブロックを終了（オプション）
}
```

**カンマのセマンティクス**:
- `;,` (セミコロン + 1カンマ) → コマンドを実行し、Wait(1)
- `;,,` (セミコロン + 2カンマ) → コマンドを実行し、Wait(2)
- `;,,,` (セミコロン + 3カンマ) → コマンドを実行し、Wait(3)
- 待機時間 = カンマの数 × ステップ期間（step(n)のn）

**ステップ期間の計算**:
- `TIME`モード: 1ステップ = n × 50ms（n × 3ティック @ 60 FPS）
- `MIDI_TIME`モード: 1ステップ = n × 32分音符

**使用例**:
```filly
// TIMEモード: step(65) = 65 × 50ms = 3.25秒/ステップ
mes(TIME) {
    step(65) {
        PlayWAVE("sound.wav");,,  // 再生し、2ステップ待機（6.5秒）
        MoveWin(0, 1);,,          // 移動し、2ステップ待機
        MoveWin(0, 2);,           // 移動し、1ステップ待機（3.25秒）
    }
}

// MIDI_TIMEモード: step(8) = 8 × 32分音符 = 4分音符/ステップ
mes(MIDI_TIME) {
    step(8) {
        MoveCast(0, x, y);,       // 移動し、1ステップ待機
        x = x + 10;,              // 更新し、1ステップ待機
    }
}
```

**注意**: 
- カンマがない場合、コマンドは待機なしで即座に実行される
- `end_step` でブロックを明示的に終了（オプション）
- この構文はKUMA2などのレガシーFILLYスクリプトで広く使用されている

### if-else
条件分岐

```filly
if (condition) {
    // 真の場合
} else {
    // 偽の場合
}
```

### for
繰り返し

```filly
for (init; condition; increment) {
    // ループ本体
}
```

### while
条件付き繰り返し

```filly
while (condition) {
    // ループ本体
}
```

### do-while
後判定繰り返し

```filly
do {
    // ループ本体
} while (condition);
```

### switch-case
多分岐

```filly
switch (value) {
    case 1:
        // value が 1 の場合
        break;
    case 2:
        // value が 2 の場合
        break;
    default:
        // その他の場合
        break;
}
```

### break
ループの中断

```filly
break;
```

### continue
ループの次の反復へ

```filly
continue;
```

---

## 特殊キーワード

### ESCキー（特別なハンドリング）

**ESCキー**は特別な扱いを受けます：

```filly
// ESCキーが押されると、プログラムは即座に終了します
// mes()ブロックや明示的なハンドラは不要
```

**動作**:
1. ESCキーが押されると、システムは終了フラグを設定
2. 現在実行中のOpCodeが完了した後、VM実行を停止
3. すべてのシーケンスを停止
4. プログラムを終了

**注意**: 
- ESCキーは`mes(KEY)`では捕捉できません
- ESCキーは常にプログラム終了として扱われます
- スクリプトからESCキーの動作を変更することはできません

### del_me
現在のシーケンスを終了

```filly
del_me;
```

**注意**: 
- 括弧なしで呼び出し可能
- 現在のシーケンスのみを終了（他のシーケンスは継続）

### del_us
同じグループのシーケンスを終了

```filly
del_us;
```

### del_all
すべてのシーケンスを終了し、リソースをクリーンアップ

```filly
del_all;
```

**注意**: 
- すべてのウィンドウを閉じる
- グラフィックスリソースをクリーンアップ
- MIDI再生は継続

### end_step
step ブロックの終了

```filly
end_step;
```

---

## サポート範囲

### son-et で実装される機能

**コア機能**:
- すべてのウィンドウ、ピクチャー、キャスト操作
- テキストレンダリング
- 基本的な描画機能
- 文字列操作
- ファイルI/O（INI、バイナリ）
- 配列操作
- MIDI/WAV再生
- メッセージシステム
- すべての制御構文

### son-et で実装されない機能

**Windows固有のAPI**:
- `PlayCD` - CD-ROMオーディオ（廃止されたハードウェア）
- `MCI`, `StrMCI` - Windows MCIコマンド
- `SetRegStr`, `GetRegStr` - Windowsレジストリアクセス
- `PlayAVI` - AVIビデオ再生（複雑なコーデックサポート）

**代替手段**:
- CD audio → `PlayMIDI` または `PlayWAVE` でデジタルオーディオファイルを使用
- MCI commands → `PlayMIDI`, `PlayWAVE`, またはプラットフォーム固有の代替手段
- Registry access → INIファイル (`WriteIniInt`, `GetIniInt`, `WriteIniStr`, `GetIniStr`)
- AVI playback → 外部プレーヤーでモダンなビデオフォーマットを使用

---

## 使用例

### 基本的なウィンドウ表示

```filly
main() {
    pic = LoadPic("image.bmp");
    OpenWin(pic);
}
```

### MIDI同期アニメーション

```filly
main() {
    pic = LoadPic("sprite.bmp");
    
    mes(MIDI_TIME) {
        step(8) {
            MoveCast(0, x, y);,
            x = x + 10;,
            end_step;
        }
    }
    
    PlayMIDI("music.mid");
}
```

### タイムモードアニメーション

```filly
main() {
    pic = LoadPic("background.bmp");
    OpenWin(pic);
    
    mes(TIME) {
        step(20) {
            // 1秒ごとに実行
            TextWrite(pic, 10, 10, "Hello World");,
            end_step;
        }
    }
}
```

---

## 参考資料

- [requirements.md](../.kiro/specs/requirements.md) - son-et の要求仕様
- [design.md](../.kiro/specs/design.md) - son-et のアーキテクチャ設計
- [tasks.md](../.kiro/specs/tasks.md) - 実装タスクリスト
