// filename: pkg/core/evaluation_resolve_test.go
package core

import (
	"errors"
	"reflect"
	"testing"
	// Assuming universal_test_helpers provides runEvalTestCase (or similar) and EvalTestCase struct
	// Assuming errors like ErrVariableNotFound are defined in errors.go
	// Assuming AST nodes like StringLiteralNode are defined in ast.go
)

// --- Test Suite for resolveValue Placeholder Substitution ---

// ResolveValueTestCase mirrors EvalTestCase but focuses on resolveValue results
type ResolveValueTestCase struct {
	Name            string
	InputNode       interface{} // Should be the AST node passed to resolveValue
	InitialVars     map[string]interface{}
	ExpectedValue   interface{} // The expected value returned by resolveValue
	WantErr         bool
	ExpectedErrorIs error // Use sentinel error or nil
}

// runResolveValueTest is a helper adapted from runEvalExpressionTest, focusing on calling resolveValue.
// NOTE: Assumes NewDefaultTestInterpreter helper exists from universal_test_helpers.go
func runResolveValueTest(t *testing.T, tt ResolveValueTestCase) {
	t.Helper() // Mark this function as a test helper

	// Set up interpreter with initial variables
	interp, _ := NewDefaultTestInterpreter(t) // Assumes this helper is available
	if tt.InitialVars != nil {
		for k, v := range tt.InitialVars {
			// Use SetVariable to ensure variables are properly stored
			if err := interp.SetVariable(k, v); err != nil {
				t.Fatalf("[%s] Failed to set initial variable '%s': %v", tt.Name, k, err)
			}
		}
	}
	// Note: resolveValue doesn't use lastCallResult directly, so no need to set it here.

	// Call the target function: resolveValue
	gotValue, err := interp.resolveValue(tt.InputNode)

	// --- Assertions ---
	if tt.WantErr {
		if err == nil {
			t.Errorf("[%s] Expected an error, but got nil", tt.Name)
			return
		}
		if tt.ExpectedErrorIs != nil {
			if !errors.Is(err, tt.ExpectedErrorIs) {
				t.Errorf("[%s] Error mismatch.\nExpected error wrapping: [%v]\nGot:                     [%v]", tt.Name, tt.ExpectedErrorIs, err)
			} else {
				// Optional: Log that the expected error type was received
				t.Logf("[%s] Got expected error wrapping [%v]: %v", tt.Name, tt.ExpectedErrorIs, err)
			}
		} else {
			// If WantErr is true but ExpectedErrorIs is nil, just check that *an* error occurred.
			t.Logf("[%s] Got expected error (specific type not checked): %v", tt.Name, err)
		}
		// If an error was expected, don't compare the returned value (it might be meaningless)
	} else { // No error wanted
		if err != nil {
			t.Errorf("[%s] Unexpected error: %v", tt.Name, err)
		} else if !reflect.DeepEqual(gotValue, tt.ExpectedValue) {
			// Provide detailed output on value mismatch
			t.Errorf("[%s] Result value mismatch.\nInput Node: %+v\nExpected:   %v (%T)\nGot:        %v (%T)",
				tt.Name, tt.InputNode, tt.ExpectedValue, tt.ExpectedValue, gotValue, gotValue)
		}
	}
}

