// NeuroScript Version: 0.8.0
// File version: 6
// Purpose: Corrects invalid script syntax and updates the panic safety test to use a zero-value struct.
// filename: pkg/api/analysis/tool_visitor_test.go
// nlines: 100+
// risk_rating: LOW

package analysis_test

import (
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
	"github.com/aprice2704/neuroscript/pkg/api/analysis"
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
)

func TestFindRequiredTools(t *testing.T) {
	tests := []struct {
		name   string
		script string
		want   map[string]struct{}
	}{
		{
			name:   "No tool calls",
			script: "func main() means\n return 1 \nendfunc",
			want:   map[string]struct{}{},
		},
		{
			name:   "Single tool call in command",
			script: "command\n call tool.fs.read(\"path\") \nendcommand",
			want:   map[string]struct{}{"fs.read": {}},
		},
		{
			name:   "Single tool call in procedure",
			script: "func main() means\n must tool.str.inspect(var) \nendfunc",
			want:   map[string]struct{}{"str.inspect": {}},
		},
		{
			name: "Multiple unique tool calls",
			script: `
                command
                    set x = tool.math.add(1, 2)
                    emit tool.str.inspect(x)
                endcommand`,
			want: map[string]struct{}{"math.add": {}, "str.inspect": {}},
		},
		{
			name: "Duplicate tool calls",
			script: `
                command
                    call tool.fs.read("a")
                    call tool.fs.read("b")
                endcommand`,
			want: map[string]struct{}{"fs.read": {}},
		},
		{
			name: "Call to regular procedure (should be ignored)",
			script: `
                command
                    call my_proc()
                endcommand`,
			want: map[string]struct{}{},
		},
		{
			name:   "Tool call inside nested expression",
			script: "command\n set x = 1 + tool.math.add(2, 3) \nendcommand",
			want:   map[string]struct{}{"math.add": {}},
		},
		{
			name:   "Script with nil step fields (for panic safety)",
			script: "func main() means\n return \nendfunc", // Return has no values
			want:   map[string]struct{}{},
		},
		{
			name: "Tool call in event handler",
			script: `
				on event "foo" do
					call tool.events.bar()
				endon
			`,
			want: map[string]struct{}{"events.bar": {}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree, err := api.Parse([]byte(tt.script), api.ParseSkipComments)
			if err != nil {
				t.Fatalf("api.Parse failed: %v", err)
			}

			if tree == nil || tree.Root == nil {
				t.Fatal("Parsing resulted in a nil tree or root node")
			}

			got := analysis.FindRequiredTools(&interfaces.Tree{Root: tree.Root})

			if len(got) == 0 {
				got = make(map[string]struct{})
			}
			if len(tt.want) == 0 {
				tt.want = make(map[string]struct{})
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindRequiredTools() mismatch:\n  Got:  %v\n  Want: %v", got, tt.want)
			}
		})
	}
}

// TestFindRequiredTools_PanicSafety ensures the visitor doesn't panic on malformed AST nodes.
func TestFindRequiredTools_PanicSafety(t *testing.T) {
	t.Run("CallableExprNode with zero-value Target", func(t *testing.T) {
		// Manually construct an AST with a zero-value Target struct.
		malformedAST := &interfaces.Tree{
			Root: &ast.Program{
				Commands: []*ast.CommandNode{
					{
						Body: []ast.Step{
							{
								Type: "expression_statement",
								ExpressionStmt: &ast.ExpressionStatementNode{
									Expression: &ast.CallableExprNode{
										Target: ast.CallTarget{}, // FIX: Use zero-value struct instead of nil
									},
								},
							},
						},
					},
				},
			},
		}

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("FindRequiredTools panicked on a malformed AST: %v", r)
			}
		}()

		// This function should now execute without panicking.
		_ = analysis.FindRequiredTools(malformedAST)
	})
}
