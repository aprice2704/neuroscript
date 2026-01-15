// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 2
// :: description: Implements basic string tools (Length, ToUpper, ToLower, TrimSpace, TrimPrefix, TrimSuffix).
// :: latestChange: Added TrimPrefix and TrimSuffix tools.
// :: filename: pkg/tool/strtools/tools_string_basic.go
// :: serialization: go

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

func toolStringTrimPrefix(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	// Corresponds to "TrimPrefix" tool with args: input_string, prefix
	if len(args) != 2 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "String.TrimPrefix: expected 2 arguments (input_string, prefix)", lang.ErrArgumentMismatch)
	}
	inputStr, okStr := args[0].(string)
	prefix, okPrefix := args[1].(string)

	if !okStr {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("String.TrimPrefix: input_string argument must be a string, got %T", args[0]), lang.ErrArgumentMismatch)
	}
	if !okPrefix {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("String.TrimPrefix: prefix argument must be a string, got %T", args[1]), lang.ErrArgumentMismatch)
	}

	result := strings.TrimPrefix(inputStr, prefix)
	interpreter.GetLogger().Debug("Tool: String.TrimPrefix", "input", inputStr, "prefix", prefix, "result", result)
	return result, nil
}

func toolStringTrimSuffix(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	// Corresponds to "TrimSuffix" tool with args: input_string, suffix
	if len(args) != 2 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "String.TrimSuffix: expected 2 arguments (input_string, suffix)", lang.ErrArgumentMismatch)
	}
	inputStr, okStr := args[0].(string)
	suffix, okSuffix := args[1].(string)

	if !okStr {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("String.TrimSuffix: input_string argument must be a string, got %T", args[0]), lang.ErrArgumentMismatch)
	}
	if !okSuffix {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("String.TrimSuffix: suffix argument must be a string, got %T", args[1]), lang.ErrArgumentMismatch)
	}

	result := strings.TrimSuffix(inputStr, suffix)
	interpreter.GetLogger().Debug("Tool: String.TrimSuffix", "input", inputStr, "suffix", suffix, "result", result)
	return result, nil
}
