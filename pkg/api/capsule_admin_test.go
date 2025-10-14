// NeuroScript Version: 0.8.0
// File version: 8
// Purpose: Corrects a panic by providing a mandatory HostContext when creating the runtime interpreter.
// filename: pkg/api/capsule_admin_test.go
// nlines: 91
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
    set s = "::id: capsule/host-persisted-prompt\n"
	set s = s + "::version: 1\n"
    set s = s + "::serialization: ns\n"
    set s = s + "::description: A test capsule.\n"
    set s = s + "This prompt was added via an admin registry.\n"
    must tool.capsule.add(s)
endcommand
`

	// 3. Create a privileged policy for the config interpreter.
	allowedTools := []string{"tool.capsule.add"}
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
	runtimePolicy := api.NewPolicyBuilder(api.ContextNormal).
		Allow("tool.capsule.getlatest").
		Build()

	runtimeInterp := api.New(
		api.WithHostContext(newTestHostContext(nil)), // <-- FIX: Add mandatory HostContext.
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
