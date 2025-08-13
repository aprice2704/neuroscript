// NeuroScript Version: 0.6.0
// File version: 12.0.0
// Purpose: Aligned tests to use the canonical types.AgentModel, resolving type assertion failures.
// filename: pkg/interpreter/interpreter_agentmodel_test.go
// nlines: 115
// risk_rating: LOW

package interpreter

import (
	"reflect"
	"sort"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// TestInterpreterAgentModelManagement provides direct unit tests for the interpreter's
// internal AgentModel handling methods, ensuring the core state management is correct.
func TestInterpreterAgentModelManagement(t *testing.T) {
	t.Run("RegisterAgentModel success", func(t *testing.T) {
		interp, _ := NewTestInterpreter(t, nil, nil, true) // Run with privileges
		config := map[string]lang.Value{
			"provider": lang.StringValue{Value: "test_provider"},
			"model":    lang.StringValue{Value: "test_model"},
		}
		agentName := types.AgentModelName("test_agent")

		err := interp.RegisterAgentModel(agentName, config)
		if err != nil {
			t.Fatalf("RegisterAgentModel() returned an unexpected error: %v", err)
		}

		modelAny, exists := interp.GetAgentModel(agentName)
		if !exists {
			t.Fatal("GetAgentModel() failed to find the newly registered agent.")
		}
		model, ok := modelAny.(types.AgentModel)
		if !ok {
			t.Fatalf("GetAgentModel() returned an unexpected type: %T", modelAny)
		}
		if model.Name != agentName || model.Provider != "test_provider" {
			t.Errorf("Registered agent has incorrect data. Got: %+v", model)
		}
	})

	t.Run("RegisterAgentModel missing required fields", func(t *testing.T) {
		interp, _ := NewTestInterpreter(t, nil, nil, true) // Run with privileges
		config := map[string]lang.Value{
			"provider": lang.StringValue{Value: "p"},
		}
		agentName := types.AgentModelName("bad_agent")
		err := interp.RegisterAgentModel(agentName, config)
		if err == nil {
			t.Fatal("RegisterAgentModel() should have failed for missing 'model' field, but it succeeded.")
		}
	})

	t.Run("List and Delete AgentModels", func(t *testing.T) {
		interp, _ := NewTestInterpreter(t, nil, nil, true) // Run with privileges
		config1 := map[string]lang.Value{"provider": lang.StringValue{Value: "p"}, "model": lang.StringValue{Value: "m1"}}
		config2 := map[string]lang.Value{"provider": lang.StringValue{Value: "p"}, "model": lang.StringValue{Value: "m2"}}
		agent1Name := types.AgentModelName("agent1")
		agent2Name := types.AgentModelName("agent2")

		_ = interp.RegisterAgentModel(agent1Name, config1)
		_ = interp.RegisterAgentModel(agent2Name, config2)

		// Test List
		initialList := interp.ListAgentModels()
		stringList := make([]string, len(initialList))
		for i, v := range initialList {
			stringList[i] = string(v)
		}
		sort.Strings(stringList)
		expected := []string{"agent1", "agent2"}
		if !reflect.DeepEqual(stringList, expected) {
			t.Errorf("ListAgentModels() mismatch. Got: %v, Want: %v", stringList, expected)
		}

		// Test Delete
		deleted := interp.DeleteAgentModel(agent1Name)
		if !deleted {
			t.Error("DeleteAgentModel() returned false for an existing agent.")
		}

		finalList := interp.ListAgentModels()
		if len(finalList) != 1 || finalList[0] != agent2Name {
			t.Errorf("ListAgentModels() after delete is incorrect. Got: %v", finalList)
		}

		// Test deleting non-existent model
		deleted = interp.DeleteAgentModel("non_existent_agent")
		if deleted {
			t.Error("DeleteAgentModel() returned true for a non-existent agent.")
		}
	})

	t.Run("UpdateAgentModel success", func(t *testing.T) {
		interp, _ := NewTestInterpreter(t, nil, nil, true) // Run with privileges
		initialConfig := map[string]lang.Value{
			"provider": lang.StringValue{Value: "p_orig"},
			"model":    lang.StringValue{Value: "m_orig"},
		}
		agentName := types.AgentModelName("agent_to_update")
		_ = interp.RegisterAgentModel(agentName, initialConfig)

		updates := map[string]lang.Value{
			"model": lang.StringValue{Value: "m_new"},
		}
		err := interp.UpdateAgentModel(agentName, updates)
		if err != nil {
			t.Fatalf("UpdateAgentModel() returned an unexpected error: %v", err)
		}

		modelAny, _ := interp.GetAgentModel(agentName)
		model, ok := modelAny.(types.AgentModel)
		if !ok {
			t.Fatalf("GetAgentModel() returned an unexpected type: %T", modelAny)
		}

		if model.Provider != "p_orig" {
			t.Error("UpdateAgentModel() incorrectly changed a non-updated field.")
		}
		if model.Model != "m_new" {
			t.Errorf("UpdateAgentModel() failed to update model field. Got: %s, Want: %s", model.Model, "m_new")
		}
	})

	t.Run("UpdateAgentModel non-existent error", func(t *testing.T) {
		interp, _ := NewTestInterpreter(t, nil, nil, true) // Run with privileges
		updates := map[string]lang.Value{"model": lang.StringValue{Value: "m_new"}}
		err := interp.UpdateAgentModel("non_existent", updates)
		if err == nil {
			t.Fatal("UpdateAgentModel() did not return an error for a non-existent agent.")
		}
	})
}
