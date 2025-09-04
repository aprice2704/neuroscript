// NeuroScript Version: 0.7.0
// File version: 1
// Purpose: Defines the tool specifications for the 'aeiou' toolset.
// filename: pkg/tool/aeiou/tooldefs_aeiou.go
// nlines: 34
// risk_rating: LOW
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
			Description: "Constructs a valid, multi-line AEIOU v3 envelope string.",
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
