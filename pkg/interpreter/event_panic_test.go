// NeuroScript Version: 0.8.0
// File version: 1
// Purpose: Tests that panics occurring within event handlers are caught and reported via the error callback.
// filename: pkg/interpreter/event_panic_test.go
// nlines: 75
// risk_rating: LOW

package interpreter_test

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// panickingTool is a simple tool designed to cause a panic when called.
var panickingTool = tool.ToolImplementation{
	Spec: tool.ToolSpec{
		Name:  "Panic",
		Group: "test",
	},
	Func: func(rt tool.Runtime, args []interface{}) (interface{}, error) {
		panic("intentional panic from test tool")
	},
}

func TestEventHandler_PanicRecovery(t *testing.T) {
	t.Logf("[DEBUG] Turn 1: Starting TestEventHandler_PanicRecovery.")
	h := NewTestHarness(t)
	interp := h.Interpreter

	// Register the tool that will cause the panic.
	if _, err := interp.ToolRegistry().RegisterTool(panickingTool); err != nil {
		t.Fatalf("Failed to register panicking tool: %v", err)
	}
	t.Logf("[DEBUG] Turn 2: Panicking tool registered.")

	// Script with an event handler that calls the panicking tool.
	script := `
	on event "cause_panic" do
		emit "Handler started..." // Should execute
		call tool.test.Panic()
		emit "Handler finished." // Should NOT execute
	endon
	`

	var (
		callbackErr     *lang.RuntimeError
		callbackInvoked bool
		mu              sync.Mutex
		emitted         []string
	)

	// Configure the error callback to capture the panic error.
	h.HostContext.EventHandlerErrorCallback = func(eventName, source string, err *lang.RuntimeError) {
		mu.Lock()
		defer mu.Unlock()
		t.Logf("[DEBUG] Turn X: EventHandlerErrorCallback invoked! Error: %v", err)
		callbackErr = err
		callbackInvoked = true
	}
	h.HostContext.EmitFunc = func(v lang.Value) {
		mu.Lock()
		defer mu.Unlock()
		emitted = append(emitted, v.String())
	}
	t.Logf("[DEBUG] Turn 3: HostContext callbacks configured.")

	// Load the script.
	tree, pErr := h.Parser.Parse(script)
	if pErr != nil {
		t.Fatalf("Failed to parse script: %v", pErr)
	}
	program, _, bErr := h.ASTBuilder.Build(tree)
	if bErr != nil {
		t.Fatalf("Failed to build AST: %v", bErr)
	}
	if err := interp.Load(&interfaces.Tree{Root: program}); err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}
	t.Logf("[DEBUG] Turn 4: Script loaded.")

	// Emit the event to trigger the handler.
	interp.EmitEvent("cause_panic", "TestSystem", nil)
	t.Logf("[DEBUG] Turn 5: Event emitted, waiting for handler.")

	// Wait briefly for the handler goroutine and callback.
	time.Sleep(100 * time.Millisecond)

	// Assertions.
	mu.Lock()
	defer mu.Unlock()

	if !callbackInvoked {
		t.Fatal("EventHandlerErrorCallback was not invoked after the panic.")
	}
	if callbackErr == nil {
		t.Fatal("Callback was invoked, but the received error was nil.")
	}
	if callbackErr.Code != lang.ErrorCodeInternal {
		t.Errorf("Expected error code %d (Internal), but got %d", lang.ErrorCodeInternal, callbackErr.Code)
	}
	expectedMsgPart := "panic executing event handler"
	expectedPanicVal := "intentional panic from test tool"
	if !strings.Contains(callbackErr.Message, expectedMsgPart) || !strings.Contains(callbackErr.Error(), expectedPanicVal) {
		t.Errorf("Error message missing expected panic details. Got: %s", callbackErr.Error())
	}

	// Verify only the first emit happened.
	if len(emitted) != 1 || emitted[0] != "Handler started..." {
		t.Errorf("Expected only 'Handler started...' to be emitted, got: %v", emitted)
	}

	t.Logf("[DEBUG] Turn 6: Test completed successfully, panic was caught and reported.")
}
