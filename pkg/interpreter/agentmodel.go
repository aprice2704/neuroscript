// NeuroScript Version: 0.8.0
// File version: 21.0.0
// Purpose: Updated all method signatures to use 'string' instead of 'types.AgentModelName' to match the purified interface.
// filename: pkg/interpreter/agentmodel.go
// nlines: 75
// risk_rating: HIGH

package interpreter

import (
	"errors"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/policygate"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// unwrapMapValues is a private helper from the original implementation.
func unwrapMapValues(m map[string]lang.Value) map[string]any {
	unwrapped := make(map[string]any, len(m))
	for k, v := range m {
		unwrapped[k] = lang.Unwrap(v)
	}
	return unwrapped
}

// RegisterAgentModel adds a new AgentModel configuration.
// FIX: Signature changed from types.AgentModelName to string
func (i *Interpreter) RegisterAgentModel(name string, config map[string]lang.Value) error {
	if err := policygate.Check(i, types.CapModelAdmin); err != nil {
		return err
	}
	if _, apiKeyExists := config["api_key"]; apiKeyExists {
		return errors.New("agent model config cannot contain 'api_key'; use 'account_name' instead")
	}
	_, pOk := config["provider"]
	_, mOk := config["model"]
	if !pOk || !mOk {
		return errors.New("agent model config missing required field(s): provider and model")
	}
	admin := i.AgentModelsAdmin()
	return admin.Register(name, unwrapMapValues(config)) // This now matches the interface
}

// UpdateAgentModel modifies an existing AgentModel's configuration.
// FIX: Signature changed from types.AgentModelName to string
func (i *Interpreter) UpdateAgentModel(name string, updates map[string]lang.Value) error {
	if err := policygate.Check(i, types.CapModelAdmin); err != nil {
		return err
	}
	admin := i.AgentModelsAdmin()
	return admin.Update(name, unwrapMapValues(updates)) // This now matches the interface
}

// DeleteAgentModel removes an AgentModel from the interpreter's state.
// FIX: Signature changed from types.AgentModelName to string
func (i *Interpreter) DeleteAgentModel(name string) bool {
	if err := policygate.Check(i, types.CapModelAdmin); err != nil {
		return false // Interface doesn't allow returning an error here.
	}
	admin := i.AgentModelsAdmin()
	return admin.Delete(name) // This now matches the interface
}

// ListAgentModels returns the names of all registered AgentModels.
// FIX: Return type changed from []types.AgentModelName to []string
func (i *Interpreter) ListAgentModels() []string {
	reader := i.AgentModels()
	return reader.List() // This now matches the interface
}

// GetAgentModel retrieves a copy of an AgentModel config.
// FIX: Signature changed from types.AgentModelName to string
func (i *Interpreter) GetAgentModel(name string) (any, bool) {
	reader := i.AgentModels()
	return reader.Get(name) // This now matches the interface
}
