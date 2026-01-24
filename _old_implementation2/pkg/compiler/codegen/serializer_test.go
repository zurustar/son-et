package codegen

import (
	"strings"
	"testing"

	"github.com/zurustar/son-et/pkg/compiler/interpreter"
	"github.com/zurustar/son-et/pkg/compiler/preprocessor"
)

func TestSerializeSimpleOpCode(t *testing.T) {
	opcodes := []interpreter.OpCode{
		{
			Cmd: interpreter.OpAssign,
			Args: []any{
				interpreter.Variable("x"),
				int64(5),
			},
		},
	}

	s := NewSerializer()
	result := s.SerializeProject("test", opcodes, nil, nil)

	// Check package declaration
	if !strings.Contains(result, "package main") {
		t.Error("Expected package declaration")
	}

	// Check imports
	if !strings.Contains(result, "import") {
		t.Error("Expected import statement")
	}

	// Check function
	if !strings.Contains(result, "func GetOpCodes()") {
		t.Error("Expected GetOpCodes function")
	}

	// Check OpCode
	if !strings.Contains(result, "interpreter.OpAssign") {
		t.Error("Expected OpAssign in output")
	}

	// Check variable
	if !strings.Contains(result, "interpreter.Variable(\"x\")") {
		t.Error("Expected Variable in output")
	}

	// Check value
	if !strings.Contains(result, "int64(5)") {
		t.Error("Expected int64(5) in output")
	}
}

func TestSerializeWithMetadata(t *testing.T) {
	metadata := &preprocessor.Metadata{
		Title:       "Test Game",
		Author:      "Test Author",
		Version:     "1.0",
		Description: "A test game",
		Custom:      map[string]string{"key": "value"},
	}

	opcodes := []interpreter.OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("x"), int64(1)}},
	}

	s := NewSerializer()
	result := s.SerializeProject("test", opcodes, metadata, nil)

	// Check metadata
	if !strings.Contains(result, "var metadata") {
		t.Error("Expected metadata variable")
	}
	if !strings.Contains(result, "\"title\": \"Test Game\"") {
		t.Error("Expected title in metadata")
	}
	if !strings.Contains(result, "\"author\": \"Test Author\"") {
		t.Error("Expected author in metadata")
	}
	if !strings.Contains(result, "\"version\": \"1.0\"") {
		t.Error("Expected version in metadata")
	}
	if !strings.Contains(result, "\"description\": \"A test game\"") {
		t.Error("Expected description in metadata")
	}
	if !strings.Contains(result, "\"key\": \"value\"") {
		t.Error("Expected custom metadata")
	}
}

func TestSerializeWithAssets(t *testing.T) {
	assets := []string{"test.bmp", "music.mid", "sound.wav"}

	opcodes := []interpreter.OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("x"), int64(1)}},
	}

	s := NewSerializer()
	result := s.SerializeProject("test", opcodes, nil, assets)

	// Check assets
	if !strings.Contains(result, "var assets") {
		t.Error("Expected assets variable")
	}
	if !strings.Contains(result, "\"test.bmp\"") {
		t.Error("Expected test.bmp in assets")
	}
	if !strings.Contains(result, "\"music.mid\"") {
		t.Error("Expected music.mid in assets")
	}
	if !strings.Contains(result, "\"sound.wav\"") {
		t.Error("Expected sound.wav in assets")
	}
}

func TestSerializeNestedOpCode(t *testing.T) {
	opcodes := []interpreter.OpCode{
		{
			Cmd: interpreter.OpAssign,
			Args: []any{
				interpreter.Variable("x"),
				interpreter.OpCode{
					Cmd: interpreter.OpBinaryOp,
					Args: []any{
						"+",
						int64(5),
						int64(3),
					},
				},
			},
		},
	}

	s := NewSerializer()
	result := s.SerializeProject("test", opcodes, nil, nil)

	// Check nested OpCode
	if !strings.Contains(result, "interpreter.OpBinaryOp") {
		t.Error("Expected OpBinaryOp in output")
	}
	if !strings.Contains(result, "\"+\"") {
		t.Error("Expected operator in output")
	}
}

func TestSerializeOpCodeSlice(t *testing.T) {
	opcodes := []interpreter.OpCode{
		{
			Cmd: interpreter.OpIf,
			Args: []any{
				interpreter.OpCode{
					Cmd:  interpreter.OpBinaryOp,
					Args: []any{">", interpreter.Variable("x"), int64(5)},
				},
				[]interpreter.OpCode{
					{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("y"), int64(10)}},
				},
				[]interpreter.OpCode{},
			},
		},
	}

	s := NewSerializer()
	result := s.SerializeProject("test", opcodes, nil, nil)

	// Check if statement
	if !strings.Contains(result, "interpreter.OpIf") {
		t.Error("Expected OpIf in output")
	}

	// Check nested OpCode slice
	if !strings.Contains(result, "[]interpreter.OpCode{") {
		t.Error("Expected OpCode slice in output")
	}
}

func TestSerializeStringValue(t *testing.T) {
	opcodes := []interpreter.OpCode{
		{
			Cmd: interpreter.OpCall,
			Args: []any{
				interpreter.Variable("LoadPic"),
				int64(1),
				"test.bmp",
			},
		},
	}

	s := NewSerializer()
	result := s.SerializeProject("test", opcodes, nil, nil)

	// Check string value
	if !strings.Contains(result, "\"test.bmp\"") {
		t.Error("Expected string value in output")
	}
}

