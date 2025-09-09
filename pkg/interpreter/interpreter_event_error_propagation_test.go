// NeuroScript Version: 0.7.1
// File version: 13
// Purpose: Removed obsolete tests for the now-deleted system:handler:error event, leaving only the test for the canonical host callback mechanism.
// filename: pkg/interpreter/interpreter_event_error_propagation_test.go
// nlines: 75
// risk_rating: LOW

package interpreter

import (
	"sync"
	"testing"
	"time"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/parser"
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
		t.Log("[DEBUG] Running Host callback test...")
		var (
			callbackErr     *lang.RuntimeError
			callbackEvent   string
			callbackSource  string
			callbackInvoked = make(chan struct{})
			mu              sync.Mutex
		)

		interp, err := NewTestInterpreter(t, nil, nil, false)
		if err != nil {
			t.Fatalf("Failed to create test interpreter: %v", err)
		}

		interp.SetEmitFunc(func(v lang.Value) {})
		interp.SetWhisperFunc(func(h, d lang.Value) {})

		interp.SetEventHandlerErrorCallback(func(eventName, source string, err *lang.RuntimeError) {
			mu.Lock()
			defer mu.Unlock()
			t.Logf("[DEBUG] Host callback invoked! event: %s, source: %s, err: %v", eventName, source, err)
			callbackErr = err
			callbackEvent = eventName
			callbackSource = source
			close(callbackInvoked)
		})

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

		interp.EmitEvent("user:login", "TestSystem", nil)

		select {
		case <-callbackInvoked:
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
		case <-time.After(500 * time.Millisecond): // Increased timeout for debugging
			t.Fatal("Test timed out: event handler error callback was not invoked.")
		}
	})
}
