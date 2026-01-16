# 埋め込みビルドガイド

このドキュメントでは、FILLYプロジェクトのスタンドアロン実行ファイルを作成する方法を説明します。

## 概要

son-etインタプリタは2つの実行モードをサポートしています：

1. **Direct Mode（直接実行モード）** - 開発時にディレクトリからTFYプロジェクトを直接実行
2. **Embedded Mode（埋め込みモード）** - 配布用にプロジェクトを埋め込んだスタンドアロン実行ファイルを作成

## 埋め込みビルドの作成

### クイックスタート

ビルドスクリプトを使用して埋め込み実行ファイルを作成します：

```bash
./build_embedded.sh kuma2
```

これにより、すべてのアセットを含むスタンドアロン実行ファイル `kuma2` が作成され、外部ファイルなしで配布できます。

### 手動ビルド

手動で埋め込みビルドを作成する場合：

```bash
go build -tags embedded -o kuma2 ./samples/kuma2
```

## 埋め込みビルド用のプロジェクト構造

プロジェクトを埋め込み可能にするには、以下の構造の `embedded_main.go` ファイルが必要です：

```go
//go:build embedded
// +build embedded

package main

import (
    "embed"
    // ... その他のインポート
)

// カレントディレクトリのすべてのファイルを埋め込む
//go:embed *
var embeddedFS embed.FS

func main() {
    // 埋め込まれたプロジェクトを実行
    // 完全な実装は samples/kuma2/embedded_main.go を参照
}
```

### 重要なポイント

1. **ビルドタグ**: `//go:build embedded` を使用して、`-tags embedded` でビルドする時のみコンパイルされるようにします

2. **他のmain関数の除外**: プロジェクトに他の `main()` 関数を持つGoファイル（テストハーネスなど）がある場合、ビルドタグで除外します：
   ```go
   //go:build !embedded
   // +build !embedded
   ```

3. **埋め込みディレクティブ**: `//go:embed *` を使用してプロジェクトディレクトリ内のすべてのファイルを埋め込みます

4. **アセット読み込み**: 埋め込み実行ファイルは `EmbedFSAssetLoader` を使用して埋め込みファイルシステムからアセットを読み込みます

## 例：kuma2プロジェクト

kuma2サンプルプロジェクトは埋め込みビルド構造を示しています：

```
samples/kuma2/
├── embedded_main.go      # 埋め込みビルドのエントリポイント（ビルドタグ: embedded）
├── kuma2_game.go         # テストハーネス（ビルドタグ: !embedded）
├── KUMA2.TFY             # FILLYスクリプト
├── KUMA2.FIL             # FILLYスクリプト
├── TITLE.BMP             # 画像アセット
├── KUMA-1.BMP
├── KUMA-2.BMP
├── ...
└── KUMA.MID              # オーディオアセット
```

### kuma2のビルド

```bash
# ビルドスクリプトを使用
./build_embedded.sh kuma2

# または手動で
go build -tags embedded -o kuma2 ./samples/kuma2
```

### 埋め込み実行ファイルの実行

```bash
./kuma2
```

実行ファイルはコマンドライン引数なしで実行でき、すべてのアセットを埋め込みファイルシステムから読み込みます。

## 新しい埋め込みプロジェクトの作成

新しいプロジェクトの埋め込みビルドを作成するには：

1. **プロジェクトディレクトリを作成** `samples/` 内に：
   ```bash
   mkdir samples/my_project
   ```

2. **FILLYスクリプトとアセットを追加**：
   ```
   samples/my_project/
   ├── script.tfy
   ├── image1.bmp
   ├── music.mid
   └── ...
   ```

3. **embedded_main.goをコピーして調整** kuma2から：
   ```bash
   cp samples/kuma2/embedded_main.go samples/my_project/
   ```

4. **プロジェクト名を更新** embedded_main.go内で必要に応じて（ログ出力用）

5. **埋め込み実行ファイルをビルド**：
   ```bash
   ./build_embedded.sh my_project
   ```

## 埋め込みビルドの利点

- **単一ファイル配布**: すべてのアセットが実行ファイルに埋め込まれます
- **外部依存なし**: アセットファイルを別途配布する必要がありません
- **簡単なデプロイ**: 実行ファイルをコピーするだけでプロジェクトを実行できます
- **アセット保護**: アセットがバイナリに埋め込まれます（基本的な難読化）

## 制限事項

- **実行ファイルサイズが大きい**: すべてのアセットを含むため（通常50-100MB）
- **実行時の変更不可**: 再ビルドせずにアセットを変更できません
- **ビルド時間**: 大きなアセットの埋め込みによりビルド時間が増加します

## トラブルシューティング

### ビルドエラー

**エラー: "main redeclared in this block"**
- 解決策: 他の `main()` 関数を持つGoファイルに `//go:build !embedded` を追加

**エラー: "pattern ... invalid pattern syntax"**
- 解決策: `//go:embed` ディレクティブが `..` を含まない相対パスを使用していることを確認

### 実行時エラー

**エラー: "no TFY files found in embedded project"**
- 解決策: TFYファイルがプロジェクトディレクトリにあり、`.gitignore` で除外されていないことを確認

**エラー: "failed to read embedded file"**
- 解決策: 参照されているすべてのアセットがプロジェクトディレクトリにあることを確認

## 関連ドキュメント

- [README.md](README.md) - son-etの一般的なドキュメント
- [build-workflow.md](.kiro/steering/build-workflow.md) - ビルドと実行のワークフロー
- [development-workflow.md](.kiro/steering/development-workflow.md) - 開発ワークフロー
