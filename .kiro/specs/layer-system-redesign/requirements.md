# 要件定義書: レイヤーシステム再設計 (Layer System Redesign)

## はじめに

このドキュメントは、FILLYエミュレータのグラフィックスレイヤーシステムの再設計要件を定義します。現在の実装では、レイヤーがPictureIDで管理されていますが、正しくはWindowIDで管理されるべきです。この再設計により、レイヤーの所属関係を正しく管理し、描画の不具合を解消します。

## 用語集

- **Window（ウィンドウ）**: 仮想デスクトップ上に表示される矩形領域。背景色と透明なレイヤースタックを持つ
- **Picture（ピクチャー）**: メモリ上の画像データ。ウィンドウに関連付けられて表示される
- **Layer（レイヤー）**: ウィンドウに属する描画要素。位置、サイズ、Z順序、可視性を持つ
- **Picture_Layer**: MovePicで作成されるレイヤー。「焼き付け」（baking）が可能
- **Text_Layer**: TextWriteで作成されるレイヤー。アンチエイリアス対策のため常に新規作成
- **Cast_Layer**: PutCastで作成されるスプライトレイヤー。移動時に残像を残さない
- **Baking（焼き付け）**: 画像をレイヤーに合成する処理。最上位がPicture_Layerの場合はそこに焼き付け、そうでなければ新規レイヤーを作成
- **Window_Layer_Set**: ウィンドウに属するレイヤーの集合
- **Z_Order**: レイヤーの描画順序。大きいほど前面に表示される
- **Background_Color**: ウィンドウの背景色。レイヤースタックの最背面に描画される

## 現在の問題点

### バグ: PictureIDとWindowIDの不一致

現在の実装では以下の不整合があります：

1. `createCastLayer`は`cast.WinID`でレイヤーを登録している（正しい）
2. `drawLayersForWindow`は`win.PicID`でレイヤーを検索している（間違い）

この不一致により：
- レイヤーが見つからない
- レイヤーが間違ったウィンドウに関連付けられる
- 描画が正しく行われない

### 設計上の問題

現在の`PictureLayerSet`はピクチャーIDをキーとして管理していますが、FILLYの設計では：
- レイヤーはウィンドウに属する
- ウィンドウは背景色を持ち、その上に透明なレイヤーが重なる
- ピクチャーはレイヤーの描画ソースとして使用される

## 要件

### 要件1: レイヤー所属モデルの変更

**ユーザーストーリー:** 開発者として、レイヤーがウィンドウに正しく所属するようにしたい。そうすることで、描画の不具合を解消し、FILLYの動作を正確に再現できる。

#### 受け入れ基準

1.1. THE Layer_Manager SHALL レイヤーをWindowIDで管理する（PictureIDではなく）
1.2. WHEN ウィンドウが開かれたとき THEN THE Layer_Manager SHALL そのウィンドウ用のWindow_Layer_Setを作成する
1.3. WHEN ウィンドウが閉じられたとき THEN THE Layer_Manager SHALL そのウィンドウに属するすべてのレイヤーを削除する
1.4. THE Window_Layer_Set SHALL 背景色とレイヤースタックを保持する
1.5. THE Layer_Manager SHALL WindowIDをキーとしてWindow_Layer_Setを検索する

### 要件2: レイヤータイプの定義

**ユーザーストーリー:** 開発者として、3種類のレイヤー（Picture、Text、Cast）を明確に区別したい。そうすることで、各レイヤータイプに適した処理を行える。

#### 受け入れ基準

2.1. THE System SHALL Picture_Layerを定義する（MovePicで作成、焼き付け可能）
2.2. THE System SHALL Text_Layerを定義する（TextWriteで作成、常に新規レイヤー）
2.3. THE System SHALL Cast_Layerを定義する（PutCastで作成、スプライト動作）
2.4. WHEN レイヤーが作成されたとき THEN THE System SHALL レイヤータイプを識別可能にする
2.5. THE Picture_Layer SHALL 焼き付け対象として機能する
2.6. THE Text_Layer SHALL アンチエイリアス対策のため常に新規作成される
2.7. THE Cast_Layer SHALL 移動時に残像を残さない（スプライト動作）

### 要件3: MovePicの焼き付けロジック

**ユーザーストーリー:** 開発者として、MovePicが正しい焼き付けロジックで動作するようにしたい。そうすることで、FILLYの描画動作を正確に再現できる。

#### 受け入れ基準

3.1. WHEN MovePicが呼び出されたとき THEN THE System SHALL 最上位レイヤーのタイプを確認する
3.2. IF 最上位レイヤーがPicture_Layerである THEN THE System SHALL そのレイヤーに画像を焼き付ける
3.3. IF 最上位レイヤーがCast_LayerまたはText_Layerである THEN THE System SHALL 新しいPicture_Layer（ウィンドウサイズの透明レイヤー）を作成し、そこに焼き付ける
3.4. IF レイヤースタックが空である THEN THE System SHALL 新しいPicture_Layerを作成し、そこに焼き付ける
3.5. THE Picture_Layer SHALL ウィンドウサイズの透明画像として初期化される
3.6. WHEN 焼き付けが行われたとき THEN THE System SHALL 焼き付け先レイヤーをダーティとしてマークする

### 要件4: Cast_Layerの動作

**ユーザーストーリー:** 開発者として、Cast_Layerがスプライトとして正しく動作するようにしたい。そうすることで、キャストの移動時に残像が残らない。

