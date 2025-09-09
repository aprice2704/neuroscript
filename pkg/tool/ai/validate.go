// NeuroScript Version: 0.3.0
// File version: 5
// Purpose: Implements the Validate tool. Correctly passes through validation errors.
// filename: pkg/tool/ai/validate.go
// nlines: 45
// risk_rating: LOW

package ai

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/json_lite"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// Validate implements the tool.ai.Validate function.
func Validate(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) < 2 || len(args) > 3 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, fmt.Sprintf("Validate: expected 2 or 3 arguments, got %d", len(args)), lang.ErrArgumentMismatch)
	}

	value, ok := args[0].(map[string]interface{})
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("Validate: value argument must be a map, got %T", args[0]), lang.ErrInvalidArgument)
	}

	shapeMap, ok := args[1].(map[string]interface{})
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("Validate: shape argument must be a map, got %T", args[1]), lang.ErrInvalidArgument)
	}

	allowExtra := false
	if len(args) == 3 {
		allowExtra, ok = args[2].(bool)
		if !ok {
			return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("Validate: allow_extra argument must be a boolean, got %T", args[2]), lang.ErrInvalidArgument)
		}
	}

	shape, err := json_lite.ParseShape(shapeMap)
	if err != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeSyntax, fmt.Sprintf("Validate: failed to parse shape: %v", err), err)
	}

	// Corrected call to use the new options struct.
	// Note: The 'case-insensitive' option is not exposed at the NeuroScript level yet.
	validateOptions := &json_lite.ValidateOptions{
		AllowExtra:      allowExtra,
		CaseInsensitive: false,
	}
	if err := shape.Validate(value, validateOptions); err != nil {
		return nil, err
	}

	return true, nil
}
