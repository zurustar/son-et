# 要件定義書: スプライトシステム (Sprite System)

## はじめに

このドキュメントは、FILLYエミュレータのスプライトシステムの要件を定義します。スプライトシステムは、すべての描画要素を統一的に管理し、階層的なZ順序による描画を実現します。

## 背景

### 現在の問題点

1. **1次元Z順序の制限**: 現在の実装では `CalculateGlobalZOrder(windowZOrder, localZOrder)` で1次元に変換している
2. **タイプ別固定範囲**: キャスト(100-999)、テキスト(1000-)とタイプ別に固定範囲が割り当てられている
3. **描画順序の不自然さ**: テキストは常にキャストより前面になってしまう
4. **操作順序の無視**: 本来は操作順序（PutCast、TextWrite、MovePicの呼び出し順）でZ順序が決まるべき

### 解決方針

すべての描画要素をスプライトとして統一的に扱い、Z順序を `[0, 1, 0]` のような配列（パス）で表現することで、親子関係に基づいた階層的な描画順序を実現します。

## 用語集

- **Sprite（スプライト）**: 画像を保持し、位置・Z順序・可視性・透明度を持つ描画単位
- **SpriteManager**: スプライトの作成・管理・描画を行うマネージャー
- **WindowSprite**: ウィンドウを表すスプライト。子スプライトの親となる
- **PictureSprite**: ピクチャを表すスプライト。MovePicで作成される
- **CastSprite**: キャストを表すスプライト。透明色処理をサポート
- **TextSprite**: テキストを表すスプライト。差分抽出方式で作成
- **ShapeSprite**: 図形を表すスプライト
- **Hierarchical_Z_Order（階層的Z順序）**: 親子関係に基づいた多次元のZ順序。配列（パス）で表現される
- **Z_Path（Zパス）**: スプライトの階層的Z順序を表す整数配列。例: `[0, 1, 2]`
- **Local_Z_Order（ローカルZ順序）**: 同じ親を持つ兄弟スプライト間での相対的なZ順序
- **Parent_Sprite（親スプライト）**: 他のスプライトを子として持つスプライト
- **Child_Sprite（子スプライト）**: 親スプライトに属するスプライト
- **Sibling_Sprite（兄弟スプライト）**: 同じ親を持つスプライト
- **Root_Sprite（ルートスプライト）**: 親を持たないトップレベルのスプライト（デスクトップ直下）
- **Operation_Order（操作順序）**: PutCast、TextWrite、MovePicなどの呼び出し順序
- **Z_Order_Counter（Z順序カウンター）**: 操作順序を追跡するためのカウンター
- **差分抽出**: テキスト描画時にアンチエイリアスを除去するための手法
- **Dirty_Flag**: スプライトの変更を追跡するフラグ

## 要件

---

### 第1部: スプライトの基本定義

---

### 要件1: 汎用スプライト

**ユーザーストーリー:** 開発者として、すべての描画要素を統一的なスプライトとして扱いたい。そうすることで、コードの重複を減らし、保守性を向上させられる。

#### 受け入れ基準

1.1. THE Sprite SHALL 一意のIDを持つ
1.2. THE Sprite SHALL *ebiten.Image を保持する
1.3. THE Sprite SHALL 位置（X, Y座標）を持つ
1.4. THE Sprite SHALL Z順序を持つ
1.5. THE Sprite SHALL 可視性フラグを持つ
1.6. THE Sprite SHALL 透明度（0.0〜1.0）を持つ
1.7. THE Sprite SHALL 親スプライトへの参照を持つ（オプション）
1.8. THE Sprite SHALL ダーティフラグを持つ

### 要件2: 親子関係

**ユーザーストーリー:** 開発者として、スプライト間に親子関係を設定したい。そうすることで、ウインドウとその内容物の関係を自然に表現できる。

#### 受け入れ基準

