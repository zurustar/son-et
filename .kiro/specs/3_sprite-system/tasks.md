# タスクリスト: 階層的Z順序システム (Hierarchical Z-Order System)

## フェーズ1: 基盤

- [x] 1. ZPath構造体の実装
  - [x] 1.1 pkg/graphics/zpath.goファイルを作成する
  - [x] 1.2 ZPath構造体を実装する（path []int フィールド）
  - [x] 1.3 NewZPath関数を実装する（可変長引数で初期化）
  - [x] 1.4 NewZPathFromParent関数を実装する（親のZ_Pathに子のLocal_Z_Orderを追加）
  - [x] 1.5 Path(), Depth(), LocalZOrder(), Parent(), String()メソッドを実装する

- [x] 2. ZPath比較関数の実装
  - [x] 2.1 Compare関数を実装する（辞書順比較、戻り値: -1, 0, 1）
  - [x] 2.2 Less関数を実装する（sort.Interface用）
  - [x] 2.3 IsPrefix関数を実装する（プレフィックス判定）
  - [x] 2.4 Equal関数を実装する（等価判定）

- [x] 3. ZOrderCounter構造体の実装
  - [x] 3.1 ZOrderCounter構造体を実装する（counters map[int]int, mu sync.RWMutex）
  - [x] 3.2 NewZOrderCounter関数を実装する
  - [x] 3.3 GetNext関数を実装する（カウンター取得とインクリメント）
  - [x] 3.4 Reset関数を実装する（特定の親のカウンターをリセット）
  - [x] 3.5 ResetAll関数を実装する（すべてのカウンターをリセット）

- [x] 4. ZPathのユニットテスト
  - [x] 4.1 pkg/graphics/zpath_test.goファイルを作成する
  - [x] 4.2 ZPath作成のテストを実装する
  - [x] 4.3 ZPath比較のテストを実装する（辞書順、プレフィックス）
  - [x] 4.4 ZOrderCounterのテストを実装する

## フェーズ2: Sprite拡張

- [x] 5. SpriteにZ_Pathフィールドを追加
  - [x] 5.1 Sprite構造体にzPath *ZPathフィールドを追加する
  - [x] 5.2 Sprite構造体にchildren []*Spriteフィールドを追加する
  - [x] 5.3 Sprite構造体にsortKey stringフィールドを追加する（キャッシュ用）
  - [x] 5.4 ZPath()メソッドを実装する
  - [x] 5.5 SetZPath()メソッドを実装する

- [x] 6. Spriteの子スプライト管理
  - [x] 6.1 GetChildren()メソッドを実装する
  - [x] 6.2 AddChild()メソッドを実装する（親のZ_Pathを継承）
  - [x] 6.3 RemoveChild()メソッドを実装する
  - [x] 6.4 HasChildren()メソッドを実装する

- [x] 7. SpriteManagerの拡張
  - [x] 7.1 SpriteManagerにzOrderCounter *ZOrderCounterフィールドを追加する
  - [x] 7.2 NewSpriteManager()でZOrderCounterを初期化する
  - [x] 7.3 CreateSpriteWithZPath()メソッドを実装する
  - [x] 7.4 CreateRootSprite()メソッドを実装する（ウインドウ用）

## フェーズ3: ソートと描画

- [x] 8. Z_Pathによるソートの実装
  - [x] 8.1 sortSprites()メソッドをZ_Path辞書順ソートに変更する
  - [x] 8.2 needSortフラグの管理を確認する
  - [x] 8.3 ソートキャッシュの有効性を確認する

- [x] 9. 描画順序の変更
  - [x] 9.1 Draw()メソッドをZ_Path順描画に対応させる
  - [x] 9.2 親スプライトが非表示の場合、子スプライトも描画しないことを確認する
  - [x] 9.3 IsEffectivelyVisible()が正しく動作することを確認する

## フェーズ4: 動的変更

- [x] 10. BringToFrontの実装
  - [x] 10.1 BringToFront()メソッドを実装する
  - [x] 10.2 子スプライトのZ_Path再計算を実装する
  - [x] 10.3 BringToFrontのテストを実装する

- [x] 11. SendToBackの実装
  - [x] 11.1 SendToBack()メソッドを実装する
  - [x] 11.2 最小Z順序の検索を実装する
  - [x] 11.3 SendToBackのテストを実装する

- [x] 12. updateChildrenZPathsの実装
  - [x] 12.1 updateChildrenZPaths()メソッドを実装する（再帰的更新）
  - [x] 12.2 親のZ_Path変更時の子スプライト更新をテストする

## フェーズ5: 既存システムとの統合

- [x] 13. WindowSpriteの更新
  - [x] 13.1 CreateWindowSprite()でZ_Pathを設定する
  - [x] 13.2 UpdateWindowZOrder()でZ_Pathを更新する
  - [x] 13.3 子スプライトのZ_Path更新を実装する

- [x] 14. CastSpriteの更新
  - [x] 14.1 CreateCastSprite()でZ_Pathを設定する（親のZ_Pathを継承）
  - [x] 14.2 操作順序でLocal_Z_Orderを割り当てる

- [x] 15. TextSpriteの更新
  - [x] 15.1 CreateTextSprite()でZ_Pathを設定する（親のZ_Pathを継承）
  - [x] 15.2 操作順序でLocal_Z_Orderを割り当てる

- [x] 16. PictureSpriteの更新
  - [x] 16.1 CreatePictureSprite()でZ_Pathを設定する（親のZ_Pathを継承）
  - [x] 16.2 操作順序でLocal_Z_Orderを割り当てる

