// NeuroScript Version: 0.6.0
// File version: 7.0.0
// Purpose: Updated all state-modifying functions to delegate to the root interpreter instance, fixing a critical bug where changes were lost in sandboxed clones.
// filename: pkg/interpreter/interpreter_agentmodel.go
// nlines: 130
// risk_rating: HIGH

package interpreter

import (
	"fmt"
	"sync"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// AgentModel holds the validated and parsed configuration for a specific AI model endpoint.
type AgentModel struct {
	Name           types.AgentModelName
	Provider       string
	Model          string
	SecretRef      string
	BaseURL        string
	BudgetCurrency string
	PriceTable     map[string]float64
	Temperature    float64
}

// interpreterAgentModelState holds the agent-model specific state.
type interpreterAgentModelState struct {
	agentModels   map[types.AgentModelName]AgentModel
	agentModelsMu sync.RWMutex
}

func newInterpreterAgentModelState() *interpreterAgentModelState {
	return &interpreterAgentModelState{
		agentModels: make(map[types.AgentModelName]AgentModel),
	}
}

// RegisterAgentModel adds a new AgentModel configuration to the interpreter's state.
func (i *Interpreter) RegisterAgentModel(name types.AgentModelName, config map[string]lang.Value) error {
	if i.root != nil {
		return i.root.RegisterAgentModel(name, config)
	}

	i.state.agentModelsMu.Lock()
	defer i.state.agentModelsMu.Unlock()

	if _, exists := i.state.agentModels[name]; exists {
		return fmt.Errorf("AgentModel '%s' is already registered", name)
	}

	provider, _ := lang.ToString(config["provider"])
	model, _ := lang.ToString(config["model"])

	if name == "" || provider == "" || model == "" {
		return fmt.Errorf("AgentModel registration for '%s' is missing required fields (name, provider, model)", name)
	}

	newAgentModel := AgentModel{
		Name:     name,
		Provider: provider,
		Model:    model,
	}

	if baseURL, ok := config["base_url"]; ok {
		newAgentModel.BaseURL, _ = lang.ToString(baseURL)
	}
	if secretRef, ok := config["api_key_ref"]; ok {
		newAgentModel.SecretRef, _ = lang.ToString(secretRef)
	}
	if currency, ok := config["budget_currency"]; ok {
		newAgentModel.BudgetCurrency, _ = lang.ToString(currency)
	}

	i.state.agentModels[name] = newAgentModel
	i.logger.Info("Registered new AgentModel", "name", string(name), "provider", provider, "model", model)
	return nil
}

// UpdateAgentModel modifies an existing AgentModel's configuration.
func (i *Interpreter) UpdateAgentModel(name types.AgentModelName, updates map[string]lang.Value) error {
	if i.root != nil {
		return i.root.UpdateAgentModel(name, updates)
	}

	i.state.agentModelsMu.Lock()
	defer i.state.agentModelsMu.Unlock()

	existing, exists := i.state.agentModels[name]
	if !exists {
		return fmt.Errorf("AgentModel '%s' not found for update", name)
	}

	if provider, ok := updates["provider"]; ok {
		existing.Provider, _ = lang.ToString(provider)
	}
	if model, ok := updates["model"]; ok {
		existing.Model, _ = lang.ToString(model)
	}
	if baseURL, ok := updates["base_url"]; ok {
		existing.BaseURL, _ = lang.ToString(baseURL)
	}
	if secretRef, ok := updates["api_key_ref"]; ok {
		existing.SecretRef, _ = lang.ToString(secretRef)
	}
	if currency, ok := updates["budget_currency"]; ok {
		existing.BudgetCurrency, _ = lang.ToString(currency)
	}

	i.state.agentModels[name] = existing
	i.logger.Info("Updated AgentModel", "name", string(name))
	return nil
}

// DeleteAgentModel removes an AgentModel from the interpreter's state.
func (i *Interpreter) DeleteAgentModel(name types.AgentModelName) bool {
	if i.root != nil {
		return i.root.DeleteAgentModel(name)
	}

	i.state.agentModelsMu.Lock()
	defer i.state.agentModelsMu.Unlock()

	if _, exists := i.state.agentModels[name]; !exists {
		return false
	}

	delete(i.state.agentModels, name)
	i.logger.Info("Deleted AgentModel", "name", string(name))
	return true
}

// ListAgentModels returns the names of all registered AgentModels.
func (i *Interpreter) ListAgentModels() []types.AgentModelName {
	if i.root != nil {
		return i.root.ListAgentModels()
	}

	i.state.agentModelsMu.RLock()
	defer i.state.agentModelsMu.RUnlock()

	names := make([]types.AgentModelName, 0, len(i.state.agentModels))
	for name := range i.state.agentModels {
		names = append(names, name)
	}
	return names
}

// GetAgentModel retrieves a copy of an AgentModel config.
func (i *Interpreter) GetAgentModel(name types.AgentModelName) (AgentModel, bool) {
	if i.root != nil {
		return i.root.GetAgentModel(name)
	}

	i.state.agentModelsMu.RLock()
	defer i.state.agentModelsMu.RUnlock()
	model, found := i.state.agentModels[name]
	return model, found
}
