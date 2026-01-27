# 要件定義書: 実行エンジン (Execution Engine)

## はじめに

このドキュメントは、Toffyスクリプト言語の実行エンジンの要件を定義します。コンパイラは既に実装済みで、OpCodeを生成できます。実行エンジンは、生成されたOpCodeを実行し、イベントドリブンなスクリプトの動作を実現します。

特に、MIDIのテンポに同期したタイミング制御が重要な機能であるため、サウンドシステムとイベント機構を最優先で実装します。描画系機能は別スペックで対応します。

また、ROBOTサンプル（samples/robot/ROBOT.TFY）の実行時に発生するエラーを修正するため、不足している組み込み関数（StrPrint、CreatePic 3引数パターン、CapTitle 1引数パターン）の実装も含みます。

## 用語集

- **StrPrint**: printf形式のフォーマット指定子を使用して文字列を生成する組み込み関数
- **Format_Specifier**: フォーマット文字列内の変換指定子（%ld, %lx, %s, %03d等）
- **CreatePic**: ピクチャーを生成する組み込み関数。複数の呼び出しパターンをサポート
- **CapTitle**: ウィンドウのキャプション（タイトル）を設定する組み込み関数
- **Virtual_Window**: FILLYスクリプトで管理される仮想ウィンドウ
- **Picture_ID**: ピクチャーを識別する整数値
- **VM（仮想マシン）**: OpCodeを実行するランタイム環境
- **OpCode**: コンパイラが生成した実行可能な命令列
- **Event_Handler**: mes()ブロックで定義されるイベント駆動の処理
- **Event_Queue**: イベントを格納し順次処理するキュー
- **Event_Dispatcher**: イベントを適切なハンドラに配信する機構
- **Step_Block**: step()構文で定義される時間制御されたコマンドシーケンス
- **Step_Counter**: ステップ実行の現在位置を追跡するカウンタ
- **MIDI_Tick**: MIDIの最小時間単位（通常は1/480拍）
- **Timer_Event**: 定期的に発生する時間イベント
- **Callback**: イベント発生時に呼び出される関数
- **Scope**: 変数の有効範囲
- **Global_Scope**: プログラム全体で有効な変数スコープ
- **Local_Scope**: 関数内でのみ有効な変数スコープ

## 要件

### 要件1: イベントシステムの基盤

**ユーザーストーリー:** 開発者として、イベントドリブンなスクリプトを実行したい。そうすることで、タイマーやMIDI再生に同期した処理を実現できる。

#### 受け入れ基準

1.1. THE System SHALL イベントを時系列順に格納するイベントキューを提供する
1.2. WHEN イベントがキューに追加されたとき、THE System SHALL タイムスタンプを割り当てる
1.3. WHEN イベントループがイベントを処理するとき、THE System SHALL 時系列順にイベントをディスパッチする
1.4. WHEN イベントがディスパッチされたとき、THE System SHALL そのイベントタイプに登録されたすべてのハンドラを呼び出す
1.5. WHEN 同じイベントタイプに複数のハンドラが登録されているとき、THE System SHALL 登録順に実行する
1.6. WHEN ハンドラが実行中のとき、THE System SHALL イベント固有のパラメータ（MesP1、MesP2、MesP3）へのアクセスを提供する
1.7. THE System SHALL 次のイベントタイプをサポートする：TIME、MIDI_TIME、MIDI_END、LBDOWN、RBDOWN、RBDBLCLK

### 要件2: mes()構文のサポート

**ユーザーストーリー:** 開発者として、mes()構文でイベントハンドラを登録したい。そうすることで、特定のイベント発生時に処理を実行できる。

#### 受け入れ基準

