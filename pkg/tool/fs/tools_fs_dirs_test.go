// NeuroScript Version: 0.4.0
// File version: 8
// Purpose: Corrected Mkdir functional test to expect the more specific ErrPathNotDirectory.
// filename: pkg/tool/fs/tools_fs_dirs_test.go
// nlines: 250 // Approximate
// risk_rating: LOW

package fs

import (
	"os"
	"path/filepath"
	"testing"
)

// --- ListDirectory Tests ---

func TestToolListDirectoryValidation(t *testing.T) {
	testCases := []ValidationTestCase{
		{
			Name:		"Correct_Args_(Path_Only)",
			InputArgs:	MakeArgs("some/dir"),
			ExpectedError:	ErrFileNotFound,
		},
		{
			Name:		"Wrong_Arg_Count",
			InputArgs:	MakeArgs("some/dir", true, "extra"),
			ExpectedError:	ErrArgumentMismatch,
		},
	}

	runValidationTestCases(t, "FS.List", testCases)
}

func TestToolListDirectoryFunctional(t *testing.T) {
	setupFunc := func(sandboxRoot string) error {
		os.RemoveAll(filepath.Join(sandboxRoot, "dir1"))
		os.RemoveAll(filepath.Join(sandboxRoot, "empty_dir"))
		os.Remove(filepath.Join(sandboxRoot, "file1.txt"))
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
			name:		"List_Root_NonRecursive",
			toolName:	"FS.List",
			args:		MakeArgs("."),
			setupFunc:	setupFunc,
			checkFunc: func(t *testing.T, interp *Interpreter, result interface{}, err error, ctx interface{}) {
				AssertNoError(t, err)
				res, ok := result.([]map[string]interface{})
				if !ok {
					t.Fatalf("Expected []map[string]interface{}, got %T", result)
				}
				if len(res) != 3 {	// dir1, empty_dir, file1.txt
					t.Errorf("Expected 3 entries in root, got %d. Result: %v", len(res), res)
				}
			},
		},
		{
			name:		"Error_PathIsFile",
			toolName:	"FS.List",
			args:		MakeArgs("file1.txt"),
			setupFunc:	setupFunc,
			wantToolErrIs:	ErrPathNotDirectory,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interp, err := NewDefaultTestInterpreter(t)
			if err != nil {
				t.Fatalf("NewDefaultTestInterpreter failed: %v", err)
			}
			sb := interp.SandboxDir()

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
		{Name: "Empty_Path", InputArgs: MakeArgs(""), ExpectedError: ErrInvalidArgument},
	}
	runValidationTestCases(t, "FS.Mkdir", testCases)
}

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
			name:		"Create_Single_Dir",
			toolName:	"FS.Mkdir",
			args:		MakeArgs("new_dir_1"),
			setupFunc:	setupFunc,
			checkFunc: func(t *testing.T, interp *Interpreter, result interface{}, err error, ctx interface{}) {
				AssertNoError(t, err)
				if _, statErr := os.Stat(filepath.Join(interp.SandboxDir(), "new_dir_1")); os.IsNotExist(statErr) {
					t.Error("Mkdir did not create the directory 'new_dir_1'")
				}
			},
		},
		{
			name:		"Error_PathIsFile",
			toolName:	"FS.Mkdir",
			args:		MakeArgs("existing_file"),
			setupFunc:	setupFunc,
			// FIX: The error returned is more specific than just 'path exists'. It's specifically not a directory.
			wantToolErrIs:	ErrPathNotDirectory,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interp, err := NewDefaultTestInterpreter(t)
			if err != nil {
				t.Fatalf("NewDefaultTestInterpreter failed: %v", err)
			}
			sb := interp.SandboxDir()

			if tt.setupFunc != nil {
				if err := tt.setupFunc(sb); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}
			testFsToolHelper(t, interp, tt)
		})
	}
}