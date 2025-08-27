// NeuroScript Version: 0.7.0
// File version: 5
// Purpose: Updated tests to use strongly-typed AgentModel fields and a valid config map for registration.
// filename: pkg/tool/agentmodel/tools_agentmodel_test.go
// nlines: 200
// risk_rating: MEDIUM

package agentmodel_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/policy/capability"
	"github.com/aprice2704/neuroscript/pkg/runtime"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/tool/agentmodel"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// agentModelTestCase defines the structure for a single agentmodel tool test case.
type agentModelTestCase struct {
	name          string
	toolName      types.ToolName
	args          []interface{}
	setupFunc     func(t *testing.T, interp *interpreter.Interpreter) error
	checkFunc     func(t *testing.T, interp tool.Runtime, result interface{}, err error)
	wantResult    interface{}
	wantToolErrIs error
}

// newAgentModelTestInterpreter creates an interpreter with policy for agentmodel testing.
func newAgentModelTestInterpreter(t *testing.T) *interpreter.Interpreter {
	t.Helper()

	testPolicy := &runtime.ExecPolicy{
		Context: runtime.ContextConfig,
		Allow:   []string{"tool.agentmodel.*"},
		Grants: capability.NewGrantSet(
			[]capability.Capability{
				{Resource: "model", Verbs: []string{"admin"}, Scopes: []string{"*"}},
			},
			capability.Limits{},
		),
	}

	interp := interpreter.NewInterpreter(interpreter.WithExecPolicy(testPolicy))

	for _, toolImpl := range agentmodel.AgentModelToolsToRegister {
		if _, err := interp.ToolRegistry().RegisterTool(toolImpl); err != nil {
			t.Fatalf("Failed to register tool '%s': %v", toolImpl.Spec.Name, err)
		}
	}
	return interp
}

// testAgentModelToolHelper provides a generic runner for agentModelTestCase tests.
func testAgentModelToolHelper(t *testing.T, tc agentModelTestCase) {
	t.Helper()

	interp := newAgentModelTestInterpreter(t)

	if tc.setupFunc != nil {
		if err := tc.setupFunc(t, interp); err != nil {
			t.Fatalf("Setup function failed for test '%s': %v", tc.name, err)
		}
	}

	fullname := types.MakeFullName(agentmodel.Group, string(tc.toolName))
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

// newValidModelConfig creates a valid model configuration map for registration.
func newValidModelConfig(provider, model string) map[string]interface{} {
	return map[string]interface{}{
		"provider":            provider,
		"model":               model,
		"tool_loop_permitted": true,
		"max_turns":           5.0, // Use float64 as ns numbers are floats
	}
}

func TestToolAgentModel_Register(t *testing.T) {
	tests := []agentModelTestCase{
		{
			name:       "Success: Register a new model with loop fields",
			toolName:   "Register",
			args:       []interface{}{"test_model_1", newValidModelConfig("p1", "m1")},
			wantResult: true,
		},
		{
			name:          "Fail: Missing required provider field",
			toolName:      "Register",
			args:          []interface{}{"test_model_2", map[string]interface{}{"model": "m2"}},
			wantToolErrIs: lang.ErrInvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testAgentModelToolHelper(t, tt)
		})
	}
}

func TestToolAgentModel_Update(t *testing.T) {
	setup := func(t *testing.T, interp *interpreter.Interpreter) error {
		config := newValidModelConfig("p1", "m1")
		// The public API for registration is via the tool, which we are testing.
		// For setup, we can call the interpreter's admin interface directly.
		return interp.AgentModelsAdmin().Register("model_to_update", config)
	}

	tests := []agentModelTestCase{
		{
			name:       "Success: Update existing model",
			toolName:   "Update",
			args:       []interface{}{"model_to_update", map[string]interface{}{"provider": "updated_provider"}},
			setupFunc:  setup,
			wantResult: true,
		},
		{
			name:          "Fail: Model not found",
			toolName:      "Update",
			args:          []interface{}{"nonexistent_model", map[string]interface{}{"provider": "p2"}},
			setupFunc:     setup,
			wantToolErrIs: lang.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testAgentModelToolHelper(t, tt)
		})
	}
}
