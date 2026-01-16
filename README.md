# son-et
FILLY Script Interpreter

## 概要

son-etは、レガシーなFILLYスクリプト(Toffyスクリプト)を直接実行するインタプリタです。
開発時は即座にスクリプトを実行でき、配布時はプロジェクトを埋め込んだスタンドアロン実行ファイルを作成できます。

## 特徴

*   **インタプリタ実行**: FILLYスクリプト（`.tfy`など）を直接実行します。コンパイル不要で即座に動作確認が可能です。
*   **シングルバイナリ配布**: 画像や音楽（MIDI）などのリソースファイルを埋め込んだスタンドアロン実行ファイルを生成できます。(ただし、MIDIを再生する場合は別途SF2ファイルが必要です。)
*   **クロスプラットフォーム対応**: macOSを含む主要なプラットフォームで動作します。
*   **MIDI/Audio対応**: MIDIファイル(SMF)の再生をサポートしており、ソフトウェアシンセサイザーにより外部音源なしでBGMを再生可能です。
*   **統一されたスコープ管理**: すべてのコードがOpCodeとして実行されるため、変数スコープが一貫して管理されます。

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

### 開発時（Direct Mode）

プロジェクトディレクトリを指定して直接実行します：

```bash
son-et samples/kuma2
```

son-etは自動的にディレクトリ内のTFYファイルからmain関数を探して実行します。

### 配布時（Embedded Mode）

プロジェクトを埋め込んだスタンドアロン実行ファイルを作成するには、son-et自体をビルドする際にプロジェクトを指定します：

```bash
# ビルド時にプロジェクトを埋め込む（詳細は build-workflow.md を参照）
go build -tags embed_kuma2 -o kuma2 ./cmd/son-et

# 生成された実行ファイルを配布
./kuma2
```

## 注意事項

*   画像ファイル（BMP）や音楽ファイル（MIDI）は、TFYスクリプトと同じディレクトリに配置してください。自動的に検出されます。
*   現状、生成される仮想画面サイズは 1280x720 です。レガシーな解像度（640x480等）のウィンドウは、この仮想デスクトップ内に表示されます。
*   MIDIファイルを再生する場合は、SoundFont（.sf2）ファイルが必要です。実行ファイルと同じディレクトリに配置するか、`-sf <path>` オプションで指定してください。

## ドキュメント

プロジェクトの詳細なドキュメントは `.kiro/specs/` ディレクトリにあります：

### Core Engine
*   **[requirements.md](.kiro/specs/core-engine/requirements.md)** - 要件定義書（EARS形式）
*   **[design.md](.kiro/specs/core-engine/design.md)** - 設計書（アーキテクチャ、コンポーネント、正確性プロパティ）
*   **[architecture.md](.kiro/specs/core-engine/architecture.md)** - アーキテクチャ詳細（実装パターン、デバッグガイド）
*   **[implementation-status.md](.kiro/specs/core-engine/implementation-status.md)** - 実装状況レポート
*   **[tasks.md](.kiro/specs/core-engine/tasks.md)** - 実装タスクリスト

### Interpreter Architecture
*   **[requirements.md](.kiro/specs/interpreter-architecture/requirements.md)** - インタプリタアーキテクチャ要件定義書

## 謝辞

FILLYの作者である内田友幸さん、またFILLYを盛り上げたFFILLYのコミュニティの皆様方に感謝申し上げます。

