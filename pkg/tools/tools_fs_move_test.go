// NeuroScript Version: 0.4.0
// File version: 2
// Purpose: Moved 'Correct_Args' test to functional tests with proper setup to fix validation failure.
// nlines: 180
// risk_rating: LOW
// filename: pkg/core/tools_fs_move_test.go
package core

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// --- MoveFile Validation Tests ---
func TestToolMoveFileValidation(t *testing.T) {
	testCases := []ValidationTestCase{
		{Name: "Wrong_Arg_Count_(None)", InputArgs: MakeArgs(), ExpectedError: ErrArgumentMismatch},
		{Name: "Wrong_Arg_Count_(One)", InputArgs: MakeArgs("src"), ExpectedError: ErrArgumentMismatch},
		{Name: "Wrong_Arg_Count_(Three)", InputArgs: MakeArgs("src", "dest", "extra"), ExpectedError: ErrArgumentMismatch},
		{Name: "Nil_First_Arg", InputArgs: MakeArgs(nil, "dest"), ExpectedError: ErrInvalidArgument},
		{Name: "Nil_Second_Arg", InputArgs: MakeArgs("src", nil), ExpectedError: ErrInvalidArgument},
		{Name: "Wrong_First_Arg_Type", InputArgs: MakeArgs(123, "dest"), ExpectedError: ErrInvalidArgument},
		{Name: "Wrong_Second_Arg_Type", InputArgs: MakeArgs("src", 456), ExpectedError: ErrInvalidArgument},
		// The "Correct_Args" case was moved to functional tests because it requires file system state.
	}
	runValidationTestCases(t, "FS.Move", testCases)
}

// --- MoveFile Functional Tests ---
func TestToolMoveFileFunctional(t *testing.T) {
	// Use t.TempDir for sandboxed filesystem operations
	sandboxDir := t.TempDir()
	interp := NewTestInterpreterWithSandbox(t, sandboxDir)

	// --- Test Setup Helper ---
	createTestFile := func(relativePath, content string) string {
		t.Helper()
		absPath := filepath.Join(sandboxDir, relativePath)
		parentDir := filepath.Dir(absPath)
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			t.Fatalf("Failed to create parent directory %s for test file: %v", parentDir, err)
		}
		err := os.WriteFile(absPath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", absPath, err)
		}
		return relativePath
	}

	// --- Test Cases ---
	testCases := []struct {
		name       string
		sourcePath string
		destPath   string
		setupFunc  func()
		wantErrIs  error
		checkFunc  func(t *testing.T)
	}{
		{
			name: "Success: Correct Args (from validation)",
			setupFunc: func() {
				createTestFile("source.txt", "content")
			},
			sourcePath: "source.txt",
			destPath:   "destination.txt",
			wantErrIs:  nil,
			checkFunc: func(t *testing.T) {
				if _, err := os.Stat(filepath.Join(sandboxDir, "source.txt")); !errors.Is(err, os.ErrNotExist) {
					t.Errorf("Source file 'source.txt' should not exist after move")
				}
				if _, err := os.Stat(filepath.Join(sandboxDir, "destination.txt")); err != nil {
					t.Errorf("Destination file 'destination.txt' should exist after move: %v", err)
				}
			},
		},
		{
			name:       "Success: Rename file",
			sourcePath: createTestFile("old.txt", "content1"),
			destPath:   "new.txt",
			checkFunc: func(t *testing.T) {
				if _, err := os.Stat(filepath.Join(sandboxDir, "old.txt")); !errors.Is(err, os.ErrNotExist) {
					t.Errorf("Source file old.txt still exists after successful move")
				}
				if _, err := os.Stat(filepath.Join(sandboxDir, "new.txt")); err != nil {
					t.Errorf("Destination file new.txt not found after successful move: %v", err)
				}
			},
		},
		{
			name: "Success: Move file into existing subdir",
			setupFunc: func() {
				createTestFile("move_me.txt", "content2")
				os.Mkdir(filepath.Join(sandboxDir, "subdir"), 0755)
			},
			sourcePath: "move_me.txt",
			destPath:   "subdir/moved.txt",
			checkFunc: func(t *testing.T) {
				if _, err := os.Stat(filepath.Join(sandboxDir, "move_me.txt")); !errors.Is(err, os.ErrNotExist) {
					t.Errorf("Source file move_me.txt still exists after successful move")
				}
				if _, err := os.Stat(filepath.Join(sandboxDir, "subdir/moved.txt")); err != nil {
					t.Errorf("Destination file subdir/moved.txt not found after successful move: %v", err)
				}
			},
		},
		{
			name:       "Fail: Source does not exist",
			sourcePath: "nonexistent_source.txt",
			destPath:   "any_dest.txt",
			wantErrIs:  ErrFileNotFound,
			checkFunc: func(t *testing.T) {
				if _, err := os.Stat(filepath.Join(sandboxDir, "any_dest.txt")); !errors.Is(err, os.ErrNotExist) {
					t.Errorf("Destination file should not exist when source is missing")
				}
			},
		},
		{
			name: "Fail: Destination exists",
			setupFunc: func() {
				createTestFile("src_exists.txt", "content3")
				createTestFile("dest_exists.txt", "content4")
			},
			sourcePath: "src_exists.txt",
			destPath:   "dest_exists.txt",
			wantErrIs:  ErrPathExists,
			checkFunc: func(t *testing.T) {
				if _, err := os.Stat(filepath.Join(sandboxDir, "src_exists.txt")); err != nil {
					t.Errorf("Source file should still exist when destination exists")
				}
			},
		},
		{
			name:       "Fail: Path outside sandbox (Source)",
			sourcePath: "../outside_src.txt",
			destPath:   "dest.txt",
			wantErrIs:  ErrPathViolation,
		},
		{
			name:       "Fail: Path outside sandbox (Destination)",
			sourcePath: createTestFile("valid_src.txt", "content5"),
			destPath:   "../outside_dest.txt",
			wantErrIs:  ErrPathViolation,
			checkFunc: func(t *testing.T) {
				if _, err := os.Stat(filepath.Join(sandboxDir, "valid_src.txt")); err != nil {
					t.Errorf("Source file should still exist when destination is invalid")
				}
			},
		},
		{
			name:       "Fail: Empty Source Path",
			sourcePath: "",
			destPath:   "some_dest.txt",
			wantErrIs:  ErrInvalidArgument,
		},
		{
			name:       "Fail: Empty Destination Path",
			sourcePath: createTestFile("another_valid_src.txt", "content6"),
			destPath:   "",
			wantErrIs:  ErrInvalidArgument,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupFunc != nil {
				tc.setupFunc()
			}
			toolImpl, _ := interp.ToolRegistry().GetTool("FS.Move")
			_, err := toolImpl.Func(interp, MakeArgs(tc.sourcePath, tc.destPath))

			if !errors.Is(err, tc.wantErrIs) {
				t.Errorf("Expected error [%v], but got [%v]", tc.wantErrIs, err)
			}

			if tc.checkFunc != nil {
				tc.checkFunc(t)
			}
		})
	}
}
