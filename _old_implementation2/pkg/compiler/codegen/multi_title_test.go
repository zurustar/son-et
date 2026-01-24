package codegen

import (
	"strings"
	"testing"

	"github.com/zurustar/son-et/pkg/compiler/interpreter"
	"github.com/zurustar/son-et/pkg/compiler/preprocessor"
)

func TestSerializeMultiTitle(t *testing.T) {
	titles := []EmbeddedTitle{
		{
			Name:        "kuma2",
			Directory:   "samples/kuma2",
			EntryPoint:  "KUMA2.TFY",
			EmbedFSName: "kuma2FS",
			OpCodes: []interpreter.OpCode{
				{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("x"), int64(1)}},
			},
			Metadata: &preprocessor.Metadata{
				Title:       "Kuma Game",
				Description: "A bear adventure",
			},
		},
		{
			Name:        "robot",
			Directory:   "samples/robot",
			EntryPoint:  "ROBOT.TFY",
			EmbedFSName: "robotFS",
			OpCodes: []interpreter.OpCode{
				{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("y"), int64(2)}},
			},
			Metadata: &preprocessor.Metadata{
				Title:       "Robot Game",
				Description: "A robot story",
			},
		},
	}

	s := NewSerializer()
	result := s.SerializeMultiTitle(titles)

	// Check package declaration
	if !strings.Contains(result, "package main") {
		t.Error("Expected package declaration")
	}

	// Check imports
	if !strings.Contains(result, "\"embed\"") {
		t.Error("Expected embed import")
	}

	// Check embed directives for each title
	if !strings.Contains(result, "//go:embed samples/kuma2") {
		t.Error("Expected embed directive for kuma2")
	}
	if !strings.Contains(result, "var kuma2FS embed.FS") {
		t.Error("Expected kuma2FS variable")
	}

	if !strings.Contains(result, "//go:embed samples/robot") {
		t.Error("Expected embed directive for robot")
	}
	if !strings.Contains(result, "var robotFS embed.FS") {
		t.Error("Expected robotFS variable")
	}

	// Check TitleInfo struct
	if !strings.Contains(result, "type TitleInfo struct") {
		t.Error("Expected TitleInfo struct")
	}
	if !strings.Contains(result, "GetOpCodes  func() []interpreter.OpCode") {
		t.Error("Expected GetOpCodes field in TitleInfo")
	}
	if !strings.Contains(result, "GetFS       func() embed.FS") {
		t.Error("Expected GetFS field in TitleInfo")
	}

	// Check title functions
	if !strings.Contains(result, "func GetKuma2OpCodes()") {
		t.Error("Expected GetKuma2OpCodes function")
	}
	if !strings.Contains(result, "func GetKuma2FS()") {
		t.Error("Expected GetKuma2FS function")
	}
	if !strings.Contains(result, "func GetRobotOpCodes()") {
		t.Error("Expected GetRobotOpCodes function")
	}
	if !strings.Contains(result, "func GetRobotFS()") {
		t.Error("Expected GetRobotFS function")
	}

	// Check GetTitles function
	if !strings.Contains(result, "func GetTitles() []TitleInfo") {
		t.Error("Expected GetTitles function")
	}

	// Check title registry entries
	if !strings.Contains(result, "Name:        \"kuma2\"") {
		t.Error("Expected kuma2 in title registry")
	}
	if !strings.Contains(result, "Title:       \"Kuma Game\"") {
		t.Error("Expected Kuma Game title")
	}
	if !strings.Contains(result, "Description: \"A bear adventure\"") {
		t.Error("Expected kuma2 description")
	}

	if !strings.Contains(result, "Name:        \"robot\"") {
		t.Error("Expected robot in title registry")
	}
	if !strings.Contains(result, "Title:       \"Robot Game\"") {
		t.Error("Expected Robot Game title")
	}

	// Check menu function
	if !strings.Contains(result, "func DisplayMenu(titles []TitleInfo) int") {
		t.Error("Expected DisplayMenu function")
	}

	// Check main function
	if !strings.Contains(result, "func main()") {
		t.Error("Expected main function")
	}
	if !strings.Contains(result, "titles := GetTitles()") {
		t.Error("Expected GetTitles call in main")
	}
}

