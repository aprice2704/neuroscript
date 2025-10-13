// NeuroScript Version: 0.8.0
// File version: 11
// Purpose: Rewrote test to correctly validate 'ask' result by checking emitted output instead of a leaked variable.
// filename: pkg/interpreter/ask_permission_test.go
// nlines: 129
// risk_rating: LOW

package interpreter_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/provider"
)

// mockPermissionProvider can be configured to return a 'continue' or 'done' signal.
type mockPermissionProvider struct {
	shouldContinue bool
	t              *testing.T
}

func (m *mockPermissionProvider) Chat(ctx context.Context, req provider.AIRequest) (*provider.AIResponse, error) {
	m.t.Logf("[DEBUG] Turn X: mockPermissionProvider.Chat called.")
	control := "done"
	if m.shouldContinue {
		control = "continue"
	}
	actionsScript := fmt.Sprintf(`
	command
		emit "one-shot success"
		set params = {"action": "%s"}
		emit tool.aeiou.magic("LOOP", params)
	endcommand`, control)

	env := &aeiou.Envelope{UserData: "{}", Actions: actionsScript}
	respText, _ := env.Compose()
	return &provider.AIResponse{TextContent: respText}, nil
}

func TestAskLoopPermission(t *testing.T) {
	restrictedAgentConfig := map[string]lang.Value{
		"provider":            lang.StringValue{Value: "mock_permission_provider"},
		"model":               lang.StringValue{Value: "restricted_model"},
		"tool_loop_permitted": lang.BoolValue{Value: false}, // Explicitly false
		"max_turns":           lang.NumberValue{Value: 2},   // Allow more than one turn to test the policy
	}

	permissiveAgentConfig := map[string]lang.Value{
		"provider":            lang.StringValue{Value: "mock_permission_provider"},
		"model":               lang.StringValue{Value: "permissive_model"},
		"tool_loop_permitted": lang.BoolValue{Value: true},
	}

	t.Run("Success: Restricted agent in one-shot mode", func(t *testing.T) {
		h := NewTestHarness(t)
		t.Logf("[DEBUG] Turn 1: Starting 'Success: Restricted agent in one-shot mode'.")
		// Provider will return a "done" signal.
		h.Interpreter.RegisterProvider("mock_permission_provider", &mockPermissionProvider{shouldContinue: false, t: t})
		_ = h.Interpreter.RegisterAgentModel("restricted_agent", restrictedAgentConfig)

		script := `command
			ask "restricted_agent", "go" into result
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
		t.Logf("[DEBUG] Turn 2: Executing script.")
		_, err := h.Interpreter.Execute(program)

		if err != nil {
			t.Fatalf("Expected one-shot ask to succeed for restricted agent, but got error: %v", err)
		}
		if len(capturedEmits) != 1 {
			t.Fatalf("Expected 1 emit, but got %d", len(capturedEmits))
		}
		if s := capturedEmits[0]; s != "one-shot success" {
			t.Errorf("Expected result 'one-shot success', got '%s'", s)
		}
		t.Logf("[DEBUG] Turn 3: Test completed successfully.")
	})

	t.Run("Failure: Restricted agent attempts to loop", func(t *testing.T) {
		h := NewTestHarness(t)
		t.Logf("[DEBUG] Turn 1: Starting 'Failure: Restricted agent attempts to loop'.")
		// Provider will return a "continue" signal, which is a violation.
		h.Interpreter.RegisterProvider("mock_permission_provider", &mockPermissionProvider{shouldContinue: true, t: t})
		_ = h.Interpreter.RegisterAgentModel("restricted_agent", restrictedAgentConfig)

		script := `command ask "restricted_agent", "go" endcommand`
		tree, _ := h.Parser.Parse(script)
		program, _, _ := h.ASTBuilder.Build(tree)
		if err := h.Interpreter.Load(&interfaces.Tree{Root: program}); err != nil {
			t.Fatalf("Failed to load program: %v", err)
		}
		t.Logf("[DEBUG] Turn 2: Executing script, expecting policy error.")
		_, err := h.Interpreter.Execute(program)

		if err == nil {
			t.Fatal("Expected ask to fail with a policy error, but it succeeded.")
		}
		var rtErr *lang.RuntimeError
		if !errors.As(err, &rtErr) || rtErr.Code != lang.ErrorCodePolicy {
			t.Errorf("Expected a RuntimeError with ErrorCodePolicy, but got: %v", err)
		}
		t.Logf("[DEBUG] Turn 3: Test completed with expected error: %v", err)
	})

	t.Run("Success: Permissive agent attempts to loop", func(t *testing.T) {
		h := NewTestHarness(t)
		t.Logf("[DEBUG] Turn 1: Starting 'Success: Permissive agent attempts to loop'.")
		// Provider will return a "continue" signal, which is allowed.
		h.Interpreter.RegisterProvider("mock_permission_provider", &mockPermissionProvider{shouldContinue: true, t: t})
		_ = h.Interpreter.RegisterAgentModel("permissive_agent", permissiveAgentConfig)

		script := `command ask "permissive_agent", "go" endcommand`
		tree, _ := h.Parser.Parse(script)
		program, _, _ := h.ASTBuilder.Build(tree)
		if err := h.Interpreter.Load(&interfaces.Tree{Root: program}); err != nil {
			t.Fatalf("Failed to load program: %v", err)
		}
		t.Logf("[DEBUG] Turn 2: Executing script.")
		_, err := h.Interpreter.Execute(program)

		if err != nil {
			t.Fatalf("Expected ask to succeed for permissive agent, but got error: %v", err)
		}
		t.Logf("[DEBUG] Turn 3: Test completed successfully.")
	})
}
