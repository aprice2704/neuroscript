// NeuroScript Version: 0.4.1
// File version: 7
// Purpose: Corrected test tool spec and sanitized test names to fix parser errors.
// filename: core/evaluation_new_types_test.go
// nlines: 140
// risk_rating: LOW

package core

import (
	"fmt"
	"math"
	"strings"
	"testing"
)

// runNewTypesTestScript is a helper to set up an interpreter and run a script.
func runNewTypesTestScript(t *testing.T, script string) (interface{}, error) {
	t.Helper()
	// Use the correct helper from helpers.go
	i, _ := NewTestInterpreter(t, nil, nil)

	// A temporary tool just for testing fuzzy value creation
	// CORRECTED: Used ArgTypeFloat instead of the string "number".
	specFuzzyTest := ToolSpec{Name: "Test.NewFuzzy", Args: []ArgSpec{{Name: "val", Type: ArgTypeFloat}}}
	toolFuzzyTest := func(_ *Interpreter, args []interface{}) (interface{}, error) {
		val, _ := toFloat64(args[0])
		return NewFuzzyValue(val), nil
	}
	// Manually register this test-only tool
	_ = i.RegisterTool(ToolImplementation{Spec: specFuzzyTest, Func: toolFuzzyTest})

	// Sanitize the test name to prevent the parser from misinterpreting special characters.
	// CORRECTED: Added replacement for '-'
	scriptNameForParser := strings.ReplaceAll(t.Name(), "/", "_")
	scriptNameForParser = strings.ReplaceAll(scriptNameForParser, "-", "_")

	// Execute the script using the correct function name and sanitized script name
	result, err := i.ExecuteScriptString(scriptNameForParser, script, nil)
	if err != nil {
		return nil, fmt.Errorf("script execution failed: %w", err)
	}
	return result, nil
}

func TestNewTypesIntegration(t *testing.T) {
	t.Run("TypeOf_New_Types", func(t *testing.T) {
		script := `
			set t = tool.Time.Now()
			set e = tool.Error.New(404, "not found")
			return typeof(t), typeof(e)
		`
		result, err := runNewTypesTestScript(t, script)
		if err != nil {
			t.Fatal(err)
		}

		resSlice, ok := result.([]interface{})
		if !ok || len(resSlice) != 2 {
			t.Fatalf("Expected a slice of 2 results, got %T", result)
		}

		if resSlice[0] != "timedate" {
			t.Errorf("typeof(t) failed: got %v, want 'timedate'", resSlice[0])
		}
		if resSlice[1] != "error" {
			t.Errorf("typeof(e) failed: got %v, want 'error'", resSlice[1])
		}
	})

	t.Run("Timedate_Comparison", func(t *testing.T) {
		script := `
			set t1 = tool.Time.Now()
			set t2 = tool.Time.Now()
			return t1 < t2, t1 <= t2
		`
		result, err := runNewTypesTestScript(t, script)
		if err != nil {
			t.Fatal(err)
		}

		resSlice, ok := result.([]interface{})
		if !ok || len(resSlice) != 2 {
			t.Fatalf("Expected a slice of 2 results, got %T", result)
		}
		if resSlice[0] != true {
			t.Errorf("t1 < t2 failed: got %v, want true", resSlice[0])
		}
		if resSlice[1] != true {
			t.Errorf("t1 <= t2 failed: got %v, want true", resSlice[1])
		}
	})

	t.Run("Fuzzy_Logic_Operators", func(t *testing.T) {
		script := `
			set f_true = tool.Test.NewFuzzy(0.8)
			set f_false = tool.Test.NewFuzzy(0.3)
			
			set res_not = not f_true      // 0.2
			set res_and = f_true and f_false  // min(0.8, 0.3) = 0.3
			set res_or = f_true or f_false    // max(0.8, 0.3) = 0.8
			set res_mixed = f_false or true // max(0.3, 1.0) = 1.0

			return res_not, res_and, res_or, res_mixed
		`
		result, err := runNewTypesTestScript(t, script)
		if err != nil {
			t.Fatal(err)
		}

		resSlice, ok := result.([]interface{})
		if !ok || len(resSlice) != 4 {
			t.Fatalf("Expected a slice of 4 results, got %T", result)
		}

		checkFuzzy := func(val interface{}, expected float64, name string) {
			fv, ok := val.(FuzzyValue)
			if !ok {
				t.Errorf("Expected FuzzyValue for %s, got %T", name, val)
				return
			}
			if math.Abs(fv.μ-expected) > 1e-9 {
				t.Errorf("%s: got fuzzy %v, want %v", name, fv.μ, expected)
			}
		}

		checkFuzzy(resSlice[0], 0.2, "res_not")
		checkFuzzy(resSlice[1], 0.3, "res_and")
		checkFuzzy(resSlice[2], 0.8, "res_or")
		checkFuzzy(resSlice[3], 1.0, "res_mixed")
	})

	t.Run("Error_Tool_and_is-error_built-in", func(t *testing.T) {
		script := `
			set my_err = tool.Error.New("E_FAIL", "it failed")
			set check_true = is_error(my_err)
			set check_false = is_error("just a string")
			return check_true, check_false, my_err
		`
		result, err := runNewTypesTestScript(t, script)
		if err != nil {
			t.Fatal(err)
		}

		resSlice, ok := result.([]interface{})
		if !ok || len(resSlice) != 3 {
			t.Fatalf("Expected a slice of 3 results, got %T", result)
		}

		if resSlice[0] != true {
			t.Errorf("is_error(my_err) failed: got %v, want true", resSlice[0])
		}
		if resSlice[1] != false {
			t.Errorf("is_error(\"a string\") failed: got %v, want false", resSlice[1])
		}
		errVal, ok := resSlice[2].(ErrorValue)
		if !ok {
			t.Fatalf("Expected third result to be ErrorValue, got %T", resSlice[2])
		}
		if code, _ := errVal.Value["code"].(StringValue); code.Value != "E_FAIL" {
			t.Errorf("Error code mismatch: got %v, want 'E_FAIL'", code.Value)
		}
	})
}
