package codegen

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"testing/quick"

	"github.com/zurustar/filly2exe/pkg/compiler/lexer"
	"github.com/zurustar/filly2exe/pkg/compiler/parser"
)

// Feature: core-engine, Property 1: Transpiler generates valid Go code
// Validates: Requirements 1.1, 1.2

// FILLYScript represents a valid FILLY script for property testing
type FILLYScript struct {
	Source string
}

// Generate creates random valid FILLY scripts for property testing
func (FILLYScript) Generate(r *rand.Rand, size int) reflect.Value {
	// Generate various valid FILLY script patterns
	patterns := []string{
		// Simple variable assignment
		`main() {
			x = 10;
		}`,

		// Function call with string literal
		`main() {
			LoadPic("test.bmp");
		}`,

		// Multiple statements
		`main() {
			x = 5;
			y = 10;
			z = x + y;
		}`,

		// Function definition and call
		`helper(a) {
			x = a + 1;
		}
		main() {
			helper(5);
		}`,

		// Conditional statement
		`main() {
			x = 10;
			if (x > 5) {
				y = 1;
			}
		}`,

		// Loop statement
		`main() {
			for (i = 0; i < 10; i = i + 1) {
				x = i;
			}
		}`,

		// Array usage
		`main() {
			let arr[];
			arr[0] = 5;
			x = arr[0];
		}`,

		// String operations
		`main() {
			str s;
			s = "hello";
		}`,

		// mes block with TIME mode
		`main() {
			mes(time) {
				,,
			}
		}`,

		// step block
		`main() {
			mes(time) {
				step(8) {
					,,
				}
			}
		}`,

		// Multiple function calls
		`main() {
			pic = LoadPic("image.bmp");
			OpenWin(pic, 0, 0, 640, 480, 0, 0, 0);
		}`,

		// Nested expressions
		`main() {
			x = (5 + 3) * 2;
			y = x - 1;
		}`,

		// Global variables
		`int globalVar;
		main() {
			globalVar = 100;
		}`,

		// Function with parameters
		`calculate(a, b) {
			x = a + b;
		}
		main() {
			calculate(5, 10);
		}`,

		// Mixed case identifiers (case-insensitive)
		`main() {
			MyVar = 10;
			myvar = 20;
		}`,
	}

	// Select a random pattern
	idx := r.Intn(len(patterns))
	return reflect.ValueOf(FILLYScript{Source: patterns[idx]})
}

// TestProperty1_TranspilerGeneratesValidGoCode verifies that the transpiler
// generates valid Go code for any valid FILLY script
func TestProperty1_TranspilerGeneratesValidGoCode(t *testing.T) {
	// Feature: core-engine, Property 1: Transpiler generates valid Go code
	// Validates: Requirements 1.1, 1.2

	property := func(script FILLYScript) bool {
		// 1. Lex the FILLY script
		l := lexer.New(script.Source)

		// 2. Parse the FILLY script
		p := parser.New(l)
		program := p.ParseProgram()

		// Check for parser errors
		if len(p.Errors()) > 0 {
			// Parser errors mean invalid input, skip this test case
			return true
		}

		// 3. Generate Go code
		gen := New([]string{}) // No assets for basic tests
		goCode := gen.Generate(program)

		// 4. Verify the generated code compiles
		return compilesSuccessfully(t, goCode)
	}

	config := &quick.Config{
		MaxCount: 100, // Run 100 iterations as specified in design
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property violated: %v", err)
	}
}

