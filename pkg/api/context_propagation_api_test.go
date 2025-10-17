// NeuroScript Version: 0.8.0
// File version: 8
// Purpose: Removes obsolete EndToEnd and NestedCall context propagation tests per user request.
// filename: pkg/api/context_propagation_api_test.go
// nlines: 201
// risk_rating: LOW

package api_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os" // DEBUG
	"testing"

	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/api"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/provider"
)

// mockContextProviderAPI is an AI provider that executes a specific callback
// inside its Chat method, simulating the 'ask' loop's action execution.
type mockContextProviderAPI struct {
	t        *testing.T
	Callback func() (string, error)
}

func (m *mockContextProviderAPI) Chat(ctx context.Context, req provider.AIRequest) (*provider.AIResponse, error) {
	// DEBUG
	fmt.Fprintf(os.Stderr, "[DEBUG] mockContextProviderAPI.Chat: Callback initiated.\n")
	actions, err := m.Callback()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[DEBUG] mockContextProviderAPI.Chat: Callback returned error: %v\n", err)
		return nil, err
	}
	fmt.Fprintf(os.Stderr, "[DEBUG] mockContextProviderAPI.Chat: Callback returned actions:\n%s\n", actions)
	env := &aeiou.Envelope{UserData: "{}", Actions: actions}
	respText, _ := env.Compose()
	return &provider.AIResponse{TextContent: respText}, nil
}

// buildProbeToolAPI creates the tool used by all tests to inspect the runtime context.
func buildProbeToolAPI() api.ToolImplementation {
	return api.ToolImplementation{
		Spec: api.ToolSpec{
			Name:       "ProbeContext",
			Group:      "test",
			ReturnType: api.ArgTypeMap,
		},
		Func: func(rt api.Runtime, args []interface{}) (interface{}, error) {
			report := make(map[string]interface{})
			fmt.Printf("[DEBUG] probeTool (API): Executing probe.\n")

			if provider, ok := rt.(interfaces.ActorProvider); ok {
				if actor, found := provider.Actor(); found {
					report["actor_did"] = actor.DID()
				} else {
					report["actor_did"] = "not_found"
				}
			} else {
				report["actor_did"] = "provider_unsupported"
			}

			if ctxProvider, ok := rt.(api.TurnContextProvider); ok {
				turnCtx := ctxProvider.GetTurnContext()
				// DEBUG: Print the context pointer received by the tool
				fmt.Printf("[DEBUG] probeTool (API): Received context pointer: %p\n", turnCtx)
				if sid, ok := turnCtx.Value(interpreter.AeiouSessionIDKey).(string); ok {
					report["aeiou_session_id"] = sid
				} else {
					report["aeiou_session_id"] = "not_found"
				}
				if tid, ok := turnCtx.Value(interpreter.AeiouTurnIndexKey).(int); ok {
					report["aeiou_turn_index"] = float64(tid)
				} else {
					report["aeiou_turn_index"] = -1.0
				}
				if tnonce, ok := turnCtx.Value(interpreter.AeiouTurnNonceKey).(string); ok {
					report["aeiou_turn_nonce"] = tnonce
				} else {
					report["aeiou_turn_nonce"] = "not_found"
				}
			} else {
				report["aeiou_session_id"] = "provider_unsupported"
			}
			fmt.Fprintf(os.Stderr, "[DEBUG] probeTool (API): Probe complete. Report: %v\n", report) // DEBUG
			return report, nil
		},
	}
}

