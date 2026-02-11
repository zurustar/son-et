package vm

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// OpenF mode flag constants
// Requirement 4.3: アクセス属性（下位2ビット）: 0=読み書き、1=書き専用、2=読み専用
const (
	FileAccessReadWrite = 0x0000 // 読み書き
	FileAccessWriteOnly = 0x0001 // 書き専用
	FileAccessReadOnly  = 0x0002 // 読み専用
	FileAccessMask      = 0x0003 // アクセス属性マスク

	// Requirement 4.4: 新規作成フラグ 0x1000
	FileCreateNew = 0x1000 // ファイルを新規作成（存在すれば切り詰め）
)

// SeekF origin constants
// Requirement 6.2: 0=ファイル先頭、1=現在位置、2=ファイル末尾
const (
	SeekSet = 0 // ファイル先頭から → io.SeekStart
	SeekCur = 1 // 現在位置から → io.SeekCurrent
	SeekEnd = 2 // ファイル末尾から → io.SeekEnd
)

// registerFileIOBuiltins registers file I/O built-in functions.
func (vm *VM) registerFileIOBuiltins() {
	vm.registerOpenF()
	vm.registerCloseF()
	vm.registerSeekF()
	vm.registerReadF()
	vm.registerWriteF()
	vm.registerStrReadF()
	vm.registerStrWriteF()
}

// registerOpenF registers the OpenF built-in function.
// OpenF(filename [, mode]) — ファイルを開く
// Requirement 4.1: 引数1つで読み込み専用
// Requirement 4.2: 引数2つでmodeフラグに従う
// Requirement 4.5: 相対パスはタイトルディレクトリから解決
// Requirement 4.6: ファイル不存在で新規作成フラグなしはエラー
// Requirement 4.7: 引数不足はエラー
func (vm *VM) registerOpenF() {
	vm.RegisterBuiltinFunction("OpenF", func(v *VM, args []any) (any, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("OpenF requires at least 1 argument (filename), got %d", len(args))
		}

		filename := toString(args[0])
		fullPath := v.resolveFilePath(filename)

		// Default: read-only (1 arg)
		mode := int64(FileAccessReadOnly)
		if len(args) >= 2 {
			var ok bool
			mode, ok = toInt64(args[1])
			if !ok {
				return nil, fmt.Errorf("OpenF mode must be integer, got %T", args[1])
			}
		}

		accessMode := mode & FileAccessMask
		createNew := mode&FileCreateNew != 0

		var flag int
		switch accessMode {
		case FileAccessReadWrite:
			flag = os.O_RDWR
		case FileAccessWriteOnly:
			flag = os.O_WRONLY
		case FileAccessReadOnly:
			flag = os.O_RDONLY
		default:
			flag = os.O_RDONLY
		}

		if createNew {
			flag |= os.O_CREATE | os.O_TRUNC
		}

		var file *os.File
		var err error

		if createNew {
			file, err = os.OpenFile(fullPath, flag, 0644)
		} else {
			file, err = os.OpenFile(fullPath, flag, 0)
		}
		if err != nil {
			return nil, fmt.Errorf("OpenF: %w", err)
		}

		handle := v.fileHandleTable.Open(file)
		v.log.Debug("OpenF called", "filename", filename, "mode", mode, "handle", handle)
		return int64(handle), nil
	})
}

// registerCloseF registers the CloseF built-in function.
// CloseF(handle) — ファイルを閉じる
// Requirement 5.1: ハンドルのファイルを閉じ、ハンドルを解放する
// Requirement 5.2: 無効なハンドルはエラー
// Requirement 5.3: 引数不足はエラー
func (vm *VM) registerCloseF() {
	vm.RegisterBuiltinFunction("CloseF", func(v *VM, args []any) (any, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("CloseF requires 1 argument (handle), got %d", len(args))
		}

		handle, ok := toInt64(args[0])
		if !ok {
			return nil, fmt.Errorf("CloseF handle must be integer, got %T", args[0])
		}

		err := v.fileHandleTable.Close(int(handle))
		if err != nil {
			return nil, fmt.Errorf("CloseF: %w", err)
		}

		v.log.Debug("CloseF called", "handle", handle)
		return nil, nil
	})
}

