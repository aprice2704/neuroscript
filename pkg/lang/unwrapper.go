// NeuroScript Version: 0.7.1
// File version: 1
// Purpose: Provides a canonical unwrapper that preserves integer types for shape validation.
// filename: pkg/lang/unwrapper.go
// nlines: 40
// risk_rating: LOW

package lang

import "math"

// UnwrapForShapeValidation converts a lang.Value to an interface{} suitable for
// the shape validator, crucially preserving integer types where possible to avoid
// the common float64 conversion problem.
func UnwrapForShapeValidation(v Value) interface{} {
	if v == nil {
		return nil
	}
	switch tv := v.(type) {
	case NumberValue:
		// If the number is a whole number, return it as an int64.
		if tv.Value == math.Trunc(tv.Value) {
			return int64(tv.Value)
		}
		return tv.Value
	case *MapValue:
		// Recursively unwrap maps to handle nested structures.
		m := make(map[string]interface{}, len(tv.Value))
		for k, val := range tv.Value {
			m[k] = UnwrapForShapeValidation(val)
		}
		return m
	case ListValue:
		// Recursively unwrap lists.
		l := make([]interface{}, len(tv.Value))
		for i, val := range tv.Value {
			l[i] = UnwrapForShapeValidation(val)
		}
		return l
	default:
		// Use the standard unwrap for all other types.
		return Unwrap(v)
	}
}
