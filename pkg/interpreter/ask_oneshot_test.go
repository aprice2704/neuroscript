// NeuroScript Version: 0.8.0
// File version: 12
// Purpose: Refactored to remove obsolete loop permission tests. 'max_turns' now controls loops.
// filename: pkg/interpreter/ask_oneshot_test.go
// nlines: 66
// risk_rating: LOW

package interpreter_test

import (
	"context"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/provider"
)

// mockPermissionProvider now just emits a simple success message.
type mockPermissionProvider struct {
	t *testing.T
}

func (m *mockPermissionProvider) Chat(ctx context.Context, req provider.AIRequest) (*provider.AIResponse, error) {
	m.t.Logf("[DEBUG] Turn X: mockPermissionProvider.Chat called.")
	// THE FIX: The AI no longer needs to send control tokens.
	actionsScript := `
	command
		emit "one-shot success"
	endcommand`

	env := &aeiou.Envelope{UserData: "{}", Actions: actionsScript}
	respText, _ := env.Compose()
	return &provider.AIResponse{TextContent: respText}, nil
}

func TestAskLoopPermission(t *testing.T) {
	// This test is simplified. The 'tool_loop_permitted' flag is no longer
	// checked by the host loop, so we just test a basic one-shot call.
	agentConfig := map[string]lang.Value{
		"provider":  lang.StringValue{Value: "mock_permission_provider"},
		"model":     lang.StringValue{Value: "restricted_model"},
		"max_turns": lang.NumberValue{Value: 1}, // This is now how loops are controlled.
	}

	t.Run("Success: Agent in one-shot mode", func(t *testing.T) {
		h := NewTestHarness(t)
		t.Logf("[DEBUG] Turn 1: Starting 'Success: Agent in one-shot mode'.")
		h.Interpreter.RegisterProvider("mock_permission_provider", &mockPermissionProvider{t: t})
		_ = h.Interpreter.RegisterAgentModel("one_shot_agent", agentConfig)

		script := `command
			ask "one_shot_agent", "go" into result
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
			t.Fatalf("Expected one-shot ask to succeed, but got error: %v", err)
		}
		if len(capturedEmits) != 1 {
			t.Fatalf("Expected 1 emit, but got %d", len(capturedEmits))
		}
		if s := capturedEmits[0]; s != "one-shot success" {
			t.Errorf("Expected result 'one-shot success', got '%s'", s)
		}
		t.Logf("[DEBUG] Turn 3: Test completed successfully.")
	})

	// NOTE: The 'Failure: Restricted agent attempts to loop' and
	// 'Success: Permissive agent attempts to loop' tests are removed
	// as they are obsolete. The 'tool_loop_permitted' flag is
	// no longer used by the Go-based ask loop.
}
