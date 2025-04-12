// filename: pkg/core/tools_list_test.go
package core

import (
	"errors" // Import errors package
	"reflect"

	// "strings" // No longer needed for error checking
	"testing"
)

// Assume newDefaultTestInterpreter and makeArgs are defined in testing_helpers.go

// testListToolHelper encapsulates the logic for validating and executing a list tool test case.
// Uses errors.Is exclusively for error checking.
func testListToolHelper(t *testing.T, interp *Interpreter, tc struct {
	name          string
	toolName      string
	args          []interface{}
	wantResult    interface{} // Expected result *if* no error
	wantToolErrIs error       // Specific Go error expected *from the tool function*
	valWantErrIs  error       // Specific Go error expected *from validation*
}) {
	t.Helper()
	t.Run(tc.name, func(t *testing.T) {
		toolImpl, found := interp.ToolRegistry().GetTool(tc.toolName)
		if !found {
			t.Fatalf("Tool %q not found in registry", tc.toolName)
		}
		spec := toolImpl.Spec

		// --- Validation ---
		convertedArgs, valErr := ValidateAndConvertArgs(spec, tc.args)

		// Check Specific Validation Error
		if tc.valWantErrIs != nil {
			if valErr == nil {
				t.Errorf("ValidateAndConvertArgs() expected error [%v], but got nil", tc.valWantErrIs)
			} else if !errors.Is(valErr, tc.valWantErrIs) {
				t.Errorf("ValidateAndConvertArgs() expected error type [%v], but got type [%T] with value: %v", tc.valWantErrIs, valErr, valErr)
			}
			// Regardless of match details, if specific error was expected, stop.
			return
		}

		// Check for Unexpected Validation Error
		if valErr != nil && tc.valWantErrIs == nil {
			t.Fatalf("ValidateAndConvertArgs() unexpected validation error: %v", valErr)
		}

		// --- Execution (Only if validation passed and wasn't expected to fail) ---
		gotResult, toolErr := toolImpl.Func(interp, convertedArgs)

		// Check Specific Tool Error
		if tc.wantToolErrIs != nil {
			if toolErr == nil {
				t.Errorf("Tool function expected error [%v], but got nil. Result: %v (%T)", tc.wantToolErrIs, gotResult, gotResult)
			} else if !errors.Is(toolErr, tc.wantToolErrIs) {
				t.Errorf("Tool function expected error type [%v], but got type [%T] with value: %v", tc.wantToolErrIs, toolErr, toolErr)
			}
			// If specific tool error was expected, don't check result
			return
		}

		// Check for Unexpected Tool Error
		if toolErr != nil && tc.wantToolErrIs == nil {
			t.Fatalf("Tool function unexpected error: %v. Result: %v (%T)", toolErr, gotResult, gotResult)
		}

		// --- Result Comparison (only if no errors occurred or were expected via wantToolErrIs) ---
		if tc.wantToolErrIs == nil { // Only compare results if no specific tool error was expected
			if !reflect.DeepEqual(gotResult, tc.wantResult) {
				t.Errorf("Tool function result mismatch:\n  Got:  %#v (%T)\n  Want: %#v (%T)",
					gotResult, gotResult, tc.wantResult, tc.wantResult)
			}
		}
	})
}

// --- Test Functions for Each Tool (with corrected expectations) ---

func TestToolListLength(t *testing.T) {
	interp, _ := newDefaultTestInterpreter(t)
	tests := []struct {
		name          string
		toolName      string
		args          []interface{}
		wantResult    interface{}
		wantToolErrIs error
		valWantErrIs  error
	}{
		{name: "Empty", toolName: "ListLength", args: makeArgs([]interface{}{}), wantResult: int64(0)},
		{name: "Simple", toolName: "ListLength", args: makeArgs([]interface{}{1, "a", true}), wantResult: int64(3)},
		// *** FIXED: Use interface{}(nil) for nil list arg ***
		{name: "Nil List", toolName: "ListLength", args: makeArgs(interface{}(nil)), valWantErrIs: ErrValidationRequiredArgNil},
		// *** End Fix ***
		{name: "Wrong Type", toolName: "ListLength", args: makeArgs("not a list"), valWantErrIs: ErrValidationTypeMismatch},
		{name: "Wrong Count", toolName: "ListLength", args: makeArgs(), valWantErrIs: ErrValidationArgCount},
	}
	for _, tt := range tests {
		testListToolHelper(t, interp, tt)
	}
}

