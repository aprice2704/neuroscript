// NeuroScript Version: 0.3.1
// File version: 0.1.9 // Fix unused variable compiler error in post-test check.
// nlines: 315 // Approximate
// risk_rating: LOW
// filename: pkg/core/tools_fs_dirs_test.go
package core

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

// --- ListDirectory Validation Tests (Unchanged) ---
func TestToolListDirectoryValidation(t *testing.T) {
	testCases := []ValidationTestCase{
		{Name: "Wrong Arg Count (None)", InputArgs: MakeArgs(), ExpectedError: ErrValidationArgCount},
		{Name: "Wrong Arg Count (Three)", InputArgs: MakeArgs("path", true, "extra"), ExpectedError: ErrValidationArgCount},
		{Name: "Nil First Arg", InputArgs: MakeArgs(nil), ExpectedError: ErrValidationRequiredArgNil},
		{Name: "Wrong First Arg Type", InputArgs: MakeArgs(123), ExpectedError: ErrValidationTypeMismatch},
		{Name: "Wrong Second Arg Type", InputArgs: MakeArgs("path", "not-a-bool"), ExpectedError: ErrValidationTypeMismatch},
		{Name: "Correct Args (Path Only)", InputArgs: MakeArgs("some/dir"), ExpectedError: nil},
		{Name: "Correct Args (Path and Recursive)", InputArgs: MakeArgs("some/dir", true), ExpectedError: nil},
		{Name: "Correct Args (Path and Nil Recursive)", InputArgs: MakeArgs("some/dir", nil), ExpectedError: nil},
	}
	runValidationTestCases(t, "FS.List", testCases)
}

