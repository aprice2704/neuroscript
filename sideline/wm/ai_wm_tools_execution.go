// NeuroScript Version: 0.4.1
// File version: 1
// Purpose: Refactored tool func to remove validation call and use direct args from bridge.
// AI Worker Management: Stateless Execution Tool
// filename: pkg/core/ai_wm_tools_execution.go
// nlines: 48

package core

import (
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/sideline/nspatch"
)

var specAIWorkerExecuteStateless = tool.ToolSpec{
	Name:     "AIWorker.ExecuteStatelessTask",
	Category: "AI Worker Management",
	Args: []tool.ArgSpec{
		{Name: "name", Type: tool.ArgTypeString, Required: true},
		{Name: "prompt", Type: tool.ArgTypeString, Required: true},
		{Name: "config_overrides", Type: tool.ArgTypeMap, Required: false},
	},
	ReturnType: "map",
}

var toolAIWorkerExecuteStateless = tool.ToolImplementation{
	Spec: specAIWorkerExecuteStateless,
	Func: func(i *neurogo.Interpreter, args []interface{}) (interface{}, error) {
		m, err := getAIWorkerManager(i)
		if err != nil {
			return nil, err
		}
		// Arguments are lang.Positional, validation is handled by the bridge
		defName, _ := args[0].(string)
		prompt, _ := args[1].(string)
		var overrides map[string]interface{}
		if args[2] != nil {
			overrides, _ = args[2].(map[string]interface{})
		}

		if i.llmClient == nil {
			return nil, lang.NewRuntimeError(lang.ErrorCodeConfiguration, "Interpreter's LLMClient is nil", lang.ErrConfiguration)
		}

		output, perfRecord, execErr := m.ExecuteStatelessTask(defName, i.llmClient, prompt, overrides)
		if execErr != nil {
			taskId := "unknown"
			if perfRecord != nil {
				taskId = perfRecord.TaskID
			}
			i.Logger().Warnf("ExecuteStatelessTask for tool %s (TaskID: %s) failed: %v", specAIWorkerExecuteStateless.Name, taskId, execErr)
			return nil, execErr
		}
		if perfRecord == nil {
			return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "ExecuteStatelessTask returned nil performance record without error", nspatch.ErrInternal)
		}
		return map[string]interface{}{"output": output, "taskId": perfRecord.TaskID, "cost": perfRecord.CostIncurred}, nil
	},
}