func TestSerializeSingleTitle(t *testing.T) {
	title := EmbeddedTitle{
		Name:        "kuma2",
		Directory:   "samples/kuma2",
		EntryPoint:  "KUMA2.TFY",
		EmbedFSName: "kuma2FS",
		OpCodes: []interpreter.OpCode{
			{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("x"), int64(5)}},
		},
		Metadata: &preprocessor.Metadata{
			Title:       "Kuma Game",
			Author:      "Test Author",
			Version:     "1.0",
			Description: "A bear adventure",
		},
	}

	s := NewSerializer()
	result := s.SerializeSingleTitle(title)

	// Check package declaration
	if !strings.Contains(result, "package main") {
		t.Error("Expected package declaration")
	}

	// Check imports
	if !strings.Contains(result, "\"embed\"") {
		t.Error("Expected embed import")
	}

	// Check embed directive
	if !strings.Contains(result, "//go:embed samples/kuma2") {
		t.Error("Expected embed directive")
	}
	if !strings.Contains(result, "var kuma2FS embed.FS") {
		t.Error("Expected kuma2FS variable")
	}

	// Check metadata
	if !strings.Contains(result, "var metadata = map[string]string{") {
		t.Error("Expected metadata variable")
	}
	if !strings.Contains(result, "\"title\": \"Kuma Game\"") {
		t.Error("Expected title in metadata")
	}
	if !strings.Contains(result, "\"author\": \"Test Author\"") {
		t.Error("Expected author in metadata")
	}

	// Check GetOpCodes function
	if !strings.Contains(result, "func GetOpCodes() []interpreter.OpCode") {
		t.Error("Expected GetOpCodes function")
	}

	// Check GetFS function
	if !strings.Contains(result, "func GetFS() embed.FS") {
		t.Error("Expected GetFS function")
	}
	if !strings.Contains(result, "return kuma2FS") {
		t.Error("Expected return kuma2FS in GetFS")
	}

	// Check main function
	if !strings.Contains(result, "func main()") {
		t.Error("Expected main function")
	}
	if !strings.Contains(result, "opcodes := GetOpCodes()") {
		t.Error("Expected GetOpCodes call in main")
	}
	if !strings.Contains(result, "titleFS := GetFS()") {
		t.Error("Expected GetFS call in main")
	}
}

func TestSanitizeName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"kuma2", "Kuma2"},
		{"robot", "Robot"},
		{"y-saru", "Y_saru"},
		{"test_game", "Test_game"},
		{"my game", "My_game"},
		{"123test", "_123test"}, // Numbers can't start identifier, so prepend underscore
		{"UPPERCASE", "UPPERCASE"},
		{"mixed-Case_Name", "Mixed_Case_Name"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := sanitizeName(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeName(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSerializeMultiTitleWithoutMetadata(t *testing.T) {
	titles := []EmbeddedTitle{
		{
			Name:        "simple",
			Directory:   "samples/simple",
			EntryPoint:  "SIMPLE.TFY",
			EmbedFSName: "simpleFS",
			OpCodes: []interpreter.OpCode{
				{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("x"), int64(1)}},
			},
			Metadata: nil, // No metadata
		},
	}

	s := NewSerializer()
	result := s.SerializeMultiTitle(titles)

	// Should still generate valid code
	if !strings.Contains(result, "package main") {
		t.Error("Expected package declaration")
	}

	// Title name should be used when metadata is missing
	if !strings.Contains(result, "Name:        \"simple\"") {
		t.Error("Expected simple in title registry")
	}
	if !strings.Contains(result, "Title:       \"simple\"") {
		t.Error("Expected simple as title when metadata missing")
	}
}

func TestSerializeTitleFunction(t *testing.T) {
	title := EmbeddedTitle{
		Name:        "test",
		Directory:   "samples/test",
		EntryPoint:  "TEST.TFY",
		EmbedFSName: "testFS",
		OpCodes: []interpreter.OpCode{
			{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("x"), int64(10)}},
			{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("y"), int64(20)}},
		},
	}

	s := NewSerializer()
	result := s.serializeTitleFunction(title)

	// Check function name
	if !strings.Contains(result, "func GetTestOpCodes()") {
		t.Error("Expected GetTestOpCodes function")
	}

	// Check return type
	if !strings.Contains(result, "[]interpreter.OpCode") {
		t.Error("Expected OpCode slice return type")
	}

	// Check OpCodes
	if !strings.Contains(result, "interpreter.OpAssign") {
		t.Error("Expected OpAssign in function")
	}
	if !strings.Contains(result, "int64(10)") {
		t.Error("Expected first value in function")
	}
	if !strings.Contains(result, "int64(20)") {
		t.Error("Expected second value in function")
	}
}