func TestToolListAppendPrepend(t *testing.T) {
	interp, _ := newDefaultTestInterpreter(t)
	baseList := []interface{}{"a", int64(1)}
	tests := []struct {
		name          string
		toolName      string // Append or Prepend
		args          []interface{}
		wantResult    interface{}
		wantToolErrIs error
		valWantErrIs  error
	}{
		{name: "Append Simple", toolName: "ListAppend", args: makeArgs(baseList, true), wantResult: []interface{}{"a", int64(1), true}},
		{name: "Append To Empty", toolName: "ListAppend", args: makeArgs([]interface{}{}, "new"), wantResult: []interface{}{"new"}},
		// *** FIXED: Use interface{}(nil) for nil list arg ***
		{name: "Append To Nil", toolName: "ListAppend", args: makeArgs(interface{}(nil), "new"), valWantErrIs: ErrValidationRequiredArgNil},
		// *** End Fix ***
		{name: "Append Wrong Type", toolName: "ListAppend", args: makeArgs("not list", "el"), valWantErrIs: ErrValidationTypeMismatch},
		{name: "Prepend Simple", toolName: "ListPrepend", args: makeArgs(baseList, true), wantResult: []interface{}{true, "a", int64(1)}},
		{name: "Prepend To Empty", toolName: "ListPrepend", args: makeArgs([]interface{}{}, "new"), wantResult: []interface{}{"new"}},
		// *** FIXED: Use interface{}(nil) for nil list arg ***
		{name: "Prepend To Nil", toolName: "ListPrepend", args: makeArgs(interface{}(nil), "new"), valWantErrIs: ErrValidationRequiredArgNil},
		// *** End Fix ***
		{name: "Prepend Wrong Count", toolName: "ListPrepend", args: makeArgs(baseList), valWantErrIs: ErrValidationArgCount},
	}
	for _, tt := range tests {
		testListToolHelper(t, interp, tt)
	}
}

func TestToolListGet(t *testing.T) {
	interp, _ := newDefaultTestInterpreter(t)
	list := []interface{}{"a", int64(1), true}
	tests := []struct {
		name          string
		toolName      string
		args          []interface{}
		wantResult    interface{}
		wantToolErrIs error
		valWantErrIs  error
	}{
		{name: "Get First", toolName: "ListGet", args: makeArgs(list, int64(0)), wantResult: "a"},
		{name: "Get Middle", toolName: "ListGet", args: makeArgs(list, int64(1)), wantResult: int64(1)},
		{name: "Get Last", toolName: "ListGet", args: makeArgs(list, int64(2)), wantResult: true},
		{name: "OOB High No Default", toolName: "ListGet", args: makeArgs(list, int64(3)), wantResult: nil},
		{name: "OOB Low No Default", toolName: "ListGet", args: makeArgs(list, int64(-1)), wantResult: nil},
		{name: "Empty No Default", toolName: "ListGet", args: makeArgs([]interface{}{}, int64(0)), wantResult: nil},
		{name: "OOB High With Default", toolName: "ListGet", args: makeArgs(list, int64(5), "default"), wantResult: "default"},
		{name: "OOB Low With Default", toolName: "ListGet", args: makeArgs(list, int64(-2), false), wantResult: false},
		{name: "Empty With Default", toolName: "ListGet", args: makeArgs([]interface{}{}, int64(0), "def"), wantResult: "def"},
		// *** FIXED: Use interface{}(nil) for nil list arg ***
		{name: "Nil List No Default", toolName: "ListGet", args: makeArgs(interface{}(nil), int64(0)), valWantErrIs: ErrValidationRequiredArgNil},
		{name: "Nil List With Default", toolName: "ListGet", args: makeArgs(interface{}(nil), int64(0), "def"), valWantErrIs: ErrValidationRequiredArgNil},
		// *** End Fix ***
		{name: "Wrong Index Type Coerced", toolName: "ListGet", args: makeArgs(list, "1"), wantResult: int64(1)},
	}
	for _, tt := range tests {
		testListToolHelper(t, interp, tt)
	}
}

