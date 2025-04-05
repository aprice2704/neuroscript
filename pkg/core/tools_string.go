// pkg/core/tools_string.go
package core

import (
	"fmt" // <<< Added fmt import
	"strings"
	"unicode/utf8"
)

// registerStringTools adds string manipulation tools to the registry.
func registerStringTools(registry *ToolRegistry) {
	// ... (other tool registrations remain the same) ...
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

	// *** MODIFIED JoinStrings Spec ***
	registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			Name:        "JoinStrings",
			Description: "Joins elements of a list (converting each to string) with a separator.",
			Args: []ArgSpec{
				// Changed expected type to SliceAny
				{Name: "input_slice", Type: ArgTypeSliceAny, Required: true, Description: "List of items to join."},
				{Name: "separator", Type: ArgTypeString, Required: true, Description: "String to place between elements."},
			},
			ReturnType: ArgTypeString,
		},
		Func: toolJoinStrings, // Use the updated function below
	})
	// *** END MODIFICATION ***

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
}

// --- String Manipulation Tools (Matching ToolFunc Signature) ---

// ... (toolStringLength, toolSubstring, etc. remain the same) ...
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
	end := int(endIndex)
	if start < 0 {
		start = 0
	}
	if end > strLen {
		end = strLen
	}
	if start >= strLen || start >= end {
		return "", nil
	}
	return string(runes[start:end]), nil
}

func toolToUpper(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	inputStr := args[0].(string)
	return strings.ToUpper(inputStr), nil
}

func toolToLower(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	inputStr := args[0].(string)
	return strings.ToLower(inputStr), nil
}

func toolTrimSpace(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	inputStr := args[0].(string)
	return strings.TrimSpace(inputStr), nil
}

func toolSplitString(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	inputStr := args[0].(string)
	delimiter := args[1].(string)
	return strings.Split(inputStr, delimiter), nil
}

func toolSplitWords(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	inputStr := args[0].(string)
	return strings.Fields(inputStr), nil
}

// *** MODIFIED toolJoinStrings Implementation ***
func toolJoinStrings(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Validation ensures args[0] is ArgTypeSliceAny (can be []interface{} or []string)
	inputSliceRaw := args[0]
	separator := args[1].(string) // Arg 1 is guaranteed string by validation

	var stringSlice []string

	// Handle both possible slice types from validation
	switch v := inputSliceRaw.(type) {
	case []string:
		stringSlice = v // Use directly if it's already []string
	case []interface{}:
		stringSlice = make([]string, len(v))
		for i, item := range v {
			if item != nil {
				stringSlice[i] = fmt.Sprintf("%v", item) // Convert each element
			} else {
				stringSlice[i] = "" // Convert nil to empty string
			}
		}
	default:
		// Should not happen if validation is correct, but handle defensively
		return nil, fmt.Errorf("internal error: JoinStrings received unexpected slice type %T after validation", inputSliceRaw)
	}

	return strings.Join(stringSlice, separator), nil
}

// *** END MODIFICATION ***

// ... (toolReplaceAll, toolContains, toolHasPrefix, toolHasSuffix remain the same) ...
func toolReplaceAll(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	inputStr := args[0].(string)
	oldSub := args[1].(string)
	newSub := args[2].(string)
	return strings.ReplaceAll(inputStr, oldSub, newSub), nil
}

func toolContains(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	inputStr := args[0].(string)
	subStr := args[1].(string)
	return strings.Contains(inputStr, subStr), nil
}

func toolHasPrefix(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	inputStr := args[0].(string)
	prefix := args[1].(string)
	return strings.HasPrefix(inputStr, prefix), nil
}

func toolHasSuffix(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	inputStr := args[0].(string)
	suffix := args[1].(string)
	return strings.HasSuffix(inputStr, suffix), nil
}
