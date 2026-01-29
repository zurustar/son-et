# 実装タスク: 描画システム (Graphics System)

## フェーズ1: 基盤

- [x] 1. GraphicsSystem構造体の実装
  - [x] 1.1 `pkg/graphics/graphics.go` を作成し、GraphicsSystem構造体を定義する
  - [x] 1.2 NewGraphicsSystem()コンストラクタを実装する
  - [x] 1.3 Update()メソッドを実装する（コマンドキュー処理用）
  - [x] 1.4 Draw()メソッドを実装する（描画処理用）
  - [x] 1.5 Shutdown()メソッドを実装する

- [x] 2. CommandQueueの実装
  - [x] 2.1 `pkg/graphics/queue.go` を作成し、CommandQueue構造体を定義する
  - [x] 2.2 Push()メソッドを実装する（スレッドセーフ）
  - [x] 2.3 PopAll()メソッドを実装する
  - [x] 2.4 CommandType定数を定義する
  - [x] 2.5 プロパティテスト: コマンド実行順序（FIFO）の検証 **Validates: 要件 7.4**
  - [x] 2.6 プロパティテスト: スレッドセーフ性の検証 **Validates: 要件 7.1**

- [x] 3. 色変換ユーティリティの実装
  - [x] 3.1 `pkg/graphics/color.go` を作成する
  - [x] 3.2 ColorFromInt()を実装する（0xRRGGBB → color.Color）
  - [x] 3.3 ColorToInt()を実装する（color.Color → 0xRRGGBB）
  - [x] 3.4 ユニットテスト: 色変換の往復テスト

## フェーズ2: ピクチャーシステム

- [x] 4. PictureManagerの実装
  - [x] 4.1 `pkg/graphics/picture.go` を作成し、Picture構造体を定義する
  - [x] 4.2 PictureManager構造体を定義する
  - [x] 4.3 NewPictureManager()コンストラクタを実装する
  - [x] 4.4 LoadPic()を実装する（BMP/PNG読み込み、大文字小文字非依存検索）
  - [x] 4.5 CreatePic()を実装する（空のピクチャー生成）
  - [x] 4.6 CreatePicFrom()を実装する（既存ピクチャーからコピー生成）
  - [x] 4.7 DelPic()を実装する（ピクチャー削除）
  - [x] 4.8 GetPic()を実装する
  - [x] 4.9 PicWidth()、PicHeight()を実装する
  - [x] 4.10 プロパティテスト: ピクチャーIDの一意性 **Validates: 要件 1.2**
  - [x] 4.11 プロパティテスト: ピクチャーサイズの正確性 **Validates: 要件 1.7, 1.8**
  - [x] 4.12 プロパティテスト: 削除後のアクセスエラー **Validates: 要件 1.9**
  - [x] 4.13 プロパティテスト: リソース制限（最大256） **Validates: 要件 9.5**

## フェーズ3: ウィンドウシステム

- [x] 5. WindowManagerの実装
  - [x] 5.1 `pkg/graphics/window.go` を作成し、Window構造体を定義する
  - [x] 5.2 WindowManager構造体を定義する
  - [x] 5.3 NewWindowManager()コンストラクタを実装する
  - [x] 5.4 OpenWin()を実装する（ウィンドウ作成）
  - [x] 5.5 MoveWin()を実装する（ウィンドウ移動・変更）
  - [x] 5.6 CloseWin()を実装する（ウィンドウ削除）
  - [x] 5.7 CloseWinAll()を実装する
  - [x] 5.8 GetWin()を実装する
  - [x] 5.9 GetWindowsOrdered()を実装する（Z順序でソート）
  - [x] 5.10 プロパティテスト: ウィンドウZ順序 **Validates: 要件 3.11**
  - [x] 5.11 プロパティテスト: リソース制限（最大64） **Validates: 要件 9.6**

## フェーズ4: キャストシステム

