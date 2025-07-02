// NeuroScript Version: 0.3.1
// File version: 0.1.0
// Defines a slice of ToolImplementation structs for AI Worker Management tools.
// filename: pkg/tool/ai/tooldefs_ai.go

package ai

import "github.com/aprice2704/neuroscript/pkg/tool"

// aiWmToolsToRegister contains ToolImplementation definitions for AI Worker Management tools.
var aiWmToolsToRegister = []tool.ToolImplementation{	//
	// Definition Tools (from ai_wm_tools_definitions.go)
	//	toolAIWorkerDefinitionAdd,  //
	wm.toolAIWorkerDefinitionGet,	//
	wm.toolAIWorkerDefinitionList,	//
	// toolAIWorkerDefinitionUpdate, //
	// toolAIWorkerDefinitionRemove, //

	// Admin/Load-Save Tools (from ai_wm_tools_admin.go)
	wm.toolAIWorkerDefinitionLoadAll,	//
	// toolAIWorkerDefinitionSaveAll,   //
	// toolAIWorkerSavePerformanceData, //
	wm.toolAIWorkerLoadPerformanceData,	//

	// Instance Tools (from ai_wm_tools_instances.go)
	wm.toolAIWorkerInstanceSpawn,		//
	wm.toolAIWorkerInstanceGet,		//
	wm.toolAIWorkerInstanceListActive,		//
	wm.toolAIWorkerInstanceRetire,		//
	wm.toolAIWorkerInstanceUpdateStatus,	//
	wm.toolAIWorkerInstanceUpdateTokenUsage,	//

	// Execution Tools (from ai_wm_tools_execution.go)
	wm.toolAIWorkerExecuteStateless,	//

	// Performance Tools (from ai_wm_tools_performance.go)
	// Ensure these variables are defined, likely in ai_wm_tools_performance.go
	wm.toolAIWorkerGetPerformanceRecords,	// This should now be defined
	wm.toolAIWorkerLogPerformance,
}