// NeuroScript Version: 0.3.1
// File version: 0.0.5
// Purpose: Standardized argument validation to use ErrArgumentMismatch instead of ErrInvalidArgument for consistency with test expectations.
// nlines: 280
// risk_rating: LOW
// filename: pkg/core/tools_string.go

package core

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// --- Tool Implementations ---

func toolStringLength(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "String.Length: expected 1 argument (input_string)", ErrArgumentMismatch)
	}
	inputStr, ok := args[0].(string)
	if !ok {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("String.Length: input_string argument must be a string, got %T", args[0]), ErrArgumentMismatch)
	}
	length := float64(utf8.RuneCountInString(inputStr))
	interpreter.Logger().Debug("Tool: String.Length", "input", inputStr, "length", length)
	return length, nil
}

func toolStringSubstring(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Corresponds to "Substring" tool with args: input_string, start_index, length
	if len(args) != 3 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "String.Substring: expected 3 arguments (input_string, start_index, length)", ErrArgumentMismatch)
	}
	inputStr, okStr := args[0].(string)
	startIndexRaw, okStart := args[1].(int64)
	lengthRaw, okLen := args[2].(int64)

	if !okStr {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("String.Substring: input_string argument must be a string, got %T", args[0]), ErrArgumentMismatch)
	}
	if !okStart {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("String.Substring: start_index argument must be an integer, got %T", args[1]), ErrArgumentMismatch)
	}
	if !okLen {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("String.Substring: length argument must be an integer, got %T", args[2]), ErrArgumentMismatch)
	}

	startIndex := int(startIndexRaw)
	length := int(lengthRaw)
	runes := []rune(inputStr)
	runeCount := len(runes)

	// Check for negative indices/length *before* clamping
	if startIndex < 0 {
		return nil, NewRuntimeError(ErrorCodeBounds, fmt.Sprintf("String.Substring: start_index (%d) cannot be negative", startIndex), ErrListIndexOutOfBounds)
	}
	if length < 0 {
		return nil, NewRuntimeError(ErrorCodeBounds, fmt.Sprintf("String.Substring: length (%d) cannot be negative", length), ErrListIndexOutOfBounds)
	}

	// Clamp start index
	if startIndex > runeCount {
		startIndex = runeCount // Clamp to end (allows empty string result)
	}

	// Calculate end index based on clamped start and requested length
	endIndex := startIndex + length

	// Clamp end index
	if endIndex > runeCount {
		endIndex = runeCount
	}

	// Handle cases resulting in empty string
	if startIndex >= endIndex || startIndex >= runeCount {
		interpreter.Logger().Debug("Tool: String.Substring (empty due to indices/length)", "input", inputStr, "start", startIndexRaw, "length", lengthRaw, "rune_count", runeCount, "result", "")
		return "", nil
	}

	substring := string(runes[startIndex:endIndex])
	interpreter.Logger().Debug("Tool: String.Substring", "input", inputStr, "start", startIndexRaw, "length", lengthRaw, "result", substring)
	return substring, nil
}

func toolStringConcat(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Corresponds to "Concat" tool with args: strings_list (ArgTypeSliceString)
	if len(args) != 1 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "String.Concat: expected 1 argument (strings_list)", ErrArgumentMismatch)
	}
	// Validation ensures this is []string
	stringsList, ok := args[0].([]string)
	if !ok {
		// This should not happen if validation worked correctly with ArgTypeSliceString
		return nil, NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("String.Concat: internal error - expected []string from validation, got %T", args[0]), ErrTypeAssertionFailed)
	}

	var builder strings.Builder
	for _, str := range stringsList {
		builder.WriteString(str)
	}

	result := builder.String()
	interpreter.Logger().Debug("Tool: String.Concat", "input_count", len(stringsList), "result_length", len(result))
	return result, nil
}

