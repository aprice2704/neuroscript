// NeuroScript Version: 0.5.2
// File version: 1
// Purpose: Implements string predicate tools (Contains, HasPrefix, HasSuffix).
// filename: pkg/tool/strtools/tools_string_predicates.go
// nlines: 66
// risk_rating: LOW

package strtools

import (
	"fmt"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

func toolStringContains(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	// Corresponds to "Contains" tool with args: input_string, substring
	if len(args) != 2 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "String.Contains: expected 2 arguments (input_string, substring)", lang.ErrArgumentMismatch)
	}
	inputStr, okStr := args[0].(string)
	substr, okSubstr := args[1].(string)

	if !okStr {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("String.Contains: input_string argument must be a string, got %T", args[0]), lang.ErrArgumentMismatch)
	}
	if !okSubstr {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("String.Contains: substring argument must be a string, got %T", args[1]), lang.ErrArgumentMismatch)
	}

	contains := strings.Contains(inputStr, substr)
	interpreter.GetLogger().Debug("Tool: String.Contains", "input", inputStr, "substring", substr, "result", contains)
	return contains, nil
}

func toolStringHasPrefix(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	// Corresponds to "HasPrefix" tool with args: input_string, prefix
	if len(args) != 2 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "String.HasPrefix: expected 2 arguments (input_string, prefix)", lang.ErrArgumentMismatch)
	}
	inputStr, okStr := args[0].(string)
	prefix, okPrefix := args[1].(string)

	if !okStr {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("String.HasPrefix: input_string argument must be a string, got %T", args[0]), lang.ErrArgumentMismatch)
	}
	if !okPrefix {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("String.HasPrefix: prefix argument must be a string, got %T", args[1]), lang.ErrArgumentMismatch)
	}

	hasPrefix := strings.HasPrefix(inputStr, prefix)
	interpreter.GetLogger().Debug("Tool: String.HasPrefix", "input", inputStr, "prefix", prefix, "result", hasPrefix)
	return hasPrefix, nil
}

func toolStringHasSuffix(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	// Corresponds to "HasSuffix" tool with args: input_string, suffix
	if len(args) != 2 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "String.HasSuffix: expected 2 arguments (input_string, suffix)", lang.ErrArgumentMismatch)
	}
	inputStr, okStr := args[0].(string)
	suffix, okSuffix := args[1].(string)

	if !okStr {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("String.HasSuffix: input_string argument must be a string, got %T", args[0]), lang.ErrArgumentMismatch)
	}
	if !okSuffix {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("String.HasSuffix: suffix argument must be a string, got %T", args[1]), lang.ErrArgumentMismatch)
	}

	hasSuffix := strings.HasSuffix(inputStr, suffix)
	interpreter.GetLogger().Debug("Tool: String.HasSuffix", "input", inputStr, "suffix", suffix, "result", hasSuffix)
	return hasSuffix, nil
}
