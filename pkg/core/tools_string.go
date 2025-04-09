// filename: pkg/core/tools_string.go
package core

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// registerStringTools adds string manipulation tools to the registry.
func registerStringTools(registry *ToolRegistry) {
	registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{Name: "StringLength", Description: "Returns the number of UTF-8 characters (runes) in a string.", Args: []ArgSpec{{Name: "input", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeInt},
		Func: toolStringLength,
	})
	registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{Name: "Substring", Description: "Returns a portion of the string (rune-based indexing).", Args: []ArgSpec{{Name: "input", Type: ArgTypeString, Required: true}, {Name: "start", Type: ArgTypeInt, Required: true, Description: "0-based start index (inclusive)."}, {Name: "end", Type: ArgTypeInt, Required: true, Description: "0-based end index (exclusive)."}}, ReturnType: ArgTypeString},
		Func: toolSubstring,
	})
	registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{Name: "ToUpper", Description: "Converts a string to uppercase.", Args: []ArgSpec{{Name: "input", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeString},
		Func: toolToUpper,
	})
	registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{Name: "ToLower", Description: "Converts a string to lowercase.", Args: []ArgSpec{{Name: "input", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeString},
		Func: toolToLower,
	})
	registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{Name: "TrimSpace", Description: "Removes leading and trailing whitespace from a string.", Args: []ArgSpec{{Name: "input", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeString},
		Func: toolTrimSpace,
	})
	registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{Name: "SplitString", Description: "Splits a string by a delimiter.", Args: []ArgSpec{{Name: "input", Type: ArgTypeString, Required: true}, {Name: "delimiter", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeSliceString},
		Func: toolSplitString,
	})
	registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{Name: "SplitWords", Description: "Splits a string into words based on whitespace.", Args: []ArgSpec{{Name: "input", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeSliceString},
		Func: toolSplitWords,
	})
	registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			Name:        "JoinStrings",
			Description: "Joins elements of a list (converting each to string) with a separator.",
			Args:        []ArgSpec{{Name: "input_slice", Type: ArgTypeSliceAny, Required: true, Description: "List of items to join."}, {Name: "separator", Type: ArgTypeString, Required: true, Description: "String to place between elements."}},
			ReturnType:  ArgTypeString,
		},
		Func: toolJoinStrings,
	})
	registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{Name: "ReplaceAll", Description: "Replaces all occurrences of a substring with another.", Args: []ArgSpec{{Name: "input", Type: ArgTypeString, Required: true}, {Name: "old", Type: ArgTypeString, Required: true}, {Name: "new", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeString},
		Func: toolReplaceAll,
	})
	registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{Name: "Contains", Description: "Checks if a string contains a substring.", Args: []ArgSpec{{Name: "input", Type: ArgTypeString, Required: true}, {Name: "substring", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeBool},
		Func: toolContains,
	})
	registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{Name: "HasPrefix", Description: "Checks if a string starts with a prefix.", Args: []ArgSpec{{Name: "input", Type: ArgTypeString, Required: true}, {Name: "prefix", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeBool},
		Func: toolHasPrefix,
	})
	registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{Name: "HasSuffix", Description: "Checks if a string ends with a suffix.", Args: []ArgSpec{{Name: "input", Type: ArgTypeString, Required: true}, {Name: "suffix", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeBool},
		Func: toolHasSuffix,
	})

	// *** ADDED LineCountString registration ***
	registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			Name:        "LineCountString", // New tool name
			Description: "Counts the number of lines in the given string content.",
			Args: []ArgSpec{
				{Name: "content", Type: ArgTypeString, Required: true, Description: "The string content in which to count lines."},
			},
			ReturnType: ArgTypeInt,
		},
		Func: toolLineCountString, // New implementation function
	})
	// *** END ADDITION ***
}

// --- Implementations ---

func toolStringLength(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	inputStr := args[0].(string)
	length := utf8.RuneCountInString(inputStr)
	return int64(length), nil
}

func toolSubstring(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	inputStr := args[0].(string)
	startIndex := args[1].(int64)
	endIndex := args[2].(int64)
	runes := []rune(inputStr)
	strLen := len(runes)
	start := int(startIndex)
	if start < 0 {
		start = 0
	}
	end := int(endIndex)
	if end > strLen {
		end = strLen
	}
	if start >= strLen || start >= end {
		return "", nil
	}
	return string(runes[start:end]), nil
}

func toolToUpper(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	return strings.ToUpper(args[0].(string)), nil
}
func toolToLower(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	return strings.ToLower(args[0].(string)), nil
}
func toolTrimSpace(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	return strings.TrimSpace(args[0].(string)), nil
}
func toolSplitString(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	return strings.Split(args[0].(string), args[1].(string)), nil
}
func toolSplitWords(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	return strings.Fields(args[0].(string)), nil
}

func toolJoinStrings(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	inputSliceRaw := args[0]
	separator := args[1].(string)
	var stringSlice []string
	switch v := inputSliceRaw.(type) {
	case []string:
		stringSlice = v
	case []interface{}:
		stringSlice = make([]string, len(v))
		for i, item := range v {
			stringSlice[i] = fmt.Sprintf("%v", item)
		} // Convert each element
	default:
		return nil, fmt.Errorf("internal error: JoinStrings received unexpected slice type %T after validation", inputSliceRaw)
	}
	return strings.Join(stringSlice, separator), nil
}

func toolReplaceAll(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	return strings.ReplaceAll(args[0].(string), args[1].(string), args[2].(string)), nil
}
func toolContains(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	return strings.Contains(args[0].(string), args[1].(string)), nil
}
func toolHasPrefix(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	return strings.HasPrefix(args[0].(string), args[1].(string)), nil
}
func toolHasSuffix(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	return strings.HasSuffix(args[0].(string), args[1].(string)), nil
}

// *** ADDED toolLineCountString IMPLEMENTATION ***
func toolLineCountString(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Validation ensures args[0] is a string
	content := args[0].(string)

	if interpreter.logger != nil {
		logSnippet := content
		if len(logSnippet) > 80 {
			logSnippet = logSnippet[:80] + "..."
		}
		interpreter.logger.Printf("[TOOL LineCountString] Counting lines in string (snippet): %q", logSnippet)
	}

	if len(content) == 0 {
		return int64(0), nil
	}
	lineCount := int64(strings.Count(content, "\n"))
	if len(content) > 0 && !strings.HasSuffix(content, "\n") {
		lineCount++
	}
	if content == "\n" {
		lineCount = 1
	} // Handle single newline case

	if interpreter.logger != nil {
		interpreter.logger.Printf("[TOOL LineCountString] Counted %d lines.", lineCount)
	}
	return lineCount, nil
}

// *** END ADDITION ***
