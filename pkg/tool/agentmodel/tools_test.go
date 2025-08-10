// NeuroScript Version: 0.6.0
// File version: 4
// Purpose: Contains unit tests for the agentmodel toolset using correct type assertions.
// filename: pkg/tool/agentmodel/tools_test.go
// nlines: 118
// risk_rating: LOW

package agentmodel

import (
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// setupTestInterp creates an interpreter. The agentmodel tools are now registered
// automatically via the init() function when this package is imported.
func setupTestInterp(t *testing.T) *interpreter.Interpreter {
	t.Helper()
	// The call to RegisterAgentModelTools is no longer needed.
	return interpreter.NewInterpreter(interpreter.WithLogger(logging.NewTestLogger(t)))
}

func TestAgentModelTools(t *testing.T) {
	t.Run("Register and List", func(t *testing.T) {
		interp := setupTestInterp(t)

		// Define the config map using lang.Value types, as expected by ExecuteTool
		config := lang.NewMapValue(map[string]lang.Value{
			"provider": lang.StringValue{Value: "openai"},
			"model":    lang.StringValue{Value: "gpt-4o"},
		})

		args := map[string]lang.Value{"name": lang.StringValue{Value: "default"}, "config": config}
		toolName := types.MakeFullName("agentmodel", "Register")
		result, err := interp.ExecuteTool(toolName, args)

		if err != nil {
			t.Fatalf("Register tool failed: %v", err)
		}
		// Correctly assert the lang.BoolValue type and its content
		if res, ok := result.(lang.BoolValue); !ok || !res.Value {
			t.Fatalf("Register tool did not return true, got: %#v", result)
		}

		// Verify registration via the List tool
		listToolName := types.MakeFullName("agentmodel", "List")
		listResultVal, err := interp.ExecuteTool(listToolName, nil)
		if err != nil {
			t.Fatalf("List tool failed: %v", err)
		}

		// Compare the result to the expected lang.ListValue
		expectedList := lang.NewListValue([]lang.Value{lang.StringValue{Value: "default"}})
		if !reflect.DeepEqual(listResultVal, expectedList) {
			t.Errorf("List result mismatch.\n  Got: %#v\n Want: %#v", listResultVal, expectedList)
		}

		_, exists := interp.GetAgentModel("default")
		if !exists {
			t.Error("GetAgentModel failed to find the registered model 'default'")
		}
	})

	t.Run("Register duplicate fails", func(t *testing.T) {
		interp := setupTestInterp(t)
		config := lang.NewMapValue(map[string]lang.Value{"provider": lang.StringValue{Value: "a"}, "model": lang.StringValue{Value: "b"}})
		args := map[string]lang.Value{"name": lang.StringValue{Value: "default"}, "config": config}
		toolName := types.MakeFullName("agentmodel", "Register")

		_, err := interp.ExecuteTool(toolName, args)
		if err != nil {
			t.Fatalf("First register call failed unexpectedly: %v", err)
		}

		_, err = interp.ExecuteTool(toolName, args)
		if err == nil {
			t.Fatal("Expected an error when registering a duplicate AgentModel, but got nil")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		interp := setupTestInterp(t)

		config := lang.NewMapValue(map[string]lang.Value{"provider": lang.StringValue{Value: "a"}, "model": lang.StringValue{Value: "b"}})
		regArgs := map[string]lang.Value{"name": lang.StringValue{Value: "to_delete"}, "config": config}
		regToolName := types.MakeFullName("agentmodel", "Register")
		_, _ = interp.ExecuteTool(regToolName, regArgs)

		delArgs := map[string]lang.Value{"name": lang.StringValue{Value: "to_delete"}}
		delToolName := types.MakeFullName("agentmodel", "Delete")
		result, err := interp.ExecuteTool(delToolName, delArgs)
		if err != nil {
			t.Fatalf("Delete tool failed: %v", err)
		}
		if res, ok := result.(lang.BoolValue); !ok || !res.Value {
			t.Fatal("Delete tool returned false for an existing model")
		}

		listToolName := types.MakeFullName("agentmodel", "List")
		listResultVal, err := interp.ExecuteTool(listToolName, nil)
		if err != nil {
			t.Fatalf("List tool failed after delete: %v", err)
		}

		// Corrected: Safely assert the type and check the length. This prevents the panic.
		list, ok := listResultVal.(lang.ListValue)
		if !ok {
			t.Fatalf("List tool did not return a ListValue, got %T", listResultVal)
		}
		if len(list.Value) != 0 {
			t.Errorf("List should be empty after delete, but got: %#v", list.Value)
		}

		delArgsNonexistent := map[string]lang.Value{"name": lang.StringValue{Value: "nonexistent"}}
		result, err = interp.ExecuteTool(delToolName, delArgsNonexistent)
		if err != nil {
			t.Fatalf("Delete tool failed for nonexistent model: %v", err)
		}
		if res, ok := result.(lang.BoolValue); !ok || res.Value {
			t.Fatal("Delete tool returned true for a nonexistent model")
		}
	})
}