- [x] 6. CastManagerの実装
  - [x] 6.1 `pkg/graphics/cast.go` を作成し、Cast構造体を定義する
  - [x] 6.2 CastManager構造体を定義する
  - [x] 6.3 NewCastManager()コンストラクタを実装する
  - [x] 6.4 PutCast()を実装する（キャスト配置）
  - [x] 6.5 MoveCast()を実装する（キャスト移動）
  - [x] 6.6 DelCast()を実装する（キャスト削除）
  - [x] 6.7 GetCast()を実装する
  - [x] 6.8 GetCastsByWindow()を実装する
  - [x] 6.9 DeleteCastsByWindow()を実装する
  - [x] 6.10 プロパティテスト: キャストIDの一意性 **Validates: 要件 4.2**
  - [x] 6.11 プロパティテスト: キャスト位置の更新 **Validates: 要件 4.3**
  - [x] 6.12 プロパティテスト: ウィンドウ削除時のキャスト削除 **Validates: 要件 9.2**
  - [x] 6.13 プロパティテスト: リソース制限（最大1024） **Validates: 要件 9.7**

## フェーズ5: ピクチャー転送

- [x] 7. ピクチャー転送機能の実装
  - [x] 7.1 `pkg/graphics/transfer.go` を作成する
  - [x] 7.2 MovePic()を実装する（mode=0: 通常コピー）
  - [x] 7.3 MovePic()を実装する（mode=1: 透明色除外）
  - [x] 7.4 TransPic()を実装する（指定透明色除外）
  - [x] 7.5 ReversePic()を実装する（左右反転）
  - [x] 7.6 MoveSPic()を実装する（拡大縮小転送）
  - [x] 7.7 クリッピング処理を実装する
  - [x] 7.8 ユニットテスト: 各転送モードの動作確認

## フェーズ6: テキストシステム

- [x] 8. TextRendererの実装
  - [x] 8.1 `pkg/graphics/text.go` を作成し、FontSettings構造体を定義する
  - [x] 8.2 TextSettings構造体を定義する
  - [x] 8.3 TextRenderer構造体を定義する
  - [x] 8.4 NewTextRenderer()コンストラクタを実装する
  - [x] 8.5 SetFont()を実装する（フォント設定）
  - [x] 8.6 フォントフォールバック機能を実装する（MSゴシック→Hiragino等）
  - [x] 8.7 SetTextColor()、SetBgColor()、SetBackMode()を実装する
  - [x] 8.8 TextWrite()を実装する（アンチエイリアス無効）
  - [x] 8.9 埋め込みフォントを追加する（NotoSansJP等）
  - [x] 8.10 ユニットテスト: フォントフォールバックの動作確認

## フェーズ7: 描画プリミティブ

- [x] 9. 描画プリミティブの実装
  - [x] 9.1 `pkg/graphics/primitives.go` を作成する
  - [x] 9.2 DrawLine()を実装する
  - [x] 9.3 DrawRect()を実装する
  - [x] 9.4 FillRect()を実装する
  - [x] 9.5 DrawCircle()を実装する
  - [x] 9.6 SetLineSize()を実装する
  - [x] 9.7 SetPaintColor()を実装する
  - [x] 9.8 GetColor()を実装する
  - [x] 9.9 ユニットテスト: 各描画プリミティブの動作確認

## フェーズ8: シーンチェンジ

- [x] 10. シーンチェンジの実装
  - [x] 10.1 `pkg/graphics/scene_change.go` を作成する
  - [x] 10.2 SceneChange構造体を定義する
  - [x] 10.3 SceneChangeMode定数を定義する（mode 2-9）
  - [x] 10.4 ワイプエフェクトを実装する（mode 2-7）
  - [x] 10.5 ランダムブロックエフェクトを実装する（mode 8）
  - [x] 10.6 フェードエフェクトを実装する（mode 9）
  - [x] 10.7 MovePic()にシーンチェンジモードを統合する
  - [x] 10.8 ユニットテスト: 各シーンチェンジモードの動作確認

## フェーズ9: 描画ループ統合

- [x] 11. 描画ループの統合
  - [x] 11.1 GraphicsSystem.Draw()でウィンドウを描画する
  - [x] 11.2 GraphicsSystem.Draw()でキャストを描画する（透明色除外）
  - [x] 11.3 Z順序に基づいた描画順序を実装する
  - [x] 11.4 仮想デスクトップのスケーリングを実装する
  - [x] 11.5 アスペクト比維持とレターボックスを実装する
  - [x] 11.6 マウス座標の仮想デスクトップ座標変換を実装する
  - [x] 11.7 統合テスト: 描画ループの動作確認