func toolStringSplit(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Corresponds to "Split" tool with args: input_string, delimiter
	if len(args) != 2 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "String.Split: expected 2 arguments (input_string, delimiter)", ErrArgumentMismatch)
	}
	inputStr, okStr := args[0].(string)
	separator, okSep := args[1].(string)

	if !okStr {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("String.Split: input_string argument must be a string, got %T", args[0]), ErrArgumentMismatch)
	}
	if !okSep {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("String.Split: delimiter argument must be a string, got %T", args[1]), ErrArgumentMismatch)
	}

	// Corrected: Directly return []string
	parts := strings.Split(inputStr, separator)

	interpreter.Logger().Debug("Tool: String.Split", "input_length", len(inputStr), "separator", separator, "parts_count", len(parts))
	return parts, nil // Return []string directly
}

func toolSplitWords(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Corresponds to "SplitWords" tool with args: input_string
	if len(args) != 1 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "String.SplitWords: expected 1 argument (input_string)", ErrArgumentMismatch)
	}
	inputStr, ok := args[0].(string)
	if !ok {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("String.SplitWords: input_string argument must be a string, got %T", args[0]), ErrArgumentMismatch)
	}

	// Corrected: Directly return []string
	parts := strings.Fields(inputStr)

	interpreter.Logger().Debug("Tool: String.SplitWords", "input_length", len(inputStr), "parts_count", len(parts))
	return parts, nil // Return []string directly
}

func toolStringJoin(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Corresponds to "Join" tool with args: string_list, separator
	if len(args) != 2 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "String.Join: expected 2 arguments (string_list, separator)", ErrArgumentMismatch)
	}
	// Corrected: Expect []string directly due to ArgTypeSliceString spec
	stringList, okList := args[0].([]string)
	separator, okSep := args[1].(string)

	if !okList {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("String.Join: string_list argument must be a list of strings, got %T", args[0]), ErrArgumentMismatch)
	}
	if !okSep {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("String.Join: separator argument must be a string, got %T", args[1]), ErrArgumentMismatch)
	}

	// No need to convert elements if input is already []string
	result := strings.Join(stringList, separator)
	interpreter.Logger().Debug("Tool: String.Join", "input_count", len(stringList), "separator", separator, "result_length", len(result))
	return result, nil
}

func toolStringContains(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Corresponds to "Contains" tool with args: input_string, substring
	if len(args) != 2 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "String.Contains: expected 2 arguments (input_string, substring)", ErrArgumentMismatch)
	}
	inputStr, okStr := args[0].(string)
	substr, okSubstr := args[1].(string)

	if !okStr {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("String.Contains: input_string argument must be a string, got %T", args[0]), ErrArgumentMismatch)
	}
	if !okSubstr {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("String.Contains: substring argument must be a string, got %T", args[1]), ErrArgumentMismatch)
	}

	contains := strings.Contains(inputStr, substr)
	interpreter.Logger().Debug("Tool: String.Contains", "input", inputStr, "substring", substr, "result", contains)
	return contains, nil
}

func toolStringHasPrefix(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Corresponds to "HasPrefix" tool with args: input_string, prefix
	if len(args) != 2 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "String.HasPrefix: expected 2 arguments (input_string, prefix)", ErrArgumentMismatch)
	}
	inputStr, okStr := args[0].(string)
	prefix, okPrefix := args[1].(string)

	if !okStr {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("String.HasPrefix: input_string argument must be a string, got %T", args[0]), ErrArgumentMismatch)
	}
	if !okPrefix {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("String.HasPrefix: prefix argument must be a string, got %T", args[1]), ErrArgumentMismatch)
	}

	hasPrefix := strings.HasPrefix(inputStr, prefix)
	interpreter.Logger().Debug("Tool: String.HasPrefix", "input", inputStr, "prefix", prefix, "result", hasPrefix)
	return hasPrefix, nil
}

func toolStringHasSuffix(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Corresponds to "HasSuffix" tool with args: input_string, suffix
	if len(args) != 2 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "String.HasSuffix: expected 2 arguments (input_string, suffix)", ErrArgumentMismatch)
	}
	inputStr, okStr := args[0].(string)
	suffix, okSuffix := args[1].(string)

	if !okStr {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("String.HasSuffix: input_string argument must be a string, got %T", args[0]), ErrArgumentMismatch)
	}
	if !okSuffix {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("String.HasSuffix: suffix argument must be a string, got %T", args[1]), ErrArgumentMismatch)
	}

	hasSuffix := strings.HasSuffix(inputStr, suffix)
	interpreter.Logger().Debug("Tool: String.HasSuffix", "input", inputStr, "suffix", suffix, "result", hasSuffix)
	return hasSuffix, nil
}