func TestResolveValuePlaceholders(t *testing.T) {
	vars := map[string]interface{}{
		"name":      "World",
		"greeting":  "Hello",
		"subject":   "there",
		"num":       int64(42),
		"boolVal":   true,
		"nilVal":    nil,
		"spacedVar": "Spaced Value",
	}

	testCases := []ResolveValueTestCase{
		// --- Success Cases ---
		{
			Name:          "Raw string basic substitution",
			InputNode:     StringLiteralNode{Value: "Test {{name}}!", IsRaw: true},
			InitialVars:   vars,
			ExpectedValue: "Test World!",
			WantErr:       false,
		},
		{
			Name:          "Raw string multiple substitutions",
			InputNode:     StringLiteralNode{Value: "{{greeting}} {{subject}} - Status {{boolVal}}", IsRaw: true},
			InitialVars:   vars,
			ExpectedValue: "Hello there - Status true",
			WantErr:       false,
		},
		{
			Name:          "Raw string numeric substitution",
			InputNode:     StringLiteralNode{Value: "The number is {{num}}.", IsRaw: true},
			InitialVars:   vars,
			ExpectedValue: "The number is 42.",
			WantErr:       false,
		},
		{
			Name:          "Raw string nil substitution",
			InputNode:     StringLiteralNode{Value: "Value: {{nilVal}}.", IsRaw: true},
			InitialVars:   vars,
			ExpectedValue: "Value: <nil>.", // fmt.Sprintf("%v", nil) results in "<nil>"
			WantErr:       false,
		},
		{
			Name:          "Raw string with surrounding whitespace in placeholder",
			InputNode:     StringLiteralNode{Value: "Data: {{ spacedVar }}", IsRaw: true},
			InitialVars:   vars,
			ExpectedValue: "Data: Spaced Value",
			WantErr:       false,
		},
		{
			Name:          "Raw string consecutive placeholders",
			InputNode:     StringLiteralNode{Value: "{{greeting}}{{subject}}", IsRaw: true},
			InitialVars:   vars,
			ExpectedValue: "Hellothere",
			WantErr:       false,
		},
		{
			Name:          "Raw string no placeholders",
			InputNode:     StringLiteralNode{Value: "Just plain text.", IsRaw: true},
			InitialVars:   vars,
			ExpectedValue: "Just plain text.",
			WantErr:       false,
		},
		{
			Name:          "Raw string empty",
			InputNode:     StringLiteralNode{Value: "", IsRaw: true},
			InitialVars:   vars,
			ExpectedValue: "",
			WantErr:       false,
		},
		{
			Name:          "Raw string only placeholder",
			InputNode:     StringLiteralNode{Value: "{{name}}", IsRaw: true},
			InitialVars:   vars,
			ExpectedValue: "World",
			WantErr:       false,
		},

		// --- Cases: No Substitution Expected ---
		{
			Name:          "Normal string with placeholder syntax",
			InputNode:     StringLiteralNode{Value: "Hello {{name}}!", IsRaw: false}, // IsRaw is false
			InitialVars:   vars,
			ExpectedValue: "Hello {{name}}!", // Expect original value
			WantErr:       false,
		},
		{
			Name:          "Normal string empty",
			InputNode:     StringLiteralNode{Value: "", IsRaw: false},
			InitialVars:   vars,
			ExpectedValue: "",
			WantErr:       false,
		},
		{
			Name:          "Normal string resembling raw syntax",
			InputNode:     StringLiteralNode{Value: "```{{name}}```", IsRaw: false}, // IsRaw is false
			InitialVars:   vars,
			ExpectedValue: "```{{name}}```", // Expect original value
			WantErr:       false,
		},

		// --- Error Cases ---
		{
			Name:            "Raw string missing variable",
			InputNode:       StringLiteralNode{Value: "Hello {{missing_var}}!", IsRaw: true},
			InitialVars:     vars,
			ExpectedValue:   nil, // Value is undefined on error
			WantErr:         true,
			ExpectedErrorIs: ErrVariableNotFound,
		},
		{
			Name:            "Raw string mixed found and missing",
			InputNode:       StringLiteralNode{Value: "{{greeting}} {{not_here}} {{name}}", IsRaw: true},
			InitialVars:     vars,
			ExpectedValue:   nil,
			WantErr:         true,
			ExpectedErrorIs: ErrVariableNotFound, // Should report the first missing variable error
		},
		{
			Name:          "Raw string malformed placeholder (no closing)",
			InputNode:     StringLiteralNode{Value: "Test {{name", IsRaw: true},
			InitialVars:   vars,
			ExpectedValue: "Test {{name", // Regex won't match, no substitution, no error
			WantErr:       false,
		},
		{
			Name:          "Raw string malformed placeholder (empty)",
			InputNode:     StringLiteralNode{Value: "Test {{}}", IsRaw: true},
			InitialVars:   vars,
			ExpectedValue: "Test {{}}", // Regex won't match invalid identifier, no substitution, no error
			WantErr:       false,
		},
		{
			Name:          "Raw string malformed placeholder (space only)",
			InputNode:     StringLiteralNode{Value: "Test {{ }}", IsRaw: true},
			InitialVars:   vars,
			ExpectedValue: "Test {{ }}", // Regex won't match identifier, no substitution, no error
			WantErr:       false,
		},
	}

	for _, tt := range testCases {
		// Run the test case using the helper
		t.Run(tt.Name, func(t *testing.T) {
			runResolveValueTest(t, tt)
		})
	}
}
