// NeuroScript Version: 0.8.0
// File version: 17
// Purpose: Rewrote tests to correctly validate 'ask' loop results by checking emitted output instead of leaked variables.
// filename: pkg/interpreter/ask_loop_test.go
// nlines: 162
// risk_rating: LOW

package interpreter_test

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/provider"
)

type mockLoopingProvider struct {
	turnCount int32
	t         *testing.T
}

func (m *mockLoopingProvider) Chat(ctx context.Context, req provider.AIRequest) (*provider.AIResponse, error) {
	turn := atomic.AddInt32(&m.turnCount, 1)
	m.t.Logf("[DEBUG] mockLoopingProvider: Chat call, turn %d", turn)

	var control, notes string
	if turn >= 3 {
		control = "done"
		notes = "Completed the task in three turns."
	} else {
		control = "continue"
		notes = fmt.Sprintf("Continuing to turn %d.", turn+1)
	}

	actionsScript := fmt.Sprintf(`
	command
		emit "This is the result from turn %d."
		set params = {"action": "%s", "notes": "%s"}
		emit tool.aeiou.magic("LOOP", params)
	endcommand
	`, turn, control, notes)

	env := &aeiou.Envelope{UserData: "{}", Actions: actionsScript}
	respText, _ := env.Compose()

	return &provider.AIResponse{
		TextContent: respText,
	}, nil
}

func TestAutoLoop_Success(t *testing.T) {
	h := NewTestHarness(t)
	t.Logf("[DEBUG] Turn 1: Starting TestAutoLoop_Success.")
	h.Interpreter.RegisterProvider("mock_looper", &mockLoopingProvider{t: t})

	modelConfig := map[string]lang.Value{
		"provider":            lang.StringValue{Value: "mock_looper"},
		"model":               lang.StringValue{Value: "looper_model"},
		"tool_loop_permitted": lang.BoolValue{Value: true},
		"max_turns":           lang.NumberValue{Value: 5},
	}
	_ = h.Interpreter.RegisterAgentModel("test_agent", modelConfig)
	t.Logf("[DEBUG] Turn 2: Mock provider and agent registered.")

	script := `command
		ask "test_agent", "start" into final_result
		emit final_result
	endcommand`

	var capturedEmits []string
	h.HostContext.EmitFunc = func(v lang.Value) {
		capturedEmits = append(capturedEmits, v.String())
	}

	tree, _ := h.Parser.Parse(script)
	program, _, _ := h.ASTBuilder.Build(tree)
	if err := h.Interpreter.Load(&interfaces.Tree{Root: program}); err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}
	t.Logf("[DEBUG] Turn 3: Script loaded. Executing commands.")

	_, err := h.Interpreter.Execute(program)

	if err != nil {
		t.Fatalf("Expected loop to succeed, but it failed: %v", err)
	}

	if len(capturedEmits) == 0 {
		t.Fatal("Expected script to emit a final result, but it emitted nothing.")
	}
	resultStr := capturedEmits[0]
	if !strings.Contains(resultStr, "result from turn 3") {
		t.Errorf("Expected final result to contain output from turn 3, but got: %s", resultStr)
	}
	t.Logf("[DEBUG] Turn 4: TestAutoLoop_Success completed.")
}

func TestAutoLoop_MaxTurnsExceeded(t *testing.T) {
	h := NewTestHarness(t)
	t.Logf("[DEBUG] Turn 1: Starting TestAutoLoop_MaxTurnsExceeded.")
	h.Interpreter.RegisterProvider("mock_looper", &mockLoopingProvider{t: t})

	modelConfig := map[string]lang.Value{
		"provider":            lang.StringValue{Value: "mock_looper"},
		"model":               lang.StringValue{Value: "looper_model"},
		"tool_loop_permitted": lang.BoolValue{Value: true},
		"max_turns":           lang.NumberValue{Value: 2}, // Set max turns to 2
	}
	_ = h.Interpreter.RegisterAgentModel("test_agent", modelConfig)
	t.Logf("[DEBUG] Turn 2: Mock provider and agent registered.")

	script := `command
		ask "test_agent", "start" into result
		emit result
	endcommand`

	var capturedEmits []string
	h.HostContext.EmitFunc = func(v lang.Value) {
		capturedEmits = append(capturedEmits, v.String())
	}

	tree, _ := h.Parser.Parse(script)
	program, _, _ := h.ASTBuilder.Build(tree)
	if err := h.Interpreter.Load(&interfaces.Tree{Root: program}); err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}
	t.Logf("[DEBUG] Turn 3: Script loaded. Executing commands.")

	_, err := h.Interpreter.Execute(program)

	if err != nil {
		t.Fatalf("Script execution failed unexpectedly: %v", err)
	}

	if len(capturedEmits) == 0 {
		t.Fatal("Expected script to emit a final result, but it emitted nothing.")
	}
	resultStr := capturedEmits[0]
	if !strings.Contains(resultStr, "result from turn 2") {
		t.Errorf("Expected result from turn 2, but got: %s", resultStr)
	}
	t.Logf("[DEBUG] Turn 4: TestAutoLoop_MaxTurnsExceeded completed.")
}

type mockAbortingProvider struct {
	t *testing.T
}

func (m *mockAbortingProvider) Chat(ctx context.Context, req provider.AIRequest) (*provider.AIResponse, error) {
	m.t.Logf("[DEBUG] mockAbortingProvider: Chat call, returning abort.")
	actionsScript := `
	command
		set params = {"action": "abort", "reason": "precondition_failed"}
		emit tool.aeiou.magic("LOOP", params)
	endcommand
	`
	env := &aeiou.Envelope{UserData: "{}", Actions: actionsScript}
	respText, _ := env.Compose()
	return &provider.AIResponse{TextContent: respText}, nil
}

func TestAutoLoop_Abort(t *testing.T) {
	h := NewTestHarness(t)
	t.Logf("[DEBUG] Turn 1: Starting TestAutoLoop_Abort.")
	h.Interpreter.RegisterProvider("mock_aborter", &mockAbortingProvider{t: t})
	modelConfig := map[string]lang.Value{
		"provider":            lang.StringValue{Value: "mock_aborter"},
		"model":               lang.StringValue{Value: "aborter_model"},
		"tool_loop_permitted": lang.BoolValue{Value: true},
		"max_turns":           lang.NumberValue{Value: 5},
	}
	_ = h.Interpreter.RegisterAgentModel("test_agent", modelConfig)
	t.Logf("[DEBUG] Turn 2: Mock provider and agent registered.")

	script := `command ask "test_agent", "start" into result; emit result endcommand`
	tree, _ := h.Parser.Parse(script)
	program, _, _ := h.ASTBuilder.Build(tree)
	if err := h.Interpreter.Load(&interfaces.Tree{Root: program}); err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}
	t.Logf("[DEBUG] Turn 3: Script loaded. Executing commands.")
	_, err := h.Interpreter.Execute(program)

	if err != nil {
		t.Fatalf("Expected loop to succeed with empty result, but it failed: %v", err)
	}
	t.Logf("[DEBUG] Turn 4: TestAutoLoop_Abort completed.")
}
