// NeuroScript Version: 0.7.0
// File version: 1
// Purpose: Implements the tool.shape.Validate function.
// filename: pkg/tool/shape/validate.go
// nlines: 68
// risk_rating: LOW

package shape

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/json_lite"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// toolShapeValidate implements the tool.shape.Validate function.
func toolShapeValidate(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
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

	options := &json_lite.ValidateOptions{}
	if len(args) == 3 {
		optionsMap, ok := args[2].(map[string]interface{})
		if !ok {
			return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("Validate: options argument must be a map, got %T", args[2]), lang.ErrInvalidArgument)
		}
		if v, present := optionsMap["allow_extra"]; present {
			if b, ok := v.(bool); ok {
				options.AllowExtra = b
			}
		}
		if v, present := optionsMap["case_insensitive"]; present {
			if b, ok := v.(bool); ok {
				options.CaseInsensitive = b
			}
		}
	}

	shape, err := json_lite.ParseShape(shapeMap)
	if err != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeSyntax, fmt.Sprintf("Validate: failed to parse shape: %v", err), err)
	}

	if err := shape.Validate(value, options); err != nil {
		// The error from Validate is already well-formed, just return it.
		return nil, err
	}

	return true, nil
}
