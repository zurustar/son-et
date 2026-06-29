# バグ調査メモ (2026-06-29 開始)

機能単位でバグチェックを進める。各タスクの結果をここに逐次記録する（セッションが切れても残すため）。

## 確定バグ一覧（severity順・要修正）

| # | severity | 場所 | 概要 |
|---|----------|------|------|
| A | 🔴 最高 | [vm.go executeFor/executeWhile](../pkg/vm/vm.go#L839-L950) | **ループ内 `return`/`wait` が伝播せず握り潰される**。ループが最後まで回り、戻り値が後続反復で上書き消失。早期return系が全滅 |
| B | 🟠 高 | [preprocessor.go expandIncludes](../pkg/compiler/preprocessor/preprocessor.go#L144-L197) | コメント/文字列内に `#include` テキストがあると展開が壊れる（lexer検出と生文字列検索の二重走査の不整合） |
| C | 🟡 中 | [parser.go parseGenericDirective](../pkg/compiler/parser/parser.go#L1571-L1575) | 未知 `#`ディレクティブが直後の文の先頭トークンを食う（黙って壊れる） |
| D | 🟡 中 | [midi.go ParseMIDITempoMap](../pkg/vm/audio/midi.go#L614-L704) | 不正/切り詰めMIDIで範囲外アクセス panic（DoS） |
| E | 🟡 中 | [bmp.go DecodeBMP](../pkg/graphics/bmp.go#L94-L171) | 負の幅/巨大寸法のBMPで makeslice/OOM panic（DoS） |
| F | 🟡 中 | [sprite.go AddChild](../pkg/sprite/sprite.go#L130-L140) | 親子サイクルのガード無し→再帰走査でスタックオーバーフロー（プロセス即死） |

詳細・再現条件・修正案は各タスクのセクション参照。軽微な所見（lexer のエラー報告漏れ、scope のロック窓、イベント順序など）も各セクションに記載。


## 進捗

- [x] 1. Lexer (pkg/compiler/lexer)
- [x] 2. Parser (pkg/compiler/parser)
- [x] 3. Preprocessor (pkg/compiler/preprocessor)
- [x] 4. Compiler/codegen (pkg/compiler/compiler)
- [x] 5. VM core (pkg/vm) ※コア完了。builtins_system/graphics/fileio は未走査
- [x] 6. VM audio/MIDI (pkg/vm/audio)
- [x] 7. Graphics (pkg/graphics) ※BMP解析＋要所スポット。Ebitengine描画/sprite全体は精査未了
- [x] 8. Sprite (pkg/sprite) ※コア sprite.go 完了。各 *_sprite.go サブタイプは軽め
- [x] 9. App/CLI/window/title 周辺 ※app.go/cli.go 完了。window/title/script は軽め

---

## 1. Lexer (pkg/compiler/lexer) — 完了

テスト: `go test ./pkg/compiler/lexer/` → PASS（ベースライン緑）。

### 検出した問題（いずれも低〜中severity、クラッシュはしない）

1. **未終端の文字列リテラルがエラーにならず黙って受理される**
   - 場所: [lexer.go:330-377](../pkg/compiler/lexer/lexer.go#L330-L377) `readString`
   - `"abc`（閉じ引用符なし）でEOFに達しても、閉じ引用符を読み飛ばすだけ（L367-369）でエラーを出さない。`TokenizeWithErrors` も ILLEGAL トークンしか拾わないため検出されない。
   - 検証済み: `x = "abc` → lexerエラー 0 件。
   - さらに `"abc\`（末尾がバックスラッシュ）だと、エスケープ処理でEOF(0)を文字列に NUL バイトとして混入させる副作用あり（L355-359 default 分岐）。

2. **未終端のマルチラインコメントがエラーにならない**
   - 場所: [lexer.go:71-90](../pkg/compiler/lexer/lexer.go#L71-L90) `skipMultiLineComment`
   - `/* never closed` がEOFまで黙って読み飛ばされ、エラー報告なし。
   - 検証済み: `a /* never closed` → エラー 0 件、トークン 2 個（a と EOF）。

3. **`0x`（16進プレフィックスのみ・桁なし）が INT トークンになる**
   - 場所: [lexer.go:168-189](../pkg/compiler/lexer/lexer.go#L168-L189) `readHexNumber`
   - リテラル `"0x"` の INT トークンを生成。ただし下流の [parser.go:594](../pkg/compiler/parser/parser.go#L594) `strconv.ParseInt(literal[2:], 16, 64)` が空文字でエラーを返し、`could not parse "0x" as integer` として報告されるため**実害は軽微**（クラッシュしない）。lexer 段で弾けると親切。

### 所感
lexer 本体（位置追跡・演算子・キーワード判定）は概ね健全。上記はエラー報告の網羅性の問題が中心。修正は任意。

---

## 2. Parser (pkg/compiler/parser) — 完了

テスト: `go test ./pkg/compiler/parser/` → PASS（ベースライン緑）。

### 検出した問題

1. **【確定・中〜高severity】未知のプリプロセッサディレクティブが直後の文の先頭トークンを食う**
   - 場所: [parser.go:1571-1575](../pkg/compiler/parser/parser.go#L1571-L1575) `parseGenericDirective`
   - パーサの規約は「各 parse 関数は最後に消費したトークン上に cur を残し、`ParseProgram` 側の `p.nextToken()`（L253）が次へ進める」。ところが `parseGenericDirective` は自分で `p.nextToken()` してから nil を返すため、`ParseProgram` の `nextToken` と合わせて**2トークン進んでしまう**。結果、未知ディレクティブの直後の文の先頭トークンが消える。
   - 検証済み: `#foo bar\nint x;\nint y;` をパース → 期待は VarDeclaration×2 だが、実際は `[0] ExpressionStatement x`（`int` が食われ単なる式 `x` に化けた）+ `[1] VarDeclaration y`。エラー報告もなし（黙って壊れる）。
   - 修正案: `parseGenericDirective` の `p.nextToken()` を削除する（ディレクティブは単一トークンなので、cur をディレクティブ上に残せば `ParseProgram` 側で正しく1つ進む）。

2. **【軽微】パーサが lexer のエラーを捨てている**
   - 場所: [parser.go:99](../pkg/compiler/parser/parser.go#L99) `tokens, _ := l.Tokenize()`
   - `TokenizeWithErrors` ではなく `Tokenize` を使うため lexer エラーを破棄。不正文字 `@` 等は後段で `no prefix parse function for ILLEGAL found` として一応報告されるが、lexer 側の「未終端文字列/コメント」（タスク1の所見1・2）はパースを通しても完全に黙殺されたまま。位置・メッセージの精度も落ちる。

### 所感
for/if/while/switch、関数定義判定（先読み+位置復元）、代入判定など制御フローのトークン管理は丁寧で堅牢。`parseGenericDirective` だけが規約を破っている。確定バグ1件は修正推奨。

---

## 3. Preprocessor (pkg/compiler/preprocessor) — 完了

テスト: `go test ./pkg/compiler/preprocessor/` → PASS（ベースライン緑）。

### 検出した問題

1. **【確定・高severity】コメント/文字列内に `#include` というテキストがあると展開が壊れる**
   - 場所: [preprocessor.go:144-197](../pkg/compiler/preprocessor/preprocessor.go#L144-L197) `expandIncludes` と [L230-238](../pkg/compiler/preprocessor/preprocessor.go#L230-L238) `findDirectiveStart`
   - 原因: INCLUDE ディレクティブの**検出**は lexer（コメント/文字列を正しくスキップ）で行うのに、ディレクティブの**位置特定**は `findDirectiveStart` = 素朴な `strings.Index(source, "#include")` で行う。2つの走査が食い違い、コメント内の `#include` を先に拾ってしまう。
   - 検証済み入力:
     ```
     // see #include "fake.tfy"
     #include "real.tfy"
     int after;
     ```
     出力:
     ```
     // see REAL_CONTENT;          ← real.tfy の中身がコメント行内に挿入される（＝コメントアウトされ無効化）
     #include "real.tfy"          ← 本物のディレクティブが未処理のまま残る
     int after;
     ```
   - 正常系（コメントなしの複数 include）は正しく動作することを別途確認済み（バグはコメント/文字列内 `#include` 限定）。
   - 修正案: lexer が返す INCLUDE トークンの位置情報（Line/Column）を使って置換範囲を決める、または lexer のトークン走査だけで再構成し、`findDirectiveStart` の素朴な文字列検索をやめる。

### その他（軽微・要検討）

2. **1文字ファイル名が無視される**: [extractIncludeFilename L206](../pkg/compiler/preprocessor/preprocessor.go#L206) `len(rest) < 2` で空文字を返す。`#include "a"` のような1文字名は弾かれる（実害は薄い）。

3. **UTF-8 ファイルが Shift-JIS として誤デコードされる可能性**: [readFileWithEncoding L262-279](../pkg/compiler/preprocessor/preprocessor.go#L262-L279)。Shift-JIS デコードは UTF-8 入力でもエラーを返さないことが多く、その場合フォールバックが効かず文字化けしうる。.TFY が Shift-JIS 前提なら設計通りだが、混在時は注意。

### 所感
循環参照検出・include ガード・スタック管理（defer pop）は正しい。最大の問題は検出（lexer）と位置特定（生文字列検索）の二重走査による不整合。確定バグ1件は修正推奨。

---

## 4. Compiler/codegen (pkg/compiler/compiler) — 完了

テスト: `go test ./pkg/compiler/compiler/` → PASS（ベースライン緑）。

### 検出した問題

1. **【確定・中severity / 要VM確認】配列のサイズ指定が完全に捨てられる**
   - 場所: [compiler.go:252-279](../pkg/compiler/compiler/compiler.go#L252-L279) `compileVarDeclaration`。パーサは `int arr[10]` のサイズ式を `vd.Sizes` に格納するが、コンパイラは `vd.Sizes` を**一度も参照しない**（grep で確認: `Sizes` の参照ゼロ）。
   - 検証済み: `int arr[10];` と `int arr[];` がどちらも同一の `Assign [arr []]`（空配列）にコンパイルされる。サイズ `10` は消失。
   - 影響: VM が添字代入で配列を自動拡張するなら実害は小さいが、未代入の `arr[5]` を**読み取る**コードがあると範囲外アクセスになりうる。→ タスク5(VM)で配列の自動拡張/範囲外読み取り挙動を要確認。

### その他（軽微・VM依存・要検討）

2. **`return` を `Call "return"` として表現**: [compileReturnStatement L724-733](../pkg/compiler/compiler/compiler.go#L724-L733)。VM が組み込み関数名 `return` と衝突せず特別扱いする前提。VMで要確認。

3. **`for(;;)` の条件が `nil` になる**: [compileForStatement L497-500](../pkg/compiler/compiler/compiler.go#L497-L500)。条件省略時 `condition` は `any(nil)`。VM が nil 条件を「常に真」と解釈しないと無限ループ for が壊れる。VMで要確認。

4. **`str` 配列の要素デフォルトが未定義**: 配列は型に関係なく `[]any{}` で初期化。要素の型別デフォルト（"" vs 0）は動的依存。

### 所感
AST→opcode の素直な変換器で、ロジックは概ね健全。確定の問題は配列サイズdrop 1件。残りはVMの挙動次第なのでタスク5で裏取りする。
→ タスク5で確認: 配列は自動拡張・範囲外読み取りは0を返す（[array.go](../pkg/vm/array.go), [executor.go:712-788](../pkg/vm/executor.go#L712-L788)）ため、配列サイズdropは**挙動上ほぼ無害**と判明。`return`/`nil条件`もVM側で対応あり（ただし下記の重大バグあり）。

---

## 5. VM core (pkg/vm) — コア完了（一部builtins未走査）

テスト: `go test ./pkg/vm/` → PASS（ベースライン緑）。

### 検出した問題

1. **【確定・最高severity】`for`/`while` ループ内の `return` が効かない（ループが最後まで回り、return値が消失する）**
   - 場所: [vm.go:839-900 `executeFor`](../pkg/vm/vm.go#L839-L900) と [vm.go:907-950 `executeWhile`](../pkg/vm/vm.go#L907-L950)
   - 原因: 両関数はループ本体の実行結果を `breakSignal` / `continueSignal` だけ判定し、**`returnMarker` と `waitMarker` を判定していない**。`executeBlock`（[vm.go:1051-1090](../pkg/vm/vm.go#L1051-L1090)）は4種すべて伝播するのに、ループ側が return/wait を素通りさせてしまう。
   - 挙動: `return` が来てもループは break せず、`lastResult = result`（＝returnMarker）に入れたまま**次の反復を継続**。後続の反復が non-return の結果を返すと `lastResult` が**上書きされ、return が完全に消失**する。さらに return すべき後のイテレーションの副作用が全部実行される。
   - 検証済み（手組みopcode）:
     - 無条件 `return` を含む `for(i=0;i<10;i++)` → i が 10 まで回りきった（本来は即 return で i=0 のはず）。
     - 条件付き `if(i==2) return 99;` + `hits[i]=i` を含む `for(i=0;i<5;i++)` → 戻り値は `returnMarker(99)` ではなく `int64(4)`（**return が消失**）。`hits=[0 1 0 3 4]`（i=3,4 の副作用が return 後に実行された）。
   - 影響: ループ内 `return` を使う関数すべてが誤動作。「最初に見つかった要素を返す」「条件成立で早期 return」等の頻出パターンが壊れる。実スクリプトへの影響大。
   - 修正案: `executeFor`/`executeWhile` の本体実行後に、break/continue 判定に加えて以下を追加（`executeBlock` と同様に即 return で伝播）:
     ```go
     if _, isReturn := result.(*returnMarker); isReturn { return result, nil }
     if _, isWait   := result.(*waitMarker);   isWait   { return result, nil }
     ```
   - 補足: `waitMarker`（step/アニメ用の待機）も同様に伝播されないため、ループ内 `wait` が機能しない可能性が高い。step のセマンティクスと突き合わせて要確認。

### その他（中〜軽微）

2. **`Scope.Set` の unlock→再lock がデータ競合の窓を作る**: [scope.go:64-87](../pkg/vm/scope.go#L64-L87)。親スコープ探索のため一度 `s.mu` を手動 Unlock して再 Lock する。ロック獲得は常に child→parent 方向で一貫しており（デッドロックは起きない）この Unlock は不要に見えるが、その間に別 goroutine が同名変数を生成すると親子両方に書く不整合が起きうる。イベントハンドラが別 goroutine で走る設計（RWMutex の存在が示唆）なら顕在化しうる。

3. **大文字小文字無視のビルトイン/ユーザー関数マッチが map を線形走査**: [executor.go:314-337](../pkg/vm/executor.go#L314-L337)。完全一致で見つからない場合に map 全体を走査。呼び出しごとに O(n) でありホットパスだと遅い。また map 走査順は不定なので、大小違いで複数候補があると非決定的（通常は無いはず）。

4. **【軽微】イベントキューのソートが非安定（同一タイムスタンプの順序が崩れる）**: [event.go:163-188 `Push`](../pkg/vm/event.go#L163-L188)。`sort.Slice`（非安定）で push の度に全体ソート。TIME イベントを短時間に大量 push するとクロック分解能以下で同一 `time.Now()` になり、登録/到着順（FIFO）が保証されない。要件 1.1/1.5（時系列・登録順）に反しうる。`sort.SliceStable` か単調増加シーケンス番号での比較が望ましい。push 毎の全体ソートは O(n log n) で効率も悪い。

### 確認できた“非バグ”（参考）
- `executeWait` は `waitMarker` を返し、`executeBlock`→ハンドラ `Execute` の PC 退避で正しく一時停止/再開する（[event.go:313-421](../pkg/vm/event.go#L313-L421)）。ただし**ループ内の wait は finding #1 により伝播しない**ため機能しない。
- `Run` の二重処理回避（[vm.go:443](../pkg/vm/vm.go#L443)）は `opcode.Cmd == "DefineFunction"` の文字列リテラル比較だが、`opcode.Cmd` は `type Cmd string` で値も一致するため**正しく動作**（スタイルは脆い）。
- 配列は自動拡張、範囲外読み取りは0返し（[executor.go:712-788](../pkg/vm/executor.go#L712-L788)）。
- builtins（array/string/math）の境界処理は概ね正常。`StrCode` の2バイトコードは生バイト連結で UTF-8 的に不正な文字列を生む可能性（Shift-JIS前提なら設計通り）。

### 未走査（次セッション）
builtins_system.go(723) / builtins_graphics.go(1062) / builtins_fileio.go(441) / builtins_audio.go / callUserFunction の動的スコープ詳細。グラフィック系はタスク7と重複するのでそちらで。

### 所感
VM core 最大の問題は **finding #1（ループ内 return/wait 不伝播）で最高severity**。これが最優先修正対象。他はデータ競合の窓・イベント順序など中〜軽微。

---

## 6. VM audio/MIDI (pkg/vm/audio) — 完了

テスト: `go test ./pkg/vm/audio/` → PASS（ベースライン緑）。

### 検出した問題

1. **【確定・中severity】不正/切り詰めMIDIファイルで `ParseMIDITempoMap` が範囲外アクセス panic**
   - 場所: [midi.go:614-704 `ParseMIDITempoMap`](../pkg/vm/audio/midi.go#L614-L704)
   - 原因: トラック走査ループが `pos < trackEnd` のみを条件とし、`trackEnd = offset+8+trackLen`（**ファイル内の宣言値**）を信頼。`trackLen` が実データ長を超えると `pos` が `len(data)` を越え、[L645 `readVarLen(data[pos:])`](../pkg/vm/audio/midi.go#L645) や [L653 `data[pos]`](../pkg/vm/audio/midi.go#L653) でスライス範囲外 panic。
   - 検証済み: MThd 正常 + MTrk 宣言長127だが実データ数バイトのみ → `slice bounds out of range [26:24]` で panic。
   - 影響: 壊れた/細工した .MID を読むとクラッシュ（DoS）。実経路は `Play()`→`meltysynth.NewMidiFile`（先に呼ばれる）が弾く可能性はあるが、`ParseMIDITempoMap` 自体が境界チェック皆無で危険。
   - 修正案: ループ条件と各インデックスアクセスを `pos < trackEnd && pos < len(data)` でガード。`trackEnd = min(offset+8+trackLen, len(data))` とするのも可。

### その他（軽微）

2. **fadeout 中の `mp.player` 直接参照**: [audio.go:279-280](../pkg/vm/audio/audio.go#L279-L280)。`as.midiPlayer.player`（MIDIPlayer の非公開フィールド）を `mp.mu` を取らずに参照。`as.mu` で AudioSystem 経由の操作は直列化されるため実害は出にくいが、ロック境界をまたぐ参照で行儀は悪い。

3. **MIDIメタ/SysEx 後の running status 未クリア**: [midi.go:660-663](../pkg/vm/audio/midi.go#L660-L663)。System メッセージ後も `lastStatus` を保持。MIDI仕様では system common はrunning statusを解除すべき。テンポ抽出専用なので実害は限定的。

### 確認できた“非バグ”（参考）
- Timer（[timer.go](../pkg/vm/audio/timer.go)）の start/stop と goroutine 終了同期（stopCh/doneCh、ロック外で待機）は正しい。TIMEイベントは別goroutineから push するが EventQueue は mutex 保護済み。
- WAV（[wav.go](../pkg/vm/audio/wav.go)）は Ebitengine の `wav.DecodeWithSampleRate` に解析を委譲しており手動境界バグなし。
- `AudioSystem` のロック順序は as→mp で一貫。

### 所感
audio は全体に堅牢。唯一の確定バグは MIDI テンポ解析の境界チェック欠如によるクラッシュ。境界ガード追加で解消。

---

## 7. Graphics (pkg/graphics) — BMP解析＋要所スポット完了（13k行のため全体精査は未了）

テスト: `go test ./pkg/graphics/` → PASS（ベースライン緑）。

### 検出した問題

1. **【確定・中severity】不正BMP（負の幅／巨大寸法）で `DecodeBMP` がクラッシュ**
   - 場所: [bmp.go:94-171](../pkg/graphics/bmp.go#L94-L171)。`width=int(infoHeader.Width)` を無検証で使用。
   - 負の幅: [decodeRGB の rowSize](../pkg/graphics/bmp.go#L160-L171) が負になり `make([]byte, rowSize)` が `makeslice: len out of range` で panic（検証済み: width=-4 で再現）。
   - 巨大な width×height: [image.NewRGBA L136](../pkg/graphics/bmp.go#L136) が width*height*4 バイト確保を試み OOM/panic。
   - MIDI と同じ「ファイル宣言値を無検証で信頼」パターン。壊れた/細工した .BMP でクラッシュ（DoS）。
   - 修正案: `width <= 0 || height == 0 || width > MAX || height > MAX` を早期に弾く。`DataOffset` の `skipBytes` も負方向は無視されるが過大値は `io.CopyN` で吸収される。

### 確認できた“非バグ”（参考）
- `SceneChange` の speed は [L79-83](../pkg/graphics/scene_change.go#L79-L83) で 1-100 にクランプ済み（speed=0 による進捗ゼロ無限ループは起きない）。
- 多くの `make([]T,...)` はポインタスライスで安全。手動ピクセル確保の危険箇所は bmp.go のみ。
- RLE8/RLE4 デコードの絶対モード/パディング処理は概ね正しい（描画は `x<width && y<height` でガード済み）。

### 未走査（13k行のため）
graphics_draw.go / transfer.go / primitives.go / headless.go / 各 *_sprite.go の描画・転送ロジック詳細。Ebitengine 依存部分が多く境界は概ねライブラリ側でガードされるが、headless.go のソフトレンダリングは要追加精査。次セッション候補。

### 所感
高リスクなファイル解析（BMP）で確定クラッシュ1件。描画系の大部分は Ebitengine 委譲で安全側。全体精査は規模的に次セッション継続。

---

## 8. Sprite (pkg/sprite) — コア sprite.go 完了

テスト: `go test ./pkg/sprite/` → PASS（ベースライン緑）。

### 検出した問題

1. **【確定・中severity】親子サイクルのガードが無く、再帰走査でスタックオーバーフロー**
   - 場所: [sprite.go:130-140 `AddChild`](../pkg/sprite/sprite.go#L130-L140)。サイクル/自己親子の検査なし。
   - 検証済み: `a.AddChild(a)`（自己親）も `b.AddChild(c); c.AddChild(b)`（相互）も許容され、`a.parent==a` / `b<->c` が成立。
   - 影響: その状態で [AbsolutePosition L191-199](../pkg/sprite/sprite.go#L191-L199) / [EffectiveAlpha L203-209](../pkg/sprite/sprite.go#L203-L209) / [IsEffectivelyVisible L213-221](../pkg/sprite/sprite.go#L213-L221) / [drawSprite L395-424](../pkg/sprite/sprite.go#L395-L424) が無限再帰 → Go の stack overflow は `recover` 不可の致命エラーでプロセス即死。スクリプトから親子関係を操作できるなら DoS。
   - 修正案: `AddChild` で「child が s の祖先（または s 自身）か」を辿って検査し、サイクルになる場合は拒否。

2. **【中severity / 要スレッドモデル確認】Sprite 個別メソッドが無ロックで共有ツリーを変更**
   - `SpriteManager` は `sm.mu`（RWMutex）で保護し `Draw` は RLock、`CreateSprite`/`DeleteSprite` は Lock を取る。しかし [`Sprite.SetPosition`](../pkg/sprite/sprite.go#L69) / [`AddChild`](../pkg/sprite/sprite.go#L130) / [`RemoveChild`](../pkg/sprite/sprite.go#L143) / [`BringToFront`](../pkg/sprite/sprite.go#L155) / [`SendToBack`](../pkg/sprite/sprite.go#L173) は `*Sprite` 上で**ロックなし**に `x/y/parent/children` を変更する。
   - `BringRootToFront`/`SendRootToBack`（[L504-557](../pkg/sprite/sprite.go#L504-L557)）は sm.mu を取るのに、`Sprite.BringToFront`/`SendToBack` は取らない——ロック方針が不整合。
   - VM（イベントハンドラ）が Ebitengine の Draw とは別 goroutine で sprite を変更する構成なら **データ競合**。スレッドモデルはタスク9で確認する。同一 goroutine（Update/Draw が同一ループ）なら実害なし。

### 所感
スライスベースの z 順・親子管理・再帰削除のロジックは妥当。確定はサイクルガード欠如。ロック不整合は app のスレッドモデル次第（タスク9で裏取り）。各 *_sprite.go（window/shape/cast/picture）は sprite.go の派生で、同じパターンを踏襲。

> **タスク9での裏取り結果**: アプリで実際に使われるのは **pkg/graphics/sprite.go** の SpriteManager（pkg/sprite はほぼミラー）。`GraphicsSystem` が `gs.mu` を持ち、`Draw` は RLock、`MoveCast` 等の変更は Lock を取るため、**実アプリでは sprite アクセスは gs.mu で粗粒度に直列化されており、ロック競合の実害は概ね回避**されている。ただし sprite パッケージ自体は自己防御しておらず、gs.mu を取らずに sprite を触る経路があれば競合する設計上の脆さは残る。**サイクルガード欠如（#1）は引き続き有効な確定バグ**。

---

## 9. App / CLI / window / title 周辺 — app.go/cli.go 完了

テスト: 全パッケージ `go build ./...` 通過（ベースライン緑）。

### 確認できた重要事項（スレッドモデル）

- **VM は専用 goroutine で実行される**: [app.go:298-300](../pkg/app/app.go#L298-L300) および [app.go:477-479](../pkg/app/app.go#L477-L479) で `go func(){ vmErrCh <- vmInstance.Run() }()`。一方 Ebitengine の Update/Draw はメイン goroutine。→ VM のイベントハンドラ（sprite/graphics 変更）と Draw（読み取り）は**別 goroutine で並行**。
- ただし [graphics_draw.go:19-28 `Draw`](../pkg/graphics/graphics_draw.go#L19-L28) は `gs.mu.RLock()`、[graphics_sprite.go:110-155 `MoveCast`](../pkg/graphics/graphics_sprite.go#L110-L155) 等の変更系は `gs.mu.Lock()` を取るため、GraphicsSystem 経由のアクセスは直列化される（タスク8 #2 の競合は実害が抑えられている）。

### 検出した問題（軽微）

1. **【軽微】`reorderArgs` の bool フラグ判定がハードコード列挙で脆い**: [cli.go:106-134](../pkg/cli/cli.go#L106-L134)。`-h/--help/--headless` のみを「値を取らないフラグ」として除外。将来 bool フラグを追加すると、直後の位置引数（パス等）を誤ってフラグ値として吸収しうる。現状の定義済みフラグでは破綻しない。

### 未走査（軽め）
window.go(657) / title.go(365) / script.go(128) の詳細ロジック（ゲームループ Update、タイトル選択 UI、スクリプトローダ）。app の結線と CLI は確認済み。

### 所感
結線・CLI は健全。最大の収穫は「VM が別 goroutine」というスレッドモデルの確定（sprite ロック議論の前提）。GraphicsSystem の粗粒度ロックで主要経路の競合は回避されている。