## フェーズ10: VM統合

- [x] 12. VM組み込み関数の実装
  - [x] 12.1 LoadPic組み込み関数を実装する
  - [x] 12.2 CreatePic、DelPic組み込み関数を実装する
  - [x] 12.3 PicWidth、PicHeight組み込み関数を実装する
  - [x] 12.4 MovePic、MoveSPic、TransPic、ReversePic組み込み関数を実装する
  - [x] 12.5 OpenWin、MoveWin、CloseWin、CloseWinAll組み込み関数を実装する
  - [x] 12.6 GetPicNo、CapTitle組み込み関数を実装する
  - [x] 12.7 PutCast、MoveCast、DelCast組み込み関数を実装する
  - [x] 12.8 SetFont、TextWrite、TextColor、BgColor、BackMode組み込み関数を実装する
  - [x] 12.9 DrawLine、DrawRect、FillRect、DrawCircle組み込み関数を実装する
  - [x] 12.10 SetLineSize、SetPaintColor、GetColor組み込み関数を実装する
  - [x] 12.11 WinInfo組み込み関数を実装する

- [x] 13. ゲームループ統合
  - [x] 13.1 `pkg/window/window.go` を更新してGraphicsSystemを統合する
  - [x] 13.2 Update()でVMイベント処理を呼び出す
  - [x] 13.3 Draw()で描画コマンドキューを処理する
  - [x] 13.4 マウスイベントをVMに伝達する
  - [x] 13.5 VMの終了時にEbitengineを終了する
  - [x] 13.6 Ebitengineウィンドウ閉じ時にVMを停止する

- [x] 14. ヘッドレスモード対応
  - [x] 14.1 ヘッドレスモード用のダミーGraphicsSystemを実装する
  - [x] 14.2 描画操作をログに記録する機能を実装する
  - [x] 14.3 VMオプションでヘッドレスモードを切り替え可能にする

## フェーズ11: テストとドキュメント

- [x] 15. 統合テスト
  - [x] 15.1 サンプルスクリプト（ftile400）での動作確認
  - [x] 15.2 サンプルスクリプト（home）での動作確認
  - [x] 15.3 サンプルスクリプト（kuma2）での動作確認
  - [x] 15.4 エラーハンドリングの動作確認
  - [x] 15.5 リソース解放の動作確認

## フェーズ12: デバッグオーバーレイ

- [x] 16. デバッグオーバーレイの実装
  - [x] 16.1 `pkg/graphics/debug.go` を作成し、DebugOverlay構造体を定義する
  - [x] 16.2 DrawWindowID()を実装する（タイトルバーに黄色で `[W1]` 形式）
  - [x] 16.3 DrawPictureID()を実装する（左上に緑色で `P1` 形式、半透明黒背景）
  - [x] 16.4 DrawCastID()を実装する（キャスト位置に黄色で `C1(P2)` 形式、半透明黒背景）
  - [x] 16.5 ログレベルに基づいた表示/非表示の切り替えを実装する
  - [x] 16.6 GraphicsSystem.Draw()にデバッグオーバーレイ描画を統合する
  - [x] 16.7 ユニットテスト: デバッグオーバーレイの表示/非表示切り替え

## フェーズ13: バグ修正

- [x] 17. キャスト透明色の修正
  - [x] 17.1 キャストの透明色（transColor）が反映されない問題を修正する
    - 症状: キャストの背景が白色で塗りつぶされたまま表示される
    - 期待動作: 透明色（例: 0xffffff）と一致するピクセルが透明になる
    - 関連ファイル: `pkg/graphics/shader.go`, `pkg/graphics/graphics.go`
    - 調査ポイント: `drawImageWithColorKey`関数が正しく呼ばれているか、ピクセル処理が正しいか
  - [x] 17.2 サンプル `y_saru` でキャストの透明色が正しく動作することを確認する

