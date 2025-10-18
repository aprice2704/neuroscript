// NeuroScript Version: 0.8.0
// File version: 1
// Purpose: Contains the coerceArg helper function, split from tools_bridge.go.
// filename: pkg/tool/tools_coerce.go
// nlines: 70+
// risk_rating: MEDIUM

package tool

import (
	"fmt"
	"os"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/utils"
)

// coerceArg attempts to convert an unwrapped Go value `x` into the Go type
// corresponding to the NeuroScript ArgType `t`.
func coerceArg(x interface{}, t ArgType) (interface{}, error) {
	fmt.Fprintf(os.Stderr, "      [coerceArg] ENTERED. Input: (%T)%#v, TargetType: %s\n", x, x, t) // DEBUG
	if x == nil {
		fmt.Fprintf(os.Stderr, "      [coerceArg] Input is nil, returning nil.\n") // DEBUG
		return nil, nil                                                            // Allow nil through; tools must handle optional args.
	}

	switch t {
	case ArgTypeString:
		s, ok := x.(string)
		if !ok {
			// For now, strict type checking. Consider allowing number->string?
			err := fmt.Errorf("expected string, got %T", x)
			fmt.Fprintf(os.Stderr, "      [coerceArg] FAILED String: %v\n", err) // DEBUG
			return nil, err
		}
		fmt.Fprintf(os.Stderr, "      [coerceArg] SUCCESS String: %s\n", s) // DEBUG
		return s, nil
	case ArgTypeInt:
		i, ok := lang.ToInt64(x) // Use lang helper which handles wrapped/unwrapped numbers/strings
		if !ok {
			err := fmt.Errorf("expected integer or integer-like value, got %T", x)
			fmt.Fprintf(os.Stderr, "      [coerceArg] FAILED Int: %v\n", err) // DEBUG
			return nil, err
		}
		fmt.Fprintf(os.Stderr, "      [coerceArg] SUCCESS Int: %d\n", i) // DEBUG
		return i, nil
	case ArgTypeFloat:
		f, ok := lang.ToFloat64(x) // Use lang helper
		if !ok {
			err := fmt.Errorf("expected float or number-like value, got %T", x)
			fmt.Fprintf(os.Stderr, "      [coerceArg] FAILED Float: %v\n", err) // DEBUG
			return nil, err
		}
		fmt.Fprintf(os.Stderr, "      [coerceArg] SUCCESS Float: %f\n", f) // DEBUG
		return f, nil
	case ArgTypeBool:
		b, ok := utils.ConvertToBool(x) // Handles various bool-like values
		if !ok {
			err := fmt.Errorf("expected boolean or boolean-like value, got %T", x)
			fmt.Fprintf(os.Stderr, "      [coerceArg] FAILED Bool: %v\n", err) // DEBUG
			return nil, err
		}
		fmt.Fprintf(os.Stderr, "      [coerceArg] SUCCESS Bool: %t\n", b) // DEBUG
		return b, nil
	case ArgTypeSliceAny:
		s, ok, convErr := utils.ConvertToSliceOfAny(x) // Handles various slice types
		if !ok {
			err := fmt.Errorf("expected list/slice, got %T (conversion err: %v)", x, convErr)
			fmt.Fprintf(os.Stderr, "      [coerceArg] FAILED SliceAny: %v\n", err) // DEBUG
			return nil, err
		}
		fmt.Fprintf(os.Stderr, "      [coerceArg] SUCCESS SliceAny: (%T)%#v\n", s, s) // DEBUG
		return s, nil
	case ArgTypeMap:
		fmt.Fprintf(os.Stderr, "      [coerceArg] Attempting Map assertion...\n") // DEBUG
		m, ok := x.(map[string]interface{})
		if !ok {
			err := fmt.Errorf("expected map[string]any, got %T", x)
			fmt.Fprintf(os.Stderr, "      [coerceArg] FAILED Map assertion: %v\n", err) // DEBUG
			return nil, err
		}
		fmt.Fprintf(os.Stderr, "      [coerceArg] SUCCESS Map: (%T)%#v\n", m, m) // DEBUG
		return m, nil
	case ArgTypeAny:
		fmt.Fprintf(os.Stderr, "      [coerceArg] SUCCESS Any (no coercion).\n") // DEBUG
		return x, nil                                                            // No coercion needed
	default:
		err := fmt.Errorf("unknown argument type specified for coercion: %s", t)
		fmt.Fprintf(os.Stderr, "      [coerceArg] FAILED Unknown type: %v\n", err) // DEBUG
		return nil, err
	}
}
