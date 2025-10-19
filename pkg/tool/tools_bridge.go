// NeuroScript Version: 0.8.0
// File version: 8
// Purpose: Removes all fmt.Fprintf(os.Stderr, ...) debug output.
// filename: pkg/tool/tools_bridge.go
// nlines: 125+
// risk_rating: MEDIUM

package tool

import (
	"errors"
	"fmt"
	"os"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/types"
	// "github.com/aprice2704/neuroscript/pkg/utils" // No longer needed here
)

// CallFromInterpreter is the single bridge between the Value-based interpreter and primitive-based tools.
// It handles policy checks, argument unwrapping/coercion, runtime context unwrapping (for internal tools),
// tool execution, and result wrapping.
func (r *ToolRegistryImpl) CallFromInterpreter(interp Runtime, fullname types.FullName, args []lang.Value) (lang.Value, error) {
	impl, ok := r.GetTool(fullname)
	if !ok {
		canonicalName := CanonicalizeToolName(string(fullname))
		errMsg := fmt.Sprintf("tool '%s' not found", canonicalName)

		// --- LOUD FAILURE (START) ---
		// Log to stderr for immediate visibility
		fmt.Fprintf(os.Stderr, "  - ERROR: Tool not found: %s\n", canonicalName)
		// Log to the host application's logger
		if interp != nil && interp.GetLogger() != nil {
			interp.GetLogger().Error("Tool call failed: "+errMsg, "tool", canonicalName)
		}
		// --- LOUD FAILURE (END) ---

		return nil, lang.NewRuntimeError(lang.ErrorCodeToolNotFound, errMsg, lang.ErrToolNotFound)
	}

	// Policy enforcement using the live interpreter context.
	if err := CanCall(interp, impl); err != nil {
		return nil, err
	}

	// --- Argument Processing ---
	rawArgs := make([]interface{}, len(args))
	for i, arg := range args {
		rawArgs[i] = lang.Unwrap(arg)
	}

	if len(rawArgs) < len(impl.Spec.Args) {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, fmt.Sprintf("tool '%s': expected at least %d args, got %d", impl.FullName, len(impl.Spec.Args), len(rawArgs)), lang.ErrArgumentMismatch)
	}

	coercedArgs := make([]interface{}, len(impl.Spec.Args))
	for i, spec := range impl.Spec.Args {
		var coercedVal interface{}
		var coerceErr error
		// Call coerceArg (now in tools_coerce.go)
		coercedVal, coerceErr = coerceArg(rawArgs[i], spec.Type)

		if coerceErr != nil {
			// Ensure error wrapping happens here if coerceArg fails
			wrappedErr := lang.NewRuntimeError(lang.ErrorCodeArgMismatch, fmt.Sprintf("tool '%s' arg '%s': %v", impl.FullName, spec.Name, coerceErr), lang.ErrArgumentMismatch)
			return nil, wrappedErr
		}
		coercedArgs[i] = coercedVal
	}

	if impl.Spec.Variadic {
		coercedArgs = append(coercedArgs, rawArgs[len(impl.Spec.Args):]...)
	}

	// --- Runtime Selection ---
	runtimeForTool := interp // Default: pass the potentially wrapped runtime
	if impl.IsInternal {
		if wrapper, ok := interp.(Wrapper); ok {
			runtimeForTool = wrapper.Unwrap() // Unwrap for internal tools
		} else {
		}
	}

	// --- Tool Invocation ---
	var out interface{}
	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = lang.NewRuntimeError(lang.ErrorCodeInternal, fmt.Sprintf("panic during tool '%s' invocation: %v", fullname, r), fmt.Errorf("panic: %v", r))
				out = nil
			}
		}()

		if impl.Func == nil {
			err = lang.NewRuntimeError(lang.ErrorCodeInternal, fmt.Sprintf("internal error: tool '%s' has nil implementation function", fullname), lang.ErrInternal)
			out = nil
			return
		}

		out, err = impl.Func(runtimeForTool, coercedArgs)
	}()

	if err != nil {
		var rtErr *lang.RuntimeError
		if !errors.As(err, &rtErr) {
			err = lang.NewRuntimeError(lang.ErrorCodeToolExecutionFailed, fmt.Sprintf("tool '%s' failed: %v", fullname, err), err)
		}
		return nil, err
	}
	wrappedOut, wrapErr := lang.Wrap(out)

	if wrapErr != nil {
		wrapRuntimeErr := lang.NewRuntimeError(lang.ErrorCodeInternal, fmt.Sprintf("failed to wrap result from tool '%s': %v", fullname, wrapErr), wrapErr)
		return nil, wrapRuntimeErr
	}

	return wrappedOut, nil
}

// ExecuteTool provides an entry point for external Go code to execute a tool using named arguments.
func (r *ToolRegistryImpl) ExecuteTool(fullname types.FullName, args map[string]lang.Value) (lang.Value, error) {
	// ... (implementation unchanged) ...
	impl, ok := r.GetTool(fullname)
	if !ok {
		canonicalName := CanonicalizeToolName(string(fullname))
		return nil, lang.NewRuntimeError(lang.ErrorCodeToolNotFound, fmt.Sprintf("tool '%s' not found", canonicalName), lang.ErrToolNotFound)
	}
	if r.interpreter == nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeConfiguration, "ToolRegistry not configured with a valid runtime context for ExecuteTool", lang.ErrConfiguration)
	}
	orderedLangArgs := make([]lang.Value, len(impl.Spec.Args))
	for i, spec := range impl.Spec.Args {
		val, ok := args[spec.Name]
		if !ok {
			if spec.Required {
				return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, fmt.Sprintf("missing required argument '%s' for tool '%s'", spec.Name, impl.FullName), lang.ErrArgumentMismatch)
			}
			orderedLangArgs[i] = lang.NilValue{}
		} else {
			orderedLangArgs[i] = val
		}
	}
	return r.CallFromInterpreter(r.interpreter, fullname, orderedLangArgs)
}
