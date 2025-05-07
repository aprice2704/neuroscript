// NeuroScript Version: 0.3.0
// File version: 0.1.0
// AI Worker Management: Stateless Execution Tool
// filename: pkg/core/ai_wm_tools_execution.go

package core

import (
	"fmt"
	// "time" // Not directly needed here
	// "github.com/google/uuid" // Not directly needed here
)

var specAIWorkerExecuteStateless = ToolSpec{
	Name: "AIWorker.ExecuteStatelessTask", Description: "Executes a stateless task using an AI Worker Definition.",
	Args: []ArgSpec{
		{Name: "definition_id", Type: ArgTypeString, Required: true},
		{Name: "prompt", Type: ArgTypeString, Required: true},
		{Name: "config_overrides", Type: ArgTypeMap, Required: false},
	},
	ReturnType: ArgTypeMap, // Returns map: {"output": string, "taskId": string, "cost": float64}
}

var toolAIWorkerExecuteStateless = ToolImplementation{
	Spec: specAIWorkerExecuteStateless,
	Func: func(i *Interpreter, argsGiven []interface{}) (interface{}, error) {
		m, err := getAIWorkerManager(i)
		if err != nil {
			return nil, err
		}
		validatedArgsList, valErr := ValidateAndConvertArgs(specAIWorkerExecuteStateless, argsGiven)
		if valErr != nil {
			return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("Validation failed for tool %s: %s", specAIWorkerExecuteStateless.Name, valErr.Error()), ErrInvalidArgument)
		}
		parsedArgs := mapValidatedArgsListToMapByName(specAIWorkerExecuteStateless.Args, validatedArgsList)
		defID, _ := parsedArgs["definition_id"].(string)
		prompt, _ := parsedArgs["prompt"].(string)
		overrides, _ := parsedArgs["config_overrides"].(map[string]interface{})

		if i.llmClient == nil {
			return nil, NewRuntimeError(ErrorCodeConfiguration, "Interpreter's LLMClient is nil, cannot execute stateless task for tool "+specAIWorkerExecuteStateless.Name, ErrConfiguration)
		}

		output, perfRecord, execErr := m.ExecuteStatelessTask(defID, i.llmClient, prompt, overrides)
		if execErr != nil {
			taskId := "unknown"
			if perfRecord != nil {
				taskId = perfRecord.TaskID
			}
			i.Logger().Warnf("ExecuteStatelessTask for tool %s (TaskID: %s) failed: %v", specAIWorkerExecuteStateless.Name, taskId, execErr)
			return nil, execErr
		}
		if perfRecord == nil {
			return nil, NewRuntimeError(ErrorCodeInternal, "ExecuteStatelessTask returned nil performance record without error for tool "+specAIWorkerExecuteStateless.Name, ErrInternal)
		}
		return map[string]interface{}{"output": output, "taskId": perfRecord.TaskID, "cost": perfRecord.CostIncurred}, nil
	},
}