// --- ListDirectory Functional Tests (Unchanged) ---
func TestToolListDirectoryFunctional(t *testing.T) {
	// --- Helper for comparing results ---
	compareResults := func(t *testing.T, _ fsTestCase, expected, actual interface{}) {
		t.Helper()
		actualSlice, ok := actual.([]map[string]interface{})
		if !ok {
			t.Fatalf("Actual result is not []map[string]interface{}, got %T", actual)
		}
		expectedSlice, ok := expected.([]map[string]interface{})
		if !ok {
			t.Fatalf("Expected value is not []map[string]interface{}, got %T", expected)
		}

		if len(expectedSlice) != len(actualSlice) {
			t.Fatalf("Expected %d entries, got %d.\nExpected: %+v\nActual:   %+v", len(expectedSlice), len(actualSlice), expectedSlice, actualSlice)
		}

		cleanExpected := make([]map[string]interface{}, len(expectedSlice))
		cleanActual := make([]map[string]interface{}, len(actualSlice))

		copyMapClean := func(src map[string]interface{}) map[string]interface{} {
			dest := make(map[string]interface{})
			isDir := false
			if isDirVal, ok := src["isDir"].(bool); ok {
				isDir = isDirVal
			}
			if name, ok := src["name"].(string); ok {
				dest["name"] = name
			}
			if path, ok := src["path"].(string); ok {
				dest["path"] = path
			}
			dest["isDir"] = isDir
			if sizeVal, ok := src["size"]; ok && !isDir {
				if size, okSize := sizeVal.(int64); okSize {
					dest["size"] = size
				}
			}
			return dest
		}

		for i := range expectedSlice {
			cleanExpected[i] = copyMapClean(expectedSlice[i])
		}
		for i := range actualSlice {
			cleanActual[i] = copyMapClean(actualSlice[i])
		}

		sort.Slice(cleanExpected, func(i, j int) bool {
			pathI, _ := cleanExpected[i]["path"].(string)
			pathJ, _ := cleanExpected[j]["path"].(string)
			return pathI < pathJ
		})
		sort.Slice(cleanActual, func(i, j int) bool {
			pathI, _ := cleanActual[i]["path"].(string)
			pathJ, _ := cleanActual[j]["path"].(string)
			return pathI < pathJ
		})

		if !reflect.DeepEqual(cleanExpected, cleanActual) {
			t.Errorf("Mismatch after cleaning/sorting (modTime ignored, dir size ignored):\nExpected (Clean): %+v\nActual (Clean):   %+v\n---\nExpected (Original): %+v\nActual (Original):   %+v",
				cleanExpected, cleanActual, expectedSlice, actualSlice)
			for i := 0; i < len(cleanExpected); i++ {
				if i >= len(cleanActual) {
					break
				}
				if !reflect.DeepEqual(cleanExpected[i], cleanActual[i]) {
					t.Logf("Detailed Mismatch at index %d:\n Expected: %+v\n Actual:   %+v", i, cleanExpected[i], actualSlice[i]) // Log original actual for detail
				}
			}
		}
	}

	// --- Test Setup Function ---
	setupListTest := func(sandboxRoot string) error {
		testDirPath := filepath.Join(sandboxRoot, "listTest")
		subDirPath := filepath.Join(testDirPath, "sub")
		nestedDirPath := filepath.Join(subDirPath, "nested")
		if err := os.MkdirAll(nestedDirPath, 0755); err != nil && !errors.Is(err, os.ErrExist) {
			return err
		}
		if err := os.WriteFile(filepath.Join(testDirPath, "file1.txt"), []byte("content1"), 0644); err != nil && !errors.Is(err, os.ErrExist) {
			return err
		}
		if err := os.WriteFile(filepath.Join(subDirPath, "file2.txt"), []byte("content22"), 0644); err != nil && !errors.Is(err, os.ErrExist) {
			return err
		}
		if err := os.WriteFile(filepath.Join(nestedDirPath, "file3.txt"), []byte("content333"), 0644); err != nil && !errors.Is(err, os.ErrExist) {
			return err
		}
		if err := os.WriteFile(filepath.Join(sandboxRoot, "file_at_root.txt"), []byte("root"), 0644); err != nil && !errors.Is(err, os.ErrExist) {
			return err
		}
		return nil
	}

	// --- Test Cases ---
	testCases := []fsTestCase{
		{
			name:      "NonRecursive_Root_TestDir",
			toolName:  "FS.List",
			args:      MakeArgs("listTest", false),
			setupFunc: setupListTest,
			wantResult: []map[string]interface{}{
				{"name": "file1.txt", "path": "file1.txt", "isDir": false, "size": int64(8)},
				{"name": "sub", "path": "sub", "isDir": true},
			},
		},
		{
			name:      "NonRecursive_SubDir",
			toolName:  "FS.List",
			args:      MakeArgs("listTest/sub", nil), // Use nil for recursive arg
			setupFunc: setupListTest,
			wantResult: []map[string]interface{}{
				{"name": "file2.txt", "path": "file2.txt", "isDir": false, "size": int64(9)},
				{"name": "nested", "path": "nested", "isDir": true},
			},
		},
		{
			name:      "NonRecursive_SandboxRoot",
			toolName:  "FS.List",
			args:      MakeArgs(".", false),
			setupFunc: setupListTest,
			wantResult: []map[string]interface{}{
				{"name": "file_at_root.txt", "path": "file_at_root.txt", "isDir": false, "size": int64(4)},
				{"name": "listTest", "path": "listTest", "isDir": true},
			},
		},
		{
			name:      "Recursive_Root_TestDir",
			toolName:  "FS.List",
			args:      MakeArgs("listTest", true),
			setupFunc: setupListTest,
			wantResult: []map[string]interface{}{
				{"name": "file1.txt", "path": "file1.txt", "isDir": false, "size": int64(8)},
				{"name": "sub", "path": "sub", "isDir": true},
				{"name": "file2.txt", "path": "sub/file2.txt", "isDir": false, "size": int64(9)},
				{"name": "nested", "path": "sub/nested", "isDir": true},
				{"name": "file3.txt", "path": "sub/nested/file3.txt", "isDir": false, "size": int64(10)},
			},
		},
		{
			name:      "Recursive_SubDir",
			toolName:  "FS.List",
			args:      MakeArgs("listTest/sub", true),
			setupFunc: setupListTest,
			wantResult: []map[string]interface{}{
				{"name": "file2.txt", "path": "file2.txt", "isDir": false, "size": int64(9)},
				{"name": "nested", "path": "nested", "isDir": true},
				{"name": "file3.txt", "path": "nested/file3.txt", "isDir": false, "size": int64(10)},
			},
		},
		{
			name:          "Error_NonExistent",
			toolName:      "FS.List",
			args:          MakeArgs("listTest/nonexistent", false),
			setupFunc:     setupListTest,
			wantResult:    "path not found 'listTest/nonexistent'", // Expect specific message
			wantToolErrIs: ErrFileNotFound,
		},
		{
			name:          "Error_IsFile",
			toolName:      "FS.List",
			args:          MakeArgs("listTest/file1.txt", false),
			setupFunc:     setupListTest,
			wantResult:    "path 'listTest/file1.txt' is not a directory", // Expect specific message
			wantToolErrIs: ErrPathNotDirectory,
		},
		{
			name:          "Error_OutsideSandbox",
			toolName:      "FS.List",
			args:          MakeArgs("../listTest", false),
			setupFunc:     setupListTest,
			wantResult:    "path resolves outside allowed directory", // Expect specific message
			wantToolErrIs: ErrPathViolation,
		},
	}

	// Run tests using the helper
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			interp, currentSandbox := NewDefaultTestInterpreter(t)
			if tc.setupFunc != nil {
				if err := tc.setupFunc(currentSandbox); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}
			testFsToolHelperWithCompare(t, interp, tc, compareResults)
		})
	}
}

