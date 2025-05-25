// NeuroScript Version: 0.3.1
// File version: 0.1.3 // Populated Category, Example, ReturnHelp, ErrorConditions fields.
// Defines ToolImplementation structs for String tools.
// filename: pkg/core/tooldefs_string.go
// nlines: 200 // Approximate
// risk_rating: LOW

package core

// stringToolsToRegister contains ToolImplementation definitions for String tools.
// These definitions are based on the tools previously registered in tools_string.go.
var stringToolsToRegister = []ToolImplementation{
	{
		Spec: ToolSpec{
			Name:        "Length",
			Description: "Returns the number of UTF-8 characters (runes) in a string.",
			Category:    "String Manipulation",
			Args: []ArgSpec{
				{Name: "input_string", Type: ArgTypeString, Required: true, Description: "The string to measure."},
			},
			ReturnType:      ArgTypeInt,
			ReturnHelp:      "Returns an integer representing the number of runes in the input string.",
			Example:         `TOOL.Length(input_string: "hello") // Returns 5`,
			ErrorConditions: "ErrInvalidArgType if input_string is not a string; ErrMissingArg if input_string is not provided.",
		},
		Func: toolStringLength,
	},
	{
		Spec: ToolSpec{
			Name:        "Substring",
			Description: "Returns a portion of the string (rune-based indexing), from start_index for a given length.",
			Category:    "String Manipulation",
			Args: []ArgSpec{
				{Name: "input_string", Type: ArgTypeString, Required: true, Description: "The string to take a substring from."},
				{Name: "start_index", Type: ArgTypeInt, Required: true, Description: "0-based start index (inclusive)."},
				{Name: "length", Type: ArgTypeInt, Required: true, Description: "Number of characters to extract."},
			},
			ReturnType:      ArgTypeString,
			ReturnHelp:      "Returns the specified substring. Returns an empty string if length is zero. Handles out-of-bounds gracefully by returning available characters.",
			Example:         `TOOL.Substring(input_string: "hello world", start_index: 6, length: 5) // Returns "world"`,
			ErrorConditions: "ErrInvalidArgType if arguments are not of the correct type; ErrMissingArg if required arguments are not provided. Returns error if start_index or length are negative.",
		},
		Func: toolStringSubstring,
	},
	{
		Spec: ToolSpec{
			Name:        "ToUpper",
			Description: "Converts a string to uppercase.",
			Category:    "String Manipulation",
			Args: []ArgSpec{
				{Name: "input_string", Type: ArgTypeString, Required: true, Description: "The string to convert."},
			},
			ReturnType:      ArgTypeString,
			ReturnHelp:      "Returns the uppercase version of the input string.",
			Example:         `TOOL.ToUpper(input_string: "hello") // Returns "HELLO"`,
			ErrorConditions: "ErrInvalidArgType if input_string is not a string; ErrMissingArg if input_string is not provided.",
		},
		Func: toolStringToUpper,
	},
	{
		Spec: ToolSpec{
			Name:        "ToLower",
			Description: "Converts a string to lowercase.",
			Category:    "String Manipulation",
			Args: []ArgSpec{
				{Name: "input_string", Type: ArgTypeString, Required: true, Description: "The string to convert."},
			},
			ReturnType:      ArgTypeString,
			ReturnHelp:      "Returns the lowercase version of the input string.",
			Example:         `TOOL.ToLower(input_string: "HELLO") // Returns "hello"`,
			ErrorConditions: "ErrInvalidArgType if input_string is not a string; ErrMissingArg if input_string is not provided.",
		},
		Func: toolStringToLower,
	},
	{
		Spec: ToolSpec{
			Name:        "TrimSpace",
			Description: "Removes leading and trailing whitespace from a string.",
			Category:    "String Manipulation",
			Args: []ArgSpec{
				{Name: "input_string", Type: ArgTypeString, Required: true, Description: "The string to trim."},
			},
			ReturnType:      ArgTypeString,
			ReturnHelp:      "Returns the string with leading and trailing whitespace removed.",
			Example:         `TOOL.TrimSpace(input_string: "  hello  ") // Returns "hello"`,
			ErrorConditions: "ErrInvalidArgType if input_string is not a string; ErrMissingArg if input_string is not provided.",
		},
		Func: toolStringTrimSpace,
	},
	{
		Spec: ToolSpec{
			Name:        "Split",
			Description: "Splits a string by a delimiter.",
			Category:    "String Manipulation",
			Args: []ArgSpec{
				{Name: "input_string", Type: ArgTypeString, Required: true, Description: "The string to split."},
				{Name: "delimiter", Type: ArgTypeString, Required: true, Description: "The delimiter string."},
			},
			ReturnType:      ArgTypeSliceString,
			ReturnHelp:      "Returns a slice of strings after splitting the input string by the delimiter.",
			Example:         `TOOL.Split(input_string: "apple,banana,orange", delimiter: ",") // Returns ["apple", "banana", "orange"]`,
			ErrorConditions: "ErrInvalidArgType if arguments are not of the correct type; ErrMissingArg if required arguments are not provided.",
		},
		Func: toolStringSplit,
	},
	{
		Spec: ToolSpec{
			Name:        "SplitWords",
			Description: "Splits a string into words based on whitespace.",
			Category:    "String Manipulation",
			Args: []ArgSpec{
				{Name: "input_string", Type: ArgTypeString, Required: true, Description: "The string to split into words."},
			},
			ReturnType:      ArgTypeSliceString,
			ReturnHelp:      "Returns a slice of strings, where each string is a word from the input string.",
			Example:         `TOOL.SplitWords(input_string: "hello world example") // Returns ["hello", "world", "example"]`,
			ErrorConditions: "ErrInvalidArgType if input_string is not a string; ErrMissingArg if input_string is not provided.",
		},
		Func: toolSplitWords,
	},
	{
		Spec: ToolSpec{
			Name:        "Join",
			Description: "Joins elements of a list of strings with a separator.",
			Category:    "String Manipulation",
			Args: []ArgSpec{
				{Name: "string_list", Type: ArgTypeSliceString, Required: true, Description: "List of strings to join."},
				{Name: "separator", Type: ArgTypeString, Required: true, Description: "String to place between elements."},
			},
			ReturnType:      ArgTypeString,
			ReturnHelp:      "Returns a single string created by joining the elements of string_list with the separator.",
			Example:         `TOOL.Join(string_list: ["apple", "banana"], separator: ", ") // Returns "apple, banana"`,
			ErrorConditions: "ErrInvalidArgType if arguments are not of the correct type; ErrMissingArg if required arguments are not provided.",
		},
		Func: toolStringJoin,
	},
	{
		Spec: ToolSpec{
			Name:        "Concat",
			Description: "Concatenates a list of strings without a separator.",
			Category:    "String Manipulation",
			Args: []ArgSpec{
				{Name: "strings_list", Type: ArgTypeSliceString, Required: true, Description: "List of strings to concatenate."},
			},
			ReturnType:      ArgTypeString,
			ReturnHelp:      "Returns a single string by concatenating all strings in the strings_list.",
			Example:         `TOOL.Concat(strings_list: ["hello", " ", "world"]) // Returns "hello world"`,
			ErrorConditions: "ErrInvalidArgType if strings_list is not a slice of strings; ErrMissingArg if strings_list is not provided.",
		},
		Func: toolStringConcat,
	},
	{
		Spec: ToolSpec{
			Name:        "Replace",
			Description: "Replaces occurrences of a substring with another, up to a specified count.",
			Category:    "String Manipulation",
			Args: []ArgSpec{
				{Name: "input_string", Type: ArgTypeString, Required: true, Description: "The string to perform replacements on."},
				{Name: "old_substring", Type: ArgTypeString, Required: true, Description: "The substring to be replaced."},
				{Name: "new_substring", Type: ArgTypeString, Required: true, Description: "The substring to replace with."},
				{Name: "count", Type: ArgTypeInt, Required: true, Description: "Maximum number of replacements. Use -1 for all."},
			},
			ReturnType:      ArgTypeString,
			ReturnHelp:      "Returns the string with specified replacements made.",
			Example:         `TOOL.Replace(input_string: "ababab", old_substring: "ab", new_substring: "cd", count: 2) // Returns "cdcdab"`,
			ErrorConditions: "ErrInvalidArgType if arguments are not of the correct type; ErrMissingArg if required arguments are not provided.",
		},
		Func: toolStringReplace,
	},
	{
		Spec: ToolSpec{
			Name:        "Contains",
			Description: "Checks if a string contains a substring.",
			Category:    "String Manipulation",
			Args: []ArgSpec{
				{Name: "input_string", Type: ArgTypeString, Required: true, Description: "The string to check."},
				{Name: "substring", Type: ArgTypeString, Required: true, Description: "The substring to search for."},
			},
			ReturnType:      ArgTypeBool,
			ReturnHelp:      "Returns true if the input_string contains the substring, false otherwise.",
			Example:         `TOOL.Contains(input_string: "hello world", substring: "world") // Returns true`,
			ErrorConditions: "ErrInvalidArgType if arguments are not of the correct type; ErrMissingArg if required arguments are not provided.",
		},
		Func: toolStringContains,
	},
	{
		Spec: ToolSpec{
			Name:        "HasPrefix",
			Description: "Checks if a string starts with a prefix.",
			Category:    "String Manipulation",
			Args: []ArgSpec{
				{Name: "input_string", Type: ArgTypeString, Required: true, Description: "The string to check."},
				{Name: "prefix", Type: ArgTypeString, Required: true, Description: "The prefix to check for."},
			},
			ReturnType:      ArgTypeBool,
			ReturnHelp:      "Returns true if the input_string starts with the prefix, false otherwise.",
			Example:         `TOOL.HasPrefix(input_string: "filename.txt", prefix: "filename") // Returns true`,
			ErrorConditions: "ErrInvalidArgType if arguments are not of the correct type; ErrMissingArg if required arguments are not provided.",
		},
		Func: toolStringHasPrefix,
	},
	{
		Spec: ToolSpec{
			Name:        "HasSuffix",
			Description: "Checks if a string ends with a suffix.",
			Category:    "String Manipulation",
			Args: []ArgSpec{
				{Name: "input_string", Type: ArgTypeString, Required: true, Description: "The string to check."},
				{Name: "suffix", Type: ArgTypeString, Required: true, Description: "The suffix to check for."},
			},
			ReturnType:      ArgTypeBool,
			ReturnHelp:      "Returns true if the input_string ends with the suffix, false otherwise.",
			Example:         `TOOL.HasSuffix(input_string: "document.doc", suffix: ".doc") // Returns true`,
			ErrorConditions: "ErrInvalidArgType if arguments are not of the correct type; ErrMissingArg if required arguments are not provided.",
		},
		Func: toolStringHasSuffix,
	},
	{
		Spec: ToolSpec{
			Name:        "LineCount",
			Description: "Counts the number of lines in the given string content.",
			Category:    "String Manipulation",
			Args: []ArgSpec{
				{Name: "content_string", Type: ArgTypeString, Required: true, Description: "The string content in which to count lines."},
			},
			ReturnType:      ArgTypeInt,
			ReturnHelp:      "Returns an integer representing the number of lines (separated by '\\n') in the string. An empty string has 1 line if not ending with \\n, or 0 if it does.",
			Example:         `TOOL.LineCount(content_string: "line1\nline2\nline3") // Returns 3`,
			ErrorConditions: "ErrInvalidArgType if content_string is not a string; ErrMissingArg if content_string is not provided.",
		},
		Func: toolLineCountString,
	},
}
