package vm

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/zurustar/son-et/pkg/opcode"
)

// helperCreateTestVM creates a VM with a temporary directory as titlePath for file I/O tests.
func helperCreateTestVM(t *testing.T) (*VM, string) {
	t.Helper()
	tmpDir := t.TempDir()
	vm := New([]opcode.OpCode{}, WithTitlePath(tmpDir))
	return vm, tmpDir
}

// helperCreateTestFile creates a file with the given content in the specified directory.
func helperCreateTestFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	return path
}

// --- OpenF Tests ---

// TestOpenF_ReadOnlyOneArg tests OpenF with 1 argument (default read-only).
// Requirement 4.1: 引数1つで読み込み専用
func TestOpenF_ReadOnlyOneArg(t *testing.T) {
	vm, tmpDir := helperCreateTestVM(t)
	helperCreateTestFile(t, tmpDir, "test.dat", "hello")

	result, err := vm.builtins["OpenF"](vm, []any{"test.dat"})
	if err != nil {
		t.Fatalf("OpenF returned error: %v", err)
	}

	handle, ok := result.(int64)
	if !ok {
		t.Fatalf("OpenF should return int64, got %T", result)
	}
	if handle < 1 {
		t.Errorf("handle should be >= 1, got %d", handle)
	}

	// Clean up: close the handle
	_, err = vm.builtins["CloseF"](vm, []any{handle})
	if err != nil {
		t.Fatalf("CloseF returned error: %v", err)
	}
}

// TestOpenF_ReadOnlyWithMode tests OpenF with explicit read-only mode flag.
// Requirements 4.2, 4.3: 引数2つでmodeフラグに従う、アクセス属性2=読み専用
func TestOpenF_ReadOnlyWithMode(t *testing.T) {
	vm, tmpDir := helperCreateTestVM(t)
	helperCreateTestFile(t, tmpDir, "test.dat", "hello")

	result, err := vm.builtins["OpenF"](vm, []any{"test.dat", int64(FileAccessReadOnly)})
	if err != nil {
		t.Fatalf("OpenF returned error: %v", err)
	}

	handle, ok := result.(int64)
	if !ok {
		t.Fatalf("OpenF should return int64, got %T", result)
	}
	if handle < 1 {
		t.Errorf("handle should be >= 1, got %d", handle)
	}

	_, err = vm.builtins["CloseF"](vm, []any{handle})
	if err != nil {
		t.Fatalf("CloseF returned error: %v", err)
	}
}

// TestOpenF_WriteOnly tests OpenF with write-only mode flag.
// Requirement 4.3: アクセス属性1=書き専用
func TestOpenF_WriteOnly(t *testing.T) {
	vm, tmpDir := helperCreateTestVM(t)
	helperCreateTestFile(t, tmpDir, "test.dat", "hello")

	result, err := vm.builtins["OpenF"](vm, []any{"test.dat", int64(FileAccessWriteOnly)})
	if err != nil {
		t.Fatalf("OpenF returned error: %v", err)
	}

	handle, ok := result.(int64)
	if !ok {
		t.Fatalf("OpenF should return int64, got %T", result)
	}
	if handle < 1 {
		t.Errorf("handle should be >= 1, got %d", handle)
	}

	_, err = vm.builtins["CloseF"](vm, []any{handle})
	if err != nil {
		t.Fatalf("CloseF returned error: %v", err)
	}
}

// TestOpenF_ReadWrite tests OpenF with read-write mode flag.
// Requirement 4.3: アクセス属性0=読み書き
func TestOpenF_ReadWrite(t *testing.T) {
	vm, tmpDir := helperCreateTestVM(t)
	helperCreateTestFile(t, tmpDir, "test.dat", "hello")

	result, err := vm.builtins["OpenF"](vm, []any{"test.dat", int64(FileAccessReadWrite)})
	if err != nil {
		t.Fatalf("OpenF returned error: %v", err)
	}

	handle, ok := result.(int64)
	if !ok {
		t.Fatalf("OpenF should return int64, got %T", result)
	}
	if handle < 1 {
		t.Errorf("handle should be >= 1, got %d", handle)
	}

	_, err = vm.builtins["CloseF"](vm, []any{handle})
	if err != nil {
		t.Fatalf("CloseF returned error: %v", err)
	}
}

// TestOpenF_CreateNew tests OpenF with create-new flag creating a new file.
// Requirement 4.4: 新規作成フラグ0x1000でファイルが存在しなければ新規作成
func TestOpenF_CreateNew(t *testing.T) {
	vm, tmpDir := helperCreateTestVM(t)

	// File does not exist yet
	newFile := "newfile.dat"
	newFilePath := filepath.Join(tmpDir, newFile)

	result, err := vm.builtins["OpenF"](vm, []any{newFile, int64(FileCreateNew | FileAccessReadWrite)})
	if err != nil {
		t.Fatalf("OpenF with create-new returned error: %v", err)
	}

	handle, ok := result.(int64)
	if !ok {
		t.Fatalf("OpenF should return int64, got %T", result)
	}

	// Verify file was created
	if _, err := os.Stat(newFilePath); os.IsNotExist(err) {
		t.Error("file should have been created by OpenF with create-new flag")
	}

	_, err = vm.builtins["CloseF"](vm, []any{handle})
	if err != nil {
		t.Fatalf("CloseF returned error: %v", err)
	}
}

