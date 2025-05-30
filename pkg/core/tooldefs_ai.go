// NeuroScript Version: 0.3.1
// File version: 0.1.0
// Defines a slice of ToolImplementation structs for AI Worker Management tools.
// filename: pkg/core/tooldefs_ai.go

package core

// aiWmToolsToRegister contains ToolImplementation definitions for AI Worker Management tools.
var aiWmToolsToRegister = []ToolImplementation{ //
	// Definition Tools (from ai_wm_tools_definitions.go)
	//	toolAIWorkerDefinitionAdd,  //
	toolAIWorkerDefinitionGet,  //
	toolAIWorkerDefinitionList, //
	// toolAIWorkerDefinitionUpdate, //
	// toolAIWorkerDefinitionRemove, //

	// Admin/Load-Save Tools (from ai_wm_tools_admin.go)
	toolAIWorkerDefinitionLoadAll, //
	// toolAIWorkerDefinitionSaveAll,   //
	// toolAIWorkerSavePerformanceData, //
	toolAIWorkerLoadPerformanceData, //

	// Instance Tools (from ai_wm_tools_instances.go)
	toolAIWorkerInstanceSpawn,            //
	toolAIWorkerInstanceGet,              //
	toolAIWorkerInstanceListActive,       //
	toolAIWorkerInstanceRetire,           //
	toolAIWorkerInstanceUpdateStatus,     //
	toolAIWorkerInstanceUpdateTokenUsage, //

	// Execution Tools (from ai_wm_tools_execution.go)
	toolAIWorkerExecuteStateless, //

	// Performance Tools (from ai_wm_tools_performance.go)
	// Ensure these variables are defined, likely in ai_wm_tools_performance.go
	toolAIWorkerGetPerformanceRecords, // This should now be defined
	// toolAIWorkerLogPerformance, // If this is not an actual tool exposed to NeuroScript, remove it from this list.
}
