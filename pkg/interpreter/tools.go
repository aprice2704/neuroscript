// NeuroScript Version: 0.8.0
// File version: 18
// Purpose: Reverted to the simple version. 'tool.aeiou.magic' is no longer handled here.
// filename: pkg/interpreter/tools.go
// nlines: 50
// risk_rating: HIGH

package interpreter

import (
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// CallTool satisfies the tool.Runtime interface. It's the bridge for tools calling other tools.
func (i *Interpreter) CallTool(toolName types.FullName, args []any) (any, error) {
	// The tool registry's ExecuteTool will handle the policy check internally.
	// We wrap the arguments and delegate directly.
	fullToolNameForLookup := types.FullName("tool." + string(toolName))

	langArgs := make(map[string]lang.Value)
	impl, ok := i.tools.GetTool(fullToolNameForLookup)
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

	resultVal, err := i.tools.ExecuteTool(fullToolNameForLookup, langArgs)
	if err != nil {
		return nil, err
	}

	return lang.Unwrap(resultVal), nil
}

// ExecuteTool is the primary entry point for the interpreter's 'call' statement.
// It delegates directly to the tool registry.
func (i *Interpreter) ExecuteTool(toolName types.FullName, args map[string]lang.Value) (lang.Value, error) {
	return i.tools.ExecuteTool(toolName, args)
}
