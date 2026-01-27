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
