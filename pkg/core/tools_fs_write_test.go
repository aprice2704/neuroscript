// NeuroScript Version: 0.4.0
// File version: 3
// Purpose: Refactored tests to use the standard runValidationTestCases helper, fixing 'Tool not found' errors for FS.Append.
// filename: pkg/core/tools_fs_write_test.go
// nlines: 95
// risk_rating: LOW

package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestToolWriteFileValidation(t *testing.T) {
	// --- FS.Write Validation ---
	writeValidationCases := []ValidationTestCase{
		{Name: "Write - Correct args", InputArgs: MakeArgs("file.txt", "content"), ExpectedError: nil},
		{Name: "Write - Wrong content type", InputArgs: MakeArgs("file.txt", 123), ExpectedError: ErrInvalidArgument},
		{Name: "Write - Wrong arg count (too few)", InputArgs: MakeArgs("file.txt"), ExpectedError: ErrArgumentMismatch},
		{Name: "Write - Path outside sandbox", InputArgs: MakeArgs("../bad.txt", "content"), ExpectedError: ErrPathViolation},
	}
	runValidationTestCases(t, "FS.Write", writeValidationCases)

	// --- FS.Append Validation ---
	appendValidationCases := []ValidationTestCase{
		{Name: "Append - Correct args", InputArgs: MakeArgs("file.txt", "content"), ExpectedError: nil},
		{Name: "Append - Wrong arg count", InputArgs: MakeArgs("file.txt"), ExpectedError: ErrArgumentMismatch},
		{Name: "Append - Path outside sandbox", InputArgs: MakeArgs("../bad.txt", "content"), ExpectedError: ErrPathViolation},
	}
	runValidationTestCases(t, "FS.Append", appendValidationCases)
}

func TestToolWriteFileFunctional(t *testing.T) {
	setup := func(sandboxRoot string) error {
		os.Remove(filepath.Join(sandboxRoot, "newfile.txt"))
		os.Remove(filepath.Join(sandboxRoot, "existing.txt"))
		os.Remove(filepath.Join(sandboxRoot, "append.txt"))
		os.Remove(filepath.Join(sandboxRoot, "newappend.txt"))
		return nil
	}

	setupAppend := func(sandboxRoot string) error {
		setup(sandboxRoot)
		return os.WriteFile(filepath.Join(sandboxRoot, "append.txt"), []byte("initial."), 0644)
	}

	testCases := []fsTestCase{
		// FS.Write
		{name: "Write to new file", toolName: "FS.Write", args: MakeArgs("newfile.txt", "hello world"), setupFunc: setup, wantContent: "hello world"},
		{name: "Overwrite existing file", toolName: "FS.Write", args: MakeArgs("existing.txt", "new content"), setupFunc: func(s string) error {
			setup(s)
			return os.WriteFile(filepath.Join(s, "existing.txt"), []byte("old content"), 0644)
		}, wantContent: "new content"},

		// FS.Append
		{name: "Append to existing file", toolName: "FS.Append", args: MakeArgs("append.txt", "appended."), setupFunc: setupAppend, wantContent: "initial.appended."},
		{name: "Append to non-existent file", toolName: "FS.Append", args: MakeArgs("newappend.txt", "content"), setupFunc: setup, wantContent: "content"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			interp, sb := NewDefaultTestInterpreter(t)

			if tc.setupFunc != nil {
				if err := tc.setupFunc(sb); err != nil {
					t.Fatalf("setup failed: %v", err)
				}
			}

			tool, ok := interp.ToolRegistry().GetTool(tc.toolName)
			if !ok {
				t.Fatalf("Tool '%s' not found in registry", tc.toolName)
			}

			_, err := tool.Func(interp, tc.args)
			if err != nil {
				t.Fatalf("unexpected error during tool execution: %v", err)
			}

			filePath := tc.args[0].(string)
			absPath := filepath.Join(sb, filePath)
			content, readErr := os.ReadFile(absPath)
			if readErr != nil {
				t.Fatalf("failed to read file for verification: %v", readErr)
			}
			if string(content) != tc.wantContent {
				t.Errorf("content mismatch:\ngot:  %q\nwant: %q", string(content), tc.wantContent)
			}
		})
	}
}
