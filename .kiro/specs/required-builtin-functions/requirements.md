# 要件定義書

## はじめに

FILLY言語インタプリタ（son-et）に残っている必須未実装ビルトイン関数9個を実装する。対象はメッセージ制御関数2個（FreezeMes, ActivateMes）とファイルI/O関数7個（OpenF, CloseF, SeekF, ReadF, WriteF, StrReadF, StrWriteF）である。これらはサンプルスクリプトの動作に必須であり、実装しないとサンプルが正常に動作しない。

## 用語集

- **VM**: FILLY言語の仮想マシン。OpCodeを実行し、ビルトイン関数を提供する
- **EventHandler**: mes()ブロックから生成されるイベントハンドラ。Active, CurrentPC, WaitCounter等の状態を持つ
- **HandlerRegistry**: EventHandlerの登録・管理を行うレジストリ。番号によるハンドラ検索をサポートする
- **FileHandleTable**: VMが管理するファイルハンドルテーブル。整数ハンドル→*os.Fileのマッピングを保持する
- **タイトルディレクトリ**: FILLYスクリプトが格納されているディレクトリ。ファイルパスの基準となる
- **Shift-JIS**: Windows 3.1/95時代の日本語文字エンコーディング。FILLYスクリプトおよびデータファイルで使用される
- **リトルエンディアン**: バイト列の下位バイトを先頭に配置するバイトオーダー。Win16/Win32のネイティブバイトオーダー

## 要件

### 要件1: メッセージブロックの一時停止（FreezeMes）

**ユーザーストーリー:** スクリプト作者として、特定のメッセージブロックを一時停止したい。それにより、一時的にイベント処理を無効化しつつ、後で再開できるようにしたい。

#### 受け入れ基準

1. WHEN FreezeMes(mes_no) が呼び出された場合、THE VM SHALL 指定された番号のEventHandlerのActiveフラグをfalseに設定する
2. WHILE EventHandlerのActiveフラグがfalseである間、THE EventDispatcher SHALL そのハンドラのExecuteメソッド呼び出しをスキップする
3. WHEN FreezeMes で一時停止されたハンドラに対してイベントが発生した場合、THE VM SHALL そのハンドラのCurrentPCおよびWaitCounterの値を保持する
4. WHEN FreezeMes に存在しないハンドラ番号が指定された場合、THE VM SHALL エラーを返さずに警告ログを出力して処理を継続する
5. WHEN FreezeMes の引数が不足している場合、THE VM SHALL 警告ログを出力して処理を継続する

### 要件2: メッセージブロックの再開（ActivateMes）

**ユーザーストーリー:** スクリプト作者として、一時停止したメッセージブロックを再開したい。それにより、停止前の実行位置から処理を継続できるようにしたい。

#### 受け入れ基準

1. WHEN ActivateMes(mes_no) が呼び出された場合、THE VM SHALL 指定された番号のEventHandlerのActiveフラグをtrueに設定する
2. WHEN ActivateMes で再開されたハンドラに次のイベントが発生した場合、THE EventHandler SHALL 停止前のCurrentPCおよびWaitCounterの値から実行を再開する
3. WHEN ActivateMes に存在しないハンドラ番号が指定された場合、THE VM SHALL エラーを返さずに警告ログを出力して処理を継続する
4. WHEN ActivateMes の引数が不足している場合、THE VM SHALL 警告ログを出力して処理を継続する

### 要件3: ファイルハンドルテーブルの管理

**ユーザーストーリー:** VM開発者として、ファイルI/O関数が使用するファイルハンドルテーブルを管理したい。それにより、整数ハンドルによるファイル操作を安全に行えるようにしたい。

#### 受け入れ基準

1. THE FileHandleTable SHALL 整数ハンドルから*os.Fileへのマッピングを管理する
2. WHEN 新しいファイルが開かれた場合、THE FileHandleTable SHALL 未使用の最小整数ハンドル（1以上）を割り当てて返す
3. WHEN ファイルが閉じられた場合、THE FileHandleTable SHALL 対応するハンドルを解放して再利用可能にする
4. WHEN VMが停止する場合、THE FileHandleTable SHALL 開いている全てのファイルを閉じてリソースを解放する
5. WHEN 無効なハンドルでファイル操作が試みられた場合、THE VM SHALL エラーを返す

### 要件4: ファイルを開く（OpenF）

**ユーザーストーリー:** スクリプト作者として、データファイルを開きたい。それにより、ファイルの読み書きを行えるようにしたい。

#### 受け入れ基準

1. WHEN OpenF(filename) が引数1つで呼び出された場合、THE VM SHALL ファイルを読み込み専用モードで開き、整数ハンドルを返す
2. WHEN OpenF(filename, mode) が引数2つで呼び出された場合、THE VM SHALL modeフラグに従ってファイルを開き、整数ハンドルを返す
3. THE VM SHALL modeフラグのアクセス属性を以下のように解釈する: 0=読み書き、1=書き専用、2=読み専用
4. WHEN modeフラグに0x1000（新規作成フラグ）が含まれる場合、THE VM SHALL ファイルが存在しなければ新規作成し、存在すれば内容を切り詰めて開く
5. WHEN ファイルパスが相対パスの場合、THE VM SHALL タイトルディレクトリからの相対パスとして解決する（既存のresolveFilePathを使用）
6. WHEN 指定されたファイルが存在せず新規作成フラグもない場合、THE VM SHALL エラーを返す
7. WHEN OpenF の引数が不足している場合、THE VM SHALL エラーを返す

