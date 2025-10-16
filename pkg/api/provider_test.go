// NeuroScript Version: 0.8.0
// File version: 14
// Purpose: Corrects the test to provide a mandatory HostContext during interpreter creation, resolving a panic.
// filename: pkg/api/provider_test.go
// nlines: 97
// risk_rating: LOW

package api_test

import (
	"context"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/provider/test"
)

func TestAPI_RegisterAndUseProvider(t *testing.T) {
	providerName := "test_provider"

	// The NeuroScript code to be executed.
	scriptContent := `
func main(returns string) means
    ask "test_agent", "ping" into result
    return result
endfunc
`
	// Create an interpreter with a trusted 'config' context to allow registration.
	// FIX: The mock provider needs permission to call the magic tool.
	configPolicy := &policy.ExecPolicy{
		Context: policy.ContextConfig,
		Allow:   []string{"tool.aeiou.magic"},
	}

	// FIX: A HostContext is now mandatory for creating an interpreter.
	hc := newTestHostContext(nil)
	interp := api.New(
		api.WithHostContext(hc),
		interpreter.WithExecPolicy(configPolicy),
	)

	interp.RegisterProvider(providerName, test.New())

	// Register an AgentModel configured to use our test provider.
	agentConfig := map[string]any{
		"provider": providerName,
		"model":    "default",
	}
	if err := interp.RegisterAgentModel("test_agent", agentConfig); err != nil {
		t.Fatalf("Failed to register agent model: %v", err)
	}

	// Parse and load the script.
	tree, err := api.Parse([]byte(scriptContent), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("api.Parse failed: %v", err)
	}
	if _, err := api.ExecWithInterpreter(context.Background(), interp, tree); err != nil {
		t.Fatalf("api.ExecWithInterpreter failed to load definitions: %v", err)
	}

	// Run the procedure.
	// The 'ask' statement will internally create a V3 envelope with a USERDATA
	// section like: {"subject":"ask","fields":{"prompt":"ping"}}
	result, err := api.RunProcedure(context.Background(), interp, "main")
	if err != nil {
		t.Fatalf("api.RunProcedure() failed: %v", err)
	}

	// Verify the final result.
	unwrapped, err := api.Unwrap(result)
	if err != nil {
		t.Fatalf("api.Unwrap failed: %v", err)
	}

	val, ok := unwrapped.(string)
	if !ok {
		t.Fatalf("Expected a string return type, but got %T", unwrapped)
	}

	// The mock provider is hard-coded to return "test_provider_ok:pong" for a "ping" prompt.
	expectedResponse := "test_provider_ok:pong"
	if !strings.Contains(val, expectedResponse) {
		t.Errorf("Expected response to contain '%s', but got '%s'", expectedResponse, val)
	}
}
