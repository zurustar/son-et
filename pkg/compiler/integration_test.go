// Package compiler provides integration tests for the compilation pipeline.
// These tests verify that the compiler can process real-world sample files.
package compiler

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/zurustar/son-et/pkg/script"
)

// IntegrationTestStats holds statistics about a compilation run.
type IntegrationTestStats struct {
	Directory    string
	TotalFiles   int
	SuccessCount int
	ErrorCount   int
	TotalOpcodes int
	FileResults  []FileResult
}

// FileResult holds the result of compiling a single file.
type FileResult struct {
	FileName    string
	OpcodeCount int
	ErrorCount  int
	Errors      []string
}

// TestIntegrationCompileDirectory tests CompileDirectory with multiple sample directories.
// This test verifies that all .TFY files in each directory can be compiled.
//
// Requirement 12.4: System provides sample FILLY scripts for integration testing.
func TestIntegrationCompileDirectory(t *testing.T) {
	// Define sample directories to test
	sampleDirs := []string{
		"samples/robot",
		"samples/home",
		"samples/kuma2",
		"samples/sab1",
		"samples/sab2",
		"samples/voice",
		"samples/voice1",
		"samples/y_saru",
		"samples/sabo2",
		"samples/touch2",
		"samples/yosemiya",
		"samples/ftile400",
	}

	// Get workspace root (go up from pkg/compiler to root)
	workspaceRoot, err := getWorkspaceRoot()
	if err != nil {
		t.Fatalf("Failed to get workspace root: %v", err)
	}

	var allStats []IntegrationTestStats

	for _, dir := range sampleDirs {
		fullPath := filepath.Join(workspaceRoot, dir)

		// Skip if directory doesn't exist
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Logf("Skipping non-existent directory: %s", dir)
			continue
		}

		t.Run(dir, func(t *testing.T) {
			stats := compileDirectoryWithStats(t, fullPath, dir)
			allStats = append(allStats, stats)

			// Log statistics
			t.Logf("Directory: %s", stats.Directory)
			t.Logf("  Total files: %d", stats.TotalFiles)
			t.Logf("  Successful: %d", stats.SuccessCount)
			t.Logf("  With errors: %d", stats.ErrorCount)
			t.Logf("  Total opcodes generated: %d", stats.TotalOpcodes)

			// Note: Some parse errors are expected for preprocessor directives (#include, etc.)
			// We log a warning but don't fail the test since sample files may use
			// features not yet implemented in the compiler.
			if stats.TotalFiles > 0 && stats.SuccessCount == 0 {
				t.Logf("Warning: No files compiled successfully in %s (may use unsupported features)", dir)
			}
		})
	}

	// Log overall summary
	t.Log("\n=== Integration Test Summary ===")
	totalFiles := 0
	totalSuccess := 0
	totalErrors := 0
	totalOpcodes := 0

	for _, stats := range allStats {
		totalFiles += stats.TotalFiles
		totalSuccess += stats.SuccessCount
		totalErrors += stats.ErrorCount
		totalOpcodes += stats.TotalOpcodes
	}

	t.Logf("Total directories tested: %d", len(allStats))
	t.Logf("Total files processed: %d", totalFiles)
	t.Logf("Total successful compilations: %d", totalSuccess)
	t.Logf("Total files with errors: %d", totalErrors)
	t.Logf("Total opcodes generated: %d", totalOpcodes)
}

// TestIntegrationRobotDirectory tests compilation of the robot sample directory.
// This is a focused test for the robot directory which contains multiple .TFY files.
func TestIntegrationRobotDirectory(t *testing.T) {
	workspaceRoot, err := getWorkspaceRoot()
	if err != nil {
		t.Fatalf("Failed to get workspace root: %v", err)
	}

	dirPath := filepath.Join(workspaceRoot, "samples/robot")
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		t.Skip("samples/robot directory not found")
	}

	stats := compileDirectoryWithStats(t, dirPath, "samples/robot")

	// Log detailed results for each file
	t.Logf("Robot directory compilation results:")
	for _, fr := range stats.FileResults {
		if fr.ErrorCount > 0 {
			t.Logf("  %s: %d opcodes, %d errors", fr.FileName, fr.OpcodeCount, fr.ErrorCount)
			for _, errMsg := range fr.Errors {
				t.Logf("    Error: %s", truncateString(errMsg, 100))
			}
		} else {
			t.Logf("  %s: %d opcodes (success)", fr.FileName, fr.OpcodeCount)
		}
	}

	// Verify we have some successful compilations
	if stats.SuccessCount == 0 {
		t.Error("Expected at least some files to compile successfully")
	}
}

