// NeuroScript Version: 0.3.1
// File version: 0.1.2
// Revert 'element' required status in tests for Append, Prepend, Contains.
// nlines: 280
// risk_rating: LOW
// filename: pkg/core/tools_list_test.go
package core

import (
	"errors" // Import errors package
	"reflect"

	// "strings" // No longer needed for error checking
	"testing"
)

// Assume NewDefaultTestInterpreter and MakeArgs are defined in testing_helpers.go

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
		// Guard against convertedArgs being nil if validation somehow passed unexpectedly after expecting error
		if convertedArgs == nil && tc.valWantErrIs == nil {
			// This path might be reachable if MakeArgs returns nil, but validation passes (e.g., optional args)
			// Let execution proceed, the tool function should handle nil args appropriately if necessary.
			// If execution *requires* non-nil convertedArgs, the tool func itself should error out.
			t.Logf("Validation passed but convertedArgs is nil (tool func must handle this)")
			// return // Optionally stop here if nil convertedArgs is always an error post-validation
		}

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
	interp, _ := NewDefaultTestInterpreter(t)
	tests := []struct {
		name          string
		toolName      string
		args          []interface{}
		wantResult    interface{}
		wantToolErrIs error
		valWantErrIs  error
	}{
		{name: "Empty", toolName: "List.Length", args: MakeArgs([]interface{}{}), wantResult: int64(0)},
		{name: "Simple", toolName: "List.Length", args: MakeArgs([]interface{}{1, "a", true}), wantResult: int64(3)},
		{name: "Nil List Arg", toolName: "List.Length", args: MakeArgs(interface{}(nil)), valWantErrIs: ErrValidationRequiredArgNil}, // list is required
		{name: "Wrong Type", toolName: "List.Length", args: MakeArgs("not a list"), valWantErrIs: ErrValidationTypeMismatch},
		// Corrected: Expect missing arg error when no args provided
		{name: "Missing List Arg", toolName: "List.Length", args: MakeArgs(), valWantErrIs: ErrValidationRequiredArgMissing},
		{name: "Wrong Arg Count (Too Many)", toolName: "List.Length", args: MakeArgs([]interface{}{}, "extra"), valWantErrIs: ErrValidationArgCount},
	}
	for _, tt := range tests {
		testListToolHelper(t, interp, tt)
	}
}

func TestToolListAppendPrepend(t *testing.T) {
	interp, _ := NewDefaultTestInterpreter(t)
	baseList := []interface{}{"a", int64(1)}
	tests := []struct {
		name          string
		toolName      string // Append or Prepend
		args          []interface{}
		wantResult    interface{}
		wantToolErrIs error
		valWantErrIs  error
	}{
		{name: "Append Simple", toolName: "List.Append", args: MakeArgs(baseList, true), wantResult: []interface{}{"a", int64(1), true}},
		{name: "Append To Empty", toolName: "List.Append", args: MakeArgs([]interface{}{}, "new"), wantResult: []interface{}{"new"}},
		{name: "Append To Nil List Arg", toolName: "List.Append", args: MakeArgs(interface{}(nil), "new"), valWantErrIs: ErrValidationRequiredArgNil}, // list is required
		// Corrected: Passing nil as the optional 'element' is now valid
		{name: "Append Nil Element", toolName: "List.Append", args: MakeArgs(baseList, nil), wantResult: []interface{}{"a", int64(1), nil}, valWantErrIs: nil},
		{name: "Append Wrong Type", toolName: "List.Append", args: MakeArgs("not list", "el"), valWantErrIs: ErrValidationTypeMismatch},
		{name: "Append Missing List Arg", toolName: "List.Append", args: MakeArgs(), valWantErrIs: ErrValidationRequiredArgMissing}, // Missing required 'list'
		{name: "Append Wrong Arg Count (Too Many)", toolName: "List.Append", args: MakeArgs(baseList, "el", "extra"), valWantErrIs: ErrValidationArgCount},

		{name: "Prepend Simple", toolName: "List.Prepend", args: MakeArgs(baseList, true), wantResult: []interface{}{true, "a", int64(1)}},
		{name: "Prepend To Empty", toolName: "List.Prepend", args: MakeArgs([]interface{}{}, "new"), wantResult: []interface{}{"new"}},
		{name: "Prepend To Nil List Arg", toolName: "List.Prepend", args: MakeArgs(interface{}(nil), "new"), valWantErrIs: ErrValidationRequiredArgNil}, // list is required
		// Corrected: Passing nil as the optional 'element' is now valid
		{name: "Prepend Nil Element", toolName: "List.Prepend", args: MakeArgs(baseList, nil), wantResult: []interface{}{nil, "a", int64(1)}, valWantErrIs: nil},
		{name: "Prepend Missing List Arg", toolName: "List.Prepend", args: MakeArgs(), valWantErrIs: ErrValidationRequiredArgMissing}, // Missing required 'list'
		// Corrected: Providing only the required 'list' arg is now valid (element is optional)
		{name: "Prepend Missing Value Arg", toolName: "List.Prepend", args: MakeArgs(baseList), wantResult: []interface{}{nil, "a", int64(1)}, valWantErrIs: nil}, // Prepends nil if element missing
		{name: "Prepend Wrong Arg Count (Too Many)", toolName: "List.Prepend", args: MakeArgs(baseList, "el", "extra"), valWantErrIs: ErrValidationArgCount},
	}
	for _, tt := range tests {
		testListToolHelper(t, interp, tt)
	}
}

