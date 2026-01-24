package codegen

import (
	"fmt"
	"strings"

	"github.com/zurustar/son-et/pkg/compiler/interpreter"
	"github.com/zurustar/son-et/pkg/compiler/preprocessor"
)

// Serializer converts OpCode sequences to Go source code for embedded mode.
type Serializer struct {
	indent int
}

// NewSerializer creates a new serializer.
func NewSerializer() *Serializer {
	return &Serializer{
		indent: 0,
	}
}

// SerializeProject generates Go source code for an embedded project.
func (s *Serializer) SerializeProject(projectName string, opcodes []interpreter.OpCode, metadata *preprocessor.Metadata, assets []string) string {
	var sb strings.Builder

	// Package declaration
	sb.WriteString("package main\n\n")

	// Imports
	sb.WriteString("import (\n")
	sb.WriteString("\t\"github.com/zurustar/son-et/pkg/compiler/interpreter\"\n")
	sb.WriteString(")\n\n")

	// Metadata
	if metadata != nil {
		sb.WriteString("// Project metadata\n")
		sb.WriteString("var metadata = map[string]string{\n")
		if metadata.Title != "" {
			sb.WriteString(fmt.Sprintf("\t\"title\": %q,\n", metadata.Title))
		}
		if metadata.Author != "" {
			sb.WriteString(fmt.Sprintf("\t\"author\": %q,\n", metadata.Author))
		}
		if metadata.Version != "" {
			sb.WriteString(fmt.Sprintf("\t\"version\": %q,\n", metadata.Version))
		}
		if metadata.Description != "" {
			sb.WriteString(fmt.Sprintf("\t\"description\": %q,\n", metadata.Description))
		}
		for k, v := range metadata.Custom {
			sb.WriteString(fmt.Sprintf("\t%q: %q,\n", k, v))
		}
		sb.WriteString("}\n\n")
	}

	// Asset list
	if len(assets) > 0 {
		sb.WriteString("// Required assets\n")
		sb.WriteString("var assets = []string{\n")
		for _, asset := range assets {
			sb.WriteString(fmt.Sprintf("\t%q,\n", asset))
		}
		sb.WriteString("}\n\n")
	}

	// OpCode data
	sb.WriteString("// Compiled OpCode sequence\n")
	sb.WriteString("func GetOpCodes() []interpreter.OpCode {\n")
	sb.WriteString("\treturn []interpreter.OpCode{\n")

	for _, op := range opcodes {
		sb.WriteString(s.serializeOpCode(op, 2))
	}

	sb.WriteString("\t}\n")
	sb.WriteString("}\n")

	return sb.String()
}

// serializeOpCode converts a single OpCode to Go source code.
func (s *Serializer) serializeOpCode(op interpreter.OpCode, indent int) string {
	var sb strings.Builder

	indentStr := strings.Repeat("\t", indent)

	sb.WriteString(indentStr)
	sb.WriteString("{\n")

	// Command
	sb.WriteString(indentStr)
	sb.WriteString(fmt.Sprintf("\tCmd: interpreter.%s,\n", op.Cmd.String()))

	// Arguments
	if len(op.Args) > 0 {
		sb.WriteString(indentStr)
		sb.WriteString("\tArgs: []any{\n")

		for _, arg := range op.Args {
			sb.WriteString(s.serializeValue(arg, indent+2))
		}

		sb.WriteString(indentStr)
		sb.WriteString("\t},\n")
	}

	sb.WriteString(indentStr)
	sb.WriteString("},\n")

	return sb.String()
}

