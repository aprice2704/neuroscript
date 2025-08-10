// NeuroScript Version: 0.6.0
// File version: 1
// Purpose: Contains unit tests for the internal AgentModel management methods on the Interpreter.
// filename: pkg/interpreter/interpreter_agentmodel_test.go
// nlines: 110
// risk_rating: LOW

package interpreter

import (
	"reflect"
	"sort"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

// TestInterpreterAgentModelManagement provides direct unit tests for the interpreter's
// internal AgentModel handling methods, ensuring the core state management is correct.
func TestInterpreterAgentModelManagement(t *testing.T) {
	t.Run("RegisterAgentModel success", func(t *testing.T) {
		interp, _ := newLocalTestInterpreter(t, nil, nil)
		config := map[string]lang.Value{
			"provider": lang.StringValue{Value: "test_provider"},
			"model":    lang.StringValue{Value: "test_model"},
		}

		err := interp.RegisterAgentModel("test_agent", config)
		if err != nil {
			t.Fatalf("RegisterAgentModel() returned an unexpected error: %v", err)
		}

		model, exists := interp.GetAgentModel("test_agent")
		if !exists {
			t.Fatal("GetAgentModel() failed to find the newly registered agent.")
		}
		if model.Name != "test_agent" || model.Provider != "test_provider" {
			t.Errorf("Registered agent has incorrect data. Got: %+v", model)
		}
	})

	t.Run("RegisterAgentModel duplicate name error", func(t *testing.T) {
		interp, _ := newLocalTestInterpreter(t, nil, nil)
		config := map[string]lang.Value{
			"provider": lang.StringValue{Value: "p"},
			"model":    lang.StringValue{Value: "m"},
		}
		_ = interp.RegisterAgentModel("duplicate_agent", config)

		// Try to register again with the same name
		err := interp.RegisterAgentModel("duplicate_agent", config)
		if err == nil {
			t.Fatal("RegisterAgentModel() did not return an error for a duplicate agent name.")
		}
	})

	t.Run("RegisterAgentModel missing required fields", func(t *testing.T) {
		interp, _ := newLocalTestInterpreter(t, nil, nil)
		badConfig := map[string]lang.Value{
			"provider": lang.StringValue{Value: "p"},
			// model is missing
		}
		err := interp.RegisterAgentModel("bad_agent", badConfig)
		if err == nil {
			t.Fatal("RegisterAgentModel() should have failed for missing 'model' field, but it succeeded.")
		}
	})

	t.Run("List and Delete AgentModels", func(t *testing.T) {
		interp, _ := newLocalTestInterpreter(t, nil, nil)
		config1 := map[string]lang.Value{"provider": lang.StringValue{Value: "p"}, "model": lang.StringValue{Value: "m1"}}
		config2 := map[string]lang.Value{"provider": lang.StringValue{Value: "p"}, "model": lang.StringValue{Value: "m2"}}

		_ = interp.RegisterAgentModel("agent1", config1)
		_ = interp.RegisterAgentModel("agent2", config2)

		// Test List
		initialList := interp.ListAgentModels()
		sort.Strings(initialList)
		expected := []string{"agent1", "agent2"}
		if !reflect.DeepEqual(initialList, expected) {
			t.Errorf("ListAgentModels() mismatch. Got: %v, Want: %v", initialList, expected)
		}

		// Test Delete
		deleted := interp.DeleteAgentModel("agent1")
		if !deleted {
			t.Error("DeleteAgentModel() returned false for an existing agent.")
		}

		finalList := interp.ListAgentModels()
		if len(finalList) != 1 || finalList[0] != "agent2" {
			t.Errorf("ListAgentModels() after delete is incorrect. Got: %v", finalList)
		}

		// Test deleting non-existent model
		deleted = interp.DeleteAgentModel("non_existent_agent")
		if deleted {
			t.Error("DeleteAgentModel() returned true for a non-existent agent.")
		}
	})

	t.Run("UpdateAgentModel success", func(t *testing.T) {
		interp, _ := newLocalTestInterpreter(t, nil, nil)
		initialConfig := map[string]lang.Value{
			"provider": lang.StringValue{Value: "p_orig"},
			"model":    lang.StringValue{Value: "m_orig"},
		}
		_ = interp.RegisterAgentModel("agent_to_update", initialConfig)

		updates := map[string]lang.Value{
			"model": lang.StringValue{Value: "m_new"},
		}
		err := interp.UpdateAgentModel("agent_to_update", updates)
		if err != nil {
			t.Fatalf("UpdateAgentModel() returned an unexpected error: %v", err)
		}

		model, _ := interp.GetAgentModel("agent_to_update")
		if model.Provider != "p_orig" {
			t.Error("UpdateAgentModel() incorrectly changed a non-updated field.")
		}
		if model.Model != "m_new" {
			t.Errorf("UpdateAgentModel() failed to update model field. Got: %s, Want: %s", model.Model, "m_new")
		}
	})

	t.Run("UpdateAgentModel non-existent error", func(t *testing.T) {
		interp, _ := newLocalTestInterpreter(t, nil, nil)
		updates := map[string]lang.Value{"model": lang.StringValue{Value: "m_new"}}
		err := interp.UpdateAgentModel("non_existent", updates)
		if err == nil {
			t.Fatal("UpdateAgentModel() did not return an error for a non-existent agent.")
		}
	})
}