// setupPropagationTestAPI provides a correctly configured interpreter for all propagation tests.
// It now accepts an actor and a policy to correctly configure the interpreter at creation time.
func setupPropagationTestAPI(t *testing.T, actor api.Actor, policy *api.ExecPolicy) (*api.Interpreter, *mockContextProviderAPI) {
	t.Helper()

	// DEBUG
	debugWriter := &bytes.Buffer{}
	t.Cleanup(func() {
		if t.Failed() {
			fmt.Printf("--- DEBUG OUTPUT for %s ---\n%s\n", t.Name(), debugWriter.String())
		}
	})

	hcBuilder := api.NewHostContextBuilder().
		WithLogger(new(mockLogger)). // Use a simple mock logger
		WithStdout(debugWriter).
		WithStderr(debugWriter). // Capture stderr for debug
		WithStdin(new(bytes.Buffer))

	if actor != nil {
		hcBuilder.WithActor(actor)
	}
	hc, err := hcBuilder.Build()
	if err != nil {
		t.Fatalf("Failed to build HostContext: %v", err)
	}

	opts := []api.Option{api.WithHostContext(hc)}
	if policy != nil {
		opts = append(opts, api.WithExecPolicy(policy))
	}

	interp := api.New(opts...)

	if _, err := interp.ToolRegistry().RegisterTool(buildProbeToolAPI()); err != nil {
		t.Fatalf("Failed to register probe tool: %v", err)
	}

	mockProv := &mockContextProviderAPI{t: t}
	interp.RegisterProvider("probe_provider_api", mockProv)

	// User confirmed this syntax error was fixed.
	setupScript := `
    func _SetupProbeAgent() means
        set config = {"provider": "probe_provider_api", "model": "probe_model_api", "tool_loop_permitted": true, "max_turns": 5}
        must tool.agentmodel.Register("probe_agent_api", config)
    endfunc
    `
	fmt.Fprintf(os.Stderr, "[DEBUG] setupPropagationTestAPI: Parsing setup script...\n") // DEBUG
	tree, pErr := api.Parse([]byte(setupScript), api.ParseSkipComments)
	if pErr != nil {
		fmt.Fprintf(os.Stderr, "[DEBUG] setupPropagationTestAPI: Setup script parse FAILED: %v\n", pErr) // DEBUG
		t.Fatalf("Setup script parse failed: %v", pErr)
	}
	fmt.Fprintf(os.Stderr, "[DEBUG] setupPropagationTestAPI: Executing setup script...\n") // DEBUG
	_, err = api.ExecWithInterpreter(context.Background(), interp, tree)
	if err != nil {
		t.Fatalf("Failed to load setup script: %v", err)
	}
	fmt.Fprintf(os.Stderr, "[DEBUG] setupPropagationTestAPI: Running _SetupProbeAgent procedure...\n") // DEBUG
	_, err = api.RunProcedure(context.Background(), interp, "_SetupProbeAgent")
	if err != nil {
		fmt.Fprintf(os.Stderr, "[DEBUG] setupPropagationTestAPI: _SetupProbeAgent FAILED: %v\n", err) // DEBUG
		t.Fatalf("Agent setup procedure failed: %v", err)
	}
	t.Logf("[DEBUG] Agent 'probe_agent_api' registered via setup script.")

	return interp, mockProv
}

// NOTE: TestContextPropagation_EndToEnd_API and TestContextPropagation_NestedCall_API
// have been removed per user request as they are considered redundant.

// TestContextPropagation_PolicyInheritance_API verifies that a restrictive policy on the parent
// interpreter is correctly inherited and enforced by the 'ask' loop sandbox using the public API.
func TestContextPropagation_PolicyInheritance_API(t *testing.T) {
	const probeToolName = "tool.test.probecontext"

	policy := api.NewPolicyBuilder(api.ContextConfig).
		Deny(probeToolName).
		// "tool.aeiou.magic" is obsolete and removed.
		Allow("tool.agentmodel.register").
		Grant("model:admin:*"). // Grant required capability
		Build()

	interp, mockProvider := setupPropagationTestAPI(t, nil, policy)
	t.Logf("Applied restrictive policy via API")

	mockProvider.Callback = func() (string, error) {
		// THE FIX: Replace 'tool.aeiou.magic' with an emit of '<<<LOOP:DONE>>>'
		return fmt.Sprintf(`
			command
				set report = %s()
				emit report
				emit "<<<LOOP:DONE>>>"
			endcommand
		`, probeToolName), nil
	}

	script := `func main() means 
	  ask "probe_agent_api", "probe" into result 
	  return result
	endfunc`
	tree, err := api.Parse([]byte(script), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	_, err = api.ExecWithInterpreter(context.Background(), interp, tree)
	if err != nil {
		t.Fatalf("ExecWithInterpreter to load failed: %v", err)
	}

	_, err = api.RunProcedure(context.Background(), interp, "main")

	if err == nil {
		t.Fatal("Execution should have failed due to policy violation, but it succeeded.")
	}

	// THE FIX: Check for the internal lang.RuntimeError and its specific code,
	// as the api.RunProcedure does not currently wrap this in api.ErrToolDenied.
	var rtErr *lang.RuntimeError
	if !errors.As(err, &rtErr) || rtErr.Code != lang.ErrorCodePolicy {
		t.Errorf("Expected a PolicyViolation error (Code 1003), but got: %v", err)
	} else {
		t.Logf("SUCCESS (API): Correctly received policy violation error: %v", err)
	}
}