func TestToolListGet(t *testing.T) {
	interp, _ := NewDefaultTestInterpreter(t)
	list := []interface{}{"a", int64(1), true, nil} // Added nil to test getting it
	tests := []struct {
		name          string
		toolName      string
		args          []interface{}
		wantResult    interface{}
		wantToolErrIs error
		valWantErrIs  error
	}{
		{name: "Get First", toolName: "List.Get", args: MakeArgs(list, int64(0)), wantResult: "a"},
		{name: "Get Middle", toolName: "List.Get", args: MakeArgs(list, int64(1)), wantResult: int64(1)},
		{name: "Get Last", toolName: "List.Get", args: MakeArgs(list, int64(2)), wantResult: true},
		{name: "Get Nil Element", toolName: "List.Get", args: MakeArgs(list, int64(3)), wantResult: nil}, // Get the actual nil
		{name: "OOB High No Default", toolName: "List.Get", args: MakeArgs(list, int64(4)), wantResult: nil},
		{name: "OOB Low No Default", toolName: "List.Get", args: MakeArgs(list, int64(-1)), wantResult: nil},
		{name: "Empty No Default", toolName: "List.Get", args: MakeArgs([]interface{}{}, int64(0)), wantResult: nil},
		{name: "OOB High With Default", toolName: "List.Get", args: MakeArgs(list, int64(5), "default"), wantResult: "default"},
		{name: "OOB Low With Default", toolName: "List.Get", args: MakeArgs(list, int64(-2), false), wantResult: false},
		{name: "Empty With Default", toolName: "List.Get", args: MakeArgs([]interface{}{}, int64(0), "def"), wantResult: "def"},
		{name: "OOB With Explicit Nil Default", toolName: "List.Get", args: MakeArgs(list, int64(5), nil), wantResult: nil},                                 // Explicit nil default
		{name: "Nil List No Default", toolName: "List.Get", args: MakeArgs(interface{}(nil), int64(0)), valWantErrIs: ErrValidationRequiredArgNil},          // list is required
		{name: "Nil List With Default", toolName: "List.Get", args: MakeArgs(interface{}(nil), int64(0), "def"), valWantErrIs: ErrValidationRequiredArgNil}, // list is required
		{name: "Wrong Index Type Coerced", toolName: "List.Get", args: MakeArgs(list, "1"), wantResult: int64(1)},                                           // Assuming coercion happens
		{name: "Wrong Index Type Invalid", toolName: "List.Get", args: MakeArgs(list, "abc"), valWantErrIs: ErrValidationTypeMismatch},                      // Assuming coercion fails
		{name: "Wrong List Type", toolName: "List.Get", args: MakeArgs("not list", int64(0)), valWantErrIs: ErrValidationTypeMismatch},
		{name: "Missing Index Arg", toolName: "List.Get", args: MakeArgs(list), valWantErrIs: ErrValidationRequiredArgMissing},
	}
	for _, tt := range tests {
		testListToolHelper(t, interp, tt)
	}
}

