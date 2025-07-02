// NeuroScript Version: 0.4.1
// File version: 1
// Purpose: Refactored all tool funcs to remove ValidateAndConvertArgs and use direct args from bridge.
// AI Worker Management: Instance Management Tools
// filename: pkg/core/ai_wm_tools_instances.go
// nlines: 212

package core

import (
	"time"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/sideline/nspatch"
	"github.com/google/uuid"
)

var specAIWorkerInstanceSpawn = tool.ToolSpec{
	Name:        "AIWorkerInstance.Spawn",
	Description: "Spawns a new AI Worker Instance and returns its details including a ConversationManager handle.",
	Category:    "AI Worker Management",
	Args: []tool.ArgSpec{
		{Name: "definition_id", Type: tool.ArgTypeString, Required: true, Description: "ID of the AIWorkerDefinition to use for spawning."},
		{Name: "config_overrides", Type: tool.ArgTypeMap, Required: false, Description: "Optional map of configuration overrides for this instance."},
		{Name: "file_contexts", Type: tool.ArgTypeSliceString, Required: false, Description: "Optional list of file context URIs for this instance."},
	},
	ReturnType: "map",
}

var toolAIWorkerInstanceSpawn = tool.ToolImplementation{
	Spec: specAIWorkerInstanceSpawn,
	Func: func(i *neurogo.Interpreter, args []interface{}) (interface{}, error) {
		m, err := getAIWorkerManager(i)
		if err != nil {
			return nil, err
		}

		defID, _ := args[0].(string)

		var overrides map[string]interface{}
		if args[1] != nil {
			overrides, _ = args[1].(map[string]interface{})
		}

		var fileContexts []string
		if args[2] != nil {
			fileContexts, _ = args[2].([]string)
		}

		instance, spawnErr := m.SpawnWorkerInstance(defID, overrides, fileContexts)
		if spawnErr != nil {
			return nil, spawnErr
		}
		if instance == nil {
			return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "SpawnWorkerInstance returned nil instance without error", nspatch.ErrInternal)
		}

		convoManager := llm.NewConversationManager(i.Logger())
		handleID, handleErr := i.RegisterHandle(convoManager, "ConversationManager-"+uuid.NewString())
		if handleErr != nil {
			m.logger.Errorf("Failed to register ConversationManager handle for instance %s: %v. Retiring instance.", instance.InstanceID, handleErr)
			_ = m.RetireWorkerInstance(instance.InstanceID, "Failed to register handle", AIWorkerInstanceStatus("error"), TokenUsageMetrics{}, nil)
			return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "failed to register ConversationManager handle", handleErr)
		}

		instanceMap := convertAIWorkerInstanceToMap(instance)
		instanceMap["conversation_manager_handle"] = handleID
		m.logger.Infof("Instance %s spawned, ConversationManager handle %s registered.", instance.InstanceID, handleID)
		return instanceMap, nil
	},
}

var specAIWorkerInstanceGet = tool.ToolSpec{
	Name:        "AIWorkerInstance.Get",
	Description: "Retrieves an active AI Worker Instance's details by its ID.",
	Category:    "AI Worker Management",
	Args:        []tool.ArgSpec{{Name: "instance_id", Type: tool.ArgTypeString, Required: true, Description: "The unique ID of the active instance to retrieve."}},
	ReturnType:  "map",
}

var toolAIWorkerInstanceGet = tool.ToolImplementation{
	Spec: specAIWorkerInstanceGet,
	Func: func(i *neurogo.Interpreter, args []interface{}) (interface{}, error) {
		m, err := getAIWorkerManager(i)
		if err != nil {
			return nil, err
		}
		id, _ := args[0].(string)
		instance, instanceErr := m.GetWorkerInstance(id)
		if instanceErr != nil {
			return nil, instanceErr
		}
		return convertAIWorkerInstanceToMap(instance), nil
	},
}

var specAIWorkerInstanceListActive = tool.ToolSpec{
	Name:        "AIWorkerInstance.ListActive",
	Description: "Lists currently active AI Worker Instances, optionally filtered.",
	Category:    "AI Worker Management",
	Args:        []tool.ArgSpec{{Name: "filters", Type: tool.ArgTypeMap, Required: false, Description: "Optional map of filters."}},
	ReturnType:  "slice",
}

var toolAIWorkerInstanceListActive = tool.ToolImplementation{
	Spec: specAIWorkerInstanceListActive,
	Func: func(i *neurogo.Interpreter, args []interface{}) (interface{}, error) {
		m, err := getAIWorkerManager(i)
		if err != nil {
			return nil, err
		}
		var filters map[string]interface{}
		if args[0] != nil {
			filters, _ = args[0].(map[string]interface{})
		}
		instances := m.ListActiveWorkerInstances(filters)
		resultList := make([]interface{}, len(instances))
		for idx, inst := range instances {
			resultList[idx] = convertAIWorkerInstanceToMap(inst)
		}
		return resultList, nil
	},
}

