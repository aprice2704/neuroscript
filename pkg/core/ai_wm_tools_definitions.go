// NeuroScript Version: 0.3.0
// File version: 0.1.0
// AI Worker Management: Definition Management Tools
// filename: pkg/core/ai_wm_tools_definitions.go

package core

import (
	"fmt"
	// "time" // Not directly needed here, but helpers might use it
	// "github.com/google/uuid" // Not directly needed here
)

var specAIWorkerDefinitionAdd = ToolSpec{
	Name:        "AIWorkerDefinition.Add",
	Description: "Adds a new AI Worker Definition. Maps (base_config, etc.) are optional.",
	Args: []ArgSpec{
		{Name: "definition_id", Type: ArgTypeString, Required: false}, {Name: "name", Type: ArgTypeString, Required: false},
		{Name: "provider", Type: ArgTypeString, Required: true}, {Name: "model_name", Type: ArgTypeString, Required: true},
		{Name: "auth", Type: ArgTypeMap, Required: true}, {Name: "interaction_models", Type: ArgTypeSliceString, Required: false},
		{Name: "capabilities", Type: ArgTypeSliceString, Required: false}, {Name: "base_config", Type: ArgTypeMap, Required: false},
		{Name: "cost_metrics", Type: ArgTypeMap, Required: false}, {Name: "rate_limits", Type: ArgTypeMap, Required: false},
		{Name: "status", Type: ArgTypeString, Required: false}, {Name: "default_file_contexts", Type: ArgTypeSliceString, Required: false},
		{Name: "metadata", Type: ArgTypeMap, Required: false},
	},
	ReturnType: ArgTypeString,
}

