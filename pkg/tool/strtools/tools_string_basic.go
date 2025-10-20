// NeuroScript Version: 0.5.2
// File version: 1
// Purpose: Implements basic string tools (Length, ToUpper, ToLower, TrimSpace).
// filename: pkg/tool/strtools/tools_string_basic.go
// nlines: 83
// risk_rating: LOW

package strtools

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

func toolStringLength(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "String.Length: expected 1 argument (input_string)", lang.ErrArgumentMismatch)
	}
	inputStr, ok := args[0].(string)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("String.Length: input_string argument must be a string, got %T", args[0]), lang.ErrArgumentMismatch)
	}
	length := float64(utf8.RuneCountInString(inputStr))
	interpreter.GetLogger().Debug("Tool: String.Length", "input", inputStr, "length", length)
	return length, nil
}

func toolStringToUpper(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	// Corresponds to "ToUpper" tool with args: input_string
	if len(args) != 1 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "String.ToUpper: expected 1 argument (input_string)", lang.ErrArgumentMismatch)
	}
	inputStr, ok := args[0].(string)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("String.ToUpper: input_string argument must be a string, got %T", args[0]), lang.ErrArgumentMismatch)
	}
	result := strings.ToUpper(inputStr)
	interpreter.GetLogger().Debug("Tool: String.ToUpper", "input", inputStr, "result", result)
	return result, nil
}

func toolStringToLower(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	// Corresponds to "ToLower" tool with args: input_string
	if len(args) != 1 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "String.ToLower: expected 1 argument (input_string)", lang.ErrArgumentMismatch)
	}
	inputStr, ok := args[0].(string)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("String.ToLower: input_string argument must be a string, got %T", args[0]), lang.ErrArgumentMismatch)
	}
	result := strings.ToLower(inputStr)
	interpreter.GetLogger().Debug("Tool: String.ToLower", "input", inputStr, "result", result)
	return result, nil
}

func toolStringTrimSpace(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	// Corresponds to "TrimSpace" tool with args: input_string
	if len(args) != 1 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "String.TrimSpace: expected 1 argument (input_string)", lang.ErrArgumentMismatch)
	}
	inputStr, ok := args[0].(string)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("String.TrimSpace: input_string argument must be a string, got %T", args[0]), lang.ErrArgumentMismatch)
	}
	result := strings.TrimSpace(inputStr)
	interpreter.GetLogger().Debug("Tool: String.TrimSpace", "input", inputStr, "result", result)
	return result, nil
}
