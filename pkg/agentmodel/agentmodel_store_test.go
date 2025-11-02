// NeuroScript Version: 0.7.0
// File version: 3
// Purpose: Updated admin/reader calls to use plain 'string' for model names instead of 'types.AgentModelName'.
// filename: pkg/agentmodel/agentmodel_store_test.go
// nlines: 198
// risk_rating: MEDIUM

package agentmodel

import (
	"errors"
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/types"
)

func TestAgentModelStore_Register(t *testing.T) {
	adminPolicy := &policy.ExecPolicy{Context: policy.ContextConfig}
	userPolicy := &policy.ExecPolicy{Context: policy.ContextNormal}

	testCases := []struct {
		name    string
		store   *AgentModelStore
		admin   interfaces.AgentModelAdmin
		model   string // Changed from types.AgentModelName
		cfg     map[string]any
		wantErr error
	}{
		{
			name:  "Success - Register new model",
			store: NewAgentModelStore(),
			admin: NewAgentModelAdmin(NewAgentModelStore(), adminPolicy),
			model: "gpt-4",
			cfg:   map[string]any{"provider": "openai", "model": "gpt-4-turbo"},
		},
		{
			name: "Fail - Duplicate model",
			store: &AgentModelStore{m: map[string]types.AgentModel{
				"gpt-4": {Name: "gpt-4"},
			}},
			admin: NewAgentModelAdmin(&AgentModelStore{m: map[string]types.AgentModel{
				"gpt-4": {Name: "gpt-4"},
			}}, adminPolicy),
			model:   "gpt-4",
			cfg:     map[string]any{"provider": "openai", "model": "gpt-4-turbo"},
			wantErr: lang.ErrDuplicateKey,
		},
		{
			name:    "Fail - Non-config context",
			store:   NewAgentModelStore(),
			admin:   NewAgentModelAdmin(NewAgentModelStore(), userPolicy),
			model:   "gpt-4",
			cfg:     map[string]any{"provider": "openai", "model": "gpt-4-turbo"},
			wantErr: policy.ErrTrust,
		},
		{
			name:    "Fail - Missing provider",
			store:   NewAgentModelStore(),
			admin:   NewAgentModelAdmin(NewAgentModelStore(), adminPolicy),
			model:   "gpt-4",
			cfg:     map[string]any{"model": "gpt-4-turbo"},
			wantErr: errors.New("'provider' and 'model' are required"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.admin.Register(tc.model, tc.cfg)

			if tc.wantErr != nil {
				if err == nil || (err.Error() != tc.wantErr.Error() && !errors.Is(err, tc.wantErr)) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestAgentModelStore_Update(t *testing.T) {
	adminPolicy := &policy.ExecPolicy{Context: policy.ContextConfig}
	userPolicy := &policy.ExecPolicy{Context: policy.ContextNormal}
	existingModel := types.AgentModel{Name: "gpt-4", Provider: "openai", Model: "gpt-4"}

	testCases := []struct {
		name    string
		store   *AgentModelStore
		admin   interfaces.AgentModelAdmin
		model   string // Changed from types.AgentModelName
		updates map[string]any
		wantErr error
	}{
		{
			name: "Success - Update existing model",
			store: &AgentModelStore{m: map[string]types.AgentModel{
				"gpt-4": existingModel,
			}},
			admin: NewAgentModelAdmin(&AgentModelStore{m: map[string]types.AgentModel{
				"gpt-4": existingModel,
			}}, adminPolicy),
			model:   "gpt-4",
			updates: map[string]any{"model": "gpt-4-turbo"},
		},
		{
			name:    "Fail - Model not found",
			store:   NewAgentModelStore(),
			admin:   NewAgentModelAdmin(NewAgentModelStore(), adminPolicy),
			model:   "gpt-4",
			updates: map[string]any{"model": "gpt-4-turbo"},
			wantErr: lang.ErrNotFound,
		},
		{
			name: "Fail - Non-config context",
			store: &AgentModelStore{m: map[string]types.AgentModel{
				"gpt-4": existingModel,
			}},
			admin:   NewAgentModelAdmin(NewAgentModelStore(), userPolicy),
			model:   "gpt-4",
			updates: map[string]any{"model": "gpt-4-turbo"},
			wantErr: policy.ErrTrust,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.admin.Update(tc.model, tc.updates)

			if tc.wantErr != nil {
				if err == nil || !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestAgentModelStore_Delete(t *testing.T) {
	adminPolicy := &policy.ExecPolicy{Context: policy.ContextConfig}
	userPolicy := &policy.ExecPolicy{Context: policy.ContextNormal}
	existingModel := types.AgentModel{Name: "gpt-4"}

	testCases := []struct {
		name      string
		store     *AgentModelStore
		admin     interfaces.AgentModelAdmin
		model     string // Changed from types.AgentModelName
		wantFound bool
	}{
		{
			name: "Success - Delete existing model",
			store: &AgentModelStore{m: map[string]types.AgentModel{
				"gpt-4": existingModel,
			}},
			admin: NewAgentModelAdmin(&AgentModelStore{m: map[string]types.AgentModel{
				"gpt-4": existingModel,
			}}, adminPolicy),
			model:     "gpt-4",
			wantFound: true,
		},
		{
			name:      "Fail - Model not found",
			store:     NewAgentModelStore(),
			admin:     NewAgentModelAdmin(NewAgentModelStore(), adminPolicy),
			model:     "gpt-4",
			wantFound: false,
		},
		{
			name: "Fail - Non-config context",
			store: &AgentModelStore{m: map[string]types.AgentModel{
				"gpt-4": existingModel,
			}},
			admin:     NewAgentModelAdmin(NewAgentModelStore(), userPolicy),
			model:     "gpt-4",
			wantFound: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			found := tc.admin.Delete(tc.model)
			if found != tc.wantFound {
				t.Fatalf("expected found to be %v, got %v", tc.wantFound, found)
			}
		})
	}
}

func TestAgentModelStore_Get_List(t *testing.T) {
	models := map[string]types.AgentModel{
		"gpt-4":    {Name: "gpt-4"},
		"claude-3": {Name: "claude-3"},
	}
	store := &AgentModelStore{m: models}
	reader := NewAgentModelReader(store)

	// Test Get
	got, ok := reader.Get("gpt-4") // Use string
	if !ok || !reflect.DeepEqual(got, models["gpt-4"]) {
		t.Errorf("Get('gpt-4') = %v, %v, want %v, true", got, ok, models["gpt-4"])
	}

	// Test List
	list := reader.List()
	if len(list) != len(models) {
		t.Errorf("List() len = %d, want %d", len(list), len(models))
	}
}
