// filename: pkg/core/tools_fs_read_test.go
package core

import (
	// Keep errors
	"fmt"           // Keep fmt
	"os"            // Keep os
	"path/filepath" // Keep filepath
	"testing"
	// Keep strings for error message check
)

// Assume testFsToolHelper is defined in testing_helpers_test.go

func TestToolReadFile(t *testing.T) {
	interp, sandboxDirAbs := newDefaultTestInterpreter(t) // Get interpreter and sandbox path

	// --- Test Setup Data ---
	testFilePathRel := "readTest.txt"
	testDirPathRel := "readTestDir"
	testContent := "Hello\nWorld!"

	// --- Setup Function ---
	// *** MODIFIED: Takes sandboxRoot string argument and uses it ***
	setupReadFileTest := func(sandboxRoot string) error {
		// Construct absolute paths *within* the sandbox for setup
		testFileAbs := filepath.Join(sandboxRoot, testFilePathRel)
		testDirAbs := filepath.Join(sandboxRoot, testDirPathRel)

		// Create file using absolute path
		if err := os.WriteFile(testFileAbs, []byte(testContent), 0644); err != nil {
			// Don't fail if it already exists, might be needed for overwrite tests implicitly
			if !os.IsExist(err) {
				return fmt.Errorf("setup WriteFile failed for %s: %w", testFileAbs, err)
			}
		}
		// Create directory using absolute path
		if err := os.Mkdir(testDirAbs, 0755); err != nil {
			// Ignore error if directory already exists
			if !os.IsExist(err) {
				return fmt.Errorf("setup Mkdir failed for %s: %w", testDirAbs, err)
			}
		}
		return nil
	}

	tests := []fsTestCase{
		{
			name:       "Read Existing File",
			toolName:   "ReadFile",
			args:       makeArgs(testFilePathRel), // Use relative path for arg
			setupFunc:  setupReadFileTest,         // Pass setup function
			wantResult: testContent,
		},
		{
			name:          "Read Non-Existent File",
			toolName:      "ReadFile",
			args:          makeArgs("nonexistent.txt"),
			setupFunc:     setupReadFileTest,
			wantResult:    "ReadFile failed: File not found at path 'nonexistent.txt'",
			wantToolErrIs: ErrInternalTool,
		},
		{
			name:      "Read Directory",
			toolName:  "ReadFile",
			args:      makeArgs(testDirPathRel), // Use relative path
			setupFunc: setupReadFileTest,
			// Construct expected absolute path for error message comparison
			wantResult:    fmt.Sprintf("ReadFile failed for '%s': read %s: is a directory", testDirPathRel, filepath.Join(sandboxDirAbs, testDirPathRel)),
			wantToolErrIs: ErrInternalTool,
		},
		{
			name:         "Validation_Wrong_Arg_Type",
			toolName:     "ReadFile",
			args:         makeArgs(123),
			valWantErrIs: ErrValidationTypeMismatch,
		},
		{
			name:         "Validation_Missing_Arg",
			toolName:     "ReadFile",
			args:         makeArgs(),
			valWantErrIs: ErrValidationArgCount,
		},
		{
			name:          "Path_Outside_Sandbox",
			toolName:      "ReadFile",
			args:          makeArgs("../outside.txt"),
			setupFunc:     setupReadFileTest,
			wantResult:    fmt.Sprintf("ReadFile path error for '../outside.txt': %s: relative path '../outside.txt' resolves to '%s' which is outside the allowed directory '%s'", ErrPathViolation.Error(), filepath.Clean(filepath.Join(sandboxDirAbs, "../outside.txt")), sandboxDirAbs),
			wantToolErrIs: ErrPathViolation,
		},
		{
			name:         "Validation_Nil_Arg",
			toolName:     "ReadFile",
			args:         makeArgs(nil),
			valWantErrIs: ErrValidationRequiredArgNil,
		},
	}

	for _, tt := range tests {
		// Pass interp and tt to the helper
		testFsToolHelper(t, interp, tt)
	}
}
