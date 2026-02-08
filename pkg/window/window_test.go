package window

import (
	"bytes"
	"log/slog"
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

// TestSetHasTitleSelection tests the SetHasTitleSelection method
// Requirements: 2.1, 3.1, 5.1
func TestSetHasTitleSelection(t *testing.T) {
	tests := []struct {
		name     string
		setValue bool
	}{
		{"Set to true", true},
		{"Set to false", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game := NewGame(ModeDesktop, nil, 0)

			// 初期状態はfalse
			if game.hasTitleSelection {
				t.Error("expected hasTitleSelection to be false initially")
			}

			// 値を設定
			game.SetHasTitleSelection(tt.setValue)

			// 設定した値が反映されていることを確認
			if game.hasTitleSelection != tt.setValue {
				t.Errorf("expected hasTitleSelection to be %v, got %v", tt.setValue, game.hasTitleSelection)
			}
		})
	}
}

// TestSetHasTitleSelection_Toggle tests toggling the hasTitleSelection flag
// Requirements: 2.1, 3.1, 5.1
func TestSetHasTitleSelection_Toggle(t *testing.T) {
	game := NewGame(ModeDesktop, nil, 0)

	// false -> true -> false の順で切り替え
	game.SetHasTitleSelection(true)
	if !game.hasTitleSelection {
		t.Error("expected hasTitleSelection to be true after setting to true")
	}

	game.SetHasTitleSelection(false)
	if game.hasTitleSelection {
		t.Error("expected hasTitleSelection to be false after setting to false")
	}
}

// TestSetOnTitleExit tests the SetOnTitleExit method
// Requirements: 2.2, 2.3, 2.4, 4.1, 4.2, 4.3
func TestSetOnTitleExit(t *testing.T) {
	game := NewGame(ModeDesktop, nil, 0)

	// 初期状態はnil
	if game.onTitleExit != nil {
		t.Error("expected onTitleExit to be nil initially")
	}

	// コールバックを設定
	callbackCalled := false
	callback := func() error {
		callbackCalled = true
		return nil
	}

	game.SetOnTitleExit(callback)

	// コールバックが設定されていることを確認
	if game.onTitleExit == nil {
		t.Fatal("expected onTitleExit to be set")
	}

	// コールバックを呼び出して動作確認
	err := game.onTitleExit()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if !callbackCalled {
		t.Error("expected callback to be called")
	}
}

// TestSetOnTitleExit_WithError tests SetOnTitleExit with a callback that returns an error
// Requirements: 4.4
func TestSetOnTitleExit_WithError(t *testing.T) {
	game := NewGame(ModeDesktop, nil, 0)

	// エラーを返すコールバックを設定
	expectedError := "cleanup error"
	callback := func() error {
		return &testError{msg: expectedError}
	}

	game.SetOnTitleExit(callback)

	// コールバックを呼び出してエラーを確認
	err := game.onTitleExit()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != expectedError {
		t.Errorf("expected error '%s', got '%s'", expectedError, err.Error())
	}
}

// TestSetOnTitleExit_Replace tests replacing the onTitleExit callback
// Requirements: 2.2, 2.3, 2.4
func TestSetOnTitleExit_Replace(t *testing.T) {
	game := NewGame(ModeDesktop, nil, 0)

	// 最初のコールバック
	firstCallbackCalled := false
	firstCallback := func() error {
		firstCallbackCalled = true
		return nil
	}

	// 2番目のコールバック
	secondCallbackCalled := false
	secondCallback := func() error {
		secondCallbackCalled = true
		return nil
	}

	// 最初のコールバックを設定
	game.SetOnTitleExit(firstCallback)

	// 2番目のコールバックで置き換え
	game.SetOnTitleExit(secondCallback)

	// コールバックを呼び出し
	err := game.onTitleExit()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// 2番目のコールバックのみが呼び出されることを確認
	if firstCallbackCalled {
		t.Error("expected first callback NOT to be called")
	}

	if !secondCallbackCalled {
		t.Error("expected second callback to be called")
	}
}