// --- Mkdir Validation Tests (Unchanged) ---
func TestToolMkdirValidation(t *testing.T) {
	testCases := []ValidationTestCase{
		{Name: "Wrong Arg Count (None)", InputArgs: MakeArgs(), ExpectedError: ErrValidationArgCount},
		{Name: "Wrong Arg Count (Two)", InputArgs: MakeArgs("path", "extra"), ExpectedError: ErrValidationArgCount},
		{Name: "Nil Arg", InputArgs: MakeArgs(nil), ExpectedError: ErrValidationRequiredArgNil},
		{Name: "Wrong Arg Type", InputArgs: MakeArgs(123), ExpectedError: ErrValidationTypeMismatch},
		{Name: "Correct Args", InputArgs: MakeArgs("new/dir"), ExpectedError: nil},
	}
	runValidationTestCases(t, "FS.Mkdir", testCases)
}

// --- Mkdir Functional Tests ---
func TestToolMkdirFunctional(t *testing.T) {
	setupMkdirTest := func(sandboxRoot string) error {
		// Clean up potential leftovers from previous runs
		os.RemoveAll(filepath.Join(sandboxRoot, "newdir"))
		os.RemoveAll(filepath.Join(sandboxRoot, "parent"))
		os.RemoveAll(filepath.Join(sandboxRoot, "existing_dir"))
		os.Remove(filepath.Join(sandboxRoot, "existing_file"))
		os.RemoveAll(filepath.Join(sandboxRoot, "outside")) // Remove potential successful creation from previous adjusted test

		// Setup initial state
		existingFilePath := filepath.Join(sandboxRoot, "existing_file")
		if err := os.WriteFile(existingFilePath, []byte("hello"), 0644); err != nil {
			return err
		}
		existingDirPath := filepath.Join(sandboxRoot, "existing_dir")
		if err := os.Mkdir(existingDirPath, 0755); err != nil {
			return err
		}
		return nil
	}

	testCases := []fsTestCase{
		{
			name:      "Create New Simple",
			toolName:  "FS.Mkdir",
			args:      MakeArgs("newdir"),
			setupFunc: setupMkdirTest,
			wantResult: map[string]interface{}{
				"status":  "success",
				"message": "Successfully created directory: newdir",
				"path":    "newdir", // Should return relative path
			},
		},
		{
			name:      "Create Nested",
			toolName:  "FS.Mkdir",
			args:      MakeArgs("parent/child"),
			setupFunc: setupMkdirTest,
			wantResult: map[string]interface{}{
				"status":  "success",
				"message": "Successfully created directory: parent/child",
				"path":    "parent/child", // Should return relative path
			},
		},
		{
			name:          "Create Existing Dir",
			toolName:      "FS.Mkdir",
			args:          MakeArgs("existing_dir"),
			setupFunc:     setupMkdirTest,
			wantResult:    "directory 'existing_dir' already exists", // Specific error message
			wantToolErrIs: ErrPathExists,
		},
		{
			name:          "Error_PathIsFile",
			toolName:      "FS.Mkdir",
			args:          MakeArgs("existing_file"),
			setupFunc:     setupMkdirTest,
			wantResult:    "path 'existing_file' already exists and is a file", // Specific error message
			wantToolErrIs: ErrPathNotDirectory,
		},
		{
			name:          "Error_OutsideSandbox_Simple",
			toolName:      "FS.Mkdir",
			args:          MakeArgs("../outside"),
			setupFunc:     setupMkdirTest,
			wantResult:    "path resolves outside allowed directory", // Specific error message
			wantToolErrIs: ErrPathViolation,
		},
		// This path resolves to allowedRoot/outside, which is considered *inside* by ResolveAndSecurePath.
		// Therefore, Mkdir should succeed.
		{
			name:      "Create_Complex_Traversal_Clean_Inside",
			toolName:  "FS.Mkdir",
			args:      MakeArgs("some/dir/../../outside"),
			setupFunc: setupMkdirTest,
			wantResult: map[string]interface{}{
				"status":  "success",
				"message": "Successfully created directory: some/dir/../../outside",
				"path":    "some/dir/../../outside", // Mkdir tool returns the *original* input path here.
			},
			wantToolErrIs: nil, // No error expected now
		},
		{
			name:          "Error_EmptyPath",
			toolName:      "FS.Mkdir",
			args:          MakeArgs(""),
			setupFunc:     setupMkdirTest,
			wantResult:    "path argument cannot be empty", // Specific error message
			wantToolErrIs: ErrInvalidArgument,
		},
		{
			name:          "Error_DotPath",
			toolName:      "FS.Mkdir",
			args:          MakeArgs("."),
			setupFunc:     setupMkdirTest,
			wantResult:    "path '.' is invalid for creating a directory", // Specific error message
			wantToolErrIs: ErrInvalidArgument,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			interp, sandboxRoot := NewDefaultTestInterpreter(t)
			if tc.setupFunc != nil {
				if err := tc.setupFunc(sandboxRoot); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}
			testFsToolHelper(t, interp, tc) // Use the standard helper

			// Post-test verification for successful cases
			if tc.wantToolErrIs == nil {
				// Check if wantResult is a map before proceeding
				if _, ok := tc.wantResult.(map[string]interface{}); ok {
					// Use the *input path* for verification, but resolve it first
					inputPathStr, inputOk := tc.args[0].(string)
					if !inputOk {
						t.Fatalf("Input path arg missing/wrong type for verification.")
					}

					createdPathAbs, errResolve := ResolveAndSecurePath(inputPathStr, sandboxRoot)
					if errResolve != nil {
						t.Errorf("Verification error resolving input path %q: %v", inputPathStr, errResolve)
						return
					}

					info, errStat := os.Stat(createdPathAbs)
					if errStat != nil {
						t.Errorf("Expected directory %q (resolved from %q) to exist after success, but stat failed: %v", createdPathAbs, inputPathStr, errStat)
					} else if !info.IsDir() {
						t.Errorf("Expected %q to be a directory after success, but it's not.", createdPathAbs)
					}

				} else {
					// If wantResult is not a map (e.g., string), this check isn't applicable
					// t.Logf("Skipping post-run verification for non-map result type %T", tc.wantResult)
				}
			} else if tc.wantToolErrIs == ErrPathExists { // Verify existing dir still exists
				inputPathStr, ok := tc.args[0].(string)
				if !ok {
					t.Fatalf("Input path arg missing/wrong type for ErrPathExists verification.")
				}
				existingPathAbs, errResolve := ResolveAndSecurePath(inputPathStr, sandboxRoot)
				if errResolve != nil {
					t.Errorf("Verification error resolving path %q for ErrPathExists: %v", inputPathStr, errResolve)
				} else {
					info, errStat := os.Stat(existingPathAbs)
					if errStat != nil {
						t.Errorf("Expected directory %q to still exist after ErrPathExists, but stat failed: %v", existingPathAbs, errStat)
					} else if !info.IsDir() {
						t.Errorf("Expected %q to still be a directory after ErrPathExists, but it's not.", existingPathAbs)
					}
				}
			}
			// Add other post-run checks if needed for different error types
		})
	}
}
