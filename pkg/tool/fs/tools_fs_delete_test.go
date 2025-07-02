// NeuroScript Version: 0.3.1
// File version: 0.0.5 // Correct t.Errorf usage from %w to %v
// nlines: 145 // Approximate
// risk_rating: MEDIUM // Test file for a destructive operation
// filename: pkg/tool/fs/tools_fs_delete_test.go
package fs

import (
	"errors"	// Keep errors
	"fmt"		// Keep fmt
	"os"		// Keep os
	"path/filepath"	// Keep filepath
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

func TestToolDeleteFile(t *testing.T) {
	// --- Test Setup Data (relative paths) ---
	fileToDeleteRel := "deleteMe.txt"
	dirToDeleteRel := "deleteMeDir"	// Should be empty
	nonEmptyDirRel := "dontDeleteMeDir"
	nonEmptyFileRel := filepath.Join(nonEmptyDirRel, "keepMe.txt")
	fileToDeleteContent := "some content"

	// --- Setup Function (runs in the specific sandbox for each test) ---
	setupDeleteFileTest := func(sandboxRoot string) error {
		fileToDeleteAbs := filepath.Join(sandboxRoot, fileToDeleteRel)
		dirToDeleteAbs := filepath.Join(sandboxRoot, dirToDeleteRel)
		nonEmptyDirAbs := filepath.Join(sandboxRoot, nonEmptyDirRel)
		nonEmptyFileAbs := filepath.Join(sandboxRoot, nonEmptyFileRel)

		// Clean up before setup
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

	// --- Verification Helper ---
	verifyDeletion := func(t *testing.T, sandboxRoot string, pathRel string, shouldExist bool) {
		t.Helper()
		if sandboxRoot == "" {
			t.Fatalf("verifyDeletion called with empty sandboxRoot for path %s", pathRel)
		}
		pathAbs := filepath.Join(sandboxRoot, pathRel)
		_, err := os.Stat(pathAbs)
		if shouldExist {
			if err != nil {
				// Corrected: Use %v for printing errors in t.Errorf
				t.Errorf("verify failed: expected '%s' (abs: %s) to exist, but got error: %v", pathRel, pathAbs, err)
			}
		} else {	// Should NOT exist
			if err == nil {
				t.Errorf("verify failed: expected '%s' (abs: %s) to be deleted, but it still exists", pathRel, pathAbs)
			} else if !errors.Is(err, os.ErrNotExist) {
				// Corrected: Use %v for printing errors in t.Errorf
				t.Errorf("verify failed: expected '%s' (abs: %s) to not exist, but got unexpected error: %v", pathRel, pathAbs, err)
			}
		}
	}

	// --- Test Cases ---
	tests := []fsTestCase{
		{
			name:		"Delete Existing File",
			toolName:	"FS.Delete",
			args:		tool.MakeArgs(fileToDeleteRel),
			setupFunc:	setupDeleteFileTest,
			checkFunc: func(t *testing.T, interp *neurogo.Interpreter, result interface{}, err error, setupCtx interface{}) {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if result != "OK" {
					t.Errorf("Expected result 'OK', got %v", result)
				}
				verifyDeletion(t, interp.SandboxDir(), fileToDeleteRel, false)
			},
		},
		{
			name:		"Delete Empty Directory",
			toolName:	"FS.Delete",
			args:		tool.MakeArgs(dirToDeleteRel),
			setupFunc:	setupDeleteFileTest,
			checkFunc: func(t *testing.T, interp *neurogo.Interpreter, result interface{}, err error, setupCtx interface{}) {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if result != "OK" {
					t.Errorf("Expected result 'OK', got %v", result)
				}
				verifyDeletion(t, interp.SandboxDir(), dirToDeleteRel, false)
			},
		},
		{
			name:		"Delete Non-Existent File",
			toolName:	"FS.Delete",
			args:		tool.MakeArgs("noSuchFile.txt"),
			setupFunc:	setupDeleteFileTest,
			wantResult:	"OK",
		},
		{
			name:		"Delete Non-Empty Directory",
			toolName:	"FS.Delete",
			args:		tool.MakeArgs(nonEmptyDirRel),
			setupFunc:	setupDeleteFileTest,
			wantToolErrIs:	lang.ErrCannotDelete,
			checkFunc: func(t *testing.T, interp *neurogo.Interpreter, result interface{}, err error, setupCtx interface{}) {
				if !errors.Is(err, lang.ErrCannotDelete) {
					t.Fatalf("Expected error ErrCannotDelete, got %v", err)
				}
				verifyDeletion(t, interp.SandboxDir(), nonEmptyDirRel, true)	// Verify it still exists
			},
		},
		{
			name:		"Validation_Wrong_Arg_Type",
			toolName:	"FS.Delete",
			args:		tool.MakeArgs(12345),
			setupFunc:	setupDeleteFileTest,
			wantToolErrIs:	lang.ErrInvalidArgument,
		},
		{
			name:		"Path_Outside_Sandbox",
			toolName:	"FS.Delete",
			args:		tool.MakeArgs("../someFile"),
			setupFunc:	setupDeleteFileTest,
			wantToolErrIs:	lang.ErrPathViolation,
		},
		{
			name:		"Validation_Missing_Arg",
			toolName:	"FS.Delete",
			args:		tool.MakeArgs(),
			setupFunc:	setupDeleteFileTest,
			wantToolErrIs:	lang.ErrArgumentMismatch,
		},
		{
			name:		"Validation_Nil_Arg",
			toolName:	"FS.Delete",
			args:		tool.MakeArgs(nil),
			setupFunc:	setupDeleteFileTest,
			wantToolErrIs:	lang.ErrInvalidArgument,
		},
		{
			name:		"Validation_Empty_String_Arg",
			toolName:	"FS.Delete",
			args:		tool.MakeArgs(""),
			setupFunc:	setupDeleteFileTest,
			wantToolErrIs:	lang.ErrInvalidArgument,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			interp, _ := llm.NewDefaultTestInterpreter(t)
			testFsToolHelper(t, interp, tt)
		})
	}
}