// TestOpenF_CreateNewTruncates tests OpenF with create-new flag truncating existing file.
// Requirement 4.4: 存在すれば内容を切り詰めて開く
func TestOpenF_CreateNewTruncates(t *testing.T) {
	vm, tmpDir := helperCreateTestVM(t)
	filePath := helperCreateTestFile(t, tmpDir, "existing.dat", "original content")

	result, err := vm.builtins["OpenF"](vm, []any{"existing.dat", int64(FileCreateNew | FileAccessReadWrite)})
	if err != nil {
		t.Fatalf("OpenF with create-new returned error: %v", err)
	}

	handle, ok := result.(int64)
	if !ok {
		t.Fatalf("OpenF should return int64, got %T", result)
	}

	_, err = vm.builtins["CloseF"](vm, []any{handle})
	if err != nil {
		t.Fatalf("CloseF returned error: %v", err)
	}

	// Verify file was truncated
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if len(data) != 0 {
		t.Errorf("file should be truncated (empty), got %d bytes: %q", len(data), data)
	}
}

// TestOpenF_NonExistentFileError tests OpenF returns error for non-existent file without create flag.
// Requirement 4.6: ファイル不存在で新規作成フラグなしはエラー
func TestOpenF_NonExistentFileError(t *testing.T) {
	vm, _ := helperCreateTestVM(t)

	_, err := vm.builtins["OpenF"](vm, []any{"nonexistent.dat"})
	if err == nil {
		t.Error("OpenF should return error for non-existent file without create flag")
	}
}

// TestOpenF_MissingArgs tests OpenF returns error with missing arguments.
// Requirement 4.7: 引数不足はエラー
func TestOpenF_MissingArgs(t *testing.T) {
	vm, _ := helperCreateTestVM(t)

	_, err := vm.builtins["OpenF"](vm, []any{})
	if err == nil {
		t.Error("OpenF should return error with no arguments")
	}
}

// --- CloseF Tests ---

// TestCloseF_ValidHandle tests CloseF closes a valid handle successfully.
// Requirement 5.1: ハンドルのファイルを閉じ、ハンドルを解放する
func TestCloseF_ValidHandle(t *testing.T) {
	vm, tmpDir := helperCreateTestVM(t)
	helperCreateTestFile(t, tmpDir, "test.dat", "hello")

	result, err := vm.builtins["OpenF"](vm, []any{"test.dat"})
	if err != nil {
		t.Fatalf("OpenF returned error: %v", err)
	}
	handle := result.(int64)

	// Close should succeed
	result, err = vm.builtins["CloseF"](vm, []any{handle})
	if err != nil {
		t.Fatalf("CloseF returned error: %v", err)
	}
	if result != nil {
		t.Errorf("CloseF should return nil, got %v", result)
	}

	// Attempting to close again should fail (handle released)
	_, err = vm.builtins["CloseF"](vm, []any{handle})
	if err == nil {
		t.Error("CloseF should return error for already closed handle")
	}
}

// TestCloseF_InvalidHandle tests CloseF returns error for invalid handle.
// Requirement 5.2: 無効なハンドルはエラー
func TestCloseF_InvalidHandle(t *testing.T) {
	vm, _ := helperCreateTestVM(t)

	_, err := vm.builtins["CloseF"](vm, []any{int64(999)})
	if err == nil {
		t.Error("CloseF should return error for invalid handle")
	}
}

// TestCloseF_MissingArgs tests CloseF returns error with missing arguments.
// Requirement 5.3: 引数不足はエラー
func TestCloseF_MissingArgs(t *testing.T) {
	vm, _ := helperCreateTestVM(t)

	_, err := vm.builtins["CloseF"](vm, []any{})
	if err == nil {
		t.Error("CloseF should return error with no arguments")
	}
}

// --- SeekF Tests ---

// TestSeekF_SeekSet tests SeekF with origin=0 (SEEK_SET, from file start).
// Requirements 6.1, 6.2: originを基準にoffsetバイト分移動、0=ファイル先頭
func TestSeekF_SeekSet(t *testing.T) {
	vm, tmpDir := helperCreateTestVM(t)
	helperCreateTestFile(t, tmpDir, "test.dat", "abcdefghij") // 10 bytes

	result, err := vm.builtins["OpenF"](vm, []any{"test.dat", int64(FileAccessReadOnly)})
	if err != nil {
		t.Fatalf("OpenF returned error: %v", err)
	}
	handle := result.(int64)
	defer vm.builtins["CloseF"](vm, []any{handle})

	// Seek to offset 5 from start
	result, err = vm.builtins["SeekF"](vm, []any{handle, int64(5), int64(SeekSet)})
	if err != nil {
		t.Fatalf("SeekF returned error: %v", err)
	}

	newPos, ok := result.(int64)
	if !ok {
		t.Fatalf("SeekF should return int64, got %T", result)
	}
	if newPos != 5 {
		t.Errorf("SeekF SEEK_SET offset=5 should return position 5, got %d", newPos)
	}
}

