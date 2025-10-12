// NeuroScript Version: 0.7.1
// File version: 3
// Purpose: Refactored to use the centralized TestHarness for robust and consistent interpreter initialization.
// filename: pkg/interpreter/interpreter_event_error_extended_test.go
// nlines: 140
// risk_rating: LOW

package interpreter_test

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// TestEventHandler_ExtendedErrors covers additional edge cases for the host callback mechanism.
func TestEventHandler_ExtendedErrors(t *testing.T) {

	t.Run("Failure with nil callback does not panic", func(t *testing.T) {
		t.Logf("[DEBUG] Turn 1: Starting 'Failure with nil callback does not panic' test.")
		h := NewTestHarness(t)
		failingHandlerScript := `
		on event "user:action" do
			call tool.a.b()
		endon
		`
		// Intentionally do NOT set the error callback on the HostContext.

		tree, pErr := h.Parser.Parse(failingHandlerScript)
		if pErr != nil {
			t.Fatalf("Failed to parse script: %v", pErr)
		}
		program, _, bErr := h.ASTBuilder.Build(tree)
		if bErr != nil {
			t.Fatalf("Failed to build AST: %v", bErr)
		}
		if err := h.Interpreter.Load(&interfaces.Tree{Root: program}); err != nil {
			t.Fatalf("Failed to load program: %v", err)
		}
		t.Logf("[DEBUG] Turn 2: Script loaded.")

		h.Interpreter.EmitEvent("user:action", "TestSystem", nil)
		t.Logf("[DEBUG] Turn 3: Event emitted. Test completed without panic.")
	})

	t.Run("One failing handler does not stop others", func(t *testing.T) {
		t.Logf("[DEBUG] Turn 1: Starting 'One failing handler does not stop others' test.")
		h := NewTestHarness(t)
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

		h.HostContext.EventHandlerErrorCallback = func(eventName, source string, err *lang.RuntimeError) {
			mu.Lock()
			defer mu.Unlock()
			t.Logf("[DEBUG] Turn X: EventHandlerErrorCallback invoked.")
			callbackInvoked = true
		}
		h.HostContext.EmitFunc = func(v lang.Value) {
			mu.Lock()
			defer mu.Unlock()
			t.Logf("[DEBUG] Turn X: EmitFunc captured: %s", v.String())
			emittedOutput = append(emittedOutput, v.String())
		}
		t.Logf("[DEBUG] Turn 2: Callbacks configured on HostContext.")

		tree, pErr := h.Parser.Parse(multiHandlerScript)
		if pErr != nil {
			t.Fatalf("Failed to parse script: %v", pErr)
		}
		program, _, bErr := h.ASTBuilder.Build(tree)
		if bErr != nil {
			t.Fatalf("Failed to build AST: %v", bErr)
		}
		if err := h.Interpreter.Load(&interfaces.Tree{Root: program}); err != nil {
			t.Fatalf("Failed to load program: %v", err)
		}
		t.Logf("[DEBUG] Turn 3: Script loaded.")

		h.Interpreter.EmitEvent("user:action", "TestSystem", nil)
		time.Sleep(100 * time.Millisecond)
		t.Logf("[DEBUG] Turn 4: Event emitted, waiting for async handlers.")

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
		t.Logf("[DEBUG] Turn 5: Assertions complete.")
	})

	t.Run("Failure via 'fail' statement triggers callback", func(t *testing.T) {
		t.Logf("[DEBUG] Turn 1: Starting 'Failure via 'fail' statement' test.")
		h := NewTestHarness(t)
		failHandlerScript := `
		on event "user:action" do
			fail "intentional failure from handler"
		endon
		`
		var (
			callbackErr     *lang.RuntimeError
			callbackInvoked = make(chan struct{})
		)

		h.HostContext.EventHandlerErrorCallback = func(eventName, source string, err *lang.RuntimeError) {
			t.Logf("[DEBUG] Turn X: EventHandlerErrorCallback invoked.")
			callbackErr = err
			close(callbackInvoked)
		}
		t.Logf("[DEBUG] Turn 2: Callback configured.")

		tree, pErr := h.Parser.Parse(failHandlerScript)
		if pErr != nil {
			t.Fatalf("Failed to parse script: %v", pErr)
		}
		program, _, bErr := h.ASTBuilder.Build(tree)
		if bErr != nil {
			t.Fatalf("Failed to build AST: %v", bErr)
		}
		if err := h.Interpreter.Load(&interfaces.Tree{Root: program}); err != nil {
			t.Fatalf("Failed to load program: %v", err)
		}
		t.Logf("[DEBUG] Turn 3: Script loaded.")

		h.Interpreter.EmitEvent("user:action", "TestSystem", nil)
		t.Logf("[DEBUG] Turn 4: Event emitted.")

		select {
		case <-callbackInvoked:
			t.Logf("[DEBUG] Turn 5: Callback was invoked as expected.")
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
		t.Logf("[DEBUG] Turn 6: Test completed.")
	})
}
