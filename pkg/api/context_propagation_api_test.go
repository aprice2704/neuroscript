// NeuroScript Version: 0.8.0
// File version: 3
// Purpose: Corrects failing tests by adding the required 'model:admin:*' grant to the execution policy.
// filename: pkg/api/context_propagation_api_test.go
// nlines: 326
// risk_rating: HIGH

package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/api"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/provider"
)

// mockContextProviderAPI is an AI provider that executes a specific callback
// inside its Chat method, simulating the 'ask' loop's action execution.
type mockContextProviderAPI struct {
	t        *testing.T
	Callback func() (string, error)
}

func (m *mockContextProviderAPI) Chat(ctx context.Context, req provider.AIRequest) (*provider.AIResponse, error) {
	actions, err := m.Callback()
	if err != nil {
		return nil, err
	}
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
			return report, nil
		},
	}
}

// setupPropagationTestAPI provides a correctly configured interpreter for all propagation tests.
// It now accepts an actor and a policy to correctly configure the interpreter at creation time.
func setupPropagationTestAPI(t *testing.T, actor api.Actor, policy *api.ExecPolicy) (*api.Interpreter, *mockContextProviderAPI) {
	t.Helper()

	hcBuilder := api.NewHostContextBuilder().
		WithLogger(new(mockLogger)).
		WithStdout(new(bytes.Buffer)).
		WithStderr(new(bytes.Buffer)).
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

	setupScript := `
    func _SetupProbeAgent() means
        set config = {\
            "provider": "probe_provider_api",\
            "model": "probe_model_api",\
            "tool_loop_permitted": true\
        }
        must tool.agentmodel.Register("probe_agent_api", config)
    endfunc
    `
	tree, pErr := api.Parse([]byte(setupScript), api.ParseSkipComments)
	if pErr != nil {
		t.Fatalf("Setup script parse failed: %v", pErr)
	}
	_, err = api.ExecWithInterpreter(context.Background(), interp, tree)
	if err != nil {
		t.Fatalf("Failed to load setup script: %v", err)
	}
	_, err = api.RunProcedure(context.Background(), interp, "_SetupProbeAgent")
	if err != nil {
		t.Fatalf("Agent setup procedure failed: %v", err)
	}
	t.Logf("[DEBUG] Agent 'probe_agent_api' registered via setup script.")

	return interp, mockProv
}

