// NeuroScript Version: 0.8.0
// File version: 20.0.0
// Purpose: Re-plumbed to use the external 'policygate' package for capability checks. Removed duplicate GetExecPolicy method.
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
func (i *Interpreter) RegisterAgentModel(name types.AgentModelName, config map[string]lang.Value) error {
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
	return admin.Register(name, unwrapMapValues(config))
}

// UpdateAgentModel modifies an existing AgentModel's configuration.
func (i *Interpreter) UpdateAgentModel(name types.AgentModelName, updates map[string]lang.Value) error {
	if err := policygate.Check(i, types.CapModelAdmin); err != nil {
		return err
	}
	admin := i.AgentModelsAdmin()
	return admin.Update(name, unwrapMapValues(updates))
}

// DeleteAgentModel removes an AgentModel from the interpreter's state.
func (i *Interpreter) DeleteAgentModel(name types.AgentModelName) bool {
	if err := policygate.Check(i, types.CapModelAdmin); err != nil {
		return false // Interface doesn't allow returning an error here.
	}
	admin := i.AgentModelsAdmin()
	return admin.Delete(name)
}

// ListAgentModels returns the names of all registered AgentModels.
func (i *Interpreter) ListAgentModels() []types.AgentModelName {
	reader := i.AgentModels()
	return reader.List()
}

// GetAgentModel retrieves a copy of an AgentModel config.
func (i *Interpreter) GetAgentModel(name types.AgentModelName) (any, bool) {
	reader := i.AgentModels()
	return reader.Get(name)
}
