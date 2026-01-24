package codegen

import (
	"fmt"
	"strings"

	"github.com/zurustar/son-et/pkg/compiler/interpreter"
	"github.com/zurustar/son-et/pkg/compiler/preprocessor"
)

// EmbeddedTitle represents a single FILLY title in embedded mode.
// Each title has its own directory that gets embedded separately.
type EmbeddedTitle struct {
	Name        string                 // Title identifier (e.g., "kuma2")
	Directory   string                 // Source directory path (e.g., "samples/kuma2")
	EntryPoint  string                 // Main TFY file (e.g., "KUMA2.TFY")
	OpCodes     []interpreter.OpCode   // Compiled OpCodes
	Metadata    *preprocessor.Metadata // Title metadata from #info
	Assets      []string               // List of asset files in the directory
	EmbedFSName string                 // Name of the embed.FS variable (e.g., "kuma2FS")
}

// SerializeMultiTitle generates Go source code for multiple embedded FILLY titles.
// Each title's directory is embedded separately to avoid asset conflicts.
func (s *Serializer) SerializeMultiTitle(titles []EmbeddedTitle) string {
	var sb strings.Builder

	// Package declaration
	sb.WriteString("package main\n\n")

	// Imports
	sb.WriteString("import (\n")
	sb.WriteString("\t\"embed\"\n")
	sb.WriteString("\t\"fmt\"\n")
	sb.WriteString("\t\"os\"\n")
	sb.WriteString("\t\"github.com/zurustar/son-et/pkg/compiler/interpreter\"\n")
	sb.WriteString(")\n\n")

	// Generate embed.FS declarations for each title directory
	for _, title := range titles {
		sb.WriteString(fmt.Sprintf("//go:embed %s\n", title.Directory))
		sb.WriteString(fmt.Sprintf("var %s embed.FS\n\n", title.EmbedFSName))
	}

	// Title registry
	sb.WriteString("// TitleInfo holds metadata for an embedded FILLY title\n")
	sb.WriteString("type TitleInfo struct {\n")
	sb.WriteString("\tName        string\n")
	sb.WriteString("\tTitle       string\n")
	sb.WriteString("\tDescription string\n")
	sb.WriteString("\tDirectory   string\n")
	sb.WriteString("\tGetOpCodes  func() []interpreter.OpCode\n")
	sb.WriteString("\tGetFS       func() embed.FS\n")
	sb.WriteString("}\n\n")

	// Generate individual title functions
	for i, title := range titles {
		sb.WriteString(fmt.Sprintf("// FILLY Title %d: %s\n", i+1, title.Name))
		sb.WriteString(s.serializeTitleFunction(title))
		sb.WriteString("\n")
		sb.WriteString(s.serializeTitleFSFunction(title))
		sb.WriteString("\n")
	}

	// Generate title registry
	sb.WriteString("// GetTitles returns all embedded FILLY titles\n")
	sb.WriteString("func GetTitles() []TitleInfo {\n")
	sb.WriteString("\treturn []TitleInfo{\n")

	for _, title := range titles {
		titleName := title.Name
		description := ""
		if title.Metadata != nil {
			if title.Metadata.Title != "" {
				titleName = title.Metadata.Title
			}
			description = title.Metadata.Description
		}

		sb.WriteString("\t\t{\n")
		sb.WriteString(fmt.Sprintf("\t\t\tName:        %q,\n", title.Name))
		sb.WriteString(fmt.Sprintf("\t\t\tTitle:       %q,\n", titleName))
		sb.WriteString(fmt.Sprintf("\t\t\tDescription: %q,\n", description))
		sb.WriteString(fmt.Sprintf("\t\t\tDirectory:   %q,\n", title.Directory))
		sb.WriteString(fmt.Sprintf("\t\t\tGetOpCodes:  Get%sOpCodes,\n", sanitizeName(title.Name)))
		sb.WriteString(fmt.Sprintf("\t\t\tGetFS:       Get%sFS,\n", sanitizeName(title.Name)))
		sb.WriteString("\t\t},\n")
	}

	sb.WriteString("\t}\n")
	sb.WriteString("}\n\n")

	// Generate menu display function
	sb.WriteString(s.generateMenuFunction())

	// Generate main function for multi-title mode
	sb.WriteString(s.generateMultiTitleMain())

	return sb.String()
}

// serializeTitleFunction generates a function that returns OpCodes for a FILLY title.
func (s *Serializer) serializeTitleFunction(title EmbeddedTitle) string {
	var sb strings.Builder

	funcName := fmt.Sprintf("Get%sOpCodes", sanitizeName(title.Name))

	sb.WriteString(fmt.Sprintf("func %s() []interpreter.OpCode {\n", funcName))
	sb.WriteString("\treturn []interpreter.OpCode{\n")

	for _, op := range title.OpCodes {
		sb.WriteString(s.serializeOpCode(op, 2))
	}

	sb.WriteString("\t}\n")
	sb.WriteString("}\n")

	return sb.String()
}

// serializeTitleFSFunction generates a function that returns the embed.FS for a title.
func (s *Serializer) serializeTitleFSFunction(title EmbeddedTitle) string {
	funcName := fmt.Sprintf("Get%sFS", sanitizeName(title.Name))
	return fmt.Sprintf("func %s() embed.FS {\n\treturn %s\n}\n", funcName, title.EmbedFSName)
}

