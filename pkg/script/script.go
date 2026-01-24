package script

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// Script はスクリプトファイルを表す
type Script struct {
	FileName string // ファイル名
	Content  string // UTF-8に変換された内容
	Size     int64  // ファイルサイズ
}

// Loader はスクリプトファイルの読み込みを行う
type Loader struct {
	titlePath string
}

// NewLoader Loaderを作成
func NewLoader(titlePath string) *Loader {
	return &Loader{
		titlePath: titlePath,
	}
}

// LoadAllScripts すべての.TFYファイルを読み込む
func (l *Loader) LoadAllScripts() ([]Script, error) {
	scriptFiles, err := l.findScriptFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to find script files: %w", err)
	}

	if len(scriptFiles) == 0 {
		return nil, fmt.Errorf("no script files found in %s", l.titlePath)
	}

	var scripts []Script
	for _, filePath := range scriptFiles {
		script, err := l.loadScript(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to load script %s: %w", filePath, err)
		}
		scripts = append(scripts, *script)
	}

	return scripts, nil
}

// findScriptFiles .TFYファイルを検出（case-insensitive）
func (l *Loader) findScriptFiles() ([]string, error) {
	var scriptFiles []string

	err := filepath.Walk(l.titlePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// 拡張子をcase-insensitiveで比較
		ext := filepath.Ext(path)
		if strings.EqualFold(ext, ".tfy") {
			scriptFiles = append(scriptFiles, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return scriptFiles, nil
}

// loadScript 単一のスクリプトファイルを読み込む
func (l *Loader) loadScript(path string) (*Script, error) {
	// ファイル情報を取得
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// ファイルを読み込む
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Shift-JISからUTF-8に変換
	content, err := convertShiftJISToUTF8(data)
	if err != nil {
		return nil, fmt.Errorf("failed to convert encoding: %w", err)
	}

	return &Script{
		FileName: filepath.Base(path),
		Content:  content,
		Size:     info.Size(),
	}, nil
}

// convertShiftJISToUTF8 Shift-JISからUTF-8に変換
func convertShiftJISToUTF8(data []byte) (string, error) {
	// Shift-JISデコーダーを作成
	decoder := japanese.ShiftJIS.NewDecoder()
	reader := transform.NewReader(strings.NewReader(string(data)), decoder)

	// UTF-8に変換
	utf8Data, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("failed to decode Shift-JIS: %w", err)
	}

	return string(utf8Data), nil
}
