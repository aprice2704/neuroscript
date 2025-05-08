// NeuroScript Version: 0.3.1
// File version: 0.0.3 // Update nil content test expectation to success.
// nlines: 135 // Approximate
// risk_rating: MEDIUM // Tests file writing
// filename: pkg/core/tools_fs_write_test.go
package core

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// Assume testFsToolHelper is defined in tools_fs_helpers_test.go

func TestToolWriteFile(t *testing.T) {
	// --- Test Setup Data ---
	writeNewFile := "newWrite.txt"
	overwriteExistingFile := "overwrite.txt"
	existingContent := "Initial Content"
	newContent := "This is the new content."
	emptyContentFile := "emptyFile.txt"
	nestedFile := filepath.Join("newdir", "nestedfile.txt")

	// --- Setup Function ---
	setupWriteFileTest := func(sandboxRoot string) error {
		existingPath := filepath.Join(sandboxRoot, overwriteExistingFile)
		if err := os.WriteFile(existingPath, []byte(existingContent), 0644); err != nil && !errors.Is(err, os.ErrExist) {
			return fmt.Errorf("setup WriteFile failed for %s: %w", existingPath, err)
		}
		os.Remove(filepath.Join(sandboxRoot, writeNewFile))
		os.Remove(filepath.Join(sandboxRoot, emptyContentFile))
		os.RemoveAll(filepath.Join(sandboxRoot, "newdir"))
		return nil
	}

	// --- Test Cases ---
	tests := []fsTestCase{
		{
			name:        "Write New File",
			toolName:    "FS.Write",
			args:        MakeArgs(writeNewFile, newContent),
			wantResult:  fmt.Sprintf("Successfully wrote %d bytes to %s", len(newContent), writeNewFile),
			wantContent: newContent,
		},
		{
			name:        "Overwrite Existing File",
			toolName:    "FS.Write",
			args:        MakeArgs(overwriteExistingFile, newContent),
			setupFunc:   setupWriteFileTest,
			wantResult:  fmt.Sprintf("Successfully wrote %d bytes to %s", len(newContent), overwriteExistingFile),
			wantContent: newContent,
		},
		{
			name:        "Write Empty Content",
			toolName:    "FS.Write",
			args:        MakeArgs(emptyContentFile, ""),
			wantResult:  fmt.Sprintf("Successfully wrote %d bytes to %s", 0, emptyContentFile),
			wantContent: "",
		},
		{
			name:        "Create Subdirectory",
			toolName:    "FS.Write",
			args:        MakeArgs(nestedFile, newContent),
			wantResult:  fmt.Sprintf("Successfully wrote %d bytes to %s", len(newContent), nestedFile),
			wantContent: newContent,
		},
		{
			name:          "Validation_Wrong_Path_Type",
			toolName:      "FS.Write",
			args:          MakeArgs(123, newContent),
			wantResult:    "filepath argument must be a string",
			wantToolErrIs: ErrInvalidArgument,
		},
		{
			name:          "Validation_Wrong_Content_Type",
			toolName:      "FS.Write",
			args:          MakeArgs(writeNewFile, 456),
			wantResult:    "content argument must be a string or nil", // Updated error check
			wantToolErrIs: ErrInvalidArgument,
		},
		{
			name:          "Validation_Missing_Content",
			toolName:      "FS.Write",
			args:          MakeArgs(writeNewFile),
			wantResult:    "expected 2 arguments",
			wantToolErrIs: ErrArgumentMismatch,
		},
		{
			name:          "Validation_Missing_Path",
			toolName:      "FS.Write",
			args:          MakeArgs(),
			wantResult:    "expected 2 arguments",
			wantToolErrIs: ErrArgumentMismatch,
		},
		{
			name:          "Validation_Empty_Path",
			toolName:      "FS.Write",
			args:          MakeArgs("", newContent),
			wantResult:    "filepath argument cannot be empty",
			wantToolErrIs: ErrInvalidArgument,
		},
		{
			name:          "Validation_Nil_Path",
			toolName:      "FS.Write",
			args:          MakeArgs(nil, newContent),
			wantResult:    "filepath argument must be a string",
			wantToolErrIs: ErrInvalidArgument,
		},
		{
			name:        "Validation_Nil_Content", // <<< CORRECTED Expectation
			toolName:    "FS.Write",
			args:        MakeArgs(writeNewFile, nil),                                       // Pass nil for content
			wantResult:  fmt.Sprintf("Successfully wrote %d bytes to %s", 0, writeNewFile), // Expect success msg for 0 bytes
			wantContent: "",                                                                // Expect empty file content
			// wantToolErrIs is nil (no error expected)
		},
		{
			name:          "Path_Outside_Sandbox",
			toolName:      "FS.Write",
			args:          MakeArgs("../outside.txt", newContent),
			wantResult:    "path resolves outside allowed directory",
			wantToolErrIs: ErrPathViolation,
		},
	}

	// Run tests using the standard helper
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interp, currentSandbox := NewDefaultTestInterpreter(t)
			if tt.setupFunc != nil {
				if err := tt.setupFunc(currentSandbox); err != nil {
					t.Fatalf("Setup function failed: %v", err)
				}
			}
			testFsToolHelper(t, interp, tt) // testFsToolHelper handles content verification
		})
	}
}
