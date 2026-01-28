# 実装計画: レイヤーシステム再設計 (Layer System Redesign)

## 概要

FILLYエミュレータのグラフィックスレイヤーシステムを再設計し、レイヤーをPictureIDではなくWindowIDで管理するように変更します。これにより、現在のバグ（`createCastLayer`がWinIDで登録、`drawLayersForWindow`がPicIDで検索）を解消します。

## タスク

- [x] 1. WindowLayerSetの実装
  - [x] 1.1 WindowLayerSet構造体を定義する
    - WinID、BgColor、Width、Height、Layers、nextZOrder、CompositeBuffer、ダーティフラグを含む
    - _Requirements: 1.1, 1.4_
  
  - [x] 1.2 LayerManagerにwindowLayersマップを追加する
    - map[int]*WindowLayerSet として定義
    - 既存のlayersマップは後方互換性のために残す
    - _Requirements: 1.1, 1.5_
  
  - [x] 1.3 WindowLayerSetのCRUD操作を実装する
    - GetOrCreateWindowLayerSet(winID)
    - GetWindowLayerSet(winID)
    - DeleteWindowLayerSet(winID)
    - _Requirements: 1.2, 1.3_
  
  - [x] 1.4 WindowLayerSetのプロパティテストを作成する
    - **Property 1: レイヤーのWindowID管理**
    - **Property 2: ウィンドウ開閉時のレイヤーセット管理**
    - **Validates: Requirements 1.1, 1.2, 1.3, 1.5**

- [x] 2. PictureLayerの実装
  - [x] 2.1 PictureLayer構造体を定義する
    - BaseLayerを埋め込み
    - image（ウィンドウサイズの透明画像）
    - bakeable（焼き付け可能フラグ）
    - _Requirements: 2.1, 2.5_
  
  - [x] 2.2 PictureLayerのメソッドを実装する
    - NewPictureLayer(id, winWidth, winHeight)
    - Bake(src, destX, destY)
    - IsBakeable()
    - GetLayerType() → LayerTypePicture
    - _Requirements: 2.1, 2.4, 2.5_
  
  - [x] 2.3 PictureLayerの単体テストを作成する
    - 焼き付け動作の確認
    - _Requirements: 2.1, 2.5_

- [x] 3. LayerTypeの追加
  - [x] 3.1 LayerType定数を定義する
    - LayerTypePicture、LayerTypeText、LayerTypeCast
    - _Requirements: 2.4_
  
  - [x] 3.2 Layerインターフェースに GetLayerType() を追加する
    - 既存のレイヤー型に実装を追加
    - _Requirements: 2.4_
  
  - [x] 3.3 レイヤータイプ識別のプロパティテストを作成する
    - **Property 3: レイヤータイプの識別**
    - **Validates: Requirements 2.4**

- [x] 4. チェックポイント - 基本構造の確認
  - すべてのテストが通ることを確認し、問題があればユーザーに質問する

- [x] 5. MovePicの焼き付けロジック実装
  - [x] 5.1 getTopmostLayer()メソッドを実装する
    - WindowLayerSetの最上位レイヤーを取得
    - _Requirements: 3.1_
  
  - [x] 5.2 movePicInternalを修正する
    - 最上位レイヤーのタイプを確認
    - Picture_Layerなら焼き付け、そうでなければ新規作成
    - _Requirements: 3.2, 3.3, 3.4_
  
  - [x] 5.3 createPictureLayerForWindow()を実装する
    - ウィンドウサイズの透明PictureLayerを作成
    - WindowLayerSetに追加
    - _Requirements: 3.5_
  
  - [x] 5.4 焼き付け後のダーティフラグ設定を実装する
    - _Requirements: 3.6_
  
  - [x] 5.5 焼き付けロジックのプロパティテストを作成する
    - **Property 4: MovePicの焼き付けロジック**
    - **Validates: Requirements 3.2, 3.3, 3.4, 3.5, 3.6**

- [x] 6. CastManagerの修正
  - [x] 6.1 createCastLayerをWindowLayerSetに登録するよう修正する
    - windowLayersマップを使用
    - _Requirements: 7.1_
  
  - [x] 6.2 updateCastLayerをWindowLayerSetを使用するよう修正する
    - _Requirements: 4.2_
  
  - [x] 6.3 deleteCastLayerをWindowLayerSetを使用するよう修正する
    - _Requirements: 4.3_
  
  - [x] 6.4 Cast_Layerのプロパティテストを作成する
    - **Property 6: Cast_Layerのスプライト動作**
    - **Validates: Requirements 4.1, 4.2, 4.3, 4.6**

