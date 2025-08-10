// NeuroScript Version: 0.6.0
// File version: 8
// Purpose: Provides tools for managing AgentModels. Get tool now redacts the API key.
// filename: pkg/tool/agentmodel/tools.go
// nlines: 168
// risk_rating: MEDIUM

package agentmodel

import (
	"errors"
	"fmt"
	"sort"

	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// agentModelToolsToRegister holds the definitions for all tools in this package.
var agentModelToolsToRegister = []tool.ToolImplementation{
	{
		Spec: tool.ToolSpec{
			Name:        "Register",
			Group:       "agentmodel",
			Description: "Registers a new AgentModel configuration. Fails if the name already exists.",
			Category:    "Configuration",
			Args: []tool.ArgSpec{
				{Name: "name", Type: tool.ArgTypeString, Required: true, Description: "The unique name for the AgentModel."},
				{Name: "config", Type: tool.ArgTypeMap, Required: true, Description: "A map containing the model's configuration."},
			},
			ReturnType: tool.ArgTypeBool,
		},
		Func: registerAgentModel,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Update",
			Group:       "agentmodel",
			Description: "Updates an existing AgentModel's configuration with new values.",
			Category:    "Configuration",
			Args: []tool.ArgSpec{
				{Name: "name", Type: tool.ArgTypeString, Required: true, Description: "The name of the AgentModel to update."},
				{Name: "updates", Type: tool.ArgTypeMap, Required: true, Description: "A map of configuration keys and values to update."},
			},
			ReturnType: tool.ArgTypeBool,
		},
		Func: updateAgentModel,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Delete",
			Group:       "agentmodel",
			Description: "Deletes a registered AgentModel.",
			Category:    "Configuration",
			Args: []tool.ArgSpec{
				{Name: "name", Type: tool.ArgTypeString, Required: true, Description: "The name of the AgentModel to delete."},
			},
			ReturnType: tool.ArgTypeBool,
		},
		Func: deleteAgentModel,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "List",
			Group:       "agentmodel",
			Description: "Lists the names of all registered AgentModels in alphabetical order.",
			Category:    "Configuration",
			ReturnType:  "list",
		},
		Func: listAgentModels,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Get",
			Group:       "agentmodel",
			Description: "Gets the configuration of a specific registered AgentModel.",
			Category:    "Configuration",
			Args: []tool.ArgSpec{
				{Name: "name", Type: tool.ArgTypeString, Required: true, Description: "The name of the AgentModel to get."},
			},
			ReturnType: tool.ArgTypeMap,
		},
		Func: getAgentModel,
	},
}

// init() runs when this package is imported, adding the toolset to the global registration list.
func init() {
	tool.AddToolsetRegistration(
		"agentmodel",
		tool.CreateRegistrationFunc("agentmodel", agentModelToolsToRegister),
	)
}

// getInterpreter is a helper to safely access the underlying interpreter from the runtime.
func getInterpreter(rt tool.Runtime) (*interpreter.Interpreter, error) {
	interp, ok := rt.(*interpreter.Interpreter)
	if !ok {
		return nil, errors.New("internal tool error: agentmodel tools require the full interpreter runtime")
	}
	return interp, nil
}

// toLangValueMap converts a map[string]interface{} from the tool system to the map[string]lang.Value
// expected by the interpreter's internal methods, using the Wrap function as per the contract.
func toLangValueMap(data map[string]interface{}) (map[string]lang.Value, error) {
	result := make(map[string]lang.Value, len(data))
	for k, v := range data {
		wrappedVal, err := lang.Wrap(v)
		if err != nil {
			return nil, fmt.Errorf("could not wrap value for key '%s': %w", k, err)
		}
		result[k] = wrappedVal
	}
	return result, nil
}

// --- Tool Implementations ---

func registerAgentModel(rt tool.Runtime, args []interface{}) (interface{}, error) {
	name := args[0].(string)
	configData := args[1].(map[string]interface{})

	config, err := toLangValueMap(configData)
	if err != nil {
		return nil, fmt.Errorf("invalid config map: %w", err)
	}

	interp, err := getInterpreter(rt)
	if err != nil {
		return nil, err
	}

	err = interp.RegisterAgentModel(name, config)
	return err == nil, err
}

func updateAgentModel(rt tool.Runtime, args []interface{}) (interface{}, error) {
	name := args[0].(string)
	updatesData := args[1].(map[string]interface{})

	updates, err := toLangValueMap(updatesData)
	if err != nil {
		return nil, fmt.Errorf("invalid updates map: %w", err)
	}

	interp, err := getInterpreter(rt)
	if err != nil {
		return nil, err
	}

	err = interp.UpdateAgentModel(name, updates)
	return err == nil, err
}

func deleteAgentModel(rt tool.Runtime, args []interface{}) (interface{}, error) {
	name := args[0].(string)
	interp, err := getInterpreter(rt)
	if err != nil {
		return nil, err
	}
	wasDeleted := interp.DeleteAgentModel(name)
	return wasDeleted, nil
}

func listAgentModels(rt tool.Runtime, args []interface{}) (interface{}, error) {
	interp, err := getInterpreter(rt)
	if err != nil {
		return nil, err
	}
	names := interp.ListAgentModels()
	sort.Strings(names)

	vals := make([]interface{}, len(names))
	for i, name := range names {
		vals[i] = name
	}
	return vals, nil
}

func getAgentModel(rt tool.Runtime, args []interface{}) (interface{}, error) {
	name := args[0].(string)
	interp, err := getInterpreter(rt)
	if err != nil {
		return nil, err
	}

	model, found := interp.GetAgentModel(name)
	if !found {
		return nil, fmt.Errorf("AgentModel '%s' not found", name)
	}

	// Convert AgentModel struct to a map[string]interface{} for the script.
	configMap := map[string]interface{}{
		"name":        model.Name,
		"provider":    model.Provider,
		"model":       model.Model,
		"api_key":     "<secret api key>", // API key is redacted for security.
		"base_url":    model.BaseURL,
		"temperature": model.Temperature,
	}
	// PriceTable is map[string]float64, which is compatible with interface{}
	if model.PriceTable != nil {
		configMap["price_table"] = model.PriceTable
	}

	return configMap, nil
}