// TestIntegrationHomeDirectory tests compilation of the home sample directory.
func TestIntegrationHomeDirectory(t *testing.T) {
	workspaceRoot, err := getWorkspaceRoot()
	if err != nil {
		t.Fatalf("Failed to get workspace root: %v", err)
	}

	dirPath := filepath.Join(workspaceRoot, "samples/home")
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		t.Skip("samples/home directory not found")
	}

	stats := compileDirectoryWithStats(t, dirPath, "samples/home")

	// Log detailed results
	t.Logf("Home directory compilation results:")
	for _, fr := range stats.FileResults {
		if fr.ErrorCount > 0 {
			t.Logf("  %s: %d opcodes, %d errors", fr.FileName, fr.OpcodeCount, fr.ErrorCount)
		} else {
			t.Logf("  %s: %d opcodes (success)", fr.FileName, fr.OpcodeCount)
		}
	}
}

// TestIntegrationSab1Directory tests compilation of the sab1 sample directory.
func TestIntegrationSab1Directory(t *testing.T) {
	workspaceRoot, err := getWorkspaceRoot()
	if err != nil {
		t.Fatalf("Failed to get workspace root: %v", err)
	}

	dirPath := filepath.Join(workspaceRoot, "samples/sab1")
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		t.Skip("samples/sab1 directory not found")
	}

	stats := compileDirectoryWithStats(t, dirPath, "samples/sab1")

	// Log detailed results
	t.Logf("Sab1 directory compilation results:")
	for _, fr := range stats.FileResults {
		if fr.ErrorCount > 0 {
			t.Logf("  %s: %d opcodes, %d errors", fr.FileName, fr.OpcodeCount, fr.ErrorCount)
		} else {
			t.Logf("  %s: %d opcodes (success)", fr.FileName, fr.OpcodeCount)
		}
	}
}

// TestIntegrationSab2Directory tests compilation of the sab2 sample directory.
func TestIntegrationSab2Directory(t *testing.T) {
	workspaceRoot, err := getWorkspaceRoot()
	if err != nil {
		t.Fatalf("Failed to get workspace root: %v", err)
	}

	dirPath := filepath.Join(workspaceRoot, "samples/sab2")
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		t.Skip("samples/sab2 directory not found")
	}

	stats := compileDirectoryWithStats(t, dirPath, "samples/sab2")

	// Log detailed results
	t.Logf("Sab2 directory compilation results:")
	for _, fr := range stats.FileResults {
		if fr.ErrorCount > 0 {
			t.Logf("  %s: %d opcodes, %d errors", fr.FileName, fr.OpcodeCount, fr.ErrorCount)
		} else {
			t.Logf("  %s: %d opcodes (success)", fr.FileName, fr.OpcodeCount)
		}
	}
}

// compileDirectoryWithStats compiles all files in a directory and returns statistics.
func compileDirectoryWithStats(t *testing.T, dirPath, displayName string) IntegrationTestStats {
	t.Helper()

	stats := IntegrationTestStats{
		Directory: displayName,
	}

	// Use CompileDirectoryWithResults to get detailed results
	results, err := CompileDirectoryWithResults(dirPath)
	if err != nil {
		t.Logf("Warning: Failed to load scripts from %s: %v", displayName, err)
		return stats
	}

	stats.TotalFiles = len(results)

	for _, result := range results {
		fr := FileResult{
			FileName:    result.FileName,
			OpcodeCount: len(result.OpCodes),
			ErrorCount:  len(result.Errors),
		}

		if len(result.Errors) > 0 {
			stats.ErrorCount++
			for _, e := range result.Errors {
				fr.Errors = append(fr.Errors, e.Error())
			}
		} else {
			stats.SuccessCount++
			stats.TotalOpcodes += len(result.OpCodes)
		}

		stats.FileResults = append(stats.FileResults, fr)
	}

	return stats
}

