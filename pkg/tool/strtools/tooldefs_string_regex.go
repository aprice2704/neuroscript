// NeuroScript Version: 0.5.2
// File version: 1
// Purpose: Defines specifications for regular expression string tools.
// filename: pkg/tool/strtools/tooldefs_string_regex.go
// nlines: 55
// risk_rating: MEDIUM

package strtools

import (
	"github.com/aprice2704/neuroscript/pkg/policy/capability"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

var stringRegexToolsToRegister = []tool.ToolImplementation{
	{
		Spec: tool.ToolSpec{
			Name:        "MatchRegex",
			Group:       group,
			Description: "Checks if a string matches a regular expression. Requires 'str:use:regex' capability.",
			Category:    "String Regex",
			Args: []tool.ArgSpec{
				{Name: "pattern", Type: tool.ArgTypeString, Required: true, Description: "The regex pattern to match."},
				{Name: "input_string", Type: tool.ArgTypeString, Required: true, Description: "The string to check."},
			},
			ReturnType:      tool.ArgTypeBool,
			ReturnHelp:      "Returns true if the input_string matches the pattern, false otherwise.",
			Example:         `str.MatchRegex(pattern: "\\d{3}-\\d{2}-\\d{4}", input_string: "123-45-6789")`,
			ErrorConditions: "ErrInvalidArgument if the regex pattern is invalid.",
		},
		Func:         toolMatchRegex,
		RequiredCaps: []capability.Capability{capability.New(group, capability.VerbUse, "regex")},
	},
	{
		Spec: tool.ToolSpec{
			Name:        "FindAllRegex",
			Group:       group,
			Description: "Finds all non-overlapping occurrences of a regex pattern in a string. Requires 'str:use:regex' capability.",
			Category:    "String Regex",
			Args: []tool.ArgSpec{
				{Name: "pattern", Type: tool.ArgTypeString, Required: true, Description: "The regex pattern to find."},
				{Name: "input_string", Type: tool.ArgTypeString, Required: true, Description: "The string to search in."},
			},
			ReturnType:      tool.ArgTypeSliceString,
			ReturnHelp:      "Returns a list of all matching substrings.",
			Example:         `str.FindAllRegex(pattern: "\\w+", input_string: "hello world 123") // Returns ["hello", "world", "123"]`,
			ErrorConditions: "ErrInvalidArgument if the regex pattern is invalid.",
		},
		Func:         toolFindAllRegex,
		RequiredCaps: []capability.Capability{capability.New(group, capability.VerbUse, "regex")},
	},
	{
		Spec: tool.ToolSpec{
			Name:        "ReplaceRegex",
			Group:       group,
			Description: "Replaces all occurrences of a regex pattern in a string with a replacement string. Requires 'str:use:regex' capability.",
			Category:    "String Regex",
			Args: []tool.ArgSpec{
				{Name: "pattern", Type: tool.ArgTypeString, Required: true, Description: "The regex pattern to find."},
				{Name: "input_string", Type: tool.ArgTypeString, Required: true, Description: "The string to search in."},
				{Name: "replacement", Type: tool.ArgTypeString, Required: true, Description: "The string to replace matches with."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns a new string with all replacements made.",
			Example:         `str.ReplaceRegex(pattern: "\\s+", input_string: "a  b c", replacement: "-") // Returns "a-b-c"`,
			ErrorConditions: "ErrInvalidArgument if the regex pattern is invalid.",
		},
		Func:         toolReplaceRegex,
		RequiredCaps: []capability.Capability{capability.New(group, capability.VerbUse, "regex")},
	},
}
