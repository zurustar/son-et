# 要件ドキュメント

## はじめに

レイヤーベース描画システムは、FILLYの描画処理を改善するための機能です。現在の実装では、キャストがスクリーンに直接描画されており、MovePicで描画された内容がキャストの下に隠れてしまう問題があります。_old_implementation2では、すべての描画（MovePic、キャスト、テキスト）がbase_picに「焼き付けられる」方式でした。

本システムでは、すべての描画要素をレイヤーとして管理し、正しいZ順序で合成する描画システムを実装します。

## 用語集

- **Layer**: 描画要素を保持する単位。位置、サイズ、Z順序、可視性を持つ
- **Layer_Manager**: レイヤーの作成、削除、合成を管理するコンポーネント
- **Background_Layer**: 背景画像（OriginalImage）を保持するレイヤー
- **Cast_Layer**: キャストを保持するレイヤー。透明色処理を含む
- **Drawing_Layer**: MovePicで描画された内容を保持するレイヤー
- **Text_Layer**: テキスト描画を保持するレイヤー
- **Dirty_Flag**: レイヤーの変更を追跡するフラグ
- **Composite_Buffer**: レイヤーを合成した結果を保持するバッファ
- **Z_Order**: レイヤーの描画順序。大きいほど前面に表示される
- **Visible_Region**: ウィンドウ内で実際に表示される領域

## 要件

### 要件 1: レイヤー構造

**ユーザーストーリー:** 開発者として、描画要素をレイヤーとして管理したい。これにより、正しいZ順序で描画要素を合成できる。

#### 受け入れ基準

1. THE Layer_Manager SHALL 背景レイヤー（Background_Layer）を管理する
2. THE Layer_Manager SHALL キャストレイヤー（Cast_Layer）をZ順序で管理する
3. THE Layer_Manager SHALL 描画エントリ（Drawing_Entry）を管理する
4. THE Layer_Manager SHALL テキストレイヤー（Text_Layer）を管理する
5. WHEN レイヤーが追加されたとき THEN THE Layer_Manager SHALL **操作順序に基づいて**Z順序を割り当てる
6. THE Layer_Manager SHALL レイヤーを**操作順序（Z順序）**で合成する
7. **重要**: 後から実行された操作（MovePic、PutCast、TextWrite）は、先に実行された操作の上に表示される

### 要件 2: レイヤー管理

**ユーザーストーリー:** 開発者として、レイヤーを動的に追加・削除・更新したい。これにより、ゲームの状態変化に応じて描画内容を変更できる。

#### 受け入れ基準

1. WHEN PutCastが呼び出されたとき THEN THE Layer_Manager SHALL 新しいCast_Layerを作成し、**現在のZ順序カウンター**を割り当てる
2. WHEN MoveCastが呼び出されたとき THEN THE Layer_Manager SHALL 対応するCast_Layerの位置を更新する（Z順序は変更しない）
3. WHEN DelCastが呼び出されたとき THEN THE Layer_Manager SHALL 対応するCast_Layerを削除する
4. WHEN MovePicが呼び出されたとき THEN THE Layer_Manager SHALL 新しいDrawing_Entryを作成し、**現在のZ順序カウンター**を割り当てる
5. WHEN TextWriteが呼び出されたとき THEN THE Layer_Manager SHALL 新しいText_Layerを作成し、**現在のZ順序カウンター**を割り当てる
6. WHEN ウィンドウが閉じられたとき THEN THE Layer_Manager SHALL そのウィンドウに属するすべてのレイヤーを削除する
7. **重要**: すべての描画操作（MovePic、PutCast、TextWrite）は同じZ順序カウンターを共有し、操作のたびにカウンターが増加する

### 要件 3: ダーティフラグによる最適化

**ユーザーストーリー:** 開発者として、変更があったレイヤーのみを再描画したい。これにより、描画パフォーマンスを向上できる。

#### 受け入れ基準

1. WHEN レイヤーの位置が変更されたとき THEN THE Layer_Manager SHALL そのレイヤーのDirty_Flagを設定する
2. WHEN レイヤーの内容が変更されたとき THEN THE Layer_Manager SHALL そのレイヤーのDirty_Flagを設定する
3. WHEN レイヤーの可視性が変更されたとき THEN THE Layer_Manager SHALL そのレイヤーのDirty_Flagを設定する
4. WHEN 合成処理が完了したとき THEN THE Layer_Manager SHALL すべてのDirty_Flagをクリアする
5. WHEN Dirty_Flagが設定されていないレイヤーがあるとき THEN THE Layer_Manager SHALL そのレイヤーのキャッシュを使用する

### 要件 4: 可視領域クリッピング

**ユーザーストーリー:** 開発者として、ウィンドウ外のレイヤーを描画処理からスキップしたい。これにより、不要な描画処理を削減できる。

#### 受け入れ基準

