// NeuroScript Version: 0.6.0
// File version: 9.1.0
// Purpose: Corrected toolSelectAgentModel to properly type-assert to the canonical types.AgentModel.
// filename: pkg/tool/agentmodel/tools_agentmodel.go
// nlines: 150
// risk_rating: HIGH

package agentmodel

import (
	"fmt"
	"sort"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// modelAdminRuntime defines the interface we expect from the runtime
// for agent model administrative operations.
type modelAdminRuntime interface {
	AgentModelsAdmin() interfaces.AgentModelAdmin
	AgentModels() interfaces.AgentModelReader
}

func getModelAdmin(rt tool.Runtime) (interfaces.AgentModelAdmin, error) {
	interp, ok := rt.(modelAdminRuntime)
	if !ok {
		return nil, fmt.Errorf("internal error: runtime does not support agent model admin operations")
	}
	return interp.AgentModelsAdmin(), nil
}

func getModelReader(rt tool.Runtime) (interfaces.AgentModelReader, error) {
	interp, ok := rt.(modelAdminRuntime)
	if !ok {
		return nil, fmt.Errorf("internal error: runtime does not support agent model read operations")
	}
	return interp.AgentModels(), nil
}

func toolRegisterAgentModel(rt tool.Runtime, args []interface{}) (interface{}, error) {
	admin, err := getModelAdmin(rt)
	if err != nil {
		return nil, err
	}
	name, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("argument 'name' must be a string")
	}
	config, ok := args[1].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("argument 'config' must be a map[string]interface{}")
	}

	if _, ok := config["model"]; !ok {
		return nil, lang.ErrInvalidArgument
	}

	err = admin.Register(types.AgentModelName(name), config)
	if err != nil {
		return nil, err
	}
	return true, nil
}

func toolUpdateAgentModel(rt tool.Runtime, args []interface{}) (interface{}, error) {
	admin, err := getModelAdmin(rt)
	if err != nil {
		return nil, err
	}
	name, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("argument 'name' must be a string")
	}
	updates, ok := args[1].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("argument 'updates' must be a map[string]interface{}")
	}
	err = admin.Update(types.AgentModelName(name), updates)
	if err != nil {
		if err.Error() == "agent model not found" {
			return nil, lang.ErrNotFound
		}
		return nil, err
	}
	return true, nil
}

func toolDeleteAgentModel(rt tool.Runtime, args []interface{}) (interface{}, error) {
	admin, err := getModelAdmin(rt)
	if err != nil {
		return nil, err
	}
	name, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("argument 'name' must be a string")
	}
	if ok := admin.Delete(types.AgentModelName(name)); !ok {
		return false, nil
	}
	return true, nil
}

func toolListAgentModels(rt tool.Runtime, args []interface{}) (interface{}, error) {
	reader, err := getModelReader(rt)
	if err != nil {
		return nil, err
	}
	names := reader.List()
	return names, nil
}

func toolSelectAgentModel(rt tool.Runtime, args []interface{}) (interface{}, error) {
	reader, err := getModelReader(rt)
	if err != nil {
		return nil, err
	}

	var name string
	if args[0] != nil {
		var ok bool
		name, ok = args[0].(string)
		if !ok {
			return nil, fmt.Errorf("argument 'name' must be a string")
		}
	}

	if name == "" {
		models := reader.List()
		if len(models) == 0 {
			return nil, fmt.Errorf("%w: no agent models available to select from", lang.ErrNotFound)
		}
		sort.Slice(models, func(i, j int) bool {
			return models[i] < models[j]
		})
		return string(models[0]), nil
	}

	info, found := reader.Get(types.AgentModelName(name))
	if !found {
		return nil, fmt.Errorf("%w: agent model %q not found", lang.ErrNotFound, name)
	}

	model, ok := info.(types.AgentModel)
	if !ok {
		return nil, fmt.Errorf("internal error: retrieved agent model is not of type types.AgentModel, but %T", info)
	}
	return string(model.Name), nil
}
