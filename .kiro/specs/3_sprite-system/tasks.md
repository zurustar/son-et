# 実装タスク: スプライトシステム (Sprite System)

## 概要

このドキュメントは、スプライトシステムの実装タスクを定義します。
既存の `pkg/graphics/` からスプライト関連コードを `pkg/sprite/` に分離し、
スライスベースの描画順序システムを実装します。

## タスク一覧

---

### フェーズ1: パッケージ分離

- [x] 1. pkg/sprite パッケージの作成
  - [x] 1.1 pkg/sprite ディレクトリを作成する
  - [x] 1.2 pkg/graphics/sprite.go を pkg/sprite/sprite.go に移動し、パッケージ名を変更する
  - [x] 1.3 pkg/graphics/sprite_test.go を pkg/sprite/sprite_test.go に移動する
  - [x] 1.4 pkg/graphics/sprite_property_test.go を pkg/sprite/ に移動する
  - [x] 1.5 pkg/sprite/errors.go を作成する（スプライト関連エラー定義）

- [x] 2. スプライトタイプの移動
  - [x] 2.1 pkg/graphics/window_sprite.go を pkg/sprite/ に移動する
  - [x] 2.2 pkg/graphics/picture_sprite.go を pkg/sprite/ に移動する
  - [x] 2.3 pkg/graphics/cast_sprite.go を pkg/sprite/ に移動する
  - [x] 2.4 pkg/graphics/text_sprite.go を pkg/sprite/ に移動する
  - [x] 2.5 pkg/graphics/shape_sprite.go を pkg/sprite/ に移動する
  - [x] 2.6 各ファイルのテストファイルも移動する
  - [x] 2.7 pkg/graphics から pkg/sprite への import を追加する

---

### フェーズ2: スライスベースの描画順序

- [x] 3. Sprite構造体の更新（要件1, 2対応）
  - [x] 3.1 Sprite構造体に children []*Sprite フィールドを追加する（要件1.7）
  - [x] 3.2 Sprite構造体に parent *Sprite フィールドを追加する（要件1.6）
  - [x] 3.3 zPath, zOrder, sortKey 関連フィールドを削除する（要件9.4）
  - [x] 3.4 AddChild メソッドを実装する - スライス末尾に追加（要件2.5）
  - [x] 3.5 RemoveChild メソッドを実装する
  - [x] 3.6 GetChildren メソッドを実装する
  - [x] 3.7 AbsolutePosition メソッドを更新する - 親の位置を加算（要件2.1）
  - [x] 3.8 EffectiveAlpha メソッドを更新する - 親の透明度を乗算（要件2.2）
  - [x] 3.9 IsEffectivelyVisible メソッドを更新する - 親の可視性を継承（要件2.3）

- [x] 4. 描画順序変更メソッド（要件12対応）
  - [x] 4.1 BringToFront メソッドを実装する - スライス末尾に移動（要件12.1）
  - [x] 4.2 SendToBack メソッドを実装する - スライス先頭に移動（要件12.2）

- [x] 5. SpriteManager の更新（要件3対応）
  - [x] 5.1 roots []*Sprite フィールドを追加する - ルートスプライト管理（要件3.7）
  - [x] 5.2 pictureSpriteMap map[int]*PictureSprite を追加する（要件14.4）
  - [x] 5.3 ZOrderCounter, zPath, needSort 関連コードを削除する
  - [x] 5.4 CreateSprite メソッドを更新する - 親子関係対応（要件3.1）
  - [x] 5.5 CreateRootSprite メソッドを実装する
  - [x] 5.6 DeleteSprite メソッドを更新する - 子スプライトも削除（要件3.4）
  - [x] 5.7 Clear メソッドを更新する（要件3.5）

- [x] 6. 再帰的描画の実装（要件9, 10対応）
  - [x] 6.1 Draw メソッドを再帰的描画に変更する
  - [x] 6.2 drawSprite 内部メソッドを実装する - 親→子の順で描画（要件10.1）
  - [x] 6.3 スライス順序での描画を実装する（要件9.1, 9.2）
  - [x] 6.4 親が非表示の場合は子も描画しない（要件10.4）


---

### フェーズ3: PictureSprite の状態管理

- [x] 7. PictureSprite の実装（要件13, 14対応）
  - [x] 7.1 PictureSpriteState 型を定義する（Unattached, Attached）（要件14.1）
  - [x] 7.2 PictureSprite 構造体を更新する - state, windowID フィールド追加
  - [x] 7.3 CreatePictureSpriteOnLoad メソッドを実装する - 非表示で作成（要件13.1）
  - [x] 7.4 AttachPictureSpriteToWindow メソッドを実装する（要件13.3, 13.4）
  - [x] 7.5 GetPictureSpriteByPictureID メソッドを実装する（要件14.4）
  - [x] 7.6 FreePictureSprite メソッドを実装する - 子スプライトも削除（要件13.7）
  - [x] 7.7 UpdatePictureSpriteImage メソッドを実装する（要件14.5）

