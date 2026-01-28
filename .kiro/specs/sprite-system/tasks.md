# 実装計画: スプライトシステム (Sprite System)

## 概要

FILLYエミュレータのグラフィックスシステムを簡素化するため、汎用スプライトシステムを実装します。

## タスク

- [x] 1. 基本スプライト構造体の実装
  - [x] 1.1 Sprite構造体を定義する（ID、Image、X、Y、ZOrder、Visible、Alpha、Parent、Dirty）
    - _Requirements: 1.1〜1.8_
  - [x] 1.2 Spriteのゲッター・セッターを実装する
    - _Requirements: 1.1〜1.8_
  - [x] 1.3 Spriteの単体テストを作成する

- [x] 2. 親子関係の実装
  - [x] 2.1 AbsolutePosition()を実装する
    - _Requirements: 2.1_
  - [x] 2.2 EffectiveAlpha()を実装する
    - _Requirements: 2.2_
  - [x] 2.3 IsEffectivelyVisible()を実装する
    - _Requirements: 2.3_
  - [x] 2.4 親子関係の単体テストを作成する

- [x] 3. SpriteManagerの実装
  - [x] 3.1 SpriteManager構造体を定義する
    - _Requirements: 3.1〜3.6_
  - [x] 3.2 CreateSprite()を実装する
    - _Requirements: 3.1_
  - [x] 3.3 CreateSpriteWithSize()を実装する
    - _Requirements: 3.2_
  - [x] 3.4 GetSprite()を実装する
    - _Requirements: 3.3_
  - [x] 3.5 RemoveSprite()を実装する
    - _Requirements: 3.4_
  - [x] 3.6 Clear()とCount()を実装する
    - _Requirements: 3.5, 3.6_
  - [x] 3.7 SpriteManagerの単体テストを作成する

- [x] 4. Z順序描画の実装
  - [x] 4.1 sortSprites()を実装する
    - _Requirements: 4.1, 10.1_
  - [x] 4.2 Draw()を実装する
    - _Requirements: 4.1〜4.5_
  - [x] 4.3 MarkNeedSort()を実装する
    - _Requirements: 10.2_
  - [x] 4.4 描画の単体テストを作成する

- [x] 5. テキストスプライトヘルパーの実装
  - [x] 5.1 CreateTextSpriteImage()を実装する（差分抽出方式）
    - _Requirements: 5.1, 5.2_
  - [x] 5.2 テキストスプライトの単体テストを作成する

- [x] 6. プロパティベーステストの作成
  - [x] 6.1 Property 1: スプライトID管理のテスト
    - **Validates: Requirements 1.1, 3.1, 3.3**
  - [x] 6.2 Property 2: 親子関係の位置計算のテスト
    - **Validates: Requirements 2.1**
  - [x] 6.3 Property 3: 親子関係の透明度計算のテスト
    - **Validates: Requirements 2.2**
  - [x] 6.4 Property 4: 親子関係の可視性のテスト
    - **Validates: Requirements 2.3, 4.3**
  - [x] 6.5 Property 5: Z順序による描画順のテスト
    - **Validates: Requirements 4.1**
  - [x] 6.6 Property 6: テキスト差分抽出のテスト
    - **Validates: Requirements 5.1, 5.2**

- [x] 7. チェックポイント - 基本機能の確認
  - すべてのテストが通ることを確認し、問題があればユーザーに質問する

- [x] 8. GraphicsSystemへのSpriteManager統合
  - [x] 8.1 GraphicsSystem構造体にSpriteManagerフィールドを追加する
  - [x] 8.2 NewGraphicsSystem()でSpriteManagerを初期化する
  - [x] 8.3 GetSpriteManager()メソッドを追加する
  - [x] 8.4 統合の単体テストを作成する

- [x] 9. ウインドウのスプライト化
  - [x] 9.1 WindowSpriteラッパー構造体を作成する（Window + Sprite）
  - [x] 9.2 OpenWin()でWindowSpriteを作成するように変更する
  - [x] 9.3 drawWindowDecoration()をスプライトベースに変更する
  - [x] 9.4 ウインドウスプライトの単体テストを作成する
  - [x] 9.5 サンプルタイトル（kuma2）でウインドウ表示を確認する

