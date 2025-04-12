// filename: pkg/core/tools_list_impl.go
package core

import (
	"fmt"
	"reflect"
	"sort"
	// Added for ListSort error message check potentially
)

// --- List Tool Implementations ---

func toolListLength(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	list, ok := args[0].([]interface{})
	if !ok {
		// Handle the case where input might be nil directly without being []interface{}
		if args[0] == nil {
			return int64(0), nil // nil list has length 0
		}
		// Should be caught by validation, but handle defensively
		return nil, fmt.Errorf("TOOL.ListLength internal error: input was not []interface{} or nil, got %T", args[0])
	}
	// If it *is* []interface{}, check length (also covers nil slice case correctly)
	return int64(len(list)), nil
}

func toolListAppend(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	list, ok := args[0].([]interface{})
	element := args[1]
	if !ok {
		if args[0] == nil { // Handle nil input list gracefully
			list = []interface{}{} // Treat nil as empty list
		} else {
			return nil, fmt.Errorf("TOOL.ListAppend internal error: input list was not []interface{} or nil, got %T", args[0])
		}
	}
	// Create a new slice with capacity for one more element
	newList := make([]interface{}, 0, len(list)+1)
	newList = append(newList, list...) // Append original elements
	newList = append(newList, element) // Append the new element
	return newList, nil
}

func toolListPrepend(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	list, ok := args[0].([]interface{})
	element := args[1]
	if !ok {
		if args[0] == nil { // Handle nil input list gracefully
			list = []interface{}{} // Treat nil as empty list
		} else {
			return nil, fmt.Errorf("TOOL.ListPrepend internal error: input list was not []interface{} or nil, got %T", args[0])
		}
	}
	// Create a new slice with capacity for one more element
	newList := make([]interface{}, 0, len(list)+1)
	newList = append(newList, element) // Prepend the new element
	newList = append(newList, list...) // Append original elements
	return newList, nil
}

func toolListGet(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	list, okList := args[0].([]interface{})
	indexVal, okIdx := args[1].(int64)
	defaultValue := interface{}(nil) // Default to nil
	hasDefault := false
	if len(args) > 2 {
		defaultValue = args[2]
		hasDefault = true
	}

	if !okList {
		if args[0] == nil { // Handle nil input list gracefully
			list = []interface{}{} // Treat nil as empty list
		} else {
			return nil, fmt.Errorf("TOOL.ListGet internal error: input list was not []interface{} or nil, got %T", args[0])
		}
	}
	if !okIdx {
		// Should be caught by validation
		return nil, fmt.Errorf("TOOL.ListGet internal error: index was not int64, got %T", args[1])
	}

	index := int(indexVal)
	if list == nil || index < 0 || index >= len(list) {
		if hasDefault {
			return defaultValue, nil
		}
		return nil, nil // Return nil if index out of bounds and no default
	}
	return list[index], nil
}

func toolListSlice(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	list, okList := args[0].([]interface{})
	startVal, okStart := args[1].(int64)
	endVal, okEnd := args[2].(int64)

	if !okList {
		if args[0] == nil { // Handle nil input list gracefully
			return []interface{}{}, nil // Return empty slice for nil input
		} else {
			return nil, fmt.Errorf("TOOL.ListSlice internal error: input list was not []interface{} or nil, got %T", args[0])
		}
	}
	if !okStart {
		return nil, fmt.Errorf("TOOL.ListSlice internal error: start index was not int64, got %T", args[1])
	}
	if !okEnd {
		return nil, fmt.Errorf("TOOL.ListSlice internal error: end index was not int64, got %T", args[2])
	}

	// Now we know list is non-nil []interface{}
	listLen := len(list)
	start := int(startVal)
	end := int(endVal)

	// Clamp indices like Go slices
	if start < 0 {
		start = 0
	}
	if end > listLen {
		end = listLen
	}
	if start > end || start >= listLen {
		return []interface{}{}, nil // Return empty slice if range is invalid
	}

	// Create and return a *new* slice
	slice := make([]interface{}, end-start)
	copy(slice, list[start:end])
	return slice, nil
}

func toolListContains(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	list, okList := args[0].([]interface{})
	element := args[1]
	if !okList {
		if args[0] == nil { // Handle nil input list gracefully
			return false, nil // Nil list cannot contain element
		} else {
			return nil, fmt.Errorf("TOOL.ListContains internal error: input list was not []interface{} or nil, got %T", args[0])
		}
	}

	// Now we know list is non-nil []interface{}
	for _, item := range list {
		if reflect.DeepEqual(item, element) {
			return true, nil
		}
	}
	return false, nil
}

func toolListReverse(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	list, okList := args[0].([]interface{})
	if !okList {
		if args[0] == nil { // Handle nil input list gracefully
			return []interface{}{}, nil // Return empty slice for nil input
		} else {
			return nil, fmt.Errorf("TOOL.ListReverse internal error: input list was not []interface{} or nil, got %T", args[0])
		}
	}

	// Now we know list is non-nil []interface{}
	listLen := len(list)
	newList := make([]interface{}, listLen)
	for i := 0; i < listLen; i++ {
		newList[i] = list[listLen-1-i]
	}
	return newList, nil
}