1. WHEN レイヤーがウィンドウの可視領域外にあるとき THEN THE Layer_Manager SHALL そのレイヤーの描画をスキップする
2. WHEN レイヤーが部分的に可視領域内にあるとき THEN THE Layer_Manager SHALL 可視部分のみを描画する
3. THE Layer_Manager SHALL 各レイヤーの境界ボックスを計算する
4. THE Layer_Manager SHALL 可視領域との交差判定を行う

### 要件 5: レイヤーキャッシュ

**ユーザーストーリー:** 開発者として、変更がないレイヤーのキャッシュを使用したい。これにより、再描画のコストを削減できる。

#### 受け入れ基準

1. THE Layer_Manager SHALL 各レイヤーの描画結果をキャッシュする
2. WHEN レイヤーの内容が変更されていないとき THEN THE Layer_Manager SHALL キャッシュされた画像を使用する
3. WHEN レイヤーの内容が変更されたとき THEN THE Layer_Manager SHALL キャッシュを無効化する
4. THE Layer_Manager SHALL テキストレイヤーのキャッシュを特に重視する（作成コストが高いため）

### 要件 6: 部分更新

**ユーザーストーリー:** 開発者として、変更があった領域のみを再合成したい。これにより、全画面再描画を避けられる。

#### 受け入れ基準

1. THE Layer_Manager SHALL 変更があった領域（ダーティ領域）を追跡する
2. WHEN 合成処理を行うとき THEN THE Layer_Manager SHALL ダーティ領域のみを再合成する
3. WHEN 複数のダーティ領域があるとき THEN THE Layer_Manager SHALL それらを統合して処理する
4. THE Layer_Manager SHALL ダーティ領域の境界ボックスを計算する

### 要件 7: 上書きスキップ

**ユーザーストーリー:** 開発者として、完全に覆われたレイヤーの描画をスキップしたい。これにより、不要な描画処理を削減できる。

#### 受け入れ基準

1. WHEN 不透明なレイヤーが別のレイヤーを完全に覆っているとき THEN THE Layer_Manager SHALL 覆われたレイヤーの描画をスキップする
2. THE Layer_Manager SHALL 各レイヤーの不透明度を追跡する
3. THE Layer_Manager SHALL レイヤー間の重なり判定を行う

### 要件 8: 既存コードとの互換性

**ユーザーストーリー:** 開発者として、既存のAPIを変更せずにレイヤーベース描画を導入したい。これにより、既存のコードを壊さずに改善できる。

#### 受け入れ基準

1. THE Layer_Manager SHALL pkg/graphics/graphics.goのDraw関数と統合する
2. THE Layer_Manager SHALL pkg/graphics/cast.goのキャスト管理と統合する
3. THE Layer_Manager SHALL pkg/graphics/text_layer.goのテキストレイヤーと統合する
4. THE Layer_Manager SHALL pkg/vm/vm.goのPutCast/MoveCast実装と統合する
5. WHEN 既存のAPIが呼び出されたとき THEN THE Layer_Manager SHALL 内部的にレイヤー操作に変換する

### 要件 9: パフォーマンス要件

**ユーザーストーリー:** 開発者として、60fpsを維持できる描画パフォーマンスを確保したい。これにより、スムーズなゲーム体験を提供できる。

#### 受け入れ基準

1. THE Layer_Manager SHALL 典型的なFILLYタイトル（10〜20レイヤー）の合成を16.67ms（60fps）以内に完了する
   - 注: 100レイヤーでの60fps達成は現実的ではないが、実際のFILLYタイトルでは10〜20レイヤー程度が典型的
   - 10レイヤーで約4.6ms（約217fps相当）を達成
2. THE Layer_Manager SHALL テキストレイヤーのキャッシュを使用して、テキスト描画のコストを削減する
3. THE Layer_Manager SHALL ダーティフラグを使用して、不要な再描画を避ける
4. THE Layer_Manager SHALL 可視領域クリッピングを使用して、不要な描画をスキップする

### 要件 10: 操作順序に基づくZ順序管理

**ユーザーストーリー:** 開発者として、FILLYの「焼き付け」方式と同等の描画結果を得たい。後から実行した描画操作が前面に表示される。

#### 受け入れ基準

1. THE Layer_Manager SHALL すべての描画操作（MovePic、PutCast、TextWrite）に対して共通のZ順序カウンターを使用する
2. WHEN 描画操作が実行されたとき THEN THE Layer_Manager SHALL 現在のZ順序カウンターを割り当て、カウンターを増加させる
3. WHEN MovePicがPutCastの後に呼び出されたとき THEN MovePicの内容はCastの上に表示される
4. WHEN PutCastがMovePicの後に呼び出されたとき THEN CastはMovePicの内容の上に表示される
5. THE Layer_Manager SHALL 合成時にすべてのレイヤーをZ順序でソートして描画する
6. **例**: MovePic(Z=1) → PutCast(Z=2) → MovePic(Z=3) の場合、Z=3のMovePicがZ=2のCastの上に表示される
