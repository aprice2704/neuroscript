// NeuroScript Version: 0.8.0
// File version: 16
// Purpose: FIX: Removed invalid type assertion on the concrete factory type.
// filename: pkg/api/provider_test.go
// nlines: 90
// risk_rating: HIGH

package api

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ax"
	"github.com/aprice2704/neuroscript/pkg/provider/test"
	"github.com/google/uuid"
)

func TestAPI_RegisterAndUseProvider(t *testing.T) {
	ctx := context.Background()
	providerName := "test_provider"

	// --- Phase 1: Factory and Provider Setup ---
	factory, err := NewAXFactory(ctx, ax.RunnerOpts{}, &mockRuntime{}, &mockID{did: "did:test:host"})
	if err != nil {
		t.Fatalf("NewAXFactory() failed: %v", err)
	}

	// In a test, we can access the internal root interpreter to register a mock provider.
	factory.root.RegisterProvider(providerName, test.New())

	// --- Phase 2: Configuration via a Config Runner ---
	configScript := fmt.Sprintf(`
	command
		must tool.agentmodel.register("test_agent", {
			"provider": "%s",
			"model":    "default"
		})
	endcommand`, providerName)

	configRunner, err := factory.NewRunner(ctx, ax.RunnerConfig, ax.RunnerOpts{})
	if err != nil {
		t.Fatalf("NewRunner(Config) failed: %v", err)
	}
	if err := configRunner.LoadScript([]byte(configScript)); err != nil {
		t.Fatalf("LoadScript(Config) failed: %v", err)
	}
	if _, err := configRunner.Execute(); err != nil {
		t.Fatalf("Execute(Config) failed: %v", err)
	}

	// --- Phase 3: Execution via a User Runner ---
	scriptContent := `
func main(returns string) means
    ask "test_agent", "ping" into result
    return result
endfunc
`
	userRunner, err := factory.NewRunner(ctx, ax.RunnerUser, ax.RunnerOpts{})
	if err != nil {
		t.Fatalf("NewRunner(User) failed: %v", err)
	}

	// Run the procedure, passing the required AEIOU turn context.
	turnCtx := ContextWithSessionID(context.Background(), uuid.NewString())
	result, err := AXRunScript(turnCtx, userRunner, []byte(scriptContent), "main")
	if err != nil {
		t.Fatalf("AXRunScript() failed: %v", err)
	}

	// Verify the final result.
	val, ok := result.(string)
	if !ok {
		t.Fatalf("Expected a string return type, but got %T", result)
	}

	// The mock provider is hard-coded to return "test_provider_ok:pong" for a "ping" prompt.
	expectedResponse := "test_provider_ok:pong"
	if !strings.Contains(val, expectedResponse) {
		t.Errorf("Expected response to contain '%s', but got '%s'", expectedResponse, val)
	}
}
