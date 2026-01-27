# 実装計画: レイヤーベース描画システム

## 概要

レイヤーベース描画システムを実装するための段階的なタスクリストです。各タスクは前のタスクに基づいて構築され、最終的にすべてのコンポーネントが統合されます。

## タスク

- [x] 1. 基本的なレイヤー構造の実装
  - [x] 1.1 Layerインターフェースの定義
    - pkg/graphics/layer.go を作成
    - GetID, GetBounds, GetZOrder, IsVisible, IsDirty, SetDirty, GetImage, Invalidate メソッドを定義
    - _Requirements: 1.1, 1.2, 1.3, 1.4_
  - [x] 1.2 BackgroundLayerの実装
    - 背景画像を保持するレイヤー
    - Z順序は常に0
    - _Requirements: 1.1_
  - [x] 1.3 DrawingLayerの実装
    - MovePicで描画された内容を保持するレイヤー
    - Z順序は常に1
    - _Requirements: 1.3_
  - [x] 1.4 CastLayerの実装
    - キャストを保持するレイヤー
    - 透明色処理、キャッシュ機能を含む
    - Z順序は100から開始
    - _Requirements: 1.2_
  - [x] 1.5 TextLayerEntryの実装
    - テキストを保持するレイヤー
    - キャッシュ機能を含む
    - Z順序は1000から開始
    - _Requirements: 1.4_
  - [x] 1.6 レイヤー構造のユニットテスト
    - 各レイヤータイプの作成テスト
    - Z順序の割り当てテスト
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5_

- [x] 2. LayerManagerの実装
  - [x] 2.1 LayerManager構造体の定義
    - pkg/graphics/layer_manager.go を作成
    - PictureLayerSetの管理
    - ミューテックスによる同期
    - _Requirements: 1.1, 1.2, 1.3, 1.4_
  - [x] 2.2 PictureLayerSetの実装
    - ピクチャーに属するレイヤーのセット
    - 背景、描画、キャスト、テキストレイヤーの管理
    - 合成バッファとダーティ領域の管理
    - _Requirements: 1.6_
  - [x] 2.3 レイヤー追加・削除メソッドの実装
    - AddCastLayer, RemoveCastLayer
    - AddTextLayer, RemoveTextLayer
    - Z順序の自動割り当て
    - _Requirements: 1.5, 2.1, 2.3, 2.5_
  - [x] 2.4 LayerManagerのユニットテスト
    - レイヤー追加・削除テスト
    - Z順序の管理テスト
    - _Requirements: 2.1, 2.3, 2.5_
  - [x] 2.5 Property 1: レイヤー管理の一貫性テスト
    - **Property 1: レイヤー管理の一貫性**
    - **Validates: Requirements 1.1, 1.2, 1.3, 1.4, 1.5, 1.6**

- [x] 3. ダーティフラグとキャッシュの実装
  - [x] 3.1 ダーティフラグの実装
    - 各レイヤーにダーティフラグを追加
    - 位置、内容、可視性の変更時にフラグを設定
    - _Requirements: 3.1, 3.2, 3.3_
  - [x] 3.2 キャッシュ管理の実装
    - レイヤーの描画結果をキャッシュ
    - ダーティフラグに基づいてキャッシュを使用/無効化
    - _Requirements: 5.1, 5.2, 5.3_
  - [x] 3.3 ダーティ領域追跡の実装
    - 変更があった領域を追跡
    - 複数のダーティ領域を統合
    - _Requirements: 6.1, 6.2, 6.3, 6.4_
  - [x] 3.4 ダーティフラグとキャッシュのユニットテスト
    - ダーティフラグの設定・クリアテスト
    - キャッシュの使用・無効化テスト
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5, 5.1, 5.2, 5.3_
  - [x] 3.5 Property 3: ダーティフラグの正確性テスト
    - **Property 3: ダーティフラグの正確性**
    - **Validates: Requirements 3.1, 3.2, 3.3, 3.4, 3.5**
  - [x] 3.6 Property 5: キャッシュ管理の正確性テスト
    - **Property 5: キャッシュ管理の正確性**
    - **Validates: Requirements 5.1, 5.2, 5.3**

- [x] 4. チェックポイント - 基本機能の確認
  - すべてのテストが通ることを確認し、問題があれば質問する

- [x] 5. 合成処理の実装
  - [x] 5.1 可視領域クリッピングの実装
    - レイヤーの境界ボックス計算
    - 可視領域との交差判定
    - 可視領域外のレイヤーをスキップ
    - _Requirements: 4.1, 4.2, 4.3, 4.4_
  - [x] 5.2 上書きスキップの実装
    - レイヤーの不透明度追跡
    - レイヤー間の重なり判定
    - 完全に覆われたレイヤーをスキップ
    - _Requirements: 7.1, 7.2, 7.3_
  - [x] 5.3 合成処理の実装
    - 背景 → 描画 → キャスト → テキストの順で合成
    - ダーティ領域のみを再合成
    - 合成後にダーティフラグをクリア
    - _Requirements: 1.6, 6.2_
  - [x] 5.4 合成処理のユニットテスト
    - 可視領域クリッピングテスト
    - 上書きスキップテスト
    - 合成順序テスト
    - _Requirements: 4.1, 4.2, 7.1, 1.6_
  - [x] 5.5 Property 4: 可視領域クリッピングの正確性テスト
    - **Property 4: 可視領域クリッピングの正確性**
    - **Validates: Requirements 4.1, 4.2, 4.3, 4.4**
  - [x] 5.6 Property 6: ダーティ領域追跡の正確性テスト
    - **Property 6: ダーティ領域追跡の正確性**
    - **Validates: Requirements 6.1, 6.2, 6.3, 6.4**
  - [x] 5.7 Property 7: 上書きスキップの正確性テスト
    - **Property 7: 上書きスキップの正確性**
    - **Validates: Requirements 7.1, 7.2, 7.3**

