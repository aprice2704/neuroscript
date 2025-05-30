// NeuroScript Version: 0.3.1
// File version: 0.1.2 // Removed persistDefinitionsUnsafe from LogPerformance tool
// AI Worker Management: Performance and Logging Tools
// filename: pkg/core/ai_wm_tools_performance.go

package core

import (
	"fmt"
	"time"
	// "github.com/google/uuid" // Not directly needed here
	// Assuming getAIWorkerManager, ValidateAndConvertArgs, mapValidatedArgsListToMapByName,
	// toInt64, convertPerformanceRecordToMap are defined elsewhere or provided
	// by the interpreter/tool infra.
)

var specAIWorkerLogPerformance = ToolSpec{ //
	Name:        "AIWorker.LogPerformance",                          //
	Description: "Logs a performance record for an AI Worker task.", //
	Category:    "AI Worker Management",                             //
	Args: []ArgSpec{ //
		{Name: "task_id", Type: ArgTypeString, Required: true, Description: "Unique ID for the task."},                                                     //
		{Name: "instance_id", Type: ArgTypeString, Required: true, Description: "ID of the AIWorkerInstance used."},                                        //
		{Name: "definition_id", Type: ArgTypeString, Required: true, Description: "ID of the AIWorkerDefinition used."},                                    //
		{Name: "timestamp_start", Type: ArgTypeString, Required: true, Description: "Start timestamp (RFC3339Nano or RFC3339 format)."},                    //
		{Name: "timestamp_end", Type: ArgTypeString, Required: true, Description: "End timestamp (RFC3339Nano or RFC3339 format)."},                        //
		{Name: "duration_ms", Type: ArgTypeInt, Required: true, Description: "Task duration in milliseconds."},                                             //
		{Name: "success", Type: ArgTypeBool, Required: true, Description: "Whether the task was successful."},                                              //
		{Name: "input_context", Type: ArgTypeMap, Required: false, Description: "Optional map of input context details."},                                  //
		{Name: "llm_metrics", Type: ArgTypeMap, Required: false, Description: "Optional map of LLM-specific metrics (e.g., token counts, finish reason)."}, //
		{Name: "cost_incurred", Type: ArgTypeFloat, Required: false, Description: "Optional cost incurred for this task."},                                 //
		{Name: "output_summary", Type: ArgTypeString, Required: false, Description: "Optional summary of the task output."},                                //
		{Name: "error_details", Type: ArgTypeString, Required: false, Description: "Optional error details if success is false."},                          //
	},
	ReturnType:      ArgTypeString,                                                                                                                                                                                                                                                                                              //
	ReturnHelp:      "Returns the TaskID string of the logged performance record.",                                                                                                                                                                                                                                              //
	Example:         `TOOL.AIWorker.LogPerformance(task_id: "task_abc", instance_id: "inst_123", definition_id: "def_xyz", timestamp_start: "2023-10-27T10:00:00.000Z", timestamp_end: "2023-10-27T10:00:05.123Z", duration_ms: 5123, success: true, llm_metrics: {"input_tokens":10, "output_tokens":50})`,                     //
	ErrorConditions: "ErrAIWorkerManagerMissing; ErrInvalidArgument if required arguments are missing/invalid type (e.g., timestamp format, duration_ms not int); Errors from AIWorkerManager.logPerformanceRecordUnsafe (e.g., definition not found). Persisting definitions error is now only a warning if it were to occur.", // Updated error conditions
}

var toolAIWorkerLogPerformance = ToolImplementation{ //
	Spec: specAIWorkerLogPerformance, //
	Func: func(i *Interpreter, argsGiven []interface{}) (interface{}, error) { //
		m, err := getAIWorkerManager(i) //
		if err != nil {                 //
			return nil, err //
		}
		validatedArgsList, valErr := ValidateAndConvertArgs(specAIWorkerLogPerformance, argsGiven) //
		if valErr != nil {                                                                         //
			return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("Validation failed for tool %s: %s", specAIWorkerLogPerformance.Name, valErr.Error()), ErrInvalidArgument) //
		}
		parsedArgs := mapValidatedArgsListToMapByName(specAIWorkerLogPerformance.Args, validatedArgsList) //

		record := PerformanceRecord{}                                 //
		record.TaskID, _ = parsedArgs["task_id"].(string)             //
		record.InstanceID, _ = parsedArgs["instance_id"].(string)     //
		record.DefinitionID, _ = parsedArgs["definition_id"].(string) //

		tsStartStr, _ := parsedArgs["timestamp_start"].(string)               //
		record.TimestampStart, err = time.Parse(time.RFC3339Nano, tsStartStr) //
		if err != nil {                                                       //
			record.TimestampStart, err = time.Parse(time.RFC3339, tsStartStr) // Fallback //
			if err != nil {                                                   //
				return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("invalid timestamp_start format '%s' for tool %s: %v", tsStartStr, specAIWorkerLogPerformance.Name, err), ErrInvalidArgument) //
			}
		}

		tsEndStr, _ := parsedArgs["timestamp_end"].(string)               //
		record.TimestampEnd, err = time.Parse(time.RFC3339Nano, tsEndStr) //
		if err != nil {                                                   //
			record.TimestampEnd, err = time.Parse(time.RFC3339, tsEndStr) // Fallback //
			if err != nil {                                               //
				return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("invalid timestamp_end format '%s' for tool %s: %v", tsEndStr, specAIWorkerLogPerformance.Name, err), ErrInvalidArgument) //
			}
		}

		durationInt64, _ := toInt64(parsedArgs["duration_ms"]) //
		record.DurationMs = durationInt64                      //
		record.Success, _ = parsedArgs["success"].(bool)       //

		if ic, ok := parsedArgs["input_context"].(map[string]interface{}); ok { //
			record.InputContext = ic //
		}
		if lm, ok := parsedArgs["llm_metrics"].(map[string]interface{}); ok { //
			record.LLMMetrics = lm //
		}
		if ci, ok := parsedArgs["cost_incurred"].(float64); ok { //
			record.CostIncurred = ci //
		}
		if osVal, ok := parsedArgs["output_summary"].(string); ok { // Renamed 'os' to 'osVal' //
			record.OutputSummary = osVal //
		}
		if ed, ok := parsedArgs["error_details"].(string); ok { //
			record.ErrorDetails = ed //
		}

		// The actual writing of the performance record to a file should happen
		// via a separate mechanism that does *not* hold m.mu.
		// For example, ExecuteStatelessTask would marshal 'record' and then call
		// a file-writing helper after all m.mu locks are released.

		// This tool function's responsibility is to update the in-memory aggregate summary.
		m.mu.Lock()         //
		defer m.mu.Unlock() //

		logErr := m.logPerformanceRecordUnsafe(&record) // This updates in-memory summary
		if logErr != nil {                              //
			return nil, logErr //
		}

		// REMOVED: m.persistDefinitionsUnsafe() call to avoid I/O under lock.
		// Definition persistence (saving all definitions) is a separate concern
		// and should not be tied to every performance log event if it involves I/O.
		// If definitions need saving because their summary changed, that should be
		// handled by a more strategic save mechanism (e.g., periodic, on shutdown, explicit admin tool).
		i.Logger().Debugf("%s: In-memory performance summary updated for TaskID: %s. File persistence of definitions NOT triggered here.", specAIWorkerLogPerformance.Name, record.TaskID)

		return record.TaskID, nil //
	},
}