// TestSeekF_SeekCur tests SeekF with origin=1 (SEEK_CUR, from current position).
// Requirements 6.1, 6.2: originを基準にoffsetバイト分移動、1=現在位置
func TestSeekF_SeekCur(t *testing.T) {
	vm, tmpDir := helperCreateTestVM(t)
	helperCreateTestFile(t, tmpDir, "test.dat", "abcdefghij") // 10 bytes

	result, err := vm.builtins["OpenF"](vm, []any{"test.dat", int64(FileAccessReadOnly)})
	if err != nil {
		t.Fatalf("OpenF returned error: %v", err)
	}
	handle := result.(int64)
	defer vm.builtins["CloseF"](vm, []any{handle})

	// First seek to position 3
	_, err = vm.builtins["SeekF"](vm, []any{handle, int64(3), int64(SeekSet)})
	if err != nil {
		t.Fatalf("SeekF (setup) returned error: %v", err)
	}

	// Then seek +4 from current position → should be at 7
	result, err = vm.builtins["SeekF"](vm, []any{handle, int64(4), int64(SeekCur)})
	if err != nil {
		t.Fatalf("SeekF returned error: %v", err)
	}

	newPos, ok := result.(int64)
	if !ok {
		t.Fatalf("SeekF should return int64, got %T", result)
	}
	if newPos != 7 {
		t.Errorf("SeekF SEEK_CUR from 3 + offset 4 should return position 7, got %d", newPos)
	}
}

// TestSeekF_SeekEnd tests SeekF with origin=2 (SEEK_END, from file end).
// Requirements 6.1, 6.2: originを基準にoffsetバイト分移動、2=ファイル末尾
func TestSeekF_SeekEnd(t *testing.T) {
	vm, tmpDir := helperCreateTestVM(t)
	helperCreateTestFile(t, tmpDir, "test.dat", "abcdefghij") // 10 bytes

	result, err := vm.builtins["OpenF"](vm, []any{"test.dat", int64(FileAccessReadOnly)})
	if err != nil {
		t.Fatalf("OpenF returned error: %v", err)
	}
	handle := result.(int64)
	defer vm.builtins["CloseF"](vm, []any{handle})

	// Seek to -3 from end → should be at position 7
	result, err = vm.builtins["SeekF"](vm, []any{handle, int64(-3), int64(SeekEnd)})
	if err != nil {
		t.Fatalf("SeekF returned error: %v", err)
	}

	newPos, ok := result.(int64)
	if !ok {
		t.Fatalf("SeekF should return int64, got %T", result)
	}
	if newPos != 7 {
		t.Errorf("SeekF SEEK_END offset=-3 on 10-byte file should return position 7, got %d", newPos)
	}
}

// TestSeekF_InvalidHandle tests SeekF returns error for invalid handle.
// Requirement 6.3: 無効なハンドルはエラー
func TestSeekF_InvalidHandle(t *testing.T) {
	vm, _ := helperCreateTestVM(t)

	_, err := vm.builtins["SeekF"](vm, []any{int64(999), int64(0), int64(SeekSet)})
	if err == nil {
		t.Error("SeekF should return error for invalid handle")
	}
}

// TestSeekF_MissingArgs tests SeekF returns error with missing arguments.
// Requirement 6.4: 引数不足はエラー
func TestSeekF_MissingArgs(t *testing.T) {
	vm, _ := helperCreateTestVM(t)

	tests := []struct {
		name string
		args []any
	}{
		{"no arguments", []any{}},
		{"one argument", []any{int64(1)}},
		{"two arguments", []any{int64(1), int64(0)}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := vm.builtins["SeekF"](vm, tt.args)
			if err == nil {
				t.Errorf("SeekF should return error with %s", tt.name)
			}
		})
	}
}

// --- ReadF Tests ---

// TestReadF_Size1 tests ReadF with size=1 reads a single byte correctly.
// Requirement 7.2: size=1 → 0〜255の整数値
func TestReadF_Size1(t *testing.T) {
	vm, tmpDir := helperCreateTestVM(t)
	// Write byte 0xAB (171 decimal)
	helperCreateTestFile(t, tmpDir, "test.dat", string([]byte{0xAB}))

	result, err := vm.builtins["OpenF"](vm, []any{"test.dat", int64(FileAccessReadOnly)})
	if err != nil {
		t.Fatalf("OpenF returned error: %v", err)
	}
	handle := result.(int64)
	defer vm.builtins["CloseF"](vm, []any{handle})

	result, err = vm.builtins["ReadF"](vm, []any{handle, int64(1)})
	if err != nil {
		t.Fatalf("ReadF returned error: %v", err)
	}

	val, ok := result.(int64)
	if !ok {
		t.Fatalf("ReadF should return int64, got %T", result)
	}
	if val != 0xAB {
		t.Errorf("ReadF size=1 should return 0xAB (171), got 0x%X (%d)", val, val)
	}
}

