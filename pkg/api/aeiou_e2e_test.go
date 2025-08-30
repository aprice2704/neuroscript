// NeuroScript Version: 0.7.0
// File version: 3
// Purpose: Corrects script syntax from 'set ..., err =' to the valid 'set ... ='.
// filename: pkg/api/aeiou_e2e_test.go
// nlines: 75
// risk_rating: LOW

package api_test

import (
	"context"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/policy"
)

// TestE2E_AeiouWorkflow verifies that the standard aeiou tool workflow
// can be successfully executed via the public API.
func TestE2E_AeiouWorkflow(t *testing.T) {
	// 1. The NeuroScript code to be executed.
	scriptContent := `
func main(returns string) means
    # Create a new envelope
    set h = tool.aeiou.new()

    # Set some content
    call tool.aeiou.set_section(h, "ACTIONS", "command { emit 'ok' }")
    call tool.aeiou.set_section(h, "ORCHESTRATION", "test prompt")

    # Compose it to a string
    set payload = tool.aeiou.compose(h)

    # Parse it back into a new handle
    set h2 = tool.aeiou.parse(payload)

    # Retrieve and return a section from the new handle to verify
    set result = tool.aeiou.get_section(h2, "ORCHESTRATION")

    return result
endfunc
`
	// 2. Create an interpreter with a policy that explicitly allows the aeiou tools.
	// Since these tools are now standard, api.New() will have them registered.
	policy := &policy.ExecPolicy{
		Context: policy.ContextNormal,
		Allow: []string{
			"tool.aeiou.new",
			"tool.aeiou.set_section",
			"tool.aeiou.get_section",
			"tool.aeiou.compose",
			"tool.aeiou.parse",
		},
	}
	interp := api.New(interpreter.WithExecPolicy(policy))

	// 3. Parse and load the script.
	tree, err := api.Parse([]byte(scriptContent), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("api.Parse failed: %v", err)
	}
	if _, err := api.ExecWithInterpreter(context.Background(), interp, tree); err != nil {
		t.Fatalf("api.ExecWithInterpreter failed to load definitions: %v", err)
	}

	// 4. Run the procedure.
	result, err := api.RunProcedure(context.Background(), interp, "main")
	if err != nil {
		t.Fatalf("api.RunProcedure failed unexpectedly: %v", err)
	}

	// 5. Verify the final result.
	unwrapped, _ := api.Unwrap(result)
	if val, ok := unwrapped.(string); !ok || val != "test prompt" {
		t.Errorf("Expected result to be 'test prompt', but got %v (type %T)", unwrapped, unwrapped)
	}
}