- [ ] 18. キャスト表示後の画面更新問題
  - [ ] 18.1 キャストを表示するシーンが終わった後、画面の表示が変わらなくなる問題を調査・修正する
    - 症状: キャストシーン終了後、画面が更新されなくなる
    - 関連ファイル: `pkg/graphics/graphics.go`, `pkg/vm/vm.go`
    - 調査ポイント: ウィンドウやキャストの削除処理、描画ループの状態

- [x] 19. RLE圧縮BMPサポートの実装
  - [x] 19.1 RLE8圧縮BMPのデコード機能を実装する
    - Go標準ライブラリの`image/bmp`はRLE圧縮をサポートしていないため、カスタムデコーダーが必要
    - 関連ファイル: `pkg/graphics/bmp.go`, `pkg/graphics/picture.go`
    - _Requirements: 1.10.1_
  - [x] 19.2 RLE4圧縮BMPのデコード機能を実装する
    - _Requirements: 1.10.1_
  - [x] 19.3 RLE圧縮BMPのユニットテストを作成する
    - samples/robot/のBMPファイルを使用してテスト
    - _Requirements: 1.10.1, 1.10.2_
  - [x] 19.4 samples/robot/ROBOT.TFY でRLE圧縮BMPが正しく読み込まれることを確認する
    - _Requirements: 1.10.1_


---

## ~~フェーズ14: レイヤーシステム再設計 (Layer System Redesign)~~ [廃止]

> **注意**: このフェーズのタスクは完了しましたが、その後フェーズ15のスプライトシステムに置き換えられました。
> 現在の実装では`GetLayerManager()`はnilを返し、すべての描画はスプライトシステム経由で行われます。
> 以下のタスクは歴史的な参照のために残されています。

<details>
<summary>廃止されたタスク（クリックで展開）</summary>

- [x] 20. WindowLayerSetの実装
  - [x] 20.1 WindowLayerSet構造体を定義する
    - WinID、BgColor、Width、Height、Layers、nextZOrder、CompositeBuffer、ダーティフラグを含む
    - _Requirements: 24.1, 24.4_
  
  - [x] 20.2 LayerManagerにwindowLayersマップを追加する
    - map[int]*WindowLayerSet として定義
    - 既存のlayersマップは後方互換性のために残す
    - _Requirements: 24.1, 24.5_
  
  - [x] 20.3 WindowLayerSetのCRUD操作を実装する
    - GetOrCreateWindowLayerSet(winID)
    - GetWindowLayerSet(winID)
    - DeleteWindowLayerSet(winID)
    - _Requirements: 24.2, 24.3_
  
  - [x] 20.4 WindowLayerSetのプロパティテストを作成する
    - **Property 11: レイヤーのWindowID管理**
    - **Property 12: ウィンドウ開閉時のレイヤーセット管理**
    - **Validates: Requirements 24.1, 24.2, 24.3, 24.5**

- [x] 21. PictureLayerの実装
  - [x] 21.1 PictureLayer構造体を定義する
    - BaseLayerを埋め込み
    - image（ウィンドウサイズの透明画像）
    - bakeable（焼き付け可能フラグ）
    - _Requirements: 25.1, 25.5_
  
  - [x] 21.2 PictureLayerのメソッドを実装する
    - NewPictureLayer(id, winWidth, winHeight)
    - Bake(src, destX, destY)
    - IsBakeable()
    - GetLayerType() → LayerTypePicture
    - _Requirements: 25.1, 25.4, 25.5_
  
  - [x] 21.3 PictureLayerの単体テストを作成する
    - 焼き付け動作の確認
    - _Requirements: 25.1, 25.5_

- [x] 22. LayerTypeの追加
  - [x] 22.1 LayerType定数を定義する
    - LayerTypePicture、LayerTypeText、LayerTypeCast
    - _Requirements: 25.4_
  
  - [x] 22.2 Layerインターフェースに GetLayerType() を追加する
    - 既存のレイヤー型に実装を追加
    - _Requirements: 25.4_
  
  - [x] 22.3 レイヤータイプ識別のプロパティテストを作成する
    - **Property 13: レイヤータイプの識別**
    - **Validates: Requirements 25.4**

- [x] 23. チェックポイント - 基本構造の確認
  - すべてのテストが通ることを確認し、問題があればユーザーに質問する

