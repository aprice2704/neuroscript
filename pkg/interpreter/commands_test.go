// NeuroScript Version: 0.5.2
// File version: 9.0.0
// Purpose: Refactored to use the centralized TestHarness and corrected the tool name to be fully qualified.
// filename: pkg/interpreter/interpreter_commands_test.go
// nlines: 80
// risk_rating: LOW

package interpreter_test

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/types"
)

func TestCommandExecution(t *testing.T) {
	t.Run("Execute Multiple Commands in Order", func(t *testing.T) {
		h := NewTestHarness(t)
		interp := h.Interpreter
		t.Logf("[DEBUG] Turn 1: Test harness created for command execution test.")

		var callLog []string
		mockToolSpec := tool.ToolSpec{Name: "Record", Group: "TestTool", Args: []tool.ArgSpec{{Name: "arg", Type: "string"}}}
		mockToolFunc := func(_ tool.Runtime, args []interface{}) (interface{}, error) {
			arg, _ := lang.ToString(args[0])
			t.Logf("[DEBUG] Turn X: Mock tool 'Record' called with arg: %s", arg)
			callLog = append(callLog, arg)
			return &lang.NilValue{}, nil
		}
		interp.ToolRegistry().RegisterTool(tool.ToolImplementation{Spec: mockToolSpec, Func: mockToolFunc})
		t.Logf("[DEBUG] Turn 2: Mock tool registered.")

		program := &ast.Program{
			Commands: []*ast.CommandNode{
				{
					Body: []ast.Step{
						{
							Type: "call",
							Call: &ast.CallableExprNode{
								Target:    ast.CallTarget{Name: string(types.MakeFullName("TestTool", "Record")), IsTool: true},
								Arguments: []ast.Expression{&ast.StringLiteralNode{Value: "first"}},
							},
						},
						{
							Type:    "set",
							LValues: []*ast.LValueNode{{Identifier: "my_arg"}},
							Values:  []ast.Expression{&ast.StringLiteralNode{Value: "second"}},
						},
						{
							Type: "call",
							Call: &ast.CallableExprNode{
								Target:    ast.CallTarget{Name: string(types.MakeFullName("TestTool", "Record")), IsTool: true},
								Arguments: []ast.Expression{&ast.VariableNode{Name: "my_arg"}},
							},
						},
					},
				},
			},
		}

		if err := interp.Load(&interfaces.Tree{Root: program}); err != nil {
			t.Fatalf("Load() returned an unexpected error: %v", err)
		}
		t.Logf("[DEBUG] Turn 3: Program with commands loaded.")

		_, err := interp.Execute(program)
		if err != nil {
			t.Fatalf("Execute() failed: %v", err)
		}
		t.Logf("[DEBUG] Turn 4: Execute() completed.")

		if len(callLog) != 2 {
			t.Fatalf("Expected mock tool to be called 2 times, but was called %d times", len(callLog))
		}
		if callLog[0] != "first" {
			t.Errorf("Expected first call argument to be 'first', got '%s'", callLog[0])
		}
		if callLog[1] != "second" {
			t.Errorf("Expected second call argument to be 'second', got '%s'", callLog[1])
		}
		t.Logf("[DEBUG] Turn 5: Assertions passed.")
	})
}
