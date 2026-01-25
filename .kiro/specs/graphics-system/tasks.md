# 実装タスク: 描画システム (Graphics System)

## フェーズ1: 基盤

- [ ] 1. GraphicsSystem構造体の実装
  - [ ] 1.1 `pkg/graphics/graphics.go` を作成し、GraphicsSystem構造体を定義する
  - [ ] 1.2 NewGraphicsSystem()コンストラクタを実装する
  - [ ] 1.3 Update()メソッドを実装する（コマンドキュー処理用）
  - [ ] 1.4 Draw()メソッドを実装する（描画処理用）
  - [ ] 1.5 Shutdown()メソッドを実装する

- [ ] 2. CommandQueueの実装
  - [ ] 2.1 `pkg/graphics/queue.go` を作成し、CommandQueue構造体を定義する
  - [ ] 2.2 Push()メソッドを実装する（スレッドセーフ）
  - [ ] 2.3 PopAll()メソッドを実装する
  - [ ] 2.4 CommandType定数を定義する
  - [ ] 2.5 プロパティテスト: コマンド実行順序（FIFO）の検証 **Validates: 要件 7.4**
  - [ ] 2.6 プロパティテスト: スレッドセーフ性の検証 **Validates: 要件 7.1**

- [ ] 3. 色変換ユーティリティの実装
  - [ ] 3.1 `pkg/graphics/color.go` を作成する
  - [ ] 3.2 ColorFromInt()を実装する（0xRRGGBB → color.Color）
  - [ ] 3.3 ColorToInt()を実装する（color.Color → 0xRRGGBB）
  - [ ] 3.4 ユニットテスト: 色変換の往復テスト

## フェーズ2: ピクチャーシステム

- [ ] 4. PictureManagerの実装
  - [ ] 4.1 `pkg/graphics/picture.go` を作成し、Picture構造体を定義する
  - [ ] 4.2 PictureManager構造体を定義する
  - [ ] 4.3 NewPictureManager()コンストラクタを実装する
  - [ ] 4.4 LoadPic()を実装する（BMP/PNG読み込み、大文字小文字非依存検索）
  - [ ] 4.5 CreatePic()を実装する（空のピクチャー生成）
  - [ ] 4.6 CreatePicFrom()を実装する（既存ピクチャーからコピー生成）
  - [ ] 4.7 DelPic()を実装する（ピクチャー削除）
  - [ ] 4.8 GetPic()を実装する
  - [ ] 4.9 PicWidth()、PicHeight()を実装する
  - [ ] 4.10 プロパティテスト: ピクチャーIDの一意性 **Validates: 要件 1.2**
  - [ ] 4.11 プロパティテスト: ピクチャーサイズの正確性 **Validates: 要件 1.7, 1.8**
  - [ ] 4.12 プロパティテスト: 削除後のアクセスエラー **Validates: 要件 1.9**
  - [ ] 4.13 プロパティテスト: リソース制限（最大256） **Validates: 要件 9.5**

## フェーズ3: ウィンドウシステム

- [ ] 5. WindowManagerの実装
  - [ ] 5.1 `pkg/graphics/window.go` を作成し、Window構造体を定義する
  - [ ] 5.2 WindowManager構造体を定義する
  - [ ] 5.3 NewWindowManager()コンストラクタを実装する
  - [ ] 5.4 OpenWin()を実装する（ウィンドウ作成）
  - [ ] 5.5 MoveWin()を実装する（ウィンドウ移動・変更）
  - [ ] 5.6 CloseWin()を実装する（ウィンドウ削除）
  - [ ] 5.7 CloseWinAll()を実装する
  - [ ] 5.8 GetWin()を実装する
  - [ ] 5.9 GetWindowsOrdered()を実装する（Z順序でソート）
  - [ ] 5.10 プロパティテスト: ウィンドウZ順序 **Validates: 要件 3.11**
  - [ ] 5.11 プロパティテスト: リソース制限（最大64） **Validates: 要件 9.6**

## フェーズ4: キャストシステム

