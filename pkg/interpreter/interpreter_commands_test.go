// NeuroScript Version: 0.5.2
// File version: 8.0.1
// Purpose: Corrected the call to interp.Load to pass the correct AST structure.
// filename: pkg/interpreter/interpreter_commands_test.go
// nlines: 80
// risk_rating: LOW

package interpreter

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

func TestCommandExecution(t *testing.T) {
	t.Run("Execute Multiple Commands in Order", func(t *testing.T) {
		interp, _ := NewTestInterpreter(t, nil, nil, false)

		// Register a mock tool for the commands to call.
		var callLog []string
		mockToolSpec := tool.ToolSpec{Name: "Record", Group: "TestTool", Args: []tool.ArgSpec{{Name: "arg", Type: "string"}}}
		mockToolFunc := func(_ tool.Runtime, args []interface{}) (interface{}, error) {
			arg, _ := lang.ToString(args[0])
			callLog = append(callLog, arg)
			return &lang.NilValue{}, nil
		}
		interp.ToolRegistry().RegisterTool(tool.ToolImplementation{Spec: mockToolSpec, Func: mockToolFunc})

		// Manually build the AST to simulate a parsed command script.
		program := &ast.Program{
			Commands: []*ast.CommandNode{
				{
					Body: []ast.Step{
						{
							Type: "call",
							Call: &ast.CallableExprNode{
								// FIX: Use the correct fully-qualified name for the tool.
								Target:    ast.CallTarget{Name: "tool.TestTool.Record", IsTool: true},
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
								// FIX: Use the correct fully-qualified name for the tool.
								Target:    ast.CallTarget{Name: "tool.TestTool.Record", IsTool: true},
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

		// Execute the loaded commands
		_, err := interp.ExecuteCommands()
		if err != nil {
			t.Fatalf("ExecuteCommands() failed: %v", err)
		}

		// Verify the tool was called correctly and in order
		if len(callLog) != 2 {
			t.Fatalf("Expected mock tool to be called 2 times, but was called %d times", len(callLog))
		}
		if callLog[0] != "first" {
			t.Errorf("Expected first call argument to be 'first', got '%s'", callLog[0])
		}
		if callLog[1] != "second" {
			t.Errorf("Expected second call argument to be 'second', got '%s'", callLog[1])
		}
	})
}
