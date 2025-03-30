package core

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// --- String Manipulation Tools (Matching ToolFunc Signature) ---
// func(interpreter *Interpreter, args []interface{}) (interface{}, error)

// Add this map definition to pkg/core/tools_string.go

// Helper map for tools needing special arg handling (e.g., non-string types)
// The interpreter checks this map. Functions listed here receive []interface{}.
var specialStringToolImplementations = map[string]bool{
	"JoinStrings": true,
	// Add other tool names here if they need raw interface{} args
	// instead of relying on ValidateAndConvertArgs for basic types.
}

func toolStringLength(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("TOOL.StringLength expects 1 argument (input_string)")
	}
	inputStr, ok := args[0].(string)
	if !ok {
		// Attempt conversion if not string
		inputStr = fmt.Sprintf("%v", args[0])
		// return nil, fmt.Errorf("TOOL.StringLength argument must be a string (got %T)", args[0])
	}

	length := utf8.RuneCountInString(inputStr)
	// Return as string for now, consistent with interpreter expectations
	return fmt.Sprintf("%d", length), nil
}

func toolSubstring(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("TOOL.Substring expects 3 arguments (input_string, start_index, end_index)")
	}

	// Arg 0: Input String
	inputStr, ok := args[0].(string)
	if !ok {
		inputStr = fmt.Sprintf("%v", args[0]) // Convert if possible
		// return nil, fmt.Errorf("TOOL.Substring: input_string must be a string (got %T)", args[0])
	}

	// Arg 1: Start Index (expect int after validation/conversion)
	startIndex, ok := args[1].(int)
	if !ok {
		return nil, fmt.Errorf("TOOL.Substring: start_index must be an integer (got %T - expected from validation)", args[1])
	}

	// Arg 2: End Index (expect int after validation/conversion)
	endIndex, ok := args[2].(int)
	if !ok {
		return nil, fmt.Errorf("TOOL.Substring: end_index must be an integer (got %T - expected from validation)", args[2])
	}

	runes := []rune(inputStr)
	strLen := len(runes)

	// Index validation (Go slice semantics: start inclusive, end exclusive)
	if startIndex < 0 {
		startIndex = 0
	}
	if endIndex > strLen {
		endIndex = strLen
	}
	if startIndex >= strLen || startIndex >= endIndex {
		return "", nil // Return empty string for invalid/empty range
	}

	return string(runes[startIndex:endIndex]), nil
}

func toolToUpper(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("TOOL.ToUpper expects 1 argument")
	}
	inputStr, ok := args[0].(string)
	if !ok {
		inputStr = fmt.Sprintf("%v", args[0])
	}
	return strings.ToUpper(inputStr), nil
}

func toolToLower(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("TOOL.ToLower expects 1 argument")
	}
	inputStr, ok := args[0].(string)
	if !ok {
		inputStr = fmt.Sprintf("%v", args[0])
	}
	return strings.ToLower(inputStr), nil
}

func toolTrimSpace(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("TOOL.TrimSpace expects 1 argument")
	}
	inputStr, ok := args[0].(string)
	if !ok {
		inputStr = fmt.Sprintf("%v", args[0])
	}
	return strings.TrimSpace(inputStr), nil
}

// toolSplitString returns a native Go []string slice.
func toolSplitString(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("TOOL.SplitString expects 2 arguments (input, delimiter)")
	}
	inputStr, ok1 := args[0].(string)
	delimiter, ok2 := args[1].(string)
	if !ok1 {
		inputStr = fmt.Sprintf("%v", args[0])
	} // Convert first if not string
	if !ok2 {
		return nil, fmt.Errorf("TOOL.SplitString: delimiter must be a string (got %T)", args[1])
	}

	return strings.Split(inputStr, delimiter), nil
}

// toolSplitWords splits by whitespace and returns a []string slice.
func toolSplitWords(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("TOOL.SplitWords expects 1 argument")
	}
	inputStr, ok := args[0].(string)
	if !ok {
		inputStr = fmt.Sprintf("%v", args[0])
	}
	return strings.Fields(inputStr), nil
}

// toolJoinStrings expects the first argument to be a []string slice
// (passed from interpreter after validation/conversion) and the second arg to be the separator string.
func toolJoinStrings(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("TOOL.JoinStrings expects 2 arguments (string_slice, separator)")
	}

	// Arg 0: Slice (expect []string after validation/conversion)
	stringSlice, ok := args[0].([]string)
	if !ok {
		// This ideally shouldn't happen if ValidateAndConvertArgs worked correctly
		return nil, fmt.Errorf("TOOL.JoinStrings: first argument is not []string (got %T - check validation)", args[0])
	}

	// Arg 1: Separator (expect string after validation/conversion)
	separator, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("TOOL.JoinStrings: second argument is not string (got %T - check validation)", args[1])
	}

	return strings.Join(stringSlice, separator), nil
}

func toolReplaceAll(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("TOOL.ReplaceAll expects 3 arguments (input, old, new)")
	}
	inputStr, ok1 := args[0].(string)
	oldSub, ok2 := args[1].(string)
	newSub, ok3 := args[2].(string)
	if !ok1 {
		inputStr = fmt.Sprintf("%v", args[0])
	}
	if !ok2 {
		return nil, fmt.Errorf("TOOL.ReplaceAll: 'old' must be a string (got %T)", args[1])
	}
	if !ok3 {
		return nil, fmt.Errorf("TOOL.ReplaceAll: 'new' must be a string (got %T)", args[2])
	}

	return strings.ReplaceAll(inputStr, oldSub, newSub), nil
}

func toolContains(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("TOOL.Contains expects 2 arguments (input, substring)")
	}
	inputStr, ok1 := args[0].(string)
	subStr, ok2 := args[1].(string)
	if !ok1 {
		inputStr = fmt.Sprintf("%v", args[0])
	}
	if !ok2 {
		return nil, fmt.Errorf("TOOL.Contains: 'substring' must be a string (got %T)", args[1])
	}

	result := strings.Contains(inputStr, subStr)
	return fmt.Sprintf("%t", result), nil // Return "true" or "false"
}

func toolHasPrefix(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("TOOL.HasPrefix expects 2 arguments (input, prefix)")
	}
	inputStr, ok1 := args[0].(string)
	prefix, ok2 := args[1].(string)
	if !ok1 {
		inputStr = fmt.Sprintf("%v", args[0])
	}
	if !ok2 {
		return nil, fmt.Errorf("TOOL.HasPrefix: 'prefix' must be a string (got %T)", args[1])
	}

	result := strings.HasPrefix(inputStr, prefix)
	return fmt.Sprintf("%t", result), nil
}

func toolHasSuffix(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("TOOL.HasSuffix expects 2 arguments (input, suffix)")
	}
	inputStr, ok1 := args[0].(string)
	suffix, ok2 := args[1].(string)
	if !ok1 {
		inputStr = fmt.Sprintf("%v", args[0])
	}
	if !ok2 {
		return nil, fmt.Errorf("TOOL.HasSuffix: 'suffix' must be a string (got %T)", args[1])
	}

	result := strings.HasSuffix(inputStr, suffix)
	return fmt.Sprintf("%t", result), nil
}
