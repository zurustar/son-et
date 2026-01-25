package title

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/zurustar/son-et/pkg/compiler/lexer"
	"github.com/zurustar/son-et/pkg/script"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// TitleConfig はtitle.jsonの構造
type TitleConfig struct {
	EntryFile string `json:"entryFile"`
}

// FillyTitle はFILLYタイトルを表す
type FillyTitle struct {
	Name       string         // タイトル名（ディレクトリ名）
	Path       string         // タイトルのパス（embedの場合は仮想パス）
	IsEmbedded bool           // embedされたタイトルかどうか
	Metadata   *TitleMetadata // #infoから抽出したメタデータ
	EntryFile  string         // エントリーポイントファイル名（空の場合は自動検出）
}

// TitleMetadata は#infoディレクティブから抽出したメタデータ
type TitleMetadata struct {
	INAM string   // タイトル名
	ICOP string   // 著作権情報
	ISBJ string   // サブジェクト（説明）
	IART string   // アーティスト
	ICMT []string // コメント（複数行可）
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
			title := FillyTitle{
				Name:       entry.Name(),
				Path:       titlePath,
				IsEmbedded: true,
			}
			// embedされたタイトルのメタデータ抽出
			title.Metadata = r.extractEmbeddedMetadata(titlePath)
			// title.jsonからエントリーポイント読み込み
			title.EntryFile = r.loadEmbeddedTitleConfig(titlePath)
			r.embeddedTitles = append(r.embeddedTitles, title)
		}
	}
}

// loadEmbeddedTitleConfig はembedされたタイトルのtitle.jsonを読み込む
func (r *FillyTitleRegistry) loadEmbeddedTitleConfig(titlePath string) string {
	configPath := filepath.Join(titlePath, "title.json")
	data, err := fs.ReadFile(r.embedFS, configPath)
	if err != nil {
		return "" // title.jsonが存在しない場合は空文字列
	}

	var config TitleConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return "" // パースエラーの場合も空文字列
	}

	return config.EntryFile
}

// extractEmbeddedMetadata はembedされたタイトルからメタデータを抽出する
func (r *FillyTitleRegistry) extractEmbeddedMetadata(titlePath string) *TitleMetadata {
	metadata := &TitleMetadata{
		ICMT: []string{},
	}

	// TFYファイルを探して読み込む
	entries, err := fs.ReadDir(r.embedFS, titlePath)
	if err != nil {
		return metadata
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := strings.ToUpper(entry.Name())
		if !strings.HasSuffix(name, ".TFY") {
			continue
		}

		filePath := filepath.Join(titlePath, entry.Name())
		data, err := fs.ReadFile(r.embedFS, filePath)
		if err != nil {
			continue
		}

		// Shift-JIS to UTF-8 conversion
		content := convertToUTF8(data)
		meta := ExtractMetadata(content)

		// マージ
		if metadata.INAM == "" && meta.INAM != "" {
			metadata.INAM = meta.INAM
		}
		if metadata.ICOP == "" && meta.ICOP != "" {
			metadata.ICOP = meta.ICOP
		}
		if metadata.ISBJ == "" && meta.ISBJ != "" {
			metadata.ISBJ = meta.ISBJ
		}
		if metadata.IART == "" && meta.IART != "" {
			metadata.IART = meta.IART
		}
		metadata.ICMT = append(metadata.ICMT, meta.ICMT...)
	}

	return metadata
}

// convertToUTF8 はShift-JISからUTF-8に変換する
func convertToUTF8(data []byte) string {
	decoder := japanese.ShiftJIS.NewDecoder()
	reader := transform.NewReader(strings.NewReader(string(data)), decoder)
	utf8Data, err := io.ReadAll(reader)
	if err != nil {
		// 変換に失敗した場合はそのまま返す
		return string(data)
	}
	return string(utf8Data)
}

// LoadExternalTitle 外部ディレクトリからタイトルを読み込む
func (r *FillyTitleRegistry) LoadExternalTitle(path string) error {
	return r.LoadExternalTitleWithEntry(path, "")
}

