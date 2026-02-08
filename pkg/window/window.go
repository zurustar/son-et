package window

import (
	"bufio"
	"context"
	"fmt"
	"image/color"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/zurustar/son-et/pkg/logger"
	"github.com/zurustar/son-et/pkg/title"
	"golang.org/x/image/font/basicfont"
)

var (
	// 背景色 #0087C8
	backgroundColor = color.RGBA{0x00, 0x87, 0xC8, 0xFF}
	// テキスト色（白）
	textColor = color.White
	// 選択中のテキスト色（黄色）
	selectedTextColor = color.RGBA{0xFF, 0xFF, 0x00, 0xFF}
	// デフォルトフォント
	defaultFace = text.NewGoXFace(basicfont.Face7x13)
)

// Mode はウィンドウの表示モードを表す
type Mode int

const (
	ModeSelection Mode = iota // タイトル選択画面
	ModeDesktop               // 仮想デスクトップ
)

// Game はEbitengineのゲームインターフェースを実装する
type Game struct {
	mode          Mode               // 現在のモード
	titles        []title.FillyTitle // 利用可能なタイトル一覧
	selectedIndex int                // 選択中のタイトルのインデックス
	selectedTitle *title.FillyTitle  // 選択されたタイトル
	timeout       time.Duration      // タイムアウト時間
	startTime     time.Time          // 開始時刻

	// Graphics system integration
	graphicsSystem GraphicsSystemInterface
	vmRunner       VMRunnerInterface
	eventPusher    MouseEventPusher

	// VM startup control
	vmStartFunc func()     // VMを開始する関数
	vmStarted   bool       // VMが開始されたかどうか
	vmErrCh     chan error // VMのエラーチャネル

	// Mode transition callback (for selection -> desktop transition)
	onTitleSelected func(title *title.FillyTitle) error
	transitionError error // モード遷移時のエラー

	// Title selection mode support (Requirements 2.1, 3.1, 5.1)
	hasTitleSelection bool         // タイトル選択画面があるかどうか（複数タイトル時true）
	onTitleExit       func() error // タイトル終了時のコールバック

	// Mouse state tracking for event generation
	lastMouseX int
	lastMouseY int
	mu         sync.RWMutex
}

// GraphicsSystemInterface defines the interface for graphics operations
type GraphicsSystemInterface interface {
	Update() error
	Draw(screen *ebiten.Image)
	Shutdown()
	// GetVirtualWidth returns the virtual desktop width for coordinate conversion
	GetVirtualWidth() int
	// GetVirtualHeight returns the virtual desktop height for coordinate conversion
	GetVirtualHeight() int
}

// VMRunnerInterface defines the interface for VM operations
type VMRunnerInterface interface {
	IsRunning() bool
	IsFullyStopped() bool
	Stop()
}

// EventQueueInterface defines the interface for pushing events to the VM
type EventQueueInterface interface {
	Push(event interface{})
}

// MouseEventPusher defines the interface for pushing mouse events
// This is used to decouple the window package from the vm package
type MouseEventPusher interface {
	PushMouseEvent(eventType string, windowID, x, y int)
	PushKeyEvent(eventType string, keyCode int)
}

// NewGame Gameを作成
func NewGame(mode Mode, titles []title.FillyTitle, timeout time.Duration) *Game {
	return &Game{
		mode:          mode,
		titles:        titles,
		selectedIndex: 0,
		timeout:       timeout,
		startTime:     time.Now(),
	}
}

// SetGraphicsSystem sets the graphics system for desktop mode
func (g *Game) SetGraphicsSystem(gs GraphicsSystemInterface) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.graphicsSystem = gs
}

// SetVMRunner sets the VM runner for desktop mode
func (g *Game) SetVMRunner(vm VMRunnerInterface) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.vmRunner = vm
}

// SetEventPusher sets the event pusher for mouse events
func (g *Game) SetEventPusher(pusher MouseEventPusher) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.eventPusher = pusher
}

// SetVMStartFunc sets the function to start the VM
// This function will be called on the first Update() call to ensure
// Ebitengine is fully initialized before VM starts
func (g *Game) SetVMStartFunc(startFunc func(), errCh chan error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.vmStartFunc = startFunc
	g.vmErrCh = errCh
}

// GetVMErrorChannel returns the VM error channel
func (g *Game) GetVMErrorChannel() chan error {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.vmErrCh
}

