// NeuroScript Version: 0.7.0
// File version: 7
// Purpose: Fixes incorrect error handling for the api.Unwrap function call.
// filename: pkg/api/autoprovider_test.go
// nlines: 93
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

// TestAPI_AutoProviderRegistration verifies that a provider registered via the
// top-level API function is correctly configured and accessible to scripts.
func TestAPI_AutoProviderRegistration(t *testing.T) {
	// 1. Create an interpreter instance.
	interp := api.New()

	// 2. Register the mock provider with a specific name on the interpreter.
	interp.RegisterProvider("mock", test.New())

	// 3. Define a script that uses this specific provider.
	scriptContent := `
func main(returns string) means
    # Create an envelope for the prompt
    set h = tool.aeiou.new()
    call tool.aeiou.set_section(h, "ORCHESTRATION", "What is a large language model?")
    set payload = tool.aeiou.compose(h)

    # Call the model with the composed envelope
    set result = tool.model.chat("mock", "default", payload)
    return result
endfunc
`
	// 4. Configure a policy to allow the necessary tools.
	policy := &policy.ExecPolicy{
		Context: policy.ContextNormal,
		Allow: []string{
			"tool.model.chat",
			"tool.aeiou.new",
			"tool.aeiou.set_section",
			"tool.aeiou.compose",
		},
	}
	// Apply the policy to the interpreter instance.
	interp = api.New(interpreter.WithExecPolicy(policy))
	interp.RegisterProvider("mock", test.New()) // Re-register after creating new interp with policy

	// 5. Parse and load the script.
	tree, err := api.Parse([]byte(scriptContent), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("api.Parse failed: %v", err)
	}
	if _, err := api.ExecWithInterpreter(context.Background(), interp, tree); err != nil {
		t.Fatalf("api.ExecWithInterpreter failed: %v", err)
	}

	// 6. Run the procedure.
	result, err := api.RunProcedure(context.Background(), interp, "main")
	if err != nil {
		t.Fatalf("api.RunProcedure failed unexpectedly: %v", err)
	}

	// 7. Verify the result from the mock provider.
	unwrapped, err := api.Unwrap(result)
	if err != nil {
		t.Fatalf("api.Unwrap failed: %v", err)
	}
	val, ok := unwrapped.(string)
	if !ok {
		t.Fatalf("Expected a string return type, but got %T", unwrapped)
	}

	expectedResponse := "A large language model is a neural network."
	if !strings.Contains(val, expectedResponse) {
		t.Errorf("Expected response to contain '%s', but got: '%s'", expectedResponse, val)
	}
}
