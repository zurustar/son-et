package title

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// FillyTitle はFILLYタイトルを表す
type FillyTitle struct {
	Name       string // タイトル名
	Path       string // タイトルのパス（embedの場合は仮想パス）
	IsEmbedded bool   // embedされたタイトルかどうか
}

// FillyTitleRegistry はFILLYタイトルの管理を行う
type FillyTitleRegistry struct {
	embeddedTitles []FillyTitle // embedされたタイトル一覧
	externalTitle  *FillyTitle  // 外部から指定されたタイトル
	embedFS        embed.FS     // embedされたファイルシステム
}

// NewFillyTitleRegistry FillyTitleRegistryを作成
func NewFillyTitleRegistry(embedFS embed.FS) *FillyTitleRegistry {
	registry := &FillyTitleRegistry{
		embedFS:        embedFS,
		embeddedTitles: []FillyTitle{},
	}

	// embedされたタイトルを検出
	registry.loadEmbeddedTitles()

	return registry
}

// loadEmbeddedTitles embedされたタイトルを検出して読み込む
func (r *FillyTitleRegistry) loadEmbeddedTitles() {
	// titlesディレクトリ内のサブディレクトリを列挙
	entries, err := fs.ReadDir(r.embedFS, "titles")
	if err != nil {
		// titlesディレクトリが存在しない、または読み込めない場合は何もしない
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			titlePath := filepath.Join("titles", entry.Name())
			r.embeddedTitles = append(r.embeddedTitles, FillyTitle{
				Name:       entry.Name(),
				Path:       titlePath,
				IsEmbedded: true,
			})
		}
	}
}

// LoadExternalTitle 外部ディレクトリからタイトルを読み込む
func (r *FillyTitleRegistry) LoadExternalTitle(path string) error {
	// ディレクトリの存在確認
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("title directory does not exist: %s", path)
		}
		return fmt.Errorf("failed to access title directory: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("title path is not a directory: %s", path)
	}

	// 絶対パスに変換
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	r.externalTitle = &FillyTitle{
		Name:       filepath.Base(absPath),
		Path:       absPath,
		IsEmbedded: false,
	}

	return nil
}

// GetAvailableTitles 利用可能なタイトル一覧を取得
func (r *FillyTitleRegistry) GetAvailableTitles() []FillyTitle {
	var titles []FillyTitle

	// 外部タイトルが指定されている場合はそれを優先
	if r.externalTitle != nil {
		titles = append(titles, *r.externalTitle)
	}

	// embedされたタイトルを追加
	titles = append(titles, r.embeddedTitles...)

	return titles
}

// SelectTitle タイトルを選択（単一の場合は自動選択）
// 戻り値: (選択されたタイトル, 選択画面が必要か, エラー)
func (r *FillyTitleRegistry) SelectTitle() (*FillyTitle, bool, error) {
	titles := r.GetAvailableTitles()

	if len(titles) == 0 {
		return nil, false, fmt.Errorf("no FILLY titles available")
	}

	if len(titles) == 1 {
		// 単一のタイトルの場合は自動選択
		return &titles[0], false, nil
	}

	// 複数のタイトルがある場合は選択画面が必要
	return nil, true, nil
}
