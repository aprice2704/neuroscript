// NeuroScript Version: 0.5.2
// File version: 9
// Purpose: Simplified to only contain methods required by the tool.Runtime interface.
// filename: pkg/interpreter/interpreter_tools.go
// nlines: 60
// risk_rating: HIGH

package interpreter

import (
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// CallTool satisfies the tool.Runtime interface. It's the bridge for tools calling other tools.
func (i *Interpreter) CallTool(toolName string, args []any) (any, error) {
	// Since this is on the Runtime, args are already primitives.
	// We need to wrap them back to lang.Value for ExecuteTool.
	langArgs := make(map[string]lang.Value)
	impl, ok := i.tools.GetTool(toolName)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeToolNotFound, "tool not found: "+toolName, lang.ErrToolNotFound)
	}

	for idx, spec := range impl.Spec.Args {
		if idx < len(args) {
			wrapped, err := lang.Wrap(args[idx])
			if err != nil {
				return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "wrapping argument failed for "+spec.Name, err)
			}
			langArgs[spec.Name] = wrapped
		}
	}

	resultVal, err := i.tools.ExecuteTool(toolName, langArgs)
	if err != nil {
		return nil, err
	}

	return lang.Unwrap(resultVal), nil
}

// ExecuteTool is the primary entry point for the interpreter's 'call' statement.
func (i *Interpreter) ExecuteTool(toolName string, args map[string]lang.Value) (lang.Value, error) {
	return i.tools.ExecuteTool(toolName, args)
}
