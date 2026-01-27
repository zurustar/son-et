# 要件ドキュメント

## 概要

ROBOTサンプル（samples/robot/ROBOT.TFY）の実行時に発生するエラーを修正するため、不足している組み込み関数を実装する。具体的には以下の3つの機能を追加する：

1. `StrPrint` 関数 - printf形式の文字列フォーマット
2. `CreatePic` の3引数パターン - ソースピクチャーを参照して指定サイズで新規作成
3. `CapTitle` の1引数パターン修正 - 全ての仮想ウィンドウのタイトルを設定

## 用語集

- **StrPrint**: printf形式のフォーマット指定子を使用して文字列を生成する組み込み関数
- **Format_Specifier**: フォーマット文字列内の変換指定子（%ld, %lx, %s, %03d等）
- **CreatePic**: ピクチャーを生成する組み込み関数。複数の呼び出しパターンをサポート
- **CapTitle**: ウィンドウのキャプション（タイトル）を設定する組み込み関数
- **Virtual_Window**: FILLYスクリプトで管理される仮想ウィンドウ
- **Picture_ID**: ピクチャーを識別する整数値

## 要件

### 要件 1: StrPrint関数の実装

**ユーザーストーリー:** 開発者として、printf形式のフォーマット指定子を使用して動的に文字列を生成したい。これにより、ファイル名の連番生成やデバッグメッセージの作成が可能になる。

#### 受け入れ基準

1.1. WHEN StrPrint がフォーマット文字列と引数で呼び出された場合、THE System SHALL フォーマット指定子に従ってフォーマットされた文字列を返す。

1.2. THE System SHALL 10進整数フォーマット用の `%ld` フォーマット指定子をサポートし、Go言語の `%d` に変換する。

1.3. THE System SHALL 16進数フォーマット用の `%lx` フォーマット指定子をサポートし、Go言語の `%x` に変換する。

1.4. THE System SHALL 文字列フォーマット用の `%s` フォーマット指定子をサポートする。

1.5. THE System SHALL ゼロパディング整数用の `%03d` などの幅とパディング指定子をサポートする。

1.6. WHEN StrPrint がエスケープシーケンス（`\n`, `\t`, `\r`）を含むフォーマット文字列で呼び出された場合、THE System SHALL それらを実際の制御文字に変換する。

1.7. WHEN StrPrint がフォーマット指定子より少ない引数で呼び出された場合、THE System SHALL クラッシュせずに適切に処理する（Go言語のfmt.Sprintfの動作に従う）。

1.8. WHEN StrPrint がフォーマット指定子より多い引数で呼び出された場合、THE System SHALL 余分な引数を無視する。

### 要件 2: CreatePic 3引数パターンの実装

**ユーザーストーリー:** 開発者として、既存のピクチャーを参照しながら任意のサイズの新しいピクチャーを作成したい。これにより、スプライトシートから個別のスプライト用バッファを効率的に作成できる。

#### 受け入れ基準

2.1. WHEN CreatePic が3つの引数（srcPicID, width, height）で呼び出された場合、THE System SHALL 指定されたサイズの新しい空のピクチャーを作成する。

2.2. WHEN CreatePic が3引数パターンで呼び出された場合、THE System SHALL ソースピクチャーの内容をコピーせず、空のピクチャーを返す。

2.3. WHEN CreatePic が存在しないソースピクチャーIDで呼び出された場合、THE System SHALL エラーを返す。

2.4. WHEN CreatePic が0以下の幅または高さで呼び出された場合、THE System SHALL エラーを返す。

2.5. THE System SHALL 既存の1引数パターン（CreatePic(srcID)）と2引数パターン（CreatePic(width, height)）との互換性を維持する。

### 要件 3: CapTitle 1引数パターンの修正

**ユーザーストーリー:** 開発者として、1つの引数でCapTitleを呼び出した際に、全ての仮想ウィンドウのタイトルを一括で設定したい。これにより、アプリケーション全体のタイトル管理が簡単になる。

#### 受け入れ基準

3.1. WHEN CapTitle が1つの引数（title）で呼び出された場合、THE System SHALL 全ての既存の仮想ウィンドウのキャプションを設定する。

3.2. WHEN CapTitle が1つの引数で呼び出され、仮想ウィンドウが存在しない場合、THE System SHALL エラーを発生させずに正常に終了する。

3.3. WHEN CapTitle が2つの引数（winID, title）で呼び出された場合、THE System SHALL 指定されたウィンドウのキャプションのみを設定する（既存動作を維持）。

3.4. WHEN CapTitle が存在しないウィンドウIDで呼び出された場合、THE System SHALL エラーを発生させずに正常に終了する。

3.5. THE System SHALL 空文字列（""）をタイトルとして受け入れ、ウィンドウのキャプションをクリアする。