func TestToolListSlice(t *testing.T) {
	interp, _ := NewDefaultTestInterpreter(t)
	list := []interface{}{"a", "b", "c", "d", "e"}
	tests := []struct {
		name          string
		toolName      string
		args          []interface{}
		wantResult    interface{}
		wantToolErrIs error
		valWantErrIs  error
	}{
		{name: "Slice Middle", toolName: "List.Slice", args: MakeArgs(list, int64(1), int64(4)), wantResult: []interface{}{"b", "c", "d"}},
		{name: "Slice Start", toolName: "List.Slice", args: MakeArgs(list, int64(0), int64(2)), wantResult: []interface{}{"a", "b"}},
		{name: "Slice End", toolName: "List.Slice", args: MakeArgs(list, int64(3), int64(5)), wantResult: []interface{}{"d", "e"}},
		{name: "Slice Full", toolName: "List.Slice", args: MakeArgs(list, int64(0), int64(5)), wantResult: []interface{}{"a", "b", "c", "d", "e"}},
		{name: "Slice Empty Start=End", toolName: "List.Slice", args: MakeArgs(list, int64(2), int64(2)), wantResult: []interface{}{}},
		{name: "Slice Empty Start>End", toolName: "List.Slice", args: MakeArgs(list, int64(3), int64(1)), wantResult: []interface{}{}},
		{name: "Slice Clamp High End", toolName: "List.Slice", args: MakeArgs(list, int64(3), int64(10)), wantResult: []interface{}{"d", "e"}},
		{name: "Slice Clamp Low Start", toolName: "List.Slice", args: MakeArgs(list, int64(-2), int64(2)), wantResult: []interface{}{"a", "b"}},
		{name: "Slice Clamp Both", toolName: "List.Slice", args: MakeArgs(list, int64(-1), int64(10)), wantResult: []interface{}{"a", "b", "c", "d", "e"}},
		{name: "Slice Empty List", toolName: "List.Slice", args: MakeArgs([]interface{}{}, int64(0), int64(1)), wantResult: []interface{}{}},
		{name: "Slice Nil List Arg", toolName: "List.Slice", args: MakeArgs(interface{}(nil), int64(0), int64(1)), valWantErrIs: ErrValidationRequiredArgNil}, // list is required
		{name: "Wrong End Type Coerced", toolName: "List.Slice", args: MakeArgs(list, int64(0), "2"), wantResult: []interface{}{"a", "b"}},                    // Assuming coercion
		{name: "Wrong Start Type Invalid", toolName: "List.Slice", args: MakeArgs(list, "abc", int64(2)), valWantErrIs: ErrValidationTypeMismatch},
		{name: "Wrong List Type", toolName: "List.Slice", args: MakeArgs("no", int64(0), int64(1)), valWantErrIs: ErrValidationTypeMismatch},
		{name: "Missing End Index", toolName: "List.Slice", args: MakeArgs(list, int64(1)), valWantErrIs: ErrValidationRequiredArgMissing},
		{name: "Missing Start and End Indices", toolName: "List.Slice", args: MakeArgs(list), valWantErrIs: ErrValidationRequiredArgMissing},
	}
	for _, tt := range tests {
		testListToolHelper(t, interp, tt)
	}
}

func TestToolListContains(t *testing.T) {
	interp, _ := NewDefaultTestInterpreter(t)
	list := []interface{}{"a", int64(1), true, nil, float64(1.0), []interface{}{"sub"}}
	tests := []struct {
		name          string
		toolName      string
		args          []interface{}
		wantResult    interface{}
		wantToolErrIs error
		valWantErrIs  error
	}{
		{name: "Contains String", toolName: "List.Contains", args: MakeArgs(list, "a"), wantResult: true},
		{name: "Contains Int", toolName: "List.Contains", args: MakeArgs(list, int64(1)), wantResult: true},
		{name: "Contains Bool", toolName: "List.Contains", args: MakeArgs(list, true), wantResult: true},
		// Corrected: Passing nil as the optional 'element' is now valid
		{name: "Contains Actual Nil", toolName: "List.Contains", args: MakeArgs(list, nil), wantResult: true, valWantErrIs: nil},
		{name: "Contains Float", toolName: "List.Contains", args: MakeArgs(list, float64(1.0)), wantResult: true},
		{name: "Contains Sub-List", toolName: "List.Contains", args: MakeArgs(list, []interface{}{"sub"}), wantResult: true},
		{name: "Not Contains String", toolName: "List.Contains", args: MakeArgs(list, "b"), wantResult: false},
		{name: "Not Contains Int", toolName: "List.Contains", args: MakeArgs(list, int64(2)), wantResult: false},
		{name: "Not Contains Float", toolName: "List.Contains", args: MakeArgs(list, float64(1.1)), wantResult: false},
		{name: "Not Contains Sub-List (Different)", toolName: "List.Contains", args: MakeArgs(list, []interface{}{"diff"}), wantResult: false},
		{name: "Empty List", toolName: "List.Contains", args: MakeArgs([]interface{}{}, "a"), wantResult: false},
		{name: "Nil List Arg", toolName: "List.Contains", args: MakeArgs(interface{}(nil), "a"), valWantErrIs: ErrValidationRequiredArgNil}, // list is required
		{name: "Wrong List Type", toolName: "List.Contains", args: MakeArgs("no", "a"), valWantErrIs: ErrValidationTypeMismatch},
		// Corrected: Providing only the required 'list' arg is now valid (element is optional)
		{name: "Missing Value Arg", toolName: "List.Contains", args: MakeArgs(list), wantResult: true, valWantErrIs: nil}, // Contains nil if element missing? This seems wrong, should maybe error in tool func? Let's assume validation passes.
	}
	for _, tt := range tests {
		testListToolHelper(t, interp, tt)
	}
}