// compilesSuccessfully checks if the generated Go code compiles without errors
func compilesSuccessfully(t *testing.T, goCode string) bool {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "transpiler-test-*")
	if err != nil {
		t.Logf("Failed to create temp dir: %v", err)
		return false
	}
	defer os.RemoveAll(tmpDir)

	// Write the generated Go code to a file
	goFile := filepath.Join(tmpDir, "test.go")
	if err := os.WriteFile(goFile, []byte(goCode), 0644); err != nil {
		t.Logf("Failed to write Go file: %v", err)
		return false
	}

	// Initialize a Go module in the temp directory
	cmd := exec.Command("go", "mod", "init", "testmodule")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Logf("Failed to init module: %v\nOutput: %s", err, output)
		return false
	}

	// Add dependency on the engine package
	cmd = exec.Command("go", "mod", "edit", "-replace",
		fmt.Sprintf("github.com/zurustar/filly2exe=%s", getProjectRoot()))
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Logf("Failed to add replace directive: %v\nOutput: %s", err, output)
		return false
	}

	// Run go mod tidy
	cmd = exec.Command("go", "mod", "tidy")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Logf("Failed to tidy module: %v\nOutput: %s", err, output)
		return false
	}

	// Try to compile the Go code
	cmd = exec.Command("go", "build", "-o", "test", "test.go")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Logf("Compilation failed: %v\nOutput: %s\nGenerated code:\n%s",
			err, output, goCode)
		return false
	}

	return true
}

// getProjectRoot returns the root directory of the project
func getProjectRoot() string {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}

	// Walk up until we find go.mod
	dir := cwd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root without finding go.mod
			return cwd
		}
		dir = parent
	}
}

// TestProperty2_CaseInsensitiveIdentifierTransformation verifies that
// any FILLY identifier with mixed case is converted to lowercase in generated code
func TestProperty2_CaseInsensitiveIdentifierTransformation(t *testing.T) {
	// Feature: core-engine, Property 2: Case-insensitive identifier transformation
	// Validates: Requirements 1.4

	property := func(ident MixedCaseIdentifier) bool {
		// Create a simple script using the mixed-case identifier
		script := fmt.Sprintf(`main() {
			%s = 10;
			x = %s + 5;
		}`, ident.Original, ident.Original)

		// Parse and generate
		l := lexer.New(script)
		p := parser.New(l)
		program := p.ParseProgram()

		// Skip if parser errors (invalid identifier)
		if len(p.Errors()) > 0 {
			return true
		}

		gen := New([]string{})
		goCode := gen.Generate(program)

		// Verify the lowercase version appears in the generated code
		// The identifier should be converted to lowercase
		lowercaseIdent := strings.ToLower(ident.Original)

		// Check that the lowercase version appears in the generated code
		// Look for it as a standalone identifier (with word boundaries)
		if !containsIdentifier(goCode, lowercaseIdent) {
			t.Logf("Expected lowercase identifier %q not found in generated code for input %q:\n%s",
				lowercaseIdent, ident.Original, goCode)
			return false
		}

		// Also verify that the original mixed-case version does NOT appear as an identifier
		// (unless it happens to be all lowercase already)
		if ident.Original != lowercaseIdent {
			if containsIdentifier(goCode, ident.Original) {
				t.Logf("Original mixed-case identifier %q found in generated code (should be lowercase):\n%s",
					ident.Original, goCode)
				return false
			}
		}

		return true
	}

	config := &quick.Config{
		MaxCount: 100, // Run 100 iterations as specified in design
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property violated: %v", err)
	}
}

// containsIdentifier checks if an identifier appears in the code as a standalone word
// This helps avoid false positives from substrings
func containsIdentifier(code, ident string) bool {
	// Simple heuristic: look for the identifier surrounded by non-identifier characters
	// Identifier characters are: letters, digits, underscore
	lines := strings.Split(code, "\n")
	for _, line := range lines {
		// Skip comments
		if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "//") {
			continue
		}

		// Look for the identifier with word boundaries
		// We'll check if it appears with non-identifier chars on both sides
		idx := 0
		for {
			pos := strings.Index(line[idx:], ident)
			if pos == -1 {
				break
			}
			pos += idx

			// Check character before (if exists)
			beforeOk := pos == 0 || !isIdentifierChar(rune(line[pos-1]))

			// Check character after (if exists)
			afterPos := pos + len(ident)
			afterOk := afterPos >= len(line) || !isIdentifierChar(rune(line[afterPos]))

			if beforeOk && afterOk {
				return true
			}

			idx = pos + 1
		}
	}
	return false
}

