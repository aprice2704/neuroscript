// NeuroScript Version: 0.8.0
// File version: 18
// Purpose: FIX: Removed invalid type assertion on the concrete factory type. Changed package to `api` for test helpers.
// filename: pkg/api/autoprovider_test.go
// nlines: 86
// risk_rating: HIGH

package api

import (
	"context"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ax"
	"github.com/aprice2704/neuroscript/pkg/provider/test"
	"github.com/google/uuid"
)

func TestAPI_AutoProviderRegistration(t *testing.T) {
	ctx := context.Background()

	// --- Phase 1: Factory and Provider Setup ---
	factory, err := NewAXFactory(ctx, ax.RunnerOpts{}, &mockRuntime{}, &mockID{did: "did:test:host"})
	if err != nil {
		t.Fatalf("NewAXFactory() failed: %v", err)
	}

	factory.root.RegisterProvider("mock", test.New())

	// --- Phase 2: Configuration via a Config Runner ---
	configScript := `
	command
		must tool.agentmodel.register("test_agent", {
			"provider": "mock",
			"model": "test-model"
		})
	endcommand
	`
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
    ask "test_agent", "What is a large language model?" into result
    return result
endfunc
`
	userRunner, err := factory.NewRunner(ctx, ax.RunnerUser, ax.RunnerOpts{})
	if err != nil {
		t.Fatalf("NewRunner(User) failed: %v", err)
	}

	turnCtx := ContextWithSessionID(context.Background(), uuid.NewString())
	result, err := AXRunScript(turnCtx, userRunner, []byte(scriptContent), "main")
	if err != nil {
		t.Fatalf("AXRunScript() failed unexpectedly: %v", err)
	}

	val, ok := result.(string)
	if !ok {
		t.Fatalf("Expected a string return type, but got %T", result)
	}

	expectedResponse := "large language model"
	if !strings.Contains(val, expectedResponse) {
		t.Errorf("Expected response to contain '%s', but got: '%s'", expectedResponse, val)
	}
}
