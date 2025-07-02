// NeuroScript Version: 0.4.0
// File version: 4
// Purpose: Fixed variable shadowing bug by correctly handling the error return from NewDefaultTestInterpreter.
// filename: pkg/tool/fs/tools_fs_write_test.go
// nlines: 95
// risk_rating: LOW

package fs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

func TestToolWriteFileValidation(t *testing.T) {
	writeValidationCases := []testutil.ValidationTestCase{
		{Name: "Write - Correct args", InputArgs: tool.MakeArgs("file.txt", "content"), ExpectedError: nil},
		{Name: "Write - Path outside sandbox", InputArgs: tool.MakeArgs("../bad.txt", "content"), ExpectedError: lang.ErrPathViolation},
	}
	testutil.runValidationTestCases(t, "FS.Write", writeValidationCases)

	appendValidationCases := []testutil.ValidationTestCase{
		{Name: "Append - Correct args", InputArgs: tool.MakeArgs("file.txt", "content"), ExpectedError: nil},
		{Name: "Append - Path outside sandbox", InputArgs: tool.MakeArgs("../bad.txt", "content"), ExpectedError: lang.ErrPathViolation},
	}
	testutil.runValidationTestCases(t, "FS.Append", appendValidationCases)
}

func TestToolWriteFileFunctional(t *testing.T) {
	setup := func(sandboxRoot string) error {
		os.Remove(filepath.Join(sandboxRoot, "newfile.txt"))
		return nil
	}

	testCases := []fsTestCase{
		{name: "Write to new file", toolName: "FS.Write", args: tool.MakeArgs("newfile.txt", "hello world"), setupFunc: setup, wantContent: "hello world"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			interp, err := llm.NewDefaultTestInterpreter(t)
			if err != nil {
				t.Fatalf("NewDefaultTestInterpreter failed: %v", err)
			}
			sb := interp.SandboxDir()

			if tc.setupFunc != nil {
				if err := tc.setupFunc(sb); err != nil {
					t.Fatalf("setup failed: %v", err)
				}
			}

			tool, ok := interp.ToolRegistry().GetTool(tc.toolName)
			if !ok {
				t.Fatalf("Tool '%s' not found in registry", tc.toolName)
			}

			_, toolErr := tool.Func(interp, tc.args)
			if toolErr != nil {
				t.Fatalf("unexpected error during tool execution: %v", toolErr)
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