### 要件5: ファイルを閉じる（CloseF）

**ユーザーストーリー:** スクリプト作者として、開いたファイルを閉じたい。それにより、リソースを適切に解放できるようにしたい。

#### 受け入れ基準

1. WHEN CloseF(handle) が呼び出された場合、THE VM SHALL 指定されたハンドルのファイルを閉じ、ハンドルをFileHandleTableから解放する
2. WHEN CloseF に無効なハンドルが指定された場合、THE VM SHALL エラーを返す
3. WHEN CloseF の引数が不足している場合、THE VM SHALL エラーを返す

### 要件6: ファイルポインタの移動（SeekF）

**ユーザーストーリー:** スクリプト作者として、ファイル内の任意の位置に移動したい。それにより、ファイルの特定の位置からデータを読み書きできるようにしたい。

#### 受け入れ基準

1. WHEN SeekF(handle, offset, origin) が呼び出された場合、THE VM SHALL 指定されたoriginを基準にoffsetバイト分ファイルポインタを移動する
2. THE VM SHALL originを以下のように解釈する: 0=ファイル先頭（SEEK_SET）、1=現在位置（SEEK_CUR）、2=ファイル末尾（SEEK_END）
3. WHEN SeekF に無効なハンドルが指定された場合、THE VM SHALL エラーを返す
4. WHEN SeekF の引数が不足している場合、THE VM SHALL エラーを返す

### 要件7: バイナリ読み込み（ReadF）

**ユーザーストーリー:** スクリプト作者として、ファイルからバイナリデータを読み込みたい。それにより、データファイルの数値データを処理できるようにしたい。

#### 受け入れ基準

1. WHEN ReadF(handle, size) が呼び出された場合、THE VM SHALL ファイルから指定されたsizeバイト（1〜4）を読み込み、リトルエンディアンの整数値として返す
2. WHEN sizeが1の場合、THE VM SHALL 1バイトを読み込み0〜255の整数値として返す
3. WHEN sizeが2の場合、THE VM SHALL 2バイトを読み込みリトルエンディアンの16ビット整数値として返す
4. WHEN sizeが4の場合、THE VM SHALL 4バイトを読み込みリトルエンディアンの32ビット整数値として返す
5. WHEN sizeが1〜4の範囲外の場合、THE VM SHALL エラーを返す
6. WHEN ReadF に無効なハンドルが指定された場合、THE VM SHALL エラーを返す
7. WHEN ReadF の引数が不足している場合、THE VM SHALL エラーを返す

### 要件8: バイナリ書き込み（WriteF）

**ユーザーストーリー:** スクリプト作者として、ファイルにバイナリデータを書き込みたい。それにより、データファイルを作成・更新できるようにしたい。

#### 受け入れ基準

1. WHEN WriteF(handle, value, length) が3引数で呼び出された場合、THE VM SHALL valueをリトルエンディアンでlengthバイト（1〜4）分ファイルに書き込む
2. WHEN WriteF(handle, value) が2引数で呼び出された場合、THE VM SHALL valueを1バイトとしてファイルに書き込む
3. WHEN lengthが1〜4の範囲外の場合、THE VM SHALL エラーを返す
4. WHEN WriteF に無効なハンドルが指定された場合、THE VM SHALL エラーを返す
5. WHEN WriteF の引数が不足している場合、THE VM SHALL エラーを返す

### 要件9: 文字列読み込み（StrReadF）

**ユーザーストーリー:** スクリプト作者として、ファイルからテキストデータを1行ずつ読み込みたい。それにより、テキストベースのデータファイルを処理できるようにしたい。

#### 受け入れ基準

1. WHEN StrReadF(handle) が呼び出された場合、THE VM SHALL ファイルから改行区切りで1行を読み込み、文字列として返す
2. WHEN ファイルがShift-JISエンコーディングの場合、THE VM SHALL 読み込んだ文字列をUTF-8に変換して返す
3. WHEN ファイルの終端（EOF）に達した場合、THE VM SHALL 空文字列 "" を返す
4. THE VM SHALL 改行文字（CR, LF, CRLF）を行区切りとして認識し、返す文字列には改行文字を含めない
5. WHEN StrReadF に無効なハンドルが指定された場合、THE VM SHALL エラーを返す
6. WHEN StrReadF の引数が不足している場合、THE VM SHALL エラーを返す

### 要件10: 文字列書き込み（StrWriteF）

**ユーザーストーリー:** スクリプト作者として、ファイルにテキストデータを書き込みたい。それにより、テキストベースのデータファイルを作成・更新できるようにしたい。

#### 受け入れ基準

1. WHEN StrWriteF(handle, str) が呼び出された場合、THE VM SHALL 文字列をファイルに書き込む（改行は付加しない）
2. WHEN 内部文字列がUTF-8の場合、THE VM SHALL Shift-JISに変換してファイルに書き込む
3. WHEN StrWriteF に無効なハンドルが指定された場合、THE VM SHALL エラーを返す
4. WHEN StrWriteF の引数が不足している場合、THE VM SHALL エラーを返す
