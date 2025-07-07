// NeuroScript Version: 0.5.2
// File version: 19
// Purpose: Corrected typeof() check for error type to pass the test after fixing the NewErrorValue constructor.
// filename: pkg/interpreter/evaluation_new_types_test.go
// nlines: 160
// risk_rating: MEDIUM

package interpreter

import (
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// runNewTypesTestScript is a helper to set up an interpreter and run a script.
func runNewTypesTestScript(t *testing.T, script string) (lang.Value, error) {
	t.Helper()
	i := NewInterpreter(WithLogger(logging.NewTestLogger(t)))

	// Register test-specific tool for fuzzy logic
	specFuzzyTest := tool.ToolSpec{Name: "Test.NewFuzzy", Args: []tool.ArgSpec{{Name: "val", Type: "float"}}}
	toolFuzzyTest := func(_ tool.Runtime, args []interface{}) (interface{}, error) {
		val, _ := lang.ToFloat64(args[0])
		return lang.NewFuzzyValue(val), nil
	}
	_ = i.ToolRegistry().RegisterTool(tool.ToolImplementation{Spec: specFuzzyTest, Func: toolFuzzyTest})

	// Manually register the Time and Error tools that are no longer auto-registered.
	specTimeNow := tool.ToolSpec{Name: "Time.Now", Args: []tool.ArgSpec{}}
	toolTimeNow := func(_ tool.Runtime, args []interface{}) (interface{}, error) {
		return lang.TimedateValue{Value: time.Now()}, nil
	}
	_ = i.ToolRegistry().RegisterTool(tool.ToolImplementation{Spec: specTimeNow, Func: toolTimeNow})

	specTimeSleep := tool.ToolSpec{Name: "Time.Sleep", Args: []tool.ArgSpec{{Name: "ms", Type: "int"}}}
	toolTimeSleep := func(_ tool.Runtime, args []interface{}) (interface{}, error) {
		ms, _ := lang.ToInt64(args[0])
		time.Sleep(time.Duration(ms) * time.Millisecond)
		return &lang.NilValue{}, nil
	}
	_ = i.ToolRegistry().RegisterTool(tool.ToolImplementation{Spec: specTimeSleep, Func: toolTimeSleep})

	specErrorNew := tool.ToolSpec{Name: "Error.New", Args: []tool.ArgSpec{{Name: "code", Type: "any"}, {Name: "msg", Type: "string"}}}
	toolErrorNew := func(_ tool.Runtime, args []interface{}) (interface{}, error) {
		var codeStr string
		if codeVal, ok := args[0].(lang.Value); ok {
			codeStr, _ = lang.ToString(codeVal)
		} else {
			codeStr = fmt.Sprintf("%v", args[0])
		}
		msg, _ := lang.ToString(args[1])
		return lang.NewErrorValue(codeStr, msg, nil), nil
	}
	_ = i.ToolRegistry().RegisterTool(tool.ToolImplementation{Spec: specErrorNew, Func: toolErrorNew})

	scriptNameForParser := strings.ReplaceAll(t.Name(), "/", "_")
	scriptNameForParser = strings.ReplaceAll(scriptNameForParser, "-", "_")

	// The result from ExecuteScriptString is already a lang.Value
	result, rErr := i.ExecuteScriptString(scriptNameForParser, script, nil)
	if rErr != nil {
		return nil, fmt.Errorf("script execution failed: %w", rErr)
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
