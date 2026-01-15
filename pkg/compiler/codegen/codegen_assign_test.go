package codegen

import (
	"strings"
	"testing"

	"github.com/zurustar/son-et/pkg/compiler/lexer"
	"github.com/zurustar/son-et/pkg/compiler/parser"
)

// TestAssignCallGeneration tests that engine.Assign() is generated for VM variables
func TestAssignCallGeneration(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		shouldContain    []string
		shouldNotContain []string
	}{
		{
			name: "VM variable gets Assign call",
			input: `
main() {
	x = 10
	mes(TIME) {
		OpenWin(x, 0, 0, 640, 480, 0, 0, 0)
	}
}`,
			shouldContain: []string{
				"x = engine.Assign(\"x\", 10).(int)",
			},
			shouldNotContain: []string{},
		},
		{
			name: "Non-VM variable gets normal assignment",
			input: `
main() {
	x = 10
	y = 20
	OpenWin(x, 0, 0, 640, 480, 0, 0, 0)
	mes(TIME) {
		OpenWin(y, 0, 0, 640, 480, 0, 0, 0)
	}
}`,
			shouldContain: []string{
				"x = 10",                             // Normal assignment
				"y = engine.Assign(\"y\", 20).(int)", // Assign call
			},
			shouldNotContain: []string{
				"x = engine.Assign(\"x\"", // x should NOT use Assign
			},
		},
		{
			name: "Multiple VM variables",
			input: `
main() {
	winW = WinInfo(0)
	winH = WinInfo(1)
	mes(MIDI_TIME) {
		OpenWin(0, winW-320, winH-240, 640, 480, 0, 0, 0)
	}
}`,
			shouldContain: []string{
				"winw = engine.Assign(\"winW\", engine.WinInfo(0)).(int)",
				"winh = engine.Assign(\"winH\", engine.WinInfo(1)).(int)",
			},
			shouldNotContain: []string{},
		},
		{
			name: "Function returning string",
			input: `
main() {
	mystr = StrCode(65)
	mes(TIME) {
		TextWrite(mystr, 0, 0, 0)
	}
}`,
			shouldContain: []string{
				"mystr = engine.Assign(\"mystr\", engine.StrCode(65)).(string)",
			},
			shouldNotContain: []string{},
		},
		{
			name: "Mixed VM and non-VM variables",
			input: `
main() {
	a = 1
	b = 2
	c = 3
	d = 4
	
	OpenWin(a, 0, 0, 100, 100, 0, 0, 0)
	
	mes(TIME) {
		MoveCast(1, 0, b, c, 0, 0, 0, 0)
	}
	
	OpenWin(d, 0, 0, 100, 100, 0, 0, 0)
}`,
			shouldContain: []string{
				"a = 1", // Not in mes()
				"b = engine.Assign(\"b\", 2).(int)",
				"c = engine.Assign(\"c\", 3).(int)",
				"d = 4", // Not in mes()
			},
			shouldNotContain: []string{
				"a = engine.Assign",
				"d = engine.Assign",
			},
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
			code := g.Generate(program)

			// Check for expected strings
			for _, expected := range tt.shouldContain {
				if !strings.Contains(code, expected) {
					t.Errorf("Expected code to contain:\n%s\n\nGenerated code:\n%s", expected, code)
				}
			}

			// Check for unexpected strings
			for _, unexpected := range tt.shouldNotContain {
				if strings.Contains(code, unexpected) {
					t.Errorf("Expected code NOT to contain:\n%s\n\nGenerated code:\n%s", unexpected, code)
				}
			}
		})
	}
}

// TestTypeInference tests the inferType function
func TestTypeInference(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		varName  string
		expected string
	}{
		{
			name: "integer literal",
			input: `
main() {
	x = 42
	mes(TIME) { OpenWin(x, 0, 0, 100, 100, 0, 0, 0) }
}`,
			varName:  "x",
			expected: ".(int)",
		},
		{
			name: "function returning string",
			input: `
main() {
	s = StrCode(65)
	mes(TIME) { TextWrite(s, 0, 0, 0) }
}`,
			varName:  "s",
			expected: ".(string)",
		},
		{
			name: "function returning int",
			input: `
main() {
	w = WinInfo(0)
	mes(TIME) { OpenWin(0, w, 0, 100, 100, 0, 0, 0) }
}`,
			varName:  "w",
			expected: ".(int)",
		},
		{
			name: "function returning string",
			input: `
main() {
	s = SubStr("hello", 0, 2)
	mes(TIME) { TextWrite(s, 0, 0, 0) }
}`,
			varName:  "s",
			expected: ".(string)",
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
			code := g.Generate(program)

			// Check that the variable has the correct type assertion
			expectedPattern := tt.varName + " = engine.Assign"
			if !strings.Contains(code, expectedPattern) {
				t.Errorf("Expected code to contain Assign call for %s", tt.varName)
				return
			}

			// Check for the type assertion
			if !strings.Contains(code, tt.expected) {
				t.Errorf("Expected type assertion %s in generated code:\n%s", tt.expected, code)
			}
		})
	}
}

// TestCaseInsensitiveAssign tests that variable names are case-insensitive
func TestCaseInsensitiveAssign(t *testing.T) {
	input := `
main() {
	WinW = WinInfo(0)
	mes(TIME) {
		OpenWin(0, winw, 0, 100, 100, 0, 0, 0)
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

	// Variable should be stored as lowercase "winw"
	if !strings.Contains(code, "winw = engine.Assign(\"WinW\"") {
		t.Errorf("Expected lowercase variable name with original case in Assign call")
	}

	// Should be accessible in mes() block as engine.Variable("winw")
	if !strings.Contains(code, "engine.Variable(\"winw\")") {
		t.Errorf("Expected lowercase variable reference in mes() block")
	}
}
