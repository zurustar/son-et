package title

import (
	"embed"
	"os"
	"path/filepath"
	"testing"
)

//go:embed testdata
var testEmbedFS embed.FS

func TestNewFillyTitleRegistry(t *testing.T) {
	registry := NewFillyTitleRegistry(testEmbedFS)
	if registry == nil {
		t.Fatal("NewFillyTitleRegistry returned nil")
	}

	// testdataディレクトリにはtitlesディレクトリがないので、embeddedTitlesは空
	if len(registry.embeddedTitles) != 0 {
		t.Errorf("expected 0 embedded titles, got %d", len(registry.embeddedTitles))
	}
}

func TestLoadExternalTitle_ValidDirectory(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tmpDir := t.TempDir()
	titleDir := filepath.Join(tmpDir, "test-title")
	if err := os.Mkdir(titleDir, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	registry := NewFillyTitleRegistry(testEmbedFS)
	err := registry.LoadExternalTitle(titleDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if registry.externalTitle == nil {
		t.Fatal("externalTitle is nil")
	}

	if registry.externalTitle.Name != "test-title" {
		t.Errorf("expected name 'test-title', got %q", registry.externalTitle.Name)
	}

	if registry.externalTitle.IsEmbedded {
		t.Error("external title should not be marked as embedded")
	}
}

func TestLoadExternalTitle_NonExistentDirectory(t *testing.T) {
	registry := NewFillyTitleRegistry(testEmbedFS)
	err := registry.LoadExternalTitle("/nonexistent/path")
	if err == nil {
		t.Error("expected error for nonexistent directory, got nil")
	}
}

func TestLoadExternalTitle_NotADirectory(t *testing.T) {
	// テスト用の一時ファイルを作成
	tmpFile, err := os.CreateTemp("", "test-file-*")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	registry := NewFillyTitleRegistry(testEmbedFS)
	err = registry.LoadExternalTitle(tmpFile.Name())
	if err == nil {
		t.Error("expected error for file path, got nil")
	}
}

func TestGetAvailableTitles_NoTitles(t *testing.T) {
	registry := NewFillyTitleRegistry(testEmbedFS)
	titles := registry.GetAvailableTitles()

	if len(titles) != 0 {
		t.Errorf("expected 0 titles, got %d", len(titles))
	}
}

func TestGetAvailableTitles_ExternalOnly(t *testing.T) {
	tmpDir := t.TempDir()
	titleDir := filepath.Join(tmpDir, "test-title")
	if err := os.Mkdir(titleDir, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	registry := NewFillyTitleRegistry(testEmbedFS)
	if err := registry.LoadExternalTitle(titleDir); err != nil {
		t.Fatalf("failed to load external title: %v", err)
	}

	titles := registry.GetAvailableTitles()
	if len(titles) != 1 {
		t.Errorf("expected 1 title, got %d", len(titles))
	}

	if titles[0].Name != "test-title" {
		t.Errorf("expected name 'test-title', got %q", titles[0].Name)
	}
}

func TestSelectTitle_NoTitles(t *testing.T) {
	registry := NewFillyTitleRegistry(testEmbedFS)
	_, _, err := registry.SelectTitle()
	if err == nil {
		t.Error("expected error when no titles available, got nil")
	}
}

func TestSelectTitle_SingleTitle(t *testing.T) {
	tmpDir := t.TempDir()
	titleDir := filepath.Join(tmpDir, "test-title")
	if err := os.Mkdir(titleDir, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	registry := NewFillyTitleRegistry(testEmbedFS)
	if err := registry.LoadExternalTitle(titleDir); err != nil {
		t.Fatalf("failed to load external title: %v", err)
	}

	title, needsSelection, err := registry.SelectTitle()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if needsSelection {
		t.Error("selection screen should not be needed for single title")
	}

	if title == nil {
		t.Fatal("selected title is nil")
	}

	if title.Name != "test-title" {
		t.Errorf("expected name 'test-title', got %q", title.Name)
	}
}

func TestSelectTitle_MultipleTitles(t *testing.T) {
	tmpDir := t.TempDir()

	// 複数のタイトルディレクトリを作成
	title1Dir := filepath.Join(tmpDir, "title1")
	title2Dir := filepath.Join(tmpDir, "title2")
	if err := os.Mkdir(title1Dir, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}
	if err := os.Mkdir(title2Dir, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	// 外部タイトルを1つ読み込み、embedに1つ追加（手動で）
	registry := NewFillyTitleRegistry(testEmbedFS)
	if err := registry.LoadExternalTitle(title1Dir); err != nil {
		t.Fatalf("failed to load external title: %v", err)
	}

	// embedされたタイトルを手動で追加（テスト用）
	registry.embeddedTitles = append(registry.embeddedTitles, FillyTitle{
		Name:       "embedded-title",
		Path:       "titles/embedded-title",
		IsEmbedded: true,
	})

	title, needsSelection, err := registry.SelectTitle()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !needsSelection {
		t.Error("selection screen should be needed for multiple titles")
	}

	if title != nil {
		t.Error("title should be nil when selection is needed")
	}
}
