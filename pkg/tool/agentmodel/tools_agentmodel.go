// NeuroScript Version: 0.7.3
// File version: 8
// Purpose: Implemented the new 'Exists' tool function.
// filename: pkg/tool/agentmodel/tools_agentmodel.go
// nlines: 161
// risk_rating: HIGH
package agentmodel

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/types"
)

type agentModelRuntime interface {
	AgentModelsAdmin() interfaces.AgentModelAdmin
	AgentModels() interfaces.AgentModelReader
}

func getAgentModelAdmin(rt tool.Runtime) (interfaces.AgentModelAdmin, error) {
	interp, ok := rt.(agentModelRuntime)
	if !ok {
		return nil, fmt.Errorf("internal error: runtime does not support agent model admin operations")
	}
	return interp.AgentModelsAdmin(), nil
}

func getAgentModelReader(rt tool.Runtime) (interfaces.AgentModelReader, error) {
	interp, ok := rt.(agentModelRuntime)
	if !ok {
		return nil, fmt.Errorf("internal error: runtime does not support agent model read operations")
	}
	return interp.AgentModels(), nil
}

func toolRegisterAgentModel(rt tool.Runtime, args []interface{}) (interface{}, error) {
	admin, err := getAgentModelAdmin(rt)
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
	err = admin.Register(types.AgentModelName(name), config)
	if err != nil {
		if strings.Contains(err.Error(), "are required") {
			return nil, fmt.Errorf("%w: %v", lang.ErrInvalidArgument, err)
		}
		return nil, err
	}
	return true, nil
}

func toolUpdateAgentModel(rt tool.Runtime, args []interface{}) (interface{}, error) {
	admin, err := getAgentModelAdmin(rt)
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
		return nil, err
	}
	return true, nil
}

func toolDeleteAgentModel(rt tool.Runtime, args []interface{}) (interface{}, error) {
	admin, err := getAgentModelAdmin(rt)
	if err != nil {
		return nil, err
	}
	name, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("argument 'name' must be a string")
	}
	return admin.Delete(types.AgentModelName(name)), nil
}

func toolListAgentModels(rt tool.Runtime, args []interface{}) (interface{}, error) {
	reader, err := getAgentModelReader(rt)
	if err != nil {
		return nil, err
	}
	names := reader.List()
	// Convert to []string for NeuroScript
	stringNames := make([]string, len(names))
	for i, n := range names {
		stringNames[i] = string(n)
	}
	return stringNames, nil
}

func toolGetAgentModel(rt tool.Runtime, args []interface{}) (interface{}, error) {
	reader, err := getAgentModelReader(rt)
	if err != nil {
		return nil, err
	}
	name, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("argument 'name' must be a string")
	}
	model, found := reader.Get(types.AgentModelName(name))
	if !found {
		return lang.NewMapValue(nil), nil // Return nil if not found
	}

	// Convert struct to map via JSON, then wrap for NeuroScript
	data, err := json.Marshal(model)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal agent model struct: %w", err)
	}
	var modelMap map[string]any
	if err := json.Unmarshal(data, &modelMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal agent model to map: %w", err)
	}

	return lang.Wrap(modelMap)
}

func toolAgentModelExists(rt tool.Runtime, args []interface{}) (interface{}, error) {
	reader, err := getAgentModelReader(rt)
	if err != nil {
		return nil, err
	}
	name, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("argument 'name' must be a string")
	}
	_, found := reader.Get(types.AgentModelName(name))
	return found, nil
}

func toolSelectAgentModel(rt tool.Runtime, args []interface{}) (interface{}, error) {
	reader, err := getAgentModelReader(rt)
	if err != nil {
		return nil, err
	}
	name := ""
	if len(args) > 0 && args[0] != nil {
		s, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("argument 'name' must be a string")
		}
		name = s
	}

	if name != "" {
		// If a name is provided, check if it exists.
		if _, found := reader.Get(types.AgentModelName(name)); !found {
			return nil, lang.ErrNotFound
		}
		return name, nil
	}

	// If no name is provided, return the first model alphabetically.
	names := reader.List()
	if len(names) == 0 {
		return nil, lang.ErrNotFound // No models are registered.
	}
	sort.Slice(names, func(i, j int) bool {
		return names[i] < names[j]
	})
	return string(names[0]), nil
}
