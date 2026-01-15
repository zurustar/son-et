# son-et
Filly Scenario (toffy) to Golang Transpiler

## 概要

son-etは、レガシーなFILLYスクリプト(Toffyスクリプト)を、現代的なGo言語のソースコードに変換（トランスパイル）するツールです。
生成されたGoコードは、マルチプラットフォーム（主にmacOS）で動作するネイティブアプリケーションとしてビルドできます。

## 特徴

*   **Source-to-Source コンパイル**: FILLYスクリプト（`.tfy`など）を解析し、Go言語のソースコードを生成します。
*   **シングルバイナリ**: 画像や音楽（MIDI）などのリソースファイルは、`//go:embed` を使用して実行ファイルに埋め込まれるため、配布が容易です。(ただし、MIDIを再生する場合は別途SF2ファイルが必要です。)
*   **クロスプラットフォーム対応**: macOSを含む主要なプラットフォームで動作するネイティブアプリケーションを生成します。
*   **MIDI/Audio対応**: MIDIファイル(SMF)の再生をサポートしており、ソフトウェアシンセサイザーにより外部音源なしでBGMを再生可能です。

## 動作環境

*   **Go**: バージョン 1.24 以上

## インストール

```bash
git clone https://github.com/zurustar/son-et.git
cd son-et
go install ./cmd/son-et
```

`son-et` コマンドが `$GOPATH/bin` にインストールされます。パスが通っていることを確認してください。

## 使い方

1.  変換したいスクリプトファイル（例: `sample.tfy`）と画像や音声ファイルなどのアセット一式を用意します。
2.  スクリプトファイルがあるディレクトリで、以下のコマンドを実行します。

```bash
son-et ./sample.tfy
```

3.  同じディレクトリに `sample_game.go` というGoのソースコードが生成されます。
4.  生成されたGoコードを実行またはビルドします。

```bash
# 直接実行する場合
go run sample_game.go

# ビルドして実行ファイルを作成する場合
go build -o mygame sample_game.go
./mygame
```

## 注意事項

*   画像ファイル（BMP）や音楽ファイル（MIDI）は、変換元のスクリプトファイルと同じディレクトリに配置してください。コンパイル時に自動的に検出され、埋め込まれます。
*   現状、生成される仮想画面サイズは 1280x720 です。レガシーな解像度（640x480等）のウィンドウは、この仮想デスクトップ内に表示されます。
*   MIDIファイルを再生する場合は、SoundFont（.sf2）ファイルが必要です。実行ファイルと同じディレクトリに配置するか、`-sf <path>` オプションで指定してください。

## ドキュメント

プロジェクトの詳細なドキュメントは `.kiro/specs/core-engine/` ディレクトリにあります：

*   **[requirements.md](.kiro/specs/core-engine/requirements.md)** - 要件定義書（EARS形式）
*   **[design.md](.kiro/specs/core-engine/design.md)** - 設計書（アーキテクチャ、コンポーネント、正確性プロパティ）
*   **[architecture.md](.kiro/specs/core-engine/architecture.md)** - アーキテクチャ詳細（実装パターン、デバッグガイド）
*   **[implementation-status.md](.kiro/specs/core-engine/implementation-status.md)** - 実装状況レポート
*   **[tasks.md](.kiro/specs/core-engine/tasks.md)** - 実装タスクリスト

## 謝辞

FILLYの作者である内田友幸さん、またFILLYを盛り上げたFFILLYのコミュニティの皆様方に感謝申し上げます。