// TestReadF_Size2 tests ReadF with size=2 reads 16-bit little-endian correctly.
// Requirement 7.3: size=2 → リトルエンディアンの16ビット整数値
func TestReadF_Size2(t *testing.T) {
	vm, tmpDir := helperCreateTestVM(t)
	// 0x0102 in little-endian: [0x02, 0x01]
	helperCreateTestFile(t, tmpDir, "test.dat", string([]byte{0x02, 0x01}))

	result, err := vm.builtins["OpenF"](vm, []any{"test.dat", int64(FileAccessReadOnly)})
	if err != nil {
		t.Fatalf("OpenF returned error: %v", err)
	}
	handle := result.(int64)
	defer vm.builtins["CloseF"](vm, []any{handle})

	result, err = vm.builtins["ReadF"](vm, []any{handle, int64(2)})
	if err != nil {
		t.Fatalf("ReadF returned error: %v", err)
	}

	val, ok := result.(int64)
	if !ok {
		t.Fatalf("ReadF should return int64, got %T", result)
	}
	if val != 0x0102 {
		t.Errorf("ReadF size=2 should return 0x0102 (258), got 0x%X (%d)", val, val)
	}
}

// TestReadF_Size4 tests ReadF with size=4 reads 32-bit little-endian correctly.
// Requirement 7.4: size=4 → リトルエンディアンの32ビット整数値
func TestReadF_Size4(t *testing.T) {
	vm, tmpDir := helperCreateTestVM(t)
	// 0x01020304 in little-endian: [0x04, 0x03, 0x02, 0x01]
	helperCreateTestFile(t, tmpDir, "test.dat", string([]byte{0x04, 0x03, 0x02, 0x01}))

	result, err := vm.builtins["OpenF"](vm, []any{"test.dat", int64(FileAccessReadOnly)})
	if err != nil {
		t.Fatalf("OpenF returned error: %v", err)
	}
	handle := result.(int64)
	defer vm.builtins["CloseF"](vm, []any{handle})

	result, err = vm.builtins["ReadF"](vm, []any{handle, int64(4)})
	if err != nil {
		t.Fatalf("ReadF returned error: %v", err)
	}

	val, ok := result.(int64)
	if !ok {
		t.Fatalf("ReadF should return int64, got %T", result)
	}
	if val != 0x01020304 {
		t.Errorf("ReadF size=4 should return 0x01020304 (16909060), got 0x%X (%d)", val, val)
	}
}

// TestReadF_InvalidSize tests ReadF returns error for invalid size values.
// Requirement 7.5: size範囲外はエラー
func TestReadF_InvalidSize(t *testing.T) {
	vm, tmpDir := helperCreateTestVM(t)
	helperCreateTestFile(t, tmpDir, "test.dat", "abcdefgh")

	result, err := vm.builtins["OpenF"](vm, []any{"test.dat", int64(FileAccessReadOnly)})
	if err != nil {
		t.Fatalf("OpenF returned error: %v", err)
	}
	handle := result.(int64)
	defer vm.builtins["CloseF"](vm, []any{handle})

	invalidSizes := []int64{0, 3, 5, -1, 8}
	for _, size := range invalidSizes {
		_, err := vm.builtins["ReadF"](vm, []any{handle, int64(size)})
		if err == nil {
			t.Errorf("ReadF should return error for invalid size %d", size)
		}
	}
}

// TestReadF_InvalidHandle tests ReadF returns error for invalid handle.
// Requirement 7.6: 無効なハンドルはエラー
func TestReadF_InvalidHandle(t *testing.T) {
	vm, _ := helperCreateTestVM(t)

	_, err := vm.builtins["ReadF"](vm, []any{int64(999), int64(1)})
	if err == nil {
		t.Error("ReadF should return error for invalid handle")
	}
}

// TestReadF_MissingArgs tests ReadF returns error with missing arguments.
// Requirement 7.7: 引数不足はエラー
func TestReadF_MissingArgs(t *testing.T) {
	vm, _ := helperCreateTestVM(t)

	tests := []struct {
		name string
		args []any
	}{
		{"no arguments", []any{}},
		{"one argument", []any{int64(1)}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := vm.builtins["ReadF"](vm, tt.args)
			if err == nil {
				t.Errorf("ReadF should return error with %s", tt.name)
			}
		})
	}
}

// --- WriteF Tests ---

// TestWriteF_TwoArgs_WritesOneByte tests WriteF with 2 args writes 1 byte.
// Requirement 8.2: 2引数でvalueを1バイトとして書き込み
func TestWriteF_TwoArgs_WritesOneByte(t *testing.T) {
	vm, tmpDir := helperCreateTestVM(t)

	result, err := vm.builtins["OpenF"](vm, []any{"write1.dat", int64(FileCreateNew | FileAccessReadWrite)})
	if err != nil {
		t.Fatalf("OpenF returned error: %v", err)
	}
	handle := result.(int64)
	defer vm.builtins["CloseF"](vm, []any{handle})

	// WriteF with 2 args: handle, value → writes 1 byte
	_, err = vm.builtins["WriteF"](vm, []any{handle, int64(0xCD)})
	if err != nil {
		t.Fatalf("WriteF returned error: %v", err)
	}

	// Verify file content
	data, err := os.ReadFile(filepath.Join(tmpDir, "write1.dat"))
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if len(data) != 1 || data[0] != 0xCD {
		t.Errorf("WriteF 2-arg should write 1 byte 0xCD, got %v", data)
	}
}