- [x] 24. MovePicの焼き付けロジック実装
  - [x] 24.1 getTopmostLayer()メソッドを実装する
    - WindowLayerSetの最上位レイヤーを取得
    - _Requirements: 26.1_
  
  - [x] 24.2 movePicInternalを修正する
    - 最上位レイヤーのタイプを確認
    - Picture_Layerなら焼き付け、そうでなければ新規作成
    - _Requirements: 26.2, 26.3, 26.4_
  
  - [x] 24.3 createPictureLayerForWindow()を実装する
    - ウィンドウサイズの透明PictureLayerを作成
    - WindowLayerSetに追加
    - _Requirements: 26.5_
  
  - [x] 24.4 焼き付け後のダーティフラグ設定を実装する
    - _Requirements: 26.6_
  
  - [x] 24.5 焼き付けロジックのプロパティテストを作成する
    - **Property 14: MovePicの焼き付けロジック**
    - **Validates: Requirements 26.2, 26.3, 26.4, 26.5, 26.6**

- [x] 25. CastManagerの修正
  - [x] 25.1 createCastLayerをWindowLayerSetに登録するよう修正する
    - windowLayersマップを使用
    - _Requirements: 29.1_
  
  - [x] 25.2 updateCastLayerをWindowLayerSetを使用するよう修正する
    - _Requirements: 27.2_
  
  - [x] 25.3 deleteCastLayerをWindowLayerSetを使用するよう修正する
    - _Requirements: 27.3_
  
  - [x] 25.4 Cast_Layerのプロパティテストを作成する
    - **Property 16: Cast_Layerのスプライト動作**
    - **Validates: Requirements 27.1, 27.2, 27.3, 27.6**

- [x] 26. TextRendererの修正
  - [x] 26.1 TextWriteをWindowLayerSetに登録するよう修正する
    - ピクチャーIDからウィンドウIDを逆引き
    - _Requirements: 29.3_
  
  - [x] 26.2 Text_Layerの新規作成を確認する
    - 既存の実装が正しいことを確認
    - _Requirements: 28.1, 28.2_
  
  - [x] 26.3 Text_Layerのプロパティテストを作成する
    - **Property 15: Text_Layerの新規作成**
    - **Validates: Requirements 25.6, 28.1, 28.2**

- [x] 27. チェックポイント - レイヤー登録の確認
  - すべてのテストが通ることを確認し、問題があればユーザーに質問する

- [x] 28. drawLayersForWindowの修正
  - [x] 28.1 drawLayersForWindowをWindowIDで検索するよう修正する
    - win.PicID → win.ID に変更
    - _Requirements: 29.4_
  
  - [x] 28.2 背景色の描画を追加する
    - WindowLayerSetのBgColorを使用
    - _Requirements: 24.4_
  
  - [x] 28.3 Z順序でのソートを確認する
    - GetAllLayersSorted()が正しく動作することを確認
    - _Requirements: 23.5_
  
  - [x] 28.4 Z順序のプロパティテストを作成する
    - **Property 17: Z順序の管理**
    - **Validates: Requirements 23.1, 23.2, 23.5**

- [x] 29. ピクチャーID逆引きの実装
  - [x] 29.1 WindowManagerにGetWinByPicIDを確認する
    - 既存の実装が正しいことを確認
    - _Requirements: 29.5_
  
  - [x] 29.2 LayerManagerにpicToWinMappingを追加する（必要な場合）
    - ピクチャーIDからウィンドウIDへのマッピング
    - _Requirements: 29.5_
    - **結論: 不要** - WindowManager.GetWinByPicID()が既に実装済みで、要件29.5を満たしている。LayerManagerに重複したマッピングを持つ必要はない。
  
  - [x] 29.3 レイヤー登録のプロパティテストを作成する
    - **Property 18: レイヤーのウィンドウ登録**
    - **Validates: Requirements 29.1, 29.2, 29.3, 29.5**

- [x] 30. ウィンドウ操作との統合
  - [x] 30.1 OpenWinでWindowLayerSetを作成するよう修正する
    - _Requirements: 24.2_
  
  - [x] 30.2 CloseWinでWindowLayerSetを削除するよう修正する
    - _Requirements: 24.3_
  
  - [x] 30.3 CloseWinAllでWindowLayerSetをすべて削除するよう修正する
    - _Requirements: 24.3_

