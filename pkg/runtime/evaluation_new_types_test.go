// NeuroScript Version: 0.4.1
// File version: 12
// Purpose: Corrected runNewTypesTestScript to use the direct return value from the interpreter, which properly handles 'return' statements. This fixes the "variable not found" and incorrect result errors.
// filename: core/evaluation_new_types_test.go
// nlines: 125
// risk_rating: LOW

package runtime

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/parser"
)

// runNewTypesTestScript is a helper to set up an interpreter and run a script.
// It now correctly returns the interpreter's final result by capturing it from ExecuteScriptString.
func runNewTypesTestScript(t *testing.T, script string) (lang.Value, error) {
	t.Helper()
	i, _ := llm.NewTestInterpreter(t, nil, nil)

	specFuzzyTest := tool.ToolSpec{Name: "Test.NewFuzzy", Args: []tool.ArgSpec{{Name: "val", Type: parser.ArgTypeFloat}}}
	toolFuzzyTest := func(_ *Interpreter, args []interface{}) (interface{}, error) {
		val, _ := lang.toFloat64(args[0])
		return lang.NewFuzzyValue(val), nil
	}
	_ = i.RegisterTool(tool.ToolImplementation{Spec: specFuzzyTest, Func: toolFuzzyTest})

	scriptNameForParser := strings.ReplaceAll(t.Name(), "/", "_")
	scriptNameForParser = strings.ReplaceAll(scriptNameForParser, "-", "_")

	// CORRECTED: Use the direct result from ExecuteScriptString.
	// It correctly handles the 'return' statement and provides its result. The previous
	// approach of using i.lastCallResult was flawed because it didn't correctly
	// capture the value from a `return` statement, leading to errors.
	result, err := i.ExecuteScriptString(scriptNameForParser, script, nil)
	if err != nil {
		return nil, fmt.Errorf("script execution failed: %w", err)
	}

	// The interpreter can return primitives or Value types. Wrap ensures
	// we always have a Value, per the function signature and value contract.
	wrappedResult, wrapErr := lang.Wrap(result)
	if wrapErr != nil {
		return nil, fmt.Errorf("failed to wrap interpreter result: %w", wrapErr)
	}

	return wrappedResult, nil
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

		unwrapped := lang.Unwrap(result)
		resSlice, ok := unwrapped.([]interface{})
		if !ok || len(resSlice) != 2 {
			t.Fatalf("Expected a slice of 2 results, got %v (%T)", unwrapped, unwrapped)
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
			// A tiny sleep is needed on fast machines to ensure Now() is different
			call tool.Time.Sleep(1) 
			set t2 = tool.Time.Now()
			return t1 < t2, t1 <= t2
		`
		result, err := runNewTypesTestScript(t, script)
		if err != nil {
			t.Fatal(err)
		}

		unwrapped := lang.Unwrap(result)
		resSlice, ok := unwrapped.([]interface{})
		if !ok || len(resSlice) != 2 {
			t.Fatalf("Expected a slice of 2 results, got %v (%T)", unwrapped, unwrapped)
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
			
			set res_not = not f_true
			set res_and = f_true and f_false
			set res_or = f_true or f_false
			set res_mixed = f_false or true

			return res_not, res_and, res_or, res_mixed
		`
		result, err := runNewTypesTestScript(t, script)
		if err != nil {
			t.Fatal(err)
		}

		unwrapped := lang.Unwrap(result)
		resSlice, ok := unwrapped.([]interface{})
		if !ok || len(resSlice) != 4 {
			t.Fatalf("Expected a slice of 4 results, got %v (%T)", unwrapped, unwrapped)
		}

		checkFuzzy := func(val interface{}, expected float64, name string) {
			// Note: unwrapValue turns FuzzyValue into its float64 representation
			fv, ok := val.(float64)
			if !ok {
				t.Errorf("Expected float64 for %s, got %T", name, val)
				return
			}
			if math.Abs(fv-expected) > 1e-9 {
				t.Errorf("%s: got fuzzy %v, want %v", name, fv, expected)
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

		unwrapped := lang.Unwrap(result)
		resSlice, ok := unwrapped.([]interface{})
		if !ok || len(resSlice) != 3 {
			t.Fatalf("Expected a slice of 3 results, got %v (%T)", unwrapped, unwrapped)
		}

		if resSlice[0] != true {
			t.Errorf("is_error(my_err) failed: got %v, want true", resSlice[0])
		}
		if resSlice[1] != false {
			t.Errorf("is_error(\"a string\") failed: got %v, want false", resSlice[1])
		}
		errMap, ok := resSlice[2].(map[string]interface{})
		if !ok {
			t.Fatalf("Expected third result to be an unwrapped error map, got %T", resSlice[2])
		}
		if code, _ := errMap["code"].(string); code != "E_FAIL" {
			t.Errorf("Error code mismatch: got %v, want 'E_FAIL'", code)
		}
	})
}