// TestWriteF_ThreeArgs_WritesLittleEndian tests WriteF with 3 args writes little-endian bytes.
// Requirement 8.1: 3引数でvalueをリトルエンディアンでlengthバイト書き込み
func TestWriteF_ThreeArgs_WritesLittleEndian(t *testing.T) {
	vm, tmpDir := helperCreateTestVM(t)

	// Test length=2: 0x0102 → [0x02, 0x01]
	result, err := vm.builtins["OpenF"](vm, []any{"write2.dat", int64(FileCreateNew | FileAccessReadWrite)})
	if err != nil {
		t.Fatalf("OpenF returned error: %v", err)
	}
	handle := result.(int64)

	_, err = vm.builtins["WriteF"](vm, []any{handle, int64(0x0102), int64(2)})
	if err != nil {
		t.Fatalf("WriteF returned error: %v", err)
	}
	vm.builtins["CloseF"](vm, []any{handle})

	data, err := os.ReadFile(filepath.Join(tmpDir, "write2.dat"))
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if len(data) != 2 || data[0] != 0x02 || data[1] != 0x01 {
		t.Errorf("WriteF 3-arg length=2 for 0x0102 should write [0x02, 0x01], got %v", data)
	}

	// Test length=4: 0x01020304 → [0x04, 0x03, 0x02, 0x01]
	result, err = vm.builtins["OpenF"](vm, []any{"write4.dat", int64(FileCreateNew | FileAccessReadWrite)})
	if err != nil {
		t.Fatalf("OpenF returned error: %v", err)
	}
	handle = result.(int64)

	_, err = vm.builtins["WriteF"](vm, []any{handle, int64(0x01020304), int64(4)})
	if err != nil {
		t.Fatalf("WriteF returned error: %v", err)
	}
	vm.builtins["CloseF"](vm, []any{handle})

	data, err = os.ReadFile(filepath.Join(tmpDir, "write4.dat"))
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if len(data) != 4 || data[0] != 0x04 || data[1] != 0x03 || data[2] != 0x02 || data[3] != 0x01 {
		t.Errorf("WriteF 3-arg length=4 for 0x01020304 should write [0x04, 0x03, 0x02, 0x01], got %v", data)
	}
}

// TestWriteF_InvalidLength tests WriteF returns error for invalid length values.
// Requirement 8.3: length範囲外はエラー
func TestWriteF_InvalidLength(t *testing.T) {
	vm, _ := helperCreateTestVM(t)

	result, err := vm.builtins["OpenF"](vm, []any{"invalid_len.dat", int64(FileCreateNew | FileAccessReadWrite)})
	if err != nil {
		t.Fatalf("OpenF returned error: %v", err)
	}
	handle := result.(int64)
	defer vm.builtins["CloseF"](vm, []any{handle})

	invalidLengths := []int64{0, 3, 5, -1, 8}
	for _, length := range invalidLengths {
		_, err := vm.builtins["WriteF"](vm, []any{handle, int64(42), int64(length)})
		if err == nil {
			t.Errorf("WriteF should return error for invalid length %d", length)
		}
	}
}

// TestWriteF_InvalidHandle tests WriteF returns error for invalid handle.
// Requirement 8.4: 無効なハンドルはエラー
func TestWriteF_InvalidHandle(t *testing.T) {
	vm, _ := helperCreateTestVM(t)

	_, err := vm.builtins["WriteF"](vm, []any{int64(999), int64(42)})
	if err == nil {
		t.Error("WriteF should return error for invalid handle")
	}
}

// TestWriteF_MissingArgs tests WriteF returns error with missing arguments.
// Requirement 8.5: 引数不足はエラー
func TestWriteF_MissingArgs(t *testing.T) {
	vm, _ := helperCreateTestVM(t)

	tests := []struct {
		name string
		args []any
	}{
		{"no arguments", []any{}},
		{"one argument", []any{int64(1)}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := vm.builtins["WriteF"](vm, tt.args)
			if err == nil {
				t.Errorf("WriteF should return error with %s", tt.name)
			}
		})
	}
}

// --- ReadF/WriteF Round-trip Tests ---

// TestReadFWriteF_RoundTrip1Byte tests write then read back 1 byte.
// Validates: Requirements 7.2, 8.2
func TestReadFWriteF_RoundTrip1Byte(t *testing.T) {
	vm, _ := helperCreateTestVM(t)

	result, err := vm.builtins["OpenF"](vm, []any{"rt1.dat", int64(FileCreateNew | FileAccessReadWrite)})
	if err != nil {
		t.Fatalf("OpenF returned error: %v", err)
	}
	handle := result.(int64)
	defer vm.builtins["CloseF"](vm, []any{handle})

	// Write 1 byte
	_, err = vm.builtins["WriteF"](vm, []any{handle, int64(0xEF)})
	if err != nil {
		t.Fatalf("WriteF returned error: %v", err)
	}

	// Seek back to start
	_, err = vm.builtins["SeekF"](vm, []any{handle, int64(0), int64(SeekSet)})
	if err != nil {
		t.Fatalf("SeekF returned error: %v", err)
	}

	// Read back
	result, err = vm.builtins["ReadF"](vm, []any{handle, int64(1)})
	if err != nil {
		t.Fatalf("ReadF returned error: %v", err)
	}

	val := result.(int64)
	if val != 0xEF {
		t.Errorf("round-trip 1 byte: wrote 0xEF, read back 0x%X", val)
	}
}

