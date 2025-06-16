// NeuroScript Version: 0.4.1
// File version: 2
// Purpose: Fixed variable shadowing bug by correctly handling the error return from NewDefaultTestInterpreter.
// filename: pkg/core/tools_fs_read_test.go
// nlines: 65
// risk_rating: LOW

package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestToolReadFile(t *testing.T) {
	readTestFile := "readTest.txt"
	readTestContent := "Hello Reader"

	setup := func(sandboxRoot string) error {
		return os.WriteFile(filepath.Join(sandboxRoot, readTestFile), []byte(readTestContent), 0644)
	}

	tests := []fsTestCase{
		{name: "Read Existing File", toolName: "FS.Read", args: MakeArgs(readTestFile), setupFunc: setup, wantResult: readTestContent},
		{name: "Read Non-Existent File", toolName: "FS.Read", args: MakeArgs("nonexistent.txt"), wantToolErrIs: ErrFileNotFound},
		{name: "Validation_Empty_Filepath_Arg", toolName: "FS.Read", args: MakeArgs(""), wantToolErrIs: ErrInvalidArgument},
		{name: "Path_Outside_Sandbox", toolName: "FS.Read", args: MakeArgs("../outside.txt"), wantToolErrIs: ErrPathViolation},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interp, err := NewDefaultTestInterpreter(t)
			if err != nil {
				t.Fatalf("NewDefaultTestInterpreter failed: %v", err)
			}
			currentSandbox := interp.SandboxDir()

			if tt.setupFunc != nil {
				if err := tt.setupFunc(currentSandbox); err != nil {
					t.Fatalf("Setup function failed: %v", err)
				}
			}
			testFsToolHelper(t, interp, tt)
		})
	}
}
