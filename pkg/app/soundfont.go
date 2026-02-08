package app

import (
	"embed"
	"os"
	"path/filepath"

	"github.com/zurustar/son-et/pkg/fileutil"
)

// SoundFontLocation represents the location of a SoundFont file.
type SoundFontLocation struct {
	// Path is the path to the SoundFont file
	Path string
	// FileSystem is the FileSystem to use for loading (nil for external files)
	FileSystem fileutil.FileSystem
	// IsEmbedded indicates whether the SoundFont is embedded
	IsEmbedded bool
}

// DefaultSoundFontName is the default SoundFont filename to search for.
const DefaultSoundFontName = "GeneralUser-GS.sf2"

// findSoundFont searches for a SoundFont file in the following order:
// 1. Embedded soundfonts directory
// 2. Embedded title directory
// 3. Current directory (external)
// 4. Title path (external)
//
// Requirement 3.1: Application SHALL 以下の優先順位でSF2ファイルを検索する
// Requirement 3.2: 埋め込みSF2ファイルが見つかる場合は埋め込みファイルを優先して使用
// Requirement 3.3: 埋め込みSF2ファイルが見つからない場合は外部ファイルシステムから検索
//
// Parameters:
//   - embedFS: The embedded file system
//   - titlePath: Path to the title directory
//   - isEmbedded: Whether the title is embedded
//
// Returns:
//   - *SoundFontLocation: Location of the SoundFont file, or nil if not found
func findSoundFont(embedFS embed.FS, titlePath string, isEmbedded bool) *SoundFontLocation {
	// 1. Check embedded soundfonts directory
	// Requirement 3.1: First priority - embedded soundfonts directory
	soundfontsPath := "soundfonts/" + DefaultSoundFontName
	if data, err := embedFS.ReadFile(soundfontsPath); err == nil && len(data) > 0 {
		return &SoundFontLocation{
			Path:       DefaultSoundFontName, // FileSystemのベースパスが"soundfonts"なので、ファイル名だけ
			FileSystem: fileutil.NewEmbedFS(embedFS, "soundfonts"),
			IsEmbedded: true,
		}
	}

	// 2. Check embedded title directory
	// Requirement 3.1: Second priority - embedded title directory
	if isEmbedded && titlePath != "" {
		titleSFPath := titlePath + "/" + DefaultSoundFontName
		if data, err := embedFS.ReadFile(titleSFPath); err == nil && len(data) > 0 {
			return &SoundFontLocation{
				Path:       DefaultSoundFontName,
				FileSystem: fileutil.NewEmbedFS(embedFS, titlePath),
				IsEmbedded: true,
			}
		}
	}

	// 3. Check current directory (external)
	// Requirement 3.3: Third priority - external current directory
	if _, err := os.Stat(DefaultSoundFontName); err == nil {
		return &SoundFontLocation{
			Path:       DefaultSoundFontName,
			FileSystem: nil,
			IsEmbedded: false,
		}
	}

	// 4. Check title path (external)
	// Requirement 3.3: Fourth priority - external title directory
	if titlePath != "" {
		titleSFPath := filepath.Join(titlePath, DefaultSoundFontName)
		if _, err := os.Stat(titleSFPath); err == nil {
			return &SoundFontLocation{
				Path:       titleSFPath,
				FileSystem: nil,
				IsEmbedded: false,
			}
		}
	}

	// Requirement 3.4: No SoundFont found
	return nil
}