// TestSetOnTitleExit_Nil tests setting onTitleExit to nil
// Requirements: 2.2, 2.3, 2.4
func TestSetOnTitleExit_Nil(t *testing.T) {
	game := NewGame(ModeDesktop, nil, 0)

	// コールバックを設定
	callback := func() error {
		return nil
	}
	game.SetOnTitleExit(callback)

	// nilで上書き
	game.SetOnTitleExit(nil)

	// nilになっていることを確認
	if game.onTitleExit != nil {
		t.Error("expected onTitleExit to be nil after setting to nil")
	}
}

// testError is a simple error type for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

// TestReturnToSelection_PreservesTitles tests that returnToSelection() does not modify the titles list
// Requirements: 2.6
func TestReturnToSelection_PreservesTitles(t *testing.T) {
	// 複数のタイトルを設定
	titles := []title.FillyTitle{
		{Name: "Title1", Path: "/path/1", IsEmbedded: false},
		{Name: "Title2", Path: "/path/2", IsEmbedded: true},
		{Name: "Title3", Path: "/path/3", IsEmbedded: false},
	}

	game := NewGame(ModeDesktop, titles, 0)
	game.SetHasTitleSelection(true)

	// 元のタイトル一覧を保存
	originalTitles := make([]title.FillyTitle, len(game.titles))
	copy(originalTitles, game.titles)

	// returnToSelection を呼び出し
	err := game.returnToSelection()
	if err != nil {
		t.Fatalf("returnToSelection failed: %v", err)
	}

	// タイトル一覧が変更されていないことを確認
	if len(game.titles) != len(originalTitles) {
		t.Errorf("expected %d titles, got %d", len(originalTitles), len(game.titles))
	}

	for i, original := range originalTitles {
		if game.titles[i].Name != original.Name {
			t.Errorf("title[%d].Name: expected %s, got %s", i, original.Name, game.titles[i].Name)
		}
		if game.titles[i].Path != original.Path {
			t.Errorf("title[%d].Path: expected %s, got %s", i, original.Path, game.titles[i].Path)
		}
		if game.titles[i].IsEmbedded != original.IsEmbedded {
			t.Errorf("title[%d].IsEmbedded: expected %v, got %v", i, original.IsEmbedded, game.titles[i].IsEmbedded)
		}
	}
}

// TestReturnToSelection_PreservesSelectedIndex tests that returnToSelection() does not modify the selectedIndex
// Requirements: 5.2
func TestReturnToSelection_PreservesSelectedIndex(t *testing.T) {
	titles := []title.FillyTitle{
		{Name: "Title1", Path: "/path/1", IsEmbedded: false},
		{Name: "Title2", Path: "/path/2", IsEmbedded: false},
		{Name: "Title3", Path: "/path/3", IsEmbedded: false},
	}

	testCases := []struct {
		name          string
		selectedIndex int
	}{
		{"First title selected", 0},
		{"Middle title selected", 1},
		{"Last title selected", 2},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			game := NewGame(ModeDesktop, titles, 0)
			game.SetHasTitleSelection(true)
			game.selectedIndex = tc.selectedIndex

			// returnToSelection を呼び出し
			err := game.returnToSelection()
			if err != nil {
				t.Fatalf("returnToSelection failed: %v", err)
			}

			// selectedIndex が変更されていないことを確認
			if game.selectedIndex != tc.selectedIndex {
				t.Errorf("expected selectedIndex %d, got %d", tc.selectedIndex, game.selectedIndex)
			}
		})
	}
}