var toolAIWorkerDefinitionAdd = ToolImplementation{
	Spec: specAIWorkerDefinitionAdd, // Keep spec here for clarity with its implementation
	Func: func(i *Interpreter, argsGiven []interface{}) (interface{}, error) {
		m, err := getAIWorkerManager(i)
		if err != nil {
			return nil, err
		}
		// Corrected: Pass the whole spec to ValidateAndConvertArgs
		validatedArgsList, valErr := ValidateAndConvertArgs(specAIWorkerDefinitionAdd, argsGiven)
		if valErr != nil {
			return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("Validation failed for tool %s: %s", specAIWorkerDefinitionAdd.Name, valErr.Error()), ErrInvalidArgument)
		}
		parsedArgs := mapValidatedArgsListToMapByName(specAIWorkerDefinitionAdd.Args, validatedArgsList)

		def := AIWorkerDefinition{} // AIWorkerDefinition is a struct

		if idVal, ok := parsedArgs["definition_id"].(string); ok {
			def.DefinitionID = idVal
		}
		if nameVal, ok := parsedArgs["name"].(string); ok {
			def.Name = nameVal
		}

		providerStr, ok := parsedArgs["provider"].(string)
		if !ok || providerStr == "" {
			return nil, NewRuntimeError(ErrorCodeArgMismatch, "provider is required for tool "+specAIWorkerDefinitionAdd.Name, ErrInvalidArgument)
		}
		def.Provider = AIWorkerProvider(providerStr)

		modelNameStr, ok := parsedArgs["model_name"].(string)
		if !ok || modelNameStr == "" {
			return nil, NewRuntimeError(ErrorCodeArgMismatch, "model_name is required for tool "+specAIWorkerDefinitionAdd.Name, ErrInvalidArgument)
		}
		def.ModelName = modelNameStr

		authMap, ok := parsedArgs["auth"].(map[string]interface{})
		if !ok {
			return nil, NewRuntimeError(ErrorCodeArgMismatch, "auth map is required for tool "+specAIWorkerDefinitionAdd.Name, ErrInvalidArgument)
		}
		authMethodStr, _ := authMap["method"].(string)
		authValueStr, _ := authMap["value"].(string)
		def.Auth = APIKeySource{Method: APIKeySourceMethod(authMethodStr), Value: authValueStr}

		if imListArg, okGet := parsedArgs["interaction_models"]; okGet && imListArg != nil {
			if imList, listOk := imListArg.([]interface{}); listOk {
				for _, im := range imList {
					if imStr, sOk := im.(string); sOk {
						def.InteractionModels = append(def.InteractionModels, InteractionModelType(imStr))
					}
				}
			}
		}
		if capListArg, okGet := parsedArgs["capabilities"]; okGet && capListArg != nil {
			if capList, listOk := capListArg.([]interface{}); listOk {
				for _, c := range capList {
					if cStr, sOk := c.(string); sOk {
						def.Capabilities = append(def.Capabilities, cStr)
					}
				}
			}
		}

		if bc, okGet := parsedArgs["base_config"].(map[string]interface{}); okGet && bc != nil {
			def.BaseConfig = bc
		} else {
			def.BaseConfig = make(map[string]interface{}) // Initialize to empty map if not provided
		}

		if cmArg, okGet := parsedArgs["cost_metrics"]; okGet && cmArg != nil {
			if cm, mapOk := cmArg.(map[string]interface{}); mapOk {
				def.CostMetrics = make(map[string]float64)
				for k, v := range cm {
					if fv, fOk := toFloat64(v); fOk { // Ensure toFloat64 handles various numeric types
						def.CostMetrics[k] = fv
					} else {
						i.Logger().Warnf("Non-float or unconvertible cost_metric '%s' with value '%v' in %s. Skipping.", k, v, specAIWorkerDefinitionAdd.Name)
					}
				}
			}
		} else {
			def.CostMetrics = make(map[string]float64) // Initialize to empty map
		}

		if rlMapArg, okGet := parsedArgs["rate_limits"]; okGet && rlMapArg != nil {
			if rlMap, mapOk := rlMapArg.(map[string]interface{}); mapOk {
				// RateLimits is a struct, initialize it before setting fields
				def.RateLimits = RateLimitPolicy{}
				if v, fOk := toInt64(rlMap["max_requests_per_minute"]); fOk {
					def.RateLimits.MaxRequestsPerMinute = int(v)
				}
				if v, fOk := toInt64(rlMap["max_tokens_per_minute"]); fOk {
					def.RateLimits.MaxTokensPerMinute = int(v)
				}
				if v, fOk := toInt64(rlMap["max_tokens_per_day"]); fOk {
					def.RateLimits.MaxTokensPerDay = int(v)
				}
				if v, fOk := toInt64(rlMap["max_concurrent_active_instances"]); fOk {
					def.RateLimits.MaxConcurrentActiveInstances = int(v)
				}
			}
		} // else RateLimits will be zero-value struct, which is fine

		if statusStr, okGet := parsedArgs["status"].(string); okGet && statusStr != "" {
			def.Status = AIWorkerDefinitionStatus(statusStr)
		} // Default status is handled by AddWorkerDefinition if empty

		if dfcListArg, okGet := parsedArgs["default_file_contexts"]; okGet && dfcListArg != nil {
			if dfcList, listOk := dfcListArg.([]interface{}); listOk {
				for _, item := range dfcList {
					if s, sOk := item.(string); sOk {
						def.DefaultFileContexts = append(def.DefaultFileContexts, s)
					}
				}
			}
		}
		if md, okGet := parsedArgs["metadata"].(map[string]interface{}); okGet && md != nil {
			def.Metadata = md
		} else {
			def.Metadata = make(map[string]interface{}) // Initialize to empty map
		}

		// Ensure AggregatePerformanceSummary is initialized as it's a pointer
		if def.AggregatePerformanceSummary == nil {
			def.AggregatePerformanceSummary = &AIWorkerPerformanceSummary{}
		}

		id, addErr := m.AddWorkerDefinition(def) // AddWorkerDefinition now takes AIWorkerDefinition by value
		if addErr != nil {
			return nil, addErr
		}
		return id, nil
	},
}

var specAIWorkerDefinitionGet = ToolSpec{
	Name:        "AIWorkerDefinition.Get",
	Description: "Retrieves an AI Worker Definition by its ID.",
	Args:        []ArgSpec{{Name: "definition_id", Type: ArgTypeString, Required: true}},
	ReturnType:  ArgTypeMap,
}
var toolAIWorkerDefinitionGet = ToolImplementation{
	Spec: specAIWorkerDefinitionGet,
	Func: func(i *Interpreter, argsGiven []interface{}) (interface{}, error) {
		m, err := getAIWorkerManager(i)
		if err != nil {
			return nil, err
		}
		validatedArgsList, valErr := ValidateAndConvertArgs(specAIWorkerDefinitionGet, argsGiven)
		if valErr != nil {
			return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("Validation failed for tool %s: %s", specAIWorkerDefinitionGet.Name, valErr.Error()), ErrInvalidArgument)
		}
		parsedArgs := mapValidatedArgsListToMapByName(specAIWorkerDefinitionGet.Args, validatedArgsList)
		id, _ := parsedArgs["definition_id"].(string)

		def, getErr := m.GetWorkerDefinition(id)
		if getErr != nil {
			return nil, getErr
		}
		return convertAIWorkerDefinitionToMap(def), nil
	},
}

