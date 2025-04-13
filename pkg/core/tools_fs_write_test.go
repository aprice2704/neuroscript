// filename: pkg/core/tools_fs_write_test.go
package core

import (
	// Keep errors
	"fmt"           // Keep fmt
	"os"            // Keep os
	"path/filepath" // Keep filepath
	"testing"
)

// Assume testFsToolHelper is defined in testing_helpers_test.go

func TestToolWriteFile(t *testing.T) {
	interp, sandboxDirAbs := newDefaultTestInterpreter(t) // Get interpreter and sandbox path

	// --- Test Setup Data ---
	newWriteFileRel := "newWrite.txt"
	overwriteTargetRel := "overwrite.txt"
	overwriteInitialContent := "initial content"
	nestedFileRel := filepath.Join("newdir", "nestedfile.txt")
	writeContent := "this is the written content"
	emptyContent := ""

	// --- Setup Function ---
	// *** MODIFIED: Takes sandboxRoot string argument and uses it ***
	setupWriteFileTest := func(sandboxRoot string) error {
		// Construct absolute path for the file to be overwritten
		overwriteTargetAbs := filepath.Join(sandboxRoot, overwriteTargetRel)

		// Create the file that will be overwritten
		if err := os.WriteFile(overwriteTargetAbs, []byte(overwriteInitialContent), 0644); err != nil {
			return fmt.Errorf("setup WriteFile failed for %s: %w", overwriteTargetAbs, err)
		}
		return nil
	}

	tests := []fsTestCase{
		{
			name:        "Write New File",
			toolName:    "WriteFile",
			args:        makeArgs(newWriteFileRel, writeContent),
			setupFunc:   nil, // No setup needed for writing new file
			wantResult:  "OK",
			wantContent: writeContent, // Verify content using helper
		},
		{
			name:        "Overwrite Existing File",
			toolName:    "WriteFile",
			args:        makeArgs(overwriteTargetRel, writeContent),
			setupFunc:   setupWriteFileTest, // Setup the initial file
			wantResult:  "OK",
			wantContent: writeContent, // Verify new content
		},
		{
			name:        "Write Empty Content",
			toolName:    "WriteFile",
			args:        makeArgs("emptyFile.txt", emptyContent),
			setupFunc:   nil,
			wantResult:  "OK",
			wantContent: emptyContent,
		},
		{
			name:        "Create Subdirectory", // WriteFile should create parent dirs
			toolName:    "WriteFile",
			args:        makeArgs(nestedFileRel, writeContent),
			setupFunc:   nil, // No setup needed, dir creation is part of the test
			wantResult:  "OK",
			wantContent: writeContent,
		},
		{
			name:         "Validation_Wrong_Path_Type",
			toolName:     "WriteFile",
			args:         makeArgs(123, writeContent),
			valWantErrIs: ErrValidationTypeMismatch,
		},
		{
			name:         "Validation_Wrong_Content_Type",
			toolName:     "WriteFile",
			args:         makeArgs(newWriteFileRel, 456),
			valWantErrIs: ErrValidationTypeMismatch,
		},
		{
			name:         "Validation_Missing_Content",
			toolName:     "WriteFile",
			args:         makeArgs(newWriteFileRel),
			valWantErrIs: ErrValidationArgCount,
		},
		{
			name:          "Path_Outside_Sandbox",
			toolName:      "WriteFile",
			args:          makeArgs("../outside.txt", writeContent),
			setupFunc:     nil,
			wantResult:    fmt.Sprintf("WriteFile path error for '../outside.txt': %s: relative path '../outside.txt' resolves to '%s' which is outside the allowed directory '%s'", ErrPathViolation.Error(), filepath.Clean(filepath.Join(sandboxDirAbs, "../outside.txt")), sandboxDirAbs),
			wantToolErrIs: ErrPathViolation,
		},
		{
			// Trying to write to a path that represents the sandbox root directory itself
			// Note: SecureFilePath prevents writing directly to the root ('./') but allows files *in* root.
			// Let's test writing *to* a directory path instead.
			name:          "Write_To_Directory_Path",
			toolName:      "WriteFile",
			args:          makeArgs("newdir", writeContent), // Path used for nestedFileRel
			setupFunc:     setupWriteFileTest,               // Creates the file to overwrite only
			wantResult:    fmt.Sprintf("WriteFile mkdir failed for dir '%s': mkdir %s: not a directory", filepath.Join(sandboxDirAbs, "newdir"), filepath.Join(sandboxDirAbs, "newdir")),
			wantToolErrIs: ErrInternalTool, // Should fail during MkdirAll check inside WriteFile
		},
		{
			name:         "Validation_Nil_Path",
			toolName:     "WriteFile",
			args:         makeArgs(nil, writeContent),
			valWantErrIs: ErrValidationRequiredArgNil,
		},
		{
			name:         "Validation_Nil_Content",
			toolName:     "WriteFile",
			args:         makeArgs(newWriteFileRel, nil),
			valWantErrIs: ErrValidationRequiredArgNil,
		},
	}

	for _, tt := range tests {
		testFsToolHelper(t, interp, tt)
		// Note: File content verification is now handled *within* testFsToolHelper via wantContent field
	}
}