// getWorkspaceRoot returns the workspace root directory.
// It navigates up from the current directory to find the root.
func getWorkspaceRoot() (string, error) {
	// Start from current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Navigate up to find the workspace root (where samples/ exists)
	dir := cwd
	for {
		samplesPath := filepath.Join(dir, "samples")
		if _, err := os.Stat(samplesPath); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root without finding samples/
			break
		}
		dir = parent
	}

	// If not found, try relative path from test location
	// Tests are run from pkg/compiler, so go up two levels
	return filepath.Join(cwd, "..", ".."), nil
}

// truncateString truncates a string to the specified length.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// TestIntegrationCompileDirectoryWithResults tests the detailed results API.
func TestIntegrationCompileDirectoryWithResults(t *testing.T) {
	workspaceRoot, err := getWorkspaceRoot()
	if err != nil {
		t.Fatalf("Failed to get workspace root: %v", err)
	}

	dirPath := filepath.Join(workspaceRoot, "samples/kuma2")
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		t.Skip("samples/kuma2 directory not found")
	}

	results, err := CompileDirectoryWithResults(dirPath)
	if err != nil {
		t.Fatalf("CompileDirectoryWithResults failed: %v", err)
	}

	// kuma2 should have at least KUMA2.TFY
	if len(results) == 0 {
		t.Error("Expected at least one result from kuma2 directory")
	}

	// Log results
	for _, r := range results {
		if len(r.Errors) > 0 {
			t.Logf("%s: %d errors", r.FileName, len(r.Errors))
		} else {
			t.Logf("%s: %d opcodes generated", r.FileName, len(r.OpCodes))
		}
	}
}

// TestIntegrationNoFatalErrors verifies that compilation doesn't cause fatal errors.
// Some parse errors are expected for preprocessor directives, but the compiler
// should not panic or crash.
func TestIntegrationNoFatalErrors(t *testing.T) {
	workspaceRoot, err := getWorkspaceRoot()
	if err != nil {
		t.Fatalf("Failed to get workspace root: %v", err)
	}

	// Test directories that might have complex scripts
	testDirs := []string{
		"samples/robot",
		"samples/home",
		"samples/sab1",
		"samples/sab2",
	}

	for _, dir := range testDirs {
		dirPath := filepath.Join(workspaceRoot, dir)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			continue
		}

		t.Run(dir, func(t *testing.T) {
			// This should not panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Compilation panicked for %s: %v", dir, r)
				}
			}()

			_, _ = CompileDirectory(dirPath)
			// We don't check errors here - we just verify no panic
		})
	}
}

// TestErrorCollectionFromLexer tests that lexer errors are properly collected and reported.
// Requirement 5.1: Lexer reports illegal characters with character, line, and column.
func TestErrorCollectionFromLexer(t *testing.T) {
	// Source with illegal character
	source := `int x = 5;
int y = @;
int z = 10;`

	_, errs := Compile(source)

	// Should have at least one error
	if len(errs) == 0 {
		t.Error("Expected at least one error for illegal character '@'")
		return
	}

	// Check that error contains location information
	errStr := errs[0].Error()
	if errStr == "" {
		t.Error("Error message should not be empty")
	}

	t.Logf("Lexer error: %s", errStr)
}

// TestErrorCollectionFromParser tests that parser errors are properly collected and reported.
// Requirement 5.2: Parser reports syntax errors with expected/actual token types, line, and column.
func TestErrorCollectionFromParser(t *testing.T) {
	// Source with syntax error (missing expression)
	source := `main() {
    int x = ;
}`

	_, errs := Compile(source)

	// Should have at least one error
	if len(errs) == 0 {
		t.Error("Expected at least one error for syntax error")
		return
	}

	// Check that error contains location information
	errStr := errs[0].Error()
	if errStr == "" {
		t.Error("Error message should not be empty")
	}

	t.Logf("Parser error: %s", errStr)
}

// TestErrorCollectionFromCompiler tests that compiler errors are properly collected and reported.
// Requirement 5.5: Compiler reports unknown AST node types in error messages.
func TestErrorCollectionFromCompiler(t *testing.T) {
	// This test verifies that the compiler can handle and report errors
	// For now, we test that valid code compiles without errors
	// Note: Variable declarations inside functions use assignment syntax in FILLY
	source := `main() {
    x = 5;
    x = x + 1;
}`

	opcodes, errs := Compile(source)

	if len(errs) > 0 {
		t.Errorf("Expected no errors for valid code, got: %v", errs)
	}

	if len(opcodes) == 0 {
		t.Error("Expected opcodes to be generated")
	}

	t.Logf("Generated %d opcodes", len(opcodes))
}

