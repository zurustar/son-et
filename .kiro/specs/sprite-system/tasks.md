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

- [ ] 8. 既存システムとの統合準備
  - [ ] 8.1 現在のLayerManagerとの共存方法を検討する
  - [ ] 8.2 移行計画を作成する

## 注意事項

- 各プロパティテストは最低100回の反復を実行するよう設定してください
- go testコマンドには `-timeout` オプションを使用してください
- 既存のテストが壊れないよう注意してください

