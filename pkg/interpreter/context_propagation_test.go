// NeuroScript Version: 0.8.0
// File version: 10
// Purpose: Corrected test to expect the correct ErrorCodePolicy (1003) instead of the wrapped ErrorCodeInternal (6).
// filename: pkg/interpreter/context_propagation_test.go
// nlines: 318
// risk_rating: LOW

package interpreter_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/capability"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/provider"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/tool/agentmodel"
)

// mockContextProvider is an AI provider that executes a specific callback
// inside its Chat method, simulating the 'ask' loop's action execution.
type mockContextProvider struct {
	t        *testing.T
	Callback func() (string, error)
}

func (m *mockContextProvider) Chat(ctx context.Context, req provider.AIRequest) (*provider.AIResponse, error) {
	actions, err := m.Callback()
	if err != nil {
		return nil, err
	}
	env := &aeiou.Envelope{UserData: "{}", Actions: actions}
	respText, _ := env.Compose()
	return &provider.AIResponse{TextContent: respText}, nil
}

// buildProbeTool creates the tool used by all tests to inspect the runtime context.
func buildProbeTool() tool.ToolImplementation {
	return tool.ToolImplementation{
		Spec: tool.ToolSpec{
			Name:       "ProbeContext",
			Group:      "test",
			ReturnType: tool.ArgTypeMap,
		},
		Func: func(rt tool.Runtime, args []interface{}) (interface{}, error) {
			report := make(map[string]interface{})
			fmt.Printf("[DEBUG] probeTool: Executing probe.\n")

			if provider, ok := rt.(interfaces.ActorProvider); ok {
				if actor, found := provider.Actor(); found {
					report["actor_did"] = actor.DID()
				} else {
					report["actor_did"] = "not_found"
				}
			} else {
				report["actor_did"] = "provider_unsupported"
			}

			if ctxProvider, ok := rt.(interpreter.TurnContextProvider); ok {
				turnCtx := ctxProvider.GetTurnContext()
				fmt.Printf("[DEBUG] probeTool: Received context pointer: %p\n", turnCtx)
				if sid, ok := turnCtx.Value(interpreter.AeiouSessionIDKey).(string); ok {
					report["aeiou_session_id"] = sid
					fmt.Printf("[DEBUG] probeTool: Found session ID: %s\n", sid)
				} else {
					report["aeiou_session_id"] = "not_found"
				}
				if tid, ok := turnCtx.Value(interpreter.AeiouTurnIndexKey).(int); ok {
					report["aeiou_turn_index"] = float64(tid) // JSON unmarshals numbers to float64
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

// setupPropagationTest provides a correctly configured interpreter for all propagation tests.
func setupPropagationTest(t *testing.T) (*interpreter.Interpreter, *mockContextProvider) {
	t.Helper()
	h := NewTestHarness(t)
	interp := h.Interpreter

	if err := tool.CreateRegistrationFunc("agentmodel", agentmodel.AgentModelToolsToRegister)(interp.ToolRegistry()); err != nil {
		t.Fatalf("Failed to register agentmodel toolset: %v", err)
	}
	if _, err := interp.ToolRegistry().RegisterTool(buildProbeTool()); err != nil {
		t.Fatalf("Failed to register probe tool: %v", err)
	}

	mockProv := &mockContextProvider{t: t}
	interp.RegisterProvider("probe_provider", mockProv)

	setupScript := `
    func _SetupProbeAgent() means
        set config = {\
            "provider": "probe_provider",\
            "model": "probe_model",\
            "tool_loop_permitted": true\
        }
        must tool.agentmodel.Register("probe_agent", config)
    endfunc
    `
	tree, pErr := interp.Parser().Parse(setupScript)
	if pErr != nil {
		t.Fatalf("Setup script parse failed: %v", pErr)
	}
	program, _, bErr := interp.ASTBuilder().Build(tree)
	if bErr != nil {
		t.Fatalf("Setup script build failed: %v", bErr)
	}
	if err := interp.Load(&interfaces.Tree{Root: program}); err != nil {
		t.Fatalf("Setup script load failed: %v", err)
	}

	_, err := interp.Run("_SetupProbeAgent")
	if err != nil {
		t.Fatalf("Agent setup procedure failed: %v", err)
	}
	t.Logf("[DEBUG] Agent 'probe_agent' registered via setup script.")

	return interp, mockProv
}

// TestContextPropagation_EndToEnd verifies that critical contexts are correctly passed
// through the ask loop's sandbox into the final tool runtime.
func TestContextPropagation_EndToEnd(t *testing.T) {
	const actorDID = "did:test:context-probe-actor"
	const probeToolName = "tool.test.probeContext"

	interp, mockProvider := setupPropagationTest(t)
	interp.HostContext().Actor = &mockActor{did: actorDID}

	mockProvider.Callback = func() (string, error) {
		// THE FIX: Removed obsolete 'tool.aeiou.magic' call.
		// The loop now terminates via the '<<<LOOP:DONE>>>' signal.
		return fmt.Sprintf(`
			command
				set report = %s()
				emit report
				emit "<<<LOOP:DONE>>>"
			endcommand
		`, probeToolName), nil
	}

	script := `
		func main() means
			ask "probe_agent", "probe" into result
			return result
		endfunc
	`
	// THE FIX: Parse and load the script before running it.
	tree, _ := interp.Parser().Parse(script)
	program, _, _ := interp.ASTBuilder().Build(tree)
	if err := interp.Load(&interfaces.Tree{Root: program}); err != nil {
		t.Fatalf("Failed to load script: %v", err)
	}
	t.Logf("[DEBUG] Main test script loaded.")

	val, err := interp.Run("main")
	if err != nil {
		t.Fatalf("Run(main) failed unexpectedly: %v", err)
	}

	resultStr, ok := val.(lang.StringValue)
	if !ok {
		t.Fatalf("Expected 'ask' to return a string, but got %T", val)
	}
	t.Logf("[DEBUG] Received raw string result from ask: %s", resultStr.Value)

	var report map[string]interface{}
	// The result is now the *emitted output*, which is a JSON string.
	if err := json.Unmarshal([]byte(resultStr.Value), &report); err != nil {
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
	if report["aeiou_turn_nonce"] == "not_found" || report["aeiou_turn_nonce"] == "" {
		t.Errorf("Probe failed to find AEIOU Turn Nonce. Got: '%v'", report["aeiou_turn_nonce"])
	}
}

// TestContextPropagation_NestedCall verifies context propagation through an 'ask' sandbox
// and then through a subsequent nested 'func' call sandbox.
func TestContextPropagation_NestedCall(t *testing.T) {
	interp, mockProvider := setupPropagationTest(t)
	interp.HostContext().Actor = &mockActor{did: "did:test:nested-call-actor"}

	mockProvider.Callback = func() (string, error) {
		// THE FIX: Removed obsolete 'tool.aeiou.magic' call.
		return `
			command
				set report = call_the_probe()
				emit report
				emit "<<<LOOP:DONE>>>"
			endcommand
		`, nil
	}

	script := `
		func call_the_probe() means
			return tool.test.probeContext()
		endfunc
		func main() means
			ask "probe_agent", "probe" into result
			return result
		endfunc
	`
	// THE FIX: Parse and load the script before running it.
	tree, _ := interp.Parser().Parse(script)
	program, _, _ := interp.ASTBuilder().Build(tree)
	if err := interp.Load(&interfaces.Tree{Root: program}); err != nil {
		t.Fatalf("Failed to load script: %v", err)
	}
	t.Logf("[DEBUG] Nested call test script loaded.")

	val, err := interp.Run("main")
	if err != nil {
		t.Fatalf("Run(main) failed unexpectedly: %v", err)
	}

	resultStr, ok := val.(lang.StringValue)
	if !ok {
		t.Fatalf("Expected probe to return a string, but got %T", val)
	}
	t.Logf("[DEBUG] Received raw string result from ask (nested): %s", resultStr.Value)

	var report map[string]interface{}
	if err := json.Unmarshal([]byte(resultStr.Value), &report); err != nil {
		t.Fatalf("Failed to unmarshal JSON from result string: %v", err)
	}

	if report["actor_did"] != "did:test:nested-call-actor" {
		t.Errorf("Probe failed to find Actor DID through nested call. Got: '%v'", report["actor_did"])
	}
	t.Logf("SUCCESS: Context correctly propagated through nested function call.")
}

// TestContextPropagation_PolicyInheritance verifies that a restrictive policy on the parent
// interpreter is correctly inherited and enforced by the 'ask' loop sandbox.
func TestContextPropagation_PolicyInheritance(t *testing.T) {
	const probeToolName = "tool.test.probeContext"
	interp, mockProvider := setupPropagationTest(t)

	// Setup a restrictive policy denying the probe tool.
	denyRule := capability.New("tool", "exec", probeToolName)
	p := policy.NewBuilder(policy.ContextNormal).Deny(denyRule.String()).Build()
	interp.ExecPolicy = p
	t.Logf("Applied restrictive policy: Deny '%s'", denyRule.String())

	// This callback now attempts to call the DENIED tool.
	mockProvider.Callback = func() (string, error) {
		// THE FIX: Removed obsolete 'tool.aeiou.magic' call.
		return fmt.Sprintf(`
			command
				set report = %s()
				emit report
				emit "<<<LOOP:DONE>>>"
			endcommand
		`, probeToolName), nil
	}

	script := `func main() means ask "probe_agent", "probe" into result; return result endfunc`
	tree, _ := interp.Parser().Parse(script)
	program, _, _ := interp.ASTBuilder().Build(tree)
	interp.Load(&interfaces.Tree{Root: program})
	_, err := interp.Run("main")

	if err == nil {
		t.Fatal("Execution should have failed due to policy violation, but it succeeded.")
	}
	fmt.Printf("[DEBUG] Policy inheritance test received error: %v (Type: %T)\n", err, err) // DEBUG

	var rtErr *lang.RuntimeError
	if errors.As(err, &rtErr) {
		// THE FIX: With the incorrect error wrapping in steps_ask_aeiou.go removed,
		// we should now receive the correct ErrorCodePolicy (1003).
		if rtErr.Code != lang.ErrorCodePolicy {
			t.Errorf("Expected error code to be ErrorCodePolicy (%d), but got %d. Message: %s", lang.ErrorCodePolicy, rtErr.Code, rtErr.Message)
		} else {
			t.Logf("SUCCESS: Correctly received policy violation error (code 1003): %v", err)
		}
	} else {
		t.Errorf("Expected a *lang.RuntimeError, but got %T: %v", err, err)
	}
}
