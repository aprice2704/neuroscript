// NeuroScript Version: 0.8.0
// File version: 23
// Purpose: Corrected calls to provider.NewAdmin to include the ExecPolicy.
// filename: pkg/interpreter/ask_loop_test.go
// nlines: 172
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

// mockLoopingProvider now emits '<<<LOOP:DONE>>>' on turn 3.
type mockLoopingProvider struct {
	turnCount int32
	t         *testing.T
}

func (m *mockLoopingProvider) Chat(ctx context.Context, req provider.AIRequest) (*provider.AIResponse, error) {
	turn := atomic.AddInt32(&m.turnCount, 1)
	m.t.Logf("[DEBUG] mockLoopingProvider: Chat call, turn %d", turn)

	// THE FIX: The AI emits its answer, and optionally emits the DONE signal.
	var actionsScript string
	if turn >= 3 {
		actionsScript = fmt.Sprintf(`
		command
			emit "This is the result from turn %d."
			emit "<<<LOOP:DONE>>>"
		endcommand
		`, turn)
	} else {
		actionsScript = fmt.Sprintf(`
		command
			emit "This is the result from turn %d."
		endcommand
		`, turn)
	}

	env := &aeiou.Envelope{UserData: "{}", Actions: actionsScript}
	respText, _ := env.Compose()

	return &provider.AIResponse{
		TextContent: respText,
	}, nil
}

// mockStuckProvider always returns the exact same response (and no DONE signal).
type mockStuckProvider struct {
	t *testing.T
}

func (m *mockStuckProvider) Chat(ctx context.Context, req provider.AIRequest) (*provider.AIResponse, error) {
	m.t.Logf("[DEBUG] mockStuckProvider: Chat call, returning a static response.")
	actionsScript := `
	command
		emit "I am stuck."
	endcommand
	`
	env := &aeiou.Envelope{UserData: "{}", Actions: actionsScript}
	respText, _ := env.Compose()
	return &provider.AIResponse{TextContent: respText}, nil
}

func TestAutoLoop_Success(t *testing.T) {
	h := NewTestHarness(t)
	t.Logf("[DEBUG] Turn 1: Starting TestAutoLoop_Success.")
	// --- FIX: Register provider via the harness's registry ---
	if err := provider.NewAdmin(h.ProviderRegistry, h.Interpreter.GetExecPolicy()).Register("mock_looper", &mockLoopingProvider{t: t}); err != nil {
		t.Fatalf("Failed to register mock provider: %v", err)
	}
	// --- End Fix ---

	modelConfig := map[string]lang.Value{
		"provider":            lang.StringValue{Value: "mock_looper"},
		"model":               lang.StringValue{Value: "looper_model"},
		"tool_loop_permitted": lang.BoolValue{Value: true},
		"max_turns":           lang.NumberValue{Value: 5}, // Set higher than the AI's stop turn.
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
	// THE FIX: We now assert that the loop correctly stopped at turn 3
	// because of the <<<LOOP:DONE>>> signal.
	if !strings.Contains(resultStr, "result from turn 3") {
		t.Errorf("Expected final result to contain output from turn 3, but got: %s", resultStr)
	}
	t.Logf("[DEBUG] Turn 4: TestAutoLoop_Success completed.")
}

func TestAutoLoop_MaxTurnsExceeded(t *testing.T) {
	h := NewTestHarness(t)
	t.Logf("[DEBUG] Turn 1: Starting TestAutoLoop_MaxTurnsExceeded.")
	// --- FIX: Register provider via the harness's registry ---
	if err := provider.NewAdmin(h.ProviderRegistry, h.Interpreter.GetExecPolicy()).Register("mock_looper", &mockLoopingProvider{t: t}); err != nil {
		t.Fatalf("Failed to register mock provider: %v", err)
	}
	// --- End Fix ---

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
	// This test is still valid. The loop is cut off by max_turns at 2,
	// before the AI can signal 'done' on turn 3.
	if !strings.Contains(resultStr, "result from turn 2") {
		t.Errorf("Expected result from turn 2, but got: %s", resultStr)
	}
	t.Logf("[DEBUG] Turn 4: TestAutoLoop_MaxTurnsExceeded completed.")
}

func TestAutoLoop_ProgressGuard(t *testing.T) {
	h := NewTestHarness(t)
	t.Logf("[DEBUG] Turn 1: Starting TestAutoLoop_ProgressGuard.")
	mockProv := &mockStuckProvider{t: t}
	// --- FIX: Register provider via the harness's registry ---
	if err := provider.NewAdmin(h.ProviderRegistry, h.Interpreter.GetExecPolicy()).Register("mock_stuck_provider", mockProv); err != nil {
		t.Fatalf("Failed to register mock provider: %v", err)
	}
	// --- End Fix ---

	modelConfig := map[string]lang.Value{
		"provider":            lang.StringValue{Value: "mock_stuck_provider"},
		"model":               lang.StringValue{Value: "stuck_model"},
		"tool_loop_permitted": lang.BoolValue{Value: true},
		"max_turns":           lang.NumberValue{Value: 10}, // High max_turns
	}
	_ = h.Interpreter.RegisterAgentModel("stuck_agent", modelConfig)
	t.Logf("[DEBUG] Turn 2: Stuck provider and agent registered.")

	script := `command
		ask "stuck_agent", "get stuck" into final_result
		emit final_result
	endcommand`

	var capturedEmits []string
	h.HostContext.EmitFunc = func(v lang.Value) {
		capturedEmits = append(capturedEmits, v.String())
	}

	tree, _ := h.Parser.Parse(script)
	program, _, _ := h.ASTBuilder.Build(tree)
	h.Interpreter.Load(&interfaces.Tree{Root: program})
	t.Logf("[DEBUG] Turn 3: Script loaded. Executing commands.")

	_, err := h.Interpreter.Execute(program)
	if err != nil {
		t.Fatalf("Expected loop to terminate gracefully, but it failed: %v", err)
	}

	if len(capturedEmits) == 0 {
		t.Fatal("Expected script to emit a final result, but it emitted nothing.")
	}
	resultStr := capturedEmits[0]
	if !strings.Contains(resultStr, "I am stuck.") {
		t.Errorf("Expected final result to contain 'I am stuck.', but got: %s", resultStr)
	}
	// This test is still valid. The mock AI never emits the DONE signal,
	// so it is correctly terminated by the progress guard.
	t.Logf("[DEBUG] Turn 4: TestAutoLoop_ProgressGuard completed, loop correctly terminated by progress guard.")
}
