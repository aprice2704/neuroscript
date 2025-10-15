// NeuroScript Version: 0.5.2
// File version: 6
// Purpose: Contains tests for the 'Inspect' string formatting tool.
// filename: pkg/tool/strtools/tools_string_format_test.go
// nlines: 99
// risk_rating: LOW

package strtools

import (
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/types"
)

func TestToolInspect(t *testing.T) {
	t.Logf("DEBUG: Creating new test interpreter for format tests with a valid HostContext.")
	hostCtx := &interpreter.HostContext{
		Logger: logging.NewTestLogger(t),
		Stdout: os.Stdout,
		Stdin:  os.Stdin,
		Stderr: os.Stderr,
	}
	interp := interpreter.NewInterpreter(
		interpreter.WithHostContext(hostCtx),
		interpreter.WithExecPolicy(policy.AllowAll()),
	)

	for _, impl := range stringFormatToolsToRegister {
		if _, err := interp.ToolRegistry().RegisterTool(impl); err != nil {
			t.Fatalf("Failed to register tool %q: %v", impl.Spec.Name, err)
		}
	}

	nestedMap := map[string]interface{}{
		"level1": map[string]interface{}{
			"level2": "level2_value",
		},
	}

	longString := "this is a very long string that should be truncated by the inspect tool to show its functionality"

	tests := []struct {
		name       string
		toolName   string
		args       []interface{}
		wantResult string
		wantErrIs  error
	}{
		{name: "Simple String", toolName: "Inspect", args: MakeArgs("hello"), wantResult: `"hello"`},
		{name: "Integer", toolName: "Inspect", args: MakeArgs(123), wantResult: `123`},
		{name: "Float", toolName: "Inspect", args: MakeArgs(123.45), wantResult: `123.45`},
		{name: "Boolean", toolName: "Inspect", args: MakeArgs(true), wantResult: `true`},
		{name: "Nil", toolName: "Inspect", args: MakeArgs(nil), wantResult: `<nil>`},
		{name: "Simple Slice", toolName: "Inspect", args: MakeArgs([]interface{}{"a", int64(1)}), wantResult: `["a", 1]`},
		{name: "Simple Map", toolName: "Inspect", args: MakeArgs(map[string]interface{}{"key": "value"}), wantResult: `{"key":"value"}`},
		{
			name:       "Long String Truncation",
			toolName:   "Inspect",
			args:       MakeArgs(longString, int64(32)),
			wantResult: `"this is a very long string th..."`, // Corrected expectation
		},
		{
			name:       "Depth Limit",
			toolName:   "Inspect",
			args:       MakeArgs(nestedMap, int64(128), int64(1)),
			wantResult: `{"level1":...}`,
		},
		{
			name:       "Sufficient Depth",
			toolName:   "Inspect",
			args:       MakeArgs(nestedMap, int64(128), int64(2)),
			wantResult: `{"level1":{"level2":"level2_value"}}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fullname := types.MakeFullName(group, tt.toolName)
			toolImpl, found := interp.ToolRegistry().GetTool(fullname)
			if !found {
				t.Fatalf("Tool %q not found", fullname)
			}
			got, err := toolImpl.Func(interp, tt.args)

			if tt.wantErrIs != nil {
				if !errors.Is(err, tt.wantErrIs) {
					t.Errorf("Expected error wrapping [%v], but got: %v", tt.wantErrIs, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			gotString, ok := got.(string)
			if !ok {
				t.Fatalf("Expected a string result, but got %T", got)
			}

			if !reflect.DeepEqual(gotString, tt.wantResult) {
				t.Errorf("Result mismatch:\n  Got:  %#v\n  Want: %#v", gotString, tt.wantResult)
			}
		})
	}
}
