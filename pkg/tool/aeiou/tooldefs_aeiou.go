// :: product: FDM/NS
// :: majorVersion: 0
// :: fileVersion: 2
// :: description: Updated description to reflect AEIOU v4.
// :: latestChange: Changed "v3" to "v4" in Description.
// :: filename: pkg/tool/aeiou/tooldefs_aeiou.go
// :: serialization: go
package aeiou

import (
	"github.com/aprice2704/neuroscript/pkg/tool"
)

const Group = "aeiou"

var AeiouToolsToRegister = []tool.ToolImplementation{
	{
		Spec: tool.ToolSpec{
			Name:        "ComposeEnvelope",
			Group:       Group,
			Description: "Constructs a valid, multi-line AEIOU v4 envelope string.",
			Args: []tool.ArgSpec{
				{Name: "userdata", Type: tool.ArgTypeString, Required: true, Description: "The JSON string for the USERDATA section."},
				{Name: "actions", Type: tool.ArgTypeString, Required: true, Description: "The NeuroScript command block for the ACTIONS section."},
				{Name: "scratchpad", Type: tool.ArgTypeString, Required: false, Description: "Optional: Content for the SCRATCHPAD section."},
				{Name: "output", Type: tool.ArgTypeString, Required: false, Description: "Optional: Content for the OUTPUT section."},
			},
			ReturnType: tool.ArgTypeString,
		},
		Func: envelopeToolFunc,
	},
}
