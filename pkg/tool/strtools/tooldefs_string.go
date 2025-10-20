// NeuroScript Version: 0.3.1
// File version: 0.1.8
// Purpose: Changed 'Join' and 'Concat' ArgType from SliceString to ArgTypeAny to allow generic lists.
// filename: pkg/tool/strtools/tooldefs_string.go
// nlines: 198
// risk_rating: LOW

package strtools

import (
	"github.com/aprice2704/neuroscript/pkg/tool"
)

const group = "str"

// stringToolsToRegister contains ToolImplementation definitions for String tools.
// These definitions are based on the tools previously registered in tools_string.go.
var stringToolsToRegister = []tool.ToolImplementation{
	{
		Spec: tool.ToolSpec{
			Name:        "Length",
			Group:       group,
			Description: "Returns the number of UTF-8 characters (runes) in a string.",
			Category:    "String Operations",
			Args: []tool.ArgSpec{
				{Name: "input_string", Type: tool.ArgTypeString, Required: true, Description: "The string to measure."},
			},
			ReturnType:      tool.ArgTypeInt,
			ReturnHelp:      "Returns an integer representing the number of runes in the input string.",
			Example:         `tool.Length("hello") // Returns 5`,
			ErrorConditions: "Returns `ErrArgumentMismatch` if the wrong number of arguments is provided. Returns `ErrInvalidArgument` (with `ErrorCodeType`) if `input_string` is not a string.",
		},
		Func: toolStringLength,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Substring",
			Group:       group,
			Description: "Returns a portion of the string (rune-based indexing), from start_index for a given length.",
			Category:    "String Operations",
			Args: []tool.ArgSpec{
				{Name: "input_string", Type: tool.ArgTypeString, Required: true, Description: "The string to take a substring from."},
				{Name: "start_index", Type: tool.ArgTypeInt, Required: true, Description: "0-based start index (inclusive)."},
				{Name: "length", Type: tool.ArgTypeInt, Required: true, Description: "Number of characters to extract."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns the specified substring (rune-based). Returns an empty string if length is zero or if start_index is out of bounds (after clamping). Gracefully handles out-of-bounds for non-negative start_index and length by returning available characters.",
			Example:         `tool.Substring("hello world", 6, 5) // Returns "world"`,
			ErrorConditions: "Returns `ErrArgumentMismatch` if the wrong number of arguments is provided. Returns `ErrInvalidArgument` (with `ErrorCodeType`) if arguments are not of the correct type. Returns `ErrListIndexOutOfBounds` (with `ErrorCodeBounds`) if `start_index` or `length` are negative.",
		},
		Func: toolStringSubstring,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "ToUpper",
			Group:       group,
			Description: "Converts a string to uppercase.",
			Category:    "String Operations",
			Args: []tool.ArgSpec{
				{Name: "input_string", Type: tool.ArgTypeString, Required: true, Description: "The string to convert."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns the uppercase version of the input string.",
			Example:         `tool.ToUpper("hello") // Returns "HELLO"`,
			ErrorConditions: "Returns `ErrArgumentMismatch` if the wrong number of arguments is provided. Returns `ErrInvalidArgument` (with `ErrorCodeType`) if `input_string` is not a string.",
		},
		Func: toolStringToUpper,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "ToLower",
			Group:       group,
			Description: "Converts a string to lowercase.",
			Category:    "String Operations",
			Args: []tool.ArgSpec{
				{Name: "input_string", Type: tool.ArgTypeString, Required: true, Description: "The string to convert."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns the lowercase version of the input string.",
			Example:         `tool.ToLower("HELLO") // Returns "hello"`,
			ErrorConditions: "Returns `ErrArgumentMismatch` if the wrong number of arguments is provided. Returns `ErrInvalidArgument` (with `ErrorCodeType`) if `input_string` is not a string.",
		},
		Func: toolStringToLower,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "TrimSpace",
			Group:       group,
			Description: "Removes leading and trailing whitespace from a string.",
			Category:    "String Operations",
			Args: []tool.ArgSpec{
				{Name: "input_string", Type: tool.ArgTypeString, Required: true, Description: "The string to trim."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns the string with leading and trailing whitespace removed.",
			Example:         `tool.TrimSpace("  hello  ") // Returns "hello"`,
			ErrorConditions: "Returns `ErrArgumentMismatch` if the wrong number of arguments is provided. Returns `ErrInvalidArgument` (with `ErrorCodeType`) if `input_string` is not a string.",
		},
		Func: toolStringTrimSpace,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Split",
			Group:       group,
			Description: "Splits a string by a delimiter.",
			Category:    "String Operations",
			Args: []tool.ArgSpec{
				{Name: "input_string", Type: tool.ArgTypeString, Required: true, Description: "The string to split."},
				{Name: "delimiter", Type: tool.ArgTypeString, Required: true, Description: "The delimiter string."},
			},
			ReturnType:      tool.ArgTypeSliceString,
			ReturnHelp:      "Returns a slice of strings after splitting the input string by the delimiter.",
			Example:         `tool.Split("apple,banana,orange", ",") // Returns ["apple", "banana", "orange"]`,
			ErrorConditions: "Returns `ErrArgumentMismatch` if the wrong number of arguments is provided. Returns `ErrInvalidArgument` (with `ErrorCodeType`) if `input_string` or `delimiter` are not strings.",
		},
		Func: toolStringSplit,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "SplitWords",
			Group:       group,
			Description: "Splits a string into words based on whitespace.",
			Category:    "String Operations",
			Args: []tool.ArgSpec{
				{Name: "input_string", Type: tool.ArgTypeString, Required: true, Description: "The string to split into words."},
			},
			ReturnType:      tool.ArgTypeSliceString,
			ReturnHelp:      "Returns a slice of strings, where each string is a word from the input string, with whitespace removed. Multiple spaces are treated as a single delimiter.",
			Example:         `tool.SplitWords("hello world  example") // Returns ["hello", "world", "example"]`,
			ErrorConditions: "Returns `ErrArgumentMismatch` if the wrong number of arguments is provided. Returns `ErrInvalidArgument` (with `ErrorCodeType`) if `input_string` is not a string.",
		},
		Func: toolSplitWords,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Join",
			Group:       group,
			Description: "Joins elements of a list of strings with a separator.",
			Category:    "String Operations",
			Args: []tool.ArgSpec{
				// FIX: Changed tool.ArgTypeSliceString to tool.ArgTypeAny
				{Name: "string_list", Type: tool.ArgTypeAny, Required: true, Description: "List of strings to join."},
				{Name: "separator", Type: tool.ArgTypeString, Required: true, Description: "String to place between elements."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns a single string created by joining the elements of string_list with the separator.",
			Example:         `tool.Join(["apple", "banana"], ", ") // Returns "apple, banana"`,
			ErrorConditions: "Returns `ErrArgumentMismatch` if the wrong number of arguments is provided. Returns `ErrInvalidArgument` (with `ErrorCodeType`) if `string_list` is not a list or `separator` is not a string.",
		},
		Func: toolStringJoin,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Concat",
			Group:       group,
			Description: "Concatenates a list of strings without a separator.",
			Category:    "String Operations",
			Args: []tool.ArgSpec{
				// FIX: Changed tool.ArgTypeSliceString to tool.ArgTypeAny
				{Name: "strings_list", Type: tool.ArgTypeAny, Required: true, Description: "List of strings to concatenate."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns a single string by concatenating all strings in the strings_list.",
			Example:         `tool.Concat(["hello", " ", "world"]) // Returns "hello world"`,
			ErrorConditions: "Returns `ErrArgumentMismatch` if the wrong number of arguments is provided. Returns `ErrInvalidArgument` (with `ErrorCodeType`) if `strings_list` is not a list.",
		},
		Func: toolStringConcat,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Replace",
			Group:       group,
			Description: "Replaces occurrences of a substring with another, up to a specified count.",
			Category:    "String Operations",
			Args: []tool.ArgSpec{
				{Name: "input_string", Type: tool.ArgTypeString, Required: true, Description: "The string to perform replacements on."},
				{Name: "old_substring", Type: tool.ArgTypeString, Required: true, Description: "The substring to be replaced."},
				{Name: "new_substring", Type: tool.ArgTypeString, Required: true, Description: "The substring to replace with."},
				{Name: "count", Type: tool.ArgTypeInt, Required: true, Description: "Maximum number of replacements. Use -1 for all."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns the string with specified replacements made.",
			Example:         `tool.Replace("ababab", "ab", "cd", 2) // Returns "cdcdab"`,
			ErrorConditions: "Returns `ErrArgumentMismatch` if the wrong number of arguments is provided. Returns `ErrInvalidArgument` (with `ErrorCodeType`) if `input_string`, `old_substring`, or `new_substring` are not strings, or if `count` is not an integer.",
		},
		Func: toolStringReplace,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Contains",
			Group:       group,
			Description: "Checks if a string contains a substring.",
			Category:    "String Operations",
			Args: []tool.ArgSpec{
				{Name: "input_string", Type: tool.ArgTypeString, Required: true, Description: "The string to check."},
				{Name: "substring", Type: tool.ArgTypeString, Required: true, Description: "The substring to search for."},
			},
			ReturnType:      tool.ArgTypeBool,
			ReturnHelp:      "Returns true if the input_string contains the substring, false otherwise.",
			Example:         `tool.Contains("hello world", "world") // Returns true`,
			ErrorConditions: "Returns `ErrArgumentMismatch` if the wrong number of arguments is provided. Returns `ErrInvalidArgument` (with `ErrorCodeType`) if `input_string` or `substring` are not strings.",
		},
		Func: toolStringContains,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "HasPrefix",
			Group:       group,
			Description: "Checks if a string starts with a prefix.",
			Category:    "String Operations",
			Args: []tool.ArgSpec{
				{Name: "input_string", Type: tool.ArgTypeString, Required: true, Description: "The string to check."},
				{Name: "prefix", Type: tool.ArgTypeString, Required: true, Description: "The prefix to check for."},
			},
			ReturnType:      tool.ArgTypeBool,
			ReturnHelp:      "Returns true if the input_string starts with the prefix, false otherwise.",
			Example:         `tool.HasPrefix("filename.txt", "filename") // Returns true`,
			ErrorConditions: "Returns `ErrArgumentMismatch` if the wrong number of arguments is provided. Returns `ErrInvalidArgument` (with `ErrorCodeType`) if `input_string` or `prefix` are not strings.",
		},
		Func: toolStringHasPrefix,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "HasSuffix",
			Group:       group,
			Description: "Checks if a string ends with a suffix.",
			Category:    "String Operations",
			Args: []tool.ArgSpec{
				{Name: "input_string", Type: tool.ArgTypeString, Required: true, Description: "The string to check."},
				{Name: "suffix", Type: tool.ArgTypeString, Required: true, Description: "The suffix to check for."},
			},
			ReturnType:      tool.ArgTypeBool,
			ReturnHelp:      "Returns true if the input_string ends with the suffix, false otherwise.",
			Example:         `tool.HasSuffix("document.doc", ".doc") // Returns true`,
			ErrorConditions: "Returns `ErrArgumentMismatch` if the wrong number of arguments is provided. Returns `ErrInvalidArgument` (with `ErrorCodeType`) if `input_string` or `suffix` are not strings.",
		},
		Func: toolStringHasSuffix,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "LineCount",
			Group:       group,
			Description: "Counts the number of lines in the given string content.",
			Category:    "String Operations",
			Args: []tool.ArgSpec{
				{Name: "content_string", Type: tool.ArgTypeString, Required: true, Description: "The string content in which to count lines."},
			},
			ReturnType: tool.ArgTypeInt,
			ReturnHelp: "Returns an integer representing the number of lines in the string. Lines are typically separated by '\\n'. An empty string results in 0 lines. If the string is not empty and does not end with a newline, the last line is still counted.",
			Example: `tool.LineCount("line1\nline2\nline3") // Returns 3
tool.LineCount("line1\nline2") // Returns 2
tool.LineCount("") // Returns 0`,
			ErrorConditions: "Returns `ErrArgumentMismatch` if the wrong number of arguments is provided. Returns `ErrInvalidArgument` (with `ErrorCodeType`) if `content_string` is not a string.",
		},
		Func: toolLineCountString,
	},
}
