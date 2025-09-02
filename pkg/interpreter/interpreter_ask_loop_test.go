// NeuroScript Version: 0.7.0
// File version: 15
// Purpose: Corrected mock providers to include a mandatory UserData section in composed envelopes, fixing the 'invalid envelope marker' parsing errors.
// filename: pkg/interpreter/interpreter_ask_loop_test.go
// nlines: 160
// risk_rating: LOW

package interpreter_test

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/parser"
	"github.com/aprice2704/neuroscript/pkg/provider"
)

type mockLoopingProvider struct {
	turnCount int32
}

func (m *mockLoopingProvider) Chat(ctx context.Context, req provider.AIRequest) (*provider.AIResponse, error) {
	turn := atomic.AddInt32(&m.turnCount, 1)

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
	interp, err := interpreter.NewTestInterpreter(t, nil, nil, true)
	if err != nil {
		t.Fatalf("Failed to create test interpreter: %v", err)
	}
	interp.RegisterProvider("mock_looper", &mockLoopingProvider{})

	modelConfig := map[string]lang.Value{
		"provider":            lang.StringValue{Value: "mock_looper"},
		"model":               lang.StringValue{Value: "looper_model"},
		"tool_loop_permitted": lang.BoolValue{Value: true},
		"max_turns":           lang.NumberValue{Value: 5},
	}
	_ = interp.RegisterAgentModel("test_agent", modelConfig)

	script := `command ask "test_agent", "start" into final_result endcommand`

	p := parser.NewParserAPI(nil)
	tree, _ := p.Parse(script)
	builder := parser.NewASTBuilder(nil)
	program, _, _ := builder.Build(tree)
	_, err = interp.Execute(program)

	if err != nil {
		t.Fatalf("Expected loop to succeed, but it failed: %v", err)
	}
	resultVar, _ := interp.GetVariable("final_result")
	resultStr, _ := lang.ToString(resultVar)
	if !strings.Contains(resultStr, "result from turn 3") {
		t.Errorf("Expected final result to contain output from turn 3, but got: %s", resultStr)
	}
}

func TestAutoLoop_MaxTurnsExceeded(t *testing.T) {
	interp, err := interpreter.NewTestInterpreter(t, nil, nil, true)
	if err != nil {
		t.Fatalf("Failed to create test interpreter: %v", err)
	}
	interp.RegisterProvider("mock_looper", &mockLoopingProvider{})

	modelConfig := map[string]lang.Value{
		"provider":            lang.StringValue{Value: "mock_looper"},
		"model":               lang.StringValue{Value: "looper_model"},
		"tool_loop_permitted": lang.BoolValue{Value: true},
		"max_turns":           lang.NumberValue{Value: 2},
	}
	_ = interp.RegisterAgentModel("test_agent", modelConfig)

	script := `command ask "test_agent", "start" into result endcommand`

	p := parser.NewParserAPI(nil)
	tree, _ := p.Parse(script)
	builder := parser.NewASTBuilder(nil)
	program, _, _ := builder.Build(tree)
	_, err = interp.Execute(program)

	if err != nil {
		t.Fatalf("Script execution failed unexpectedly: %v", err)
	}
	resultVar, _ := interp.GetVariable("result")
	resultStr, _ := lang.ToString(resultVar)
	if !strings.Contains(resultStr, "result from turn 2") {
		t.Errorf("Expected result from turn 2, but got: %s", resultStr)
	}
}

type mockAbortingProvider struct{}

func (m *mockAbortingProvider) Chat(ctx context.Context, req provider.AIRequest) (*provider.AIResponse, error) {
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
	interp, err := interpreter.NewTestInterpreter(t, nil, nil, true)
	if err != nil {
		t.Fatalf("Failed to create test interpreter: %v", err)
	}
	interp.RegisterProvider("mock_aborter", &mockAbortingProvider{})
	modelConfig := map[string]lang.Value{
		"provider":            lang.StringValue{Value: "mock_aborter"},
		"model":               lang.StringValue{Value: "aborter_model"},
		"tool_loop_permitted": lang.BoolValue{Value: true},
		"max_turns":           lang.NumberValue{Value: 5},
	}
	_ = interp.RegisterAgentModel("test_agent", modelConfig)

	script := `command ask "test_agent", "start" endcommand`
	p := parser.NewParserAPI(nil)
	tree, _ := p.Parse(script)
	builder := parser.NewASTBuilder(nil)
	program, _, _ := builder.Build(tree)
	_, err = interp.Execute(program)

	if err != nil {
		t.Fatalf("Expected loop to succeed with empty result, but it failed: %v", err)
	}
}
