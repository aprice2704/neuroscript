// NeuroScript Version: 0.7.2
// File version: 1
// Purpose: Defines and implements the 'metadata' toolset for NeuroScript.
// filename: pkg/tool/metadata/tools.go
// nlines: 90
// risk_rating: LOW
package metadata

import (
	"strings"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/metadata"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// Group is the official tool group name for this toolset.
const Group = "metadata"

// MetadataToolsToRegister is the list of tool implementations for registration.
var MetadataToolsToRegister = []tool.ToolImplementation{
	{
		Spec: tool.ToolSpec{
			Name:        "Detect",
			Group:       Group,
			Description: "Detects the serialization format ('md' or 'ns') of a string content by checking for a '::serialization:' key.",
			Args: []tool.ArgSpec{
				{Name: "content", Type: tool.ArgTypeString, Required: true},
			},
			ReturnType: tool.ArgTypeString,
		},
		Func: detectFunc,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Parse",
			Group:       Group,
			Description: "Auto-detects serialization and parses content into a metadata map and a content body string.",
			Args: []tool.ArgSpec{
				{Name: "content", Type: tool.ArgTypeString, Required: true},
			},
			ReturnType: tool.ArgTypeMap,
		},
		Func: parseFunc,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "NormalizeKey",
			Group:       Group,
			Description: "Normalizes a metadata key by converting it to lowercase and removing '.', '_', and '-' characters.",
			Args: []tool.ArgSpec{
				{Name: "key", Type: tool.ArgTypeString, Required: true},
			},
			ReturnType: tool.ArgTypeString,
		},
		Func: normalizeKeyFunc,
	},
}

func detectFunc(rt tool.Runtime, args []interface{}) (interface{}, error) {
	content, ok := args[0].(string)
	if !ok {
		return nil, lang.ErrInvalidArgument
	}
	return metadata.DetectSerialization(strings.NewReader(content))
}

func parseFunc(rt tool.Runtime, args []interface{}) (interface{}, error) {
	content, ok := args[0].(string)
	if !ok {
		return nil, lang.ErrInvalidArgument
	}

	meta, body, ser, err := metadata.ParseWithAutoDetect(strings.NewReader(content))
	if err != nil {
		return nil, err
	}

	// Convert metadata.Store (map[string]string) to map[string]interface{}
	metaMap := make(map[string]interface{}, len(meta))
	for k, v := range meta {
		metaMap[k] = v
	}

	return map[string]interface{}{
		"serialization": ser,
		"metadata":      metaMap,
		"body":          string(body),
	}, nil
}

func normalizeKeyFunc(rt tool.Runtime, args []interface{}) (interface{}, error) {
	key, ok := args[0].(string)
	if !ok {
		return nil, lang.ErrInvalidArgument
	}
	return metadata.NormalizeKey(key), nil
}
