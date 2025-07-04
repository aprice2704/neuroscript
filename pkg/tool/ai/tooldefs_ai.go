// NeuroScript Version: 0.3.1
// File version: 0.1.0
// Defines a slice of ToolImplementation structs for AI Worker Management tools.
// filename: pkg/tool/ai/tooldefs_ai.go

package ai

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/google/generative-ai-go/genai"
)

// ToGenaiType converts the internal ArgType to the corresponding genai.Type.
func ToGenaiType(at tool.ArgType) (genai.Type, error) {
	switch at {
	case tool.ArgTypeString, tool.ArgTypeAny:
		return genai.TypeString, nil
	case tool.ArgTypeInt:
		return genai.TypeInteger, nil
	case tool.ArgTypeFloat:
		return genai.TypeNumber, nil
	case tool.ArgTypeBool:
		return genai.TypeBoolean, nil
	case tool.ArgTypeMap:
		return genai.TypeObject, nil
	case tool.ArgTypeSlice, tool.ArgTypeSliceString, tool.ArgTypeSliceInt, tool.ArgTypeSliceFloat, tool.ArgTypeSliceBool, tool.ArgTypeSliceMap, tool.ArgTypeSliceAny:
		return genai.TypeArray, nil
	case tool.ArgTypeNil:
		return genai.TypeUnspecified, fmt.Errorf("cannot convert ArgTypeNil to a genai.Type for LLM function declaration expecting a specific type")
	default:
		return genai.TypeUnspecified, fmt.Errorf("unsupported ArgType '%s' cannot be converted to genai.Type", at)
	}
}