var specAIWorkerInstanceRetire = tool.ToolSpec{
	Name:        "AIWorkerInstance.Retire",
	Description: "Retires an active AI Worker Instance.",
	Category:    "AI Worker Management",
	Args: []tool.ArgSpec{
		{Name: "instance_id", Type: tool.ArgTypeString, Required: true},
		{Name: "conversation_manager_handle", Type: tool.ArgTypeString, Required: true},
		{Name: "reason", Type: tool.ArgTypeString, Required: true},
		{Name: "final_status", Type: tool.ArgTypeString, Required: true},
		{Name: "final_session_token_usage", Type: tool.ArgTypeMap, Required: true},
		{Name: "performance_records", Type: tool.ArgTypeSliceMap, Required: false},
	},
	ReturnType: "nil",
}

var toolAIWorkerInstanceRetire = tool.ToolImplementation{
	Spec: specAIWorkerInstanceRetire,
	Func: func(i *neurogo.Interpreter, args []interface{}) (interface{}, error) {
		m, err := getAIWorkerManager(i)
		if err != nil {
			return nil, err
		}
		instanceID, _ := args[0].(string)
		handleID, _ := args[1].(string)
		reason, _ := args[2].(string)
		finalStatusStr, _ := args[3].(string)
		usageMap, _ := args[4].(map[string]interface{})

		finalUsage := TokenUsageMetrics{}
		if v, ok := lang.toInt64(usageMap["input_tokens"]); ok {
			finalUsage.InputTokens = v
		}
		if v, ok := lang.toInt64(usageMap["output_tokens"]); ok {
			finalUsage.OutputTokens = v
		}
		finalUsage.TotalTokens = finalUsage.InputTokens + finalUsage.OutputTokens

		var perfRecords []*PerformanceRecord
		if args[5] != nil {
			if prList, ok := args[5].([]map[string]interface{}); ok {
				for _, prMap := range prList {
					// This parsing logic could be a separate helper function
					pr := PerformanceRecord{InstanceID: instanceID}
					pr.TaskID, _ = prMap["task_id"].(string)
					pr.DefinitionID, _ = prMap["definition_id"].(string)
					if tsStartStr, tsOk := prMap["timestamp_start"].(string); tsOk {
						pr.TimestampStart, _ = time.Parse(time.RFC3339Nano, tsStartStr)
					}
					if tsEndStr, tsOk := prMap["timestamp_end"].(string); tsOk {
						pr.TimestampEnd, _ = time.Parse(time.RFC3339Nano, tsEndStr)
					}
					perfRecords = append(perfRecords, &pr)
				}
			}
		}

		retireErr := m.RetireWorkerInstance(instanceID, reason, AIWorkerInstanceStatus(finalStatusStr), finalUsage, perfRecords)
		if retireErr != nil {
			return nil, retireErr
		}
		if !i.RemoveHandle(handleID) {
			i.Logger().Warnf("Failed to remove ConversationManager handle %s for retired instance %s", handleID, instanceID)
		}
		return nil, nil
	},
}

var specAIWorkerInstanceUpdateStatus = tool.ToolSpec{
	Name:        "AIWorkerInstance.UpdateStatus",
	Description: "Updates the status of an active AI Worker Instance.",
	Category:    "AI Worker Management",
	Args: []tool.ArgSpec{
		{Name: "instance_id", Type: tool.ArgTypeString, Required: true},
		{Name: "status", Type: tool.ArgTypeString, Required: true},
		{Name: "last_error", Type: tool.ArgTypeString, Required: false},
	},
	ReturnType: "nil",
}

var toolAIWorkerInstanceUpdateStatus = tool.ToolImplementation{
	Spec: specAIWorkerInstanceUpdateStatus,
	Func: func(i *neurogo.Interpreter, args []interface{}) (interface{}, error) {
		m, err := getAIWorkerManager(i)
		if err != nil {
			return nil, err
		}
		instanceID, _ := args[0].(string)
		statusStr, _ := args[1].(string)
		var lastError string
		if args[2] != nil {
			lastError, _ = args[2].(string)
		}
		return nil, m.UpdateInstanceStatus(instanceID, AIWorkerInstanceStatus(statusStr), lastError)
	},
}

var specAIWorkerInstanceUpdateTokenUsage = tool.ToolSpec{
	Name:        "AIWorkerInstance.UpdateTokenUsage",
	Description: "Updates the session token usage for an active AI Worker Instance.",
	Category:    "AI Worker Management",
	Args: []tool.ArgSpec{
		{Name: "instance_id", Type: tool.ArgTypeString, Required: true},
		{Name: "input_tokens", Type: tool.ArgTypeInt, Required: true},
		{Name: "output_tokens", Type: tool.ArgTypeInt, Required: true},
	},
	ReturnType: "nil",
}

var toolAIWorkerInstanceUpdateTokenUsage = tool.ToolImplementation{
	Spec: specAIWorkerInstanceUpdateTokenUsage,
	Func: func(i *neurogo.Interpreter, args []interface{}) (interface{}, error) {
		m, err := getAIWorkerManager(i)
		if err != nil {
			return nil, err
		}
		instanceID, _ := args[0].(string)
		inputTokens, _ := lang.toInt64(args[1])
		outputTokens, _ := lang.toInt64(args[2])
		return nil, m.UpdateInstanceSessionTokenUsage(instanceID, inputTokens, outputTokens)
	},
}
