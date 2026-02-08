// Package audio provides audio-related components for the FILLY virtual machine.
// This file implements SoundFont loading utilities with FileSystem support.
package audio

import (
	"bytes"
	"fmt"
	"os"

	"github.com/sinshu/go-meltysynth/meltysynth"
	"github.com/zurustar/son-et/pkg/fileutil"
)

// ReadSoundFontFS reads a SoundFont file using the FileSystem interface.
// This supports both real file system and embedded file system.
//
// Requirement 2.1: SoundFont_Loader SHALL FileSystemインターフェースを使用してSF2ファイルを読み込む
// Requirement 2.2: FileSystemが設定されていない場合は従来通り os.ReadFile を使用
// Requirement 2.3: FileSystemが設定されている場合は FileSystem.ReadFile を使用
//
// Parameters:
//   - fs: The FileSystem interface to use for file access (can be nil for regular file system)
//   - path: Path to the SoundFont (.sf2) file
//
// Returns:
//   - []byte: The SoundFont file contents
//   - error: Error if the file cannot be read
func ReadSoundFontFS(fs fileutil.FileSystem, path string) ([]byte, error) {
	if fs == nil {
		// Requirement 2.2: Fall back to regular file system
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("%w: %s", ErrSoundFontNotFound, path)
			}
			return nil, fmt.Errorf("failed to read SoundFont file: %w", err)
		}
		return data, nil
	}

	// Requirement 2.3: Use FileSystem interface
	data, err := fs.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrSoundFontNotFound, path)
	}
	return data, nil
}

// LoadSoundFontFS loads and parses a SoundFont file using the FileSystem interface.
// This is a convenience function that combines reading and parsing.
//
// Parameters:
//   - fs: The FileSystem interface to use for file access (can be nil for regular file system)
//   - path: Path to the SoundFont (.sf2) file
//
// Returns:
//   - *meltysynth.SoundFont: The parsed SoundFont
//   - error: Error if the file cannot be read or parsed
func LoadSoundFontFS(fs fileutil.FileSystem, path string) (*meltysynth.SoundFont, error) {
	data, err := ReadSoundFontFS(fs, path)
	if err != nil {
		return nil, err
	}

	soundFont, err := meltysynth.NewSoundFont(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to parse SoundFont: %w", err)
	}

	return soundFont, nil
}