- [x] 31. チェックポイント - 統合の確認
  - すべてのテストが通ることを確認し、問題があればユーザーに質問する

- [x] 32. ダーティフラグとキャッシュの最適化
  - [x] 32.1 WindowLayerSetのダーティ領域追跡を実装する
    - _Requirements: 31.1, 31.2_
  
  - [x] 32.2 合成処理完了後のダーティフラグクリアを実装する
    - _Requirements: 31.5_
  
  - [x] 32.3 ダーティフラグのプロパティテストを作成する
    - **Property 19: ダーティフラグの動作**
    - **Validates: Requirements 31.1, 31.2**

- [x] 33. エラーハンドリングの実装
  - [x] 33.1 存在しないウィンドウIDのエラー処理を実装する
    - _Requirements: 32.1_
  
  - [x] 33.2 存在しないレイヤーIDのエラー処理を実装する
    - _Requirements: 32.2_
  
  - [x] 33.3 レイヤー作成失敗のエラー処理を実装する
    - _Requirements: 32.3_
  
  - [x] 33.4 エラーハンドリングのプロパティテストを作成する
    - **Property 20: エラーハンドリング**
    - **Validates: Requirements 32.1, 32.2, 32.3, 32.4**

- [x] 34. 後方互換性の確認
  - [x] 34.1 既存のCastManager APIが動作することを確認する
    - _Requirements: 30.1_
  
  - [x] 34.2 既存のTextRenderer APIが動作することを確認する
    - _Requirements: 30.2_
  
  - [x] 34.3 既存のMovePic APIが動作することを確認する
    - _Requirements: 30.3_
  
  - [x] 34.4 既存のテストがすべて通ることを確認する
    - _Requirements: 30.4_

- [x] 35. 最終チェックポイント
  - すべてのテストが通ることを確認し、問題があればユーザーに質問する

## 注意事項（レイヤーシステム再設計）

- 各プロパティテストは最低100回の反復を実行するよう設定してください
- Go言語のプロパティベーステストには `testing/quick` または `github.com/leanovate/gopter` を使用してください
- 既存のテストが壊れないよう、後方互換性に注意してください
- go testコマンドには `-timeout` オプションを使用してください（無限ループ対策）

</details>

## フェーズ15: スプライトシステム (Sprite System)

- [x] 36. 基本スプライト構造体の実装
  - [x] 36.1 Sprite構造体を定義する（ID、Image、X、Y、ZOrder、Visible、Alpha、Parent、Dirty）
    - _Requirements: 33.1〜33.8_
  - [x] 36.2 Spriteのゲッター・セッターを実装する
    - _Requirements: 33.1〜33.8_
  - [x] 36.3 Spriteの単体テストを作成する

- [x] 37. 親子関係の実装
  - [x] 37.1 AbsolutePosition()を実装する
    - _Requirements: 34.1_
  - [x] 37.2 EffectiveAlpha()を実装する
    - _Requirements: 34.2_
  - [x] 37.3 IsEffectivelyVisible()を実装する
    - _Requirements: 34.3_
  - [x] 37.4 親子関係の単体テストを作成する

- [x] 38. SpriteManagerの実装
  - [x] 38.1 SpriteManager構造体を定義する
    - _Requirements: 35.1〜35.6_
  - [x] 38.2 CreateSprite()を実装する
    - _Requirements: 35.1_
  - [x] 38.3 CreateSpriteWithSize()を実装する
    - _Requirements: 35.2_
  - [x] 38.4 GetSprite()を実装する
    - _Requirements: 35.3_
  - [x] 38.5 RemoveSprite()を実装する
    - _Requirements: 35.4_
  - [x] 38.6 Clear()とCount()を実装する
    - _Requirements: 35.5, 35.6_
  - [x] 38.7 SpriteManagerの単体テストを作成する