- [x] 17. 既存APIとの互換性
  - [x] 17.1 SetZOrder()の互換性ラッパーを実装する
  - [x] 17.2 ConvertFlatZOrderToZPath()を実装する
  - [x] 17.3 既存のCalculateGlobalZOrder()との共存を確認する

## フェーズ6: デバッグ支援

- [x] 18. Z_Pathの可視化
  - [x] 18.1 ZPathString()メソッドを実装する
  - [x] 18.2 PrintHierarchy()メソッドを実装する（ツリー形式出力）
  - [x] 18.3 PrintDrawOrder()メソッドを実装する（描画順序リスト）

- [x] 19. デバッグオーバーレイ
  - [x] 19.1 DrawDebugOverlay()メソッドを実装する
  - [x] 19.2 Z_Pathをスプライト位置に表示する機能を実装する

## フェーズ7: プロパティベーステスト

- [x] 20. ZPathのプロパティベーステスト
  - [x] 20.1 pkg/graphics/zpath_property_test.goファイルを作成する
  - [x] 20.2 プロパティ1: Z_Pathの一意性テストを実装する
  - [x] 20.3 プロパティ2: Z_Pathの継承テストを実装する
  - [x] 20.4 プロパティ3: ルートスプライトのZ_Pathテストを実装する
  - [x] 20.5 プロパティ11: 辞書順比較の正確性テストを実装する

- [x] 21. 操作順序のプロパティベーステスト
  - [x] 21.1 プロパティ4: 操作順序の反映テストを実装する
  - [x] 21.2 プロパティ5: タイプ非依存性テストを実装する

- [x] 22. 描画順序のプロパティベーステスト
  - [x] 22.1 プロパティ6: 親子描画順序テストを実装する
  - [x] 22.2 プロパティ7: 兄弟描画順序テストを実装する
  - [x] 22.3 プロパティ8: 可視性の継承テストを実装する

- [x] 23. ウインドウ間のプロパティベーステスト
  - [x] 23.1 プロパティ9: ウインドウ間の描画順序テストを実装する
  - [x] 23.2 プロパティ10: ウインドウZ順序更新の伝播テストを実装する

- [x] 24. 動的変更のプロパティベーステスト
  - [x] 24.1 プロパティ14: 最前面移動テストを実装する
  - [x] 24.2 プロパティ15: 最背面移動テストを実装する

## フェーズ8: 統合テスト

- [x] 25. 統合テスト
  - [x] 25.1 既存のサンプルスクリプト（kuma2）で動作確認する
  - [ ]* 25.2 既存のサンプルスクリプト（home）で動作確認する（スキップ：未検証サンプル）
  - [x] 25.3 既存のサンプルスクリプト（y_saru）で動作確認する
  - [x] 25.4 描画順序が操作順序に従っていることを視覚的に確認する

## フェーズ9: ピクチャのスプライト化

- [x] 26. PictureSpriteの状態管理
  - [x] 26.1 PictureSpriteState型を定義する（Unattached, Attached）
  - [x] 26.2 PictureSprite構造体にstate, windowIDフィールドを追加する
  - [x] 26.3 SpriteManagerにpictureSpriteMap map[int]*PictureSpriteを追加する

- [x] 27. LoadPic時のPictureSprite作成
  - [x] 27.1 CreatePictureSpriteOnLoad()メソッドを実装する
  - [x] 27.2 非表示状態でPictureSpriteを作成する
  - [x] 27.3 pictureSpriteMapに登録する
  - [x] 27.4 既存のPictureSpriteがあれば削除する

- [x] 28. SetPic時のPictureSprite関連付け
  - [x] 28.1 AttachPictureSpriteToWindow()メソッドを実装する
  - [x] 28.2 PictureSpriteをWindowSpriteの子として追加する
  - [x] 28.3 Z_Pathを設定する
  - [x] 28.4 状態をAttachedに変更し、表示状態にする
  - [x] 28.5 既存の子スプライトのZ_Pathを更新する

- [x] 29. ピクチャ番号からのPictureSprite検索
  - [x] 29.1 GetPictureSpriteByPictureID()メソッドを実装する
  - [x] 29.2 CastSet時にピクチャ番号からPictureSpriteを取得して親に設定する
  - [x] 29.3 TextWrite時にピクチャ番号からPictureSpriteを取得して親に設定する

- [x] 30. FreePic時のPictureSprite削除
  - [x] 30.1 FreePictureSprite()メソッドを実装する
  - [x] 30.2 子スプライトを再帰的に削除する
  - [x] 30.3 pictureSpriteMapから削除する

- [x] 31. MovePicとの連携
  - [x] 31.1 UpdatePictureSpriteImage()メソッドを実装する
  - [x] 31.2 MovePic時に転送先PictureSpriteの画像を更新する

- [x] 32. 描画時の状態チェック
  - [x] 32.1 PictureSprite.IsEffectivelyVisible()を実装する
  - [x] 32.2 未関連付け状態では描画しないことを確認する
  - [x] 32.3 関連付け済み状態では親の可視性に従うことを確認する

- [x] 33. ピクチャスプライト化のテスト
  - [x] 33.1 LoadPic時のPictureSprite作成テストを実装する
  - [x] 33.2 SetPic時の関連付けテストを実装する
  - [x] 33.3 未関連付けピクチャへのTextWrite/CastSetテストを実装する
  - [x] 33.4 FreePic時の削除テストを実装する

- [x] 34. 統合テスト（ピクチャスプライト化）
  - [x] 34.1 y_saruサンプルで動作確認する
  - [x] 34.2 hasParent=falseのログが出なくなることを確認する
  - [x] 34.3 画面表示が正しいことを確認する