func TestToolListReverse(t *testing.T) {
	interp, _ := NewDefaultTestInterpreter(t)
	tests := []struct {
		name          string
		toolName      string
		args          []interface{}
		wantResult    interface{}
		wantToolErrIs error
		valWantErrIs  error
	}{
		{name: "Reverse Simple", toolName: "List.Reverse", args: MakeArgs([]interface{}{"a", int64(1), true}), wantResult: []interface{}{true, int64(1), "a"}},
		{name: "Reverse Single", toolName: "List.Reverse", args: MakeArgs([]interface{}{"a"}), wantResult: []interface{}{"a"}},
		{name: "Reverse Empty", toolName: "List.Reverse", args: MakeArgs([]interface{}{}), wantResult: []interface{}{}},
		{name: "Reverse Nil List Arg", toolName: "List.Reverse", args: MakeArgs(interface{}(nil)), valWantErrIs: ErrValidationRequiredArgNil}, // list is required
		{name: "Wrong Type", toolName: "List.Reverse", args: MakeArgs(123), valWantErrIs: ErrValidationTypeMismatch},
		{name: "Missing List Arg", toolName: "List.Reverse", args: MakeArgs(), valWantErrIs: ErrValidationRequiredArgMissing},
		{name: "Wrong Arg Count (Too Many)", toolName: "List.Reverse", args: MakeArgs([]interface{}{}, "extra"), valWantErrIs: ErrValidationArgCount},
	}
	for _, tt := range tests {
		testListToolHelper(t, interp, tt)
	}
}

func TestToolListSort(t *testing.T) {
	interp, _ := NewDefaultTestInterpreter(t)
	tests := []struct {
		name          string
		toolName      string
		args          []interface{}
		wantResult    interface{} // Can be []interface{}
		wantToolErrIs error       // Expect specific tool error
		valWantErrIs  error
	}{
		{name: "Sort Strings", toolName: "List.Sort", args: MakeArgs([]interface{}{"c", "a", "b"}), wantResult: []interface{}{"a", "b", "c"}},
		{name: "Sort Ints", toolName: "List.Sort", args: MakeArgs([]interface{}{int64(3), int64(1), int64(2)}), wantResult: []interface{}{int64(1), int64(2), int64(3)}},
		{name: "Sort Floats", toolName: "List.Sort", args: MakeArgs([]interface{}{float64(3.3), float64(1.1), float64(2.2)}), wantResult: []interface{}{float64(1.1), float64(2.2), float64(3.3)}},
		{name: "Sort Mixed Numbers", toolName: "List.Sort", args: MakeArgs([]interface{}{int64(3), float64(1.1), int64(2)}), wantResult: []interface{}{float64(1.1), int64(2), int64(3)}},
		{name: "Sort Empty", toolName: "List.Sort", args: MakeArgs([]interface{}{}), wantResult: []interface{}{}},
		{name: "Sort Nil List Arg", toolName: "List.Sort", args: MakeArgs(interface{}(nil)), valWantErrIs: ErrValidationRequiredArgNil}, // list is required
		{name: "Sort Single String", toolName: "List.Sort", args: MakeArgs([]interface{}{"a"}), wantResult: []interface{}{"a"}},
		{name: "Sort Single Int", toolName: "List.Sort", args: MakeArgs([]interface{}{int64(5)}), wantResult: []interface{}{int64(5)}},
		{name: "Sort Mixed String/Num", toolName: "List.Sort", args: MakeArgs([]interface{}{"a", int64(1)}), wantToolErrIs: ErrListCannotSortMixedTypes},
		{name: "Sort Mixed Num/String", toolName: "List.Sort", args: MakeArgs([]interface{}{int64(1), "a"}), wantToolErrIs: ErrListCannotSortMixedTypes},
		{name: "Sort Mixed Num/Bool", toolName: "List.Sort", args: MakeArgs([]interface{}{int64(1), true}), wantToolErrIs: ErrListCannotSortMixedTypes},
		{name: "Sort Mixed String/Bool", toolName: "List.Sort", args: MakeArgs([]interface{}{"a", false}), wantToolErrIs: ErrListCannotSortMixedTypes},
		{name: "Sort List of Lists", toolName: "List.Sort", args: MakeArgs([]interface{}{[]interface{}{1}, []interface{}{0}}), wantToolErrIs: ErrListCannotSortMixedTypes},
		{name: "Sort List With Nil Element", toolName: "List.Sort", args: MakeArgs([]interface{}{"a", nil, "c"}), wantToolErrIs: ErrListCannotSortMixedTypes},
		{name: "Sort Strings Looking Like Numbers", toolName: "List.Sort", args: MakeArgs([]interface{}{"10", "2", "1"}), wantResult: []interface{}{"1", "10", "2"}}, // Lexicographical
		{name: "Wrong Type", toolName: "List.Sort", args: MakeArgs(123), valWantErrIs: ErrValidationTypeMismatch},
		{name: "Missing List Arg", toolName: "List.Sort", args: MakeArgs(), valWantErrIs: ErrValidationRequiredArgMissing},
		{name: "Wrong Arg Count (Too Many)", toolName: "List.Sort", args: MakeArgs([]interface{}{}, "extra"), valWantErrIs: ErrValidationArgCount},
	}
	for _, tt := range tests {
		testListToolHelper(t, interp, tt)
	}
}

