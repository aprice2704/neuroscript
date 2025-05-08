// NeuroScript Version: 0.3.1
// File version: 0.0.3 // Remove unused interp variable.
// nlines: 110 // Approximate
// risk_rating: LOW
// filename: pkg/core/tools_fs_read_test.go
package core

import (
	"errors" // Required for errors package
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// Assume testFsToolHelper is defined in tools_fs_helpers_test.go

func TestToolReadFile(t *testing.T) {
	// --- Test Setup Data ---
	readTestFile := "readTest.txt"
	readTestContent := "Hello Reader"
	readTestDir := "readTestDir" // Directory to test reading

	// --- Setup Function ---
	setupReadFileTest := func(sandboxRoot string) error {
		filePath := filepath.Join(sandboxRoot, readTestFile)
		dirPath := filepath.Join(sandboxRoot, readTestDir)

		// Clean up potential leftovers
		os.Remove(filePath)
		os.RemoveAll(dirPath) // Use RemoveAll for directories

		// Create file
		if err := os.WriteFile(filePath, []byte(readTestContent), 0644); err != nil {
			return fmt.Errorf("setup WriteFile failed for %s: %w", filePath, err)
		}
		// Create directory
		if err := os.Mkdir(dirPath, 0755); err != nil && !errors.Is(err, os.ErrExist) {
			return fmt.Errorf("setup Mkdir failed for %s: %w", dirPath, err)
		}
		return nil
	}

	// --- Test Cases ---
	// Using fsTestCase now requires defining toolName for the helper
	tests := []fsTestCase{
		{
			name:       "Read Existing File",
			toolName:   "FS.Read", // Added toolName
			args:       MakeArgs(readTestFile),
			setupFunc:  setupReadFileTest,
			wantResult: readTestContent, // Expect file content string
		},
		{
			name:          "Read Non-Existent File",
			toolName:      "FS.Read", // Added toolName
			args:          MakeArgs("nonexistent.txt"),
			setupFunc:     setupReadFileTest,
			wantResult:    "file not found 'nonexistent.txt'", // Expect specific error message substring
			wantToolErrIs: ErrFileNotFound,                    // Expect specific error type
		},
		{
			name:          "Read Directory",
			toolName:      "FS.Read", // Added toolName
			args:          MakeArgs(readTestDir),
			setupFunc:     setupReadFileTest,
			wantResult:    "path 'readTestDir' is a directory", // Expect specific error message substring
			wantToolErrIs: ErrPathNotFile,                      // Expect specific error type
		},
		{
			name:          "Validation_Wrong_Arg_Type",
			toolName:      "FS.Read", // Added toolName
			args:          MakeArgs(123),
			wantResult:    "filepath argument must be a string", // Error substring
			wantToolErrIs: ErrInvalidArgument,
		},
		{
			name:          "Validation_Missing_Arg",
			toolName:      "FS.Read", // Added toolName
			args:          MakeArgs(),
			wantResult:    "expected 1 argument", // Error substring
			wantToolErrIs: ErrArgumentMismatch,
		},
		{
			name:          "Path_Outside_Sandbox",
			toolName:      "FS.Read", // Added toolName
			args:          MakeArgs("../outside.txt"),
			setupFunc:     setupReadFileTest,
			wantResult:    "path resolves outside allowed directory", // Error substring
			wantToolErrIs: ErrPathViolation,                          // Expect specific error type
		},
		{
			name:          "Validation_Nil_Arg",
			toolName:      "FS.Read", // Added toolName
			args:          MakeArgs(nil),
			wantResult:    "filepath argument must be a string", // Error substring
			wantToolErrIs: ErrInvalidArgument,
		},
		{
			name:          "Validation_Empty_Filepath_Arg",
			toolName:      "FS.Read", // Added toolName
			args:          MakeArgs(""),
			wantResult:    "filepath argument cannot be empty", // Error substring
			wantToolErrIs: ErrInvalidArgument,
		},
	}

	// Run tests using the helper
	// Removed outer interp declaration - helper creates one per test case
	for _, tt := range tests {
		// Run tests within t.Run for isolation
		t.Run(tt.name, func(t *testing.T) {
			interp, currentSandbox := NewDefaultTestInterpreter(t)
			// Ensure setup runs in the correct sandbox
			if tt.setupFunc != nil {
				if err := tt.setupFunc(currentSandbox); err != nil {
					t.Fatalf("Setup function failed: %v", err)
				}
			}
			// Pass the interpreter created for this subtest
			testFsToolHelper(t, interp, tt)
		})
	}
}
