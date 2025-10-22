// NeuroScript Version: 0.8.0
// File version: 9
// Purpose: Fixes Unwrap to correctly handle *NilValue pointers.
// filename: pkg/lang/value_helpers.go
// nlines: 194
// risk_rating: HIGH

package lang

import (
	"fmt"
	"os"
	"reflect"
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
		// Ensure MapValue pointers from incorrect wrapping are dereferenced
		if mvPtr, ok := v.(*MapValue); ok && mvPtr != nil {
			return *mvPtr, nil // Return the value, not the pointer
		}
		// FIX: Ensure NilValue pointers are handled correctly if passed in
		if _, ok := v.(*NilValue); ok {
			// Even if a pointer is passed, return the value type consistently
			return NilValue{}, nil
		}
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
		// THE FIX: Return MapValue by value, not by pointer.
		return MapValue{Value: newMap}, nil

	default:
		// Use reflection to handle slices of any string-based type.
		val := reflect.ValueOf(x)
		if val.Kind() == reflect.Slice {
			elemType := val.Type().Elem()
			if elemType.Kind() == reflect.String {
				elems := make([]Value, val.Len())
				for i := 0; i < val.Len(); i++ {
					elems[i] = StringValue{Value: val.Index(i).String()}
				}
				return ListValue{Value: elems}, nil
			}
		}

		// --- REMOVED DANGEROUS STRUCT-TO-JSON FALLBACK ---
		// if val.Kind() == reflect.Struct { ... }

		// FIX: Add explicit logging and error on failure
		errMsg := fmt.Sprintf("core.Wrap: unsupported type %T. Tool authors must return map[string]any, not structs.", v)
		fmt.Fprintf(os.Stderr, "--- CRITICAL: lang.Wrap failed: %s ---\n", errMsg)
		// FIX: Use "%s" to satisfy the vet linter
		return nil, fmt.Errorf("%s", errMsg)
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
	// --- ADD THIS CASE ---
	case *NilValue: // Handle pointer to NilValue explicitly
		return nil
	// --- END ADD ---
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

	// Handle both MapValue and *MapValue during unwrap for robustness
	case MapValue:
		out := make(map[string]any)
		for k, e := range t.Value {
			out[k] = Unwrap(e)
		}
		return out
	case *MapValue:
		if t == nil { // Handle nil pointer case
			return nil
		}
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