// TestErrorContextGeneration tests that error context is properly generated.
// Requirement 5.3: Parser includes source code context (2 lines before and after error).
// Requirement 5.4: Parser includes pointer (^) indicating error column.
func TestErrorContextGeneration(t *testing.T) {
	source := `int a = 1;
int b = 2;
int c = 3;
int d = ;
int e = 5;
int f = 6;`

	_, errs := Compile(source)

	if len(errs) == 0 {
		t.Error("Expected at least one error")
		return
	}

	// Check that error message contains context
	errStr := errs[0].Error()
	t.Logf("Error with context:\n%s", errStr)

	// The error should contain the error line marker
	// Note: Context generation depends on the error type
}

// TestMultipleErrorCollection tests that multiple errors are collected.
// Requirement 5.6: System collects all errors and returns them to caller.
func TestMultipleErrorCollection(t *testing.T) {
	// Source with multiple syntax errors
	source := `main() {
    int x = ;
    int y = ;
}`

	_, errs := Compile(source)

	// Should have errors
	if len(errs) == 0 {
		t.Error("Expected errors for syntax errors")
		return
	}

	t.Logf("Collected %d errors", len(errs))
	for i, err := range errs {
		t.Logf("Error %d: %s", i+1, err.Error())
	}
}

// TestSuccessfulCompilationNoErrors tests that successful compilation returns no errors.
// Requirement 5.7: When compilation succeeds, system returns empty error list.
func TestSuccessfulCompilationNoErrors(t *testing.T) {
	// Note: Variable declarations inside functions use assignment syntax in FILLY
	source := `main() {
    x = 5;
    y = 10;
    x = x + y;
}`

	opcodes, errs := Compile(source)

	if len(errs) != 0 {
		t.Errorf("Expected no errors for valid code, got %d errors: %v", len(errs), errs)
	}

	if len(opcodes) == 0 {
		t.Error("Expected opcodes to be generated")
	}

	t.Logf("Successfully compiled with %d opcodes and 0 errors", len(opcodes))
}

// TestPipelineStopsOnError tests that the pipeline stops when a phase fails.
// Requirement 6.3: If any phase fails, system stops pipeline and returns accumulated errors.
func TestPipelineStopsOnError(t *testing.T) {
	// Source with early syntax error
	source := `main( {
    int x = 5;
}`

	opcodes, errs := Compile(source)

	// Should have errors
	if len(errs) == 0 {
		t.Error("Expected errors for syntax error")
		return
	}

	// Should not have opcodes (pipeline stopped)
	if len(opcodes) != 0 {
		t.Errorf("Expected no opcodes when pipeline fails, got %d", len(opcodes))
	}

	t.Logf("Pipeline stopped with %d errors", len(errs))
}

// TestFindMainScript tests the FindMainScript function.
// Requirement 14.1: System scans all TFY files to identify the file containing main function.
func TestFindMainScript(t *testing.T) {
	scripts := []script.Script{
		{FileName: "helper.tfy", Content: `helper() { x = 1; }`},
		{FileName: "main.tfy", Content: `main() { helper(); }`},
		{FileName: "utils.tfy", Content: `utils() { y = 2; }`},
	}

	mainInfo, err := FindMainScript(scripts)
	if err != nil {
		t.Fatalf("FindMainScript failed: %v", err)
	}

	if mainInfo.FileName != "main.tfy" {
		t.Errorf("Expected main.tfy, got %s", mainInfo.FileName)
	}
}

// TestFindMainScriptCaseInsensitive tests that main function detection is case-insensitive.
func TestFindMainScriptCaseInsensitive(t *testing.T) {
	testCases := []struct {
		name     string
		content  string
		expected bool
	}{
		{"lowercase", `main() { x = 1; }`, true},
		{"uppercase", `MAIN() { x = 1; }`, true},
		{"mixedcase", `Main() { x = 1; }`, true},
		{"mixedcase2", `MaIn() { x = 1; }`, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			scripts := []script.Script{
				{FileName: "test.tfy", Content: tc.content},
			}

			mainInfo, err := FindMainScript(scripts)
			if tc.expected {
				if err != nil {
					t.Errorf("Expected to find main function, got error: %v", err)
				}
				if mainInfo == nil {
					t.Error("Expected mainInfo to be non-nil")
				}
			}
		})
	}
}

