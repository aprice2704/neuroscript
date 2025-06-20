// NeuroScript Version: 0.4.1
// File version: 1
// Purpose: Refactored tool funcs to remove ValidateAndConvertArgs and use direct args from bridge.
// AI Worker Management: Performance and Logging Tools
// filename: pkg/core/ai_wm_tools_performance.go
// nlines: 121

package core

import (
	"fmt"
	"time"
)

var specAIWorkerLogPerformance = ToolSpec{
	Name:        "AIWorker.LogPerformance",
	Description: "Logs a performance record for an AI Worker task.",
	Category:    "AI Worker Management",
	Args: []ArgSpec{
		{Name: "task_id", Type: ArgTypeString, Required: true},
		{Name: "instance_id", Type: ArgTypeString, Required: true},
		{Name: "definition_id", Type: ArgTypeString, Required: true},
		{Name: "timestamp_start", Type: ArgTypeString, Required: true},
		{Name: "timestamp_end", Type: ArgTypeString, Required: true},
		{Name: "duration_ms", Type: ArgTypeInt, Required: true},
		{Name: "success", Type: ArgTypeBool, Required: true},
		{Name: "input_context", Type: ArgTypeMap, Required: false},
		{Name: "llm_metrics", Type: ArgTypeMap, Required: false},
		{Name: "cost_incurred", Type: ArgTypeFloat, Required: false},
		{Name: "output_summary", Type: ArgTypeString, Required: false},
		{Name: "error_details", Type: ArgTypeString, Required: false},
	},
	ReturnType: "string",
}

var toolAIWorkerLogPerformance = ToolImplementation{
	Spec: specAIWorkerLogPerformance,
	Func: func(i *Interpreter, args []interface{}) (interface{}, error) {
		m, err := getAIWorkerManager(i)
		if err != nil {
			return nil, err
		}

		// Arguments are now positional from the bridge
		record := PerformanceRecord{}
		record.TaskID, _ = args[0].(string)
		record.InstanceID, _ = args[1].(string)
		record.DefinitionID, _ = args[2].(string)
		tsStartStr, _ := args[3].(string)
		tsEndStr, _ := args[4].(string)
		record.DurationMs, _ = toInt64(args[5])
		record.Success, _ = args[6].(bool)

		if args[7] != nil {
			record.InputContext, _ = args[7].(map[string]interface{})
		}
		if args[8] != nil {
			record.LLMMetrics, _ = args[8].(map[string]interface{})
		}
		if args[9] != nil {
			record.CostIncurred, _ = args[9].(float64)
		}
		if args[10] != nil {
			record.OutputSummary, _ = args[10].(string)
		}
		if args[11] != nil {
			record.ErrorDetails, _ = args[11].(string)
		}

		// Timestamps require parsing with fallback
		record.TimestampStart, err = time.Parse(time.RFC3339Nano, tsStartStr)
		if err != nil {
			record.TimestampStart, err = time.Parse(time.RFC3339, tsStartStr)
			if err != nil {
				return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("invalid timestamp_start format '%s'", tsStartStr), err)
			}
		}
		record.TimestampEnd, err = time.Parse(time.RFC3339Nano, tsEndStr)
		if err != nil {
			record.TimestampEnd, err = time.Parse(time.RFC3339, tsEndStr)
			if err != nil {
				return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("invalid timestamp_end format '%s'", tsEndStr), err)
			}
		}

		logErr := m.logPerformanceRecordUnsafe(&record)
		if logErr != nil {
			return nil, logErr
		}

		i.Logger().Debugf("%s: In-memory summary updated for TaskID: %s.", specAIWorkerLogPerformance.Name, record.TaskID)
		return record.TaskID, nil
	},
}

var specAIWorkerGetPerformanceRecords = ToolSpec{
	Name:        "AIWorker.GetPerformanceRecords",
	Description: "Retrieves performance records for an AI Worker Definition.",
	Category:    "AI Worker Management",
	Args: []ArgSpec{
		{Name: "definition_id", Type: ArgTypeString, Required: true},
		{Name: "filters", Type: ArgTypeMap, Required: false},
	},
	ReturnType: "slice",
}

var toolAIWorkerGetPerformanceRecords = ToolImplementation{
	Spec: specAIWorkerGetPerformanceRecords,
	Func: func(i *Interpreter, args []interface{}) (interface{}, error) {
		m, err := getAIWorkerManager(i)
		if err != nil {
			return nil, err
		}
		defID, _ := args[0].(string)
		var filters map[string]interface{}
		if args[1] != nil {
			filters, _ = args[1].(map[string]interface{})
		}

		records, getErr := m.GetPerformanceRecordsForDefinition(defID, filters)
		if getErr != nil {
			return nil, getErr
		}
		resultList := make([]interface{}, len(records))
		for idx, rec := range records {
			resultList[idx] = convertPerformanceRecordToMap(rec)
		}
		return resultList, nil
	},
}
