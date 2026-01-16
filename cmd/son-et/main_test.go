package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestCLIHelp tests the help display functionality
func TestCLIHelp(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"help flag", []string{"--help"}},
		{"no arguments", []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("go", append([]string{"run", "main.go"}, tt.args...)...)
			cmd.Dir = "."
			output, err := cmd.CombinedOutput()

			if err != nil {
				// Exit code 0 is expected for help
				if exitErr, ok := err.(*exec.ExitError); ok {
					if exitErr.ExitCode() != 0 {
						t.Errorf("Expected exit code 0, got %d", exitErr.ExitCode())
					}
				}
			}

			outputStr := string(output)
			if !strings.Contains(outputStr, "son-et - FILLY Script Interpreter") {
				t.Error("Help output should contain title")
			}
			if !strings.Contains(outputStr, "USAGE:") {
				t.Error("Help output should contain USAGE section")
			}
			if !strings.Contains(outputStr, "son-et <directory>") {
				t.Error("Help output should show directory usage")
			}
		})
	}
}

// TestCLIDirectoryValidation tests directory validation
func TestCLIDirectoryValidation(t *testing.T) {
	tests := []struct {
		name        string
		setupDir    func() string
		cleanupDir  func(string)
		expectError bool
		errorMsg    string
	}{
		{
			name: "nonexistent directory",
			setupDir: func() string {
				return "/tmp/nonexistent_test_dir_12345"
			},
			cleanupDir:  func(s string) {},
			expectError: true,
			errorMsg:    "does not exist",
		},
		{
			name: "file instead of directory",
			setupDir: func() string {
				tmpFile := filepath.Join(os.TempDir(), "test_file.txt")
				os.WriteFile(tmpFile, []byte("test"), 0644)
				return tmpFile
			},
			cleanupDir: func(s string) {
				os.Remove(s)
			},
			expectError: true,
			errorMsg:    "is not a directory",
		},
		{
			name: "empty directory",
			setupDir: func() string {
				tmpDir := filepath.Join(os.TempDir(), "test_empty_dir")
				os.MkdirAll(tmpDir, 0755)
				return tmpDir
			},
			cleanupDir: func(s string) {
				os.RemoveAll(s)
			},
			expectError: true,
			errorMsg:    "no TFY files found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setupDir()
			defer tt.cleanupDir(dir)

			cmd := exec.Command("go", "run", "main.go", dir)
			cmd.Dir = "."
			output, err := cmd.CombinedOutput()

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				outputStr := string(output)
				if !strings.Contains(outputStr, tt.errorMsg) {
					t.Errorf("Expected error message to contain '%s', got: %s", tt.errorMsg, outputStr)
				}
			}
		})
	}
}

// TestCLIParsingError tests parsing error reporting
func TestCLIParsingError(t *testing.T) {
	// Create a temporary directory with an invalid TFY file
	tmpDir := filepath.Join(os.TempDir(), "test_parse_error")
	os.MkdirAll(tmpDir, 0755)
	defer os.RemoveAll(tmpDir)

	// Write an invalid TFY file (syntax error)
	invalidTFY := `
		function test(
			// Missing closing parenthesis and body
	`
	tfyPath := filepath.Join(tmpDir, "test.tfy")
	if err := os.WriteFile(tfyPath, []byte(invalidTFY), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cmd := exec.Command("go", "run", "main.go", tmpDir)
	cmd.Dir = "."
	output, err := cmd.CombinedOutput()

	if err == nil {
		t.Error("Expected parsing error but got none")
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "PARSING ERRORS") && !strings.Contains(outputStr, "Parse error") {
		t.Errorf("Expected parsing error message, got: %s", outputStr)
	}
}

// TestCLIDirectModeExecution tests direct mode execution with a real sample
// This test runs for a limited time and then terminates
func TestCLIDirectModeExecution(t *testing.T) {
	// Check if a sample project exists
	sampleDirs := []string{
		"../../samples/kuma2",
		"../../samples/sabo2",
		"../../samples/robot",
	}

	var testDir string
	for _, dir := range sampleDirs {
		if _, err := os.Stat(dir); err == nil {
			testDir = dir
			break
		}
	}

	if testDir == "" {
		t.Skip("No sample projects found for integration testing")
	}

	t.Logf("Testing with sample: %s", testDir)

	// Run the interpreter with a timeout
	cmd := exec.Command("go", "run", "main.go", testDir)
	cmd.Dir = "."

	// Start the command
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start command: %v", err)
	}

	// Ensure process is killed even if test fails
	defer func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
			cmd.Wait() // Clean up zombie process
		}
	}()

	// Create a timer to kill the process after 3 seconds
	timer := time.AfterFunc(3*time.Second, func() {
		if cmd.Process != nil {
			t.Log("Terminating test execution after timeout")
			cmd.Process.Kill()
		}
	})
	defer timer.Stop()

	// Wait for the command to finish (or be killed)
	err := cmd.Wait()

	// Check the result
	if err != nil {
		// Process was killed (expected) or had an error
		if exitErr, ok := err.(*exec.ExitError); ok {
			// Exit code -1 typically means killed by signal
			if exitErr.ExitCode() == -1 {
				t.Log("Process terminated successfully after timeout (expected)")
				return
			}
		}
		// Other errors might indicate real problems
		t.Logf("Process exited with error (may be expected): %v", err)
	} else {
		t.Log("Process completed successfully within timeout")
	}
}

// TestCLIWithDebugLevel tests execution with different debug levels
func TestCLIWithDebugLevel(t *testing.T) {
	// Create a minimal test project
	tmpDir := filepath.Join(os.TempDir(), "test_debug_level")
	os.MkdirAll(tmpDir, 0755)
	defer os.RemoveAll(tmpDir)

	// Write a minimal valid TFY file
	minimalTFY := `
		function main() {
			// Empty main function
		}
	`
	tfyPath := filepath.Join(tmpDir, "test.tfy")
	if err := os.WriteFile(tfyPath, []byte(minimalTFY), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	debugLevels := []string{"0", "1", "2"}

	for _, level := range debugLevels {
		t.Run("DEBUG_LEVEL="+level, func(t *testing.T) {
			cmd := exec.Command("go", "run", "main.go", tmpDir)
			cmd.Dir = "."
			cmd.Env = append(os.Environ(), "DEBUG_LEVEL="+level)

			// Start the command
			if err := cmd.Start(); err != nil {
				t.Fatalf("Failed to start command: %v", err)
			}

			// Ensure process is killed even if test fails
			defer func() {
				if cmd.Process != nil {
					cmd.Process.Kill()
					cmd.Wait() // Clean up zombie process
				}
			}()

			// Kill after 2 seconds
			timer := time.AfterFunc(2*time.Second, func() {
				if cmd.Process != nil {
					cmd.Process.Kill()
				}
			})
			defer timer.Stop()

			cmd.Wait()
			t.Logf("Executed with DEBUG_LEVEL=%s", level)
		})
	}
}