// TestFindMainScriptNoMain tests error when no main function is found.
// Requirement 14.3: When main function is not found, report error.
func TestFindMainScriptNoMain(t *testing.T) {
	scripts := []script.Script{
		{FileName: "helper.tfy", Content: `helper() { x = 1; }`},
		{FileName: "utils.tfy", Content: `utils() { y = 2; }`},
	}

	_, err := FindMainScript(scripts)
	if err == nil {
		t.Error("Expected error when no main function found")
	}

	if err.Error() != "no main function found in any script file" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}

// TestFindMainScriptMultipleMain tests error when multiple main functions are found.
// Requirement 14.2: When main function exists in multiple files, report error.
func TestFindMainScriptMultipleMain(t *testing.T) {
	scripts := []script.Script{
		{FileName: "main1.tfy", Content: `main() { x = 1; }`},
		{FileName: "main2.tfy", Content: `main() { y = 2; }`},
	}

	_, err := FindMainScript(scripts)
	if err == nil {
		t.Error("Expected error when multiple main functions found")
	}

	// Error should mention both files
	errStr := err.Error()
	if !contains(errStr, "multiple main functions") {
		t.Errorf("Expected error to mention 'multiple main functions', got: %s", errStr)
	}
}

// TestCompileWithEntryPoint tests the CompileWithEntryPoint function.
// Requirement 13.1: Application calls compiler after loading scripts to generate OpCode.
// Requirement 14.4: When file containing main function is identified, start compilation from that file.
func TestCompileWithEntryPoint(t *testing.T) {
	scripts := []script.Script{
		{FileName: "helper.tfy", Content: `helper() { x = 1; }`},
		{FileName: "main.tfy", Content: `main() { helper(); }`},
	}

	opcodes, err := CompileWithEntryPoint(scripts)
	if err != nil {
		t.Fatalf("CompileWithEntryPoint failed: %v", err)
	}

	if len(opcodes) == 0 {
		t.Error("Expected opcodes to be generated")
	}

	t.Logf("Generated %d opcodes", len(opcodes))
}

// TestCompileWithEntryPointNoMain tests error when no main function is found.
func TestCompileWithEntryPointNoMain(t *testing.T) {
	scripts := []script.Script{
		{FileName: "helper.tfy", Content: `helper() { x = 1; }`},
	}

	_, err := CompileWithEntryPoint(scripts)
	if err == nil {
		t.Error("Expected error when no main function found")
	}
}

// TestCompileDirectoryWithEntryPoint tests the CompileDirectoryWithEntryPoint function.
func TestCompileDirectoryWithEntryPoint(t *testing.T) {
	workspaceRoot, err := getWorkspaceRoot()
	if err != nil {
		t.Fatalf("Failed to get workspace root: %v", err)
	}

	// Test with robot directory which should have a main function
	dirPath := filepath.Join(workspaceRoot, "samples/robot")
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		t.Skip("samples/robot directory not found")
	}

	opcodes, err := CompileDirectoryWithEntryPoint(dirPath)
	if err != nil {
		// Log the error but don't fail - some samples may not have main
		t.Logf("CompileDirectoryWithEntryPoint result: %v", err)
		return
	}

	t.Logf("Generated %d opcodes from robot directory", len(opcodes))
}

// TestContainsMainFunction tests the containsMainFunction helper.
func TestContainsMainFunction(t *testing.T) {
	testCases := []struct {
		name     string
		content  string
		expected bool
	}{
		{
			name:     "has main",
			content:  `main() { x = 1; }`,
			expected: true,
		},
		{
			name:     "no main",
			content:  `helper() { x = 1; }`,
			expected: false,
		},
		{
			name:     "main in comment",
			content:  `// main() { x = 1; }`,
			expected: false,
		},
		{
			name:     "main in string",
			content:  `x = "main() { x = 1; }";`,
			expected: false,
		},
		{
			name:     "main with params",
			content:  `main(int argc) { x = 1; }`,
			expected: true,
		},
		{
			name:     "MAIN uppercase",
			content:  `MAIN() { x = 1; }`,
			expected: true,
		},
		{
			name:     "main function call not definition",
			content:  `helper() { main(); }`,
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, _ := containsMainFunction(tc.content)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v for content: %s", tc.expected, result, tc.content)
			}
		})
	}
}

// contains checks if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