- [x] 7. TextRendererの修正
  - [x] 7.1 TextWriteをWindowLayerSetに登録するよう修正する
    - ピクチャーIDからウィンドウIDを逆引き
    - _Requirements: 7.3_
  
  - [x] 7.2 Text_Layerの新規作成を確認する
    - 既存の実装が正しいことを確認
    - _Requirements: 5.1, 5.2_
  
  - [x] 7.3 Text_Layerのプロパティテストを作成する
    - **Property 5: Text_Layerの新規作成**
    - **Validates: Requirements 2.6, 5.1, 5.2**

- [x] 8. チェックポイント - レイヤー登録の確認
  - すべてのテストが通ることを確認し、問題があればユーザーに質問する

- [x] 9. drawLayersForWindowの修正
  - [x] 9.1 drawLayersForWindowをWindowIDで検索するよう修正する
    - win.PicID → win.ID に変更
    - _Requirements: 7.4_
  
  - [x] 9.2 背景色の描画を追加する
    - WindowLayerSetのBgColorを使用
    - _Requirements: 6.1_
  
  - [x] 9.3 Z順序でのソートを確認する
    - GetAllLayersSorted()が正しく動作することを確認
    - _Requirements: 6.2_
  
  - [x] 9.4 Z順序のプロパティテストを作成する
    - **Property 7: Z順序の管理**
    - **Validates: Requirements 6.2, 6.3, 6.4**

- [x] 10. ピクチャーID逆引きの実装
  - [x] 10.1 WindowManagerにGetWinByPicIDを確認する
    - 既存の実装が正しいことを確認
    - _Requirements: 7.5_
  
  - [x] 10.2 LayerManagerにpicToWinMappingを追加する（必要な場合）
    - ピクチャーIDからウィンドウIDへのマッピング
    - _Requirements: 7.5_
    - **結論: 不要** - WindowManager.GetWinByPicID()が既に実装済みで、要件7.5を満たしている。LayerManagerに重複したマッピングを持つ必要はない。
  
  - [x] 10.3 レイヤー登録のプロパティテストを作成する
    - **Property 8: レイヤーのウィンドウ登録**
    - **Validates: Requirements 7.1, 7.2, 7.3, 7.5**

- [x] 11. ウィンドウ操作との統合
  - [x] 11.1 OpenWinでWindowLayerSetを作成するよう修正する
    - _Requirements: 1.2_
  
  - [x] 11.2 CloseWinでWindowLayerSetを削除するよう修正する
    - _Requirements: 1.3_
  
  - [x] 11.3 CloseWinAllでWindowLayerSetをすべて削除するよう修正する
    - _Requirements: 1.3_

- [x] 12. チェックポイント - 統合の確認
  - すべてのテストが通ることを確認し、問題があればユーザーに質問する

- [x] 13. ダーティフラグとキャッシュの最適化
  - [x] 13.1 WindowLayerSetのダーティ領域追跡を実装する
    - _Requirements: 9.1, 9.2_
  
  - [x] 13.2 合成処理完了後のダーティフラグクリアを実装する
    - _Requirements: 9.3_
  
  - [x] 13.3 ダーティフラグのプロパティテストを作成する
    - **Property 9: ダーティフラグの動作**
    - **Validates: Requirements 9.1, 9.2**

- [x] 14. エラーハンドリングの実装
  - [x] 14.1 存在しないウィンドウIDのエラー処理を実装する
    - _Requirements: 10.1_
  
  - [x] 14.2 存在しないレイヤーIDのエラー処理を実装する
    - _Requirements: 10.2_
  
  - [x] 14.3 レイヤー作成失敗のエラー処理を実装する
    - _Requirements: 10.3_
  
  - [x] 14.4 エラーハンドリングのプロパティテストを作成する
    - **Property 10: エラーハンドリング**
    - **Validates: Requirements 10.1, 10.2, 10.3, 10.4**

- [x] 15. 後方互換性の確認
  - [x] 15.1 既存のCastManager APIが動作することを確認する
    - _Requirements: 8.1_
  
  - [x] 15.2 既存のTextRenderer APIが動作することを確認する
    - _Requirements: 8.2_
  
  - [x] 15.3 既存のMovePic APIが動作することを確認する
    - _Requirements: 8.3_
  
  - [x] 15.4 既存のテストがすべて通ることを確認する
    - _Requirements: 8.4_

- [x] 16. 最終チェックポイント
  - すべてのテストが通ることを確認し、問題があればユーザーに質問する

## 注意事項

- 各プロパティテストは最低100回の反復を実行するよう設定してください
- Go言語のプロパティベーステストには `testing/quick` または `github.com/leanovate/gopter` を使用してください
- 既存のテストが壊れないよう、後方互換性に注意してください
- go testコマンドには `-timeout` オプションを使用してください（無限ループ対策）