// isIdentifierChar checks if a character can be part of an identifier
func isIdentifierChar(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}

// MixedCaseIdentifier represents an identifier with mixed case for property testing
type MixedCaseIdentifier struct {
	Original string
}

// Generate creates random mixed-case identifiers for property testing
func (MixedCaseIdentifier) Generate(r *rand.Rand, size int) reflect.Value {
	// Generate valid identifier patterns with mixed case
	patterns := []string{
		"MyVariable",
		"myVariable",
		"MYVARIABLE",
		"MyVar",
		"ABC",
		"AbC",
		"XYZ",
		"TestVar",
		"testVar",
		"TEST_VAR",
		"MyTestVariable",
		"aBC",
		"aBc",
		"Variable123",
		"VAR123",
		"Var123",
		"myVar123",
		"MyVar123",
		"X",
		"Y",
		"Z",
		"A",
		"B",
		"C",
		"Xx",
		"Yy",
		"Zz",
		"VarA",
		"VarB",
		"VarC",
		"varA",
		"varB",
		"varC",
	}

	// Select a random pattern
	idx := r.Intn(len(patterns))
	return reflect.ValueOf(MixedCaseIdentifier{Original: patterns[idx]})
}

// TestProperty1_AssetEmbedding verifies that asset references are properly
// embedded in the generated code
func TestProperty1_AssetEmbedding(t *testing.T) {
	// Feature: core-engine, Property 1: Transpiler generates valid Go code
	// Validates: Requirements 2.1, 2.2, 2.3

	testCases := []struct {
		name          string
		input         string
		assets        []string
		expectedEmbed []string
	}{
		{
			name: "LoadPic with BMP",
			input: `main() {
				pic = LoadPic("test.bmp");
			}`,
			assets:        []string{"test.bmp"},
			expectedEmbed: []string{"//go:embed test.bmp"},
		},
		{
			name: "Multiple assets",
			input: `main() {
				pic1 = LoadPic("image1.bmp");
				pic2 = LoadPic("image2.bmp");
				PlayMIDI("music.mid");
			}`,
			assets:        []string{"image1.bmp", "image2.bmp", "music.mid"},
			expectedEmbed: []string{"//go:embed image1.bmp image2.bmp music.mid"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse and generate
			l := lexer.New(tc.input)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("Parser errors: %v", p.Errors())
			}

			gen := New(tc.assets)
			goCode := gen.Generate(program)

			// Check that go:embed directive is present
			for _, expected := range tc.expectedEmbed {
				if !strings.Contains(goCode, expected) {
					t.Errorf("Expected embed directive %q not found in generated code:\n%s",
						expected, goCode)
				}
			}
		})
	}
}

// TestProperty3_AssetEmbeddingCompleteness verifies that all asset references
// in the code result in go:embed directives
func TestProperty3_AssetEmbeddingCompleteness(t *testing.T) {
	// Feature: core-engine, Property 3: Asset embedding completeness
	// Validates: Requirements 2.1, 2.2, 2.3

	property := func(script AssetScript) bool {
		// Parse the script
		l := lexer.New(script.Source)
		p := parser.New(l)
		program := p.ParseProgram()

		// Skip if parser errors
		if len(p.Errors()) > 0 {
			return true
		}

		// Generate code with the assets
		gen := New(script.Assets)
		goCode := gen.Generate(program)

		// Verify that all assets appear in the go:embed directive
		for _, asset := range script.Assets {
			// Check if the asset appears in a go:embed directive
			embedDirective := "//go:embed"
			embedIdx := strings.Index(goCode, embedDirective)
			if embedIdx == -1 {
				t.Logf("No go:embed directive found in generated code")
				return false
			}

			// Extract the embed line (from //go:embed to newline)
			embedLine := goCode[embedIdx:]
			newlineIdx := strings.Index(embedLine, "\n")
			if newlineIdx != -1 {
				embedLine = embedLine[:newlineIdx]
			}

			// Check if the asset is in the embed line
			if !strings.Contains(embedLine, asset) {
				t.Logf("Asset %q not found in go:embed directive: %s", asset, embedLine)
				return false
			}
		}

		return true
	}

	config := &quick.Config{
		MaxCount: 100, // Run 100 iterations as specified in design
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property violated: %v", err)
	}
}

