// NeuroScript Version: 0.3.1
// File version: 0.1.0
// Defines a slice of ToolImplementation structs for AI Worker Management tools.
// filename: pkg/core/tooldefs_ai_wm.go

package core

// aiWmToolsToRegister contains ToolImplementation definitions for AI Worker Management tools.
// These are typically global variables defined in their respective ai_wm_tools_*.go files.
// This array is intended to be concatenated with other similar arrays in a central
// registrar (e.g., zz_core_tools_registrar.go) to be processed by AddToolImplementations.
//
// If these tools are registered via this array, their registration should be removed
// from other functions like RegisterAIWorkerTools to avoid duplicates.
var aiWmToolsToRegister = []ToolImplementation{
	// Definition Tools (from ai_wm_tools_definitions.go)
	toolAIWorkerDefinitionAdd,
	toolAIWorkerDefinitionGet,
	toolAIWorkerDefinitionList,
	toolAIWorkerDefinitionUpdate,
	toolAIWorkerDefinitionRemove,

	// Admin/Load-Save Tools (from ai_wm_tools_admin.go)
	toolAIWorkerDefinitionLoadAll,
	toolAIWorkerDefinitionSaveAll,
	toolAIWorkerSavePerformanceData,
	toolAIWorkerLoadPerformanceData,

	// Instance Tools (from ai_wm_tools_instances.go)
	toolAIWorkerInstanceSpawn,
	toolAIWorkerInstanceGet,
	toolAIWorkerInstanceListActive,
	toolAIWorkerInstanceRetire,
	toolAIWorkerInstanceUpdateStatus,
	toolAIWorkerInstanceUpdateTokenUsage,

	// Execution Tools (from ai_wm_tools_execution.go)
	toolAIWorkerExecuteStateless,

	// Performance Tools (from ai_wm_tools_performance.go)
	toolAIWorkerLogPerformance,
	toolAIWorkerGetPerformanceRecords,
}
