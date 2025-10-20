// NeuroScript Version: 0.5.2
// File version: 1
// Purpose: Provides a shared helper function for coercing arguments to int64.
// filename: pkg/tool/strtools/tools_string_helpers.go
// nlines: 21
// risk_rating: LOW

package strtools

// toInt64 robustly converts an interface{} to int64, handling float64 and nil.
func toInt64(v interface{}) (int64, bool) {
	if v == nil {
		return 0, true // Per request, treat nil as a valid 0
	}
	if i, ok := v.(int64); ok {
		return i, true
	}
	if f, ok := v.(float64); ok {
		return int64(f), true
	}
	return 0, false
}
