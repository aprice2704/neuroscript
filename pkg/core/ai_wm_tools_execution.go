// NeuroScript Version: 0.3.1
// File version: 0.1.1 // Populated Category, Example, ReturnHelp, ErrorConditions for ToolSpec.
// AI Worker Management: Stateless Execution Tool
// filename: pkg/core/ai_wm_tools_execution.go
// nlines: 60 // Approximate

package core

import (
	"fmt"
	// "time" // Not directly needed here
	// "github.com/google/uuid" // Not directly needed here
)

var specAIWorkerExecuteStateless = ToolSpec{
	Name:        "AIWorker.ExecuteStatelessTask",
	Description: "Executes a stateless task using an AI Worker Definition.",
	Category:    "AI Worker Management",
	Args: []ArgSpec{
		{Name: "name", Type: ArgTypeString, Required: true, Description: "name of the AIWorkerDefinition to use."},
		{Name: "prompt", Type: ArgTypeString, Required: true, Description: "The prompt/input text for the LLM."},
		{Name: "config_overrides", Type: ArgTypeMap, Required: false, Description: "Optional map of configuration overrides for this specific execution."},
	},
	ReturnType:      ArgTypeMap,
	ReturnHelp:      "Returns a map: {'output': string (LLM response), 'taskId': string, 'cost': float64}. Returns nil on error.",
	Example:         `TOOL.AIWorker.ExecuteStatelessTask(definition_id: "google-gemini-1.5-flash", prompt: "Translate 'hello' to French.")`,
	ErrorConditions: "ErrAIWorkerManagerMissing; ErrInvalidArgument for missing/invalid args; ErrConfiguration if interpreter's LLMClient is nil; Errors from AIWorkerManager.ExecuteStatelessTask (e.g., ErrDefinitionNotFound, LLM communication errors, rate limits); ErrInternal if performance record is nil without error.",
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
		defname, _ := parsedArgs["name"].(string)
		prompt, _ := parsedArgs["prompt"].(string)
		overrides, _ := parsedArgs["config_overrides"].(map[string]interface{})

		if i.llmClient == nil {
			return nil, NewRuntimeError(ErrorCodeConfiguration, "Interpreter's LLMClient is nil, cannot execute stateless task for tool "+specAIWorkerExecuteStateless.Name, ErrConfiguration)
		}

		output, perfRecord, execErr := m.ExecuteStatelessTask(defname, i.llmClient, prompt, overrides)
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