2.1. WHEN 子スプライトの絶対位置を計算するとき THEN THE System SHALL 親の位置を加算する
2.2. WHEN 子スプライトの実効透明度を計算するとき THEN THE System SHALL 親の透明度を乗算する
2.3. WHEN 親スプライトが非表示のとき THEN THE System SHALL 子スプライトも非表示として扱う
2.4. THE Sprite SHALL 親を変更できる
2.5. THE Sprite SHALL 子スプライトのリストを保持する
2.6. THE Sprite SHALL ローカルZ順序（兄弟間での順序）を持つ
2.7. WHEN 子スプライトが追加されたとき THEN THE System SHALL 操作順序に基づいてローカルZ順序を割り当てる
2.8. THE System SHALL 親子関係の深さに制限を設けない（n次元の階層をサポート）

### 要件3: スプライトマネージャー

**ユーザーストーリー:** 開発者として、スプライトを一元管理したい。そうすることで、描画順序の管理やスプライトの検索が容易になる。

#### 受け入れ基準

3.1. THE SpriteManager SHALL スプライトを作成できる
3.2. THE SpriteManager SHALL 指定サイズの空のスプライトを作成できる
3.3. THE SpriteManager SHALL IDでスプライトを取得できる
3.4. THE SpriteManager SHALL スプライトを削除できる
3.5. THE SpriteManager SHALL すべてのスプライトをクリアできる
3.6. THE SpriteManager SHALL 登録されているスプライト数を返せる

---

### 第2部: スプライトタイプ

---

### 要件4: ウインドウスプライト

**ユーザーストーリー:** 開発者として、仮想ウインドウをスプライトとして表現したい。そうすることで、ウインドウとその内容物を親子関係で管理できる。

#### 受け入れ基準

4.1. THE System SHALL 指定サイズ・背景色のウインドウスプライトを作成できる
4.2. THE System SHALL ウインドウスプライトを親として子スプライトを追加できる
4.3. WHEN ウインドウが閉じられたとき THEN THE System SHALL ウインドウとその子スプライトを削除する

### 要件5: ピクチャスプライト

**ユーザーストーリー:** 開発者として、BMPファイルをスプライトとして表示したい。そうすることで、FILLYのピクチャ機能を実現できる。

#### 受け入れ基準

5.1. THE System SHALL BMPファイルからスプライトを作成できる
5.2. THE System SHALL 透明色を指定できる
5.3. THE System SHALL ピクチャの一部を切り出してスプライトにできる

### 要件6: キャストスプライト

**ユーザーストーリー:** 開発者として、キャスト（アニメーションスプライト）を表現したい。そうすることで、FILLYのキャスト機能を実現できる。

#### 受け入れ基準

6.1. THE System SHALL キャストをスプライトとして作成できる
6.2. THE System SHALL キャストの位置を移動できる（残像なし）
6.3. THE System SHALL キャストを削除できる
6.4. THE System SHALL 透明色処理をサポートする

### 要件7: テキストスプライト（差分抽出方式）

**ユーザーストーリー:** 開発者として、テキストをアンチエイリアスなしで描画したい。そうすることで、FILLYの見た目を正確に再現できる。

#### 受け入れ基準

7.1. THE System SHALL 背景色の上にテキストを描画し、差分を抽出する
7.2. THE System SHALL 差分抽出結果を透過スプライトとして生成する
7.3. THE System SHALL テキストの色を指定できる
7.4. THE System SHALL 複数のテキストを重ねて描画できる
7.5. WHEN 同じ位置に異なるテキストを描画したとき THEN THE System SHALL 前のテキストの影を残さない

### 要件8: 図形スプライト

**ユーザーストーリー:** 開発者として、図形（線、矩形、円など）をスプライトとして描画したい。そうすることで、FILLYの描画機能を実現できる。

#### 受け入れ基準

8.1. THE System SHALL 線を描画したスプライトを作成できる
8.2. THE System SHALL 矩形を描画したスプライトを作成できる
8.3. THE System SHALL 塗りつぶし矩形を描画したスプライトを作成できる

---

### 第3部: 階層的Z順序システム

---

### 要件9: 階層的Z順序の表現

**ユーザーストーリー:** 開発者として、スプライトのZ順序を階層的に表現したい。そうすることで、親子関係に基づいた直感的な描画順序を実現できる。

#### 受け入れ基準