func TestToolListSlice(t *testing.T) {
	interp, _ := newDefaultTestInterpreter(t)
	list := []interface{}{"a", "b", "c", "d", "e"}
	tests := []struct {
		name          string
		toolName      string
		args          []interface{}
		wantResult    interface{}
		wantToolErrIs error
		valWantErrIs  error
	}{
		{name: "Slice Middle", toolName: "ListSlice", args: makeArgs(list, int64(1), int64(4)), wantResult: []interface{}{"b", "c", "d"}},
		{name: "Slice Start", toolName: "ListSlice", args: makeArgs(list, int64(0), int64(2)), wantResult: []interface{}{"a", "b"}},
		{name: "Slice End", toolName: "ListSlice", args: makeArgs(list, int64(3), int64(5)), wantResult: []interface{}{"d", "e"}},
		{name: "Slice Full", toolName: "ListSlice", args: makeArgs(list, int64(0), int64(5)), wantResult: []interface{}{"a", "b", "c", "d", "e"}},
		{name: "Slice Empty Start=End", toolName: "ListSlice", args: makeArgs(list, int64(2), int64(2)), wantResult: []interface{}{}},
		{name: "Slice Empty Start>End", toolName: "ListSlice", args: makeArgs(list, int64(3), int64(1)), wantResult: []interface{}{}},
		{name: "Slice Clamp High End", toolName: "ListSlice", args: makeArgs(list, int64(3), int64(10)), wantResult: []interface{}{"d", "e"}},
		{name: "Slice Clamp Low Start", toolName: "ListSlice", args: makeArgs(list, int64(-2), int64(2)), wantResult: []interface{}{"a", "b"}},
		{name: "Slice Clamp Both", toolName: "ListSlice", args: makeArgs(list, int64(-1), int64(10)), wantResult: []interface{}{"a", "b", "c", "d", "e"}},
		{name: "Slice Empty List", toolName: "ListSlice", args: makeArgs([]interface{}{}, int64(0), int64(1)), wantResult: []interface{}{}},
		// *** FIXED: Use interface{}(nil) for nil list arg ***
		{name: "Slice Nil List", toolName: "ListSlice", args: makeArgs(interface{}(nil), int64(0), int64(1)), valWantErrIs: ErrValidationRequiredArgNil},
		// *** End Fix ***
		{name: "Wrong End Type Coerced", toolName: "ListSlice", args: makeArgs(list, int64(0), "2"), wantResult: []interface{}{"a", "b"}},
	}
	for _, tt := range tests {
		testListToolHelper(t, interp, tt)
	}
}

func TestToolListContains(t *testing.T) {
	interp, _ := newDefaultTestInterpreter(t)
	list := []interface{}{"a", int64(1), true, nil, float64(1.0)}
	tests := []struct {
		name          string
		toolName      string
		args          []interface{}
		wantResult    interface{}
		wantToolErrIs error
		valWantErrIs  error
	}{
		{name: "Contains String", toolName: "ListContains", args: makeArgs(list, "a"), wantResult: true},
		{name: "Contains Int", toolName: "ListContains", args: makeArgs(list, int64(1)), wantResult: true},
		{name: "Contains Bool", toolName: "ListContains", args: makeArgs(list, true), wantResult: true},
		// *** Note: Element arg validation error expected for nil ***
		{name: "Contains Nil Element Arg", toolName: "ListContains", args: makeArgs(list, nil), valWantErrIs: ErrValidationRequiredArgNil},
		{name: "Contains Float", toolName: "ListContains", args: makeArgs(list, float64(1.0)), wantResult: true},
		{name: "Not Contains String", toolName: "ListContains", args: makeArgs(list, "b"), wantResult: false},
		{name: "Not Contains Int", toolName: "ListContains", args: makeArgs(list, int64(2)), wantResult: false},
		{name: "Not Contains Float", toolName: "ListContains", args: makeArgs(list, float64(1.1)), wantResult: false},
		{name: "Empty List", toolName: "ListContains", args: makeArgs([]interface{}{}, "a"), wantResult: false},
		// *** FIXED: Use interface{}(nil) for nil list arg ***
		{name: "Nil List", toolName: "ListContains", args: makeArgs(interface{}(nil), "a"), valWantErrIs: ErrValidationRequiredArgNil},
		// *** End Fix ***
	}
	for _, tt := range tests {
		testListToolHelper(t, interp, tt)
	}
}

