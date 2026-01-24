package engine

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
)

func TestNewParseError(t *testing.T) {
	err := NewParseError(10, 5, "unexpected token '%s'", "}")

	if err.Type != ErrorTypeParser {
		t.Errorf("Expected ErrorTypeParser, got %v", err.Type)
	}
	if err.Line != 10 {
		t.Errorf("Expected line 10, got %d", err.Line)
	}
	if err.Column != 5 {
		t.Errorf("Expected column 5, got %d", err.Column)
	}

	expected := "Parse error at line 10, column 5: unexpected token '}'"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}
}

func TestNewRuntimeError(t *testing.T) {
	err := NewRuntimeError("LoadPic", "1, \"test.bmp\"", "file not found")

	if err.Type != ErrorTypeRuntime {
		t.Errorf("Expected ErrorTypeRuntime, got %v", err.Type)
	}
	if err.OpCode != "LoadPic" {
		t.Errorf("Expected OpCode 'LoadPic', got '%s'", err.OpCode)
	}

	expected := "Runtime error in LoadPic(1, \"test.bmp\"): file not found"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}
}

func TestNewRuntimeError_NoArgs(t *testing.T) {
	err := NewRuntimeError("Wait", "", "invalid wait count")

	expected := "Runtime error in Wait: invalid wait count"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}
}

func TestNewAssetError(t *testing.T) {
	cause := errors.New("file not found")
	err := NewAssetError("test.bmp", "failed to load image", cause)

	if err.Type != ErrorTypeAsset {
		t.Errorf("Expected ErrorTypeAsset, got %v", err.Type)
	}
	if err.Filename != "test.bmp" {
		t.Errorf("Expected filename 'test.bmp', got '%s'", err.Filename)
	}
	if err.Cause != cause {
		t.Error("Expected cause to be set")
	}

	expected := "Asset error loading 'test.bmp': failed to load image (cause: file not found)"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}
}

func TestNewAssetError_NoCause(t *testing.T) {
	err := NewAssetError("test.bmp", "invalid format", nil)

	expected := "Asset error loading 'test.bmp': invalid format"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}
}

func TestEngineError_Unwrap(t *testing.T) {
	cause := errors.New("underlying error")
	err := NewAssetError("test.bmp", "failed", cause)

	unwrapped := err.Unwrap()
	if unwrapped != cause {
		t.Error("Unwrap should return the cause")
	}
}

func TestEngine_ReportError(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	engine := NewEngine(&MockRenderer{}, &MockAssetLoader{Files: make(map[string][]byte)}, &MockImageDecoder{})

	err := NewParseError(5, 10, "test error")
	engine.ReportError(err)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "Parse error at line 5, column 10: test error") {
		t.Errorf("Expected parse error in output, got: %s", output)
	}
}

func TestEngine_ReportParseError(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	engine := NewEngine(&MockRenderer{}, &MockAssetLoader{Files: make(map[string][]byte)}, &MockImageDecoder{})
	engine.ReportParseError(15, 20, "unexpected token '%s'", "}")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "Parse error at line 15, column 20") {
		t.Errorf("Expected parse error in output, got: %s", output)
	}
	if !strings.Contains(output, "unexpected token '}'") {
		t.Errorf("Expected error message in output, got: %s", output)
	}
}

func TestEngine_ReportRuntimeError(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	engine := NewEngine(&MockRenderer{}, &MockAssetLoader{Files: make(map[string][]byte)}, &MockImageDecoder{})
	engine.ReportRuntimeError("LoadPic", "1, \"test.bmp\"", "file not found")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "Runtime error in LoadPic") {
		t.Errorf("Expected runtime error in output, got: %s", output)
	}
	if !strings.Contains(output, "file not found") {
		t.Errorf("Expected error message in output, got: %s", output)
	}
}

func TestEngine_ReportAssetError(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	engine := NewEngine(&MockRenderer{}, &MockAssetLoader{Files: make(map[string][]byte)}, &MockImageDecoder{})
	cause := errors.New("io error")
	engine.ReportAssetError("test.bmp", "failed to load", cause)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "Asset error loading 'test.bmp'") {
		t.Errorf("Expected asset error in output, got: %s", output)
	}
	if !strings.Contains(output, "failed to load") {
		t.Errorf("Expected error message in output, got: %s", output)
	}
}

func TestEngine_ReportError_GenericError(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	engine := NewEngine(&MockRenderer{}, &MockAssetLoader{Files: make(map[string][]byte)}, &MockImageDecoder{})
	err := errors.New("generic error")
	engine.ReportError(err)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "generic error") {
		t.Errorf("Expected generic error in output, got: %s", output)
	}
}

func TestErrorFormatting_WithArgs(t *testing.T) {
	err := NewRuntimeError("TestOp", "arg1, arg2", "test message with %d args", 2)

	expected := "Runtime error in TestOp(arg1, arg2): test message with 2 args"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}
}

func TestParseError_MultipleFormattingArgs(t *testing.T) {
	err := NewParseError(1, 1, "expected %s but got %s", "identifier", "number")

	if !strings.Contains(err.Error(), "expected identifier but got number") {
		t.Errorf("Expected formatted message, got: %s", err.Error())
	}
}
