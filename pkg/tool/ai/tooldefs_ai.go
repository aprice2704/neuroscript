// NeuroScript Version: 0.3.1
// File version: 0.1.0
// Defines a slice of ToolImplementation structs for AI Worker Management tools.
// filename: pkg/tool/ai/tooldefs_ai.go

package ai

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/tool"
	"google.golang.org/genai"
)

// ToGenaiType converts the internal ArgType to the corresponding genai.Type.
func (at ArgType) ToGenaiType() (genai.Type, error) {
	switch at {
	case ArgTypeString, ArgTypeAny:
		return genai.TypeString, nil
	case ArgTypeInt:
		return genai.TypeInteger, nil
	case ArgTypeFloat:
		return genai.TypeNumber, nil
	case ArgTypeBool:
		return genai.TypeBoolean, nil
	case ArgTypeMap:
		return genai.TypeObject, nil
	case ArgTypeSlice, ArgTypeSliceString, ArgTypeSliceInt, ArgTypeSliceFloat, ArgTypeSliceBool, ArgTypeSliceMap, ArgTypeSliceAny:
		return genai.TypeArray, nil
	case ArgTypeNil:
		return genai.TypeUnspecified, fmt.Errorf("cannot convert ArgTypeNil to a genai.Type for LLM function declaration expecting a specific type")
	default:
		return genai.TypeUnspecified, fmt.Errorf("unsupported ArgType '%s' cannot be converted to genai.Type", at)
	}
}

// aiWmToolsToRegister contains ToolImplementation definitions for AI Worker Management tools.
var aiWmToolsToRegister = []tool.ToolImplementation{ //
	// Definition Tools (from ai_wm_tools_definitions.go)
	//	toolAIWorkerDefinitionAdd,  //
	wm.toolAIWorkerDefinitionGet,  //
	wm.toolAIWorkerDefinitionList, //
	// toolAIWorkerDefinitionUpdate, //
	// toolAIWorkerDefinitionRemove, //

	// Admin/Load-Save Tools (from ai_wm_tools_admin.go)
	wm.toolAIWorkerDefinitionLoadAll, //
	// toolAIWorkerDefinitionSaveAll,   //
	// toolAIWorkerSavePerformanceData, //
	wm.toolAIWorkerLoadPerformanceData, //

	// Instance Tools (from ai_wm_tools_instances.go)
	wm.toolAIWorkerInstanceSpawn,            //
	wm.toolAIWorkerInstanceGet,              //
	wm.toolAIWorkerInstanceListActive,       //
	wm.toolAIWorkerInstanceRetire,           //
	wm.toolAIWorkerInstanceUpdateStatus,     //
	wm.toolAIWorkerInstanceUpdateTokenUsage, //

	// Execution Tools (from ai_wm_tools_execution.go)
	wm.toolAIWorkerExecuteStateless, //

	// Performance Tools (from ai_wm_tools_performance.go)
	// Ensure these variables are defined, likely in ai_wm_tools_performance.go
	wm.toolAIWorkerGetPerformanceRecords, // This should now be defined
	wm.toolAIWorkerLogPerformance,
}
