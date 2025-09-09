// NeuroScript Version: 0.3.0
// File version: 6
// Purpose: Implements the Select tool. Correctly handles both string and array-form paths.
// filename: pkg/tool/ai/select.go
// nlines: 81
// risk_rating: MEDIUM

package ai

import (
	"errors"
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/json_lite"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// Select implements the tool.ai.Select function.
func Select(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) < 2 || len(args) > 3 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, fmt.Sprintf("Select: expected 2 or 3 arguments, got %d", len(args)), lang.ErrArgumentMismatch)
	}

	value := args[0]
	pathArg := args[1]
	missingOK := false
	if len(args) == 3 {
		var ok bool
		missingOK, ok = args[2].(bool)
		if !ok {
			return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("Select: missing_ok argument must be a boolean, got %T", args[2]), lang.ErrInvalidArgument)
		}
	}

	var path json_lite.Path
	var err error

	// Use a type switch to handle both string paths and array-form paths.
	switch p := pathArg.(type) {
	case string:
		path, err = json_lite.ParsePath(p)
	case []interface{}:
		// This logic is adapted from the test helper in path_arrayform_test.go
		// to build a path from a slice of keys/indices.
		path, err = buildPathFromArray(p)
	default:
		err = lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("Select: path argument must be a string or a list, got %T", pathArg), lang.ErrInvalidArgument)
	}

	if err != nil {
		// Ensure parsing/building errors are wrapped correctly.
		return nil, lang.NewRuntimeError(lang.ErrorCodeSyntax, fmt.Sprintf("Select: failed to process path: %v", err), err)
	}

	result, selErr := json_lite.Select(value, path, nil) // Corrected call
	if selErr != nil {
		if missingOK {
			// If missing is OK, we swallow errors like key not found, but not syntax errors.
			switch {
			case errors.Is(selErr, json_lite.ErrMapKeyNotFound), errors.Is(selErr, json_lite.ErrListIndexOutOfBounds), errors.Is(selErr, json_lite.ErrCollectionIsNil):
				return nil, nil
			}
		}
		// Return the specific selection error (e.g., ErrMapKeyNotFound) wrapped in a RuntimeError.
		return nil, lang.NewRuntimeError(lang.ErrorCodeKeyNotFound, selErr.Error(), selErr)
	}

	return result, nil
}

// buildPathFromArray converts an array-form path from the interpreter (e.g., []any{"items", 1, "id"})
// into a json_lite.Path, which is required for selection.
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
		case float64: // Numbers from NeuroScript will be float64
			p = append(p, json_lite.PathSegment{Index: int(v), IsKey: false})
		case int: // Also handle int for robustness
			p = append(p, json_lite.PathSegment{Index: v, IsKey: false})
		default:
			return nil, fmt.Errorf("%w: path array elements must be string or int, got %T", lang.ErrInvalidArgument, v)
		}
	}
	return p, nil
}
