---
inclusion: manual
---

# サンプル実行方法

## 基本コマンド（ワンライナー）

```bash
T=30; LOG=kuma2_test.log; SAMPLE=samples/kuma2; go run ./cmd/son-et --timeout $T --log-level debug $SAMPLE > $LOG 2>&1 & PID=$!; sleep $T; kill $PID 2>/dev/null; cat $LOG
```

## 変数の説明

| 変数 | 説明 | 例 |
|------|------|-----|
| `T` | タイムアウト秒数 | `30`, `60` |
| `LOG` | ログファイル名 | `kuma2_test.log` |
| `SAMPLE` | サンプルディレクトリ | `samples/kuma2` |


## 実行例

```bash
# kuma2を30秒で実行
T=30; LOG=kuma2_test.log; SAMPLE=samples/kuma2; go run ./cmd/son-et --timeout $T --log-level debug $SAMPLE > $LOG 2>&1 & PID=$!; sleep $T; kill $PID 2>/dev/null; cat $LOG

# y_saruを65秒で実行(Castの場面が65秒程度実行しないと完了しない)
T=65; LOG=y_saru_test.log; SAMPLE=samples/y_saru; go run ./cmd/son-et --timeout $T --log-level debug $SAMPLE > $LOG 2>&1 & PID=$!; sleep $T; kill $PID 2>/dev/null; cat $LOG
```

## コマンドの動作

1. `go run` でビルド＆実行（ビルド忘れ防止）
2. アプリの `--timeout` オプションで内部タイムアウト
3. `sleep $T; kill $PID` で外部からも強制終了（暴走対策）
4. 全出力をログファイルに保存
5. 最後に `cat $LOG` でログを表示

## 注意事項

- macOSでは `timeout` コマンドが使えないため、このパターンを使用すること
- 暴走時は `kill $PID` で強制終了される
