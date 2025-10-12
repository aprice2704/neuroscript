// NeuroScript Version: 0.7.1
// File version: 14
// Purpose: Refactored to use the centralized TestHarness for robust and consistent interpreter initialization.
// filename: pkg/interpreter/interpreter_event_error_propagation_test.go
// nlines: 80
// risk_rating: LOW

package interpreter_test

import (
	"sync"
	"testing"
	"time"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// TestEventHandler_ErrorPropagation now exclusively covers the host-callback error reporting system.
func TestEventHandler_ErrorPropagation(t *testing.T) {
	failingHandlerScript := `
on event "user:login" do
	emit "handler started"
	call tool.debug.dumpClones()
	call tool.marmot.painter()
	emit "handler finished"
endon
`
	t.Run("Host callback receives handler error", func(t *testing.T) {
		t.Logf("[DEBUG] Turn 1: Starting 'Host callback receives handler error' test.")
		h := NewTestHarness(t)
		var (
			callbackErr     *lang.RuntimeError
			callbackEvent   string
			callbackSource  string
			callbackInvoked = make(chan struct{})
			mu              sync.Mutex
		)

		h.HostContext.EventHandlerErrorCallback = func(eventName, source string, err *lang.RuntimeError) {
			mu.Lock()
			defer mu.Unlock()
			t.Logf("[DEBUG] Turn X: Host callback invoked! event: %s, source: %s, err: %v", eventName, source, err)
			callbackErr = err
			callbackEvent = eventName
			callbackSource = source
			close(callbackInvoked)
		}
		t.Logf("[DEBUG] Turn 2: HostContext callback configured.")

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
		t.Logf("[DEBUG] Turn 3: Script loaded.")

		h.Interpreter.EmitEvent("user:login", "TestSystem", nil)
		t.Logf("[DEBUG] Turn 4: Event emitted.")

		select {
		case <-callbackInvoked:
			t.Logf("[DEBUG] Turn 5: Callback invoked as expected.")
			mu.Lock()
			defer mu.Unlock()
			if callbackEvent != "user:login" {
				t.Errorf("Callback received wrong event name. Got: '%s', Want: 'user:login'", callbackEvent)
			}
			if callbackSource != "TestSystem" {
				t.Errorf("Callback received wrong source. Got: '%s', Want: 'TestSystem'", callbackSource)
			}
			if callbackErr == nil || callbackErr.Code != lang.ErrorCodeToolNotFound {
				t.Errorf("Callback received wrong error. Got: %v, Want: ToolNotFound", callbackErr)
			}
		case <-time.After(500 * time.Millisecond):
			t.Fatal("Test timed out: event handler error callback was not invoked.")
		}
		t.Logf("[DEBUG] Turn 6: Test completed.")
	})
}
