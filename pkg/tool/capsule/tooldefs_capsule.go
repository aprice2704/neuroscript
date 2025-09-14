// NeuroScript Version: 0.7.1
// File version: 3
// Purpose: Defines the tool specifications for managing documentation capsules.
// filename: pkg/tool/capsule/tooldefs_capsule.go
// nlines: 75
// risk_rating: HIGH
package capsule

import (
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// Group is the official tool group name for this toolset.
const Group = "capsule"

// CapsuleToolsToRegister is the list of tool implementations that this
// package provides for registration.
var CapsuleToolsToRegister = []tool.ToolImplementation{
	{
		Spec: tool.ToolSpec{
			Name:        "List",
			Group:       Group,
			Description: "Lists the IDs of all available documentation capsules.",
			ReturnType:  tool.ArgTypeSliceString,
			Example:     `capsule.List()`,
		},
		Func:          toolListCapsules,
		RequiresTrust: false,
		Effects:       []string{"readonly"},
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Read",
			Group:       Group,
			Description: "Reads the content and metadata of a specific capsule by its full ID (e.g., 'capsule/aeiou@2').",
			Args: []tool.ArgSpec{
				{Name: "id", Type: tool.ArgTypeString, Description: "The fully qualified ID of the capsule.", Required: true},
			},
			ReturnType: tool.ArgTypeMap,
			Example:    `capsule.Read("capsule/aeiou@2")`,
		},
		Func:          toolReadCapsule,
		RequiresTrust: false,
		Effects:       []string{"readonly"},
	},
	{
		Spec: tool.ToolSpec{
			Name:        "GetLatest",
			Group:       Group,
			Description: "Gets the latest version of a capsule by its logical name (e.g., 'capsule/aeiou').",
			Args: []tool.ArgSpec{
				{Name: "name", Type: tool.ArgTypeString, Description: "The logical name of the capsule.", Required: true},
			},
			ReturnType: tool.ArgTypeMap,
			Example:    `capsule.GetLatest("capsule/aeiou")`,
		},
		Func:          toolGetLatestCapsule,
		RequiresTrust: false,
		Effects:       []string{"readonly"},
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Add",
			Group:       Group,
			Description: "Adds a new capsule to the runtime registry. Requires a privileged interpreter.",
			Args: []tool.ArgSpec{
				{Name: "capsuleData", Type: tool.ArgTypeMap, Description: "A map containing the capsule fields (name, version, content, etc.).", Required: true},
			},
			ReturnType: tool.ArgTypeNil,
			Example:    `capsule.Add({"name":"capsule/my-new-one","version":"1","content":"Hello"})`,
		},
		Func:          toolAddCapsule,
		RequiresTrust: true,
		Effects:       []string{"capsule:write"},
	},
}
