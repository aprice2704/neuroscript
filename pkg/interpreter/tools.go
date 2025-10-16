// NeuroScript Version: 0.8.0
// File version: 21
// Purpose: Pass the current interpreter 'i' as the runtime to ExecuteTool, ensuring the tool registry gets the correct *clone* for internal tools.
// filename: pkg/interpreter/tools.go
// nlines: 54
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

	// THE FIX: We must pass 'i' (the current interpreter clone) as the runtime.
	// This allows the tool registry to differentiate:
	// 1. For Internal tools: Pass 'i' (the clone) so they get the correct ephemeral context.
	// 2. For External tools: Pass 'i.PublicAPI' (the wrapper) so they get the identity context.
	resultVal, err := i.tools.ExecuteTool(i, fullToolNameForLookup, langArgs)
	if err != nil {
		return nil, err
	}

	return lang.Unwrap(resultVal), nil
}

// ExecuteTool is the primary entry point for the interpreter's 'call' statement.
// It delegates directly to the tool registry, passing itself ('i') as the
// active tool.Runtime. This is critical for propagating the ephemeral turn context
// to internal tools like 'tool.aeiou.magic'.
func (i *Interpreter) ExecuteTool(toolName types.FullName, args map[string]lang.Value) (lang.Value, error) {
	// THE FIX: Pass 'i' (the clone) as the runtime.
	return i.tools.ExecuteTool(i, toolName, args)
}
