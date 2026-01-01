// :: product: NS
// :: majorVersion: 1
// :: fileVersion: 4
// :: description: Refactored coercion to include NodeID, EntityID, and Handle validation.
// :: latestChange: Added validation logic for ArgTypeNodeID, ArgTypeEntityID, and ArgTypeHandle.
// :: filename: pkg/tool/tools_coerce.go
// :: serialization: go

package tool

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/utils"
)

// coerceArg attempts to convert an unwrapped Go value `x` into the Go type
// corresponding to the NeuroScript ArgType `t`.
func coerceArg(x interface{}, t ArgType) (interface{}, error) {
	if x == nil {
		return nil, nil // Allow nil through; tools must handle optional args.
	}

	var (
		coerced interface{}
		ok      bool
		err     error
	)

	switch t {
	case ArgTypeString:
		coerced, ok = x.(string)
		if !ok {
			err = fmt.Errorf("expected string, got %T", x)
		}

	case ArgTypeInt:
		coerced, ok = lang.ToInt64(x)
		if !ok {
			err = fmt.Errorf("expected integer or integer-like value, got %T", x)
		}

	case ArgTypeFloat:
		coerced, ok = lang.ToFloat64(x)
		if !ok {
			err = fmt.Errorf("expected float or number-like value, got %T", x)
		}

	case ArgTypeBool:
		coerced, ok = utils.ConvertToBool(x)
		if !ok {
			err = fmt.Errorf("expected boolean or boolean-like value, got %T", x)
		}

	case ArgTypeMap:
		coerced, ok = x.(map[string]interface{})
		if !ok {
			err = fmt.Errorf("expected map[string]any, got %T", x)
		}

	case ArgTypeNil:
		return nil, nil // Type spec explicitly wants nil

	case ArgTypeHandle:
		str, ok := x.(string)
		if !ok {
			return nil, fmt.Errorf("expected string for Handle, got %T", x)
		}
		if !interfaces.IsNSHandle(str) {
			return nil, fmt.Errorf("invalid handle format: %s", str)
		}
		return str, nil

	case ArgTypeNodeID:
		str, ok := x.(string)
		if !ok {
			return nil, fmt.Errorf("expected string for NodeID, got %T", x)
		}
		if !interfaces.IsNodeID(str) {
			return nil, fmt.Errorf("invalid NodeID: must start with 'N_': %s", str)
		}
		return str, nil

	case ArgTypeEntityID:
		str, ok := x.(string)
		if !ok {
			return nil, fmt.Errorf("expected string for EntityID, got %T", x)
		}
		if !interfaces.IsEntityID(str) {
			return nil, fmt.Errorf("invalid EntityID: must start with 'E_': %s", str)
		}
		return str, nil

	case ArgTypeSlice, ArgTypeSliceAny: // ArgTypeSlice is an alias for SliceAny
		coerced, ok, err = utils.ConvertToSliceOfAny(x)
		// Note: ok/err pattern is different for this helper

	case ArgTypeSliceString:
		coerced, ok, err = utils.ConvertToSliceOfString(x)

	case ArgTypeSliceInt:
		coerced, ok, err = utils.ConvertToSliceOfInt64(x)

	case ArgTypeSliceFloat:
		coerced, ok, err = utils.ConvertToSliceOfFloat64(x)

	case ArgTypeSliceBool:
		coerced, ok, err = utils.ConvertToSliceOfBool(x)

	case ArgTypeSliceMap:
		coerced, ok, err = utils.ConvertToSliceOfMap(x)

	case ArgTypeAny:
		return x, nil // No coercion needed

	default:
		return nil, fmt.Errorf("unknown argument type specified for coercion: %s", t)
	}

	// Handle errors from the switch cases
	if err != nil {
		return nil, err // Return the specific conversion error
	}
	if !ok {
		// This should be caught by err checks, but as a fallback
		return nil, fmt.Errorf("coercion to %s failed for type %T", t, x)
	}

	return coerced, nil
}
