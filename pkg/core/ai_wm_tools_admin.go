// NeuroScript Version: 0.3.1
// File version: 0.1.1 // Populated Category, Example, ReturnHelp, ErrorConditions for ToolSpecs.
// AI Worker Management: Administrative Tools (Load/Save All)
// filename: pkg/core/ai_wm_tools_admin.go
// nlines: 100 // Approximate

package core

import (
	"fmt"
	// "time" // Not directly needed here
	// "github.com/google/uuid" // Not directly needed here
)

var specAIWorkerDefinitionLoadAll = ToolSpec{
	Name:            "AIWorkerDefinition.LoadAll",
	Description:     "Reloads all worker definitions from the configured JSON file.",
	Category:        "AI Worker Management",
	Args:            []ArgSpec{},
	ReturnType:      ArgTypeString,
	ReturnHelp:      "Returns a string message indicating the number of definitions reloaded, e.g., 'Reloaded X worker definitions.'.",
	Example:         `TOOL.AIWorkerDefinition.LoadAll()`,
	ErrorConditions: "ErrAIWorkerManagerMissing; Errors from AIWorkerManager.LoadWorkerDefinitionsFromFile (e.g., related to file I/O, JSON parsing, or validation of loaded definitions).",
}

var toolAIWorkerDefinitionLoadAll = ToolImplementation{
	Spec: specAIWorkerDefinitionLoadAll,
	Func: func(i *Interpreter, argsGiven []interface{}) (interface{}, error) {
		m, err := getAIWorkerManager(i)
		if err != nil {
			return nil, err
		}
		// No need to validate args for a zero-arg tool
		loadErr := m.LoadWorkerDefinitionsFromFile()
		if loadErr != nil {
			return nil, loadErr
		}
		return fmt.Sprintf("Reloaded %d worker definitions.", len(m.ListWorkerDefinitions(nil))), nil
	},
}

var specAIWorkerDefinitionSaveAll = ToolSpec{
	Name:            "AIWorkerDefinition.SaveAll",
	Description:     "Saves all current worker definitions to the configured JSON file.",
	Category:        "AI Worker Management",
	Args:            []ArgSpec{},
	ReturnType:      ArgTypeString,
	ReturnHelp:      "Returns a string message indicating the number of definitions saved, e.g., 'Saved X worker definitions.'.",
	Example:         `TOOL.AIWorkerDefinition.SaveAll()`,
	ErrorConditions: "ErrAIWorkerManagerMissing; Errors from AIWorkerManager.SaveWorkerDefinitionsToFile (e.g., related to file I/O or JSON serialization).",
}

// var toolAIWorkerDefinitionSaveAll = ToolImplementation{
// 	Spec: specAIWorkerDefinitionSaveAll,
// 	Func: func(i *Interpreter, argsGiven []interface{}) (interface{}, error) {
// 		m, err := getAIWorkerManager(i)
// 		if err != nil {
// 			return nil, err
// 		}
// 		// No need to validate args for a zero-arg tool
// 		saveErr := m.SaveWorkerDefinitionsToFile()
// 		if saveErr != nil {
// 			return nil, saveErr
// 		}
// 		return fmt.Sprintf("Saved %d worker definitions.", len(m.ListWorkerDefinitions(nil))), nil
// 	},
// }

var specAIWorkerSavePerformanceData = ToolSpec{
	Name:            "AIWorker.SavePerformanceData",
	Description:     "Explicitly triggers saving of all worker definitions (which include performance summaries). Raw performance data for instances is appended when an instance is retired.",
	Category:        "AI Worker Management",
	Args:            []ArgSpec{},
	ReturnType:      ArgTypeString,
	ReturnHelp:      "Returns a string message: 'Ensured definitions (with summaries) are saved. Raw performance data appends automatically.'.",
	Example:         `TOOL.AIWorker.SavePerformanceData()`,
	ErrorConditions: "ErrAIWorkerManagerMissing; Errors from AIWorkerManager.persistDefinitionsUnsafe (e.g., file I/O or JSON serialization errors).",
}

// var toolAIWorkerSavePerformanceData = ToolImplementation{
// 	Spec: specAIWorkerSavePerformanceData,
// 	Func: func(i *Interpreter, argsGiven []interface{}) (interface{}, error) {
// 		m, err := getAIWorkerManager(i)
// 		if err != nil {
// 			return nil, err
// 		}
// 		m.mu.Lock()
// 		defer m.mu.Unlock()
// 		// This tool's main purpose is to ensure the definition file (with summaries) is saved.
// 		// Raw performance data for instances is appended when an instance is retired.
// 		// There isn't a separate "save all raw performance data" command in the current manager design
// 		// beyond what happens at instance retirement.
// 		if saveErr := m.persistDefinitionsUnsafe(); saveErr != nil { // Corrected call
// 			return nil, saveErr
// 		}
// 		i.Logger().Debug("AIWorker.SavePerformanceData: Called. Ensured definitions (with summaries) are saved. Raw performance data appends automatically on instance retirement.")
// 		return "Ensured definitions (with summaries) are saved. Raw performance data appends automatically.", nil
// 	},
// }

var specAIWorkerLoadPerformanceData = ToolSpec{
	Name:            "AIWorker.LoadPerformanceData",
	Description:     "Reloads all worker definitions, which implicitly re-processes performance summaries from persisted data.",
	Category:        "AI Worker Management",
	Args:            []ArgSpec{},
	ReturnType:      ArgTypeString,
	ReturnHelp:      "Returns a string message: 'Worker definitions and associated performance summaries reloaded.'.",
	Example:         `TOOL.AIWorker.LoadPerformanceData()`,
	ErrorConditions: "ErrAIWorkerManagerMissing; Errors from AIWorkerManager.LoadWorkerDefinitionsFromFile (e.g., file I/O, JSON parsing).",
}

var toolAIWorkerLoadPerformanceData = ToolImplementation{
	Spec: specAIWorkerLoadPerformanceData,
	Func: func(i *Interpreter, argsGiven []interface{}) (interface{}, error) {
		m, err := getAIWorkerManager(i)
		if err != nil {
			return nil, err
		}
		loadErr := m.LoadWorkerDefinitionsFromFile()
		if loadErr != nil {
			return nil, loadErr
		}
		return "Worker definitions and associated performance summaries reloaded.", nil
	},
}
