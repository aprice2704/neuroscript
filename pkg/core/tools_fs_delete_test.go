// NeuroScript Version: 0.3.1
// File version: 0.0.4 // Fix scope issues with verifyDeletion helper.
// nlines: 145 // Approximate
// risk_rating: MEDIUM // Test file for a destructive operation
// filename: pkg/core/tools_fs_delete_test.go
package core

import (
	"errors"        // Keep errors
	"fmt"           // Keep fmt
	"os"            // Keep os
	"path/filepath" // Keep filepath

	// For checking error substrings
	"testing"
)

// Assume testFsToolHelper is defined in tools_fs_helpers_test.go

func TestToolDeleteFile(t *testing.T) {

	// --- Test Setup Data (relative paths) ---
	fileToDeleteRel := "deleteMe.txt"
	dirToDeleteRel := "deleteMeDir" // Should be empty
	nonEmptyDirRel := "dontDeleteMeDir"
	nonEmptyFileRel := filepath.Join(nonEmptyDirRel, "keepMe.txt")
	fileToDeleteContent := "some content"

	// --- Setup Function (runs in the specific sandbox for each test) ---
	setupDeleteFileTest := func(sandboxRoot string) error {
		fileToDeleteAbs := filepath.Join(sandboxRoot, fileToDeleteRel)
		dirToDeleteAbs := filepath.Join(sandboxRoot, dirToDeleteRel)
		nonEmptyDirAbs := filepath.Join(sandboxRoot, nonEmptyDirRel)
		nonEmptyFileAbs := filepath.Join(sandboxRoot, nonEmptyFileRel)

		os.Remove(fileToDeleteAbs)
		os.Remove(nonEmptyFileAbs)
		os.Remove(dirToDeleteAbs)
		os.Remove(nonEmptyDirAbs)

		if err := os.WriteFile(fileToDeleteAbs, []byte(fileToDeleteContent), 0644); err != nil {
			return fmt.Errorf("setup WriteFile failed for %s: %w", fileToDeleteAbs, err)
		}
		if err := os.Mkdir(dirToDeleteAbs, 0755); err != nil && !os.IsExist(err) {
			return fmt.Errorf("setup Mkdir failed for %s: %w", dirToDeleteAbs, err)
		}
		if err := os.Mkdir(nonEmptyDirAbs, 0755); err != nil && !os.IsExist(err) {
			return fmt.Errorf("setup Mkdir failed for %s: %w", nonEmptyDirAbs, err)
		}
		if err := os.WriteFile(nonEmptyFileAbs, []byte("do not delete"), 0644); err != nil {
			return fmt.Errorf("setup WriteFile failed for %s: %w", nonEmptyFileAbs, err)
		}
		return nil
	}

	// --- Verification Helper (defined outside loop) ---
	// Takes sandboxRoot explicitly now.
	verifyDeletion := func(sandboxRoot string, pathRel string, shouldExist bool) error {
		if sandboxRoot == "" {
			return fmt.Errorf("verifyDeletion called with empty sandboxRoot for path %s", pathRel)
		}
		pathAbs := filepath.Join(sandboxRoot, pathRel)
		_, err := os.Stat(pathAbs)
		if shouldExist {
			if err != nil {
				return fmt.Errorf("verify failed: expected '%s' (abs: %s) to exist, but got error: %w", pathRel, pathAbs, err)
			}
		} else { // Should NOT exist
			if err == nil {
				return fmt.Errorf("verify failed: expected '%s' (abs: %s) to be deleted, but it still exists", pathRel, pathAbs)
			}
			if !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("verify failed: expected '%s' (abs: %s) to not exist, but got unexpected error: %w", pathRel, pathAbs, err)
			}
		}
		return nil
	}

	// --- Test Cases ---
	tests := []fsTestCase{
		{
			name:       "Delete Existing File",
			toolName:   "FS.Delete",
			args:       MakeArgs(fileToDeleteRel),
			setupFunc:  setupDeleteFileTest,
			wantResult: "OK",
			// Original cleanup func calls the *outer* verifyDeletion
			cleanupFunc: func(sb string) error { return verifyDeletion(sb, fileToDeleteRel, false) },
		},
		{
			name:        "Delete Empty Directory",
			toolName:    "FS.Delete",
			args:        MakeArgs(dirToDeleteRel),
			setupFunc:   setupDeleteFileTest,
			wantResult:  "OK",
			cleanupFunc: func(sb string) error { return verifyDeletion(sb, dirToDeleteRel, false) },
		},
		{
			name:       "Delete Non-Existent File",
			toolName:   "FS.Delete",
			args:       MakeArgs("noSuchFile.txt"),
			setupFunc:  setupDeleteFileTest,
			wantResult: "OK",
			// No cleanup needed as nothing should change
		},
		{
			name:          "Delete Non-Empty Directory",
			toolName:      "FS.Delete",
			args:          MakeArgs(nonEmptyDirRel),
			setupFunc:     setupDeleteFileTest,
			wantResult:    "directory not empty", // Expect substring
			wantToolErrIs: ErrCannotDelete,
			cleanupFunc:   func(sb string) error { return verifyDeletion(sb, nonEmptyDirRel, true) }, // Verify it still exists
		},
		{
			name:          "Validation_Wrong_Arg_Type",
			toolName:      "FS.Delete",
			args:          MakeArgs(12345),
			wantResult:    "path argument must be a string",
			wantToolErrIs: ErrInvalidArgument,
		},
		{
			name:          "Path_Outside_Sandbox",
			toolName:      "FS.Delete",
			args:          MakeArgs("../someFile"),
			setupFunc:     setupDeleteFileTest,
			wantResult:    "path resolves outside allowed directory",
			wantToolErrIs: ErrPathViolation,
		},
		{
			name:          "Validation_Missing_Arg",
			toolName:      "FS.Delete",
			args:          MakeArgs(),
			wantResult:    "expected 1 argument",
			wantToolErrIs: ErrArgumentMismatch,
		},
		{
			name:          "Validation_Nil_Arg",
			toolName:      "FS.Delete",
			args:          MakeArgs(nil),
			wantResult:    "path argument must be a string",
			wantToolErrIs: ErrInvalidArgument,
		},
		{
			name:          "Validation_Empty_String_Arg",
			toolName:      "FS.Delete",
			args:          MakeArgs(""),
			wantResult:    "path cannot be empty",
			wantToolErrIs: ErrInvalidArgument,
		},
	}

	for _, tt := range tests {
		// Capture range variable for safety, although not running in parallel here
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			interp, currentSandbox := NewDefaultTestInterpreter(t)

			// *** Crucial Step: Update cleanupFunc for the current test iteration ***
			// If the test case has a cleanup function defined...
			if tt.cleanupFunc != nil {
				// Store the original function defined in the test case slice.
				originalCleanup := tt.cleanupFunc
				// Replace the test case's cleanupFunc with a *new* closure.
				// This new closure captures the 'currentSandbox' from this specific t.Run call.
				// It also calls the *original* cleanup logic, passing it the correct sandbox.
				tt.cleanupFunc = func(ignoredSandbox string) error {
					// Call the original cleanup function, providing the correct sandbox path.
					// The 'ignoredSandbox' argument comes from the test helper's t.Cleanup,
					// but we use the 'currentSandbox' captured here.
					return originalCleanup(currentSandbox)
				}
			}

			// Now call the test helper. It will use the *updated* tt.cleanupFunc
			// which correctly passes the currentSandbox to the underlying verifyDeletion call.
			testFsToolHelper(t, interp, tt)

			// No explicit cleanup call here, as testFsToolHelper handles t.Cleanup internally.
		})
	}
}