func TestSerializeFloatValue(t *testing.T) {
	opcodes := []interpreter.OpCode{
		{
			Cmd: interpreter.OpAssign,
			Args: []any{
				interpreter.Variable("x"),
				float64(3.14),
			},
		},
	}

	s := NewSerializer()
	result := s.SerializeProject("test", opcodes, nil, nil)

	// Check float value
	if !strings.Contains(result, "float64(3.14") {
		t.Error("Expected float64 value in output")
	}
}

func TestSerializeArrayValue(t *testing.T) {
	opcodes := []interpreter.OpCode{
		{
			Cmd: interpreter.OpAssign,
			Args: []any{
				interpreter.Variable("arr"),
				[]any{int64(1), int64(2), int64(3)},
			},
		},
	}

	s := NewSerializer()
	result := s.SerializeProject("test", opcodes, nil, nil)

	// Check array value
	if !strings.Contains(result, "[]any{") {
		t.Error("Expected array in output")
	}
	if !strings.Contains(result, "int64(1)") {
		t.Error("Expected array element in output")
	}
}

func TestTrackAssetReferences(t *testing.T) {
	opcodes := []interpreter.OpCode{
		{
			Cmd: interpreter.OpCall,
			Args: []any{
				interpreter.OpCode{
					Cmd: interpreter.OpCall,
					Args: []any{
						interpreter.Variable("LoadPic"),
						int64(1),
						"test.bmp",
					},
				},
			},
		},
		{
			Cmd: interpreter.OpCall,
			Args: []any{
				interpreter.OpCode{
					Cmd: interpreter.OpCall,
					Args: []any{
						interpreter.Variable("PlayMIDI"),
						"music.mid",
					},
				},
			},
		},
	}

	s := NewSerializer()
	assets := s.TrackAssetReferences(opcodes)

	if len(assets) != 2 {
		t.Errorf("Expected 2 assets, got %d", len(assets))
	}

	assetMap := make(map[string]bool)
	for _, asset := range assets {
		assetMap[asset] = true
	}

	if !assetMap["test.bmp"] {
		t.Error("Expected test.bmp in assets")
	}
	if !assetMap["music.mid"] {
		t.Error("Expected music.mid in assets")
	}
}

func TestSerializeFunctionDefinition(t *testing.T) {
	body := []interpreter.OpCode{
		{
			Cmd: interpreter.OpAssign,
			Args: []any{
				interpreter.Variable("result"),
				interpreter.OpCode{
					Cmd: interpreter.OpBinaryOp,
					Args: []any{
						"+",
						interpreter.Variable("a"),
						interpreter.Variable("b"),
					},
				},
			},
		},
	}

	s := NewSerializer()
	result := s.SerializeFunctionDefinition("add", []string{"a", "b"}, body)

	// Check function signature
	if !strings.Contains(result, "func add(a any, b any)") {
		t.Error("Expected function signature")
	}

	// Check return type
	if !strings.Contains(result, "[]interpreter.OpCode") {
		t.Error("Expected return type")
	}

	// Check body
	if !strings.Contains(result, "interpreter.OpAssign") {
		t.Error("Expected OpAssign in function body")
	}
}

func TestSerializeVariableDeclaration(t *testing.T) {
	s := NewSerializer()

	tests := []struct {
		name     string
		varName  string
		value    any
		expected string
	}{
		{"string", "name", "test", "var name = \"test\""},
		{"int", "count", 42, "var count = 42"},
		{"float", "pi", 3.14, "var pi = 3.14"},
		{"bool", "flag", true, "var flag = true"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.SerializeVariableDeclaration(tt.varName, tt.value)
			if !strings.Contains(result, tt.expected) {
				t.Errorf("Expected %q in result, got %q", tt.expected, result)
			}
		})
	}
}

func TestSerializeComplexProgram(t *testing.T) {
	// Simulate a complex program with multiple statement types
	opcodes := []interpreter.OpCode{
		// Variable assignment
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("x"), int64(5)}},
		// If statement
		{
			Cmd: interpreter.OpIf,
			Args: []any{
				interpreter.OpCode{
					Cmd:  interpreter.OpBinaryOp,
					Args: []any{">", interpreter.Variable("x"), int64(0)},
				},
				[]interpreter.OpCode{
					{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("y"), int64(10)}},
				},
				[]interpreter.OpCode{},
			},
		},
		// Function call
		{
			Cmd: interpreter.OpCall,
			Args: []any{
				interpreter.OpCode{
					Cmd: interpreter.OpCall,
					Args: []any{
						interpreter.Variable("LoadPic"),
						int64(1),
						"test.bmp",
					},
				},
			},
		},
	}

	metadata := &preprocessor.Metadata{
		Title:  "Complex Test",
		Author: "Tester",
	}

	s := NewSerializer()
	result := s.SerializeProject("complex", opcodes, metadata, nil)

	// Verify all components are present
	if !strings.Contains(result, "package main") {
		t.Error("Missing package declaration")
	}
	if !strings.Contains(result, "var metadata") {
		t.Error("Missing metadata")
	}
	if !strings.Contains(result, "func GetOpCodes()") {
		t.Error("Missing GetOpCodes function")
	}
	if !strings.Contains(result, "interpreter.OpAssign") {
		t.Error("Missing OpAssign")
	}
	if !strings.Contains(result, "interpreter.OpIf") {
		t.Error("Missing OpIf")
	}
	if !strings.Contains(result, "interpreter.OpCall") {
		t.Error("Missing OpCall")
	}
}
