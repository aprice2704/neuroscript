// NeuroScript Version: 0.3.1
// File version: 0.1.2 // Align Substring, ReplaceAll; Add ConcatStrings, SplitWords, LineCountString funcs; Correct Func pointers.
// Defines ToolImplementation structs for String tools.
// filename: pkg/core/tooldefs_string.go

package core

// stringToolsToRegister contains ToolImplementation definitions for String tools.
// These definitions are based on the tools previously registered in tools_string.go.
var stringToolsToRegister = []ToolImplementation{
	{
		Spec: ToolSpec{
			Name:        "Length",
			Description: "Returns the number of UTF-8 characters (runes) in a string.",
			Args: []ArgSpec{
				{Name: "input_string", Type: ArgTypeString, Required: true, Description: "The string to measure."},
			},
			ReturnType: ArgTypeInt,
		},
		Func: toolStringLength,
	},
	{
		Spec: ToolSpec{
			Name:        "Substring", // Renamed for clarity if preferred, or keep "Substring"
			Description: "Returns a portion of the string (rune-based indexing), from start_index for a given length.",
			Args: []ArgSpec{
				{Name: "input_string", Type: ArgTypeString, Required: true, Description: "The string to take a substring from."},
				{Name: "start_index", Type: ArgTypeInt, Required: true, Description: "0-based start index (inclusive)."},
				{Name: "length", Type: ArgTypeInt, Required: true, Description: "Number of characters to extract."},
			},
			ReturnType: ArgTypeString,
		},
		Func: toolStringSubstring,
	},
	{
		Spec: ToolSpec{
			Name:        "ToUpper", // Renamed for clarity
			Description: "Converts a string to uppercase.",
			Args: []ArgSpec{
				{Name: "input_string", Type: ArgTypeString, Required: true, Description: "The string to convert."},
			},
			ReturnType: ArgTypeString,
		},
		Func: toolStringToUpper,
	},
	{
		Spec: ToolSpec{
			Name:        "ToLower", // Renamed for clarity
			Description: "Converts a string to lowercase.",
			Args: []ArgSpec{
				{Name: "input_string", Type: ArgTypeString, Required: true, Description: "The string to convert."},
			},
			ReturnType: ArgTypeString,
		},
		Func: toolStringToLower,
	},
	{
		Spec: ToolSpec{
			Name:        "TrimSpace", // Renamed for clarity
			Description: "Removes leading and trailing whitespace from a string.",
			Args: []ArgSpec{
				{Name: "input_string", Type: ArgTypeString, Required: true, Description: "The string to trim."},
			},
			ReturnType: ArgTypeString,
		},
		Func: toolStringTrimSpace,
	},
	{
		Spec: ToolSpec{
			Name:        "Split", // Renamed for clarity
			Description: "Splits a string by a delimiter.",
			Args: []ArgSpec{
				{Name: "input_string", Type: ArgTypeString, Required: true, Description: "The string to split."},
				{Name: "delimiter", Type: ArgTypeString, Required: true, Description: "The delimiter string."},
			},
			ReturnType: ArgTypeSliceString,
		},
		Func: toolStringSplit,
	},
	{
		Spec: ToolSpec{
			Name:        "SplitWords", // Renamed for clarity
			Description: "Splits a string into words based on whitespace.",
			Args: []ArgSpec{
				{Name: "input_string", Type: ArgTypeString, Required: true, Description: "The string to split into words."},
			},
			ReturnType: ArgTypeSliceString,
		},
		Func: toolSplitWords, // Assumes toolSplitWords will be defined in pkg/core/tools_string.go
	},
	{
		Spec: ToolSpec{
			Name:        "Join", // Renamed for clarity
			Description: "Joins elements of a list of strings with a separator.",
			Args: []ArgSpec{
				{Name: "string_list", Type: ArgTypeSliceString, Required: true, Description: "List of strings to join."},
				{Name: "separator", Type: ArgTypeString, Required: true, Description: "String to place between elements."},
			},
			ReturnType: ArgTypeString,
		},
		Func: toolStringJoin, // Note: toolStringJoin in tools_string.go currently takes []interface{} and converts. ArgTypeSliceString should be fine.
	},
	{
		Spec: ToolSpec{
			Name:        "Concat", // New tool, was toolStringConcat
			Description: "Concatenates a list of strings without a separator.",
			Args: []ArgSpec{
				{Name: "strings_list", Type: ArgTypeSliceString, Required: true, Description: "List of strings to concatenate."},
			},
			ReturnType: ArgTypeString,
		},
		Func: toolStringConcat,
	},
	{
		Spec: ToolSpec{
			Name:        "Replace", // Was ReplaceAll
			Description: "Replaces occurrences of a substring with another, up to a specified count.",
			Args: []ArgSpec{
				{Name: "input_string", Type: ArgTypeString, Required: true, Description: "The string to perform replacements on."},
				{Name: "old_substring", Type: ArgTypeString, Required: true, Description: "The substring to be replaced."},
				{Name: "new_substring", Type: ArgTypeString, Required: true, Description: "The substring to replace with."},
				{Name: "count", Type: ArgTypeInt, Required: true, Description: "Maximum number of replacements. Use -1 for all."},
			},
			ReturnType: ArgTypeString,
		},
		Func: toolStringReplace,
	},
	{
		Spec: ToolSpec{
			Name:        "Contains", // Renamed for clarity
			Description: "Checks if a string contains a substring.",
			Args: []ArgSpec{
				{Name: "input_string", Type: ArgTypeString, Required: true, Description: "The string to check."},
				{Name: "substring", Type: ArgTypeString, Required: true, Description: "The substring to search for."},
			},
			ReturnType: ArgTypeBool,
		},
		Func: toolStringContains,
	},
	{
		Spec: ToolSpec{
			Name:        "HasPrefix", // Renamed for clarity
			Description: "Checks if a string starts with a prefix.",
			Args: []ArgSpec{
				{Name: "input_string", Type: ArgTypeString, Required: true, Description: "The string to check."},
				{Name: "prefix", Type: ArgTypeString, Required: true, Description: "The prefix to check for."},
			},
			ReturnType: ArgTypeBool,
		},
		Func: toolStringHasPrefix,
	},
	{
		Spec: ToolSpec{
			Name:        "HasSuffix", // Renamed for clarity
			Description: "Checks if a string ends with a suffix.",
			Args: []ArgSpec{
				{Name: "input_string", Type: ArgTypeString, Required: true, Description: "The string to check."},
				{Name: "suffix", Type: ArgTypeString, Required: true, Description: "The suffix to check for."},
			},
			ReturnType: ArgTypeBool,
		},
		Func: toolStringHasSuffix,
	},
	{
		Spec: ToolSpec{
			Name:        "LineCount", // Renamed for clarity
			Description: "Counts the number of lines in the given string content.",
			Args: []ArgSpec{
				{Name: "content_string", Type: ArgTypeString, Required: true, Description: "The string content in which to count lines."},
			},
			ReturnType: ArgTypeInt,
		},
		Func: toolLineCountString, // Assumes toolLineCountString will be defined in pkg/core/tools_string.go
	},
}
