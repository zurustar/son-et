package vm

import (
	"bufio"
	"fmt"
	"os"
	"sync"
)

// minHandleID is the minimum handle ID assigned by FileHandleTable.
// Handles start from 1 (not 0) to distinguish from uninitialized values.
const minHandleID = 1

// fileEntry はファイルハンドルテーブルの1エントリ。
type fileEntry struct {
	file   *os.File
	reader *bufio.Reader // StrReadF用。初回StrReadF呼び出し時に遅延初期化。
}

// FileHandleTable は整数ハンドル→*os.Fileのマッピングを管理する。
// ハンドルは1から始まる正の整数。
// Requirement 3.1: 整数ハンドルから*os.Fileへのマッピングを管理する。
type FileHandleTable struct {
	files map[int]*fileEntry
	mu    sync.Mutex
}

// NewFileHandleTable は新しいFileHandleTableを生成して返す。
func NewFileHandleTable() *FileHandleTable {
	return &FileHandleTable{
		files: make(map[int]*fileEntry),
	}
}

// Open はファイルを登録し、未使用の最小整数ハンドル（1以上）を割り当てて返す。
// Requirement 3.2: 未使用の最小整数ハンドル（1以上）を割り当てて返す。
func (fht *FileHandleTable) Open(file *os.File) int {
	fht.mu.Lock()
	defer fht.mu.Unlock()

	handle := minHandleID
	for {
		if _, exists := fht.files[handle]; !exists {
			break
		}
		handle++
	}

	fht.files[handle] = &fileEntry{
		file: file,
	}
	return handle
}

// Get はハンドルに対応するfileEntryを返す。
// 無効なハンドルの場合はエラーを返す。
// Requirement 3.5: 無効なハンドルでファイル操作が試みられた場合、エラーを返す。
func (fht *FileHandleTable) Get(handle int) (*fileEntry, error) {
	fht.mu.Lock()
	defer fht.mu.Unlock()

	entry, exists := fht.files[handle]
	if !exists {
		return nil, fmt.Errorf("invalid file handle: %d", handle)
	}
	return entry, nil
}

// Close はハンドルのファイルを閉じてハンドルを解放する。
// 解放されたハンドルは後続のOpen呼び出しで再利用される。
// Requirement 3.3: ファイルが閉じられた場合、対応するハンドルを解放して再利用可能にする。
// Requirement 3.5: 無効なハンドルでファイル操作が試みられた場合、エラーを返す。
func (fht *FileHandleTable) Close(handle int) error {
	fht.mu.Lock()
	defer fht.mu.Unlock()

	entry, exists := fht.files[handle]
	if !exists {
		return fmt.Errorf("invalid file handle: %d", handle)
	}

	err := entry.file.Close()
	delete(fht.files, handle)
	return err
}

// CloseAll は開いている全てのファイルを閉じてリソースを解放する。
// 個別のCloseエラーはログに記録するが、クリーンアップ処理は継続する。
// Requirement 3.4: VMが停止する場合、開いている全てのファイルを閉じてリソースを解放する。
func (fht *FileHandleTable) CloseAll() {
	fht.mu.Lock()
	defer fht.mu.Unlock()

	for handle, entry := range fht.files {
		_ = entry.file.Close()
		delete(fht.files, handle)
	}
}

// ResetReader はハンドルに対応するfileEntryのbufio.Readerをリセットする。
// SeekF呼び出し時に使用し、バッファリングとファイルポインタの整合性を保つ。
func (fht *FileHandleTable) ResetReader(handle int) {
	fht.mu.Lock()
	defer fht.mu.Unlock()

	entry, exists := fht.files[handle]
	if !exists {
		return
	}

	// readerがnilでない場合のみリセット（遅延初期化されていない場合は何もしない）
	if entry.reader != nil {
		entry.reader.Reset(entry.file)
	}
}
