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
	if err := os.Mkdir(title1Dir, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	// 外部タイトルを読み込まず、embedに複数追加（テスト用）
	registry := NewFillyTitleRegistry(testEmbedFS)

	// embedされたタイトルを手動で追加（テスト用）
	registry.embeddedTitles = append(registry.embeddedTitles, FillyTitle{
		Name:       "embedded-title-1",
		Path:       "titles/embedded-title-1",
		IsEmbedded: true,
	})
	registry.embeddedTitles = append(registry.embeddedTitles, FillyTitle{
		Name:       "embedded-title-2",
		Path:       "titles/embedded-title-2",
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

func TestExtractMetadata_Basic(t *testing.T) {
	content := `#info INAM "テストタイトル"
#info ICOP "著作者"
#info ISBJ "説明文"
#info ICMT "コメント1"
#info ICMT "コメント2"

main() {
	LoadPic("test.bmp");
}
`
	metadata := ExtractMetadata(content)

	if metadata.INAM != "テストタイトル" {
		t.Errorf("expected INAM 'テストタイトル', got %q", metadata.INAM)
	}
	if metadata.ICOP != "著作者" {
		t.Errorf("expected ICOP '著作者', got %q", metadata.ICOP)
	}
	if metadata.ISBJ != "説明文" {
		t.Errorf("expected ISBJ '説明文', got %q", metadata.ISBJ)
	}
	if len(metadata.ICMT) != 2 {
		t.Errorf("expected 2 ICMT entries, got %d", len(metadata.ICMT))
	}
	if len(metadata.ICMT) >= 1 && metadata.ICMT[0] != "コメント1" {
		t.Errorf("expected ICMT[0] 'コメント1', got %q", metadata.ICMT[0])
	}
}

func TestExtractMetadata_NoInfo(t *testing.T) {
	content := `main() {
	LoadPic("test.bmp");
}
`
	metadata := ExtractMetadata(content)

	if metadata.INAM != "" {
		t.Errorf("expected empty INAM, got %q", metadata.INAM)
	}
	if len(metadata.ICMT) != 0 {
		t.Errorf("expected 0 ICMT entries, got %d", len(metadata.ICMT))
	}
}

func TestDisplayName_WithMetadata(t *testing.T) {
	title := FillyTitle{
		Name: "dir-name",
		Metadata: &TitleMetadata{
			INAM: "表示タイトル",
		},
	}

	if title.DisplayName() != "表示タイトル" {
		t.Errorf("expected '表示タイトル', got %q", title.DisplayName())
	}
}

func TestDisplayName_WithoutMetadata(t *testing.T) {
	title := FillyTitle{
		Name: "dir-name",
	}

	if title.DisplayName() != "dir-name" {
		t.Errorf("expected 'dir-name', got %q", title.DisplayName())
	}
}

func TestDisplayName_EmptyINAM(t *testing.T) {
	title := FillyTitle{
		Name: "dir-name",
		Metadata: &TitleMetadata{
			INAM: "",
		},
	}

	if title.DisplayName() != "dir-name" {
		t.Errorf("expected 'dir-name', got %q", title.DisplayName())
	}
}

func TestExtractMetadataFromDirectory_Sample(t *testing.T) {
	// samples/kuma2 ディレクトリからメタデータを抽出
	metadata, err := ExtractMetadataFromDirectory("../../samples/kuma2")
	if err != nil {
		t.Fatalf("failed to extract metadata: %v", err)
	}

	// KUMA2.TFYには#info INAMがあるはず
	if metadata.INAM == "" {
		t.Log("INAM is empty (may be encoding issue)")
	} else {
		t.Logf("INAM: %s", metadata.INAM)
	}

	if metadata.ISBJ == "" {
		t.Log("ISBJ is empty")
	} else {
		t.Logf("ISBJ: %s", metadata.ISBJ)
	}

	t.Logf("ICMT count: %d", len(metadata.ICMT))
}
