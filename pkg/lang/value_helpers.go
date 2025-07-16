// NeuroScript Version: 0.4.1
// File version: 1
// Purpose: Implements the core Value wrapping/unwrapping contract.
// filename: pkg/lang/value_helpers.go
// nlines: 151
// risk_rating: MEDIUM

package lang

import (
	"fmt"
	"time"
)

//----------------------------------------------------------------------------
// Public helpers – the ONLY wrapping / unwrapping API external code should use
//----------------------------------------------------------------------------

// Wrap turns ordinary Go values into the tagged-union `Value` types used
// internally by the interpreter.
func Wrap(x any) (Value, error) {
	switch v := x.(type) {
	case nil:
		return NilValue{}, nil

	case Value: // already wrapped
		return v, nil

	case string:
		return StringValue{Value: v}, nil
	case []byte:
		return BytesValue{Value: v}, nil
	case bool:
		return BoolValue{Value: v}, nil
	case int:
		return NumberValue{Value: float64(v)}, nil
	case int64:
		return NumberValue{Value: float64(v)}, nil
	case float64:
		return NumberValue{Value: v}, nil
	case time.Time:
		return TimedateValue{Value: v}, nil

	case []any:
		elems := make([]Value, len(v))
		for i, item := range v {
			wrappedItem, err := Wrap(item)
			if err != nil {
				return nil, fmt.Errorf("error wrapping element %d in slice: %w", i, err)
			}
			elems[i] = wrappedItem
		}
		return ListValue{Value: elems}, nil

	case map[string]any:
		newMap := make(map[string]Value)
		for key, val := range v {
			wrappedVal, err := Wrap(val)
			if err != nil {
				return nil, fmt.Errorf("error wrapping value for key %q in map: %w", key, err)
			}
			newMap[key] = wrappedVal
		}
		return &MapValue{Value: newMap}, nil

	default:
		return nil, fmt.Errorf("core.Wrap: unsupported type %T", v)
	}
}

// Unwrap converts a wrapper back to its underlying primitive Go form.
// (It is intentionally lossy – metadata on wrappers is dropped.)
func Unwrap(v Value) any {
	if v == nil {
		return nil
	}
	switch t := v.(type) {
	case NilValue:
		return nil
	case StringValue:
		return t.Value
	case BytesValue:
		return t.Value
	case BoolValue:
		return t.Value
	case NumberValue:
		return t.Value
	case TimedateValue:
		return t.Value
	case FuzzyValue:
		return t.μ
	case FunctionValue:
		return t.Value // Returns the raw Procedure struct
	case ToolValue:
		return t.Value // Returns the raw ToolImplementation struct

	case ListValue:
		out := make([]any, len(t.Value))
		for i, e := range t.Value {
			out[i] = Unwrap(e)
		}
		return out

	case *MapValue:
		out := make(map[string]any)
		for k, e := range t.Value {
			out[k] = Unwrap(e)
		}
		return out

	case ErrorValue: // Errors are just maps
		out := make(map[string]any)
		for k, e := range t.Value {
			out[k] = Unwrap(e)
		}
		return out

	case EventValue: // Events are just maps
		out := make(map[string]any)
		for k, e := range t.Value {
			out[k] = Unwrap(e)
		}
		return out

	default:
		// This should ideally not be reached if all types are handled.
		// Return the wrapper itself as a fallback.
		return t
	}
}

// UnwrapSlice is a convenience helper for unwrapping a slice of Values.
func UnwrapSlice(in []Value) ([]any, error) {
	if in == nil {
		return nil, nil
	}

	out := make([]any, len(in))
	for i, v := range in {
		// Note: Unwrap currently doesn't return an error. If it did,
		// we would handle it here.
		out[i] = Unwrap(v)
	}
	return out, nil
}
