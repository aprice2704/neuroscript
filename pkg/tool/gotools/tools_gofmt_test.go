// NeuroScript Version: 0.4.0
// File version: 3
// Purpose: Corrected expected errors in Go.Imports tests to match specific implementation errors.
// filename: pkg/tool/gotools/tools_gofmt_test.go
// nlines: 115
// risk_rating: LOW

package gotools

import (
	"errors"
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// testGoFormatToolHelper tests a go formatter tool implementation directly.
func testGoFormatToolHelper(t *testing.T, interp tool.RunTime, tc struct {
	name       string
	toolName   string
	args       []interface{}
	wantResult interface{}
	wantErrIs  error
}) {
	t.Helper()
	t.Run(tc.name, func(t *testing.T) {
		toolImpl, found := interp.ToolRegistry().GetTool(tc.toolName)
		if !found {
			t.Fatalf("Tool %q not found", tc.toolName)
		}

		gotResult, toolErr := toolImpl.Func(interp, tc.args)

		if tc.wantErrIs != nil {
			if !errors.Is(toolErr, tc.wantErrIs) {
				t.Errorf("Expected error wrapping [%v], but got: %v", tc.wantErrIs, toolErr)
			}
			// On error, the tool returns a map. Check the success flag.
			if gotMap, ok := gotResult.(map[string]interface{}); ok {
				if success, _ := gotMap["success"].(bool); success {
					t.Errorf("Expected success=false in result map on error, but was true")
				}
			} else if toolErr != nil {
				// Allow nil map if a toolErr is returned, but log it as a potential issue.
				// t.Logf("Warning: tool returned an error but the result was not a map (%T)", gotResult)
			}
			return
		}

		if toolErr != nil {
			t.Fatalf("Unexpected error: %v", toolErr)
		}

		if !reflect.DeepEqual(gotResult, tc.wantResult) {
			t.Errorf("Result mismatch:\nGot:\n%q\nWant:\n%q", gotResult, tc.wantResult)
		}
	})
}

func TestToolGoFmt(t *testing.T) {
	interp, _ := llm.NewDefaultTestInterpreter(t)
	unformatted := "package main\nfunc  main() {}"
	formatted := "package main\n\nfunc main() {}\n"
	invalid := "package main func main() {"

	tests := []struct {
		name       string
		toolName   string
		args       []interface{}
		wantResult interface{}
		wantErrIs  error
	}{
		{name: "Format valid code", toolName: "Go.Fmt", args: tool.MakeArgs(unformatted), wantResult: formatted},
		{name: "Format already formatted code", toolName: "Go.Fmt", args: tool.MakeArgs(formatted), wantResult: formatted},
		{name: "Format invalid code", toolName: "Go.Fmt", args: tool.MakeArgs(invalid), wantErrIs: lang.ErrToolExecutionFailed},
		{name: "Wrong arg type", toolName: "Go.Fmt", args: tool.MakeArgs(123), wantErrIs: lang.ErrInvalidArgument},
	}

	for _, tt := range tests {
		testGoFormatToolHelper(t, interp, tt)
	}
}

func TestToolGoImports(t *testing.T) {
	interp, _ := llm.NewDefaultTestInterpreter(t)
	needsImport := "package main\nfunc main() { fmt.Println() }"
	wantsImport := "package main\n\nimport \"fmt\"\n\nfunc main() { fmt.Println() }\n"
	hasUnusedImport := "package main\nimport \"os\"\nfunc main() {}"
	wantsUnusedRemoved := "package main\n\nfunc main() {}\n"

	tests := []struct {
		name       string
		toolName   string
		args       []interface{}
		wantResult interface{}
		wantErrIs  error
	}{
		{name: "Add missing import", toolName: "Go.Imports", args: tool.MakeArgs(needsImport), wantResult: wantsImport},
		{name: "Remove unused import", toolName: "Go.Imports", args: tool.MakeArgs(hasUnusedImport), wantResult: wantsUnusedRemoved},
		{name: "Invalid source", toolName: "Go.Imports", args: tool.MakeArgs("package main func {"), wantErrIs: lang.ErrToolExecutionFailed}, // CORRECTED
		{name: "Wrong arg type", toolName: "Go.Imports", args: tool.MakeArgs(12345), wantErrIs: lang.ErrInvalidArgument},                     // CORRECTED
	}

	for _, tt := range tests {
		testGoFormatToolHelper(t, interp, tt)
	}
}
