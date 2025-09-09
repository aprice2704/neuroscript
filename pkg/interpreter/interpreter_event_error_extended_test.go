// NeuroScript Version: 0.7.1
// File version: 2
// Purpose: Corrected scripts to be syntactically valid and added required I/O function setup to prevent panics, fixing all test failures.
// filename: pkg/interpreter/interpreter_event_error_extended_test.go
// nlines: 135
// risk_rating: LOW

package interpreter

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/parser"
)

// TestEventHandler_ExtendedErrors covers additional edge cases for the host callback mechanism.
func TestEventHandler_ExtendedErrors(t *testing.T) {

	// Test Case 1: No-Op on Nil Callback
	t.Run("Failure with nil callback does not panic", func(t *testing.T) {
		t.Log("[DEBUG] Running nil callback test...")
		// FIX: Use a syntactically valid but unregistered tool name.
		failingHandlerScript := `
		on event "user:action" do
			call tool.a.b()
		endon
		`
		// This test passes if it completes without panicking.
		interp, err := NewTestInterpreter(t, nil, nil, false)
		if err != nil {
			t.Fatalf("Failed to create test interpreter: %v", err)
		}

		// Intentionally do NOT set the error callback.
		// interp.SetEventHandlerErrorCallback(...)

		// FIX: We must set the I/O funcs if a handler exists, even if they do nothing.
		interp.SetEmitFunc(func(v lang.Value) {})
		interp.SetWhisperFunc(func(h, d lang.Value) {})

		tree, pErr := parser.NewParserAPI(nil).Parse(failingHandlerScript)
		if pErr != nil {
			t.Fatalf("Failed to parse script: %v", pErr)
		}
		program, _, bErr := parser.NewASTBuilder(nil).Build(tree)
		if bErr != nil {
			t.Fatalf("Failed to build AST: %v", bErr)
		}
		if err := interp.Load(&interfaces.Tree{Root: program}); err != nil {
			t.Fatalf("Failed to load program: %v", err)
		}

		// This should cause an error internally, but since there's no callback,
		// it should be handled gracefully without crashing.
		interp.EmitEvent("user:action", "TestSystem", nil)
		t.Log("[DEBUG] Test completed without panic.")
	})

	// Test Case 2: Handler Isolation
	t.Run("One failing handler does not stop others", func(t *testing.T) {
		t.Log("[DEBUG] Running handler isolation test...")
		// FIX: Use a syntactically valid but unregistered tool name.
		multiHandlerScript := `
		on event "user:action" do
			emit "handler_A_running"
			call tool.a.b()
			emit "handler_A_finished"
		endon

		on event "user:action" do
			emit "handler_B_running"
		endon
		`
		var (
			callbackInvoked bool
			emittedOutput   []string
			mu              sync.Mutex
		)

		interp, err := NewTestInterpreter(t, nil, nil, false)
		if err != nil {
			t.Fatalf("Failed to create test interpreter: %v", err)
		}

		interp.SetEventHandlerErrorCallback(func(eventName, source string, err *lang.RuntimeError) {
			mu.Lock()
			defer mu.Unlock()
			callbackInvoked = true
		})

		interp.SetEmitFunc(func(v lang.Value) {
			mu.Lock()
			defer mu.Unlock()
			emittedOutput = append(emittedOutput, v.String())
		})
		// FIX: Add whisper func for completeness.
		interp.SetWhisperFunc(func(h, d lang.Value) {})

		tree, pErr := parser.NewParserAPI(nil).Parse(multiHandlerScript)
		if pErr != nil {
			t.Fatalf("Failed to parse script: %v", pErr)
		}
		program, _, bErr := parser.NewASTBuilder(nil).Build(tree)
		if bErr != nil {
			t.Fatalf("Failed to build AST: %v", bErr)
		}
		if err := interp.Load(&interfaces.Tree{Root: program}); err != nil {
			t.Fatalf("Failed to load program: %v", err)
		}

		interp.EmitEvent("user:action", "TestSystem", nil)
		time.Sleep(100 * time.Millisecond) // Give async handlers time to run

		mu.Lock()
		defer mu.Unlock()

		if !callbackInvoked {
			t.Error("Error callback was not invoked for the failing handler.")
		}

		output := strings.Join(emittedOutput, "|")
		if !strings.Contains(output, "handler_A_running") {
			t.Error("Handler A did not start executing.")
		}
		if !strings.Contains(output, "handler_B_running") {
			t.Error("Handler B did not execute after Handler A failed.")
		}
		if strings.Contains(output, "handler_A_finished") {
			t.Error("Handler A should not have finished after the error.")
		}
	})

	// Test Case 3: Failure via `fail` statement
	t.Run("Failure via 'fail' statement triggers callback", func(t *testing.T) {
		t.Log("[DEBUG] Running 'fail' statement test...")
		failHandlerScript := `
		on event "user:action" do
			fail "intentional failure from handler"
		endon
		`
		var (
			callbackErr     *lang.RuntimeError
			callbackInvoked = make(chan struct{})
		)

		interp, err := NewTestInterpreter(t, nil, nil, false)
		if err != nil {
			t.Fatalf("Failed to create test interpreter: %v", err)
		}

		// FIX: Add required I/O funcs to prevent panic.
		interp.SetEmitFunc(func(v lang.Value) {})
		interp.SetWhisperFunc(func(h, d lang.Value) {})

		interp.SetEventHandlerErrorCallback(func(eventName, source string, err *lang.RuntimeError) {
			callbackErr = err
			close(callbackInvoked)
		})

		tree, pErr := parser.NewParserAPI(nil).Parse(failHandlerScript)
		if pErr != nil {
			t.Fatalf("Failed to parse script: %v", pErr)
		}
		program, _, bErr := parser.NewASTBuilder(nil).Build(tree)
		if bErr != nil {
			t.Fatalf("Failed to build AST: %v", bErr)
		}
		if err := interp.Load(&interfaces.Tree{Root: program}); err != nil {
			t.Fatalf("Failed to load program: %v", err)
		}

		interp.EmitEvent("user:action", "TestSystem", nil)

		select {
		case <-callbackInvoked:
			if callbackErr == nil {
				t.Fatal("Callback was invoked but error was nil.")
			}
			if callbackErr.Code != lang.ErrorCodeFailStatement {
				t.Errorf("Callback received wrong error code. Got: %v, Want: ErrorCodeFailStatement", callbackErr.Code)
			}
			if !strings.Contains(callbackErr.Message, "intentional failure") {
				t.Errorf("Callback received wrong error message. Got: '%s'", callbackErr.Message)
			}
		case <-time.After(500 * time.Millisecond):
			t.Fatal("Test timed out: error callback was not invoked for 'fail' statement.")
		}
	})
}
