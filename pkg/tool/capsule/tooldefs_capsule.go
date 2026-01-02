// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 8
// :: description: Tool definitions for the capsule toolset, providing List, Read, GetLatest, Add, and Parse.
// :: latestChange: Added Parse tool definition to allow metadata extraction from raw content strings.
// :: filename: pkg/tool/capsule/tooldefs.go
// :: serialization: go

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
	{
		Spec: tool.ToolSpec{
			Name:        "Parse",
			Group:       Group,
			Description: "Parses a raw capsule string (Markdown or NeuroScript) and returns its metadata fields.",
			Args: []tool.ArgSpec{
				{Name: "content", Type: tool.ArgTypeString, Required: true},
			},
			ReturnType: tool.ArgTypeMap,
		},
		Func:          parseCapsuleFunc,
		RequiresTrust: false,
	},
}
