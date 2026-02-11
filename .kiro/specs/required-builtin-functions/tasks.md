# 実装計画: 必須ビルトイン関数

## 概要

メッセージ制御関数2個とファイルI/O関数7個を実装する。メッセージ制御は既存パターンに従い軽量に実装し、ファイルI/OはFileHandleTableを新規導入してから各関数を順次実装する。TDDサイクル（Red→Green→Refactor）に従う。

## タスク

- [x] 1. メッセージ制御関数の実装（FreezeMes, ActivateMes）
  - [x] 1.1 FreezeMes/ActivateMesを `pkg/vm/builtins_system.go` に実装する
    - `handlerRegistry.GetHandlerByNumber(seqID)` でハンドラを取得
    - FreezeMes: `handler.Active = false`（Remove()は呼ばない）
    - ActivateMes: `handler.Active = true`
    - 引数不足・存在しないハンドラ番号は警告ログのみ（既存のDelMes/GetMesNoパターンに従う）
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 2.1, 2.2, 2.3, 2.4_
  - [x] 1.2 FreezeMes/ActivateMesのユニットテストを `pkg/vm/builtins_message_test.go` に作成する
    - 正常系: FreezeMesでActive=false、ActivateMesでActive=true
    - 引数不足テスト
    - 存在しないハンドラ番号テスト
    - 冪等性テスト（既にfreezeされたハンドラへのFreezeMes）
    - _Requirements: 1.1, 1.4, 1.5, 2.1, 2.3, 2.4_
  - [x] 1.3 FreezeMes/ActivateMesのプロパティテストを `pkg/vm/builtins_message_property_test.go` に作成する
    - **Property 1: FreezeMesはハンドラを無効化する**
    - **Validates: Requirements 1.1, 1.2**
    - **Property 2: Freeze/Activateラウンドトリップはハンドラ状態を保持する**
    - **Validates: Requirements 1.3, 2.1, 2.2**

- [x] 2. チェックポイント - メッセージ制御関数
  - 全テストがパスすることを確認し、疑問があればユーザーに確認する。

- [x] 3. FileHandleTableの実装
  - [x] 3.1 `pkg/vm/file_handle_table.go` にFileHandleTable構造体を実装する
    - `fileEntry` 構造体（`*os.File` + `*bufio.Reader`）
    - `Open(file *os.File) int` — 未使用の最小整数ハンドル（1以上）を割り当て
    - `Get(handle int) (*fileEntry, error)` — ハンドルからエントリを取得
    - `Close(handle int) error` — ファイルを閉じてハンドルを解放
    - `CloseAll()` — 全ファイルを閉じる
    - `ResetReader(handle int)` — SeekF時にbufio.Readerをリセット
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5_
  - [x] 3.2 FileHandleTableのユニットテストを `pkg/vm/file_handle_table_test.go` に作成する
    - Open/Get/Close/CloseAllの基本操作テスト
    - 無効なハンドルへのGet/Closeテスト
    - CloseAll後の全ハンドル無効化テスト
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5_
  - [x] 3.3 FileHandleTableのプロパティテストを `pkg/vm/file_handle_table_property_test.go` に作成する
    - **Property 3: FileHandleTableのハンドル割り当て不変条件**
    - **Validates: Requirements 3.2, 3.3**

- [x] 4. VMへのFileHandleTable統合
  - [x] 4.1 VM structに `fileHandleTable *FileHandleTable` フィールドを追加し、New()で初期化する
    - `registerDefaultBuiltins()` に `vm.registerFileIOBuiltins()` 呼び出しを追加
    - VM停止時のクリーンアップに `fileHandleTable.CloseAll()` を追加
    - _Requirements: 3.4_

- [x] 5. チェックポイント - FileHandleTable
  - 全テストがパスすることを確認し、疑問があればユーザーに確認する。

- [x] 6. ファイルI/O関数の実装（OpenF, CloseF, SeekF）
  - [x] 6.1 `pkg/vm/builtins_fileio.go` に `registerFileIOBuiltins()` メソッドとOpenF/CloseF/SeekFを実装する
    - OpenF: modeフラグ定数の定義、アクセス属性解釈、0x1000新規作成フラグ、resolveFilePath使用
    - CloseF: FileHandleTable.Close()呼び出し
    - SeekF: origin定数のマッピング（0→io.SeekStart, 1→io.SeekCurrent, 2→io.SeekEnd）、bufio.Readerリセット
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 4.6, 4.7, 5.1, 5.2, 5.3, 6.1, 6.2, 6.3, 6.4_
  - [x] 6.2 OpenF/CloseF/SeekFのユニットテストを `pkg/vm/builtins_fileio_test.go` に作成する
    - OpenFの各モード（読み込み専用、書き込み専用、読み書き、新規作成）
    - OpenFのファイル不存在エラー
    - OpenFの引数不足テスト
    - CloseFの正常系・無効ハンドル・引数不足テスト
    - SeekFの各origin（SEEK_SET, SEEK_CUR, SEEK_END）テスト
    - SeekFの無効ハンドル・引数不足テスト
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.6, 4.7, 5.1, 5.2, 5.3, 6.1, 6.2, 6.3, 6.4_