// SetOnTitleSelected sets the callback function called when a title is selected
// This allows the application to set up VM and graphics system before transitioning to desktop mode
func (g *Game) SetOnTitleSelected(callback func(title *title.FillyTitle) error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.onTitleSelected = callback
}

// SetHasTitleSelection sets whether the title selection screen is available
// When true, pressing ESC in desktop mode returns to the selection screen
// When false, pressing ESC in desktop mode exits the program
// Requirements: 2.1, 3.1, 5.1
func (g *Game) SetHasTitleSelection(has bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.hasTitleSelection = has
}

// SetOnTitleExit sets the callback function called when exiting a title
// This callback is used for resource cleanup (VM, Graphics, Audio)
// Requirements: 2.2, 2.3, 2.4, 4.1, 4.2, 4.3
func (g *Game) SetOnTitleExit(callback func() error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.onTitleExit = callback
}

// GetTransitionError returns any error that occurred during mode transition
func (g *Game) GetTransitionError() error {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.transitionError
}

// Update ゲームロジックの更新（Ebitengineが毎フレーム呼び出す）
func (g *Game) Update() error {
	// タイムアウトチェック
	if g.timeout > 0 && time.Since(g.startTime) >= g.timeout {
		return ebiten.Termination
	}

	switch g.mode {
	case ModeSelection:
		return g.updateSelection()
	case ModeDesktop:
		return g.updateDesktop()
	}

	return nil
}

// updateSelection タイトル選択画面の更新
func (g *Game) updateSelection() error {
	// 上矢印キー（1回だけ反応）
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		if g.selectedIndex > 0 {
			g.selectedIndex--
		}
	}

	// 下矢印キー（1回だけ反応）
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		if g.selectedIndex < len(g.titles)-1 {
			g.selectedIndex++
		}
	}

	// Enterキー（1回だけ反応）
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		g.selectedTitle = &g.titles[g.selectedIndex]

		// コールバックが設定されている場合は、デスクトップモードに遷移
		g.mu.RLock()
		callback := g.onTitleSelected
		g.mu.RUnlock()

		if callback != nil {
			// コールバックを呼び出してVM/GraphicsSystemをセットアップ
			if err := callback(g.selectedTitle); err != nil {
				g.mu.Lock()
				g.transitionError = err
				g.mu.Unlock()
				return ebiten.Termination
			}
			// デスクトップモードに遷移
			g.mu.Lock()
			g.mode = ModeDesktop
			g.startTime = time.Now() // タイムアウトをリセット
			g.mu.Unlock()
			return nil
		}

		// コールバックがない場合は終了（従来の動作）
		return ebiten.Termination
	}

	// Escキー（1回だけ反応）
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}

	return nil
}

// updateDesktop 仮想デスクトップの更新
// 要件 14.1: EbitengineのUpdate()内でVMのイベント処理を呼び出す
// 要件 14.4: VMが終了したとき、Ebitengineのゲームループを終了する
// 要件 14.5: Ebitengineのウィンドウが閉じられたとき、VMを停止する
func (g *Game) updateDesktop() error {
	// VMの開始（最初のUpdate()呼び出し時に実行）
	// これにより、Ebitengineが完全に初期化された後にVMが開始される
	g.mu.Lock()
	if !g.vmStarted && g.vmStartFunc != nil {
		g.vmStarted = true
		g.vmStartFunc()
	}
	g.mu.Unlock()

	// Escキーで終了または選択画面に戻る（1回だけ反応）
	// 要件 2.1: hasTitleSelection=trueの場合、タイトル選択画面に戻る
	// 要件 3.2: hasTitleSelection=falseの場合、プログラムを終了する
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.mu.RLock()
		hasTitleSelection := g.hasTitleSelection
		g.mu.RUnlock()

		if hasTitleSelection {
			// 複数タイトル環境: タイトル選択画面に戻る
			return g.returnToSelection()
		}

		// 単一タイトル環境: プログラムを終了
		g.mu.RLock()
		vmRunner := g.vmRunner
		g.mu.RUnlock()
		if vmRunner != nil {
			vmRunner.Stop()
		}
		return ebiten.Termination
	}

	// VMが完全に停止しても、ユーザーが明示的に終了するまでウィンドウは開いたまま
	// Escキーまたはウィンドウを閉じることで終了する
	// 要件変更: タイトル終了後もウィンドウを閉じない

	// マウスイベントを処理
	// 要件 14.6: マウスイベントをEbitengineから取得し、VMのイベントキューに追加する
	g.processMouseEvents()

	// キーボードイベントを処理
	g.processKeyboardEvents()

	// GraphicsSystemの更新（コマンドキューの処理）
	// 要件 14.2: EbitengineのDraw()内で描画コマンドキューを処理する
	// Note: 実際のコマンドキュー処理はUpdate()で行う（Ebitengineの推奨）
	g.mu.RLock()
	graphicsSystem := g.graphicsSystem
	g.mu.RUnlock()
	if graphicsSystem != nil {
		if err := graphicsSystem.Update(); err != nil {
			return err
		}
	}

	return nil
}

