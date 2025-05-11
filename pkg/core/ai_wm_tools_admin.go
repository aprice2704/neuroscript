// NeuroScript Version: 0.3.0
// File version: 0.1.0
// AI Worker Management: Administrative Tools (Load/Save All)
// filename: pkg/core/ai_wm_tools_admin.go

package core

import (
	"fmt"
	// "time" // Not directly needed here
	// "github.com/google/uuid" // Not directly needed here
)

var specAIWorkerDefinitionLoadAll = ToolSpec{
	Name:        "AIWorkerDefinition.LoadAll",
	Description: "Reloads all worker definitions from the configured JSON file.",
	Args:        []ArgSpec{},
	ReturnType:  ArgTypeString,
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
	Name:        "AIWorkerDefinition.SaveAll",
	Description: "Saves all current worker definitions to the configured JSON file.",
	Args:        []ArgSpec{},
	ReturnType:  ArgTypeString,
}

var toolAIWorkerDefinitionSaveAll = ToolImplementation{
	Spec: specAIWorkerDefinitionSaveAll,
	Func: func(i *Interpreter, argsGiven []interface{}) (interface{}, error) {
		m, err := getAIWorkerManager(i)
		if err != nil {
			return nil, err
		}
		// No need to validate args for a zero-arg tool
		saveErr := m.SaveWorkerDefinitionsToFile()
		if saveErr != nil {
			return nil, saveErr
		}
		return fmt.Sprintf("Saved %d worker definitions.", len(m.ListWorkerDefinitions(nil))), nil
	},
}

var specAIWorkerSavePerformanceData = ToolSpec{
	Name: "AIWorker.SavePerformanceData", Description: "Explicitly triggers saving of all retired instance performance data. Usually handled automatically on retire.",
	Args:       []ArgSpec{},
	ReturnType: ArgTypeString,
}

var toolAIWorkerSavePerformanceData = ToolImplementation{
	Spec: specAIWorkerSavePerformanceData,
	Func: func(i *Interpreter, argsGiven []interface{}) (interface{}, error) {
		m, err := getAIWorkerManager(i)
		if err != nil {
			return nil, err
		}
		m.mu.Lock()
		defer m.mu.Unlock()
		// This tool's main purpose is to ensure the definition file (with summaries) is saved.
		// Raw performance data for instances is appended when an instance is retired.
		// There isn't a separate "save all raw performance data" command in the current manager design
		// beyond what happens at instance retirement.
		if saveErr := m.persistDefinitionsUnsafe(); saveErr != nil { // Corrected call
			return nil, saveErr
		}
		i.Logger().Debug("AIWorker.SavePerformanceData: Called. Ensured definitions (with summaries) are saved. Raw performance data appends automatically on instance retirement.")
		return "Ensured definitions (with summaries) are saved. Raw performance data appends automatically.", nil
	},
}

var specAIWorkerLoadPerformanceData = ToolSpec{
	Name: "AIWorker.LoadPerformanceData", Description: "Reloads all worker definitions, which implicitly re-processes performance summaries from persisted data.",
	Args:       []ArgSpec{},
	ReturnType: ArgTypeString,
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