- [x] 39. Z順序描画の実装
  - [x] 39.1 sortSprites()を実装する
    - _Requirements: 36.1, 42.1_
  - [x] 39.2 Draw()を実装する
    - _Requirements: 36.1〜36.5_
  - [x] 39.3 MarkNeedSort()を実装する
    - _Requirements: 42.2_
  - [x] 39.4 描画の単体テストを作成する

- [x] 40. テキストスプライトヘルパーの実装
  - [x] 40.1 CreateTextSpriteImage()を実装する（差分抽出方式）
    - _Requirements: 37.1, 37.2_
  - [x] 40.2 テキストスプライトの単体テストを作成する

- [x] 41. プロパティベーステストの作成
  - [x] 41.1 Property 21: スプライトID管理のテスト
    - **Validates: Requirements 33.1, 35.1, 35.3**
  - [x] 41.2 Property 22: 親子関係の位置計算のテスト
    - **Validates: Requirements 34.1**
  - [x] 41.3 Property 23: 親子関係の透明度計算のテスト
    - **Validates: Requirements 34.2**
  - [x] 41.4 Property 24: 親子関係の可視性のテスト
    - **Validates: Requirements 34.3, 36.3**
  - [x] 41.5 Property 25: Z順序による描画順のテスト
    - **Validates: Requirements 36.1**
  - [x] 41.6 Property 27: テキスト差分抽出のテスト
    - **Validates: Requirements 37.1, 37.2**

- [x] 42. チェックポイント - スプライト基本機能の確認
  - すべてのテストが通ることを確認し、問題があればユーザーに質問する

- [x] 43. GraphicsSystemへのSpriteManager統合
  - [x] 43.1 GraphicsSystem構造体にSpriteManagerフィールドを追加する
  - [x] 43.2 NewGraphicsSystem()でSpriteManagerを初期化する
  - [x] 43.3 GetSpriteManager()メソッドを追加する
  - [x] 43.4 統合の単体テストを作成する

- [x] 44. ウインドウのスプライト化
  - [x] 44.1 WindowSpriteラッパー構造体を作成する（Window + Sprite）
  - [x] 44.2 OpenWin()でWindowSpriteを作成するように変更する
  - [x] 44.3 drawWindowDecoration()をスプライトベースに変更する
  - [x] 44.4 ウインドウスプライトの単体テストを作成する
  - [x] 44.5 サンプルタイトル（kuma2）でウインドウ表示を確認する

- [x] 45. ピクチャ描画（MovePic）のスプライト化
  - [x] 45.1 DrawingEntryをSpriteベースに変換するアダプタを作成する
  - [x] 45.2 MovePic()でスプライトを作成するように変更する
  - [x] 45.3 drawDrawingEntryOnScreen()をスプライトベースに変更する
  - [x] 45.4 ピクチャ描画の単体テストを作成する
  - [x] 45.5 サンプルタイトル（kuma2）でピクチャ描画を確認する

- [x] 46. キャスト描画のスプライト化
  - [x] 46.1 CastSpriteラッパー構造体を作成する（Cast + Sprite）
  - [x] 46.2 PutCast()でCastSpriteを作成するように変更する
  - [x] 46.3 MoveCast()でCastSpriteを更新するように変更する
  - [x] 46.4 DelCast()でCastSpriteを削除するように変更する
  - [x] 46.5 drawCastLayerOnScreen()をスプライトベースに変更する
  - [x] 46.6 透明色処理をスプライト描画に統合する
  - [x] 46.7 キャスト描画の単体テストを作成する
  - [x] 46.8 サンプルタイトル（home）でキャスト描画を確認する

- [x] 47. テキスト描画のスプライト化
  - [x] 47.1 TextSpriteラッパー構造体を作成する（TextLayerEntry + Sprite）
  - [x] 47.2 TextWrite()でTextSpriteを作成するように変更する
  - [x] 47.3 drawTextLayerOnScreen()をスプライトベースに変更する
  - [x] 47.4 テキスト描画の単体テストを作成する
  - [x] 47.5 サンプルタイトル（kuma2）でテキスト描画を確認する

- [ ] 48. 図形描画のスプライト化
  - [x] 48.1 ShapeSpriteラッパー構造体を作成する（図形情報 + Sprite）
  - [x] 48.2 DrawLine/DrawRect/DrawCircle/FillRectでShapeSpriteを作成するように変更する
  - [x] 48.3 図形描画の単体テストを作成する
  - [ ] 48.4 サンプルタイトル（ftile400）で図形描画を確認する

