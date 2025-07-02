// NeuroScript Version: 0.4.0
// File version: 2
// Purpose: Fixed variable shadowing bug by correctly handling the error return from NewDefaultTestInterpreter.
// filename: pkg/core/tools_go_test.go
// nlines: 119
// risk_rating: MEDIUM

package core

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// testGoGetModuleInfoHelper tests the toolGoGetModuleInfo implementation directly.
func testGoGetModuleInfoHelper(t *testing.T, tc struct {
	name       string
	dirArg     interface{}
	setupFunc  func(t *testing.T, sandboxRoot string)
	wantResult map[string]interface{}
	wantErrIs  error
}) {
	t.Helper()
	t.Run(tc.name, func(t *testing.T) {
		// FIX: Correctly handle the (*Interpreter, error) return values.
		interp, err := NewDefaultTestInterpreter(t)
		if err != nil {
			t.Fatalf("NewDefaultTestInterpreter() failed: %v", err)
		}
		// FIX: Get the sandbox directory from the interpreter instance.
		sandboxRoot := interp.SandboxDir()

		toolImpl, _ := interp.ToolRegistry().GetTool("Go.GetModuleInfo")

		// Per-test setup
		if tc.setupFunc != nil {
			tc.setupFunc(t, sandboxRoot)
		}

		args := MakeArgs(tc.dirArg)
		gotResult, toolErr := toolImpl.Func(interp, args)

		if tc.wantErrIs != nil {
			if !errors.Is(toolErr, tc.wantErrIs) {
				t.Errorf("Expected error wrapping [%v], but got: %v", tc.wantErrIs, toolErr)
			}
			return
		}
		if toolErr != nil {
			t.Fatalf("Unexpected error: %v", toolErr)
		}

		// Dynamically update the expected rootDir to match the temporary sandbox directory
		if tc.wantResult != nil {
			tc.wantResult["rootDir"] = sandboxRoot
		}

		if !reflect.DeepEqual(gotResult, tc.wantResult) {
			gotMap, _ := gotResult.(map[string]interface{})
			t.Errorf("Result map mismatch:\nGot:  %#v\nWant: %#v", gotMap, tc.wantResult)
		}
	})
}

func TestToolGoGetModuleInfo(t *testing.T) {
	moduleName := "example.com/modinfo_test"
	goVersion := "1.21"
	goModContent := fmt.Sprintf("module %s\n\ngo %s\n", moduleName, goVersion)

	// setupFunc creates a standard go.mod and a subdir.
	setupFunc := func(t *testing.T, sandboxRoot string) {
		if err := os.WriteFile(filepath.Join(sandboxRoot, "go.mod"), []byte(goModContent), 0644); err != nil {
			t.Fatalf("Setup: Failed to write go.mod: %v", err)
		}
		if err := os.Mkdir(filepath.Join(sandboxRoot, "subdir"), 0755); err != nil {
			t.Fatalf("Setup: Failed to create subdir: %v", err)
		}
	}

	tests := []struct {
		name       string
		dirArg     interface{}
		setupFunc  func(t *testing.T, sandboxRoot string)
		wantResult map[string]interface{}
		wantErrIs  error
	}{
		{
			name:      "From root dir (.)",
			dirArg:    ".",
			setupFunc: setupFunc,
			wantResult: map[string]interface{}{
				"modulePath": moduleName,
				"goVersion":  goVersion,
				"rootDir":    "", // This will be replaced by the helper
				"requires":   []map[string]interface{}{},
				"replaces":   []map[string]interface{}{},
			},
		},
		{
			name:      "From subdir",
			dirArg:    "subdir",
			setupFunc: setupFunc,
			wantResult: map[string]interface{}{
				"modulePath": moduleName,
				"goVersion":  goVersion,
				"rootDir":    "", // This will be replaced by the helper
				"requires":   []map[string]interface{}{},
				"replaces":   []map[string]interface{}{},
			},
		},
		{
			name:      "Directory outside sandbox",
			dirArg:    "../outside",
			setupFunc: setupFunc,
			wantErrIs: ErrPathViolation,
		},
		{
			name: "Go.mod not found",
			setupFunc: func(t *testing.T, sandboxRoot string) {
				// No go.mod is created for this test
			},
			dirArg:     ".",
			wantResult: nil, // Expect nil result and nil error
			wantErrIs:  nil,
		},
		{
			name:      "Wrong arg type",
			dirArg:    123,
			setupFunc: setupFunc,
			wantErrIs: ErrValidationTypeMismatch,
		},
	}

	for _, tt := range tests {
		testGoGetModuleInfoHelper(t, tt)
	}
}
