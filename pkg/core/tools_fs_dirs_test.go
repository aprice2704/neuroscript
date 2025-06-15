// NeuroScript Version: 0.4.0
// File version: 6
// Purpose: Corrected error expectation for Mkdir on an existing file path.
// filename: pkg/core/tools_fs_dirs_test.go
// nlines: 250 // Approximate
// risk_rating: LOW

package core

import (
	"os"
	"path/filepath"
	"testing"
)

// --- ListDirectory Tests ---

func TestToolListDirectoryValidation(t *testing.T) {
	testCases := []ValidationTestCase{
		{
			Name:          "Correct_Args_(Path_Only)",
			InputArgs:     MakeArgs("some/dir"),
			ExpectedError: ErrFileNotFound,
		},
		{
			Name:          "Correct_Args_(Path_and_Recursive)",
			InputArgs:     MakeArgs("some/dir", true),
			ExpectedError: ErrFileNotFound,
		},
		{
			Name:          "Correct_Args_(Path_and_Nil_Recursive)",
			InputArgs:     MakeArgs("some/dir", nil),
			ExpectedError: ErrFileNotFound,
		},
		{
			Name:          "Wrong_Arg_Count",
			InputArgs:     MakeArgs("some/dir", true, "extra"),
			ExpectedError: ErrArgumentMismatch,
		},
		{
			Name:          "Wrong_Path_Type",
			InputArgs:     MakeArgs(123),
			ExpectedError: ErrInvalidArgument,
		},
		{
			Name:          "Wrong_Recursive_Type",
			InputArgs:     MakeArgs(".", "not_a_bool"),
			ExpectedError: ErrInvalidArgument,
		},
	}

	runValidationTestCases(t, "FS.List", testCases)
}

func TestToolListDirectoryFunctional(t *testing.T) {
	// Setup a directory structure for testing. Made idempotent to prevent test pollution.
	setupFunc := func(sandboxRoot string) error {
		// Clean up previous artifacts to ensure idempotency
		os.RemoveAll(filepath.Join(sandboxRoot, "dir1"))
		os.RemoveAll(filepath.Join(sandboxRoot, "empty_dir"))
		os.Remove(filepath.Join(sandboxRoot, "file1.txt"))

		// Dirs
		if err := os.MkdirAll(filepath.Join(sandboxRoot, "dir1", "subdir1"), 0755); err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Join(sandboxRoot, "empty_dir"), 0755); err != nil {
			return err
		}
		// Files
		if err := os.WriteFile(filepath.Join(sandboxRoot, "file1.txt"), []byte("file1"), 0644); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(sandboxRoot, "dir1", "file2.txt"), []byte("file2"), 0644); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(sandboxRoot, "dir1", "subdir1", "file3.txt"), []byte("file3"), 0644); err != nil {
			return err
		}
		return nil
	}

	tests := []fsTestCase{
		{
			name:      "List_Root_NonRecursive",
			toolName:  "FS.List",
			args:      MakeArgs("."),
			setupFunc: setupFunc,
			checkFunc: func(t *testing.T, interp *Interpreter, result interface{}, err error, ctx interface{}) {
				AssertNoError(t, err)
				res, ok := result.([]map[string]interface{})
				if !ok {
					t.Fatalf("Expected []map[string]interface{}, got %T", result)
				}
				if len(res) != 3 { // dir1, empty_dir, file1.txt
					t.Errorf("Expected 3 entries in root, got %d. Result: %v", len(res), res)
				}
			},
		},
		{
			name:      "List_Root_Recursive",
			toolName:  "FS.List",
			args:      MakeArgs(".", true),
			setupFunc: setupFunc,
			checkFunc: func(t *testing.T, interp *Interpreter, result interface{}, err error, ctx interface{}) {
				AssertNoError(t, err)
				res, ok := result.([]map[string]interface{})
				if !ok {
					t.Fatalf("Expected []map[string]interface{}, got %T", result)
				}
				if len(res) != 6 { // file1, dir1, file2, subdir1, file3, empty_dir
					t.Errorf("Expected 6 entries in recursive list, got %d. Result: %v", len(res), res)
				}
			},
		},
		{
			name:          "Error_PathIsFile",
			toolName:      "FS.List",
			args:          MakeArgs("file1.txt"),
			setupFunc:     setupFunc,
			wantToolErrIs: ErrPathNotDirectory,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interp, sb := NewDefaultTestInterpreter(t)
			if tt.setupFunc != nil {
				if err := tt.setupFunc(sb); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}
			testFsToolHelper(t, interp, tt)
		})
	}
}

// --- Mkdir Tests ---

func TestToolMkdirValidation(t *testing.T) {
	testCases := []ValidationTestCase{
		{Name: "Correct_Args", InputArgs: MakeArgs("new/dir"), ExpectedError: nil},
		{Name: "Wrong_Arg_Count", InputArgs: MakeArgs("dir1", "dir2"), ExpectedError: ErrArgumentMismatch},
		{Name: "Wrong_Arg_Type", InputArgs: MakeArgs(123), ExpectedError: ErrInvalidArgument},
		{Name: "Empty_Path", InputArgs: MakeArgs(""), ExpectedError: ErrInvalidArgument},
		{Name: "Current_Dir_Path", InputArgs: MakeArgs("."), ExpectedError: ErrInvalidArgument},
	}
	runValidationTestCases(t, "FS.Mkdir", testCases)
}

func TestToolMkdirFunctional(t *testing.T) {
	// Made idempotent to prevent test pollution.
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
			toolName:  "FS.Mkdir",
			args:      MakeArgs("new_dir_1"),
			setupFunc: setupFunc,
			checkFunc: func(t *testing.T, interp *Interpreter, result interface{}, err error, ctx interface{}) {
				AssertNoError(t, err)
				if _, statErr := os.Stat(filepath.Join(interp.SandboxDir(), "new_dir_1")); os.IsNotExist(statErr) {
					t.Error("Mkdir did not create the directory 'new_dir_1'")
				}
			},
		},
		{
			name:      "Create_Nested_Dir",
			toolName:  "FS.Mkdir",
			args:      MakeArgs(filepath.Join("new_dir_2", "nested")),
			setupFunc: setupFunc,
			checkFunc: func(t *testing.T, interp *Interpreter, result interface{}, err error, ctx interface{}) {
				AssertNoError(t, err)
				if _, statErr := os.Stat(filepath.Join(interp.SandboxDir(), "new_dir_2", "nested")); os.IsNotExist(statErr) {
					t.Error("Mkdir did not create the nested directory")
				}
			},
		},
		{
			name:          "Error_PathIsFile",
			toolName:      "FS.Mkdir",
			args:          MakeArgs("existing_file"),
			setupFunc:     setupFunc,
			wantToolErrIs: ErrPathNotDirectory, // CORRECTED: The tool correctly returns this more specific error.
		},
		{
			name:      "Success_DirExists",
			toolName:  "FS.Mkdir",
			args:      MakeArgs("existing_dir"),
			setupFunc: setupFunc,
			checkFunc: func(t *testing.T, interp *Interpreter, result interface{}, err error, ctx interface{}) {
				AssertNoError(t, err) // This should succeed without error (like mkdir -p)
				if _, statErr := os.Stat(filepath.Join(interp.SandboxDir(), "existing_dir")); os.IsNotExist(statErr) {
					t.Error("Mkdir unexpectedly removed existing directory")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interp, sb := NewDefaultTestInterpreter(t)
			if tt.setupFunc != nil {
				if err := tt.setupFunc(sb); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}
			testFsToolHelper(t, interp, tt)
		})
	}
}
