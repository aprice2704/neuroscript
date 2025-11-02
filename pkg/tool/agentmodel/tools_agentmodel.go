// NeuroScript Version: 0.7.3
// File version: 10
// Purpose: Aligned all admin/reader calls to use plain 'string' for model names, fixing compiler errors.
// filename: pkg/tool/agentmodel/tools_agentmodel.go
// nlines: 171
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
	// FIX: Pass 'name' (string) directly, not types.AgentModelName(name)
	err = admin.Register(name, config)
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
	// FIX: Pass 'name' (string) directly, not types.AgentModelName(name)
	err = admin.Update(name, updates)
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
	// FIX: Pass 'name' (string) directly, not types.AgentModelName(name)
	return admin.Delete(name), nil
}

func toolListAgentModels(rt tool.Runtime, args []interface{}) (interface{}, error) {
	reader, err := getAgentModelReader(rt)
	if err != nil {
		return nil, err
	}
	// FIX: reader.List() now correctly returns []string, so no conversion is needed.
	return reader.List(), nil
}

func toolGetAgentModel(rt tool.Runtime, args []interface{}) (interface{}, error) {
	// fmt.Fprintf(os.Stderr, "DEBUG: Entered toolGetAgentModel\n")
	reader, err := getAgentModelReader(rt)
	if err != nil {
		// fmt.Fprintf(os.Stderr, "DEBUG: toolGetAgentModel: Failed to get reader: %v\n", err)
		return nil, err
	}
	name, ok := args[0].(string)
	if !ok {
		// fmt.Fprintf(os.Stderr, "DEBUG: toolGetAgentModel: Argument 'name' is not a string (got %T)\n", args[0])
		return nil, fmt.Errorf("argument 'name' must be a string")
	}
	// fmt.Fprintf(os.Stderr, "DEBUG: toolGetAgentModel: Attempting to get model '%s'\n", name)
	// FIX: Pass 'name' (string) directly, not types.AgentModelName(name)
	model, found := reader.Get(name)
	if !found {
		// fmt.Fprintf(os.Stderr, "DEBUG: toolGetAgentModel: Model '%s' not found, returning nil map\n", name)
		return lang.NewMapValue(nil), nil // Return nil if not found
	}

	// Convert struct to map via JSON, then wrap for NeuroScript
	data, err := json.Marshal(model)
	if err != nil {
		// fmt.Fprintf(os.Stderr, "DEBUG: toolGetAgentModel: Failed to marshal agent model struct: %v\n", err)
		return nil, fmt.Errorf("failed to marshal agent model struct: %w", err)
	}
	var modelMap map[string]any
	if err := json.Unmarshal(data, &modelMap); err != nil {
		// fmt.Fprintf(os.Stderr, "DEBUG: toolGetAgentModel: Failed to unmarshal agent model to map: %v\n", err)
		return nil, fmt.Errorf("failed to unmarshal agent model to map: %w", err)
	}

	// fmt.Fprintf(os.Stderr, "DEBUG: toolGetAgentModel: Wrapping and returning model map. Keys: %d\n", len(modelMap))
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
	// FIX: Pass 'name' (string) directly, not types.AgentModelName(name)
	_, found := reader.Get(name)
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
		// FIX: Pass 'name' (string) directly, not types.AgentModelName(name)
		if _, found := reader.Get(name); !found {
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
	// FIX: list() now returns []string, so just return the element.
	return names[0], nil
}
