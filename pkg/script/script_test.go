package script

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

func TestNewLoader(t *testing.T) {
	loader := NewLoader("/test/path")
	if loader == nil {
		t.Fatal("NewLoader returned nil")
	}
	if loader.fs.BasePath() != "/test/path" {
		t.Errorf("expected basePath '/test/path', got %q", loader.fs.BasePath())
	}
}

func TestFindScriptFiles_CaseInsensitive(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tmpDir := t.TempDir()

	// 様々な大文字小文字の.TFYファイルを作成
	testFiles := []string{
		"test.tfy",
		"script.TFY",
		"helper.Tfy",
		"other.txt", // これは検出されないはず
	}

	for _, filename := range testFiles {
		filePath := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(filePath, []byte("test content"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	loader := NewLoader(tmpDir)
	scriptFiles, err := loader.findScriptFiles()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// .TFYファイルは3つ検出されるはず
	if len(scriptFiles) != 3 {
		t.Errorf("expected 3 script files, got %d", len(scriptFiles))
	}

	// other.txtが含まれていないことを確認
	for _, file := range scriptFiles {
		if filepath.Base(file) == "other.txt" {
			t.Error("other.txt should not be detected as a script file")
		}
	}
}

func TestLoadScript_UTF8(t *testing.T) {
	// UTF-8のテストファイルを作成（ASCII文字のみ）
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.tfy")
	testContent := "Hello World\nThis is a test"

	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	loader := NewLoader(tmpDir)
	// loadScriptは相対パスを期待するので、ファイル名のみを渡す
	script, err := loader.loadScript("test.tfy")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if script.FileName != "test.tfy" {
		t.Errorf("expected filename 'test.tfy', got %q", script.FileName)
	}

	if script.Content != testContent {
		t.Errorf("content mismatch:\nexpected: %q\ngot: %q", testContent, script.Content)
	}

	if script.Size == 0 {
		t.Error("script size should not be 0")
	}
}

func TestLoadScript_ShiftJIS(t *testing.T) {
	// Shift-JISのテストファイルを作成
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.tfy")
	testContent := "これはShift-JISのテストです"

	// UTF-8からShift-JISに変換
	encoder := japanese.ShiftJIS.NewEncoder()
	shiftJISContent, _, err := transform.String(encoder, testContent)
	if err != nil {
		t.Fatalf("failed to encode to Shift-JIS: %v", err)
	}

	if err := os.WriteFile(testFile, []byte(shiftJISContent), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	loader := NewLoader(tmpDir)
	// loadScriptは相対パスを期待するので、ファイル名のみを渡す
	script, err := loader.loadScript("test.tfy")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if script.Content != testContent {
		t.Errorf("content mismatch:\nexpected: %q\ngot: %q", testContent, script.Content)
	}
}

func TestLoadAllScripts(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tmpDir := t.TempDir()

	// 複数のスクリプトファイルを作成
	testFiles := map[string]string{
		"main.tfy":   "main script",
		"sub.TFY":    "sub script",
		"helper.Tfy": "helper script",
	}

	for filename, content := range testFiles {
		filePath := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	loader := NewLoader(tmpDir)
	scripts, err := loader.LoadAllScripts()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(scripts) != 3 {
		t.Errorf("expected 3 scripts, got %d", len(scripts))
	}

	// すべてのスクリプトが読み込まれていることを確認
	foundFiles := make(map[string]bool)
	for _, script := range scripts {
		foundFiles[script.FileName] = true
	}

	for filename := range testFiles {
		if !foundFiles[filename] {
			t.Errorf("script file %q was not loaded", filename)
		}
	}
}

func TestLoadAllScripts_NoScripts(t *testing.T) {
	// スクリプトファイルがないディレクトリ
	tmpDir := t.TempDir()

	loader := NewLoader(tmpDir)
	_, err := loader.LoadAllScripts()
	if err == nil {
		t.Error("expected error when no script files found, got nil")
	}
}

func TestLoadAllScripts_NonExistentDirectory(t *testing.T) {
	loader := NewLoader("/nonexistent/path")
	_, err := loader.LoadAllScripts()
	if err == nil {
		t.Error("expected error for nonexistent directory, got nil")
	}
}

func TestConvertShiftJISToUTF8(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "日本語テキスト",
			input:   "こんにちは世界",
			wantErr: false,
		},
		{
			name:    "英数字",
			input:   "Hello World 123",
			wantErr: false,
		},
		{
			name:    "混在",
			input:   "Hello こんにちは 123",
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// UTF-8からShift-JISに変換
			encoder := japanese.ShiftJIS.NewEncoder()
			shiftJISData, _, err := transform.String(encoder, tc.input)
			if err != nil {
				t.Fatalf("failed to encode to Shift-JIS: %v", err)
			}

			// Shift-JISからUTF-8に変換
			result, err := convertShiftJISToUTF8([]byte(shiftJISData))
			if (err != nil) != tc.wantErr {
				t.Errorf("convertShiftJISToUTF8() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if !tc.wantErr && result != tc.input {
				t.Errorf("convertShiftJISToUTF8() = %q, want %q", result, tc.input)
			}
		})
	}
}