- [x] 10. ピクチャ描画（MovePic）のスプライト化
  - [x] 10.1 DrawingEntryをSpriteベースに変換するアダプタを作成する
  - [x] 10.2 MovePic()でスプライトを作成するように変更する
  - [x] 10.3 drawDrawingEntryOnScreen()をスプライトベースに変更する
  - [x] 10.4 ピクチャ描画の単体テストを作成する
  - [x] 10.5 サンプルタイトル（kuma2）でピクチャ描画を確認する

- [x] 11. キャスト描画のスプライト化
  - [x] 11.1 CastSpriteラッパー構造体を作成する（Cast + Sprite）
  - [x] 11.2 PutCast()でCastSpriteを作成するように変更する
  - [x] 11.3 MoveCast()でCastSpriteを更新するように変更する
  - [x] 11.4 DelCast()でCastSpriteを削除するように変更する
  - [x] 11.5 drawCastLayerOnScreen()をスプライトベースに変更する
  - [x] 11.6 透明色処理をスプライト描画に統合する
  - [x] 11.7 キャスト描画の単体テストを作成する
  - [x] 11.8 サンプルタイトル（home）でキャスト描画を確認する

- [x] 12. テキスト描画のスプライト化
  - [x] 12.1 TextSpriteラッパー構造体を作成する（TextLayerEntry + Sprite）
  - [x] 12.2 TextWrite()でTextSpriteを作成するように変更する
  - [x] 12.3 drawTextLayerOnScreen()をスプライトベースに変更する
  - [x] 12.4 テキスト描画の単体テストを作成する
  - [x] 12.5 サンプルタイトル（kuma2）でテキスト描画を確認する

- [ ] 13. 図形描画のスプライト化
  - [x] 13.1 ShapeSpriteラッパー構造体を作成する（図形情報 + Sprite）
  - [x] 13.2 DrawLine/DrawRect/DrawCircle/FillRectでShapeSpriteを作成するように変更する
  - [x] 13.3 図形描画の単体テストを作成する
  - [ ] 13.4 サンプルタイトル（ftile400）で図形描画を確認する

- [ ] 14. 描画システムの統合
  - [x] 14.1 GraphicsSystem.Draw()をSpriteManager.Draw()ベースに変更する
  - [x] 14.2 ウインドウ内のスプライトをウインドウの子スプライトとして管理する
  - [x] 14.3 Z順序の統一（ウインドウ間、ウインドウ内）を実装する
  - [x] 14.4 統合テストを作成する
  - [ ] 14.5 サンプルタイトル（kuma2, home）で全体動作を確認する

- [-] 15. 旧レイヤーシステムの削除
  - [x] 15.1 LayerManagerの使用箇所を確認し、スプライトシステムに置き換える
  - [x] 15.2 WindowLayerSetの使用箇所を確認し、スプライトシステムに置き換える
  - [x] 15.3 PictureLayerSetの使用箇所を確認し、スプライトシステムに置き換える
  - [x] 15.4 CastLayer, DrawingEntry, TextLayerEntryの使用箇所を確認する
  - [x] 15.5 不要になったレイヤー関連ファイルを削除する
  - [x] 15.6 不要になったテストファイルを削除する
  - [x] 15.7 全テストがパスすることを確認する

- [ ] 16. 最終動作確認
  - [ ] 16.1 サンプルタイトル（kuma2）で動作確認
  - [ ] 16.2 サンプルタイトル（home）で動作確認
  - [ ] 16.3 サンプルタイトル（ftile400）で動作確認
  - [ ] 16.4 既存テストがすべてパスすることを確認

## 注意事項

- 各プロパティテストは最低100回の反復を実行するよう設定してください
- go testコマンドには `-timeout` オプションを使用してください
- 既存のテストが壊れないよう注意してください
- 各タスク完了後にサンプルタイトルで動作確認を行い、問題があれば早期に発見してください
- 旧システムと新システムを並行稼働させながら段階的に移行してください