- [x] 6. 既存コードとの統合
  - [x] 6.1 GraphicsSystemへのLayerManager統合
    - pkg/graphics/graphics.go を修正
    - LayerManagerをGraphicsSystemに追加
    - Draw関数でLayerManagerを使用
    - _Requirements: 8.1_
  - [x] 6.2 CastManagerとの統合
    - pkg/graphics/cast.go を修正
    - PutCast/MoveCast/DelCastでLayerManagerを使用
    - _Requirements: 8.2, 2.1, 2.2, 2.3_
  - [x] 6.3 TextRendererとの統合
    - pkg/graphics/text_layer.go を修正
    - TextWriteでLayerManagerを使用
    - _Requirements: 8.3, 2.5_
  - [x] 6.4 VMとの統合
    - pkg/vm/vm.go を修正（必要に応じて）
    - PutCast/MoveCast実装がLayerManagerを使用することを確認
    - _Requirements: 8.4, 8.5_
  - [x] 6.5 統合テスト
    - GraphicsSystem統合テスト
    - VM統合テスト
    - _Requirements: 8.1, 8.2, 8.3, 8.4_
  - [x] 6.6 Property 2: レイヤー操作の整合性テスト
    - **Property 2: レイヤー操作の整合性**
    - **Validates: Requirements 2.1, 2.2, 2.3, 2.4, 2.5, 2.6**

- [x] 7. チェックポイント - 統合の確認
  - すべてのテストが通ることを確認し、問題があれば質問する

- [x] 8. _old_implementation2互換性の実装 **[スキップ]**
  - レイヤ方式を採用したことで、_old_implementation2の「焼き付け」方式との互換性は自動的に達成される
  - Z順序による合成（背景 < 描画 < キャスト < テキスト）で同等の視覚的結果が得られる
  - 残像問題もレイヤ方式で自動的に解決される

- [x] 9. パフォーマンス最適化
  - [x] 9.1 ベンチマークテストの実装
    - レイヤー合成ベンチマーク
    - テキストレイヤー作成ベンチマーク
    - ダーティ領域更新ベンチマーク
    - _Requirements: 9.1, 9.2_
  - [x] 9.2 パフォーマンスチューニング
    - ベンチマーク結果に基づいて最適化
    - 60fps維持を確認
    - _Requirements: 9.1, 9.2, 9.3, 9.4_

- [x] 10. 最終チェックポイント
  - すべてのテストが通ることを確認し、問題があれば質問する

- [x] 11. 操作順序に基づくZ順序管理の実装
  - [x] 11.1 DrawingEntryの実装
    - MovePicで描画された内容を保持するエントリ
    - 各MovePic呼び出しで新しいDrawingEntryを作成
    - _Requirements: 1.3, 2.4, 10.3, 10.4_
  - [x] 11.2 PictureLayerSetの構造変更
    - Casts/Texts/Drawingを統合したLayersスライスに変更
    - nextZOrderカウンターを追加（すべての操作で共有）
    - _Requirements: 1.5, 1.7, 10.1, 10.2_
  - [x] 11.3 Z順序カウンターの実装
    - AddCastLayer、AddTextLayer、AddDrawingEntryで共通のカウンターを使用
    - 操作のたびにカウンターを増加
    - _Requirements: 2.1, 2.4, 2.5, 2.7_
  - [x] 11.4 合成処理の修正
    - すべてのレイヤーをZ順序でソートして描画
    - 背景レイヤーは常にZ=0で最背面
    - _Requirements: 1.6, 10.5_
  - [x] 11.5 MovePicの統合
    - MovePicでDrawingEntryを作成するように変更
    - GraphicsSystem.MovePicを修正
    - _Requirements: 2.4, 10.3, 10.4_
  - [x] 11.6 テストの更新
    - 操作順序に基づくZ順序のテスト
    - MovePic → PutCast → MovePicの順序テスト
    - _Requirements: 10.3, 10.4, 10.6_

- [x] 12. 操作順序Z順序のチェックポイント
  - すべてのテストが通ることを確認
  - y_saruサンプルで動作確認

## 注意事項

- すべてのタスクは必須です
- 各プロパティテストは最低100回のイテレーションを実行
- プロパティベーステストには `github.com/leanovate/gopter` ライブラリを使用
- ベンチマークテストは既存の `pkg/graphics/layer_bench_test.go` を参考に
