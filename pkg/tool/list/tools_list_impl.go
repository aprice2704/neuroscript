// NeuroScript Version: 0.4.0
// File version: 8
// Purpose: Added toolListAppendInPlace implementation (currently same as toolListAppend).
// filename: pkg/tool/list/tools_list_impl.go
// nlines: 247
// risk_rating: LOW

package list

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// --- List Tool Implementations (Primitive-Aware) ---
// These functions are called by the bridge and receive unwrapped, native Go types.
// They return native Go types, which the bridge will then wrap.

func toolListLength(_ tool.Runtime, args []interface{}) (interface{}, error) {
	list, ok := args[0].([]interface{})
	if !ok {
		// Attempt to handle []string as well, though ArgTypeSliceAny should make this []interface{}
		if strList, okStr := args[0].([]string); okStr {
			return float64(len(strList)), nil
		}
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, fmt.Sprintf("len expects a list, got %T", args[0]), lang.ErrArgumentMismatch)
	}
	return float64(len(list)), nil
}

func toolListAppend(_ tool.Runtime, args []interface{}) (interface{}, error) {
	list, ok := args[0].([]interface{})
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "append expects a list for the first argument", lang.ErrArgumentMismatch)
	}
	element := args[1]
	// Uses Go's built-in append, which handles allocation and copying efficiently.
	// Create a new slice with capacity for one more element
	newList := make([]interface{}, len(list), len(list)+1)
	copy(newList, list)
	// Append the new element
	newList = append(newList, element)
	return newList, nil
}

// toolListAppendInPlace uses Go's append, returning the potentially new slice header.
func toolListAppendInPlace(_ tool.Runtime, args []interface{}) (interface{}, error) {
	list, ok := args[0].([]interface{})
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "AppendInPlace expects a list for the first argument", lang.ErrArgumentMismatch)
	}
	element := args[1]
	// Go's append efficiently reuses underlying array if capacity allows,
	// otherwise allocates a new one and copies. It returns the (potentially new) slice header.
	newList := append(list, element)
	return newList, nil
}

func toolListPrepend(_ tool.Runtime, args []interface{}) (interface{}, error) {
	list, ok := args[0].([]interface{})
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "prepend expects a list for the first argument", lang.ErrArgumentMismatch)
	}
	element := args[1]
	// Create a new slice with capacity for the new element + old list
	newList := make([]interface{}, 1, len(list)+1)
	newList[0] = element
	newList = append(newList, list...)
	return newList, nil
}

func toolListGet(_ tool.Runtime, args []interface{}) (interface{}, error) {
	list, ok := args[0].([]interface{})
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "get expects a list", lang.ErrArgumentMismatch)
	}

	// Use helper to handle numeric types robustly for index
	indexRaw, okIndex := toInt64(args[1])
	if !okIndex {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("get expects an integer for index, got %T", args[1]), lang.ErrListInvalidIndexType)
	}
	index := int(indexRaw)

	var defaultValue interface{} = nil
	if len(args) > 2 {
		defaultValue = args[2]
	}

	if index < 0 || index >= len(list) {
		return defaultValue, nil
	}
	return list[index], nil
}

func toolListSlice(_ tool.Runtime, args []interface{}) (interface{}, error) {
	list, ok := args[0].([]interface{})
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "slice expects a list", lang.ErrArgumentMismatch)
	}
	// Use helper to handle numeric types robustly
	startRaw, okStart := toInt64(args[1])
	endRaw, okEnd := toInt64(args[2])

	if !okStart {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("slice expects an integer for start index, got %T", args[1]), lang.ErrArgumentMismatch)
	}
	if !okEnd {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("slice expects an integer for end index, got %T", args[2]), lang.ErrArgumentMismatch)
	}

	listLen := len(list)
	s := int(startRaw)
	e := int(endRaw)

	// Clamp indices according to Go slice semantics
	if s < 0 {
		s = 0
	}
	// Allow s == listLen for empty slice result
	if s > listLen {
		s = listLen
	}
	if e < s { // If end is before start after clamping start, result is empty
		e = s
	}
	if e > listLen {
		e = listLen
	}

	// Return a *copy* of the slice to maintain immutability principle
	subSlice := list[s:e]
	newList := make([]interface{}, len(subSlice))
	copy(newList, subSlice)
	return newList, nil
}

func toolListContains(_ tool.Runtime, args []interface{}) (interface{}, error) {
	list, ok := args[0].([]interface{})
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "contains expects a list", lang.ErrArgumentMismatch)
	}
	element := args[1]
	for _, item := range list {
		if reflect.DeepEqual(item, element) {
			return true, nil
		}
	}
	return false, nil
}

func toolListReverse(_ tool.Runtime, args []interface{}) (interface{}, error) {
	list, ok := args[0].([]interface{})
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "reverse expects a list", lang.ErrArgumentMismatch)
	}
	listLen := len(list)
	// Create a new list for the result
	newList := make([]interface{}, listLen)
	for i := 0; i < listLen; i++ {
		newList[i] = list[listLen-1-i]
	}
	return newList, nil
}

