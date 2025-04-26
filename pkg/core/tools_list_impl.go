// filename: pkg/core/tools_list_impl.go
package core

import (
	"fmt"
	"reflect"
	"sort"
)

// --- List Tool Implementations ---

func toolListLength(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Validation ensures args[0] is []interface{}
	list := args[0].([]interface{})
	return int64(len(list)), nil
}

func toolListAppend(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Validation ensures args[0] is []interface{} and args[1] is any
	list := args[0].([]interface{})
	element := args[1]

	// Create a new slice with capacity for one more element
	newList := make([]interface{}, 0, len(list)+1)
	newList = append(newList, list...) // Append original elements
	newList = append(newList, element) // Append the new element
	return newList, nil
}

func toolListPrepend(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Validation ensures args[0] is []interface{} and args[1] is any
	list := args[0].([]interface{})
	element := args[1]

	// Create a new slice with capacity for one more element
	newList := make([]interface{}, 0, len(list)+1)
	newList = append(newList, element) // Prepend the new element
	newList = append(newList, list...) // Append original elements
	return newList, nil
}

func toolListGet(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Validation ensures args[0] is []interface{}, args[1] is int64, args[2] is optional any
	list := args[0].([]interface{})
	indexVal := args[1].(int64)
	defaultValue := interface{}(nil) // Default to nil
	hasDefault := false
	if len(args) > 2 && args[2] != nil { // Check if default was provided and is not explicitly nil
		defaultValue = args[2]
		hasDefault = true
	}

	index := int(indexVal)
	if index < 0 || index >= len(list) {
		if hasDefault {
			return defaultValue, nil
		}
		// Return nil if index out of bounds and no default OR default was explicitly nil
		return nil, nil
	}
	return list[index], nil
}

func toolListSlice(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Validation ensures args[0] is []interface{}, args[1] and args[2] are int64
	list := args[0].([]interface{})
	startVal := args[1].(int64)
	endVal := args[2].(int64)

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
	// Handle invalid ranges resulting from clamping or initial values
	if start > end || start >= listLen {
		return []interface{}{}, nil // Return empty slice if range is invalid
	}

	// Create and return a *new* slice
	slice := make([]interface{}, end-start)
	copy(slice, list[start:end])
	return slice, nil
}

func toolListContains(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Validation ensures args[0] is []interface{} and args[1] is any (and not nil)
	list := args[0].([]interface{})
	element := args[1]

	for _, item := range list {
		// Handle nil in list explicitly during comparison
		if item == nil && element == nil {
			return true, nil
		}
		if item != nil && element != nil && reflect.DeepEqual(item, element) {
			return true, nil
		}
	}
	return false, nil
}

func toolListReverse(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Validation ensures args[0] is []interface{}
	list := args[0].([]interface{})

	listLen := len(list)
	newList := make([]interface{}, listLen)
	for i := 0; i < listLen; i++ {
		newList[i] = list[listLen-1-i]
	}
	return newList, nil
}