#### 受け入れ基準

4.1. WHEN PutCastが呼び出されたとき THEN THE System SHALL 新しいCast_Layerを作成する
4.2. WHEN MoveCastが呼び出されたとき THEN THE System SHALL Cast_Layerの位置を更新する（残像なし）
4.3. WHEN DelCastが呼び出されたとき THEN THE System SHALL Cast_Layerを削除する
4.4. THE Cast_Layer SHALL ソースピクチャーへの参照を保持する（画像データのコピーではなく）
4.5. THE Cast_Layer SHALL 透明色処理をサポートする
4.6. WHEN Cast_Layerが移動したとき THEN THE System SHALL 古い位置と新しい位置をダーティ領域としてマークする

### 要件5: Text_Layerの動作

**ユーザーストーリー:** 開発者として、Text_Layerが常に新規作成されるようにしたい。そうすることで、テキストのアンチエイリアスアーティファクトを防げる。

#### 受け入れ基準

5.1. WHEN TextWriteが呼び出されたとき THEN THE System SHALL 常に新しいText_Layerを作成する
5.2. THE Text_Layer SHALL 既存のレイヤーを再利用しない
5.3. THE Text_Layer SHALL テキストの境界ボックスサイズで作成される
5.4. WHEN 同じ位置に異なるテキストが描画されたとき THEN THE System SHALL 前のテキストの影を残さない
5.5. THE Text_Layer SHALL 透明な背景を持つ

### 要件6: 描画順序（Z順序）

**ユーザーストーリー:** 開発者として、レイヤーが正しいZ順序で描画されるようにしたい。そうすることで、FILLYの描画結果を正確に再現できる。

#### 受け入れ基準

6.1. THE System SHALL ウィンドウの背景色を最初に描画する
6.2. THE System SHALL レイヤースタックをZ順序（小さい順）で描画する
6.3. WHEN 新しいレイヤーが作成されたとき THEN THE System SHALL 現在のZ順序カウンターを割り当て、カウンターを増加させる
6.4. THE System SHALL すべてのレイヤータイプ（Picture、Text、Cast）で共通のZ順序カウンターを使用する
6.5. WHEN 描画が行われたとき THEN THE System SHALL 背景色 → レイヤースタック（Z順序順）の順で合成する

### 要件7: ウィンドウとレイヤーの関連付け

**ユーザーストーリー:** 開発者として、レイヤーがウィンドウに正しく関連付けられるようにしたい。そうすることで、ウィンドウごとに独立したレイヤー管理ができる。

#### 受け入れ基準

7.1. WHEN PutCastが呼び出されたとき THEN THE System SHALL Cast_LayerをウィンドウIDで登録する
7.2. WHEN MovePicが呼び出されたとき THEN THE System SHALL Picture_Layerを転送先ピクチャーに関連付けられたウィンドウに登録する
7.3. WHEN TextWriteが呼び出されたとき THEN THE System SHALL Text_Layerを対象ピクチャーに関連付けられたウィンドウに登録する
7.4. WHEN drawLayersForWindowが呼び出されたとき THEN THE System SHALL ウィンドウIDでレイヤーを検索する
7.5. THE System SHALL ピクチャーIDからウィンドウIDへの逆引きをサポートする

### 要件8: 既存コードとの互換性

**ユーザーストーリー:** 開発者として、既存のFILLYスクリプトが正しく動作し続けるようにしたい。そうすることで、既存のタイトルを壊さずに改善できる。

#### 受け入れ基準

8.1. THE System SHALL 既存のCastManager APIを維持する
8.2. THE System SHALL 既存のTextRenderer APIを維持する
8.3. THE System SHALL 既存のMovePic APIを維持する
8.4. WHEN 既存のスクリプトが実行されたとき THEN THE System SHALL 同等の描画結果を生成する
8.5. THE System SHALL 既存のデバッグオーバーレイ機能を維持する

### 要件9: パフォーマンス最適化

**ユーザーストーリー:** 開発者として、レイヤーシステムが効率的に動作するようにしたい。そうすることで、60FPSでのスムーズな描画を維持できる。

#### 受け入れ基準

9.1. THE System SHALL ダーティフラグによる部分更新をサポートする
9.2. THE System SHALL 変更のないレイヤーのキャッシュを使用する
9.3. THE System SHALL 可視領域外のレイヤーの描画をスキップする
9.4. THE System SHALL 完全に覆われたレイヤーの描画をスキップする
9.5. WHEN レイヤーが変更されたとき THEN THE System SHALL ダーティ領域のみを再合成する

### 要件10: エラーハンドリング

**ユーザーストーリー:** 開発者として、レイヤー操作のエラーが適切に処理されるようにしたい。そうすることで、問題発生時にも実行が継続される。

#### 受け入れ基準

10.1. WHEN 存在しないウィンドウIDが指定されたとき THEN THE System SHALL エラーをログに記録し、処理をスキップする
10.2. WHEN 存在しないレイヤーIDが指定されたとき THEN THE System SHALL エラーをログに記録し、処理をスキップする
10.3. WHEN レイヤー作成に失敗したとき THEN THE System SHALL エラーをログに記録し、nilを返す
10.4. THE System SHALL 致命的でないエラーの後も実行を継続する
10.5. THE System SHALL エラーメッセージに関数名、ウィンドウID、レイヤーIDを含める