// registerSeekF registers the SeekF built-in function.
// SeekF(handle, offset, origin) — ファイルポインタ移動
// Requirement 6.1: originを基準にoffsetバイト分移動
// Requirement 6.2: origin: 0=SEEK_SET, 1=SEEK_CUR, 2=SEEK_END
// Requirement 6.3: 無効なハンドルはエラー
// Requirement 6.4: 引数不足はエラー
func (vm *VM) registerSeekF() {
	vm.RegisterBuiltinFunction("SeekF", func(v *VM, args []any) (any, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("SeekF requires 3 arguments (handle, offset, origin), got %d", len(args))
		}

		handle, ok := toInt64(args[0])
		if !ok {
			return nil, fmt.Errorf("SeekF handle must be integer, got %T", args[0])
		}

		offset, ok := toInt64(args[1])
		if !ok {
			return nil, fmt.Errorf("SeekF offset must be integer, got %T", args[1])
		}

		origin, ok := toInt64(args[2])
		if !ok {
			return nil, fmt.Errorf("SeekF origin must be integer, got %T", args[2])
		}

		entry, err := v.fileHandleTable.Get(int(handle))
		if err != nil {
			return nil, fmt.Errorf("SeekF: %w", err)
		}

		// Map FILLY origin constants to Go io.Seek* constants
		var whence int
		switch origin {
		case SeekSet:
			whence = io.SeekStart
		case SeekCur:
			whence = io.SeekCurrent
		case SeekEnd:
			whence = io.SeekEnd
		default:
			return nil, fmt.Errorf("SeekF: invalid origin %d (must be 0, 1, or 2)", origin)
		}

		newPos, err := entry.file.Seek(offset, whence)
		if err != nil {
			return nil, fmt.Errorf("SeekF: %w", err)
		}

		// Reset bufio.Reader to maintain consistency with file pointer
		v.fileHandleTable.ResetReader(int(handle))

		v.log.Debug("SeekF called", "handle", handle, "offset", offset, "origin", origin, "newPos", newPos)
		return int64(newPos), nil
	})
}

// Valid byte sizes for ReadF and WriteF operations.
const (
	readWriteSize1 = 1
	readWriteSize2 = 2
	readWriteSize4 = 4
)

// registerReadF registers the ReadF built-in function.
// ReadF(handle, size) — バイナリ読み込み（1〜4バイト→リトルエンディアン整数）
// Requirement 7.1: sizeバイト（1〜4）を読み込み、リトルエンディアンの整数値として返す
// Requirement 7.2: size=1 → 0〜255の整数値
// Requirement 7.3: size=2 → リトルエンディアンの16ビット整数値
// Requirement 7.4: size=4 → リトルエンディアンの32ビット整数値
// Requirement 7.5: size範囲外はエラー
// Requirement 7.6: 無効なハンドルはエラー
// Requirement 7.7: 引数不足はエラー
func (vm *VM) registerReadF() {
	vm.RegisterBuiltinFunction("ReadF", func(v *VM, args []any) (any, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("ReadF requires 2 arguments (handle, size), got %d", len(args))
		}

		handle, ok := toInt64(args[0])
		if !ok {
			return nil, fmt.Errorf("ReadF handle must be integer, got %T", args[0])
		}

		size, ok := toInt64(args[1])
		if !ok {
			return nil, fmt.Errorf("ReadF size must be integer, got %T", args[1])
		}

		entry, err := v.fileHandleTable.Get(int(handle))
		if err != nil {
			return nil, fmt.Errorf("ReadF: %w", err)
		}

		var result int64
		switch size {
		case readWriteSize1:
			buf := make([]byte, readWriteSize1)
			if _, err := io.ReadFull(entry.file, buf); err != nil {
				return nil, fmt.Errorf("ReadF: %w", err)
			}
			result = int64(buf[0])
		case readWriteSize2:
			buf := make([]byte, readWriteSize2)
			if _, err := io.ReadFull(entry.file, buf); err != nil {
				return nil, fmt.Errorf("ReadF: %w", err)
			}
			result = int64(binary.LittleEndian.Uint16(buf))
		case readWriteSize4:
			buf := make([]byte, readWriteSize4)
			if _, err := io.ReadFull(entry.file, buf); err != nil {
				return nil, fmt.Errorf("ReadF: %w", err)
			}
			result = int64(binary.LittleEndian.Uint32(buf))
		default:
			return nil, fmt.Errorf("ReadF: invalid size %d (must be 1, 2, or 4)", size)
		}

		v.log.Debug("ReadF called", "handle", handle, "size", size, "result", result)
		return result, nil
	})
}

// registerWriteF registers the WriteF built-in function.
// WriteF(handle, value [, length]) — バイナリ書き込み
// Requirement 8.1: 3引数でvalueをリトルエンディアンでlengthバイト書き込み
// Requirement 8.2: 2引数でvalueを1バイトとして書き込み
// Requirement 8.3: length範囲外はエラー
// Requirement 8.4: 無効なハンドルはエラー
// Requirement 8.5: 引数不足はエラー
func (vm *VM) registerWriteF() {
	vm.RegisterBuiltinFunction("WriteF", func(v *VM, args []any) (any, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("WriteF requires at least 2 arguments (handle, value), got %d", len(args))
		}

		handle, ok := toInt64(args[0])
		if !ok {
			return nil, fmt.Errorf("WriteF handle must be integer, got %T", args[0])
		}

		value, ok := toInt64(args[1])
		if !ok {
			return nil, fmt.Errorf("WriteF value must be integer, got %T", args[1])
		}

		// Default length is 1 byte when called with 2 args
		length := int64(readWriteSize1)
		if len(args) >= 3 {
			length, ok = toInt64(args[2])
			if !ok {
				return nil, fmt.Errorf("WriteF length must be integer, got %T", args[2])
			}
		}

		entry, err := v.fileHandleTable.Get(int(handle))
		if err != nil {
			return nil, fmt.Errorf("WriteF: %w", err)
		}

		var buf []byte
		switch length {
		case readWriteSize1:
			buf = []byte{byte(value)}
		case readWriteSize2:
			buf = make([]byte, readWriteSize2)
			binary.LittleEndian.PutUint16(buf, uint16(value))
		case readWriteSize4:
			buf = make([]byte, readWriteSize4)
			binary.LittleEndian.PutUint32(buf, uint32(value))
		default:
			return nil, fmt.Errorf("WriteF: invalid length %d (must be 1, 2, or 4)", length)
		}

		if _, err := entry.file.Write(buf); err != nil {
			return nil, fmt.Errorf("WriteF: %w", err)
		}

		v.log.Debug("WriteF called", "handle", handle, "value", value, "length", length)
		return nil, nil
	})
}