2.1. WHEN OpRegisterEventHandler OpCodeが実行されたとき、THE System SHALL 指定されたイベントタイプのハンドラを登録する
2.2. WHEN mes(TIME)ハンドラが登録されたとき、THE System SHALL タイマーティックごとにそれを呼び出す
2.3. WHEN mes(MIDI_TIME)ハンドラが登録されたとき、THE System SHALL MIDIティックごとにそれを呼び出す
2.4. WHEN mes(MIDI_END)ハンドラが登録されたとき、THE System SHALL MIDI再生完了時にそれを呼び出す
2.5. WHEN mes(LBDOWN)ハンドラが登録されたとき、THE System SHALL マウス左ボタン押下時にそれを呼び出す
2.6. WHEN mes(RBDOWN)ハンドラが登録されたとき、THE System SHALL マウス右ボタン押下時にそれを呼び出す
2.7. WHEN mes(RBDBLCLK)ハンドラが登録されたとき、THE System SHALL マウス右ボタンダブルクリック時にそれを呼び出す
2.8. WHEN ハンドラが別のハンドラ内で登録されたとき、THE System SHALL ネストされたハンドラ登録をサポートする
2.9. WHEN del_meが呼ばれたとき、THE System SHALL 現在実行中のハンドラを削除する
2.10. WHEN del_allが呼ばれたとき、THE System SHALL すべての登録済みハンドラを削除する
2.11. WHEN del_usが呼ばれたとき、THE System SHALL 現在実行中のハンドラを削除する（del_meと同じ）

### 要件3: タイマーシステム

**ユーザーストーリー:** 開発者として、定期的なタイマーイベントを受け取りたい。そうすることで、時間経過に応じた処理を実行できる。

#### 受け入れ基準

3.1. THE System SHALL 定期的にTIMEイベントを生成する
3.2. WHEN タイマー間隔が設定されたとき、THE System SHALL その間隔をTIMEイベント生成に使用する
3.3. THE System SHALL デフォルトのタイマー間隔として50ミリ秒を提供する
3.4. WHEN TIMEイベントが生成されたとき、THE System SHALL それをイベントキューに追加する
3.5. WHEN 複数のTIMEハンドラが登録されているとき、THE System SHALL 各TIMEイベントに対してすべてを呼び出す
3.6. THE System SHALL ハンドラの実行に時間がかかっても正確なタイミングを維持する

### 要件4: MIDI再生とMIDI_TIMEイベント

**ユーザーストーリー:** 開発者として、MIDIファイルを再生し、そのテンポに同期したイベントを受け取りたい。そうすることで、音楽に合わせたアニメーションを実現できる。

#### 受け入れ基準

4.1. WHEN PlayMIDI(filename)が呼ばれたとき、THE System SHALL 指定されたMIDIファイルの再生を開始する
4.2. WHEN MIDI再生が開始されたとき、THE System SHALL MIDIファイルからテンポ情報を抽出する
4.3. WHEN MIDIが再生中のとき、THE System SHALL MIDIテンポに同期したMIDI_TIMEイベントを生成する
4.4. WHEN MIDIテンポが120 BPMで解像度が480 ticks per beatのとき、THE System SHALL 1.04ミリ秒ごとにMIDI_TIMEイベントを生成する
4.5. WHEN MIDI再生が完了したとき、THE System SHALL MIDI_ENDイベントを生成する
4.6. WHEN 別のMIDIが再生中にPlayMIDIが呼ばれたとき、THE System SHALL 前のMIDIを停止して新しいMIDIを開始する
4.7. THE System SHALL Standard MIDI File (SMF)フォーマットをサポートする
4.8. THE System SHALL ソフトウェアシンセサイザーを使用してMIDIオーディオをレンダリングする
4.9. WHEN SoundFontファイルが提供されたとき、THE System SHALL それをMIDI合成に使用する
4.10. WHEN SoundFontが提供されないとき、THE System SHALL エラーを報告する

### 要件5: WAV再生

**ユーザーストーリー:** 開発者として、WAVファイルを再生したい。そうすることで、効果音を鳴らすことができる。

#### 受け入れ基準

5.1. WHEN PlayWAVE(filename)が呼ばれたとき、THE System SHALL 指定されたWAVファイルの再生を開始する
5.2. WHEN 複数のPlayWAVE呼び出しが行われたとき、THE System SHALL すべてのWAVファイルを同時に再生する
5.3. THE System SHALL 標準的なWAVファイルフォーマット（PCM、8ビット、16ビット）をサポートする
5.4. WHEN WAVファイルが見つからないとき、THE System SHALL エラーをログに記録して実行を継続する
5.5. WHEN WAVファイルが破損しているとき、THE System SHALL エラーをログに記録して実行を継続する
5.6. THE System SHALL 複数のWAVストリームを単一のオーディオ出力にミックスする