9.1. THE Sprite SHALL Z_Pathを持つ（整数配列として表現）
9.2. THE Z_Path SHALL 親のZ_Pathに自身のLocal_Z_Orderを追加した形式である
9.3. THE Root_Sprite SHALL 単一要素のZ_Path（例: `[0]`）を持つ
9.4. WHEN 子スプライトが作成されたとき THEN THE System SHALL 親のZ_Pathを継承し、自身のLocal_Z_Orderを追加する
9.5. THE System SHALL Z_Pathの辞書順比較でスプライトの描画順序を決定する

### 要件10: 操作順序によるZ順序決定

**ユーザーストーリー:** 開発者として、スプライトのZ順序が操作順序で決まるようにしたい。そうすることで、タイプに関係なく後から作成されたスプライトが前面に表示される。

#### 受け入れ基準

10.1. THE System SHALL 各親スプライトごとにZ_Order_Counterを管理する
10.2. WHEN PutCastが呼び出されたとき THEN THE System SHALL 現在のZ_Order_Counterを使用してLocal_Z_Orderを割り当てる
10.3. WHEN TextWriteが呼び出されたとき THEN THE System SHALL 現在のZ_Order_Counterを使用してLocal_Z_Orderを割り当てる
10.4. WHEN MovePicが呼び出されたとき THEN THE System SHALL 現在のZ_Order_Counterを使用してLocal_Z_Orderを割り当てる
10.5. WHEN スプライトが作成されたとき THEN THE System SHALL Z_Order_Counterをインクリメントする
10.6. THE System SHALL タイプ（キャスト、テキスト、ピクチャ）に関係なく、操作順でZ順序を決定する

### 要件11: Z_Pathの比較

**ユーザーストーリー:** 開発者として、Z_Pathを効率的に比較したい。そうすることで、描画順序のソートが高速に行える。

#### 受け入れ基準

11.1. THE System SHALL Z_Pathを辞書順（lexicographic order）で比較する
11.2. WHEN Z_Path Aの先頭がZ_Path Bの先頭と一致するとき THEN THE System SHALL 次の要素を比較する
11.3. WHEN Z_Path Aが Z_Path Bのプレフィックスであるとき THEN THE System SHALL AをBより前（背面）と判定する
11.4. THE System SHALL 比較結果をキャッシュして再利用する

### 要件12: 階層的Z順序による描画

**ユーザーストーリー:** 開発者として、スプライトが階層的Z順序に従って正しく描画されるようにしたい。そうすることで、ウインドウ間の重なりとウインドウ内の要素の重なりを両方正しく制御できる。

#### 受け入れ基準

12.1. THE SpriteManager SHALL ルートスプライト（親を持たないスプライト）をローカルZ順序で描画する
12.2. WHEN スプライトを描画した後 THEN THE System SHALL そのスプライトの子をローカルZ順序で再帰的に描画する
12.3. THE SpriteManager SHALL 非表示のスプライトとその子孫を描画しない
12.4. THE SpriteManager SHALL 透明度を適用して描画する
12.5. THE SpriteManager SHALL 親子関係を考慮した絶対位置で描画する
12.6. WHEN 前面のウインドウが存在するとき THEN THE System SHALL 背面のウインドウの内容を前面のウインドウで隠す
12.7. THE System SHALL タイプ別の固定Z順序範囲を使用しない（操作順序でZ順序を決定する）

### 要件13: 親子関係に基づく描画順序

**ユーザーストーリー:** 開発者として、親スプライトの子が親の描画後に描画されるようにしたい。そうすることで、ウィンドウの内容がウィンドウの上に正しく表示される。

#### 受け入れ基準

13.1. WHEN 描画が行われるとき THEN THE System SHALL 親スプライトを先に描画し、その後に子スプライトを描画する
13.2. WHEN 同じ親を持つ子スプライトを描画するとき THEN THE System SHALL Local_Z_Order順で描画する
13.3. THE System SHALL 任意の深さの親子関係に対応する
13.4. WHEN 親スプライトが非表示のとき THEN THE System SHALL 子スプライトも描画しない

### 要件14: ウィンドウ間の描画順序

**ユーザーストーリー:** 開発者として、前面のウィンドウの内容が背面のウィンドウの内容を完全に覆い隠すようにしたい。そうすることで、ウィンドウの重なりが正しく表現される。