func TestSerializeTitleFSFunction(t *testing.T) {
	title := EmbeddedTitle{
		Name:        "test",
		Directory:   "samples/test",
		EntryPoint:  "TEST.TFY",
		EmbedFSName: "testFS",
	}

	s := NewSerializer()
	result := s.serializeTitleFSFunction(title)

	// Check function name
	if !strings.Contains(result, "func GetTestFS()") {
		t.Error("Expected GetTestFS function")
	}

	// Check return type
	if !strings.Contains(result, "embed.FS") {
		t.Error("Expected embed.FS return type")
	}

	// Check return statement
	if !strings.Contains(result, "return testFS") {
		t.Error("Expected return testFS")
	}
}

func TestGenerateMenuFunction(t *testing.T) {
	s := NewSerializer()
	result := s.generateMenuFunction()

	// Check function signature
	if !strings.Contains(result, "func DisplayMenu(titles []TitleInfo) int") {
		t.Error("Expected DisplayMenu function signature")
	}

	// Check menu elements
	if !strings.Contains(result, "FILLY Title Launcher") {
		t.Error("Expected menu title")
	}
	if !strings.Contains(result, "Select title:") {
		t.Error("Expected selection prompt")
	}
	if !strings.Contains(result, "0. Exit") {
		t.Error("Expected exit option")
	}
}

func TestGenerateMultiTitleMain(t *testing.T) {
	s := NewSerializer()
	result := s.generateMultiTitleMain()

	// Check function signature
	if !strings.Contains(result, "func main()") {
		t.Error("Expected main function")
	}

	// Check main logic
	if !strings.Contains(result, "titles := GetTitles()") {
		t.Error("Expected GetTitles call")
	}
	if !strings.Contains(result, "choice := DisplayMenu(titles)") {
		t.Error("Expected DisplayMenu call")
	}
	if !strings.Contains(result, "opcodes := selectedTitle.GetOpCodes()") {
		t.Error("Expected GetOpCodes call")
	}
	if !strings.Contains(result, "titleFS := selectedTitle.GetFS()") {
		t.Error("Expected GetFS call")
	}

	// Check exit handling
	if !strings.Contains(result, "if choice == 0") {
		t.Error("Expected exit condition")
	}
	if !strings.Contains(result, "os.Exit(0)") {
		t.Error("Expected os.Exit call")
	}
}

func TestGenerateSingleTitleMain(t *testing.T) {
	s := NewSerializer()
	result := s.generateSingleTitleMain()

	// Check function signature
	if !strings.Contains(result, "func main()") {
		t.Error("Expected main function")
	}

	// Check main logic
	if !strings.Contains(result, "opcodes := GetOpCodes()") {
		t.Error("Expected GetOpCodes call")
	}
	if !strings.Contains(result, "titleFS := GetFS()") {
		t.Error("Expected GetFS call")
	}

	// Check TODO comment
	if !strings.Contains(result, "TODO: Execute opcodes") {
		t.Error("Expected TODO comment")
	}
}