### 要件6: step()構文のサポート

**ユーザーストーリー:** 開発者として、step()構文でステップ実行を制御したい。そうすることで、イベントごとに次のステップに進む処理を実現できる。

#### 受け入れ基準

6.1. WHEN OpSetStep OpCodeが実行されたとき、THE System SHALL 指定されたカウントでステップカウンタを初期化する（step(n)のnはカンマ1つあたりに待機するイベント数を表す）
6.2. WHEN OpWait OpCodeが実行されたとき、THE System SHALL 次のイベントが発生するまで実行を一時停止する
6.3. WHEN ステップ実行中にイベントが発生したとき、THE System SHALL 次のステップに進む
6.4. WHEN ステップに複数のコマンドが含まれるとき、THE System SHALL 待機する前にすべてのコマンドを実行する
6.5. WHEN ステップが空（カンマのみ）のとき、THE System SHALL コマンドを実行せずに次のイベントを待つ
6.6. WHEN 連続するカンマが現れたとき、THE System SHALL 複数のイベントを待つ（カンマ数 × step(n)のn回のイベント）
6.7. WHEN end_stepが呼ばれたとき、THE System SHALL ステップブロックの実行を終了する
6.8. WHEN すべてのステップが完了したとき、THE System SHALL 自動的にステップブロックを終了する
6.9. WHEN step()がブロックなしで使用されたとき、THE System SHALL 指定された数のイベントを待つ
6.10. WHEN step(n)がn=0で呼ばれたとき、THE System SHALL 待機せずに即座に実行する
6.11. WHEN mes(TIME)内でstep(n)が使用されたとき、THE System SHALL カンマ1つにつきn回のTIMEイベント（n × 50ms）を待機する
6.12. WHEN mes(MIDI_TIME)内でstep(n)が使用されたとき、THE System SHALL カンマ1つにつきn回のMIDI_TIMEイベントを待機する

### 要件7: マウスイベント

**ユーザーストーリー:** 開発者として、マウスイベントを受け取りたい。そうすることで、ユーザーのマウス操作に応じた処理を実行できる。

#### 受け入れ基準

7.1. WHEN マウス左ボタンが押されたとき、THE System SHALL LBDOWNイベントを生成する
7.2. WHEN マウス右ボタンが押されたとき、THE System SHALL RBDOWNイベントを生成する
7.3. WHEN マウス右ボタンがダブルクリックされたとき、THE System SHALL RBDBLCLKイベントを生成する
7.4. WHEN マウスイベントが生成されたとき、THE System SHALL MesP1にウィンドウIDを設定する
7.5. WHEN マウスイベントが生成されたとき、THE System SHALL MesP2にX座標を設定する
7.6. WHEN マウスイベントが生成されたとき、THE System SHALL MesP3にY座標を設定する
7.7. THE System SHALL ウィンドウ相対のマウス座標を提供する
7.8. WHEN マウスカーソルの下にウィンドウがないとき、THE System SHALL MesP1を-1に設定する

### 要件8: OpCode実行

**ユーザーストーリー:** 開発者として、コンパイルされたOpCodeを実行したい。そうすることで、スクリプトの動作を実現できる。

#### 受け入れ基準

