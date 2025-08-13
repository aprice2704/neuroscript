// NeuroScript Version: 0.6.0
// File version: 4
// Purpose: Contains functional tests for the agentmodel toolset. Corrected test data setup to properly create lang.Value maps.
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
func newValidModelConfig(name, provider, model string) map[string]interface{} {
	return map[string]interface{}{
		"name":     name,
		"provider": provider,
		"model":    model,
	}
}

// toLangValueMap converts a map[string]interface{} to a map[string]lang.Value.
func toLangValueMap(t *testing.T, m map[string]interface{}) map[string]lang.Value {
	t.Helper()
	langMap := make(map[string]lang.Value)
	for k, v := range m {
		val, err := lang.Wrap(v)
		if err != nil {
			t.Fatalf("Failed to wrap value for key '%s': %v", k, err)
		}
		langMap[k] = val
	}
	return langMap
}

func TestToolAgentModel_Register(t *testing.T) {
	tests := []agentModelTestCase{
		{
			name:       "Success: Register a new model",
			toolName:   "Register",
			args:       []interface{}{"test_model_1", newValidModelConfig("test_model_1", "p1", "m1")},
			wantResult: true,
		},
		{
			name:          "Fail: Missing required fields",
			toolName:      "Register",
			args:          []interface{}{"test_model_2", map[string]interface{}{"provider": "p2"}},
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
		config := toLangValueMap(t, newValidModelConfig("model_to_update", "p1", "m1"))
		return interp.RegisterAgentModel("model_to_update", config)
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
			wantResult:    false,
			wantToolErrIs: lang.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testAgentModelToolHelper(t, tt)
		})
	}
}

func TestToolAgentModel_Delete(t *testing.T) {
	setup := func(t *testing.T, interp *interpreter.Interpreter) error {
		config := toLangValueMap(t, newValidModelConfig("model_to_delete", "p1", "m1"))
		return interp.RegisterAgentModel("model_to_delete", config)
	}

	tests := []agentModelTestCase{
		{
			name:       "Success: Delete existing model",
			toolName:   "Delete",
			args:       []interface{}{"model_to_delete"},
			setupFunc:  setup,
			wantResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testAgentModelToolHelper(t, tt)
		})
	}
}

func TestToolAgentModel_ListAndSelect(t *testing.T) {
	setup := func(t *testing.T, interp *interpreter.Interpreter) error {
		config1 := toLangValueMap(t, newValidModelConfig("z_model", "p1", "m1"))
		if err := interp.RegisterAgentModel("z_model", config1); err != nil {
			return err
		}
		config2 := toLangValueMap(t, newValidModelConfig("a_model", "p2", "m2"))
		if err := interp.RegisterAgentModel("a_model", config2); err != nil {
			return err
		}
		return nil
	}

	listTest := agentModelTestCase{
		name:      "Success: List models",
		toolName:  "List",
		args:      []interface{}{},
		setupFunc: setup,
		checkFunc: func(t *testing.T, interp tool.Runtime, result interface{}, err error) {
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			resSlice, ok := result.([]types.AgentModelName)
			if !ok {
				t.Fatalf("Expected a slice of AgentModelName, got %T", result)
			}
			if len(resSlice) != 2 {
				t.Fatalf("Expected 2 models, got %d", len(resSlice))
			}
		},
	}

	selectTest := agentModelTestCase{
		name:      "Success: Select first available model",
		toolName:  "Select",
		args:      []interface{}{nil},
		setupFunc: setup,
	}

	t.Run(listTest.name, func(t *testing.T) {
		testAgentModelToolHelper(t, listTest)
	})
	t.Run(selectTest.name, func(t *testing.T) {
		testAgentModelToolHelper(t, selectTest)
	})
}