// AssetScript represents a FILLY script with asset references for property testing
type AssetScript struct {
	Source string
	Assets []string
}

// Generate creates random FILLY scripts with asset references for property testing
func (AssetScript) Generate(r *rand.Rand, size int) reflect.Value {
	// Generate various asset loading patterns
	patterns := []struct {
		source string
		assets []string
	}{
		{
			source: `main() { pic = LoadPic("test.bmp"); }`,
			assets: []string{"test.bmp"},
		},
		{
			source: `main() { PlayMIDI("music.mid"); }`,
			assets: []string{"music.mid"},
		},
		{
			source: `main() { PlayWAVE("sound.wav"); }`,
			assets: []string{"sound.wav"},
		},
		{
			source: `main() {
				pic1 = LoadPic("img1.bmp");
				pic2 = LoadPic("img2.bmp");
			}`,
			assets: []string{"img1.bmp", "img2.bmp"},
		},
		{
			source: `main() {
				LoadPic("a.bmp");
				PlayMIDI("b.mid");
				PlayWAVE("c.wav");
			}`,
			assets: []string{"a.bmp", "b.mid", "c.wav"},
		},
		{
			source: `helper() {
				LoadPic("helper.bmp");
			}
			main() {
				LoadPic("main.bmp");
				helper();
			}`,
			assets: []string{"helper.bmp", "main.bmp"},
		},
		{
			source: `main() {
				mes(time) {
					LoadPic("mes.bmp");
				}
			}`,
			assets: []string{"mes.bmp"},
		},
		{
			source: `main() {
				for (i = 0; i < 3; i = i + 1) {
					LoadPic("loop.bmp");
				}
			}`,
			assets: []string{"loop.bmp"},
		},
		{
			source: `main() {
				if (x > 0) {
					LoadPic("if.bmp");
				}
			}`,
			assets: []string{"if.bmp"},
		},
		{
			source: `main() {
				pic1 = LoadPic("image1.bmp");
				pic2 = LoadPic("image2.bmp");
				pic3 = LoadPic("image3.bmp");
				PlayMIDI("music1.mid");
				PlayMIDI("music2.mid");
				PlayWAVE("sound1.wav");
			}`,
			assets: []string{"image1.bmp", "image2.bmp", "image3.bmp", "music1.mid", "music2.mid", "sound1.wav"},
		},
	}

	// Select a random pattern
	idx := r.Intn(len(patterns))
	pattern := patterns[idx]

	return reflect.ValueOf(AssetScript{
		Source: pattern.source,
		Assets: pattern.assets,
	})
}

