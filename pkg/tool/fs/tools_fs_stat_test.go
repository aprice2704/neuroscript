// NeuroScript Version: 0.4.1
// File version: 1
// Purpose: Refactored to use the central fs test helper.
// filename: pkg/tool/fs/tools_fs_stat_test.go
// nlines: 88
// risk_rating: LOW

package fs

import (
	"os"
	"path/filepath"
	"testing"
)

// compareStatResults is a custom check function for stat results.
func compareStatResults(t *testing.T, expectedResult map[string]interface{}, actual interface{}) {
	t.Helper()
	actualMap, okA := actual.(map[string]interface{})
	if !okA {
		t.Fatalf("Actual result is not map[string]interface{}, got %T", actual)
	}

	// Compare specific fields, ignoring dynamic ones like timestamps
	if actualMap["is_dir"] != expectedResult["is_dir"] {
		t.Errorf("is_dir mismatch: got %v, want %v", actualMap["is_dir"], expectedResult["is_dir"])
	}
	if actualMap["path"] != expectedResult["path"] {
		t.Errorf("path mismatch: got %v, want %v", actualMap["path"], expectedResult["path"])
	}
}

func TestToolStat(t *testing.T) {
	testFileName := "test_file.txt"
	testFileContent := "hello"
	testDirName := "test_subdir"

	setup := func(sandboxRoot string) error {
		if err := os.WriteFile(filepath.Join(sandboxRoot, testFileName), []byte(testFileContent), 0644); err != nil {
			return err
		}
		return os.Mkdir(filepath.Join(sandboxRoot, testDirName), 0755)
	}

	testCases := []fsTestCase{
		{
			name:		"Stat Existing File",
			toolName:	"FS.Stat",
			args:		MakeArgs(testFileName),
			setupFunc:	setup,
			checkFunc: func(t *testing.T, interp *Interpreter, result interface{}, err error, setupCtx interface{}) {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				compareStatResults(t, map[string]interface{}{"is_dir": false, "path": testFileName}, result)
			},
		},
		{
			name:		"Stat Existing Directory",
			toolName:	"FS.Stat",
			args:		MakeArgs(testDirName),
			setupFunc:	setup,
			checkFunc: func(t *testing.T, interp *Interpreter, result interface{}, err error, setupCtx interface{}) {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				compareStatResults(t, map[string]interface{}{"is_dir": true, "path": testDirName}, result)
			},
		},
		{name: "Stat Non-existent File", toolName: "FS.Stat", args: MakeArgs("nonexistent.txt"), setupFunc: setup, wantToolErrIs: ErrFileNotFound},
		{name: "Stat Path Outside Sandbox", toolName: "FS.Stat", args: MakeArgs("../outside.txt"), setupFunc: setup, wantToolErrIs: ErrPathViolation},
		{name: "Stat Empty Path", toolName: "FS.Stat", args: MakeArgs(""), setupFunc: setup, wantToolErrIs: ErrInvalidArgument},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			interp, _ := NewDefaultTestInterpreter(t)
			testFsToolHelper(t, interp, tc)
		})
	}
}