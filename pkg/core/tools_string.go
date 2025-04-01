package core

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// --- String Manipulation Tools (Matching ToolFunc Signature) ---

func toolStringLength(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("TOOL.StringLength internal error: expected 1 arg, got %d", len(args))
	}
	inputStr, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("TOOL.StringLength internal error: expected string arg, got %T", args[0])
	}
	length := utf8.RuneCountInString(inputStr)
	return int64(length), nil
}

func toolSubstring(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("TOOL.Substring internal error: expected 3 args, got %d", len(args))
	}
	inputStr, ok0 := args[0].(string)
	startIndex, ok1 := args[1].(int64) // Expect int64
	endIndex, ok2 := args[2].(int64)   // Expect int64

	if !ok0 {
		return nil, fmt.Errorf("TOOL.Substring internal error: input_string not string (got %T)", args[0])
	}
	// Updated error messages to expect int64
	if !ok1 {
		return nil, fmt.Errorf("TOOL.Substring internal error: start_index must be int64 (got %T)", args[1])
	}
	if !ok2 {
		return nil, fmt.Errorf("TOOL.Substring internal error: end_index must be int64 (got %T)", args[2])
	}

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

func toolToUpper(interpreter *Interpreter, args []interface{}) (interface{}, error) { /* ... as before ... */
	if len(args) != 1 {
		return nil, fmt.Errorf("TOOL.ToUpper internal error: expected 1 arg")
	}
	inputStr, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("TOOL.ToUpper internal error: expected string arg, got %T", args[0])
	}
	return strings.ToUpper(inputStr), nil
}
func toolToLower(interpreter *Interpreter, args []interface{}) (interface{}, error) { /* ... as before ... */
	if len(args) != 1 {
		return nil, fmt.Errorf("TOOL.ToLower internal error: expected 1 arg")
	}
	inputStr, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("TOOL.ToLower internal error: expected string arg, got %T", args[0])
	}
	return strings.ToLower(inputStr), nil
}
func toolTrimSpace(interpreter *Interpreter, args []interface{}) (interface{}, error) { /* ... as before ... */
	if len(args) != 1 {
		return nil, fmt.Errorf("TOOL.TrimSpace internal error: expected 1 arg")
	}
	inputStr, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("TOOL.TrimSpace internal error: expected string arg, got %T", args[0])
	}
	return strings.TrimSpace(inputStr), nil
}
func toolSplitString(interpreter *Interpreter, args []interface{}) (interface{}, error) { /* ... as before ... */
	if len(args) != 2 {
		return nil, fmt.Errorf("TOOL.SplitString internal error: expected 2 args")
	}
	inputStr, ok1 := args[0].(string)
	delimiter, ok2 := args[1].(string)
	if !ok1 {
		return nil, fmt.Errorf("TOOL.SplitString internal error: expected string arg 0, got %T", args[0])
	}
	if !ok2 {
		return nil, fmt.Errorf("TOOL.SplitString internal error: expected string arg 1, got %T", args[1])
	}
	return strings.Split(inputStr, delimiter), nil
}
func toolSplitWords(interpreter *Interpreter, args []interface{}) (interface{}, error) { /* ... as before ... */
	if len(args) != 1 {
		return nil, fmt.Errorf("TOOL.SplitWords internal error: expected 1 arg")
	}
	inputStr, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("TOOL.SplitWords internal error: expected string arg, got %T", args[0])
	}
	return strings.Fields(inputStr), nil
}
func toolJoinStrings(interpreter *Interpreter, args []interface{}) (interface{}, error) { /* ... as before ... */
	if len(args) != 2 {
		return nil, fmt.Errorf("TOOL.JoinStrings internal error: expected 2 args")
	}
	stringSlice, ok1 := args[0].([]string)
	separator, ok2 := args[1].(string)
	if !ok1 {
		return nil, fmt.Errorf("TOOL.JoinStrings internal error: expected []string arg 0, got %T", args[0])
	}
	if !ok2 {
		return nil, fmt.Errorf("TOOL.JoinStrings internal error: expected string arg 1, got %T", args[1])
	}
	return strings.Join(stringSlice, separator), nil
}
func toolReplaceAll(interpreter *Interpreter, args []interface{}) (interface{}, error) { /* ... as before ... */
	if len(args) != 3 {
		return nil, fmt.Errorf("TOOL.ReplaceAll internal error: expected 3 args")
	}
	inputStr, ok1 := args[0].(string)
	oldSub, ok2 := args[1].(string)
	newSub, ok3 := args[2].(string)
	if !ok1 {
		return nil, fmt.Errorf("TOOL.ReplaceAll internal error: expected string arg 0, got %T", args[0])
	}
	if !ok2 {
		return nil, fmt.Errorf("TOOL.ReplaceAll internal error: expected string arg 1, got %T", args[1])
	}
	if !ok3 {
		return nil, fmt.Errorf("TOOL.ReplaceAll internal error: expected string arg 2, got %T", args[2])
	}
	return strings.ReplaceAll(inputStr, oldSub, newSub), nil
}
func toolContains(interpreter *Interpreter, args []interface{}) (interface{}, error) { /* ... as before ... */
	if len(args) != 2 {
		return nil, fmt.Errorf("TOOL.Contains internal error: expected 2 args")
	}
	inputStr, ok1 := args[0].(string)
	subStr, ok2 := args[1].(string)
	if !ok1 {
		return nil, fmt.Errorf("TOOL.Contains internal error: expected string arg 0, got %T", args[0])
	}
	if !ok2 {
		return nil, fmt.Errorf("TOOL.Contains internal error: expected string arg 1, got %T", args[1])
	}
	return strings.Contains(inputStr, subStr), nil
}
func toolHasPrefix(interpreter *Interpreter, args []interface{}) (interface{}, error) { /* ... as before ... */
	if len(args) != 2 {
		return nil, fmt.Errorf("TOOL.HasPrefix internal error: expected 2 args")
	}
	inputStr, ok1 := args[0].(string)
	prefix, ok2 := args[1].(string)
	if !ok1 {
		return nil, fmt.Errorf("TOOL.HasPrefix internal error: expected string arg 0, got %T", args[0])
	}
	if !ok2 {
		return nil, fmt.Errorf("TOOL.HasPrefix internal error: expected string arg 1, got %T", args[1])
	}
	return strings.HasPrefix(inputStr, prefix), nil
}
func toolHasSuffix(interpreter *Interpreter, args []interface{}) (interface{}, error) { /* ... as before ... */
	if len(args) != 2 {
		return nil, fmt.Errorf("TOOL.HasSuffix internal error: expected 2 args")
	}
	inputStr, ok1 := args[0].(string)
	suffix, ok2 := args[1].(string)
	if !ok1 {
		return nil, fmt.Errorf("TOOL.HasSuffix internal error: expected string arg 0, got %T", args[0])
	}
	if !ok2 {
		return nil, fmt.Errorf("TOOL.HasSuffix internal error: expected string arg 1, got %T", args[1])
	}
	return strings.HasSuffix(inputStr, suffix), nil
}
