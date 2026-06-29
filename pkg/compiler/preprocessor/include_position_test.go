package preprocessor

import (
	"strings"
	"testing"
	"testing/fstest"
)

// TestCommentedIncludeNotExpanded is a regression test for the bug where an
// "#include" appearing inside a comment was matched by a naive textual search,
// corrupting the output and leaving the real directive unprocessed.
// See docs/bug-hunt-findings.md finding B.
func TestCommentedIncludeNotExpanded(t *testing.T) {
	mfs := fstest.MapFS{
		"main.tfy": {Data: []byte("// see #include \"fake.tfy\"\n#include \"real.tfy\"\nint after;\n")},
		"real.tfy": {Data: []byte("REAL_CONTENT;\n")},
		"fake.tfy": {Data: []byte("FAKE_CONTENT;\n")},
	}
	p := NewWithFS("", mfs)
	res, err := p.PreprocessFile("main.tfy")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The real include must be expanded.
	if !strings.Contains(res.Source, "REAL_CONTENT;") {
		t.Errorf("real.tfy was not expanded:\n%s", res.Source)
	}
	// The commented-out fake include must NOT be expanded.
	if strings.Contains(res.Source, "FAKE_CONTENT;") {
		t.Errorf("fake.tfy inside a comment was wrongly expanded:\n%s", res.Source)
	}
	// The literal directive text must not survive in the output.
	if strings.Contains(res.Source, "#include \"real.tfy\"") {
		t.Errorf("real #include directive left unprocessed:\n%s", res.Source)
	}
	// The comment line itself should be preserved intact.
	if !strings.Contains(res.Source, "// see #include \"fake.tfy\"") {
		t.Errorf("comment line was corrupted:\n%s", res.Source)
	}
	// Content after the include should be preserved.
	if !strings.Contains(res.Source, "int after;") {
		t.Errorf("trailing content lost:\n%s", res.Source)
	}
	for _, f := range res.IncludedFiles {
		if f == "fake.tfy" {
			t.Errorf("fake.tfy must not be in IncludedFiles: %v", res.IncludedFiles)
		}
	}
}

// TestMultipleIncludesStillWork ensures the new position logic preserves the
// normal multi-include behavior (no regression).
func TestMultipleIncludesStillWork(t *testing.T) {
	mfs := fstest.MapFS{
		"main.tfy": {Data: []byte("int x;\n#include \"a.tfy\"\nint y;\n#include \"b.tfy\"\nint z;\n")},
		"a.tfy":    {Data: []byte("A;\n")},
		"b.tfy":    {Data: []byte("B;\n")},
	}
	p := NewWithFS("", mfs)
	res, err := p.PreprocessFile("main.tfy")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, want := range []string{"int x;", "A;", "int y;", "B;", "int z;"} {
		if !strings.Contains(res.Source, want) {
			t.Errorf("missing %q in output:\n%s", want, res.Source)
		}
	}
	if strings.Contains(res.Source, "#include") {
		t.Errorf("an #include directive was left unprocessed:\n%s", res.Source)
	}
}