8.1. WHEN VMがOpCodeシーケンスを受け取ったとき、THE System SHALL 各OpCodeを順番に実行する
8.2. WHEN OpAssignが実行されたとき、THE System SHALL 指定された変数に値を代入する
8.3. WHEN OpArrayAssignが実行されたとき、THE System SHALL 指定された配列要素に値を代入する
8.4. WHEN OpCallが実行されたとき、THE System SHALL 指定された関数を引数とともに呼び出す
8.5. WHEN OpIfが実行されたとき、THE System SHALL 条件を評価して適切な分岐を実行する
8.6. WHEN OpForが実行されたとき、THE System SHALL 初期化、条件、インクリメントを伴うループを実行する
8.7. WHEN OpWhileが実行されたとき、THE System SHALL 条件が真の間ループを実行する
8.8. WHEN OpSwitchが実行されたとき、THE System SHALL 値を評価して一致するcaseを実行する
8.9. WHEN OpBreakが実行されたとき、THE System SHALL 現在のループを抜ける
8.10. WHEN OpContinueが実行されたとき、THE System SHALL 次のループ反復にスキップする
8.11. WHEN OpBinaryOpが実行されたとき、THE System SHALL 二項演算を評価して結果を返す
8.12. WHEN OpUnaryOpが実行されたとき、THE System SHALL 単項演算を評価して結果を返す
8.13. WHEN OpArrayAccessが実行されたとき、THE System SHALL 指定された配列インデックスの値を返す

### 要件9: 変数スコープ管理

**ユーザーストーリー:** 開発者として、適切な変数スコープ管理が欲しい。そうすることで、ローカル変数とグローバル変数が正しく動作する。

#### 受け入れ基準

9.1. WHEN 変数がトップレベルで宣言されたとき、THE System SHALL それをグローバルスコープに格納する
9.2. WHEN 変数が関数内で宣言されたとき、THE System SHALL それをローカルスコープに格納する
9.3. WHEN 関数が呼ばれたとき、THE System SHALL 新しいローカルスコープを作成する
9.4. WHEN 関数が戻るとき、THE System SHALL ローカルスコープを破棄する
9.5. WHEN 変数がアクセスされたとき、THE System SHALL 最初にローカルスコープを検索し、次にグローバルスコープを検索する
9.6. WHEN 変数が事前の宣言なしに代入されたとき、THE System SHALL 現在のスコープにそれを作成する
9.7. WHEN 関数パラメータが渡されたとき、THE System SHALL それらをローカルスコープにバインドする
9.8. THE System SHALL 独立したローカルスコープを持つネストされた関数呼び出しをサポートする

### 要件10: 組み込み関数の実行

**ユーザーストーリー:** 開発者として、組み込み関数を呼び出したい。そうすることで、システム機能を利用できる。

#### 受け入れ基準

10.1. WHEN PlayMIDIが呼ばれたとき、THE System SHALL MIDI再生関数を呼び出す
10.2. WHEN PlayWAVEが呼ばれたとき、THE System SHALL WAV再生関数を呼び出す
10.3. WHEN del_meが呼ばれたとき、THE System SHALL 現在のイベントハンドラを削除する
10.4. WHEN del_allが呼ばれたとき、THE System SHALL すべてのイベントハンドラを削除する
10.5. WHEN del_usが呼ばれたとき、THE System SHALL 現在のイベントハンドラを削除する
10.6. WHEN end_stepが呼ばれたとき、THE System SHALL 現在のステップブロックを終了する
10.7. WHEN ExitTitleが呼ばれたとき、THE System SHALL プログラムを終了する
10.8. WHEN 未知の関数が呼ばれたとき、THE System SHALL エラーをログに記録して実行を継続する
10.9. THE System SHALL 組み込み関数のレジストリを提供する
10.10. THE System SHALL カスタム組み込み関数の登録を許可する

### 要件11: エラーハンドリング

**ユーザーストーリー:** 開発者として、実行時エラーが発生したときに適切に処理したい。そうすることで、プログラムがクラッシュせずに動作を継続できる。

#### 受け入れ基準

11.1. WHEN 実行時エラーが発生したとき、THE System SHALL コンテキスト情報とともにエラーをログに記録する
11.2. WHEN ファイルが見つからないとき、THE System SHALL エラーをログに記録して実行を継続する
11.3. WHEN ゼロ除算が発生したとき、THE System SHALL エラーをログに記録してゼロを返す
11.4. WHEN 配列インデックスが範囲外のとき、THE System SHALL エラーをログに記録してゼロを返す
11.5. WHEN 変数が見つからないとき、THE System SHALL デフォルト値でそれを作成する
11.6. WHEN 関数が見つからないとき、THE System SHALL エラーをログに記録して実行を継続する
11.7. THE System SHALL 利用可能な場合は行番号を含むエラーメッセージを提供する
11.8. THE System SHALL 致命的でないエラーの後も実行を継続する

