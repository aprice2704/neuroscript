// NeuroScript Version: 0.3.1
// File version: 0.0.2 // Added toolSplitWords, toolLineCountString. Aligned error messages and arg names with tooldefs.
// nlines: 280 // Approximate
// risk_rating: LOW
// filename: pkg/core/tools_string.go

package core

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// NOTE: The init() function that previously registered tools here has been removed.
// Registration now happens via the stringToolsToRegister variable in tooldefs_string.go
// being processed by zz_core_tools_registrar.go.

// --- Tool Implementations ---

func toolStringLength(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "String.Length: expected 1 argument (input_string)", ErrArgumentMismatch)
	}
	inputStr, ok := args[0].(string)
	if !ok {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("String.Length: input_string argument must be a string, got %T", args[0]), ErrInvalidArgument)
	}
	length := int64(utf8.RuneCountInString(inputStr))
	interpreter.Logger().Debug("Tool: String.Length", "input", inputStr, "length", length)
	return length, nil
}

func toolStringSubstring(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) != 3 {
		// Corresponds to "StringSubstring" tool with args: input_string, start_index, length
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "String.Substring: expected 3 arguments (input_string, start_index, length)", ErrArgumentMismatch)
	}
	inputStr, okStr := args[0].(string)
	startIndexRaw, okStart := args[1].(int64)
	lengthRaw, okLen := args[2].(int64)

	if !okStr {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("String.Substring: input_string argument must be a string, got %T", args[0]), ErrInvalidArgument)
	}
	if !okStart {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("String.Substring: start_index argument must be an integer, got %T", args[1]), ErrInvalidArgument)
	}
	if !okLen {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("String.Substring: length argument must be an integer, got %T", args[2]), ErrInvalidArgument)
	}

	startIndex := int(startIndexRaw)
	length := int(lengthRaw)
	runes := []rune(inputStr)
	runeCount := len(runes)

	if startIndex < 0 {
		return nil, NewRuntimeError(ErrorCodeBounds, fmt.Sprintf("String.Substring: start_index (%d) cannot be negative", startIndex), ErrListIndexOutOfBounds)
	}
	if length < 0 {
		return nil, NewRuntimeError(ErrorCodeBounds, fmt.Sprintf("String.Substring: length (%d) cannot be negative", length), ErrListIndexOutOfBounds)
	}
	if startIndex > runeCount {
		// Allow startIndex == runeCount for zero-length substring at end
		if startIndex == runeCount && length == 0 {
			interpreter.Logger().Debug("Tool: String.Substring (empty at end)", "input", inputStr, "start", startIndex, "length", length, "result", "")
			return "", nil
		}
		return nil, NewRuntimeError(ErrorCodeBounds, fmt.Sprintf("String.Substring: start_index (%d) is out of bounds for string length %d", startIndex, runeCount), ErrListIndexOutOfBounds)
	}

	endIndex := startIndex + length
	if endIndex > runeCount {
		endIndex = runeCount
	}

	if startIndex >= endIndex {
		interpreter.Logger().Debug("Tool: String.Substring (empty due to indices/length)", "input", inputStr, "start", startIndex, "length", length, "rune_count", runeCount, "result", "")
		return "", nil
	}

	substring := string(runes[startIndex:endIndex])
	interpreter.Logger().Debug("Tool: String.Substring", "input", inputStr, "start", startIndex, "length", length, "result", substring)
	return substring, nil
}

func toolStringConcat(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Corresponds to "StringConcat" tool with args: strings_list
	if len(args) != 1 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "String.Concat: expected 1 argument (strings_list)", ErrArgumentMismatch)
	}
	stringsList, ok := args[0].([]interface{})
	if !ok {
		// This case should ideally be caught by ValidateAndConvertArgs if ArgTypeSliceString is strict.
		// If it can be []interface{} from validation, then this check is fine.
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("String.Concat: strings_list argument must be a list of strings, got %T", args[0]), ErrInvalidArgument)
	}

	var builder strings.Builder
	for i, item := range stringsList {
		str, ok := item.(string)
		if !ok {
			return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("String.Concat: list element at index %d must be a string, got %T", i, item), ErrInvalidArgument)
		}
		builder.WriteString(str)
	}

	result := builder.String()
	interpreter.Logger().Debug("Tool: String.Concat", "input_count", len(stringsList), "result_length", len(result))
	return result, nil
}

