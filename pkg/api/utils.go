// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 1
// :: description: Utility functions for NeuroScript API interactions.
// :: latestChange: Created file with DeepSanitize to bridge Go types to Interpreter types.
// :: filename: pkg/api/utils.go
// :: serialization: go

package api

// DeepSanitize recursively converts typed Go maps and slices into
// generic []any and map[string]any structures required by the
// NeuroScript interpreter (lang.Wrap).
//
// The interpreter requires generics to function correctly, but Go libraries
// often return strongly typed collections (e.g., map[string]int, []string).
// This function bridges that gap.
func DeepSanitize(v any) any {
	switch t := v.(type) {
	// --- Generic Containers (Recurse) ---
	case []map[string]any:
		out := make([]any, len(t))
		for i, item := range t {
			out[i] = DeepSanitize(item)
		}
		return out
	case []any:
		out := make([]any, len(t))
		for i, item := range t {
			out[i] = DeepSanitize(item)
		}
		return out
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, val := range t {
			out[k] = DeepSanitize(val)
		}
		return out

	// --- Common Typed Maps (Convert) ---
	case map[string]int:
		out := make(map[string]any, len(t))
		for k, val := range t {
			out[k] = val
		}
		return out
	case map[string]int64:
		out := make(map[string]any, len(t))
		for k, val := range t {
			out[k] = val
		}
		return out
	case map[string]string:
		out := make(map[string]any, len(t))
		for k, val := range t {
			out[k] = val
		}
		return out

	// --- Common Typed Slices (Convert) ---
	case []string:
		out := make([]any, len(t))
		for i, v := range t {
			out[i] = v
		}
		return out
	case []int:
		out := make([]any, len(t))
		for i, v := range t {
			out[i] = v
		}
		return out

	// --- Passthrough ---
	default:
		return v
	}
}
