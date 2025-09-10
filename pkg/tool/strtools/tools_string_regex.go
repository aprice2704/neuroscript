// NeuroScript Version: 0.5.2
// File version: 1
// Purpose: Implements regular expression tools for string manipulation.
// filename: pkg/tool/strtools/tools_string_regex.go
// nlines: 66
// risk_rating: HIGH

package strtools

import (
	"fmt"
	"regexp"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

func toolMatchRegex(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "MatchRegex: expected 2 arguments (pattern, input_string)", lang.ErrArgumentMismatch)
	}
	pattern, ok1 := args[0].(string)
	input, ok2 := args[1].(string)
	if !ok1 || !ok2 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, "MatchRegex: arguments must be strings", lang.ErrArgumentMismatch)
	}

	matched, err := regexp.MatchString(pattern, input)
	if err != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInvalidValue, fmt.Sprintf("MatchRegex: invalid regex pattern: %v", err), lang.ErrInvalidArgument)
	}
	interpreter.GetLogger().Debug("Tool: MatchRegex", "pattern", pattern, "result", matched)
	return matched, nil
}

func toolFindAllRegex(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "FindAllRegex: expected 2 arguments (pattern, input_string)", lang.ErrArgumentMismatch)
	}
	pattern, ok1 := args[0].(string)
	input, ok2 := args[1].(string)
	if !ok1 || !ok2 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, "FindAllRegex: arguments must be strings", lang.ErrArgumentMismatch)
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInvalidValue, fmt.Sprintf("FindAllRegex: invalid regex pattern: %v", err), lang.ErrInvalidArgument)
	}
	matches := re.FindAllString(input, -1)
	interpreter.GetLogger().Debug("Tool: FindAllRegex", "pattern", pattern, "matches_count", len(matches))
	return matches, nil
}

func toolReplaceRegex(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) != 3 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "ReplaceRegex: expected 3 arguments (pattern, input_string, replacement)", lang.ErrArgumentMismatch)
	}
	pattern, ok1 := args[0].(string)
	input, ok2 := args[1].(string)
	replacement, ok3 := args[2].(string)
	if !ok1 || !ok2 || !ok3 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, "ReplaceRegex: arguments must be strings", lang.ErrArgumentMismatch)
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInvalidValue, fmt.Sprintf("ReplaceRegex: invalid regex pattern: %v", err), lang.ErrInvalidArgument)
	}
	result := re.ReplaceAllString(input, replacement)
	interpreter.GetLogger().Debug("Tool: ReplaceRegex", "pattern", pattern)
	return result, nil
}