// LoadExternalTitleWithEntry 外部ディレクトリからタイトルを読み込む（エントリーファイル指定付き）
func (r *FillyTitleRegistry) LoadExternalTitleWithEntry(path string, entryFile string) error {
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

	// メタデータを抽出
	metadata, _ := ExtractMetadataFromDirectory(absPath)

	// エントリーファイルの決定
	// 1. 引数で指定されていればそれを使用
	// 2. title.jsonがあればそれを使用
	// 3. どちらもなければ空（自動検出）
	finalEntryFile := entryFile
	if finalEntryFile == "" {
		finalEntryFile = loadTitleConfig(absPath)
	}

	r.externalTitle = &FillyTitle{
		Name:       filepath.Base(absPath),
		Path:       absPath,
		IsEmbedded: false,
		Metadata:   metadata,
		EntryFile:  finalEntryFile,
	}

	return nil
}

// loadTitleConfig は外部タイトルのtitle.jsonを読み込む
func loadTitleConfig(dirPath string) string {
	configPath := filepath.Join(dirPath, "title.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return "" // title.jsonが存在しない場合は空文字列
	}

	var config TitleConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return "" // パースエラーの場合も空文字列
	}

	return config.EntryFile
}

// GetAvailableTitles 利用可能なタイトル一覧を取得
func (r *FillyTitleRegistry) GetAvailableTitles() []FillyTitle {
	var titles []FillyTitle

	// 外部タイトルが指定されている場合はそれのみを返す
	if r.externalTitle != nil {
		titles = append(titles, *r.externalTitle)
		return titles
	}

	// 外部タイトルがない場合はembedされたタイトルを返す
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

// ExtractMetadata はTFYファイルから#infoメタデータを抽出する
// フルコンパイルせずにLexerのみを使用して軽量に抽出する
func ExtractMetadata(content string) *TitleMetadata {
	metadata := &TitleMetadata{
		ICMT: []string{},
	}

	l := lexer.New(content)
	for {
		tok := l.NextToken()
		if tok.Type == lexer.TOKEN_EOF {
			break
		}

		if tok.Type == lexer.TOKEN_INFO {
			// Literal format: "#info KEY value" or "#info KEY \"value\""
			parseInfoDirective(tok.Literal, metadata)
		}
	}

	return metadata
}

// parseInfoDirective は#infoディレクティブをパースしてメタデータに追加する
func parseInfoDirective(literal string, metadata *TitleMetadata) {
	// Remove "#info " prefix
	if !strings.HasPrefix(literal, "#info ") {
		return
	}
	rest := strings.TrimPrefix(literal, "#info ")
	rest = strings.TrimSpace(rest)

	// Split into key and value
	parts := strings.SplitN(rest, " ", 2)
	if len(parts) < 2 {
		return
	}

	key := strings.ToUpper(strings.TrimSpace(parts[0]))
	value := strings.TrimSpace(parts[1])

	// Remove quotes if present
	if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
		value = value[1 : len(value)-1]
	}

	switch key {
	case "INAM":
		metadata.INAM = value
	case "ICOP":
		metadata.ICOP = value
	case "ISBJ":
		metadata.ISBJ = value
	case "IART":
		metadata.IART = value
	case "ICMT":
		metadata.ICMT = append(metadata.ICMT, value)
	}
}

// ExtractMetadataFromDirectory はディレクトリ内のTFYファイルからメタデータを抽出する
func ExtractMetadataFromDirectory(dirPath string) (*TitleMetadata, error) {
	loader := script.NewLoader(dirPath)
	scripts, err := loader.LoadAllScripts()
	if err != nil {
		return nil, fmt.Errorf("failed to load scripts: %w", err)
	}

	// 全スクリプトからメタデータを収集（最初に見つかったものを優先）
	combined := &TitleMetadata{
		ICMT: []string{},
	}

	for _, s := range scripts {
		meta := ExtractMetadata(s.Content)
		if combined.INAM == "" && meta.INAM != "" {
			combined.INAM = meta.INAM
		}
		if combined.ICOP == "" && meta.ICOP != "" {
			combined.ICOP = meta.ICOP
		}
		if combined.ISBJ == "" && meta.ISBJ != "" {
			combined.ISBJ = meta.ISBJ
		}
		if combined.IART == "" && meta.IART != "" {
			combined.IART = meta.IART
		}
		combined.ICMT = append(combined.ICMT, meta.ICMT...)
	}

	return combined, nil
}

// DisplayName はタイトルの表示名を返す
// INAMがあればそれを、なければディレクトリ名を返す
func (t *FillyTitle) DisplayName() string {
	if t.Metadata != nil && t.Metadata.INAM != "" {
		return t.Metadata.INAM
	}
	return t.Name
}