// returnToSelection はデスクトップモードからタイトル選択画面に戻る
// 要件 2.1, 2.5, 5.1: エスケープキーでタイトル選択画面に戻る
// 注: 完全な実装はタスク2.2で行う
func (g *Game) returnToSelection() error {
	// VMを停止
	g.mu.RLock()
	vmRunner := g.vmRunner
	onTitleExit := g.onTitleExit
	g.mu.RUnlock()

	if vmRunner != nil {
		vmRunner.Stop()
	}

	// リソースクリーンアップコールバックを呼び出す
	if onTitleExit != nil {
		if err := onTitleExit(); err != nil {
			logger.GetLogger().Error("onTitleExit callback failed", "error", err)
		}
	}

	// モードをModeSelectionに変更
	g.mu.Lock()
	g.mode = ModeSelection
	g.vmStarted = false
	g.graphicsSystem = nil
	g.vmRunner = nil
	g.eventPusher = nil
	g.mu.Unlock()

	return nil
}

// processMouseEvents はマウスイベントを処理してVMに伝達する
// 要件 14.6: マウスイベントをEbitengineから取得し、VMのイベントキューに追加する
// 要件 8.7: マウスイベントが発生したとき、仮想デスクトップ座標に変換してMesP2、MesP3に設定する
func (g *Game) processMouseEvents() {
	g.mu.RLock()
	eventPusher := g.eventPusher
	graphicsSystem := g.graphicsSystem
	g.mu.RUnlock()

	if eventPusher == nil {
		return
	}

	// マウス座標を取得
	mouseX, mouseY := ebiten.CursorPosition()

	// 仮想デスクトップ座標に変換
	virtualX, virtualY := g.screenToVirtual(mouseX, mouseY, graphicsSystem)

	// 左ボタン押し下げ (LBDOWN)
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		// windowID は 0 (メインウィンドウ) として扱う
		// 実際のウィンドウIDの判定は後のフェーズで実装
		eventPusher.PushMouseEvent("LBDOWN", 0, virtualX, virtualY)
	}

	// 左ボタン離し (CLICK) - ボタンを離した時点でクリック完了
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		eventPusher.PushMouseEvent("CLICK", 0, virtualX, virtualY)
	}

	// 右ボタン押し下げ (RBDOWN)
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		eventPusher.PushMouseEvent("RBDOWN", 0, virtualX, virtualY)
	}

	// 右ダブルクリック（Ebitengineでは直接サポートされていないため、
	// 短時間内の2回クリックで判定する必要がある - 将来の拡張）
	// TODO: 詳細はdocs/unimplemented-features.mdを参照

	// マウス座標を保存
	g.mu.Lock()
	g.lastMouseX = virtualX
	g.lastMouseY = virtualY
	g.mu.Unlock()
}