var specAIWorkerDefinitionList = ToolSpec{
	Name:        "AIWorkerDefinition.List",
	Description: "Lists all AI Worker Definitions, optionally filtered.",
	Args:        []ArgSpec{{Name: "filters", Type: ArgTypeMap, Required: false}},
	ReturnType:  ArgTypeSliceMap,
}
var toolAIWorkerDefinitionList = ToolImplementation{
	Spec: specAIWorkerDefinitionList,
	Func: func(i *Interpreter, argsGiven []interface{}) (interface{}, error) {
		m, err := getAIWorkerManager(i)
		if err != nil {
			return nil, err
		}
		validatedArgsList, valErr := ValidateAndConvertArgs(specAIWorkerDefinitionList, argsGiven)
		if valErr != nil {
			return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("Validation failed for tool %s: %s", specAIWorkerDefinitionList.Name, valErr.Error()), ErrInvalidArgument)
		}
		parsedArgs := mapValidatedArgsListToMapByName(specAIWorkerDefinitionList.Args, validatedArgsList)
		filters, _ := parsedArgs["filters"].(map[string]interface{})

		defs := m.ListWorkerDefinitions(filters)
		resultList := make([]interface{}, len(defs))
		for idx, def := range defs {
			resultList[idx] = convertAIWorkerDefinitionToMap(def)
		}
		return resultList, nil
	},
}

var specAIWorkerDefinitionUpdate = ToolSpec{
	Name:        "AIWorkerDefinition.Update",
	Description: "Updates fields of an existing AI Worker Definition.",
	Args: []ArgSpec{
		{Name: "definition_id", Type: ArgTypeString, Required: true},
		{Name: "updates", Type: ArgTypeMap, Required: true},
	},
	ReturnType: ArgTypeNil,
}
var toolAIWorkerDefinitionUpdate = ToolImplementation{
	Spec: specAIWorkerDefinitionUpdate,
	Func: func(i *Interpreter, argsGiven []interface{}) (interface{}, error) {
		m, err := getAIWorkerManager(i)
		if err != nil {
			return nil, err
		}
		validatedArgsList, valErr := ValidateAndConvertArgs(specAIWorkerDefinitionUpdate, argsGiven)
		if valErr != nil {
			return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("Validation failed for tool %s: %s", specAIWorkerDefinitionUpdate.Name, valErr.Error()), ErrInvalidArgument)
		}
		parsedArgs := mapValidatedArgsListToMapByName(specAIWorkerDefinitionUpdate.Args, validatedArgsList)
		id, _ := parsedArgs["definition_id"].(string)
		updates, _ := parsedArgs["updates"].(map[string]interface{})

		updateErr := m.UpdateWorkerDefinition(id, updates)
		if updateErr != nil {
			return nil, updateErr
		}
		return nil, nil
	},
}

var specAIWorkerDefinitionRemove = ToolSpec{
	Name:        "AIWorkerDefinition.Remove",
	Description: "Removes an AI Worker Definition if it has no active instances.",
	Args:        []ArgSpec{{Name: "definition_id", Type: ArgTypeString, Required: true}},
	ReturnType:  ArgTypeNil,
}
var toolAIWorkerDefinitionRemove = ToolImplementation{
	Spec: specAIWorkerDefinitionRemove,
	Func: func(i *Interpreter, argsGiven []interface{}) (interface{}, error) {
		m, err := getAIWorkerManager(i)
		if err != nil {
			return nil, err
		}
		validatedArgsList, valErr := ValidateAndConvertArgs(specAIWorkerDefinitionRemove, argsGiven)
		if valErr != nil {
			return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("Validation failed for tool %s: %s", specAIWorkerDefinitionRemove.Name, valErr.Error()), ErrInvalidArgument)
		}
		parsedArgs := mapValidatedArgsListToMapByName(specAIWorkerDefinitionRemove.Args, validatedArgsList)
		id, _ := parsedArgs["definition_id"].(string)

		removeErr := m.RemoveWorkerDefinition(id)
		if removeErr != nil {
			return nil, removeErr
		}
		return nil, nil
	},
}
