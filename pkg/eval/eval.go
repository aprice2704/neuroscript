// NeuroScript Version: 0.8.0
// File version: 4
// Purpose: Adds the 'Required' field to the local ArgSpec to fully decouple it from the tool package.
// filename: pkg/eval/eval.go
// nlines: 43
// risk_rating: HIGH

package eval

import (
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// ArgSpec defines the minimal specification for a tool argument needed by the evaluator.
type ArgSpec struct {
	Name     string
	Type     string
	Required bool
}

// ToolSpec defines the minimal tool specification needed by the evaluator.
// This decouples the eval package from the broader tool package.
type ToolSpec struct {
	FullName types.FullName
	Args     []ArgSpec
}

// Runtime defines the interface the evaluator needs to interact with the interpreter.
type Runtime interface {
	GetVariable(name string) (lang.Value, bool)
	ExecuteTool(toolName types.FullName, args map[string]lang.Value) (lang.Value, error)
	RunProcedure(procName string, args ...lang.Value) (lang.Value, error)
	GetToolSpec(toolName types.FullName) (ToolSpec, bool)
}

// Expression evaluates an AST expression node within the given runtime.
func Expression(rt Runtime, node ast.Expression) (lang.Value, error) {
	if node == nil {
		return &lang.NilValue{}, nil
	}
	e := &evaluation{rt: rt}
	return e.Expression(node)
}