// registerStrReadF registers the StrReadF built-in function.
// StrReadF(handle) — 1行読み込み（Shift-JIS→UTF-8変換）
// Requirement 9.1: ファイルから改行区切りで1行を読み込み、文字列として返す
// Requirement 9.2: Shift-JISエンコーディングの場合、UTF-8に変換して返す
// Requirement 9.3: EOF時は空文字列を返す
// Requirement 9.4: 改行文字（CR, LF, CRLF）を行区切りとして認識し、返す文字列には改行文字を含めない
// Requirement 9.5: 無効なハンドルはエラー
// Requirement 9.6: 引数不足はエラー
func (vm *VM) registerStrReadF() {
	vm.RegisterBuiltinFunction("StrReadF", func(v *VM, args []any) (any, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("StrReadF requires 1 argument (handle), got %d", len(args))
		}

		handle, ok := toInt64(args[0])
		if !ok {
			return nil, fmt.Errorf("StrReadF handle must be integer, got %T", args[0])
		}

		entry, err := v.fileHandleTable.Get(int(handle))
		if err != nil {
			return nil, fmt.Errorf("StrReadF: %w", err)
		}

		// Lazy-initialize bufio.Reader on first StrReadF call
		if entry.reader == nil {
			entry.reader = bufio.NewReader(entry.file)
		}

		line, err := readLineFromReader(entry.reader)
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("StrReadF: %w", err)
		}

		// EOF with no data → return empty string
		if err == io.EOF && len(line) == 0 {
			v.log.Debug("StrReadF called (EOF)", "handle", handle)
			return "", nil
		}

		// Decode Shift-JIS to UTF-8
		decoder := japanese.ShiftJIS.NewDecoder()
		utf8Str, _, decErr := transform.String(decoder, string(line))
		if decErr != nil {
			return nil, fmt.Errorf("StrReadF: Shift-JIS decode error: %w", decErr)
		}

		v.log.Debug("StrReadF called", "handle", handle, "result", utf8Str)
		return utf8Str, nil
	})
}

// readLineFromReader reads a single line from a bufio.Reader, handling CR, LF, and CRLF delimiters.
// Returns the line content without the trailing line delimiter.
func readLineFromReader(r *bufio.Reader) ([]byte, error) {
	var line []byte
	for {
		b, err := r.ReadByte()
		if err != nil {
			// EOF: return whatever we have so far
			return line, err
		}

		if b == '\n' {
			// LF or CRLF — line is complete
			return line, nil
		}

		if b == '\r' {
			// Could be CR alone or CRLF
			next, err := r.Peek(1)
			if err == nil && len(next) > 0 && next[0] == '\n' {
				// CRLF — consume the LF
				_, _ = r.ReadByte()
			}
			// CR or CRLF — line is complete
			return line, nil
		}

		line = append(line, b)
	}
}

// registerStrWriteF registers the StrWriteF built-in function.
// StrWriteF(handle, str) — 文字列書き込み（UTF-8→Shift-JIS変換、改行なし）
// Requirement 10.1: 文字列をファイルに書き込む（改行は付加しない）
// Requirement 10.2: UTF-8からShift-JISに変換して書き込む
// Requirement 10.3: 無効なハンドルはエラー
// Requirement 10.4: 引数不足はエラー
func (vm *VM) registerStrWriteF() {
	vm.RegisterBuiltinFunction("StrWriteF", func(v *VM, args []any) (any, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("StrWriteF requires 2 arguments (handle, str), got %d", len(args))
		}

		handle, ok := toInt64(args[0])
		if !ok {
			return nil, fmt.Errorf("StrWriteF handle must be integer, got %T", args[0])
		}

		str := toString(args[1])

		entry, err := v.fileHandleTable.Get(int(handle))
		if err != nil {
			return nil, fmt.Errorf("StrWriteF: %w", err)
		}

		// Encode UTF-8 to Shift-JIS
		encoder := japanese.ShiftJIS.NewEncoder()
		sjisStr, _, encErr := transform.String(encoder, str)
		if encErr != nil {
			return nil, fmt.Errorf("StrWriteF: Shift-JIS encode error: %w", encErr)
		}

		// Write without appending newline
		if _, err := io.WriteString(entry.file, sjisStr); err != nil {
			return nil, fmt.Errorf("StrWriteF: %w", err)
		}

		v.log.Debug("StrWriteF called", "handle", handle, "str", str)
		return nil, nil
	})
}

