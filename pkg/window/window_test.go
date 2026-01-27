package window

import (
	"strings"
	"testing"
	"time"

	"github.com/zurustar/son-et/pkg/title"
)

func TestNewGame(t *testing.T) {
	titles := []title.FillyTitle{
		{Name: "Title1", Path: "/path/1", IsEmbedded: false},
		{Name: "Title2", Path: "/path/2", IsEmbedded: false},
	}

	game := NewGame(ModeSelection, titles, 10*time.Second)

	if game == nil {
		t.Fatal("NewGame returned nil")
	}

	if game.mode != ModeSelection {
		t.Errorf("expected mode ModeSelection, got %v", game.mode)
	}

	if len(game.titles) != 2 {
		t.Errorf("expected 2 titles, got %d", len(game.titles))
	}

	if game.selectedIndex != 0 {
		t.Errorf("expected selectedIndex 0, got %d", game.selectedIndex)
	}

	if game.timeout != 10*time.Second {
		t.Errorf("expected timeout 10s, got %v", game.timeout)
	}
}

func TestLayout(t *testing.T) {
	game := NewGame(ModeDesktop, nil, 0)

	width, height := game.Layout(0, 0)

	if width != 1024 {
		t.Errorf("expected width 1024, got %d", width)
	}

	if height != 768 {
		t.Errorf("expected height 768, got %d", height)
	}
}

func TestGetSelectedTitle(t *testing.T) {
	titles := []title.FillyTitle{
		{Name: "Title1", Path: "/path/1", IsEmbedded: false},
	}

	game := NewGame(ModeSelection, titles, 0)

	// 初期状態ではnil
	if game.GetSelectedTitle() != nil {
		t.Error("expected nil before selection")
	}

	// タイトルを選択
	game.selectedTitle = &titles[0]

	selected := game.GetSelectedTitle()
	if selected == nil {
		t.Fatal("GetSelectedTitle returned nil")
	}

	if selected.Name != "Title1" {
		t.Errorf("expected Title1, got %s", selected.Name)
	}
}

func TestUpdateSelection_Timeout(t *testing.T) {
	titles := []title.FillyTitle{
		{Name: "Title1", Path: "/path/1", IsEmbedded: false},
	}

	// 非常に短いタイムアウトを設定
	game := NewGame(ModeSelection, titles, 1*time.Nanosecond)

	// 少し待機
	time.Sleep(10 * time.Millisecond)

	// Update を呼び出すとタイムアウトで終了するはず
	err := game.Update()
	if err == nil {
		t.Error("expected timeout error, got nil")
	}
}

func TestMode(t *testing.T) {
	tests := []struct {
		name string
		mode Mode
	}{
		{"Selection mode", ModeSelection},
		{"Desktop mode", ModeDesktop},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game := NewGame(tt.mode, nil, 0)
			if game.mode != tt.mode {
				t.Errorf("expected mode %v, got %v", tt.mode, game.mode)
			}
		})
	}
}

func TestUpdateDesktop(t *testing.T) {
	titles := []title.FillyTitle{
		{Name: "Title1", Path: "/path/1", IsEmbedded: false},
	}

	game := NewGame(ModeDesktop, titles, 0)

	// 通常の更新では nil を返す
	err := game.updateDesktop()
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestUpdateDesktop_Timeout(t *testing.T) {
	titles := []title.FillyTitle{
		{Name: "Title1", Path: "/path/1", IsEmbedded: false},
	}

	// 非常に短いタイムアウトを設定
	game := NewGame(ModeDesktop, titles, 1*time.Nanosecond)

	// 少し待機
	time.Sleep(10 * time.Millisecond)

	// Update を呼び出すとタイムアウトで終了するはず
	err := game.Update()
	if err == nil {
		t.Error("expected timeout error, got nil")
	}
}

func TestRunHeadless_SingleTitle(t *testing.T) {
	titles := []title.FillyTitle{
		{Name: "Title1", Path: "/path/1", IsEmbedded: false},
	}

	var output strings.Builder
	input := strings.NewReader("")

	selected, err := RunHeadless(titles, 0, input, &output)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if selected == nil {
		t.Fatal("expected selected title, got nil")
	}

	if selected.Name != "Title1" {
		t.Errorf("expected Title1, got %s", selected.Name)
	}

	if !strings.Contains(output.String(), "Auto-selecting") {
		t.Error("expected auto-selection message")
	}
}

func TestRunHeadless_MultipleTitle_ValidSelection(t *testing.T) {
	titles := []title.FillyTitle{
		{Name: "Title1", Path: "/path/1", IsEmbedded: false},
		{Name: "Title2", Path: "/path/2", IsEmbedded: false},
		{Name: "Title3", Path: "/path/3", IsEmbedded: false},
	}

	var output strings.Builder
	input := strings.NewReader("2\n")

	selected, err := RunHeadless(titles, 0, input, &output)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if selected == nil {
		t.Fatal("expected selected title, got nil")
	}

	if selected.Name != "Title2" {
		t.Errorf("expected Title2, got %s", selected.Name)
	}

	outputStr := output.String()
	if !strings.Contains(outputStr, "Available FILLY Titles:") {
		t.Error("expected title list header")
	}
	if !strings.Contains(outputStr, "Title1") {
		t.Error("expected Title1 in list")
	}
	if !strings.Contains(outputStr, "Selected: Title2") {
		t.Error("expected selection confirmation")
	}
}

func TestRunHeadless_InvalidThenValid(t *testing.T) {
	titles := []title.FillyTitle{
		{Name: "Title1", Path: "/path/1", IsEmbedded: false},
		{Name: "Title2", Path: "/path/2", IsEmbedded: false},
	}

	var output strings.Builder
	// 無効な入力の後に有効な入力
	input := strings.NewReader("abc\n0\n3\n1\n")

	selected, err := RunHeadless(titles, 0, input, &output)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if selected == nil {
		t.Fatal("expected selected title, got nil")
	}

	if selected.Name != "Title1" {
		t.Errorf("expected Title1, got %s", selected.Name)
	}

	outputStr := output.String()
	if !strings.Contains(outputStr, "Invalid input") {
		t.Error("expected invalid input message")
	}
	if !strings.Contains(outputStr, "Invalid selection") {
		t.Error("expected invalid selection message")
	}
}

func TestRunHeadless_Quit(t *testing.T) {
	titles := []title.FillyTitle{
		{Name: "Title1", Path: "/path/1", IsEmbedded: false},
		{Name: "Title2", Path: "/path/2", IsEmbedded: false},
	}

	var output strings.Builder
	input := strings.NewReader("q\n")

	selected, err := RunHeadless(titles, 0, input, &output)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if selected != nil {
		t.Errorf("expected nil, got %v", selected)
	}

	if !strings.Contains(err.Error(), "cancelled") {
		t.Errorf("expected 'cancelled' error, got %v", err)
	}
}

func TestRunHeadless_Timeout(t *testing.T) {
	titles := []title.FillyTitle{
		{Name: "Title1", Path: "/path/1", IsEmbedded: false},
		{Name: "Title2", Path: "/path/2", IsEmbedded: false},
	}

	var output strings.Builder
	// 入力を遅延させるために、入力なしで実行
	input := strings.NewReader("")

	// 非常に短いタイムアウトを設定
	selected, err := RunHeadless(titles, 10*time.Millisecond, input, &output)

	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}

	if selected != nil {
		t.Errorf("expected nil, got %v", selected)
	}

	if !strings.Contains(err.Error(), "timeout") && !strings.Contains(err.Error(), "input closed") {
		t.Errorf("expected 'timeout' or 'input closed' error, got %v", err)
	}
}