// serializeValue converts a value to Go source code.
func (s *Serializer) serializeValue(value any, indent int) string {
	indentStr := strings.Repeat("\t", indent)

	switch v := value.(type) {
	case interpreter.Variable:
		return indentStr + fmt.Sprintf("interpreter.Variable(%q),\n", string(v))

	case string:
		return indentStr + fmt.Sprintf("%q,\n", v)

	case int:
		return indentStr + fmt.Sprintf("int64(%d),\n", v)

	case int64:
		return indentStr + fmt.Sprintf("int64(%d),\n", v)

	case float64:
		return indentStr + fmt.Sprintf("float64(%f),\n", v)

	case bool:
		return indentStr + fmt.Sprintf("%t,\n", v)

	case interpreter.OpCode:
		return s.serializeOpCode(v, indent)

	case []interpreter.OpCode:
		var sb strings.Builder
		sb.WriteString(indentStr)
		sb.WriteString("[]interpreter.OpCode{\n")
		for _, op := range v {
			sb.WriteString(s.serializeOpCode(op, indent+1))
		}
		sb.WriteString(indentStr)
		sb.WriteString("},\n")
		return sb.String()

	case []any:
		var sb strings.Builder
		sb.WriteString(indentStr)
		sb.WriteString("[]any{\n")
		for _, item := range v {
			sb.WriteString(s.serializeValue(item, indent+1))
		}
		sb.WriteString(indentStr)
		sb.WriteString("},\n")
		return sb.String()

	default:
		return indentStr + fmt.Sprintf("nil, // unsupported type: %T\n", v)
	}
}

// SerializeFunctionDefinition generates Go code for a function definition.
func (s *Serializer) SerializeFunctionDefinition(name string, params []string, body []interpreter.OpCode) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("func %s(", name))
	for i, param := range params {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("%s any", param))
	}
	sb.WriteString(") []interpreter.OpCode {\n")
	sb.WriteString("\treturn []interpreter.OpCode{\n")

	for _, op := range body {
		sb.WriteString(s.serializeOpCode(op, 2))
	}

	sb.WriteString("\t}\n")
	sb.WriteString("}\n")

	return sb.String()
}

// SerializeVariableDeclaration generates Go code for a variable declaration.
func (s *Serializer) SerializeVariableDeclaration(name string, value any) string {
	return fmt.Sprintf("var %s = %v\n", name, s.formatValue(value))
}

// formatValue formats a value for Go source code.
func (s *Serializer) formatValue(value any) string {
	switch v := value.(type) {
	case string:
		return fmt.Sprintf("%q", v)
	case int, int64:
		return fmt.Sprintf("%d", v)
	case float64:
		return fmt.Sprintf("%f", v)
	case bool:
		return fmt.Sprintf("%t", v)
	default:
		return "nil"
	}
}

// TrackAssetReferences extracts asset file references from OpCode sequences.
func (s *Serializer) TrackAssetReferences(opcodes []interpreter.OpCode) []string {
	assets := make(map[string]bool)
	s.extractAssets(opcodes, assets)

	result := make([]string, 0, len(assets))
	for asset := range assets {
		result = append(result, asset)
	}
	return result
}

// extractAssets recursively extracts asset references from OpCodes.
func (s *Serializer) extractAssets(opcodes []interpreter.OpCode, assets map[string]bool) {
	for _, op := range opcodes {
		// Check if this is a LoadPic, PlayMIDI, PlayWAVE, etc.
		if op.Cmd == interpreter.OpCall && len(op.Args) > 0 {
			// Check if first arg is a Variable (function name)
			if funcName, ok := op.Args[0].(interpreter.Variable); ok {
				funcStr := string(funcName)
				// Check for asset-loading functions
				if funcStr == "LoadPic" || funcStr == "PlayMIDI" || funcStr == "PlayWAVE" {
					// Look for filename in remaining args
					for i := 1; i < len(op.Args); i++ {
						if filename, ok := op.Args[i].(string); ok {
							assets[filename] = true
						}
					}
				}
			}
		}

		// Recursively check nested OpCodes in all arguments
		for _, arg := range op.Args {
			switch v := arg.(type) {
			case interpreter.OpCode:
				s.extractAssets([]interpreter.OpCode{v}, assets)
			case []interpreter.OpCode:
				s.extractAssets(v, assets)
			case []any:
				for _, item := range v {
					if nestedOp, ok := item.(interpreter.OpCode); ok {
						s.extractAssets([]interpreter.OpCode{nestedOp}, assets)
					} else if nestedOps, ok := item.([]interpreter.OpCode); ok {
						s.extractAssets(nestedOps, assets)
					}
				}
			}
		}
	}
}
