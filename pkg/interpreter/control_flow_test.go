// NeuroScript Version: 0.5.2
// File version: 17.0.0
// Purpose: Refactored to use the NewTestHarness and updated Run calls to align with the post-refactor API.
// filename: pkg/interpreter/control_flow_test.go
// nlines: 200
// risk_rating: MEDIUM

package interpreter_test

import (
	"errors"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// runControlFlowTest is a new helper for this file, using the TestHarness.
func runControlFlowTest(t *testing.T, script string) (lang.Value, error) {
	t.Helper()
	h := NewTestHarness(t)
	h.T.Logf("[DEBUG] Turn 1: Harness created for script:\n%s", script)

	tree, pErr := h.Parser.Parse(script)
	if pErr != nil {
		h.T.Logf("[DEBUG] Turn 2: Parser failed: %v", pErr)
		return nil, pErr
	}
	h.T.Logf("[DEBUG] Turn 2: Script parsed successfully.")

	program, _, bErr := h.ASTBuilder.Build(tree)
	if bErr != nil {
		h.T.Logf("[DEBUG] Turn 3: AST Builder failed: %v", bErr)
		return nil, bErr
	}
	h.T.Logf("[DEBUG] Turn 3: AST built successfully.")

	if err := h.Interpreter.Load(&interfaces.Tree{Root: program}); err != nil {
		h.T.Logf("[DEBUG] Turn 4: Load failed: %v", err)
		return nil, err
	}
	h.T.Logf("[DEBUG] Turn 4: Script loaded successfully.")

	// Run main and return its result directly.
	h.T.Logf("[DEBUG] Turn 5: Calling Run('main').")
	result, runErr := h.Interpreter.Run("main")
	h.T.Logf("[DEBUG] Turn 6: Run('main') completed. Result: %#v, Error: %v", result, runErr)
	return result, runErr
}

func TestErrorHandlingControlFlow(t *testing.T) {
	t.Run("on_error_catches_a_fail_statement", func(t *testing.T) {
		script := `
			func main(returns result) means
				set result = "not caught"
				on error do
					set result = "caught it"
					clear_error
				endon

				fail "this is a test failure"

				return result
			endfunc
		`
		val, err := runControlFlowTest(t, script)
		if err != nil {
			t.Fatalf("runControlFlowTest returned an unexpected error: %v", err)
		}

		resultStr, _ := lang.ToString(val)
		expected := "caught it"
		if resultStr != expected {
			t.Errorf("Expected result '%s', but got '%s'", expected, resultStr)
		}
	})

	t.Run("must_failure_is_caught_by_on_error", func(t *testing.T) {
		script := `
			func main(returns result) means
				set result = "unhandled"
				on error do
					set result = "must failed as expected"
					clear_error
				endon

				must 1 > 2

				return result
			endfunc
		`
		val, err := runControlFlowTest(t, script)
		if err != nil {
			t.Fatalf("runControlFlowTest returned an unexpected error: %v", err)
		}

		resultStr, _ := lang.ToString(val)
		expected := "must failed as expected"
		if resultStr != expected {
			t.Errorf("Expected result '%s', but got '%s'", expected, resultStr)
		}
	})

	t.Run("clear_error_prevents_error_propagation", func(t *testing.T) {
		script := `
			func will_fail_but_clear() means
				on error do
					clear_error
				endon
				fail "this should be cleared"
			endfunc

			func main(returns result) means
				set result = "not continued"
				call will_fail_but_clear()
				set result = "successfully continued"
				return result
			endfunc
		`
		val, err := runControlFlowTest(t, script)
		if err != nil {
			t.Fatalf("Expected script to succeed due to clear_error, but it failed: %v", err)
		}

		resultStr, _ := lang.ToString(val)
		expected := "successfully continued"
		if resultStr != expected {
			t.Errorf("Expected final result '%s', got '%s'", expected, resultStr)
		}
	})

	t.Run("error_propagates_if_not_cleared", func(t *testing.T) {
		script := `
			func just_fails() means
				fail "propagating error"
			endfunc

			func main(returns result) means
				set result = "not caught"
				on error do
					set result = "main caught it"
				endon

				call just_fails()

				return "should not reach here"
			endfunc
		`
		_, err := runControlFlowTest(t, script)
		if err == nil {
			t.Fatal("Script was expected to fail, but it succeeded.")
		}

		var rtErr *lang.RuntimeError
		if !errors.As(err, &rtErr) {
			t.Fatalf("Expected a RuntimeError, but got %T: %v", err, err)
		}

		if rtErr.Message != "propagating error" {
			t.Errorf("Expected propagated error message to be 'propagating error', got '%s'", rtErr.Message)
		}
	})

	t.Run("fail_with_expression", func(t *testing.T) {
		script := `
			func main(returns result) means
				set result = "not handled"
				on error do
					set result = "handled"
					clear_error
				endon
				set err_msg = "custom failure"
				fail err_msg
				return result
			endfunc
		`
		val, err := runControlFlowTest(t, script)
		if err != nil {
			t.Fatalf("Script returned an unexpected Go error: %v", err)
		}

		resultStr, _ := lang.ToString(val)
		expected := "handled"
		if resultStr != expected {
			t.Errorf("Expected result '%s', but got '%s'", expected, resultStr)
		}
	})

	t.Run("for_each_with_nil_collection", func(t *testing.T) {
		script := `
			func main(returns result) means
				set my_collection = nil
				set loop_did_run = false

				for each item in my_collection
					set loop_did_run = true
				endfor

				return loop_did_run
			endfunc
		`
		val, err := runControlFlowTest(t, script)
		if err != nil {
			t.Fatalf("runControlFlowTest returned an unexpected error: %v", err)
		}

		resultBool, ok := val.(lang.BoolValue)
		if !ok {
			t.Fatalf("Expected boolean result, but got %T", val)
		}
		if resultBool.Value {
			t.Error("Expected loop_did_run to be false, but it was true")
		}
	})
}