// TestReturnToSelection_PreservesTitlesAndSelectedIndex tests that both titles and selectedIndex are preserved
// Requirements: 2.6, 5.2
func TestReturnToSelection_PreservesTitlesAndSelectedIndex(t *testing.T) {
	titles := []title.FillyTitle{
		{Name: "Title1", Path: "/path/1", IsEmbedded: false},
		{Name: "Title2", Path: "/path/2", IsEmbedded: true},
	}

	game := NewGame(ModeDesktop, titles, 0)
	game.SetHasTitleSelection(true)
	game.selectedIndex = 1 // 2番目のタイトルを選択

	// 元の状態を保存
	originalTitlesCount := len(game.titles)
	originalSelectedIndex := game.selectedIndex

	// returnToSelection を呼び出し
	err := game.returnToSelection()
	if err != nil {
		t.Fatalf("returnToSelection failed: %v", err)
	}

	// モードが ModeSelection に変更されていることを確認
	if game.mode != ModeSelection {
		t.Errorf("expected mode ModeSelection, got %v", game.mode)
	}

	// タイトル一覧が保持されていることを確認
	if len(game.titles) != originalTitlesCount {
		t.Errorf("expected %d titles, got %d", originalTitlesCount, len(game.titles))
	}

	// selectedIndex が保持されていることを確認
	if game.selectedIndex != originalSelectedIndex {
		t.Errorf("expected selectedIndex %d, got %d", originalSelectedIndex, game.selectedIndex)
	}
}

// TestReturnToSelection_NilError_NoErrorLog tests that when onTitleExit callback
// returns nil, no error log is produced and mode transitions to ModeSelection.
// Requirements: 1.3
func TestReturnToSelection_NilError_NoErrorLog(t *testing.T) {
	// Set up a buffer to capture log output
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelError,
	})
	originalDefault := slog.Default()
	slog.SetDefault(slog.New(handler))
	defer slog.SetDefault(originalDefault)

	titles := []title.FillyTitle{
		{Name: "Title1", Path: "/path/1", IsEmbedded: false},
	}

	game := NewGame(ModeDesktop, titles, 0)
	game.SetHasTitleSelection(true)

	// コールバックがnilを返す（エラーなし）
	callbackCalled := false
	game.SetOnTitleExit(func() error {
		callbackCalled = true
		return nil
	})

	// returnToSelection を呼び出し
	err := game.returnToSelection()
	if err != nil {
		t.Fatalf("returnToSelection failed: %v", err)
	}

	// コールバックが呼び出されたことを確認
	if !callbackCalled {
		t.Error("expected onTitleExit callback to be called")
	}

	// エラーログが出力されていないことを確認
	logOutput := buf.String()
	if strings.Contains(logOutput, "level=ERROR") {
		t.Errorf("expected no error log output, but got: %s", logOutput)
	}

	// モードがModeSelectionに遷移していることを確認
	if game.mode != ModeSelection {
		t.Errorf("expected mode ModeSelection, got %v", game.mode)
	}
}

// TestReturnToSelection_NilCallback_NoPanic tests that when onTitleExit callback
// is not set (nil), no panic occurs and mode transitions to ModeSelection.
// Requirements: 1.4
func TestReturnToSelection_NilCallback_NoPanic(t *testing.T) {
	// Set up a buffer to capture log output
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelError,
	})
	originalDefault := slog.Default()
	slog.SetDefault(slog.New(handler))
	defer slog.SetDefault(originalDefault)

	titles := []title.FillyTitle{
		{Name: "Title1", Path: "/path/1", IsEmbedded: false},
	}

	game := NewGame(ModeDesktop, titles, 0)
	game.SetHasTitleSelection(true)
	// onTitleExit は設定しない（nilのまま）

	// returnToSelection を呼び出し — パニックしないことを確認
	err := game.returnToSelection()
	if err != nil {
		t.Fatalf("returnToSelection failed: %v", err)
	}

	// ログ出力がないことを確認
	logOutput := buf.String()
	if strings.Contains(logOutput, "level=ERROR") {
		t.Errorf("expected no log output when callback is nil, but got: %s", logOutput)
	}

	// モードがModeSelectionに遷移していることを確認
	if game.mode != ModeSelection {
		t.Errorf("expected mode ModeSelection, got %v", game.mode)
	}
}