- [ ] 6. CastManagerの実装
  - [ ] 6.1 `pkg/graphics/cast.go` を作成し、Cast構造体を定義する
  - [ ] 6.2 CastManager構造体を定義する
  - [ ] 6.3 NewCastManager()コンストラクタを実装する
  - [ ] 6.4 PutCast()を実装する（キャスト配置）
  - [ ] 6.5 MoveCast()を実装する（キャスト移動）
  - [ ] 6.6 DelCast()を実装する（キャスト削除）
  - [ ] 6.7 GetCast()を実装する
  - [ ] 6.8 GetCastsByWindow()を実装する
  - [ ] 6.9 DeleteCastsByWindow()を実装する
  - [ ] 6.10 プロパティテスト: キャストIDの一意性 **Validates: 要件 4.2**
  - [ ] 6.11 プロパティテスト: キャスト位置の更新 **Validates: 要件 4.3**
  - [ ] 6.12 プロパティテスト: ウィンドウ削除時のキャスト削除 **Validates: 要件 9.2**
  - [ ] 6.13 プロパティテスト: リソース制限（最大1024） **Validates: 要件 9.7**

## フェーズ5: ピクチャー転送

- [ ] 7. ピクチャー転送機能の実装
  - [ ] 7.1 `pkg/graphics/transfer.go` を作成する
  - [ ] 7.2 MovePic()を実装する（mode=0: 通常コピー）
  - [ ] 7.3 MovePic()を実装する（mode=1: 透明色除外）
  - [ ] 7.4 TransPic()を実装する（指定透明色除外）
  - [ ] 7.5 ReversePic()を実装する（左右反転）
  - [ ] 7.6 MoveSPic()を実装する（拡大縮小転送）
  - [ ] 7.7 クリッピング処理を実装する
  - [ ] 7.8 ユニットテスト: 各転送モードの動作確認

## フェーズ6: テキストシステム

- [ ] 8. TextRendererの実装
  - [ ] 8.1 `pkg/graphics/text.go` を作成し、FontSettings構造体を定義する
  - [ ] 8.2 TextSettings構造体を定義する
  - [ ] 8.3 TextRenderer構造体を定義する
  - [ ] 8.4 NewTextRenderer()コンストラクタを実装する
  - [ ] 8.5 SetFont()を実装する（フォント設定）
  - [ ] 8.6 フォントフォールバック機能を実装する（MSゴシック→Hiragino等）
  - [ ] 8.7 SetTextColor()、SetBgColor()、SetBackMode()を実装する
  - [ ] 8.8 TextWrite()を実装する（アンチエイリアス無効）
  - [ ] 8.9 埋め込みフォントを追加する（NotoSansJP等）
  - [ ] 8.10 ユニットテスト: フォントフォールバックの動作確認

## フェーズ7: 描画プリミティブ

- [ ] 9. 描画プリミティブの実装
  - [ ] 9.1 `pkg/graphics/primitives.go` を作成する
  - [ ] 9.2 DrawLine()を実装する
  - [ ] 9.3 DrawRect()を実装する
  - [ ] 9.4 FillRect()を実装する
  - [ ] 9.5 DrawCircle()を実装する
  - [ ] 9.6 SetLineSize()を実装する
  - [ ] 9.7 SetPaintColor()を実装する
  - [ ] 9.8 GetColor()を実装する
  - [ ] 9.9 ユニットテスト: 各描画プリミティブの動作確認

## フェーズ8: シーンチェンジ

- [ ] 10. シーンチェンジの実装
  - [ ] 10.1 `pkg/graphics/scene_change.go` を作成する
  - [ ] 10.2 SceneChange構造体を定義する
  - [ ] 10.3 SceneChangeMode定数を定義する（mode 2-9）
  - [ ] 10.4 ワイプエフェクトを実装する（mode 2-7）
  - [ ] 10.5 ランダムブロックエフェクトを実装する（mode 8）
  - [ ] 10.6 フェードエフェクトを実装する（mode 9）
  - [ ] 10.7 MovePic()にシーンチェンジモードを統合する
  - [ ] 10.8 ユニットテスト: 各シーンチェンジモードの動作確認