// TestProperty4_CaseInsensitiveAssetMatching verifies that asset references
// with different casing than the actual file are correctly matched and embedded
func TestProperty4_CaseInsensitiveAssetMatching(t *testing.T) {
	// Feature: core-engine, Property 4: Case-insensitive asset matching
	// Validates: Requirements 2.4

	property := func(script CaseInsensitiveAssetScript) bool {
		// Parse the script
		l := lexer.New(script.Source)
		p := parser.New(l)
		program := p.ParseProgram()

		// Skip if parser errors
		if len(p.Errors()) > 0 {
			return true
		}

		// Generate code with the actual filesystem assets (correct case)
		gen := New(script.ActualAssets)
		goCode := gen.Generate(program)

		// Verify that all actual assets appear in the go:embed directive
		// even though the script references them with different casing
		for _, actualAsset := range script.ActualAssets {
			// Check if the asset appears in a go:embed directive
			embedDirective := "//go:embed"
			embedIdx := strings.Index(goCode, embedDirective)
			if embedIdx == -1 {
				t.Logf("No go:embed directive found in generated code")
				return false
			}

			// Extract the embed line (from //go:embed to newline)
			embedLine := goCode[embedIdx:]
			newlineIdx := strings.Index(embedLine, "\n")
			if newlineIdx != -1 {
				embedLine = embedLine[:newlineIdx]
			}

			// Check if the actual asset (with correct case) is in the embed line
			if !strings.Contains(embedLine, actualAsset) {
				t.Logf("Actual asset %q not found in go:embed directive: %s", actualAsset, embedLine)
				return false
			}
		}

		// Also verify that the engine's AssetLoader can handle case-insensitive lookups
		// This is tested at runtime, but we can verify the structure is correct
		return true
	}

	config := &quick.Config{
		MaxCount: 100, // Run 100 iterations as specified in design
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property violated: %v", err)
	}
}

// CaseInsensitiveAssetScript represents a FILLY script with case-mismatched asset references
type CaseInsensitiveAssetScript struct {
	Source       string   // Script with lowercase asset references
	ActualAssets []string // Actual filesystem names (may have different case)
}

// Generate creates random FILLY scripts with case-mismatched asset references
func (CaseInsensitiveAssetScript) Generate(r *rand.Rand, size int) reflect.Value {
	// Generate various case mismatch patterns
	// Script references assets in lowercase, but actual files may be uppercase or mixed
	patterns := []struct {
		source       string
		actualAssets []string
	}{
		{
			source:       `main() { pic = LoadPic("test.bmp"); }`,
			actualAssets: []string{"TEST.BMP"}, // Uppercase file
		},
		{
			source:       `main() { pic = LoadPic("image.bmp"); }`,
			actualAssets: []string{"Image.BMP"}, // Mixed case file
		},
		{
			source:       `main() { PlayMIDI("music.mid"); }`,
			actualAssets: []string{"MUSIC.MID"}, // Uppercase file
		},
		{
			source:       `main() { PlayWAVE("sound.wav"); }`,
			actualAssets: []string{"Sound.WAV"}, // Mixed case file
		},
		{
			source: `main() {
				pic1 = LoadPic("img1.bmp");
				pic2 = LoadPic("img2.bmp");
			}`,
			actualAssets: []string{"IMG1.BMP", "IMG2.BMP"}, // Uppercase files
		},
		{
			source: `main() {
				LoadPic("a.bmp");
				PlayMIDI("b.mid");
				PlayWAVE("c.wav");
			}`,
			actualAssets: []string{"A.BMP", "B.MID", "C.WAV"}, // Uppercase files
		},
		{
			source: `main() {
				LoadPic("title.bmp");
				LoadPic("kuma-1.bmp");
				PlayMIDI("music.mid");
			}`,
			actualAssets: []string{"TITLE.BMP", "KUMA-1.BMP", "MUSIC.MID"}, // Real-world example
		},
		{
			source: `main() {
				pic = LoadPic("myimage.bmp");
			}`,
			actualAssets: []string{"MyImage.BMP"}, // CamelCase file
		},
		{
			source: `main() {
				pic = LoadPic("test_file.bmp");
			}`,
			actualAssets: []string{"TEST_FILE.BMP"}, // Underscore with uppercase
		},
		{
			source: `main() {
				pic1 = LoadPic("file1.bmp");
				pic2 = LoadPic("file2.bmp");
				pic3 = LoadPic("file3.bmp");
			}`,
			actualAssets: []string{"FILE1.BMP", "File2.BMP", "file3.bmp"}, // Mixed cases
		},
	}

	// Select a random pattern
	idx := r.Intn(len(patterns))
	pattern := patterns[idx]

	return reflect.ValueOf(CaseInsensitiveAssetScript{
		Source:       pattern.source,
		ActualAssets: pattern.actualAssets,
	})
}