- [ ] 49. 描画システムの統合
  - [x] 49.1 GraphicsSystem.Draw()をSpriteManager.Draw()ベースに変更する
  - [x] 49.2 ウインドウ内のスプライトをウインドウの子スプライトとして管理する
  - [x] 49.3 Z順序の統一（ウインドウ間、ウインドウ内）を実装する
  - [x] 49.4 統合テストを作成する
  - [ ] 49.5 サンプルタイトル（kuma2, home）で全体動作を確認する

- [-] 50. 旧レイヤーシステムの削除
  - [x] 50.1 LayerManagerの使用箇所を確認し、スプライトシステムに置き換える
  - [x] 50.2 WindowLayerSetの使用箇所を確認し、スプライトシステムに置き換える
  - [x] 50.3 PictureLayerSetの使用箇所を確認し、スプライトシステムに置き換える
  - [x] 50.4 CastLayer, DrawingEntry, TextLayerEntryの使用箇所を確認する
  - [x] 50.5 不要になったレイヤー関連ファイルを削除する
  - [x] 50.6 不要になったテストファイルを削除する
  - [x] 50.7 全テストがパスすることを確認する

- [ ] 51. スプライトシステム最終動作確認
  - [x] 51.1 サンプルタイトル（kuma2）で動作確認
  - [ ] 51.2 サンプルタイトル（home）で動作確認
  - [ ] 51.3 サンプルタイトル（ftile400）で動作確認
  - [ ] 51.4 既存テストがすべてパスすることを確認

- [-] 52. スプライトシステム完全統合
  - [x] 52.1 bakeToPictureLayerからピクチャーへの直接描画を削除する
    - _Requirements: 43.1, 43.3_
    - _Note: 完了。PictureSpriteのみを作成するように変更済み_
  - [x] 52.2 TextSpriteを描画する際に下のスプライトの画像を参照してブレンドする
    - _Requirements: 43.2, 43.4_
    - _Note: drawTextSpriteWithBackgroundを実装。ReadPixelsでscreenから背景を読み取り、テキストを再描画する_
  - [ ] 52.3 drawWindowDecorationからピクチャー画像の直接描画を削除する
    - _Requirements: 43.2_
    - _Note: 試行したが問題が発生したため元に戻した_
  - [ ] 52.4 Draw()をスプライトシステムのみで描画するように変更する
    - _Requirements: 43.2, 43.4_
  - [ ] 52.5 統合テストを作成する
  - [ ] 52.6 サンプルタイトル（kuma2）で動作確認

- [ ] 53. ピクチャースプライトの融合機能
  - [ ] 53.1 PictureSpriteManagerにFindMergeableSprite()を追加する
    - _Requirements: 44.2_
  - [ ] 53.2 PictureSpriteにMergeImage()を追加する
    - _Requirements: 44.3_
  - [ ] 53.3 PictureSpriteManagerにMergeOrCreatePictureSprite()を追加する
    - _Requirements: 44.1_
  - [ ] 53.4 MovePicで融合機能を使用するように変更する
    - _Requirements: 44.1, 44.5_
  - [ ] 53.5 融合時の領域拡張を実装する
    - _Requirements: 44.4_
  - [ ] 53.6 融合機能の単体テストを作成する
  - [ ] 53.7 サンプルタイトル（y_saru）で動作確認

- [ ] 54. スプライトシステム最終検証
  - [ ] 54.1 すべてのサンプルタイトルで動作確認
  - [ ] 54.2 既存テストがすべてパスすることを確認
  - [ ] 54.3 パフォーマンス測定（スプライト数、FPS）

## 注意事項（スプライトシステム）

- 各プロパティテストは最低100回の反復を実行するよう設定してください
- go testコマンドには `-timeout` オプションを使用してください
- 既存のテストが壊れないよう注意してください
- 各タスク完了後にサンプルタイトルで動作確認を行い、問題があれば早期に発見してください
- 旧システムと新システムを並行稼働させながら段階的に移行してください