### 要件12: ヘッドレスモード

**ユーザーストーリー:** 開発者として、GUIなしでスクリプトを実行したい。そうすることで、自動テストやデバッグを行える。

#### 受け入れ基準

12.1. WHEN ヘッドレスモードが有効のとき、THE System SHALL オーディオシステムを初期化する
12.2. WHEN ヘッドレスモードが有効のとき、THE System SHALL すべてのオーディオ出力をミュートする
12.3. WHEN ヘッドレスモードが有効のとき、THE System SHALL MIDI_TIMEイベントを通常通り生成する
12.4. WHEN ヘッドレスモードが有効のとき、THE System SHALL すべての描画操作をスキップする
12.5. WHEN ヘッドレスモードが有効のとき、THE System SHALL すべての描画操作をログに記録する
12.6. WHEN ヘッドレスモードが有効のとき、THE System SHALL すべてのログメッセージにタイムスタンプを追加する
12.7. THE System SHALL コマンドラインフラグによるヘッドレスモードの有効化をサポートする
12.8. THE System SHALL 環境変数によるヘッドレスモードの有効化をサポートする

### 要件13: タイムアウト機能

**ユーザーストーリー:** 開発者として、指定時間後にプログラムを自動終了したい。そうすることで、テスト実行を制御できる。

#### 受け入れ基準

13.1. WHEN タイムアウトが指定されたとき、THE System SHALL 指定された期間後に実行を終了する
13.2. WHEN タイムアウトが期限切れになったとき、THE System SHALL すべてのリソースをクリーンアップする
13.3. WHEN タイムアウトが期限切れになったとき、THE System SHALL タイムアウトメッセージをログに記録する
13.4. THE System SHALL 秒単位でのタイムアウト指定をサポートする
13.5. THE System SHALL コマンドラインフラグによるタイムアウト指定をサポートする
13.6. WHEN タイムアウトが指定されないとき、THE System SHALL 無期限に実行する

### 要件14: 実行ループ

**ユーザーストーリー:** 開発者として、効率的な実行ループが欲しい。そうすることで、イベント処理とOpCode実行がスムーズに動作する。

#### 受け入れ基準

14.1. THE System SHALL イベントを処理しOpCodeを実行するメインイベントループを実行する
14.2. WHEN イベントキューが空のとき、THE System SHALL 次のイベントを待つ
14.3. WHEN イベントが利用可能なとき、THE System SHALL それらを順番に処理する
14.4. WHEN OpCode実行が進行中のとき、THE System SHALL 待機ポイントまで継続する
14.5. WHEN 待機ポイントに到達したとき、THE System SHALL イベントループに制御を戻す
14.6. THE System SHALL イベント処理とOpCode実行のバランスを維持する
14.7. THE System SHALL キューサイズを制限してイベントキューのオーバーフローを防ぐ
14.8. WHEN キューが満杯のとき、THE System SHALL 最も古いイベントを破棄する

### 要件15: プログラム終了

**ユーザーストーリー:** 開発者として、プログラムを適切に終了したい。そうすることで、リソースが正しく解放される。

#### 受け入れ基準

15.1. WHEN ExitTitleが呼ばれたとき、THE System SHALL すべてのオーディオ再生を停止する
15.2. WHEN ExitTitleが呼ばれたとき、THE System SHALL すべてのウィンドウを閉じる
15.3. WHEN ExitTitleが呼ばれたとき、THE System SHALL すべてのリソースをクリーンアップする
15.4. WHEN ExitTitleが呼ばれたとき、THE System SHALL イベントループを終了する
15.5. WHEN すべてのハンドラが削除されたとき、THE System SHALL 明示的な終了まで実行を継続する
15.6. WHEN main関数が完了したとき、THE System SHALL イベント処理を継続する
15.7. THE System SHALL グレースフルシャットダウン機構を提供する

### 要件16: デバッグサポート

**ユーザーストーリー:** 開発者として、実行状態をデバッグしたい。そうすることで、問題を特定して修正できる。

#### 受け入れ基準

