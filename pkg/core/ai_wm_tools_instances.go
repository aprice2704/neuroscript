// NeuroScript Version: 0.3.0
// File version: 0.1.0
// AI Worker Management: Instance Management Tools
// filename: pkg/core/ai_wm_tools_instances.go

package core

import (
	"fmt"
	"time" // Required for parsing time in Retire tool

	"github.com/google/uuid" // Required for Spawn tool to create ConversationManager handle
)

var specAIWorkerInstanceSpawn = ToolSpec{
	Name: "AIWorkerInstance.Spawn", Description: "Spawns a new AI Worker Instance and returns its details including a ConversationManager handle.",
	Args: []ArgSpec{
		{Name: "definition_id", Type: ArgTypeString, Required: true},
		{Name: "config_overrides", Type: ArgTypeMap, Required: false},
		{Name: "file_contexts", Type: ArgTypeSliceString, Required: false},
	}, ReturnType: ArgTypeMap,
}

var toolAIWorkerInstanceSpawn = ToolImplementation{
	Spec: specAIWorkerInstanceSpawn,
	Func: func(i *Interpreter, argsGiven []interface{}) (interface{}, error) {
		m, err := getAIWorkerManager(i)
		if err != nil {
			return nil, err
		}
		validatedArgsList, valErr := ValidateAndConvertArgs(specAIWorkerInstanceSpawn, argsGiven)
		if valErr != nil {
			return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("Validation failed for tool %s: %s", specAIWorkerInstanceSpawn.Name, valErr.Error()), ErrInvalidArgument)
		}
		parsedArgs := mapValidatedArgsListToMapByName(specAIWorkerInstanceSpawn.Args, validatedArgsList)
		defID, _ := parsedArgs["definition_id"].(string)
		overrides, _ := parsedArgs["config_overrides"].(map[string]interface{})
		var fileContexts []string
		if fcListArg, okGet := parsedArgs["file_contexts"]; okGet && fcListArg != nil {
			if fcList, listOk := fcListArg.([]interface{}); listOk {
				for _, item := range fcList {
					if s, sOk := item.(string); sOk {
						fileContexts = append(fileContexts, s)
					}
				}
			}
		}
		instance, spawnErr := m.SpawnWorkerInstance(defID, overrides, fileContexts)
		if spawnErr != nil {
			return nil, spawnErr
		}
		if instance == nil {
			return nil, NewRuntimeError(ErrorCodeInternal, "SpawnWorkerInstance returned nil instance without error for tool "+specAIWorkerInstanceSpawn.Name, ErrInternal)
		}

		convoManager := NewConversationManager(i.Logger())
		handleID, handleErr := i.RegisterHandle(convoManager, "ConversationManager-"+uuid.NewString()) // Make handle unique
		if handleErr != nil {
			m.logger.Errorf("Failed to register ConversationManager handle for instance %s (tool %s): %v. Retiring instance.", instance.InstanceID, specAIWorkerInstanceSpawn.Name, handleErr)
			_ = m.RetireWorkerInstance(instance.InstanceID, "Failed to register ConversationManager handle", InstanceStatusRetiredError, TokenUsageMetrics{}, nil)
			if re, ok := handleErr.(*RuntimeError); ok {
				return nil, re
			}
			return nil, NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("failed to register ConversationManager handle for instance %s (tool %s)", instance.InstanceID, specAIWorkerInstanceSpawn.Name), handleErr)
		}

		instanceMap := convertAIWorkerInstanceToMap(instance)
		instanceMap["conversation_manager_handle"] = handleID
		m.logger.Infof("Instance %s spawned (tool %s), ConversationManager registered with handle %s", instance.InstanceID, specAIWorkerInstanceSpawn.Name, handleID)
		return instanceMap, nil
	},
}

var specAIWorkerInstanceGet = ToolSpec{
	Name:        "AIWorkerInstance.Get",
	Description: "Retrieves an active AI Worker Instance's details by its ID.",
	Args:        []ArgSpec{{Name: "instance_id", Type: ArgTypeString, Required: true}},
	ReturnType:  ArgTypeMap,
}

var toolAIWorkerInstanceGet = ToolImplementation{
	Spec: specAIWorkerInstanceGet,
	Func: func(i *Interpreter, argsGiven []interface{}) (interface{}, error) {
		m, managerErr := getAIWorkerManager(i)
		if managerErr != nil {
			return nil, managerErr
		}
		validatedArgsList, valErr := ValidateAndConvertArgs(specAIWorkerInstanceGet, argsGiven)
		if valErr != nil {
			return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("Validation failed for tool %s: %s", specAIWorkerInstanceGet.Name, valErr.Error()), ErrInvalidArgument)
		}
		parsedArgs := mapValidatedArgsListToMapByName(specAIWorkerInstanceGet.Args, validatedArgsList)
		id, _ := parsedArgs["instance_id"].(string)
		instance, instanceErr := m.GetWorkerInstance(id)
		if instanceErr != nil {
			return nil, instanceErr
		}
		return convertAIWorkerInstanceToMap(instance), nil
	},
}

