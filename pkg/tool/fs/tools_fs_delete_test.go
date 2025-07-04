// NeuroScript Version: 0.3.1
// File version: 0.1.0 // Refactored to be self-contained and use local helpers.
// nlines: 145 // Approximate
// risk_rating: MEDIUM // Test file for a destructive operation
// filename: pkg/tool/fs/tools_fs_delete_test.go
package fs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

func TestToolDeleteFile(t *testing.T) {
	// --- Test Setup Data (relative paths) ---
	fileToDeleteRel := "deleteMe.txt"
	dirToDeleteRel := "deleteMeDir"
	nonEmptyDirRel := "dontDeleteMeDir"
	nonEmptyFileRel := filepath.Join(nonEmptyDirRel, "keepMe.txt")
	fileToDeleteContent := "some content"

	// --- Test Cases ---
	tests := []fsTestCase{
		{
			name:     "Delete Existing File",
			toolName: "FS.Delete",
			args:     []interface{}{fileToDeleteRel},
			setupFunc: func(s string) error {
				mustWriteFile(t, filepath.Join(s, fileToDeleteRel), fileToDeleteContent)
				return nil
			},
			checkFunc: func(t *testing.T, interp tool.Runtime, result interface{}, err error, setupCtx interface{}) {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if result != "OK" {
					t.Errorf("Expected result 'OK', got %v", result)
				}
				if _, statErr := os.Stat(filepath.Join(interp.SandboxDir(), fileToDeleteRel)); !os.IsNotExist(statErr) {
					t.Error("Expected file to be deleted, but it still exists.")
				}
			},
		},
		{
			name:     "Delete Empty Directory",
			toolName: "FS.Delete",
			args:     []interface{}{dirToDeleteRel},
			setupFunc: func(s string) error {
				mustMkdir(t, filepath.Join(s, dirToDeleteRel))
				return nil
			},
			checkFunc: func(t *testing.T, interp tool.Runtime, result interface{}, err error, setupCtx interface{}) {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if _, statErr := os.Stat(filepath.Join(interp.SandboxDir(), dirToDeleteRel)); !os.IsNotExist(statErr) {
					t.Error("Expected directory to be deleted, but it still exists.")
				}
			},
		},
		{
			name:          "Delete Non-Existent File",
			toolName:      "FS.Delete",
			args:          []interface{}{"noSuchFile.txt"},
			wantResult:    "OK",
			wantToolErrIs: nil,
		},
		{
			name:     "Delete Non-Empty Directory",
			toolName: "FS.Delete",
			args:     []interface{}{nonEmptyDirRel},
			setupFunc: func(s string) error {
				mustMkdir(t, filepath.Join(s, nonEmptyDirRel))
				mustWriteFile(t, filepath.Join(s, nonEmptyFileRel), "content")
				return nil
			},
			wantToolErrIs: lang.ErrCannotDelete,
		},
		{
			name:          "Path_Outside_Sandbox",
			toolName:      "FS.Delete",
			args:          []interface{}{"../someFile"},
			wantToolErrIs: lang.ErrPathViolation,
		},
		{
			name:          "Validation_Missing_Arg",
			toolName:      "FS.Delete",
			args:          []interface{}{},
			wantToolErrIs: lang.ErrArgumentMismatch,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// Use the local helper from tools_fs_helpers_test.go
			interp := newFsTestInterpreter(t)
			testFsToolHelper(t, interp, tt)
		})
	}
}