16.1. WHEN デバッグモードが有効のとき、THE System SHALL 各OpCode実行をログに記録する
16.2. WHEN デバッグモードが有効のとき、THE System SHALL イベントディスパッチをログに記録する
16.3. WHEN デバッグモードが有効のとき、THE System SHALL ハンドラの登録と削除をログに記録する
16.4. WHEN デバッグモードが有効のとき、THE System SHALL 変数代入をログに記録する
16.5. WHEN デバッグモードが有効のとき、THE System SHALL 関数呼び出しをログに記録する
16.6. THE System SHALL 異なるログレベル（debug、info、warn、error）をサポートする
16.7. THE System SHALL コマンドラインフラグによるログレベル設定を許可する
16.8. THE System SHALL ログメッセージにタイムスタンプを含める

### 要件17: Wait()関数のサポート

**ユーザーストーリー:** 開発者として、Wait()関数で指定ステップ数だけ待機したい。そうすることで、タイミング制御を柔軟に行える。

#### 受け入れ基準

17.1. WHEN Wait(n)が呼ばれたとき、THE System SHALL nイベント発生分だけ実行を一時停止する
17.2. WHEN Wait(n)がmes(TIME)ハンドラ内で呼ばれたとき、THE System SHALL n回のTIMEイベントを待つ
17.3. WHEN Wait(n)がmes(MIDI_TIME)ハンドラ内で呼ばれたとき、THE System SHALL n回のMIDI_TIMEイベントを待つ
17.4. WHEN Wait(0)が呼ばれたとき、THE System SHALL 即座に実行を継続する
17.5. WHEN Wait(n)がn<0で呼ばれたとき、THE System SHALL それをWait(0)として扱う
17.6. THE System SHALL 各ハンドラに対して個別の待機カウンタを維持する
17.7. WHEN Wait()中にハンドラが削除されたとき、THE System SHALL 待機をキャンセルする

### 要件18: MIDIテンポ変更のサポート

**ユーザーストーリー:** 開発者として、MIDI再生中のテンポ変更に対応したい。そうすることで、テンポが変わる曲でも正しく同期できる。

#### 受け入れ基準

18.1. WHEN MIDIファイルにテンポ変更イベントが含まれるとき、THE System SHALL それらを検出する
18.2. WHEN テンポ変更イベントに遭遇したとき、THE System SHALL MIDI_TIMEイベント間隔を更新する
18.3. WHEN テンポが120 BPMから60 BPMに変更されたとき、THE System SHALL MIDI_TIMEイベント間隔を2倍にする
18.4. THE System SHALL テンポ変更を跨いで正確な同期を維持する
18.5. WHEN 複数のテンポ変更が発生したとき、THE System SHALL 各変更を正しく処理する

### 要件19: 配列のサポート

**ユーザーストーリー:** 開発者として、配列を使用したい。そうすることで、複数の値を効率的に管理できる。

#### 受け入れ基準

19.1. WHEN 配列が宣言されたとき、THE System SHALL 配列用のストレージを割り当てる
19.2. WHEN 配列要素がアクセスされたとき、THE System SHALL 指定されたインデックスの値を返す
19.3. WHEN 配列要素が代入されたとき、THE System SHALL 指定されたインデックスに値を格納する
19.4. WHEN 配列インデックスが負のとき、THE System SHALL エラーをログに記録してゼロを返す
19.5. WHEN 配列インデックスが配列サイズを超えるとき、THE System SHALL 自動的に配列を拡張する
19.6. THE System SHALL 動的な配列リサイズをサポートする
19.7. THE System SHALL 新しい配列要素をゼロで初期化する
19.8. WHEN 配列が関数に渡されたとき、THE System SHALL それを参照渡しする

### 要件20: 関数呼び出しとスタック管理

**ユーザーストーリー:** 開発者として、関数を呼び出して戻り値を受け取りたい。そうすることで、コードを構造化できる。

#### 受け入れ基準

