// NeuroScript Version: 0.4.1
// File version: 2
// Purpose: Removed Add, Update, and Remove tools as their underlying manager methods were removed (definitions are immutable post-load).
// AI Worker Management: Definition Management Tools
// filename: pkg/core/ai_wm_tools_definitions.go
// nlines: 65

package core

var specAIWorkerDefinitionGet = ToolSpec{
	Name:            "AIWorkerDefinition.Get",
	Description:     "Retrieves an AI Worker Definition by its ID.",
	Category:        "AI Worker Management",
	Args:            []ArgSpec{{Name: "definition_id", Type: ArgTypeString, Required: true, Description: "The unique ID of the definition to retrieve."}},
	ReturnType:      "map",
	ReturnHelp:      "Returns a map representing the AIWorkerDefinition struct. Returns nil if not found or on error.",
	Example:         `TOOL.AIWorkerDefinition.Get(definition_id: "google-gemini-1.5-pro")`,
	ErrorConditions: "ErrAIWorkerManagerMissing; ErrInvalidArgument if definition_id is not provided or not a string; ErrDefinitionNotFound if definition with ID does not exist.",
}
var toolAIWorkerDefinitionGet = ToolImplementation{
	Spec: specAIWorkerDefinitionGet,
	Func: func(i *Interpreter, args []interface{}) (interface{}, error) {
		m, err := getAIWorkerManager(i)
		if err != nil {
			return nil, err
		}
		id, _ := args[0].(string)

		def, getErr := m.GetWorkerDefinition(id)
		if getErr != nil {
			return nil, getErr
		}
		return convertAIWorkerDefinitionToMap(def), nil
	},
}

var specAIWorkerDefinitionList = ToolSpec{
	Name:            "AIWorkerDefinition.List",
	Description:     "Lists all AI Worker Definitions, optionally filtered.",
	Category:        "AI Worker Management",
	Args:            []ArgSpec{{Name: "filters", Type: ArgTypeMap, Required: false, Description: "Optional map of filters (e.g., {'provider':'google', 'status':'active'})."}},
	ReturnType:      "slice",
	ReturnHelp:      "Returns a slice of maps, where each map represents an AIWorkerDefinition. Returns an empty slice if no definitions match or exist.",
	Example:         `TOOL.AIWorkerDefinition.List(filters: {"provider":"google"})`,
	ErrorConditions: "ErrAIWorkerManagerMissing; ErrInvalidArgument if filters is not a map.",
}
var toolAIWorkerDefinitionList = ToolImplementation{
	Spec: specAIWorkerDefinitionList,
	Func: func(i *Interpreter, args []interface{}) (interface{}, error) {
		m, err := getAIWorkerManager(i)
		if err != nil {
			return nil, err
		}
		var filters map[string]interface{}
		if len(args) > 0 && args[0] != nil {
			filters, _ = args[0].(map[string]interface{})
		}

		defs := m.ListWorkerDefinitions(filters)
		resultList := make([]interface{}, len(defs))
		for idx, def := range defs {
			resultList[idx] = convertAIWorkerDefinitionToMap(def)
		}
		return resultList, nil
	},
}

// Note: toolAIWorkerDefinitionAdd, toolAIWorkerDefinitionUpdate, and
// toolAIWorkerDefinitionRemove have been removed as the underlying
// AIWorkerManager methods were deleted to make definitions immutable post-load.
