// NeuroScript Version: 0.7.2
// File version: 3
// Purpose: Corrects the test by adding a policy to the runtime interpreter that allows the capsule tool to run.
// filename: pkg/api/capsule_admin_test.go
// nlines: 88
// risk_rating: MEDIUM

package api_test

import (
	"context"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
)

// TestAdminCapsuleRegistry_PersistencePattern verifies the full, two-phase lifecycle
// of a host-managed capsule registry, using the newly exposed explicit types.
func TestAdminCapsuleRegistry_PersistencePattern(t *testing.T) {
	// --- Phase 1: Trusted Configuration ---

	// 1. The host application (e.g., FDM) creates and owns the admin registry.
	liveAdminRegistry := api.NewAdminCapsuleRegistry()
	if liveAdminRegistry == nil {
		t.Fatal("NewAdminCapsuleRegistry() returned nil")
	}

	// 2. Define the trusted script that will add a new capsule.
	configScript := `
command
    must tool.capsule.Add({\
        "name": "capsule/host-persisted-prompt",\
        "version": "1.1",\
        "content": "This prompt was added via an admin registry."\
    })
endcommand
`
	// 3. Create a privileged policy for the config interpreter.
	allowedTools := []string{"tool.capsule.Add"}
	requiredGrants := []api.Capability{
		api.NewCapability(api.ResCapsule, api.VerbWrite, "*"),
	}

	// 4. Create a special config interpreter, injecting the LIVE admin registry.
	configInterp := api.NewConfigInterpreter(
		allowedTools,
		requiredGrants,
		api.WithCapsuleAdminRegistry(liveAdminRegistry), // <-- Give it write access
	)

	// 5. Run the script to populate the liveAdminRegistry.
	tree, err := api.Parse([]byte(configScript), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("Phase 1: api.Parse() failed: %v", err)
	}
	_, err = api.ExecWithInterpreter(context.Background(), configInterp, tree)
	if err != nil {
		t.Fatalf("Phase 1: api.ExecWithInterpreter() failed: %v", err)
	}

	// --- Phase 2: Unprivileged Runtime ---

	// 1. Define a normal, unprivileged script that reads the capsule.
	runtimeScript := `
func main(returns string) means
    set my_cap = tool.capsule.GetLatest("capsule/host-persisted-prompt")
    return my_cap["content"]
endfunc
`
	// 2. Create a standard, unprivileged interpreter.
	// FIX: Add a policy that allows the 'GetLatest' tool to run.
	runtimePolicy := api.NewPolicyBuilder(api.ContextNormal).
		Allow("tool.capsule.GetLatest").
		Build()

	runtimeInterp := api.New(
		// 3. Add the populated registry as a new, read-only layer.
		api.WithCapsuleRegistry(liveAdminRegistry),
		api.WithExecPolicy(runtimePolicy),
	)

	// 4. Load and run the script.
	tree, _ = api.Parse([]byte(runtimeScript), api.ParseSkipComments)
	api.ExecWithInterpreter(context.Background(), runtimeInterp, tree)

	result, err := api.RunProcedure(context.Background(), runtimeInterp, "main")
	if err != nil {
		t.Fatalf("Phase 2: api.RunProcedure() failed: %v", err)
	}

	// 5. Verify the result.
	unwrapped, _ := api.Unwrap(result)
	content, ok := unwrapped.(string)
	if !ok {
		t.Fatalf("Expected a string result, but got %T", unwrapped)
	}

	expectedContent := "This prompt was added via an admin registry."
	if !strings.Contains(content, expectedContent) {
		t.Errorf("Read incorrect capsule content.\n  Expected to contain: %q\n  Got: %q", expectedContent, content)
	}
}
