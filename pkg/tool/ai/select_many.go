// NeuroScript Version: 0.3.0
// File version: 2
// Purpose: Implements the SelectMany tool for the AI toolset.
// filename: pkg/tool/ai/select_many.go
// nlines: 62
// risk_rating: MEDIUM

package ai

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/json_lite"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// SelectMany implements the tool.ai.SelectMany function.
func SelectMany(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) < 2 || len(args) > 3 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, fmt.Sprintf("SelectMany: expected 2 or 3 arguments, got %d", len(args)), lang.ErrArgumentMismatch)
	}

	value := args[0]

	extracts, ok := args[1].(map[string]interface{})
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("SelectMany: extracts argument must be a map, got %T", args[1]), lang.ErrInvalidArgument)
	}

	missingOK := false
	if len(args) == 3 {
		missingOK, ok = args[2].(bool)
		if !ok {
			return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("SelectMany: missing_ok argument must be a boolean, got %T", args[2]), lang.ErrInvalidArgument)
		}
	}

	result := make(map[string]interface{})

	for key, pathArg := range extracts {
		path, err := json_lite.ParsePath(fmt.Sprintf("%v", pathArg))
		if err != nil {
			return nil, lang.NewRuntimeError(lang.ErrorCodeSyntax, fmt.Sprintf("SelectMany: failed to parse path for key '%s': %v", key, err), err)
		}

		extractedValue, err := json_lite.Select(value, path, nil) // Corrected call
		if err != nil {
			if missingOK {
				continue
			}
			return nil, lang.NewRuntimeError(lang.ErrorCodeKeyNotFound, fmt.Sprintf("at key '%s': %v", key, err), err)
		}
		result[key] = extractedValue
	}

	return result, nil
}