#### 受け入れ基準

14.1. THE System SHALL ウィンドウをRoot_Spriteとして扱う
14.2. WHEN ウィンドウAがウィンドウBより前面にあるとき THEN THE System SHALL ウィンドウAのすべての子スプライトをウィンドウBのすべての子スプライトより前面に描画する
14.3. THE System SHALL ウィンドウのZ順序変更時に、そのウィンドウの子スプライトのZ_Pathを更新する
14.4. WHEN ウィンドウが前面に移動したとき THEN THE System SHALL そのウィンドウのZ_Pathを更新する

### 要件15: 動的なZ順序変更

**ユーザーストーリー:** 開発者として、スプライトのZ順序を動的に変更したい。そうすることで、スプライトを前面や背面に移動できる。

#### 受け入れ基準

15.1. THE System SHALL スプライトのLocal_Z_Orderを変更できる
15.2. WHEN Local_Z_Orderが変更されたとき THEN THE System SHALL Z_Pathを再計算する
15.3. WHEN 親スプライトが変更されたとき THEN THE System SHALL 子スプライトのZ_Pathを再計算する
15.4. THE System SHALL スプライトを最前面に移動するメソッドを提供する
15.5. THE System SHALL スプライトを最背面に移動するメソッドを提供する

---

### 第4部: ピクチャとスプライトの統合

---

### 要件16: ピクチャのスプライト化

**ユーザーストーリー:** 開発者として、ピクチャをロードした時点でスプライトとして管理したい。そうすることで、ウインドウに関連付けられる前でもキャストやテキストの親として機能できる。

#### 背景

FILLYでは以下のパターンが存在する：
1. ピクチャをロード（LoadPic）してからウインドウに関連付ける（SetPic）
2. ウインドウに関連付けられていないピクチャにテキストを描画し、MovePicで転送する
3. キャストやテキストはピクチャ番号を指定して配置される

現在の実装では、ウインドウに関連付けられたときに初めてPictureSpriteが作成されるため、
関連付け前のピクチャに対するCastSetやTextWriteで親が見つからない問題がある。

#### 受け入れ基準

16.1. WHEN LoadPicが呼び出されたとき THEN THE System SHALL 非表示のPictureSpriteを作成する
16.2. THE PictureSprite SHALL ピクチャ番号をキーとして管理される
16.3. WHEN SetPicが呼び出されたとき THEN THE System SHALL 既存のPictureSpriteをウインドウの子として関連付ける
16.4. WHEN SetPicが呼び出されたとき THEN THE System SHALL PictureSpriteを表示状態にする
16.5. WHEN ウインドウに関連付けられていないピクチャにCastSetが呼び出されたとき THEN THE System SHALL キャストをそのピクチャのPictureSpriteの子として管理する
16.6. WHEN ウインドウに関連付けられていないピクチャにTextWriteが呼び出されたとき THEN THE System SHALL テキストをそのピクチャのPictureSpriteの子として管理する
16.7. THE System SHALL ピクチャがウインドウに関連付けられたとき、既存の子スプライト（キャスト、テキスト）のZ_Pathを更新する
16.8. WHEN ピクチャが解放されたとき THEN THE System SHALL 対応するPictureSpriteとその子スプライトを削除する

### 要件17: ピクチャスプライトの状態管理

**ユーザーストーリー:** 開発者として、ピクチャスプライトの状態（表示/非表示、ウインドウ関連付け）を追跡したい。そうすることで、適切なタイミングで描画できる。

#### 受け入れ基準

17.1. THE PictureSprite SHALL 「未関連付け」「関連付け済み」の状態を持つ
17.2. WHEN PictureSpriteが「未関連付け」状態のとき THEN THE System SHALL そのスプライトを描画しない
17.3. WHEN PictureSpriteが「関連付け済み」状態のとき THEN THE System SHALL 親ウインドウの可視性に従って描画する
17.4. THE System SHALL ピクチャ番号からPictureSpriteを効率的に検索できる
17.5. WHEN MovePicが呼び出されたとき THEN THE System SHALL 転送先ピクチャのPictureSpriteの画像を更新する
17.6. WHEN TextWriteが呼び出されたとき THEN THE System SHALL 対象ピクチャのPictureSpriteの画像を更新する

