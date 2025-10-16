// NeuroScript Version: 0.8.0
// File version: 6
// Purpose: Corrects test failure by using the api.NewPolicyBuilder to create a policy that allows the test tool.
// filename: pkg/api/cmd_block_context_test.go
// nlines: 77
// risk_rating: LOW

package api_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// TestExecWithInterpreterToolRuntimeContext verifies that a tool called from a sandboxed
// 'command' block (via ExecWithInterpreter) receives the public, identity-aware
// *api.Interpreter as its runtime, not the internal *interpreter.Interpreter.
func TestExecWithInterpreterToolRuntimeContext(t *testing.T) {
	// 1. Define a tool that checks the type of its runtime.
	runtimeCheckTool := tool.ToolImplementation{
		Spec: tool.ToolSpec{
			Name:  "check_runtime",
			Group: "test",
		},
		Func: func(rt tool.Runtime, args []interface{}) (interface{}, error) {
			if _, ok := rt.(interfaces.ActorProvider); !ok {
				return nil, fmt.Errorf("FATAL: runtime is of type %T, does not implement ActorProvider", rt)
			}
			return "OK", nil
		},
	}

	// 2. Set up the host context and an execution policy that allows the test tool.
	hostCtx := newTestHostContext(nil)

	// FIX: Use the fluent builder from the api package to construct the policy.
	execPolicy := api.NewPolicyBuilder(api.ContextNormal).
		Allow("tool.test.check_runtime").
		Build()

	interp := api.New(
		api.WithHostContext(hostCtx),
		api.WithExecPolicy(execPolicy), // Apply the permissive policy
	)

	_, err := interp.ToolRegistry().RegisterTool(runtimeCheckTool)
	if err != nil {
		t.Fatalf("failed to register tool: %v", err)
	}

	// 3. Define a script that calls the test tool from a command block.
	script := `
        command
            call tool.test.check_runtime()
        endcommand
    `
	tree, err := api.Parse([]byte(script), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("failed to parse script: %v", err)
	}

	// 4. Execute the script's command blocks.
	_, execErr := api.ExecWithInterpreter(context.Background(), interp, tree)

	// 5. Assert that the execution succeeded without any errors.
	if execErr != nil {
		t.Errorf("ExecWithInterpreter failed: expected no error, but got: %v", execErr)
	}
}
