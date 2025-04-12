// filename: pkg/core/tools_fs_read_test.go
package core

import (
	// Keep errors
	"fmt" // Keep fmt
	"os"
	"path/filepath" // Keep filepath

	// Keep strings for error message check
	"testing"
)

// Remove the duplicate testFsToolHelper function definition here
/*
func testFsToolHelper(...) {
    ... // REMOVED
}
*/

func TestToolReadFile(t *testing.T) {
	interp, sandboxDir := newDefaultTestInterpreter(t) // Get interpreter and sandbox

	// Setup test file
	testFilePathRel := "readTest.txt"
	// testFilePathAbs := filepath.Join(sandboxDir, testFilePathRel) // Not needed directly
	testContent := "Hello\nWorld!"
	err := os.WriteFile(testFilePathRel, []byte(testContent), 0644) // Write relative to sandbox (CWD)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	// Cleanup handled by t.TempDir()

	// Create a directory to test reading a directory
	testDirPathRel := "readTestDir"
	// testDirPathAbs := filepath.Join(sandboxDir, testDirPathRel) // Not needed directly
	os.Mkdir(testDirPathRel, 0755) // Create relative to sandbox (CWD)
	// Cleanup handled by t.TempDir()

	// Use the unified fsTestCase struct
	tests := []fsTestCase{
		{
			name:       "Read Existing File",
			toolName:   "ReadFile",
			args:       makeArgs(testFilePathRel), // Use relative path for arg
			wantResult: testContent,
		},
		{
			name:       "Read Non-Existent File",
			toolName:   "ReadFile",
			args:       makeArgs("nonexistent.txt"),
			wantResult: "ReadFile failed: File not found at path 'nonexistent.txt'", // Expect error message string
			// Expect the tool to return ErrInternalTool wrapping the os error
			wantToolErrIs: ErrInternalTool,
		},
		{
			name:     "Read Directory",
			toolName: "ReadFile",
			args:     makeArgs(testDirPathRel), // Use relative path
			// Expect specific error message string
			wantResult:    fmt.Sprintf("ReadFile failed for '%s': read %s: is a directory", testDirPathRel, filepath.Join(sandboxDir, testDirPathRel)),
			wantToolErrIs: ErrInternalTool, // Expect the tool to return ErrInternalTool wrapping the os error
		},
		{
			name:         "Validation_Wrong_Arg_Type",
			toolName:     "ReadFile",
			args:         makeArgs(123),             // Invalid type
			valWantErrIs: ErrValidationTypeMismatch, // Expect validation type error
		},
		{
			name:         "Validation_Missing_Arg",
			toolName:     "ReadFile",
			args:         makeArgs(),            // Missing arg
			valWantErrIs: ErrValidationArgCount, // Expect validation count error
		},
		{
			name:       "Path_Outside_Sandbox",
			toolName:   "ReadFile",
			args:       makeArgs("../outside.txt"),
			wantResult: fmt.Sprintf("ReadFile path error for '../outside.txt': %s: relative path '../outside.txt' resolves to '%s' which is outside the allowed directory '%s'", ErrPathViolation.Error(), filepath.Clean(filepath.Join(sandboxDir, "../outside.txt")), sandboxDir), // Construct expected error message
			// *** MODIFIED: Expect ErrPathViolation from tool ***
			wantToolErrIs: ErrPathViolation,
		},
		{
			name:         "Validation_Nil_Arg",
			toolName:     "ReadFile",
			args:         makeArgs(nil),               // Pass nil for path
			valWantErrIs: ErrValidationRequiredArgNil, // Expect validation error
		},
	}

	for _, tt := range tests {
		// Pass interp and tt to the helper in testing_helpers_test.go
		// Need to adjust the helper slightly if we want to check *both*
		// the specific error type (wantToolErrIs) AND the string result (wantResult)
		// when an error occurs. Let's assume helper checks error first, then result if no error or if wantResult is set.
		testFsToolHelper(t, interp, tt)
	}
}