func TestToolListReverse(t *testing.T) {
	interp, _ := newDefaultTestInterpreter(t)
	tests := []struct {
		name          string
		toolName      string
		args          []interface{}
		wantResult    interface{}
		wantToolErrIs error
		valWantErrIs  error
	}{
		{name: "Reverse Simple", toolName: "ListReverse", args: makeArgs([]interface{}{"a", int64(1), true}), wantResult: []interface{}{true, int64(1), "a"}},
		{name: "Reverse Single", toolName: "ListReverse", args: makeArgs([]interface{}{"a"}), wantResult: []interface{}{"a"}},
		{name: "Reverse Empty", toolName: "ListReverse", args: makeArgs([]interface{}{}), wantResult: []interface{}{}},
		// *** FIXED: Use interface{}(nil) for nil list arg ***
		{name: "Reverse Nil", toolName: "ListReverse", args: makeArgs(interface{}(nil)), valWantErrIs: ErrValidationRequiredArgNil},
		// *** End Fix ***
	}
	for _, tt := range tests {
		testListToolHelper(t, interp, tt)
	}
}

func TestToolListSort(t *testing.T) {
	interp, _ := newDefaultTestInterpreter(t)
	tests := []struct {
		name          string
		toolName      string
		args          []interface{}
		wantResult    interface{} // Can be []interface{}
		wantToolErrIs error       // Expect specific tool error
		valWantErrIs  error
	}{
		{name: "Sort Strings", toolName: "ListSort", args: makeArgs([]interface{}{"c", "a", "b"}), wantResult: []interface{}{"a", "b", "c"}},
		{name: "Sort Ints", toolName: "ListSort", args: makeArgs([]interface{}{int64(3), int64(1), int64(2)}), wantResult: []interface{}{int64(1), int64(2), int64(3)}},
		{name: "Sort Floats", toolName: "ListSort", args: makeArgs([]interface{}{float64(3.3), float64(1.1), float64(2.2)}), wantResult: []interface{}{float64(1.1), float64(2.2), float64(3.3)}},
		{name: "Sort Mixed Numbers", toolName: "ListSort", args: makeArgs([]interface{}{int64(3), float64(1.1), int64(2)}), wantResult: []interface{}{float64(1.1), int64(2), int64(3)}},
		{name: "Sort Empty", toolName: "ListSort", args: makeArgs([]interface{}{}), wantResult: []interface{}{}},
		// *** FIXED: Use interface{}(nil) for nil list arg ***
		{name: "Sort Nil", toolName: "ListSort", args: makeArgs(interface{}(nil)), valWantErrIs: ErrValidationRequiredArgNil},
		// *** End Fix ***
		{name: "Sort Single", toolName: "ListSort", args: makeArgs([]interface{}{"a"}), wantResult: []interface{}{"a"}},
		{name: "Sort Mixed String/Num", toolName: "ListSort", args: makeArgs([]interface{}{"a", int64(1)}), wantToolErrIs: ErrListCannotSortMixedTypes},
		{name: "Sort Mixed Num/Bool", toolName: "ListSort", args: makeArgs([]interface{}{int64(1), true}), wantToolErrIs: ErrListCannotSortMixedTypes},
		{name: "Sort List of Lists", toolName: "ListSort", args: makeArgs([]interface{}{[]interface{}{1}, []interface{}{0}}), wantToolErrIs: ErrListCannotSortMixedTypes},
		{name: "Sort List With Nil Element", toolName: "ListSort", args: makeArgs([]interface{}{"a", nil, "c"}), wantToolErrIs: ErrListCannotSortMixedTypes},
	}
	for _, tt := range tests {
		testListToolHelper(t, interp, tt)
	}
}