## フェーズ9: 描画ループ統合

- [ ] 11. 描画ループの統合
  - [ ] 11.1 GraphicsSystem.Draw()でウィンドウを描画する
  - [ ] 11.2 GraphicsSystem.Draw()でキャストを描画する（透明色除外）
  - [ ] 11.3 Z順序に基づいた描画順序を実装する
  - [ ] 11.4 仮想デスクトップのスケーリングを実装する
  - [ ] 11.5 アスペクト比維持とレターボックスを実装する
  - [ ] 11.6 マウス座標の仮想デスクトップ座標変換を実装する
  - [ ] 11.7 統合テスト: 描画ループの動作確認

## フェーズ10: VM統合

- [ ] 12. VM組み込み関数の実装
  - [ ] 12.1 LoadPic組み込み関数を実装する
  - [ ] 12.2 CreatePic、DelPic組み込み関数を実装する
  - [ ] 12.3 PicWidth、PicHeight組み込み関数を実装する
  - [ ] 12.4 MovePic、MoveSPic、TransPic、ReversePic組み込み関数を実装する
  - [ ] 12.5 OpenWin、MoveWin、CloseWin、CloseWinAll組み込み関数を実装する
  - [ ] 12.6 GetPicNo、CapTitle組み込み関数を実装する
  - [ ] 12.7 PutCast、MoveCast、DelCast組み込み関数を実装する
  - [ ] 12.8 SetFont、TextWrite、TextColor、BgColor、BackMode組み込み関数を実装する
  - [ ] 12.9 DrawLine、DrawRect、FillRect、DrawCircle組み込み関数を実装する
  - [ ] 12.10 SetLineSize、SetPaintColor、GetColor組み込み関数を実装する
  - [ ] 12.11 WinInfo組み込み関数を実装する

- [ ] 13. ゲームループ統合
  - [ ] 13.1 `pkg/window/window.go` を更新してGraphicsSystemを統合する
  - [ ] 13.2 Update()でVMイベント処理を呼び出す
  - [ ] 13.3 Draw()で描画コマンドキューを処理する
  - [ ] 13.4 マウスイベントをVMに伝達する
  - [ ] 13.5 VMの終了時にEbitengineを終了する
  - [ ] 13.6 Ebitengineウィンドウ閉じ時にVMを停止する

- [ ] 14. ヘッドレスモード対応
  - [ ] 14.1 ヘッドレスモード用のダミーGraphicsSystemを実装する
  - [ ] 14.2 描画操作をログに記録する機能を実装する
  - [ ] 14.3 VMオプションでヘッドレスモードを切り替え可能にする

## フェーズ11: テストとドキュメント

- [ ] 15. 統合テスト
  - [ ] 15.1 サンプルスクリプト（ftile400）での動作確認
  - [ ] 15.2 サンプルスクリプト（home）での動作確認
  - [ ] 15.3 サンプルスクリプト（kuma2）での動作確認
  - [ ] 15.4 エラーハンドリングの動作確認
  - [ ] 15.5 リソース解放の動作確認

## フェーズ12: デバッグオーバーレイ

- [ ] 16. デバッグオーバーレイの実装
  - [ ] 16.1 `pkg/graphics/debug.go` を作成し、DebugOverlay構造体を定義する
  - [ ] 16.2 DrawWindowID()を実装する（タイトルバーに黄色で `[W1]` 形式）
  - [ ] 16.3 DrawPictureID()を実装する（左上に緑色で `P1` 形式、半透明黒背景）
  - [ ] 16.4 DrawCastID()を実装する（キャスト位置に黄色で `C1(P2)` 形式、半透明黒背景）
  - [ ] 16.5 ログレベルに基づいた表示/非表示の切り替えを実装する
  - [ ] 16.6 GraphicsSystem.Draw()にデバッグオーバーレイ描画を統合する
  - [ ] 16.7 ユニットテスト: デバッグオーバーレイの表示/非表示切り替え