var specAIWorkerInstanceListActive = ToolSpec{
	Name:        "AIWorkerInstance.ListActive",
	Description: "Lists currently active AI Worker Instances, optionally filtered.",
	Args:        []ArgSpec{{Name: "filters", Type: ArgTypeMap, Required: false}},
	ReturnType:  ArgTypeSliceMap,
}

var toolAIWorkerInstanceListActive = ToolImplementation{
	Spec: specAIWorkerInstanceListActive,
	Func: func(i *Interpreter, argsGiven []interface{}) (interface{}, error) {
		m, err := getAIWorkerManager(i)
		if err != nil {
			return nil, err
		}
		validatedArgsList, valErr := ValidateAndConvertArgs(specAIWorkerInstanceListActive, argsGiven)
		if valErr != nil {
			return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("Validation failed for tool %s: %s", specAIWorkerInstanceListActive.Name, valErr.Error()), ErrInvalidArgument)
		}
		parsedArgs := mapValidatedArgsListToMapByName(specAIWorkerInstanceListActive.Args, validatedArgsList)
		filters, _ := parsedArgs["filters"].(map[string]interface{})

		instances := m.ListActiveWorkerInstances(filters)
		resultList := make([]interface{}, len(instances))
		for idx, inst := range instances {
			resultList[idx] = convertAIWorkerInstanceToMap(inst)
		}
		return resultList, nil
	},
}

var specAIWorkerInstanceRetire = ToolSpec{
	Name:        "AIWorkerInstance.Retire",
	Description: "Retires an active AI Worker Instance, persisting its final state and performance.",
	Args: []ArgSpec{
		{Name: "instance_id", Type: ArgTypeString, Required: true},
		{Name: "conversation_manager_handle", Type: ArgTypeString, Required: true},
		{Name: "reason", Type: ArgTypeString, Required: true},
		{Name: "final_status", Type: ArgTypeString, Required: true},
		{Name: "final_session_token_usage", Type: ArgTypeMap, Required: true},
		{Name: "performance_records", Type: ArgTypeSliceMap, Required: false},
	},
	ReturnType: ArgTypeNil,
}

var toolAIWorkerInstanceRetire = ToolImplementation{
	Spec: specAIWorkerInstanceRetire,
	Func: func(i *Interpreter, argsGiven []interface{}) (interface{}, error) {
		m, err := getAIWorkerManager(i)
		if err != nil {
			return nil, err
		}
		validatedArgsList, valErr := ValidateAndConvertArgs(specAIWorkerInstanceRetire, argsGiven)
		if valErr != nil {
			return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("Validation failed for tool %s: %s", specAIWorkerInstanceRetire.Name, valErr.Error()), ErrInvalidArgument)
		}
		parsedArgs := mapValidatedArgsListToMapByName(specAIWorkerInstanceRetire.Args, validatedArgsList)

		instanceID, _ := parsedArgs["instance_id"].(string)
		handleID, _ := parsedArgs["conversation_manager_handle"].(string)
		reason, _ := parsedArgs["reason"].(string)
		finalStatusStr, _ := parsedArgs["final_status"].(string)
		finalStatus := AIWorkerInstanceStatus(finalStatusStr)

		usageMap, ok := parsedArgs["final_session_token_usage"].(map[string]interface{})
		if !ok {
			return nil, NewRuntimeError(ErrorCodeArgMismatch, "final_session_token_usage must be a map for tool "+specAIWorkerInstanceRetire.Name, ErrInvalidArgument)
		}
		finalUsage := TokenUsageMetrics{}
		if v, fOk := toInt64(usageMap["input_tokens"]); fOk {
			finalUsage.InputTokens = v
		}
		if v, fOk := toInt64(usageMap["output_tokens"]); fOk {
			finalUsage.OutputTokens = v
		}
		finalUsage.TotalTokens = finalUsage.InputTokens + finalUsage.OutputTokens

		var perfRecords []*PerformanceRecord
		if prListArg, okGet := parsedArgs["performance_records"].([]interface{}); okGet && prListArg != nil {
			for _, prMapArg := range prListArg {
				if prMap, mapOk := prMapArg.(map[string]interface{}); mapOk {
					pr := PerformanceRecord{}
					pr.TaskID, _ = prMap["task_id"].(string)
					pr.InstanceID = instanceID
					if defIDVal, defIDOk := prMap["definition_id"].(string); defIDOk {
						pr.DefinitionID = defIDVal
					} else {
						// Attempt to get DefID from the instance being retired
						inst, _ := m.GetWorkerInstance(instanceID) // RLock is fine here
						if inst != nil {
							pr.DefinitionID = inst.DefinitionID
						}
					}

					if tsStartStr, tsOk := prMap["timestamp_start"].(string); tsOk {
						pr.TimestampStart, _ = time.Parse(time.RFC3339Nano, tsStartStr)
					}
					if tsEndStr, tsOk := prMap["timestamp_end"].(string); tsOk {
						pr.TimestampEnd, _ = time.Parse(time.RFC3339Nano, tsEndStr)
					}
					if dur, durOk := toInt64(prMap["duration_ms"]); durOk {
						pr.DurationMs = dur
					}
					pr.Success, _ = prMap["success"].(bool)
					if ic, icOk := prMap["input_context"].(map[string]interface{}); icOk {
						pr.InputContext = ic
					}
					if lm, lmOk := prMap["llm_metrics"].(map[string]interface{}); lmOk {
						pr.LLMMetrics = lm
					}
					if cost, costOk := prMap["cost_incurred"].(float64); costOk {
						pr.CostIncurred = cost
					}
					pr.OutputSummary, _ = prMap["output_summary"].(string)
					pr.ErrorDetails, _ = prMap["error_details"].(string)
					// SupervisorFeedback would need similar parsing if included
					perfRecords = append(perfRecords, &pr)
				}
			}
		}

		retireErr := m.RetireWorkerInstance(instanceID, reason, finalStatus, finalUsage, perfRecords)
		if retireErr != nil {
			return nil, retireErr
		}

		if !i.RemoveHandle(handleID) {
			i.Logger().Warnf("Failed to remove ConversationManager handle %s for retired instance %s", handleID, instanceID)
		}
		return nil, nil
	},
}

