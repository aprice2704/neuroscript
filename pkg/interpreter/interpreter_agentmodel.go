// NeuroScript Version: 0.6.0
// File version: 1
// Purpose: Defines the AgentModel struct and the interpreter's internal machinery for managing them.
// filename: pkg/interpreter/interpreter_agentmodel.go
// nlines: 115
// risk_rating: MEDIUM

package interpreter

import (
	"fmt"
	"sync"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

// AgentModel holds the validated and parsed configuration for a specific AI model endpoint.
type AgentModel struct {
	Name        string
	Provider    string
	Model       string
	APIKey      string
	BaseURL     string
	PriceTable  map[string]float64
	Temperature float64
	// Add other configuration fields here as needed
}

// interpreterAgentModelState holds the agent-model specific state.
type interpreterAgentModelState struct {
	agentModels   map[string]AgentModel
	agentModelsMu sync.RWMutex
}

func newInterpreterAgentModelState() *interpreterAgentModelState {
	return &interpreterAgentModelState{
		agentModels: make(map[string]AgentModel),
	}
}

// RegisterAgentModel adds a new AgentModel configuration to the interpreter's state.
// This is the underlying method called by the corresponding tool.
func (i *Interpreter) RegisterAgentModel(name string, config map[string]lang.Value) error {
	i.state.agentModelsMu.Lock()
	defer i.state.agentModelsMu.Unlock()

	if _, exists := i.state.agentModels[name]; exists {
		return fmt.Errorf("AgentModel '%s' is already registered", name)
	}

	// Basic validation and parsing of the config map
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
	// ... (parse other optional fields like api_key, temperature, etc.)

	i.state.agentModels[name] = newAgentModel
	i.logger.Info("Registered new AgentModel", "name", name, "provider", provider, "model", model)
	return nil
}

// UpdateAgentModel modifies an existing AgentModel's configuration.
func (i *Interpreter) UpdateAgentModel(name string, updates map[string]lang.Value) error {
	i.state.agentModelsMu.Lock()
	defer i.state.agentModelsMu.Unlock()

	existing, exists := i.state.agentModels[name]
	if !exists {
		return fmt.Errorf("AgentModel '%s' not found for update", name)
	}

	// Merge updates into the existing config
	if provider, ok := updates["provider"]; ok {
		existing.Provider, _ = lang.ToString(provider)
	}
	if model, ok := updates["model"]; ok {
		existing.Model, _ = lang.ToString(model)
	}
	// ... (update other fields)

	i.state.agentModels[name] = existing
	i.logger.Info("Updated AgentModel", "name", name)
	return nil
}

// DeleteAgentModel removes an AgentModel from the interpreter's state.
func (i *Interpreter) DeleteAgentModel(name string) bool {
	i.state.agentModelsMu.Lock()
	defer i.state.agentModelsMu.Unlock()

	if _, exists := i.state.agentModels[name]; !exists {
		return false
	}

	delete(i.state.agentModels, name)
	i.logger.Info("Deleted AgentModel", "name", name)
	return true
}

// ListAgentModels returns the names of all registered AgentModels.
func (i *Interpreter) ListAgentModels() []string {
	i.state.agentModelsMu.RLock()
	defer i.state.agentModelsMu.RUnlock()

	names := make([]string, 0, len(i.state.agentModels))
	for name := range i.state.agentModels {
		names = append(names, name)
	}
	return names
}

// GetAgentModel retrieves a copy of an AgentModel config.
func (i *Interpreter) GetAgentModel(name string) (AgentModel, bool) {
	i.state.agentModelsMu.RLock()
	defer i.state.agentModelsMu.RUnlock()
	model, found := i.state.agentModels[name]
	return model, found
}
