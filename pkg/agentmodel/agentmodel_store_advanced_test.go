// NeuroScript Version: 0.7.0
// File version: 2
// Purpose: Provides advanced tests for AgentModelStore, covering concurrency, parsing, and updates.
// filename: pkg/agentmodel/agentmodel_store_advanced_test.go
// nlines: 231
// risk_rating: HIGH

package agentmodel

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// newFullConfig creates a map with every possible AgentModel field set.
func newFullConfig(name, provider, model string) (map[string]interface{}, types.AgentModel) {
	seed := int64(12345)
	cfg := map[string]interface{}{
		"provider":            provider,
		"model":               model,
		"api_key_ref":         "SOME_SECRET",
		"base_url":            "https://api.example.com/v1",
		"budget_currency":     "USD",
		"notes":               "Full config test",
		"disabled":            false,
		"context_ktok":        128.0,
		"max_turns":           10.0,
		"max_retries":         2.0,
		"temperature":         0.85,
		"top_p":               0.95,
		"top_k":               50.0,
		"max_output_tokens":   4096.0,
		"stop_sequences":      []interface{}{"stop1", "stop2"},
		"presence_penalty":    0.1,
		"frequency_penalty":   0.2,
		"repetition_penalty":  1.1,
		"seed":                float64(*(&seed)),
		"log_probs":           true,
		"response_format":     "json_object",
		"tool_loop_permitted": true,
		"auto_loop_enabled":   false,
		"tool_choice":         "auto",
		"safe_prompt":         true,
		"safety_settings": map[string]interface{}{
			"HARM_CATEGORY_HARASSMENT": "BLOCK_LOW_AND_ABOVE",
		},
	}

	expectedModel := types.AgentModel{
		Name:           types.AgentModelName(name),
		Provider:       provider,
		Model:          model,
		SecretRef:      "SOME_SECRET",
		BaseURL:        "https://api.example.com/v1",
		BudgetCurrency: "USD",
		Notes:          "Full config test",
		Disabled:       false,
		ContextKTok:    128,
		MaxTurns:       10,
		MaxRetries:     2,
		Generation: types.GenerationConfig{
			Temperature:       0.85,
			TopP:              0.95,
			TopK:              50,
			MaxOutputTokens:   4096,
			StopSequences:     []string{"stop1", "stop2"},
			PresencePenalty:   0.1,
			FrequencyPenalty:  0.2,
			RepetitionPenalty: 1.1,
			Seed:              &seed,
			LogProbs:          true,
			ResponseFormat:    types.ResponseFormatJSON,
		},
		Tools: types.ToolConfig{
			ToolLoopPermitted: true,
			AutoLoopEnabled:   false,
			ToolChoice:        types.ToolChoiceAuto,
		},
		Safety: types.SafetyConfig{
			SafePrompt: true,
			Settings:   map[string]string{"HARM_CATEGORY_HARASSMENT": "BLOCK_LOW_AND_ABOVE"},
		},
		// Deprecated fields
		Temperature:       0.85,
		ToolLoopPermitted: true,
		AutoLoopEnabled:   false,
	}
	return cfg, expectedModel
}

func TestAgentModelStore_FullConfigParsing(t *testing.T) {
	store := NewAgentModelStore()
	admin := NewAgentModelAdmin(store, &policy.ExecPolicy{Context: policy.ContextConfig})
	cfg, expectedModel := newFullConfig("full-model", "test-provider", "test-model-123")

	if err := admin.Register("full-model", cfg); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	rawModel, ok := store.m["full-model"]
	if !ok {
		t.Fatal("Model not found in store after registration")
	}

	// Zero out fields we don't set in the config for deep equal comparison
	expectedModel.PriceTable = rawModel.PriceTable

	if !reflect.DeepEqual(rawModel, expectedModel) {
		t.Errorf("Parsed model does not match expected model.\nGot:    %+v\nWanted: %+v", rawModel, expectedModel)
	}
}

func TestAgentModelStore_PartialUpdate(t *testing.T) {
	store := NewAgentModelStore()
	admin := NewAgentModelAdmin(store, &policy.ExecPolicy{Context: policy.ContextConfig})
	cfg, _ := newFullConfig("partial-update-model", "p1", "m1")

	if err := admin.Register("partial-update-model", cfg); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	updates := map[string]interface{}{
		"notes":       "Updated notes.",
		"temperature": 0.99,
		"disabled":    true,
	}

	if err := admin.Update("partial-update-model", updates); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	model, _ := store.m["partial-update-model"]
	if model.Notes != "Updated notes." {
		t.Errorf("Expected notes to be updated, got %q", model.Notes)
	}
	if model.Generation.Temperature != 0.99 {
		t.Errorf("Expected temperature to be updated, got %f", model.Generation.Temperature)
	}
	if !model.Disabled {
		t.Error("Expected disabled to be updated to true")
	}
	if model.Provider != "p1" {
		t.Error("Expected provider to be unchanged")
	}
}

func TestAgentModelStore_ConfigTypeMismatch(t *testing.T) {
	store := NewAgentModelStore()
	admin := NewAgentModelAdmin(store, &policy.ExecPolicy{Context: policy.ContextConfig})

	testCases := []struct {
		name    string
		field   string
		value   interface{}
		wantErr string
	}{
		{"Bad temperature", "temperature", "warm", "expected number"},
		{"Bad disabled", "disabled", "false", "expected bool"},
		{"Bad max_turns", "max_turns", "ten", "expected number"},
		{"Bad stop_sequences", "stop_sequences", "stop", "expected slice"},
		{"Bad stop_sequences item", "stop_sequences", []interface{}{123}, "expected string"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := map[string]interface{}{
				"provider": "p",
				"model":    "m",
				tc.field:   tc.value,
			}
			err := admin.Register("bad-model", cfg)
			if err == nil {
				t.Fatalf("Expected error but got nil")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("Expected error to contain %q, but got: %v", tc.wantErr, err)
			}
		})
	}
}

func TestAgentModelStore_Concurrency(t *testing.T) {
	store := NewAgentModelStore()
	admin := NewAgentModelAdmin(store, &policy.ExecPolicy{Context: policy.ContextConfig})
	reader := NewAgentModelReader(store)
	numGoroutines := 50
	numModels := 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			modelID := id % numModels
			name := fmt.Sprintf("model-%d", modelID)
			cfg := map[string]interface{}{"provider": "p", "model": "m"}

			// Mix of register, update, get, list, delete
			switch id % 4 {
			case 0: // Register
				_ = admin.Register(types.AgentModelName(name), cfg)
			case 1: // Update
				updates := map[string]interface{}{"notes": fmt.Sprintf("goroutine %d", id)}
				_ = admin.Update(types.AgentModelName(name), updates)
			case 2: // Get
				_, _ = reader.Get(types.AgentModelName(name))
			case 3: // List & Delete
				_ = reader.List()
				_ = admin.Delete(types.AgentModelName(name))
			}
		}(i)
	}

	wg.Wait()
	// The test passes if it completes without the race detector firing.
	// The final state of the store is non-deterministic and not checked.
	t.Logf("Concurrency test finished with %d models in the store.", len(store.m))
}