// generateMenuFunction generates the menu display function.
func (s *Serializer) generateMenuFunction() string {
	return `// DisplayMenu shows the FILLY title selection menu
func DisplayMenu(titles []TitleInfo) int {
	fmt.Println("=================================")
	fmt.Println("  FILLY Title Launcher")
	fmt.Println("=================================")
	fmt.Println()
	
	for i, title := range titles {
		fmt.Printf("%d. %s\n", i+1, title.Title)
		if title.Description != "" {
			fmt.Printf("   %s\n", title.Description)
		}
	}
	
	fmt.Println()
	fmt.Println("0. Exit")
	fmt.Println()
	fmt.Print("Select title: ")
	
	var choice int
	_, err := fmt.Scanf("%d", &choice)
	if err != nil {
		return -1
	}
	
	return choice
}

`
}

// generateMultiTitleMain generates the main function for multi-title mode.
func (s *Serializer) generateMultiTitleMain() string {
	return `// main is the entry point for multi-title mode
func main() {
	titles := GetTitles()
	
	for {
		choice := DisplayMenu(titles)
		
		if choice == 0 {
			fmt.Println("Goodbye!")
			os.Exit(0)
		}
		
		if choice < 1 || choice > len(titles) {
			fmt.Println("Invalid choice. Please try again.")
			continue
		}
		
		selectedTitle := titles[choice-1]
		fmt.Printf("\nLaunching: %s\n\n", selectedTitle.Title)
		
		// Get the title's OpCodes and embedded filesystem
		opcodes := selectedTitle.GetOpCodes()
		titleFS := selectedTitle.GetFS()
		
		// TODO: Execute title with opcodes and titleFS
		// The engine will use titleFS as the AssetLoader
		// This will be implemented in Phase 3 (Execution Layer)
		
		_ = opcodes
		_ = titleFS
		
		fmt.Println("\nTitle completed. Press Enter to return to menu...")
		fmt.Scanln()
	}
}
`
}

// SerializeSingleTitle generates Go source code for a single embedded FILLY title.
func (s *Serializer) SerializeSingleTitle(title EmbeddedTitle) string {
	var sb strings.Builder

	// Package declaration
	sb.WriteString("package main\n\n")

	// Imports
	sb.WriteString("import (\n")
	sb.WriteString("\t\"embed\"\n")
	sb.WriteString("\t\"github.com/zurustar/son-et/pkg/compiler/interpreter\"\n")
	sb.WriteString(")\n\n")

	// Embed the title directory
	sb.WriteString(fmt.Sprintf("//go:embed %s\n", title.Directory))
	sb.WriteString(fmt.Sprintf("var %s embed.FS\n\n", title.EmbedFSName))

	// Metadata
	if title.Metadata != nil {
		sb.WriteString("// Title metadata\n")
		sb.WriteString("var metadata = map[string]string{\n")
		if title.Metadata.Title != "" {
			sb.WriteString(fmt.Sprintf("\t\"title\": %q,\n", title.Metadata.Title))
		}
		if title.Metadata.Author != "" {
			sb.WriteString(fmt.Sprintf("\t\"author\": %q,\n", title.Metadata.Author))
		}
		if title.Metadata.Version != "" {
			sb.WriteString(fmt.Sprintf("\t\"version\": %q,\n", title.Metadata.Version))
		}
		if title.Metadata.Description != "" {
			sb.WriteString(fmt.Sprintf("\t\"description\": %q,\n", title.Metadata.Description))
		}
		for k, v := range title.Metadata.Custom {
			sb.WriteString(fmt.Sprintf("\t%q: %q,\n", k, v))
		}
		sb.WriteString("}\n\n")
	}

	// OpCode function
	sb.WriteString("// GetOpCodes returns the compiled OpCode sequence\n")
	sb.WriteString("func GetOpCodes() []interpreter.OpCode {\n")
	sb.WriteString("\treturn []interpreter.OpCode{\n")

	for _, op := range title.OpCodes {
		sb.WriteString(s.serializeOpCode(op, 2))
	}

	sb.WriteString("\t}\n")
	sb.WriteString("}\n\n")

	// FS function
	sb.WriteString("// GetFS returns the embedded filesystem\n")
	sb.WriteString(fmt.Sprintf("func GetFS() embed.FS {\n\treturn %s\n}\n\n", title.EmbedFSName))

	// Main function
	sb.WriteString(s.generateSingleTitleMain())

	return sb.String()
}

// generateSingleTitleMain generates the main function for single-title mode.
func (s *Serializer) generateSingleTitleMain() string {
	return `// main is the entry point for single-title mode
func main() {
	opcodes := GetOpCodes()
	titleFS := GetFS()
	
	// TODO: Execute opcodes with titleFS as AssetLoader
	// This will be implemented in Phase 3 (Execution Layer)
	
	_ = opcodes
	_ = titleFS
}
`
}

// sanitizeName converts a FILLY title name to a valid Go identifier.
func sanitizeName(name string) string {
	// Replace invalid characters with underscores
	result := ""
	for i, ch := range name {
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') {
			result += string(ch)
		} else if ch >= '0' && ch <= '9' {
			// Numbers are valid, but not at the start
			if i == 0 || len(result) == 0 {
				result += "_"
			}
			result += string(ch)
		} else if ch == '-' || ch == '_' || ch == ' ' {
			result += "_"
		}
	}

	// Ensure we have at least something
	if len(result) == 0 {
		result = "Title"
	}

	// Capitalize first letter
	first := result[0]
	if first >= 'a' && first <= 'z' {
		result = string(first-32) + result[1:]
	}

	return result
}
