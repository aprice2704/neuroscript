// NeuroScript Version: 0.5.2
// File version: 1
// Purpose: Implements string splitting and joining tools (Split, SplitWords, Join, Concat).
// filename: pkg/tool/strtools/tools_string_split_join.go
// nlines: 111
// risk_rating: LOW

package strtools

import (
	"fmt"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

func toolStringConcat(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	// Corresponds to "Concat" tool with args: strings_list (ArgTypeAny)
	if len(args) != 1 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "String.Concat: expected 1 argument (strings_list)", lang.ErrArgumentMismatch)
	}

	var builder strings.Builder

	// Handle both []string (fast path) and []interface{} (robust path)
	switch list := args[0].(type) {
	case []string:
		// FAST PATH: Already []string
		for _, str := range list {
			builder.WriteString(str)
		}
		interpreter.GetLogger().Debug("Tool: String.Concat (fast path)", "input_count", len(list), "result_length", builder.Len())
		return builder.String(), nil

	case []interface{}:
		// ROBUST PATH: []interface{}, must coerce each element
		for i, item := range list {
			s, ok := item.(string)
			if !ok {
				s = fmt.Sprint(item)
				interpreter.GetLogger().Warn("Tool: String.Concat coercing non-string element to string", "index", i, "type", fmt.Sprintf("%T", item))
			}
			builder.WriteString(s)
		}
		interpreter.GetLogger().Debug("Tool: String.Concat (builder path)", "input_count", len(list), "result_length", builder.Len())
		return builder.String(), nil

	default:
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("String.Concat: strings_list argument must be a list, got %T", args[0]), lang.ErrArgumentMismatch)
	}
}

func toolStringSplit(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	// Corresponds to "Split" tool with args: input_string, delimiter
	if len(args) != 2 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "String.Split: expected 2 arguments (input_string, delimiter)", lang.ErrArgumentMismatch)
	}
	inputStr, okStr := args[0].(string)
	separator, okSep := args[1].(string)

	if !okStr {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("String.Split: input_string argument must be a string, got %T", args[0]), lang.ErrArgumentMismatch)
	}
	if !okSep {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("String.Split: delimiter argument must be a string, got %T", args[1]), lang.ErrArgumentMismatch)
	}

	// Corrected: Directly return []string
	parts := strings.Split(inputStr, separator)

	interpreter.GetLogger().Debug("Tool: String.Split", "input_length", len(inputStr), "separator", separator, "parts_count", len(parts))
	return parts, nil // Return []string directly
}

func toolSplitWords(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	// Corresponds to "SplitWords" tool with args: input_string
	if len(args) != 1 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "String.SplitWords: expected 1 argument (input_string)", lang.ErrArgumentMismatch)
	}
	inputStr, ok := args[0].(string)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("String.SplitWords: input_string argument must be a string, got %T", args[0]), lang.ErrArgumentMismatch)
	}

	// Corrected: Directly return []string
	parts := strings.Fields(inputStr)

	interpreter.GetLogger().Debug("Tool: String.SplitWords", "input_length", len(inputStr), "parts_count", len(parts))
	return parts, nil // Return []string directly
}

func toolStringJoin(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	// Corresponds to "Join" tool with args: string_list, separator
	if len(args) != 2 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "String.Join: expected 2 arguments (string_list, separator)", lang.ErrArgumentMismatch)
	}
	separator, okSep := args[1].(string)
	if !okSep {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("String.Join: separator argument must be a string, got %T", args[1]), lang.ErrArgumentMismatch)
	}

	var builder strings.Builder
	var listLen int

	// Handle both []string (fast path) and []interface{} (robust path)
	switch list := args[0].(type) {
	case []string:
		// FAST PATH: Already []string
		listLen = len(list)
		for i, str := range list {
			if i > 0 {
				builder.WriteString(separator)
			}
			builder.WriteString(str)
		}
		interpreter.GetLogger().Debug("Tool: String.Join (fast path)", "input_count", listLen, "separator", separator, "result_length", builder.Len())
		return builder.String(), nil

	case []interface{}:
		// ROBUST PATH: []interface{}, must coerce each element
		listLen = len(list)
		for i, item := range list {
			if i > 0 {
				builder.WriteString(separator)
			}
			s, ok := item.(string)
			if !ok {
				s = fmt.Sprint(item)
				interpreter.GetLogger().Warn("Tool: String.Join coercing non-string element to string", "index", i, "type", fmt.Sprintf("%T", item))
			}
			builder.WriteString(s)
		}
		interpreter.GetLogger().Debug("Tool: String.Join (builder path)", "input_count", listLen, "separator", separator, "result_length", builder.Len())
		return builder.String(), nil

	default:
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("String.Join: string_list argument must be a list, got %T", args[0]), lang.ErrArgumentMismatch)
	}
}
