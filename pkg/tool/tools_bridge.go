// NeuroScript Version: 0.8.0
// File version: 7
// Purpose: Made tool-not-found error "louder" by logging to stderr and the runtime logger.
// filename: pkg/tool/tools_bridge.go
// nlines: 125+
// risk_rating: HIGH

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
	// DEBUG: Logging context
	fmt.Fprintf(os.Stderr, "--- DEBUG: CallFromInterpreter for tool '%s' ---\n", fullname)
	fmt.Fprintf(os.Stderr, "  - Runtime from argument (interp): %T\n", interp)
	fmt.Fprintf(os.Stderr, "  - Runtime from registry (r.interpreter): %T\n", r.interpreter)

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
	fmt.Fprintf(os.Stderr, "  - DEBUG: Found ToolImplementation. Name: %s, IsInternal: %v\n", impl.FullName, impl.IsInternal) // DEBUG

	// Policy enforcement using the live interpreter context.
	if err := CanCall(interp, impl); err != nil {
		fmt.Fprintf(os.Stderr, "  - DEBUG: CanCall failed: %v\n", err) // DEBUG
		return nil, err
	}
	fmt.Fprintf(os.Stderr, "  - DEBUG: CanCall succeeded.\n") // DEBUG

	// --- Argument Processing ---
	fmt.Fprintf(os.Stderr, "  - DEBUG: Unwrapping %d lang.Value arguments...\n", len(args)) // DEBUG
	rawArgs := make([]interface{}, len(args))
	for i, arg := range args {
		// DEBUG: Log before unwrapping
		fmt.Fprintf(os.Stderr, "    - DEBUG: Unwrapping args[%d]: (%T) %#v\n", i, arg, arg) // DEBUG
		rawArgs[i] = lang.Unwrap(arg)
		// DEBUG: Log after unwrapping
		fmt.Fprintf(os.Stderr, "    - DEBUG:   -> rawArgs[%d]: (%T) %#v\n", i, rawArgs[i], rawArgs[i]) // DEBUG
	}
	fmt.Fprintf(os.Stderr, "  - DEBUG: Unwrapping complete.\n") // DEBUG

	if len(rawArgs) < len(impl.Spec.Args) {
		fmt.Fprintf(os.Stderr, "  - DEBUG: Arg count mismatch. Expected %d, got %d.\n", len(impl.Spec.Args), len(rawArgs)) // DEBUG
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, fmt.Sprintf("tool '%s': expected at least %d args, got %d", impl.FullName, len(impl.Spec.Args), len(rawArgs)), lang.ErrArgumentMismatch)
	}

	fmt.Fprintf(os.Stderr, "  - DEBUG: Coercing %d arguments to spec...\n", len(impl.Spec.Args)) // DEBUG
	coercedArgs := make([]interface{}, len(impl.Spec.Args))
	for i, spec := range impl.Spec.Args {
		fmt.Fprintf(os.Stderr, "    - DEBUG: Coercing arg %d ('%s') to type '%s'. Input value: (%T) %#v\n", i, spec.Name, spec.Type, rawArgs[i], rawArgs[i]) // DEBUG
		var coercedVal interface{}
		var coerceErr error
		// Call coerceArg (now in tools_coerce.go)
		coercedVal, coerceErr = coerceArg(rawArgs[i], spec.Type)
		// DEBUG: Log result of coerceArg IMMEDIATELY
		fmt.Fprintf(os.Stderr, "    - DEBUG: coerceArg returned: value=(%T)%#v, error=%v\n", coercedVal, coercedVal, coerceErr) // DEBUG

		if coerceErr != nil {
			fmt.Fprintf(os.Stderr, "    - DEBUG: Coercion FAILED: %v\n", coerceErr) // DEBUG
			// Ensure error wrapping happens here if coerceArg fails
			wrappedErr := lang.NewRuntimeError(lang.ErrorCodeArgMismatch, fmt.Sprintf("tool '%s' arg '%s': %v", impl.FullName, spec.Name, coerceErr), lang.ErrArgumentMismatch)
			fmt.Fprintf(os.Stderr, "    - DEBUG: Returning wrapped coercion error: %v\n", wrappedErr) // DEBUG
			return nil, wrappedErr
		}
		coercedArgs[i] = coercedVal
		fmt.Fprintf(os.Stderr, "    - DEBUG: Coerced arg %d ('%s') successful. Final Value: (%T) %#v\n", i, spec.Name, coercedArgs[i], coercedArgs[i]) // DEBUG
	}
	fmt.Fprintf(os.Stderr, "  - DEBUG: Coercion loop complete.\n") // DEBUG

	fmt.Fprintf(os.Stderr, "  - DEBUG: Checking for variadic args. IsVariadic: %v\n", impl.Spec.Variadic) // DEBUG
	if impl.Spec.Variadic {
		coercedArgs = append(coercedArgs, rawArgs[len(impl.Spec.Args):]...)
		fmt.Fprintf(os.Stderr, "  - DEBUG: Appended %d variadic args.\n", len(rawArgs)-len(impl.Spec.Args)) // DEBUG
	}

	// --- Runtime Selection ---
	fmt.Fprintf(os.Stderr, "  - DEBUG: Selecting runtime for tool call...\n") // DEBUG
	runtimeForTool := interp                                                  // Default: pass the potentially wrapped runtime
	if impl.IsInternal {
		if wrapper, ok := interp.(Wrapper); ok {
			runtimeForTool = wrapper.Unwrap()                                                                           // Unwrap for internal tools
			fmt.Fprintf(os.Stderr, "  - DEBUG: Unwrapped runtime for internal tool. Type is now: %T\n", runtimeForTool) // DEBUG
		} else {
			fmt.Fprintf(os.Stderr, "  - DEBUG: WARNING: Internal tool '%s' running with wrapped runtime %T (does not implement tool.Wrapper).\n", fullname, interp) // DEBUG
		}
	}
	fmt.Fprintf(os.Stderr, "  - DEBUG: Runtime selection complete. Type: %T\n", runtimeForTool) // DEBUG

	// <<< ADDED POINTER LOGGING HERE >>>
	fmt.Fprintf(os.Stderr, "  - DEBUG: Pointer stored in impl.Func before invocation: %p\n", impl.Func) // DEBUG
	// <<< END POINTER LOGGING >>>

	// --- Tool Invocation ---
	fmt.Fprintf(os.Stderr, "  - DEBUG: >>> INVOKING TOOL: impl.Func(runtimeForTool, coercedArgs) <<<\n") // DEBUG

	var out interface{}
	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Fprintf(os.Stderr, "  - DEBUG: !!! PANIC CAUGHT DURING impl.Func INVOCATION !!!: %v\n", r) // DEBUG
				err = lang.NewRuntimeError(lang.ErrorCodeInternal, fmt.Sprintf("panic during tool '%s' invocation: %v", fullname, r), fmt.Errorf("panic: %v", r))
				out = nil
				fmt.Fprintf(os.Stderr, "  - DEBUG: Panic converted to error: %v\n", err) // DEBUG
			}
		}()

		if impl.Func == nil {
			fmt.Fprintf(os.Stderr, "  - DEBUG: !!! ERROR: impl.Func is nil for tool '%s' !!!\n", fullname) // DEBUG
			err = lang.NewRuntimeError(lang.ErrorCodeInternal, fmt.Sprintf("internal error: tool '%s' has nil implementation function", fullname), lang.ErrInternal)
			out = nil
			return
		}

		fmt.Fprintf(os.Stderr, "  - DEBUG: --- Calling impl.Func now ---\n") // DEBUG
		out, err = impl.Func(runtimeForTool, coercedArgs)
		fmt.Fprintf(os.Stderr, "  - DEBUG: --- impl.Func returned ---\n") // DEBUG
	}()

	fmt.Fprintf(os.Stderr, "  - DEBUG: <<< TOOL RETURNED (or recovered): out=(%T)%#v, err=%v >>>\n", out, out, err) // DEBUG

	if err != nil {
		fmt.Fprintf(os.Stderr, "  - DEBUG: Tool invocation returned an error (or panic recovered as error): %v\n", err) // DEBUG
		var rtErr *lang.RuntimeError
		if !errors.As(err, &rtErr) {
			err = lang.NewRuntimeError(lang.ErrorCodeToolExecutionFailed, fmt.Sprintf("tool '%s' failed: %v", fullname, err), err)
			fmt.Fprintf(os.Stderr, "  - DEBUG: Wrapped non-runtime tool error: %v\n", err) // DEBUG
		}
		fmt.Fprintf(os.Stderr, "  - DEBUG: Returning error from CallFromInterpreter (Tool error case): %v\n", err) // DEBUG
		return nil, err
	}
	fmt.Fprintf(os.Stderr, "  - DEBUG: Tool invocation successful. Wrapping result...\n") // DEBUG
	wrappedOut, wrapErr := lang.Wrap(out)
	fmt.Fprintf(os.Stderr, "  - DEBUG: lang.Wrap returned: value=(%T)%#v, error=%v\n", wrappedOut, wrappedOut, wrapErr) // DEBUG

	if wrapErr != nil {
		fmt.Fprintf(os.Stderr, "  - DEBUG: lang.Wrap failed: %v\n", wrapErr) // DEBUG
		wrapRuntimeErr := lang.NewRuntimeError(lang.ErrorCodeInternal, fmt.Sprintf("failed to wrap result from tool '%s': %v", fullname, wrapErr), wrapErr)
		fmt.Fprintf(os.Stderr, "  - DEBUG: Returning error from CallFromInterpreter (Wrap error case): %v\n", wrapRuntimeErr) // DEBUG
		return nil, wrapRuntimeErr
	}

	fmt.Fprintf(os.Stderr, "  - DEBUG: Returning result from CallFromInterpreter (Success case): (%T)%#v\n", wrappedOut, wrappedOut) // DEBUG
	fmt.Fprintln(os.Stderr, "-------------------------------------------------")                                                     // DEBUG
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