func toolStringSplit(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Corresponds to "StringSplit" tool with args: input_string, delimiter
	if len(args) != 2 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "String.Split: expected 2 arguments (input_string, delimiter)", ErrArgumentMismatch)
	}
	inputStr, okStr := args[0].(string)
	separator, okSep := args[1].(string)

	if !okStr {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("String.Split: input_string argument must be a string, got %T", args[0]), ErrInvalidArgument)
	}
	if !okSep {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("String.Split: delimiter argument must be a string, got %T", args[1]), ErrInvalidArgument)
	}

	parts := strings.Split(inputStr, separator)
	result := make([]interface{}, len(parts))
	for i, part := range parts {
		result[i] = part
	}

	interpreter.Logger().Debug("Tool: String.Split", "input_length", len(inputStr), "separator", separator, "parts_count", len(parts))
	return result, nil
}

func toolSplitWords(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Corresponds to "StringSplitWords" tool with args: input_string
	if len(args) != 1 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "String.SplitWords: expected 1 argument (input_string)", ErrArgumentMismatch)
	}
	inputStr, ok := args[0].(string)
	if !ok {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("String.SplitWords: input_string argument must be a string, got %T", args[0]), ErrInvalidArgument)
	}

	parts := strings.Fields(inputStr)
	result := make([]interface{}, len(parts))
	for i, part := range parts {
		result[i] = part
	}
	interpreter.Logger().Debug("Tool: String.SplitWords", "input_length", len(inputStr), "parts_count", len(parts))
	return result, nil
}

func toolStringJoin(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Corresponds to "StringJoin" tool with args: string_list, separator
	if len(args) != 2 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "String.Join: expected 2 arguments (string_list, separator)", ErrArgumentMismatch)
	}
	stringListRaw, okList := args[0].([]interface{})
	separator, okSep := args[1].(string)

	if !okList {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("String.Join: string_list argument must be a list, got %T", args[0]), ErrInvalidArgument)
	}
	if !okSep {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("String.Join: separator argument must be a string, got %T", args[1]), ErrInvalidArgument)
	}

	stringList := make([]string, 0, len(stringListRaw))
	for i, item := range stringListRaw {
		str, ok := item.(string)
		if !ok {
			return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("String.Join: list element at index %d must be a string, got %T", i, item), ErrInvalidArgument)
		}
		stringList = append(stringList, str)
	}

	result := strings.Join(stringList, separator)
	interpreter.Logger().Debug("Tool: String.Join", "input_count", len(stringList), "separator", separator, "result_length", len(result))
	return result, nil
}

func toolStringContains(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Corresponds to "StringContains" tool with args: input_string, substring
	if len(args) != 2 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "String.Contains: expected 2 arguments (input_string, substring)", ErrArgumentMismatch)
	}
	inputStr, okStr := args[0].(string)
	substr, okSubstr := args[1].(string)

	if !okStr {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("String.Contains: input_string argument must be a string, got %T", args[0]), ErrInvalidArgument)
	}
	if !okSubstr {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("String.Contains: substring argument must be a string, got %T", args[1]), ErrInvalidArgument)
	}

	contains := strings.Contains(inputStr, substr)
	interpreter.Logger().Debug("Tool: String.Contains", "input", inputStr, "substring", substr, "result", contains)
	return contains, nil
}

func toolStringHasPrefix(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Corresponds to "StringHasPrefix" tool with args: input_string, prefix
	if len(args) != 2 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "String.HasPrefix: expected 2 arguments (input_string, prefix)", ErrArgumentMismatch)
	}
	inputStr, okStr := args[0].(string)
	prefix, okPrefix := args[1].(string)

	if !okStr {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("String.HasPrefix: input_string argument must be a string, got %T", args[0]), ErrInvalidArgument)
	}
	if !okPrefix {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("String.HasPrefix: prefix argument must be a string, got %T", args[1]), ErrInvalidArgument)
	}

	hasPrefix := strings.HasPrefix(inputStr, prefix)
	interpreter.Logger().Debug("Tool: String.HasPrefix", "input", inputStr, "prefix", prefix, "result", hasPrefix)
	return hasPrefix, nil
}

