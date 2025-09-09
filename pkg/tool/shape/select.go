// NeuroScript Version: 0.7.0
// File version: 1
// Purpose: Implements the tool.shape.Select function.
// filename: pkg/tool/shape/select.go
// nlines: 100
// risk_rating: MEDIUM

package shape

import (
	"errors"
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/json_lite"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// toolShapeSelect implements the tool.shape.Select function.
func toolShapeSelect(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) < 2 || len(args) > 3 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, fmt.Sprintf("Select: expected 2 or 3 arguments, got %d", len(args)), lang.ErrArgumentMismatch)
	}

	value := args[0]
	pathArg := args[1]

	options := &json_lite.SelectOptions{}
	missingOK := false
	if len(args) == 3 {
		optionsMap, ok := args[2].(map[string]interface{})
		if !ok {
			return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("Select: options argument must be a map, got %T", args[2]), lang.ErrInvalidArgument)
		}
		if v, present := optionsMap["case_insensitive"]; present {
			if b, ok := v.(bool); ok {
				options.CaseInsensitive = b
			}
		}
		if v, present := optionsMap["missing_ok"]; present {
			if b, ok := v.(bool); ok {
				missingOK = b
			}
		}
	}

	var path json_lite.Path
	var err error

	switch p := pathArg.(type) {
	case string:
		path, err = json_lite.ParsePath(p)
	case []interface{}:
		path, err = buildPathFromArray(p)
	default:
		err = lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("Select: path argument must be a string or a list, got %T", pathArg), lang.ErrInvalidArgument)
	}

	if err != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeSyntax, fmt.Sprintf("Select: failed to process path: %v", err), err)
	}

	result, selErr := json_lite.Select(value, path, options)
	if selErr != nil {
		if missingOK {
			switch {
			case errors.Is(selErr, json_lite.ErrMapKeyNotFound), errors.Is(selErr, json_lite.ErrListIndexOutOfBounds), errors.Is(selErr, json_lite.ErrCollectionIsNil):
				return nil, nil // Suppress error and return nil as requested.
			}
		}
		// For other errors, or if missing_ok is false, return the error.
		return nil, lang.NewRuntimeError(lang.ErrorCodeKeyNotFound, selErr.Error(), selErr)
	}

	return result, nil
}

// buildPathFromArray is a helper to construct a json_lite.Path from a NeuroScript list.
func buildPathFromArray(arr []interface{}) (json_lite.Path, error) {
	if len(arr) == 0 {
		return nil, fmt.Errorf("%w: path array cannot be empty", lang.ErrInvalidPath)
	}
	p := make(json_lite.Path, 0, len(arr))
	for _, el := range arr {
		switch v := el.(type) {
		case string:
			if v == "" {
				return nil, fmt.Errorf("%w: path segment cannot be an empty string", lang.ErrInvalidArgument)
			}
			p = append(p, json_lite.PathSegment{Key: v, IsKey: true})
		case float64:
			p = append(p, json_lite.PathSegment{Index: int(v), IsKey: false})
		case int:
			p = append(p, json_lite.PathSegment{Index: v, IsKey: false})
		default:
			return nil, fmt.Errorf("%w: path array elements must be string or number, got %T", lang.ErrInvalidArgument, v)
		}
	}
	return p, nil
}