func toolStringToUpper(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Corresponds to "ToUpper" tool with args: input_string
	if len(args) != 1 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "String.ToUpper: expected 1 argument (input_string)", ErrArgumentMismatch)
	}
	inputStr, ok := args[0].(string)
	if !ok {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("String.ToUpper: input_string argument must be a string, got %T", args[0]), ErrArgumentMismatch)
	}
	result := strings.ToUpper(inputStr)
	interpreter.Logger().Debug("Tool: String.ToUpper", "input", inputStr, "result", result)
	return result, nil
}

func toolStringToLower(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Corresponds to "ToLower" tool with args: input_string
	if len(args) != 1 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "String.ToLower: expected 1 argument (input_string)", ErrArgumentMismatch)
	}
	inputStr, ok := args[0].(string)
	if !ok {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("String.ToLower: input_string argument must be a string, got %T", args[0]), ErrArgumentMismatch)
	}
	result := strings.ToLower(inputStr)
	interpreter.Logger().Debug("Tool: String.ToLower", "input", inputStr, "result", result)
	return result, nil
}

func toolStringTrimSpace(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Corresponds to "TrimSpace" tool with args: input_string
	if len(args) != 1 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "String.TrimSpace: expected 1 argument (input_string)", ErrArgumentMismatch)
	}
	inputStr, ok := args[0].(string)
	if !ok {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("String.TrimSpace: input_string argument must be a string, got %T", args[0]), ErrArgumentMismatch)
	}
	result := strings.TrimSpace(inputStr)
	interpreter.Logger().Debug("Tool: String.TrimSpace", "input", inputStr, "result", result)
	return result, nil
}

func toolStringReplace(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Corresponds to "Replace" tool with args: input_string, old_substring, new_substring, count
	if len(args) != 4 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "String.Replace: expected 4 arguments (input_string, old_substring, new_substring, count)", ErrArgumentMismatch)
	}
	inputStr, okStr := args[0].(string)
	oldSubstr, okOld := args[1].(string)
	newSubstr, okNew := args[2].(string)
	countRaw, okCount := args[3].(int64)

	if !okStr {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("String.Replace: input_string argument must be a string, got %T", args[0]), ErrArgumentMismatch)
	}
	if !okOld {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("String.Replace: old_substring argument must be a string, got %T", args[1]), ErrArgumentMismatch)
	}
	if !okNew {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("String.Replace: new_substring argument must be a string, got %T", args[2]), ErrArgumentMismatch)
	}
	if !okCount {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("String.Replace: count argument must be an integer, got %T", args[3]), ErrArgumentMismatch)
	}

	count := int(countRaw)
	result := strings.Replace(inputStr, oldSubstr, newSubstr, count)
	interpreter.Logger().Debug("Tool: String.Replace", "input", inputStr, "old", oldSubstr, "new", newSubstr, "count", count, "result", result)
	return result, nil
}

func toolLineCountString(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Corresponds to "LineCount" tool with args: content_string
	if len(args) != 1 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "String.LineCount: expected 1 argument (content_string)", ErrArgumentMismatch)
	}
	content, ok := args[0].(string)
	if !ok {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("String.LineCount: content_string argument must be a string, got %T", args[0]), ErrArgumentMismatch)
	}

	if content == "" {
		interpreter.Logger().Debug("Tool: String.LineCount", "content", content, "line_count", 0)
		return float64(0), nil
	}
	// Count occurrences of newline character
	lineCount := float64(strings.Count(content, "\n"))
	// Add 1 if the string doesn't end with a newline (to count the last line)
	if !strings.HasSuffix(content, "\n") {
		lineCount++
	}

	interpreter.Logger().Debug("Tool: String.LineCount", "content_len", len(content), "line_count", lineCount)
	return lineCount, nil
}
