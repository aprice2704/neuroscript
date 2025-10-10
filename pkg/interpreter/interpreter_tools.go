// NeuroScript Version: 0.8.0
// File version: 16
// Purpose: FIX: Replaces direct access to the removed 'tools' field with calls to the ToolRegistry() method.
// filename: pkg/interpreter/interpreter_tools.go
// nlines: 60
// risk_rating: HIGH

package interpreter

import (
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// CallTool satisfies the tool.Runtime interface. It's the bridge for tools calling other tools.
func (i *Interpreter) CallTool(toolName types.FullName, args []any) (any, error) {
	// A tool calling another tool provides the group-qualified name (e.g., "my.test.tools.echo").
	// We must prepend "tool." to construct the full canonical name used as the key in the registry.
	fullToolNameForLookup := types.FullName("tool." + string(toolName))

	// Since this is on the Runtime, args are already primitives.
	// We need to wrap them back to lang.Value for ExecuteTool.
	langArgs := make(map[string]lang.Value)
	impl, ok := i.ToolRegistry().GetTool(fullToolNameForLookup)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeToolNotFound, "tool not found: "+string(fullToolNameForLookup), lang.ErrToolNotFound)
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

	resultVal, err := i.ToolRegistry().ExecuteTool(fullToolNameForLookup, langArgs)
	if err != nil {
		return nil, err
	}

	return lang.Unwrap(resultVal), nil
}

// ExecuteTool is the primary entry point for the interpreter's 'call' statement.
func (i *Interpreter) ExecuteTool(toolName types.FullName, args map[string]lang.Value) (lang.Value, error) {
	return i.ToolRegistry().ExecuteTool(toolName, args)
}
