// NeuroScript Version: 0.8.0
// File version: 12
// Purpose: Adds a defer/recover block to ExecuteTool for panic safety, matching CallFromInterpreter.
// filename: pkg/tool/tools_bridge.go
// nlines: 178
// risk_rating: HIGH

package tool

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/types"
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
		fmt.Fprintf(os.Stderr, "  - ERROR: Tool not found: %s\n", canonicalName)
		if interp != nil && interp.GetLogger() != nil {
			interp.GetLogger().Error("Tool call failed: "+errMsg, "tool", canonicalName)
		}
		// --- LOUD FAILURE (END) ---

		return nil, lang.NewRuntimeError(lang.ErrorCodeToolNotFound, errMsg, lang.ErrToolNotFound)
	}

	// Policy enforcement using the live interpreter context.
	if err := CanCall(interp, impl); err != nil {
		return nil, err // Return policy violation error directly
	}

	// --- Argument Unwrapping ---
	rawArgs := make([]interface{}, len(args))
	for i, arg := range args {
		rawArgs[i] = lang.Unwrap(arg)
	}

	// --- Centralized Validation and Coercion ---
	coercedArgs, validationErr := validateAndCoerceArgs(impl.FullName, rawArgs, impl.Spec)
	if validationErr != nil {
		return nil, validationErr
	}

	// --- Runtime Selection ---
	runtimeForTool := interp
	if impl.IsInternal {
		if wrapper, ok := interp.(Wrapper); ok {
			runtimeForTool = wrapper.Unwrap()
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

	// --- Result Handling ---
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
// It now uses the centralized validation logic.
func (r *ToolRegistryImpl) ExecuteTool(fullname types.FullName, args map[string]lang.Value) (lang.Value, error) {
	impl, ok := r.GetTool(fullname)
	if !ok {
		canonicalName := CanonicalizeToolName(string(fullname))
		return nil, lang.NewRuntimeError(lang.ErrorCodeToolNotFound, fmt.Sprintf("tool '%s' not found", canonicalName), lang.ErrToolNotFound)
	}
	if r.interpreter == nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeConfiguration, "ToolRegistry not configured with a valid runtime context for ExecuteTool", lang.ErrConfiguration)
	}

	// --- [FIX] Policy Enforcement ---
	//fmt.Fprintf(os.Stderr, "[DEBUG][ExecuteTool] Checking policy for tool: %s\n", impl.FullName)
	if err := CanCall(r.interpreter, impl); err != nil {
		//	fmt.Fprintf(os.Stderr, "[DEBUG][ExecuteTool] Policy check FAILED for %s: %v\n", impl.FullName, err)
		return nil, err // Return policy violation error directly
	}
	//fmt.Fprintf(os.Stderr, "[DEBUG][ExecuteTool] Policy check PASSED for: %s\n", impl.FullName)
	// --- End [FIX] ---

	// --- Build positional rawArgs from named args map ---
	numSpecArgs := len(impl.Spec.Args)
	rawArgs := make([]any, numSpecArgs) // Size exactly to spec
	var missingRequired []string

	for i, spec := range impl.Spec.Args {
		val, found := args[spec.Name]
		if !found {
			if spec.Required {
				missingRequired = append(missingRequired, spec.Name)
			}
			rawArgs[i] = nil // Not provided, insert nil
		} else {
			rawArgs[i] = lang.Unwrap(val) // Found, unwrap it
		}
	}

	// Check for missing required args *before* validation call for a clearer error
	if len(missingRequired) > 0 {
		errMsg := fmt.Sprintf("missing required arguments for tool '%s': %s", impl.FullName, strings.Join(missingRequired, ", "))
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, errMsg, lang.ErrArgumentMismatch)
	}

	// --- Centralized Validation and Coercion ---
	coercedArgs, validationErr := validateAndCoerceArgs(impl.FullName, rawArgs, impl.Spec)
	if validationErr != nil {
		return nil, validationErr // Return detailed error from validator
	}

	// --- Tool Invocation (Simplified, assumes external calls don't need runtime unwrapping) ---
	// Note: ExecuteTool uses the registry's base interpreter context.
	// --- [NEW] Add panic recovery ---
	var out interface{}
	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				// This now creates the error message the test was originally (and correctly) expecting.
				err = lang.NewRuntimeError(lang.ErrorCodeInternal, fmt.Sprintf("panic during tool '%s' invocation: %v", fullname, r), fmt.Errorf("panic: %v", r))
				out = nil
				//	fmt.Fprintf(os.Stderr, "[DEBUG][ExecuteTool] Recovered panic from tool %s: %v\n", fullname, r)
			}
		}()
		out, err = impl.Func(r.interpreter, coercedArgs) // Pass validated args
	}()
	// --- End [NEW] ---

	// --- Result Handling ---
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