20.1. WHEN 関数が呼ばれたとき、THE System SHALL 新しいスタックフレームをプッシュする
20.2. WHEN 関数が戻るとき、THE System SHALL スタックフレームをポップする
20.3. WHEN 関数が戻り値を持つとき、THE System SHALL それを呼び出し元に渡す
20.4. WHEN 関数が戻り値を持たないとき、THE System SHALL ゼロを返す
20.5. THE System SHALL 再帰的な関数呼び出しをサポートする
20.6. THE System SHALL スタックオーバーフローを検出してエラーを報告する
20.7. THE System SHALL 最大スタック深度を1000フレームに維持する
20.8. WHEN スタックオーバーフローが発生したとき、THE System SHALL エラーをログに記録して実行を終了する

### 要件21: StrPrint関数の実装

**ユーザーストーリー:** 開発者として、printf形式のフォーマット指定子を使用して動的に文字列を生成したい。これにより、ファイル名の連番生成やデバッグメッセージの作成が可能になる。

#### 受け入れ基準

21.1. WHEN StrPrint がフォーマット文字列と引数で呼び出された場合、THE System SHALL フォーマット指定子に従ってフォーマットされた文字列を返す。

21.2. THE System SHALL 10進整数フォーマット用の `%ld` フォーマット指定子をサポートし、Go言語の `%d` に変換する。

21.3. THE System SHALL 16進数フォーマット用の `%lx` フォーマット指定子をサポートし、Go言語の `%x` に変換する。

21.4. THE System SHALL 文字列フォーマット用の `%s` フォーマット指定子をサポートする。

21.5. THE System SHALL ゼロパディング整数用の `%03d` などの幅とパディング指定子をサポートする。

21.6. WHEN StrPrint がエスケープシーケンス（`\n`, `\t`, `\r`）を含むフォーマット文字列で呼び出された場合、THE System SHALL それらを実際の制御文字に変換する。

21.7. WHEN StrPrint がフォーマット指定子より少ない引数で呼び出された場合、THE System SHALL クラッシュせずに適切に処理する（Go言語のfmt.Sprintfの動作に従う）。

21.8. WHEN StrPrint がフォーマット指定子より多い引数で呼び出された場合、THE System SHALL 余分な引数を無視する。

### 要件22: CreatePic 3引数パターンの実装

**ユーザーストーリー:** 開発者として、既存のピクチャーを参照しながら任意のサイズの新しいピクチャーを作成したい。これにより、スプライトシートから個別のスプライト用バッファを効率的に作成できる。

#### 受け入れ基準

22.1. WHEN CreatePic が3つの引数（srcPicID, width, height）で呼び出された場合、THE System SHALL 指定されたサイズの新しい空のピクチャーを作成する。

22.2. WHEN CreatePic が3引数パターンで呼び出された場合、THE System SHALL ソースピクチャーの内容をコピーせず、空のピクチャーを返す。

22.3. WHEN CreatePic が存在しないソースピクチャーIDで呼び出された場合、THE System SHALL エラーを返す。

22.4. WHEN CreatePic が0以下の幅または高さで呼び出された場合、THE System SHALL エラーを返す。

22.5. THE System SHALL 既存の1引数パターン（CreatePic(srcID)）と2引数パターン（CreatePic(width, height)）との互換性を維持する。

### 要件23: CapTitle 1引数パターンの修正

**ユーザーストーリー:** 開発者として、1つの引数でCapTitleを呼び出した際に、全ての仮想ウィンドウのタイトルを一括で設定したい。これにより、アプリケーション全体のタイトル管理が簡単になる。

#### 受け入れ基準

23.1. WHEN CapTitle が1つの引数（title）で呼び出された場合、THE System SHALL 全ての既存の仮想ウィンドウのキャプションを設定する。

23.2. WHEN CapTitle が1つの引数で呼び出され、仮想ウィンドウが存在しない場合、THE System SHALL エラーを発生させずに正常に終了する。

23.3. WHEN CapTitle が2つの引数（winID, title）で呼び出された場合、THE System SHALL 指定されたウィンドウのキャプションのみを設定する（既存動作を維持）。

23.4. WHEN CapTitle が存在しないウィンドウIDで呼び出された場合、THE System SHALL エラーを発生させずに正常に終了する。

23.5. THE System SHALL 空文字列（""）をタイトルとして受け入れ、ウィンドウのキャプションをクリアする。
