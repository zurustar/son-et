# 要件定義書: 描画システム (Graphics System)

## はじめに

このドキュメントは、FILLYスクリプト言語の描画システムの要件を定義します。実行エンジン（VM）は既に実装済みで、描画関連の関数はダミー実装として登録されています。本スペックでは、Ebitengineを使用した実際の描画機能を実装します。

描画システムは以下の5つの主要機能で構成されます：
1. **ウィンドウシステム**: 仮想ウィンドウの管理
2. **ピクチャーシステム**: 画像データの管理と転送
3. **キャストシステム**: スプライト（キャスト）の管理
4. **テキストシステム**: 文字列の描画
5. **描画プリミティブ**: 基本図形の描画

## 用語集

- **Picture（ピクチャー）**: メモリ上の画像データ。BMPファイルから読み込むか、CreatePicで生成する
- **Window（ウィンドウ）**: 仮想デスクトップ上に表示される矩形領域。ピクチャーを表示する
- **Cast（キャスト）**: ウィンドウ上に配置されるスプライト。ピクチャーの一部を切り出して表示する
- **Virtual_Desktop**: 描画対象となる仮想的なデスクトップ領域
- **Transparent_Color**: 透明色として扱う色（通常は黒 0x000000）
- **Drawing_Command_Queue**: メインスレッドで実行される描画コマンドのキュー
- **Ebitengine**: Go言語用の2Dゲームエンジン。描画とオーディオを提供する

## 技術的制約

### Ebitengineのメインスレッド制約

Ebitengineの描画APIはメインスレッドでのみ呼び出し可能です。イベントハンドラ（mes()ブロック）から描画を行う場合、以下のアプローチを採用します：