// TestReadFWriteF_RoundTrip2Bytes tests write then read back 2 bytes (little-endian).
// Validates: Requirements 7.3, 8.1
func TestReadFWriteF_RoundTrip2Bytes(t *testing.T) {
	vm, _ := helperCreateTestVM(t)

	result, err := vm.builtins["OpenF"](vm, []any{"rt2.dat", int64(FileCreateNew | FileAccessReadWrite)})
	if err != nil {
		t.Fatalf("OpenF returned error: %v", err)
	}
	handle := result.(int64)
	defer vm.builtins["CloseF"](vm, []any{handle})

	// Write 2 bytes: 0xBEEF
	_, err = vm.builtins["WriteF"](vm, []any{handle, int64(0xBEEF), int64(2)})
	if err != nil {
		t.Fatalf("WriteF returned error: %v", err)
	}

	// Seek back to start
	_, err = vm.builtins["SeekF"](vm, []any{handle, int64(0), int64(SeekSet)})
	if err != nil {
		t.Fatalf("SeekF returned error: %v", err)
	}

	// Read back
	result, err = vm.builtins["ReadF"](vm, []any{handle, int64(2)})
	if err != nil {
		t.Fatalf("ReadF returned error: %v", err)
	}

	val := result.(int64)
	if val != 0xBEEF {
		t.Errorf("round-trip 2 bytes: wrote 0xBEEF, read back 0x%X", val)
	}
}

// TestReadFWriteF_RoundTrip4Bytes tests write then read back 4 bytes (little-endian).
// Validates: Requirements 7.4, 8.1
func TestReadFWriteF_RoundTrip4Bytes(t *testing.T) {
	vm, _ := helperCreateTestVM(t)

	result, err := vm.builtins["OpenF"](vm, []any{"rt4.dat", int64(FileCreateNew | FileAccessReadWrite)})
	if err != nil {
		t.Fatalf("OpenF returned error: %v", err)
	}
	handle := result.(int64)
	defer vm.builtins["CloseF"](vm, []any{handle})

	// Write 4 bytes: 0xDEADBEEF
	_, err = vm.builtins["WriteF"](vm, []any{handle, int64(0xDEADBEEF), int64(4)})
	if err != nil {
		t.Fatalf("WriteF returned error: %v", err)
	}

	// Seek back to start
	_, err = vm.builtins["SeekF"](vm, []any{handle, int64(0), int64(SeekSet)})
	if err != nil {
		t.Fatalf("SeekF returned error: %v", err)
	}

	// Read back
	result, err = vm.builtins["ReadF"](vm, []any{handle, int64(4)})
	if err != nil {
		t.Fatalf("ReadF returned error: %v", err)
	}

	val := result.(int64)
	if val != 0xDEADBEEF {
		t.Errorf("round-trip 4 bytes: wrote 0xDEADBEEF, read back 0x%X", val)
	}
}

// --- StrReadF Tests ---

