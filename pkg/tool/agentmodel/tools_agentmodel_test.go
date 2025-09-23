// NeuroScript Version: 0.7.3
// File version: 12
// Purpose: Adds a new test case for the 'Exists' tool.
// filename: pkg/tool/agentmodel/tools_agentmodel_test.go
// nlines: 284
// risk_rating: MEDIUM

package agentmodel_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/capability"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/policy"
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

	testPolicy := &policy.ExecPolicy{
		Context: policy.ContextConfig,
		Allow:   []string{"tool.agentmodel.*"},
		Grants: capability.NewGrantSet(
			[]capability.Capability{
				{Resource: "model", Verbs: []string{"admin", "read"}, Scopes: []string{"*"}},
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
		"temperature":         0.8,
		"top_p":               0.9,
		"top_k":               40.0,
		"max_output_tokens":   2048.0,
		"response_format":     "json_object",
		"tool_choice":         "auto",
		"safe_prompt":         true,
	}
}

func TestToolAgentModel_Register(t *testing.T) {
	tests := []agentModelTestCase{
		{
			name:       "Success: Register a new model with loop fields",
			toolName:   "Register",
			args:       []interface{}{"test_model_1", newValidModelConfig("p1", "m1")},
			wantResult: true,
			checkFunc: func(t *testing.T, interp tool.Runtime, result interface{}, err error) {
				if err != nil {
					t.Fatalf("Unexpected registration error: %v", err)
				}
				reader := interp.(interface {
					AgentModels() interfaces.AgentModelReader
				}).AgentModels()
				modelAny, found := reader.Get("test_model_1")
				if !found {
					t.Fatal("Registered model not found")
				}
				model, ok := modelAny.(types.AgentModel)
				if !ok {
					t.Fatalf("Retrieved model is not of type types.AgentModel")
				}
				if model.Generation.Temperature != 0.8 {
					t.Errorf("Expected temperature 0.8, got %f", model.Generation.Temperature)
				}
				if model.Tools.ToolChoice != types.ToolChoiceAuto {
					t.Errorf("Expected tool_choice 'auto', got %q", model.Tools.ToolChoice)
				}
				if !model.Safety.SafePrompt {
					t.Error("Expected safe_prompt to be true")
				}
			},
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
		return interp.AgentModelsAdmin().Register("model_to_update", config)
	}

	tests := []agentModelTestCase{
		{
			name:       "Success: Update existing model",
			toolName:   "Update",
			args:       []interface{}{"model_to_update", map[string]interface{}{"provider": "updated_provider", "temperature": 0.5}},
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

func TestToolAgentModel_Get(t *testing.T) {
	setup := func(t *testing.T, interp *interpreter.Interpreter) error {
		config := newValidModelConfig("p1", "m1")
		config["account_name"] = "test-account"
		return interp.AgentModelsAdmin().Register("model_to_get", config)
	}

	tests := []agentModelTestCase{
		{
			name:      "Success: Get existing model and verify account_name key",
			toolName:  "Get",
			args:      []interface{}{"model_to_get"},
			setupFunc: setup,
			checkFunc: func(t *testing.T, interp tool.Runtime, result interface{}, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				modelMapVal, ok := result.(*lang.MapValue)
				if !ok {
					t.Fatalf("Expected *lang.MapValue, got %T", result)
				}
				unwrapped := lang.Unwrap(modelMapVal)
				modelMap, ok := unwrapped.(map[string]any)
				if !ok {
					t.Fatalf("Expected unwrapped map, got %T", unwrapped)
				}

				if _, ok := modelMap["AccountName"]; ok {
					t.Error("Found unexpected key 'AccountName'; should be 'account_name'")
				}

				val, ok := modelMap["account_name"]
				if !ok {
					t.Errorf("Expected 'account_name' key in map, but it was not found. Map keys: %v", reflect.ValueOf(modelMap).MapKeys())
				} else if val != "test-account" {
					t.Errorf("Expected account_name 'test-account', got %v", val)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testAgentModelToolHelper(t, tt)
		})
	}
}

func TestToolAgentModel_Exists(t *testing.T) {
	setup := func(t *testing.T, interp *interpreter.Interpreter) error {
		config := newValidModelConfig("p1", "m1")
		return interp.AgentModelsAdmin().Register("existing-model", config)
	}

	tests := []agentModelTestCase{
		{
			name:       "Success: Model exists",
			toolName:   "Exists",
			args:       []interface{}{"existing-model"},
			setupFunc:  setup,
			wantResult: true,
		},
		{
			name:       "Success: Model does not exist",
			toolName:   "Exists",
			args:       []interface{}{"non-existent-model"},
			setupFunc:  setup,
			wantResult: false,
		},
		{
			name:       "Success: Model exists (case-insensitive)",
			toolName:   "Exists",
			args:       []interface{}{"EXISTING-MODEL"},
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
