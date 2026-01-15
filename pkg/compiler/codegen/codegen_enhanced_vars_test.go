package codegen

import (
	"strings"
	"testing"

	"github.com/zurustar/son-et/pkg/compiler/ast"
	"github.com/zurustar/son-et/pkg/compiler/lexer"
	"github.com/zurustar/son-et/pkg/compiler/parser"
)

// TestEnhancedVariableDetection tests that all variable references are detected
func TestEnhancedVariableDetection(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name: "variables in infix expressions",
			input: `
main() {
	a = 100
	b = 200
	mes(TIME) {
		OpenWin(0, a-50, b+100, 640, 480, 0, 0, 0)
	}
}`,
			expected: []string{"a", "b"},
		},
		{
			name: "variables in nested function arguments",
			input: `
main() {
	x = 10
	y = 20
	mes(TIME) {
		MoveCast(1, 0, WinInfo(x), WinInfo(y), 0, 0, 0, 0)
	}
}`,
			expected: []string{"x", "y"},
		},
		{
			name: "variables in array subscripts",
			input: `
main() {
	i = 5
	arr[0] = 100
	mes(TIME) {
		OpenWin(arr[i], 0, 0, 640, 480, 0, 0, 0)
	}
}`,
			expected: []string{"arr", "i"},
		},
		{
			name: "variables in complex nested expressions",
			input: `
main() {
	w = 1280
	h = 720
	offset = 50
	mes(TIME) {
		OpenWin(0, (w-640)/2+offset, (h-480)/2-offset, 640, 480, 0, 0, 0)
	}
}`,
			expected: []string{"h", "offset", "w"},
		},
		{
			name: "variables in if conditions inside mes",
			input: `
main() {
	flag = 1
	x = 100
	mes(TIME) {
		if (flag == 1) {
			MoveCast(1, 0, x, 0, 0, 0, 0, 0)
		}
	}
}`,
			expected: []string{"flag", "x"},
		},
		{
			name: "variables in for loops inside mes",
			input: `
main() {
	start = 0
	end = 10
	mes(TIME) {
		for (i = start; i < end; i = i + 1) {
			MoveCast(i, 0, 100, 100, 0, 0, 0, 0)
		}
	}
}`,
			expected: []string{"end", "i", "start"},
		},
		{
			name: "variables in switch statements inside mes",
			input: `
main() {
	mode = 1
	x = 100
	y = 200
	mes(TIME) {
		switch (mode) {
			case 1:
				MoveCast(1, 0, x, 0, 0, 0, 0, 0)
			case 2:
				MoveCast(1, 0, y, 0, 0, 0, 0, 0)
		}
	}
}`,
			expected: []string{"mode", "x", "y"},
		},
		{
			name: "exclude function names",
			input: `
main() {
	x = 10
	mes(TIME) {
		OpenWin(x, 0, 0, 640, 480, 0, 0, 0)
		MoveCast(1, 0, 100, 100, 0, 0, 0, 0)
	}
}`,
			expected: []string{"x"}, // OpenWin and MoveCast should NOT be included
		},
		{
			name: "exclude constants",
			input: `
main() {
	x = 10
	mes(TIME) {
		OpenWin(x, 0, 0, 640, 480, 0, 0, 0)
	}
	mes(MIDI_TIME) {
		MoveCast(1, 0, x, 0, 0, 0, 0, 0)
	}
}`,
			expected: []string{"x"}, // TIME and MIDI_TIME should NOT be included
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

			// Check each expected variable is present
			for _, exp := range tt.expected {
				if !vmVars[exp] {
					t.Errorf("Expected variable %q not found in vmVars. Found: %v", exp, vmVars)
				}
			}

			// Check no unexpected variables
			if len(vmVars) != len(tt.expected) {
				var found []string
				for v := range vmVars {
					found = append(found, v)
				}
				t.Errorf("Expected %d variables, got %d. Expected: %v, Found: %v",
					len(tt.expected), len(vmVars), tt.expected, found)
			}
		})
	}
}

// TestEnhancedVariableDetectionInGeneratedCode tests the full integration
func TestEnhancedVariableDetectionInGeneratedCode(t *testing.T) {
	input := `
main() {
	winW = WinInfo(0)
	winH = WinInfo(1)
	offsetX = 50
	offsetY = 100
	
	mes(MIDI_TIME) {
		OpenWin(0, (winW-640)/2+offsetX, (winH-480)/2-offsetY, 640, 480, 0, 0, 0)
	}
}
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	g := New([]string{})
	code := g.Generate(program)

	// All four variables should use Assign()
	expectedVars := []string{"winw", "winh", "offsetx", "offsety"}
	for _, v := range expectedVars {
		assignPattern := v + " = engine.Assign("
		if !strings.Contains(code, assignPattern) {
			t.Errorf("Expected variable %q to use engine.Assign(), but not found in:\n%s", v, code)
		}
	}

	// Verify they're passed to RegisterSequence
	if !strings.Contains(code, "map[string]any{") {
		t.Error("Expected RegisterSequence to have variable map")
	}
}