1. **描画コマンドキュー**: イベントハンドラは描画コマンドをキューに追加する
2. **メインループでの実行**: Ebitengineのゲームループ（Update/Draw）でキューを処理する
3. **同期**: 必要に応じて描画完了を待機する機構を提供する

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  Event Handler  │ --> │  Command Queue  │ --> │  Main Loop      │
│  (mes block)    │     │  (thread-safe)  │     │  (Ebitengine)   │
└─────────────────┘     └─────────────────┘     └─────────────────┘
```

## 要件

### 要件1: ピクチャーシステム

**ユーザーストーリー:** 開発者として、画像ファイルを読み込んでメモリ上で操作したい。そうすることで、ウィンドウやキャストに表示する画像を準備できる。

#### 受け入れ基準

1.1. WHEN LoadPic(filename)が呼ばれたとき、THE System SHALL 指定されたBMPファイルを読み込み、ピクチャーIDを返す
1.2. WHEN LoadPicが成功したとき、THE System SHALL 0から始まる連番のピクチャーIDを割り当てる
1.3. WHEN LoadPicで指定されたファイルが存在しないとき、THE System SHALL エラーをログに記録し、-1を返す
1.4. WHEN CreatePic(pic_no, width, height)が呼ばれたとき、THE System SHALL 指定されたサイズの空のピクチャーを生成する
1.5. WHEN CreatePicが成功したとき、THE System SHALL 新しいピクチャーIDを返す
1.6. WHEN DelPic(pic_no)が呼ばれたとき、THE System SHALL 指定されたピクチャーを破棄し、メモリを解放する
1.7. WHEN PicWidth(pic_no)が呼ばれたとき、THE System SHALL 指定されたピクチャーの幅を返す
1.8. WHEN PicHeight(pic_no)が呼ばれたとき、THE System SHALL 指定されたピクチャーの高さを返す
1.9. WHEN 存在しないピクチャーIDが指定されたとき、THE System SHALL エラーをログに記録し、0を返す
1.10. THE System SHALL BMP形式（24ビット、8ビット）の画像ファイルをサポートする
1.11. THE System SHALL PNG形式の画像ファイルもサポートする（拡張機能）
1.12. WHEN ファイル名の大文字小文字が異なるとき、THE System SHALL 大文字小文字を区別せずにファイルを検索する（Windows互換性）

### 要件2: ピクチャー転送

**ユーザーストーリー:** 開発者として、ピクチャー間で画像データを転送したい。そうすることで、画像の合成やアニメーションを実現できる。

#### 受け入れ基準

2.1. WHEN MovePic(src_pic, src_x, src_y, width, height, dst_pic, dst_x, dst_y, mode)が呼ばれたとき、THE System SHALL 指定された領域を転送する
2.2. WHEN MovePicのmodeが0のとき、THE System SHALL 通常コピーを行う
2.3. WHEN MovePicのmodeが1のとき、THE System SHALL 透明色（黒 0x000000）を除いて転送する
2.4. WHEN MovePicのmodeが2のとき、THE System SHALL シーンチェンジモードで転送する
2.5. WHEN MoveSPic(src_pic, src_x, src_y, src_w, src_h, dst_pic, dst_x, dst_y, dst_w, dst_h)が呼ばれたとき、THE System SHALL 拡大縮小して転送する
2.6. WHEN MoveSPicに透明色が指定されたとき、THE System SHALL 透明色を除いて拡大縮小転送する
2.7. WHEN ReversePic(src_pic, src_x, src_y, width, height, dst_pic, dst_x, dst_y)が呼ばれたとき、THE System SHALL 左右反転して転送する
2.8. WHEN TransPic(src_pic, src_x, src_y, width, height, dst_pic, dst_x, dst_y, trans_color)が呼ばれたとき、THE System SHALL 指定された透明色を除いて転送する
2.9. WHEN 転送元または転送先のピクチャーが存在しないとき、THE System SHALL エラーをログに記録し、処理をスキップする
2.10. WHEN 転送領域がピクチャーの範囲外のとき、THE System SHALL クリッピングを行い、有効な領域のみ転送する

### 要件3: ウィンドウシステム

**ユーザーストーリー:** 開発者として、仮想デスクトップ上にウィンドウを開いて画像を表示したい。そうすることで、ユーザーに視覚的なコンテンツを提供できる。

#### 受け入れ基準

3.1. WHEN OpenWin(pic)が呼ばれたとき、THE System SHALL ピクチャー全体を表示するウィンドウを開き、ウィンドウIDを返す
3.2. WHEN OpenWin(pic, x, y, width, height, pic_x, pic_y, color)が呼ばれたとき、THE System SHALL 指定された位置とサイズでウィンドウを開く
3.3. WHEN OpenWinが成功したとき、THE System SHALL 0から始まる連番のウィンドウIDを割り当てる
3.4. WHEN MoveWin(win, pic)が呼ばれたとき、THE System SHALL ウィンドウに関連付けられたピクチャーを変更する
3.5. WHEN MoveWin(win, pic, x, y, width, height, pic_x, pic_y)が呼ばれたとき、THE System SHALL ウィンドウの位置、サイズ、ピクチャー参照位置を変更する
3.6. WHEN CloseWin(win_no)が呼ばれたとき、THE System SHALL 指定されたウィンドウを閉じる
3.7. WHEN CloseWinAllが呼ばれたとき、THE System SHALL すべてのウィンドウを閉じる
3.8. WHEN CapTitle(win_no, title)が呼ばれたとき、THE System SHALL ウィンドウのキャプションを設定する
3.9. WHEN GetPicNo(win_no)が呼ばれたとき、THE System SHALL ウィンドウに関連付けられたピクチャー番号を返す
3.10. WHEN 存在しないウィンドウIDが指定されたとき、THE System SHALL エラーをログに記録し、処理をスキップする
3.11. THE System SHALL ウィンドウをZ順序で管理し、後から開いたウィンドウを前面に表示する
3.12. THE System SHALL ウィンドウの背景色（color引数）を適用する

### 要件4: キャストシステム

**ユーザーストーリー:** 開発者として、ウィンドウ上にスプライト（キャスト）を配置して動かしたい。そうすることで、アニメーションやインタラクティブな要素を実現できる。

#### 受け入れ基準

4.1. WHEN PutCast(win_no, pic_no, x, y, src_x, src_y, width, height)が呼ばれたとき、THE System SHALL 指定されたウィンドウにキャストを配置し、キャストIDを返す
4.2. WHEN PutCastが成功したとき、THE System SHALL 0から始まる連番のキャストIDを割り当てる
4.3. WHEN MoveCast(cast_no, x, y)が呼ばれたとき、THE System SHALL キャストの位置を変更する
4.4. WHEN MoveCast(cast_no, x, y, src_x, src_y, width, height)が呼ばれたとき、THE System SHALL キャストの位置とソース領域を変更する
4.5. WHEN MoveCast(cast_no, pic_no, x, y)が呼ばれたとき、THE System SHALL キャストのピクチャーと位置を変更する
4.6. WHEN DelCast(cast_no)が呼ばれたとき、THE System SHALL 指定されたキャストを削除する
4.7. WHEN 存在しないキャストIDが指定されたとき、THE System SHALL エラーをログに記録し、処理をスキップする
4.8. THE System SHALL キャストを透明色（黒 0x000000）を除いて描画する
4.9. THE System SHALL キャストをZ順序で管理し、後から配置したキャストを前面に表示する
4.10. THE System SHALL キャストの位置をウィンドウ相対座標で管理する

### 要件5: テキストシステム

**ユーザーストーリー:** 開発者として、ピクチャー上に文字列を描画したい。そうすることで、テキスト情報を表示できる。

#### 受け入れ基準

5.1. WHEN SetFont(font_name, size, charset, weight, italic, underline, strikeout)が呼ばれたとき、THE System SHALL フォント設定を変更する
5.2. WHEN TextWrite(pic_no, x, y, text)が呼ばれたとき、THE System SHALL 指定されたピクチャーに文字列を描画する
5.3. WHEN TextColor(color)が呼ばれたとき、THE System SHALL 文字色を設定する
5.4. WHEN BgColor(color)が呼ばれたとき、THE System SHALL 背景色を設定する
5.5. WHEN BackMode(mode)が呼ばれたとき、THE System SHALL 背景モードを設定する（0=透明, 1=不透明）
5.6. THE System SHALL 日本語（Shift-JIS由来のUTF-8）を正しく描画する
5.7. THE System SHALL デフォルトフォントとして日本語対応フォントを使用する
5.8. WHEN 指定されたフォントが見つからないとき、THE System SHALL デフォルトフォントを使用する
5.9. THE System SHALL フォントサイズをピクセル単位で指定できる
5.10. THE System SHALL 太字（weight >= 700）とイタリックをサポートする

### 要件6: 描画プリミティブ

**ユーザーストーリー:** 開発者として、基本的な図形を描画したい。そうすることで、UIやグラフィカルな要素を作成できる。

#### 受け入れ基準

6.1. WHEN DrawLine(pic_no, x1, y1, x2, y2)が呼ばれたとき、THE System SHALL 指定されたピクチャーに直線を描画する
6.2. WHEN DrawCircle(pic_no, x, y, radius, fill_mode)が呼ばれたとき、THE System SHALL 円を描画する
6.3. WHEN DrawCircleのfill_modeが0のとき、THE System SHALL 輪郭のみ描画する
6.4. WHEN DrawCircleのfill_modeが2のとき、THE System SHALL 塗りつぶして描画する
6.5. WHEN DrawRect(pic_no, x1, y1, x2, y2, fill_mode)が呼ばれたとき、THE System SHALL 矩形を描画する
6.6. WHEN FillRect(pic_no, x1, y1, x2, y2, color)が呼ばれたとき、THE System SHALL 指定された色で矩形を塗りつぶす
6.7. WHEN SetLineSize(size)が呼ばれたとき、THE System SHALL 線の太さを設定する
6.8. WHEN SetPaintColor(color)が呼ばれたとき、THE System SHALL 描画色を設定する
6.9. WHEN GetColor(pic_no, x, y)が呼ばれたとき、THE System SHALL 指定された座標のピクセル色を返す
6.10. THE System SHALL 色を24ビットRGB（0xRRGGBB）形式で扱う
6.11. WHEN 描画対象のピクチャーが存在しないとき、THE System SHALL エラーをログに記録し、処理をスキップする

### 要件7: 描画コマンドキュー

**ユーザーストーリー:** 開発者として、イベントハンドラから描画操作を行いたい。そうすることで、MIDI同期やタイマーイベントに応じた描画を実現できる。

#### 受け入れ基準

7.1. THE System SHALL スレッドセーフな描画コマンドキューを提供する
7.2. WHEN イベントハンドラから描画関数が呼ばれたとき、THE System SHALL 描画コマンドをキューに追加する
7.3. WHEN Ebitengineのゲームループが実行されたとき、THE System SHALL キュー内のコマンドを順次実行する
7.4. THE System SHALL 描画コマンドの実行順序を保証する（FIFO）
7.5. WHEN キューが空のとき、THE System SHALL 待機せずに次のフレームに進む
7.6. THE System SHALL 1フレームあたりの描画コマンド数に制限を設けない
7.7. WHEN 描画コマンドの実行中にエラーが発生したとき、THE System SHALL エラーをログに記録し、次のコマンドを実行する

### 要件8: 仮想デスクトップ

**ユーザーストーリー:** 開発者として、固定サイズの仮想デスクトップ上で描画を行いたい。そうすることで、オリジナルのFILLYスクリプトと同じ見た目を再現できる。

#### 受け入れ基準

8.1. THE System SHALL デフォルトで640x480ピクセルの仮想デスクトップを提供する
8.2. THE System SHALL WinInfo(0)で仮想デスクトップの幅を返す
8.3. THE System SHALL WinInfo(1)で仮想デスクトップの高さを返す
8.4. THE System SHALL 仮想デスクトップを実際のウィンドウサイズに合わせてスケーリングする
8.5. THE System SHALL アスペクト比を維持してスケーリングする
8.6. THE System SHALL スケーリング時にレターボックス（黒帯）を表示する
8.7. WHEN マウスイベントが発生したとき、THE System SHALL 仮想デスクトップ座標に変換してMesP2、MesP3に設定する

### 要件9: リソース管理

**ユーザーストーリー:** 開発者として、描画リソースが適切に管理されることを期待する。そうすることで、メモリリークを防ぎ、安定した動作を実現できる。

#### 受け入れ基準

9.1. WHEN ピクチャーが削除されたとき、THE System SHALL 関連するEbitengine画像リソースを解放する
9.2. WHEN ウィンドウが閉じられたとき、THE System SHALL 関連するキャストを削除する
9.3. WHEN プログラムが終了したとき、THE System SHALL すべての描画リソースを解放する
9.4. THE System SHALL ピクチャーIDの再利用を許可する（削除後に同じIDを再割り当て可能）
9.5. THE System SHALL 同時に管理できるピクチャー数を最大256に制限する
9.6. THE System SHALL 同時に管理できるウィンドウ数を最大64に制限する
9.7. THE System SHALL 同時に管理できるキャスト数を最大1024に制限する
9.8. WHEN リソース制限に達したとき、THE System SHALL エラーをログに記録し、-1を返す

### 要件10: VM統合

**ユーザーストーリー:** 開発者として、既存のVMから描画関数を呼び出したい。そうすることで、FILLYスクリプトから描画機能を利用できる。

#### 受け入れ基準

10.1. THE System SHALL pkg/vm/vm.goの組み込み関数として描画関数を登録する
10.2. THE System SHALL 既存のダミー実装を実際の実装に置き換える
10.3. THE System SHALL 描画システムをVMのオプションとして初期化可能にする
10.4. WHEN ヘッドレスモードが有効のとき、THE System SHALL 描画操作をログに記録するのみで実際の描画を行わない
10.5. THE System SHALL 描画システムの初期化失敗時にエラーを報告する
10.6. THE System SHALL 描画システムのシャットダウン時にすべてのリソースを解放する

### 要件11: エラーハンドリング

**ユーザーストーリー:** 開発者として、描画エラーが発生しても実行が継続されることを期待する。そうすることで、一部の画像が見つからなくてもプログラムが動作する。

#### 受け入れ基準

11.1. WHEN 画像ファイルが見つからないとき、THE System SHALL エラーをログに記録し、実行を継続する
11.2. WHEN 画像ファイルが破損しているとき、THE System SHALL エラーをログに記録し、実行を継続する
11.3. WHEN 無効なピクチャーIDが指定されたとき、THE System SHALL エラーをログに記録し、デフォルト値を返す
11.4. WHEN 無効なウィンドウIDが指定されたとき、THE System SHALL エラーをログに記録し、処理をスキップする
11.5. WHEN 無効なキャストIDが指定されたとき、THE System SHALL エラーをログに記録し、処理をスキップする
11.6. THE System SHALL エラーメッセージに関数名と引数を含める
11.7. THE System SHALL 致命的でないエラーの後も実行を継続する

### 要件12: パフォーマンス

**ユーザーストーリー:** 開発者として、描画処理が高速に実行されることを期待する。そうすることで、スムーズなアニメーションを実現できる。

#### 受け入れ基準

12.1. THE System SHALL 60 FPSでの描画を目標とする
12.2. THE System SHALL ピクチャー転送をGPUアクセラレーションで実行する（Ebitengine経由）
12.3. THE System SHALL 変更のないウィンドウの再描画を最小化する
12.4. THE System SHALL キャストの描画をバッチ処理で最適化する
12.5. WHEN 描画負荷が高いとき、THE System SHALL フレームスキップを行わずに処理を継続する

### 要件13: シーンチェンジ（MovePicのmode引数）

**ユーザーストーリー:** 開発者として、画像転送時にフェードやワイプなどのエフェクトを適用したい。そうすることで、視覚的に魅力的な画面遷移を実現できる。

#### 受け入れ基準

13.1. WHEN MovePicのmodeが1のとき、THE System SHALL 透明色（黒 0x000000）を除いて転送する
13.2. WHEN MovePicのmodeが2のとき、THE System SHALL 上から下へのワイプで転送する
13.3. WHEN MovePicのmodeが3のとき、THE System SHALL 左から右へのワイプで転送する
13.4. WHEN MovePicのmodeが4のとき、THE System SHALL 右から左へのワイプで転送する
13.5. WHEN MovePicのmodeが5のとき、THE System SHALL 下から上へのワイプで転送する
13.6. WHEN MovePicのmodeが6のとき、THE System SHALL 中央から外側へのワイプで転送する
13.7. WHEN MovePicのmodeが7のとき、THE System SHALL 外側から中央へのワイプで転送する
13.8. WHEN MovePicのmodeが8のとき、THE System SHALL ランダムなブロックで転送する
13.9. WHEN MovePicのmodeが9のとき、THE System SHALL フェードイン/アウトで転送する
13.10. WHEN MovePicにspeed引数が指定されたとき、THE System SHALL エフェクトの速度を調整する
13.11. THE System SHALL シーンチェンジを非同期で実行し、完了を待たずに次の処理に進む

### 要件14: ゲームループ統合

**ユーザーストーリー:** 開発者として、VMのイベントループとEbitengineのゲームループを統合したい。そうすることで、描画とスクリプト実行が協調して動作する。

#### 受け入れ基準

14.1. THE System SHALL EbitengineのUpdate()内でVMのイベント処理を呼び出す
14.2. THE System SHALL EbitengineのDraw()内で描画コマンドキューを処理する
14.3. THE System SHALL EbitengineのUpdate()内でオーディオシステムの更新を呼び出す
14.4. WHEN VMが終了したとき、THE System SHALL Ebitengineのゲームループを終了する
14.5. WHEN Ebitengineのウィンドウが閉じられたとき、THE System SHALL VMを停止する
14.6. THE System SHALL マウスイベントをEbitengineから取得し、VMのイベントキューに追加する
14.7. THE System SHALL キーボードイベントをEbitengineから取得し、VMのイベントキューに追加する（将来拡張）

### 要件15: デバッグオーバーレイ

**ユーザーストーリー:** 開発者として、デバッグ時にピクチャー、ウィンドウ、キャストのIDを画面上で確認したい。そうすることで、描画の問題を素早く特定できる。

#### 受け入れ基準

15.1. WHEN ログレベルがDebug（レベル2）以上のとき、THE System SHALL ウィンドウIDをタイトルバーに表示する
15.2. WHEN ログレベルがDebug以上のとき、THE System SHALL ピクチャーIDをウィンドウ内容の左上に表示する
15.3. WHEN ログレベルがDebug以上のとき、THE System SHALL キャストIDとソースピクチャーIDをキャスト位置に表示する
15.4. THE System SHALL デバッグラベルを半透明の背景付きで表示し、視認性を確保する
15.5. THE System SHALL ウィンドウIDを黄色、ピクチャーIDを緑色、キャストIDを黄色で表示する
15.6. THE System SHALL デバッグラベルの表示形式を以下とする：ウィンドウ `[W1]`、ピクチャー `P1`、キャスト `C1(P2)`
15.7. WHEN ログレベルがDebug未満のとき、THE System SHALL デバッグオーバーレイを表示しない
15.8. THE System SHALL デバッグオーバーレイの表示/非表示をログレベル設定で切り替え可能にする