// TestContextPropagation_EndToEnd_API verifies that critical contexts are correctly passed
// through the ask loop's sandbox into the final tool runtime via the public API.
func TestContextPropagation_EndToEnd_API(t *testing.T) {
	const actorDID = "did:test:context-probe-actor-api"
	const probeToolName = "tool.test.probecontext"

	actor := &mockActorImpl{did: actorDID}
	policy := api.NewPolicyBuilder(api.ContextConfig).
		Allow("tool.agentmodel.register").
		Allow(probeToolName).
		Allow("tool.aeiou.magic").
		Grant("model:admin:*"). // <-- FIX: Grant required capability
		Build()

	interp, mockProvider := setupPropagationTestAPI(t, actor, policy)

	mockProvider.Callback = func() (string, error) {
		return fmt.Sprintf(`
			command
				set report = %s()
				emit report
				set p = {"action": "done"}
				emit tool.aeiou.magic("LOOP", p)
			endcommand
		`, probeToolName), nil
	}

	script := `
		func main() means
			ask "probe_agent_api", "probe" into result
			return result
		endfunc
	`
	tree, err := api.Parse([]byte(script), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	_, err = api.ExecWithInterpreter(context.Background(), interp, tree)
	if err != nil {
		t.Fatalf("ExecWithInterpreter to load failed: %v", err)
	}
	t.Logf("[DEBUG] Main test script loaded via API.")

	val, err := api.RunProcedure(context.Background(), interp, "main")
	if err != nil {
		t.Fatalf("RunProcedure(main) failed unexpectedly: %v", err)
	}

	unwrapped, _ := api.Unwrap(val)
	resultStr, ok := unwrapped.(string)
	if !ok {
		t.Fatalf("Expected 'ask' to return a string, but got %T", val)
	}
	t.Logf("[DEBUG] Received raw string result from ask (API): %s", resultStr)

	var report map[string]interface{}
	if err := json.Unmarshal([]byte(resultStr), &report); err != nil {
		t.Fatalf("Failed to unmarshal JSON from result string: %v", err)
	}

	if report["actor_did"] != actorDID {
		t.Errorf("Probe failed to find correct Actor DID. Got: '%v', Want: '%s'", report["actor_did"], actorDID)
	}
	if report["aeiou_session_id"] == "not_found" || report["aeiou_session_id"] == "" {
		t.Errorf("Probe failed to find AEIOU Session ID. Got: '%v'", report["aeiou_session_id"])
	}
	if report["aeiou_turn_index"] != 1.0 {
		t.Errorf("Expected AEIOU Turn Index to be 1, but got %v", report["aeiou_turn_index"])
	}
}

// TestContextPropagation_NestedCall_API verifies context propagation through an 'ask' sandbox
// and then through a subsequent nested 'func' call sandbox using the public API.
func TestContextPropagation_NestedCall_API(t *testing.T) {
	const actorDID = "did:test:nested-call-actor-api"
	actor := &mockActorImpl{did: actorDID}
	policy := api.NewPolicyBuilder(api.ContextConfig).
		Allow("tool.agentmodel.register").
		Allow("tool.test.probecontext").
		Allow("tool.aeiou.magic").
		Grant("model:admin:*"). // <-- FIX: Grant required capability
		Build()

	interp, mockProvider := setupPropagationTestAPI(t, actor, policy)

	mockProvider.Callback = func() (string, error) {
		return `
			command
				set report = call_the_probe()
				emit report
				set p = {"action": "done"}
				emit tool.aeiou.magic("LOOP", p)
			endcommand
		`, nil
	}

	script := `
		func call_the_probe() means
			return tool.test.probecontext()
		endfunc
		func main() means
			ask "probe_agent_api", "probe" into result
			return result
		endfunc
	`
	tree, err := api.Parse([]byte(script), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	_, err = api.ExecWithInterpreter(context.Background(), interp, tree)
	if err != nil {
		t.Fatalf("ExecWithInterpreter to load failed: %v", err)
	}

	val, err := api.RunProcedure(context.Background(), interp, "main")
	if err != nil {
		t.Fatalf("RunProcedure(main) failed unexpectedly: %v", err)
	}

	unwrapped, _ := api.Unwrap(val)
	resultStr, ok := unwrapped.(string)
	if !ok {
		t.Fatalf("Expected probe to return a string, but got %T", val)
	}

	var report map[string]interface{}
	if err := json.Unmarshal([]byte(resultStr), &report); err != nil {
		t.Fatalf("Failed to unmarshal JSON from result string: %v", err)
	}

	if report["actor_did"] != actorDID {
		t.Errorf("Probe failed to find Actor DID through nested call. Got: '%v'", report["actor_did"])
	}
	t.Logf("SUCCESS (API): Context correctly propagated through nested function call.")
}

// TestContextPropagation_PolicyInheritance_API verifies that a restrictive policy on the parent
// interpreter is correctly inherited and enforced by the 'ask' loop sandbox using the public API.
func TestContextPropagation_PolicyInheritance_API(t *testing.T) {
	const probeToolName = "tool.test.probecontext"

	policy := api.NewPolicyBuilder(api.ContextConfig).
		Deny(probeToolName).
		Allow("tool.aeiou.magic").
		Allow("tool.agentmodel.register").
		Grant("model:admin:*"). // <-- FIX: Grant required capability
		Build()

	interp, mockProvider := setupPropagationTestAPI(t, nil, policy)
	t.Logf("Applied restrictive policy via API")

	mockProvider.Callback = func() (string, error) {
		return fmt.Sprintf(`
			command
				set report = %s()
				emit report
				set p = {"action": "done"}
				emit tool.aeiou.magic("LOOP", p)
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

	if !errors.Is(err, api.ErrToolDenied) {
		t.Errorf("Expected a policy violation error wrapping api.ErrToolDenied, but got: %v", err)
	} else {
		t.Logf("SUCCESS (API): Correctly received policy violation error: %v", err)
	}
}