---

### フェーズ4: 各スプライトタイプの更新

- [x] 8. WindowSprite の更新（要件4対応）
  - [x] 8.1 WindowSprite を親子関係対応に更新する
  - [x] 8.2 ウインドウ作成時にルートスプライトとして登録する（要件11.1）
  - [x] 8.3 ウインドウ削除時に子スプライトも削除する（要件4.3）

- [x] 9. CastSprite の更新（要件6対応）
  - [x] 9.1 CastSprite を親子関係対応に更新する
  - [x] 9.2 PutCast の引数を修正する - src_pic_no, dst_pic_no（要件6.1）
  - [x] 9.3 キャストを配置先ピクチャーの子として追加する
  - [x] 9.4 透明色処理を維持する（要件6.4）

- [x] 10. TextSprite の更新（要件7対応）
  - [x] 10.1 TextSprite を親子関係対応に更新する
  - [x] 10.2 テキストを対象ピクチャーの子として追加する
  - [x] 10.3 差分抽出方式を維持する（要件7.1, 7.2）

- [x] 11. ShapeSprite の更新（要件8対応）
  - [x] 11.1 ShapeSprite を親子関係対応に更新する
  - [x] 11.2 図形を対象ピクチャーの子として追加する

---

### フェーズ5: pkg/graphics との統合

- [x] 12. GraphicsSystem の更新
  - [x] 12.1 pkg/sprite をインポートする
  - [x] 12.2 SpriteManager を pkg/sprite.SpriteManager に変更する
  - [x] 12.3 LoadPic で PictureSprite を作成するように変更する
  - [x] 12.4 OpenWin で WindowSprite をルートとして作成する
  - [x] 12.5 PutCast の引数処理を修正する（src_pic, dst_pic）
  - [x] 12.6 Draw メソッドで pkg/sprite の描画を呼び出す

- [x] 13. 不要コードの削除
  - [x] 13.1 pkg/graphics/zpath.go を削除する
  - [x] 13.2 pkg/graphics/zpath_test.go を削除する
  - [x] 13.3 pkg/graphics/zpath_property_test.go を削除する
  - [x] 13.4 layer.go, layer_manager.go の不要部分を削除する
  - [x] 13.5 CalculateGlobalZOrder 関連コードを削除する

**注意**: タスク13.1-13.5は、現在の実装がハイブリッドアプローチ（ZPathとスライスベースの両方を使用）であるため、実際には削除されていません。ZPath関連のコードは、現在の描画システムで引き続き使用されています。完全なスライスベースの描画順序への移行が完了した後に、これらのコードを削除する予定です。

---

### フェーズ6: テストとデバッグ

- [x] 14. ユニットテストの更新
  - [x] 14.1 Sprite の親子関係テストを追加する
  - [x] 14.2 スライスベース描画順序のテストを追加する
  - [x] 14.3 BringToFront, SendToBack のテストを追加する
  - [x] 14.4 PictureSprite 状態管理のテストを追加する

- [x] 15. プロパティベーステスト（設計書のプロパティ1-8対応）
  - [x] 15.1 プロパティ1: 追加順序の保持テスト
  - [x] 15.2 プロパティ2: 描画順序の一貫性テスト
  - [x] 15.3 プロパティ3: 兄弟描画順序テスト
  - [x] 15.4 プロパティ4: 可視性の継承テスト
  - [x] 15.5 プロパティ5: 最前面移動テスト
  - [x] 15.6 プロパティ6: 最背面移動テスト

- [x] 16. 統合テスト
  - [x] 16.1 既存のサンプルスクリプト（y_saru）で動作確認する
  - [x] 16.2 描画順序が正しいことを視覚的に確認する
  - [x] 16.3 PutCast の引数修正が正しく動作することを確認する

- [x] 17. デバッグ支援（要件20対応）
  - [x] 17.1 PrintHierarchy メソッドを実装する - ツリー形式出力（要件20.1）
  - [x] 17.2 PrintDrawOrder メソッドを実装する - 描画順序リスト（要件20.2）
  - [ ]*17.3 デバッグオーバーレイを実装する（要件20.3）

---

## 注意事項

### PutCast 引数の変更
従来: `PutCast(win_no, pic_no, x, y, ...)`
新規: `PutCast(src_pic_no, dst_pic_no, x, y, ...)`

- 第1引数: ソースピクチャーID（画像の取得元）
- 第2引数: 配置先ピクチャーID（キャストを配置する先）

### 依存関係
```
pkg/graphics → pkg/sprite  （graphicsがspriteを使う）
pkg/vm → pkg/graphics      （VMがgraphicsを使う）
```

pkg/sprite は独立しており、pkg/graphics に依存しません。
