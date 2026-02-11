package vm

import (
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/zurustar/son-et/pkg/opcode"
)

// helperOpenTempFile opens a new temp file via OpenF with create-new + read-write mode.
// Returns the integer handle. The file is created in the VM's titlePath directory.
func helperOpenTempFile(vm *VM, filename string) (int64, error) {
	mode := int64(FileCreateNew | FileAccessReadWrite)
	result, err := vm.builtins["OpenF"](vm, []any{filename, mode})
	if err != nil {
		return 0, err
	}
	return result.(int64), nil
}

// Feature: required-builtin-functions, Property 4: ReadF/WriteFバイナリラウンドトリップ
// 任意の整数値valueと任意の有効なsize（1, 2, 4）に対して、WriteF(handle, value, size)で
// 書き込んだ後にファイルポインタを先頭に戻してReadF(handle, size)で読み込むと、
// 書き込んだ値と同じ値が得られる（valueをsizeバイトで表現可能な範囲に制限した場合）。
// **Validates: Requirements 7.1, 7.2, 7.3, 7.4, 8.1, 8.2**
func TestProperty4_ReadFWriteFBinaryRoundTrip(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property: WriteF then ReadF round-trip for size=1 (0-255)
	properties.Property("WriteF/ReadF round-trip for size=1", prop.ForAll(
		func(value int64) bool {
			vm := New([]opcode.OpCode{}, WithTitlePath(t.TempDir()))

			handle, err := helperOpenTempFile(vm, "prop4_s1.dat")
			if err != nil {
				return false
			}
			defer vm.builtins["CloseF"](vm, []any{handle})

			// Write value with size=1
			if _, err := vm.builtins["WriteF"](vm, []any{handle, value, int64(1)}); err != nil {
				return false
			}

			// Seek back to start
			if _, err := vm.builtins["SeekF"](vm, []any{handle, int64(0), int64(SeekSet)}); err != nil {
				return false
			}

			// Read back with size=1
			result, err := vm.builtins["ReadF"](vm, []any{handle, int64(1)})
			if err != nil {
				return false
			}

			return result.(int64) == value
		},
		gen.Int64Range(0, 255),
	))

	// Property: WriteF/ReadF round-trip for size=2 (0-65535)
	properties.Property("WriteF/ReadF round-trip for size=2", prop.ForAll(
		func(value int64) bool {
			vm := New([]opcode.OpCode{}, WithTitlePath(t.TempDir()))

			handle, err := helperOpenTempFile(vm, "prop4_s2.dat")
			if err != nil {
				return false
			}
			defer vm.builtins["CloseF"](vm, []any{handle})

			// Write value with size=2
			if _, err := vm.builtins["WriteF"](vm, []any{handle, value, int64(2)}); err != nil {
				return false
			}

			// Seek back to start
			if _, err := vm.builtins["SeekF"](vm, []any{handle, int64(0), int64(SeekSet)}); err != nil {
				return false
			}

			// Read back with size=2
			result, err := vm.builtins["ReadF"](vm, []any{handle, int64(2)})
			if err != nil {
				return false
			}

			return result.(int64) == value
		},
		gen.Int64Range(0, 65535),
	))

	// Property: WriteF/ReadF round-trip for size=4 (0-4294967295)
	properties.Property("WriteF/ReadF round-trip for size=4", prop.ForAll(
		func(value int64) bool {
			vm := New([]opcode.OpCode{}, WithTitlePath(t.TempDir()))

			handle, err := helperOpenTempFile(vm, "prop4_s4.dat")
			if err != nil {
				return false
			}
			defer vm.builtins["CloseF"](vm, []any{handle})

			// Write value with size=4
			if _, err := vm.builtins["WriteF"](vm, []any{handle, value, int64(4)}); err != nil {
				return false
			}

			// Seek back to start
			if _, err := vm.builtins["SeekF"](vm, []any{handle, int64(0), int64(SeekSet)}); err != nil {
				return false
			}

			// Read back with size=4
			result, err := vm.builtins["ReadF"](vm, []any{handle, int64(4)})
			if err != nil {
				return false
			}

			return result.(int64) == value
		},
		gen.Int64Range(0, 4294967295),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: required-builtin-functions, Property 6: SeekFによるランダムアクセスの整合性
// 任意のバイト列が書き込まれたファイルと任意の有効なオフセットに対して、
// SeekF(handle, offset, 0)でファイルポインタを移動した後にReadF(handle, 1)で
// 読み込むと、そのオフセット位置のバイト値が得られる。
// **Validates: Requirements 6.1, 6.2**
func TestProperty6_SeekFRandomAccessConsistency(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property: SeekF to any valid offset then ReadF(1) returns the byte at that offset
	properties.Property("SeekF random access returns correct byte at any offset", prop.ForAll(
		func(data []byte, offsetIdx int) bool {
			if len(data) == 0 {
				return true // skip empty data
			}

			vm := New([]opcode.OpCode{}, WithTitlePath(t.TempDir()))

			handle, err := helperOpenTempFile(vm, "prop6.dat")
			if err != nil {
				return false
			}
			defer vm.builtins["CloseF"](vm, []any{handle})

			// Write each byte using WriteF(handle, value, 1)
			for _, b := range data {
				if _, err := vm.builtins["WriteF"](vm, []any{handle, int64(b), int64(1)}); err != nil {
					return false
				}
			}

			// Pick a valid offset within the written data
			offset := int64(offsetIdx % len(data))

			// Seek to that offset from file start
			if _, err := vm.builtins["SeekF"](vm, []any{handle, offset, int64(SeekSet)}); err != nil {
				return false
			}

			// Read 1 byte
			result, err := vm.builtins["ReadF"](vm, []any{handle, int64(1)})
			if err != nil {
				return false
			}

			// The read byte must match the byte at that offset in the original data
			return result.(int64) == int64(data[offset])
		},
		gen.SliceOfN(50, gen.UInt8()).SuchThat(func(v interface{}) bool {
			return len(v.([]byte)) > 0
		}).WithShrinker(nil),
		gen.IntRange(0, 49),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: required-builtin-functions, Property 5: StrWriteF/StrReadFテキストラウンドトリップ
// 任意の改行を含まないShift-JIS変換可能な文字列に対して、StrWriteF(handle, str)で書き込み、
// 改行バイト(0x0a)をWriteFで追記し、ファイルポインタを先頭に戻してStrReadF(handle)で
// 読み込むと、元の文字列と同じ値が得られる。
// **Validates: Requirements 9.1, 9.2, 10.1, 10.2**
func TestProperty5_StrWriteFStrReadFTextRoundTrip(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Generator for ASCII printable strings (0x20-0x7E), no CR/LF.
	// These characters round-trip perfectly through Shift-JIS.
	asciiPrintableGen := gen.SliceOfN(50, gen.IntRange(0x20, 0x7E)).
		SuchThat(func(v interface{}) bool {
			return len(v.([]int)) > 0
		}).
		Map(func(ints []int) string {
			runes := make([]rune, len(ints))
			for i, c := range ints {
				runes[i] = rune(c)
			}
			return string(runes)
		})

	// Property: StrWriteF then StrReadF round-trip for ASCII printable strings
	properties.Property("StrWriteF/StrReadF round-trip for ASCII printable strings", prop.ForAll(
		func(str string) bool {
			vm := New([]opcode.OpCode{}, WithTitlePath(t.TempDir()))

			handle, err := helperOpenTempFile(vm, "prop5_ascii.dat")
			if err != nil {
				return false
			}
			defer vm.builtins["CloseF"](vm, []any{handle})

			// StrWriteF the string
			if _, err := vm.builtins["StrWriteF"](vm, []any{handle, str}); err != nil {
				return false
			}

			// WriteF a LF byte (0x0A) as line delimiter
			if _, err := vm.builtins["WriteF"](vm, []any{handle, int64(0x0A)}); err != nil {
				return false
			}

			// SeekF to start
			if _, err := vm.builtins["SeekF"](vm, []any{handle, int64(0), int64(SeekSet)}); err != nil {
				return false
			}

			// StrReadF and compare
			result, err := vm.builtins["StrReadF"](vm, []any{handle})
			if err != nil {
				return false
			}

			return result.(string) == str
		},
		asciiPrintableGen,
	))

	// Generator for strings containing Japanese hiragana (あ-ん, U+3041-U+3093).
	// These are all valid in Shift-JIS and round-trip correctly.
	hiraganaGen := gen.SliceOfN(20, gen.IntRange(0x3041, 0x3093)).
		SuchThat(func(v interface{}) bool {
			return len(v.([]int)) > 0
		}).
		Map(func(ints []int) string {
			runes := make([]rune, len(ints))
			for i, c := range ints {
				runes[i] = rune(c)
			}
			return string(runes)
		})

	// Property: StrWriteF then StrReadF round-trip for Japanese hiragana strings
	properties.Property("StrWriteF/StrReadF round-trip for hiragana strings", prop.ForAll(
		func(str string) bool {
			vm := New([]opcode.OpCode{}, WithTitlePath(t.TempDir()))

			handle, err := helperOpenTempFile(vm, "prop5_hira.dat")
			if err != nil {
				return false
			}
			defer vm.builtins["CloseF"](vm, []any{handle})

			// StrWriteF the string
			if _, err := vm.builtins["StrWriteF"](vm, []any{handle, str}); err != nil {
				return false
			}

			// WriteF a LF byte (0x0A) as line delimiter
			if _, err := vm.builtins["WriteF"](vm, []any{handle, int64(0x0A)}); err != nil {
				return false
			}

			// SeekF to start
			if _, err := vm.builtins["SeekF"](vm, []any{handle, int64(0), int64(SeekSet)}); err != nil {
				return false
			}

			// StrReadF and compare
			result, err := vm.builtins["StrReadF"](vm, []any{handle})
			if err != nil {
				return false
			}

			return result.(string) == str
		},
		hiraganaGen,
	))

	// Property: StrWriteF then StrReadF round-trip for mixed ASCII + hiragana strings
	properties.Property("StrWriteF/StrReadF round-trip for mixed ASCII+hiragana strings", prop.ForAll(
		func(asciiPart string, hiraganaPart string) bool {
			mixed := asciiPart + hiraganaPart
			if len(mixed) == 0 {
				return true
			}

			vm := New([]opcode.OpCode{}, WithTitlePath(t.TempDir()))

			handle, err := helperOpenTempFile(vm, "prop5_mixed.dat")
			if err != nil {
				return false
			}
			defer vm.builtins["CloseF"](vm, []any{handle})

			// StrWriteF the string
			if _, err := vm.builtins["StrWriteF"](vm, []any{handle, mixed}); err != nil {
				return false
			}

			// WriteF a LF byte (0x0A) as line delimiter
			if _, err := vm.builtins["WriteF"](vm, []any{handle, int64(0x0A)}); err != nil {
				return false
			}

			// SeekF to start
			if _, err := vm.builtins["SeekF"](vm, []any{handle, int64(0), int64(SeekSet)}); err != nil {
				return false
			}

			// StrReadF and compare
			result, err := vm.builtins["StrReadF"](vm, []any{handle})
			if err != nil {
				return false
			}

			return result.(string) == mixed
		},
		gen.SliceOfN(10, gen.IntRange(0x20, 0x7E)).Map(func(ints []int) string {
			runes := make([]rune, len(ints))
			for i, c := range ints {
				runes[i] = rune(c)
			}
			return string(runes)
		}),
		gen.SliceOfN(10, gen.IntRange(0x3041, 0x3093)).Map(func(ints []int) string {
			runes := make([]rune, len(ints))
			for i, c := range ints {
				runes[i] = rune(c)
			}
			return string(runes)
		}),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
