// NeuroScript Version: 0.7.0
// File version: 1
// Purpose: Provides advanced tests for agentmodel tools, covering selection logic and arg validation.
// filename: pkg/tool/agentmodel/tools_agentmodel_advanced_test.go
// nlines: 105
// risk_rating: MEDIUM

package agentmodel_test

import (
	"errors"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/types"
)

func TestToolAgentModel_Select(t *testing.T) {
	setup := func(t *testing.T, interp *interpreter.Interpreter) error {
		cfg1 := map[string]interface{}{"provider": "p", "model": "m1"}
		cfg2 := map[string]interface{}{"provider": "p", "model": "m2"}
		if err := interp.AgentModelsAdmin().Register("b-model", cfg1); err != nil {
			return err
		}
		if err := interp.AgentModelsAdmin().Register("a-model", cfg2); err != nil {
			return err
		}
		return nil
	}

	tests := []agentModelTestCase{
		{
			name:       "Success: Select by specific name",
			toolName:   "Select",
			args:       []interface{}{"b-model"},
			setupFunc:  setup,
			wantResult: "b-model",
		},
		{
			name:       "Success: Select default (alphabetically first)",
			toolName:   "Select",
			args:       []interface{}{""}, // or nil
			setupFunc:  setup,
			wantResult: "a-model",
		},
		{
			name:          "Fail: Select non-existent model",
			toolName:      "Select",
			args:          []interface{}{"c-model"},
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

func TestToolAgentModel_ArgumentValidation(t *testing.T) {
	tests := []agentModelTestCase{
		{
			name:          "Register - name not a string",
			toolName:      "Register",
			args:          []interface{}{123, map[string]interface{}{}},
			wantToolErrIs: errors.New("argument 'name' must be a string"),
		},
		{
			name:          "Register - config not a map",
			toolName:      "Register",
			args:          []interface{}{"a-model", "config_string"},
			wantToolErrIs: errors.New("argument 'config' must be a map[string]interface{}"),
		},
		{
			name:          "Update - updates not a map",
			toolName:      "Update",
			args:          []interface{}{"a-model", "updates_string"},
			wantToolErrIs: errors.New("argument 'updates' must be a map[string]interface{}"),
		},
		{
			name:          "Delete - name not a string",
			toolName:      "Delete",
			args:          []interface{}{false},
			wantToolErrIs: errors.New("argument 'name' must be a string"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use a simplified check for these basic validation errors.
			interp := newAgentModelTestInterpreter(t)
			fullname := types.MakeFullName("agentmodel", string(tt.toolName))
			toolImpl, _ := interp.ToolRegistry().GetTool(fullname)
			_, err := toolImpl.Func(interp, tt.args)

			if err == nil {
				t.Fatalf("Expected error but got nil")
			}
			if err.Error() != tt.wantToolErrIs.Error() {
				t.Errorf("Expected error message %q, got %q", tt.wantToolErrIs.Error(), err.Error())
			}
		})
	}
}
