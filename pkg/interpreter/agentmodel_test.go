// NeuroScript Version: 0.8.0
// File version: 14.0.0
// Purpose: Refactored to use the centralized TestHarness for robust and consistent interpreter initialization.
// filename: pkg/interpreter/interpreter_agentmodel_test.go
// nlines: 120
// risk_rating: LOW

package interpreter_test

import (
	"reflect"
	"sort"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/types"
)

func TestInterpreterAgentModelManagement(t *testing.T) {
	t.Run("RegisterAgentModel success", func(t *testing.T) {
		t.Logf("[DEBUG] Turn 1: Starting 'RegisterAgentModel success' test.")
		h := NewTestHarness(t)
		h.Interpreter.ExecPolicy = &policy.ExecPolicy{Context: policy.ContextConfig, Allow: []string{"*"}}
		interp := h.Interpreter

		config := map[string]lang.Value{
			"provider": lang.StringValue{Value: "test_provider"},
			"model":    lang.StringValue{Value: "test_model"},
		}
		agentName := types.AgentModelName("test_agent")
		t.Logf("[DEBUG] Turn 2: Registering agent model.")
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
		t.Logf("[DEBUG] Turn 3: Assertions passed.")
	})

	t.Run("RegisterAgentModel missing required fields", func(t *testing.T) {
		t.Logf("[DEBUG] Turn 1: Starting 'RegisterAgentModel missing required fields' test.")
		h := NewTestHarness(t)
		h.Interpreter.ExecPolicy = &policy.ExecPolicy{Context: policy.ContextConfig, Allow: []string{"*"}}
		interp := h.Interpreter

		config := map[string]lang.Value{
			"provider": lang.StringValue{Value: "p"},
		}
		agentName := types.AgentModelName("bad_agent")
		t.Logf("[DEBUG] Turn 2: Attempting to register agent with missing fields.")
		err := interp.RegisterAgentModel(agentName, config)
		if err == nil {
			t.Fatal("RegisterAgentModel() should have failed for missing 'model' field, but it succeeded.")
		}
		t.Logf("[DEBUG] Turn 3: Correctly received expected error: %v", err)
	})

	t.Run("List and Delete AgentModels", func(t *testing.T) {
		t.Logf("[DEBUG] Turn 1: Starting 'List and Delete AgentModels' test.")
		h := NewTestHarness(t)
		h.Interpreter.ExecPolicy = &policy.ExecPolicy{Context: policy.ContextConfig, Allow: []string{"*"}}
		interp := h.Interpreter

		config1 := map[string]lang.Value{"provider": lang.StringValue{Value: "p"}, "model": lang.StringValue{Value: "m1"}}
		config2 := map[string]lang.Value{"provider": lang.StringValue{Value: "p"}, "model": lang.StringValue{Value: "m2"}}
		agent1Name := types.AgentModelName("agent1")
		agent2Name := types.AgentModelName("agent2")

		_ = interp.RegisterAgentModel(agent1Name, config1)
		_ = interp.RegisterAgentModel(agent2Name, config2)
		t.Logf("[DEBUG] Turn 2: Two agent models registered.")

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
		t.Logf("[DEBUG] Turn 3: List assertion passed.")

		deleted := interp.DeleteAgentModel(agent1Name)
		if !deleted {
			t.Error("DeleteAgentModel() returned false for an existing agent.")
		}
		t.Logf("[DEBUG] Turn 4: Agent 'agent1' deleted.")

		finalList := interp.ListAgentModels()
		if len(finalList) != 1 || finalList[0] != agent2Name {
			t.Errorf("ListAgentModels() after delete is incorrect. Got: %v", finalList)
		}
		t.Logf("[DEBUG] Turn 5: Final list assertion passed.")
	})

	t.Run("UpdateAgentModel success", func(t *testing.T) {
		t.Logf("[DEBUG] Turn 1: Starting 'UpdateAgentModel success' test.")
		h := NewTestHarness(t)
		h.Interpreter.ExecPolicy = &policy.ExecPolicy{Context: policy.ContextConfig, Allow: []string{"*"}}
		interp := h.Interpreter

		initialConfig := map[string]lang.Value{
			"provider": lang.StringValue{Value: "p_orig"},
			"model":    lang.StringValue{Value: "m_orig"},
		}
		agentName := types.AgentModelName("agent_to_update")
		_ = interp.RegisterAgentModel(agentName, initialConfig)
		t.Logf("[DEBUG] Turn 2: Initial agent registered.")

		updates := map[string]lang.Value{
			"model": lang.StringValue{Value: "m_new"},
		}
		err := interp.UpdateAgentModel(agentName, updates)
		if err != nil {
			t.Fatalf("UpdateAgentModel() returned an unexpected error: %v", err)
		}
		t.Logf("[DEBUG] Turn 3: Agent model updated.")

		modelAny, _ := interp.GetAgentModel(agentName)
		model, _ := modelAny.(types.AgentModel)

		if model.Provider != "p_orig" {
			t.Error("UpdateAgentModel() incorrectly changed a non-updated field.")
		}
		if model.Model != "m_new" {
			t.Errorf("UpdateAgentModel() failed to update model field. Got: %s, Want: %s", model.Model, "m_new")
		}
		t.Logf("[DEBUG] Turn 4: Assertions passed.")
	})
}
