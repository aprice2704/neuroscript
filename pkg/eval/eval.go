// NeuroScript Version: 0.8.0
// File version: 2
// Purpose: Decouples expression evaluation from the interpreter implementation.
// filename: pkg/eval/eval.go
// nlines: 30
// risk_rating: HIGH

package eval

import (
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// Runtime defines the interface the evaluator needs to interact with the interpreter.
type Runtime interface {
	GetVariable(name string) (lang.Value, bool)
	ExecuteTool(toolName types.FullName, args map[string]lang.Value) (lang.Value, error)
	RunProcedure(procName string, args ...lang.Value) (lang.Value, error)
	GetToolSpec(toolName types.FullName) (tool.ToolSpec, bool)
}

// Expression evaluates an AST expression node within the given runtime.
func Expression(rt Runtime, node ast.Expression) (lang.Value, error) {
	if node == nil {
		return &lang.NilValue{}, nil
	}
	e := &evaluation{rt: rt}
	return e.Expression(node)
}
