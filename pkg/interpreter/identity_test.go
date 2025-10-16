// NeuroScript Version: 0.8.0
// File version: 2
// Purpose: Verifies that a tool can access an actor's identity via the HostContext. Corrected error handling to use errors.As.
// filename: pkg/interpreter/identity_test.go
// nlines: 82
// risk_rating: LOW

package interpreter_test

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// mockActor is a concrete implementation of the Actor interface for testing.
type mockActor struct {
	did string
}

func (m *mockActor) DID() string { return m.did }

// Statically assert that mockActor implements the required interface.
var _ interfaces.Actor = (*mockActor)(nil)

// TestIdentity_ViaHostContext verifies that a tool can access an actor's
// identity when it is provided via the HostContext at interpreter creation time.
func TestIdentity_ViaHostContext(t *testing.T) {
	// --- ARRANGE ---
	t.Logf("[DEBUG] Turn 1: Starting TestIdentity_ViaHostContext.")

	// 1. A tool that requires identity.
	identityAwareTool := tool.ToolImplementation{
		Spec: tool.ToolSpec{
			Name:       "GetActorDID",
			Group:      "test",
			ReturnType: tool.ArgTypeString,
		},
		Func: func(rt tool.Runtime, args []interface{}) (interface{}, error) {
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
		set my_id = tool.test.GetActorDID()
		emit my_id
	endcommand
	`
	policy := policy.NewBuilder(policy.ContextNormal).Allow("tool.test.GetActorDID").Build()
	var stdout bytes.Buffer

	// 2. Create the actor identity.
	expectedDID := "did:test:host-context-actor"
	actor := &mockActor{did: expectedDID}

	// 3. Create a HostContext containing the actor identity using the builder.
	harness := NewTestHarness(t)
	hc, err := interpreter.NewHostContextBuilder().
		WithLogger(harness.Logger).
		WithStdout(&stdout).
		WithStdin(&bytes.Buffer{}).
		WithStderr(&bytes.Buffer{}).
		WithActor(actor). // <-- Set identity directly on the builder.
		Build()
	if err != nil {
		t.Fatalf("Failed to build host context: %v", err)
	}
	t.Logf("[DEBUG] Turn 2: HostContext with actor created.")

	// 4. Create an interpreter with the identity-aware context.
	interp := interpreter.NewInterpreter(
		interpreter.WithHostContext(hc),
		interpreter.WithExecPolicy(policy),
	)
	if _, err := interp.ToolRegistry().RegisterTool(identityAwareTool); err != nil {
		t.Fatalf("Failed to register tool: %v", err)
	}
	t.Logf("[DEBUG] Turn 3: Interpreter and tool created.")

	// --- ACT ---
	tree, err := interp.Parser().Parse(script)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	program, _, err := interp.ASTBuilder().Build(tree)
	if err != nil {
		t.Fatalf("AST build failed: %v", err)
	}
	if err := interp.Load(&interfaces.Tree{Root: program}); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	_, err = interp.Execute(program)
	t.Logf("[DEBUG] Turn 4: Script executed.")

	// --- ASSERT ---
	if err != nil {
		var rtErr *lang.RuntimeError
		if errors.As(err, &rtErr) {
			t.Fatalf("Execute failed: %v\nUnderlying error: %v", err, rtErr.Unwrap())
		}
		t.Fatalf("ExecWithInterpreter failed: %v", err)
	}

	output := strings.TrimSpace(stdout.String())
	if output != expectedDID {
		t.Errorf("Expected stdout to contain '%s', but got '%s'", expectedDID, output)
	}

	t.Logf("[DEBUG] Turn 5: SUCCESS: Correct DID was retrieved by the tool.")
}
