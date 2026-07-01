package vm

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/zurustar/son-et/pkg/opcode"
)

// TestMixedReadFAndStrReadF is a regression test for finding P2-b: ReadF (binary)
// and StrReadF (line) must share the same file position so they can be
// interleaved without one skipping data the other buffered.
func TestMixedReadFAndStrReadF(t *testing.T) {
	dir := t.TempDir()
	// Layout: 1 byte 0x41('A'), then "line1\n", then 1 byte 0x42('B').
	content := append([]byte{0x41}, []byte("line1\n")...)
	content = append(content, 0x42)
	if err := os.WriteFile(filepath.Join(dir, "mix.dat"), content, 0644); err != nil {
		t.Fatal(err)
	}

	vm := New([]opcode.OpCode{}, WithTitlePath(dir))

	h, err := vm.builtins["OpenF"](vm, []any{"mix.dat", int64(FileAccessReadOnly)})
	if err != nil {
		t.Fatalf("OpenF: %v", err)
	}
	handle := h.(int64)
	defer vm.builtins["CloseF"](vm, []any{handle})

	// ReadF size=1 → 0x41
	r1, err := vm.builtins["ReadF"](vm, []any{handle, int64(1)})
	if err != nil {
		t.Fatalf("ReadF#1: %v", err)
	}
	if v, _ := toInt64(r1); v != 0x41 {
		t.Fatalf("ReadF#1 = 0x%X, want 0x41", v)
	}

	// StrReadF → "line1" (must start right after the first byte, not skip ahead)
	s, err := vm.builtins["StrReadF"](vm, []any{handle})
	if err != nil {
		t.Fatalf("StrReadF: %v", err)
	}
	if s.(string) != "line1" {
		t.Fatalf("StrReadF = %q, want %q (ReadF/StrReadF desynced)", s, "line1")
	}

	// ReadF size=1 → 0x42 (must be exactly after the newline)
	r2, err := vm.builtins["ReadF"](vm, []any{handle, int64(1)})
	if err != nil {
		t.Fatalf("ReadF#2: %v", err)
	}
	if v, _ := toInt64(r2); v != 0x42 {
		t.Fatalf("ReadF#2 = 0x%X, want 0x42 (position drifted after StrReadF)", v)
	}
}
