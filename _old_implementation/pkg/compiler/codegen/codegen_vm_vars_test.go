package codegen

import (
	"strings"
	"testing"

	"github.com/zurustar/son-et/pkg/compiler/ast"
	"github.com/zurustar/son-et/pkg/compiler/lexer"
	"github.com/zurustar/son-et/pkg/compiler/parser"
)

// TestScanMesBlocksForVMVars tests that variables used in mes() blocks are correctly identified
func TestScanMesBlocksForVMVars(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name: "simple variable in mes block",
			input: `
main() {
	x = 10
	mes(TIME) {
		OpenWin(x, 0, 0, 640, 480, 0, 0, 0)
	}
}`,
			expected: []string{"x"},
		},
		{
			name: "multiple variables in mes block",
			input: `
main() {
	winW = WinInfo(0)
	winH = WinInfo(1)
	mes(MIDI_TIME) {
		OpenWin(0, winW-320, winH-240, 640, 480, 0, 0, 0)
	}
}`,
			expected: []string{"winh", "winw"},
		},
		{
			name: "variables in nested expressions",
			input: `
main() {
	a = 100
	b = 200
	c = 300
	mes(TIME) {
		MoveCast(1, 0, a+b, c*2, 0, 0, 0, 0)
	}
}`,
			expected: []string{"a", "b", "c"},
		},
		{
			name: "multiple mes blocks",
			input: `
main() {
	x = 10
	y = 20
	z = 30
	mes(TIME) {
		OpenWin(x, 0, 0, 640, 480, 0, 0, 0)
	}
	mes(MIDI_TIME) {
		MoveCast(1, 0, y, z, 0, 0, 0, 0)
	}
}`,
			expected: []string{"x", "y", "z"},
		},
		{
			name: "no mes blocks",
			input: `
main() {
	x = 10
	OpenWin(x, 0, 0, 640, 480, 0, 0, 0)
}`,
			expected: []string{},
		},
		{
			name: "variables in step blocks inside mes",
			input: `
main() {
	pic = LoadPic("test.bmp")
	mes(TIME) {
		step(8) {
			MoveCast(1, pic, 100, 100, 0, 0, 0, 0)
		}
	}
}`,
			expected: []string{"pic"},
		},
		{
			name: "variables in control flow inside mes",
			input: `
main() {
	x = 10
	y = 20
	mes(TIME) {
		if (x > 5) {
			MoveCast(1, 0, y, 0, 0, 0, 0, 0)
		}
	}
}`,
			expected: []string{"x", "y"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("Parser errors: %v", p.Errors())
			}

			g := New([]string{})

			// Find main function
			var mainFunc *ast.FunctionStatement
			for _, stmt := range program.Statements {
				if fn, ok := stmt.(*ast.FunctionStatement); ok && fn.Name.Value == "main" {
					mainFunc = fn
					break
				}
			}

			if mainFunc == nil {
				t.Fatal("No main function found")
			}

			// Scan for VM variables
			vmVars := g.scanMesBlocksForVMVars(mainFunc.Body)

			// Convert to sorted slice for comparison
			var result []string
			for v := range vmVars {
				result = append(result, v)
			}

			// Sort both slices for comparison
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d variables, got %d: %v", len(tt.expected), len(result), result)
				return
			}

			// Check each expected variable is present
			for _, exp := range tt.expected {
				if !vmVars[exp] {
					t.Errorf("Expected variable %q not found in vmVars", exp)
				}
			}
		})
	}
}

// TestVMVarsInGeneratedCode tests that the generator properly tracks VM variables
func TestVMVarsInGeneratedCode(t *testing.T) {
	input := `
main() {
	winW = WinInfo(0)
	winH = WinInfo(1)
	p39 = LoadPic("P39.BMP")
	
	mes(MIDI_TIME) {
		OpenWin(p39, winW-320, winH-240, 640, 480, 0, 0, 0)
	}
}
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	g := New([]string{"P39.BMP"})
	code := g.Generate(program)

	// After generation, vmVars should contain the variables used in mes() blocks
	if g.vmVars == nil {
		t.Fatal("vmVars not initialized")
	}

	expectedVars := []string{"winw", "winh", "p39"}
	for _, v := range expectedVars {
		if !g.vmVars[v] {
			t.Errorf("Expected variable %q in vmVars, but not found", v)
		}
	}

	// Verify the generated code contains variable declarations
	if !strings.Contains(code, "var winw int") {
		t.Error("Generated code should contain 'var winw int'")
	}
	if !strings.Contains(code, "var winh int") {
		t.Error("Generated code should contain 'var winh int'")
	}
	if !strings.Contains(code, "var p39 int") {
		t.Error("Generated code should contain 'var p39 int'")
	}
}