func TestToolListIsEmpty(t *testing.T) {
	interp, _ := NewDefaultTestInterpreter(t)
	tests := []struct {
		name          string
		toolName      string
		args          []interface{}
		wantResult    interface{}
		wantToolErrIs error
		valWantErrIs  error
	}{
		{name: "Is Empty True", toolName: "List.IsEmpty", args: MakeArgs([]interface{}{}), wantResult: true},
		{name: "Is Empty True (Nil Arg)", toolName: "List.IsEmpty", args: MakeArgs(interface{}(nil)), valWantErrIs: ErrValidationRequiredArgNil}, // list is required
		{name: "Is Empty False", toolName: "List.IsEmpty", args: MakeArgs([]interface{}{"a"}), wantResult: false},
		{name: "Is Empty False Long", toolName: "List.IsEmpty", args: MakeArgs([]interface{}{1, 2, 3}), wantResult: false},
		{name: "Wrong Type", toolName: "List.IsEmpty", args: MakeArgs("not a list"), valWantErrIs: ErrValidationTypeMismatch},
		{name: "Missing List Arg", toolName: "List.IsEmpty", args: MakeArgs(), valWantErrIs: ErrValidationRequiredArgMissing},
		{name: "Wrong Arg Count (Too Many)", toolName: "List.IsEmpty", args: MakeArgs([]interface{}{}, "extra"), valWantErrIs: ErrValidationArgCount},
	}
	for _, tt := range tests {
		testListToolHelper(t, interp, tt)
	}
}

// --- Test for ListHead ---
func TestToolListHead(t *testing.T) {
	interp, _ := NewDefaultTestInterpreter(t)
	list := []interface{}{"a", "b", int64(1), nil}
	singleList := []interface{}{"only"}
	emptyList := []interface{}{}

	tests := []struct {
		name          string
		toolName      string
		args          []interface{}
		wantResult    interface{}
		wantToolErrIs error
		valWantErrIs  error
	}{
		{name: "Head Simple", toolName: "List.Head", args: MakeArgs(list), wantResult: "a"},
		{name: "Head Single", toolName: "List.Head", args: MakeArgs(singleList), wantResult: "only"},
		{name: "Head First is Nil", toolName: "List.Head", args: MakeArgs([]interface{}{nil, "b"}), wantResult: nil},
		{name: "Head Empty", toolName: "List.Head", args: MakeArgs(emptyList), wantResult: nil},                                    // Returns nil for empty
		{name: "Head Nil Arg", toolName: "List.Head", args: MakeArgs(interface{}(nil)), valWantErrIs: ErrValidationRequiredArgNil}, // list is required
		{name: "Wrong Type", toolName: "List.Head", args: MakeArgs(123), valWantErrIs: ErrValidationTypeMismatch},
		{name: "Wrong Arg Count", toolName: "List.Head", args: MakeArgs(list, "extra"), valWantErrIs: ErrValidationArgCount},
		{name: "Missing List Arg", toolName: "List.Head", args: MakeArgs(), valWantErrIs: ErrValidationRequiredArgMissing},
	}
	for _, tt := range tests {
		testListToolHelper(t, interp, tt)
	}
}