func toolListSort(_ tool.Runtime, args []interface{}) (interface{}, error) {
	list, ok := args[0].([]interface{})
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "sort expects a list", lang.ErrArgumentMismatch)
	}
	if len(list) == 0 {
		return []interface{}{}, nil
	}

	// Create a copy to sort, leaving original untouched
	newList := make([]interface{}, len(list))
	copy(newList, list)

	// Determine type for sorting (string, number, or mixed/invalid)
	var sortType string // "string", "number", "mixed"
	for i, v := range newList {
		currentType := ""
		if v == nil { // Cannot sort lists with nil
			sortType = "mixed"
			break
		}
		if _, ok := v.(string); ok {
			currentType = "string"
		} else if _, ok := lang.ToFloat64(v); ok { // Check if convertible to number
			currentType = "number"
		} else {
			// Check for explicitly non-sortable types
			switch v.(type) {
			case bool, []interface{}, map[string]interface{}:
				sortType = "mixed" // Mark as mixed/invalid if non-sortable found
			default:
				currentType = "mixed" // Unknown type, treat as non-sortable for now
			}
		}

		if sortType == "mixed" { // Already determined invalid
			break
		}

		if i == 0 {
			sortType = currentType
		} else if sortType != currentType {
			sortType = "mixed" // Found different types
			break
		}
	}

	if sortType == "mixed" {
		// Provide a more specific error
		for _, v := range newList {
			if v == nil {
				return nil, lang.NewRuntimeError(lang.ErrorCodeType, "cannot sort list with nil elements", lang.ErrListCannotSortMixedTypes)
			}
			switch v.(type) {
			case bool, []interface{}, map[string]interface{}:
				return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("cannot sort list with non-sortable type %T", v), lang.ErrListCannotSortMixedTypes)
			}
		}
		// If no explicitly non-sortable types found, it must be mixed numbers/strings
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, "cannot sort list with mixed string and number types", lang.ErrListCannotSortMixedTypes)
	}

	// Perform the sort based on determined type
	if sortType == "string" {
		sort.SliceStable(newList, func(i, j int) bool {
			// We already know these are strings
			return newList[i].(string) < newList[j].(string)
		})
	} else { // sortType == "number"
		sort.SliceStable(newList, func(i, j int) bool {
			// We know these are convertible to float64
			// Use Must because we've already validated convertibility
			ni, _ := lang.ToFloat64(newList[i])
			nj, _ := lang.ToFloat64(newList[j])
			return ni < nj
		})
	}

	return newList, nil
}

func toolListHead(_ tool.Runtime, args []interface{}) (interface{}, error) {
	list, ok := args[0].([]interface{})
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "head expects a list", lang.ErrArgumentMismatch)
	}
	if len(list) == 0 {
		return nil, nil
	}
	return list[0], nil
}

func toolListRest(_ tool.Runtime, args []interface{}) (interface{}, error) {
	list, ok := args[0].([]interface{})
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "rest expects a list", lang.ErrArgumentMismatch)
	}
	if len(list) <= 1 {
		return []interface{}{}, nil
	}
	// Return a *copy* of the rest of the slice
	subSlice := list[1:]
	newList := make([]interface{}, len(subSlice))
	copy(newList, subSlice)
	return newList, nil
}

func toolListTail(_ tool.Runtime, args []interface{}) (interface{}, error) {
	list, ok := args[0].([]interface{})
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "tail expects a list", lang.ErrArgumentMismatch)
	}
	// Use helper to handle numeric types robustly
	countRaw, okCount := toInt64(args[1])
	if !okCount {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("tail expects an integer for count, got %T", args[1]), lang.ErrArgumentMismatch)
	}

	listLen := len(list)
	c := int(countRaw)

	if c <= 0 {
		return []interface{}{}, nil
	}
	// Clamp start index calculation
	start := listLen - c
	if start < 0 {
		start = 0
	}
	// The end index is implicitly listLen

	// Return a *copy* to maintain immutability principle for most list ops
	tailSlice := list[start:]
	newList := make([]interface{}, len(tailSlice))
	copy(newList, tailSlice)
	return newList, nil

}

func toolListIsEmpty(_ tool.Runtime, args []interface{}) (interface{}, error) {
	list, ok := args[0].([]interface{})
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "isEmpty expects a list", lang.ErrArgumentMismatch)
	}
	return len(list) == 0, nil
}

// toInt64 robustly converts an interface{} to int64, handling float64.
// It returns false if the input is nil or not a whole number.
func toInt64(v interface{}) (int64, bool) {
	if v == nil {
		return 0, false // Treat nil as invalid for list indices/counts
	}
	if i, ok := v.(int64); ok {
		return i, true
	}
	if f, ok := v.(float64); ok {
		// Ensure it's a whole number
		if f == float64(int64(f)) {
			return int64(f), true
		}
	}
	// Allow int type as well
	if i, ok := v.(int); ok {
		return int64(i), true
	}

	return 0, false
}
