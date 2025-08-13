// NeuroScript Version: 0.6.0
// File version: 17.0.0
// Purpose: Corrected the call to admin.Register to pass the expected map[string]any, fixing the core type mismatch and compiler errors.
// filename: pkg/interpreter/interpreter_agentmodel.go
// nlines: 80
// risk_rating: HIGH

package interpreter

import (
	"errors"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// unwrapMapValues converts a map of lang.Value to a map of any by unwrapping each value.
func unwrapMapValues(m map[string]lang.Value) map[string]any {
	unwrapped := make(map[string]any, len(m))
	for k, v := range m {
		unwrapped[k] = lang.Unwrap(v)
	}
	return unwrapped
}

// RegisterAgentModel adds a new AgentModel configuration to the interpreter's state.
func (i *Interpreter) RegisterAgentModel(name types.AgentModelName, config map[string]lang.Value) error {
	if i.root != nil {
		return i.root.RegisterAgentModel(name, config)
	}

	// Validate required fields before proceeding.
	_, pOk := config["provider"]
	_, mOk := config["model"]
	if !pOk || !mOk {
		return errors.New("agent model config missing required field(s): provider and model")
	}

	// The AgentModelAdmin is responsible for creating the AgentModel from the map.
	admin := i.AgentModelsAdmin()
	return admin.Register(name, unwrapMapValues(config))
}

// UpdateAgentModel modifies an existing AgentModel's configuration by delegating to the central model store.
func (i *Interpreter) UpdateAgentModel(name types.AgentModelName, updates map[string]lang.Value) error {
	if i.root != nil {
		return i.root.UpdateAgentModel(name, updates)
	}
	admin := i.AgentModelsAdmin()
	return admin.Update(name, unwrapMapValues(updates))
}

// DeleteAgentModel removes an AgentModel from the interpreter's state by delegating to the central model store.
func (i *Interpreter) DeleteAgentModel(name types.AgentModelName) bool {
	if i.root != nil {
		return i.root.DeleteAgentModel(name)
	}
	admin := i.AgentModelsAdmin()
	return admin.Delete(name)
}

// ListAgentModels returns the names of all registered AgentModels from the central model store.
func (i *Interpreter) ListAgentModels() []types.AgentModelName {
	if i.root != nil {
		return i.root.ListAgentModels()
	}
	reader := i.AgentModels()
	return reader.List()
}

// GetAgentModel retrieves a copy of an AgentModel config from the central model store.
func (i *Interpreter) GetAgentModel(name types.AgentModelName) (any, bool) {
	if i.root != nil {
		return i.root.GetAgentModel(name)
	}

	reader := i.AgentModels()
	return reader.Get(name)
}
