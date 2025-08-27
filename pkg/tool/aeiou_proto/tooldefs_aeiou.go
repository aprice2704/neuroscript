// NeuroScript Version: 0.7.0
// File version: 4
// Purpose: Adds the missing tool definition for 'tool.aeiou.magic' to fix the compiler error.
// filename: pkg/tool/aeiou_proto/tooldefs_aeiou.go
// nlines: 147
// risk_rating: LOW

package aeiou_proto

import (
	"github.com/aprice2704/neuroscript/pkg/tool"
)

const group = "aeiou"

var aeiouToolsToRegister = []tool.ToolImplementation{
	{
		Spec: tool.ToolSpec{
			Name:        "new",
			Group:       group,
			Description: "Creates a new, empty AEIOU envelope object and returns a handle to it.",
			Category:    "AEIOU Operations",
			Args:        []tool.ArgSpec{},
			ReturnType:  tool.ArgTypeString, // Handles are strings
			ReturnHelp:  "Returns a string handle to the newly created envelope.",
			Example:     `set handle = tool.aeiou.new()`,
		},
		Func: toolAeiouNew,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "parse",
			Group:       group,
			Description: "Parses a raw string payload into an envelope object and returns a handle.",
			Category:    "AEIOU Operations",
			Args: []tool.ArgSpec{
				{Name: "payload", Type: tool.ArgTypeString, Required: true, Description: "The raw envelope string to parse."},
			},
			ReturnType:      tool.ArgTypeString, // Handles are strings
			ReturnHelp:      "Returns a handle to the parsed envelope, or an error if parsing fails.",
			Example:         `set payload = tool.aeiou.compose(handle)\nset parsed_handle = tool.aeiou.parse(payload)`,
			ErrorConditions: "Returns an error if the payload is malformed or no V2 envelope is found.",
		},
		Func: toolAeiouParse,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "get_section",
			Group:       group,
			Description: "Retrieves the content of a specific section from an envelope.",
			Category:    "AEIOU Operations",
			Args: []tool.ArgSpec{
				{Name: "handle", Type: tool.ArgTypeString, Required: true, Description: "The handle of the envelope."},
				{Name: "section_name", Type: tool.ArgTypeString, Required: true, Description: "The name of the section (e.g., 'ACTIONS', 'EVENTS')."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns the content of the specified section.",
			Example:         `set content = tool.aeiou.get_section(handle, "ACTIONS")`,
			ErrorConditions: "Returns an error if the handle is invalid or the section name is unknown.",
		},
		Func: toolAeiouGetSection,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "set_section",
			Group:       group,
			Description: "Sets the content of a specific section in an envelope.",
			Category:    "AEIOU Operations",
			Args: []tool.ArgSpec{
				{Name: "handle", Type: tool.ArgTypeString, Required: true, Description: "The handle of the envelope."},
				{Name: "section_name", Type: tool.ArgTypeString, Required: true, Description: "The name of the section to set."},
				{Name: "content", Type: tool.ArgTypeString, Required: true, Description: "The new content for the section."},
			},
			ReturnType:      tool.ArgTypeNil,
			ReturnHelp:      "Returns nil on success.",
			Example:         `call tool.aeiou.set_section(handle, "ACTIONS", "command {}")`,
			ErrorConditions: "Returns an error if the handle is invalid or the section name is unknown.",
		},
		Func: toolAeiouSetSection,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "compose",
			Group:       group,
			Description: "Composes an envelope object into its final, V2 checksummed string representation.",
			Category:    "AEIOU Operations",
			Args: []tool.ArgSpec{
				{Name: "handle", Type: tool.ArgTypeString, Required: true, Description: "The handle of the envelope to compose."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns the complete, ready-to-send V2 envelope payload as a string.",
			Example:         `set payload = tool.aeiou.compose(handle)`,
			ErrorConditions: "Returns an error if the handle is invalid.",
		},
		Func: toolAeiouCompose,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "validate",
			Group:       group,
			Description: "Validates the internal consistency of an envelope's contents (currently a placeholder).",
			Category:    "AEIOU Operations",
			Args: []tool.ArgSpec{
				{Name: "handle", Type: tool.ArgTypeString, Required: true, Description: "The handle of the envelope to validate."},
			},
			ReturnType:      tool.ArgTypeSliceString,
			ReturnHelp:      "Returns a list of validation error messages. An empty list means the envelope is valid.",
			Example:         `set errors = tool.aeiou.validate(handle)`,
			ErrorConditions: "Returns an error if the handle is invalid.",
		},
		Func: toolAeiouValidate,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "magic",
			Group:       group,
			Description: "Creates a V2 magic string for use in envelope control (e.g., loop signals).",
			Category:    "AEIOU Operations",
			Args: []tool.ArgSpec{
				{Name: "type", Type: tool.ArgTypeString, Required: true, Description: "The marker type (e.g., 'LOOP', 'DIAGNOSTIC')."},
				{Name: "payload", Type: tool.ArgTypeMap, Required: false, Description: "An optional JSON object to include as the payload."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns the formatted V2 magic string.",
			Example:         `set signal = tool.aeiou.magic("LOOP", {"control":"continue"})`,
			ErrorConditions: "Returns an error if the payload is not valid JSON.",
		},
		Func: toolAeiouMagic,
	},
}
