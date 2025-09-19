// NeuroScript Version: 0.7.2
// File version: 7
// Purpose: Updates the 'Add' tool's return type to map.
// filename: pkg/tool/capsule/tooldefs.go
// nlines: 78
// risk_rating: HIGH
package capsule

import (
	"github.com/aprice2704/neuroscript/pkg/capability"
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
		},
		Func:          listCapsulesFunc,
		RequiresTrust: false,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Read",
			Group:       Group,
			Description: "Reads a capsule by its full ID ('name@version') or the latest version by name.",
			Args: []tool.ArgSpec{
				{Name: "id", Type: tool.ArgTypeString, Required: true},
			},
			ReturnType: tool.ArgTypeMap,
		},
		Func:          readCapsuleFunc,
		RequiresTrust: false,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "GetLatest",
			Group:       Group,
			Description: "Gets the latest version of a capsule by its logical name.",
			Args: []tool.ArgSpec{
				{Name: "name", Type: tool.ArgTypeString, Required: true},
			},
			ReturnType: tool.ArgTypeMap,
		},
		Func:          getLatestCapsuleFunc,
		RequiresTrust: false,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Add",
			Group:       Group,
			Description: "Adds a new capsule to the runtime registry by parsing its content. Requires a privileged interpreter.",
			Args: []tool.ArgSpec{
				{Name: "capsuleContent", Type: tool.ArgTypeString, Required: true},
			},
			ReturnType: tool.ArgTypeMap,
		},
		Func:          addCapsuleFunc,
		RequiresTrust: true,
		RequiredCaps: []capability.Capability{
			{Resource: "capsule", Verbs: []string{"write"}, Scopes: []string{"*"}},
		},
	},
}