// TestStrReadF_LF tests StrReadF reads a line delimited by LF.
// Validates: Requirements 9.1, 9.4
func TestStrReadF_LF(t *testing.T) {
	vm, tmpDir := helperCreateTestVM(t)

	// Write raw bytes with LF line endings (ASCII is valid Shift-JIS)
	path := filepath.Join(tmpDir, "lf.dat")
	if err := os.WriteFile(path, []byte("hello\nworld\n"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	result, err := vm.builtins["OpenF"](vm, []any{"lf.dat"})
	if err != nil {
		t.Fatalf("OpenF returned error: %v", err)
	}
	handle := result.(int64)
	defer vm.builtins["CloseF"](vm, []any{handle})

	// Read first line
	result, err = vm.builtins["StrReadF"](vm, []any{handle})
	if err != nil {
		t.Fatalf("StrReadF returned error: %v", err)
	}
	if result != "hello" {
		t.Errorf("expected %q, got %q", "hello", result)
	}

	// Read second line
	result, err = vm.builtins["StrReadF"](vm, []any{handle})
	if err != nil {
		t.Fatalf("StrReadF returned error: %v", err)
	}
	if result != "world" {
		t.Errorf("expected %q, got %q", "world", result)
	}
}

// TestStrReadF_CRLF tests StrReadF reads a line delimited by CRLF.
// Validates: Requirements 9.1, 9.4
func TestStrReadF_CRLF(t *testing.T) {
	vm, tmpDir := helperCreateTestVM(t)

	path := filepath.Join(tmpDir, "crlf.dat")
	if err := os.WriteFile(path, []byte("hello\r\nworld\r\n"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	result, err := vm.builtins["OpenF"](vm, []any{"crlf.dat"})
	if err != nil {
		t.Fatalf("OpenF returned error: %v", err)
	}
	handle := result.(int64)
	defer vm.builtins["CloseF"](vm, []any{handle})

	result, err = vm.builtins["StrReadF"](vm, []any{handle})
	if err != nil {
		t.Fatalf("StrReadF returned error: %v", err)
	}
	if result != "hello" {
		t.Errorf("expected %q, got %q", "hello", result)
	}

	result, err = vm.builtins["StrReadF"](vm, []any{handle})
	if err != nil {
		t.Fatalf("StrReadF returned error: %v", err)
	}
	if result != "world" {
		t.Errorf("expected %q, got %q", "world", result)
	}
}

// TestStrReadF_CR tests StrReadF reads a line delimited by CR only.
// Validates: Requirements 9.1, 9.4
func TestStrReadF_CR(t *testing.T) {
	vm, tmpDir := helperCreateTestVM(t)

	path := filepath.Join(tmpDir, "cr.dat")
	if err := os.WriteFile(path, []byte("hello\rworld\r"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	result, err := vm.builtins["OpenF"](vm, []any{"cr.dat"})
	if err != nil {
		t.Fatalf("OpenF returned error: %v", err)
	}
	handle := result.(int64)
	defer vm.builtins["CloseF"](vm, []any{handle})

	result, err = vm.builtins["StrReadF"](vm, []any{handle})
	if err != nil {
		t.Fatalf("StrReadF returned error: %v", err)
	}
	if result != "hello" {
		t.Errorf("expected %q, got %q", "hello", result)
	}

	result, err = vm.builtins["StrReadF"](vm, []any{handle})
	if err != nil {
		t.Fatalf("StrReadF returned error: %v", err)
	}
	if result != "world" {
		t.Errorf("expected %q, got %q", "world", result)
	}
}

// TestStrReadF_EOF tests StrReadF returns empty string at EOF.
// Validates: Requirement 9.3
func TestStrReadF_EOF(t *testing.T) {
	vm, tmpDir := helperCreateTestVM(t)

	// Create an empty file
	path := filepath.Join(tmpDir, "empty.dat")
	if err := os.WriteFile(path, []byte{}, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	result, err := vm.builtins["OpenF"](vm, []any{"empty.dat"})
	if err != nil {
		t.Fatalf("OpenF returned error: %v", err)
	}
	handle := result.(int64)
	defer vm.builtins["CloseF"](vm, []any{handle})

	result, err = vm.builtins["StrReadF"](vm, []any{handle})
	if err != nil {
		t.Fatalf("StrReadF returned error: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty string at EOF, got %q", result)
	}
}

// TestStrReadF_EOFAfterData tests StrReadF returns empty string after all data is read.
// Validates: Requirement 9.3
func TestStrReadF_EOFAfterData(t *testing.T) {
	vm, tmpDir := helperCreateTestVM(t)

	path := filepath.Join(tmpDir, "one_line.dat")
	if err := os.WriteFile(path, []byte("only\n"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	result, err := vm.builtins["OpenF"](vm, []any{"one_line.dat"})
	if err != nil {
		t.Fatalf("OpenF returned error: %v", err)
	}
	handle := result.(int64)
	defer vm.builtins["CloseF"](vm, []any{handle})

	// Read the one line
	_, err = vm.builtins["StrReadF"](vm, []any{handle})
	if err != nil {
		t.Fatalf("StrReadF returned error: %v", err)
	}

	// Next read should return empty string (EOF)
	result, err = vm.builtins["StrReadF"](vm, []any{handle})
	if err != nil {
		t.Fatalf("StrReadF returned error: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty string at EOF, got %q", result)
	}
}

// TestStrReadF_ShiftJIS tests StrReadF converts Shift-JIS to UTF-8.
// Validates: Requirement 9.2
func TestStrReadF_ShiftJIS(t *testing.T) {
	vm, tmpDir := helperCreateTestVM(t)

	// Shift-JIS bytes for "こんにちは" followed by LF
	sjisBytes := []byte{0x82, 0xB1, 0x82, 0xF1, 0x82, 0xC9, 0x82, 0xBF, 0x82, 0xCD, 0x0A}
	path := filepath.Join(tmpDir, "sjis.dat")
	if err := os.WriteFile(path, sjisBytes, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	result, err := vm.builtins["OpenF"](vm, []any{"sjis.dat"})
	if err != nil {
		t.Fatalf("OpenF returned error: %v", err)
	}
	handle := result.(int64)
	defer vm.builtins["CloseF"](vm, []any{handle})

	result, err = vm.builtins["StrReadF"](vm, []any{handle})
	if err != nil {
		t.Fatalf("StrReadF returned error: %v", err)
	}
	if result != "こんにちは" {
		t.Errorf("expected %q, got %q", "こんにちは", result)
	}
}

// TestStrReadF_InvalidHandle tests StrReadF with an invalid handle.
// Validates: Requirement 9.5
func TestStrReadF_InvalidHandle(t *testing.T) {
	vm, _ := helperCreateTestVM(t)

	_, err := vm.builtins["StrReadF"](vm, []any{int64(999)})
	if err == nil {
		t.Fatal("expected error for invalid handle, got nil")
	}
}

// TestStrReadF_MissingArgs tests StrReadF with no arguments.
// Validates: Requirement 9.6
func TestStrReadF_MissingArgs(t *testing.T) {
	vm, _ := helperCreateTestVM(t)

	_, err := vm.builtins["StrReadF"](vm, []any{})
	if err == nil {
		t.Fatal("expected error for missing args, got nil")
	}
}

// --- StrWriteF Tests ---

// TestStrWriteF_NoNewline tests StrWriteF writes string without appending newline.
// Validates: Requirement 10.1
func TestStrWriteF_NoNewline(t *testing.T) {
	vm, tmpDir := helperCreateTestVM(t)

	result, err := vm.builtins["OpenF"](vm, []any{"no_nl.dat", int64(FileCreateNew | FileAccessReadWrite)})
	if err != nil {
		t.Fatalf("OpenF returned error: %v", err)
	}
	handle := result.(int64)
	defer vm.builtins["CloseF"](vm, []any{handle})

	// Write two strings — neither should have a newline appended
	_, err = vm.builtins["StrWriteF"](vm, []any{handle, "AB"})
	if err != nil {
		t.Fatalf("StrWriteF returned error: %v", err)
	}
	_, err = vm.builtins["StrWriteF"](vm, []any{handle, "CD"})
	if err != nil {
		t.Fatalf("StrWriteF returned error: %v", err)
	}

	// Read raw file bytes and verify no newline was inserted
	path := filepath.Join(tmpDir, "no_nl.dat")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	// ASCII is valid Shift-JIS, so "AB" + "CD" → Shift-JIS "ABCD"
	if string(data) != "ABCD" {
		t.Errorf("expected file content %q, got %q", "ABCD", string(data))
	}
}

// TestStrWriteF_ShiftJIS tests StrWriteF converts UTF-8 to Shift-JIS.
// Validates: Requirement 10.2
func TestStrWriteF_ShiftJIS(t *testing.T) {
	vm, tmpDir := helperCreateTestVM(t)

	result, err := vm.builtins["OpenF"](vm, []any{"sjis_out.dat", int64(FileCreateNew | FileAccessReadWrite)})
	if err != nil {
		t.Fatalf("OpenF returned error: %v", err)
	}
	handle := result.(int64)
	defer vm.builtins["CloseF"](vm, []any{handle})

	// Write UTF-8 Japanese string
	_, err = vm.builtins["StrWriteF"](vm, []any{handle, "こんにちは"})
	if err != nil {
		t.Fatalf("StrWriteF returned error: %v", err)
	}

	// Read raw file bytes and verify they are Shift-JIS encoded
	path := filepath.Join(tmpDir, "sjis_out.dat")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	expectedSJIS := []byte{0x82, 0xB1, 0x82, 0xF1, 0x82, 0xC9, 0x82, 0xBF, 0x82, 0xCD}
	if len(data) != len(expectedSJIS) {
		t.Fatalf("expected %d bytes, got %d bytes", len(expectedSJIS), len(data))
	}
	for i, b := range expectedSJIS {
		if data[i] != b {
			t.Errorf("byte %d: expected 0x%02X, got 0x%02X", i, b, data[i])
		}
	}
}

// TestStrWriteF_InvalidHandle tests StrWriteF with an invalid handle.
// Validates: Requirement 10.3
func TestStrWriteF_InvalidHandle(t *testing.T) {
	vm, _ := helperCreateTestVM(t)

	_, err := vm.builtins["StrWriteF"](vm, []any{int64(999), "test"})
	if err == nil {
		t.Fatal("expected error for invalid handle, got nil")
	}
}

// TestStrWriteF_MissingArgs tests StrWriteF with insufficient arguments.
// Validates: Requirement 10.4
func TestStrWriteF_MissingArgs(t *testing.T) {
	vm, _ := helperCreateTestVM(t)

	// No arguments
	_, err := vm.builtins["StrWriteF"](vm, []any{})
	if err == nil {
		t.Fatal("expected error for missing args (0 args), got nil")
	}

	// Only handle, missing string
	_, err = vm.builtins["StrWriteF"](vm, []any{int64(1)})
	if err == nil {
		t.Fatal("expected error for missing args (1 arg), got nil")
	}
}

// --- StrWriteF/StrReadF Round-trip Test ---

// TestStrWriteFStrReadF_RoundTrip tests writing with StrWriteF, adding a newline with WriteF,
// seeking to start, and reading back with StrReadF.
// Validates: Requirements 9.1, 9.2, 10.1, 10.2
func TestStrWriteFStrReadF_RoundTrip(t *testing.T) {
	vm, _ := helperCreateTestVM(t)

	result, err := vm.builtins["OpenF"](vm, []any{"rt_str.dat", int64(FileCreateNew | FileAccessReadWrite)})
	if err != nil {
		t.Fatalf("OpenF returned error: %v", err)
	}
	handle := result.(int64)
	defer vm.builtins["CloseF"](vm, []any{handle})

	// Write a string with StrWriteF
	_, err = vm.builtins["StrWriteF"](vm, []any{handle, "こんにちは"})
	if err != nil {
		t.Fatalf("StrWriteF returned error: %v", err)
	}

	// Write a newline byte with WriteF (LF = 0x0A)
	_, err = vm.builtins["WriteF"](vm, []any{handle, int64(0x0A)})
	if err != nil {
		t.Fatalf("WriteF returned error: %v", err)
	}

	// Seek back to start
	_, err = vm.builtins["SeekF"](vm, []any{handle, int64(0), int64(SeekSet)})
	if err != nil {
		t.Fatalf("SeekF returned error: %v", err)
	}

	// Read back with StrReadF
	result, err = vm.builtins["StrReadF"](vm, []any{handle})
	if err != nil {
		t.Fatalf("StrReadF returned error: %v", err)
	}
	if result != "こんにちは" {
		t.Errorf("round-trip: expected %q, got %q", "こんにちは", result)
	}
}
