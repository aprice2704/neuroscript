// filename: pkg/core/tools_fs_delete_test.go
package core

import (
	"errors"        // Keep errors
	"fmt"           // Keep fmt
	"os"            // Keep os
	"path/filepath" // Keep filepath
	"testing"
)

// Assume testFsToolHelper is defined in tools_fs_helpers_test.go

func TestToolDeleteFile(t *testing.T) {
	interp, _ := NewDefaultTestInterpreter(t) // Get interpreter and sandbox path

	// --- Test Setup Data ---
	fileToDeleteRel := "deleteMe.txt"
	dirToDeleteRel := "deleteMeDir" // Should be empty
	nonEmptyDirRel := "dontDeleteMeDir"
	nonEmptyFileRel := filepath.Join(nonEmptyDirRel, "keepMe.txt")
	fileToDeleteContent := "some content"

	// --- Setup Function ---
	setupDeleteFileTest := func(sandboxRoot string) error {
		fileToDeleteAbs := filepath.Join(sandboxRoot, fileToDeleteRel)
		dirToDeleteAbs := filepath.Join(sandboxRoot, dirToDeleteRel)
		nonEmptyDirAbs := filepath.Join(sandboxRoot, nonEmptyDirRel)
		nonEmptyFileAbs := filepath.Join(sandboxRoot, nonEmptyFileRel)

		// Create file to delete
		if err := os.WriteFile(fileToDeleteAbs, []byte(fileToDeleteContent), 0644); err != nil {
			return fmt.Errorf("setup WriteFile failed for %s: %w", fileToDeleteAbs, err)
		}
		// Create empty directory to delete
		if err := os.Mkdir(dirToDeleteAbs, 0755); err != nil && !os.IsExist(err) {
			return fmt.Errorf("setup Mkdir failed for %s: %w", dirToDeleteAbs, err)
		}
		// Create non-empty directory (should fail deletion)
		if err := os.Mkdir(nonEmptyDirAbs, 0755); err != nil && !os.IsExist(err) {
			return fmt.Errorf("setup Mkdir failed for %s: %w", nonEmptyDirAbs, err)
		}
		if err := os.WriteFile(nonEmptyFileAbs, []byte("do not delete"), 0644); err != nil {
			return fmt.Errorf("setup WriteFile failed for %s: %w", nonEmptyFileAbs, err)
		}
		return nil
	}

	// --- Cleanup Function (Verify deletion) ---
	verifyDeletion := func(sandboxRoot string, pathRel string, shouldExist bool) error {
		pathAbs := filepath.Join(sandboxRoot, pathRel)
		_, err := os.Stat(pathAbs)
		if shouldExist {
			if err != nil {
				return fmt.Errorf("verify failed: expected '%s' to exist, but got error: %w", pathRel, err)
			}
		} else { // Should NOT exist
			if err == nil {
				return fmt.Errorf("verify failed: expected '%s' to be deleted, but it still exists", pathRel)
			}
			if !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("verify failed: expected '%s' to not exist, but got unexpected error: %w", pathRel, err)
			}
		}
		return nil
	}

	tests := []fsTestCase{
		{
			name:        "Delete Existing File",
			toolName:    "DeleteFile",
			args:        MakeArgs(fileToDeleteRel),
			setupFunc:   setupDeleteFileTest,
			wantResult:  "OK",
			cleanupFunc: func(sb string) error { return verifyDeletion(sb, fileToDeleteRel, false) },
		},
		{
			name:        "Delete Empty Directory",
			toolName:    "DeleteFile",
			args:        MakeArgs(dirToDeleteRel),
			setupFunc:   setupDeleteFileTest,
			wantResult:  "OK",
			cleanupFunc: func(sb string) error { return verifyDeletion(sb, dirToDeleteRel, false) },
		},
		{
			name:       "Delete Non-Existent File",
			toolName:   "DeleteFile",
			args:       MakeArgs("noSuchFile.txt"),
			setupFunc:  setupDeleteFileTest, // Setup other files, doesn't matter for this one
			wantResult: "OK",                // Returns OK if not found
		},
		{
			name:          "Delete Non-Empty Directory",
			toolName:      "DeleteFile",
			args:          MakeArgs(nonEmptyDirRel),
			setupFunc:     setupDeleteFileTest,
			wantResult:    fmt.Sprintf("DeleteFile failed for '%s': remove %s: directory not empty", nonEmptyDirRel, filepath.Join(interp.sandboxDir, nonEmptyDirRel)),
			wantToolErrIs: ErrCannotDelete,
			cleanupFunc:   func(sb string) error { return verifyDeletion(sb, nonEmptyDirRel, true) }, // Ensure it still exists
		},
		{
			name:         "Validation_Wrong_Arg_Type",
			toolName:     "DeleteFile",
			args:         MakeArgs(12345),
			valWantErrIs: ErrValidationTypeMismatch,
		},
		{
			name:          "Path_Outside_Sandbox",
			toolName:      "DeleteFile",
			args:          MakeArgs("../someFile"),
			setupFunc:     setupDeleteFileTest, // Setup doesn't matter
			wantResult:    fmt.Sprintf("DeleteFile path error for '../someFile': %s: relative path '../someFile' resolves to '%s' which is outside the allowed directory '%s'", ErrPathViolation.Error(), filepath.Clean(filepath.Join(interp.sandboxDir, "../someFile")), interp.sandboxDir),
			wantToolErrIs: ErrPathViolation,
		},
		{
			name:         "Validation_Missing_Arg",
			toolName:     "DeleteFile",
			args:         MakeArgs(),
			valWantErrIs: ErrValidationArgCount,
		},
		{
			name:         "Validation_Nil_Arg",
			toolName:     "DeleteFile",
			args:         MakeArgs(nil),
			valWantErrIs: ErrValidationRequiredArgNil,
		},
	}

	for _, tt := range tests {
		testFsToolHelper(t, interp, tt)
		// Verification is handled by cleanup funcs where appropriate
	}
}
