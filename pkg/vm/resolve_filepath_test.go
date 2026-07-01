package vm

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/zurustar/son-et/pkg/opcode"
)

// TestResolveFilePathConfinement is a regression test for the path-traversal
// vulnerability (docs/bug-hunt-findings.md finding P2-a): resolveFilePath must
// confine script-provided paths to the title directory, rejecting absolute
// paths and ".." escapes.
func TestResolveFilePathConfinement(t *testing.T) {
	titleDir := t.TempDir()
	vm := New([]opcode.OpCode{}, WithTitlePath(titleDir))

	t.Run("valid relative path is allowed and confined", func(t *testing.T) {
		got, err := vm.resolveFilePath("data/save.dat")
		if err != nil {
			t.Fatalf("unexpected error for valid path: %v", err)
		}
		want := filepath.Join(titleDir, "data", "save.dat")
		if got != want {
			t.Errorf("resolved path = %q, want %q", got, want)
		}
	})

	t.Run("absolute path is rejected", func(t *testing.T) {
		abs := "/etc/passwd"
		if runtime.GOOS == "windows" {
			abs = `C:\Windows\System32\drivers\etc\hosts`
		}
		if _, err := vm.resolveFilePath(abs); err == nil {
			t.Errorf("expected error for absolute path %q, got nil", abs)
		}
	})

	t.Run("dotdot escape is rejected", func(t *testing.T) {
		for _, p := range []string{
			"../../../../etc/passwd",
			"..",
			"sub/../../outside.txt",
		} {
			if _, err := vm.resolveFilePath(p); err == nil {
				t.Errorf("expected error for escaping path %q, got nil", p)
			}
		}
	})

	t.Run("dotdot that stays inside is allowed", func(t *testing.T) {
		// a/../b resolves to <title>/b — still inside, should be allowed
		got, err := vm.resolveFilePath("a/../b.dat")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.HasPrefix(got, titleDir) {
			t.Errorf("resolved path %q is not within title dir %q", got, titleDir)
		}
	})
}