- [x] 7. バイナリI/O関数の実装（ReadF, WriteF）
  - [x] 7.1 `pkg/vm/builtins_fileio.go` にReadF/WriteFを追加する
    - ReadF: `encoding/binary.LittleEndian` でバイト列→整数変換、size 1〜4のみ許可
    - WriteF: 2引数版（1バイト）と3引数版（lengthバイト）、リトルエンディアン書き込み
    - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5, 7.6, 7.7, 8.1, 8.2, 8.3, 8.4, 8.5_
  - [x] 7.2 ReadF/WriteFのユニットテストを `pkg/vm/builtins_fileio_test.go` に追加する
    - ReadFのsize範囲外テスト
    - WriteFの引数不足・無効ハンドルテスト
    - 具体的なバイト値の読み書きテスト
    - _Requirements: 7.5, 7.6, 7.7, 8.3, 8.4, 8.5_
  - [x] 7.3 ReadF/WriteFのプロパティテストを `pkg/vm/builtins_fileio_property_test.go` に作成する
    - **Property 4: ReadF/WriteFバイナリラウンドトリップ**
    - **Validates: Requirements 7.1, 7.2, 7.3, 7.4, 8.1, 8.2**
    - **Property 6: SeekFによるランダムアクセスの整合性**
    - **Validates: Requirements 6.1, 6.2**

- [x] 8. チェックポイント - バイナリI/O
  - 全テストがパスすることを確認し、疑問があればユーザーに確認する。

- [x] 9. 文字列I/O関数の実装（StrReadF, StrWriteF）
  - [x] 9.1 `pkg/vm/builtins_fileio.go` にStrReadF/StrWriteFを追加する
    - StrReadF: `bufio.Reader` で1行読み込み、`japanese.ShiftJIS.NewDecoder()` でUTF-8変換、EOF時は空文字列
    - StrWriteF: `japanese.ShiftJIS.NewEncoder()` でShift-JIS変換、改行は付加しない
    - fileEntryのreaderフィールドの遅延初期化
    - _Requirements: 9.1, 9.2, 9.3, 9.4, 9.5, 9.6, 10.1, 10.2, 10.3, 10.4_
  - [x] 9.2 StrReadF/StrWriteFのユニットテストを `pkg/vm/builtins_fileio_test.go` に追加する
    - StrReadFのEOF時の空文字列返却テスト
    - StrReadFの各改行パターン（CR, LF, CRLF）テスト
    - StrWriteFの改行非付加テスト
    - Shift-JIS変換テスト（日本語文字列の読み書き）
    - 引数不足・無効ハンドルテスト
    - _Requirements: 9.2, 9.3, 9.4, 9.5, 9.6, 10.2, 10.3, 10.4_
  - [x] 9.3 StrReadF/StrWriteFのプロパティテストを `pkg/vm/builtins_fileio_property_test.go` に追加する
    - **Property 5: StrWriteF/StrReadFテキストラウンドトリップ**
    - **Validates: Requirements 9.1, 9.2, 10.1, 10.2**

- [x] 10. ドキュメントの更新
  - [x] 10.1 `docs/unimplemented-functions.md` を更新する
    - メッセージ関連関数セクションからFreezeMes/ActivateMesを削除し、実装済み関数セクションに移動
    - ファイル操作関連関数セクションからOpenF/CloseF/SeekF/ReadF/WriteF/StrReadF/StrWriteFを削除し、実装済み関数セクションに移動
    - 統計の数値を更新（未実装関数合計、必須の数）
  - [x] 10.2 `docs/language-spec.md` のファイルI/Oセクションに詳細仕様を追記する
    - OpenFのmodeフラグの詳細（アクセス属性、新規作成フラグ0x1000）
    - ReadFのsizeが1〜4バイトであること、リトルエンディアン整数値を返すこと
    - WriteFの2引数版と3引数版の違い
    - SeekFのorigin値（0=先頭, 1=現在位置, 2=末尾）
    - StrReadFのEOF時の空文字列返却、改行の扱い
    - StrWriteFの改行非付加

- [x] 11. 最終チェックポイント
  - 全テストがパスすることを確認し、疑問があればユーザーに確認する。

## 備考

- `*` マーク付きのタスクはオプション（スキップ可能）
- 各タスクは特定の要件にトレースされている
- チェックポイントで段階的に検証を行う
- プロパティテストは普遍的な正当性を検証し、ユニットテストは具体例とエッジケースを検証する
- `go test` は必ず `-timeout` オプション付きで実行すること
