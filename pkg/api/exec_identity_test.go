// NeuroScript Version: 0.8.0
// File version: 6
// Purpose: Corrects the identity test to use the canonical HostContextBuilder pattern for injecting identity.
// filename: pkg/api/exec_identity_test.go
// nlines: 85
// risk_rating: LOW

package api_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
)

// mockActorImpl is a concrete implementation of the Actor interface for testing.
type mockActorImpl struct {
	did string
}

func (m *mockActorImpl) DID() string { return m.did }

// Statically assert that mockActorImpl implements the required interface.
var _ api.Actor = (*mockActorImpl)(nil)

// TestExec_IdentityViaHostContext verifies that a tool can access an actor's
// identity when it is provided via the HostContext at interpreter creation time.
func TestExec_IdentityViaHostContext(t *testing.T) {
	// --- ARRANGE ---

	// 1. A tool that requires identity.
	identityAwareTool := api.ToolImplementation{
		Spec: api.ToolSpec{
			Name:       "GetActorDID",
			Group:      "test",
			ReturnType: "string",
		},
		Func: func(rt api.Runtime, args []any) (any, error) {
			provider, ok := rt.(interfaces.ActorProvider)
			if !ok {
				return nil, fmt.Errorf("runtime does not provide an actor")
			}
			actor, found := provider.Actor()
			if !found {
				return "actor_not_found", nil
			}
			return actor.DID(), nil
		},
	}

	script := `
	command
		set my_id = tool.test.getactordid()
		emit my_id
	endcommand
	`
	policy := api.NewPolicyBuilder(api.ContextNormal).Allow("tool.test.getactordid").Build()
	var stdout bytes.Buffer

	// 2. Create the actor identity.
	expectedDID := "did:test:host-context-agent"
	actor := &mockActorImpl{did: expectedDID}

	// 3. Create a HostContext containing the actor identity using the builder.
	hc, err := api.NewHostContextBuilder().
		WithLogger(&mockLogger{}).
		WithStdout(&stdout).
		WithStdin(os.Stdin).
		WithStderr(os.Stderr).
		WithActor(actor). // <-- Set identity directly on the builder.
		Build()
	if err != nil {
		t.Fatalf("Failed to build host context: %v", err)
	}

	// 4. Create a standard interpreter. It becomes identity-aware via its context.
	interp := api.New(
		api.WithHostContext(hc),
		api.WithExecPolicy(policy),
	)
	if _, err := interp.ToolRegistry().RegisterTool(identityAwareTool); err != nil {
		t.Fatalf("Failed to register tool: %v", err)
	}

	// --- ACT ---
	tree, err := api.Parse([]byte(script), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	_, err = api.ExecWithInterpreter(context.Background(), interp, tree)

	// --- ASSERT ---
	if err != nil {
		t.Fatalf("ExecWithInterpreter failed: %v", err)
	}

	output := strings.TrimSpace(stdout.String())
	if output != expectedDID {
		t.Errorf("Expected stdout to contain '%s', but got '%s'", expectedDID, output)
	}

	t.Log("SUCCESS: ExecWithInterpreter correctly executed an identity-aware tool using HostContext.")
}