var specAIWorkerInstanceUpdateStatus = ToolSpec{
	Name: "AIWorkerInstance.UpdateStatus", Description: "Updates the status and optionally the last error of an active AI Worker Instance.",
	Args: []ArgSpec{
		{Name: "instance_id", Type: ArgTypeString, Required: true},
		{Name: "status", Type: ArgTypeString, Required: true},
		{Name: "last_error", Type: ArgTypeString, Required: false},
	}, ReturnType: ArgTypeNil,
}

var toolAIWorkerInstanceUpdateStatus = ToolImplementation{
	Spec: specAIWorkerInstanceUpdateStatus,
	Func: func(i *Interpreter, argsGiven []interface{}) (interface{}, error) {
		m, err := getAIWorkerManager(i)
		if err != nil {
			return nil, err
		}
		validatedArgsList, valErr := ValidateAndConvertArgs(specAIWorkerInstanceUpdateStatus, argsGiven)
		if valErr != nil {
			return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("Validation failed for tool %s: %s", specAIWorkerInstanceUpdateStatus.Name, valErr.Error()), ErrInvalidArgument)
		}
		parsedArgs := mapValidatedArgsListToMapByName(specAIWorkerInstanceUpdateStatus.Args, validatedArgsList)

		instanceID, _ := parsedArgs["instance_id"].(string)
		statusStr, _ := parsedArgs["status"].(string)
		lastError, _ := parsedArgs["last_error"].(string)

		updateErr := m.UpdateInstanceStatus(instanceID, AIWorkerInstanceStatus(statusStr), lastError)
		if updateErr != nil {
			return nil, updateErr
		}
		return nil, nil
	},
}

var specAIWorkerInstanceUpdateTokenUsage = ToolSpec{
	Name: "AIWorkerInstance.UpdateTokenUsage", Description: "Updates the session token usage for an active AI Worker Instance.",
	Args: []ArgSpec{
		{Name: "instance_id", Type: ArgTypeString, Required: true},
		{Name: "input_tokens", Type: ArgTypeInt, Required: true},
		{Name: "output_tokens", Type: ArgTypeInt, Required: true},
	}, ReturnType: ArgTypeNil,
}

var toolAIWorkerInstanceUpdateTokenUsage = ToolImplementation{
	Spec: specAIWorkerInstanceUpdateTokenUsage,
	Func: func(i *Interpreter, argsGiven []interface{}) (interface{}, error) {
		m, err := getAIWorkerManager(i)
		if err != nil {
			return nil, err
		}
		validatedArgsList, valErr := ValidateAndConvertArgs(specAIWorkerInstanceUpdateTokenUsage, argsGiven)
		if valErr != nil {
			return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("Validation failed for tool %s: %s", specAIWorkerInstanceUpdateTokenUsage.Name, valErr.Error()), ErrInvalidArgument)
		}
		parsedArgs := mapValidatedArgsListToMapByName(specAIWorkerInstanceUpdateTokenUsage.Args, validatedArgsList)

		instanceID, _ := parsedArgs["instance_id"].(string)
		inputTokens, _ := toInt64(parsedArgs["input_tokens"])
		outputTokens, _ := toInt64(parsedArgs["output_tokens"])

		updateErr := m.UpdateInstanceSessionTokenUsage(instanceID, inputTokens, outputTokens)
		if updateErr != nil {
			return nil, updateErr
		}
		return nil, nil
	},
}
