// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 2
// :: description: Defines tool specifications for ANSI colorization and manipulation tools.
// :: latestChange: Expanded Colorize help text to list all supported tags.
// :: filename: pkg/tool/strtools/tooldefs_string_ansi.go
// :: serialization: go

package strtools

import (
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// stringAnsiToolsToRegister contains ToolImplementation definitions for ANSI string tools.
var stringAnsiToolsToRegister = []tool.ToolImplementation{
	{
		Spec: tool.ToolSpec{
			Name:  "Color",
			Group: group,
			Description: "Replaces supported color tags with ANSI escape sequences.\n" +
				"Supported Tags:\n" +
				"  Resets: [reset], [default], [bg-default]\n" +
				"  Styles: [bold], [dim], [italic], [underline], [blink], [reverse], [hidden], [strike]\n" +
				"  Colors: [black], [red], [green], [yellow], [blue], [magenta], [cyan], [white]\n" +
				"  Bright: [bright-black] (or [gray]), [bright-red], [bright-green], etc.\n" +
				"  Backgrounds: [bg-red], [bg-bright-red], etc.",
			Category: "String Formatting",
			Args: []tool.ArgSpec{
				{Name: "input_string", Type: tool.ArgTypeString, Required: true, Description: "The string containing color tags to process."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns the string with tags replaced by ANSI codes.",
			Example:         `tool.Colorize("[bold][blue]Info:[reset] [bright-red]Error details")`,
			ErrorConditions: "Returns `ErrArgumentMismatch` if the wrong number of arguments is provided. Returns `ErrInvalidArgument` (with `ErrorCodeType`) if `input_string` is not a string.",
		},
		Func: toolStringColorize,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "StripAnsi",
			Group:       group,
			Description: "Removes all ANSI escape sequences from a string.",
			Category:    "String Formatting",
			Args: []tool.ArgSpec{
				{Name: "input_string", Type: tool.ArgTypeString, Required: true, Description: "The string to strip ANSI codes from."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns the clean string without ANSI codes.",
			Example:         `tool.StripAnsi("\x1b[31mError\x1b[0m") // Returns "Error"`,
			ErrorConditions: "Returns `ErrArgumentMismatch` if the wrong number of arguments is provided. Returns `ErrInvalidArgument` (with `ErrorCodeType`) if `input_string` is not a string.",
		},
		Func: toolStringStripAnsi,
	},
}