func toolStringHasSuffix(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Corresponds to "StringHasSuffix" tool with args: input_string, suffix
	if len(args) != 2 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "String.HasSuffix: expected 2 arguments (input_string, suffix)", ErrArgumentMismatch)
	}
	inputStr, okStr := args[0].(string)
	suffix, okSuffix := args[1].(string)

	if !okStr {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("String.HasSuffix: input_string argument must be a string, got %T", args[0]), ErrInvalidArgument)
	}
	if !okSuffix {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("String.HasSuffix: suffix argument must be a string, got %T", args[1]), ErrInvalidArgument)
	}

	hasSuffix := strings.HasSuffix(inputStr, suffix)
	interpreter.Logger().Debug("Tool: String.HasSuffix", "input", inputStr, "suffix", suffix, "result", hasSuffix)
	return hasSuffix, nil
}

func toolStringToUpper(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Corresponds to "StringToUpper" tool with args: input_string
	if len(args) != 1 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "String.ToUpper: expected 1 argument (input_string)", ErrArgumentMismatch)
	}
	inputStr, ok := args[0].(string)
	if !ok {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("String.ToUpper: input_string argument must be a string, got %T", args[0]), ErrInvalidArgument)
	}
	result := strings.ToUpper(inputStr)
	interpreter.Logger().Debug("Tool: String.ToUpper", "input", inputStr, "result", result)
	return result, nil
}

func toolStringToLower(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Corresponds to "StringToLower" tool with args: input_string
	if len(args) != 1 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "String.ToLower: expected 1 argument (input_string)", ErrArgumentMismatch)
	}
	inputStr, ok := args[0].(string)
	if !ok {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("String.ToLower: input_string argument must be a string, got %T", args[0]), ErrInvalidArgument)
	}
	result := strings.ToLower(inputStr)
	interpreter.Logger().Debug("Tool: String.ToLower", "input", inputStr, "result", result)
	return result, nil
}

func toolStringTrimSpace(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Corresponds to "StringTrimSpace" tool with args: input_string
	if len(args) != 1 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "String.TrimSpace: expected 1 argument (input_string)", ErrArgumentMismatch)
	}
	inputStr, ok := args[0].(string)
	if !ok {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("String.TrimSpace: input_string argument must be a string, got %T", args[0]), ErrInvalidArgument)
	}
	result := strings.TrimSpace(inputStr)
	interpreter.Logger().Debug("Tool: String.TrimSpace", "input", inputStr, "result", result)
	return result, nil
}

func toolStringReplace(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Corresponds to "StringReplace" tool with args: input_string, old_substring, new_substring, count
	if len(args) != 4 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "String.Replace: expected 4 arguments (input_string, old_substring, new_substring, count)", ErrArgumentMismatch)
	}
	inputStr, okStr := args[0].(string)
	oldSubstr, okOld := args[1].(string)
	newSubstr, okNew := args[2].(string)
	countRaw, okCount := args[3].(int64)

	if !okStr {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("String.Replace: input_string argument must be a string, got %T", args[0]), ErrInvalidArgument)
	}
	if !okOld {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("String.Replace: old_substring argument must be a string, got %T", args[1]), ErrInvalidArgument)
	}
	if !okNew {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("String.Replace: new_substring argument must be a string, got %T", args[2]), ErrInvalidArgument)
	}
	if !okCount {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("String.Replace: count argument must be an integer, got %T", args[3]), ErrInvalidArgument)
	}

	count := int(countRaw)
	result := strings.Replace(inputStr, oldSubstr, newSubstr, count)
	interpreter.Logger().Debug("Tool: String.Replace", "input", inputStr, "old", oldSubstr, "new", newSubstr, "count", count, "result", result)
	return result, nil
}

func toolLineCountString(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Corresponds to "StringLineCount" tool with args: content_string
	if len(args) != 1 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "String.LineCount: expected 1 argument (content_string)", ErrArgumentMismatch)
	}
	content, ok := args[0].(string)
	if !ok {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("String.LineCount: content_string argument must be a string, got %T", args[0]), ErrInvalidArgument)
	}

	if content == "" {
		interpreter.Logger().Debug("Tool: String.LineCount", "content", content, "line_count", 0)
		return int64(0), nil
	}
	lineCount := int64(strings.Count(content, "\n"))
	if !strings.HasSuffix(content, "\n") {
		lineCount++
	}

	interpreter.Logger().Debug("Tool: String.LineCount", "content", content, "line_count", lineCount)
	return lineCount, nil
}