func toolListSort(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Validation ensures args[0] is []interface{}
	list := args[0].([]interface{})

	if len(list) == 0 {
		return []interface{}{}, nil // Return empty slice for empty input
	}

	// --- Refined Restriction Check ---
	canSortNumerically := true
	canSortLexicographically := true
	firstElemTypeKnown := false
	var firstKind reflect.Kind

	for i, elem := range list {
		var currentKind reflect.Kind
		isNumeric := false
		isString := false

		// Determine type/kind of current element
		if elem == nil { // Cannot sort lists with nil elements easily
			errMsg := fmt.Sprintf("list contains nil element at index %d", i)
			if interpreter.logger != nil {
				interpreter.logger.Info("Tool: ListSort] Error: %s", errMsg)
			}
			return nil, fmt.Errorf("%w: %s", ErrListCannotSortMixedTypes, errMsg)
		}

		currentKind = reflect.TypeOf(elem).Kind()
		// Check numeric convertibility *specifically*
		if _, numOK := ToNumeric(elem); numOK {
			// Further check if it's *actually* a number type or a string that looks like one
			switch currentKind {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
				reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
				reflect.Float32, reflect.Float64:
				isNumeric = true
				currentKind = reflect.Float64 // Use Float64 as the canonical numeric kind for comparison
			case reflect.String:
				// It's a string that might parse as a number, but treat as string for sort unless *all* are numeric
				isString = true
				currentKind = reflect.String
			default:
				// Neither a standard number nor string
				isNumeric = false
				isString = false
			}
		} else if currentKind == reflect.String {
			isString = true
			currentKind = reflect.String
		} else {
			// Not convertible to number and not a string
			isNumeric = false
			isString = false
		}

		// Initialize first kind on first element
		if !firstElemTypeKnown {
			firstKind = currentKind
			firstElemTypeKnown = true
			// Set initial sortability based on the first element
			canSortNumerically = isNumeric
			canSortLexicographically = isString // Can only sort strings if first element IS a string
		} else {
			// Check consistency with the first element's determined sortable type
			// If we started numeric, this one must also be numeric
			if canSortNumerically && (!isNumeric || currentKind != reflect.Float64) {
				canSortNumerically = false
			}
			// If we started string, this one must also be string
			if canSortLexicographically && (!isString || currentKind != reflect.String) {
				canSortLexicographically = false
			}
		}

		// Early exit if neither type is possible anymore
		if !canSortNumerically && !canSortLexicographically {
			errMsg := fmt.Sprintf("list contains mixed or non-sortable types (e.g., element %d [%v] type %T is incompatible with first element type kind %v)", i, elem, elem, firstKind)
			if interpreter.logger != nil {
				interpreter.logger.Info("Tool: ListSort] Error: %s", errMsg)
			}
			return nil, fmt.Errorf("%w: %s", ErrListCannotSortMixedTypes, errMsg)

		}
	}

	// If neither type is uniformly possible after checking all elements (should have exited above, but double-check)
	if !canSortNumerically && !canSortLexicographically {
		errMsg := fmt.Sprintf("list contains mixed or non-sortable types (final check, first kind: %v)", firstKind)
		if interpreter.logger != nil {
			interpreter.logger.Info("Tool: ListSort] Error: %s", errMsg)
		}
		return nil, fmt.Errorf("%w: %s", ErrListCannotSortMixedTypes, errMsg)
	}

	// --- Perform Sorting ---
	newList := make([]interface{}, len(list))
	copy(newList, list)

	if canSortNumerically {
		// Sort numerically based on float64 conversion
		if interpreter.logger != nil {
			interpreter.logger.Info("Tool: ListSort] Sorting numerically.")
		}
		sort.SliceStable(newList, func(i, j int) bool {
			// Safe conversion due to checks above (all elements were verified convertible to numeric)
			fI, _ := toFloat64(newList[i])
			fJ, _ := toFloat64(newList[j])
			return fI < fJ
		})
	} else { // Must be lexicographically sortable (all strings)
		if interpreter.logger != nil {
			interpreter.logger.Info("Tool: ListSort] Sorting lexicographically.")
		}
		sort.SliceStable(newList, func(i, j int) bool {
			// Safe to assume string type here due to checks above
			sI := newList[i].(string) // Direct assertion is safe now
			sJ := newList[j].(string)
			return sI < sJ
		})
	}

	return newList, nil // Return sorted list and nil error
}

// --- NEW ListHead ---
func toolListHead(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Validation ensures args[0] is []interface{}
	list := args[0].([]interface{})

	if len(list) == 0 {
		return nil, nil // Return nil for empty list
	}
	return list[0], nil
}

// --- NEW ListRest ---
func toolListRest(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Validation ensures args[0] is []interface{}
	list := args[0].([]interface{})

	if len(list) <= 1 {
		return []interface{}{}, nil // Return empty slice if 0 or 1 element
	}
	// Create and return a new slice containing elements from index 1 onwards
	newList := make([]interface{}, len(list)-1)
	copy(newList, list[1:])
	return newList, nil
}

// --- NEW ListTail ---
func toolListTail(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Validation ensures args[0] is []interface{}, args[1] is int64
	list := args[0].([]interface{})
	countVal := args[1].(int64)

	listLen := len(list)
	count := int(countVal)

	// Handle count logic
	if count <= 0 {
		return []interface{}{}, nil // Return empty if count is non-positive
	}
	if count >= listLen {
		// Return a copy of the original list if count is >= length
		newList := make([]interface{}, listLen)
		copy(newList, list)
		return newList, nil
	}

	// Calculate start index for the tail
	startIndex := listLen - count
	// Create and return a new slice containing the last 'count' elements
	newList := make([]interface{}, count)
	copy(newList, list[startIndex:])
	return newList, nil
}

// --- Existing ListIsEmpty ---
func toolListIsEmpty(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Validation ensures args[0] is []interface{}
	list := args[0].([]interface{})
	return len(list) == 0, nil
}