func toolListSort(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	list, okList := args[0].([]interface{})
	if !okList {
		// Handle nil list case explicitly based on validation outcome
		if args[0] == nil {
			// If validation allowed optional nil, return empty list
			return []interface{}{}, nil
		}
		return nil, fmt.Errorf("%w: TOOL.ListSort internal error: input list was not []interface{}, got %T", ErrInternalTool, args[0])
	}

	if list == nil || len(list) == 0 {
		return []interface{}{}, nil // Return empty slice for empty input
	}

	// --- Refined Restriction Check ---
	canSortNumerically := true
	canSortLexicographically := true
	firstElemTypeKnown := false
	var firstKind reflect.Kind

	for _, elem := range list {
		var currentKind reflect.Kind
		isNumeric := false
		isString := false

		// Determine type/kind of current element
		if elem == nil { // Cannot sort lists with nil elements easily
			canSortNumerically = false
			canSortLexicographically = false
			break
		}
		currentKind = reflect.TypeOf(elem).Kind()
		if _, numOK := ToNumeric(elem); numOK { // Check numeric convertibility
			isNumeric = true
			if currentKind != reflect.String { // Treat actual numbers as numeric kind
				currentKind = reflect.Float64 // Use Float64 as the canonical numeric kind for comparison
			}
		}
		if _, strOK := elem.(string); strOK {
			isString = true
			currentKind = reflect.String // Use String kind
		}

		// Initialize first kind on first element
		if !firstElemTypeKnown {
			firstKind = currentKind
			firstElemTypeKnown = true
			// Set initial sortability based on the first element
			canSortNumerically = isNumeric
			canSortLexicographically = isString
		} else {
			// Check consistency with the first element's sortable type
			if canSortNumerically && (!isNumeric || currentKind != reflect.Float64) {
				// If we thought it was numeric, but this isn't, invalidate numeric sort
				canSortNumerically = false
			}
			if canSortLexicographically && (!isString || currentKind != reflect.String) {
				// If we thought it was string, but this isn't, invalidate string sort
				canSortLexicographically = false
			}
		}

		// Early exit if neither type is possible anymore
		if !canSortNumerically && !canSortLexicographically {
			break
		}
	}

	// If neither type is uniformly possible, return defined error
	if !canSortNumerically && !canSortLexicographically {
		errMsg := fmt.Sprintf("list contains mixed or non-sortable types (e.g., first element type kind: %v)", firstKind)
		if interpreter.logger != nil {
			interpreter.logger.Printf("[TOOL ListSort] Error: %s", errMsg)
		}
		// *** Return defined error ***
		return nil, fmt.Errorf("%w: %s", ErrListCannotSortMixedTypes, errMsg)
	}

	// --- Perform Sorting ---
	newList := make([]interface{}, len(list))
	copy(newList, list)

	if canSortNumerically {
		// Sort numerically based on float64 conversion
		if interpreter.logger != nil {
			interpreter.logger.Printf("[TOOL ListSort] Sorting numerically.")
		}
		sort.SliceStable(newList, func(i, j int) bool {
			fI, _ := toFloat64(newList[i]) // Safe conversion due to checks above
			fJ, _ := toFloat64(newList[j])
			return fI < fJ
		})
	} else { // Must be lexicographically sortable (all strings)
		if interpreter.logger != nil {
			interpreter.logger.Printf("[TOOL ListSort] Sorting lexicographically.")
		}
		sort.SliceStable(newList, func(i, j int) bool {
			// Safe to assume string type here due to checks above
			sI := fmt.Sprintf("%v", newList[i])
			sJ := fmt.Sprintf("%v", newList[j])
			return sI < sJ
		})
	}

	return newList, nil // Return sorted list and nil error
}

func toolListHead(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	list, okList := args[0].([]interface{})
	if !okList {
		if args[0] == nil { // Handle nil input list gracefully
			return nil, nil // Return nil for nil input
		} else {
			return nil, fmt.Errorf("TOOL.ListHead internal error: input list was not []interface{} or nil, got %T", args[0])
		}
	}

	if list == nil || len(list) == 0 {
		return nil, nil // Return nil for empty list
	}
	return list[0], nil
}

func toolListRest(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	list, okList := args[0].([]interface{})
	if !okList {
		if args[0] == nil { // Handle nil input list gracefully
			return []interface{}{}, nil // Return empty for nil input
		} else {
			return nil, fmt.Errorf("TOOL.ListRest internal error: input list was not []interface{} or nil, got %T", args[0])
		}
	}

	if list == nil || len(list) <= 1 {
		return []interface{}{}, nil // Return empty slice
	}
	// Create and return a new slice
	newList := make([]interface{}, len(list)-1)
	copy(newList, list[1:])
	return newList, nil
}

func toolListTail(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	list, okList := args[0].([]interface{})
	countVal, okCount := args[1].(int64)

	if !okList {
		if args[0] == nil { // Handle nil input list gracefully
			return []interface{}{}, nil // Return empty for nil input
		} else {
			return nil, fmt.Errorf("TOOL.ListTail internal error: input list was not []interface{} or nil, got %T", args[0])
		}
	}
	if !okCount {
		return nil, fmt.Errorf("TOOL.ListTail internal error: count was not int64, got %T", args[1])
	}

	// Now we know list is non-nil []interface{}
	listLen := len(list)
	count := int(countVal)

	if count <= 0 {
		return []interface{}{}, nil // Return empty if count is non-positive
	}
	if count >= listLen {
		// Return a copy of the original list
		newList := make([]interface{}, listLen)
		copy(newList, list)
		return newList, nil
	}

	// Calculate start index for the tail
	startIndex := listLen - count
	// Create and return a new slice
	newList := make([]interface{}, count)
	copy(newList, list[startIndex:])
	return newList, nil
}

func toolListIsEmpty(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	list, okList := args[0].([]interface{})
	if !okList {
		if args[0] == nil {
			return true, nil // nil list is empty
		}
		// If it's not []interface{} and not nil, validation should have caught it.
		// Return an internal error.
		return nil, fmt.Errorf("TOOL.ListIsEmpty internal error: input was not []interface{} or nil, got %T", args[0])
	}
	// If it *is* []interface{}, check length (also covers nil slice case correctly)
	return len(list) == 0, nil
}
