// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 5
// :: description: Tests for enhanced error formatting including stack traces, definition locations, covering events and command blocks.
// :: latestChange: Updated assertions to verify presence of definition locations.
// :: filename: pkg/interpreter/error_formatting_test.go
// :: serialization: go

package interpreter_test

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

func TestErrorFormatting_StackTrace(t *testing.T) {
	script := `
func fail_here() means
	fail "intentional failure"
endfunc

func intermediate() means
	call fail_here()
endfunc

func main() means
	call intermediate()
endfunc
`
	t.Logf("[DEBUG] Turn 1: Starting TestErrorFormatting_StackTrace")
	h := NewTestHarness(t)
	interp := h.Interpreter

	tree, _ := h.Parser.Parse(script)
	program, _, _ := h.ASTBuilder.Build(tree)
	interp.Load(&interfaces.Tree{Root: program})

	_, err := interp.Run("main")

	if err == nil {
		t.Fatal("Expected execution to fail, but it succeeded.")
	}

	errMsg := err.Error()
	t.Logf("\n--- [Simple Call Stack Output] ---\n%s\n----------------------------------", errMsg)

	expectedParts := []string{
		"intentional failure",
		"Stack Trace:",
		"main",
		"intermediate",
		"fail_here",
		"(defined at", // Verify definition location is present
	}

	for _, part := range expectedParts {
		if !strings.Contains(errMsg, part) {
			t.Errorf("Error message missing expected part: %q", part)
		}
	}
}

func TestErrorFormatting_DeepRecursion(t *testing.T) {
	script := `
func recursive(needs n) means
	if n <= 0
		fail "bottom of recursion"
	endif
	call recursive(n - 1)
endfunc

func main() means
	call recursive(5)
endfunc
`
	t.Logf("[DEBUG] Turn 1: Starting TestErrorFormatting_DeepRecursion")
	h := NewTestHarness(t)
	interp := h.Interpreter

	tree, _ := h.Parser.Parse(script)
	program, _, _ := h.ASTBuilder.Build(tree)
	interp.Load(&interfaces.Tree{Root: program})

	_, err := interp.Run("main")

	if err == nil {
		t.Fatal("Expected execution to fail, but it succeeded.")
	}

	errMsg := err.Error()
	t.Logf("\n--- [Deep Recursion Stack Output] ---\n%s\n-------------------------------------", errMsg)

	if !strings.Contains(errMsg, "bottom of recursion") {
		t.Errorf("Error message missing failure reason")
	}

	// Count occurrences of "recursive" in the stack trace
	// 5 down to 0 is 6 calls total.
	count := strings.Count(errMsg, "recursive")
	if count < 6 {
		t.Errorf("Expected at least 6 frames of 'recursive', found %d", count)
	}

	// Verify definition info is present
	if !strings.Contains(errMsg, "(defined at") {
		t.Error("Stack trace missing definition locations")
	}
}

func TestErrorFormatting_ToolFailure(t *testing.T) {
	t.Logf("[DEBUG] Turn 1: Starting TestErrorFormatting_ToolFailure")
	h := NewTestHarness(t)
	interp := h.Interpreter

	// Register a tool that returns a specific error
	failingTool := tool.ToolImplementation{
		Spec: tool.ToolSpec{Name: "Fail", Group: "test", ReturnType: "string"},
		Func: func(rt tool.Runtime, args []interface{}) (interface{}, error) {
			return nil, fmt.Errorf("tool-level error occurred")
		},
	}
	interp.ToolRegistry().RegisterTool(failingTool)

	script := `
func invoke_tool() means
	call tool.test.Fail()
endfunc

func main() means
	call invoke_tool()
endfunc
`
	tree, _ := h.Parser.Parse(script)
	program, _, _ := h.ASTBuilder.Build(tree)
	interp.Load(&interfaces.Tree{Root: program})

	_, err := interp.Run("main")

	if err == nil {
		t.Fatal("Expected execution to fail, but it succeeded.")
	}

	errMsg := err.Error()
	t.Logf("\n--- [Tool Failure Stack Output] ---\n%s\n-----------------------------------", errMsg)

	// Check for the tool error message
	if !strings.Contains(errMsg, "tool-level error occurred") {
		t.Errorf("Error message missing underlying tool error")
	}

	// Check stack trace validity
	// It should show the NeuroScript function that called the tool.
	expectedParts := []string{
		"Stack Trace:",
		"main",
		"invoke_tool",
		"(defined at",
	}
	for _, part := range expectedParts {
		if !strings.Contains(errMsg, part) {
			t.Errorf("Error message missing expected stack frame part: %q", part)
		}
	}
}

func TestErrorFormatting_CommandBlock(t *testing.T) {
	script := `
command
	set x = 10
	fail "failure in command block"
endcommand
`
	t.Logf("[DEBUG] Turn 1: Starting TestErrorFormatting_CommandBlock")
	h := NewTestHarness(t)
	interp := h.Interpreter

	tree, _ := h.Parser.Parse(script)
	program, _, _ := h.ASTBuilder.Build(tree)

	// Execute the command block
	_, err := interp.Execute(program)

	if err == nil {
		t.Fatal("Expected execution to fail, but it succeeded.")
	}

	errMsg := err.Error()
	t.Logf("\n--- [Command Block Stack Output] ---\n%s\n------------------------------------", errMsg)

	if !strings.Contains(errMsg, "failure in command block") {
		t.Errorf("Error message missing failure reason")
	}

	// Verify command block location context
	if !strings.Contains(errMsg, "Command Block (at") {
		t.Errorf("Error message missing command block location context")
	}
}

func TestErrorFormatting_EventHandler(t *testing.T) {
	script := `
on event "test:fail" do
	call fail_helper()
endon

func fail_helper() means
	fail "failure in event handler"
endfunc
`
	t.Logf("[DEBUG] Turn 1: Starting TestErrorFormatting_EventHandler")
	h := NewTestHarness(t)
	interp := h.Interpreter

	// 1. Setup error capture for event handler
	var capturedErr error
	var wg sync.WaitGroup
	wg.Add(1)

	h.HostContext.EventHandlerErrorCallback = func(eventName, source string, err *lang.RuntimeError) {
		capturedErr = err
		wg.Done()
	}

	// 2. Load script
	tree, _ := h.Parser.Parse(script)
	program, _, _ := h.ASTBuilder.Build(tree)
	interp.Load(&interfaces.Tree{Root: program})

	// 3. Emit event
	interp.EmitEvent("test:fail", "test_source", nil)

	// 4. Wait for handler
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for event handler error callback")
	}

	if capturedErr == nil {
		t.Fatal("Expected an error from event handler, got nil")
	}

	errMsg := capturedErr.Error()
	t.Logf("\n--- [Event Handler Stack Output] ---\n%s\n------------------------------------", errMsg)

	if !strings.Contains(errMsg, "failure in event handler") {
		t.Errorf("Error message missing failure reason")
	}

	// Check stack trace
	// Should contain: fail_helper -> (event handler context)
	expectedParts := []string{
		"Stack Trace:",
		"fail_helper",
		"(defined at",
		"Event: test:fail (defined at",
	}
	for _, part := range expectedParts {
		if !strings.Contains(errMsg, part) {
			t.Errorf("Error message missing expected part: %q", part)
		}
	}
}
