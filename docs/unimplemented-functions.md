# 未実装関数一覧

このドキュメントは、`language_spec.md`で定義されているがまだ実装されていない関数の一覧です。
Windows固有の関数（PlayCD, MCI, StrMCI, SetRegStr, GetRegStr, PlayAVI）は除外しています。

最終更新: 2026-02-02

---

## 文字列関連関数

| 関数名 | 説明 | 優先度 | 使用サンプル |
|--------|------|--------|--------------|
| `StrInput(prompt, default_value)` | 文字列の入力（ダイアログ表示） | 低 | - |

---

## 配列操作関連関数

現在、未実装の配列操作関連関数はありません。

---

## 整数関連関数

現在、未実装の整数関連関数はありません。

---

## ファイル操作関連関数

### ファイル管理

| 関数名 | 説明 | 優先度 | 使用サンプル |
|--------|------|--------|--------------|
| `CopyFile(src_file, dst_file)` | ファイルのコピー | 低 | - |
| `DelFile(filename)` | ファイルの削除 | 低 | - |
| `IsExist(filename)` | ファイルの存在確認 | 中 | - |
| `MkDir(dirname)` | ディレクトリの作成 | 低 | - |
| `RmDir(dirname)` | ディレクトリの削除 | 低 | - |
| `ChDir(dirname)` | カレントディレクトリの変更 | 低 | - |
| `GetCwd()` | カレントディレクトリの取得 | 低 | - |

---

## オーディオ関連関数（リソース管理）

| 関数名 | 説明 | 優先度 | 使用サンプル |
|--------|------|--------|--------------|
| `LoadRsc(id, filename)` | リソースの読み込み | 低 | - |
| `PlayRsc(id)` | リソースの再生 | 低 | - |
| `DelRsc(id)` | リソースの削除 | 低 | - |

---

## メッセージ関連関数

現在、未実装のメッセージ関連関数はありません。

---

## システム関連関数

| 関数名 | 説明 | 優先度 | 使用サンプル |
|--------|------|--------|--------------|
| `WhatDay()` | 日付の取得 | 低 | - |
| `WhatTime()` | 時刻の取得 | 低 | - |
| `GetCmdLine()` | コマンドライン引数の取得 | 低 | - |

---

## 描画関連関数

| 関数名 | 説明 | 優先度 | 使用サンプル |
|--------|------|--------|--------------|
| `SetROP(rop_mode)` | ラスタオペレーションの設定（COPYPEN, XORPEN, MERGEPEN等） | 低 | - |

---

## 統計

- **未実装関数合計: 15個**
- **必須（サンプルで使用）: 0個**
- 高優先度: 0個
- 中優先度: 1個
- 低優先度: 14個

---

## 優先度の基準

- **必須**: サンプルスクリプトで実際に使用されている（実装しないとサンプルが動作しない）
- **高**: 基本的な機能で使用頻度が高いと予想される
- **中**: 一部のスクリプトで使用される可能性がある
- **低**: 特殊な用途、またはレガシー機能

---

## 実装済み関数（参考）

### ウィンドウ関連
`OpenWin`, `MoveWin`, `CloseWin`, `CloseWinAll`, `CapTitle`, `GetPicNo`

### ピクチャー関連
`LoadPic`, `MovePic`, `MoveSPic`, `DelPic`, `CreatePic`, `PicWidth`, `PicHeight`, `ReversePic`, `TransPic`

### キャスト関連
`PutCast`, `MoveCast`, `DelCast`

### 文字表示関連
`SetFont`, `TextWrite`, `TextColor`, `BgColor`, `BackMode`

### 描画関連
`DrawLine`, `DrawCircle`, `DrawRect`, `FillRect`, `SetLineSize`, `SetPaintColor`, `GetColor`, `SetColor`

### 文字列関連
`StrLen`, `StrPrint`, `StrCode`, `SubStr`, `StrFind`, `StrUp`, `StrLow`, `CharCode`

### 配列操作関連
`ArraySize`, `DelArrayAll`, `DelArrayAt`, `InsArrayAt`

### 整数関連
`MakeLong`, `GetHiWord`, `GetLowWord`

### ファイル操作（バイナリ/文字列I/O）
`OpenF`, `CloseF`, `SeekF`, `ReadF`, `WriteF`, `StrReadF`, `StrWriteF`

### ファイル操作（INI）
`WriteIniInt`, `GetIniInt`, `WriteIniStr`, `GetIniStr`

### オーディオ関連
`PlayMIDI`, `PlayWAVE`

### メッセージ関連
`GetMesNo`, `DelMes`, `PostMes`, `FreezeMes`, `ActivateMes`

### システム関連
`WinInfo`, `GetSysTime`, `Random`, `MsgBox`, `Debug`

### 制御関連
`del_me`, `del_us`, `del_all`, `end_step`, `Wait`, `ExitTitle`

### Windows固有（スタブ実装）
`Shell`, `MCI`, `StrMCI`
