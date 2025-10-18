// NeuroScript Version: 0.5.2
// File version: 3
// Purpose: Switched to use central newStringTestInterpreter but re-applies the required regex policy.
// filename: pkg/tool/strtools/tools_string_regex_test.go
// nlines: 75
// risk_rating: MEDIUM

package strtools

import (
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/types"
)

func TestToolStringRegex(t *testing.T) {
	// We must create a custom interpreter for this test to set the
	// specific policy required for regex tools.
	t.Logf("DEBUG: Creating new test interpreter for regex tests with a valid HostContext.")
	hostCtx := &interpreter.HostContext{
		Logger: logging.NewTestLogger(t),
		Stdout: os.Stdout,
		Stdin:  os.Stdin,
		Stderr: os.Stderr,
	}
	// Setup: Create a policy that grants regex capability.
	testPolicy := policy.NewBuilder(policy.ContextConfig).
		Allow("tool.str.*").
		Grant("str:use:regex").
		Build()

	// We still use interpreter.NewInterpreter() which will
	// trigger the init() function in register.go to register all tools.
	interp := interpreter.NewInterpreter(
		interpreter.WithHostContext(hostCtx),
		interpreter.WithExecPolicy(testPolicy),
	)
	// We NO LONGER need the manual registration loop.
	/*
		for _, impl := range stringRegexToolsToRegister {
			if _, err := interp.ToolRegistry().RegisterTool(impl); err != nil {
				t.Fatalf("Failed to register tool %q: %v", impl.Spec.Name, err)
			}
		}
	*/

	tests := []struct {
		name       string
		toolName   string
		args       []interface{}
		wantResult interface{}
		wantErrIs  error
	}{
		// MatchRegex
		{name: "MatchRegex True", toolName: "MatchRegex", args: MakeArgs(`\d+`, "123"), wantResult: true},
		{name: "MatchRegex False", toolName: "MatchRegex", args: MakeArgs(`\d+`, "abc"), wantResult: false},
		{name: "MatchRegex Invalid Pattern", toolName: "MatchRegex", args: MakeArgs(`[`, "abc"), wantErrIs: lang.ErrInvalidArgument},

		// FindAllRegex
		{name: "FindAllRegex Simple", toolName: "FindAllRegex", args: MakeArgs(`\w+`, "one two,three"), wantResult: []string{"one", "two", "three"}},
		{name: "FindAllRegex No Match", toolName: "FindAllRegex", args: MakeArgs(`\d+`, "abc def"), wantResult: []string(nil)}, // No matches returns nil slice from regexp
		{name: "FindAllRegex Invalid Pattern", toolName: "FindAllRegex", args: MakeArgs(`(`, "abc"), wantErrIs: lang.ErrInvalidArgument},

		// ReplaceRegex
		{name: "ReplaceRegex Simple", toolName: "ReplaceRegex", args: MakeArgs(`\s+`, "one  two\tthree", "_"), wantResult: "one_two_three"},
		{name: "ReplaceRegex No Match", toolName: "ReplaceRegex", args: MakeArgs(`\d+`, "abc", "_"), wantResult: "abc"},
		{name: "ReplaceRegex Invalid Pattern", toolName: "ReplaceRegex", args: MakeArgs(`\p{`, "abc", "_"), wantErrIs: lang.ErrInvalidArgument},
	}

	for _, tt := range tests {
		// Use the helper from tools_string_basic_test.go
		// We can't use the simple helper because it doesn't set up the required policy.
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
			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Errorf("Result mismatch:\n  Got:  %#v\n  Want: %#v", got, tt.wantResult)
			}
		})
	}
}
