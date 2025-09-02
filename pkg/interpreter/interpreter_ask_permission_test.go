// NeuroScript Version: 0.7.0
// File version: 9
// Purpose: Corrected the looping test by explicitly setting 'max_turns' to 2, ensuring the policy violation is actually triggered on the second turn.
// filename: pkg/interpreter/interpreter_ask_permission_test.go
// nlines: 121
// risk_rating: LOW

package interpreter_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/parser"
	"github.com/aprice2704/neuroscript/pkg/provider"
)

// mockPermissionProvider can be configured to return a 'continue' or 'done' signal.
type mockPermissionProvider struct {
	shouldContinue bool
}

func (m *mockPermissionProvider) Chat(ctx context.Context, req provider.AIRequest) (*provider.AIResponse, error) {
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
		interp, err := interpreter.NewTestInterpreter(t, nil, nil, true)
		if err != nil {
			t.Fatal(err)
		}
		// Provider will return a "done" signal.
		interp.RegisterProvider("mock_permission_provider", &mockPermissionProvider{shouldContinue: false})
		_ = interp.RegisterAgentModel("restricted_agent", restrictedAgentConfig)

		script := `command ask "restricted_agent", "go" into result endcommand`
		p := parser.NewParserAPI(nil)
		tree, _ := p.Parse(script)
		program, _, _ := parser.NewASTBuilder(nil).Build(tree)
		_, err = interp.Execute(program)

		if err != nil {
			t.Fatalf("Expected one-shot ask to succeed for restricted agent, but got error: %v", err)
		}
		result, _ := interp.GetVariable("result")
		if s, _ := lang.ToString(result); s != "one-shot success" {
			t.Errorf("Expected result 'one-shot success', got '%s'", s)
		}
	})

	t.Run("Failure: Restricted agent attempts to loop", func(t *testing.T) {
		interp, err := interpreter.NewTestInterpreter(t, nil, nil, true)
		if err != nil {
			t.Fatal(err)
		}
		// Provider will return a "continue" signal, which is a violation.
		interp.RegisterProvider("mock_permission_provider", &mockPermissionProvider{shouldContinue: true})
		_ = interp.RegisterAgentModel("restricted_agent", restrictedAgentConfig)

		script := `command ask "restricted_agent", "go" endcommand`
		p := parser.NewParserAPI(nil)
		tree, _ := p.Parse(script)
		program, _, _ := parser.NewASTBuilder(nil).Build(tree)
		_, err = interp.Execute(program)

		if err == nil {
			t.Fatal("Expected ask to fail with a policy error, but it succeeded.")
		}
		var rtErr *lang.RuntimeError
		if !errors.As(err, &rtErr) || rtErr.Code != lang.ErrorCodePolicy {
			t.Errorf("Expected a RuntimeError with ErrorCodePolicy, but got: %v", err)
		}
	})

	t.Run("Success: Permissive agent attempts to loop", func(t *testing.T) {
		interp, err := interpreter.NewTestInterpreter(t, nil, nil, true)
		if err != nil {
			t.Fatal(err)
		}
		// Provider will return a "continue" signal, which is allowed.
		// The loop will terminate after one turn because maxTurns defaults to 1.
		interp.RegisterProvider("mock_permission_provider", &mockPermissionProvider{shouldContinue: true})
		_ = interp.RegisterAgentModel("permissive_agent", permissiveAgentConfig)

		script := `command ask "permissive_agent", "go" endcommand`
		p := parser.NewParserAPI(nil)
		tree, _ := p.Parse(script)
		program, _, _ := parser.NewASTBuilder(nil).Build(tree)
		_, err = interp.Execute(program)

		if err != nil {
			t.Fatalf("Expected ask to succeed for permissive agent, but got error: %v", err)
		}
	})
}
