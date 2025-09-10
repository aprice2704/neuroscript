// NeuroScript Version: 0.6.0
// File version: 4
// Purpose: Provides test helpers for the 'os' tool package. Updated to use exported tool list.
// filename: pkg/tool/os/tools_os_helpers_test.go
// nlines: 80
// risk_rating: LOW

package os_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/tool/os"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// osTestCase defines the structure for a single os tool test case.
type osTestCase struct {
	name          string
	toolName      types.ToolName
	args          []interface{}
	setupFunc     func(t *testing.T) error
	checkFunc     func(t *testing.T, interp tool.Runtime, result interface{}, err error)
	wantResult    interface{}
	wantToolErrIs error
}

// newOsTestInterpreter creates a self-contained interpreter for os tool testing.
func newOsTestInterpreter(t *testing.T) *interpreter.Interpreter {
	t.Helper()

	// Define a policy that allows the os tools to run.
	testPolicy := policy.NewBuilder(policy.ContextConfig).
		Allow("tool.os.*").
		Grant("env:read:*").
		Grant("os:exec:sleep").
		LimitPerRunCents("USD", 100). // Example limit
		LimitToolCalls("tool.os.Getenv", 10).
		Build()
	// Manually set the sleep limit for tests.
	testPolicy.Grants.Limits.TimeMaxSleepSeconds = 5

	interp := interpreter.NewInterpreter(interpreter.WithExecPolicy(testPolicy))

	// Manually register all os tools for the test interpreter
	allOsTools := append(os.OsToolsToRegister, os.OsProcToolsToRegister...)
	for _, toolImpl := range allOsTools {
		if _, err := interp.ToolRegistry().RegisterTool(toolImpl); err != nil {
			t.Fatalf("Failed to register tool '%s': %v", toolImpl.Spec.Name, err)
		}
	}
	return interp
}

// testOsToolHelper provides a generic runner for osTestCase tests.
func testOsToolHelper(t *testing.T, tc osTestCase) {
	t.Helper()

	if tc.setupFunc != nil {
		if err := tc.setupFunc(t); err != nil {
			t.Fatalf("Setup function failed for test '%s': %v", tc.name, err)
		}
	}

	interp := newOsTestInterpreter(t)

	fullname := types.MakeFullName(os.Group, string(tc.toolName))
	toolImpl, found := interp.ToolRegistry().GetTool(fullname)
	if !found {
		t.Fatalf("Tool %q not found in registry", tc.toolName)
	}

	result, err := toolImpl.Func(interp, tc.args)

	if tc.checkFunc != nil {
		tc.checkFunc(t, interp, result, err)
		return
	}

	if tc.wantToolErrIs != nil {
		if !errors.Is(err, tc.wantToolErrIs) {
			t.Errorf("Expected error wrapping [%v], but got: %v", tc.wantToolErrIs, err)
		}
	} else if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if err == nil {
		if tc.wantResult != nil {
			if !reflect.DeepEqual(result, tc.wantResult) {
				t.Errorf("Result mismatch.\nGot:    %#v\nWanted: %#v", result, tc.wantResult)
			}
		}
	}
}
