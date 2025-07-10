// NeuroScript Version: 0.4.0
// File version: 10
// Purpose: Corrected tool names to align with the new registration system.
// filename: pkg/tool/fs/tools_fs_dirs_test.go
// nlines: 250 // Approximate
// risk_rating: LOW

package fs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// --- ListDirectory Tests ---

func TestToolListDirectoryFunctional(t *testing.T) {
	setupFunc := func(sandboxRoot string) error {
		// Clean up previous test runs if any
		os.RemoveAll(filepath.Join(sandboxRoot, "dir1"))
		os.RemoveAll(filepath.Join(sandboxRoot, "empty_dir"))
		os.Remove(filepath.Join(sandboxRoot, "file1.txt"))
		// Setup for current test
		if err := os.MkdirAll(filepath.Join(sandboxRoot, "dir1", "subdir1"), 0755); err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Join(sandboxRoot, "empty_dir"), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(sandboxRoot, "file1.txt"), []byte("file1"), 0644); err != nil {
			return err
		}
		return nil
	}

	tests := []fsTestCase{
		{
			name:      "List_Root_NonRecursive",
			toolName:  "List",
			args:      []interface{}{"."},
			setupFunc: setupFunc,
			checkFunc: func(t *testing.T, interp tool.Runtime, result interface{}, err error, ctx interface{}) {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				res, ok := result.([]map[string]interface{})
				if !ok {
					t.Fatalf("Expected []map[string]interface{}, got %T", result)
				}
				// Expecting dir1, empty_dir, file1.txt
				if len(res) != 3 {
					t.Errorf("Expected 3 entries in root, got %d. Result: %v", len(res), res)
				}
			},
		},
		{
			name:          "Error_PathIsFile",
			toolName:      "List",
			args:          []interface{}{"file1.txt"},
			setupFunc:     setupFunc,
			wantToolErrIs: lang.ErrPathNotDirectory,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interp := newFsTestInterpreter(t)
			testFsToolHelper(t, interp, tt)
		})
	}
}

// --- Mkdir Tests ---

func TestToolMkdirFunctional(t *testing.T) {
	setupFunc := func(sandboxRoot string) error {
		os.Remove(filepath.Join(sandboxRoot, "existing_file"))
		os.RemoveAll(filepath.Join(sandboxRoot, "existing_dir"))

		if err := os.WriteFile(filepath.Join(sandboxRoot, "existing_file"), []byte(""), 0644); err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Join(sandboxRoot, "existing_dir"), 0755); err != nil {
			return err
		}
		return nil
	}

	tests := []fsTestCase{
		{
			name:      "Create_Single_Dir",
			toolName:  "Mkdir",
			args:      []interface{}{"new_dir_1"},
			setupFunc: setupFunc,
			checkFunc: func(t *testing.T, interp tool.Runtime, result interface{}, err error, ctx interface{}) {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				// FIX: Cast interp to the concrete *interpreter.Interpreter type to access SandboxDir.
				interpImpl, ok := interp.(*interpreter.Interpreter)
				if !ok {
					t.Fatal("Interpreter provided to checkFunc is not the expected concrete type.")
				}
				if _, statErr := os.Stat(filepath.Join(interpImpl.SandboxDir(), "new_dir_1")); os.IsNotExist(statErr) {
					t.Error("Mkdir did not create the directory 'new_dir_1'")
				}
			},
		},
		{
			name:          "Error_PathIsFile",
			toolName:      "Mkdir",
			args:          []interface{}{"existing_file"},
			setupFunc:     setupFunc,
			wantToolErrIs: lang.ErrPathNotDirectory,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interp := newFsTestInterpreter(t)
			testFsToolHelper(t, interp, tt)
		})
	}
}
