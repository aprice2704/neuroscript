// NeuroScript Version: 0.5.2
// File version: 4.0.0
// Purpose: Refactored to use the centralized TestHarness for robust and consistent interpreter initialization.
// filename: pkg/interpreter/interpreter_tool_calling_test.go

package interpreter_test

import (
	"fmt"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/types"
)

func TestToolToToolCalling_WithDottedGroup(t *testing.T) {
	t.Logf("[DEBUG] Turn 1: Starting TestToolToToolCalling_WithDottedGroup.")
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
			t.Logf("[DEBUG] Turn X: Callee tool 'Greet' executed.")
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
			t.Logf("[DEBUG] Turn X: Caller tool 'DoCall' executed.")
			if len(args) < 2 {
				return nil, fmt.Errorf("caller expects two arguments")
			}
			targetToolName, _ := args[0].(string)
			targetToolArg, _ := args[1].(string)

			return rt.CallTool(types.FullName(targetToolName), []interface{}{targetToolArg})
		},
	}

	h := NewTestHarness(t)
	interp := h.Interpreter
	t.Logf("[DEBUG] Turn 2: Test harness created.")

	if _, err := interp.ToolRegistry().RegisterTool(calleeTool); err != nil {
		t.Fatalf("Failed to register callee tool: %v", err)
	}
	if _, err := interp.ToolRegistry().RegisterTool(callerTool); err != nil {
		t.Fatalf("Failed to register caller tool: %v", err)
	}
	t.Logf("[DEBUG] Turn 3: Caller and callee tools registered.")

	callerToolKey := types.MakeFullNameTyped(callerTool.Spec.Group, callerTool.Spec.Name)
	calleeToolNameForArg := types.FullName(string(calleeTool.Spec.Group) + "." + string(calleeTool.Spec.Name))

	argsForCaller := map[string]lang.Value{
		"targetTool": lang.StringValue{Value: string(calleeToolNameForArg)},
		"targetArg":  lang.StringValue{Value: "World"},
	}

	t.Logf("[DEBUG] Turn 4: Executing caller tool.")
	result, execErr := interp.ExecuteTool(callerToolKey, argsForCaller)
	if execErr != nil {
		t.Fatalf("ExecuteTool failed with an unexpected error: %v", execErr)
	}
	t.Logf("[DEBUG] Turn 5: Execution complete.")

	expected := "Hello, World"
	got, _ := lang.ToString(result)
	if got != expected {
		t.Errorf("Tool-to-tool call returned an incorrect value. \n  Expected: %q\n       Got: %q", expected, got)
	}
	t.Logf("[DEBUG] Turn 6: Assertion passed.")
}
