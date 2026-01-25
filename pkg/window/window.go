package window

import (
	"bufio"
	"context"
	"fmt"
	"image/color"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
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
}

// GraphicsSystemInterface defines the interface for graphics operations
type GraphicsSystemInterface interface {
	Update() error
	Draw(screen *ebiten.Image)
	Shutdown()
}

// VMRunnerInterface defines the interface for VM operations
type VMRunnerInterface interface {
	IsRunning() bool
	Stop()
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
	g.graphicsSystem = gs
}

// SetVMRunner sets the VM runner for desktop mode
func (g *Game) SetVMRunner(vm VMRunnerInterface) {
	g.vmRunner = vm
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
		return ebiten.Termination
	}

	// Escキー（1回だけ反応）
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}

	return nil
}

// updateDesktop 仮想デスクトップの更新
func (g *Game) updateDesktop() error {
	// Escキーで終了（1回だけ反応）
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}

	// VMが停止していたら終了
	if g.vmRunner != nil && !g.vmRunner.IsRunning() {
		return ebiten.Termination
	}

	// GraphicsSystemの更新
	if g.graphicsSystem != nil {
		if err := g.graphicsSystem.Update(); err != nil {
			return err
		}
	}

	return nil
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
	// skelton要件 3.2: ウィンドウサイズは 1280x720 ピクセル
	return 1280, 720
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
	ebiten.SetWindowSize(1280, 720)
	ebiten.SetWindowTitle("FILLY - son-et")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeDisabled)

	// ゲームを実行
	if err := ebiten.RunGame(game); err != nil {
		return nil, fmt.Errorf("failed to run game: %w", err)
	}

	return game.GetSelectedTitle(), nil
}
