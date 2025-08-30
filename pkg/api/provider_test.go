// NeuroScript Version: 0.7.0
// File version: 6
// Purpose: Fixes incorrect error handling for the api.Unwrap function call.
// filename: pkg/api/provider_test.go
// nlines: 87
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
	// 1. Manually register a provider instance with the interpreter.
	providerName := "test_provider"
	interp := api.New()
	interp.RegisterProvider(providerName, test.New())

	// 2. The NeuroScript code to be executed.
	scriptContent := `
func main(returns string) means
    # Create an envelope for the prompt
    set h = tool.aeiou.new()
    call tool.aeiou.set_section(h, "ORCHESTRATION", "ping")
    set payload = tool.aeiou.compose(h)

    # Call the model and store the result
    set result = tool.model.chat("test_provider", "default", payload)
    return result
endfunc
`
	// 3. Create an interpreter with a policy that allows the necessary tools.
	policy := &policy.ExecPolicy{
		Context: policy.ContextNormal,
		Allow: []string{
			"tool.model.chat",
			"tool.aeiou.new",
			"tool.aeiou.set_section",
			"tool.aeiou.compose",
		},
	}

	// We must re-create the interpreter with the policy.
	interp = api.New(interpreter.WithExecPolicy(policy))
	interp.RegisterProvider(providerName, test.New())

	// 4. Parse and load the script.
	tree, err := api.Parse([]byte(scriptContent), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("api.Parse failed: %v", err)
	}
	if _, err := api.ExecWithInterpreter(context.Background(), interp, tree); err != nil {
		t.Fatalf("api.ExecWithInterpreter failed to load definitions: %v", err)
	}

	// 5. Run the procedure.
	result, err := api.RunProcedure(context.Background(), interp, "main")
	if err != nil {
		t.Fatalf("api.RunProcedure() failed: %v", err)
	}

	// 6. Verify the final result.
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