// processKeyboardEvents はキーボードイベントを処理してVMに伝達する
func (g *Game) processKeyboardEvents() {
	g.mu.RLock()
	eventPusher := g.eventPusher
	g.mu.RUnlock()

	if eventPusher == nil {
		return
	}

	// A-Zキーをチェック
	// 小文字のASCIIコード（97-122）を送信する
	// TFYスクリプトは小文字のASCIIコードを期待している（例: ka = 97）
	keys := []struct {
		key  ebiten.Key
		char rune
	}{
		{ebiten.KeyA, 'a'}, {ebiten.KeyB, 'b'}, {ebiten.KeyC, 'c'}, {ebiten.KeyD, 'd'},
		{ebiten.KeyE, 'e'}, {ebiten.KeyF, 'f'}, {ebiten.KeyG, 'g'}, {ebiten.KeyH, 'h'},
		{ebiten.KeyI, 'i'}, {ebiten.KeyJ, 'j'}, {ebiten.KeyK, 'k'}, {ebiten.KeyL, 'l'},
		{ebiten.KeyM, 'm'}, {ebiten.KeyN, 'n'}, {ebiten.KeyO, 'o'}, {ebiten.KeyP, 'p'},
		{ebiten.KeyQ, 'q'}, {ebiten.KeyR, 'r'}, {ebiten.KeyS, 's'}, {ebiten.KeyT, 't'},
		{ebiten.KeyU, 'u'}, {ebiten.KeyV, 'v'}, {ebiten.KeyW, 'w'}, {ebiten.KeyX, 'x'},
		{ebiten.KeyY, 'y'}, {ebiten.KeyZ, 'z'},
	}

	for _, k := range keys {
		if inpututil.IsKeyJustPressed(k.key) {
			// CHARイベントを生成
			// MesP2にキーコード（ASCIIコード）を設定
			eventPusher.PushKeyEvent("CHAR", int(k.char))
		}
	}
}

// screenToVirtual はスクリーン座標を仮想デスクトップ座標に変換する
// 要件 8.7: マウスイベントが発生したとき、仮想デスクトップ座標に変換する
func (g *Game) screenToVirtual(screenX, screenY int, gs GraphicsSystemInterface) (int, int) {
	// 仮想デスクトップのサイズを取得
	virtualWidth := 1024
	virtualHeight := 768
	if gs != nil {
		virtualWidth = gs.GetVirtualWidth()
		virtualHeight = gs.GetVirtualHeight()
	}

	// 実際のウィンドウサイズを取得
	screenWidth, screenHeight := ebiten.WindowSize()
	if screenWidth == 0 || screenHeight == 0 {
		// ウィンドウサイズが取得できない場合はそのまま返す
		return screenX, screenY
	}

	// スケーリング係数を計算（アスペクト比を維持）
	scaleX := float64(screenWidth) / float64(virtualWidth)
	scaleY := float64(screenHeight) / float64(virtualHeight)
	scale := scaleX
	if scaleY < scaleX {
		scale = scaleY
	}

	// レターボックスのオフセットを計算
	offsetX := (float64(screenWidth) - float64(virtualWidth)*scale) / 2
	offsetY := (float64(screenHeight) - float64(virtualHeight)*scale) / 2

	// 仮想デスクトップ座標に変換
	virtualX := int((float64(screenX) - offsetX) / scale)
	virtualY := int((float64(screenY) - offsetY) / scale)

	// 範囲チェック
	if virtualX < 0 {
		virtualX = 0
	}
	if virtualX >= virtualWidth {
		virtualX = virtualWidth - 1
	}
	if virtualY < 0 {
		virtualY = 0
	}
	if virtualY >= virtualHeight {
		virtualY = virtualHeight - 1
	}

	return virtualX, virtualY
}

// Draw 画面描画（Ebitengineが毎フレーム呼び出す）
func (g *Game) Draw(screen *ebiten.Image) {
	// skelton要件 3.2: 背景色は #0087C8
	screen.Fill(backgroundColor)

	switch g.mode {
	case ModeSelection:
		g.drawSelection(screen)
	case ModeDesktop:
		g.drawDesktop(screen)
	}
}

// drawSelection タイトル選択画面の描画
func (g *Game) drawSelection(screen *ebiten.Image) {
	// タイトルを表示
	titleText := "Select a FILLY Title"
	titleOp := &text.DrawOptions{}
	titleOp.GeoM.Translate(50, 50)
	titleOp.ColorScale.ScaleWithColor(textColor)
	text.Draw(screen, titleText, defaultFace, titleOp)

	// タイトル一覧を表示
	for i, t := range g.titles {
		y := 120 + float64(i*40)

		// 選択中のタイトルは色を変える
		prefix := "  "
		if i == g.selectedIndex {
			prefix = "> "
		}

		titleName := prefix + t.Name
		op := &text.DrawOptions{}
		op.GeoM.Translate(70, y)
		if i == g.selectedIndex {
			op.ColorScale.ScaleWithColor(selectedTextColor)
		} else {
			op.ColorScale.ScaleWithColor(textColor)
		}
		text.Draw(screen, titleName, defaultFace, op)
	}

	// 操作説明を表示
	helpText := "Use UP/DOWN to select, ENTER to confirm, ESC to exit"
	helpOp := &text.DrawOptions{}
	helpOp.GeoM.Translate(50, 650)
	helpOp.ColorScale.ScaleWithColor(textColor)
	text.Draw(screen, helpText, defaultFace, helpOp)
}