var specAIWorkerGetPerformanceRecords = ToolSpec{ //
	Name:        "AIWorker.GetPerformanceRecords",                                               //
	Description: "Retrieves persisted performance records for a specific AI Worker Definition.", //
	Category:    "AI Worker Management",                                                         //
	Args: []ArgSpec{ //
		{Name: "definition_id", Type: ArgTypeString, Required: true, Description: "ID of the AIWorkerDefinition for which to retrieve records."},       //
		{Name: "filters", Type: ArgTypeMap, Required: false, Description: "Optional map of filters to apply to the records (e.g., {'success':true})."}, //
	},
	ReturnType:      ArgTypeSliceMap,                                                                                                                                                                                 //
	ReturnHelp:      "Returns a slice of maps, where each map represents a PerformanceRecord. Returns an empty slice if no records match or exist.",                                                                  //
	Example:         `TOOL.AIWorker.GetPerformanceRecords(definition_id: "google-gemini-1.5-pro", filters: {"success":true})`,                                                                                        //
	ErrorConditions: "ErrAIWorkerManagerMissing; ErrInvalidArgument for missing/invalid args; Errors from AIWorkerManager.GetPerformanceRecordsForDefinition (e.g., file I/O for persistence, JSON parsing errors).", //
}

var toolAIWorkerGetPerformanceRecords = ToolImplementation{ //
	Spec: specAIWorkerGetPerformanceRecords, //
	Func: func(i *Interpreter, argsGiven []interface{}) (interface{}, error) { //
		m, err := getAIWorkerManager(i) //
		if err != nil {                 //
			return nil, err //
		}
		validatedArgsList, valErr := ValidateAndConvertArgs(specAIWorkerGetPerformanceRecords, argsGiven) //
		if valErr != nil {                                                                                //
			return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("Validation failed for tool %s: %s", specAIWorkerGetPerformanceRecords.Name, valErr.Error()), ErrInvalidArgument) //
		}
		parsedArgs := mapValidatedArgsListToMapByName(specAIWorkerGetPerformanceRecords.Args, validatedArgsList) //
		defID, _ := parsedArgs["definition_id"].(string)                                                         //
		filters, _ := parsedArgs["filters"].(map[string]interface{})                                             //

		// This call now relies on the refactored m.GetPerformanceRecordsForDefinition
		// in pkg/core/ai_wm_performance.go to handle its locking and I/O correctly.
		records, getErr := m.GetPerformanceRecordsForDefinition(defID, filters) //
		if getErr != nil {                                                      //
			return nil, getErr //
		}
		resultList := make([]interface{}, len(records)) //
		for idx, rec := range records {                 //
			resultList[idx] = convertPerformanceRecordToMap(rec) //
		}
		return resultList, nil //
	},
}

// Helper function (assuming it's defined elsewhere or you'll add it)
// func getAIWorkerManager(i *Interpreter) (*AIWorkerManager, error) {
// 	if i.aiWorkerManager == nil {
// 		return nil, NewRuntimeError(ErrorCodeConfiguration, "AIWorkerManager not available in interpreter context", nil)
// 	}
// 	return i.aiWorkerManager, nil
// }

// Helper function (assuming it's defined elsewhere)
// func convertPerformanceRecordToMap(record *PerformanceRecord) map[string]interface{} {
//   // Implementation to convert struct to map, possibly using json.Marshal then Unmarshal, or struct tags
//   data, _ := json.Marshal(record)
//   var recordMap map[string]interface{}
//   _ = json.Unmarshal(data, &recordMap)
//   return recordMap
// }

// Helpers like ValidateAndConvertArgs, mapValidatedArgsListToMapByName, toInt64
// are assumed to be part of your existing tool infrastructure.
