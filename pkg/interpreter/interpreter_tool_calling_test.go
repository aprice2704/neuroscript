// NeuroScript Version: 0.5.2
// File version: 3.0.0
// Purpose: Corrected call to the renamed test helper function 'NewTestInterpreter'.
// filename: pkg/interpreter/interpreter_tool_calling_test.go

package interpreter

import (
	"fmt"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/types"
)

func TestToolToToolCalling_WithDottedGroup(t *testing.T) {
	// 1. Define the "callee" tool with a dotted group name. This is the tool we want to call.
	calleeTool := tool.ToolImplementation{
		Spec: tool.ToolSpec{
			Name:        "Greet",
			Group:       "testing.callee.with.dots",
			Description: "A simple tool to be called by another tool.",
			Args: []tool.ArgSpec{
				{Name: "name", Type: "string", Required: true},
			},
			ReturnType: "string",
		},
		Func: func(rt tool.Runtime, args []interface{}) (interface{}, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("callee expects at least one argument")
			}
			name, ok := args[0].(string)
			if !ok {
				return nil, fmt.Errorf("callee expects the first argument to be a string, got %T", args[0])
			}
			return "Hello, " + name, nil
		},
	}

	// 2. Define the "caller" tool. This tool will use the runtime to call the callee.
	callerTool := tool.ToolImplementation{
		Spec: tool.ToolSpec{
			Name:        "DoCall",
			Group:       "testing.caller",
			Description: "A tool that calls another tool.",
			Args: []tool.ArgSpec{
				{Name: "targetTool", Type: "string", Required: true},
				{Name: "targetArg", Type: "string", Required: true},
			},
			ReturnType: "any",
		},
		Func: func(rt tool.Runtime, args []interface{}) (interface{}, error) {
			if len(args) < 2 {
				return nil, fmt.Errorf("caller expects two arguments")
			}
			targetToolName, _ := args[0].(string)
			targetToolArg, _ := args[1].(string)

			// This is the core of the test: using the runtime to call another tool.
			// The `CallTool` method is responsible for correctly resolving the dotted name.
			return rt.CallTool(types.FullName(targetToolName), []interface{}{targetToolArg})
		},
	}

	// 3. Setup the interpreter and register the tools.
	interp, err := NewTestInterpreter(t, nil, nil, false)
	if err != nil {
		t.Fatalf("Failed to create test interpreter: %v", err)
	}

	if _, err := interp.ToolRegistry().RegisterTool(calleeTool); err != nil {
		t.Fatalf("Failed to register callee tool: %v", err)
	}
	if _, err := interp.ToolRegistry().RegisterTool(callerTool); err != nil {
		t.Fatalf("Failed to register caller tool: %v", err)
	}

	// 4. Execute the caller tool, telling it to call the callee.
	// The key for the registry is the full "tool.group.name".
	callerToolKey := types.MakeFullNameTyped(callerTool.Spec.Group, callerTool.Spec.Name)
	// FIX: The name we pass to the caller is the group-qualified name (group.name),
	// which is what the `CallTool` runtime method expects.
	calleeToolNameForArg := types.FullName(string(calleeTool.Spec.Group) + "." + string(calleeTool.Spec.Name))

	argsForCaller := map[string]lang.Value{
		"targetTool": lang.StringValue{Value: string(calleeToolNameForArg)},
		"targetArg":  lang.StringValue{Value: "World"},
	}

	result, execErr := interp.ExecuteTool(callerToolKey, argsForCaller)
	if execErr != nil {
		t.Fatalf("ExecuteTool failed with an unexpected error: %v", execErr)
	}

	// 5. Verify the result.
	expected := "Hello, World"
	got, _ := lang.ToString(result)
	if got != expected {
		t.Errorf("Tool-to-tool call returned an incorrect value. \n  Expected: %q\n       Got: %q", expected, got)
	}

	t.Log("Successfully verified that a tool can call another tool with a dotted group name.")
}