### 要件18: ピクチャ内のスプライト

**ユーザーストーリー:** 開発者として、ピクチャの中にキャストやテキストを配置したい。そうすることで、複雑な階層構造を表現できる。

#### 受け入れ基準

18.1. THE PictureSprite SHALL 子スプライトを持てる
18.2. WHEN ピクチャ内にキャストが配置されたとき THEN THE System SHALL キャストをピクチャの子スプライトとして管理する
18.3. WHEN ピクチャ内にテキストが配置されたとき THEN THE System SHALL テキストをピクチャの子スプライトとして管理する
18.4. THE System SHALL ピクチャの移動時に子スプライトも一緒に移動する

### 要件19: スプライトシステム完全統合

**ユーザーストーリー:** 開発者として、すべての描画をスプライトシステム経由で行いたい。そうすることで、描画の一貫性を保ち、Z順序の問題を解消できる。

#### 受け入れ基準

19.1. WHEN MovePicが呼び出されたとき THEN THE System SHALL ピクチャー画像への直接描画を行わず、PictureSpriteのみを作成する
19.2. WHEN Draw()が呼び出されたとき THEN THE System SHALL スプライトシステムのみを使用して描画する
19.3. THE System SHALL ピクチャー画像への焼き付けを廃止する
19.4. THE System SHALL すべての描画要素をスプライトとして管理する

### 要件20: ピクチャースプライトの融合

**ユーザーストーリー:** 開発者として、同じ領域に重なるピクチャースプライトを融合したい。そうすることで、スプライト数を削減し、パフォーマンスを向上させられる。

#### 受け入れ基準

20.1. WHEN 同じピクチャーIDに対してMovePicが複数回呼び出されたとき THEN THE System SHALL 既存のPictureSpriteに画像を合成できる
20.2. THE PictureSpriteManager SHALL 融合可能なスプライトを検索できる
20.3. THE PictureSprite SHALL 新しい画像を既存の画像に合成できる
20.4. WHEN 融合が行われたとき THEN THE System SHALL スプライトの領域を適切に拡張する
20.5. THE System SHALL 融合によりスプライト数を削減する

---

### 第5部: パフォーマンスと互換性

---

### 要件21: パフォーマンス

**ユーザーストーリー:** 開発者として、スプライトシステムが効率的に動作するようにしたい。そうすることで、60FPSでのスムーズな描画を維持できる。

#### 受け入れ基準

21.1. THE SpriteManager SHALL Z順序のソート結果をキャッシュする
21.2. THE SpriteManager SHALL スプライトの変更時にソートが必要であることをマークする
21.3. THE System SHALL 変更のあったサブツリーのみを再ソートする
21.4. THE System SHALL Z_Pathの比較を効率的に行う（O(depth)）
21.5. THE System SHALL スレッドセーフな操作を提供する

### 要件22: 既存システムとの互換性

**ユーザーストーリー:** 開発者として、既存のFILLYスクリプトが正しく動作し続けるようにしたい。そうすることで、既存のタイトルを壊さずに改善できる。

#### 受け入れ基準

22.1. THE System SHALL 既存のスプライトAPIを維持する
22.2. THE System SHALL 既存のウィンドウ管理APIを維持する
22.3. WHEN 既存のスクリプトが実行されたとき THEN THE System SHALL 同等の描画結果を生成する
22.4. THE System SHALL 1次元Z順序から階層的Z順序への移行パスを提供する

### 要件23: デバッグ支援

**ユーザーストーリー:** 開発者として、階層的Z順序の状態を確認したい。そうすることで、描画順序の問題をデバッグできる。

#### 受け入れ基準

23.1. THE System SHALL スプライトのZ_Pathを文字列として取得できる
23.2. THE System SHALL スプライト階層をツリー形式で出力できる
23.3. THE System SHALL 描画順序のリストを出力できる
23.4. WHEN デバッグモードが有効なとき THEN THE System SHALL Z_Pathをオーバーレイ表示できる
