// NeuroScript Version: 0.8.0
// File version: 17
// Purpose: Corrected Capability struct instantiation to use Verbs and Scopes slices.
// filename: pkg/interpreter/tools.go
// nlines: 60
// risk_rating: HIGH

package interpreter

import (
	"github.com/aprice2704/neuroscript/pkg/capability"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/policygate"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// CallTool satisfies the tool.Runtime interface. It's the bridge for tools calling other tools.
func (i *Interpreter) CallTool(toolName types.FullName, args []any) (any, error) {
	fullToolNameForLookup := types.FullName("tool." + string(toolName))

	// Policy check for tool-to-tool calls.
	cap := capability.Capability{
		Resource: capability.ResTool,
		Verbs:    []string{capability.VerbExec},
		Scopes:   []string{string(fullToolNameForLookup)},
	}
	if err := policygate.Check(i, cap); err != nil {
		return nil, err
	}

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
func (i *Interpreter) ExecuteTool(toolName types.FullName, args map[string]lang.Value) (lang.Value, error) {
	// Policy check for script-to-tool calls.
	cap := capability.Capability{
		Resource: capability.ResTool,
		Verbs:    []string{capability.VerbExec},
		Scopes:   []string{string(toolName)},
	}
	if err := policygate.Check(i, cap); err != nil {
		return nil, err
	}
	return i.tools.ExecuteTool(toolName, args)
}