func TestToolListHeadRest(t *testing.T) {
	interp, _ := newDefaultTestInterpreter(t)
	list := []interface{}{"a", "b", int64(1)}
	singleList := []interface{}{"only"}
	emptyList := []interface{}{}
	// var nilList []interface{} = nil // No longer needed

	tests := []struct {
		name          string
		toolName      string // Head or Rest
		args          []interface{}
		wantResult    interface{}
		wantToolErrIs error
		valWantErrIs  error
	}{
		// Head
		{name: "Head Simple", toolName: "ListHead", args: makeArgs(list), wantResult: "a"},
		{name: "Head Single", toolName: "ListHead", args: makeArgs(singleList), wantResult: "only"},
		{name: "Head Empty", toolName: "ListHead", args: makeArgs(emptyList), wantResult: nil}, // Returns nil
		// *** FIXED: Use interface{}(nil) for nil list arg ***
		{name: "Head Nil", toolName: "ListHead", args: makeArgs(interface{}(nil)), valWantErrIs: ErrValidationRequiredArgNil},
		// *** End Fix ***
		// Rest
		{name: "Rest Simple", toolName: "ListRest", args: makeArgs(list), wantResult: []interface{}{"b", int64(1)}},
		{name: "Rest Single", toolName: "ListRest", args: makeArgs(singleList), wantResult: []interface{}{}},
		{name: "Rest Empty", toolName: "ListRest", args: makeArgs(emptyList), wantResult: []interface{}{}},
		// *** FIXED: Use interface{}(nil) for nil list arg ***
		{name: "Rest Nil", toolName: "ListRest", args: makeArgs(interface{}(nil)), valWantErrIs: ErrValidationRequiredArgNil},
		// *** End Fix ***
	}
	for _, tt := range tests {
		testListToolHelper(t, interp, tt)
	}
}

func TestToolListTail(t *testing.T) {
	interp, _ := newDefaultTestInterpreter(t)
	list := []interface{}{"a", "b", "c", "d", "e"}
	tests := []struct {
		name          string
		toolName      string
		args          []interface{}
		wantResult    interface{}
		wantToolErrIs error
		valWantErrIs  error
	}{
		{name: "Tail Last 2", toolName: "ListTail", args: makeArgs(list, int64(2)), wantResult: []interface{}{"d", "e"}},
		{name: "Tail Last 1", toolName: "ListTail", args: makeArgs(list, int64(1)), wantResult: []interface{}{"e"}},
		{name: "Tail Last 5 (All)", toolName: "ListTail", args: makeArgs(list, int64(5)), wantResult: []interface{}{"a", "b", "c", "d", "e"}},
		{name: "Tail Last 6 (>Len)", toolName: "ListTail", args: makeArgs(list, int64(6)), wantResult: []interface{}{"a", "b", "c", "d", "e"}},
		{name: "Tail Count 0", toolName: "ListTail", args: makeArgs(list, int64(0)), wantResult: []interface{}{}},
		{name: "Tail Count Negative", toolName: "ListTail", args: makeArgs(list, int64(-1)), wantResult: []interface{}{}},
		{name: "Tail Empty List", toolName: "ListTail", args: makeArgs([]interface{}{}, int64(2)), wantResult: []interface{}{}},
		// *** FIXED: Use interface{}(nil) for nil list arg ***
		{name: "Tail Nil List", toolName: "ListTail", args: makeArgs(interface{}(nil), int64(2)), valWantErrIs: ErrValidationRequiredArgNil},
		// *** End Fix ***
		{name: "Wrong Count Type Coerced", toolName: "ListTail", args: makeArgs(list, "2"), wantResult: []interface{}{"d", "e"}},
	}
	for _, tt := range tests {
		testListToolHelper(t, interp, tt)
	}
}

func TestToolListIsEmpty(t *testing.T) {
	interp, _ := newDefaultTestInterpreter(t)
	tests := []struct {
		name          string
		toolName      string
		args          []interface{}
		wantResult    interface{}
		wantToolErrIs error
		valWantErrIs  error
	}{
		{name: "Is Empty True", toolName: "ListIsEmpty", args: makeArgs([]interface{}{}), wantResult: true},
		// *** FIXED: Use interface{}(nil) for nil list arg ***
		{name: "Is Empty True (Nil)", toolName: "ListIsEmpty", args: makeArgs(interface{}(nil)), valWantErrIs: ErrValidationRequiredArgNil},
		// *** End Fix ***
		{name: "Is Empty False", toolName: "ListIsEmpty", args: makeArgs([]interface{}{"a"}), wantResult: false},
		{name: "Is Empty False Long", toolName: "ListIsEmpty", args: makeArgs([]interface{}{1, 2, 3}), wantResult: false},
		{name: "Wrong Type", toolName: "ListIsEmpty", args: makeArgs("not a list"), valWantErrIs: ErrValidationTypeMismatch},
	}
	for _, tt := range tests {
		testListToolHelper(t, interp, tt)
	}
}