// drawDesktop 仮想デスクトップの描画
func (g *Game) drawDesktop(screen *ebiten.Image) {
	// GraphicsSystemで描画（背景色の上に描画される）
	if g.graphicsSystem != nil {
		g.graphicsSystem.Draw(screen)
	}
	// GraphicsSystemが設定されていない場合は、背景色のみ表示される
}

// Layout 画面サイズを返す
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	// skelton要件 3.2: ウィンドウサイズは 1024x768 ピクセル
	return 1024, 768
}

// GetSelectedTitle 選択されたタイトルを取得
func (g *Game) GetSelectedTitle() *title.FillyTitle {
	return g.selectedTitle
}

// RunHeadless ヘッドレスモードでタイトル選択を実行
func RunHeadless(titles []title.FillyTitle, timeout time.Duration, reader io.Reader, writer io.Writer) (*title.FillyTitle, error) {
	// タイトルが1つの場合は自動選択
	if len(titles) == 1 {
		fmt.Fprintf(writer, "Auto-selecting title: %s\n", titles[0].Name)
		return &titles[0], nil
	}

	// タイムアウト処理用のコンテキスト
	ctx := context.Background()
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	// タイトル一覧を表示
	fmt.Fprintln(writer, "Available FILLY Titles:")
	for i, t := range titles {
		fmt.Fprintf(writer, "  %d: %s\n", i+1, t.Name)
	}
	fmt.Fprintln(writer)

	// 選択を受け付ける
	scanner := bufio.NewScanner(reader)
	resultCh := make(chan *title.FillyTitle)
	errCh := make(chan error)

	go func() {
		for {
			fmt.Fprint(writer, "Select a title (1-", len(titles), ") or 'q' to quit: ")
			if !scanner.Scan() {
				if err := scanner.Err(); err != nil {
					errCh <- fmt.Errorf("failed to read input: %w", err)
				} else {
					errCh <- fmt.Errorf("input closed")
				}
				return
			}

			input := strings.TrimSpace(scanner.Text())

			// 終了コマンド
			if input == "q" || input == "Q" {
				errCh <- fmt.Errorf("user cancelled")
				return
			}

			// 数値に変換
			num, err := strconv.Atoi(input)
			if err != nil {
				fmt.Fprintln(writer, "Invalid input. Please enter a number.")
				continue
			}

			// 範囲チェック
			if num < 1 || num > len(titles) {
				fmt.Fprintf(writer, "Invalid selection. Please enter a number between 1 and %d.\n", len(titles))
				continue
			}

			// 選択されたタイトルを返す
			selected := &titles[num-1]
			fmt.Fprintf(writer, "Selected: %s\n", selected.Name)
			resultCh <- selected
			return
		}
	}()

	// タイムアウトまたは選択完了を待つ
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("timeout")
	case err := <-errCh:
		return nil, err
	case selected := <-resultCh:
		return selected, nil
	}
}

// Run GUIモードでウィンドウを実行
func Run(mode Mode, titles []title.FillyTitle, timeout time.Duration) (*title.FillyTitle, error) {
	game := NewGame(mode, titles, timeout)

	// ウィンドウ設定
	ebiten.SetWindowSize(1024, 768)
	ebiten.SetWindowTitle("son-et - FILLY interpreter")
	// 要件 8.5: アスペクト比を維持してスケーリングする
	// 要件 8.6: スケーリング時にレターボックス（黒帯）を表示する
	// WindowResizingModeEnabledを使用してウィンドウのリサイズを許可
	// Ebitengineが自動的にアスペクト比を維持してスケーリングし、
	// レターボックスを表示する
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	// ゲームを実行
	if err := ebiten.RunGame(game); err != nil {
		return nil, fmt.Errorf("failed to run game: %w", err)
	}

	return game.GetSelectedTitle(), nil
}
