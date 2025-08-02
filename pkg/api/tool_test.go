// NeuroScript Version: 0.6.0
// File version: 6
// Purpose: Reverted test to use the now-fixed WithTool function.
// filename: pkg/api/tool_test.go
// nlines: 93
// risk_rating: LOW

package api_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
)

// TestAPI_BuiltinToolExecution verifies that a standard, built-in tool can be
// successfully called via the public API, following the standard user workflow.
func TestAPI_BuiltinToolExecution(t *testing.T) {
	// 1. The NeuroScript code to be executed.
	scriptContent := `
func do_math(returns number) means
    return tool.math.Add(10, 32)
endfunc
`
	// 2. Create a standard interpreter. The standard tools are auto-bundled.
	interp := api.New()

	// 3. Parse and load the script.
	tree, err := api.Parse([]byte(scriptContent), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("api.Parse failed: %v", err)
	}
	if _, err := api.ExecWithInterpreter(context.Background(), interp, tree); err != nil {
		t.Fatalf("api.ExecWithInterpreter failed to load definitions: %v", err)
	}

	// 4. Run the procedure and assert that it succeeds.
	result, err := api.RunProcedure(context.Background(), interp, "do_math")
	if err != nil {
		t.Fatalf("api.RunProcedure failed unexpectedly: %v", err)
	}

	// 5. If it passes, verify the result.
	unwrapped, _ := api.Unwrap(result)
	if val, ok := unwrapped.(float64); !ok || val != 42.0 {
		t.Errorf("Expected result to be 42.0, but got %v (type %T)", unwrapped, unwrapped)
	}
}

// TestAPI_CustomToolWithDottedGroup is the critical test. It verifies that a
// custom-defined tool with a dot in its group name can be registered and
// resolved correctly through the public API. This directly tests the fix
// for the interpreter bug.
func TestAPI_CustomToolWithDottedGroup(t *testing.T) {
	// 1. Define the custom tool's Go implementation.
	echoToolFunc := func(rt api.Runtime, args []any) (any, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("echo tool expects exactly 1 argument")
		}
		return args[0], nil
	}

	// 2. Define the tool's implementation struct using only public API types.
	echoToolImpl := api.ToolImplementation{
		Spec: api.ToolSpec{
			Name:  "bleat",
			Group: "xx",
			Args: []api.ArgSpec{
				{Name: "value", Type: "any", Required: true},
			},
			ReturnType: "any",
		},
		Func: echoToolFunc,
	}

	// 3. Create an interpreter and register the custom tool using the now-fixed WithTool.
	interp := api.New(api.WithTool(echoToolImpl))

	// 4. The script that calls the custom tool.
	scriptContent := `
func do_echo(returns string) means
    return tool.xx.bleat("hello, world")
endfunc
`
	// 5. Parse and load the script.
	tree, err := api.Parse([]byte(scriptContent), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("api.Parse failed: %v", err)
	}
	if _, err := api.ExecWithInterpreter(context.Background(), interp, tree); err != nil {
		t.Fatalf("api.ExecWithInterpreter failed to load definitions: %v", err)
	}

	// 6. Run the procedure and assert that it succeeds.
	result, err := api.RunProcedure(context.Background(), interp, "do_echo")
	if err != nil {
		t.Fatalf("api.RunProcedure failed for custom tool unexpectedly: %v", err)
	}

	// 7. Unwrap and verify the result.
	unwrapped, _ := api.Unwrap(result)
	if val, ok := unwrapped.(string); !ok || val != "hello, world" {
		t.Errorf("Expected result to be 'hello, world', but got %v (type %T)", unwrapped, unwrapped)
	}

	t.Log("Successfully registered and called custom tool with dotted group name.")
}