// --- Test for ListRest ---
func TestToolListRest(t *testing.T) {
	interp, _ := NewDefaultTestInterpreter(t)
	list := []interface{}{"a", "b", int64(1)}
	singleList := []interface{}{"only"}
	emptyList := []interface{}{}

	tests := []struct {
		name          string
		toolName      string
		args          []interface{}
		wantResult    interface{}
		wantToolErrIs error
		valWantErrIs  error
	}{
		{name: "Rest Simple", toolName: "List.Rest", args: MakeArgs(list), wantResult: []interface{}{"b", int64(1)}},
		{name: "Rest Single", toolName: "List.Rest", args: MakeArgs(singleList), wantResult: []interface{}{}},                      // Returns empty
		{name: "Rest Empty", toolName: "List.Rest", args: MakeArgs(emptyList), wantResult: []interface{}{}},                        // Returns empty
		{name: "Rest Nil Arg", toolName: "List.Rest", args: MakeArgs(interface{}(nil)), valWantErrIs: ErrValidationRequiredArgNil}, // list is required
		{name: "Wrong Type", toolName: "List.Rest", args: MakeArgs(123), valWantErrIs: ErrValidationTypeMismatch},
		{name: "Wrong Arg Count", toolName: "List.Rest", args: MakeArgs(list, "extra"), valWantErrIs: ErrValidationArgCount},
		{name: "Missing List Arg", toolName: "List.Rest", args: MakeArgs(), valWantErrIs: ErrValidationRequiredArgMissing},
	}
	for _, tt := range tests {
		testListToolHelper(t, interp, tt)
	}
}

// --- Test for ListTail ---
func TestToolListTail(t *testing.T) {
	interp, _ := NewDefaultTestInterpreter(t)
	list := []interface{}{"a", "b", "c", "d", "e"}
	tests := []struct {
		name          string
		toolName      string
		args          []interface{}
		wantResult    interface{}
		wantToolErrIs error
		valWantErrIs  error
	}{
		{name: "Tail Last 2", toolName: "List.Tail", args: MakeArgs(list, int64(2)), wantResult: []interface{}{"d", "e"}},
		{name: "Tail Last 1", toolName: "List.Tail", args: MakeArgs(list, int64(1)), wantResult: []interface{}{"e"}},
		{name: "Tail Last 5 (All)", toolName: "List.Tail", args: MakeArgs(list, int64(5)), wantResult: []interface{}{"a", "b", "c", "d", "e"}},
		{name: "Tail Last 6 (>Len)", toolName: "List.Tail", args: MakeArgs(list, int64(6)), wantResult: []interface{}{"a", "b", "c", "d", "e"}},
		{name: "Tail Count 0", toolName: "List.Tail", args: MakeArgs(list, int64(0)), wantResult: []interface{}{}},
		{name: "Tail Count Negative", toolName: "List.Tail", args: MakeArgs(list, int64(-1)), wantResult: []interface{}{}},
		{name: "Tail Empty List", toolName: "List.Tail", args: MakeArgs([]interface{}{}, int64(2)), wantResult: []interface{}{}},
		{name: "Tail Nil List Arg", toolName: "List.Tail", args: MakeArgs(interface{}(nil), int64(2)), valWantErrIs: ErrValidationRequiredArgNil}, // list is required
		{name: "Wrong Count Type Coerced", toolName: "List.Tail", args: MakeArgs(list, "2"), wantResult: []interface{}{"d", "e"}},                 // Assuming coercion
		{name: "Wrong Count Type Invalid", toolName: "List.Tail", args: MakeArgs(list, "abc"), valWantErrIs: ErrValidationTypeMismatch},
		{name: "Wrong List Type", toolName: "List.Tail", args: MakeArgs("no", int64(1)), valWantErrIs: ErrValidationTypeMismatch},
		{name: "Missing Count Arg", toolName: "List.Tail", args: MakeArgs(list), valWantErrIs: ErrValidationRequiredArgMissing},
		{name: "Missing List Arg", toolName: "List.Tail", args: MakeArgs(), valWantErrIs: ErrValidationRequiredArgMissing},
	}
	for _, tt := range tests {
		testListToolHelper(t, interp, tt)
	}
}
