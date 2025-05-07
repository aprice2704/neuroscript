// NeuroScript Version: 0.3.1
// File version: 0.1.1 // Reformatted for readability
// Defines ToolImplementation structs for String tools.
// filename: pkg/core/tooldefs_string.go

package core

// stringToolsToRegister contains ToolImplementation definitions for String tools.
// These definitions are based on the tools previously registered in tools_string.go.
var stringToolsToRegister = []ToolImplementation{
	{
		Spec: ToolSpec{
			Name:        "StringLength",
			Description: "Returns the number of UTF-8 characters (runes) in a string.",
			Args: []ArgSpec{
				{Name: "input", Type: ArgTypeString, Required: true},
			},
			ReturnType: ArgTypeInt,
		},
		Func: toolStringLength, // Assumes toolStringLength is defined in pkg/core/tools_string.go
	},
	{
		Spec: ToolSpec{
			Name:        "Substring",
			Description: "Returns a portion of the string (rune-based indexing).",
			Args: []ArgSpec{
				{Name: "input", Type: ArgTypeString, Required: true},
				{Name: "start", Type: ArgTypeInt, Required: true, Description: "0-based start index (inclusive)."},
				{Name: "end", Type: ArgTypeInt, Required: true, Description: "0-based end index (exclusive)."},
			},
			ReturnType: ArgTypeString,
		},
		Func: toolSubstring, // Assumes toolSubstring is defined in pkg/core/tools_string.go
	},
	{
		Spec: ToolSpec{
			Name:        "ToUpper",
			Description: "Converts a string to uppercase.",
			Args: []ArgSpec{
				{Name: "input", Type: ArgTypeString, Required: true},
			},
			ReturnType: ArgTypeString,
		},
		Func: toolToUpper, // Assumes toolToUpper is defined in pkg/core/tools_string.go
	},
	{
		Spec: ToolSpec{
			Name:        "ToLower",
			Description: "Converts a string to lowercase.",
			Args: []ArgSpec{
				{Name: "input", Type: ArgTypeString, Required: true},
			},
			ReturnType: ArgTypeString,
		},
		Func: toolToLower, // Assumes toolToLower is defined in pkg/core/tools_string.go
	},
	{
		Spec: ToolSpec{
			Name:        "TrimSpace",
			Description: "Removes leading and trailing whitespace from a string.",
			Args: []ArgSpec{
				{Name: "input", Type: ArgTypeString, Required: true},
			},
			ReturnType: ArgTypeString,
		},
		Func: toolTrimSpace, // Assumes toolTrimSpace is defined in pkg/core/tools_string.go
	},
	{
		Spec: ToolSpec{
			Name:        "SplitString",
			Description: "Splits a string by a delimiter.",
			Args: []ArgSpec{
				{Name: "input", Type: ArgTypeString, Required: true},
				{Name: "delimiter", Type: ArgTypeString, Required: true},
			},
			ReturnType: ArgTypeSliceString,
		},
		Func: toolSplitString, // Assumes toolSplitString is defined in pkg/core/tools_string.go
	},
	{
		Spec: ToolSpec{
			Name:        "SplitWords",
			Description: "Splits a string into words based on whitespace.",
			Args: []ArgSpec{
				{Name: "input", Type: ArgTypeString, Required: true},
			},
			ReturnType: ArgTypeSliceString,
		},
		Func: toolSplitWords, // Assumes toolSplitWords is defined in pkg/core/tools_string.go
	},
	{
		Spec: ToolSpec{
			Name:        "JoinStrings",
			Description: "Joins elements of a list (converting each to string) with a separator.",
			Args: []ArgSpec{
				{Name: "input_slice", Type: ArgTypeSliceAny, Required: true, Description: "List of items to join."},
				{Name: "separator", Type: ArgTypeString, Required: true, Description: "String to place between elements."},
			},
			ReturnType: ArgTypeString,
		},
		Func: toolJoinStrings, // Assumes toolJoinStrings is defined in pkg/core/tools_string.go
	},
	{
		Spec: ToolSpec{
			Name:        "ReplaceAll",
			Description: "Replaces all occurrences of a substring with another.",
			Args: []ArgSpec{
				{Name: "input", Type: ArgTypeString, Required: true},
				{Name: "old", Type: ArgTypeString, Required: true},
				{Name: "new", Type: ArgTypeString, Required: true},
			},
			ReturnType: ArgTypeString,
		},
		Func: toolReplaceAll, // Assumes toolReplaceAll is defined in pkg/core/tools_string.go
	},
	{
		Spec: ToolSpec{
			Name:        "Contains",
			Description: "Checks if a string contains a substring.",
			Args: []ArgSpec{
				{Name: "input", Type: ArgTypeString, Required: true},
				{Name: "substring", Type: ArgTypeString, Required: true},
			},
			ReturnType: ArgTypeBool,
		},
		Func: toolContains, // Assumes toolContains is defined in pkg/core/tools_string.go
	},
	{
		Spec: ToolSpec{
			Name:        "HasPrefix",
			Description: "Checks if a string starts with a prefix.",
			Args: []ArgSpec{
				{Name: "input", Type: ArgTypeString, Required: true},
				{Name: "prefix", Type: ArgTypeString, Required: true},
			},
			ReturnType: ArgTypeBool,
		},
		Func: toolHasPrefix, // Assumes toolHasPrefix is defined in pkg/core/tools_string.go
	},
	{
		Spec: ToolSpec{
			Name:        "HasSuffix",
			Description: "Checks if a string ends with a suffix.",
			Args: []ArgSpec{
				{Name: "input", Type: ArgTypeString, Required: true},
				{Name: "suffix", Type: ArgTypeString, Required: true},
			},
			ReturnType: ArgTypeBool,
		},
		Func: toolHasSuffix, // Assumes toolHasSuffix is defined in pkg/core/tools_string.go
	},
	{
		Spec: ToolSpec{
			Name:        "LineCountString",
			Description: "Counts the number of lines in the given string content.",
			Args: []ArgSpec{
				{Name: "content", Type: ArgTypeString, Required: true, Description: "The string content in which to count lines."},
			},
			ReturnType: ArgTypeInt,
		},
		Func: toolLineCountString, // Assumes toolLineCountString is defined in pkg/core/tools_string.go
	},
}
