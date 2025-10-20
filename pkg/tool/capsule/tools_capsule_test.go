// NeuroScript Version: 0.7.2
// File version: 14
// Purpose: Corrected test setup to initialize the interpreter with a valid HostContext, fixing a panic.
// filename: pkg/tool/capsule/tools_test.go
// nlines: 111
// risk_rating: MEDIUM
package capsule_test

import (
	"bytes"
	"errors"
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/capsule"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/tool"
	toolcapsule "github.com/aprice2704/neuroscript/pkg/tool/capsule"
	"github.com/aprice2704/neuroscript/pkg/types"
)

type capsuleTestCase struct {
	name          string
	toolName      types.ToolName
	args          []interface{}
	setupFunc     func(t *testing.T, interp *interpreter.Interpreter) error
	checkFunc     func(t *testing.T, interp tool.Runtime, result interface{}, err error)
	wantResult    interface{}
	wantToolErrIs error
	isPrivileged  bool
}

func newCapsuleTestInterpreter(t *testing.T, isPrivileged bool) *interpreter.Interpreter {
	t.Helper()

	hostCtx := &interpreter.HostContext{
		Logger: logging.NewTestLogger(t),
		Stdout: &bytes.Buffer{},
		Stdin:  &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	}

	var testPolicy *policy.ExecPolicy
	var opts []interpreter.InterpreterOption

	if isPrivileged {
		testPolicy = policy.NewBuilder(policy.ContextConfig).
			Allow("tool.capsule.*").
			Grant("capsule:write:*").
			Build()
		adminRegistry := capsule.NewRegistry()
		opts = append(opts, interpreter.WithCapsuleAdminRegistry(adminRegistry))
	} else {
		testPolicy = policy.NewBuilder(policy.ContextNormal).
			Allow(
				"tool.capsule.List",
				"tool.capsule.Read",
				"tool.capsule.GetLatest",
			).Build()
	}
	opts = append(opts, interpreter.WithExecPolicy(testPolicy))
	opts = append(opts, interpreter.WithHostContext(hostCtx))

	interp := interpreter.NewInterpreter(opts...)

	// for _, toolImpl := range toolcapsule.CapsuleToolsToRegister {
	// 	if _, err := interp.ToolRegistry().RegisterTool(toolImpl); err != nil {
	// 		t.Fatalf("Failed to register tool '%s': %v", toolImpl.Spec.Name, err)
	// 	}
	// }
	return interp
}

func testCapsuleToolHelper(t *testing.T, tc capsuleTestCase) {
	t.Helper()

	interp := newCapsuleTestInterpreter(t, tc.isPrivileged)

	if tc.setupFunc != nil {
		if err := tc.setupFunc(t, interp); err != nil {
			t.Fatalf("Setup function failed for test '%s': %v", tc.name, err)
		}
	}

	fullname := types.MakeFullName(toolcapsule.Group, string(tc.toolName))
	toolImpl, found := interp.ToolRegistry().GetTool(fullname)

	if !found {
		t.Fatalf("Tool %q not found in registry", fullname)
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
		t.Fatalf("Unexpected error for test '%s': %v", tc.name, err)
	}

	if err == nil {
		if tc.wantResult != nil && !reflect.DeepEqual(result, tc.wantResult) {
			t.Errorf("Result mismatch.\nGot:    %#v\nWanted: %#v", result, tc.wantResult)
		}
	}
}
