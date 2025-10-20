// NeuroScript Version: 0.5.2
// File version: 1
// Purpose: Implements string substring and replacement tools.
// filename: pkg/tool/strtools/tools_string_substring.go
// nlines: 98
// risk_rating: LOW

package strtools

import (
	"fmt"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

func toolStringSubstring(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	// Corresponds to "Substring" tool with args: input_string, start_index, length
	if len(args) != 3 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "String.Substring: expected 3 arguments (input_string, start_index, length)", lang.ErrArgumentMismatch)
	}
	inputStr, okStr := args[0].(string)
	// FIX: Use the robust toInt64 helper to coerce numeric types.
	startIndexRaw, okStart := toInt64(args[1])
	lengthRaw, okLen := toInt64(args[2])

	if !okStr {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("String.Substring: input_string argument must be a string, got %T", args[0]), lang.ErrArgumentMismatch)
	}
	if !okStart {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("String.Substring: start_index argument must be an integer, got %T", args[1]), lang.ErrArgumentMismatch)
	}
	if !okLen {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("String.Substring: length argument must be an integer, got %T", args[2]), lang.ErrArgumentMismatch)
	}

	startIndex := int(startIndexRaw)
	length := int(lengthRaw)
	runes := []rune(inputStr)
	runeCount := len(runes)

	// Check for negative indices/length *before* clamping
	if startIndex < 0 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeBounds, fmt.Sprintf("String.Substring: start_index (%d) cannot be negative", startIndex), lang.ErrListIndexOutOfBounds)
	}
	if length < 0 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeBounds, fmt.Sprintf("String.Substring: length (%d) cannot be negative", length), lang.ErrListIndexOutOfBounds)
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
		interpreter.GetLogger().Debug("Tool: String.Substring (empty due to indices/length)", "input", inputStr, "start", startIndexRaw, "length", lengthRaw, "rune_count", runeCount, "result", "")
		return "", nil
	}

	substring := string(runes[startIndex:endIndex])
	interpreter.GetLogger().Debug("Tool: String.Substring", "input", inputStr, "start", startIndexRaw, "length", lengthRaw, "result", substring)
	return substring, nil
}

func toolStringReplace(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	// Corresponds to "Replace" tool with args: input_string, old_substring, new_substring, count
	if len(args) != 4 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "String.Replace: expected 4 arguments (input_string, old_substring, new_substring, count)", lang.ErrArgumentMismatch)
	}
	inputStr, okStr := args[0].(string)
	oldSubstr, okOld := args[1].(string)
	newSubstr, okNew := args[2].(string)
	// FIX: Use the robust toInt64 helper to coerce numeric types.
	countRaw, okCount := toInt64(args[3])

	if !okStr {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("String.Replace: input_string argument must be a string, got %T", args[0]), lang.ErrArgumentMismatch)
	}
	if !okOld {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("String.Replace: old_substring argument must be a string, got %T", args[1]), lang.ErrArgumentMismatch)
	}
	if !okNew {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("String.Replace: new_substring argument must be a string, got %T", args[2]), lang.ErrArgumentMismatch)
	}
	if !okCount {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("String.Replace: count argument must be an integer, got %T", args[3]), lang.ErrArgumentMismatch)
	}

	count := int(countRaw)
	result := strings.Replace(inputStr, oldSubstr, newSubstr, count)
	interpreter.GetLogger().Debug("Tool: String.Replace", "input", inputStr, "old", oldSubstr, "new", newSubstr, "count", count, "result", result)
	return result, nil
}
